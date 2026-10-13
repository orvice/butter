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

import 'package:protobuf/protobuf.dart' as $pb;

class AgentChannelPlatform extends $pb.ProtobufEnum {
  static const AgentChannelPlatform AGENT_CHANNEL_PLATFORM_UNSPECIFIED =
      AgentChannelPlatform._(
          0, _omitEnumNames ? '' : 'AGENT_CHANNEL_PLATFORM_UNSPECIFIED');
  static const AgentChannelPlatform AGENT_CHANNEL_PLATFORM_TELEGRAM =
      AgentChannelPlatform._(
          1, _omitEnumNames ? '' : 'AGENT_CHANNEL_PLATFORM_TELEGRAM');
  static const AgentChannelPlatform AGENT_CHANNEL_PLATFORM_DISCORD =
      AgentChannelPlatform._(
          2, _omitEnumNames ? '' : 'AGENT_CHANNEL_PLATFORM_DISCORD');

  static const $core.List<AgentChannelPlatform> values = <AgentChannelPlatform>[
    AGENT_CHANNEL_PLATFORM_UNSPECIFIED,
    AGENT_CHANNEL_PLATFORM_TELEGRAM,
    AGENT_CHANNEL_PLATFORM_DISCORD,
  ];

  static final $core.List<AgentChannelPlatform?> _byValue =
      $pb.ProtobufEnum.$_initByValueList(values, 2);
  static AgentChannelPlatform? valueOf($core.int value) =>
      value < 0 || value >= _byValue.length ? null : _byValue[value];

  const AgentChannelPlatform._(super.value, super.name);
}

class AgentTriggerType extends $pb.ProtobufEnum {
  static const AgentTriggerType AGENT_TRIGGER_TYPE_UNSPECIFIED =
      AgentTriggerType._(
          0, _omitEnumNames ? '' : 'AGENT_TRIGGER_TYPE_UNSPECIFIED');
  static const AgentTriggerType AGENT_TRIGGER_TYPE_MESSAGE =
      AgentTriggerType._(1, _omitEnumNames ? '' : 'AGENT_TRIGGER_TYPE_MESSAGE');
  static const AgentTriggerType AGENT_TRIGGER_TYPE_COMMAND =
      AgentTriggerType._(2, _omitEnumNames ? '' : 'AGENT_TRIGGER_TYPE_COMMAND');
  static const AgentTriggerType AGENT_TRIGGER_TYPE_MENTION =
      AgentTriggerType._(3, _omitEnumNames ? '' : 'AGENT_TRIGGER_TYPE_MENTION');
  static const AgentTriggerType AGENT_TRIGGER_TYPE_PRIVATE_CHAT =
      AgentTriggerType._(
          4, _omitEnumNames ? '' : 'AGENT_TRIGGER_TYPE_PRIVATE_CHAT');

  static const $core.List<AgentTriggerType> values = <AgentTriggerType>[
    AGENT_TRIGGER_TYPE_UNSPECIFIED,
    AGENT_TRIGGER_TYPE_MESSAGE,
    AGENT_TRIGGER_TYPE_COMMAND,
    AGENT_TRIGGER_TYPE_MENTION,
    AGENT_TRIGGER_TYPE_PRIVATE_CHAT,
  ];

  static final $core.List<AgentTriggerType?> _byValue =
      $pb.ProtobufEnum.$_initByValueList(values, 4);
  static AgentTriggerType? valueOf($core.int value) =>
      value < 0 || value >= _byValue.length ? null : _byValue[value];

  const AgentTriggerType._(super.value, super.name);
}

class AgentReplyMode extends $pb.ProtobufEnum {
  static const AgentReplyMode AGENT_REPLY_MODE_UNSPECIFIED =
      AgentReplyMode._(0, _omitEnumNames ? '' : 'AGENT_REPLY_MODE_UNSPECIFIED');
  static const AgentReplyMode AGENT_REPLY_MODE_REPLY =
      AgentReplyMode._(1, _omitEnumNames ? '' : 'AGENT_REPLY_MODE_REPLY');
  static const AgentReplyMode AGENT_REPLY_MODE_THREAD =
      AgentReplyMode._(2, _omitEnumNames ? '' : 'AGENT_REPLY_MODE_THREAD');
  static const AgentReplyMode AGENT_REPLY_MODE_NEW_MESSAGE =
      AgentReplyMode._(3, _omitEnumNames ? '' : 'AGENT_REPLY_MODE_NEW_MESSAGE');

  static const $core.List<AgentReplyMode> values = <AgentReplyMode>[
    AGENT_REPLY_MODE_UNSPECIFIED,
    AGENT_REPLY_MODE_REPLY,
    AGENT_REPLY_MODE_THREAD,
    AGENT_REPLY_MODE_NEW_MESSAGE,
  ];

  static final $core.List<AgentReplyMode?> _byValue =
      $pb.ProtobufEnum.$_initByValueList(values, 3);
  static AgentReplyMode? valueOf($core.int value) =>
      value < 0 || value >= _byValue.length ? null : _byValue[value];

  const AgentReplyMode._(super.value, super.name);
}

class AgentSessionScope extends $pb.ProtobufEnum {
  static const AgentSessionScope AGENT_SESSION_SCOPE_UNSPECIFIED =
      AgentSessionScope._(
          0, _omitEnumNames ? '' : 'AGENT_SESSION_SCOPE_UNSPECIFIED');
  static const AgentSessionScope AGENT_SESSION_SCOPE_USER =
      AgentSessionScope._(1, _omitEnumNames ? '' : 'AGENT_SESSION_SCOPE_USER');
  static const AgentSessionScope AGENT_SESSION_SCOPE_CHAT =
      AgentSessionScope._(2, _omitEnumNames ? '' : 'AGENT_SESSION_SCOPE_CHAT');
  static const AgentSessionScope AGENT_SESSION_SCOPE_THREAD =
      AgentSessionScope._(
          3, _omitEnumNames ? '' : 'AGENT_SESSION_SCOPE_THREAD');
  static const AgentSessionScope AGENT_SESSION_SCOPE_CUSTOM =
      AgentSessionScope._(
          4, _omitEnumNames ? '' : 'AGENT_SESSION_SCOPE_CUSTOM');

  static const $core.List<AgentSessionScope> values = <AgentSessionScope>[
    AGENT_SESSION_SCOPE_UNSPECIFIED,
    AGENT_SESSION_SCOPE_USER,
    AGENT_SESSION_SCOPE_CHAT,
    AGENT_SESSION_SCOPE_THREAD,
    AGENT_SESSION_SCOPE_CUSTOM,
  ];

  static final $core.List<AgentSessionScope?> _byValue =
      $pb.ProtobufEnum.$_initByValueList(values, 4);
  static AgentSessionScope? valueOf($core.int value) =>
      value < 0 || value >= _byValue.length ? null : _byValue[value];

  const AgentSessionScope._(super.value, super.name);
}

const $core.bool _omitEnumNames =
    $core.bool.fromEnvironment('protobuf.omit_enum_names');
