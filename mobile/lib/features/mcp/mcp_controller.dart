import 'dart:async';

import 'package:connectrpc/connect.dart' show Code, ConnectException;
import 'package:flutter/foundation.dart';

import '../../core/butter_api.dart';
import '../../core/config_errors.dart';
import '../../core/config_resource_controller.dart';
import '../../gen/agents/v1/agent.pb.dart';
import '../../gen/agents/v1/agent_service.pb.dart';
import '../../gen/agents/v1/dashboard.pb.dart';
import 'mcp_draft.dart';

enum McpStatusKind {
  probing,
  connected,
  disconnected,
  configured,
  error,
  failed,
}

/// The outcome of one explicit connection test.
class McpStatusView {
  const McpStatusView(this.kind, {this.reason = '', this.toolCount = 0});
  final McpStatusKind kind;

  /// A safe sentence: classified from the probe, never the server's own text.
  final String reason;
  final int toolCount;
}

/// The tools one server offered, or why they could not be listed.
class McpToolsView {
  const McpToolsView(this.state, {this.tools = const [], this.error});
  final ConfigLoad state;
  final List<MCPTool> tools;
  final String? error;
}

enum McpOAuthPhase {
  idle,
  starting,
  waiting,
  connected,
  failed,
  cancelled,
  timedOut,
}

/// How long the server keeps an authorization open; the App stops waiting then.
const mcpOAuthFlowTtl = Duration(minutes: 10);

/// How often a flow being waited on is looked at while the app is in the
/// foreground. Each look is one read-only status request; at most one is ever
/// in flight, and none is made after the flow ends or the app is suspended.
/// A flow is therefore read at most `mcpOAuthFlowTtl / mcpOAuthPollInterval`
/// times.
const mcpOAuthPollInterval = Duration(seconds: 5);

/// What watches one waiting flow: a poll timer and a deadline timer.
class _FlowWatch {
  _FlowWatch(this.view);
  final McpOAuthView view;
  Timer? poll;
  Timer? deadline;
}

/// One server's authorization, as the App last read it, and the flow it is
/// waiting on.
///
/// Butter completes the flow itself: the provider sends the browser back to
/// the server's callback, which exchanges the code and stores the connection.
/// The App therefore never sees a code or token; "connected" means the server
/// reports a connection newer than the one that existed when the flow started.
class McpOAuthView {
  McpOAuthView({
    this.phase = McpOAuthPhase.idle,
    this.message,
    this.status,
    this.deadline,
    this.baseline,
    this.baselineConnected = false,
    this.launching = false,
  });

  McpOAuthPhase phase;
  String? message;

  /// The last state read from the server (null: not read yet).
  MCPOAuthConnectionStatus? status;
  bool reading = false;
  DateTime? deadline;

  /// `connected_at` before the flow started, and whether a connection existed.
  DateTime? baseline;
  bool baselineConnected;

  /// The authorization address was handed to the browser launcher and its
  /// answer is not in yet. The flow is already recorded as waiting: the
  /// launcher may be slow, and the app may go to the background meanwhile.
  bool launching;

  bool get waiting => phase == McpOAuthPhase.waiting;
}

/// Shown when the server cannot keep an installed preset's hidden credentials.
const mcpKeepUnsupported =
    'This Butter server is too old to edit an installed preset without losing credentials the app cannot see, so nothing was changed. Ask the platform administrator to upgrade Butter. Until then, an edit that replaces or removes every hidden credential still works, and reinstalling the preset restores it.';

const _oauthAfterSuspend =
    'The app was closed while this was starting, so it is unknown whether the authorization page was prepared. Start again.';

