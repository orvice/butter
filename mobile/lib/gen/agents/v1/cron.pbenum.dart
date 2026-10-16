// This is a generated file - do not edit.
//
// Generated from agents/v1/cron.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names, prefer_relative_imports

import 'dart:core' as $core;

import 'package:protobuf/protobuf.dart' as $pb;

class CronConcurrencyPolicy extends $pb.ProtobufEnum {
  static const CronConcurrencyPolicy CRON_CONCURRENCY_POLICY_UNSPECIFIED =
      CronConcurrencyPolicy._(
          0, _omitEnumNames ? '' : 'CRON_CONCURRENCY_POLICY_UNSPECIFIED');
  static const CronConcurrencyPolicy CRON_CONCURRENCY_POLICY_SKIP =
      CronConcurrencyPolicy._(
          1, _omitEnumNames ? '' : 'CRON_CONCURRENCY_POLICY_SKIP');
  static const CronConcurrencyPolicy CRON_CONCURRENCY_POLICY_QUEUE =
      CronConcurrencyPolicy._(
          2, _omitEnumNames ? '' : 'CRON_CONCURRENCY_POLICY_QUEUE');
  static const CronConcurrencyPolicy CRON_CONCURRENCY_POLICY_REPLACE =
      CronConcurrencyPolicy._(
          3, _omitEnumNames ? '' : 'CRON_CONCURRENCY_POLICY_REPLACE');
  static const CronConcurrencyPolicy CRON_CONCURRENCY_POLICY_ALLOW =
      CronConcurrencyPolicy._(
          4, _omitEnumNames ? '' : 'CRON_CONCURRENCY_POLICY_ALLOW');

  static const $core.List<CronConcurrencyPolicy> values =
      <CronConcurrencyPolicy>[
    CRON_CONCURRENCY_POLICY_UNSPECIFIED,
    CRON_CONCURRENCY_POLICY_SKIP,
    CRON_CONCURRENCY_POLICY_QUEUE,
    CRON_CONCURRENCY_POLICY_REPLACE,
    CRON_CONCURRENCY_POLICY_ALLOW,
  ];

  static final $core.List<CronConcurrencyPolicy?> _byValue =
      $pb.ProtobufEnum.$_initByValueList(values, 4);
  static CronConcurrencyPolicy? valueOf($core.int value) =>
      value < 0 || value >= _byValue.length ? null : _byValue[value];

  const CronConcurrencyPolicy._(super.value, super.name);
}

class CronNotifyOn extends $pb.ProtobufEnum {
  static const CronNotifyOn CRON_NOTIFY_ON_UNSPECIFIED =
      CronNotifyOn._(0, _omitEnumNames ? '' : 'CRON_NOTIFY_ON_UNSPECIFIED');
  static const CronNotifyOn CRON_NOTIFY_ON_ALWAYS =
      CronNotifyOn._(1, _omitEnumNames ? '' : 'CRON_NOTIFY_ON_ALWAYS');
  static const CronNotifyOn CRON_NOTIFY_ON_FAILURE =
      CronNotifyOn._(2, _omitEnumNames ? '' : 'CRON_NOTIFY_ON_FAILURE');
  static const CronNotifyOn CRON_NOTIFY_ON_SUCCESS =
      CronNotifyOn._(3, _omitEnumNames ? '' : 'CRON_NOTIFY_ON_SUCCESS');

  static const $core.List<CronNotifyOn> values = <CronNotifyOn>[
    CRON_NOTIFY_ON_UNSPECIFIED,
    CRON_NOTIFY_ON_ALWAYS,
    CRON_NOTIFY_ON_FAILURE,
    CRON_NOTIFY_ON_SUCCESS,
  ];

  static final $core.List<CronNotifyOn?> _byValue =
      $pb.ProtobufEnum.$_initByValueList(values, 3);
  static CronNotifyOn? valueOf($core.int value) =>
      value < 0 || value >= _byValue.length ? null : _byValue[value];

  const CronNotifyOn._(super.value, super.name);
}

/// CronDeliveryType selects how cron results are delivered.
class CronDeliveryType extends $pb.ProtobufEnum {
  static const CronDeliveryType CRON_DELIVERY_TYPE_UNSPECIFIED =
      CronDeliveryType._(
          0, _omitEnumNames ? '' : 'CRON_DELIVERY_TYPE_UNSPECIFIED');
  static const CronDeliveryType CRON_DELIVERY_TYPE_LOG =
      CronDeliveryType._(1, _omitEnumNames ? '' : 'CRON_DELIVERY_TYPE_LOG');
  static const CronDeliveryType CRON_DELIVERY_TYPE_WEBHOOK =
      CronDeliveryType._(2, _omitEnumNames ? '' : 'CRON_DELIVERY_TYPE_WEBHOOK');

  /// Deprecated: legacy AgentChannel + raw chat ID delivery. Existing jobs
  /// still decode, but new create/update calls reject it for Telegram; use
  /// CRON_DELIVERY_TYPE_TELEGRAM_DESTINATION instead (issue #264).
  static const CronDeliveryType CRON_DELIVERY_TYPE_CHANNEL =
      CronDeliveryType._(3, _omitEnumNames ? '' : 'CRON_DELIVERY_TYPE_CHANNEL');
  static const CronDeliveryType CRON_DELIVERY_TYPE_NOTIFY_GROUP =
      CronDeliveryType._(
          4, _omitEnumNames ? '' : 'CRON_DELIVERY_TYPE_NOTIFY_GROUP');

