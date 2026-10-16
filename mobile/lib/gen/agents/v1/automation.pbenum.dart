// This is a generated file - do not edit.
//
// Generated from agents/v1/automation.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names, prefer_relative_imports

import 'dart:core' as $core;

import 'package:protobuf/protobuf.dart' as $pb;

class AutomationTriggerType extends $pb.ProtobufEnum {
  static const AutomationTriggerType AUTOMATION_TRIGGER_TYPE_UNSPECIFIED =
      AutomationTriggerType._(
          0, _omitEnumNames ? '' : 'AUTOMATION_TRIGGER_TYPE_UNSPECIFIED');
  static const AutomationTriggerType AUTOMATION_TRIGGER_TYPE_MANUAL =
      AutomationTriggerType._(
          1, _omitEnumNames ? '' : 'AUTOMATION_TRIGGER_TYPE_MANUAL');
  static const AutomationTriggerType AUTOMATION_TRIGGER_TYPE_SCHEDULE =
      AutomationTriggerType._(
          2, _omitEnumNames ? '' : 'AUTOMATION_TRIGGER_TYPE_SCHEDULE');
  static const AutomationTriggerType AUTOMATION_TRIGGER_TYPE_WEBHOOK =
      AutomationTriggerType._(
          3, _omitEnumNames ? '' : 'AUTOMATION_TRIGGER_TYPE_WEBHOOK');
  static const AutomationTriggerType AUTOMATION_TRIGGER_TYPE_FORUM_EVENT =
      AutomationTriggerType._(
          4, _omitEnumNames ? '' : 'AUTOMATION_TRIGGER_TYPE_FORUM_EVENT');
  static const AutomationTriggerType AUTOMATION_TRIGGER_TYPE_CHANNEL_EVENT =
      AutomationTriggerType._(
          5, _omitEnumNames ? '' : 'AUTOMATION_TRIGGER_TYPE_CHANNEL_EVENT');
  static const AutomationTriggerType AUTOMATION_TRIGGER_TYPE_DAEMON_EVENT =
      AutomationTriggerType._(
          6, _omitEnumNames ? '' : 'AUTOMATION_TRIGGER_TYPE_DAEMON_EVENT');

  static const $core.List<AutomationTriggerType> values =
      <AutomationTriggerType>[
    AUTOMATION_TRIGGER_TYPE_UNSPECIFIED,
    AUTOMATION_TRIGGER_TYPE_MANUAL,
    AUTOMATION_TRIGGER_TYPE_SCHEDULE,
    AUTOMATION_TRIGGER_TYPE_WEBHOOK,
    AUTOMATION_TRIGGER_TYPE_FORUM_EVENT,
    AUTOMATION_TRIGGER_TYPE_CHANNEL_EVENT,
    AUTOMATION_TRIGGER_TYPE_DAEMON_EVENT,
  ];

  static final $core.List<AutomationTriggerType?> _byValue =
      $pb.ProtobufEnum.$_initByValueList(values, 6);
  static AutomationTriggerType? valueOf($core.int value) =>
      value < 0 || value >= _byValue.length ? null : _byValue[value];

  const AutomationTriggerType._(super.value, super.name);
}

