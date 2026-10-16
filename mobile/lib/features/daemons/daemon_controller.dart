import 'dart:async';

import '../../core/butter_api.dart';
import '../../core/config_errors.dart';
import '../../core/config_resource_controller.dart';
import '../../core/daemon_errors.dart';
import '../../gen/agents/v1/agent.pb.dart';
import '../../gen/agents/v1/api_token.pb.dart';
import '../../gen/agents/v1/dashboard.pb.dart';

/// A worker token's secret, held only while it is on screen.
class MintedSecret {
  MintedSecret({required this.runtimeId, required this.token, this.secret});
  final String runtimeId;
  final APIToken token;

  /// Null once wiped.
  String? secret;
}

/// A task that was in flight on one read and is gone on the next. The server
/// lists in-flight tasks only, so it never says how a task ended.
class EndedTask {
  const EndedTask({
    required this.task,
    required this.cancelDelivered,
    required this.seenAt,
  });
  final DaemonTaskInFlight task;
  final bool cancelDelivered;
  final DateTime seenAt;
}

/// Daemon runtimes of one server, signed-in user and Workspace, and what runs
/// on them: connected workers, tasks in flight, server diagnostics and the
/// Workspace's worker tokens.
///
/// A runtime is a whole-object configuration write like any other. A worker
/// token is separate from every other API token: only the DAEMON kind is ever
/// listed or revoked from here, and the one-time secret of a new one lives
/// only on [minted] until it is dismissed, the app leaves the foreground or
/// the context is disposed. Minting stores another token every time, so it is
/// never repeated on its own: after a lost reply the person decides, from the
/// token list, what to do.
class DaemonController extends ConfigResourceController<DaemonRuntime> {
  DaemonController({
    required super.newApi,
    required super.token,
    required super.workspaceId,
    required super.onUnauthenticated,
    DateTime Function()? now,
  }) : _now = now ?? DateTime.now,
       super(kind: ConfigKind.daemon);

  final DateTime Function() _now;

  @override
  String idOf(DaemonRuntime item) => item.id;

  @override
  Future<List<DaemonRuntime>> readAll(ButterApi api) =>
      api.daemonRuntimes(token, workspaceId);

  @override
  Future<DaemonRuntime> readOne(ButterApi api, String id) =>
      api.daemonRuntime(token, workspaceId, id);

  @override
  Future<DaemonRuntime> createOne(ButterApi api, DaemonRuntime draft) =>
      api.createDaemonRuntime(token, workspaceId, draft);

  @override
  Future<DaemonRuntime> updateOne(ButterApi api, DaemonRuntime draft) =>
      api.updateDaemonRuntime(token, workspaceId, draft);

  @override
  Future<void> deleteOne(ButterApi api, String id) =>
      api.deleteDaemonRuntime(token, workspaceId, id);

  // ---- Connected workers ----

  List<DaemonStatus> online = const [];
  ConfigLoad onlineState = ConfigLoad.idle;
  String? onlineError;
  int _onlineSeq = 0;
  bool _reloadOnline = false;

  /// The worker connected for [runtimeId], or null when it is offline.
  DaemonStatus? workerOf(String runtimeId) {
    for (final w in online) {
      if (w.daemonRuntimeId == runtimeId) return w;
    }
    return null;
  }

  Future<void> loadOnline() async {
    if (disposed) return;
    final epoch = this.epoch, seq = ++_onlineSeq;
    _reloadOnline = true;
    onlineState = ConfigLoad.loading;
    onlineError = null;
    emit();
    try {
      final list = await api.daemons(token, workspaceId);
      if (!live(epoch) || seq != _onlineSeq) return;
      online = List.unmodifiable(list);
      onlineState = ConfigLoad.ready;
    } catch (failure) {
      if (!live(epoch) || seq != _onlineSeq || expired(failure)) return;
      onlineState = ConfigLoad.failed;
      onlineError = configFailureMessage(failure, kind, ConfigOperation.status);
    }
    emit();
  }

