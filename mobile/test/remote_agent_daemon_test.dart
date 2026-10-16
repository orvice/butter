import 'dart:async';
import 'dart:io';

import 'package:butter_mobile/app/client_context.dart';
import 'package:butter_mobile/core/config_resource_controller.dart';
import 'package:butter_mobile/gen/agents/v1/agent.pb.dart';
import 'package:butter_mobile/gen/agents/v1/dashboard.pb.dart';
import 'package:flutter_test/flutter_test.dart';

import 'support/daemon_server.dart';
import 'support/fake_oauth_platform.dart';
import 'support/memory_storage.dart';

RemoteAgent daemonAgent(
  String id,
  String runtimeId, {
  String executor = 'codex',
}) => RemoteAgent(
  id: id,
  name: 'Agent $id',
  protocol: RemoteAgentProtocol.REMOTE_AGENT_PROTOCOL_DAEMON,
  daemonRuntimeId: runtimeId,
  acpRuntime: executor,
);

Future<(ClientContext, DaemonServerFake)> start({
  void Function(DaemonServerFake)? seed,
}) async {
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
  return (client, server);
}

void main() {
  test('runtimes of this Workspace only feed the picker', () async {
    final (client, server) = await start(
      seed: (s) {
        s.seedRuntime('alpha', DaemonRuntime(id: 'rt1', name: 'One'));
        s.seedRuntime('beta', DaemonRuntime(id: 'rtb', name: 'Other'));
      },
    );
    final remote = client.remoteAgents!;
    await remote.loadRuntimes();
    expect(remote.runtimeState, ConfigLoad.ready);
    expect(remote.runtimes!.map((r) => r.id), ['rt1']);
    expect(server.daemonCalls.map((c) => c.workspace).toSet(), {'alpha'});
  });

  test('a runtime that is not in the Workspace is a reference error', () async {
    final (client, _) = await start(
      seed: (s) => s.seedRuntime('alpha', DaemonRuntime(id: 'rt1')),
    );
    final remote = client.remoteAgents!;
    // Before the runtimes are known nothing is claimed.
    expect(remote.referenceProblem(daemonAgent('a', 'ghost')), isNull);
    await remote.loadRuntimes();
    expect(remote.referenceProblem(daemonAgent('a', 'rt1')), isNull);
    expect(
      remote.referenceProblem(daemonAgent('a', 'ghost')),
      contains('does not exist in this Workspace'),
    );
    expect(
      remote.referenceProblem(daemonAgent('a', '')),
      contains('No runtime is chosen'),
    );
    // Other protocols have no runtime reference.
    expect(
      remote.referenceProblem(
        RemoteAgent(
          id: 'x',
          protocol: RemoteAgentProtocol.REMOTE_AGENT_PROTOCOL_A2A,
        ),
      ),
      isNull,
    );
  });

  test('a connected worker that lacks the executor is reported', () async {
    final (client, _) = await start(
      seed: (s) {
        s.seedRuntime('alpha', DaemonRuntime(id: 'rt1'));
        s.onlineWorkers['alpha']!.add(
          DaemonStatus(
            daemonRuntimeId: 'rt1',
            state: DaemonStatus_State.STATE_ONLINE,
            acpRuntimes: ['opencode'],
          ),
        );
      },
    );
    final remote = client.remoteAgents!;
    await remote.loadRuntimes();
    expect(
      remote.referenceProblem(daemonAgent('a', 'rt1', executor: 'opencode')),
      isNull,
    );
    expect(
      remote.referenceProblem(daemonAgent('a', 'rt1')),
      contains('does not offer executor "codex"'),
    );
  });

  test('a failed runtimes read says so and claims nothing', () async {
    final (client, server) = await start();
    server.daemonRefusals['ListDaemonRuntimes'] = (503, 'unavailable', 'down');
    final remote = client.remoteAgents!;
    await remote.loadRuntimes();
    expect(remote.runtimeState, ConfigLoad.failed);
    expect(remote.runtimeError, isNotNull);
    expect(remote.referenceProblem(daemonAgent('a', 'ghost')), isNull);
  });

  test('suspend drops the runtimes and recover reads them again', () async {
    final (client, server) = await start(
      seed: (s) => s.seedRuntime('alpha', DaemonRuntime(id: 'rt1')),
    );
    final remote = client.remoteAgents!;
    await remote.loadRuntimes();
    remote.suspend();
    expect(remote.runtimes, isNull);
    await remote.recover();
    expect(remote.runtimes!.map((r) => r.id), ['rt1']);
    expect(server.daemonCount('ListDaemonRuntimes'), 2);
  });

  group('a runtimes read that is superseded asks nothing more', () {
    test(
      'after the app leaves, the delayed reply does not read workers',
      () async {
        final (client, server) = await start(
          seed: (s) => s.seedRuntime('alpha', DaemonRuntime(id: 'rt1')),
        );
        final gate = Completer<void>();
        server.daemonBefore = (m) async {
          if (m == 'ListDaemonRuntimes') await gate.future;
        };
        final remote = client.remoteAgents!;
        final pending = remote.loadRuntimes();
        await Future<void>.delayed(const Duration(milliseconds: 100));
        remote.suspend();
        gate.complete();
        await pending;
        await Future<void>.delayed(const Duration(milliseconds: 100));
        expect(server.daemonCount('ListDaemons'), 0);
        expect(remote.runtimes, isNull);
      },
    );

    test('after the context is disposed, no worker read follows', () async {
      final (client, server) = await start(
        seed: (s) => s.seedRuntime('alpha', DaemonRuntime(id: 'rt1')),
      );
      final gate = Completer<void>();
      server.daemonBefore = (m) async {
        if (m == 'ListDaemonRuntimes') await gate.future;
      };
      final remote = client.remoteAgents!;
      final pending = remote.loadRuntimes();
      await Future<void>.delayed(const Duration(milliseconds: 100));
      await client.selectWorkspace('beta');
      expect(remote.disposed, isTrue);
      gate.complete();
      await pending;
      await Future<void>.delayed(const Duration(milliseconds: 100));
      expect(server.daemonCount('ListDaemons'), 0);
    });

    test(
      'a newer read that began meanwhile is the only one to go on',
      () async {
        final (client, server) = await start(
          seed: (s) => s.seedRuntime('alpha', DaemonRuntime(id: 'rt1')),
        );
        final gate = Completer<void>();
        var held = false;
        server.daemonBefore = (m) async {
          if (m == 'ListDaemonRuntimes' && !held) {
            held = true;
            await gate.future;
          }
        };
        final remote = client.remoteAgents!;
        final first = remote.loadRuntimes();
        await Future<void>.delayed(const Duration(milliseconds: 100));
        await remote.loadRuntimes();
        expect(server.daemonCount('ListDaemons'), 1);
        gate.complete();
        await first;
        await Future<void>.delayed(const Duration(milliseconds: 100));
        expect(server.daemonCount('ListDaemons'), 1);
        expect(remote.runtimeState, ConfigLoad.ready);
      },
    );
  });
}
