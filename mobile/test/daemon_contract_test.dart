import 'dart:convert';
import 'dart:io' as io;

import 'package:butter_mobile/core/butter_api.dart';
import 'package:butter_mobile/core/config_resource_controller.dart';
import 'package:butter_mobile/core/daemon_errors.dart';
import 'package:butter_mobile/core/server_address.dart';
import 'package:butter_mobile/features/daemons/daemon_controller.dart';
import 'package:butter_mobile/features/daemons/daemon_draft.dart';
import 'package:butter_mobile/features/remote_agents/remote_agent_controller.dart';
import 'package:butter_mobile/features/remote_agents/remote_agent_draft.dart';
import 'package:butter_mobile/gen/agents/v1/api_token.pb.dart';
import 'package:butter_mobile/gen/agents/v1/dashboard.pb.dart';
import 'package:flutter_test/flutter_test.dart';

// DaemonService, APITokenService (list and revoke), RemoteAgentService, the
// production runner with its daemon Bridge and the production worker poll
// transport are served by the contract fixture (mobile/tool/contractserver/
// daemon.go). The only double is the worker: a deterministic fake that speaks
// the real Register / Poll / ReportTaskUpdate / Unregister protocol with the
// worker token the App minted, driven over the fixture's loopback control API.
// No worker runs on the phone, no real opencode or codex runs, and nothing
// here is a device run.
const workspace = 'daemon';
const peerWorkspace = 'daemon-peer';
const memberWorkspace = 'mcp';
const deniedWorkspace = 'mcp-denied';

class Control {
  Control(this.base);
  final String base;

  Future<(int, Map<String, dynamic>)> call(
    String method,
    String path, [
    Map<String, Object?>? body,
  ]) async {
    final http = io.HttpClient();
    try {
      final request = await http.openUrl(
        method,
        Uri.parse('$base/fake-daemon$path'),
      );
      request.headers.contentType = io.ContentType.json;
      if (body != null) request.write(jsonEncode(body));
      final response = await request.close();
      final text = await utf8.decodeStream(response);
      final json = text.isEmpty
          ? <String, dynamic>{}
          : jsonDecode(text) as Map<String, dynamic>;
      return (response.statusCode, json);
    } finally {
      http.close(force: true);
    }
  }
}

Future<void> eventually(
  Future<bool> Function() check,
  String what, {
  Duration timeout = const Duration(seconds: 10),
}) async {
  final end = DateTime.now().add(timeout);
  while (DateTime.now().isBefore(end)) {
    if (await check()) return;
    await Future<void>.delayed(const Duration(milliseconds: 50));
  }
  fail('Timed out waiting for $what.');
}