  /// Deliver to one Telegram Destination through the unified sender.
  static const CronDeliveryType CRON_DELIVERY_TYPE_TELEGRAM_DESTINATION =
      CronDeliveryType._(
          5, _omitEnumNames ? '' : 'CRON_DELIVERY_TYPE_TELEGRAM_DESTINATION');

  static const $core.List<CronDeliveryType> values = <CronDeliveryType>[
    CRON_DELIVERY_TYPE_UNSPECIFIED,
    CRON_DELIVERY_TYPE_LOG,
    CRON_DELIVERY_TYPE_WEBHOOK,
    CRON_DELIVERY_TYPE_CHANNEL,
    CRON_DELIVERY_TYPE_NOTIFY_GROUP,
    CRON_DELIVERY_TYPE_TELEGRAM_DESTINATION,
  ];

  static final $core.List<CronDeliveryType?> _byValue =
      $pb.ProtobufEnum.$_initByValueList(values, 5);
  static CronDeliveryType? valueOf($core.int value) =>
      value < 0 || value >= _byValue.length ? null : _byValue[value];

  const CronDeliveryType._(super.value, super.name);
}

/// CronExecutionStatus indicates the outcome of a cron execution.
class CronExecutionStatus extends $pb.ProtobufEnum {
  static const CronExecutionStatus CRON_EXECUTION_STATUS_UNSPECIFIED =
      CronExecutionStatus._(
          0, _omitEnumNames ? '' : 'CRON_EXECUTION_STATUS_UNSPECIFIED');
  static const CronExecutionStatus CRON_EXECUTION_STATUS_SUCCESS =
      CronExecutionStatus._(
          1, _omitEnumNames ? '' : 'CRON_EXECUTION_STATUS_SUCCESS');
  static const CronExecutionStatus CRON_EXECUTION_STATUS_ERROR =
      CronExecutionStatus._(
          2, _omitEnumNames ? '' : 'CRON_EXECUTION_STATUS_ERROR');
  static const CronExecutionStatus CRON_EXECUTION_STATUS_SKIPPED =
      CronExecutionStatus._(
          3, _omitEnumNames ? '' : 'CRON_EXECUTION_STATUS_SKIPPED');
  static const CronExecutionStatus CRON_EXECUTION_STATUS_CANCELLED =
      CronExecutionStatus._(
          4, _omitEnumNames ? '' : 'CRON_EXECUTION_STATUS_CANCELLED');

  /// The run paused on a Human Input node and waits for a human answer via
  /// SessionService.ReplySession (ADR 0003). Not a terminal state.
  static const CronExecutionStatus CRON_EXECUTION_STATUS_WAITING_INPUT =
      CronExecutionStatus._(
          5, _omitEnumNames ? '' : 'CRON_EXECUTION_STATUS_WAITING_INPUT');

  static const $core.List<CronExecutionStatus> values = <CronExecutionStatus>[
    CRON_EXECUTION_STATUS_UNSPECIFIED,
    CRON_EXECUTION_STATUS_SUCCESS,
    CRON_EXECUTION_STATUS_ERROR,
    CRON_EXECUTION_STATUS_SKIPPED,
    CRON_EXECUTION_STATUS_CANCELLED,
    CRON_EXECUTION_STATUS_WAITING_INPUT,
  ];

  static final $core.List<CronExecutionStatus?> _byValue =
      $pb.ProtobufEnum.$_initByValueList(values, 5);
  static CronExecutionStatus? valueOf($core.int value) =>
      value < 0 || value >= _byValue.length ? null : _byValue[value];

  const CronExecutionStatus._(super.value, super.name);
}

class CronExecutionTriggerType extends $pb.ProtobufEnum {
  static const CronExecutionTriggerType
      CRON_EXECUTION_TRIGGER_TYPE_UNSPECIFIED = CronExecutionTriggerType._(
          0, _omitEnumNames ? '' : 'CRON_EXECUTION_TRIGGER_TYPE_UNSPECIFIED');
  static const CronExecutionTriggerType CRON_EXECUTION_TRIGGER_TYPE_SCHEDULE =
      CronExecutionTriggerType._(
          1, _omitEnumNames ? '' : 'CRON_EXECUTION_TRIGGER_TYPE_SCHEDULE');
  static const CronExecutionTriggerType CRON_EXECUTION_TRIGGER_TYPE_MANUAL =
      CronExecutionTriggerType._(
          2, _omitEnumNames ? '' : 'CRON_EXECUTION_TRIGGER_TYPE_MANUAL');

  static const $core.List<CronExecutionTriggerType> values =
      <CronExecutionTriggerType>[
    CRON_EXECUTION_TRIGGER_TYPE_UNSPECIFIED,
    CRON_EXECUTION_TRIGGER_TYPE_SCHEDULE,
    CRON_EXECUTION_TRIGGER_TYPE_MANUAL,
  ];

  static final $core.List<CronExecutionTriggerType?> _byValue =
      $pb.ProtobufEnum.$_initByValueList(values, 2);
  static CronExecutionTriggerType? valueOf($core.int value) =>
      value < 0 || value >= _byValue.length ? null : _byValue[value];

  const CronExecutionTriggerType._(super.value, super.name);
}

const $core.bool _omitEnumNames =
    $core.bool.fromEnvironment('protobuf.omit_enum_names');
