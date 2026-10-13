import 'dart:async';
import 'dart:convert';

import 'package:butter_mobile/app/app_preferences.dart';
import 'package:butter_mobile/app/butter_app.dart';
import 'package:butter_mobile/app/client_context.dart';
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:shared_preferences_platform_interface/shared_preferences_async_platform_interface.dart';

import 'client_context_test.dart' show MemoryStorage;
import 'device_storage_test.dart' show PreferencesPlatform;

void main() {
  testWidgets(
    'delayed restoration cannot be overwritten by an early settings form',
    (tester) async {
      final release = Completer<void>();
      final platform = PreferencesPlatform()
        ..values['butter.appearance'] =
            '{"theme":"dark","showIdentifiers":false}'
        ..beforeRead = () => release.future;
      SharedPreferencesAsyncPlatform.instance = platform;
      final client = ClientContext(storage: MemoryStorage());
      addTearDown(client.dispose);
      await tester.pumpWidget(ButterApp(context: client));
      expect(
        tester
            .widget<IconButton>(find.widgetWithIcon(IconButton, Icons.palette))
            .onPressed,
        isNull,
      );
      release.complete();
      await tester.pumpAndSettle();
      await tester.tap(find.byTooltip('Appearance and display'));
      await tester.pumpAndSettle();
      expect(find.text('dark'), findsOneWidget);
      await tester.tap(find.text('Show resource identifiers'));
      await tester.tap(find.text('Save appearance'));
      await tester.pumpAndSettle();
      expect(
        jsonDecode(platform.values['butter.appearance']!)['theme'],
        'dark',
      );
    },
  );

  testWidgets(
    'appearance failure keeps choices; explicit save applies and restores theme/display',
    (tester) async {
      final platform = PreferencesPlatform();
      SharedPreferencesAsyncPlatform.instance = platform;
      final settings = AppPreferences();
      final client = ClientContext(storage: MemoryStorage());
      addTearDown(() {
        settings.dispose();
        client.dispose();
      });
      await tester.runAsync(settings.start);
      await tester.pumpWidget(
        ButterApp(context: client, preferences: settings),
      );
      await tester.tap(find.byTooltip('Appearance and display'));
      await tester.pumpAndSettle();
      await tester.tap(find.text('system'));
      await tester.pumpAndSettle();
      await tester.tap(find.text('dark').last);
      await tester.pumpAndSettle();
      await tester.tap(find.text('Show resource identifiers'));
      platform.failSave = true;
      await tester.tap(find.text('Save appearance'));
      await tester.pumpAndSettle();
      expect(find.textContaining('could not be saved'), findsOneWidget);
      expect(settings.theme, ThemeMode.system);
      expect(find.text('dark'), findsOneWidget);
      expect(
        tester.widget<CheckboxListTile>(find.byType(CheckboxListTile)).value,
        isTrue,
      );
      platform.failSave = false;
      await tester.tap(find.text('Save appearance'));
      await tester.pumpAndSettle();
      expect(find.text('Appearance saved.'), findsOneWidget);
      expect(
        Theme.of(tester.element(find.text('Save appearance'))).brightness,
        Brightness.dark,
      );
      final restored = AppPreferences();
      await tester.runAsync(restored.start);
      expect(restored.theme, ThemeMode.dark);
      expect(restored.showIdentifiers, isTrue);
      restored.dispose();
    },
  );
}
