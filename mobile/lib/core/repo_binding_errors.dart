import 'package:connectrpc/connect.dart';

import '../gen/agents/v1/repobinding.pb.dart';

enum RepoBindingOperation {
  load,
  save,
  credential,
  validate,
  onboard,
  sync,
  publish,
  detach,
  webhook,
  baseline,
  browse,
}

String repoBindingFailureMessage(
  Object failure,
  RepoBindingOperation operation,
) {
  if (failure is ConnectException) {
    final message = failure.message.toLowerCase();
    switch (failure.code) {
      case Code.permissionDenied:
        return 'You need the owner or admin role in this Workspace to change the repository binding. Nothing was changed.';
      case Code.notFound:
        if (operation == RepoBindingOperation.browse) {
          return 'Nothing is cached for that path yet. Sync the repository first.';
        }
        return 'This Workspace has no repository binding. Refresh the page.';
      case Code.failedPrecondition:
        if (message.contains('active revision') ||
            message.contains('snapshot') ||
            message.contains('materialize')) {
          return 'There is no valid Active Revision to materialize, so the binding was not removed. Choose a recovery option or publish first.';
        }
        if (message.contains('credential') || message.contains('pat')) {
          return 'Set a personal access token before validating or syncing.';
        }
        if (message.contains('git host')) {
          return 'The selected Git host is no longer configured. Pick another host.';
        }
        return 'The server refused this step. Read the binding status and try again.';
      case Code.invalidArgument:
        return 'The server rejected these values. Check the form and try again.';
      case Code.unauthenticated:
        return 'Session expired. Sign in again.';
      default:
        break;
    }
  }
  if (failure is FormatException) return failure.message;
  return switch (operation) {
    RepoBindingOperation.load => 'The repository binding could not be loaded.',
    RepoBindingOperation.save => 'The binding was not saved.',
    RepoBindingOperation.credential => 'The token was not saved.',
    RepoBindingOperation.validate => 'Validation did not complete.',
    RepoBindingOperation.onboard => 'Onboarding did not complete.',
    RepoBindingOperation.sync => 'Sync did not complete.',
    RepoBindingOperation.publish => 'Publish did not complete.',
    RepoBindingOperation.detach => 'The binding was not removed.',
    RepoBindingOperation.webhook => 'Webhook secret was not configured.',
    RepoBindingOperation.baseline => 'Baseline was not accepted.',
    RepoBindingOperation.browse => 'Repository cache could not be read.',
  };
}

String connectionStateLabel(
  RepoBindingConnectionState state,
) => switch (state) {
  RepoBindingConnectionState.REPO_BINDING_CONNECTION_STATE_OK => 'Connected',
  RepoBindingConnectionState.REPO_BINDING_CONNECTION_STATE_FAILED => 'Failed',
  RepoBindingConnectionState.REPO_BINDING_CONNECTION_STATE_DEGRADED =>
    'Degraded',
  RepoBindingConnectionState.REPO_BINDING_CONNECTION_STATE_DIVERGED =>
    'Diverged',
  RepoBindingConnectionState.REPO_BINDING_CONNECTION_STATE_UNVALIDATED ||
  RepoBindingConnectionState.REPO_BINDING_CONNECTION_STATE_UNSPECIFIED =>
    'Unvalidated',
  _ => 'Unknown',
};
