// This is a generated file - do not edit.
//
// Generated from agents/v1/automation.proto.

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

import 'package:protobuf/well_known_types/google/protobuf/duration.pbjson.dart'
    as $1;
import 'package:protobuf/well_known_types/google/protobuf/timestamp.pbjson.dart'
    as $0;

@$core.Deprecated('Use automationTriggerTypeDescriptor instead')
const AutomationTriggerType$json = {
  '1': 'AutomationTriggerType',
  '2': [
    {'1': 'AUTOMATION_TRIGGER_TYPE_UNSPECIFIED', '2': 0},
    {'1': 'AUTOMATION_TRIGGER_TYPE_MANUAL', '2': 1},
    {'1': 'AUTOMATION_TRIGGER_TYPE_SCHEDULE', '2': 2},
    {'1': 'AUTOMATION_TRIGGER_TYPE_WEBHOOK', '2': 3},
    {'1': 'AUTOMATION_TRIGGER_TYPE_FORUM_EVENT', '2': 4},
    {'1': 'AUTOMATION_TRIGGER_TYPE_CHANNEL_EVENT', '2': 5},
    {'1': 'AUTOMATION_TRIGGER_TYPE_DAEMON_EVENT', '2': 6},
  ],
};

/// Descriptor for `AutomationTriggerType`. Decode as a `google.protobuf.EnumDescriptorProto`.
final $typed_data.Uint8List automationTriggerTypeDescriptor = $convert.base64Decode(
    'ChVBdXRvbWF0aW9uVHJpZ2dlclR5cGUSJwojQVVUT01BVElPTl9UUklHR0VSX1RZUEVfVU5TUE'
    'VDSUZJRUQQABIiCh5BVVRPTUFUSU9OX1RSSUdHRVJfVFlQRV9NQU5VQUwQARIkCiBBVVRPTUFU'
    'SU9OX1RSSUdHRVJfVFlQRV9TQ0hFRFVMRRACEiMKH0FVVE9NQVRJT05fVFJJR0dFUl9UWVBFX1'
    'dFQkhPT0sQAxInCiNBVVRPTUFUSU9OX1RSSUdHRVJfVFlQRV9GT1JVTV9FVkVOVBAEEikKJUFV'
    'VE9NQVRJT05fVFJJR0dFUl9UWVBFX0NIQU5ORUxfRVZFTlQQBRIoCiRBVVRPTUFUSU9OX1RSSU'
    'dHRVJfVFlQRV9EQUVNT05fRVZFTlQQBg==');

@$core.Deprecated('Use automationConditionOperatorDescriptor instead')
const AutomationConditionOperator$json = {
  '1': 'AutomationConditionOperator',
  '2': [
    {'1': 'AUTOMATION_CONDITION_OPERATOR_UNSPECIFIED', '2': 0},
    {'1': 'AUTOMATION_CONDITION_OPERATOR_EQUALS', '2': 1},
    {'1': 'AUTOMATION_CONDITION_OPERATOR_NOT_EQUALS', '2': 2},
    {'1': 'AUTOMATION_CONDITION_OPERATOR_CONTAINS', '2': 3},
    {'1': 'AUTOMATION_CONDITION_OPERATOR_REGEX_MATCH', '2': 4},
    {'1': 'AUTOMATION_CONDITION_OPERATOR_EXISTS', '2': 5},
    {'1': 'AUTOMATION_CONDITION_OPERATOR_NOT_EXISTS', '2': 6},
  ],
};

/// Descriptor for `AutomationConditionOperator`. Decode as a `google.protobuf.EnumDescriptorProto`.
final $typed_data.Uint8List automationConditionOperatorDescriptor = $convert.base64Decode(
    'ChtBdXRvbWF0aW9uQ29uZGl0aW9uT3BlcmF0b3ISLQopQVVUT01BVElPTl9DT05ESVRJT05fT1'
    'BFUkFUT1JfVU5TUEVDSUZJRUQQABIoCiRBVVRPTUFUSU9OX0NPTkRJVElPTl9PUEVSQVRPUl9F'
    'UVVBTFMQARIsCihBVVRPTUFUSU9OX0NPTkRJVElPTl9PUEVSQVRPUl9OT1RfRVFVQUxTEAISKg'
    'omQVVUT01BVElPTl9DT05ESVRJT05fT1BFUkFUT1JfQ09OVEFJTlMQAxItCilBVVRPTUFUSU9O'
    'X0NPTkRJVElPTl9PUEVSQVRPUl9SRUdFWF9NQVRDSBAEEigKJEFVVE9NQVRJT05fQ09ORElUSU'
    '9OX09QRVJBVE9SX0VYSVNUUxAFEiwKKEFVVE9NQVRJT05fQ09ORElUSU9OX09QRVJBVE9SX05P'
    'VF9FWElTVFMQBg==');

@$core.Deprecated('Use automationStepTypeDescriptor instead')
const AutomationStepType$json = {
  '1': 'AutomationStepType',
  '2': [
    {'1': 'AUTOMATION_STEP_TYPE_UNSPECIFIED', '2': 0},
    {'1': 'AUTOMATION_STEP_TYPE_INVOKE_AGENT', '2': 1},
    {'1': 'AUTOMATION_STEP_TYPE_CALL_WEBHOOK', '2': 2},
    {'1': 'AUTOMATION_STEP_TYPE_SEND_NOTIFY_GROUP', '2': 3},
    {'1': 'AUTOMATION_STEP_TYPE_CREATE_FORUM_POST', '2': 4},
  ],
};

/// Descriptor for `AutomationStepType`. Decode as a `google.protobuf.EnumDescriptorProto`.
final $typed_data.Uint8List automationStepTypeDescriptor = $convert.base64Decode(
    'ChJBdXRvbWF0aW9uU3RlcFR5cGUSJAogQVVUT01BVElPTl9TVEVQX1RZUEVfVU5TUEVDSUZJRU'
    'QQABIlCiFBVVRPTUFUSU9OX1NURVBfVFlQRV9JTlZPS0VfQUdFTlQQARIlCiFBVVRPTUFUSU9O'
    'X1NURVBfVFlQRV9DQUxMX1dFQkhPT0sQAhIqCiZBVVRPTUFUSU9OX1NURVBfVFlQRV9TRU5EX0'
    '5PVElGWV9HUk9VUBADEioKJkFVVE9NQVRJT05fU1RFUF9UWVBFX0NSRUFURV9GT1JVTV9QT1NU'
    'EAQ=');

@$core.Deprecated('Use automationConcurrencyPolicyDescriptor instead')
const AutomationConcurrencyPolicy$json = {
  '1': 'AutomationConcurrencyPolicy',
  '2': [
    {'1': 'AUTOMATION_CONCURRENCY_POLICY_UNSPECIFIED', '2': 0},
    {'1': 'AUTOMATION_CONCURRENCY_POLICY_SKIP', '2': 1},
    {'1': 'AUTOMATION_CONCURRENCY_POLICY_QUEUE', '2': 2},
    {'1': 'AUTOMATION_CONCURRENCY_POLICY_REPLACE', '2': 3},
    {'1': 'AUTOMATION_CONCURRENCY_POLICY_ALLOW', '2': 4},
  ],
};

/// Descriptor for `AutomationConcurrencyPolicy`. Decode as a `google.protobuf.EnumDescriptorProto`.
final $typed_data.Uint8List automationConcurrencyPolicyDescriptor = $convert.base64Decode(
    'ChtBdXRvbWF0aW9uQ29uY3VycmVuY3lQb2xpY3kSLQopQVVUT01BVElPTl9DT05DVVJSRU5DWV'
    '9QT0xJQ1lfVU5TUEVDSUZJRUQQABImCiJBVVRPTUFUSU9OX0NPTkNVUlJFTkNZX1BPTElDWV9T'
    'S0lQEAESJwojQVVUT01BVElPTl9DT05DVVJSRU5DWV9QT0xJQ1lfUVVFVUUQAhIpCiVBVVRPTU'
    'FUSU9OX0NPTkNVUlJFTkNZX1BPTElDWV9SRVBMQUNFEAMSJwojQVVUT01BVElPTl9DT05DVVJS'
    'RU5DWV9QT0xJQ1lfQUxMT1cQBA==');

