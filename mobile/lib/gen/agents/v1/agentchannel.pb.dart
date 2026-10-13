// This is a generated file - do not edit.
//
// Generated from agents/v1/agentchannel.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names, prefer_relative_imports

import 'dart:core' as $core;

import 'package:fixnum/fixnum.dart' as $fixnum;
import 'package:protobuf/protobuf.dart' as $pb;

import 'agentchannel.pbenum.dart';

export 'package:protobuf/protobuf.dart' show GeneratedMessageGenericExtensions;

export 'agentchannel.pbenum.dart';

/// AgentChannel binds an agent to an external interaction platform.
class AgentChannel extends $pb.GeneratedMessage {
  factory AgentChannel({
    $core.String? name,
    $core.String? agentName,
    AgentChannelPlatform? platform,
    $core.bool? enabled,
    $core.Iterable<AgentTrigger>? triggers,
    AgentDelivery? delivery,
    AgentSessionBinding? session,
    $core.String? agentId,
    TelegramChannelConfig? telegram,
    $core.String? model,
    DiscordChannelConfig? discord,
    $core.Iterable<$core.MapEntry<$core.String, $core.String>>? metadata,
    $core.String? workspaceId,
  }) {
    final result = AgentChannel._();
    if (name != null) result.name = name;
    if (agentName != null) result.agentName = agentName;
    if (platform != null) result.platform = platform;
    if (enabled != null) result.enabled = enabled;
    if (triggers != null) result.triggers.addAll(triggers);
    if (delivery != null) result.delivery = delivery;
    if (session != null) result.session = session;
    if (agentId != null) result.agentId = agentId;
    if (telegram != null) result.telegram = telegram;
    if (model != null) result.model = model;
    if (discord != null) result.discord = discord;
    if (metadata != null) result.metadata.addEntries(metadata);
    if (workspaceId != null) result.workspaceId = workspaceId;
    return result;
  }

  AgentChannel._();