/// MCP servers and the global presets of one server, signed-in user and
/// Workspace.
///
/// Presets are only read and installed. Installing a preset is one write
/// (create or overwrite of the Workspace server with the preset's ID), fenced
/// like any other. Probing a server and listing its tools are reads that make
/// the server connect out: they are explicit, bounded, and never retried.
class McpController extends ConfigResourceController<MCPServer> {
  McpController({
    required super.newApi,
    required super.token,
    required super.workspaceId,
    required super.onUnauthenticated,
    this.openBrowser,
    this.allowHttp = false,
    DateTime Function()? now,
    this.pollInterval = mcpOAuthPollInterval,
    Timer Function(Duration, void Function())? timer,
  }) : _now = now ?? DateTime.now,
       _timer = timer ?? Timer.new,
       super(kind: ConfigKind.mcp);

  /// Opens a URL in the system browser; false when it could not.
  final Future<bool> Function(Uri url)? openBrowser;

  /// Whether a plain-http authorization page may be opened (development
  /// servers only).
  final bool allowHttp;
  final DateTime Function() _now;

  /// How often a waiting flow is looked at in the foreground.
  final Duration pollInterval;
  final Timer Function(Duration, void Function()) _timer;

  @override
  String idOf(MCPServer item) => item.id;

  @override
  Future<List<MCPServer>> readAll(ButterApi api) =>
      api.mcpServers(token, workspaceId);

  @override
  Future<MCPServer> readOne(ButterApi api, String id) =>
      api.mcpServer(token, workspaceId, id);

  @override
  Future<MCPServer> createOne(ButterApi api, MCPServer draft) =>
      api.createMcpServer(token, workspaceId, draft);

  @override
  Future<MCPServer> updateOne(ButterApi api, MCPServer draft) =>
      api.updateMcpServer(token, workspaceId, draft);

  @override
  Future<void> deleteOne(ButterApi api, String id) =>
      api.deleteMcpServer(token, workspaceId, id);

  /// Whether the server can keep credentials the App cannot read while an
  /// installed preset is edited: null until found out, then what the server
  /// answered. Forgotten when the app returns to the foreground, since the
  /// server may have been upgraded meanwhile.
  bool? keepSupported;

  /// Saves an edit of a server whose credentials the form may not hold (an
  /// installed preset's). Without credentials to keep it is an ordinary
  /// whole-object update, which every server understands. With them it needs
  /// `UpdateMCPServerKeepingCredentials`: an older server ignores unknown
  /// fields of the ordinary request and would delete what was meant to be kept,
  /// so the App never sends keep fields there. It calls the dedicated method,
  /// and a server that answers it as unknown stored nothing; the edit is then
  /// refused with an upgrade explanation, with no fallback to the ordinary
  /// request. One call, never repeated on its own.
  Future<bool> saveEdit(McpSave save) => runWrite(id: save.server.id, (
    epoch,
  ) async {
    if (!save.keepsCredentials) {
      final saved = await api.updateMcpServer(token, workspaceId, save.server);
      if (live(epoch)) stored(saved);
      return;
    }
    if (keepSupported == false) {
      throw const ConfigUnsupported(mcpKeepUnsupported);
    }
    final MCPServer saved;
    try {
      saved = await api.updateMcpServerKeeping(
        token,
        workspaceId,
        save.server,
        keepHeaderKeys: save.keepHeaderKeys,
        keepClientSecret: save.keepClientSecret,
      );
    } on ConnectException catch (failure) {
      if (failure.code != Code.unimplemented) rethrow;
      if (live(epoch)) keepSupported = false;
      throw const ConfigUnsupported(mcpKeepUnsupported);
    }
    if (live(epoch)) {
      keepSupported = true;
      stored(saved);
    }
  }, ConfigOperation.update);

  // ---- Probes, tools: reads fenced against writes to the same server ----

  final Map<String, McpStatusView> statuses = {};
  final Map<String, McpToolsView> tools = {};
  final Map<String, int> _probeSeq = {};
  final Map<String, int> _toolsSeq = {};

  @override
  void objectChanging(String? id) {
    // Anything read about the server before a write to it may predate it.
    if (id == null) return;
    // A start still on its way to the browser belongs to the server as it
    // was; it must not open anything now.
    _flowSeq[id] = (_flowSeq[id] ?? 0) + 1;
    _probeSeq[id] = (_probeSeq[id] ?? 0) + 1;
    _toolsSeq[id] = (_toolsSeq[id] ?? 0) + 1;
    statuses.remove(id);
    tools.remove(id);
  }