void main() {
  const server = String.fromEnvironment('BUTTER_TEST_SERVER');
  for (final prefix in ['', '/butter']) {
    group(
      'Daemon production contract $prefix',
      skip: server.isEmpty ? 'Run tool/contract.sh.' : false,
      () {
        final base = '$server$prefix';
        final tag = prefix.isEmpty ? 'root' : 'prefix';
        final control = Control(server);
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

        DaemonController daemons({String ws = workspace}) {
          final c = DaemonController(
            newApi: () => ButterApi(ServerAddress.parse(base, allowHttp: true)),
            token: token,
            workspaceId: ws,
            onUnauthenticated: () {},
          );
          addTearDown(c.dispose);
          return c;
        }

        RemoteAgentController remotes({String ws = workspace}) {
          final c = RemoteAgentController(
            newApi: () => ButterApi(ServerAddress.parse(base, allowHttp: true)),
            token: token,
            workspaceId: ws,
            onUnauthenticated: () {},
          );
          addTearDown(c.dispose);
          return c;
        }

        /// A runtime plus its minted token, as the App does it.
        Future<String> runtimeWithSecret(
          DaemonController c,
          String id, {
          String ws = workspace,
        }) async {
          final draft = DaemonRuntimeDraft.create()
            ..id = id
            ..name = 'Runtime $id';
          expect(await c.create(draft.build()), isTrue, reason: c.writeError);
          expect(await c.mint(id, 'Worker $id'), isTrue, reason: c.writeError);
          final secret = c.minted!.secret!;
          c.dismissSecret();
          return secret;
        }

        Future<void> startWorker(
          String ws,
          String runtime,
          String secret, {
          String mode = 'stubborn',
        }) async {
          final (code, body) = await control.call('POST', '/worker/start', {
            'workspace': ws,
            'runtime_id': runtime,
            'token': secret,
            'mode': mode,
          });
          expect(code, 200, reason: '$body');
          addTearDown(
            () => control.call('POST', '/worker/drop', {
              'workspace': ws,
              'runtime_id': runtime,
            }),
          );
        }

        Future<void> daemonAgent(
          RemoteAgentController r,
          String id,
          String runtimeId,
        ) async {
          final draft = RemoteAgentDraft.create()
            ..id = id
            ..name = 'Agent $id'
            ..kind = RemoteKind.daemon
            ..daemonRuntimeId = runtimeId
            ..acpRuntime = 'codex';
          expect(draft.firstError(), isNull);
          expect(await r.create(draft.build()), isTrue, reason: r.writeError);
        }

        Future<String> dispatch(String ws, String remoteAgent) async {
          final (code, body) = await control.call('POST', '/dispatch', {
            'workspace': ws,
            'remote_agent_id': remoteAgent,
            'input': 'build it',
          });
          expect(code, 200, reason: '$body');
          return body['run_id'] as String;
        }

        Future<DaemonTaskInFlight> waitForTask(
          DaemonController c,
          String runtimeId,
        ) async {
          DaemonTaskInFlight? found;
          await eventually(() async {
            await c.loadTasks();
            found = c.tasks
                .where((t) => t.daemonRuntimeId == runtimeId)
                .firstOrNull;
            return found != null;
          }, 'a task on $runtimeId');
          return found!;
        }

        test(
          'runtimes round-trip, keep labels, and never cross Workspaces',
          () async {
            final c = daemons();
            final seeded = DaemonRuntime(id: 'crud-$tag', name: 'Crud')
              ..labels['team'] = 'platform';
            await api.createDaemonRuntime(token, workspace, seeded);
            await c.load();
            expect(c.items.map((r) => r.id), contains('crud-$tag'));
            expect(c.items.every((r) => r.workspaceId == workspace), isTrue);

            await c.loadItem('crud-$tag');
            final draft = DaemonRuntimeDraft.edit(c.item!)..name = 'Renamed';
            expect(await c.update(draft.build()), isTrue, reason: c.writeError);
            final after = await api.daemonRuntime(
              token,
              workspace,
              'crud-$tag',
            );
            expect(after.name, 'Renamed');
            expect(after.labels['team'], 'platform');

            // Labels edited as rows replace the stored map as a whole.
            await c.loadItem('crud-$tag');
            final relabel = DaemonRuntimeDraft.edit(c.item!)
              ..labels.removeWhere((r) => r.key == 'team')
              ..labels.add(LabelRow(key: 'region', value: 'eu'));
            expect(
              await c.update(relabel.build()),
              isTrue,
              reason: c.writeError,
            );
            final relabelled = await api.daemonRuntime(
              token,
              workspace,
              'crud-$tag',
            );
            expect(relabelled.labels, {'region': 'eu'});
            expect(relabelled.name, 'Renamed');

            // A duplicate ID is a definite refusal.
            final dup = DaemonRuntimeDraft.create()
              ..id = 'crud-$tag'
              ..name = 'Again';
            expect(await c.create(dup.build()), isFalse);
            expect(c.writeError, contains('already exists'));
            expect(c.unconfirmed, isFalse);

            // The peer Workspace does not see it, nor can it read or delete it.
            final peer = daemons(ws: peerWorkspace);
            await peer.load();
            expect(peer.items.map((r) => r.id), isNot(contains('crud-$tag')));
            await peer.loadItem('crud-$tag');
            expect(peer.itemState, ConfigLoad.failed);
            expect(await peer.delete('crud-$tag'), isFalse);
            await c.load();
            expect(c.items.map((r) => r.id), contains('crud-$tag'));

            expect(await c.delete('crud-$tag'), isTrue, reason: c.writeError);
            await c.load();
            expect(c.items.map((r) => r.id), isNot(contains('crud-$tag')));
          },
        );

        test(
          'a worker token expires only when an expiry was asked for, down to '
          'half an hour',
          () async {
            final c = daemons();
            final id = 'ttl-$tag';
            await api.createDaemonRuntime(
              token,
              workspace,
              DaemonRuntime(id: id, name: 'Ttl'),
            );
            addTearDown(() => api.deleteDaemonRuntime(token, workspace, id));
            Future<APIToken> mint(String name, {Duration? ttl}) async {
              final before = DateTime.now().toUtc();
              expect(
                await c.mint(id, name, ttl: ttl),
                isTrue,
                reason: c.writeError,
              );
              final minted = c.minted!.token;
              c.dismissSecret();
              if (ttl != null) {
                expect(minted.hasExpiresAt(), isTrue, reason: name);
                final left = minted.expiresAt.toDateTime().difference(before);
                expect(left, greaterThan(ttl - const Duration(seconds: 5)));
                expect(left, lessThan(ttl + const Duration(seconds: 5)));
              }
              return minted;
            }

            expect((await mint('forever')).hasExpiresAt(), isFalse);
            await mint('half an hour', ttl: const Duration(minutes: 30));
            await mint('an hour', ttl: const Duration(hours: 1));
            await mint('twelve hours', ttl: const Duration(hours: 12));
            await mint('a week', ttl: const Duration(days: 7));
          },
        );

        test(
          'a Workspace the account is not in is refused with a safe message',
          () async {
            final c = daemons(ws: deniedWorkspace);
            await c.load();
            expect(c.listState, ConfigLoad.failed);
            expect(c.listError, contains('access'));
            await c.loadOnline();
            expect(c.onlineState, ConfigLoad.failed);
          },
        );

        test(
          'a minted worker token connects its worker and is listed without a secret',
          () async {
            final c = daemons();
            final id = 'tok-$tag';
            final draft = DaemonRuntimeDraft.create()
              ..id = id
              ..name = 'Tok';
            await c.create(draft.build());
            expect(await c.mint(id, 'Laptop'), isTrue, reason: c.writeError);
            final minted = c.minted!;
            expect(minted.secret, isNotEmpty);
            expect(minted.token.kind, APITokenKind.API_TOKEN_KIND_DAEMON);
            expect(minted.token.daemonRuntimeId, id);
            final secret = minted.secret!;
            c.dismissSecret();
            expect(minted.secret, isNull);

            await startWorker(workspace, id, secret);
            await eventually(() async {
              await c.loadOnline();
              return c.workerOf(id) != null;
            }, 'the worker to appear online');
            final worker = c.workerOf(id)!;
            expect(worker.name, 'fake-worker');
            expect(worker.executors, ['fake']);
            expect(worker.acpRuntimes, containsAll(['opencode', 'codex']));
            expect(worker.version, isNotEmpty);
            expect(worker.state, isNot(DaemonStatus_State.STATE_OFFLINE));

            await c.loadTokens();
            final listed = c.tokensOf(id).single;
            expect(listed.id, minted.token.id);
            expect(listed.kind, APITokenKind.API_TOKEN_KIND_DAEMON);
            expect(listed.toProto3Json().toString(), isNot(contains(secret)));
            // Every token here is a worker token of some runtime.
            expect(
              c.workerTokens.every(
                (t) => t.kind == APITokenKind.API_TOKEN_KIND_DAEMON,
              ),
              isTrue,
            );

            // The owner revokes it; the revoked secret registers no worker.
            expect(await c.revoke(listed.id), isTrue, reason: c.writeError);
            expect(c.tokensOf(id).single.revoked, isTrue);
            final (code, _) = await control.call('POST', '/worker/start', {
              'workspace': workspace,
              'runtime_id': 'revoked-$tag',
              'token': secret,
              'mode': 'stubborn',
            });
            expect(code, 401);
          },
        );

        test(
          'a plain member may list worker tokens but not revoke them',
          () async {
            final c = daemons(ws: memberWorkspace);
            final id = 'member-$tag';
            final draft = DaemonRuntimeDraft.create()
              ..id = id
              ..name = 'Member';
            expect(await c.create(draft.build()), isTrue, reason: c.writeError);
            expect(await c.mint(id, 'w'), isTrue, reason: c.writeError);
            final tokenId = c.minted!.token.id;
            c.dismissSecret();
            await c.loadTokens();
            expect(c.tokensOf(id).single.id, tokenId);
            expect(await c.revoke(tokenId), isFalse);
            expect(c.writeError, contains('owner or admin'));
            expect(c.unconfirmed, isFalse);
            await c.loadTokens();
            expect(c.tokensOf(id).single.revoked, isFalse);
          },
        );

        test(
          'a Daemon Remote Agent dispatches through the worker; cancel is delivery only',
          () async {
            final d = daemons();
            final r = remotes();
            final runtime = 'run-$tag';
            final secret = await runtimeWithSecret(d, runtime);
            await startWorker(workspace, runtime, secret);
            await r.loadRuntimes();
            final agentId = 'ra-run-$tag';
            await daemonAgent(r, agentId, runtime);
            expect(r.referenceProblem(r.byId(agentId)!), isNull);
            await r.probe(agentId);
            expect(r.probes[agentId]!.kind, RemoteProbeKind.reachable);

            final runId = await dispatch(workspace, agentId);
            final task = await waitForTask(d, runtime);
            expect(task.acpRuntime, 'codex');
            expect(task.agentName, 'Agent $agentId');

            await control.call('POST', '/task/progress', {
              'workspace': workspace,
              'runtime_id': runtime,
              'task_id': task.taskId,
              'step': 'Compiling',
              'progress': 40,
            });
            await eventually(() async {
              await d.loadTasks();
              final t = d.tasks
                  .where((t) => t.taskId == task.taskId)
                  .firstOrNull;
              return t != null &&
                  t.currentStep == 'Compiling' &&
                  t.progress == 40;
            }, 'progress in the list');
            expect(
              d.tasks.single.elapsed.seconds.toInt(),
              greaterThanOrEqualTo(0),
            );

            await d.cancel(d.tasks.single);
            expect(d.cancels[task.taskId]!.kind, CancelKind.delivered);
            expect(d.cancels[task.taskId]!.message, contains('still running'));
            await eventually(() async {
              final (_, state) = await control.call(
                'GET',
                '/worker/state?workspace=$workspace&runtime_id=$runtime',
              );
              return (state['cancels'] as List).contains(task.taskId);
            }, 'the worker to see the cancel');
            // A stubborn worker keeps going: still listed, run not over.
            await Future<void>.delayed(const Duration(milliseconds: 300));
            await d.loadTasks();
            expect(d.tasks.map((t) => t.taskId), contains(task.taskId));
            final (_, running) = await control.call('GET', '/run?id=$runId');
            expect(running['done'], isFalse);

            // The worker's own report ends it; the App learns only that it left.
            await control.call('POST', '/task/fail', {
              'workspace': workspace,
              'runtime_id': runtime,
              'task_id': task.taskId,
              'error': 'stopped by worker',
            });
            await eventually(() async {
              await d.loadTasks();
              return d.tasks.every((t) => t.taskId != task.taskId);
            }, 'the task to leave the list');
            expect(d.ended.first.task.taskId, task.taskId);
            expect(d.ended.first.cancelDelivered, isTrue);
            expect(d.cancels.containsKey(task.taskId), isFalse);
          },
        );

        test(
          'cancel of a task whose worker is gone, or that belongs to another Workspace, delivers nothing',
          () async {
            final d = daemons();
            final peer = daemons(ws: peerWorkspace);
            final rWs = remotes();
            final runtime = 'gone-$tag';
            final secret = await runtimeWithSecret(d, runtime);
            await startWorker(workspace, runtime, secret);
            await rWs.loadRuntimes();
            final agentId = 'ra-gone-$tag';
            await daemonAgent(rWs, agentId, runtime);
            await dispatch(workspace, agentId);
            final task = await waitForTask(d, runtime);

            // The peer Workspace neither sees nor can cancel it.
            await peer.loadTasks();
            expect(peer.tasks.where((t) => t.taskId == task.taskId), isEmpty);
            await peer.cancel(task);
            expect(peer.cancels[task.taskId]!.kind, CancelKind.gone);
            await Future<void>.delayed(const Duration(milliseconds: 200));
            final (_, state) = await control.call(
              'GET',
              '/worker/state?workspace=$workspace&runtime_id=$runtime',
            );
            expect(state['cancels'], isEmpty);

            // The worker leaves; its task goes with it and a cancel finds nothing.
            await control.call('POST', '/worker/stop', {
              'workspace': workspace,
              'runtime_id': runtime,
            });
            await eventually(() async {
              await d.loadOnline();
              return d.workerOf(runtime) == null;
            }, 'the worker to go offline');
            await d.cancel(task);
            expect(d.cancels[task.taskId]!.kind, CancelKind.gone);
          },
        );

        test(
          'deleting a runtime is unguarded: the references are found first and then flagged',
          () async {
            final d = daemons();
            final r = remotes();
            final runtime = 'del-$tag';
            await runtimeWithSecret(d, runtime);
            await r.loadRuntimes();
            final agentId = 'ra-del-$tag';
            await daemonAgent(r, agentId, runtime);

            final users = await d.remoteAgentsOn(runtime);
            expect(users!.map((a) => a.id), [agentId]);
            expect(await d.delete(runtime), isTrue, reason: d.writeError);

            await r.loadRuntimes();
            await r.load();
            expect(
              r.referenceProblem(r.byId(agentId)!),
              contains('does not exist in this Workspace'),
            );
            // The server still holds the agent, and says the worker is not there.
            await r.probe(agentId);
            expect(r.probes[agentId]!.kind, isNot(RemoteProbeKind.reachable));
            // Its worker token stays listed until someone revokes it.
            await d.loadTokens();
            expect(d.tokensOf(runtime), isNotEmpty);
          },
        );

        test(
          'a Daemon agent may not name a runtime of another Workspace',
          () async {
            final d = daemons(ws: peerWorkspace);
            final runtime = 'peer-only-$tag';
            await runtimeWithSecret(d, runtime, ws: peerWorkspace);
            final r = remotes();
            await r.loadRuntimes();
            final draft = RemoteAgentDraft.create()
              ..id = 'ra-cross-$tag'
              ..name = 'Cross'
              ..kind = RemoteKind.daemon
              ..daemonRuntimeId = runtime
              ..acpRuntime = 'codex';
            expect(r.runtimes!.map((x) => x.id), isNot(contains(runtime)));
            expect(await r.create(draft.build()), isFalse);
            expect(r.writeError, isNotNull);
            expect(r.unconfirmed, isFalse);
          },
        );

        test('diagnostics are the server\'s own', () async {
          final c = daemons();
          await c.loadDiagnostics();
          expect(
            c.diagnosticsState,
            ConfigLoad.ready,
            reason: c.diagnosticsError,
          );
          expect(c.diagnostics!.goroutines, greaterThan(0));
        });
      },
    );
  }
}
