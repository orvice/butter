// This is a generated file - do not edit.
//
// Generated from agents/v1/agent_file.proto.

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

@$core.Deprecated('Use agentFileMountPermissionDescriptor instead')
const AgentFileMountPermission$json = {
  '1': 'AgentFileMountPermission',
  '2': [
    {'1': 'AGENT_FILE_MOUNT_PERMISSION_UNSPECIFIED', '2': 0},
    {'1': 'AGENT_FILE_MOUNT_PERMISSION_READ', '2': 1},
    {'1': 'AGENT_FILE_MOUNT_PERMISSION_READ_WRITE', '2': 2},
    {'1': 'AGENT_FILE_MOUNT_PERMISSION_READ_WRITE_DELETE', '2': 3},
  ],
};

/// Descriptor for `AgentFileMountPermission`. Decode as a `google.protobuf.EnumDescriptorProto`.
final $typed_data.Uint8List agentFileMountPermissionDescriptor = $convert.base64Decode(
    'ChhBZ2VudEZpbGVNb3VudFBlcm1pc3Npb24SKwonQUdFTlRfRklMRV9NT1VOVF9QRVJNSVNTSU'
    '9OX1VOU1BFQ0lGSUVEEAASJAogQUdFTlRfRklMRV9NT1VOVF9QRVJNSVNTSU9OX1JFQUQQARIq'
    'CiZBR0VOVF9GSUxFX01PVU5UX1BFUk1JU1NJT05fUkVBRF9XUklURRACEjEKLUFHRU5UX0ZJTE'
    'VfTU9VTlRfUEVSTUlTU0lPTl9SRUFEX1dSSVRFX0RFTEVURRAD');

@$core.Deprecated('Use agentFileSpaceDescriptor instead')
const AgentFileSpace$json = {
  '1': 'AgentFileSpace',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
    {'1': 'name', '3': 2, '4': 1, '5': 9, '8': {}, '10': 'name'},
    {'1': 'description', '3': 3, '4': 1, '5': 9, '10': 'description'},
    {
      '1': 'metadata',
      '3': 4,
      '4': 3,
      '5': 11,
      '6': '.agents.v1.AgentFileSpace.MetadataEntry',
      '10': 'metadata'
    },
    {
      '1': 'created_at',
      '3': 5,
      '4': 1,
      '5': 11,
      '6': '.google.protobuf.Timestamp',
      '10': 'createdAt'
    },
    {
      '1': 'updated_at',
      '3': 6,
      '4': 1,
      '5': 11,
      '6': '.google.protobuf.Timestamp',
      '10': 'updatedAt'
    },
    {'1': 'workspace_id', '3': 100, '4': 1, '5': 9, '10': 'workspaceId'},
  ],
  '3': [AgentFileSpace_MetadataEntry$json],
};

@$core.Deprecated('Use agentFileSpaceDescriptor instead')
const AgentFileSpace_MetadataEntry$json = {
  '1': 'MetadataEntry',
  '2': [
    {'1': 'key', '3': 1, '4': 1, '5': 9, '10': 'key'},
    {'1': 'value', '3': 2, '4': 1, '5': 9, '10': 'value'},
  ],
  '7': {'7': true},
};

/// Descriptor for `AgentFileSpace`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List agentFileSpaceDescriptor = $convert.base64Decode(
    'Cg5BZ2VudEZpbGVTcGFjZRIOCgJpZBgBIAEoCVICaWQSGwoEbmFtZRgCIAEoCUIH+kIEcgIQAV'
    'IEbmFtZRIgCgtkZXNjcmlwdGlvbhgDIAEoCVILZGVzY3JpcHRpb24SQwoIbWV0YWRhdGEYBCAD'
    'KAsyJy5hZ2VudHMudjEuQWdlbnRGaWxlU3BhY2UuTWV0YWRhdGFFbnRyeVIIbWV0YWRhdGESOQ'
    'oKY3JlYXRlZF9hdBgFIAEoCzIaLmdvb2dsZS5wcm90b2J1Zi5UaW1lc3RhbXBSCWNyZWF0ZWRB'
    'dBI5Cgp1cGRhdGVkX2F0GAYgASgLMhouZ29vZ2xlLnByb3RvYnVmLlRpbWVzdGFtcFIJdXBkYX'
    'RlZEF0EiEKDHdvcmtzcGFjZV9pZBhkIAEoCVILd29ya3NwYWNlSWQaOwoNTWV0YWRhdGFFbnRy'
    'eRIQCgNrZXkYASABKAlSA2tleRIUCgV2YWx1ZRgCIAEoCVIFdmFsdWU6AjgB');

