import 'package:butter_mobile/app/client_context.dart';
import 'package:butter_mobile/gen/agents/v1/auth.pb.dart';
import 'package:flutter_test/flutter_test.dart';

import 'management_permission_matrix.dart';
import 'support/memory_storage.dart';
import 'support/management_server.dart';

void main() {
  group('management permission matrix (ClientContext)', () {
    late ManagementServer server;
    late ClientContext client;

    setUp(() async {
      server = ManagementServer();
      await server.start();
      client = ClientContext(storage: MemoryStorage(), allowHttp: true);
      await client.setServer(server.url);
      await client.login('alice', 'password');
    });

    tearDown(() async {
      client.dispose();
      await server.close();
    });

    for (final entry in managementPermissionMatrix.entries) {
      final (wsRole, globalRole) = entry.key;
      final row = entry.value;
      test(row.label, () async {
        final membership = switch (wsRole) {
          WorkspaceRole.member => 'member',
          WorkspaceRole.workspaceAdmin => 'admin',
          WorkspaceRole.owner => 'owner',
        };
        server.role = membership;
        server.profile = User(
          id: server.profile.id,
          username: server.profile.username,
          displayName: server.profile.displayName,
          avatarUrl: server.profile.avatarUrl,
          role: globalRole == GlobalRole.globalAdmin ? 'admin' : '',
        );
        if (globalRole == GlobalRole.globalAdmin) {
          await client.revalidate();
        }
        await client.openWorkspace('alpha');
        expect(client.managedWorkspace?.id, 'alpha');
        expect(client.canManageWorkspace, row.canManageWorkspace);
        expect(
          client.canManageWorkspaceResources,
          row.canManageWorkspaceResources,
        );
      });
    }
  });
}
