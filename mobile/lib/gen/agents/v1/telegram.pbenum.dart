// This is a generated file - do not edit.
//
// Generated from agents/v1/telegram.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names, prefer_relative_imports

import 'dart:core' as $core;

import 'package:protobuf/protobuf.dart' as $pb;

/// TelegramReceiveMode selects how a Channel consumes Telegram updates.
class TelegramReceiveMode extends $pb.ProtobufEnum {
  static const TelegramReceiveMode TELEGRAM_RECEIVE_MODE_UNSPECIFIED =
      TelegramReceiveMode._(
          0, _omitEnumNames ? '' : 'TELEGRAM_RECEIVE_MODE_UNSPECIFIED');

  /// Telegram pushes updates to a per-Channel callback URL. The default, and
  /// the only mode that works correctly across multiple Pods.
  static const TelegramReceiveMode TELEGRAM_RECEIVE_MODE_WEBHOOK =
      TelegramReceiveMode._(
          1, _omitEnumNames ? '' : 'TELEGRAM_RECEIVE_MODE_WEBHOOK');

  /// One leader Pod long-polls getUpdates. Intended for local development.
  static const TelegramReceiveMode TELEGRAM_RECEIVE_MODE_LONG_POLLING =
      TelegramReceiveMode._(
          2, _omitEnumNames ? '' : 'TELEGRAM_RECEIVE_MODE_LONG_POLLING');

  static const $core.List<TelegramReceiveMode> values = <TelegramReceiveMode>[
    TELEGRAM_RECEIVE_MODE_UNSPECIFIED,
    TELEGRAM_RECEIVE_MODE_WEBHOOK,
    TELEGRAM_RECEIVE_MODE_LONG_POLLING,
  ];

  static final $core.List<TelegramReceiveMode?> _byValue =
      $pb.ProtobufEnum.$_initByValueList(values, 2);
  static TelegramReceiveMode? valueOf($core.int value) =>
      value < 0 || value >= _byValue.length ? null : _byValue[value];

  const TelegramReceiveMode._(super.value, super.name);
}

/// TelegramCredentialState reports whether a Channel has a usable Bot Token
/// without ever revealing it.
class TelegramCredentialState extends $pb.ProtobufEnum {
  static const TelegramCredentialState TELEGRAM_CREDENTIAL_STATE_UNSPECIFIED =
      TelegramCredentialState._(
          0, _omitEnumNames ? '' : 'TELEGRAM_CREDENTIAL_STATE_UNSPECIFIED');

  /// No credential is stored for this Channel.
  static const TelegramCredentialState TELEGRAM_CREDENTIAL_STATE_MISSING =
      TelegramCredentialState._(
          1, _omitEnumNames ? '' : 'TELEGRAM_CREDENTIAL_STATE_MISSING');

  /// A credential is stored and last resolved to the Channel's Bot ID.
  static const TelegramCredentialState TELEGRAM_CREDENTIAL_STATE_VALID =
      TelegramCredentialState._(
          2, _omitEnumNames ? '' : 'TELEGRAM_CREDENTIAL_STATE_VALID');

  /// A credential is stored but Telegram rejected it on last use.
  static const TelegramCredentialState TELEGRAM_CREDENTIAL_STATE_INVALID =
      TelegramCredentialState._(
          3, _omitEnumNames ? '' : 'TELEGRAM_CREDENTIAL_STATE_INVALID');

  static const $core.List<TelegramCredentialState> values =
      <TelegramCredentialState>[
    TELEGRAM_CREDENTIAL_STATE_UNSPECIFIED,
    TELEGRAM_CREDENTIAL_STATE_MISSING,
    TELEGRAM_CREDENTIAL_STATE_VALID,
    TELEGRAM_CREDENTIAL_STATE_INVALID,
  ];

  static final $core.List<TelegramCredentialState?> _byValue =
      $pb.ProtobufEnum.$_initByValueList(values, 3);
  static TelegramCredentialState? valueOf($core.int value) =>
      value < 0 || value >= _byValue.length ? null : _byValue[value];

