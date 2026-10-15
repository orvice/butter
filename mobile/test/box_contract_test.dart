import 'package:butter_mobile/app/client_context.dart';
import 'package:butter_mobile/core/butter_api.dart';
import 'package:butter_mobile/core/server_address.dart';
import 'package:butter_mobile/features/butterboxes/box_controller.dart';
import 'package:butter_mobile/features/chat/chat_controller.dart';
import 'package:butter_mobile/gen/agents/v1/agent.pb.dart';
import 'package:connectrpc/connect.dart';
import 'package:flutter_test/flutter_test.dart';

import 'support/memory_storage.dart';

// The Box here is an in-process test double served by the contract fixture
// (mobile/tool/contractserver/boxes.go) through the real generated Pi and
// Cursor handlers. Butter's ButterBoxService, AgentService and runner are the
// production ones. Nothing in this file is a production Box or a device run.
const workspace = 'boxes';
const goodToken = 'box-secret-token';

Matcher code(Code code) =>
    isA<ConnectException>().having((e) => e.code, 'code', code);

Future<void> until(bool Function() done, {String? reason}) async {
  final stop = DateTime.now().add(const Duration(seconds: 30));
  while (!done()) {
    if (DateTime.now().isAfter(stop)) fail(reason ?? 'condition never held');
    await Future<void>.delayed(const Duration(milliseconds: 20));
  }
}

