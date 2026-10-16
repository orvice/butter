import 'dart:async';
import 'dart:io';

import 'package:butter_mobile/app/client_context.dart';
import 'package:butter_mobile/core/config_resource_controller.dart';
import 'package:butter_mobile/core/daemon_errors.dart';
import 'package:butter_mobile/features/daemons/daemon_controller.dart';
import 'package:butter_mobile/features/daemons/daemon_draft.dart';
import 'package:butter_mobile/gen/agents/v1/agent.pb.dart';
import 'package:butter_mobile/gen/agents/v1/api_token.pb.dart';
import 'package:butter_mobile/gen/agents/v1/dashboard.pb.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:protobuf/well_known_types/google/protobuf/duration.pb.dart'
    as pbd;
import 'package:fixnum/fixnum.dart';

import 'support/daemon_server.dart';
import 'support/fake_oauth_platform.dart';
import 'support/memory_storage.dart';

DaemonRuntime runtime(String id, {String name = ''}) =>
    DaemonRuntime(id: id, name: name.isEmpty ? 'Runtime $id' : name);

DaemonTaskInFlight task(
  String id, {
  String runtimeId = 'rt1',
  int progress = 10,
  String step = 'cloning',
}) => DaemonTaskInFlight(
  taskId: id,
  daemonRuntimeId: runtimeId,
  daemonName: 'Runtime $runtimeId',
  acpRuntime: 'codex',
  agentName: 'Remote',
  currentStep: step,
  progress: progress,
  elapsed: pbd.Duration(seconds: Int64(75)),
);

class Harness {
  Harness(this.client, this.server);
  final ClientContext client;
  final DaemonServerFake server;
  DaemonController get daemons => client.daemons!;
}

Future<Harness> start({void Function(DaemonServerFake)? seed}) async {
  final old = HttpOverrides.current;
  HttpOverrides.global = null;
  final server = DaemonServerFake();
  seed?.call(server);
  await server.start();
  final client = ClientContext(
    storage: MemoryStorage(),
    allowHttp: true,
    oauth: FakeOAuthPlatform(),
  );
  addTearDown(() async {
    client.dispose();
    await server.close();
    HttpOverrides.global = old;
  });
  await client.setServer(server.url);
  await client.login('alice', 'password');
  await client.selectWorkspace('alpha');
  return Harness(client, server);
}

