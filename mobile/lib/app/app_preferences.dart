import 'dart:convert';

import 'package:flutter/material.dart';
import 'package:shared_preferences/shared_preferences.dart';

/// Device-only, non-sensitive appearance preferences. Never stores drafts.
final class AppPreferences extends ChangeNotifier {
  late final _storage = SharedPreferencesAsync();
  ThemeMode theme = ThemeMode.system;
  bool showIdentifiers = false;
  bool busy = false;
  String? error;
  bool _disposed = false;
  void _emit() {
    if (!_disposed) notifyListeners();
  }

  Future<void> start() async {
    busy = true;
    _emit();
    try {
      final raw = await _storage.getString('butter.appearance');
      if (raw != null) {
        final value = jsonDecode(raw) as Map<String, dynamic>;
        theme = ThemeMode.values.firstWhere(
          (mode) => mode.name == value['theme'],
          orElse: () => ThemeMode.system,
        );
        showIdentifiers = value['showIdentifiers'] == true;
      }
    } catch (_) {
      error = 'Appearance settings could not be restored.';
    } finally {
      busy = false;
      _emit();
    }
  }

  Future<bool> save({
    required ThemeMode theme,
    required bool showIdentifiers,
  }) async {
    if (busy) return false;
    busy = true;
    error = null;
    _emit();
    try {
      await _storage.setString(
        'butter.appearance',
        jsonEncode({'theme': theme.name, 'showIdentifiers': showIdentifiers}),
      );
      this.theme = theme;
      this.showIdentifiers = showIdentifiers;
      return true;
    } catch (_) {
      error = 'Appearance settings could not be saved. Retry explicitly.';
      return false;
    } finally {
      busy = false;
      _emit();
    }
  }

  @override
  void dispose() {
    _disposed = true;
    super.dispose();
  }
}
