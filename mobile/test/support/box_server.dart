import 'dart:async';
import 'dart:convert';
import 'dart:io';

import 'package:butter_mobile/gen/agents/v1/butterbox.pb.dart';

import 'agent_server.dart';

/// External HTTP boundary for ButterBoxService (on top of AgentService), with
/// controllable late, refused and lost replies. The real semantics (write-only
/// token, reference-blocked delete, catalog errors) are proven against the
/// production Go handlers in box_contract_test; this only drives UI timing and
/// failure shapes. The phone never reaches a Box: only these Butter calls exist.
class BoxServer extends AgentServer {
  final boxes = <String, Map<String, ButterBox>>{'alpha': {}, 'beta': {}};

  /// The access token the server holds per Workspace and box. It is never in a
  /// reply, so a test that finds it on screen found a leak.
  final secrets = <String, Map<String, String>>{'alpha': {}, 'beta': {}};
  final boxCalls = <({String method, String? workspace, List<int> body})>[];
  int _next = 0;

  static const writeMethods = {
    'CreateButterBox',
    'UpdateButterBox',
    'DeleteButterBox',
    'SetButterBoxToken',
  };
  int get boxWrites =>
      boxCalls.where((c) => writeMethods.contains(c.method)).length;

  Future<void> Function()? beforeList;
  Future<void> Function()? beforeGetBox;
  Future<void> Function()? beforeStatus;
  Future<void> Function()? beforeCatalog;

  /// A write answered with this (status, code, message) instead of applied.
  (int, String, String)? boxRefusal;

  /// A write that is applied but whose answer is lost (503).
  bool loseBoxReply = false;

  /// Answered for GetButterBoxStatus.
  GetButterBoxStatusResponse status = GetButterBoxStatusResponse(
    reachable: true,
    activeSessions: 2,
  );
  (int, String, String)? statusRefusal;
  List<ButterBoxModel> piModels = [];
  (int, String, String)? piRefusal;
  List<ButterBoxCursorModel> cursorModels = [];
  (int, String, String)? cursorRefusal;

  /// Ids that agents reference; deleting them is refused like production.
  final referencedBy = <String, List<String>>{};

  BoxServer() {
    final inherited = authorizedReply!;
    authorizedReply = (r, bytes) async {
      if (!r.uri.path.contains('ButterBoxService')) {
        return inherited(r, bytes);
      }
      final method = r.uri.path.split('/').last;
      final ws = r.headers.value('x-workspace-id');
      boxCalls.add((method: method, workspace: ws, body: bytes));
      final bucket = boxes[ws];
      if (bucket == null) {
        _fail(r, 403, 'permission_denied', 'workspace access denied');
        return true;
      }
      if (writeMethods.contains(method)) {
        if (boxRefusal case final f?) {
          _fail(r, f.$1, f.$2, f.$3);
          return true;
        }
      }
      ButterBox? find(String id) => bucket[id];
      switch (method) {
        case 'ListButterBoxes':
          // The snapshot is taken when the request arrives, so a delayed
          // answer is a stale one, as when a read overlaps a write.
          final snapshot = [for (final b in bucket.values) b.deepCopy()];
          await beforeList?.call();
          write(r, ListButterBoxesResponse(butterBoxes: snapshot));
        case 'GetButterBox':
          final id = GetButterBoxRequest.fromBuffer(bytes).id;
          final found = find(id)?.deepCopy();
          await beforeGetBox?.call();
          if (found == null) {
            _fail(r, 404, 'not_found', 'butter box not found');
          } else {
            write(r, GetButterBoxResponse(butterBox: found));
          }
        case 'CreateButterBox':
          final req = CreateButterBoxRequest.fromBuffer(bytes);
          if (bucket.values.any((b) => b.name == req.name)) {
            _fail(r, 409, 'already_exists', 'butter box already exists');
            return true;
          }
          final box = ButterBox(
            id: 'box-${++_next}',
            name: req.name,
            baseUrl: req.baseUrl,
            enabled: req.enabled,
            credentialSet: req.token.isNotEmpty,
            workspaceId: ws,
          );
          bucket[box.id] = box;
          if (req.token.isNotEmpty) secrets[ws]![box.id] = req.token;
          _answer(r, CreateButterBoxResponse(butterBox: box));
        case 'UpdateButterBox':
          final req = UpdateButterBoxRequest.fromBuffer(bytes);
          final found = find(req.id);
          if (found == null) {
            _fail(r, 404, 'not_found', 'butter box not found');
            return true;
          }
          final stored = found.deepCopy()
            ..name = req.name
            ..baseUrl = req.baseUrl
            ..enabled = req.enabled;
          bucket[req.id] = stored;
          _answer(r, UpdateButterBoxResponse(butterBox: stored));
        case 'SetButterBoxToken':
          final req = SetButterBoxTokenRequest.fromBuffer(bytes);
          final found = find(req.id);
          if (found == null) {
            _fail(r, 404, 'not_found', 'butter box not found');
            return true;
          }
          final stored = found.deepCopy()..credentialSet = req.token.isNotEmpty;
          bucket[req.id] = stored;
          if (req.token.isEmpty) {
            secrets[ws]!.remove(req.id);
          } else {
            secrets[ws]![req.id] = req.token;
          }
          _answer(r, SetButterBoxTokenResponse(butterBox: stored));
        case 'DeleteButterBox':
          final id = DeleteButterBoxRequest.fromBuffer(bytes).id;
          if (find(id) == null) {
            _fail(r, 404, 'not_found', 'butter box not found');
            return true;
          }
          final refs = referencedBy[id];
          if (refs != null && refs.isNotEmpty) {
            _fail(
              r,
              400,
              'failed_precondition',
              'butter box "$id" is referenced by agents: ${refs.join(', ')}; update or delete those agents first',
            );
            return true;
          }
          bucket.remove(id);
          secrets[ws]!.remove(id);
          _answer(r, DeleteButterBoxResponse());
        case 'GetButterBoxStatus':
          await beforeStatus?.call();
          if (statusRefusal case final f?) {
            _fail(r, f.$1, f.$2, f.$3);
          } else {
            write(r, status);
          }
        case 'ListButterBoxModels':
          await beforeCatalog?.call();
          if (piRefusal case final f?) {
            _fail(r, f.$1, f.$2, f.$3);
          } else {
            write(r, ListButterBoxModelsResponse(models: piModels));
          }
        case 'ListButterBoxCursorModels':
          await beforeCatalog?.call();
          if (cursorRefusal case final f?) {
            _fail(r, f.$1, f.$2, f.$3);
          } else {
            write(r, ListButterBoxCursorModelsResponse(models: cursorModels));
          }
        default:
          return false;
      }
      return true;
    };
  }

  /// Adds an already registered box without going through a call.
  ButterBox seed(
    String ws,
    String id,
    String name, {
    bool enabled = true,
    bool credentialSet = true,
  }) {
    final box = ButterBox(
      id: id,
      name: name,
      baseUrl: 'https://$id.box.test',
      enabled: enabled,
      credentialSet: credentialSet,
      workspaceId: ws,
    );
    boxes[ws]![id] = box;
    if (credentialSet) secrets[ws]![id] = 'secret-of-$id';
    return box;
  }

  void _answer(HttpRequest r, dynamic message) {
    if (loseBoxReply) {
      _fail(r, 503, 'unavailable', 'connection lost');
    } else {
      write(r, message);
    }
  }

  void _fail(HttpRequest r, int status, String code, String message) {
    r.response.statusCode = status;
    r.response.headers.contentType = ContentType.json;
    r.response.write(jsonEncode({'code': code, 'message': message}));
  }
}
