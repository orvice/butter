import 'dart:io';

import 'package:butter_mobile/app/butter_app.dart';
import 'package:butter_mobile/app/client_context.dart';
import 'package:butter_mobile/gen/agents/v1/agent_service.pb.dart';
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';

import 'support/memory_storage.dart';
import 'support/provider_fixtures.dart';
import 'support/provider_server.dart';

Future<void> tapRequest(
  WidgetTester tester,
  ClientContext client,
  Finder finder,
) async {
  await tester.runAsync(() async {
    await tester.tap(finder);
    for (var i = 0; i < 200 && client.busy; i++) {
      await Future<void>.delayed(const Duration(milliseconds: 10));
    }
  });
  expect(client.busy, isFalse);
  await tester.pumpAndSettle();
}

/// Waits until a request started by a screen action has been answered: the
/// server saw [writes] write calls and the client is idle again.
Future<void> settle(
  WidgetTester tester,
  ClientContext client,
  ProviderServer server,
  int writes,
) async {
  for (var i = 0; i < 300; i++) {
    await tester.runAsync(
      () => Future<void>.delayed(const Duration(milliseconds: 10)),
    );
    await tester.pump();
    if (!client.busy && server.writes >= writes) break;
  }
  expect(client.busy, isFalse);
  expect(server.writes, writes);
  await tester.pumpAndSettle();
}

/// Taps a control whose handler awaits a dialog and then sends a request. The
/// handler's continuation must run in the real zone for its HTTP call to
/// complete, so the tap that starts it happens inside [WidgetTester.runAsync].
Future<void> tapReal(WidgetTester tester, Finder finder) async {
  await tester.ensureVisible(finder);
  await tester.runAsync(() => tester.tap(finder));
  await tester.pumpAndSettle();
}

Future<void> tapScreen(WidgetTester tester, Finder finder) async {
  await tester.ensureVisible(finder);
  await tester.tap(finder);
  await tester.pumpAndSettle();
}

Future<void> enter(WidgetTester tester, String label, String text) async {
  final field = find.widgetWithText(TextFormField, label);
  await tester.ensureVisible(field);
  await tester.enterText(field, text);
  await tester.pump();
}

bool hasFocus(WidgetTester tester, String label) => tester
    .widget<EditableText>(
      find.descendant(
        of: find.widgetWithText(TextFormField, label),
        matching: find.byType(EditableText),
      ),
    )
    .focusNode
    .hasFocus;

class Harness {
  Harness(this.tester);
  final WidgetTester tester;
  final server = ProviderServer();
  final client = ClientContext(storage: MemoryStorage(), allowHttp: true);
  final _old = HttpOverrides.current;

  /// Signs in, enters Workspace Alpha and opens the Model providers list.
  Future<void> start({Size? size, double textScale = 1}) async {
    HttpOverrides.global = null;
    tester.view.physicalSize = size ?? const Size(800, 3000);
    tester.view.devicePixelRatio = 1;
    tester.platformDispatcher.textScaleFactorTestValue = textScale;
    addTearDown(() async {
      tester.platformDispatcher.clearTextScaleFactorTestValue();
      tester.view.resetPhysicalSize();
      tester.view.resetDevicePixelRatio();
      client.dispose();
      await server.close();
      HttpOverrides.global = _old;
    });
    await tester.pumpWidget(ButterApp(context: client));
    await tester.runAsync(server.start);
    await tester.runAsync(() => client.setServer(server.url));
    await tester.runAsync(() => client.login('alice', 'password'));
    await tester.pumpAndSettle();
    await tapRequest(tester, client, find.text('Alpha'));
  }

  Future<void> openList() async {
    await tapRequest(tester, client, find.text('Model providers'));
  }

  Future<void> openDetail(String name) async {
    await tapRequest(tester, client, find.text(name));
  }
}

