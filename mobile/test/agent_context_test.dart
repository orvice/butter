import 'dart:async';

import 'package:butter_mobile/app/client_context.dart';
import 'package:butter_mobile/gen/agents/v1/agent.pb.dart';
import 'package:butter_mobile/gen/agents/v1/agent_service.pb.dart';
import 'package:butter_mobile/gen/agents/v1/repobinding.pb.dart';
import 'package:flutter_test/flutter_test.dart';

import 'support/agent_server.dart';
import 'support/memory_storage.dart';

Future<ClientContext> signedIn(AgentServer server) async {
  await server.start();
  addTearDown(server.close);
  final client = ClientContext(storage: MemoryStorage(), allowHttp: true);
  addTearDown(client.dispose);
  await client.setServer(server.url);
  await client.login('alice', 'password');
  await client.selectWorkspace('alpha');
  return client;
}

void main() {
  test(
    'bound but not onboarded content remains editable through the DB-owned API',
    () async {
      final server = AgentServer()
        ..bound = true
        ..activeCommitSha = '';
      server.agents['alpha']!['stable'] = Agent(
        agentId: 'stable',
        name: 'runtime',
        type: AgentType.AGENT_TYPE_LLM,
      );
      final client = await signedIn(server);
      await client.loadAgent('stable');
      final before = client.managedAgent!;
      final draft = before.deepCopy();
      draft.ensureConfig().instruction = 'Database owned';
      expect(
        await client.saveAgent(draft, create: false, previous: before),
        isTrue,
      );
      expect(server.calls.last.method, 'UpdateAgent');
    },
  );

  test(
    'late Agent detail and in-flight writes cannot fill another Workspace',
    () async {
      final held = Completer<void>();
      final reached = Completer<void>();
      final server = AgentServer();
      server.agents['alpha']!['same'] = Agent(agentId: 'same', name: 'alpha');
      server.agents['beta']!['same'] = Agent(agentId: 'same', name: 'beta');
      final client = await signedIn(server);
      server.beforeGet = () async {
        reached.complete();
        await held.future;
      };
      final read = client.loadAgent('same');
      await reached.future;
      await client.selectWorkspace('beta');
      held.complete();
      await read;
      expect(client.managedAgent, isNull);
      expect(client.agentChoices, isEmpty);
      server.beforeGet = null;
      await client.loadAgent('same');
      expect(client.managedAgent?.name, 'beta');

      final writeHeld = Completer<void>(), writeReached = Completer<void>();
      server.beforeWrite = () async {
        writeReached.complete();
        await writeHeld.future;
      };
      final writing = client.saveAgent(
        client.managedAgent!.deepCopy()..displayName = 'pending',
        create: false,
      );
      await writeReached.future;
      await client.selectWorkspace('alpha');
      writeHeld.complete();
      expect(await writing, isFalse);
      expect(client.managedAgent, isNull);
      expect(client.agents.single.name, 'alpha');
    },
  );

  test(
    'ambiguous save is never replayed by foreground recovery and expires cleanly',
    () async {
      final server = AgentServer()..loseWriteReply = true;
      final client = await signedIn(server);
      final draft = Agent(
        agentId: 'lost',
        name: 'lost',
        type: AgentType.AGENT_TYPE_LLM,
      );
      expect(await client.saveAgent(draft, create: true), isFalse);
      expect(client.writeNotice, contains('may already have been applied'));
      client.suspend();
      await client.revalidate();
      expect(client.agents.single.agentId, 'lost');
      expect(server.calls.where((c) => c.method == 'CreateAgent').length, 1);
      expect(client.writeNotice, isNotNull);
      await client.loadAgent('lost');
      expect(client.writeNotice, isNull);
      expect(client.managedAgent?.agentId, 'lost');
      await client.prepareAgentEditor();
      expect(client.agentChoices, isNotEmpty);
      server.expired = true;
      await client.loadAgent('lost');
      expect(client.user, isNull);
      expect(client.managedAgent, isNull);
      expect(client.agentChoices, isEmpty);
      expect(client.agentModelChoices, isEmpty);
    },
  );

  test(
    'server validation is specific and safe; permission denial keeps the session and draft',
    () async {
      final server = AgentServer()
        ..refusal = (
          400,
          'invalid_argument',
          'agent "secret": config.context_guard.max_tokens is not supported with the sliding window strategy; sk-secret-should-not-echo',
        );
      final client = await signedIn(server);
      final draft = Agent(agentId: 'draft', name: 'draft');
      expect(await client.saveAgent(draft, create: true), isFalse);
      expect(client.error, contains('Sliding Window uses the entry limit'));
      expect(client.error, isNot(contains('sk-secret')));
      server.refusal = (
        400,
        'invalid_argument',
        'agent "secret": config.context_guard.strategy 99 is not a known strategy',
      );
      expect(await client.saveAgent(draft, create: true), isFalse);
      expect(client.error, contains('unsupported ContextGuard strategy'));
      expect(client.writeNotice, isNull);
      server.refusal = (403, 'permission_denied', 'secret');
      expect(await client.saveAgent(draft, create: true), isFalse);
      expect(client.user?.id, 'alice');
      expect(client.error, contains('permission'));
      expect(draft.agentId, 'draft');
    },
  );

  test(
    'Git-bound save uses observed version and only changed content paths',
    () async {
      final server = AgentServer()..bound = true;
      server.agents['alpha']!['stable'] = Agent(
        agentId: 'stable',
        name: 'runtime',
        type: AgentType.AGENT_TYPE_LLM,
        config: AgentConfig(instruction: 'Before'),
      );
      await server.start();
      addTearDown(server.close);
      final client = ClientContext(storage: MemoryStorage(), allowHttp: true);
      addTearDown(client.dispose);
      await client.setServer(server.url);
      await client.login('alice', 'password');
      await client.selectWorkspace('alpha');
      await client.loadAgent('stable');
      final before = client.managedAgent!;
      final draft = before.deepCopy()..config.instruction = 'After';
      expect(
        await client.saveAgent(draft, create: false, previous: before),
        isTrue,
      );
      final call = server.calls.lastWhere(
        (c) => c.method == 'UpdateAgentConfiguration',
      );
      final request = UpdateAgentConfigurationRequest.fromBuffer(call.body);
      expect(request.expectedAgentVersion, before.version);
      expect(request.contentChanges.single.path, 'agents/stable/prompt.md');
      expect(request.contentChanges.single.content, 'After');
      expect(
        request.contentChanges.single.operation,
        ContentFileOperation.CONTENT_FILE_OPERATION_PUT,
      );
    },
  );

  group(
    'Git-owned composite save shows the Active Revision, not the DB reply',
    () {
      Future<ClientContext> gitBound(AgentServer server) async {
        server.bound = true;
        server.agents['alpha']!['stable'] = Agent(
          agentId: 'stable',
          name: 'runtime',
          type: AgentType.AGENT_TYPE_LLM,
          description: 'DB description',
          config: AgentConfig(instruction: 'DB prompt'),
        );
        server.gitContent['alpha']!['stable'] = {
          'prompt.md': 'Git prompt',
          'description.md': 'Git description',
        };
        final client = await signedIn(server);
        await client.loadAgent('stable');
        expect(client.managedAgent!.config.instruction, 'Git prompt');
        return client;
      }

      int count(AgentServer server, String method) =>
          server.calls.where((c) => c.method == method).length;

      test(
        'a prompt change reads back instead of adopting the DB reply',
        () async {
          final server = AgentServer();
          final client = await gitBound(server);
          final before = client.managedAgent!;
          final draft = before.deepCopy()..config.instruction = 'New prompt';
          expect(
            await client.saveAgent(draft, create: false, previous: before),
            isTrue,
          );
          expect(client.managedAgent!.config.instruction, 'New prompt');
          expect(client.managedAgent!.description, 'Git description');
          final listed = client.agents.singleWhere(
            (a) => a.agentId == 'stable',
          );
          expect(listed.config.instruction, 'New prompt');
          // Only a read, never a replayed write.
          expect(count(server, 'UpdateAgentConfiguration'), 1);
          expect(server.calls.last.method, isNot('UpdateAgentConfiguration'));
        },
      );

      test('a labels-only change keeps the Git content on screen', () async {
        final server = AgentServer();
        final client = await gitBound(server);
        final before = client.managedAgent!;
        final draft = before.deepCopy()..labels['team'] = 'core';
        expect(
          await client.saveAgent(draft, create: false, previous: before),
          isTrue,
        );
        expect(client.managedAgent!.labels['team'], 'core');
        expect(client.managedAgent!.config.instruction, 'Git prompt');
        expect(client.managedAgent!.description, 'Git description');
        final listed = client.agents.singleWhere((a) => a.agentId == 'stable');
        expect(listed.config.instruction, 'Git prompt');
        expect(count(server, 'UpdateAgentConfiguration'), 1);
      });

      test(
        'a failed read-back never shows the DB reply and is not replayed',
        () async {
          final server = AgentServer();
          final client = await gitBound(server);
          final before = client.managedAgent!;
          final draft = before.deepCopy()..config.instruction = 'New prompt';
          server.getRefusal = (503, 'unavailable', 'down');
          expect(
            await client.saveAgent(draft, create: false, previous: before),
            isFalse,
          );
          expect(client.error, contains('could not be reloaded'));
          expect(client.managedAgent, isNull);
          expect(
            client.agents.every((a) => a.config.instruction != 'DB prompt'),
            isTrue,
          );
          expect(client.writeNotice, isNull);

          server.getRefusal = null;
          client.suspend();
          await client.revalidate();
          expect(client.managedAgent?.config.instruction, 'New prompt');
          expect(count(server, 'UpdateAgentConfiguration'), 1);
        },
      );

      test(
        'a read-back that outlives its Workspace cannot fill another',
        () async {
          final server = AgentServer();
          final client = await gitBound(server);
          server.agents['beta']!['stable'] = Agent(
            agentId: 'stable',
            name: 'beta',
          );
          final before = client.managedAgent!;
          final draft = before.deepCopy()..config.instruction = 'New prompt';
          final held = Completer<void>(), reached = Completer<void>();
          server.beforeGet = () async {
            reached.complete();
            await held.future;
          };
          final writing = client.saveAgent(
            draft,
            create: false,
            previous: before,
          );
          await reached.future;
          await client.selectWorkspace('beta');
          held.complete();
          expect(await writing, isFalse);
          expect(client.managedAgent, isNull);
          expect(client.agents.single.name, 'beta');
        },
      );
    },
  );

  test(
    'editor choices load every page and store model aliases, not credentials',
    () async {
      final server = AgentServer();
      for (var i = 0; i < 55; i++) {
        server.agents['alpha']!['child-$i'] = Agent(
          agentId: 'child-$i',
          name: 'duplicate label',
          type: AgentType.AGENT_TYPE_LLM,
        );
      }
      await server.start();
      addTearDown(server.close);
      final client = ClientContext(storage: MemoryStorage(), allowHttp: true);
      addTearDown(client.dispose);
      await client.setServer(server.url);
      await client.login('alice', 'password');
      await client.selectWorkspace('alpha');
      await client.prepareAgentEditor();
      expect(client.agentChoices.length, 55);
      expect(client.agentModelChoices, ['fast']);
      expect(client.agentSkillChoices, ['writing']);
      expect(client.modelProviders, isEmpty);
      expect(client.agentEditorReady, isTrue);
    },
  );

  test(
    'a member creates and edits an Agent addressed only by immutable ID',
    () async {
      final server = AgentServer();
      await server.start();
      addTearDown(server.close);
      final client = ClientContext(storage: MemoryStorage(), allowHttp: true);
      addTearDown(client.dispose);
      await client.setServer(server.url);
      await client.login('alice', 'password');
      await client.selectWorkspace('alpha');
      final created = Agent(
        agentId: 'stable',
        name: 'runtime',
        displayName: 'Visible',
        type: AgentType.AGENT_TYPE_LLM,
        config: AgentConfig(instruction: 'Hello'),
      );
      expect(await client.saveAgent(created, create: true), isTrue);
      await client.loadAgent('stable');
      expect(client.managedAgent?.displayName, 'Visible');
      final draft = client.managedAgent!.deepCopy()..displayName = 'Renamed';
      expect(
        await client.saveAgent(
          draft,
          create: false,
          previous: client.managedAgent,
        ),
        isTrue,
      );
      await client.loadAgent('stable');
      expect(client.managedAgent?.displayName, 'Renamed');
      expect(client.agentStatus?.agentId, 'stable');
      expect(server.calls.every((c) => c.workspace == 'alpha'), isTrue);
      expect(
        server.requests
            .where((r) => r.path.contains('AgentService'))
            .every((r) => r.token == 'Bearer alice'),
        isTrue,
      );
    },
  );
}