  // ---- Tasks in flight ----

  List<DaemonTaskInFlight> tasks = const [];
  ConfigLoad taskState = ConfigLoad.idle;
  String? taskError;
  DateTime? tasksReadAt;
  int _taskSeq = 0;
  bool _reloadTasks = false;

  /// Tasks that left the in-flight list since the page opened, newest first.
  List<EndedTask> ended = const [];
  static const _endedKept = 20;

  final Map<String, CancelView> cancels = {};

  Future<void>? _tasksRead;

  /// Reads the tasks. A call while a read is in flight joins it instead of
  /// starting another, so a slow server is asked one question at a time and
  /// the answer that arrives is never discarded for a newer, unanswered one.
  Future<void> loadTasks() {
    if (disposed) return Future.value();
    final running = _tasksRead;
    if (running != null) return running;
    late final Future<void> run;
    run = _readTasks().whenComplete(() {
      if (identical(_tasksRead, run)) _tasksRead = null;
    });
    _tasksRead = run;
    return run;
  }

  Future<void> _readTasks() async {
    final epoch = this.epoch, seq = ++_taskSeq;
    _reloadTasks = true;
    if (taskState != ConfigLoad.ready) taskState = ConfigLoad.loading;
    taskError = null;
    emit();
    try {
      final list = await api.daemonTasks(token, workspaceId);
      if (!live(epoch) || seq != _taskSeq) return;
      _noteEnded(list);
      tasks = List.unmodifiable(list);
      taskState = ConfigLoad.ready;
      tasksReadAt = _now();
    } catch (failure) {
      if (!live(epoch) || seq != _taskSeq || expired(failure)) return;
      // The last list stays, marked by [taskError] and [tasksReadAt]: it is
      // old, and the page says so.
      taskState = tasks.isEmpty && tasksReadAt == null
          ? ConfigLoad.failed
          : ConfigLoad.ready;
      taskError = configFailureMessage(failure, kind, ConfigOperation.status);
    }
    emit();
  }

  void _noteEnded(List<DaemonTaskInFlight> next) {
    final still = {for (final t in next) t.taskId};
    final gone = [
      for (final t in tasks)
        if (!still.contains(t.taskId))
          EndedTask(
            task: t,
            cancelDelivered: cancels[t.taskId]?.kind == CancelKind.delivered,
            seenAt: _now(),
          ),
    ];
    if (gone.isNotEmpty) {
      ended = List.unmodifiable([...gone.reversed, ...ended].take(_endedKept));
    }
    // A cancel outcome only means something while its task is listed.
    cancels.removeWhere((id, _) => !still.contains(id));
  }

  /// Delivers a cancel for [task]. A repeat is harmless, so only a cancel in
  /// flight blocks another. The outcome in [cancels] says what is known: the
  /// task itself is still listed until the worker reports its own end.
  Future<void> cancel(DaemonTaskInFlight task) async {
    if (disposed || cancels[task.taskId]?.sending == true) return;
    final epoch = this.epoch, id = task.taskId;
    cancels[id] = cancelSending;
    emit();
    try {
      final runtime = await api.cancelDaemonTask(
        token,
        workspaceId,
        id,
        task.daemonRuntimeId,
      );
      if (!live(epoch)) return;
      cancels[id] = cancelDelivered(
        runtime.isEmpty ? task.daemonRuntimeId : runtime,
      );
    } catch (failure) {
      if (!live(epoch) || expired(failure)) return;
      cancels[id] = cancelFailureView(failure);
    }
    emit();
  }

  // ---- Diagnostics ----

  BridgeDiagnostics? diagnostics;
  ConfigLoad diagnosticsState = ConfigLoad.idle;
  String? diagnosticsError;
  int _diagnosticsSeq = 0;
  bool _reloadDiagnostics = false;

