// This is a generated file - do not edit.
//
// Generated from agents/v1/workspace_memory.proto.

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

@$core.Deprecated('Use workspaceMemoryScopeDescriptor instead')
const WorkspaceMemoryScope$json = {
  '1': 'WorkspaceMemoryScope',
  '2': [
    {'1': 'WORKSPACE_MEMORY_SCOPE_UNSPECIFIED', '2': 0},
    {'1': 'WORKSPACE_MEMORY_SCOPE_WORKSPACE', '2': 1},
    {'1': 'WORKSPACE_MEMORY_SCOPE_AGENT', '2': 2},
  ],
};

/// Descriptor for `WorkspaceMemoryScope`. Decode as a `google.protobuf.EnumDescriptorProto`.
final $typed_data.Uint8List workspaceMemoryScopeDescriptor = $convert.base64Decode(
    'ChRXb3Jrc3BhY2VNZW1vcnlTY29wZRImCiJXT1JLU1BBQ0VfTUVNT1JZX1NDT1BFX1VOU1BFQ0'
    'lGSUVEEAASJAogV09SS1NQQUNFX01FTU9SWV9TQ09QRV9XT1JLU1BBQ0UQARIgChxXT1JLU1BB'
    'Q0VfTUVNT1JZX1NDT1BFX0FHRU5UEAI=');

@$core.Deprecated('Use workspaceMemoryConfigDescriptor instead')
const WorkspaceMemoryConfig$json = {
  '1': 'WorkspaceMemoryConfig',
  '2': [
    {'1': 'workspace_id', '3': 1, '4': 1, '5': 9, '10': 'workspaceId'},
    {'1': 'base_url', '3': 2, '4': 1, '5': 9, '10': 'baseUrl'},
    {'1': 'enabled', '3': 3, '4': 1, '5': 8, '10': 'enabled'},
    {'1': 'credential_set', '3': 4, '4': 1, '5': 8, '10': 'credentialSet'},
    {
      '1': 'credential_updated_at',
      '3': 5,
      '4': 1,
      '5': 11,
      '6': '.google.protobuf.Timestamp',
      '10': 'credentialUpdatedAt'
    },
    {
      '1': 'created_at',
      '3': 6,
      '4': 1,
      '5': 11,
      '6': '.google.protobuf.Timestamp',
      '10': 'createdAt'
    },
    {
      '1': 'updated_at',
      '3': 7,
      '4': 1,
      '5': 11,
      '6': '.google.protobuf.Timestamp',
      '10': 'updatedAt'
    },
  ],
};

/// Descriptor for `WorkspaceMemoryConfig`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List workspaceMemoryConfigDescriptor = $convert.base64Decode(
    'ChVXb3Jrc3BhY2VNZW1vcnlDb25maWcSIQoMd29ya3NwYWNlX2lkGAEgASgJUgt3b3Jrc3BhY2'
    'VJZBIZCghiYXNlX3VybBgCIAEoCVIHYmFzZVVybBIYCgdlbmFibGVkGAMgASgIUgdlbmFibGVk'
    'EiUKDmNyZWRlbnRpYWxfc2V0GAQgASgIUg1jcmVkZW50aWFsU2V0Ek4KFWNyZWRlbnRpYWxfdX'
    'BkYXRlZF9hdBgFIAEoCzIaLmdvb2dsZS5wcm90b2J1Zi5UaW1lc3RhbXBSE2NyZWRlbnRpYWxV'
    'cGRhdGVkQXQSOQoKY3JlYXRlZF9hdBgGIAEoCzIaLmdvb2dsZS5wcm90b2J1Zi5UaW1lc3RhbX'
    'BSCWNyZWF0ZWRBdBI5Cgp1cGRhdGVkX2F0GAcgASgLMhouZ29vZ2xlLnByb3RvYnVmLlRpbWVz'
    'dGFtcFIJdXBkYXRlZEF0');

@$core.Deprecated('Use getWorkspaceMemoryConfigRequestDescriptor instead')
const GetWorkspaceMemoryConfigRequest$json = {
  '1': 'GetWorkspaceMemoryConfigRequest',
};

/// Descriptor for `GetWorkspaceMemoryConfigRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getWorkspaceMemoryConfigRequestDescriptor =
    $convert.base64Decode('Ch9HZXRXb3Jrc3BhY2VNZW1vcnlDb25maWdSZXF1ZXN0');

