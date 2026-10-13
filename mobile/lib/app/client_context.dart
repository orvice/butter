import 'dart:async';

import 'package:connectrpc/connect.dart';
import 'package:flutter/foundation.dart';

import '../core/butter_api.dart';
import '../core/client_storage.dart';
import '../core/native_oauth.dart';
import '../core/server_address.dart';
import '../gen/agents/v1/agent.pb.dart';
import '../gen/agents/v1/auth.pb.dart';
import '../gen/agents/v1/workspace.pb.dart';

enum ClientStage { server, login, workspaces, agents }

const _timedOut = 'Sign-in timed out. Start again.';

/// A sign-in failure whose message is safe and useful to show as is.
final class SignInFailure implements Exception {
  const SignInFailure(this.message);
  final String message;
}

/// Server answers to Begin that deserve a specific explanation.
SignInFailure? _beginFailure(
  ConnectException failure,
) => switch (failure.code) {
  Code.failedPrecondition => const SignInFailure(
    'Sign-in from the app is not enabled on this server. Use your password or ask an administrator.',
  ),
  Code.invalidArgument => const SignInFailure(
    'This server does not accept sign-in from this version of the app.',
  ),
  Code.notFound => const SignInFailure(
    'That sign-in provider is no longer available.',
  ),
  _ => null,
};

/// A refused handoff is never retried: it is single-use, so start again.
SignInFailure? _redeemFailure(ConnectException failure) =>
    switch (failure.code) {
      Code.permissionDenied ||
      Code.invalidArgument ||
      Code.failedPrecondition ||
      Code.notFound => const SignInFailure(
        'Sign-in could not be completed. Start again.',
      ),
      _ => null,
    };

/// Owns one server/user/Workspace scope. Every replacement closes its network
/// scope and fences late responses; future observers must use the same lifetime.
final class ClientContext extends ChangeNotifier {
  ClientContext({
    required ClientStorage storage,
    this.allowHttp = false,
    NativeOAuthPlatform? oauth,
    String Function()? newVerifier,
    DateTime Function()? now,
    this.oauthTimeout = const Duration(minutes: 10),
  }) : _storage = storage,
       _oauth = oauth,
       _newVerifier = newVerifier ?? generatePkceVerifier,
       _now = now ?? DateTime.now;
  final ClientStorage _storage;
  final bool allowHttp;
  final NativeOAuthPlatform? _oauth;
  final String Function() _newVerifier;
  final DateTime Function() _now;

  /// How long the browser leg may take; matches the server's state lifetime.
  final Duration oauthTimeout;
  ServerAddress? address;
  User? user;
  Workspace? workspace;
  List<Workspace> workspaces = const [];
  List<Agent> agents = const [];
  Workspace? managedWorkspace;
  List<WorkspaceMember> members = const [];
  int scope = 0;
  bool get canManageWorkspace =>
      managedWorkspace != null &&
      (user?.role == 'admin' ||
          members.any((m) => m.userId == user?.id && m.role == 'owner'));

  void _clearManagement() {
    managedWorkspace = null;
    members = const [];
  }

  bool busy = false;
  String? error;
  String? writeNotice;
  int? _writingGeneration;
  static const _unconfirmedWrite =
      'Change not confirmed; it may already have been applied. Refresh before retrying.';

  /// OAuth providers the server offers for sign-in.
  List<OAuthProvider> oauthProviders = const [];

  /// Display name of the provider whose browser sign-in is in progress.
  String? oauthWaitingFor;
  PendingOAuth? _pending;
  Timer? _oauthTimer;
  StreamSubscription<Uri>? _links;
  Future<void> _oauthTail = Future.value();

  // Completes once start() has read secure storage, so a link that launched
  // the App cold is matched against the restored flow, never before it.
  Future<void> _ready = Future.value();

  /// Completes when every return link received so far has been handled.
  Future<void> get oauthIdle => _oauthTail;
  String _nextPage = '';
  bool get hasMoreAgents => _nextPage.isNotEmpty;
  ClientStage get stage => address == null
      ? ClientStage.server
      : user == null
      ? ClientStage.login
      : workspace == null
      ? ClientStage.workspaces
      : ClientStage.agents;

