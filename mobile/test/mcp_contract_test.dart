import 'dart:io' as io;

import 'package:butter_mobile/app/client_context.dart';
import 'package:butter_mobile/core/butter_api.dart';
import 'package:butter_mobile/core/config_resource_controller.dart';
import 'package:butter_mobile/core/native_oauth.dart';
import 'package:butter_mobile/core/server_address.dart';
import 'package:butter_mobile/core/url_validation.dart';
import 'package:butter_mobile/features/mcp/mcp_controller.dart';
import 'package:butter_mobile/features/mcp/mcp_draft.dart';
import 'package:butter_mobile/features/remote_agents/remote_agent_controller.dart';
import 'package:butter_mobile/features/remote_agents/remote_agent_draft.dart';
import 'package:butter_mobile/gen/agents/v1/agent.pb.dart';
import 'package:butter_mobile/gen/agents/v1/agent_service.pb.dart';
import 'package:connectrpc/connect.dart';
import 'package:fixnum/fixnum.dart';
import 'package:flutter_test/flutter_test.dart';

import 'support/memory_storage.dart';

// The MCP servers, the authorization server, the A2A card and the OpenCode
// health endpoint here are in-process test doubles served by the contract
// fixture (mobile/tool/contractserver/mcp.go). MCPServerService,
// GlobalMCPServerService, RemoteAgentService, the OAuth service and the public
// callback are the production ones. Nothing in this file is a production MCP
// server, a real provider, or a device run.
const workspace = 'mcp';
const deniedWorkspace = 'mcp-denied';
const staticToken = 'mcp-static-secret';
const clientId = 'contract-client';
const clientSecret = 'contract-client-secret';

Matcher code(Code code) =>
    isA<ConnectException>().having((e) => e.code, 'code', code);

/// The system browser, played by the test: it follows the authorization URL to
/// the authorization server and then to Butter's public callback, as a browser
/// would, and keeps the last redirect.
class BrowsingPlatform implements NativeOAuthPlatform {
  Uri? lastRedirect;
  final opened = <Uri>[];

  @override
  Future<bool> openBrowser(Uri url) async {
    opened.add(url);
    final http = io.HttpClient();
    try {
      var next = url;
      for (var hop = 0; hop < 2; hop++) {
        final request = await http.getUrl(next)
          ..followRedirects = false;
        final response = await request.close();
        await response.drain<void>();
        if (response.statusCode != 302) return true;
        next = next.resolve(response.headers.value('location')!);
        lastRedirect = next;
      }
    } finally {
      http.close(force: true);
    }
    return true;
  }

  @override
  Stream<Uri> get links => const Stream.empty();
}