@$core.Deprecated('Use getWorkspaceMemoryConfigResponseDescriptor instead')
const GetWorkspaceMemoryConfigResponse$json = {
  '1': 'GetWorkspaceMemoryConfigResponse',
  '2': [
    {
      '1': 'config',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.WorkspaceMemoryConfig',
      '10': 'config'
    },
  ],
};

/// Descriptor for `GetWorkspaceMemoryConfigResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getWorkspaceMemoryConfigResponseDescriptor =
    $convert.base64Decode(
        'CiBHZXRXb3Jrc3BhY2VNZW1vcnlDb25maWdSZXNwb25zZRI4CgZjb25maWcYASABKAsyIC5hZ2'
        'VudHMudjEuV29ya3NwYWNlTWVtb3J5Q29uZmlnUgZjb25maWc=');

@$core.Deprecated('Use putWorkspaceMemoryConfigRequestDescriptor instead')
const PutWorkspaceMemoryConfigRequest$json = {
  '1': 'PutWorkspaceMemoryConfigRequest',
  '2': [
    {'1': 'base_url', '3': 1, '4': 1, '5': 9, '10': 'baseUrl'},
    {'1': 'enabled', '3': 2, '4': 1, '5': 8, '10': 'enabled'},
    {
      '1': 'api_key',
      '3': 3,
      '4': 1,
      '5': 9,
      '9': 0,
      '10': 'apiKey',
      '17': true
    },
  ],
  '8': [
    {'1': '_api_key'},
  ],
};

/// Descriptor for `PutWorkspaceMemoryConfigRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List putWorkspaceMemoryConfigRequestDescriptor =
    $convert.base64Decode(
        'Ch9QdXRXb3Jrc3BhY2VNZW1vcnlDb25maWdSZXF1ZXN0EhkKCGJhc2VfdXJsGAEgASgJUgdiYX'
        'NlVXJsEhgKB2VuYWJsZWQYAiABKAhSB2VuYWJsZWQSHAoHYXBpX2tleRgDIAEoCUgAUgZhcGlL'
        'ZXmIAQFCCgoIX2FwaV9rZXk=');

@$core.Deprecated('Use putWorkspaceMemoryConfigResponseDescriptor instead')
const PutWorkspaceMemoryConfigResponse$json = {
  '1': 'PutWorkspaceMemoryConfigResponse',
  '2': [
    {
      '1': 'config',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.WorkspaceMemoryConfig',
      '10': 'config'
    },
    {'1': 'warning', '3': 2, '4': 1, '5': 9, '10': 'warning'},
  ],
};

/// Descriptor for `PutWorkspaceMemoryConfigResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List putWorkspaceMemoryConfigResponseDescriptor =
    $convert.base64Decode(
        'CiBQdXRXb3Jrc3BhY2VNZW1vcnlDb25maWdSZXNwb25zZRI4CgZjb25maWcYASABKAsyIC5hZ2'
        'VudHMudjEuV29ya3NwYWNlTWVtb3J5Q29uZmlnUgZjb25maWcSGAoHd2FybmluZxgCIAEoCVIH'
        'd2FybmluZw==');

@$core.Deprecated('Use deleteWorkspaceMemoryConfigRequestDescriptor instead')
const DeleteWorkspaceMemoryConfigRequest$json = {
  '1': 'DeleteWorkspaceMemoryConfigRequest',
};

/// Descriptor for `DeleteWorkspaceMemoryConfigRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List deleteWorkspaceMemoryConfigRequestDescriptor =
    $convert.base64Decode('CiJEZWxldGVXb3Jrc3BhY2VNZW1vcnlDb25maWdSZXF1ZXN0');

@$core.Deprecated('Use deleteWorkspaceMemoryConfigResponseDescriptor instead')
const DeleteWorkspaceMemoryConfigResponse$json = {
  '1': 'DeleteWorkspaceMemoryConfigResponse',
};

/// Descriptor for `DeleteWorkspaceMemoryConfigResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List deleteWorkspaceMemoryConfigResponseDescriptor =
    $convert
        .base64Decode('CiNEZWxldGVXb3Jrc3BhY2VNZW1vcnlDb25maWdSZXNwb25zZQ==');

@$core.Deprecated('Use testWorkspaceMemoryConnectionRequestDescriptor instead')
const TestWorkspaceMemoryConnectionRequest$json = {
  '1': 'TestWorkspaceMemoryConnectionRequest',
};

