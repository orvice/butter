import 'dart:async';
import 'dart:convert';
import 'dart:io';

import 'package:butter_mobile/gen/agents/v1/agent_service.pb.dart';

import 'server.dart';

/// One run request as the server received it.
class RunPost {
  RunPost({
    required this.agentId,
    required this.token,
    required this.workspace,
    required this.body,
    required this.contentType,
  });
  final String agentId;
  final String? token;
  final String? workspace;
  final String? contentType;
  final Map<String, dynamic> body;

  String get threadId => body['threadId'] as String;
  String get runId => body['runId'] as String;
  List<dynamic> get messages => body['messages'] as List<dynamic>;
  Object? get content => (messages.single as Map)['content'];

  /// Completes when the client closed its side of the connection.
  final disconnected = Completer<void>();
}

/// The response of one run, written by hand on the connection so a test
/// controls exactly what the client sees and when: headers, events, silence,
/// or a connection that drops.
class RunResponse {
  RunResponse(this._socket, this._post) {
    _socket.listen(
      (_) {},
      onDone: _closed,
      onError: (_) => _closed(),
      cancelOnError: true,
    );
  }
  final Socket _socket;
  final RunPost _post;
  var _open = false;
  var _finished = false;

  void _closed() {
    _finished = true;
    if (!_post.disconnected.isCompleted) _post.disconnected.complete();
  }

  bool get clientGone => _post.disconnected.isCompleted;

  void _raw(List<int> bytes) {
    if (_finished) return;
    try {
      _socket.add(bytes);
    } catch (_) {
      _closed();
    }
  }

  /// Answers with an HTTP error before any stream.
  Future<void> fail(int status, [Map<String, Object?>? body]) async {
    final payload = utf8.encode(jsonEncode(body ?? {'error': 'refused'}));
    _raw(
      utf8.encode(
        'HTTP/1.1 $status X\r\nContent-Type: application/json\r\n'
        'Content-Length: ${payload.length}\r\nConnection: close\r\n\r\n',
      ),
    );
    _raw(payload);
    await _end();
  }

  /// Opens the 200 event stream.
  void open() {
    if (_open) return;
    _open = true;
    _raw(
      utf8.encode(
        'HTTP/1.1 200 OK\r\nContent-Type: text/event-stream\r\n'
        'Transfer-Encoding: chunked\r\nConnection: close\r\n\r\n',
      ),
    );
  }

  void _chunk(String text) {
    final bytes = utf8.encode(text);
    _raw(utf8.encode('${bytes.length.toRadixString(16)}\r\n'));
    _raw(bytes);
    _raw(utf8.encode('\r\n'));
  }

  Future<void> flush() async {
    if (_finished) return;
    try {
      await _socket.flush();
    } catch (_) {
      _closed();
    }
  }

  /// Sends one AG-UI event.
  Future<void> event(Map<String, Object?> event) async {
    open();
    _chunk('data: ${jsonEncode(event)}\n\n');
    await flush();
  }

  /// Sends raw text as one chunk, for framing tests.
  Future<void> rawChunk(String text) async {
    open();
    _chunk(text);
    await flush();
  }

  Future<void> heartbeat() async {
    open();
    _chunk(': heartbeat\n\n');
    await flush();
  }

  /// Keeps the connection open, with a heartbeat now and then so a client
  /// that left is noticed, until [until] completes or the client is gone.
  Future<void> hold(Future<void> until) async {
    var released = false;
    unawaited(until.then((_) => released = true));
    while (!released && !clientGone) {
      await Future<void>.delayed(const Duration(milliseconds: 10));
    }
  }

  /// Ends the stream cleanly.
  Future<void> finish() async {
    open();
    _chunk('');
    await _end();
  }

  /// Drops the connection with no end, as a lost network does.
  Future<void> drop() async {
    _socket.destroy();
    _closed();
  }

  Future<void> _end() async {
    if (_finished) return;
    try {
      await _socket.flush();
      await _socket.close();
    } catch (_) {}
    _closed();
  }

  /// Lets the harness end a response a script left open.
  Future<void> close() => _end();
}

typedef RunScript = Future<void> Function(RunPost post, RunResponse response);
typedef AttachScript =
    Future<void> Function(String agent, String thread, RunResponse response);

Map<String, Object?> runStarted(String thread, String run) => {
  'type': 'RUN_STARTED',
  'threadId': thread,
  'runId': run,
};
Map<String, Object?> runFinished(String thread, String run) => {
  'type': 'RUN_FINISHED',
  'threadId': thread,
  'runId': run,
  'outcome': {'type': 'success'},
};
Map<String, Object?> textStart(String id) => {
  'type': 'TEXT_MESSAGE_START',
  'messageId': id,
  'role': 'assistant',
};
Map<String, Object?> textDelta(String id, String delta) => {
  'type': 'TEXT_MESSAGE_CONTENT',
  'messageId': id,
  'delta': delta,
};
Map<String, Object?> textEnd(String id) => {
  'type': 'TEXT_MESSAGE_END',
  'messageId': id,
};

/// The AG-UI and session endpoints on top of the shared [TestServer]. Only
/// the external HTTP boundary is faked: what the server would answer, and
/// when.
class ChatServer {
  ChatServer(this.server) {
    server.authorizedReply = _handle;
  }
  final TestServer server;

