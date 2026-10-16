// This is a generated file - do not edit.
//
// Generated from agents/v1/linear.proto.

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

@$core.Deprecated('Use linearAppCredentialStateDescriptor instead')
const LinearAppCredentialState$json = {
  '1': 'LinearAppCredentialState',
  '2': [
    {'1': 'LINEAR_APP_CREDENTIAL_STATE_UNSPECIFIED', '2': 0},
    {'1': 'LINEAR_APP_CREDENTIAL_STATE_INCOMPLETE', '2': 1},
    {'1': 'LINEAR_APP_CREDENTIAL_STATE_COMPLETE', '2': 2},
  ],
};

/// Descriptor for `LinearAppCredentialState`. Decode as a `google.protobuf.EnumDescriptorProto`.
final $typed_data.Uint8List linearAppCredentialStateDescriptor = $convert.base64Decode(
    'ChhMaW5lYXJBcHBDcmVkZW50aWFsU3RhdGUSKwonTElORUFSX0FQUF9DUkVERU5USUFMX1NUQV'
    'RFX1VOU1BFQ0lGSUVEEAASKgomTElORUFSX0FQUF9DUkVERU5USUFMX1NUQVRFX0lOQ09NUExF'
    'VEUQARIoCiRMSU5FQVJfQVBQX0NSRURFTlRJQUxfU1RBVEVfQ09NUExFVEUQAg==');

@$core.Deprecated('Use linearInstallationCredentialStateDescriptor instead')
const LinearInstallationCredentialState$json = {
  '1': 'LinearInstallationCredentialState',
  '2': [
    {'1': 'LINEAR_INSTALLATION_CREDENTIAL_STATE_UNSPECIFIED', '2': 0},
    {'1': 'LINEAR_INSTALLATION_CREDENTIAL_STATE_VALID', '2': 1},
    {'1': 'LINEAR_INSTALLATION_CREDENTIAL_STATE_NEEDS_REINSTALL', '2': 2},
  ],
};

/// Descriptor for `LinearInstallationCredentialState`. Decode as a `google.protobuf.EnumDescriptorProto`.
final $typed_data.Uint8List linearInstallationCredentialStateDescriptor =
    $convert.base64Decode(
        'CiFMaW5lYXJJbnN0YWxsYXRpb25DcmVkZW50aWFsU3RhdGUSNAowTElORUFSX0lOU1RBTExBVE'
        'lPTl9DUkVERU5USUFMX1NUQVRFX1VOU1BFQ0lGSUVEEAASLgoqTElORUFSX0lOU1RBTExBVElP'
        'Tl9DUkVERU5USUFMX1NUQVRFX1ZBTElEEAESOAo0TElORUFSX0lOU1RBTExBVElPTl9DUkVERU'
        '5USUFMX1NUQVRFX05FRURTX1JFSU5TVEFMTBAC');

@$core.Deprecated('Use linearProcessingStatusDescriptor instead')
const LinearProcessingStatus$json = {
  '1': 'LinearProcessingStatus',
  '2': [
    {'1': 'LINEAR_PROCESSING_STATUS_UNSPECIFIED', '2': 0},
    {'1': 'LINEAR_PROCESSING_STATUS_QUEUED', '2': 1},
    {'1': 'LINEAR_PROCESSING_STATUS_RECEIVED', '2': 2},
    {'1': 'LINEAR_PROCESSING_STATUS_PROCESSING', '2': 3},
    {'1': 'LINEAR_PROCESSING_STATUS_READY_TO_DELIVER', '2': 4},
    {'1': 'LINEAR_PROCESSING_STATUS_SUCCEEDED', '2': 5},
    {'1': 'LINEAR_PROCESSING_STATUS_FAILED', '2': 6},
    {'1': 'LINEAR_PROCESSING_STATUS_FAILED_UNCERTAIN', '2': 7},
    {'1': 'LINEAR_PROCESSING_STATUS_CANCELLED', '2': 8},
  ],
};

/// Descriptor for `LinearProcessingStatus`. Decode as a `google.protobuf.EnumDescriptorProto`.
final $typed_data.Uint8List linearProcessingStatusDescriptor = $convert.base64Decode(
    'ChZMaW5lYXJQcm9jZXNzaW5nU3RhdHVzEigKJExJTkVBUl9QUk9DRVNTSU5HX1NUQVRVU19VTl'
    'NQRUNJRklFRBAAEiMKH0xJTkVBUl9QUk9DRVNTSU5HX1NUQVRVU19RVUVVRUQQARIlCiFMSU5F'
    'QVJfUFJPQ0VTU0lOR19TVEFUVVNfUkVDRUlWRUQQAhInCiNMSU5FQVJfUFJPQ0VTU0lOR19TVE'
    'FUVVNfUFJPQ0VTU0lORxADEi0KKUxJTkVBUl9QUk9DRVNTSU5HX1NUQVRVU19SRUFEWV9UT19E'
    'RUxJVkVSEAQSJgoiTElORUFSX1BST0NFU1NJTkdfU1RBVFVTX1NVQ0NFRURFRBAFEiMKH0xJTk'
    'VBUl9QUk9DRVNTSU5HX1NUQVRVU19GQUlMRUQQBhItCilMSU5FQVJfUFJPQ0VTU0lOR19TVEFU'
    'VVNfRkFJTEVEX1VOQ0VSVEFJThAHEiYKIkxJTkVBUl9QUk9DRVNTSU5HX1NUQVRVU19DQU5DRU'
    'xMRUQQCA==');

@$core.Deprecated('Use linearAppDescriptor instead')
const LinearApp$json = {
  '1': 'LinearApp',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
    {'1': 'display_name', '3': 2, '4': 1, '5': 9, '10': 'displayName'},
    {'1': 'client_id', '3': 3, '4': 1, '5': 9, '10': 'clientId'},
    {'1': 'agent_id', '3': 4, '4': 1, '5': 9, '10': 'agentId'},
    {'1': 'inbound_enabled', '3': 5, '4': 1, '5': 8, '10': 'inboundEnabled'},
    {'1': 'allowed_user_ids', '3': 6, '4': 3, '5': 9, '10': 'allowedUserIds'},
    {
      '1': 'max_run_seconds',
      '3': 7,
      '4': 1,
      '5': 5,
      '9': 0,
      '10': 'maxRunSeconds',
      '17': true
    },
    {'1': 'revision', '3': 8, '4': 1, '5': 3, '10': 'revision'},
    {
      '1': 'credential_state',
      '3': 9,
      '4': 1,
      '5': 14,
      '6': '.agents.v1.LinearAppCredentialState',
      '10': 'credentialState'
    },
    {
      '1': 'client_secret_set',
      '3': 10,
      '4': 1,
      '5': 8,
      '10': 'clientSecretSet'
    },
    {
      '1': 'webhook_secret_set',
      '3': 11,
      '4': 1,
      '5': 8,
      '10': 'webhookSecretSet'
    },
    {'1': 'callback_url', '3': 12, '4': 1, '5': 9, '10': 'callbackUrl'},
    {'1': 'webhook_url', '3': 13, '4': 1, '5': 9, '10': 'webhookUrl'},
    {
      '1': 'created_at',
      '3': 14,
      '4': 1,
      '5': 11,
      '6': '.google.protobuf.Timestamp',
      '10': 'createdAt'
    },
    {
      '1': 'updated_at',
      '3': 15,
      '4': 1,
      '5': 11,
      '6': '.google.protobuf.Timestamp',
      '10': 'updatedAt'
    },
    {'1': 'workspace_id', '3': 100, '4': 1, '5': 9, '10': 'workspaceId'},
  ],
  '8': [
    {'1': '_max_run_seconds'},
  ],
};

