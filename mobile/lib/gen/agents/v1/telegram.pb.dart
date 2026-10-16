// This is a generated file - do not edit.
//
// Generated from agents/v1/telegram.proto.

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
import 'package:protobuf/well_known_types/google/protobuf/timestamp.pb.dart'
    as $0;

import 'telegram.pbenum.dart';

export 'package:protobuf/protobuf.dart' show GeneratedMessageGenericExtensions;

export 'telegram.pbenum.dart';

/// TelegramBotCapabilities mirrors the subset of Telegram `getMe` flags that
/// change what a Channel can actually do in groups and Topics.
class TelegramBotCapabilities extends $pb.GeneratedMessage {
  factory TelegramBotCapabilities({
    $core.bool? canJoinGroups,
    $core.bool? canReadAllGroupMessages,
    $core.bool? supportsInlineQueries,
  }) {
    final result = TelegramBotCapabilities._();
    if (canJoinGroups != null) result.canJoinGroups = canJoinGroups;
    if (canReadAllGroupMessages != null)
      result.canReadAllGroupMessages = canReadAllGroupMessages;
    if (supportsInlineQueries != null)
      result.supportsInlineQueries = supportsInlineQueries;
    return result;
  }

  TelegramBotCapabilities._();

  factory TelegramBotCapabilities.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      TelegramBotCapabilities()..mergeFromBuffer(data, registry);
  factory TelegramBotCapabilities.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      TelegramBotCapabilities()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'TelegramBotCapabilities',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: TelegramBotCapabilities.$_createMessage)
    ..aOB(1, _omitFieldNames ? '' : 'canJoinGroups')
    ..aOB(2, _omitFieldNames ? '' : 'canReadAllGroupMessages')
    ..aOB(3, _omitFieldNames ? '' : 'supportsInlineQueries')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  TelegramBotCapabilities clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  TelegramBotCapabilities copyWith(
          void Function(TelegramBotCapabilities) updates) =>
      super.copyWith((message) => updates(message as TelegramBotCapabilities))
          as TelegramBotCapabilities;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use TelegramBotCapabilities() / TelegramBotCapabilities.new instead')
  static TelegramBotCapabilities create() => TelegramBotCapabilities._();
  static $pb.GeneratedMessage $_createMessage() => TelegramBotCapabilities._();
  @$core.override
  TelegramBotCapabilities createEmptyInstance() => TelegramBotCapabilities._();
  @$core.pragma('dart2js:noInline')
  static TelegramBotCapabilities getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<TelegramBotCapabilities>(
          TelegramBotCapabilities.$_createMessage);
  static TelegramBotCapabilities? _defaultInstance;

  /// @gotags: json:"can_join_groups,omitempty" bson:"can_join_groups,omitempty"
  @$pb.TagNumber(1)
  $core.bool get canJoinGroups => $_getBF(0);
  @$pb.TagNumber(1)
  set canJoinGroups($core.bool value) => $_setBool(0, value);
  @$pb.TagNumber(1)
  $core.bool hasCanJoinGroups() => $_has(0);
  @$pb.TagNumber(1)
  void clearCanJoinGroups() => $_clearField(1);

  /// False when BotFather Group Privacy is enabled, in which case ALL trigger
  /// mode only sees commands and replies. Surfaced, never changed by Butter.
  /// @gotags: json:"can_read_all_group_messages,omitempty" bson:"can_read_all_group_messages,omitempty"
  @$pb.TagNumber(2)
  $core.bool get canReadAllGroupMessages => $_getBF(1);
  @$pb.TagNumber(2)
  set canReadAllGroupMessages($core.bool value) => $_setBool(1, value);
  @$pb.TagNumber(2)
  $core.bool hasCanReadAllGroupMessages() => $_has(1);
  @$pb.TagNumber(2)
  void clearCanReadAllGroupMessages() => $_clearField(2);

  /// @gotags: json:"supports_inline_queries,omitempty" bson:"supports_inline_queries,omitempty"
  @$pb.TagNumber(3)
  $core.bool get supportsInlineQueries => $_getBF(2);
  @$pb.TagNumber(3)
  set supportsInlineQueries($core.bool value) => $_setBool(2, value);
  @$pb.TagNumber(3)
  $core.bool hasSupportsInlineQueries() => $_has(2);
  @$pb.TagNumber(3)
  void clearSupportsInlineQueries() => $_clearField(3);
}

/// TelegramChannel is one Telegram Bot transport in a workspace.
class TelegramChannel extends $pb.GeneratedMessage {
  factory TelegramChannel({
    $core.String? id,
    $core.String? key,
    $core.String? name,
    $core.String? botId,
    $core.String? botUsername,
    TelegramBotCapabilities? botCapabilities,
    TelegramReceiveMode? receiveMode,
    $core.bool? inboundEnabled,
    $core.bool? outboundEnabled,
    $fixnum.Int64? revision,
    $0.Timestamp? createdAt,
    $0.Timestamp? updatedAt,
    TelegramCredentialState? credentialState,
    $0.Timestamp? credentialUpdatedAt,
    $core.bool? webhookSecretSet,
    $core.String? lastCredentialError,
    $core.String? workspaceId,
  }) {
    final result = TelegramChannel._();
    if (id != null) result.id = id;
    if (key != null) result.key = key;
    if (name != null) result.name = name;
    if (botId != null) result.botId = botId;
    if (botUsername != null) result.botUsername = botUsername;
    if (botCapabilities != null) result.botCapabilities = botCapabilities;
    if (receiveMode != null) result.receiveMode = receiveMode;
    if (inboundEnabled != null) result.inboundEnabled = inboundEnabled;
    if (outboundEnabled != null) result.outboundEnabled = outboundEnabled;
    if (revision != null) result.revision = revision;
    if (createdAt != null) result.createdAt = createdAt;
    if (updatedAt != null) result.updatedAt = updatedAt;
    if (credentialState != null) result.credentialState = credentialState;
    if (credentialUpdatedAt != null)
      result.credentialUpdatedAt = credentialUpdatedAt;
    if (webhookSecretSet != null) result.webhookSecretSet = webhookSecretSet;
    if (lastCredentialError != null)
      result.lastCredentialError = lastCredentialError;
    if (workspaceId != null) result.workspaceId = workspaceId;
    return result;
  }

  TelegramChannel._();

  factory TelegramChannel.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      TelegramChannel()..mergeFromBuffer(data, registry);
  factory TelegramChannel.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      TelegramChannel()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'TelegramChannel',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: TelegramChannel.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..aOS(2, _omitFieldNames ? '' : 'key')
    ..aOS(3, _omitFieldNames ? '' : 'name')
    ..aOS(4, _omitFieldNames ? '' : 'botId')
    ..aOS(5, _omitFieldNames ? '' : 'botUsername')
    ..aOM<TelegramBotCapabilities>(6, _omitFieldNames ? '' : 'botCapabilities',
        subBuilder: TelegramBotCapabilities.$_createMessage)
    ..aE<TelegramReceiveMode>(7, _omitFieldNames ? '' : 'receiveMode',
        enumValues: TelegramReceiveMode.values)
    ..aOB(8, _omitFieldNames ? '' : 'inboundEnabled')
    ..aOB(9, _omitFieldNames ? '' : 'outboundEnabled')
    ..aInt64(10, _omitFieldNames ? '' : 'revision')
    ..aOM<$0.Timestamp>(11, _omitFieldNames ? '' : 'createdAt',
        subBuilder: $0.Timestamp.$_createMessage)
    ..aOM<$0.Timestamp>(12, _omitFieldNames ? '' : 'updatedAt',
        subBuilder: $0.Timestamp.$_createMessage)
    ..aE<TelegramCredentialState>(13, _omitFieldNames ? '' : 'credentialState',
        enumValues: TelegramCredentialState.values)
    ..aOM<$0.Timestamp>(14, _omitFieldNames ? '' : 'credentialUpdatedAt',
        subBuilder: $0.Timestamp.$_createMessage)
    ..aOB(15, _omitFieldNames ? '' : 'webhookSecretSet')
    ..aOS(16, _omitFieldNames ? '' : 'lastCredentialError')
    ..aOS(100, _omitFieldNames ? '' : 'workspaceId')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  TelegramChannel clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  TelegramChannel copyWith(void Function(TelegramChannel) updates) =>
      super.copyWith((message) => updates(message as TelegramChannel))
          as TelegramChannel;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use TelegramChannel() / TelegramChannel.new instead')
  static TelegramChannel create() => TelegramChannel._();
  static $pb.GeneratedMessage $_createMessage() => TelegramChannel._();
  @$core.override
  TelegramChannel createEmptyInstance() => TelegramChannel._();
  @$core.pragma('dart2js:noInline')
  static TelegramChannel getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<TelegramChannel>(
          TelegramChannel.$_createMessage);
  static TelegramChannel? _defaultInstance;

