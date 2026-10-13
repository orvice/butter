// This is a generated file - do not edit.
//
// Generated from agents/v1/auth.proto.

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

import 'workspace.pbjson.dart' as $1;

@$core.Deprecated('Use userDescriptor instead')
const User$json = {
  '1': 'User',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
    {'1': 'username', '3': 2, '4': 1, '5': 9, '10': 'username'},
    {'1': 'display_name', '3': 3, '4': 1, '5': 9, '10': 'displayName'},
    {'1': 'role', '3': 4, '4': 1, '5': 9, '10': 'role'},
    {'1': 'disabled', '3': 5, '4': 1, '5': 8, '10': 'disabled'},
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
    {'1': 'avatar_url', '3': 8, '4': 1, '5': 9, '10': 'avatarUrl'},
    {'1': 'email', '3': 9, '4': 1, '5': 9, '10': 'email'},
    {'1': 'provider', '3': 10, '4': 1, '5': 9, '10': 'provider'},
    {'1': 'external_id', '3': 11, '4': 1, '5': 9, '10': 'externalId'},
  ],
};

/// Descriptor for `User`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List userDescriptor = $convert.base64Decode(
    'CgRVc2VyEg4KAmlkGAEgASgJUgJpZBIaCgh1c2VybmFtZRgCIAEoCVIIdXNlcm5hbWUSIQoMZG'
    'lzcGxheV9uYW1lGAMgASgJUgtkaXNwbGF5TmFtZRISCgRyb2xlGAQgASgJUgRyb2xlEhoKCGRp'
    'c2FibGVkGAUgASgIUghkaXNhYmxlZBI5CgpjcmVhdGVkX2F0GAYgASgLMhouZ29vZ2xlLnByb3'
    'RvYnVmLlRpbWVzdGFtcFIJY3JlYXRlZEF0EjkKCnVwZGF0ZWRfYXQYByABKAsyGi5nb29nbGUu'
    'cHJvdG9idWYuVGltZXN0YW1wUgl1cGRhdGVkQXQSHQoKYXZhdGFyX3VybBgIIAEoCVIJYXZhdG'
    'FyVXJsEhQKBWVtYWlsGAkgASgJUgVlbWFpbBIaCghwcm92aWRlchgKIAEoCVIIcHJvdmlkZXIS'
    'HwoLZXh0ZXJuYWxfaWQYCyABKAlSCmV4dGVybmFsSWQ=');

@$core.Deprecated('Use loginRequestDescriptor instead')
const LoginRequest$json = {
  '1': 'LoginRequest',
  '2': [
    {'1': 'username', '3': 1, '4': 1, '5': 9, '8': {}, '10': 'username'},
    {'1': 'password', '3': 2, '4': 1, '5': 9, '8': {}, '10': 'password'},
  ],
};

/// Descriptor for `LoginRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List loginRequestDescriptor = $convert.base64Decode(
    'CgxMb2dpblJlcXVlc3QSIwoIdXNlcm5hbWUYASABKAlCB/pCBHICEAFSCHVzZXJuYW1lEiMKCH'
    'Bhc3N3b3JkGAIgASgJQgf6QgRyAhABUghwYXNzd29yZA==');

@$core.Deprecated('Use loginResponseDescriptor instead')
const LoginResponse$json = {
  '1': 'LoginResponse',
  '2': [
    {'1': 'token', '3': 1, '4': 1, '5': 9, '10': 'token'},
    {
      '1': 'user',
      '3': 2,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.User',
      '10': 'user'
    },
    {
      '1': 'expires_at',
      '3': 3,
      '4': 1,
      '5': 11,
      '6': '.google.protobuf.Timestamp',
      '10': 'expiresAt'
    },
    {
      '1': 'workspaces',
      '3': 4,
      '4': 3,
      '5': 11,
      '6': '.agents.v1.Workspace',
      '10': 'workspaces'
    },
  ],
};

/// Descriptor for `LoginResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List loginResponseDescriptor = $convert.base64Decode(
    'Cg1Mb2dpblJlc3BvbnNlEhQKBXRva2VuGAEgASgJUgV0b2tlbhIjCgR1c2VyGAIgASgLMg8uYW'
    'dlbnRzLnYxLlVzZXJSBHVzZXISOQoKZXhwaXJlc19hdBgDIAEoCzIaLmdvb2dsZS5wcm90b2J1'
    'Zi5UaW1lc3RhbXBSCWV4cGlyZXNBdBI0Cgp3b3Jrc3BhY2VzGAQgAygLMhQuYWdlbnRzLnYxLl'
    'dvcmtzcGFjZVIKd29ya3NwYWNlcw==');

