// This is a generated file - do not edit.
//
// Generated from agents/v1/butterbox.proto.

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

@$core.Deprecated('Use butterBoxDescriptor instead')
const ButterBox$json = {
  '1': 'ButterBox',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
    {'1': 'name', '3': 2, '4': 1, '5': 9, '8': {}, '10': 'name'},
    {'1': 'base_url', '3': 3, '4': 1, '5': 9, '8': {}, '10': 'baseUrl'},
    {'1': 'enabled', '3': 4, '4': 1, '5': 8, '10': 'enabled'},
    {'1': 'credential_set', '3': 5, '4': 1, '5': 8, '10': 'credentialSet'},
    {
      '1': 'credential_updated_at',
      '3': 6,
      '4': 1,
      '5': 11,
      '6': '.google.protobuf.Timestamp',
      '10': 'credentialUpdatedAt'
    },
    {
      '1': 'created_at',
      '3': 7,
      '4': 1,
      '5': 11,
      '6': '.google.protobuf.Timestamp',
      '10': 'createdAt'
    },
    {
      '1': 'updated_at',
      '3': 8,
      '4': 1,
      '5': 11,
      '6': '.google.protobuf.Timestamp',
      '10': 'updatedAt'
    },
    {'1': 'workspace_id', '3': 100, '4': 1, '5': 9, '10': 'workspaceId'},
  ],
};

/// Descriptor for `ButterBox`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List butterBoxDescriptor = $convert.base64Decode(
    'CglCdXR0ZXJCb3gSDgoCaWQYASABKAlSAmlkEhsKBG5hbWUYAiABKAlCB/pCBHICEAFSBG5hbW'
    'USIgoIYmFzZV91cmwYAyABKAlCB/pCBHICEAFSB2Jhc2VVcmwSGAoHZW5hYmxlZBgEIAEoCFIH'
    'ZW5hYmxlZBIlCg5jcmVkZW50aWFsX3NldBgFIAEoCFINY3JlZGVudGlhbFNldBJOChVjcmVkZW'
    '50aWFsX3VwZGF0ZWRfYXQYBiABKAsyGi5nb29nbGUucHJvdG9idWYuVGltZXN0YW1wUhNjcmVk'
    'ZW50aWFsVXBkYXRlZEF0EjkKCmNyZWF0ZWRfYXQYByABKAsyGi5nb29nbGUucHJvdG9idWYuVG'
    'ltZXN0YW1wUgljcmVhdGVkQXQSOQoKdXBkYXRlZF9hdBgIIAEoCzIaLmdvb2dsZS5wcm90b2J1'
    'Zi5UaW1lc3RhbXBSCXVwZGF0ZWRBdBIhCgx3b3Jrc3BhY2VfaWQYZCABKAlSC3dvcmtzcGFjZU'
    'lk');

@$core.Deprecated('Use listButterBoxesRequestDescriptor instead')
const ListButterBoxesRequest$json = {
  '1': 'ListButterBoxesRequest',
};

/// Descriptor for `ListButterBoxesRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listButterBoxesRequestDescriptor =
    $convert.base64Decode('ChZMaXN0QnV0dGVyQm94ZXNSZXF1ZXN0');

@$core.Deprecated('Use listButterBoxesResponseDescriptor instead')
const ListButterBoxesResponse$json = {
  '1': 'ListButterBoxesResponse',
  '2': [
    {
      '1': 'butter_boxes',
      '3': 1,
      '4': 3,
      '5': 11,
      '6': '.agents.v1.ButterBox',
      '10': 'butterBoxes'
    },
  ],
};

/// Descriptor for `ListButterBoxesResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listButterBoxesResponseDescriptor =
    $convert.base64Decode(
        'ChdMaXN0QnV0dGVyQm94ZXNSZXNwb25zZRI3CgxidXR0ZXJfYm94ZXMYASADKAsyFC5hZ2VudH'
        'MudjEuQnV0dGVyQm94UgtidXR0ZXJCb3hlcw==');

@$core.Deprecated('Use getButterBoxRequestDescriptor instead')
const GetButterBoxRequest$json = {
  '1': 'GetButterBoxRequest',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '8': {}, '10': 'id'},
  ],
};

