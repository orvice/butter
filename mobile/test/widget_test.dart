import 'dart:io';

import 'package:butter_mobile/app/butter_app.dart';
import 'package:butter_mobile/app/client_context.dart';
import 'package:butter_mobile/gen/agents/v1/agent_service.pb.dart';
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';

import 'support/memory_storage.dart';
import 'support/server.dart';

void main() {
  testWidgets(
    'permission failure is visible and successful empty refresh is not a failure',
    (tester) async {
      final oldOverrides = HttpOverrides.current;
      HttpOverrides.global = null;
      final server = TestServer();
      final context = ClientContext(storage: MemoryStorage(), allowHttp: true);
      addTearDown(() async {
        context.dispose();
        await server.close();
        HttpOverrides.global = oldOverrides;
      });
      await tester.runAsync(server.start);
      await tester.pumpWidget(ButterApp(context: context));
      await tester.runAsync(() => context.setServer(server.url));
      await tester.pumpAndSettle();
      await tester.runAsync(() => context.login('alice', 'password'));
      await tester.pumpAndSettle();
      server.denyAgents = true;
      await tester.runAsync(() => context.selectWorkspace('alpha'));
      await tester.pumpAndSettle();
      expect(
        find.text('You do not have permission for this operation.'),
        findsOneWidget,
      );
      expect(find.text('No Agents in this Workspace.'), findsNothing);
      await tester.runAsync(() async {
        server.denyAgents = false;
        server.agentReply = (_, _) async => ListAgentsResponse();
        await context.refreshAgents();
      });
      await tester.pumpAndSettle();
      expect(find.text('No Agents in this Workspace.'), findsOneWidget);
      expect(
        find.text('You do not have permission for this operation.'),
        findsNothing,
      );
    },
  );
  testWidgets(
    'connection and password forms reach authorized Agent list, including admin without global controls',
    (tester) async {
      final oldOverrides = HttpOverrides.current;
      HttpOverrides.global = null;
      final server = TestServer();
      final context = ClientContext(storage: MemoryStorage(), allowHttp: true);
      addTearDown(() async {
        context.dispose();
        await server.close();
        HttpOverrides.global = oldOverrides;
      });
      await tester.runAsync(server.start);
      await tester.pumpWidget(ButterApp(context: context));
      expect(find.text('Connect to Butter'), findsOneWidget);
      await tester.tap(find.text('Verify and save'));
      await tester.pump();
      expect(find.text('Enter a server address.'), findsOneWidget);
      await tester.runAsync(() => context.setServer(server.url));
      await tester.pumpAndSettle();
      expect(find.text('Sign in'), findsWidgets);
      await tester.tap(find.byKey(const Key('sign-in')));
      await tester.pump();
      expect(find.text('Enter your username.'), findsOneWidget);
      expect(find.text('Enter your password.'), findsOneWidget);
      await tester.runAsync(() => context.login('admin', 'password'));
      await tester.pumpAndSettle();
      expect(find.text('Choose Workspace'), findsOneWidget);
      await tester.runAsync(() => context.selectWorkspace('alpha'));
      await tester.pumpAndSettle();
      expect(find.text('alpha agent'), findsOneWidget);
      expect(find.text('Global administration'), findsNothing);
      expect(find.byKey(const Key('password')), findsNothing);
      expect(find.text(server.url), findsOneWidget);
    },
  );
}
