// This is a generated file - do not edit.
//
// Generated from agents/v1/agentchannel.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names, prefer_relative_imports
// ignore_for_file: unused_import

import 'dart:convert' as $convert;
import 'dart:core' as $core;
import 'dart:typed_data' as $typed_data;

@$core.Deprecated('Use agentChannelPlatformDescriptor instead')
const AgentChannelPlatform$json = {
  '1': 'AgentChannelPlatform',
  '2': [
    {'1': 'AGENT_CHANNEL_PLATFORM_UNSPECIFIED', '2': 0},
    {'1': 'AGENT_CHANNEL_PLATFORM_TELEGRAM', '2': 1},
    {'1': 'AGENT_CHANNEL_PLATFORM_DISCORD', '2': 2},
  ],
};

/// Descriptor for `AgentChannelPlatform`. Decode as a `google.protobuf.EnumDescriptorProto`.
final $typed_data.Uint8List agentChannelPlatformDescriptor = $convert.base64Decode(
    'ChRBZ2VudENoYW5uZWxQbGF0Zm9ybRImCiJBR0VOVF9DSEFOTkVMX1BMQVRGT1JNX1VOU1BFQ0'
    'lGSUVEEAASIwofQUdFTlRfQ0hBTk5FTF9QTEFURk9STV9URUxFR1JBTRABEiIKHkFHRU5UX0NI'
    'QU5ORUxfUExBVEZPUk1fRElTQ09SRBAC');

@$core.Deprecated('Use agentTriggerTypeDescriptor instead')
const AgentTriggerType$json = {
  '1': 'AgentTriggerType',
  '2': [
    {'1': 'AGENT_TRIGGER_TYPE_UNSPECIFIED', '2': 0},
    {'1': 'AGENT_TRIGGER_TYPE_MESSAGE', '2': 1},
    {'1': 'AGENT_TRIGGER_TYPE_COMMAND', '2': 2},
    {'1': 'AGENT_TRIGGER_TYPE_MENTION', '2': 3},
    {'1': 'AGENT_TRIGGER_TYPE_PRIVATE_CHAT', '2': 4},
  ],
};

/// Descriptor for `AgentTriggerType`. Decode as a `google.protobuf.EnumDescriptorProto`.
final $typed_data.Uint8List agentTriggerTypeDescriptor = $convert.base64Decode(
    'ChBBZ2VudFRyaWdnZXJUeXBlEiIKHkFHRU5UX1RSSUdHRVJfVFlQRV9VTlNQRUNJRklFRBAAEh'
    '4KGkFHRU5UX1RSSUdHRVJfVFlQRV9NRVNTQUdFEAESHgoaQUdFTlRfVFJJR0dFUl9UWVBFX0NP'
    'TU1BTkQQAhIeChpBR0VOVF9UUklHR0VSX1RZUEVfTUVOVElPThADEiMKH0FHRU5UX1RSSUdHRV'
    'JfVFlQRV9QUklWQVRFX0NIQVQQBA==');

@$core.Deprecated('Use agentReplyModeDescriptor instead')
const AgentReplyMode$json = {
  '1': 'AgentReplyMode',
  '2': [
    {'1': 'AGENT_REPLY_MODE_UNSPECIFIED', '2': 0},
    {'1': 'AGENT_REPLY_MODE_REPLY', '2': 1},
    {'1': 'AGENT_REPLY_MODE_THREAD', '2': 2},
    {'1': 'AGENT_REPLY_MODE_NEW_MESSAGE', '2': 3},
  ],
};

/// Descriptor for `AgentReplyMode`. Decode as a `google.protobuf.EnumDescriptorProto`.
final $typed_data.Uint8List agentReplyModeDescriptor = $convert.base64Decode(
    'Cg5BZ2VudFJlcGx5TW9kZRIgChxBR0VOVF9SRVBMWV9NT0RFX1VOU1BFQ0lGSUVEEAASGgoWQU'
    'dFTlRfUkVQTFlfTU9ERV9SRVBMWRABEhsKF0FHRU5UX1JFUExZX01PREVfVEhSRUFEEAISIAoc'
    'QUdFTlRfUkVQTFlfTU9ERV9ORVdfTUVTU0FHRRAD');

