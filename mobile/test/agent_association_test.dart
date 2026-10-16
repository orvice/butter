import 'dart:io';

import 'package:butter_mobile/app/client_context.dart';
import 'package:butter_mobile/features/agents/agent_draft.dart';
import 'package:butter_mobile/features/agents/agent_pages.dart';
import 'package:butter_mobile/gen/agents/v1/agent.pb.dart';
import 'package:butter_mobile/gen/agents/v1/agent_file.pb.dart';
import 'package:butter_mobile/gen/agents/v1/agent_service.pb.dart';
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';

import 'support/file_server.dart';
import 'support/memory_storage.dart';

typedef Editor = ({ClientContext client, FileServerFake server});

const _read = AgentFileMountPermission.AGENT_FILE_MOUNT_PERMISSION_READ;
const _write = AgentFileMountPermission.AGENT_FILE_MOUNT_PERMISSION_READ_WRITE;

Future<Editor> editor(
  WidgetTester tester, {
  Agent? base,
  void Function(FileServerFake)? seed,
}) async {
  final old = HttpOverrides.current;
  HttpOverrides.global = null;
  final server = FileServerFake();
  server.servers['alpha']!['docs-mcp'] = MCPServer(
    id: 'docs-mcp',
    name: 'Docs MCP',
  );
  server.servers['alpha']!['search'] = MCPServer(id: 'search', name: 'Search');
  server.remotes['alpha']!['ra'] = RemoteAgent(id: 'ra', name: 'Remote A');
  server.seedSpace('alpha', 'docs', 'Docs');
  server.seedSpace('alpha', 'notes', 'Team Notes');
  seed?.call(server);
  late ClientContext client;
  addTearDown(() async {
    client.dispose();
    await server.close();
    HttpOverrides.global = old;
  });
  tester.view.physicalSize = const Size(800, 4000);
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
  await tester.pumpWidget(
    MaterialApp(
      home: AgentFormPage(client: client, base: base),
    ),
  );
  await tester.pumpAndSettle();
  return (client: client, server: server);
}

Agent llm({
  List<String> mcp = const [],
  List<String> remote = const [],
  List<String> skills = const [],
  List<AgentFileMount> mounts = const [],
}) => Agent(
  agentId: 'stable',
  name: 'runtime',
  type: AgentType.AGENT_TYPE_LLM,
  config: AgentConfig(
    mcpServerIds: mcp,
    remoteAgentIds: remote,
    skills: skills,
    fileMounts: mounts,
  ),
);

Future<void> show(WidgetTester tester, Finder finder) async {
  FocusManager.instance.primaryFocus?.unfocus();
  await tester.pumpAndSettle();
  await tester.ensureVisible(finder);
  await tester.pumpAndSettle();
}

Future<void> tap(WidgetTester tester, Finder finder) async {
  await show(tester, finder);
  await tester.tap(finder);
  await tester.pumpAndSettle();
}

Future<void> pick(WidgetTester tester, Key dropdown, String label) async {
  await tap(tester, find.byKey(dropdown));
  await tester.tap(find.text(label).last);
  await tester.pumpAndSettle();
}

Future<void> save(WidgetTester tester, Editor h) async {
  final button = find.byKey(const Key('save-agent'));
  await show(tester, button);
  await tester.runAsync(() async {
    await tester.tap(button);
    for (var i = 0; i < 300 && h.client.busy; i++) {
      await Future<void>.delayed(const Duration(milliseconds: 10));
    }
  });
  await tester.pumpAndSettle();
}

Agent sent(Editor h) => UpdateAgentRequest.fromBuffer(
  h.server.calls.lastWhere((c) => c.method == 'UpdateAgent').body,
).agent;

int updates(Editor h) =>
    h.server.calls.where((c) => c.method == 'UpdateAgent').length;

