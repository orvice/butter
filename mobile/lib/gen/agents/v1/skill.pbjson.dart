// This is a generated file - do not edit.
//
// Generated from agents/v1/skill.proto.

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

@$core.Deprecated('Use skillDescriptor instead')
const Skill$json = {
  '1': 'Skill',
  '2': [
    {'1': 'name', '3': 1, '4': 1, '5': 9, '8': {}, '10': 'name'},
    {'1': 'description', '3': 2, '4': 1, '5': 9, '10': 'description'},
    {'1': 'license', '3': 3, '4': 1, '5': 9, '10': 'license'},
    {'1': 'compatibility', '3': 4, '4': 1, '5': 9, '10': 'compatibility'},
    {
      '1': 'metadata',
      '3': 5,
      '4': 3,
      '5': 11,
      '6': '.agents.v1.Skill.MetadataEntry',
      '10': 'metadata'
    },
    {'1': 'allowed_tools', '3': 6, '4': 3, '5': 9, '10': 'allowedTools'},
    {'1': 'size_bytes', '3': 7, '4': 1, '5': 3, '10': 'sizeBytes'},
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
  '3': [Skill_MetadataEntry$json],
};

@$core.Deprecated('Use skillDescriptor instead')
const Skill_MetadataEntry$json = {
  '1': 'MetadataEntry',
  '2': [
    {'1': 'key', '3': 1, '4': 1, '5': 9, '10': 'key'},
    {'1': 'value', '3': 2, '4': 1, '5': 9, '10': 'value'},
  ],
  '7': {'7': true},
};

/// Descriptor for `Skill`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List skillDescriptor = $convert.base64Decode(
    'CgVTa2lsbBIbCgRuYW1lGAEgASgJQgf6QgRyAhABUgRuYW1lEiAKC2Rlc2NyaXB0aW9uGAIgAS'
    'gJUgtkZXNjcmlwdGlvbhIYCgdsaWNlbnNlGAMgASgJUgdsaWNlbnNlEiQKDWNvbXBhdGliaWxp'
    'dHkYBCABKAlSDWNvbXBhdGliaWxpdHkSOgoIbWV0YWRhdGEYBSADKAsyHi5hZ2VudHMudjEuU2'
    'tpbGwuTWV0YWRhdGFFbnRyeVIIbWV0YWRhdGESIwoNYWxsb3dlZF90b29scxgGIAMoCVIMYWxs'
    'b3dlZFRvb2xzEh0KCnNpemVfYnl0ZXMYByABKANSCXNpemVCeXRlcxI5CgpjcmVhdGVkX2F0GA'
    'ggASgLMhouZ29vZ2xlLnByb3RvYnVmLlRpbWVzdGFtcFIJY3JlYXRlZEF0EjkKCnVwZGF0ZWRf'
    'YXQYCSABKAsyGi5nb29nbGUucHJvdG9idWYuVGltZXN0YW1wUgl1cGRhdGVkQXQSIQoMd29ya3'
    'NwYWNlX2lkGGQgASgJUgt3b3Jrc3BhY2VJZBo7Cg1NZXRhZGF0YUVudHJ5EhAKA2tleRgBIAEo'
    'CVIDa2V5EhQKBXZhbHVlGAIgASgJUgV2YWx1ZToCOAE=');

@$core.Deprecated('Use skillResourceDescriptor instead')
const SkillResource$json = {
  '1': 'SkillResource',
  '2': [
    {'1': 'path', '3': 1, '4': 1, '5': 9, '8': {}, '10': 'path'},
    {'1': 'size_bytes', '3': 2, '4': 1, '5': 3, '10': 'sizeBytes'},
    {'1': 'content_type', '3': 3, '4': 1, '5': 9, '10': 'contentType'},
    {
      '1': 'created_at',
      '3': 4,
      '4': 1,
      '5': 11,
      '6': '.google.protobuf.Timestamp',
      '10': 'createdAt'
    },
    {
      '1': 'updated_at',
      '3': 5,
      '4': 1,
      '5': 11,
      '6': '.google.protobuf.Timestamp',
      '10': 'updatedAt'
    },
  ],
};