  Future<void> loadDiagnostics() async {
    if (disposed) return;
    final epoch = this.epoch, seq = ++_diagnosticsSeq;
    _reloadDiagnostics = true;
    diagnosticsState = ConfigLoad.loading;
    diagnosticsError = null;
    emit();
    try {
      final reply = await api.bridgeDiagnostics(token, workspaceId);
      if (!live(epoch) || seq != _diagnosticsSeq) return;
      diagnostics = reply;
      diagnosticsState = ConfigLoad.ready;
    } catch (failure) {
      if (!live(epoch) || seq != _diagnosticsSeq || expired(failure)) return;
      diagnosticsState = ConfigLoad.failed;
      diagnosticsError = configFailureMessage(
        failure,
        kind,
        ConfigOperation.status,
      );
    }
    emit();
  }

  // ---- What depends on a runtime ----

  /// The Remote Agents of this Workspace that run on [runtimeId], read now.
  /// Null when the read failed: the caller must say it could not check.
  Future<List<RemoteAgent>?> remoteAgentsOn(String runtimeId) async {
    if (disposed) return null;
    final epoch = this.epoch;
    try {
      final all = await api.remoteAgents(token, workspaceId);
      if (!live(epoch)) return null;
      return [
        for (final a in all)
          if (a.protocol == RemoteAgentProtocol.REMOTE_AGENT_PROTOCOL_DAEMON &&
              a.daemonRuntimeId == runtimeId)
            a,
      ];
    } catch (failure) {
      if (live(epoch)) expired(failure);
      return null;
    }
  }

  // ---- Worker tokens ----

  /// The DAEMON tokens of the Workspace. Tokens of any other kind are dropped
  /// on arrival and never reach the screen.
  List<APIToken> workerTokens = const [];
  ConfigLoad tokenState = ConfigLoad.idle;
  String? tokenError;
  int _tokenSeq = 0;
  bool _reloadTokens = false;

  List<APIToken> tokensOf(String runtimeId) => [
    for (final t in workerTokens)
      if (t.daemonRuntimeId == runtimeId) t,
  ];

  /// One read of the token list. It settles unknown mints and revocations
  /// that ended before it began, but never tells which token a lost mint made.
  Future<void> loadTokens() async {
    if (disposed) return;
    final epoch = this.epoch, seq = ++_tokenSeq, started = beginRead();
    _reloadTokens = true;
    tokenState = ConfigLoad.loading;
    tokenError = null;
    emit();
    try {
      final list = await api.apiTokens(token, workspaceId);
      if (!live(epoch) || seq != _tokenSeq) return;
      if (!readIsFresh(started)) {
        tokenState = workerTokens.isEmpty && tokenState != ConfigLoad.ready
            ? ConfigLoad.idle
            : ConfigLoad.ready;
        emit();
        return;
      }
      workerTokens = List.unmodifiable([
        for (final t in list)
          if (t.kind == APITokenKind.API_TOKEN_KIND_DAEMON) t,
      ]);
      tokenState = ConfigLoad.ready;
      reconcile(started, scope: 'mint');
      reconcile(started, scope: 'revoke');
    } catch (failure) {
      if (!live(epoch) || seq != _tokenSeq) return;
      if (!readIsFresh(started)) {
        tokenState = ConfigLoad.idle;
        emit();
        return;
      }
      if (expired(failure)) return;
      tokenState = ConfigLoad.failed;
      tokenError = configFailureMessage(failure, kind, ConfigOperation.load);
    }
    emit();
  }

  /// The secret on screen, or null.
  MintedSecret? minted;

  int _screenVisit = 0;

  /// The screen that showed (or was about to show) a secret was left or
  /// covered. The secret is wiped, and a mint still in flight is fenced: its
  /// reply is not kept, because nobody is looking at the screen it belongs to.
  /// [notify] is false from `dispose`, when rebuilding is not allowed.
  void leaveSecretScreen(String runtimeId, {bool notify = true}) {
    _screenVisit++;
    final m = minted;
    if (m == null || m.runtimeId != runtimeId) return;
    m.secret = null;
    minted = null;
    if (notify) emit();
  }

