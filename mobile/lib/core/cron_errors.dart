import 'package:connectrpc/connect.dart';

enum CronOperation { load, create, update, delete, run, executions, reply }

/// Safe wording for cron failures; never echoes raw server text.
String? describeCronFailure(Object failure, CronOperation operation) {
  if (failure is! ConnectException) return null;
  final message = failure.message.toLowerCase();
  switch (failure.code) {
    case Code.invalidArgument:
      if (message.contains('channel delivery is no longer accepted') ||
          message.contains('use telegram_destination')) {
        return 'Legacy CHANNEL delivery is not accepted. Choose Log, Webhook, Notify Group or Telegram Destination.';
      }
      if (message.contains('telegram_destination_id')) {
        return 'Choose a Telegram Destination that exists and has outbound delivery enabled.';
      }
      if (message.contains('schedule') || message.contains('timezone')) {
        return 'Check the schedule expression and timezone (IANA name, or leave empty for UTC).';
      }
      if (operation == CronOperation.create ||
          operation == CronOperation.update) {
        return 'The server rejected these values. Check the schedule, Agent, delivery fields and timeout/retry numbers.';
      }
      return null;
    case Code.notFound:
      return switch (operation) {
        CronOperation.load || CronOperation.executions =>
          'This cron job no longer exists. Refresh the list.',
        CronOperation.run =>
          'This cron job no longer exists, so nothing was run. Refresh the list.',
        _ =>
          'This cron job no longer exists, so nothing was changed. Refresh the list.',
      };
    case Code.permissionDenied:
      return 'You do not have access to this Workspace, or your role does not allow this.';
    default:
      return null;
  }
}

String cronFailureMessage(Object failure, CronOperation operation) {
  final specific = describeCronFailure(failure, operation);
  if (specific != null) return specific;
  if (failure is ConnectException) {
    return switch (failure.code) {
      Code.permissionDenied => 'You do not have permission for this operation.',
      Code.invalidArgument => 'Check the entered values.',
      Code.notFound => 'This cron job is no longer available.',
      Code.failedPrecondition => 'The server is not ready for this operation.',
      Code.deadlineExceeded || Code.unavailable =>
        'Cannot reach the server. Check your connection and retry.',
      _ => 'The request failed. Retry explicitly.',
    };
  }
  return 'The operation failed. Check your connection, then retry.';
}

bool cronWriteOutcomeUnknown(Object failure) =>
    failure is ConnectException &&
    const [
      Code.unavailable,
      Code.deadlineExceeded,
      Code.unknown,
      Code.internal,
    ].contains(failure.code);