/// Descriptor for `SkillResource`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List skillResourceDescriptor = $convert.base64Decode(
    'Cg1Ta2lsbFJlc291cmNlEhsKBHBhdGgYASABKAlCB/pCBHICEAFSBHBhdGgSHQoKc2l6ZV9ieX'
    'RlcxgCIAEoA1IJc2l6ZUJ5dGVzEiEKDGNvbnRlbnRfdHlwZRgDIAEoCVILY29udGVudFR5cGUS'
    'OQoKY3JlYXRlZF9hdBgEIAEoCzIaLmdvb2dsZS5wcm90b2J1Zi5UaW1lc3RhbXBSCWNyZWF0ZW'
    'RBdBI5Cgp1cGRhdGVkX2F0GAUgASgLMhouZ29vZ2xlLnByb3RvYnVmLlRpbWVzdGFtcFIJdXBk'
    'YXRlZEF0');

@$core.Deprecated('Use listSkillsRequestDescriptor instead')
const ListSkillsRequest$json = {
  '1': 'ListSkillsRequest',
};

/// Descriptor for `ListSkillsRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listSkillsRequestDescriptor =
    $convert.base64Decode('ChFMaXN0U2tpbGxzUmVxdWVzdA==');

@$core.Deprecated('Use listSkillsResponseDescriptor instead')
const ListSkillsResponse$json = {
  '1': 'ListSkillsResponse',
  '2': [
    {
      '1': 'skills',
      '3': 1,
      '4': 3,
      '5': 11,
      '6': '.agents.v1.Skill',
      '10': 'skills'
    },
  ],
};

/// Descriptor for `ListSkillsResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listSkillsResponseDescriptor = $convert.base64Decode(
    'ChJMaXN0U2tpbGxzUmVzcG9uc2USKAoGc2tpbGxzGAEgAygLMhAuYWdlbnRzLnYxLlNraWxsUg'
    'Zza2lsbHM=');

@$core.Deprecated('Use getSkillRequestDescriptor instead')
const GetSkillRequest$json = {
  '1': 'GetSkillRequest',
  '2': [
    {'1': 'name', '3': 1, '4': 1, '5': 9, '8': {}, '10': 'name'},
  ],
};

/// Descriptor for `GetSkillRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getSkillRequestDescriptor = $convert.base64Decode(
    'Cg9HZXRTa2lsbFJlcXVlc3QSGwoEbmFtZRgBIAEoCUIH+kIEcgIQAVIEbmFtZQ==');

@$core.Deprecated('Use getSkillResponseDescriptor instead')
const GetSkillResponse$json = {
  '1': 'GetSkillResponse',
  '2': [
    {
      '1': 'skill',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.Skill',
      '10': 'skill'
    },
    {'1': 'skill_md', '3': 2, '4': 1, '5': 9, '10': 'skillMd'},
  ],
};

/// Descriptor for `GetSkillResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getSkillResponseDescriptor = $convert.base64Decode(
    'ChBHZXRTa2lsbFJlc3BvbnNlEiYKBXNraWxsGAEgASgLMhAuYWdlbnRzLnYxLlNraWxsUgVza2'
    'lsbBIZCghza2lsbF9tZBgCIAEoCVIHc2tpbGxNZA==');

@$core.Deprecated('Use createSkillRequestDescriptor instead')
const CreateSkillRequest$json = {
  '1': 'CreateSkillRequest',
  '2': [
    {'1': 'name', '3': 1, '4': 1, '5': 9, '8': {}, '10': 'name'},
    {'1': 'skill_md', '3': 2, '4': 1, '5': 9, '8': {}, '10': 'skillMd'},
  ],
};

/// Descriptor for `CreateSkillRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List createSkillRequestDescriptor = $convert.base64Decode(
    'ChJDcmVhdGVTa2lsbFJlcXVlc3QSGwoEbmFtZRgBIAEoCUIH+kIEcgIQAVIEbmFtZRIiCghza2'
    'lsbF9tZBgCIAEoCUIH+kIEcgIQAVIHc2tpbGxNZA==');