@$core.Deprecated('Use automationRunStatusDescriptor instead')
const AutomationRunStatus$json = {
  '1': 'AutomationRunStatus',
  '2': [
    {'1': 'AUTOMATION_RUN_STATUS_UNSPECIFIED', '2': 0},
    {'1': 'AUTOMATION_RUN_STATUS_RUNNING', '2': 1},
    {'1': 'AUTOMATION_RUN_STATUS_SUCCEEDED', '2': 2},
    {'1': 'AUTOMATION_RUN_STATUS_FAILED', '2': 3},
    {'1': 'AUTOMATION_RUN_STATUS_SKIPPED', '2': 4},
    {'1': 'AUTOMATION_RUN_STATUS_CANCELLED', '2': 5},
    {'1': 'AUTOMATION_RUN_STATUS_WAITING_INPUT', '2': 6},
  ],
};

/// Descriptor for `AutomationRunStatus`. Decode as a `google.protobuf.EnumDescriptorProto`.
final $typed_data.Uint8List automationRunStatusDescriptor = $convert.base64Decode(
    'ChNBdXRvbWF0aW9uUnVuU3RhdHVzEiUKIUFVVE9NQVRJT05fUlVOX1NUQVRVU19VTlNQRUNJRk'
    'lFRBAAEiEKHUFVVE9NQVRJT05fUlVOX1NUQVRVU19SVU5OSU5HEAESIwofQVVUT01BVElPTl9S'
    'VU5fU1RBVFVTX1NVQ0NFRURFRBACEiAKHEFVVE9NQVRJT05fUlVOX1NUQVRVU19GQUlMRUQQAx'
    'IhCh1BVVRPTUFUSU9OX1JVTl9TVEFUVVNfU0tJUFBFRBAEEiMKH0FVVE9NQVRJT05fUlVOX1NU'
    'QVRVU19DQU5DRUxMRUQQBRInCiNBVVRPTUFUSU9OX1JVTl9TVEFUVVNfV0FJVElOR19JTlBVVB'
    'AG');

@$core.Deprecated('Use automationStepRunStatusDescriptor instead')
const AutomationStepRunStatus$json = {
  '1': 'AutomationStepRunStatus',
  '2': [
    {'1': 'AUTOMATION_STEP_RUN_STATUS_UNSPECIFIED', '2': 0},
    {'1': 'AUTOMATION_STEP_RUN_STATUS_RUNNING', '2': 1},
    {'1': 'AUTOMATION_STEP_RUN_STATUS_SUCCEEDED', '2': 2},
    {'1': 'AUTOMATION_STEP_RUN_STATUS_FAILED', '2': 3},
    {'1': 'AUTOMATION_STEP_RUN_STATUS_SKIPPED', '2': 4},
    {'1': 'AUTOMATION_STEP_RUN_STATUS_CANCELLED', '2': 5},
  ],
};

/// Descriptor for `AutomationStepRunStatus`. Decode as a `google.protobuf.EnumDescriptorProto`.
final $typed_data.Uint8List automationStepRunStatusDescriptor = $convert.base64Decode(
    'ChdBdXRvbWF0aW9uU3RlcFJ1blN0YXR1cxIqCiZBVVRPTUFUSU9OX1NURVBfUlVOX1NUQVRVU1'
    '9VTlNQRUNJRklFRBAAEiYKIkFVVE9NQVRJT05fU1RFUF9SVU5fU1RBVFVTX1JVTk5JTkcQARIo'
    'CiRBVVRPTUFUSU9OX1NURVBfUlVOX1NUQVRVU19TVUNDRUVERUQQAhIlCiFBVVRPTUFUSU9OX1'
    'NURVBfUlVOX1NUQVRVU19GQUlMRUQQAxImCiJBVVRPTUFUSU9OX1NURVBfUlVOX1NUQVRVU19T'
    'S0lQUEVEEAQSKAokQVVUT01BVElPTl9TVEVQX1JVTl9TVEFUVVNfQ0FOQ0VMTEVEEAU=');

@$core.Deprecated('Use automationDescriptor instead')
const Automation$json = {
  '1': 'Automation',
  '2': [
    {'1': 'name', '3': 1, '4': 1, '5': 9, '8': {}, '10': 'name'},
    {'1': 'enabled', '3': 2, '4': 1, '5': 8, '10': 'enabled'},
    {
      '1': 'trigger',
      '3': 3,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.AutomationTrigger',
      '10': 'trigger'
    },
    {
      '1': 'conditions',
      '3': 4,
      '4': 3,
      '5': 11,
      '6': '.agents.v1.AutomationCondition',
      '10': 'conditions'
    },
    {
      '1': 'steps',
      '3': 5,
      '4': 3,
      '5': 11,
      '6': '.agents.v1.AutomationStep',
      '10': 'steps'
    },
    {
      '1': 'policy',
      '3': 6,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.AutomationPolicy',
      '10': 'policy'
    },
    {
      '1': 'metadata',
      '3': 20,
      '4': 3,
      '5': 11,
      '6': '.agents.v1.Automation.MetadataEntry',
      '10': 'metadata'
    },
    {
      '1': 'created_at',
      '3': 30,
      '4': 1,
      '5': 11,
      '6': '.google.protobuf.Timestamp',
      '10': 'createdAt'
    },
    {
      '1': 'updated_at',
      '3': 31,
      '4': 1,
      '5': 11,
      '6': '.google.protobuf.Timestamp',
      '10': 'updatedAt'
    },
    {'1': 'workspace_id', '3': 100, '4': 1, '5': 9, '10': 'workspaceId'},
  ],
  '3': [Automation_MetadataEntry$json],
};

@$core.Deprecated('Use automationDescriptor instead')
const Automation_MetadataEntry$json = {
  '1': 'MetadataEntry',
  '2': [
    {'1': 'key', '3': 1, '4': 1, '5': 9, '10': 'key'},
    {'1': 'value', '3': 2, '4': 1, '5': 9, '10': 'value'},
  ],
  '7': {'7': true},
};

/// Descriptor for `Automation`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List automationDescriptor = $convert.base64Decode(
    'CgpBdXRvbWF0aW9uEhsKBG5hbWUYASABKAlCB/pCBHICEAFSBG5hbWUSGAoHZW5hYmxlZBgCIA'
    'EoCFIHZW5hYmxlZBI2Cgd0cmlnZ2VyGAMgASgLMhwuYWdlbnRzLnYxLkF1dG9tYXRpb25Ucmln'
    'Z2VyUgd0cmlnZ2VyEj4KCmNvbmRpdGlvbnMYBCADKAsyHi5hZ2VudHMudjEuQXV0b21hdGlvbk'
    'NvbmRpdGlvblIKY29uZGl0aW9ucxIvCgVzdGVwcxgFIAMoCzIZLmFnZW50cy52MS5BdXRvbWF0'
    'aW9uU3RlcFIFc3RlcHMSMwoGcG9saWN5GAYgASgLMhsuYWdlbnRzLnYxLkF1dG9tYXRpb25Qb2'
    'xpY3lSBnBvbGljeRI/CghtZXRhZGF0YRgUIAMoCzIjLmFnZW50cy52MS5BdXRvbWF0aW9uLk1l'
    'dGFkYXRhRW50cnlSCG1ldGFkYXRhEjkKCmNyZWF0ZWRfYXQYHiABKAsyGi5nb29nbGUucHJvdG'
    '9idWYuVGltZXN0YW1wUgljcmVhdGVkQXQSOQoKdXBkYXRlZF9hdBgfIAEoCzIaLmdvb2dsZS5w'
    'cm90b2J1Zi5UaW1lc3RhbXBSCXVwZGF0ZWRBdBIhCgx3b3Jrc3BhY2VfaWQYZCABKAlSC3dvcm'
    'tzcGFjZUlkGjsKDU1ldGFkYXRhRW50cnkSEAoDa2V5GAEgASgJUgNrZXkSFAoFdmFsdWUYAiAB'
    'KAlSBXZhbHVlOgI4AQ==');

