import 'dart:async';
import 'dart:convert';
import 'dart:io';

import 'package:butter_mobile/gen/agents/v1/api_token.pb.dart';
import 'package:butter_mobile/gen/agents/v1/dashboard.pb.dart';
import 'package:fixnum/fixnum.dart';
import 'package:protobuf/well_known_types/google/protobuf/timestamp.pb.dart';

import 'mcp_server.dart';

/// External HTTP boundary for DaemonService and APITokenService with
/// controllable late, refused and lost replies. The real semantics (tenant
/// isolation, worker tokens, dispatch, status, cancel) are proven against the
/// production Go handlers and a fake worker in daemon_contract_test; this only
/// drives UI timing and failure shapes. It extends the Remote Agent fake so one
/// server serves both.
class DaemonServerFake extends McpServerFake {
  final runtimes = <String, Map<String, DaemonRuntime>>{
    'alpha': {},
    'beta': {},
  };
  final onlineWorkers = <String, List<DaemonStatus>>{'alpha': [], 'beta': []};
  final taskList = <String, List<DaemonTaskInFlight>>{'alpha': [], 'beta': []};
  final tokens = <String, List<APIToken>>{'alpha': [], 'beta': []};
  final daemonCalls = <({String method, String? workspace, List<int> body})>[];

  BridgeDiagnostics diagnostics = BridgeDiagnostics(
    cpuPercent: 12.5,
    memoryUsedBytes: Int64(256 * 1024 * 1024),
    memoryLimitBytes: Int64(1024 * 1024 * 1024),
    goroutines: 42,
    checkedAt: Timestamp(seconds: Int64(1760000000)),
  );

  static const daemonWriteMethods = {
    'CreateDaemonRuntime',
    'UpdateDaemonRuntime',
    'DeleteDaemonRuntime',
    'CreateDaemonRuntimeToken',
    'CancelDaemonTask',
    'RevokeAPIToken',
  };
  int daemonCount(String method) =>
      daemonCalls.where((c) => c.method == method).length;

  Future<void> Function(String method)? daemonBefore;
  final daemonRefusals = <String, (int, String, String)>{};

  /// Writes that are applied but whose answer is lost (503).
  final lostReplies = <String>{};

  int _minted = 0;
  final secrets = <String>[];

