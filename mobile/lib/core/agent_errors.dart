import 'package:connectrpc/connect.dart';

/// Safe actionable explanations for the production validators. Never echo
/// arbitrary server text (which can contain instructions, metadata or keys).
String? describeAgentFailure(Object failure) {
  if (failure is! ConnectException) return null;
  if (failure.code == Code.aborted) {
    return 'This Agent changed since it was read. Reload and compare before saving; nothing is automatically retried.';
  }
  if (failure.code == Code.alreadyExists) {
    return 'This Agent ID or runtime name is already reserved in this Workspace. Choose another ID or name.';
  }
  if (failure.code == Code.internal) {
    // The server rebuilds the runtime after a save and rolls the save back when
    // a reference does not resolve; it reports that as an internal error.
    final message = failure.message;
    if (message.contains('unknown mcp_server_id')) {
      return 'Nothing was saved: an MCP server this Agent references no longer exists in this Workspace. Reload the choices and remove or replace it.';
    }
    if (message.contains('unknown remote_agent_id')) {
      return 'Nothing was saved: a Remote Agent this Agent references no longer exists in this Workspace. Reload the choices and remove or replace it.';
    }
    return null;
  }
  if (failure.code == Code.invalidArgument) {
    final message = failure.message;
    const guardErrors = {
      'config.context_guard.max_tokens must not be negative':
          'Agent Context Override must not be negative.',
      'config.context_guard.max_turns must not be negative':
          'Sliding Window entries must not be negative.',
      'config.context_guard.max_turns is not supported with the threshold':
          'Threshold uses the input-context override, not the Sliding Window entry count.',
      'config.context_guard.max_tokens is not supported with the sliding window':
          'Sliding Window uses the entry limit, not an input-context override.',
      'config.context_guard.strategy is required':
          'Choose Threshold or Sliding Window, or turn ContextGuard Off.',
      'config.context_guard is not supported':
          'ContextGuard is only supported on LLM Agents. Configure it on the independent child LLM.',
    };
    for (final entry in guardErrors.entries) {
      if (message.contains(entry.key)) {
        return 'Server validation: ${entry.value}';
      }
    }
    if (message.contains('config.context_guard.strategy')) {
      return 'Server validation: unsupported ContextGuard strategy. The stored value is retained; use a compatible server/app before editing this strategy.';
    }
    const workflowErrors = {
      'at least one node': 'Add at least one node to the workflow.',
      'reserved for the entry sentinel': 'A node uses the reserved name START.',
      'duplicate node name': 'Two nodes share a name.',
      'requires an agent_id reference': 'An Agent node has no Agent selected.',
      'requires a question': 'A Human input node has no question.',
      'references unknown node':
          'An edge refers to a node that does not exist.',
      'mutually exclusive':
          'An edge has both a route label and the default flag.',
      'match the same input': 'Two routes from one node match the same text.',
      'must not target JOIN': 'A routed or default edge cannot target a Join.',
      'entry edge': 'Connect START to the first node.',
      'requires a default outgoing edge':
          'A Router needs a default outgoing edge.',
      'parallel_worker is only supported':
          'Parallel worker applies only to Agent nodes.',
      'supported on only HUMAN_INPUT':
          'A form is supported only on Human input nodes.',
      'already claimed by agent':
          'An Agent referenced here already has another parent.',
    };
    for (final entry in workflowErrors.entries) {
      if (message.contains(entry.key)) {
        return 'Server validation: ${entry.value}';
      }
    }
    if (message.contains('form field') ||
        message.contains('a form ') ||
        message.contains('single-choice') ||
        message.contains('form title')) {
      return 'Server validation: a Human input form was rejected. Check each field\'s name, label, type and options.';
    }
    if (message.contains('workflow node') &&
        message.contains('not found in workspace')) {
      return 'Server validation: a workflow node references an Agent that is not available in this Workspace.';
    }
    return 'The server rejected the Agent configuration. Check the ID, type and child references (existing IDs, no cycles, only one parent).';
  }
  return null;
}
