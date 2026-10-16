// This is a generated file - do not edit.
//
// Generated from agents/v1/telegram.proto.

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

import 'package:protobuf/well_known_types/google/protobuf/timestamp.pbjson.dart'
    as $0;

@$core.Deprecated('Use telegramReceiveModeDescriptor instead')
const TelegramReceiveMode$json = {
  '1': 'TelegramReceiveMode',
  '2': [
    {'1': 'TELEGRAM_RECEIVE_MODE_UNSPECIFIED', '2': 0},
    {'1': 'TELEGRAM_RECEIVE_MODE_WEBHOOK', '2': 1},
    {'1': 'TELEGRAM_RECEIVE_MODE_LONG_POLLING', '2': 2},
  ],
};

/// Descriptor for `TelegramReceiveMode`. Decode as a `google.protobuf.EnumDescriptorProto`.
final $typed_data.Uint8List telegramReceiveModeDescriptor = $convert.base64Decode(
    'ChNUZWxlZ3JhbVJlY2VpdmVNb2RlEiUKIVRFTEVHUkFNX1JFQ0VJVkVfTU9ERV9VTlNQRUNJRk'
    'lFRBAAEiEKHVRFTEVHUkFNX1JFQ0VJVkVfTU9ERV9XRUJIT09LEAESJgoiVEVMRUdSQU1fUkVD'
    'RUlWRV9NT0RFX0xPTkdfUE9MTElORxAC');

@$core.Deprecated('Use telegramCredentialStateDescriptor instead')
const TelegramCredentialState$json = {
  '1': 'TelegramCredentialState',
  '2': [
    {'1': 'TELEGRAM_CREDENTIAL_STATE_UNSPECIFIED', '2': 0},
    {'1': 'TELEGRAM_CREDENTIAL_STATE_MISSING', '2': 1},
    {'1': 'TELEGRAM_CREDENTIAL_STATE_VALID', '2': 2},
    {'1': 'TELEGRAM_CREDENTIAL_STATE_INVALID', '2': 3},
  ],
};

/// Descriptor for `TelegramCredentialState`. Decode as a `google.protobuf.EnumDescriptorProto`.
final $typed_data.Uint8List telegramCredentialStateDescriptor = $convert.base64Decode(
    'ChdUZWxlZ3JhbUNyZWRlbnRpYWxTdGF0ZRIpCiVURUxFR1JBTV9DUkVERU5USUFMX1NUQVRFX1'
    'VOU1BFQ0lGSUVEEAASJQohVEVMRUdSQU1fQ1JFREVOVElBTF9TVEFURV9NSVNTSU5HEAESIwof'
    'VEVMRUdSQU1fQ1JFREVOVElBTF9TVEFURV9WQUxJRBACEiUKIVRFTEVHUkFNX0NSRURFTlRJQU'
    'xfU1RBVEVfSU5WQUxJRBAD');

@$core.Deprecated('Use telegramTriggerModeDescriptor instead')
const TelegramTriggerMode$json = {
  '1': 'TelegramTriggerMode',
  '2': [
    {'1': 'TELEGRAM_TRIGGER_MODE_UNSPECIFIED', '2': 0},
    {'1': 'TELEGRAM_TRIGGER_MODE_ALL', '2': 1},
    {'1': 'TELEGRAM_TRIGGER_MODE_MENTION', '2': 2},
    {'1': 'TELEGRAM_TRIGGER_MODE_COMMAND', '2': 3},
    {'1': 'TELEGRAM_TRIGGER_MODE_MENTION_OR_COMMAND', '2': 4},
  ],
};

/// Descriptor for `TelegramTriggerMode`. Decode as a `google.protobuf.EnumDescriptorProto`.
final $typed_data.Uint8List telegramTriggerModeDescriptor = $convert.base64Decode(
    'ChNUZWxlZ3JhbVRyaWdnZXJNb2RlEiUKIVRFTEVHUkFNX1RSSUdHRVJfTU9ERV9VTlNQRUNJRk'
    'lFRBAAEh0KGVRFTEVHUkFNX1RSSUdHRVJfTU9ERV9BTEwQARIhCh1URUxFR1JBTV9UUklHR0VS'
    'X01PREVfTUVOVElPThACEiEKHVRFTEVHUkFNX1RSSUdHRVJfTU9ERV9DT01NQU5EEAMSLAooVE'
    'VMRUdSQU1fVFJJR0dFUl9NT0RFX01FTlRJT05fT1JfQ09NTUFORBAE');

@$core.Deprecated('Use telegramSessionPolicyDescriptor instead')
const TelegramSessionPolicy$json = {
  '1': 'TelegramSessionPolicy',
  '2': [
    {'1': 'TELEGRAM_SESSION_POLICY_UNSPECIFIED', '2': 0},
    {'1': 'TELEGRAM_SESSION_POLICY_DESTINATION', '2': 1},
    {'1': 'TELEGRAM_SESSION_POLICY_USER', '2': 2},
  ],
};

/// Descriptor for `TelegramSessionPolicy`. Decode as a `google.protobuf.EnumDescriptorProto`.
final $typed_data.Uint8List telegramSessionPolicyDescriptor = $convert.base64Decode(
    'ChVUZWxlZ3JhbVNlc3Npb25Qb2xpY3kSJwojVEVMRUdSQU1fU0VTU0lPTl9QT0xJQ1lfVU5TUE'
    'VDSUZJRUQQABInCiNURUxFR1JBTV9TRVNTSU9OX1BPTElDWV9ERVNUSU5BVElPThABEiAKHFRF'
    'TEVHUkFNX1NFU1NJT05fUE9MSUNZX1VTRVIQAg==');

@$core.Deprecated('Use telegramReplyModeDescriptor instead')
const TelegramReplyMode$json = {
  '1': 'TelegramReplyMode',
  '2': [
    {'1': 'TELEGRAM_REPLY_MODE_UNSPECIFIED', '2': 0},
    {'1': 'TELEGRAM_REPLY_MODE_REPLY', '2': 1},
    {'1': 'TELEGRAM_REPLY_MODE_NEW_MESSAGE', '2': 2},
  ],
};

/// Descriptor for `TelegramReplyMode`. Decode as a `google.protobuf.EnumDescriptorProto`.
final $typed_data.Uint8List telegramReplyModeDescriptor = $convert.base64Decode(
    'ChFUZWxlZ3JhbVJlcGx5TW9kZRIjCh9URUxFR1JBTV9SRVBMWV9NT0RFX1VOU1BFQ0lGSUVEEA'
    'ASHQoZVEVMRUdSQU1fUkVQTFlfTU9ERV9SRVBMWRABEiMKH1RFTEVHUkFNX1JFUExZX01PREVf'
    'TkVXX01FU1NBR0UQAg==');

@$core.Deprecated('Use telegramWebhookStateDescriptor instead')
const TelegramWebhookState$json = {
  '1': 'TelegramWebhookState',
  '2': [
    {'1': 'TELEGRAM_WEBHOOK_STATE_UNSPECIFIED', '2': 0},
    {'1': 'TELEGRAM_WEBHOOK_STATE_NOT_APPLICABLE', '2': 1},
    {'1': 'TELEGRAM_WEBHOOK_STATE_REGISTERED', '2': 2},
    {'1': 'TELEGRAM_WEBHOOK_STATE_PENDING', '2': 3},
    {'1': 'TELEGRAM_WEBHOOK_STATE_FAILED', '2': 4},
  ],
};

/// Descriptor for `TelegramWebhookState`. Decode as a `google.protobuf.EnumDescriptorProto`.
final $typed_data.Uint8List telegramWebhookStateDescriptor = $convert.base64Decode(
    'ChRUZWxlZ3JhbVdlYmhvb2tTdGF0ZRImCiJURUxFR1JBTV9XRUJIT09LX1NUQVRFX1VOU1BFQ0'
    'lGSUVEEAASKQolVEVMRUdSQU1fV0VCSE9PS19TVEFURV9OT1RfQVBQTElDQUJMRRABEiUKIVRF'
    'TEVHUkFNX1dFQkhPT0tfU1RBVEVfUkVHSVNURVJFRBACEiIKHlRFTEVHUkFNX1dFQkhPT0tfU1'
    'RBVEVfUEVORElORxADEiEKHVRFTEVHUkFNX1dFQkhPT0tfU1RBVEVfRkFJTEVEEAQ=');

@$core.Deprecated('Use telegramProcessingStatusDescriptor instead')
const TelegramProcessingStatus$json = {
  '1': 'TelegramProcessingStatus',
  '2': [
    {'1': 'TELEGRAM_PROCESSING_STATUS_UNSPECIFIED', '2': 0},
    {'1': 'TELEGRAM_PROCESSING_STATUS_RECEIVED', '2': 1},
    {'1': 'TELEGRAM_PROCESSING_STATUS_PROCESSING', '2': 2},
    {'1': 'TELEGRAM_PROCESSING_STATUS_READY_TO_DELIVER', '2': 3},
    {'1': 'TELEGRAM_PROCESSING_STATUS_SUCCEEDED', '2': 4},
    {'1': 'TELEGRAM_PROCESSING_STATUS_FAILED', '2': 5},
    {'1': 'TELEGRAM_PROCESSING_STATUS_FAILED_UNCERTAIN', '2': 6},
  ],
};

/// Descriptor for `TelegramProcessingStatus`. Decode as a `google.protobuf.EnumDescriptorProto`.
final $typed_data.Uint8List telegramProcessingStatusDescriptor = $convert.base64Decode(
    'ChhUZWxlZ3JhbVByb2Nlc3NpbmdTdGF0dXMSKgomVEVMRUdSQU1fUFJPQ0VTU0lOR19TVEFUVV'
    'NfVU5TUEVDSUZJRUQQABInCiNURUxFR1JBTV9QUk9DRVNTSU5HX1NUQVRVU19SRUNFSVZFRBAB'
    'EikKJVRFTEVHUkFNX1BST0NFU1NJTkdfU1RBVFVTX1BST0NFU1NJTkcQAhIvCitURUxFR1JBTV'
    '9QUk9DRVNTSU5HX1NUQVRVU19SRUFEWV9UT19ERUxJVkVSEAMSKAokVEVMRUdSQU1fUFJPQ0VT'
    'U0lOR19TVEFUVVNfU1VDQ0VFREVEEAQSJQohVEVMRUdSQU1fUFJPQ0VTU0lOR19TVEFUVVNfRk'
    'FJTEVEEAUSLworVEVMRUdSQU1fUFJPQ0VTU0lOR19TVEFUVVNfRkFJTEVEX1VOQ0VSVEFJThAG');

@$core.Deprecated('Use telegramBotCapabilitiesDescriptor instead')
const TelegramBotCapabilities$json = {
  '1': 'TelegramBotCapabilities',
  '2': [
    {'1': 'can_join_groups', '3': 1, '4': 1, '5': 8, '10': 'canJoinGroups'},
    {
      '1': 'can_read_all_group_messages',
      '3': 2,
      '4': 1,
      '5': 8,
      '10': 'canReadAllGroupMessages'
    },
    {
      '1': 'supports_inline_queries',
      '3': 3,
      '4': 1,
      '5': 8,
      '10': 'supportsInlineQueries'
    },
  ],
};

