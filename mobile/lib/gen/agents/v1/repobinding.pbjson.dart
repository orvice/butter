// This is a generated file - do not edit.
//
// Generated from agents/v1/repobinding.proto.

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

@$core.Deprecated('Use repoBindingWriteModeDescriptor instead')
const RepoBindingWriteMode$json = {
  '1': 'RepoBindingWriteMode',
  '2': [
    {'1': 'REPO_BINDING_WRITE_MODE_UNSPECIFIED', '2': 0},
    {'1': 'REPO_BINDING_WRITE_MODE_DIRECT_COMMIT', '2': 1},
    {'1': 'REPO_BINDING_WRITE_MODE_CHANGE_REQUEST', '2': 2},
  ],
};

/// Descriptor for `RepoBindingWriteMode`. Decode as a `google.protobuf.EnumDescriptorProto`.
final $typed_data.Uint8List repoBindingWriteModeDescriptor = $convert.base64Decode(
    'ChRSZXBvQmluZGluZ1dyaXRlTW9kZRInCiNSRVBPX0JJTkRJTkdfV1JJVEVfTU9ERV9VTlNQRU'
    'NJRklFRBAAEikKJVJFUE9fQklORElOR19XUklURV9NT0RFX0RJUkVDVF9DT01NSVQQARIqCiZS'
    'RVBPX0JJTkRJTkdfV1JJVEVfTU9ERV9DSEFOR0VfUkVRVUVTVBAC');

@$core.Deprecated('Use repoBindingOnboardingModeDescriptor instead')
const RepoBindingOnboardingMode$json = {
  '1': 'RepoBindingOnboardingMode',
  '2': [
    {'1': 'REPO_BINDING_ONBOARDING_MODE_UNSPECIFIED', '2': 0},
    {'1': 'REPO_BINDING_ONBOARDING_MODE_EXPORT_CURRENT', '2': 1},
    {'1': 'REPO_BINDING_ONBOARDING_MODE_IMPORT_REPOSITORY', '2': 2},
  ],
};

/// Descriptor for `RepoBindingOnboardingMode`. Decode as a `google.protobuf.EnumDescriptorProto`.
final $typed_data.Uint8List repoBindingOnboardingModeDescriptor = $convert.base64Decode(
    'ChlSZXBvQmluZGluZ09uYm9hcmRpbmdNb2RlEiwKKFJFUE9fQklORElOR19PTkJPQVJESU5HX0'
    '1PREVfVU5TUEVDSUZJRUQQABIvCitSRVBPX0JJTkRJTkdfT05CT0FSRElOR19NT0RFX0VYUE9S'
    'VF9DVVJSRU5UEAESMgouUkVQT19CSU5ESU5HX09OQk9BUkRJTkdfTU9ERV9JTVBPUlRfUkVQT1'
    'NJVE9SWRAC');

@$core.Deprecated('Use repoBindingDetachRecoveryDescriptor instead')
const RepoBindingDetachRecovery$json = {
  '1': 'RepoBindingDetachRecovery',
  '2': [
    {'1': 'REPO_BINDING_DETACH_RECOVERY_UNSPECIFIED', '2': 0},
    {'1': 'REPO_BINDING_DETACH_RECOVERY_KEEP_DATABASE', '2': 1},
  ],
};

/// Descriptor for `RepoBindingDetachRecovery`. Decode as a `google.protobuf.EnumDescriptorProto`.
final $typed_data.Uint8List repoBindingDetachRecoveryDescriptor = $convert.base64Decode(
    'ChlSZXBvQmluZGluZ0RldGFjaFJlY292ZXJ5EiwKKFJFUE9fQklORElOR19ERVRBQ0hfUkVDT1'
    'ZFUllfVU5TUEVDSUZJRUQQABIuCipSRVBPX0JJTkRJTkdfREVUQUNIX1JFQ09WRVJZX0tFRVBf'
    'REFUQUJBU0UQAQ==');

@$core.Deprecated('Use repoBindingConnectionStateDescriptor instead')
const RepoBindingConnectionState$json = {
  '1': 'RepoBindingConnectionState',
  '2': [
    {'1': 'REPO_BINDING_CONNECTION_STATE_UNSPECIFIED', '2': 0},
    {'1': 'REPO_BINDING_CONNECTION_STATE_UNVALIDATED', '2': 1},
    {'1': 'REPO_BINDING_CONNECTION_STATE_OK', '2': 2},
    {'1': 'REPO_BINDING_CONNECTION_STATE_FAILED', '2': 3},
    {'1': 'REPO_BINDING_CONNECTION_STATE_DEGRADED', '2': 4},
    {'1': 'REPO_BINDING_CONNECTION_STATE_DIVERGED', '2': 5},
  ],
};

/// Descriptor for `RepoBindingConnectionState`. Decode as a `google.protobuf.EnumDescriptorProto`.
final $typed_data.Uint8List repoBindingConnectionStateDescriptor = $convert.base64Decode(
    'ChpSZXBvQmluZGluZ0Nvbm5lY3Rpb25TdGF0ZRItCilSRVBPX0JJTkRJTkdfQ09OTkVDVElPTl'
    '9TVEFURV9VTlNQRUNJRklFRBAAEi0KKVJFUE9fQklORElOR19DT05ORUNUSU9OX1NUQVRFX1VO'
    'VkFMSURBVEVEEAESJAogUkVQT19CSU5ESU5HX0NPTk5FQ1RJT05fU1RBVEVfT0sQAhIoCiRSRV'
    'BPX0JJTkRJTkdfQ09OTkVDVElPTl9TVEFURV9GQUlMRUQQAxIqCiZSRVBPX0JJTkRJTkdfQ09O'
    'TkVDVElPTl9TVEFURV9ERUdSQURFRBAEEioKJlJFUE9fQklORElOR19DT05ORUNUSU9OX1NUQV'
    'RFX0RJVkVSR0VEEAU=');

@$core.Deprecated('Use contentFileOperationDescriptor instead')
const ContentFileOperation$json = {
  '1': 'ContentFileOperation',
  '2': [
    {'1': 'CONTENT_FILE_OPERATION_UNSPECIFIED', '2': 0},
    {'1': 'CONTENT_FILE_OPERATION_PUT', '2': 1},
    {'1': 'CONTENT_FILE_OPERATION_DELETE', '2': 2},
  ],
};

/// Descriptor for `ContentFileOperation`. Decode as a `google.protobuf.EnumDescriptorProto`.
final $typed_data.Uint8List contentFileOperationDescriptor = $convert.base64Decode(
    'ChRDb250ZW50RmlsZU9wZXJhdGlvbhImCiJDT05URU5UX0ZJTEVfT1BFUkFUSU9OX1VOU1BFQ0'
    'lGSUVEEAASHgoaQ09OVEVOVF9GSUxFX09QRVJBVElPTl9QVVQQARIhCh1DT05URU5UX0ZJTEVf'
    'T1BFUkFUSU9OX0RFTEVURRAC');

@$core.Deprecated('Use repoCacheEntryKindDescriptor instead')
const RepoCacheEntryKind$json = {
  '1': 'RepoCacheEntryKind',
  '2': [
    {'1': 'REPO_CACHE_ENTRY_KIND_UNSPECIFIED', '2': 0},
    {'1': 'REPO_CACHE_ENTRY_KIND_FILE', '2': 1},
    {'1': 'REPO_CACHE_ENTRY_KIND_DIRECTORY', '2': 2},
    {'1': 'REPO_CACHE_ENTRY_KIND_SYMLINK', '2': 3},
    {'1': 'REPO_CACHE_ENTRY_KIND_SUBMODULE', '2': 4},
  ],
};

/// Descriptor for `RepoCacheEntryKind`. Decode as a `google.protobuf.EnumDescriptorProto`.
final $typed_data.Uint8List repoCacheEntryKindDescriptor = $convert.base64Decode(
    'ChJSZXBvQ2FjaGVFbnRyeUtpbmQSJQohUkVQT19DQUNIRV9FTlRSWV9LSU5EX1VOU1BFQ0lGSU'
    'VEEAASHgoaUkVQT19DQUNIRV9FTlRSWV9LSU5EX0ZJTEUQARIjCh9SRVBPX0NBQ0hFX0VOVFJZ'
    'X0tJTkRfRElSRUNUT1JZEAISIQodUkVQT19DQUNIRV9FTlRSWV9LSU5EX1NZTUxJTksQAxIjCh'
    '9SRVBPX0NBQ0hFX0VOVFJZX0tJTkRfU1VCTU9EVUxFEAQ=');

