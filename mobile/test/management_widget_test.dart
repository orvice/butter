import 'dart:io';

import 'package:butter_mobile/app/butter_app.dart';
import 'package:butter_mobile/app/client_context.dart';
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';

import 'support/memory_storage.dart';
import 'support/management_server.dart';

Future<void> tapRequest(
  WidgetTester tester,
  ClientContext client,
  Finder finder,
) async {
  await tester.runAsync(() async {
    await tester.tap(finder);
    for (var i = 0; i < 200 && client.busy; i++) {
      await Future<void>.delayed(const Duration(milliseconds: 10));
    }
  });
  expect(client.busy, isFalse);
}

void main() {
  testWidgets('managing Alpha then Beta opens Beta on the first tap', (
    tester,
  ) async {
    final old = HttpOverrides.current;
    HttpOverrides.global = null;
    final server = ManagementServer();
    final client = ClientContext(storage: MemoryStorage(), allowHttp: true);
    addTearDown(() async {
      client.dispose();
      await server.close();
      HttpOverrides.global = old;
    });
    await tester.pumpWidget(ButterApp(context: client));
    await tester.runAsync(server.start);
    await tester.runAsync(() => client.setServer(server.url));
    await tester.runAsync(() => client.login('alice', 'password'));
    await tester.pumpAndSettle();
    await tapRequest(tester, client, find.byTooltip('Manage Alpha'));
    await tester.pumpAndSettle();
    expect(find.text('ID: alpha'), findsOneWidget);
    await tester.pageBack();
    await tester.pumpAndSettle();
    await tapRequest(tester, client, find.byTooltip('Manage Beta'));
    await tester.pumpAndSettle();
    expect(find.text('ID: beta'), findsOneWidget);
  });

  testWidgets(
    'owner forms retain safe input on denial and members cannot edit',
    (tester) async {
      final old = HttpOverrides.current;
      HttpOverrides.global = null;
      final server = ManagementServer();
      final client = ClientContext(storage: MemoryStorage(), allowHttp: true);
      addTearDown(() async {
        client.dispose();
        await server.close();
        HttpOverrides.global = old;
      });
      await tester.pumpWidget(ButterApp(context: client));
      await tester.runAsync(server.start);
      await tester.runAsync(() => client.setServer(server.url));
      await tester.runAsync(() => client.login('alice', 'password'));
      await tester.pumpAndSettle();
      await tapRequest(tester, client, find.byTooltip('Manage Alpha'));
      await tester.pumpAndSettle();
      expect(find.text('Edit Workspace'), findsOneWidget);
      await tester.tap(find.text('Edit Workspace'));
      await tester.pumpAndSettle();
      await tester.enterText(
        find.widgetWithText(TextFormField, 'Name'),
        'Safe draft',
      );
      server.role = 'member';
      await tapRequest(tester, client, find.text('Save Workspace'));
      await tester.pumpAndSettle();
      expect(find.text('Safe draft'), findsOneWidget);
      expect(find.textContaining('permission'), findsOneWidget);
      await tester.runAsync(client.logout);
      await tester.pumpAndSettle();
      expect(find.text('Safe draft'), findsNothing);
      expect(
        find.text('Context changed. Return to the home screen.'),
        findsWidgets,
      );
    },
  );

  testWidgets(
    'Workspace creation, member form and delete confirmation complete their UI loops',
    (tester) async {
      final old = HttpOverrides.current;
      HttpOverrides.global = null;
      final server = ManagementServer();
      final client = ClientContext(storage: MemoryStorage(), allowHttp: true);
      addTearDown(() async {
        client.dispose();
        await server.close();
        HttpOverrides.global = old;
      });
      await tester.pumpWidget(ButterApp(context: client));
      await tester.runAsync(server.start);
      await tester.runAsync(() => client.setServer(server.url));
      await tester.runAsync(() => client.login('alice', 'password'));
      await tester.pumpAndSettle();
      await tester.tap(find.text('Create Workspace'));
      await tester.pumpAndSettle();
      await tester.tap(find.text('Save Workspace'));
      await tester.pump();
      expect(find.text('Required.'), findsNWidgets(2));
      await tester.enterText(
        find.widgetWithText(TextFormField, 'Name'),
        'Created here',
      );
      await tester.enterText(
        find.widgetWithText(TextFormField, 'Slug'),
        'created-here',
      );
      await tester.pageBack();
      await tester.pumpAndSettle();
      expect(find.text('Discard unsaved changes?'), findsOneWidget);
      await tester.tap(find.text('Cancel'));
      await tester.pumpAndSettle();
      expect(find.text('created-here'), findsOneWidget);
      expect(server.writes, 0);
      await tapRequest(tester, client, find.text('Save Workspace'));
      await tester.pumpAndSettle();
      expect(find.text('Workspace saved.'), findsNothing);
      expect(find.text('Created here'), findsOneWidget);
      await tapRequest(tester, client, find.byTooltip('Manage Created here'));
      await tester.pumpAndSettle();
      await tester.tap(find.text('Add member'));
      await tester.pumpAndSettle();
      await tester.enterText(
        find.widgetWithText(TextFormField, 'Existing user ID'),
        'bob',
      );
      await tapRequest(tester, client, find.text('Save member'));
      await tester.pumpAndSettle();
      expect(find.text('Add bob?'), findsOneWidget);
      await tapRequest(tester, client, find.text('Confirm'));
      await tester.pumpAndSettle();
      expect(find.text('bob'), findsOneWidget);
      await tester.ensureVisible(find.text('Remove bob'));
      await tapRequest(tester, client, find.text('Remove bob'));
      await tester.pumpAndSettle();
      await tapRequest(tester, client, find.text('Remove member'));
      await tester.pumpAndSettle();
      expect(find.text('bob'), findsNothing);
      await tester.ensureVisible(find.text('Delete Workspace'));
      await tapRequest(tester, client, find.text('Delete Workspace'));
      await tester.pumpAndSettle();
      await tester.tap(find.text('Cancel'));
      await tester.pumpAndSettle();
      expect(client.workspaces.any((w) => w.id == 'created'), isTrue);
      await tapRequest(tester, client, find.text('Delete Workspace'));
      await tester.pumpAndSettle();
      await tapRequest(tester, client, find.text('Delete Workspace').last);
      await tester.pumpAndSettle();
      expect(find.text('Created here'), findsNothing);
      expect(tester.takeException(), isNull);
    },
  );

  testWidgets(
    'profile and settings remain accessible at large system text scale',
    (tester) async {
      final old = HttpOverrides.current;
      HttpOverrides.global = null;
      final server = ManagementServer();
      final client = ClientContext(storage: MemoryStorage(), allowHttp: true);
      tester.platformDispatcher.textScaleFactorTestValue = 2;
      tester.view.physicalSize = const Size(390, 844);
      tester.view.devicePixelRatio = 1;
      final semantics = tester.ensureSemantics();
      addTearDown(() async {
        tester.platformDispatcher.clearTextScaleFactorTestValue();
        tester.view.resetPhysicalSize();
        tester.view.resetDevicePixelRatio();
        client.dispose();
        await server.close();
        HttpOverrides.global = old;
      });
      await tester.pumpWidget(ButterApp(context: client));
      await tester.runAsync(server.start);
      await tester.runAsync(() => client.setServer(server.url));
      await tester.runAsync(() => client.login('alice', 'password'));
      await tester.pumpAndSettle();
      expect(
        tester.getSemantics(find.byTooltip('My profile')).label,
        contains('My profile'),
      );
      await tester.tap(find.byTooltip('My profile'));
      await tester.pumpAndSettle();
      expect(find.text('Username: alice'), findsOneWidget);
      expect(find.text('Create user'), findsNothing);
      await tester.scrollUntilVisible(
        find.widgetWithText(TextFormField, 'Display name'),
        200,
        scrollable: find.byType(Scrollable).first,
      );
      await tester.enterText(
        find.widgetWithText(TextFormField, 'Display name'),
        'New name',
      );
      await tester.ensureVisible(find.text('Save profile'));
      await tapRequest(tester, client, find.text('Save profile'));
      await tester.pumpAndSettle();
      expect(client.user!.displayName, 'New name');
      await tester.scrollUntilVisible(
        find.text('Profile saved.'),
        -200,
        scrollable: find.byType(Scrollable).first,
      );
      expect(find.text('Profile saved.'), findsOneWidget);
      await tester.scrollUntilVisible(
        find.widgetWithText(TextFormField, 'Current password'),
        300,
        scrollable: find.byType(Scrollable).first,
      );
      await tester.enterText(
        find.widgetWithText(TextFormField, 'Current password'),
        'wrong',
      );
      await tester.enterText(
        find.widgetWithText(TextFormField, 'New password'),
        'secret',
      );
      await tester.enterText(
        find.widgetWithText(TextFormField, 'Confirm new password'),
        'secret',
      );
      await tester.ensureVisible(find.text('Change password'));
      await tapRequest(tester, client, find.text('Change password'));
      await tester.pumpAndSettle();
      for (final label in [
        'Current password',
        'New password',
        'Confirm new password',
      ]) {
        expect(
          tester
              .widget<TextFormField>(find.widgetWithText(TextFormField, label))
              .controller!
              .text,
          '',
        );
      }
      expect(tester.takeException(), isNull);
      semantics.dispose();
    },
  );
}
