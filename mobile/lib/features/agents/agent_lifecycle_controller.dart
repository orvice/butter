import 'dart:async';

import 'package:connectrpc/connect.dart';
import 'package:flutter/foundation.dart';

import '../../core/agent_errors.dart';
import '../../core/agent_lifecycle_errors.dart';
import '../../core/butter_api.dart';
import '../../core/config_resource_controller.dart';
import '../../gen/agents/v1/agent_operation.pb.dart';
import 'agent_content_draft.dart';

enum LifecycleLoad { idle, loading, ready, failed }

const lifecycleUnconfirmed = configUnconfirmed;

/// Agent lifecycle operations and Git content commits for one Workspace (#459).
class AgentLifecycleController extends ChangeNotifier {
  AgentLifecycleController({
    required ButterApi Function() newApi,
    required this.token,
    required this.workspaceId,
    required this.canManage,
    required this.onUnauthenticated,
    this.onAgentsChanged,
    this.onBindingChanged,
  }) : _newApi = newApi;

  final ButterApi Function() _newApi;
  final String token;
  final String workspaceId;
  final bool Function() canManage;
  final void Function() onUnauthenticated;
  final void Function()? onAgentsChanged;
  final void Function()? onBindingChanged;

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

  bool writing = false;
  String? writeError;
  bool unconfirmed = false;

  List<AgentOperation> operations = const [];
  LifecycleLoad operationsState = LifecycleLoad.idle;
  String? operationsError;
  String operationsNextPage = '';
  AgentOperationStatus? operationsFilter;
  int _opsSeq = 0;

  AgentOperation? operationDetail;
  LifecycleLoad detailState = LifecycleLoad.idle;
  String? detailError;
  int _detailSeq = 0;

  AgentContentDraft? contentDraft;
  LifecycleLoad contentState = LifecycleLoad.idle;
  String? contentError;
  String? contentValidationErrors;
  String? lastCommitSha;
  int _contentSeq = 0;

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

  void _writeFinished({
    required bool ok,
    String? error,
    bool ambiguous = false,
  }) {
    _lastWriteTick = ++_clock;
    writing = false;
    if (ok) {
      unconfirmed = false;
      writeError = null;
    } else if (error != null) {
      writeError = error;
      unconfirmed = false;
    } else if (ambiguous) {
      unconfirmed = true;
      writeError = lifecycleUnconfirmed;
    }
    _emit();
  }

  Future<void> loadOperations({bool more = false}) async {
    if (_disposed) return;
    final epoch = _epoch, seq = ++_opsSeq, started = ++_clock;
    if (!more) {
      operations = const [];
      operationsNextPage = '';
    }
    operationsState = LifecycleLoad.loading;
    operationsError = null;
    _emit();
    try {
      final reply = await _api.listAgentOperations(
        token,
        workspaceId,
        status:
            operationsFilter ??
            AgentOperationStatus.AGENT_OPERATION_STATUS_UNSPECIFIED,
        pageToken: more ? operationsNextPage : '',
      );
      if (!_live(epoch) || seq != _opsSeq || !_fresh(started)) return;
      operations = List.unmodifiable([...operations, ...reply.operations]);
      operationsNextPage = reply.nextPageToken;
      operationsState = LifecycleLoad.ready;
    } catch (failure) {
      if (!_live(epoch) || seq != _opsSeq) return;
      if (_expired(failure)) return;
      operationsState = LifecycleLoad.failed;
      operationsError =
          describeAgentLifecycleFailure(failure) ??
          'Could not load Agent operations.';
    }
    _emit();
  }

  Future<void> loadOperation(String id) async {
    if (_disposed) return;
    final epoch = _epoch, seq = ++_detailSeq, started = ++_clock;
    detailState = LifecycleLoad.loading;
    detailError = null;
    _emit();
    try {
      final op = await _api.getAgentOperation(token, workspaceId, id);
      if (!_live(epoch) || seq != _detailSeq || !_fresh(started)) return;
      operationDetail = op;
      detailState = LifecycleLoad.ready;
    } catch (failure) {
      if (!_live(epoch) || seq != _detailSeq) return;
      if (_expired(failure)) return;
      detailState = LifecycleLoad.failed;
      detailError =
          describeAgentLifecycleFailure(failure) ??
          'Could not load this operation.';
    }
    _emit();
  }