@$core.Deprecated('Use workspaceRepoBindingDescriptor instead')
const WorkspaceRepoBinding$json = {
  '1': 'WorkspaceRepoBinding',
  '2': [
    {'1': 'git_host_id', '3': 1, '4': 1, '5': 9, '8': {}, '10': 'gitHostId'},
    {'1': 'repository', '3': 2, '4': 1, '5': 9, '8': {}, '10': 'repository'},
    {'1': 'branch', '3': 3, '4': 1, '5': 9, '8': {}, '10': 'branch'},
    {'1': 'root_path', '3': 4, '4': 1, '5': 9, '10': 'rootPath'},
    {
      '1': 'write_mode',
      '3': 5,
      '4': 1,
      '5': 14,
      '6': '.agents.v1.RepoBindingWriteMode',
      '10': 'writeMode'
    },
    {
      '1': 'content_schema_version',
      '3': 6,
      '4': 1,
      '5': 5,
      '10': 'contentSchemaVersion'
    },
    {'1': 'credential_set', '3': 7, '4': 1, '5': 8, '10': 'credentialSet'},
    {
      '1': 'credential_updated_at',
      '3': 8,
      '4': 1,
      '5': 11,
      '6': '.google.protobuf.Timestamp',
      '10': 'credentialUpdatedAt'
    },
    {
      '1': 'status',
      '3': 9,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.RepoBindingStatus',
      '10': 'status'
    },
    {
      '1': 'created_at',
      '3': 10,
      '4': 1,
      '5': 11,
      '6': '.google.protobuf.Timestamp',
      '10': 'createdAt'
    },
    {
      '1': 'updated_at',
      '3': 11,
      '4': 1,
      '5': 11,
      '6': '.google.protobuf.Timestamp',
      '10': 'updatedAt'
    },
    {
      '1': 'observed_commit_sha',
      '3': 12,
      '4': 1,
      '5': 9,
      '10': 'observedCommitSha'
    },
    {
      '1': 'active_commit_sha',
      '3': 13,
      '4': 1,
      '5': 9,
      '10': 'activeCommitSha'
    },
    {
      '1': 'last_synced_at',
      '3': 14,
      '4': 1,
      '5': 11,
      '6': '.google.protobuf.Timestamp',
      '10': 'lastSyncedAt'
    },
    {'1': 'last_sync_error', '3': 15, '4': 1, '5': 9, '10': 'lastSyncError'},
    {
      '1': 'last_publication_error',
      '3': 16,
      '4': 1,
      '5': 9,
      '10': 'lastPublicationError'
    },
    {
      '1': 'publication_errors',
      '3': 17,
      '4': 3,
      '5': 9,
      '10': 'publicationErrors'
    },
    {
      '1': 'webhook_secret_set',
      '3': 18,
      '4': 1,
      '5': 8,
      '10': 'webhookSecretSet'
    },
    {'1': 'workspace_id', '3': 100, '4': 1, '5': 9, '10': 'workspaceId'},
  ],
};

/// Descriptor for `WorkspaceRepoBinding`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List workspaceRepoBindingDescriptor = $convert.base64Decode(
    'ChRXb3Jrc3BhY2VSZXBvQmluZGluZxInCgtnaXRfaG9zdF9pZBgBIAEoCUIH+kIEcgIQAVIJZ2'
    'l0SG9zdElkEicKCnJlcG9zaXRvcnkYAiABKAlCB/pCBHICEAFSCnJlcG9zaXRvcnkSHwoGYnJh'
    'bmNoGAMgASgJQgf6QgRyAhABUgZicmFuY2gSGwoJcm9vdF9wYXRoGAQgASgJUghyb290UGF0aB'
    'I+Cgp3cml0ZV9tb2RlGAUgASgOMh8uYWdlbnRzLnYxLlJlcG9CaW5kaW5nV3JpdGVNb2RlUgl3'
    'cml0ZU1vZGUSNAoWY29udGVudF9zY2hlbWFfdmVyc2lvbhgGIAEoBVIUY29udGVudFNjaGVtYV'
    'ZlcnNpb24SJQoOY3JlZGVudGlhbF9zZXQYByABKAhSDWNyZWRlbnRpYWxTZXQSTgoVY3JlZGVu'
    'dGlhbF91cGRhdGVkX2F0GAggASgLMhouZ29vZ2xlLnByb3RvYnVmLlRpbWVzdGFtcFITY3JlZG'
    'VudGlhbFVwZGF0ZWRBdBI0CgZzdGF0dXMYCSABKAsyHC5hZ2VudHMudjEuUmVwb0JpbmRpbmdT'
    'dGF0dXNSBnN0YXR1cxI5CgpjcmVhdGVkX2F0GAogASgLMhouZ29vZ2xlLnByb3RvYnVmLlRpbW'
    'VzdGFtcFIJY3JlYXRlZEF0EjkKCnVwZGF0ZWRfYXQYCyABKAsyGi5nb29nbGUucHJvdG9idWYu'
    'VGltZXN0YW1wUgl1cGRhdGVkQXQSLgoTb2JzZXJ2ZWRfY29tbWl0X3NoYRgMIAEoCVIRb2JzZX'
    'J2ZWRDb21taXRTaGESKgoRYWN0aXZlX2NvbW1pdF9zaGEYDSABKAlSD2FjdGl2ZUNvbW1pdFNo'
    'YRJACg5sYXN0X3N5bmNlZF9hdBgOIAEoCzIaLmdvb2dsZS5wcm90b2J1Zi5UaW1lc3RhbXBSDG'
    'xhc3RTeW5jZWRBdBImCg9sYXN0X3N5bmNfZXJyb3IYDyABKAlSDWxhc3RTeW5jRXJyb3ISNAoW'
    'bGFzdF9wdWJsaWNhdGlvbl9lcnJvchgQIAEoCVIUbGFzdFB1YmxpY2F0aW9uRXJyb3ISLQoScH'
    'VibGljYXRpb25fZXJyb3JzGBEgAygJUhFwdWJsaWNhdGlvbkVycm9ycxIsChJ3ZWJob29rX3Nl'
    'Y3JldF9zZXQYEiABKAhSEHdlYmhvb2tTZWNyZXRTZXQSIQoMd29ya3NwYWNlX2lkGGQgASgJUg'
    't3b3Jrc3BhY2VJZA==');

@$core.Deprecated('Use repoBindingStatusDescriptor instead')
const RepoBindingStatus$json = {
  '1': 'RepoBindingStatus',
  '2': [
    {
      '1': 'state',
      '3': 1,
      '4': 1,
      '5': 14,
      '6': '.agents.v1.RepoBindingConnectionState',
      '10': 'state'
    },
    {
      '1': 'last_validated_at',
      '3': 2,
      '4': 1,
      '5': 11,
      '6': '.google.protobuf.Timestamp',
      '10': 'lastValidatedAt'
    },
    {'1': 'error', '3': 3, '4': 1, '5': 9, '10': 'error'},
    {
      '1': 'checks',
      '3': 4,
      '4': 3,
      '5': 11,
      '6': '.agents.v1.RepoBindingCheck',
      '10': 'checks'
    },
  ],
};

/// Descriptor for `RepoBindingStatus`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List repoBindingStatusDescriptor = $convert.base64Decode(
    'ChFSZXBvQmluZGluZ1N0YXR1cxI7CgVzdGF0ZRgBIAEoDjIlLmFnZW50cy52MS5SZXBvQmluZG'
    'luZ0Nvbm5lY3Rpb25TdGF0ZVIFc3RhdGUSRgoRbGFzdF92YWxpZGF0ZWRfYXQYAiABKAsyGi5n'
    'b29nbGUucHJvdG9idWYuVGltZXN0YW1wUg9sYXN0VmFsaWRhdGVkQXQSFAoFZXJyb3IYAyABKA'
    'lSBWVycm9yEjMKBmNoZWNrcxgEIAMoCzIbLmFnZW50cy52MS5SZXBvQmluZGluZ0NoZWNrUgZj'
    'aGVja3M=');