  ButterApi? _api;
  String? _token;
  int _generation = 0;
  bool _disposed = false;
  Future<void> _storageTail = Future.value();

  bool _current(int generation) => !_disposed && generation == _generation;
  void _emit() {
    if (!_disposed) notifyListeners();
  }

  int _begin({bool clearIdentity = false}) {
    if (_writingGeneration == _generation) writeNotice = _unconfirmedWrite;
    _generation++;
    _api?.close();
    _api = address == null ? null : ButterApi(address!);
    if (clearIdentity) _clearIdentity();
    busy = true;
    error = null;
    _emit();
    return _generation;
  }

  void _clearIdentity() {
    scope++;
    writeNotice = null;
    _writingGeneration = null;
    _clearManagement();
    _token = null;
    user = null;
    workspace = null;
    workspaces = const [];
    agents = const [];
    _nextPage = '';
  }

  // Serialize reads too: the OS adapter may remove corrupt credentials during
  // readSession. That cleanup must finish before a newer login can write.
  Future<T> _persist<T>(Future<T> Function() action) {
    final result = _storageTail.then((_) => action());
    _storageTail = result.then((_) {}, onError: (Object _) {});
    return result;
  }

  Future<void> _run(
    int generation,
    Future<void> Function() action, {
    bool clearOnFailure = false,
  }) async {
    try {
      await action();
    } catch (failure) {
      if (!_current(generation)) return;
      if (failure is ConnectException &&
          failure.code == Code.unauthenticated &&
          _token != null) {
        final expiredGeneration = _begin(clearIdentity: true);
        String notice = 'Session expired. Sign in again.';
        try {
          await _persist(_storage.clearSession);
        } catch (_) {
          notice =
              'Session expired. Secure storage could not be cleared; try signing out again.';
        }
        if (_current(expiredGeneration)) {
          busy = false;
          error = notice;
          _emit();
        }
        return;
      }
      if (clearOnFailure) {
        _clearIdentity();
        _api?.close();
        String? cleanupWarning;
        try {
          await _persist(_storage.clearSession);
        } catch (_) {
          cleanupWarning =
              'Secure storage could not be cleared. Try signing out again.';
        }
        if (!_current(generation)) return;
        error = [
          errorMessage(failure),
          cleanupWarning,
        ].whereType<String>().join(' ');
        return;
      }
      error = errorMessage(failure);
    } finally {
      if (_current(generation)) {
        busy = false;
        _emit();
      }
    }
  }

  Future<void> start() async {
    _links ??= _oauth?.links.listen(_onLink);
    final ready = Completer<void>();
    _ready = ready.future;
    try {
      await _start();
    } finally {
      ready.complete();
    }
  }

  Future<void> _start() async {
    final generation = _begin(clearIdentity: true);
    await _run(generation, () async {
      final saved = await _persist(_storage.readServer);
      if (!_current(generation) || saved == null) return;
      address = ServerAddress.parse(saved, allowHttp: allowHttp);
      _api = ButterApi(address!);
      final credential = await _persist(
        () async => _current(generation) ? await _storage.readSession() : null,
      );
      if (!_current(generation)) return;
      if (credential == null || credential.server != address.toString()) {
        if (credential != null) await _persist(_storage.clearSession);
        await _loadProviders(generation);
        await _restorePending(generation);
        return;
      }
      _token = credential.token;
      await _identity(generation);
    });
  }

  Future<void> setServer(String input) async {
    ServerAddress candidate;
    try {
      candidate = ServerAddress.parse(input, allowHttp: allowHttp);
    } on FormatException catch (failure) {
      error = failure.message.toString();
      _emit();
      return;
    }
    final generation = _begin(clearIdentity: true);
    _dropOAuth();
    address = null;
    _api?.close();
    final api = ButterApi(candidate);
    _api = api;
    _emit();
    await _run(generation, () async {
      await _persist(_storage.clearSession);
      await _persist(_storage.clearPendingOAuth);
      if (!_current(generation)) return;
      final providers = await api.verify();
      if (!_current(generation)) return;
      await _persist(() async {
        if (_current(generation)) {
          await _storage.saveServer(candidate.toString());
        }
      });
      if (_current(generation)) {
        address = candidate;
        oauthProviders = List.unmodifiable(providers);
      }
    });
  }

