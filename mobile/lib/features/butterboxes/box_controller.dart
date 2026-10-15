import 'dart:async';

import 'package:connectrpc/connect.dart';
import 'package:flutter/foundation.dart';

import '../../core/box_errors.dart';
import '../../core/butter_api.dart';
import '../../gen/agents/v1/butterbox.pb.dart';

enum BoxLoad { idle, loading, ready, failed }

enum BoxProbeState { idle, probing, reachable, unreachable, failed }

/// The outcome of one explicit connection test.
class BoxProbe {
  const BoxProbe(this.state, {this.activeSessions = 0, this.detail = ''});
  final BoxProbeState state;
  final int activeSessions;

  /// What the probe said (an unreachable box's reason) or a safe failure
  /// message. Bounded; never a credential.
  final String detail;
}

/// A model catalog the box reported, or why it could not.
class BoxCatalog<T> {
  const BoxCatalog(this.state, {this.models = const [], this.error});
  final BoxLoad state;
  final List<T> models;
  final String? error;
}

const _unconfirmed =
    'Change not confirmed; it may already have been applied. Refresh before retrying.';

String _bounded(String text) =>
    text.length > 300 ? '${text.substring(0, 300)}…' : text;

/// ButterBoxes of one server, signed-in user and Workspace.
///
/// Its lifetime is that scope: the owner disposes it when the server, account
/// or Workspace changes, which closes its network use and drops everything it
/// holds, so no late response can reach another context. It is never reused.
/// The phone never talks to a Box: every call goes to the Butter server, which
/// holds the Box credential and does the probing.
class ButterBoxController extends ChangeNotifier {
  ButterBoxController({
    required ButterApi Function() newApi,
    required this.token,
    required this.workspaceId,
    required this.onUnauthenticated,
  }) : _newApi = newApi;

  final ButterApi Function() _newApi;
  final String token;
  final String workspaceId;

  /// Called when a request is refused as unauthenticated; the owner signs the
  /// person out. Nothing is retried.
  final void Function() onUnauthenticated;

  ButterApi? _apiOrNull;
  ButterApi get _api => _apiOrNull ??= _newApi();
  bool _disposed = false;
  bool get disposed => _disposed;
  int _epoch = 0;
  bool _live(int epoch) => !_disposed && epoch == _epoch;
  void _emit() {
    if (!_disposed) notifyListeners();
  }

  // ---- List and detail ----

  List<ButterBox> boxes = const [];
  BoxLoad listState = BoxLoad.idle;
  String? listError;
  int _listSeq = 0;

  ButterBox? box;
  BoxLoad boxState = BoxLoad.idle;
  String? boxError;
  int _boxSeq = 0;

  ButterBox? byId(String id) {
    for (final b in boxes) {
      if (b.id == id) return b;
    }
    return null;
  }

  bool _expired(Object failure) {
    final expired =
        failure is ConnectException && failure.code == Code.unauthenticated;
    if (expired && !_disposed) onUnauthenticated();
    return expired;
  }

  Future<void> loadBoxes() async {
    if (_disposed) return;
    final epoch = _epoch, seq = ++_listSeq;
    listState = BoxLoad.loading;
    listError = null;
    _emit();
    try {
      final list = await _api.butterBoxes(token, workspaceId);
      if (!_live(epoch) || seq != _listSeq) return;
      boxes = List.unmodifiable(list);
      listState = BoxLoad.ready;
      // A successful authoritative read is the reconciliation step.
      unconfirmed = false;
    } catch (failure) {
      if (!_live(epoch) || seq != _listSeq || _expired(failure)) return;
      listState = BoxLoad.failed;
      listError = boxFailureMessage(failure, BoxOperation.load);
    }
    _emit();
  }