  const TelegramCredentialState._(super.value, super.name);
}

/// TelegramTriggerMode selects which messages at a Destination invoke the
/// Agent. Defaults to ALL.
class TelegramTriggerMode extends $pb.ProtobufEnum {
  static const TelegramTriggerMode TELEGRAM_TRIGGER_MODE_UNSPECIFIED =
      TelegramTriggerMode._(
          0, _omitEnumNames ? '' : 'TELEGRAM_TRIGGER_MODE_UNSPECIFIED');

  /// Every supported message invokes the Agent.
  static const TelegramTriggerMode TELEGRAM_TRIGGER_MODE_ALL =
      TelegramTriggerMode._(
          1, _omitEnumNames ? '' : 'TELEGRAM_TRIGGER_MODE_ALL');

  /// Only messages that mention the current Bot.
  static const TelegramTriggerMode TELEGRAM_TRIGGER_MODE_MENTION =
      TelegramTriggerMode._(
          2, _omitEnumNames ? '' : 'TELEGRAM_TRIGGER_MODE_MENTION');

  /// Only messages that start with a command entity.
  static const TelegramTriggerMode TELEGRAM_TRIGGER_MODE_COMMAND =
      TelegramTriggerMode._(
          3, _omitEnumNames ? '' : 'TELEGRAM_TRIGGER_MODE_COMMAND');

  /// Either of the above.
  static const TelegramTriggerMode TELEGRAM_TRIGGER_MODE_MENTION_OR_COMMAND =
      TelegramTriggerMode._(
          4, _omitEnumNames ? '' : 'TELEGRAM_TRIGGER_MODE_MENTION_OR_COMMAND');

  static const $core.List<TelegramTriggerMode> values = <TelegramTriggerMode>[
    TELEGRAM_TRIGGER_MODE_UNSPECIFIED,
    TELEGRAM_TRIGGER_MODE_ALL,
    TELEGRAM_TRIGGER_MODE_MENTION,
    TELEGRAM_TRIGGER_MODE_COMMAND,
    TELEGRAM_TRIGGER_MODE_MENTION_OR_COMMAND,
  ];

  static final $core.List<TelegramTriggerMode?> _byValue =
      $pb.ProtobufEnum.$_initByValueList(values, 4);
  static TelegramTriggerMode? valueOf($core.int value) =>
      value < 0 || value >= _byValue.length ? null : _byValue[value];

  const TelegramTriggerMode._(super.value, super.name);
}

/// TelegramSessionPolicy selects the session subject inside one Destination.
class TelegramSessionPolicy extends $pb.ProtobufEnum {
  static const TelegramSessionPolicy TELEGRAM_SESSION_POLICY_UNSPECIFIED =
      TelegramSessionPolicy._(
          0, _omitEnumNames ? '' : 'TELEGRAM_SESSION_POLICY_UNSPECIFIED');

  /// All admitted users at the Destination share one conversation. Default.
  static const TelegramSessionPolicy TELEGRAM_SESSION_POLICY_DESTINATION =
      TelegramSessionPolicy._(
          1, _omitEnumNames ? '' : 'TELEGRAM_SESSION_POLICY_DESTINATION');

  /// Each admitted Telegram user keeps a separate conversation.
  static const TelegramSessionPolicy TELEGRAM_SESSION_POLICY_USER =
      TelegramSessionPolicy._(
          2, _omitEnumNames ? '' : 'TELEGRAM_SESSION_POLICY_USER');

  static const $core.List<TelegramSessionPolicy> values =
      <TelegramSessionPolicy>[
    TELEGRAM_SESSION_POLICY_UNSPECIFIED,
    TELEGRAM_SESSION_POLICY_DESTINATION,
    TELEGRAM_SESSION_POLICY_USER,
  ];