@$core.Deprecated('Use automationTriggerDescriptor instead')
const AutomationTrigger$json = {
  '1': 'AutomationTrigger',
  '2': [
    {
      '1': 'type',
      '3': 1,
      '4': 1,
      '5': 14,
      '6': '.agents.v1.AutomationTriggerType',
      '10': 'type'
    },
    {
      '1': 'schedule',
      '3': 2,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.AutomationScheduleTrigger',
      '10': 'schedule'
    },
    {
      '1': 'webhook',
      '3': 3,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.AutomationWebhookTrigger',
      '10': 'webhook'
    },
    {
      '1': 'event',
      '3': 4,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.AutomationEventTrigger',
      '10': 'event'
    },
  ],
};

/// Descriptor for `AutomationTrigger`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List automationTriggerDescriptor = $convert.base64Decode(
    'ChFBdXRvbWF0aW9uVHJpZ2dlchI0CgR0eXBlGAEgASgOMiAuYWdlbnRzLnYxLkF1dG9tYXRpb2'
    '5UcmlnZ2VyVHlwZVIEdHlwZRJACghzY2hlZHVsZRgCIAEoCzIkLmFnZW50cy52MS5BdXRvbWF0'
    'aW9uU2NoZWR1bGVUcmlnZ2VyUghzY2hlZHVsZRI9Cgd3ZWJob29rGAMgASgLMiMuYWdlbnRzLn'
    'YxLkF1dG9tYXRpb25XZWJob29rVHJpZ2dlclIHd2ViaG9vaxI3CgVldmVudBgEIAEoCzIhLmFn'
    'ZW50cy52MS5BdXRvbWF0aW9uRXZlbnRUcmlnZ2VyUgVldmVudA==');

@$core.Deprecated('Use automationScheduleTriggerDescriptor instead')
const AutomationScheduleTrigger$json = {
  '1': 'AutomationScheduleTrigger',
  '2': [
    {'1': 'schedule', '3': 1, '4': 1, '5': 9, '8': {}, '10': 'schedule'},
    {'1': 'timezone', '3': 2, '4': 1, '5': 9, '10': 'timezone'},
  ],
};

/// Descriptor for `AutomationScheduleTrigger`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List automationScheduleTriggerDescriptor =
    $convert.base64Decode(
        'ChlBdXRvbWF0aW9uU2NoZWR1bGVUcmlnZ2VyEiMKCHNjaGVkdWxlGAEgASgJQgf6QgRyAhABUg'
        'hzY2hlZHVsZRIaCgh0aW1lem9uZRgCIAEoCVIIdGltZXpvbmU=');

@$core.Deprecated('Use automationWebhookTriggerDescriptor instead')
const AutomationWebhookTrigger$json = {
  '1': 'AutomationWebhookTrigger',
  '2': [
    {'1': 'token', '3': 1, '4': 1, '5': 9, '10': 'token'},
  ],
};

/// Descriptor for `AutomationWebhookTrigger`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List automationWebhookTriggerDescriptor =
    $convert.base64Decode(
        'ChhBdXRvbWF0aW9uV2ViaG9va1RyaWdnZXISFAoFdG9rZW4YASABKAlSBXRva2Vu');

@$core.Deprecated('Use automationEventTriggerDescriptor instead')
const AutomationEventTrigger$json = {
  '1': 'AutomationEventTrigger',
  '2': [
    {'1': 'event_type', '3': 1, '4': 1, '5': 9, '10': 'eventType'},
    {
      '1': 'filters',
      '3': 2,
      '4': 3,
      '5': 11,
      '6': '.agents.v1.AutomationEventTrigger.FiltersEntry',
      '10': 'filters'
    },
  ],
  '3': [AutomationEventTrigger_FiltersEntry$json],
};

@$core.Deprecated('Use automationEventTriggerDescriptor instead')
const AutomationEventTrigger_FiltersEntry$json = {
  '1': 'FiltersEntry',
  '2': [
    {'1': 'key', '3': 1, '4': 1, '5': 9, '10': 'key'},
    {'1': 'value', '3': 2, '4': 1, '5': 9, '10': 'value'},
  ],
  '7': {'7': true},
};

/// Descriptor for `AutomationEventTrigger`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List automationEventTriggerDescriptor = $convert.base64Decode(
    'ChZBdXRvbWF0aW9uRXZlbnRUcmlnZ2VyEh0KCmV2ZW50X3R5cGUYASABKAlSCWV2ZW50VHlwZR'
    'JICgdmaWx0ZXJzGAIgAygLMi4uYWdlbnRzLnYxLkF1dG9tYXRpb25FdmVudFRyaWdnZXIuRmls'
    'dGVyc0VudHJ5UgdmaWx0ZXJzGjoKDEZpbHRlcnNFbnRyeRIQCgNrZXkYASABKAlSA2tleRIUCg'
    'V2YWx1ZRgCIAEoCVIFdmFsdWU6AjgB');

@$core.Deprecated('Use automationConditionDescriptor instead')
const AutomationCondition$json = {
  '1': 'AutomationCondition',
  '2': [
    {'1': 'selector', '3': 1, '4': 1, '5': 9, '8': {}, '10': 'selector'},
    {
      '1': 'operator',
      '3': 2,
      '4': 1,
      '5': 14,
      '6': '.agents.v1.AutomationConditionOperator',
      '10': 'operator'
    },
    {'1': 'value', '3': 3, '4': 1, '5': 9, '10': 'value'},
  ],
};

/// Descriptor for `AutomationCondition`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List automationConditionDescriptor = $convert.base64Decode(
    'ChNBdXRvbWF0aW9uQ29uZGl0aW9uEiMKCHNlbGVjdG9yGAEgASgJQgf6QgRyAhABUghzZWxlY3'
    'RvchJCCghvcGVyYXRvchgCIAEoDjImLmFnZW50cy52MS5BdXRvbWF0aW9uQ29uZGl0aW9uT3Bl'
    'cmF0b3JSCG9wZXJhdG9yEhQKBXZhbHVlGAMgASgJUgV2YWx1ZQ==');

@$core.Deprecated('Use automationStepDescriptor instead')
const AutomationStep$json = {
  '1': 'AutomationStep',
  '2': [
    {'1': 'name', '3': 1, '4': 1, '5': 9, '8': {}, '10': 'name'},
    {
      '1': 'type',
      '3': 2,
      '4': 1,
      '5': 14,
      '6': '.agents.v1.AutomationStepType',
      '10': 'type'
    },
    {
      '1': 'invoke_agent',
      '3': 3,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.AutomationInvokeAgentStep',
      '10': 'invokeAgent'
    },
    {
      '1': 'call_webhook',
      '3': 4,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.AutomationCallWebhookStep',
      '10': 'callWebhook'
    },
    {
      '1': 'send_notify_group',
      '3': 5,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.AutomationSendNotifyGroupStep',
      '10': 'sendNotifyGroup'
    },
    {
      '1': 'create_forum_post',
      '3': 6,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.AutomationCreateForumPostStep',
      '10': 'createForumPost'
    },
    {
      '1': 'policy',
      '3': 10,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.AutomationPolicy',
      '10': 'policy'
    },
  ],
};

