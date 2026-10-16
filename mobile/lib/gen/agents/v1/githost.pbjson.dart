// This is a generated file - do not edit.
//
// Generated from agents/v1/githost.proto.

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

@$core.Deprecated('Use gitHostKindDescriptor instead')
const GitHostKind$json = {
  '1': 'GitHostKind',
  '2': [
    {'1': 'GIT_HOST_KIND_UNSPECIFIED', '2': 0},
    {'1': 'GIT_HOST_KIND_GITHUB', '2': 1},
    {'1': 'GIT_HOST_KIND_GITLAB', '2': 2},
  ],
};

/// Descriptor for `GitHostKind`. Decode as a `google.protobuf.EnumDescriptorProto`.
final $typed_data.Uint8List gitHostKindDescriptor = $convert.base64Decode(
    'CgtHaXRIb3N0S2luZBIdChlHSVRfSE9TVF9LSU5EX1VOU1BFQ0lGSUVEEAASGAoUR0lUX0hPU1'
    'RfS0lORF9HSVRIVUIQARIYChRHSVRfSE9TVF9LSU5EX0dJVExBQhAC');

@$core.Deprecated('Use gitHostDescriptor instead')
const GitHost$json = {
  '1': 'GitHost',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
    {'1': 'name', '3': 2, '4': 1, '5': 9, '8': {}, '10': 'name'},
    {
      '1': 'kind',
      '3': 3,
      '4': 1,
      '5': 14,
      '6': '.agents.v1.GitHostKind',
      '10': 'kind'
    },
    {'1': 'api_base_url', '3': 4, '4': 1, '5': 9, '8': {}, '10': 'apiBaseUrl'},
    {'1': 'web_base_url', '3': 5, '4': 1, '5': 9, '10': 'webBaseUrl'},
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

/// Descriptor for `GitHost`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List gitHostDescriptor = $convert.base64Decode(
    'CgdHaXRIb3N0Eg4KAmlkGAEgASgJUgJpZBIbCgRuYW1lGAIgASgJQgf6QgRyAhABUgRuYW1lEi'
    'oKBGtpbmQYAyABKA4yFi5hZ2VudHMudjEuR2l0SG9zdEtpbmRSBGtpbmQSKQoMYXBpX2Jhc2Vf'
    'dXJsGAQgASgJQgf6QgRyAhABUgphcGlCYXNlVXJsEiAKDHdlYl9iYXNlX3VybBgFIAEoCVIKd2'
    'ViQmFzZVVybBI5CgpjcmVhdGVkX2F0GAYgASgLMhouZ29vZ2xlLnByb3RvYnVmLlRpbWVzdGFt'
    'cFIJY3JlYXRlZEF0EjkKCnVwZGF0ZWRfYXQYByABKAsyGi5nb29nbGUucHJvdG9idWYuVGltZX'
    'N0YW1wUgl1cGRhdGVkQXQ=');

@$core.Deprecated('Use listGitHostsRequestDescriptor instead')
const ListGitHostsRequest$json = {
  '1': 'ListGitHostsRequest',
};

/// Descriptor for `ListGitHostsRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listGitHostsRequestDescriptor =
    $convert.base64Decode('ChNMaXN0R2l0SG9zdHNSZXF1ZXN0');

@$core.Deprecated('Use listGitHostsResponseDescriptor instead')
const ListGitHostsResponse$json = {
  '1': 'ListGitHostsResponse',
  '2': [
    {
      '1': 'hosts',
      '3': 1,
      '4': 3,
      '5': 11,
      '6': '.agents.v1.GitHost',
      '10': 'hosts'
    },
  ],
};

/// Descriptor for `ListGitHostsResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listGitHostsResponseDescriptor = $convert.base64Decode(
    'ChRMaXN0R2l0SG9zdHNSZXNwb25zZRIoCgVob3N0cxgBIAMoCzISLmFnZW50cy52MS5HaXRIb3'
    'N0UgVob3N0cw==');

@$core.Deprecated('Use getGitHostRequestDescriptor instead')
const GetGitHostRequest$json = {
  '1': 'GetGitHostRequest',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '8': {}, '10': 'id'},
  ],
};

