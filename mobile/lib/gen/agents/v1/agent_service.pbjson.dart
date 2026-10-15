// This is a generated file - do not edit.
//
// Generated from agents/v1/agent_service.proto.

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
    as $7;
import 'package:protobuf/well_known_types/google/protobuf/struct.pbjson.dart'
    as $6;
import 'package:protobuf/well_known_types/google/protobuf/timestamp.pbjson.dart'
    as $1;

import 'agent.pbjson.dart' as $0;
import 'agent_file.pbjson.dart' as $9;
import 'agent_operation.pbjson.dart' as $3;
import 'agentchannel.pbjson.dart' as $5;
import 'content.pbjson.dart' as $2;
import 'dashboard.pbjson.dart' as $8;
import 'repobinding.pbjson.dart' as $4;

@$core.Deprecated('Use agentRuntimeStateDescriptor instead')
const AgentRuntimeState$json = {
  '1': 'AgentRuntimeState',
  '2': [
    {'1': 'AGENT_RUNTIME_STATE_UNSPECIFIED', '2': 0},
    {'1': 'AGENT_RUNTIME_STATE_IDLE', '2': 1},
    {'1': 'AGENT_RUNTIME_STATE_RUNNING', '2': 2},
    {'1': 'AGENT_RUNTIME_STATE_FAILED', '2': 3},
  ],
};

/// Descriptor for `AgentRuntimeState`. Decode as a `google.protobuf.EnumDescriptorProto`.
final $typed_data.Uint8List agentRuntimeStateDescriptor = $convert.base64Decode(
    'ChFBZ2VudFJ1bnRpbWVTdGF0ZRIjCh9BR0VOVF9SVU5USU1FX1NUQVRFX1VOU1BFQ0lGSUVEEA'
    'ASHAoYQUdFTlRfUlVOVElNRV9TVEFURV9JRExFEAESHwobQUdFTlRfUlVOVElNRV9TVEFURV9S'
    'VU5OSU5HEAISHgoaQUdFTlRfUlVOVElNRV9TVEFURV9GQUlMRUQQAw==');

@$core.Deprecated('Use migrationReadinessDescriptor instead')
const MigrationReadiness$json = {
  '1': 'MigrationReadiness',
  '2': [
    {'1': 'MIGRATION_READINESS_UNSPECIFIED', '2': 0},
    {'1': 'MIGRATION_READINESS_READY', '2': 1},
    {'1': 'MIGRATION_READINESS_MISSING_ID', '2': 2},
    {'1': 'MIGRATION_READINESS_CONFLICT', '2': 3},
    {'1': 'MIGRATION_READINESS_INCOMPLETE_DEPS', '2': 4},
  ],
};

/// Descriptor for `MigrationReadiness`. Decode as a `google.protobuf.EnumDescriptorProto`.
final $typed_data.Uint8List migrationReadinessDescriptor = $convert.base64Decode(
    'ChJNaWdyYXRpb25SZWFkaW5lc3MSIwofTUlHUkFUSU9OX1JFQURJTkVTU19VTlNQRUNJRklFRB'
    'AAEh0KGU1JR1JBVElPTl9SRUFESU5FU1NfUkVBRFkQARIiCh5NSUdSQVRJT05fUkVBRElORVNT'
    'X01JU1NJTkdfSUQQAhIgChxNSUdSQVRJT05fUkVBRElORVNTX0NPTkZMSUNUEAMSJwojTUlHUk'
    'FUSU9OX1JFQURJTkVTU19JTkNPTVBMRVRFX0RFUFMQBA==');

@$core.Deprecated('Use migrateModeDescriptor instead')
const MigrateMode$json = {
  '1': 'MigrateMode',
  '2': [
    {'1': 'MIGRATE_MODE_UNSPECIFIED', '2': 0},
    {'1': 'MIGRATE_MODE_DRY_RUN', '2': 1},
    {'1': 'MIGRATE_MODE_APPLY', '2': 2},
    {'1': 'MIGRATE_MODE_VERIFY', '2': 3},
  ],
};

/// Descriptor for `MigrateMode`. Decode as a `google.protobuf.EnumDescriptorProto`.
final $typed_data.Uint8List migrateModeDescriptor = $convert.base64Decode(
    'CgtNaWdyYXRlTW9kZRIcChhNSUdSQVRFX01PREVfVU5TUEVDSUZJRUQQABIYChRNSUdSQVRFX0'
    '1PREVfRFJZX1JVThABEhYKEk1JR1JBVEVfTU9ERV9BUFBMWRACEhcKE01JR1JBVEVfTU9ERV9W'
    'RVJJRlkQAw==');

@$core.Deprecated('Use invocationStatusDescriptor instead')
const InvocationStatus$json = {
  '1': 'InvocationStatus',
  '2': [
    {'1': 'INVOCATION_STATUS_UNSPECIFIED', '2': 0},
    {'1': 'INVOCATION_STATUS_RUNNING', '2': 1},
    {'1': 'INVOCATION_STATUS_SUCCEEDED', '2': 2},
    {'1': 'INVOCATION_STATUS_FAILED', '2': 3},
    {'1': 'INVOCATION_STATUS_QUEUED', '2': 4},
    {'1': 'INVOCATION_STATUS_CANCELLED', '2': 5},
  ],
};

/// Descriptor for `InvocationStatus`. Decode as a `google.protobuf.EnumDescriptorProto`.
final $typed_data.Uint8List invocationStatusDescriptor = $convert.base64Decode(
    'ChBJbnZvY2F0aW9uU3RhdHVzEiEKHUlOVk9DQVRJT05fU1RBVFVTX1VOU1BFQ0lGSUVEEAASHQ'
    'oZSU5WT0NBVElPTl9TVEFUVVNfUlVOTklORxABEh8KG0lOVk9DQVRJT05fU1RBVFVTX1NVQ0NF'
    'RURFRBACEhwKGElOVk9DQVRJT05fU1RBVFVTX0ZBSUxFRBADEhwKGElOVk9DQVRJT05fU1RBVF'
    'VTX1FVRVVFRBAEEh8KG0lOVk9DQVRJT05fU1RBVFVTX0NBTkNFTExFRBAF');

@$core.Deprecated('Use mCPOAuthConnectionStateDescriptor instead')
const MCPOAuthConnectionState$json = {
  '1': 'MCPOAuthConnectionState',
  '2': [
    {'1': 'MCPO_AUTH_CONNECTION_STATE_UNSPECIFIED', '2': 0},
    {'1': 'MCPO_AUTH_CONNECTION_STATE_DISCONNECTED', '2': 1},
    {'1': 'MCPO_AUTH_CONNECTION_STATE_CONNECTED', '2': 2},
    {'1': 'MCPO_AUTH_CONNECTION_STATE_REAUTHORIZATION_REQUIRED', '2': 3},
    {'1': 'MCPO_AUTH_CONNECTION_STATE_ERROR', '2': 4},
  ],
};

/// Descriptor for `MCPOAuthConnectionState`. Decode as a `google.protobuf.EnumDescriptorProto`.
final $typed_data.Uint8List mCPOAuthConnectionStateDescriptor = $convert.base64Decode(
    'ChdNQ1BPQXV0aENvbm5lY3Rpb25TdGF0ZRIqCiZNQ1BPX0FVVEhfQ09OTkVDVElPTl9TVEFURV'
    '9VTlNQRUNJRklFRBAAEisKJ01DUE9fQVVUSF9DT05ORUNUSU9OX1NUQVRFX0RJU0NPTk5FQ1RF'
    'RBABEigKJE1DUE9fQVVUSF9DT05ORUNUSU9OX1NUQVRFX0NPTk5FQ1RFRBACEjcKM01DUE9fQV'
    'VUSF9DT05ORUNUSU9OX1NUQVRFX1JFQVVUSE9SSVpBVElPTl9SRVFVSVJFRBADEiQKIE1DUE9f'
    'QVVUSF9DT05ORUNUSU9OX1NUQVRFX0VSUk9SEAQ=');

@$core.Deprecated('Use listAgentsRequestDescriptor instead')
const ListAgentsRequest$json = {
  '1': 'ListAgentsRequest',
  '2': [
    {'1': 'page_size', '3': 1, '4': 1, '5': 5, '10': 'pageSize'},
    {'1': 'page_token', '3': 2, '4': 1, '5': 9, '10': 'pageToken'},
  ],
};

/// Descriptor for `ListAgentsRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listAgentsRequestDescriptor = $convert.base64Decode(
    'ChFMaXN0QWdlbnRzUmVxdWVzdBIbCglwYWdlX3NpemUYASABKAVSCHBhZ2VTaXplEh0KCnBhZ2'
    'VfdG9rZW4YAiABKAlSCXBhZ2VUb2tlbg==');

@$core.Deprecated('Use listAgentsResponseDescriptor instead')
const ListAgentsResponse$json = {
  '1': 'ListAgentsResponse',
  '2': [
    {
      '1': 'agents',
      '3': 1,
      '4': 3,
      '5': 11,
      '6': '.agents.v1.Agent',
      '10': 'agents'
    },
    {'1': 'next_page_token', '3': 2, '4': 1, '5': 9, '10': 'nextPageToken'},
    {'1': 'total', '3': 3, '4': 1, '5': 5, '10': 'total'},
  ],
};

/// Descriptor for `ListAgentsResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listAgentsResponseDescriptor = $convert.base64Decode(
    'ChJMaXN0QWdlbnRzUmVzcG9uc2USKAoGYWdlbnRzGAEgAygLMhAuYWdlbnRzLnYxLkFnZW50Ug'
    'ZhZ2VudHMSJgoPbmV4dF9wYWdlX3Rva2VuGAIgASgJUg1uZXh0UGFnZVRva2VuEhQKBXRvdGFs'
    'GAMgASgFUgV0b3RhbA==');

@$core.Deprecated('Use reloadAgentsRequestDescriptor instead')
const ReloadAgentsRequest$json = {
  '1': 'ReloadAgentsRequest',
};

/// Descriptor for `ReloadAgentsRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List reloadAgentsRequestDescriptor =
    $convert.base64Decode('ChNSZWxvYWRBZ2VudHNSZXF1ZXN0');

@$core.Deprecated('Use reloadAgentsResponseDescriptor instead')
const ReloadAgentsResponse$json = {
  '1': 'ReloadAgentsResponse',
  '2': [
    {
      '1': 'reloaded_at',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.google.protobuf.Timestamp',
      '10': 'reloadedAt'
    },
  ],
};

/// Descriptor for `ReloadAgentsResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List reloadAgentsResponseDescriptor = $convert.base64Decode(
    'ChRSZWxvYWRBZ2VudHNSZXNwb25zZRI7CgtyZWxvYWRlZF9hdBgBIAEoCzIaLmdvb2dsZS5wcm'
    '90b2J1Zi5UaW1lc3RhbXBSCnJlbG9hZGVkQXQ=');

@$core.Deprecated('Use agentRuntimeStatusDescriptor instead')
const AgentRuntimeStatus$json = {
  '1': 'AgentRuntimeStatus',
  '2': [
    {'1': 'name', '3': 1, '4': 1, '5': 9, '10': 'name'},
    {
      '1': 'state',
      '3': 2,
      '4': 1,
      '5': 14,
      '6': '.agents.v1.AgentRuntimeState',
      '10': 'state'
    },
    {
      '1': 'last_run_at',
      '3': 3,
      '4': 1,
      '5': 11,
      '6': '.google.protobuf.Timestamp',
      '10': 'lastRunAt'
    },
    {
      '1': 'last_invocation_id',
      '3': 4,
      '4': 1,
      '5': 9,
      '10': 'lastInvocationId'
    },
    {'1': 'in_flight', '3': 5, '4': 1, '5': 5, '10': 'inFlight'},
    {'1': 'agent_id', '3': 6, '4': 1, '5': 9, '10': 'agentId'},
  ],
};

/// Descriptor for `AgentRuntimeStatus`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List agentRuntimeStatusDescriptor = $convert.base64Decode(
    'ChJBZ2VudFJ1bnRpbWVTdGF0dXMSEgoEbmFtZRgBIAEoCVIEbmFtZRIyCgVzdGF0ZRgCIAEoDj'
    'IcLmFnZW50cy52MS5BZ2VudFJ1bnRpbWVTdGF0ZVIFc3RhdGUSOgoLbGFzdF9ydW5fYXQYAyAB'
    'KAsyGi5nb29nbGUucHJvdG9idWYuVGltZXN0YW1wUglsYXN0UnVuQXQSLAoSbGFzdF9pbnZvY2'
    'F0aW9uX2lkGAQgASgJUhBsYXN0SW52b2NhdGlvbklkEhsKCWluX2ZsaWdodBgFIAEoBVIIaW5G'
    'bGlnaHQSGQoIYWdlbnRfaWQYBiABKAlSB2FnZW50SWQ=');

@$core.Deprecated('Use getAgentRuntimeStatusRequestDescriptor instead')
const GetAgentRuntimeStatusRequest$json = {
  '1': 'GetAgentRuntimeStatusRequest',
  '2': [
    {
      '1': 'name',
      '3': 1,
      '4': 1,
      '5': 9,
      '8': {'3': true},
      '10': 'name',
    },
    {'1': 'agent_id', '3': 2, '4': 1, '5': 9, '10': 'agentId'},
  ],
};

/// Descriptor for `GetAgentRuntimeStatusRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getAgentRuntimeStatusRequestDescriptor =
    $convert.base64Decode(
        'ChxHZXRBZ2VudFJ1bnRpbWVTdGF0dXNSZXF1ZXN0EhYKBG5hbWUYASABKAlCAhgBUgRuYW1lEh'
        'kKCGFnZW50X2lkGAIgASgJUgdhZ2VudElk');

@$core.Deprecated('Use getAgentRuntimeStatusResponseDescriptor instead')
const GetAgentRuntimeStatusResponse$json = {
  '1': 'GetAgentRuntimeStatusResponse',
  '2': [
    {
      '1': 'status',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.AgentRuntimeStatus',
      '10': 'status'
    },
  ],
};

/// Descriptor for `GetAgentRuntimeStatusResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getAgentRuntimeStatusResponseDescriptor =
    $convert.base64Decode(
        'Ch1HZXRBZ2VudFJ1bnRpbWVTdGF0dXNSZXNwb25zZRI1CgZzdGF0dXMYASABKAsyHS5hZ2VudH'
        'MudjEuQWdlbnRSdW50aW1lU3RhdHVzUgZzdGF0dXM=');

@$core.Deprecated('Use listAgentRuntimeStatusesRequestDescriptor instead')
const ListAgentRuntimeStatusesRequest$json = {
  '1': 'ListAgentRuntimeStatusesRequest',
  '2': [
    {
      '1': 'names',
      '3': 1,
      '4': 3,
      '5': 9,
      '8': {'3': true},
      '10': 'names',
    },
    {'1': 'agent_ids', '3': 2, '4': 3, '5': 9, '10': 'agentIds'},
  ],
};

/// Descriptor for `ListAgentRuntimeStatusesRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listAgentRuntimeStatusesRequestDescriptor =
    $convert.base64Decode(
        'Ch9MaXN0QWdlbnRSdW50aW1lU3RhdHVzZXNSZXF1ZXN0EhgKBW5hbWVzGAEgAygJQgIYAVIFbm'
        'FtZXMSGwoJYWdlbnRfaWRzGAIgAygJUghhZ2VudElkcw==');

@$core.Deprecated('Use listAgentRuntimeStatusesResponseDescriptor instead')
const ListAgentRuntimeStatusesResponse$json = {
  '1': 'ListAgentRuntimeStatusesResponse',
  '2': [
    {
      '1': 'statuses',
      '3': 1,
      '4': 3,
      '5': 11,
      '6': '.agents.v1.AgentRuntimeStatus',
      '10': 'statuses'
    },
  ],
};

/// Descriptor for `ListAgentRuntimeStatusesResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listAgentRuntimeStatusesResponseDescriptor =
    $convert.base64Decode(
        'CiBMaXN0QWdlbnRSdW50aW1lU3RhdHVzZXNSZXNwb25zZRI5CghzdGF0dXNlcxgBIAMoCzIdLm'
        'FnZW50cy52MS5BZ2VudFJ1bnRpbWVTdGF0dXNSCHN0YXR1c2Vz');

@$core.Deprecated('Use cancelAgentInvocationRequestDescriptor instead')
const CancelAgentInvocationRequest$json = {
  '1': 'CancelAgentInvocationRequest',
  '2': [
    {'1': 'invocation_id', '3': 1, '4': 1, '5': 9, '10': 'invocationId'},
  ],
};

/// Descriptor for `CancelAgentInvocationRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List cancelAgentInvocationRequestDescriptor =
    $convert.base64Decode(
        'ChxDYW5jZWxBZ2VudEludm9jYXRpb25SZXF1ZXN0EiMKDWludm9jYXRpb25faWQYASABKAlSDG'
        'ludm9jYXRpb25JZA==');

@$core.Deprecated('Use cancelAgentInvocationResponseDescriptor instead')
const CancelAgentInvocationResponse$json = {
  '1': 'CancelAgentInvocationResponse',
  '2': [
    {'1': 'cancelled', '3': 1, '4': 1, '5': 8, '10': 'cancelled'},
  ],
};

/// Descriptor for `CancelAgentInvocationResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List cancelAgentInvocationResponseDescriptor =
    $convert.base64Decode(
        'Ch1DYW5jZWxBZ2VudEludm9jYXRpb25SZXNwb25zZRIcCgljYW5jZWxsZWQYASABKAhSCWNhbm'
        'NlbGxlZA==');

@$core.Deprecated('Use streamAgentRequestDescriptor instead')
const StreamAgentRequest$json = {
  '1': 'StreamAgentRequest',
  '2': [
    {
      '1': 'agent_name',
      '3': 1,
      '4': 1,
      '5': 9,
      '8': {'3': true},
      '10': 'agentName',
    },
    {'1': 'message', '3': 2, '4': 1, '5': 9, '10': 'message'},
    {'1': 'app_name', '3': 3, '4': 1, '5': 9, '10': 'appName'},
    {'1': 'user_id', '3': 4, '4': 1, '5': 9, '10': 'userId'},
    {'1': 'session_id', '3': 5, '4': 1, '5': 9, '10': 'sessionId'},
    {'1': 'model_override', '3': 6, '4': 1, '5': 9, '10': 'modelOverride'},
    {
      '1': 'parts',
      '3': 7,
      '4': 3,
      '5': 11,
      '6': '.agents.v1.InputPart',
      '10': 'parts'
    },
    {'1': 'agent_id', '3': 8, '4': 1, '5': 9, '10': 'agentId'},
  ],
};

/// Descriptor for `StreamAgentRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List streamAgentRequestDescriptor = $convert.base64Decode(
    'ChJTdHJlYW1BZ2VudFJlcXVlc3QSIQoKYWdlbnRfbmFtZRgBIAEoCUICGAFSCWFnZW50TmFtZR'
    'IYCgdtZXNzYWdlGAIgASgJUgdtZXNzYWdlEhkKCGFwcF9uYW1lGAMgASgJUgdhcHBOYW1lEhcK'
    'B3VzZXJfaWQYBCABKAlSBnVzZXJJZBIdCgpzZXNzaW9uX2lkGAUgASgJUglzZXNzaW9uSWQSJQ'
    'oObW9kZWxfb3ZlcnJpZGUYBiABKAlSDW1vZGVsT3ZlcnJpZGUSKgoFcGFydHMYByADKAsyFC5h'
    'Z2VudHMudjEuSW5wdXRQYXJ0UgVwYXJ0cxIZCghhZ2VudF9pZBgIIAEoCVIHYWdlbnRJZA==');

@$core.Deprecated('Use streamAgentResponseDescriptor instead')
const StreamAgentResponse$json = {
  '1': 'StreamAgentResponse',
  '2': [
    {
      '1': 'started',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.StreamAgentStarted',
      '9': 0,
      '10': 'started'
    },
    {
      '1': 'run_event',
      '3': 2,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.StreamAgentRunEvent',
      '9': 0,
      '10': 'runEvent'
    },
    {
      '1': 'text_delta',
      '3': 3,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.StreamAgentTextDelta',
      '9': 0,
      '10': 'textDelta'
    },
    {
      '1': 'final',
      '3': 4,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.StreamAgentFinal',
      '9': 0,
      '10': 'final'
    },
  ],
  '8': [
    {'1': 'event'},
  ],
};

/// Descriptor for `StreamAgentResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List streamAgentResponseDescriptor = $convert.base64Decode(
    'ChNTdHJlYW1BZ2VudFJlc3BvbnNlEjkKB3N0YXJ0ZWQYASABKAsyHS5hZ2VudHMudjEuU3RyZW'
    'FtQWdlbnRTdGFydGVkSABSB3N0YXJ0ZWQSPQoJcnVuX2V2ZW50GAIgASgLMh4uYWdlbnRzLnYx'
    'LlN0cmVhbUFnZW50UnVuRXZlbnRIAFIIcnVuRXZlbnQSQAoKdGV4dF9kZWx0YRgDIAEoCzIfLm'
    'FnZW50cy52MS5TdHJlYW1BZ2VudFRleHREZWx0YUgAUgl0ZXh0RGVsdGESMwoFZmluYWwYBCAB'
    'KAsyGy5hZ2VudHMudjEuU3RyZWFtQWdlbnRGaW5hbEgAUgVmaW5hbEIHCgVldmVudA==');

@$core.Deprecated('Use streamAgentStartedDescriptor instead')
const StreamAgentStarted$json = {
  '1': 'StreamAgentStarted',
  '2': [
    {'1': 'invocation_id', '3': 1, '4': 1, '5': 9, '10': 'invocationId'},
    {'1': 'session_id', '3': 2, '4': 1, '5': 9, '10': 'sessionId'},
    {'1': 'agent_name', '3': 3, '4': 1, '5': 9, '10': 'agentName'},
    {'1': 'agent_id', '3': 4, '4': 1, '5': 9, '10': 'agentId'},
  ],
};

/// Descriptor for `StreamAgentStarted`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List streamAgentStartedDescriptor = $convert.base64Decode(
    'ChJTdHJlYW1BZ2VudFN0YXJ0ZWQSIwoNaW52b2NhdGlvbl9pZBgBIAEoCVIMaW52b2NhdGlvbk'
    'lkEh0KCnNlc3Npb25faWQYAiABKAlSCXNlc3Npb25JZBIdCgphZ2VudF9uYW1lGAMgASgJUglh'
    'Z2VudE5hbWUSGQoIYWdlbnRfaWQYBCABKAlSB2FnZW50SWQ=');

@$core.Deprecated('Use streamAgentRunEventDescriptor instead')
const StreamAgentRunEvent$json = {
  '1': 'StreamAgentRunEvent',
  '2': [
    {'1': 'invocation_id', '3': 1, '4': 1, '5': 9, '10': 'invocationId'},
    {'1': 'session_id', '3': 2, '4': 1, '5': 9, '10': 'sessionId'},
    {'1': 'agent_name', '3': 3, '4': 1, '5': 9, '10': 'agentName'},
    {'1': 'event_id', '3': 4, '4': 1, '5': 9, '10': 'eventId'},
    {'1': 'author', '3': 5, '4': 1, '5': 9, '10': 'author'},
    {'1': 'branch', '3': 6, '4': 1, '5': 9, '10': 'branch'},
    {'1': 'partial', '3': 7, '4': 1, '5': 8, '10': 'partial'},
    {'1': 'final_response', '3': 8, '4': 1, '5': 8, '10': 'finalResponse'},
    {'1': 'content_json', '3': 9, '4': 1, '5': 9, '10': 'contentJson'},
    {
      '1': 'timestamp',
      '3': 10,
      '4': 1,
      '5': 11,
      '6': '.google.protobuf.Timestamp',
      '10': 'timestamp'
    },
    {'1': 'agent_id', '3': 11, '4': 1, '5': 9, '10': 'agentId'},
  ],
};

/// Descriptor for `StreamAgentRunEvent`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List streamAgentRunEventDescriptor = $convert.base64Decode(
    'ChNTdHJlYW1BZ2VudFJ1bkV2ZW50EiMKDWludm9jYXRpb25faWQYASABKAlSDGludm9jYXRpb2'
    '5JZBIdCgpzZXNzaW9uX2lkGAIgASgJUglzZXNzaW9uSWQSHQoKYWdlbnRfbmFtZRgDIAEoCVIJ'
    'YWdlbnROYW1lEhkKCGV2ZW50X2lkGAQgASgJUgdldmVudElkEhYKBmF1dGhvchgFIAEoCVIGYX'
    'V0aG9yEhYKBmJyYW5jaBgGIAEoCVIGYnJhbmNoEhgKB3BhcnRpYWwYByABKAhSB3BhcnRpYWwS'
    'JQoOZmluYWxfcmVzcG9uc2UYCCABKAhSDWZpbmFsUmVzcG9uc2USIQoMY29udGVudF9qc29uGA'
    'kgASgJUgtjb250ZW50SnNvbhI4Cgl0aW1lc3RhbXAYCiABKAsyGi5nb29nbGUucHJvdG9idWYu'
    'VGltZXN0YW1wUgl0aW1lc3RhbXASGQoIYWdlbnRfaWQYCyABKAlSB2FnZW50SWQ=');