@$core.Deprecated('Use agentFileDescriptor instead')
const AgentFile$json = {
  '1': 'AgentFile',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
    {'1': 'space_id', '3': 2, '4': 1, '5': 9, '10': 'spaceId'},
    {'1': 'path', '3': 3, '4': 1, '5': 9, '8': {}, '10': 'path'},
    {'1': 'content_type', '3': 4, '4': 1, '5': 9, '10': 'contentType'},
    {'1': 'size_bytes', '3': 5, '4': 1, '5': 3, '10': 'sizeBytes'},
    {'1': 'version', '3': 6, '4': 1, '5': 3, '10': 'version'},
    {
      '1': 'metadata',
      '3': 7,
      '4': 3,
      '5': 11,
      '6': '.agents.v1.AgentFile.MetadataEntry',
      '10': 'metadata'
    },
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
    {'1': 'workspace_id', '3': 100, '4': 1, '5': 9, '10': 'workspaceId'},
  ],
  '3': [AgentFile_MetadataEntry$json],
};

@$core.Deprecated('Use agentFileDescriptor instead')
const AgentFile_MetadataEntry$json = {
  '1': 'MetadataEntry',
  '2': [
    {'1': 'key', '3': 1, '4': 1, '5': 9, '10': 'key'},
    {'1': 'value', '3': 2, '4': 1, '5': 9, '10': 'value'},
  ],
  '7': {'7': true},
};

/// Descriptor for `AgentFile`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List agentFileDescriptor = $convert.base64Decode(
    'CglBZ2VudEZpbGUSDgoCaWQYASABKAlSAmlkEhkKCHNwYWNlX2lkGAIgASgJUgdzcGFjZUlkEh'
    'sKBHBhdGgYAyABKAlCB/pCBHICEAFSBHBhdGgSIQoMY29udGVudF90eXBlGAQgASgJUgtjb250'
    'ZW50VHlwZRIdCgpzaXplX2J5dGVzGAUgASgDUglzaXplQnl0ZXMSGAoHdmVyc2lvbhgGIAEoA1'
    'IHdmVyc2lvbhI+CghtZXRhZGF0YRgHIAMoCzIiLmFnZW50cy52MS5BZ2VudEZpbGUuTWV0YWRh'
    'dGFFbnRyeVIIbWV0YWRhdGESOQoKY3JlYXRlZF9hdBgIIAEoCzIaLmdvb2dsZS5wcm90b2J1Zi'
    '5UaW1lc3RhbXBSCWNyZWF0ZWRBdBI5Cgp1cGRhdGVkX2F0GAkgASgLMhouZ29vZ2xlLnByb3Rv'
    'YnVmLlRpbWVzdGFtcFIJdXBkYXRlZEF0EiEKDHdvcmtzcGFjZV9pZBhkIAEoCVILd29ya3NwYW'
    'NlSWQaOwoNTWV0YWRhdGFFbnRyeRIQCgNrZXkYASABKAlSA2tleRIUCgV2YWx1ZRgCIAEoCVIF'
    'dmFsdWU6AjgB');

@$core.Deprecated('Use agentFileMountDescriptor instead')
const AgentFileMount$json = {
  '1': 'AgentFileMount',
  '2': [
    {'1': 'space_id', '3': 1, '4': 1, '5': 9, '8': {}, '10': 'spaceId'},
    {'1': 'mount_path', '3': 2, '4': 1, '5': 9, '10': 'mountPath'},
    {
      '1': 'permission',
      '3': 3,
      '4': 1,
      '5': 14,
      '6': '.agents.v1.AgentFileMountPermission',
      '10': 'permission'
    },
  ],
};