  static final $core.List<TelegramSessionPolicy?> _byValue =
      $pb.ProtobufEnum.$_initByValueList(values, 2);
  static TelegramSessionPolicy? valueOf($core.int value) =>
      value < 0 || value >= _byValue.length ? null : _byValue[value];

  const TelegramSessionPolicy._(super.value, super.name);
}

/// TelegramReplyMode selects whether responses quote the inbound message.
/// It is independent of Forum Topic targeting: every outbound message always
/// carries the Destination's optional thread ID regardless of this setting.
class TelegramReplyMode extends $pb.ProtobufEnum {
  static const TelegramReplyMode TELEGRAM_REPLY_MODE_UNSPECIFIED =
      TelegramReplyMode._(
          0, _omitEnumNames ? '' : 'TELEGRAM_REPLY_MODE_UNSPECIFIED');

  /// Quote the inbound message.
  static const TelegramReplyMode TELEGRAM_REPLY_MODE_REPLY =
      TelegramReplyMode._(1, _omitEnumNames ? '' : 'TELEGRAM_REPLY_MODE_REPLY');

  /// Send a standalone message.
  static const TelegramReplyMode TELEGRAM_REPLY_MODE_NEW_MESSAGE =
      TelegramReplyMode._(
          2, _omitEnumNames ? '' : 'TELEGRAM_REPLY_MODE_NEW_MESSAGE');

  static const $core.List<TelegramReplyMode> values = <TelegramReplyMode>[
    TELEGRAM_REPLY_MODE_UNSPECIFIED,
    TELEGRAM_REPLY_MODE_REPLY,
    TELEGRAM_REPLY_MODE_NEW_MESSAGE,
  ];

  static final $core.List<TelegramReplyMode?> _byValue =
      $pb.ProtobufEnum.$_initByValueList(values, 2);
  static TelegramReplyMode? valueOf($core.int value) =>
      value < 0 || value >= _byValue.length ? null : _byValue[value];

  const TelegramReplyMode._(super.value, super.name);
}

/// TelegramWebhookState reports how the Channel's callback registration
/// compares with what Telegram currently has.
class TelegramWebhookState extends $pb.ProtobufEnum {
  static const TelegramWebhookState TELEGRAM_WEBHOOK_STATE_UNSPECIFIED =
      TelegramWebhookState._(
          0, _omitEnumNames ? '' : 'TELEGRAM_WEBHOOK_STATE_UNSPECIFIED');

  /// The Channel does not use Webhook mode.
  static const TelegramWebhookState TELEGRAM_WEBHOOK_STATE_NOT_APPLICABLE =
      TelegramWebhookState._(
          1, _omitEnumNames ? '' : 'TELEGRAM_WEBHOOK_STATE_NOT_APPLICABLE');

  /// Desired and observed registration agree.
  static const TelegramWebhookState TELEGRAM_WEBHOOK_STATE_REGISTERED =
      TelegramWebhookState._(
          2, _omitEnumNames ? '' : 'TELEGRAM_WEBHOOK_STATE_REGISTERED');

  /// Telegram has no webhook, or a different URL, for this Bot.
  static const TelegramWebhookState TELEGRAM_WEBHOOK_STATE_PENDING =
      TelegramWebhookState._(
          3, _omitEnumNames ? '' : 'TELEGRAM_WEBHOOK_STATE_PENDING');

  /// The last reconciliation attempt failed.
  static const TelegramWebhookState TELEGRAM_WEBHOOK_STATE_FAILED =
      TelegramWebhookState._(
          4, _omitEnumNames ? '' : 'TELEGRAM_WEBHOOK_STATE_FAILED');

  static const $core.List<TelegramWebhookState> values = <TelegramWebhookState>[
    TELEGRAM_WEBHOOK_STATE_UNSPECIFIED,
    TELEGRAM_WEBHOOK_STATE_NOT_APPLICABLE,
    TELEGRAM_WEBHOOK_STATE_REGISTERED,
    TELEGRAM_WEBHOOK_STATE_PENDING,
    TELEGRAM_WEBHOOK_STATE_FAILED,
  ];