class AutomationConditionOperator extends $pb.ProtobufEnum {
  static const AutomationConditionOperator
      AUTOMATION_CONDITION_OPERATOR_UNSPECIFIED = AutomationConditionOperator._(
          0, _omitEnumNames ? '' : 'AUTOMATION_CONDITION_OPERATOR_UNSPECIFIED');
  static const AutomationConditionOperator
      AUTOMATION_CONDITION_OPERATOR_EQUALS = AutomationConditionOperator._(
          1, _omitEnumNames ? '' : 'AUTOMATION_CONDITION_OPERATOR_EQUALS');
  static const AutomationConditionOperator
      AUTOMATION_CONDITION_OPERATOR_NOT_EQUALS = AutomationConditionOperator._(
          2, _omitEnumNames ? '' : 'AUTOMATION_CONDITION_OPERATOR_NOT_EQUALS');
  static const AutomationConditionOperator
      AUTOMATION_CONDITION_OPERATOR_CONTAINS = AutomationConditionOperator._(
          3, _omitEnumNames ? '' : 'AUTOMATION_CONDITION_OPERATOR_CONTAINS');
  static const AutomationConditionOperator
      AUTOMATION_CONDITION_OPERATOR_REGEX_MATCH = AutomationConditionOperator._(
          4, _omitEnumNames ? '' : 'AUTOMATION_CONDITION_OPERATOR_REGEX_MATCH');
  static const AutomationConditionOperator
      AUTOMATION_CONDITION_OPERATOR_EXISTS = AutomationConditionOperator._(
          5, _omitEnumNames ? '' : 'AUTOMATION_CONDITION_OPERATOR_EXISTS');
  static const AutomationConditionOperator
      AUTOMATION_CONDITION_OPERATOR_NOT_EXISTS = AutomationConditionOperator._(
          6, _omitEnumNames ? '' : 'AUTOMATION_CONDITION_OPERATOR_NOT_EXISTS');

  static const $core.List<AutomationConditionOperator> values =
      <AutomationConditionOperator>[
    AUTOMATION_CONDITION_OPERATOR_UNSPECIFIED,
    AUTOMATION_CONDITION_OPERATOR_EQUALS,
    AUTOMATION_CONDITION_OPERATOR_NOT_EQUALS,
    AUTOMATION_CONDITION_OPERATOR_CONTAINS,
    AUTOMATION_CONDITION_OPERATOR_REGEX_MATCH,
    AUTOMATION_CONDITION_OPERATOR_EXISTS,
    AUTOMATION_CONDITION_OPERATOR_NOT_EXISTS,
  ];

  static final $core.List<AutomationConditionOperator?> _byValue =
      $pb.ProtobufEnum.$_initByValueList(values, 6);
  static AutomationConditionOperator? valueOf($core.int value) =>
      value < 0 || value >= _byValue.length ? null : _byValue[value];

  const AutomationConditionOperator._(super.value, super.name);
}

class AutomationStepType extends $pb.ProtobufEnum {
  static const AutomationStepType AUTOMATION_STEP_TYPE_UNSPECIFIED =
      AutomationStepType._(
          0, _omitEnumNames ? '' : 'AUTOMATION_STEP_TYPE_UNSPECIFIED');
  static const AutomationStepType AUTOMATION_STEP_TYPE_INVOKE_AGENT =
      AutomationStepType._(
          1, _omitEnumNames ? '' : 'AUTOMATION_STEP_TYPE_INVOKE_AGENT');
  static const AutomationStepType AUTOMATION_STEP_TYPE_CALL_WEBHOOK =
      AutomationStepType._(
          2, _omitEnumNames ? '' : 'AUTOMATION_STEP_TYPE_CALL_WEBHOOK');
  static const AutomationStepType AUTOMATION_STEP_TYPE_SEND_NOTIFY_GROUP =
      AutomationStepType._(
          3, _omitEnumNames ? '' : 'AUTOMATION_STEP_TYPE_SEND_NOTIFY_GROUP');
  static const AutomationStepType AUTOMATION_STEP_TYPE_CREATE_FORUM_POST =
      AutomationStepType._(
          4, _omitEnumNames ? '' : 'AUTOMATION_STEP_TYPE_CREATE_FORUM_POST');

  static const $core.List<AutomationStepType> values = <AutomationStepType>[
    AUTOMATION_STEP_TYPE_UNSPECIFIED,
    AUTOMATION_STEP_TYPE_INVOKE_AGENT,
    AUTOMATION_STEP_TYPE_CALL_WEBHOOK,
    AUTOMATION_STEP_TYPE_SEND_NOTIFY_GROUP,
    AUTOMATION_STEP_TYPE_CREATE_FORUM_POST,
  ];