@$core.Deprecated('Use repoBindingCheckDescriptor instead')
const RepoBindingCheck$json = {
  '1': 'RepoBindingCheck',
  '2': [
    {'1': 'name', '3': 1, '4': 1, '5': 9, '10': 'name'},
    {'1': 'ok', '3': 2, '4': 1, '5': 8, '10': 'ok'},
    {'1': 'required', '3': 3, '4': 1, '5': 8, '10': 'required'},
    {'1': 'detail', '3': 4, '4': 1, '5': 9, '10': 'detail'},
  ],
};

/// Descriptor for `RepoBindingCheck`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List repoBindingCheckDescriptor = $convert.base64Decode(
    'ChBSZXBvQmluZGluZ0NoZWNrEhIKBG5hbWUYASABKAlSBG5hbWUSDgoCb2sYAiABKAhSAm9rEh'
    'oKCHJlcXVpcmVkGAMgASgIUghyZXF1aXJlZBIWCgZkZXRhaWwYBCABKAlSBmRldGFpbA==');

@$core.Deprecated('Use repoBindingOverlapDescriptor instead')
const RepoBindingOverlap$json = {
  '1': 'RepoBindingOverlap',
  '2': [
    {'1': 'workspace_id', '3': 1, '4': 1, '5': 9, '10': 'workspaceId'},
    {'1': 'workspace_name', '3': 2, '4': 1, '5': 9, '10': 'workspaceName'},
  ],
};

/// Descriptor for `RepoBindingOverlap`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List repoBindingOverlapDescriptor = $convert.base64Decode(
    'ChJSZXBvQmluZGluZ092ZXJsYXASIQoMd29ya3NwYWNlX2lkGAEgASgJUgt3b3Jrc3BhY2VJZB'
    'IlCg53b3Jrc3BhY2VfbmFtZRgCIAEoCVINd29ya3NwYWNlTmFtZQ==');

@$core.Deprecated('Use getWorkspaceRepoBindingRequestDescriptor instead')
const GetWorkspaceRepoBindingRequest$json = {
  '1': 'GetWorkspaceRepoBindingRequest',
};

/// Descriptor for `GetWorkspaceRepoBindingRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getWorkspaceRepoBindingRequestDescriptor =
    $convert.base64Decode('Ch5HZXRXb3Jrc3BhY2VSZXBvQmluZGluZ1JlcXVlc3Q=');

@$core.Deprecated('Use getWorkspaceRepoBindingResponseDescriptor instead')
const GetWorkspaceRepoBindingResponse$json = {
  '1': 'GetWorkspaceRepoBindingResponse',
  '2': [
    {
      '1': 'binding',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.WorkspaceRepoBinding',
      '10': 'binding'
    },
    {
      '1': 'overlaps',
      '3': 2,
      '4': 3,
      '5': 11,
      '6': '.agents.v1.RepoBindingOverlap',
      '10': 'overlaps'
    },
  ],
};

/// Descriptor for `GetWorkspaceRepoBindingResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getWorkspaceRepoBindingResponseDescriptor =
    $convert.base64Decode(
        'Ch9HZXRXb3Jrc3BhY2VSZXBvQmluZGluZ1Jlc3BvbnNlEjkKB2JpbmRpbmcYASABKAsyHy5hZ2'
        'VudHMudjEuV29ya3NwYWNlUmVwb0JpbmRpbmdSB2JpbmRpbmcSOQoIb3ZlcmxhcHMYAiADKAsy'
        'HS5hZ2VudHMudjEuUmVwb0JpbmRpbmdPdmVybGFwUghvdmVybGFwcw==');

@$core.Deprecated('Use putWorkspaceRepoBindingRequestDescriptor instead')
const PutWorkspaceRepoBindingRequest$json = {
  '1': 'PutWorkspaceRepoBindingRequest',
  '2': [
    {
      '1': 'binding',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.WorkspaceRepoBinding',
      '10': 'binding'
    },
  ],
};

/// Descriptor for `PutWorkspaceRepoBindingRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List putWorkspaceRepoBindingRequestDescriptor =
    $convert.base64Decode(
        'Ch5QdXRXb3Jrc3BhY2VSZXBvQmluZGluZ1JlcXVlc3QSOQoHYmluZGluZxgBIAEoCzIfLmFnZW'
        '50cy52MS5Xb3Jrc3BhY2VSZXBvQmluZGluZ1IHYmluZGluZw==');

@$core.Deprecated('Use putWorkspaceRepoBindingResponseDescriptor instead')
const PutWorkspaceRepoBindingResponse$json = {
  '1': 'PutWorkspaceRepoBindingResponse',
  '2': [
    {
      '1': 'binding',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.WorkspaceRepoBinding',
      '10': 'binding'
    },
  ],
};

/// Descriptor for `PutWorkspaceRepoBindingResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List putWorkspaceRepoBindingResponseDescriptor =
    $convert.base64Decode(
        'Ch9QdXRXb3Jrc3BhY2VSZXBvQmluZGluZ1Jlc3BvbnNlEjkKB2JpbmRpbmcYASABKAsyHy5hZ2'
        'VudHMudjEuV29ya3NwYWNlUmVwb0JpbmRpbmdSB2JpbmRpbmc=');

@$core.Deprecated('Use deleteWorkspaceRepoBindingRequestDescriptor instead')
const DeleteWorkspaceRepoBindingRequest$json = {
  '1': 'DeleteWorkspaceRepoBindingRequest',
  '2': [
    {
      '1': 'recovery',
      '3': 1,
      '4': 1,
      '5': 14,
      '6': '.agents.v1.RepoBindingDetachRecovery',
      '10': 'recovery'
    },
  ],
};

/// Descriptor for `DeleteWorkspaceRepoBindingRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List deleteWorkspaceRepoBindingRequestDescriptor =
    $convert.base64Decode(
        'CiFEZWxldGVXb3Jrc3BhY2VSZXBvQmluZGluZ1JlcXVlc3QSQAoIcmVjb3ZlcnkYASABKA4yJC'
        '5hZ2VudHMudjEuUmVwb0JpbmRpbmdEZXRhY2hSZWNvdmVyeVIIcmVjb3Zlcnk=');

@$core.Deprecated('Use deleteWorkspaceRepoBindingResponseDescriptor instead')
const DeleteWorkspaceRepoBindingResponse$json = {
  '1': 'DeleteWorkspaceRepoBindingResponse',
  '2': [
    {
      '1': 'agents_materialized',
      '3': 1,
      '4': 1,
      '5': 5,
      '10': 'agentsMaterialized'
    },
  ],
};

/// Descriptor for `DeleteWorkspaceRepoBindingResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List deleteWorkspaceRepoBindingResponseDescriptor =
    $convert.base64Decode(
        'CiJEZWxldGVXb3Jrc3BhY2VSZXBvQmluZGluZ1Jlc3BvbnNlEi8KE2FnZW50c19tYXRlcmlhbG'
        'l6ZWQYASABKAVSEmFnZW50c01hdGVyaWFsaXplZA==');

@$core.Deprecated('Use onboardWorkspaceRepositoryRequestDescriptor instead')
const OnboardWorkspaceRepositoryRequest$json = {
  '1': 'OnboardWorkspaceRepositoryRequest',
  '2': [
    {
      '1': 'mode',
      '3': 1,
      '4': 1,
      '5': 14,
      '6': '.agents.v1.RepoBindingOnboardingMode',
      '10': 'mode'
    },
  ],
};

/// Descriptor for `OnboardWorkspaceRepositoryRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List onboardWorkspaceRepositoryRequestDescriptor =
    $convert.base64Decode(
        'CiFPbmJvYXJkV29ya3NwYWNlUmVwb3NpdG9yeVJlcXVlc3QSOAoEbW9kZRgBIAEoDjIkLmFnZW'
        '50cy52MS5SZXBvQmluZGluZ09uYm9hcmRpbmdNb2RlUgRtb2Rl');

