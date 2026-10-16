import 'package:butter_mobile/features/cron/cron_draft.dart';
import 'package:butter_mobile/gen/agents/v1/cron.pb.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  test('rejects legacy CHANNEL delivery', () {
    final draft = CronJobDraft.create()
      ..name = 'nightly'
      ..schedule = '@daily'
      ..agentId = 'a1'
      ..deliveryType = CronDeliveryType.CRON_DELIVERY_TYPE_CHANNEL;
    expect(draft.validateDelivery(), contains('Legacy CHANNEL'));
  });

  test('builds telegram destination delivery', () {
    final draft = CronJobDraft.create()
      ..name = 'tg-job'
      ..schedule = '@hourly'
      ..agentId = 'a1'
      ..deliveryType = CronDeliveryType.CRON_DELIVERY_TYPE_TELEGRAM_DESTINATION
      ..telegramDestinationId = 'dest-1';
    expect(draft.firstError(), isNull);
    final job = draft.build();
    expect(job.delivery.telegramDestinationId, 'dest-1');
  });
}
