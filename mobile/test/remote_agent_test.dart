import 'dart:async';
import 'dart:io';

import 'package:butter_mobile/app/client_context.dart';
import 'package:butter_mobile/core/config_resource_controller.dart';
import 'package:butter_mobile/core/url_validation.dart';
import 'package:butter_mobile/features/remote_agents/remote_agent_controller.dart';
import 'package:butter_mobile/features/remote_agents/remote_agent_draft.dart';
import 'package:butter_mobile/gen/agents/v1/agent.pb.dart';
import 'package:butter_mobile/gen/agents/v1/dashboard.pb.dart';
import 'package:fixnum/fixnum.dart';
import 'package:flutter_test/flutter_test.dart';

import 'support/mcp_server.dart';
import 'support/memory_storage.dart';

RemoteAgent a2a(String id) => RemoteAgent(
  id: id,
  name: 'Agent $id',
  url: 'https://a2a.example.com',
  protocol: RemoteAgentProtocol.REMOTE_AGENT_PROTOCOL_A2A,
);

RemoteAgent opencode(String id) => RemoteAgent(
  id: id,
  name: 'Code $id',
  url: 'https://oc.example.com',
  protocol: RemoteAgentProtocol.REMOTE_AGENT_PROTOCOL_OPENCODE_HTTP,
  opencodeAgent: 'build',
  opencodeModel: 'anthropic/claude',
  username: 'oc-user',
  password: 'oc-secret',
);

RemoteAgent daemon(String id) => RemoteAgent(
  id: id,
  name: 'Daemon $id',
  protocol: RemoteAgentProtocol.REMOTE_AGENT_PROTOCOL_DAEMON,
  daemonRuntimeId: 'rt-1',
  acpRuntime: 'opencode',
);

Future<({ClientContext client, McpServerFake server})> start({
  void Function(McpServerFake)? seed,
}) async {
  final old = HttpOverrides.current;
  HttpOverrides.global = null;
  final server = McpServerFake();
  seed?.call(server);
  await server.start();
  final client = ClientContext(storage: MemoryStorage(), allowHttp: true);
  addTearDown(() async {
    client.dispose();
    await server.close();
    HttpOverrides.global = old;
  });
  await client.setServer(server.url);
  await client.login('alice', 'password');
  await client.selectWorkspace('alpha');
  return (client: client, server: server);
}