/// Descriptor for `GetGitHostRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getGitHostRequestDescriptor = $convert.base64Decode(
    'ChFHZXRHaXRIb3N0UmVxdWVzdBIXCgJpZBgBIAEoCUIH+kIEcgIQAVICaWQ=');

@$core.Deprecated('Use getGitHostResponseDescriptor instead')
const GetGitHostResponse$json = {
  '1': 'GetGitHostResponse',
  '2': [
    {
      '1': 'host',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.GitHost',
      '10': 'host'
    },
  ],
};

/// Descriptor for `GetGitHostResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getGitHostResponseDescriptor = $convert.base64Decode(
    'ChJHZXRHaXRIb3N0UmVzcG9uc2USJgoEaG9zdBgBIAEoCzISLmFnZW50cy52MS5HaXRIb3N0Ug'
    'Rob3N0');

@$core.Deprecated('Use createGitHostRequestDescriptor instead')
const CreateGitHostRequest$json = {
  '1': 'CreateGitHostRequest',
  '2': [
    {
      '1': 'host',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.GitHost',
      '10': 'host'
    },
  ],
};

/// Descriptor for `CreateGitHostRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List createGitHostRequestDescriptor = $convert.base64Decode(
    'ChRDcmVhdGVHaXRIb3N0UmVxdWVzdBImCgRob3N0GAEgASgLMhIuYWdlbnRzLnYxLkdpdEhvc3'
    'RSBGhvc3Q=');

@$core.Deprecated('Use createGitHostResponseDescriptor instead')
const CreateGitHostResponse$json = {
  '1': 'CreateGitHostResponse',
  '2': [
    {
      '1': 'host',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.GitHost',
      '10': 'host'
    },
  ],
};

/// Descriptor for `CreateGitHostResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List createGitHostResponseDescriptor = $convert.base64Decode(
    'ChVDcmVhdGVHaXRIb3N0UmVzcG9uc2USJgoEaG9zdBgBIAEoCzISLmFnZW50cy52MS5HaXRIb3'
    'N0UgRob3N0');

@$core.Deprecated('Use updateGitHostRequestDescriptor instead')
const UpdateGitHostRequest$json = {
  '1': 'UpdateGitHostRequest',
  '2': [
    {
      '1': 'host',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.GitHost',
      '10': 'host'
    },
  ],
};

/// Descriptor for `UpdateGitHostRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List updateGitHostRequestDescriptor = $convert.base64Decode(
    'ChRVcGRhdGVHaXRIb3N0UmVxdWVzdBImCgRob3N0GAEgASgLMhIuYWdlbnRzLnYxLkdpdEhvc3'
    'RSBGhvc3Q=');

@$core.Deprecated('Use updateGitHostResponseDescriptor instead')
const UpdateGitHostResponse$json = {
  '1': 'UpdateGitHostResponse',
  '2': [
    {
      '1': 'host',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.GitHost',
      '10': 'host'
    },
  ],
};

/// Descriptor for `UpdateGitHostResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List updateGitHostResponseDescriptor = $convert.base64Decode(
    'ChVVcGRhdGVHaXRIb3N0UmVzcG9uc2USJgoEaG9zdBgBIAEoCzISLmFnZW50cy52MS5HaXRIb3'
    'N0UgRob3N0');

@$core.Deprecated('Use deleteGitHostRequestDescriptor instead')
const DeleteGitHostRequest$json = {
  '1': 'DeleteGitHostRequest',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '8': {}, '10': 'id'},
  ],
};

/// Descriptor for `DeleteGitHostRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List deleteGitHostRequestDescriptor =
    $convert.base64Decode(
        'ChREZWxldGVHaXRIb3N0UmVxdWVzdBIXCgJpZBgBIAEoCUIH+kIEcgIQAVICaWQ=');

@$core.Deprecated('Use deleteGitHostResponseDescriptor instead')
const DeleteGitHostResponse$json = {
  '1': 'DeleteGitHostResponse',
};

