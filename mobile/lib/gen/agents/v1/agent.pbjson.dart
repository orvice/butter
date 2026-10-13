// This is a generated file - do not edit.
//
// Generated from agents/v1/agent.proto.

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

@$core.Deprecated('Use workflowNodeKindDescriptor instead')
const WorkflowNodeKind$json = {
  '1': 'WorkflowNodeKind',
  '2': [
    {'1': 'WORKFLOW_NODE_KIND_UNSPECIFIED', '2': 0},
    {'1': 'WORKFLOW_NODE_KIND_AGENT', '2': 1},
    {'1': 'WORKFLOW_NODE_KIND_HUMAN_INPUT', '2': 2},
    {'1': 'WORKFLOW_NODE_KIND_ROUTER', '2': 3},
    {'1': 'WORKFLOW_NODE_KIND_JOIN', '2': 4},
  ],
};

/// Descriptor for `WorkflowNodeKind`. Decode as a `google.protobuf.EnumDescriptorProto`.
final $typed_data.Uint8List workflowNodeKindDescriptor = $convert.base64Decode(
    'ChBXb3JrZmxvd05vZGVLaW5kEiIKHldPUktGTE9XX05PREVfS0lORF9VTlNQRUNJRklFRBAAEh'
    'wKGFdPUktGTE9XX05PREVfS0lORF9BR0VOVBABEiIKHldPUktGTE9XX05PREVfS0lORF9IVU1B'
    'Tl9JTlBVVBACEh0KGVdPUktGTE9XX05PREVfS0lORF9ST1VURVIQAxIbChdXT1JLRkxPV19OT0'
    'RFX0tJTkRfSk9JThAE');

@$core.Deprecated('Use humanInputFormFieldTypeDescriptor instead')
const HumanInputFormFieldType$json = {
  '1': 'HumanInputFormFieldType',
  '2': [
    {'1': 'HUMAN_INPUT_FORM_FIELD_TYPE_UNSPECIFIED', '2': 0},
    {'1': 'HUMAN_INPUT_FORM_FIELD_TYPE_TEXT', '2': 1},
    {'1': 'HUMAN_INPUT_FORM_FIELD_TYPE_SINGLE_CHOICE', '2': 2},
  ],
};

/// Descriptor for `HumanInputFormFieldType`. Decode as a `google.protobuf.EnumDescriptorProto`.
final $typed_data.Uint8List humanInputFormFieldTypeDescriptor = $convert.base64Decode(
    'ChdIdW1hbklucHV0Rm9ybUZpZWxkVHlwZRIrCidIVU1BTl9JTlBVVF9GT1JNX0ZJRUxEX1RZUE'
    'VfVU5TUEVDSUZJRUQQABIkCiBIVU1BTl9JTlBVVF9GT1JNX0ZJRUxEX1RZUEVfVEVYVBABEi0K'
    'KUhVTUFOX0lOUFVUX0ZPUk1fRklFTERfVFlQRV9TSU5HTEVfQ0hPSUNFEAI=');

@$core.Deprecated('Use mCPServerAuthTypeDescriptor instead')
const MCPServerAuthType$json = {
  '1': 'MCPServerAuthType',
  '2': [
    {'1': 'MCP_SERVER_AUTH_TYPE_UNSPECIFIED', '2': 0},
    {'1': 'MCP_SERVER_AUTH_TYPE_NONE', '2': 1},
    {'1': 'MCP_SERVER_AUTH_TYPE_STATIC_HEADERS', '2': 2},
    {'1': 'MCP_SERVER_AUTH_TYPE_OAUTH2', '2': 3},
  ],
};

/// Descriptor for `MCPServerAuthType`. Decode as a `google.protobuf.EnumDescriptorProto`.
final $typed_data.Uint8List mCPServerAuthTypeDescriptor = $convert.base64Decode(
    'ChFNQ1BTZXJ2ZXJBdXRoVHlwZRIkCiBNQ1BfU0VSVkVSX0FVVEhfVFlQRV9VTlNQRUNJRklFRB'
    'AAEh0KGU1DUF9TRVJWRVJfQVVUSF9UWVBFX05PTkUQARInCiNNQ1BfU0VSVkVSX0FVVEhfVFlQ'
    'RV9TVEFUSUNfSEVBREVSUxACEh8KG01DUF9TRVJWRVJfQVVUSF9UWVBFX09BVVRIMhAD');

@$core.Deprecated('Use streamingModeDescriptor instead')
const StreamingMode$json = {
  '1': 'StreamingMode',
  '2': [
    {'1': 'STREAMING_MODE_UNSPECIFIED', '2': 0},
    {'1': 'STREAMING_MODE_NONE', '2': 1},
    {'1': 'STREAMING_MODE_SSE', '2': 2},
  ],
};

/// Descriptor for `StreamingMode`. Decode as a `google.protobuf.EnumDescriptorProto`.
final $typed_data.Uint8List streamingModeDescriptor = $convert.base64Decode(
    'Cg1TdHJlYW1pbmdNb2RlEh4KGlNUUkVBTUlOR19NT0RFX1VOU1BFQ0lGSUVEEAASFwoTU1RSRU'
    'FNSU5HX01PREVfTk9ORRABEhYKElNUUkVBTUlOR19NT0RFX1NTRRAC');

@$core.Deprecated('Use agentLifecycleStatusDescriptor instead')
const AgentLifecycleStatus$json = {
  '1': 'AgentLifecycleStatus',
  '2': [
    {'1': 'AGENT_LIFECYCLE_STATUS_UNSPECIFIED', '2': 0},
    {'1': 'AGENT_LIFECYCLE_STATUS_ACTIVE', '2': 1},
    {'1': 'AGENT_LIFECYCLE_STATUS_MIGRATION_REQUIRED', '2': 2},
    {'1': 'AGENT_LIFECYCLE_STATUS_PROVISIONING', '2': 3},
    {'1': 'AGENT_LIFECYCLE_STATUS_ERROR', '2': 4},
    {'1': 'AGENT_LIFECYCLE_STATUS_DELETING', '2': 5},
    {'1': 'AGENT_LIFECYCLE_STATUS_DELETED', '2': 6},
  ],
};

/// Descriptor for `AgentLifecycleStatus`. Decode as a `google.protobuf.EnumDescriptorProto`.
final $typed_data.Uint8List agentLifecycleStatusDescriptor = $convert.base64Decode(
    'ChRBZ2VudExpZmVjeWNsZVN0YXR1cxImCiJBR0VOVF9MSUZFQ1lDTEVfU1RBVFVTX1VOU1BFQ0'
    'lGSUVEEAASIQodQUdFTlRfTElGRUNZQ0xFX1NUQVRVU19BQ1RJVkUQARItCilBR0VOVF9MSUZF'
    'Q1lDTEVfU1RBVFVTX01JR1JBVElPTl9SRVFVSVJFRBACEicKI0FHRU5UX0xJRkVDWUNMRV9TVE'
    'FUVVNfUFJPVklTSU9OSU5HEAMSIAocQUdFTlRfTElGRUNZQ0xFX1NUQVRVU19FUlJPUhAEEiMK'
    'H0FHRU5UX0xJRkVDWUNMRV9TVEFUVVNfREVMRVRJTkcQBRIiCh5BR0VOVF9MSUZFQ1lDTEVfU1'
    'RBVFVTX0RFTEVURUQQBg==');

@$core.Deprecated('Use agentTypeDescriptor instead')
const AgentType$json = {
  '1': 'AgentType',
  '2': [
    {'1': 'AGENT_TYPE_UNSPECIFIED', '2': 0},
    {'1': 'AGENT_TYPE_LLM', '2': 1},
    {'1': 'AGENT_TYPE_LOOP', '2': 2},
    {'1': 'AGENT_TYPE_SEQUENTIAL', '2': 3},
    {'1': 'AGENT_TYPE_PARALLEL', '2': 4},
    {'1': 'AGENT_TYPE_WORKFLOW', '2': 5},
    {'1': 'AGENT_TYPE_PI', '2': 6},
    {'1': 'AGENT_TYPE_CURSOR', '2': 7},
  ],
};

/// Descriptor for `AgentType`. Decode as a `google.protobuf.EnumDescriptorProto`.
final $typed_data.Uint8List agentTypeDescriptor = $convert.base64Decode(
    'CglBZ2VudFR5cGUSGgoWQUdFTlRfVFlQRV9VTlNQRUNJRklFRBAAEhIKDkFHRU5UX1RZUEVfTE'
    'xNEAESEwoPQUdFTlRfVFlQRV9MT09QEAISGQoVQUdFTlRfVFlQRV9TRVFVRU5USUFMEAMSFwoT'
    'QUdFTlRfVFlQRV9QQVJBTExFTBAEEhcKE0FHRU5UX1RZUEVfV09SS0ZMT1cQBRIRCg1BR0VOVF'
    '9UWVBFX1BJEAYSFQoRQUdFTlRfVFlQRV9DVVJTT1IQBw==');