/// Descriptor for `TestWorkspaceMemoryConnectionRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List testWorkspaceMemoryConnectionRequestDescriptor =
    $convert
        .base64Decode('CiRUZXN0V29ya3NwYWNlTWVtb3J5Q29ubmVjdGlvblJlcXVlc3Q=');

@$core.Deprecated('Use testWorkspaceMemoryConnectionResponseDescriptor instead')
const TestWorkspaceMemoryConnectionResponse$json = {
  '1': 'TestWorkspaceMemoryConnectionResponse',
  '2': [
    {'1': 'ok', '3': 1, '4': 1, '5': 8, '10': 'ok'},
    {'1': 'error', '3': 2, '4': 1, '5': 9, '10': 'error'},
  ],
};

/// Descriptor for `TestWorkspaceMemoryConnectionResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List testWorkspaceMemoryConnectionResponseDescriptor =
    $convert.base64Decode(
        'CiVUZXN0V29ya3NwYWNlTWVtb3J5Q29ubmVjdGlvblJlc3BvbnNlEg4KAm9rGAEgASgIUgJvax'
        'IUCgVlcnJvchgCIAEoCVIFZXJyb3I=');

@$core.Deprecated('Use workspaceMemoryDescriptor instead')
const WorkspaceMemory$json = {
  '1': 'WorkspaceMemory',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
    {'1': 'memory', '3': 2, '4': 1, '5': 9, '10': 'memory'},
    {
      '1': 'scope',
      '3': 3,
      '4': 1,
      '5': 14,
      '6': '.agents.v1.WorkspaceMemoryScope',
      '10': 'scope'
    },
    {'1': 'agent_id', '3': 4, '4': 1, '5': 9, '10': 'agentId'},
    {'1': 'channel', '3': 5, '4': 1, '5': 9, '10': 'channel'},
    {'1': 'principal', '3': 6, '4': 1, '5': 9, '10': 'principal'},
    {'1': 'session_id', '3': 7, '4': 1, '5': 9, '10': 'sessionId'},
    {
      '1': 'created_at',
      '3': 8,
      '4': 1,
      '5': 11,
      '6': '.google.protobuf.Timestamp',
      '10': 'createdAt'
    },
    {
      '1': 'updated_at',
      '3': 9,
      '4': 1,
      '5': 11,
      '6': '.google.protobuf.Timestamp',
      '10': 'updatedAt'
    },
    {'1': 'score', '3': 10, '4': 1, '5': 1, '10': 'score'},
  ],
};

/// Descriptor for `WorkspaceMemory`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List workspaceMemoryDescriptor = $convert.base64Decode(
    'Cg9Xb3Jrc3BhY2VNZW1vcnkSDgoCaWQYASABKAlSAmlkEhYKBm1lbW9yeRgCIAEoCVIGbWVtb3'
    'J5EjUKBXNjb3BlGAMgASgOMh8uYWdlbnRzLnYxLldvcmtzcGFjZU1lbW9yeVNjb3BlUgVzY29w'
    'ZRIZCghhZ2VudF9pZBgEIAEoCVIHYWdlbnRJZBIYCgdjaGFubmVsGAUgASgJUgdjaGFubmVsEh'
    'wKCXByaW5jaXBhbBgGIAEoCVIJcHJpbmNpcGFsEh0KCnNlc3Npb25faWQYByABKAlSCXNlc3Np'
    'b25JZBI5CgpjcmVhdGVkX2F0GAggASgLMhouZ29vZ2xlLnByb3RvYnVmLlRpbWVzdGFtcFIJY3'
    'JlYXRlZEF0EjkKCnVwZGF0ZWRfYXQYCSABKAsyGi5nb29nbGUucHJvdG9idWYuVGltZXN0YW1w'
    'Ugl1cGRhdGVkQXQSFAoFc2NvcmUYCiABKAFSBXNjb3Jl');

@$core.Deprecated('Use listWorkspaceMemoriesRequestDescriptor instead')
const ListWorkspaceMemoriesRequest$json = {
  '1': 'ListWorkspaceMemoriesRequest',
  '2': [
    {
      '1': 'scope',
      '3': 1,
      '4': 1,
      '5': 14,
      '6': '.agents.v1.WorkspaceMemoryScope',
      '10': 'scope'
    },
    {'1': 'agent_id', '3': 2, '4': 1, '5': 9, '10': 'agentId'},
  ],
};

