import 'package:connectrpc/connect.dart';

enum OpsArea { overview, invocations, apiToken }

String opsFailureMessage(Object failure, OpsArea area) {
  if (failure is ConnectException) {
    return switch (failure.code) {
      Code.permissionDenied => switch (area) {
        OpsArea.overview =>
          'Dashboard metrics are limited to global administrators on this server.',
        OpsArea.apiToken =>
          'Only workspace owners and admins can create or revoke API tokens.',
        OpsArea.invocations => 'You do not have permission for this operation.',
      },
      Code.unauthenticated => 'Sign in again to continue.',
      Code.notFound => switch (area) {
        OpsArea.invocations => 'This invocation is no longer available.',
        _ => 'This resource is no longer available.',
      },
      Code.failedPrecondition =>
        'This operation is not available on the server right now.',
      Code.unavailable =>
        'The server could not complete the request. Try again.',
      _ => 'Something went wrong. Refresh and try again.',
    };
  }
  return 'Something went wrong. Refresh and try again.';
}