/// Descriptor for `AgentFileMount`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List agentFileMountDescriptor = $convert.base64Decode(
    'Cg5BZ2VudEZpbGVNb3VudBIiCghzcGFjZV9pZBgBIAEoCUIH+kIEcgIQAVIHc3BhY2VJZBIdCg'
    'ptb3VudF9wYXRoGAIgASgJUgltb3VudFBhdGgSQwoKcGVybWlzc2lvbhgDIAEoDjIjLmFnZW50'
    'cy52MS5BZ2VudEZpbGVNb3VudFBlcm1pc3Npb25SCnBlcm1pc3Npb24=');

@$core.Deprecated('Use listAgentFileSpacesRequestDescriptor instead')
const ListAgentFileSpacesRequest$json = {
  '1': 'ListAgentFileSpacesRequest',
};

/// Descriptor for `ListAgentFileSpacesRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listAgentFileSpacesRequestDescriptor =
    $convert.base64Decode('ChpMaXN0QWdlbnRGaWxlU3BhY2VzUmVxdWVzdA==');

@$core.Deprecated('Use listAgentFileSpacesResponseDescriptor instead')
const ListAgentFileSpacesResponse$json = {
  '1': 'ListAgentFileSpacesResponse',
  '2': [
    {
      '1': 'spaces',
      '3': 1,
      '4': 3,
      '5': 11,
      '6': '.agents.v1.AgentFileSpace',
      '10': 'spaces'
    },
  ],
};

/// Descriptor for `ListAgentFileSpacesResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listAgentFileSpacesResponseDescriptor =
    $convert.base64Decode(
        'ChtMaXN0QWdlbnRGaWxlU3BhY2VzUmVzcG9uc2USMQoGc3BhY2VzGAEgAygLMhkuYWdlbnRzLn'
        'YxLkFnZW50RmlsZVNwYWNlUgZzcGFjZXM=');

@$core.Deprecated('Use getAgentFileSpaceRequestDescriptor instead')
const GetAgentFileSpaceRequest$json = {
  '1': 'GetAgentFileSpaceRequest',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
  ],
};

/// Descriptor for `GetAgentFileSpaceRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getAgentFileSpaceRequestDescriptor = $convert
    .base64Decode('ChhHZXRBZ2VudEZpbGVTcGFjZVJlcXVlc3QSDgoCaWQYASABKAlSAmlk');

@$core.Deprecated('Use getAgentFileSpaceResponseDescriptor instead')
const GetAgentFileSpaceResponse$json = {
  '1': 'GetAgentFileSpaceResponse',
  '2': [
    {
      '1': 'space',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.AgentFileSpace',
      '10': 'space'
    },
  ],
};

/// Descriptor for `GetAgentFileSpaceResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getAgentFileSpaceResponseDescriptor =
    $convert.base64Decode(
        'ChlHZXRBZ2VudEZpbGVTcGFjZVJlc3BvbnNlEi8KBXNwYWNlGAEgASgLMhkuYWdlbnRzLnYxLk'
        'FnZW50RmlsZVNwYWNlUgVzcGFjZQ==');

@$core.Deprecated('Use createAgentFileSpaceRequestDescriptor instead')
const CreateAgentFileSpaceRequest$json = {
  '1': 'CreateAgentFileSpaceRequest',
  '2': [
    {
      '1': 'space',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.AgentFileSpace',
      '10': 'space'
    },
  ],
};

/// Descriptor for `CreateAgentFileSpaceRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List createAgentFileSpaceRequestDescriptor =
    $convert.base64Decode(
        'ChtDcmVhdGVBZ2VudEZpbGVTcGFjZVJlcXVlc3QSLwoFc3BhY2UYASABKAsyGS5hZ2VudHMudj'
        'EuQWdlbnRGaWxlU3BhY2VSBXNwYWNl');

@$core.Deprecated('Use createAgentFileSpaceResponseDescriptor instead')
const CreateAgentFileSpaceResponse$json = {
  '1': 'CreateAgentFileSpaceResponse',
  '2': [
    {
      '1': 'space',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.AgentFileSpace',
      '10': 'space'
    },
  ],
};

/// Descriptor for `CreateAgentFileSpaceResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List createAgentFileSpaceResponseDescriptor =
    $convert.base64Decode(
        'ChxDcmVhdGVBZ2VudEZpbGVTcGFjZVJlc3BvbnNlEi8KBXNwYWNlGAEgASgLMhkuYWdlbnRzLn'
        'YxLkFnZW50RmlsZVNwYWNlUgVzcGFjZQ==');