  @override
  void objectRemoved(String id) {
    _flowSeq[id] = (_flowSeq[id] ?? 0) + 1;
    statuses.remove(id);
    tools.remove(id);
    _unwatch(id);
    _oauth.remove(id);
  }

  /// Makes the server connect to [id] now. A server that cannot be reached is
  /// an answer, not a failure of the call.
  Future<void> probe(String id) async {
    if (disposed) return;
    final epoch = this.epoch, seq = (_probeSeq[id] ?? 0) + 1;
    _probeSeq[id] = seq;
    statuses[id] = const McpStatusView(McpStatusKind.probing);
    emit();
    try {
      final reply = await api.mcpServerStatus(token, workspaceId, id);
      if (!live(epoch) || _probeSeq[id] != seq) return;
      statuses[id] = switch (reply.state) {
        MCPServerStatus_State.STATE_CONNECTED => McpStatusView(
          McpStatusKind.connected,
          reason: 'Connected.',
          toolCount: reply.toolCount,
        ),
        MCPServerStatus_State.STATE_DISCONNECTED => McpStatusView(
          McpStatusKind.disconnected,
          reason: mcpProbeReason(reply.detail),
        ),
        MCPServerStatus_State.STATE_CONFIGURED => const McpStatusView(
          McpStatusKind.configured,
          reason: 'Configured; the server did not connect to it.',
        ),
        _ => McpStatusView(
          McpStatusKind.error,
          reason: mcpProbeReason(reply.detail),
        ),
      };
    } catch (failure) {
      if (!live(epoch) || _probeSeq[id] != seq || expired(failure)) return;
      statuses[id] = McpStatusView(
        McpStatusKind.failed,
        reason: configFailureMessage(failure, kind, ConfigOperation.status),
      );
    }
    emit();
  }

  /// Lists the tools [id] offers, with which of them its filter allows.
  Future<void> loadTools(String id) async {
    if (disposed) return;
    final epoch = this.epoch, seq = (_toolsSeq[id] ?? 0) + 1;
    _toolsSeq[id] = seq;
    tools[id] = const McpToolsView(ConfigLoad.loading);
    emit();
    try {
      final reply = await api.mcpTools(token, workspaceId, id);
      if (!live(epoch) || _toolsSeq[id] != seq) return;
      final failure = reply.errors[id];
      tools[id] = failure != null
          ? McpToolsView(ConfigLoad.failed, error: mcpProbeReason(failure))
          : McpToolsView(
              ConfigLoad.ready,
              tools: List.unmodifiable(
                reply.tools.where(
                  (t) => t.serverId == id || t.serverId.isEmpty,
                ),
              ),
            );
    } catch (failure) {
      if (!live(epoch) || _toolsSeq[id] != seq || expired(failure)) return;
      tools[id] = McpToolsView(
        ConfigLoad.failed,
        error: configFailureMessage(failure, kind, ConfigOperation.tools),
      );
    }
    emit();
  }

  // ---- Global presets: read and install only ----

  List<MCPServer> presets = const [];
  ConfigLoad presetState = ConfigLoad.idle;
  String? presetError;
  int _presetSeq = 0;

  Future<void> loadPresets() async {
    if (disposed) return;
    final epoch = this.epoch, seq = ++_presetSeq;
    presetState = ConfigLoad.loading;
    presetError = null;
    emit();
    try {
      final list = await api.globalMcpPresets(token, workspaceId);
      if (!live(epoch) || seq != _presetSeq) return;
      presets = List.unmodifiable(list);
      presetState = ConfigLoad.ready;
    } catch (failure) {
      if (!live(epoch) || seq != _presetSeq || expired(failure)) return;
      presetState = ConfigLoad.failed;
      presetError = configFailureMessage(
        failure,
        ConfigKind.preset,
        ConfigOperation.load,
      );
    }
    emit();
  }