/// Descriptor for `TelegramBotCapabilities`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List telegramBotCapabilitiesDescriptor = $convert.base64Decode(
    'ChdUZWxlZ3JhbUJvdENhcGFiaWxpdGllcxImCg9jYW5fam9pbl9ncm91cHMYASABKAhSDWNhbk'
    'pvaW5Hcm91cHMSPAobY2FuX3JlYWRfYWxsX2dyb3VwX21lc3NhZ2VzGAIgASgIUhdjYW5SZWFk'
    'QWxsR3JvdXBNZXNzYWdlcxI2ChdzdXBwb3J0c19pbmxpbmVfcXVlcmllcxgDIAEoCFIVc3VwcG'
    '9ydHNJbmxpbmVRdWVyaWVz');

@$core.Deprecated('Use telegramChannelDescriptor instead')
const TelegramChannel$json = {
  '1': 'TelegramChannel',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
    {'1': 'key', '3': 2, '4': 1, '5': 9, '8': {}, '10': 'key'},
    {'1': 'name', '3': 3, '4': 1, '5': 9, '10': 'name'},
    {'1': 'bot_id', '3': 4, '4': 1, '5': 9, '10': 'botId'},
    {'1': 'bot_username', '3': 5, '4': 1, '5': 9, '10': 'botUsername'},
    {
      '1': 'bot_capabilities',
      '3': 6,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.TelegramBotCapabilities',
      '10': 'botCapabilities'
    },
    {
      '1': 'receive_mode',
      '3': 7,
      '4': 1,
      '5': 14,
      '6': '.agents.v1.TelegramReceiveMode',
      '10': 'receiveMode'
    },
    {'1': 'inbound_enabled', '3': 8, '4': 1, '5': 8, '10': 'inboundEnabled'},
    {'1': 'outbound_enabled', '3': 9, '4': 1, '5': 8, '10': 'outboundEnabled'},
    {'1': 'revision', '3': 10, '4': 1, '5': 3, '10': 'revision'},
    {
      '1': 'created_at',
      '3': 11,
      '4': 1,
      '5': 11,
      '6': '.google.protobuf.Timestamp',
      '10': 'createdAt'
    },
    {
      '1': 'updated_at',
      '3': 12,
      '4': 1,
      '5': 11,
      '6': '.google.protobuf.Timestamp',
      '10': 'updatedAt'
    },
    {
      '1': 'credential_state',
      '3': 13,
      '4': 1,
      '5': 14,
      '6': '.agents.v1.TelegramCredentialState',
      '10': 'credentialState'
    },
    {
      '1': 'credential_updated_at',
      '3': 14,
      '4': 1,
      '5': 11,
      '6': '.google.protobuf.Timestamp',
      '10': 'credentialUpdatedAt'
    },
    {
      '1': 'webhook_secret_set',
      '3': 15,
      '4': 1,
      '5': 8,
      '10': 'webhookSecretSet'
    },
    {
      '1': 'last_credential_error',
      '3': 16,
      '4': 1,
      '5': 9,
      '10': 'lastCredentialError'
    },
    {'1': 'workspace_id', '3': 100, '4': 1, '5': 9, '10': 'workspaceId'},
  ],
};

/// Descriptor for `TelegramChannel`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List telegramChannelDescriptor = $convert.base64Decode(
    'Cg9UZWxlZ3JhbUNoYW5uZWwSDgoCaWQYASABKAlSAmlkEhkKA2tleRgCIAEoCUIH+kIEcgIQAV'
    'IDa2V5EhIKBG5hbWUYAyABKAlSBG5hbWUSFQoGYm90X2lkGAQgASgJUgVib3RJZBIhCgxib3Rf'
    'dXNlcm5hbWUYBSABKAlSC2JvdFVzZXJuYW1lEk0KEGJvdF9jYXBhYmlsaXRpZXMYBiABKAsyIi'
    '5hZ2VudHMudjEuVGVsZWdyYW1Cb3RDYXBhYmlsaXRpZXNSD2JvdENhcGFiaWxpdGllcxJBCgxy'
    'ZWNlaXZlX21vZGUYByABKA4yHi5hZ2VudHMudjEuVGVsZWdyYW1SZWNlaXZlTW9kZVILcmVjZW'
    'l2ZU1vZGUSJwoPaW5ib3VuZF9lbmFibGVkGAggASgIUg5pbmJvdW5kRW5hYmxlZBIpChBvdXRi'
    'b3VuZF9lbmFibGVkGAkgASgIUg9vdXRib3VuZEVuYWJsZWQSGgoIcmV2aXNpb24YCiABKANSCH'
    'JldmlzaW9uEjkKCmNyZWF0ZWRfYXQYCyABKAsyGi5nb29nbGUucHJvdG9idWYuVGltZXN0YW1w'
    'UgljcmVhdGVkQXQSOQoKdXBkYXRlZF9hdBgMIAEoCzIaLmdvb2dsZS5wcm90b2J1Zi5UaW1lc3'
    'RhbXBSCXVwZGF0ZWRBdBJNChBjcmVkZW50aWFsX3N0YXRlGA0gASgOMiIuYWdlbnRzLnYxLlRl'
    'bGVncmFtQ3JlZGVudGlhbFN0YXRlUg9jcmVkZW50aWFsU3RhdGUSTgoVY3JlZGVudGlhbF91cG'
    'RhdGVkX2F0GA4gASgLMhouZ29vZ2xlLnByb3RvYnVmLlRpbWVzdGFtcFITY3JlZGVudGlhbFVw'
    'ZGF0ZWRBdBIsChJ3ZWJob29rX3NlY3JldF9zZXQYDyABKAhSEHdlYmhvb2tTZWNyZXRTZXQSMg'
    'oVbGFzdF9jcmVkZW50aWFsX2Vycm9yGBAgASgJUhNsYXN0Q3JlZGVudGlhbEVycm9yEiEKDHdv'
    'cmtzcGFjZV9pZBhkIAEoCVILd29ya3NwYWNlSWQ=');

@$core.Deprecated('Use telegramDestinationConfigDescriptor instead')
const TelegramDestinationConfig$json = {
  '1': 'TelegramDestinationConfig',
  '2': [
    {'1': 'agent_id', '3': 1, '4': 1, '5': 9, '10': 'agentId'},
    {'1': 'model', '3': 2, '4': 1, '5': 9, '10': 'model'},
    {
      '1': 'selectable_agent_ids',
      '3': 3,
      '4': 3,
      '5': 9,
      '10': 'selectableAgentIds'
    },
    {
      '1': 'selectable_models',
      '3': 4,
      '4': 3,
      '5': 9,
      '10': 'selectableModels'
    },
    {
      '1': 'trigger_mode',
      '3': 5,
      '4': 1,
      '5': 14,
      '6': '.agents.v1.TelegramTriggerMode',
      '10': 'triggerMode'
    },
    {
      '1': 'session_policy',
      '3': 6,
      '4': 1,
      '5': 14,
      '6': '.agents.v1.TelegramSessionPolicy',
      '10': 'sessionPolicy'
    },
    {'1': 'allowed_user_ids', '3': 7, '4': 3, '5': 9, '10': 'allowedUserIds'},
    {
      '1': 'controller_user_ids',
      '3': 8,
      '4': 3,
      '5': 9,
      '10': 'controllerUserIds'
    },
    {
      '1': 'reply_mode',
      '3': 9,
      '4': 1,
      '5': 14,
      '6': '.agents.v1.TelegramReplyMode',
      '10': 'replyMode'
    },
    {
      '1': 'debug_default',
      '3': 10,
      '4': 1,
      '5': 8,
      '9': 0,
      '10': 'debugDefault',
      '17': true
    },
  ],
  '8': [
    {'1': '_debug_default'},
  ],
};

/// Descriptor for `TelegramDestinationConfig`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List telegramDestinationConfigDescriptor = $convert.base64Decode(
    'ChlUZWxlZ3JhbURlc3RpbmF0aW9uQ29uZmlnEhkKCGFnZW50X2lkGAEgASgJUgdhZ2VudElkEh'
    'QKBW1vZGVsGAIgASgJUgVtb2RlbBIwChRzZWxlY3RhYmxlX2FnZW50X2lkcxgDIAMoCVISc2Vs'
    'ZWN0YWJsZUFnZW50SWRzEisKEXNlbGVjdGFibGVfbW9kZWxzGAQgAygJUhBzZWxlY3RhYmxlTW'
    '9kZWxzEkEKDHRyaWdnZXJfbW9kZRgFIAEoDjIeLmFnZW50cy52MS5UZWxlZ3JhbVRyaWdnZXJN'
    'b2RlUgt0cmlnZ2VyTW9kZRJHCg5zZXNzaW9uX3BvbGljeRgGIAEoDjIgLmFnZW50cy52MS5UZW'
    'xlZ3JhbVNlc3Npb25Qb2xpY3lSDXNlc3Npb25Qb2xpY3kSKAoQYWxsb3dlZF91c2VyX2lkcxgH'
    'IAMoCVIOYWxsb3dlZFVzZXJJZHMSLgoTY29udHJvbGxlcl91c2VyX2lkcxgIIAMoCVIRY29udH'
    'JvbGxlclVzZXJJZHMSOwoKcmVwbHlfbW9kZRgJIAEoDjIcLmFnZW50cy52MS5UZWxlZ3JhbVJl'
    'cGx5TW9kZVIJcmVwbHlNb2RlEigKDWRlYnVnX2RlZmF1bHQYCiABKAhIAFIMZGVidWdEZWZhdW'
    'x0iAEBQhAKDl9kZWJ1Z19kZWZhdWx0');

@$core.Deprecated('Use telegramDestinationVerificationDescriptor instead')
const TelegramDestinationVerification$json = {
  '1': 'TelegramDestinationVerification',
  '2': [
    {'1': 'verified', '3': 1, '4': 1, '5': 8, '10': 'verified'},
    {
      '1': 'verified_at',
      '3': 2,
      '4': 1,
      '5': 11,
      '6': '.google.protobuf.Timestamp',
      '10': 'verifiedAt'
    },
    {
      '1': 'last_inbound_at',
      '3': 3,
      '4': 1,
      '5': 11,
      '6': '.google.protobuf.Timestamp',
      '10': 'lastInboundAt'
    },
    {
      '1': 'last_outbound_at',
      '3': 4,
      '4': 1,
      '5': 11,
      '6': '.google.protobuf.Timestamp',
      '10': 'lastOutboundAt'
    },
    {
      '1': 'last_outbound_error',
      '3': 5,
      '4': 1,
      '5': 9,
      '10': 'lastOutboundError'
    },
  ],
};