@$core.Deprecated('Use updateAgentFileSpaceRequestDescriptor instead')
const UpdateAgentFileSpaceRequest$json = {
  '1': 'UpdateAgentFileSpaceRequest',
  '2': [
    {
      '1': 'space',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.AgentFileSpace',
      '10': 'space'
    },
  ],
};

/// Descriptor for `UpdateAgentFileSpaceRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List updateAgentFileSpaceRequestDescriptor =
    $convert.base64Decode(
        'ChtVcGRhdGVBZ2VudEZpbGVTcGFjZVJlcXVlc3QSLwoFc3BhY2UYASABKAsyGS5hZ2VudHMudj'
        'EuQWdlbnRGaWxlU3BhY2VSBXNwYWNl');

@$core.Deprecated('Use updateAgentFileSpaceResponseDescriptor instead')
const UpdateAgentFileSpaceResponse$json = {
  '1': 'UpdateAgentFileSpaceResponse',
  '2': [
    {
      '1': 'space',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.AgentFileSpace',
      '10': 'space'
    },
  ],
};

/// Descriptor for `UpdateAgentFileSpaceResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List updateAgentFileSpaceResponseDescriptor =
    $convert.base64Decode(
        'ChxVcGRhdGVBZ2VudEZpbGVTcGFjZVJlc3BvbnNlEi8KBXNwYWNlGAEgASgLMhkuYWdlbnRzLn'
        'YxLkFnZW50RmlsZVNwYWNlUgVzcGFjZQ==');

@$core.Deprecated('Use deleteAgentFileSpaceRequestDescriptor instead')
const DeleteAgentFileSpaceRequest$json = {
  '1': 'DeleteAgentFileSpaceRequest',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
  ],
};

/// Descriptor for `DeleteAgentFileSpaceRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List deleteAgentFileSpaceRequestDescriptor =
    $convert.base64Decode(
        'ChtEZWxldGVBZ2VudEZpbGVTcGFjZVJlcXVlc3QSDgoCaWQYASABKAlSAmlk');

@$core.Deprecated('Use deleteAgentFileSpaceResponseDescriptor instead')
const DeleteAgentFileSpaceResponse$json = {
  '1': 'DeleteAgentFileSpaceResponse',
};

/// Descriptor for `DeleteAgentFileSpaceResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List deleteAgentFileSpaceResponseDescriptor =
    $convert.base64Decode('ChxEZWxldGVBZ2VudEZpbGVTcGFjZVJlc3BvbnNl');

@$core.Deprecated('Use listAgentFilesRequestDescriptor instead')
const ListAgentFilesRequest$json = {
  '1': 'ListAgentFilesRequest',
  '2': [
    {'1': 'space_id', '3': 1, '4': 1, '5': 9, '10': 'spaceId'},
    {'1': 'path_prefix', '3': 2, '4': 1, '5': 9, '10': 'pathPrefix'},
  ],
};

/// Descriptor for `ListAgentFilesRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listAgentFilesRequestDescriptor = $convert.base64Decode(
    'ChVMaXN0QWdlbnRGaWxlc1JlcXVlc3QSGQoIc3BhY2VfaWQYASABKAlSB3NwYWNlSWQSHwoLcG'
    'F0aF9wcmVmaXgYAiABKAlSCnBhdGhQcmVmaXg=');

@$core.Deprecated('Use listAgentFilesResponseDescriptor instead')
const ListAgentFilesResponse$json = {
  '1': 'ListAgentFilesResponse',
  '2': [
    {
      '1': 'files',
      '3': 1,
      '4': 3,
      '5': 11,
      '6': '.agents.v1.AgentFile',
      '10': 'files'
    },
  ],
};

/// Descriptor for `ListAgentFilesResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listAgentFilesResponseDescriptor =
    $convert.base64Decode(
        'ChZMaXN0QWdlbnRGaWxlc1Jlc3BvbnNlEioKBWZpbGVzGAEgAygLMhQuYWdlbnRzLnYxLkFnZW'
        '50RmlsZVIFZmlsZXM=');