  /// Reads one box fresh; an edit starts from this copy, not a list row.
  Future<void> loadBox(String id) async {
    if (_disposed) return;
    final epoch = _epoch, seq = ++_boxSeq;
    if (box?.id != id) box = null;
    boxState = BoxLoad.loading;
    boxError = null;
    _emit();
    try {
      final result = await _api.butterBox(token, workspaceId, id);
      if (!_live(epoch) || seq != _boxSeq) return;
      box = result;
      boxState = BoxLoad.ready;
      unconfirmed = false;
      _replaceInList(result);
    } catch (failure) {
      if (!_live(epoch) || seq != _boxSeq || _expired(failure)) return;
      box = null;
      boxState = BoxLoad.failed;
      boxError = boxFailureMessage(failure, BoxOperation.load);
    }
    _emit();
  }

  void _replaceInList(ButterBox result) {
    if (boxes.any((b) => b.id == result.id)) {
      boxes = List.unmodifiable([
        for (final b in boxes) b.id == result.id ? result : b,
      ]);
    }
  }

  // ---- Connection test and model catalogs (read-only, never blocking) ----

  final Map<String, BoxProbe> probes = {};
  final Map<String, int> _probeSeq = {};

  /// Asks the server to probe the box now. An unreachable box is an answer.
  Future<void> probe(String id) async {
    if (_disposed) return;
    final epoch = _epoch, seq = (_probeSeq[id] ?? 0) + 1;
    _probeSeq[id] = seq;
    probes[id] = const BoxProbe(BoxProbeState.probing);
    _emit();
    try {
      final reply = await _api.butterBoxStatus(token, workspaceId, id);
      if (!_live(epoch) || _probeSeq[id] != seq) return;
      probes[id] = reply.reachable
          ? BoxProbe(
              BoxProbeState.reachable,
              activeSessions: reply.activeSessions,
            )
          : BoxProbe(BoxProbeState.unreachable, detail: _bounded(reply.error));
    } catch (failure) {
      if (!_live(epoch) || _probeSeq[id] != seq || _expired(failure)) return;
      probes[id] = BoxProbe(
        BoxProbeState.failed,
        detail: boxFailureMessage(failure, BoxOperation.status),
      );
    }
    _emit();
  }

  final Map<String, BoxCatalog<ButterBoxModel>> piCatalogs = {};
  final Map<String, BoxCatalog<ButterBoxCursorModel>> cursorCatalogs = {};
  final Map<String, int> _catalogSeq = {};

  Future<void> loadPiModels(String id) => _catalog(
    'pi:$id',
    piCatalogs,
    id,
    () => _api.butterBoxModels(token, workspaceId, id),
    BoxOperation.piModels,
  );

  Future<void> loadCursorModels(String id) => _catalog(
    'cursor:$id',
    cursorCatalogs,
    id,
    () => _api.butterBoxCursorModels(token, workspaceId, id),
    BoxOperation.cursorModels,
  );

  Future<void> _catalog<T>(
    String key,
    Map<String, BoxCatalog<T>> into,
    String id,
    Future<List<T>> Function() read,
    BoxOperation operation,
  ) async {
    if (_disposed || id.isEmpty) return;
    final epoch = _epoch, seq = (_catalogSeq[key] ?? 0) + 1;
    _catalogSeq[key] = seq;
    into[id] = const BoxCatalog(BoxLoad.loading);
    _emit();
    try {
      final models = await read();
      if (!_live(epoch) || _catalogSeq[key] != seq) return;
      into[id] = BoxCatalog(BoxLoad.ready, models: List.unmodifiable(models));
    } catch (failure) {
      if (!_live(epoch) || _catalogSeq[key] != seq || _expired(failure)) {
        return;
      }
      into[id] = BoxCatalog(
        BoxLoad.failed,
        error: boxFailureMessage(failure, operation),
      );
    }
    _emit();
  }

  // ---- Writes: one explicit call each, never repeated on their own ----

  bool writing = false;
  String? writeError;

  /// A write whose outcome is unknown. Another write is refused until a read
  /// reconciles it.
  bool unconfirmed = false;
  String? get writeNotice => unconfirmed ? _unconfirmed : null;