void main() {
  group('draft', () {
    test('a2a needs an ID, a name and an absolute URL', () {
      final d = RemoteAgentDraft.create();
      expect(d.firstError(), isNotNull);
      d.id = 'a';
      d.name = 'A';
      d.url = 'ftp://x';
      expect(d.firstError(), contains('http'));
      d.url = 'https://a2a.example.com';
      expect(d.firstError(), isNull);
      final a = d.build();
      expect(a.protocol, RemoteAgentProtocol.REMOTE_AGENT_PROTOCOL_A2A);
      expect(a.hasUsername(), isFalse);
    });

    test('switching an opencode agent to a2a clears opencode fields', () {
      final d = RemoteAgentDraft.edit(opencode('o1'))..kind = RemoteKind.a2a;
      final a = d.build();
      expect(a.protocol, RemoteAgentProtocol.REMOTE_AGENT_PROTOCOL_A2A);
      expect(a.opencodeAgent, '');
      expect(a.username, '');
      expect(a.password, '');
    });

    test(
      'the password is kept, replaced or cleared, never blanked by accident',
      () {
        final d = RemoteAgentDraft.edit(opencode('o1'));
        expect(d.dirty, isFalse);
        expect(d.hasStoredPassword, isTrue);
        expect(d.build(), opencode('o1'));
        d.passwordEdit = SecretEdit.replace;
        expect(d.validatePassword(), isNotNull);
        d.newPassword = 'fresh';
        expect(d.build().password, 'fresh');
        d.passwordEdit = SecretEdit.clear;
        // The username now has no password: the server would refuse it.
        expect(d.validatePassword(), contains('needs a password'));
        d.username = '';
        expect(d.validatePassword(), isNull);
        expect(d.build().password, '');
      },
    );

    test('an opencode password alone, without a username, is allowed', () {
      final d = RemoteAgentDraft.create()
        ..id = 'o'
        ..name = 'O'
        ..url = 'http://oc.example.com'
        ..kind = RemoteKind.opencode
        ..passwordEdit = SecretEdit.replace
        ..newPassword = 'pw';
      expect(d.firstError(), isNull);
      expect(d.sendsCredentialsInClear, isTrue);
      expect(d.build().password, 'pw');
    });

    test('unknown fields and the workspace survive an edit', () {
      final stored = a2a('a1')..workspaceId = 'alpha';
      stored.unknownFields.mergeVarintField(99, Int64(5));
      final a = (RemoteAgentDraft.edit(stored)..name = 'New').build();
      expect(a.workspaceId, 'alpha');
      expect(a.unknownFields.getField(99)?.varints.single.toInt(), 5);
    });

    test('Daemon agents are not editable here', () {
      expect(remoteAgentEditable(daemon('d1')), isFalse);
      expect(remoteAgentEditable(a2a('a')), isTrue);
      expect(isDaemonAgent(daemon('d1')), isTrue);
    });

    test('create is clean until something is entered', () {
      final d = RemoteAgentDraft.create();
      expect(d.dirty, isFalse);
      d.name = 'x';
      expect(d.dirty, isTrue);
    });
  });

  group('controller', () {
    test(
      'lists the active Workspace only, including its Daemon agents',
      () async {
        final h = await start(
          seed: (s) {
            s.seedRemote('alpha', a2a('a1'));
            s.seedRemote('alpha', daemon('d1'));
            s.seedRemote('beta', a2a('b1'));
          },
        );
        final c = h.client.remoteAgents!;
        await c.load();
        expect(c.items.map((a) => a.id).toSet(), {'a1', 'd1'});
      },
    );

    test('create, edit keeping the password, and delete', () async {
      final h = await start();
      final c = h.client.remoteAgents!;
      final draft = RemoteAgentDraft.create()
        ..id = 'o1'
        ..name = 'Code'
        ..url = 'https://oc.example.com'
        ..kind = RemoteKind.opencode
        ..username = 'u'
        ..passwordEdit = SecretEdit.replace
        ..newPassword = 'pw1';
      expect(await c.create(draft.build()), isTrue);
      expect(h.server.remotes['alpha']!['o1']!.password, 'pw1');

      await c.loadItem('o1');
      final edit = RemoteAgentDraft.edit(c.item!)..name = 'Code 2';
      expect(await c.update(edit.build()), isTrue);
      final stored = h.server.remotes['alpha']!['o1']!;
      expect(stored.name, 'Code 2');
      expect(stored.password, 'pw1');
      expect(stored.username, 'u');

      expect(await c.delete('o1'), isTrue);
      expect(h.server.remotes['alpha'], isEmpty);
    });

    test('a lost reply is unconfirmed and never repeated', () async {
      final h = await start();
      final c = h.client.remoteAgents!;
      h.server.loseReply = true;
      final draft = RemoteAgentDraft.create()
        ..id = 'a1'
        ..name = 'A'
        ..url = 'https://a2a.example.com';
      expect(await c.create(draft.build()), isFalse);
      expect(c.unconfirmed, isTrue);
      h.server.loseReply = false;
      expect(await c.create(draft.build()), isFalse);
      expect(h.server.count('CreateRemoteAgent'), 1);
      await c.load();
      expect(c.unconfirmed, isFalse);
      expect(c.items.single.id, 'a1');
    });

    test(
      'a refused write for lack of access is clear and not unknown',
      () async {
        final h = await start(seed: (s) => s.seedRemote('alpha', a2a('a1')));
        final c = h.client.remoteAgents!;
        h.server.refusals['DeleteRemoteAgent'] = (
          403,
          'permission_denied',
          'nope',
        );
        expect(await c.delete('a1'), isFalse);
        expect(c.writeError, contains('do not have access'));
        expect(c.unconfirmed, isFalse);
      },
    );

    test('probe states are classified and never echo server text', () async {
      final h = await start(seed: (s) => s.seedRemote('alpha', a2a('a1')));
      final c = h.client.remoteAgents!;
      h.server.remoteState = RemoteAgentStatus_State.STATE_UNREACHABLE;
      h.server.remoteDetail =
          'Get "https://user:pw-leak@a2a.example.com/.well-known/agent.json": dial tcp: refused';
      await c.probe('a1');
      var view = c.probes['a1']!;
      expect(view.kind, RemoteProbeKind.unreachable);
      expect(view.reason, isNot(contains('pw-leak')));

      h.server.remoteDetail = 'agent card returned 404';
      await c.probe('a1');
      expect(c.probes['a1']!.reason, contains('HTTP 404'));

      h.server.remoteState = RemoteAgentStatus_State.STATE_ERROR;
      h.server.remoteDetail =
          'opencode server returned 401 (check username/password)';
      await c.probe('a1');
      view = c.probes['a1']!;
      expect(view.kind, RemoteProbeKind.error);
      expect(view.reason, contains('rejected the username or password'));

      h.server.remoteState = RemoteAgentStatus_State.STATE_ACTIVE;
      await c.probe('a1');
      expect(c.probes['a1']!.kind, RemoteProbeKind.reachable);
      expect(c.probes['a1']!.latencyMs, 12);
    });

    test('suspending fences a probe in flight', () async {
      final h = await start(seed: (s) => s.seedRemote('alpha', a2a('a1')));
      final c = h.client.remoteAgents!;
      final gate = Completer<void>();
      h.server.before = (m) =>
          m == 'GetRemoteAgentStatus' ? gate.future : Future.value();
      final probe = c.probe('a1');
      await Future<void>.delayed(const Duration(milliseconds: 50));
      h.client.suspend();
      gate.complete();
      await probe;
      expect(c.probes['a1'], isNull);
    });

    test(
      'credentials are dropped on suspend and read again on recover',
      () async {
        final h = await start(
          seed: (s) => s.seedRemote('alpha', opencode('o1')),
        );
        final c = h.client.remoteAgents!;
        await c.load();
        h.client.suspend();
        expect(c.items, isEmpty);
        await c.recover();
        expect(c.items.single.password, 'oc-secret');
        expect(c.listState, ConfigLoad.ready);
      },
    );
  });
}
