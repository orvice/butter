import 'dart:async';
import 'dart:convert';
import 'dart:math';
import 'dart:io';
import 'dart:typed_data';

import 'package:connectrpc/connect.dart' as connect;
import 'package:connectrpc/protobuf.dart';
import 'package:connectrpc/protocol/connect.dart' as protocol;
import 'package:fixnum/fixnum.dart';
import 'package:protobuf/well_known_types/google/protobuf/duration.pb.dart'
    as pbd;

import '../gen/agents/v1/agent.pb.dart';
import '../gen/agents/v1/agent_service.connect.client.dart';
import '../gen/agents/v1/agent_service.pb.dart';
import '../gen/agents/v1/api_token.connect.client.dart';
import '../gen/agents/v1/api_token.pb.dart';
import '../gen/agents/v1/auth.connect.client.dart';
import '../gen/agents/v1/auth.pb.dart';
import '../gen/agents/v1/butterbox.connect.client.dart';
import '../gen/agents/v1/butterbox.pb.dart';
import '../gen/agents/v1/dashboard.connect.client.dart';
import '../gen/agents/v1/dashboard.pb.dart';
import '../gen/agents/v1/repobinding.connect.client.dart';
import '../gen/agents/v1/repobinding.pb.dart';
import '../gen/agents/v1/skill.connect.client.dart';
import '../gen/agents/v1/skill.pb.dart';
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
    _repoBindings = WorkspaceRepoBindingServiceClient(transport);
    _skills = SkillServiceClient(transport);
    _modelProviders = ModelProviderServiceClient(transport);
    _sessions = SessionServiceClient(transport);
    _boxes = ButterBoxServiceClient(transport);
    _mcp = MCPServerServiceClient(transport);
    _globalMcp = GlobalMCPServerServiceClient(transport);
    _remoteAgents = RemoteAgentServiceClient(transport);
    _daemons = DaemonServiceClient(transport);
    _apiTokens = APITokenServiceClient(transport);
  }

  final ServerAddress address;
  final HttpClient _client = HttpClient();
  late final AuthServiceClient _auth;
  late final WorkspaceServiceClient _workspaces;
  late final AgentServiceClient _agents;
  late final WorkspaceRepoBindingServiceClient _repoBindings;
  late final SkillServiceClient _skills;
  late final ModelProviderServiceClient _modelProviders;
  late final SessionServiceClient _sessions;
  late final ButterBoxServiceClient _boxes;
  late final MCPServerServiceClient _mcp;
  late final GlobalMCPServerServiceClient _globalMcp;
  late final RemoteAgentServiceClient _remoteAgents;
  late final DaemonServiceClient _daemons;
  late final APITokenServiceClient _apiTokens;

  connect.AbortSignal get _deadline =>
      connect.TimeoutSignal(const Duration(seconds: 20));

  /// The Cursor model catalog is fetched by the box through a transient
  /// bridge process; the server allows it 75 seconds, so the client waits a
  /// little longer rather than abandoning a call the server still honors.
  static const cursorCatalogTimeout = Duration(seconds: 90);
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
  Future<Agent> agent(String token, String workspace, String id) async {
    final reply = await _agents.getAgent(
      GetAgentRequest(agentId: id),
      headers: _headers(token, workspace),
      signal: _deadline,
    );
    if (!reply.hasAgent()) {
      throw connect.ConnectException(connect.Code.notFound, 'Agent missing.');
    }
    return reply.agent;
  }

  Future<AgentRuntimeStatus?> agentStatus(
    String token,
    String workspace,
    String id,
  ) async {
    final reply = await _agents.getAgentRuntimeStatus(
      GetAgentRuntimeStatusRequest(agentId: id),
      headers: _headers(token, workspace),
      signal: _deadline,
    );
    return reply.hasStatus() ? reply.status : null;
  }

  Future<WorkspaceRepoBinding?> repositoryBinding(
    String token,
    String workspace,
  ) async {
    final reply = await _repoBindings.getWorkspaceRepoBinding(
      GetWorkspaceRepoBindingRequest(),
      headers: _headers(token, workspace),
      signal: _deadline,
    );
    return reply.hasBinding() ? reply.binding : null;
  }

  /// Read-only catalog for Agent associations, not standalone Skill management.
  Future<List<String>> skillNames(String token, String workspace) async => [
    for (final skill in (await _skills.listSkills(
      ListSkillsRequest(),
      headers: _headers(token, workspace),
      signal: _deadline,
    )).skills)
      skill.name,
  ];

  Future<Agent> saveAgent(
    String token,
    String workspace,
    Agent agent, {
    required bool create,
    bool contentGitOwned = false,
    String baseCommitSha = '',
    Agent? previous,
  }) async {
    final headers = _headers(token, workspace);
    final Agent? saved;
    if (create) {
      final reply = await _agents.createAgent(
        CreateAgentRequest(
          agent: agent,
          initialContent: AgentContentInput(
            description: agent.description,
            prompt: agent.config.instruction,
            globalPrompt: agent.config.globalInstruction,
          ),
        ),
        headers: headers,
        signal: _deadline,
      );
      saved = reply.hasAgent() ? reply.agent : null;
    } else if (contentGitOwned) {
      if (previous == null || previous.agentId != agent.agentId) {
        throw const FormatException('Reload this Agent before editing.');
      }
      final fields = [
        ('description.md', previous.description, agent.description),
        ('prompt.md', previous.config.instruction, agent.config.instruction),
        (
          'global-prompt.md',
          previous.config.globalInstruction,
          agent.config.globalInstruction,
        ),
      ];
      final reply = await _agents.updateAgentConfiguration(
        UpdateAgentConfigurationRequest(
          agentPatch: agent,
          expectedAgentVersion: previous.version,
          baseCommitSha: baseCommitSha,
          contentChanges: [
            for (final field in fields)
              if (field.$2 != field.$3)
                ContentFileAction(
                  path: 'agents/${agent.agentId}/${field.$1}',
                  content: field.$3,
                  operation: field.$3.isEmpty
                      ? ContentFileOperation.CONTENT_FILE_OPERATION_DELETE
                      : ContentFileOperation.CONTENT_FILE_OPERATION_PUT,
                ),
          ],
        ),
        headers: headers,
        signal: _deadline,
      );
      if (reply.validationErrors.isNotEmpty) {
        throw const FormatException(
          'Agent Content failed server validation. Review the prompt and references before saving.',
        );
      }
      saved = reply.hasAgent() ? reply.agent : null;
    } else {
      final reply = await _agents.updateAgent(
        UpdateAgentRequest(agent: agent),
        headers: headers,
        signal: _deadline,
      );
      saved = reply.hasAgent() ? reply.agent : null;
    }
    if (saved == null) {
      throw connect.ConnectException(
        connect.Code.unknown,
        'Agent reply missing.',
      );
    }
    if (!create &&
        !contentGitOwned &&
        (saved.description != agent.description ||
            saved.config.instruction != agent.config.instruction ||
            saved.config.globalInstruction != agent.config.globalInstruction)) {
      // Ownership may have changed between discovery and the write. Some
      // operational fields may already have saved; never call that success.
      throw connect.ConnectException(
        connect.Code.unknown,
        'Agent Content was not confirmed.',
      );
    }
    return saved;
  }

  Future<List<ModelProvider>> modelProviders(
    String token,
    String workspace,
  ) async => (await _modelProviders.listModelProviders(
    ListModelProvidersRequest(),
    headers: _headers(token, workspace),
    signal: _deadline,
  )).modelProviders;
  Future<ModelProvider> modelProvider(
    String token,
    String workspace,
    String name,
  ) async {
    final reply = await _modelProviders.getModelProvider(
      GetModelProviderRequest(name: name),
      headers: _headers(token, workspace),
      signal: _deadline,
    );
    if (!reply.hasModelProvider()) {
      throw connect.ConnectException(
        connect.Code.notFound,
        'Model provider missing.',
      );
    }
    return reply.modelProvider;
  }

  /// Create and update are different calls with different failure modes
  /// (already exists vs. not found), so the caller states which one it means.
  Future<ModelProvider> saveModelProvider(
    String token,
    String workspace,
    ModelProvider provider, {
    required bool create,
  }) async {
    final headers = _headers(token, workspace);
    final ModelProvider? saved;
    if (create) {
      final reply = await _modelProviders.createModelProvider(
        CreateModelProviderRequest(modelProvider: provider),
        headers: headers,
        signal: _deadline,
      );
      saved = reply.hasModelProvider() ? reply.modelProvider : null;
    } else {
      final reply = await _modelProviders.updateModelProvider(
        UpdateModelProviderRequest(modelProvider: provider),
        headers: headers,
        signal: _deadline,
      );
      saved = reply.hasModelProvider() ? reply.modelProvider : null;
    }
    if (saved == null) {
      throw connect.ConnectException(
        connect.Code.unknown,
        'Model provider reply missing.',
      );
    }
    return saved;
  }

  Future<void> deleteModelProvider(
    String token,
    String workspace,
    String name,
  ) => _modelProviders.deleteModelProvider(
    DeleteModelProviderRequest(name: name),
    headers: _headers(token, workspace),
    signal: _deadline,
  );

  // ---- ButterBox (ButterBoxService). Every call carries the selected
  // Workspace. The access token is write-only: no reply carries it, only
  // `credential_set`. ----

  Future<List<ButterBox>> butterBoxes(String token, String workspace) async =>
      (await _boxes.listButterBoxes(
        ListButterBoxesRequest(),
        headers: _headers(token, workspace),
        signal: _deadline,
      )).butterBoxes;

  Future<ButterBox> butterBox(String token, String workspace, String id) async {
    final reply = await _boxes.getButterBox(
      GetButterBoxRequest(id: id),
      headers: _headers(token, workspace),
      signal: _deadline,
    );
    if (!reply.hasButterBox()) {
      throw connect.ConnectException(connect.Code.notFound, 'Box missing.');
    }
    return reply.butterBox;
  }

  /// One create. [accessToken] empty stores no token.
  Future<ButterBox> createButterBox(
    String token,
    String workspace, {
    required String name,
    required String baseUrl,
    required bool enabled,
    String accessToken = '',
  }) async {
    final reply = await _boxes.createButterBox(
      CreateButterBoxRequest(
        name: name,
        baseUrl: baseUrl,
        enabled: enabled,
        token: accessToken,
      ),
      headers: _headers(token, workspace),
      signal: _deadline,
    );
    if (!reply.hasButterBox()) {
      throw connect.ConnectException(
        connect.Code.unknown,
        'Box reply missing.',
      );
    }
    return reply.butterBox;
  }

  /// Updates name, URL and enabled only; the token has its own call.
  Future<ButterBox> updateButterBox(
    String token,
    String workspace, {
    required String id,
    required String name,
    required String baseUrl,
    required bool enabled,
  }) async {
    final reply = await _boxes.updateButterBox(
      UpdateButterBoxRequest(
        id: id,
        name: name,
        baseUrl: baseUrl,
        enabled: enabled,
      ),
      headers: _headers(token, workspace),
      signal: _deadline,
    );
    if (!reply.hasButterBox()) {
      throw connect.ConnectException(
        connect.Code.unknown,
        'Box reply missing.',
      );
    }
    return reply.butterBox;
  }

  Future<void> deleteButterBox(String token, String workspace, String id) =>
      _boxes.deleteButterBox(
        DeleteButterBoxRequest(id: id),
        headers: _headers(token, workspace),
        signal: _deadline,
      );

  /// Sets or rotates the box access token; an empty [accessToken] clears it.
  Future<ButterBox> setButterBoxToken(
    String token,
    String workspace,
    String id,
    String accessToken,
  ) async {
    final reply = await _boxes.setButterBoxToken(
      SetButterBoxTokenRequest(id: id, token: accessToken),
      headers: _headers(token, workspace),
      signal: _deadline,
    );
    if (!reply.hasButterBox()) {
      throw connect.ConnectException(
        connect.Code.unknown,
        'Box reply missing.',
      );
    }
    return reply.butterBox;
  }

  /// Probes the box through the server. Unreachable is a normal answer.
  Future<GetButterBoxStatusResponse> butterBoxStatus(
    String token,
    String workspace,
    String id,
  ) => _boxes.getButterBoxStatus(
    GetButterBoxStatusRequest(id: id),
    headers: _headers(token, workspace),
    signal: _deadline,
  );

  Future<List<ButterBoxModel>> butterBoxModels(
    String token,
    String workspace,
    String id,
  ) async => (await _boxes.listButterBoxModels(
    ListButterBoxModelsRequest(id: id),
    headers: _headers(token, workspace),
    signal: _deadline,
  )).models;

  Future<List<ButterBoxCursorModel>> butterBoxCursorModels(
    String token,
    String workspace,
    String id,
  ) async => (await _boxes.listButterBoxCursorModels(
    ListButterBoxCursorModelsRequest(id: id),
    headers: _headers(token, workspace),
    signal: connect.TimeoutSignal(cursorCatalogTimeout),
  )).models;

  // ---- MCP servers (MCPServerService), global presets
  // (GlobalMCPServerService: list and install only) and Remote Agents
  // (RemoteAgentService). Every call carries the selected Workspace. ----

  /// A probe makes the server connect to the MCP server (its own timeout is
  /// per server, and ListMCPTools probes servers one after another), so the
  /// client waits longer than for a plain read rather than abandoning a call
  /// the server still honors.
  static const mcpProbeTimeout = Duration(seconds: 60);

  Future<List<MCPServer>> mcpServers(String token, String workspace) async =>
      (await _mcp.listMCPServers(
        ListMCPServersRequest(),
        headers: _headers(token, workspace),
        signal: _deadline,
      )).mcpServers;

  MCPServer _mcpReply(bool has, MCPServer Function() read) {
    if (!has) {
      throw connect.ConnectException(
        connect.Code.unknown,
        'MCP server reply missing.',
      );
    }
    return read();
  }

  Future<MCPServer> mcpServer(String token, String workspace, String id) async {
    final reply = await _mcp.getMCPServer(
      GetMCPServerRequest(id: id),
      headers: _headers(token, workspace),
      signal: _deadline,
    );
    if (!reply.hasMcpServer()) {
      throw connect.ConnectException(
        connect.Code.notFound,
        'MCP server missing.',
      );
    }
    return reply.mcpServer;
  }

  /// One create. [server] is the complete message to store.
  Future<MCPServer> createMcpServer(
    String token,
    String workspace,
    MCPServer server,
  ) async {
    final reply = await _mcp.createMCPServer(
      CreateMCPServerRequest(mcpServer: server),
      headers: _headers(token, workspace),
      signal: _deadline,
    );
    return _mcpReply(reply.hasMcpServer(), () => reply.mcpServer);
  }

  /// One whole-object update: what is sent is what is stored.
  Future<MCPServer> updateMcpServer(
    String token,
    String workspace,
    MCPServer server,
  ) async {
    final reply = await _mcp.updateMCPServer(
      UpdateMCPServerRequest(mcpServer: server),
      headers: _headers(token, workspace),
      signal: _deadline,
    );
    return _mcpReply(reply.hasMcpServer(), () => reply.mcpServer);
  }

  /// A whole-object update that also keeps the stored credentials named in
  /// [keepHeaderKeys] / [keepClientSecret], which reads of an installed preset
  /// hide. It is a method of its own, not fields of [updateMcpServer]: a server
  /// that predates it answers Unimplemented before storing anything, whereas
  /// an older update would ignore the keep fields and delete the credentials
  /// left out of the message. The answer hides credentials too.
  Future<MCPServer> updateMcpServerKeeping(
    String token,
    String workspace,
    MCPServer server, {
    List<String> keepHeaderKeys = const [],
    bool keepClientSecret = false,
  }) async {
    final reply = await _mcp.updateMCPServerKeepingCredentials(
      UpdateMCPServerKeepingCredentialsRequest(
        mcpServer: server,
        keepHeaderKeys: keepHeaderKeys,
        keepOauthClientSecret: keepClientSecret ? true : null,
      ),
      headers: _headers(token, workspace),
      signal: _deadline,
    );
    return _mcpReply(reply.hasMcpServer(), () => reply.mcpServer);
  }

  Future<void> deleteMcpServer(String token, String workspace, String id) =>
      _mcp.deleteMCPServer(
        DeleteMCPServerRequest(id: id),
        headers: _headers(token, workspace),
        signal: _deadline,
      );

  /// Connects to the server through Butter. A server that cannot be reached
  /// is an answer (`DISCONNECTED`), not an RPC failure.
  Future<MCPServerStatus> mcpServerStatus(
    String token,
    String workspace,
    String id,
  ) async {
    final reply = await _mcp.getMCPServerStatus(
      GetMCPServerStatusRequest(id: id),
      headers: _headers(token, workspace),
      signal: connect.TimeoutSignal(mcpProbeTimeout),
    );
    if (!reply.hasStatus()) {
      throw connect.ConnectException(
        connect.Code.unknown,
        'MCP status reply missing.',
      );
    }
    return reply.status;
  }

  /// Lists the tools of one server; failures come back per server.
  Future<ListMCPToolsResponse> mcpTools(
    String token,
    String workspace,
    String id,
  ) => _mcp.listMCPTools(
    ListMCPToolsRequest(serverId: id),
    headers: _headers(token, workspace),
    signal: connect.TimeoutSignal(mcpProbeTimeout),
  );

  Future<StartMCPServerOAuthResponse> startMcpOAuth(
    String token,
    String workspace,
    String id,
  ) => _mcp.startMCPServerOAuth(
    // No return URL: the server's callback finishes the exchange itself and
    // the App reads the resulting state, so no address is invented for it.
    StartMCPServerOAuthRequest(serverId: id),
    headers: _headers(token, workspace),
    signal: _deadline,
  );

  Future<MCPOAuthConnectionStatus> mcpOAuthStatus(
    String token,
    String workspace,
    String id,
  ) async {
    final reply = await _mcp.getMCPServerOAuthStatus(
      GetMCPServerOAuthStatusRequest(serverId: id),
      headers: _headers(token, workspace),
      signal: _deadline,
    );
    if (!reply.hasStatus()) {
      throw connect.ConnectException(
        connect.Code.unknown,
        'MCP OAuth status reply missing.',
      );
    }
    return reply.status;
  }

  Future<void> disconnectMcpOAuth(String token, String workspace, String id) =>
      _mcp.disconnectMCPServerOAuth(
        DisconnectMCPServerOAuthRequest(serverId: id),
        headers: _headers(token, workspace),
        signal: _deadline,
      );

  Future<List<MCPServer>> globalMcpPresets(
    String token,
    String workspace,
  ) async => (await _globalMcp.listGlobalMCPServers(
    ListGlobalMCPServersRequest(),
    headers: _headers(token, workspace),
    signal: _deadline,
  )).mcpServers;

  /// Installs a preset into [workspace]: create-or-overwrite of the
  /// Workspace's server with the preset's ID. The reply is redacted.
  Future<MCPServer> installGlobalMcpPreset(
    String token,
    String workspace,
    String presetId,
  ) async {
    final reply = await _globalMcp.installGlobalMCPServer(
      InstallGlobalMCPServerRequest(id: presetId),
      headers: _headers(token, workspace),
      signal: _deadline,
    );
    return _mcpReply(reply.hasMcpServer(), () => reply.mcpServer);
  }

  Future<List<RemoteAgent>> remoteAgents(
    String token,
    String workspace,
  ) async => (await _remoteAgents.listRemoteAgents(
    ListRemoteAgentsRequest(),
    headers: _headers(token, workspace),
    signal: _deadline,
  )).remoteAgents;

  RemoteAgent _remoteReply(bool has, RemoteAgent Function() read) {
    if (!has) {
      throw connect.ConnectException(
        connect.Code.unknown,
        'Remote Agent reply missing.',
      );
    }
    return read();
  }

  Future<RemoteAgent> remoteAgent(
    String token,
    String workspace,
    String id,
  ) async {
    final reply = await _remoteAgents.getRemoteAgent(
      GetRemoteAgentRequest(id: id),
      headers: _headers(token, workspace),
      signal: _deadline,
    );
    if (!reply.hasRemoteAgent()) {
      throw connect.ConnectException(
        connect.Code.notFound,
        'Remote Agent missing.',
      );
    }
    return reply.remoteAgent;
  }

  Future<RemoteAgent> createRemoteAgent(
    String token,
    String workspace,
    RemoteAgent agent,
  ) async {
    final reply = await _remoteAgents.createRemoteAgent(
      CreateRemoteAgentRequest(remoteAgent: agent),
      headers: _headers(token, workspace),
      signal: _deadline,
    );
    return _remoteReply(reply.hasRemoteAgent(), () => reply.remoteAgent);
  }

  Future<RemoteAgent> updateRemoteAgent(
    String token,
    String workspace,
    RemoteAgent agent,
  ) async {
    final reply = await _remoteAgents.updateRemoteAgent(
      UpdateRemoteAgentRequest(remoteAgent: agent),
      headers: _headers(token, workspace),
      signal: _deadline,
    );
    return _remoteReply(reply.hasRemoteAgent(), () => reply.remoteAgent);
  }

  Future<void> deleteRemoteAgent(String token, String workspace, String id) =>
      _remoteAgents.deleteRemoteAgent(
        DeleteRemoteAgentRequest(id: id),
        headers: _headers(token, workspace),
        signal: _deadline,
      );

  /// Probes the endpoint through the server. Unreachable is a normal answer.
  Future<RemoteAgentStatus> remoteAgentStatus(
    String token,
    String workspace,
    String id,
  ) async {
    final reply = await _remoteAgents.getRemoteAgentStatus(
      GetRemoteAgentStatusRequest(id: id),
      headers: _headers(token, workspace),
      signal: _deadline,
    );
    if (!reply.hasStatus()) {
      throw connect.ConnectException(
        connect.Code.unknown,
        'Remote Agent status reply missing.',
      );
    }
    return reply.status;
  }

  // ---- Daemon execution (DaemonService). Every call carries the selected
  // Workspace. A worker token is minted only by [createDaemonRuntimeToken];
  // this client never creates an ordinary API token. ----

  Future<List<DaemonRuntime>> daemonRuntimes(
    String token,
    String workspace,
  ) async => (await _daemons.listDaemonRuntimes(
    ListDaemonRuntimesRequest(),
    headers: _headers(token, workspace),
    signal: _deadline,
  )).runtimes;

  Future<DaemonRuntime> daemonRuntime(
    String token,
    String workspace,
    String id,
  ) async {
    final reply = await _daemons.getDaemonRuntime(
      GetDaemonRuntimeRequest(id: id),
      headers: _headers(token, workspace),
      signal: _deadline,
    );
    if (!reply.hasRuntime()) {
      throw connect.ConnectException(
        connect.Code.notFound,
        'Daemon runtime missing.',
      );
    }
    return reply.runtime;
  }

  DaemonRuntime _runtimeReply(bool has, DaemonRuntime Function() read) {
    if (!has) {
      throw connect.ConnectException(
        connect.Code.unknown,
        'Daemon runtime reply missing.',
      );
    }
    return read();
  }

  Future<DaemonRuntime> createDaemonRuntime(
    String token,
    String workspace,
    DaemonRuntime runtime,
  ) async {
    final reply = await _daemons.createDaemonRuntime(
      CreateDaemonRuntimeRequest(runtime: runtime),
      headers: _headers(token, workspace),
      signal: _deadline,
    );
    return _runtimeReply(reply.hasRuntime(), () => reply.runtime);
  }

  /// One whole-object update: what is sent is what is stored.
  Future<DaemonRuntime> updateDaemonRuntime(
    String token,
    String workspace,
    DaemonRuntime runtime,
  ) async {
    final reply = await _daemons.updateDaemonRuntime(
      UpdateDaemonRuntimeRequest(runtime: runtime),
      headers: _headers(token, workspace),
      signal: _deadline,
    );
    return _runtimeReply(reply.hasRuntime(), () => reply.runtime);
  }

  Future<void> deleteDaemonRuntime(String token, String workspace, String id) =>
      _daemons.deleteDaemonRuntime(
        DeleteDaemonRuntimeRequest(id: id),
        headers: _headers(token, workspace),
        signal: _deadline,
      );

  /// Mints one worker token bound to [runtimeId]. Not idempotent: every call
  /// stores another token, so a caller must never repeat it on its own. The
  /// reply carries the secret exactly once.
  Future<CreateDaemonRuntimeTokenResponse> createDaemonRuntimeToken(
    String token,
    String workspace,
    String runtimeId,
    String name, {
    Duration? ttl,
  }) async {
    if (ttl != null && ttl <= Duration.zero) {
      throw ArgumentError.value(ttl, 'ttl', 'must be positive, or null');
    }
    final reply = await _daemons.createDaemonRuntimeToken(
      CreateDaemonRuntimeTokenRequest(
        daemonRuntimeId: runtimeId,
        name: name,
        // Absent means the token does not expire. The server reads a zero or
        // negative duration the same way, so one is never sent for a token
        // that was meant to expire.
        ttl: ttl == null
            ? null
            : pbd.Duration(
                seconds: Int64(
                  ttl.inMicroseconds ~/ Duration.microsecondsPerSecond,
                ),
                nanos:
                    (ttl.inMicroseconds % Duration.microsecondsPerSecond) *
                    1000,
              ),
      ),
      headers: _headers(token, workspace),
      signal: _deadline,
    );
    if (reply.secret.isEmpty || !reply.hasToken()) {
      throw connect.ConnectException(
        connect.Code.unknown,
        'Worker token reply incomplete.',
      );
    }
    return reply;
  }

  /// The daemons connected right now. A runtime that is not listed is offline.
  Future<List<DaemonStatus>> daemons(String token, String workspace) async =>
      (await _daemons.listDaemons(
        ListDaemonsRequest(),
        headers: _headers(token, workspace),
        signal: _deadline,
      )).daemons;

  /// Tasks in flight only; a finished task is not listed anywhere.
  Future<List<DaemonTaskInFlight>> daemonTasks(
    String token,
    String workspace,
  ) async => (await _daemons.listDaemonTasks(
    ListDaemonTasksRequest(),
    headers: _headers(token, workspace),
    signal: _deadline,
  )).tasks;

  /// Asks the server to deliver a cancel to the worker running [taskId]. The
  /// reply means the cancel was handed over, never that the task stopped.
  /// Returns the runtime that received it.
  Future<String> cancelDaemonTask(
    String token,
    String workspace,
    String taskId,
    String runtimeId,
  ) async => (await _daemons.cancelDaemonTask(
    CancelDaemonTaskRequest(taskId: taskId, daemonRuntimeId: runtimeId),
    headers: _headers(token, workspace),
    signal: _deadline,
  )).daemonRuntimeId;

  /// Process metrics of the Butter server, shared by every Workspace.
  Future<BridgeDiagnostics> bridgeDiagnostics(
    String token,
    String workspace,
  ) async {
    final reply = await _daemons.getBridgeDiagnostics(
      GetBridgeDiagnosticsRequest(),
      headers: _headers(token, workspace),
      signal: _deadline,
    );
    if (!reply.hasDiagnostics()) {
      throw connect.ConnectException(
        connect.Code.unknown,
        'Diagnostics reply missing.',
      );
    }
    return reply.diagnostics;
  }

  /// Every token of the Workspace, of every kind. Callers must keep only the
  /// worker tokens they manage and never show another kind.
  Future<List<APIToken>> apiTokens(String token, String workspace) async =>
      (await _apiTokens.listAPITokens(
        ListAPITokensRequest(),
        headers: _headers(token, workspace),
        signal: _deadline,
      )).tokens;

  Future<void> revokeApiToken(String token, String workspace, String id) =>
      _apiTokens.revokeAPIToken(
        RevokeAPITokenRequest(id: id),
        headers: _headers(token, workspace),
        signal: _deadline,
      );

  /// One page of the caller's AG-UI thread sessions in [workspace]. The server
  /// scopes the list to that Workspace; which sessions are bound threads is
  /// decided by the caller from the session state.
  Future<ListSessionsResponse> threadSessions(
    String token,
    String workspace, {
    required String userId,
    required int pageSize,
    String pageToken = '',
  }) => _sessions.listSessions(
    ListSessionsRequest(
      appName: 'agui',
      userId: userId,
      workspaceScoped: true,
      pageSize: pageSize,
      pageToken: pageToken,
    ),
    headers: _headers(token, workspace),
    signal: _deadline,
  );

  /// Sets a session's title. Never repeated on its own.
  Future<SessionInfo> renameSession(
    String token,
    String workspace, {
    required String appName,
    required String userId,
    required String sessionId,
    required String title,
  }) async {
    final reply = await _sessions.updateSessionTitle(
      UpdateSessionTitleRequest(
        appName: appName,
        userId: userId,
        sessionId: sessionId,
        title: title,
      ),
      headers: _headers(token, workspace),
      signal: _deadline,
    );
    if (!reply.hasSession()) {
      throw connect.ConnectException(
        connect.Code.unknown,
        'Session reply missing.',
      );
    }
    return reply.session;
  }

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