@$core.Deprecated('Use agentSessionScopeDescriptor instead')
const AgentSessionScope$json = {
  '1': 'AgentSessionScope',
  '2': [
    {'1': 'AGENT_SESSION_SCOPE_UNSPECIFIED', '2': 0},
    {'1': 'AGENT_SESSION_SCOPE_USER', '2': 1},
    {'1': 'AGENT_SESSION_SCOPE_CHAT', '2': 2},
    {'1': 'AGENT_SESSION_SCOPE_THREAD', '2': 3},
    {'1': 'AGENT_SESSION_SCOPE_CUSTOM', '2': 4},
  ],
};

/// Descriptor for `AgentSessionScope`. Decode as a `google.protobuf.EnumDescriptorProto`.
final $typed_data.Uint8List agentSessionScopeDescriptor = $convert.base64Decode(
    'ChFBZ2VudFNlc3Npb25TY29wZRIjCh9BR0VOVF9TRVNTSU9OX1NDT1BFX1VOU1BFQ0lGSUVEEA'
    'ASHAoYQUdFTlRfU0VTU0lPTl9TQ09QRV9VU0VSEAESHAoYQUdFTlRfU0VTU0lPTl9TQ09QRV9D'
    'SEFUEAISHgoaQUdFTlRfU0VTU0lPTl9TQ09QRV9USFJFQUQQAxIeChpBR0VOVF9TRVNTSU9OX1'
    'NDT1BFX0NVU1RPTRAE');

@$core.Deprecated('Use agentChannelDescriptor instead')
const AgentChannel$json = {
  '1': 'AgentChannel',
  '2': [
    {'1': 'name', '3': 1, '4': 1, '5': 9, '8': {}, '10': 'name'},
    {'1': 'agent_name', '3': 2, '4': 1, '5': 9, '10': 'agentName'},
    {'1': 'agent_id', '3': 8, '4': 1, '5': 9, '10': 'agentId'},
    {
      '1': 'platform',
      '3': 3,
      '4': 1,
      '5': 14,
      '6': '.agents.v1.AgentChannelPlatform',
      '10': 'platform'
    },
    {'1': 'enabled', '3': 4, '4': 1, '5': 8, '10': 'enabled'},
    {
      '1': 'triggers',
      '3': 5,
      '4': 3,
      '5': 11,
      '6': '.agents.v1.AgentTrigger',
      '10': 'triggers'
    },
    {
      '1': 'delivery',
      '3': 6,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.AgentDelivery',
      '10': 'delivery'
    },
    {
      '1': 'session',
      '3': 7,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.AgentSessionBinding',
      '10': 'session'
    },
    {
      '1': 'telegram',
      '3': 10,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.TelegramChannelConfig',
      '10': 'telegram'
    },
    {
      '1': 'discord',
      '3': 12,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.DiscordChannelConfig',
      '10': 'discord'
    },
    {'1': 'model', '3': 11, '4': 1, '5': 9, '10': 'model'},
    {
      '1': 'metadata',
      '3': 20,
      '4': 3,
      '5': 11,
      '6': '.agents.v1.AgentChannel.MetadataEntry',
      '10': 'metadata'
    },
    {'1': 'workspace_id', '3': 100, '4': 1, '5': 9, '10': 'workspaceId'},
  ],
  '3': [AgentChannel_MetadataEntry$json],
};

@$core.Deprecated('Use agentChannelDescriptor instead')
const AgentChannel_MetadataEntry$json = {
  '1': 'MetadataEntry',
  '2': [
    {'1': 'key', '3': 1, '4': 1, '5': 9, '10': 'key'},
    {'1': 'value', '3': 2, '4': 1, '5': 9, '10': 'value'},
  ],
  '7': {'7': true},
};