  Future<void> login(String username, String password) async {
    if (address == null) return;
    final generation = _begin(clearIdentity: true);
    _dropOAuth();
    final api = _api!;
    await _run(generation, () async {
      await _persist(_storage.clearSession);
      await _persist(_storage.clearPendingOAuth);
      if (!_current(generation)) return;
      final reply = await api.login(username.trim(), password);
      await _finishSignIn(generation, reply.token);
    }, clearOnFailure: true);
  }

  /// Stores the new session and reads the identity it belongs to.
  Future<void> _finishSignIn(int generation, String token) async {
    if (!_current(generation)) return;
    if (token.isEmpty) throw StateError('Missing session token');
    _token = token;
    await _persist(() async {
      if (_current(generation)) {
        await _storage.saveSession(
          StoredSession(server: address.toString(), token: token),
        );
      }
    });
    if (_current(generation)) await _identity(generation);
  }

  /// Signs in through the provider's page in the system browser. The App then
  /// waits for the server's return link ([_onLink]); the session is only
  /// issued to this App, proven by the PKCE verifier that never left it.
  Future<void> beginOAuth(String provider) async {
    final platform = _oauth;
    if (address == null || platform == null) return;
    final generation = _begin(clearIdentity: true);
    _dropOAuth();
    final api = _api!;
    await _run(generation, () async {
      await _persist(_storage.clearSession);
      await _persist(_storage.clearPendingOAuth);
      if (!_current(generation)) return;
      final verifier = _newVerifier();
      final BeginOAuthFlowResponse reply;
      try {
        reply = await api.beginOAuth(
          provider,
          nativeReturnUri,
          pkceChallenge(verifier),
        );
      } on ConnectException catch (failure) {
        throw _beginFailure(failure) ?? failure;
      }
      if (!_current(generation)) return;
      // Only a web page may be opened. The server is trusted to name the
      // provider, not to choose which app or scheme the OS launches.
      if (!_isWebUrl(reply.authorizeUrl)) {
        throw const SignInFailure(
          'The server returned an unusable sign-in address.',
        );
      }
      final pending = PendingOAuth(
        server: address.toString(),
        provider: provider,
        state: reply.state,
        verifier: verifier,
        expiresAt: _now().add(oauthTimeout),
      );
      await _persist(() async {
        if (_current(generation)) await _storage.savePendingOAuth(pending);
      });
      if (!_current(generation)) return;
      _adoptPending(pending);
      _emit();
      var opened = false;
      try {
        opened = await platform.openBrowser(Uri.parse(reply.authorizeUrl));
      } catch (_) {
        opened = false;
      }
      if (!opened) {
        _dropOAuth();
        await _persist(_storage.clearPendingOAuth);
        throw const SignInFailure(
          'Could not open the browser to sign in. Try again.',
        );
      }
    });
  }

  /// Gives up on a sign-in still waiting for the browser.
  Future<void> cancelOAuth() async {
    _dropOAuth();
    _emit();
    await _persist(_storage.clearPendingOAuth);
  }

  bool _isWebUrl(String value) {
    final url = Uri.tryParse(value);
    if (url == null || !url.hasAuthority || url.host.isEmpty) return false;
    return url.scheme == 'https' || (allowHttp && url.scheme == 'http');
  }

  void _adoptPending(PendingOAuth pending) {
    _pending = pending;
    oauthWaitingFor = oauthProviders
        .where((p) => p.name == pending.provider)
        .map((p) => p.displayName)
        .followedBy([pending.provider])
        .first;
    _oauthTimer?.cancel();
    final remaining = pending.expiresAt.difference(_now());
    _oauthTimer = Timer(remaining.isNegative ? Duration.zero : remaining, () {
      if (!identical(_pending, pending)) return;
      _dropOAuth();
      error = _timedOut;
      _emit();
      _persist(_storage.clearPendingOAuth).catchError((_) {});
    });
  }