@$core.Deprecated('Use createSkillResponseDescriptor instead')
const CreateSkillResponse$json = {
  '1': 'CreateSkillResponse',
  '2': [
    {
      '1': 'skill',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.Skill',
      '10': 'skill'
    },
  ],
};

/// Descriptor for `CreateSkillResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List createSkillResponseDescriptor = $convert.base64Decode(
    'ChNDcmVhdGVTa2lsbFJlc3BvbnNlEiYKBXNraWxsGAEgASgLMhAuYWdlbnRzLnYxLlNraWxsUg'
    'Vza2lsbA==');

@$core.Deprecated('Use updateSkillRequestDescriptor instead')
const UpdateSkillRequest$json = {
  '1': 'UpdateSkillRequest',
  '2': [
    {'1': 'name', '3': 1, '4': 1, '5': 9, '8': {}, '10': 'name'},
    {'1': 'skill_md', '3': 2, '4': 1, '5': 9, '8': {}, '10': 'skillMd'},
  ],
};

/// Descriptor for `UpdateSkillRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List updateSkillRequestDescriptor = $convert.base64Decode(
    'ChJVcGRhdGVTa2lsbFJlcXVlc3QSGwoEbmFtZRgBIAEoCUIH+kIEcgIQAVIEbmFtZRIiCghza2'
    'lsbF9tZBgCIAEoCUIH+kIEcgIQAVIHc2tpbGxNZA==');

@$core.Deprecated('Use updateSkillResponseDescriptor instead')
const UpdateSkillResponse$json = {
  '1': 'UpdateSkillResponse',
  '2': [
    {
      '1': 'skill',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.Skill',
      '10': 'skill'
    },
  ],
};

/// Descriptor for `UpdateSkillResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List updateSkillResponseDescriptor = $convert.base64Decode(
    'ChNVcGRhdGVTa2lsbFJlc3BvbnNlEiYKBXNraWxsGAEgASgLMhAuYWdlbnRzLnYxLlNraWxsUg'
    'Vza2lsbA==');

@$core.Deprecated('Use deleteSkillRequestDescriptor instead')
const DeleteSkillRequest$json = {
  '1': 'DeleteSkillRequest',
  '2': [
    {'1': 'name', '3': 1, '4': 1, '5': 9, '8': {}, '10': 'name'},
  ],
};

/// Descriptor for `DeleteSkillRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List deleteSkillRequestDescriptor =
    $convert.base64Decode(
        'ChJEZWxldGVTa2lsbFJlcXVlc3QSGwoEbmFtZRgBIAEoCUIH+kIEcgIQAVIEbmFtZQ==');

@$core.Deprecated('Use deleteSkillResponseDescriptor instead')
const DeleteSkillResponse$json = {
  '1': 'DeleteSkillResponse',
};

/// Descriptor for `DeleteSkillResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List deleteSkillResponseDescriptor =
    $convert.base64Decode('ChNEZWxldGVTa2lsbFJlc3BvbnNl');

@$core.Deprecated('Use listSkillResourcesRequestDescriptor instead')
const ListSkillResourcesRequest$json = {
  '1': 'ListSkillResourcesRequest',
  '2': [
    {'1': 'skill_name', '3': 1, '4': 1, '5': 9, '8': {}, '10': 'skillName'},
  ],
};

/// Descriptor for `ListSkillResourcesRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listSkillResourcesRequestDescriptor =
    $convert.base64Decode(
        'ChlMaXN0U2tpbGxSZXNvdXJjZXNSZXF1ZXN0EiYKCnNraWxsX25hbWUYASABKAlCB/pCBHICEA'
        'FSCXNraWxsTmFtZQ==');

@$core.Deprecated('Use listSkillResourcesResponseDescriptor instead')
const ListSkillResourcesResponse$json = {
  '1': 'ListSkillResourcesResponse',
  '2': [
    {
      '1': 'resources',
      '3': 1,
      '4': 3,
      '5': 11,
      '6': '.agents.v1.SkillResource',
      '10': 'resources'
    },
  ],
};