  /// Installs the preset into this Workspace: creates the server with the
  /// preset's ID, or overwrites the one an earlier install made. One call; an
  /// unknown outcome is settled by a read, never by installing again.
  Future<bool> install(String presetId) =>
      runWrite(id: presetId, (epoch) async {
        final saved = await api.installGlobalMcpPreset(
          token,
          workspaceId,
          presetId,
        );
        if (live(epoch)) stored(saved);
      }, ConfigOperation.install);

  // ---- OAuth ----

  final Map<String, McpOAuthView> _oauth = {};
  final Map<String, int> _oauthSeq = {};

  /// Per server: bumped by every start and by every write aimed at the server
  /// (edit, disconnect, delete, install, also when one begins). A start
  /// continuation that finds it moved has been overtaken.
  final Map<String, int> _flowSeq = {};
  McpOAuthView oauthOf(String id) => _oauth.putIfAbsent(id, McpOAuthView.new);

  bool _isWebUrl(Uri url) =>
      url.hasAuthority &&
      url.host.isNotEmpty &&
      (url.scheme == 'https' || (allowHttp && url.scheme == 'http'));

  /// Reads where the server's authorization of [id] stands. While a flow is
  /// being waited on, this is also how its end is noticed.
  Future<void> loadOAuth(String id) async {
    if (disposed) return;
    final view = oauthOf(id);
    final epoch = this.epoch, seq = (_oauthSeq[id] ?? 0) + 1;
    _oauthSeq[id] = seq;
    final started = beginRead();
    view.reading = true;
    emit();
    try {
      final status = await api.mcpOAuthStatus(token, workspaceId, id);
      if (!live(epoch) || _oauthSeq[id] != seq) return;
      view.reading = false;
      // A read that began before a write to this server ended may predate it
      // (a disconnect, an edit): it must neither show nor settle anything.
      if (!identical(_oauth[id], view) || !readIsFresh(started)) {
        emit();
        return;
      }
      view.status = status;
      // A read that began after a disconnect ended settles it.
      reconcile(started, id: id, scope: 'oauth');
      _settleWaiting(view, status);
      // A flow that has ended is not watched any longer.
      if (!view.waiting && identical(_watches[id]?.view, view)) _unwatch(id);
    } catch (failure) {
      if (!live(epoch) || _oauthSeq[id] != seq) return;
      view.reading = false;
      if (!identical(_oauth[id], view) || !readIsFresh(started)) {
        emit();
        return;
      }
      if (expired(failure)) return;
      if (view.waiting) {
        view.message =
            'Could not read the authorization state yet. ${configFailureMessage(failure, kind, ConfigOperation.oauthStatus)}';
      } else if (view.phase == McpOAuthPhase.cancelled ||
          view.phase == McpOAuthPhase.timedOut) {
        // The end of the flow is already said; a failed look afterwards does
        // not replace it.
      } else {
        view.message = configFailureMessage(
          failure,
          kind,
          ConfigOperation.oauthStatus,
        );
      }
    }
    emit();
  }

  /// Whether [status] shows the connection the waiting flow was started for.
  void _settleWaiting(McpOAuthView view, MCPOAuthConnectionStatus status) {
    // A flow the App stopped waiting for (stopped by the person, or expired)
    // can still have been finished in the browser: the server's state, against
    // the baseline taken before it started, is what decides.
    final stoppedWaiting =
        view.phase == McpOAuthPhase.cancelled ||
        view.phase == McpOAuthPhase.timedOut;
    if (!view.waiting && !stoppedWaiting) return;
    final connectedNow =
        status.state ==
        MCPOAuthConnectionState.MCPO_AUTH_CONNECTION_STATE_CONNECTED;
    final at = status.hasConnectedAt() ? status.connectedAt.toDateTime() : null;
    final newer =
        at != null && (view.baseline == null || at.isAfter(view.baseline!));
    if (connectedNow && (newer || !view.baselineConnected)) {
      view
        ..phase = McpOAuthPhase.connected
        ..message = stoppedWaiting
            ? 'Connected. The authorization was finished after the app stopped waiting for it.'
            : 'Connected.';
      return;
    }
    if (stoppedWaiting) return;
    if (_now().isAfter(view.deadline!)) {
      view
        ..phase = McpOAuthPhase.timedOut
        ..message =
            'No connection was made within ${mcpOAuthFlowTtl.inMinutes} minutes, so the authorization expired. If you declined it or the provider refused, nothing was connected. Start again.';
      return;
    }
    view.message =
        'Not connected yet. Finish signing in in the browser, then come back and check again.';
  }

