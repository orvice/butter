import 'dart:async';

import 'package:connectrpc/connect.dart';
import 'package:flutter/foundation.dart';

import '../../core/butter_api.dart';
import '../../core/linear_errors.dart';
import '../../core/native_oauth.dart';
import '../../gen/agents/v1/linear.pb.dart';
import 'linear_draft.dart';
import 'linear_install_flow.dart';

enum LinearLoad { idle, loading, ready, failed }

const linearUnconfirmed =
    'Change not confirmed; it may already have been applied. Refresh before retrying.';

/// Linear Apps, installations, platform URL status and processing records.
class LinearController extends ChangeNotifier {
  LinearController({
    required ButterApi Function() newApi,
    required this.token,
    required this.workspaceId,
    required this.canManage,
    required this.onUnauthenticated,
    NativeOAuthPlatform? browser,
    Duration? pollInterval,
    DateTime Function()? now,
  }) : _newApi = newApi,
       browser = browser ?? DeviceOAuthPlatform(),
       pollInterval = pollInterval ?? linearInstallPollInterval,
       _now = now ?? DateTime.now;

  final ButterApi Function() _newApi;
  final String token;
  final String workspaceId;
  final bool Function() canManage;
  final void Function() onUnauthenticated;
  final NativeOAuthPlatform browser;
  final Duration pollInterval;
  final DateTime Function() _now;

  ButterApi? _apiOrNull;
  ButterApi get _api => _apiOrNull ??= _newApi();
  bool _disposed = false;
  bool get disposed => _disposed;
  int _epoch = 0;
  bool _live(int epoch) => !_disposed && epoch == _epoch;
  void _emit() {
    if (!_disposed) notifyListeners();
  }

  int _clock = 0;
  int _lastWriteTick = 0;
  bool _fresh(int startTick) => startTick > _lastWriteTick;

  List<LinearApp> apps = const [];
  LinearLoad listState = LinearLoad.idle;
  String? listError;
  int _listSeq = 0;

  LinearApp? selected;
  LinearLoad detailState = LinearLoad.idle;
  String? detailError;
  int _detailSeq = 0;

  List<LinearInstallation> installations = const [];
  LinearLoad installListState = LinearLoad.idle;
  String? installListError;
  int _installSeq = 0;

  String? platformBaseUrl;
  bool platformSettingsDenied = false;
  LinearLoad platformState = LinearLoad.idle;

  List<LinearProcessingRecord> records = const [];
  LinearLoad recordsState = LinearLoad.idle;
  String? recordsError;
  int _recordsSeq = 0;
  LinearProcessingRecord? selectedRecord;

  bool writing = false;
  String? writeError;
  bool unconfirmed = false;

  LinearInstallView install = LinearInstallView();
  Timer? _installPoll;
  Timer? _installDeadline;
  int _installFlowSeq = 0;

  bool _expired(Object failure) {
    final expired =
        failure is ConnectException && failure.code == Code.unauthenticated;
    if (expired && !_disposed) onUnauthenticated();
    return expired;
  }

  void _writeStarted() {
    _lastWriteTick = ++_clock;
    writing = true;
    writeError = null;
    _emit();
  }

  void _writeFinished(bool ok, {String? error, bool unknown = false}) {
    _lastWriteTick = ++_clock;
    writing = false;
    if (ok) {
      unconfirmed = false;
      writeError = null;
    } else if (error != null) {
      writeError = error;
      unconfirmed = false;
    } else if (unknown) {
      unconfirmed = true;
      writeError = linearUnconfirmed;
    }
    _emit();
  }

  Future<void> loadApps() async {
    final seq = ++_listSeq;
    final tick = ++_clock;
    listState = LinearLoad.loading;
    listError = null;
    _emit();
    try {
      final items = await _api.listLinearApps(token, workspaceId);
      if (!_live(_epoch) || seq != _listSeq || !_fresh(tick)) return;
      apps = List.unmodifiable(items);
      listState = LinearLoad.ready;
    } catch (failure) {
      if (!_live(_epoch) || seq != _listSeq || !_fresh(tick)) return;
      if (_expired(failure)) return;
      listState = LinearLoad.failed;
      listError = linearFailureMessage(failure, LinearOperation.load);
    }
    _emit();
  }