void main() {
  testWidgets('an LLM Agent saves MCP, Remote Agent, mount and Skill '
      'associations that read back as chosen', (tester) async {
    final h = await editor(tester, base: llm());
    await tap(tester, find.byKey(const ValueKey('mcp-docs-mcp')));
    await tap(tester, find.byKey(const ValueKey('remote-ra')));
    await tap(tester, find.byKey(const ValueKey('skill-writing')));
    await pick(tester, const ValueKey('add-mount-0'), 'Team Notes (notes)');
    // The suggested path is the space name as a slug, with an explicit READ.
    expect(
      find.text('Mounted at /team-notes. An empty path mounts at /.'),
      findsOneWidget,
    );
    await pick(
      tester,
      const ValueKey('mount-0-permission-null-1'),
      'Read and write',
    );
    expect(
      find.text('This Agent can change files in Team Notes when it runs.'),
      findsOneWidget,
    );
    await save(tester, h);
    final a = sent(h);
    expect(a.config.mcpServerIds, ['docs-mcp']);
    expect(a.config.remoteAgentIds, ['ra']);
    expect(a.config.skills, ['writing']);
    expect(a.config.fileMounts, hasLength(1));
    expect(a.config.fileMounts.single.spaceId, 'notes');
    expect(a.config.fileMounts.single.mountPath, '/team-notes');
    expect(a.config.fileMounts.single.permission, _write);
  });

  testWidgets('a mount keeps unknown fields and its unspecified permission '
      'until the user changes them', (tester) async {
    final odd = AgentFileMount(spaceId: 'docs', mountPath: '/d')
      ..mergeFromBuffer([0xd0, 0x05, 7]); // future field 90
    final unknownPermission = AgentFileMount(spaceId: 'notes', mountPath: '/n')
      ..mergeFromBuffer([0x18, 77]); // permission = future enum 77
    final h = await editor(tester, base: llm(mounts: [odd, unknownPermission]));
    expect(find.byKey(const Key('mount-1-unknown')), findsOneWidget);
    // Editing something else sends both mounts byte for byte.
    await tester.enterText(find.byKey(const Key('display-name')), 'Renamed');
    await save(tester, h);
    expect(sent(h).config.fileMounts.map((m) => m.writeToBuffer()).toList(), [
      odd.writeToBuffer(),
      unknownPermission.writeToBuffer(),
    ]);
    expect(sent(h).config.fileMounts.first.permission.value, 0);
  });

  testWidgets('choosing a permission replaces an unknown one', (tester) async {
    final unknownPermission = AgentFileMount(spaceId: 'notes', mountPath: '/n')
      ..mergeFromBuffer([0x18, 77]);
    final h = await editor(tester, base: llm(mounts: [unknownPermission]));
    await pick(tester, const ValueKey('mount-0-permission-77-0'), 'Read only');
    await save(tester, h);
    final m = sent(h).config.fileMounts.single;
    expect(m.permission, _read);
    expect(m.unknownFields.isEmpty, isTrue);
  });

  testWidgets('two mounts at the same path are refused before any write', (
    tester,
  ) async {
    final h = await editor(
      tester,
      base: llm(
        mounts: [
          AgentFileMount(
            spaceId: 'docs',
            mountPath: '/shared',
            permission: _read,
          ),
          AgentFileMount(
            spaceId: 'notes',
            mountPath: 'shared/',
            permission: _read,
          ),
        ],
      ),
    );
    await save(tester, h);
    expect(updates(h), 0);
    expect(
      find.text(
        'Another mount already uses /shared. The runtime would drop one of them.',
      ),
      findsWidgets,
    );
    // Moving one aside clears the problem and the save goes through.
    await tester.enterText(find.byKey(const Key('mount-1-path')), '/other');
    await save(tester, h);
    expect(updates(h), 1);
    expect(sent(h).config.fileMounts.map((m) => m.mountPath), [
      '/shared',
      '/other',
    ]);
  });

  testWidgets('a mount of a space that is not there is flagged, not hidden', (
    tester,
  ) async {
    await editor(
      tester,
      base: llm(
        mounts: [
          AgentFileMount(spaceId: 'gone', mountPath: '/g', permission: _read),
        ],
      ),
    );
    expect(find.byKey(const Key('mount-0-missing')), findsOneWidget);
    expect(find.text('Space gone'), findsOneWidget);
  });

  testWidgets('a reference that is no longer there stays selected and flagged '
      'until removed', (tester) async {
    final h = await editor(
      tester,
      base: llm(mcp: ['gone', 'docs-mcp'], remote: ['lost']),
    );
    expect(
      find.textContaining('refuses to save a missing MCP server'),
      findsOneWidget,
    );
    expect(
      find.textContaining('refuses to save a missing Remote Agent'),
      findsOneWidget,
    );
    await tap(tester, find.byKey(const ValueKey('mcp-gone')));
    await tap(tester, find.byKey(const ValueKey('remote-lost')));
    await save(tester, h);
    expect(sent(h).config.mcpServerIds, ['docs-mcp']);
    expect(sent(h).config.remoteAgentIds, isEmpty);
  });

  testWidgets('a refused list does not block editing and keeps the selection', (
    tester,
  ) async {
    final h = await editor(
      tester,
      base: llm(
        mcp: ['docs-mcp'],
        mounts: [
          AgentFileMount(spaceId: 'docs', mountPath: '/d', permission: _read),
        ],
      ),
      seed: (s) {
        s.refusals['ListMCPServers'] = (
          403,
          'permission_denied',
          'secret server words',
        );
        s.fileRefusals['ListAgentFileSpaces'] = (
          500,
          'internal',
          'secret server words',
        );
      },
    );
    expect(h.client.agentEditorReady, isTrue);
    expect(find.byKey(const Key('mcp-note')), findsOneWidget);
    expect(find.textContaining('Your role cannot list these'), findsOneWidget);
    expect(
      find.textContaining('File spaces could not be loaded'),
      findsOneWidget,
    );
    expect(find.textContaining('secret server words'), findsNothing);
    // Nothing can be added from a list that is not there.
    expect(find.byKey(const ValueKey('add-mount-1')), findsNothing);
    expect(find.byKey(const ValueKey('mcp-search')), findsNothing);
    await tester.enterText(find.byKey(const Key('display-name')), 'Renamed');
    await save(tester, h);
    expect(sent(h).config.mcpServerIds, ['docs-mcp']);
    expect(sent(h).config.fileMounts.single.spaceId, 'docs');
  });

  testWidgets('a Loop offers Remote Agents only; a legacy MCP id can be '
      'removed but not added', (tester) async {
    final loop = Agent(
      agentId: 'loop',
      name: 'runtime',
      type: AgentType.AGENT_TYPE_LOOP,
      config: AgentConfig(maxIterations: 2, mcpServerIds: ['docs-mcp']),
    );
    final h = await editor(tester, base: loop);
    expect(find.byKey(const ValueKey('remote-ra')), findsOneWidget);
    expect(find.byKey(const ValueKey('skill-writing')), findsNothing);
    expect(find.text('File mounts'), findsNothing);
    // The retained id is listed; nothing else on the list is offered.
    expect(find.byKey(const ValueKey('mcp-docs-mcp')), findsOneWidget);
    expect(find.byKey(const ValueKey('mcp-search')), findsNothing);
    expect(find.textContaining('The runtime ignores these'), findsOneWidget);
    await tap(tester, find.byKey(const ValueKey('remote-ra')));
    await save(tester, h);
    expect(sent(h).config.remoteAgentIds, ['ra']);
    expect(sent(h).config.mcpServerIds, ['docs-mcp']);
  });

  testWidgets('a Workflow offers none of them', (tester) async {
    final flow = Agent(
      agentId: 'flow',
      name: 'runtime',
      type: AgentType.AGENT_TYPE_WORKFLOW,
      config: AgentConfig(workflow: WorkflowConfig()),
    );
    await editor(tester, base: flow);
    for (final key in ['mcp-docs-mcp', 'remote-ra', 'skill-writing']) {
      expect(find.byKey(ValueKey(key)), findsNothing, reason: key);
    }
    expect(find.text('File mounts'), findsNothing);
  });

  testWidgets('a refused save for a missing reference says nothing was saved '
      'and does not echo the server', (tester) async {
    final h = await editor(
      tester,
      base: llm(mcp: ['docs-mcp']),
      seed: (s) => s.refusal = (
        500,
        'internal',
        'internal: rebuilding agent "x": unknown mcp_server_id "docs-mcp" private-detail',
      ),
    );
    await tester.enterText(find.byKey(const Key('display-name')), 'Renamed');
    await save(tester, h);
    expect(
      find.textContaining('Nothing was saved: an MCP server this Agent'),
      findsOneWidget,
    );
    expect(find.textContaining('private-detail'), findsNothing);
    expect(updates(h), 1, reason: 'a refused save is not replayed');
  });

  test('a new Agent is not sent associations its final type ignores', () {
    final draft = AgentDraft();
    draft.config
      ..mcpServerIds.add('docs-mcp')
      ..skills.add('writing')
      ..remoteAgentIds.add('ra')
      ..fileMounts.add(AgentFileMount(spaceId: 'docs', mountPath: '/d'));
    // Switching away and back loses nothing from the draft itself.
    draft.setType(AgentType.AGENT_TYPE_LOOP);
    expect(draft.build().config.mcpServerIds, ['docs-mcp']);
    final loop = draft.buildForCreate();
    expect(loop.config.mcpServerIds, isEmpty);
    expect(loop.config.skills, isEmpty);
    expect(loop.config.fileMounts, isEmpty);
    expect(loop.config.remoteAgentIds, ['ra']);
    draft.setType(AgentType.AGENT_TYPE_WORKFLOW);
    expect(draft.buildForCreate().config.remoteAgentIds, isEmpty);
    draft.setType(AgentType.AGENT_TYPE_LLM);
    final back = draft.buildForCreate();
    expect(back.config.mcpServerIds, ['docs-mcp']);
    expect(back.config.fileMounts, hasLength(1));
  });
}
