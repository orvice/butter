import 'dart:typed_data';

import 'package:butter_mobile/gen/agents/v1/workspace.pb.dart';
import 'package:butter_mobile/core/butter_api.dart';
import 'package:butter_mobile/core/server_address.dart';
import 'package:connectrpc/connect.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  const server = String.fromEnvironment('BUTTER_TEST_SERVER');
  for (final prefix in ['', '/butter']) {
    test(
      'Dart binary unary interoperates with real Butter Go services at $prefix',
      () async {
        final api = ButterApi(
          ServerAddress.parse('$server$prefix', allowHttp: true),
        );
        addTearDown(api.close);
        await api.verify();
        await expectLater(
          api.login('mobile-test', 'wrong'),
          throwsA(
            isA<ConnectException>().having(
              (e) => e.code,
              'code',
              Code.unauthenticated,
            ),
          ),
        );
        final login = await api.login('mobile-test', 'contract-test-password');
        expect(login.token, isNotEmpty);
        final me = await api.me(login.token);
        expect(me.user.username, 'mobile-test');
        final workspaces = await api.workspaces(login.token);
        expect(
          workspaces.workspaces.map((w) => w.id),
          containsAll(['alpha', 'beta']),
        );
        final alpha = await api.agents(login.token, 'alpha');
        final beta = await api.agents(login.token, 'beta');
        expect(alpha.agents.single.agentId, 'shared');
        expect(beta.agents.single.agentId, 'shared');
        expect(alpha.agents.single.name, isNot(beta.agents.single.name));
        await expectLater(
          api.agents(login.token, 'forbidden'),
          throwsA(
            isA<ConnectException>().having(
              (e) => e.code,
              'code',
              Code.permissionDenied,
            ),
          ),
        );
        final original = (await api.workspace(login.token, 'alpha')).workspace;
        await expectLater(
          api.saveWorkspace(login.token, original.deepCopy()..name = 'Denied'),
          throwsA(
            isA<ConnectException>().having(
              (e) => e.code,
              'code',
              Code.permissionDenied,
            ),
          ),
        );
        final owned = await api.saveWorkspace(
          login.token,
          Workspace(
            name: 'Mobile owned',
            slug: prefix.isEmpty ? 'mobile-root' : 'mobile-prefix',
          ),
        );
        expect(
          (await api.members(login.token, owned.id)).members.single.role,
          'owner',
        );
        final edited = Workspace.fromBuffer([
          ...owned.writeToBuffer(),
          0x98,
          0x06,
          7,
        ])..description = 'Edited description';
        final saved = await api.saveWorkspace(login.token, edited);
        expect(saved.description, 'Edited description');
        expect(saved.unknownFields.hasField(99), isTrue);
        expect(
          (await api.addMember(
            login.token,
            owned.id,
            'other-user',
            'member',
          )).role,
          'member',
        );
        expect(
          (await api.updateMember(
            login.token,
            owned.id,
            'other-user',
            'admin',
          )).role,
          'admin',
        );
        await api.removeMember(login.token, owned.id, 'other-user');
        expect((await api.members(login.token, owned.id)).members.length, 1);
        await api.deleteWorkspace(login.token, owned.id);
        await expectLater(
          api.workspace(login.token, owned.id),
          throwsA(
            isA<ConnectException>().having(
              (e) => e.code,
              'code',
              Code.notFound,
            ),
          ),
        );
        final roleTest = await api.saveWorkspace(
          login.token,
          Workspace(
            name: 'Role test',
            slug: prefix.isEmpty ? 'role-root' : 'role-prefix',
          ),
        );
        await api.updateMember(login.token, roleTest.id, me.user.id, 'admin');
        await expectLater(
          api.saveWorkspace(login.token, roleTest..name = 'Denied admin'),
          throwsA(
            isA<ConnectException>().having(
              (e) => e.code,
              'code',
              Code.permissionDenied,
            ),
          ),
        );
        expect(
          (await api.profile(
            login.token,
            'Mobile Profile',
            avatarUrl: 'https://cdn.example/avatar.png',
          )).user.avatarUrl,
          'https://cdn.example/avatar.png',
        );
        expect(
          (await api.profile(login.token, 'Renamed Profile')).user.avatarUrl,
          'https://cdn.example/avatar.png',
        );
        expect(
          (await api.profile(
            login.token,
            'Renamed Profile',
            avatarUrl: '',
          )).user.avatarUrl,
          '',
        );
        await expectLater(
          api.changePassword(login.token, 'wrong', 'new-password'),
          throwsA(
            isA<ConnectException>().having(
              (e) => e.code,
              'code',
              Code.permissionDenied,
            ),
          ),
        );
        await api.changePassword(
          login.token,
          'contract-test-password',
          'new-password',
        );
        await api.changePassword(
          login.token,
          'new-password',
          'contract-test-password',
        );
        await expectLater(
          api.uploadAvatar(
            login.token,
            Uint8List.fromList([137, 80, 78, 71, 13, 10, 26, 10]),
          ),
          throwsA(
            isA<ConnectException>().having(
              (e) => e.code,
              'code',
              Code.unavailable,
            ),
          ),
        );
        await api.logout(login.token);
        await expectLater(
          api.me(login.token),
          throwsA(
            isA<ConnectException>().having(
              (e) => e.code,
              'code',
              Code.unauthenticated,
            ),
          ),
        );
      },
      skip: server.isEmpty
          ? 'Run tool/contract.sh against the real Go handlers.'
          : false,
    );
  }
}
