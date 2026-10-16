import 'dart:async';

import '../../core/butter_api.dart';
import '../../core/config_errors.dart';
import '../../core/config_resource_controller.dart';
import '../../core/cron_errors.dart';
import '../../gen/agents/v1/agent.pb.dart';
import '../../gen/agents/v1/cron.pb.dart';
import '../../gen/agents/v1/telegram.pb.dart';
import '../sessions/session_controller.dart';
import '../sessions/session_key.dart';

/// Cron jobs and executions for one server, user and Workspace.
class CronController extends ConfigResourceController<CronJob> {
  CronController({
    required super.newApi,
    required super.token,
    required super.workspaceId,
    required super.onUnauthenticated,
    required this.sessions,
    required this.agents,
  }) : super(kind: ConfigKind.cron);

  /// Authorized ReplySession seam (#463).
  final SessionController Function() sessions;

  /// Agent list for pickers (immutable IDs).
  final List<Agent> Function() agents;

  @override
  String idOf(CronJob item) => item.name;

  @override
  Future<List<CronJob>> readAll(ButterApi api) =>
      api.listCronJobs(token, workspaceId);

  @override
  Future<CronJob> readOne(ButterApi api, String id) =>
      api.getCronJob(token, workspaceId, id);

  @override
  Future<CronJob> createOne(ButterApi api, CronJob draft) =>
      api.createCronJob(token, workspaceId, draft);

  @override
  Future<CronJob> updateOne(ButterApi api, CronJob draft) =>
      api.updateCronJob(token, workspaceId, draft);

  @override
  Future<void> deleteOne(ButterApi api, String id) =>
      api.deleteCronJob(token, workspaceId, id);

  @override
  String writeFailureMessage(
    Object failure,
    ConfigOperation operation,
    String scope,
  ) {
    if (scope == 'run') {
      return cronFailureMessage(failure, CronOperation.run);
    }
    final cronOp = switch (operation) {
      ConfigOperation.create => CronOperation.create,
      ConfigOperation.update => CronOperation.update,
      ConfigOperation.delete => CronOperation.delete,
      _ => CronOperation.load,
    };
    return cronFailureMessage(failure, cronOp);
  }

  // ---- Reference lists for delivery pickers ----

  List<NotifyGroup> notifyGroups = const [];
  List<TelegramDestination> telegramDestinations = const [];
  ConfigLoad referenceState = ConfigLoad.idle;
  String? referenceError;
  int _referenceSeq = 0;

  Future<void> loadReferences() async {
    if (disposed) return;
    final epoch = this.epoch, seq = ++_referenceSeq;
    referenceState = ConfigLoad.loading;
    referenceError = null;
    emit();
    try {
      final results = await Future.wait([
        api.listNotifyGroups(token, workspaceId),
        api.listTelegramDestinations(token, workspaceId),
      ]);
      if (!live(epoch) || seq != _referenceSeq) return;
      notifyGroups = List.unmodifiable(results[0] as List<NotifyGroup>);
      telegramDestinations = List.unmodifiable(
        results[1] as List<TelegramDestination>,
      );
      referenceState = ConfigLoad.ready;
    } catch (failure) {
      if (!live(epoch) || seq != _referenceSeq) return;
      if (expired(failure)) return;
      referenceState = ConfigLoad.failed;
      referenceError = cronFailureMessage(failure, CronOperation.load);
    }
    emit();
  }

  // ---- Executions ----

  List<CronExecution> executions = const [];
  ConfigLoad executionState = ConfigLoad.idle;
  String? executionError;
  String executionNextToken = '';
  String? _executionJob;
  int _executionSeq = 0;

  CronExecution? selectedExecution;