@$core.Deprecated('Use meRequestDescriptor instead')
const MeRequest$json = {
  '1': 'MeRequest',
};

/// Descriptor for `MeRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List meRequestDescriptor =
    $convert.base64Decode('CglNZVJlcXVlc3Q=');

@$core.Deprecated('Use meResponseDescriptor instead')
const MeResponse$json = {
  '1': 'MeResponse',
  '2': [
    {
      '1': 'user',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.User',
      '10': 'user'
    },
  ],
};

/// Descriptor for `MeResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List meResponseDescriptor = $convert.base64Decode(
    'CgpNZVJlc3BvbnNlEiMKBHVzZXIYASABKAsyDy5hZ2VudHMudjEuVXNlclIEdXNlcg==');

@$core.Deprecated('Use logoutRequestDescriptor instead')
const LogoutRequest$json = {
  '1': 'LogoutRequest',
};

/// Descriptor for `LogoutRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List logoutRequestDescriptor =
    $convert.base64Decode('Cg1Mb2dvdXRSZXF1ZXN0');

@$core.Deprecated('Use logoutResponseDescriptor instead')
const LogoutResponse$json = {
  '1': 'LogoutResponse',
};

/// Descriptor for `LogoutResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List logoutResponseDescriptor =
    $convert.base64Decode('Cg5Mb2dvdXRSZXNwb25zZQ==');

@$core.Deprecated('Use listUsersRequestDescriptor instead')
const ListUsersRequest$json = {
  '1': 'ListUsersRequest',
};

/// Descriptor for `ListUsersRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listUsersRequestDescriptor =
    $convert.base64Decode('ChBMaXN0VXNlcnNSZXF1ZXN0');

@$core.Deprecated('Use listUsersResponseDescriptor instead')
const ListUsersResponse$json = {
  '1': 'ListUsersResponse',
  '2': [
    {
      '1': 'users',
      '3': 1,
      '4': 3,
      '5': 11,
      '6': '.agents.v1.User',
      '10': 'users'
    },
  ],
};

/// Descriptor for `ListUsersResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listUsersResponseDescriptor = $convert.base64Decode(
    'ChFMaXN0VXNlcnNSZXNwb25zZRIlCgV1c2VycxgBIAMoCzIPLmFnZW50cy52MS5Vc2VyUgV1c2'
    'Vycw==');

@$core.Deprecated('Use createUserRequestDescriptor instead')
const CreateUserRequest$json = {
  '1': 'CreateUserRequest',
  '2': [
    {'1': 'username', '3': 1, '4': 1, '5': 9, '8': {}, '10': 'username'},
    {'1': 'password', '3': 2, '4': 1, '5': 9, '8': {}, '10': 'password'},
    {'1': 'display_name', '3': 3, '4': 1, '5': 9, '10': 'displayName'},
    {'1': 'role', '3': 4, '4': 1, '5': 9, '10': 'role'},
    {'1': 'disabled', '3': 5, '4': 1, '5': 8, '10': 'disabled'},
  ],
};

/// Descriptor for `CreateUserRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List createUserRequestDescriptor = $convert.base64Decode(
    'ChFDcmVhdGVVc2VyUmVxdWVzdBIjCgh1c2VybmFtZRgBIAEoCUIH+kIEcgIQAVIIdXNlcm5hbW'
    'USIwoIcGFzc3dvcmQYAiABKAlCB/pCBHICEAFSCHBhc3N3b3JkEiEKDGRpc3BsYXlfbmFtZRgD'
    'IAEoCVILZGlzcGxheU5hbWUSEgoEcm9sZRgEIAEoCVIEcm9sZRIaCghkaXNhYmxlZBgFIAEoCF'
    'IIZGlzYWJsZWQ=');

@$core.Deprecated('Use createUserResponseDescriptor instead')
const CreateUserResponse$json = {
  '1': 'CreateUserResponse',
  '2': [
    {
      '1': 'user',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.User',
      '10': 'user'
    },
  ],
};

/// Descriptor for `CreateUserResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List createUserResponseDescriptor = $convert.base64Decode(
    'ChJDcmVhdGVVc2VyUmVzcG9uc2USIwoEdXNlchgBIAEoCzIPLmFnZW50cy52MS5Vc2VyUgR1c2'
    'Vy');

