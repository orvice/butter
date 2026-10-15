import 'dart:async';
import 'dart:io';

import 'package:butter_mobile/app/client_context.dart';
import 'package:butter_mobile/core/butter_api.dart';
import 'package:butter_mobile/features/mcp/mcp_controller.dart';
import 'package:butter_mobile/gen/agents/v1/agent.pb.dart';
import 'package:flutter_test/flutter_test.dart';

import 'support/fake_oauth_platform.dart';
import 'support/manual_clock.dart';
import 'support/mcp_server.dart';
import 'support/memory_storage.dart';

/// How a flow being waited on is observed in the foreground: bounded read-only
/// polling and a deadline, on a clock and timers the test moves by hand.
void main() {
  const status = 'GetMCPServerOAuthStatus';
  const poll = Duration(seconds: 5);
  final t0 = DateTime.utc(2026, 10, 16, 12);

  MCPServer oauthServer(String id) {
    final s = MCPServer(
      id: id,
      name: 'Server $id',
      url: 'https://mcp.example.com/mcp',
      transport: MCPServerTransport.MCP_SERVER_TRANSPORT_STREAMABLE_HTTP,
    );
    s.ensureAuth()
      ..type = MCPServerAuthType.MCP_SERVER_AUTH_TYPE_OAUTH2
      ..ensureOauth2();
    return s;
  }

  late McpServerFake server;
  late ManualClock clock;
  late FakeOAuthPlatform platform;
  late McpController c;

  /// A controller of its own, on the manual clock, talking to a fake server
  /// that holds one OAuth server `a1`.
  Future<void> rig({
    Duration pollEvery = poll,
    void Function(McpServerFake)? seed,
  }) async {
    final old = HttpOverrides.current;
    HttpOverrides.global = null;
    server = McpServerFake()..seed('alpha', oauthServer('a1'));
    seed?.call(server);
    await server.start();
    platform = FakeOAuthPlatform();
    final client = ClientContext(
      storage: MemoryStorage(),
      allowHttp: true,
      oauth: platform,
    );
    await client.setServer(server.url);
    await client.login('alice', 'password');
    await client.selectWorkspace('alpha');
    final login = await ButterApi(client.address!).login('alice', 'password');
    clock = ManualClock(t0);
    c = McpController(
      newApi: () => ButterApi(client.address!),
      token: login.token,
      workspaceId: 'alpha',
      onUnauthenticated: () {},
      openBrowser: platform.openBrowser,
      allowHttp: true,
      now: () => clock.now,
      pollInterval: pollEvery,
      timer: clock.schedule,
    );
    addTearDown(() async {
      c.dispose();
      server.before = null;
      client.dispose();
      await server.close();
      HttpOverrides.global = old;
    });
  }

  Future<void> until(bool Function() done, {String? reason}) async {
    for (var i = 0; i < 300 && !done(); i++) {
      await Future<void>.delayed(const Duration(milliseconds: 10));
    }
    expect(done(), isTrue, reason: reason ?? 'condition never held');
  }

  /// The look that a tick started has finished and the next one is scheduled
  /// (or the watch ended).
  Future<void> settled() async {
    await until(() => !c.oauthOf('a1').reading);
    await Future<void>.delayed(const Duration(milliseconds: 5));
  }

  /// Holds every status read until the returned gate opens.
  Completer<void> holdStatus() {
    final gate = Completer<void>();
    server.before = (m) => m == status ? gate.future : Future.value();
    return gate;
  }

  Future<void> waiting() async {
    await c.connectOAuth('a1');
    expect(c.oauthOf('a1').phase, McpOAuthPhase.waiting);
    expect(platform.opened, hasLength(1));
  }

  group('a flow being waited on', () {
    test('is noticed finishing without anyone asking', () async {
      await rig();
      await waiting();
      // A poll timer and the deadline.
      expect(clock.pending, 2);
      expect(server.count(status), 1, reason: 'the baseline read only');

      server.connect(
        'alpha',
        'a1',
        at: clock.now.add(const Duration(minutes: 1)),
      );
      await clock.advance(poll);
      await until(() => c.oauthOf('a1').phase == McpOAuthPhase.connected);

      expect(c.oauthOf('a1').message, 'Connected.');
      expect(
        clock.pending,
        0,
        reason: 'nothing keeps watching a finished flow',
      );
      expect(server.count(status), 2);
      await clock.advance(const Duration(hours: 1));
      expect(server.count(status), 2);
    });

    test('is looked at once per interval, with reads only', () async {
      await rig();
      await waiting();
      for (var i = 1; i <= 4; i++) {
        await clock.advance(poll);
        await settled();
        expect(server.count(status), 1 + i, reason: 'one look per interval');
        expect(c.oauthOf('a1').phase, McpOAuthPhase.waiting);
        expect(c.oauthOf('a1').message, contains('Not connected yet'));
      }
      // Nothing was ever sent but the one Start the person asked for.
      expect(server.count('StartMCPServerOAuth'), 1);
      expect(server.writes, 1);
      expect(platform.opened, hasLength(1));
    });

    test('a look that is slow is not stacked upon', () async {
      await rig();
      await waiting();
      final gate = holdStatus();
      await clock.advance(poll);
      await until(() => server.count(status) == 2);
      // Time passes while that read hangs: no second one starts.
      await clock.advance(poll * 5);
      expect(server.count(status), 2);
      expect(c.oauthOf('a1').phase, McpOAuthPhase.waiting);

      server.before = null;
      gate.complete();
      await settled();
      await clock.advance(poll);
      await settled();
      expect(server.count(status), 3, reason: 'polling resumes after it');
    });

    test('expires at its deadline, says so, and stops looking', () async {
      await rig();
      await waiting();
      await clock.advance(mcpOAuthFlowTtl - const Duration(seconds: 1));
      await settled();
      expect(c.oauthOf('a1').phase, McpOAuthPhase.waiting);

      await clock.advance(const Duration(seconds: 2));
      expect(c.oauthOf('a1').phase, McpOAuthPhase.timedOut);
      expect(c.oauthOf('a1').message, contains('expired'));
      await settled();
      expect(clock.pending, 0);
      final reads = server.count(status);

      await clock.advance(const Duration(hours: 2));
      expect(
        server.count(status),
        reads,
        reason: 'nothing looks after the end',
      );
      expect(c.hasWaitingFlow, isFalse);
      expect(server.count('StartMCPServerOAuth'), 1, reason: 'never replayed');
      expect(server.writes, 1);
    });

    test('the number of looks is bounded by the flow lifetime', () async {
      await rig();
      await waiting();
      for (
        var i = 0;
        i < mcpOAuthFlowTtl.inSeconds ~/ poll.inSeconds + 5;
        i++
      ) {
        await clock.advance(poll);
        await settled();
      }
      // The baseline, at most one look per interval, and the final one.
      final bound = 1 + mcpOAuthFlowTtl.inSeconds ~/ poll.inSeconds + 1;
      expect(server.count(status), lessThanOrEqualTo(bound));
      expect(c.oauthOf('a1').phase, McpOAuthPhase.timedOut);
    });

    test('ends at its deadline even while a status read hangs', () async {
      await rig();
      await waiting();
      final gate = holdStatus();
      await clock.advance(poll);
      await until(() => server.count(status) == 2);
      expect(c.oauthOf('a1').reading, isTrue);

      // The read never answers; the deadline still ends the waiting.
      await clock.advance(mcpOAuthFlowTtl);
      expect(c.oauthOf('a1').phase, McpOAuthPhase.timedOut);
      expect(c.hasWaitingFlow, isFalse);

      // An answer that arrives afterwards cannot bring it back.
      server.before = null;
      gate.complete();
      await Future<void>.delayed(const Duration(milliseconds: 100));
      expect(c.oauthOf('a1').phase, McpOAuthPhase.timedOut);
      expect(clock.pending, 0);
    });

    test('ends at its deadline when every status read fails', () async {
      await rig();
      await waiting();
      server.refusals[status] = (503, 'unavailable', 'down');
      await clock.advance(poll);
      await settled();
      expect(c.oauthOf('a1').phase, McpOAuthPhase.waiting);
      expect(
        c.oauthOf('a1').message,
        contains('Could not read the authorization state'),
      );
      await clock.advance(mcpOAuthFlowTtl);
      await settled();
      expect(c.oauthOf('a1').phase, McpOAuthPhase.timedOut);
      expect(
        c.oauthOf('a1').message,
        contains('expired'),
        reason: 'a failed last look does not replace the end of the flow',
      );
      expect(clock.pending, 0);
    });

    test('a connection finished in the last moment is reported as it is, '
        'against the baseline', () async {
      // Looks are an hour apart, so the deadline comes first.
      await rig(pollEvery: const Duration(hours: 1));
      await waiting();
      server.connect(
        'alpha',
        'a1',
        at: clock.now.add(const Duration(minutes: 9)),
      );
      await clock.advance(mcpOAuthFlowTtl);
      await until(() => c.oauthOf('a1').phase == McpOAuthPhase.connected);
      expect(
        c.oauthOf('a1').message,
        contains('after the app stopped waiting'),
      );
    });

    test(
      'an earlier connection is not taken for the end of this flow',
      () async {
        await rig(
          seed: (s) => s.connect(
            'alpha',
            'a1',
            at: t0.subtract(const Duration(days: 1)),
          ),
        );
        await waiting();
        for (var i = 0; i < 3; i++) {
          await clock.advance(poll);
          await settled();
        }
        expect(c.oauthOf('a1').phase, McpOAuthPhase.waiting);
        await clock.advance(mcpOAuthFlowTtl);
        await settled();
        expect(c.oauthOf('a1').phase, McpOAuthPhase.timedOut);
      },
    );

    test('while the browser launcher is still pending, the watch already '
        'runs', () async {
      await rig();
      platform.openGate = Completer<bool>();
      final start = c.connectOAuth('a1');
      await until(() => platform.opened.length == 1);
      expect(c.oauthOf('a1').launching, isTrue);
      expect(clock.pending, 2);

      server.connect(
        'alpha',
        'a1',
        at: clock.now.add(const Duration(minutes: 1)),
      );
      await clock.advance(poll);
      await until(() => c.oauthOf('a1').phase == McpOAuthPhase.connected);
      platform.openGate!.complete(true);
      await start;
      expect(c.oauthOf('a1').phase, McpOAuthPhase.connected);
    });
  });

  group('the watch is invalidated with the flow', () {
    test(
      'stopping waiting stops the looks; a later finish still shows',
      () async {
        await rig();
        await waiting();
        c.cancelOAuth('a1');
        expect(clock.pending, 0);
        final reads = server.count(status);
        await clock.advance(mcpOAuthFlowTtl * 2);
        expect(server.count(status), reads);
        expect(c.oauthOf('a1').phase, McpOAuthPhase.cancelled);

        server.connect('alpha', 'a1', at: clock.now);
        await c.loadOAuth('a1');
        expect(c.oauthOf('a1').phase, McpOAuthPhase.connected);
        expect(
          c.oauthOf('a1').message,
          contains('after the app stopped waiting'),
        );
      },
    );

    test(
      'a look still on its way when the person stops waiting is dropped',
      () async {
        await rig();
        await waiting();
        final gate = holdStatus();
        await clock.advance(poll);
        await until(() => server.count(status) == 2);
        c.cancelOAuth('a1');
        server.before = null;
        // The answer was taken before this connection existed.
        server.connect(
          'alpha',
          'a1',
          at: clock.now.add(const Duration(minutes: 1)),
        );
        gate.complete();
        await Future<void>.delayed(const Duration(milliseconds: 100));
        expect(c.oauthOf('a1').phase, McpOAuthPhase.cancelled);
        expect(clock.pending, 0);
      },
    );

    test('disconnecting ends the watch', () async {
      await rig();
      await waiting();
      expect(await c.disconnectOAuth('a1'), isTrue);
      expect(clock.pending, 0);
      final reads = server.count(status);
      await clock.advance(mcpOAuthFlowTtl * 2);
      expect(server.count(status), reads);
      expect(c.oauthOf('a1').phase, McpOAuthPhase.idle);
    });

    test('deleting the server ends the watch', () async {
      await rig();
      await c.load();
      await waiting();
      expect(await c.delete('a1'), isTrue);
      expect(clock.pending, 0);
      final reads = server.count(status);
      await clock.advance(mcpOAuthFlowTtl * 2);
      expect(server.count(status), reads);
    });

    test('a flow that failed to open the browser is not watched', () async {
      await rig();
      platform.openResult = false;
      await c.connectOAuth('a1');
      expect(c.oauthOf('a1').phase, McpOAuthPhase.failed);
      expect(clock.pending, 0);
    });

    test('a new flow replaces the old watch instead of doubling it', () async {
      await rig();
      await waiting();
      c.cancelOAuth('a1');
      await c.connectOAuth('a1');
      expect(c.oauthOf('a1').phase, McpOAuthPhase.waiting);
      expect(clock.pending, 2);
      final reads = server.count(status);
      await clock.advance(poll);
      await settled();
      expect(server.count(status), reads + 1, reason: 'one chain of looks');
    });
  });

  group('the watch follows the app', () {
    test('suspending stops it; returning looks at once and resumes', () async {
      await rig();
      await waiting();
      c.suspend();
      expect(clock.pending, 0);
      final reads = server.count(status);
      await clock.advance(poll * 10);
      expect(
        server.count(status),
        reads,
        reason: 'nothing runs in the background',
      );
      expect(c.oauthOf('a1').phase, McpOAuthPhase.waiting);

      await c.recover();
      expect(server.count(status), reads + 1, reason: 'a look on return');
      expect(clock.pending, 2, reason: 'watching again');
      server.connect('alpha', 'a1', at: clock.now);
      await clock.advance(poll);
      await until(() => c.oauthOf('a1').phase == McpOAuthPhase.connected);
      expect(clock.pending, 0);
    });

    test('a finish while away is found on return', () async {
      await rig();
      await waiting();
      c.suspend();
      server.connect(
        'alpha',
        'a1',
        at: clock.now.add(const Duration(minutes: 2)),
      );
      await c.recover();
      expect(c.oauthOf('a1').phase, McpOAuthPhase.connected);
      expect(clock.pending, 0);
    });

    test('returning after the deadline ends the waiting even when the read '
        'hangs', () async {
      await rig();
      await waiting();
      c.suspend();
      await clock.advance(mcpOAuthFlowTtl + const Duration(minutes: 1));
      final gate = holdStatus();
      await c.recover();
      expect(c.oauthOf('a1').phase, McpOAuthPhase.timedOut);
      expect(c.hasWaitingFlow, isFalse);
      expect(clock.pending, 0);
      server.before = null;
      gate.complete();
    });

    test(
      'disposing cancels every timer and nothing is requested after',
      () async {
        await rig();
        await waiting();
        c.dispose();
        expect(clock.pending, 0);
        final reads = server.count(status);
        await clock.advance(mcpOAuthFlowTtl * 2);
        expect(server.count(status), reads);
      },
    );

    test('a look that is answered after disposal changes nothing', () async {
      await rig();
      await waiting();
      final gate = holdStatus();
      await clock.advance(poll);
      await until(() => server.count(status) == 2);
      c.dispose();
      server.before = null;
      gate.complete();
      await Future<void>.delayed(const Duration(milliseconds: 50));
      expect(clock.pending, 0);
    });
  });
}