@$core.Deprecated('Use getAgentFileRequestDescriptor instead')
const GetAgentFileRequest$json = {
  '1': 'GetAgentFileRequest',
  '2': [
    {'1': 'space_id', '3': 1, '4': 1, '5': 9, '10': 'spaceId'},
    {'1': 'path', '3': 2, '4': 1, '5': 9, '10': 'path'},
    {'1': 'version', '3': 3, '4': 1, '5': 3, '10': 'version'},
  ],
};

/// Descriptor for `GetAgentFileRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getAgentFileRequestDescriptor = $convert.base64Decode(
    'ChNHZXRBZ2VudEZpbGVSZXF1ZXN0EhkKCHNwYWNlX2lkGAEgASgJUgdzcGFjZUlkEhIKBHBhdG'
    'gYAiABKAlSBHBhdGgSGAoHdmVyc2lvbhgDIAEoA1IHdmVyc2lvbg==');

@$core.Deprecated('Use getAgentFileResponseDescriptor instead')
const GetAgentFileResponse$json = {
  '1': 'GetAgentFileResponse',
  '2': [
    {
      '1': 'file',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.AgentFile',
      '10': 'file'
    },
    {'1': 'content', '3': 2, '4': 1, '5': 9, '10': 'content'},
  ],
};

/// Descriptor for `GetAgentFileResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getAgentFileResponseDescriptor = $convert.base64Decode(
    'ChRHZXRBZ2VudEZpbGVSZXNwb25zZRIoCgRmaWxlGAEgASgLMhQuYWdlbnRzLnYxLkFnZW50Rm'
    'lsZVIEZmlsZRIYCgdjb250ZW50GAIgASgJUgdjb250ZW50');

@$core.Deprecated('Use writeAgentFileRequestDescriptor instead')
const WriteAgentFileRequest$json = {
  '1': 'WriteAgentFileRequest',
  '2': [
    {'1': 'space_id', '3': 1, '4': 1, '5': 9, '10': 'spaceId'},
    {'1': 'path', '3': 2, '4': 1, '5': 9, '10': 'path'},
    {'1': 'content', '3': 3, '4': 1, '5': 9, '10': 'content'},
    {'1': 'content_type', '3': 4, '4': 1, '5': 9, '10': 'contentType'},
    {
      '1': 'metadata',
      '3': 5,
      '4': 3,
      '5': 11,
      '6': '.agents.v1.WriteAgentFileRequest.MetadataEntry',
      '10': 'metadata'
    },
  ],
  '3': [WriteAgentFileRequest_MetadataEntry$json],
};

@$core.Deprecated('Use writeAgentFileRequestDescriptor instead')
const WriteAgentFileRequest_MetadataEntry$json = {
  '1': 'MetadataEntry',
  '2': [
    {'1': 'key', '3': 1, '4': 1, '5': 9, '10': 'key'},
    {'1': 'value', '3': 2, '4': 1, '5': 9, '10': 'value'},
  ],
  '7': {'7': true},
};

/// Descriptor for `WriteAgentFileRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List writeAgentFileRequestDescriptor = $convert.base64Decode(
    'ChVXcml0ZUFnZW50RmlsZVJlcXVlc3QSGQoIc3BhY2VfaWQYASABKAlSB3NwYWNlSWQSEgoEcG'
    'F0aBgCIAEoCVIEcGF0aBIYCgdjb250ZW50GAMgASgJUgdjb250ZW50EiEKDGNvbnRlbnRfdHlw'
    'ZRgEIAEoCVILY29udGVudFR5cGUSSgoIbWV0YWRhdGEYBSADKAsyLi5hZ2VudHMudjEuV3JpdG'
    'VBZ2VudEZpbGVSZXF1ZXN0Lk1ldGFkYXRhRW50cnlSCG1ldGFkYXRhGjsKDU1ldGFkYXRhRW50'
    'cnkSEAoDa2V5GAEgASgJUgNrZXkSFAoFdmFsdWUYAiABKAlSBXZhbHVlOgI4AQ==');

@$core.Deprecated('Use writeAgentFileResponseDescriptor instead')
const WriteAgentFileResponse$json = {
  '1': 'WriteAgentFileResponse',
  '2': [
    {
      '1': 'file',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.AgentFile',
      '10': 'file'
    },
  ],
};