  static final $core.List<AutomationStepType?> _byValue =
      $pb.ProtobufEnum.$_initByValueList(values, 4);
  static AutomationStepType? valueOf($core.int value) =>
      value < 0 || value >= _byValue.length ? null : _byValue[value];

  const AutomationStepType._(super.value, super.name);
}

class AutomationConcurrencyPolicy extends $pb.ProtobufEnum {
  static const AutomationConcurrencyPolicy
      AUTOMATION_CONCURRENCY_POLICY_UNSPECIFIED = AutomationConcurrencyPolicy._(
          0, _omitEnumNames ? '' : 'AUTOMATION_CONCURRENCY_POLICY_UNSPECIFIED');
  static const AutomationConcurrencyPolicy AUTOMATION_CONCURRENCY_POLICY_SKIP =
      AutomationConcurrencyPolicy._(
          1, _omitEnumNames ? '' : 'AUTOMATION_CONCURRENCY_POLICY_SKIP');
  static const AutomationConcurrencyPolicy AUTOMATION_CONCURRENCY_POLICY_QUEUE =
      AutomationConcurrencyPolicy._(
          2, _omitEnumNames ? '' : 'AUTOMATION_CONCURRENCY_POLICY_QUEUE');
  static const AutomationConcurrencyPolicy
      AUTOMATION_CONCURRENCY_POLICY_REPLACE = AutomationConcurrencyPolicy._(
          3, _omitEnumNames ? '' : 'AUTOMATION_CONCURRENCY_POLICY_REPLACE');
  static const AutomationConcurrencyPolicy AUTOMATION_CONCURRENCY_POLICY_ALLOW =
      AutomationConcurrencyPolicy._(
          4, _omitEnumNames ? '' : 'AUTOMATION_CONCURRENCY_POLICY_ALLOW');

  static const $core.List<AutomationConcurrencyPolicy> values =
      <AutomationConcurrencyPolicy>[
    AUTOMATION_CONCURRENCY_POLICY_UNSPECIFIED,
    AUTOMATION_CONCURRENCY_POLICY_SKIP,
    AUTOMATION_CONCURRENCY_POLICY_QUEUE,
    AUTOMATION_CONCURRENCY_POLICY_REPLACE,
    AUTOMATION_CONCURRENCY_POLICY_ALLOW,
  ];

  static final $core.List<AutomationConcurrencyPolicy?> _byValue =
      $pb.ProtobufEnum.$_initByValueList(values, 4);
  static AutomationConcurrencyPolicy? valueOf($core.int value) =>
      value < 0 || value >= _byValue.length ? null : _byValue[value];

  const AutomationConcurrencyPolicy._(super.value, super.name);
}

class AutomationRunStatus extends $pb.ProtobufEnum {
  static const AutomationRunStatus AUTOMATION_RUN_STATUS_UNSPECIFIED =
      AutomationRunStatus._(
          0, _omitEnumNames ? '' : 'AUTOMATION_RUN_STATUS_UNSPECIFIED');
  static const AutomationRunStatus AUTOMATION_RUN_STATUS_RUNNING =
      AutomationRunStatus._(
          1, _omitEnumNames ? '' : 'AUTOMATION_RUN_STATUS_RUNNING');
  static const AutomationRunStatus AUTOMATION_RUN_STATUS_SUCCEEDED =
      AutomationRunStatus._(
          2, _omitEnumNames ? '' : 'AUTOMATION_RUN_STATUS_SUCCEEDED');
  static const AutomationRunStatus AUTOMATION_RUN_STATUS_FAILED =
      AutomationRunStatus._(
          3, _omitEnumNames ? '' : 'AUTOMATION_RUN_STATUS_FAILED');
  static const AutomationRunStatus AUTOMATION_RUN_STATUS_SKIPPED =
      AutomationRunStatus._(
          4, _omitEnumNames ? '' : 'AUTOMATION_RUN_STATUS_SKIPPED');
  static const AutomationRunStatus AUTOMATION_RUN_STATUS_CANCELLED =
      AutomationRunStatus._(
          5, _omitEnumNames ? '' : 'AUTOMATION_RUN_STATUS_CANCELLED');