@$core.Deprecated('Use updateUserPasswordRequestDescriptor instead')
const UpdateUserPasswordRequest$json = {
  '1': 'UpdateUserPasswordRequest',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '8': {}, '10': 'id'},
    {'1': 'password', '3': 2, '4': 1, '5': 9, '8': {}, '10': 'password'},
  ],
};

/// Descriptor for `UpdateUserPasswordRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List updateUserPasswordRequestDescriptor =
    $convert.base64Decode(
        'ChlVcGRhdGVVc2VyUGFzc3dvcmRSZXF1ZXN0EhcKAmlkGAEgASgJQgf6QgRyAhABUgJpZBIjCg'
        'hwYXNzd29yZBgCIAEoCUIH+kIEcgIQAVIIcGFzc3dvcmQ=');

@$core.Deprecated('Use updateUserPasswordResponseDescriptor instead')
const UpdateUserPasswordResponse$json = {
  '1': 'UpdateUserPasswordResponse',
  '2': [
    {
      '1': 'user',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.User',
      '10': 'user'
    },
  ],
};

/// Descriptor for `UpdateUserPasswordResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List updateUserPasswordResponseDescriptor =
    $convert.base64Decode(
        'ChpVcGRhdGVVc2VyUGFzc3dvcmRSZXNwb25zZRIjCgR1c2VyGAEgASgLMg8uYWdlbnRzLnYxLl'
        'VzZXJSBHVzZXI=');

@$core.Deprecated('Use setUserDisabledRequestDescriptor instead')
const SetUserDisabledRequest$json = {
  '1': 'SetUserDisabledRequest',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '8': {}, '10': 'id'},
    {'1': 'disabled', '3': 2, '4': 1, '5': 8, '10': 'disabled'},
  ],
};

/// Descriptor for `SetUserDisabledRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List setUserDisabledRequestDescriptor =
    $convert.base64Decode(
        'ChZTZXRVc2VyRGlzYWJsZWRSZXF1ZXN0EhcKAmlkGAEgASgJQgf6QgRyAhABUgJpZBIaCghkaX'
        'NhYmxlZBgCIAEoCFIIZGlzYWJsZWQ=');

@$core.Deprecated('Use setUserDisabledResponseDescriptor instead')
const SetUserDisabledResponse$json = {
  '1': 'SetUserDisabledResponse',
  '2': [
    {
      '1': 'user',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.User',
      '10': 'user'
    },
  ],
};

/// Descriptor for `SetUserDisabledResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List setUserDisabledResponseDescriptor =
    $convert.base64Decode(
        'ChdTZXRVc2VyRGlzYWJsZWRSZXNwb25zZRIjCgR1c2VyGAEgASgLMg8uYWdlbnRzLnYxLlVzZX'
        'JSBHVzZXI=');

@$core.Deprecated('Use updateProfileRequestDescriptor instead')
const UpdateProfileRequest$json = {
  '1': 'UpdateProfileRequest',
  '2': [
    {'1': 'display_name', '3': 1, '4': 1, '5': 9, '8': {}, '10': 'displayName'},
    {
      '1': 'avatar_url',
      '3': 2,
      '4': 1,
      '5': 9,
      '9': 0,
      '10': 'avatarUrl',
      '17': true
    },
  ],
  '8': [
    {'1': '_avatar_url'},
  ],
};

/// Descriptor for `UpdateProfileRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List updateProfileRequestDescriptor = $convert.base64Decode(
    'ChRVcGRhdGVQcm9maWxlUmVxdWVzdBIqCgxkaXNwbGF5X25hbWUYASABKAlCB/pCBHICEAFSC2'
    'Rpc3BsYXlOYW1lEiIKCmF2YXRhcl91cmwYAiABKAlIAFIJYXZhdGFyVXJsiAEBQg0KC19hdmF0'
    'YXJfdXJs');

@$core.Deprecated('Use updateProfileResponseDescriptor instead')
const UpdateProfileResponse$json = {
  '1': 'UpdateProfileResponse',
  '2': [
    {
      '1': 'user',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.User',
      '10': 'user'
    },
  ],
};

/// Descriptor for `UpdateProfileResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List updateProfileResponseDescriptor = $convert.base64Decode(
    'ChVVcGRhdGVQcm9maWxlUmVzcG9uc2USIwoEdXNlchgBIAEoCzIPLmFnZW50cy52MS5Vc2VyUg'
    'R1c2Vy');

