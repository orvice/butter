import 'dart:async';
import 'dart:convert';
import 'dart:io';

import 'package:butter_mobile/gen/agents/v1/agent.pb.dart';
import 'package:butter_mobile/gen/agents/v1/agent_operation.pb.dart';
import 'package:butter_mobile/gen/agents/v1/agent_service.pb.dart';
import 'package:butter_mobile/gen/agents/v1/repobinding.pb.dart';
import 'package:butter_mobile/gen/agents/v1/skill.pb.dart';

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
  String activeCommitSha = 'active-revision';

  /// Active Revision content per Workspace and Agent, keyed by file name
  /// (`prompt.md`, `description.md`, `global-prompt.md`). Like production,
  /// reads overlay it onto the stored Agent, while a composite save answers
  /// with the stored record and never writes Git content back into it.
  final gitContent = <String, Map<String, Map<String, String>>>{
    'alpha': {},
    'beta': {},
  };

  /// Refuses `GetAgent` while set, e.g. the read-back after a composite save.
  (int, String, String)? getRefusal;

  /// Refuses `ListAgents` while set.
  (int, String, String)? agentsListRefusal;

  final operations = <AgentOperation>[];

  Agent _overlaid(String ws, Agent stored) {
    final files = gitContent[ws]?[stored.agentId];
    if (files == null) return stored;
    final agent = stored.deepCopy();
    final description = files['description.md'];
    final prompt = files['prompt.md'];
    final global = files['global-prompt.md'];
    if (description != null) agent.description = description;
    if (prompt != null) agent.ensureConfig().instruction = prompt;
    if (global != null) agent.ensureConfig().globalInstruction = global;
    return agent;
  }

  AgentServer() {
    authorizedReply = (r, bytes) async {
      final method = r.uri.path.split('/').last;
      final ws = r.headers.value('x-workspace-id');
      if (method == 'GetWorkspaceRepoBinding') {
        write(
          r,
          GetWorkspaceRepoBindingResponse(
            binding: bound
                ? WorkspaceRepoBinding(
                    workspaceId: ws,
                    activeCommitSha: activeCommitSha,
                  )
                : null,
          ),
        );
        return true;
      }
      if (method == 'ListSkills') {
        write(r, ListSkillsResponse(skills: [Skill(name: 'writing')]));
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
          if (agentsListRefusal case final failure?) {
            _fail(r, failure.$1, failure.$2, failure.$3);
            return true;
          }
          final req = ListAgentsRequest.fromBuffer(bytes);
          final list = bucket.values.toList();
          final offset = int.tryParse(req.pageToken) ?? 0;
          write(
            r,
            ListAgentsResponse(
              agents: list.skip(offset).take(50).map((a) => _overlaid(ws!, a)),
              nextPageToken: list.length > offset + 50 ? '${offset + 50}' : '',
            ),
          );
        case 'GetAgent':
          await beforeGet?.call();
          if (getRefusal case final failure?) {
            _fail(r, failure.$1, failure.$2, failure.$3);
            return true;
          }
          final found = bucket[GetAgentRequest.fromBuffer(bytes).agentId];
          if (found == null) {
            _fail(r, 404, 'not_found', 'Agent missing');
          } else {
            write(r, GetAgentResponse(agent: _overlaid(ws!, found)));
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
          if (method == 'UpdateAgentConfiguration') {
            // Production keeps the stored description/instructions: Git owns
            // them, so the patch's values for those fields are not written.
            final request = UpdateAgentConfigurationRequest.fromBuffer(bytes);
            final previous = bucket[stored.agentId];
            if (previous != null) {
              stored.description = previous.description;
              stored.ensureConfig()
                ..instruction = previous.config.instruction
                ..globalInstruction = previous.config.globalInstruction;
            }
            for (final change in request.contentChanges) {
              final parts = change.path.split('/');
              final files = gitContent[ws]!.putIfAbsent(parts[1], () => {});
              if (change.operation ==
                  ContentFileOperation.CONTENT_FILE_OPERATION_DELETE) {
                files.remove(parts.last);
              } else {
                files[parts.last] = change.content;
              }
            }
          }
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
        case 'DeleteAgent':
          await beforeWrite?.call();
          if (refusal case final failure?) {
            _fail(r, failure.$1, failure.$2, failure.$3);
            return true;
          }
          if (loseWriteReply) {
            _fail(r, 503, 'unavailable', 'Reply lost');
            return true;
          }
          final delId = DeleteAgentRequest.fromBuffer(bytes).agentId;
          final victim = bucket[delId];
          if (victim == null) {
            _fail(r, 404, 'not_found', 'Agent missing');
          } else {
            victim.lifecycleStatus =
                AgentLifecycleStatus.AGENT_LIFECYCLE_STATUS_DELETED;
            operations.add(
              AgentOperation(
                id: 'op-delete-$delId',
                agentId: delId,
                type: AgentOperationType.AGENT_OPERATION_TYPE_DELETE,
                status: AgentOperationStatus.AGENT_OPERATION_STATUS_SUCCEEDED,
              ),
            );
            write(r, DeleteAgentResponse(operation: operations.last));
          }
        case 'RestoreAgent':
          final restoreId = RestoreAgentRequest.fromBuffer(bytes).agentId;
          final tomb = bucket[restoreId];
          if (tomb == null) {
            _fail(r, 404, 'not_found', 'Agent missing');
          } else {
            tomb.lifecycleStatus =
                AgentLifecycleStatus.AGENT_LIFECYCLE_STATUS_ACTIVE;
            write(r, RestoreAgentResponse(agent: _overlaid(ws!, tomb)));
          }
        case 'ListAgentOperations':
          write(r, ListAgentOperationsResponse(operations: operations));
        case 'GetAgentOperation':
          final opId = GetAgentOperationRequest.fromBuffer(bytes).operationId;
          final op = operations.cast<AgentOperation?>().firstWhere(
            (o) => o!.id == opId,
            orElse: () => null,
          );
          if (op == null) {
            _fail(r, 404, 'not_found', 'Operation missing');
          } else {
            write(r, GetAgentOperationResponse(operation: op));
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
