import 'dart:convert';

import 'package:flutter_secure_storage/flutter_secure_storage.dart';
import 'package:shared_preferences/shared_preferences.dart';

final class StoredSession {
  const StoredSession({required this.server, required this.token});
  final String server;
  final String token;
}

/// OS identity storage boundary. Only the server address uses ordinary storage;
/// device appearance preferences are managed separately by AppPreferences.
abstract interface class ClientStorage {
  Future<String?> readServer();
  Future<void> saveServer(String address);
  Future<StoredSession?> readSession();
  Future<void> saveSession(StoredSession session);
  Future<void> clearSession();
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
}