/// Descriptor for `GetButterBoxRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getButterBoxRequestDescriptor =
    $convert.base64Decode(
        'ChNHZXRCdXR0ZXJCb3hSZXF1ZXN0EhcKAmlkGAEgASgJQgf6QgRyAhABUgJpZA==');

@$core.Deprecated('Use getButterBoxResponseDescriptor instead')
const GetButterBoxResponse$json = {
  '1': 'GetButterBoxResponse',
  '2': [
    {
      '1': 'butter_box',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.ButterBox',
      '10': 'butterBox'
    },
  ],
};

/// Descriptor for `GetButterBoxResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getButterBoxResponseDescriptor = $convert.base64Decode(
    'ChRHZXRCdXR0ZXJCb3hSZXNwb25zZRIzCgpidXR0ZXJfYm94GAEgASgLMhQuYWdlbnRzLnYxLk'
    'J1dHRlckJveFIJYnV0dGVyQm94');

@$core.Deprecated('Use createButterBoxRequestDescriptor instead')
const CreateButterBoxRequest$json = {
  '1': 'CreateButterBoxRequest',
  '2': [
    {'1': 'name', '3': 1, '4': 1, '5': 9, '8': {}, '10': 'name'},
    {'1': 'base_url', '3': 2, '4': 1, '5': 9, '8': {}, '10': 'baseUrl'},
    {'1': 'enabled', '3': 3, '4': 1, '5': 8, '10': 'enabled'},
    {'1': 'token', '3': 4, '4': 1, '5': 9, '10': 'token'},
  ],
};

/// Descriptor for `CreateButterBoxRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List createButterBoxRequestDescriptor = $convert.base64Decode(
    'ChZDcmVhdGVCdXR0ZXJCb3hSZXF1ZXN0EhsKBG5hbWUYASABKAlCB/pCBHICEAFSBG5hbWUSIg'
    'oIYmFzZV91cmwYAiABKAlCB/pCBHICEAFSB2Jhc2VVcmwSGAoHZW5hYmxlZBgDIAEoCFIHZW5h'
    'YmxlZBIUCgV0b2tlbhgEIAEoCVIFdG9rZW4=');

@$core.Deprecated('Use createButterBoxResponseDescriptor instead')
const CreateButterBoxResponse$json = {
  '1': 'CreateButterBoxResponse',
  '2': [
    {
      '1': 'butter_box',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.ButterBox',
      '10': 'butterBox'
    },
  ],
};

/// Descriptor for `CreateButterBoxResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List createButterBoxResponseDescriptor =
    $convert.base64Decode(
        'ChdDcmVhdGVCdXR0ZXJCb3hSZXNwb25zZRIzCgpidXR0ZXJfYm94GAEgASgLMhQuYWdlbnRzLn'
        'YxLkJ1dHRlckJveFIJYnV0dGVyQm94');

@$core.Deprecated('Use updateButterBoxRequestDescriptor instead')
const UpdateButterBoxRequest$json = {
  '1': 'UpdateButterBoxRequest',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '8': {}, '10': 'id'},
    {'1': 'name', '3': 2, '4': 1, '5': 9, '8': {}, '10': 'name'},
    {'1': 'base_url', '3': 3, '4': 1, '5': 9, '8': {}, '10': 'baseUrl'},
    {'1': 'enabled', '3': 4, '4': 1, '5': 8, '10': 'enabled'},
  ],
};

/// Descriptor for `UpdateButterBoxRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List updateButterBoxRequestDescriptor = $convert.base64Decode(
    'ChZVcGRhdGVCdXR0ZXJCb3hSZXF1ZXN0EhcKAmlkGAEgASgJQgf6QgRyAhABUgJpZBIbCgRuYW'
    '1lGAIgASgJQgf6QgRyAhABUgRuYW1lEiIKCGJhc2VfdXJsGAMgASgJQgf6QgRyAhABUgdiYXNl'
    'VXJsEhgKB2VuYWJsZWQYBCABKAhSB2VuYWJsZWQ=');

