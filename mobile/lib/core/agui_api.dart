import 'dart:async';
import 'dart:convert';
import 'dart:io';
import 'dart:typed_data';

import 'server_address.dart';
import 'sse.dart';

/// A refused AG-UI request: the server answered with an HTTP status before any
/// stream opened, so the run did not start.
final class AguiHttpError implements Exception {
  const AguiHttpError(this.status, [this.code]);
  final int status;

  /// The `code` of the server's JSON error body, when it sent one.
  final String? code;
  @override
  String toString() => 'AguiHttpError($status${code == null ? '' : ', $code'})';
}

/// The request never left this device (no connection was made), so the server
/// cannot have seen it.
final class AguiNotSent implements Exception {
  const AguiNotSent();
  @override
  String toString() => 'AguiNotSent';
}

/// The request was written but no usable answer came back: the server may or
/// may not have started the run. Never resolved by sending again.
final class AguiAmbiguous implements Exception {
  const AguiAmbiguous();
  @override
  String toString() => 'AguiAmbiguous';
}

/// An open stream that went silent for longer than the idle timeout.
final class AguiIdle implements Exception {
  const AguiIdle();
  @override
  String toString() => 'AguiIdle';
}

/// An open run: its events, and the one way to end this device's connection
/// to it. Aborting closes the socket at once; the run itself carries on at
/// the server.
final class AguiRun {
  const AguiRun(this.events, this.abort);
  final Stream<SseEvent> events;
  final void Function() abort;
}

/// The run a Stop reached (202). It ends shortly after on the server.
final class AguiStoppedRun {
  const AguiStoppedRun({
    required this.threadId,
    required this.runId,
    required this.invocationId,
  });
  final String threadId;
  final String runId;
  final String invocationId;
}

/// Plain `dart:io` HTTP for the AG-UI endpoints. Redirects are never
/// followed (a redirect would carry the credential elsewhere), nothing is
/// retried, and a response body is only ever read within a bound.
final class AguiApi {
  AguiApi(this.address, {this.idleTimeout = const Duration(seconds: 60)});

  final ServerAddress address;

  /// How long an open stream may stay silent. The server sends a heartbeat
  /// comment while a run is quiet, so silence means the connection is gone.
  final Duration idleTimeout;

  HttpClient _client = _newClient();
  bool _closed = false;

  static HttpClient _newClient() =>
      HttpClient()..connectionTimeout = const Duration(seconds: 10);

  static const _headerTimeout = Duration(seconds: 30);
  static const _errorBodyLimit = 64 * 1024;
  static const _historyLimit = 64 * 1024 * 1024;

  Uri _run(String agentId) =>
      address.endpoint('api/agui/${Uri.encodeComponent(agentId)}');

  Uri _threadResource(String agentId, String threadId, String resource) =>
      address.endpoint(
        'api/agui/${Uri.encodeComponent(agentId)}/threads/'
        '${Uri.encodeComponent(threadId)}/$resource',
      );

  Uri _messages(String agentId, String threadId) =>
      _threadResource(agentId, threadId, 'messages');

  // The abort of every run still open, so a suspend can end them all.
  final _live = <void Function()>{};

  /// Posts a run and returns it once the server answered 200.
  ///
  /// Each run has its own connection, so ending it never touches another
  /// request. Throws [AguiNotSent] when nothing could be sent,
  /// [AguiHttpError] for a refusal, and [AguiAmbiguous] when the request went
  /// out but the answer was lost or unusable. The events' errors are the
  /// connection's own, or [AguiIdle].
  Future<AguiRun> run({
    required String token,
    required String workspace,
    required String agentId,
    required Uint8List body,
  }) async {
    if (_closed) throw const AguiNotSent();
    final client = _newClient();
    var aborted = false;
    late final void Function() abort;
    abort = () {
      aborted = true;
      _live.remove(abort);
      client.close(force: true);
    };
    _live.add(abort);
    final HttpClientRequest request;
    try {
      request = await client.postUrl(_run(agentId));
    } catch (_) {
      abort();
      throw const AguiNotSent();
    }
    final HttpClientResponse response;
    try {
      request.followRedirects = false;
      request.headers
        ..set('Authorization', 'Bearer $token')
        ..set('X-Workspace-ID', workspace)
        ..set('Accept', 'text/event-stream')
        ..contentType = ContentType('application', 'json', charset: 'utf-8')
        ..contentLength = body.length;
      request.add(body);
      response = await request.close().timeout(_headerTimeout);
    } catch (_) {
      abort();
      throw const AguiAmbiguous();
    }
    if (response.statusCode != 200) {
      final refusal = await _refusal(response);
      abort();
      throw refusal;
    }
    if (response.headers.contentType?.mimeType != 'text/event-stream') {
      abort();
      throw const AguiAmbiguous();
    }
    final bytes = response
        .timeout(
          idleTimeout,
          onTimeout: (sink) {
            sink.addError(const AguiIdle());
            sink.close();
          },
        )
        // What the close of an aborted run raises has no one left to hear it.
        .handleError((Object _) {}, test: (_) => aborted);
    return AguiRun(parseSse(bytes), abort);
  }

