import 'dart:async';

import 'package:connectrpc/connect.dart';
import 'package:flutter/foundation.dart';

import '../../core/butter_api.dart';
import '../../core/config_resource_controller.dart';
import '../../core/ops_errors.dart';
import '../../gen/agents/v1/agent_service.pb.dart';
import '../../gen/agents/v1/api_token.pb.dart';
import '../../gen/agents/v1/cron.pb.dart';
import '../../gen/agents/v1/dashboard.pb.dart';
import 'api_token_logic.dart';
import 'invocation_logic.dart';

/// One-time secret for a newly created personal API token.
class MintedApiSecret {
  MintedApiSecret({required this.token, required this.secret});
  final APIToken token;
  String? secret;
}

/// Counts for the signed-in Workspace (never server-wide aggregates).
class WorkspaceScopedCounts {
  const WorkspaceScopedCounts({
    required this.agents,
    required this.mcpServers,
    required this.remoteAgents,
    required this.cronJobs,
    required this.daemonRuntimes,
  });
  final int agents;
  final int mcpServers;
  final int remoteAgents;
  final int cronJobs;
  final int daemonRuntimes;
}

/// Overview, invocation operations and personal API tokens for one server,
/// user and Workspace.
class WorkspaceOpsController extends ChangeNotifier {
  WorkspaceOpsController({
    required ButterApi Function() newApi,
    required this.token,
    required this.workspaceId,
    required this.isGlobalAdmin,
    required this.onUnauthenticated,
    required this.agentCount,
  }) : _newApi = newApi;

  final ButterApi Function() _newApi;
  final String token;
  final String workspaceId;
  final bool Function() isGlobalAdmin;
  final void Function() onUnauthenticated;
  final int Function() agentCount;

  ButterApi get api => _newApi();
  bool disposed = false;
  int epoch = 0;

  bool live(int gen) => !disposed && gen == epoch;

  void emit() {
    if (!disposed) notifyListeners();
  }

  void suspend() {
    epoch++;
    clearMintedSecret(notify: false);
    overviewState = ConfigLoad.idle;
    invocationsState = ConfigLoad.idle;
    tokenState = ConfigLoad.idle;
    emit();
  }

  Future<void> recover() async {
    await loadOverview();
  }

  @override
  void dispose() {
    disposed = true;
    clearMintedSecret(notify: false);
    super.dispose();
  }

  bool expired(Object failure) {
    if (failure is ConnectException && failure.code == Code.unauthenticated) {
      onUnauthenticated();
      return true;
    }
    return false;
  }

  // ---- Overview ----

  ConfigLoad overviewState = ConfigLoad.idle;
  String? overviewError;

  /// When true, [serverOverview] and related fields are server-wide (admin).
  bool overviewIsServerWide = false;

  GetOverviewResponse? serverOverview;
  List<ActivityEvent> activityEvents = const [];
  GetActivityMetricsResponse? activityMetrics;
  List<CronExecutionBucket> cronBuckets = const [];
  BridgeDiagnostics? serverHealth;
  WorkspaceScopedCounts? workspaceCounts;
  List<CronExecution> recentCronExecutions = const [];

  int _overviewSeq = 0;