  Future<void> loadPlatformSettings() async {
    platformState = LinearLoad.loading;
    platformSettingsDenied = false;
    _emit();
    try {
      final settings = await _api.getLinearSettings(token);
      platformBaseUrl = settings?.publicBaseUrl ?? '';
      platformState = LinearLoad.ready;
    } catch (failure) {
      if (_expired(failure)) return;
      if (failure is ConnectException &&
          failure.code == Code.permissionDenied) {
        platformSettingsDenied = true;
        platformBaseUrl = null;
        platformState = LinearLoad.ready;
      } else {
        platformState = LinearLoad.failed;
      }
    }
    _emit();
  }

  Future<void> loadApp(String id) async {
    final seq = ++_detailSeq;
    final tick = ++_clock;
    detailState = LinearLoad.loading;
    detailError = null;
    _emit();
    try {
      final app = await _api.getLinearApp(token, workspaceId, id);
      if (!_live(_epoch) || seq != _detailSeq || !_fresh(tick)) return;
      selected = app;
      detailState = LinearLoad.ready;
    } catch (failure) {
      if (!_live(_epoch) || seq != _detailSeq || !_fresh(tick)) return;
      if (_expired(failure)) return;
      detailState = LinearLoad.failed;
      detailError = linearFailureMessage(failure, LinearOperation.load);
    }
    _emit();
  }

  Future<bool> createApp(LinearAppDraft draft) async {
    if (writing || unconfirmed) return false;
    final err = draft.validate(creating: true);
    if (err != null) {
      writeError = err;
      _emit();
      return false;
    }
    _writeStarted();
    try {
      final created = await _api.createLinearApp(
        token,
        workspaceId,
        draft.buildForCreate(),
        clientSecret: draft.clientSecret.trim(),
        webhookSecret: draft.webhookSecret.trim(),
      );
      if (!_live(_epoch)) return false;
      selected = created;
      _writeFinished(true);
      await loadApps();
      return true;
    } catch (failure) {
      if (!_live(_epoch)) return false;
      if (_expired(failure)) return false;
      _writeFinished(
        false,
        error: linearFailureMessage(failure, LinearOperation.create),
        unknown: linearWriteOutcomeUnknown(failure),
      );
      return false;
    }
  }

  Future<bool> updateApp(LinearAppDraft draft) async {
    if (writing || unconfirmed) return false;
    final err = draft.validate(creating: false);
    if (err != null) {
      writeError = err;
      _emit();
      return false;
    }
    _writeStarted();
    try {
      final updated = await _api.updateLinearApp(
        token,
        workspaceId,
        draft.buildForUpdate(),
      );
      if (!_live(_epoch)) return false;
      selected = updated;
      _writeFinished(true);
      await loadApps();
      return true;
    } catch (failure) {
      if (!_live(_epoch)) return false;
      if (_expired(failure)) return false;
      _writeFinished(
        false,
        error: linearFailureMessage(failure, LinearOperation.update),
        unknown: linearWriteOutcomeUnknown(failure),
      );
      return false;
    }
  }

  Future<bool> putCredentials(
    String appId, {
    String? clientSecret,
    String? webhookSecret,
  }) async {
    if (writing || unconfirmed) return false;
    _writeStarted();
    try {
      final updated = await _api.putLinearAppCredentials(
        token,
        workspaceId,
        appId,
        clientSecret: clientSecret,
        webhookSecret: webhookSecret,
      );
      if (!_live(_epoch)) return false;
      if (selected?.id == appId) selected = updated;
      _writeFinished(true);
      await loadApps();
      return true;
    } catch (failure) {
      if (!_live(_epoch)) return false;
      if (_expired(failure)) return false;
      _writeFinished(
        false,
        error: linearFailureMessage(failure, LinearOperation.credentials),
        unknown: linearWriteOutcomeUnknown(failure),
      );
      return false;
    }
  }

  Future<bool> deleteApp(String id) async {
    if (writing || unconfirmed) return false;
    _writeStarted();
    try {
      await _api.deleteLinearApp(token, workspaceId, id);
      if (!_live(_epoch)) return false;
      if (selected?.id == id) selected = null;
      _writeFinished(true);
      await loadApps();
      return true;
    } catch (failure) {
      if (!_live(_epoch)) return false;
      if (_expired(failure)) return false;
      _writeFinished(
        false,
        error: linearFailureMessage(failure, LinearOperation.delete),
        unknown: linearWriteOutcomeUnknown(failure),
      );
      return false;
    }
  }

