import 'dart:async';

import 'package:connectrpc/connect.dart';
import 'package:flutter/foundation.dart';
import '../../core/butter_api.dart';
import '../../core/session_errors.dart';
import '../../gen/agents/v1/agent_service.pb.dart';
import 'session_key.dart';

enum SessionLoad { idle, loading, ready, failed }

/// Filters for [ListSessions]. Empty strings mean no filter on that field.
class SessionFilters {
  const SessionFilters({
    this.appName = '',
    this.userId = '',
    this.startTime,
    this.endTime,
  });

  final String appName;
  final String userId;
  final DateTime? startTime;
  final DateTime? endTime;

  SessionFilters copyWith({
    String? appName,
    String? userId,
    DateTime? startTime,
    DateTime? endTime,
    bool clearStart = false,
    bool clearEnd = false,
  }) => SessionFilters(
    appName: appName ?? this.appName,
    userId: userId ?? this.userId,
    startTime: clearStart ? null : (startTime ?? this.startTime),
    endTime: clearEnd ? null : (endTime ?? this.endTime),
  );
}

/// Outcome of an explicit [SessionController.reply]; exposed for Cron (#464).
class SessionReplyResult {
  const SessionReplyResult.ok(this.response)
    : error = null,
      unconfirmed = false;
  const SessionReplyResult.failed(this.error)
    : response = null,
      unconfirmed = false;
  const SessionReplyResult.unconfirmed()
    : response = null,
      error = null,
      unconfirmed = true;

  final String? response;
  final String? error;
  final bool unconfirmed;
}

/// Authorized session list, detail, rename, delete and [ReplySession] for one
/// server, user and Workspace.
///
/// Writes are never repeated after an ambiguous outcome. A read that began
/// before a write ended cannot settle that write.
class SessionController extends ChangeNotifier {
  SessionController({
    required ButterApi Function() newApi,
    required this.token,
    required this.workspaceId,
    required this.onUnauthenticated,
  }) : _newApi = newApi;

  final ButterApi Function() _newApi;
  final String token;
  final String workspaceId;
  final void Function() onUnauthenticated;

  static const pageSize = 20;
  static const recentEventsLimit = 50;

  ButterApi? _apiOrNull;
  ButterApi get api => _apiOrNull ??= _newApi();
  bool _disposed = false;
  int _epoch = 0;

  bool get disposed => _disposed;
  bool _live(int epoch) => !_disposed && epoch == _epoch;

  int _clock = 0;
  int _lastWriteTick = 0;
  bool _fresh(int startTick) => startTick > _lastWriteTick;

  void _emit() {
    if (!_disposed) notifyListeners();
  }

  bool _expired(Object failure) {
    final expired =
        failure is ConnectException && failure.code == Code.unauthenticated;
    if (expired && !_disposed) onUnauthenticated();
    return expired;
  }

  // ---- List ----

  SessionFilters appliedFilters = const SessionFilters();
  SessionFilters draftFilters = const SessionFilters();
  final List<String> _pageTokens = [''];
  int _pageIndex = 0;

  List<SessionInfo> sessions = const [];
  int total = 0;
  String? nextPageToken;
  SessionLoad listState = SessionLoad.idle;
  String? listError;
  int _listSeq = 0;
  Future<void>? _listTail;

  Future<void> _whenListIdle() => _listTail ?? Future<void>.value();

  int get pageIndex => _pageIndex;
  bool get canPrevPage => _pageIndex > 0;
  bool get canNextPage => nextPageToken != null && nextPageToken!.isNotEmpty;

  void resetPagination() {
    _pageTokens
      ..clear()
      ..add('');
    _pageIndex = 0;
  }

  /// Applies [draftFilters] and loads page 1.
  Future<void> applyFilters() async {
    appliedFilters = draftFilters;
    resetPagination();
    await loadPage();
  }