@$core.Deprecated('Use lLMIncludeContentsDescriptor instead')
const LLMIncludeContents$json = {
  '1': 'LLMIncludeContents',
  '2': [
    {'1': 'LLM_INCLUDE_CONTENTS_UNSPECIFIED', '2': 0},
    {'1': 'LLM_INCLUDE_CONTENTS_DEFAULT', '2': 1},
    {'1': 'LLM_INCLUDE_CONTENTS_NONE', '2': 2},
  ],
};

/// Descriptor for `LLMIncludeContents`. Decode as a `google.protobuf.EnumDescriptorProto`.
final $typed_data.Uint8List lLMIncludeContentsDescriptor = $convert.base64Decode(
    'ChJMTE1JbmNsdWRlQ29udGVudHMSJAogTExNX0lOQ0xVREVfQ09OVEVOVFNfVU5TUEVDSUZJRU'
    'QQABIgChxMTE1fSU5DTFVERV9DT05URU5UU19ERUZBVUxUEAESHQoZTExNX0lOQ0xVREVfQ09O'
    'VEVOVFNfTk9ORRAC');

@$core.Deprecated('Use mCPServerTransportDescriptor instead')
const MCPServerTransport$json = {
  '1': 'MCPServerTransport',
  '2': [
    {'1': 'MCP_SERVER_TRANSPORT_UNSPECIFIED', '2': 0},
    {'1': 'MCP_SERVER_TRANSPORT_STREAMABLE_HTTP', '2': 2},
    {'1': 'MCP_SERVER_TRANSPORT_SSE', '2': 3},
  ],
  '4': [
    {'1': 1, '2': 1},
  ],
  '5': ['MCP_SERVER_TRANSPORT_STDIO'],
};

/// Descriptor for `MCPServerTransport`. Decode as a `google.protobuf.EnumDescriptorProto`.
final $typed_data.Uint8List mCPServerTransportDescriptor = $convert.base64Decode(
    'ChJNQ1BTZXJ2ZXJUcmFuc3BvcnQSJAogTUNQX1NFUlZFUl9UUkFOU1BPUlRfVU5TUEVDSUZJRU'
    'QQABIoCiRNQ1BfU0VSVkVSX1RSQU5TUE9SVF9TVFJFQU1BQkxFX0hUVFAQAhIcChhNQ1BfU0VS'
    'VkVSX1RSQU5TUE9SVF9TU0UQAyIECAEQASoaTUNQX1NFUlZFUl9UUkFOU1BPUlRfU1RESU8=');

@$core.Deprecated('Use remoteAgentProtocolDescriptor instead')
const RemoteAgentProtocol$json = {
  '1': 'RemoteAgentProtocol',
  '2': [
    {'1': 'REMOTE_AGENT_PROTOCOL_UNSPECIFIED', '2': 0},
    {'1': 'REMOTE_AGENT_PROTOCOL_A2A', '2': 1},
    {'1': 'REMOTE_AGENT_PROTOCOL_DAEMON', '2': 2},
    {'1': 'REMOTE_AGENT_PROTOCOL_OPENCODE_HTTP', '2': 3},
  ],
};

/// Descriptor for `RemoteAgentProtocol`. Decode as a `google.protobuf.EnumDescriptorProto`.
final $typed_data.Uint8List remoteAgentProtocolDescriptor = $convert.base64Decode(
    'ChNSZW1vdGVBZ2VudFByb3RvY29sEiUKIVJFTU9URV9BR0VOVF9QUk9UT0NPTF9VTlNQRUNJRk'
    'lFRBAAEh0KGVJFTU9URV9BR0VOVF9QUk9UT0NPTF9BMkEQARIgChxSRU1PVEVfQUdFTlRfUFJP'
    'VE9DT0xfREFFTU9OEAISJwojUkVNT1RFX0FHRU5UX1BST1RPQ09MX09QRU5DT0RFX0hUVFAQAw'
    '==');

@$core.Deprecated('Use contextGuardStrategyDescriptor instead')
const ContextGuardStrategy$json = {
  '1': 'ContextGuardStrategy',
  '2': [
    {'1': 'CONTEXT_GUARD_STRATEGY_UNSPECIFIED', '2': 0},
    {'1': 'CONTEXT_GUARD_STRATEGY_THRESHOLD', '2': 1},
    {'1': 'CONTEXT_GUARD_STRATEGY_SLIDING_WINDOW', '2': 2},
  ],
};

/// Descriptor for `ContextGuardStrategy`. Decode as a `google.protobuf.EnumDescriptorProto`.
final $typed_data.Uint8List contextGuardStrategyDescriptor = $convert.base64Decode(
    'ChRDb250ZXh0R3VhcmRTdHJhdGVneRImCiJDT05URVhUX0dVQVJEX1NUUkFURUdZX1VOU1BFQ0'
    'lGSUVEEAASJAogQ09OVEVYVF9HVUFSRF9TVFJBVEVHWV9USFJFU0hPTEQQARIpCiVDT05URVhU'
    'X0dVQVJEX1NUUkFURUdZX1NMSURJTkdfV0lORE9XEAI=');

@$core.Deprecated('Use resultCardGenerationDescriptor instead')
const ResultCardGeneration$json = {
  '1': 'ResultCardGeneration',
  '2': [
    {'1': 'RESULT_CARD_GENERATION_UNSPECIFIED', '2': 0},
    {'1': 'RESULT_CARD_GENERATION_DISABLED', '2': 1},
  ],
};

/// Descriptor for `ResultCardGeneration`. Decode as a `google.protobuf.EnumDescriptorProto`.
final $typed_data.Uint8List resultCardGenerationDescriptor = $convert.base64Decode(
    'ChRSZXN1bHRDYXJkR2VuZXJhdGlvbhImCiJSRVNVTFRfQ0FSRF9HRU5FUkFUSU9OX1VOU1BFQ0'
    'lGSUVEEAASIwofUkVTVUxUX0NBUkRfR0VORVJBVElPTl9ESVNBQkxFRBAB');

@$core.Deprecated('Use resultCardPresentationDescriptor instead')
const ResultCardPresentation$json = {
  '1': 'ResultCardPresentation',
  '2': [
    {'1': 'RESULT_CARD_PRESENTATION_UNSPECIFIED', '2': 0},
    {'1': 'RESULT_CARD_PRESENTATION_AUTO', '2': 1},
    {'1': 'RESULT_CARD_PRESENTATION_PREFERRED', '2': 2},
  ],
};

/// Descriptor for `ResultCardPresentation`. Decode as a `google.protobuf.EnumDescriptorProto`.
final $typed_data.Uint8List resultCardPresentationDescriptor = $convert.base64Decode(
    'ChZSZXN1bHRDYXJkUHJlc2VudGF0aW9uEigKJFJFU1VMVF9DQVJEX1BSRVNFTlRBVElPTl9VTl'
    'NQRUNJRklFRBAAEiEKHVJFU1VMVF9DQVJEX1BSRVNFTlRBVElPTl9BVVRPEAESJgoiUkVTVUxU'
    'X0NBUkRfUFJFU0VOVEFUSU9OX1BSRUZFUlJFRBAC');

@$core.Deprecated('Use notifyTargetTypeDescriptor instead')
const NotifyTargetType$json = {
  '1': 'NotifyTargetType',
  '2': [
    {'1': 'NOTIFY_TARGET_TYPE_UNSPECIFIED', '2': 0},
    {'1': 'NOTIFY_TARGET_TYPE_TELEGRAM', '2': 1},
    {'1': 'NOTIFY_TARGET_TYPE_LARK_WEBHOOK', '2': 2},
    {'1': 'NOTIFY_TARGET_TYPE_DISCORD_WEBHOOK', '2': 3},
  ],
};

/// Descriptor for `NotifyTargetType`. Decode as a `google.protobuf.EnumDescriptorProto`.
final $typed_data.Uint8List notifyTargetTypeDescriptor = $convert.base64Decode(
    'ChBOb3RpZnlUYXJnZXRUeXBlEiIKHk5PVElGWV9UQVJHRVRfVFlQRV9VTlNQRUNJRklFRBAAEh'
    '8KG05PVElGWV9UQVJHRVRfVFlQRV9URUxFR1JBTRABEiMKH05PVElGWV9UQVJHRVRfVFlQRV9M'
    'QVJLX1dFQkhPT0sQAhImCiJOT1RJRllfVEFSR0VUX1RZUEVfRElTQ09SRF9XRUJIT09LEAM=');

