import 'package:connectrpc/connect.dart';

import '../gen/agents/v1/agent_operation.pbenum.dart';

/// Safe explanations for Agent lifecycle and Git content RPC failures.
String? describeAgentLifecycleFailure(Object failure) {
  if (failure is! ConnectException) return null;
  if (failure.code == Code.aborted) {
    return 'This Agent changed since it was read. Reload and compare before saving; nothing is automatically retried.';
  }
  if (failure.code == Code.failedPrecondition) {
    final message = failure.message;
    if (message.contains('agent is deleted')) {
      return 'This Agent is tombstoned. Restore it before changing its configuration.';
    }
    if (message.contains('agent is not deleted')) {
      return 'This Agent is not tombstoned; restore does not apply.';
    }
    if (message.contains('lifecycle operations are not available')) {
      return 'Agent lifecycle operations need a Git repository binding in this Workspace.';
    }
    if (message.contains('referenced by')) {
      return 'This Agent is still referenced elsewhere and cannot be removed yet. Update those references first.';
    }
    if (message.contains('orphan')) {
      return 'Other Agents still reference this one. Remove or reassign those references before deleting.';
    }
    if (message.contains('claimed')) {
      return 'Git content for this Agent ID is still claimed by an active Agent. Tombstone or purge only after nothing active uses this path.';
    }
  }
  if (failure.code == Code.permissionDenied) {
    return 'You do not have permission for this Agent lifecycle action.';
  }
  if (failure.code == Code.invalidArgument &&
      failure.message.contains('validation')) {
    return 'The server rejected the Agent Content change. Review the fields and try again after reloading.';
  }
  return null;
}

String operationTypeLabel(AgentOperationType type) => switch (type) {
  AgentOperationType.AGENT_OPERATION_TYPE_CREATE => 'Create',
  AgentOperationType.AGENT_OPERATION_TYPE_UPDATE_CONFIGURATION =>
    'Update configuration',
  AgentOperationType.AGENT_OPERATION_TYPE_DELETE => 'Delete (tombstone)',
  AgentOperationType.AGENT_OPERATION_TYPE_RESTORE => 'Restore',
  _ => 'Unknown operation',
};

String operationStatusLabel(AgentOperationStatus status) => switch (status) {
  AgentOperationStatus.AGENT_OPERATION_STATUS_PENDING => 'Pending',
  AgentOperationStatus.AGENT_OPERATION_STATUS_RUNNING => 'Running',
  AgentOperationStatus.AGENT_OPERATION_STATUS_SUCCEEDED => 'Succeeded',
  AgentOperationStatus.AGENT_OPERATION_STATUS_FAILED => 'Failed',
  _ => 'Unknown',
};

String operationStepStatusLabel(AgentOperationStepStatus status) =>
    switch (status) {
      AgentOperationStepStatus.AGENT_OPERATION_STEP_STATUS_PENDING => 'Pending',
      AgentOperationStepStatus.AGENT_OPERATION_STEP_STATUS_SUCCEEDED =>
        'Succeeded',
      AgentOperationStepStatus.AGENT_OPERATION_STEP_STATUS_FAILED => 'Failed',
      AgentOperationStepStatus.AGENT_OPERATION_STEP_STATUS_SKIPPED => 'Skipped',
      _ => 'Unknown',
    };

String operationStepLabel(AgentOperationStepKind kind) => switch (kind) {
  AgentOperationStepKind.AGENT_OPERATION_STEP_KIND_DB_PROVISION =>
    'Provision database record',
  AgentOperationStepKind.AGENT_OPERATION_STEP_KIND_CONTENT_COMMIT =>
    'Commit agent content',
  AgentOperationStepKind.AGENT_OPERATION_STEP_KIND_SYNC_PUBLISH =>
    'Sync and publish',
  AgentOperationStepKind.AGENT_OPERATION_STEP_KIND_ACTIVATE => 'Activate agent',
  AgentOperationStepKind.AGENT_OPERATION_STEP_KIND_TOMBSTONE =>
    'Tombstone agent',
  AgentOperationStepKind.AGENT_OPERATION_STEP_KIND_RESTORE_DB =>
    'Restore database record',
  AgentOperationStepKind.AGENT_OPERATION_STEP_KIND_DB_PATCH =>
    'Apply configuration patch',
  _ => 'Unknown step',
};