  /// Service-generated immutable identifier.
  /// @gotags: json:"id,omitempty" bson:"id,omitempty"
  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);

  /// Workspace-unique immutable key chosen by the operator. Stable handle for
  /// configuration and logs.
  /// @gotags: json:"key,omitempty" bson:"key,omitempty"
  @$pb.TagNumber(2)
  $core.String get key => $_getSZ(1);
  @$pb.TagNumber(2)
  set key($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasKey() => $_has(1);
  @$pb.TagNumber(2)
  void clearKey() => $_clearField(2);

  /// Mutable display name.
  /// @gotags: json:"name,omitempty" bson:"name,omitempty"
  @$pb.TagNumber(3)
  $core.String get name => $_getSZ(2);
  @$pb.TagNumber(3)
  set name($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasName() => $_has(2);
  @$pb.TagNumber(3)
  void clearName() => $_clearField(3);

  /// Immutable Telegram Bot ID as a decimal string, resolved by `getMe` at
  /// creation. Credential rotation may never change it.
  /// @gotags: json:"bot_id,omitempty" bson:"bot_id,omitempty"
  @$pb.TagNumber(4)
  $core.String get botId => $_getSZ(3);
  @$pb.TagNumber(4)
  set botId($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasBotId() => $_has(3);
  @$pb.TagNumber(4)
  void clearBotId() => $_clearField(4);

  /// Bot username without the leading @, refreshed on credential validation.
  /// @gotags: json:"bot_username,omitempty" bson:"bot_username,omitempty"
  @$pb.TagNumber(5)
  $core.String get botUsername => $_getSZ(4);
  @$pb.TagNumber(5)
  set botUsername($core.String value) => $_setString(4, value);
  @$pb.TagNumber(5)
  $core.bool hasBotUsername() => $_has(4);
  @$pb.TagNumber(5)
  void clearBotUsername() => $_clearField(5);

  /// @gotags: json:"bot_capabilities,omitempty" bson:"bot_capabilities,omitempty"
  @$pb.TagNumber(6)
  TelegramBotCapabilities get botCapabilities => $_getN(5);
  @$pb.TagNumber(6)
  set botCapabilities(TelegramBotCapabilities value) => $_setField(6, value);
  @$pb.TagNumber(6)
  $core.bool hasBotCapabilities() => $_has(5);
  @$pb.TagNumber(6)
  void clearBotCapabilities() => $_clearField(6);
  @$pb.TagNumber(6)
  TelegramBotCapabilities ensureBotCapabilities() => $_ensure(5);

  /// Defaults to WEBHOOK when unspecified on create.
  /// @gotags: json:"receive_mode,omitempty" bson:"receive_mode,omitempty"
  @$pb.TagNumber(7)
  TelegramReceiveMode get receiveMode => $_getN(6);
  @$pb.TagNumber(7)
  set receiveMode(TelegramReceiveMode value) => $_setField(7, value);
  @$pb.TagNumber(7)
  $core.bool hasReceiveMode() => $_has(6);
  @$pb.TagNumber(7)
  void clearReceiveMode() => $_clearField(7);

  /// Desired inbound state: whether this Channel consumes Telegram updates.
  /// @gotags: json:"inbound_enabled,omitempty" bson:"inbound_enabled,omitempty"
  @$pb.TagNumber(8)
  $core.bool get inboundEnabled => $_getBF(7);
  @$pb.TagNumber(8)
  set inboundEnabled($core.bool value) => $_setBool(7, value);
  @$pb.TagNumber(8)
  $core.bool hasInboundEnabled() => $_has(7);
  @$pb.TagNumber(8)
  void clearInboundEnabled() => $_clearField(8);

  /// Desired outbound state: whether this Channel may send messages.
  /// Inbound-enabled + outbound-disabled is rejected: every accepted
  /// interaction must be able to produce a reply.
  /// @gotags: json:"outbound_enabled,omitempty" bson:"outbound_enabled,omitempty"
  @$pb.TagNumber(9)
  $core.bool get outboundEnabled => $_getBF(8);
  @$pb.TagNumber(9)
  set outboundEnabled($core.bool value) => $_setBool(8, value);
  @$pb.TagNumber(9)
  $core.bool hasOutboundEnabled() => $_has(8);
  @$pb.TagNumber(9)
  void clearOutboundEnabled() => $_clearField(9);

  /// Optimistic concurrency token. Updates must send the revision they read.
  /// @gotags: json:"revision,omitempty" bson:"revision,omitempty"
  @$pb.TagNumber(10)
  $fixnum.Int64 get revision => $_getI64(9);
  @$pb.TagNumber(10)
  set revision($fixnum.Int64 value) => $_setInt64(9, value);
  @$pb.TagNumber(10)
  $core.bool hasRevision() => $_has(9);
  @$pb.TagNumber(10)
  void clearRevision() => $_clearField(10);

  /// @gotags: json:"created_at,omitempty" bson:"created_at,omitempty"
  @$pb.TagNumber(11)
  $0.Timestamp get createdAt => $_getN(10);
  @$pb.TagNumber(11)
  set createdAt($0.Timestamp value) => $_setField(11, value);
  @$pb.TagNumber(11)
  $core.bool hasCreatedAt() => $_has(10);
  @$pb.TagNumber(11)
  void clearCreatedAt() => $_clearField(11);
  @$pb.TagNumber(11)
  $0.Timestamp ensureCreatedAt() => $_ensure(10);

  /// @gotags: json:"updated_at,omitempty" bson:"updated_at,omitempty"
  @$pb.TagNumber(12)
  $0.Timestamp get updatedAt => $_getN(11);
  @$pb.TagNumber(12)
  set updatedAt($0.Timestamp value) => $_setField(12, value);
  @$pb.TagNumber(12)
  $core.bool hasUpdatedAt() => $_has(11);
  @$pb.TagNumber(12)
  void clearUpdatedAt() => $_clearField(12);
  @$pb.TagNumber(12)
  $0.Timestamp ensureUpdatedAt() => $_ensure(11);

  /// Server-owned. Never carries credential material.
  /// @gotags: json:"credential_state,omitempty" bson:"credential_state,omitempty"
  @$pb.TagNumber(13)
  TelegramCredentialState get credentialState => $_getN(12);
  @$pb.TagNumber(13)
  set credentialState(TelegramCredentialState value) => $_setField(13, value);
  @$pb.TagNumber(13)
  $core.bool hasCredentialState() => $_has(12);
  @$pb.TagNumber(13)
  void clearCredentialState() => $_clearField(13);

  /// When the Bot Token was last set or rotated. Server-owned.
  /// @gotags: json:"credential_updated_at,omitempty" bson:"credential_updated_at,omitempty"
  @$pb.TagNumber(14)
  $0.Timestamp get credentialUpdatedAt => $_getN(13);
  @$pb.TagNumber(14)
  set credentialUpdatedAt($0.Timestamp value) => $_setField(14, value);
  @$pb.TagNumber(14)
  $core.bool hasCredentialUpdatedAt() => $_has(13);
  @$pb.TagNumber(14)
  void clearCredentialUpdatedAt() => $_clearField(14);
  @$pb.TagNumber(14)
  $0.Timestamp ensureCredentialUpdatedAt() => $_ensure(13);

  /// Whether a per-Channel Telegram Webhook secret is stored. Server-owned.
  /// @gotags: json:"webhook_secret_set,omitempty" bson:"webhook_secret_set,omitempty"
  @$pb.TagNumber(15)
  $core.bool get webhookSecretSet => $_getBF(14);
  @$pb.TagNumber(15)
  set webhookSecretSet($core.bool value) => $_setBool(14, value);
  @$pb.TagNumber(15)
  $core.bool hasWebhookSecretSet() => $_has(14);
  @$pb.TagNumber(15)
  void clearWebhookSecretSet() => $_clearField(15);

  /// Sanitized failure summary from the last credential validation. Empty on
  /// success; never contains credential material.
  /// @gotags: json:"last_credential_error,omitempty" bson:"last_credential_error,omitempty"
  @$pb.TagNumber(16)
  $core.String get lastCredentialError => $_getSZ(15);
  @$pb.TagNumber(16)
  set lastCredentialError($core.String value) => $_setString(15, value);
  @$pb.TagNumber(16)
  $core.bool hasLastCredentialError() => $_has(15);
  @$pb.TagNumber(16)
  void clearLastCredentialError() => $_clearField(16);

  /// @gotags: json:"workspace_id,omitempty" bson:"workspace_id,omitempty"
  @$pb.TagNumber(100)
  $core.String get workspaceId => $_getSZ(16);
  @$pb.TagNumber(100)
  set workspaceId($core.String value) => $_setString(16, value);
  @$pb.TagNumber(100)
  $core.bool hasWorkspaceId() => $_has(16);
  @$pb.TagNumber(100)
  void clearWorkspaceId() => $_clearField(100);
}

/// TelegramDestinationConfig is the inbound routing and interaction policy of
/// one Destination. Only meaningful when the Destination is inbound-enabled,
/// but always persisted so enabling inbound later needs no reconfiguration.
class TelegramDestinationConfig extends $pb.GeneratedMessage {
  factory TelegramDestinationConfig({
    $core.String? agentId,
    $core.String? model,
    $core.Iterable<$core.String>? selectableAgentIds,
    $core.Iterable<$core.String>? selectableModels,
    TelegramTriggerMode? triggerMode,
    TelegramSessionPolicy? sessionPolicy,
    $core.Iterable<$core.String>? allowedUserIds,
    $core.Iterable<$core.String>? controllerUserIds,
    TelegramReplyMode? replyMode,
    $core.bool? debugDefault,
  }) {
    final result = TelegramDestinationConfig._();
    if (agentId != null) result.agentId = agentId;
    if (model != null) result.model = model;
    if (selectableAgentIds != null)
      result.selectableAgentIds.addAll(selectableAgentIds);
    if (selectableModels != null)
      result.selectableModels.addAll(selectableModels);
    if (triggerMode != null) result.triggerMode = triggerMode;
    if (sessionPolicy != null) result.sessionPolicy = sessionPolicy;
    if (allowedUserIds != null) result.allowedUserIds.addAll(allowedUserIds);
    if (controllerUserIds != null)
      result.controllerUserIds.addAll(controllerUserIds);
    if (replyMode != null) result.replyMode = replyMode;
    if (debugDefault != null) result.debugDefault = debugDefault;
    return result;
  }

  TelegramDestinationConfig._();

  factory TelegramDestinationConfig.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      TelegramDestinationConfig()..mergeFromBuffer(data, registry);
  factory TelegramDestinationConfig.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      TelegramDestinationConfig()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'TelegramDestinationConfig',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: TelegramDestinationConfig.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'agentId')
    ..aOS(2, _omitFieldNames ? '' : 'model')
    ..pPS(3, _omitFieldNames ? '' : 'selectableAgentIds')
    ..pPS(4, _omitFieldNames ? '' : 'selectableModels')
    ..aE<TelegramTriggerMode>(5, _omitFieldNames ? '' : 'triggerMode',
        enumValues: TelegramTriggerMode.values)
    ..aE<TelegramSessionPolicy>(6, _omitFieldNames ? '' : 'sessionPolicy',
        enumValues: TelegramSessionPolicy.values)
    ..pPS(7, _omitFieldNames ? '' : 'allowedUserIds')
    ..pPS(8, _omitFieldNames ? '' : 'controllerUserIds')
    ..aE<TelegramReplyMode>(9, _omitFieldNames ? '' : 'replyMode',
        enumValues: TelegramReplyMode.values)
    ..aOB(10, _omitFieldNames ? '' : 'debugDefault')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  TelegramDestinationConfig clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  TelegramDestinationConfig copyWith(
          void Function(TelegramDestinationConfig) updates) =>
      super.copyWith((message) => updates(message as TelegramDestinationConfig))
          as TelegramDestinationConfig;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use TelegramDestinationConfig() / TelegramDestinationConfig.new instead')
  static TelegramDestinationConfig create() => TelegramDestinationConfig._();
  static $pb.GeneratedMessage $_createMessage() =>
      TelegramDestinationConfig._();
  @$core.override
  TelegramDestinationConfig createEmptyInstance() =>
      TelegramDestinationConfig._();
  @$core.pragma('dart2js:noInline')
  static TelegramDestinationConfig getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<TelegramDestinationConfig>(
          TelegramDestinationConfig.$_createMessage);
  static TelegramDestinationConfig? _defaultInstance;

  /// Default Agent ID invoked at this address. Required for inbound
  /// Destinations; must be an ACTIVE Agent in the same workspace.
  /// @gotags: json:"agent_id,omitempty" bson:"agent_id,omitempty"
  @$pb.TagNumber(1)
  $core.String get agentId => $_getSZ(0);
  @$pb.TagNumber(1)
  set agentId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasAgentId() => $_has(0);
  @$pb.TagNumber(1)
  void clearAgentId() => $_clearField(1);

  /// Optional explicit Model alias. Empty inherits the Agent's default Model.
  /// @gotags: json:"model,omitempty" bson:"model,omitempty"
  @$pb.TagNumber(2)
  $core.String get model => $_getSZ(1);
  @$pb.TagNumber(2)
  set model($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasModel() => $_has(1);
  @$pb.TagNumber(2)
  void clearModel() => $_clearField(2);

  /// Agents a controller may switch to with /agent. Empty locks Agent
  /// selection to `agent_id`.
  /// @gotags: json:"selectable_agent_ids,omitempty" bson:"selectable_agent_ids,omitempty"
  @$pb.TagNumber(3)
  $pb.PbList<$core.String> get selectableAgentIds => $_getList(2);

  /// Model aliases a controller may switch to with /model. Empty locks Model
  /// selection.
  /// @gotags: json:"selectable_models,omitempty" bson:"selectable_models,omitempty"
  @$pb.TagNumber(4)
  $pb.PbList<$core.String> get selectableModels => $_getList(3);

  /// Defaults to ALL when unspecified.
  /// @gotags: json:"trigger_mode,omitempty" bson:"trigger_mode,omitempty"
  @$pb.TagNumber(5)
  TelegramTriggerMode get triggerMode => $_getN(4);
  @$pb.TagNumber(5)
  set triggerMode(TelegramTriggerMode value) => $_setField(5, value);
  @$pb.TagNumber(5)
  $core.bool hasTriggerMode() => $_has(4);
  @$pb.TagNumber(5)
  void clearTriggerMode() => $_clearField(5);

  /// Defaults to DESTINATION when unspecified.
  /// @gotags: json:"session_policy,omitempty" bson:"session_policy,omitempty"
  @$pb.TagNumber(6)
  TelegramSessionPolicy get sessionPolicy => $_getN(5);
  @$pb.TagNumber(6)
  set sessionPolicy(TelegramSessionPolicy value) => $_setField(6, value);
  @$pb.TagNumber(6)
  $core.bool hasSessionPolicy() => $_has(5);
  @$pb.TagNumber(6)
  void clearSessionPolicy() => $_clearField(6);

  /// Telegram user IDs (decimal strings) admitted at this address. Empty
  /// admits every real Telegram user that can reach the address.
  /// @gotags: json:"allowed_user_ids,omitempty" bson:"allowed_user_ids,omitempty"
  @$pb.TagNumber(7)
  $pb.PbList<$core.String> get allowedUserIds => $_getList(6);

  /// Telegram user IDs allowed to run management commands (/agent, /model,
  /// /debug, /clear). Controllers must also satisfy ordinary admission.
  /// @gotags: json:"controller_user_ids,omitempty" bson:"controller_user_ids,omitempty"
  @$pb.TagNumber(8)
  $pb.PbList<$core.String> get controllerUserIds => $_getList(7);

  /// Defaults to REPLY when unspecified.
  /// @gotags: json:"reply_mode,omitempty" bson:"reply_mode,omitempty"
  @$pb.TagNumber(9)
  TelegramReplyMode get replyMode => $_getN(8);
  @$pb.TagNumber(9)
  set replyMode(TelegramReplyMode value) => $_setField(9, value);
  @$pb.TagNumber(9)
  $core.bool hasReplyMode() => $_has(8);
  @$pb.TagNumber(9)
  void clearReplyMode() => $_clearField(9);

  /// Whether debug output starts enabled for new sessions at this address.
  /// Unset means enabled: debug output is on unless explicitly turned off.
  /// @gotags: json:"debug_default,omitempty" bson:"debug_default,omitempty"
  @$pb.TagNumber(10)
  $core.bool get debugDefault => $_getBF(9);
  @$pb.TagNumber(10)
  set debugDefault($core.bool value) => $_setBool(9, value);
  @$pb.TagNumber(10)
  $core.bool hasDebugDefault() => $_has(9);
  @$pb.TagNumber(10)
  void clearDebugDefault() => $_clearField(10);
}

/// TelegramDestinationVerification is server-owned evidence that the address
/// actually works. It never rewrites the configured address.
class TelegramDestinationVerification extends $pb.GeneratedMessage {
  factory TelegramDestinationVerification({
    $core.bool? verified,
    $0.Timestamp? verifiedAt,
    $0.Timestamp? lastInboundAt,
    $0.Timestamp? lastOutboundAt,
    $core.String? lastOutboundError,
  }) {
    final result = TelegramDestinationVerification._();
    if (verified != null) result.verified = verified;
    if (verifiedAt != null) result.verifiedAt = verifiedAt;
    if (lastInboundAt != null) result.lastInboundAt = lastInboundAt;
    if (lastOutboundAt != null) result.lastOutboundAt = lastOutboundAt;
    if (lastOutboundError != null) result.lastOutboundError = lastOutboundError;
    return result;
  }

  TelegramDestinationVerification._();

  factory TelegramDestinationVerification.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      TelegramDestinationVerification()..mergeFromBuffer(data, registry);
  factory TelegramDestinationVerification.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      TelegramDestinationVerification()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'TelegramDestinationVerification',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: TelegramDestinationVerification.$_createMessage)
    ..aOB(1, _omitFieldNames ? '' : 'verified')
    ..aOM<$0.Timestamp>(2, _omitFieldNames ? '' : 'verifiedAt',
        subBuilder: $0.Timestamp.$_createMessage)
    ..aOM<$0.Timestamp>(3, _omitFieldNames ? '' : 'lastInboundAt',
        subBuilder: $0.Timestamp.$_createMessage)
    ..aOM<$0.Timestamp>(4, _omitFieldNames ? '' : 'lastOutboundAt',
        subBuilder: $0.Timestamp.$_createMessage)
    ..aOS(5, _omitFieldNames ? '' : 'lastOutboundError')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  TelegramDestinationVerification clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  TelegramDestinationVerification copyWith(
          void Function(TelegramDestinationVerification) updates) =>
      super.copyWith(
              (message) => updates(message as TelegramDestinationVerification))
          as TelegramDestinationVerification;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use TelegramDestinationVerification() / TelegramDestinationVerification.new instead')
  static TelegramDestinationVerification create() =>
      TelegramDestinationVerification._();
  static $pb.GeneratedMessage $_createMessage() =>
      TelegramDestinationVerification._();
  @$core.override
  TelegramDestinationVerification createEmptyInstance() =>
      TelegramDestinationVerification._();
  @$core.pragma('dart2js:noInline')
  static TelegramDestinationVerification getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<TelegramDestinationVerification>(
          TelegramDestinationVerification.$_createMessage);
  static TelegramDestinationVerification? _defaultInstance;

  /// True once a message was successfully sent to, or received from, this
  /// exact address.
  /// @gotags: json:"verified,omitempty" bson:"verified,omitempty"
  @$pb.TagNumber(1)
  $core.bool get verified => $_getBF(0);
  @$pb.TagNumber(1)
  set verified($core.bool value) => $_setBool(0, value);
  @$pb.TagNumber(1)
  $core.bool hasVerified() => $_has(0);
  @$pb.TagNumber(1)
  void clearVerified() => $_clearField(1);

  /// @gotags: json:"verified_at,omitempty" bson:"verified_at,omitempty"
  @$pb.TagNumber(2)
  $0.Timestamp get verifiedAt => $_getN(1);
  @$pb.TagNumber(2)
  set verifiedAt($0.Timestamp value) => $_setField(2, value);
  @$pb.TagNumber(2)
  $core.bool hasVerifiedAt() => $_has(1);
  @$pb.TagNumber(2)
  void clearVerifiedAt() => $_clearField(2);
  @$pb.TagNumber(2)
  $0.Timestamp ensureVerifiedAt() => $_ensure(1);

  /// @gotags: json:"last_inbound_at,omitempty" bson:"last_inbound_at,omitempty"
  @$pb.TagNumber(3)
  $0.Timestamp get lastInboundAt => $_getN(2);
  @$pb.TagNumber(3)
  set lastInboundAt($0.Timestamp value) => $_setField(3, value);
  @$pb.TagNumber(3)
  $core.bool hasLastInboundAt() => $_has(2);
  @$pb.TagNumber(3)
  void clearLastInboundAt() => $_clearField(3);
  @$pb.TagNumber(3)
  $0.Timestamp ensureLastInboundAt() => $_ensure(2);

  /// @gotags: json:"last_outbound_at,omitempty" bson:"last_outbound_at,omitempty"
  @$pb.TagNumber(4)
  $0.Timestamp get lastOutboundAt => $_getN(3);
  @$pb.TagNumber(4)
  set lastOutboundAt($0.Timestamp value) => $_setField(4, value);
  @$pb.TagNumber(4)
  $core.bool hasLastOutboundAt() => $_has(3);
  @$pb.TagNumber(4)
  void clearLastOutboundAt() => $_clearField(4);
  @$pb.TagNumber(4)
  $0.Timestamp ensureLastOutboundAt() => $_ensure(3);

  /// Sanitized summary of the most recent outbound failure. Empty on success.
  /// @gotags: json:"last_outbound_error,omitempty" bson:"last_outbound_error,omitempty"
  @$pb.TagNumber(5)
  $core.String get lastOutboundError => $_getSZ(4);
  @$pb.TagNumber(5)
  set lastOutboundError($core.String value) => $_setString(4, value);
  @$pb.TagNumber(5)
  $core.bool hasLastOutboundError() => $_has(4);
  @$pb.TagNumber(5)
  void clearLastOutboundError() => $_clearField(5);
}

/// TelegramDestination is one exact Telegram address under a Channel.
class TelegramDestination extends $pb.GeneratedMessage {
  factory TelegramDestination({
    $core.String? id,
    $core.String? key,
    $core.String? name,
    $core.String? channelId,
    $core.String? chatId,
    $core.String? messageThreadId,
    $core.bool? inboundEnabled,
    $core.bool? outboundEnabled,
    TelegramDestinationConfig? config,
    $fixnum.Int64? revision,
    $0.Timestamp? createdAt,
    $0.Timestamp? updatedAt,
    TelegramDestinationVerification? verification,
    $core.String? workspaceId,
  }) {
    final result = TelegramDestination._();
    if (id != null) result.id = id;
    if (key != null) result.key = key;
    if (name != null) result.name = name;
    if (channelId != null) result.channelId = channelId;
    if (chatId != null) result.chatId = chatId;
    if (messageThreadId != null) result.messageThreadId = messageThreadId;
    if (inboundEnabled != null) result.inboundEnabled = inboundEnabled;
    if (outboundEnabled != null) result.outboundEnabled = outboundEnabled;
    if (config != null) result.config = config;
    if (revision != null) result.revision = revision;
    if (createdAt != null) result.createdAt = createdAt;
    if (updatedAt != null) result.updatedAt = updatedAt;
    if (verification != null) result.verification = verification;
    if (workspaceId != null) result.workspaceId = workspaceId;
    return result;
  }

  TelegramDestination._();

  factory TelegramDestination.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      TelegramDestination()..mergeFromBuffer(data, registry);
  factory TelegramDestination.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      TelegramDestination()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'TelegramDestination',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: TelegramDestination.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..aOS(2, _omitFieldNames ? '' : 'key')
    ..aOS(3, _omitFieldNames ? '' : 'name')
    ..aOS(4, _omitFieldNames ? '' : 'channelId')
    ..aOS(5, _omitFieldNames ? '' : 'chatId')
    ..aOS(6, _omitFieldNames ? '' : 'messageThreadId')
    ..aOB(7, _omitFieldNames ? '' : 'inboundEnabled')
    ..aOB(8, _omitFieldNames ? '' : 'outboundEnabled')
    ..aOM<TelegramDestinationConfig>(9, _omitFieldNames ? '' : 'config',
        subBuilder: TelegramDestinationConfig.$_createMessage)
    ..aInt64(10, _omitFieldNames ? '' : 'revision')
    ..aOM<$0.Timestamp>(11, _omitFieldNames ? '' : 'createdAt',
        subBuilder: $0.Timestamp.$_createMessage)
    ..aOM<$0.Timestamp>(12, _omitFieldNames ? '' : 'updatedAt',
        subBuilder: $0.Timestamp.$_createMessage)
    ..aOM<TelegramDestinationVerification>(
        13, _omitFieldNames ? '' : 'verification',
        subBuilder: TelegramDestinationVerification.$_createMessage)
    ..aOS(100, _omitFieldNames ? '' : 'workspaceId')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  TelegramDestination clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  TelegramDestination copyWith(void Function(TelegramDestination) updates) =>
      super.copyWith((message) => updates(message as TelegramDestination))
          as TelegramDestination;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core
      .Deprecated('Use TelegramDestination() / TelegramDestination.new instead')
  static TelegramDestination create() => TelegramDestination._();
  static $pb.GeneratedMessage $_createMessage() => TelegramDestination._();
  @$core.override
  TelegramDestination createEmptyInstance() => TelegramDestination._();
  @$core.pragma('dart2js:noInline')
  static TelegramDestination getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<TelegramDestination>(
          TelegramDestination.$_createMessage);
  static TelegramDestination? _defaultInstance;

  /// Service-generated immutable identifier. This is what Cron jobs and
  /// Notify Groups persist.
  /// @gotags: json:"id,omitempty" bson:"id,omitempty"
  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);

  /// Workspace-unique immutable key chosen by the operator.
  /// @gotags: json:"key,omitempty" bson:"key,omitempty"
  @$pb.TagNumber(2)
  $core.String get key => $_getSZ(1);
  @$pb.TagNumber(2)
  set key($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasKey() => $_has(1);
  @$pb.TagNumber(2)
  void clearKey() => $_clearField(2);

  /// Mutable display name.
  /// @gotags: json:"name,omitempty" bson:"name,omitempty"
  @$pb.TagNumber(3)
  $core.String get name => $_getSZ(2);
  @$pb.TagNumber(3)
  set name($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasName() => $_has(2);
  @$pb.TagNumber(3)
  void clearName() => $_clearField(3);

  /// Immutable owning Channel.
  /// @gotags: json:"channel_id,omitempty" bson:"channel_id,omitempty"
  @$pb.TagNumber(4)
  $core.String get channelId => $_getSZ(3);
  @$pb.TagNumber(4)
  set channelId($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasChannelId() => $_has(3);
  @$pb.TagNumber(4)
  void clearChannelId() => $_clearField(4);

  /// Immutable Telegram chat ID as a canonical decimal string. Negative for
  /// groups and supergroups.
  /// @gotags: json:"chat_id,omitempty" bson:"chat_id,omitempty"
  @$pb.TagNumber(5)
  $core.String get chatId => $_getSZ(4);
  @$pb.TagNumber(5)
  set chatId($core.String value) => $_setString(4, value);
  @$pb.TagNumber(5)
  $core.bool hasChatId() => $_has(4);
  @$pb.TagNumber(5)
  void clearChatId() => $_clearField(5);

  /// Immutable optional Forum Topic ID as a canonical decimal string. Empty
  /// means "not a Topic address" and is distinct from any real thread ID; a
  /// present value must be positive.
  /// @gotags: json:"message_thread_id,omitempty" bson:"message_thread_id,omitempty"
  @$pb.TagNumber(6)
  $core.String get messageThreadId => $_getSZ(5);
  @$pb.TagNumber(6)
  set messageThreadId($core.String value) => $_setString(5, value);
  @$pb.TagNumber(6)
  $core.bool hasMessageThreadId() => $_has(5);
  @$pb.TagNumber(6)
  void clearMessageThreadId() => $_clearField(6);

  /// Whether Telegram updates at this address may invoke an Agent.
  /// @gotags: json:"inbound_enabled,omitempty" bson:"inbound_enabled,omitempty"
  @$pb.TagNumber(7)
  $core.bool get inboundEnabled => $_getBF(6);
  @$pb.TagNumber(7)
  set inboundEnabled($core.bool value) => $_setBool(6, value);
  @$pb.TagNumber(7)
  $core.bool hasInboundEnabled() => $_has(6);
  @$pb.TagNumber(7)
  void clearInboundEnabled() => $_clearField(7);

  /// Whether proactive senders (Cron, Notify Groups, test messages) may
  /// target this address.
  /// @gotags: json:"outbound_enabled,omitempty" bson:"outbound_enabled,omitempty"
  @$pb.TagNumber(8)
  $core.bool get outboundEnabled => $_getBF(7);
  @$pb.TagNumber(8)
  set outboundEnabled($core.bool value) => $_setBool(7, value);
  @$pb.TagNumber(8)
  $core.bool hasOutboundEnabled() => $_has(7);
  @$pb.TagNumber(8)
  void clearOutboundEnabled() => $_clearField(8);

  /// @gotags: json:"config,omitempty" bson:"config,omitempty"
  @$pb.TagNumber(9)
  TelegramDestinationConfig get config => $_getN(8);
  @$pb.TagNumber(9)
  set config(TelegramDestinationConfig value) => $_setField(9, value);
  @$pb.TagNumber(9)
  $core.bool hasConfig() => $_has(8);
  @$pb.TagNumber(9)
  void clearConfig() => $_clearField(9);
  @$pb.TagNumber(9)
  TelegramDestinationConfig ensureConfig() => $_ensure(8);

  /// Optimistic concurrency token. Also frozen into inbound queue events so a
  /// worker can tell that policy changed mid-flight.
  /// @gotags: json:"revision,omitempty" bson:"revision,omitempty"
  @$pb.TagNumber(10)
  $fixnum.Int64 get revision => $_getI64(9);
  @$pb.TagNumber(10)
  set revision($fixnum.Int64 value) => $_setInt64(9, value);
  @$pb.TagNumber(10)
  $core.bool hasRevision() => $_has(9);
  @$pb.TagNumber(10)
  void clearRevision() => $_clearField(10);

  /// @gotags: json:"created_at,omitempty" bson:"created_at,omitempty"
  @$pb.TagNumber(11)
  $0.Timestamp get createdAt => $_getN(10);
  @$pb.TagNumber(11)
  set createdAt($0.Timestamp value) => $_setField(11, value);
  @$pb.TagNumber(11)
  $core.bool hasCreatedAt() => $_has(10);
  @$pb.TagNumber(11)
  void clearCreatedAt() => $_clearField(11);
  @$pb.TagNumber(11)
  $0.Timestamp ensureCreatedAt() => $_ensure(10);

  /// @gotags: json:"updated_at,omitempty" bson:"updated_at,omitempty"
  @$pb.TagNumber(12)
  $0.Timestamp get updatedAt => $_getN(11);
  @$pb.TagNumber(12)
  set updatedAt($0.Timestamp value) => $_setField(12, value);
  @$pb.TagNumber(12)
  $core.bool hasUpdatedAt() => $_has(11);
  @$pb.TagNumber(12)
  void clearUpdatedAt() => $_clearField(12);
  @$pb.TagNumber(12)
  $0.Timestamp ensureUpdatedAt() => $_ensure(11);

  /// @gotags: json:"verification,omitempty" bson:"verification,omitempty"
  @$pb.TagNumber(13)
  TelegramDestinationVerification get verification => $_getN(12);
  @$pb.TagNumber(13)
  set verification(TelegramDestinationVerification value) =>
      $_setField(13, value);
  @$pb.TagNumber(13)
  $core.bool hasVerification() => $_has(12);
  @$pb.TagNumber(13)
  void clearVerification() => $_clearField(13);
  @$pb.TagNumber(13)
  TelegramDestinationVerification ensureVerification() => $_ensure(12);

  /// @gotags: json:"workspace_id,omitempty" bson:"workspace_id,omitempty"
  @$pb.TagNumber(100)
  $core.String get workspaceId => $_getSZ(13);
  @$pb.TagNumber(100)
  set workspaceId($core.String value) => $_setString(13, value);
  @$pb.TagNumber(100)
  $core.bool hasWorkspaceId() => $_has(13);
  @$pb.TagNumber(100)
  void clearWorkspaceId() => $_clearField(100);
}

class ListTelegramChannelsRequest extends $pb.GeneratedMessage {
  factory ListTelegramChannelsRequest() => ListTelegramChannelsRequest._();

  ListTelegramChannelsRequest._();

  factory ListTelegramChannelsRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListTelegramChannelsRequest()..mergeFromBuffer(data, registry);
  factory ListTelegramChannelsRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListTelegramChannelsRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListTelegramChannelsRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: ListTelegramChannelsRequest.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListTelegramChannelsRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListTelegramChannelsRequest copyWith(
          void Function(ListTelegramChannelsRequest) updates) =>
      super.copyWith(
              (message) => updates(message as ListTelegramChannelsRequest))
          as ListTelegramChannelsRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use ListTelegramChannelsRequest() / ListTelegramChannelsRequest.new instead')
  static ListTelegramChannelsRequest create() =>
      ListTelegramChannelsRequest._();
  static $pb.GeneratedMessage $_createMessage() =>
      ListTelegramChannelsRequest._();
  @$core.override
  ListTelegramChannelsRequest createEmptyInstance() =>
      ListTelegramChannelsRequest._();
  @$core.pragma('dart2js:noInline')
  static ListTelegramChannelsRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ListTelegramChannelsRequest>(
          ListTelegramChannelsRequest.$_createMessage);
  static ListTelegramChannelsRequest? _defaultInstance;
}

class ListTelegramChannelsResponse extends $pb.GeneratedMessage {
  factory ListTelegramChannelsResponse({
    $core.Iterable<TelegramChannel>? channels,
  }) {
    final result = ListTelegramChannelsResponse._();
    if (channels != null) result.channels.addAll(channels);
    return result;
  }

  ListTelegramChannelsResponse._();

  factory ListTelegramChannelsResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListTelegramChannelsResponse()..mergeFromBuffer(data, registry);
  factory ListTelegramChannelsResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListTelegramChannelsResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListTelegramChannelsResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: ListTelegramChannelsResponse.$_createMessage)
    ..pPM<TelegramChannel>(1, _omitFieldNames ? '' : 'channels',
        subBuilder: TelegramChannel.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListTelegramChannelsResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListTelegramChannelsResponse copyWith(
          void Function(ListTelegramChannelsResponse) updates) =>
      super.copyWith(
              (message) => updates(message as ListTelegramChannelsResponse))
          as ListTelegramChannelsResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use ListTelegramChannelsResponse() / ListTelegramChannelsResponse.new instead')
  static ListTelegramChannelsResponse create() =>
      ListTelegramChannelsResponse._();
  static $pb.GeneratedMessage $_createMessage() =>
      ListTelegramChannelsResponse._();
  @$core.override
  ListTelegramChannelsResponse createEmptyInstance() =>
      ListTelegramChannelsResponse._();
  @$core.pragma('dart2js:noInline')
  static ListTelegramChannelsResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ListTelegramChannelsResponse>(
          ListTelegramChannelsResponse.$_createMessage);
  static ListTelegramChannelsResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $pb.PbList<TelegramChannel> get channels => $_getList(0);
}

class GetTelegramChannelRequest extends $pb.GeneratedMessage {
  factory GetTelegramChannelRequest({
    $core.String? id,
  }) {
    final result = GetTelegramChannelRequest._();
    if (id != null) result.id = id;
    return result;
  }

  GetTelegramChannelRequest._();

  factory GetTelegramChannelRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetTelegramChannelRequest()..mergeFromBuffer(data, registry);
  factory GetTelegramChannelRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetTelegramChannelRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetTelegramChannelRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: GetTelegramChannelRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetTelegramChannelRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetTelegramChannelRequest copyWith(
          void Function(GetTelegramChannelRequest) updates) =>
      super.copyWith((message) => updates(message as GetTelegramChannelRequest))
          as GetTelegramChannelRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use GetTelegramChannelRequest() / GetTelegramChannelRequest.new instead')
  static GetTelegramChannelRequest create() => GetTelegramChannelRequest._();
  static $pb.GeneratedMessage $_createMessage() =>
      GetTelegramChannelRequest._();
  @$core.override
  GetTelegramChannelRequest createEmptyInstance() =>
      GetTelegramChannelRequest._();
  @$core.pragma('dart2js:noInline')
  static GetTelegramChannelRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<GetTelegramChannelRequest>(
          GetTelegramChannelRequest.$_createMessage);
  static GetTelegramChannelRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);
}

class GetTelegramChannelResponse extends $pb.GeneratedMessage {
  factory GetTelegramChannelResponse({
    TelegramChannel? channel,
  }) {
    final result = GetTelegramChannelResponse._();
    if (channel != null) result.channel = channel;
    return result;
  }

  GetTelegramChannelResponse._();

  factory GetTelegramChannelResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetTelegramChannelResponse()..mergeFromBuffer(data, registry);
  factory GetTelegramChannelResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetTelegramChannelResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetTelegramChannelResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: GetTelegramChannelResponse.$_createMessage)
    ..aOM<TelegramChannel>(1, _omitFieldNames ? '' : 'channel',
        subBuilder: TelegramChannel.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetTelegramChannelResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetTelegramChannelResponse copyWith(
          void Function(GetTelegramChannelResponse) updates) =>
      super.copyWith(
              (message) => updates(message as GetTelegramChannelResponse))
          as GetTelegramChannelResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use GetTelegramChannelResponse() / GetTelegramChannelResponse.new instead')
  static GetTelegramChannelResponse create() => GetTelegramChannelResponse._();
  static $pb.GeneratedMessage $_createMessage() =>
      GetTelegramChannelResponse._();
  @$core.override
  GetTelegramChannelResponse createEmptyInstance() =>
      GetTelegramChannelResponse._();
  @$core.pragma('dart2js:noInline')
  static GetTelegramChannelResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<GetTelegramChannelResponse>(
          GetTelegramChannelResponse.$_createMessage);
  static GetTelegramChannelResponse? _defaultInstance;

  @$pb.TagNumber(1)
  TelegramChannel get channel => $_getN(0);
  @$pb.TagNumber(1)
  set channel(TelegramChannel value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasChannel() => $_has(0);
  @$pb.TagNumber(1)
  void clearChannel() => $_clearField(1);
  @$pb.TagNumber(1)
  TelegramChannel ensureChannel() => $_ensure(0);
}

class CreateTelegramChannelRequest extends $pb.GeneratedMessage {
  factory CreateTelegramChannelRequest({
    TelegramChannel? channel,
    $core.String? botToken,
  }) {
    final result = CreateTelegramChannelRequest._();
    if (channel != null) result.channel = channel;
    if (botToken != null) result.botToken = botToken;
    return result;
  }

  CreateTelegramChannelRequest._();

  factory CreateTelegramChannelRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateTelegramChannelRequest()..mergeFromBuffer(data, registry);
  factory CreateTelegramChannelRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateTelegramChannelRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CreateTelegramChannelRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: CreateTelegramChannelRequest.$_createMessage)
    ..aOM<TelegramChannel>(1, _omitFieldNames ? '' : 'channel',
        subBuilder: TelegramChannel.$_createMessage)
    ..aOS(2, _omitFieldNames ? '' : 'botToken')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateTelegramChannelRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateTelegramChannelRequest copyWith(
          void Function(CreateTelegramChannelRequest) updates) =>
      super.copyWith(
              (message) => updates(message as CreateTelegramChannelRequest))
          as CreateTelegramChannelRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use CreateTelegramChannelRequest() / CreateTelegramChannelRequest.new instead')
  static CreateTelegramChannelRequest create() =>
      CreateTelegramChannelRequest._();
  static $pb.GeneratedMessage $_createMessage() =>
      CreateTelegramChannelRequest._();
  @$core.override
  CreateTelegramChannelRequest createEmptyInstance() =>
      CreateTelegramChannelRequest._();
  @$core.pragma('dart2js:noInline')
  static CreateTelegramChannelRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<CreateTelegramChannelRequest>(
          CreateTelegramChannelRequest.$_createMessage);
  static CreateTelegramChannelRequest? _defaultInstance;

  /// Only key, name, and receive_mode are read from this message. Bot
  /// identity, revision, timestamps, and credential state are server-owned.
  @$pb.TagNumber(1)
  TelegramChannel get channel => $_getN(0);
  @$pb.TagNumber(1)
  set channel(TelegramChannel value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasChannel() => $_has(0);
  @$pb.TagNumber(1)
  void clearChannel() => $_clearField(1);
  @$pb.TagNumber(1)
  TelegramChannel ensureChannel() => $_ensure(0);

  /// Bot Token, write-only. Validated with `getMe`, encrypted before storage,
  /// and never returned by any RPC or written to logs or error details.
  @$pb.TagNumber(2)
  $core.String get botToken => $_getSZ(1);
  @$pb.TagNumber(2)
  set botToken($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasBotToken() => $_has(1);
  @$pb.TagNumber(2)
  void clearBotToken() => $_clearField(2);
}

class CreateTelegramChannelResponse extends $pb.GeneratedMessage {
  factory CreateTelegramChannelResponse({
    TelegramChannel? channel,
  }) {
    final result = CreateTelegramChannelResponse._();
    if (channel != null) result.channel = channel;
    return result;
  }

  CreateTelegramChannelResponse._();

  factory CreateTelegramChannelResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateTelegramChannelResponse()..mergeFromBuffer(data, registry);
  factory CreateTelegramChannelResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateTelegramChannelResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CreateTelegramChannelResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: CreateTelegramChannelResponse.$_createMessage)
    ..aOM<TelegramChannel>(1, _omitFieldNames ? '' : 'channel',
        subBuilder: TelegramChannel.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateTelegramChannelResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateTelegramChannelResponse copyWith(
          void Function(CreateTelegramChannelResponse) updates) =>
      super.copyWith(
              (message) => updates(message as CreateTelegramChannelResponse))
          as CreateTelegramChannelResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use CreateTelegramChannelResponse() / CreateTelegramChannelResponse.new instead')
  static CreateTelegramChannelResponse create() =>
      CreateTelegramChannelResponse._();
  static $pb.GeneratedMessage $_createMessage() =>
      CreateTelegramChannelResponse._();
  @$core.override
  CreateTelegramChannelResponse createEmptyInstance() =>
      CreateTelegramChannelResponse._();
  @$core.pragma('dart2js:noInline')
  static CreateTelegramChannelResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<CreateTelegramChannelResponse>(
          CreateTelegramChannelResponse.$_createMessage);
  static CreateTelegramChannelResponse? _defaultInstance;

  @$pb.TagNumber(1)
  TelegramChannel get channel => $_getN(0);
  @$pb.TagNumber(1)
  set channel(TelegramChannel value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasChannel() => $_has(0);
  @$pb.TagNumber(1)
  void clearChannel() => $_clearField(1);
  @$pb.TagNumber(1)
  TelegramChannel ensureChannel() => $_ensure(0);
}

class UpdateTelegramChannelRequest extends $pb.GeneratedMessage {
  factory UpdateTelegramChannelRequest({
    TelegramChannel? channel,
  }) {
    final result = UpdateTelegramChannelRequest._();
    if (channel != null) result.channel = channel;
    return result;
  }

  UpdateTelegramChannelRequest._();

  factory UpdateTelegramChannelRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpdateTelegramChannelRequest()..mergeFromBuffer(data, registry);
  factory UpdateTelegramChannelRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpdateTelegramChannelRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'UpdateTelegramChannelRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: UpdateTelegramChannelRequest.$_createMessage)
    ..aOM<TelegramChannel>(1, _omitFieldNames ? '' : 'channel',
        subBuilder: TelegramChannel.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateTelegramChannelRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateTelegramChannelRequest copyWith(
          void Function(UpdateTelegramChannelRequest) updates) =>
      super.copyWith(
              (message) => updates(message as UpdateTelegramChannelRequest))
          as UpdateTelegramChannelRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use UpdateTelegramChannelRequest() / UpdateTelegramChannelRequest.new instead')
  static UpdateTelegramChannelRequest create() =>
      UpdateTelegramChannelRequest._();
  static $pb.GeneratedMessage $_createMessage() =>
      UpdateTelegramChannelRequest._();
  @$core.override
  UpdateTelegramChannelRequest createEmptyInstance() =>
      UpdateTelegramChannelRequest._();
  @$core.pragma('dart2js:noInline')
  static UpdateTelegramChannelRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<UpdateTelegramChannelRequest>(
          UpdateTelegramChannelRequest.$_createMessage);
  static UpdateTelegramChannelRequest? _defaultInstance;

  /// `id` selects the Channel and `revision` must match the stored one.
  @$pb.TagNumber(1)
  TelegramChannel get channel => $_getN(0);
  @$pb.TagNumber(1)
  set channel(TelegramChannel value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasChannel() => $_has(0);
  @$pb.TagNumber(1)
  void clearChannel() => $_clearField(1);
  @$pb.TagNumber(1)
  TelegramChannel ensureChannel() => $_ensure(0);
}

class UpdateTelegramChannelResponse extends $pb.GeneratedMessage {
  factory UpdateTelegramChannelResponse({
    TelegramChannel? channel,
  }) {
    final result = UpdateTelegramChannelResponse._();
    if (channel != null) result.channel = channel;
    return result;
  }

  UpdateTelegramChannelResponse._();

  factory UpdateTelegramChannelResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpdateTelegramChannelResponse()..mergeFromBuffer(data, registry);
  factory UpdateTelegramChannelResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpdateTelegramChannelResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'UpdateTelegramChannelResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: UpdateTelegramChannelResponse.$_createMessage)
    ..aOM<TelegramChannel>(1, _omitFieldNames ? '' : 'channel',
        subBuilder: TelegramChannel.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateTelegramChannelResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateTelegramChannelResponse copyWith(
          void Function(UpdateTelegramChannelResponse) updates) =>
      super.copyWith(
              (message) => updates(message as UpdateTelegramChannelResponse))
          as UpdateTelegramChannelResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use UpdateTelegramChannelResponse() / UpdateTelegramChannelResponse.new instead')
  static UpdateTelegramChannelResponse create() =>
      UpdateTelegramChannelResponse._();
  static $pb.GeneratedMessage $_createMessage() =>
      UpdateTelegramChannelResponse._();
  @$core.override
  UpdateTelegramChannelResponse createEmptyInstance() =>
      UpdateTelegramChannelResponse._();
  @$core.pragma('dart2js:noInline')
  static UpdateTelegramChannelResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<UpdateTelegramChannelResponse>(
          UpdateTelegramChannelResponse.$_createMessage);
  static UpdateTelegramChannelResponse? _defaultInstance;

  @$pb.TagNumber(1)
  TelegramChannel get channel => $_getN(0);
  @$pb.TagNumber(1)
  set channel(TelegramChannel value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasChannel() => $_has(0);
  @$pb.TagNumber(1)
  void clearChannel() => $_clearField(1);
  @$pb.TagNumber(1)
  TelegramChannel ensureChannel() => $_ensure(0);
}

class RotateTelegramChannelCredentialRequest extends $pb.GeneratedMessage {
  factory RotateTelegramChannelCredentialRequest({
    $core.String? channelId,
    $core.String? botToken,
  }) {
    final result = RotateTelegramChannelCredentialRequest._();
    if (channelId != null) result.channelId = channelId;
    if (botToken != null) result.botToken = botToken;
    return result;
  }

  RotateTelegramChannelCredentialRequest._();

  factory RotateTelegramChannelCredentialRequest.fromBuffer(
          $core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      RotateTelegramChannelCredentialRequest()..mergeFromBuffer(data, registry);
  factory RotateTelegramChannelCredentialRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      RotateTelegramChannelCredentialRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'RotateTelegramChannelCredentialRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance:
          RotateTelegramChannelCredentialRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'channelId')
    ..aOS(2, _omitFieldNames ? '' : 'botToken')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  RotateTelegramChannelCredentialRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  RotateTelegramChannelCredentialRequest copyWith(
          void Function(RotateTelegramChannelCredentialRequest) updates) =>
      super.copyWith((message) =>
              updates(message as RotateTelegramChannelCredentialRequest))
          as RotateTelegramChannelCredentialRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use RotateTelegramChannelCredentialRequest() / RotateTelegramChannelCredentialRequest.new instead')
  static RotateTelegramChannelCredentialRequest create() =>
      RotateTelegramChannelCredentialRequest._();
  static $pb.GeneratedMessage $_createMessage() =>
      RotateTelegramChannelCredentialRequest._();
  @$core.override
  RotateTelegramChannelCredentialRequest createEmptyInstance() =>
      RotateTelegramChannelCredentialRequest._();
  @$core.pragma('dart2js:noInline')
  static RotateTelegramChannelCredentialRequest getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<
              RotateTelegramChannelCredentialRequest>(
          RotateTelegramChannelCredentialRequest.$_createMessage);
  static RotateTelegramChannelCredentialRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get channelId => $_getSZ(0);
  @$pb.TagNumber(1)
  set channelId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasChannelId() => $_has(0);
  @$pb.TagNumber(1)
  void clearChannelId() => $_clearField(1);

  /// The replacement Bot Token. Must resolve to the Channel's existing Bot ID.
  @$pb.TagNumber(2)
  $core.String get botToken => $_getSZ(1);
  @$pb.TagNumber(2)
  set botToken($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasBotToken() => $_has(1);
  @$pb.TagNumber(2)
  void clearBotToken() => $_clearField(2);
}

class RotateTelegramChannelCredentialResponse extends $pb.GeneratedMessage {
  factory RotateTelegramChannelCredentialResponse({
    TelegramChannel? channel,
  }) {
    final result = RotateTelegramChannelCredentialResponse._();
    if (channel != null) result.channel = channel;
    return result;
  }

  RotateTelegramChannelCredentialResponse._();

  factory RotateTelegramChannelCredentialResponse.fromBuffer(
          $core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      RotateTelegramChannelCredentialResponse()
        ..mergeFromBuffer(data, registry);
  factory RotateTelegramChannelCredentialResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      RotateTelegramChannelCredentialResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'RotateTelegramChannelCredentialResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance:
          RotateTelegramChannelCredentialResponse.$_createMessage)
    ..aOM<TelegramChannel>(1, _omitFieldNames ? '' : 'channel',
        subBuilder: TelegramChannel.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  RotateTelegramChannelCredentialResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  RotateTelegramChannelCredentialResponse copyWith(
          void Function(RotateTelegramChannelCredentialResponse) updates) =>
      super.copyWith((message) =>
              updates(message as RotateTelegramChannelCredentialResponse))
          as RotateTelegramChannelCredentialResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use RotateTelegramChannelCredentialResponse() / RotateTelegramChannelCredentialResponse.new instead')
  static RotateTelegramChannelCredentialResponse create() =>
      RotateTelegramChannelCredentialResponse._();
  static $pb.GeneratedMessage $_createMessage() =>
      RotateTelegramChannelCredentialResponse._();
  @$core.override
  RotateTelegramChannelCredentialResponse createEmptyInstance() =>
      RotateTelegramChannelCredentialResponse._();
  @$core.pragma('dart2js:noInline')
  static RotateTelegramChannelCredentialResponse getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<
              RotateTelegramChannelCredentialResponse>(
          RotateTelegramChannelCredentialResponse.$_createMessage);
  static RotateTelegramChannelCredentialResponse? _defaultInstance;

  @$pb.TagNumber(1)
  TelegramChannel get channel => $_getN(0);
  @$pb.TagNumber(1)
  set channel(TelegramChannel value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasChannel() => $_has(0);
  @$pb.TagNumber(1)
  void clearChannel() => $_clearField(1);
  @$pb.TagNumber(1)
  TelegramChannel ensureChannel() => $_ensure(0);
}

class SetTelegramChannelEnabledRequest extends $pb.GeneratedMessage {
  factory SetTelegramChannelEnabledRequest({
    $core.String? channelId,
    $fixnum.Int64? revision,
    $core.bool? inboundEnabled,
    $core.bool? outboundEnabled,
  }) {
    final result = SetTelegramChannelEnabledRequest._();
    if (channelId != null) result.channelId = channelId;
    if (revision != null) result.revision = revision;
    if (inboundEnabled != null) result.inboundEnabled = inboundEnabled;
    if (outboundEnabled != null) result.outboundEnabled = outboundEnabled;
    return result;
  }

  SetTelegramChannelEnabledRequest._();

  factory SetTelegramChannelEnabledRequest.fromBuffer(
          $core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      SetTelegramChannelEnabledRequest()..mergeFromBuffer(data, registry);
  factory SetTelegramChannelEnabledRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      SetTelegramChannelEnabledRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'SetTelegramChannelEnabledRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: SetTelegramChannelEnabledRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'channelId')
    ..aInt64(2, _omitFieldNames ? '' : 'revision')
    ..aOB(3, _omitFieldNames ? '' : 'inboundEnabled')
    ..aOB(4, _omitFieldNames ? '' : 'outboundEnabled')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  SetTelegramChannelEnabledRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  SetTelegramChannelEnabledRequest copyWith(
          void Function(SetTelegramChannelEnabledRequest) updates) =>
      super.copyWith(
              (message) => updates(message as SetTelegramChannelEnabledRequest))
          as SetTelegramChannelEnabledRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use SetTelegramChannelEnabledRequest() / SetTelegramChannelEnabledRequest.new instead')
  static SetTelegramChannelEnabledRequest create() =>
      SetTelegramChannelEnabledRequest._();
  static $pb.GeneratedMessage $_createMessage() =>
      SetTelegramChannelEnabledRequest._();
  @$core.override
  SetTelegramChannelEnabledRequest createEmptyInstance() =>
      SetTelegramChannelEnabledRequest._();
  @$core.pragma('dart2js:noInline')
  static SetTelegramChannelEnabledRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<SetTelegramChannelEnabledRequest>(
          SetTelegramChannelEnabledRequest.$_createMessage);
  static SetTelegramChannelEnabledRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get channelId => $_getSZ(0);
  @$pb.TagNumber(1)
  set channelId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasChannelId() => $_has(0);
  @$pb.TagNumber(1)
  void clearChannelId() => $_clearField(1);

  @$pb.TagNumber(2)
  $fixnum.Int64 get revision => $_getI64(1);
  @$pb.TagNumber(2)
  set revision($fixnum.Int64 value) => $_setInt64(1, value);
  @$pb.TagNumber(2)
  $core.bool hasRevision() => $_has(1);
  @$pb.TagNumber(2)
  void clearRevision() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.bool get inboundEnabled => $_getBF(2);
  @$pb.TagNumber(3)
  set inboundEnabled($core.bool value) => $_setBool(2, value);
  @$pb.TagNumber(3)
  $core.bool hasInboundEnabled() => $_has(2);
  @$pb.TagNumber(3)
  void clearInboundEnabled() => $_clearField(3);

  @$pb.TagNumber(4)
  $core.bool get outboundEnabled => $_getBF(3);
  @$pb.TagNumber(4)
  set outboundEnabled($core.bool value) => $_setBool(3, value);
  @$pb.TagNumber(4)
  $core.bool hasOutboundEnabled() => $_has(3);
  @$pb.TagNumber(4)
  void clearOutboundEnabled() => $_clearField(4);
}

class SetTelegramChannelEnabledResponse extends $pb.GeneratedMessage {
  factory SetTelegramChannelEnabledResponse({
    TelegramChannel? channel,
    $core.Iterable<$core.String>? warnings,
  }) {
    final result = SetTelegramChannelEnabledResponse._();
    if (channel != null) result.channel = channel;
    if (warnings != null) result.warnings.addAll(warnings);
    return result;
  }

  SetTelegramChannelEnabledResponse._();

  factory SetTelegramChannelEnabledResponse.fromBuffer(
          $core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      SetTelegramChannelEnabledResponse()..mergeFromBuffer(data, registry);
  factory SetTelegramChannelEnabledResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      SetTelegramChannelEnabledResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'SetTelegramChannelEnabledResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: SetTelegramChannelEnabledResponse.$_createMessage)
    ..aOM<TelegramChannel>(1, _omitFieldNames ? '' : 'channel',
        subBuilder: TelegramChannel.$_createMessage)
    ..pPS(2, _omitFieldNames ? '' : 'warnings')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  SetTelegramChannelEnabledResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  SetTelegramChannelEnabledResponse copyWith(
          void Function(SetTelegramChannelEnabledResponse) updates) =>
      super.copyWith((message) =>
              updates(message as SetTelegramChannelEnabledResponse))
          as SetTelegramChannelEnabledResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use SetTelegramChannelEnabledResponse() / SetTelegramChannelEnabledResponse.new instead')
  static SetTelegramChannelEnabledResponse create() =>
      SetTelegramChannelEnabledResponse._();
  static $pb.GeneratedMessage $_createMessage() =>
      SetTelegramChannelEnabledResponse._();
  @$core.override
  SetTelegramChannelEnabledResponse createEmptyInstance() =>
      SetTelegramChannelEnabledResponse._();
  @$core.pragma('dart2js:noInline')
  static SetTelegramChannelEnabledResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<SetTelegramChannelEnabledResponse>(
          SetTelegramChannelEnabledResponse.$_createMessage);
  static SetTelegramChannelEnabledResponse? _defaultInstance;

  @$pb.TagNumber(1)
  TelegramChannel get channel => $_getN(0);
  @$pb.TagNumber(1)
  set channel(TelegramChannel value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasChannel() => $_has(0);
  @$pb.TagNumber(1)
  void clearChannel() => $_clearField(1);
  @$pb.TagNumber(1)
  TelegramChannel ensureChannel() => $_ensure(0);

  /// Human-readable preflight findings when the request was accepted but a
  /// capability is degraded (e.g. Group Privacy is on).
  @$pb.TagNumber(2)
  $pb.PbList<$core.String> get warnings => $_getList(1);
}

class DeleteTelegramChannelRequest extends $pb.GeneratedMessage {
  factory DeleteTelegramChannelRequest({
    $core.String? id,
  }) {
    final result = DeleteTelegramChannelRequest._();
    if (id != null) result.id = id;
    return result;
  }

  DeleteTelegramChannelRequest._();

  factory DeleteTelegramChannelRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DeleteTelegramChannelRequest()..mergeFromBuffer(data, registry);
  factory DeleteTelegramChannelRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DeleteTelegramChannelRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'DeleteTelegramChannelRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: DeleteTelegramChannelRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteTelegramChannelRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteTelegramChannelRequest copyWith(
          void Function(DeleteTelegramChannelRequest) updates) =>
      super.copyWith(
              (message) => updates(message as DeleteTelegramChannelRequest))
          as DeleteTelegramChannelRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use DeleteTelegramChannelRequest() / DeleteTelegramChannelRequest.new instead')
  static DeleteTelegramChannelRequest create() =>
      DeleteTelegramChannelRequest._();
  static $pb.GeneratedMessage $_createMessage() =>
      DeleteTelegramChannelRequest._();
  @$core.override
  DeleteTelegramChannelRequest createEmptyInstance() =>
      DeleteTelegramChannelRequest._();
  @$core.pragma('dart2js:noInline')
  static DeleteTelegramChannelRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<DeleteTelegramChannelRequest>(
          DeleteTelegramChannelRequest.$_createMessage);
  static DeleteTelegramChannelRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);
}

class DeleteTelegramChannelResponse extends $pb.GeneratedMessage {
  factory DeleteTelegramChannelResponse() => DeleteTelegramChannelResponse._();

  DeleteTelegramChannelResponse._();

  factory DeleteTelegramChannelResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DeleteTelegramChannelResponse()..mergeFromBuffer(data, registry);
  factory DeleteTelegramChannelResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DeleteTelegramChannelResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'DeleteTelegramChannelResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: DeleteTelegramChannelResponse.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteTelegramChannelResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteTelegramChannelResponse copyWith(
          void Function(DeleteTelegramChannelResponse) updates) =>
      super.copyWith(
              (message) => updates(message as DeleteTelegramChannelResponse))
          as DeleteTelegramChannelResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use DeleteTelegramChannelResponse() / DeleteTelegramChannelResponse.new instead')
  static DeleteTelegramChannelResponse create() =>
      DeleteTelegramChannelResponse._();
  static $pb.GeneratedMessage $_createMessage() =>
      DeleteTelegramChannelResponse._();
  @$core.override
  DeleteTelegramChannelResponse createEmptyInstance() =>
      DeleteTelegramChannelResponse._();
  @$core.pragma('dart2js:noInline')
  static DeleteTelegramChannelResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<DeleteTelegramChannelResponse>(
          DeleteTelegramChannelResponse.$_createMessage);
  static DeleteTelegramChannelResponse? _defaultInstance;
}

class GetTelegramChannelStatusRequest extends $pb.GeneratedMessage {
  factory GetTelegramChannelStatusRequest({
    $core.String? channelId,
  }) {
    final result = GetTelegramChannelStatusRequest._();
    if (channelId != null) result.channelId = channelId;
    return result;
  }

  GetTelegramChannelStatusRequest._();

  factory GetTelegramChannelStatusRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetTelegramChannelStatusRequest()..mergeFromBuffer(data, registry);
  factory GetTelegramChannelStatusRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetTelegramChannelStatusRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetTelegramChannelStatusRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: GetTelegramChannelStatusRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'channelId')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetTelegramChannelStatusRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetTelegramChannelStatusRequest copyWith(
          void Function(GetTelegramChannelStatusRequest) updates) =>
      super.copyWith(
              (message) => updates(message as GetTelegramChannelStatusRequest))
          as GetTelegramChannelStatusRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use GetTelegramChannelStatusRequest() / GetTelegramChannelStatusRequest.new instead')
  static GetTelegramChannelStatusRequest create() =>
      GetTelegramChannelStatusRequest._();
  static $pb.GeneratedMessage $_createMessage() =>
      GetTelegramChannelStatusRequest._();
  @$core.override
  GetTelegramChannelStatusRequest createEmptyInstance() =>
      GetTelegramChannelStatusRequest._();
  @$core.pragma('dart2js:noInline')
  static GetTelegramChannelStatusRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<GetTelegramChannelStatusRequest>(
          GetTelegramChannelStatusRequest.$_createMessage);
  static GetTelegramChannelStatusRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get channelId => $_getSZ(0);
  @$pb.TagNumber(1)
  set channelId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasChannelId() => $_has(0);
  @$pb.TagNumber(1)
  void clearChannelId() => $_clearField(1);
}

/// TelegramChannelStatus is derived on every read. Nothing here is persisted
/// as configuration — availability is observed, not declared.
class TelegramChannelStatus extends $pb.GeneratedMessage {
  factory TelegramChannelStatus({
    $core.String? channelId,
    $core.bool? inboundDesired,
    $core.bool? outboundDesired,
    TelegramCredentialState? credentialState,
    TelegramReceiveMode? receiveMode,
    $core.int? inboundDestinationCount,
    $core.int? outboundDestinationCount,
    $core.Iterable<$core.String>? blockers,
    $core.Iterable<$core.String>? warnings,
    $core.String? lastCredentialError,
    TelegramWebhookState? webhookState,
    $core.String? webhookUrl,
    $core.String? lastWebhookError,
    $0.Timestamp? lastReconciledAt,
    $core.bool? pollingLeader,
    $core.bool? queueReady,
    $fixnum.Int64? lastFetchedUpdateId,
    $fixnum.Int64? lastAcceptedUpdateId,
    $0.Timestamp? lastPolledAt,
    $core.String? lastReceiveError,
  }) {
    final result = TelegramChannelStatus._();
    if (channelId != null) result.channelId = channelId;
    if (inboundDesired != null) result.inboundDesired = inboundDesired;
    if (outboundDesired != null) result.outboundDesired = outboundDesired;
    if (credentialState != null) result.credentialState = credentialState;
    if (receiveMode != null) result.receiveMode = receiveMode;
    if (inboundDestinationCount != null)
      result.inboundDestinationCount = inboundDestinationCount;
    if (outboundDestinationCount != null)
      result.outboundDestinationCount = outboundDestinationCount;
    if (blockers != null) result.blockers.addAll(blockers);
    if (warnings != null) result.warnings.addAll(warnings);
    if (lastCredentialError != null)
      result.lastCredentialError = lastCredentialError;
    if (webhookState != null) result.webhookState = webhookState;
    if (webhookUrl != null) result.webhookUrl = webhookUrl;
    if (lastWebhookError != null) result.lastWebhookError = lastWebhookError;
    if (lastReconciledAt != null) result.lastReconciledAt = lastReconciledAt;
    if (pollingLeader != null) result.pollingLeader = pollingLeader;
    if (queueReady != null) result.queueReady = queueReady;
    if (lastFetchedUpdateId != null)
      result.lastFetchedUpdateId = lastFetchedUpdateId;
    if (lastAcceptedUpdateId != null)
      result.lastAcceptedUpdateId = lastAcceptedUpdateId;
    if (lastPolledAt != null) result.lastPolledAt = lastPolledAt;
    if (lastReceiveError != null) result.lastReceiveError = lastReceiveError;
    return result;
  }

  TelegramChannelStatus._();

  factory TelegramChannelStatus.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      TelegramChannelStatus()..mergeFromBuffer(data, registry);
  factory TelegramChannelStatus.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      TelegramChannelStatus()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'TelegramChannelStatus',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: TelegramChannelStatus.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'channelId')
    ..aOB(2, _omitFieldNames ? '' : 'inboundDesired')
    ..aOB(3, _omitFieldNames ? '' : 'outboundDesired')
    ..aE<TelegramCredentialState>(4, _omitFieldNames ? '' : 'credentialState',
        enumValues: TelegramCredentialState.values)
    ..aE<TelegramReceiveMode>(5, _omitFieldNames ? '' : 'receiveMode',
        enumValues: TelegramReceiveMode.values)
    ..aI(6, _omitFieldNames ? '' : 'inboundDestinationCount')
    ..aI(7, _omitFieldNames ? '' : 'outboundDestinationCount')
    ..pPS(8, _omitFieldNames ? '' : 'blockers')
    ..pPS(9, _omitFieldNames ? '' : 'warnings')
    ..aOS(10, _omitFieldNames ? '' : 'lastCredentialError')
    ..aE<TelegramWebhookState>(11, _omitFieldNames ? '' : 'webhookState',
        enumValues: TelegramWebhookState.values)
    ..aOS(12, _omitFieldNames ? '' : 'webhookUrl')
    ..aOS(13, _omitFieldNames ? '' : 'lastWebhookError')
    ..aOM<$0.Timestamp>(14, _omitFieldNames ? '' : 'lastReconciledAt',
        subBuilder: $0.Timestamp.$_createMessage)
    ..aOB(15, _omitFieldNames ? '' : 'pollingLeader')
    ..aOB(16, _omitFieldNames ? '' : 'queueReady')
    ..aInt64(17, _omitFieldNames ? '' : 'lastFetchedUpdateId')
    ..aInt64(18, _omitFieldNames ? '' : 'lastAcceptedUpdateId')
    ..aOM<$0.Timestamp>(19, _omitFieldNames ? '' : 'lastPolledAt',
        subBuilder: $0.Timestamp.$_createMessage)
    ..aOS(20, _omitFieldNames ? '' : 'lastReceiveError')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  TelegramChannelStatus clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  TelegramChannelStatus copyWith(
          void Function(TelegramChannelStatus) updates) =>
      super.copyWith((message) => updates(message as TelegramChannelStatus))
          as TelegramChannelStatus;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use TelegramChannelStatus() / TelegramChannelStatus.new instead')
  static TelegramChannelStatus create() => TelegramChannelStatus._();
  static $pb.GeneratedMessage $_createMessage() => TelegramChannelStatus._();
  @$core.override
  TelegramChannelStatus createEmptyInstance() => TelegramChannelStatus._();
  @$core.pragma('dart2js:noInline')
  static TelegramChannelStatus getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<TelegramChannelStatus>(
          TelegramChannelStatus.$_createMessage);
  static TelegramChannelStatus? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get channelId => $_getSZ(0);
  @$pb.TagNumber(1)
  set channelId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasChannelId() => $_has(0);
  @$pb.TagNumber(1)
  void clearChannelId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.bool get inboundDesired => $_getBF(1);
  @$pb.TagNumber(2)
  set inboundDesired($core.bool value) => $_setBool(1, value);
  @$pb.TagNumber(2)
  $core.bool hasInboundDesired() => $_has(1);
  @$pb.TagNumber(2)
  void clearInboundDesired() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.bool get outboundDesired => $_getBF(2);
  @$pb.TagNumber(3)
  set outboundDesired($core.bool value) => $_setBool(2, value);
  @$pb.TagNumber(3)
  $core.bool hasOutboundDesired() => $_has(2);
  @$pb.TagNumber(3)
  void clearOutboundDesired() => $_clearField(3);

  @$pb.TagNumber(4)
  TelegramCredentialState get credentialState => $_getN(3);
  @$pb.TagNumber(4)
  set credentialState(TelegramCredentialState value) => $_setField(4, value);
  @$pb.TagNumber(4)
  $core.bool hasCredentialState() => $_has(3);
  @$pb.TagNumber(4)
  void clearCredentialState() => $_clearField(4);

  @$pb.TagNumber(5)
  TelegramReceiveMode get receiveMode => $_getN(4);
  @$pb.TagNumber(5)
  set receiveMode(TelegramReceiveMode value) => $_setField(5, value);
  @$pb.TagNumber(5)
  $core.bool hasReceiveMode() => $_has(4);
  @$pb.TagNumber(5)
  void clearReceiveMode() => $_clearField(5);

  /// Number of enabled inbound Destinations under this Channel.
  @$pb.TagNumber(6)
  $core.int get inboundDestinationCount => $_getIZ(5);
  @$pb.TagNumber(6)
  set inboundDestinationCount($core.int value) => $_setSignedInt32(5, value);
  @$pb.TagNumber(6)
  $core.bool hasInboundDestinationCount() => $_has(5);
  @$pb.TagNumber(6)
  void clearInboundDestinationCount() => $_clearField(6);

  /// Number of enabled outbound Destinations under this Channel.
  @$pb.TagNumber(7)
  $core.int get outboundDestinationCount => $_getIZ(6);
  @$pb.TagNumber(7)
  set outboundDestinationCount($core.int value) => $_setSignedInt32(6, value);
  @$pb.TagNumber(7)
  $core.bool hasOutboundDestinationCount() => $_has(6);
  @$pb.TagNumber(7)
  void clearOutboundDestinationCount() => $_clearField(7);

  /// Unmet prerequisites that currently block enabling the Channel.
  @$pb.TagNumber(8)
  $pb.PbList<$core.String> get blockers => $_getList(7);

  /// Accepted but degraded conditions, e.g. Group Privacy enabled.
  @$pb.TagNumber(9)
  $pb.PbList<$core.String> get warnings => $_getList(8);

  @$pb.TagNumber(10)
  $core.String get lastCredentialError => $_getSZ(9);
  @$pb.TagNumber(10)
  set lastCredentialError($core.String value) => $_setString(9, value);
  @$pb.TagNumber(10)
  $core.bool hasLastCredentialError() => $_has(9);
  @$pb.TagNumber(10)
  void clearLastCredentialError() => $_clearField(10);

  /// Webhook registration state as of the last reconciliation.
  @$pb.TagNumber(11)
  TelegramWebhookState get webhookState => $_getN(10);
  @$pb.TagNumber(11)
  set webhookState(TelegramWebhookState value) => $_setField(11, value);
  @$pb.TagNumber(11)
  $core.bool hasWebhookState() => $_has(10);
  @$pb.TagNumber(11)
  void clearWebhookState() => $_clearField(11);

  /// The callback URL derived from the global base URL and Channel ID.
  /// Empty when no global base URL is configured.
  @$pb.TagNumber(12)
  $core.String get webhookUrl => $_getSZ(11);
  @$pb.TagNumber(12)
  set webhookUrl($core.String value) => $_setString(11, value);
  @$pb.TagNumber(12)
  $core.bool hasWebhookUrl() => $_has(11);
  @$pb.TagNumber(12)
  void clearWebhookUrl() => $_clearField(12);

  /// Sanitized failure summary from the last reconciliation. Never contains
  /// the per-Channel secret.
  @$pb.TagNumber(13)
  $core.String get lastWebhookError => $_getSZ(12);
  @$pb.TagNumber(13)
  set lastWebhookError($core.String value) => $_setString(12, value);
  @$pb.TagNumber(13)
  $core.bool hasLastWebhookError() => $_has(12);
  @$pb.TagNumber(13)
  void clearLastWebhookError() => $_clearField(13);

  /// When the reconciler last compared desired and observed state.
  @$pb.TagNumber(14)
  $0.Timestamp get lastReconciledAt => $_getN(13);
  @$pb.TagNumber(14)
  set lastReconciledAt($0.Timestamp value) => $_setField(14, value);
  @$pb.TagNumber(14)
  $core.bool hasLastReconciledAt() => $_has(13);
  @$pb.TagNumber(14)
  void clearLastReconciledAt() => $_clearField(14);
  @$pb.TagNumber(14)
  $0.Timestamp ensureLastReconciledAt() => $_ensure(13);

  /// Whether this Pod currently holds the Long Polling lease for the Channel.
  @$pb.TagNumber(15)
  $core.bool get pollingLeader => $_getBF(14);
  @$pb.TagNumber(15)
  set pollingLeader($core.bool value) => $_setBool(14, value);
  @$pb.TagNumber(15)
  $core.bool hasPollingLeader() => $_has(14);
  @$pb.TagNumber(15)
  void clearPollingLeader() => $_clearField(15);

  /// Whether the durable queue is reachable and correctly configured.
  @$pb.TagNumber(16)
  $core.bool get queueReady => $_getBF(15);
  @$pb.TagNumber(16)
  set queueReady($core.bool value) => $_setBool(15, value);
  @$pb.TagNumber(16)
  $core.bool hasQueueReady() => $_has(15);
  @$pb.TagNumber(16)
  void clearQueueReady() => $_clearField(16);

  /// Long Polling observations from the Pod answering this request. Empty on
  /// a Pod that does not hold the Channel's lease.
  @$pb.TagNumber(17)
  $fixnum.Int64 get lastFetchedUpdateId => $_getI64(16);
  @$pb.TagNumber(17)
  set lastFetchedUpdateId($fixnum.Int64 value) => $_setInt64(16, value);
  @$pb.TagNumber(17)
  $core.bool hasLastFetchedUpdateId() => $_has(16);
  @$pb.TagNumber(17)
  void clearLastFetchedUpdateId() => $_clearField(17);

  @$pb.TagNumber(18)
  $fixnum.Int64 get lastAcceptedUpdateId => $_getI64(17);
  @$pb.TagNumber(18)
  set lastAcceptedUpdateId($fixnum.Int64 value) => $_setInt64(17, value);
  @$pb.TagNumber(18)
  $core.bool hasLastAcceptedUpdateId() => $_has(17);
  @$pb.TagNumber(18)
  void clearLastAcceptedUpdateId() => $_clearField(18);

  @$pb.TagNumber(19)
  $0.Timestamp get lastPolledAt => $_getN(18);
  @$pb.TagNumber(19)
  set lastPolledAt($0.Timestamp value) => $_setField(19, value);
  @$pb.TagNumber(19)
  $core.bool hasLastPolledAt() => $_has(18);
  @$pb.TagNumber(19)
  void clearLastPolledAt() => $_clearField(19);
  @$pb.TagNumber(19)
  $0.Timestamp ensureLastPolledAt() => $_ensure(18);

  /// Sanitized failure summary from the receive path. Never contains
  /// credential material.
  @$pb.TagNumber(20)
  $core.String get lastReceiveError => $_getSZ(19);
  @$pb.TagNumber(20)
  set lastReceiveError($core.String value) => $_setString(19, value);
  @$pb.TagNumber(20)
  $core.bool hasLastReceiveError() => $_has(19);
  @$pb.TagNumber(20)
  void clearLastReceiveError() => $_clearField(20);
}

class GetTelegramChannelStatusResponse extends $pb.GeneratedMessage {
  factory GetTelegramChannelStatusResponse({
    TelegramChannelStatus? status,
  }) {
    final result = GetTelegramChannelStatusResponse._();
    if (status != null) result.status = status;
    return result;
  }

  GetTelegramChannelStatusResponse._();

  factory GetTelegramChannelStatusResponse.fromBuffer(
          $core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetTelegramChannelStatusResponse()..mergeFromBuffer(data, registry);
  factory GetTelegramChannelStatusResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetTelegramChannelStatusResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetTelegramChannelStatusResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: GetTelegramChannelStatusResponse.$_createMessage)
    ..aOM<TelegramChannelStatus>(1, _omitFieldNames ? '' : 'status',
        subBuilder: TelegramChannelStatus.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetTelegramChannelStatusResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetTelegramChannelStatusResponse copyWith(
          void Function(GetTelegramChannelStatusResponse) updates) =>
      super.copyWith(
              (message) => updates(message as GetTelegramChannelStatusResponse))
          as GetTelegramChannelStatusResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use GetTelegramChannelStatusResponse() / GetTelegramChannelStatusResponse.new instead')
  static GetTelegramChannelStatusResponse create() =>
      GetTelegramChannelStatusResponse._();
  static $pb.GeneratedMessage $_createMessage() =>
      GetTelegramChannelStatusResponse._();
  @$core.override
  GetTelegramChannelStatusResponse createEmptyInstance() =>
      GetTelegramChannelStatusResponse._();
  @$core.pragma('dart2js:noInline')
  static GetTelegramChannelStatusResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<GetTelegramChannelStatusResponse>(
          GetTelegramChannelStatusResponse.$_createMessage);
  static GetTelegramChannelStatusResponse? _defaultInstance;

  @$pb.TagNumber(1)
  TelegramChannelStatus get status => $_getN(0);
  @$pb.TagNumber(1)
  set status(TelegramChannelStatus value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasStatus() => $_has(0);
  @$pb.TagNumber(1)
  void clearStatus() => $_clearField(1);
  @$pb.TagNumber(1)
  TelegramChannelStatus ensureStatus() => $_ensure(0);
}

class ListTelegramDestinationsRequest extends $pb.GeneratedMessage {
  factory ListTelegramDestinationsRequest({
    $core.String? channelId,
  }) {
    final result = ListTelegramDestinationsRequest._();
    if (channelId != null) result.channelId = channelId;
    return result;
  }

  ListTelegramDestinationsRequest._();

  factory ListTelegramDestinationsRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListTelegramDestinationsRequest()..mergeFromBuffer(data, registry);
  factory ListTelegramDestinationsRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListTelegramDestinationsRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListTelegramDestinationsRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: ListTelegramDestinationsRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'channelId')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListTelegramDestinationsRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListTelegramDestinationsRequest copyWith(
          void Function(ListTelegramDestinationsRequest) updates) =>
      super.copyWith(
              (message) => updates(message as ListTelegramDestinationsRequest))
          as ListTelegramDestinationsRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use ListTelegramDestinationsRequest() / ListTelegramDestinationsRequest.new instead')
  static ListTelegramDestinationsRequest create() =>
      ListTelegramDestinationsRequest._();
  static $pb.GeneratedMessage $_createMessage() =>
      ListTelegramDestinationsRequest._();
  @$core.override
  ListTelegramDestinationsRequest createEmptyInstance() =>
      ListTelegramDestinationsRequest._();
  @$core.pragma('dart2js:noInline')
  static ListTelegramDestinationsRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ListTelegramDestinationsRequest>(
          ListTelegramDestinationsRequest.$_createMessage);
  static ListTelegramDestinationsRequest? _defaultInstance;

  /// Optional Channel filter. Empty lists every Destination in the workspace.
  @$pb.TagNumber(1)
  $core.String get channelId => $_getSZ(0);
  @$pb.TagNumber(1)
  set channelId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasChannelId() => $_has(0);
  @$pb.TagNumber(1)
  void clearChannelId() => $_clearField(1);
}

class ListTelegramDestinationsResponse extends $pb.GeneratedMessage {
  factory ListTelegramDestinationsResponse({
    $core.Iterable<TelegramDestination>? destinations,
  }) {
    final result = ListTelegramDestinationsResponse._();
    if (destinations != null) result.destinations.addAll(destinations);
    return result;
  }

  ListTelegramDestinationsResponse._();

  factory ListTelegramDestinationsResponse.fromBuffer(
          $core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListTelegramDestinationsResponse()..mergeFromBuffer(data, registry);
  factory ListTelegramDestinationsResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListTelegramDestinationsResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListTelegramDestinationsResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: ListTelegramDestinationsResponse.$_createMessage)
    ..pPM<TelegramDestination>(1, _omitFieldNames ? '' : 'destinations',
        subBuilder: TelegramDestination.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListTelegramDestinationsResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListTelegramDestinationsResponse copyWith(
          void Function(ListTelegramDestinationsResponse) updates) =>
      super.copyWith(
              (message) => updates(message as ListTelegramDestinationsResponse))
          as ListTelegramDestinationsResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use ListTelegramDestinationsResponse() / ListTelegramDestinationsResponse.new instead')
  static ListTelegramDestinationsResponse create() =>
      ListTelegramDestinationsResponse._();
  static $pb.GeneratedMessage $_createMessage() =>
      ListTelegramDestinationsResponse._();
  @$core.override
  ListTelegramDestinationsResponse createEmptyInstance() =>
      ListTelegramDestinationsResponse._();
  @$core.pragma('dart2js:noInline')
  static ListTelegramDestinationsResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ListTelegramDestinationsResponse>(
          ListTelegramDestinationsResponse.$_createMessage);
  static ListTelegramDestinationsResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $pb.PbList<TelegramDestination> get destinations => $_getList(0);
}

class GetTelegramDestinationRequest extends $pb.GeneratedMessage {
  factory GetTelegramDestinationRequest({
    $core.String? id,
  }) {
    final result = GetTelegramDestinationRequest._();
    if (id != null) result.id = id;
    return result;
  }

  GetTelegramDestinationRequest._();

  factory GetTelegramDestinationRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetTelegramDestinationRequest()..mergeFromBuffer(data, registry);
  factory GetTelegramDestinationRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetTelegramDestinationRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetTelegramDestinationRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: GetTelegramDestinationRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetTelegramDestinationRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetTelegramDestinationRequest copyWith(
          void Function(GetTelegramDestinationRequest) updates) =>
      super.copyWith(
              (message) => updates(message as GetTelegramDestinationRequest))
          as GetTelegramDestinationRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use GetTelegramDestinationRequest() / GetTelegramDestinationRequest.new instead')
  static GetTelegramDestinationRequest create() =>
      GetTelegramDestinationRequest._();
  static $pb.GeneratedMessage $_createMessage() =>
      GetTelegramDestinationRequest._();
  @$core.override
  GetTelegramDestinationRequest createEmptyInstance() =>
      GetTelegramDestinationRequest._();
  @$core.pragma('dart2js:noInline')
  static GetTelegramDestinationRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<GetTelegramDestinationRequest>(
          GetTelegramDestinationRequest.$_createMessage);
  static GetTelegramDestinationRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);
}

class GetTelegramDestinationResponse extends $pb.GeneratedMessage {
  factory GetTelegramDestinationResponse({
    TelegramDestination? destination,
  }) {
    final result = GetTelegramDestinationResponse._();
    if (destination != null) result.destination = destination;
    return result;
  }

  GetTelegramDestinationResponse._();

  factory GetTelegramDestinationResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetTelegramDestinationResponse()..mergeFromBuffer(data, registry);
  factory GetTelegramDestinationResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetTelegramDestinationResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetTelegramDestinationResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: GetTelegramDestinationResponse.$_createMessage)
    ..aOM<TelegramDestination>(1, _omitFieldNames ? '' : 'destination',
        subBuilder: TelegramDestination.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetTelegramDestinationResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetTelegramDestinationResponse copyWith(
          void Function(GetTelegramDestinationResponse) updates) =>
      super.copyWith(
              (message) => updates(message as GetTelegramDestinationResponse))
          as GetTelegramDestinationResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use GetTelegramDestinationResponse() / GetTelegramDestinationResponse.new instead')
  static GetTelegramDestinationResponse create() =>
      GetTelegramDestinationResponse._();
  static $pb.GeneratedMessage $_createMessage() =>
      GetTelegramDestinationResponse._();
  @$core.override
  GetTelegramDestinationResponse createEmptyInstance() =>
      GetTelegramDestinationResponse._();
  @$core.pragma('dart2js:noInline')
  static GetTelegramDestinationResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<GetTelegramDestinationResponse>(
          GetTelegramDestinationResponse.$_createMessage);
  static GetTelegramDestinationResponse? _defaultInstance;

  @$pb.TagNumber(1)
  TelegramDestination get destination => $_getN(0);
  @$pb.TagNumber(1)
  set destination(TelegramDestination value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasDestination() => $_has(0);
  @$pb.TagNumber(1)
  void clearDestination() => $_clearField(1);
  @$pb.TagNumber(1)
  TelegramDestination ensureDestination() => $_ensure(0);
}

class CreateTelegramDestinationRequest extends $pb.GeneratedMessage {
  factory CreateTelegramDestinationRequest({
    TelegramDestination? destination,
  }) {
    final result = CreateTelegramDestinationRequest._();
    if (destination != null) result.destination = destination;
    return result;
  }

  CreateTelegramDestinationRequest._();

  factory CreateTelegramDestinationRequest.fromBuffer(
          $core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateTelegramDestinationRequest()..mergeFromBuffer(data, registry);
  factory CreateTelegramDestinationRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateTelegramDestinationRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CreateTelegramDestinationRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: CreateTelegramDestinationRequest.$_createMessage)
    ..aOM<TelegramDestination>(1, _omitFieldNames ? '' : 'destination',
        subBuilder: TelegramDestination.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateTelegramDestinationRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateTelegramDestinationRequest copyWith(
          void Function(CreateTelegramDestinationRequest) updates) =>
      super.copyWith(
              (message) => updates(message as CreateTelegramDestinationRequest))
          as CreateTelegramDestinationRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use CreateTelegramDestinationRequest() / CreateTelegramDestinationRequest.new instead')
  static CreateTelegramDestinationRequest create() =>
      CreateTelegramDestinationRequest._();
  static $pb.GeneratedMessage $_createMessage() =>
      CreateTelegramDestinationRequest._();
  @$core.override
  CreateTelegramDestinationRequest createEmptyInstance() =>
      CreateTelegramDestinationRequest._();
  @$core.pragma('dart2js:noInline')
  static CreateTelegramDestinationRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<CreateTelegramDestinationRequest>(
          CreateTelegramDestinationRequest.$_createMessage);
  static CreateTelegramDestinationRequest? _defaultInstance;

  /// `id`, `revision`, timestamps, and `verification` are server-owned.
  @$pb.TagNumber(1)
  TelegramDestination get destination => $_getN(0);
  @$pb.TagNumber(1)
  set destination(TelegramDestination value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasDestination() => $_has(0);
  @$pb.TagNumber(1)
  void clearDestination() => $_clearField(1);
  @$pb.TagNumber(1)
  TelegramDestination ensureDestination() => $_ensure(0);
}

class CreateTelegramDestinationResponse extends $pb.GeneratedMessage {
  factory CreateTelegramDestinationResponse({
    TelegramDestination? destination,
  }) {
    final result = CreateTelegramDestinationResponse._();
    if (destination != null) result.destination = destination;
    return result;
  }

  CreateTelegramDestinationResponse._();

  factory CreateTelegramDestinationResponse.fromBuffer(
          $core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateTelegramDestinationResponse()..mergeFromBuffer(data, registry);
  factory CreateTelegramDestinationResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateTelegramDestinationResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CreateTelegramDestinationResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: CreateTelegramDestinationResponse.$_createMessage)
    ..aOM<TelegramDestination>(1, _omitFieldNames ? '' : 'destination',
        subBuilder: TelegramDestination.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateTelegramDestinationResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateTelegramDestinationResponse copyWith(
          void Function(CreateTelegramDestinationResponse) updates) =>
      super.copyWith((message) =>
              updates(message as CreateTelegramDestinationResponse))
          as CreateTelegramDestinationResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use CreateTelegramDestinationResponse() / CreateTelegramDestinationResponse.new instead')
  static CreateTelegramDestinationResponse create() =>
      CreateTelegramDestinationResponse._();
  static $pb.GeneratedMessage $_createMessage() =>
      CreateTelegramDestinationResponse._();
  @$core.override
  CreateTelegramDestinationResponse createEmptyInstance() =>
      CreateTelegramDestinationResponse._();
  @$core.pragma('dart2js:noInline')
  static CreateTelegramDestinationResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<CreateTelegramDestinationResponse>(
          CreateTelegramDestinationResponse.$_createMessage);
  static CreateTelegramDestinationResponse? _defaultInstance;

  @$pb.TagNumber(1)
  TelegramDestination get destination => $_getN(0);
  @$pb.TagNumber(1)
  set destination(TelegramDestination value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasDestination() => $_has(0);
  @$pb.TagNumber(1)
  void clearDestination() => $_clearField(1);
  @$pb.TagNumber(1)
  TelegramDestination ensureDestination() => $_ensure(0);
}

class UpdateTelegramDestinationRequest extends $pb.GeneratedMessage {
  factory UpdateTelegramDestinationRequest({
    TelegramDestination? destination,
  }) {
    final result = UpdateTelegramDestinationRequest._();
    if (destination != null) result.destination = destination;
    return result;
  }

  UpdateTelegramDestinationRequest._();

  factory UpdateTelegramDestinationRequest.fromBuffer(
          $core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpdateTelegramDestinationRequest()..mergeFromBuffer(data, registry);
  factory UpdateTelegramDestinationRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpdateTelegramDestinationRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'UpdateTelegramDestinationRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: UpdateTelegramDestinationRequest.$_createMessage)
    ..aOM<TelegramDestination>(1, _omitFieldNames ? '' : 'destination',
        subBuilder: TelegramDestination.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateTelegramDestinationRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateTelegramDestinationRequest copyWith(
          void Function(UpdateTelegramDestinationRequest) updates) =>
      super.copyWith(
              (message) => updates(message as UpdateTelegramDestinationRequest))
          as UpdateTelegramDestinationRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use UpdateTelegramDestinationRequest() / UpdateTelegramDestinationRequest.new instead')
  static UpdateTelegramDestinationRequest create() =>
      UpdateTelegramDestinationRequest._();
  static $pb.GeneratedMessage $_createMessage() =>
      UpdateTelegramDestinationRequest._();
  @$core.override
  UpdateTelegramDestinationRequest createEmptyInstance() =>
      UpdateTelegramDestinationRequest._();
  @$core.pragma('dart2js:noInline')
  static UpdateTelegramDestinationRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<UpdateTelegramDestinationRequest>(
          UpdateTelegramDestinationRequest.$_createMessage);
  static UpdateTelegramDestinationRequest? _defaultInstance;

  /// `id` selects the Destination and `revision` must match the stored one.
  @$pb.TagNumber(1)
  TelegramDestination get destination => $_getN(0);
  @$pb.TagNumber(1)
  set destination(TelegramDestination value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasDestination() => $_has(0);
  @$pb.TagNumber(1)
  void clearDestination() => $_clearField(1);
  @$pb.TagNumber(1)
  TelegramDestination ensureDestination() => $_ensure(0);
}

class UpdateTelegramDestinationResponse extends $pb.GeneratedMessage {
  factory UpdateTelegramDestinationResponse({
    TelegramDestination? destination,
  }) {
    final result = UpdateTelegramDestinationResponse._();
    if (destination != null) result.destination = destination;
    return result;
  }

  UpdateTelegramDestinationResponse._();

  factory UpdateTelegramDestinationResponse.fromBuffer(
          $core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpdateTelegramDestinationResponse()..mergeFromBuffer(data, registry);
  factory UpdateTelegramDestinationResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpdateTelegramDestinationResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'UpdateTelegramDestinationResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: UpdateTelegramDestinationResponse.$_createMessage)
    ..aOM<TelegramDestination>(1, _omitFieldNames ? '' : 'destination',
        subBuilder: TelegramDestination.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateTelegramDestinationResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateTelegramDestinationResponse copyWith(
          void Function(UpdateTelegramDestinationResponse) updates) =>
      super.copyWith((message) =>
              updates(message as UpdateTelegramDestinationResponse))
          as UpdateTelegramDestinationResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use UpdateTelegramDestinationResponse() / UpdateTelegramDestinationResponse.new instead')
  static UpdateTelegramDestinationResponse create() =>
      UpdateTelegramDestinationResponse._();
  static $pb.GeneratedMessage $_createMessage() =>
      UpdateTelegramDestinationResponse._();
  @$core.override
  UpdateTelegramDestinationResponse createEmptyInstance() =>
      UpdateTelegramDestinationResponse._();
  @$core.pragma('dart2js:noInline')
  static UpdateTelegramDestinationResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<UpdateTelegramDestinationResponse>(
          UpdateTelegramDestinationResponse.$_createMessage);
  static UpdateTelegramDestinationResponse? _defaultInstance;

  @$pb.TagNumber(1)
  TelegramDestination get destination => $_getN(0);
  @$pb.TagNumber(1)
  set destination(TelegramDestination value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasDestination() => $_has(0);
  @$pb.TagNumber(1)
  void clearDestination() => $_clearField(1);
  @$pb.TagNumber(1)
  TelegramDestination ensureDestination() => $_ensure(0);
}

class DeleteTelegramDestinationRequest extends $pb.GeneratedMessage {
  factory DeleteTelegramDestinationRequest({
    $core.String? id,
  }) {
    final result = DeleteTelegramDestinationRequest._();
    if (id != null) result.id = id;
    return result;
  }

  DeleteTelegramDestinationRequest._();

  factory DeleteTelegramDestinationRequest.fromBuffer(
          $core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DeleteTelegramDestinationRequest()..mergeFromBuffer(data, registry);
  factory DeleteTelegramDestinationRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DeleteTelegramDestinationRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'DeleteTelegramDestinationRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: DeleteTelegramDestinationRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteTelegramDestinationRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteTelegramDestinationRequest copyWith(
          void Function(DeleteTelegramDestinationRequest) updates) =>
      super.copyWith(
              (message) => updates(message as DeleteTelegramDestinationRequest))
          as DeleteTelegramDestinationRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use DeleteTelegramDestinationRequest() / DeleteTelegramDestinationRequest.new instead')
  static DeleteTelegramDestinationRequest create() =>
      DeleteTelegramDestinationRequest._();
  static $pb.GeneratedMessage $_createMessage() =>
      DeleteTelegramDestinationRequest._();
  @$core.override
  DeleteTelegramDestinationRequest createEmptyInstance() =>
      DeleteTelegramDestinationRequest._();
  @$core.pragma('dart2js:noInline')
  static DeleteTelegramDestinationRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<DeleteTelegramDestinationRequest>(
          DeleteTelegramDestinationRequest.$_createMessage);
  static DeleteTelegramDestinationRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);
}

class DeleteTelegramDestinationResponse extends $pb.GeneratedMessage {
  factory DeleteTelegramDestinationResponse() =>
      DeleteTelegramDestinationResponse._();

  DeleteTelegramDestinationResponse._();

  factory DeleteTelegramDestinationResponse.fromBuffer(
          $core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DeleteTelegramDestinationResponse()..mergeFromBuffer(data, registry);
  factory DeleteTelegramDestinationResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DeleteTelegramDestinationResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'DeleteTelegramDestinationResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: DeleteTelegramDestinationResponse.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteTelegramDestinationResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteTelegramDestinationResponse copyWith(
          void Function(DeleteTelegramDestinationResponse) updates) =>
      super.copyWith((message) =>
              updates(message as DeleteTelegramDestinationResponse))
          as DeleteTelegramDestinationResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use DeleteTelegramDestinationResponse() / DeleteTelegramDestinationResponse.new instead')
  static DeleteTelegramDestinationResponse create() =>
      DeleteTelegramDestinationResponse._();
  static $pb.GeneratedMessage $_createMessage() =>
      DeleteTelegramDestinationResponse._();
  @$core.override
  DeleteTelegramDestinationResponse createEmptyInstance() =>
      DeleteTelegramDestinationResponse._();
  @$core.pragma('dart2js:noInline')
  static DeleteTelegramDestinationResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<DeleteTelegramDestinationResponse>(
          DeleteTelegramDestinationResponse.$_createMessage);
  static DeleteTelegramDestinationResponse? _defaultInstance;
}

class SendTelegramTestMessageRequest extends $pb.GeneratedMessage {
  factory SendTelegramTestMessageRequest({
    $core.String? destinationId,
    $core.String? text,
  }) {
    final result = SendTelegramTestMessageRequest._();
    if (destinationId != null) result.destinationId = destinationId;
    if (text != null) result.text = text;
    return result;
  }

  SendTelegramTestMessageRequest._();

  factory SendTelegramTestMessageRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      SendTelegramTestMessageRequest()..mergeFromBuffer(data, registry);
  factory SendTelegramTestMessageRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      SendTelegramTestMessageRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'SendTelegramTestMessageRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: SendTelegramTestMessageRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'destinationId')
    ..aOS(2, _omitFieldNames ? '' : 'text')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  SendTelegramTestMessageRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  SendTelegramTestMessageRequest copyWith(
          void Function(SendTelegramTestMessageRequest) updates) =>
      super.copyWith(
              (message) => updates(message as SendTelegramTestMessageRequest))
          as SendTelegramTestMessageRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use SendTelegramTestMessageRequest() / SendTelegramTestMessageRequest.new instead')
  static SendTelegramTestMessageRequest create() =>
      SendTelegramTestMessageRequest._();
  static $pb.GeneratedMessage $_createMessage() =>
      SendTelegramTestMessageRequest._();
  @$core.override
  SendTelegramTestMessageRequest createEmptyInstance() =>
      SendTelegramTestMessageRequest._();
  @$core.pragma('dart2js:noInline')
  static SendTelegramTestMessageRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<SendTelegramTestMessageRequest>(
          SendTelegramTestMessageRequest.$_createMessage);
  static SendTelegramTestMessageRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get destinationId => $_getSZ(0);
  @$pb.TagNumber(1)
  set destinationId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasDestinationId() => $_has(0);
  @$pb.TagNumber(1)
  void clearDestinationId() => $_clearField(1);

  /// Optional message body. Empty sends a default confirmation text.
  @$pb.TagNumber(2)
  $core.String get text => $_getSZ(1);
  @$pb.TagNumber(2)
  set text($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasText() => $_has(1);
  @$pb.TagNumber(2)
  void clearText() => $_clearField(2);
}

class SendTelegramTestMessageResponse extends $pb.GeneratedMessage {
  factory SendTelegramTestMessageResponse({
    TelegramDestination? destination,
    $core.Iterable<$core.String>? messageIds,
  }) {
    final result = SendTelegramTestMessageResponse._();
    if (destination != null) result.destination = destination;
    if (messageIds != null) result.messageIds.addAll(messageIds);
    return result;
  }

  SendTelegramTestMessageResponse._();

  factory SendTelegramTestMessageResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      SendTelegramTestMessageResponse()..mergeFromBuffer(data, registry);
  factory SendTelegramTestMessageResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      SendTelegramTestMessageResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'SendTelegramTestMessageResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: SendTelegramTestMessageResponse.$_createMessage)
    ..aOM<TelegramDestination>(1, _omitFieldNames ? '' : 'destination',
        subBuilder: TelegramDestination.$_createMessage)
    ..pPS(2, _omitFieldNames ? '' : 'messageIds')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  SendTelegramTestMessageResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  SendTelegramTestMessageResponse copyWith(
          void Function(SendTelegramTestMessageResponse) updates) =>
      super.copyWith(
              (message) => updates(message as SendTelegramTestMessageResponse))
          as SendTelegramTestMessageResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use SendTelegramTestMessageResponse() / SendTelegramTestMessageResponse.new instead')
  static SendTelegramTestMessageResponse create() =>
      SendTelegramTestMessageResponse._();
  static $pb.GeneratedMessage $_createMessage() =>
      SendTelegramTestMessageResponse._();
  @$core.override
  SendTelegramTestMessageResponse createEmptyInstance() =>
      SendTelegramTestMessageResponse._();
  @$core.pragma('dart2js:noInline')
  static SendTelegramTestMessageResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<SendTelegramTestMessageResponse>(
          SendTelegramTestMessageResponse.$_createMessage);
  static SendTelegramTestMessageResponse? _defaultInstance;

  /// The Destination with refreshed verification metadata.
  @$pb.TagNumber(1)
  TelegramDestination get destination => $_getN(0);
  @$pb.TagNumber(1)
  set destination(TelegramDestination value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasDestination() => $_has(0);
  @$pb.TagNumber(1)
  void clearDestination() => $_clearField(1);
  @$pb.TagNumber(1)
  TelegramDestination ensureDestination() => $_ensure(0);

  /// Telegram message IDs produced by the send, in delivery order.
  @$pb.TagNumber(2)
  $pb.PbList<$core.String> get messageIds => $_getList(1);
}

/// TelegramSettings holds platform-level Telegram configuration.
class TelegramSettings extends $pb.GeneratedMessage {
  factory TelegramSettings({
    $core.String? webhookBaseUrl,
    $0.Timestamp? updatedAt,
  }) {
    final result = TelegramSettings._();
    if (webhookBaseUrl != null) result.webhookBaseUrl = webhookBaseUrl;
    if (updatedAt != null) result.updatedAt = updatedAt;
    return result;
  }

  TelegramSettings._();

  factory TelegramSettings.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      TelegramSettings()..mergeFromBuffer(data, registry);
  factory TelegramSettings.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      TelegramSettings()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'TelegramSettings',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: TelegramSettings.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'webhookBaseUrl')
    ..aOM<$0.Timestamp>(2, _omitFieldNames ? '' : 'updatedAt',
        subBuilder: $0.Timestamp.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  TelegramSettings clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  TelegramSettings copyWith(void Function(TelegramSettings) updates) =>
      super.copyWith((message) => updates(message as TelegramSettings))
          as TelegramSettings;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use TelegramSettings() / TelegramSettings.new instead')
  static TelegramSettings create() => TelegramSettings._();
  static $pb.GeneratedMessage $_createMessage() => TelegramSettings._();
  @$core.override
  TelegramSettings createEmptyInstance() => TelegramSettings._();
  @$core.pragma('dart2js:noInline')
  static TelegramSettings getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<TelegramSettings>(
          TelegramSettings.$_createMessage);
  static TelegramSettings? _defaultInstance;

  /// Public base URL Telegram callbacks are delivered to, e.g.
  /// "https://butter.example.com". Per-Channel callback URLs are derived from
  /// this and the immutable Channel ID, so rotating the host does not require
  /// touching any Channel.
  /// @gotags: json:"webhook_base_url,omitempty" bson:"webhook_base_url,omitempty"
  @$pb.TagNumber(1)
  $core.String get webhookBaseUrl => $_getSZ(0);
  @$pb.TagNumber(1)
  set webhookBaseUrl($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasWebhookBaseUrl() => $_has(0);
  @$pb.TagNumber(1)
  void clearWebhookBaseUrl() => $_clearField(1);

  /// @gotags: json:"updated_at,omitempty" bson:"updated_at,omitempty"
  @$pb.TagNumber(2)
  $0.Timestamp get updatedAt => $_getN(1);
  @$pb.TagNumber(2)
  set updatedAt($0.Timestamp value) => $_setField(2, value);
  @$pb.TagNumber(2)
  $core.bool hasUpdatedAt() => $_has(1);
  @$pb.TagNumber(2)
  void clearUpdatedAt() => $_clearField(2);
  @$pb.TagNumber(2)
  $0.Timestamp ensureUpdatedAt() => $_ensure(1);
}

class GetTelegramSettingsRequest extends $pb.GeneratedMessage {
  factory GetTelegramSettingsRequest() => GetTelegramSettingsRequest._();

  GetTelegramSettingsRequest._();

  factory GetTelegramSettingsRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetTelegramSettingsRequest()..mergeFromBuffer(data, registry);
  factory GetTelegramSettingsRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetTelegramSettingsRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetTelegramSettingsRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: GetTelegramSettingsRequest.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetTelegramSettingsRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetTelegramSettingsRequest copyWith(
          void Function(GetTelegramSettingsRequest) updates) =>
      super.copyWith(
              (message) => updates(message as GetTelegramSettingsRequest))
          as GetTelegramSettingsRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use GetTelegramSettingsRequest() / GetTelegramSettingsRequest.new instead')
  static GetTelegramSettingsRequest create() => GetTelegramSettingsRequest._();
  static $pb.GeneratedMessage $_createMessage() =>
      GetTelegramSettingsRequest._();
  @$core.override
  GetTelegramSettingsRequest createEmptyInstance() =>
      GetTelegramSettingsRequest._();
  @$core.pragma('dart2js:noInline')
  static GetTelegramSettingsRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<GetTelegramSettingsRequest>(
          GetTelegramSettingsRequest.$_createMessage);
  static GetTelegramSettingsRequest? _defaultInstance;
}

class GetTelegramSettingsResponse extends $pb.GeneratedMessage {
  factory GetTelegramSettingsResponse({
    TelegramSettings? settings,
  }) {
    final result = GetTelegramSettingsResponse._();
    if (settings != null) result.settings = settings;
    return result;
  }

  GetTelegramSettingsResponse._();

  factory GetTelegramSettingsResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetTelegramSettingsResponse()..mergeFromBuffer(data, registry);
  factory GetTelegramSettingsResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetTelegramSettingsResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetTelegramSettingsResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: GetTelegramSettingsResponse.$_createMessage)
    ..aOM<TelegramSettings>(1, _omitFieldNames ? '' : 'settings',
        subBuilder: TelegramSettings.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetTelegramSettingsResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetTelegramSettingsResponse copyWith(
          void Function(GetTelegramSettingsResponse) updates) =>
      super.copyWith(
              (message) => updates(message as GetTelegramSettingsResponse))
          as GetTelegramSettingsResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use GetTelegramSettingsResponse() / GetTelegramSettingsResponse.new instead')
  static GetTelegramSettingsResponse create() =>
      GetTelegramSettingsResponse._();
  static $pb.GeneratedMessage $_createMessage() =>
      GetTelegramSettingsResponse._();
  @$core.override
  GetTelegramSettingsResponse createEmptyInstance() =>
      GetTelegramSettingsResponse._();
  @$core.pragma('dart2js:noInline')
  static GetTelegramSettingsResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<GetTelegramSettingsResponse>(
          GetTelegramSettingsResponse.$_createMessage);
  static GetTelegramSettingsResponse? _defaultInstance;

  @$pb.TagNumber(1)
  TelegramSettings get settings => $_getN(0);
  @$pb.TagNumber(1)
  set settings(TelegramSettings value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasSettings() => $_has(0);
  @$pb.TagNumber(1)
  void clearSettings() => $_clearField(1);
  @$pb.TagNumber(1)
  TelegramSettings ensureSettings() => $_ensure(0);
}

class UpdateTelegramSettingsRequest extends $pb.GeneratedMessage {
  factory UpdateTelegramSettingsRequest({
    TelegramSettings? settings,
  }) {
    final result = UpdateTelegramSettingsRequest._();
    if (settings != null) result.settings = settings;
    return result;
  }

  UpdateTelegramSettingsRequest._();

  factory UpdateTelegramSettingsRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpdateTelegramSettingsRequest()..mergeFromBuffer(data, registry);
  factory UpdateTelegramSettingsRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpdateTelegramSettingsRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'UpdateTelegramSettingsRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: UpdateTelegramSettingsRequest.$_createMessage)
    ..aOM<TelegramSettings>(1, _omitFieldNames ? '' : 'settings',
        subBuilder: TelegramSettings.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateTelegramSettingsRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateTelegramSettingsRequest copyWith(
          void Function(UpdateTelegramSettingsRequest) updates) =>
      super.copyWith(
              (message) => updates(message as UpdateTelegramSettingsRequest))
          as UpdateTelegramSettingsRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use UpdateTelegramSettingsRequest() / UpdateTelegramSettingsRequest.new instead')
  static UpdateTelegramSettingsRequest create() =>
      UpdateTelegramSettingsRequest._();
  static $pb.GeneratedMessage $_createMessage() =>
      UpdateTelegramSettingsRequest._();
  @$core.override
  UpdateTelegramSettingsRequest createEmptyInstance() =>
      UpdateTelegramSettingsRequest._();
  @$core.pragma('dart2js:noInline')
  static UpdateTelegramSettingsRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<UpdateTelegramSettingsRequest>(
          UpdateTelegramSettingsRequest.$_createMessage);
  static UpdateTelegramSettingsRequest? _defaultInstance;

  @$pb.TagNumber(1)
  TelegramSettings get settings => $_getN(0);
  @$pb.TagNumber(1)
  set settings(TelegramSettings value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasSettings() => $_has(0);
  @$pb.TagNumber(1)
  void clearSettings() => $_clearField(1);
  @$pb.TagNumber(1)
  TelegramSettings ensureSettings() => $_ensure(0);
}

class UpdateTelegramSettingsResponse extends $pb.GeneratedMessage {
  factory UpdateTelegramSettingsResponse({
    TelegramSettings? settings,
  }) {
    final result = UpdateTelegramSettingsResponse._();
    if (settings != null) result.settings = settings;
    return result;
  }

  UpdateTelegramSettingsResponse._();

  factory UpdateTelegramSettingsResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpdateTelegramSettingsResponse()..mergeFromBuffer(data, registry);
  factory UpdateTelegramSettingsResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpdateTelegramSettingsResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'UpdateTelegramSettingsResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: UpdateTelegramSettingsResponse.$_createMessage)
    ..aOM<TelegramSettings>(1, _omitFieldNames ? '' : 'settings',
        subBuilder: TelegramSettings.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateTelegramSettingsResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateTelegramSettingsResponse copyWith(
          void Function(UpdateTelegramSettingsResponse) updates) =>
      super.copyWith(
              (message) => updates(message as UpdateTelegramSettingsResponse))
          as UpdateTelegramSettingsResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use UpdateTelegramSettingsResponse() / UpdateTelegramSettingsResponse.new instead')
  static UpdateTelegramSettingsResponse create() =>
      UpdateTelegramSettingsResponse._();
  static $pb.GeneratedMessage $_createMessage() =>
      UpdateTelegramSettingsResponse._();
  @$core.override
  UpdateTelegramSettingsResponse createEmptyInstance() =>
      UpdateTelegramSettingsResponse._();
  @$core.pragma('dart2js:noInline')
  static UpdateTelegramSettingsResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<UpdateTelegramSettingsResponse>(
          UpdateTelegramSettingsResponse.$_createMessage);
  static UpdateTelegramSettingsResponse? _defaultInstance;

  @$pb.TagNumber(1)
  TelegramSettings get settings => $_getN(0);
  @$pb.TagNumber(1)
  set settings(TelegramSettings value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasSettings() => $_has(0);
  @$pb.TagNumber(1)
  void clearSettings() => $_clearField(1);
  @$pb.TagNumber(1)
  TelegramSettings ensureSettings() => $_ensure(0);
}

/// TelegramDeliverySegment is one piece of a response plus its delivery state.
class TelegramDeliverySegment extends $pb.GeneratedMessage {
  factory TelegramDeliverySegment({
    $core.int? index,
    $core.String? text,
    $core.String? status,
    $core.String? messageId,
    $core.String? error,
  }) {
    final result = TelegramDeliverySegment._();
    if (index != null) result.index = index;
    if (text != null) result.text = text;
    if (status != null) result.status = status;
    if (messageId != null) result.messageId = messageId;
    if (error != null) result.error = error;
    return result;
  }

  TelegramDeliverySegment._();

  factory TelegramDeliverySegment.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      TelegramDeliverySegment()..mergeFromBuffer(data, registry);
  factory TelegramDeliverySegment.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      TelegramDeliverySegment()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'TelegramDeliverySegment',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: TelegramDeliverySegment.$_createMessage)
    ..aI(1, _omitFieldNames ? '' : 'index')
    ..aOS(2, _omitFieldNames ? '' : 'text')
    ..aOS(3, _omitFieldNames ? '' : 'status')
    ..aOS(4, _omitFieldNames ? '' : 'messageId')
    ..aOS(5, _omitFieldNames ? '' : 'error')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  TelegramDeliverySegment clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  TelegramDeliverySegment copyWith(
          void Function(TelegramDeliverySegment) updates) =>
      super.copyWith((message) => updates(message as TelegramDeliverySegment))
          as TelegramDeliverySegment;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use TelegramDeliverySegment() / TelegramDeliverySegment.new instead')
  static TelegramDeliverySegment create() => TelegramDeliverySegment._();
  static $pb.GeneratedMessage $_createMessage() => TelegramDeliverySegment._();
  @$core.override
  TelegramDeliverySegment createEmptyInstance() => TelegramDeliverySegment._();
  @$core.pragma('dart2js:noInline')
  static TelegramDeliverySegment getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<TelegramDeliverySegment>(
          TelegramDeliverySegment.$_createMessage);
  static TelegramDeliverySegment? _defaultInstance;

  /// @gotags: json:"index,omitempty" bson:"index,omitempty"
  @$pb.TagNumber(1)
  $core.int get index => $_getIZ(0);
  @$pb.TagNumber(1)
  set index($core.int value) => $_setSignedInt32(0, value);
  @$pb.TagNumber(1)
  $core.bool hasIndex() => $_has(0);
  @$pb.TagNumber(1)
  void clearIndex() => $_clearField(1);

  /// @gotags: json:"text,omitempty" bson:"text,omitempty"
  @$pb.TagNumber(2)
  $core.String get text => $_getSZ(1);
  @$pb.TagNumber(2)
  set text($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasText() => $_has(1);
  @$pb.TagNumber(2)
  void clearText() => $_clearField(2);

  /// One of "pending", "sending", "sent", "failed". "sending" is persisted
  /// before the Telegram request and requires operator review after a crash.
  /// @gotags: json:"status,omitempty" bson:"status,omitempty"
  @$pb.TagNumber(3)
  $core.String get status => $_getSZ(2);
  @$pb.TagNumber(3)
  set status($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasStatus() => $_has(2);
  @$pb.TagNumber(3)
  void clearStatus() => $_clearField(3);

  /// Telegram message ID once sent.
  /// @gotags: json:"message_id,omitempty" bson:"message_id,omitempty"
  @$pb.TagNumber(4)
  $core.String get messageId => $_getSZ(3);
  @$pb.TagNumber(4)
  set messageId($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasMessageId() => $_has(3);
  @$pb.TagNumber(4)
  void clearMessageId() => $_clearField(4);

  /// Sanitized failure summary. Never contains credential material.
  /// @gotags: json:"error,omitempty" bson:"error,omitempty"
  @$pb.TagNumber(5)
  $core.String get error => $_getSZ(4);
  @$pb.TagNumber(5)
  set error($core.String value) => $_setString(4, value);
  @$pb.TagNumber(5)
  $core.bool hasError() => $_has(4);
  @$pb.TagNumber(5)
  void clearError() => $_clearField(5);
}

/// TelegramProcessingRecord is the auditable state of one accepted update.
class TelegramProcessingRecord extends $pb.GeneratedMessage {
  factory TelegramProcessingRecord({
    $core.String? id,
    $core.String? workspaceId,
    $core.String? channelId,
    $core.String? destinationId,
    $fixnum.Int64? updateId,
    TelegramProcessingStatus? status,
    $core.String? invocationId,
    $core.int? attempts,
    $fixnum.Int64? destinationRevision,
    $core.String? error,
    $core.String? output,
    $core.Iterable<TelegramDeliverySegment>? segments,
    $0.Timestamp? createdAt,
    $0.Timestamp? updatedAt,
    $0.Timestamp? expiresAt,
    $core.bool? deadLettered,
  }) {
    final result = TelegramProcessingRecord._();
    if (id != null) result.id = id;
    if (workspaceId != null) result.workspaceId = workspaceId;
    if (channelId != null) result.channelId = channelId;
    if (destinationId != null) result.destinationId = destinationId;
    if (updateId != null) result.updateId = updateId;
    if (status != null) result.status = status;
    if (invocationId != null) result.invocationId = invocationId;
    if (attempts != null) result.attempts = attempts;
    if (destinationRevision != null)
      result.destinationRevision = destinationRevision;
    if (error != null) result.error = error;
    if (output != null) result.output = output;
    if (segments != null) result.segments.addAll(segments);
    if (createdAt != null) result.createdAt = createdAt;
    if (updatedAt != null) result.updatedAt = updatedAt;
    if (expiresAt != null) result.expiresAt = expiresAt;
    if (deadLettered != null) result.deadLettered = deadLettered;
    return result;
  }

  TelegramProcessingRecord._();

  factory TelegramProcessingRecord.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      TelegramProcessingRecord()..mergeFromBuffer(data, registry);
  factory TelegramProcessingRecord.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      TelegramProcessingRecord()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'TelegramProcessingRecord',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: TelegramProcessingRecord.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..aOS(2, _omitFieldNames ? '' : 'workspaceId')
    ..aOS(3, _omitFieldNames ? '' : 'channelId')
    ..aOS(4, _omitFieldNames ? '' : 'destinationId')
    ..aInt64(5, _omitFieldNames ? '' : 'updateId')
    ..aE<TelegramProcessingStatus>(6, _omitFieldNames ? '' : 'status',
        enumValues: TelegramProcessingStatus.values)
    ..aOS(7, _omitFieldNames ? '' : 'invocationId')
    ..aI(8, _omitFieldNames ? '' : 'attempts')
    ..aInt64(9, _omitFieldNames ? '' : 'destinationRevision')
    ..aOS(10, _omitFieldNames ? '' : 'error')
    ..aOS(11, _omitFieldNames ? '' : 'output')
    ..pPM<TelegramDeliverySegment>(12, _omitFieldNames ? '' : 'segments',
        subBuilder: TelegramDeliverySegment.$_createMessage)
    ..aOM<$0.Timestamp>(13, _omitFieldNames ? '' : 'createdAt',
        subBuilder: $0.Timestamp.$_createMessage)
    ..aOM<$0.Timestamp>(14, _omitFieldNames ? '' : 'updatedAt',
        subBuilder: $0.Timestamp.$_createMessage)
    ..aOM<$0.Timestamp>(15, _omitFieldNames ? '' : 'expiresAt',
        subBuilder: $0.Timestamp.$_createMessage)
    ..aOB(16, _omitFieldNames ? '' : 'deadLettered')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  TelegramProcessingRecord clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  TelegramProcessingRecord copyWith(
          void Function(TelegramProcessingRecord) updates) =>
      super.copyWith((message) => updates(message as TelegramProcessingRecord))
          as TelegramProcessingRecord;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use TelegramProcessingRecord() / TelegramProcessingRecord.new instead')
  static TelegramProcessingRecord create() => TelegramProcessingRecord._();
  static $pb.GeneratedMessage $_createMessage() => TelegramProcessingRecord._();
  @$core.override
  TelegramProcessingRecord createEmptyInstance() =>
      TelegramProcessingRecord._();
  @$core.pragma('dart2js:noInline')
  static TelegramProcessingRecord getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<TelegramProcessingRecord>(
          TelegramProcessingRecord.$_createMessage);
  static TelegramProcessingRecord? _defaultInstance;

  /// Service-generated identifier.
  /// @gotags: json:"id,omitempty" bson:"id,omitempty"
  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);

  /// @gotags: json:"workspace_id,omitempty" bson:"workspace_id,omitempty"
  @$pb.TagNumber(2)
  $core.String get workspaceId => $_getSZ(1);
  @$pb.TagNumber(2)
  set workspaceId($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasWorkspaceId() => $_has(1);
  @$pb.TagNumber(2)
  void clearWorkspaceId() => $_clearField(2);

  /// @gotags: json:"channel_id,omitempty" bson:"channel_id,omitempty"
  @$pb.TagNumber(3)
  $core.String get channelId => $_getSZ(2);
  @$pb.TagNumber(3)
  set channelId($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasChannelId() => $_has(2);
  @$pb.TagNumber(3)
  void clearChannelId() => $_clearField(3);

  /// @gotags: json:"destination_id,omitempty" bson:"destination_id,omitempty"
  @$pb.TagNumber(4)
  $core.String get destinationId => $_getSZ(3);
  @$pb.TagNumber(4)
  set destinationId($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasDestinationId() => $_has(3);
  @$pb.TagNumber(4)
  void clearDestinationId() => $_clearField(4);

  /// Telegram's own update identifier; unique with channel_id.
  /// @gotags: json:"update_id,omitempty" bson:"update_id,omitempty"
  @$pb.TagNumber(5)
  $fixnum.Int64 get updateId => $_getI64(4);
  @$pb.TagNumber(5)
  set updateId($fixnum.Int64 value) => $_setInt64(4, value);
  @$pb.TagNumber(5)
  $core.bool hasUpdateId() => $_has(4);
  @$pb.TagNumber(5)
  void clearUpdateId() => $_clearField(5);

  /// @gotags: json:"status,omitempty" bson:"status,omitempty"
  @$pb.TagNumber(6)
  TelegramProcessingStatus get status => $_getN(5);
  @$pb.TagNumber(6)
  set status(TelegramProcessingStatus value) => $_setField(6, value);
  @$pb.TagNumber(6)
  $core.bool hasStatus() => $_has(5);
  @$pb.TagNumber(6)
  void clearStatus() => $_clearField(6);

  /// Stable across retries so an operator can correlate one logical unit of
  /// work with the Agent invocation it produced.
  /// @gotags: json:"invocation_id,omitempty" bson:"invocation_id,omitempty"
  @$pb.TagNumber(7)
  $core.String get invocationId => $_getSZ(6);
  @$pb.TagNumber(7)
  set invocationId($core.String value) => $_setString(6, value);
  @$pb.TagNumber(7)
  $core.bool hasInvocationId() => $_has(6);
  @$pb.TagNumber(7)
  void clearInvocationId() => $_clearField(7);

  /// How many times this record has been claimed.
  /// @gotags: json:"attempts,omitempty" bson:"attempts,omitempty"
  @$pb.TagNumber(8)
  $core.int get attempts => $_getIZ(7);
  @$pb.TagNumber(8)
  set attempts($core.int value) => $_setSignedInt32(7, value);
  @$pb.TagNumber(8)
  $core.bool hasAttempts() => $_has(7);
  @$pb.TagNumber(8)
  void clearAttempts() => $_clearField(8);

  /// Destination revision frozen at acceptance.
  /// @gotags: json:"destination_revision,omitempty" bson:"destination_revision,omitempty"
  @$pb.TagNumber(9)
  $fixnum.Int64 get destinationRevision => $_getI64(8);
  @$pb.TagNumber(9)
  set destinationRevision($fixnum.Int64 value) => $_setInt64(8, value);
  @$pb.TagNumber(9)
  $core.bool hasDestinationRevision() => $_has(8);
  @$pb.TagNumber(9)
  void clearDestinationRevision() => $_clearField(9);

  /// Sanitized failure summary. Never contains credential material or the
  /// full update body.
  /// @gotags: json:"error,omitempty" bson:"error,omitempty"
  @$pb.TagNumber(10)
  $core.String get error => $_getSZ(9);
  @$pb.TagNumber(10)
  set error($core.String value) => $_setString(9, value);
  @$pb.TagNumber(10)
  $core.bool hasError() => $_has(9);
  @$pb.TagNumber(10)
  void clearError() => $_clearField(10);

  /// The complete Agent response, persisted before delivery begins so a
  /// delivery retry never has to re-run the Agent.
  /// @gotags: json:"output,omitempty" bson:"output,omitempty"
  @$pb.TagNumber(11)
  $core.String get output => $_getSZ(10);
  @$pb.TagNumber(11)
  set output($core.String value) => $_setString(10, value);
  @$pb.TagNumber(11)
  $core.bool hasOutput() => $_has(10);
  @$pb.TagNumber(11)
  void clearOutput() => $_clearField(11);

  /// Per-segment delivery state.
  /// @gotags: json:"segments,omitempty" bson:"segments,omitempty"
  @$pb.TagNumber(12)
  $pb.PbList<TelegramDeliverySegment> get segments => $_getList(11);

  /// @gotags: json:"created_at,omitempty" bson:"created_at,omitempty"
  @$pb.TagNumber(13)
  $0.Timestamp get createdAt => $_getN(12);
  @$pb.TagNumber(13)
  set createdAt($0.Timestamp value) => $_setField(13, value);
  @$pb.TagNumber(13)
  $core.bool hasCreatedAt() => $_has(12);
  @$pb.TagNumber(13)
  void clearCreatedAt() => $_clearField(13);
  @$pb.TagNumber(13)
  $0.Timestamp ensureCreatedAt() => $_ensure(12);

  /// @gotags: json:"updated_at,omitempty" bson:"updated_at,omitempty"
  @$pb.TagNumber(14)
  $0.Timestamp get updatedAt => $_getN(13);
  @$pb.TagNumber(14)
  set updatedAt($0.Timestamp value) => $_setField(14, value);
  @$pb.TagNumber(14)
  $core.bool hasUpdatedAt() => $_has(13);
  @$pb.TagNumber(14)
  void clearUpdatedAt() => $_clearField(14);
  @$pb.TagNumber(14)
  $0.Timestamp ensureUpdatedAt() => $_ensure(13);

  /// When the record becomes eligible for TTL cleanup (30 days).
  /// @gotags: json:"expires_at,omitempty" bson:"expires_at,omitempty"
  @$pb.TagNumber(15)
  $0.Timestamp get expiresAt => $_getN(14);
  @$pb.TagNumber(15)
  set expiresAt($0.Timestamp value) => $_setField(15, value);
  @$pb.TagNumber(15)
  $core.bool hasExpiresAt() => $_has(14);
  @$pb.TagNumber(15)
  void clearExpiresAt() => $_clearField(15);
  @$pb.TagNumber(15)
  $0.Timestamp ensureExpiresAt() => $_ensure(14);

  /// Whether the record is in dead-letter handling and needs an operator.
  /// @gotags: json:"dead_lettered,omitempty" bson:"dead_lettered,omitempty"
  @$pb.TagNumber(16)
  $core.bool get deadLettered => $_getBF(15);
  @$pb.TagNumber(16)
  set deadLettered($core.bool value) => $_setBool(15, value);
  @$pb.TagNumber(16)
  $core.bool hasDeadLettered() => $_has(15);
  @$pb.TagNumber(16)
  void clearDeadLettered() => $_clearField(16);
}

class ListTelegramProcessingRecordsRequest extends $pb.GeneratedMessage {
  factory ListTelegramProcessingRecordsRequest({
    $core.String? channelId,
    $core.String? destinationId,
    TelegramProcessingStatus? status,
    $core.int? pageSize,
  }) {
    final result = ListTelegramProcessingRecordsRequest._();
    if (channelId != null) result.channelId = channelId;
    if (destinationId != null) result.destinationId = destinationId;
    if (status != null) result.status = status;
    if (pageSize != null) result.pageSize = pageSize;
    return result;
  }

  ListTelegramProcessingRecordsRequest._();

  factory ListTelegramProcessingRecordsRequest.fromBuffer(
          $core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListTelegramProcessingRecordsRequest()..mergeFromBuffer(data, registry);
  factory ListTelegramProcessingRecordsRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListTelegramProcessingRecordsRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListTelegramProcessingRecordsRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: ListTelegramProcessingRecordsRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'channelId')
    ..aOS(2, _omitFieldNames ? '' : 'destinationId')
    ..aE<TelegramProcessingStatus>(3, _omitFieldNames ? '' : 'status',
        enumValues: TelegramProcessingStatus.values)
    ..aI(4, _omitFieldNames ? '' : 'pageSize')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListTelegramProcessingRecordsRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListTelegramProcessingRecordsRequest copyWith(
          void Function(ListTelegramProcessingRecordsRequest) updates) =>
      super.copyWith((message) =>
              updates(message as ListTelegramProcessingRecordsRequest))
          as ListTelegramProcessingRecordsRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use ListTelegramProcessingRecordsRequest() / ListTelegramProcessingRecordsRequest.new instead')
  static ListTelegramProcessingRecordsRequest create() =>
      ListTelegramProcessingRecordsRequest._();
  static $pb.GeneratedMessage $_createMessage() =>
      ListTelegramProcessingRecordsRequest._();
  @$core.override
  ListTelegramProcessingRecordsRequest createEmptyInstance() =>
      ListTelegramProcessingRecordsRequest._();
  @$core.pragma('dart2js:noInline')
  static ListTelegramProcessingRecordsRequest getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<
              ListTelegramProcessingRecordsRequest>(
          ListTelegramProcessingRecordsRequest.$_createMessage);
  static ListTelegramProcessingRecordsRequest? _defaultInstance;

  /// Optional filters.
  @$pb.TagNumber(1)
  $core.String get channelId => $_getSZ(0);
  @$pb.TagNumber(1)
  set channelId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasChannelId() => $_has(0);
  @$pb.TagNumber(1)
  void clearChannelId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get destinationId => $_getSZ(1);
  @$pb.TagNumber(2)
  set destinationId($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasDestinationId() => $_has(1);
  @$pb.TagNumber(2)
  void clearDestinationId() => $_clearField(2);

  @$pb.TagNumber(3)
  TelegramProcessingStatus get status => $_getN(2);
  @$pb.TagNumber(3)
  set status(TelegramProcessingStatus value) => $_setField(3, value);
  @$pb.TagNumber(3)
  $core.bool hasStatus() => $_has(2);
  @$pb.TagNumber(3)
  void clearStatus() => $_clearField(3);

  /// Maximum records to return. Defaults to 50, capped at 200.
  @$pb.TagNumber(4)
  $core.int get pageSize => $_getIZ(3);
  @$pb.TagNumber(4)
  set pageSize($core.int value) => $_setSignedInt32(3, value);
  @$pb.TagNumber(4)
  $core.bool hasPageSize() => $_has(3);
  @$pb.TagNumber(4)
  void clearPageSize() => $_clearField(4);
}

class ListTelegramProcessingRecordsResponse extends $pb.GeneratedMessage {
  factory ListTelegramProcessingRecordsResponse({
    $core.Iterable<TelegramProcessingRecord>? records,
  }) {
    final result = ListTelegramProcessingRecordsResponse._();
    if (records != null) result.records.addAll(records);
    return result;
  }

  ListTelegramProcessingRecordsResponse._();

  factory ListTelegramProcessingRecordsResponse.fromBuffer(
          $core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListTelegramProcessingRecordsResponse()..mergeFromBuffer(data, registry);
  factory ListTelegramProcessingRecordsResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListTelegramProcessingRecordsResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListTelegramProcessingRecordsResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance:
          ListTelegramProcessingRecordsResponse.$_createMessage)
    ..pPM<TelegramProcessingRecord>(1, _omitFieldNames ? '' : 'records',
        subBuilder: TelegramProcessingRecord.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListTelegramProcessingRecordsResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListTelegramProcessingRecordsResponse copyWith(
          void Function(ListTelegramProcessingRecordsResponse) updates) =>
      super.copyWith((message) =>
              updates(message as ListTelegramProcessingRecordsResponse))
          as ListTelegramProcessingRecordsResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use ListTelegramProcessingRecordsResponse() / ListTelegramProcessingRecordsResponse.new instead')
  static ListTelegramProcessingRecordsResponse create() =>
      ListTelegramProcessingRecordsResponse._();
  static $pb.GeneratedMessage $_createMessage() =>
      ListTelegramProcessingRecordsResponse._();
  @$core.override
  ListTelegramProcessingRecordsResponse createEmptyInstance() =>
      ListTelegramProcessingRecordsResponse._();
  @$core.pragma('dart2js:noInline')
  static ListTelegramProcessingRecordsResponse getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<
              ListTelegramProcessingRecordsResponse>(
          ListTelegramProcessingRecordsResponse.$_createMessage);
  static ListTelegramProcessingRecordsResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $pb.PbList<TelegramProcessingRecord> get records => $_getList(0);
}

class GetTelegramProcessingRecordRequest extends $pb.GeneratedMessage {
  factory GetTelegramProcessingRecordRequest({
    $core.String? id,
  }) {
    final result = GetTelegramProcessingRecordRequest._();
    if (id != null) result.id = id;
    return result;
  }

  GetTelegramProcessingRecordRequest._();

  factory GetTelegramProcessingRecordRequest.fromBuffer(
          $core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetTelegramProcessingRecordRequest()..mergeFromBuffer(data, registry);
  factory GetTelegramProcessingRecordRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetTelegramProcessingRecordRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetTelegramProcessingRecordRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: GetTelegramProcessingRecordRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetTelegramProcessingRecordRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetTelegramProcessingRecordRequest copyWith(
          void Function(GetTelegramProcessingRecordRequest) updates) =>
      super.copyWith((message) =>
              updates(message as GetTelegramProcessingRecordRequest))
          as GetTelegramProcessingRecordRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use GetTelegramProcessingRecordRequest() / GetTelegramProcessingRecordRequest.new instead')
  static GetTelegramProcessingRecordRequest create() =>
      GetTelegramProcessingRecordRequest._();
  static $pb.GeneratedMessage $_createMessage() =>
      GetTelegramProcessingRecordRequest._();
  @$core.override
  GetTelegramProcessingRecordRequest createEmptyInstance() =>
      GetTelegramProcessingRecordRequest._();
  @$core.pragma('dart2js:noInline')
  static GetTelegramProcessingRecordRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<GetTelegramProcessingRecordRequest>(
          GetTelegramProcessingRecordRequest.$_createMessage);
  static GetTelegramProcessingRecordRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);
}

class GetTelegramProcessingRecordResponse extends $pb.GeneratedMessage {
  factory GetTelegramProcessingRecordResponse({
    TelegramProcessingRecord? record,
  }) {
    final result = GetTelegramProcessingRecordResponse._();
    if (record != null) result.record = record;
    return result;
  }

  GetTelegramProcessingRecordResponse._();

  factory GetTelegramProcessingRecordResponse.fromBuffer(
          $core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetTelegramProcessingRecordResponse()..mergeFromBuffer(data, registry);
  factory GetTelegramProcessingRecordResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetTelegramProcessingRecordResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetTelegramProcessingRecordResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: GetTelegramProcessingRecordResponse.$_createMessage)
    ..aOM<TelegramProcessingRecord>(1, _omitFieldNames ? '' : 'record',
        subBuilder: TelegramProcessingRecord.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetTelegramProcessingRecordResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetTelegramProcessingRecordResponse copyWith(
          void Function(GetTelegramProcessingRecordResponse) updates) =>
      super.copyWith((message) =>
              updates(message as GetTelegramProcessingRecordResponse))
          as GetTelegramProcessingRecordResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use GetTelegramProcessingRecordResponse() / GetTelegramProcessingRecordResponse.new instead')
  static GetTelegramProcessingRecordResponse create() =>
      GetTelegramProcessingRecordResponse._();
  static $pb.GeneratedMessage $_createMessage() =>
      GetTelegramProcessingRecordResponse._();
  @$core.override
  GetTelegramProcessingRecordResponse createEmptyInstance() =>
      GetTelegramProcessingRecordResponse._();
  @$core.pragma('dart2js:noInline')
  static GetTelegramProcessingRecordResponse getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<
              GetTelegramProcessingRecordResponse>(
          GetTelegramProcessingRecordResponse.$_createMessage);
  static GetTelegramProcessingRecordResponse? _defaultInstance;

  @$pb.TagNumber(1)
  TelegramProcessingRecord get record => $_getN(0);
  @$pb.TagNumber(1)
  set record(TelegramProcessingRecord value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasRecord() => $_has(0);
  @$pb.TagNumber(1)
  void clearRecord() => $_clearField(1);
  @$pb.TagNumber(1)
  TelegramProcessingRecord ensureRecord() => $_ensure(0);
}

class ResendTelegramReplyRequest extends $pb.GeneratedMessage {
  factory ResendTelegramReplyRequest({
    $core.String? id,
  }) {
    final result = ResendTelegramReplyRequest._();
    if (id != null) result.id = id;
    return result;
  }

  ResendTelegramReplyRequest._();

  factory ResendTelegramReplyRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ResendTelegramReplyRequest()..mergeFromBuffer(data, registry);
  factory ResendTelegramReplyRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ResendTelegramReplyRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ResendTelegramReplyRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: ResendTelegramReplyRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ResendTelegramReplyRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ResendTelegramReplyRequest copyWith(
          void Function(ResendTelegramReplyRequest) updates) =>
      super.copyWith(
              (message) => updates(message as ResendTelegramReplyRequest))
          as ResendTelegramReplyRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use ResendTelegramReplyRequest() / ResendTelegramReplyRequest.new instead')
  static ResendTelegramReplyRequest create() => ResendTelegramReplyRequest._();
  static $pb.GeneratedMessage $_createMessage() =>
      ResendTelegramReplyRequest._();
  @$core.override
  ResendTelegramReplyRequest createEmptyInstance() =>
      ResendTelegramReplyRequest._();
  @$core.pragma('dart2js:noInline')
  static ResendTelegramReplyRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ResendTelegramReplyRequest>(
          ResendTelegramReplyRequest.$_createMessage);
  static ResendTelegramReplyRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);
}

class ResendTelegramReplyResponse extends $pb.GeneratedMessage {
  factory ResendTelegramReplyResponse({
    TelegramProcessingRecord? record,
    $core.Iterable<$core.String>? messageIds,
  }) {
    final result = ResendTelegramReplyResponse._();
    if (record != null) result.record = record;
    if (messageIds != null) result.messageIds.addAll(messageIds);
    return result;
  }

  ResendTelegramReplyResponse._();

  factory ResendTelegramReplyResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ResendTelegramReplyResponse()..mergeFromBuffer(data, registry);
  factory ResendTelegramReplyResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ResendTelegramReplyResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ResendTelegramReplyResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: ResendTelegramReplyResponse.$_createMessage)
    ..aOM<TelegramProcessingRecord>(1, _omitFieldNames ? '' : 'record',
        subBuilder: TelegramProcessingRecord.$_createMessage)
    ..pPS(2, _omitFieldNames ? '' : 'messageIds')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ResendTelegramReplyResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ResendTelegramReplyResponse copyWith(
          void Function(ResendTelegramReplyResponse) updates) =>
      super.copyWith(
              (message) => updates(message as ResendTelegramReplyResponse))
          as ResendTelegramReplyResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use ResendTelegramReplyResponse() / ResendTelegramReplyResponse.new instead')
  static ResendTelegramReplyResponse create() =>
      ResendTelegramReplyResponse._();
  static $pb.GeneratedMessage $_createMessage() =>
      ResendTelegramReplyResponse._();
  @$core.override
  ResendTelegramReplyResponse createEmptyInstance() =>
      ResendTelegramReplyResponse._();
  @$core.pragma('dart2js:noInline')
  static ResendTelegramReplyResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ResendTelegramReplyResponse>(
          ResendTelegramReplyResponse.$_createMessage);
  static ResendTelegramReplyResponse? _defaultInstance;

  @$pb.TagNumber(1)
  TelegramProcessingRecord get record => $_getN(0);
  @$pb.TagNumber(1)
  set record(TelegramProcessingRecord value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasRecord() => $_has(0);
  @$pb.TagNumber(1)
  void clearRecord() => $_clearField(1);
  @$pb.TagNumber(1)
  TelegramProcessingRecord ensureRecord() => $_ensure(0);

  /// Telegram message IDs produced by this resend, in order.
  @$pb.TagNumber(2)
  $pb.PbList<$core.String> get messageIds => $_getList(1);
}

/// =============================================================================
/// TelegramChannelService
///
/// Reads are open to every workspace member (the auth middleware already
/// verified membership for the X-Workspace-ID header). Mutations require the
/// caller to hold the workspace "owner" or "admin" role; global admins bypass.
/// No RPC ever returns a Bot Token.
/// =============================================================================
class TelegramChannelServiceApi {
  final $pb.RpcClient _client;

  TelegramChannelServiceApi(this._client);

  /// ListTelegramChannels returns every Channel in the current workspace.
  $async.Future<ListTelegramChannelsResponse> listTelegramChannels(
          $pb.ClientContext? ctx, ListTelegramChannelsRequest request) =>
      _client.invoke<ListTelegramChannelsResponse>(
          ctx,
          'TelegramChannelService',
          'ListTelegramChannels',
          request,
          ListTelegramChannelsResponse());

  /// GetTelegramChannel returns one Channel by ID.
  $async.Future<GetTelegramChannelResponse> getTelegramChannel(
          $pb.ClientContext? ctx, GetTelegramChannelRequest request) =>
      _client.invoke<GetTelegramChannelResponse>(ctx, 'TelegramChannelService',
          'GetTelegramChannel', request, GetTelegramChannelResponse());

  /// CreateTelegramChannel validates the supplied Bot Token with Telegram
  /// `getMe` before any Channel or credential record is committed, then stores
  /// the resolved immutable Bot identity. Channels are always created with
  /// inbound and outbound disabled; enablement is a separate action.
  $async.Future<CreateTelegramChannelResponse> createTelegramChannel(
          $pb.ClientContext? ctx, CreateTelegramChannelRequest request) =>
      _client.invoke<CreateTelegramChannelResponse>(
          ctx,
          'TelegramChannelService',
          'CreateTelegramChannel',
          request,
          CreateTelegramChannelResponse());

  /// UpdateTelegramChannel changes the mutable configuration (name, receive
  /// mode). It requires the expected revision and rejects changes to the
  /// immutable key and Bot identity.
  $async.Future<UpdateTelegramChannelResponse> updateTelegramChannel(
          $pb.ClientContext? ctx, UpdateTelegramChannelRequest request) =>
      _client.invoke<UpdateTelegramChannelResponse>(
          ctx,
          'TelegramChannelService',
          'UpdateTelegramChannel',
          request,
          UpdateTelegramChannelResponse());

  /// RotateTelegramChannelCredential validates the new Bot Token first and
  /// accepts it only when it resolves to the same immutable Bot ID. The prior
  /// credential stays active if validation or persistence fails.
  $async.Future<RotateTelegramChannelCredentialResponse>
      rotateTelegramChannelCredential($pb.ClientContext? ctx,
              RotateTelegramChannelCredentialRequest request) =>
          _client.invoke<RotateTelegramChannelCredentialResponse>(
              ctx,
              'TelegramChannelService',
              'RotateTelegramChannelCredential',
              request,
              RotateTelegramChannelCredentialResponse());

  /// SetTelegramChannelEnabled applies the desired inbound/outbound state
  /// after a strict preflight. Enabling inbound without outbound is rejected.
  $async.Future<SetTelegramChannelEnabledResponse> setTelegramChannelEnabled(
          $pb.ClientContext? ctx, SetTelegramChannelEnabledRequest request) =>
      _client.invoke<SetTelegramChannelEnabledResponse>(
          ctx,
          'TelegramChannelService',
          'SetTelegramChannelEnabled',
          request,
          SetTelegramChannelEnabledResponse());

  /// DeleteTelegramChannel removes a Channel and its stored credential. It is
  /// refused while any Destination still references the Channel.
  $async.Future<DeleteTelegramChannelResponse> deleteTelegramChannel(
          $pb.ClientContext? ctx, DeleteTelegramChannelRequest request) =>
      _client.invoke<DeleteTelegramChannelResponse>(
          ctx,
          'TelegramChannelService',
          'DeleteTelegramChannel',
          request,
          DeleteTelegramChannelResponse());

  /// GetTelegramChannelStatus returns configured desired state plus derived
  /// runtime readiness. Runtime availability is never persisted as
  /// configuration.
  $async.Future<GetTelegramChannelStatusResponse> getTelegramChannelStatus(
          $pb.ClientContext? ctx, GetTelegramChannelStatusRequest request) =>
      _client.invoke<GetTelegramChannelStatusResponse>(
          ctx,
          'TelegramChannelService',
          'GetTelegramChannelStatus',
          request,
          GetTelegramChannelStatusResponse());
}

/// =============================================================================
/// TelegramDestinationService
/// =============================================================================
class TelegramDestinationServiceApi {
  final $pb.RpcClient _client;

  TelegramDestinationServiceApi(this._client);

  /// ListTelegramDestinations returns the workspace's Destinations, optionally
  /// narrowed to one Channel.
  $async.Future<ListTelegramDestinationsResponse> listTelegramDestinations(
          $pb.ClientContext? ctx, ListTelegramDestinationsRequest request) =>
      _client.invoke<ListTelegramDestinationsResponse>(
          ctx,
          'TelegramDestinationService',
          'ListTelegramDestinations',
          request,
          ListTelegramDestinationsResponse());

  /// GetTelegramDestination returns one Destination by ID.
  $async.Future<GetTelegramDestinationResponse> getTelegramDestination(
          $pb.ClientContext? ctx, GetTelegramDestinationRequest request) =>
      _client.invoke<GetTelegramDestinationResponse>(
          ctx,
          'TelegramDestinationService',
          'GetTelegramDestination',
          request,
          GetTelegramDestinationResponse());

  /// CreateTelegramDestination registers one exact Telegram address. Creating
  /// a Destination sends no Telegram message.
  $async.Future<CreateTelegramDestinationResponse> createTelegramDestination(
          $pb.ClientContext? ctx, CreateTelegramDestinationRequest request) =>
      _client.invoke<CreateTelegramDestinationResponse>(
          ctx,
          'TelegramDestinationService',
          'CreateTelegramDestination',
          request,
          CreateTelegramDestinationResponse());

  /// UpdateTelegramDestination changes name, enablement, and policy. It
  /// requires the expected revision and rejects any change to the immutable
  /// address; redirecting traffic requires creating a new Destination.
  $async.Future<UpdateTelegramDestinationResponse> updateTelegramDestination(
          $pb.ClientContext? ctx, UpdateTelegramDestinationRequest request) =>
      _client.invoke<UpdateTelegramDestinationResponse>(
          ctx,
          'TelegramDestinationService',
          'UpdateTelegramDestination',
          request,
          UpdateTelegramDestinationResponse());

  /// DeleteTelegramDestination removes a Destination. It is refused while a
  /// Notify Group or Cron job still references it.
  $async.Future<DeleteTelegramDestinationResponse> deleteTelegramDestination(
          $pb.ClientContext? ctx, DeleteTelegramDestinationRequest request) =>
      _client.invoke<DeleteTelegramDestinationResponse>(
          ctx,
          'TelegramDestinationService',
          'DeleteTelegramDestination',
          request,
          DeleteTelegramDestinationResponse());

  /// SendTelegramTestMessage delivers a message to the Destination through the
  /// unified sender. It is the explicit action that proves an address works:
  /// creating a Destination never sends anything. A successful send marks an
  /// unverified Destination verified.
  $async.Future<SendTelegramTestMessageResponse> sendTelegramTestMessage(
          $pb.ClientContext? ctx, SendTelegramTestMessageRequest request) =>
      _client.invoke<SendTelegramTestMessageResponse>(
          ctx,
          'TelegramDestinationService',
          'SendTelegramTestMessage',
          request,
          SendTelegramTestMessageResponse());
}

/// =============================================================================
/// TelegramAdminService manages platform-level Telegram settings.
///
/// The Webhook base URL is global, not per-workspace: it names the public
/// address of this deployment behind its load balancer, which is a platform
/// fact rather than a tenant choice. Only global administrators may read or
/// change it, so workspace input can never redirect Telegram callbacks.
/// =============================================================================
class TelegramAdminServiceApi {
  final $pb.RpcClient _client;

  TelegramAdminServiceApi(this._client);

  /// GetTelegramSettings returns the platform Telegram settings.
  $async.Future<GetTelegramSettingsResponse> getTelegramSettings(
          $pb.ClientContext? ctx, GetTelegramSettingsRequest request) =>
      _client.invoke<GetTelegramSettingsResponse>(ctx, 'TelegramAdminService',
          'GetTelegramSettings', request, GetTelegramSettingsResponse());

  /// UpdateTelegramSettings replaces the platform Telegram settings.
  $async.Future<UpdateTelegramSettingsResponse> updateTelegramSettings(
          $pb.ClientContext? ctx, UpdateTelegramSettingsRequest request) =>
      _client.invoke<UpdateTelegramSettingsResponse>(
          ctx,
          'TelegramAdminService',
          'UpdateTelegramSettings',
          request,
          UpdateTelegramSettingsResponse());
}

/// =============================================================================
/// TelegramProcessingService
///
/// Workspace members may read processing records and redacted failures for
/// their workspace; owners/admins may resend an already-produced reply.
/// There is deliberately no Agent rerun action: once Agent work may have
/// produced side effects, repeating it is an operator decision made with
/// context this API does not have.
/// =============================================================================
class TelegramProcessingServiceApi {
  final $pb.RpcClient _client;

  TelegramProcessingServiceApi(this._client);

  $async.Future<ListTelegramProcessingRecordsResponse>
      listTelegramProcessingRecords($pb.ClientContext? ctx,
              ListTelegramProcessingRecordsRequest request) =>
          _client.invoke<ListTelegramProcessingRecordsResponse>(
              ctx,
              'TelegramProcessingService',
              'ListTelegramProcessingRecords',
              request,
              ListTelegramProcessingRecordsResponse());
  $async.Future<GetTelegramProcessingRecordResponse>
      getTelegramProcessingRecord($pb.ClientContext? ctx,
              GetTelegramProcessingRecordRequest request) =>
          _client.invoke<GetTelegramProcessingRecordResponse>(
              ctx,
              'TelegramProcessingService',
              'GetTelegramProcessingRecord',
              request,
              GetTelegramProcessingRecordResponse());

  /// ResendTelegramReply re-delivers unsent or failed segments of a response
  /// that was already produced. It never invokes the Agent, and it is refused
  /// when no complete output exists.
  $async.Future<ResendTelegramReplyResponse> resendTelegramReply(
          $pb.ClientContext? ctx, ResendTelegramReplyRequest request) =>
      _client.invoke<ResendTelegramReplyResponse>(
          ctx,
          'TelegramProcessingService',
          'ResendTelegramReply',
          request,
          ResendTelegramReplyResponse());
}

const $core.bool _omitFieldNames =
    $core.bool.fromEnvironment('protobuf.omit_field_names');
const $core.bool _omitMessageNames =
    $core.bool.fromEnvironment('protobuf.omit_message_names');
