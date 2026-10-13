import 'dart:async';
import 'dart:convert';
import 'dart:math';
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
import 'avatar.dart';

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
  /// Returns the OAuth providers that RPC reported.
  Future<List<OAuthProvider>> verify() async {
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
    return (await _auth.listOAuthProviders(
      ListOAuthProvidersRequest(),
      signal: _deadline,
    )).providers;
  }

  Future<List<OAuthProvider>> oauthProviders() async =>
      (await _auth.listOAuthProviders(
        ListOAuthProvidersRequest(),
        signal: _deadline,
      )).providers;

  /// Public: begins a native sign-in with the App's return URI and PKCE
  /// challenge. Carries no credential.
  Future<BeginOAuthFlowResponse> beginOAuth(
    String provider,
    String returnUri,
    String codeChallenge,
  ) => _auth.beginOAuthFlow(
    BeginOAuthFlowRequest(
      provider: provider,
      native: NativeOAuthClient(
        returnUri: returnUri,
        codeChallenge: codeChallenge,
      ),
    ),
    signal: _deadline,
  );

  /// Public: exchanges the browser's one-time handoff code, proven by the PKCE
  /// verifier, for a session. Carries no credential.
  Future<RedeemOAuthHandoffResponse> redeemOAuth(
    String provider,
    String state,
    String handoffCode,
    String codeVerifier,
  ) => _auth.redeemOAuthHandoff(
    RedeemOAuthHandoffRequest(
      provider: provider,
      state: state,
      handoffCode: handoffCode,
      codeVerifier: codeVerifier,
    ),
    signal: _deadline,
  );

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

  Future<GetWorkspaceResponse> workspace(String token, String id) =>
      _workspaces.getWorkspace(
        GetWorkspaceRequest(id: id),
        headers: _headers(token),
        signal: _deadline,
      );
  Future<ListWorkspaceMembersResponse> members(String token, String id) =>
      _workspaces.listWorkspaceMembers(
        ListWorkspaceMembersRequest(workspaceId: id),
        headers: _headers(token),
        signal: _deadline,
      );
  Future<Workspace> saveWorkspace(String token, Workspace workspace) async {
    if (workspace.id.isEmpty) {
      return (await _workspaces.createWorkspace(
        CreateWorkspaceRequest(workspace: workspace),
        headers: _headers(token),
        signal: _deadline,
      )).workspace;
    }
    return (await _workspaces.updateWorkspace(
      UpdateWorkspaceRequest(workspace: workspace),
      headers: _headers(token),
      signal: _deadline,
    )).workspace;
  }

  Future<void> deleteWorkspace(String token, String id) =>
      _workspaces.deleteWorkspace(
        DeleteWorkspaceRequest(id: id),
        headers: _headers(token),
        signal: _deadline,
      );
  Future<WorkspaceMember> addMember(
    String token,
    String id,
    String userId,
    String role,
  ) async => (await _workspaces.addWorkspaceMember(
    AddWorkspaceMemberRequest(workspaceId: id, userId: userId, role: role),
    headers: _headers(token),
    signal: _deadline,
  )).member;
  Future<WorkspaceMember> updateMember(
    String token,
    String id,
    String userId,
    String role,
  ) async => (await _workspaces.updateWorkspaceMember(
    UpdateWorkspaceMemberRequest(workspaceId: id, userId: userId, role: role),
    headers: _headers(token),
    signal: _deadline,
  )).member;
  Future<void> removeMember(String token, String id, String userId) =>
      _workspaces.removeWorkspaceMember(
        RemoveWorkspaceMemberRequest(workspaceId: id, userId: userId),
        headers: _headers(token),
        signal: _deadline,
      );

  Future<UpdateProfileResponse> profile(
    String token,
    String name, {
    String? avatarUrl,
  }) => _auth.updateProfile(
    UpdateProfileRequest(displayName: name, avatarUrl: avatarUrl),
    headers: _headers(token),
    signal: _deadline,
  );
  Future<ChangePasswordResponse> changePassword(
    String token,
    String current,
    String next,
  ) => _auth.changePassword(
    ChangePasswordRequest(currentPassword: current, newPassword: next),
    headers: _headers(token),
    signal: _deadline,
  );

  Future<String> uploadAvatar(String token, Uint8List bytes) async {
    final type = avatarContentType(bytes);
    final suffix = type.split('/').last == 'jpeg'
        ? 'jpg'
        : type.split('/').last;
    final random = Random.secure();
    final boundary =
        'butter-${List.generate(24, (_) => random.nextInt(256).toRadixString(16).padLeft(2, '0')).join()}';
    final request = await _client.postUrl(
      address.endpoint('api/uploads/avatar'),
    );
    request.followRedirects = false;
    request.headers.set(HttpHeaders.authorizationHeader, 'Bearer $token');
    request.headers.set(
      HttpHeaders.contentTypeHeader,
      'multipart/form-data; boundary=$boundary',
    );
    request.add(
      utf8.encode(
        '--$boundary\r\nContent-Disposition: form-data; name="file"; filename="avatar.$suffix"\r\nContent-Type: $type\r\n\r\n',
      ),
    );
    request.add(bytes);
    request.add(utf8.encode('\r\n--$boundary--\r\n'));
    try {
      final response = await request.close().timeout(
        const Duration(seconds: 20),
      );
      final body = await response
          .fold<List<int>>([], (buffer, chunk) {
            if (buffer.length + chunk.length > 64 * 1024) {
              throw StateError('Upload reply too large');
            }
            return buffer..addAll(chunk);
          })
          .timeout(const Duration(seconds: 20));
      if (response.statusCode != 200) {
        if (response.statusCode == 413) {
          throw const FormatException(
            'The image exceeds the server upload limit. Choose a smaller image.',
          );
        }
        if (response.statusCode == 415) {
          throw const FormatException('The server rejected this image type.');
        }
        final code = switch (response.statusCode) {
          401 => connect.Code.unauthenticated,
          403 => connect.Code.permissionDenied,
          400 => connect.Code.invalidArgument,
          503 => connect.Code.unavailable,
          _ => connect.Code.unknown,
        };
        throw connect.ConnectException(code, 'Avatar upload failed');
      }
      final value = jsonDecode(utf8.decode(body)) as Map<String, dynamic>;
      final url = value['url'] as String;
      final uri = Uri.parse(url);
      if (!['https', 'http'].contains(uri.scheme) ||
          uri.host.isEmpty ||
          uri.userInfo.isNotEmpty) {
        throw StateError('Invalid avatar URL');
      }
      return url;
    } catch (failure) {
      request.abort();
      rethrow;
    }
  }

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
