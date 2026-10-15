import 'dart:async';
import 'dart:convert';
import 'dart:io';

import 'package:butter_mobile/gen/agents/v1/agent.pb.dart';
import 'package:butter_mobile/gen/agents/v1/agent_service.pb.dart';
import 'package:butter_mobile/gen/agents/v1/dashboard.pb.dart';
import 'package:fixnum/fixnum.dart';
import 'package:protobuf/well_known_types/google/protobuf/timestamp.pb.dart';

import 'agent_server.dart';

/// External HTTP boundary for MCPServerService, GlobalMCPServerService and
/// RemoteAgentService with controllable late, refused and lost replies. The
/// real semantics (credentials, redaction, probes, OAuth completion) are proven
/// against the production Go handlers in mcp_contract_test; this only drives UI
/// timing and failure shapes.
class McpServerFake extends AgentServer {
  final servers = <String, Map<String, MCPServer>>{'alpha': {}, 'beta': {}};
  final remotes = <String, Map<String, RemoteAgent>>{'alpha': {}, 'beta': {}};
  final presets = <MCPServer>[];
  final oauth = <String, Map<String, MCPOAuthConnectionStatus>>{
    'alpha': {},
    'beta': {},
  };
  final mcpCalls = <({String method, String? workspace, List<int> body})>[];

  static const writeMethods = {
    'CreateMCPServer',
    'UpdateMCPServer',
    'DeleteMCPServer',
    'DisconnectMCPServerOAuth',
    'StartMCPServerOAuth',
    'InstallGlobalMCPServer',
    'CreateRemoteAgent',
    'UpdateRemoteAgent',
    'DeleteRemoteAgent',
  };
  int count(String method) => mcpCalls.where((c) => c.method == method).length;
  int get writes =>
      mcpCalls.where((c) => writeMethods.contains(c.method)).length;

  Future<void> Function(String method)? before;

  /// Refuses a method (status, code, message), decided when it arrives.
  final refusals = <String, (int, String, String)>{};

  /// A write that is applied but whose answer is lost (503).
  bool loseReply = false;

  MCPServerStatus_State statusState = MCPServerStatus_State.STATE_CONNECTED;
  String statusDetail = '';
  List<MCPTool> toolList = [];
  Map<String, String> toolErrors = {};
  RemoteAgentStatus_State remoteState = RemoteAgentStatus_State.STATE_ACTIVE;
  String remoteDetail = '';

  /// What StartMCPServerOAuth answers.
  String authorizationUrl = 'https://auth.example.test/authorize?x=1';