@$core.Deprecated('Use agentDescriptor instead')
const Agent$json = {
  '1': 'Agent',
  '2': [
    {'1': 'name', '3': 1, '4': 1, '5': 9, '8': {}, '10': 'name'},
    {'1': 'description', '3': 2, '4': 1, '5': 9, '10': 'description'},
    {
      '1': 'sub_agents',
      '3': 3,
      '4': 3,
      '5': 11,
      '6': '.agents.v1.Agent',
      '8': {'3': true},
      '10': 'subAgents',
    },
    {
      '1': 'labels',
      '3': 4,
      '4': 3,
      '5': 11,
      '6': '.agents.v1.Agent.LabelsEntry',
      '10': 'labels'
    },
    {
      '1': 'metadata',
      '3': 5,
      '4': 3,
      '5': 11,
      '6': '.agents.v1.Agent.MetadataEntry',
      '10': 'metadata'
    },
    {
      '1': 'config',
      '3': 6,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.AgentConfig',
      '10': 'config'
    },
    {
      '1': 'type',
      '3': 20,
      '4': 1,
      '5': 14,
      '6': '.agents.v1.AgentType',
      '10': 'type'
    },
    {'1': 'enable_a2a', '3': 21, '4': 1, '5': 8, '10': 'enableA2a'},
    {
      '1': 'enable_openai_api',
      '3': 22,
      '4': 1,
      '5': 8,
      '10': 'enableOpenaiApi'
    },
    {'1': 'enable_agui', '3': 32, '4': 1, '5': 8, '10': 'enableAgui'},
    {'1': 'agent_id', '3': 23, '4': 1, '5': 9, '10': 'agentId'},
    {'1': 'display_name', '3': 24, '4': 1, '5': 9, '10': 'displayName'},
    {'1': 'child_agent_ids', '3': 25, '4': 3, '5': 9, '10': 'childAgentIds'},
    {
      '1': 'lifecycle_status',
      '3': 26,
      '4': 1,
      '5': 14,
      '6': '.agents.v1.AgentLifecycleStatus',
      '10': 'lifecycleStatus'
    },
    {'1': 'legacy_name', '3': 27, '4': 1, '5': 9, '10': 'legacyName'},
    {'1': 'version', '3': 28, '4': 1, '5': 3, '10': 'version'},
    {
      '1': 'created_at',
      '3': 29,
      '4': 1,
      '5': 11,
      '6': '.google.protobuf.Timestamp',
      '10': 'createdAt'
    },
    {
      '1': 'updated_at',
      '3': 30,
      '4': 1,
      '5': 11,
      '6': '.google.protobuf.Timestamp',
      '10': 'updatedAt'
    },
    {
      '1': 'deleted_at',
      '3': 31,
      '4': 1,
      '5': 11,
      '6': '.google.protobuf.Timestamp',
      '10': 'deletedAt'
    },
    {'1': 'workspace_id', '3': 100, '4': 1, '5': 9, '10': 'workspaceId'},
  ],
  '3': [Agent_LabelsEntry$json, Agent_MetadataEntry$json],
};

@$core.Deprecated('Use agentDescriptor instead')
const Agent_LabelsEntry$json = {
  '1': 'LabelsEntry',
  '2': [
    {'1': 'key', '3': 1, '4': 1, '5': 9, '10': 'key'},
    {'1': 'value', '3': 2, '4': 1, '5': 9, '10': 'value'},
  ],
  '7': {'7': true},
};

@$core.Deprecated('Use agentDescriptor instead')
const Agent_MetadataEntry$json = {
  '1': 'MetadataEntry',
  '2': [
    {'1': 'key', '3': 1, '4': 1, '5': 9, '10': 'key'},
    {'1': 'value', '3': 2, '4': 1, '5': 9, '10': 'value'},
  ],
  '7': {'7': true},
};

/// Descriptor for `Agent`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List agentDescriptor = $convert.base64Decode(
    'CgVBZ2VudBIhCgRuYW1lGAEgASgJQg36QgpyCBABWgR1c2VyUgRuYW1lEiAKC2Rlc2NyaXB0aW'
    '9uGAIgASgJUgtkZXNjcmlwdGlvbhIzCgpzdWJfYWdlbnRzGAMgAygLMhAuYWdlbnRzLnYxLkFn'
    'ZW50QgIYAVIJc3ViQWdlbnRzEjQKBmxhYmVscxgEIAMoCzIcLmFnZW50cy52MS5BZ2VudC5MYW'
    'JlbHNFbnRyeVIGbGFiZWxzEjoKCG1ldGFkYXRhGAUgAygLMh4uYWdlbnRzLnYxLkFnZW50Lk1l'
    'dGFkYXRhRW50cnlSCG1ldGFkYXRhEi4KBmNvbmZpZxgGIAEoCzIWLmFnZW50cy52MS5BZ2VudE'
    'NvbmZpZ1IGY29uZmlnEigKBHR5cGUYFCABKA4yFC5hZ2VudHMudjEuQWdlbnRUeXBlUgR0eXBl'
    'Eh0KCmVuYWJsZV9hMmEYFSABKAhSCWVuYWJsZUEyYRIqChFlbmFibGVfb3BlbmFpX2FwaRgWIA'
    'EoCFIPZW5hYmxlT3BlbmFpQXBpEh8KC2VuYWJsZV9hZ3VpGCAgASgIUgplbmFibGVBZ3VpEhkK'
    'CGFnZW50X2lkGBcgASgJUgdhZ2VudElkEiEKDGRpc3BsYXlfbmFtZRgYIAEoCVILZGlzcGxheU'
    '5hbWUSJgoPY2hpbGRfYWdlbnRfaWRzGBkgAygJUg1jaGlsZEFnZW50SWRzEkoKEGxpZmVjeWNs'
    'ZV9zdGF0dXMYGiABKA4yHy5hZ2VudHMudjEuQWdlbnRMaWZlY3ljbGVTdGF0dXNSD2xpZmVjeW'
    'NsZVN0YXR1cxIfCgtsZWdhY3lfbmFtZRgbIAEoCVIKbGVnYWN5TmFtZRIYCgd2ZXJzaW9uGBwg'
    'ASgDUgd2ZXJzaW9uEjkKCmNyZWF0ZWRfYXQYHSABKAsyGi5nb29nbGUucHJvdG9idWYuVGltZX'
    'N0YW1wUgljcmVhdGVkQXQSOQoKdXBkYXRlZF9hdBgeIAEoCzIaLmdvb2dsZS5wcm90b2J1Zi5U'
    'aW1lc3RhbXBSCXVwZGF0ZWRBdBI5CgpkZWxldGVkX2F0GB8gASgLMhouZ29vZ2xlLnByb3RvYn'
    'VmLlRpbWVzdGFtcFIJZGVsZXRlZEF0EiEKDHdvcmtzcGFjZV9pZBhkIAEoCVILd29ya3NwYWNl'
    'SWQaOQoLTGFiZWxzRW50cnkSEAoDa2V5GAEgASgJUgNrZXkSFAoFdmFsdWUYAiABKAlSBXZhbH'
    'VlOgI4ARo7Cg1NZXRhZGF0YUVudHJ5EhAKA2tleRgBIAEoCVIDa2V5EhQKBXZhbHVlGAIgASgJ'
    'UgV2YWx1ZToCOAE=');

@$core.Deprecated('Use agentConfigDescriptor instead')
const AgentConfig$json = {
  '1': 'AgentConfig',
  '2': [
    {
      '1': 'runtime',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.AgentRuntime',
      '10': 'runtime'
    },
    {
      '1': 'mcp_servers',
      '3': 2,
      '4': 3,
      '5': 11,
      '6': '.agents.v1.MCPServer',
      '10': 'mcpServers'
    },
    {
      '1': 'context_guard',
      '3': 3,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.ContextGuardConfig',
      '10': 'contextGuard'
    },
    {'1': 'mcp_server_ids', '3': 4, '4': 3, '5': 9, '10': 'mcpServerIds'},
    {'1': 'remote_agent_ids', '3': 5, '4': 3, '5': 9, '10': 'remoteAgentIds'},
    {
      '1': 'file_mounts',
      '3': 6,
      '4': 3,
      '5': 11,
      '6': '.agents.v1.AgentFileMount',
      '10': 'fileMounts'
    },
    {'1': 'skills', '3': 7, '4': 3, '5': 9, '10': 'skills'},
    {
      '1': 'memory',
      '3': 8,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.MemoryConfig',
      '10': 'memory'
    },
    {
      '1': 'result_cards',
      '3': 9,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.ResultCardConfig',
      '10': 'resultCards'
    },
    {'1': 'model', '3': 20, '4': 1, '5': 9, '10': 'model'},
    {'1': 'instruction', '3': 21, '4': 1, '5': 9, '10': 'instruction'},
    {
      '1': 'global_instruction',
      '3': 22,
      '4': 1,
      '5': 9,
      '10': 'globalInstruction'
    },
    {
      '1': 'disallow_transfer_to_parent',
      '3': 23,
      '4': 1,
      '5': 8,
      '10': 'disallowTransferToParent'
    },
    {
      '1': 'disallow_transfer_to_peers',
      '3': 24,
      '4': 1,
      '5': 8,
      '10': 'disallowTransferToPeers'
    },
    {
      '1': 'include_contents',
      '3': 25,
      '4': 1,
      '5': 14,
      '6': '.agents.v1.LLMIncludeContents',
      '10': 'includeContents'
    },
    {'1': 'output_key', '3': 26, '4': 1, '5': 9, '10': 'outputKey'},
    {
      '1': 'input_schema_json',
      '3': 27,
      '4': 1,
      '5': 9,
      '10': 'inputSchemaJson'
    },
    {
      '1': 'output_schema_json',
      '3': 28,
      '4': 1,
      '5': 9,
      '10': 'outputSchemaJson'
    },
    {'1': 'max_iterations', '3': 30, '4': 1, '5': 13, '10': 'maxIterations'},
    {
      '1': 'workflow',
      '3': 40,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.WorkflowConfig',
      '10': 'workflow'
    },
    {
      '1': 'pi',
      '3': 50,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.PiAgentConfig',
      '10': 'pi'
    },
    {
      '1': 'cursor',
      '3': 51,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.CursorAgentConfig',
      '10': 'cursor'
    },
  ],
};

