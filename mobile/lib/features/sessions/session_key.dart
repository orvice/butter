import '../../gen/agents/v1/agent_service.pb.dart';

/// The server address of one ADK session: app (channel), user and session id.
class SessionKey {
  const SessionKey({
    required this.appName,
    required this.userId,
    required this.sessionId,
  });

  factory SessionKey.fromInfo(SessionInfo info) => SessionKey(
    appName: info.appName,
    userId: info.userId,
    sessionId: info.sessionId,
  );

  final String appName;
  final String userId;
  final String sessionId;

  String get storageKey => '$appName/$userId/$sessionId';

  /// True when this session is the dashboard's AG-UI thread namespace.
  bool get isAguiThread => appName == 'agui' && sessionId.startsWith('agui-');

  @override
  bool operator ==(Object other) =>
      other is SessionKey &&
      appName == other.appName &&
      userId == other.userId &&
      sessionId == other.sessionId;

  @override
  int get hashCode => Object.hash(appName, userId, sessionId);
}