@$core.Deprecated('Use onboardWorkspaceRepositoryResponseDescriptor instead')
const OnboardWorkspaceRepositoryResponse$json = {
  '1': 'OnboardWorkspaceRepositoryResponse',
  '2': [
    {
      '1': 'binding',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.WorkspaceRepoBinding',
      '10': 'binding'
    },
    {'1': 'commit_sha', '3': 2, '4': 1, '5': 9, '10': 'commitSha'},
    {'1': 'published', '3': 3, '4': 1, '5': 8, '10': 'published'},
    {
      '1': 'validation_errors',
      '3': 4,
      '4': 3,
      '5': 9,
      '10': 'validationErrors'
    },
    {'1': 'agents_exported', '3': 5, '4': 1, '5': 5, '10': 'agentsExported'},
    {'1': 'agents_imported', '3': 6, '4': 1, '5': 5, '10': 'agentsImported'},
  ],
};

/// Descriptor for `OnboardWorkspaceRepositoryResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List onboardWorkspaceRepositoryResponseDescriptor = $convert.base64Decode(
    'CiJPbmJvYXJkV29ya3NwYWNlUmVwb3NpdG9yeVJlc3BvbnNlEjkKB2JpbmRpbmcYASABKAsyHy'
    '5hZ2VudHMudjEuV29ya3NwYWNlUmVwb0JpbmRpbmdSB2JpbmRpbmcSHQoKY29tbWl0X3NoYRgC'
    'IAEoCVIJY29tbWl0U2hhEhwKCXB1Ymxpc2hlZBgDIAEoCFIJcHVibGlzaGVkEisKEXZhbGlkYX'
    'Rpb25fZXJyb3JzGAQgAygJUhB2YWxpZGF0aW9uRXJyb3JzEicKD2FnZW50c19leHBvcnRlZBgF'
    'IAEoBVIOYWdlbnRzRXhwb3J0ZWQSJwoPYWdlbnRzX2ltcG9ydGVkGAYgASgFUg5hZ2VudHNJbX'
    'BvcnRlZA==');

@$core.Deprecated(
    'Use setWorkspaceRepoBindingCredentialRequestDescriptor instead')
const SetWorkspaceRepoBindingCredentialRequest$json = {
  '1': 'SetWorkspaceRepoBindingCredentialRequest',
  '2': [
    {'1': 'pat', '3': 1, '4': 1, '5': 9, '8': {}, '10': 'pat'},
  ],
};

/// Descriptor for `SetWorkspaceRepoBindingCredentialRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List setWorkspaceRepoBindingCredentialRequestDescriptor =
    $convert.base64Decode(
        'CihTZXRXb3Jrc3BhY2VSZXBvQmluZGluZ0NyZWRlbnRpYWxSZXF1ZXN0EhkKA3BhdBgBIAEoCU'
        'IH+kIEcgIQAVIDcGF0');

@$core.Deprecated(
    'Use setWorkspaceRepoBindingCredentialResponseDescriptor instead')
const SetWorkspaceRepoBindingCredentialResponse$json = {
  '1': 'SetWorkspaceRepoBindingCredentialResponse',
  '2': [
    {
      '1': 'binding',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.WorkspaceRepoBinding',
      '10': 'binding'
    },
  ],
};

/// Descriptor for `SetWorkspaceRepoBindingCredentialResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List
    setWorkspaceRepoBindingCredentialResponseDescriptor = $convert.base64Decode(
        'CilTZXRXb3Jrc3BhY2VSZXBvQmluZGluZ0NyZWRlbnRpYWxSZXNwb25zZRI5CgdiaW5kaW5nGA'
        'EgASgLMh8uYWdlbnRzLnYxLldvcmtzcGFjZVJlcG9CaW5kaW5nUgdiaW5kaW5n');

@$core.Deprecated('Use validateWorkspaceRepoBindingRequestDescriptor instead')
const ValidateWorkspaceRepoBindingRequest$json = {
  '1': 'ValidateWorkspaceRepoBindingRequest',
};

/// Descriptor for `ValidateWorkspaceRepoBindingRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List validateWorkspaceRepoBindingRequestDescriptor =
    $convert
        .base64Decode('CiNWYWxpZGF0ZVdvcmtzcGFjZVJlcG9CaW5kaW5nUmVxdWVzdA==');

@$core.Deprecated('Use validateWorkspaceRepoBindingResponseDescriptor instead')
const ValidateWorkspaceRepoBindingResponse$json = {
  '1': 'ValidateWorkspaceRepoBindingResponse',
  '2': [
    {
      '1': 'binding',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.WorkspaceRepoBinding',
      '10': 'binding'
    },
  ],
};

/// Descriptor for `ValidateWorkspaceRepoBindingResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List validateWorkspaceRepoBindingResponseDescriptor =
    $convert.base64Decode(
        'CiRWYWxpZGF0ZVdvcmtzcGFjZVJlcG9CaW5kaW5nUmVzcG9uc2USOQoHYmluZGluZxgBIAEoCz'
        'IfLmFnZW50cy52MS5Xb3Jrc3BhY2VSZXBvQmluZGluZ1IHYmluZGluZw==');

@$core.Deprecated('Use syncWorkspaceRepositoryRequestDescriptor instead')
const SyncWorkspaceRepositoryRequest$json = {
  '1': 'SyncWorkspaceRepositoryRequest',
};

/// Descriptor for `SyncWorkspaceRepositoryRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List syncWorkspaceRepositoryRequestDescriptor =
    $convert.base64Decode('Ch5TeW5jV29ya3NwYWNlUmVwb3NpdG9yeVJlcXVlc3Q=');

@$core.Deprecated('Use syncWorkspaceRepositoryResponseDescriptor instead')
const SyncWorkspaceRepositoryResponse$json = {
  '1': 'SyncWorkspaceRepositoryResponse',
  '2': [
    {
      '1': 'binding',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.WorkspaceRepoBinding',
      '10': 'binding'
    },
    {'1': 'entries_synced', '3': 2, '4': 1, '5': 5, '10': 'entriesSynced'},
    {'1': 'published', '3': 3, '4': 1, '5': 8, '10': 'published'},
    {
      '1': 'publication_errors',
      '3': 4,
      '4': 3,
      '5': 9,
      '10': 'publicationErrors'
    },
  ],
};

/// Descriptor for `SyncWorkspaceRepositoryResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List syncWorkspaceRepositoryResponseDescriptor =
    $convert.base64Decode(
        'Ch9TeW5jV29ya3NwYWNlUmVwb3NpdG9yeVJlc3BvbnNlEjkKB2JpbmRpbmcYASABKAsyHy5hZ2'
        'VudHMudjEuV29ya3NwYWNlUmVwb0JpbmRpbmdSB2JpbmRpbmcSJQoOZW50cmllc19zeW5jZWQY'
        'AiABKAVSDWVudHJpZXNTeW5jZWQSHAoJcHVibGlzaGVkGAMgASgIUglwdWJsaXNoZWQSLQoScH'
        'VibGljYXRpb25fZXJyb3JzGAQgAygJUhFwdWJsaWNhdGlvbkVycm9ycw==');

@$core.Deprecated('Use getRepositorySyncStatusRequestDescriptor instead')
const GetRepositorySyncStatusRequest$json = {
  '1': 'GetRepositorySyncStatusRequest',
};

/// Descriptor for `GetRepositorySyncStatusRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getRepositorySyncStatusRequestDescriptor =
    $convert.base64Decode('Ch5HZXRSZXBvc2l0b3J5U3luY1N0YXR1c1JlcXVlc3Q=');

@$core.Deprecated('Use getRepositorySyncStatusResponseDescriptor instead')
const GetRepositorySyncStatusResponse$json = {
  '1': 'GetRepositorySyncStatusResponse',
  '2': [
    {
      '1': 'binding',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.WorkspaceRepoBinding',
      '10': 'binding'
    },
  ],
};

/// Descriptor for `GetRepositorySyncStatusResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getRepositorySyncStatusResponseDescriptor =
    $convert.base64Decode(
        'Ch9HZXRSZXBvc2l0b3J5U3luY1N0YXR1c1Jlc3BvbnNlEjkKB2JpbmRpbmcYASABKAsyHy5hZ2'
        'VudHMudjEuV29ya3NwYWNlUmVwb0JpbmRpbmdSB2JpbmRpbmc=');

@$core.Deprecated('Use listRepositoryEntriesRequestDescriptor instead')
const ListRepositoryEntriesRequest$json = {
  '1': 'ListRepositoryEntriesRequest',
  '2': [
    {'1': 'path', '3': 1, '4': 1, '5': 9, '10': 'path'},
  ],
};