@$core.Deprecated('Use updateButterBoxResponseDescriptor instead')
const UpdateButterBoxResponse$json = {
  '1': 'UpdateButterBoxResponse',
  '2': [
    {
      '1': 'butter_box',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.ButterBox',
      '10': 'butterBox'
    },
  ],
};

/// Descriptor for `UpdateButterBoxResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List updateButterBoxResponseDescriptor =
    $convert.base64Decode(
        'ChdVcGRhdGVCdXR0ZXJCb3hSZXNwb25zZRIzCgpidXR0ZXJfYm94GAEgASgLMhQuYWdlbnRzLn'
        'YxLkJ1dHRlckJveFIJYnV0dGVyQm94');

@$core.Deprecated('Use deleteButterBoxRequestDescriptor instead')
const DeleteButterBoxRequest$json = {
  '1': 'DeleteButterBoxRequest',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '8': {}, '10': 'id'},
  ],
};

/// Descriptor for `DeleteButterBoxRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List deleteButterBoxRequestDescriptor =
    $convert.base64Decode(
        'ChZEZWxldGVCdXR0ZXJCb3hSZXF1ZXN0EhcKAmlkGAEgASgJQgf6QgRyAhABUgJpZA==');

@$core.Deprecated('Use deleteButterBoxResponseDescriptor instead')
const DeleteButterBoxResponse$json = {
  '1': 'DeleteButterBoxResponse',
};

/// Descriptor for `DeleteButterBoxResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List deleteButterBoxResponseDescriptor =
    $convert.base64Decode('ChdEZWxldGVCdXR0ZXJCb3hSZXNwb25zZQ==');

@$core.Deprecated('Use setButterBoxTokenRequestDescriptor instead')
const SetButterBoxTokenRequest$json = {
  '1': 'SetButterBoxTokenRequest',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '8': {}, '10': 'id'},
    {'1': 'token', '3': 2, '4': 1, '5': 9, '10': 'token'},
  ],
};

/// Descriptor for `SetButterBoxTokenRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List setButterBoxTokenRequestDescriptor =
    $convert.base64Decode(
        'ChhTZXRCdXR0ZXJCb3hUb2tlblJlcXVlc3QSFwoCaWQYASABKAlCB/pCBHICEAFSAmlkEhQKBX'
        'Rva2VuGAIgASgJUgV0b2tlbg==');

@$core.Deprecated('Use setButterBoxTokenResponseDescriptor instead')
const SetButterBoxTokenResponse$json = {
  '1': 'SetButterBoxTokenResponse',
  '2': [
    {
      '1': 'butter_box',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.ButterBox',
      '10': 'butterBox'
    },
  ],
};

/// Descriptor for `SetButterBoxTokenResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List setButterBoxTokenResponseDescriptor =
    $convert.base64Decode(
        'ChlTZXRCdXR0ZXJCb3hUb2tlblJlc3BvbnNlEjMKCmJ1dHRlcl9ib3gYASABKAsyFC5hZ2VudH'
        'MudjEuQnV0dGVyQm94UglidXR0ZXJCb3g=');

@$core.Deprecated('Use getButterBoxStatusRequestDescriptor instead')
const GetButterBoxStatusRequest$json = {
  '1': 'GetButterBoxStatusRequest',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '8': {}, '10': 'id'},
  ],
};

/// Descriptor for `GetButterBoxStatusRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getButterBoxStatusRequestDescriptor =
    $convert.base64Decode(
        'ChlHZXRCdXR0ZXJCb3hTdGF0dXNSZXF1ZXN0EhcKAmlkGAEgASgJQgf6QgRyAhABUgJpZA==');

@$core.Deprecated('Use getButterBoxStatusResponseDescriptor instead')
const GetButterBoxStatusResponse$json = {
  '1': 'GetButterBoxStatusResponse',
  '2': [
    {'1': 'reachable', '3': 1, '4': 1, '5': 8, '10': 'reachable'},
    {'1': 'active_sessions', '3': 2, '4': 1, '5': 5, '10': 'activeSessions'},
    {'1': 'error', '3': 3, '4': 1, '5': 9, '10': 'error'},
  ],
};