void main() {
  test('lists and reads the active Workspace only', () async {
    final h = await start(
      seed: (s) {
        s.seedRuntime('alpha', runtime('a1'));
        s.seedRuntime('beta', runtime('b1'));
      },
    );
    await h.daemons.load();
    expect(h.daemons.listState, ConfigLoad.ready);
    expect(h.daemons.items.map((r) => r.id), ['a1']);
    await h.daemons.loadItem('a1');
    expect(h.daemons.item?.name, 'Runtime a1');
    expect(h.server.daemonCalls.map((c) => c.workspace).toSet(), {'alpha'});
  });

  test('create refuses a duplicate ID without calling it unknown', () async {
    final h = await start();
    final draft = DaemonRuntimeDraft.create()
      ..id = 'rt1'
      ..name = 'One';
    expect(await h.daemons.create(draft.build()), isTrue);
    expect(h.server.runtimes['alpha']!['rt1']!.name, 'One');
    expect(await h.daemons.create(draft.build()), isFalse);
    expect(h.daemons.writeError, contains('already exists'));
    expect(h.daemons.unconfirmed, isFalse);
  });

  test(
    'an edit carries labels and every field the form does not touch',
    () async {
      final h = await start(
        seed: (s) {
          final r = runtime('rt1')
            ..labels['team'] = 'platform'
            ..createdBy = 'bob';
          s.seedRuntime('alpha', r);
        },
      );
      await h.daemons.loadItem('rt1');
      final draft = DaemonRuntimeDraft.edit(h.daemons.item!)..name = 'Renamed';
      expect(await h.daemons.update(draft.build()), isTrue);
      final stored = h.server.runtimes['alpha']!['rt1']!;
      expect(stored.name, 'Renamed');
      expect(stored.labels['team'], 'platform');
      expect(stored.createdBy, 'bob');
    },
  );

  test(
    'a lost create reply is unconfirmed, never repeated, settled by a read',
    () async {
      final h = await start();
      h.server.lostReplies.add('CreateDaemonRuntime');
      final draft = DaemonRuntimeDraft.create()
        ..id = 'rt1'
        ..name = 'One';
      expect(await h.daemons.create(draft.build()), isFalse);
      expect(h.daemons.unconfirmed, isTrue);
      expect(await h.daemons.create(draft.build()), isFalse);
      expect(h.server.daemonCount('CreateDaemonRuntime'), 1);
      await h.daemons.load();
      expect(h.daemons.unconfirmed, isFalse);
      expect(h.daemons.items.map((r) => r.id), ['rt1']);
    },
  );

  test('connected workers are keyed by runtime; offline is absence', () async {
    final h = await start(
      seed: (s) {
        s.onlineWorkers['alpha']!.add(
          DaemonStatus(
            daemonRuntimeId: 'rt1',
            name: 'Laptop',
            executors: ['codex'],
            acpRuntimes: ['opencode', 'codex'],
            state: DaemonStatus_State.STATE_ONLINE,
            activeTasks: 1,
          ),
        );
        s.onlineWorkers['beta']!.add(DaemonStatus(daemonRuntimeId: 'rt9'));
      },
    );
    await h.daemons.loadOnline();
    expect(h.daemons.onlineState, ConfigLoad.ready);
    expect(h.daemons.workerOf('rt1')?.name, 'Laptop');
    expect(h.daemons.workerOf('rt2'), isNull);
    expect(h.daemons.workerOf('rt9'), isNull);
  });

  test(
    'tasks show step, progress and elapsed; a vanished task is noted, not guessed',
    () async {
      final h = await start(
        seed: (s) => s.taskList['alpha']!.addAll([task('t1'), task('t2')]),
      );
      await h.daemons.loadTasks();
      expect(h.daemons.tasks.map((t) => t.taskId), ['t1', 't2']);
      expect(h.daemons.tasks.first.currentStep, 'cloning');
      expect(h.daemons.tasks.first.elapsed.seconds.toInt(), 75);
      h.server.taskList['alpha']!.removeWhere((t) => t.taskId == 't1');
      await h.daemons.loadTasks();
      expect(h.daemons.tasks.map((t) => t.taskId), ['t2']);
      expect(h.daemons.ended.single.task.taskId, 't1');
      expect(h.daemons.ended.single.cancelDelivered, isFalse);
    },
  );

  test('a failed refresh keeps the last list and says it is old', () async {
    final h = await start(seed: (s) => s.taskList['alpha']!.add(task('t1')));
    await h.daemons.loadTasks();
    h.server.daemonRefusals['ListDaemonTasks'] = (503, 'unavailable', 'down');
    await h.daemons.loadTasks();
    expect(h.daemons.tasks.map((t) => t.taskId), ['t1']);
    expect(h.daemons.taskError, isNotNull);
    expect(h.daemons.tasksReadAt, isNotNull);
  });

  test('diagnostics are read and a refusal is reported', () async {
    final h = await start();
    await h.daemons.loadDiagnostics();
    expect(h.daemons.diagnostics?.goroutines, 42);
    h.server.daemonRefusals['GetBridgeDiagnostics'] = (
      403,
      'permission_denied',
      'zzz-server-text',
    );
    await h.daemons.loadDiagnostics();
    expect(h.daemons.diagnosticsState, ConfigLoad.failed);
    expect(h.daemons.diagnosticsError, isNotNull);
    expect(h.daemons.diagnosticsError, isNot(contains('zzz-server-text')));
  });

  group('worker tokens', () {
    test(
      'only worker tokens are listed; user tokens never reach the app state',
      () async {
        final h = await start(
          seed: (s) {
            s.seedRuntime('alpha', runtime('rt1'));
            s.tokens['alpha']!.addAll([
              APIToken(
                id: 'u1',
                name: 'my laptop cli',
                prefix: 'btr_u',
                kind: APITokenKind.API_TOKEN_KIND_USER,
              ),
              APIToken(
                id: 'w1',
                name: 'worker',
                prefix: 'btr_w',
                kind: APITokenKind.API_TOKEN_KIND_DAEMON,
                daemonRuntimeId: 'rt1',
              ),
            ]);
          },
        );
        await h.daemons.loadTokens();
        expect(h.daemons.workerTokens.map((t) => t.id), ['w1']);
        expect(h.daemons.tokensOf('rt1').single.id, 'w1');
        expect(h.daemons.tokensOf('rt2'), isEmpty);
      },
    );

    test(
      'mint shows the secret once; dismissing wipes it and the token stays by prefix',
      () async {
        final h = await start(
          seed: (s) => s.seedRuntime('alpha', runtime('rt1')),
        );
        expect(await h.daemons.mint('rt1', 'Laptop worker'), isTrue);
        final m = h.daemons.minted!;
        expect(m.secret, 'worker-secret-1');
        expect(m.token.kind, APITokenKind.API_TOKEN_KIND_DAEMON);
        expect(m.token.daemonRuntimeId, 'rt1');
        // Another secret cannot replace the one on screen unseen.
        expect(await h.daemons.mint('rt1', 'second'), isFalse);
        expect(h.server.daemonCount('CreateDaemonRuntimeToken'), 1);
        h.daemons.dismissSecret();
        expect(h.daemons.minted, isNull);
        expect(m.secret, isNull);
        expect(h.daemons.tokensOf('rt1').single.prefix, 'btr_w1');
        // The secret is not part of any stored token field.
        expect(
          h.daemons.workerTokens.single.toString(),
          isNot(contains('worker-secret')),
        );
      },
    );

    test(
      'a lost mint reply is never repeated and never shows a secret',
      () async {
        final h = await start(
          seed: (s) => s.seedRuntime('alpha', runtime('rt1')),
        );
        h.server.lostReplies.add('CreateDaemonRuntimeToken');
        expect(await h.daemons.mint('rt1', 'w'), isFalse);
        expect(h.daemons.minted, isNull);
        expect(h.daemons.unconfirmed, isTrue);
        expect(h.daemons.writeNotice, contains('never shown'));
        expect(await h.daemons.mint('rt1', 'w'), isFalse);
        expect(h.server.daemonCount('CreateDaemonRuntimeToken'), 1);
        // A stored runtime edit is also held back until the tokens were read.
        expect(await h.daemons.delete('rt1'), isFalse);
        expect(h.server.daemonCount('DeleteDaemonRuntime'), 0);
        await h.daemons.loadTokens();
        expect(h.daemons.unconfirmed, isFalse);
        expect(h.daemons.tokensOf('rt1'), hasLength(1));
        // The person may now decide: revoke the one they cannot identify.
        expect(await h.daemons.revoke('tok-1'), isTrue);
        expect(h.daemons.tokensOf('rt1').single.revoked, isTrue);
      },
    );

    test(
      'suspend wipes the secret and marks an in-flight mint unconfirmed',
      () async {
        final gate = Completer<void>();
        final h = await start(
          seed: (s) => s.seedRuntime('alpha', runtime('rt1')),
        );
        h.server.daemonBefore = (m) async {
          if (m == 'CreateDaemonRuntimeToken') await gate.future;
        };
        final pending = h.daemons.mint('rt1', 'w');
        await Future<void>.delayed(const Duration(milliseconds: 100));
        h.daemons.suspend();
        gate.complete();
        expect(await pending, isFalse);
        expect(h.daemons.minted, isNull);
        expect(h.daemons.unconfirmed, isTrue);
        await h.daemons.recover();
        expect(h.daemons.unconfirmed, isFalse);
        expect(h.daemons.minted, isNull);
        expect(h.server.daemonCount('CreateDaemonRuntimeToken'), 1);
      },
    );

    test('suspend after a mint drops the secret from memory', () async {
      final h = await start(
        seed: (s) => s.seedRuntime('alpha', runtime('rt1')),
      );
      await h.daemons.mint('rt1', 'w');
      final m = h.daemons.minted!;
      h.daemons.suspend();
      expect(h.daemons.minted, isNull);
      expect(m.secret, isNull);
    });

    test('leaving the Workspace wipes the secret', () async {
      final h = await start(
        seed: (s) => s.seedRuntime('alpha', runtime('rt1')),
      );
      await h.daemons.mint('rt1', 'w');
      final controller = h.daemons;
      final m = controller.minted!;
      await h.client.selectWorkspace('beta');
      expect(controller.disposed, isTrue);
      expect(m.secret, isNull);
      expect(controller.minted, isNull);
    });

    test('minting for a deleted runtime says so and creates nothing', () async {
      final h = await start();
      expect(await h.daemons.mint('ghost', 'w'), isFalse);
      expect(h.daemons.writeError, contains('no longer exists'));
      expect(h.daemons.unconfirmed, isFalse);
      expect(h.server.tokens['alpha'], isEmpty);
    });

    test(
      'revoke refused for a member names the role; nothing changes',
      () async {
        final h = await start(
          seed: (s) => s.tokens['alpha']!.add(
            APIToken(
              id: 'w1',
              kind: APITokenKind.API_TOKEN_KIND_DAEMON,
              daemonRuntimeId: 'rt1',
            ),
          ),
        );
        await h.daemons.loadTokens();
        h.server.daemonRefusals['RevokeAPIToken'] = (
          403,
          'permission_denied',
          'forbidden',
        );
        expect(await h.daemons.revoke('w1'), isFalse);
        expect(h.daemons.writeError, contains('owner or admin'));
        expect(h.daemons.workerTokens.single.revoked, isFalse);
        expect(h.daemons.unconfirmed, isFalse);
      },
    );
  });

  group('cancel', () {
    test('success is delivery only: the task stays listed', () async {
      final h = await start(seed: (s) => s.taskList['alpha']!.add(task('t1')));
      await h.daemons.loadTasks();
      await h.daemons.cancel(h.daemons.tasks.single);
      final view = h.daemons.cancels['t1']!;
      expect(view.kind, CancelKind.delivered);
      expect(view.message, contains('still running'));
      expect(h.daemons.tasks.single.taskId, 't1');
      // The worker later reports its end; the task leaves the list and the
      // outcome stays unknown to the app.
      h.server.taskList['alpha']!.clear();
      await h.daemons.loadTasks();
      expect(h.daemons.ended.single.cancelDelivered, isTrue);
      expect(h.daemons.cancels, isEmpty);
    });

    test(
      'a task that is gone, a worker that cannot be reached and a refusal are told apart',
      () async {
        final h = await start(
          seed: (s) => s.taskList['alpha']!.add(task('t1')),
        );
        await h.daemons.loadTasks();
        final t = h.daemons.tasks.single;
        h.server.daemonRefusals['CancelDaemonTask'] = (404, 'not_found', 'x');
        await h.daemons.cancel(t);
        expect(h.daemons.cancels['t1']!.kind, CancelKind.gone);
        h.server.daemonRefusals['CancelDaemonTask'] = (
          500,
          'internal',
          'worker down',
        );
        await h.daemons.cancel(t);
        expect(h.daemons.cancels['t1']!.kind, CancelKind.workerUnreachable);
        expect(
          h.daemons.cancels['t1']!.message,
          isNot(contains('worker down')),
        );
        h.server.daemonRefusals['CancelDaemonTask'] = (
          403,
          'permission_denied',
          'x',
        );
        await h.daemons.cancel(t);
        expect(h.daemons.cancels['t1']!.kind, CancelKind.refused);
      },
    );

    test('a lost reply is unknown and may be sent again', () async {
      final h = await start(seed: (s) => s.taskList['alpha']!.add(task('t1')));
      await h.daemons.loadTasks();
      h.server.lostReplies.add('CancelDaemonTask');
      await h.daemons.cancel(h.daemons.tasks.single);
      expect(h.daemons.cancels['t1']!.kind, CancelKind.unknown);
      h.server.lostReplies.clear();
      await h.daemons.cancel(h.daemons.tasks.single);
      expect(h.daemons.cancels['t1']!.kind, CancelKind.delivered);
    });

    test(
      'a cancel in flight when the app leaves is reported as not confirmed',
      () async {
        final gate = Completer<void>();
        final h = await start(
          seed: (s) => s.taskList['alpha']!.add(task('t1')),
        );
        await h.daemons.loadTasks();
        h.server.daemonBefore = (m) async {
          if (m == 'CancelDaemonTask') await gate.future;
        };
        final pending = h.daemons.cancel(h.daemons.tasks.single);
        await Future<void>.delayed(const Duration(milliseconds: 100));
        expect(h.daemons.cancels['t1']!.sending, isTrue);
        h.daemons.suspend();
        gate.complete();
        await pending;
        expect(h.daemons.cancels['t1']!.kind, CancelKind.unknown);
        await h.daemons.recover();
        expect(h.daemons.tasks.single.taskId, 't1');
        expect(h.daemons.cancels['t1']!.kind, CancelKind.unknown);
      },
    );
  });

  test(
    'remoteAgentsOn finds the Daemon agents of a runtime in this Workspace',
    () async {
      final h = await start(
        seed: (s) {
          s.seedRemote(
            'alpha',
            RemoteAgent(
              id: 'ra1',
              name: 'On rt1',
              protocol: RemoteAgentProtocol.REMOTE_AGENT_PROTOCOL_DAEMON,
              daemonRuntimeId: 'rt1',
            ),
          );
          s.seedRemote(
            'alpha',
            RemoteAgent(
              id: 'ra2',
              name: 'On rt2',
              protocol: RemoteAgentProtocol.REMOTE_AGENT_PROTOCOL_DAEMON,
              daemonRuntimeId: 'rt2',
            ),
          );
          s.seedRemote(
            'alpha',
            RemoteAgent(
              id: 'ra3',
              name: 'A2A',
              protocol: RemoteAgentProtocol.REMOTE_AGENT_PROTOCOL_A2A,
              url: 'https://x.example.com',
            ),
          );
          s.seedRemote(
            'beta',
            RemoteAgent(
              id: 'rb1',
              protocol: RemoteAgentProtocol.REMOTE_AGENT_PROTOCOL_DAEMON,
              daemonRuntimeId: 'rt1',
            ),
          );
        },
      );
      final found = await h.daemons.remoteAgentsOn('rt1');
      expect(found!.map((a) => a.id), ['ra1']);
      h.server.refusals['ListRemoteAgents'] = (503, 'unavailable', 'down');
      expect(await h.daemons.remoteAgentsOn('rt1'), isNull);
    },
  );

  group('task reads', () {
    test('overlapping reads share the one in flight', () async {
      final gate = Completer<void>();
      final h = await start(seed: (s) => s.taskList['alpha']!.add(task('t1')));
      h.server.daemonBefore = (m) async {
        if (m == 'ListDaemonTasks') await gate.future;
      };
      final a = h.daemons.loadTasks();
      final b = h.daemons.loadTasks();
      final c = h.daemons.loadTasks();
      await Future<void>.delayed(const Duration(milliseconds: 100));
      expect(h.server.daemonCount('ListDaemonTasks'), 1);
      gate.complete();
      await Future.wait([a, b, c]);
      expect(h.daemons.tasks.single.taskId, 't1');
      // Once it settled, the next read is a new request.
      await h.daemons.loadTasks();
      expect(h.server.daemonCount('ListDaemonTasks'), 2);
    });

    test('a read begun before the app left is not reused after it', () async {
      final gate = Completer<void>();
      final h = await start(seed: (s) => s.taskList['alpha']!.add(task('t1')));
      var held = false;
      h.server.daemonBefore = (m) async {
        if (m == 'ListDaemonTasks' && !held) {
          held = true;
          await gate.future;
        }
      };
      final stale = h.daemons.loadTasks();
      await Future<void>.delayed(const Duration(milliseconds: 100));
      h.daemons.suspend();
      await h.daemons.recover();
      expect(h.daemons.tasks.single.taskId, 't1');
      gate.complete();
      await stale;
      expect(h.server.daemonCount('ListDaemonTasks'), 2);
    });
  });

  group('minting a worker token', () {
    test(
      'sends the expiry it was given, and none when there is none',
      () async {
        final h = await start(
          seed: (s) => s.seedRuntime('alpha', runtime('rt1')),
        );
        expect(await h.daemons.mint('rt1', 'forever'), isTrue);
        h.daemons.dismissSecret();
        expect(
          await h.daemons.mint('rt1', 'a day', ttl: const Duration(hours: 24)),
          isTrue,
        );
        final bodies = [
          for (final c in h.server.daemonCalls)
            if (c.method == 'CreateDaemonRuntimeToken')
              CreateDaemonRuntimeTokenRequest.fromBuffer(c.body),
        ];
        expect(bodies.map((b) => b.name), ['forever', 'a day']);
        expect(bodies[0].hasTtl(), isFalse);
        expect(bodies[1].ttl.seconds.toInt(), 86400);
      },
    );

    test('leaving the screen wipes the secret that was shown', () async {
      final h = await start(
        seed: (s) => s.seedRuntime('alpha', runtime('rt1')),
      );
      await h.daemons.mint('rt1', 'w');
      final m = h.daemons.minted!;
      h.daemons.leaveSecretScreen('rt1', notify: false);
      expect(h.daemons.minted, isNull);
      expect(m.secret, isNull);
    });

    test(
      'a reply that arrives after the screen was left is not kept',
      () async {
        final gate = Completer<void>();
        final h = await start(
          seed: (s) => s.seedRuntime('alpha', runtime('rt1')),
        );
        h.server.daemonBefore = (m) async {
          if (m == 'CreateDaemonRuntimeToken') await gate.future;
        };
        final pending = h.daemons.mint('rt1', 'w');
        await Future<void>.delayed(const Duration(milliseconds: 100));
        h.daemons.leaveSecretScreen('rt1', notify: false);
        gate.complete();
        expect(await pending, isFalse);
        expect(h.daemons.minted, isNull);
        // The token exists but its secret is gone: that must be said, and no
        // second token is made until the person has looked at the list.
        expect(h.daemons.unconfirmed, isTrue);
        expect(h.daemons.writeError, contains('discarded'));
        expect(await h.daemons.mint('rt1', 'again'), isFalse);
        expect(h.server.daemonCount('CreateDaemonRuntimeToken'), 1);
        await h.daemons.loadTokens();
        expect(h.daemons.unconfirmed, isFalse);
        expect(h.daemons.tokensOf('rt1'), hasLength(1));
      },
    );

    test('a mint started after leaving another screen is kept', () async {
      final h = await start(
        seed: (s) => s.seedRuntime('alpha', runtime('rt1')),
      );
      h.daemons.leaveSecretScreen('rt1', notify: false);
      expect(await h.daemons.mint('rt1', 'w'), isTrue);
      expect(h.daemons.minted?.secret, 'worker-secret-1');
    });
  });
}