  /// The run's invoke_agent step paused on a Human Input node and waits for a
  /// human reply on the recorded session coordinates (ADR-0003).
  static const AutomationRunStatus AUTOMATION_RUN_STATUS_WAITING_INPUT =
      AutomationRunStatus._(
          6, _omitEnumNames ? '' : 'AUTOMATION_RUN_STATUS_WAITING_INPUT');

  static const $core.List<AutomationRunStatus> values = <AutomationRunStatus>[
    AUTOMATION_RUN_STATUS_UNSPECIFIED,
    AUTOMATION_RUN_STATUS_RUNNING,
    AUTOMATION_RUN_STATUS_SUCCEEDED,
    AUTOMATION_RUN_STATUS_FAILED,
    AUTOMATION_RUN_STATUS_SKIPPED,
    AUTOMATION_RUN_STATUS_CANCELLED,
    AUTOMATION_RUN_STATUS_WAITING_INPUT,
  ];

  static final $core.List<AutomationRunStatus?> _byValue =
      $pb.ProtobufEnum.$_initByValueList(values, 6);
  static AutomationRunStatus? valueOf($core.int value) =>
      value < 0 || value >= _byValue.length ? null : _byValue[value];

  const AutomationRunStatus._(super.value, super.name);
}

class AutomationStepRunStatus extends $pb.ProtobufEnum {
  static const AutomationStepRunStatus AUTOMATION_STEP_RUN_STATUS_UNSPECIFIED =
      AutomationStepRunStatus._(
          0, _omitEnumNames ? '' : 'AUTOMATION_STEP_RUN_STATUS_UNSPECIFIED');
  static const AutomationStepRunStatus AUTOMATION_STEP_RUN_STATUS_RUNNING =
      AutomationStepRunStatus._(
          1, _omitEnumNames ? '' : 'AUTOMATION_STEP_RUN_STATUS_RUNNING');
  static const AutomationStepRunStatus AUTOMATION_STEP_RUN_STATUS_SUCCEEDED =
      AutomationStepRunStatus._(
          2, _omitEnumNames ? '' : 'AUTOMATION_STEP_RUN_STATUS_SUCCEEDED');
  static const AutomationStepRunStatus AUTOMATION_STEP_RUN_STATUS_FAILED =
      AutomationStepRunStatus._(
          3, _omitEnumNames ? '' : 'AUTOMATION_STEP_RUN_STATUS_FAILED');
  static const AutomationStepRunStatus AUTOMATION_STEP_RUN_STATUS_SKIPPED =
      AutomationStepRunStatus._(
          4, _omitEnumNames ? '' : 'AUTOMATION_STEP_RUN_STATUS_SKIPPED');
  static const AutomationStepRunStatus AUTOMATION_STEP_RUN_STATUS_CANCELLED =
      AutomationStepRunStatus._(
          5, _omitEnumNames ? '' : 'AUTOMATION_STEP_RUN_STATUS_CANCELLED');

  static const $core.List<AutomationStepRunStatus> values =
      <AutomationStepRunStatus>[
    AUTOMATION_STEP_RUN_STATUS_UNSPECIFIED,
    AUTOMATION_STEP_RUN_STATUS_RUNNING,
    AUTOMATION_STEP_RUN_STATUS_SUCCEEDED,
    AUTOMATION_STEP_RUN_STATUS_FAILED,
    AUTOMATION_STEP_RUN_STATUS_SKIPPED,
    AUTOMATION_STEP_RUN_STATUS_CANCELLED,
  ];

  static final $core.List<AutomationStepRunStatus?> _byValue =
      $pb.ProtobufEnum.$_initByValueList(values, 5);
  static AutomationStepRunStatus? valueOf($core.int value) =>
      value < 0 || value >= _byValue.length ? null : _byValue[value];

  const AutomationStepRunStatus._(super.value, super.name);
}

const $core.bool _omitEnumNames =
    $core.bool.fromEnvironment('protobuf.omit_enum_names');
