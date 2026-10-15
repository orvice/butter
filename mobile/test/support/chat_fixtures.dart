import 'dart:convert';

import 'package:butter_mobile/gen/agents/v1/agent.pb.dart';
import 'package:butter_mobile/gen/agents/v1/agent_service.pb.dart';
import 'package:fixnum/fixnum.dart';
import 'package:protobuf/well_known_types/google/protobuf/struct.pb.dart';
import 'package:protobuf/well_known_types/google/protobuf/timestamp.pb.dart';

/// A session as ListSessions returns an AG-UI thread's: `agui-{threadId}`
/// under app `agui`, with the hidden UI binding as a JSON string in state.
SessionInfo session(
  String threadId, {
  String app = 'agui',
  String? sessionId,
  String title = '',
  Value? binding,
  String workspace = 'alpha',
  String agent = 'agent-1',
  String? bindingThread,
  int updated = 0,
  String user = 'alice',
}) {
  final value =
      binding ??
      Value(
        stringValue: jsonEncode({
          'workspace_id': workspace,
          'agent_id': agent,
          'thread_id': bindingThread ?? threadId,
        }),
      );
  return SessionInfo(
    sessionId: sessionId ?? 'agui-$threadId',
    appName: app,
    userId: user,
    title: title,
    state: Struct(fields: {'butter:a2ui:binding': value}.entries),
    lastUpdateTime: Timestamp(seconds: Int64(updated)),
  );
}

SessionInfo unbound(String threadId) =>
    SessionInfo(sessionId: 'agui-$threadId', appName: 'agui', userId: 'alice');

Agent agent(String id, String name, {String displayName = ''}) =>
    Agent(agentId: id, name: name, displayName: displayName);