/// Descriptor for `ListWorkspaceMemoriesRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listWorkspaceMemoriesRequestDescriptor =
    $convert.base64Decode(
        'ChxMaXN0V29ya3NwYWNlTWVtb3JpZXNSZXF1ZXN0EjUKBXNjb3BlGAEgASgOMh8uYWdlbnRzLn'
        'YxLldvcmtzcGFjZU1lbW9yeVNjb3BlUgVzY29wZRIZCghhZ2VudF9pZBgCIAEoCVIHYWdlbnRJ'
        'ZA==');

@$core.Deprecated('Use listWorkspaceMemoriesResponseDescriptor instead')
const ListWorkspaceMemoriesResponse$json = {
  '1': 'ListWorkspaceMemoriesResponse',
  '2': [
    {
      '1': 'memories',
      '3': 1,
      '4': 3,
      '5': 11,
      '6': '.agents.v1.WorkspaceMemory',
      '10': 'memories'
    },
    {'1': 'truncated', '3': 2, '4': 1, '5': 8, '10': 'truncated'},
    {'1': 'limit', '3': 3, '4': 1, '5': 5, '10': 'limit'},
  ],
};

/// Descriptor for `ListWorkspaceMemoriesResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listWorkspaceMemoriesResponseDescriptor =
    $convert.base64Decode(
        'Ch1MaXN0V29ya3NwYWNlTWVtb3JpZXNSZXNwb25zZRI2CghtZW1vcmllcxgBIAMoCzIaLmFnZW'
        '50cy52MS5Xb3Jrc3BhY2VNZW1vcnlSCG1lbW9yaWVzEhwKCXRydW5jYXRlZBgCIAEoCFIJdHJ1'
        'bmNhdGVkEhQKBWxpbWl0GAMgASgFUgVsaW1pdA==');

@$core.Deprecated('Use searchWorkspaceMemoriesRequestDescriptor instead')
const SearchWorkspaceMemoriesRequest$json = {
  '1': 'SearchWorkspaceMemoriesRequest',
  '2': [
    {'1': 'query', '3': 1, '4': 1, '5': 9, '10': 'query'},
    {
      '1': 'scope',
      '3': 2,
      '4': 1,
      '5': 14,
      '6': '.agents.v1.WorkspaceMemoryScope',
      '10': 'scope'
    },
    {'1': 'agent_id', '3': 3, '4': 1, '5': 9, '10': 'agentId'},
    {'1': 'top_k', '3': 4, '4': 1, '5': 5, '10': 'topK'},
  ],
};

/// Descriptor for `SearchWorkspaceMemoriesRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List searchWorkspaceMemoriesRequestDescriptor =
    $convert.base64Decode(
        'Ch5TZWFyY2hXb3Jrc3BhY2VNZW1vcmllc1JlcXVlc3QSFAoFcXVlcnkYASABKAlSBXF1ZXJ5Ej'
        'UKBXNjb3BlGAIgASgOMh8uYWdlbnRzLnYxLldvcmtzcGFjZU1lbW9yeVNjb3BlUgVzY29wZRIZ'
        'CghhZ2VudF9pZBgDIAEoCVIHYWdlbnRJZBITCgV0b3BfaxgEIAEoBVIEdG9wSw==');

@$core.Deprecated('Use searchWorkspaceMemoriesResponseDescriptor instead')
const SearchWorkspaceMemoriesResponse$json = {
  '1': 'SearchWorkspaceMemoriesResponse',
  '2': [
    {
      '1': 'memories',
      '3': 1,
      '4': 3,
      '5': 11,
      '6': '.agents.v1.WorkspaceMemory',
      '10': 'memories'
    },
  ],
};

/// Descriptor for `SearchWorkspaceMemoriesResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List searchWorkspaceMemoriesResponseDescriptor =
    $convert.base64Decode(
        'Ch9TZWFyY2hXb3Jrc3BhY2VNZW1vcmllc1Jlc3BvbnNlEjYKCG1lbW9yaWVzGAEgAygLMhouYW'
        'dlbnRzLnYxLldvcmtzcGFjZU1lbW9yeVIIbWVtb3JpZXM=');

@$core.Deprecated('Use deleteWorkspaceMemoryRequestDescriptor instead')
const DeleteWorkspaceMemoryRequest$json = {
  '1': 'DeleteWorkspaceMemoryRequest',
  '2': [
    {'1': 'memory_id', '3': 1, '4': 1, '5': 9, '10': 'memoryId'},
  ],
};