/// Descriptor for `AutomationStep`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List automationStepDescriptor = $convert.base64Decode(
    'Cg5BdXRvbWF0aW9uU3RlcBIbCgRuYW1lGAEgASgJQgf6QgRyAhABUgRuYW1lEjEKBHR5cGUYAi'
    'ABKA4yHS5hZ2VudHMudjEuQXV0b21hdGlvblN0ZXBUeXBlUgR0eXBlEkcKDGludm9rZV9hZ2Vu'
    'dBgDIAEoCzIkLmFnZW50cy52MS5BdXRvbWF0aW9uSW52b2tlQWdlbnRTdGVwUgtpbnZva2VBZ2'
    'VudBJHCgxjYWxsX3dlYmhvb2sYBCABKAsyJC5hZ2VudHMudjEuQXV0b21hdGlvbkNhbGxXZWJo'
    'b29rU3RlcFILY2FsbFdlYmhvb2sSVAoRc2VuZF9ub3RpZnlfZ3JvdXAYBSABKAsyKC5hZ2VudH'
    'MudjEuQXV0b21hdGlvblNlbmROb3RpZnlHcm91cFN0ZXBSD3NlbmROb3RpZnlHcm91cBJUChFj'
    'cmVhdGVfZm9ydW1fcG9zdBgGIAEoCzIoLmFnZW50cy52MS5BdXRvbWF0aW9uQ3JlYXRlRm9ydW'
    '1Qb3N0U3RlcFIPY3JlYXRlRm9ydW1Qb3N0EjMKBnBvbGljeRgKIAEoCzIbLmFnZW50cy52MS5B'
    'dXRvbWF0aW9uUG9saWN5UgZwb2xpY3k=');

@$core.Deprecated('Use automationInvokeAgentStepDescriptor instead')
const AutomationInvokeAgentStep$json = {
  '1': 'AutomationInvokeAgentStep',
  '2': [
    {'1': 'agent_name', '3': 1, '4': 1, '5': 9, '10': 'agentName'},
    {'1': 'agent_id', '3': 4, '4': 1, '5': 9, '10': 'agentId'},
    {'1': 'input', '3': 2, '4': 1, '5': 9, '10': 'input'},
    {'1': 'model_override', '3': 3, '4': 1, '5': 9, '10': 'modelOverride'},
  ],
};

/// Descriptor for `AutomationInvokeAgentStep`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List automationInvokeAgentStepDescriptor = $convert.base64Decode(
    'ChlBdXRvbWF0aW9uSW52b2tlQWdlbnRTdGVwEh0KCmFnZW50X25hbWUYASABKAlSCWFnZW50Tm'
    'FtZRIZCghhZ2VudF9pZBgEIAEoCVIHYWdlbnRJZBIUCgVpbnB1dBgCIAEoCVIFaW5wdXQSJQoO'
    'bW9kZWxfb3ZlcnJpZGUYAyABKAlSDW1vZGVsT3ZlcnJpZGU=');

@$core.Deprecated('Use automationCallWebhookStepDescriptor instead')
const AutomationCallWebhookStep$json = {
  '1': 'AutomationCallWebhookStep',
  '2': [
    {'1': 'url', '3': 1, '4': 1, '5': 9, '8': {}, '10': 'url'},
    {'1': 'method', '3': 2, '4': 1, '5': 9, '10': 'method'},
    {'1': 'payload_json', '3': 3, '4': 1, '5': 9, '10': 'payloadJson'},
    {
      '1': 'headers',
      '3': 4,
      '4': 3,
      '5': 11,
      '6': '.agents.v1.AutomationCallWebhookStep.HeadersEntry',
      '10': 'headers'
    },
  ],
  '3': [AutomationCallWebhookStep_HeadersEntry$json],
};

@$core.Deprecated('Use automationCallWebhookStepDescriptor instead')
const AutomationCallWebhookStep_HeadersEntry$json = {
  '1': 'HeadersEntry',
  '2': [
    {'1': 'key', '3': 1, '4': 1, '5': 9, '10': 'key'},
    {'1': 'value', '3': 2, '4': 1, '5': 9, '10': 'value'},
  ],
  '7': {'7': true},
};

/// Descriptor for `AutomationCallWebhookStep`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List automationCallWebhookStepDescriptor = $convert.base64Decode(
    'ChlBdXRvbWF0aW9uQ2FsbFdlYmhvb2tTdGVwEhkKA3VybBgBIAEoCUIH+kIEcgIQAVIDdXJsEh'
    'YKBm1ldGhvZBgCIAEoCVIGbWV0aG9kEiEKDHBheWxvYWRfanNvbhgDIAEoCVILcGF5bG9hZEpz'
    'b24SSwoHaGVhZGVycxgEIAMoCzIxLmFnZW50cy52MS5BdXRvbWF0aW9uQ2FsbFdlYmhvb2tTdG'
    'VwLkhlYWRlcnNFbnRyeVIHaGVhZGVycxo6CgxIZWFkZXJzRW50cnkSEAoDa2V5GAEgASgJUgNr'
    'ZXkSFAoFdmFsdWUYAiABKAlSBXZhbHVlOgI4AQ==');

@$core.Deprecated('Use automationSendNotifyGroupStepDescriptor instead')
const AutomationSendNotifyGroupStep$json = {
  '1': 'AutomationSendNotifyGroupStep',
  '2': [
    {
      '1': 'notify_group_name',
      '3': 1,
      '4': 1,
      '5': 9,
      '8': {},
      '10': 'notifyGroupName'
    },
    {'1': 'title', '3': 2, '4': 1, '5': 9, '10': 'title'},
    {'1': 'message', '3': 3, '4': 1, '5': 9, '10': 'message'},
  ],
};

/// Descriptor for `AutomationSendNotifyGroupStep`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List automationSendNotifyGroupStepDescriptor =
    $convert.base64Decode(
        'Ch1BdXRvbWF0aW9uU2VuZE5vdGlmeUdyb3VwU3RlcBIzChFub3RpZnlfZ3JvdXBfbmFtZRgBIA'
        'EoCUIH+kIEcgIQAVIPbm90aWZ5R3JvdXBOYW1lEhQKBXRpdGxlGAIgASgJUgV0aXRsZRIYCgdt'
        'ZXNzYWdlGAMgASgJUgdtZXNzYWdl');

@$core.Deprecated('Use automationCreateForumPostStepDescriptor instead')
const AutomationCreateForumPostStep$json = {
  '1': 'AutomationCreateForumPostStep',
  '2': [
    {'1': 'thread_id', '3': 1, '4': 1, '5': 9, '8': {}, '10': 'threadId'},
    {'1': 'body', '3': 2, '4': 1, '5': 9, '8': {}, '10': 'body'},
  ],
};

/// Descriptor for `AutomationCreateForumPostStep`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List automationCreateForumPostStepDescriptor =
    $convert.base64Decode(
        'Ch1BdXRvbWF0aW9uQ3JlYXRlRm9ydW1Qb3N0U3RlcBIkCgl0aHJlYWRfaWQYASABKAlCB/pCBH'
        'ICEAFSCHRocmVhZElkEhsKBGJvZHkYAiABKAlCB/pCBHICEAFSBGJvZHk=');

@$core.Deprecated('Use automationPolicyDescriptor instead')
const AutomationPolicy$json = {
  '1': 'AutomationPolicy',
  '2': [
    {
      '1': 'timeout',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.google.protobuf.Duration',
      '10': 'timeout'
    },
    {
      '1': 'retry',
      '3': 2,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.AutomationRetryPolicy',
      '10': 'retry'
    },
    {
      '1': 'concurrency',
      '3': 3,
      '4': 1,
      '5': 14,
      '6': '.agents.v1.AutomationConcurrencyPolicy',
      '10': 'concurrency'
    },
    {'1': 'max_output_bytes', '3': 4, '4': 1, '5': 5, '10': 'maxOutputBytes'},
  ],
};

