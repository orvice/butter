import 'package:butter_mobile/core/butter_api.dart';
import 'package:butter_mobile/core/config_resource_controller.dart';
import 'package:butter_mobile/core/server_address.dart';
import 'package:butter_mobile/features/automation/automation_controller.dart';
import 'package:butter_mobile/gen/agents/v1/automation.pb.dart';
import 'package:flutter_test/flutter_test.dart';

const workspace = 'automation';
const seededManual = 'seeded-manual';
const condFail = 'cond-fail';
const webhookFail = 'webhook-fail';

void main() {
  const server = String.fromEnvironment('BUTTER_TEST_SERVER');
  group(
    'Automation production contract',
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

      AutomationController auto() {
        final c = AutomationController(
          newApi: () => ButterApi(ServerAddress.parse(base, allowHttp: true)),
          token: token,
          workspaceId: workspace,
          onUnauthenticated: () {},
          agents: () => const [],
        );
        addTearDown(c.dispose);
        return c;
      }

      test('lists seeded automations', () async {
        final c = auto();
        await c.load();
        expect(c.listState, ConfigLoad.ready);
        expect(c.items.any((a) => a.name == seededManual), isTrue);
      });

      test('manual run succeeds with notify step', () async {
        final c = auto();
        expect(await c.runNow(seededManual), isTrue, reason: c.runError);
        final settled = await c.pollRunUntilSettled(c.lastRun!.id);
        expect(
          settled?.status,
          AutomationRunStatus.AUTOMATION_RUN_STATUS_SUCCEEDED,
        );
        await c.loadStepRuns(settled!.id);
        expect(c.stepRuns.length, greaterThanOrEqualTo(2));
      });

      test('condition mismatch skips run', () async {
        final c = auto();
        expect(await c.runNow(condFail, triggerPayloadJson: '{}'), isTrue);
        final settled = await c.pollRunUntilSettled(c.lastRun!.id);
        expect(
          settled?.status,
          AutomationRunStatus.AUTOMATION_RUN_STATUS_SKIPPED,
        );
      });

      test('webhook failure fails run', () async {
        final c = auto();
        expect(await c.runNow(webhookFail), isTrue);
        final settled = await c.pollRunUntilSettled(c.lastRun!.id);
        expect(
          settled?.status,
          AutomationRunStatus.AUTOMATION_RUN_STATUS_FAILED,
        );
        await c.loadStepRuns(settled!.id);
        expect(
          c.stepRuns.first.status,
          AutomationStepRunStatus.AUTOMATION_STEP_RUN_STATUS_FAILED,
        );
      });
    },
  );
}