/// Descriptor for `ListSkillResourcesResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listSkillResourcesResponseDescriptor =
    $convert.base64Decode(
        'ChpMaXN0U2tpbGxSZXNvdXJjZXNSZXNwb25zZRI2CglyZXNvdXJjZXMYASADKAsyGC5hZ2VudH'
        'MudjEuU2tpbGxSZXNvdXJjZVIJcmVzb3VyY2Vz');

@$core.Deprecated('Use getSkillResourceRequestDescriptor instead')
const GetSkillResourceRequest$json = {
  '1': 'GetSkillResourceRequest',
  '2': [
    {'1': 'skill_name', '3': 1, '4': 1, '5': 9, '8': {}, '10': 'skillName'},
    {'1': 'path', '3': 2, '4': 1, '5': 9, '8': {}, '10': 'path'},
  ],
};

/// Descriptor for `GetSkillResourceRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getSkillResourceRequestDescriptor =
    $convert.base64Decode(
        'ChdHZXRTa2lsbFJlc291cmNlUmVxdWVzdBImCgpza2lsbF9uYW1lGAEgASgJQgf6QgRyAhABUg'
        'lza2lsbE5hbWUSGwoEcGF0aBgCIAEoCUIH+kIEcgIQAVIEcGF0aA==');

@$core.Deprecated('Use getSkillResourceResponseDescriptor instead')
const GetSkillResourceResponse$json = {
  '1': 'GetSkillResourceResponse',
  '2': [
    {
      '1': 'resource',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.SkillResource',
      '10': 'resource'
    },
    {'1': 'content', '3': 2, '4': 1, '5': 12, '10': 'content'},
  ],
};

/// Descriptor for `GetSkillResourceResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getSkillResourceResponseDescriptor =
    $convert.base64Decode(
        'ChhHZXRTa2lsbFJlc291cmNlUmVzcG9uc2USNAoIcmVzb3VyY2UYASABKAsyGC5hZ2VudHMudj'
        'EuU2tpbGxSZXNvdXJjZVIIcmVzb3VyY2USGAoHY29udGVudBgCIAEoDFIHY29udGVudA==');

@$core.Deprecated('Use putSkillResourceRequestDescriptor instead')
const PutSkillResourceRequest$json = {
  '1': 'PutSkillResourceRequest',
  '2': [
    {'1': 'skill_name', '3': 1, '4': 1, '5': 9, '8': {}, '10': 'skillName'},
    {'1': 'path', '3': 2, '4': 1, '5': 9, '8': {}, '10': 'path'},
    {'1': 'content', '3': 3, '4': 1, '5': 12, '10': 'content'},
    {'1': 'content_type', '3': 4, '4': 1, '5': 9, '10': 'contentType'},
  ],
};

/// Descriptor for `PutSkillResourceRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List putSkillResourceRequestDescriptor = $convert.base64Decode(
    'ChdQdXRTa2lsbFJlc291cmNlUmVxdWVzdBImCgpza2lsbF9uYW1lGAEgASgJQgf6QgRyAhABUg'
    'lza2lsbE5hbWUSGwoEcGF0aBgCIAEoCUIH+kIEcgIQAVIEcGF0aBIYCgdjb250ZW50GAMgASgM'
    'Ugdjb250ZW50EiEKDGNvbnRlbnRfdHlwZRgEIAEoCVILY29udGVudFR5cGU=');

@$core.Deprecated('Use putSkillResourceResponseDescriptor instead')
const PutSkillResourceResponse$json = {
  '1': 'PutSkillResourceResponse',
  '2': [
    {
      '1': 'resource',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.SkillResource',
      '10': 'resource'
    },
  ],
};

/// Descriptor for `PutSkillResourceResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List putSkillResourceResponseDescriptor =
    $convert.base64Decode(
        'ChhQdXRTa2lsbFJlc291cmNlUmVzcG9uc2USNAoIcmVzb3VyY2UYASABKAsyGC5hZ2VudHMudj'
        'EuU2tpbGxSZXNvdXJjZVIIcmVzb3VyY2U=');