  Future<void> loadPage({int? pageIndex}) async {
    if (_disposed) return;
    final previous = _listTail ?? Future<void>.value();
    final done = Completer<void>();
    _listTail = done.future;
    await previous;
    if (_disposed) {
      done.complete();
      return;
    }
    try {
      if (pageIndex != null) {
        if (pageIndex < 0 || pageIndex >= _pageTokens.length) return;
        _pageIndex = pageIndex;
      }
      final epoch = _epoch, seq = ++_listSeq, started = ++_clock;
      listState = SessionLoad.loading;
      listError = null;
      _emit();
      try {
        final pageToken = _pageTokens[_pageIndex];
        final reply = await api.listSessions(
          token,
          workspaceId,
          appName: appliedFilters.appName,
          userId: appliedFilters.userId,
          startTime: appliedFilters.startTime,
          endTime: appliedFilters.endTime,
          pageSize: pageSize,
          pageToken: pageToken,
        );
        if (!_live(epoch) || seq != _listSeq) return;
        if (!_fresh(started)) {
          listState = SessionLoad.ready;
          _emit();
          return;
        }
        sessions = List.unmodifiable(reply.sessions);
        total = reply.total;
        nextPageToken = reply.nextPageToken.isEmpty
            ? null
            : reply.nextPageToken;
        listState = SessionLoad.ready;
      } catch (failure) {
        if (!_live(epoch) || seq != _listSeq) return;
        if (!_fresh(started)) {
          listState = sessions.isEmpty ? SessionLoad.failed : SessionLoad.ready;
          _emit();
          return;
        }
        if (_expired(failure)) return;
        listState = SessionLoad.failed;
        listError = sessionFailureMessage(failure, SessionOperation.list);
      }
      _emit();
    } finally {
      done.complete();
    }
  }

  Future<void> nextPage() async {
    final token = nextPageToken;
    if (token == null || token.isEmpty) return;
    if (_pageIndex == _pageTokens.length - 1) {
      _pageTokens.add(token);
    }
    await loadPage(pageIndex: _pageIndex + 1);
  }

  Future<void> prevPage() async {
    if (_pageIndex <= 0) return;
    await loadPage(pageIndex: _pageIndex - 1);
  }

  // ---- Detail ----

  SessionKey? openKey;
  SessionDetail? detail;
  SessionLoad detailState = SessionLoad.idle;
  String? detailError;
  bool showAllEvents = false;
  int _detailSeq = 0;

  /// Opens [key] and loads its detail. Pass [fullHistory] true to request every
  /// event (num_recent_events = 0).
  Future<void> openSession(SessionKey key, {bool fullHistory = false}) async {
    openKey = key;
    showAllEvents = fullHistory;
    await reloadDetail();
  }

  void closeDetail() {
    openKey = null;
    detail = null;
    detailState = SessionLoad.idle;
    detailError = null;
    showAllEvents = false;
    _detailSeq++;
    _emit();
  }

  Future<void> reloadDetail() async {
    final key = openKey;
    if (_disposed || key == null) return;
    final epoch = _epoch, seq = ++_detailSeq, started = ++_clock;
    detailState = SessionLoad.loading;
    detailError = null;
    _emit();
    try {
      final loaded = await api.getSession(
        token,
        workspaceId,
        appName: key.appName,
        userId: key.userId,
        sessionId: key.sessionId,
        numRecentEvents: showAllEvents ? 0 : recentEventsLimit,
      );
      if (!_live(epoch) || seq != _detailSeq) return;
      if (!_fresh(started)) {
        detailState = detail == null ? SessionLoad.failed : SessionLoad.ready;
        _emit();
        return;
      }
      detail = loaded;
      detailState = SessionLoad.ready;
      _reconcileRename(key, loaded.session);
    } catch (failure) {
      if (!_live(epoch) || seq != _detailSeq) return;
      if (!_fresh(started)) {
        detailState = detail == null ? SessionLoad.failed : SessionLoad.ready;
        _emit();
        return;
      }
      if (_expired(failure)) return;
      detailState = SessionLoad.failed;
      detailError = sessionFailureMessage(failure, SessionOperation.read);
    }
    _emit();
  }

  Future<void> toggleFullHistory() async {
    showAllEvents = !showAllEvents;
    await reloadDetail();
  }

  void _reconcileRename(SessionKey key, SessionInfo info) {
    if (_renameUnconfirmed != key.storageKey) return;
    _renameUnconfirmed = null;
    _renameNotice = null;
    _replaceInList(info);
  }

  void _replaceInList(SessionInfo info) {
    final key = SessionKey.fromInfo(info);
    var found = false;
    final next = sessions.map((s) {
      if (SessionKey.fromInfo(s) == key) {
        found = true;
        return info;
      }
      return s;
    }).toList();
    if (found) sessions = List.unmodifiable(next);
  }

  // ---- Rename ----

  String? _renameUnconfirmed;
  String? _renameNotice;
  bool renameSending = false;
  String? renameError;

  bool get renameBlocked => _renameUnconfirmed != null || renameSending;

  String? get renameOutcomeNotice => _renameNotice;