  /// Sessions ListSessions pages through, newest first.
  final sessions = <SessionInfo>[];

  /// Thread reads by `agentId/threadId`; absent means an empty history.
  final histories = <String, String>{};

  final posts = <RunPost>[];
  final historyReads = <String>[];
  final stops = <String>[];
  final attaches = <String>[];
  AttachScript? attachScript;

  /// When set, Stop answers 202 with this body; otherwise 204.
  Map<String, Object?>? stopAccepted;
  final listCalls = <ListSessionsRequest>[];
  final listWorkspaces = <String?>[];
  final renames = <UpdateSessionTitleRequest>[];

  /// What a run request does. The default is a short, successful reply.
  RunScript? script;
  (int, String)? renameFailure;
  (int, String)? listFailure;
  Future<void> Function()? beforeList;

  Future<bool> _handle(HttpRequest r, List<int> bytes) async {
    final path = r.uri.path;
    final segments = r.uri.pathSegments;
    if (path.endsWith('/ListSessions')) {
      final input = ListSessionsRequest.fromBuffer(bytes);
      listCalls.add(input);
      listWorkspaces.add(r.headers.value('x-workspace-id'));
      await beforeList?.call();
      if (listFailure != null) {
        server.error(r, listFailure!.$1, listFailure!.$2);
        return true;
      }
      final size = input.pageSize == 0 ? 20 : input.pageSize;
      final start = input.pageToken.isEmpty ? 0 : int.parse(input.pageToken);
      final end = start + size > sessions.length
          ? sessions.length
          : start + size;
      server.write(
        r,
        ListSessionsResponse(
          sessions: sessions.sublist(start, end),
          nextPageToken: end < sessions.length ? '$end' : '',
          total: sessions.length,
        ),
      );
      return true;
    }
    if (path.endsWith('/UpdateSessionTitle')) {
      final input = UpdateSessionTitleRequest.fromBuffer(bytes);
      renames.add(input);
      if (renameFailure != null) {
        server.error(r, renameFailure!.$1, renameFailure!.$2);
        return true;
      }
      final index = sessions.indexWhere((s) => s.sessionId == input.sessionId);
      if (index < 0) {
        server.error(r, 404, 'not_found');
        return true;
      }
      sessions[index] = sessions[index].deepCopy()..title = input.title;
      server.write(r, UpdateSessionTitleResponse(session: sessions[index]));
      return true;
    }
    if (segments.length >= 3 && segments[0] == 'api' && segments[1] == 'agui') {
      // /api/agui/:agent or /api/agui/:agent/threads/:thread/{messages,stop}
      final agent = segments[2];
      if (segments.length == 3 && r.method == 'POST') {
        await _run(r, agent, bytes);
        return true;
      }
      if (segments.length == 6 && segments[3] == 'threads') {
        final thread = segments[4];
        if (segments[5] == 'messages' && r.method == 'GET') {
          historyReads.add('$agent/$thread');
          r.response.headers.contentType = ContentType.json;
          r.response.write(
            histories['$agent/$thread'] ??
                jsonEncode({'threadId': thread, 'messages': const []}),
          );
          return true;
        }
        if (segments[5] == 'stop' && r.method == 'POST') {
          stops.add('$agent/$thread');
          if (stopAccepted != null) {
            r.response.statusCode = 202;
            r.response.write(jsonEncode(stopAccepted));
          } else {
            r.response.statusCode = 204;
          }
          return true;
        }
        if (segments[5] == 'run' && r.method == 'GET') {
          attaches.add('$agent/$thread');
          if (attachScript == null) {
            r.response.statusCode = 204;
            return true;
          }
          final socket = await r.response.detachSocket(writeHeaders: false);
          final response = RunResponse(
            socket,
            RunPost(
              agentId: agent,
              token: r.headers.value('authorization'),
              workspace: r.headers.value('x-workspace-id'),
              contentType: null,
              body: {'threadId': thread, 'runId': 'attach'},
            ),
          );
          try {
            await attachScript!(agent, thread, response);
          } finally {
            await response.close();
          }
          return true;
        }
      }
    }
    return false;
  }

  Future<void> _run(HttpRequest r, String agent, List<int> bytes) async {
    final post = RunPost(
      agentId: agent,
      token: r.headers.value('authorization'),
      workspace: r.headers.value('x-workspace-id'),
      contentType: r.headers.contentType?.mimeType,
      body: jsonDecode(utf8.decode(bytes)) as Map<String, dynamic>,
    );
    posts.add(post);
    final socket = await r.response.detachSocket(writeHeaders: false);
    final response = RunResponse(socket, post);
    try {
      await (script ?? _defaultScript)(post, response);
    } finally {
      await response.close();
    }
  }

  static Future<void> _defaultScript(RunPost post, RunResponse r) async {
    await r.event(runStarted(post.threadId, post.runId));
    await r.event(textStart('m1'));
    await r.event(textDelta('m1', 'ok'));
    await r.event(textEnd('m1'));
    await r.event(runFinished(post.threadId, post.runId));
    await r.finish();
  }
}