/// Descriptor for `GetButterBoxStatusResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getButterBoxStatusResponseDescriptor =
    $convert.base64Decode(
        'ChpHZXRCdXR0ZXJCb3hTdGF0dXNSZXNwb25zZRIcCglyZWFjaGFibGUYASABKAhSCXJlYWNoYW'
        'JsZRInCg9hY3RpdmVfc2Vzc2lvbnMYAiABKAVSDmFjdGl2ZVNlc3Npb25zEhQKBWVycm9yGAMg'
        'ASgJUgVlcnJvcg==');

@$core.Deprecated('Use listButterBoxModelsRequestDescriptor instead')
const ListButterBoxModelsRequest$json = {
  '1': 'ListButterBoxModelsRequest',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '8': {}, '10': 'id'},
  ],
};

/// Descriptor for `ListButterBoxModelsRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listButterBoxModelsRequestDescriptor =
    $convert.base64Decode(
        'ChpMaXN0QnV0dGVyQm94TW9kZWxzUmVxdWVzdBIXCgJpZBgBIAEoCUIH+kIEcgIQAVICaWQ=');

@$core.Deprecated('Use listButterBoxModelsResponseDescriptor instead')
const ListButterBoxModelsResponse$json = {
  '1': 'ListButterBoxModelsResponse',
  '2': [
    {
      '1': 'models',
      '3': 1,
      '4': 3,
      '5': 11,
      '6': '.agents.v1.ButterBoxModel',
      '10': 'models'
    },
  ],
};

/// Descriptor for `ListButterBoxModelsResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listButterBoxModelsResponseDescriptor =
    $convert.base64Decode(
        'ChtMaXN0QnV0dGVyQm94TW9kZWxzUmVzcG9uc2USMQoGbW9kZWxzGAEgAygLMhkuYWdlbnRzLn'
        'YxLkJ1dHRlckJveE1vZGVsUgZtb2RlbHM=');

@$core.Deprecated('Use butterBoxModelDescriptor instead')
const ButterBoxModel$json = {
  '1': 'ButterBoxModel',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
    {'1': 'provider', '3': 2, '4': 1, '5': 9, '10': 'provider'},
    {'1': 'name', '3': 3, '4': 1, '5': 9, '10': 'name'},
    {'1': 'reasoning', '3': 4, '4': 1, '5': 8, '10': 'reasoning'},
  ],
};

/// Descriptor for `ButterBoxModel`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List butterBoxModelDescriptor = $convert.base64Decode(
    'Cg5CdXR0ZXJCb3hNb2RlbBIOCgJpZBgBIAEoCVICaWQSGgoIcHJvdmlkZXIYAiABKAlSCHByb3'
    'ZpZGVyEhIKBG5hbWUYAyABKAlSBG5hbWUSHAoJcmVhc29uaW5nGAQgASgIUglyZWFzb25pbmc=');

@$core.Deprecated('Use listButterBoxCursorModelsRequestDescriptor instead')
const ListButterBoxCursorModelsRequest$json = {
  '1': 'ListButterBoxCursorModelsRequest',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '8': {}, '10': 'id'},
  ],
};

/// Descriptor for `ListButterBoxCursorModelsRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listButterBoxCursorModelsRequestDescriptor =
    $convert.base64Decode(
        'CiBMaXN0QnV0dGVyQm94Q3Vyc29yTW9kZWxzUmVxdWVzdBIXCgJpZBgBIAEoCUIH+kIEcgIQAV'
        'ICaWQ=');

@$core.Deprecated('Use listButterBoxCursorModelsResponseDescriptor instead')
const ListButterBoxCursorModelsResponse$json = {
  '1': 'ListButterBoxCursorModelsResponse',
  '2': [
    {
      '1': 'models',
      '3': 1,
      '4': 3,
      '5': 11,
      '6': '.agents.v1.ButterBoxCursorModel',
      '10': 'models'
    },
  ],
};

