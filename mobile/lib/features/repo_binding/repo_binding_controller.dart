import 'dart:async';

import 'package:connectrpc/connect.dart';
import 'package:flutter/foundation.dart';

import '../../core/butter_api.dart';
import '../../core/repo_binding_errors.dart';
import '../../gen/agents/v1/githost.pb.dart';
import '../../gen/agents/v1/repobinding.pb.dart';
import 'repo_binding_draft.dart';

enum RepoBindingLoad { idle, loading, ready, failed }

const repoBindingUnconfirmed =
    'Change not confirmed; it may already have been applied. Refresh before retrying.';

/// Git repository binding for one server, user and Workspace (#458).
///
/// Exposes binding state for Agent Content (#459): [agentContentGitOwned] is
/// true once the binding has an Active Revision (`active_commit_sha`).
class RepoBindingController extends ChangeNotifier {
  RepoBindingController({
    required ButterApi Function() newApi,
    required this.token,
    required this.workspaceId,
    required this.canManage,
    required this.onUnauthenticated,
    this.onBindingChanged,
  }) : _newApi = newApi;

  final ButterApi Function() _newApi;
  final String token;
  final String workspaceId;
  final bool Function() canManage;
  final void Function() onUnauthenticated;

  /// Called after onboarding, publish, sync or detach changes Git ownership.
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

  WorkspaceRepoBinding? binding;
  List<RepoBindingOverlap> overlaps = const [];
  RepoBindingLoad loadState = RepoBindingLoad.idle;
  String? loadError;
  int _loadSeq = 0;

  List<GitHost> gitHosts = const [];
  RepoBindingLoad hostsState = RepoBindingLoad.idle;
  String? hostsError;

  bool writing = false;
  String? writeError;
  bool unconfirmed = false;
  String? _unconfirmedScope;

  /// One-time webhook secret from the last ConfigureWebhookSecret answer.
  WebhookSecret? webhookSecret;

  /// Cached tree browser state (DB cache only).
  String browsePath = '';
  List<RepoCacheEntry> entries = const [];
  RepoBindingLoad entriesState = RepoBindingLoad.idle;
  String? entriesError;
  String? selectedFilePath;
  String? fileContent;
  RepoBindingLoad fileState = RepoBindingLoad.idle;
  String? fileError;

  bool get hasBinding => binding != null;
  bool get hasActiveRevision =>
      binding != null && binding!.activeCommitSha.isNotEmpty;

  /// Whether Agent Content edits must use Git paths (#459 seam).
  bool get agentContentGitOwned => hasActiveRevision;

  bool _expired(Object failure) {
    final expired =
        failure is ConnectException && failure.code == Code.unauthenticated;
    if (expired && !_disposed) onUnauthenticated();
    return expired;
  }

  void _writeStarted(String scope) {
    _lastWriteTick = ++_clock;
    writing = true;
    writeError = null;
    _unconfirmedScope = scope;
    _emit();
  }

  void _writeFinished(bool ok, {String? error}) {
    _lastWriteTick = ++_clock;
    writing = false;
    if (ok) {
      unconfirmed = false;
      _unconfirmedScope = null;
      writeError = null;
    } else if (error != null) {
      writeError = error;
    } else {
      unconfirmed = true;
      writeError = repoBindingUnconfirmed;
    }
    _emit();
  }

  void _reconcile(int readStart, {String? scope}) {
    if (!unconfirmed || _unconfirmedScope == null) return;
    if (!_fresh(readStart)) return;
    if (scope != null && scope != _unconfirmedScope) return;
    unconfirmed = false;
    _unconfirmedScope = null;
    writeError = null;
    _emit();
  }

  Future<void> load() async {
    if (_disposed) return;
    final epoch = _epoch, seq = ++_loadSeq, started = ++_clock;
    final before = loadState == RepoBindingLoad.loading
        ? RepoBindingLoad.idle
        : loadState;
    loadState = RepoBindingLoad.loading;
    loadError = null;
    _emit();
    try {
      final snap = await _api.workspaceRepoBinding(token, workspaceId);
      if (!_live(epoch) || seq != _loadSeq) return;
      if (!_fresh(started)) {
        loadState = before;
        _emit();
        return;
      }
      binding = snap.binding;
      overlaps = List.unmodifiable(snap.overlaps);
      loadState = RepoBindingLoad.ready;
      _reconcile(started, scope: 'binding');
    } catch (failure) {
      if (!_live(epoch) || seq != _loadSeq) return;
      if (!_fresh(started)) {
        loadState = before;
        _emit();
        return;
      }
      if (_expired(failure)) return;
      loadState = RepoBindingLoad.failed;
      loadError = repoBindingFailureMessage(failure, RepoBindingOperation.load);
    }
    _emit();
  }

