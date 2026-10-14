import 'dart:io';

import 'package:butter_mobile/app/client_context.dart';
import 'package:butter_mobile/features/agents/agent_pages.dart';
import 'package:butter_mobile/core/management_page.dart';
import 'package:butter_mobile/gen/agents/v1/agent.pb.dart';
import 'package:butter_mobile/gen/agents/v1/agent_service.pb.dart';
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';

import 'support/agent_server.dart';
import 'support/memory_storage.dart';

Future<({ClientContext client, AgentServer server})> editor(
  WidgetTester tester, {
  Agent? base,
  void Function(AgentServer)? seed,
}) async {
  final old = HttpOverrides.current;
  HttpOverrides.global = null;
  final server = AgentServer();
  seed?.call(server);
  late ClientContext client;
  addTearDown(() async {
    client.dispose();
    await server.close();
    HttpOverrides.global = old;
  });
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
  await tester.pumpWidget(
    MaterialApp(
      home: AgentFormPage(client: client, base: base),
    ),
  );
  return (client: client, server: server);
}

Future<void> visible(WidgetTester tester, Finder finder) async {
  FocusManager.instance.primaryFocus?.unfocus();
  await tester.pumpAndSettle();
  await tester.scrollUntilVisible(
    finder,
    300,
    scrollable: find.byType(Scrollable).first,
  );
  await tester.pumpAndSettle();
}

Future<void> saveForm(WidgetTester tester, ClientContext client) async {
  final save = find.byKey(const Key('save-agent'));
  await visible(tester, save);
  await tester.runAsync(() async {
    await tester.tap(save);
    for (var i = 0; i < 200 && client.busy; i++) {
      await Future<void>.delayed(const Duration(milliseconds: 10));
    }
  });
  expect(client.busy, isFalse);
  await tester.pumpAndSettle();
}

