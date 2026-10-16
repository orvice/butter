import 'package:butter_mobile/core/butter_api.dart';
import 'package:butter_mobile/core/server_address.dart';
import 'package:butter_mobile/features/ops/api_token_logic.dart';
import 'package:butter_mobile/features/ops/invocation_logic.dart';
import 'package:butter_mobile/features/ops/workspace_ops_controller.dart';
import 'package:butter_mobile/gen/agents/v1/agent_service.pb.dart';
import 'package:butter_mobile/gen/agents/v1/api_token.pb.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  test('user API tokens exclude daemon kind', () {
    expect(
      isUserApiToken(APIToken(kind: APITokenKind.API_TOKEN_KIND_USER)),
      isTrue,
    );
    expect(
      isUserApiToken(APIToken(kind: APITokenKind.API_TOKEN_KIND_DAEMON)),
      isFalse,
    );
  });

  test('active invocations support cancel', () {
    expect(
      invocationSupportsCancel(
        Invocation(status: InvocationStatus.INVOCATION_STATUS_RUNNING),
      ),
      isTrue,
    );
    expect(
      invocationSupportsCancel(
        Invocation(status: InvocationStatus.INVOCATION_STATUS_SUCCEEDED),
      ),
      isFalse,
    );
  });

  test('suspend clears a minted API secret from memory', () {
    final c = WorkspaceOpsController(
      newApi: () => ButterApi(ServerAddress.parse('https://example.test')),
      token: 't',
      workspaceId: 'ws',
      isGlobalAdmin: () => false,
      onUnauthenticated: () {},
      agentCount: () => 0,
    );
    addTearDown(c.dispose);
    c.minted = MintedApiSecret(
      token: APIToken(id: '1', name: 'n'),
      secret: 'sekret',
    );
    c.suspend();
    expect(c.minted, isNull);
  });

  test('invocation filters trim and apply to list calls', () {
    final c = WorkspaceOpsController(
      newApi: () => throw UnimplementedError(),
      token: 't',
      workspaceId: 'ws',
      isGlobalAdmin: () => false,
      onUnauthenticated: () {},
      agentCount: () => 0,
    );
    addTearDown(c.dispose);
    c.setInvocationFilters(agentId: '  a1 ', sessionId: ' s2 ');
    expect(c.invocationFilterAgentId, 'a1');
    expect(c.invocationFilterSessionId, 's2');
  });
}
