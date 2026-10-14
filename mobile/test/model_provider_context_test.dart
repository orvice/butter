import 'dart:async';

import 'package:butter_mobile/app/client_context.dart';
import 'package:butter_mobile/features/model_providers/provider_draft.dart';
import 'package:butter_mobile/gen/agents/v1/agent.pb.dart';
import 'package:butter_mobile/gen/agents/v1/agent_service.pb.dart';
import 'package:flutter_test/flutter_test.dart';

import 'support/memory_storage.dart';
import 'support/provider_fixtures.dart';
import 'support/provider_server.dart';

ModelProvider edited(
  ModelProvider base, {
  KeyAction key = KeyAction.keep,
  String newKey = '',
  String? baseUrl,
  List<int>? keep,
}) => buildModelProvider(
  base: base,
  type: base.type,
  baseUrl: baseUrl ?? base.baseUrl,
  key: key,
  newKey: newKey,
  rows: [
    for (final i in keep ?? List.generate(base.models.length, (i) => i))
      ModelRowInput(
        original: base.models[i],
        name: base.models[i].name,
        alias: base.models[i].alias,
        contextWindow: base.models[i].contextWindowTokens == 0
            ? ''
            : '${base.models[i].contextWindowTokens}',
      ),
  ],
);

void main() {
  late ProviderServer server;
  late ClientContext client;
  setUp(() async {
    server = ProviderServer();
    await server.start();
    client = ClientContext(storage: MemoryStorage(), allowHttp: true);
    await client.setServer(server.url);
    await client.login('alice', 'password');
    await client.selectWorkspace('alpha');
  });
  tearDown(() async {
    client.dispose();
    await server.close();
  });

  test('reads are scoped to the selected Workspace header', () async {
    server.providers['alpha']!['one'] = providerWithUnknown('one');
    server.providers['beta']!['two'] = providerWithUnknown('two');
    await client.loadModelProviders();
    expect(client.modelProviders.map((p) => p.name), ['one']);
    expect(client.modelProvidersLoaded, isTrue);
    await client.loadModelProvider('one');
    expect(client.modelProvider!.name, 'one');
    expect(
      server.calls.map((c) => c.workspace),
      everyElement('alpha'),
      reason: 'every provider call carries X-Workspace-ID',
    );
    await client.selectWorkspace('beta');
    expect(client.modelProviders, isEmpty);
    expect(client.modelProvider, isNull);
    await client.loadModelProviders();
    expect(client.modelProviders.map((p) => p.name), ['two']);
  });

  test(
    'an edit sends the whole object and keeps what was not edited',
    () async {
      server.providers['alpha']!['one'] = providerWithUnknown('one');
      await client.loadModelProvider('one');
      final base = client.modelProvider!;
      expect(
        await client.saveModelProvider(
          edited(base, baseUrl: 'https://new.example/v1', keep: [0, 2]),
          create: false,
        ),
        isTrue,
      );
      final sent = UpdateModelProviderRequest.fromBuffer(
        server.lastBody('UpdateModelProvider'),
      ).modelProvider;
      expect(sent.apiKey, 'sk-secret-value', reason: 'keep resends the key');
      expect(sent.workspaceId, 'alpha');
      expect(unknownVarint(sent, 99), 7);
      expect([for (final m in sent.models) m.name], ['gpt-4o', 'gpt-nano']);
      expect([for (final m in sent.models) unknownVarint(m, 15)], [1, 3]);
      final stored = server.providers['alpha']!['one']!;
      expect(stored.baseUrl, 'https://new.example/v1');
      expect(client.modelProviders.single.baseUrl, 'https://new.example/v1');
      expect(client.modelProvider!.models.length, 2);
    },
  );

  test('replacing and clearing the key are explicit', () async {
    server.providers['alpha']!['one'] = providerWithUnknown('one');
    await client.loadModelProvider('one');
    expect(
      await client.saveModelProvider(
        edited(
          client.modelProvider!,
          key: KeyAction.replace,
          newKey: 'sk-new-value',
        ),
        create: false,
      ),
      isTrue,
    );
    expect(server.providers['alpha']!['one']!.apiKey, 'sk-new-value');
    expect(
      await client.saveModelProvider(
        edited(client.modelProvider!, key: KeyAction.clear),
        create: false,
      ),
      isTrue,
    );
    expect(server.providers['alpha']!['one']!.apiKey, '');
    expect(server.providers['alpha']!['one']!.models.length, 3);
  });

  test(
    'create is a distinct call, and a duplicate name is explained',
    () async {
      final draft = buildModelProvider(
        name: 'fresh',
        type: 'gemini',
        baseUrl: '',
        key: KeyAction.replace,
        newKey: 'sk-fresh-value',
        rows: const [
          ModelRowInput(
            name: 'gemini-pro',
            alias: 'pro',
            contextWindow: '4294967295',
          ),
        ],
      );
      expect(await client.saveModelProvider(draft, create: true), isTrue);
      expect(server.calls.last.method, 'CreateModelProvider');
      expect(client.modelProviders.single.models.single.alias, 'pro');
      expect(await client.saveModelProvider(draft, create: true), isFalse);
      expect(client.error, contains('already exists'));
      expect(client.error, isNot(contains('sk-fresh-value')));
      expect(client.modelProviders, hasLength(1));
      expect(server.writes, 2, reason: 'a refused create is not retried');
    },
  );

  test(
    'a refusal that names Telegram references is shown and nothing is repeated',
    () async {
      server.providers['alpha']!['one'] = providerWithUnknown('one');
      await client.loadModelProviders();
      server.refusal = (
        412,
        'failed_precondition',
        'models are referenced by telegram destinations: fast (dest-1, dest-2)',
      );
      expect(await client.deleteModelProvider('one'), isFalse);
      expect(client.error, contains('cannot be deleted'));
      expect(client.error, contains('fast (dest-1, dest-2)'));
      expect(client.modelProviders, hasLength(1));
      await client.loadModelProvider('one');
      expect(
        await client.saveModelProvider(
          edited(client.modelProvider!, keep: [1, 2]),
          create: false,
        ),
        isFalse,
      );
      expect(client.error, contains('cannot be removed'));
      expect(client.error, contains('dest-2'));
      expect(client.writeNotice, isNull, reason: 'a refusal is not ambiguous');
      expect(server.writes, 2);
      expect(server.providers['alpha']!['one']!.models, hasLength(3));
    },
  );

  test('a lost reply is reported as unconfirmed and never replayed', () async {
    server.providers['alpha']!['one'] = providerWithUnknown('one');
    await client.loadModelProvider('one');
    server.loseWriteReply = true;
    expect(
      await client.saveModelProvider(
        edited(client.modelProvider!, baseUrl: 'https://x.example'),
        create: false,
      ),
      isFalse,
    );
    expect(client.writeNotice, contains('not confirmed'));
    expect(client.error, isNot(contains('sk-secret-value')));
    // The write did land; the app did not claim otherwise and did not retry.
    expect(server.writes, 1);
    expect(server.providers['alpha']!['one']!.baseUrl, 'https://x.example');
    server.loseWriteReply = false;
    await client.revalidate();
    expect(server.writes, 1, reason: 'foreground recovery only reads');
  });

  test('permission denial keeps the provider untouched', () async {
    server.providers['alpha']!['one'] = providerWithUnknown('one');
    await client.loadModelProvider('one');
    server.refusal = (403, 'permission_denied', 'denied');
    expect(await client.deleteModelProvider('one'), isFalse);
    expect(client.error, contains('permission'));
    expect(client.modelProvider!.name, 'one');
    expect(server.writes, 1);
  });

  test('late responses never restore data into another scope', () async {
    server.providers['alpha']!['one'] = providerWithUnknown('one');
    final entered = Completer<void>();
    final release = Completer<void>();
    server.beforeList = () async {
      entered.complete();
      await release.future;
    };
    final loading = client.loadModelProviders();
    await entered.future;
    await client.selectWorkspace('beta');
    release.complete();
    await loading;
    expect(client.workspace!.id, 'beta');
    expect(client.modelProviders, isEmpty);
    expect(client.modelProvidersLoaded, isFalse);

    // The same for a detail read and for a write answered after a switch.
    await client.selectWorkspace('alpha');
    final detailEntered = Completer<void>();
    final detailRelease = Completer<void>();
    server.beforeGet = () async {
      detailEntered.complete();
      await detailRelease.future;
    };
    final detail = client.loadModelProvider('one');
    await detailEntered.future;
    await client.selectWorkspace('beta');
    detailRelease.complete();
    await detail;
    expect(client.modelProvider, isNull);
    server.beforeGet = null;

    await client.selectWorkspace('alpha');
    await client.loadModelProvider('one');
    final base = client.modelProvider!;
    final writing = Completer<void>();
    final writeRelease = Completer<void>();
    server.beforeWrite = () async {
      writing.complete();
      await writeRelease.future;
    };
    final saving = client.saveModelProvider(
      edited(base, baseUrl: 'https://late.example'),
      create: false,
    );
    await writing.future;
    await client.selectWorkspace('beta');
    writeRelease.complete();
    expect(await saving, isFalse);
    expect(client.modelProviders, isEmpty);
    expect(client.modelProvider, isNull);
    expect(server.writes, 1, reason: 'the interrupted write is not replayed');
  });

  test('an expired session or another account drops every provider', () async {
    server.providers['alpha']!['one'] = providerWithUnknown('one');
    await client.loadModelProvider('one');
    await client.loadModelProviders();
    server.expired = true;
    await client.loadModelProviders();
    expect(client.user, isNull);
    expect(client.modelProviders, isEmpty);
    expect(client.modelProvider, isNull);

    server.expired = false;
    await client.login('alice', 'password');
    await client.selectWorkspace('alpha');
    await client.loadModelProvider('one');
    await client.logout();
    await client.login('bob', 'password');
    expect(client.modelProviders, isEmpty);
    expect(client.modelProvider, isNull);
    expect(client.modelProvidersLoaded, isFalse);
  });

  test('pausing drops keys; resuming only re-reads what was open', () async {
    server.providers['alpha']!['one'] = providerWithUnknown('one');
    await client.loadModelProviders();
    await client.loadModelProvider('one');
    final entered = Completer<void>();
    final release = Completer<void>();
    server.beforeList = () async {
      entered.complete();
      await release.future;
    };
    final reload = client.loadModelProviders();
    await entered.future;
    client.suspend();
    release.complete();
    await reload;
    expect(client.modelProviders, isEmpty, reason: 'a late read after pause');
    expect(client.modelProvider, isNull);
    server.beforeList = null;

    final writesBefore = server.writes;
    await client.loadModelProviders();
    await client.loadModelProvider('one');
    client.suspend();
    expect(client.modelProviders, isEmpty);
    await client.revalidate();
    expect(client.modelProviders.map((p) => p.name), ['one']);
    expect(client.modelProvider!.name, 'one');
    expect(server.writes, writesBefore);
  });

  test(
    'losing access to the Workspace drops its providers on resume',
    () async {
      server.providers['alpha']!['one'] = providerWithUnknown('one');
      await client.loadModelProviders();
      server.spaces.removeWhere((w) => w.id == 'alpha');
      client.suspend();
      await client.revalidate();
      expect(client.workspace, isNull);
      expect(client.modelProviders, isEmpty);
    },
  );

  group('resume failures keep their own wording', () {
    // Only provider reads may say "model provider"; a failing identity,
    // Workspace-management or Agent read must not be blamed on a provider.
    Future<void> managedReadFailsOnResume() async {
      await client.openWorkspace('alpha');
      expect(client.managedWorkspace!.id, 'alpha');
      client.suspend();
      server.missingWorkspaces.add('alpha');
      await client.revalidate();
      expect(client.error, isNotNull);
      expect(client.error!.toLowerCase(), isNot(contains('provider')));
      expect(client.error, isNot(contains('no longer exists')));
    }

    test(
      'a managed Workspace read failing, providers never opened',
      managedReadFailsOnResume,
    );

    test('a managed Workspace read failing, providers opened', () async {
      server.providers['alpha']!['one'] = providerWithUnknown('one');
      await client.loadModelProvider('one');
      final before = server.calls.length;
      await managedReadFailsOnResume();
      expect(
        server.calls.length,
        before,
        reason: 'the failed management read stops the resume before providers',
      );
    });

    test('a provider detail that vanished while paused says so', () async {
      server.providers['alpha']!['one'] = providerWithUnknown('one');
      await client.loadModelProvider('one');
      client.suspend();
      server.providers['alpha']!.remove('one');
      await client.revalidate();
      expect(client.error, contains('no longer exists'));
      expect(client.modelProvider, isNull);
    });
  });

  test('keys appear in no error or notice text', () async {
    server.providers['alpha']!['one'] = providerWithUnknown('one');
    await client.loadModelProvider('one');
    final seen = <String>[];
    client.addListener(() {
      seen
        ..add(client.error ?? '')
        ..add(client.writeNotice ?? '');
    });
    for (final failure in [
      (400, 'invalid_argument', 'sk-secret-value bad'),
      (500, 'internal', 'boom sk-secret-value'),
      (503, 'unavailable', 'sk-secret-value'),
      (412, 'failed_precondition', 'workspace sk-secret-value'),
    ]) {
      server.refusal = failure;
      await client.saveModelProvider(
        edited(client.modelProvider!, key: KeyAction.replace, newKey: 'sk-new'),
        create: false,
      );
    }
    expect(seen.where((s) => s.contains('sk-')), isEmpty);
    expect(seen.where((s) => s.isNotEmpty), isNotEmpty);
  });
}
