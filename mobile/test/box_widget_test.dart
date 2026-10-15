import 'dart:io';

import 'package:butter_mobile/app/butter_app.dart';
import 'package:butter_mobile/app/client_context.dart';
import 'package:butter_mobile/features/agents/agent_pages.dart';
import 'package:butter_mobile/features/agents/leaf_draft.dart';
import 'package:butter_mobile/features/butterboxes/box_pages.dart';
import 'package:butter_mobile/gen/agents/v1/agent.pb.dart';
import 'package:butter_mobile/gen/agents/v1/agent_service.pb.dart';
import 'package:butter_mobile/gen/agents/v1/butterbox.pb.dart';
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';

import 'support/box_server.dart';
import 'support/memory_storage.dart';

Future<({ClientContext client, BoxServer server})> signedIn(
  WidgetTester tester, {
  void Function(BoxServer)? seed,
}) async {
  final old = HttpOverrides.current;
  HttpOverrides.global = null;
  final server = BoxServer();
  seed?.call(server);
  late ClientContext client;
  addTearDown(() async {
    client.dispose();
    await server.close();
    HttpOverrides.global = old;
  });
  tester.view.physicalSize = const Size(800, 3000);
  tester.view.devicePixelRatio = 1;
  addTearDown(tester.view.reset);
  await tester.pumpWidget(const MaterialApp(home: SizedBox()));
  await tester.runAsync(() async {
    await server.start();
    client = ClientContext(storage: MemoryStorage(), allowHttp: true);
    await client.setServer(server.url);
    await client.login('alice', 'password');
    await client.selectWorkspace('alpha');
    await client.prepareAgentEditor();
  });
  expect(client.error, isNull);
  return (client: client, server: server);
}

/// Lets real HTTP complete while the widget tree keeps pumping.
Future<void> waitFor(WidgetTester tester, bool Function() done) async {
  for (var i = 0; i < 300 && !done(); i++) {
    await tester.runAsync(
      () => Future<void>.delayed(const Duration(milliseconds: 10)),
    );
    await tester.pump();
  }
  expect(done(), isTrue, reason: 'the request did not finish');
  await tester.pumpAndSettle();
}

Future<void> tapReal(WidgetTester tester, Finder finder) async {
  await tester.ensureVisible(finder);
  await tester.runAsync(() => tester.tap(finder));
  // A request is in flight and its progress bar never settles: callers wait
  // for the outcome with waitFor.
  await tester.pump();
}

/// Taps a dialog's action. What follows is a request, whose progress bar never
/// settles, so only the dialog's own closing is pumped.
Future<void> tapDialog(WidgetTester tester, Finder finder) async {
  await tester.tap(finder);
  await tester.pump();
  await tester.pump(const Duration(milliseconds: 400));
}

Future<void> tapScreen(WidgetTester tester, Finder finder) async {
  await tester.ensureVisible(finder);
  await tester.tap(finder);
  await tester.pumpAndSettle();
}

Future<void> enter(WidgetTester tester, Key key, String text) async {
  final field = find.byKey(key);
  await tester.ensureVisible(field);
  await tester.enterText(field, text);
  await tester.pump();
}

Future<void> choose(WidgetTester tester, Finder dropdown, String label) async {
  await tester.ensureVisible(dropdown);
  await tester.tap(dropdown);
  await tester.pumpAndSettle();
  await tester.tap(find.text(label).last);
  await tester.pumpAndSettle();
}

Future<void> openList(WidgetTester tester, ClientContext client) async {
  await tester.pumpWidget(MaterialApp(home: ButterBoxListPage(client: client)));
  await tester.runAsync(() => client.boxes!.loadBoxes());
  await tester.pumpAndSettle();
}

Future<void> openDetail(
  WidgetTester tester,
  ClientContext client,
  String id,
) async {
  // Built in the real zone so the page's own first read runs on real timers.
  await tester.runAsync(
    () => tester.pumpWidget(
      MaterialApp(
        home: ButterBoxDetailPage(client: client, id: id),
      ),
    ),
  );
  await waitFor(tester, () => client.boxes!.box?.id == id);
}

