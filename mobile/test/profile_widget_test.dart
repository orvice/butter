import 'dart:convert';
import 'dart:io';

import 'package:butter_mobile/app/butter_app.dart';
import 'package:butter_mobile/app/client_context.dart';
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:image_picker_platform_interface/image_picker_platform_interface.dart';

import 'client_context_test.dart' show MemoryStorage;
import 'management_widget_test.dart' show tapRequest;
import 'support/management_server.dart';

class PhotoPicker extends ImagePickerPlatform {
  @override
  Future<XFile?> getImageFromSource({
    required ImageSource source,
    ImagePickerOptions options = const ImagePickerOptions(),
  }) async {
    expect(source, ImageSource.gallery);
    return XFile.fromData(
      base64Decode(
        'iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVQIHWP4z8DwHwAFgAI/ScLbtAAAAABJRU5ErkJggg==',
      ),
      name: 'photo.png',
    );
  }
}

void main() {
  testWidgets(
    'photo selection is local, failed upload retains safe draft, expiry destroys it',
    (tester) async {
      final oldHttp = HttpOverrides.current;
      final oldPicker = ImagePickerPlatform.instance;
      HttpOverrides.global = null;
      ImagePickerPlatform.instance = PhotoPicker();
      final server = ManagementServer()..profile.avatarUrl = '';
      final storage = MemoryStorage();
      final client = ClientContext(storage: storage, allowHttp: true);
      addTearDown(() async {
        client.dispose();
        await server.close();
        HttpOverrides.global = oldHttp;
        ImagePickerPlatform.instance = oldPicker;
      });
      await tester.pumpWidget(ButterApp(context: client));
      await tester.runAsync(server.start);
      await tester.runAsync(() => client.setServer(server.url));
      await tester.runAsync(() => client.login('alice', 'password'));
      await tester.pumpAndSettle();
      await tester.tap(find.byTooltip('My profile'));
      await tester.pumpAndSettle();
      await tester.tap(find.text('Choose avatar'));
      await tester.pumpAndSettle();
      expect(server.writes, 0);
      expect(find.byType(Image), findsOneWidget);
      await tester.enterText(
        find.widgetWithText(TextFormField, 'Display name'),
        'Safe profile draft',
      );
      server.rejectWrites = true;
      await tester.ensureVisible(find.text('Save profile'));
      await tapRequest(tester, client, find.text('Save profile'));
      await tester.pumpAndSettle();
      expect(server.writes, 1);
      expect(find.text('Safe profile draft'), findsOneWidget);
      await tester.scrollUntilVisible(
        find.text('Choose avatar'),
        -150,
        scrollable: find.byType(Scrollable).first,
      );
      expect(find.byType(Image), findsOneWidget);
      expect(client.writeNotice, contains('not confirmed'));
      server.expired = true;
      await tester.ensureVisible(find.text('Save profile'));
      await tapRequest(tester, client, find.text('Save profile'));
      await tester.pumpAndSettle();
      expect(find.text('Safe profile draft'), findsNothing);
      expect(find.byType(Image), findsNothing);
      expect(storage.credential, isNull);
      expect(tester.takeException(), isNull);
    },
  );
}