/// Descriptor for `ListRepositoryEntriesRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listRepositoryEntriesRequestDescriptor =
    $convert.base64Decode(
        'ChxMaXN0UmVwb3NpdG9yeUVudHJpZXNSZXF1ZXN0EhIKBHBhdGgYASABKAlSBHBhdGg=');

@$core.Deprecated('Use listRepositoryEntriesResponseDescriptor instead')
const ListRepositoryEntriesResponse$json = {
  '1': 'ListRepositoryEntriesResponse',
  '2': [
    {'1': 'commit_sha', '3': 1, '4': 1, '5': 9, '10': 'commitSha'},
    {
      '1': 'entries',
      '3': 2,
      '4': 3,
      '5': 11,
      '6': '.agents.v1.RepoCacheEntry',
      '10': 'entries'
    },
    {
      '1': 'observed_commit_sha',
      '3': 3,
      '4': 1,
      '5': 9,
      '10': 'observedCommitSha'
    },
    {'1': 'active_commit_sha', '3': 4, '4': 1, '5': 9, '10': 'activeCommitSha'},
  ],
};

/// Descriptor for `ListRepositoryEntriesResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listRepositoryEntriesResponseDescriptor = $convert.base64Decode(
    'Ch1MaXN0UmVwb3NpdG9yeUVudHJpZXNSZXNwb25zZRIdCgpjb21taXRfc2hhGAEgASgJUgljb2'
    '1taXRTaGESMwoHZW50cmllcxgCIAMoCzIZLmFnZW50cy52MS5SZXBvQ2FjaGVFbnRyeVIHZW50'
    'cmllcxIuChNvYnNlcnZlZF9jb21taXRfc2hhGAMgASgJUhFvYnNlcnZlZENvbW1pdFNoYRIqCh'
    'FhY3RpdmVfY29tbWl0X3NoYRgEIAEoCVIPYWN0aXZlQ29tbWl0U2hh');

@$core.Deprecated('Use getRepositoryFileRequestDescriptor instead')
const GetRepositoryFileRequest$json = {
  '1': 'GetRepositoryFileRequest',
  '2': [
    {'1': 'path', '3': 1, '4': 1, '5': 9, '10': 'path'},
  ],
};

/// Descriptor for `GetRepositoryFileRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getRepositoryFileRequestDescriptor =
    $convert.base64Decode(
        'ChhHZXRSZXBvc2l0b3J5RmlsZVJlcXVlc3QSEgoEcGF0aBgBIAEoCVIEcGF0aA==');

@$core.Deprecated('Use getRepositoryFileResponseDescriptor instead')
const GetRepositoryFileResponse$json = {
  '1': 'GetRepositoryFileResponse',
  '2': [
    {'1': 'commit_sha', '3': 1, '4': 1, '5': 9, '10': 'commitSha'},
    {
      '1': 'entry',
      '3': 2,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.RepoCacheEntry',
      '10': 'entry'
    },
    {'1': 'content', '3': 3, '4': 1, '5': 9, '10': 'content'},
    {
      '1': 'observed_commit_sha',
      '3': 4,
      '4': 1,
      '5': 9,
      '10': 'observedCommitSha'
    },
    {'1': 'active_commit_sha', '3': 5, '4': 1, '5': 9, '10': 'activeCommitSha'},
  ],
};

/// Descriptor for `GetRepositoryFileResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getRepositoryFileResponseDescriptor = $convert.base64Decode(
    'ChlHZXRSZXBvc2l0b3J5RmlsZVJlc3BvbnNlEh0KCmNvbW1pdF9zaGEYASABKAlSCWNvbW1pdF'
    'NoYRIvCgVlbnRyeRgCIAEoCzIZLmFnZW50cy52MS5SZXBvQ2FjaGVFbnRyeVIFZW50cnkSGAoH'
    'Y29udGVudBgDIAEoCVIHY29udGVudBIuChNvYnNlcnZlZF9jb21taXRfc2hhGAQgASgJUhFvYn'
    'NlcnZlZENvbW1pdFNoYRIqChFhY3RpdmVfY29tbWl0X3NoYRgFIAEoCVIPYWN0aXZlQ29tbWl0'
    'U2hh');

@$core.Deprecated('Use publishWorkspaceRepositoryRequestDescriptor instead')
const PublishWorkspaceRepositoryRequest$json = {
  '1': 'PublishWorkspaceRepositoryRequest',
};

/// Descriptor for `PublishWorkspaceRepositoryRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List publishWorkspaceRepositoryRequestDescriptor =
    $convert.base64Decode('CiFQdWJsaXNoV29ya3NwYWNlUmVwb3NpdG9yeVJlcXVlc3Q=');

@$core.Deprecated('Use publishWorkspaceRepositoryResponseDescriptor instead')
const PublishWorkspaceRepositoryResponse$json = {
  '1': 'PublishWorkspaceRepositoryResponse',
  '2': [
    {
      '1': 'binding',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.WorkspaceRepoBinding',
      '10': 'binding'
    },
    {
      '1': 'validation_errors',
      '3': 2,
      '4': 3,
      '5': 9,
      '10': 'validationErrors'
    },
  ],
};

/// Descriptor for `PublishWorkspaceRepositoryResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List publishWorkspaceRepositoryResponseDescriptor =
    $convert.base64Decode(
        'CiJQdWJsaXNoV29ya3NwYWNlUmVwb3NpdG9yeVJlc3BvbnNlEjkKB2JpbmRpbmcYASABKAsyHy'
        '5hZ2VudHMudjEuV29ya3NwYWNlUmVwb0JpbmRpbmdSB2JpbmRpbmcSKwoRdmFsaWRhdGlvbl9l'
        'cnJvcnMYAiADKAlSEHZhbGlkYXRpb25FcnJvcnM=');

@$core.Deprecated('Use configureWebhookSecretRequestDescriptor instead')
const ConfigureWebhookSecretRequest$json = {
  '1': 'ConfigureWebhookSecretRequest',
};

/// Descriptor for `ConfigureWebhookSecretRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List configureWebhookSecretRequestDescriptor =
    $convert.base64Decode('Ch1Db25maWd1cmVXZWJob29rU2VjcmV0UmVxdWVzdA==');

@$core.Deprecated('Use configureWebhookSecretResponseDescriptor instead')
const ConfigureWebhookSecretResponse$json = {
  '1': 'ConfigureWebhookSecretResponse',
  '2': [
    {
      '1': 'binding',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.WorkspaceRepoBinding',
      '10': 'binding'
    },
    {'1': 'webhook_secret', '3': 2, '4': 1, '5': 9, '10': 'webhookSecret'},
    {'1': 'callback_url', '3': 3, '4': 1, '5': 9, '10': 'callbackUrl'},
  ],
};

/// Descriptor for `ConfigureWebhookSecretResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List configureWebhookSecretResponseDescriptor =
    $convert.base64Decode(
        'Ch5Db25maWd1cmVXZWJob29rU2VjcmV0UmVzcG9uc2USOQoHYmluZGluZxgBIAEoCzIfLmFnZW'
        '50cy52MS5Xb3Jrc3BhY2VSZXBvQmluZGluZ1IHYmluZGluZxIlCg53ZWJob29rX3NlY3JldBgC'
        'IAEoCVINd2ViaG9va1NlY3JldBIhCgxjYWxsYmFja191cmwYAyABKAlSC2NhbGxiYWNrVXJs');

@$core.Deprecated('Use acceptRepositoryBaselineRequestDescriptor instead')
const AcceptRepositoryBaselineRequest$json = {
  '1': 'AcceptRepositoryBaselineRequest',
};

/// Descriptor for `AcceptRepositoryBaselineRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List acceptRepositoryBaselineRequestDescriptor =
    $convert.base64Decode('Ch9BY2NlcHRSZXBvc2l0b3J5QmFzZWxpbmVSZXF1ZXN0');

@$core.Deprecated('Use acceptRepositoryBaselineResponseDescriptor instead')
const AcceptRepositoryBaselineResponse$json = {
  '1': 'AcceptRepositoryBaselineResponse',
  '2': [
    {
      '1': 'binding',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.WorkspaceRepoBinding',
      '10': 'binding'
    },
  ],
};