  Future<void> loadInstallations(String appId) async {
    final seq = ++_installSeq;
    final tick = ++_clock;
    installListState = LinearLoad.loading;
    installListError = null;
    _emit();
    try {
      final items = await _api.listLinearInstallations(
        token,
        workspaceId,
        appId,
      );
      if (!_live(_epoch) || seq != _installSeq || !_fresh(tick)) return;
      installations = List.unmodifiable(items);
      installListState = LinearLoad.ready;
    } catch (failure) {
      if (!_live(_epoch) || seq != _installSeq || !_fresh(tick)) return;
      if (_expired(failure)) return;
      installListState = LinearLoad.failed;
      installListError = linearFailureMessage(failure, LinearOperation.load);
    }
    _emit();
  }

  Future<bool> deleteInstallation(String appId, String installationId) async {
    if (writing || unconfirmed) return false;
    _writeStarted();
    try {
      await _api.deleteLinearInstallation(
        token,
        workspaceId,
        appId,
        installationId,
      );
      if (!_live(_epoch)) return false;
      _writeFinished(true);
      await loadInstallations(appId);
      return true;
    } catch (failure) {
      if (!_live(_epoch)) return false;
      if (_expired(failure)) return false;
      _writeFinished(
        false,
        error: linearFailureMessage(failure, LinearOperation.delete),
        unknown: linearWriteOutcomeUnknown(failure),
      );
      return false;
    }
  }

  Future<void> loadRecords({
    String appId = '',
    LinearProcessingStatus? status,
  }) async {
    final seq = ++_recordsSeq;
    final tick = ++_clock;
    recordsState = LinearLoad.loading;
    recordsError = null;
    _emit();
    try {
      final items = status == null
          ? await _api.listLinearProcessingRecords(
              token,
              workspaceId,
              appId: appId,
            )
          : await _api.listLinearProcessingRecords(
              token,
              workspaceId,
              appId: appId,
              status: status,
            );
      if (!_live(_epoch) || seq != _recordsSeq || !_fresh(tick)) return;
      records = List.unmodifiable(items);
      recordsState = LinearLoad.ready;
    } catch (failure) {
      if (!_live(_epoch) || seq != _recordsSeq || !_fresh(tick)) return;
      if (_expired(failure)) return;
      recordsState = LinearLoad.failed;
      recordsError = linearFailureMessage(failure, LinearOperation.load);
    }
    _emit();
  }

  Future<void> loadRecord(String id) async {
    try {
      selectedRecord = await _api.getLinearProcessingRecord(
        token,
        workspaceId,
        id,
      );
      _emit();
    } catch (failure) {
      if (_expired(failure)) return;
      recordsError = linearFailureMessage(failure, LinearOperation.load);
      _emit();
    }
  }

  Future<bool> resendReply(String recordId) async {
    if (writing || unconfirmed) return false;
    _writeStarted();
    try {
      final updated = await _api.resendLinearReply(
        token,
        workspaceId,
        recordId,
      );
      if (!_live(_epoch)) return false;
      selectedRecord = updated;
      _writeFinished(true);
      await loadRecords();
      return true;
    } catch (failure) {
      if (!_live(_epoch)) return false;
      if (_expired(failure)) return false;
      _writeFinished(
        false,
        error: linearFailureMessage(failure, LinearOperation.resend),
        unknown: linearWriteOutcomeUnknown(failure),
      );
      return false;
    }
  }

  void _clearInstallTimers() {
    _installPoll?.cancel();
    _installPoll = null;
    _installDeadline?.cancel();
    _installDeadline = null;
  }

  void _setInstallPhase(
    LinearInstallPhase phase, {
    String? message,
    required int seq,
  }) {
    if (seq != _installFlowSeq || _disposed) return;
    install.phase = phase;
    install.message = message;
    _emit();
  }

  Future<void> _pollInstallations(int seq) async {
    if (seq != _installFlowSeq || _disposed) return;
    final appId = install.appId;
    try {
      final items = await _api.listLinearInstallations(
        token,
        workspaceId,
        appId,
      );
      if (seq != _installFlowSeq || _disposed) return;
      installations = List.unmodifiable(items);
      final newOnes = items
          .where((i) => !install.baselineInstallationIds.contains(i.id))
          .toList();
      if (newOnes.isNotEmpty) {
        _clearInstallTimers();
        _setInstallPhase(
          LinearInstallPhase.installed,
          message: 'Installed into ${newOnes.first.organizationName}.',
          seq: seq,
        );
        return;
      }
    } catch (_) {
      // Transient read failures do not end the wait.
    }
    if (seq != _installFlowSeq || _disposed) return;
    _installPoll = Timer(
      pollInterval,
      () => unawaited(_pollInstallations(seq)),
    );
  }

