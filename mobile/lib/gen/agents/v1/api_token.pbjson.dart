// This is a generated file - do not edit.
//
// Generated from agents/v1/api_token.proto.

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

@$core.Deprecated('Use aPITokenKindDescriptor instead')
const APITokenKind$json = {
  '1': 'APITokenKind',
  '2': [
    {'1': 'API_TOKEN_KIND_UNSPECIFIED', '2': 0},
    {'1': 'API_TOKEN_KIND_USER', '2': 1},
    {'1': 'API_TOKEN_KIND_DAEMON', '2': 2},
  ],
};

/// Descriptor for `APITokenKind`. Decode as a `google.protobuf.EnumDescriptorProto`.
final $typed_data.Uint8List aPITokenKindDescriptor = $convert.base64Decode(
    'CgxBUElUb2tlbktpbmQSHgoaQVBJX1RPS0VOX0tJTkRfVU5TUEVDSUZJRUQQABIXChNBUElfVE'
    '9LRU5fS0lORF9VU0VSEAESGQoVQVBJX1RPS0VOX0tJTkRfREFFTU9OEAI=');

@$core.Deprecated('Use aPITokenDescriptor instead')
const APIToken$json = {
  '1': 'APIToken',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
    {'1': 'name', '3': 2, '4': 1, '5': 9, '10': 'name'},
    {'1': 'prefix', '3': 3, '4': 1, '5': 9, '10': 'prefix'},
    {
      '1': 'created_at',
      '3': 4,
      '4': 1,
      '5': 11,
      '6': '.google.protobuf.Timestamp',
      '10': 'createdAt'
    },
    {
      '1': 'last_used_at',
      '3': 5,
      '4': 1,
      '5': 11,
      '6': '.google.protobuf.Timestamp',
      '10': 'lastUsedAt'
    },
    {'1': 'revoked', '3': 6, '4': 1, '5': 8, '10': 'revoked'},
    {
      '1': 'kind',
      '3': 7,
      '4': 1,
      '5': 14,
      '6': '.agents.v1.APITokenKind',
      '10': 'kind'
    },
    {'1': 'scopes', '3': 8, '4': 3, '5': 9, '10': 'scopes'},
    {
      '1': 'expires_at',
      '3': 9,
      '4': 1,
      '5': 11,
      '6': '.google.protobuf.Timestamp',
      '10': 'expiresAt'
    },
    {
      '1': 'daemon_runtime_id',
      '3': 10,
      '4': 1,
      '5': 9,
      '10': 'daemonRuntimeId'
    },
    {'1': 'workspace_id', '3': 100, '4': 1, '5': 9, '10': 'workspaceId'},
  ],
};

/// Descriptor for `APIToken`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List aPITokenDescriptor = $convert.base64Decode(
    'CghBUElUb2tlbhIOCgJpZBgBIAEoCVICaWQSEgoEbmFtZRgCIAEoCVIEbmFtZRIWCgZwcmVmaX'
    'gYAyABKAlSBnByZWZpeBI5CgpjcmVhdGVkX2F0GAQgASgLMhouZ29vZ2xlLnByb3RvYnVmLlRp'
    'bWVzdGFtcFIJY3JlYXRlZEF0EjwKDGxhc3RfdXNlZF9hdBgFIAEoCzIaLmdvb2dsZS5wcm90b2'
    'J1Zi5UaW1lc3RhbXBSCmxhc3RVc2VkQXQSGAoHcmV2b2tlZBgGIAEoCFIHcmV2b2tlZBIrCgRr'
    'aW5kGAcgASgOMhcuYWdlbnRzLnYxLkFQSVRva2VuS2luZFIEa2luZBIWCgZzY29wZXMYCCADKA'
    'lSBnNjb3BlcxI5CgpleHBpcmVzX2F0GAkgASgLMhouZ29vZ2xlLnByb3RvYnVmLlRpbWVzdGFt'
    'cFIJZXhwaXJlc0F0EioKEWRhZW1vbl9ydW50aW1lX2lkGAogASgJUg9kYWVtb25SdW50aW1lSW'
    'QSIQoMd29ya3NwYWNlX2lkGGQgASgJUgt3b3Jrc3BhY2VJZA==');

@$core.Deprecated('Use listAPITokensRequestDescriptor instead')
const ListAPITokensRequest$json = {
  '1': 'ListAPITokensRequest',
};

/// Descriptor for `ListAPITokensRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listAPITokensRequestDescriptor =
    $convert.base64Decode('ChRMaXN0QVBJVG9rZW5zUmVxdWVzdA==');

@$core.Deprecated('Use listAPITokensResponseDescriptor instead')
const ListAPITokensResponse$json = {
  '1': 'ListAPITokensResponse',
  '2': [
    {
      '1': 'tokens',
      '3': 1,
      '4': 3,
      '5': 11,
      '6': '.agents.v1.APIToken',
      '10': 'tokens'
    },
  ],
};

/// Descriptor for `ListAPITokensResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listAPITokensResponseDescriptor = $convert.base64Decode(
    'ChVMaXN0QVBJVG9rZW5zUmVzcG9uc2USKwoGdG9rZW5zGAEgAygLMhMuYWdlbnRzLnYxLkFQSV'
    'Rva2VuUgZ0b2tlbnM=');