  /// Forgets the in-memory flow. Callers also clear secure storage.
  void _dropOAuth() {
    _oauthTimer?.cancel();
    _oauthTimer = null;
    _pending = null;
    oauthWaitingFor = null;
  }

  Future<void> _loadProviders(int generation) async {
    try {
      final providers = await _api!.oauthProviders();
      if (_current(generation)) {
        oauthProviders = List.unmodifiable(providers);
      }
    } catch (_) {
      // Password sign-in does not depend on the provider list.
    }
  }

  /// After a restart, resumes a flow the App was killed during, but only one
  /// that is for this server and has not expired.
  Future<void> _restorePending(int generation) async {
    if (_oauth == null) return;
    final stored = await _persist(_storage.readPendingOAuth);
    if (!_current(generation) || stored == null) return;
    if (stored.server != address.toString() ||
        !stored.expiresAt.isAfter(_now())) {
      await _persist(_storage.clearPendingOAuth);
      return;
    }
    _adoptPending(stored);
  }

  // Links are handled one at a time, in arrival order. A user-visible failure
  // is reported by the sign-in itself (`_run`); the catch only keeps one bad
  // link from breaking the chain for the next.
  void _onLink(Uri uri) {
    _oauthTail = _oauthTail.then((_) => _handleLink(uri)).catchError((_) {});
  }

  Future<void> _handleLink(Uri uri) async {
    await _ready;
    final pending = _pending;
    if (pending == null || _disposed) return;
    final expected = Uri.parse(nativeReturnUri);
    // Anything that is not this flow's own return is not ours to act on, and
    // must not end the flow either: a stray link cannot cancel a real sign-in.
    if (uri.scheme != expected.scheme ||
        uri.host.isNotEmpty ||
        uri.path != expected.path ||
        uri.queryParameters['state'] != pending.state) {
      return;
    }
    // From here the flow is spent whatever happens: a handoff is single-use.
    _dropOAuth();
    final generation = _begin(clearIdentity: true);
    final api = _api;
    await _run(generation, () async {
      await _persist(_storage.clearPendingOAuth);
      if (!_current(generation)) return;
      if (pending.server != address.toString() || api == null) return;
      if (!pending.expiresAt.isAfter(_now())) {
        throw const SignInFailure(_timedOut);
      }
      final handoff = uri.queryParameters['handoff'] ?? '';
      if (handoff.isEmpty) {
        throw const SignInFailure('Sign-in was not completed. Start again.');
      }
      final RedeemOAuthHandoffResponse reply;
      try {
        reply = await api.redeemOAuth(
          pending.provider,
          pending.state,
          handoff,
          pending.verifier,
        );
      } on ConnectException catch (failure) {
        throw _redeemFailure(failure) ?? failure;
      }
      await _finishSignIn(generation, reply.token);
    }, clearOnFailure: true);
  }

  Future<void> _identity(int generation) async {
    final api = _api!;
    final token = _token!;
    final me = await api.me(token);
    if (!_current(generation)) return;
    if (!me.hasUser() || me.user.id.isEmpty || me.user.disabled) {
      throw ConnectException(Code.unauthenticated, 'Invalid session');
    }
    final list = await api.workspaces(token);
    if (!_current(generation)) return;
    user = me.user;
    workspaces = List.unmodifiable(list.workspaces);
    if (managedWorkspace != null &&
        !workspaces.any((w) => w.id == managedWorkspace!.id)) {
      scope++;
      _clearManagement();
    }
    if (workspace != null) {
      final matching = workspaces.where((w) => w.id == workspace!.id);
      if (matching.isEmpty) {
        scope++;
        workspace = null;
        agents = const [];
        _nextPage = '';
      } else {
        workspace = matching.first;
      }
    }
  }

  void suspend() {
    if (_writingGeneration != null) writeNotice = _unconfirmedWrite;
    _generation++;
    _api?.close();
    busy = false;
    members = const [];
    _emit();
  }