  Future<void> loadExecutions(String jobName, {bool more = false}) async {
    if (disposed) return;
    final epoch = this.epoch, seq = ++_executionSeq;
    final started = beginRead();
    if (!more) {
      _executionJob = jobName;
      executions = const [];
      executionNextToken = '';
    }
    executionState = ConfigLoad.loading;
    executionError = null;
    emit();
    try {
      final page = await api.listCronExecutions(
        token,
        workspaceId,
        jobName: jobName,
        pageSize: 20,
        pageToken: more ? executionNextToken : '',
      );
      if (!live(epoch) || seq != _executionSeq) return;
      if (!readIsFresh(started)) {
        executionState = executions.isEmpty
            ? ConfigLoad.idle
            : ConfigLoad.ready;
        emit();
        return;
      }
      executions = List.unmodifiable([...executions, ...page.executions]);
      executionNextToken = page.nextPageToken;
      executionState = ConfigLoad.ready;
      reconcile(started, id: jobName);
      if (_runUnconfirmed == jobName) {
        _runUnconfirmed = null;
      }
    } catch (failure) {
      if (!live(epoch) || seq != _executionSeq) return;
      if (!readIsFresh(started)) {
        executionState = executions.isEmpty
            ? ConfigLoad.idle
            : ConfigLoad.ready;
        emit();
        return;
      }
      if (expired(failure)) return;
      executionState = ConfigLoad.failed;
      executionError = cronFailureMessage(failure, CronOperation.executions);
    }
    emit();
  }

  /// Re-reads one execution by id from the current list page, then the server.
  Future<void> refreshExecution(CronExecution exec) async {
    if (_executionJob == null) return;
    await loadExecutions(_executionJob!);
    final match = executions.where((e) => e.id == exec.id).firstOrNull;
    selectedExecution = match ?? exec;
    emit();
  }

  // ---- Run now (never auto-retried) ----

  bool runSending = false;
  String? runError;
  String? _runUnconfirmed;
  CronExecution? lastRun;

  bool get runBlocked => runSending || _runUnconfirmed != null;

  Future<bool> runNow(String jobName) async {
    if (disposed || runBlocked || unconfirmed) return false;
    final epoch = this.epoch;
    runSending = true;
    runError = null;
    emit();
    try {
      final exec = await api.runCronJobNow(token, workspaceId, jobName);
      if (!live(epoch)) {
        runSending = false;
        return false;
      }
      _runUnconfirmed = null;
      runSending = false;
      lastRun = exec;
      if (_executionJob == jobName) {
        unawaited(loadExecutions(jobName));
      }
      emit();
      return true;
    } catch (failure) {
      if (!live(epoch)) return false;
      runSending = false;
      if (expired(failure)) return false;
      runError = cronFailureMessage(failure, CronOperation.run);
      if (cronWriteOutcomeUnknown(failure)) {
        _runUnconfirmed = jobName;
      }
      emit();
      return false;
    }
  }

  void clearRunUnconfirmed(String jobName) {
    if (_runUnconfirmed == jobName) {
      _runUnconfirmed = null;
      emit();
    }
  }

  Future<bool> toggleEnabled(CronJob job, bool enabled) async {
    final draft = job.deepCopy()..enabled = enabled;
    return update(draft);
  }

  // ---- WAITING_INPUT reply via SessionController ----

  Future<SessionReplyResult> replyToExecution({
    required CronExecution execution,
    required String message,
  }) async {
    if (execution.status !=
        CronExecutionStatus.CRON_EXECUTION_STATUS_WAITING_INPUT) {
      return const SessionReplyResult.failed(
        'This execution is not waiting for input.',
      );
    }
    final agentId = execution.agentId;
    if (agentId.isEmpty) {
      return const SessionReplyResult.failed('Execution has no Agent ID.');
    }
    final key = SessionKey(
      appName: execution.sessionAppName,
      userId: execution.sessionUserId,
      sessionId: execution.sessionId,
    );
    final result = await sessions().reply(
      key: key,
      agentId: agentId,
      message: message,
    );
    if (!result.unconfirmed &&
        result.error == null &&
        _executionJob == execution.jobName) {
      unawaited(loadExecutions(execution.jobName));
    }
    return result;
  }

  @override
  void onSuspend() {
    executions = const [];
    executionState = ConfigLoad.idle;
    selectedExecution = null;
    notifyGroups = const [];
    telegramDestinations = const [];
    referenceState = ConfigLoad.idle;
    lastRun = null;
    if (runSending) {
      _runUnconfirmed = _executionJob;
    }
    runSending = false;
  }

  @override
  Future<void> onRecover() async {
    if (_executionJob != null) {
      await loadExecutions(_executionJob!);
    }
    await loadReferences();
  }
}
