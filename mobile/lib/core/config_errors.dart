import 'package:connectrpc/connect.dart';

import '../gen/agents/v1/dashboard.pb.dart';

/// What a configuration screen is about, for wording.
enum ConfigKind { mcp, preset, remote, daemon }

enum ConfigOperation {
  load,
  create,
  update,
  delete,
  status,
  tools,
  oauthStart,
  oauthStatus,
  oauthDisconnect,
  install,

  /// Minting a worker token for a daemon runtime.
  mint,

  /// Revoking a worker token.
  revoke,
}

/// A change the server cannot make safely, found before anything was sent or
/// answered by the server as an unknown method, so nothing was stored. It is a
/// definite answer, never an unknown outcome.
class ConfigUnsupported implements Exception {
  const ConfigUnsupported(this.message);
  final String message;

  @override
  String toString() => message;
}

String _subject(ConfigKind kind) => switch (kind) {
  ConfigKind.mcp => 'MCP server',
  ConfigKind.preset => 'preset',
  ConfigKind.remote => 'Remote Agent',
  ConfigKind.daemon => 'daemon runtime',
};

/// Specific, safe wording for the failures an MCP or Remote Agent operation
/// can meet, or null to fall back to the generic message.
///
/// Server text is never echoed. A failure message can carry the failing
/// server's own words (a probe error, an OAuth discovery error), which may
/// hold URLs and credentials; only its status code, the operation and a few
/// of the server's stable, documented phrases decide the wording.
String? describeConfigFailure(
  Object failure,
  ConfigKind kind,
  ConfigOperation operation,
) {
  if (failure is! ConnectException) return null;
  final message = failure.message.toLowerCase();
  final subject = _subject(kind);
  switch (failure.code) {
    case Code.permissionDenied:
      if (operation == ConfigOperation.revoke) {
        return 'Revoking a worker token needs the owner or admin role in this Workspace. Nothing was changed.';
      }
      return 'You do not have access to this Workspace, or your role does not allow this. Nothing was changed.';
    case Code.alreadyExists:
      if (operation == ConfigOperation.install) {
        return 'This Workspace already has an MCP server with this ID that was not installed from this preset, so nothing was installed. Delete or rename that server first.';
      }
      return 'A $subject with this ID already exists in this Workspace. Choose another ID.';
    case Code.notFound:
      if (operation == ConfigOperation.install) {
        return 'This preset is no longer offered, so nothing was installed. Refresh the presets.';
      }
      if (operation == ConfigOperation.mint) {
        return 'This daemon runtime no longer exists in this Workspace, so no token was created. Refresh the list.';
      }
      if (operation == ConfigOperation.revoke) {
        return 'This token no longer exists, so nothing was changed. Refresh the tokens.';
      }
      return switch (operation) {
        ConfigOperation.load ||
        ConfigOperation.status ||
        ConfigOperation.tools ||
        ConfigOperation.oauthStatus =>
          'This $subject no longer exists. Refresh the list.',
        _ =>
          'This $subject no longer exists, so nothing was changed. Refresh the list.',
      };
    case Code.invalidArgument:
      if (operation == ConfigOperation.create ||
          operation == ConfigOperation.update) {
        return switch (kind) {
          ConfigKind.remote =>
            'The server rejected these values. Check the URL (absolute http or https), the protocol fields and that a username comes with a password.',
          ConfigKind.daemon =>
            'The server rejected these values. Check that the name is not empty.',
          _ =>
            'The server rejected these values. Check the URL (absolute http or https), the authentication fields and that a client secret comes with a client ID.',
        };
      }
      return null;
    case Code.failedPrecondition:
      if (operation == ConfigOperation.oauthStart) {
        if (message.contains('oauth service is not configured')) {
          return 'MCP OAuth is not enabled on this server. Ask the platform administrator.';
        }
        if (message.contains('callback_base_url')) {
          return 'This server has no public address configured for MCP OAuth. Ask the platform administrator.';
        }
        if (message.contains('not configured for oauth2')) {
          return 'This MCP server is not set to OAuth 2.0. Edit it first.';
        }
        return 'Butter could not start the authorization. Check the OAuth settings (client, endpoints, resource); if the provider requires a pre-registered client, enter its client ID.';
      }
      if (operation == ConfigOperation.install) {
        return 'This server cannot install presets right now.';
      }
      return null;
    case Code.unavailable || Code.deadlineExceeded:
      return null;
    default:
      return null;
  }
}