/// Descriptor for `LinearApp`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List linearAppDescriptor = $convert.base64Decode(
    'CglMaW5lYXJBcHASDgoCaWQYASABKAlSAmlkEiEKDGRpc3BsYXlfbmFtZRgCIAEoCVILZGlzcG'
    'xheU5hbWUSGwoJY2xpZW50X2lkGAMgASgJUghjbGllbnRJZBIZCghhZ2VudF9pZBgEIAEoCVIH'
    'YWdlbnRJZBInCg9pbmJvdW5kX2VuYWJsZWQYBSABKAhSDmluYm91bmRFbmFibGVkEigKEGFsbG'
    '93ZWRfdXNlcl9pZHMYBiADKAlSDmFsbG93ZWRVc2VySWRzEisKD21heF9ydW5fc2Vjb25kcxgH'
    'IAEoBUgAUg1tYXhSdW5TZWNvbmRziAEBEhoKCHJldmlzaW9uGAggASgDUghyZXZpc2lvbhJOCh'
    'BjcmVkZW50aWFsX3N0YXRlGAkgASgOMiMuYWdlbnRzLnYxLkxpbmVhckFwcENyZWRlbnRpYWxT'
    'dGF0ZVIPY3JlZGVudGlhbFN0YXRlEioKEWNsaWVudF9zZWNyZXRfc2V0GAogASgIUg9jbGllbn'
    'RTZWNyZXRTZXQSLAoSd2ViaG9va19zZWNyZXRfc2V0GAsgASgIUhB3ZWJob29rU2VjcmV0U2V0'
    'EiEKDGNhbGxiYWNrX3VybBgMIAEoCVILY2FsbGJhY2tVcmwSHwoLd2ViaG9va191cmwYDSABKA'
    'lSCndlYmhvb2tVcmwSOQoKY3JlYXRlZF9hdBgOIAEoCzIaLmdvb2dsZS5wcm90b2J1Zi5UaW1l'
    'c3RhbXBSCWNyZWF0ZWRBdBI5Cgp1cGRhdGVkX2F0GA8gASgLMhouZ29vZ2xlLnByb3RvYnVmLl'
    'RpbWVzdGFtcFIJdXBkYXRlZEF0EiEKDHdvcmtzcGFjZV9pZBhkIAEoCVILd29ya3NwYWNlSWRC'
    'EgoQX21heF9ydW5fc2Vjb25kcw==');

@$core.Deprecated('Use listLinearAppsRequestDescriptor instead')
const ListLinearAppsRequest$json = {
  '1': 'ListLinearAppsRequest',
};

/// Descriptor for `ListLinearAppsRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listLinearAppsRequestDescriptor =
    $convert.base64Decode('ChVMaXN0TGluZWFyQXBwc1JlcXVlc3Q=');

@$core.Deprecated('Use listLinearAppsResponseDescriptor instead')
const ListLinearAppsResponse$json = {
  '1': 'ListLinearAppsResponse',
  '2': [
    {
      '1': 'apps',
      '3': 1,
      '4': 3,
      '5': 11,
      '6': '.agents.v1.LinearApp',
      '10': 'apps'
    },
  ],
};

/// Descriptor for `ListLinearAppsResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listLinearAppsResponseDescriptor =
    $convert.base64Decode(
        'ChZMaXN0TGluZWFyQXBwc1Jlc3BvbnNlEigKBGFwcHMYASADKAsyFC5hZ2VudHMudjEuTGluZW'
        'FyQXBwUgRhcHBz');

@$core.Deprecated('Use getLinearAppRequestDescriptor instead')
const GetLinearAppRequest$json = {
  '1': 'GetLinearAppRequest',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '8': {}, '10': 'id'},
  ],
};

/// Descriptor for `GetLinearAppRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getLinearAppRequestDescriptor =
    $convert.base64Decode(
        'ChNHZXRMaW5lYXJBcHBSZXF1ZXN0EhcKAmlkGAEgASgJQgf6QgRyAhABUgJpZA==');

@$core.Deprecated('Use getLinearAppResponseDescriptor instead')
const GetLinearAppResponse$json = {
  '1': 'GetLinearAppResponse',
  '2': [
    {
      '1': 'app',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.LinearApp',
      '10': 'app'
    },
  ],
};

/// Descriptor for `GetLinearAppResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getLinearAppResponseDescriptor = $convert.base64Decode(
    'ChRHZXRMaW5lYXJBcHBSZXNwb25zZRImCgNhcHAYASABKAsyFC5hZ2VudHMudjEuTGluZWFyQX'
    'BwUgNhcHA=');

@$core.Deprecated('Use createLinearAppRequestDescriptor instead')
const CreateLinearAppRequest$json = {
  '1': 'CreateLinearAppRequest',
  '2': [
    {
      '1': 'app',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.LinearApp',
      '10': 'app'
    },
    {
      '1': 'client_secret',
      '3': 2,
      '4': 1,
      '5': 9,
      '9': 0,
      '10': 'clientSecret',
      '17': true
    },
    {
      '1': 'webhook_secret',
      '3': 3,
      '4': 1,
      '5': 9,
      '9': 1,
      '10': 'webhookSecret',
      '17': true
    },
  ],
  '8': [
    {'1': '_client_secret'},
    {'1': '_webhook_secret'},
  ],
};

/// Descriptor for `CreateLinearAppRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List createLinearAppRequestDescriptor = $convert.base64Decode(
    'ChZDcmVhdGVMaW5lYXJBcHBSZXF1ZXN0EiYKA2FwcBgBIAEoCzIULmFnZW50cy52MS5MaW5lYX'
    'JBcHBSA2FwcBIoCg1jbGllbnRfc2VjcmV0GAIgASgJSABSDGNsaWVudFNlY3JldIgBARIqCg53'
    'ZWJob29rX3NlY3JldBgDIAEoCUgBUg13ZWJob29rU2VjcmV0iAEBQhAKDl9jbGllbnRfc2Vjcm'
    'V0QhEKD193ZWJob29rX3NlY3JldA==');