/// Descriptor for `WriteAgentFileResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List writeAgentFileResponseDescriptor =
    $convert.base64Decode(
        'ChZXcml0ZUFnZW50RmlsZVJlc3BvbnNlEigKBGZpbGUYASABKAsyFC5hZ2VudHMudjEuQWdlbn'
        'RGaWxlUgRmaWxl');

@$core.Deprecated('Use deleteAgentFileRequestDescriptor instead')
const DeleteAgentFileRequest$json = {
  '1': 'DeleteAgentFileRequest',
  '2': [
    {'1': 'space_id', '3': 1, '4': 1, '5': 9, '10': 'spaceId'},
    {'1': 'path', '3': 2, '4': 1, '5': 9, '10': 'path'},
  ],
};

/// Descriptor for `DeleteAgentFileRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List deleteAgentFileRequestDescriptor =
    $convert.base64Decode(
        'ChZEZWxldGVBZ2VudEZpbGVSZXF1ZXN0EhkKCHNwYWNlX2lkGAEgASgJUgdzcGFjZUlkEhIKBH'
        'BhdGgYAiABKAlSBHBhdGg=');

@$core.Deprecated('Use deleteAgentFileResponseDescriptor instead')
const DeleteAgentFileResponse$json = {
  '1': 'DeleteAgentFileResponse',
};

/// Descriptor for `DeleteAgentFileResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List deleteAgentFileResponseDescriptor =
    $convert.base64Decode('ChdEZWxldGVBZ2VudEZpbGVSZXNwb25zZQ==');

@$core.Deprecated('Use searchAgentFilesRequestDescriptor instead')
const SearchAgentFilesRequest$json = {
  '1': 'SearchAgentFilesRequest',
  '2': [
    {'1': 'space_id', '3': 1, '4': 1, '5': 9, '10': 'spaceId'},
    {'1': 'query', '3': 2, '4': 1, '5': 9, '10': 'query'},
    {'1': 'limit', '3': 3, '4': 1, '5': 5, '10': 'limit'},
  ],
};

/// Descriptor for `SearchAgentFilesRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List searchAgentFilesRequestDescriptor =
    $convert.base64Decode(
        'ChdTZWFyY2hBZ2VudEZpbGVzUmVxdWVzdBIZCghzcGFjZV9pZBgBIAEoCVIHc3BhY2VJZBIUCg'
        'VxdWVyeRgCIAEoCVIFcXVlcnkSFAoFbGltaXQYAyABKAVSBWxpbWl0');

@$core.Deprecated('Use searchAgentFilesResponseDescriptor instead')
const SearchAgentFilesResponse$json = {
  '1': 'SearchAgentFilesResponse',
  '2': [
    {
      '1': 'results',
      '3': 1,
      '4': 3,
      '5': 11,
      '6': '.agents.v1.AgentFileSearchResult',
      '10': 'results'
    },
  ],
};

/// Descriptor for `SearchAgentFilesResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List searchAgentFilesResponseDescriptor =
    $convert.base64Decode(
        'ChhTZWFyY2hBZ2VudEZpbGVzUmVzcG9uc2USOgoHcmVzdWx0cxgBIAMoCzIgLmFnZW50cy52MS'
        '5BZ2VudEZpbGVTZWFyY2hSZXN1bHRSB3Jlc3VsdHM=');

@$core.Deprecated('Use agentFileSearchResultDescriptor instead')
const AgentFileSearchResult$json = {
  '1': 'AgentFileSearchResult',
  '2': [
    {
      '1': 'file',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.AgentFile',
      '10': 'file'
    },
    {'1': 'snippets', '3': 2, '4': 3, '5': 9, '10': 'snippets'},
  ],
};

/// Descriptor for `AgentFileSearchResult`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List agentFileSearchResultDescriptor = $convert.base64Decode(
    'ChVBZ2VudEZpbGVTZWFyY2hSZXN1bHQSKAoEZmlsZRgBIAEoCzIULmFnZW50cy52MS5BZ2VudE'
    'ZpbGVSBGZpbGUSGgoIc25pcHBldHMYAiADKAlSCHNuaXBwZXRz');

