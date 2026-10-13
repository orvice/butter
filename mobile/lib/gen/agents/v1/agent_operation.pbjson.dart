// This is a generated file - do not edit.
//
// Generated from agents/v1/agent_operation.proto.

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

@$core.Deprecated('Use agentOperationTypeDescriptor instead')
const AgentOperationType$json = {
  '1': 'AgentOperationType',
  '2': [
    {'1': 'AGENT_OPERATION_TYPE_UNSPECIFIED', '2': 0},
    {'1': 'AGENT_OPERATION_TYPE_CREATE', '2': 1},
    {'1': 'AGENT_OPERATION_TYPE_UPDATE_CONFIGURATION', '2': 2},
    {'1': 'AGENT_OPERATION_TYPE_DELETE', '2': 3},
    {'1': 'AGENT_OPERATION_TYPE_RESTORE', '2': 4},
  ],
};

/// Descriptor for `AgentOperationType`. Decode as a `google.protobuf.EnumDescriptorProto`.
final $typed_data.Uint8List agentOperationTypeDescriptor = $convert.base64Decode(
    'ChJBZ2VudE9wZXJhdGlvblR5cGUSJAogQUdFTlRfT1BFUkFUSU9OX1RZUEVfVU5TUEVDSUZJRU'
    'QQABIfChtBR0VOVF9PUEVSQVRJT05fVFlQRV9DUkVBVEUQARItCilBR0VOVF9PUEVSQVRJT05f'
    'VFlQRV9VUERBVEVfQ09ORklHVVJBVElPThACEh8KG0FHRU5UX09QRVJBVElPTl9UWVBFX0RFTE'
    'VURRADEiAKHEFHRU5UX09QRVJBVElPTl9UWVBFX1JFU1RPUkUQBA==');

@$core.Deprecated('Use agentOperationStatusDescriptor instead')
const AgentOperationStatus$json = {
  '1': 'AgentOperationStatus',
  '2': [
    {'1': 'AGENT_OPERATION_STATUS_UNSPECIFIED', '2': 0},
    {'1': 'AGENT_OPERATION_STATUS_PENDING', '2': 1},
    {'1': 'AGENT_OPERATION_STATUS_RUNNING', '2': 2},
    {'1': 'AGENT_OPERATION_STATUS_SUCCEEDED', '2': 3},
    {'1': 'AGENT_OPERATION_STATUS_FAILED', '2': 4},
  ],
};

/// Descriptor for `AgentOperationStatus`. Decode as a `google.protobuf.EnumDescriptorProto`.
final $typed_data.Uint8List agentOperationStatusDescriptor = $convert.base64Decode(
    'ChRBZ2VudE9wZXJhdGlvblN0YXR1cxImCiJBR0VOVF9PUEVSQVRJT05fU1RBVFVTX1VOU1BFQ0'
    'lGSUVEEAASIgoeQUdFTlRfT1BFUkFUSU9OX1NUQVRVU19QRU5ESU5HEAESIgoeQUdFTlRfT1BF'
    'UkFUSU9OX1NUQVRVU19SVU5OSU5HEAISJAogQUdFTlRfT1BFUkFUSU9OX1NUQVRVU19TVUNDRU'
    'VERUQQAxIhCh1BR0VOVF9PUEVSQVRJT05fU1RBVFVTX0ZBSUxFRBAE');

@$core.Deprecated('Use agentOperationStepKindDescriptor instead')
const AgentOperationStepKind$json = {
  '1': 'AgentOperationStepKind',
  '2': [
    {'1': 'AGENT_OPERATION_STEP_KIND_UNSPECIFIED', '2': 0},
    {'1': 'AGENT_OPERATION_STEP_KIND_DB_PROVISION', '2': 1},
    {'1': 'AGENT_OPERATION_STEP_KIND_CONTENT_COMMIT', '2': 2},
    {'1': 'AGENT_OPERATION_STEP_KIND_SYNC_PUBLISH', '2': 3},
    {'1': 'AGENT_OPERATION_STEP_KIND_ACTIVATE', '2': 4},
    {'1': 'AGENT_OPERATION_STEP_KIND_TOMBSTONE', '2': 5},
    {'1': 'AGENT_OPERATION_STEP_KIND_RESTORE_DB', '2': 6},
    {'1': 'AGENT_OPERATION_STEP_KIND_DB_PATCH', '2': 7},
  ],
};

