import 'dart:async';

import 'package:connectrpc/connect.dart';
import 'package:flutter/foundation.dart';
import 'package:uuid/uuid.dart';

import '../core/butter_api.dart';
import '../core/agent_errors.dart';
import '../core/client_storage.dart';
import '../core/model_provider_errors.dart';
import '../core/native_oauth.dart';
import '../core/server_address.dart';
import '../gen/agents/v1/agent.pb.dart';
import '../gen/agents/v1/agent_service.pb.dart';
import '../gen/agents/v1/auth.pb.dart';
import '../gen/agents/v1/workspace.pb.dart';
import '../core/agui_api.dart';
import '../features/butterboxes/box_controller.dart';
import '../features/chat/chat_controller.dart';

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
    this.chatPageSize = 50,
    this.chatIdleTimeout = const Duration(seconds: 60),
    String Function()? newChatId,
  }) : _newChatId = newChatId ?? (() => const Uuid().v4()),
       _storage = storage,
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

  /// Threads read per request by the chat list.
  final int chatPageSize;

  /// How long a live reply may be silent before the connection counts as lost.
  final Duration chatIdleTimeout;
  final String Function() _newChatId;
  ServerAddress? address;
  User? user;
  Workspace? workspace;
  List<Workspace> workspaces = const [];
  List<Agent> agents = const [];
  Agent? managedAgent;
  AgentRuntimeStatus? agentStatus;
  String? _agentId;
  List<Agent> agentChoices = const [];
  List<String> agentModelChoices = const [];
  List<String> agentSkillChoices = const [];
  bool agentContentGitOwned = false;
  bool agentEditorReady = false;
  bool agentRepositoryBound = false;
  bool _agentEditorShown = false;

  void _clearAgentEditor() {
    agentChoices = const [];
    agentModelChoices = const [];
    agentEditorReady = false;
    agentRepositoryBound = false;
    agentContentGitOwned = false;
    agentSkillChoices = const [];
    _agentEditorShown = false;
  }

  void _clearAgentDetail() {
    managedAgent = null;
    agentStatus = null;
    _agentId = null;
  }

  Workspace? managedWorkspace;
  List<WorkspaceMember> members = const [];
  int scope = 0;

  /// Model Providers of the selected Workspace. They carry API keys as the
  /// server returns them, so they are in memory only, never displayed, and
  /// dropped with every identity/Workspace change and when the app pauses.
  List<ModelProvider> modelProviders = const [];
  bool modelProvidersLoaded = false;
  ModelProvider? modelProvider;
  bool _providersShown = false;
  String? _providerName;
  ChatController? _chat;

  /// Chat for the current server, user and Workspace; null until signed in
  /// with a Workspace selected. One controller serves one such scope: it is
  /// disposed, with its streams and everything it holds, whenever the scope
  /// changes.
  ChatController? get chat {
    final api = address;
    final token = _token;
    final id = workspace?.id;
    final me = user?.id;
    if (_disposed || api == null || token == null || id == null || me == null) {
      return null;
    }
    return _chat ??= ChatController(
      api: ButterApi(api),
      agui: AguiApi(api, idleTimeout: chatIdleTimeout),
      newId: _newChatId,
      token: token,
      workspaceId: id,
      userId: me,
      pageSize: chatPageSize,
      onUnauthenticated: () {
        final self = _chat;
        // Only the scope that is still current may sign the person out.
        if (self != null && !self.disposed && _token == token) {
          unawaited(_expire());
        }
      },
    );
  }

  ButterBoxController? _boxes;

  /// ButterBoxes for the current server, user and Workspace; null until
  /// signed in with a Workspace selected. Like [chat], one controller serves
  /// one scope and is disposed with it. The server holds every Box credential
  /// and does every Box call; the phone never contacts a Box.
  ButterBoxController? get boxes {
    final api = address;
    final token = _token;
    final id = workspace?.id;
    if (_disposed || api == null || token == null || id == null) return null;
    return _boxes ??= ButterBoxController(
      newApi: () => ButterApi(api),
      token: token,
      workspaceId: id,
      onUnauthenticated: () {
        final self = _boxes;
        if (self != null && !self.disposed && _token == token) {
          unawaited(_expire());
        }
      },
    );
  }

  void _disposeChat() {
    final old = _chat;
    _chat = null;
    old?.dispose();
    final oldBoxes = _boxes;
    _boxes = null;
    oldBoxes?.dispose();
  }

  bool get canManageWorkspace =>
      managedWorkspace != null &&
      (user?.role == 'admin' ||
          members.any((m) => m.userId == user?.id && m.role == 'owner'));

  void _clearManagement() {
    _clearAgentDetail();
    _clearAgentEditor();
    managedWorkspace = null;
    members = const [];
  }

  void _clearModelProviders() {
    modelProviders = const [];
    modelProvidersLoaded = false;
    modelProvider = null;
    _providersShown = false;
    _providerName = null;
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
    _disposeChat();
    writeNotice = null;
    _writingGeneration = null;
    _clearManagement();
    _clearModelProviders();
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

  /// The server refused the session: sign the person out, clear the stored
  /// credential and say why.
  Future<void> _expire() async {
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
  }

  Future<void> _run(
    int generation,
    Future<void> Function() action, {
    bool clearOnFailure = false,
    String? Function(Object failure)? describe,
  }) async {
    try {
      await action();
    } catch (failure) {
      if (!_current(generation)) return;
      if (failure is ConnectException &&
          failure.code == Code.unauthenticated &&
          _token != null) {
        await _expire();
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
      error = describe?.call(failure) ?? errorMessage(failure);
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
        _disposeChat();
        workspace = null;
        agents = const [];
        _nextPage = '';
        _clearModelProviders();
        _clearAgentDetail();
        _clearAgentEditor();
      } else {
        workspace = matching.first;
      }
    }
  }

  void suspend() {
    _chat?.suspend();
    _boxes?.suspend();
    if (_writingGeneration != null) writeNotice = _unconfirmedWrite;
    _generation++;
    _api?.close();
    busy = false;
    members = const [];
    managedAgent = null;
    agentStatus = null;
    agentChoices = const [];
    agentModelChoices = const [];
    agentEditorReady = false;
    agentSkillChoices = const [];
    // Keys must not wait in memory while the app is in the background; the
    // read-only foreground recovery reloads what was open.
    modelProviders = const [];
    modelProvidersLoaded = false;
    modelProvider = null;
    _emit();
  }

  /// Foreground restoration is read-only; never repeats Login or a write.
  Future<void> revalidate() async {
    if (_token == null) return;
    final generation = _begin();
    // Provider wording applies only while the provider reads run; identity,
    // Workspace management and Agent failures keep the generic handling.
    var readingProviders = false;
    await _run(
      generation,
      () async {
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
          if (!_current(generation)) return;
          if (_agentEditorShown) {
            await _loadAgentEditor(generation);
            if (!_current(generation)) return;
          }
          final openAgent = _agentId;
          if (openAgent != null) {
            await _loadAgent(generation, openAgent);
            if (!_current(generation)) return;
          }
          if (_providersShown) {
            readingProviders = true;
            await _loadModelProviders(generation);
            final open = _providerName;
            if (open != null && _current(generation)) {
              await _loadModelProvider(generation, open);
            }
          }
        }
      },
      describe: (failure) => readingProviders ? _describeLoad(failure) : null,
    );
    // Read-only too: only what a lost connection left uncertain is read again.
    unawaited(_chat?.recover());
  }

  Future<void> refreshWorkspaces() async {
    if (_token == null) return;
    scope++;
    _disposeChat();
    _clearManagement();
    _clearModelProviders();
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
    _disposeChat();
    _clearManagement();
    _clearModelProviders();
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

  Future<void> prepareAgentEditor() async {
    if (_token == null || workspace == null) return;
    _agentEditorShown = true;
    agentEditorReady = false;
    final generation = _begin();
    await _run(generation, () => _loadAgentEditor(generation));
  }

  Future<void> _loadAgentEditor(int generation) async {
    final api = _api!;
    final token = _token!;
    final ws = workspace!.id;
    final choices = <Agent>[];
    final seen = <String>{};
    var page = '';
    do {
      final reply = await api.agents(token, ws, pageToken: page);
      if (!_current(generation)) return;
      choices.addAll(reply.agents);
      page = reply.nextPageToken;
      if (page.isNotEmpty && !seen.add(page)) {
        throw const FormatException(
          'Agent pagination did not advance. Refresh before editing.',
        );
      }
    } while (page.isNotEmpty);
    final providers = await api.modelProviders(token, ws);
    if (!_current(generation)) return;
    final binding = await api.repositoryBinding(token, ws);
    if (!_current(generation)) return;
    final skills = await api.skillNames(token, ws);
    if (!_current(generation)) return;
    agentChoices = List.unmodifiable(choices);
    agentModelChoices = List.unmodifiable({
      for (final p in providers)
        for (final m in p.models)
          if (m.alias.isNotEmpty) m.alias else if (m.name.isNotEmpty) m.name,
    });
    agentRepositoryBound = binding != null;
    agentContentGitOwned = binding?.activeCommitSha.isNotEmpty ?? false;
    agentSkillChoices = List.unmodifiable(skills);
    agentEditorReady = true;
  }

  Future<void> loadAgent(String id) async {
    if (_token == null || workspace == null) return;
    _clearAgentDetail();
    _agentId = id;
    final generation = _begin();
    await _run(generation, () async {
      await _loadAgent(generation, id);
      // An explicit authoritative detail reload is the user's reconciliation
      // step. Foreground recovery does not silently dismiss the notice.
      if (_current(generation)) writeNotice = null;
    }, describe: describeAgentFailure);
  }

  Future<void> _loadAgent(int generation, String id) async {
    final api = _api!;
    final ws = workspace!.id;
    final token = _token!;
    final result = await api.agent(token, ws, id);
    if (!_current(generation)) return;
    managedAgent = result;
    final status = await api.agentStatus(token, ws, id);
    if (_current(generation)) agentStatus = status;
  }

  Future<bool> saveAgent(Agent draft, {required bool create, Agent? previous}) {
    final ws = workspace?.id;
    if (ws == null) return Future.value(false);
    return _write((generation) async {
      final api = _api!;
      final token = _token!;
      // Re-read binding for each explicit save; a failed read must never route
      // a Git-owned prompt through UpdateAgent, where it would be ignored.
      final binding = await api.repositoryBinding(token, ws);
      if (!_current(generation)) return;
      final gitOwned = binding?.activeCommitSha.isNotEmpty ?? false;
      var saved = await api.saveAgent(
        token,
        ws,
        draft,
        create: create,
        contentGitOwned: gitOwned,
        baseCommitSha: binding?.activeCommitSha ?? '',
        previous: previous,
      );

      if (!_current(generation)) return;
      if (gitOwned && !create) {
        // A composite save answers with the stored DB record, whose
        // description and instructions are not the Git content it just
        // saved. Only an authoritative read carries the Active Revision.
        // This is a read: the write is never repeated.
        try {
          saved = await api.agent(token, ws, saved.agentId);
        } catch (failure) {
          if (!_current(generation)) return;
          if (failure is ConnectException &&
              failure.code == Code.unauthenticated) {
            rethrow;
          }
          // The save went through; show nothing rather than the DB record.
          managedAgent = null;
          agentStatus = null;
          _agentId = saved.agentId;
          throw const FormatException(
            'The change was saved, but the Agent could not be reloaded. Reload it before editing again.',
          );
        }
        if (!_current(generation)) return;
      }
      managedAgent = saved;
      _agentId = saved.agentId;
      agents = List.unmodifiable([
        for (final a in agents) a.agentId == saved.agentId ? saved : a,
        if (!agents.any((a) => a.agentId == saved.agentId)) saved,
      ]);
    }, describe: describeAgentFailure);
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
  Future<bool> _write(
    Future<void> Function(int) action, {
    String? Function(Object failure)? describe,
  }) async {
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
    }, describe: describe);
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
        _clearModelProviders();
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
          _clearModelProviders();
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

  /// Reads the selected Workspace's Model Providers. A response that arrives
  /// after the scope changed is dropped, never shown in the new one.
  Future<void> loadModelProviders() async {
    if (_token == null || workspace == null) return;
    _providersShown = true;
    _providerName = null;
    modelProvider = null;
    final generation = _begin();
    await _run(
      generation,
      () => _loadModelProviders(generation),
      describe: _describeLoad,
    );
  }

  Future<void> _loadModelProviders(int generation) async {
    final id = workspace!.id;
    final list = await _api!.modelProviders(_token!, id);
    if (!_current(generation) || workspace?.id != id) return;
    modelProviders = List.unmodifiable(list);
    modelProvidersLoaded = true;
  }

  /// Reads one provider fresh from the server; its copy is what an edit is
  /// built from, so an edit never starts from a stale list row.
  Future<void> loadModelProvider(String name) async {
    if (_token == null || workspace == null) return;
    _providersShown = true;
    _providerName = name;
    modelProvider = null;
    final generation = _begin();
    await _run(
      generation,
      () => _loadModelProvider(generation, name),
      describe: _describeLoad,
    );
  }

  Future<void> _loadModelProvider(int generation, String name) async {
    final id = workspace!.id;
    final result = await _api!.modelProvider(_token!, id, name);
    if (!_current(generation) || workspace?.id != id) return;
    modelProvider = result;
    _providerName = name;
  }

  String? _describeLoad(Object failure) =>
      describeModelProviderFailure(failure, ModelProviderOperation.load);

  /// One explicit create or whole-object update; never repeated on its own.
  /// [draft] is the complete message to store, built from the server's own copy.
  Future<bool> saveModelProvider(ModelProvider draft, {required bool create}) {
    final id = workspace?.id;
    if (id == null) return Future.value(false);
    return _write(
      (generation) async {
        final result = await _api!.saveModelProvider(
          _token!,
          id,
          draft,
          create: create,
        );
        if (!_current(generation) || workspace?.id != id) return;
        modelProviders = List.unmodifiable([
          for (final p in modelProviders) p.name == result.name ? result : p,
          if (!modelProviders.any((p) => p.name == result.name)) result,
        ]);
        modelProvider = result;
        _providerName = result.name;
        modelProvidersLoaded = true;
      },
      describe: (failure) => describeModelProviderFailure(
        failure,
        create ? ModelProviderOperation.create : ModelProviderOperation.update,
      ),
    );
  }

  Future<bool> deleteModelProvider(String name) {
    final id = workspace?.id;
    if (id == null) return Future.value(false);
    return _write(
      (generation) async {
        await _api!.deleteModelProvider(_token!, id, name);
        if (!_current(generation) || workspace?.id != id) return;
        modelProviders = List.unmodifiable(
          modelProviders.where((p) => p.name != name),
        );
        if (modelProvider?.name == name) modelProvider = null;
        if (_providerName == name) _providerName = null;
      },
      describe: (failure) =>
          describeModelProviderFailure(failure, ModelProviderOperation.delete),
    );
  }

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
    _disposeChat();
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