  Future<void> loadOverview() async {
    if (disposed) return;
    final gen = epoch, seq = ++_overviewSeq;
    overviewState = ConfigLoad.loading;
    overviewError = null;
    emit();
    try {
      if (isGlobalAdmin()) {
        overviewIsServerWide = true;
        final results = await Future.wait([
          api.getDashboardOverview(token),
          api.getActivityFeed(token, limit: 20),
          api.getActivityMetrics(
            token,
            GetActivityMetricsRequest_Range.RANGE_7D,
          ),
          api.getCronExecutionTimeseries(
            token,
            GetCronExecutionTimeseriesRequest_Range.RANGE_7D,
          ),
        ]);
        if (!live(gen) || seq != _overviewSeq) return;
        serverOverview = results[0] as GetOverviewResponse;
        activityEvents = List.unmodifiable(
          (results[1] as GetActivityFeedResponse).events,
        );
        activityMetrics = results[2] as GetActivityMetricsResponse;
        cronBuckets = List.unmodifiable(
          (results[3] as GetCronExecutionTimeseriesResponse).buckets,
        );
        workspaceCounts = null;
        recentCronExecutions = const [];
      } else {
        overviewIsServerWide = false;
        serverOverview = null;
        activityMetrics = null;
        cronBuckets = const [];
        final results = await Future.wait([
          api.mcpServers(token, workspaceId),
          api.remoteAgents(token, workspaceId),
          api.listCronJobs(token, workspaceId),
          api.daemonRuntimes(token, workspaceId),
          api.listAgentInvocations(token, workspaceId, pageSize: 20),
          api.listCronExecutions(token, workspaceId, pageSize: 50),
          api.bridgeDiagnostics(token, workspaceId),
        ]);
        if (!live(gen) || seq != _overviewSeq) return;
        workspaceCounts = WorkspaceScopedCounts(
          agents: agentCount(),
          mcpServers: (results[0] as List).length,
          remoteAgents: (results[1] as List).length,
          cronJobs: (results[2] as List).length,
          daemonRuntimes: (results[3] as List).length,
        );
        final invPage = results[4] as ListAgentInvocationsResponse;
        activityEvents = List.unmodifiable(
          invPage.invocations.map(_invocationActivity),
        );
        recentCronExecutions = List.unmodifiable(
          (results[5]
                  as ({List<CronExecution> executions, String nextPageToken}))
              .executions,
        );
        serverHealth = results[6] as BridgeDiagnostics;
      }
      overviewState = ConfigLoad.ready;
    } catch (failure) {
      if (!live(gen) || seq != _overviewSeq) return;
      if (expired(failure)) return;
      overviewState = ConfigLoad.failed;
      overviewError = opsFailureMessage(failure, OpsArea.overview);
    }
    emit();
  }

  ActivityEvent _invocationActivity(Invocation inv) {
    return ActivityEvent(
      id: inv.id,
      kind: 'invocation',
      actor: inv.agentName.isEmpty ? inv.agentId : inv.agentName,
      message: '${invocationStatusLabel(inv.status)} · ${inv.sessionId}',
      timestamp: inv.hasStartedAt() ? inv.startedAt : null,
    );
  }

  // ---- Invocations ----

  ConfigLoad invocationsState = ConfigLoad.idle;
  String? invocationsError;
  List<Invocation> invocations = const [];
  String invocationsNextPage = '';
  int invocationsTotal = 0;
  String invocationFilterAgentId = '';
  String invocationFilterSessionId = '';
  Invocation? selectedInvocation;
  String? cancelMessage;
  bool cancelBusy = false;

  int _invSeq = 0;

  void setInvocationFilters({String? agentId, String? sessionId}) {
    invocationFilterAgentId = agentId?.trim() ?? '';
    invocationFilterSessionId = sessionId?.trim() ?? '';
    emit();
  }

  Future<void> loadInvocations({bool more = false}) async {
    if (disposed) return;
    final gen = epoch, seq = ++_invSeq;
    if (!more) {
      invocationsState = ConfigLoad.loading;
      invocations = const [];
      invocationsNextPage = '';
    }
    invocationsError = null;
    emit();
    try {
      final page = await api.listAgentInvocations(
        token,
        workspaceId,
        agentId: invocationFilterAgentId,
        sessionId: invocationFilterSessionId,
        pageSize: 30,
        pageToken: more ? invocationsNextPage : '',
      );
      if (!live(gen) || seq != _invSeq) return;
      invocations = List.unmodifiable([
        if (more) ...invocations,
        ...page.invocations,
      ]);
      invocationsNextPage = page.nextPageToken;
      invocationsTotal = page.total;
      invocationsState = ConfigLoad.ready;
    } catch (failure) {
      if (!live(gen) || seq != _invSeq) return;
      if (expired(failure)) return;
      invocationsState = ConfigLoad.failed;
      invocationsError = opsFailureMessage(failure, OpsArea.invocations);
    }
    emit();
  }