/// Descriptor for `AutomationPolicy`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List automationPolicyDescriptor = $convert.base64Decode(
    'ChBBdXRvbWF0aW9uUG9saWN5EjMKB3RpbWVvdXQYASABKAsyGS5nb29nbGUucHJvdG9idWYuRH'
    'VyYXRpb25SB3RpbWVvdXQSNgoFcmV0cnkYAiABKAsyIC5hZ2VudHMudjEuQXV0b21hdGlvblJl'
    'dHJ5UG9saWN5UgVyZXRyeRJICgtjb25jdXJyZW5jeRgDIAEoDjImLmFnZW50cy52MS5BdXRvbW'
    'F0aW9uQ29uY3VycmVuY3lQb2xpY3lSC2NvbmN1cnJlbmN5EigKEG1heF9vdXRwdXRfYnl0ZXMY'
    'BCABKAVSDm1heE91dHB1dEJ5dGVz');

@$core.Deprecated('Use automationRetryPolicyDescriptor instead')
const AutomationRetryPolicy$json = {
  '1': 'AutomationRetryPolicy',
  '2': [
    {'1': 'max_attempts', '3': 1, '4': 1, '5': 5, '10': 'maxAttempts'},
    {
      '1': 'backoff',
      '3': 2,
      '4': 1,
      '5': 11,
      '6': '.google.protobuf.Duration',
      '10': 'backoff'
    },
  ],
};

/// Descriptor for `AutomationRetryPolicy`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List automationRetryPolicyDescriptor = $convert.base64Decode(
    'ChVBdXRvbWF0aW9uUmV0cnlQb2xpY3kSIQoMbWF4X2F0dGVtcHRzGAEgASgFUgttYXhBdHRlbX'
    'B0cxIzCgdiYWNrb2ZmGAIgASgLMhkuZ29vZ2xlLnByb3RvYnVmLkR1cmF0aW9uUgdiYWNrb2Zm');

@$core.Deprecated('Use automationRunDescriptor instead')
const AutomationRun$json = {
  '1': 'AutomationRun',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
    {'1': 'automation_name', '3': 2, '4': 1, '5': 9, '10': 'automationName'},
    {
      '1': 'trigger_type',
      '3': 3,
      '4': 1,
      '5': 14,
      '6': '.agents.v1.AutomationTriggerType',
      '10': 'triggerType'
    },
    {
      '1': 'status',
      '3': 4,
      '4': 1,
      '5': 14,
      '6': '.agents.v1.AutomationRunStatus',
      '10': 'status'
    },
    {
      '1': 'trigger_payload_json',
      '3': 5,
      '4': 1,
      '5': 9,
      '10': 'triggerPayloadJson'
    },
    {'1': 'error', '3': 6, '4': 1, '5': 9, '10': 'error'},
    {
      '1': 'started_at',
      '3': 7,
      '4': 1,
      '5': 11,
      '6': '.google.protobuf.Timestamp',
      '10': 'startedAt'
    },
    {
      '1': 'finished_at',
      '3': 8,
      '4': 1,
      '5': 11,
      '6': '.google.protobuf.Timestamp',
      '10': 'finishedAt'
    },
    {'1': 'duration_ms', '3': 9, '4': 1, '5': 3, '10': 'durationMs'},
    {'1': 'session_app_name', '3': 10, '4': 1, '5': 9, '10': 'sessionAppName'},
    {'1': 'session_user_id', '3': 11, '4': 1, '5': 9, '10': 'sessionUserId'},
    {'1': 'session_id', '3': 12, '4': 1, '5': 9, '10': 'sessionId'},
    {'1': 'agent_name', '3': 13, '4': 1, '5': 9, '10': 'agentName'},
    {'1': 'agent_id', '3': 14, '4': 1, '5': 9, '10': 'agentId'},
    {'1': 'workspace_id', '3': 100, '4': 1, '5': 9, '10': 'workspaceId'},
  ],
};

/// Descriptor for `AutomationRun`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List automationRunDescriptor = $convert.base64Decode(
    'Cg1BdXRvbWF0aW9uUnVuEg4KAmlkGAEgASgJUgJpZBInCg9hdXRvbWF0aW9uX25hbWUYAiABKA'
    'lSDmF1dG9tYXRpb25OYW1lEkMKDHRyaWdnZXJfdHlwZRgDIAEoDjIgLmFnZW50cy52MS5BdXRv'
    'bWF0aW9uVHJpZ2dlclR5cGVSC3RyaWdnZXJUeXBlEjYKBnN0YXR1cxgEIAEoDjIeLmFnZW50cy'
    '52MS5BdXRvbWF0aW9uUnVuU3RhdHVzUgZzdGF0dXMSMAoUdHJpZ2dlcl9wYXlsb2FkX2pzb24Y'
    'BSABKAlSEnRyaWdnZXJQYXlsb2FkSnNvbhIUCgVlcnJvchgGIAEoCVIFZXJyb3ISOQoKc3Rhcn'
    'RlZF9hdBgHIAEoCzIaLmdvb2dsZS5wcm90b2J1Zi5UaW1lc3RhbXBSCXN0YXJ0ZWRBdBI7Cgtm'
    'aW5pc2hlZF9hdBgIIAEoCzIaLmdvb2dsZS5wcm90b2J1Zi5UaW1lc3RhbXBSCmZpbmlzaGVkQX'
    'QSHwoLZHVyYXRpb25fbXMYCSABKANSCmR1cmF0aW9uTXMSKAoQc2Vzc2lvbl9hcHBfbmFtZRgK'
    'IAEoCVIOc2Vzc2lvbkFwcE5hbWUSJgoPc2Vzc2lvbl91c2VyX2lkGAsgASgJUg1zZXNzaW9uVX'
    'NlcklkEh0KCnNlc3Npb25faWQYDCABKAlSCXNlc3Npb25JZBIdCgphZ2VudF9uYW1lGA0gASgJ'
    'UglhZ2VudE5hbWUSGQoIYWdlbnRfaWQYDiABKAlSB2FnZW50SWQSIQoMd29ya3NwYWNlX2lkGG'
    'QgASgJUgt3b3Jrc3BhY2VJZA==');

@$core.Deprecated('Use automationStepRunDescriptor instead')
const AutomationStepRun$json = {
  '1': 'AutomationStepRun',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
    {'1': 'run_id', '3': 2, '4': 1, '5': 9, '10': 'runId'},
    {'1': 'automation_name', '3': 3, '4': 1, '5': 9, '10': 'automationName'},
    {'1': 'step_name', '3': 4, '4': 1, '5': 9, '10': 'stepName'},
    {
      '1': 'step_type',
      '3': 5,
      '4': 1,
      '5': 14,
      '6': '.agents.v1.AutomationStepType',
      '10': 'stepType'
    },
    {
      '1': 'status',
      '3': 6,
      '4': 1,
      '5': 14,
      '6': '.agents.v1.AutomationStepRunStatus',
      '10': 'status'
    },
    {'1': 'attempt_count', '3': 7, '4': 1, '5': 5, '10': 'attemptCount'},
    {'1': 'input_json', '3': 8, '4': 1, '5': 9, '10': 'inputJson'},
    {'1': 'output_json', '3': 9, '4': 1, '5': 9, '10': 'outputJson'},
    {'1': 'error', '3': 10, '4': 1, '5': 9, '10': 'error'},
    {'1': 'invocation_id', '3': 11, '4': 1, '5': 9, '10': 'invocationId'},
    {
      '1': 'started_at',
      '3': 12,
      '4': 1,
      '5': 11,
      '6': '.google.protobuf.Timestamp',
      '10': 'startedAt'
    },
    {
      '1': 'finished_at',
      '3': 13,
      '4': 1,
      '5': 11,
      '6': '.google.protobuf.Timestamp',
      '10': 'finishedAt'
    },
    {'1': 'duration_ms', '3': 14, '4': 1, '5': 3, '10': 'durationMs'},
    {'1': 'order', '3': 15, '4': 1, '5': 5, '10': 'order'},
    {'1': 'truncated', '3': 16, '4': 1, '5': 8, '10': 'truncated'},
    {'1': 'workspace_id', '3': 100, '4': 1, '5': 9, '10': 'workspaceId'},
  ],
};

