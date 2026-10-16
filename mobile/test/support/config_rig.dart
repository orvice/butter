import 'dart:io';

import 'package:butter_mobile/app/client_context.dart';
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';

import 'fake_oauth_platform.dart';
import 'mcp_server.dart';
import 'memory_storage.dart';

typedef ConfigRig = ({
  ClientContext client,
  McpServerFake server,
  FakeOAuthPlatform platform,
});

/// A signed-in context on the `alpha` Workspace against [McpServerFake].
Future<ConfigRig> signedIn(
  WidgetTester tester, {
  void Function(McpServerFake)? seed,
  Duration? mcpPollInterval,
  Duration? daemonPollInterval,
  McpServerFake? using,
}) async {
  final old = HttpOverrides.current;
  HttpOverrides.global = null;
  final server = using ?? McpServerFake();
  seed?.call(server);
  final platform = FakeOAuthPlatform();
  late ClientContext client;
  addTearDown(() async {
    client.dispose();
    await server.close();
    HttpOverrides.global = old;
  });
  tester.view.physicalSize = const Size(800, 3000);
  tester.view.devicePixelRatio = 1;
  addTearDown(tester.view.reset);
  await tester.pumpWidget(const MaterialApp(home: SizedBox()));
  await tester.runAsync(() async {
    await server.start();
    client = ClientContext(
      storage: MemoryStorage(),
      allowHttp: true,
      oauth: platform,
      mcpPollInterval: mcpPollInterval ?? const Duration(seconds: 5),
      daemonPollInterval: daemonPollInterval ?? const Duration(seconds: 5),
    );
    await client.setServer(server.url);
    await client.login('alice', 'password');
    await client.selectWorkspace('alpha');
  });
  expect(client.error, isNull);
  return (client: client, server: server, platform: platform);
}

/// Lets real HTTP complete while the widget tree keeps pumping.
Future<void> waitFor(WidgetTester tester, bool Function() done) async {
  for (var i = 0; i < 300 && !done(); i++) {
    await tester.runAsync(
      () => Future<void>.delayed(const Duration(milliseconds: 10)),
    );
    await tester.pump();
  }
  expect(done(), isTrue, reason: 'the request did not finish');
  // A read that follows the write may still be on the wire, and its progress
  // bar never settles until real time passes.
  for (var i = 0; ; i++) {
    try {
      await tester.pumpAndSettle();
      return;
    } on FlutterError {
      if (i >= 100) rethrow;
      await tester.runAsync(
        () => Future<void>.delayed(const Duration(milliseconds: 20)),
      );
    }
  }
}

Future<void> tapReal(WidgetTester tester, Finder finder) async {
  await tester.ensureVisible(finder);
  await tester.runAsync(() => tester.tap(finder));
  await tester.pump();
}

/// Taps a dialog's action; what follows may be a request whose progress bar
/// never settles, so only the dialog's own closing is pumped.
Future<void> tapDialog(WidgetTester tester, Finder finder) async {
  await tester.tap(
    find.descendant(of: find.byType(AlertDialog), matching: finder),
  );
  await tester.pump();
  await tester.pump(const Duration(milliseconds: 400));
}

Future<void> tapScreen(WidgetTester tester, Finder finder) async {
  await tester.ensureVisible(finder);
  await tester.tap(finder);
  await tester.pumpAndSettle();
}

Future<void> enter(WidgetTester tester, Key key, String text) async {
  final field = find.byKey(key);
  await tester.ensureVisible(field);
  await tester.enterText(field, text);
  await tester.pump();
}

/// Builds [page] in the real zone so its own first read runs on real timers.
Future<void> openPage(WidgetTester tester, Widget page) async {
  await tester.runAsync(() => tester.pumpWidget(MaterialApp(home: page)));
  await tester.pumpAndSettle();
}

/// Taps a dialog action that goes on to make a request: the tap runs in the
/// real zone so the request's timers do too.
Future<void> tapDialogReal(WidgetTester tester, Finder finder) async {
  await tester.runAsync(
    () => tester.tap(
      find.descendant(of: find.byType(AlertDialog), matching: finder),
    ),
  );
  await tester.pump();
  await tester.pump(const Duration(milliseconds: 400));
}
