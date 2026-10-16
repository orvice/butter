import 'package:butter_mobile/app/client_context.dart';
import 'package:butter_mobile/features/telegram/telegram_controller.dart';
import 'package:butter_mobile/features/telegram/telegram_pages.dart';
import 'package:flutter_test/flutter_test.dart';

import 'support/config_rig.dart';
import 'support/memory_storage.dart';

void main() {
  testWidgets('lists seeded channel on contract fixture', (tester) async {
    const server = String.fromEnvironment('BUTTER_TEST_SERVER');
    if (server.isEmpty) return;

    late ClientContext client;
    addTearDown(() => client.dispose());
    await tester.runAsync(() async {
      client = ClientContext(storage: MemoryStorage(), allowHttp: true);
      await client.setServer(server);
      await client.login('mobile-test', 'contract-test-password');
      await client.selectWorkspace('telegram');
    });

    await openPage(tester, TelegramChannelListPage(client: client));
    await tester.runAsync(() => client.telegram!.loadChannels());
    await waitFor(
      tester,
      () => client.telegram!.channelListState == TelegramLoad.ready,
    );
    expect(find.text('Seeded'), findsOneWidget);
  });

  testWidgets('shows administrator hint when platform webhook base is denied', (
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
      await client.selectWorkspace('telegram');
    });

    await openPage(tester, TelegramChannelListPage(client: client));
    await tester.runAsync(() async {
      await client.telegram!.loadChannels();
      await client.telegram!.loadPlatformSettings();
    });
    await waitFor(tester, () => client.telegram!.platformSettingsDenied);
    expect(find.textContaining('platform administrator'), findsOneWidget);
  });
}