/// Descriptor for `ListButterBoxCursorModelsResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listButterBoxCursorModelsResponseDescriptor =
    $convert.base64Decode(
        'CiFMaXN0QnV0dGVyQm94Q3Vyc29yTW9kZWxzUmVzcG9uc2USNwoGbW9kZWxzGAEgAygLMh8uYW'
        'dlbnRzLnYxLkJ1dHRlckJveEN1cnNvck1vZGVsUgZtb2RlbHM=');

@$core.Deprecated('Use butterBoxCursorModelDescriptor instead')
const ButterBoxCursorModel$json = {
  '1': 'ButterBoxCursorModel',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
    {'1': 'name', '3': 2, '4': 1, '5': 9, '10': 'name'},
  ],
};

/// Descriptor for `ButterBoxCursorModel`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List butterBoxCursorModelDescriptor = $convert.base64Decode(
    'ChRCdXR0ZXJCb3hDdXJzb3JNb2RlbBIOCgJpZBgBIAEoCVICaWQSEgoEbmFtZRgCIAEoCVIEbm'
    'FtZQ==');

const $core.Map<$core.String, $core.dynamic> ButterBoxServiceBase$json = {
  '1': 'ButterBoxService',
  '2': [
    {
      '1': 'ListButterBoxes',
      '2': '.agents.v1.ListButterBoxesRequest',
      '3': '.agents.v1.ListButterBoxesResponse'
    },
    {
      '1': 'GetButterBox',
      '2': '.agents.v1.GetButterBoxRequest',
      '3': '.agents.v1.GetButterBoxResponse'
    },
    {
      '1': 'CreateButterBox',
      '2': '.agents.v1.CreateButterBoxRequest',
      '3': '.agents.v1.CreateButterBoxResponse'
    },
    {
      '1': 'UpdateButterBox',
      '2': '.agents.v1.UpdateButterBoxRequest',
      '3': '.agents.v1.UpdateButterBoxResponse'
    },
    {
      '1': 'DeleteButterBox',
      '2': '.agents.v1.DeleteButterBoxRequest',
      '3': '.agents.v1.DeleteButterBoxResponse'
    },
    {
      '1': 'SetButterBoxToken',
      '2': '.agents.v1.SetButterBoxTokenRequest',
      '3': '.agents.v1.SetButterBoxTokenResponse'
    },
    {
      '1': 'GetButterBoxStatus',
      '2': '.agents.v1.GetButterBoxStatusRequest',
      '3': '.agents.v1.GetButterBoxStatusResponse'
    },
    {
      '1': 'ListButterBoxModels',
      '2': '.agents.v1.ListButterBoxModelsRequest',
      '3': '.agents.v1.ListButterBoxModelsResponse'
    },
    {
      '1': 'ListButterBoxCursorModels',
      '2': '.agents.v1.ListButterBoxCursorModelsRequest',
      '3': '.agents.v1.ListButterBoxCursorModelsResponse'
    },
  ],
};

@$core.Deprecated('Use butterBoxServiceDescriptor instead')
const $core.Map<$core.String, $core.Map<$core.String, $core.dynamic>>
    ButterBoxServiceBase$messageJson = {
  '.agents.v1.ListButterBoxesRequest': ListButterBoxesRequest$json,
  '.agents.v1.ListButterBoxesResponse': ListButterBoxesResponse$json,
  '.agents.v1.ButterBox': ButterBox$json,
  '.google.protobuf.Timestamp': $0.Timestamp$json,
  '.agents.v1.GetButterBoxRequest': GetButterBoxRequest$json,
  '.agents.v1.GetButterBoxResponse': GetButterBoxResponse$json,
  '.agents.v1.CreateButterBoxRequest': CreateButterBoxRequest$json,
  '.agents.v1.CreateButterBoxResponse': CreateButterBoxResponse$json,
  '.agents.v1.UpdateButterBoxRequest': UpdateButterBoxRequest$json,
  '.agents.v1.UpdateButterBoxResponse': UpdateButterBoxResponse$json,
  '.agents.v1.DeleteButterBoxRequest': DeleteButterBoxRequest$json,
  '.agents.v1.DeleteButterBoxResponse': DeleteButterBoxResponse$json,
  '.agents.v1.SetButterBoxTokenRequest': SetButterBoxTokenRequest$json,
  '.agents.v1.SetButterBoxTokenResponse': SetButterBoxTokenResponse$json,
  '.agents.v1.GetButterBoxStatusRequest': GetButterBoxStatusRequest$json,
  '.agents.v1.GetButterBoxStatusResponse': GetButterBoxStatusResponse$json,
  '.agents.v1.ListButterBoxModelsRequest': ListButterBoxModelsRequest$json,
  '.agents.v1.ListButterBoxModelsResponse': ListButterBoxModelsResponse$json,
  '.agents.v1.ButterBoxModel': ButterBoxModel$json,
  '.agents.v1.ListButterBoxCursorModelsRequest':
      ListButterBoxCursorModelsRequest$json,
  '.agents.v1.ListButterBoxCursorModelsResponse':
      ListButterBoxCursorModelsResponse$json,
  '.agents.v1.ButterBoxCursorModel': ButterBoxCursorModel$json,
};