  /// Starts authorizing [id] in the system browser.
  ///
  /// The current state is read first so a connection that already exists is
  /// not mistaken for the result of this flow; if that read fails nothing is
  /// started. Every step after an await checks that this start is still the
  /// current one for this server: a disconnect, edit, delete or install that
  /// began meanwhile, or a newer start, overtakes it, and an overtaken start
  /// opens nothing and changes nothing. Once the server has prepared the
  /// authorization the flow is recorded as waiting *before* the browser is
  /// asked to open it, so a suspension while the launcher is pending still
  /// finds the flow and reads its real state on return.
  Future<void> connectOAuth(String id) async {
    if (disposed) return;
    final view = oauthOf(id);
    if (view.phase == McpOAuthPhase.starting || view.launching || writing) {
      return;
    }
    if (unconfirmed) {
      view
        ..phase = McpOAuthPhase.failed
        ..message =
            'An earlier change is not confirmed. Refresh the server first.';
      emit();
      return;
    }
    final epoch = this.epoch;
    final flow = (_flowSeq[id] ?? 0) + 1;
    _flowSeq[id] = flow;
    view
      ..phase = McpOAuthPhase.starting
      ..message = null
      ..deadline = null;
    emit();

    // True while this start is still the one that may act.
    bool current() =>
        !disposed && identical(_oauth[id], view) && _flowSeq[id] == flow;

    // Ends an overtaken start without touching state it no longer owns.
    void overtaken() {
      if (disposed) return;
      if (identical(_oauth[id], view) && view.phase == McpOAuthPhase.starting) {
        view
          ..phase = McpOAuthPhase.failed
          ..message =
              'The server changed while authorization was starting, so the browser was not opened. Start again if you still want to connect.';
        emit();
      }
    }

    final readTick = beginRead();
    MCPOAuthConnectionStatus before;
    try {
      before = await api.mcpOAuthStatus(token, workspaceId, id);
    } catch (failure) {
      if (!current()) return overtaken();
      if (!live(epoch)) return _suspendedDuringStart(view);
      if (expired(failure)) return;
      view
        ..phase = McpOAuthPhase.failed
        ..message =
            'Could not read the current authorization, so nothing was started. ${configFailureMessage(failure, kind, ConfigOperation.oauthStatus)}';
      emit();
      return;
    }
    if (!current() || !readIsFresh(readTick)) return overtaken();
    if (!live(epoch)) return _suspendedDuringStart(view);

    StartMCPServerOAuthResponse started;
    try {
      started = await api.startMcpOAuth(token, workspaceId, id);
    } catch (failure) {
      if (!current()) return overtaken();
      if (!live(epoch)) return _suspendedDuringStart(view);
      if (expired(failure)) return;
      view
        ..phase = McpOAuthPhase.failed
        ..message = configFailureMessage(
          failure,
          kind,
          ConfigOperation.oauthStart,
        );
      emit();
      return;
    }
    if (!current()) return overtaken();
    if (!live(epoch)) return _suspendedDuringStart(view);

    final url = Uri.tryParse(started.authorizationUrl);
    final open = openBrowser;
    if (url == null || !_isWebUrl(url)) {
      view
        ..phase = McpOAuthPhase.failed
        ..message =
            'The server returned an authorization address the app will not open. Nothing was connected.';
      emit();
      return;
    }
    if (open == null) {
      view
        ..phase = McpOAuthPhase.failed
        ..message =
            'This app cannot open the browser to authorize. Nothing was connected.';
      emit();
      return;
    }

    // Recorded before the launcher runs: from here a suspension keeps the
    // flow, and the return reads its real state.
    view
      ..phase = McpOAuthPhase.waiting
      ..launching = true
      ..message = 'Opening the browser…'
      ..baseline = before.hasConnectedAt()
          ? before.connectedAt.toDateTime()
          : null
      ..baselineConnected =
          before.state ==
          MCPOAuthConnectionState.MCPO_AUTH_CONNECTION_STATE_CONNECTED
      ..deadline = _now().add(mcpOAuthFlowTtl)
      ..status = before;
    // Looked at from here on, whatever the launcher does.
    _watch(id, view);
    emit();

    var opened = false;
    try {
      opened = await open(url);
    } catch (_) {
      opened = false;
    }
    if (disposed) return;
    if (!identical(_oauth[id], view)) {
      // Replaced meanwhile (disconnect, delete). A page that did open cannot
      // be withdrawn, so say so on whatever is shown now.
      final now = _oauth[id];
      if (opened && now != null && !now.waiting) {
        now.message =
            'An authorization page was opened just before this change. Finishing it in the browser would connect again.';
        emit();
      }
      return;
    }
    // Stopped or expired while the launcher was pending: what the person
    // sees now is already the end of this flow.
    if (!view.launching) return;
    view.launching = false;
    if (!opened) {
      _unwatch(id);
      view
        ..phase = McpOAuthPhase.failed
        ..message =
            'Could not open the browser to authorize. Nothing was connected; try again.';
      emit();
      return;
    }
    if (view.waiting) {
      view.message =
          'Finish signing in in the browser, then return to Butter. The link works for ${mcpOAuthFlowTtl.inMinutes} minutes.';
    }
    emit();
  }

