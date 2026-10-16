import 'package:connectrpc/connect.dart';

enum NotifyGroupOperation { load, create, update, delete }

String? describeNotifyGroupFailure(
  Object failure,
  NotifyGroupOperation operation,
) {
  if (failure is! ConnectException) return null;
  final message = failure.message.toLowerCase();
  switch (failure.code) {
    case Code.invalidArgument:
      if (message.contains('bot_token') ||
          message.contains('chat_id') ||
          message.contains('parse_mode') ||
          message.contains('message_thread_id')) {
        return 'Legacy raw Telegram fields are not accepted. Choose a Telegram Destination instead.';
      }
      if (message.contains('destination_id')) {
        return 'Choose a Telegram Destination that exists in this Workspace.';
      }
      if (operation == NotifyGroupOperation.create ||
          operation == NotifyGroupOperation.update) {
        return 'Check each target type and its required fields.';
      }
      return null;
    case Code.notFound:
      return 'This Notify Group no longer exists. Refresh the list.';
    case Code.permissionDenied:
      return 'You do not have access to this Workspace, or your role does not allow this.';
    default:
      return null;
  }
}

String notifyGroupFailureMessage(
  Object failure,
  NotifyGroupOperation operation,
) {
  final specific = describeNotifyGroupFailure(failure, operation);
  if (specific != null) return specific;
  if (failure is ConnectException) {
    return switch (failure.code) {
      Code.permissionDenied => 'You do not have permission for this operation.',
      Code.invalidArgument => 'Check the entered values.',
      Code.notFound => 'This Notify Group is no longer available.',
      Code.deadlineExceeded || Code.unavailable =>
        'Cannot reach the server. Check your connection and retry.',
      _ => 'The request failed. Retry explicitly.',
    };
  }
  return 'The operation failed. Check your connection, then retry.';
}

bool notifyGroupWriteOutcomeUnknown(Object failure) =>
    failure is ConnectException &&
    const [
      Code.unavailable,
      Code.deadlineExceeded,
      Code.unknown,
      Code.internal,
    ].contains(failure.code);