@$core.Deprecated('Use createLinearAppResponseDescriptor instead')
const CreateLinearAppResponse$json = {
  '1': 'CreateLinearAppResponse',
  '2': [
    {
      '1': 'app',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.LinearApp',
      '10': 'app'
    },
  ],
};

/// Descriptor for `CreateLinearAppResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List createLinearAppResponseDescriptor =
    $convert.base64Decode(
        'ChdDcmVhdGVMaW5lYXJBcHBSZXNwb25zZRImCgNhcHAYASABKAsyFC5hZ2VudHMudjEuTGluZW'
        'FyQXBwUgNhcHA=');

@$core.Deprecated('Use updateLinearAppRequestDescriptor instead')
const UpdateLinearAppRequest$json = {
  '1': 'UpdateLinearAppRequest',
  '2': [
    {
      '1': 'app',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.LinearApp',
      '10': 'app'
    },
  ],
};

/// Descriptor for `UpdateLinearAppRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List updateLinearAppRequestDescriptor =
    $convert.base64Decode(
        'ChZVcGRhdGVMaW5lYXJBcHBSZXF1ZXN0EiYKA2FwcBgBIAEoCzIULmFnZW50cy52MS5MaW5lYX'
        'JBcHBSA2FwcA==');

@$core.Deprecated('Use updateLinearAppResponseDescriptor instead')
const UpdateLinearAppResponse$json = {
  '1': 'UpdateLinearAppResponse',
  '2': [
    {
      '1': 'app',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.LinearApp',
      '10': 'app'
    },
  ],
};

/// Descriptor for `UpdateLinearAppResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List updateLinearAppResponseDescriptor =
    $convert.base64Decode(
        'ChdVcGRhdGVMaW5lYXJBcHBSZXNwb25zZRImCgNhcHAYASABKAsyFC5hZ2VudHMudjEuTGluZW'
        'FyQXBwUgNhcHA=');

@$core.Deprecated('Use putLinearAppCredentialsRequestDescriptor instead')
const PutLinearAppCredentialsRequest$json = {
  '1': 'PutLinearAppCredentialsRequest',
  '2': [
    {'1': 'app_id', '3': 1, '4': 1, '5': 9, '8': {}, '10': 'appId'},
    {
      '1': 'client_secret',
      '3': 2,
      '4': 1,
      '5': 9,
      '9': 0,
      '10': 'clientSecret',
      '17': true
    },
    {
      '1': 'webhook_secret',
      '3': 3,
      '4': 1,
      '5': 9,
      '9': 1,
      '10': 'webhookSecret',
      '17': true
    },
  ],
  '8': [
    {'1': '_client_secret'},
    {'1': '_webhook_secret'},
  ],
};

/// Descriptor for `PutLinearAppCredentialsRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List putLinearAppCredentialsRequestDescriptor =
    $convert.base64Decode(
        'Ch5QdXRMaW5lYXJBcHBDcmVkZW50aWFsc1JlcXVlc3QSHgoGYXBwX2lkGAEgASgJQgf6QgRyAh'
        'ABUgVhcHBJZBIoCg1jbGllbnRfc2VjcmV0GAIgASgJSABSDGNsaWVudFNlY3JldIgBARIqCg53'
        'ZWJob29rX3NlY3JldBgDIAEoCUgBUg13ZWJob29rU2VjcmV0iAEBQhAKDl9jbGllbnRfc2Vjcm'
        'V0QhEKD193ZWJob29rX3NlY3JldA==');

@$core.Deprecated('Use putLinearAppCredentialsResponseDescriptor instead')
const PutLinearAppCredentialsResponse$json = {
  '1': 'PutLinearAppCredentialsResponse',
  '2': [
    {
      '1': 'app',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.LinearApp',
      '10': 'app'
    },
  ],
};

/// Descriptor for `PutLinearAppCredentialsResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List putLinearAppCredentialsResponseDescriptor =
    $convert.base64Decode(
        'Ch9QdXRMaW5lYXJBcHBDcmVkZW50aWFsc1Jlc3BvbnNlEiYKA2FwcBgBIAEoCzIULmFnZW50cy'
        '52MS5MaW5lYXJBcHBSA2FwcA==');

@$core.Deprecated('Use deleteLinearAppRequestDescriptor instead')
const DeleteLinearAppRequest$json = {
  '1': 'DeleteLinearAppRequest',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '8': {}, '10': 'id'},
  ],
};

/// Descriptor for `DeleteLinearAppRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List deleteLinearAppRequestDescriptor =
    $convert.base64Decode(
        'ChZEZWxldGVMaW5lYXJBcHBSZXF1ZXN0EhcKAmlkGAEgASgJQgf6QgRyAhABUgJpZA==');

@$core.Deprecated('Use deleteLinearAppResponseDescriptor instead')
const DeleteLinearAppResponse$json = {
  '1': 'DeleteLinearAppResponse',
};

/// Descriptor for `DeleteLinearAppResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List deleteLinearAppResponseDescriptor =
    $convert.base64Decode('ChdEZWxldGVMaW5lYXJBcHBSZXNwb25zZQ==');

@$core.Deprecated('Use linearInstallationDescriptor instead')
const LinearInstallation$json = {
  '1': 'LinearInstallation',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
    {'1': 'app_id', '3': 2, '4': 1, '5': 9, '10': 'appId'},
    {'1': 'organization_id', '3': 3, '4': 1, '5': 9, '10': 'organizationId'},
    {
      '1': 'organization_name',
      '3': 4,
      '4': 1,
      '5': 9,
      '10': 'organizationName'
    },
    {'1': 'app_user_id', '3': 5, '4': 1, '5': 9, '10': 'appUserId'},
    {'1': 'scopes', '3': 6, '4': 3, '5': 9, '10': 'scopes'},
    {
      '1': 'credential_state',
      '3': 7,
      '4': 1,
      '5': 14,
      '6': '.agents.v1.LinearInstallationCredentialState',
      '10': 'credentialState'
    },
    {
      '1': 'last_credential_error',
      '3': 8,
      '4': 1,
      '5': 9,
      '10': 'lastCredentialError'
    },
    {
      '1': 'installed_at',
      '3': 9,
      '4': 1,
      '5': 11,
      '6': '.google.protobuf.Timestamp',
      '10': 'installedAt'
    },
    {
      '1': 'updated_at',
      '3': 10,
      '4': 1,
      '5': 11,
      '6': '.google.protobuf.Timestamp',
      '10': 'updatedAt'
    },
    {'1': 'workspace_id', '3': 100, '4': 1, '5': 9, '10': 'workspaceId'},
  ],
};