  void _suspendedDuringStart(McpOAuthView view) {
    if (disposed) return;
    if (view.phase == McpOAuthPhase.starting) {
      view
        ..phase = McpOAuthPhase.failed
        ..message = _oauthAfterSuspend;
    }
  }

  /// Stops waiting. The server has no way to withdraw an authorization that was
  /// opened, so the link stays valid until it expires; a connection finished in
  /// the browser afterwards still shows when the state is read.
  void cancelOAuth(String id) {
    final view = oauthOf(id);
    if (!view.waiting) return;
    _unwatch(id);
    view
      ..launching = false
      ..phase = McpOAuthPhase.cancelled
      ..message =
          'Stopped waiting. The authorization link stays valid for up to ${mcpOAuthFlowTtl.inMinutes} minutes; finishing it in the browser would still connect, so check the state to see.';
    emit();
  }

  /// Deletes the stored connection. One call; an unknown outcome is settled by
  /// reading the state, never by calling again.
  Future<bool> disconnectOAuth(String id) => runWrite(id: id, scope: 'oauth', (
    epoch,
  ) async {
    await api.disconnectMcpOAuth(token, workspaceId, id);
    if (!live(epoch)) return;
    _oauthSeq[id] = (_oauthSeq[id] ?? 0) + 1;
    _unwatch(id);
    _oauth[id] = McpOAuthView(
      phase: McpOAuthPhase.idle,
      message: 'Disconnected.',
      status: MCPOAuthConnectionStatus(
        serverId: id,
        state: MCPOAuthConnectionState.MCPO_AUTH_CONNECTION_STATE_DISCONNECTED,
      ),
    );
    statuses.remove(id);
    tools.remove(id);
  }, ConfigOperation.oauthDisconnect);

  // ---- Lifecycle ----

  @override
  void onSuspend() {
    statuses.removeWhere((_, s) => s.kind == McpStatusKind.probing);
    tools.removeWhere((_, t) => t.state == ConfigLoad.loading);
    if (presetState == ConfigLoad.loading) presetState = ConfigLoad.idle;
    // A flow being waited on survives: the app goes to the background while
    // the browser is in front, which is the point. A start that was in
    // flight is of unknown outcome.
    // Nothing watches while the app is in the background; the return looks.
    _unwatchAll();
    for (final view in _oauth.values) {
      view.reading = false;
      if (view.phase == McpOAuthPhase.starting) {
        view
          ..phase = McpOAuthPhase.failed
          ..message = _oauthAfterSuspend;
      }
    }
  }