  McpServerFake() {
    final inherited = authorizedReply!;
    authorizedReply = (r, bytes) async {
      final path = r.uri.path;
      if (!path.contains('MCPServerService') &&
          !path.contains('GlobalMCPServerService') &&
          !path.contains('RemoteAgentService')) {
        return inherited(r, bytes);
      }
      final method = path.split('/').last;
      final ws = r.headers.value('x-workspace-id');
      mcpCalls.add((method: method, workspace: ws, body: bytes));
      final bucket = servers[ws];
      if (bucket == null) {
        _fail(r, 403, 'permission_denied', 'workspace access denied');
        return true;
      }
      final refused = refusals[method];
      if (refused != null &&
          !method.startsWith('Get') &&
          !method.startsWith('List')) {
        _fail(r, refused.$1, refused.$2, refused.$3);
        return true;
      }
      final remote = remotes[ws]!;
      // A write is held here (when a test asks) before it is applied, so its
      // answer is late the way a slow network makes it.
      if (writeMethods.contains(method)) await before?.call(method);
      // Reads capture their answer when the request arrives, so a delayed
      // answer is a stale one.
      switch (method) {
        case 'ListMCPServers':
          final snapshot = [for (final s in bucket.values) _view(s)];
          await before?.call(method);
          if (refused != null) {
            return _fail(r, refused.$1, refused.$2, refused.$3);
          }
          write(r, ListMCPServersResponse(mcpServers: snapshot));
        case 'GetMCPServer':
          final id = GetMCPServerRequest.fromBuffer(bytes).id;
          final found = bucket[id] == null ? null : _view(bucket[id]!);
          await before?.call(method);
          if (refused != null) {
            return _fail(r, refused.$1, refused.$2, refused.$3);
          }
          if (found == null) {
            return _fail(r, 404, 'not_found', 'mcp server not found');
          }
          write(r, GetMCPServerResponse(mcpServer: found));
        case 'CreateMCPServer':
          final s = CreateMCPServerRequest.fromBuffer(bytes).mcpServer;
          if (bucket.containsKey(s.id)) {
            return _fail(r, 409, 'already_exists', 'mcp server already exists');
          }
          final stored = s.deepCopy()..workspaceId = ws!;
          bucket[s.id] = stored;
          _answer(r, CreateMCPServerResponse(mcpServer: stored));
        case 'UpdateMCPServer':
          final req = UpdateMCPServerRequest.fromBuffer(bytes);
          final s = req.mcpServer.deepCopy();
          final prev = bucket[s.id];
          if (prev == null) {
            return _fail(r, 404, 'not_found', 'mcp server not found');
          }
          // Mirrors the server: kept credentials are named, never inferred.
          final keeping =
              req.keepHeaderKeys.isNotEmpty || req.keepOauthClientSecret;
          if (keeping) {
            for (final k in req.keepHeaderKeys) {
              s.headers[k] = prev.headers[k] ?? '';
            }
            if (req.keepOauthClientSecret) {
              s.ensureAuth().ensureOauth2().clientSecret =
                  prev.auth.oauth2.clientSecret;
            }
            final marker = prev.metadata[presetKey];
            if (marker != null) s.metadata[presetKey] = marker;
          }
          bucket[s.id] = s.deepCopy();
          _answer(
            r,
            UpdateMCPServerResponse(mcpServer: keeping ? _masked(s) : s),
          );
        case 'DeleteMCPServer':
          final id = DeleteMCPServerRequest.fromBuffer(bytes).id;
          if (bucket.remove(id) == null) {
            return _fail(r, 404, 'not_found', 'mcp server not found');
          }
          _answer(r, DeleteMCPServerResponse());
        case 'GetMCPServerStatus':
          final id = GetMCPServerStatusRequest.fromBuffer(bytes).id;
          await before?.call(method);
          if (refused != null) {
            return _fail(r, refused.$1, refused.$2, refused.$3);
          }
          write(
            r,
            GetMCPServerStatusResponse(
              status: MCPServerStatus(
                id: id,
                state: statusState,
                detail: statusDetail,
                toolCount: toolList.length,
              ),
            ),
          );
        case 'ListMCPTools':
          await before?.call(method);
          if (refused != null) {
            return _fail(r, refused.$1, refused.$2, refused.$3);
          }
          write(
            r,
            ListMCPToolsResponse(tools: toolList, errors: toolErrors.entries),
          );
        case 'StartMCPServerOAuth':
          final id = StartMCPServerOAuthRequest.fromBuffer(bytes);
          // The App sends no return URL: the server's own callback finishes.
          if (id.returnUrl.isNotEmpty) {
            return _fail(r, 400, 'invalid_argument', 'unexpected return_url');
          }
          await before?.call(method);
          _answer(
            r,
            StartMCPServerOAuthResponse(
              authorizationUrl: authorizationUrl,
              flowId: 'flow-1',
            ),
          );
        case 'GetMCPServerOAuthStatus':
          final id = GetMCPServerOAuthStatusRequest.fromBuffer(bytes).serverId;
          final status =
              oauth[ws]![id] ??
              MCPOAuthConnectionStatus(
                serverId: id,
                state: MCPOAuthConnectionState
                    .MCPO_AUTH_CONNECTION_STATE_DISCONNECTED,
              );
          final snapshot = status.deepCopy();
          await before?.call(method);
          if (refused != null) {
            return _fail(r, refused.$1, refused.$2, refused.$3);
          }
          write(r, GetMCPServerOAuthStatusResponse(status: snapshot));
        case 'DisconnectMCPServerOAuth':
          final id = DisconnectMCPServerOAuthRequest.fromBuffer(bytes).serverId;
          oauth[ws]!.remove(id);
          _answer(r, DisconnectMCPServerOAuthResponse());
        case 'ListGlobalMCPServers':
          await before?.call(method);
          if (refused != null) {
            return _fail(r, refused.$1, refused.$2, refused.$3);
          }
          write(
            r,
            ListGlobalMCPServersResponse(
              mcpServers: [for (final p in presets) p.deepCopy()],
            ),
          );
        case 'InstallGlobalMCPServer':
          final id = InstallGlobalMCPServerRequest.fromBuffer(bytes).id;
          final preset = presets.where((p) => p.id == id).firstOrNull;
          if (preset == null) {
            return _fail(r, 404, 'not_found', 'global mcp server not found');
          }
          final existing = bucket[id];
          if (existing != null &&
              (existing.metadata['butter.global_mcp_preset_id'] ?? '') != id) {
            return _fail(r, 409, 'already_exists', 'mcp server already exists');
          }
          final installed = preset.deepCopy()
            ..workspaceId = ws!
            ..metadata['butter.global_mcp_preset_id'] = id;
          bucket[id] = installed;
          _answer(
            r,
            InstallGlobalMCPServerResponse(mcpServer: _masked(installed)),
          );
        case 'ListRemoteAgents':
          final snapshot = [for (final a in remote.values) a.deepCopy()];
          await before?.call(method);
          if (refused != null) {
            return _fail(r, refused.$1, refused.$2, refused.$3);
          }
          write(r, ListRemoteAgentsResponse(remoteAgents: snapshot));
        case 'GetRemoteAgent':
          final id = GetRemoteAgentRequest.fromBuffer(bytes).id;
          final found = remote[id]?.deepCopy();
          await before?.call(method);
          if (refused != null) {
            return _fail(r, refused.$1, refused.$2, refused.$3);
          }
          if (found == null) {
            return _fail(r, 404, 'not_found', 'remote agent not found');
          }
          write(r, GetRemoteAgentResponse(remoteAgent: found));
        case 'CreateRemoteAgent':
          final a = CreateRemoteAgentRequest.fromBuffer(bytes).remoteAgent;
          if (remote.containsKey(a.id)) {
            return _fail(
              r,
              409,
              'already_exists',
              'remote agent already exists',
            );
          }
          final stored = a.deepCopy()..workspaceId = ws!;
          remote[a.id] = stored;
          _answer(r, CreateRemoteAgentResponse(remoteAgent: stored));
        case 'UpdateRemoteAgent':
          final a = UpdateRemoteAgentRequest.fromBuffer(bytes).remoteAgent;
          if (!remote.containsKey(a.id)) {
            return _fail(r, 404, 'not_found', 'remote agent not found');
          }
          remote[a.id] = a.deepCopy();
          _answer(r, UpdateRemoteAgentResponse(remoteAgent: a));
        case 'DeleteRemoteAgent':
          final id = DeleteRemoteAgentRequest.fromBuffer(bytes).id;
          if (remote.remove(id) == null) {
            return _fail(r, 404, 'not_found', 'remote agent not found');
          }
          _answer(r, DeleteRemoteAgentResponse());
        case 'GetRemoteAgentStatus':
          final id = GetRemoteAgentStatusRequest.fromBuffer(bytes).id;
          await before?.call(method);
          if (refused != null) {
            return _fail(r, refused.$1, refused.$2, refused.$3);
          }
          write(
            r,
            GetRemoteAgentStatusResponse(
              status: RemoteAgentStatus(
                id: id,
                state: remoteState,
                detail: remoteDetail,
                latencyMs: Int64(12),
                checkedAt: Timestamp(seconds: Int64(1)),
              ),
            ),
          );
        default:
          return false;
      }
      return true;
    };
  }