/// Descriptor for `AgentOperationStepKind`. Decode as a `google.protobuf.EnumDescriptorProto`.
final $typed_data.Uint8List agentOperationStepKindDescriptor = $convert.base64Decode(
    'ChZBZ2VudE9wZXJhdGlvblN0ZXBLaW5kEikKJUFHRU5UX09QRVJBVElPTl9TVEVQX0tJTkRfVU'
    '5TUEVDSUZJRUQQABIqCiZBR0VOVF9PUEVSQVRJT05fU1RFUF9LSU5EX0RCX1BST1ZJU0lPThAB'
    'EiwKKEFHRU5UX09QRVJBVElPTl9TVEVQX0tJTkRfQ09OVEVOVF9DT01NSVQQAhIqCiZBR0VOVF'
    '9PUEVSQVRJT05fU1RFUF9LSU5EX1NZTkNfUFVCTElTSBADEiYKIkFHRU5UX09QRVJBVElPTl9T'
    'VEVQX0tJTkRfQUNUSVZBVEUQBBInCiNBR0VOVF9PUEVSQVRJT05fU1RFUF9LSU5EX1RPTUJTVE'
    '9ORRAFEigKJEFHRU5UX09QRVJBVElPTl9TVEVQX0tJTkRfUkVTVE9SRV9EQhAGEiYKIkFHRU5U'
    'X09QRVJBVElPTl9TVEVQX0tJTkRfREJfUEFUQ0gQBw==');

@$core.Deprecated('Use agentOperationStepStatusDescriptor instead')
const AgentOperationStepStatus$json = {
  '1': 'AgentOperationStepStatus',
  '2': [
    {'1': 'AGENT_OPERATION_STEP_STATUS_UNSPECIFIED', '2': 0},
    {'1': 'AGENT_OPERATION_STEP_STATUS_PENDING', '2': 1},
    {'1': 'AGENT_OPERATION_STEP_STATUS_SUCCEEDED', '2': 2},
    {'1': 'AGENT_OPERATION_STEP_STATUS_FAILED', '2': 3},
    {'1': 'AGENT_OPERATION_STEP_STATUS_SKIPPED', '2': 4},
  ],
};

/// Descriptor for `AgentOperationStepStatus`. Decode as a `google.protobuf.EnumDescriptorProto`.
final $typed_data.Uint8List agentOperationStepStatusDescriptor = $convert.base64Decode(
    'ChhBZ2VudE9wZXJhdGlvblN0ZXBTdGF0dXMSKwonQUdFTlRfT1BFUkFUSU9OX1NURVBfU1RBVF'
    'VTX1VOU1BFQ0lGSUVEEAASJwojQUdFTlRfT1BFUkFUSU9OX1NURVBfU1RBVFVTX1BFTkRJTkcQ'
    'ARIpCiVBR0VOVF9PUEVSQVRJT05fU1RFUF9TVEFUVVNfU1VDQ0VFREVEEAISJgoiQUdFTlRfT1'
    'BFUkFUSU9OX1NURVBfU1RBVFVTX0ZBSUxFRBADEicKI0FHRU5UX09QRVJBVElPTl9TVEVQX1NU'
    'QVRVU19TS0lQUEVEEAQ=');

@$core.Deprecated('Use agentOperationDescriptor instead')
const AgentOperation$json = {
  '1': 'AgentOperation',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
    {'1': 'agent_id', '3': 2, '4': 1, '5': 9, '10': 'agentId'},
    {'1': 'agent_name', '3': 3, '4': 1, '5': 9, '10': 'agentName'},
    {
      '1': 'type',
      '3': 4,
      '4': 1,
      '5': 14,
      '6': '.agents.v1.AgentOperationType',
      '10': 'type'
    },
    {
      '1': 'status',
      '3': 5,
      '4': 1,
      '5': 14,
      '6': '.agents.v1.AgentOperationStatus',
      '10': 'status'
    },
    {'1': 'attempt_count', '3': 6, '4': 1, '5': 5, '10': 'attemptCount'},
    {
      '1': 'steps',
      '3': 7,
      '4': 3,
      '5': 11,
      '6': '.agents.v1.AgentOperationStep',
      '10': 'steps'
    },
    {'1': 'error', '3': 8, '4': 1, '5': 9, '10': 'error'},
    {'1': 'request_json', '3': 9, '4': 1, '5': 9, '10': 'requestJson'},
    {'1': 'base_commit_sha', '3': 10, '4': 1, '5': 9, '10': 'baseCommitSha'},
    {
      '1': 'expected_agent_version',
      '3': 11,
      '4': 1,
      '5': 3,
      '10': 'expectedAgentVersion'
    },
    {'1': 'committed_sha', '3': 12, '4': 1, '5': 9, '10': 'committedSha'},
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
};

