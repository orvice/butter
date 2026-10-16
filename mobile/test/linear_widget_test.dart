import 'package:butter_mobile/app/client_context.dart';
import 'package:butter_mobile/features/linear/linear_controller.dart';
import 'package:butter_mobile/features/linear/linear_pages.dart';
import 'package:flutter_test/flutter_test.dart';

import 'support/config_rig.dart';
import 'support/memory_storage.dart';

void main() {
  testWidgets('lists seeded Linear App on contract fixture', (tester) async {
    const server = String.fromEnvironment('BUTTER_TEST_SERVER');
    if (server.isEmpty) return;

    late ClientContext client;
    addTearDown(() => client.dispose());
    await tester.runAsync(() async {
      client = ClientContext(storage: MemoryStorage(), allowHttp: true);
      await client.setServer(server);
      await client.login('mobile-test', 'contract-test-password');
      await client.selectWorkspace('linear');
    });

    await openPage(tester, LinearListPage(client: client));
    await tester.runAsync(() => client.linear!.loadApps());
    await waitFor(tester, () => client.linear!.listState == LinearLoad.ready);
    expect(find.textContaining('contract-linear'), findsWidgets);
  });
}
