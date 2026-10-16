import 'package:butter_mobile/core/butter_api.dart';
import 'package:butter_mobile/core/server_address.dart';
import 'package:butter_mobile/features/agents/agent_lifecycle_controller.dart';
import 'package:butter_mobile/gen/agents/v1/agent.pb.dart';
import 'package:flutter_test/flutter_test.dart';

import 'support/agent_server.dart';

void main() {
  test('delete tombstones and clears unconfirmed without auto-retry', () async {
    final server = AgentServer();
    server.agents['alpha']!['victim'] = Agent(
      agentId: 'victim',
      name: 'victim',
      lifecycleStatus: AgentLifecycleStatus.AGENT_LIFECYCLE_STATUS_ACTIVE,
    );
    await server.start();
    addTearDown(server.close);
    final api = ButterApi(ServerAddress.parse(server.url, allowHttp: true));
    addTearDown(api.close);
    final token = (await api.login('u', 'password')).token;
    final c = AgentLifecycleController(
      newApi: () => ButterApi(ServerAddress.parse(server.url, allowHttp: true)),
      token: token,
      workspaceId: 'alpha',
      canManage: () => true,
      onUnauthenticated: () {},
    );
    addTearDown(c.dispose);
    server.loseWriteReply = true;
    expect(await c.deleteAgent('victim', displayName: 'victim'), isFalse);
    expect(c.unconfirmed, isTrue);
    expect(
      server.agents['alpha']!['victim']!.lifecycleStatus,
      AgentLifecycleStatus.AGENT_LIFECYCLE_STATUS_ACTIVE,
    );
    server.loseWriteReply = false;
    c.unconfirmed = false;
    expect(await c.deleteAgent('victim', displayName: 'victim'), isTrue);
    expect(
      server.agents['alpha']!['victim']!.lifecycleStatus,
      AgentLifecycleStatus.AGENT_LIFECYCLE_STATUS_DELETED,
    );
  });
}