  factory AgentChannel.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      AgentChannel()..mergeFromBuffer(data, registry);
  factory AgentChannel.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      AgentChannel()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'AgentChannel',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: AgentChannel.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'name')
    ..aOS(2, _omitFieldNames ? '' : 'agentName')
    ..aE<AgentChannelPlatform>(3, _omitFieldNames ? '' : 'platform',
        enumValues: AgentChannelPlatform.values)
    ..aOB(4, _omitFieldNames ? '' : 'enabled')
    ..pPM<AgentTrigger>(5, _omitFieldNames ? '' : 'triggers',
        subBuilder: AgentTrigger.$_createMessage)
    ..aOM<AgentDelivery>(6, _omitFieldNames ? '' : 'delivery',
        subBuilder: AgentDelivery.$_createMessage)
    ..aOM<AgentSessionBinding>(7, _omitFieldNames ? '' : 'session',
        subBuilder: AgentSessionBinding.$_createMessage)
    ..aOS(8, _omitFieldNames ? '' : 'agentId')
    ..aOM<TelegramChannelConfig>(10, _omitFieldNames ? '' : 'telegram',
        subBuilder: TelegramChannelConfig.$_createMessage)
    ..aOS(11, _omitFieldNames ? '' : 'model')
    ..aOM<DiscordChannelConfig>(12, _omitFieldNames ? '' : 'discord',
        subBuilder: DiscordChannelConfig.$_createMessage)
    ..m<$core.String, $core.String>(20, _omitFieldNames ? '' : 'metadata',
        entryClassName: 'AgentChannel.MetadataEntry',
        keyFieldType: $pb.PbFieldType.OS,
        valueFieldType: $pb.PbFieldType.OS,
        packageName: const $pb.PackageName('agents.v1'))
    ..aOS(100, _omitFieldNames ? '' : 'workspaceId')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  AgentChannel clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  AgentChannel copyWith(void Function(AgentChannel) updates) =>
      super.copyWith((message) => updates(message as AgentChannel))
          as AgentChannel;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use AgentChannel() / AgentChannel.new instead')
  static AgentChannel create() => AgentChannel._();
  static $pb.GeneratedMessage $_createMessage() => AgentChannel._();
  @$core.override
  AgentChannel createEmptyInstance() => AgentChannel._();
  @$core.pragma('dart2js:noInline')
  static AgentChannel getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<AgentChannel>(
          AgentChannel.$_createMessage);
  static AgentChannel? _defaultInstance;

  /// @gotags: json:"name,omitempty" yaml:"name,omitempty"
  @$pb.TagNumber(1)
  $core.String get name => $_getSZ(0);
  @$pb.TagNumber(1)
  set name($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasName() => $_has(0);
  @$pb.TagNumber(1)
  void clearName() => $_clearField(1);

  /// Legacy agent name binding. Deprecated: use agent_id. Retained so
  /// historical records stay readable; the service layer backfills agent_id.
  /// @gotags: json:"agent_name,omitempty" yaml:"agent_name,omitempty"
  @$pb.TagNumber(2)
  $core.String get agentName => $_getSZ(1);
  @$pb.TagNumber(2)
  set agentName($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasAgentName() => $_has(1);
  @$pb.TagNumber(2)
  void clearAgentName() => $_clearField(2);

  /// @gotags: json:"platform,omitempty" yaml:"platform,omitempty"
  @$pb.TagNumber(3)
  AgentChannelPlatform get platform => $_getN(2);
  @$pb.TagNumber(3)
  set platform(AgentChannelPlatform value) => $_setField(3, value);
  @$pb.TagNumber(3)
  $core.bool hasPlatform() => $_has(2);
  @$pb.TagNumber(3)
  void clearPlatform() => $_clearField(3);

  /// @gotags: json:"enabled,omitempty" yaml:"enabled,omitempty"
  @$pb.TagNumber(4)
  $core.bool get enabled => $_getBF(3);
  @$pb.TagNumber(4)
  set enabled($core.bool value) => $_setBool(3, value);
  @$pb.TagNumber(4)
  $core.bool hasEnabled() => $_has(3);
  @$pb.TagNumber(4)
  void clearEnabled() => $_clearField(4);

  /// @gotags: json:"triggers,omitempty" yaml:"triggers,omitempty"
  @$pb.TagNumber(5)
  $pb.PbList<AgentTrigger> get triggers => $_getList(4);

  /// @gotags: json:"delivery,omitempty" yaml:"delivery,omitempty"
  @$pb.TagNumber(6)
  AgentDelivery get delivery => $_getN(5);
  @$pb.TagNumber(6)
  set delivery(AgentDelivery value) => $_setField(6, value);
  @$pb.TagNumber(6)
  $core.bool hasDelivery() => $_has(5);
  @$pb.TagNumber(6)
  void clearDelivery() => $_clearField(6);
  @$pb.TagNumber(6)
  AgentDelivery ensureDelivery() => $_ensure(5);

  /// @gotags: json:"session,omitempty" yaml:"session,omitempty"
  @$pb.TagNumber(7)
  AgentSessionBinding get session => $_getN(6);
  @$pb.TagNumber(7)
  set session(AgentSessionBinding value) => $_setField(7, value);
  @$pb.TagNumber(7)
  $core.bool hasSession() => $_has(6);
  @$pb.TagNumber(7)
  void clearSession() => $_clearField(7);
  @$pb.TagNumber(7)
  AgentSessionBinding ensureSession() => $_ensure(6);

  /// Immutable agent_id of the bound agent. Preferred over agent_name.
  /// @gotags: json:"agent_id,omitempty" yaml:"agent_id,omitempty" bson:"agent_id,omitempty"
  @$pb.TagNumber(8)
  $core.String get agentId => $_getSZ(7);
  @$pb.TagNumber(8)
  set agentId($core.String value) => $_setString(7, value);
  @$pb.TagNumber(8)
  $core.bool hasAgentId() => $_has(7);
  @$pb.TagNumber(8)
  void clearAgentId() => $_clearField(8);

  /// @gotags: json:"telegram,omitempty" yaml:"telegram,omitempty"
  @$pb.TagNumber(10)
  TelegramChannelConfig get telegram => $_getN(8);
  @$pb.TagNumber(10)
  set telegram(TelegramChannelConfig value) => $_setField(10, value);
  @$pb.TagNumber(10)
  $core.bool hasTelegram() => $_has(8);
  @$pb.TagNumber(10)
  void clearTelegram() => $_clearField(10);
  @$pb.TagNumber(10)
  TelegramChannelConfig ensureTelegram() => $_ensure(8);

  /// Optional model override (alias or name). Overrides the agent's configured model.
  /// @gotags: json:"model,omitempty" yaml:"model,omitempty"
  @$pb.TagNumber(11)
  $core.String get model => $_getSZ(9);
  @$pb.TagNumber(11)
  set model($core.String value) => $_setString(9, value);
  @$pb.TagNumber(11)
  $core.bool hasModel() => $_has(9);
  @$pb.TagNumber(11)
  void clearModel() => $_clearField(11);

  /// @gotags: json:"discord,omitempty" yaml:"discord,omitempty"
  @$pb.TagNumber(12)
  DiscordChannelConfig get discord => $_getN(10);
  @$pb.TagNumber(12)
  set discord(DiscordChannelConfig value) => $_setField(12, value);
  @$pb.TagNumber(12)
  $core.bool hasDiscord() => $_has(10);
  @$pb.TagNumber(12)
  void clearDiscord() => $_clearField(12);
  @$pb.TagNumber(12)
  DiscordChannelConfig ensureDiscord() => $_ensure(10);

  /// @gotags: json:"metadata,omitempty" yaml:"metadata,omitempty"
  @$pb.TagNumber(20)
  $pb.PbMap<$core.String, $core.String> get metadata => $_getMap(11);

  /// ID of the workspace this channel belongs to.
  /// @gotags: json:"workspace_id,omitempty" yaml:"workspace_id,omitempty" bson:"workspace_id,omitempty"
  @$pb.TagNumber(100)
  $core.String get workspaceId => $_getSZ(12);
  @$pb.TagNumber(100)
  set workspaceId($core.String value) => $_setString(12, value);
  @$pb.TagNumber(100)
  $core.bool hasWorkspaceId() => $_has(12);
  @$pb.TagNumber(100)
  void clearWorkspaceId() => $_clearField(100);
}

/// AgentTrigger defines when a channel should invoke the agent.
class AgentTrigger extends $pb.GeneratedMessage {
  factory AgentTrigger({
    AgentTriggerType? type,
    $core.Iterable<$core.String>? commands,
    $core.Iterable<$core.String>? prefixes,
    $core.bool? requireMention,
  }) {
    final result = AgentTrigger._();
    if (type != null) result.type = type;
    if (commands != null) result.commands.addAll(commands);
    if (prefixes != null) result.prefixes.addAll(prefixes);
    if (requireMention != null) result.requireMention = requireMention;
    return result;
  }

  AgentTrigger._();

  factory AgentTrigger.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      AgentTrigger()..mergeFromBuffer(data, registry);
  factory AgentTrigger.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      AgentTrigger()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'AgentTrigger',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: AgentTrigger.$_createMessage)
    ..aE<AgentTriggerType>(1, _omitFieldNames ? '' : 'type',
        enumValues: AgentTriggerType.values)
    ..pPS(2, _omitFieldNames ? '' : 'commands')
    ..pPS(3, _omitFieldNames ? '' : 'prefixes')
    ..aOB(4, _omitFieldNames ? '' : 'requireMention')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  AgentTrigger clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  AgentTrigger copyWith(void Function(AgentTrigger) updates) =>
      super.copyWith((message) => updates(message as AgentTrigger))
          as AgentTrigger;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use AgentTrigger() / AgentTrigger.new instead')
  static AgentTrigger create() => AgentTrigger._();
  static $pb.GeneratedMessage $_createMessage() => AgentTrigger._();
  @$core.override
  AgentTrigger createEmptyInstance() => AgentTrigger._();
  @$core.pragma('dart2js:noInline')
  static AgentTrigger getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<AgentTrigger>(
          AgentTrigger.$_createMessage);
  static AgentTrigger? _defaultInstance;

  /// @gotags: json:"type,omitempty" yaml:"type,omitempty"
  @$pb.TagNumber(1)
  AgentTriggerType get type => $_getN(0);
  @$pb.TagNumber(1)
  set type(AgentTriggerType value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasType() => $_has(0);
  @$pb.TagNumber(1)
  void clearType() => $_clearField(1);

  /// Commands used when type is COMMAND.
  /// @gotags: json:"commands,omitempty" yaml:"commands,omitempty"
  @$pb.TagNumber(2)
  $pb.PbList<$core.String> get commands => $_getList(1);

  /// Prefixes used when type is MESSAGE.
  /// @gotags: json:"prefixes,omitempty" yaml:"prefixes,omitempty"
  @$pb.TagNumber(3)
  $pb.PbList<$core.String> get prefixes => $_getList(2);

  /// When true, the agent is triggered only if explicitly mentioned.
  /// @gotags: json:"require_mention,omitempty" yaml:"require_mention,omitempty"
  @$pb.TagNumber(4)
  $core.bool get requireMention => $_getBF(3);
  @$pb.TagNumber(4)
  set requireMention($core.bool value) => $_setBool(3, value);
  @$pb.TagNumber(4)
  $core.bool hasRequireMention() => $_has(3);
  @$pb.TagNumber(4)
  void clearRequireMention() => $_clearField(4);
}

/// AgentDelivery defines how agent output is delivered back to the platform.
class AgentDelivery extends $pb.GeneratedMessage {
  factory AgentDelivery({
    AgentReplyMode? replyMode,
    $core.bool? streamingEnabled,
    $core.bool? sendTyping,
  }) {
    final result = AgentDelivery._();
    if (replyMode != null) result.replyMode = replyMode;
    if (streamingEnabled != null) result.streamingEnabled = streamingEnabled;
    if (sendTyping != null) result.sendTyping = sendTyping;
    return result;
  }

  AgentDelivery._();

  factory AgentDelivery.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      AgentDelivery()..mergeFromBuffer(data, registry);
  factory AgentDelivery.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      AgentDelivery()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'AgentDelivery',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: AgentDelivery.$_createMessage)
    ..aE<AgentReplyMode>(1, _omitFieldNames ? '' : 'replyMode',
        enumValues: AgentReplyMode.values)
    ..aOB(2, _omitFieldNames ? '' : 'streamingEnabled')
    ..aOB(3, _omitFieldNames ? '' : 'sendTyping')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  AgentDelivery clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  AgentDelivery copyWith(void Function(AgentDelivery) updates) =>
      super.copyWith((message) => updates(message as AgentDelivery))
          as AgentDelivery;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use AgentDelivery() / AgentDelivery.new instead')
  static AgentDelivery create() => AgentDelivery._();
  static $pb.GeneratedMessage $_createMessage() => AgentDelivery._();
  @$core.override
  AgentDelivery createEmptyInstance() => AgentDelivery._();
  @$core.pragma('dart2js:noInline')
  static AgentDelivery getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<AgentDelivery>(
          AgentDelivery.$_createMessage);
  static AgentDelivery? _defaultInstance;

  /// @gotags: json:"reply_mode,omitempty" yaml:"reply_mode,omitempty"
  @$pb.TagNumber(1)
  AgentReplyMode get replyMode => $_getN(0);
  @$pb.TagNumber(1)
  set replyMode(AgentReplyMode value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasReplyMode() => $_has(0);
  @$pb.TagNumber(1)
  void clearReplyMode() => $_clearField(1);

  /// @gotags: json:"streaming_enabled,omitempty" yaml:"streaming_enabled,omitempty"
  @$pb.TagNumber(2)
  $core.bool get streamingEnabled => $_getBF(1);
  @$pb.TagNumber(2)
  set streamingEnabled($core.bool value) => $_setBool(1, value);
  @$pb.TagNumber(2)
  $core.bool hasStreamingEnabled() => $_has(1);
  @$pb.TagNumber(2)
  void clearStreamingEnabled() => $_clearField(2);

  /// @gotags: json:"send_typing,omitempty" yaml:"send_typing,omitempty"
  @$pb.TagNumber(3)
  $core.bool get sendTyping => $_getBF(2);
  @$pb.TagNumber(3)
  set sendTyping($core.bool value) => $_setBool(2, value);
  @$pb.TagNumber(3)
  $core.bool hasSendTyping() => $_has(2);
  @$pb.TagNumber(3)
  void clearSendTyping() => $_clearField(3);
}

/// AgentSessionBinding defines how platform context maps to agent sessions.
class AgentSessionBinding extends $pb.GeneratedMessage {
  factory AgentSessionBinding({
    AgentSessionScope? scope,
    $core.String? keyTemplate,
  }) {
    final result = AgentSessionBinding._();
    if (scope != null) result.scope = scope;
    if (keyTemplate != null) result.keyTemplate = keyTemplate;
    return result;
  }

  AgentSessionBinding._();

  factory AgentSessionBinding.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      AgentSessionBinding()..mergeFromBuffer(data, registry);
  factory AgentSessionBinding.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      AgentSessionBinding()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'AgentSessionBinding',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: AgentSessionBinding.$_createMessage)
    ..aE<AgentSessionScope>(1, _omitFieldNames ? '' : 'scope',
        enumValues: AgentSessionScope.values)
    ..aOS(2, _omitFieldNames ? '' : 'keyTemplate')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  AgentSessionBinding clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  AgentSessionBinding copyWith(void Function(AgentSessionBinding) updates) =>
      super.copyWith((message) => updates(message as AgentSessionBinding))
          as AgentSessionBinding;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core
      .Deprecated('Use AgentSessionBinding() / AgentSessionBinding.new instead')
  static AgentSessionBinding create() => AgentSessionBinding._();
  static $pb.GeneratedMessage $_createMessage() => AgentSessionBinding._();
  @$core.override
  AgentSessionBinding createEmptyInstance() => AgentSessionBinding._();
  @$core.pragma('dart2js:noInline')
  static AgentSessionBinding getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<AgentSessionBinding>(
          AgentSessionBinding.$_createMessage);
  static AgentSessionBinding? _defaultInstance;

  /// @gotags: json:"scope,omitempty" yaml:"scope,omitempty"
  @$pb.TagNumber(1)
  AgentSessionScope get scope => $_getN(0);
  @$pb.TagNumber(1)
  set scope(AgentSessionScope value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasScope() => $_has(0);
  @$pb.TagNumber(1)
  void clearScope() => $_clearField(1);

  /// Optional template for custom session keys.
  /// @gotags: json:"key_template,omitempty" yaml:"key_template,omitempty"
  @$pb.TagNumber(2)
  $core.String get keyTemplate => $_getSZ(1);
  @$pb.TagNumber(2)
  set keyTemplate($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasKeyTemplate() => $_has(1);
  @$pb.TagNumber(2)
  void clearKeyTemplate() => $_clearField(2);
}

class TelegramChannelConfig extends $pb.GeneratedMessage {
  factory TelegramChannelConfig({
    $core.String? botToken,
    $core.String? webhookUrl,
    $core.String? webhookSecret,
    $core.Iterable<$fixnum.Int64>? allowedChatIds,
    $core.Iterable<$fixnum.Int64>? allowedUserIds,
    $core.bool? debug,
  }) {
    final result = TelegramChannelConfig._();
    if (botToken != null) result.botToken = botToken;
    if (webhookUrl != null) result.webhookUrl = webhookUrl;
    if (webhookSecret != null) result.webhookSecret = webhookSecret;
    if (allowedChatIds != null) result.allowedChatIds.addAll(allowedChatIds);
    if (allowedUserIds != null) result.allowedUserIds.addAll(allowedUserIds);
    if (debug != null) result.debug = debug;
    return result;
  }

  TelegramChannelConfig._();

  factory TelegramChannelConfig.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      TelegramChannelConfig()..mergeFromBuffer(data, registry);
  factory TelegramChannelConfig.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      TelegramChannelConfig()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'TelegramChannelConfig',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: TelegramChannelConfig.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'botToken')
    ..aOS(2, _omitFieldNames ? '' : 'webhookUrl')
    ..aOS(3, _omitFieldNames ? '' : 'webhookSecret')
    ..p<$fixnum.Int64>(
        4, _omitFieldNames ? '' : 'allowedChatIds', $pb.PbFieldType.K6)
    ..p<$fixnum.Int64>(
        5, _omitFieldNames ? '' : 'allowedUserIds', $pb.PbFieldType.K6)
    ..aOB(6, _omitFieldNames ? '' : 'debug')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  TelegramChannelConfig clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  TelegramChannelConfig copyWith(
          void Function(TelegramChannelConfig) updates) =>
      super.copyWith((message) => updates(message as TelegramChannelConfig))
          as TelegramChannelConfig;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use TelegramChannelConfig() / TelegramChannelConfig.new instead')
  static TelegramChannelConfig create() => TelegramChannelConfig._();
  static $pb.GeneratedMessage $_createMessage() => TelegramChannelConfig._();
  @$core.override
  TelegramChannelConfig createEmptyInstance() => TelegramChannelConfig._();
  @$core.pragma('dart2js:noInline')
  static TelegramChannelConfig getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<TelegramChannelConfig>(
          TelegramChannelConfig.$_createMessage);
  static TelegramChannelConfig? _defaultInstance;

  /// @gotags: json:"bot_token,omitempty" yaml:"bot_token,omitempty"
  @$pb.TagNumber(1)
  $core.String get botToken => $_getSZ(0);
  @$pb.TagNumber(1)
  set botToken($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasBotToken() => $_has(0);
  @$pb.TagNumber(1)
  void clearBotToken() => $_clearField(1);

  /// @gotags: json:"webhook_url,omitempty" yaml:"webhook_url,omitempty"
  @$pb.TagNumber(2)
  $core.String get webhookUrl => $_getSZ(1);
  @$pb.TagNumber(2)
  set webhookUrl($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasWebhookUrl() => $_has(1);
  @$pb.TagNumber(2)
  void clearWebhookUrl() => $_clearField(2);

  /// @gotags: json:"webhook_secret,omitempty" yaml:"webhook_secret,omitempty"
  @$pb.TagNumber(3)
  $core.String get webhookSecret => $_getSZ(2);
  @$pb.TagNumber(3)
  set webhookSecret($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasWebhookSecret() => $_has(2);
  @$pb.TagNumber(3)
  void clearWebhookSecret() => $_clearField(3);

  /// @gotags: json:"allowed_chat_ids,omitempty" yaml:"allowed_chat_ids,omitempty"
  @$pb.TagNumber(4)
  $pb.PbList<$fixnum.Int64> get allowedChatIds => $_getList(3);

  /// @gotags: json:"allowed_user_ids,omitempty" yaml:"allowed_user_ids,omitempty"
  @$pb.TagNumber(5)
  $pb.PbList<$fixnum.Int64> get allowedUserIds => $_getList(4);

  /// @gotags: json:"debug,omitempty" yaml:"debug,omitempty"
  @$pb.TagNumber(6)
  $core.bool get debug => $_getBF(5);
  @$pb.TagNumber(6)
  set debug($core.bool value) => $_setBool(5, value);
  @$pb.TagNumber(6)
  $core.bool hasDebug() => $_has(5);
  @$pb.TagNumber(6)
  void clearDebug() => $_clearField(6);
}

class DiscordChannelConfig extends $pb.GeneratedMessage {
  factory DiscordChannelConfig({
    $core.String? botToken,
    $core.Iterable<$core.String>? allowedGuildIds,
    $core.Iterable<$core.String>? allowedChannelIds,
    $core.bool? debug,
  }) {
    final result = DiscordChannelConfig._();
    if (botToken != null) result.botToken = botToken;
    if (allowedGuildIds != null) result.allowedGuildIds.addAll(allowedGuildIds);
    if (allowedChannelIds != null)
      result.allowedChannelIds.addAll(allowedChannelIds);
    if (debug != null) result.debug = debug;
    return result;
  }

  DiscordChannelConfig._();

  factory DiscordChannelConfig.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DiscordChannelConfig()..mergeFromBuffer(data, registry);
  factory DiscordChannelConfig.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DiscordChannelConfig()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'DiscordChannelConfig',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: DiscordChannelConfig.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'botToken')
    ..pPS(2, _omitFieldNames ? '' : 'allowedGuildIds')
    ..pPS(3, _omitFieldNames ? '' : 'allowedChannelIds')
    ..aOB(4, _omitFieldNames ? '' : 'debug')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DiscordChannelConfig clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DiscordChannelConfig copyWith(void Function(DiscordChannelConfig) updates) =>
      super.copyWith((message) => updates(message as DiscordChannelConfig))
          as DiscordChannelConfig;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use DiscordChannelConfig() / DiscordChannelConfig.new instead')
  static DiscordChannelConfig create() => DiscordChannelConfig._();
  static $pb.GeneratedMessage $_createMessage() => DiscordChannelConfig._();
  @$core.override
  DiscordChannelConfig createEmptyInstance() => DiscordChannelConfig._();
  @$core.pragma('dart2js:noInline')
  static DiscordChannelConfig getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<DiscordChannelConfig>(
          DiscordChannelConfig.$_createMessage);
  static DiscordChannelConfig? _defaultInstance;

  /// @gotags: json:"bot_token,omitempty" yaml:"bot_token,omitempty"
  @$pb.TagNumber(1)
  $core.String get botToken => $_getSZ(0);
  @$pb.TagNumber(1)
  set botToken($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasBotToken() => $_has(0);
  @$pb.TagNumber(1)
  void clearBotToken() => $_clearField(1);

  /// @gotags: json:"allowed_guild_ids,omitempty" yaml:"allowed_guild_ids,omitempty"
  @$pb.TagNumber(2)
  $pb.PbList<$core.String> get allowedGuildIds => $_getList(1);

  /// @gotags: json:"allowed_channel_ids,omitempty" yaml:"allowed_channel_ids,omitempty"
  @$pb.TagNumber(3)
  $pb.PbList<$core.String> get allowedChannelIds => $_getList(2);

  /// @gotags: json:"debug,omitempty" yaml:"debug,omitempty"
  @$pb.TagNumber(4)
  $core.bool get debug => $_getBF(3);
  @$pb.TagNumber(4)
  set debug($core.bool value) => $_setBool(3, value);
  @$pb.TagNumber(4)
  $core.bool hasDebug() => $_has(3);
  @$pb.TagNumber(4)
  void clearDebug() => $_clearField(4);
}

const $core.bool _omitFieldNames =
    $core.bool.fromEnvironment('protobuf.omit_field_names');
const $core.bool _omitMessageNames =
    $core.bool.fromEnvironment('protobuf.omit_message_names');
