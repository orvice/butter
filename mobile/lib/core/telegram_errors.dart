import 'package:connectrpc/connect.dart';

enum TelegramOperation {
  load,
  createChannel,
  updateChannel,
  rotateToken,
  setEnabled,
  deleteChannel,
  status,
  createDestination,
  updateDestination,
  deleteDestination,
  testSend,
  resend,
  processing,
}

const telegramUnconfirmed =
    'Change not confirmed; it may already have been applied. Refresh before retrying.';

String? describeTelegramFailure(Object failure, TelegramOperation operation) {
  if (failure is! ConnectException) return null;
  final message = failure.message;
  switch (failure.code) {
    case Code.permissionDenied:
      return switch (operation) {
        TelegramOperation.resend ||
        TelegramOperation.createChannel ||
        TelegramOperation.updateChannel ||
        TelegramOperation.rotateToken ||
        TelegramOperation.setEnabled ||
        TelegramOperation.deleteChannel ||
        TelegramOperation.createDestination ||
        TelegramOperation.updateDestination ||
        TelegramOperation.deleteDestination =>
          'You need workspace owner or admin access for this action.',
        _ => 'You do not have permission to do this.',
      };
    case Code.failedPrecondition:
      if (message.contains('webhook_base_url') ||
          message.contains('Telegram webhook base URL')) {
        return 'Telegram webhook delivery is not configured on this server. Ask a platform administrator to set the public webhook base URL.';
      }
      if (message.contains('queue') || message.contains('Redis')) {
        return 'Telegram inbound delivery needs a configured queue on this server. Ask a platform administrator.';
      }
      if (operation == TelegramOperation.deleteChannel &&
          message.contains('destination')) {
        return 'Delete every Destination under this Channel first, then try again.';
      }
      if (operation == TelegramOperation.deleteDestination &&
          (message.contains('Notify') || message.contains('Cron'))) {
        return 'This Destination is still used by a Notify Group or Cron job. Remove those references first.';
      }
      if (operation == TelegramOperation.rotateToken &&
          message.contains('bot id')) {
        return 'The new Bot token must belong to the same Bot as this Channel.';
      }
      if (operation == TelegramOperation.resend) {
        if (message.contains('no complete response')) {
          return 'Nothing was stored to resend. The Agent may not have finished.';
        }
        if (message.contains('already delivered')) {
          return 'Every part of this reply was already delivered.';
        }
        if (message.contains('already being delivered')) {
          return 'This reply is already being delivered. Wait and refresh.';
        }
      }
      if (operation == TelegramOperation.setEnabled &&
          message.contains('outbound')) {
        return 'Turn on sending before turning on receiving: every accepted message must be able to get a reply.';
      }
      return null;
    case Code.invalidArgument:
      if (message.contains('chat_id') ||
          message.contains('message_thread_id')) {
        return 'Use Telegram chat and topic IDs as decimal strings (for example -1001234567890).';
      }
      if (message.contains('bot_token') || message.contains('getMe')) {
        return 'Telegram rejected this Bot token. Check the token from BotFather and try again.';
      }
      if (message.contains('agent_id') || message.contains('Agent')) {
        return 'Choose an active Agent in this Workspace.';
      }
      if (message.contains('model') || message.contains('alias')) {
        return 'Choose a model alias from your Model providers, or leave model empty to inherit.';
      }
      return null;
    case Code.notFound:
      return switch (operation) {
        TelegramOperation.load ||
        TelegramOperation.status ||
        TelegramOperation.processing =>
          'This item no longer exists. Refresh the list.',
        _ =>
          'This item no longer exists, so nothing was changed. Refresh the list.',
      };
    case Code.aborted:
      if (operation == TelegramOperation.resend) {
        return 'This reply is already being delivered. Wait and refresh.';
      }
      return null;
    default:
      return null;
  }
}

String telegramFailureMessage(Object failure, TelegramOperation operation) {
  final specific = describeTelegramFailure(failure, operation);
  if (specific != null) return specific;
  if (failure is ConnectException) {
    if (failure.code == Code.unauthenticated) {
      return 'Your session expired. Sign in again.';
    }
    return failure.message.isEmpty
        ? 'The server refused this Telegram request.'
        : failure.message;
  }
  return 'Something went wrong. Try again after refreshing.';
}
