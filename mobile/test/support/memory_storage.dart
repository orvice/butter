import 'package:butter_mobile/core/client_storage.dart';

/// In-memory stand-in for the OS storage boundary.
class MemoryStorage implements ClientStorage {
  String? server;
  StoredSession? credential;
  PendingOAuth? pending;
  bool failSave = false;
  Future<void> Function()? beforeSave;
  Future<StoredSession?> Function()? readSessionOverride;
  bool failClear = false;
  @override
  Future<String?> readServer() async => server;
  @override
  Future<void> saveServer(String address) async {
    server = address;
  }

  @override
  Future<StoredSession?> readSession() async =>
      readSessionOverride == null ? credential : await readSessionOverride!();
  @override
  Future<void> saveSession(StoredSession session) async {
    if (failSave) throw StateError('OS vault is locked');
    await beforeSave?.call();
    credential = session;
  }

  @override
  Future<void> clearSession() async {
    if (failClear) throw StateError('OS vault is locked');
    credential = null;
  }

  @override
  Future<PendingOAuth?> readPendingOAuth() async => pending;
  @override
  Future<void> savePendingOAuth(PendingOAuth value) async {
    pending = value;
  }

  @override
  Future<void> clearPendingOAuth() async {
    pending = null;
  }
}