/// Descriptor for `LinearInstallation`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List linearInstallationDescriptor = $convert.base64Decode(
    'ChJMaW5lYXJJbnN0YWxsYXRpb24SDgoCaWQYASABKAlSAmlkEhUKBmFwcF9pZBgCIAEoCVIFYX'
    'BwSWQSJwoPb3JnYW5pemF0aW9uX2lkGAMgASgJUg5vcmdhbml6YXRpb25JZBIrChFvcmdhbml6'
    'YXRpb25fbmFtZRgEIAEoCVIQb3JnYW5pemF0aW9uTmFtZRIeCgthcHBfdXNlcl9pZBgFIAEoCV'
    'IJYXBwVXNlcklkEhYKBnNjb3BlcxgGIAMoCVIGc2NvcGVzElcKEGNyZWRlbnRpYWxfc3RhdGUY'
    'ByABKA4yLC5hZ2VudHMudjEuTGluZWFySW5zdGFsbGF0aW9uQ3JlZGVudGlhbFN0YXRlUg9jcm'
    'VkZW50aWFsU3RhdGUSMgoVbGFzdF9jcmVkZW50aWFsX2Vycm9yGAggASgJUhNsYXN0Q3JlZGVu'
    'dGlhbEVycm9yEj0KDGluc3RhbGxlZF9hdBgJIAEoCzIaLmdvb2dsZS5wcm90b2J1Zi5UaW1lc3'
    'RhbXBSC2luc3RhbGxlZEF0EjkKCnVwZGF0ZWRfYXQYCiABKAsyGi5nb29nbGUucHJvdG9idWYu'
    'VGltZXN0YW1wUgl1cGRhdGVkQXQSIQoMd29ya3NwYWNlX2lkGGQgASgJUgt3b3Jrc3BhY2VJZA'
    '==');

@$core.Deprecated('Use beginLinearInstallRequestDescriptor instead')
const BeginLinearInstallRequest$json = {
  '1': 'BeginLinearInstallRequest',
  '2': [
    {'1': 'app_id', '3': 1, '4': 1, '5': 9, '8': {}, '10': 'appId'},
    {'1': 'return_url', '3': 2, '4': 1, '5': 9, '10': 'returnUrl'},
  ],
};

/// Descriptor for `BeginLinearInstallRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List beginLinearInstallRequestDescriptor =
    $convert.base64Decode(
        'ChlCZWdpbkxpbmVhckluc3RhbGxSZXF1ZXN0Eh4KBmFwcF9pZBgBIAEoCUIH+kIEcgIQAVIFYX'
        'BwSWQSHQoKcmV0dXJuX3VybBgCIAEoCVIJcmV0dXJuVXJs');

@$core.Deprecated('Use beginLinearInstallResponseDescriptor instead')
const BeginLinearInstallResponse$json = {
  '1': 'BeginLinearInstallResponse',
  '2': [
    {'1': 'authorize_url', '3': 1, '4': 1, '5': 9, '10': 'authorizeUrl'},
  ],
};

/// Descriptor for `BeginLinearInstallResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List beginLinearInstallResponseDescriptor =
    $convert.base64Decode(
        'ChpCZWdpbkxpbmVhckluc3RhbGxSZXNwb25zZRIjCg1hdXRob3JpemVfdXJsGAEgASgJUgxhdX'
        'Rob3JpemVVcmw=');

@$core.Deprecated('Use listLinearInstallationsRequestDescriptor instead')
const ListLinearInstallationsRequest$json = {
  '1': 'ListLinearInstallationsRequest',
  '2': [
    {'1': 'app_id', '3': 1, '4': 1, '5': 9, '8': {}, '10': 'appId'},
  ],
};

/// Descriptor for `ListLinearInstallationsRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listLinearInstallationsRequestDescriptor =
    $convert.base64Decode(
        'Ch5MaXN0TGluZWFySW5zdGFsbGF0aW9uc1JlcXVlc3QSHgoGYXBwX2lkGAEgASgJQgf6QgRyAh'
        'ABUgVhcHBJZA==');

@$core.Deprecated('Use listLinearInstallationsResponseDescriptor instead')
const ListLinearInstallationsResponse$json = {
  '1': 'ListLinearInstallationsResponse',
  '2': [
    {
      '1': 'installations',
      '3': 1,
      '4': 3,
      '5': 11,
      '6': '.agents.v1.LinearInstallation',
      '10': 'installations'
    },
  ],
};

/// Descriptor for `ListLinearInstallationsResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listLinearInstallationsResponseDescriptor =
    $convert.base64Decode(
        'Ch9MaXN0TGluZWFySW5zdGFsbGF0aW9uc1Jlc3BvbnNlEkMKDWluc3RhbGxhdGlvbnMYASADKA'
        'syHS5hZ2VudHMudjEuTGluZWFySW5zdGFsbGF0aW9uUg1pbnN0YWxsYXRpb25z');

@$core.Deprecated('Use deleteLinearInstallationRequestDescriptor instead')
const DeleteLinearInstallationRequest$json = {
  '1': 'DeleteLinearInstallationRequest',
  '2': [
    {'1': 'app_id', '3': 1, '4': 1, '5': 9, '8': {}, '10': 'appId'},
    {'1': 'id', '3': 2, '4': 1, '5': 9, '8': {}, '10': 'id'},
  ],
};

/// Descriptor for `DeleteLinearInstallationRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List deleteLinearInstallationRequestDescriptor =
    $convert.base64Decode(
        'Ch9EZWxldGVMaW5lYXJJbnN0YWxsYXRpb25SZXF1ZXN0Eh4KBmFwcF9pZBgBIAEoCUIH+kIEcg'
        'IQAVIFYXBwSWQSFwoCaWQYAiABKAlCB/pCBHICEAFSAmlk');

@$core.Deprecated('Use deleteLinearInstallationResponseDescriptor instead')
const DeleteLinearInstallationResponse$json = {
  '1': 'DeleteLinearInstallationResponse',
};

/// Descriptor for `DeleteLinearInstallationResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List deleteLinearInstallationResponseDescriptor =
    $convert.base64Decode('CiBEZWxldGVMaW5lYXJJbnN0YWxsYXRpb25SZXNwb25zZQ==');

@$core.Deprecated('Use linearSettingsDescriptor instead')
const LinearSettings$json = {
  '1': 'LinearSettings',
  '2': [
    {'1': 'public_base_url', '3': 1, '4': 1, '5': 9, '10': 'publicBaseUrl'},
    {
      '1': 'updated_at',
      '3': 2,
      '4': 1,
      '5': 11,
      '6': '.google.protobuf.Timestamp',
      '10': 'updatedAt'
    },
  ],
};