  Future<bool> retryOperation(String id) async {
    if (_disposed || writing || !canManage()) return false;
    final epoch = _epoch;
    _writeStarted();
    var ok = false;
    var ambiguous = false;
    String? validation;
    try {
      final reply = await _api.retryAgentOperation(token, workspaceId, id);
      if (!_live(epoch)) {
        _writeFinished(ok: false, ambiguous: true);
        return false;
      }
      if (reply.validationErrors.isNotEmpty) {
        validation = reply.validationErrors.join('\n');
      }
      operationDetail = reply.hasOperation()
          ? reply.operation
          : operationDetail;
      ok =
          reply.hasOperation() &&
          reply.operation.status ==
              AgentOperationStatus.AGENT_OPERATION_STATUS_SUCCEEDED;
      onAgentsChanged?.call();
      onBindingChanged?.call();
      unawaited(loadOperations());
    } catch (failure) {
      if (!_live(epoch)) return false;
      if (_expired(failure)) {
        _writeFinished(ok: false);
        return false;
      }
      ambiguous =
          failure is! ConnectException ||
          [
            Code.unavailable,
            Code.deadlineExceeded,
            Code.unknown,
          ].contains(failure.code);
      writeError =
          describeAgentLifecycleFailure(failure) ??
          'Retry failed. Refresh the operation before trying again.';
    }
    _writeFinished(
      ok: ok,
      error: validation ?? writeError,
      ambiguous: ambiguous,
    );
    return ok;
  }

  Future<bool> deleteAgent(
    String agentId, {
    required String displayName,
  }) async {
    if (_disposed || writing) return false;
    final epoch = _epoch;
    _writeStarted();
    var ok = false;
    var ambiguous = false;
    try {
      await _api.deleteAgent(token, workspaceId, agentId);
      if (!_live(epoch)) {
        _writeFinished(ok: false, ambiguous: true);
        return false;
      }
      ok = true;
      onAgentsChanged?.call();
    } catch (failure) {
      if (!_live(epoch)) return false;
      if (_expired(failure)) {
        _writeFinished(ok: false);
        return false;
      }
      ambiguous =
          failure is! ConnectException ||
          [
            Code.unavailable,
            Code.deadlineExceeded,
            Code.unknown,
          ].contains(failure.code);
      writeError =
          describeAgentLifecycleFailure(failure) ??
          describeAgentFailure(failure) ??
          'Could not tombstone Agent $displayName ($agentId).';
    }
    _writeFinished(ok: ok, ambiguous: ambiguous);
    return ok;
  }

  Future<bool> restoreAgent(String agentId) async {
    if (_disposed || writing || !canManage()) return false;
    final epoch = _epoch;
    _writeStarted();
    var ok = false;
    var ambiguous = false;
    try {
      final reply = await _api.restoreAgent(token, workspaceId, agentId);
      if (!_live(epoch)) {
        _writeFinished(ok: false, ambiguous: true);
        return false;
      }
      if (reply.validationErrors.isNotEmpty) {
        writeError = reply.validationErrors.join('\n');
      } else {
        ok = reply.hasAgent();
      }
      onAgentsChanged?.call();
      onBindingChanged?.call();
    } catch (failure) {
      if (!_live(epoch)) return false;
      if (_expired(failure)) {
        _writeFinished(ok: false);
        return false;
      }
      ambiguous =
          failure is! ConnectException ||
          [
            Code.unavailable,
            Code.deadlineExceeded,
            Code.unknown,
          ].contains(failure.code);
      writeError =
          describeAgentLifecycleFailure(failure) ??
          'Could not restore this Agent.';
    }
    _writeFinished(ok: ok, ambiguous: ambiguous);
    return ok;
  }

  Future<bool> purgeContent(String agentId) async {
    if (_disposed || writing || !canManage()) return false;
    final epoch = _epoch;
    _writeStarted();
    var ok = false;
    var ambiguous = false;
    try {
      final reply = await _api.purgeAgentContent(token, workspaceId, agentId);
      if (!_live(epoch)) {
        _writeFinished(ok: false, ambiguous: true);
        return false;
      }
      if (reply.validationErrors.isNotEmpty) {
        writeError = reply.validationErrors.join('\n');
      } else {
        ok = true;
        lastCommitSha = reply.commitSha;
      }
      onBindingChanged?.call();
    } catch (failure) {
      if (!_live(epoch)) return false;
      if (_expired(failure)) {
        _writeFinished(ok: false);
        return false;
      }
      ambiguous =
          failure is! ConnectException ||
          [
            Code.unavailable,
            Code.deadlineExceeded,
            Code.unknown,
          ].contains(failure.code);
      writeError =
          describeAgentLifecycleFailure(failure) ??
          'Could not purge Git content for this Agent.';
    }
    _writeFinished(ok: ok, ambiguous: ambiguous);
    return ok;
  }