@$core.Deprecated('Use streamAgentTextDeltaDescriptor instead')
const StreamAgentTextDelta$json = {
  '1': 'StreamAgentTextDelta',
  '2': [
    {'1': 'invocation_id', '3': 1, '4': 1, '5': 9, '10': 'invocationId'},
    {'1': 'session_id', '3': 2, '4': 1, '5': 9, '10': 'sessionId'},
    {'1': 'agent_name', '3': 3, '4': 1, '5': 9, '10': 'agentName'},
    {'1': 'text', '3': 4, '4': 1, '5': 9, '10': 'text'},
    {'1': 'agent_id', '3': 5, '4': 1, '5': 9, '10': 'agentId'},
  ],
};

/// Descriptor for `StreamAgentTextDelta`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List streamAgentTextDeltaDescriptor = $convert.base64Decode(
    'ChRTdHJlYW1BZ2VudFRleHREZWx0YRIjCg1pbnZvY2F0aW9uX2lkGAEgASgJUgxpbnZvY2F0aW'
    '9uSWQSHQoKc2Vzc2lvbl9pZBgCIAEoCVIJc2Vzc2lvbklkEh0KCmFnZW50X25hbWUYAyABKAlS'
    'CWFnZW50TmFtZRISCgR0ZXh0GAQgASgJUgR0ZXh0EhkKCGFnZW50X2lkGAUgASgJUgdhZ2VudE'
    'lk');

@$core.Deprecated('Use streamAgentFinalDescriptor instead')
const StreamAgentFinal$json = {
  '1': 'StreamAgentFinal',
  '2': [
    {'1': 'invocation_id', '3': 1, '4': 1, '5': 9, '10': 'invocationId'},
    {'1': 'session_id', '3': 2, '4': 1, '5': 9, '10': 'sessionId'},
    {'1': 'agent_name', '3': 3, '4': 1, '5': 9, '10': 'agentName'},
    {'1': 'response', '3': 4, '4': 1, '5': 9, '10': 'response'},
    {'1': 'agent_id', '3': 5, '4': 1, '5': 9, '10': 'agentId'},
  ],
};

/// Descriptor for `StreamAgentFinal`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List streamAgentFinalDescriptor = $convert.base64Decode(
    'ChBTdHJlYW1BZ2VudEZpbmFsEiMKDWludm9jYXRpb25faWQYASABKAlSDGludm9jYXRpb25JZB'
    'IdCgpzZXNzaW9uX2lkGAIgASgJUglzZXNzaW9uSWQSHQoKYWdlbnRfbmFtZRgDIAEoCVIJYWdl'
    'bnROYW1lEhoKCHJlc3BvbnNlGAQgASgJUghyZXNwb25zZRIZCghhZ2VudF9pZBgFIAEoCVIHYW'
    'dlbnRJZA==');

@$core.Deprecated('Use getAgentRequestDescriptor instead')
const GetAgentRequest$json = {
  '1': 'GetAgentRequest',
  '2': [
    {
      '1': 'name',
      '3': 1,
      '4': 1,
      '5': 9,
      '8': {'3': true},
      '10': 'name',
    },
    {'1': 'agent_id', '3': 2, '4': 1, '5': 9, '10': 'agentId'},
  ],
};

/// Descriptor for `GetAgentRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getAgentRequestDescriptor = $convert.base64Decode(
    'Cg9HZXRBZ2VudFJlcXVlc3QSFgoEbmFtZRgBIAEoCUICGAFSBG5hbWUSGQoIYWdlbnRfaWQYAi'
    'ABKAlSB2FnZW50SWQ=');

@$core.Deprecated('Use getAgentResponseDescriptor instead')
const GetAgentResponse$json = {
  '1': 'GetAgentResponse',
  '2': [
    {
      '1': 'agent',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.Agent',
      '10': 'agent'
    },
  ],
};

/// Descriptor for `GetAgentResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getAgentResponseDescriptor = $convert.base64Decode(
    'ChBHZXRBZ2VudFJlc3BvbnNlEiYKBWFnZW50GAEgASgLMhAuYWdlbnRzLnYxLkFnZW50UgVhZ2'
    'VudA==');

@$core.Deprecated('Use createAgentRequestDescriptor instead')
const CreateAgentRequest$json = {
  '1': 'CreateAgentRequest',
  '2': [
    {
      '1': 'agent',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.Agent',
      '10': 'agent'
    },
    {
      '1': 'initial_content',
      '3': 2,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.AgentContentInput',
      '10': 'initialContent'
    },
    {'1': 'operation_id', '3': 3, '4': 1, '5': 9, '10': 'operationId'},
  ],
};

/// Descriptor for `CreateAgentRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List createAgentRequestDescriptor = $convert.base64Decode(
    'ChJDcmVhdGVBZ2VudFJlcXVlc3QSJgoFYWdlbnQYASABKAsyEC5hZ2VudHMudjEuQWdlbnRSBW'
    'FnZW50EkUKD2luaXRpYWxfY29udGVudBgCIAEoCzIcLmFnZW50cy52MS5BZ2VudENvbnRlbnRJ'
    'bnB1dFIOaW5pdGlhbENvbnRlbnQSIQoMb3BlcmF0aW9uX2lkGAMgASgJUgtvcGVyYXRpb25JZA'
    '==');

@$core.Deprecated('Use createAgentResponseDescriptor instead')
const CreateAgentResponse$json = {
  '1': 'CreateAgentResponse',
  '2': [
    {
      '1': 'agent',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.Agent',
      '10': 'agent'
    },
    {
      '1': 'operation',
      '3': 2,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.AgentOperation',
      '10': 'operation'
    },
  ],
};

/// Descriptor for `CreateAgentResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List createAgentResponseDescriptor = $convert.base64Decode(
    'ChNDcmVhdGVBZ2VudFJlc3BvbnNlEiYKBWFnZW50GAEgASgLMhAuYWdlbnRzLnYxLkFnZW50Ug'
    'VhZ2VudBI3CglvcGVyYXRpb24YAiABKAsyGS5hZ2VudHMudjEuQWdlbnRPcGVyYXRpb25SCW9w'
    'ZXJhdGlvbg==');

@$core.Deprecated('Use agentContentInputDescriptor instead')
const AgentContentInput$json = {
  '1': 'AgentContentInput',
  '2': [
    {'1': 'description', '3': 1, '4': 1, '5': 9, '10': 'description'},
    {'1': 'prompt', '3': 2, '4': 1, '5': 9, '10': 'prompt'},
    {'1': 'global_prompt', '3': 3, '4': 1, '5': 9, '10': 'globalPrompt'},
  ],
};

/// Descriptor for `AgentContentInput`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List agentContentInputDescriptor = $convert.base64Decode(
    'ChFBZ2VudENvbnRlbnRJbnB1dBIgCgtkZXNjcmlwdGlvbhgBIAEoCVILZGVzY3JpcHRpb24SFg'
    'oGcHJvbXB0GAIgASgJUgZwcm9tcHQSIwoNZ2xvYmFsX3Byb21wdBgDIAEoCVIMZ2xvYmFsUHJv'
    'bXB0');

@$core.Deprecated('Use updateAgentConfigurationRequestDescriptor instead')
const UpdateAgentConfigurationRequest$json = {
  '1': 'UpdateAgentConfigurationRequest',
  '2': [
    {
      '1': 'agent_patch',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.Agent',
      '10': 'agentPatch'
    },
    {
      '1': 'content_changes',
      '3': 2,
      '4': 3,
      '5': 11,
      '6': '.agents.v1.ContentFileAction',
      '10': 'contentChanges'
    },
    {
      '1': 'expected_agent_version',
      '3': 3,
      '4': 1,
      '5': 3,
      '10': 'expectedAgentVersion'
    },
    {'1': 'base_commit_sha', '3': 4, '4': 1, '5': 9, '10': 'baseCommitSha'},
    {'1': 'operation_id', '3': 5, '4': 1, '5': 9, '10': 'operationId'},
  ],
};

/// Descriptor for `UpdateAgentConfigurationRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List updateAgentConfigurationRequestDescriptor = $convert.base64Decode(
    'Ch9VcGRhdGVBZ2VudENvbmZpZ3VyYXRpb25SZXF1ZXN0EjEKC2FnZW50X3BhdGNoGAEgASgLMh'
    'AuYWdlbnRzLnYxLkFnZW50UgphZ2VudFBhdGNoEkUKD2NvbnRlbnRfY2hhbmdlcxgCIAMoCzIc'
    'LmFnZW50cy52MS5Db250ZW50RmlsZUFjdGlvblIOY29udGVudENoYW5nZXMSNAoWZXhwZWN0ZW'
    'RfYWdlbnRfdmVyc2lvbhgDIAEoA1IUZXhwZWN0ZWRBZ2VudFZlcnNpb24SJgoPYmFzZV9jb21t'
    'aXRfc2hhGAQgASgJUg1iYXNlQ29tbWl0U2hhEiEKDG9wZXJhdGlvbl9pZBgFIAEoCVILb3Blcm'
    'F0aW9uSWQ=');

@$core.Deprecated('Use updateAgentConfigurationResponseDescriptor instead')
const UpdateAgentConfigurationResponse$json = {
  '1': 'UpdateAgentConfigurationResponse',
  '2': [
    {
      '1': 'agent',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.Agent',
      '10': 'agent'
    },
    {
      '1': 'operation',
      '3': 2,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.AgentOperation',
      '10': 'operation'
    },
    {
      '1': 'validation_errors',
      '3': 3,
      '4': 3,
      '5': 9,
      '10': 'validationErrors'
    },
  ],
};

/// Descriptor for `UpdateAgentConfigurationResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List updateAgentConfigurationResponseDescriptor =
    $convert.base64Decode(
        'CiBVcGRhdGVBZ2VudENvbmZpZ3VyYXRpb25SZXNwb25zZRImCgVhZ2VudBgBIAEoCzIQLmFnZW'
        '50cy52MS5BZ2VudFIFYWdlbnQSNwoJb3BlcmF0aW9uGAIgASgLMhkuYWdlbnRzLnYxLkFnZW50'
        'T3BlcmF0aW9uUglvcGVyYXRpb24SKwoRdmFsaWRhdGlvbl9lcnJvcnMYAyADKAlSEHZhbGlkYX'
        'Rpb25FcnJvcnM=');

@$core.Deprecated('Use restoreAgentRequestDescriptor instead')
const RestoreAgentRequest$json = {
  '1': 'RestoreAgentRequest',
  '2': [
    {'1': 'agent_id', '3': 1, '4': 1, '5': 9, '10': 'agentId'},
    {'1': 'operation_id', '3': 2, '4': 1, '5': 9, '10': 'operationId'},
  ],
};

/// Descriptor for `RestoreAgentRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List restoreAgentRequestDescriptor = $convert.base64Decode(
    'ChNSZXN0b3JlQWdlbnRSZXF1ZXN0EhkKCGFnZW50X2lkGAEgASgJUgdhZ2VudElkEiEKDG9wZX'
    'JhdGlvbl9pZBgCIAEoCVILb3BlcmF0aW9uSWQ=');

@$core.Deprecated('Use restoreAgentResponseDescriptor instead')
const RestoreAgentResponse$json = {
  '1': 'RestoreAgentResponse',
  '2': [
    {
      '1': 'agent',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.Agent',
      '10': 'agent'
    },
    {
      '1': 'operation',
      '3': 2,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.AgentOperation',
      '10': 'operation'
    },
    {
      '1': 'validation_errors',
      '3': 3,
      '4': 3,
      '5': 9,
      '10': 'validationErrors'
    },
  ],
};

/// Descriptor for `RestoreAgentResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List restoreAgentResponseDescriptor = $convert.base64Decode(
    'ChRSZXN0b3JlQWdlbnRSZXNwb25zZRImCgVhZ2VudBgBIAEoCzIQLmFnZW50cy52MS5BZ2VudF'
    'IFYWdlbnQSNwoJb3BlcmF0aW9uGAIgASgLMhkuYWdlbnRzLnYxLkFnZW50T3BlcmF0aW9uUglv'
    'cGVyYXRpb24SKwoRdmFsaWRhdGlvbl9lcnJvcnMYAyADKAlSEHZhbGlkYXRpb25FcnJvcnM=');

@$core.Deprecated('Use getAgentOperationRequestDescriptor instead')
const GetAgentOperationRequest$json = {
  '1': 'GetAgentOperationRequest',
  '2': [
    {'1': 'operation_id', '3': 1, '4': 1, '5': 9, '10': 'operationId'},
  ],
};

/// Descriptor for `GetAgentOperationRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getAgentOperationRequestDescriptor =
    $convert.base64Decode(
        'ChhHZXRBZ2VudE9wZXJhdGlvblJlcXVlc3QSIQoMb3BlcmF0aW9uX2lkGAEgASgJUgtvcGVyYX'
        'Rpb25JZA==');

@$core.Deprecated('Use getAgentOperationResponseDescriptor instead')
const GetAgentOperationResponse$json = {
  '1': 'GetAgentOperationResponse',
  '2': [
    {
      '1': 'operation',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.AgentOperation',
      '10': 'operation'
    },
  ],
};

/// Descriptor for `GetAgentOperationResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getAgentOperationResponseDescriptor =
    $convert.base64Decode(
        'ChlHZXRBZ2VudE9wZXJhdGlvblJlc3BvbnNlEjcKCW9wZXJhdGlvbhgBIAEoCzIZLmFnZW50cy'
        '52MS5BZ2VudE9wZXJhdGlvblIJb3BlcmF0aW9u');

@$core.Deprecated('Use listAgentOperationsRequestDescriptor instead')
const ListAgentOperationsRequest$json = {
  '1': 'ListAgentOperationsRequest',
  '2': [
    {
      '1': 'status',
      '3': 1,
      '4': 1,
      '5': 14,
      '6': '.agents.v1.AgentOperationStatus',
      '10': 'status'
    },
    {'1': 'page_size', '3': 2, '4': 1, '5': 5, '10': 'pageSize'},
    {'1': 'page_token', '3': 3, '4': 1, '5': 9, '10': 'pageToken'},
  ],
};

/// Descriptor for `ListAgentOperationsRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listAgentOperationsRequestDescriptor =
    $convert.base64Decode(
        'ChpMaXN0QWdlbnRPcGVyYXRpb25zUmVxdWVzdBI3CgZzdGF0dXMYASABKA4yHy5hZ2VudHMudj'
        'EuQWdlbnRPcGVyYXRpb25TdGF0dXNSBnN0YXR1cxIbCglwYWdlX3NpemUYAiABKAVSCHBhZ2VT'
        'aXplEh0KCnBhZ2VfdG9rZW4YAyABKAlSCXBhZ2VUb2tlbg==');

@$core.Deprecated('Use listAgentOperationsResponseDescriptor instead')
const ListAgentOperationsResponse$json = {
  '1': 'ListAgentOperationsResponse',
  '2': [
    {
      '1': 'operations',
      '3': 1,
      '4': 3,
      '5': 11,
      '6': '.agents.v1.AgentOperation',
      '10': 'operations'
    },
    {'1': 'next_page_token', '3': 2, '4': 1, '5': 9, '10': 'nextPageToken'},
  ],
};

/// Descriptor for `ListAgentOperationsResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listAgentOperationsResponseDescriptor =
    $convert.base64Decode(
        'ChtMaXN0QWdlbnRPcGVyYXRpb25zUmVzcG9uc2USOQoKb3BlcmF0aW9ucxgBIAMoCzIZLmFnZW'
        '50cy52MS5BZ2VudE9wZXJhdGlvblIKb3BlcmF0aW9ucxImCg9uZXh0X3BhZ2VfdG9rZW4YAiAB'
        'KAlSDW5leHRQYWdlVG9rZW4=');

@$core.Deprecated('Use retryAgentOperationRequestDescriptor instead')
const RetryAgentOperationRequest$json = {
  '1': 'RetryAgentOperationRequest',
  '2': [
    {'1': 'operation_id', '3': 1, '4': 1, '5': 9, '10': 'operationId'},
  ],
};

/// Descriptor for `RetryAgentOperationRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List retryAgentOperationRequestDescriptor =
    $convert.base64Decode(
        'ChpSZXRyeUFnZW50T3BlcmF0aW9uUmVxdWVzdBIhCgxvcGVyYXRpb25faWQYASABKAlSC29wZX'
        'JhdGlvbklk');

@$core.Deprecated('Use retryAgentOperationResponseDescriptor instead')
const RetryAgentOperationResponse$json = {
  '1': 'RetryAgentOperationResponse',
  '2': [
    {
      '1': 'agent',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.Agent',
      '10': 'agent'
    },
    {
      '1': 'operation',
      '3': 2,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.AgentOperation',
      '10': 'operation'
    },
    {
      '1': 'validation_errors',
      '3': 3,
      '4': 3,
      '5': 9,
      '10': 'validationErrors'
    },
  ],
};

/// Descriptor for `RetryAgentOperationResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List retryAgentOperationResponseDescriptor = $convert.base64Decode(
    'ChtSZXRyeUFnZW50T3BlcmF0aW9uUmVzcG9uc2USJgoFYWdlbnQYASABKAsyEC5hZ2VudHMudj'
    'EuQWdlbnRSBWFnZW50EjcKCW9wZXJhdGlvbhgCIAEoCzIZLmFnZW50cy52MS5BZ2VudE9wZXJh'
    'dGlvblIJb3BlcmF0aW9uEisKEXZhbGlkYXRpb25fZXJyb3JzGAMgAygJUhB2YWxpZGF0aW9uRX'
    'Jyb3Jz');

@$core.Deprecated('Use updateAgentRequestDescriptor instead')
const UpdateAgentRequest$json = {
  '1': 'UpdateAgentRequest',
  '2': [
    {
      '1': 'agent',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.Agent',
      '10': 'agent'
    },
  ],
};

/// Descriptor for `UpdateAgentRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List updateAgentRequestDescriptor = $convert.base64Decode(
    'ChJVcGRhdGVBZ2VudFJlcXVlc3QSJgoFYWdlbnQYASABKAsyEC5hZ2VudHMudjEuQWdlbnRSBW'
    'FnZW50');

@$core.Deprecated('Use updateAgentResponseDescriptor instead')
const UpdateAgentResponse$json = {
  '1': 'UpdateAgentResponse',
  '2': [
    {
      '1': 'agent',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.Agent',
      '10': 'agent'
    },
  ],
};

/// Descriptor for `UpdateAgentResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List updateAgentResponseDescriptor = $convert.base64Decode(
    'ChNVcGRhdGVBZ2VudFJlc3BvbnNlEiYKBWFnZW50GAEgASgLMhAuYWdlbnRzLnYxLkFnZW50Ug'
    'VhZ2VudA==');

@$core.Deprecated('Use deleteAgentRequestDescriptor instead')
const DeleteAgentRequest$json = {
  '1': 'DeleteAgentRequest',
  '2': [
    {
      '1': 'name',
      '3': 1,
      '4': 1,
      '5': 9,
      '8': {'3': true},
      '10': 'name',
    },
    {'1': 'agent_id', '3': 2, '4': 1, '5': 9, '10': 'agentId'},
    {'1': 'operation_id', '3': 3, '4': 1, '5': 9, '10': 'operationId'},
  ],
};

/// Descriptor for `DeleteAgentRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List deleteAgentRequestDescriptor = $convert.base64Decode(
    'ChJEZWxldGVBZ2VudFJlcXVlc3QSFgoEbmFtZRgBIAEoCUICGAFSBG5hbWUSGQoIYWdlbnRfaW'
    'QYAiABKAlSB2FnZW50SWQSIQoMb3BlcmF0aW9uX2lkGAMgASgJUgtvcGVyYXRpb25JZA==');

@$core.Deprecated('Use deleteAgentResponseDescriptor instead')
const DeleteAgentResponse$json = {
  '1': 'DeleteAgentResponse',
  '2': [
    {
      '1': 'operation',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.AgentOperation',
      '10': 'operation'
    },
  ],
};

/// Descriptor for `DeleteAgentResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List deleteAgentResponseDescriptor = $convert.base64Decode(
    'ChNEZWxldGVBZ2VudFJlc3BvbnNlEjcKCW9wZXJhdGlvbhgBIAEoCzIZLmFnZW50cy52MS5BZ2'
    'VudE9wZXJhdGlvblIJb3BlcmF0aW9u');

@$core.Deprecated('Use assignAgentIDRequestDescriptor instead')
const AssignAgentIDRequest$json = {
  '1': 'AssignAgentIDRequest',
  '2': [
    {'1': 'name', '3': 1, '4': 1, '5': 9, '10': 'name'},
    {'1': 'agent_id', '3': 2, '4': 1, '5': 9, '10': 'agentId'},
  ],
  '7': {'3': true},
};

/// Descriptor for `AssignAgentIDRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List assignAgentIDRequestDescriptor = $convert.base64Decode(
    'ChRBc3NpZ25BZ2VudElEUmVxdWVzdBISCgRuYW1lGAEgASgJUgRuYW1lEhkKCGFnZW50X2lkGA'
    'IgASgJUgdhZ2VudElkOgIYAQ==');

@$core.Deprecated('Use assignAgentIDResponseDescriptor instead')
const AssignAgentIDResponse$json = {
  '1': 'AssignAgentIDResponse',
  '2': [
    {
      '1': 'agent',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.Agent',
      '10': 'agent'
    },
  ],
};

/// Descriptor for `AssignAgentIDResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List assignAgentIDResponseDescriptor = $convert.base64Decode(
    'ChVBc3NpZ25BZ2VudElEUmVzcG9uc2USJgoFYWdlbnQYASABKAsyEC5hZ2VudHMudjEuQWdlbn'
    'RSBWFnZW50');

@$core.Deprecated('Use getMigrationReadinessRequestDescriptor instead')
const GetMigrationReadinessRequest$json = {
  '1': 'GetMigrationReadinessRequest',
};

/// Descriptor for `GetMigrationReadinessRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getMigrationReadinessRequestDescriptor =
    $convert.base64Decode('ChxHZXRNaWdyYXRpb25SZWFkaW5lc3NSZXF1ZXN0');

@$core.Deprecated('Use getMigrationReadinessResponseDescriptor instead')
const GetMigrationReadinessResponse$json = {
  '1': 'GetMigrationReadinessResponse',
  '2': [
    {
      '1': 'statuses',
      '3': 1,
      '4': 3,
      '5': 11,
      '6': '.agents.v1.AgentMigrationStatus',
      '10': 'statuses'
    },
  ],
};

/// Descriptor for `GetMigrationReadinessResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getMigrationReadinessResponseDescriptor =
    $convert.base64Decode(
        'Ch1HZXRNaWdyYXRpb25SZWFkaW5lc3NSZXNwb25zZRI7CghzdGF0dXNlcxgBIAMoCzIfLmFnZW'
        '50cy52MS5BZ2VudE1pZ3JhdGlvblN0YXR1c1IIc3RhdHVzZXM=');

@$core.Deprecated('Use agentMigrationStatusDescriptor instead')
const AgentMigrationStatus$json = {
  '1': 'AgentMigrationStatus',
  '2': [
    {'1': 'name', '3': 1, '4': 1, '5': 9, '10': 'name'},
    {'1': 'agent_id', '3': 2, '4': 1, '5': 9, '10': 'agentId'},
    {
      '1': 'readiness',
      '3': 3,
      '4': 1,
      '5': 14,
      '6': '.agents.v1.MigrationReadiness',
      '10': 'readiness'
    },
    {'1': 'detail', '3': 4, '4': 1, '5': 9, '10': 'detail'},
  ],
};

/// Descriptor for `AgentMigrationStatus`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List agentMigrationStatusDescriptor = $convert.base64Decode(
    'ChRBZ2VudE1pZ3JhdGlvblN0YXR1cxISCgRuYW1lGAEgASgJUgRuYW1lEhkKCGFnZW50X2lkGA'
    'IgASgJUgdhZ2VudElkEjsKCXJlYWRpbmVzcxgDIAEoDjIdLmFnZW50cy52MS5NaWdyYXRpb25S'
    'ZWFkaW5lc3NSCXJlYWRpbmVzcxIWCgZkZXRhaWwYBCABKAlSBmRldGFpbA==');

@$core.Deprecated('Use migrateAgentsV2RequestDescriptor instead')
const MigrateAgentsV2Request$json = {
  '1': 'MigrateAgentsV2Request',
  '2': [
    {
      '1': 'mode',
      '3': 1,
      '4': 1,
      '5': 14,
      '6': '.agents.v1.MigrateMode',
      '10': 'mode'
    },
  ],
};

/// Descriptor for `MigrateAgentsV2Request`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List migrateAgentsV2RequestDescriptor =
    $convert.base64Decode(
        'ChZNaWdyYXRlQWdlbnRzVjJSZXF1ZXN0EioKBG1vZGUYASABKA4yFi5hZ2VudHMudjEuTWlncm'
        'F0ZU1vZGVSBG1vZGU=');

@$core.Deprecated('Use migrateAgentResultDescriptor instead')
const MigrateAgentResult$json = {
  '1': 'MigrateAgentResult',
  '2': [
    {'1': 'name', '3': 1, '4': 1, '5': 9, '10': 'name'},
    {'1': 'agent_id', '3': 2, '4': 1, '5': 9, '10': 'agentId'},
    {'1': 'action', '3': 3, '4': 1, '5': 9, '10': 'action'},
    {'1': 'detail', '3': 4, '4': 1, '5': 9, '10': 'detail'},
  ],
};

/// Descriptor for `MigrateAgentResult`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List migrateAgentResultDescriptor = $convert.base64Decode(
    'ChJNaWdyYXRlQWdlbnRSZXN1bHQSEgoEbmFtZRgBIAEoCVIEbmFtZRIZCghhZ2VudF9pZBgCIA'
    'EoCVIHYWdlbnRJZBIWCgZhY3Rpb24YAyABKAlSBmFjdGlvbhIWCgZkZXRhaWwYBCABKAlSBmRl'
    'dGFpbA==');