@$core.Deprecated('Use createAPITokenRequestDescriptor instead')
const CreateAPITokenRequest$json = {
  '1': 'CreateAPITokenRequest',
  '2': [
    {'1': 'name', '3': 1, '4': 1, '5': 9, '8': {}, '10': 'name'},
  ],
};

/// Descriptor for `CreateAPITokenRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List createAPITokenRequestDescriptor = $convert.base64Decode(
    'ChVDcmVhdGVBUElUb2tlblJlcXVlc3QSGwoEbmFtZRgBIAEoCUIH+kIEcgIQAVIEbmFtZQ==');

@$core.Deprecated('Use createAPITokenResponseDescriptor instead')
const CreateAPITokenResponse$json = {
  '1': 'CreateAPITokenResponse',
  '2': [
    {
      '1': 'token',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.APIToken',
      '10': 'token'
    },
    {'1': 'secret', '3': 2, '4': 1, '5': 9, '10': 'secret'},
  ],
};

/// Descriptor for `CreateAPITokenResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List createAPITokenResponseDescriptor =
    $convert.base64Decode(
        'ChZDcmVhdGVBUElUb2tlblJlc3BvbnNlEikKBXRva2VuGAEgASgLMhMuYWdlbnRzLnYxLkFQSV'
        'Rva2VuUgV0b2tlbhIWCgZzZWNyZXQYAiABKAlSBnNlY3JldA==');

@$core.Deprecated('Use revokeAPITokenRequestDescriptor instead')
const RevokeAPITokenRequest$json = {
  '1': 'RevokeAPITokenRequest',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '8': {}, '10': 'id'},
  ],
};

/// Descriptor for `RevokeAPITokenRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List revokeAPITokenRequestDescriptor =
    $convert.base64Decode(
        'ChVSZXZva2VBUElUb2tlblJlcXVlc3QSFwoCaWQYASABKAlCB/pCBHICEAFSAmlk');

@$core.Deprecated('Use revokeAPITokenResponseDescriptor instead')
const RevokeAPITokenResponse$json = {
  '1': 'RevokeAPITokenResponse',
  '2': [
    {
      '1': 'token',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.APIToken',
      '10': 'token'
    },
  ],
};

/// Descriptor for `RevokeAPITokenResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List revokeAPITokenResponseDescriptor =
    $convert.base64Decode(
        'ChZSZXZva2VBUElUb2tlblJlc3BvbnNlEikKBXRva2VuGAEgASgLMhMuYWdlbnRzLnYxLkFQSV'
        'Rva2VuUgV0b2tlbg==');

const $core.Map<$core.String, $core.dynamic> APITokenServiceBase$json = {
  '1': 'APITokenService',
  '2': [
    {
      '1': 'ListAPITokens',
      '2': '.agents.v1.ListAPITokensRequest',
      '3': '.agents.v1.ListAPITokensResponse'
    },
    {
      '1': 'CreateAPIToken',
      '2': '.agents.v1.CreateAPITokenRequest',
      '3': '.agents.v1.CreateAPITokenResponse'
    },
    {
      '1': 'RevokeAPIToken',
      '2': '.agents.v1.RevokeAPITokenRequest',
      '3': '.agents.v1.RevokeAPITokenResponse'
    },
  ],
};

@$core.Deprecated('Use aPITokenServiceDescriptor instead')
const $core.Map<$core.String, $core.Map<$core.String, $core.dynamic>>
    APITokenServiceBase$messageJson = {
  '.agents.v1.ListAPITokensRequest': ListAPITokensRequest$json,
  '.agents.v1.ListAPITokensResponse': ListAPITokensResponse$json,
  '.agents.v1.APIToken': APIToken$json,
  '.google.protobuf.Timestamp': $0.Timestamp$json,
  '.agents.v1.CreateAPITokenRequest': CreateAPITokenRequest$json,
  '.agents.v1.CreateAPITokenResponse': CreateAPITokenResponse$json,
  '.agents.v1.RevokeAPITokenRequest': RevokeAPITokenRequest$json,
  '.agents.v1.RevokeAPITokenResponse': RevokeAPITokenResponse$json,
};

/// Descriptor for `APITokenService`. Decode as a `google.protobuf.ServiceDescriptorProto`.
final $typed_data.Uint8List aPITokenServiceDescriptor = $convert.base64Decode(
    'Cg9BUElUb2tlblNlcnZpY2USUgoNTGlzdEFQSVRva2VucxIfLmFnZW50cy52MS5MaXN0QVBJVG'
    '9rZW5zUmVxdWVzdBogLmFnZW50cy52MS5MaXN0QVBJVG9rZW5zUmVzcG9uc2USVQoOQ3JlYXRl'
    'QVBJVG9rZW4SIC5hZ2VudHMudjEuQ3JlYXRlQVBJVG9rZW5SZXF1ZXN0GiEuYWdlbnRzLnYxLk'
    'NyZWF0ZUFQSVRva2VuUmVzcG9uc2USVQoOUmV2b2tlQVBJVG9rZW4SIC5hZ2VudHMudjEuUmV2'
    'b2tlQVBJVG9rZW5SZXF1ZXN0GiEuYWdlbnRzLnYxLlJldm9rZUFQSVRva2VuUmVzcG9uc2U=');