/// Descriptor for `AcceptRepositoryBaselineResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List acceptRepositoryBaselineResponseDescriptor =
    $convert.base64Decode(
        'CiBBY2NlcHRSZXBvc2l0b3J5QmFzZWxpbmVSZXNwb25zZRI5CgdiaW5kaW5nGAEgASgLMh8uYW'
        'dlbnRzLnYxLldvcmtzcGFjZVJlcG9CaW5kaW5nUgdiaW5kaW5n');

@$core.Deprecated('Use contentFileActionDescriptor instead')
const ContentFileAction$json = {
  '1': 'ContentFileAction',
  '2': [
    {'1': 'path', '3': 1, '4': 1, '5': 9, '8': {}, '10': 'path'},
    {
      '1': 'operation',
      '3': 2,
      '4': 1,
      '5': 14,
      '6': '.agents.v1.ContentFileOperation',
      '10': 'operation'
    },
    {'1': 'content', '3': 3, '4': 1, '5': 9, '10': 'content'},
  ],
};

/// Descriptor for `ContentFileAction`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List contentFileActionDescriptor = $convert.base64Decode(
    'ChFDb250ZW50RmlsZUFjdGlvbhIbCgRwYXRoGAEgASgJQgf6QgRyAhABUgRwYXRoEj0KCW9wZX'
    'JhdGlvbhgCIAEoDjIfLmFnZW50cy52MS5Db250ZW50RmlsZU9wZXJhdGlvblIJb3BlcmF0aW9u'
    'EhgKB2NvbnRlbnQYAyABKAlSB2NvbnRlbnQ=');

@$core.Deprecated('Use commitAgentContentRequestDescriptor instead')
const CommitAgentContentRequest$json = {
  '1': 'CommitAgentContentRequest',
  '2': [
    {
      '1': 'actions',
      '3': 1,
      '4': 3,
      '5': 11,
      '6': '.agents.v1.ContentFileAction',
      '8': {},
      '10': 'actions'
    },
    {'1': 'message', '3': 2, '4': 1, '5': 9, '10': 'message'},
    {'1': 'base_revision', '3': 3, '4': 1, '5': 9, '10': 'baseRevision'},
  ],
};

/// Descriptor for `CommitAgentContentRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List commitAgentContentRequestDescriptor = $convert.base64Decode(
    'ChlDb21taXRBZ2VudENvbnRlbnRSZXF1ZXN0EkAKB2FjdGlvbnMYASADKAsyHC5hZ2VudHMudj'
    'EuQ29udGVudEZpbGVBY3Rpb25CCPpCBZIBAggBUgdhY3Rpb25zEhgKB21lc3NhZ2UYAiABKAlS'
    'B21lc3NhZ2USIwoNYmFzZV9yZXZpc2lvbhgDIAEoCVIMYmFzZVJldmlzaW9u');

@$core.Deprecated('Use commitAgentContentResponseDescriptor instead')
const CommitAgentContentResponse$json = {
  '1': 'CommitAgentContentResponse',
  '2': [
    {
      '1': 'binding',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.WorkspaceRepoBinding',
      '10': 'binding'
    },
    {'1': 'commit_sha', '3': 2, '4': 1, '5': 9, '10': 'commitSha'},
    {
      '1': 'validation_errors',
      '3': 3,
      '4': 3,
      '5': 9,
      '10': 'validationErrors'
    },
    {
      '1': 'change_request_url',
      '3': 4,
      '4': 1,
      '5': 9,
      '10': 'changeRequestUrl'
    },
  ],
};

/// Descriptor for `CommitAgentContentResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List commitAgentContentResponseDescriptor = $convert.base64Decode(
    'ChpDb21taXRBZ2VudENvbnRlbnRSZXNwb25zZRI5CgdiaW5kaW5nGAEgASgLMh8uYWdlbnRzLn'
    'YxLldvcmtzcGFjZVJlcG9CaW5kaW5nUgdiaW5kaW5nEh0KCmNvbW1pdF9zaGEYAiABKAlSCWNv'
    'bW1pdFNoYRIrChF2YWxpZGF0aW9uX2Vycm9ycxgDIAMoCVIQdmFsaWRhdGlvbkVycm9ycxIsCh'
    'JjaGFuZ2VfcmVxdWVzdF91cmwYBCABKAlSEGNoYW5nZVJlcXVlc3RVcmw=');

@$core.Deprecated('Use rollbackAgentContentRequestDescriptor instead')
const RollbackAgentContentRequest$json = {
  '1': 'RollbackAgentContentRequest',
  '2': [
    {
      '1': 'target_commit_sha',
      '3': 1,
      '4': 1,
      '5': 9,
      '8': {},
      '10': 'targetCommitSha'
    },
  ],
};

/// Descriptor for `RollbackAgentContentRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List rollbackAgentContentRequestDescriptor =
    $convert.base64Decode(
        'ChtSb2xsYmFja0FnZW50Q29udGVudFJlcXVlc3QSMwoRdGFyZ2V0X2NvbW1pdF9zaGEYASABKA'
        'lCB/pCBHICEAFSD3RhcmdldENvbW1pdFNoYQ==');

@$core.Deprecated('Use rollbackAgentContentResponseDescriptor instead')
const RollbackAgentContentResponse$json = {
  '1': 'RollbackAgentContentResponse',
  '2': [
    {
      '1': 'binding',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.WorkspaceRepoBinding',
      '10': 'binding'
    },
    {'1': 'commit_sha', '3': 2, '4': 1, '5': 9, '10': 'commitSha'},
    {
      '1': 'validation_errors',
      '3': 3,
      '4': 3,
      '5': 9,
      '10': 'validationErrors'
    },
    {
      '1': 'change_request_url',
      '3': 4,
      '4': 1,
      '5': 9,
      '10': 'changeRequestUrl'
    },
  ],
};

/// Descriptor for `RollbackAgentContentResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List rollbackAgentContentResponseDescriptor = $convert.base64Decode(
    'ChxSb2xsYmFja0FnZW50Q29udGVudFJlc3BvbnNlEjkKB2JpbmRpbmcYASABKAsyHy5hZ2VudH'
    'MudjEuV29ya3NwYWNlUmVwb0JpbmRpbmdSB2JpbmRpbmcSHQoKY29tbWl0X3NoYRgCIAEoCVIJ'
    'Y29tbWl0U2hhEisKEXZhbGlkYXRpb25fZXJyb3JzGAMgAygJUhB2YWxpZGF0aW9uRXJyb3JzEi'
    'wKEmNoYW5nZV9yZXF1ZXN0X3VybBgEIAEoCVIQY2hhbmdlUmVxdWVzdFVybA==');

@$core.Deprecated('Use purgeAgentContentRequestDescriptor instead')
const PurgeAgentContentRequest$json = {
  '1': 'PurgeAgentContentRequest',
  '2': [
    {'1': 'agent_id', '3': 1, '4': 1, '5': 9, '8': {}, '10': 'agentId'},
  ],
};

/// Descriptor for `PurgeAgentContentRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List purgeAgentContentRequestDescriptor =
    $convert.base64Decode(
        'ChhQdXJnZUFnZW50Q29udGVudFJlcXVlc3QSIgoIYWdlbnRfaWQYASABKAlCB/pCBHICEAFSB2'
        'FnZW50SWQ=');

@$core.Deprecated('Use purgeAgentContentResponseDescriptor instead')
const PurgeAgentContentResponse$json = {
  '1': 'PurgeAgentContentResponse',
  '2': [
    {
      '1': 'binding',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.WorkspaceRepoBinding',
      '10': 'binding'
    },
    {'1': 'commit_sha', '3': 2, '4': 1, '5': 9, '10': 'commitSha'},
    {
      '1': 'validation_errors',
      '3': 3,
      '4': 3,
      '5': 9,
      '10': 'validationErrors'
    },
    {
      '1': 'change_request_url',
      '3': 4,
      '4': 1,
      '5': 9,
      '10': 'changeRequestUrl'
    },
  ],
};