@$core.Deprecated('Use migrateAgentsV2ResponseDescriptor instead')
const MigrateAgentsV2Response$json = {
  '1': 'MigrateAgentsV2Response',
  '2': [
    {
      '1': 'mode',
      '3': 1,
      '4': 1,
      '5': 14,
      '6': '.agents.v1.MigrateMode',
      '10': 'mode'
    },
    {
      '1': 'results',
      '3': 2,
      '4': 3,
      '5': 11,
      '6': '.agents.v1.MigrateAgentResult',
      '10': 'results'
    },
    {'1': 'total', '3': 3, '4': 1, '5': 5, '10': 'total'},
    {'1': 'migrated', '3': 4, '4': 1, '5': 5, '10': 'migrated'},
    {'1': 'skipped', '3': 5, '4': 1, '5': 5, '10': 'skipped'},
    {'1': 'errors', '3': 6, '4': 1, '5': 5, '10': 'errors'},
  ],
};

/// Descriptor for `MigrateAgentsV2Response`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List migrateAgentsV2ResponseDescriptor = $convert.base64Decode(
    'ChdNaWdyYXRlQWdlbnRzVjJSZXNwb25zZRIqCgRtb2RlGAEgASgOMhYuYWdlbnRzLnYxLk1pZ3'
    'JhdGVNb2RlUgRtb2RlEjcKB3Jlc3VsdHMYAiADKAsyHS5hZ2VudHMudjEuTWlncmF0ZUFnZW50'
    'UmVzdWx0UgdyZXN1bHRzEhQKBXRvdGFsGAMgASgFUgV0b3RhbBIaCghtaWdyYXRlZBgEIAEoBV'
    'IIbWlncmF0ZWQSGAoHc2tpcHBlZBgFIAEoBVIHc2tpcHBlZBIWCgZlcnJvcnMYBiABKAVSBmVy'
    'cm9ycw==');

@$core.Deprecated('Use verifyAgentIDCutoverRequestDescriptor instead')
const VerifyAgentIDCutoverRequest$json = {
  '1': 'VerifyAgentIDCutoverRequest',
};

/// Descriptor for `VerifyAgentIDCutoverRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List verifyAgentIDCutoverRequestDescriptor =
    $convert.base64Decode('ChtWZXJpZnlBZ2VudElEQ3V0b3ZlclJlcXVlc3Q=');

@$core.Deprecated('Use agentIDCutoverFindingDescriptor instead')
const AgentIDCutoverFinding$json = {
  '1': 'AgentIDCutoverFinding',
  '2': [
    {'1': 'workspace_id', '3': 1, '4': 1, '5': 9, '10': 'workspaceId'},
    {'1': 'check', '3': 2, '4': 1, '5': 9, '10': 'check'},
    {'1': 'entity_kind', '3': 3, '4': 1, '5': 9, '10': 'entityKind'},
    {'1': 'entity', '3': 4, '4': 1, '5': 9, '10': 'entity'},
    {'1': 'detail', '3': 5, '4': 1, '5': 9, '10': 'detail'},
  ],
};

/// Descriptor for `AgentIDCutoverFinding`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List agentIDCutoverFindingDescriptor = $convert.base64Decode(
    'ChVBZ2VudElEQ3V0b3ZlckZpbmRpbmcSIQoMd29ya3NwYWNlX2lkGAEgASgJUgt3b3Jrc3BhY2'
    'VJZBIUCgVjaGVjaxgCIAEoCVIFY2hlY2sSHwoLZW50aXR5X2tpbmQYAyABKAlSCmVudGl0eUtp'
    'bmQSFgoGZW50aXR5GAQgASgJUgZlbnRpdHkSFgoGZGV0YWlsGAUgASgJUgZkZXRhaWw=');

@$core.Deprecated('Use verifyAgentIDCutoverResponseDescriptor instead')
const VerifyAgentIDCutoverResponse$json = {
  '1': 'VerifyAgentIDCutoverResponse',
  '2': [
    {'1': 'passed', '3': 1, '4': 1, '5': 8, '10': 'passed'},
    {
      '1': 'findings',
      '3': 2,
      '4': 3,
      '5': 11,
      '6': '.agents.v1.AgentIDCutoverFinding',
      '10': 'findings'
    },
  ],
};

/// Descriptor for `VerifyAgentIDCutoverResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List verifyAgentIDCutoverResponseDescriptor =
    $convert.base64Decode(
        'ChxWZXJpZnlBZ2VudElEQ3V0b3ZlclJlc3BvbnNlEhYKBnBhc3NlZBgBIAEoCFIGcGFzc2VkEj'
        'wKCGZpbmRpbmdzGAIgAygLMiAuYWdlbnRzLnYxLkFnZW50SURDdXRvdmVyRmluZGluZ1IIZmlu'
        'ZGluZ3M=');

@$core.Deprecated('Use invokeAgentRequestDescriptor instead')
const InvokeAgentRequest$json = {
  '1': 'InvokeAgentRequest',
  '2': [
    {
      '1': 'agent_name',
      '3': 1,
      '4': 1,
      '5': 9,
      '8': {'3': true},
      '10': 'agentName',
    },
    {'1': 'input', '3': 2, '4': 1, '5': 9, '10': 'input'},
    {'1': 'app_name', '3': 3, '4': 1, '5': 9, '10': 'appName'},
    {'1': 'user_id', '3': 4, '4': 1, '5': 9, '10': 'userId'},
    {'1': 'session_id', '3': 5, '4': 1, '5': 9, '10': 'sessionId'},
    {'1': 'model_override', '3': 6, '4': 1, '5': 9, '10': 'modelOverride'},
    {'1': 'agent_id', '3': 7, '4': 1, '5': 9, '10': 'agentId'},
  ],
};

/// Descriptor for `InvokeAgentRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List invokeAgentRequestDescriptor = $convert.base64Decode(
    'ChJJbnZva2VBZ2VudFJlcXVlc3QSIQoKYWdlbnRfbmFtZRgBIAEoCUICGAFSCWFnZW50TmFtZR'
    'IUCgVpbnB1dBgCIAEoCVIFaW5wdXQSGQoIYXBwX25hbWUYAyABKAlSB2FwcE5hbWUSFwoHdXNl'
    'cl9pZBgEIAEoCVIGdXNlcklkEh0KCnNlc3Npb25faWQYBSABKAlSCXNlc3Npb25JZBIlCg5tb2'
    'RlbF9vdmVycmlkZRgGIAEoCVINbW9kZWxPdmVycmlkZRIZCghhZ2VudF9pZBgHIAEoCVIHYWdl'
    'bnRJZA==');

@$core.Deprecated('Use invokeAgentResponseDescriptor instead')
const InvokeAgentResponse$json = {
  '1': 'InvokeAgentResponse',
  '2': [
    {'1': 'session_id', '3': 1, '4': 1, '5': 9, '10': 'sessionId'},
    {'1': 'response', '3': 2, '4': 1, '5': 9, '10': 'response'},
  ],
};

/// Descriptor for `InvokeAgentResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List invokeAgentResponseDescriptor = $convert.base64Decode(
    'ChNJbnZva2VBZ2VudFJlc3BvbnNlEh0KCnNlc3Npb25faWQYASABKAlSCXNlc3Npb25JZBIaCg'
    'hyZXNwb25zZRgCIAEoCVIIcmVzcG9uc2U=');

@$core.Deprecated('Use invocationDescriptor instead')
const Invocation$json = {
  '1': 'Invocation',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
    {'1': 'agent_name', '3': 2, '4': 1, '5': 9, '10': 'agentName'},
    {'1': 'app_name', '3': 3, '4': 1, '5': 9, '10': 'appName'},
    {'1': 'user_id', '3': 4, '4': 1, '5': 9, '10': 'userId'},
    {'1': 'session_id', '3': 5, '4': 1, '5': 9, '10': 'sessionId'},
    {
      '1': 'status',
      '3': 6,
      '4': 1,
      '5': 14,
      '6': '.agents.v1.InvocationStatus',
      '10': 'status'
    },
    {'1': 'input', '3': 7, '4': 1, '5': 9, '10': 'input'},
    {'1': 'output', '3': 8, '4': 1, '5': 9, '10': 'output'},
    {'1': 'error', '3': 9, '4': 1, '5': 9, '10': 'error'},
    {
      '1': 'started_at',
      '3': 10,
      '4': 1,
      '5': 11,
      '6': '.google.protobuf.Timestamp',
      '10': 'startedAt'
    },
    {
      '1': 'finished_at',
      '3': 11,
      '4': 1,
      '5': 11,
      '6': '.google.protobuf.Timestamp',
      '10': 'finishedAt'
    },
    {'1': 'latency_ms', '3': 12, '4': 1, '5': 3, '10': 'latencyMs'},
    {'1': 'model_override', '3': 13, '4': 1, '5': 9, '10': 'modelOverride'},
    {'1': 'source', '3': 14, '4': 1, '5': 9, '10': 'source'},
    {'1': 'agent_id', '3': 15, '4': 1, '5': 9, '10': 'agentId'},
    {
      '1': 'agent_display_name',
      '3': 16,
      '4': 1,
      '5': 9,
      '10': 'agentDisplayName'
    },
    {'1': 'request_id', '3': 17, '4': 1, '5': 9, '10': 'requestId'},
    {'1': 'workspace_id', '3': 100, '4': 1, '5': 9, '10': 'workspaceId'},
  ],
};

/// Descriptor for `Invocation`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List invocationDescriptor = $convert.base64Decode(
    'CgpJbnZvY2F0aW9uEg4KAmlkGAEgASgJUgJpZBIdCgphZ2VudF9uYW1lGAIgASgJUglhZ2VudE'
    '5hbWUSGQoIYXBwX25hbWUYAyABKAlSB2FwcE5hbWUSFwoHdXNlcl9pZBgEIAEoCVIGdXNlcklk'
    'Eh0KCnNlc3Npb25faWQYBSABKAlSCXNlc3Npb25JZBIzCgZzdGF0dXMYBiABKA4yGy5hZ2VudH'
    'MudjEuSW52b2NhdGlvblN0YXR1c1IGc3RhdHVzEhQKBWlucHV0GAcgASgJUgVpbnB1dBIWCgZv'
    'dXRwdXQYCCABKAlSBm91dHB1dBIUCgVlcnJvchgJIAEoCVIFZXJyb3ISOQoKc3RhcnRlZF9hdB'
    'gKIAEoCzIaLmdvb2dsZS5wcm90b2J1Zi5UaW1lc3RhbXBSCXN0YXJ0ZWRBdBI7CgtmaW5pc2hl'
    'ZF9hdBgLIAEoCzIaLmdvb2dsZS5wcm90b2J1Zi5UaW1lc3RhbXBSCmZpbmlzaGVkQXQSHQoKbG'
    'F0ZW5jeV9tcxgMIAEoA1IJbGF0ZW5jeU1zEiUKDm1vZGVsX292ZXJyaWRlGA0gASgJUg1tb2Rl'
    'bE92ZXJyaWRlEhYKBnNvdXJjZRgOIAEoCVIGc291cmNlEhkKCGFnZW50X2lkGA8gASgJUgdhZ2'
    'VudElkEiwKEmFnZW50X2Rpc3BsYXlfbmFtZRgQIAEoCVIQYWdlbnREaXNwbGF5TmFtZRIdCgpy'
    'ZXF1ZXN0X2lkGBEgASgJUglyZXF1ZXN0SWQSIQoMd29ya3NwYWNlX2lkGGQgASgJUgt3b3Jrc3'
    'BhY2VJZA==');

@$core.Deprecated('Use listAgentInvocationsRequestDescriptor instead')
const ListAgentInvocationsRequest$json = {
  '1': 'ListAgentInvocationsRequest',
  '2': [
    {'1': 'agent_name', '3': 1, '4': 1, '5': 9, '10': 'agentName'},
    {'1': 'session_id', '3': 2, '4': 1, '5': 9, '10': 'sessionId'},
    {'1': 'page_size', '3': 3, '4': 1, '5': 5, '10': 'pageSize'},
    {'1': 'page_token', '3': 4, '4': 1, '5': 9, '10': 'pageToken'},
    {'1': 'agent_id', '3': 5, '4': 1, '5': 9, '10': 'agentId'},
  ],
};

/// Descriptor for `ListAgentInvocationsRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listAgentInvocationsRequestDescriptor = $convert.base64Decode(
    'ChtMaXN0QWdlbnRJbnZvY2F0aW9uc1JlcXVlc3QSHQoKYWdlbnRfbmFtZRgBIAEoCVIJYWdlbn'
    'ROYW1lEh0KCnNlc3Npb25faWQYAiABKAlSCXNlc3Npb25JZBIbCglwYWdlX3NpemUYAyABKAVS'
    'CHBhZ2VTaXplEh0KCnBhZ2VfdG9rZW4YBCABKAlSCXBhZ2VUb2tlbhIZCghhZ2VudF9pZBgFIA'
    'EoCVIHYWdlbnRJZA==');

@$core.Deprecated('Use listAgentInvocationsResponseDescriptor instead')
const ListAgentInvocationsResponse$json = {
  '1': 'ListAgentInvocationsResponse',
  '2': [
    {
      '1': 'invocations',
      '3': 1,
      '4': 3,
      '5': 11,
      '6': '.agents.v1.Invocation',
      '10': 'invocations'
    },
    {'1': 'next_page_token', '3': 2, '4': 1, '5': 9, '10': 'nextPageToken'},
    {'1': 'total', '3': 3, '4': 1, '5': 5, '10': 'total'},
  ],
};

/// Descriptor for `ListAgentInvocationsResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listAgentInvocationsResponseDescriptor =
    $convert.base64Decode(
        'ChxMaXN0QWdlbnRJbnZvY2F0aW9uc1Jlc3BvbnNlEjcKC2ludm9jYXRpb25zGAEgAygLMhUuYW'
        'dlbnRzLnYxLkludm9jYXRpb25SC2ludm9jYXRpb25zEiYKD25leHRfcGFnZV90b2tlbhgCIAEo'
        'CVINbmV4dFBhZ2VUb2tlbhIUCgV0b3RhbBgDIAEoBVIFdG90YWw=');

@$core.Deprecated('Use getAgentInvocationRequestDescriptor instead')
const GetAgentInvocationRequest$json = {
  '1': 'GetAgentInvocationRequest',
  '2': [
    {'1': 'invocation_id', '3': 1, '4': 1, '5': 9, '10': 'invocationId'},
  ],
  '9': [
    {'1': 2, '2': 3},
    {'1': 3, '2': 4},
    {'1': 4, '2': 5},
  ],
  '10': ['session_id', 'latest', 'include_input_parts'],
};

/// Descriptor for `GetAgentInvocationRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getAgentInvocationRequestDescriptor = $convert.base64Decode(
    'ChlHZXRBZ2VudEludm9jYXRpb25SZXF1ZXN0EiMKDWludm9jYXRpb25faWQYASABKAlSDGludm'
    '9jYXRpb25JZEoECAIQA0oECAMQBEoECAQQBVIKc2Vzc2lvbl9pZFIGbGF0ZXN0UhNpbmNsdWRl'
    'X2lucHV0X3BhcnRz');

@$core.Deprecated('Use getAgentInvocationResponseDescriptor instead')
const GetAgentInvocationResponse$json = {
  '1': 'GetAgentInvocationResponse',
  '2': [
    {
      '1': 'invocation',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.Invocation',
      '10': 'invocation'
    },
  ],
  '9': [
    {'1': 2, '2': 3},
  ],
  '10': ['input_parts'],
};

/// Descriptor for `GetAgentInvocationResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getAgentInvocationResponseDescriptor =
    $convert.base64Decode(
        'ChpHZXRBZ2VudEludm9jYXRpb25SZXNwb25zZRI1CgppbnZvY2F0aW9uGAEgASgLMhUuYWdlbn'
        'RzLnYxLkludm9jYXRpb25SCmludm9jYXRpb25KBAgCEANSC2lucHV0X3BhcnRz');

@$core.Deprecated('Use listMCPServersRequestDescriptor instead')
const ListMCPServersRequest$json = {
  '1': 'ListMCPServersRequest',
};

/// Descriptor for `ListMCPServersRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listMCPServersRequestDescriptor =
    $convert.base64Decode('ChVMaXN0TUNQU2VydmVyc1JlcXVlc3Q=');

@$core.Deprecated('Use listMCPServersResponseDescriptor instead')
const ListMCPServersResponse$json = {
  '1': 'ListMCPServersResponse',
  '2': [
    {
      '1': 'mcp_servers',
      '3': 1,
      '4': 3,
      '5': 11,
      '6': '.agents.v1.MCPServer',
      '10': 'mcpServers'
    },
  ],
};

/// Descriptor for `ListMCPServersResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listMCPServersResponseDescriptor =
    $convert.base64Decode(
        'ChZMaXN0TUNQU2VydmVyc1Jlc3BvbnNlEjUKC21jcF9zZXJ2ZXJzGAEgAygLMhQuYWdlbnRzLn'
        'YxLk1DUFNlcnZlclIKbWNwU2VydmVycw==');

@$core.Deprecated('Use getMCPServerRequestDescriptor instead')
const GetMCPServerRequest$json = {
  '1': 'GetMCPServerRequest',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
  ],
};

/// Descriptor for `GetMCPServerRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getMCPServerRequestDescriptor = $convert
    .base64Decode('ChNHZXRNQ1BTZXJ2ZXJSZXF1ZXN0Eg4KAmlkGAEgASgJUgJpZA==');

@$core.Deprecated('Use getMCPServerResponseDescriptor instead')
const GetMCPServerResponse$json = {
  '1': 'GetMCPServerResponse',
  '2': [
    {
      '1': 'mcp_server',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.MCPServer',
      '10': 'mcpServer'
    },
  ],
};

/// Descriptor for `GetMCPServerResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getMCPServerResponseDescriptor = $convert.base64Decode(
    'ChRHZXRNQ1BTZXJ2ZXJSZXNwb25zZRIzCgptY3Bfc2VydmVyGAEgASgLMhQuYWdlbnRzLnYxLk'
    '1DUFNlcnZlclIJbWNwU2VydmVy');

@$core.Deprecated('Use createMCPServerRequestDescriptor instead')
const CreateMCPServerRequest$json = {
  '1': 'CreateMCPServerRequest',
  '2': [
    {
      '1': 'mcp_server',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.MCPServer',
      '10': 'mcpServer'
    },
  ],
};

/// Descriptor for `CreateMCPServerRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List createMCPServerRequestDescriptor =
    $convert.base64Decode(
        'ChZDcmVhdGVNQ1BTZXJ2ZXJSZXF1ZXN0EjMKCm1jcF9zZXJ2ZXIYASABKAsyFC5hZ2VudHMudj'
        'EuTUNQU2VydmVyUgltY3BTZXJ2ZXI=');

@$core.Deprecated('Use createMCPServerResponseDescriptor instead')
const CreateMCPServerResponse$json = {
  '1': 'CreateMCPServerResponse',
  '2': [
    {
      '1': 'mcp_server',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.MCPServer',
      '10': 'mcpServer'
    },
  ],
};

/// Descriptor for `CreateMCPServerResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List createMCPServerResponseDescriptor =
    $convert.base64Decode(
        'ChdDcmVhdGVNQ1BTZXJ2ZXJSZXNwb25zZRIzCgptY3Bfc2VydmVyGAEgASgLMhQuYWdlbnRzLn'
        'YxLk1DUFNlcnZlclIJbWNwU2VydmVy');

@$core.Deprecated('Use updateMCPServerRequestDescriptor instead')
const UpdateMCPServerRequest$json = {
  '1': 'UpdateMCPServerRequest',
  '2': [
    {
      '1': 'mcp_server',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.MCPServer',
      '10': 'mcpServer'
    },
  ],
};

/// Descriptor for `UpdateMCPServerRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List updateMCPServerRequestDescriptor =
    $convert.base64Decode(
        'ChZVcGRhdGVNQ1BTZXJ2ZXJSZXF1ZXN0EjMKCm1jcF9zZXJ2ZXIYASABKAsyFC5hZ2VudHMudj'
        'EuTUNQU2VydmVyUgltY3BTZXJ2ZXI=');

@$core.Deprecated('Use updateMCPServerResponseDescriptor instead')
const UpdateMCPServerResponse$json = {
  '1': 'UpdateMCPServerResponse',
  '2': [
    {
      '1': 'mcp_server',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.MCPServer',
      '10': 'mcpServer'
    },
  ],
};

/// Descriptor for `UpdateMCPServerResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List updateMCPServerResponseDescriptor =
    $convert.base64Decode(
        'ChdVcGRhdGVNQ1BTZXJ2ZXJSZXNwb25zZRIzCgptY3Bfc2VydmVyGAEgASgLMhQuYWdlbnRzLn'
        'YxLk1DUFNlcnZlclIJbWNwU2VydmVy');

@$core.Deprecated(
    'Use updateMCPServerKeepingCredentialsRequestDescriptor instead')
const UpdateMCPServerKeepingCredentialsRequest$json = {
  '1': 'UpdateMCPServerKeepingCredentialsRequest',
  '2': [
    {
      '1': 'mcp_server',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.MCPServer',
      '10': 'mcpServer'
    },
    {'1': 'keep_header_keys', '3': 2, '4': 3, '5': 9, '10': 'keepHeaderKeys'},
    {
      '1': 'keep_oauth_client_secret',
      '3': 3,
      '4': 1,
      '5': 8,
      '10': 'keepOauthClientSecret'
    },
  ],
};

/// Descriptor for `UpdateMCPServerKeepingCredentialsRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List updateMCPServerKeepingCredentialsRequestDescriptor =
    $convert.base64Decode(
        'CihVcGRhdGVNQ1BTZXJ2ZXJLZWVwaW5nQ3JlZGVudGlhbHNSZXF1ZXN0EjMKCm1jcF9zZXJ2ZX'
        'IYASABKAsyFC5hZ2VudHMudjEuTUNQU2VydmVyUgltY3BTZXJ2ZXISKAoQa2VlcF9oZWFkZXJf'
        'a2V5cxgCIAMoCVIOa2VlcEhlYWRlcktleXMSNwoYa2VlcF9vYXV0aF9jbGllbnRfc2VjcmV0GA'
        'MgASgIUhVrZWVwT2F1dGhDbGllbnRTZWNyZXQ=');

@$core.Deprecated(
    'Use updateMCPServerKeepingCredentialsResponseDescriptor instead')
const UpdateMCPServerKeepingCredentialsResponse$json = {
  '1': 'UpdateMCPServerKeepingCredentialsResponse',
  '2': [
    {
      '1': 'mcp_server',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.MCPServer',
      '10': 'mcpServer'
    },
  ],
};

/// Descriptor for `UpdateMCPServerKeepingCredentialsResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List
    updateMCPServerKeepingCredentialsResponseDescriptor = $convert.base64Decode(
        'CilVcGRhdGVNQ1BTZXJ2ZXJLZWVwaW5nQ3JlZGVudGlhbHNSZXNwb25zZRIzCgptY3Bfc2Vydm'
        'VyGAEgASgLMhQuYWdlbnRzLnYxLk1DUFNlcnZlclIJbWNwU2VydmVy');

@$core.Deprecated('Use deleteMCPServerRequestDescriptor instead')
const DeleteMCPServerRequest$json = {
  '1': 'DeleteMCPServerRequest',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
  ],
};

/// Descriptor for `DeleteMCPServerRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List deleteMCPServerRequestDescriptor = $convert
    .base64Decode('ChZEZWxldGVNQ1BTZXJ2ZXJSZXF1ZXN0Eg4KAmlkGAEgASgJUgJpZA==');

@$core.Deprecated('Use deleteMCPServerResponseDescriptor instead')
const DeleteMCPServerResponse$json = {
  '1': 'DeleteMCPServerResponse',
};

/// Descriptor for `DeleteMCPServerResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List deleteMCPServerResponseDescriptor =
    $convert.base64Decode('ChdEZWxldGVNQ1BTZXJ2ZXJSZXNwb25zZQ==');

@$core.Deprecated('Use startMCPServerOAuthRequestDescriptor instead')
const StartMCPServerOAuthRequest$json = {
  '1': 'StartMCPServerOAuthRequest',
  '2': [
    {'1': 'server_id', '3': 1, '4': 1, '5': 9, '10': 'serverId'},
    {'1': 'return_url', '3': 2, '4': 1, '5': 9, '10': 'returnUrl'},
  ],
};

/// Descriptor for `StartMCPServerOAuthRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List startMCPServerOAuthRequestDescriptor =
    $convert.base64Decode(
        'ChpTdGFydE1DUFNlcnZlck9BdXRoUmVxdWVzdBIbCglzZXJ2ZXJfaWQYASABKAlSCHNlcnZlck'
        'lkEh0KCnJldHVybl91cmwYAiABKAlSCXJldHVyblVybA==');

@$core.Deprecated('Use startMCPServerOAuthResponseDescriptor instead')
const StartMCPServerOAuthResponse$json = {
  '1': 'StartMCPServerOAuthResponse',
  '2': [
    {
      '1': 'authorization_url',
      '3': 1,
      '4': 1,
      '5': 9,
      '10': 'authorizationUrl'
    },
    {'1': 'flow_id', '3': 2, '4': 1, '5': 9, '10': 'flowId'},
  ],
};