/// Descriptor for `DeleteGitHostResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List deleteGitHostResponseDescriptor =
    $convert.base64Decode('ChVEZWxldGVHaXRIb3N0UmVzcG9uc2U=');

const $core.Map<$core.String, $core.dynamic> GitHostServiceBase$json = {
  '1': 'GitHostService',
  '2': [
    {
      '1': 'ListGitHosts',
      '2': '.agents.v1.ListGitHostsRequest',
      '3': '.agents.v1.ListGitHostsResponse'
    },
    {
      '1': 'GetGitHost',
      '2': '.agents.v1.GetGitHostRequest',
      '3': '.agents.v1.GetGitHostResponse'
    },
    {
      '1': 'CreateGitHost',
      '2': '.agents.v1.CreateGitHostRequest',
      '3': '.agents.v1.CreateGitHostResponse'
    },
    {
      '1': 'UpdateGitHost',
      '2': '.agents.v1.UpdateGitHostRequest',
      '3': '.agents.v1.UpdateGitHostResponse'
    },
    {
      '1': 'DeleteGitHost',
      '2': '.agents.v1.DeleteGitHostRequest',
      '3': '.agents.v1.DeleteGitHostResponse'
    },
  ],
};

@$core.Deprecated('Use gitHostServiceDescriptor instead')
const $core.Map<$core.String, $core.Map<$core.String, $core.dynamic>>
    GitHostServiceBase$messageJson = {
  '.agents.v1.ListGitHostsRequest': ListGitHostsRequest$json,
  '.agents.v1.ListGitHostsResponse': ListGitHostsResponse$json,
  '.agents.v1.GitHost': GitHost$json,
  '.google.protobuf.Timestamp': $0.Timestamp$json,
  '.agents.v1.GetGitHostRequest': GetGitHostRequest$json,
  '.agents.v1.GetGitHostResponse': GetGitHostResponse$json,
  '.agents.v1.CreateGitHostRequest': CreateGitHostRequest$json,
  '.agents.v1.CreateGitHostResponse': CreateGitHostResponse$json,
  '.agents.v1.UpdateGitHostRequest': UpdateGitHostRequest$json,
  '.agents.v1.UpdateGitHostResponse': UpdateGitHostResponse$json,
  '.agents.v1.DeleteGitHostRequest': DeleteGitHostRequest$json,
  '.agents.v1.DeleteGitHostResponse': DeleteGitHostResponse$json,
};

/// Descriptor for `GitHostService`. Decode as a `google.protobuf.ServiceDescriptorProto`.
final $typed_data.Uint8List gitHostServiceDescriptor = $convert.base64Decode(
    'Cg5HaXRIb3N0U2VydmljZRJPCgxMaXN0R2l0SG9zdHMSHi5hZ2VudHMudjEuTGlzdEdpdEhvc3'
    'RzUmVxdWVzdBofLmFnZW50cy52MS5MaXN0R2l0SG9zdHNSZXNwb25zZRJJCgpHZXRHaXRIb3N0'
    'EhwuYWdlbnRzLnYxLkdldEdpdEhvc3RSZXF1ZXN0Gh0uYWdlbnRzLnYxLkdldEdpdEhvc3RSZX'
    'Nwb25zZRJSCg1DcmVhdGVHaXRIb3N0Eh8uYWdlbnRzLnYxLkNyZWF0ZUdpdEhvc3RSZXF1ZXN0'
    'GiAuYWdlbnRzLnYxLkNyZWF0ZUdpdEhvc3RSZXNwb25zZRJSCg1VcGRhdGVHaXRIb3N0Eh8uYW'
    'dlbnRzLnYxLlVwZGF0ZUdpdEhvc3RSZXF1ZXN0GiAuYWdlbnRzLnYxLlVwZGF0ZUdpdEhvc3RS'
    'ZXNwb25zZRJSCg1EZWxldGVHaXRIb3N0Eh8uYWdlbnRzLnYxLkRlbGV0ZUdpdEhvc3RSZXF1ZX'
    'N0GiAuYWdlbnRzLnYxLkRlbGV0ZUdpdEhvc3RSZXNwb25zZQ==');
