import 'package:butter_mobile/app/client_context.dart';
import 'package:butter_mobile/core/config_resource_controller.dart';
import 'package:butter_mobile/features/ops/ops_pages.dart';
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';

import 'support/config_rig.dart';
import 'support/memory_storage.dart';

void main() {
  testWidgets('ops overview and invocations smoke on contract fixture', (
    tester,
  ) async {
    const server = String.fromEnvironment('BUTTER_TEST_SERVER');
    if (server.isEmpty) return;

    late ClientContext client;
    addTearDown(() => client.dispose());
    await tester.runAsync(() async {
      client = ClientContext(storage: MemoryStorage(), allowHttp: true);
      await client.setServer(server);
      await client.login('mobile-test', 'contract-test-password');
      await client.selectWorkspace('ops468');
    });

    await openPage(tester, _SmokeOpsHost(client: client));
    await tester.runAsync(() => client.ops!.loadOverview());
    await waitFor(tester, () => client.ops!.overviewState == ConfigLoad.ready);
    expect(find.text('Overview'), findsOneWidget);

    await tester.runAsync(() => client.ops!.loadInvocations());
    await waitFor(tester, () => client.ops!.invocations.isNotEmpty);
    expect(find.textContaining('ops-inv'), findsWidgets);
  });

  testWidgets('leaving API tokens page clears minted secret', (tester) async {
    const server = String.fromEnvironment('BUTTER_TEST_SERVER');
    if (server.isEmpty) return;

    late ClientContext client;
    addTearDown(() => client.dispose());
    await tester.runAsync(() async {
      client = ClientContext(storage: MemoryStorage(), allowHttp: true);
      await client.setServer(server);
      await client.login('mobile-test', 'contract-test-password');
      await client.selectWorkspace('ops468');
    });

    await openPage(tester, _SmokeOpsHost(client: client, openTokens: true));
    await tester.runAsync(() async {
      await client.ops!.loadPersonalTokens();
      await client.ops!.createPersonalToken('widget-smoke-token');
    });
    await waitFor(tester, () => client.ops!.minted != null);
    expect(find.textContaining('Copy'), findsOneWidget);

    await tester.pageBack();
    await tester.pumpAndSettle();
    expect(client.ops!.minted, isNull);
  });
}

class _SmokeOpsHost extends StatelessWidget {
  const _SmokeOpsHost({required this.client, this.openTokens = false});
  final ClientContext client;
  final bool openTokens;

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: Center(
        child: FilledButton(
          onPressed: () async {
            if (openTokens) {
              await openApiTokens(context, client);
            } else {
              await openOverview(context, client);
            }
          },
          child: Text(openTokens ? 'tokens' : 'overview'),
        ),
      ),
    );
  }
}
