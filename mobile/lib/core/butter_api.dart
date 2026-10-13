import 'dart:async';
import 'dart:io';
import 'dart:typed_data';

import 'package:connectrpc/connect.dart' as connect;
import 'package:connectrpc/protobuf.dart';
import 'package:connectrpc/protocol/connect.dart' as protocol;

import '../gen/agents/v1/agent_service.connect.client.dart';
import '../gen/agents/v1/agent_service.pb.dart';
import '../gen/agents/v1/auth.connect.client.dart';
import '../gen/agents/v1/auth.pb.dart';
import '../gen/agents/v1/workspace.connect.client.dart';
import '../gen/agents/v1/workspace.pb.dart';
import 'server_address.dart';

/// A disposable request scope. No retry interceptor or shared credentials.
final class ButterApi {
  ButterApi(this.address) {
    _client.connectionTimeout = const Duration(seconds: 10);
    final transport = protocol.Transport(
      baseUrl: address.endpoint('api').toString(),
      codec: const ProtoCodec(),
      httpClient: _send,
    );
    _auth = AuthServiceClient(transport);
    _workspaces = WorkspaceServiceClient(transport);
    _agents = AgentServiceClient(transport);
  }

  final ServerAddress address;
  final HttpClient _client = HttpClient();
  late final AuthServiceClient _auth;
  late final WorkspaceServiceClient _workspaces;
  late final AgentServiceClient _agents;

  connect.AbortSignal get _deadline =>
      connect.TimeoutSignal(const Duration(seconds: 20));
  connect.Headers _headers(String token, [String? workspace]) =>
      connect.Headers()
        ..set('Authorization', ['Bearer $token'])
        ..addAll(
          workspace == null
              ? connect.Headers()
              : (connect.Headers()..set('X-Workspace-ID', [workspace])),
        );

  /// Uses a public generated RPC too, so an arbitrary HTTP 200 is not enough.
  Future<void> verify() async {
    final request = await _client.getUrl(address.endpoint('ping'));
    request.followRedirects = false;
    final response = await request.close().timeout(const Duration(seconds: 10));
    await response.drain<void>().timeout(const Duration(seconds: 10));
    if (response.statusCode != 200) {
      throw connect.ConnectException(
        connect.Code.unavailable,
        'Server check failed.',
      );
    }
    await _auth.listOAuthProviders(
      ListOAuthProvidersRequest(),
      signal: _deadline,
    );
  }

  Future<LoginResponse> login(String username, String password) => _auth.login(
    LoginRequest(username: username, password: password),
    signal: _deadline,
  );
  Future<MeResponse> me(String token) =>
      _auth.me(MeRequest(), headers: _headers(token), signal: _deadline);
  Future<ListWorkspacesResponse> workspaces(String token) =>
      _workspaces.listWorkspaces(
        ListWorkspacesRequest(),
        headers: _headers(token),
        signal: _deadline,
      );
  Future<ListAgentsResponse> agents(
    String token,
    String workspace, {
    String pageToken = '',
  }) => _agents.listAgents(
    ListAgentsRequest(pageSize: 50, pageToken: pageToken),
    headers: _headers(token, workspace),
    signal: _deadline,
  );
  Future<LogoutResponse> logout(String token) => _auth.logout(
    LogoutRequest(),
    headers: _headers(token),
    signal: _deadline,
  );

  // The stock native adapter follows redirects. Explicitly refuse them so an
  // endpoint cannot redirect a credential-bearing request to another address.
  // Binary serialization and Connect error decoding remain in the SDK.
  Future<connect.HttpResponse> _send(connect.HttpRequest input) async {
    final request = await _client.openUrl(input.method, Uri.parse(input.url));
    request.followRedirects = false;
    for (final header in input.header.entries) {
      request.headers.add(header.name, header.value);
    }
    input.signal?.future.then((error) => request.abort(error));
    if (input.body case final body?) {
      await for (final chunk in body) {
        request.add(chunk);
      }
    }
    final response = await request.close();
    final headers = connect.Headers();
    response.headers.forEach((key, values) {
      // dart:io automatically decompresses gzip; remove the stale headers.
      if (response.headers.value('content-encoding') == 'gzip' &&
          (key == 'content-encoding' || key == 'content-length')) {
        return;
      }
      for (final value in values) {
        headers.add(key, value);
      }
    });
    return connect.HttpResponse(
      response.statusCode,
      headers,
      response.map(Uint8List.fromList),
      connect.Headers(),
    );
  }

  void close() => _client.close(force: true);
}
