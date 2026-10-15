import 'package:connectrpc/connect.dart';

enum BoxOperation {
  load,
  create,
  update,
  delete,
  token,
  status,
  piModels,
  cursorModels,
}

const _referencePrefix = 'is referenced by agents:';

const _manualHint = 'You can still enter the value by hand.';

/// Specific, safe wording for the failures a ButterBox operation can meet, or
/// null to fall back to the generic message.
///
/// Server text is never echoed. The one thing taken from it is the list of
/// Agent IDs that block a delete: the server builds it only from Agent IDs,
/// and it is the reason the delete was refused. Everything else is decided
/// from the status code, the operation and the server's stable, documented
/// phrases (`CURSOR_API_KEY`, `CursorService`, credential encryption).
String? describeBoxFailure(Object failure, BoxOperation operation) {
  if (failure is! ConnectException) return null;
  final message = failure.message;
  switch (failure.code) {
    case Code.failedPrecondition:
      if (message.contains('credential encryption is not configured')) {
        return 'This server cannot store ButterBox access tokens because credential encryption is not configured. Ask the platform administrator.';
      }
      if (operation == BoxOperation.delete) {
        final at = message.indexOf(_referencePrefix);
        if (at < 0) return null;
        var references = message
            .substring(at + _referencePrefix.length)
            .split(';')
            .first
            .trim();
        if (references.length > 400) {
          references = '${references.substring(0, 400)}…';
        }
        return 'This ButterBox cannot be deleted while Agents use it: $references. Repoint or delete those Agents first, then try again.';
      }
      if (operation == BoxOperation.cursorModels) {
        if (message.contains('does not serve CursorService')) {
          return 'This ButterBox does not serve Cursor. Upgrade butter-box on the box. $_manualHint';
        }
        if (message.contains('CURSOR_API_KEY')) {
          return 'The ButterBox refused the Cursor model catalog. Check the box access token and that CURSOR_API_KEY is set on the box. $_manualHint';
        }
      }
      return null;
    case Code.alreadyExists:
      return 'A ButterBox with this name already exists in this Workspace. Choose another name.';
    case Code.notFound:
      return switch (operation) {
        BoxOperation.load =>
          'This ButterBox no longer exists. Refresh the list.',
        BoxOperation.piModels ||
        BoxOperation.cursorModels ||
        BoxOperation.status =>
          'This ButterBox no longer exists. Refresh the list.',
        _ =>
          'This ButterBox no longer exists, so nothing was changed. Refresh the list.',
      };
    case Code.invalidArgument:
      return operation == BoxOperation.create ||
              operation == BoxOperation.update
          ? 'The server rejected these values. Enter a name and an absolute http or https base URL.'
          : null;
    case Code.resourceExhausted:
      if (operation == BoxOperation.piModels) {
        return 'The ButterBox is at its session capacity, so it could not list models. Wait and try again, or raise PI_API_MAX_SESSIONS on the box. $_manualHint';
      }
      return null;
    case Code.unavailable || Code.deadlineExceeded:
      if (operation == BoxOperation.piModels) {
        return 'The ButterBox did not answer the model catalog. Check that it is running and that its access token is valid. $_manualHint';
      }
      if (operation == BoxOperation.cursorModels) {
        return 'The ButterBox did not answer the Cursor model catalog. Check that it is running. $_manualHint';
      }
      return null;
    default:
      return null;
  }
}

/// [describeBoxFailure], or the generic wording for the status code. Never
/// echoes server text or credentials.
String boxFailureMessage(Object failure, BoxOperation operation) {
  final specific = describeBoxFailure(failure, operation);
  if (specific != null) return specific;
  if (failure is FormatException) return failure.message;
  if (failure is ConnectException) {
    return switch (failure.code) {
      Code.permissionDenied => 'You do not have permission for this operation.',
      Code.invalidArgument => 'Check the entered values.',
      Code.notFound => 'This ButterBox is no longer available.',
      Code.failedPrecondition => 'The server is not ready for this operation.',
      Code.deadlineExceeded || Code.unavailable =>
        'Cannot reach the server. Check your connection and retry.',
      _ => 'The request failed. Retry explicitly.',
    };
  }
  return 'The operation failed. Check your connection, then retry.';
}

/// True when a failed write may nevertheless have been applied, so it must be
/// reconciled by a read and never repeated on its own.
bool boxWriteOutcomeUnknown(Object failure) =>
    failure is! FormatException &&
    (failure is! ConnectException ||
        const [
          Code.unavailable,
          Code.deadlineExceeded,
          Code.unknown,
          Code.internal,
        ].contains(failure.code));