/// Descriptor for `PurgeAgentContentResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List purgeAgentContentResponseDescriptor = $convert.base64Decode(
    'ChlQdXJnZUFnZW50Q29udGVudFJlc3BvbnNlEjkKB2JpbmRpbmcYASABKAsyHy5hZ2VudHMudj'
    'EuV29ya3NwYWNlUmVwb0JpbmRpbmdSB2JpbmRpbmcSHQoKY29tbWl0X3NoYRgCIAEoCVIJY29t'
    'bWl0U2hhEisKEXZhbGlkYXRpb25fZXJyb3JzGAMgAygJUhB2YWxpZGF0aW9uRXJyb3JzEiwKEm'
    'NoYW5nZV9yZXF1ZXN0X3VybBgEIAEoCVIQY2hhbmdlUmVxdWVzdFVybA==');

@$core.Deprecated('Use repoCacheEntryDescriptor instead')
const RepoCacheEntry$json = {
  '1': 'RepoCacheEntry',
  '2': [
    {'1': 'path', '3': 1, '4': 1, '5': 9, '10': 'path'},
    {
      '1': 'kind',
      '3': 2,
      '4': 1,
      '5': 14,
      '6': '.agents.v1.RepoCacheEntryKind',
      '10': 'kind'
    },
    {'1': 'size', '3': 3, '4': 1, '5': 3, '10': 'size'},
    {'1': 'content_hash', '3': 4, '4': 1, '5': 9, '10': 'contentHash'},
    {'1': 'claimed', '3': 5, '4': 1, '5': 8, '10': 'claimed'},
  ],
};

/// Descriptor for `RepoCacheEntry`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List repoCacheEntryDescriptor = $convert.base64Decode(
    'Cg5SZXBvQ2FjaGVFbnRyeRISCgRwYXRoGAEgASgJUgRwYXRoEjEKBGtpbmQYAiABKA4yHS5hZ2'
    'VudHMudjEuUmVwb0NhY2hlRW50cnlLaW5kUgRraW5kEhIKBHNpemUYAyABKANSBHNpemUSIQoM'
    'Y29udGVudF9oYXNoGAQgASgJUgtjb250ZW50SGFzaBIYCgdjbGFpbWVkGAUgASgIUgdjbGFpbW'
    'Vk');

const $core.Map<$core.String, $core.dynamic>
    WorkspaceRepoBindingServiceBase$json = {
  '1': 'WorkspaceRepoBindingService',
  '2': [
    {
      '1': 'GetWorkspaceRepoBinding',
      '2': '.agents.v1.GetWorkspaceRepoBindingRequest',
      '3': '.agents.v1.GetWorkspaceRepoBindingResponse'
    },
    {
      '1': 'PutWorkspaceRepoBinding',
      '2': '.agents.v1.PutWorkspaceRepoBindingRequest',
      '3': '.agents.v1.PutWorkspaceRepoBindingResponse'
    },
    {
      '1': 'DeleteWorkspaceRepoBinding',
      '2': '.agents.v1.DeleteWorkspaceRepoBindingRequest',
      '3': '.agents.v1.DeleteWorkspaceRepoBindingResponse'
    },
    {
      '1': 'OnboardWorkspaceRepository',
      '2': '.agents.v1.OnboardWorkspaceRepositoryRequest',
      '3': '.agents.v1.OnboardWorkspaceRepositoryResponse'
    },
    {
      '1': 'SetWorkspaceRepoBindingCredential',
      '2': '.agents.v1.SetWorkspaceRepoBindingCredentialRequest',
      '3': '.agents.v1.SetWorkspaceRepoBindingCredentialResponse'
    },
    {
      '1': 'ValidateWorkspaceRepoBinding',
      '2': '.agents.v1.ValidateWorkspaceRepoBindingRequest',
      '3': '.agents.v1.ValidateWorkspaceRepoBindingResponse'
    },
    {
      '1': 'SyncWorkspaceRepository',
      '2': '.agents.v1.SyncWorkspaceRepositoryRequest',
      '3': '.agents.v1.SyncWorkspaceRepositoryResponse'
    },
    {
      '1': 'GetRepositorySyncStatus',
      '2': '.agents.v1.GetRepositorySyncStatusRequest',
      '3': '.agents.v1.GetRepositorySyncStatusResponse'
    },
    {
      '1': 'ListRepositoryEntries',
      '2': '.agents.v1.ListRepositoryEntriesRequest',
      '3': '.agents.v1.ListRepositoryEntriesResponse'
    },
    {
      '1': 'GetRepositoryFile',
      '2': '.agents.v1.GetRepositoryFileRequest',
      '3': '.agents.v1.GetRepositoryFileResponse'
    },
    {
      '1': 'PublishWorkspaceRepository',
      '2': '.agents.v1.PublishWorkspaceRepositoryRequest',
      '3': '.agents.v1.PublishWorkspaceRepositoryResponse'
    },
    {
      '1': 'ConfigureWebhookSecret',
      '2': '.agents.v1.ConfigureWebhookSecretRequest',
      '3': '.agents.v1.ConfigureWebhookSecretResponse'
    },
    {
      '1': 'AcceptRepositoryBaseline',
      '2': '.agents.v1.AcceptRepositoryBaselineRequest',
      '3': '.agents.v1.AcceptRepositoryBaselineResponse'
    },
    {
      '1': 'CommitAgentContent',
      '2': '.agents.v1.CommitAgentContentRequest',
      '3': '.agents.v1.CommitAgentContentResponse'
    },
    {
      '1': 'RollbackAgentContent',
      '2': '.agents.v1.RollbackAgentContentRequest',
      '3': '.agents.v1.RollbackAgentContentResponse'
    },
    {
      '1': 'PurgeAgentContent',
      '2': '.agents.v1.PurgeAgentContentRequest',
      '3': '.agents.v1.PurgeAgentContentResponse'
    },
  ],
};

@$core.Deprecated('Use workspaceRepoBindingServiceDescriptor instead')
const $core.Map<$core.String, $core.Map<$core.String, $core.dynamic>>
    WorkspaceRepoBindingServiceBase$messageJson = {
  '.agents.v1.GetWorkspaceRepoBindingRequest':
      GetWorkspaceRepoBindingRequest$json,
  '.agents.v1.GetWorkspaceRepoBindingResponse':
      GetWorkspaceRepoBindingResponse$json,
  '.agents.v1.WorkspaceRepoBinding': WorkspaceRepoBinding$json,
  '.google.protobuf.Timestamp': $0.Timestamp$json,
  '.agents.v1.RepoBindingStatus': RepoBindingStatus$json,
  '.agents.v1.RepoBindingCheck': RepoBindingCheck$json,
  '.agents.v1.RepoBindingOverlap': RepoBindingOverlap$json,
  '.agents.v1.PutWorkspaceRepoBindingRequest':
      PutWorkspaceRepoBindingRequest$json,
  '.agents.v1.PutWorkspaceRepoBindingResponse':
      PutWorkspaceRepoBindingResponse$json,
  '.agents.v1.DeleteWorkspaceRepoBindingRequest':
      DeleteWorkspaceRepoBindingRequest$json,
  '.agents.v1.DeleteWorkspaceRepoBindingResponse':
      DeleteWorkspaceRepoBindingResponse$json,
  '.agents.v1.OnboardWorkspaceRepositoryRequest':
      OnboardWorkspaceRepositoryRequest$json,
  '.agents.v1.OnboardWorkspaceRepositoryResponse':
      OnboardWorkspaceRepositoryResponse$json,
  '.agents.v1.SetWorkspaceRepoBindingCredentialRequest':
      SetWorkspaceRepoBindingCredentialRequest$json,
  '.agents.v1.SetWorkspaceRepoBindingCredentialResponse':
      SetWorkspaceRepoBindingCredentialResponse$json,
  '.agents.v1.ValidateWorkspaceRepoBindingRequest':
      ValidateWorkspaceRepoBindingRequest$json,
  '.agents.v1.ValidateWorkspaceRepoBindingResponse':
      ValidateWorkspaceRepoBindingResponse$json,
  '.agents.v1.SyncWorkspaceRepositoryRequest':
      SyncWorkspaceRepositoryRequest$json,
  '.agents.v1.SyncWorkspaceRepositoryResponse':
      SyncWorkspaceRepositoryResponse$json,
  '.agents.v1.GetRepositorySyncStatusRequest':
      GetRepositorySyncStatusRequest$json,
  '.agents.v1.GetRepositorySyncStatusResponse':
      GetRepositorySyncStatusResponse$json,
  '.agents.v1.ListRepositoryEntriesRequest': ListRepositoryEntriesRequest$json,
  '.agents.v1.ListRepositoryEntriesResponse':
      ListRepositoryEntriesResponse$json,
  '.agents.v1.RepoCacheEntry': RepoCacheEntry$json,
  '.agents.v1.GetRepositoryFileRequest': GetRepositoryFileRequest$json,
  '.agents.v1.GetRepositoryFileResponse': GetRepositoryFileResponse$json,
  '.agents.v1.PublishWorkspaceRepositoryRequest':
      PublishWorkspaceRepositoryRequest$json,
  '.agents.v1.PublishWorkspaceRepositoryResponse':
      PublishWorkspaceRepositoryResponse$json,
  '.agents.v1.ConfigureWebhookSecretRequest':
      ConfigureWebhookSecretRequest$json,
  '.agents.v1.ConfigureWebhookSecretResponse':
      ConfigureWebhookSecretResponse$json,
  '.agents.v1.AcceptRepositoryBaselineRequest':
      AcceptRepositoryBaselineRequest$json,
  '.agents.v1.AcceptRepositoryBaselineResponse':
      AcceptRepositoryBaselineResponse$json,
  '.agents.v1.CommitAgentContentRequest': CommitAgentContentRequest$json,
  '.agents.v1.ContentFileAction': ContentFileAction$json,
  '.agents.v1.CommitAgentContentResponse': CommitAgentContentResponse$json,
  '.agents.v1.RollbackAgentContentRequest': RollbackAgentContentRequest$json,
  '.agents.v1.RollbackAgentContentResponse': RollbackAgentContentResponse$json,
  '.agents.v1.PurgeAgentContentRequest': PurgeAgentContentRequest$json,
  '.agents.v1.PurgeAgentContentResponse': PurgeAgentContentResponse$json,
};

