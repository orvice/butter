// This is a generated file - do not edit.
//
// Generated from agents/v1/cron.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names, prefer_relative_imports

import 'dart:async' as $async;
import 'dart:core' as $core;

import 'package:fixnum/fixnum.dart' as $fixnum;
import 'package:protobuf/protobuf.dart' as $pb;
import 'package:protobuf/well_known_types/google/protobuf/duration.pb.dart'
    as $0;
import 'package:protobuf/well_known_types/google/protobuf/timestamp.pb.dart'
    as $1;

import 'cron.pbenum.dart';

export 'package:protobuf/protobuf.dart' show GeneratedMessageGenericExtensions;

export 'cron.pbenum.dart';

/// CronJob defines a scheduled agent execution.
class CronJob extends $pb.GeneratedMessage {
  factory CronJob({
    $core.String? name,
    $core.String? schedule,
    $core.String? agentName,
    $core.String? input,
    $core.String? timezone,
    $core.bool? enabled,
    CronDelivery? delivery,
    $0.Duration? timeout,
    CronRetryPolicy? retry,
    $core.Iterable<$core.MapEntry<$core.String, $core.String>>? metadata,
    CronConcurrencyPolicy? concurrencyPolicy,
    CronNotifyOn? notifyOn,
    $core.int? maxOutputBytes,
    $core.String? agentId,
    $core.String? workspaceId,
  }) {
    final result = CronJob._();
    if (name != null) result.name = name;
    if (schedule != null) result.schedule = schedule;
    if (agentName != null) result.agentName = agentName;
    if (input != null) result.input = input;
    if (timezone != null) result.timezone = timezone;
    if (enabled != null) result.enabled = enabled;
    if (delivery != null) result.delivery = delivery;
    if (timeout != null) result.timeout = timeout;
    if (retry != null) result.retry = retry;
    if (metadata != null) result.metadata.addEntries(metadata);
    if (concurrencyPolicy != null) result.concurrencyPolicy = concurrencyPolicy;
    if (notifyOn != null) result.notifyOn = notifyOn;
    if (maxOutputBytes != null) result.maxOutputBytes = maxOutputBytes;
    if (agentId != null) result.agentId = agentId;
    if (workspaceId != null) result.workspaceId = workspaceId;
    return result;
  }

  CronJob._();

  factory CronJob.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CronJob()..mergeFromBuffer(data, registry);
  factory CronJob.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CronJob()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CronJob',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: CronJob.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'name')
    ..aOS(2, _omitFieldNames ? '' : 'schedule')
    ..aOS(3, _omitFieldNames ? '' : 'agentName')
    ..aOS(4, _omitFieldNames ? '' : 'input')
    ..aOS(5, _omitFieldNames ? '' : 'timezone')
    ..aOB(6, _omitFieldNames ? '' : 'enabled')
    ..aOM<CronDelivery>(7, _omitFieldNames ? '' : 'delivery',
        subBuilder: CronDelivery.$_createMessage)
    ..aOM<$0.Duration>(8, _omitFieldNames ? '' : 'timeout',
        subBuilder: $0.Duration.$_createMessage)
    ..aOM<CronRetryPolicy>(9, _omitFieldNames ? '' : 'retry',
        subBuilder: CronRetryPolicy.$_createMessage)
    ..m<$core.String, $core.String>(10, _omitFieldNames ? '' : 'metadata',
        entryClassName: 'CronJob.MetadataEntry',
        keyFieldType: $pb.PbFieldType.OS,
        valueFieldType: $pb.PbFieldType.OS,
        packageName: const $pb.PackageName('agents.v1'))
    ..aE<CronConcurrencyPolicy>(11, _omitFieldNames ? '' : 'concurrencyPolicy',
        enumValues: CronConcurrencyPolicy.values)
    ..aE<CronNotifyOn>(12, _omitFieldNames ? '' : 'notifyOn',
        enumValues: CronNotifyOn.values)
    ..aI(13, _omitFieldNames ? '' : 'maxOutputBytes')
    ..aOS(14, _omitFieldNames ? '' : 'agentId')
    ..aOS(100, _omitFieldNames ? '' : 'workspaceId')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CronJob clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CronJob copyWith(void Function(CronJob) updates) =>
      super.copyWith((message) => updates(message as CronJob)) as CronJob;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use CronJob() / CronJob.new instead')
  static CronJob create() => CronJob._();
  static $pb.GeneratedMessage $_createMessage() => CronJob._();
  @$core.override
  CronJob createEmptyInstance() => CronJob._();
  @$core.pragma('dart2js:noInline')
  static CronJob getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<CronJob>(CronJob.$_createMessage);
  static CronJob? _defaultInstance;