void main() {
  testWidgets(
    'Agent can associate existing Skills without standalone Skill management',
    (tester) async {
      final base = Agent(
        agentId: 'stable',
        name: 'runtime',
        type: AgentType.AGENT_TYPE_LLM,
        config: AgentConfig(skills: ['missing-but-retained']),
      );
      final h = await editor(tester, base: base);
      final skill = find.byKey(const ValueKey('skill-writing'));
      await visible(tester, skill);
      await tester.tap(skill);
      await tester.pumpAndSettle();
      await saveForm(tester, h.client);
      final sent = UpdateAgentRequest.fromBuffer(
        h.server.calls.lastWhere((c) => c.method == 'UpdateAgent').body,
      ).agent;
      expect(sent.config.skills, ['missing-but-retained', 'writing']);
      expect(
        h.server.requests
            .where((r) => r.path.contains('SkillService'))
            .map((r) => r.path.split('/').last),
        ['ListSkills'],
      );
    },
  );

  testWidgets(
    'editing a label keeps unknown include/guard enum and unedited nested configuration',
    (tester) async {
      final base = Agent(
        agentId: 'stable',
        name: 'runtime',
        type: AgentType.AGENT_TYPE_LLM,
        config: AgentConfig(
          skills: ['existing'],
          outputSchemaJson: '{"type":"object"}',
          contextGuard: ContextGuardConfig()..mergeFromBuffer([8, 99]),
        )..mergeFromBuffer([0xc8, 0x01, 98]),
      );
      final h = await editor(tester, base: base);
      final display = find.byKey(const Key('display-name'));
      await visible(tester, display);
      await tester.enterText(display, 'Renamed');
      expect(
        find.text('Include contents: unknown value 98 retained unchanged'),
        findsOneWidget,
      );
      expect(
        find.text('ContextGuard: unknown strategy 99 retained unchanged'),
        findsOneWidget,
      );
      await saveForm(tester, h.client);
      final sent = UpdateAgentRequest.fromBuffer(
        h.server.calls.lastWhere((c) => c.method == 'UpdateAgent').body,
      ).agent;
      expect(sent.displayName, 'Renamed');
      expect(sent.config.writeToBuffer(), base.config.writeToBuffer());
    },
  );

  testWidgets(
    'Loop edits reorder child IDs without editing independent children',
    (tester) async {
      final base = Agent(
        agentId: 'loop',
        name: 'runtime',
        type: AgentType.AGENT_TYPE_LOOP,
        childAgentIds: ['child-b', 'child-a'],
        config: AgentConfig(maxIterations: 3),
      );
      final h = await editor(
        tester,
        base: base,
        seed: (s) {
          for (final id in ['child-a', 'child-b']) {
            s.agents['alpha']![id] = Agent(
              agentId: id,
              name: id,
              displayName: 'Same label',
              type: AgentType.AGENT_TYPE_LLM,
            );
          }
        },
      );
      expect(find.byKey(const Key('instruction')), findsNothing);
      expect(find.byKey(const Key('guard-mode')), findsNothing);
      final iterations = find.byKey(const Key('max-iterations'));
      await visible(tester, iterations);
      await tester.enterText(iterations, '5');
      final move = find.byTooltip('Move child-a up');
      await visible(tester, move);
      await tester.tap(move);
      await tester.pumpAndSettle();
      await saveForm(tester, h.client);
      final sent = UpdateAgentRequest.fromBuffer(
        h.server.calls.lastWhere((c) => c.method == 'UpdateAgent').body,
      ).agent;
      expect(sent.childAgentIds, ['child-a', 'child-b']);
      expect(sent.config.maxIterations, 5);
      expect(h.server.agents['alpha']!['child-a']!.displayName, 'Same label');
    },
  );

  testWidgets('Workspace change unmounts the old Agent draft route', (
    tester,
  ) async {
    final h = await editor(tester);
    await tester.pumpWidget(
      MaterialApp(
        home: Builder(
          builder: (context) => Scaffold(
            body: TextButton(
              onPressed: () => openManagementPage(
                context,
                h.client,
                AgentFormPage(client: h.client),
              ),
              child: const Text('Open editor'),
            ),
          ),
        ),
      ),
    );
    await tester.tap(find.text('Open editor'));
    await tester.pumpAndSettle();
    await tester.enterText(find.byKey(const Key('agent-id')), 'private-draft');
    await tester.runAsync(() => h.client.selectWorkspace('beta'));
    await tester.pumpAndSettle();
    expect(
      find.text('Context changed. Return to the home screen.'),
      findsOneWidget,
    );
    expect(find.byKey(const Key('agent-id')), findsNothing);
  });

  testWidgets(
    'phone layout scrolls all structured fields at twice system text size',
    (tester) async {
      tester.view.physicalSize = const Size(390, 844);
      tester.view.devicePixelRatio = 1;
      tester.platformDispatcher.textScaleFactorTestValue = 2;
      addTearDown(() {
        tester.view.resetPhysicalSize();
        tester.view.resetDevicePixelRatio();
        tester.platformDispatcher.clearTextScaleFactorTestValue();
      });
      final h = await editor(tester);
      await visible(tester, find.byKey(const Key('save-agent')));
      expect(tester.takeException(), isNull);
      expect(h.client.agentEditorReady, isTrue);
    },
  );

  testWidgets(
    'model choices update the visible model field and invalid off-screen fields prevent writes',
    (tester) async {
      tester.view.physicalSize = const Size(800, 3000);
      tester.view.devicePixelRatio = 1;
      addTearDown(() {
        tester.view.resetPhysicalSize();
        tester.view.resetDevicePixelRatio();
      });
      final h = await editor(tester);
      await tester.enterText(find.byKey(const Key('agent-id')), 'new-llm');
      await tester.enterText(find.byKey(const Key('agent-name')), 'runtime');
      final choice = find.byKey(const ValueKey('model-choice-'));
      await visible(tester, choice);
      await tester.tap(choice);
      await tester.pumpAndSettle();
      await tester.tap(find.text('fast').last);
      await tester.pumpAndSettle();
      final model = find.byKey(const Key('model'));
      await visible(tester, model);
      expect(
        tester
            .widget<EditableText>(
              find.descendant(of: model, matching: find.byType(EditableText)),
            )
            .controller
            .text,
        'fast',
      );
      final schema = find.byKey(const Key('output-schema'));
      await visible(tester, schema);
      await tester.enterText(schema, '{bad json');
      final save = find.byKey(const Key('save-agent'));
      await visible(tester, save);
      await tester.runAsync(() => tester.tap(save));
      await tester.pumpAndSettle();
      expect(
        h.server.calls.where((c) => c.method.startsWith('Create')),
        isEmpty,
      );
      expect(
        find.text('Enter a JSON schema object, or leave this empty.'),
        findsOneWidget,
      );
    },
  );

  testWidgets(
    'LLM editor saves instructions and uses only the Threshold override',
    (tester) async {
      final oldOverrides = HttpOverrides.current;
      HttpOverrides.global = null;
      addTearDown(() => HttpOverrides.global = oldOverrides);
      final server = AgentServer();
      await tester.runAsync(server.start);
      addTearDown(server.close);
      final client = ClientContext(storage: MemoryStorage(), allowHttp: true);
      addTearDown(client.dispose);
      await tester.runAsync(() async {
        await client.setServer(server.url);
        await client.login('alice', 'password');
        await client.selectWorkspace('alpha');
        await client.prepareAgentEditor();
      });
      expect(client.error, isNull);
      expect(client.agentEditorReady, isTrue);
      await tester.pumpWidget(MaterialApp(home: AgentFormPage(client: client)));
      await tester.enterText(find.byKey(const Key('agent-id')), 'new-llm');
      await tester.enterText(
        find.byKey(const Key('agent-name')),
        'runtime-name',
      );
      final instruction = find.byKey(const Key('instruction'));
      await tester.scrollUntilVisible(
        instruction,
        250,
        scrollable: find.byType(Scrollable).first,
      );
      await tester.enterText(instruction, 'Follow my instruction');
      final guard = find.byKey(const Key('guard-mode'));
      await tester.scrollUntilVisible(
        guard,
        250,
        scrollable: find.byType(Scrollable).first,
      );
      await tester.tap(guard);
      await tester.pumpAndSettle();
      await tester.tap(find.text('Threshold').last);
      await tester.pumpAndSettle();
      final override = find.byKey(const Key('guard-tokens'));
      await tester.ensureVisible(override);
      await tester.enterText(override, '64000');
      expect(find.byKey(const Key('guard-turns')), findsNothing);
      FocusManager.instance.primaryFocus?.unfocus();
      await tester.pumpAndSettle();
      final save = find.byKey(const Key('save-agent'));
      await tester.scrollUntilVisible(
        save,
        300,
        scrollable: find.byType(Scrollable).first,
      );
      await tester.pumpAndSettle();
      await tester.runAsync(() async {
        await tester.tap(save);
        for (var i = 0; i < 100 && client.busy; i++) {
          await Future<void>.delayed(const Duration(milliseconds: 10));
        }
      });
      await tester.pumpAndSettle();
      final call = server.calls.lastWhere((c) => c.method == 'CreateAgent');
      final saved = CreateAgentRequest.fromBuffer(call.body).agent;
      expect(saved.agentId, 'new-llm');
      expect(saved.config.instruction, 'Follow my instruction');
      expect(saved.config.contextGuard.maxTokens, 64000);
      expect(saved.config.contextGuard.hasMaxTurns(), isFalse);
      expect(saved.type, AgentType.AGENT_TYPE_LLM);
    },
  );
}
