import 'dart:convert';

import 'package:fixnum/fixnum.dart';
import 'package:protobuf/well_known_types/google/protobuf/duration.pb.dart'
    as pbd;

import '../../core/automation_errors.dart';
import '../../gen/agents/v1/automation.pb.dart';

class ConditionDraft {
  ConditionDraft({
    this.selector = '',
    this.operator =
        AutomationConditionOperator.AUTOMATION_CONDITION_OPERATOR_EQUALS,
    this.value = '',
  });

  String selector;
  AutomationConditionOperator operator;
  String value;

  factory ConditionDraft.fromProto(AutomationCondition c) => ConditionDraft(
    selector: c.selector,
    operator: c.operator,
    value: c.value,
  );

  AutomationCondition toProto() => AutomationCondition(
    selector: selector.trim(),
    operator: operator,
    value: value,
  );

  String? validate() =>
      selector.trim().isEmpty ? 'Condition selector is required.' : null;
}

class StepDraft {
  StepDraft({
    this.name = '',
    this.type = AutomationStepType.AUTOMATION_STEP_TYPE_CALL_WEBHOOK,
  });

  String name;
  AutomationStepType type;
  String agentId = '';
  String agentInput = '';
  String modelOverride = '';
  String webhookUrl = '';
  String webhookMethod = 'POST';
  String webhookPayloadJson = '';
  String notifyGroupName = '';
  String notifyTitle = '';
  String notifyMessage = '';
  String forumThreadId = '';
  String forumBody = '';

  factory StepDraft.fromProto(AutomationStep s) {
    final d = StepDraft(name: s.name, type: s.type);
    if (s.hasInvokeAgent()) {
      d.agentId = s.invokeAgent.agentId;
      d.agentInput = s.invokeAgent.input;
      d.modelOverride = s.invokeAgent.modelOverride;
    }
    if (s.hasCallWebhook()) {
      d.webhookUrl = s.callWebhook.url;
      d.webhookMethod = s.callWebhook.method.isEmpty
          ? 'POST'
          : s.callWebhook.method;
      d.webhookPayloadJson = s.callWebhook.payloadJson;
    }
    if (s.hasSendNotifyGroup()) {
      d.notifyGroupName = s.sendNotifyGroup.notifyGroupName;
      d.notifyTitle = s.sendNotifyGroup.title;
      d.notifyMessage = s.sendNotifyGroup.message;
    }
    if (s.hasCreateForumPost()) {
      d.forumThreadId = s.createForumPost.threadId;
      d.forumBody = s.createForumPost.body;
    }
    return d;
  }

  AutomationStep toProto() {
    final s = AutomationStep(name: name.trim(), type: type);
    switch (type) {
      case AutomationStepType.AUTOMATION_STEP_TYPE_INVOKE_AGENT:
        s.invokeAgent = AutomationInvokeAgentStep(
          agentId: agentId.trim(),
          input: agentInput,
          modelOverride: modelOverride.trim(),
        );
      case AutomationStepType.AUTOMATION_STEP_TYPE_CALL_WEBHOOK:
        s.callWebhook = AutomationCallWebhookStep(
          url: webhookUrl.trim(),
          method: webhookMethod.trim().isEmpty ? 'POST' : webhookMethod.trim(),
          payloadJson: webhookPayloadJson,
        );
      case AutomationStepType.AUTOMATION_STEP_TYPE_SEND_NOTIFY_GROUP:
        s.sendNotifyGroup = AutomationSendNotifyGroupStep(
          notifyGroupName: notifyGroupName.trim(),
          title: notifyTitle,
          message: notifyMessage,
        );
      case AutomationStepType.AUTOMATION_STEP_TYPE_CREATE_FORUM_POST:
        s.createForumPost = AutomationCreateForumPostStep(
          threadId: forumThreadId.trim(),
          body: forumBody,
        );
      default:
        break;
    }
    return s;
  }

  String? validate() {
    if (name.trim().isEmpty) return 'Step name is required.';
    switch (type) {
      case AutomationStepType.AUTOMATION_STEP_TYPE_INVOKE_AGENT:
        if (agentId.trim().isEmpty) return 'Choose an Agent for this step.';
      case AutomationStepType.AUTOMATION_STEP_TYPE_CALL_WEBHOOK:
        if (webhookUrl.trim().isEmpty) return 'Webhook URL is required.';
        if (webhookPayloadJson.trim().isNotEmpty) {
          try {
            jsonDecode(webhookPayloadJson);
          } catch (_) {
            return 'Webhook payload must be valid JSON.';
          }
        }
      case AutomationStepType.AUTOMATION_STEP_TYPE_SEND_NOTIFY_GROUP:
        if (notifyGroupName.trim().isEmpty) {
          return 'Choose a Notify Group for this step.';
        }
      case AutomationStepType.AUTOMATION_STEP_TYPE_CREATE_FORUM_POST:
        if (forumThreadId.trim().isEmpty || forumBody.trim().isEmpty) {
          return 'Forum thread ID and body are required.';
        }
      default:
        return 'Choose a step type.';
    }
    return null;
  }
}

class AutomationDraft {
  AutomationDraft.create()
    : base = null,
      triggerType = AutomationTriggerType.AUTOMATION_TRIGGER_TYPE_MANUAL {
    _initial = snapshot();
  }