/// Descriptor for `StartMCPServerOAuthResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List startMCPServerOAuthResponseDescriptor =
    $convert.base64Decode(
        'ChtTdGFydE1DUFNlcnZlck9BdXRoUmVzcG9uc2USKwoRYXV0aG9yaXphdGlvbl91cmwYASABKA'
        'lSEGF1dGhvcml6YXRpb25VcmwSFwoHZmxvd19pZBgCIAEoCVIGZmxvd0lk');

@$core.Deprecated('Use completeMCPServerOAuthRequestDescriptor instead')
const CompleteMCPServerOAuthRequest$json = {
  '1': 'CompleteMCPServerOAuthRequest',
  '2': [
    {'1': 'flow_id', '3': 1, '4': 1, '5': 9, '10': 'flowId'},
    {'1': 'code', '3': 2, '4': 1, '5': 9, '10': 'code'},
    {'1': 'state', '3': 3, '4': 1, '5': 9, '10': 'state'},
  ],
};

/// Descriptor for `CompleteMCPServerOAuthRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List completeMCPServerOAuthRequestDescriptor =
    $convert.base64Decode(
        'Ch1Db21wbGV0ZU1DUFNlcnZlck9BdXRoUmVxdWVzdBIXCgdmbG93X2lkGAEgASgJUgZmbG93SW'
        'QSEgoEY29kZRgCIAEoCVIEY29kZRIUCgVzdGF0ZRgDIAEoCVIFc3RhdGU=');

@$core.Deprecated('Use completeMCPServerOAuthResponseDescriptor instead')
const CompleteMCPServerOAuthResponse$json = {
  '1': 'CompleteMCPServerOAuthResponse',
  '2': [
    {
      '1': 'status',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.MCPOAuthConnectionStatus',
      '10': 'status'
    },
  ],
};

/// Descriptor for `CompleteMCPServerOAuthResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List completeMCPServerOAuthResponseDescriptor =
    $convert.base64Decode(
        'Ch5Db21wbGV0ZU1DUFNlcnZlck9BdXRoUmVzcG9uc2USOwoGc3RhdHVzGAEgASgLMiMuYWdlbn'
        'RzLnYxLk1DUE9BdXRoQ29ubmVjdGlvblN0YXR1c1IGc3RhdHVz');

@$core.Deprecated('Use getMCPServerOAuthStatusRequestDescriptor instead')
const GetMCPServerOAuthStatusRequest$json = {
  '1': 'GetMCPServerOAuthStatusRequest',
  '2': [
    {'1': 'server_id', '3': 1, '4': 1, '5': 9, '10': 'serverId'},
  ],
};

/// Descriptor for `GetMCPServerOAuthStatusRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getMCPServerOAuthStatusRequestDescriptor =
    $convert.base64Decode(
        'Ch5HZXRNQ1BTZXJ2ZXJPQXV0aFN0YXR1c1JlcXVlc3QSGwoJc2VydmVyX2lkGAEgASgJUghzZX'
        'J2ZXJJZA==');

@$core.Deprecated('Use getMCPServerOAuthStatusResponseDescriptor instead')
const GetMCPServerOAuthStatusResponse$json = {
  '1': 'GetMCPServerOAuthStatusResponse',
  '2': [
    {
      '1': 'status',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.MCPOAuthConnectionStatus',
      '10': 'status'
    },
  ],
};

/// Descriptor for `GetMCPServerOAuthStatusResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getMCPServerOAuthStatusResponseDescriptor =
    $convert.base64Decode(
        'Ch9HZXRNQ1BTZXJ2ZXJPQXV0aFN0YXR1c1Jlc3BvbnNlEjsKBnN0YXR1cxgBIAEoCzIjLmFnZW'
        '50cy52MS5NQ1BPQXV0aENvbm5lY3Rpb25TdGF0dXNSBnN0YXR1cw==');

@$core.Deprecated('Use disconnectMCPServerOAuthRequestDescriptor instead')
const DisconnectMCPServerOAuthRequest$json = {
  '1': 'DisconnectMCPServerOAuthRequest',
  '2': [
    {'1': 'server_id', '3': 1, '4': 1, '5': 9, '10': 'serverId'},
  ],
};

/// Descriptor for `DisconnectMCPServerOAuthRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List disconnectMCPServerOAuthRequestDescriptor =
    $convert.base64Decode(
        'Ch9EaXNjb25uZWN0TUNQU2VydmVyT0F1dGhSZXF1ZXN0EhsKCXNlcnZlcl9pZBgBIAEoCVIIc2'
        'VydmVySWQ=');

@$core.Deprecated('Use disconnectMCPServerOAuthResponseDescriptor instead')
const DisconnectMCPServerOAuthResponse$json = {
  '1': 'DisconnectMCPServerOAuthResponse',
  '2': [
    {
      '1': 'status',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.MCPOAuthConnectionStatus',
      '10': 'status'
    },
  ],
};

/// Descriptor for `DisconnectMCPServerOAuthResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List disconnectMCPServerOAuthResponseDescriptor =
    $convert.base64Decode(
        'CiBEaXNjb25uZWN0TUNQU2VydmVyT0F1dGhSZXNwb25zZRI7CgZzdGF0dXMYASABKAsyIy5hZ2'
        'VudHMudjEuTUNQT0F1dGhDb25uZWN0aW9uU3RhdHVzUgZzdGF0dXM=');

@$core.Deprecated('Use mCPOAuthConnectionStatusDescriptor instead')
const MCPOAuthConnectionStatus$json = {
  '1': 'MCPOAuthConnectionStatus',
  '2': [
    {'1': 'server_id', '3': 1, '4': 1, '5': 9, '10': 'serverId'},
    {
      '1': 'state',
      '3': 2,
      '4': 1,
      '5': 14,
      '6': '.agents.v1.MCPOAuthConnectionState',
      '10': 'state'
    },
    {'1': 'detail', '3': 3, '4': 1, '5': 9, '10': 'detail'},
    {'1': 'scopes', '3': 4, '4': 3, '5': 9, '10': 'scopes'},
    {
      '1': 'connected_at',
      '3': 5,
      '4': 1,
      '5': 11,
      '6': '.google.protobuf.Timestamp',
      '10': 'connectedAt'
    },
    {
      '1': 'expires_at',
      '3': 6,
      '4': 1,
      '5': 11,
      '6': '.google.protobuf.Timestamp',
      '10': 'expiresAt'
    },
    {
      '1': 'checked_at',
      '3': 7,
      '4': 1,
      '5': 11,
      '6': '.google.protobuf.Timestamp',
      '10': 'checkedAt'
    },
  ],
};

/// Descriptor for `MCPOAuthConnectionStatus`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List mCPOAuthConnectionStatusDescriptor = $convert.base64Decode(
    'ChhNQ1BPQXV0aENvbm5lY3Rpb25TdGF0dXMSGwoJc2VydmVyX2lkGAEgASgJUghzZXJ2ZXJJZB'
    'I4CgVzdGF0ZRgCIAEoDjIiLmFnZW50cy52MS5NQ1BPQXV0aENvbm5lY3Rpb25TdGF0ZVIFc3Rh'
    'dGUSFgoGZGV0YWlsGAMgASgJUgZkZXRhaWwSFgoGc2NvcGVzGAQgAygJUgZzY29wZXMSPQoMY2'
    '9ubmVjdGVkX2F0GAUgASgLMhouZ29vZ2xlLnByb3RvYnVmLlRpbWVzdGFtcFILY29ubmVjdGVk'
    'QXQSOQoKZXhwaXJlc19hdBgGIAEoCzIaLmdvb2dsZS5wcm90b2J1Zi5UaW1lc3RhbXBSCWV4cG'
    'lyZXNBdBI5CgpjaGVja2VkX2F0GAcgASgLMhouZ29vZ2xlLnByb3RvYnVmLlRpbWVzdGFtcFIJ'
    'Y2hlY2tlZEF0');

@$core.Deprecated('Use listGlobalMCPServersRequestDescriptor instead')
const ListGlobalMCPServersRequest$json = {
  '1': 'ListGlobalMCPServersRequest',
};

/// Descriptor for `ListGlobalMCPServersRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listGlobalMCPServersRequestDescriptor =
    $convert.base64Decode('ChtMaXN0R2xvYmFsTUNQU2VydmVyc1JlcXVlc3Q=');

@$core.Deprecated('Use listGlobalMCPServersResponseDescriptor instead')
const ListGlobalMCPServersResponse$json = {
  '1': 'ListGlobalMCPServersResponse',
  '2': [
    {
      '1': 'mcp_servers',
      '3': 1,
      '4': 3,
      '5': 11,
      '6': '.agents.v1.MCPServer',
      '10': 'mcpServers'
    },
  ],
};

/// Descriptor for `ListGlobalMCPServersResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listGlobalMCPServersResponseDescriptor =
    $convert.base64Decode(
        'ChxMaXN0R2xvYmFsTUNQU2VydmVyc1Jlc3BvbnNlEjUKC21jcF9zZXJ2ZXJzGAEgAygLMhQuYW'
        'dlbnRzLnYxLk1DUFNlcnZlclIKbWNwU2VydmVycw==');

@$core.Deprecated('Use createGlobalMCPServerRequestDescriptor instead')
const CreateGlobalMCPServerRequest$json = {
  '1': 'CreateGlobalMCPServerRequest',
  '2': [
    {
      '1': 'mcp_server',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.MCPServer',
      '10': 'mcpServer'
    },
  ],
};

/// Descriptor for `CreateGlobalMCPServerRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List createGlobalMCPServerRequestDescriptor =
    $convert.base64Decode(
        'ChxDcmVhdGVHbG9iYWxNQ1BTZXJ2ZXJSZXF1ZXN0EjMKCm1jcF9zZXJ2ZXIYASABKAsyFC5hZ2'
        'VudHMudjEuTUNQU2VydmVyUgltY3BTZXJ2ZXI=');

@$core.Deprecated('Use createGlobalMCPServerResponseDescriptor instead')
const CreateGlobalMCPServerResponse$json = {
  '1': 'CreateGlobalMCPServerResponse',
  '2': [
    {
      '1': 'mcp_server',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.MCPServer',
      '10': 'mcpServer'
    },
  ],
};

/// Descriptor for `CreateGlobalMCPServerResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List createGlobalMCPServerResponseDescriptor =
    $convert.base64Decode(
        'Ch1DcmVhdGVHbG9iYWxNQ1BTZXJ2ZXJSZXNwb25zZRIzCgptY3Bfc2VydmVyGAEgASgLMhQuYW'
        'dlbnRzLnYxLk1DUFNlcnZlclIJbWNwU2VydmVy');

@$core.Deprecated('Use updateGlobalMCPServerRequestDescriptor instead')
const UpdateGlobalMCPServerRequest$json = {
  '1': 'UpdateGlobalMCPServerRequest',
  '2': [
    {
      '1': 'mcp_server',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.MCPServer',
      '10': 'mcpServer'
    },
  ],
};

/// Descriptor for `UpdateGlobalMCPServerRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List updateGlobalMCPServerRequestDescriptor =
    $convert.base64Decode(
        'ChxVcGRhdGVHbG9iYWxNQ1BTZXJ2ZXJSZXF1ZXN0EjMKCm1jcF9zZXJ2ZXIYASABKAsyFC5hZ2'
        'VudHMudjEuTUNQU2VydmVyUgltY3BTZXJ2ZXI=');

@$core.Deprecated('Use updateGlobalMCPServerResponseDescriptor instead')
const UpdateGlobalMCPServerResponse$json = {
  '1': 'UpdateGlobalMCPServerResponse',
  '2': [
    {
      '1': 'mcp_server',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.MCPServer',
      '10': 'mcpServer'
    },
  ],
};

/// Descriptor for `UpdateGlobalMCPServerResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List updateGlobalMCPServerResponseDescriptor =
    $convert.base64Decode(
        'Ch1VcGRhdGVHbG9iYWxNQ1BTZXJ2ZXJSZXNwb25zZRIzCgptY3Bfc2VydmVyGAEgASgLMhQuYW'
        'dlbnRzLnYxLk1DUFNlcnZlclIJbWNwU2VydmVy');

@$core.Deprecated('Use deleteGlobalMCPServerRequestDescriptor instead')
const DeleteGlobalMCPServerRequest$json = {
  '1': 'DeleteGlobalMCPServerRequest',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
  ],
};

/// Descriptor for `DeleteGlobalMCPServerRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List deleteGlobalMCPServerRequestDescriptor =
    $convert.base64Decode(
        'ChxEZWxldGVHbG9iYWxNQ1BTZXJ2ZXJSZXF1ZXN0Eg4KAmlkGAEgASgJUgJpZA==');

@$core.Deprecated('Use deleteGlobalMCPServerResponseDescriptor instead')
const DeleteGlobalMCPServerResponse$json = {
  '1': 'DeleteGlobalMCPServerResponse',
};

/// Descriptor for `DeleteGlobalMCPServerResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List deleteGlobalMCPServerResponseDescriptor =
    $convert.base64Decode('Ch1EZWxldGVHbG9iYWxNQ1BTZXJ2ZXJSZXNwb25zZQ==');

@$core.Deprecated('Use installGlobalMCPServerRequestDescriptor instead')
const InstallGlobalMCPServerRequest$json = {
  '1': 'InstallGlobalMCPServerRequest',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
    {'1': 'workspace_id', '3': 2, '4': 1, '5': 9, '10': 'workspaceId'},
  ],
};

/// Descriptor for `InstallGlobalMCPServerRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List installGlobalMCPServerRequestDescriptor =
    $convert.base64Decode(
        'Ch1JbnN0YWxsR2xvYmFsTUNQU2VydmVyUmVxdWVzdBIOCgJpZBgBIAEoCVICaWQSIQoMd29ya3'
        'NwYWNlX2lkGAIgASgJUgt3b3Jrc3BhY2VJZA==');

@$core.Deprecated('Use installGlobalMCPServerResponseDescriptor instead')
const InstallGlobalMCPServerResponse$json = {
  '1': 'InstallGlobalMCPServerResponse',
  '2': [
    {
      '1': 'mcp_server',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.MCPServer',
      '10': 'mcpServer'
    },
  ],
};

/// Descriptor for `InstallGlobalMCPServerResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List installGlobalMCPServerResponseDescriptor =
    $convert.base64Decode(
        'Ch5JbnN0YWxsR2xvYmFsTUNQU2VydmVyUmVzcG9uc2USMwoKbWNwX3NlcnZlchgBIAEoCzIULm'
        'FnZW50cy52MS5NQ1BTZXJ2ZXJSCW1jcFNlcnZlcg==');

@$core.Deprecated('Use listModelProvidersRequestDescriptor instead')
const ListModelProvidersRequest$json = {
  '1': 'ListModelProvidersRequest',
};

/// Descriptor for `ListModelProvidersRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listModelProvidersRequestDescriptor =
    $convert.base64Decode('ChlMaXN0TW9kZWxQcm92aWRlcnNSZXF1ZXN0');

@$core.Deprecated('Use listModelProvidersResponseDescriptor instead')
const ListModelProvidersResponse$json = {
  '1': 'ListModelProvidersResponse',
  '2': [
    {
      '1': 'model_providers',
      '3': 1,
      '4': 3,
      '5': 11,
      '6': '.agents.v1.ModelProvider',
      '10': 'modelProviders'
    },
  ],
};

/// Descriptor for `ListModelProvidersResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listModelProvidersResponseDescriptor =
    $convert.base64Decode(
        'ChpMaXN0TW9kZWxQcm92aWRlcnNSZXNwb25zZRJBCg9tb2RlbF9wcm92aWRlcnMYASADKAsyGC'
        '5hZ2VudHMudjEuTW9kZWxQcm92aWRlclIObW9kZWxQcm92aWRlcnM=');

@$core.Deprecated('Use getModelProviderRequestDescriptor instead')
const GetModelProviderRequest$json = {
  '1': 'GetModelProviderRequest',
  '2': [
    {'1': 'name', '3': 1, '4': 1, '5': 9, '10': 'name'},
  ],
};

/// Descriptor for `GetModelProviderRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getModelProviderRequestDescriptor =
    $convert.base64Decode(
        'ChdHZXRNb2RlbFByb3ZpZGVyUmVxdWVzdBISCgRuYW1lGAEgASgJUgRuYW1l');

@$core.Deprecated('Use getModelProviderResponseDescriptor instead')
const GetModelProviderResponse$json = {
  '1': 'GetModelProviderResponse',
  '2': [
    {
      '1': 'model_provider',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.ModelProvider',
      '10': 'modelProvider'
    },
  ],
};

/// Descriptor for `GetModelProviderResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getModelProviderResponseDescriptor =
    $convert.base64Decode(
        'ChhHZXRNb2RlbFByb3ZpZGVyUmVzcG9uc2USPwoObW9kZWxfcHJvdmlkZXIYASABKAsyGC5hZ2'
        'VudHMudjEuTW9kZWxQcm92aWRlclINbW9kZWxQcm92aWRlcg==');

@$core.Deprecated('Use createModelProviderRequestDescriptor instead')
const CreateModelProviderRequest$json = {
  '1': 'CreateModelProviderRequest',
  '2': [
    {
      '1': 'model_provider',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.ModelProvider',
      '10': 'modelProvider'
    },
  ],
};

/// Descriptor for `CreateModelProviderRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List createModelProviderRequestDescriptor =
    $convert.base64Decode(
        'ChpDcmVhdGVNb2RlbFByb3ZpZGVyUmVxdWVzdBI/Cg5tb2RlbF9wcm92aWRlchgBIAEoCzIYLm'
        'FnZW50cy52MS5Nb2RlbFByb3ZpZGVyUg1tb2RlbFByb3ZpZGVy');

@$core.Deprecated('Use createModelProviderResponseDescriptor instead')
const CreateModelProviderResponse$json = {
  '1': 'CreateModelProviderResponse',
  '2': [
    {
      '1': 'model_provider',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.ModelProvider',
      '10': 'modelProvider'
    },
  ],
};

/// Descriptor for `CreateModelProviderResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List createModelProviderResponseDescriptor =
    $convert.base64Decode(
        'ChtDcmVhdGVNb2RlbFByb3ZpZGVyUmVzcG9uc2USPwoObW9kZWxfcHJvdmlkZXIYASABKAsyGC'
        '5hZ2VudHMudjEuTW9kZWxQcm92aWRlclINbW9kZWxQcm92aWRlcg==');

@$core.Deprecated('Use updateModelProviderRequestDescriptor instead')
const UpdateModelProviderRequest$json = {
  '1': 'UpdateModelProviderRequest',
  '2': [
    {
      '1': 'model_provider',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.ModelProvider',
      '10': 'modelProvider'
    },
  ],
};

/// Descriptor for `UpdateModelProviderRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List updateModelProviderRequestDescriptor =
    $convert.base64Decode(
        'ChpVcGRhdGVNb2RlbFByb3ZpZGVyUmVxdWVzdBI/Cg5tb2RlbF9wcm92aWRlchgBIAEoCzIYLm'
        'FnZW50cy52MS5Nb2RlbFByb3ZpZGVyUg1tb2RlbFByb3ZpZGVy');

@$core.Deprecated('Use updateModelProviderResponseDescriptor instead')
const UpdateModelProviderResponse$json = {
  '1': 'UpdateModelProviderResponse',
  '2': [
    {
      '1': 'model_provider',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.ModelProvider',
      '10': 'modelProvider'
    },
  ],
};

/// Descriptor for `UpdateModelProviderResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List updateModelProviderResponseDescriptor =
    $convert.base64Decode(
        'ChtVcGRhdGVNb2RlbFByb3ZpZGVyUmVzcG9uc2USPwoObW9kZWxfcHJvdmlkZXIYASABKAsyGC'
        '5hZ2VudHMudjEuTW9kZWxQcm92aWRlclINbW9kZWxQcm92aWRlcg==');

@$core.Deprecated('Use deleteModelProviderRequestDescriptor instead')
const DeleteModelProviderRequest$json = {
  '1': 'DeleteModelProviderRequest',
  '2': [
    {'1': 'name', '3': 1, '4': 1, '5': 9, '10': 'name'},
  ],
};

/// Descriptor for `DeleteModelProviderRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List deleteModelProviderRequestDescriptor =
    $convert.base64Decode(
        'ChpEZWxldGVNb2RlbFByb3ZpZGVyUmVxdWVzdBISCgRuYW1lGAEgASgJUgRuYW1l');

@$core.Deprecated('Use deleteModelProviderResponseDescriptor instead')
const DeleteModelProviderResponse$json = {
  '1': 'DeleteModelProviderResponse',
};

/// Descriptor for `DeleteModelProviderResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List deleteModelProviderResponseDescriptor =
    $convert.base64Decode('ChtEZWxldGVNb2RlbFByb3ZpZGVyUmVzcG9uc2U=');

@$core.Deprecated('Use listNotifyGroupsRequestDescriptor instead')
const ListNotifyGroupsRequest$json = {
  '1': 'ListNotifyGroupsRequest',
};

/// Descriptor for `ListNotifyGroupsRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listNotifyGroupsRequestDescriptor =
    $convert.base64Decode('ChdMaXN0Tm90aWZ5R3JvdXBzUmVxdWVzdA==');

@$core.Deprecated('Use listNotifyGroupsResponseDescriptor instead')
const ListNotifyGroupsResponse$json = {
  '1': 'ListNotifyGroupsResponse',
  '2': [
    {
      '1': 'notify_groups',
      '3': 1,
      '4': 3,
      '5': 11,
      '6': '.agents.v1.NotifyGroup',
      '10': 'notifyGroups'
    },
  ],
};

/// Descriptor for `ListNotifyGroupsResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listNotifyGroupsResponseDescriptor =
    $convert.base64Decode(
        'ChhMaXN0Tm90aWZ5R3JvdXBzUmVzcG9uc2USOwoNbm90aWZ5X2dyb3VwcxgBIAMoCzIWLmFnZW'
        '50cy52MS5Ob3RpZnlHcm91cFIMbm90aWZ5R3JvdXBz');

@$core.Deprecated('Use getNotifyGroupRequestDescriptor instead')
const GetNotifyGroupRequest$json = {
  '1': 'GetNotifyGroupRequest',
  '2': [
    {'1': 'name', '3': 1, '4': 1, '5': 9, '10': 'name'},
  ],
};

/// Descriptor for `GetNotifyGroupRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getNotifyGroupRequestDescriptor =
    $convert.base64Decode(
        'ChVHZXROb3RpZnlHcm91cFJlcXVlc3QSEgoEbmFtZRgBIAEoCVIEbmFtZQ==');

@$core.Deprecated('Use getNotifyGroupResponseDescriptor instead')
const GetNotifyGroupResponse$json = {
  '1': 'GetNotifyGroupResponse',
  '2': [
    {
      '1': 'notify_group',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.NotifyGroup',
      '10': 'notifyGroup'
    },
  ],
};

/// Descriptor for `GetNotifyGroupResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getNotifyGroupResponseDescriptor =
    $convert.base64Decode(
        'ChZHZXROb3RpZnlHcm91cFJlc3BvbnNlEjkKDG5vdGlmeV9ncm91cBgBIAEoCzIWLmFnZW50cy'
        '52MS5Ob3RpZnlHcm91cFILbm90aWZ5R3JvdXA=');

@$core.Deprecated('Use createNotifyGroupRequestDescriptor instead')
const CreateNotifyGroupRequest$json = {
  '1': 'CreateNotifyGroupRequest',
  '2': [
    {
      '1': 'notify_group',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.NotifyGroup',
      '10': 'notifyGroup'
    },
  ],
};

/// Descriptor for `CreateNotifyGroupRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List createNotifyGroupRequestDescriptor =
    $convert.base64Decode(
        'ChhDcmVhdGVOb3RpZnlHcm91cFJlcXVlc3QSOQoMbm90aWZ5X2dyb3VwGAEgASgLMhYuYWdlbn'
        'RzLnYxLk5vdGlmeUdyb3VwUgtub3RpZnlHcm91cA==');

@$core.Deprecated('Use createNotifyGroupResponseDescriptor instead')
const CreateNotifyGroupResponse$json = {
  '1': 'CreateNotifyGroupResponse',
  '2': [
    {
      '1': 'notify_group',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.NotifyGroup',
      '10': 'notifyGroup'
    },
  ],
};

/// Descriptor for `CreateNotifyGroupResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List createNotifyGroupResponseDescriptor =
    $convert.base64Decode(
        'ChlDcmVhdGVOb3RpZnlHcm91cFJlc3BvbnNlEjkKDG5vdGlmeV9ncm91cBgBIAEoCzIWLmFnZW'
        '50cy52MS5Ob3RpZnlHcm91cFILbm90aWZ5R3JvdXA=');

@$core.Deprecated('Use updateNotifyGroupRequestDescriptor instead')
const UpdateNotifyGroupRequest$json = {
  '1': 'UpdateNotifyGroupRequest',
  '2': [
    {
      '1': 'notify_group',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.NotifyGroup',
      '10': 'notifyGroup'
    },
  ],
};

/// Descriptor for `UpdateNotifyGroupRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List updateNotifyGroupRequestDescriptor =
    $convert.base64Decode(
        'ChhVcGRhdGVOb3RpZnlHcm91cFJlcXVlc3QSOQoMbm90aWZ5X2dyb3VwGAEgASgLMhYuYWdlbn'
        'RzLnYxLk5vdGlmeUdyb3VwUgtub3RpZnlHcm91cA==');