/// [describeConfigFailure], or the generic wording for the status code. Never
/// echoes server text or credentials.
String configFailureMessage(
  Object failure,
  ConfigKind kind,
  ConfigOperation operation,
) {
  final specific = describeConfigFailure(failure, kind, operation);
  if (specific != null) return specific;
  if (failure is FormatException) return failure.message;
  if (failure is ConfigUnsupported) return failure.message;
  if (failure is ConnectException) {
    return switch (failure.code) {
      Code.permissionDenied => 'You do not have permission for this operation.',
      Code.invalidArgument => 'Check the entered values.',
      Code.notFound => 'This ${_subject(kind)} is no longer available.',
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
bool configWriteOutcomeUnknown(Object failure) =>
    failure is! FormatException &&
    failure is! ConfigUnsupported &&
    (failure is! ConnectException ||
        const [
          Code.unavailable,
          Code.deadlineExceeded,
          Code.unknown,
          Code.internal,
        ].contains(failure.code));

/// What to show for an MCP server that did not connect, from the probe text
/// the server reports. That text is the raw error of the connection attempt:
/// a transport error, the MCP server's own answer or an OAuth failure, and it
/// can carry whatever the server (or an intermediary) put in it. It is
/// therefore only classified, never displayed.
String mcpProbeReason(String serverText) {
  final text = serverText.toLowerCase();
  bool any(List<String> words) => words.any(text.contains);
  if (text.trim().isEmpty) return 'The MCP server did not answer.';
  if (any(['timeout_seconds'])) {
    return 'This server\'s timeout setting is not valid. Edit it.';
  }
  if (any(['requires reauthorization', 'refresh failed', 'not usable'])) {
    return 'The authorization expired or was revoked. Connect the account again.';
  }
  if (any(['requires authorization'])) {
    return 'Authorization is required. Connect the account.';
  }
  if (any(['unauthorized', '401', 'forbidden', '403'])) {
    return 'The MCP server rejected the credentials. Check the headers, or connect the account.';
  }
  if (any(['deadline', 'timeout', 'timed out'])) {
    return 'The MCP server did not answer in time.';
  }
  if (any([
    'refused',
    'no such host',
    'dial',
    'connect',
    'eof',
    'certificate',
    'tls',
    'unreachable',
    'unavailable',
  ])) {
    return 'The server could not connect to the MCP server. Check its URL and that it is running.';
  }
  return 'The MCP server could not be used.';
}

/// What to show for a Remote Agent from its probe. The detail is the server's
/// text and is classified like [mcpProbeReason]; only a three-digit HTTP
/// status the probe reports is carried over.
String remoteProbeReason(RemoteAgentStatus_State state, String serverText) {
  final text = serverText.toLowerCase();
  bool any(List<String> words) => words.any(text.contains);
  final status = RegExp(r'returned (\d{3})').firstMatch(text)?.group(1);
  switch (state) {
    case RemoteAgentStatus_State.STATE_ACTIVE:
    case RemoteAgentStatus_State.STATE_IDLE:
      return 'Reachable.';
    case RemoteAgentStatus_State.STATE_CONFIGURED:
      return 'Configured; the server does not probe this protocol.';
    case RemoteAgentStatus_State.STATE_UNREACHABLE:
      if (any(['not online'])) return 'The daemon runtime is not online.';
      if (status != null) {
        return 'The endpoint answered HTTP $status instead of a healthy response.';
      }
      if (any(['deadline', 'timeout', 'timed out'])) {
        return 'The endpoint did not answer in time.';
      }
      return 'The server could not reach the endpoint. Check its URL and that it is running.';
    case RemoteAgentStatus_State.STATE_ERROR:
      if (any(['401', 'username/password'])) {
        return 'The endpoint rejected the username or password.';
      }
      if (any(['acp_runtime'])) {
        return 'The executor must be opencode or codex.';
      }
      if (any(['is required'])) {
        return 'The configuration is incomplete for this protocol.';
      }
      return 'The probe failed. Check the configuration.';
    default:
      return 'The server could not tell whether this endpoint is reachable.';
  }
}