/// Descriptor for `TelegramDestinationVerification`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List telegramDestinationVerificationDescriptor = $convert.base64Decode(
    'Ch9UZWxlZ3JhbURlc3RpbmF0aW9uVmVyaWZpY2F0aW9uEhoKCHZlcmlmaWVkGAEgASgIUgh2ZX'
    'JpZmllZBI7Cgt2ZXJpZmllZF9hdBgCIAEoCzIaLmdvb2dsZS5wcm90b2J1Zi5UaW1lc3RhbXBS'
    'CnZlcmlmaWVkQXQSQgoPbGFzdF9pbmJvdW5kX2F0GAMgASgLMhouZ29vZ2xlLnByb3RvYnVmLl'
    'RpbWVzdGFtcFINbGFzdEluYm91bmRBdBJEChBsYXN0X291dGJvdW5kX2F0GAQgASgLMhouZ29v'
    'Z2xlLnByb3RvYnVmLlRpbWVzdGFtcFIObGFzdE91dGJvdW5kQXQSLgoTbGFzdF9vdXRib3VuZF'
    '9lcnJvchgFIAEoCVIRbGFzdE91dGJvdW5kRXJyb3I=');

@$core.Deprecated('Use telegramDestinationDescriptor instead')
const TelegramDestination$json = {
  '1': 'TelegramDestination',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
    {'1': 'key', '3': 2, '4': 1, '5': 9, '8': {}, '10': 'key'},
    {'1': 'name', '3': 3, '4': 1, '5': 9, '10': 'name'},
    {'1': 'channel_id', '3': 4, '4': 1, '5': 9, '8': {}, '10': 'channelId'},
    {'1': 'chat_id', '3': 5, '4': 1, '5': 9, '8': {}, '10': 'chatId'},
    {'1': 'message_thread_id', '3': 6, '4': 1, '5': 9, '10': 'messageThreadId'},
    {'1': 'inbound_enabled', '3': 7, '4': 1, '5': 8, '10': 'inboundEnabled'},
    {'1': 'outbound_enabled', '3': 8, '4': 1, '5': 8, '10': 'outboundEnabled'},
    {
      '1': 'config',
      '3': 9,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.TelegramDestinationConfig',
      '10': 'config'
    },
    {'1': 'revision', '3': 10, '4': 1, '5': 3, '10': 'revision'},
    {
      '1': 'created_at',
      '3': 11,
      '4': 1,
      '5': 11,
      '6': '.google.protobuf.Timestamp',
      '10': 'createdAt'
    },
    {
      '1': 'updated_at',
      '3': 12,
      '4': 1,
      '5': 11,
      '6': '.google.protobuf.Timestamp',
      '10': 'updatedAt'
    },
    {
      '1': 'verification',
      '3': 13,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.TelegramDestinationVerification',
      '10': 'verification'
    },
    {'1': 'workspace_id', '3': 100, '4': 1, '5': 9, '10': 'workspaceId'},
  ],
};

/// Descriptor for `TelegramDestination`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List telegramDestinationDescriptor = $convert.base64Decode(
    'ChNUZWxlZ3JhbURlc3RpbmF0aW9uEg4KAmlkGAEgASgJUgJpZBIZCgNrZXkYAiABKAlCB/pCBH'
    'ICEAFSA2tleRISCgRuYW1lGAMgASgJUgRuYW1lEiYKCmNoYW5uZWxfaWQYBCABKAlCB/pCBHIC'
    'EAFSCWNoYW5uZWxJZBIgCgdjaGF0X2lkGAUgASgJQgf6QgRyAhABUgZjaGF0SWQSKgoRbWVzc2'
    'FnZV90aHJlYWRfaWQYBiABKAlSD21lc3NhZ2VUaHJlYWRJZBInCg9pbmJvdW5kX2VuYWJsZWQY'
    'ByABKAhSDmluYm91bmRFbmFibGVkEikKEG91dGJvdW5kX2VuYWJsZWQYCCABKAhSD291dGJvdW'
    '5kRW5hYmxlZBI8CgZjb25maWcYCSABKAsyJC5hZ2VudHMudjEuVGVsZWdyYW1EZXN0aW5hdGlv'
    'bkNvbmZpZ1IGY29uZmlnEhoKCHJldmlzaW9uGAogASgDUghyZXZpc2lvbhI5CgpjcmVhdGVkX2'
    'F0GAsgASgLMhouZ29vZ2xlLnByb3RvYnVmLlRpbWVzdGFtcFIJY3JlYXRlZEF0EjkKCnVwZGF0'
    'ZWRfYXQYDCABKAsyGi5nb29nbGUucHJvdG9idWYuVGltZXN0YW1wUgl1cGRhdGVkQXQSTgoMdm'
    'VyaWZpY2F0aW9uGA0gASgLMiouYWdlbnRzLnYxLlRlbGVncmFtRGVzdGluYXRpb25WZXJpZmlj'
    'YXRpb25SDHZlcmlmaWNhdGlvbhIhCgx3b3Jrc3BhY2VfaWQYZCABKAlSC3dvcmtzcGFjZUlk');

@$core.Deprecated('Use listTelegramChannelsRequestDescriptor instead')
const ListTelegramChannelsRequest$json = {
  '1': 'ListTelegramChannelsRequest',
};

/// Descriptor for `ListTelegramChannelsRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listTelegramChannelsRequestDescriptor =
    $convert.base64Decode('ChtMaXN0VGVsZWdyYW1DaGFubmVsc1JlcXVlc3Q=');

@$core.Deprecated('Use listTelegramChannelsResponseDescriptor instead')
const ListTelegramChannelsResponse$json = {
  '1': 'ListTelegramChannelsResponse',
  '2': [
    {
      '1': 'channels',
      '3': 1,
      '4': 3,
      '5': 11,
      '6': '.agents.v1.TelegramChannel',
      '10': 'channels'
    },
  ],
};

/// Descriptor for `ListTelegramChannelsResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listTelegramChannelsResponseDescriptor =
    $convert.base64Decode(
        'ChxMaXN0VGVsZWdyYW1DaGFubmVsc1Jlc3BvbnNlEjYKCGNoYW5uZWxzGAEgAygLMhouYWdlbn'
        'RzLnYxLlRlbGVncmFtQ2hhbm5lbFIIY2hhbm5lbHM=');

@$core.Deprecated('Use getTelegramChannelRequestDescriptor instead')
const GetTelegramChannelRequest$json = {
  '1': 'GetTelegramChannelRequest',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '8': {}, '10': 'id'},
  ],
};

/// Descriptor for `GetTelegramChannelRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getTelegramChannelRequestDescriptor =
    $convert.base64Decode(
        'ChlHZXRUZWxlZ3JhbUNoYW5uZWxSZXF1ZXN0EhcKAmlkGAEgASgJQgf6QgRyAhABUgJpZA==');

@$core.Deprecated('Use getTelegramChannelResponseDescriptor instead')
const GetTelegramChannelResponse$json = {
  '1': 'GetTelegramChannelResponse',
  '2': [
    {
      '1': 'channel',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.TelegramChannel',
      '10': 'channel'
    },
  ],
};

/// Descriptor for `GetTelegramChannelResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getTelegramChannelResponseDescriptor =
    $convert.base64Decode(
        'ChpHZXRUZWxlZ3JhbUNoYW5uZWxSZXNwb25zZRI0CgdjaGFubmVsGAEgASgLMhouYWdlbnRzLn'
        'YxLlRlbGVncmFtQ2hhbm5lbFIHY2hhbm5lbA==');

@$core.Deprecated('Use createTelegramChannelRequestDescriptor instead')
const CreateTelegramChannelRequest$json = {
  '1': 'CreateTelegramChannelRequest',
  '2': [
    {
      '1': 'channel',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.TelegramChannel',
      '10': 'channel'
    },
    {'1': 'bot_token', '3': 2, '4': 1, '5': 9, '8': {}, '10': 'botToken'},
  ],
};

/// Descriptor for `CreateTelegramChannelRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List createTelegramChannelRequestDescriptor =
    $convert.base64Decode(
        'ChxDcmVhdGVUZWxlZ3JhbUNoYW5uZWxSZXF1ZXN0EjQKB2NoYW5uZWwYASABKAsyGi5hZ2VudH'
        'MudjEuVGVsZWdyYW1DaGFubmVsUgdjaGFubmVsEiQKCWJvdF90b2tlbhgCIAEoCUIH+kIEcgIQ'
        'AVIIYm90VG9rZW4=');

@$core.Deprecated('Use createTelegramChannelResponseDescriptor instead')
const CreateTelegramChannelResponse$json = {
  '1': 'CreateTelegramChannelResponse',
  '2': [
    {
      '1': 'channel',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.TelegramChannel',
      '10': 'channel'
    },
  ],
};

/// Descriptor for `CreateTelegramChannelResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List createTelegramChannelResponseDescriptor =
    $convert.base64Decode(
        'Ch1DcmVhdGVUZWxlZ3JhbUNoYW5uZWxSZXNwb25zZRI0CgdjaGFubmVsGAEgASgLMhouYWdlbn'
        'RzLnYxLlRlbGVncmFtQ2hhbm5lbFIHY2hhbm5lbA==');

@$core.Deprecated('Use updateTelegramChannelRequestDescriptor instead')
const UpdateTelegramChannelRequest$json = {
  '1': 'UpdateTelegramChannelRequest',
  '2': [
    {
      '1': 'channel',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.TelegramChannel',
      '10': 'channel'
    },
  ],
};

/// Descriptor for `UpdateTelegramChannelRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List updateTelegramChannelRequestDescriptor =
    $convert.base64Decode(
        'ChxVcGRhdGVUZWxlZ3JhbUNoYW5uZWxSZXF1ZXN0EjQKB2NoYW5uZWwYASABKAsyGi5hZ2VudH'
        'MudjEuVGVsZWdyYW1DaGFubmVsUgdjaGFubmVs');

@$core.Deprecated('Use updateTelegramChannelResponseDescriptor instead')
const UpdateTelegramChannelResponse$json = {
  '1': 'UpdateTelegramChannelResponse',
  '2': [
    {
      '1': 'channel',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.TelegramChannel',
      '10': 'channel'
    },
  ],
};

/// Descriptor for `UpdateTelegramChannelResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List updateTelegramChannelResponseDescriptor =
    $convert.base64Decode(
        'Ch1VcGRhdGVUZWxlZ3JhbUNoYW5uZWxSZXNwb25zZRI0CgdjaGFubmVsGAEgASgLMhouYWdlbn'
        'RzLnYxLlRlbGVncmFtQ2hhbm5lbFIHY2hhbm5lbA==');

@$core
    .Deprecated('Use rotateTelegramChannelCredentialRequestDescriptor instead')
const RotateTelegramChannelCredentialRequest$json = {
  '1': 'RotateTelegramChannelCredentialRequest',
  '2': [
    {'1': 'channel_id', '3': 1, '4': 1, '5': 9, '8': {}, '10': 'channelId'},
    {'1': 'bot_token', '3': 2, '4': 1, '5': 9, '8': {}, '10': 'botToken'},
  ],
};

/// Descriptor for `RotateTelegramChannelCredentialRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List rotateTelegramChannelCredentialRequestDescriptor =
    $convert.base64Decode(
        'CiZSb3RhdGVUZWxlZ3JhbUNoYW5uZWxDcmVkZW50aWFsUmVxdWVzdBImCgpjaGFubmVsX2lkGA'
        'EgASgJQgf6QgRyAhABUgljaGFubmVsSWQSJAoJYm90X3Rva2VuGAIgASgJQgf6QgRyAhABUghi'
        'b3RUb2tlbg==');