/// Descriptor for `AutomationStepRun`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List automationStepRunDescriptor = $convert.base64Decode(
    'ChFBdXRvbWF0aW9uU3RlcFJ1bhIOCgJpZBgBIAEoCVICaWQSFQoGcnVuX2lkGAIgASgJUgVydW'
    '5JZBInCg9hdXRvbWF0aW9uX25hbWUYAyABKAlSDmF1dG9tYXRpb25OYW1lEhsKCXN0ZXBfbmFt'
    'ZRgEIAEoCVIIc3RlcE5hbWUSOgoJc3RlcF90eXBlGAUgASgOMh0uYWdlbnRzLnYxLkF1dG9tYX'
    'Rpb25TdGVwVHlwZVIIc3RlcFR5cGUSOgoGc3RhdHVzGAYgASgOMiIuYWdlbnRzLnYxLkF1dG9t'
    'YXRpb25TdGVwUnVuU3RhdHVzUgZzdGF0dXMSIwoNYXR0ZW1wdF9jb3VudBgHIAEoBVIMYXR0ZW'
    '1wdENvdW50Eh0KCmlucHV0X2pzb24YCCABKAlSCWlucHV0SnNvbhIfCgtvdXRwdXRfanNvbhgJ'
    'IAEoCVIKb3V0cHV0SnNvbhIUCgVlcnJvchgKIAEoCVIFZXJyb3ISIwoNaW52b2NhdGlvbl9pZB'
    'gLIAEoCVIMaW52b2NhdGlvbklkEjkKCnN0YXJ0ZWRfYXQYDCABKAsyGi5nb29nbGUucHJvdG9i'
    'dWYuVGltZXN0YW1wUglzdGFydGVkQXQSOwoLZmluaXNoZWRfYXQYDSABKAsyGi5nb29nbGUucH'
    'JvdG9idWYuVGltZXN0YW1wUgpmaW5pc2hlZEF0Eh8KC2R1cmF0aW9uX21zGA4gASgDUgpkdXJh'
    'dGlvbk1zEhQKBW9yZGVyGA8gASgFUgVvcmRlchIcCgl0cnVuY2F0ZWQYECABKAhSCXRydW5jYX'
    'RlZBIhCgx3b3Jrc3BhY2VfaWQYZCABKAlSC3dvcmtzcGFjZUlk');

@$core.Deprecated('Use listAutomationsRequestDescriptor instead')
const ListAutomationsRequest$json = {
  '1': 'ListAutomationsRequest',
};

/// Descriptor for `ListAutomationsRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listAutomationsRequestDescriptor =
    $convert.base64Decode('ChZMaXN0QXV0b21hdGlvbnNSZXF1ZXN0');

@$core.Deprecated('Use listAutomationsResponseDescriptor instead')
const ListAutomationsResponse$json = {
  '1': 'ListAutomationsResponse',
  '2': [
    {
      '1': 'automations',
      '3': 1,
      '4': 3,
      '5': 11,
      '6': '.agents.v1.Automation',
      '10': 'automations'
    },
  ],
};

/// Descriptor for `ListAutomationsResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listAutomationsResponseDescriptor =
    $convert.base64Decode(
        'ChdMaXN0QXV0b21hdGlvbnNSZXNwb25zZRI3CgthdXRvbWF0aW9ucxgBIAMoCzIVLmFnZW50cy'
        '52MS5BdXRvbWF0aW9uUgthdXRvbWF0aW9ucw==');

@$core.Deprecated('Use getAutomationRequestDescriptor instead')
const GetAutomationRequest$json = {
  '1': 'GetAutomationRequest',
  '2': [
    {'1': 'name', '3': 1, '4': 1, '5': 9, '8': {}, '10': 'name'},
  ],
};

/// Descriptor for `GetAutomationRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getAutomationRequestDescriptor =
    $convert.base64Decode(
        'ChRHZXRBdXRvbWF0aW9uUmVxdWVzdBIbCgRuYW1lGAEgASgJQgf6QgRyAhABUgRuYW1l');

@$core.Deprecated('Use getAutomationResponseDescriptor instead')
const GetAutomationResponse$json = {
  '1': 'GetAutomationResponse',
  '2': [
    {
      '1': 'automation',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.Automation',
      '10': 'automation'
    },
  ],
};

/// Descriptor for `GetAutomationResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getAutomationResponseDescriptor = $convert.base64Decode(
    'ChVHZXRBdXRvbWF0aW9uUmVzcG9uc2USNQoKYXV0b21hdGlvbhgBIAEoCzIVLmFnZW50cy52MS'
    '5BdXRvbWF0aW9uUgphdXRvbWF0aW9u');

@$core.Deprecated('Use createAutomationRequestDescriptor instead')
const CreateAutomationRequest$json = {
  '1': 'CreateAutomationRequest',
  '2': [
    {
      '1': 'automation',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.Automation',
      '10': 'automation'
    },
  ],
};

/// Descriptor for `CreateAutomationRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List createAutomationRequestDescriptor =
    $convert.base64Decode(
        'ChdDcmVhdGVBdXRvbWF0aW9uUmVxdWVzdBI1CgphdXRvbWF0aW9uGAEgASgLMhUuYWdlbnRzLn'
        'YxLkF1dG9tYXRpb25SCmF1dG9tYXRpb24=');

@$core.Deprecated('Use createAutomationResponseDescriptor instead')
const CreateAutomationResponse$json = {
  '1': 'CreateAutomationResponse',
  '2': [
    {
      '1': 'automation',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.Automation',
      '10': 'automation'
    },
  ],
};

/// Descriptor for `CreateAutomationResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List createAutomationResponseDescriptor =
    $convert.base64Decode(
        'ChhDcmVhdGVBdXRvbWF0aW9uUmVzcG9uc2USNQoKYXV0b21hdGlvbhgBIAEoCzIVLmFnZW50cy'
        '52MS5BdXRvbWF0aW9uUgphdXRvbWF0aW9u');

@$core.Deprecated('Use updateAutomationRequestDescriptor instead')
const UpdateAutomationRequest$json = {
  '1': 'UpdateAutomationRequest',
  '2': [
    {
      '1': 'automation',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.Automation',
      '10': 'automation'
    },
  ],
};

/// Descriptor for `UpdateAutomationRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List updateAutomationRequestDescriptor =
    $convert.base64Decode(
        'ChdVcGRhdGVBdXRvbWF0aW9uUmVxdWVzdBI1CgphdXRvbWF0aW9uGAEgASgLMhUuYWdlbnRzLn'
        'YxLkF1dG9tYXRpb25SCmF1dG9tYXRpb24=');

@$core.Deprecated('Use updateAutomationResponseDescriptor instead')
const UpdateAutomationResponse$json = {
  '1': 'UpdateAutomationResponse',
  '2': [
    {
      '1': 'automation',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.Automation',
      '10': 'automation'
    },
  ],
};

/// Descriptor for `UpdateAutomationResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List updateAutomationResponseDescriptor =
    $convert.base64Decode(
        'ChhVcGRhdGVBdXRvbWF0aW9uUmVzcG9uc2USNQoKYXV0b21hdGlvbhgBIAEoCzIVLmFnZW50cy'
        '52MS5BdXRvbWF0aW9uUgphdXRvbWF0aW9u');

@$core.Deprecated('Use deleteAutomationRequestDescriptor instead')
const DeleteAutomationRequest$json = {
  '1': 'DeleteAutomationRequest',
  '2': [
    {'1': 'name', '3': 1, '4': 1, '5': 9, '8': {}, '10': 'name'},
  ],
};

