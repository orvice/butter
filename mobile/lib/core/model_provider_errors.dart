import 'package:connectrpc/connect.dart';

enum ModelProviderOperation { load, create, update, delete }

const _referencePrefix = 'models are referenced by telegram destinations:';

/// Specific, safe wording for the failures a Model Provider operation can
/// meet, or null to fall back to the generic message.
///
/// Server text is never echoed except the reference list, which the server
/// builds only from model aliases and Telegram Destination IDs (never the
/// provider's credential) and which is the reason an operation is refused.
String? describeModelProviderFailure(
  Object failure,
  ModelProviderOperation operation,
) {
  if (failure is! ConnectException) return null;
  switch (failure.code) {
    case Code.failedPrecondition:
      final message = failure.message;
      final at = message.indexOf(_referencePrefix);
      if (at < 0) return null;
      var references = message.substring(at + _referencePrefix.length).trim();
      if (references.length > 400) {
        references = '${references.substring(0, 400)}…';
      }
      return operation == ModelProviderOperation.delete
          ? 'This provider cannot be deleted while Telegram destinations use its models: $references. Change those destinations first, then try again.'
          : 'These models cannot be removed while Telegram destinations use them: $references. Keep the models or change those destinations first.';
    case Code.alreadyExists:
      return 'A model provider with this name already exists in this Workspace. Choose another name.';
    case Code.notFound:
      return operation == ModelProviderOperation.load
          ? 'This model provider no longer exists. Refresh the list.'
          : 'This model provider no longer exists, so nothing was changed. Refresh the list.';
    case Code.invalidArgument:
      return 'The server rejected these values. Use an absolute http or https base URL, or leave it empty.';
    default:
      return null;
  }
}
