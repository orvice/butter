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
    return 'The server rejected the Agent configuration. Check the ID, type and child references (existing IDs, no cycles, only one parent).';
  }
  return null;
}
