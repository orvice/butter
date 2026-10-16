import 'dart:async';

import 'package:connectrpc/connect.dart';
import 'package:flutter/foundation.dart';
import 'package:protobuf/protobuf.dart' show GeneratedMessage;

import 'butter_api.dart';
import 'config_errors.dart';

enum ConfigLoad { idle, loading, ready, failed }

const configUnconfirmed =
    'Change not confirmed; it may already have been applied. Refresh before retrying.';

/// A Workspace-scoped configuration resource that is listed, read, created,
/// updated and deleted one explicit call at a time (MCP servers, Remote
/// Agents).
///
/// Its lifetime is one server, signed-in user and Workspace: the owner
/// disposes it when any of them changes, which closes its network use and
/// drops everything it holds, so no late response can reach another context.
///
/// Reads are ordered against writes. A read that began before a write ended
/// may carry the state from before it, so its answer is dropped and it can
/// never settle that write either. A write whose outcome is unknown (the
/// connection broke, the app went to the background) is never repeated; every
/// further write is refused until a later read that covers the same object
/// has seen the server's state.
abstract class ConfigResourceController<T extends GeneratedMessage>
    extends ChangeNotifier {
  ConfigResourceController({
    required ButterApi Function() newApi,
    required this.token,
    required this.workspaceId,
    required this.onUnauthenticated,
    required this.kind,
  }) : _newApi = newApi;

  final ButterApi Function() _newApi;
  final String token;
  final String workspaceId;
  final ConfigKind kind;

  /// Called when a request is refused as unauthenticated; the owner signs the
  /// person out. Nothing is retried.
  final void Function() onUnauthenticated;

  // ---- What a resource type provides ----

  @protected
  String idOf(T item);
  @protected
  Future<List<T>> readAll(ButterApi api);
  @protected
  Future<T> readOne(ButterApi api, String id);
  @protected
  Future<T> createOne(ButterApi api, T draft);
  @protected
  Future<T> updateOne(ButterApi api, T draft);
  @protected
  Future<void> deleteOne(ButterApi api, String id);

  /// A write aimed at [id] (null: none known) starts or has just ended;
  /// anything read for that object before now is obsolete.
  @protected
  void objectChanging(String? id) {}

  /// The object is gone; drop whatever is cached for it.
  @protected
  void objectRemoved(String id) {}

  @protected
  void onSuspend() {}

  /// The app is back in the foreground: re-read what [onSuspend] dropped.
  @protected
  Future<void> onRecover() async {}
  @protected
  void onDispose() {}

  // ---- Lifetime ----

  ButterApi? _apiOrNull;
  @protected
  ButterApi get api => _apiOrNull ??= _newApi();
  bool _disposed = false;
  bool get disposed => _disposed;
  int _epoch = 0;
  @protected
  int get epoch => _epoch;
  @protected
  bool live(int epoch) => !_disposed && epoch == _epoch;
  @protected
  void emit() {
    if (!_disposed) notifyListeners();
  }

  @protected
  bool expired(Object failure) {
    final isExpired =
        failure is ConnectException && failure.code == Code.unauthenticated;
    if (isExpired && !_disposed) onUnauthenticated();
    return isExpired;
  }

  // Ordering of reads against writes. Every read takes a tick when it starts;
  // every write takes one when it starts and another when it ends.
  int _clock = 0;
  int _lastWriteTick = 0;
  bool _fresh(int startTick) => startTick > _lastWriteTick;

  /// Whether a read that began at [startTick] began after the last write
  /// started or ended. An older read may carry the state from before it.
  @protected
  bool readIsFresh(int startTick) => _fresh(startTick);

  // ---- List and detail ----

  List<T> items = const [];
  ConfigLoad listState = ConfigLoad.idle;
  String? listError;
  int _listSeq = 0;

  T? item;
  ConfigLoad itemState = ConfigLoad.idle;
  String? itemError;
  int _itemSeq = 0;
  String? _openId;
  bool _reloadList = false;

  T? byId(String id) {
    for (final i in items) {
      if (idOf(i) == id) return i;
    }
    return null;
  }

  Future<void> load() async {
    if (_disposed) return;
    final epoch = _epoch, seq = ++_listSeq, started = ++_clock;
    final before = listState == ConfigLoad.loading
        ? ConfigLoad.idle
        : listState;
    listState = ConfigLoad.loading;
    listError = null;
    emit();
    try {
      final list = await readAll(api);
      if (!live(epoch) || seq != _listSeq) return;
      if (!_fresh(started)) {
        listState = before;
        emit();
        return;
      }
      items = List.unmodifiable(list);
      listState = ConfigLoad.ready;
      reconcile(started);
    } catch (failure) {
      if (!live(epoch) || seq != _listSeq) return;
      if (!_fresh(started)) {
        // An obsolete failed read says nothing about the present, nor does an
        // authentication failure it carries.
        listState = before;
        emit();
        return;
      }
      if (expired(failure)) return;
      listState = ConfigLoad.failed;
      listError = configFailureMessage(failure, kind, ConfigOperation.load);
    }
    emit();
  }

  /// Reads one object fresh; an edit starts from this copy, not a list row.
  Future<void> loadItem(String id) async {
    if (_disposed) return;
    final epoch = _epoch, seq = ++_itemSeq, started = ++_clock;
    _openId = id;
    if (item != null && idOf(item as T) != id) item = null;
    final before = itemState == ConfigLoad.loading
        ? ConfigLoad.idle
        : itemState;
    itemState = ConfigLoad.loading;
    itemError = null;
    emit();
    try {
      final result = await readOne(api, id);
      if (!live(epoch) || seq != _itemSeq) return;
      if (!_fresh(started)) {
        _keepStored(id, before);
        return;
      }
      item = result;
      itemState = ConfigLoad.ready;
      reconcile(started, id: id);
      _replaceInList(result);
    } catch (failure) {
      if (!live(epoch) || seq != _itemSeq) return;
      if (!_fresh(started)) {
        _keepStored(id, before);
        return;
      }
      if (expired(failure)) return;
      if (failure is ConnectException && failure.code == Code.notFound) {
        // Absent after every write that ended before this read began: that
        // settles them (a delete or a create that did not happen).
        reconcile(started, id: id);
        items = List.unmodifiable(items.where((i) => idOf(i) != id));
        objectRemoved(id);
      }
      item = null;
      itemState = ConfigLoad.failed;
      itemError = configFailureMessage(failure, kind, ConfigOperation.load);
    }
    emit();
  }

  void _keepStored(String id, ConfigLoad before) {
    itemState = item != null && idOf(item as T) == id
        ? ConfigLoad.ready
        : before;
    itemError = null;
    emit();
  }

  void _replaceInList(T result) {
    final id = idOf(result);
    if (items.any((i) => idOf(i) == id)) {
      items = List.unmodifiable([
        for (final i in items) idOf(i) == id ? result : i,
      ]);
    }
  }

  // ---- Writes: one explicit call each, never repeated on their own ----

  bool writing = false;
  String? writeError;

  /// Writes whose outcome is unknown: the tick at which each ended, the object
  /// it aimed at and what kind of read can settle it.
  final List<({int tick, String? id, String scope})> _unknown = [];
  bool get unconfirmed => _unknown.isNotEmpty;
  String? get writeNotice => unconfirmed ? configUnconfirmed : null;

  String? _inFlightId;
  String _inFlightScope = 'config';

  /// A read that began at [startTick] and covers [id] (every object when null)
  /// settles the unknown writes of [scope] that ended before it began.
  @protected
  void reconcile(int startTick, {String? id, String scope = 'config'}) {
    _unknown.removeWhere(
      (w) =>
          w.scope == scope && w.tick < startTick && (id == null || w.id == id),
    );
  }

  @protected
  int beginRead() => ++_clock;

  /// Runs [action] as the one write in flight. Refused while another write is
  /// in flight or unconfirmed.
  @protected
  Future<bool> runWrite(
    Future<void> Function(int epoch) action,
    ConfigOperation operation, {
    String? id,
    String scope = 'config',
  }) async {
    if (_disposed || writing || unconfirmed) return false;
    final epoch = _epoch;
    writing = true;
    writeError = null;
    _inFlightId = id;
    _inFlightScope = scope;
    _lastWriteTick = ++_clock;
    objectChanging(id);
    emit();
    var ok = false;
    try {
      await action(epoch);
      ok = live(epoch);
    } catch (failure) {
      if (!live(epoch) || expired(failure)) return false;
      writeError = configFailureMessage(failure, kind, operation);
      if (configWriteOutcomeUnknown(failure)) {
        _unknown.add((tick: ++_clock, id: id, scope: scope));
      }
    } finally {
      if (live(epoch)) {
        writing = false;
        _inFlightId = null;
        _lastWriteTick = ++_clock;
        objectChanging(id);
        emit();
      }
    }
    return ok;
  }

  @protected
  void stored(T saved) {
    final id = idOf(saved);
    item = saved;
    itemState = ConfigLoad.ready;
    items = List.unmodifiable([
      for (final i in items) idOf(i) == id ? saved : i,
      if (!items.any((i) => idOf(i) == id)) saved,
    ]);
  }

  /// [draft] is the whole object to create; the server stores what is sent.
  Future<bool> create(T draft) => runWrite(id: idOf(draft), (epoch) async {
    final saved = await createOne(api, draft);
    if (live(epoch)) stored(saved);
  }, ConfigOperation.create);

  /// [draft] is the whole object to store: an update replaces it, so it starts
  /// from the stored copy and carries every field it does not edit.
  Future<bool> update(T draft) => runWrite(id: idOf(draft), (epoch) async {
    final saved = await updateOne(api, draft);
    if (live(epoch)) stored(saved);
  }, ConfigOperation.update);

  Future<bool> delete(String id) => runWrite(id: id, (epoch) async {
    await deleteOne(api, id);
    if (!live(epoch)) return;
    items = List.unmodifiable(items.where((i) => idOf(i) != id));
    if (item != null && idOf(item as T) == id) item = null;
    objectRemoved(id);
  }, ConfigOperation.delete);

  /// The app is going to the background: in-flight work is closed and fenced,
  /// and a write that was in flight is unconfirmed, never repeated.
  void suspend() {
    if (_disposed) return;
    _epoch++;
    _apiOrNull?.close();
    _apiOrNull = null;
    if (writing) {
      _unknown.add((tick: ++_clock, id: _inFlightId, scope: _inFlightScope));
      _lastWriteTick = _clock;
      objectChanging(_inFlightId);
    }
    writing = false;
    _inFlightId = null;
    // Stored credentials (header values, client secrets, passwords) must not
    // wait in memory while the app is in the background; [recover] reads
    // again what was open.
    _reloadList =
        _reloadList ||
        listState == ConfigLoad.ready ||
        listState == ConfigLoad.loading;
    items = const [];
    item = null;
    listState = ConfigLoad.idle;
    itemState = ConfigLoad.idle;
    onSuspend();
    emit();
  }

  /// Read-only recovery after the app returns to the foreground. It never
  /// repeats a write: an unconfirmed one stays so until these reads settle it.
  Future<void> recover() async {
    if (_disposed) return;
    final reload = _reloadList, open = _openId;
    _reloadList = false;
    await Future.wait([
      if (reload || unconfirmed) load(),
      if (open != null && (reload || unconfirmed)) loadItem(open),
      onRecover(),
    ]);
  }

  @override
  void dispose() {
    if (_disposed) return;
    _disposed = true;
    _epoch++;
    items = const [];
    item = null;
    onDispose();
    _apiOrNull?.close();
    _apiOrNull = null;
    super.dispose();
  }
}
