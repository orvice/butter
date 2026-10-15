import 'package:connectrpc/connect.dart';

import 'agui_api.dart';

/// Stable, non-sensitive text for a chat transport or request failure. Server
/// payloads are never echoed: the message depends on the status or code alone.
/// (A run's own failure message is run content, not a transport error; the
/// controller shows it bounded, as the dashboard does.)
String describeChatFailure(Object failure) {
  if (failure is AguiHttpError) return describeAguiRefusal(failure);
  if (failure is AguiNotSent) return 'Could not reach the server.';
  if (failure is ConnectException) {
    return switch (failure.code) {
      Code.permissionDenied =>
        'You do not have access to chat in this Workspace.',
      Code.notFound => 'That thread is no longer available.',
      Code.invalidArgument => 'The server rejected that request.',
      Code.deadlineExceeded || Code.unavailable =>
        'Cannot reach the server. Check your connection and retry.',
      _ => 'The request failed. Retry explicitly.',
    };
  }
  if (failure is FormatException) return failure.message;
  return 'The request failed. Check your connection, then retry.';
}

/// What a refused run request means for the person, by status and code.
String describeAguiRefusal(AguiHttpError error) => switch ((
  error.status,
  error.code,
)) {
  (400, 'thread_unbound') =>
    'This thread was created before background runs existed and cannot be continued here. Start a new thread.',
  (400, _) => 'The server rejected this message. Check its text and images.',
  (403, _) =>
    'This thread is not available with this agent. Start a new thread.',
  (404, _) => 'This agent is not available in this Workspace.',
  (409, 'run_exists') => 'That message was already sent.',
  (409, _) => 'This thread is busy with another run.',
  (413, _) => 'This message is too large for the server.',
  (503, _) => 'The server is not ready. Try again in a moment.',
  _ => 'The request failed. Retry explicitly.',
};