  Future<void> loadGitHosts() async {
    if (_disposed) return;
    final epoch = _epoch, started = ++_clock;
    hostsState = RepoBindingLoad.loading;
    hostsError = null;
    _emit();
    try {
      final list = await _api.gitHosts(token);
      if (!_live(epoch)) return;
      if (!_fresh(started)) return;
      gitHosts = List.unmodifiable(list);
      hostsState = RepoBindingLoad.ready;
    } catch (failure) {
      if (!_live(epoch)) return;
      if (!_fresh(started)) return;
      if (_expired(failure)) return;
      hostsState = RepoBindingLoad.failed;
      hostsError = 'Git hosts could not be loaded.';
    }
    _emit();
  }

  Future<bool> saveBinding(RepoBindingDraft draft) async {
    if (_disposed || writing || unconfirmed) return false;
    if (!canManage()) {
      writeError =
          'You need the owner or admin role in this Workspace to change the repository binding. Nothing was changed.';
      _emit();
      return false;
    }
    final err = draft.validate();
    if (err != null) {
      writeError = err;
      _emit();
      return false;
    }
    final epoch = _epoch;
    _writeStarted('binding');
    try {
      final saved = await _api.putWorkspaceRepoBinding(
        token,
        workspaceId,
        draft.toPutMessage(),
      );
      if (!_live(epoch)) return false;
      binding = saved;
      overlaps = const [];
      loadState = RepoBindingLoad.ready;
      _writeFinished(true);
      unawaited(load());
      onBindingChanged?.call();
      return true;
    } catch (failure) {
      if (!_live(epoch)) return false;
      if (_expired(failure)) return false;
      _writeFinished(
        false,
        error: repoBindingFailureMessage(failure, RepoBindingOperation.save),
      );
      return false;
    }
  }

  Future<bool> setCredential(String pat) async {
    if (_disposed || writing || unconfirmed || !canManage()) return false;
    if (pat.trim().isEmpty) {
      writeError = 'Enter a personal access token.';
      _emit();
      return false;
    }
    final epoch = _epoch;
    _writeStarted('credential');
    try {
      final saved = await _api.setRepoBindingCredential(
        token,
        workspaceId,
        pat,
      );
      if (!_live(epoch)) return false;
      binding = saved;
      _writeFinished(true);
      return true;
    } catch (failure) {
      if (!_live(epoch)) return false;
      if (_expired(failure)) return false;
      _writeFinished(
        false,
        error: repoBindingFailureMessage(
          failure,
          RepoBindingOperation.credential,
        ),
      );
      return false;
    }
  }

  Future<bool> validate() async {
    if (_disposed || writing || unconfirmed || !canManage()) return false;
    final epoch = _epoch;
    _writeStarted('validate');
    try {
      final saved = await _api.validateRepoBinding(token, workspaceId);
      if (!_live(epoch)) return false;
      binding = saved;
      _writeFinished(true);
      return true;
    } catch (failure) {
      if (!_live(epoch)) return false;
      if (_expired(failure)) return false;
      _writeFinished(
        false,
        error: repoBindingFailureMessage(
          failure,
          RepoBindingOperation.validate,
        ),
      );
      return false;
    }
  }

  Future<OnboardWorkspaceRepositoryResponse?> onboard(
    RepoBindingOnboardingMode mode,
  ) async {
    if (_disposed || writing || unconfirmed || !canManage()) return null;
    final epoch = _epoch;
    _writeStarted('onboard');
    try {
      final result = await _api.onboardWorkspaceRepository(
        token,
        workspaceId,
        mode,
      );
      if (!_live(epoch)) return null;
      binding = result.hasBinding() ? result.binding : binding;
      _writeFinished(true);
      onBindingChanged?.call();
      return result;
    } catch (failure) {
      if (!_live(epoch)) return null;
      if (_expired(failure)) return null;
      _writeFinished(
        false,
        error: repoBindingFailureMessage(failure, RepoBindingOperation.onboard),
      );
      return null;
    }
  }

  Future<bool> syncRepository() async {
    if (_disposed || writing || unconfirmed || !canManage()) return false;
    final epoch = _epoch;
    _writeStarted('sync');
    try {
      final result = await _api.syncWorkspaceRepository(token, workspaceId);
      if (!_live(epoch)) return false;
      binding = result.binding;
      _writeFinished(true);
      browsePath = '';
      entries = const [];
      onBindingChanged?.call();
      return true;
    } catch (failure) {
      if (!_live(epoch)) return false;
      if (_expired(failure)) return false;
      _writeFinished(
        false,
        error: repoBindingFailureMessage(failure, RepoBindingOperation.sync),
      );
      return false;
    }
  }

  Future<bool> publish() async {
    if (_disposed || writing || unconfirmed || !canManage()) return false;
    final epoch = _epoch;
    _writeStarted('publish');
    try {
      final saved = await _api.publishWorkspaceRepository(token, workspaceId);
      if (!_live(epoch)) return false;
      binding = saved;
      _writeFinished(true);
      onBindingChanged?.call();
      return true;
    } catch (failure) {
      if (!_live(epoch)) return false;
      if (_expired(failure)) return false;
      _writeFinished(
        false,
        error: repoBindingFailureMessage(failure, RepoBindingOperation.publish),
      );
      return false;
    }
  }

