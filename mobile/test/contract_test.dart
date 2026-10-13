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
