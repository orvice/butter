import 'dart:convert';

import 'package:flutter_secure_storage/flutter_secure_storage.dart';
import 'package:shared_preferences/shared_preferences.dart';

final class StoredSession {
  const StoredSession({required this.server, required this.token});
  final String server;
  final String token;
}

/// A native OAuth sign-in that began but has not been redeemed. It holds the
/// PKCE verifier, so it lives in secure storage, and it is bound to the server
/// that issued the state. It lets a sign-in finish after the App was killed
/// while the user was in the browser.
final class PendingOAuth {
  const PendingOAuth({
    required this.server,
    required this.provider,
    required this.state,
    required this.verifier,
    required this.expiresAt,
  });
  final String server;
  final String provider;
  final String state;
  final String verifier;
  final DateTime expiresAt;
}

/// OS identity storage boundary. Only the server address uses ordinary storage;
/// device appearance preferences are managed separately by AppPreferences.
abstract interface class ClientStorage {
  Future<String?> readServer();
  Future<void> saveServer(String address);
  Future<StoredSession?> readSession();
  Future<void> saveSession(StoredSession session);
  Future<void> clearSession();
  Future<PendingOAuth?> readPendingOAuth();
  Future<void> savePendingOAuth(PendingOAuth pending);
  Future<void> clearPendingOAuth();
}

final class DeviceStorage implements ClientStorage {
  DeviceStorage({
    FlutterSecureStorage? secure,
    SharedPreferencesAsync? preferences,
  }) : _secure =
           secure ??
           const FlutterSecureStorage(
             iOptions: IOSOptions(
               accessibility: KeychainAccessibility.first_unlock_this_device,
             ),
           ),
       _preferences = preferences ?? SharedPreferencesAsync();
  final FlutterSecureStorage _secure;
  final SharedPreferencesAsync _preferences;
  static const _sessionKey = 'butter.session';
  static const _serverKey = 'butter.server';
  static const _pendingKey = 'butter.oauth.pending';

  @override
  Future<String?> readServer() => _preferences.getString(_serverKey);
  @override
  Future<void> saveServer(String address) =>
      _preferences.setString(_serverKey, address);
  @override
  Future<StoredSession?> readSession() async {
    final raw = await _secure.read(key: _sessionKey);
    if (raw == null) return null;
    try {
      final json = jsonDecode(raw) as Map<String, dynamic>;
      final server = json['server'] as String;
      final token = json['token'] as String;
      if (server.isEmpty || token.isEmpty) throw const FormatException();
      return StoredSession(server: server, token: token);
    } on Object {
      await clearSession();
      return null;
    }
  }

  @override
  Future<void> saveSession(StoredSession session) => _secure.write(
    key: _sessionKey,
    value: jsonEncode({'server': session.server, 'token': session.token}),
  );
  @override
  Future<void> clearSession() => _secure.delete(key: _sessionKey);

  @override
  Future<PendingOAuth?> readPendingOAuth() async {
    final raw = await _secure.read(key: _pendingKey);
    if (raw == null) return null;
    try {
      final json = jsonDecode(raw) as Map<String, dynamic>;
      final pending = PendingOAuth(
        server: json['server'] as String,
        provider: json['provider'] as String,
        state: json['state'] as String,
        verifier: json['verifier'] as String,
        expiresAt: DateTime.parse(json['expires_at'] as String).toUtc(),
      );
      if (pending.server.isEmpty ||
          pending.provider.isEmpty ||
          pending.state.isEmpty ||
          pending.verifier.isEmpty) {
        throw const FormatException();
      }
      return pending;
    } on Object {
      await clearPendingOAuth();
      return null;
    }
  }

  @override
  Future<void> savePendingOAuth(PendingOAuth pending) => _secure.write(
    key: _pendingKey,
    value: jsonEncode({
      'server': pending.server,
      'provider': pending.provider,
      'state': pending.state,
      'verifier': pending.verifier,
      'expires_at': pending.expiresAt.toUtc().toIso8601String(),
    }),
  );

  @override
  Future<void> clearPendingOAuth() => _secure.delete(key: _pendingKey);
}