/// Descriptor for `AgentConfig`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List agentConfigDescriptor = $convert.base64Decode(
    'CgtBZ2VudENvbmZpZxIxCgdydW50aW1lGAEgASgLMhcuYWdlbnRzLnYxLkFnZW50UnVudGltZV'
    'IHcnVudGltZRI1CgttY3Bfc2VydmVycxgCIAMoCzIULmFnZW50cy52MS5NQ1BTZXJ2ZXJSCm1j'
    'cFNlcnZlcnMSQgoNY29udGV4dF9ndWFyZBgDIAEoCzIdLmFnZW50cy52MS5Db250ZXh0R3Vhcm'
    'RDb25maWdSDGNvbnRleHRHdWFyZBIkCg5tY3Bfc2VydmVyX2lkcxgEIAMoCVIMbWNwU2VydmVy'
    'SWRzEigKEHJlbW90ZV9hZ2VudF9pZHMYBSADKAlSDnJlbW90ZUFnZW50SWRzEjoKC2ZpbGVfbW'
    '91bnRzGAYgAygLMhkuYWdlbnRzLnYxLkFnZW50RmlsZU1vdW50UgpmaWxlTW91bnRzEhYKBnNr'
    'aWxscxgHIAMoCVIGc2tpbGxzEi8KBm1lbW9yeRgIIAEoCzIXLmFnZW50cy52MS5NZW1vcnlDb2'
    '5maWdSBm1lbW9yeRI+CgxyZXN1bHRfY2FyZHMYCSABKAsyGy5hZ2VudHMudjEuUmVzdWx0Q2Fy'
    'ZENvbmZpZ1ILcmVzdWx0Q2FyZHMSFAoFbW9kZWwYFCABKAlSBW1vZGVsEiAKC2luc3RydWN0aW'
    '9uGBUgASgJUgtpbnN0cnVjdGlvbhItChJnbG9iYWxfaW5zdHJ1Y3Rpb24YFiABKAlSEWdsb2Jh'
    'bEluc3RydWN0aW9uEj0KG2Rpc2FsbG93X3RyYW5zZmVyX3RvX3BhcmVudBgXIAEoCFIYZGlzYW'
    'xsb3dUcmFuc2ZlclRvUGFyZW50EjsKGmRpc2FsbG93X3RyYW5zZmVyX3RvX3BlZXJzGBggASgI'
    'UhdkaXNhbGxvd1RyYW5zZmVyVG9QZWVycxJIChBpbmNsdWRlX2NvbnRlbnRzGBkgASgOMh0uYW'
    'dlbnRzLnYxLkxMTUluY2x1ZGVDb250ZW50c1IPaW5jbHVkZUNvbnRlbnRzEh0KCm91dHB1dF9r'
    'ZXkYGiABKAlSCW91dHB1dEtleRIqChFpbnB1dF9zY2hlbWFfanNvbhgbIAEoCVIPaW5wdXRTY2'
    'hlbWFKc29uEiwKEm91dHB1dF9zY2hlbWFfanNvbhgcIAEoCVIQb3V0cHV0U2NoZW1hSnNvbhIl'
    'Cg5tYXhfaXRlcmF0aW9ucxgeIAEoDVINbWF4SXRlcmF0aW9ucxI1Cgh3b3JrZmxvdxgoIAEoCz'
    'IZLmFnZW50cy52MS5Xb3JrZmxvd0NvbmZpZ1IId29ya2Zsb3cSKAoCcGkYMiABKAsyGC5hZ2Vu'
    'dHMudjEuUGlBZ2VudENvbmZpZ1ICcGkSNAoGY3Vyc29yGDMgASgLMhwuYWdlbnRzLnYxLkN1cn'
    'NvckFnZW50Q29uZmlnUgZjdXJzb3I=');

@$core.Deprecated('Use cursorAgentConfigDescriptor instead')
const CursorAgentConfig$json = {
  '1': 'CursorAgentConfig',
  '2': [
    {'1': 'butterbox_id', '3': 1, '4': 1, '5': 9, '10': 'butterboxId'},
    {'1': 'working_dir', '3': 2, '4': 1, '5': 9, '10': 'workingDir'},
    {'1': 'model', '3': 3, '4': 1, '5': 9, '10': 'model'},
    {'1': 'mode', '3': 4, '4': 1, '5': 9, '10': 'mode'},
    {
      '1': 'max_run_seconds',
      '3': 5,
      '4': 1,
      '5': 5,
      '9': 0,
      '10': 'maxRunSeconds',
      '17': true
    },
  ],
  '8': [
    {'1': '_max_run_seconds'},
  ],
};

/// Descriptor for `CursorAgentConfig`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List cursorAgentConfigDescriptor = $convert.base64Decode(
    'ChFDdXJzb3JBZ2VudENvbmZpZxIhCgxidXR0ZXJib3hfaWQYASABKAlSC2J1dHRlcmJveElkEh'
    '8KC3dvcmtpbmdfZGlyGAIgASgJUgp3b3JraW5nRGlyEhQKBW1vZGVsGAMgASgJUgVtb2RlbBIS'
    'CgRtb2RlGAQgASgJUgRtb2RlEisKD21heF9ydW5fc2Vjb25kcxgFIAEoBUgAUg1tYXhSdW5TZW'
    'NvbmRziAEBQhIKEF9tYXhfcnVuX3NlY29uZHM=');

@$core.Deprecated('Use piAgentConfigDescriptor instead')
const PiAgentConfig$json = {
  '1': 'PiAgentConfig',
  '2': [
    {'1': 'butterbox_id', '3': 1, '4': 1, '5': 9, '10': 'butterboxId'},
    {'1': 'working_dir', '3': 2, '4': 1, '5': 9, '10': 'workingDir'},
    {'1': 'provider', '3': 3, '4': 1, '5': 9, '10': 'provider'},
    {'1': 'model', '3': 4, '4': 1, '5': 9, '10': 'model'},
    {'1': 'thinking_level', '3': 5, '4': 1, '5': 9, '10': 'thinkingLevel'},
    {
      '1': 'max_run_seconds',
      '3': 6,
      '4': 1,
      '5': 5,
      '9': 0,
      '10': 'maxRunSeconds',
      '17': true
    },
  ],
  '8': [
    {'1': '_max_run_seconds'},
  ],
};

/// Descriptor for `PiAgentConfig`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List piAgentConfigDescriptor = $convert.base64Decode(
    'Cg1QaUFnZW50Q29uZmlnEiEKDGJ1dHRlcmJveF9pZBgBIAEoCVILYnV0dGVyYm94SWQSHwoLd2'
    '9ya2luZ19kaXIYAiABKAlSCndvcmtpbmdEaXISGgoIcHJvdmlkZXIYAyABKAlSCHByb3ZpZGVy'
    'EhQKBW1vZGVsGAQgASgJUgVtb2RlbBIlCg50aGlua2luZ19sZXZlbBgFIAEoCVINdGhpbmtpbm'
    'dMZXZlbBIrCg9tYXhfcnVuX3NlY29uZHMYBiABKAVIAFINbWF4UnVuU2Vjb25kc4gBAUISChBf'
    'bWF4X3J1bl9zZWNvbmRz');

@$core.Deprecated('Use workflowConfigDescriptor instead')
const WorkflowConfig$json = {
  '1': 'WorkflowConfig',
  '2': [
    {
      '1': 'nodes',
      '3': 1,
      '4': 3,
      '5': 11,
      '6': '.agents.v1.WorkflowNode',
      '10': 'nodes'
    },
    {
      '1': 'edges',
      '3': 2,
      '4': 3,
      '5': 11,
      '6': '.agents.v1.WorkflowEdge',
      '10': 'edges'
    },
  ],
};

/// Descriptor for `WorkflowConfig`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List workflowConfigDescriptor = $convert.base64Decode(
    'Cg5Xb3JrZmxvd0NvbmZpZxItCgVub2RlcxgBIAMoCzIXLmFnZW50cy52MS5Xb3JrZmxvd05vZG'
    'VSBW5vZGVzEi0KBWVkZ2VzGAIgAygLMhcuYWdlbnRzLnYxLldvcmtmbG93RWRnZVIFZWRnZXM=');