  static final $core.List<TelegramWebhookState?> _byValue =
      $pb.ProtobufEnum.$_initByValueList(values, 4);
  static TelegramWebhookState? valueOf($core.int value) =>
      value < 0 || value >= _byValue.length ? null : _byValue[value];

  const TelegramWebhookState._(super.value, super.name);
}

class TelegramProcessingStatus extends $pb.ProtobufEnum {
  static const TelegramProcessingStatus TELEGRAM_PROCESSING_STATUS_UNSPECIFIED =
      TelegramProcessingStatus._(
          0, _omitEnumNames ? '' : 'TELEGRAM_PROCESSING_STATUS_UNSPECIFIED');

  /// Claimed from the queue, no Agent work started.
  static const TelegramProcessingStatus TELEGRAM_PROCESSING_STATUS_RECEIVED =
      TelegramProcessingStatus._(
          1, _omitEnumNames ? '' : 'TELEGRAM_PROCESSING_STATUS_RECEIVED');

  /// Agent work may have started. A crash from here is not safely retryable.
  static const TelegramProcessingStatus TELEGRAM_PROCESSING_STATUS_PROCESSING =
      TelegramProcessingStatus._(
          2, _omitEnumNames ? '' : 'TELEGRAM_PROCESSING_STATUS_PROCESSING');

  /// Agent output is safely persisted; only delivery remains.
  static const TelegramProcessingStatus
      TELEGRAM_PROCESSING_STATUS_READY_TO_DELIVER = TelegramProcessingStatus._(
          3,
          _omitEnumNames ? '' : 'TELEGRAM_PROCESSING_STATUS_READY_TO_DELIVER');

  /// Delivered.
  static const TelegramProcessingStatus TELEGRAM_PROCESSING_STATUS_SUCCEEDED =
      TelegramProcessingStatus._(
          4, _omitEnumNames ? '' : 'TELEGRAM_PROCESSING_STATUS_SUCCEEDED');

  /// Failed before Agent work started, or failed to deliver. Safe to retry.
  static const TelegramProcessingStatus TELEGRAM_PROCESSING_STATUS_FAILED =
      TelegramProcessingStatus._(
          5, _omitEnumNames ? '' : 'TELEGRAM_PROCESSING_STATUS_FAILED');

  /// Failed after Agent work may have started. Deliberately NOT retried
  /// automatically: a rerun could repeat tool side effects.
  static const TelegramProcessingStatus
      TELEGRAM_PROCESSING_STATUS_FAILED_UNCERTAIN = TelegramProcessingStatus._(
          6,
          _omitEnumNames ? '' : 'TELEGRAM_PROCESSING_STATUS_FAILED_UNCERTAIN');

  static const $core.List<TelegramProcessingStatus> values =
      <TelegramProcessingStatus>[
    TELEGRAM_PROCESSING_STATUS_UNSPECIFIED,
    TELEGRAM_PROCESSING_STATUS_RECEIVED,
    TELEGRAM_PROCESSING_STATUS_PROCESSING,
    TELEGRAM_PROCESSING_STATUS_READY_TO_DELIVER,
    TELEGRAM_PROCESSING_STATUS_SUCCEEDED,
    TELEGRAM_PROCESSING_STATUS_FAILED,
    TELEGRAM_PROCESSING_STATUS_FAILED_UNCERTAIN,
  ];

  static final $core.List<TelegramProcessingStatus?> _byValue =
      $pb.ProtobufEnum.$_initByValueList(values, 6);
  static TelegramProcessingStatus? valueOf($core.int value) =>
      value < 0 || value >= _byValue.length ? null : _byValue[value];

  const TelegramProcessingStatus._(super.value, super.name);
}

const $core.bool _omitEnumNames =
    $core.bool.fromEnvironment('protobuf.omit_enum_names');