  Future<void> beginInstall(String appId) async {
    if (!canManage()) {
      writeError = linearFailureMessage(
        ConnectException(Code.permissionDenied, 'denied'),
        LinearOperation.install,
      );
      _emit();
      return;
    }
    _clearInstallTimers();
    final seq = ++_installFlowSeq;
    install = LinearInstallView(
      phase: LinearInstallPhase.starting,
      appId: appId,
      startedAt: _now(),
      deadline: _now().add(linearInstallFlowTtl),
    )..sequence = seq;
    _emit();
    await loadInstallations(appId);
    if (seq != _installFlowSeq) return;
    install.baselineInstallationIds = installations.map((i) => i.id).toSet();
    try {
      final url = await _api.beginLinearInstall(
        token,
        workspaceId,
        appId,
        returnUrl: '/linear-apps/$appId',
      );
      if (seq != _installFlowSeq) return;
      final opened = await openLinearAuthorizeUrl(url, platform: browser);
      if (seq != _installFlowSeq) return;
      if (!opened) {
        _setInstallPhase(
          LinearInstallPhase.failed,
          message: 'Could not open the system browser.',
          seq: seq,
        );
        return;
      }
      install.phase = LinearInstallPhase.waiting;
      install.message =
          'Finish approval in the browser, then return here. Waiting up to ${linearInstallFlowTtl.inMinutes} minutes.';
      _emit();
      _installDeadline = Timer(linearInstallFlowTtl, () async {
        if (seq != _installFlowSeq) return;
        _clearInstallTimers();
        await loadInstallations(appId);
        final newOnes = installations
            .where((i) => !install.baselineInstallationIds.contains(i.id))
            .toList();
        if (newOnes.isNotEmpty) {
          _setInstallPhase(
            LinearInstallPhase.installed,
            message: 'Installed into ${newOnes.first.organizationName}.',
            seq: seq,
          );
        } else {
          _setInstallPhase(
            LinearInstallPhase.timedOut,
            message:
                'No new installation appeared within ${linearInstallFlowTtl.inMinutes} minutes. If you finished in the browser, refresh installations.',
            seq: seq,
          );
        }
      });
      unawaited(_pollInstallations(seq));
    } catch (failure) {
      if (seq != _installFlowSeq) return;
      if (_expired(failure)) return;
      _clearInstallTimers();
      _setInstallPhase(
        LinearInstallPhase.failed,
        message: linearFailureMessage(failure, LinearOperation.install),
        seq: seq,
      );
    }
  }

  void stopWaitingForInstall() {
    _clearInstallTimers();
    final seq = install.sequence;
    _setInstallPhase(
      LinearInstallPhase.cancelled,
      message:
          'Stopped waiting. The authorization link may still work in the browser; refresh installations to see whether it finished.',
      seq: seq,
    );
  }

  void recoverInstallWait() {
    final seq = install.sequence;
    if (install.phase != LinearInstallPhase.waiting) return;
    if (_now().isAfter(install.deadline)) {
      stopWaitingForInstall();
      install.phase = LinearInstallPhase.timedOut;
      install.message =
          'The install wait expired. Refresh installations to see whether Linear finished.';
      _emit();
      return;
    }
    unawaited(_pollInstallations(seq));
  }

  void suspend() {
    _epoch++;
    _clearInstallTimers();
    _apiOrNull?.close();
    _apiOrNull = null;
    apps = const [];
    selected = null;
    installations = const [];
    records = const [];
    listState = LinearLoad.idle;
    detailState = LinearLoad.idle;
    install = LinearInstallView();
    writing = false;
    unconfirmed = false;
    writeError = null;
  }

  Future<void> recover() async {
    if (install.phase == LinearInstallPhase.waiting) {
      recoverInstallWait();
    }
    if (listState == LinearLoad.ready) {
      await loadApps();
    }
  }

  @override
  void dispose() {
    _disposed = true;
    _epoch++;
    _clearInstallTimers();
    _apiOrNull?.close();
    super.dispose();
  }
}
