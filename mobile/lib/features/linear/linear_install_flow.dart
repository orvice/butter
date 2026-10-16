import 'dart:async';

import '../../core/native_oauth.dart';

/// Linear App install uses the system browser and the server's OAuth callback;
/// nothing returns to the App with a token. The App learns the outcome by
/// re-reading installations.
const linearInstallFlowTtl = Duration(minutes: 10);
const linearInstallPollInterval = Duration(seconds: 5);

enum LinearInstallPhase {
  idle,
  starting,
  waiting,
  installed,
  failed,
  cancelled,
  timedOut,
}

class LinearInstallView {
  LinearInstallView({
    this.phase = LinearInstallPhase.idle,
    this.message,
    this.appId = '',
    this.baselineInstallationIds = const {},
    DateTime? startedAt,
    DateTime? deadline,
  }) : startedAt = startedAt ?? DateTime.fromMillisecondsSinceEpoch(0),
       deadline = deadline ?? DateTime.fromMillisecondsSinceEpoch(0);

  LinearInstallPhase phase;
  String? message;
  String appId;
  Set<String> baselineInstallationIds;
  DateTime startedAt;
  DateTime deadline;
  int sequence = 0;
}

/// Opens [authorizeUrl] in the system browser; returns false if the OS refused.
Future<bool> openLinearAuthorizeUrl(
  String authorizeUrl, {
  NativeOAuthPlatform? platform,
}) {
  platform ??= DeviceOAuthPlatform();
  final uri = Uri.tryParse(authorizeUrl);
  if (uri == null || !uri.hasScheme || uri.scheme != 'https') {
    return Future.value(false);
  }
  return platform.openBrowser(uri);
}