@$core.Deprecated('Use workflowNodeDescriptor instead')
const WorkflowNode$json = {
  '1': 'WorkflowNode',
  '2': [
    {'1': 'name', '3': 1, '4': 1, '5': 9, '10': 'name'},
    {
      '1': 'kind',
      '3': 2,
      '4': 1,
      '5': 14,
      '6': '.agents.v1.WorkflowNodeKind',
      '10': 'kind'
    },
    {
      '1': 'agent',
      '3': 3,
      '4': 1,
      '5': 9,
      '8': {'3': true},
      '10': 'agent',
    },
    {'1': 'question', '3': 4, '4': 1, '5': 9, '10': 'question'},
    {'1': 'agent_id', '3': 5, '4': 1, '5': 9, '10': 'agentId'},
    {'1': 'parallel_worker', '3': 10, '4': 1, '5': 8, '10': 'parallelWorker'},
    {
      '1': 'retry',
      '3': 11,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.WorkflowRetryConfig',
      '10': 'retry'
    },
    {'1': 'timeout_seconds', '3': 12, '4': 1, '5': 5, '10': 'timeoutSeconds'},
    {
      '1': 'form',
      '3': 13,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.HumanInputForm',
      '10': 'form'
    },
  ],
};

/// Descriptor for `WorkflowNode`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List workflowNodeDescriptor = $convert.base64Decode(
    'CgxXb3JrZmxvd05vZGUSEgoEbmFtZRgBIAEoCVIEbmFtZRIvCgRraW5kGAIgASgOMhsuYWdlbn'
    'RzLnYxLldvcmtmbG93Tm9kZUtpbmRSBGtpbmQSGAoFYWdlbnQYAyABKAlCAhgBUgVhZ2VudBIa'
    'CghxdWVzdGlvbhgEIAEoCVIIcXVlc3Rpb24SGQoIYWdlbnRfaWQYBSABKAlSB2FnZW50SWQSJw'
    'oPcGFyYWxsZWxfd29ya2VyGAogASgIUg5wYXJhbGxlbFdvcmtlchI0CgVyZXRyeRgLIAEoCzIe'
    'LmFnZW50cy52MS5Xb3JrZmxvd1JldHJ5Q29uZmlnUgVyZXRyeRInCg90aW1lb3V0X3NlY29uZH'
    'MYDCABKAVSDnRpbWVvdXRTZWNvbmRzEi0KBGZvcm0YDSABKAsyGS5hZ2VudHMudjEuSHVtYW5J'
    'bnB1dEZvcm1SBGZvcm0=');

@$core.Deprecated('Use humanInputFormDescriptor instead')
const HumanInputForm$json = {
  '1': 'HumanInputForm',
  '2': [
    {'1': 'title', '3': 1, '4': 1, '5': 9, '10': 'title'},
    {
      '1': 'fields',
      '3': 2,
      '4': 3,
      '5': 11,
      '6': '.agents.v1.HumanInputFormField',
      '10': 'fields'
    },
  ],
};

/// Descriptor for `HumanInputForm`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List humanInputFormDescriptor = $convert.base64Decode(
    'Cg5IdW1hbklucHV0Rm9ybRIUCgV0aXRsZRgBIAEoCVIFdGl0bGUSNgoGZmllbGRzGAIgAygLMh'
    '4uYWdlbnRzLnYxLkh1bWFuSW5wdXRGb3JtRmllbGRSBmZpZWxkcw==');

@$core.Deprecated('Use humanInputFormFieldDescriptor instead')
const HumanInputFormField$json = {
  '1': 'HumanInputFormField',
  '2': [
    {'1': 'name', '3': 1, '4': 1, '5': 9, '10': 'name'},
    {'1': 'label', '3': 2, '4': 1, '5': 9, '10': 'label'},
    {'1': 'hint', '3': 3, '4': 1, '5': 9, '10': 'hint'},
    {'1': 'required', '3': 4, '4': 1, '5': 8, '10': 'required'},
    {
      '1': 'type',
      '3': 5,
      '4': 1,
      '5': 14,
      '6': '.agents.v1.HumanInputFormFieldType',
      '10': 'type'
    },
    {'1': 'max_length', '3': 6, '4': 1, '5': 5, '10': 'maxLength'},
    {
      '1': 'options',
      '3': 7,
      '4': 3,
      '5': 11,
      '6': '.agents.v1.HumanInputFormOption',
      '10': 'options'
    },
  ],
};

/// Descriptor for `HumanInputFormField`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List humanInputFormFieldDescriptor = $convert.base64Decode(
    'ChNIdW1hbklucHV0Rm9ybUZpZWxkEhIKBG5hbWUYASABKAlSBG5hbWUSFAoFbGFiZWwYAiABKA'
    'lSBWxhYmVsEhIKBGhpbnQYAyABKAlSBGhpbnQSGgoIcmVxdWlyZWQYBCABKAhSCHJlcXVpcmVk'
    'EjYKBHR5cGUYBSABKA4yIi5hZ2VudHMudjEuSHVtYW5JbnB1dEZvcm1GaWVsZFR5cGVSBHR5cG'
    'USHQoKbWF4X2xlbmd0aBgGIAEoBVIJbWF4TGVuZ3RoEjkKB29wdGlvbnMYByADKAsyHy5hZ2Vu'
    'dHMudjEuSHVtYW5JbnB1dEZvcm1PcHRpb25SB29wdGlvbnM=');

@$core.Deprecated('Use humanInputFormOptionDescriptor instead')
const HumanInputFormOption$json = {
  '1': 'HumanInputFormOption',
  '2': [
    {'1': 'value', '3': 1, '4': 1, '5': 9, '10': 'value'},
    {'1': 'label', '3': 2, '4': 1, '5': 9, '10': 'label'},
  ],
};

/// Descriptor for `HumanInputFormOption`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List humanInputFormOptionDescriptor = $convert.base64Decode(
    'ChRIdW1hbklucHV0Rm9ybU9wdGlvbhIUCgV2YWx1ZRgBIAEoCVIFdmFsdWUSFAoFbGFiZWwYAi'
    'ABKAlSBWxhYmVs');

@$core.Deprecated('Use workflowRetryConfigDescriptor instead')
const WorkflowRetryConfig$json = {
  '1': 'WorkflowRetryConfig',
  '2': [
    {'1': 'max_attempts', '3': 1, '4': 1, '5': 5, '10': 'maxAttempts'},
    {
      '1': 'initial_delay_seconds',
      '3': 2,
      '4': 1,
      '5': 5,
      '10': 'initialDelaySeconds'
    },
    {'1': 'max_delay_seconds', '3': 3, '4': 1, '5': 5, '10': 'maxDelaySeconds'},
    {'1': 'backoff_factor', '3': 4, '4': 1, '5': 1, '10': 'backoffFactor'},
  ],
};

/// Descriptor for `WorkflowRetryConfig`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List workflowRetryConfigDescriptor = $convert.base64Decode(
    'ChNXb3JrZmxvd1JldHJ5Q29uZmlnEiEKDG1heF9hdHRlbXB0cxgBIAEoBVILbWF4QXR0ZW1wdH'
    'MSMgoVaW5pdGlhbF9kZWxheV9zZWNvbmRzGAIgASgFUhNpbml0aWFsRGVsYXlTZWNvbmRzEioK'
    'EW1heF9kZWxheV9zZWNvbmRzGAMgASgFUg9tYXhEZWxheVNlY29uZHMSJQoOYmFja29mZl9mYW'
    'N0b3IYBCABKAFSDWJhY2tvZmZGYWN0b3I=');

@$core.Deprecated('Use workflowEdgeDescriptor instead')
const WorkflowEdge$json = {
  '1': 'WorkflowEdge',
  '2': [
    {'1': 'from', '3': 1, '4': 1, '5': 9, '10': 'from'},
    {'1': 'to', '3': 2, '4': 1, '5': 9, '10': 'to'},
    {'1': 'route', '3': 3, '4': 1, '5': 9, '10': 'route'},
    {'1': 'is_default', '3': 4, '4': 1, '5': 8, '10': 'isDefault'},
  ],
};

/// Descriptor for `WorkflowEdge`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List workflowEdgeDescriptor = $convert.base64Decode(
    'CgxXb3JrZmxvd0VkZ2USEgoEZnJvbRgBIAEoCVIEZnJvbRIOCgJ0bxgCIAEoCVICdG8SFAoFcm'
    '91dGUYAyABKAlSBXJvdXRlEh0KCmlzX2RlZmF1bHQYBCABKAhSCWlzRGVmYXVsdA==');

@$core.Deprecated('Use agentRuntimeDescriptor instead')
const AgentRuntime$json = {
  '1': 'AgentRuntime',
  '2': [
    {
      '1': 'streaming_mode',
      '3': 1,
      '4': 1,
      '5': 14,
      '6': '.agents.v1.StreamingMode',
      '10': 'streamingMode'
    },
    {
      '1': 'save_input_blobs_as_artifacts',
      '3': 2,
      '4': 1,
      '5': 8,
      '10': 'saveInputBlobsAsArtifacts'
    },
  ],
};

/// Descriptor for `AgentRuntime`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List agentRuntimeDescriptor = $convert.base64Decode(
    'CgxBZ2VudFJ1bnRpbWUSPwoOc3RyZWFtaW5nX21vZGUYASABKA4yGC5hZ2VudHMudjEuU3RyZW'
    'FtaW5nTW9kZVINc3RyZWFtaW5nTW9kZRJACh1zYXZlX2lucHV0X2Jsb2JzX2FzX2FydGlmYWN0'
    'cxgCIAEoCFIZc2F2ZUlucHV0QmxvYnNBc0FydGlmYWN0cw==');