@$core.Deprecated('Use changePasswordRequestDescriptor instead')
const ChangePasswordRequest$json = {
  '1': 'ChangePasswordRequest',
  '2': [
    {
      '1': 'current_password',
      '3': 1,
      '4': 1,
      '5': 9,
      '8': {},
      '10': 'currentPassword'
    },
    {'1': 'new_password', '3': 2, '4': 1, '5': 9, '8': {}, '10': 'newPassword'},
  ],
};

/// Descriptor for `ChangePasswordRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List changePasswordRequestDescriptor = $convert.base64Decode(
    'ChVDaGFuZ2VQYXNzd29yZFJlcXVlc3QSMgoQY3VycmVudF9wYXNzd29yZBgBIAEoCUIH+kIEcg'
    'IQAVIPY3VycmVudFBhc3N3b3JkEioKDG5ld19wYXNzd29yZBgCIAEoCUIH+kIEcgIQAVILbmV3'
    'UGFzc3dvcmQ=');

@$core.Deprecated('Use changePasswordResponseDescriptor instead')
const ChangePasswordResponse$json = {
  '1': 'ChangePasswordResponse',
  '2': [
    {
      '1': 'user',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.User',
      '10': 'user'
    },
  ],
};

/// Descriptor for `ChangePasswordResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List changePasswordResponseDescriptor =
    $convert.base64Decode(
        'ChZDaGFuZ2VQYXNzd29yZFJlc3BvbnNlEiMKBHVzZXIYASABKAsyDy5hZ2VudHMudjEuVXNlcl'
        'IEdXNlcg==');

@$core.Deprecated('Use oAuthProviderDescriptor instead')
const OAuthProvider$json = {
  '1': 'OAuthProvider',
  '2': [
    {'1': 'name', '3': 1, '4': 1, '5': 9, '10': 'name'},
    {'1': 'display_name', '3': 2, '4': 1, '5': 9, '10': 'displayName'},
  ],
};

/// Descriptor for `OAuthProvider`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List oAuthProviderDescriptor = $convert.base64Decode(
    'Cg1PQXV0aFByb3ZpZGVyEhIKBG5hbWUYASABKAlSBG5hbWUSIQoMZGlzcGxheV9uYW1lGAIgAS'
    'gJUgtkaXNwbGF5TmFtZQ==');

@$core.Deprecated('Use listOAuthProvidersRequestDescriptor instead')
const ListOAuthProvidersRequest$json = {
  '1': 'ListOAuthProvidersRequest',
};

/// Descriptor for `ListOAuthProvidersRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listOAuthProvidersRequestDescriptor =
    $convert.base64Decode('ChlMaXN0T0F1dGhQcm92aWRlcnNSZXF1ZXN0');

@$core.Deprecated('Use listOAuthProvidersResponseDescriptor instead')
const ListOAuthProvidersResponse$json = {
  '1': 'ListOAuthProvidersResponse',
  '2': [
    {
      '1': 'providers',
      '3': 1,
      '4': 3,
      '5': 11,
      '6': '.agents.v1.OAuthProvider',
      '10': 'providers'
    },
  ],
};

/// Descriptor for `ListOAuthProvidersResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listOAuthProvidersResponseDescriptor =
    $convert.base64Decode(
        'ChpMaXN0T0F1dGhQcm92aWRlcnNSZXNwb25zZRI2Cglwcm92aWRlcnMYASADKAsyGC5hZ2VudH'
        'MudjEuT0F1dGhQcm92aWRlclIJcHJvdmlkZXJz');

@$core.Deprecated('Use beginOAuthFlowRequestDescriptor instead')
const BeginOAuthFlowRequest$json = {
  '1': 'BeginOAuthFlowRequest',
  '2': [
    {'1': 'provider', '3': 1, '4': 1, '5': 9, '8': {}, '10': 'provider'},
    {'1': 'redirect_uri', '3': 2, '4': 1, '5': 9, '10': 'redirectUri'},
    {
      '1': 'native',
      '3': 3,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.NativeOAuthClient',
      '10': 'native'
    },
  ],
};

/// Descriptor for `BeginOAuthFlowRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List beginOAuthFlowRequestDescriptor = $convert.base64Decode(
    'ChVCZWdpbk9BdXRoRmxvd1JlcXVlc3QSIwoIcHJvdmlkZXIYASABKAlCB/pCBHICEAFSCHByb3'
    'ZpZGVyEiEKDHJlZGlyZWN0X3VyaRgCIAEoCVILcmVkaXJlY3RVcmkSNAoGbmF0aXZlGAMgASgL'
    'MhwuYWdlbnRzLnYxLk5hdGl2ZU9BdXRoQ2xpZW50UgZuYXRpdmU=');