@$core.Deprecated('Use deleteSkillResourceRequestDescriptor instead')
const DeleteSkillResourceRequest$json = {
  '1': 'DeleteSkillResourceRequest',
  '2': [
    {'1': 'skill_name', '3': 1, '4': 1, '5': 9, '8': {}, '10': 'skillName'},
    {'1': 'path', '3': 2, '4': 1, '5': 9, '8': {}, '10': 'path'},
  ],
};

/// Descriptor for `DeleteSkillResourceRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List deleteSkillResourceRequestDescriptor =
    $convert.base64Decode(
        'ChpEZWxldGVTa2lsbFJlc291cmNlUmVxdWVzdBImCgpza2lsbF9uYW1lGAEgASgJQgf6QgRyAh'
        'ABUglza2lsbE5hbWUSGwoEcGF0aBgCIAEoCUIH+kIEcgIQAVIEcGF0aA==');

@$core.Deprecated('Use deleteSkillResourceResponseDescriptor instead')
const DeleteSkillResourceResponse$json = {
  '1': 'DeleteSkillResourceResponse',
};

/// Descriptor for `DeleteSkillResourceResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List deleteSkillResourceResponseDescriptor =
    $convert.base64Decode('ChtEZWxldGVTa2lsbFJlc291cmNlUmVzcG9uc2U=');

const $core.Map<$core.String, $core.dynamic> SkillServiceBase$json = {
  '1': 'SkillService',
  '2': [
    {
      '1': 'ListSkills',
      '2': '.agents.v1.ListSkillsRequest',
      '3': '.agents.v1.ListSkillsResponse'
    },
    {
      '1': 'GetSkill',
      '2': '.agents.v1.GetSkillRequest',
      '3': '.agents.v1.GetSkillResponse'
    },
    {
      '1': 'CreateSkill',
      '2': '.agents.v1.CreateSkillRequest',
      '3': '.agents.v1.CreateSkillResponse'
    },
    {
      '1': 'UpdateSkill',
      '2': '.agents.v1.UpdateSkillRequest',
      '3': '.agents.v1.UpdateSkillResponse'
    },
    {
      '1': 'DeleteSkill',
      '2': '.agents.v1.DeleteSkillRequest',
      '3': '.agents.v1.DeleteSkillResponse'
    },
    {
      '1': 'ListSkillResources',
      '2': '.agents.v1.ListSkillResourcesRequest',
      '3': '.agents.v1.ListSkillResourcesResponse'
    },
    {
      '1': 'GetSkillResource',
      '2': '.agents.v1.GetSkillResourceRequest',
      '3': '.agents.v1.GetSkillResourceResponse'
    },
    {
      '1': 'PutSkillResource',
      '2': '.agents.v1.PutSkillResourceRequest',
      '3': '.agents.v1.PutSkillResourceResponse'
    },
    {
      '1': 'DeleteSkillResource',
      '2': '.agents.v1.DeleteSkillResourceRequest',
      '3': '.agents.v1.DeleteSkillResourceResponse'
    },
  ],
};