@$core.Deprecated('Use updateNotifyGroupResponseDescriptor instead')
const UpdateNotifyGroupResponse$json = {
  '1': 'UpdateNotifyGroupResponse',
  '2': [
    {
      '1': 'notify_group',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.NotifyGroup',
      '10': 'notifyGroup'
    },
  ],
};

/// Descriptor for `UpdateNotifyGroupResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List updateNotifyGroupResponseDescriptor =
    $convert.base64Decode(
        'ChlVcGRhdGVOb3RpZnlHcm91cFJlc3BvbnNlEjkKDG5vdGlmeV9ncm91cBgBIAEoCzIWLmFnZW'
        '50cy52MS5Ob3RpZnlHcm91cFILbm90aWZ5R3JvdXA=');

@$core.Deprecated('Use deleteNotifyGroupRequestDescriptor instead')
const DeleteNotifyGroupRequest$json = {
  '1': 'DeleteNotifyGroupRequest',
  '2': [
    {'1': 'name', '3': 1, '4': 1, '5': 9, '10': 'name'},
  ],
};

/// Descriptor for `DeleteNotifyGroupRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List deleteNotifyGroupRequestDescriptor =
    $convert.base64Decode(
        'ChhEZWxldGVOb3RpZnlHcm91cFJlcXVlc3QSEgoEbmFtZRgBIAEoCVIEbmFtZQ==');

@$core.Deprecated('Use deleteNotifyGroupResponseDescriptor instead')
const DeleteNotifyGroupResponse$json = {
  '1': 'DeleteNotifyGroupResponse',
};

/// Descriptor for `DeleteNotifyGroupResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List deleteNotifyGroupResponseDescriptor =
    $convert.base64Decode('ChlEZWxldGVOb3RpZnlHcm91cFJlc3BvbnNl');

@$core.Deprecated('Use listRemoteAgentsRequestDescriptor instead')
const ListRemoteAgentsRequest$json = {
  '1': 'ListRemoteAgentsRequest',
};

/// Descriptor for `ListRemoteAgentsRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listRemoteAgentsRequestDescriptor =
    $convert.base64Decode('ChdMaXN0UmVtb3RlQWdlbnRzUmVxdWVzdA==');

@$core.Deprecated('Use listRemoteAgentsResponseDescriptor instead')
const ListRemoteAgentsResponse$json = {
  '1': 'ListRemoteAgentsResponse',
  '2': [
    {
      '1': 'remote_agents',
      '3': 1,
      '4': 3,
      '5': 11,
      '6': '.agents.v1.RemoteAgent',
      '10': 'remoteAgents'
    },
  ],
};

/// Descriptor for `ListRemoteAgentsResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listRemoteAgentsResponseDescriptor =
    $convert.base64Decode(
        'ChhMaXN0UmVtb3RlQWdlbnRzUmVzcG9uc2USOwoNcmVtb3RlX2FnZW50cxgBIAMoCzIWLmFnZW'
        '50cy52MS5SZW1vdGVBZ2VudFIMcmVtb3RlQWdlbnRz');

@$core.Deprecated('Use getRemoteAgentRequestDescriptor instead')
const GetRemoteAgentRequest$json = {
  '1': 'GetRemoteAgentRequest',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
  ],
};

/// Descriptor for `GetRemoteAgentRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getRemoteAgentRequestDescriptor = $convert
    .base64Decode('ChVHZXRSZW1vdGVBZ2VudFJlcXVlc3QSDgoCaWQYASABKAlSAmlk');

@$core.Deprecated('Use getRemoteAgentResponseDescriptor instead')
const GetRemoteAgentResponse$json = {
  '1': 'GetRemoteAgentResponse',
  '2': [
    {
      '1': 'remote_agent',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.RemoteAgent',
      '10': 'remoteAgent'
    },
  ],
};

/// Descriptor for `GetRemoteAgentResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getRemoteAgentResponseDescriptor =
    $convert.base64Decode(
        'ChZHZXRSZW1vdGVBZ2VudFJlc3BvbnNlEjkKDHJlbW90ZV9hZ2VudBgBIAEoCzIWLmFnZW50cy'
        '52MS5SZW1vdGVBZ2VudFILcmVtb3RlQWdlbnQ=');

@$core.Deprecated('Use createRemoteAgentRequestDescriptor instead')
const CreateRemoteAgentRequest$json = {
  '1': 'CreateRemoteAgentRequest',
  '2': [
    {
      '1': 'remote_agent',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.RemoteAgent',
      '10': 'remoteAgent'
    },
  ],
};

/// Descriptor for `CreateRemoteAgentRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List createRemoteAgentRequestDescriptor =
    $convert.base64Decode(
        'ChhDcmVhdGVSZW1vdGVBZ2VudFJlcXVlc3QSOQoMcmVtb3RlX2FnZW50GAEgASgLMhYuYWdlbn'
        'RzLnYxLlJlbW90ZUFnZW50UgtyZW1vdGVBZ2VudA==');

@$core.Deprecated('Use createRemoteAgentResponseDescriptor instead')
const CreateRemoteAgentResponse$json = {
  '1': 'CreateRemoteAgentResponse',
  '2': [
    {
      '1': 'remote_agent',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.RemoteAgent',
      '10': 'remoteAgent'
    },
  ],
};

/// Descriptor for `CreateRemoteAgentResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List createRemoteAgentResponseDescriptor =
    $convert.base64Decode(
        'ChlDcmVhdGVSZW1vdGVBZ2VudFJlc3BvbnNlEjkKDHJlbW90ZV9hZ2VudBgBIAEoCzIWLmFnZW'
        '50cy52MS5SZW1vdGVBZ2VudFILcmVtb3RlQWdlbnQ=');

@$core.Deprecated('Use updateRemoteAgentRequestDescriptor instead')
const UpdateRemoteAgentRequest$json = {
  '1': 'UpdateRemoteAgentRequest',
  '2': [
    {
      '1': 'remote_agent',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.RemoteAgent',
      '10': 'remoteAgent'
    },
  ],
};

/// Descriptor for `UpdateRemoteAgentRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List updateRemoteAgentRequestDescriptor =
    $convert.base64Decode(
        'ChhVcGRhdGVSZW1vdGVBZ2VudFJlcXVlc3QSOQoMcmVtb3RlX2FnZW50GAEgASgLMhYuYWdlbn'
        'RzLnYxLlJlbW90ZUFnZW50UgtyZW1vdGVBZ2VudA==');

@$core.Deprecated('Use updateRemoteAgentResponseDescriptor instead')
const UpdateRemoteAgentResponse$json = {
  '1': 'UpdateRemoteAgentResponse',
  '2': [
    {
      '1': 'remote_agent',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.RemoteAgent',
      '10': 'remoteAgent'
    },
  ],
};

/// Descriptor for `UpdateRemoteAgentResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List updateRemoteAgentResponseDescriptor =
    $convert.base64Decode(
        'ChlVcGRhdGVSZW1vdGVBZ2VudFJlc3BvbnNlEjkKDHJlbW90ZV9hZ2VudBgBIAEoCzIWLmFnZW'
        '50cy52MS5SZW1vdGVBZ2VudFILcmVtb3RlQWdlbnQ=');

@$core.Deprecated('Use deleteRemoteAgentRequestDescriptor instead')
const DeleteRemoteAgentRequest$json = {
  '1': 'DeleteRemoteAgentRequest',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
  ],
};

/// Descriptor for `DeleteRemoteAgentRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List deleteRemoteAgentRequestDescriptor = $convert
    .base64Decode('ChhEZWxldGVSZW1vdGVBZ2VudFJlcXVlc3QSDgoCaWQYASABKAlSAmlk');

@$core.Deprecated('Use deleteRemoteAgentResponseDescriptor instead')
const DeleteRemoteAgentResponse$json = {
  '1': 'DeleteRemoteAgentResponse',
};

/// Descriptor for `DeleteRemoteAgentResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List deleteRemoteAgentResponseDescriptor =
    $convert.base64Decode('ChlEZWxldGVSZW1vdGVBZ2VudFJlc3BvbnNl');

@$core.Deprecated('Use listChannelsRequestDescriptor instead')
const ListChannelsRequest$json = {
  '1': 'ListChannelsRequest',
};

/// Descriptor for `ListChannelsRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listChannelsRequestDescriptor =
    $convert.base64Decode('ChNMaXN0Q2hhbm5lbHNSZXF1ZXN0');

@$core.Deprecated('Use listChannelsResponseDescriptor instead')
const ListChannelsResponse$json = {
  '1': 'ListChannelsResponse',
  '2': [
    {
      '1': 'channels',
      '3': 1,
      '4': 3,
      '5': 11,
      '6': '.agents.v1.AgentChannel',
      '10': 'channels'
    },
  ],
};

/// Descriptor for `ListChannelsResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listChannelsResponseDescriptor = $convert.base64Decode(
    'ChRMaXN0Q2hhbm5lbHNSZXNwb25zZRIzCghjaGFubmVscxgBIAMoCzIXLmFnZW50cy52MS5BZ2'
    'VudENoYW5uZWxSCGNoYW5uZWxz');

@$core.Deprecated('Use getChannelRequestDescriptor instead')
const GetChannelRequest$json = {
  '1': 'GetChannelRequest',
  '2': [
    {'1': 'name', '3': 1, '4': 1, '5': 9, '10': 'name'},
  ],
};

/// Descriptor for `GetChannelRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getChannelRequestDescriptor = $convert
    .base64Decode('ChFHZXRDaGFubmVsUmVxdWVzdBISCgRuYW1lGAEgASgJUgRuYW1l');

@$core.Deprecated('Use getChannelResponseDescriptor instead')
const GetChannelResponse$json = {
  '1': 'GetChannelResponse',
  '2': [
    {
      '1': 'channel',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.AgentChannel',
      '10': 'channel'
    },
  ],
};

/// Descriptor for `GetChannelResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getChannelResponseDescriptor = $convert.base64Decode(
    'ChJHZXRDaGFubmVsUmVzcG9uc2USMQoHY2hhbm5lbBgBIAEoCzIXLmFnZW50cy52MS5BZ2VudE'
    'NoYW5uZWxSB2NoYW5uZWw=');

@$core.Deprecated('Use createChannelRequestDescriptor instead')
const CreateChannelRequest$json = {
  '1': 'CreateChannelRequest',
  '2': [
    {
      '1': 'channel',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.AgentChannel',
      '10': 'channel'
    },
  ],
};

/// Descriptor for `CreateChannelRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List createChannelRequestDescriptor = $convert.base64Decode(
    'ChRDcmVhdGVDaGFubmVsUmVxdWVzdBIxCgdjaGFubmVsGAEgASgLMhcuYWdlbnRzLnYxLkFnZW'
    '50Q2hhbm5lbFIHY2hhbm5lbA==');

@$core.Deprecated('Use createChannelResponseDescriptor instead')
const CreateChannelResponse$json = {
  '1': 'CreateChannelResponse',
  '2': [
    {
      '1': 'channel',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.AgentChannel',
      '10': 'channel'
    },
  ],
};

/// Descriptor for `CreateChannelResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List createChannelResponseDescriptor = $convert.base64Decode(
    'ChVDcmVhdGVDaGFubmVsUmVzcG9uc2USMQoHY2hhbm5lbBgBIAEoCzIXLmFnZW50cy52MS5BZ2'
    'VudENoYW5uZWxSB2NoYW5uZWw=');

@$core.Deprecated('Use updateChannelRequestDescriptor instead')
const UpdateChannelRequest$json = {
  '1': 'UpdateChannelRequest',
  '2': [
    {
      '1': 'channel',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.AgentChannel',
      '10': 'channel'
    },
  ],
};

/// Descriptor for `UpdateChannelRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List updateChannelRequestDescriptor = $convert.base64Decode(
    'ChRVcGRhdGVDaGFubmVsUmVxdWVzdBIxCgdjaGFubmVsGAEgASgLMhcuYWdlbnRzLnYxLkFnZW'
    '50Q2hhbm5lbFIHY2hhbm5lbA==');

@$core.Deprecated('Use updateChannelResponseDescriptor instead')
const UpdateChannelResponse$json = {
  '1': 'UpdateChannelResponse',
  '2': [
    {
      '1': 'channel',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.AgentChannel',
      '10': 'channel'
    },
  ],
};

/// Descriptor for `UpdateChannelResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List updateChannelResponseDescriptor = $convert.base64Decode(
    'ChVVcGRhdGVDaGFubmVsUmVzcG9uc2USMQoHY2hhbm5lbBgBIAEoCzIXLmFnZW50cy52MS5BZ2'
    'VudENoYW5uZWxSB2NoYW5uZWw=');

@$core.Deprecated('Use deleteChannelRequestDescriptor instead')
const DeleteChannelRequest$json = {
  '1': 'DeleteChannelRequest',
  '2': [
    {'1': 'name', '3': 1, '4': 1, '5': 9, '10': 'name'},
  ],
};

/// Descriptor for `DeleteChannelRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List deleteChannelRequestDescriptor = $convert
    .base64Decode('ChREZWxldGVDaGFubmVsUmVxdWVzdBISCgRuYW1lGAEgASgJUgRuYW1l');

@$core.Deprecated('Use deleteChannelResponseDescriptor instead')
const DeleteChannelResponse$json = {
  '1': 'DeleteChannelResponse',
};

/// Descriptor for `DeleteChannelResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List deleteChannelResponseDescriptor =
    $convert.base64Decode('ChVEZWxldGVDaGFubmVsUmVzcG9uc2U=');

@$core.Deprecated('Use restartChannelRequestDescriptor instead')
const RestartChannelRequest$json = {
  '1': 'RestartChannelRequest',
  '2': [
    {'1': 'name', '3': 1, '4': 1, '5': 9, '10': 'name'},
  ],
};

/// Descriptor for `RestartChannelRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List restartChannelRequestDescriptor =
    $convert.base64Decode(
        'ChVSZXN0YXJ0Q2hhbm5lbFJlcXVlc3QSEgoEbmFtZRgBIAEoCVIEbmFtZQ==');

@$core.Deprecated('Use restartChannelResponseDescriptor instead')
const RestartChannelResponse$json = {
  '1': 'RestartChannelResponse',
  '2': [
    {
      '1': 'channel',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.AgentChannel',
      '10': 'channel'
    },
  ],
};

/// Descriptor for `RestartChannelResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List restartChannelResponseDescriptor =
    $convert.base64Decode(
        'ChZSZXN0YXJ0Q2hhbm5lbFJlc3BvbnNlEjEKB2NoYW5uZWwYASABKAsyFy5hZ2VudHMudjEuQW'
        'dlbnRDaGFubmVsUgdjaGFubmVs');

@$core.Deprecated('Use pauseChannelRequestDescriptor instead')
const PauseChannelRequest$json = {
  '1': 'PauseChannelRequest',
  '2': [
    {'1': 'name', '3': 1, '4': 1, '5': 9, '10': 'name'},
  ],
};

/// Descriptor for `PauseChannelRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List pauseChannelRequestDescriptor = $convert
    .base64Decode('ChNQYXVzZUNoYW5uZWxSZXF1ZXN0EhIKBG5hbWUYASABKAlSBG5hbWU=');

@$core.Deprecated('Use pauseChannelResponseDescriptor instead')
const PauseChannelResponse$json = {
  '1': 'PauseChannelResponse',
  '2': [
    {
      '1': 'channel',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.AgentChannel',
      '10': 'channel'
    },
  ],
};

/// Descriptor for `PauseChannelResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List pauseChannelResponseDescriptor = $convert.base64Decode(
    'ChRQYXVzZUNoYW5uZWxSZXNwb25zZRIxCgdjaGFubmVsGAEgASgLMhcuYWdlbnRzLnYxLkFnZW'
    '50Q2hhbm5lbFIHY2hhbm5lbA==');

@$core.Deprecated('Use resumeChannelRequestDescriptor instead')
const ResumeChannelRequest$json = {
  '1': 'ResumeChannelRequest',
  '2': [
    {'1': 'name', '3': 1, '4': 1, '5': 9, '10': 'name'},
  ],
};

/// Descriptor for `ResumeChannelRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List resumeChannelRequestDescriptor = $convert
    .base64Decode('ChRSZXN1bWVDaGFubmVsUmVxdWVzdBISCgRuYW1lGAEgASgJUgRuYW1l');

@$core.Deprecated('Use resumeChannelResponseDescriptor instead')
const ResumeChannelResponse$json = {
  '1': 'ResumeChannelResponse',
  '2': [
    {
      '1': 'channel',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.AgentChannel',
      '10': 'channel'
    },
  ],
};

/// Descriptor for `ResumeChannelResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List resumeChannelResponseDescriptor = $convert.base64Decode(
    'ChVSZXN1bWVDaGFubmVsUmVzcG9uc2USMQoHY2hhbm5lbBgBIAEoCzIXLmFnZW50cy52MS5BZ2'
    'VudENoYW5uZWxSB2NoYW5uZWw=');

@$core.Deprecated('Use createSessionRequestDescriptor instead')
const CreateSessionRequest$json = {
  '1': 'CreateSessionRequest',
  '2': [
    {'1': 'app_name', '3': 1, '4': 1, '5': 9, '10': 'appName'},
    {'1': 'user_id', '3': 2, '4': 1, '5': 9, '10': 'userId'},
    {'1': 'session_id', '3': 3, '4': 1, '5': 9, '10': 'sessionId'},
    {
      '1': 'state',
      '3': 4,
      '4': 1,
      '5': 11,
      '6': '.google.protobuf.Struct',
      '10': 'state'
    },
  ],
};

/// Descriptor for `CreateSessionRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List createSessionRequestDescriptor = $convert.base64Decode(
    'ChRDcmVhdGVTZXNzaW9uUmVxdWVzdBIZCghhcHBfbmFtZRgBIAEoCVIHYXBwTmFtZRIXCgd1c2'
    'VyX2lkGAIgASgJUgZ1c2VySWQSHQoKc2Vzc2lvbl9pZBgDIAEoCVIJc2Vzc2lvbklkEi0KBXN0'
    'YXRlGAQgASgLMhcuZ29vZ2xlLnByb3RvYnVmLlN0cnVjdFIFc3RhdGU=');

@$core.Deprecated('Use createSessionResponseDescriptor instead')
const CreateSessionResponse$json = {
  '1': 'CreateSessionResponse',
  '2': [
    {
      '1': 'session',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.SessionInfo',
      '10': 'session'
    },
  ],
};

/// Descriptor for `CreateSessionResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List createSessionResponseDescriptor = $convert.base64Decode(
    'ChVDcmVhdGVTZXNzaW9uUmVzcG9uc2USMAoHc2Vzc2lvbhgBIAEoCzIWLmFnZW50cy52MS5TZX'
    'NzaW9uSW5mb1IHc2Vzc2lvbg==');

@$core.Deprecated('Use getSessionRequestDescriptor instead')
const GetSessionRequest$json = {
  '1': 'GetSessionRequest',
  '2': [
    {'1': 'app_name', '3': 1, '4': 1, '5': 9, '10': 'appName'},
    {'1': 'user_id', '3': 2, '4': 1, '5': 9, '10': 'userId'},
    {'1': 'session_id', '3': 3, '4': 1, '5': 9, '10': 'sessionId'},
    {'1': 'num_recent_events', '3': 4, '4': 1, '5': 5, '10': 'numRecentEvents'},
  ],
};

/// Descriptor for `GetSessionRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getSessionRequestDescriptor = $convert.base64Decode(
    'ChFHZXRTZXNzaW9uUmVxdWVzdBIZCghhcHBfbmFtZRgBIAEoCVIHYXBwTmFtZRIXCgd1c2VyX2'
    'lkGAIgASgJUgZ1c2VySWQSHQoKc2Vzc2lvbl9pZBgDIAEoCVIJc2Vzc2lvbklkEioKEW51bV9y'
    'ZWNlbnRfZXZlbnRzGAQgASgFUg9udW1SZWNlbnRFdmVudHM=');

@$core.Deprecated('Use getSessionResponseDescriptor instead')
const GetSessionResponse$json = {
  '1': 'GetSessionResponse',
  '2': [
    {
      '1': 'session_detail',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.SessionDetail',
      '10': 'sessionDetail'
    },
  ],
};

/// Descriptor for `GetSessionResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getSessionResponseDescriptor = $convert.base64Decode(
    'ChJHZXRTZXNzaW9uUmVzcG9uc2USPwoOc2Vzc2lvbl9kZXRhaWwYASABKAsyGC5hZ2VudHMudj'
    'EuU2Vzc2lvbkRldGFpbFINc2Vzc2lvbkRldGFpbA==');

@$core.Deprecated('Use listSessionsRequestDescriptor instead')
const ListSessionsRequest$json = {
  '1': 'ListSessionsRequest',
  '2': [
    {'1': 'app_name', '3': 1, '4': 1, '5': 9, '10': 'appName'},
    {'1': 'user_id', '3': 2, '4': 1, '5': 9, '10': 'userId'},
    {
      '1': 'start_time',
      '3': 3,
      '4': 1,
      '5': 11,
      '6': '.google.protobuf.Timestamp',
      '10': 'startTime'
    },
    {
      '1': 'end_time',
      '3': 4,
      '4': 1,
      '5': 11,
      '6': '.google.protobuf.Timestamp',
      '10': 'endTime'
    },
    {'1': 'page_size', '3': 5, '4': 1, '5': 5, '10': 'pageSize'},
    {'1': 'page_token', '3': 6, '4': 1, '5': 9, '10': 'pageToken'},
    {'1': 'workspace_scoped', '3': 7, '4': 1, '5': 8, '10': 'workspaceScoped'},
  ],
};

/// Descriptor for `ListSessionsRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listSessionsRequestDescriptor = $convert.base64Decode(
    'ChNMaXN0U2Vzc2lvbnNSZXF1ZXN0EhkKCGFwcF9uYW1lGAEgASgJUgdhcHBOYW1lEhcKB3VzZX'
    'JfaWQYAiABKAlSBnVzZXJJZBI5CgpzdGFydF90aW1lGAMgASgLMhouZ29vZ2xlLnByb3RvYnVm'
    'LlRpbWVzdGFtcFIJc3RhcnRUaW1lEjUKCGVuZF90aW1lGAQgASgLMhouZ29vZ2xlLnByb3RvYn'
    'VmLlRpbWVzdGFtcFIHZW5kVGltZRIbCglwYWdlX3NpemUYBSABKAVSCHBhZ2VTaXplEh0KCnBh'
    'Z2VfdG9rZW4YBiABKAlSCXBhZ2VUb2tlbhIpChB3b3Jrc3BhY2Vfc2NvcGVkGAcgASgIUg93b3'
    'Jrc3BhY2VTY29wZWQ=');

@$core.Deprecated('Use listSessionsResponseDescriptor instead')
const ListSessionsResponse$json = {
  '1': 'ListSessionsResponse',
  '2': [
    {
      '1': 'sessions',
      '3': 1,
      '4': 3,
      '5': 11,
      '6': '.agents.v1.SessionInfo',
      '10': 'sessions'
    },
    {'1': 'next_page_token', '3': 2, '4': 1, '5': 9, '10': 'nextPageToken'},
    {'1': 'total', '3': 3, '4': 1, '5': 5, '10': 'total'},
  ],
};

/// Descriptor for `ListSessionsResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listSessionsResponseDescriptor = $convert.base64Decode(
    'ChRMaXN0U2Vzc2lvbnNSZXNwb25zZRIyCghzZXNzaW9ucxgBIAMoCzIWLmFnZW50cy52MS5TZX'
    'NzaW9uSW5mb1IIc2Vzc2lvbnMSJgoPbmV4dF9wYWdlX3Rva2VuGAIgASgJUg1uZXh0UGFnZVRv'
    'a2VuEhQKBXRvdGFsGAMgASgFUgV0b3RhbA==');

@$core.Deprecated('Use deleteSessionRequestDescriptor instead')
const DeleteSessionRequest$json = {
  '1': 'DeleteSessionRequest',
  '2': [
    {'1': 'app_name', '3': 1, '4': 1, '5': 9, '10': 'appName'},
    {'1': 'user_id', '3': 2, '4': 1, '5': 9, '10': 'userId'},
    {'1': 'session_id', '3': 3, '4': 1, '5': 9, '10': 'sessionId'},
  ],
};

/// Descriptor for `DeleteSessionRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List deleteSessionRequestDescriptor = $convert.base64Decode(
    'ChREZWxldGVTZXNzaW9uUmVxdWVzdBIZCghhcHBfbmFtZRgBIAEoCVIHYXBwTmFtZRIXCgd1c2'
    'VyX2lkGAIgASgJUgZ1c2VySWQSHQoKc2Vzc2lvbl9pZBgDIAEoCVIJc2Vzc2lvbklk');

@$core.Deprecated('Use deleteSessionResponseDescriptor instead')
const DeleteSessionResponse$json = {
  '1': 'DeleteSessionResponse',
};

/// Descriptor for `DeleteSessionResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List deleteSessionResponseDescriptor =
    $convert.base64Decode('ChVEZWxldGVTZXNzaW9uUmVzcG9uc2U=');