/// Descriptor for `AgentChannel`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List agentChannelDescriptor = $convert.base64Decode(
    'CgxBZ2VudENoYW5uZWwSGwoEbmFtZRgBIAEoCUIH+kIEcgIQAVIEbmFtZRIdCgphZ2VudF9uYW'
    '1lGAIgASgJUglhZ2VudE5hbWUSGQoIYWdlbnRfaWQYCCABKAlSB2FnZW50SWQSOwoIcGxhdGZv'
    'cm0YAyABKA4yHy5hZ2VudHMudjEuQWdlbnRDaGFubmVsUGxhdGZvcm1SCHBsYXRmb3JtEhgKB2'
    'VuYWJsZWQYBCABKAhSB2VuYWJsZWQSMwoIdHJpZ2dlcnMYBSADKAsyFy5hZ2VudHMudjEuQWdl'
    'bnRUcmlnZ2VyUgh0cmlnZ2VycxI0CghkZWxpdmVyeRgGIAEoCzIYLmFnZW50cy52MS5BZ2VudE'
    'RlbGl2ZXJ5UghkZWxpdmVyeRI4CgdzZXNzaW9uGAcgASgLMh4uYWdlbnRzLnYxLkFnZW50U2Vz'
    'c2lvbkJpbmRpbmdSB3Nlc3Npb24SPAoIdGVsZWdyYW0YCiABKAsyIC5hZ2VudHMudjEuVGVsZW'
    'dyYW1DaGFubmVsQ29uZmlnUgh0ZWxlZ3JhbRI5CgdkaXNjb3JkGAwgASgLMh8uYWdlbnRzLnYx'
    'LkRpc2NvcmRDaGFubmVsQ29uZmlnUgdkaXNjb3JkEhQKBW1vZGVsGAsgASgJUgVtb2RlbBJBCg'
    'htZXRhZGF0YRgUIAMoCzIlLmFnZW50cy52MS5BZ2VudENoYW5uZWwuTWV0YWRhdGFFbnRyeVII'
    'bWV0YWRhdGESIQoMd29ya3NwYWNlX2lkGGQgASgJUgt3b3Jrc3BhY2VJZBo7Cg1NZXRhZGF0YU'
    'VudHJ5EhAKA2tleRgBIAEoCVIDa2V5EhQKBXZhbHVlGAIgASgJUgV2YWx1ZToCOAE=');

@$core.Deprecated('Use agentTriggerDescriptor instead')
const AgentTrigger$json = {
  '1': 'AgentTrigger',
  '2': [
    {
      '1': 'type',
      '3': 1,
      '4': 1,
      '5': 14,
      '6': '.agents.v1.AgentTriggerType',
      '10': 'type'
    },
    {'1': 'commands', '3': 2, '4': 3, '5': 9, '10': 'commands'},
    {'1': 'prefixes', '3': 3, '4': 3, '5': 9, '10': 'prefixes'},
    {'1': 'require_mention', '3': 4, '4': 1, '5': 8, '10': 'requireMention'},
  ],
};

/// Descriptor for `AgentTrigger`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List agentTriggerDescriptor = $convert.base64Decode(
    'CgxBZ2VudFRyaWdnZXISLwoEdHlwZRgBIAEoDjIbLmFnZW50cy52MS5BZ2VudFRyaWdnZXJUeX'
    'BlUgR0eXBlEhoKCGNvbW1hbmRzGAIgAygJUghjb21tYW5kcxIaCghwcmVmaXhlcxgDIAMoCVII'
    'cHJlZml4ZXMSJwoPcmVxdWlyZV9tZW50aW9uGAQgASgIUg5yZXF1aXJlTWVudGlvbg==');

@$core.Deprecated('Use agentDeliveryDescriptor instead')
const AgentDelivery$json = {
  '1': 'AgentDelivery',
  '2': [
    {
      '1': 'reply_mode',
      '3': 1,
      '4': 1,
      '5': 14,
      '6': '.agents.v1.AgentReplyMode',
      '10': 'replyMode'
    },
    {
      '1': 'streaming_enabled',
      '3': 2,
      '4': 1,
      '5': 8,
      '10': 'streamingEnabled'
    },
    {'1': 'send_typing', '3': 3, '4': 1, '5': 8, '10': 'sendTyping'},
  ],
};

/// Descriptor for `AgentDelivery`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List agentDeliveryDescriptor = $convert.base64Decode(
    'Cg1BZ2VudERlbGl2ZXJ5EjgKCnJlcGx5X21vZGUYASABKA4yGS5hZ2VudHMudjEuQWdlbnRSZX'
    'BseU1vZGVSCXJlcGx5TW9kZRIrChFzdHJlYW1pbmdfZW5hYmxlZBgCIAEoCFIQc3RyZWFtaW5n'
    'RW5hYmxlZBIfCgtzZW5kX3R5cGluZxgDIAEoCFIKc2VuZFR5cGluZw==');