/// Descriptor for `LinearSettings`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List linearSettingsDescriptor = $convert.base64Decode(
    'Cg5MaW5lYXJTZXR0aW5ncxImCg9wdWJsaWNfYmFzZV91cmwYASABKAlSDXB1YmxpY0Jhc2VVcm'
    'wSOQoKdXBkYXRlZF9hdBgCIAEoCzIaLmdvb2dsZS5wcm90b2J1Zi5UaW1lc3RhbXBSCXVwZGF0'
    'ZWRBdA==');

@$core.Deprecated('Use getLinearSettingsRequestDescriptor instead')
const GetLinearSettingsRequest$json = {
  '1': 'GetLinearSettingsRequest',
};

/// Descriptor for `GetLinearSettingsRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getLinearSettingsRequestDescriptor =
    $convert.base64Decode('ChhHZXRMaW5lYXJTZXR0aW5nc1JlcXVlc3Q=');

@$core.Deprecated('Use getLinearSettingsResponseDescriptor instead')
const GetLinearSettingsResponse$json = {
  '1': 'GetLinearSettingsResponse',
  '2': [
    {
      '1': 'settings',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.LinearSettings',
      '10': 'settings'
    },
  ],
};

/// Descriptor for `GetLinearSettingsResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getLinearSettingsResponseDescriptor =
    $convert.base64Decode(
        'ChlHZXRMaW5lYXJTZXR0aW5nc1Jlc3BvbnNlEjUKCHNldHRpbmdzGAEgASgLMhkuYWdlbnRzLn'
        'YxLkxpbmVhclNldHRpbmdzUghzZXR0aW5ncw==');

@$core.Deprecated('Use updateLinearSettingsRequestDescriptor instead')
const UpdateLinearSettingsRequest$json = {
  '1': 'UpdateLinearSettingsRequest',
  '2': [
    {
      '1': 'settings',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.LinearSettings',
      '10': 'settings'
    },
  ],
};

/// Descriptor for `UpdateLinearSettingsRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List updateLinearSettingsRequestDescriptor =
    $convert.base64Decode(
        'ChtVcGRhdGVMaW5lYXJTZXR0aW5nc1JlcXVlc3QSNQoIc2V0dGluZ3MYASABKAsyGS5hZ2VudH'
        'MudjEuTGluZWFyU2V0dGluZ3NSCHNldHRpbmdz');

@$core.Deprecated('Use updateLinearSettingsResponseDescriptor instead')
const UpdateLinearSettingsResponse$json = {
  '1': 'UpdateLinearSettingsResponse',
  '2': [
    {
      '1': 'settings',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.LinearSettings',
      '10': 'settings'
    },
  ],
};

/// Descriptor for `UpdateLinearSettingsResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List updateLinearSettingsResponseDescriptor =
    $convert.base64Decode(
        'ChxVcGRhdGVMaW5lYXJTZXR0aW5nc1Jlc3BvbnNlEjUKCHNldHRpbmdzGAEgASgLMhkuYWdlbn'
        'RzLnYxLkxpbmVhclNldHRpbmdzUghzZXR0aW5ncw==');

@$core.Deprecated('Use linearProcessingRecordDescriptor instead')
const LinearProcessingRecord$json = {
  '1': 'LinearProcessingRecord',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
    {'1': 'workspace_id', '3': 2, '4': 1, '5': 9, '10': 'workspaceId'},
    {'1': 'app_id', '3': 3, '4': 1, '5': 9, '10': 'appId'},
    {'1': 'installation_id', '3': 4, '4': 1, '5': 9, '10': 'installationId'},
    {'1': 'agent_session_id', '3': 5, '4': 1, '5': 9, '10': 'agentSessionId'},
    {'1': 'action', '3': 6, '4': 1, '5': 9, '10': 'action'},
    {'1': 'delivery_id', '3': 7, '4': 1, '5': 9, '10': 'deliveryId'},
    {
      '1': 'status',
      '3': 8,
      '4': 1,
      '5': 14,
      '6': '.agents.v1.LinearProcessingStatus',
      '10': 'status'
    },
    {'1': 'invocation_id', '3': 9, '4': 1, '5': 9, '10': 'invocationId'},
    {'1': 'attempts', '3': 10, '4': 1, '5': 5, '10': 'attempts'},
    {'1': 'app_revision', '3': 11, '4': 1, '5': 3, '10': 'appRevision'},
    {'1': 'error', '3': 12, '4': 1, '5': 9, '10': 'error'},
    {'1': 'output', '3': 13, '4': 1, '5': 9, '10': 'output'},
    {'1': 'output_type', '3': 14, '4': 1, '5': 9, '10': 'outputType'},
    {'1': 'output_signal', '3': 15, '4': 1, '5': 9, '10': 'outputSignal'},
    {
      '1': 'output_signal_metadata',
      '3': 16,
      '4': 1,
      '5': 9,
      '10': 'outputSignalMetadata'
    },
    {'1': 'delivered', '3': 17, '4': 1, '5': 8, '10': 'delivered'},
    {'1': 'dead_lettered', '3': 18, '4': 1, '5': 8, '10': 'deadLettered'},
    {
      '1': 'prompting_user_id',
      '3': 19,
      '4': 1,
      '5': 9,
      '10': 'promptingUserId'
    },
    {'1': 'issue_identifier', '3': 20, '4': 1, '5': 9, '10': 'issueIdentifier'},
    {'1': 'agent_id', '3': 24, '4': 1, '5': 9, '10': 'agentId'},
    {
      '1': 'created_at',
      '3': 21,
      '4': 1,
      '5': 11,
      '6': '.google.protobuf.Timestamp',
      '10': 'createdAt'
    },
    {
      '1': 'updated_at',
      '3': 22,
      '4': 1,
      '5': 11,
      '6': '.google.protobuf.Timestamp',
      '10': 'updatedAt'
    },
    {
      '1': 'expires_at',
      '3': 23,
      '4': 1,
      '5': 11,
      '6': '.google.protobuf.Timestamp',
      '10': 'expiresAt'
    },
  ],
};