@$core.Deprecated('Use replySessionRequestDescriptor instead')
const ReplySessionRequest$json = {
  '1': 'ReplySessionRequest',
  '2': [
    {
      '1': 'agent_name',
      '3': 1,
      '4': 1,
      '5': 9,
      '8': {'3': true},
      '10': 'agentName',
    },
    {'1': 'app_name', '3': 2, '4': 1, '5': 9, '10': 'appName'},
    {'1': 'user_id', '3': 3, '4': 1, '5': 9, '10': 'userId'},
    {'1': 'session_id', '3': 4, '4': 1, '5': 9, '10': 'sessionId'},
    {'1': 'message', '3': 5, '4': 1, '5': 9, '10': 'message'},
    {'1': 'model_override', '3': 6, '4': 1, '5': 9, '10': 'modelOverride'},
    {
      '1': 'parts',
      '3': 7,
      '4': 3,
      '5': 11,
      '6': '.agents.v1.InputPart',
      '10': 'parts'
    },
    {'1': 'agent_id', '3': 8, '4': 1, '5': 9, '10': 'agentId'},
  ],
};

/// Descriptor for `ReplySessionRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List replySessionRequestDescriptor = $convert.base64Decode(
    'ChNSZXBseVNlc3Npb25SZXF1ZXN0EiEKCmFnZW50X25hbWUYASABKAlCAhgBUglhZ2VudE5hbW'
    'USGQoIYXBwX25hbWUYAiABKAlSB2FwcE5hbWUSFwoHdXNlcl9pZBgDIAEoCVIGdXNlcklkEh0K'
    'CnNlc3Npb25faWQYBCABKAlSCXNlc3Npb25JZBIYCgdtZXNzYWdlGAUgASgJUgdtZXNzYWdlEi'
    'UKDm1vZGVsX292ZXJyaWRlGAYgASgJUg1tb2RlbE92ZXJyaWRlEioKBXBhcnRzGAcgAygLMhQu'
    'YWdlbnRzLnYxLklucHV0UGFydFIFcGFydHMSGQoIYWdlbnRfaWQYCCABKAlSB2FnZW50SWQ=');

@$core.Deprecated('Use replySessionResponseDescriptor instead')
const ReplySessionResponse$json = {
  '1': 'ReplySessionResponse',
  '2': [
    {'1': 'response', '3': 1, '4': 1, '5': 9, '10': 'response'},
  ],
};

/// Descriptor for `ReplySessionResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List replySessionResponseDescriptor =
    $convert.base64Decode(
        'ChRSZXBseVNlc3Npb25SZXNwb25zZRIaCghyZXNwb25zZRgBIAEoCVIIcmVzcG9uc2U=');

@$core.Deprecated('Use updateSessionTitleRequestDescriptor instead')
const UpdateSessionTitleRequest$json = {
  '1': 'UpdateSessionTitleRequest',
  '2': [
    {'1': 'app_name', '3': 1, '4': 1, '5': 9, '10': 'appName'},
    {'1': 'user_id', '3': 2, '4': 1, '5': 9, '10': 'userId'},
    {'1': 'session_id', '3': 3, '4': 1, '5': 9, '10': 'sessionId'},
    {'1': 'title', '3': 4, '4': 1, '5': 9, '10': 'title'},
  ],
};

/// Descriptor for `UpdateSessionTitleRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List updateSessionTitleRequestDescriptor = $convert.base64Decode(
    'ChlVcGRhdGVTZXNzaW9uVGl0bGVSZXF1ZXN0EhkKCGFwcF9uYW1lGAEgASgJUgdhcHBOYW1lEh'
    'cKB3VzZXJfaWQYAiABKAlSBnVzZXJJZBIdCgpzZXNzaW9uX2lkGAMgASgJUglzZXNzaW9uSWQS'
    'FAoFdGl0bGUYBCABKAlSBXRpdGxl');

@$core.Deprecated('Use updateSessionTitleResponseDescriptor instead')
const UpdateSessionTitleResponse$json = {
  '1': 'UpdateSessionTitleResponse',
  '2': [
    {
      '1': 'session',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.SessionInfo',
      '10': 'session'
    },
  ],
};

/// Descriptor for `UpdateSessionTitleResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List updateSessionTitleResponseDescriptor =
    $convert.base64Decode(
        'ChpVcGRhdGVTZXNzaW9uVGl0bGVSZXNwb25zZRIwCgdzZXNzaW9uGAEgASgLMhYuYWdlbnRzLn'
        'YxLlNlc3Npb25JbmZvUgdzZXNzaW9u');

@$core.Deprecated('Use sessionInfoDescriptor instead')
const SessionInfo$json = {
  '1': 'SessionInfo',
  '2': [
    {'1': 'session_id', '3': 1, '4': 1, '5': 9, '10': 'sessionId'},
    {'1': 'app_name', '3': 2, '4': 1, '5': 9, '10': 'appName'},
    {'1': 'user_id', '3': 3, '4': 1, '5': 9, '10': 'userId'},
    {
      '1': 'state',
      '3': 4,
      '4': 1,
      '5': 11,
      '6': '.google.protobuf.Struct',
      '10': 'state'
    },
    {
      '1': 'last_update_time',
      '3': 5,
      '4': 1,
      '5': 11,
      '6': '.google.protobuf.Timestamp',
      '10': 'lastUpdateTime'
    },
    {'1': 'turn_count', '3': 6, '4': 1, '5': 5, '10': 'turnCount'},
    {'1': 'title', '3': 7, '4': 1, '5': 9, '10': 'title'},
    {
      '1': 'latest_invocation_id',
      '3': 8,
      '4': 1,
      '5': 9,
      '10': 'latestInvocationId'
    },
    {
      '1': 'latest_invocation_status',
      '3': 9,
      '4': 1,
      '5': 14,
      '6': '.agents.v1.InvocationStatus',
      '10': 'latestInvocationStatus'
    },
    {
      '1': 'has_pending_interrupt',
      '3': 10,
      '4': 1,
      '5': 8,
      '10': 'hasPendingInterrupt'
    },
    {'1': 'unread', '3': 11, '4': 1, '5': 8, '10': 'unread'},
    {
      '1': 'last_read_at',
      '3': 12,
      '4': 1,
      '5': 11,
      '6': '.google.protobuf.Timestamp',
      '10': 'lastReadAt'
    },
    {'1': 'workspace_id', '3': 100, '4': 1, '5': 9, '10': 'workspaceId'},
  ],
};

/// Descriptor for `SessionInfo`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List sessionInfoDescriptor = $convert.base64Decode(
    'CgtTZXNzaW9uSW5mbxIdCgpzZXNzaW9uX2lkGAEgASgJUglzZXNzaW9uSWQSGQoIYXBwX25hbW'
    'UYAiABKAlSB2FwcE5hbWUSFwoHdXNlcl9pZBgDIAEoCVIGdXNlcklkEi0KBXN0YXRlGAQgASgL'
    'MhcuZ29vZ2xlLnByb3RvYnVmLlN0cnVjdFIFc3RhdGUSRAoQbGFzdF91cGRhdGVfdGltZRgFIA'
    'EoCzIaLmdvb2dsZS5wcm90b2J1Zi5UaW1lc3RhbXBSDmxhc3RVcGRhdGVUaW1lEh0KCnR1cm5f'
    'Y291bnQYBiABKAVSCXR1cm5Db3VudBIUCgV0aXRsZRgHIAEoCVIFdGl0bGUSMAoUbGF0ZXN0X2'
    'ludm9jYXRpb25faWQYCCABKAlSEmxhdGVzdEludm9jYXRpb25JZBJVChhsYXRlc3RfaW52b2Nh'
    'dGlvbl9zdGF0dXMYCSABKA4yGy5hZ2VudHMudjEuSW52b2NhdGlvblN0YXR1c1IWbGF0ZXN0SW'
    '52b2NhdGlvblN0YXR1cxIyChVoYXNfcGVuZGluZ19pbnRlcnJ1cHQYCiABKAhSE2hhc1BlbmRp'
    'bmdJbnRlcnJ1cHQSFgoGdW5yZWFkGAsgASgIUgZ1bnJlYWQSPAoMbGFzdF9yZWFkX2F0GAwgAS'
    'gLMhouZ29vZ2xlLnByb3RvYnVmLlRpbWVzdGFtcFIKbGFzdFJlYWRBdBIhCgx3b3Jrc3BhY2Vf'
    'aWQYZCABKAlSC3dvcmtzcGFjZUlk');

@$core.Deprecated('Use sessionDetailDescriptor instead')
const SessionDetail$json = {
  '1': 'SessionDetail',
  '2': [
    {
      '1': 'session',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.SessionInfo',
      '10': 'session'
    },
    {
      '1': 'events',
      '3': 2,
      '4': 3,
      '5': 11,
      '6': '.agents.v1.SessionEvent',
      '10': 'events'
    },
    {
      '1': 'duration',
      '3': 3,
      '4': 1,
      '5': 11,
      '6': '.google.protobuf.Duration',
      '10': 'duration'
    },
  ],
};

/// Descriptor for `SessionDetail`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List sessionDetailDescriptor = $convert.base64Decode(
    'Cg1TZXNzaW9uRGV0YWlsEjAKB3Nlc3Npb24YASABKAsyFi5hZ2VudHMudjEuU2Vzc2lvbkluZm'
    '9SB3Nlc3Npb24SLwoGZXZlbnRzGAIgAygLMhcuYWdlbnRzLnYxLlNlc3Npb25FdmVudFIGZXZl'
    'bnRzEjUKCGR1cmF0aW9uGAMgASgLMhkuZ29vZ2xlLnByb3RvYnVmLkR1cmF0aW9uUghkdXJhdG'
    'lvbg==');

@$core.Deprecated('Use generateSessionTitleRequestDescriptor instead')
const GenerateSessionTitleRequest$json = {
  '1': 'GenerateSessionTitleRequest',
  '2': [
    {'1': 'app_name', '3': 1, '4': 1, '5': 9, '10': 'appName'},
    {'1': 'user_id', '3': 2, '4': 1, '5': 9, '10': 'userId'},
    {'1': 'session_id', '3': 3, '4': 1, '5': 9, '10': 'sessionId'},
  ],
};

/// Descriptor for `GenerateSessionTitleRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List generateSessionTitleRequestDescriptor =
    $convert.base64Decode(
        'ChtHZW5lcmF0ZVNlc3Npb25UaXRsZVJlcXVlc3QSGQoIYXBwX25hbWUYASABKAlSB2FwcE5hbW'
        'USFwoHdXNlcl9pZBgCIAEoCVIGdXNlcklkEh0KCnNlc3Npb25faWQYAyABKAlSCXNlc3Npb25J'
        'ZA==');

@$core.Deprecated('Use generateSessionTitleResponseDescriptor instead')
const GenerateSessionTitleResponse$json = {
  '1': 'GenerateSessionTitleResponse',
  '2': [
    {
      '1': 'session',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.SessionInfo',
      '10': 'session'
    },
    {'1': 'generated', '3': 2, '4': 1, '5': 8, '10': 'generated'},
  ],
};

/// Descriptor for `GenerateSessionTitleResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List generateSessionTitleResponseDescriptor =
    $convert.base64Decode(
        'ChxHZW5lcmF0ZVNlc3Npb25UaXRsZVJlc3BvbnNlEjAKB3Nlc3Npb24YASABKAsyFi5hZ2VudH'
        'MudjEuU2Vzc2lvbkluZm9SB3Nlc3Npb24SHAoJZ2VuZXJhdGVkGAIgASgIUglnZW5lcmF0ZWQ=');

@$core.Deprecated('Use markSessionReadRequestDescriptor instead')
const MarkSessionReadRequest$json = {
  '1': 'MarkSessionReadRequest',
  '2': [
    {'1': 'app_name', '3': 1, '4': 1, '5': 9, '10': 'appName'},
    {'1': 'user_id', '3': 2, '4': 1, '5': 9, '10': 'userId'},
    {'1': 'session_id', '3': 3, '4': 1, '5': 9, '10': 'sessionId'},
    {
      '1': 'last_read_event_id',
      '3': 4,
      '4': 1,
      '5': 9,
      '10': 'lastReadEventId'
    },
  ],
};

/// Descriptor for `MarkSessionReadRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List markSessionReadRequestDescriptor = $convert.base64Decode(
    'ChZNYXJrU2Vzc2lvblJlYWRSZXF1ZXN0EhkKCGFwcF9uYW1lGAEgASgJUgdhcHBOYW1lEhcKB3'
    'VzZXJfaWQYAiABKAlSBnVzZXJJZBIdCgpzZXNzaW9uX2lkGAMgASgJUglzZXNzaW9uSWQSKwoS'
    'bGFzdF9yZWFkX2V2ZW50X2lkGAQgASgJUg9sYXN0UmVhZEV2ZW50SWQ=');

@$core.Deprecated('Use markSessionReadResponseDescriptor instead')
const MarkSessionReadResponse$json = {
  '1': 'MarkSessionReadResponse',
  '2': [
    {
      '1': 'session',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.SessionInfo',
      '10': 'session'
    },
  ],
};

/// Descriptor for `MarkSessionReadResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List markSessionReadResponseDescriptor =
    $convert.base64Decode(
        'ChdNYXJrU2Vzc2lvblJlYWRSZXNwb25zZRIwCgdzZXNzaW9uGAEgASgLMhYuYWdlbnRzLnYxLl'
        'Nlc3Npb25JbmZvUgdzZXNzaW9u');

@$core.Deprecated('Use sessionEventDescriptor instead')
const SessionEvent$json = {
  '1': 'SessionEvent',
  '2': [
    {'1': 'event_id', '3': 1, '4': 1, '5': 9, '10': 'eventId'},
    {'1': 'invocation_id', '3': 2, '4': 1, '5': 9, '10': 'invocationId'},
    {'1': 'author', '3': 3, '4': 1, '5': 9, '10': 'author'},
    {'1': 'branch', '3': 4, '4': 1, '5': 9, '10': 'branch'},
    {'1': 'content_json', '3': 5, '4': 1, '5': 9, '10': 'contentJson'},
    {
      '1': 'timestamp',
      '3': 6,
      '4': 1,
      '5': 11,
      '6': '.google.protobuf.Timestamp',
      '10': 'timestamp'
    },
    {'1': 'trace_id', '3': 7, '4': 1, '5': 9, '10': 'traceId'},
    {'1': 'trace_url', '3': 8, '4': 1, '5': 9, '10': 'traceUrl'},
  ],
};

/// Descriptor for `SessionEvent`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List sessionEventDescriptor = $convert.base64Decode(
    'CgxTZXNzaW9uRXZlbnQSGQoIZXZlbnRfaWQYASABKAlSB2V2ZW50SWQSIwoNaW52b2NhdGlvbl'
    '9pZBgCIAEoCVIMaW52b2NhdGlvbklkEhYKBmF1dGhvchgDIAEoCVIGYXV0aG9yEhYKBmJyYW5j'
    'aBgEIAEoCVIGYnJhbmNoEiEKDGNvbnRlbnRfanNvbhgFIAEoCVILY29udGVudEpzb24SOAoJdG'
    'ltZXN0YW1wGAYgASgLMhouZ29vZ2xlLnByb3RvYnVmLlRpbWVzdGFtcFIJdGltZXN0YW1wEhkK'
    'CHRyYWNlX2lkGAcgASgJUgd0cmFjZUlkEhsKCXRyYWNlX3VybBgIIAEoCVIIdHJhY2VVcmw=');

const $core.Map<$core.String, $core.dynamic> AgentServiceBase$json = {
  '1': 'AgentService',
  '2': [
    {
      '1': 'ListAgents',
      '2': '.agents.v1.ListAgentsRequest',
      '3': '.agents.v1.ListAgentsResponse'
    },
    {
      '1': 'GetAgent',
      '2': '.agents.v1.GetAgentRequest',
      '3': '.agents.v1.GetAgentResponse'
    },
    {
      '1': 'CreateAgent',
      '2': '.agents.v1.CreateAgentRequest',
      '3': '.agents.v1.CreateAgentResponse'
    },
    {
      '1': 'UpdateAgent',
      '2': '.agents.v1.UpdateAgentRequest',
      '3': '.agents.v1.UpdateAgentResponse'
    },
    {
      '1': 'DeleteAgent',
      '2': '.agents.v1.DeleteAgentRequest',
      '3': '.agents.v1.DeleteAgentResponse'
    },
    {
      '1': 'InvokeAgent',
      '2': '.agents.v1.InvokeAgentRequest',
      '3': '.agents.v1.InvokeAgentResponse'
    },
    {
      '1': 'ListAgentInvocations',
      '2': '.agents.v1.ListAgentInvocationsRequest',
      '3': '.agents.v1.ListAgentInvocationsResponse'
    },
    {
      '1': 'ReloadAgents',
      '2': '.agents.v1.ReloadAgentsRequest',
      '3': '.agents.v1.ReloadAgentsResponse'
    },
    {
      '1': 'GetAgentRuntimeStatus',
      '2': '.agents.v1.GetAgentRuntimeStatusRequest',
      '3': '.agents.v1.GetAgentRuntimeStatusResponse'
    },
    {
      '1': 'ListAgentRuntimeStatuses',
      '2': '.agents.v1.ListAgentRuntimeStatusesRequest',
      '3': '.agents.v1.ListAgentRuntimeStatusesResponse'
    },
    {
      '1': 'CancelAgentInvocation',
      '2': '.agents.v1.CancelAgentInvocationRequest',
      '3': '.agents.v1.CancelAgentInvocationResponse'
    },
    {
      '1': 'StreamAgent',
      '2': '.agents.v1.StreamAgentRequest',
      '3': '.agents.v1.StreamAgentResponse',
      '6': true
    },
    {
      '1': 'AssignAgentID',
      '2': '.agents.v1.AssignAgentIDRequest',
      '3': '.agents.v1.AssignAgentIDResponse',
      '4': {'33': true},
    },
    {
      '1': 'GetMigrationReadiness',
      '2': '.agents.v1.GetMigrationReadinessRequest',
      '3': '.agents.v1.GetMigrationReadinessResponse',
      '4': {'33': true},
    },
    {
      '1': 'MigrateAgentsV2',
      '2': '.agents.v1.MigrateAgentsV2Request',
      '3': '.agents.v1.MigrateAgentsV2Response',
      '4': {'33': true},
    },
    {
      '1': 'VerifyAgentIDCutover',
      '2': '.agents.v1.VerifyAgentIDCutoverRequest',
      '3': '.agents.v1.VerifyAgentIDCutoverResponse'
    },
    {
      '1': 'UpdateAgentConfiguration',
      '2': '.agents.v1.UpdateAgentConfigurationRequest',
      '3': '.agents.v1.UpdateAgentConfigurationResponse'
    },
    {
      '1': 'RestoreAgent',
      '2': '.agents.v1.RestoreAgentRequest',
      '3': '.agents.v1.RestoreAgentResponse'
    },
    {
      '1': 'GetAgentInvocation',
      '2': '.agents.v1.GetAgentInvocationRequest',
      '3': '.agents.v1.GetAgentInvocationResponse'
    },
    {
      '1': 'GetAgentOperation',
      '2': '.agents.v1.GetAgentOperationRequest',
      '3': '.agents.v1.GetAgentOperationResponse'
    },
    {
      '1': 'ListAgentOperations',
      '2': '.agents.v1.ListAgentOperationsRequest',
      '3': '.agents.v1.ListAgentOperationsResponse'
    },
    {
      '1': 'RetryAgentOperation',
      '2': '.agents.v1.RetryAgentOperationRequest',
      '3': '.agents.v1.RetryAgentOperationResponse'
    },
  ],
};

@$core.Deprecated('Use agentServiceDescriptor instead')
const $core.Map<$core.String, $core.Map<$core.String, $core.dynamic>>
    AgentServiceBase$messageJson = {
  '.agents.v1.ListAgentsRequest': ListAgentsRequest$json,
  '.agents.v1.ListAgentsResponse': ListAgentsResponse$json,
  '.agents.v1.Agent': $0.Agent$json,
  '.agents.v1.Agent.LabelsEntry': $0.Agent_LabelsEntry$json,
  '.agents.v1.Agent.MetadataEntry': $0.Agent_MetadataEntry$json,
  '.agents.v1.AgentConfig': $0.AgentConfig$json,
  '.agents.v1.AgentRuntime': $0.AgentRuntime$json,
  '.agents.v1.MCPServer': $0.MCPServer$json,
  '.agents.v1.MCPServer.HeadersEntry': $0.MCPServer_HeadersEntry$json,
  '.agents.v1.MCPServer.MetadataEntry': $0.MCPServer_MetadataEntry$json,
  '.agents.v1.MCPServerAuth': $0.MCPServerAuth$json,
  '.agents.v1.MCPServerOAuth2Config': $0.MCPServerOAuth2Config$json,
  '.agents.v1.ContextGuardConfig': $0.ContextGuardConfig$json,
  '.agents.v1.AgentFileMount': $9.AgentFileMount$json,
  '.agents.v1.MemoryConfig': $0.MemoryConfig$json,
  '.agents.v1.ResultCardConfig': $0.ResultCardConfig$json,
  '.agents.v1.WorkflowConfig': $0.WorkflowConfig$json,
  '.agents.v1.WorkflowNode': $0.WorkflowNode$json,
  '.agents.v1.WorkflowRetryConfig': $0.WorkflowRetryConfig$json,
  '.agents.v1.HumanInputForm': $0.HumanInputForm$json,
  '.agents.v1.HumanInputFormField': $0.HumanInputFormField$json,
  '.agents.v1.HumanInputFormOption': $0.HumanInputFormOption$json,
  '.agents.v1.WorkflowEdge': $0.WorkflowEdge$json,
  '.agents.v1.PiAgentConfig': $0.PiAgentConfig$json,
  '.agents.v1.CursorAgentConfig': $0.CursorAgentConfig$json,
  '.google.protobuf.Timestamp': $1.Timestamp$json,
  '.agents.v1.GetAgentRequest': GetAgentRequest$json,
  '.agents.v1.GetAgentResponse': GetAgentResponse$json,
  '.agents.v1.CreateAgentRequest': CreateAgentRequest$json,
  '.agents.v1.AgentContentInput': AgentContentInput$json,
  '.agents.v1.CreateAgentResponse': CreateAgentResponse$json,
  '.agents.v1.AgentOperation': $3.AgentOperation$json,
  '.agents.v1.AgentOperationStep': $3.AgentOperationStep$json,
  '.agents.v1.UpdateAgentRequest': UpdateAgentRequest$json,
  '.agents.v1.UpdateAgentResponse': UpdateAgentResponse$json,
  '.agents.v1.DeleteAgentRequest': DeleteAgentRequest$json,
  '.agents.v1.DeleteAgentResponse': DeleteAgentResponse$json,
  '.agents.v1.InvokeAgentRequest': InvokeAgentRequest$json,
  '.agents.v1.InvokeAgentResponse': InvokeAgentResponse$json,
  '.agents.v1.ListAgentInvocationsRequest': ListAgentInvocationsRequest$json,
  '.agents.v1.ListAgentInvocationsResponse': ListAgentInvocationsResponse$json,
  '.agents.v1.Invocation': Invocation$json,
  '.agents.v1.ReloadAgentsRequest': ReloadAgentsRequest$json,
  '.agents.v1.ReloadAgentsResponse': ReloadAgentsResponse$json,
  '.agents.v1.GetAgentRuntimeStatusRequest': GetAgentRuntimeStatusRequest$json,
  '.agents.v1.GetAgentRuntimeStatusResponse':
      GetAgentRuntimeStatusResponse$json,
  '.agents.v1.AgentRuntimeStatus': AgentRuntimeStatus$json,
  '.agents.v1.ListAgentRuntimeStatusesRequest':
      ListAgentRuntimeStatusesRequest$json,
  '.agents.v1.ListAgentRuntimeStatusesResponse':
      ListAgentRuntimeStatusesResponse$json,
  '.agents.v1.CancelAgentInvocationRequest': CancelAgentInvocationRequest$json,
  '.agents.v1.CancelAgentInvocationResponse':
      CancelAgentInvocationResponse$json,
  '.agents.v1.StreamAgentRequest': StreamAgentRequest$json,
  '.agents.v1.InputPart': $2.InputPart$json,
  '.agents.v1.InlineData': $2.InlineData$json,
  '.agents.v1.StreamAgentResponse': StreamAgentResponse$json,
  '.agents.v1.StreamAgentStarted': StreamAgentStarted$json,
  '.agents.v1.StreamAgentRunEvent': StreamAgentRunEvent$json,
  '.agents.v1.StreamAgentTextDelta': StreamAgentTextDelta$json,
  '.agents.v1.StreamAgentFinal': StreamAgentFinal$json,
  '.agents.v1.AssignAgentIDRequest': AssignAgentIDRequest$json,
  '.agents.v1.AssignAgentIDResponse': AssignAgentIDResponse$json,
  '.agents.v1.GetMigrationReadinessRequest': GetMigrationReadinessRequest$json,
  '.agents.v1.GetMigrationReadinessResponse':
      GetMigrationReadinessResponse$json,
  '.agents.v1.AgentMigrationStatus': AgentMigrationStatus$json,
  '.agents.v1.MigrateAgentsV2Request': MigrateAgentsV2Request$json,
  '.agents.v1.MigrateAgentsV2Response': MigrateAgentsV2Response$json,
  '.agents.v1.MigrateAgentResult': MigrateAgentResult$json,
  '.agents.v1.VerifyAgentIDCutoverRequest': VerifyAgentIDCutoverRequest$json,
  '.agents.v1.VerifyAgentIDCutoverResponse': VerifyAgentIDCutoverResponse$json,
  '.agents.v1.AgentIDCutoverFinding': AgentIDCutoverFinding$json,
  '.agents.v1.UpdateAgentConfigurationRequest':
      UpdateAgentConfigurationRequest$json,
  '.agents.v1.ContentFileAction': $4.ContentFileAction$json,
  '.agents.v1.UpdateAgentConfigurationResponse':
      UpdateAgentConfigurationResponse$json,
  '.agents.v1.RestoreAgentRequest': RestoreAgentRequest$json,
  '.agents.v1.RestoreAgentResponse': RestoreAgentResponse$json,
  '.agents.v1.GetAgentInvocationRequest': GetAgentInvocationRequest$json,
  '.agents.v1.GetAgentInvocationResponse': GetAgentInvocationResponse$json,
  '.agents.v1.GetAgentOperationRequest': GetAgentOperationRequest$json,
  '.agents.v1.GetAgentOperationResponse': GetAgentOperationResponse$json,
  '.agents.v1.ListAgentOperationsRequest': ListAgentOperationsRequest$json,
  '.agents.v1.ListAgentOperationsResponse': ListAgentOperationsResponse$json,
  '.agents.v1.RetryAgentOperationRequest': RetryAgentOperationRequest$json,
  '.agents.v1.RetryAgentOperationResponse': RetryAgentOperationResponse$json,
};

