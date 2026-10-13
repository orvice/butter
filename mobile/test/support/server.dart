import 'dart:async';
import 'dart:convert';
import 'dart:io';

import 'package:butter_mobile/gen/agents/v1/agent.pb.dart';
import 'package:butter_mobile/gen/agents/v1/agent_service.pb.dart';
import 'package:butter_mobile/gen/agents/v1/auth.pb.dart';
import 'package:butter_mobile/gen/agents/v1/workspace.pb.dart';
import 'package:protobuf/protobuf.dart';

/// Controllable external HTTP boundary; no mock of application collaborators.
class TestServer {
  late final HttpServer _server;
  String get url => 'http://127.0.0.1:${_server.port}';
  final requests = <({String path, String? token, String? workspace})>[];
  Future<ListAgentsResponse> Function(String, ListAgentsRequest)? agentReply;
  Future<LoginResponse> Function(LoginRequest)? loginReply;
  bool expired = false;
  bool denyAgents = false;
  bool denyWorkspaces = false;
  bool failLogout = false;
  Future<void> Function()? beforeLogout;

  Future<void> start() async {
    _server = await HttpServer.bind(InternetAddress.loopbackIPv4, 0);
    _server.listen((r) async {
      final path = r.uri.path;
      final token = r.headers.value('authorization');
      final workspace = r.headers.value('x-workspace-id');
      requests.add((path: path, token: token, workspace: workspace));
      final bytes = await r.fold<List<int>>([], (a, b) => a..addAll(b));
      if (path == '/ping') {
        r.response.write('{"service":"butter"}');
      } else if (path.endsWith('/Login')) {
        final input = LoginRequest.fromBuffer(bytes);
        if (input.password != 'password') {
          error(r, 401, 'unauthenticated');
          return;
        }
        write(
          r,
          loginReply == null
              ? LoginResponse(
                  token: input.username,
                  user: User(id: input.username, username: input.username),
                )
              : await loginReply!(input),
        );
      } else if (path.endsWith('/ListOAuthProviders')) {
        write(r, ListOAuthProvidersResponse());
      } else if (token == null || expired) {
        error(r, 401, 'unauthenticated');
        return;
      } else if (path.endsWith('/Me')) {
        final name = token.substring(7);
        write(
          r,
          MeResponse(
            user: User(
              id: name,
              username: name,
              role: name == 'admin' ? 'admin' : 'user',
            ),
          ),
        );
      } else if (path.endsWith('/ListWorkspaces')) {
        if (denyWorkspaces) {
          error(r, 403, 'permission_denied');
          return;
        }
        write(
          r,
          ListWorkspacesResponse(
            workspaces: [
              Workspace(id: 'alpha', name: 'Alpha'),
              Workspace(id: 'beta', name: 'Beta'),
            ],
          ),
        );
      } else if (path.endsWith('/ListAgents')) {
        if (denyAgents) {
          error(r, 403, 'permission_denied');
          return;
        }
        final input = ListAgentsRequest.fromBuffer(bytes);
        write(
          r,
          agentReply == null
              ? ListAgentsResponse(
                  agents: [Agent(agentId: 'shared', name: '$workspace agent')],
                )
              : await agentReply!(workspace!, input),
        );
      } else if (path.endsWith('/Logout')) {
        await beforeLogout?.call();
        if (failLogout) {
          error(r, 503, 'unavailable');
          return;
        }
        write(r, LogoutResponse());
      } else {
        error(r, 404, 'not_found');
        return;
      }
      await r.response.close();
    });
  }

  void write(HttpRequest r, GeneratedMessage message) {
    r.response.headers.contentType = ContentType('application', 'proto');
    r.response.add(message.writeToBuffer());
  }

  void error(HttpRequest r, int status, String code) {
    r.response.statusCode = status;
    r.response.headers.contentType = ContentType.json;
    r.response.write(jsonEncode({'code': code, 'message': 'Request failed'}));
    r.response.close();
  }

  Future<void> close() => _server.close(force: true);
}