void main() {
  const server = String.fromEnvironment('BUTTER_TEST_SERVER');
  for (final prefix in ['', '/butter']) {
    group(
      'ButterBox and leaf production contract $prefix',
      skip: server.isEmpty ? 'Run tool/contract.sh.' : false,
      () {
        final base = '$server$prefix';
        final tag = prefix.isEmpty ? 'root' : 'prefix';
        // The double is reached by the Butter server itself, at the
        // fixture's own address, never by this device.
        String boxUrl(String profile) => '$server/fake-box/$profile';
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

        Future<ClientContext> signedIn() async {
          final client = ClientContext(
            storage: MemoryStorage(),
            allowHttp: true,
            chatIdleTimeout: const Duration(seconds: 30),
          );
          addTearDown(client.dispose);
          await client.setServer(base);
          await client.login('mobile-test', 'contract-test-password');
          await client.selectWorkspace(workspace);
          expect(client.error, isNull);
          return client;
        }

        Future<Agent> saveLeaf(Agent agent) =>
            api.saveAgent(token, workspace, agent, create: true);

        test('boxes round-trip with a write-only token', () async {
          final box = await api.createButterBox(
            token,
            workspace,
            name: 'laptop-$tag',
            baseUrl: boxUrl('ok'),
            enabled: true,
            accessToken: goodToken,
          );
          expect(box.credentialSet, isTrue);
          expect(box.workspaceId, workspace);
          expect(box.writeToBuffer().toString(), isNot(contains(goodToken)));
          expect(
            (await api.butterBoxes(token, workspace)).map((b) => b.id),
            contains(box.id),
          );
          // Another Workspace does not see it.
          expect(await api.butterBoxes(token, 'alpha'), isEmpty);
          await expectLater(
            api.createButterBox(
              token,
              workspace,
              name: 'laptop-$tag',
              baseUrl: boxUrl('ok'),
              enabled: true,
            ),
            throwsA(code(Code.alreadyExists)),
          );

          // Editing never carries the token: the box still answers.
          final edited = await api.updateButterBox(
            token,
            workspace,
            id: box.id,
            name: 'laptop-$tag',
            baseUrl: boxUrl('ok'),
            enabled: true,
          );
          expect(edited.credentialSet, isTrue);
          expect(
            (await api.butterBoxStatus(token, workspace, box.id)).reachable,
            isTrue,
          );

          // Replace with a wrong token, then clear: separate writes.
          final replaced = await api.setButterBoxToken(
            token,
            workspace,
            box.id,
            'wrong',
          );
          expect(replaced.credentialSet, isTrue);
          final refused = await api.butterBoxStatus(token, workspace, box.id);
          expect(refused.reachable, isFalse);
          expect(refused.error, isNotEmpty);
          final cleared = await api.setButterBoxToken(
            token,
            workspace,
            box.id,
            '',
          );
          expect(cleared.credentialSet, isFalse);
        });

        test(
          'an unreachable box is data; catalogs report real errors',
          () async {
            final dead = await api.createButterBox(
              token,
              workspace,
              name: 'dead-$tag',
              baseUrl: 'http://127.0.0.1:1',
              enabled: true,
            );
            final status = await api.butterBoxStatus(token, workspace, dead.id);
            expect(status.reachable, isFalse);
            expect(status.error, isNotEmpty);

            final ok = await api.createButterBox(
              token,
              workspace,
              name: 'catalog-$tag',
              baseUrl: boxUrl('ok'),
              enabled: true,
              accessToken: goodToken,
            );
            expect(
              (await api.butterBoxModels(
                token,
                workspace,
                ok.id,
              )).map((m) => '${m.provider}/${m.id}'),
              ['anthropic/claude-fable-5', 'openai/gpt-5.6'],
            );
            expect(
              (await api.butterBoxCursorModels(
                token,
                workspace,
                ok.id,
              )).map((m) => m.id),
              ['composer-2.5', 'auto-smart'],
            );

            final full = await api.createButterBox(
              token,
              workspace,
              name: 'full-$tag',
              baseUrl: boxUrl('capacity'),
              enabled: true,
            );
            await expectLater(
              api.butterBoxModels(token, workspace, full.id),
              throwsA(code(Code.resourceExhausted)),
            );
            final nokey = await api.createButterBox(
              token,
              workspace,
              name: 'nokey-$tag',
              baseUrl: boxUrl('nokey'),
              enabled: true,
            );
            await expectLater(
              api.butterBoxCursorModels(token, workspace, nokey.id),
              throwsA(
                isA<ConnectException>()
                    .having((e) => e.code, 'code', Code.failedPrecondition)
                    .having(
                      (e) => e.message,
                      'message',
                      contains('CURSOR_API_KEY'),
                    ),
              ),
            );
          },
        );

        test(
          'the controller turns those errors into actionable text',
          () async {
            final client = await signedIn();
            final boxes = client.boxes!;
            final full = await api.createButterBox(
              token,
              workspace,
              name: 'ctrl-full-$tag',
              baseUrl: boxUrl('capacity'),
              enabled: true,
            );
            final nokey = await api.createButterBox(
              token,
              workspace,
              name: 'ctrl-nokey-$tag',
              baseUrl: boxUrl('nokey'),
              enabled: true,
            );
            await boxes.loadPiModels(full.id);
            expect(boxes.piCatalogs[full.id]!.state, BoxLoad.failed);
            expect(boxes.piCatalogs[full.id]!.error, contains('capacity'));
            expect(boxes.piCatalogs[full.id]!.error, contains('by hand'));
            await boxes.loadCursorModels(nokey.id);
            expect(boxes.cursorCatalogs[nokey.id]!.state, BoxLoad.failed);
            expect(
              boxes.cursorCatalogs[nokey.id]!.error,
              contains('CURSOR_API_KEY'),
            );
            await boxes.probe(nokey.id);
            expect(boxes.probes[nokey.id]!.state, BoxProbeState.reachable);
          },
        );

        test('leaf Agents save, round-trip the run limit and keep the box '
            'from deletion', () async {
          final box = await api.createButterBox(
            token,
            workspace,
            name: 'leafbox-$tag',
            baseUrl: boxUrl('ok'),
            enabled: true,
            accessToken: goodToken,
          );
          final limits = <String, int?>{
            'default': null,
            'unlimited': 0,
            'timed': 90,
          };
          for (final entry in limits.entries) {
            final id = 'pi-${entry.key}-$tag';
            final saved = await saveLeaf(
              Agent(
                agentId: id,
                name: id,
                type: AgentType.AGENT_TYPE_PI,
                config: AgentConfig(
                  pi: PiAgentConfig(
                    butterboxId: box.id,
                    workingDir: '/srv/repo',
                    provider: 'anthropic',
                    model: 'claude-fable-5',
                    thinkingLevel: 'high',
                    maxRunSeconds: entry.value,
                  ),
                ),
              ),
            );
            final readBack = await api.agent(token, workspace, id);
            for (final pi in [saved.config.pi, readBack.config.pi]) {
              expect(pi.hasMaxRunSeconds(), entry.value != null, reason: id);
              if (entry.value != null) {
                expect(pi.maxRunSeconds, entry.value, reason: id);
              }
              expect(pi.butterboxId, box.id);
              expect(pi.model, 'claude-fable-5');
            }
          }

          // A Butter-side behavior field on a leaf is told, not silently kept.
          await expectLater(
            saveLeaf(
              Agent(
                agentId: 'pi-bad-$tag',
                name: 'pi-bad-$tag',
                type: AgentType.AGENT_TYPE_PI,
                config: AgentConfig(
                  instruction: 'no',
                  pi: PiAgentConfig(butterboxId: box.id),
                ),
              ),
            ),
            throwsA(code(Code.invalidArgument)),
          );

          // The reference-blocked delete names the Agents, and the controller
          // says so without removing the box.
          final client = await signedIn();
          final boxes = client.boxes!;
          await boxes.loadBoxes();
          expect(await boxes.delete(box.id), isFalse);
          expect(boxes.writeError, contains('pi-timed-$tag'));
          expect(boxes.byId(box.id), isNotNull);
          expect(boxes.unconfirmed, isFalse);

          // An unreferenced box deletes.
          final spare = await api.createButterBox(
            token,
            workspace,
            name: 'spare-$tag',
            baseUrl: boxUrl('ok'),
            enabled: true,
          );
          await boxes.loadBoxes();
          expect(await boxes.delete(spare.id), isTrue);
          expect(boxes.byId(spare.id), isNull);
        });

        test('Pi and Cursor leaves run through the real bridges with session '
            'continuity and a fresh session after repointing', () async {
          final ok = await api.createButterBox(
            token,
            workspace,
            name: 'runbox-$tag',
            baseUrl: boxUrl('ok'),
            enabled: true,
            accessToken: goodToken,
          );
          await saveLeaf(
            Agent(
              agentId: 'pi-run-$tag',
              name: 'pi-run-$tag',
              type: AgentType.AGENT_TYPE_PI,
              config: AgentConfig(
                pi: PiAgentConfig(
                  butterboxId: ok.id,
                  workingDir: '/srv/a',
                  provider: 'anthropic',
                  model: 'claude-fable-5',
                ),
              ),
            ),
          );
          await saveLeaf(
            Agent(
              agentId: 'cursor-run-$tag',
              name: 'cursor-run-$tag',
              type: AgentType.AGENT_TYPE_CURSOR,
              config: AgentConfig(
                cursor: CursorAgentConfig(
                  butterboxId: ok.id,
                  workingDir: '/srv/c',
                  model: 'composer-2.5',
                  mode: 'plan',
                ),
              ),
            ),
          );

          final client = await signedIn();
          final chat = client.chat!;
          await chat.loadThreads();

          Future<String> say(
            String agentId,
            String text, {
            bool fresh = false,
          }) async {
            if (fresh) {
              chat.startThread(
                chat.agents.firstWhere((a) => a.agentId == agentId),
              );
            }
            expect(
              await chat.send(text, const []),
              isTrue,
              reason: chat.sendError,
            );
            await until(
              () => chat.phase == ChatPhase.idle,
              reason: 'run ended',
            );
            expect(chat.notice, isNull, reason: chat.notice?.text);
            return chat.messages.last.text;
          }

          final first = await say('pi-run-$tag', 'hello', fresh: true);
          expect(first, contains('cwd=/srv/a'));
          expect(first, contains('model=anthropic/claude-fable-5'));
          final session = RegExp(
            r'session=([^,]+),',
          ).firstMatch(first)!.group(1);
          final second = await say('pi-run-$tag', 'again');
          expect(second, contains('session=$session,'));
          final other = await say('pi-run-$tag', 'elsewhere', fresh: true);
          expect(other, isNot(contains('session=$session,')));

          final cursor = await say('cursor-run-$tag', 'build', fresh: true);
          expect(cursor, contains('cwd=/srv/c'));
          expect(cursor, contains('mode=plan'));
          expect(cursor, contains('model=composer-2.5'));
        });

        test('capacity, an unreachable box and a missing Cursor key fail with '
            'actionable text', () async {
          final full = await api.createButterBox(
            token,
            workspace,
            name: 'failfull-$tag',
            baseUrl: boxUrl('capacity'),
            enabled: true,
          );
          final nokey = await api.createButterBox(
            token,
            workspace,
            name: 'failnokey-$tag',
            baseUrl: boxUrl('nokey'),
            enabled: true,
          );
          final dead = await api.createButterBox(
            token,
            workspace,
            name: 'faildead-$tag',
            baseUrl: boxUrl('ok'),
            enabled: true,
          );
          await saveLeaf(
            Agent(
              agentId: 'cursor-dead-$tag',
              name: 'cursor-dead-$tag',
              type: AgentType.AGENT_TYPE_CURSOR,
              config: AgentConfig(
                cursor: CursorAgentConfig(butterboxId: dead.id),
              ),
            ),
          );
          // The box moves out from under the saved leaf.
          await api.updateButterBox(
            token,
            workspace,
            id: dead.id,
            name: 'faildead-$tag',
            baseUrl: 'http://127.0.0.1:1',
            enabled: true,
          );
          await saveLeaf(
            Agent(
              agentId: 'pi-full-$tag',
              name: 'pi-full-$tag',
              type: AgentType.AGENT_TYPE_PI,
              config: AgentConfig(pi: PiAgentConfig(butterboxId: full.id)),
            ),
          );
          await saveLeaf(
            Agent(
              agentId: 'cursor-nokey-$tag',
              name: 'cursor-nokey-$tag',
              type: AgentType.AGENT_TYPE_CURSOR,
              config: AgentConfig(
                cursor: CursorAgentConfig(butterboxId: nokey.id),
              ),
            ),
          );

          final client = await signedIn();
          final chat = client.chat!;
          await chat.loadThreads();
          for (final (agentId, wanted) in [
            ('pi-full-$tag', 'capacity'),
            ('cursor-nokey-$tag', 'CURSOR_API_KEY'),
            ('cursor-dead-$tag', 'unreachable'),
          ]) {
            chat.startThread(
              chat.agents.firstWhere((a) => a.agentId == agentId),
            );
            expect(await chat.send('go', const []), isTrue);
            await until(
              () => chat.phase == ChatPhase.idle,
              reason: 'run ended',
            );
            expect(chat.notice?.kind, NoticeKind.failed, reason: agentId);
            expect(chat.notice!.text, contains(wanted), reason: agentId);
          }
        });
      },
    );
  }
}