const $core.Map<$core.String, $core.dynamic> AgentFileServiceBase$json = {
  '1': 'AgentFileService',
  '2': [
    {
      '1': 'ListAgentFileSpaces',
      '2': '.agents.v1.ListAgentFileSpacesRequest',
      '3': '.agents.v1.ListAgentFileSpacesResponse'
    },
    {
      '1': 'GetAgentFileSpace',
      '2': '.agents.v1.GetAgentFileSpaceRequest',
      '3': '.agents.v1.GetAgentFileSpaceResponse'
    },
    {
      '1': 'CreateAgentFileSpace',
      '2': '.agents.v1.CreateAgentFileSpaceRequest',
      '3': '.agents.v1.CreateAgentFileSpaceResponse'
    },
    {
      '1': 'UpdateAgentFileSpace',
      '2': '.agents.v1.UpdateAgentFileSpaceRequest',
      '3': '.agents.v1.UpdateAgentFileSpaceResponse'
    },
    {
      '1': 'DeleteAgentFileSpace',
      '2': '.agents.v1.DeleteAgentFileSpaceRequest',
      '3': '.agents.v1.DeleteAgentFileSpaceResponse'
    },
    {
      '1': 'ListAgentFiles',
      '2': '.agents.v1.ListAgentFilesRequest',
      '3': '.agents.v1.ListAgentFilesResponse'
    },
    {
      '1': 'GetAgentFile',
      '2': '.agents.v1.GetAgentFileRequest',
      '3': '.agents.v1.GetAgentFileResponse'
    },
    {
      '1': 'WriteAgentFile',
      '2': '.agents.v1.WriteAgentFileRequest',
      '3': '.agents.v1.WriteAgentFileResponse'
    },
    {
      '1': 'DeleteAgentFile',
      '2': '.agents.v1.DeleteAgentFileRequest',
      '3': '.agents.v1.DeleteAgentFileResponse'
    },
    {
      '1': 'SearchAgentFiles',
      '2': '.agents.v1.SearchAgentFilesRequest',
      '3': '.agents.v1.SearchAgentFilesResponse'
    },
  ],
};

@$core.Deprecated('Use agentFileServiceDescriptor instead')
const $core.Map<$core.String, $core.Map<$core.String, $core.dynamic>>
    AgentFileServiceBase$messageJson = {
  '.agents.v1.ListAgentFileSpacesRequest': ListAgentFileSpacesRequest$json,
  '.agents.v1.ListAgentFileSpacesResponse': ListAgentFileSpacesResponse$json,
  '.agents.v1.AgentFileSpace': AgentFileSpace$json,
  '.agents.v1.AgentFileSpace.MetadataEntry': AgentFileSpace_MetadataEntry$json,
  '.google.protobuf.Timestamp': $0.Timestamp$json,
  '.agents.v1.GetAgentFileSpaceRequest': GetAgentFileSpaceRequest$json,
  '.agents.v1.GetAgentFileSpaceResponse': GetAgentFileSpaceResponse$json,
  '.agents.v1.CreateAgentFileSpaceRequest': CreateAgentFileSpaceRequest$json,
  '.agents.v1.CreateAgentFileSpaceResponse': CreateAgentFileSpaceResponse$json,
  '.agents.v1.UpdateAgentFileSpaceRequest': UpdateAgentFileSpaceRequest$json,
  '.agents.v1.UpdateAgentFileSpaceResponse': UpdateAgentFileSpaceResponse$json,
  '.agents.v1.DeleteAgentFileSpaceRequest': DeleteAgentFileSpaceRequest$json,
  '.agents.v1.DeleteAgentFileSpaceResponse': DeleteAgentFileSpaceResponse$json,
  '.agents.v1.ListAgentFilesRequest': ListAgentFilesRequest$json,
  '.agents.v1.ListAgentFilesResponse': ListAgentFilesResponse$json,
  '.agents.v1.AgentFile': AgentFile$json,
  '.agents.v1.AgentFile.MetadataEntry': AgentFile_MetadataEntry$json,
  '.agents.v1.GetAgentFileRequest': GetAgentFileRequest$json,
  '.agents.v1.GetAgentFileResponse': GetAgentFileResponse$json,
  '.agents.v1.WriteAgentFileRequest': WriteAgentFileRequest$json,
  '.agents.v1.WriteAgentFileRequest.MetadataEntry':
      WriteAgentFileRequest_MetadataEntry$json,
  '.agents.v1.WriteAgentFileResponse': WriteAgentFileResponse$json,
  '.agents.v1.DeleteAgentFileRequest': DeleteAgentFileRequest$json,
  '.agents.v1.DeleteAgentFileResponse': DeleteAgentFileResponse$json,
  '.agents.v1.SearchAgentFilesRequest': SearchAgentFilesRequest$json,
  '.agents.v1.SearchAgentFilesResponse': SearchAgentFilesResponse$json,
  '.agents.v1.AgentFileSearchResult': AgentFileSearchResult$json,
};