@$core.Deprecated('Use nativeOAuthClientDescriptor instead')
const NativeOAuthClient$json = {
  '1': 'NativeOAuthClient',
  '2': [
    {'1': 'return_uri', '3': 1, '4': 1, '5': 9, '8': {}, '10': 'returnUri'},
    {
      '1': 'code_challenge',
      '3': 2,
      '4': 1,
      '5': 9,
      '8': {},
      '10': 'codeChallenge'
    },
  ],
};

/// Descriptor for `NativeOAuthClient`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List nativeOAuthClientDescriptor = $convert.base64Decode(
    'ChFOYXRpdmVPQXV0aENsaWVudBImCgpyZXR1cm5fdXJpGAEgASgJQgf6QgRyAhABUglyZXR1cm'
    '5VcmkSLwoOY29kZV9jaGFsbGVuZ2UYAiABKAlCCPpCBXIDmAErUg1jb2RlQ2hhbGxlbmdl');

@$core.Deprecated('Use beginOAuthFlowResponseDescriptor instead')
const BeginOAuthFlowResponse$json = {
  '1': 'BeginOAuthFlowResponse',
  '2': [
    {'1': 'authorize_url', '3': 1, '4': 1, '5': 9, '10': 'authorizeUrl'},
    {'1': 'state', '3': 2, '4': 1, '5': 9, '10': 'state'},
  ],
};

/// Descriptor for `BeginOAuthFlowResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List beginOAuthFlowResponseDescriptor =
    $convert.base64Decode(
        'ChZCZWdpbk9BdXRoRmxvd1Jlc3BvbnNlEiMKDWF1dGhvcml6ZV91cmwYASABKAlSDGF1dGhvcm'
        'l6ZVVybBIUCgVzdGF0ZRgCIAEoCVIFc3RhdGU=');

@$core.Deprecated('Use completeOAuthFlowRequestDescriptor instead')
const CompleteOAuthFlowRequest$json = {
  '1': 'CompleteOAuthFlowRequest',
  '2': [
    {'1': 'provider', '3': 1, '4': 1, '5': 9, '8': {}, '10': 'provider'},
    {'1': 'code', '3': 2, '4': 1, '5': 9, '8': {}, '10': 'code'},
    {'1': 'state', '3': 3, '4': 1, '5': 9, '8': {}, '10': 'state'},
  ],
};

/// Descriptor for `CompleteOAuthFlowRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List completeOAuthFlowRequestDescriptor = $convert.base64Decode(
    'ChhDb21wbGV0ZU9BdXRoRmxvd1JlcXVlc3QSIwoIcHJvdmlkZXIYASABKAlCB/pCBHICEAFSCH'
    'Byb3ZpZGVyEhsKBGNvZGUYAiABKAlCB/pCBHICEAFSBGNvZGUSHQoFc3RhdGUYAyABKAlCB/pC'
    'BHICEAFSBXN0YXRl');

@$core.Deprecated('Use completeOAuthFlowResponseDescriptor instead')
const CompleteOAuthFlowResponse$json = {
  '1': 'CompleteOAuthFlowResponse',
  '2': [
    {'1': 'token', '3': 1, '4': 1, '5': 9, '10': 'token'},
    {
      '1': 'user',
      '3': 2,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.User',
      '10': 'user'
    },
    {
      '1': 'expires_at',
      '3': 3,
      '4': 1,
      '5': 11,
      '6': '.google.protobuf.Timestamp',
      '10': 'expiresAt'
    },
    {
      '1': 'workspaces',
      '3': 4,
      '4': 3,
      '5': 11,
      '6': '.agents.v1.Workspace',
      '10': 'workspaces'
    },
    {
      '1': 'native_handoff',
      '3': 5,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.NativeOAuthHandoff',
      '10': 'nativeHandoff'
    },
  ],
};

