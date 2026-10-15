import 'dart:async';
import 'dart:io';

import 'package:butter_mobile/app/client_context.dart';
import 'package:butter_mobile/core/config_resource_controller.dart';
import 'package:butter_mobile/core/butter_api.dart';
import 'package:butter_mobile/features/mcp/mcp_controller.dart';
import 'package:butter_mobile/features/mcp/mcp_draft.dart';
import 'package:butter_mobile/gen/agents/v1/agent.pb.dart';
import 'package:butter_mobile/gen/agents/v1/agent_service.pb.dart';
import 'package:butter_mobile/gen/agents/v1/dashboard.pb.dart';
import 'package:flutter_test/flutter_test.dart';

import 'support/fake_oauth_platform.dart';
import 'support/mcp_server.dart';
import 'support/memory_storage.dart';

MCPServer server(
  String id, {
  String name = '',
  String url = 'https://mcp.example.com/mcp',
}) => MCPServer(
  id: id,
  name: name.isEmpty ? 'Server $id' : name,
  url: url,
  transport: MCPServerTransport.MCP_SERVER_TRANSPORT_STREAMABLE_HTTP,
);

MCPServer oauthServer(String id) {
  final s = server(id);
  s.ensureAuth()
    ..type = MCPServerAuthType.MCP_SERVER_AUTH_TYPE_OAUTH2
    ..ensureOauth2();
  return s;
}

class Harness {
  Harness(this.client, this.server, this.platform);
  final ClientContext client;
  final McpServerFake server;
  final FakeOAuthPlatform platform;
  McpController get mcp => client.mcp!;
}

Future<Harness> start({
  void Function(McpServerFake)? seed,
  DateTime Function()? now,
}) async {
  final old = HttpOverrides.current;
  HttpOverrides.global = null;
  final server = McpServerFake();
  seed?.call(server);
  await server.start();
  final platform = FakeOAuthPlatform();
  final client = ClientContext(
    storage: MemoryStorage(),
    allowHttp: true,
    oauth: platform,
  );
  addTearDown(() async {
    client.dispose();
    await server.close();
    HttpOverrides.global = old;
  });
  await client.setServer(server.url);
  await client.login('alice', 'password');
  await client.selectWorkspace('alpha');
  return Harness(client, server, platform);
}