  Future<void> loadInvocationDetail(String id) async {
    if (disposed) return;
    final gen = epoch;
    cancelMessage = null;
    emit();
    try {
      selectedInvocation = await api.getAgentInvocation(token, workspaceId, id);
      if (!live(gen)) return;
    } catch (failure) {
      if (!live(gen)) return;
      if (expired(failure)) return;
      selectedInvocation = null;
      invocationsError = opsFailureMessage(failure, OpsArea.invocations);
    }
    emit();
  }

  Future<bool> cancelInvocation(String id) async {
    if (disposed || cancelBusy) return false;
    cancelBusy = true;
    cancelMessage = null;
    emit();
    final gen = epoch;
    try {
      final reply = await api.cancelAgentInvocation(token, workspaceId, id);
      if (!live(gen)) return false;
      cancelMessage = reply.cancelled
          ? 'Cancel was accepted by the server.'
          : 'The server did not stop this invocation. It may already have ended, '
                'run on another instance, or use an entry point this cancel path '
                'does not reach.';
      await loadInvocations();
      if (selectedInvocation?.id == id) {
        await loadInvocationDetail(id);
      }
      return reply.cancelled;
    } catch (failure) {
      if (!live(gen)) return false;
      if (expired(failure)) return false;
      cancelMessage = opsFailureMessage(failure, OpsArea.invocations);
      return false;
    } finally {
      cancelBusy = false;
      emit();
    }
  }

  // ---- Personal API tokens ----

  ConfigLoad tokenState = ConfigLoad.idle;
  String? tokenError;
  String? tokenWriteError;
  List<APIToken> personalTokens = const [];
  MintedApiSecret? minted;
  int _tokenSeq = 0;
  int _secretVisit = 0;

  Future<void> loadPersonalTokens() async {
    if (disposed) return;
    final gen = epoch, seq = ++_tokenSeq;
    tokenState = ConfigLoad.loading;
    tokenError = null;
    emit();
    try {
      final all = await api.apiTokens(token, workspaceId);
      if (!live(gen) || seq != _tokenSeq) return;
      personalTokens = List.unmodifiable([
        for (final t in all)
          if (isUserApiToken(t)) t,
      ]);
      tokenState = ConfigLoad.ready;
    } catch (failure) {
      if (!live(gen) || seq != _tokenSeq) return;
      if (expired(failure)) return;
      tokenState = ConfigLoad.failed;
      tokenError = opsFailureMessage(failure, OpsArea.apiToken);
    }
    emit();
  }

  void leaveTokenScreen({bool notify = true}) {
    _secretVisit++;
    clearMintedSecret(notify: notify);
  }

  void clearMintedSecret({bool notify = true}) {
    final m = minted;
    if (m == null) return;
    m.secret = null;
    minted = null;
    if (notify) emit();
  }

  Future<bool> createPersonalToken(String name) async {
    if (disposed || minted != null) return false;
    tokenWriteError = null;
    final visit = _secretVisit;
    final gen = epoch;
    emit();
    try {
      final reply = await api.createApiToken(token, workspaceId, name);
      if (!live(gen)) return false;
      if (visit != _secretVisit) {
        tokenWriteError =
            'A token was created, but you left the screen before its secret '
            'could be shown. Revoke the new token and create another.';
        await loadPersonalTokens();
        return false;
      }
      minted = MintedApiSecret(token: reply.token, secret: reply.secret);
      personalTokens = List.unmodifiable([...personalTokens, reply.token]);
      return true;
    } catch (failure) {
      if (!live(gen)) return false;
      if (expired(failure)) return false;
      tokenWriteError = opsFailureMessage(failure, OpsArea.apiToken);
      return false;
    } finally {
      emit();
    }
  }

  Future<bool> revokePersonalToken(String id) async {
    if (disposed) return false;
    tokenWriteError = null;
    final gen = epoch;
    emit();
    try {
      await api.revokeApiToken(token, workspaceId, id);
      if (!live(gen)) return false;
      personalTokens = List.unmodifiable([
        for (final t in personalTokens)
          if (t.id == id) (t.deepCopy()..revoked = true) else t,
      ]);
      return true;
    } catch (failure) {
      if (!live(gen)) return false;
      if (expired(failure)) return false;
      tokenWriteError = opsFailureMessage(failure, OpsArea.apiToken);
      return false;
    } finally {
      emit();
    }
  }
}
