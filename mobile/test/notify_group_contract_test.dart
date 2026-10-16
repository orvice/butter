import 'package:butter_mobile/core/butter_api.dart';
import 'package:butter_mobile/core/server_address.dart';
import 'package:butter_mobile/features/notify_groups/notify_group_controller.dart';
import 'package:butter_mobile/features/notify_groups/notify_group_draft.dart';
import 'package:butter_mobile/gen/agents/v1/agent.pb.dart';
import 'package:flutter_test/flutter_test.dart';

const workspace = 'automation';

void main() {
  const server = String.fromEnvironment('BUTTER_TEST_SERVER');
  group(
    'Notify Group production contract',
    skip: server.isEmpty ? 'Run tool/contract.sh.' : false,
    () {
      final base = server;
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

      test('lists seeded group and creates discord target', () async {
        final c = NotifyGroupController(
          newApi: () => ButterApi(ServerAddress.parse(base, allowHttp: true)),
          token: token,
          workspaceId: workspace,
          onUnauthenticated: () {},
        );
        addTearDown(c.dispose);
        await c.load();
        expect(c.items.any((g) => g.name == 'alerts'), isTrue);

        final name = 'contract-ng-${DateTime.now().millisecondsSinceEpoch}';
        final draft = NotifyGroupDraft.create()
          ..name = name
          ..targets.add(
            NotifyTargetDraft(
              type: NotifyTargetType.NOTIFY_TARGET_TYPE_DISCORD_WEBHOOK,
            )..discordWebhookUrl = 'https://discord.test/hook',
          );
        expect(await c.create(draft.build()), isTrue, reason: c.writeError);
        expect(await c.delete(name), isTrue);
      });

      test('rejects legacy telegram bot_token on create', () async {
        final c = NotifyGroupController(
          newApi: () => ButterApi(ServerAddress.parse(base, allowHttp: true)),
          token: token,
          workspaceId: workspace,
          onUnauthenticated: () {},
        );
        addTearDown(c.dispose);
        final bad = NotifyGroup(
          name: 'bad-legacy',
          enabled: true,
          targets: [
            NotifyTarget(
              type: NotifyTargetType.NOTIFY_TARGET_TYPE_TELEGRAM,
              telegram: TelegramNotifyTarget(
                // ignore: deprecated_member_use_from_same_package
                botToken: 'secret',
              ),
            ),
          ],
        );
        expect(await c.create(bad), isFalse);
        expect(c.writeError, contains('Legacy raw Telegram'));
      });
    },
  );
}