/// Descriptor for `CompleteOAuthFlowResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List completeOAuthFlowResponseDescriptor = $convert.base64Decode(
    'ChlDb21wbGV0ZU9BdXRoRmxvd1Jlc3BvbnNlEhQKBXRva2VuGAEgASgJUgV0b2tlbhIjCgR1c2'
    'VyGAIgASgLMg8uYWdlbnRzLnYxLlVzZXJSBHVzZXISOQoKZXhwaXJlc19hdBgDIAEoCzIaLmdv'
    'b2dsZS5wcm90b2J1Zi5UaW1lc3RhbXBSCWV4cGlyZXNBdBI0Cgp3b3Jrc3BhY2VzGAQgAygLMh'
    'QuYWdlbnRzLnYxLldvcmtzcGFjZVIKd29ya3NwYWNlcxJECg5uYXRpdmVfaGFuZG9mZhgFIAEo'
    'CzIdLmFnZW50cy52MS5OYXRpdmVPQXV0aEhhbmRvZmZSDW5hdGl2ZUhhbmRvZmY=');

@$core.Deprecated('Use nativeOAuthHandoffDescriptor instead')
const NativeOAuthHandoff$json = {
  '1': 'NativeOAuthHandoff',
  '2': [
    {'1': 'return_url', '3': 1, '4': 1, '5': 9, '10': 'returnUrl'},
  ],
};

/// Descriptor for `NativeOAuthHandoff`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List nativeOAuthHandoffDescriptor =
    $convert.base64Decode(
        'ChJOYXRpdmVPQXV0aEhhbmRvZmYSHQoKcmV0dXJuX3VybBgBIAEoCVIJcmV0dXJuVXJs');

@$core.Deprecated('Use redeemOAuthHandoffRequestDescriptor instead')
const RedeemOAuthHandoffRequest$json = {
  '1': 'RedeemOAuthHandoffRequest',
  '2': [
    {'1': 'provider', '3': 1, '4': 1, '5': 9, '8': {}, '10': 'provider'},
    {'1': 'state', '3': 2, '4': 1, '5': 9, '8': {}, '10': 'state'},
    {'1': 'handoff_code', '3': 3, '4': 1, '5': 9, '8': {}, '10': 'handoffCode'},
    {
      '1': 'code_verifier',
      '3': 4,
      '4': 1,
      '5': 9,
      '8': {},
      '10': 'codeVerifier'
    },
  ],
};

/// Descriptor for `RedeemOAuthHandoffRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List redeemOAuthHandoffRequestDescriptor = $convert.base64Decode(
    'ChlSZWRlZW1PQXV0aEhhbmRvZmZSZXF1ZXN0EiMKCHByb3ZpZGVyGAEgASgJQgf6QgRyAhABUg'
    'hwcm92aWRlchIdCgVzdGF0ZRgCIAEoCUIH+kIEcgIQAVIFc3RhdGUSKgoMaGFuZG9mZl9jb2Rl'
    'GAMgASgJQgf6QgRyAhABUgtoYW5kb2ZmQ29kZRIvCg1jb2RlX3ZlcmlmaWVyGAQgASgJQgr6Qg'
    'dyBRArGIABUgxjb2RlVmVyaWZpZXI=');

@$core.Deprecated('Use redeemOAuthHandoffResponseDescriptor instead')
const RedeemOAuthHandoffResponse$json = {
  '1': 'RedeemOAuthHandoffResponse',
  '2': [
    {'1': 'token', '3': 1, '4': 1, '5': 9, '10': 'token'},
    {
      '1': 'user',
      '3': 2,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.User',
      '10': 'user'
    },
    {
      '1': 'expires_at',
      '3': 3,
      '4': 1,
      '5': 11,
      '6': '.google.protobuf.Timestamp',
      '10': 'expiresAt'
    },
    {
      '1': 'workspaces',
      '3': 4,
      '4': 3,
      '5': 11,
      '6': '.agents.v1.Workspace',
      '10': 'workspaces'
    },
  ],
};

/// Descriptor for `RedeemOAuthHandoffResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List redeemOAuthHandoffResponseDescriptor = $convert.base64Decode(
    'ChpSZWRlZW1PQXV0aEhhbmRvZmZSZXNwb25zZRIUCgV0b2tlbhgBIAEoCVIFdG9rZW4SIwoEdX'
    'NlchgCIAEoCzIPLmFnZW50cy52MS5Vc2VyUgR1c2VyEjkKCmV4cGlyZXNfYXQYAyABKAsyGi5n'
    'b29nbGUucHJvdG9idWYuVGltZXN0YW1wUglleHBpcmVzQXQSNAoKd29ya3NwYWNlcxgEIAMoCz'
    'IULmFnZW50cy52MS5Xb3Jrc3BhY2VSCndvcmtzcGFjZXM=');

