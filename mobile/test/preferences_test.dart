import 'package:butter_mobile/app/app_preferences.dart';
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';

import 'device_storage_test.dart' show PreferencesPlatform;
import 'package:shared_preferences_platform_interface/shared_preferences_async_platform_interface.dart';

void main() {
  TestWidgetsFlutterBinding.ensureInitialized();
  test(
    'theme and display settings survive restart without storing identity',
    () async {
      final platform = PreferencesPlatform();
      SharedPreferencesAsyncPlatform.instance = platform;
      final preferences = AppPreferences();
      await preferences.start();
      expect(preferences.theme, ThemeMode.system);
      await preferences.save(theme: ThemeMode.dark, showIdentifiers: true);
      final restored = AppPreferences();
      await restored.start();
      expect(restored.theme, ThemeMode.dark);
      expect(restored.showIdentifiers, isTrue);
      expect(platform.values.keys, ['butter.appearance']);
      preferences.dispose();
      restored.dispose();
    },
  );
}
