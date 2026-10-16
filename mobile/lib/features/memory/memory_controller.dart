import 'dart:async';

import 'package:connectrpc/connect.dart';
import 'package:flutter/foundation.dart';

import '../../core/butter_api.dart';
import '../../core/config_resource_controller.dart';
import '../../core/memory_errors.dart';
import '../../gen/agents/v1/workspace_memory.pb.dart';
import 'memory_draft.dart';

enum MemoryLoad { idle, loading, ready, failed }

const memoryUnconfirmed = configUnconfirmed;

class MemoryController extends ChangeNotifier {
  MemoryController({
    required ButterApi Function() newApi,
    required this.token,
    required this.workspaceId,
    required this.canManage,
    required this.onUnauthenticated,
  }) : _newApi = newApi;

  final ButterApi Function() _newApi;
  final String token;
  final String workspaceId;
  final bool Function() canManage;
  final void Function() onUnauthenticated;

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

  WorkspaceMemoryConfig? config;
  MemoryLoad configState = MemoryLoad.idle;
  String? configError;
  int _configSeq = 0;

  String? saveWarning;
  bool? testOk;
  String? testMessage;
  bool testing = false;

  MemoryLoad memoriesState = MemoryLoad.idle;
  String? memoriesError;
  List<WorkspaceMemory> memories = const [];
  bool memoriesTruncated = false;
  int memoriesLimit = 0;
  String memoryQuery = '';
  int _memSeq = 0;

  bool writing = false;
  String? writeError;
  bool unconfirmed = false;

