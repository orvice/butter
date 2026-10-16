import 'dart:async';

import '../../core/butter_api.dart';
import '../../core/config_errors.dart';
import '../../core/config_resource_controller.dart';
import '../../gen/agents/v1/agent.pb.dart';
import '../../gen/agents/v1/dashboard.pb.dart';

enum RemoteProbeKind {
  probing,
  reachable,
  unreachable,
  error,
  configured,
  failed,
}

/// The outcome of one explicit probe of a Remote Agent endpoint.
class RemoteProbeView {
  const RemoteProbeView(this.kind, {this.reason = '', this.latencyMs});
  final RemoteProbeKind kind;

  /// A safe sentence, classified from the probe; never the server's own text.
  final String reason;
  final int? latencyMs;
}

/// Remote Agents of one server, signed-in user and Workspace.
///
/// Create, update and delete are whole-object writes fenced like every
/// configuration write. The probe makes the server call the endpoint; it is
/// explicit, bounded and never retried. Daemon-protocol agents belong to the
/// Daemon feature: they are listed and probed, never written from here.
class RemoteAgentController extends ConfigResourceController<RemoteAgent> {
  RemoteAgentController({
    required super.newApi,
    required super.token,
    required super.workspaceId,
    required super.onUnauthenticated,
  }) : super(kind: ConfigKind.remote);

  @override
  String idOf(RemoteAgent item) => item.id;

  @override
  Future<List<RemoteAgent>> readAll(ButterApi api) =>
      api.remoteAgents(token, workspaceId);

  @override
  Future<RemoteAgent> readOne(ButterApi api, String id) =>
      api.remoteAgent(token, workspaceId, id);

  @override
  Future<RemoteAgent> createOne(ButterApi api, RemoteAgent draft) =>
      api.createRemoteAgent(token, workspaceId, draft);

  @override
  Future<RemoteAgent> updateOne(ButterApi api, RemoteAgent draft) =>
      api.updateRemoteAgent(token, workspaceId, draft);

  @override
  Future<void> deleteOne(ButterApi api, String id) =>
      api.deleteRemoteAgent(token, workspaceId, id);

  final Map<String, RemoteProbeView> probes = {};
  final Map<String, int> _probeSeq = {};

  @override
  void objectChanging(String? id) {
    if (id == null) return;
    _probeSeq[id] = (_probeSeq[id] ?? 0) + 1;
    probes.remove(id);
  }

  @override
  void objectRemoved(String id) => probes.remove(id);

  Future<void> probe(String id) async {
    if (disposed) return;
    final epoch = this.epoch, seq = (_probeSeq[id] ?? 0) + 1;
    _probeSeq[id] = seq;
    probes[id] = const RemoteProbeView(RemoteProbeKind.probing);
    emit();
    try {
      final reply = await api.remoteAgentStatus(token, workspaceId, id);
      if (!live(epoch) || _probeSeq[id] != seq) return;
      final latency = reply.latencyMs.toInt();
      probes[id] = RemoteProbeView(
        switch (reply.state) {
          RemoteAgentStatus_State.STATE_ACTIVE ||
          RemoteAgentStatus_State.STATE_IDLE => RemoteProbeKind.reachable,
          RemoteAgentStatus_State.STATE_UNREACHABLE =>
            RemoteProbeKind.unreachable,
          RemoteAgentStatus_State.STATE_CONFIGURED =>
            RemoteProbeKind.configured,
          _ => RemoteProbeKind.error,
        },
        reason: remoteProbeReason(reply.state, reply.detail),
        latencyMs: latency > 0 ? latency : null,
      );
    } catch (failure) {
      if (!live(epoch) || _probeSeq[id] != seq || expired(failure)) return;
      probes[id] = RemoteProbeView(
        RemoteProbeKind.failed,
        reason: configFailureMessage(failure, kind, ConfigOperation.status),
      );
    }
    emit();
  }

  // ---- The runtimes a Daemon agent can run on ----

  /// This Workspace's daemon runtimes, null until read (and again after the
  /// app leaves the foreground).
  List<DaemonRuntime>? runtimes;

  /// The workers connected now; null when unknown.
  List<DaemonStatus>? online;
  ConfigLoad runtimeState = ConfigLoad.idle;
  String? runtimeError;
  int _runtimeSeq = 0;
  bool _reloadRuntimes = false;

  Future<void> loadRuntimes() async {
    if (disposed) return;
    final epoch = this.epoch, seq = ++_runtimeSeq;
    _reloadRuntimes = true;
    runtimeState = ConfigLoad.loading;
    runtimeError = null;
    emit();
    List<DaemonRuntime>? read;
    String? problem;
    try {
      read = await api.daemonRuntimes(token, workspaceId);
      // A read that was superseded, or whose app left the foreground, asks
      // nothing more: `api` would open a fresh client for it.
      if (!live(epoch) || seq != _runtimeSeq) return;
    } catch (failure) {
      if (!live(epoch) || seq != _runtimeSeq || expired(failure)) return;
      problem = configFailureMessage(
        failure,
        ConfigKind.daemon,
        ConfigOperation.load,
      );
    }
    List<DaemonStatus>? workers;
    if (read != null) {
      try {
        workers = await api.daemons(token, workspaceId);
      } catch (failure) {
        if (!live(epoch) || seq != _runtimeSeq || expired(failure)) return;
        // Which workers are connected is a hint, not a requirement.
      }
    }
    if (!live(epoch) || seq != _runtimeSeq) return;
    runtimes = read == null ? null : List.unmodifiable(read);
    online = workers == null ? null : List.unmodifiable(workers);
    runtimeState = read == null ? ConfigLoad.failed : ConfigLoad.ready;
    runtimeError = problem;
    emit();
  }

  /// The runtime of [id] when it is known to be missing, else null.
  DaemonRuntime? runtimeOf(String id) {
    for (final r in runtimes ?? const <DaemonRuntime>[]) {
      if (r.id == id) return r;
    }
    return null;
  }

  /// Why a Daemon agent cannot work as configured, or null when nothing is
  /// known to be wrong (including before the runtimes are read: nothing is
  /// claimed without evidence). Never the server's text.
  String? referenceProblem(RemoteAgent agent) {
    if (agent.protocol != RemoteAgentProtocol.REMOTE_AGENT_PROTOCOL_DAEMON) {
      return null;
    }
    if (runtimes == null) return null;
    final id = agent.daemonRuntimeId;
    if (id.isEmpty) {
      return 'No runtime is chosen. Choose one of this Workspace\'s daemon runtimes.';
    }
    if (runtimeOf(id) == null) {
      return 'Runtime "$id" does not exist in this Workspace, so runs through this agent fail. Choose another runtime.';
    }
    final executor = agent.acpRuntime;
    for (final w in online ?? const <DaemonStatus>[]) {
      if (w.daemonRuntimeId != id) continue;
      if (executor.isNotEmpty &&
          w.acpRuntimes.isNotEmpty &&
          !w.acpRuntimes.contains(executor)) {
        return 'The worker connected for runtime "$id" does not offer executor "$executor".';
      }
    }
    return null;
  }

  @override
  void onSuspend() {
    probes.removeWhere((_, p) => p.kind == RemoteProbeKind.probing);
    runtimes = null;
    online = null;
    runtimeState = ConfigLoad.idle;
  }

  @override
  Future<void> onRecover() async {
    if (_reloadRuntimes) await loadRuntimes();
  }
}