@$core
    .Deprecated('Use rotateTelegramChannelCredentialResponseDescriptor instead')
const RotateTelegramChannelCredentialResponse$json = {
  '1': 'RotateTelegramChannelCredentialResponse',
  '2': [
    {
      '1': 'channel',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.TelegramChannel',
      '10': 'channel'
    },
  ],
};

/// Descriptor for `RotateTelegramChannelCredentialResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List rotateTelegramChannelCredentialResponseDescriptor =
    $convert.base64Decode(
        'CidSb3RhdGVUZWxlZ3JhbUNoYW5uZWxDcmVkZW50aWFsUmVzcG9uc2USNAoHY2hhbm5lbBgBIA'
        'EoCzIaLmFnZW50cy52MS5UZWxlZ3JhbUNoYW5uZWxSB2NoYW5uZWw=');

@$core.Deprecated('Use setTelegramChannelEnabledRequestDescriptor instead')
const SetTelegramChannelEnabledRequest$json = {
  '1': 'SetTelegramChannelEnabledRequest',
  '2': [
    {'1': 'channel_id', '3': 1, '4': 1, '5': 9, '8': {}, '10': 'channelId'},
    {'1': 'revision', '3': 2, '4': 1, '5': 3, '10': 'revision'},
    {'1': 'inbound_enabled', '3': 3, '4': 1, '5': 8, '10': 'inboundEnabled'},
    {'1': 'outbound_enabled', '3': 4, '4': 1, '5': 8, '10': 'outboundEnabled'},
  ],
};

/// Descriptor for `SetTelegramChannelEnabledRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List setTelegramChannelEnabledRequestDescriptor =
    $convert.base64Decode(
        'CiBTZXRUZWxlZ3JhbUNoYW5uZWxFbmFibGVkUmVxdWVzdBImCgpjaGFubmVsX2lkGAEgASgJQg'
        'f6QgRyAhABUgljaGFubmVsSWQSGgoIcmV2aXNpb24YAiABKANSCHJldmlzaW9uEicKD2luYm91'
        'bmRfZW5hYmxlZBgDIAEoCFIOaW5ib3VuZEVuYWJsZWQSKQoQb3V0Ym91bmRfZW5hYmxlZBgEIA'
        'EoCFIPb3V0Ym91bmRFbmFibGVk');

@$core.Deprecated('Use setTelegramChannelEnabledResponseDescriptor instead')
const SetTelegramChannelEnabledResponse$json = {
  '1': 'SetTelegramChannelEnabledResponse',
  '2': [
    {
      '1': 'channel',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.TelegramChannel',
      '10': 'channel'
    },
    {'1': 'warnings', '3': 2, '4': 3, '5': 9, '10': 'warnings'},
  ],
};

/// Descriptor for `SetTelegramChannelEnabledResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List setTelegramChannelEnabledResponseDescriptor =
    $convert.base64Decode(
        'CiFTZXRUZWxlZ3JhbUNoYW5uZWxFbmFibGVkUmVzcG9uc2USNAoHY2hhbm5lbBgBIAEoCzIaLm'
        'FnZW50cy52MS5UZWxlZ3JhbUNoYW5uZWxSB2NoYW5uZWwSGgoId2FybmluZ3MYAiADKAlSCHdh'
        'cm5pbmdz');

@$core.Deprecated('Use deleteTelegramChannelRequestDescriptor instead')
const DeleteTelegramChannelRequest$json = {
  '1': 'DeleteTelegramChannelRequest',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '8': {}, '10': 'id'},
  ],
};

/// Descriptor for `DeleteTelegramChannelRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List deleteTelegramChannelRequestDescriptor =
    $convert.base64Decode(
        'ChxEZWxldGVUZWxlZ3JhbUNoYW5uZWxSZXF1ZXN0EhcKAmlkGAEgASgJQgf6QgRyAhABUgJpZA'
        '==');

@$core.Deprecated('Use deleteTelegramChannelResponseDescriptor instead')
const DeleteTelegramChannelResponse$json = {
  '1': 'DeleteTelegramChannelResponse',
};

/// Descriptor for `DeleteTelegramChannelResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List deleteTelegramChannelResponseDescriptor =
    $convert.base64Decode('Ch1EZWxldGVUZWxlZ3JhbUNoYW5uZWxSZXNwb25zZQ==');

@$core.Deprecated('Use getTelegramChannelStatusRequestDescriptor instead')
const GetTelegramChannelStatusRequest$json = {
  '1': 'GetTelegramChannelStatusRequest',
  '2': [
    {'1': 'channel_id', '3': 1, '4': 1, '5': 9, '8': {}, '10': 'channelId'},
  ],
};

/// Descriptor for `GetTelegramChannelStatusRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getTelegramChannelStatusRequestDescriptor =
    $convert.base64Decode(
        'Ch9HZXRUZWxlZ3JhbUNoYW5uZWxTdGF0dXNSZXF1ZXN0EiYKCmNoYW5uZWxfaWQYASABKAlCB/'
        'pCBHICEAFSCWNoYW5uZWxJZA==');

@$core.Deprecated('Use telegramChannelStatusDescriptor instead')
const TelegramChannelStatus$json = {
  '1': 'TelegramChannelStatus',
  '2': [
    {'1': 'channel_id', '3': 1, '4': 1, '5': 9, '10': 'channelId'},
    {'1': 'inbound_desired', '3': 2, '4': 1, '5': 8, '10': 'inboundDesired'},
    {'1': 'outbound_desired', '3': 3, '4': 1, '5': 8, '10': 'outboundDesired'},
    {
      '1': 'credential_state',
      '3': 4,
      '4': 1,
      '5': 14,
      '6': '.agents.v1.TelegramCredentialState',
      '10': 'credentialState'
    },
    {
      '1': 'receive_mode',
      '3': 5,
      '4': 1,
      '5': 14,
      '6': '.agents.v1.TelegramReceiveMode',
      '10': 'receiveMode'
    },
    {
      '1': 'inbound_destination_count',
      '3': 6,
      '4': 1,
      '5': 5,
      '10': 'inboundDestinationCount'
    },
    {
      '1': 'outbound_destination_count',
      '3': 7,
      '4': 1,
      '5': 5,
      '10': 'outboundDestinationCount'
    },
    {'1': 'blockers', '3': 8, '4': 3, '5': 9, '10': 'blockers'},
    {'1': 'warnings', '3': 9, '4': 3, '5': 9, '10': 'warnings'},
    {
      '1': 'last_credential_error',
      '3': 10,
      '4': 1,
      '5': 9,
      '10': 'lastCredentialError'
    },
    {
      '1': 'webhook_state',
      '3': 11,
      '4': 1,
      '5': 14,
      '6': '.agents.v1.TelegramWebhookState',
      '10': 'webhookState'
    },
    {'1': 'webhook_url', '3': 12, '4': 1, '5': 9, '10': 'webhookUrl'},
    {
      '1': 'last_webhook_error',
      '3': 13,
      '4': 1,
      '5': 9,
      '10': 'lastWebhookError'
    },
    {
      '1': 'last_reconciled_at',
      '3': 14,
      '4': 1,
      '5': 11,
      '6': '.google.protobuf.Timestamp',
      '10': 'lastReconciledAt'
    },
    {'1': 'polling_leader', '3': 15, '4': 1, '5': 8, '10': 'pollingLeader'},
    {'1': 'queue_ready', '3': 16, '4': 1, '5': 8, '10': 'queueReady'},
    {
      '1': 'last_fetched_update_id',
      '3': 17,
      '4': 1,
      '5': 3,
      '10': 'lastFetchedUpdateId'
    },
    {
      '1': 'last_accepted_update_id',
      '3': 18,
      '4': 1,
      '5': 3,
      '10': 'lastAcceptedUpdateId'
    },
    {
      '1': 'last_polled_at',
      '3': 19,
      '4': 1,
      '5': 11,
      '6': '.google.protobuf.Timestamp',
      '10': 'lastPolledAt'
    },
    {
      '1': 'last_receive_error',
      '3': 20,
      '4': 1,
      '5': 9,
      '10': 'lastReceiveError'
    },
  ],
};

/// Descriptor for `TelegramChannelStatus`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List telegramChannelStatusDescriptor = $convert.base64Decode(
    'ChVUZWxlZ3JhbUNoYW5uZWxTdGF0dXMSHQoKY2hhbm5lbF9pZBgBIAEoCVIJY2hhbm5lbElkEi'
    'cKD2luYm91bmRfZGVzaXJlZBgCIAEoCFIOaW5ib3VuZERlc2lyZWQSKQoQb3V0Ym91bmRfZGVz'
    'aXJlZBgDIAEoCFIPb3V0Ym91bmREZXNpcmVkEk0KEGNyZWRlbnRpYWxfc3RhdGUYBCABKA4yIi'
    '5hZ2VudHMudjEuVGVsZWdyYW1DcmVkZW50aWFsU3RhdGVSD2NyZWRlbnRpYWxTdGF0ZRJBCgxy'
    'ZWNlaXZlX21vZGUYBSABKA4yHi5hZ2VudHMudjEuVGVsZWdyYW1SZWNlaXZlTW9kZVILcmVjZW'
    'l2ZU1vZGUSOgoZaW5ib3VuZF9kZXN0aW5hdGlvbl9jb3VudBgGIAEoBVIXaW5ib3VuZERlc3Rp'
    'bmF0aW9uQ291bnQSPAoab3V0Ym91bmRfZGVzdGluYXRpb25fY291bnQYByABKAVSGG91dGJvdW'
    '5kRGVzdGluYXRpb25Db3VudBIaCghibG9ja2VycxgIIAMoCVIIYmxvY2tlcnMSGgoId2Fybmlu'
    'Z3MYCSADKAlSCHdhcm5pbmdzEjIKFWxhc3RfY3JlZGVudGlhbF9lcnJvchgKIAEoCVITbGFzdE'
    'NyZWRlbnRpYWxFcnJvchJECg13ZWJob29rX3N0YXRlGAsgASgOMh8uYWdlbnRzLnYxLlRlbGVn'
    'cmFtV2ViaG9va1N0YXRlUgx3ZWJob29rU3RhdGUSHwoLd2ViaG9va191cmwYDCABKAlSCndlYm'
    'hvb2tVcmwSLAoSbGFzdF93ZWJob29rX2Vycm9yGA0gASgJUhBsYXN0V2ViaG9va0Vycm9yEkgK'
    'Emxhc3RfcmVjb25jaWxlZF9hdBgOIAEoCzIaLmdvb2dsZS5wcm90b2J1Zi5UaW1lc3RhbXBSEG'
    'xhc3RSZWNvbmNpbGVkQXQSJQoOcG9sbGluZ19sZWFkZXIYDyABKAhSDXBvbGxpbmdMZWFkZXIS'
    'HwoLcXVldWVfcmVhZHkYECABKAhSCnF1ZXVlUmVhZHkSMwoWbGFzdF9mZXRjaGVkX3VwZGF0ZV'
    '9pZBgRIAEoA1ITbGFzdEZldGNoZWRVcGRhdGVJZBI1ChdsYXN0X2FjY2VwdGVkX3VwZGF0ZV9p'
    'ZBgSIAEoA1IUbGFzdEFjY2VwdGVkVXBkYXRlSWQSQAoObGFzdF9wb2xsZWRfYXQYEyABKAsyGi'
    '5nb29nbGUucHJvdG9idWYuVGltZXN0YW1wUgxsYXN0UG9sbGVkQXQSLAoSbGFzdF9yZWNlaXZl'
    'X2Vycm9yGBQgASgJUhBsYXN0UmVjZWl2ZUVycm9y');