/// Descriptor for `DeleteAutomationRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List deleteAutomationRequestDescriptor =
    $convert.base64Decode(
        'ChdEZWxldGVBdXRvbWF0aW9uUmVxdWVzdBIbCgRuYW1lGAEgASgJQgf6QgRyAhABUgRuYW1l');

@$core.Deprecated('Use deleteAutomationResponseDescriptor instead')
const DeleteAutomationResponse$json = {
  '1': 'DeleteAutomationResponse',
  '2': [
    {
      '1': 'automation',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.Automation',
      '10': 'automation'
    },
  ],
};

/// Descriptor for `DeleteAutomationResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List deleteAutomationResponseDescriptor =
    $convert.base64Decode(
        'ChhEZWxldGVBdXRvbWF0aW9uUmVzcG9uc2USNQoKYXV0b21hdGlvbhgBIAEoCzIVLmFnZW50cy'
        '52MS5BdXRvbWF0aW9uUgphdXRvbWF0aW9u');

@$core.Deprecated('Use runAutomationNowRequestDescriptor instead')
const RunAutomationNowRequest$json = {
  '1': 'RunAutomationNowRequest',
  '2': [
    {'1': 'name', '3': 1, '4': 1, '5': 9, '8': {}, '10': 'name'},
    {
      '1': 'trigger_payload_json',
      '3': 2,
      '4': 1,
      '5': 9,
      '10': 'triggerPayloadJson'
    },
  ],
};

/// Descriptor for `RunAutomationNowRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List runAutomationNowRequestDescriptor =
    $convert.base64Decode(
        'ChdSdW5BdXRvbWF0aW9uTm93UmVxdWVzdBIbCgRuYW1lGAEgASgJQgf6QgRyAhABUgRuYW1lEj'
        'AKFHRyaWdnZXJfcGF5bG9hZF9qc29uGAIgASgJUhJ0cmlnZ2VyUGF5bG9hZEpzb24=');

@$core.Deprecated('Use runAutomationNowResponseDescriptor instead')
const RunAutomationNowResponse$json = {
  '1': 'RunAutomationNowResponse',
  '2': [
    {
      '1': 'run',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.AutomationRun',
      '10': 'run'
    },
  ],
};

/// Descriptor for `RunAutomationNowResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List runAutomationNowResponseDescriptor =
    $convert.base64Decode(
        'ChhSdW5BdXRvbWF0aW9uTm93UmVzcG9uc2USKgoDcnVuGAEgASgLMhguYWdlbnRzLnYxLkF1dG'
        '9tYXRpb25SdW5SA3J1bg==');

@$core.Deprecated('Use listAutomationRunsRequestDescriptor instead')
const ListAutomationRunsRequest$json = {
  '1': 'ListAutomationRunsRequest',
  '2': [
    {'1': 'automation_name', '3': 1, '4': 1, '5': 9, '10': 'automationName'},
    {'1': 'page_size', '3': 2, '4': 1, '5': 5, '10': 'pageSize'},
    {'1': 'page_token', '3': 3, '4': 1, '5': 9, '10': 'pageToken'},
  ],
};

/// Descriptor for `ListAutomationRunsRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listAutomationRunsRequestDescriptor = $convert.base64Decode(
    'ChlMaXN0QXV0b21hdGlvblJ1bnNSZXF1ZXN0EicKD2F1dG9tYXRpb25fbmFtZRgBIAEoCVIOYX'
    'V0b21hdGlvbk5hbWUSGwoJcGFnZV9zaXplGAIgASgFUghwYWdlU2l6ZRIdCgpwYWdlX3Rva2Vu'
    'GAMgASgJUglwYWdlVG9rZW4=');

@$core.Deprecated('Use listAutomationRunsResponseDescriptor instead')
const ListAutomationRunsResponse$json = {
  '1': 'ListAutomationRunsResponse',
  '2': [
    {
      '1': 'runs',
      '3': 1,
      '4': 3,
      '5': 11,
      '6': '.agents.v1.AutomationRun',
      '10': 'runs'
    },
    {'1': 'next_page_token', '3': 2, '4': 1, '5': 9, '10': 'nextPageToken'},
  ],
};

/// Descriptor for `ListAutomationRunsResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listAutomationRunsResponseDescriptor =
    $convert.base64Decode(
        'ChpMaXN0QXV0b21hdGlvblJ1bnNSZXNwb25zZRIsCgRydW5zGAEgAygLMhguYWdlbnRzLnYxLk'
        'F1dG9tYXRpb25SdW5SBHJ1bnMSJgoPbmV4dF9wYWdlX3Rva2VuGAIgASgJUg1uZXh0UGFnZVRv'
        'a2Vu');

@$core.Deprecated('Use getAutomationRunRequestDescriptor instead')
const GetAutomationRunRequest$json = {
  '1': 'GetAutomationRunRequest',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '8': {}, '10': 'id'},
  ],
};

/// Descriptor for `GetAutomationRunRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getAutomationRunRequestDescriptor =
    $convert.base64Decode(
        'ChdHZXRBdXRvbWF0aW9uUnVuUmVxdWVzdBIXCgJpZBgBIAEoCUIH+kIEcgIQAVICaWQ=');

@$core.Deprecated('Use getAutomationRunResponseDescriptor instead')
const GetAutomationRunResponse$json = {
  '1': 'GetAutomationRunResponse',
  '2': [
    {
      '1': 'run',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.AutomationRun',
      '10': 'run'
    },
  ],
};

/// Descriptor for `GetAutomationRunResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getAutomationRunResponseDescriptor =
    $convert.base64Decode(
        'ChhHZXRBdXRvbWF0aW9uUnVuUmVzcG9uc2USKgoDcnVuGAEgASgLMhguYWdlbnRzLnYxLkF1dG'
        '9tYXRpb25SdW5SA3J1bg==');

@$core.Deprecated('Use listAutomationStepRunsRequestDescriptor instead')
const ListAutomationStepRunsRequest$json = {
  '1': 'ListAutomationStepRunsRequest',
  '2': [
    {'1': 'run_id', '3': 1, '4': 1, '5': 9, '8': {}, '10': 'runId'},
  ],
};

/// Descriptor for `ListAutomationStepRunsRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listAutomationStepRunsRequestDescriptor =
    $convert.base64Decode(
        'Ch1MaXN0QXV0b21hdGlvblN0ZXBSdW5zUmVxdWVzdBIeCgZydW5faWQYASABKAlCB/pCBHICEA'
        'FSBXJ1bklk');

@$core.Deprecated('Use listAutomationStepRunsResponseDescriptor instead')
const ListAutomationStepRunsResponse$json = {
  '1': 'ListAutomationStepRunsResponse',
  '2': [
    {
      '1': 'step_runs',
      '3': 1,
      '4': 3,
      '5': 11,
      '6': '.agents.v1.AutomationStepRun',
      '10': 'stepRuns'
    },
  ],
};

/// Descriptor for `ListAutomationStepRunsResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listAutomationStepRunsResponseDescriptor =
    $convert.base64Decode(
        'Ch5MaXN0QXV0b21hdGlvblN0ZXBSdW5zUmVzcG9uc2USOQoJc3RlcF9ydW5zGAEgAygLMhwuYW'
        'dlbnRzLnYxLkF1dG9tYXRpb25TdGVwUnVuUghzdGVwUnVucw==');

