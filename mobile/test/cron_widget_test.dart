import 'package:butter_mobile/app/client_context.dart';
import 'package:butter_mobile/core/config_resource_controller.dart';
import 'package:butter_mobile/features/cron/cron_pages.dart';
import 'package:flutter_test/flutter_test.dart';

import 'support/config_rig.dart';
import 'support/memory_storage.dart';

void main() {
  testWidgets('lists seeded cron job on contract fixture', (tester) async {
    const server = String.fromEnvironment('BUTTER_TEST_SERVER');
    if (server.isEmpty) return;

    late ClientContext client;
    addTearDown(() => client.dispose());
    await tester.runAsync(() async {
      client = ClientContext(storage: MemoryStorage(), allowHttp: true);
      await client.setServer(server);
      await client.login('mobile-test', 'contract-test-password');
      await client.selectWorkspace('cron');
    });

    await openPage(tester, CronListPage(client: client));
    await tester.runAsync(() => client.cron!.load());
    await waitFor(tester, () => client.cron!.listState == ConfigLoad.ready);
    expect(find.text('seeded-daily'), findsOneWidget);
  });
}
