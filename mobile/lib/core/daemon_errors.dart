import 'package:connectrpc/connect.dart';

/// How far a cancel got. The server's success means the cancel was handed to
/// the worker, never that the task stopped: the worker decides, and reports
/// its own end.
enum CancelKind {
  sending,

  /// Handed to the worker. The task keeps running until the worker reports.
  delivered,

  /// No connected worker runs the task any more.
  gone,

  /// The server could not hand it to the worker.
  workerUnreachable,

  /// The server refused the request.
  refused,

  /// The answer was lost; the cancel may or may not have been delivered.
  unknown,
  failed,
}

class CancelView {
  const CancelView(this.kind, this.message);
  final CancelKind kind;

  /// A safe sentence; never the server's own text.
  final String message;

  bool get sending => kind == CancelKind.sending;
}

const cancelSending = CancelView(
  CancelKind.sending,
  'Sending the cancel to the worker…',
);

CancelView cancelDelivered(String runtimeId) => CancelView(
  CancelKind.delivered,
  'Cancel delivered to runtime "$runtimeId". The task is still running until the worker reports back; the worker decides whether and how it stops.',
);

const cancelLostWithApp = CancelView(
  CancelKind.unknown,
  'The cancel was not confirmed before the app left the foreground. It may have been delivered. Refresh the tasks; send it again only if the task is still listed.',
);

/// What a failed cancel call tells the person. Repeating a cancel is harmless,
/// so no outcome blocks another attempt; each only says what is known.
CancelView cancelFailureView(Object failure) {
  if (failure is ConnectException) {
    switch (failure.code) {
      case Code.notFound:
        return const CancelView(
          CancelKind.gone,
          'No connected worker is running this task any more, so nothing was delivered. It may have finished, or the worker disconnected. Refresh the tasks.',
        );
      case Code.internal:
        return const CancelView(
          CancelKind.workerUnreachable,
          'The server could not hand the cancel to the worker; it may have just disconnected. The task may still be running. Refresh the tasks, then retry if it is still listed.',
        );
      case Code.permissionDenied:
        return const CancelView(
          CancelKind.refused,
          'You do not have access to this Workspace, so nothing was cancelled.',
        );
      case Code.failedPrecondition:
        return const CancelView(
          CancelKind.refused,
          'The server needs a selected Workspace for this, so nothing was cancelled.',
        );
      case Code.unavailable || Code.deadlineExceeded || Code.unknown:
        return const CancelView(
          CancelKind.unknown,
          'The connection to the server was lost. The cancel may or may not have been delivered. Refresh the tasks; send it again only if the task is still listed.',
        );
      default:
        return const CancelView(
          CancelKind.failed,
          'The server did not accept the cancel. The task may still be running.',
        );
    }
  }
  return const CancelView(
    CancelKind.unknown,
    'The cancel was not confirmed. It may or may not have been delivered. Refresh the tasks.',
  );
}