/// Descriptor for `AgentService`. Decode as a `google.protobuf.ServiceDescriptorProto`.
final $typed_data.Uint8List agentServiceDescriptor = $convert.base64Decode(
    'CgxBZ2VudFNlcnZpY2USSQoKTGlzdEFnZW50cxIcLmFnZW50cy52MS5MaXN0QWdlbnRzUmVxdW'
    'VzdBodLmFnZW50cy52MS5MaXN0QWdlbnRzUmVzcG9uc2USQwoIR2V0QWdlbnQSGi5hZ2VudHMu'
    'djEuR2V0QWdlbnRSZXF1ZXN0GhsuYWdlbnRzLnYxLkdldEFnZW50UmVzcG9uc2USTAoLQ3JlYX'
    'RlQWdlbnQSHS5hZ2VudHMudjEuQ3JlYXRlQWdlbnRSZXF1ZXN0Gh4uYWdlbnRzLnYxLkNyZWF0'
    'ZUFnZW50UmVzcG9uc2USTAoLVXBkYXRlQWdlbnQSHS5hZ2VudHMudjEuVXBkYXRlQWdlbnRSZX'
    'F1ZXN0Gh4uYWdlbnRzLnYxLlVwZGF0ZUFnZW50UmVzcG9uc2USTAoLRGVsZXRlQWdlbnQSHS5h'
    'Z2VudHMudjEuRGVsZXRlQWdlbnRSZXF1ZXN0Gh4uYWdlbnRzLnYxLkRlbGV0ZUFnZW50UmVzcG'
    '9uc2USTAoLSW52b2tlQWdlbnQSHS5hZ2VudHMudjEuSW52b2tlQWdlbnRSZXF1ZXN0Gh4uYWdl'
    'bnRzLnYxLkludm9rZUFnZW50UmVzcG9uc2USZwoUTGlzdEFnZW50SW52b2NhdGlvbnMSJi5hZ2'
    'VudHMudjEuTGlzdEFnZW50SW52b2NhdGlvbnNSZXF1ZXN0GicuYWdlbnRzLnYxLkxpc3RBZ2Vu'
    'dEludm9jYXRpb25zUmVzcG9uc2USTwoMUmVsb2FkQWdlbnRzEh4uYWdlbnRzLnYxLlJlbG9hZE'
    'FnZW50c1JlcXVlc3QaHy5hZ2VudHMudjEuUmVsb2FkQWdlbnRzUmVzcG9uc2USagoVR2V0QWdl'
    'bnRSdW50aW1lU3RhdHVzEicuYWdlbnRzLnYxLkdldEFnZW50UnVudGltZVN0YXR1c1JlcXVlc3'
    'QaKC5hZ2VudHMudjEuR2V0QWdlbnRSdW50aW1lU3RhdHVzUmVzcG9uc2UScwoYTGlzdEFnZW50'
    'UnVudGltZVN0YXR1c2VzEiouYWdlbnRzLnYxLkxpc3RBZ2VudFJ1bnRpbWVTdGF0dXNlc1JlcX'
    'Vlc3QaKy5hZ2VudHMudjEuTGlzdEFnZW50UnVudGltZVN0YXR1c2VzUmVzcG9uc2USagoVQ2Fu'
    'Y2VsQWdlbnRJbnZvY2F0aW9uEicuYWdlbnRzLnYxLkNhbmNlbEFnZW50SW52b2NhdGlvblJlcX'
    'Vlc3QaKC5hZ2VudHMudjEuQ2FuY2VsQWdlbnRJbnZvY2F0aW9uUmVzcG9uc2USTgoLU3RyZWFt'
    'QWdlbnQSHS5hZ2VudHMudjEuU3RyZWFtQWdlbnRSZXF1ZXN0Gh4uYWdlbnRzLnYxLlN0cmVhbU'
    'FnZW50UmVzcG9uc2UwARJXCg1Bc3NpZ25BZ2VudElEEh8uYWdlbnRzLnYxLkFzc2lnbkFnZW50'
    'SURSZXF1ZXN0GiAuYWdlbnRzLnYxLkFzc2lnbkFnZW50SURSZXNwb25zZSIDiAIBEm8KFUdldE'
    '1pZ3JhdGlvblJlYWRpbmVzcxInLmFnZW50cy52MS5HZXRNaWdyYXRpb25SZWFkaW5lc3NSZXF1'
    'ZXN0GiguYWdlbnRzLnYxLkdldE1pZ3JhdGlvblJlYWRpbmVzc1Jlc3BvbnNlIgOIAgESXQoPTW'
    'lncmF0ZUFnZW50c1YyEiEuYWdlbnRzLnYxLk1pZ3JhdGVBZ2VudHNWMlJlcXVlc3QaIi5hZ2Vu'
    'dHMudjEuTWlncmF0ZUFnZW50c1YyUmVzcG9uc2UiA4gCARJnChRWZXJpZnlBZ2VudElEQ3V0b3'
    'ZlchImLmFnZW50cy52MS5WZXJpZnlBZ2VudElEQ3V0b3ZlclJlcXVlc3QaJy5hZ2VudHMudjEu'
    'VmVyaWZ5QWdlbnRJREN1dG92ZXJSZXNwb25zZRJzChhVcGRhdGVBZ2VudENvbmZpZ3VyYXRpb2'
    '4SKi5hZ2VudHMudjEuVXBkYXRlQWdlbnRDb25maWd1cmF0aW9uUmVxdWVzdBorLmFnZW50cy52'
    'MS5VcGRhdGVBZ2VudENvbmZpZ3VyYXRpb25SZXNwb25zZRJPCgxSZXN0b3JlQWdlbnQSHi5hZ2'
    'VudHMudjEuUmVzdG9yZUFnZW50UmVxdWVzdBofLmFnZW50cy52MS5SZXN0b3JlQWdlbnRSZXNw'
    'b25zZRJhChJHZXRBZ2VudEludm9jYXRpb24SJC5hZ2VudHMudjEuR2V0QWdlbnRJbnZvY2F0aW'
    '9uUmVxdWVzdBolLmFnZW50cy52MS5HZXRBZ2VudEludm9jYXRpb25SZXNwb25zZRJeChFHZXRB'
    'Z2VudE9wZXJhdGlvbhIjLmFnZW50cy52MS5HZXRBZ2VudE9wZXJhdGlvblJlcXVlc3QaJC5hZ2'
    'VudHMudjEuR2V0QWdlbnRPcGVyYXRpb25SZXNwb25zZRJkChNMaXN0QWdlbnRPcGVyYXRpb25z'
    'EiUuYWdlbnRzLnYxLkxpc3RBZ2VudE9wZXJhdGlvbnNSZXF1ZXN0GiYuYWdlbnRzLnYxLkxpc3'
    'RBZ2VudE9wZXJhdGlvbnNSZXNwb25zZRJkChNSZXRyeUFnZW50T3BlcmF0aW9uEiUuYWdlbnRz'
    'LnYxLlJldHJ5QWdlbnRPcGVyYXRpb25SZXF1ZXN0GiYuYWdlbnRzLnYxLlJldHJ5QWdlbnRPcG'
    'VyYXRpb25SZXNwb25zZQ==');

const $core.Map<$core.String, $core.dynamic> MCPServerServiceBase$json = {
  '1': 'MCPServerService',
  '2': [
    {
      '1': 'ListMCPServers',
      '2': '.agents.v1.ListMCPServersRequest',
      '3': '.agents.v1.ListMCPServersResponse'
    },
    {
      '1': 'GetMCPServer',
      '2': '.agents.v1.GetMCPServerRequest',
      '3': '.agents.v1.GetMCPServerResponse'
    },
    {
      '1': 'CreateMCPServer',
      '2': '.agents.v1.CreateMCPServerRequest',
      '3': '.agents.v1.CreateMCPServerResponse'
    },
    {
      '1': 'UpdateMCPServer',
      '2': '.agents.v1.UpdateMCPServerRequest',
      '3': '.agents.v1.UpdateMCPServerResponse'
    },
    {
      '1': 'UpdateMCPServerKeepingCredentials',
      '2': '.agents.v1.UpdateMCPServerKeepingCredentialsRequest',
      '3': '.agents.v1.UpdateMCPServerKeepingCredentialsResponse'
    },
    {
      '1': 'DeleteMCPServer',
      '2': '.agents.v1.DeleteMCPServerRequest',
      '3': '.agents.v1.DeleteMCPServerResponse'
    },
    {
      '1': 'GetMCPServerStatus',
      '2': '.agents.v1.GetMCPServerStatusRequest',
      '3': '.agents.v1.GetMCPServerStatusResponse'
    },
    {
      '1': 'ListMCPTools',
      '2': '.agents.v1.ListMCPToolsRequest',
      '3': '.agents.v1.ListMCPToolsResponse'
    },
    {
      '1': 'StartMCPServerOAuth',
      '2': '.agents.v1.StartMCPServerOAuthRequest',
      '3': '.agents.v1.StartMCPServerOAuthResponse'
    },
    {
      '1': 'CompleteMCPServerOAuth',
      '2': '.agents.v1.CompleteMCPServerOAuthRequest',
      '3': '.agents.v1.CompleteMCPServerOAuthResponse'
    },
    {
      '1': 'GetMCPServerOAuthStatus',
      '2': '.agents.v1.GetMCPServerOAuthStatusRequest',
      '3': '.agents.v1.GetMCPServerOAuthStatusResponse'
    },
    {
      '1': 'DisconnectMCPServerOAuth',
      '2': '.agents.v1.DisconnectMCPServerOAuthRequest',
      '3': '.agents.v1.DisconnectMCPServerOAuthResponse'
    },
  ],
};

@$core.Deprecated('Use mCPServerServiceDescriptor instead')
const $core.Map<$core.String, $core.Map<$core.String, $core.dynamic>>
    MCPServerServiceBase$messageJson = {
  '.agents.v1.ListMCPServersRequest': ListMCPServersRequest$json,
  '.agents.v1.ListMCPServersResponse': ListMCPServersResponse$json,
  '.agents.v1.MCPServer': $0.MCPServer$json,
  '.agents.v1.MCPServer.HeadersEntry': $0.MCPServer_HeadersEntry$json,
  '.agents.v1.MCPServer.MetadataEntry': $0.MCPServer_MetadataEntry$json,
  '.agents.v1.MCPServerAuth': $0.MCPServerAuth$json,
  '.agents.v1.MCPServerOAuth2Config': $0.MCPServerOAuth2Config$json,
  '.agents.v1.GetMCPServerRequest': GetMCPServerRequest$json,
  '.agents.v1.GetMCPServerResponse': GetMCPServerResponse$json,
  '.agents.v1.CreateMCPServerRequest': CreateMCPServerRequest$json,
  '.agents.v1.CreateMCPServerResponse': CreateMCPServerResponse$json,
  '.agents.v1.UpdateMCPServerRequest': UpdateMCPServerRequest$json,
  '.agents.v1.UpdateMCPServerResponse': UpdateMCPServerResponse$json,
  '.agents.v1.UpdateMCPServerKeepingCredentialsRequest':
      UpdateMCPServerKeepingCredentialsRequest$json,
  '.agents.v1.UpdateMCPServerKeepingCredentialsResponse':
      UpdateMCPServerKeepingCredentialsResponse$json,
  '.agents.v1.DeleteMCPServerRequest': DeleteMCPServerRequest$json,
  '.agents.v1.DeleteMCPServerResponse': DeleteMCPServerResponse$json,
  '.agents.v1.GetMCPServerStatusRequest': $8.GetMCPServerStatusRequest$json,
  '.agents.v1.GetMCPServerStatusResponse': $8.GetMCPServerStatusResponse$json,
  '.agents.v1.MCPServerStatus': $8.MCPServerStatus$json,
  '.google.protobuf.Timestamp': $1.Timestamp$json,
  '.agents.v1.ListMCPToolsRequest': $8.ListMCPToolsRequest$json,
  '.agents.v1.ListMCPToolsResponse': $8.ListMCPToolsResponse$json,
  '.agents.v1.MCPTool': $8.MCPTool$json,
  '.agents.v1.ListMCPToolsResponse.ErrorsEntry':
      $8.ListMCPToolsResponse_ErrorsEntry$json,
  '.agents.v1.StartMCPServerOAuthRequest': StartMCPServerOAuthRequest$json,
  '.agents.v1.StartMCPServerOAuthResponse': StartMCPServerOAuthResponse$json,
  '.agents.v1.CompleteMCPServerOAuthRequest':
      CompleteMCPServerOAuthRequest$json,
  '.agents.v1.CompleteMCPServerOAuthResponse':
      CompleteMCPServerOAuthResponse$json,
  '.agents.v1.MCPOAuthConnectionStatus': MCPOAuthConnectionStatus$json,
  '.agents.v1.GetMCPServerOAuthStatusRequest':
      GetMCPServerOAuthStatusRequest$json,
  '.agents.v1.GetMCPServerOAuthStatusResponse':
      GetMCPServerOAuthStatusResponse$json,
  '.agents.v1.DisconnectMCPServerOAuthRequest':
      DisconnectMCPServerOAuthRequest$json,
  '.agents.v1.DisconnectMCPServerOAuthResponse':
      DisconnectMCPServerOAuthResponse$json,
};

/// Descriptor for `MCPServerService`. Decode as a `google.protobuf.ServiceDescriptorProto`.
final $typed_data.Uint8List mCPServerServiceDescriptor = $convert.base64Decode(
    'ChBNQ1BTZXJ2ZXJTZXJ2aWNlElUKDkxpc3RNQ1BTZXJ2ZXJzEiAuYWdlbnRzLnYxLkxpc3RNQ1'
    'BTZXJ2ZXJzUmVxdWVzdBohLmFnZW50cy52MS5MaXN0TUNQU2VydmVyc1Jlc3BvbnNlEk8KDEdl'
    'dE1DUFNlcnZlchIeLmFnZW50cy52MS5HZXRNQ1BTZXJ2ZXJSZXF1ZXN0Gh8uYWdlbnRzLnYxLk'
    'dldE1DUFNlcnZlclJlc3BvbnNlElgKD0NyZWF0ZU1DUFNlcnZlchIhLmFnZW50cy52MS5DcmVh'
    'dGVNQ1BTZXJ2ZXJSZXF1ZXN0GiIuYWdlbnRzLnYxLkNyZWF0ZU1DUFNlcnZlclJlc3BvbnNlEl'
    'gKD1VwZGF0ZU1DUFNlcnZlchIhLmFnZW50cy52MS5VcGRhdGVNQ1BTZXJ2ZXJSZXF1ZXN0GiIu'
    'YWdlbnRzLnYxLlVwZGF0ZU1DUFNlcnZlclJlc3BvbnNlEo4BCiFVcGRhdGVNQ1BTZXJ2ZXJLZW'
    'VwaW5nQ3JlZGVudGlhbHMSMy5hZ2VudHMudjEuVXBkYXRlTUNQU2VydmVyS2VlcGluZ0NyZWRl'
    'bnRpYWxzUmVxdWVzdBo0LmFnZW50cy52MS5VcGRhdGVNQ1BTZXJ2ZXJLZWVwaW5nQ3JlZGVudG'
    'lhbHNSZXNwb25zZRJYCg9EZWxldGVNQ1BTZXJ2ZXISIS5hZ2VudHMudjEuRGVsZXRlTUNQU2Vy'
    'dmVyUmVxdWVzdBoiLmFnZW50cy52MS5EZWxldGVNQ1BTZXJ2ZXJSZXNwb25zZRJhChJHZXRNQ1'
    'BTZXJ2ZXJTdGF0dXMSJC5hZ2VudHMudjEuR2V0TUNQU2VydmVyU3RhdHVzUmVxdWVzdBolLmFn'
    'ZW50cy52MS5HZXRNQ1BTZXJ2ZXJTdGF0dXNSZXNwb25zZRJPCgxMaXN0TUNQVG9vbHMSHi5hZ2'
    'VudHMudjEuTGlzdE1DUFRvb2xzUmVxdWVzdBofLmFnZW50cy52MS5MaXN0TUNQVG9vbHNSZXNw'
    'b25zZRJkChNTdGFydE1DUFNlcnZlck9BdXRoEiUuYWdlbnRzLnYxLlN0YXJ0TUNQU2VydmVyT0'
    'F1dGhSZXF1ZXN0GiYuYWdlbnRzLnYxLlN0YXJ0TUNQU2VydmVyT0F1dGhSZXNwb25zZRJtChZD'
    'b21wbGV0ZU1DUFNlcnZlck9BdXRoEiguYWdlbnRzLnYxLkNvbXBsZXRlTUNQU2VydmVyT0F1dG'
    'hSZXF1ZXN0GikuYWdlbnRzLnYxLkNvbXBsZXRlTUNQU2VydmVyT0F1dGhSZXNwb25zZRJwChdH'
    'ZXRNQ1BTZXJ2ZXJPQXV0aFN0YXR1cxIpLmFnZW50cy52MS5HZXRNQ1BTZXJ2ZXJPQXV0aFN0YX'
    'R1c1JlcXVlc3QaKi5hZ2VudHMudjEuR2V0TUNQU2VydmVyT0F1dGhTdGF0dXNSZXNwb25zZRJz'
    'ChhEaXNjb25uZWN0TUNQU2VydmVyT0F1dGgSKi5hZ2VudHMudjEuRGlzY29ubmVjdE1DUFNlcn'
    'Zlck9BdXRoUmVxdWVzdBorLmFnZW50cy52MS5EaXNjb25uZWN0TUNQU2VydmVyT0F1dGhSZXNw'
    'b25zZQ==');

const $core.Map<$core.String, $core.dynamic> GlobalMCPServerServiceBase$json = {
  '1': 'GlobalMCPServerService',
  '2': [
    {
      '1': 'ListGlobalMCPServers',
      '2': '.agents.v1.ListGlobalMCPServersRequest',
      '3': '.agents.v1.ListGlobalMCPServersResponse'
    },
    {
      '1': 'CreateGlobalMCPServer',
      '2': '.agents.v1.CreateGlobalMCPServerRequest',
      '3': '.agents.v1.CreateGlobalMCPServerResponse'
    },
    {
      '1': 'UpdateGlobalMCPServer',
      '2': '.agents.v1.UpdateGlobalMCPServerRequest',
      '3': '.agents.v1.UpdateGlobalMCPServerResponse'
    },
    {
      '1': 'DeleteGlobalMCPServer',
      '2': '.agents.v1.DeleteGlobalMCPServerRequest',
      '3': '.agents.v1.DeleteGlobalMCPServerResponse'
    },
    {
      '1': 'InstallGlobalMCPServer',
      '2': '.agents.v1.InstallGlobalMCPServerRequest',
      '3': '.agents.v1.InstallGlobalMCPServerResponse'
    },
  ],
};

@$core.Deprecated('Use globalMCPServerServiceDescriptor instead')
const $core.Map<$core.String, $core.Map<$core.String, $core.dynamic>>
    GlobalMCPServerServiceBase$messageJson = {
  '.agents.v1.ListGlobalMCPServersRequest': ListGlobalMCPServersRequest$json,
  '.agents.v1.ListGlobalMCPServersResponse': ListGlobalMCPServersResponse$json,
  '.agents.v1.MCPServer': $0.MCPServer$json,
  '.agents.v1.MCPServer.HeadersEntry': $0.MCPServer_HeadersEntry$json,
  '.agents.v1.MCPServer.MetadataEntry': $0.MCPServer_MetadataEntry$json,
  '.agents.v1.MCPServerAuth': $0.MCPServerAuth$json,
  '.agents.v1.MCPServerOAuth2Config': $0.MCPServerOAuth2Config$json,
  '.agents.v1.CreateGlobalMCPServerRequest': CreateGlobalMCPServerRequest$json,
  '.agents.v1.CreateGlobalMCPServerResponse':
      CreateGlobalMCPServerResponse$json,
  '.agents.v1.UpdateGlobalMCPServerRequest': UpdateGlobalMCPServerRequest$json,
  '.agents.v1.UpdateGlobalMCPServerResponse':
      UpdateGlobalMCPServerResponse$json,
  '.agents.v1.DeleteGlobalMCPServerRequest': DeleteGlobalMCPServerRequest$json,
  '.agents.v1.DeleteGlobalMCPServerResponse':
      DeleteGlobalMCPServerResponse$json,
  '.agents.v1.InstallGlobalMCPServerRequest':
      InstallGlobalMCPServerRequest$json,
  '.agents.v1.InstallGlobalMCPServerResponse':
      InstallGlobalMCPServerResponse$json,
};

/// Descriptor for `GlobalMCPServerService`. Decode as a `google.protobuf.ServiceDescriptorProto`.
final $typed_data.Uint8List globalMCPServerServiceDescriptor = $convert.base64Decode(
    'ChZHbG9iYWxNQ1BTZXJ2ZXJTZXJ2aWNlEmcKFExpc3RHbG9iYWxNQ1BTZXJ2ZXJzEiYuYWdlbn'
    'RzLnYxLkxpc3RHbG9iYWxNQ1BTZXJ2ZXJzUmVxdWVzdBonLmFnZW50cy52MS5MaXN0R2xvYmFs'
    'TUNQU2VydmVyc1Jlc3BvbnNlEmoKFUNyZWF0ZUdsb2JhbE1DUFNlcnZlchInLmFnZW50cy52MS'
    '5DcmVhdGVHbG9iYWxNQ1BTZXJ2ZXJSZXF1ZXN0GiguYWdlbnRzLnYxLkNyZWF0ZUdsb2JhbE1D'
    'UFNlcnZlclJlc3BvbnNlEmoKFVVwZGF0ZUdsb2JhbE1DUFNlcnZlchInLmFnZW50cy52MS5VcG'
    'RhdGVHbG9iYWxNQ1BTZXJ2ZXJSZXF1ZXN0GiguYWdlbnRzLnYxLlVwZGF0ZUdsb2JhbE1DUFNl'
    'cnZlclJlc3BvbnNlEmoKFURlbGV0ZUdsb2JhbE1DUFNlcnZlchInLmFnZW50cy52MS5EZWxldG'
    'VHbG9iYWxNQ1BTZXJ2ZXJSZXF1ZXN0GiguYWdlbnRzLnYxLkRlbGV0ZUdsb2JhbE1DUFNlcnZl'
    'clJlc3BvbnNlEm0KFkluc3RhbGxHbG9iYWxNQ1BTZXJ2ZXISKC5hZ2VudHMudjEuSW5zdGFsbE'
    'dsb2JhbE1DUFNlcnZlclJlcXVlc3QaKS5hZ2VudHMudjEuSW5zdGFsbEdsb2JhbE1DUFNlcnZl'
    'clJlc3BvbnNl');

const $core.Map<$core.String, $core.dynamic> ModelProviderServiceBase$json = {
  '1': 'ModelProviderService',
  '2': [
    {
      '1': 'ListModelProviders',
      '2': '.agents.v1.ListModelProvidersRequest',
      '3': '.agents.v1.ListModelProvidersResponse'
    },
    {
      '1': 'GetModelProvider',
      '2': '.agents.v1.GetModelProviderRequest',
      '3': '.agents.v1.GetModelProviderResponse'
    },
    {
      '1': 'CreateModelProvider',
      '2': '.agents.v1.CreateModelProviderRequest',
      '3': '.agents.v1.CreateModelProviderResponse'
    },
    {
      '1': 'UpdateModelProvider',
      '2': '.agents.v1.UpdateModelProviderRequest',
      '3': '.agents.v1.UpdateModelProviderResponse'
    },
    {
      '1': 'DeleteModelProvider',
      '2': '.agents.v1.DeleteModelProviderRequest',
      '3': '.agents.v1.DeleteModelProviderResponse'
    },
  ],
};

@$core.Deprecated('Use modelProviderServiceDescriptor instead')
const $core.Map<$core.String, $core.Map<$core.String, $core.dynamic>>
    ModelProviderServiceBase$messageJson = {
  '.agents.v1.ListModelProvidersRequest': ListModelProvidersRequest$json,
  '.agents.v1.ListModelProvidersResponse': ListModelProvidersResponse$json,
  '.agents.v1.ModelProvider': $0.ModelProvider$json,
  '.agents.v1.ModelConfig': $0.ModelConfig$json,
  '.agents.v1.GetModelProviderRequest': GetModelProviderRequest$json,
  '.agents.v1.GetModelProviderResponse': GetModelProviderResponse$json,
  '.agents.v1.CreateModelProviderRequest': CreateModelProviderRequest$json,
  '.agents.v1.CreateModelProviderResponse': CreateModelProviderResponse$json,
  '.agents.v1.UpdateModelProviderRequest': UpdateModelProviderRequest$json,
  '.agents.v1.UpdateModelProviderResponse': UpdateModelProviderResponse$json,
  '.agents.v1.DeleteModelProviderRequest': DeleteModelProviderRequest$json,
  '.agents.v1.DeleteModelProviderResponse': DeleteModelProviderResponse$json,
};

