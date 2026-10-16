import 'package:butter_mobile/app/client_context.dart';
import 'package:butter_mobile/core/config_resource_controller.dart';
import 'package:butter_mobile/features/notify_groups/notify_group_pages.dart';
import 'package:flutter_test/flutter_test.dart';

import 'support/config_rig.dart';
import 'support/memory_storage.dart';

void main() {
  testWidgets('lists seeded notify group on contract fixture', (tester) async {
    const server = String.fromEnvironment('BUTTER_TEST_SERVER');
    if (server.isEmpty) return;

    late ClientContext client;
    addTearDown(() => client.dispose());
    await tester.runAsync(() async {
      client = ClientContext(storage: MemoryStorage(), allowHttp: true);
      await client.setServer(server);
      await client.login('mobile-test', 'contract-test-password');
      await client.selectWorkspace('automation');
    });

    await openPage(tester, NotifyGroupListPage(client: client));
    await tester.runAsync(() => client.notifyGroups!.load());
    await waitFor(
      tester,
      () => client.notifyGroups!.listState == ConfigLoad.ready,
    );
    expect(find.text('alerts'), findsOneWidget);
  });
}