@$core.Deprecated('Use getTelegramChannelStatusResponseDescriptor instead')
const GetTelegramChannelStatusResponse$json = {
  '1': 'GetTelegramChannelStatusResponse',
  '2': [
    {
      '1': 'status',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.TelegramChannelStatus',
      '10': 'status'
    },
  ],
};

/// Descriptor for `GetTelegramChannelStatusResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getTelegramChannelStatusResponseDescriptor =
    $convert.base64Decode(
        'CiBHZXRUZWxlZ3JhbUNoYW5uZWxTdGF0dXNSZXNwb25zZRI4CgZzdGF0dXMYASABKAsyIC5hZ2'
        'VudHMudjEuVGVsZWdyYW1DaGFubmVsU3RhdHVzUgZzdGF0dXM=');

@$core.Deprecated('Use listTelegramDestinationsRequestDescriptor instead')
const ListTelegramDestinationsRequest$json = {
  '1': 'ListTelegramDestinationsRequest',
  '2': [
    {'1': 'channel_id', '3': 1, '4': 1, '5': 9, '10': 'channelId'},
  ],
};

/// Descriptor for `ListTelegramDestinationsRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listTelegramDestinationsRequestDescriptor =
    $convert.base64Decode(
        'Ch9MaXN0VGVsZWdyYW1EZXN0aW5hdGlvbnNSZXF1ZXN0Eh0KCmNoYW5uZWxfaWQYASABKAlSCW'
        'NoYW5uZWxJZA==');

@$core.Deprecated('Use listTelegramDestinationsResponseDescriptor instead')
const ListTelegramDestinationsResponse$json = {
  '1': 'ListTelegramDestinationsResponse',
  '2': [
    {
      '1': 'destinations',
      '3': 1,
      '4': 3,
      '5': 11,
      '6': '.agents.v1.TelegramDestination',
      '10': 'destinations'
    },
  ],
};

/// Descriptor for `ListTelegramDestinationsResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listTelegramDestinationsResponseDescriptor =
    $convert.base64Decode(
        'CiBMaXN0VGVsZWdyYW1EZXN0aW5hdGlvbnNSZXNwb25zZRJCCgxkZXN0aW5hdGlvbnMYASADKA'
        'syHi5hZ2VudHMudjEuVGVsZWdyYW1EZXN0aW5hdGlvblIMZGVzdGluYXRpb25z');

@$core.Deprecated('Use getTelegramDestinationRequestDescriptor instead')
const GetTelegramDestinationRequest$json = {
  '1': 'GetTelegramDestinationRequest',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '8': {}, '10': 'id'},
  ],
};

/// Descriptor for `GetTelegramDestinationRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getTelegramDestinationRequestDescriptor =
    $convert.base64Decode(
        'Ch1HZXRUZWxlZ3JhbURlc3RpbmF0aW9uUmVxdWVzdBIXCgJpZBgBIAEoCUIH+kIEcgIQAVICaW'
        'Q=');

@$core.Deprecated('Use getTelegramDestinationResponseDescriptor instead')
const GetTelegramDestinationResponse$json = {
  '1': 'GetTelegramDestinationResponse',
  '2': [
    {
      '1': 'destination',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.TelegramDestination',
      '10': 'destination'
    },
  ],
};

/// Descriptor for `GetTelegramDestinationResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getTelegramDestinationResponseDescriptor =
    $convert.base64Decode(
        'Ch5HZXRUZWxlZ3JhbURlc3RpbmF0aW9uUmVzcG9uc2USQAoLZGVzdGluYXRpb24YASABKAsyHi'
        '5hZ2VudHMudjEuVGVsZWdyYW1EZXN0aW5hdGlvblILZGVzdGluYXRpb24=');

@$core.Deprecated('Use createTelegramDestinationRequestDescriptor instead')
const CreateTelegramDestinationRequest$json = {
  '1': 'CreateTelegramDestinationRequest',
  '2': [
    {
      '1': 'destination',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.TelegramDestination',
      '10': 'destination'
    },
  ],
};

/// Descriptor for `CreateTelegramDestinationRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List createTelegramDestinationRequestDescriptor =
    $convert.base64Decode(
        'CiBDcmVhdGVUZWxlZ3JhbURlc3RpbmF0aW9uUmVxdWVzdBJACgtkZXN0aW5hdGlvbhgBIAEoCz'
        'IeLmFnZW50cy52MS5UZWxlZ3JhbURlc3RpbmF0aW9uUgtkZXN0aW5hdGlvbg==');

@$core.Deprecated('Use createTelegramDestinationResponseDescriptor instead')
const CreateTelegramDestinationResponse$json = {
  '1': 'CreateTelegramDestinationResponse',
  '2': [
    {
      '1': 'destination',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.TelegramDestination',
      '10': 'destination'
    },
  ],
};

/// Descriptor for `CreateTelegramDestinationResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List createTelegramDestinationResponseDescriptor =
    $convert.base64Decode(
        'CiFDcmVhdGVUZWxlZ3JhbURlc3RpbmF0aW9uUmVzcG9uc2USQAoLZGVzdGluYXRpb24YASABKA'
        'syHi5hZ2VudHMudjEuVGVsZWdyYW1EZXN0aW5hdGlvblILZGVzdGluYXRpb24=');

@$core.Deprecated('Use updateTelegramDestinationRequestDescriptor instead')
const UpdateTelegramDestinationRequest$json = {
  '1': 'UpdateTelegramDestinationRequest',
  '2': [
    {
      '1': 'destination',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.TelegramDestination',
      '10': 'destination'
    },
  ],
};

/// Descriptor for `UpdateTelegramDestinationRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List updateTelegramDestinationRequestDescriptor =
    $convert.base64Decode(
        'CiBVcGRhdGVUZWxlZ3JhbURlc3RpbmF0aW9uUmVxdWVzdBJACgtkZXN0aW5hdGlvbhgBIAEoCz'
        'IeLmFnZW50cy52MS5UZWxlZ3JhbURlc3RpbmF0aW9uUgtkZXN0aW5hdGlvbg==');

@$core.Deprecated('Use updateTelegramDestinationResponseDescriptor instead')
const UpdateTelegramDestinationResponse$json = {
  '1': 'UpdateTelegramDestinationResponse',
  '2': [
    {
      '1': 'destination',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.TelegramDestination',
      '10': 'destination'
    },
  ],
};

/// Descriptor for `UpdateTelegramDestinationResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List updateTelegramDestinationResponseDescriptor =
    $convert.base64Decode(
        'CiFVcGRhdGVUZWxlZ3JhbURlc3RpbmF0aW9uUmVzcG9uc2USQAoLZGVzdGluYXRpb24YASABKA'
        'syHi5hZ2VudHMudjEuVGVsZWdyYW1EZXN0aW5hdGlvblILZGVzdGluYXRpb24=');

@$core.Deprecated('Use deleteTelegramDestinationRequestDescriptor instead')
const DeleteTelegramDestinationRequest$json = {
  '1': 'DeleteTelegramDestinationRequest',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '8': {}, '10': 'id'},
  ],
};

/// Descriptor for `DeleteTelegramDestinationRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List deleteTelegramDestinationRequestDescriptor =
    $convert.base64Decode(
        'CiBEZWxldGVUZWxlZ3JhbURlc3RpbmF0aW9uUmVxdWVzdBIXCgJpZBgBIAEoCUIH+kIEcgIQAV'
        'ICaWQ=');

@$core.Deprecated('Use deleteTelegramDestinationResponseDescriptor instead')
const DeleteTelegramDestinationResponse$json = {
  '1': 'DeleteTelegramDestinationResponse',
};

/// Descriptor for `DeleteTelegramDestinationResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List deleteTelegramDestinationResponseDescriptor =
    $convert.base64Decode('CiFEZWxldGVUZWxlZ3JhbURlc3RpbmF0aW9uUmVzcG9uc2U=');

@$core.Deprecated('Use sendTelegramTestMessageRequestDescriptor instead')
const SendTelegramTestMessageRequest$json = {
  '1': 'SendTelegramTestMessageRequest',
  '2': [
    {
      '1': 'destination_id',
      '3': 1,
      '4': 1,
      '5': 9,
      '8': {},
      '10': 'destinationId'
    },
    {'1': 'text', '3': 2, '4': 1, '5': 9, '10': 'text'},
  ],
};

/// Descriptor for `SendTelegramTestMessageRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List sendTelegramTestMessageRequestDescriptor =
    $convert.base64Decode(
        'Ch5TZW5kVGVsZWdyYW1UZXN0TWVzc2FnZVJlcXVlc3QSLgoOZGVzdGluYXRpb25faWQYASABKA'
        'lCB/pCBHICEAFSDWRlc3RpbmF0aW9uSWQSEgoEdGV4dBgCIAEoCVIEdGV4dA==');

@$core.Deprecated('Use sendTelegramTestMessageResponseDescriptor instead')
const SendTelegramTestMessageResponse$json = {
  '1': 'SendTelegramTestMessageResponse',
  '2': [
    {
      '1': 'destination',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.TelegramDestination',
      '10': 'destination'
    },
    {'1': 'message_ids', '3': 2, '4': 3, '5': 9, '10': 'messageIds'},
  ],
};

/// Descriptor for `SendTelegramTestMessageResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List sendTelegramTestMessageResponseDescriptor =
    $convert.base64Decode(
        'Ch9TZW5kVGVsZWdyYW1UZXN0TWVzc2FnZVJlc3BvbnNlEkAKC2Rlc3RpbmF0aW9uGAEgASgLMh'
        '4uYWdlbnRzLnYxLlRlbGVncmFtRGVzdGluYXRpb25SC2Rlc3RpbmF0aW9uEh8KC21lc3NhZ2Vf'
        'aWRzGAIgAygJUgptZXNzYWdlSWRz');

@$core.Deprecated('Use telegramSettingsDescriptor instead')
const TelegramSettings$json = {
  '1': 'TelegramSettings',
  '2': [
    {'1': 'webhook_base_url', '3': 1, '4': 1, '5': 9, '10': 'webhookBaseUrl'},
    {
      '1': 'updated_at',
      '3': 2,
      '4': 1,
      '5': 11,
      '6': '.google.protobuf.Timestamp',
      '10': 'updatedAt'
    },
  ],
};

