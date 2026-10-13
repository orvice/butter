import 'dart:async';

import 'package:connectrpc/connect.dart';
import 'package:flutter/foundation.dart';

import '../core/butter_api.dart';
import '../core/client_storage.dart';
import '../core/server_address.dart';
import '../gen/agents/v1/agent.pb.dart';
import '../gen/agents/v1/auth.pb.dart';
import '../gen/agents/v1/workspace.pb.dart';

enum ClientStage { server, login, workspaces, agents }

/// Owns one server/user/Workspace scope. Every replacement closes its network
/// scope and fences late responses; future observers must use the same lifetime.
final class ClientContext extends ChangeNotifier {
  ClientContext({required ClientStorage storage, this.allowHttp = false})
    : _storage = storage;
  final ClientStorage _storage;
  final bool allowHttp;
  ServerAddress? address;
  User? user;
  Workspace? workspace;
  List<Workspace> workspaces = const [];
  List<Agent> agents = const [];
  bool busy = false;
  String? error;
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
      if (credential == null) return;
      if (credential.server != address.toString()) {
        await _persist(_storage.clearSession);
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
    address = null;
    _api?.close();
    final api = ButterApi(candidate);
    _api = api;
    _emit();
    await _run(generation, () async {
      await _persist(_storage.clearSession);
      if (!_current(generation)) return;
      await api.verify();
      if (!_current(generation)) return;
      await _persist(() async {
        if (_current(generation)) {
          await _storage.saveServer(candidate.toString());
        }
      });
      if (_current(generation)) address = candidate;
    });
  }

  Future<void> login(String username, String password) async {
    if (address == null) return;
    final generation = _begin(clearIdentity: true);
    final api = _api!;
    await _run(generation, () async {
      await _persist(_storage.clearSession);
      if (!_current(generation)) return;
      final reply = await api.login(username.trim(), password);
      if (!_current(generation)) return;
      if (reply.token.isEmpty) throw StateError('Missing session token');
      _token = reply.token;
      await _persist(() async {
        if (_current(generation)) {
          await _storage.saveSession(
            StoredSession(server: address.toString(), token: reply.token),
          );
        }
      });
      if (_current(generation)) await _identity(generation);
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
  }

  void suspend() {
    _generation++;
    _api?.close();
    busy = false;
    _emit();
  }

  /// Foreground restoration is read-only; never repeats Login or a write.
  Future<void> revalidate() async {
    if (_token == null) return;
    final generation = _begin();
    await _run(generation, () async {
      await _identity(generation);
      if (!_current(generation)) return;
      if (workspace != null && !workspaces.any((w) => w.id == workspace!.id)) {
        workspace = null;
        agents = const [];
        _nextPage = '';
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
    final generation = _begin();
    workspace = null;
    agents = const [];
    _nextPage = '';
    _emit();
    await _run(generation, () => _identity(generation));
  }

  Future<void> selectWorkspace(String id) async {
    if (_token == null || !workspaces.any((w) => w.id == id)) return;
    final generation = _begin();
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

  Future<void> logout() async {
    final oldAddress = address;
    final oldToken = _token;
    final generation = _begin(clearIdentity: true);
    await _run(generation, () async {
      await _persist(_storage.clearSession);
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
    _generation++;
    _api?.close();
    super.dispose();
  }
}

/// Stable, non-sensitive UI errors; never stringify transport payloads/tokens.
String errorMessage(Object failure) {
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