  bool get configEnabled => config?.enabled ?? false;

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
    saveWarning = null;
    _emit();
  }

  void _writeFinished(bool ok, {String? error}) {
    _lastWriteTick = ++_clock;
    writing = false;
    if (ok) {
      unconfirmed = false;
      writeError = null;
    } else if (error != null) {
      writeError = error;
    } else {
      unconfirmed = true;
      writeError = memoryUnconfirmed;
    }
    _emit();
  }

  Future<void> loadConfig() async {
    final seq = ++_configSeq;
    final tick = ++_clock;
    configState = MemoryLoad.loading;
    configError = null;
    _emit();
    try {
      final loaded = await _api.workspaceMemoryConfig(token, workspaceId);
      if (!_live(_epoch) || seq != _configSeq || !_fresh(tick)) return;
      config = loaded;
      configState = MemoryLoad.ready;
      configError = null;
    } catch (failure) {
      if (!_live(_epoch) || seq != _configSeq || !_fresh(tick)) return;
      if (_expired(failure)) return;
      configState = MemoryLoad.failed;
      configError = memoryFailureMessage(failure, MemoryOperation.loadConfig);
    }
    _emit();
  }

  Future<bool> saveConfig(MemoryConfigDraft draft) async {
    if (!canManage()) {
      writeError = memoryFailureMessage(
        ConnectException(Code.permissionDenied, 'permission denied'),
        MemoryOperation.saveConfig,
      );
      _emit();
      return false;
    }
    _writeStarted();
    try {
      final result = await _api.putWorkspaceMemoryConfig(
        token,
        workspaceId,
        draft.buildPut(),
      );
      if (!_live(_epoch)) {
        _writeFinished(false);
        return false;
      }
      config = result.config;
      configState = MemoryLoad.ready;
      if (result.warning.isNotEmpty) {
        saveWarning =
            'Saved, but the connection check could not reach the mem0 server.';
      }
      draft.clearTypedSecret();
      _writeFinished(true);
      return true;
    } catch (failure) {
      if (!_live(_epoch)) return false;
      draft.clearTypedSecret();
      if (_expired(failure)) {
        _writeFinished(false, error: writeError);
        return false;
      }
      _writeFinished(
        false,
        error: memoryFailureMessage(failure, MemoryOperation.saveConfig),
      );
      return false;
    }
  }

  Future<bool> deleteConfig() async {
    if (!canManage()) return false;
    _writeStarted();
    try {
      await _api.deleteWorkspaceMemoryConfig(token, workspaceId);
      if (!_live(_epoch)) {
        _writeFinished(false);
        return false;
      }
      config = null;
      memories = const [];
      memoryQuery = '';
      _writeFinished(true);
      return true;
    } catch (failure) {
      if (!_live(_epoch)) return false;
      if (_expired(failure)) {
        _writeFinished(false);
        return false;
      }
      _writeFinished(
        false,
        error: memoryFailureMessage(failure, MemoryOperation.deleteConfig),
      );
      return false;
    }
  }

  Future<void> testConnection() async {
    testing = true;
    testOk = null;
    testMessage = null;
    _emit();
    final tick = ++_clock;
    try {
      final result = await _api.testWorkspaceMemoryConnection(
        token,
        workspaceId,
      );
      if (!_live(_epoch) || !_fresh(tick)) return;
      testOk = result.ok;
      testMessage = result.ok
          ? 'The mem0 server answered and accepted the API key.'
          : classifyMem0ProbeError(result.error);
    } catch (failure) {
      if (!_live(_epoch) || !_fresh(tick)) return;
      if (_expired(failure)) return;
      testOk = false;
      testMessage = memoryFailureMessage(
        failure,
        MemoryOperation.testConnection,
      );
    } finally {
      testing = false;
      _emit();
    }
  }

  Future<void> loadMemories({
    WorkspaceMemoryScope scope =
        WorkspaceMemoryScope.WORKSPACE_MEMORY_SCOPE_WORKSPACE,
    String agentId = '',
    String query = '',
  }) async {
    final seq = ++_memSeq;
    final tick = ++_clock;
    memoriesState = MemoryLoad.loading;
    memoriesError = null;
    _emit();
    try {
      final List<WorkspaceMemory> rows;
      if (query.trim().isEmpty) {
        final reply = await _api.listWorkspaceMemories(
          token,
          workspaceId,
          scope: scope,
          agentId: agentId,
        );
        rows = reply.memories;
        memoriesTruncated = reply.truncated;
        memoriesLimit = reply.limit;
      } else {
        final reply = await _api.searchWorkspaceMemories(
          token,
          workspaceId,
          query: query.trim(),
          scope: scope,
          agentId: agentId,
        );
        rows = reply.memories;
        memoriesTruncated = false;
        memoriesLimit = 0;
      }
      if (!_live(_epoch) || seq != _memSeq || !_fresh(tick)) return;
      memories = rows;
      memoryQuery = query;
      memoriesState = MemoryLoad.ready;
    } catch (failure) {
      if (!_live(_epoch) || seq != _memSeq || !_fresh(tick)) return;
      if (_expired(failure)) return;
      memoriesState = MemoryLoad.failed;
      memoriesError = memoryFailureMessage(
        failure,
        query.trim().isEmpty
            ? MemoryOperation.listMemories
            : MemoryOperation.searchMemories,
      );
    }
    _emit();
  }

  Future<bool> deleteMemory(String id) async {
    if (!canManage()) {
      writeError =
          'You need the owner or admin role in this Workspace to delete memories. Nothing was changed.';
      _emit();
      return false;
    }
    _writeStarted();
    try {
      await _api.deleteWorkspaceMemory(token, workspaceId, id);
      if (!_live(_epoch)) {
        _writeFinished(false);
        return false;
      }
      memories = memories.where((m) => m.id != id).toList();
      _writeFinished(true);
      return true;
    } catch (failure) {
      if (!_live(_epoch)) return false;
      if (_expired(failure)) {
        _writeFinished(false);
        return false;
      }
      _writeFinished(
        false,
        error: memoryFailureMessage(failure, MemoryOperation.deleteMemory),
      );
      return false;
    }
  }

  void suspend() {
    _epoch++;
    _apiOrNull = null;
    testing = false;
    writing = false;
    _emit();
  }

  Future<void> recover() async {
    await loadConfig();
    if (configEnabled) {
      await loadMemories();
    }
  }

  @override
  void dispose() {
    _disposed = true;
    _epoch++;
    super.dispose();
  }
}