  DaemonServerFake() {
    final inherited = authorizedReply!;
    authorizedReply = (r, bytes) async {
      final path = r.uri.path;
      if (!path.contains('DaemonService') &&
          !path.contains('APITokenService')) {
        return inherited(r, bytes);
      }
      final method = path.split('/').last;
      final ws = r.headers.value('x-workspace-id');
      daemonCalls.add((method: method, workspace: ws, body: bytes));
      final bucket = runtimes[ws];
      if (bucket == null) return _fail(r, 403, 'permission_denied', 'denied');
      final refused = daemonRefusals[method];
      if (refused != null) return _fail(r, refused.$1, refused.$2, refused.$3);
      // Reads capture their answer when the request arrives, so a delayed
      // answer is a stale one.
      switch (method) {
        case 'ListDaemonRuntimes':
          final snapshot = [for (final x in bucket.values) x.deepCopy()];
          await daemonBefore?.call(method);
          write(r, ListDaemonRuntimesResponse(runtimes: snapshot));
        case 'GetDaemonRuntime':
          final id = GetDaemonRuntimeRequest.fromBuffer(bytes).id;
          final found = bucket[id]?.deepCopy();
          await daemonBefore?.call(method);
          if (found == null) return _fail(r, 404, 'not_found', 'not found');
          write(r, GetDaemonRuntimeResponse(runtime: found));
        case 'CreateDaemonRuntime':
          await daemonBefore?.call(method);
          final x = CreateDaemonRuntimeRequest.fromBuffer(bytes).runtime;
          if (bucket.containsKey(x.id)) {
            return _fail(r, 409, 'already_exists', 'exists');
          }
          final stored = x.deepCopy()..workspaceId = ws!;
          bucket[x.id] = stored;
          _answer(method, r, CreateDaemonRuntimeResponse(runtime: stored));
        case 'UpdateDaemonRuntime':
          await daemonBefore?.call(method);
          final x = UpdateDaemonRuntimeRequest.fromBuffer(bytes).runtime;
          if (!bucket.containsKey(x.id)) {
            return _fail(r, 404, 'not_found', 'not found');
          }
          bucket[x.id] = x.deepCopy();
          _answer(method, r, UpdateDaemonRuntimeResponse(runtime: x));
        case 'DeleteDaemonRuntime':
          await daemonBefore?.call(method);
          final id = DeleteDaemonRuntimeRequest.fromBuffer(bytes).id;
          if (bucket.remove(id) == null) {
            return _fail(r, 404, 'not_found', 'not found');
          }
          _answer(method, r, DeleteDaemonRuntimeResponse());
        case 'ListDaemons':
          final snapshot = [for (final d in onlineWorkers[ws]!) d.deepCopy()];
          await daemonBefore?.call(method);
          write(r, ListDaemonsResponse(daemons: snapshot));
        case 'ListDaemonTasks':
          final snapshot = [for (final t in taskList[ws]!) t.deepCopy()];
          await daemonBefore?.call(method);
          write(r, ListDaemonTasksResponse(tasks: snapshot));
        case 'GetBridgeDiagnostics':
          await daemonBefore?.call(method);
          write(r, GetBridgeDiagnosticsResponse(diagnostics: diagnostics));
        case 'CancelDaemonTask':
          await daemonBefore?.call(method);
          final req = CancelDaemonTaskRequest.fromBuffer(bytes);
          final task = taskList[ws]!
              .where((t) => t.taskId == req.taskId)
              .firstOrNull;
          if (task == null) return _fail(r, 404, 'not_found', 'not found');
          _answer(
            method,
            r,
            CancelDaemonTaskResponse(daemonRuntimeId: task.daemonRuntimeId),
          );
        case 'CreateDaemonRuntimeToken':
          await daemonBefore?.call(method);
          final req = CreateDaemonRuntimeTokenRequest.fromBuffer(bytes);
          if (!bucket.containsKey(req.daemonRuntimeId)) {
            return _fail(r, 404, 'not_found', 'runtime not found');
          }
          final n = ++_minted;
          final secret = 'worker-secret-$n';
          secrets.add(secret);
          final token = APIToken(
            id: 'tok-$n',
            name: req.name,
            prefix: 'btr_w$n',
            kind: APITokenKind.API_TOKEN_KIND_DAEMON,
            scopes: ['daemon:connect'],
            daemonRuntimeId: req.daemonRuntimeId,
            workspaceId: ws,
          );
          tokens[ws]!.add(token);
          _answer(
            method,
            r,
            CreateDaemonRuntimeTokenResponse(token: token, secret: secret),
          );
        case 'ListAPITokens':
          final snapshot = [for (final t in tokens[ws]!) t.deepCopy()];
          await daemonBefore?.call(method);
          write(r, ListAPITokensResponse(tokens: snapshot));
        case 'RevokeAPIToken':
          await daemonBefore?.call(method);
          final id = RevokeAPITokenRequest.fromBuffer(bytes).id;
          final found = tokens[ws]!.where((t) => t.id == id).firstOrNull;
          if (found == null) return _fail(r, 404, 'not_found', 'not found');
          found.revoked = true;
          _answer(method, r, RevokeAPITokenResponse());
        default:
          return false;
      }
      return true;
    };
  }

  DaemonRuntime seedRuntime(String ws, DaemonRuntime runtime) {
    runtimes[ws]![runtime.id] = runtime..workspaceId = ws;
    return runtime;
  }

  void _answer(String method, HttpRequest r, dynamic message) {
    if (lostReplies.contains(method)) {
      _fail(r, 503, 'unavailable', 'connection lost');
    } else {
      write(r, message);
    }
  }

  bool _fail(HttpRequest r, int status, String code, String message) {
    r.response.statusCode = status;
    r.response.headers.contentType = ContentType.json;
    r.response.write(jsonEncode({'code': code, 'message': message}));
    return true;
  }
}