  Future<void> loadContent(String agentId) async {
    if (_disposed) return;
    final epoch = _epoch, seq = ++_contentSeq, started = ++_clock;
    contentState = LifecycleLoad.loading;
    contentError = null;
    contentValidationErrors = null;
    _emit();
    try {
      final binding = await _api.repositoryBinding(token, workspaceId);
      final agent = await _api.agent(token, workspaceId, agentId);
      if (!_live(epoch) || seq != _contentSeq || !_fresh(started)) return;
      contentDraft = AgentContentDraft.fromAgent(
        agentId,
        agent,
        baseRevision: binding?.activeCommitSha ?? '',
      );
      contentState = LifecycleLoad.ready;
    } catch (failure) {
      if (!_live(epoch) || seq != _contentSeq) return;
      if (_expired(failure)) return;
      contentState = LifecycleLoad.failed;
      contentError =
          describeAgentLifecycleFailure(failure) ??
          'Could not load Agent Content.';
    }
    _emit();
  }

  Future<bool> commitContent() async {
    final draft = contentDraft;
    if (draft == null || _disposed || writing || !canManage() || !draft.dirty) {
      return false;
    }
    final actions = draft.commitActions();
    if (actions.isEmpty) return false;
    final epoch = _epoch;
    _writeStarted();
    var ok = false;
    var ambiguous = false;
    try {
      final reply = await _api.commitAgentContent(
        token,
        workspaceId,
        actions: actions,
        baseRevision: draft.baseRevision,
      );
      if (!_live(epoch)) {
        _writeFinished(ok: false, ambiguous: true);
        return false;
      }
      if (reply.validationErrors.isNotEmpty) {
        contentValidationErrors = reply.validationErrors.join('\n');
        writeError =
            'Server validation failed. Reload to compare with the repository; nothing was overwritten automatically.';
      } else {
        ok = true;
        lastCommitSha = reply.commitSha;
        draft.markSaved();
        contentValidationErrors = null;
        onBindingChanged?.call();
        onAgentsChanged?.call();
      }
    } catch (failure) {
      if (!_live(epoch)) return false;
      if (_expired(failure)) {
        _writeFinished(ok: false);
        return false;
      }
      ambiguous =
          failure is! ConnectException ||
          [
            Code.unavailable,
            Code.deadlineExceeded,
            Code.unknown,
          ].contains(failure.code);
      writeError =
          describeAgentLifecycleFailure(failure) ??
          'Commit failed. Refresh before retrying; writes are never auto-repeated.';
    }
    _writeFinished(ok: ok, ambiguous: ambiguous);
    return ok;
  }

  Future<bool> rollbackContent(String targetCommitSha) async {
    if (_disposed || writing || !canManage() || targetCommitSha.isEmpty) {
      return false;
    }
    final epoch = _epoch;
    _writeStarted();
    var ok = false;
    var ambiguous = false;
    try {
      final reply = await _api.rollbackAgentContent(
        token,
        workspaceId,
        targetCommitSha,
      );
      if (!_live(epoch)) {
        _writeFinished(ok: false, ambiguous: true);
        return false;
      }
      if (reply.validationErrors.isNotEmpty) {
        writeError = reply.validationErrors.join('\n');
      } else {
        ok = true;
        lastCommitSha = reply.commitSha;
        onBindingChanged?.call();
        onAgentsChanged?.call();
        if (contentDraft != null) {
          unawaited(loadContent(contentDraft!.agentId));
        }
      }
    } catch (failure) {
      if (!_live(epoch)) return false;
      if (_expired(failure)) {
        _writeFinished(ok: false);
        return false;
      }
      ambiguous =
          failure is! ConnectException ||
          [
            Code.unavailable,
            Code.deadlineExceeded,
            Code.unknown,
          ].contains(failure.code);
      writeError =
          describeAgentLifecycleFailure(failure) ??
          'Rollback failed. Refresh the binding before retrying.';
    }
    _writeFinished(ok: ok, ambiguous: ambiguous);
    return ok;
  }

  void suspend() {
    _epoch++;
    _apiOrNull?.close();
    _apiOrNull = null;
  }

  @override
  void dispose() {
    _disposed = true;
    _epoch++;
    super.dispose();
  }
}
