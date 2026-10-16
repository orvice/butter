import '../../core/automation_errors.dart';
import '../../core/butter_api.dart';
import '../../core/config_errors.dart';
import '../../core/config_resource_controller.dart';
import '../../gen/agents/v1/agent.pb.dart';
import '../../gen/agents/v1/automation.pb.dart';

class AutomationController extends ConfigResourceController<Automation> {
  AutomationController({
    required super.newApi,
    required super.token,
    required super.workspaceId,
    required super.onUnauthenticated,
    required this.agents,
  }) : super(kind: ConfigKind.automation);

  final List<Agent> Function() agents;

  @override
  String idOf(Automation item) => item.name;

  @override
  Future<List<Automation>> readAll(ButterApi api) =>
      api.listAutomations(token, workspaceId);

  @override
  Future<Automation> readOne(ButterApi api, String id) =>
      api.getAutomation(token, workspaceId, id);

  @override
  Future<Automation> createOne(ButterApi api, Automation draft) =>
      api.createAutomation(token, workspaceId, draft);

  @override
  Future<Automation> updateOne(ButterApi api, Automation draft) =>
      api.updateAutomation(token, workspaceId, draft);

  @override
  Future<void> deleteOne(ButterApi api, String id) =>
      api.deleteAutomation(token, workspaceId, id);

  @override
  String writeFailureMessage(
    Object failure,
    ConfigOperation operation,
    String scope,
  ) {
    if (scope == 'run') {
      return automationFailureMessage(failure, AutomationOperation.run);
    }
    final op = switch (operation) {
      ConfigOperation.create => AutomationOperation.create,
      ConfigOperation.update => AutomationOperation.update,
      ConfigOperation.delete => AutomationOperation.delete,
      _ => AutomationOperation.load,
    };
    return automationFailureMessage(failure, op);
  }

  List<NotifyGroup> notifyGroups = const [];
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
      notifyGroups = List.unmodifiable(
        await api.listNotifyGroups(token, workspaceId),
      );
      if (!live(epoch) || seq != _referenceSeq) return;
      referenceState = ConfigLoad.ready;
    } catch (failure) {
      if (!live(epoch) || seq != _referenceSeq) return;
      if (expired(failure)) return;
      referenceState = ConfigLoad.failed;
      referenceError = automationFailureMessage(
        failure,
        AutomationOperation.load,
      );
    }
    emit();
  }

  List<AutomationRun> runs = const [];
  ConfigLoad runState = ConfigLoad.idle;
  String? runsLoadError;
  String runNextToken = '';
  String? _runAutomation;
  int _runSeq = 0;

  List<AutomationStepRun> stepRuns = const [];
  ConfigLoad stepState = ConfigLoad.idle;
  String? stepError;
  String? _stepRunId;
  int _stepSeq = 0;

  AutomationRun? selectedRun;

  Future<void> loadRuns(String automationName, {bool more = false}) async {
    if (disposed) return;
    final epoch = this.epoch, seq = ++_runSeq;
    final started = beginRead();
    if (!more) {
      _runAutomation = automationName;
      runs = const [];
      runNextToken = '';
    }
    runState = ConfigLoad.loading;
    runsLoadError = null;
    emit();
    try {
      final page = await api.listAutomationRuns(
        token,
        workspaceId,
        automationName: automationName,
        pageToken: more ? runNextToken : '',
      );
      if (!live(epoch) || seq != _runSeq) return;
      if (!readIsFresh(started)) {
        runState = runs.isEmpty ? ConfigLoad.idle : ConfigLoad.ready;
        emit();
        return;
      }
      runs = List.unmodifiable([...runs, ...page.runs]);
      runNextToken = page.nextPageToken;
      runState = ConfigLoad.ready;
      reconcile(started, id: automationName);
      if (_runUnconfirmed == automationName) {
        _runUnconfirmed = null;
      }
    } catch (failure) {
      if (!live(epoch) || seq != _runSeq) return;
      if (!readIsFresh(started)) {
        runState = runs.isEmpty ? ConfigLoad.idle : ConfigLoad.ready;
        emit();
        return;
      }
      if (expired(failure)) return;
      runState = ConfigLoad.failed;
      runsLoadError = automationFailureMessage(
        failure,
        AutomationOperation.runs,
      );
    }
    emit();
  }

  Future<void> loadStepRuns(String runId) async {
    if (disposed) return;
    final epoch = this.epoch, seq = ++_stepSeq;
    _stepRunId = runId;
    stepState = ConfigLoad.loading;
    stepError = null;
    emit();
    try {
      final steps = await api.listAutomationStepRuns(token, workspaceId, runId);
      if (!live(epoch) || seq != _stepSeq) return;
      stepRuns = List.unmodifiable(steps);
      stepState = ConfigLoad.ready;
    } catch (failure) {
      if (!live(epoch) || seq != _stepSeq) return;
      if (expired(failure)) return;
      stepState = ConfigLoad.failed;
      stepError = automationFailureMessage(failure, AutomationOperation.steps);
    }
    emit();
  }

  Future<AutomationRun?> pollRunUntilSettled(
    String runId, {
    int attempts = 30,
  }) async {
    for (var i = 0; i < attempts; i++) {
      if (disposed) return null;
      try {
        final run = await api.getAutomationRun(token, workspaceId, runId);
        if (run.status != AutomationRunStatus.AUTOMATION_RUN_STATUS_RUNNING &&
            run.status !=
                AutomationRunStatus.AUTOMATION_RUN_STATUS_UNSPECIFIED) {
          return run;
        }
      } catch (_) {}
      await Future<void>.delayed(const Duration(milliseconds: 200));
    }
    return null;
  }

  bool runSending = false;
  String? runError;
  String? _runUnconfirmed;
  AutomationRun? lastRun;

  bool get runBlocked => runSending || _runUnconfirmed != null;

  Future<bool> runNow(String name, {String triggerPayloadJson = '{}'}) async {
    if (disposed || runBlocked || unconfirmed) return false;
    final epoch = this.epoch;
    runSending = true;
    runError = null;
    emit();
    try {
      final run = await api.runAutomationNow(
        token,
        workspaceId,
        name,
        triggerPayloadJson: triggerPayloadJson,
      );
      if (!live(epoch)) {
        runSending = false;
        return false;
      }
      _runUnconfirmed = null;
      runSending = false;
      lastRun = run;
      if (_runAutomation == name) {
        runs = List.unmodifiable([run, ...runs]);
      }
      emit();
      return true;
    } catch (failure) {
      if (!live(epoch)) return false;
      if (expired(failure)) return false;
      runSending = false;
      runError = automationFailureMessage(failure, AutomationOperation.run);
      if (automationWriteOutcomeUnknown(failure)) {
        _runUnconfirmed = name;
      }
      emit();
      return false;
    }
  }

  @override
  void onSuspend() {
    notifyGroups = const [];
    referenceState = ConfigLoad.idle;
    runs = const [];
    stepRuns = const [];
    runState = ConfigLoad.idle;
    stepState = ConfigLoad.idle;
  }

  @override
  Future<void> onRecover() async {
    await loadReferences();
    if (_runAutomation != null) {
      await loadRuns(_runAutomation!);
    }
    if (_stepRunId != null) {
      await loadStepRuns(_stepRunId!);
    }
  }
}
