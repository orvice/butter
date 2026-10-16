import 'package:butter_mobile/core/butter_api.dart';
import 'package:butter_mobile/core/server_address.dart';
import 'package:butter_mobile/features/agents/agent_lifecycle_controller.dart';
import 'package:butter_mobile/gen/agents/v1/agent.pb.dart';
import 'package:flutter_test/flutter_test.dart';

/// mobile-test is owner; no Git repo binding (restore/delete sagas skip publish).
const workspace = 'daemon';

void main() {
  const server = String.fromEnvironment('BUTTER_TEST_SERVER');
  for (final prefix in ['', '/butter']) {
    group(
      'Agent lifecycle production contract $prefix',
      skip: server.isEmpty ? 'Run tool/contract.sh.' : false,
      () {
        final base = '$server$prefix';
        late ButterApi api;
        late String token;

        setUp(() async {
          api = ButterApi(ServerAddress.parse(base, allowHttp: true));
          addTearDown(api.close);
          token = (await api.login(
            'mobile-test',
            'contract-test-password',
          )).token;
        });

        Future<void> ensureActiveAgent(String id) async {
          try {
            final agent = await api.agent(token, workspace, id);
            if (agent.lifecycleStatus ==
                AgentLifecycleStatus.AGENT_LIFECYCLE_STATUS_DELETED) {
              await api.restoreAgent(token, workspace, id);
            }
            return;
          } catch (_) {}
          await api.saveAgent(
            token,
            workspace,
            Agent(
              agentId: id,
              name: id,
              type: AgentType.AGENT_TYPE_LLM,
              config: AgentConfig(model: 'fast', instruction: 'hi'),
            ),
            create: true,
          );
        }

        test('tombstone restore and list lifecycle operations', () async {
          const id = 'lifecycle-dart-agent';
          await ensureActiveAgent(id);

          final lifecycle = AgentLifecycleController(
            newApi: () => ButterApi(ServerAddress.parse(base, allowHttp: true)),
            token: token,
            workspaceId: workspace,
            canManage: () => true,
            onUnauthenticated: () {},
          );
          addTearDown(lifecycle.dispose);

          expect(
            await lifecycle.deleteAgent(id, displayName: id),
            isTrue,
            reason: lifecycle.writeError,
          );
          final tomb = await api.agent(token, workspace, id);
          expect(
            tomb.lifecycleStatus,
            AgentLifecycleStatus.AGENT_LIFECYCLE_STATUS_DELETED,
          );

          await lifecycle.loadOperations();
          expect(lifecycle.operations, isNotEmpty);

          expect(
            await lifecycle.restoreAgent(id),
            isTrue,
            reason: lifecycle.writeError,
          );
          final active = await api.agent(token, workspace, id);
          expect(
            active.lifecycleStatus,
            AgentLifecycleStatus.AGENT_LIFECYCLE_STATUS_ACTIVE,
          );
        });
      },
    );
  }
}
