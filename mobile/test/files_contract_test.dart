import 'package:butter_mobile/app/client_context.dart';
import 'package:butter_mobile/core/agent_errors.dart';
import 'package:butter_mobile/core/butter_api.dart';
import 'package:butter_mobile/core/config_errors.dart';
import 'package:butter_mobile/core/server_address.dart';
import 'package:butter_mobile/features/chat/chat_controller.dart';
import 'package:butter_mobile/features/files/file_controller.dart';
import 'package:butter_mobile/features/files/file_rules.dart';
import 'package:butter_mobile/gen/agents/v1/agent.pb.dart';
import 'package:butter_mobile/gen/agents/v1/agent_file.pb.dart';
import 'package:connectrpc/connect.dart';
import 'package:flutter_test/flutter_test.dart';

import 'support/memory_storage.dart';

// AgentFileService, AgentService, MCPServerService and the AG-UI chat route are
// the production handlers, served by the contract fixture
// (mobile/tool/contractserver/files.go and chat.go) over in-memory
// repositories. The model is a scripted OpenAI-compatible test double and the
// MCP server is an in-process double. The chat runs below are real runner
// turns: what the Agent can do is decided by the configuration this App saved.
// Nothing here is a real provider, S3 or object store, or a device run.
const workspace = 'files';
const peerWorkspace = 'files-peer';
const deniedWorkspace = 'mcp-denied';

Matcher code(Code code) =>
    isA<ConnectException>().having((e) => e.code, 'code', code);

Future<void> until(bool Function() done, {String? reason}) async {
  final stop = DateTime.now().add(const Duration(seconds: 20));
  while (!done()) {
    if (DateTime.now().isAfter(stop)) fail(reason ?? 'condition never held');
    await Future<void>.delayed(const Duration(milliseconds: 20));
  }
}