void main() {
  const server = String.fromEnvironment('BUTTER_TEST_SERVER');
  for (final prefix in ['', '/butter']) {
    group(
      'MCP and Remote Agent production contract $prefix',
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

        McpController mcpController({
          String ws = workspace,
          NativeOAuthPlatform? platform,
          DateTime Function()? now,
        }) {
          final c = McpController(
            newApi: () => ButterApi(ServerAddress.parse(base, allowHttp: true)),
            token: token,
            workspaceId: ws,
            onUnauthenticated: () {},
            openBrowser: platform?.openBrowser,
            allowHttp: true,
            now: now,
          );
          addTearDown(c.dispose);
          return c;
        }

        RemoteAgentController remoteController({String ws = workspace}) {
          final c = RemoteAgentController(
            newApi: () => ButterApi(ServerAddress.parse(base, allowHttp: true)),
            token: token,
            workspaceId: ws,
            onUnauthenticated: () {},
          );
          addTearDown(c.dispose);
          return c;
        }

        MCPServer httpServer(String id, String path) => MCPServer(
          id: id,
          name: id,
          transport: MCPServerTransport.MCP_SERVER_TRANSPORT_STREAMABLE_HTTP,
          url: '$server/fake-mcp/$path',
        );

        MCPServer oauthServer(String id, {List<String> scopes = const []}) {
          final s = httpServer(id, 'oauth');
          s.ensureAuth()
            ..type = MCPServerAuthType.MCP_SERVER_AUTH_TYPE_OAUTH2
            ..ensureOauth2().clientId = clientId;
          s.auth.oauth2
            ..clientSecret = clientSecret
            ..scopes.addAll(scopes)
            ..authorizationUrl = '$server/fake-as/authorize'
            ..tokenUrl = '$server/fake-as/token';
          return s;
        }

        test('an MCP server round-trips; unknown fields, metadata and the '
            'filter survive an edit through the draft', () async {
          final c = mcpController();
          final created = httpServer('open-$tag', 'open')
            ..toolFilter.add('echo')
            ..metadata['owner'] = 'someone';
          created.unknownFields.mergeVarintField(99, Int64(7));
          expect(await c.create(created), isTrue, reason: c.writeError);
          expect(c.items.single.workspaceId, workspace);

          await c.loadItem('open-$tag');
          final stored = c.item!;
          expect(stored.metadata['owner'], 'someone');
          expect(stored.unknownFields.getField(99)?.varints.single.toInt(), 7);

          final draft = McpDraft.edit(stored)..name = 'Open renamed';
          expect(draft.firstError(), isNull);
          expect(await c.update(draft.build()), isTrue, reason: c.writeError);
          final after = await api.mcpServer(token, workspace, 'open-$tag');
          expect(after.name, 'Open renamed');
          expect(after.toolFilter, ['echo']);
          expect(after.metadata['owner'], 'someone');
          expect(after.unknownFields.getField(99)?.varints.single.toInt(), 7);

          // Another Workspace does not see it, and a duplicate is refused.
          expect(await api.mcpServers(token, 'alpha'), isEmpty);
          expect(await c.create(created), isFalse);
          expect(c.writeError, contains('already exists'));
          expect(c.unconfirmed, isFalse);

          expect(await c.delete('open-$tag'), isTrue);
          expect(
            (await api.mcpServers(token, workspace)).map((s) => s.id),
            isNot(contains('open-$tag')),
          );
        });

        test(
          'the probe and the tool list come from a real MCP server',
          () async {
            final c = mcpController();
            final s = httpServer('probe-$tag', 'open')..toolFilter.add('echo');
            expect(await c.create(s), isTrue);
            await c.probe('probe-$tag');
            final status = c.statuses['probe-$tag']!;
            expect(status.kind, McpStatusKind.connected);
            expect(status.toolCount, 1);

            await c.loadTools('probe-$tag');
            final view = c.tools['probe-$tag']!;
            expect(view.state, ConfigLoad.ready);
            final allowed = {for (final t in view.tools) t.name: t.allowed};
            expect(allowed, {'echo': true, 'danger': false});
            await c.delete('probe-$tag');
          },
        );

        test('static header credentials are the ones the server sends; a '
            'wrong one reads as rejected credentials, never as text', () async {
          final c = mcpController();
          final good = httpServer('hdr-ok-$tag', 'bearer');
          good.ensureAuth().type =
              MCPServerAuthType.MCP_SERVER_AUTH_TYPE_STATIC_HEADERS;
          good.headers['Authorization'] = 'Bearer $staticToken';
          expect(await c.create(good), isTrue);
          await c.probe('hdr-ok-$tag');
          expect(c.statuses['hdr-ok-$tag']!.kind, McpStatusKind.connected);

          // Replace through the draft with a wrong value.
          await c.loadItem('hdr-ok-$tag');
          final draft = McpDraft.edit(c.item!);
          draft.headers.single.replacement = 'Bearer wrong-token-xyz';
          expect(await c.update(draft.build()), isTrue);
          await c.probe('hdr-ok-$tag');
          final view = c.statuses['hdr-ok-$tag']!;
          expect(
            view.kind,
            anyOf(McpStatusKind.disconnected, McpStatusKind.error),
          );
          expect(view.reason, contains('rejected the credentials'));
          expect(view.reason, isNot(contains('wrong-token-xyz')));
          await c.delete('hdr-ok-$tag');
        });

        test('a server that cannot be reached is an answer with a safe '
            'reason', () async {
          final c = mcpController();
          final s = httpServer('down-$tag', 'broken');
          expect(await c.create(s), isTrue);
          await c.probe('down-$tag');
          final view = c.statuses['down-$tag']!;
          expect(view.kind, isNot(McpStatusKind.connected));
          expect(view.kind, isNot(McpStatusKind.failed));
          expect(view.reason, isNotEmpty);
          await c.delete('down-$tag');

          // Nothing listens there at all.
          final dead = MCPServer(
            id: 'dead-$tag',
            name: 'dead',
            transport: MCPServerTransport.MCP_SERVER_TRANSPORT_STREAMABLE_HTTP,
            url: 'http://127.0.0.1:1/mcp',
          );
          expect(await c.create(dead), isTrue);
          await c.probe('dead-$tag');
          expect(
            c.statuses['dead-$tag']!.reason,
            contains('could not connect'),
          );
          await c.loadTools('dead-$tag');
          expect(c.tools['dead-$tag']!.state, ConfigLoad.failed);
          expect(c.tools['dead-$tag']!.error, contains('could not connect'));
          await c.delete('dead-$tag');
        });

        test(
          'OAuth completes through the server callback and the App reads the '
          'state; disconnect deletes it',
          () async {
            final platform = BrowsingPlatform();
            final c = mcpController(platform: platform);
            expect(await c.create(oauthServer('oauth-$tag')), isTrue);

            await c.loadOAuth('oauth-$tag');
            expect(
              c.oauthOf('oauth-$tag').status!.state,
              MCPOAuthConnectionState.MCPO_AUTH_CONNECTION_STATE_DISCONNECTED,
            );

            await c.connectOAuth('oauth-$tag');
            expect(c.oauthOf('oauth-$tag').phase, McpOAuthPhase.waiting);
            // No return link was sent: the server's callback redirected to its
            // own relative page, and the App never saw a code or token.
            expect(platform.lastRedirect!.path, '/mcp-servers');
            expect(
              platform.lastRedirect!.queryParameters['mcp_oauth'],
              'success',
            );
            expect(
              platform.opened.single.queryParameters['redirect_uri'],
              endsWith('/api/mcp/oauth/callback'),
            );

            await c.loadOAuth('oauth-$tag');
            final view = c.oauthOf('oauth-$tag');
            expect(view.phase, McpOAuthPhase.connected);
            expect(
              view.status!.state,
              MCPOAuthConnectionState.MCPO_AUTH_CONNECTION_STATE_CONNECTED,
            );

            // The stored authorization is what the probe uses.
            await c.probe('oauth-$tag');
            expect(c.statuses['oauth-$tag']!.kind, McpStatusKind.connected);

            expect(await c.disconnectOAuth('oauth-$tag'), isTrue);
            await c.loadOAuth('oauth-$tag');
            expect(
              c.oauthOf('oauth-$tag').status!.state,
              MCPOAuthConnectionState.MCPO_AUTH_CONNECTION_STATE_DISCONNECTED,
            );
            await c.probe('oauth-$tag');
            expect(
              c.statuses['oauth-$tag']!.kind,
              isNot(McpStatusKind.connected),
            );
            expect(c.statuses['oauth-$tag']!.reason, contains('Authorization'));
            await c.delete('oauth-$tag');
          },
        );

        for (final refusal in ['deny', 'tokenfail']) {
          test('a $refusal authorization connects nothing and the wait '
              'times out explicitly', () async {
            final platform = BrowsingPlatform();
            var clock = DateTime.now();
            final c = mcpController(platform: platform, now: () => clock);
            final id = 'oauth-$refusal-$tag';
            expect(await c.create(oauthServer(id, scopes: [refusal])), isTrue);
            await c.connectOAuth(id);
            expect(c.oauthOf(id).phase, McpOAuthPhase.waiting);
            expect(
              platform.lastRedirect!.queryParameters['mcp_oauth'],
              'error',
            );

            await c.loadOAuth(id);
            expect(c.oauthOf(id).phase, McpOAuthPhase.waiting);
            expect(
              c.oauthOf(id).status!.state,
              MCPOAuthConnectionState.MCPO_AUTH_CONNECTION_STATE_DISCONNECTED,
            );

            clock = clock.add(const Duration(minutes: 11));
            await c.loadOAuth(id);
            expect(c.oauthOf(id).phase, McpOAuthPhase.timedOut);
            await c.delete(id);
          });
        }

        test('starting OAuth on a server that is not OAuth, or has an '
            'incomplete client, is refused with a reason', () async {
          final platform = BrowsingPlatform();
          final c = mcpController(platform: platform);
          expect(await c.create(httpServer('plain-$tag', 'open')), isTrue);
          await c.connectOAuth('plain-$tag');
          expect(c.oauthOf('plain-$tag').phase, McpOAuthPhase.failed);
          expect(c.oauthOf('plain-$tag').message, isNotNull);
          expect(platform.opened, isEmpty);
          await c.delete('plain-$tag');
        });

        test('global presets are listed redacted and installed; an installed '
            'preset works and cannot be edited in the App', () async {
          final c = mcpController();
          await c.loadPresets();
          final ids = c.presets.map((p) => p.id).toSet();
          expect(
            ids,
            containsAll(['preset-open', 'preset-secret', 'preset-oauth']),
          );
          for (final p in c.presets) {
            expect(p.headers.values.every((v) => v == '***'), isTrue);
            expect(p.writeToBuffer().toString(), isNot(contains(staticToken)));
            expect(p.auth.oauth2.clientSecret, isEmpty);
          }

          expect(
            await c.install('preset-secret'),
            isTrue,
            reason: c.writeError,
          );
          final installed = c.byId('preset-secret')!;
          expect(isInstalledPreset(installed), isTrue);
          expect(installed.headers.values.every((v) => v == '***'), isTrue);
          expect(mcpEditBlock(installed), isNotNull);
          // The server still sends the real header.
          await c.probe('preset-secret');
          expect(c.statuses['preset-secret']!.kind, McpStatusKind.connected);

          // Reinstalling is the one way to change it.
          expect(await c.install('preset-secret'), isTrue);

          // No write to presets is offered by the client at all.
          expect(await c.install('preset-open'), isTrue);
          expect(mcpEditBlock(c.byId('preset-open')!), isNull);

          await c.delete('preset-secret');
          await c.delete('preset-open');
        });

        test('installing a preset that does not exist is NotFound, and one '
            'over a server that is not that preset is refused', () async {
          final c = mcpController();
          expect(await c.install('no-such-preset'), isFalse);
          expect(c.writeError, contains('no longer offered'));
          expect(c.unconfirmed, isFalse);

          expect(await c.create(httpServer('preset-open', 'open')), isTrue);
          expect(await c.install('preset-open'), isFalse);
          expect(c.writeError, contains('not installed from this preset'));
          await c.delete('preset-open');
        });

        test('a Workspace the user is not a member of is refused, as '
            'permission denied, for reads and writes alike', () async {
          await expectLater(
            api.mcpServers(token, deniedWorkspace),
            throwsA(code(Code.permissionDenied)),
          );
          final c = mcpController(ws: deniedWorkspace);
          await c.load();
          expect(c.listState, ConfigLoad.failed);
          expect(c.listError, contains('do not have access'));
          expect(await c.create(httpServer('x-$tag', 'open')), isFalse);
          expect(c.writeError, contains('do not have access'));
          expect(c.unconfirmed, isFalse);

          final r = remoteController(ws: deniedWorkspace);
          await r.load();
          expect(r.listState, ConfigLoad.failed);
        });

        test('Remote Agents: A2A, OpenCode credentials and probes against the '
            'real service', () async {
          final c = remoteController();
          final a2a = RemoteAgentDraft.create()
            ..id = 'a2a-$tag'
            ..name = 'A2A'
            ..url = '$server/fake-a2a/ok';
          expect(a2a.firstError(), isNull);
          expect(await c.create(a2a.build()), isTrue, reason: c.writeError);
          await c.probe('a2a-$tag');
          expect(c.probes['a2a-$tag']!.kind, RemoteProbeKind.reachable);

          final gone = RemoteAgentDraft.create()
            ..id = 'a2a-gone-$tag'
            ..name = 'Gone'
            ..url = '$server/fake-a2a/gone';
          expect(await c.create(gone.build()), isTrue);
          await c.probe('a2a-gone-$tag');
          expect(c.probes['a2a-gone-$tag']!.kind, RemoteProbeKind.unreachable);
          expect(c.probes['a2a-gone-$tag']!.reason, contains('HTTP 404'));

          final oc = RemoteAgentDraft.create()
            ..id = 'oc-$tag'
            ..name = 'Code'
            ..kind = RemoteKind.opencode
            ..url = '$server/fake-opencode/ok'
            ..opencodeAgent = 'build'
            ..username = 'opencode-user'
            ..passwordEdit = SecretEdit.replace
            ..newPassword = 'opencode-secret';
          expect(await c.create(oc.build()), isTrue, reason: c.writeError);
          await c.probe('oc-$tag');
          expect(c.probes['oc-$tag']!.kind, RemoteProbeKind.reachable);

          // Keeping the password through an edit keeps the endpoint working.
          await c.loadItem('oc-$tag');
          final keep = RemoteAgentDraft.edit(c.item!)..opencodeModel = 'a/b';
          expect(await c.update(keep.build()), isTrue);
          await c.probe('oc-$tag');
          expect(c.probes['oc-$tag']!.kind, RemoteProbeKind.reachable);
          expect(c.item!.opencodeModel, 'a/b');

          // Replacing it with a wrong one is seen by the probe.
          await c.loadItem('oc-$tag');
          final wrong = RemoteAgentDraft.edit(c.item!)
            ..passwordEdit = SecretEdit.replace
            ..newPassword = 'wrong';
          expect(await c.update(wrong.build()), isTrue);
          await c.probe('oc-$tag');
          expect(c.probes['oc-$tag']!.kind, RemoteProbeKind.error);
          expect(
            c.probes['oc-$tag']!.reason,
            contains('rejected the username or password'),
          );

          // Clearing the credentials (username too) sends none.
          await c.loadItem('oc-$tag');
          final clear = RemoteAgentDraft.edit(c.item!)
            ..username = ''
            ..passwordEdit = SecretEdit.clear;
          expect(clear.firstError(), isNull);
          expect(await c.update(clear.build()), isTrue);
          final stored = await api.remoteAgent(token, workspace, 'oc-$tag');
          expect(stored.password, isEmpty);
          expect(stored.username, isEmpty);

          for (final id in ['a2a-$tag', 'a2a-gone-$tag', 'oc-$tag']) {
            expect(await c.delete(id), isTrue);
          }
        });

        test('Remote Agent refusals: a username without a password, and '
            'a daemon runtime on a non-daemon protocol', () async {
          final c = remoteController();
          final bad = RemoteAgent(
            id: 'bad-$tag',
            name: 'Bad',
            protocol: RemoteAgentProtocol.REMOTE_AGENT_PROTOCOL_OPENCODE_HTTP,
            url: '$server/fake-opencode/open',
            username: 'u',
          );
          expect(await c.create(bad), isFalse);
          expect(c.writeError, contains('rejected these values'));
          expect(c.unconfirmed, isFalse);

          // The draft never sends daemon routing, even from a stored copy.
          final stored = RemoteAgent(
            id: 'bad2-$tag',
            name: 'Bad2',
            protocol: RemoteAgentProtocol.REMOTE_AGENT_PROTOCOL_A2A,
            url: '$server/fake-a2a/ok',
            daemonRuntimeId: 'rt',
          );
          expect(
            RemoteAgentDraft.edit(stored).build().daemonRuntimeId,
            isEmpty,
          );
        });

        test(
          'the signed-in context scopes both controllers and drops them with '
          'the Workspace',
          () async {
            final client = ClientContext(
              storage: MemoryStorage(),
              allowHttp: true,
            );
            addTearDown(client.dispose);
            await client.setServer(base);
            await client.login('mobile-test', 'contract-test-password');
            await client.selectWorkspace(workspace);
            final mcp = client.mcp!;
            final remote = client.remoteAgents!;
            await mcp.loadPresets();
            expect(mcp.presets, isNotEmpty);
            await remote.load();
            expect(remote.listState, ConfigLoad.ready);
            await client.selectWorkspace('alpha');
            expect(mcp.disposed, isTrue);
            expect(remote.disposed, isTrue);
            expect(identical(client.mcp, mcp), isFalse);
          },
        );
      },
    );
  }
}