/// Descriptor for `AgentOperation`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List agentOperationDescriptor = $convert.base64Decode(
    'Cg5BZ2VudE9wZXJhdGlvbhIOCgJpZBgBIAEoCVICaWQSGQoIYWdlbnRfaWQYAiABKAlSB2FnZW'
    '50SWQSHQoKYWdlbnRfbmFtZRgDIAEoCVIJYWdlbnROYW1lEjEKBHR5cGUYBCABKA4yHS5hZ2Vu'
    'dHMudjEuQWdlbnRPcGVyYXRpb25UeXBlUgR0eXBlEjcKBnN0YXR1cxgFIAEoDjIfLmFnZW50cy'
    '52MS5BZ2VudE9wZXJhdGlvblN0YXR1c1IGc3RhdHVzEiMKDWF0dGVtcHRfY291bnQYBiABKAVS'
    'DGF0dGVtcHRDb3VudBIzCgVzdGVwcxgHIAMoCzIdLmFnZW50cy52MS5BZ2VudE9wZXJhdGlvbl'
    'N0ZXBSBXN0ZXBzEhQKBWVycm9yGAggASgJUgVlcnJvchIhCgxyZXF1ZXN0X2pzb24YCSABKAlS'
    'C3JlcXVlc3RKc29uEiYKD2Jhc2VfY29tbWl0X3NoYRgKIAEoCVINYmFzZUNvbW1pdFNoYRI0Ch'
    'ZleHBlY3RlZF9hZ2VudF92ZXJzaW9uGAsgASgDUhRleHBlY3RlZEFnZW50VmVyc2lvbhIjCg1j'
    'b21taXR0ZWRfc2hhGAwgASgJUgxjb21taXR0ZWRTaGESOQoKY3JlYXRlZF9hdBgeIAEoCzIaLm'
    'dvb2dsZS5wcm90b2J1Zi5UaW1lc3RhbXBSCWNyZWF0ZWRBdBI5Cgp1cGRhdGVkX2F0GB8gASgL'
    'MhouZ29vZ2xlLnByb3RvYnVmLlRpbWVzdGFtcFIJdXBkYXRlZEF0EiEKDHdvcmtzcGFjZV9pZB'
    'hkIAEoCVILd29ya3NwYWNlSWQ=');

@$core.Deprecated('Use agentOperationStepDescriptor instead')
const AgentOperationStep$json = {
  '1': 'AgentOperationStep',
  '2': [
    {
      '1': 'kind',
      '3': 1,
      '4': 1,
      '5': 14,
      '6': '.agents.v1.AgentOperationStepKind',
      '10': 'kind'
    },
    {
      '1': 'status',
      '3': 2,
      '4': 1,
      '5': 14,
      '6': '.agents.v1.AgentOperationStepStatus',
      '10': 'status'
    },
    {'1': 'attempt_count', '3': 3, '4': 1, '5': 5, '10': 'attemptCount'},
    {'1': 'error', '3': 4, '4': 1, '5': 9, '10': 'error'},
    {
      '1': 'started_at',
      '3': 5,
      '4': 1,
      '5': 11,
      '6': '.google.protobuf.Timestamp',
      '10': 'startedAt'
    },
    {
      '1': 'finished_at',
      '3': 6,
      '4': 1,
      '5': 11,
      '6': '.google.protobuf.Timestamp',
      '10': 'finishedAt'
    },
  ],
};

/// Descriptor for `AgentOperationStep`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List agentOperationStepDescriptor = $convert.base64Decode(
    'ChJBZ2VudE9wZXJhdGlvblN0ZXASNQoEa2luZBgBIAEoDjIhLmFnZW50cy52MS5BZ2VudE9wZX'
    'JhdGlvblN0ZXBLaW5kUgRraW5kEjsKBnN0YXR1cxgCIAEoDjIjLmFnZW50cy52MS5BZ2VudE9w'
    'ZXJhdGlvblN0ZXBTdGF0dXNSBnN0YXR1cxIjCg1hdHRlbXB0X2NvdW50GAMgASgFUgxhdHRlbX'
    'B0Q291bnQSFAoFZXJyb3IYBCABKAlSBWVycm9yEjkKCnN0YXJ0ZWRfYXQYBSABKAsyGi5nb29n'
    'bGUucHJvdG9idWYuVGltZXN0YW1wUglzdGFydGVkQXQSOwoLZmluaXNoZWRfYXQYBiABKAsyGi'
    '5nb29nbGUucHJvdG9idWYuVGltZXN0YW1wUgpmaW5pc2hlZEF0');
