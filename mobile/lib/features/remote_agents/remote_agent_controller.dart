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

  @override
  void onSuspend() =>
      probes.removeWhere((_, p) => p.kind == RemoteProbeKind.probing);
}
