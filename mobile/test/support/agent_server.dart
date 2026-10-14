import 'dart:async';
import 'dart:convert';
import 'dart:io';

import 'package:butter_mobile/gen/agents/v1/agent.pb.dart';
import 'package:butter_mobile/gen/agents/v1/agent_service.pb.dart';
import 'package:butter_mobile/gen/agents/v1/repobinding.pb.dart';

import 'server.dart';

/// External HTTP control for late/ambiguous replies; production semantics are
/// checked separately by agent_contract_test against the real Go handlers.
class AgentServer extends TestServer {
  final agents = <String, Map<String, Agent>>{'alpha': {}, 'beta': {}};
  final calls = <({String method, String? workspace, List<int> body})>[];
  Future<void> Function()? beforeGet;
  Future<void> Function()? beforeWrite;
  (int, String, String)? refusal;
  bool loseWriteReply = false;
  bool bound = false;

  AgentServer() {
    authorizedReply = (r, bytes) async {
      final method = r.uri.path.split('/').last;
      final ws = r.headers.value('x-workspace-id');
      if (method == 'GetWorkspaceRepoBinding') {
        write(
          r,
          GetWorkspaceRepoBindingResponse(
            binding: bound ? WorkspaceRepoBinding(workspaceId: ws) : null,
          ),
        );
        return true;
      }
      if (method == 'ListModelProviders') {
        write(
          r,
          ListModelProvidersResponse(
            modelProviders: [
              ModelProvider(
                name: 'provider',
                apiKey: 'never-show',
                models: [ModelConfig(name: 'provider-model', alias: 'fast')],
              ),
            ],
          ),
        );
        return true;
      }
      if (!r.uri.path.contains('AgentService')) return false;
      calls.add((method: method, workspace: ws, body: bytes));
      final bucket = agents[ws];
      if (bucket == null) {
        _fail(r, 403, 'permission_denied', 'Workspace denied');
        return true;
      }
      switch (method) {
        case 'ListAgents':
          final req = ListAgentsRequest.fromBuffer(bytes);
          final list = bucket.values.toList();
          final offset = int.tryParse(req.pageToken) ?? 0;
          write(
            r,
            ListAgentsResponse(
              agents: list.skip(offset).take(50),
              nextPageToken: list.length > offset + 50 ? '${offset + 50}' : '',
            ),
          );
        case 'GetAgent':
          await beforeGet?.call();
          final found = bucket[GetAgentRequest.fromBuffer(bytes).agentId];
          if (found == null) {
            _fail(r, 404, 'not_found', 'Agent missing');
          } else {
            write(r, GetAgentResponse(agent: found));
          }
        case 'GetAgentRuntimeStatus':
          write(
            r,
            GetAgentRuntimeStatusResponse(
              status: AgentRuntimeStatus(
                agentId: GetAgentRuntimeStatusRequest.fromBuffer(bytes).agentId,
                state: AgentRuntimeState.AGENT_RUNTIME_STATE_IDLE,
              ),
            ),
          );
        case 'CreateAgent' || 'UpdateAgent' || 'UpdateAgentConfiguration':
          await beforeWrite?.call();
          if (refusal case final failure?) {
            _fail(r, failure.$1, failure.$2, failure.$3);
            return true;
          }
          final Agent incoming;
          if (method == 'CreateAgent') {
            incoming = CreateAgentRequest.fromBuffer(bytes).agent;
          } else if (method == 'UpdateAgent') {
            incoming = UpdateAgentRequest.fromBuffer(bytes).agent;
          } else {
            incoming = UpdateAgentConfigurationRequest.fromBuffer(
              bytes,
            ).agentPatch;
          }
          final stored = incoming.deepCopy()..workspaceId = ws!;
          bucket[stored.agentId] = stored;
          if (loseWriteReply) {
            _fail(r, 503, 'unavailable', 'Reply lost');
          } else if (method == 'CreateAgent') {
            write(r, CreateAgentResponse(agent: stored));
          } else if (method == 'UpdateAgent') {
            write(r, UpdateAgentResponse(agent: stored));
          } else {
            write(r, UpdateAgentConfigurationResponse(agent: stored));
          }
        default:
          return false;
      }
      return true;
    };
  }

  void _fail(HttpRequest r, int status, String code, String message) {
    r.response.statusCode = status;
    r.response.headers.contentType = ContentType.json;
    r.response.write(jsonEncode({'code': code, 'message': message}));
  }
}
