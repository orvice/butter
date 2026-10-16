import 'package:butter_mobile/core/butter_api.dart';
import 'package:butter_mobile/core/server_address.dart';
import 'package:butter_mobile/features/ops/invocation_logic.dart';
import 'package:butter_mobile/gen/agents/v1/agent_service.pb.dart';
import 'package:butter_mobile/gen/agents/v1/api_token.pb.dart';
import 'package:connectrpc/connect.dart';
import 'package:flutter_test/flutter_test.dart';

const opsWorkspace = 'ops468';

void main() {
  const server = String.fromEnvironment('BUTTER_TEST_SERVER');
  group(
    'Ops production contract',
    skip: server.isEmpty ? 'Run tool/contract.sh.' : false,
    () {
      late ButterApi api;
      late String token;

      setUpAll(() async {
        api = ButterApi(ServerAddress.parse(server, allowHttp: true));
        addTearDown(api.close);
        token = (await api.login(
          'mobile-test',
          'contract-test-password',
        )).token;
      });

      test('Dashboard GetOverview is denied for a normal member', () async {
        await expectLater(
          api.getDashboardOverview(token),
          throwsA(
            isA<ConnectException>().having(
              (e) => e.code,
              'code',
              Code.permissionDenied,
            ),
          ),
        );
      });

      test(
        'ListAgentInvocations returns seeded ops workspace records',
        () async {
          final page = await api.listAgentInvocations(
            token,
            opsWorkspace,
            pageSize: 10,
          );
          expect(page.total, greaterThanOrEqualTo(3));
          expect(
            page.invocations.map((i) => i.id),
            containsAll(['ops-inv-running', 'ops-inv-agui', 'ops-inv-done']),
          );
        },
      );

      test('ListAgentInvocations agent filter narrows results', () async {
        final page = await api.listAgentInvocations(
          token,
          opsWorkspace,
          agentId: 'ops-agent',
          pageSize: 10,
        );
        expect(page.invocations, isNotEmpty);
        expect(page.invocations.every((i) => i.agentId == 'ops-agent'), isTrue);
      });

      test('Cancel stream invocation is not stopped like AG-UI', () async {
        final reply = await api.cancelAgentInvocation(
          token,
          opsWorkspace,
          'ops-inv-running',
        );
        expect(reply.cancelled, isFalse);
        final inv = await api.getAgentInvocation(
          token,
          opsWorkspace,
          'ops-inv-running',
        );
        expect(inv.status, InvocationStatus.INVOCATION_STATUS_RUNNING);
        expect(invocationCancelExplanation(inv), isNot(contains('AG-UI Stop')));
      });

      test('Cancel detached AG-UI invocation accepts Stop path', () async {
        final reply = await api.cancelAgentInvocation(
          token,
          opsWorkspace,
          'ops-inv-agui',
        );
        expect(reply.cancelled, isTrue);
        final inv = await api.getAgentInvocation(
          token,
          opsWorkspace,
          'ops-inv-agui',
        );
        expect(inv.status, InvocationStatus.INVOCATION_STATUS_CANCELLED);
      });

      test(
        'Create and revoke personal API token with one-time secret',
        () async {
          final created = await api.createApiToken(
            token,
            opsWorkspace,
            'contract-ops-token',
          );
          expect(created.secret, isNotEmpty);
          expect(created.token.kind, APITokenKind.API_TOKEN_KIND_USER);
          final listed = await api.apiTokens(token, opsWorkspace);
          expect(
            listed.any(
              (t) =>
                  t.id == created.token.id &&
                  t.kind == APITokenKind.API_TOKEN_KIND_USER,
            ),
            isTrue,
          );
          expect(
            listed.any((t) => t.kind == APITokenKind.API_TOKEN_KIND_DAEMON),
            isFalse,
            reason: 'Daemon tokens must not appear on the personal token path',
          );
          await api.revokeApiToken(token, opsWorkspace, created.token.id);
        },
      );

      test(
        'Member workspace mcp is denied create token on cron only workspace',
        () async {
          await expectLater(
            api.createApiToken(token, 'alpha', 'should-fail'),
            throwsA(
              isA<ConnectException>().having(
                (e) => e.code,
                'code',
                Code.permissionDenied,
              ),
            ),
          );
        },
      );
    },
  );

  test('invocation cancel explanation mentions AG-UI for detached source', () {
    final inv = Invocation(source: aguiDetachedSource);
    expect(invocationCancelExplanation(inv), contains('AG-UI'));
  });
}