  Future<AguiHttpError> _refusal(HttpClientResponse response) async {
    String? code;
    try {
      final bytes = await _readBounded(
        response,
        _errorBodyLimit,
      ).timeout(const Duration(seconds: 10));
      final decoded = jsonDecode(utf8.decode(bytes));
      if (decoded is Map && decoded['code'] is String) {
        code = decoded['code'] as String;
      }
    } catch (_) {
      // The status alone decides what the person is told.
    }
    return AguiHttpError(response.statusCode, code);
  }

  /// Reads at most [limit] bytes; throws when the body is larger.
  Future<List<int>> _readBounded(Stream<List<int>> body, int limit) async {
    final out = BytesBuilder(copy: false);
    await for (final chunk in body) {
      out.add(chunk);
      if (out.length > limit) {
        throw const FormatException('The server sent too much.');
      }
    }
    return out.takeBytes();
  }

  /// Follows the thread's detached run log (GET …/run). Null on 204.
  Future<AguiRun?> attach({
    required String token,
    required String workspace,
    required String agentId,
    required String threadId,
  }) async {
    if (_closed) throw const AguiNotSent();
    final client = _newClient();
    var aborted = false;
    late final void Function() abort;
    abort = () {
      aborted = true;
      _live.remove(abort);
      client.close(force: true);
    };
    _live.add(abort);
    final HttpClientRequest request;
    try {
      request = await client.getUrl(_threadResource(agentId, threadId, 'run'));
    } catch (_) {
      abort();
      throw const AguiNotSent();
    }
    final HttpClientResponse response;
    try {
      request.followRedirects = false;
      request.headers
        ..set('Authorization', 'Bearer $token')
        ..set('X-Workspace-ID', workspace)
        ..set('Accept', 'text/event-stream');
      response = await request.close().timeout(_headerTimeout);
    } catch (_) {
      abort();
      throw const AguiAmbiguous();
    }
    if (response.statusCode == 204) {
      abort();
      return null;
    }
    if (response.statusCode != 200) {
      final refusal = await _refusal(response);
      abort();
      throw refusal;
    }
    if (response.headers.contentType?.mimeType != 'text/event-stream') {
      abort();
      throw const AguiAmbiguous();
    }
    final bytes = response
        .timeout(
          idleTimeout,
          onTimeout: (sink) {
            sink.addError(const AguiIdle());
            sink.close();
          },
        )
        .handleError((Object _) {}, test: (_) => aborted);
    return AguiRun(parseSse(bytes), abort);
  }

  /// Stops the thread's detached run. Null on 204 (nothing to stop).
  Future<AguiStoppedRun?> stop({
    required String token,
    required String workspace,
    required String agentId,
    required String threadId,
  }) async {
    if (_closed) throw const AguiNotSent();
    final request = await _client
        .postUrl(_threadResource(agentId, threadId, 'stop'))
        .timeout(const Duration(seconds: 20));
    request.followRedirects = false;
    request.headers
      ..set('Authorization', 'Bearer $token')
      ..set('X-Workspace-ID', workspace)
      ..contentLength = 0;
    final response = await request.close().timeout(const Duration(seconds: 20));
    if (response.statusCode == 204) return null;
    if (response.statusCode != 202) throw await _refusal(response);
    final bytes = await _readBounded(
      response,
      _errorBodyLimit,
    ).timeout(const Duration(seconds: 10));
    final decoded = jsonDecode(utf8.decode(bytes));
    if (decoded is! Map) throw const FormatException('Unreadable stop answer.');
    final runId = decoded['runId'];
    final invocationId = decoded['invocationId'];
    final tid = decoded['threadId'];
    if (runId is! String || invocationId is! String || tid is! String) {
      throw const FormatException('Unreadable stop answer.');
    }
    return AguiStoppedRun(
      threadId: tid,
      runId: runId,
      invocationId: invocationId,
    );
  }

  /// A thread's history as the server's JSON text.
  Future<String> history({
    required String token,
    required String workspace,
    required String agentId,
    required String threadId,
  }) async {
    if (_closed) throw const AguiNotSent();
    final request = await _client
        .getUrl(_messages(agentId, threadId))
        .timeout(const Duration(seconds: 20));
    request.followRedirects = false;
    request.headers
      ..set('Authorization', 'Bearer $token')
      ..set('X-Workspace-ID', workspace)
      ..set('Accept', 'application/json');
    final response = await request.close().timeout(const Duration(seconds: 20));
    if (response.statusCode != 200) throw await _refusal(response);
    final bytes = await _readBounded(
      response,
      _historyLimit,
    ).timeout(const Duration(seconds: 60));
    return utf8.decode(bytes);
  }

  /// Aborts everything in flight (streams, the request being sent, reads)
  /// and stays usable: the next request opens a fresh connection.
  void abortAll() {
    if (_closed) return;
    for (final abort in _live.toList()) {
      abort();
    }
    final old = _client;
    _client = _newClient();
    old.close(force: true);
  }

  void close() {
    _closed = true;
    for (final abort in _live.toList()) {
      abort();
    }
    _client.close(force: true);
  }
}