const $core.Map<$core.String, $core.dynamic> AuthServiceBase$json = {
  '1': 'AuthService',
  '2': [
    {
      '1': 'Login',
      '2': '.agents.v1.LoginRequest',
      '3': '.agents.v1.LoginResponse'
    },
    {'1': 'Me', '2': '.agents.v1.MeRequest', '3': '.agents.v1.MeResponse'},
    {
      '1': 'Logout',
      '2': '.agents.v1.LogoutRequest',
      '3': '.agents.v1.LogoutResponse'
    },
    {
      '1': 'ListUsers',
      '2': '.agents.v1.ListUsersRequest',
      '3': '.agents.v1.ListUsersResponse'
    },
    {
      '1': 'CreateUser',
      '2': '.agents.v1.CreateUserRequest',
      '3': '.agents.v1.CreateUserResponse'
    },
    {
      '1': 'UpdateUserPassword',
      '2': '.agents.v1.UpdateUserPasswordRequest',
      '3': '.agents.v1.UpdateUserPasswordResponse'
    },
    {
      '1': 'SetUserDisabled',
      '2': '.agents.v1.SetUserDisabledRequest',
      '3': '.agents.v1.SetUserDisabledResponse'
    },
    {
      '1': 'UpdateProfile',
      '2': '.agents.v1.UpdateProfileRequest',
      '3': '.agents.v1.UpdateProfileResponse'
    },
    {
      '1': 'ChangePassword',
      '2': '.agents.v1.ChangePasswordRequest',
      '3': '.agents.v1.ChangePasswordResponse'
    },
    {
      '1': 'ListOAuthProviders',
      '2': '.agents.v1.ListOAuthProvidersRequest',
      '3': '.agents.v1.ListOAuthProvidersResponse'
    },
    {
      '1': 'BeginOAuthFlow',
      '2': '.agents.v1.BeginOAuthFlowRequest',
      '3': '.agents.v1.BeginOAuthFlowResponse'
    },
    {
      '1': 'CompleteOAuthFlow',
      '2': '.agents.v1.CompleteOAuthFlowRequest',
      '3': '.agents.v1.CompleteOAuthFlowResponse'
    },
    {
      '1': 'RedeemOAuthHandoff',
      '2': '.agents.v1.RedeemOAuthHandoffRequest',
      '3': '.agents.v1.RedeemOAuthHandoffResponse'
    },
  ],
};

@$core.Deprecated('Use authServiceDescriptor instead')
const $core.Map<$core.String, $core.Map<$core.String, $core.dynamic>>
    AuthServiceBase$messageJson = {
  '.agents.v1.LoginRequest': LoginRequest$json,
  '.agents.v1.LoginResponse': LoginResponse$json,
  '.agents.v1.User': User$json,
  '.google.protobuf.Timestamp': $0.Timestamp$json,
  '.agents.v1.Workspace': $1.Workspace$json,
  '.agents.v1.MeRequest': MeRequest$json,
  '.agents.v1.MeResponse': MeResponse$json,
  '.agents.v1.LogoutRequest': LogoutRequest$json,
  '.agents.v1.LogoutResponse': LogoutResponse$json,
  '.agents.v1.ListUsersRequest': ListUsersRequest$json,
  '.agents.v1.ListUsersResponse': ListUsersResponse$json,
  '.agents.v1.CreateUserRequest': CreateUserRequest$json,
  '.agents.v1.CreateUserResponse': CreateUserResponse$json,
  '.agents.v1.UpdateUserPasswordRequest': UpdateUserPasswordRequest$json,
  '.agents.v1.UpdateUserPasswordResponse': UpdateUserPasswordResponse$json,
  '.agents.v1.SetUserDisabledRequest': SetUserDisabledRequest$json,
  '.agents.v1.SetUserDisabledResponse': SetUserDisabledResponse$json,
  '.agents.v1.UpdateProfileRequest': UpdateProfileRequest$json,
  '.agents.v1.UpdateProfileResponse': UpdateProfileResponse$json,
  '.agents.v1.ChangePasswordRequest': ChangePasswordRequest$json,
  '.agents.v1.ChangePasswordResponse': ChangePasswordResponse$json,
  '.agents.v1.ListOAuthProvidersRequest': ListOAuthProvidersRequest$json,
  '.agents.v1.ListOAuthProvidersResponse': ListOAuthProvidersResponse$json,
  '.agents.v1.OAuthProvider': OAuthProvider$json,
  '.agents.v1.BeginOAuthFlowRequest': BeginOAuthFlowRequest$json,
  '.agents.v1.NativeOAuthClient': NativeOAuthClient$json,
  '.agents.v1.BeginOAuthFlowResponse': BeginOAuthFlowResponse$json,
  '.agents.v1.CompleteOAuthFlowRequest': CompleteOAuthFlowRequest$json,
  '.agents.v1.CompleteOAuthFlowResponse': CompleteOAuthFlowResponse$json,
  '.agents.v1.NativeOAuthHandoff': NativeOAuthHandoff$json,
  '.agents.v1.RedeemOAuthHandoffRequest': RedeemOAuthHandoffRequest$json,
  '.agents.v1.RedeemOAuthHandoffResponse': RedeemOAuthHandoffResponse$json,
};