const $core.Map<$core.String, $core.dynamic> AutomationServiceBase$json = {
  '1': 'AutomationService',
  '2': [
    {
      '1': 'ListAutomations',
      '2': '.agents.v1.ListAutomationsRequest',
      '3': '.agents.v1.ListAutomationsResponse'
    },
    {
      '1': 'GetAutomation',
      '2': '.agents.v1.GetAutomationRequest',
      '3': '.agents.v1.GetAutomationResponse'
    },
    {
      '1': 'CreateAutomation',
      '2': '.agents.v1.CreateAutomationRequest',
      '3': '.agents.v1.CreateAutomationResponse'
    },
    {
      '1': 'UpdateAutomation',
      '2': '.agents.v1.UpdateAutomationRequest',
      '3': '.agents.v1.UpdateAutomationResponse'
    },
    {
      '1': 'DeleteAutomation',
      '2': '.agents.v1.DeleteAutomationRequest',
      '3': '.agents.v1.DeleteAutomationResponse'
    },
    {
      '1': 'RunAutomationNow',
      '2': '.agents.v1.RunAutomationNowRequest',
      '3': '.agents.v1.RunAutomationNowResponse'
    },
    {
      '1': 'ListAutomationRuns',
      '2': '.agents.v1.ListAutomationRunsRequest',
      '3': '.agents.v1.ListAutomationRunsResponse'
    },
    {
      '1': 'GetAutomationRun',
      '2': '.agents.v1.GetAutomationRunRequest',
      '3': '.agents.v1.GetAutomationRunResponse'
    },
    {
      '1': 'ListAutomationStepRuns',
      '2': '.agents.v1.ListAutomationStepRunsRequest',
      '3': '.agents.v1.ListAutomationStepRunsResponse'
    },
  ],
};

@$core.Deprecated('Use automationServiceDescriptor instead')
const $core.Map<$core.String, $core.Map<$core.String, $core.dynamic>>
    AutomationServiceBase$messageJson = {
  '.agents.v1.ListAutomationsRequest': ListAutomationsRequest$json,
  '.agents.v1.ListAutomationsResponse': ListAutomationsResponse$json,
  '.agents.v1.Automation': Automation$json,
  '.agents.v1.AutomationTrigger': AutomationTrigger$json,
  '.agents.v1.AutomationScheduleTrigger': AutomationScheduleTrigger$json,
  '.agents.v1.AutomationWebhookTrigger': AutomationWebhookTrigger$json,
  '.agents.v1.AutomationEventTrigger': AutomationEventTrigger$json,
  '.agents.v1.AutomationEventTrigger.FiltersEntry':
      AutomationEventTrigger_FiltersEntry$json,
  '.agents.v1.AutomationCondition': AutomationCondition$json,
  '.agents.v1.AutomationStep': AutomationStep$json,
  '.agents.v1.AutomationInvokeAgentStep': AutomationInvokeAgentStep$json,
  '.agents.v1.AutomationCallWebhookStep': AutomationCallWebhookStep$json,
  '.agents.v1.AutomationCallWebhookStep.HeadersEntry':
      AutomationCallWebhookStep_HeadersEntry$json,
  '.agents.v1.AutomationSendNotifyGroupStep':
      AutomationSendNotifyGroupStep$json,
  '.agents.v1.AutomationCreateForumPostStep':
      AutomationCreateForumPostStep$json,
  '.agents.v1.AutomationPolicy': AutomationPolicy$json,
  '.google.protobuf.Duration': $1.Duration$json,
  '.agents.v1.AutomationRetryPolicy': AutomationRetryPolicy$json,
  '.agents.v1.Automation.MetadataEntry': Automation_MetadataEntry$json,
  '.google.protobuf.Timestamp': $0.Timestamp$json,
  '.agents.v1.GetAutomationRequest': GetAutomationRequest$json,
  '.agents.v1.GetAutomationResponse': GetAutomationResponse$json,
  '.agents.v1.CreateAutomationRequest': CreateAutomationRequest$json,
  '.agents.v1.CreateAutomationResponse': CreateAutomationResponse$json,
  '.agents.v1.UpdateAutomationRequest': UpdateAutomationRequest$json,
  '.agents.v1.UpdateAutomationResponse': UpdateAutomationResponse$json,
  '.agents.v1.DeleteAutomationRequest': DeleteAutomationRequest$json,
  '.agents.v1.DeleteAutomationResponse': DeleteAutomationResponse$json,
  '.agents.v1.RunAutomationNowRequest': RunAutomationNowRequest$json,
  '.agents.v1.RunAutomationNowResponse': RunAutomationNowResponse$json,
  '.agents.v1.AutomationRun': AutomationRun$json,
  '.agents.v1.ListAutomationRunsRequest': ListAutomationRunsRequest$json,
  '.agents.v1.ListAutomationRunsResponse': ListAutomationRunsResponse$json,
  '.agents.v1.GetAutomationRunRequest': GetAutomationRunRequest$json,
  '.agents.v1.GetAutomationRunResponse': GetAutomationRunResponse$json,
  '.agents.v1.ListAutomationStepRunsRequest':
      ListAutomationStepRunsRequest$json,
  '.agents.v1.ListAutomationStepRunsResponse':
      ListAutomationStepRunsResponse$json,
  '.agents.v1.AutomationStepRun': AutomationStepRun$json,
};

/// Descriptor for `AutomationService`. Decode as a `google.protobuf.ServiceDescriptorProto`.
final $typed_data.Uint8List automationServiceDescriptor = $convert.base64Decode(
    'ChFBdXRvbWF0aW9uU2VydmljZRJYCg9MaXN0QXV0b21hdGlvbnMSIS5hZ2VudHMudjEuTGlzdE'
    'F1dG9tYXRpb25zUmVxdWVzdBoiLmFnZW50cy52MS5MaXN0QXV0b21hdGlvbnNSZXNwb25zZRJS'
    'Cg1HZXRBdXRvbWF0aW9uEh8uYWdlbnRzLnYxLkdldEF1dG9tYXRpb25SZXF1ZXN0GiAuYWdlbn'
    'RzLnYxLkdldEF1dG9tYXRpb25SZXNwb25zZRJbChBDcmVhdGVBdXRvbWF0aW9uEiIuYWdlbnRz'
    'LnYxLkNyZWF0ZUF1dG9tYXRpb25SZXF1ZXN0GiMuYWdlbnRzLnYxLkNyZWF0ZUF1dG9tYXRpb2'
    '5SZXNwb25zZRJbChBVcGRhdGVBdXRvbWF0aW9uEiIuYWdlbnRzLnYxLlVwZGF0ZUF1dG9tYXRp'
    'b25SZXF1ZXN0GiMuYWdlbnRzLnYxLlVwZGF0ZUF1dG9tYXRpb25SZXNwb25zZRJbChBEZWxldG'
    'VBdXRvbWF0aW9uEiIuYWdlbnRzLnYxLkRlbGV0ZUF1dG9tYXRpb25SZXF1ZXN0GiMuYWdlbnRz'
    'LnYxLkRlbGV0ZUF1dG9tYXRpb25SZXNwb25zZRJbChBSdW5BdXRvbWF0aW9uTm93EiIuYWdlbn'
    'RzLnYxLlJ1bkF1dG9tYXRpb25Ob3dSZXF1ZXN0GiMuYWdlbnRzLnYxLlJ1bkF1dG9tYXRpb25O'
    'b3dSZXNwb25zZRJhChJMaXN0QXV0b21hdGlvblJ1bnMSJC5hZ2VudHMudjEuTGlzdEF1dG9tYX'
    'Rpb25SdW5zUmVxdWVzdBolLmFnZW50cy52MS5MaXN0QXV0b21hdGlvblJ1bnNSZXNwb25zZRJb'
    'ChBHZXRBdXRvbWF0aW9uUnVuEiIuYWdlbnRzLnYxLkdldEF1dG9tYXRpb25SdW5SZXF1ZXN0Gi'
    'MuYWdlbnRzLnYxLkdldEF1dG9tYXRpb25SdW5SZXNwb25zZRJtChZMaXN0QXV0b21hdGlvblN0'
    'ZXBSdW5zEiguYWdlbnRzLnYxLkxpc3RBdXRvbWF0aW9uU3RlcFJ1bnNSZXF1ZXN0GikuYWdlbn'
    'RzLnYxLkxpc3RBdXRvbWF0aW9uU3RlcFJ1bnNSZXNwb25zZQ==');