/// Descriptor for `ButterBoxService`. Decode as a `google.protobuf.ServiceDescriptorProto`.
final $typed_data.Uint8List butterBoxServiceDescriptor = $convert.base64Decode(
    'ChBCdXR0ZXJCb3hTZXJ2aWNlElgKD0xpc3RCdXR0ZXJCb3hlcxIhLmFnZW50cy52MS5MaXN0Qn'
    'V0dGVyQm94ZXNSZXF1ZXN0GiIuYWdlbnRzLnYxLkxpc3RCdXR0ZXJCb3hlc1Jlc3BvbnNlEk8K'
    'DEdldEJ1dHRlckJveBIeLmFnZW50cy52MS5HZXRCdXR0ZXJCb3hSZXF1ZXN0Gh8uYWdlbnRzLn'
    'YxLkdldEJ1dHRlckJveFJlc3BvbnNlElgKD0NyZWF0ZUJ1dHRlckJveBIhLmFnZW50cy52MS5D'
    'cmVhdGVCdXR0ZXJCb3hSZXF1ZXN0GiIuYWdlbnRzLnYxLkNyZWF0ZUJ1dHRlckJveFJlc3Bvbn'
    'NlElgKD1VwZGF0ZUJ1dHRlckJveBIhLmFnZW50cy52MS5VcGRhdGVCdXR0ZXJCb3hSZXF1ZXN0'
    'GiIuYWdlbnRzLnYxLlVwZGF0ZUJ1dHRlckJveFJlc3BvbnNlElgKD0RlbGV0ZUJ1dHRlckJveB'
    'IhLmFnZW50cy52MS5EZWxldGVCdXR0ZXJCb3hSZXF1ZXN0GiIuYWdlbnRzLnYxLkRlbGV0ZUJ1'
    'dHRlckJveFJlc3BvbnNlEl4KEVNldEJ1dHRlckJveFRva2VuEiMuYWdlbnRzLnYxLlNldEJ1dH'
    'RlckJveFRva2VuUmVxdWVzdBokLmFnZW50cy52MS5TZXRCdXR0ZXJCb3hUb2tlblJlc3BvbnNl'
    'EmEKEkdldEJ1dHRlckJveFN0YXR1cxIkLmFnZW50cy52MS5HZXRCdXR0ZXJCb3hTdGF0dXNSZX'
    'F1ZXN0GiUuYWdlbnRzLnYxLkdldEJ1dHRlckJveFN0YXR1c1Jlc3BvbnNlEmQKE0xpc3RCdXR0'
    'ZXJCb3hNb2RlbHMSJS5hZ2VudHMudjEuTGlzdEJ1dHRlckJveE1vZGVsc1JlcXVlc3QaJi5hZ2'
    'VudHMudjEuTGlzdEJ1dHRlckJveE1vZGVsc1Jlc3BvbnNlEnYKGUxpc3RCdXR0ZXJCb3hDdXJz'
    'b3JNb2RlbHMSKy5hZ2VudHMudjEuTGlzdEJ1dHRlckJveEN1cnNvck1vZGVsc1JlcXVlc3QaLC'
    '5hZ2VudHMudjEuTGlzdEJ1dHRlckJveEN1cnNvck1vZGVsc1Jlc3BvbnNl');