/// Descriptor for `AuthService`. Decode as a `google.protobuf.ServiceDescriptorProto`.
final $typed_data.Uint8List authServiceDescriptor = $convert.base64Decode(
    'CgtBdXRoU2VydmljZRI6CgVMb2dpbhIXLmFnZW50cy52MS5Mb2dpblJlcXVlc3QaGC5hZ2VudH'
    'MudjEuTG9naW5SZXNwb25zZRIxCgJNZRIULmFnZW50cy52MS5NZVJlcXVlc3QaFS5hZ2VudHMu'
    'djEuTWVSZXNwb25zZRI9CgZMb2dvdXQSGC5hZ2VudHMudjEuTG9nb3V0UmVxdWVzdBoZLmFnZW'
    '50cy52MS5Mb2dvdXRSZXNwb25zZRJGCglMaXN0VXNlcnMSGy5hZ2VudHMudjEuTGlzdFVzZXJz'
    'UmVxdWVzdBocLmFnZW50cy52MS5MaXN0VXNlcnNSZXNwb25zZRJJCgpDcmVhdGVVc2VyEhwuYW'
    'dlbnRzLnYxLkNyZWF0ZVVzZXJSZXF1ZXN0Gh0uYWdlbnRzLnYxLkNyZWF0ZVVzZXJSZXNwb25z'
    'ZRJhChJVcGRhdGVVc2VyUGFzc3dvcmQSJC5hZ2VudHMudjEuVXBkYXRlVXNlclBhc3N3b3JkUm'
    'VxdWVzdBolLmFnZW50cy52MS5VcGRhdGVVc2VyUGFzc3dvcmRSZXNwb25zZRJYCg9TZXRVc2Vy'
    'RGlzYWJsZWQSIS5hZ2VudHMudjEuU2V0VXNlckRpc2FibGVkUmVxdWVzdBoiLmFnZW50cy52MS'
    '5TZXRVc2VyRGlzYWJsZWRSZXNwb25zZRJSCg1VcGRhdGVQcm9maWxlEh8uYWdlbnRzLnYxLlVw'
    'ZGF0ZVByb2ZpbGVSZXF1ZXN0GiAuYWdlbnRzLnYxLlVwZGF0ZVByb2ZpbGVSZXNwb25zZRJVCg'
    '5DaGFuZ2VQYXNzd29yZBIgLmFnZW50cy52MS5DaGFuZ2VQYXNzd29yZFJlcXVlc3QaIS5hZ2Vu'
    'dHMudjEuQ2hhbmdlUGFzc3dvcmRSZXNwb25zZRJhChJMaXN0T0F1dGhQcm92aWRlcnMSJC5hZ2'
    'VudHMudjEuTGlzdE9BdXRoUHJvdmlkZXJzUmVxdWVzdBolLmFnZW50cy52MS5MaXN0T0F1dGhQ'
    'cm92aWRlcnNSZXNwb25zZRJVCg5CZWdpbk9BdXRoRmxvdxIgLmFnZW50cy52MS5CZWdpbk9BdX'
    'RoRmxvd1JlcXVlc3QaIS5hZ2VudHMudjEuQmVnaW5PQXV0aEZsb3dSZXNwb25zZRJeChFDb21w'
    'bGV0ZU9BdXRoRmxvdxIjLmFnZW50cy52MS5Db21wbGV0ZU9BdXRoRmxvd1JlcXVlc3QaJC5hZ2'
    'VudHMudjEuQ29tcGxldGVPQXV0aEZsb3dSZXNwb25zZRJhChJSZWRlZW1PQXV0aEhhbmRvZmYS'
    'JC5hZ2VudHMudjEuUmVkZWVtT0F1dGhIYW5kb2ZmUmVxdWVzdBolLmFnZW50cy52MS5SZWRlZW'
    '1PQXV0aEhhbmRvZmZSZXNwb25zZQ==');