/// Descriptor for `DeleteWorkspaceMemoryRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List deleteWorkspaceMemoryRequestDescriptor =
    $convert.base64Decode(
        'ChxEZWxldGVXb3Jrc3BhY2VNZW1vcnlSZXF1ZXN0EhsKCW1lbW9yeV9pZBgBIAEoCVIIbWVtb3'
        'J5SWQ=');

@$core.Deprecated('Use deleteWorkspaceMemoryResponseDescriptor instead')
const DeleteWorkspaceMemoryResponse$json = {
  '1': 'DeleteWorkspaceMemoryResponse',
};

/// Descriptor for `DeleteWorkspaceMemoryResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List deleteWorkspaceMemoryResponseDescriptor =
    $convert.base64Decode('Ch1EZWxldGVXb3Jrc3BhY2VNZW1vcnlSZXNwb25zZQ==');

const $core.Map<$core.String, $core.dynamic>
    WorkspaceMemoryConfigServiceBase$json = {
  '1': 'WorkspaceMemoryConfigService',
  '2': [
    {
      '1': 'GetWorkspaceMemoryConfig',
      '2': '.agents.v1.GetWorkspaceMemoryConfigRequest',
      '3': '.agents.v1.GetWorkspaceMemoryConfigResponse'
    },
    {
      '1': 'PutWorkspaceMemoryConfig',
      '2': '.agents.v1.PutWorkspaceMemoryConfigRequest',
      '3': '.agents.v1.PutWorkspaceMemoryConfigResponse'
    },
    {
      '1': 'DeleteWorkspaceMemoryConfig',
      '2': '.agents.v1.DeleteWorkspaceMemoryConfigRequest',
      '3': '.agents.v1.DeleteWorkspaceMemoryConfigResponse'
    },
    {
      '1': 'TestWorkspaceMemoryConnection',
      '2': '.agents.v1.TestWorkspaceMemoryConnectionRequest',
      '3': '.agents.v1.TestWorkspaceMemoryConnectionResponse'
    },
  ],
};

@$core.Deprecated('Use workspaceMemoryConfigServiceDescriptor instead')
const $core.Map<$core.String, $core.Map<$core.String, $core.dynamic>>
    WorkspaceMemoryConfigServiceBase$messageJson = {
  '.agents.v1.GetWorkspaceMemoryConfigRequest':
      GetWorkspaceMemoryConfigRequest$json,
  '.agents.v1.GetWorkspaceMemoryConfigResponse':
      GetWorkspaceMemoryConfigResponse$json,
  '.agents.v1.WorkspaceMemoryConfig': WorkspaceMemoryConfig$json,
  '.google.protobuf.Timestamp': $0.Timestamp$json,
  '.agents.v1.PutWorkspaceMemoryConfigRequest':
      PutWorkspaceMemoryConfigRequest$json,
  '.agents.v1.PutWorkspaceMemoryConfigResponse':
      PutWorkspaceMemoryConfigResponse$json,
  '.agents.v1.DeleteWorkspaceMemoryConfigRequest':
      DeleteWorkspaceMemoryConfigRequest$json,
  '.agents.v1.DeleteWorkspaceMemoryConfigResponse':
      DeleteWorkspaceMemoryConfigResponse$json,
  '.agents.v1.TestWorkspaceMemoryConnectionRequest':
      TestWorkspaceMemoryConnectionRequest$json,
  '.agents.v1.TestWorkspaceMemoryConnectionResponse':
      TestWorkspaceMemoryConnectionResponse$json,
};

/// Descriptor for `WorkspaceMemoryConfigService`. Decode as a `google.protobuf.ServiceDescriptorProto`.
final $typed_data.Uint8List workspaceMemoryConfigServiceDescriptor = $convert.base64Decode(
    'ChxXb3Jrc3BhY2VNZW1vcnlDb25maWdTZXJ2aWNlEnMKGEdldFdvcmtzcGFjZU1lbW9yeUNvbm'
    'ZpZxIqLmFnZW50cy52MS5HZXRXb3Jrc3BhY2VNZW1vcnlDb25maWdSZXF1ZXN0GisuYWdlbnRz'
    'LnYxLkdldFdvcmtzcGFjZU1lbW9yeUNvbmZpZ1Jlc3BvbnNlEnMKGFB1dFdvcmtzcGFjZU1lbW'
    '9yeUNvbmZpZxIqLmFnZW50cy52MS5QdXRXb3Jrc3BhY2VNZW1vcnlDb25maWdSZXF1ZXN0Gisu'
    'YWdlbnRzLnYxLlB1dFdvcmtzcGFjZU1lbW9yeUNvbmZpZ1Jlc3BvbnNlEnwKG0RlbGV0ZVdvcm'
    'tzcGFjZU1lbW9yeUNvbmZpZxItLmFnZW50cy52MS5EZWxldGVXb3Jrc3BhY2VNZW1vcnlDb25m'
    'aWdSZXF1ZXN0Gi4uYWdlbnRzLnYxLkRlbGV0ZVdvcmtzcGFjZU1lbW9yeUNvbmZpZ1Jlc3Bvbn'
    'NlEoIBCh1UZXN0V29ya3NwYWNlTWVtb3J5Q29ubmVjdGlvbhIvLmFnZW50cy52MS5UZXN0V29y'
    'a3NwYWNlTWVtb3J5Q29ubmVjdGlvblJlcXVlc3QaMC5hZ2VudHMudjEuVGVzdFdvcmtzcGFjZU'
    '1lbW9yeUNvbm5lY3Rpb25SZXNwb25zZQ==');