  /// The browser finished the authorization: the server now holds a connection.
  void connect(String ws, String id, {required DateTime at}) {
    oauth[ws]![id] = MCPOAuthConnectionStatus(
      serverId: id,
      state: MCPOAuthConnectionState.MCPO_AUTH_CONNECTION_STATE_CONNECTED,
      connectedAt: Timestamp.fromDateTime(at.toUtc()),
    );
  }

  MCPServer seed(String ws, MCPServer server) {
    servers[ws]![server.id] = server..workspaceId = ws;
    return server;
  }

  RemoteAgent seedRemote(String ws, RemoteAgent agent) {
    remotes[ws]![agent.id] = agent..workspaceId = ws;
    return agent;
  }

  static const presetKey = 'butter.global_mcp_preset_id';

  /// What a read of the server shows: an installed preset hides its
  /// credentials, as the server does.
  MCPServer _view(MCPServer s) =>
      (s.metadata[presetKey] ?? '').isEmpty ? s.deepCopy() : _masked(s);

  MCPServer _masked(MCPServer s) {
    final c = s.deepCopy();
    for (final k in c.headers.keys.toList()) {
      c.headers[k] = '***';
    }
    if (c.hasAuth() && c.auth.hasOauth2()) c.auth.oauth2.clearClientSecret();
    return c;
  }

  void _answer(HttpRequest r, dynamic message) {
    if (loseReply) {
      _fail(r, 503, 'unavailable', 'connection lost');
    } else {
      write(r, message);
    }
  }

  bool _fail(HttpRequest r, int status, String code, String message) {
    r.response.statusCode = status;
    r.response.headers.contentType = ContentType.json;
    r.response.write(jsonEncode({'code': code, 'message': message}));
    return true;
  }
}