/// Descriptor for `TelegramSettings`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List telegramSettingsDescriptor = $convert.base64Decode(
    'ChBUZWxlZ3JhbVNldHRpbmdzEigKEHdlYmhvb2tfYmFzZV91cmwYASABKAlSDndlYmhvb2tCYX'
    'NlVXJsEjkKCnVwZGF0ZWRfYXQYAiABKAsyGi5nb29nbGUucHJvdG9idWYuVGltZXN0YW1wUgl1'
    'cGRhdGVkQXQ=');

@$core.Deprecated('Use getTelegramSettingsRequestDescriptor instead')
const GetTelegramSettingsRequest$json = {
  '1': 'GetTelegramSettingsRequest',
};

/// Descriptor for `GetTelegramSettingsRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getTelegramSettingsRequestDescriptor =
    $convert.base64Decode('ChpHZXRUZWxlZ3JhbVNldHRpbmdzUmVxdWVzdA==');

@$core.Deprecated('Use getTelegramSettingsResponseDescriptor instead')
const GetTelegramSettingsResponse$json = {
  '1': 'GetTelegramSettingsResponse',
  '2': [
    {
      '1': 'settings',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.TelegramSettings',
      '10': 'settings'
    },
  ],
};

/// Descriptor for `GetTelegramSettingsResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getTelegramSettingsResponseDescriptor =
    $convert.base64Decode(
        'ChtHZXRUZWxlZ3JhbVNldHRpbmdzUmVzcG9uc2USNwoIc2V0dGluZ3MYASABKAsyGy5hZ2VudH'
        'MudjEuVGVsZWdyYW1TZXR0aW5nc1IIc2V0dGluZ3M=');

@$core.Deprecated('Use updateTelegramSettingsRequestDescriptor instead')
const UpdateTelegramSettingsRequest$json = {
  '1': 'UpdateTelegramSettingsRequest',
  '2': [
    {
      '1': 'settings',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.TelegramSettings',
      '10': 'settings'
    },
  ],
};

/// Descriptor for `UpdateTelegramSettingsRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List updateTelegramSettingsRequestDescriptor =
    $convert.base64Decode(
        'Ch1VcGRhdGVUZWxlZ3JhbVNldHRpbmdzUmVxdWVzdBI3CghzZXR0aW5ncxgBIAEoCzIbLmFnZW'
        '50cy52MS5UZWxlZ3JhbVNldHRpbmdzUghzZXR0aW5ncw==');

@$core.Deprecated('Use updateTelegramSettingsResponseDescriptor instead')
const UpdateTelegramSettingsResponse$json = {
  '1': 'UpdateTelegramSettingsResponse',
  '2': [
    {
      '1': 'settings',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.TelegramSettings',
      '10': 'settings'
    },
  ],
};

/// Descriptor for `UpdateTelegramSettingsResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List updateTelegramSettingsResponseDescriptor =
    $convert.base64Decode(
        'Ch5VcGRhdGVUZWxlZ3JhbVNldHRpbmdzUmVzcG9uc2USNwoIc2V0dGluZ3MYASABKAsyGy5hZ2'
        'VudHMudjEuVGVsZWdyYW1TZXR0aW5nc1IIc2V0dGluZ3M=');

@$core.Deprecated('Use telegramDeliverySegmentDescriptor instead')
const TelegramDeliverySegment$json = {
  '1': 'TelegramDeliverySegment',
  '2': [
    {'1': 'index', '3': 1, '4': 1, '5': 5, '10': 'index'},
    {'1': 'text', '3': 2, '4': 1, '5': 9, '10': 'text'},
    {'1': 'status', '3': 3, '4': 1, '5': 9, '10': 'status'},
    {'1': 'message_id', '3': 4, '4': 1, '5': 9, '10': 'messageId'},
    {'1': 'error', '3': 5, '4': 1, '5': 9, '10': 'error'},
  ],
};

/// Descriptor for `TelegramDeliverySegment`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List telegramDeliverySegmentDescriptor = $convert.base64Decode(
    'ChdUZWxlZ3JhbURlbGl2ZXJ5U2VnbWVudBIUCgVpbmRleBgBIAEoBVIFaW5kZXgSEgoEdGV4dB'
    'gCIAEoCVIEdGV4dBIWCgZzdGF0dXMYAyABKAlSBnN0YXR1cxIdCgptZXNzYWdlX2lkGAQgASgJ'
    'UgltZXNzYWdlSWQSFAoFZXJyb3IYBSABKAlSBWVycm9y');

@$core.Deprecated('Use telegramProcessingRecordDescriptor instead')
const TelegramProcessingRecord$json = {
  '1': 'TelegramProcessingRecord',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
    {'1': 'workspace_id', '3': 2, '4': 1, '5': 9, '10': 'workspaceId'},
    {'1': 'channel_id', '3': 3, '4': 1, '5': 9, '10': 'channelId'},
    {'1': 'destination_id', '3': 4, '4': 1, '5': 9, '10': 'destinationId'},
    {'1': 'update_id', '3': 5, '4': 1, '5': 3, '10': 'updateId'},
    {
      '1': 'status',
      '3': 6,
      '4': 1,
      '5': 14,
      '6': '.agents.v1.TelegramProcessingStatus',
      '10': 'status'
    },
    {'1': 'invocation_id', '3': 7, '4': 1, '5': 9, '10': 'invocationId'},
    {'1': 'attempts', '3': 8, '4': 1, '5': 5, '10': 'attempts'},
    {
      '1': 'destination_revision',
      '3': 9,
      '4': 1,
      '5': 3,
      '10': 'destinationRevision'
    },
    {'1': 'error', '3': 10, '4': 1, '5': 9, '10': 'error'},
    {'1': 'output', '3': 11, '4': 1, '5': 9, '10': 'output'},
    {
      '1': 'segments',
      '3': 12,
      '4': 3,
      '5': 11,
      '6': '.agents.v1.TelegramDeliverySegment',
      '10': 'segments'
    },
    {
      '1': 'created_at',
      '3': 13,
      '4': 1,
      '5': 11,
      '6': '.google.protobuf.Timestamp',
      '10': 'createdAt'
    },
    {
      '1': 'updated_at',
      '3': 14,
      '4': 1,
      '5': 11,
      '6': '.google.protobuf.Timestamp',
      '10': 'updatedAt'
    },
    {
      '1': 'expires_at',
      '3': 15,
      '4': 1,
      '5': 11,
      '6': '.google.protobuf.Timestamp',
      '10': 'expiresAt'
    },
    {'1': 'dead_lettered', '3': 16, '4': 1, '5': 8, '10': 'deadLettered'},
  ],
};

/// Descriptor for `TelegramProcessingRecord`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List telegramProcessingRecordDescriptor = $convert.base64Decode(
    'ChhUZWxlZ3JhbVByb2Nlc3NpbmdSZWNvcmQSDgoCaWQYASABKAlSAmlkEiEKDHdvcmtzcGFjZV'
    '9pZBgCIAEoCVILd29ya3NwYWNlSWQSHQoKY2hhbm5lbF9pZBgDIAEoCVIJY2hhbm5lbElkEiUK'
    'DmRlc3RpbmF0aW9uX2lkGAQgASgJUg1kZXN0aW5hdGlvbklkEhsKCXVwZGF0ZV9pZBgFIAEoA1'
    'IIdXBkYXRlSWQSOwoGc3RhdHVzGAYgASgOMiMuYWdlbnRzLnYxLlRlbGVncmFtUHJvY2Vzc2lu'
    'Z1N0YXR1c1IGc3RhdHVzEiMKDWludm9jYXRpb25faWQYByABKAlSDGludm9jYXRpb25JZBIaCg'
    'hhdHRlbXB0cxgIIAEoBVIIYXR0ZW1wdHMSMQoUZGVzdGluYXRpb25fcmV2aXNpb24YCSABKANS'
    'E2Rlc3RpbmF0aW9uUmV2aXNpb24SFAoFZXJyb3IYCiABKAlSBWVycm9yEhYKBm91dHB1dBgLIA'
    'EoCVIGb3V0cHV0Ej4KCHNlZ21lbnRzGAwgAygLMiIuYWdlbnRzLnYxLlRlbGVncmFtRGVsaXZl'
    'cnlTZWdtZW50UghzZWdtZW50cxI5CgpjcmVhdGVkX2F0GA0gASgLMhouZ29vZ2xlLnByb3RvYn'
    'VmLlRpbWVzdGFtcFIJY3JlYXRlZEF0EjkKCnVwZGF0ZWRfYXQYDiABKAsyGi5nb29nbGUucHJv'
    'dG9idWYuVGltZXN0YW1wUgl1cGRhdGVkQXQSOQoKZXhwaXJlc19hdBgPIAEoCzIaLmdvb2dsZS'
    '5wcm90b2J1Zi5UaW1lc3RhbXBSCWV4cGlyZXNBdBIjCg1kZWFkX2xldHRlcmVkGBAgASgIUgxk'
    'ZWFkTGV0dGVyZWQ=');

@$core.Deprecated('Use listTelegramProcessingRecordsRequestDescriptor instead')
const ListTelegramProcessingRecordsRequest$json = {
  '1': 'ListTelegramProcessingRecordsRequest',
  '2': [
    {'1': 'channel_id', '3': 1, '4': 1, '5': 9, '10': 'channelId'},
    {'1': 'destination_id', '3': 2, '4': 1, '5': 9, '10': 'destinationId'},
    {
      '1': 'status',
      '3': 3,
      '4': 1,
      '5': 14,
      '6': '.agents.v1.TelegramProcessingStatus',
      '10': 'status'
    },
    {'1': 'page_size', '3': 4, '4': 1, '5': 5, '10': 'pageSize'},
  ],
};

/// Descriptor for `ListTelegramProcessingRecordsRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listTelegramProcessingRecordsRequestDescriptor =
    $convert.base64Decode(
        'CiRMaXN0VGVsZWdyYW1Qcm9jZXNzaW5nUmVjb3Jkc1JlcXVlc3QSHQoKY2hhbm5lbF9pZBgBIA'
        'EoCVIJY2hhbm5lbElkEiUKDmRlc3RpbmF0aW9uX2lkGAIgASgJUg1kZXN0aW5hdGlvbklkEjsK'
        'BnN0YXR1cxgDIAEoDjIjLmFnZW50cy52MS5UZWxlZ3JhbVByb2Nlc3NpbmdTdGF0dXNSBnN0YX'
        'R1cxIbCglwYWdlX3NpemUYBCABKAVSCHBhZ2VTaXpl');

@$core.Deprecated('Use listTelegramProcessingRecordsResponseDescriptor instead')
const ListTelegramProcessingRecordsResponse$json = {
  '1': 'ListTelegramProcessingRecordsResponse',
  '2': [
    {
      '1': 'records',
      '3': 1,
      '4': 3,
      '5': 11,
      '6': '.agents.v1.TelegramProcessingRecord',
      '10': 'records'
    },
  ],
};

/// Descriptor for `ListTelegramProcessingRecordsResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listTelegramProcessingRecordsResponseDescriptor =
    $convert.base64Decode(
        'CiVMaXN0VGVsZWdyYW1Qcm9jZXNzaW5nUmVjb3Jkc1Jlc3BvbnNlEj0KB3JlY29yZHMYASADKA'
        'syIy5hZ2VudHMudjEuVGVsZWdyYW1Qcm9jZXNzaW5nUmVjb3JkUgdyZWNvcmRz');