@$core.Deprecated('Use agentSessionBindingDescriptor instead')
const AgentSessionBinding$json = {
  '1': 'AgentSessionBinding',
  '2': [
    {
      '1': 'scope',
      '3': 1,
      '4': 1,
      '5': 14,
      '6': '.agents.v1.AgentSessionScope',
      '10': 'scope'
    },
    {'1': 'key_template', '3': 2, '4': 1, '5': 9, '10': 'keyTemplate'},
  ],
};

/// Descriptor for `AgentSessionBinding`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List agentSessionBindingDescriptor = $convert.base64Decode(
    'ChNBZ2VudFNlc3Npb25CaW5kaW5nEjIKBXNjb3BlGAEgASgOMhwuYWdlbnRzLnYxLkFnZW50U2'
    'Vzc2lvblNjb3BlUgVzY29wZRIhCgxrZXlfdGVtcGxhdGUYAiABKAlSC2tleVRlbXBsYXRl');

@$core.Deprecated('Use telegramChannelConfigDescriptor instead')
const TelegramChannelConfig$json = {
  '1': 'TelegramChannelConfig',
  '2': [
    {'1': 'bot_token', '3': 1, '4': 1, '5': 9, '10': 'botToken'},
    {'1': 'webhook_url', '3': 2, '4': 1, '5': 9, '10': 'webhookUrl'},
    {'1': 'webhook_secret', '3': 3, '4': 1, '5': 9, '10': 'webhookSecret'},
    {'1': 'allowed_chat_ids', '3': 4, '4': 3, '5': 3, '10': 'allowedChatIds'},
    {'1': 'allowed_user_ids', '3': 5, '4': 3, '5': 3, '10': 'allowedUserIds'},
    {'1': 'debug', '3': 6, '4': 1, '5': 8, '10': 'debug'},
  ],
};

/// Descriptor for `TelegramChannelConfig`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List telegramChannelConfigDescriptor = $convert.base64Decode(
    'ChVUZWxlZ3JhbUNoYW5uZWxDb25maWcSGwoJYm90X3Rva2VuGAEgASgJUghib3RUb2tlbhIfCg'
    't3ZWJob29rX3VybBgCIAEoCVIKd2ViaG9va1VybBIlCg53ZWJob29rX3NlY3JldBgDIAEoCVIN'
    'd2ViaG9va1NlY3JldBIoChBhbGxvd2VkX2NoYXRfaWRzGAQgAygDUg5hbGxvd2VkQ2hhdElkcx'
    'IoChBhbGxvd2VkX3VzZXJfaWRzGAUgAygDUg5hbGxvd2VkVXNlcklkcxIUCgVkZWJ1ZxgGIAEo'
    'CFIFZGVidWc=');

@$core.Deprecated('Use discordChannelConfigDescriptor instead')
const DiscordChannelConfig$json = {
  '1': 'DiscordChannelConfig',
  '2': [
    {'1': 'bot_token', '3': 1, '4': 1, '5': 9, '10': 'botToken'},
    {'1': 'allowed_guild_ids', '3': 2, '4': 3, '5': 9, '10': 'allowedGuildIds'},
    {
      '1': 'allowed_channel_ids',
      '3': 3,
      '4': 3,
      '5': 9,
      '10': 'allowedChannelIds'
    },
    {'1': 'debug', '3': 4, '4': 1, '5': 8, '10': 'debug'},
  ],
};

/// Descriptor for `DiscordChannelConfig`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List discordChannelConfigDescriptor = $convert.base64Decode(
    'ChREaXNjb3JkQ2hhbm5lbENvbmZpZxIbCglib3RfdG9rZW4YASABKAlSCGJvdFRva2VuEioKEW'
    'FsbG93ZWRfZ3VpbGRfaWRzGAIgAygJUg9hbGxvd2VkR3VpbGRJZHMSLgoTYWxsb3dlZF9jaGFu'
    'bmVsX2lkcxgDIAMoCVIRYWxsb3dlZENoYW5uZWxJZHMSFAoFZGVidWcYBCABKAhSBWRlYnVn');