void main() {
  testWidgets('ButterBoxes opens from the Workspace screen and is scoped to '
      'it', (tester) async {
    final old = HttpOverrides.current;
    HttpOverrides.global = null;
    final server = BoxServer();
    server.seed('alpha', 'a1', 'Alpha laptop');
    server.seed('beta', 'b1', 'Beta laptop');
    final client = ClientContext(storage: MemoryStorage(), allowHttp: true);
    addTearDown(() async {
      client.dispose();
      await server.close();
      HttpOverrides.global = old;
    });
    tester.view.physicalSize = const Size(800, 3000);
    tester.view.devicePixelRatio = 1;
    addTearDown(tester.view.reset);
    await tester.pumpWidget(ButterApp(context: client));
    await tester.runAsync(server.start);
    await tester.runAsync(() => client.setServer(server.url));
    await tester.runAsync(() => client.login('alice', 'password'));
    await tester.pumpAndSettle();
    await tester.runAsync(() async {
      await tester.tap(find.text('Alpha'));
      for (var i = 0; i < 200 && client.busy; i++) {
        await Future<void>.delayed(const Duration(milliseconds: 10));
      }
    });
    await tester.pumpAndSettle();

    await tapReal(tester, find.byKey(const Key('open-boxes')));
    await waitFor(tester, () => client.boxes!.listState.name == 'ready');
    expect(find.text('Alpha laptop'), findsOneWidget);
    expect(find.text('Beta laptop'), findsNothing);
    expect(server.boxCalls.map((c) => c.workspace).toSet(), {'alpha'});
  });

  group('ButterBox pages', () {
    testWidgets('registering a box sends its token once and never shows it', (
      tester,
    ) async {
      final h = await signedIn(tester);
      await openList(tester, h.client);
      expect(find.text('No ButterBoxes yet'), findsOneWidget);

      await tapScreen(tester, find.byKey(const Key('register-box')));
      await enter(tester, const Key('box-name'), 'Laptop');
      await enter(tester, const Key('box-url'), 'https://box.example.com');
      await enter(tester, const Key('box-token'), 'tok-very-secret');
      await tapReal(tester, find.byKey(const Key('save-box')));
      await waitFor(tester, () => h.server.boxWrites == 1);
      await tester.pumpAndSettle();

      final stored = h.server.boxes['alpha']!.values.single;
      expect(stored.name, 'Laptop');
      expect(stored.baseUrl, 'https://box.example.com');
      expect(h.server.secrets['alpha']![stored.id], 'tok-very-secret');
      expect(find.byKey(Key('box-${stored.id}')), findsOneWidget);
      expect(find.textContaining('Access token set'), findsOneWidget);
      expect(find.textContaining('tok-very-secret'), findsNothing);
    });

    testWidgets('an invalid name or URL is refused before any request', (
      tester,
    ) async {
      final h = await signedIn(tester);
      await openList(tester, h.client);
      await tapScreen(tester, find.byKey(const Key('register-box')));
      await enter(tester, const Key('box-name'), 'Laptop');
      await enter(tester, const Key('box-url'), 'ftp://box');
      await tapScreen(tester, find.byKey(const Key('save-box')));
      expect(find.textContaining('http'), findsWidgets);
      expect(h.server.boxWrites, 0);

      await enter(tester, const Key('box-name'), '');
      await enter(tester, const Key('box-url'), 'https://box.test');
      await tapScreen(tester, find.byKey(const Key('save-box')));
      expect(h.server.boxWrites, 0);
    });

    testWidgets('a duplicate name is explained and nothing is repeated', (
      tester,
    ) async {
      final h = await signedIn(
        tester,
        seed: (s) => s.seed('alpha', 'a1', 'Laptop'),
      );
      await openList(tester, h.client);
      await tapScreen(tester, find.byKey(const Key('register-box')));
      await enter(tester, const Key('box-name'), 'Laptop');
      await enter(tester, const Key('box-url'), 'https://box.test');
      await tapReal(tester, find.byKey(const Key('save-box')));
      await waitFor(tester, () => h.server.boxWrites == 1);
      expect(find.byKey(const Key('box-write-error')), findsOneWidget);
      expect(h.server.boxWrites, 1);
    });

    testWidgets('the connection test reports reachable and unreachable', (
      tester,
    ) async {
      final h = await signedIn(
        tester,
        seed: (s) => s.seed('alpha', 'a1', 'Laptop'),
      );
      await openDetail(tester, h.client, 'a1');
      expect(find.text('Connection not tested'), findsOneWidget);

      await tapReal(tester, find.byKey(const Key('probe-box')));
      await waitFor(
        tester,
        () => h.client.boxes!.probes['a1']?.state.name == 'reachable',
      );
      expect(find.text('Reachable · 2 active sessions'), findsOneWidget);

      h.server.status = GetButterBoxStatusResponse(
        reachable: false,
        error: 'unavailable: connection refused token=sk-SECRET',
      );
      await tapReal(tester, find.byKey(const Key('probe-box')));
      await waitFor(
        tester,
        () => h.client.boxes!.probes['a1']?.state.name == 'unreachable',
      );
      expect(
        find.text(
          'Unreachable: The server could not connect to the Box. Check its URL and that it is running.',
        ),
        findsOneWidget,
      );
      expect(find.textContaining('sk-SECRET'), findsNothing);
      expect(h.server.boxWrites, 0);
    });

    testWidgets('replacing and clearing the token are separate explicit '
        'actions', (tester) async {
      final h = await signedIn(
        tester,
        seed: (s) => s.seed('alpha', 'a1', 'Laptop'),
      );
      await openDetail(tester, h.client, 'a1');
      expect(find.textContaining('Access token: set'), findsOneWidget);

      await tapReal(tester, find.byKey(const Key('replace-token')));
      // The save action is off while the field is empty: no blank can erase.
      FilledButton save() =>
          tester.widget<FilledButton>(find.byKey(const Key('save-token')));
      expect(save().onPressed, isNull);
      await tester.enterText(find.byKey(const Key('token-field')), 'rotated');
      await tester.pump();
      expect(save().onPressed, isNotNull);
      await tapDialog(tester, find.byKey(const Key('save-token')));
      await waitFor(tester, () => h.server.boxWrites == 1);
      expect(h.server.secrets['alpha']!['a1'], 'rotated');
      expect(find.textContaining('rotated'), findsNothing);

      await tapReal(tester, find.byKey(const Key('clear-token')));
      expect(find.text('Clear access token?'), findsOneWidget);
      await tapDialog(tester, find.text('Clear token'));
      await waitFor(tester, () => h.server.boxWrites == 2);
      expect(h.server.secrets['alpha']!.containsKey('a1'), isFalse);
      expect(find.textContaining('Access token: none'), findsOneWidget);
      expect(find.byKey(const Key('clear-token')), findsNothing);
    });

    testWidgets('cancelling a destructive confirmation sends nothing', (
      tester,
    ) async {
      final h = await signedIn(
        tester,
        seed: (s) => s.seed('alpha', 'a1', 'Laptop'),
      );
      await openDetail(tester, h.client, 'a1');
      await tapScreen(tester, find.byKey(const Key('delete-box')));
      await tapScreen(tester, find.text('Cancel'));
      await tapScreen(tester, find.byKey(const Key('clear-token')));
      await tapScreen(tester, find.text('Cancel'));
      expect(h.server.boxWrites, 0);
    });

    testWidgets('deleting a box that Agents use names them and keeps it', (
      tester,
    ) async {
      final h = await signedIn(
        tester,
        seed: (s) {
          s.seed('alpha', 'a1', 'Laptop');
          s.referencedBy['a1'] = ['coder', 'planner'];
        },
      );
      await openDetail(tester, h.client, 'a1');
      await tapReal(tester, find.byKey(const Key('delete-box')));
      await tapDialog(tester, find.text('Delete').last);
      await waitFor(tester, () => h.server.boxWrites == 1);
      final error = tester.widget<Text>(
        find.byKey(const Key('box-write-error')),
      );
      expect(error.data, allOf(contains('coder'), contains('planner')));
      expect(h.server.boxes['alpha']!.containsKey('a1'), isTrue);
      expect(find.text('Laptop'), findsWidgets);
    });

    testWidgets('editing a box never sends or changes its token', (
      tester,
    ) async {
      final h = await signedIn(
        tester,
        seed: (s) => s.seed('alpha', 'a1', 'Laptop'),
      );
      await openDetail(tester, h.client, 'a1');
      await tapScreen(tester, find.byKey(const Key('edit-box')));
      expect(find.byKey(const Key('box-token')), findsNothing);
      await enter(tester, const Key('box-name'), 'Workstation');
      await tapReal(tester, find.byKey(const Key('save-box')));
      await waitFor(tester, () => h.server.boxWrites == 1);
      expect(h.server.boxes['alpha']!['a1']!.name, 'Workstation');
      expect(h.server.secrets['alpha']!['a1'], 'secret-of-a1');
      expect(
        h.server.boxCalls.where((c) => c.method == 'SetButterBoxToken'),
        isEmpty,
      );
    });

    testWidgets('a lost reply asks for a refresh instead of resending', (
      tester,
    ) async {
      final h = await signedIn(tester);
      await openList(tester, h.client);
      h.server.loseBoxReply = true;
      await tapScreen(tester, find.byKey(const Key('register-box')));
      await enter(tester, const Key('box-name'), 'Laptop');
      await enter(tester, const Key('box-url'), 'https://box.test');
      await tapReal(tester, find.byKey(const Key('save-box')));
      await waitFor(tester, () => h.client.boxes!.unconfirmed);
      expect(find.textContaining('not confirmed'), findsWidgets);
      final save = tester.widget<FilledButton>(
        find.byKey(const Key('save-box')),
      );
      expect(save.onPressed, isNull);
      expect(h.server.boxWrites, 1);
    });
  });

  group('Pi and Cursor Agent editor', () {
    Future<void> openEditor(
      WidgetTester tester,
      ClientContext client, {
      Agent? base,
    }) async {
      await tester.runAsync(() => client.boxes!.loadBoxes());
      await tester.pumpWidget(
        MaterialApp(
          home: AgentFormPage(client: client, base: base),
        ),
      );
      await tester.pumpAndSettle();
    }

    Future<void> identify(WidgetTester tester, String id) async {
      await enter(tester, const Key('agent-id'), id);
      await enter(tester, const Key('agent-name'), id);
    }

    Future<Agent> saveAndRead(
      WidgetTester tester,
      ClientContext client,
      BoxServer server,
      String method,
    ) async {
      final save = find.byKey(const Key('save-agent'));
      await tester.ensureVisible(save);
      await tester.runAsync(() async {
        await tester.tap(save);
        for (var i = 0; i < 300 && client.busy; i++) {
          await Future<void>.delayed(const Duration(milliseconds: 10));
        }
      });
      await tester.pumpAndSettle();
      final body = server.calls.lastWhere((c) => c.method == method).body;
      return method == 'CreateAgent'
          ? CreateAgentRequest.fromBuffer(body).agent
          : UpdateAgentRequest.fromBuffer(body).agent;
    }

    Future<void> pickBox(WidgetTester tester, String name) async {
      await choose(
        tester,
        find.byType(DropdownButtonFormField<String>).first,
        name,
      );
    }

    testWidgets('creating a Pi leaf sends only box-side fields', (
      tester,
    ) async {
      final h = await signedIn(
        tester,
        seed: (s) {
          s.seed('alpha', 'a1', 'Laptop');
          s.seed('alpha', 'a2', 'Off box', enabled: false);
          s.piModels = [
            ButterBoxModel(
              id: 'sonnet',
              provider: 'anthropic',
              name: 'Claude Sonnet',
              reasoning: true,
            ),
          ];
        },
      );
      await openEditor(tester, h.client);
      await identify(tester, 'pi-coder');
      await choose(tester, find.byKey(const Key('agent-type')), 'PI');

      // Butter-side behavior fields do not exist for a leaf.
      expect(find.byKey(const Key('instruction')), findsNothing);
      expect(find.byKey(const ValueKey('skill-writing')), findsNothing);
      expect(find.text('Pi configuration'), findsOneWidget);

      await waitFor(tester, () => h.client.boxes!.listState.name == 'ready');
      await pickBox(tester, 'Laptop');
      await tapReal(tester, find.byKey(const Key('leaf-load-models')));
      await waitFor(
        tester,
        () => h.client.boxes!.piCatalogs['a1']?.state.name == 'ready',
      );
      await choose(
        tester,
        find.widgetWithText(
          DropdownButtonFormField<String>,
          'Model from catalog',
        ),
        'anthropic · Claude Sonnet (sonnet)',
      );
      await choose(
        tester,
        find.widgetWithText(DropdownButtonFormField<String>, 'Thinking level'),
        'high',
      );
      await enter(tester, const Key('leaf-working-dir'), '/srv/repo');
      await choose(
        tester,
        find.widgetWithText(
          DropdownButtonFormField<RunLimitMode>,
          'Maximum run time',
        ),
        'Unlimited',
      );

      final sent = await saveAndRead(tester, h.client, h.server, 'CreateAgent');
      expect(sent.type, AgentType.AGENT_TYPE_PI);
      expect(sent.config.pi.butterboxId, 'a1');
      expect(sent.config.pi.provider, 'anthropic');
      expect(sent.config.pi.model, 'sonnet');
      expect(sent.config.pi.thinkingLevel, 'high');
      expect(sent.config.pi.workingDir, '/srv/repo');
      expect(sent.config.pi.hasMaxRunSeconds(), isTrue);
      expect(sent.config.pi.maxRunSeconds, 0);
      expect(sent.config.instruction, isEmpty);
      expect(sent.config.skills, isEmpty);
      expect(sent.config.mcpServerIds, isEmpty);
      expect(sent.config.hasMemory(), isFalse);
      expect(sent.config.hasResultCards(), isFalse);
      expect(sent.childAgentIds, isEmpty);
    });

    testWidgets(
      'a Cursor leaf takes mode and seconds; default leaves the limit '
      'absent',
      (tester) async {
        final h = await signedIn(
          tester,
          seed: (s) {
            s.seed('alpha', 'a1', 'Laptop');
            s.cursorModels = [
              ButterBoxCursorModel(id: 'composer-2.5', name: 'Composer 2.5'),
            ];
          },
        );
        await openEditor(tester, h.client);
        await identify(tester, 'cursor-coder');
        await choose(tester, find.byKey(const Key('agent-type')), 'CURSOR');
        expect(find.text('Cursor configuration'), findsOneWidget);
        await pickBox(tester, 'Laptop');
        await tapReal(tester, find.byKey(const Key('leaf-load-models')));
        await waitFor(
          tester,
          () => h.client.boxes!.cursorCatalogs['a1']?.state.name == 'ready',
        );
        await choose(
          tester,
          find.widgetWithText(
            DropdownButtonFormField<String>,
            'Model from catalog',
          ),
          'Composer 2.5 (composer-2.5)',
        );
        await choose(
          tester,
          find.widgetWithText(DropdownButtonFormField<String>, 'Mode'),
          'plan',
        );
        var sent = await saveAndRead(tester, h.client, h.server, 'CreateAgent');
        expect(sent.type, AgentType.AGENT_TYPE_CURSOR);
        expect(sent.config.cursor.butterboxId, 'a1');
        expect(sent.config.cursor.model, 'composer-2.5');
        expect(sent.config.cursor.mode, 'plan');
        expect(sent.config.cursor.hasMaxRunSeconds(), isFalse);
        expect(sent.config.hasPi(), isFalse);
      },
    );

    testWidgets('an explicit number of seconds is sent; an invalid one blocks '
        'the save', (tester) async {
      final h = await signedIn(
        tester,
        seed: (s) => s.seed('alpha', 'a1', 'Laptop'),
      );
      await openEditor(tester, h.client);
      await identify(tester, 'pi-timed');
      await choose(tester, find.byKey(const Key('agent-type')), 'PI');
      await pickBox(tester, 'Laptop');
      await choose(
        tester,
        find.widgetWithText(
          DropdownButtonFormField<RunLimitMode>,
          'Maximum run time',
        ),
        'Set seconds',
      );
      await enter(tester, const Key('leaf-run-seconds'), '0');
      await tapScreen(tester, find.byKey(const Key('save-agent')));
      expect(h.server.calls.where((c) => c.method == 'CreateAgent'), isEmpty);
      expect(find.textContaining('whole number of seconds'), findsOneWidget);

      await enter(tester, const Key('leaf-run-seconds'), '120');
      final sent = await saveAndRead(tester, h.client, h.server, 'CreateAgent');
      expect(sent.config.pi.maxRunSeconds, 120);
      expect(sent.config.pi.hasMaxRunSeconds(), isTrue);
    });

    testWidgets('a box is required, and a disabled one is marked', (
      tester,
    ) async {
      final h = await signedIn(
        tester,
        seed: (s) {
          s.seed('alpha', 'a1', 'Laptop');
          s.seed('alpha', 'a2', 'Retired', enabled: false);
        },
      );
      await openEditor(tester, h.client);
      await identify(tester, 'pi-nobox');
      await choose(tester, find.byKey(const Key('agent-type')), 'PI');
      await tapScreen(tester, find.byKey(const Key('save-agent')));
      expect(find.text('Select a ButterBox.'), findsOneWidget);
      expect(h.server.calls.where((c) => c.method == 'CreateAgent'), isEmpty);
      await tester.tap(find.byType(DropdownButtonFormField<String>).first);
      await tester.pumpAndSettle();
      expect(find.text('Retired (disabled)'), findsOneWidget);
    });

    testWidgets('a failed catalog says what to do and typing by hand still '
        'saves', (tester) async {
      final h = await signedIn(
        tester,
        seed: (s) {
          s.seed('alpha', 'a1', 'Laptop');
          s.piRefusal = (429, 'resource_exhausted', 'pi box is at capacity');
        },
      );
      await openEditor(tester, h.client);
      await identify(tester, 'pi-manual');
      await choose(tester, find.byKey(const Key('agent-type')), 'PI');
      await pickBox(tester, 'Laptop');
      await tapReal(tester, find.byKey(const Key('leaf-load-models')));
      await waitFor(
        tester,
        () => h.client.boxes!.piCatalogs['a1']?.state.name == 'failed',
      );
      final text = tester.widget<Text>(
        find.byKey(const Key('leaf-catalog-error')),
      );
      expect(text.data, allOf(contains('capacity'), contains('by hand')));
      await enter(tester, const Key('leaf-provider'), 'openai');
      await enter(tester, const Key('leaf-model'), 'gpt-5');
      final sent = await saveAndRead(tester, h.client, h.server, 'CreateAgent');
      expect(sent.config.pi.provider, 'openai');
      expect(sent.config.pi.model, 'gpt-5');
    });

    testWidgets('editing a leaf keeps an explicit unlimited run and unknown '
        'fields, and repointing changes only the box', (tester) async {
      final pi = PiAgentConfig(
        butterboxId: 'a1',
        workingDir: '/work',
        model: 'sonnet',
        thinkingLevel: 'future-level',
        maxRunSeconds: 0,
      )..mergeFromBuffer([0x98, 0x06, 7]);
      final base = Agent(
        agentId: 'pi-existing',
        name: 'pi-existing',
        type: AgentType.AGENT_TYPE_PI,
        config: AgentConfig(pi: pi),
      );
      final h = await signedIn(
        tester,
        seed: (s) {
          s.seed('alpha', 'a1', 'Laptop');
          s.seed('alpha', 'a2', 'Workstation');
          s.agents['alpha']!['pi-existing'] = base;
        },
      );
      await openEditor(tester, h.client, base: base);
      expect(find.text('Unlimited'), findsOneWidget);
      expect(find.text('future-level (retained)'), findsOneWidget);

      await pickBox(tester, 'Workstation');
      final sent = await saveAndRead(tester, h.client, h.server, 'UpdateAgent');
      expect(sent.config.pi.butterboxId, 'a2');
      expect(sent.config.pi.workingDir, '/work');
      expect(sent.config.pi.thinkingLevel, 'future-level');
      // The model of the old box is not carried to a box that may not have it.
      expect(sent.config.pi.model, isEmpty);
      expect(sent.config.pi.hasMaxRunSeconds(), isTrue);
      expect(sent.config.pi.maxRunSeconds, 0);
      expect(sent.config.pi.unknownFields.isNotEmpty, isTrue);
    });

    testWidgets('an Agent that points at a box missing from this Workspace '
        'stays unchanged unless repointed', (tester) async {
      final base = Agent(
        agentId: 'pi-orphan',
        name: 'pi-orphan',
        type: AgentType.AGENT_TYPE_PI,
        config: AgentConfig(pi: PiAgentConfig(butterboxId: 'gone')),
      );
      final h = await signedIn(
        tester,
        seed: (s) {
          s.seed('alpha', 'a1', 'Laptop');
          s.agents['alpha']!['pi-orphan'] = base;
        },
      );
      await openEditor(tester, h.client, base: base);
      expect(find.textContaining('not in this Workspace'), findsOneWidget);
      final sent = await saveAndRead(tester, h.client, h.server, 'UpdateAgent');
      expect(sent.config.pi.butterboxId, 'gone');
    });
  });
}