  Future<bool> _write(
    Future<void> Function(int epoch) action,
    BoxOperation operation,
  ) async {
    if (_disposed || writing || unconfirmed) return false;
    final epoch = _epoch;
    writing = true;
    writeError = null;
    _emit();
    var ok = false;
    try {
      await action(epoch);
      ok = _live(epoch);
    } catch (failure) {
      if (!_live(epoch) || _expired(failure)) return false;
      writeError = boxFailureMessage(failure, operation);
      if (boxWriteOutcomeUnknown(failure)) unconfirmed = true;
    } finally {
      if (_live(epoch)) {
        writing = false;
        _emit();
      }
    }
    return ok;
  }

  void _stored(ButterBox saved) {
    box = saved;
    boxState = BoxLoad.ready;
    boxes = List.unmodifiable([
      for (final b in boxes) b.id == saved.id ? saved : b,
      if (!boxes.any((b) => b.id == saved.id)) saved,
    ]);
  }

  /// [accessToken] is write-only; it is sent once and never kept.
  Future<bool> create({
    required String name,
    required String baseUrl,
    required bool enabled,
    String accessToken = '',
  }) => _write((epoch) async {
    final saved = await _api.createButterBox(
      token,
      workspaceId,
      name: name.trim(),
      baseUrl: baseUrl.trim(),
      enabled: enabled,
      accessToken: accessToken,
    );
    if (_live(epoch)) _stored(saved);
  }, BoxOperation.create);

  Future<bool> update({
    required String id,
    required String name,
    required String baseUrl,
    required bool enabled,
  }) => _write((epoch) async {
    final saved = await _api.updateButterBox(
      token,
      workspaceId,
      id: id,
      name: name.trim(),
      baseUrl: baseUrl.trim(),
      enabled: enabled,
    );
    if (_live(epoch)) _stored(saved);
  }, BoxOperation.update);

  /// Replaces the stored access token. An empty value is refused here: clearing
  /// is the separate [clearToken], so a blank field can never erase a token.
  Future<bool> replaceToken(String id, String accessToken) {
    if (accessToken.trim().isEmpty) {
      writeError = 'Enter the new access token.';
      _emit();
      return Future.value(false);
    }
    return _write((epoch) async {
      final saved = await _api.setButterBoxToken(
        token,
        workspaceId,
        id,
        accessToken.trim(),
      );
      if (_live(epoch)) {
        _stored(saved);
        probes.remove(id);
      }
    }, BoxOperation.token);
  }

  Future<bool> clearToken(String id) => _write((epoch) async {
    final saved = await _api.setButterBoxToken(token, workspaceId, id, '');
    if (_live(epoch)) {
      _stored(saved);
      probes.remove(id);
    }
  }, BoxOperation.token);

  Future<bool> delete(String id) => _write((epoch) async {
    await _api.deleteButterBox(token, workspaceId, id);
    if (!_live(epoch)) return;
    boxes = List.unmodifiable(boxes.where((b) => b.id != id));
    if (box?.id == id) box = null;
    probes.remove(id);
    piCatalogs.remove(id);
    cursorCatalogs.remove(id);
  }, BoxOperation.delete);

  /// The app is going to the background: in-flight work is closed and fenced,
  /// and a write that was in flight is unconfirmed, never repeated.
  void suspend() {
    if (_disposed) return;
    _epoch++;
    _apiOrNull?.close();
    _apiOrNull = null;
    if (writing) unconfirmed = true;
    writing = false;
    if (listState == BoxLoad.loading) listState = BoxLoad.idle;
    if (boxState == BoxLoad.loading) boxState = BoxLoad.idle;
    probes.removeWhere((_, p) => p.state == BoxProbeState.probing);
    piCatalogs.removeWhere((_, c) => c.state == BoxLoad.loading);
    cursorCatalogs.removeWhere((_, c) => c.state == BoxLoad.loading);
    _emit();
  }

  @override
  void dispose() {
    if (_disposed) return;
    _disposed = true;
    _epoch++;
    boxes = const [];
    box = null;
    probes.clear();
    piCatalogs.clear();
    cursorCatalogs.clear();
    _apiOrNull?.close();
    _apiOrNull = null;
    super.dispose();
  }
}
