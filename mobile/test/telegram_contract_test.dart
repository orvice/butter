import 'package:butter_mobile/core/butter_api.dart';
import 'package:butter_mobile/core/server_address.dart';
import 'package:butter_mobile/features/telegram/telegram_controller.dart';
import 'package:butter_mobile/features/telegram/telegram_draft.dart';
import 'package:flutter_test/flutter_test.dart';

const workspace = 'telegram';
const seededChannelId = 'tg-seeded-channel';
const seededDestId = 'tg-seeded-dest';
const resendRecordId = 'tg-resend-record';
const contractBotToken = '222222:contract-tg-bot-2';

void main() {
  const server = String.fromEnvironment('BUTTER_TEST_SERVER');
  for (final prefix in ['', '/butter']) {
    group(
      'Telegram production contract $prefix',
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

        TelegramController tg() {
          final c = TelegramController(
            newApi: () => ButterApi(ServerAddress.parse(base, allowHttp: true)),
            token: token,
            workspaceId: workspace,
            onUnauthenticated: () {},
            agents: () => const [],
            modelAliases: () => const ['fast'],
          );
          addTearDown(c.dispose);
          return c;
        }

        test(
          'lists seeded channel, destination and processing record',
          () async {
            final c = tg();
            await c.loadChannels();
            expect(c.channelListState, TelegramLoad.ready);
            expect(c.channels.any((ch) => ch.id == seededChannelId), isTrue);

            await c.loadDestinations(channelId: seededChannelId);
            expect(c.destinationListState, TelegramLoad.ready);
            expect(c.destinations.any((d) => d.id == seededDestId), isTrue);

            await c.loadProcessingRecords();
            expect(c.processingState, TelegramLoad.ready);
            expect(c.records.any((r) => r.id == resendRecordId), isTrue);
          },
        );

        test('channel status and read-only platform settings', () async {
          final c = tg();
          await c.loadChannelStatus(seededChannelId);
          expect(c.statusState, TelegramLoad.ready);
          expect(c.channelStatus, isNotNull);

          await c.loadPlatformSettings();
          expect(c.platformSettingsDenied, isTrue);
        });

        test(
          'test send and resend on seeded fixtures',
          skip: server.isEmpty
              ? 'Run tool/contract.sh.'
              : prefix.isNotEmpty
              ? 'Shared fixture; run at root prefix.'
              : false,
          () async {
            final c = tg();
            final send = await c.sendTestMessage(
              seededDestId,
              text: 'contract ping',
            );
            expect(send.ok, isTrue, reason: c.writeError);

            await c.loadProcessingRecord(resendRecordId);
            expect(c.record?.id, resendRecordId);
            expect(
              await c.resendReply(resendRecordId),
              isTrue,
              reason: c.writeError,
            );
          },
        );

        test(
          'creates and deletes a channel with validated bot token',
          skip: server.isEmpty
              ? 'Run tool/contract.sh.'
              : prefix.isNotEmpty
              ? 'Shared fixture; run at root prefix.'
              : false,
          () async {
            final c = tg();
            final key = 'contract-${DateTime.now().millisecondsSinceEpoch}';
            final draft = TelegramChannelDraft.create()
              ..key = key
              ..name = 'Contract channel';
            expect(
              await c.createChannel(draft, contractBotToken),
              isTrue,
              reason: c.writeError,
            );
            final created = c.channels.firstWhere((ch) => ch.key == key);
            expect(await c.deleteChannel(created.id), isTrue);
          },
        );
      },
    );
  }
}