  Future<bool> rename(SessionKey key, String title) async {
    if (_disposed || renameBlocked) return false;
    final trimmed = title.trim();
    if (trimmed.isEmpty) {
      renameError = 'Enter a non-empty title.';
      _emit();
      return false;
    }
    renameError = null;
    renameSending = true;
    _emit();
    await _whenListIdle();
    _lastWriteTick = ++_clock;
    try {
      final info = await api.renameSession(
        token,
        workspaceId,
        appName: key.appName,
        userId: key.userId,
        sessionId: key.sessionId,
        title: trimmed,
      );
      if (_disposed) return false;
      _renameUnconfirmed = null;
      _renameNotice = null;
      _replaceInList(info);
      if (openKey == key && detail != null) {
        detail!.session = info;
      }
      renameSending = false;
      _emit();
      return true;
    } catch (failure) {
      renameSending = false;
      if (_disposed) return false;
      if (_expired(failure)) return false;
      _renameUnconfirmed = key.storageKey;
      _renameNotice =
          'Rename not confirmed; it may already have been applied. Reload this session before trying again.';
      renameError = sessionFailureMessage(failure, SessionOperation.rename);
      _emit();
      return false;
    }
  }

  // ---- Delete ----

  bool deleteSending = false;
  String? deleteError;
  String? _deleteUnconfirmed;

  bool get deleteBlocked => deleteSending || _deleteUnconfirmed != null;

  Future<bool> deleteSession(SessionKey key) async {
    if (_disposed || deleteSending) return false;
    final epoch = _epoch;
    deleteSending = true;
    deleteError = null;
    _emit();
    await _whenListIdle();
    _lastWriteTick = ++_clock;
    try {
      await api.deleteSession(
        token,
        workspaceId,
        appName: key.appName,
        userId: key.userId,
        sessionId: key.sessionId,
      );
      if (!_live(epoch)) {
        deleteSending = false;
        return false;
      }
      _deleteUnconfirmed = null;
      sessions = List.unmodifiable([
        for (final s in sessions)
          if (SessionKey.fromInfo(s) != key) s,
      ]);
      if (openKey == key) closeDetail();
      deleteSending = false;
      _emit();
      return true;
    } catch (failure) {
      if (!_live(epoch)) return false;
      deleteSending = false;
      if (_expired(failure)) return false;
      _deleteUnconfirmed = key.storageKey;
      deleteError = sessionFailureMessage(failure, SessionOperation.delete);
      _emit();
      return false;
    }
  }

  // ---- Reply (for Cron and authorized session resume) ----

  bool replySending = false;
  String? replyError;
  String? _replyUnconfirmed;

  bool get replyBlocked => replySending || _replyUnconfirmed != null;

  /// Sends one user message through [ReplySession]. Never retried automatically.
  Future<SessionReplyResult> reply({
    required SessionKey key,
    required String agentId,
    required String message,
    String modelOverride = '',
  }) async {
    if (_disposed || replyBlocked) {
      return const SessionReplyResult.failed(
        'A reply is already in progress or waiting to be settled.',
      );
    }
    final trimmed = message.trim();
    if (trimmed.isEmpty) {
      return const SessionReplyResult.failed('Enter a message.');
    }
    if (agentId.trim().isEmpty) {
      return const SessionReplyResult.failed('Choose an Agent.');
    }
    final epoch = _epoch;
    replySending = true;
    replyError = null;
    _emit();
    await _whenListIdle();
    _lastWriteTick = ++_clock;
    try {
      final text = await api.replySession(
        token,
        workspaceId,
        appName: key.appName,
        userId: key.userId,
        sessionId: key.sessionId,
        agentId: agentId.trim(),
        message: trimmed,
        modelOverride: modelOverride,
      );
      if (!_live(epoch)) {
        replySending = false;
        return const SessionReplyResult.unconfirmed();
      }
      _replyUnconfirmed = null;
      replySending = false;
      if (openKey == key) {
        unawaited(reloadDetail());
      }
      _emit();
      return SessionReplyResult.ok(text);
    } catch (failure) {
      if (!_live(epoch)) {
        return const SessionReplyResult.unconfirmed();
      }
      replySending = false;
      if (_expired(failure)) {
        return SessionReplyResult.failed(
          sessionFailureMessage(failure, SessionOperation.reply),
        );
      }
      _replyUnconfirmed = key.storageKey;
      replyError = sessionFailureMessage(failure, SessionOperation.reply);
      _emit();
      return SessionReplyResult.failed(replyError!);
    }
  }

  void suspend() {
    _epoch++;
    _apiOrNull?.close();
    _apiOrNull = null;
    renameSending = false;
    deleteSending = false;
    replySending = false;
    _emit();
  }

  Future<void> recover() async {
    if (_disposed) return;
    if (listState == SessionLoad.ready || listError != null) {
      await loadPage();
    }
    if (openKey != null) {
      await reloadDetail();
    }
  }

  @override
  void dispose() {
    _disposed = true;
    _epoch++;
    _apiOrNull?.close();
    _apiOrNull = null;
    super.dispose();
  }
}
