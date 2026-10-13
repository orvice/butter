import 'dart:async';
import 'dart:typed_data';

import 'package:butter_mobile/app/client_context.dart';
import 'package:butter_mobile/gen/agents/v1/workspace.pb.dart';
import 'package:flutter_test/flutter_test.dart';

import 'client_context_test.dart' show MemoryStorage;
import 'support/management_server.dart';

void main() {
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

  test(
    'profile preserves absent avatar, explicitly clears it and uploads multipart',
    () async {
      expect(await client.saveProfile('Changed'), isTrue);
      expect(client.user!.avatarUrl, 'https://cdn.example/original.png');
      expect(server.lastProfile!.hasAvatarUrl(), isFalse);
      expect(await client.saveProfile('Changed', avatarUrl: ''), isTrue);
      expect(client.user!.avatarUrl, '');
      expect(
        await client.saveProfile(
          'With image',
          avatar: Uint8List.fromList([137, 80, 78, 71, 13, 10, 26, 10, 1]),
        ),
        isTrue,
      );
      expect(client.user!.avatarUrl, 'https://cdn.example/new.png');
      expect(server.uploadType, startsWith('multipart/form-data; boundary='));
      expect(server.uploadBody, contains('name="file"; filename="avatar.png"'));
      expect(server.uploadBody, contains('Content-Type: image/png'));
      expect(
        server.requests
            .where((r) => r.path.contains('uploads'))
            .single
            .workspace,
        isNull,
      );
      expect(await client.changePassword('wrong', 'new password'), isFalse);
      expect(client.user, isNotNull);
      expect(await client.changePassword('password', 'new password'), isTrue);
    },
  );

  test(
    'denied owner write is not repeated and an expired write clears identity',
    () async {
      await client.openWorkspace('alpha');
      server.role = 'member';
      expect(
        await client.saveWorkspace(
          client.managedWorkspace!.deepCopy()..name = 'Denied',
        ),
        isFalse,
      );
      expect(client.error, contains('permission'));
      expect(client.workspaces.first.name, 'Alpha');
      expect(server.writes, 1);
      await client.openWorkspace('alpha');
      expect(client.canManageWorkspace, isFalse);
      server.role = 'admin';
      await client.openWorkspace('alpha');
      expect(client.canManageWorkspace, isFalse);
      server.expired = true;
      expect(await client.saveProfile('Expired'), isFalse);
      expect(client.user, isNull);
      expect(client.members, isEmpty);
      expect(client.managedWorkspace, isNull);
    },
  );

  test(
    'late Workspace membership cannot grant permissions in another scope',
    () async {
      final entered = Completer<void>();
      final release = Completer<void>();
      server.beforeMembers = () async {
        entered.complete();
        await release.future;
      };
      final pending = client.openWorkspace('alpha');
      await entered.future;
      await client.selectWorkspace('beta');
      release.complete();
      await pending;
      expect(client.workspace!.id, 'beta');
      expect(client.managedWorkspace, isNull);
      expect(client.members, isEmpty);
      expect(client.canManageWorkspace, isFalse);
    },
  );

  test(
    'background interruption keeps an honest unconfirmed-write notice without replay',
    () async {
      final entered = Completer<void>();
      final release = Completer<void>();
      server.beforeProfile = () async {
        entered.complete();
        await release.future;
      };
      final writing = client.saveProfile('Possibly applied');
      await entered.future;
      client.suspend();
      release.complete();
      expect(await writing, isFalse);
      await client.revalidate();
      expect(client.writeNotice, contains('not confirmed'));
      expect(server.writes, 1);
    },
  );

  test(
    'foreground revalidation revokes management data after access removal',
    () async {
      await client.openWorkspace('alpha');
      server.spaces.removeWhere((w) => w.id == 'alpha');
      final scope = client.scope;
      client.suspend();
      await client.revalidate();
      expect(client.scope, greaterThan(scope));
      expect(client.managedWorkspace, isNull);
      expect(client.members, isEmpty);
      expect(client.user, isNotNull);
    },
  );

  test(
    'global admin retains access after removing their own Workspace membership',
    () async {
      server.profile.role = 'admin';
      await client.revalidate();
      await client.selectWorkspace('alpha');
      await client.openWorkspace('alpha');
      expect(await client.removeMember('alice'), isTrue);
      expect(client.workspaces.map((w) => w.id), contains('alpha'));
      expect(client.workspace!.id, 'alpha');
      expect(client.managedWorkspace!.id, 'alpha');
      expect(client.canManageWorkspace, isTrue);
      expect(client.members, isEmpty);
    },
  );

  test(
    'owner can edit Workspace and manage members through authorized requests',
    () async {
      await client.openWorkspace('alpha');
      expect(client.canManageWorkspace, isTrue);
      final edit = client.managedWorkspace!.deepCopy()..name = 'Renamed';
      expect(await client.saveWorkspace(edit), isTrue);
      expect(client.workspaces.first.name, 'Renamed');
      expect(client.workspaces.first.description, 'Original');
      expect(await client.addMember('bob', 'member'), isTrue);
      expect(client.members.last.userId, 'bob');
      expect(await client.updateMember('bob', 'admin'), isTrue);
      expect(client.members.last.role, 'admin');
      expect(await client.removeMember('bob'), isTrue);
      expect(client.members.map((m) => m.userId), ['alice']);
      expect(
        await client.saveWorkspace(Workspace(name: 'New', slug: 'new')),
        isTrue,
      );
      expect(client.workspaces.last.id, 'created');
      expect(await client.deleteWorkspace(), isTrue);
      expect(client.workspaces.map((w) => w.id), ['beta', 'created']);
      expect(
        server.requests
            .where((r) => r.path.contains('WorkspaceService'))
            .every((r) => r.workspace == null),
        isTrue,
      );
    },
  );
}