void main() {
  const server = String.fromEnvironment('BUTTER_TEST_SERVER');
  for (final prefix in ['', '/butter']) {
    group(
      'File service production Connect contract $prefix',
      skip: server.isEmpty ? 'Run tool/contract.sh.' : false,
      () {
        final base = '$server$prefix';
        final tag = prefix.isEmpty ? 'root' : 'prefix';
        late ButterApi api;
        late String token;
        setUp(() async {
          api = ButterApi(ServerAddress.parse(base, allowHttp: true));
          addTearDown(api.close);
          token = (await api.login(
            'mobile-test',
            'contract-test-password',
          )).token;
        });

        Future<String> newSpace(String name, {String ws = workspace}) async =>
            (await api.createFileSpace(
              token,
              ws,
              AgentFileSpace(id: '$name-$tag', name: name),
            )).id;

        test('spaces are created, listed, updated and deleted; a delete takes '
            'the files with it', () async {
          final id = await newSpace('Crud');
          final listed = await api.fileSpaces(token, workspace);
          expect(listed.where((s) => s.id == id), hasLength(1));
          final read = await api.fileSpace(token, workspace, id);
          expect(read.workspaceId, workspace);
          final edited = read.deepCopy()..description = 'changed';
          final saved = await api.updateFileSpace(token, workspace, edited);
          expect(saved.description, 'changed');
          await api.writeAgentFile(
            token,
            workspace,
            id,
            '/a.md',
            'x',
            contentType: defaultFileContentType,
          );
          await expectLater(
            api.createFileSpace(
              token,
              workspace,
              AgentFileSpace(id: id, name: 'again'),
            ),
            throwsA(code(Code.alreadyExists)),
          );
          await expectLater(
            api.createFileSpace(token, workspace, AgentFileSpace(id: 'x-$tag')),
            throwsA(code(Code.invalidArgument)),
          );
          await api.deleteFileSpace(token, workspace, id);
          await expectLater(
            api.fileSpace(token, workspace, id),
            throwsA(code(Code.notFound)),
          );
          await expectLater(
            api.readAgentFile(token, workspace, id, '/a.md'),
            throwsA(anything),
          );
        });

        test('the App normalizes a path exactly as the server does', () async {
          final id = await newSpace('Paths');
          const inputs = [
            'a.md',
            '/x//y/../z.md',
            ' /padded.md ',
            '/dir/sub/',
            '/./dot.md',
            '/Mixed Case.MD',
            '/ünï/çødé.md',
            '',
            '/',
            '/..',
            '..',
            '/a/../..',
          ];
          for (final input in inputs) {
            final expected = normalizeFilePath(input);
            if (expected == null) {
              await expectLater(
                api.writeAgentFile(
                  token,
                  workspace,
                  id,
                  input,
                  'x',
                  contentType: defaultFileContentType,
                ),
                throwsA(code(Code.invalidArgument)),
                reason: 'path "$input" should be refused',
              );
              continue;
            }
            final stored = await api.writeAgentFile(
              token,
              workspace,
              id,
              input,
              'x',
              contentType: defaultFileContentType,
            );
            expect(stored.path, expected, reason: 'path "$input"');
          }
          for (final prefixInput in ['', '/', '/dir', 'dir/', '/..']) {
            final expected = normalizeFilePrefix(prefixInput);
            if (expected == null) {
              await expectLater(
                api.agentFiles(token, workspace, id, pathPrefix: prefixInput),
                throwsA(code(Code.invalidArgument)),
                reason: 'prefix "$prefixInput" should be refused',
              );
            } else {
              final listed = await api.agentFiles(
                token,
                workspace,
                id,
                pathPrefix: prefixInput,
              );
              expect(
                listed.every((f) => f.path.startsWith(expected)),
                isTrue,
                reason: 'prefix "$prefixInput"',
              );
              if (expected.isEmpty) expect(listed, isNotEmpty);
            }
          }
        });

        test(
          'every write is a new version, readable by number; metadata '
          'belongs to the latest write; a missing version is not found',
          () async {
            final id = await newSpace('Versions');
            Future<AgentFile> write(String content, [Map<String, String>? m]) =>
                api.writeAgentFile(
                  token,
                  workspace,
                  id,
                  '/v.md',
                  content,
                  contentType: 'text/markdown',
                  metadata: m ?? const {},
                );
            expect((await write('one', {'k': 'v'})).version.toInt(), 1);
            expect((await write('two')).version.toInt(), 2);
            expect((await write('three')).version.toInt(), 3);
            final v1 = await api.readAgentFile(
              token,
              workspace,
              id,
              '/v.md',
              version: 1,
            );
            expect(v1.content, 'one');
            final latest = await api.readAgentFile(
              token,
              workspace,
              id,
              '/v.md',
            );
            expect(latest.content, 'three');
            expect(latest.file.version.toInt(), 3);
            expect(
              latest.file.metadata,
              isEmpty,
              reason: 'a write replaces it',
            );
            await expectLater(
              api.readAgentFile(token, workspace, id, '/v.md', version: 9),
              throwsA(code(Code.notFound)),
            );
            await api.deleteAgentFile(token, workspace, id, '/v.md');
            await expectLater(
              api.readAgentFile(token, workspace, id, '/v.md'),
              throwsA(code(Code.notFound)),
            );
          },
        );

        test('search finds text; an empty query is refused', () async {
          final id = await newSpace('Search');
          await api.writeAgentFile(
            token,
            workspace,
            id,
            '/n.md',
            'a unique-needle-$tag here',
            contentType: defaultFileContentType,
          );
          final hits = await api.searchAgentFiles(
            token,
            workspace,
            id,
            'unique-needle-$tag',
          );
          expect(hits.map((h) => h.file.path), ['/n.md']);
          await expectLater(
            api.searchAgentFiles(token, workspace, id, ''),
            throwsA(code(Code.invalidArgument)),
          );
        });

        test(
          'the size limit counts UTF-8 bytes, and the refusal names it',
          () async {
            final id = await newSpace('Limits');
            Future<AgentFile> put(String content) => api.writeAgentFile(
              token,
              workspace,
              id,
              '/big.txt',
              content,
              contentType: defaultFileContentType,
            );
            expect(
              (await put('a' * defaultFileMaxBytes)).sizeBytes.toInt(),
              defaultFileMaxBytes,
            );
            // 131,073 two-byte characters are 262,146 bytes: over the limit
            // though far under it in characters.
            final long = 'é' * (defaultFileMaxBytes ~/ 2 + 1);
            expect(utf8Length(long), defaultFileMaxBytes + 2);
            await expectLater(
              put(long),
              throwsA(
                isA<ConnectException>()
                    .having((e) => e.code, 'code', Code.invalidArgument)
                    .having(
                      (e) => fileSizeLimitFrom(e.message),
                      'limit',
                      defaultFileMaxBytes,
                    ),
              ),
            );
            // The refused write stored nothing new.
            expect(
              (await api.readAgentFile(
                token,
                workspace,
                id,
                '/big.txt',
              )).file.version.toInt(),
              1,
            );
          },
        );

        test('a Workspace never sees another\'s spaces, and a non-member is '
            'refused', () async {
          final id = await newSpace('Private');
          await api.writeAgentFile(
            token,
            workspace,
            id,
            '/p.md',
            'private',
            contentType: defaultFileContentType,
          );
          expect(
            (await api.fileSpaces(
              token,
              peerWorkspace,
            )).where((s) => s.id == id),
            isEmpty,
          );
          await expectLater(
            api.fileSpace(token, peerWorkspace, id),
            throwsA(code(Code.notFound)),
          );
          await expectLater(
            api.readAgentFile(token, peerWorkspace, id, '/p.md'),
            throwsA(anything),
          );
          await expectLater(
            api.fileSpaces(token, deniedWorkspace),
            throwsA(code(Code.permissionDenied)),
          );
          await expectLater(
            api.writeAgentFile(
              token,
              deniedWorkspace,
              id,
              '/x.md',
              'x',
              contentType: defaultFileContentType,
            ),
            throwsA(code(Code.permissionDenied)),
          );
        });

        test('the controller refuses a save over a version it did not see; '
            'only an explicit overwrite replaces it', () async {
          final id = await newSpace('Conflict');
          Future<void> put(String content) => api.writeAgentFile(
            token,
            workspace,
            id,
            '/c.md',
            content,
            contentType: defaultFileContentType,
          );
          await put('one');
          final client = ClientContext(
            storage: MemoryStorage(),
            allowHttp: true,
          );
          addTearDown(client.dispose);
          await client.setServer(base);
          await client.login('mobile-test', 'contract-test-password');
          await client.selectWorkspace(workspace);
          final files = client.fileSpaces!;

          // A new file over an existing one.
          expect(
            await files.saveFile(
              space: id,
              path: 'c.md',
              content: 'mine',
              baseVersion: null,
            ),
            isNull,
          );
          expect(files.conflict?.kind, FileConflictKind.exists);
          // Opened at version 1, changed elsewhere to version 2.
          await put('two');
          expect(
            await files.saveFile(
              space: id,
              path: '/c.md',
              content: 'mine',
              baseVersion: 1,
            ),
            isNull,
          );
          expect(files.conflict?.kind, FileConflictKind.changed);
          expect(files.conflict?.currentVersion, 2);
          expect(
            (await api.readAgentFile(token, workspace, id, '/c.md')).content,
            'two',
            reason: 'a refused save writes nothing',
          );
          final over = await files.saveFile(
            space: id,
            path: '/c.md',
            content: 'mine',
            baseVersion: 1,
            overwrite: true,
          );
          expect(over?.version.toInt(), 3);
          // Deleted elsewhere since it was opened.
          await api.deleteAgentFile(token, workspace, id, '/c.md');
          expect(
            await files.saveFile(
              space: id,
              path: '/c.md',
              content: 'again',
              baseVersion: 3,
            ),
            isNull,
          );
          expect(files.conflict?.kind, FileConflictKind.deleted);
        });

        test('saved associations decide what a real run can do', () async {
          final id = await newSpace('Run');
          await api.writeAgentFile(
            token,
            workspace,
            id,
            '/guide.md',
            'mounted content $tag',
            contentType: defaultFileContentType,
          );
          final mcpId = 'open-$tag';
          await api.createMcpServer(
            token,
            workspace,
            MCPServer(
              id: mcpId,
              name: mcpId,
              transport:
                  MCPServerTransport.MCP_SERVER_TRANSPORT_STREAMABLE_HTTP,
              url: '$server/fake-mcp/open',
            ),
          );
          final agentId = 'talker-$tag';
          Agent agent(AgentFileMountPermission permission) => Agent(
            agentId: agentId,
            name: agentId,
            type: AgentType.AGENT_TYPE_LLM,
            config: AgentConfig(
              model: 'fixture-model',
              instruction: 'Be brief.',
              mcpServerIds: [mcpId],
              skills: ['files-skill'],
              fileMounts: [
                AgentFileMount(
                  spaceId: id,
                  mountPath: '/docs',
                  permission: permission,
                ),
              ],
            ),
          );
          await api.saveAgent(
            token,
            workspace,
            agent(AgentFileMountPermission.AGENT_FILE_MOUNT_PERMISSION_READ),
            create: true,
          );
          // What was saved reads back as chosen.
          final reread = await api.agent(token, workspace, agentId);
          expect(reread.config.mcpServerIds, [mcpId]);
          expect(reread.config.skills, ['files-skill']);
          expect(reread.config.fileMounts.single.spaceId, id);
          expect(reread.config.fileMounts.single.mountPath, '/docs');
          expect(
            reread.config.fileMounts.single.permission,
            AgentFileMountPermission.AGENT_FILE_MOUNT_PERMISSION_READ,
          );

          final client = ClientContext(
            storage: MemoryStorage(),
            allowHttp: true,
            chatIdleTimeout: const Duration(seconds: 30),
          );
          addTearDown(client.dispose);
          await client.setServer(base);
          await client.login('mobile-test', 'contract-test-password');
          await client.selectWorkspace(workspace);
          final chat = client.chat!;
          await chat.loadThreads();
          chat.startThread(chat.agents.firstWhere((a) => a.agentId == agentId));

          Future<String> say(String text) async {
            expect(
              await chat.send(text, const []),
              isTrue,
              reason: chat.sendError,
            );
            await until(
              () => chat.phase == ChatPhase.idle,
              reason: 'run ended',
            );
            return chat.messages.last.text;
          }

          final offered = await say('tools?');
          for (final tool in ['agent_files_read_file', 'echo', 'list_skills']) {
            expect(offered, contains(tool));
          }
          expect(
            await say('tool:agent_files_read_file {"path":"/docs/guide.md"}'),
            contains('mounted content $tag'),
          );
          expect(
            await say(
              'tool:agent_files_write_file {"path":"/docs/guide.md","content":"changed"}',
            ),
            contains('does not allow write'),
          );
          expect(
            (await api.readAgentFile(
              token,
              workspace,
              id,
              '/guide.md',
            )).content,
            'mounted content $tag',
            reason: 'a READ mount never changes the space',
          );
          expect(
            await say('tool:echo {"text":"hello-$tag"}'),
            contains('hello-$tag'),
          );

          // Raising the permission in the saved Agent lets the next run write.
          final stored = await api.agent(token, workspace, agentId);
          stored.config.fileMounts.single.permission =
              AgentFileMountPermission.AGENT_FILE_MOUNT_PERMISSION_READ_WRITE;
          await api.saveAgent(token, workspace, stored, create: false);
          chat.startThread(chat.agents.firstWhere((a) => a.agentId == agentId));
          final wrote = await say(
            'tool:agent_files_write_file {"path":"/docs/guide.md","content":"changed"}',
          );
          expect(wrote, isNot(contains('does not allow')));
          final after = await api.readAgentFile(
            token,
            workspace,
            id,
            '/guide.md',
          );
          expect(after.content, 'changed');
          expect(after.file.version.toInt(), 2);
          expect(
            await say('tool:agent_files_delete_file {"path":"/docs/guide.md"}'),
            contains('does not allow delete'),
          );
        });

        test('a save that names a missing MCP server or Remote Agent is '
            'refused and stores nothing', () async {
          final agent = Agent(
            agentId: 'dangling-$tag',
            name: 'dangling-$tag',
            type: AgentType.AGENT_TYPE_LLM,
            config: AgentConfig(
              model: 'fixture-model',
              mcpServerIds: ['no-such-server'],
            ),
          );
          Object? failure;
          try {
            await api.saveAgent(token, workspace, agent, create: true);
          } catch (e) {
            failure = e;
          }
          expect(failure, isA<ConnectException>());
          expect(
            describeAgentFailure(failure!),
            contains('Nothing was saved: an MCP server'),
          );
          await expectLater(
            api.agent(token, workspace, 'dangling-$tag'),
            throwsA(code(Code.notFound)),
          );
        });
      },
    );
  }
}