  Future<int?> detach({
    RepoBindingDetachRecovery recovery =
        RepoBindingDetachRecovery.REPO_BINDING_DETACH_RECOVERY_UNSPECIFIED,
  }) async {
    if (_disposed || writing || unconfirmed || !canManage()) return null;
    final epoch = _epoch;
    _writeStarted('detach');
    try {
      final count = await _api.deleteWorkspaceRepoBinding(
        token,
        workspaceId,
        recovery,
      );
      if (!_live(epoch)) return null;
      binding = null;
      overlaps = const [];
      browsePath = '';
      entries = const [];
      selectedFilePath = null;
      fileContent = null;
      _writeFinished(true);
      onBindingChanged?.call();
      return count;
    } catch (failure) {
      if (!_live(epoch)) return null;
      if (_expired(failure)) return null;
      _writeFinished(
        false,
        error: repoBindingFailureMessage(failure, RepoBindingOperation.detach),
      );
      return null;
    }
  }

  Future<bool> configureWebhook() async {
    if (_disposed || writing || unconfirmed || !canManage()) return false;
    final epoch = _epoch;
    _writeStarted('webhook');
    try {
      final result = await _api.configureRepoWebhookSecret(token, workspaceId);
      if (!_live(epoch)) return false;
      binding = result.binding;
      webhookSecret = WebhookSecret(
        callbackUrl: result.callbackUrl,
        secret: result.webhookSecret,
      );
      _writeFinished(true);
      return true;
    } catch (failure) {
      if (!_live(epoch)) return false;
      if (_expired(failure)) return false;
      _writeFinished(
        false,
        error: repoBindingFailureMessage(failure, RepoBindingOperation.webhook),
      );
      return false;
    }
  }

  void dismissWebhookSecret() {
    webhookSecret = null;
    _emit();
  }

  Future<bool> acceptBaseline() async {
    if (_disposed || writing || unconfirmed || !canManage()) return false;
    final epoch = _epoch;
    _writeStarted('baseline');
    try {
      final saved = await _api.acceptRepositoryBaseline(token, workspaceId);
      if (!_live(epoch)) return false;
      binding = saved;
      _writeFinished(true);
      onBindingChanged?.call();
      return true;
    } catch (failure) {
      if (!_live(epoch)) return false;
      if (_expired(failure)) return false;
      _writeFinished(
        false,
        error: repoBindingFailureMessage(
          failure,
          RepoBindingOperation.baseline,
        ),
      );
      return false;
    }
  }

  Future<void> loadEntries({String? path}) async {
    if (_disposed) return;
    final epoch = _epoch, started = ++_clock;
    if (path != null) browsePath = path;
    entriesState = RepoBindingLoad.loading;
    entriesError = null;
    _emit();
    try {
      final reply = await _api.listRepositoryEntries(
        token,
        workspaceId,
        browsePath,
      );
      if (!_live(epoch)) return;
      if (!_fresh(started)) return;
      entries = List.unmodifiable(reply.entries);
      entriesState = RepoBindingLoad.ready;
    } catch (failure) {
      if (!_live(epoch)) return;
      if (!_fresh(started)) return;
      if (_expired(failure)) return;
      entriesState = RepoBindingLoad.failed;
      entriesError = repoBindingFailureMessage(
        failure,
        RepoBindingOperation.browse,
      );
    }
    _emit();
  }

  Future<void> loadFile(String path) async {
    if (_disposed) return;
    final epoch = _epoch, started = ++_clock;
    selectedFilePath = path;
    fileState = RepoBindingLoad.loading;
    fileError = null;
    fileContent = null;
    _emit();
    try {
      final reply = await _api.getRepositoryFile(token, workspaceId, path);
      if (!_live(epoch)) return;
      if (!_fresh(started)) return;
      fileContent = reply.content;
      fileState = RepoBindingLoad.ready;
    } catch (failure) {
      if (!_live(epoch)) return;
      if (!_fresh(started)) return;
      if (_expired(failure)) return;
      fileState = RepoBindingLoad.failed;
      fileError = repoBindingFailureMessage(
        failure,
        RepoBindingOperation.browse,
      );
    }
    _emit();
  }

  void suspend() {
    _epoch++;
    _apiOrNull?.close();
    _apiOrNull = null;
    webhookSecret = null;
    _emit();
  }

  Future<void> recover() async {
    if (_disposed) return;
    await load();
    if (canManage()) await loadGitHosts();
  }

  @override
  void dispose() {
    _disposed = true;
    _epoch++;
    _apiOrNull?.close();
    _apiOrNull = null;
    super.dispose();
  }
}

class WebhookSecret {
  const WebhookSecret({required this.callbackUrl, required this.secret});
  final String callbackUrl;
  final String secret;
}