/// Descriptor for `LinearProcessingRecord`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List linearProcessingRecordDescriptor = $convert.base64Decode(
    'ChZMaW5lYXJQcm9jZXNzaW5nUmVjb3JkEg4KAmlkGAEgASgJUgJpZBIhCgx3b3Jrc3BhY2VfaW'
    'QYAiABKAlSC3dvcmtzcGFjZUlkEhUKBmFwcF9pZBgDIAEoCVIFYXBwSWQSJwoPaW5zdGFsbGF0'
    'aW9uX2lkGAQgASgJUg5pbnN0YWxsYXRpb25JZBIoChBhZ2VudF9zZXNzaW9uX2lkGAUgASgJUg'
    '5hZ2VudFNlc3Npb25JZBIWCgZhY3Rpb24YBiABKAlSBmFjdGlvbhIfCgtkZWxpdmVyeV9pZBgH'
    'IAEoCVIKZGVsaXZlcnlJZBI5CgZzdGF0dXMYCCABKA4yIS5hZ2VudHMudjEuTGluZWFyUHJvY2'
    'Vzc2luZ1N0YXR1c1IGc3RhdHVzEiMKDWludm9jYXRpb25faWQYCSABKAlSDGludm9jYXRpb25J'
    'ZBIaCghhdHRlbXB0cxgKIAEoBVIIYXR0ZW1wdHMSIQoMYXBwX3JldmlzaW9uGAsgASgDUgthcH'
    'BSZXZpc2lvbhIUCgVlcnJvchgMIAEoCVIFZXJyb3ISFgoGb3V0cHV0GA0gASgJUgZvdXRwdXQS'
    'HwoLb3V0cHV0X3R5cGUYDiABKAlSCm91dHB1dFR5cGUSIwoNb3V0cHV0X3NpZ25hbBgPIAEoCV'
    'IMb3V0cHV0U2lnbmFsEjQKFm91dHB1dF9zaWduYWxfbWV0YWRhdGEYECABKAlSFG91dHB1dFNp'
    'Z25hbE1ldGFkYXRhEhwKCWRlbGl2ZXJlZBgRIAEoCFIJZGVsaXZlcmVkEiMKDWRlYWRfbGV0dG'
    'VyZWQYEiABKAhSDGRlYWRMZXR0ZXJlZBIqChFwcm9tcHRpbmdfdXNlcl9pZBgTIAEoCVIPcHJv'
    'bXB0aW5nVXNlcklkEikKEGlzc3VlX2lkZW50aWZpZXIYFCABKAlSD2lzc3VlSWRlbnRpZmllch'
    'IZCghhZ2VudF9pZBgYIAEoCVIHYWdlbnRJZBI5CgpjcmVhdGVkX2F0GBUgASgLMhouZ29vZ2xl'
    'LnByb3RvYnVmLlRpbWVzdGFtcFIJY3JlYXRlZEF0EjkKCnVwZGF0ZWRfYXQYFiABKAsyGi5nb2'
    '9nbGUucHJvdG9idWYuVGltZXN0YW1wUgl1cGRhdGVkQXQSOQoKZXhwaXJlc19hdBgXIAEoCzIa'
    'Lmdvb2dsZS5wcm90b2J1Zi5UaW1lc3RhbXBSCWV4cGlyZXNBdA==');

@$core.Deprecated('Use listLinearProcessingRecordsRequestDescriptor instead')
const ListLinearProcessingRecordsRequest$json = {
  '1': 'ListLinearProcessingRecordsRequest',
  '2': [
    {'1': 'app_id', '3': 1, '4': 1, '5': 9, '10': 'appId'},
    {
      '1': 'status',
      '3': 2,
      '4': 1,
      '5': 14,
      '6': '.agents.v1.LinearProcessingStatus',
      '10': 'status'
    },
    {'1': 'page_size', '3': 3, '4': 1, '5': 5, '10': 'pageSize'},
  ],
};

/// Descriptor for `ListLinearProcessingRecordsRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listLinearProcessingRecordsRequestDescriptor =
    $convert.base64Decode(
        'CiJMaXN0TGluZWFyUHJvY2Vzc2luZ1JlY29yZHNSZXF1ZXN0EhUKBmFwcF9pZBgBIAEoCVIFYX'
        'BwSWQSOQoGc3RhdHVzGAIgASgOMiEuYWdlbnRzLnYxLkxpbmVhclByb2Nlc3NpbmdTdGF0dXNS'
        'BnN0YXR1cxIbCglwYWdlX3NpemUYAyABKAVSCHBhZ2VTaXpl');

@$core.Deprecated('Use listLinearProcessingRecordsResponseDescriptor instead')
const ListLinearProcessingRecordsResponse$json = {
  '1': 'ListLinearProcessingRecordsResponse',
  '2': [
    {
      '1': 'records',
      '3': 1,
      '4': 3,
      '5': 11,
      '6': '.agents.v1.LinearProcessingRecord',
      '10': 'records'
    },
  ],
};

/// Descriptor for `ListLinearProcessingRecordsResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listLinearProcessingRecordsResponseDescriptor =
    $convert.base64Decode(
        'CiNMaXN0TGluZWFyUHJvY2Vzc2luZ1JlY29yZHNSZXNwb25zZRI7CgdyZWNvcmRzGAEgAygLMi'
        'EuYWdlbnRzLnYxLkxpbmVhclByb2Nlc3NpbmdSZWNvcmRSB3JlY29yZHM=');

@$core.Deprecated('Use getLinearProcessingRecordRequestDescriptor instead')
const GetLinearProcessingRecordRequest$json = {
  '1': 'GetLinearProcessingRecordRequest',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '8': {}, '10': 'id'},
  ],
};

/// Descriptor for `GetLinearProcessingRecordRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getLinearProcessingRecordRequestDescriptor =
    $convert.base64Decode(
        'CiBHZXRMaW5lYXJQcm9jZXNzaW5nUmVjb3JkUmVxdWVzdBIXCgJpZBgBIAEoCUIH+kIEcgIQAV'
        'ICaWQ=');

@$core.Deprecated('Use getLinearProcessingRecordResponseDescriptor instead')
const GetLinearProcessingRecordResponse$json = {
  '1': 'GetLinearProcessingRecordResponse',
  '2': [
    {
      '1': 'record',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.LinearProcessingRecord',
      '10': 'record'
    },
  ],
};

/// Descriptor for `GetLinearProcessingRecordResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getLinearProcessingRecordResponseDescriptor =
    $convert.base64Decode(
        'CiFHZXRMaW5lYXJQcm9jZXNzaW5nUmVjb3JkUmVzcG9uc2USOQoGcmVjb3JkGAEgASgLMiEuYW'
        'dlbnRzLnYxLkxpbmVhclByb2Nlc3NpbmdSZWNvcmRSBnJlY29yZA==');

@$core.Deprecated('Use resendLinearReplyRequestDescriptor instead')
const ResendLinearReplyRequest$json = {
  '1': 'ResendLinearReplyRequest',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '8': {}, '10': 'id'},
  ],
};

/// Descriptor for `ResendLinearReplyRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List resendLinearReplyRequestDescriptor =
    $convert.base64Decode(
        'ChhSZXNlbmRMaW5lYXJSZXBseVJlcXVlc3QSFwoCaWQYASABKAlCB/pCBHICEAFSAmlk');