  /// Unique name identifying this cron job.
  /// @gotags: json:"name,omitempty" yaml:"name,omitempty"
  @$pb.TagNumber(1)
  $core.String get name => $_getSZ(0);
  @$pb.TagNumber(1)
  set name($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasName() => $_has(0);
  @$pb.TagNumber(1)
  void clearName() => $_clearField(1);

  /// Cron expression (5-field) or predefined schedule (@every, @daily, etc.).
  /// @gotags: json:"schedule,omitempty" yaml:"schedule,omitempty"
  @$pb.TagNumber(2)
  $core.String get schedule => $_getSZ(1);
  @$pb.TagNumber(2)
  set schedule($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasSchedule() => $_has(1);
  @$pb.TagNumber(2)
  void clearSchedule() => $_clearField(2);

  /// Legacy agent name. Deprecated: use agent_id. Retained so historical
  /// records stay readable; the service layer backfills agent_id.
  /// @gotags: json:"agent_name,omitempty" yaml:"agent_name,omitempty"
  @$pb.TagNumber(3)
  $core.String get agentName => $_getSZ(2);
  @$pb.TagNumber(3)
  set agentName($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasAgentName() => $_has(2);
  @$pb.TagNumber(3)
  void clearAgentName() => $_clearField(3);

  /// Input message to send to the agent on each execution.
  /// @gotags: json:"input,omitempty" yaml:"input,omitempty"
  @$pb.TagNumber(4)
  $core.String get input => $_getSZ(3);
  @$pb.TagNumber(4)
  set input($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasInput() => $_has(3);
  @$pb.TagNumber(4)
  void clearInput() => $_clearField(4);

  /// IANA timezone name (e.g. "Asia/Shanghai"). Defaults to UTC if empty.
  /// @gotags: json:"timezone,omitempty" yaml:"timezone,omitempty"
  @$pb.TagNumber(5)
  $core.String get timezone => $_getSZ(4);
  @$pb.TagNumber(5)
  set timezone($core.String value) => $_setString(4, value);
  @$pb.TagNumber(5)
  $core.bool hasTimezone() => $_has(4);
  @$pb.TagNumber(5)
  void clearTimezone() => $_clearField(5);

  /// Whether this cron job is enabled. Defaults to true.
  /// @gotags: json:"enabled,omitempty" yaml:"enabled,omitempty"
  @$pb.TagNumber(6)
  $core.bool get enabled => $_getBF(5);
  @$pb.TagNumber(6)
  set enabled($core.bool value) => $_setBool(5, value);
  @$pb.TagNumber(6)
  $core.bool hasEnabled() => $_has(5);
  @$pb.TagNumber(6)
  void clearEnabled() => $_clearField(6);

  /// Delivery configuration for execution results.
  /// @gotags: json:"delivery,omitempty" yaml:"delivery,omitempty"
  @$pb.TagNumber(7)
  CronDelivery get delivery => $_getN(6);
  @$pb.TagNumber(7)
  set delivery(CronDelivery value) => $_setField(7, value);
  @$pb.TagNumber(7)
  $core.bool hasDelivery() => $_has(6);
  @$pb.TagNumber(7)
  void clearDelivery() => $_clearField(7);
  @$pb.TagNumber(7)
  CronDelivery ensureDelivery() => $_ensure(6);

  /// Maximum duration for a single logical execution. 0 means no explicit
  /// cron-level timeout beyond the caller context.
  /// @gotags: json:"timeout,omitempty" yaml:"timeout,omitempty"
  @$pb.TagNumber(8)
  $0.Duration get timeout => $_getN(7);
  @$pb.TagNumber(8)
  set timeout($0.Duration value) => $_setField(8, value);
  @$pb.TagNumber(8)
  $core.bool hasTimeout() => $_has(7);
  @$pb.TagNumber(8)
  void clearTimeout() => $_clearField(8);
  @$pb.TagNumber(8)
  $0.Duration ensureTimeout() => $_ensure(7);

  /// Retry behavior for failed invocations.
  /// @gotags: json:"retry,omitempty" yaml:"retry,omitempty"
  @$pb.TagNumber(9)
  CronRetryPolicy get retry => $_getN(8);
  @$pb.TagNumber(9)
  set retry(CronRetryPolicy value) => $_setField(9, value);
  @$pb.TagNumber(9)
  $core.bool hasRetry() => $_has(8);
  @$pb.TagNumber(9)
  void clearRetry() => $_clearField(9);
  @$pb.TagNumber(9)
  CronRetryPolicy ensureRetry() => $_ensure(8);

  /// @gotags: json:"metadata,omitempty" yaml:"metadata,omitempty"
  @$pb.TagNumber(10)
  $pb.PbMap<$core.String, $core.String> get metadata => $_getMap(9);

  /// Controls overlapping executions of the same job.
  /// @gotags: json:"concurrency_policy,omitempty" yaml:"concurrency_policy,omitempty"
  @$pb.TagNumber(11)
  CronConcurrencyPolicy get concurrencyPolicy => $_getN(10);
  @$pb.TagNumber(11)
  set concurrencyPolicy(CronConcurrencyPolicy value) => $_setField(11, value);
  @$pb.TagNumber(11)
  $core.bool hasConcurrencyPolicy() => $_has(10);
  @$pb.TagNumber(11)
  void clearConcurrencyPolicy() => $_clearField(11);

  /// Controls which execution outcomes are delivered to configured targets.
  /// @gotags: json:"notify_on,omitempty" yaml:"notify_on,omitempty"
  @$pb.TagNumber(12)
  CronNotifyOn get notifyOn => $_getN(11);
  @$pb.TagNumber(12)
  set notifyOn(CronNotifyOn value) => $_setField(12, value);
  @$pb.TagNumber(12)
  $core.bool hasNotifyOn() => $_has(11);
  @$pb.TagNumber(12)
  void clearNotifyOn() => $_clearField(12);

  /// Maximum bytes stored for execution output previews. 0 uses server default.
  /// @gotags: json:"max_output_bytes,omitempty" yaml:"max_output_bytes,omitempty"
  @$pb.TagNumber(13)
  $core.int get maxOutputBytes => $_getIZ(12);
  @$pb.TagNumber(13)
  set maxOutputBytes($core.int value) => $_setSignedInt32(12, value);
  @$pb.TagNumber(13)
  $core.bool hasMaxOutputBytes() => $_has(12);
  @$pb.TagNumber(13)
  void clearMaxOutputBytes() => $_clearField(13);

  /// Immutable agent_id of the agent to execute. Preferred over agent_name.
  /// @gotags: json:"agent_id,omitempty" yaml:"agent_id,omitempty" bson:"agent_id,omitempty"
  @$pb.TagNumber(14)
  $core.String get agentId => $_getSZ(13);
  @$pb.TagNumber(14)
  set agentId($core.String value) => $_setString(13, value);
  @$pb.TagNumber(14)
  $core.bool hasAgentId() => $_has(13);
  @$pb.TagNumber(14)
  void clearAgentId() => $_clearField(14);

  /// ID of the workspace this cron job belongs to.
  /// @gotags: json:"workspace_id,omitempty" yaml:"workspace_id,omitempty" bson:"workspace_id,omitempty"
  @$pb.TagNumber(100)
  $core.String get workspaceId => $_getSZ(14);
  @$pb.TagNumber(100)
  set workspaceId($core.String value) => $_setString(14, value);
  @$pb.TagNumber(100)
  $core.bool hasWorkspaceId() => $_has(14);
  @$pb.TagNumber(100)
  void clearWorkspaceId() => $_clearField(100);
}

class CronRetryPolicy extends $pb.GeneratedMessage {
  factory CronRetryPolicy({
    $core.int? maxAttempts,
    $0.Duration? backoff,
  }) {
    final result = CronRetryPolicy._();
    if (maxAttempts != null) result.maxAttempts = maxAttempts;
    if (backoff != null) result.backoff = backoff;
    return result;
  }

  CronRetryPolicy._();

  factory CronRetryPolicy.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CronRetryPolicy()..mergeFromBuffer(data, registry);
  factory CronRetryPolicy.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CronRetryPolicy()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CronRetryPolicy',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: CronRetryPolicy.$_createMessage)
    ..aI(1, _omitFieldNames ? '' : 'maxAttempts')
    ..aOM<$0.Duration>(2, _omitFieldNames ? '' : 'backoff',
        subBuilder: $0.Duration.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CronRetryPolicy clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CronRetryPolicy copyWith(void Function(CronRetryPolicy) updates) =>
      super.copyWith((message) => updates(message as CronRetryPolicy))
          as CronRetryPolicy;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use CronRetryPolicy() / CronRetryPolicy.new instead')
  static CronRetryPolicy create() => CronRetryPolicy._();
  static $pb.GeneratedMessage $_createMessage() => CronRetryPolicy._();
  @$core.override
  CronRetryPolicy createEmptyInstance() => CronRetryPolicy._();
  @$core.pragma('dart2js:noInline')
  static CronRetryPolicy getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<CronRetryPolicy>(
          CronRetryPolicy.$_createMessage);
  static CronRetryPolicy? _defaultInstance;

  /// Number of retries after the initial attempt. 0 disables retries.
  /// @gotags: json:"max_attempts,omitempty" yaml:"max_attempts,omitempty"
  @$pb.TagNumber(1)
  $core.int get maxAttempts => $_getIZ(0);
  @$pb.TagNumber(1)
  set maxAttempts($core.int value) => $_setSignedInt32(0, value);
  @$pb.TagNumber(1)
  $core.bool hasMaxAttempts() => $_has(0);
  @$pb.TagNumber(1)
  void clearMaxAttempts() => $_clearField(1);

  /// Backoff between retry attempts. 0 means immediate retry.
  /// @gotags: json:"backoff,omitempty" yaml:"backoff,omitempty"
  @$pb.TagNumber(2)
  $0.Duration get backoff => $_getN(1);
  @$pb.TagNumber(2)
  set backoff($0.Duration value) => $_setField(2, value);
  @$pb.TagNumber(2)
  $core.bool hasBackoff() => $_has(1);
  @$pb.TagNumber(2)
  void clearBackoff() => $_clearField(2);
  @$pb.TagNumber(2)
  $0.Duration ensureBackoff() => $_ensure(1);
}

/// CronDelivery defines how cron execution results are delivered.
class CronDelivery extends $pb.GeneratedMessage {
  factory CronDelivery({
    CronDeliveryType? type,
    $core.String? webhookUrl,
    $core.String? channelName,
    $core.String? chatId,
    $core.String? notifyGroupName,
    $core.String? telegramDestinationId,
  }) {
    final result = CronDelivery._();
    if (type != null) result.type = type;
    if (webhookUrl != null) result.webhookUrl = webhookUrl;
    if (channelName != null) result.channelName = channelName;
    if (chatId != null) result.chatId = chatId;
    if (notifyGroupName != null) result.notifyGroupName = notifyGroupName;
    if (telegramDestinationId != null)
      result.telegramDestinationId = telegramDestinationId;
    return result;
  }

  CronDelivery._();

  factory CronDelivery.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CronDelivery()..mergeFromBuffer(data, registry);
  factory CronDelivery.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CronDelivery()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CronDelivery',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: CronDelivery.$_createMessage)
    ..aE<CronDeliveryType>(1, _omitFieldNames ? '' : 'type',
        enumValues: CronDeliveryType.values)
    ..aOS(2, _omitFieldNames ? '' : 'webhookUrl')
    ..aOS(3, _omitFieldNames ? '' : 'channelName')
    ..aOS(4, _omitFieldNames ? '' : 'chatId')
    ..aOS(5, _omitFieldNames ? '' : 'notifyGroupName')
    ..aOS(6, _omitFieldNames ? '' : 'telegramDestinationId')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CronDelivery clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CronDelivery copyWith(void Function(CronDelivery) updates) =>
      super.copyWith((message) => updates(message as CronDelivery))
          as CronDelivery;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use CronDelivery() / CronDelivery.new instead')
  static CronDelivery create() => CronDelivery._();
  static $pb.GeneratedMessage $_createMessage() => CronDelivery._();
  @$core.override
  CronDelivery createEmptyInstance() => CronDelivery._();
  @$core.pragma('dart2js:noInline')
  static CronDelivery getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<CronDelivery>(
          CronDelivery.$_createMessage);
  static CronDelivery? _defaultInstance;

  /// @gotags: json:"type,omitempty" yaml:"type,omitempty"
  @$pb.TagNumber(1)
  CronDeliveryType get type => $_getN(0);
  @$pb.TagNumber(1)
  set type(CronDeliveryType value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasType() => $_has(0);
  @$pb.TagNumber(1)
  void clearType() => $_clearField(1);

  /// Webhook URL for WEBHOOK delivery type.
  /// @gotags: json:"webhook_url,omitempty" yaml:"webhook_url,omitempty"
  @$pb.TagNumber(2)
  $core.String get webhookUrl => $_getSZ(1);
  @$pb.TagNumber(2)
  set webhookUrl($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasWebhookUrl() => $_has(1);
  @$pb.TagNumber(2)
  void clearWebhookUrl() => $_clearField(2);

  /// Channel name referencing an existing AgentChannel for CHANNEL delivery type.
  /// @gotags: json:"channel_name,omitempty" yaml:"channel_name,omitempty"
  @$pb.TagNumber(3)
  $core.String get channelName => $_getSZ(2);
  @$pb.TagNumber(3)
  set channelName($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasChannelName() => $_has(2);
  @$pb.TagNumber(3)
  void clearChannelName() => $_clearField(3);

  /// Target chat/conversation ID for CHANNEL delivery type.
  /// @gotags: json:"chat_id,omitempty" yaml:"chat_id,omitempty"
  @$pb.TagNumber(4)
  $core.String get chatId => $_getSZ(3);
  @$pb.TagNumber(4)
  set chatId($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasChatId() => $_has(3);
  @$pb.TagNumber(4)
  void clearChatId() => $_clearField(4);

  /// Notify group name for NOTIFY_GROUP delivery type.
  /// @gotags: json:"notify_group_name,omitempty" yaml:"notify_group_name,omitempty"
  @$pb.TagNumber(5)
  $core.String get notifyGroupName => $_getSZ(4);
  @$pb.TagNumber(5)
  set notifyGroupName($core.String value) => $_setString(4, value);
  @$pb.TagNumber(5)
  $core.bool hasNotifyGroupName() => $_has(4);
  @$pb.TagNumber(5)
  void clearNotifyGroupName() => $_clearField(5);

  /// Telegram Destination ID for TELEGRAM_DESTINATION delivery type. The
  /// Destination owns the chat, the optional Forum Topic, and the credential,
  /// so no Bot Token or chat ID is stored on the job (issue #264).
  /// @gotags: json:"telegram_destination_id,omitempty" yaml:"telegram_destination_id,omitempty" bson:"telegram_destination_id,omitempty"
  @$pb.TagNumber(6)
  $core.String get telegramDestinationId => $_getSZ(5);
  @$pb.TagNumber(6)
  set telegramDestinationId($core.String value) => $_setString(5, value);
  @$pb.TagNumber(6)
  $core.bool hasTelegramDestinationId() => $_has(5);
  @$pb.TagNumber(6)
  void clearTelegramDestinationId() => $_clearField(6);
}

/// CronExecution records a single cron job execution.
class CronExecution extends $pb.GeneratedMessage {
  factory CronExecution({
    $core.String? id,
    $core.String? jobName,
    $core.String? agentName,
    CronExecutionStatus? status,
    $core.String? input,
    $core.String? output,
    $1.Timestamp? startedAt,
    $1.Timestamp? finishedAt,
    $core.String? error,
    $fixnum.Int64? durationMs,
    $core.int? attemptCount,
    CronExecutionTriggerType? triggerType,
    $core.String? skippedReason,
    $core.bool? truncated,
    $core.String? sessionAppName,
    $core.String? sessionUserId,
    $core.String? sessionId,
    $core.String? agentId,
    $core.String? workspaceId,
  }) {
    final result = CronExecution._();
    if (id != null) result.id = id;
    if (jobName != null) result.jobName = jobName;
    if (agentName != null) result.agentName = agentName;
    if (status != null) result.status = status;
    if (input != null) result.input = input;
    if (output != null) result.output = output;
    if (startedAt != null) result.startedAt = startedAt;
    if (finishedAt != null) result.finishedAt = finishedAt;
    if (error != null) result.error = error;
    if (durationMs != null) result.durationMs = durationMs;
    if (attemptCount != null) result.attemptCount = attemptCount;
    if (triggerType != null) result.triggerType = triggerType;
    if (skippedReason != null) result.skippedReason = skippedReason;
    if (truncated != null) result.truncated = truncated;
    if (sessionAppName != null) result.sessionAppName = sessionAppName;
    if (sessionUserId != null) result.sessionUserId = sessionUserId;
    if (sessionId != null) result.sessionId = sessionId;
    if (agentId != null) result.agentId = agentId;
    if (workspaceId != null) result.workspaceId = workspaceId;
    return result;
  }

  CronExecution._();

  factory CronExecution.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CronExecution()..mergeFromBuffer(data, registry);
  factory CronExecution.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CronExecution()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CronExecution',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: CronExecution.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..aOS(2, _omitFieldNames ? '' : 'jobName')
    ..aOS(3, _omitFieldNames ? '' : 'agentName')
    ..aE<CronExecutionStatus>(4, _omitFieldNames ? '' : 'status',
        enumValues: CronExecutionStatus.values)
    ..aOS(5, _omitFieldNames ? '' : 'input')
    ..aOS(6, _omitFieldNames ? '' : 'output')
    ..aOM<$1.Timestamp>(7, _omitFieldNames ? '' : 'startedAt',
        subBuilder: $1.Timestamp.$_createMessage)
    ..aOM<$1.Timestamp>(8, _omitFieldNames ? '' : 'finishedAt',
        subBuilder: $1.Timestamp.$_createMessage)
    ..aOS(9, _omitFieldNames ? '' : 'error')
    ..aInt64(10, _omitFieldNames ? '' : 'durationMs')
    ..aI(11, _omitFieldNames ? '' : 'attemptCount')
    ..aE<CronExecutionTriggerType>(12, _omitFieldNames ? '' : 'triggerType',
        enumValues: CronExecutionTriggerType.values)
    ..aOS(13, _omitFieldNames ? '' : 'skippedReason')
    ..aOB(14, _omitFieldNames ? '' : 'truncated')
    ..aOS(15, _omitFieldNames ? '' : 'sessionAppName')
    ..aOS(16, _omitFieldNames ? '' : 'sessionUserId')
    ..aOS(17, _omitFieldNames ? '' : 'sessionId')
    ..aOS(18, _omitFieldNames ? '' : 'agentId')
    ..aOS(100, _omitFieldNames ? '' : 'workspaceId')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CronExecution clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CronExecution copyWith(void Function(CronExecution) updates) =>
      super.copyWith((message) => updates(message as CronExecution))
          as CronExecution;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use CronExecution() / CronExecution.new instead')
  static CronExecution create() => CronExecution._();
  static $pb.GeneratedMessage $_createMessage() => CronExecution._();
  @$core.override
  CronExecution createEmptyInstance() => CronExecution._();
  @$core.pragma('dart2js:noInline')
  static CronExecution getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<CronExecution>(
          CronExecution.$_createMessage);
  static CronExecution? _defaultInstance;

  /// Unique execution identifier.
  /// @gotags: json:"id,omitempty" yaml:"id,omitempty" bson:"_id,omitempty"
  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);

  /// Name of the cron job that was executed.
  /// @gotags: json:"job_name,omitempty" yaml:"job_name,omitempty" bson:"job_name,omitempty"
  @$pb.TagNumber(2)
  $core.String get jobName => $_getSZ(1);
  @$pb.TagNumber(2)
  set jobName($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasJobName() => $_has(1);
  @$pb.TagNumber(2)
  void clearJobName() => $_clearField(2);

  /// Legacy name of the agent that was executed. Historical records written
  /// before the Agent ID migration carry only this field.
  /// @gotags: json:"agent_name,omitempty" yaml:"agent_name,omitempty" bson:"agent_name,omitempty"
  @$pb.TagNumber(3)
  $core.String get agentName => $_getSZ(2);
  @$pb.TagNumber(3)
  set agentName($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasAgentName() => $_has(2);
  @$pb.TagNumber(3)
  void clearAgentName() => $_clearField(3);

  /// Execution status.
  /// @gotags: json:"status,omitempty" yaml:"status,omitempty" bson:"status,omitempty"
  @$pb.TagNumber(4)
  CronExecutionStatus get status => $_getN(3);
  @$pb.TagNumber(4)
  set status(CronExecutionStatus value) => $_setField(4, value);
  @$pb.TagNumber(4)
  $core.bool hasStatus() => $_has(3);
  @$pb.TagNumber(4)
  void clearStatus() => $_clearField(4);

  /// Input message sent to the agent.
  /// @gotags: json:"input,omitempty" yaml:"input,omitempty" bson:"input,omitempty"
  @$pb.TagNumber(5)
  $core.String get input => $_getSZ(4);
  @$pb.TagNumber(5)
  set input($core.String value) => $_setString(4, value);
  @$pb.TagNumber(5)
  $core.bool hasInput() => $_has(4);
  @$pb.TagNumber(5)
  void clearInput() => $_clearField(5);

  /// Agent output text or error message.
  /// @gotags: json:"output,omitempty" yaml:"output,omitempty" bson:"output,omitempty"
  @$pb.TagNumber(6)
  $core.String get output => $_getSZ(5);
  @$pb.TagNumber(6)
  set output($core.String value) => $_setString(5, value);
  @$pb.TagNumber(6)
  $core.bool hasOutput() => $_has(5);
  @$pb.TagNumber(6)
  void clearOutput() => $_clearField(6);

  /// When the execution started.
  /// @gotags: json:"started_at,omitempty" yaml:"started_at,omitempty" bson:"started_at,omitempty"
  @$pb.TagNumber(7)
  $1.Timestamp get startedAt => $_getN(6);
  @$pb.TagNumber(7)
  set startedAt($1.Timestamp value) => $_setField(7, value);
  @$pb.TagNumber(7)
  $core.bool hasStartedAt() => $_has(6);
  @$pb.TagNumber(7)
  void clearStartedAt() => $_clearField(7);
  @$pb.TagNumber(7)
  $1.Timestamp ensureStartedAt() => $_ensure(6);

  /// When the execution finished.
  /// @gotags: json:"finished_at,omitempty" yaml:"finished_at,omitempty" bson:"finished_at,omitempty"
  @$pb.TagNumber(8)
  $1.Timestamp get finishedAt => $_getN(7);
  @$pb.TagNumber(8)
  set finishedAt($1.Timestamp value) => $_setField(8, value);
  @$pb.TagNumber(8)
  $core.bool hasFinishedAt() => $_has(7);
  @$pb.TagNumber(8)
  void clearFinishedAt() => $_clearField(8);
  @$pb.TagNumber(8)
  $1.Timestamp ensureFinishedAt() => $_ensure(7);

  /// Error text when status is ERROR.
  /// @gotags: json:"error,omitempty" yaml:"error,omitempty" bson:"error,omitempty"
  @$pb.TagNumber(9)
  $core.String get error => $_getSZ(8);
  @$pb.TagNumber(9)
  set error($core.String value) => $_setString(8, value);
  @$pb.TagNumber(9)
  $core.bool hasError() => $_has(8);
  @$pb.TagNumber(9)
  void clearError() => $_clearField(9);

  /// Execution duration in milliseconds.
  /// @gotags: json:"duration_ms,omitempty" yaml:"duration_ms,omitempty" bson:"duration_ms,omitempty"
  @$pb.TagNumber(10)
  $fixnum.Int64 get durationMs => $_getI64(9);
  @$pb.TagNumber(10)
  set durationMs($fixnum.Int64 value) => $_setInt64(9, value);
  @$pb.TagNumber(10)
  $core.bool hasDurationMs() => $_has(9);
  @$pb.TagNumber(10)
  void clearDurationMs() => $_clearField(10);

  /// Number of attempts used for this logical execution.
  /// @gotags: json:"attempt_count,omitempty" yaml:"attempt_count,omitempty" bson:"attempt_count,omitempty"
  @$pb.TagNumber(11)
  $core.int get attemptCount => $_getIZ(10);
  @$pb.TagNumber(11)
  set attemptCount($core.int value) => $_setSignedInt32(10, value);
  @$pb.TagNumber(11)
  $core.bool hasAttemptCount() => $_has(10);
  @$pb.TagNumber(11)
  void clearAttemptCount() => $_clearField(11);

  /// What initiated this execution.
  /// @gotags: json:"trigger_type,omitempty" yaml:"trigger_type,omitempty" bson:"trigger_type,omitempty"
  @$pb.TagNumber(12)
  CronExecutionTriggerType get triggerType => $_getN(11);
  @$pb.TagNumber(12)
  set triggerType(CronExecutionTriggerType value) => $_setField(12, value);
  @$pb.TagNumber(12)
  $core.bool hasTriggerType() => $_has(11);
  @$pb.TagNumber(12)
  void clearTriggerType() => $_clearField(12);

  /// Reason populated when status is SKIPPED.
  /// @gotags: json:"skipped_reason,omitempty" yaml:"skipped_reason,omitempty" bson:"skipped_reason,omitempty"
  @$pb.TagNumber(13)
  $core.String get skippedReason => $_getSZ(12);
  @$pb.TagNumber(13)
  set skippedReason($core.String value) => $_setString(12, value);
  @$pb.TagNumber(13)
  $core.bool hasSkippedReason() => $_has(12);
  @$pb.TagNumber(13)
  void clearSkippedReason() => $_clearField(13);

  /// True when the stored output was truncated by max_output_bytes.
  /// @gotags: json:"truncated,omitempty" yaml:"truncated,omitempty" bson:"truncated,omitempty"
  @$pb.TagNumber(14)
  $core.bool get truncated => $_getBF(13);
  @$pb.TagNumber(14)
  set truncated($core.bool value) => $_setBool(13, value);
  @$pb.TagNumber(14)
  $core.bool hasTruncated() => $_has(13);
  @$pb.TagNumber(14)
  void clearTruncated() => $_clearField(14);

  /// Session coordinates for answering a WAITING_INPUT execution via
  /// SessionService.ReplySession (app_name / user_id / session_id).
  /// Populated when status is WAITING_INPUT.
  /// @gotags: json:"session_app_name,omitempty" yaml:"session_app_name,omitempty" bson:"session_app_name,omitempty"
  @$pb.TagNumber(15)
  $core.String get sessionAppName => $_getSZ(14);
  @$pb.TagNumber(15)
  set sessionAppName($core.String value) => $_setString(14, value);
  @$pb.TagNumber(15)
  $core.bool hasSessionAppName() => $_has(14);
  @$pb.TagNumber(15)
  void clearSessionAppName() => $_clearField(15);

  /// @gotags: json:"session_user_id,omitempty" yaml:"session_user_id,omitempty" bson:"session_user_id,omitempty"
  @$pb.TagNumber(16)
  $core.String get sessionUserId => $_getSZ(15);
  @$pb.TagNumber(16)
  set sessionUserId($core.String value) => $_setString(15, value);
  @$pb.TagNumber(16)
  $core.bool hasSessionUserId() => $_has(15);
  @$pb.TagNumber(16)
  void clearSessionUserId() => $_clearField(16);

  /// @gotags: json:"session_id,omitempty" yaml:"session_id,omitempty" bson:"session_id,omitempty"
  @$pb.TagNumber(17)
  $core.String get sessionId => $_getSZ(16);
  @$pb.TagNumber(17)
  set sessionId($core.String value) => $_setString(16, value);
  @$pb.TagNumber(17)
  $core.bool hasSessionId() => $_has(16);
  @$pb.TagNumber(17)
  void clearSessionId() => $_clearField(17);

  /// Immutable agent_id of the executed agent, when known.
  /// @gotags: json:"agent_id,omitempty" yaml:"agent_id,omitempty" bson:"agent_id,omitempty"
  @$pb.TagNumber(18)
  $core.String get agentId => $_getSZ(17);
  @$pb.TagNumber(18)
  set agentId($core.String value) => $_setString(17, value);
  @$pb.TagNumber(18)
  $core.bool hasAgentId() => $_has(17);
  @$pb.TagNumber(18)
  void clearAgentId() => $_clearField(18);

  /// ID of the workspace this execution belongs to.
  /// @gotags: json:"workspace_id,omitempty" yaml:"workspace_id,omitempty" bson:"workspace_id,omitempty"
  @$pb.TagNumber(100)
  $core.String get workspaceId => $_getSZ(18);
  @$pb.TagNumber(100)
  set workspaceId($core.String value) => $_setString(18, value);
  @$pb.TagNumber(100)
  $core.bool hasWorkspaceId() => $_has(18);
  @$pb.TagNumber(100)
  void clearWorkspaceId() => $_clearField(100);
}

class ListCronJobsRequest extends $pb.GeneratedMessage {
  factory ListCronJobsRequest() => ListCronJobsRequest._();

  ListCronJobsRequest._();

  factory ListCronJobsRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListCronJobsRequest()..mergeFromBuffer(data, registry);
  factory ListCronJobsRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListCronJobsRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListCronJobsRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: ListCronJobsRequest.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListCronJobsRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListCronJobsRequest copyWith(void Function(ListCronJobsRequest) updates) =>
      super.copyWith((message) => updates(message as ListCronJobsRequest))
          as ListCronJobsRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core
      .Deprecated('Use ListCronJobsRequest() / ListCronJobsRequest.new instead')
  static ListCronJobsRequest create() => ListCronJobsRequest._();
  static $pb.GeneratedMessage $_createMessage() => ListCronJobsRequest._();
  @$core.override
  ListCronJobsRequest createEmptyInstance() => ListCronJobsRequest._();
  @$core.pragma('dart2js:noInline')
  static ListCronJobsRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ListCronJobsRequest>(
          ListCronJobsRequest.$_createMessage);
  static ListCronJobsRequest? _defaultInstance;
}

class ListCronJobsResponse extends $pb.GeneratedMessage {
  factory ListCronJobsResponse({
    $core.Iterable<CronJob>? cronJobs,
  }) {
    final result = ListCronJobsResponse._();
    if (cronJobs != null) result.cronJobs.addAll(cronJobs);
    return result;
  }

  ListCronJobsResponse._();

  factory ListCronJobsResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListCronJobsResponse()..mergeFromBuffer(data, registry);
  factory ListCronJobsResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListCronJobsResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListCronJobsResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: ListCronJobsResponse.$_createMessage)
    ..pPM<CronJob>(1, _omitFieldNames ? '' : 'cronJobs',
        subBuilder: CronJob.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListCronJobsResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListCronJobsResponse copyWith(void Function(ListCronJobsResponse) updates) =>
      super.copyWith((message) => updates(message as ListCronJobsResponse))
          as ListCronJobsResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use ListCronJobsResponse() / ListCronJobsResponse.new instead')
  static ListCronJobsResponse create() => ListCronJobsResponse._();
  static $pb.GeneratedMessage $_createMessage() => ListCronJobsResponse._();
  @$core.override
  ListCronJobsResponse createEmptyInstance() => ListCronJobsResponse._();
  @$core.pragma('dart2js:noInline')
  static ListCronJobsResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ListCronJobsResponse>(
          ListCronJobsResponse.$_createMessage);
  static ListCronJobsResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $pb.PbList<CronJob> get cronJobs => $_getList(0);
}

class GetCronJobRequest extends $pb.GeneratedMessage {
  factory GetCronJobRequest({
    $core.String? name,
  }) {
    final result = GetCronJobRequest._();
    if (name != null) result.name = name;
    return result;
  }

  GetCronJobRequest._();

  factory GetCronJobRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetCronJobRequest()..mergeFromBuffer(data, registry);
  factory GetCronJobRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetCronJobRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetCronJobRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: GetCronJobRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'name')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetCronJobRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetCronJobRequest copyWith(void Function(GetCronJobRequest) updates) =>
      super.copyWith((message) => updates(message as GetCronJobRequest))
          as GetCronJobRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use GetCronJobRequest() / GetCronJobRequest.new instead')
  static GetCronJobRequest create() => GetCronJobRequest._();
  static $pb.GeneratedMessage $_createMessage() => GetCronJobRequest._();
  @$core.override
  GetCronJobRequest createEmptyInstance() => GetCronJobRequest._();
  @$core.pragma('dart2js:noInline')
  static GetCronJobRequest getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<GetCronJobRequest>(
          GetCronJobRequest.$_createMessage);
  static GetCronJobRequest? _defaultInstance;

  /// @gotags: json:"name,omitempty" yaml:"name,omitempty"
  @$pb.TagNumber(1)
  $core.String get name => $_getSZ(0);
  @$pb.TagNumber(1)
  set name($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasName() => $_has(0);
  @$pb.TagNumber(1)
  void clearName() => $_clearField(1);
}

class CreateCronJobRequest extends $pb.GeneratedMessage {
  factory CreateCronJobRequest({
    CronJob? cronJob,
  }) {
    final result = CreateCronJobRequest._();
    if (cronJob != null) result.cronJob = cronJob;
    return result;
  }

  CreateCronJobRequest._();

  factory CreateCronJobRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateCronJobRequest()..mergeFromBuffer(data, registry);
  factory CreateCronJobRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateCronJobRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CreateCronJobRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: CreateCronJobRequest.$_createMessage)
    ..aOM<CronJob>(1, _omitFieldNames ? '' : 'cronJob',
        subBuilder: CronJob.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateCronJobRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateCronJobRequest copyWith(void Function(CreateCronJobRequest) updates) =>
      super.copyWith((message) => updates(message as CreateCronJobRequest))
          as CreateCronJobRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use CreateCronJobRequest() / CreateCronJobRequest.new instead')
  static CreateCronJobRequest create() => CreateCronJobRequest._();
  static $pb.GeneratedMessage $_createMessage() => CreateCronJobRequest._();
  @$core.override
  CreateCronJobRequest createEmptyInstance() => CreateCronJobRequest._();
  @$core.pragma('dart2js:noInline')
  static CreateCronJobRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<CreateCronJobRequest>(
          CreateCronJobRequest.$_createMessage);
  static CreateCronJobRequest? _defaultInstance;

  @$pb.TagNumber(1)
  CronJob get cronJob => $_getN(0);
  @$pb.TagNumber(1)
  set cronJob(CronJob value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasCronJob() => $_has(0);
  @$pb.TagNumber(1)
  void clearCronJob() => $_clearField(1);
  @$pb.TagNumber(1)
  CronJob ensureCronJob() => $_ensure(0);
}

class UpdateCronJobRequest extends $pb.GeneratedMessage {
  factory UpdateCronJobRequest({
    CronJob? cronJob,
  }) {
    final result = UpdateCronJobRequest._();
    if (cronJob != null) result.cronJob = cronJob;
    return result;
  }

  UpdateCronJobRequest._();

  factory UpdateCronJobRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpdateCronJobRequest()..mergeFromBuffer(data, registry);
  factory UpdateCronJobRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpdateCronJobRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'UpdateCronJobRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: UpdateCronJobRequest.$_createMessage)
    ..aOM<CronJob>(1, _omitFieldNames ? '' : 'cronJob',
        subBuilder: CronJob.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateCronJobRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateCronJobRequest copyWith(void Function(UpdateCronJobRequest) updates) =>
      super.copyWith((message) => updates(message as UpdateCronJobRequest))
          as UpdateCronJobRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use UpdateCronJobRequest() / UpdateCronJobRequest.new instead')
  static UpdateCronJobRequest create() => UpdateCronJobRequest._();
  static $pb.GeneratedMessage $_createMessage() => UpdateCronJobRequest._();
  @$core.override
  UpdateCronJobRequest createEmptyInstance() => UpdateCronJobRequest._();
  @$core.pragma('dart2js:noInline')
  static UpdateCronJobRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<UpdateCronJobRequest>(
          UpdateCronJobRequest.$_createMessage);
  static UpdateCronJobRequest? _defaultInstance;

  @$pb.TagNumber(1)
  CronJob get cronJob => $_getN(0);
  @$pb.TagNumber(1)
  set cronJob(CronJob value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasCronJob() => $_has(0);
  @$pb.TagNumber(1)
  void clearCronJob() => $_clearField(1);
  @$pb.TagNumber(1)
  CronJob ensureCronJob() => $_ensure(0);
}

class DeleteCronJobRequest extends $pb.GeneratedMessage {
  factory DeleteCronJobRequest({
    $core.String? name,
  }) {
    final result = DeleteCronJobRequest._();
    if (name != null) result.name = name;
    return result;
  }

  DeleteCronJobRequest._();

  factory DeleteCronJobRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DeleteCronJobRequest()..mergeFromBuffer(data, registry);
  factory DeleteCronJobRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DeleteCronJobRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'DeleteCronJobRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: DeleteCronJobRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'name')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteCronJobRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteCronJobRequest copyWith(void Function(DeleteCronJobRequest) updates) =>
      super.copyWith((message) => updates(message as DeleteCronJobRequest))
          as DeleteCronJobRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use DeleteCronJobRequest() / DeleteCronJobRequest.new instead')
  static DeleteCronJobRequest create() => DeleteCronJobRequest._();
  static $pb.GeneratedMessage $_createMessage() => DeleteCronJobRequest._();
  @$core.override
  DeleteCronJobRequest createEmptyInstance() => DeleteCronJobRequest._();
  @$core.pragma('dart2js:noInline')
  static DeleteCronJobRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<DeleteCronJobRequest>(
          DeleteCronJobRequest.$_createMessage);
  static DeleteCronJobRequest? _defaultInstance;

  /// @gotags: json:"name,omitempty" yaml:"name,omitempty"
  @$pb.TagNumber(1)
  $core.String get name => $_getSZ(0);
  @$pb.TagNumber(1)
  set name($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasName() => $_has(0);
  @$pb.TagNumber(1)
  void clearName() => $_clearField(1);
}

class GetCronJobResponse extends $pb.GeneratedMessage {
  factory GetCronJobResponse({
    CronJob? cronJob,
  }) {
    final result = GetCronJobResponse._();
    if (cronJob != null) result.cronJob = cronJob;
    return result;
  }

  GetCronJobResponse._();

  factory GetCronJobResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetCronJobResponse()..mergeFromBuffer(data, registry);
  factory GetCronJobResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetCronJobResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetCronJobResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: GetCronJobResponse.$_createMessage)
    ..aOM<CronJob>(1, _omitFieldNames ? '' : 'cronJob',
        subBuilder: CronJob.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetCronJobResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetCronJobResponse copyWith(void Function(GetCronJobResponse) updates) =>
      super.copyWith((message) => updates(message as GetCronJobResponse))
          as GetCronJobResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use GetCronJobResponse() / GetCronJobResponse.new instead')
  static GetCronJobResponse create() => GetCronJobResponse._();
  static $pb.GeneratedMessage $_createMessage() => GetCronJobResponse._();
  @$core.override
  GetCronJobResponse createEmptyInstance() => GetCronJobResponse._();
  @$core.pragma('dart2js:noInline')
  static GetCronJobResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<GetCronJobResponse>(
          GetCronJobResponse.$_createMessage);
  static GetCronJobResponse? _defaultInstance;

  @$pb.TagNumber(1)
  CronJob get cronJob => $_getN(0);
  @$pb.TagNumber(1)
  set cronJob(CronJob value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasCronJob() => $_has(0);
  @$pb.TagNumber(1)
  void clearCronJob() => $_clearField(1);
  @$pb.TagNumber(1)
  CronJob ensureCronJob() => $_ensure(0);
}

class CreateCronJobResponse extends $pb.GeneratedMessage {
  factory CreateCronJobResponse({
    CronJob? cronJob,
  }) {
    final result = CreateCronJobResponse._();
    if (cronJob != null) result.cronJob = cronJob;
    return result;
  }

  CreateCronJobResponse._();

  factory CreateCronJobResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateCronJobResponse()..mergeFromBuffer(data, registry);
  factory CreateCronJobResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateCronJobResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CreateCronJobResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: CreateCronJobResponse.$_createMessage)
    ..aOM<CronJob>(1, _omitFieldNames ? '' : 'cronJob',
        subBuilder: CronJob.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateCronJobResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateCronJobResponse copyWith(
          void Function(CreateCronJobResponse) updates) =>
      super.copyWith((message) => updates(message as CreateCronJobResponse))
          as CreateCronJobResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use CreateCronJobResponse() / CreateCronJobResponse.new instead')
  static CreateCronJobResponse create() => CreateCronJobResponse._();
  static $pb.GeneratedMessage $_createMessage() => CreateCronJobResponse._();
  @$core.override
  CreateCronJobResponse createEmptyInstance() => CreateCronJobResponse._();
  @$core.pragma('dart2js:noInline')
  static CreateCronJobResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<CreateCronJobResponse>(
          CreateCronJobResponse.$_createMessage);
  static CreateCronJobResponse? _defaultInstance;

  @$pb.TagNumber(1)
  CronJob get cronJob => $_getN(0);
  @$pb.TagNumber(1)
  set cronJob(CronJob value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasCronJob() => $_has(0);
  @$pb.TagNumber(1)
  void clearCronJob() => $_clearField(1);
  @$pb.TagNumber(1)
  CronJob ensureCronJob() => $_ensure(0);
}

class UpdateCronJobResponse extends $pb.GeneratedMessage {
  factory UpdateCronJobResponse({
    CronJob? cronJob,
  }) {
    final result = UpdateCronJobResponse._();
    if (cronJob != null) result.cronJob = cronJob;
    return result;
  }

  UpdateCronJobResponse._();

  factory UpdateCronJobResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpdateCronJobResponse()..mergeFromBuffer(data, registry);
  factory UpdateCronJobResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpdateCronJobResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'UpdateCronJobResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: UpdateCronJobResponse.$_createMessage)
    ..aOM<CronJob>(1, _omitFieldNames ? '' : 'cronJob',
        subBuilder: CronJob.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateCronJobResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateCronJobResponse copyWith(
          void Function(UpdateCronJobResponse) updates) =>
      super.copyWith((message) => updates(message as UpdateCronJobResponse))
          as UpdateCronJobResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use UpdateCronJobResponse() / UpdateCronJobResponse.new instead')
  static UpdateCronJobResponse create() => UpdateCronJobResponse._();
  static $pb.GeneratedMessage $_createMessage() => UpdateCronJobResponse._();
  @$core.override
  UpdateCronJobResponse createEmptyInstance() => UpdateCronJobResponse._();
  @$core.pragma('dart2js:noInline')
  static UpdateCronJobResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<UpdateCronJobResponse>(
          UpdateCronJobResponse.$_createMessage);
  static UpdateCronJobResponse? _defaultInstance;

  @$pb.TagNumber(1)
  CronJob get cronJob => $_getN(0);
  @$pb.TagNumber(1)
  set cronJob(CronJob value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasCronJob() => $_has(0);
  @$pb.TagNumber(1)
  void clearCronJob() => $_clearField(1);
  @$pb.TagNumber(1)
  CronJob ensureCronJob() => $_ensure(0);
}

class DeleteCronJobResponse extends $pb.GeneratedMessage {
  factory DeleteCronJobResponse({
    CronJob? cronJob,
  }) {
    final result = DeleteCronJobResponse._();
    if (cronJob != null) result.cronJob = cronJob;
    return result;
  }

  DeleteCronJobResponse._();

  factory DeleteCronJobResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DeleteCronJobResponse()..mergeFromBuffer(data, registry);
  factory DeleteCronJobResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DeleteCronJobResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'DeleteCronJobResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: DeleteCronJobResponse.$_createMessage)
    ..aOM<CronJob>(1, _omitFieldNames ? '' : 'cronJob',
        subBuilder: CronJob.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteCronJobResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteCronJobResponse copyWith(
          void Function(DeleteCronJobResponse) updates) =>
      super.copyWith((message) => updates(message as DeleteCronJobResponse))
          as DeleteCronJobResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use DeleteCronJobResponse() / DeleteCronJobResponse.new instead')
  static DeleteCronJobResponse create() => DeleteCronJobResponse._();
  static $pb.GeneratedMessage $_createMessage() => DeleteCronJobResponse._();
  @$core.override
  DeleteCronJobResponse createEmptyInstance() => DeleteCronJobResponse._();
  @$core.pragma('dart2js:noInline')
  static DeleteCronJobResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<DeleteCronJobResponse>(
          DeleteCronJobResponse.$_createMessage);
  static DeleteCronJobResponse? _defaultInstance;

  @$pb.TagNumber(1)
  CronJob get cronJob => $_getN(0);
  @$pb.TagNumber(1)
  set cronJob(CronJob value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasCronJob() => $_has(0);
  @$pb.TagNumber(1)
  void clearCronJob() => $_clearField(1);
  @$pb.TagNumber(1)
  CronJob ensureCronJob() => $_ensure(0);
}

class ListCronExecutionsRequest extends $pb.GeneratedMessage {
  factory ListCronExecutionsRequest({
    $core.String? jobName,
    $core.int? pageSize,
    $core.String? pageToken,
  }) {
    final result = ListCronExecutionsRequest._();
    if (jobName != null) result.jobName = jobName;
    if (pageSize != null) result.pageSize = pageSize;
    if (pageToken != null) result.pageToken = pageToken;
    return result;
  }

  ListCronExecutionsRequest._();

  factory ListCronExecutionsRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListCronExecutionsRequest()..mergeFromBuffer(data, registry);
  factory ListCronExecutionsRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListCronExecutionsRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListCronExecutionsRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: ListCronExecutionsRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'jobName')
    ..aI(2, _omitFieldNames ? '' : 'pageSize')
    ..aOS(3, _omitFieldNames ? '' : 'pageToken')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListCronExecutionsRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListCronExecutionsRequest copyWith(
          void Function(ListCronExecutionsRequest) updates) =>
      super.copyWith((message) => updates(message as ListCronExecutionsRequest))
          as ListCronExecutionsRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use ListCronExecutionsRequest() / ListCronExecutionsRequest.new instead')
  static ListCronExecutionsRequest create() => ListCronExecutionsRequest._();
  static $pb.GeneratedMessage $_createMessage() =>
      ListCronExecutionsRequest._();
  @$core.override
  ListCronExecutionsRequest createEmptyInstance() =>
      ListCronExecutionsRequest._();
  @$core.pragma('dart2js:noInline')
  static ListCronExecutionsRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ListCronExecutionsRequest>(
          ListCronExecutionsRequest.$_createMessage);
  static ListCronExecutionsRequest? _defaultInstance;

  /// Optional: filter by job name. Empty returns all.
  /// @gotags: json:"job_name,omitempty" yaml:"job_name,omitempty"
  @$pb.TagNumber(1)
  $core.String get jobName => $_getSZ(0);
  @$pb.TagNumber(1)
  set jobName($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasJobName() => $_has(0);
  @$pb.TagNumber(1)
  void clearJobName() => $_clearField(1);

  /// Maximum number of records to return. Defaults to 20.
  /// @gotags: json:"page_size,omitempty" yaml:"page_size,omitempty"
  @$pb.TagNumber(2)
  $core.int get pageSize => $_getIZ(1);
  @$pb.TagNumber(2)
  set pageSize($core.int value) => $_setSignedInt32(1, value);
  @$pb.TagNumber(2)
  $core.bool hasPageSize() => $_has(1);
  @$pb.TagNumber(2)
  void clearPageSize() => $_clearField(2);

  /// Pagination token from a previous response.
  /// @gotags: json:"page_token,omitempty" yaml:"page_token,omitempty"
  @$pb.TagNumber(3)
  $core.String get pageToken => $_getSZ(2);
  @$pb.TagNumber(3)
  set pageToken($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasPageToken() => $_has(2);
  @$pb.TagNumber(3)
  void clearPageToken() => $_clearField(3);
}

class RunCronJobNowRequest extends $pb.GeneratedMessage {
  factory RunCronJobNowRequest({
    $core.String? name,
  }) {
    final result = RunCronJobNowRequest._();
    if (name != null) result.name = name;
    return result;
  }

  RunCronJobNowRequest._();

  factory RunCronJobNowRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      RunCronJobNowRequest()..mergeFromBuffer(data, registry);
  factory RunCronJobNowRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      RunCronJobNowRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'RunCronJobNowRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: RunCronJobNowRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'name')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  RunCronJobNowRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  RunCronJobNowRequest copyWith(void Function(RunCronJobNowRequest) updates) =>
      super.copyWith((message) => updates(message as RunCronJobNowRequest))
          as RunCronJobNowRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use RunCronJobNowRequest() / RunCronJobNowRequest.new instead')
  static RunCronJobNowRequest create() => RunCronJobNowRequest._();
  static $pb.GeneratedMessage $_createMessage() => RunCronJobNowRequest._();
  @$core.override
  RunCronJobNowRequest createEmptyInstance() => RunCronJobNowRequest._();
  @$core.pragma('dart2js:noInline')
  static RunCronJobNowRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<RunCronJobNowRequest>(
          RunCronJobNowRequest.$_createMessage);
  static RunCronJobNowRequest? _defaultInstance;

  /// @gotags: json:"name,omitempty" yaml:"name,omitempty"
  @$pb.TagNumber(1)
  $core.String get name => $_getSZ(0);
  @$pb.TagNumber(1)
  set name($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasName() => $_has(0);
  @$pb.TagNumber(1)
  void clearName() => $_clearField(1);
}

class RunCronJobNowResponse extends $pb.GeneratedMessage {
  factory RunCronJobNowResponse({
    CronExecution? execution,
  }) {
    final result = RunCronJobNowResponse._();
    if (execution != null) result.execution = execution;
    return result;
  }

  RunCronJobNowResponse._();

  factory RunCronJobNowResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      RunCronJobNowResponse()..mergeFromBuffer(data, registry);
  factory RunCronJobNowResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      RunCronJobNowResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'RunCronJobNowResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: RunCronJobNowResponse.$_createMessage)
    ..aOM<CronExecution>(1, _omitFieldNames ? '' : 'execution',
        subBuilder: CronExecution.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  RunCronJobNowResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  RunCronJobNowResponse copyWith(
          void Function(RunCronJobNowResponse) updates) =>
      super.copyWith((message) => updates(message as RunCronJobNowResponse))
          as RunCronJobNowResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use RunCronJobNowResponse() / RunCronJobNowResponse.new instead')
  static RunCronJobNowResponse create() => RunCronJobNowResponse._();
  static $pb.GeneratedMessage $_createMessage() => RunCronJobNowResponse._();
  @$core.override
  RunCronJobNowResponse createEmptyInstance() => RunCronJobNowResponse._();
  @$core.pragma('dart2js:noInline')
  static RunCronJobNowResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<RunCronJobNowResponse>(
          RunCronJobNowResponse.$_createMessage);
  static RunCronJobNowResponse? _defaultInstance;

  @$pb.TagNumber(1)
  CronExecution get execution => $_getN(0);
  @$pb.TagNumber(1)
  set execution(CronExecution value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasExecution() => $_has(0);
  @$pb.TagNumber(1)
  void clearExecution() => $_clearField(1);
  @$pb.TagNumber(1)
  CronExecution ensureExecution() => $_ensure(0);
}

class ListCronExecutionsResponse extends $pb.GeneratedMessage {
  factory ListCronExecutionsResponse({
    $core.Iterable<CronExecution>? executions,
    $core.String? nextPageToken,
  }) {
    final result = ListCronExecutionsResponse._();
    if (executions != null) result.executions.addAll(executions);
    if (nextPageToken != null) result.nextPageToken = nextPageToken;
    return result;
  }

  ListCronExecutionsResponse._();

  factory ListCronExecutionsResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListCronExecutionsResponse()..mergeFromBuffer(data, registry);
  factory ListCronExecutionsResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListCronExecutionsResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListCronExecutionsResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: ListCronExecutionsResponse.$_createMessage)
    ..pPM<CronExecution>(1, _omitFieldNames ? '' : 'executions',
        subBuilder: CronExecution.$_createMessage)
    ..aOS(2, _omitFieldNames ? '' : 'nextPageToken')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListCronExecutionsResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListCronExecutionsResponse copyWith(
          void Function(ListCronExecutionsResponse) updates) =>
      super.copyWith(
              (message) => updates(message as ListCronExecutionsResponse))
          as ListCronExecutionsResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use ListCronExecutionsResponse() / ListCronExecutionsResponse.new instead')
  static ListCronExecutionsResponse create() => ListCronExecutionsResponse._();
  static $pb.GeneratedMessage $_createMessage() =>
      ListCronExecutionsResponse._();
  @$core.override
  ListCronExecutionsResponse createEmptyInstance() =>
      ListCronExecutionsResponse._();
  @$core.pragma('dart2js:noInline')
  static ListCronExecutionsResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ListCronExecutionsResponse>(
          ListCronExecutionsResponse.$_createMessage);
  static ListCronExecutionsResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $pb.PbList<CronExecution> get executions => $_getList(0);

  /// Token for the next page. Empty if no more results.
  /// @gotags: json:"next_page_token,omitempty" yaml:"next_page_token,omitempty"
  @$pb.TagNumber(2)
  $core.String get nextPageToken => $_getSZ(1);
  @$pb.TagNumber(2)
  set nextPageToken($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasNextPageToken() => $_has(1);
  @$pb.TagNumber(2)
  void clearNextPageToken() => $_clearField(2);
}

/// CronJobService manages cron jobs and their execution history.
class CronJobServiceApi {
  final $pb.RpcClient _client;

  CronJobServiceApi(this._client);

  /// ListCronJobs returns all cron jobs.
  $async.Future<ListCronJobsResponse> listCronJobs(
          $pb.ClientContext? ctx, ListCronJobsRequest request) =>
      _client.invoke<ListCronJobsResponse>(ctx, 'CronJobService',
          'ListCronJobs', request, ListCronJobsResponse());

  /// GetCronJob returns a single cron job by name.
  $async.Future<GetCronJobResponse> getCronJob(
          $pb.ClientContext? ctx, GetCronJobRequest request) =>
      _client.invoke<GetCronJobResponse>(
          ctx, 'CronJobService', 'GetCronJob', request, GetCronJobResponse());

  /// CreateCronJob creates a new cron job and schedules it.
  $async.Future<CreateCronJobResponse> createCronJob(
          $pb.ClientContext? ctx, CreateCronJobRequest request) =>
      _client.invoke<CreateCronJobResponse>(ctx, 'CronJobService',
          'CreateCronJob', request, CreateCronJobResponse());

  /// UpdateCronJob updates an existing cron job and reschedules it.
  $async.Future<UpdateCronJobResponse> updateCronJob(
          $pb.ClientContext? ctx, UpdateCronJobRequest request) =>
      _client.invoke<UpdateCronJobResponse>(ctx, 'CronJobService',
          'UpdateCronJob', request, UpdateCronJobResponse());

  /// DeleteCronJob removes a cron job and unschedules it.
  $async.Future<DeleteCronJobResponse> deleteCronJob(
          $pb.ClientContext? ctx, DeleteCronJobRequest request) =>
      _client.invoke<DeleteCronJobResponse>(ctx, 'CronJobService',
          'DeleteCronJob', request, DeleteCronJobResponse());

  /// ListCronExecutions returns execution records, optionally filtered by job name.
  $async.Future<ListCronExecutionsResponse> listCronExecutions(
          $pb.ClientContext? ctx, ListCronExecutionsRequest request) =>
      _client.invoke<ListCronExecutionsResponse>(ctx, 'CronJobService',
          'ListCronExecutions', request, ListCronExecutionsResponse());

  /// RunCronJobNow triggers an immediate execution of the cron job, bypassing
  /// the schedule. The execution is recorded the same way scheduled runs are.
  $async.Future<RunCronJobNowResponse> runCronJobNow(
          $pb.ClientContext? ctx, RunCronJobNowRequest request) =>
      _client.invoke<RunCronJobNowResponse>(ctx, 'CronJobService',
          'RunCronJobNow', request, RunCronJobNowResponse());
}

const $core.bool _omitFieldNames =
    $core.bool.fromEnvironment('protobuf.omit_field_names');
const $core.bool _omitMessageNames =
    $core.bool.fromEnvironment('protobuf.omit_message_names');