@$core.Deprecated('Use skillServiceDescriptor instead')
const $core.Map<$core.String, $core.Map<$core.String, $core.dynamic>>
    SkillServiceBase$messageJson = {
  '.agents.v1.ListSkillsRequest': ListSkillsRequest$json,
  '.agents.v1.ListSkillsResponse': ListSkillsResponse$json,
  '.agents.v1.Skill': Skill$json,
  '.agents.v1.Skill.MetadataEntry': Skill_MetadataEntry$json,
  '.google.protobuf.Timestamp': $0.Timestamp$json,
  '.agents.v1.GetSkillRequest': GetSkillRequest$json,
  '.agents.v1.GetSkillResponse': GetSkillResponse$json,
  '.agents.v1.CreateSkillRequest': CreateSkillRequest$json,
  '.agents.v1.CreateSkillResponse': CreateSkillResponse$json,
  '.agents.v1.UpdateSkillRequest': UpdateSkillRequest$json,
  '.agents.v1.UpdateSkillResponse': UpdateSkillResponse$json,
  '.agents.v1.DeleteSkillRequest': DeleteSkillRequest$json,
  '.agents.v1.DeleteSkillResponse': DeleteSkillResponse$json,
  '.agents.v1.ListSkillResourcesRequest': ListSkillResourcesRequest$json,
  '.agents.v1.ListSkillResourcesResponse': ListSkillResourcesResponse$json,
  '.agents.v1.SkillResource': SkillResource$json,
  '.agents.v1.GetSkillResourceRequest': GetSkillResourceRequest$json,
  '.agents.v1.GetSkillResourceResponse': GetSkillResourceResponse$json,
  '.agents.v1.PutSkillResourceRequest': PutSkillResourceRequest$json,
  '.agents.v1.PutSkillResourceResponse': PutSkillResourceResponse$json,
  '.agents.v1.DeleteSkillResourceRequest': DeleteSkillResourceRequest$json,
  '.agents.v1.DeleteSkillResourceResponse': DeleteSkillResourceResponse$json,
};

/// Descriptor for `SkillService`. Decode as a `google.protobuf.ServiceDescriptorProto`.
final $typed_data.Uint8List skillServiceDescriptor = $convert.base64Decode(
    'CgxTa2lsbFNlcnZpY2USSQoKTGlzdFNraWxscxIcLmFnZW50cy52MS5MaXN0U2tpbGxzUmVxdW'
    'VzdBodLmFnZW50cy52MS5MaXN0U2tpbGxzUmVzcG9uc2USQwoIR2V0U2tpbGwSGi5hZ2VudHMu'
    'djEuR2V0U2tpbGxSZXF1ZXN0GhsuYWdlbnRzLnYxLkdldFNraWxsUmVzcG9uc2USTAoLQ3JlYX'
    'RlU2tpbGwSHS5hZ2VudHMudjEuQ3JlYXRlU2tpbGxSZXF1ZXN0Gh4uYWdlbnRzLnYxLkNyZWF0'
    'ZVNraWxsUmVzcG9uc2USTAoLVXBkYXRlU2tpbGwSHS5hZ2VudHMudjEuVXBkYXRlU2tpbGxSZX'
    'F1ZXN0Gh4uYWdlbnRzLnYxLlVwZGF0ZVNraWxsUmVzcG9uc2USTAoLRGVsZXRlU2tpbGwSHS5h'
    'Z2VudHMudjEuRGVsZXRlU2tpbGxSZXF1ZXN0Gh4uYWdlbnRzLnYxLkRlbGV0ZVNraWxsUmVzcG'
    '9uc2USYQoSTGlzdFNraWxsUmVzb3VyY2VzEiQuYWdlbnRzLnYxLkxpc3RTa2lsbFJlc291cmNl'
    'c1JlcXVlc3QaJS5hZ2VudHMudjEuTGlzdFNraWxsUmVzb3VyY2VzUmVzcG9uc2USWwoQR2V0U2'
    'tpbGxSZXNvdXJjZRIiLmFnZW50cy52MS5HZXRTa2lsbFJlc291cmNlUmVxdWVzdBojLmFnZW50'
    'cy52MS5HZXRTa2lsbFJlc291cmNlUmVzcG9uc2USWwoQUHV0U2tpbGxSZXNvdXJjZRIiLmFnZW'
    '50cy52MS5QdXRTa2lsbFJlc291cmNlUmVxdWVzdBojLmFnZW50cy52MS5QdXRTa2lsbFJlc291'
    'cmNlUmVzcG9uc2USZAoTRGVsZXRlU2tpbGxSZXNvdXJjZRIlLmFnZW50cy52MS5EZWxldGVTa2'
    'lsbFJlc291cmNlUmVxdWVzdBomLmFnZW50cy52MS5EZWxldGVTa2lsbFJlc291cmNlUmVzcG9u'
    'c2U=');