@$core.Deprecated('Use resendLinearReplyResponseDescriptor instead')
const ResendLinearReplyResponse$json = {
  '1': 'ResendLinearReplyResponse',
  '2': [
    {
      '1': 'record',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.LinearProcessingRecord',
      '10': 'record'
    },
  ],
};

/// Descriptor for `ResendLinearReplyResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List resendLinearReplyResponseDescriptor =
    $convert.base64Decode(
        'ChlSZXNlbmRMaW5lYXJSZXBseVJlc3BvbnNlEjkKBnJlY29yZBgBIAEoCzIhLmFnZW50cy52MS'
        '5MaW5lYXJQcm9jZXNzaW5nUmVjb3JkUgZyZWNvcmQ=');

const $core.Map<$core.String, $core.dynamic> LinearAppServiceBase$json = {
  '1': 'LinearAppService',
  '2': [
    {
      '1': 'ListLinearApps',
      '2': '.agents.v1.ListLinearAppsRequest',
      '3': '.agents.v1.ListLinearAppsResponse'
    },
    {
      '1': 'GetLinearApp',
      '2': '.agents.v1.GetLinearAppRequest',
      '3': '.agents.v1.GetLinearAppResponse'
    },
    {
      '1': 'CreateLinearApp',
      '2': '.agents.v1.CreateLinearAppRequest',
      '3': '.agents.v1.CreateLinearAppResponse'
    },
    {
      '1': 'UpdateLinearApp',
      '2': '.agents.v1.UpdateLinearAppRequest',
      '3': '.agents.v1.UpdateLinearAppResponse'
    },
    {
      '1': 'PutLinearAppCredentials',
      '2': '.agents.v1.PutLinearAppCredentialsRequest',
      '3': '.agents.v1.PutLinearAppCredentialsResponse'
    },
    {
      '1': 'DeleteLinearApp',
      '2': '.agents.v1.DeleteLinearAppRequest',
      '3': '.agents.v1.DeleteLinearAppResponse'
    },
    {
      '1': 'BeginLinearInstall',
      '2': '.agents.v1.BeginLinearInstallRequest',
      '3': '.agents.v1.BeginLinearInstallResponse'
    },
    {
      '1': 'ListLinearInstallations',
      '2': '.agents.v1.ListLinearInstallationsRequest',
      '3': '.agents.v1.ListLinearInstallationsResponse'
    },
    {
      '1': 'DeleteLinearInstallation',
      '2': '.agents.v1.DeleteLinearInstallationRequest',
      '3': '.agents.v1.DeleteLinearInstallationResponse'
    },
  ],
};

@$core.Deprecated('Use linearAppServiceDescriptor instead')
const $core.Map<$core.String, $core.Map<$core.String, $core.dynamic>>
    LinearAppServiceBase$messageJson = {
  '.agents.v1.ListLinearAppsRequest': ListLinearAppsRequest$json,
  '.agents.v1.ListLinearAppsResponse': ListLinearAppsResponse$json,
  '.agents.v1.LinearApp': LinearApp$json,
  '.google.protobuf.Timestamp': $0.Timestamp$json,
  '.agents.v1.GetLinearAppRequest': GetLinearAppRequest$json,
  '.agents.v1.GetLinearAppResponse': GetLinearAppResponse$json,
  '.agents.v1.CreateLinearAppRequest': CreateLinearAppRequest$json,
  '.agents.v1.CreateLinearAppResponse': CreateLinearAppResponse$json,
  '.agents.v1.UpdateLinearAppRequest': UpdateLinearAppRequest$json,
  '.agents.v1.UpdateLinearAppResponse': UpdateLinearAppResponse$json,
  '.agents.v1.PutLinearAppCredentialsRequest':
      PutLinearAppCredentialsRequest$json,
  '.agents.v1.PutLinearAppCredentialsResponse':
      PutLinearAppCredentialsResponse$json,
  '.agents.v1.DeleteLinearAppRequest': DeleteLinearAppRequest$json,
  '.agents.v1.DeleteLinearAppResponse': DeleteLinearAppResponse$json,
  '.agents.v1.BeginLinearInstallRequest': BeginLinearInstallRequest$json,
  '.agents.v1.BeginLinearInstallResponse': BeginLinearInstallResponse$json,
  '.agents.v1.ListLinearInstallationsRequest':
      ListLinearInstallationsRequest$json,
  '.agents.v1.ListLinearInstallationsResponse':
      ListLinearInstallationsResponse$json,
  '.agents.v1.LinearInstallation': LinearInstallation$json,
  '.agents.v1.DeleteLinearInstallationRequest':
      DeleteLinearInstallationRequest$json,
  '.agents.v1.DeleteLinearInstallationResponse':
      DeleteLinearInstallationResponse$json,
};

/// Descriptor for `LinearAppService`. Decode as a `google.protobuf.ServiceDescriptorProto`.
final $typed_data.Uint8List linearAppServiceDescriptor = $convert.base64Decode(
    'ChBMaW5lYXJBcHBTZXJ2aWNlElUKDkxpc3RMaW5lYXJBcHBzEiAuYWdlbnRzLnYxLkxpc3RMaW'
    '5lYXJBcHBzUmVxdWVzdBohLmFnZW50cy52MS5MaXN0TGluZWFyQXBwc1Jlc3BvbnNlEk8KDEdl'
    'dExpbmVhckFwcBIeLmFnZW50cy52MS5HZXRMaW5lYXJBcHBSZXF1ZXN0Gh8uYWdlbnRzLnYxLk'
    'dldExpbmVhckFwcFJlc3BvbnNlElgKD0NyZWF0ZUxpbmVhckFwcBIhLmFnZW50cy52MS5DcmVh'
    'dGVMaW5lYXJBcHBSZXF1ZXN0GiIuYWdlbnRzLnYxLkNyZWF0ZUxpbmVhckFwcFJlc3BvbnNlEl'
    'gKD1VwZGF0ZUxpbmVhckFwcBIhLmFnZW50cy52MS5VcGRhdGVMaW5lYXJBcHBSZXF1ZXN0GiIu'
    'YWdlbnRzLnYxLlVwZGF0ZUxpbmVhckFwcFJlc3BvbnNlEnAKF1B1dExpbmVhckFwcENyZWRlbn'
    'RpYWxzEikuYWdlbnRzLnYxLlB1dExpbmVhckFwcENyZWRlbnRpYWxzUmVxdWVzdBoqLmFnZW50'
    'cy52MS5QdXRMaW5lYXJBcHBDcmVkZW50aWFsc1Jlc3BvbnNlElgKD0RlbGV0ZUxpbmVhckFwcB'
    'IhLmFnZW50cy52MS5EZWxldGVMaW5lYXJBcHBSZXF1ZXN0GiIuYWdlbnRzLnYxLkRlbGV0ZUxp'
    'bmVhckFwcFJlc3BvbnNlEmEKEkJlZ2luTGluZWFySW5zdGFsbBIkLmFnZW50cy52MS5CZWdpbk'
    'xpbmVhckluc3RhbGxSZXF1ZXN0GiUuYWdlbnRzLnYxLkJlZ2luTGluZWFySW5zdGFsbFJlc3Bv'
    'bnNlEnAKF0xpc3RMaW5lYXJJbnN0YWxsYXRpb25zEikuYWdlbnRzLnYxLkxpc3RMaW5lYXJJbn'
    'N0YWxsYXRpb25zUmVxdWVzdBoqLmFnZW50cy52MS5MaXN0TGluZWFySW5zdGFsbGF0aW9uc1Jl'
    'c3BvbnNlEnMKGERlbGV0ZUxpbmVhckluc3RhbGxhdGlvbhIqLmFnZW50cy52MS5EZWxldGVMaW'
    '5lYXJJbnN0YWxsYXRpb25SZXF1ZXN0GisuYWdlbnRzLnYxLkRlbGV0ZUxpbmVhckluc3RhbGxh'
    'dGlvblJlc3BvbnNl');