/// Descriptor for `ModelProviderService`. Decode as a `google.protobuf.ServiceDescriptorProto`.
final $typed_data.Uint8List modelProviderServiceDescriptor = $convert.base64Decode(
    'ChRNb2RlbFByb3ZpZGVyU2VydmljZRJhChJMaXN0TW9kZWxQcm92aWRlcnMSJC5hZ2VudHMudj'
    'EuTGlzdE1vZGVsUHJvdmlkZXJzUmVxdWVzdBolLmFnZW50cy52MS5MaXN0TW9kZWxQcm92aWRl'
    'cnNSZXNwb25zZRJbChBHZXRNb2RlbFByb3ZpZGVyEiIuYWdlbnRzLnYxLkdldE1vZGVsUHJvdm'
    'lkZXJSZXF1ZXN0GiMuYWdlbnRzLnYxLkdldE1vZGVsUHJvdmlkZXJSZXNwb25zZRJkChNDcmVh'
    'dGVNb2RlbFByb3ZpZGVyEiUuYWdlbnRzLnYxLkNyZWF0ZU1vZGVsUHJvdmlkZXJSZXF1ZXN0Gi'
    'YuYWdlbnRzLnYxLkNyZWF0ZU1vZGVsUHJvdmlkZXJSZXNwb25zZRJkChNVcGRhdGVNb2RlbFBy'
    'b3ZpZGVyEiUuYWdlbnRzLnYxLlVwZGF0ZU1vZGVsUHJvdmlkZXJSZXF1ZXN0GiYuYWdlbnRzLn'
    'YxLlVwZGF0ZU1vZGVsUHJvdmlkZXJSZXNwb25zZRJkChNEZWxldGVNb2RlbFByb3ZpZGVyEiUu'
    'YWdlbnRzLnYxLkRlbGV0ZU1vZGVsUHJvdmlkZXJSZXF1ZXN0GiYuYWdlbnRzLnYxLkRlbGV0ZU'
    '1vZGVsUHJvdmlkZXJSZXNwb25zZQ==');

const $core.Map<$core.String, $core.dynamic> NotifyGroupServiceBase$json = {
  '1': 'NotifyGroupService',
  '2': [
    {
      '1': 'ListNotifyGroups',
      '2': '.agents.v1.ListNotifyGroupsRequest',
      '3': '.agents.v1.ListNotifyGroupsResponse'
    },
    {
      '1': 'GetNotifyGroup',
      '2': '.agents.v1.GetNotifyGroupRequest',
      '3': '.agents.v1.GetNotifyGroupResponse'
    },
    {
      '1': 'CreateNotifyGroup',
      '2': '.agents.v1.CreateNotifyGroupRequest',
      '3': '.agents.v1.CreateNotifyGroupResponse'
    },
    {
      '1': 'UpdateNotifyGroup',
      '2': '.agents.v1.UpdateNotifyGroupRequest',
      '3': '.agents.v1.UpdateNotifyGroupResponse'
    },
    {
      '1': 'DeleteNotifyGroup',
      '2': '.agents.v1.DeleteNotifyGroupRequest',
      '3': '.agents.v1.DeleteNotifyGroupResponse'
    },
  ],
};

@$core.Deprecated('Use notifyGroupServiceDescriptor instead')
const $core.Map<$core.String, $core.Map<$core.String, $core.dynamic>>
    NotifyGroupServiceBase$messageJson = {
  '.agents.v1.ListNotifyGroupsRequest': ListNotifyGroupsRequest$json,
  '.agents.v1.ListNotifyGroupsResponse': ListNotifyGroupsResponse$json,
  '.agents.v1.NotifyGroup': $0.NotifyGroup$json,
  '.agents.v1.NotifyTarget': $0.NotifyTarget$json,
  '.agents.v1.TelegramNotifyTarget': $0.TelegramNotifyTarget$json,
  '.agents.v1.LarkNotifyTarget': $0.LarkNotifyTarget$json,
  '.agents.v1.DiscordNotifyTarget': $0.DiscordNotifyTarget$json,
  '.agents.v1.NotifyTarget.MetadataEntry': $0.NotifyTarget_MetadataEntry$json,
  '.agents.v1.NotifyGroup.MetadataEntry': $0.NotifyGroup_MetadataEntry$json,
  '.agents.v1.GetNotifyGroupRequest': GetNotifyGroupRequest$json,
  '.agents.v1.GetNotifyGroupResponse': GetNotifyGroupResponse$json,
  '.agents.v1.CreateNotifyGroupRequest': CreateNotifyGroupRequest$json,
  '.agents.v1.CreateNotifyGroupResponse': CreateNotifyGroupResponse$json,
  '.agents.v1.UpdateNotifyGroupRequest': UpdateNotifyGroupRequest$json,
  '.agents.v1.UpdateNotifyGroupResponse': UpdateNotifyGroupResponse$json,
  '.agents.v1.DeleteNotifyGroupRequest': DeleteNotifyGroupRequest$json,
  '.agents.v1.DeleteNotifyGroupResponse': DeleteNotifyGroupResponse$json,
};

/// Descriptor for `NotifyGroupService`. Decode as a `google.protobuf.ServiceDescriptorProto`.
final $typed_data.Uint8List notifyGroupServiceDescriptor = $convert.base64Decode(
    'ChJOb3RpZnlHcm91cFNlcnZpY2USWwoQTGlzdE5vdGlmeUdyb3VwcxIiLmFnZW50cy52MS5MaX'
    'N0Tm90aWZ5R3JvdXBzUmVxdWVzdBojLmFnZW50cy52MS5MaXN0Tm90aWZ5R3JvdXBzUmVzcG9u'
    'c2USVQoOR2V0Tm90aWZ5R3JvdXASIC5hZ2VudHMudjEuR2V0Tm90aWZ5R3JvdXBSZXF1ZXN0Gi'
    'EuYWdlbnRzLnYxLkdldE5vdGlmeUdyb3VwUmVzcG9uc2USXgoRQ3JlYXRlTm90aWZ5R3JvdXAS'
    'Iy5hZ2VudHMudjEuQ3JlYXRlTm90aWZ5R3JvdXBSZXF1ZXN0GiQuYWdlbnRzLnYxLkNyZWF0ZU'
    '5vdGlmeUdyb3VwUmVzcG9uc2USXgoRVXBkYXRlTm90aWZ5R3JvdXASIy5hZ2VudHMudjEuVXBk'
    'YXRlTm90aWZ5R3JvdXBSZXF1ZXN0GiQuYWdlbnRzLnYxLlVwZGF0ZU5vdGlmeUdyb3VwUmVzcG'
    '9uc2USXgoRRGVsZXRlTm90aWZ5R3JvdXASIy5hZ2VudHMudjEuRGVsZXRlTm90aWZ5R3JvdXBS'
    'ZXF1ZXN0GiQuYWdlbnRzLnYxLkRlbGV0ZU5vdGlmeUdyb3VwUmVzcG9uc2U=');

const $core.Map<$core.String, $core.dynamic> RemoteAgentServiceBase$json = {
  '1': 'RemoteAgentService',
  '2': [
    {
      '1': 'ListRemoteAgents',
      '2': '.agents.v1.ListRemoteAgentsRequest',
      '3': '.agents.v1.ListRemoteAgentsResponse'
    },
    {
      '1': 'GetRemoteAgent',
      '2': '.agents.v1.GetRemoteAgentRequest',
      '3': '.agents.v1.GetRemoteAgentResponse'
    },
    {
      '1': 'CreateRemoteAgent',
      '2': '.agents.v1.CreateRemoteAgentRequest',
      '3': '.agents.v1.CreateRemoteAgentResponse'
    },
    {
      '1': 'UpdateRemoteAgent',
      '2': '.agents.v1.UpdateRemoteAgentRequest',
      '3': '.agents.v1.UpdateRemoteAgentResponse'
    },
    {
      '1': 'DeleteRemoteAgent',
      '2': '.agents.v1.DeleteRemoteAgentRequest',
      '3': '.agents.v1.DeleteRemoteAgentResponse'
    },
    {
      '1': 'GetRemoteAgentStatus',
      '2': '.agents.v1.GetRemoteAgentStatusRequest',
      '3': '.agents.v1.GetRemoteAgentStatusResponse'
    },
  ],
};

@$core.Deprecated('Use remoteAgentServiceDescriptor instead')
const $core.Map<$core.String, $core.Map<$core.String, $core.dynamic>>
    RemoteAgentServiceBase$messageJson = {
  '.agents.v1.ListRemoteAgentsRequest': ListRemoteAgentsRequest$json,
  '.agents.v1.ListRemoteAgentsResponse': ListRemoteAgentsResponse$json,
  '.agents.v1.RemoteAgent': $0.RemoteAgent$json,
  '.agents.v1.GetRemoteAgentRequest': GetRemoteAgentRequest$json,
  '.agents.v1.GetRemoteAgentResponse': GetRemoteAgentResponse$json,
  '.agents.v1.CreateRemoteAgentRequest': CreateRemoteAgentRequest$json,
  '.agents.v1.CreateRemoteAgentResponse': CreateRemoteAgentResponse$json,
  '.agents.v1.UpdateRemoteAgentRequest': UpdateRemoteAgentRequest$json,
  '.agents.v1.UpdateRemoteAgentResponse': UpdateRemoteAgentResponse$json,
  '.agents.v1.DeleteRemoteAgentRequest': DeleteRemoteAgentRequest$json,
  '.agents.v1.DeleteRemoteAgentResponse': DeleteRemoteAgentResponse$json,
  '.agents.v1.GetRemoteAgentStatusRequest': $8.GetRemoteAgentStatusRequest$json,
  '.agents.v1.GetRemoteAgentStatusResponse':
      $8.GetRemoteAgentStatusResponse$json,
  '.agents.v1.RemoteAgentStatus': $8.RemoteAgentStatus$json,
  '.google.protobuf.Timestamp': $1.Timestamp$json,
};

/// Descriptor for `RemoteAgentService`. Decode as a `google.protobuf.ServiceDescriptorProto`.
final $typed_data.Uint8List remoteAgentServiceDescriptor = $convert.base64Decode(
    'ChJSZW1vdGVBZ2VudFNlcnZpY2USWwoQTGlzdFJlbW90ZUFnZW50cxIiLmFnZW50cy52MS5MaX'
    'N0UmVtb3RlQWdlbnRzUmVxdWVzdBojLmFnZW50cy52MS5MaXN0UmVtb3RlQWdlbnRzUmVzcG9u'
    'c2USVQoOR2V0UmVtb3RlQWdlbnQSIC5hZ2VudHMudjEuR2V0UmVtb3RlQWdlbnRSZXF1ZXN0Gi'
    'EuYWdlbnRzLnYxLkdldFJlbW90ZUFnZW50UmVzcG9uc2USXgoRQ3JlYXRlUmVtb3RlQWdlbnQS'
    'Iy5hZ2VudHMudjEuQ3JlYXRlUmVtb3RlQWdlbnRSZXF1ZXN0GiQuYWdlbnRzLnYxLkNyZWF0ZV'
    'JlbW90ZUFnZW50UmVzcG9uc2USXgoRVXBkYXRlUmVtb3RlQWdlbnQSIy5hZ2VudHMudjEuVXBk'
    'YXRlUmVtb3RlQWdlbnRSZXF1ZXN0GiQuYWdlbnRzLnYxLlVwZGF0ZVJlbW90ZUFnZW50UmVzcG'
    '9uc2USXgoRRGVsZXRlUmVtb3RlQWdlbnQSIy5hZ2VudHMudjEuRGVsZXRlUmVtb3RlQWdlbnRS'
    'ZXF1ZXN0GiQuYWdlbnRzLnYxLkRlbGV0ZVJlbW90ZUFnZW50UmVzcG9uc2USZwoUR2V0UmVtb3'
    'RlQWdlbnRTdGF0dXMSJi5hZ2VudHMudjEuR2V0UmVtb3RlQWdlbnRTdGF0dXNSZXF1ZXN0Gicu'
    'YWdlbnRzLnYxLkdldFJlbW90ZUFnZW50U3RhdHVzUmVzcG9uc2U=');

const $core.Map<$core.String, $core.dynamic> ChannelServiceBase$json = {
  '1': 'ChannelService',
  '2': [
    {
      '1': 'ListChannels',
      '2': '.agents.v1.ListChannelsRequest',
      '3': '.agents.v1.ListChannelsResponse'
    },
    {
      '1': 'GetChannel',
      '2': '.agents.v1.GetChannelRequest',
      '3': '.agents.v1.GetChannelResponse'
    },
    {
      '1': 'CreateChannel',
      '2': '.agents.v1.CreateChannelRequest',
      '3': '.agents.v1.CreateChannelResponse'
    },
    {
      '1': 'UpdateChannel',
      '2': '.agents.v1.UpdateChannelRequest',
      '3': '.agents.v1.UpdateChannelResponse'
    },
    {
      '1': 'DeleteChannel',
      '2': '.agents.v1.DeleteChannelRequest',
      '3': '.agents.v1.DeleteChannelResponse'
    },
    {
      '1': 'GetChannelStatus',
      '2': '.agents.v1.GetChannelStatusRequest',
      '3': '.agents.v1.GetChannelStatusResponse'
    },
    {
      '1': 'RestartChannel',
      '2': '.agents.v1.RestartChannelRequest',
      '3': '.agents.v1.RestartChannelResponse'
    },
    {
      '1': 'PauseChannel',
      '2': '.agents.v1.PauseChannelRequest',
      '3': '.agents.v1.PauseChannelResponse'
    },
    {
      '1': 'ResumeChannel',
      '2': '.agents.v1.ResumeChannelRequest',
      '3': '.agents.v1.ResumeChannelResponse'
    },
  ],
};

@$core.Deprecated('Use channelServiceDescriptor instead')
const $core.Map<$core.String, $core.Map<$core.String, $core.dynamic>>
    ChannelServiceBase$messageJson = {
  '.agents.v1.ListChannelsRequest': ListChannelsRequest$json,
  '.agents.v1.ListChannelsResponse': ListChannelsResponse$json,
  '.agents.v1.AgentChannel': $5.AgentChannel$json,
  '.agents.v1.AgentTrigger': $5.AgentTrigger$json,
  '.agents.v1.AgentDelivery': $5.AgentDelivery$json,
  '.agents.v1.AgentSessionBinding': $5.AgentSessionBinding$json,
  '.agents.v1.TelegramChannelConfig': $5.TelegramChannelConfig$json,
  '.agents.v1.DiscordChannelConfig': $5.DiscordChannelConfig$json,
  '.agents.v1.AgentChannel.MetadataEntry': $5.AgentChannel_MetadataEntry$json,
  '.agents.v1.GetChannelRequest': GetChannelRequest$json,
  '.agents.v1.GetChannelResponse': GetChannelResponse$json,
  '.agents.v1.CreateChannelRequest': CreateChannelRequest$json,
  '.agents.v1.CreateChannelResponse': CreateChannelResponse$json,
  '.agents.v1.UpdateChannelRequest': UpdateChannelRequest$json,
  '.agents.v1.UpdateChannelResponse': UpdateChannelResponse$json,
  '.agents.v1.DeleteChannelRequest': DeleteChannelRequest$json,
  '.agents.v1.DeleteChannelResponse': DeleteChannelResponse$json,
  '.agents.v1.GetChannelStatusRequest': $8.GetChannelStatusRequest$json,
  '.agents.v1.GetChannelStatusResponse': $8.GetChannelStatusResponse$json,
  '.agents.v1.ChannelStatus': $8.ChannelStatus$json,
  '.google.protobuf.Timestamp': $1.Timestamp$json,
  '.agents.v1.RestartChannelRequest': RestartChannelRequest$json,
  '.agents.v1.RestartChannelResponse': RestartChannelResponse$json,
  '.agents.v1.PauseChannelRequest': PauseChannelRequest$json,
  '.agents.v1.PauseChannelResponse': PauseChannelResponse$json,
  '.agents.v1.ResumeChannelRequest': ResumeChannelRequest$json,
  '.agents.v1.ResumeChannelResponse': ResumeChannelResponse$json,
};

/// Descriptor for `ChannelService`. Decode as a `google.protobuf.ServiceDescriptorProto`.
final $typed_data.Uint8List channelServiceDescriptor = $convert.base64Decode(
    'Cg5DaGFubmVsU2VydmljZRJPCgxMaXN0Q2hhbm5lbHMSHi5hZ2VudHMudjEuTGlzdENoYW5uZW'
    'xzUmVxdWVzdBofLmFnZW50cy52MS5MaXN0Q2hhbm5lbHNSZXNwb25zZRJJCgpHZXRDaGFubmVs'
    'EhwuYWdlbnRzLnYxLkdldENoYW5uZWxSZXF1ZXN0Gh0uYWdlbnRzLnYxLkdldENoYW5uZWxSZX'
    'Nwb25zZRJSCg1DcmVhdGVDaGFubmVsEh8uYWdlbnRzLnYxLkNyZWF0ZUNoYW5uZWxSZXF1ZXN0'
    'GiAuYWdlbnRzLnYxLkNyZWF0ZUNoYW5uZWxSZXNwb25zZRJSCg1VcGRhdGVDaGFubmVsEh8uYW'
    'dlbnRzLnYxLlVwZGF0ZUNoYW5uZWxSZXF1ZXN0GiAuYWdlbnRzLnYxLlVwZGF0ZUNoYW5uZWxS'
    'ZXNwb25zZRJSCg1EZWxldGVDaGFubmVsEh8uYWdlbnRzLnYxLkRlbGV0ZUNoYW5uZWxSZXF1ZX'
    'N0GiAuYWdlbnRzLnYxLkRlbGV0ZUNoYW5uZWxSZXNwb25zZRJbChBHZXRDaGFubmVsU3RhdHVz'
    'EiIuYWdlbnRzLnYxLkdldENoYW5uZWxTdGF0dXNSZXF1ZXN0GiMuYWdlbnRzLnYxLkdldENoYW'
    '5uZWxTdGF0dXNSZXNwb25zZRJVCg5SZXN0YXJ0Q2hhbm5lbBIgLmFnZW50cy52MS5SZXN0YXJ0'
    'Q2hhbm5lbFJlcXVlc3QaIS5hZ2VudHMudjEuUmVzdGFydENoYW5uZWxSZXNwb25zZRJPCgxQYX'
    'VzZUNoYW5uZWwSHi5hZ2VudHMudjEuUGF1c2VDaGFubmVsUmVxdWVzdBofLmFnZW50cy52MS5Q'
    'YXVzZUNoYW5uZWxSZXNwb25zZRJSCg1SZXN1bWVDaGFubmVsEh8uYWdlbnRzLnYxLlJlc3VtZU'
    'NoYW5uZWxSZXF1ZXN0GiAuYWdlbnRzLnYxLlJlc3VtZUNoYW5uZWxSZXNwb25zZQ==');

const $core.Map<$core.String, $core.dynamic> SessionServiceBase$json = {
  '1': 'SessionService',
  '2': [
    {
      '1': 'CreateSession',
      '2': '.agents.v1.CreateSessionRequest',
      '3': '.agents.v1.CreateSessionResponse'
    },
    {
      '1': 'GetSession',
      '2': '.agents.v1.GetSessionRequest',
      '3': '.agents.v1.GetSessionResponse'
    },
    {
      '1': 'ListSessions',
      '2': '.agents.v1.ListSessionsRequest',
      '3': '.agents.v1.ListSessionsResponse'
    },
    {
      '1': 'DeleteSession',
      '2': '.agents.v1.DeleteSessionRequest',
      '3': '.agents.v1.DeleteSessionResponse'
    },
    {
      '1': 'ReplySession',
      '2': '.agents.v1.ReplySessionRequest',
      '3': '.agents.v1.ReplySessionResponse'
    },
    {
      '1': 'UpdateSessionTitle',
      '2': '.agents.v1.UpdateSessionTitleRequest',
      '3': '.agents.v1.UpdateSessionTitleResponse'
    },
    {
      '1': 'GenerateSessionTitle',
      '2': '.agents.v1.GenerateSessionTitleRequest',
      '3': '.agents.v1.GenerateSessionTitleResponse'
    },
    {
      '1': 'MarkSessionRead',
      '2': '.agents.v1.MarkSessionReadRequest',
      '3': '.agents.v1.MarkSessionReadResponse'
    },
  ],
};

@$core.Deprecated('Use sessionServiceDescriptor instead')
const $core.Map<$core.String, $core.Map<$core.String, $core.dynamic>>
    SessionServiceBase$messageJson = {
  '.agents.v1.CreateSessionRequest': CreateSessionRequest$json,
  '.google.protobuf.Struct': $6.Struct$json,
  '.google.protobuf.Struct.FieldsEntry': $6.Struct_FieldsEntry$json,
  '.google.protobuf.Value': $6.Value$json,
  '.google.protobuf.ListValue': $6.ListValue$json,
  '.agents.v1.CreateSessionResponse': CreateSessionResponse$json,
  '.agents.v1.SessionInfo': SessionInfo$json,
  '.google.protobuf.Timestamp': $1.Timestamp$json,
  '.agents.v1.GetSessionRequest': GetSessionRequest$json,
  '.agents.v1.GetSessionResponse': GetSessionResponse$json,
  '.agents.v1.SessionDetail': SessionDetail$json,
  '.agents.v1.SessionEvent': SessionEvent$json,
  '.google.protobuf.Duration': $7.Duration$json,
  '.agents.v1.ListSessionsRequest': ListSessionsRequest$json,
  '.agents.v1.ListSessionsResponse': ListSessionsResponse$json,
  '.agents.v1.DeleteSessionRequest': DeleteSessionRequest$json,
  '.agents.v1.DeleteSessionResponse': DeleteSessionResponse$json,
  '.agents.v1.ReplySessionRequest': ReplySessionRequest$json,
  '.agents.v1.InputPart': $2.InputPart$json,
  '.agents.v1.InlineData': $2.InlineData$json,
  '.agents.v1.ReplySessionResponse': ReplySessionResponse$json,
  '.agents.v1.UpdateSessionTitleRequest': UpdateSessionTitleRequest$json,
  '.agents.v1.UpdateSessionTitleResponse': UpdateSessionTitleResponse$json,
  '.agents.v1.GenerateSessionTitleRequest': GenerateSessionTitleRequest$json,
  '.agents.v1.GenerateSessionTitleResponse': GenerateSessionTitleResponse$json,
  '.agents.v1.MarkSessionReadRequest': MarkSessionReadRequest$json,
  '.agents.v1.MarkSessionReadResponse': MarkSessionReadResponse$json,
};

/// Descriptor for `SessionService`. Decode as a `google.protobuf.ServiceDescriptorProto`.
final $typed_data.Uint8List sessionServiceDescriptor = $convert.base64Decode(
    'Cg5TZXNzaW9uU2VydmljZRJSCg1DcmVhdGVTZXNzaW9uEh8uYWdlbnRzLnYxLkNyZWF0ZVNlc3'
    'Npb25SZXF1ZXN0GiAuYWdlbnRzLnYxLkNyZWF0ZVNlc3Npb25SZXNwb25zZRJJCgpHZXRTZXNz'
    'aW9uEhwuYWdlbnRzLnYxLkdldFNlc3Npb25SZXF1ZXN0Gh0uYWdlbnRzLnYxLkdldFNlc3Npb2'
    '5SZXNwb25zZRJPCgxMaXN0U2Vzc2lvbnMSHi5hZ2VudHMudjEuTGlzdFNlc3Npb25zUmVxdWVz'
    'dBofLmFnZW50cy52MS5MaXN0U2Vzc2lvbnNSZXNwb25zZRJSCg1EZWxldGVTZXNzaW9uEh8uYW'
    'dlbnRzLnYxLkRlbGV0ZVNlc3Npb25SZXF1ZXN0GiAuYWdlbnRzLnYxLkRlbGV0ZVNlc3Npb25S'
    'ZXNwb25zZRJPCgxSZXBseVNlc3Npb24SHi5hZ2VudHMudjEuUmVwbHlTZXNzaW9uUmVxdWVzdB'
    'ofLmFnZW50cy52MS5SZXBseVNlc3Npb25SZXNwb25zZRJhChJVcGRhdGVTZXNzaW9uVGl0bGUS'
    'JC5hZ2VudHMudjEuVXBkYXRlU2Vzc2lvblRpdGxlUmVxdWVzdBolLmFnZW50cy52MS5VcGRhdG'
    'VTZXNzaW9uVGl0bGVSZXNwb25zZRJnChRHZW5lcmF0ZVNlc3Npb25UaXRsZRImLmFnZW50cy52'
    'MS5HZW5lcmF0ZVNlc3Npb25UaXRsZVJlcXVlc3QaJy5hZ2VudHMudjEuR2VuZXJhdGVTZXNzaW'
    '9uVGl0bGVSZXNwb25zZRJYCg9NYXJrU2Vzc2lvblJlYWQSIS5hZ2VudHMudjEuTWFya1Nlc3Np'
    'b25SZWFkUmVxdWVzdBoiLmFnZW50cy52MS5NYXJrU2Vzc2lvblJlYWRSZXNwb25zZQ==');