/// Descriptor for `AgentFileService`. Decode as a `google.protobuf.ServiceDescriptorProto`.
final $typed_data.Uint8List agentFileServiceDescriptor = $convert.base64Decode(
    'ChBBZ2VudEZpbGVTZXJ2aWNlEmQKE0xpc3RBZ2VudEZpbGVTcGFjZXMSJS5hZ2VudHMudjEuTG'
    'lzdEFnZW50RmlsZVNwYWNlc1JlcXVlc3QaJi5hZ2VudHMudjEuTGlzdEFnZW50RmlsZVNwYWNl'
    'c1Jlc3BvbnNlEl4KEUdldEFnZW50RmlsZVNwYWNlEiMuYWdlbnRzLnYxLkdldEFnZW50RmlsZV'
    'NwYWNlUmVxdWVzdBokLmFnZW50cy52MS5HZXRBZ2VudEZpbGVTcGFjZVJlc3BvbnNlEmcKFENy'
    'ZWF0ZUFnZW50RmlsZVNwYWNlEiYuYWdlbnRzLnYxLkNyZWF0ZUFnZW50RmlsZVNwYWNlUmVxdW'
    'VzdBonLmFnZW50cy52MS5DcmVhdGVBZ2VudEZpbGVTcGFjZVJlc3BvbnNlEmcKFFVwZGF0ZUFn'
    'ZW50RmlsZVNwYWNlEiYuYWdlbnRzLnYxLlVwZGF0ZUFnZW50RmlsZVNwYWNlUmVxdWVzdBonLm'
    'FnZW50cy52MS5VcGRhdGVBZ2VudEZpbGVTcGFjZVJlc3BvbnNlEmcKFERlbGV0ZUFnZW50Rmls'
    'ZVNwYWNlEiYuYWdlbnRzLnYxLkRlbGV0ZUFnZW50RmlsZVNwYWNlUmVxdWVzdBonLmFnZW50cy'
    '52MS5EZWxldGVBZ2VudEZpbGVTcGFjZVJlc3BvbnNlElUKDkxpc3RBZ2VudEZpbGVzEiAuYWdl'
    'bnRzLnYxLkxpc3RBZ2VudEZpbGVzUmVxdWVzdBohLmFnZW50cy52MS5MaXN0QWdlbnRGaWxlc1'
    'Jlc3BvbnNlEk8KDEdldEFnZW50RmlsZRIeLmFnZW50cy52MS5HZXRBZ2VudEZpbGVSZXF1ZXN0'
    'Gh8uYWdlbnRzLnYxLkdldEFnZW50RmlsZVJlc3BvbnNlElUKDldyaXRlQWdlbnRGaWxlEiAuYW'
    'dlbnRzLnYxLldyaXRlQWdlbnRGaWxlUmVxdWVzdBohLmFnZW50cy52MS5Xcml0ZUFnZW50Rmls'
    'ZVJlc3BvbnNlElgKD0RlbGV0ZUFnZW50RmlsZRIhLmFnZW50cy52MS5EZWxldGVBZ2VudEZpbG'
    'VSZXF1ZXN0GiIuYWdlbnRzLnYxLkRlbGV0ZUFnZW50RmlsZVJlc3BvbnNlElsKEFNlYXJjaEFn'
    'ZW50RmlsZXMSIi5hZ2VudHMudjEuU2VhcmNoQWdlbnRGaWxlc1JlcXVlc3QaIy5hZ2VudHMudj'
    'EuU2VhcmNoQWdlbnRGaWxlc1Jlc3BvbnNl');