@$core.Deprecated('Use getTelegramProcessingRecordRequestDescriptor instead')
const GetTelegramProcessingRecordRequest$json = {
  '1': 'GetTelegramProcessingRecordRequest',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '8': {}, '10': 'id'},
  ],
};

/// Descriptor for `GetTelegramProcessingRecordRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getTelegramProcessingRecordRequestDescriptor =
    $convert.base64Decode(
        'CiJHZXRUZWxlZ3JhbVByb2Nlc3NpbmdSZWNvcmRSZXF1ZXN0EhcKAmlkGAEgASgJQgf6QgRyAh'
        'ABUgJpZA==');

@$core.Deprecated('Use getTelegramProcessingRecordResponseDescriptor instead')
const GetTelegramProcessingRecordResponse$json = {
  '1': 'GetTelegramProcessingRecordResponse',
  '2': [
    {
      '1': 'record',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.TelegramProcessingRecord',
      '10': 'record'
    },
  ],
};

/// Descriptor for `GetTelegramProcessingRecordResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getTelegramProcessingRecordResponseDescriptor =
    $convert.base64Decode(
        'CiNHZXRUZWxlZ3JhbVByb2Nlc3NpbmdSZWNvcmRSZXNwb25zZRI7CgZyZWNvcmQYASABKAsyIy'
        '5hZ2VudHMudjEuVGVsZWdyYW1Qcm9jZXNzaW5nUmVjb3JkUgZyZWNvcmQ=');

@$core.Deprecated('Use resendTelegramReplyRequestDescriptor instead')
const ResendTelegramReplyRequest$json = {
  '1': 'ResendTelegramReplyRequest',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '8': {}, '10': 'id'},
  ],
};

/// Descriptor for `ResendTelegramReplyRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List resendTelegramReplyRequestDescriptor =
    $convert.base64Decode(
        'ChpSZXNlbmRUZWxlZ3JhbVJlcGx5UmVxdWVzdBIXCgJpZBgBIAEoCUIH+kIEcgIQAVICaWQ=');

@$core.Deprecated('Use resendTelegramReplyResponseDescriptor instead')
const ResendTelegramReplyResponse$json = {
  '1': 'ResendTelegramReplyResponse',
  '2': [
    {
      '1': 'record',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.TelegramProcessingRecord',
      '10': 'record'
    },
    {'1': 'message_ids', '3': 2, '4': 3, '5': 9, '10': 'messageIds'},
  ],
};

/// Descriptor for `ResendTelegramReplyResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List resendTelegramReplyResponseDescriptor =
    $convert.base64Decode(
        'ChtSZXNlbmRUZWxlZ3JhbVJlcGx5UmVzcG9uc2USOwoGcmVjb3JkGAEgASgLMiMuYWdlbnRzLn'
        'YxLlRlbGVncmFtUHJvY2Vzc2luZ1JlY29yZFIGcmVjb3JkEh8KC21lc3NhZ2VfaWRzGAIgAygJ'
        'UgptZXNzYWdlSWRz');

const $core.Map<$core.String, $core.dynamic> TelegramChannelServiceBase$json = {
  '1': 'TelegramChannelService',
  '2': [
    {
      '1': 'ListTelegramChannels',
      '2': '.agents.v1.ListTelegramChannelsRequest',
      '3': '.agents.v1.ListTelegramChannelsResponse'
    },
    {
      '1': 'GetTelegramChannel',
      '2': '.agents.v1.GetTelegramChannelRequest',
      '3': '.agents.v1.GetTelegramChannelResponse'
    },
    {
      '1': 'CreateTelegramChannel',
      '2': '.agents.v1.CreateTelegramChannelRequest',
      '3': '.agents.v1.CreateTelegramChannelResponse'
    },
    {
      '1': 'UpdateTelegramChannel',
      '2': '.agents.v1.UpdateTelegramChannelRequest',
      '3': '.agents.v1.UpdateTelegramChannelResponse'
    },
    {
      '1': 'RotateTelegramChannelCredential',
      '2': '.agents.v1.RotateTelegramChannelCredentialRequest',
      '3': '.agents.v1.RotateTelegramChannelCredentialResponse'
    },
    {
      '1': 'SetTelegramChannelEnabled',
      '2': '.agents.v1.SetTelegramChannelEnabledRequest',
      '3': '.agents.v1.SetTelegramChannelEnabledResponse'
    },
    {
      '1': 'DeleteTelegramChannel',
      '2': '.agents.v1.DeleteTelegramChannelRequest',
      '3': '.agents.v1.DeleteTelegramChannelResponse'
    },
    {
      '1': 'GetTelegramChannelStatus',
      '2': '.agents.v1.GetTelegramChannelStatusRequest',
      '3': '.agents.v1.GetTelegramChannelStatusResponse'
    },
  ],
};

@$core.Deprecated('Use telegramChannelServiceDescriptor instead')
const $core.Map<$core.String, $core.Map<$core.String, $core.dynamic>>
    TelegramChannelServiceBase$messageJson = {
  '.agents.v1.ListTelegramChannelsRequest': ListTelegramChannelsRequest$json,
  '.agents.v1.ListTelegramChannelsResponse': ListTelegramChannelsResponse$json,
  '.agents.v1.TelegramChannel': TelegramChannel$json,
  '.agents.v1.TelegramBotCapabilities': TelegramBotCapabilities$json,
  '.google.protobuf.Timestamp': $0.Timestamp$json,
  '.agents.v1.GetTelegramChannelRequest': GetTelegramChannelRequest$json,
  '.agents.v1.GetTelegramChannelResponse': GetTelegramChannelResponse$json,
  '.agents.v1.CreateTelegramChannelRequest': CreateTelegramChannelRequest$json,
  '.agents.v1.CreateTelegramChannelResponse':
      CreateTelegramChannelResponse$json,
  '.agents.v1.UpdateTelegramChannelRequest': UpdateTelegramChannelRequest$json,
  '.agents.v1.UpdateTelegramChannelResponse':
      UpdateTelegramChannelResponse$json,
  '.agents.v1.RotateTelegramChannelCredentialRequest':
      RotateTelegramChannelCredentialRequest$json,
  '.agents.v1.RotateTelegramChannelCredentialResponse':
      RotateTelegramChannelCredentialResponse$json,
  '.agents.v1.SetTelegramChannelEnabledRequest':
      SetTelegramChannelEnabledRequest$json,
  '.agents.v1.SetTelegramChannelEnabledResponse':
      SetTelegramChannelEnabledResponse$json,
  '.agents.v1.DeleteTelegramChannelRequest': DeleteTelegramChannelRequest$json,
  '.agents.v1.DeleteTelegramChannelResponse':
      DeleteTelegramChannelResponse$json,
  '.agents.v1.GetTelegramChannelStatusRequest':
      GetTelegramChannelStatusRequest$json,
  '.agents.v1.GetTelegramChannelStatusResponse':
      GetTelegramChannelStatusResponse$json,
  '.agents.v1.TelegramChannelStatus': TelegramChannelStatus$json,
};

/// Descriptor for `TelegramChannelService`. Decode as a `google.protobuf.ServiceDescriptorProto`.
final $typed_data.Uint8List telegramChannelServiceDescriptor = $convert.base64Decode(
    'ChZUZWxlZ3JhbUNoYW5uZWxTZXJ2aWNlEmcKFExpc3RUZWxlZ3JhbUNoYW5uZWxzEiYuYWdlbn'
    'RzLnYxLkxpc3RUZWxlZ3JhbUNoYW5uZWxzUmVxdWVzdBonLmFnZW50cy52MS5MaXN0VGVsZWdy'
    'YW1DaGFubmVsc1Jlc3BvbnNlEmEKEkdldFRlbGVncmFtQ2hhbm5lbBIkLmFnZW50cy52MS5HZX'
    'RUZWxlZ3JhbUNoYW5uZWxSZXF1ZXN0GiUuYWdlbnRzLnYxLkdldFRlbGVncmFtQ2hhbm5lbFJl'
    'c3BvbnNlEmoKFUNyZWF0ZVRlbGVncmFtQ2hhbm5lbBInLmFnZW50cy52MS5DcmVhdGVUZWxlZ3'
    'JhbUNoYW5uZWxSZXF1ZXN0GiguYWdlbnRzLnYxLkNyZWF0ZVRlbGVncmFtQ2hhbm5lbFJlc3Bv'
    'bnNlEmoKFVVwZGF0ZVRlbGVncmFtQ2hhbm5lbBInLmFnZW50cy52MS5VcGRhdGVUZWxlZ3JhbU'
    'NoYW5uZWxSZXF1ZXN0GiguYWdlbnRzLnYxLlVwZGF0ZVRlbGVncmFtQ2hhbm5lbFJlc3BvbnNl'
    'EogBCh9Sb3RhdGVUZWxlZ3JhbUNoYW5uZWxDcmVkZW50aWFsEjEuYWdlbnRzLnYxLlJvdGF0ZV'
    'RlbGVncmFtQ2hhbm5lbENyZWRlbnRpYWxSZXF1ZXN0GjIuYWdlbnRzLnYxLlJvdGF0ZVRlbGVn'
    'cmFtQ2hhbm5lbENyZWRlbnRpYWxSZXNwb25zZRJ2ChlTZXRUZWxlZ3JhbUNoYW5uZWxFbmFibG'
    'VkEisuYWdlbnRzLnYxLlNldFRlbGVncmFtQ2hhbm5lbEVuYWJsZWRSZXF1ZXN0GiwuYWdlbnRz'
    'LnYxLlNldFRlbGVncmFtQ2hhbm5lbEVuYWJsZWRSZXNwb25zZRJqChVEZWxldGVUZWxlZ3JhbU'
    'NoYW5uZWwSJy5hZ2VudHMudjEuRGVsZXRlVGVsZWdyYW1DaGFubmVsUmVxdWVzdBooLmFnZW50'
    'cy52MS5EZWxldGVUZWxlZ3JhbUNoYW5uZWxSZXNwb25zZRJzChhHZXRUZWxlZ3JhbUNoYW5uZW'
    'xTdGF0dXMSKi5hZ2VudHMudjEuR2V0VGVsZWdyYW1DaGFubmVsU3RhdHVzUmVxdWVzdBorLmFn'
    'ZW50cy52MS5HZXRUZWxlZ3JhbUNoYW5uZWxTdGF0dXNSZXNwb25zZQ==');