@$core.Deprecated('Use mCPServerDescriptor instead')
const MCPServer$json = {
  '1': 'MCPServer',
  '2': [
    {'1': 'name', '3': 1, '4': 1, '5': 9, '8': {}, '10': 'name'},
    {
      '1': 'transport',
      '3': 2,
      '4': 1,
      '5': 14,
      '6': '.agents.v1.MCPServerTransport',
      '10': 'transport'
    },
    {'1': 'id', '3': 10, '4': 1, '5': 9, '10': 'id'},
    {'1': 'url', '3': 6, '4': 1, '5': 9, '10': 'url'},
    {
      '1': 'headers',
      '3': 7,
      '4': 3,
      '5': 11,
      '6': '.agents.v1.MCPServer.HeadersEntry',
      '10': 'headers'
    },
    {'1': 'tool_filter', '3': 8, '4': 3, '5': 9, '10': 'toolFilter'},
    {
      '1': 'metadata',
      '3': 9,
      '4': 3,
      '5': 11,
      '6': '.agents.v1.MCPServer.MetadataEntry',
      '10': 'metadata'
    },
    {'1': 'timeout_seconds', '3': 11, '4': 1, '5': 5, '10': 'timeoutSeconds'},
    {
      '1': 'auth',
      '3': 12,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.MCPServerAuth',
      '10': 'auth'
    },
    {'1': 'workspace_id', '3': 100, '4': 1, '5': 9, '10': 'workspaceId'},
  ],
  '3': [MCPServer_HeadersEntry$json, MCPServer_MetadataEntry$json],
  '9': [
    {'1': 3, '2': 6},
  ],
  '10': ['command', 'args', 'env'],
};

@$core.Deprecated('Use mCPServerDescriptor instead')
const MCPServer_HeadersEntry$json = {
  '1': 'HeadersEntry',
  '2': [
    {'1': 'key', '3': 1, '4': 1, '5': 9, '10': 'key'},
    {'1': 'value', '3': 2, '4': 1, '5': 9, '10': 'value'},
  ],
  '7': {'7': true},
};

@$core.Deprecated('Use mCPServerDescriptor instead')
const MCPServer_MetadataEntry$json = {
  '1': 'MetadataEntry',
  '2': [
    {'1': 'key', '3': 1, '4': 1, '5': 9, '10': 'key'},
    {'1': 'value', '3': 2, '4': 1, '5': 9, '10': 'value'},
  ],
  '7': {'7': true},
};

/// Descriptor for `MCPServer`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List mCPServerDescriptor = $convert.base64Decode(
    'CglNQ1BTZXJ2ZXISGwoEbmFtZRgBIAEoCUIH+kIEcgIQAVIEbmFtZRI7Cgl0cmFuc3BvcnQYAi'
    'ABKA4yHS5hZ2VudHMudjEuTUNQU2VydmVyVHJhbnNwb3J0Ugl0cmFuc3BvcnQSDgoCaWQYCiAB'
    'KAlSAmlkEhAKA3VybBgGIAEoCVIDdXJsEjsKB2hlYWRlcnMYByADKAsyIS5hZ2VudHMudjEuTU'
    'NQU2VydmVyLkhlYWRlcnNFbnRyeVIHaGVhZGVycxIfCgt0b29sX2ZpbHRlchgIIAMoCVIKdG9v'
    'bEZpbHRlchI+CghtZXRhZGF0YRgJIAMoCzIiLmFnZW50cy52MS5NQ1BTZXJ2ZXIuTWV0YWRhdG'
    'FFbnRyeVIIbWV0YWRhdGESJwoPdGltZW91dF9zZWNvbmRzGAsgASgFUg50aW1lb3V0U2Vjb25k'
    'cxIsCgRhdXRoGAwgASgLMhguYWdlbnRzLnYxLk1DUFNlcnZlckF1dGhSBGF1dGgSIQoMd29ya3'
    'NwYWNlX2lkGGQgASgJUgt3b3Jrc3BhY2VJZBo6CgxIZWFkZXJzRW50cnkSEAoDa2V5GAEgASgJ'
    'UgNrZXkSFAoFdmFsdWUYAiABKAlSBXZhbHVlOgI4ARo7Cg1NZXRhZGF0YUVudHJ5EhAKA2tleR'
    'gBIAEoCVIDa2V5EhQKBXZhbHVlGAIgASgJUgV2YWx1ZToCOAFKBAgDEAZSB2NvbW1hbmRSBGFy'
    'Z3NSA2Vudg==');

@$core.Deprecated('Use mCPServerAuthDescriptor instead')
const MCPServerAuth$json = {
  '1': 'MCPServerAuth',
  '2': [
    {
      '1': 'type',
      '3': 1,
      '4': 1,
      '5': 14,
      '6': '.agents.v1.MCPServerAuthType',
      '10': 'type'
    },
    {
      '1': 'oauth2',
      '3': 2,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.MCPServerOAuth2Config',
      '10': 'oauth2'
    },
  ],
};

/// Descriptor for `MCPServerAuth`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List mCPServerAuthDescriptor = $convert.base64Decode(
    'Cg1NQ1BTZXJ2ZXJBdXRoEjAKBHR5cGUYASABKA4yHC5hZ2VudHMudjEuTUNQU2VydmVyQXV0aF'
    'R5cGVSBHR5cGUSOAoGb2F1dGgyGAIgASgLMiAuYWdlbnRzLnYxLk1DUFNlcnZlck9BdXRoMkNv'
    'bmZpZ1IGb2F1dGgy');

@$core.Deprecated('Use mCPServerOAuth2ConfigDescriptor instead')
const MCPServerOAuth2Config$json = {
  '1': 'MCPServerOAuth2Config',
  '2': [
    {'1': 'client_id', '3': 1, '4': 1, '5': 9, '10': 'clientId'},
    {'1': 'client_secret', '3': 2, '4': 1, '5': 9, '10': 'clientSecret'},
    {'1': 'scopes', '3': 3, '4': 3, '5': 9, '10': 'scopes'},
    {
      '1': 'authorization_url',
      '3': 4,
      '4': 1,
      '5': 9,
      '10': 'authorizationUrl'
    },
    {'1': 'token_url', '3': 5, '4': 1, '5': 9, '10': 'tokenUrl'},
    {
      '1': 'resource_metadata_url',
      '3': 6,
      '4': 1,
      '5': 9,
      '10': 'resourceMetadataUrl'
    },
    {
      '1': 'authorization_server_url',
      '3': 7,
      '4': 1,
      '5': 9,
      '10': 'authorizationServerUrl'
    },
    {'1': 'resource', '3': 8, '4': 1, '5': 9, '10': 'resource'},
  ],
};

/// Descriptor for `MCPServerOAuth2Config`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List mCPServerOAuth2ConfigDescriptor = $convert.base64Decode(
    'ChVNQ1BTZXJ2ZXJPQXV0aDJDb25maWcSGwoJY2xpZW50X2lkGAEgASgJUghjbGllbnRJZBIjCg'
    '1jbGllbnRfc2VjcmV0GAIgASgJUgxjbGllbnRTZWNyZXQSFgoGc2NvcGVzGAMgAygJUgZzY29w'
    'ZXMSKwoRYXV0aG9yaXphdGlvbl91cmwYBCABKAlSEGF1dGhvcml6YXRpb25VcmwSGwoJdG9rZW'
    '5fdXJsGAUgASgJUgh0b2tlblVybBIyChVyZXNvdXJjZV9tZXRhZGF0YV91cmwYBiABKAlSE3Jl'
    'c291cmNlTWV0YWRhdGFVcmwSOAoYYXV0aG9yaXphdGlvbl9zZXJ2ZXJfdXJsGAcgASgJUhZhdX'
    'Rob3JpemF0aW9uU2VydmVyVXJsEhoKCHJlc291cmNlGAggASgJUghyZXNvdXJjZQ==');

@$core.Deprecated('Use contextGuardConfigDescriptor instead')
const ContextGuardConfig$json = {
  '1': 'ContextGuardConfig',
  '2': [
    {
      '1': 'strategy',
      '3': 1,
      '4': 1,
      '5': 14,
      '6': '.agents.v1.ContextGuardStrategy',
      '10': 'strategy'
    },
    {'1': 'max_turns', '3': 2, '4': 1, '5': 5, '10': 'maxTurns'},
    {'1': 'max_tokens', '3': 3, '4': 1, '5': 5, '10': 'maxTokens'},
  ],
};

/// Descriptor for `ContextGuardConfig`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List contextGuardConfigDescriptor = $convert.base64Decode(
    'ChJDb250ZXh0R3VhcmRDb25maWcSOwoIc3RyYXRlZ3kYASABKA4yHy5hZ2VudHMudjEuQ29udG'
    'V4dEd1YXJkU3RyYXRlZ3lSCHN0cmF0ZWd5EhsKCW1heF90dXJucxgCIAEoBVIIbWF4VHVybnMS'
    'HQoKbWF4X3Rva2VucxgDIAEoBVIJbWF4VG9rZW5z');