  /// Foreground restoration is read-only; never repeats Login or a write.
  Future<void> revalidate() async {
    if (_token == null) return;
    final generation = _begin();
    await _run(generation, () async {
      await _identity(generation);
      if (!_current(generation)) return;
      final managedId = managedWorkspace?.id;
      if (managedId != null) {
        _clearManagement();
        await _loadWorkspace(generation, managedId);
        if (!_current(generation)) return;
      }
      if (workspace != null) {
        agents = const [];
        _nextPage = '';
        await _loadAgents(generation, '');
      }
    });
  }

  Future<void> refreshWorkspaces() async {
    if (_token == null) return;
    scope++;
    _clearManagement();
    final generation = _begin();
    writeNotice = null;
    workspace = null;
    agents = const [];
    _nextPage = '';
    _emit();
    await _run(generation, () => _identity(generation));
  }

  Future<void> selectWorkspace(String id) async {
    if (_token == null || !workspaces.any((w) => w.id == id)) return;
    scope++;
    _clearManagement();
    final generation = _begin();
    writeNotice = null;
    workspace = workspaces.firstWhere((w) => w.id == id);
    agents = const [];
    _nextPage = '';
    _emit();
    await _run(generation, () => _loadAgents(generation, ''));
  }

  Future<void> refreshAgents() async {
    if (workspace == null) return;
    final generation = _begin();
    agents = const [];
    _nextPage = '';
    _emit();
    await _run(generation, () => _loadAgents(generation, ''));
  }

  Future<void> moreAgents() async {
    if (busy || !hasMoreAgents || workspace == null) return;
    final generation = _begin();
    await _run(generation, () => _loadAgents(generation, _nextPage));
  }

  Future<void> _loadAgents(int generation, String page) async {
    final reply = await _api!.agents(_token!, workspace!.id, pageToken: page);
    if (!_current(generation)) return;
    agents = List.unmodifiable([...agents, ...reply.agents]);
    _nextPage = reply.nextPageToken;
  }

  Future<void> openWorkspace(String id) async {
    if (_token == null || busy) return;
    if (managedWorkspace != null && managedWorkspace!.id != id) scope++;
    _clearManagement();
    final generation = _begin();
    await _run(generation, () => _loadWorkspace(generation, id));
  }

  Future<void> _loadWorkspace(int generation, String id) async {
    final api = _api!;
    final token = _token!;
    final detail = await api.workspace(token, id);
    if (!_current(generation)) return;
    final membership = await api.members(token, id);
    if (!_current(generation)) return;
    managedWorkspace = detail.workspace;
    members = List.unmodifiable(membership.members);
  }

  /// A single explicit write; success is never inferred from a stale response.
  Future<bool> _write(Future<void> Function(int) action) async {
    if (_token == null || busy) return false;
    writeNotice = null;
    final generation = _begin();
    _writingGeneration = generation;
    var succeeded = false;
    var ambiguous = false;
    await _run(generation, () async {
      try {
        await action(generation);
        succeeded = _current(generation);
      } catch (failure) {
        ambiguous =
            failure is! FormatException &&
            (failure is! ConnectException ||
                [
                  Code.unavailable,
                  Code.deadlineExceeded,
                  Code.unknown,
                  Code.internal,
                ].contains(failure.code));
        rethrow;
      }
    });
    if (_writingGeneration == generation) _writingGeneration = null;
    if (!succeeded && ambiguous && _current(generation) && user != null) {
      writeNotice = _unconfirmedWrite;
      _emit();
    }
    return succeeded && _current(generation);
  }

  Future<bool> saveWorkspace(Workspace draft) => _write((generation) async {
    final result = await _api!.saveWorkspace(_token!, draft);
    if (!_current(generation)) return;
    workspaces = List.unmodifiable([
      for (final w in workspaces) w.id == result.id ? result : w,
      if (!workspaces.any((w) => w.id == result.id)) result,
    ]);
    if (workspace?.id == result.id) workspace = result;
    if (managedWorkspace?.id == result.id) managedWorkspace = result;
  });

  Future<bool> deleteWorkspace() {
    final id = managedWorkspace?.id;
    if (id == null) return Future.value(false);
    return _write((generation) async {
      await _api!.deleteWorkspace(_token!, id);
      if (!_current(generation)) return;
      workspaces = List.unmodifiable(workspaces.where((w) => w.id != id));
      _clearManagement();
      if (workspace?.id == id) {
        scope++;
        workspace = null;
        agents = const [];
        _nextPage = '';
      }
    });
  }

