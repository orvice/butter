import 'package:connectrpc/connect.dart';

import '../gen/agents/v1/automation.pb.dart';

enum AutomationOperation { load, create, update, delete, run, runs, steps }

/// Trigger types the server executes today (manual + schedule).
const automationTriggerImplemented = {
  AutomationTriggerType.AUTOMATION_TRIGGER_TYPE_MANUAL,
  AutomationTriggerType.AUTOMATION_TRIGGER_TYPE_SCHEDULE,
};

bool automationTriggerUsable(AutomationTriggerType type) =>
    automationTriggerImplemented.contains(type);

String? describeAutomationFailure(
  Object failure,
  AutomationOperation operation,
) {
  if (failure is! ConnectException) return null;
  final message = failure.message.toLowerCase();
  switch (failure.code) {
    case Code.invalidArgument:
      if (message.contains('trigger_payload_json')) {
        return 'Trigger payload must be valid JSON.';
      }
      if (message.contains('agent_id')) {
        return 'Choose an Agent that exists in this Workspace.';
      }
      if (operation == AutomationOperation.create ||
          operation == AutomationOperation.update) {
        return 'Check the trigger, conditions, steps and policy fields.';
      }
      return null;
    case Code.notFound:
      return switch (operation) {
        AutomationOperation.run =>
          'This automation no longer exists, so nothing was run. Refresh the list.',
        AutomationOperation.runs || AutomationOperation.steps =>
          'This run is no longer available. Refresh.',
        _ =>
          'This automation no longer exists, so nothing was changed. Refresh the list.',
      };
    case Code.failedPrecondition:
      if (message.contains('disabled')) {
        return 'Enable the automation before running it.';
      }
      return 'The server is not ready for this operation.';
    case Code.permissionDenied:
      return 'You do not have access to this Workspace, or your role does not allow this.';
    default:
      return null;
  }
}

String automationFailureMessage(Object failure, AutomationOperation operation) {
  final specific = describeAutomationFailure(failure, operation);
  if (specific != null) return specific;
  if (failure is ConnectException) {
    return switch (failure.code) {
      Code.permissionDenied => 'You do not have permission for this operation.',
      Code.invalidArgument => 'Check the entered values.',
      Code.notFound => 'This automation is no longer available.',
      Code.failedPrecondition => 'The server rejected this run.',
      Code.deadlineExceeded || Code.unavailable =>
        'Cannot reach the server. Check your connection and retry.',
      _ => 'The request failed. Retry explicitly.',
    };
  }
  return 'The operation failed. Check your connection, then retry.';
}

bool automationWriteOutcomeUnknown(Object failure) =>
    failure is ConnectException &&
    const [
      Code.unavailable,
      Code.deadlineExceeded,
      Code.unknown,
      Code.internal,
    ].contains(failure.code);

String automationRunStatusLabel(AutomationRunStatus s) => switch (s) {
  AutomationRunStatus.AUTOMATION_RUN_STATUS_RUNNING => 'Running',
  AutomationRunStatus.AUTOMATION_RUN_STATUS_SUCCEEDED => 'Succeeded',
  AutomationRunStatus.AUTOMATION_RUN_STATUS_FAILED => 'Failed',
  AutomationRunStatus.AUTOMATION_RUN_STATUS_SKIPPED => 'Skipped',
  AutomationRunStatus.AUTOMATION_RUN_STATUS_CANCELLED => 'Cancelled',
  AutomationRunStatus.AUTOMATION_RUN_STATUS_WAITING_INPUT =>
    'Waiting for input',
  _ => 'Unknown',
};

String automationStepRunStatusLabel(AutomationStepRunStatus s) => switch (s) {
  AutomationStepRunStatus.AUTOMATION_STEP_RUN_STATUS_RUNNING => 'Running',
  AutomationStepRunStatus.AUTOMATION_STEP_RUN_STATUS_SUCCEEDED => 'Succeeded',
  AutomationStepRunStatus.AUTOMATION_STEP_RUN_STATUS_FAILED => 'Failed',
  AutomationStepRunStatus.AUTOMATION_STEP_RUN_STATUS_SKIPPED => 'Skipped',
  AutomationStepRunStatus.AUTOMATION_STEP_RUN_STATUS_CANCELLED => 'Cancelled',
  _ => 'Unknown',
};

String automationTriggerLabel(AutomationTriggerType t) => switch (t) {
  AutomationTriggerType.AUTOMATION_TRIGGER_TYPE_MANUAL => 'Manual',
  AutomationTriggerType.AUTOMATION_TRIGGER_TYPE_SCHEDULE => 'Schedule',
  AutomationTriggerType.AUTOMATION_TRIGGER_TYPE_WEBHOOK =>
    'Webhook (not wired)',
  AutomationTriggerType.AUTOMATION_TRIGGER_TYPE_FORUM_EVENT =>
    'Forum event (not wired)',
  AutomationTriggerType.AUTOMATION_TRIGGER_TYPE_CHANNEL_EVENT =>
    'Channel event (not wired)',
  AutomationTriggerType.AUTOMATION_TRIGGER_TYPE_DAEMON_EVENT =>
    'Daemon event (not wired)',
  _ => 'Unspecified',
};
