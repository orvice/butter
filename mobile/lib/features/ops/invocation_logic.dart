import '../../gen/agents/v1/agent_service.pb.dart';

/// Server-side source for detached AG-UI runs (cross-Pod Stop path).
const aguiDetachedSource = 'agui-detached';

bool invocationIsActive(Invocation inv) =>
    inv.status == InvocationStatus.INVOCATION_STATUS_RUNNING ||
    inv.status == InvocationStatus.INVOCATION_STATUS_QUEUED;

/// Whether [CancelAgentInvocation] is meaningful for this record.
bool invocationSupportsCancel(Invocation inv) => invocationIsActive(inv);

/// User-facing explanation of what cancel will do for this invocation.
String invocationCancelExplanation(Invocation inv) {
  if (inv.source == aguiDetachedSource) {
    return 'This is a detached Chat run. Cancel uses the AG-UI Stop path and '
        'can reach the run on another server instance when the server supports it. '
        'It does not guarantee the run has ended until the record shows a terminal state.';
  }
  return 'Cancel asks the agent runner on this server to stop the invocation. '
      'Runs on another instance or entry points without runner registration may not stop.';
}

String invocationStatusLabel(InvocationStatus status) => switch (status) {
  InvocationStatus.INVOCATION_STATUS_RUNNING => 'Running',
  InvocationStatus.INVOCATION_STATUS_QUEUED => 'Queued',
  InvocationStatus.INVOCATION_STATUS_SUCCEEDED => 'Succeeded',
  InvocationStatus.INVOCATION_STATUS_FAILED => 'Failed',
  InvocationStatus.INVOCATION_STATUS_CANCELLED => 'Cancelled',
  _ => 'Unknown',
};
