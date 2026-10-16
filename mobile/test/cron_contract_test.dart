import 'package:butter_mobile/core/butter_api.dart';
import 'package:butter_mobile/core/config_resource_controller.dart';
import 'package:butter_mobile/core/server_address.dart';
import 'package:butter_mobile/features/cron/cron_controller.dart';
import 'package:butter_mobile/features/cron/cron_draft.dart';
import 'package:butter_mobile/features/sessions/session_controller.dart';
import 'package:butter_mobile/gen/agents/v1/cron.pb.dart';
import 'package:flutter_test/flutter_test.dart';

const workspace = 'cron';
const cronInstantJob = 'instant-log';
const cronSeededJob = 'seeded-daily';

void main() {
  const server = String.fromEnvironment('BUTTER_TEST_SERVER');
  for (final prefix in ['', '/butter']) {
    group(
      'Cron production contract $prefix',
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

        CronController cron({SessionController Function()? sessions}) {
          final c = CronController(
            newApi: () => ButterApi(ServerAddress.parse(base, allowHttp: true)),
            token: token,
            workspaceId: workspace,
            onUnauthenticated: () {},
            sessions: sessions ?? () => throw StateError('no sessions'),
            agents: () => const [],
          );
          addTearDown(c.dispose);
          return c;
        }

        test('lists seeded job, executions and delivery references', () async {
          final c = cron();
          await c.load();
          expect(c.listState, ConfigLoad.ready);
          expect(c.items.any((j) => j.name == 'seeded-daily'), isTrue);

          await c.loadReferences();
          expect(c.notifyGroups.any((g) => g.name == 'ops'), isTrue);
          expect(
            c.telegramDestinations.any((d) => d.id == 'cron-destination'),
            isTrue,
          );

          await c.loadExecutions('seeded-daily');
          expect(c.executions.any((e) => e.id == 'exec-waiting'), isTrue);
          final waiting = c.executions.firstWhere(
            (e) => e.id == 'exec-waiting',
          );
          if (prefix.isEmpty) {
            expect(
              waiting.status,
              CronExecutionStatus.CRON_EXECUTION_STATUS_WAITING_INPUT,
            );
          } else {
            expect(
              waiting.status,
              anyOf(
                CronExecutionStatus.CRON_EXECUTION_STATUS_WAITING_INPUT,
                CronExecutionStatus.CRON_EXECUTION_STATUS_SUCCESS,
              ),
            );
          }
        });

        test('create rejects legacy CHANNEL delivery', () async {
          final c = cron();
          final draft = CronJobDraft.create()
            ..name = 'bad-channel-$prefix'
            ..schedule = '@daily'
            ..agentId = 'cron-agent'
            ..deliveryType = CronDeliveryType.CRON_DELIVERY_TYPE_CHANNEL;
          expect(await c.create(draft.build()), isFalse);
          expect(c.writeError, contains('Legacy CHANNEL'));
        });

        test(
          'run now on instant job and reply on WAITING_INPUT execution',
          skip: server.isEmpty
              ? 'Run tool/contract.sh.'
              : prefix.isNotEmpty
              ? 'Shared fixture; run at root prefix.'
              : false,
          () async {
            final sessions = SessionController(
              newApi: () =>
                  ButterApi(ServerAddress.parse(base, allowHttp: true)),
              token: token,
              workspaceId: workspace,
              onUnauthenticated: () {},
            );
            addTearDown(sessions.dispose);
            final c = cron(sessions: () => sessions);
            expect(await c.runNow(cronInstantJob), isTrue, reason: c.runError);
            expect(c.lastRun?.jobName, cronInstantJob);

            await c.loadExecutions(cronSeededJob);
            final waiting = c.executions.firstWhere(
              (e) => e.id == 'exec-waiting',
            );
            final reply = await c.replyToExecution(
              execution: waiting,
              message: 'approved',
            );
            expect(reply.error, isNull);
            expect(reply.unconfirmed, isFalse);

            await c.loadExecutions(cronSeededJob);
            final after = c.executions.firstWhere(
              (e) => e.id == 'exec-waiting',
            );
            expect(
              after.status,
              CronExecutionStatus.CRON_EXECUTION_STATUS_SUCCESS,
            );
          },
        );

        test(
          'creates log-delivery job and deletes it',
          skip: server.isEmpty
              ? 'Run tool/contract.sh.'
              : prefix.isNotEmpty
              ? 'Shared fixture; run at root prefix.'
              : false,
          () async {
            final c = cron();
            final name = 'contract-${DateTime.now().millisecondsSinceEpoch}';
            final job =
                (CronJobDraft.create()
                      ..name = name
                      ..schedule = '@every 1h'
                      ..agentId = 'cron-agent'
                      ..deliveryType = CronDeliveryType.CRON_DELIVERY_TYPE_LOG)
                    .build();
            expect(await c.create(job), isTrue, reason: c.writeError);
            await c.load();
            expect(c.items.any((j) => j.name == name), isTrue);
            expect(await c.delete(name), isTrue);
          },
        );
      },
    );
  }
}
