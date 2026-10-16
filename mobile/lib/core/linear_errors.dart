import 'package:connectrpc/connect.dart';

enum LinearOperation {
  load,
  create,
  update,
  delete,
  credentials,
  install,
  resend,
}

String linearFailureMessage(Object failure, LinearOperation operation) {
  if (failure is ConnectException) {
    final msg = failure.message.toLowerCase();
    switch (failure.code) {
      case Code.permissionDenied:
        return switch (operation) {
          LinearOperation.create ||
          LinearOperation.update ||
          LinearOperation.delete ||
          LinearOperation.credentials ||
          LinearOperation.install ||
          LinearOperation.resend =>
            'This operation needs the workspace owner or admin role. Nothing was changed.',
          _ => 'You do not have access to this Workspace.',
        };
      case Code.alreadyExists:
        return 'This Linear client ID is already registered in Butter. Use another OAuth app.';
      case Code.aborted:
        return 'Someone else changed this Linear App. Reload and try again.';
      case Code.notFound:
        return switch (operation) {
          LinearOperation.load => 'This Linear App is no longer available.',
          LinearOperation.resend =>
            'This delivery record is no longer available, so nothing was resent.',
          _ =>
            'This Linear App no longer exists, so nothing was changed. Refresh the list.',
        };
      case Code.invalidArgument:
        return 'The server rejected these values. Check the display name, agent, allowlist and run limit.';
      case Code.failedPrecondition:
        if (msg.contains('public base url')) {
          return 'A platform administrator must configure the Linear public base URL before inbound traffic or installation can work.';
        }
        if (msg.contains('client secret before installing')) {
          return 'Set the OAuth client secret on this App before installing it into Linear.';
        }
        if (msg.contains('redis') || msg.contains('queue')) {
          return 'Inbound deliveries need a durable queue on the server. Ask the platform administrator.';
        }
        if (msg.contains('while inbound is on') ||
            msg.contains('inbound is enabled')) {
          return 'Turn inbound off before clearing secrets that inbound requires.';
        }
        if (msg.contains('not resendable') || msg.contains('no reply')) {
          return 'This record has no persisted reply to resend, or it is not in a failed state.';
        }
        if (msg.contains('linear runtime')) {
          return 'The Linear delivery service is not running on this server.';
        }
        return 'The server is not ready for this operation.';
      case Code.unavailable:
      case Code.deadlineExceeded:
        return 'Cannot reach the server. Check your connection and retry.';
      default:
        return 'The request failed. Retry explicitly.';
    }
  }
  return 'The operation failed. Check your connection, then retry.';
}

bool linearWriteOutcomeUnknown(Object failure) =>
    failure is ConnectException &&
    const [
      Code.unavailable,
      Code.deadlineExceeded,
      Code.unknown,
      Code.internal,
    ].contains(failure.code);