  AutomationDraft.edit(Automation a) : base = a.deepCopy() {
    name = a.name;
    enabled = a.enabled;
    if (a.hasTrigger()) {
      triggerType = a.trigger.type;
      schedule = a.trigger.schedule.schedule;
      timezone = a.trigger.schedule.timezone;
    }
    conditions.addAll(a.conditions.map(ConditionDraft.fromProto));
    steps.addAll(a.steps.map(StepDraft.fromProto));
    if (a.hasPolicy()) {
      if (a.policy.hasTimeout()) {
        timeoutSeconds = a.policy.timeout.seconds.toString();
      }
      if (a.policy.hasRetry()) {
        retryMaxAttempts = a.policy.retry.maxAttempts.toString();
        if (a.policy.retry.hasBackoff()) {
          retryBackoffSeconds = a.policy.retry.backoff.seconds.toString();
        }
      }
      concurrency = a.policy.concurrency;
      if (a.policy.maxOutputBytes != 0) {
        maxOutputBytes = a.policy.maxOutputBytes.toString();
      }
    }
    metadata.addAll(a.metadata);
    _initial = snapshot();
  }

  final Automation? base;
  bool get isCreate => base == null;

  String name = '';
  bool enabled = true;
  AutomationTriggerType triggerType =
      AutomationTriggerType.AUTOMATION_TRIGGER_TYPE_MANUAL;
  String schedule = '@daily';
  String timezone = '';
  final List<ConditionDraft> conditions = [];
  final List<StepDraft> steps = [];
  String timeoutSeconds = '';
  String retryMaxAttempts = '0';
  String retryBackoffSeconds = '0';
  AutomationConcurrencyPolicy concurrency =
      AutomationConcurrencyPolicy.AUTOMATION_CONCURRENCY_POLICY_SKIP;
  String maxOutputBytes = '';
  final Map<String, String> metadata = {};

  late final String _initial;

  String snapshot() => [
    name,
    enabled.toString(),
    triggerType.value.toString(),
    schedule,
    timezone,
    for (final c in conditions) '${c.selector}|${c.operator.value}|${c.value}',
    for (final s in steps) '${s.name}|${s.type.value}|${s.webhookUrl}',
    timeoutSeconds,
    retryMaxAttempts,
    concurrency.value.toString(),
    maxOutputBytes,
    for (final e in metadata.entries) '${e.key}\u0001${e.value}',
  ].join('\u0000');

  bool get dirty => snapshot() != _initial;

  String? validate() {
    if (name.trim().isEmpty) return 'Name is required.';
    if (!automationTriggerUsable(triggerType)) {
      return 'This trigger type is not available on the server yet.';
    }
    if (triggerType == AutomationTriggerType.AUTOMATION_TRIGGER_TYPE_SCHEDULE &&
        schedule.trim().isEmpty) {
      return 'Schedule expression is required.';
    }
    for (final c in conditions) {
      final err = c.validate();
      if (err != null) return err;
    }
    if (steps.isEmpty) return 'Add at least one step.';
    final names = <String>{};
    for (final s in steps) {
      final err = s.validate();
      if (err != null) return err;
      if (!names.add(s.name.trim())) {
        return 'Step names must be unique.';
      }
    }
    return null;
  }

  Automation build() {
    final a = Automation(name: name.trim(), enabled: enabled)
      ..metadata.addAll(metadata);
    if (base != null) {
      a.workspaceId = base!.workspaceId;
    }
    a.trigger = AutomationTrigger(
      type: triggerType,
      schedule: AutomationScheduleTrigger(
        schedule: schedule.trim(),
        timezone: timezone.trim(),
      ),
    );
    a.conditions.addAll(conditions.map((c) => c.toProto()));
    a.steps.addAll(steps.map((s) => s.toProto()));
    final policy = AutomationPolicy(concurrency: concurrency);
    if (timeoutSeconds.trim().isNotEmpty) {
      final sec = int.tryParse(timeoutSeconds.trim());
      if (sec != null && sec > 0) {
        policy.timeout = pbd.Duration(seconds: Int64(sec));
      }
    }
    final attempts = int.tryParse(retryMaxAttempts.trim()) ?? 0;
    final backoff = int.tryParse(retryBackoffSeconds.trim()) ?? 0;
    if (attempts > 0 || backoff > 0) {
      policy.retry = AutomationRetryPolicy(maxAttempts: attempts);
      if (backoff > 0) {
        policy.retry.backoff = pbd.Duration(seconds: Int64(backoff));
      }
    }
    final maxOut = int.tryParse(maxOutputBytes.trim());
    if (maxOut != null && maxOut > 0) {
      policy.maxOutputBytes = maxOut;
    }
    a.policy = policy;
    return a;
  }
}

String automationStepTypeLabel(AutomationStepType t) => switch (t) {
  AutomationStepType.AUTOMATION_STEP_TYPE_INVOKE_AGENT => 'Invoke Agent',
  AutomationStepType.AUTOMATION_STEP_TYPE_CALL_WEBHOOK => 'Call webhook',
  AutomationStepType.AUTOMATION_STEP_TYPE_SEND_NOTIFY_GROUP =>
    'Send Notify Group',
  AutomationStepType.AUTOMATION_STEP_TYPE_CREATE_FORUM_POST =>
    'Create forum post',
  _ => 'Unspecified',
};

String automationConditionOperatorLabel(
  AutomationConditionOperator o,
) => switch (o) {
  AutomationConditionOperator.AUTOMATION_CONDITION_OPERATOR_EQUALS => 'Equals',
  AutomationConditionOperator.AUTOMATION_CONDITION_OPERATOR_NOT_EQUALS =>
    'Not equals',
  AutomationConditionOperator.AUTOMATION_CONDITION_OPERATOR_CONTAINS =>
    'Contains',
  AutomationConditionOperator.AUTOMATION_CONDITION_OPERATOR_REGEX_MATCH =>
    'Regex match',
  AutomationConditionOperator.AUTOMATION_CONDITION_OPERATOR_EXISTS => 'Exists',
  AutomationConditionOperator.AUTOMATION_CONDITION_OPERATOR_NOT_EXISTS =>
    'Not exists',
  _ => 'Unspecified',
};