/// Descriptor for `WorkspaceRepoBindingService`. Decode as a `google.protobuf.ServiceDescriptorProto`.
final $typed_data.Uint8List workspaceRepoBindingServiceDescriptor = $convert.base64Decode(
    'ChtXb3Jrc3BhY2VSZXBvQmluZGluZ1NlcnZpY2UScAoXR2V0V29ya3NwYWNlUmVwb0JpbmRpbm'
    'cSKS5hZ2VudHMudjEuR2V0V29ya3NwYWNlUmVwb0JpbmRpbmdSZXF1ZXN0GiouYWdlbnRzLnYx'
    'LkdldFdvcmtzcGFjZVJlcG9CaW5kaW5nUmVzcG9uc2UScAoXUHV0V29ya3NwYWNlUmVwb0Jpbm'
    'RpbmcSKS5hZ2VudHMudjEuUHV0V29ya3NwYWNlUmVwb0JpbmRpbmdSZXF1ZXN0GiouYWdlbnRz'
    'LnYxLlB1dFdvcmtzcGFjZVJlcG9CaW5kaW5nUmVzcG9uc2USeQoaRGVsZXRlV29ya3NwYWNlUm'
    'Vwb0JpbmRpbmcSLC5hZ2VudHMudjEuRGVsZXRlV29ya3NwYWNlUmVwb0JpbmRpbmdSZXF1ZXN0'
    'Gi0uYWdlbnRzLnYxLkRlbGV0ZVdvcmtzcGFjZVJlcG9CaW5kaW5nUmVzcG9uc2USeQoaT25ib2'
    'FyZFdvcmtzcGFjZVJlcG9zaXRvcnkSLC5hZ2VudHMudjEuT25ib2FyZFdvcmtzcGFjZVJlcG9z'
    'aXRvcnlSZXF1ZXN0Gi0uYWdlbnRzLnYxLk9uYm9hcmRXb3Jrc3BhY2VSZXBvc2l0b3J5UmVzcG'
    '9uc2USjgEKIVNldFdvcmtzcGFjZVJlcG9CaW5kaW5nQ3JlZGVudGlhbBIzLmFnZW50cy52MS5T'
    'ZXRXb3Jrc3BhY2VSZXBvQmluZGluZ0NyZWRlbnRpYWxSZXF1ZXN0GjQuYWdlbnRzLnYxLlNldF'
    'dvcmtzcGFjZVJlcG9CaW5kaW5nQ3JlZGVudGlhbFJlc3BvbnNlEn8KHFZhbGlkYXRlV29ya3Nw'
    'YWNlUmVwb0JpbmRpbmcSLi5hZ2VudHMudjEuVmFsaWRhdGVXb3Jrc3BhY2VSZXBvQmluZGluZ1'
    'JlcXVlc3QaLy5hZ2VudHMudjEuVmFsaWRhdGVXb3Jrc3BhY2VSZXBvQmluZGluZ1Jlc3BvbnNl'
    'EnAKF1N5bmNXb3Jrc3BhY2VSZXBvc2l0b3J5EikuYWdlbnRzLnYxLlN5bmNXb3Jrc3BhY2VSZX'
    'Bvc2l0b3J5UmVxdWVzdBoqLmFnZW50cy52MS5TeW5jV29ya3NwYWNlUmVwb3NpdG9yeVJlc3Bv'
    'bnNlEnAKF0dldFJlcG9zaXRvcnlTeW5jU3RhdHVzEikuYWdlbnRzLnYxLkdldFJlcG9zaXRvcn'
    'lTeW5jU3RhdHVzUmVxdWVzdBoqLmFnZW50cy52MS5HZXRSZXBvc2l0b3J5U3luY1N0YXR1c1Jl'
    'c3BvbnNlEmoKFUxpc3RSZXBvc2l0b3J5RW50cmllcxInLmFnZW50cy52MS5MaXN0UmVwb3NpdG'
    '9yeUVudHJpZXNSZXF1ZXN0GiguYWdlbnRzLnYxLkxpc3RSZXBvc2l0b3J5RW50cmllc1Jlc3Bv'
    'bnNlEl4KEUdldFJlcG9zaXRvcnlGaWxlEiMuYWdlbnRzLnYxLkdldFJlcG9zaXRvcnlGaWxlUm'
    'VxdWVzdBokLmFnZW50cy52MS5HZXRSZXBvc2l0b3J5RmlsZVJlc3BvbnNlEnkKGlB1Ymxpc2hX'
    'b3Jrc3BhY2VSZXBvc2l0b3J5EiwuYWdlbnRzLnYxLlB1Ymxpc2hXb3Jrc3BhY2VSZXBvc2l0b3'
    'J5UmVxdWVzdBotLmFnZW50cy52MS5QdWJsaXNoV29ya3NwYWNlUmVwb3NpdG9yeVJlc3BvbnNl'
    'Em0KFkNvbmZpZ3VyZVdlYmhvb2tTZWNyZXQSKC5hZ2VudHMudjEuQ29uZmlndXJlV2ViaG9va1'
    'NlY3JldFJlcXVlc3QaKS5hZ2VudHMudjEuQ29uZmlndXJlV2ViaG9va1NlY3JldFJlc3BvbnNl'
    'EnMKGEFjY2VwdFJlcG9zaXRvcnlCYXNlbGluZRIqLmFnZW50cy52MS5BY2NlcHRSZXBvc2l0b3'
    'J5QmFzZWxpbmVSZXF1ZXN0GisuYWdlbnRzLnYxLkFjY2VwdFJlcG9zaXRvcnlCYXNlbGluZVJl'
    'c3BvbnNlEmEKEkNvbW1pdEFnZW50Q29udGVudBIkLmFnZW50cy52MS5Db21taXRBZ2VudENvbn'
    'RlbnRSZXF1ZXN0GiUuYWdlbnRzLnYxLkNvbW1pdEFnZW50Q29udGVudFJlc3BvbnNlEmcKFFJv'
    'bGxiYWNrQWdlbnRDb250ZW50EiYuYWdlbnRzLnYxLlJvbGxiYWNrQWdlbnRDb250ZW50UmVxdW'
    'VzdBonLmFnZW50cy52MS5Sb2xsYmFja0FnZW50Q29udGVudFJlc3BvbnNlEl4KEVB1cmdlQWdl'
    'bnRDb250ZW50EiMuYWdlbnRzLnYxLlB1cmdlQWdlbnRDb250ZW50UmVxdWVzdBokLmFnZW50cy'
    '52MS5QdXJnZUFnZW50Q29udGVudFJlc3BvbnNl');
