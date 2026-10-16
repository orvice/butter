import 'package:connectrpc/connect.dart';

/// Safe, actionable wording for SessionService failures. Server payloads are
/// never echoed.
String sessionFailureMessage(Object failure, SessionOperation operation) {
  if (failure is ConnectException) {
    final code = failure.code;
    final message = failure.message.toLowerCase();
    if (code == Code.unauthenticated) {
      return 'Session expired. Sign in again.';
    }
    if (code == Code.permissionDenied) {
      return switch (operation) {
        SessionOperation.list =>
          'You may not list sessions with these filters.',
        SessionOperation.read => 'You may not view this session.',
        SessionOperation.rename => 'You may not rename this session.',
        SessionOperation.delete => 'You may not delete this session.',
        SessionOperation.reply => 'You may not reply to this session.',
      };
    }
    if (code == Code.notFound) {
      return 'Session not found.';
    }
    if (code == Code.invalidArgument) {
      if (message.contains('title')) {
        return 'Enter a non-empty title up to 100 characters.';
      }
      if (message.contains('agent')) {
        return 'Choose an Agent that exists in this Workspace.';
      }
      return 'The server refused this session request.';
    }
    if (code == Code.failedPrecondition) {
      if (message.contains('agent') && message.contains('not found')) {
        return 'Choose an Agent that exists in this Workspace.';
      }
      return 'The server is not ready for session operations.';
    }
    if (code == Code.unavailable) {
      if (operation == SessionOperation.delete) {
        if (message.contains('hold') || message.contains('run')) {
          return 'A run still holds this thread. Stop it or wait, then try deleting again.';
        }
        return 'The session could not be deleted right now. Try again later.';
      }
      return 'The server is temporarily unavailable.';
    }
    if (code == Code.deadlineExceeded || code == Code.unavailable) {
      return 'The request timed out.';
    }
  }
  return switch (operation) {
    SessionOperation.list => 'Could not load sessions.',
    SessionOperation.read => 'Could not load this session.',
    SessionOperation.rename => 'Could not rename this session.',
    SessionOperation.delete => 'Could not delete this session.',
    SessionOperation.reply => 'Could not send the reply.',
  };
}

enum SessionOperation { list, read, rename, delete, reply }
