import 'dart:convert';
import 'dart:io';

import 'package:butter_mobile/gen/agents/v1/agent_service.pb.dart';
import 'package:fixnum/fixnum.dart';
import 'package:protobuf/well_known_types/google/protobuf/duration.pb.dart'
    as pbd;
import 'package:protobuf/well_known_types/google/protobuf/timestamp.pb.dart';

import 'file_server.dart';

/// Controllable SessionService boundary for widget and controller tests.
class SessionServerFake extends FileServerFake {
  final sessions = <String, _StoredSession>{};
  final sessionCalls = <({String method, String? workspace, List<int> body})>[];

  static const writeMethods = {
    'UpdateSessionTitle',
    'DeleteSession',
    'ReplySession',
  };

  int sessionWriteCount(String method) =>
      sessionCalls.where((c) => c.method == method).length;

  Future<void> Function(String method)? sessionBefore;
  final sessionRefusals = <String, (int, String, String)>{};
  final lostSessionReplies = <String>{};

  SessionServerFake() {
    final inherited = authorizedReply!;
    authorizedReply = (r, bytes) async {
      final path = r.uri.path;
      if (!path.contains('SessionService')) return inherited(r, bytes);
      final method = path.split('/').last;
      final ws = r.headers.value('x-workspace-id');
      sessionCalls.add((method: method, workspace: ws, body: bytes));
      if (sessionBefore != null) await sessionBefore!(method);
      final refusal = sessionRefusals[method];
      if (refusal != null) {
        rpcFail(r, refusal.$1, refusal.$2, refusal.$3);
        return true;
      }
      if (writeMethods.contains(method) &&
          lostSessionReplies.contains(method)) {
        rpcFail(r, 503, 'unavailable', 'Reply lost');
        return true;
      }
      return _sessionReply(r, method, bytes);
    };
  }

  void seedSession({
    required String workspace,
    required String app,
    required String user,
    required String id,
    String title = '',
    List<SessionEvent> events = const [],
  }) {
    final key = '$workspace/$app/$user/$id';
    sessions[key] = _StoredSession(
      workspace: workspace,
      app: app,
      user: user,
      id: id,
      title: title,
      events: List.of(events),
    );
  }

  bool _sessionReply(HttpRequest r, String method, List<int> body) {
    switch (method) {
      case 'ListSessions':
        return _list(r, body);
      case 'GetSession':
        return _get(r, body);
      case 'UpdateSessionTitle':
        return _rename(r, body);
      case 'DeleteSession':
        return _delete(r, body);
      case 'ReplySession':
        return _reply(r, body);
      default:
        r.response.statusCode = 404;
        return true;
    }
  }

  bool _list(HttpRequest r, List<int> body) {
    final req = ListSessionsRequest.fromBuffer(body);
    final ws = r.headers.value('x-workspace-id') ?? '';
    final all = sessions.values.where((s) {
      if (s.workspace != ws) return false;
      if (req.appName.isNotEmpty && s.app != req.appName) return false;
      if (req.userId.isNotEmpty && s.user != req.userId) return false;
      return true;
    }).toList();
    final pageSize = req.pageSize <= 0 ? 20 : req.pageSize;
    final start = req.pageToken.isEmpty
        ? 0
        : int.tryParse(utf8.decode(base64Decode(req.pageToken))) ?? 0;
    final slice = all.skip(start).take(pageSize).toList();
    final next = start + pageSize < all.length
        ? base64Encode(utf8.encode('${start + pageSize}'))
        : '';
    write(
      r,
      ListSessionsResponse(
        sessions: [for (final s in slice) s.info()],
        nextPageToken: next,
        total: all.length,
      ),
    );
    return true;
  }

  bool _get(HttpRequest r, List<int> body) {
    final req = GetSessionRequest.fromBuffer(body);
    final ws = r.headers.value('x-workspace-id') ?? '';
    final s = sessions['$ws/${req.appName}/${req.userId}/${req.sessionId}'];
    if (s == null) {
      rpcFail(r, 404, 'not_found', 'session not found');
      return true;
    }
    final events = req.numRecentEvents <= 0
        ? s.events
        : s.events.reversed
              .take(req.numRecentEvents)
              .toList()
              .reversed
              .toList();
    write(
      r,
      GetSessionResponse(
        sessionDetail: SessionDetail(
          session: s.info(),
          events: events,
          duration: pbd.Duration(seconds: Int64(12)),
        ),
      ),
    );
    return true;
  }

  bool _rename(HttpRequest r, List<int> body) {
    final req = UpdateSessionTitleRequest.fromBuffer(body);
    final ws = r.headers.value('x-workspace-id') ?? '';
    final s = sessions['$ws/${req.appName}/${req.userId}/${req.sessionId}'];
    if (s == null) {
      rpcFail(r, 404, 'not_found', 'session not found');
      return true;
    }
    s.title = req.title;
    write(r, UpdateSessionTitleResponse(session: s.info()));
    return true;
  }

  bool _delete(HttpRequest r, List<int> body) {
    final req = DeleteSessionRequest.fromBuffer(body);
    final ws = r.headers.value('x-workspace-id') ?? '';
    sessions.remove('$ws/${req.appName}/${req.userId}/${req.sessionId}');
    write(r, DeleteSessionResponse());
    return true;
  }

  bool _reply(HttpRequest r, List<int> body) {
    final req = ReplySessionRequest.fromBuffer(body);
    if (req.agentId.isEmpty) {
      rpcFail(r, 400, 'invalid_argument', 'agent required');
      return true;
    }
    write(r, ReplySessionResponse(response: 'fixture(${req.message})'));
    return true;
  }
}

class _StoredSession {
  _StoredSession({
    required this.workspace,
    required this.app,
    required this.user,
    required this.id,
    required this.title,
    required this.events,
  });
  final String workspace;
  final String app;
  final String user;
  final String id;
  String title;
  final List<SessionEvent> events;

  SessionInfo info() => SessionInfo(
    sessionId: id,
    appName: app,
    userId: user,
    title: title,
    turnCount: events.length,
    lastUpdateTime: Timestamp(
      seconds: Int64(DateTime.now().millisecondsSinceEpoch ~/ 1000),
    ),
  );
}