const $core.Map<$core.String, $core.dynamic> WorkspaceMemoryServiceBase$json = {
  '1': 'WorkspaceMemoryService',
  '2': [
    {
      '1': 'ListWorkspaceMemories',
      '2': '.agents.v1.ListWorkspaceMemoriesRequest',
      '3': '.agents.v1.ListWorkspaceMemoriesResponse'
    },
    {
      '1': 'SearchWorkspaceMemories',
      '2': '.agents.v1.SearchWorkspaceMemoriesRequest',
      '3': '.agents.v1.SearchWorkspaceMemoriesResponse'
    },
    {
      '1': 'DeleteWorkspaceMemory',
      '2': '.agents.v1.DeleteWorkspaceMemoryRequest',
      '3': '.agents.v1.DeleteWorkspaceMemoryResponse'
    },
  ],
};

@$core.Deprecated('Use workspaceMemoryServiceDescriptor instead')
const $core.Map<$core.String, $core.Map<$core.String, $core.dynamic>>
    WorkspaceMemoryServiceBase$messageJson = {
  '.agents.v1.ListWorkspaceMemoriesRequest': ListWorkspaceMemoriesRequest$json,
  '.agents.v1.ListWorkspaceMemoriesResponse':
      ListWorkspaceMemoriesResponse$json,
  '.agents.v1.WorkspaceMemory': WorkspaceMemory$json,
  '.google.protobuf.Timestamp': $0.Timestamp$json,
  '.agents.v1.SearchWorkspaceMemoriesRequest':
      SearchWorkspaceMemoriesRequest$json,
  '.agents.v1.SearchWorkspaceMemoriesResponse':
      SearchWorkspaceMemoriesResponse$json,
  '.agents.v1.DeleteWorkspaceMemoryRequest': DeleteWorkspaceMemoryRequest$json,
  '.agents.v1.DeleteWorkspaceMemoryResponse':
      DeleteWorkspaceMemoryResponse$json,
};

/// Descriptor for `WorkspaceMemoryService`. Decode as a `google.protobuf.ServiceDescriptorProto`.
final $typed_data.Uint8List workspaceMemoryServiceDescriptor = $convert.base64Decode(
    'ChZXb3Jrc3BhY2VNZW1vcnlTZXJ2aWNlEmoKFUxpc3RXb3Jrc3BhY2VNZW1vcmllcxInLmFnZW'
    '50cy52MS5MaXN0V29ya3NwYWNlTWVtb3JpZXNSZXF1ZXN0GiguYWdlbnRzLnYxLkxpc3RXb3Jr'
    'c3BhY2VNZW1vcmllc1Jlc3BvbnNlEnAKF1NlYXJjaFdvcmtzcGFjZU1lbW9yaWVzEikuYWdlbn'
    'RzLnYxLlNlYXJjaFdvcmtzcGFjZU1lbW9yaWVzUmVxdWVzdBoqLmFnZW50cy52MS5TZWFyY2hX'
    'b3Jrc3BhY2VNZW1vcmllc1Jlc3BvbnNlEmoKFURlbGV0ZVdvcmtzcGFjZU1lbW9yeRInLmFnZW'
    '50cy52MS5EZWxldGVXb3Jrc3BhY2VNZW1vcnlSZXF1ZXN0GiguYWdlbnRzLnYxLkRlbGV0ZVdv'
    'cmtzcGFjZU1lbW9yeVJlc3BvbnNl');