@$core.Deprecated('Use memoryConfigDescriptor instead')
const MemoryConfig$json = {
  '1': 'MemoryConfig',
  '2': [
    {'1': 'enabled', '3': 1, '4': 1, '5': 8, '10': 'enabled'},
    {
      '1': 'disable_auto_recall',
      '3': 2,
      '4': 1,
      '5': 8,
      '10': 'disableAutoRecall'
    },
    {
      '1': 'disable_auto_capture',
      '3': 3,
      '4': 1,
      '5': 8,
      '10': 'disableAutoCapture'
    },
    {'1': 'enable_tools', '3': 4, '4': 1, '5': 8, '10': 'enableTools'},
    {
      '1': 'allow_agent_scope_write',
      '3': 5,
      '4': 1,
      '5': 8,
      '10': 'allowAgentScopeWrite'
    },
    {'1': 'top_k', '3': 6, '4': 1, '5': 5, '9': 0, '10': 'topK', '17': true},
    {
      '1': 'threshold',
      '3': 7,
      '4': 1,
      '5': 2,
      '9': 1,
      '10': 'threshold',
      '17': true
    },
  ],
  '8': [
    {'1': '_top_k'},
    {'1': '_threshold'},
  ],
};

/// Descriptor for `MemoryConfig`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List memoryConfigDescriptor = $convert.base64Decode(
    'CgxNZW1vcnlDb25maWcSGAoHZW5hYmxlZBgBIAEoCFIHZW5hYmxlZBIuChNkaXNhYmxlX2F1dG'
    '9fcmVjYWxsGAIgASgIUhFkaXNhYmxlQXV0b1JlY2FsbBIwChRkaXNhYmxlX2F1dG9fY2FwdHVy'
    'ZRgDIAEoCFISZGlzYWJsZUF1dG9DYXB0dXJlEiEKDGVuYWJsZV90b29scxgEIAEoCFILZW5hYm'
    'xlVG9vbHMSNQoXYWxsb3dfYWdlbnRfc2NvcGVfd3JpdGUYBSABKAhSFGFsbG93QWdlbnRTY29w'
    'ZVdyaXRlEhgKBXRvcF9rGAYgASgFSABSBHRvcEuIAQESIQoJdGhyZXNob2xkGAcgASgCSAFSCX'
    'RocmVzaG9sZIgBAUIICgZfdG9wX2tCDAoKX3RocmVzaG9sZA==');

@$core.Deprecated('Use resultCardConfigDescriptor instead')
const ResultCardConfig$json = {
  '1': 'ResultCardConfig',
  '2': [
    {
      '1': 'generation',
      '3': 1,
      '4': 1,
      '5': 14,
      '6': '.agents.v1.ResultCardGeneration',
      '10': 'generation'
    },
    {
      '1': 'presentation',
      '3': 2,
      '4': 1,
      '5': 14,
      '6': '.agents.v1.ResultCardPresentation',
      '10': 'presentation'
    },
  ],
};

/// Descriptor for `ResultCardConfig`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List resultCardConfigDescriptor = $convert.base64Decode(
    'ChBSZXN1bHRDYXJkQ29uZmlnEj8KCmdlbmVyYXRpb24YASABKA4yHy5hZ2VudHMudjEuUmVzdW'
    'x0Q2FyZEdlbmVyYXRpb25SCmdlbmVyYXRpb24SRQoMcHJlc2VudGF0aW9uGAIgASgOMiEuYWdl'
    'bnRzLnYxLlJlc3VsdENhcmRQcmVzZW50YXRpb25SDHByZXNlbnRhdGlvbg==');

@$core.Deprecated('Use remoteAgentDescriptor instead')
const RemoteAgent$json = {
  '1': 'RemoteAgent',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '8': {}, '10': 'id'},
    {'1': 'name', '3': 2, '4': 1, '5': 9, '8': {}, '10': 'name'},
    {'1': 'url', '3': 3, '4': 1, '5': 9, '10': 'url'},
    {
      '1': 'protocol',
      '3': 4,
      '4': 1,
      '5': 14,
      '6': '.agents.v1.RemoteAgentProtocol',
      '10': 'protocol'
    },
    {'1': 'daemon_runtime_id', '3': 6, '4': 1, '5': 9, '10': 'daemonRuntimeId'},
    {'1': 'acp_runtime', '3': 7, '4': 1, '5': 9, '10': 'acpRuntime'},
    {'1': 'opencode_agent', '3': 8, '4': 1, '5': 9, '10': 'opencodeAgent'},
    {'1': 'opencode_model', '3': 9, '4': 1, '5': 9, '10': 'opencodeModel'},
    {'1': 'username', '3': 10, '4': 1, '5': 9, '10': 'username'},
    {'1': 'password', '3': 11, '4': 1, '5': 9, '10': 'password'},
    {'1': 'workspace_id', '3': 100, '4': 1, '5': 9, '10': 'workspaceId'},
  ],
  '9': [
    {'1': 5, '2': 6},
  ],
  '10': ['daemon_capability'],
};

/// Descriptor for `RemoteAgent`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List remoteAgentDescriptor = $convert.base64Decode(
    'CgtSZW1vdGVBZ2VudBIXCgJpZBgBIAEoCUIH+kIEcgIQAVICaWQSGwoEbmFtZRgCIAEoCUIH+k'
    'IEcgIQAVIEbmFtZRIQCgN1cmwYAyABKAlSA3VybBI6Cghwcm90b2NvbBgEIAEoDjIeLmFnZW50'
    'cy52MS5SZW1vdGVBZ2VudFByb3RvY29sUghwcm90b2NvbBIqChFkYWVtb25fcnVudGltZV9pZB'
    'gGIAEoCVIPZGFlbW9uUnVudGltZUlkEh8KC2FjcF9ydW50aW1lGAcgASgJUgphY3BSdW50aW1l'
    'EiUKDm9wZW5jb2RlX2FnZW50GAggASgJUg1vcGVuY29kZUFnZW50EiUKDm9wZW5jb2RlX21vZG'
    'VsGAkgASgJUg1vcGVuY29kZU1vZGVsEhoKCHVzZXJuYW1lGAogASgJUgh1c2VybmFtZRIaCghw'
    'YXNzd29yZBgLIAEoCVIIcGFzc3dvcmQSIQoMd29ya3NwYWNlX2lkGGQgASgJUgt3b3Jrc3BhY2'
    'VJZEoECAUQBlIRZGFlbW9uX2NhcGFiaWxpdHk=');

@$core.Deprecated('Use modelConfigDescriptor instead')
const ModelConfig$json = {
  '1': 'ModelConfig',
  '2': [
    {'1': 'name', '3': 1, '4': 1, '5': 9, '10': 'name'},
    {'1': 'alias', '3': 2, '4': 1, '5': 9, '10': 'alias'},
    {
      '1': 'context_window_tokens',
      '3': 3,
      '4': 1,
      '5': 13,
      '10': 'contextWindowTokens'
    },
  ],
};

/// Descriptor for `ModelConfig`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List modelConfigDescriptor = $convert.base64Decode(
    'CgtNb2RlbENvbmZpZxISCgRuYW1lGAEgASgJUgRuYW1lEhQKBWFsaWFzGAIgASgJUgVhbGlhcx'
    'IyChVjb250ZXh0X3dpbmRvd190b2tlbnMYAyABKA1SE2NvbnRleHRXaW5kb3dUb2tlbnM=');

@$core.Deprecated('Use modelProviderDescriptor instead')
const ModelProvider$json = {
  '1': 'ModelProvider',
  '2': [
    {'1': 'name', '3': 1, '4': 1, '5': 9, '10': 'name'},
    {'1': 'type', '3': 2, '4': 1, '5': 9, '10': 'type'},
    {'1': 'api_key', '3': 3, '4': 1, '5': 9, '10': 'apiKey'},
    {'1': 'base_url', '3': 4, '4': 1, '5': 9, '10': 'baseUrl'},
    {
      '1': 'models',
      '3': 5,
      '4': 3,
      '5': 11,
      '6': '.agents.v1.ModelConfig',
      '10': 'models'
    },
    {'1': 'workspace_id', '3': 100, '4': 1, '5': 9, '10': 'workspaceId'},
  ],
};

/// Descriptor for `ModelProvider`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List modelProviderDescriptor = $convert.base64Decode(
    'Cg1Nb2RlbFByb3ZpZGVyEhIKBG5hbWUYASABKAlSBG5hbWUSEgoEdHlwZRgCIAEoCVIEdHlwZR'
    'IXCgdhcGlfa2V5GAMgASgJUgZhcGlLZXkSGQoIYmFzZV91cmwYBCABKAlSB2Jhc2VVcmwSLgoG'
    'bW9kZWxzGAUgAygLMhYuYWdlbnRzLnYxLk1vZGVsQ29uZmlnUgZtb2RlbHMSIQoMd29ya3NwYW'
    'NlX2lkGGQgASgJUgt3b3Jrc3BhY2VJZA==');