const $core.Map<$core.String, $core.dynamic>
    TelegramDestinationServiceBase$json = {
  '1': 'TelegramDestinationService',
  '2': [
    {
      '1': 'ListTelegramDestinations',
      '2': '.agents.v1.ListTelegramDestinationsRequest',
      '3': '.agents.v1.ListTelegramDestinationsResponse'
    },
    {
      '1': 'GetTelegramDestination',
      '2': '.agents.v1.GetTelegramDestinationRequest',
      '3': '.agents.v1.GetTelegramDestinationResponse'
    },
    {
      '1': 'CreateTelegramDestination',
      '2': '.agents.v1.CreateTelegramDestinationRequest',
      '3': '.agents.v1.CreateTelegramDestinationResponse'
    },
    {
      '1': 'UpdateTelegramDestination',
      '2': '.agents.v1.UpdateTelegramDestinationRequest',
      '3': '.agents.v1.UpdateTelegramDestinationResponse'
    },
    {
      '1': 'DeleteTelegramDestination',
      '2': '.agents.v1.DeleteTelegramDestinationRequest',
      '3': '.agents.v1.DeleteTelegramDestinationResponse'
    },
    {
      '1': 'SendTelegramTestMessage',
      '2': '.agents.v1.SendTelegramTestMessageRequest',
      '3': '.agents.v1.SendTelegramTestMessageResponse'
    },
  ],
};

@$core.Deprecated('Use telegramDestinationServiceDescriptor instead')
const $core.Map<$core.String, $core.Map<$core.String, $core.dynamic>>
    TelegramDestinationServiceBase$messageJson = {
  '.agents.v1.ListTelegramDestinationsRequest':
      ListTelegramDestinationsRequest$json,
  '.agents.v1.ListTelegramDestinationsResponse':
      ListTelegramDestinationsResponse$json,
  '.agents.v1.TelegramDestination': TelegramDestination$json,
  '.agents.v1.TelegramDestinationConfig': TelegramDestinationConfig$json,
  '.google.protobuf.Timestamp': $0.Timestamp$json,
  '.agents.v1.TelegramDestinationVerification':
      TelegramDestinationVerification$json,
  '.agents.v1.GetTelegramDestinationRequest':
      GetTelegramDestinationRequest$json,
  '.agents.v1.GetTelegramDestinationResponse':
      GetTelegramDestinationResponse$json,
  '.agents.v1.CreateTelegramDestinationRequest':
      CreateTelegramDestinationRequest$json,
  '.agents.v1.CreateTelegramDestinationResponse':
      CreateTelegramDestinationResponse$json,
  '.agents.v1.UpdateTelegramDestinationRequest':
      UpdateTelegramDestinationRequest$json,
  '.agents.v1.UpdateTelegramDestinationResponse':
      UpdateTelegramDestinationResponse$json,
  '.agents.v1.DeleteTelegramDestinationRequest':
      DeleteTelegramDestinationRequest$json,
  '.agents.v1.DeleteTelegramDestinationResponse':
      DeleteTelegramDestinationResponse$json,
  '.agents.v1.SendTelegramTestMessageRequest':
      SendTelegramTestMessageRequest$json,
  '.agents.v1.SendTelegramTestMessageResponse':
      SendTelegramTestMessageResponse$json,
};

/// Descriptor for `TelegramDestinationService`. Decode as a `google.protobuf.ServiceDescriptorProto`.
final $typed_data.Uint8List telegramDestinationServiceDescriptor = $convert.base64Decode(
    'ChpUZWxlZ3JhbURlc3RpbmF0aW9uU2VydmljZRJzChhMaXN0VGVsZWdyYW1EZXN0aW5hdGlvbn'
    'MSKi5hZ2VudHMudjEuTGlzdFRlbGVncmFtRGVzdGluYXRpb25zUmVxdWVzdBorLmFnZW50cy52'
    'MS5MaXN0VGVsZWdyYW1EZXN0aW5hdGlvbnNSZXNwb25zZRJtChZHZXRUZWxlZ3JhbURlc3Rpbm'
    'F0aW9uEiguYWdlbnRzLnYxLkdldFRlbGVncmFtRGVzdGluYXRpb25SZXF1ZXN0GikuYWdlbnRz'
    'LnYxLkdldFRlbGVncmFtRGVzdGluYXRpb25SZXNwb25zZRJ2ChlDcmVhdGVUZWxlZ3JhbURlc3'
    'RpbmF0aW9uEisuYWdlbnRzLnYxLkNyZWF0ZVRlbGVncmFtRGVzdGluYXRpb25SZXF1ZXN0Giwu'
    'YWdlbnRzLnYxLkNyZWF0ZVRlbGVncmFtRGVzdGluYXRpb25SZXNwb25zZRJ2ChlVcGRhdGVUZW'
    'xlZ3JhbURlc3RpbmF0aW9uEisuYWdlbnRzLnYxLlVwZGF0ZVRlbGVncmFtRGVzdGluYXRpb25S'
    'ZXF1ZXN0GiwuYWdlbnRzLnYxLlVwZGF0ZVRlbGVncmFtRGVzdGluYXRpb25SZXNwb25zZRJ2Ch'
    'lEZWxldGVUZWxlZ3JhbURlc3RpbmF0aW9uEisuYWdlbnRzLnYxLkRlbGV0ZVRlbGVncmFtRGVz'
    'dGluYXRpb25SZXF1ZXN0GiwuYWdlbnRzLnYxLkRlbGV0ZVRlbGVncmFtRGVzdGluYXRpb25SZX'
    'Nwb25zZRJwChdTZW5kVGVsZWdyYW1UZXN0TWVzc2FnZRIpLmFnZW50cy52MS5TZW5kVGVsZWdy'
    'YW1UZXN0TWVzc2FnZVJlcXVlc3QaKi5hZ2VudHMudjEuU2VuZFRlbGVncmFtVGVzdE1lc3NhZ2'
    'VSZXNwb25zZQ==');

const $core.Map<$core.String, $core.dynamic> TelegramAdminServiceBase$json = {
  '1': 'TelegramAdminService',
  '2': [
    {
      '1': 'GetTelegramSettings',
      '2': '.agents.v1.GetTelegramSettingsRequest',
      '3': '.agents.v1.GetTelegramSettingsResponse'
    },
    {
      '1': 'UpdateTelegramSettings',
      '2': '.agents.v1.UpdateTelegramSettingsRequest',
      '3': '.agents.v1.UpdateTelegramSettingsResponse'
    },
  ],
};

@$core.Deprecated('Use telegramAdminServiceDescriptor instead')
const $core.Map<$core.String, $core.Map<$core.String, $core.dynamic>>
    TelegramAdminServiceBase$messageJson = {
  '.agents.v1.GetTelegramSettingsRequest': GetTelegramSettingsRequest$json,
  '.agents.v1.GetTelegramSettingsResponse': GetTelegramSettingsResponse$json,
  '.agents.v1.TelegramSettings': TelegramSettings$json,
  '.google.protobuf.Timestamp': $0.Timestamp$json,
  '.agents.v1.UpdateTelegramSettingsRequest':
      UpdateTelegramSettingsRequest$json,
  '.agents.v1.UpdateTelegramSettingsResponse':
      UpdateTelegramSettingsResponse$json,
};

/// Descriptor for `TelegramAdminService`. Decode as a `google.protobuf.ServiceDescriptorProto`.
final $typed_data.Uint8List telegramAdminServiceDescriptor = $convert.base64Decode(
    'ChRUZWxlZ3JhbUFkbWluU2VydmljZRJkChNHZXRUZWxlZ3JhbVNldHRpbmdzEiUuYWdlbnRzLn'
    'YxLkdldFRlbGVncmFtU2V0dGluZ3NSZXF1ZXN0GiYuYWdlbnRzLnYxLkdldFRlbGVncmFtU2V0'
    'dGluZ3NSZXNwb25zZRJtChZVcGRhdGVUZWxlZ3JhbVNldHRpbmdzEiguYWdlbnRzLnYxLlVwZG'
    'F0ZVRlbGVncmFtU2V0dGluZ3NSZXF1ZXN0GikuYWdlbnRzLnYxLlVwZGF0ZVRlbGVncmFtU2V0'
    'dGluZ3NSZXNwb25zZQ==');

const $core.Map<$core.String, $core.dynamic>
    TelegramProcessingServiceBase$json = {
  '1': 'TelegramProcessingService',
  '2': [
    {
      '1': 'ListTelegramProcessingRecords',
      '2': '.agents.v1.ListTelegramProcessingRecordsRequest',
      '3': '.agents.v1.ListTelegramProcessingRecordsResponse'
    },
    {
      '1': 'GetTelegramProcessingRecord',
      '2': '.agents.v1.GetTelegramProcessingRecordRequest',
      '3': '.agents.v1.GetTelegramProcessingRecordResponse'
    },
    {
      '1': 'ResendTelegramReply',
      '2': '.agents.v1.ResendTelegramReplyRequest',
      '3': '.agents.v1.ResendTelegramReplyResponse'
    },
  ],
};

@$core.Deprecated('Use telegramProcessingServiceDescriptor instead')
const $core.Map<$core.String, $core.Map<$core.String, $core.dynamic>>
    TelegramProcessingServiceBase$messageJson = {
  '.agents.v1.ListTelegramProcessingRecordsRequest':
      ListTelegramProcessingRecordsRequest$json,
  '.agents.v1.ListTelegramProcessingRecordsResponse':
      ListTelegramProcessingRecordsResponse$json,
  '.agents.v1.TelegramProcessingRecord': TelegramProcessingRecord$json,
  '.agents.v1.TelegramDeliverySegment': TelegramDeliverySegment$json,
  '.google.protobuf.Timestamp': $0.Timestamp$json,
  '.agents.v1.GetTelegramProcessingRecordRequest':
      GetTelegramProcessingRecordRequest$json,
  '.agents.v1.GetTelegramProcessingRecordResponse':
      GetTelegramProcessingRecordResponse$json,
  '.agents.v1.ResendTelegramReplyRequest': ResendTelegramReplyRequest$json,
  '.agents.v1.ResendTelegramReplyResponse': ResendTelegramReplyResponse$json,
};

/// Descriptor for `TelegramProcessingService`. Decode as a `google.protobuf.ServiceDescriptorProto`.
final $typed_data.Uint8List telegramProcessingServiceDescriptor = $convert.base64Decode(
    'ChlUZWxlZ3JhbVByb2Nlc3NpbmdTZXJ2aWNlEoIBCh1MaXN0VGVsZWdyYW1Qcm9jZXNzaW5nUm'
    'Vjb3JkcxIvLmFnZW50cy52MS5MaXN0VGVsZWdyYW1Qcm9jZXNzaW5nUmVjb3Jkc1JlcXVlc3Qa'
    'MC5hZ2VudHMudjEuTGlzdFRlbGVncmFtUHJvY2Vzc2luZ1JlY29yZHNSZXNwb25zZRJ8ChtHZX'
    'RUZWxlZ3JhbVByb2Nlc3NpbmdSZWNvcmQSLS5hZ2VudHMudjEuR2V0VGVsZWdyYW1Qcm9jZXNz'
    'aW5nUmVjb3JkUmVxdWVzdBouLmFnZW50cy52MS5HZXRUZWxlZ3JhbVByb2Nlc3NpbmdSZWNvcm'
    'RSZXNwb25zZRJkChNSZXNlbmRUZWxlZ3JhbVJlcGx5EiUuYWdlbnRzLnYxLlJlc2VuZFRlbGVn'
    'cmFtUmVwbHlSZXF1ZXN0GiYuYWdlbnRzLnYxLlJlc2VuZFRlbGVncmFtUmVwbHlSZXNwb25zZQ'
    '==');