void main() {
  test('lists and reads the active Workspace only', () async {
    final h = await start(
      seed: (s) {
        s.seed('alpha', server('a1'));
        s.seed('beta', server('b1'));
      },
    );
    await h.mcp.load();
    expect(h.mcp.listState, ConfigLoad.ready);
    expect(h.mcp.items.map((s) => s.id), ['a1']);
    await h.mcp.loadItem('a1');
    expect(h.mcp.item?.name, 'Server a1');
    expect(h.server.mcpCalls.map((c) => c.workspace).toSet(), {'alpha'});
  });

  test('create stores the whole message, a duplicate ID is refused', () async {
    final h = await start();
    final draft = McpDraft.create()
      ..id = 'docs'
      ..name = 'Docs'
      ..url = 'https://mcp.example.com/mcp'
      ..authKind = McpAuthKind.staticHeaders;
    draft.headers.add(McpHeaderRow.added(key: 'X-Key', replacement: 'v1'));
    expect(await h.mcp.create(draft.build()), isTrue);
    final stored = h.server.servers['alpha']!['docs']!;
    expect(stored.headers, {'X-Key': 'v1'});
    expect(h.mcp.items.single.id, 'docs');

    expect(await h.mcp.create(draft.build()), isFalse);
    expect(h.mcp.writeError, contains('already exists'));
    // A definitive refusal is not an unknown outcome.
    expect(h.mcp.unconfirmed, isFalse);
  });

  test('an edit carries every field the form does not touch', () async {
    final h = await start(
      seed: (s) {
        final stored = server('a1')..metadata['owner'] = 'platform';
        stored.toolFilter.add('echo');
        s.seed('alpha', stored);
      },
    );
    await h.mcp.loadItem('a1');
    final draft = McpDraft.edit(h.mcp.item!)..name = 'Renamed';
    expect(await h.mcp.update(draft.build()), isTrue);
    final stored = h.server.servers['alpha']!['a1']!;
    expect(stored.name, 'Renamed');
    expect(stored.metadata['owner'], 'platform');
    expect(stored.toolFilter, ['echo']);
    expect(stored.workspaceId, 'alpha');
  });

  test(
    'refused for lack of access: a clear message, nothing unknown',
    () async {
      final h = await start(seed: (s) => s.seed('alpha', server('a1')));
      await h.mcp.load();
      h.server.refusals['DeleteMCPServer'] = (
        403,
        'permission_denied',
        'role not allowed: token abc123',
      );
      expect(await h.mcp.delete('a1'), isFalse);
      expect(h.mcp.writeError, contains('do not have access'));
      expect(h.mcp.writeError, isNot(contains('abc123')));
      expect(h.mcp.unconfirmed, isFalse);
      expect(h.mcp.items, hasLength(1));
      expect(h.server.count('DeleteMCPServer'), 1);
    },
  );

  group('a write with an unknown outcome', () {
    test(
      'is never repeated and blocks further writes until a read settles it',
      () async {
        final h = await start();
        h.server.loseReply = true;
        final draft = McpDraft.create()
          ..id = 'docs'
          ..name = 'Docs'
          ..url = 'https://mcp.example.com/mcp';
        expect(await h.mcp.create(draft.build()), isFalse);
        expect(h.mcp.unconfirmed, isTrue);
        expect(h.mcp.writeNotice, contains('not confirmed'));
        // It was in fact applied.
        expect(h.server.servers['alpha'], contains('docs'));

        h.server.loseReply = false;
        expect(await h.mcp.create(draft.build()), isFalse);
        expect(await h.mcp.delete('docs'), isFalse);
        expect(h.server.count('CreateMCPServer'), 1);
        expect(h.server.count('DeleteMCPServer'), 0);

        await h.mcp.load();
        expect(h.mcp.unconfirmed, isFalse);
        expect(h.mcp.items.map((s) => s.id), ['docs']);
      },
    );

    test('a read of the object settles it, including its absence', () async {
      final h = await start(seed: (s) => s.seed('alpha', server('a1')));
      await h.mcp.load();
      h.server.loseReply = true;
      expect(await h.mcp.delete('a1'), isFalse);
      expect(h.mcp.unconfirmed, isTrue);
      h.server.loseReply = false;
      await h.mcp.loadItem('a1');
      expect(h.mcp.unconfirmed, isFalse);
      expect(h.mcp.items, isEmpty);
    });

    test('a read of another server does not settle it', () async {
      final h = await start(
        seed: (s) {
          s.seed('alpha', server('a1'));
          s.seed('alpha', server('a2'));
        },
      );
      await h.mcp.load();
      h.server.loseReply = true;
      await h.mcp.delete('a1');
      h.server.loseReply = false;
      await h.mcp.loadItem('a2');
      expect(h.mcp.unconfirmed, isTrue);
    });

    test('a read that began before the write ended cannot settle it', () async {
      final h = await start(seed: (s) => s.seed('alpha', server('a1')));
      final gate = Completer<void>();
      h.server.before = (m) =>
          m == 'ListMCPServers' ? gate.future : Future.value();
      final stale = h.mcp.load();
      await Future<void>.delayed(const Duration(milliseconds: 50));
      h.server.before = null;
      h.server.loseReply = true;
      await h.mcp.delete('a1');
      expect(h.mcp.unconfirmed, isTrue);
      gate.complete();
      await stale;
      expect(h.mcp.unconfirmed, isTrue);
      expect(h.mcp.items, isEmpty);
    });

    test('suspending mid-write marks exactly it unknown', () async {
      final h = await start();
      final gate = Completer<void>();
      h.server.before = (m) =>
          m == 'CreateMCPServer' ? gate.future : Future.value();
      final draft = McpDraft.create()
        ..id = 'docs'
        ..name = 'Docs'
        ..url = 'https://mcp.example.com/mcp';
      final write = h.mcp.create(draft.build());
      await Future<void>.delayed(const Duration(milliseconds: 50));
      h.client.suspend();
      gate.complete();
      expect(await write, isFalse);
      expect(h.mcp.writing, isFalse);
      expect(h.mcp.unconfirmed, isTrue);
      expect(h.server.count('CreateMCPServer'), 1);
    });
  });

  test('a read answered after a write never undoes it', () async {
    final h = await start(seed: (s) => s.seed('alpha', server('a1')));
    final gate = Completer<void>();
    h.server.before = (m) =>
        m == 'ListMCPServers' ? gate.future : Future.value();
    final stale = h.mcp.load();
    await Future<void>.delayed(const Duration(milliseconds: 50));
    h.server.before = null;
    final draft = McpDraft.create()
      ..id = 'a2'
      ..name = 'Two'
      ..url = 'https://two.example/mcp';
    expect(await h.mcp.create(draft.build()), isTrue);
    gate.complete();
    await stale;
    expect(h.mcp.items.map((s) => s.id), contains('a2'));
  });

  test(
    'stored credentials are dropped on suspend and read again on recover',
    () async {
      final h = await start(
        seed: (s) => s.seed(
          'alpha',
          server('a1')
            ..headers['Authorization'] = 'Bearer secret-token'
            ..ensureAuth().type =
                MCPServerAuthType.MCP_SERVER_AUTH_TYPE_STATIC_HEADERS,
        ),
      );
      await h.mcp.load();
      await h.mcp.loadItem('a1');
      expect(h.mcp.item!.headers['Authorization'], 'Bearer secret-token');
      h.client.suspend();
      expect(h.mcp.items, isEmpty);
      expect(h.mcp.item, isNull);
      await h.mcp.recover();
      expect(
        h.mcp.items.single.headers['Authorization'],
        'Bearer secret-token',
      );
      expect(h.mcp.item?.id, 'a1');
    },
  );

  group('connection test and tools', () {
    test('classify the probe and never show the server text', () async {
      final h = await start(seed: (s) => s.seed('alpha', server('a1')));
      h.server.statusState = MCPServerStatus_State.STATE_DISCONNECTED;
      h.server.statusDetail =
          'dial tcp 10.0.0.5:443: connection refused (Authorization: Bearer leak-123)';
      await h.mcp.probe('a1');
      final view = h.mcp.statuses['a1']!;
      expect(view.kind, McpStatusKind.disconnected);
      expect(view.reason, contains('could not connect'));
      expect(view.reason, isNot(contains('leak-123')));
      expect(view.reason, isNot(contains('10.0.0.5')));

      h.server.statusState = MCPServerStatus_State.STATE_ERROR;
      h.server.statusDetail = 'sending initialize: Unauthorized';
      await h.mcp.probe('a1');
      expect(
        h.mcp.statuses['a1']!.reason,
        contains('rejected the credentials'),
      );

      h.server.statusDetail = 'mcp oauth connection requires authorization';
      await h.mcp.probe('a1');
      expect(h.mcp.statuses['a1']!.reason, contains('Connect the account'));

      h.server.statusState = MCPServerStatus_State.STATE_CONNECTED;
      await h.mcp.probe('a1');
      expect(h.mcp.statuses['a1']!.kind, McpStatusKind.connected);
    });

    test(
      'a failed call is a failure, distinct from an unreachable server',
      () async {
        final h = await start(seed: (s) => s.seed('alpha', server('a1')));
        h.server.refusals['GetMCPServerStatus'] = (503, 'unavailable', 'x');
        await h.mcp.probe('a1');
        expect(h.mcp.statuses['a1']!.kind, McpStatusKind.failed);
        expect(h.server.count('GetMCPServerStatus'), 1);
      },
    );

    test(
      'tools show which the filter allows; a probe failure is per server',
      () async {
        final h = await start(seed: (s) => s.seed('alpha', server('a1')));
        h.server.toolList = [
          MCPTool(name: 'echo', serverId: 'a1', allowed: true),
          MCPTool(name: 'danger', serverId: 'a1', allowed: false),
        ];
        await h.mcp.loadTools('a1');
        final view = h.mcp.tools['a1']!;
        expect(view.state, ConfigLoad.ready);
        expect(view.tools.map((t) => '${t.name}:${t.allowed}'), [
          'echo:true',
          'danger:false',
        ]);

        h.server.toolErrors = {'a1': 'list tools: Unauthorized'};
        await h.mcp.loadTools('a1');
        expect(h.mcp.tools['a1']!.state, ConfigLoad.failed);
        expect(h.mcp.tools['a1']!.error, contains('rejected the credentials'));
      },
    );

    test('a write to the server drops an answer still on its way', () async {
      final h = await start(seed: (s) => s.seed('alpha', server('a1')));
      final gate = Completer<void>();
      h.server.before = (m) =>
          m == 'GetMCPServerStatus' ? gate.future : Future.value();
      final probe = h.mcp.probe('a1');
      await Future<void>.delayed(const Duration(milliseconds: 50));
      h.server.before = null;
      await h.mcp.loadItem('a1');
      final draft = McpDraft.edit(h.mcp.item!)
        ..url = 'https://other.example/mcp';
      await h.mcp.update(draft.build());
      gate.complete();
      await probe;
      expect(h.mcp.statuses['a1'], isNull);
    });

    test('suspending fences a probe in flight', () async {
      final h = await start(seed: (s) => s.seed('alpha', server('a1')));
      final gate = Completer<void>();
      h.server.before = (m) =>
          m == 'GetMCPServerStatus' ? gate.future : Future.value();
      final probe = h.mcp.probe('a1');
      await Future<void>.delayed(const Duration(milliseconds: 50));
      h.client.suspend();
      gate.complete();
      await probe;
      expect(h.mcp.statuses['a1'], isNull);
    });
  });

  group('global presets', () {
    MCPServer preset(String id) => server(id)..headers['Authorization'] = '***';

    test('are listed read-only and installed with one call', () async {
      final h = await start(seed: (s) => s.presets.add(preset('p1')));
      await h.mcp.loadPresets();
      expect(h.mcp.presets.single.id, 'p1');
      expect(await h.mcp.install('p1'), isTrue);
      expect(h.server.servers['alpha'], contains('p1'));
      expect(h.mcp.items.single.metadata[mcpPresetMetadataKey], 'p1');
      expect(
        h.server.mcpCalls.map((c) => c.method).toSet().intersection({
          'CreateGlobalMCPServer',
          'UpdateGlobalMCPServer',
          'DeleteGlobalMCPServer',
        }),
        isEmpty,
      );
    });

    test('installing over a server that is not this preset says so', () async {
      final h = await start(
        seed: (s) {
          s.presets.add(preset('p1'));
          s.seed('alpha', server('p1'));
        },
      );
      expect(await h.mcp.install('p1'), isFalse);
      expect(h.mcp.writeError, contains('not installed from this preset'));
      expect(h.mcp.unconfirmed, isFalse);
    });

    test(
      'installing a preset that is gone is a NotFound, not a failure',
      () async {
        final h = await start();
        expect(await h.mcp.install('missing'), isFalse);
        expect(h.mcp.writeError, contains('no longer offered'));
      },
    );

    test('a lost install reply is unconfirmed and never repeated', () async {
      final h = await start(seed: (s) => s.presets.add(preset('p1')));
      h.server.loseReply = true;
      expect(await h.mcp.install('p1'), isFalse);
      expect(h.mcp.unconfirmed, isTrue);
      h.server.loseReply = false;
      expect(await h.mcp.install('p1'), isFalse);
      expect(h.server.count('InstallGlobalMCPServer'), 1);
      await h.mcp.loadItem('p1');
      expect(h.mcp.unconfirmed, isFalse);
    });
  });

  group('OAuth in the system browser', () {
    final t0 = DateTime.utc(2026, 10, 16, 12);
    var clock = t0;

    // A controller of its own so the test owns the clock; it talks to the
    // same server as the signed-in context.
    Future<(Harness, McpController, FakeOAuthPlatform)> rig({
      bool open = true,
      void Function(McpServerFake)? seed,
    }) async {
      clock = t0;
      final h = await start(
        seed: (s) {
          s.seed('alpha', oauthServer('a1'));
          seed?.call(s);
        },
      );
      final login = await ButterApi(
        h.client.address!,
      ).login('alice', 'password');
      final platform = FakeOAuthPlatform()..openResult = open;
      final c = McpController(
        newApi: () => ButterApi(h.client.address!),
        token: login.token,
        workspaceId: 'alpha',
        onUnauthenticated: () {},
        openBrowser: platform.openBrowser,
        allowHttp: true,
        now: () => clock,
      );
      addTearDown(c.dispose);
      return (h, c, platform);
    }

    test(
      'starts with no return link, opens the system browser and waits',
      () async {
        final (h, c, platform) = await rig();
        await c.connectOAuth('a1');
        final view = c.oauthOf('a1');
        expect(view.phase, McpOAuthPhase.waiting);
        expect(
          platform.opened.single.toString(),
          'https://auth.example.test/authorize?x=1',
        );
        expect(h.server.count('StartMCPServerOAuth'), 1);
        expect(view.deadline, t0.add(mcpOAuthFlowTtl));
        // Reads before starting; nothing but reads and the one start.
        expect(
          h.server.mcpCalls
              .map((call) => call.method)
              .where((m) => m.contains('OAuth')),
          ['GetMCPServerOAuthStatus', 'StartMCPServerOAuth'],
        );
      },
    );

    test(
      'is connected only when the server reports a newer connection',
      () async {
        final (h, c, _) = await rig();
        await c.connectOAuth('a1');
        await c.loadOAuth('a1');
        expect(c.oauthOf('a1').phase, McpOAuthPhase.waiting);
        expect(c.oauthOf('a1').message, contains('Not connected yet'));

        h.server.connect('alpha', 'a1', at: t0.add(const Duration(minutes: 1)));
        await c.loadOAuth('a1');
        expect(c.oauthOf('a1').phase, McpOAuthPhase.connected);
      },
    );

    test('an earlier connection is not mistaken for this flow', () async {
      final (h, c, _) = await rig(
        seed: (s) =>
            s.connect('alpha', 'a1', at: t0.subtract(const Duration(days: 1))),
      );
      await c.connectOAuth('a1');
      await c.loadOAuth('a1');
      expect(c.oauthOf('a1').phase, McpOAuthPhase.waiting);

      // The user authorizes again: a newer connection.
      h.server.connect('alpha', 'a1', at: t0.add(const Duration(minutes: 2)));
      await c.loadOAuth('a1');
      expect(c.oauthOf('a1').phase, McpOAuthPhase.connected);
    });

    test('a refused or failed authorization leaves it unconnected until the '
        'link expires, then says so', () async {
      final (_, c, _) = await rig();
      await c.connectOAuth('a1');
      clock = t0.add(const Duration(minutes: 3));
      await c.loadOAuth('a1');
      expect(c.oauthOf('a1').phase, McpOAuthPhase.waiting);

      clock = t0.add(const Duration(minutes: 11));
      await c.loadOAuth('a1');
      final view = c.oauthOf('a1');
      expect(view.phase, McpOAuthPhase.timedOut);
      expect(view.message, contains('expired'));
      expect(view.message, contains('declined'));
    });

    test(
      'cancel stops waiting locally and says the link stays valid',
      () async {
        final (h, c, _) = await rig();
        await c.connectOAuth('a1');
        c.cancelOAuth('a1');
        expect(c.oauthOf('a1').phase, McpOAuthPhase.cancelled);
        expect(c.oauthOf('a1').message, contains('stays valid'));
        expect(h.server.count('DisconnectMCPServerOAuth'), 0);

        // A connection finished in the browser anyway is still the truth.
        h.server.connect('alpha', 'a1', at: t0.add(const Duration(minutes: 1)));
        await c.loadOAuth('a1');
        expect(
          c.oauthOf('a1').status!.state,
          MCPOAuthConnectionState.MCPO_AUTH_CONNECTION_STATE_CONNECTED,
        );
      },
    );

    test('a browser that will not open leaves a clear failure', () async {
      final (_, c, platform) = await rig(open: false);
      await c.connectOAuth('a1');
      expect(c.oauthOf('a1').phase, McpOAuthPhase.failed);
      expect(c.oauthOf('a1').message, contains('Could not open the browser'));
      expect(platform.opened, hasLength(1));
    });

    test('an address that is not a web page is never opened', () async {
      final (h, c, platform) = await rig();
      h.server.authorizationUrl = 'javascript:alert(1)';
      await c.connectOAuth('a1');
      expect(c.oauthOf('a1').phase, McpOAuthPhase.failed);
      expect(platform.opened, isEmpty);
    });

    test('if the current state cannot be read, nothing is started', () async {
      final (h, c, _) = await rig();
      h.server.refusals['GetMCPServerOAuthStatus'] = (503, 'unavailable', 'x');
      await c.connectOAuth('a1');
      expect(c.oauthOf('a1').phase, McpOAuthPhase.failed);
      expect(c.oauthOf('a1').message, contains('nothing was started'));
      expect(h.server.count('StartMCPServerOAuth'), 0);
    });

    test('the server refusing to start says what to fix', () async {
      final (h, c, platform) = await rig();
      h.server.refusals['StartMCPServerOAuth'] = (
        400,
        'failed_precondition',
        'mcp oauth service is not configured',
      );
      await c.connectOAuth('a1');
      expect(c.oauthOf('a1').phase, McpOAuthPhase.failed);
      expect(c.oauthOf('a1').message, contains('not enabled'));
      expect(platform.opened, isEmpty);
    });

    test('a flow being waited on survives the app going to the background and '
        'is checked when it returns', () async {
      final (h, c, _) = await rig();
      await c.connectOAuth('a1');
      c.suspend();
      expect(c.oauthOf('a1').phase, McpOAuthPhase.waiting);
      h.server.connect('alpha', 'a1', at: t0.add(const Duration(minutes: 1)));
      await c.recover();
      expect(c.oauthOf('a1').phase, McpOAuthPhase.connected);
    });

    test('a start still in flight when the app is suspended is unconfirmed and '
        'opens nothing', () async {
      final (h, c, platform) = await rig();
      final gate = Completer<void>();
      h.server.before = (m) =>
          m == 'StartMCPServerOAuth' ? gate.future : Future.value();
      final start = c.connectOAuth('a1');
      await Future<void>.delayed(const Duration(milliseconds: 100));
      c.suspend();
      gate.complete();
      await start;
      expect(c.oauthOf('a1').phase, McpOAuthPhase.failed);
      expect(c.oauthOf('a1').message, contains('Start again'));
      expect(platform.opened, isEmpty);
      expect(h.server.count('StartMCPServerOAuth'), 1);
    });

    test(
      'disconnect is one write; a lost reply is settled by reading the state',
      () async {
        final (h, c, _) = await rig(
          seed: (s) => s.connect('alpha', 'a1', at: t0),
        );
        await c.loadOAuth('a1');
        h.server.loseReply = true;
        expect(await c.disconnectOAuth('a1'), isFalse);
        expect(c.unconfirmed, isTrue);
        h.server.loseReply = false;
        expect(await c.disconnectOAuth('a1'), isFalse);
        expect(h.server.count('DisconnectMCPServerOAuth'), 1);

        // Reading another server's configuration does not settle it.
        await c.loadItem('a1');
        expect(c.unconfirmed, isTrue);
        await c.loadOAuth('a1');
        expect(c.unconfirmed, isFalse);
        expect(
          c.oauthOf('a1').status!.state,
          MCPOAuthConnectionState.MCPO_AUTH_CONNECTION_STATE_DISCONNECTED,
        );
      },
    );

    test('disconnect clears the connection and what was known of it', () async {
      final (h, c, _) = await rig(
        seed: (s) => s.connect('alpha', 'a1', at: t0),
      );
      expect(await c.disconnectOAuth('a1'), isTrue);
      expect(
        c.oauthOf('a1').status!.state,
        MCPOAuthConnectionState.MCPO_AUTH_CONNECTION_STATE_DISCONNECTED,
      );
      expect(h.server.oauth['alpha'], isEmpty);
    });

    test('while a change is unconfirmed no authorization is started', () async {
      final (h, c, _) = await rig();
      h.server.loseReply = true;
      await c.delete('a1');
      h.server.loseReply = false;
      await c.connectOAuth('a1');
      expect(c.oauthOf('a1').phase, McpOAuthPhase.failed);
      expect(h.server.count('StartMCPServerOAuth'), 0);
    });

    // ---- Continuations: every await re-checks that this start is current ----

    Future<void> until(bool Function() done) async {
      for (var i = 0; i < 200 && !done(); i++) {
        await Future<void>.delayed(const Duration(milliseconds: 10));
      }
      expect(done(), isTrue, reason: 'condition never held');
    }

    Completer<void> hold(Harness h, String method) {
      final gate = Completer<void>();
      h.server.before = (m) => m == method ? gate.future : Future.value();
      return gate;
    }

    group('a start that was overtaken opens nothing', () {
      Future<void> overtakenBy(
        String name,
        Future<void> Function(Harness, McpController) overtake, {
        String hold_ = 'StartMCPServerOAuth',
      }) async {
        final (h, c, platform) = await rig(
          seed: (s) => s.connect('alpha', 'a1', at: t0),
        );
        final gate = hold(h, hold_);
        final start = c.connectOAuth('a1');
        await until(() => h.server.count(hold_) == 1);
        h.server.before = null;
        await overtake(h, c);
        gate.complete();
        await start;
        expect(
          platform.opened,
          isEmpty,
          reason: '$name must stop the browser from opening',
        );
        expect(c.hasWaitingFlow, isFalse, reason: name);
      }

      test('by a disconnect: the browser stays closed, the state stays '
          'disconnected', () async {
        final (h, c, platform) = await rig(
          seed: (s) => s.connect('alpha', 'a1', at: t0),
        );
        final gate = hold(h, 'StartMCPServerOAuth');
        final start = c.connectOAuth('a1');
        await until(() => h.server.count('StartMCPServerOAuth') == 1);
        h.server.before = null;
        expect(await c.disconnectOAuth('a1'), isTrue);
        gate.complete();
        await start;
        expect(platform.opened, isEmpty);
        final view = c.oauthOf('a1');
        expect(view.phase, McpOAuthPhase.idle);
        expect(
          view.status!.state,
          MCPOAuthConnectionState.MCPO_AUTH_CONNECTION_STATE_DISCONNECTED,
        );
        expect(h.server.oauth['alpha'], isEmpty);
      });

      test('by an edit of the server', () async {
        await overtakenBy('an edit', (h, c) async {
          await c.loadItem('a1');
          final draft = McpDraft.edit(c.item!)..name = 'Renamed';
          expect(await c.update(draft.build()), isTrue);
        });
      });

      test('by deleting the server', () async {
        await overtakenBy('a delete', (h, c) async {
          expect(await c.delete('a1'), isTrue);
        });
      });

      test('by reinstalling the server from its preset', () async {
        await overtakenBy('an install', (h, c) async {
          h.server.presets.add(
            server('a1')
              ..metadata['x'] = 'y'
              ..ensureAuth().type =
                  MCPServerAuthType.MCP_SERVER_AUTH_TYPE_OAUTH2,
          );
          h.server.servers['alpha']!['a1']!.metadata[McpServerFake.presetKey] =
              'a1';
          expect(await c.install('a1'), isTrue);
        });
      });

      test(
        'by a disconnect while the first read is still on the wire',
        () async {
          final (h, c, platform) = await rig(
            seed: (s) => s.connect('alpha', 'a1', at: t0),
          );
          final gate = hold(h, 'GetMCPServerOAuthStatus');
          final start = c.connectOAuth('a1');
          await until(() => h.server.count('GetMCPServerOAuthStatus') == 1);
          h.server.before = null;
          expect(await c.disconnectOAuth('a1'), isTrue);
          gate.complete();
          await start;
          expect(h.server.count('StartMCPServerOAuth'), 0);
          expect(platform.opened, isEmpty);
          expect(c.hasWaitingFlow, isFalse);
        },
      );

      test(
        'and says why when nothing else replaced what it was showing',
        () async {
          // A write that fails leaves the same view; the start must not stay
          // "starting" for ever.
          final (h, c, platform) = await rig();
          final gate = hold(h, 'StartMCPServerOAuth');
          final start = c.connectOAuth('a1');
          await until(() => h.server.count('StartMCPServerOAuth') == 1);
          h.server.before = null;
          h.server.refusals['DeleteMCPServer'] = (
            403,
            'permission_denied',
            'no',
          );
          expect(await c.delete('a1'), isFalse);
          gate.complete();
          await start;
          expect(platform.opened, isEmpty);
          expect(c.oauthOf('a1').phase, McpOAuthPhase.failed);
          expect(c.oauthOf('a1').message, contains('browser was not opened'));
        },
      );
    });

    group('the browser launcher', () {
      test(
        'the flow is recorded as waiting before the launcher answers',
        () async {
          final (h, c, platform) = await rig();
          platform.openGate = Completer<bool>();
          final start = c.connectOAuth('a1');
          await until(() => platform.opened.isNotEmpty);
          final view = c.oauthOf('a1');
          expect(view.phase, McpOAuthPhase.waiting);
          expect(view.launching, isTrue);
          expect(c.hasWaitingFlow, isTrue);
          platform.openGate!.complete(true);
          await start;
          expect(view.launching, isFalse);
          expect(view.phase, McpOAuthPhase.waiting);
          expect(view.message, contains('Finish signing in'));
        },
      );

      test('suspending while the launcher is pending still finds the real '
          'authorization on return', () async {
        final (h, c, platform) = await rig();
        platform.openGate = Completer<bool>();
        final start = c.connectOAuth('a1');
        await until(() => platform.opened.isNotEmpty);
        c.suspend();
        // The person finished authorizing while the app was away.
        h.server.connect('alpha', 'a1', at: t0.add(const Duration(minutes: 1)));
        expect(c.oauthOf('a1').waiting, isTrue);
        await c.recover();
        expect(c.oauthOf('a1').phase, McpOAuthPhase.connected);
        // The launcher answers late; the answer changes nothing.
        platform.openGate!.complete(true);
        await start;
        expect(c.oauthOf('a1').phase, McpOAuthPhase.connected);
        expect(c.oauthOf('a1').launching, isFalse);
        expect(h.server.count('StartMCPServerOAuth'), 1);
      });

      test('a launcher that refuses is a failure, even if answered after a '
          'suspension', () async {
        final (h, c, platform) = await rig();
        platform.openGate = Completer<bool>();
        final start = c.connectOAuth('a1');
        await until(() => platform.opened.isNotEmpty);
        c.suspend();
        platform.openGate!.complete(false);
        await start;
        final view = c.oauthOf('a1');
        expect(view.phase, McpOAuthPhase.failed);
        expect(view.message, contains('Could not open the browser'));
        expect(c.hasWaitingFlow, isFalse);
      });

      test('a launcher that throws is a failure, not a flow', () async {
        final (h, c, platform) = await rig();
        platform.openThrows = true;
        await c.connectOAuth('a1');
        expect(c.oauthOf('a1').phase, McpOAuthPhase.failed);
        expect(c.oauthOf('a1').message, contains('Could not open the browser'));
        expect(c.hasWaitingFlow, isFalse);
      });

      test('an app with no launcher says so and starts nothing it cannot '
          'open', () async {
        final h = await start(seed: (s) => s.seed('alpha', oauthServer('a1')));
        final login = await ButterApi(
          h.client.address!,
        ).login('alice', 'password');
        final c = McpController(
          newApi: () => ButterApi(h.client.address!),
          token: login.token,
          workspaceId: 'alpha',
          onUnauthenticated: () {},
          allowHttp: true,
        );
        addTearDown(c.dispose);
        await c.connectOAuth('a1');
        expect(c.oauthOf('a1').phase, McpOAuthPhase.failed);
        expect(c.hasWaitingFlow, isFalse);
      });

      test('a disconnect while the launcher is pending cannot withdraw the '
          'link, and the state says so', () async {
        final (h, c, platform) = await rig(
          seed: (s) => s.connect('alpha', 'a1', at: t0),
        );
        platform.openGate = Completer<bool>();
        final start = c.connectOAuth('a1');
        await until(() => platform.opened.isNotEmpty);
        expect(await c.disconnectOAuth('a1'), isTrue);
        platform.openGate!.complete(true);
        await start;
        final view = c.oauthOf('a1');
        expect(view.phase, McpOAuthPhase.idle);
        expect(view.message, contains('opened just before this change'));
        expect(c.hasWaitingFlow, isFalse);
      });

      test('a second start is refused while the launcher is pending', () async {
        final (h, c, platform) = await rig();
        platform.openGate = Completer<bool>();
        final first = c.connectOAuth('a1');
        await until(() => platform.opened.isNotEmpty);
        await c.connectOAuth('a1');
        expect(h.server.count('StartMCPServerOAuth'), 1);
        platform.openGate!.complete(true);
        await first;
      });
    });

    group('stale authorization answers', () {
      test('a state read begun before a disconnect is dropped, even when the '
          'disconnect then fails', () async {
        final (h, c, _) = await rig(
          seed: (s) => s.connect('alpha', 'a1', at: t0),
        );
        await c.loadOAuth('a1');
        final shown = c.oauthOf('a1').status!.connectedAt.toDateTime();
        // The connection is renewed; a read taken now carries the newer one.
        h.server.connect('alpha', 'a1', at: t0.add(const Duration(hours: 1)));
        final gate = hold(h, 'GetMCPServerOAuthStatus');
        final stale = c.loadOAuth('a1');
        await until(() => h.server.count('GetMCPServerOAuthStatus') == 2);
        h.server.before = null;
        h.server.refusals['DisconnectMCPServerOAuth'] = (
          403,
          'permission_denied',
          'no',
        );
        expect(await c.disconnectOAuth('a1'), isFalse);
        gate.complete();
        await stale;
        expect(
          c.oauthOf('a1').status!.connectedAt.toDateTime(),
          shown,
          reason: 'a read older than the write cannot be applied after it',
        );
        expect(c.oauthOf('a1').reading, isFalse);
      });

      test('a read begun before a disconnect cannot show the old connection '
          'afterwards', () async {
        final (h, c, _) = await rig(
          seed: (s) => s.connect('alpha', 'a1', at: t0),
        );
        final gate = hold(h, 'GetMCPServerOAuthStatus');
        final stale = c.loadOAuth('a1');
        await until(() => h.server.count('GetMCPServerOAuthStatus') == 1);
        h.server.before = null;
        expect(await c.disconnectOAuth('a1'), isTrue);
        gate.complete();
        await stale;
        expect(
          c.oauthOf('a1').status!.state,
          MCPOAuthConnectionState.MCPO_AUTH_CONNECTION_STATE_DISCONNECTED,
        );
      });

      test('a failed read begun before a write says nothing', () async {
        final (h, c, _) = await rig(
          seed: (s) => s.connect('alpha', 'a1', at: t0),
        );
        final gate = hold(h, 'GetMCPServerOAuthStatus');
        h.server.refusals['GetMCPServerOAuthStatus'] = (
          503,
          'unavailable',
          'later',
        );
        final stale = c.loadOAuth('a1');
        await until(() => h.server.count('GetMCPServerOAuthStatus') == 1);
        h.server.before = null;
        h.server.refusals.clear();
        h.server.refusals['DisconnectMCPServerOAuth'] = (
          403,
          'permission_denied',
          'no',
        );
        expect(await c.disconnectOAuth('a1'), isFalse);
        gate.complete();
        await stale;
        expect(c.oauthOf('a1').message, isNull);
      });

      test(
        'a state read begun after the flow started is what ends a wait',
        () async {
          final (h, c, platform) = await rig();
          await c.connectOAuth('a1');
          expect(c.oauthOf('a1').waiting, isTrue);
          h.server.connect(
            'alpha',
            'a1',
            at: t0.add(const Duration(minutes: 1)),
          );
          await c.loadOAuth('a1');
          expect(c.oauthOf('a1').phase, McpOAuthPhase.connected);
          expect(platform.opened, hasLength(1));
        },
      );
    });
  });

  test(
    'returning to the foreground checks a flow that was being waited on',
    () async {
      final h = await start(seed: (s) => s.seed('alpha', oauthServer('a1')));
      await h.mcp.connectOAuth('a1');
      expect(h.mcp.oauthOf('a1').phase, McpOAuthPhase.waiting);
      h.client.suspend();
      expect(h.mcp.oauthOf('a1').phase, McpOAuthPhase.waiting);
      h.server.connect(
        'alpha',
        'a1',
        at: DateTime.now().add(const Duration(seconds: 5)),
      );
      await h.client.revalidate();
      for (
        var i = 0;
        i < 100 && h.mcp.oauthOf('a1').phase != McpOAuthPhase.connected;
        i++
      ) {
        await Future<void>.delayed(const Duration(milliseconds: 20));
      }
      expect(h.mcp.oauthOf('a1').phase, McpOAuthPhase.connected);
    },
  );

  test('an expired session signs the person out and retries nothing', () async {
    final h = await start(seed: (s) => s.seed('alpha', server('a1')));
    h.server.refusals['ListMCPServers'] = (401, 'unauthenticated', 'expired');
    await h.mcp.load();
    expect(h.server.count('ListMCPServers'), 1);
    await Future<void>.delayed(const Duration(milliseconds: 50));
    expect(h.client.error, contains('Session expired'));
    expect(h.client.mcp, isNull);
  });
}