  Future<bool> addMember(String userId, String role) =>
      _saveMember(userId, role, add: true);
  Future<bool> updateMember(String userId, String role) =>
      _saveMember(userId, role, add: false);
  Future<bool> _saveMember(String userId, String role, {required bool add}) {
    final id = managedWorkspace?.id;
    if (id == null) return Future.value(false);
    return _write((generation) async {
      final result = add
          ? await _api!.addMember(_token!, id, userId, role)
          : await _api!.updateMember(_token!, id, userId, role);
      if (!_current(generation)) return;
      members = List.unmodifiable([
        for (final m in members)
          if (m.userId != userId) m,
        result,
      ]);
    });
  }

  Future<bool> removeMember(String userId) {
    final id = managedWorkspace?.id;
    if (id == null) return Future.value(false);
    return _write((generation) async {
      await _api!.removeMember(_token!, id, userId);
      if (!_current(generation)) return;
      members = List.unmodifiable(members.where((m) => m.userId != userId));
      // Global admins still have access without an explicit membership.
      if (userId == user?.id && user?.role != 'admin') {
        scope++;
        _clearManagement();
        workspaces = List.unmodifiable(workspaces.where((w) => w.id != id));
        if (workspace?.id == id) {
          workspace = null;
          agents = const [];
          _nextPage = '';
        }
      }
    });
  }

  Future<bool> saveProfile(
    String name, {
    String? avatarUrl,
    Uint8List? avatar,
  }) => _write((generation) async {
    final api = _api!;
    final token = _token!;
    var url = avatarUrl;
    if (avatar != null) {
      url = await api.uploadAvatar(token, avatar);
      if (!_current(generation)) return;
    }
    final reply = await api.profile(token, name.trim(), avatarUrl: url);
    if (_current(generation)) user = reply.user;
  });

  Future<bool> changePassword(String current, String next) =>
      _write((generation) async {
        final reply = await _api!.changePassword(_token!, current, next);
        if (_current(generation)) user = reply.user;
      });

  Future<void> logout() async {
    final oldAddress = address;
    final oldToken = _token;
    final generation = _begin(clearIdentity: true);
    _dropOAuth();
    await _run(generation, () async {
      await _persist(_storage.clearSession);
      await _persist(_storage.clearPendingOAuth);
      if (!_current(generation) || oldAddress == null || oldToken == null) {
        return;
      }
      final revoke = _api!;
      // If the scope is closed before the response arrives, this honest notice
      // remains; only a confirmed response clears it. No revocation retry.
      error = 'Signed out locally. Server revocation not yet confirmed.';
      _emit();
      try {
        await revoke.logout(oldToken);
        if (_current(generation)) error = null;
      } catch (_) {
        if (_current(generation)) {
          error =
              'Signed out locally. Server revocation could not be confirmed.';
        }
      } finally {
        revoke.close();
      }
    });
  }

  @override
  void dispose() {
    _disposed = true;
    _oauthTimer?.cancel();
    _links?.cancel();
    _generation++;
    _api?.close();
    super.dispose();
  }
}

/// Stable, non-sensitive UI errors; never stringify transport payloads/tokens.
String errorMessage(Object failure) {
  if (failure is SignInFailure) return failure.message;
  if (failure is FormatException) return failure.message.toString();
  if (failure is ConnectException) {
    return switch (failure.code) {
      Code.unauthenticated => 'Sign-in failed. Check your credentials.',
      Code.permissionDenied => 'You do not have permission for this operation.',
      Code.invalidArgument => 'Check the entered values.',
      Code.notFound => 'This resource is no longer available.',
      Code.failedPrecondition => 'The server is not ready for this operation.',
      Code.deadlineExceeded || Code.unavailable =>
        'Cannot reach the server. Check your connection and retry.',
      _ => 'The request failed. Retry explicitly.',
    };
  }
  return 'The operation failed. Check your connection and secure storage, then retry.';
}