void main() {
  testWidgets('empty list explains itself; create validates, then saves', (
    tester,
  ) async {
    final h = Harness(tester);
    await h.start();
    await h.openList();
    expect(find.text('No model providers yet'), findsOneWidget);
    expect(find.text('Add provider'), findsOneWidget);

    await tapScreen(tester, find.text('Add provider'));
    await tapScreen(tester, find.text('Create provider'));
    expect(find.text('Enter a name for this provider.'), findsOneWidget);
    expect(find.text('Enter the model ID the provider uses.'), findsOneWidget);
    expect(hasFocus(tester, 'Name'), isTrue, reason: 'first invalid field');
    expect(h.server.writes, 0);

    await enter(tester, 'Name', 'local');
    await enter(tester, 'Base URL', 'ftp://x');
    await enter(tester, 'Model ID', 'llama3');
    await enter(tester, 'Context window (tokens)', '4294967296');
    await tapScreen(tester, find.text('Create provider'));
    expect(find.textContaining('absolute http or https address'), findsOne);
    expect(
      find.textContaining('whole number from 0 to 4,294,967,295'),
      findsOne,
    );
    expect(hasFocus(tester, 'Base URL'), isTrue);
    expect(h.server.writes, 0);

    await enter(tester, 'Base URL', 'http://localhost:11434/v1');
    await enter(tester, 'Context window (tokens)', '0');
    await enter(tester, 'Alias', ' local-llama ');
    await enter(tester, 'API key', 'sk-typed-secret');
    await tapRequest(tester, h.client, find.text('Create provider'));

    final sent = CreateModelProviderRequest.fromBuffer(
      h.server.lastBody('CreateModelProvider'),
    ).modelProvider;
    expect(sent.name, 'local');
    expect(sent.type, 'openai');
    expect(sent.apiKey, 'sk-typed-secret');
    expect(sent.baseUrl, 'http://localhost:11434/v1');
    expect(sent.models.single.name, 'llama3');
    expect(sent.models.single.alias, 'local-llama');
    expect(sent.models.single.contextWindowTokens, 0);

    // Back on the list: the new provider, never its key.
    expect(find.text('Model providers'), findsWidgets);
    expect(find.text('local'), findsOneWidget);
    expect(find.textContaining('API key set'), findsOneWidget);
    expect(find.text('local saved.'), findsOneWidget);
    expect(find.textContaining('sk-typed-secret'), findsNothing);
    expect(tester.takeException(), isNull);
  });

  testWidgets('detail shows every field but never the key', (tester) async {
    final h = Harness(tester);
    h.server.providers['alpha']!['one'] = providerWithUnknown('one');
    await h.start();
    await h.openList();
    expect(find.textContaining('OpenAI · 3 models'), findsOneWidget);
    expect(find.textContaining('API key set'), findsOneWidget);
    expect(find.textContaining('sk-secret-value'), findsNothing);
    await h.openDetail('one');
    expect(find.text('Type: OpenAI'), findsOneWidget);
    expect(find.text('Base URL: https://llm.example/v1'), findsOneWidget);
    expect(find.text('API key: Set'), findsOneWidget);
    expect(find.text('Model ID: gpt-4o'), findsOneWidget);
    expect(find.text('Alias: fast'), findsOneWidget);
    expect(find.text('Context window: 128,000 tokens'), findsOneWidget);
    expect(find.text('Alias: None'), findsOneWidget);
    expect(find.textContaining('Not set. The server uses'), findsWidgets);
    expect(find.textContaining('how many tokens of input'), findsOneWidget);
    expect(
      find.textContaining('not the maximum output tokens'),
      findsOneWidget,
    );
    expect(find.textContaining('no setting for output length'), findsOneWidget);
    expect(find.textContaining('max_tokens'), findsNothing);
    expect(find.textContaining('sk-secret-value'), findsNothing);
  });

  testWidgets('editing keeps, replaces or clears the key as asked', (
    tester,
  ) async {
    final h = Harness(tester);
    h.server.providers['alpha']!['one'] = providerWithUnknown('one');
    await h.start();
    await h.openList();
    await h.openDetail('one');

    // Keep (the default): only the base URL changes; the key and every
    // unknown field go back unchanged.
    await tapScreen(tester, find.text('Edit provider'));
    expect(
      find.text(
        'A key is set for this provider. This app never shows saved keys.',
      ),
      findsOneWidget,
    );
    expect(find.textContaining('sk-secret-value'), findsNothing);
    expect(
      tester
          .widget<TextFormField>(find.widgetWithText(TextFormField, 'Name'))
          .controller!
          .text,
      'one',
    );
    await enter(tester, 'Base URL', 'https://changed.example/v1');
    await tapRequest(tester, h.client, find.text('Save changes'));
    var stored = h.server.providers['alpha']!['one']!;
    expect(stored.apiKey, 'sk-secret-value');
    expect(stored.baseUrl, 'https://changed.example/v1');
    expect(unknownVarint(stored, 99), 7);
    expect([for (final m in stored.models) unknownVarint(m, 15)], [1, 2, 3]);
    expect(find.text('Base URL: https://changed.example/v1'), findsOneWidget);

    // Replace needs a value; nothing is sent without one.
    await tapScreen(tester, find.text('Edit provider'));
    await tapScreen(tester, find.text('Replace key'));
    final writes = h.server.writes;
    await tapScreen(tester, find.text('Save changes'));
    expect(find.text('Enter the new API key.'), findsOneWidget);
    expect(hasFocus(tester, 'New API key'), isTrue);
    expect(h.server.writes, writes);
    expect(
      tester
          .widget<TextField>(
            find.descendant(
              of: find.widgetWithText(TextFormField, 'New API key'),
              matching: find.byType(TextField),
            ),
          )
          .obscureText,
      isTrue,
    );
    await enter(tester, 'New API key', 'sk-replacement');
    await tapRequest(tester, h.client, find.text('Save changes'));
    stored = h.server.providers['alpha']!['one']!;
    expect(stored.apiKey, 'sk-replacement');
    expect(stored.models, hasLength(3));
    expect(find.textContaining('sk-replacement'), findsNothing);

    // Clear asks first; cancelling writes nothing.
    await tapScreen(tester, find.text('Edit provider'));
    await tapScreen(tester, find.text('Remove key'));
    final before = h.server.writes;
    await tapReal(tester, find.text('Save changes'));
    expect(find.text('Remove the API key?'), findsOneWidget);
    await tapScreen(tester, find.text('Cancel'));
    expect(h.server.writes, before);
    await tapReal(tester, find.text('Save changes'));
    await tapScreen(tester, find.text('Remove key and save'));
    await settle(tester, h.client, h.server, before + 1);
    expect(h.server.providers['alpha']!['one']!.apiKey, '');
    expect(find.text('API key: Not set'), findsOneWidget);

    // Without a key there is nothing to remove.
    await tapScreen(tester, find.text('Edit provider'));
    expect(find.text('No key is set for this provider.'), findsOneWidget);
    expect(find.text('Keep without a key'), findsOneWidget);
    expect(find.text('Set a key'), findsOneWidget);
    expect(find.text('Remove key'), findsNothing);
    expect(tester.takeException(), isNull);
  });

  testWidgets('removing a model row keeps every other row\'s own fields', (
    tester,
  ) async {
    final h = Harness(tester);
    h.server.providers['alpha']!['one'] = providerWithUnknown('one');
    await h.start();
    await h.openList();
    await h.openDetail('one');
    await tapScreen(tester, find.text('Edit provider'));
    await tapScreen(tester, find.byTooltip('Remove model gpt-mini'));
    expect(find.text('Model ID'), findsNWidgets(2));
    await tapRequest(tester, h.client, find.text('Save changes'));
    final stored = h.server.providers['alpha']!['one']!;
    expect([for (final m in stored.models) m.name], ['gpt-4o', 'gpt-nano']);
    expect([for (final m in stored.models) unknownVarint(m, 15)], [1, 3]);
    expect(stored.models[0].contextWindowTokens, 128000);
    expect(stored.models[1].alias, 'tiny');

    // A provider cannot be saved without any model, as on the web.
    await tapScreen(tester, find.text('Edit provider'));
    await tapScreen(tester, find.byTooltip('Remove model gpt-4o'));
    await tapScreen(tester, find.byTooltip('Remove model gpt-nano'));
    final writes = h.server.writes;
    await tapScreen(tester, find.text('Save changes'));
    expect(
      find.text('Add at least one model to this provider.'),
      findsOneWidget,
    );
    expect(h.server.writes, writes);
  });

  testWidgets('delete confirms, shows the server\'s reason, then deletes', (
    tester,
  ) async {
    final h = Harness(tester);
    h.server.providers['alpha']!['one'] = providerWithUnknown('one');
    await h.start();
    await h.openList();
    await h.openDetail('one');
    await tapReal(tester, find.text('Delete provider'));
    expect(find.text('Delete one?'), findsOneWidget);
    expect(find.textContaining('may stop working'), findsOneWidget);
    await tapScreen(tester, find.text('Cancel'));
    expect(h.server.writes, 0);

    h.server.refusal = (
      412,
      'failed_precondition',
      'models are referenced by telegram destinations: fast (dest-1)',
    );
    await tapReal(tester, find.text('Delete provider'));
    await tapScreen(tester, find.text('Delete provider').last);
    await settle(tester, h.client, h.server, 1);
    expect(find.textContaining('fast (dest-1)'), findsOneWidget);
    expect(find.text('Model provider'), findsOneWidget, reason: 'still open');
    expect(h.server.writes, 1, reason: 'a refused delete is not retried');

    h.server.refusal = null;
    await tapReal(tester, find.text('Delete provider'));
    await tapScreen(tester, find.text('Delete provider').last);
    await settle(tester, h.client, h.server, 2);
    expect(h.server.providers['alpha'], isEmpty);
    expect(find.text('No model providers yet'), findsOneWidget);
    expect(h.server.writes, 2);
  });

  testWidgets('leaving a changed form asks first and writes nothing', (
    tester,
  ) async {
    final h = Harness(tester);
    await h.start();
    await h.openList();
    await tapScreen(tester, find.text('Add provider'));
    await enter(tester, 'Name', 'draft');
    await tester.pageBack();
    await tester.pumpAndSettle();
    expect(find.text('Discard unsaved changes?'), findsOneWidget);
    await tapScreen(tester, find.text('Cancel'));
    expect(find.text('Create provider'), findsOneWidget);
    await tester.pageBack();
    await tester.pumpAndSettle();
    await tapScreen(tester, find.text('Discard'));
    expect(find.text('No model providers yet'), findsOneWidget);
    expect(h.server.writes, 0);
  });

  testWidgets('a context change removes a draft that holds a typed key', (
    tester,
  ) async {
    final h = Harness(tester);
    await h.start();
    await h.openList();
    await tapScreen(tester, find.text('Add provider'));
    await enter(tester, 'API key', 'sk-typed-secret');
    expect(find.byType(EditableText), findsWidgets);
    await tester.runAsync(h.client.logout);
    await tester.pumpAndSettle();
    expect(find.textContaining('Context changed'), findsWidgets);
    expect(find.byType(EditableText), findsNothing);
    expect(find.textContaining('sk-typed-secret'), findsNothing);
    expect(h.client.modelProviders, isEmpty);
  });

  testWidgets('switching Workspace closes provider pages', (tester) async {
    final h = Harness(tester);
    h.server.providers['alpha']!['one'] = providerWithUnknown('one');
    await h.start();
    await h.openList();
    await h.openDetail('one');
    await tester.runAsync(() => h.client.selectWorkspace('beta'));
    await tester.pumpAndSettle();
    expect(find.text('one'), findsNothing);
    expect(find.textContaining('Context changed'), findsWidgets);
  });

  testWidgets('phone width at 2x text: readable, labelled, no overflow', (
    tester,
  ) async {
    final h = Harness(tester);
    h.server.providers['alpha']!['one'] = providerWithUnknown('one');
    final semantics = tester.ensureSemantics();
    await h.start(size: const Size(390, 844), textScale: 2);
    await tester.ensureVisible(find.text('Model providers'));
    await tester.pumpAndSettle();
    await tapRequest(tester, h.client, find.text('Model providers'));
    await tester.scrollUntilVisible(
      find.text('one'),
      200,
      scrollable: find.byType(Scrollable).first,
    );
    await tapRequest(tester, h.client, find.text('one'));
    await tapScreen(tester, find.text('Edit provider'));
    final scrollable = find.byType(Scrollable).first;
    await tester.scrollUntilVisible(
      find.byTooltip('Remove model gpt-4o'),
      300,
      scrollable: scrollable,
    );
    expect(
      tester.getSemantics(find.byTooltip('Remove model gpt-4o')).tooltip,
      'Remove model gpt-4o',
    );
    await tester.scrollUntilVisible(
      find.text('Save changes'),
      300,
      scrollable: scrollable,
    );
    await tester.scrollUntilVisible(
      find.text('Replace key'),
      -300,
      scrollable: scrollable,
    );
    expect(
      tester.getSemantics(find.text('Replace key')).label,
      contains('Replace key'),
    );
    expect(tester.takeException(), isNull);
    semantics.dispose();
  });
}