  /// Mints a worker token for [runtimeId] that expires after [ttl], or never
  /// when it is null. Refused while another secret is still on screen, so one
  /// is never replaced unseen.
  Future<bool> mint(String runtimeId, String name, {Duration? ttl}) async {
    if (minted != null) return false;
    if (ttl != null && ttl <= Duration.zero) {
      // Nothing was asked of the server: a zero lifetime is never sent, since
      // the server would read it as a token that does not expire.
      writeError = 'The expiry must be a positive length of time.';
      emit();
      return false;
    }
    final visit = _screenVisit;
    final ok = await runWrite(id: runtimeId, scope: 'mint', (epoch) async {
      final reply = await api.createDaemonRuntimeToken(
        token,
        workspaceId,
        runtimeId,
        name,
        ttl: ttl,
      );
      if (!live(epoch)) return;
      if (visit != _screenVisit) {
        // The token exists, but its secret has nowhere to be shown. It is
        // never kept for a later screen; the person decides from the list.
        markUnconfirmed(id: runtimeId, scope: 'mint');
        writeError =
            'A worker token was created, but you left the screen before its secret could be shown, so the secret was discarded. Revoke that token and create another.';
        return;
      }
      minted = MintedSecret(
        runtimeId: runtimeId,
        token: reply.token,
        secret: reply.secret,
      );
      workerTokens = List.unmodifiable([...workerTokens, reply.token]);
    }, ConfigOperation.mint);
    return ok && minted != null;
  }

  /// Wipes the secret. The token stays listed by its prefix.
  void dismissSecret() {
    final m = minted;
    if (m == null) return;
    m.secret = null;
    minted = null;
    emit();
  }

  Future<bool> revoke(String tokenId) =>
      runWrite(id: tokenId, scope: 'revoke', (epoch) async {
        await api.revokeApiToken(token, workspaceId, tokenId);
        if (!live(epoch)) return;
        workerTokens = List.unmodifiable([
          for (final t in workerTokens)
            if (t.id == tokenId) (t.deepCopy()..revoked = true) else t,
        ]);
      }, ConfigOperation.revoke);

  @override
  String? get writeNotice {
    if (!unconfirmed) return null;
    final scopes = unconfirmedScopes;
    if (scopes.contains('mint')) {
      return 'A worker token may have been created, but its secret was never shown and cannot be shown again. Check the tokens, and revoke any you do not recognise. Nothing was minted again.';
    }
    if (scopes.contains('revoke')) {
      return 'A token revocation was not confirmed. Check the tokens before changing anything else.';
    }
    return configUnconfirmed;
  }

  // ---- Lifetime ----

  @override
  void onSuspend() {
    // The secret must not wait in memory while the app is in the background.
    dismissSecret();
    // A read begun before the app left must not be joined after it came back.
    _tasksRead = null;
    for (final e in cancels.entries.toList()) {
      if (e.value.sending) cancels[e.key] = cancelLostWithApp;
    }
    online = const [];
    onlineState = ConfigLoad.idle;
    tasks = const [];
    taskState = ConfigLoad.idle;
    tasksReadAt = null;
    ended = const [];
    diagnostics = null;
    diagnosticsState = ConfigLoad.idle;
    workerTokens = const [];
    tokenState = ConfigLoad.idle;
  }

  @override
  Future<void> onRecover() => Future.wait([
    if (_reloadOnline) loadOnline(),
    if (_reloadTasks) loadTasks(),
    if (_reloadDiagnostics) loadDiagnostics(),
    if (_reloadTokens || unconfirmed) loadTokens(),
  ]);

  @override
  void onDispose() {
    _tasksRead = null;
    final m = minted;
    if (m != null) m.secret = null;
    minted = null;
    online = const [];
    tasks = const [];
    ended = const [];
    diagnostics = null;
    workerTokens = const [];
  }
}