  /// Back in the foreground: read again what was open, and look for the end of
  /// any authorization being waited on.
  @override
  Future<void> onRecover() async {
    keepSupported = null;
    final reads = <Future<void>>[];
    for (final e in _oauth.entries.toList()) {
      final view = e.value;
      if (!view.waiting) continue;
      // The deadline passed while the app was away: waiting ends now, and the
      // read below only finds out whether it finished anyway.
      if (!_now().isBefore(view.deadline!)) {
        _expire(e.key, view);
        continue;
      }
      _watch(e.key, view);
      reads.add(loadOAuth(e.key));
    }
    await Future.wait(reads);
  }

  @override
  void onDispose() => _unwatchAll();

  // ---- Watching a waiting flow ----

  final Map<String, _FlowWatch> _watches = {};

  /// Starts looking at [view] (waiting, with its deadline) every
  /// [pollInterval] and ends it at its deadline. Replaces any earlier watch of
  /// [id]. Only reads: a watch never starts a flow and never writes.
  void _watch(String id, McpOAuthView view) {
    _unwatch(id);
    final watch = _FlowWatch(view);
    _watches[id] = watch;
    final left = view.deadline!.difference(_now());
    watch.deadline = _timer(
      left.isNegative ? Duration.zero : left,
      () => _deadlineReached(id, watch),
    );
    watch.poll = _timer(pollInterval, () => _poll(id, watch));
  }

  void _unwatch(String id) {
    final watch = _watches.remove(id);
    watch?.poll?.cancel();
    watch?.deadline?.cancel();
  }

  void _unwatchAll() {
    for (final id in _watches.keys.toList()) {
      _unwatch(id);
    }
  }

  /// Whether [watch] is still the one that may act: the controller is alive,
  /// it is still the watch of this server, and the flow it watches is still the
  /// one being waited on.
  bool _watching(String id, _FlowWatch watch) =>
      !disposed &&
      identical(_watches[id], watch) &&
      identical(_oauth[id], watch.view) &&
      watch.view.waiting;

  Future<void> _poll(String id, _FlowWatch watch) async {
    watch.poll = null;
    if (!_watching(id, watch)) {
      if (identical(_watches[id], watch)) _unwatch(id);
      return;
    }
    // One look at a time: a request that is slow or hangs is not stacked
    // upon (the request's own deadline ends it, and the flow's deadline ends
    // the waiting whatever it does).
    if (!watch.view.reading) await loadOAuth(id);
    if (!_watching(id, watch)) {
      if (identical(_watches[id], watch)) _unwatch(id);
      return;
    }
    watch.poll = _timer(pollInterval, () => _poll(id, watch));
  }

  void _deadlineReached(String id, _FlowWatch watch) {
    watch.deadline = null;
    if (!_watching(id, watch)) return;
    _expire(id, watch.view);
  }

  /// Ends the waiting for [view] at its deadline, whatever the state of any
  /// status request, then reads once more to learn whether the browser
  /// finished it in the last moment.
  void _expire(String id, McpOAuthView view) {
    if (disposed || !identical(_oauth[id], view) || !view.waiting) return;
    _unwatch(id);
    // An answer still on its way predates the end and is dropped; the read
    // below is the one that counts.
    _oauthSeq[id] = (_oauthSeq[id] ?? 0) + 1;
    view
      ..reading = false
      ..launching = false
      ..phase = McpOAuthPhase.timedOut
      ..message =
          'No connection was made within ${mcpOAuthFlowTtl.inMinutes} minutes, so the authorization expired. If you declined it or the provider refused, nothing was connected. Start again.';
    emit();
    unawaited(loadOAuth(id));
  }

  @visibleForTesting
  bool get hasWaitingFlow => _oauth.values.any((v) => v.waiting);
}
