import 'package:butter_mobile/core/automation_errors.dart';
import 'package:butter_mobile/features/automation/automation_draft.dart';
import 'package:butter_mobile/gen/agents/v1/automation.pb.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  test('rejects unwired trigger types', () {
    final d = AutomationDraft.create()
      ..name = 'x'
      ..triggerType = AutomationTriggerType.AUTOMATION_TRIGGER_TYPE_WEBHOOK
      ..steps.add(StepDraft(name: 'a'));
    expect(d.validate(), contains('not available'));
  });

  test('manual trigger with webhook step validates', () {
    final d = AutomationDraft.create()
      ..name = 'flow'
      ..steps.add(
        StepDraft(name: 'hook')..webhookUrl = 'https://example.test/h',
      );
    expect(d.validate(), isNull);
    expect(
      d.build().trigger.type,
      AutomationTriggerType.AUTOMATION_TRIGGER_TYPE_MANUAL,
    );
  });

  test('automationTriggerUsable matches server', () {
    expect(
      automationTriggerUsable(
        AutomationTriggerType.AUTOMATION_TRIGGER_TYPE_MANUAL,
      ),
      isTrue,
    );
    expect(
      automationTriggerUsable(
        AutomationTriggerType.AUTOMATION_TRIGGER_TYPE_DAEMON_EVENT,
      ),
      isFalse,
    );
  });
}
