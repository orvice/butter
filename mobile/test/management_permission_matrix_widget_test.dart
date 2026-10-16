import 'dart:io';

import 'package:butter_mobile/app/butter_app.dart';
import 'package:butter_mobile/app/client_context.dart';
import 'package:butter_mobile/features/workspaces/workspace_pages.dart';
import 'package:butter_mobile/gen/agents/v1/auth.pb.dart';
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';

import 'management_permission_matrix.dart';
import 'management_widget_test.dart' show tapRequest;
import 'support/memory_storage.dart';
import 'support/management_server.dart';

void main() {
  group('management permission matrix (UI)', () {
    for (final entry in managementPermissionMatrix.entries) {
      final (wsRole, globalRole) = entry.key;
      final row = entry.value;
      testWidgets('workspace detail gates: ${row.label}', (tester) async {
        final old = HttpOverrides.current;
        HttpOverrides.global = null;
        final server = ManagementServer();
        final client = ClientContext(storage: MemoryStorage(), allowHttp: true);
        addTearDown(() async {
          client.dispose();
          await server.close();
          HttpOverrides.global = old;
        });
        server.role = switch (wsRole) {
          WorkspaceRole.member => 'member',
          WorkspaceRole.workspaceAdmin => 'admin',
          WorkspaceRole.owner => 'owner',
        };
        server.profile = User(
          id: 'alice',
          username: 'alice',
          displayName: 'Alice',
          avatarUrl: '',
          role: globalRole == GlobalRole.globalAdmin ? 'admin' : '',
        );
        await tester.runAsync(server.start);
        await tester.runAsync(() => client.setServer(server.url));
        await tester.runAsync(() => client.login('alice', 'password'));
        if (globalRole == GlobalRole.globalAdmin) {
          await tester.runAsync(client.revalidate);
        }
        await tester.runAsync(() => client.openWorkspace('alpha'));
        await tester.pumpWidget(
          MaterialApp(
            home: WorkspaceDetailPage(client: client, id: 'alpha'),
          ),
        );
        await tester.pumpAndSettle();
        if (row.canManageWorkspace) {
          expect(find.text('Edit Workspace'), findsOneWidget);
        } else {
          expect(find.text('Edit Workspace'), findsNothing);
          expect(find.textContaining('Read-only'), findsOneWidget);
        }
      });
    }

    testWidgets('global admin sees no out-of-scope global admin nav', (
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
      server.role = 'member';
      server.profile = User(
        id: 'alice',
        username: 'alice',
        displayName: 'Alice',
        role: 'admin',
      );
      await tester.runAsync(server.start);
      await tester.runAsync(() => client.setServer(server.url));
      await tester.runAsync(() => client.login('alice', 'password'));
      await tester.runAsync(client.revalidate);
      await tester.pumpWidget(ButterApp(context: client));
      await tester.pumpAndSettle();
      await tester.runAsync(() => client.selectWorkspace('alpha'));
      await tester.pumpAndSettle();
      for (final label in excludedGlobalAdminNavLabels) {
        expect(find.text(label), findsNothing);
      }
      await tapRequest(tester, client, find.byTooltip('My profile'));
      await tester.pumpAndSettle();
      for (final label in excludedGlobalAdminNavLabels) {
        expect(find.text(label), findsNothing);
      }
    });
  });
}