const $core.Map<$core.String, $core.dynamic> LinearAdminServiceBase$json = {
  '1': 'LinearAdminService',
  '2': [
    {
      '1': 'GetLinearSettings',
      '2': '.agents.v1.GetLinearSettingsRequest',
      '3': '.agents.v1.GetLinearSettingsResponse'
    },
    {
      '1': 'UpdateLinearSettings',
      '2': '.agents.v1.UpdateLinearSettingsRequest',
      '3': '.agents.v1.UpdateLinearSettingsResponse'
    },
  ],
};

@$core.Deprecated('Use linearAdminServiceDescriptor instead')
const $core.Map<$core.String, $core.Map<$core.String, $core.dynamic>>
    LinearAdminServiceBase$messageJson = {
  '.agents.v1.GetLinearSettingsRequest': GetLinearSettingsRequest$json,
  '.agents.v1.GetLinearSettingsResponse': GetLinearSettingsResponse$json,
  '.agents.v1.LinearSettings': LinearSettings$json,
  '.google.protobuf.Timestamp': $0.Timestamp$json,
  '.agents.v1.UpdateLinearSettingsRequest': UpdateLinearSettingsRequest$json,
  '.agents.v1.UpdateLinearSettingsResponse': UpdateLinearSettingsResponse$json,
};

/// Descriptor for `LinearAdminService`. Decode as a `google.protobuf.ServiceDescriptorProto`.
final $typed_data.Uint8List linearAdminServiceDescriptor = $convert.base64Decode(
    'ChJMaW5lYXJBZG1pblNlcnZpY2USXgoRR2V0TGluZWFyU2V0dGluZ3MSIy5hZ2VudHMudjEuR2'
    'V0TGluZWFyU2V0dGluZ3NSZXF1ZXN0GiQuYWdlbnRzLnYxLkdldExpbmVhclNldHRpbmdzUmVz'
    'cG9uc2USZwoUVXBkYXRlTGluZWFyU2V0dGluZ3MSJi5hZ2VudHMudjEuVXBkYXRlTGluZWFyU2'
    'V0dGluZ3NSZXF1ZXN0GicuYWdlbnRzLnYxLlVwZGF0ZUxpbmVhclNldHRpbmdzUmVzcG9uc2U=');

const $core.Map<$core.String, $core.dynamic> LinearProcessingServiceBase$json =
    {
  '1': 'LinearProcessingService',
  '2': [
    {
      '1': 'ListLinearProcessingRecords',
      '2': '.agents.v1.ListLinearProcessingRecordsRequest',
      '3': '.agents.v1.ListLinearProcessingRecordsResponse'
    },
    {
      '1': 'GetLinearProcessingRecord',
      '2': '.agents.v1.GetLinearProcessingRecordRequest',
      '3': '.agents.v1.GetLinearProcessingRecordResponse'
    },
    {
      '1': 'ResendLinearReply',
      '2': '.agents.v1.ResendLinearReplyRequest',
      '3': '.agents.v1.ResendLinearReplyResponse'
    },
  ],
};

@$core.Deprecated('Use linearProcessingServiceDescriptor instead')
const $core.Map<$core.String, $core.Map<$core.String, $core.dynamic>>
    LinearProcessingServiceBase$messageJson = {
  '.agents.v1.ListLinearProcessingRecordsRequest':
      ListLinearProcessingRecordsRequest$json,
  '.agents.v1.ListLinearProcessingRecordsResponse':
      ListLinearProcessingRecordsResponse$json,
  '.agents.v1.LinearProcessingRecord': LinearProcessingRecord$json,
  '.google.protobuf.Timestamp': $0.Timestamp$json,
  '.agents.v1.GetLinearProcessingRecordRequest':
      GetLinearProcessingRecordRequest$json,
  '.agents.v1.GetLinearProcessingRecordResponse':
      GetLinearProcessingRecordResponse$json,
  '.agents.v1.ResendLinearReplyRequest': ResendLinearReplyRequest$json,
  '.agents.v1.ResendLinearReplyResponse': ResendLinearReplyResponse$json,
};

/// Descriptor for `LinearProcessingService`. Decode as a `google.protobuf.ServiceDescriptorProto`.
final $typed_data.Uint8List linearProcessingServiceDescriptor = $convert.base64Decode(
    'ChdMaW5lYXJQcm9jZXNzaW5nU2VydmljZRJ8ChtMaXN0TGluZWFyUHJvY2Vzc2luZ1JlY29yZH'
    'MSLS5hZ2VudHMudjEuTGlzdExpbmVhclByb2Nlc3NpbmdSZWNvcmRzUmVxdWVzdBouLmFnZW50'
    'cy52MS5MaXN0TGluZWFyUHJvY2Vzc2luZ1JlY29yZHNSZXNwb25zZRJ2ChlHZXRMaW5lYXJQcm'
    '9jZXNzaW5nUmVjb3JkEisuYWdlbnRzLnYxLkdldExpbmVhclByb2Nlc3NpbmdSZWNvcmRSZXF1'
    'ZXN0GiwuYWdlbnRzLnYxLkdldExpbmVhclByb2Nlc3NpbmdSZWNvcmRSZXNwb25zZRJeChFSZX'
    'NlbmRMaW5lYXJSZXBseRIjLmFnZW50cy52MS5SZXNlbmRMaW5lYXJSZXBseVJlcXVlc3QaJC5h'
    'Z2VudHMudjEuUmVzZW5kTGluZWFyUmVwbHlSZXNwb25zZQ==');
