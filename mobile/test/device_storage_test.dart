import 'package:butter_mobile/core/client_storage.dart';
import 'package:flutter_secure_storage/flutter_secure_storage.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:shared_preferences_platform_interface/shared_preferences_async_platform_interface.dart';
import 'package:shared_preferences_platform_interface/types.dart';

// OS plugin boundary, not an application store replacement.
final class PreferencesPlatform extends SharedPreferencesAsyncPlatform {
  final values = <String, String>{};
  bool failSave = false;
  Future<void> Function()? beforeRead;
  @override
  Future<String?> getString(
    String key,
    SharedPreferencesOptions options,
  ) async {
    await beforeRead?.call();
    return values[key];
  }

  @override
  Future<void> setString(
    String key,
    String value,
    SharedPreferencesOptions options,
  ) async {
    if (failSave) throw StateError('OS preferences unavailable');
    values[key] = value;
  }

  @override
  dynamic noSuchMethod(Invocation invocation) => super.noSuchMethod(invocation);
}

void main() {
  TestWidgetsFlutterBinding.ensureInitialized();
  late PreferencesPlatform preferences;
  setUp(() {
    preferences = PreferencesPlatform();
    SharedPreferencesAsyncPlatform.instance = preferences;
    FlutterSecureStorage.setMockInitialValues({});
  });
  test(
    'OS secure credential round-trips independently of ordinary preferences',
    () async {
      final storage = DeviceStorage();
      await storage.saveServer('https://butter.example');
      await storage.saveSession(
        const StoredSession(
          server: 'https://butter.example',
          token: 'sensitive-token',
        ),
      );
      final restored = DeviceStorage();
      expect(await restored.readServer(), 'https://butter.example');
      expect((await restored.readSession())!.token, 'sensitive-token');
      expect(preferences.values.values, ['https://butter.example']);
      await restored.clearSession();
      expect(await storage.readSession(), isNull);
      expect(await storage.readServer(), 'https://butter.example');
    },
  );
  test(
    'pending OAuth is secure, isolated from session and appearance preferences',
    () async {
      preferences.values['butter.appearance'] =
          '{"theme":"dark","showIdentifiers":true}';
      final storage = DeviceStorage();
      await storage.saveServer('https://butter.example');
      await storage.saveSession(
        const StoredSession(server: 'https://butter.example', token: 'tok'),
      );
      await storage.savePendingOAuth(
        PendingOAuth(
          server: 'https://butter.example',
          provider: 'github',
          state: 's',
          verifier: 'secret-verifier',
          expiresAt: DateTime.utc(2026, 10, 14),
        ),
      );
      // The verifier never reaches ordinary preferences.
      expect(preferences.values.values.join(), isNot(contains('secret')));

      await storage.clearPendingOAuth();
      expect(await storage.readPendingOAuth(), isNull);
      expect((await storage.readSession())!.token, 'tok');

      await storage.savePendingOAuth(
        PendingOAuth(
          server: 'https://butter.example',
          provider: 'github',
          state: 's',
          verifier: 'secret-verifier',
          expiresAt: DateTime.utc(2026, 10, 14),
        ),
      );
      await storage.clearSession();
      expect(await storage.readSession(), isNull);
      expect((await storage.readPendingOAuth())!.verifier, 'secret-verifier');
      expect(preferences.values['butter.appearance'], contains('dark'));
      expect(await storage.readServer(), 'https://butter.example');
    },
  );
  test(
    'corrupted pending OAuth is removed without touching the rest',
    () async {
      FlutterSecureStorage.setMockInitialValues({
        'butter.oauth.pending': '{"server":1}',
        'butter.session': '{"server":"https://butter.example","token":"tok"}',
      });
      preferences.values['butter.appearance'] = '{"theme":"light"}';
      final storage = DeviceStorage();
      expect(await storage.readPendingOAuth(), isNull);
      expect(
        await const FlutterSecureStorage().read(key: 'butter.oauth.pending'),
        isNull,
      );
      expect((await storage.readSession())!.token, 'tok');
      expect(preferences.values['butter.appearance'], '{"theme":"light"}');
    },
  );
  test('corrupted secure records fail closed and are removed', () async {
    FlutterSecureStorage.setMockInitialValues({
      'butter.session': '{"token":123}',
    });
    final storage = DeviceStorage();
    expect(await storage.readSession(), isNull);
    expect(
      await const FlutterSecureStorage().read(key: 'butter.session'),
      isNull,
    );
  });
}