@$core.Deprecated('Use notifyGroupDescriptor instead')
const NotifyGroup$json = {
  '1': 'NotifyGroup',
  '2': [
    {'1': 'name', '3': 1, '4': 1, '5': 9, '10': 'name'},
    {'1': 'enabled', '3': 2, '4': 1, '5': 8, '10': 'enabled'},
    {
      '1': 'targets',
      '3': 3,
      '4': 3,
      '5': 11,
      '6': '.agents.v1.NotifyTarget',
      '10': 'targets'
    },
    {
      '1': 'metadata',
      '3': 10,
      '4': 3,
      '5': 11,
      '6': '.agents.v1.NotifyGroup.MetadataEntry',
      '10': 'metadata'
    },
    {'1': 'workspace_id', '3': 100, '4': 1, '5': 9, '10': 'workspaceId'},
  ],
  '3': [NotifyGroup_MetadataEntry$json],
};

@$core.Deprecated('Use notifyGroupDescriptor instead')
const NotifyGroup_MetadataEntry$json = {
  '1': 'MetadataEntry',
  '2': [
    {'1': 'key', '3': 1, '4': 1, '5': 9, '10': 'key'},
    {'1': 'value', '3': 2, '4': 1, '5': 9, '10': 'value'},
  ],
  '7': {'7': true},
};

/// Descriptor for `NotifyGroup`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List notifyGroupDescriptor = $convert.base64Decode(
    'CgtOb3RpZnlHcm91cBISCgRuYW1lGAEgASgJUgRuYW1lEhgKB2VuYWJsZWQYAiABKAhSB2VuYW'
    'JsZWQSMQoHdGFyZ2V0cxgDIAMoCzIXLmFnZW50cy52MS5Ob3RpZnlUYXJnZXRSB3RhcmdldHMS'
    'QAoIbWV0YWRhdGEYCiADKAsyJC5hZ2VudHMudjEuTm90aWZ5R3JvdXAuTWV0YWRhdGFFbnRyeV'
    'IIbWV0YWRhdGESIQoMd29ya3NwYWNlX2lkGGQgASgJUgt3b3Jrc3BhY2VJZBo7Cg1NZXRhZGF0'
    'YUVudHJ5EhAKA2tleRgBIAEoCVIDa2V5EhQKBXZhbHVlGAIgASgJUgV2YWx1ZToCOAE=');

@$core.Deprecated('Use notifyTargetDescriptor instead')
const NotifyTarget$json = {
  '1': 'NotifyTarget',
  '2': [
    {'1': 'name', '3': 1, '4': 1, '5': 9, '10': 'name'},
    {'1': 'enabled', '3': 2, '4': 1, '5': 8, '10': 'enabled'},
    {
      '1': 'type',
      '3': 3,
      '4': 1,
      '5': 14,
      '6': '.agents.v1.NotifyTargetType',
      '10': 'type'
    },
    {
      '1': 'telegram',
      '3': 4,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.TelegramNotifyTarget',
      '10': 'telegram'
    },
    {
      '1': 'lark',
      '3': 5,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.LarkNotifyTarget',
      '10': 'lark'
    },
    {
      '1': 'discord',
      '3': 6,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.DiscordNotifyTarget',
      '10': 'discord'
    },
    {
      '1': 'metadata',
      '3': 10,
      '4': 3,
      '5': 11,
      '6': '.agents.v1.NotifyTarget.MetadataEntry',
      '10': 'metadata'
    },
  ],
  '3': [NotifyTarget_MetadataEntry$json],
};

@$core.Deprecated('Use notifyTargetDescriptor instead')
const NotifyTarget_MetadataEntry$json = {
  '1': 'MetadataEntry',
  '2': [
    {'1': 'key', '3': 1, '4': 1, '5': 9, '10': 'key'},
    {'1': 'value', '3': 2, '4': 1, '5': 9, '10': 'value'},
  ],
  '7': {'7': true},
};

/// Descriptor for `NotifyTarget`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List notifyTargetDescriptor = $convert.base64Decode(
    'CgxOb3RpZnlUYXJnZXQSEgoEbmFtZRgBIAEoCVIEbmFtZRIYCgdlbmFibGVkGAIgASgIUgdlbm'
    'FibGVkEi8KBHR5cGUYAyABKA4yGy5hZ2VudHMudjEuTm90aWZ5VGFyZ2V0VHlwZVIEdHlwZRI7'
    'Cgh0ZWxlZ3JhbRgEIAEoCzIfLmFnZW50cy52MS5UZWxlZ3JhbU5vdGlmeVRhcmdldFIIdGVsZW'
    'dyYW0SLwoEbGFyaxgFIAEoCzIbLmFnZW50cy52MS5MYXJrTm90aWZ5VGFyZ2V0UgRsYXJrEjgK'
    'B2Rpc2NvcmQYBiABKAsyHi5hZ2VudHMudjEuRGlzY29yZE5vdGlmeVRhcmdldFIHZGlzY29yZB'
    'JBCghtZXRhZGF0YRgKIAMoCzIlLmFnZW50cy52MS5Ob3RpZnlUYXJnZXQuTWV0YWRhdGFFbnRy'
    'eVIIbWV0YWRhdGEaOwoNTWV0YWRhdGFFbnRyeRIQCgNrZXkYASABKAlSA2tleRIUCgV2YWx1ZR'
    'gCIAEoCVIFdmFsdWU6AjgB');

@$core.Deprecated('Use telegramNotifyTargetDescriptor instead')
const TelegramNotifyTarget$json = {
  '1': 'TelegramNotifyTarget',
  '2': [
    {
      '1': 'bot_token',
      '3': 1,
      '4': 1,
      '5': 9,
      '8': {'3': true},
      '10': 'botToken',
    },
    {
      '1': 'chat_id',
      '3': 2,
      '4': 1,
      '5': 9,
      '8': {'3': true},
      '10': 'chatId',
    },
    {
      '1': 'parse_mode',
      '3': 3,
      '4': 1,
      '5': 9,
      '8': {'3': true},
      '10': 'parseMode',
    },
    {
      '1': 'message_thread_id',
      '3': 4,
      '4': 1,
      '5': 3,
      '8': {'3': true},
      '10': 'messageThreadId',
    },
    {'1': 'destination_id', '3': 5, '4': 1, '5': 9, '10': 'destinationId'},
  ],
};

/// Descriptor for `TelegramNotifyTarget`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List telegramNotifyTargetDescriptor = $convert.base64Decode(
    'ChRUZWxlZ3JhbU5vdGlmeVRhcmdldBIfCglib3RfdG9rZW4YASABKAlCAhgBUghib3RUb2tlbh'
    'IbCgdjaGF0X2lkGAIgASgJQgIYAVIGY2hhdElkEiEKCnBhcnNlX21vZGUYAyABKAlCAhgBUglw'
    'YXJzZU1vZGUSLgoRbWVzc2FnZV90aHJlYWRfaWQYBCABKANCAhgBUg9tZXNzYWdlVGhyZWFkSW'
    'QSJQoOZGVzdGluYXRpb25faWQYBSABKAlSDWRlc3RpbmF0aW9uSWQ=');

@$core.Deprecated('Use larkNotifyTargetDescriptor instead')
const LarkNotifyTarget$json = {
  '1': 'LarkNotifyTarget',
  '2': [
    {'1': 'webhook_url', '3': 1, '4': 1, '5': 9, '10': 'webhookUrl'},
    {'1': 'secret', '3': 2, '4': 1, '5': 9, '10': 'secret'},
  ],
};

/// Descriptor for `LarkNotifyTarget`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List larkNotifyTargetDescriptor = $convert.base64Decode(
    'ChBMYXJrTm90aWZ5VGFyZ2V0Eh8KC3dlYmhvb2tfdXJsGAEgASgJUgp3ZWJob29rVXJsEhYKBn'
    'NlY3JldBgCIAEoCVIGc2VjcmV0');

@$core.Deprecated('Use discordNotifyTargetDescriptor instead')
const DiscordNotifyTarget$json = {
  '1': 'DiscordNotifyTarget',
  '2': [
    {'1': 'webhook_url', '3': 1, '4': 1, '5': 9, '10': 'webhookUrl'},
    {'1': 'username', '3': 2, '4': 1, '5': 9, '10': 'username'},
    {'1': 'avatar_url', '3': 3, '4': 1, '5': 9, '10': 'avatarUrl'},
    {'1': 'thread_id', '3': 4, '4': 1, '5': 9, '10': 'threadId'},
  ],
};

/// Descriptor for `DiscordNotifyTarget`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List discordNotifyTargetDescriptor = $convert.base64Decode(
    'ChNEaXNjb3JkTm90aWZ5VGFyZ2V0Eh8KC3dlYmhvb2tfdXJsGAEgASgJUgp3ZWJob29rVXJsEh'
    'oKCHVzZXJuYW1lGAIgASgJUgh1c2VybmFtZRIdCgphdmF0YXJfdXJsGAMgASgJUglhdmF0YXJV'
    'cmwSGwoJdGhyZWFkX2lkGAQgASgJUgh0aHJlYWRJZA==');
