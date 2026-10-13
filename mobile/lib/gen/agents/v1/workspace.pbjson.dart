// This is a generated file - do not edit.
//
// Generated from agents/v1/workspace.proto.

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

@$core.Deprecated('Use workspaceDescriptor instead')
const Workspace$json = {
  '1': 'Workspace',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
    {'1': 'name', '3': 2, '4': 1, '5': 9, '8': {}, '10': 'name'},
    {'1': 'slug', '3': 3, '4': 1, '5': 9, '8': {}, '10': 'slug'},
    {'1': 'description', '3': 4, '4': 1, '5': 9, '10': 'description'},
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
  ],
};

/// Descriptor for `Workspace`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List workspaceDescriptor = $convert.base64Decode(
    'CglXb3Jrc3BhY2USDgoCaWQYASABKAlSAmlkEhsKBG5hbWUYAiABKAlCB/pCBHICEAFSBG5hbW'
    'USGwoEc2x1ZxgDIAEoCUIH+kIEcgIQAVIEc2x1ZxIgCgtkZXNjcmlwdGlvbhgEIAEoCVILZGVz'
    'Y3JpcHRpb24SOQoKY3JlYXRlZF9hdBgFIAEoCzIaLmdvb2dsZS5wcm90b2J1Zi5UaW1lc3RhbX'
    'BSCWNyZWF0ZWRBdBI5Cgp1cGRhdGVkX2F0GAYgASgLMhouZ29vZ2xlLnByb3RvYnVmLlRpbWVz'
    'dGFtcFIJdXBkYXRlZEF0');

@$core.Deprecated('Use workspaceMemberDescriptor instead')
const WorkspaceMember$json = {
  '1': 'WorkspaceMember',
  '2': [
    {'1': 'workspace_id', '3': 1, '4': 1, '5': 9, '10': 'workspaceId'},
    {'1': 'user_id', '3': 2, '4': 1, '5': 9, '10': 'userId'},
    {'1': 'role', '3': 3, '4': 1, '5': 9, '10': 'role'},
    {
      '1': 'created_at',
      '3': 4,
      '4': 1,
      '5': 11,
      '6': '.google.protobuf.Timestamp',
      '10': 'createdAt'
    },
  ],
};

/// Descriptor for `WorkspaceMember`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List workspaceMemberDescriptor = $convert.base64Decode(
    'Cg9Xb3Jrc3BhY2VNZW1iZXISIQoMd29ya3NwYWNlX2lkGAEgASgJUgt3b3Jrc3BhY2VJZBIXCg'
    'd1c2VyX2lkGAIgASgJUgZ1c2VySWQSEgoEcm9sZRgDIAEoCVIEcm9sZRI5CgpjcmVhdGVkX2F0'
    'GAQgASgLMhouZ29vZ2xlLnByb3RvYnVmLlRpbWVzdGFtcFIJY3JlYXRlZEF0');

@$core.Deprecated('Use listWorkspacesRequestDescriptor instead')
const ListWorkspacesRequest$json = {
  '1': 'ListWorkspacesRequest',
};

/// Descriptor for `ListWorkspacesRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listWorkspacesRequestDescriptor =
    $convert.base64Decode('ChVMaXN0V29ya3NwYWNlc1JlcXVlc3Q=');

@$core.Deprecated('Use listWorkspacesResponseDescriptor instead')
const ListWorkspacesResponse$json = {
  '1': 'ListWorkspacesResponse',
  '2': [
    {
      '1': 'workspaces',
      '3': 1,
      '4': 3,
      '5': 11,
      '6': '.agents.v1.Workspace',
      '10': 'workspaces'
    },
  ],
};

/// Descriptor for `ListWorkspacesResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listWorkspacesResponseDescriptor =
    $convert.base64Decode(
        'ChZMaXN0V29ya3NwYWNlc1Jlc3BvbnNlEjQKCndvcmtzcGFjZXMYASADKAsyFC5hZ2VudHMudj'
        'EuV29ya3NwYWNlUgp3b3Jrc3BhY2Vz');

@$core.Deprecated('Use getWorkspaceRequestDescriptor instead')
const GetWorkspaceRequest$json = {
  '1': 'GetWorkspaceRequest',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '8': {}, '10': 'id'},
  ],
};

/// Descriptor for `GetWorkspaceRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getWorkspaceRequestDescriptor =
    $convert.base64Decode(
        'ChNHZXRXb3Jrc3BhY2VSZXF1ZXN0EhcKAmlkGAEgASgJQgf6QgRyAhABUgJpZA==');

@$core.Deprecated('Use getWorkspaceResponseDescriptor instead')
const GetWorkspaceResponse$json = {
  '1': 'GetWorkspaceResponse',
  '2': [
    {
      '1': 'workspace',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.Workspace',
      '10': 'workspace'
    },
  ],
};

/// Descriptor for `GetWorkspaceResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getWorkspaceResponseDescriptor = $convert.base64Decode(
    'ChRHZXRXb3Jrc3BhY2VSZXNwb25zZRIyCgl3b3Jrc3BhY2UYASABKAsyFC5hZ2VudHMudjEuV2'
    '9ya3NwYWNlUgl3b3Jrc3BhY2U=');

@$core.Deprecated('Use createWorkspaceRequestDescriptor instead')
const CreateWorkspaceRequest$json = {
  '1': 'CreateWorkspaceRequest',
  '2': [
    {
      '1': 'workspace',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.Workspace',
      '10': 'workspace'
    },
  ],
};

/// Descriptor for `CreateWorkspaceRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List createWorkspaceRequestDescriptor =
    $convert.base64Decode(
        'ChZDcmVhdGVXb3Jrc3BhY2VSZXF1ZXN0EjIKCXdvcmtzcGFjZRgBIAEoCzIULmFnZW50cy52MS'
        '5Xb3Jrc3BhY2VSCXdvcmtzcGFjZQ==');

@$core.Deprecated('Use createWorkspaceResponseDescriptor instead')
const CreateWorkspaceResponse$json = {
  '1': 'CreateWorkspaceResponse',
  '2': [
    {
      '1': 'workspace',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.Workspace',
      '10': 'workspace'
    },
  ],
};

/// Descriptor for `CreateWorkspaceResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List createWorkspaceResponseDescriptor =
    $convert.base64Decode(
        'ChdDcmVhdGVXb3Jrc3BhY2VSZXNwb25zZRIyCgl3b3Jrc3BhY2UYASABKAsyFC5hZ2VudHMudj'
        'EuV29ya3NwYWNlUgl3b3Jrc3BhY2U=');

@$core.Deprecated('Use updateWorkspaceRequestDescriptor instead')
const UpdateWorkspaceRequest$json = {
  '1': 'UpdateWorkspaceRequest',
  '2': [
    {
      '1': 'workspace',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.Workspace',
      '10': 'workspace'
    },
  ],
};

/// Descriptor for `UpdateWorkspaceRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List updateWorkspaceRequestDescriptor =
    $convert.base64Decode(
        'ChZVcGRhdGVXb3Jrc3BhY2VSZXF1ZXN0EjIKCXdvcmtzcGFjZRgBIAEoCzIULmFnZW50cy52MS'
        '5Xb3Jrc3BhY2VSCXdvcmtzcGFjZQ==');

@$core.Deprecated('Use updateWorkspaceResponseDescriptor instead')
const UpdateWorkspaceResponse$json = {
  '1': 'UpdateWorkspaceResponse',
  '2': [
    {
      '1': 'workspace',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.Workspace',
      '10': 'workspace'
    },
  ],
};

/// Descriptor for `UpdateWorkspaceResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List updateWorkspaceResponseDescriptor =
    $convert.base64Decode(
        'ChdVcGRhdGVXb3Jrc3BhY2VSZXNwb25zZRIyCgl3b3Jrc3BhY2UYASABKAsyFC5hZ2VudHMudj'
        'EuV29ya3NwYWNlUgl3b3Jrc3BhY2U=');

@$core.Deprecated('Use deleteWorkspaceRequestDescriptor instead')
const DeleteWorkspaceRequest$json = {
  '1': 'DeleteWorkspaceRequest',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '8': {}, '10': 'id'},
  ],
};

/// Descriptor for `DeleteWorkspaceRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List deleteWorkspaceRequestDescriptor =
    $convert.base64Decode(
        'ChZEZWxldGVXb3Jrc3BhY2VSZXF1ZXN0EhcKAmlkGAEgASgJQgf6QgRyAhABUgJpZA==');

@$core.Deprecated('Use deleteWorkspaceResponseDescriptor instead')
const DeleteWorkspaceResponse$json = {
  '1': 'DeleteWorkspaceResponse',
};

/// Descriptor for `DeleteWorkspaceResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List deleteWorkspaceResponseDescriptor =
    $convert.base64Decode('ChdEZWxldGVXb3Jrc3BhY2VSZXNwb25zZQ==');

@$core.Deprecated('Use listWorkspaceMembersRequestDescriptor instead')
const ListWorkspaceMembersRequest$json = {
  '1': 'ListWorkspaceMembersRequest',
  '2': [
    {'1': 'workspace_id', '3': 1, '4': 1, '5': 9, '8': {}, '10': 'workspaceId'},
  ],
};

/// Descriptor for `ListWorkspaceMembersRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listWorkspaceMembersRequestDescriptor =
    $convert.base64Decode(
        'ChtMaXN0V29ya3NwYWNlTWVtYmVyc1JlcXVlc3QSKgoMd29ya3NwYWNlX2lkGAEgASgJQgf6Qg'
        'RyAhABUgt3b3Jrc3BhY2VJZA==');

@$core.Deprecated('Use listWorkspaceMembersResponseDescriptor instead')
const ListWorkspaceMembersResponse$json = {
  '1': 'ListWorkspaceMembersResponse',
  '2': [
    {
      '1': 'members',
      '3': 1,
      '4': 3,
      '5': 11,
      '6': '.agents.v1.WorkspaceMember',
      '10': 'members'
    },
  ],
};

/// Descriptor for `ListWorkspaceMembersResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listWorkspaceMembersResponseDescriptor =
    $convert.base64Decode(
        'ChxMaXN0V29ya3NwYWNlTWVtYmVyc1Jlc3BvbnNlEjQKB21lbWJlcnMYASADKAsyGi5hZ2VudH'
        'MudjEuV29ya3NwYWNlTWVtYmVyUgdtZW1iZXJz');

@$core.Deprecated('Use addWorkspaceMemberRequestDescriptor instead')
const AddWorkspaceMemberRequest$json = {
  '1': 'AddWorkspaceMemberRequest',
  '2': [
    {'1': 'workspace_id', '3': 1, '4': 1, '5': 9, '8': {}, '10': 'workspaceId'},
    {'1': 'user_id', '3': 2, '4': 1, '5': 9, '8': {}, '10': 'userId'},
    {'1': 'role', '3': 3, '4': 1, '5': 9, '10': 'role'},
  ],
};

/// Descriptor for `AddWorkspaceMemberRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List addWorkspaceMemberRequestDescriptor = $convert.base64Decode(
    'ChlBZGRXb3Jrc3BhY2VNZW1iZXJSZXF1ZXN0EioKDHdvcmtzcGFjZV9pZBgBIAEoCUIH+kIEcg'
    'IQAVILd29ya3NwYWNlSWQSIAoHdXNlcl9pZBgCIAEoCUIH+kIEcgIQAVIGdXNlcklkEhIKBHJv'
    'bGUYAyABKAlSBHJvbGU=');

@$core.Deprecated('Use addWorkspaceMemberResponseDescriptor instead')
const AddWorkspaceMemberResponse$json = {
  '1': 'AddWorkspaceMemberResponse',
  '2': [
    {
      '1': 'member',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.WorkspaceMember',
      '10': 'member'
    },
  ],
};

/// Descriptor for `AddWorkspaceMemberResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List addWorkspaceMemberResponseDescriptor =
    $convert.base64Decode(
        'ChpBZGRXb3Jrc3BhY2VNZW1iZXJSZXNwb25zZRIyCgZtZW1iZXIYASABKAsyGi5hZ2VudHMudj'
        'EuV29ya3NwYWNlTWVtYmVyUgZtZW1iZXI=');

@$core.Deprecated('Use updateWorkspaceMemberRequestDescriptor instead')
const UpdateWorkspaceMemberRequest$json = {
  '1': 'UpdateWorkspaceMemberRequest',
  '2': [
    {'1': 'workspace_id', '3': 1, '4': 1, '5': 9, '8': {}, '10': 'workspaceId'},
    {'1': 'user_id', '3': 2, '4': 1, '5': 9, '8': {}, '10': 'userId'},
    {'1': 'role', '3': 3, '4': 1, '5': 9, '8': {}, '10': 'role'},
  ],
};

/// Descriptor for `UpdateWorkspaceMemberRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List updateWorkspaceMemberRequestDescriptor =
    $convert.base64Decode(
        'ChxVcGRhdGVXb3Jrc3BhY2VNZW1iZXJSZXF1ZXN0EioKDHdvcmtzcGFjZV9pZBgBIAEoCUIH+k'
        'IEcgIQAVILd29ya3NwYWNlSWQSIAoHdXNlcl9pZBgCIAEoCUIH+kIEcgIQAVIGdXNlcklkEhsK'
        'BHJvbGUYAyABKAlCB/pCBHICEAFSBHJvbGU=');

@$core.Deprecated('Use updateWorkspaceMemberResponseDescriptor instead')
const UpdateWorkspaceMemberResponse$json = {
  '1': 'UpdateWorkspaceMemberResponse',
  '2': [
    {
      '1': 'member',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.agents.v1.WorkspaceMember',
      '10': 'member'
    },
  ],
};

/// Descriptor for `UpdateWorkspaceMemberResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List updateWorkspaceMemberResponseDescriptor =
    $convert.base64Decode(
        'Ch1VcGRhdGVXb3Jrc3BhY2VNZW1iZXJSZXNwb25zZRIyCgZtZW1iZXIYASABKAsyGi5hZ2VudH'
        'MudjEuV29ya3NwYWNlTWVtYmVyUgZtZW1iZXI=');

@$core.Deprecated('Use removeWorkspaceMemberRequestDescriptor instead')
const RemoveWorkspaceMemberRequest$json = {
  '1': 'RemoveWorkspaceMemberRequest',
  '2': [
    {'1': 'workspace_id', '3': 1, '4': 1, '5': 9, '8': {}, '10': 'workspaceId'},
    {'1': 'user_id', '3': 2, '4': 1, '5': 9, '8': {}, '10': 'userId'},
  ],
};

/// Descriptor for `RemoveWorkspaceMemberRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List removeWorkspaceMemberRequestDescriptor =
    $convert.base64Decode(
        'ChxSZW1vdmVXb3Jrc3BhY2VNZW1iZXJSZXF1ZXN0EioKDHdvcmtzcGFjZV9pZBgBIAEoCUIH+k'
        'IEcgIQAVILd29ya3NwYWNlSWQSIAoHdXNlcl9pZBgCIAEoCUIH+kIEcgIQAVIGdXNlcklk');

@$core.Deprecated('Use removeWorkspaceMemberResponseDescriptor instead')
const RemoveWorkspaceMemberResponse$json = {
  '1': 'RemoveWorkspaceMemberResponse',
};

/// Descriptor for `RemoveWorkspaceMemberResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List removeWorkspaceMemberResponseDescriptor =
    $convert.base64Decode('Ch1SZW1vdmVXb3Jrc3BhY2VNZW1iZXJSZXNwb25zZQ==');

const $core.Map<$core.String, $core.dynamic> WorkspaceServiceBase$json = {
  '1': 'WorkspaceService',
  '2': [
    {
      '1': 'ListWorkspaces',
      '2': '.agents.v1.ListWorkspacesRequest',
      '3': '.agents.v1.ListWorkspacesResponse'
    },
    {
      '1': 'GetWorkspace',
      '2': '.agents.v1.GetWorkspaceRequest',
      '3': '.agents.v1.GetWorkspaceResponse'
    },
    {
      '1': 'CreateWorkspace',
      '2': '.agents.v1.CreateWorkspaceRequest',
      '3': '.agents.v1.CreateWorkspaceResponse'
    },
    {
      '1': 'UpdateWorkspace',
      '2': '.agents.v1.UpdateWorkspaceRequest',
      '3': '.agents.v1.UpdateWorkspaceResponse'
    },
    {
      '1': 'DeleteWorkspace',
      '2': '.agents.v1.DeleteWorkspaceRequest',
      '3': '.agents.v1.DeleteWorkspaceResponse'
    },
    {
      '1': 'ListWorkspaceMembers',
      '2': '.agents.v1.ListWorkspaceMembersRequest',
      '3': '.agents.v1.ListWorkspaceMembersResponse'
    },
    {
      '1': 'AddWorkspaceMember',
      '2': '.agents.v1.AddWorkspaceMemberRequest',
      '3': '.agents.v1.AddWorkspaceMemberResponse'
    },
    {
      '1': 'UpdateWorkspaceMember',
      '2': '.agents.v1.UpdateWorkspaceMemberRequest',
      '3': '.agents.v1.UpdateWorkspaceMemberResponse'
    },
    {
      '1': 'RemoveWorkspaceMember',
      '2': '.agents.v1.RemoveWorkspaceMemberRequest',
      '3': '.agents.v1.RemoveWorkspaceMemberResponse'
    },
  ],
};

@$core.Deprecated('Use workspaceServiceDescriptor instead')
const $core.Map<$core.String, $core.Map<$core.String, $core.dynamic>>
    WorkspaceServiceBase$messageJson = {
  '.agents.v1.ListWorkspacesRequest': ListWorkspacesRequest$json,
  '.agents.v1.ListWorkspacesResponse': ListWorkspacesResponse$json,
  '.agents.v1.Workspace': Workspace$json,
  '.google.protobuf.Timestamp': $0.Timestamp$json,
  '.agents.v1.GetWorkspaceRequest': GetWorkspaceRequest$json,
  '.agents.v1.GetWorkspaceResponse': GetWorkspaceResponse$json,
  '.agents.v1.CreateWorkspaceRequest': CreateWorkspaceRequest$json,
  '.agents.v1.CreateWorkspaceResponse': CreateWorkspaceResponse$json,
  '.agents.v1.UpdateWorkspaceRequest': UpdateWorkspaceRequest$json,
  '.agents.v1.UpdateWorkspaceResponse': UpdateWorkspaceResponse$json,
  '.agents.v1.DeleteWorkspaceRequest': DeleteWorkspaceRequest$json,
  '.agents.v1.DeleteWorkspaceResponse': DeleteWorkspaceResponse$json,
  '.agents.v1.ListWorkspaceMembersRequest': ListWorkspaceMembersRequest$json,
  '.agents.v1.ListWorkspaceMembersResponse': ListWorkspaceMembersResponse$json,
  '.agents.v1.WorkspaceMember': WorkspaceMember$json,
  '.agents.v1.AddWorkspaceMemberRequest': AddWorkspaceMemberRequest$json,
  '.agents.v1.AddWorkspaceMemberResponse': AddWorkspaceMemberResponse$json,
  '.agents.v1.UpdateWorkspaceMemberRequest': UpdateWorkspaceMemberRequest$json,
  '.agents.v1.UpdateWorkspaceMemberResponse':
      UpdateWorkspaceMemberResponse$json,
  '.agents.v1.RemoveWorkspaceMemberRequest': RemoveWorkspaceMemberRequest$json,
  '.agents.v1.RemoveWorkspaceMemberResponse':
      RemoveWorkspaceMemberResponse$json,
};

/// Descriptor for `WorkspaceService`. Decode as a `google.protobuf.ServiceDescriptorProto`.
final $typed_data.Uint8List workspaceServiceDescriptor = $convert.base64Decode(
    'ChBXb3Jrc3BhY2VTZXJ2aWNlElUKDkxpc3RXb3Jrc3BhY2VzEiAuYWdlbnRzLnYxLkxpc3RXb3'
    'Jrc3BhY2VzUmVxdWVzdBohLmFnZW50cy52MS5MaXN0V29ya3NwYWNlc1Jlc3BvbnNlEk8KDEdl'
    'dFdvcmtzcGFjZRIeLmFnZW50cy52MS5HZXRXb3Jrc3BhY2VSZXF1ZXN0Gh8uYWdlbnRzLnYxLk'
    'dldFdvcmtzcGFjZVJlc3BvbnNlElgKD0NyZWF0ZVdvcmtzcGFjZRIhLmFnZW50cy52MS5DcmVh'
    'dGVXb3Jrc3BhY2VSZXF1ZXN0GiIuYWdlbnRzLnYxLkNyZWF0ZVdvcmtzcGFjZVJlc3BvbnNlEl'
    'gKD1VwZGF0ZVdvcmtzcGFjZRIhLmFnZW50cy52MS5VcGRhdGVXb3Jrc3BhY2VSZXF1ZXN0GiIu'
    'YWdlbnRzLnYxLlVwZGF0ZVdvcmtzcGFjZVJlc3BvbnNlElgKD0RlbGV0ZVdvcmtzcGFjZRIhLm'
    'FnZW50cy52MS5EZWxldGVXb3Jrc3BhY2VSZXF1ZXN0GiIuYWdlbnRzLnYxLkRlbGV0ZVdvcmtz'
    'cGFjZVJlc3BvbnNlEmcKFExpc3RXb3Jrc3BhY2VNZW1iZXJzEiYuYWdlbnRzLnYxLkxpc3RXb3'
    'Jrc3BhY2VNZW1iZXJzUmVxdWVzdBonLmFnZW50cy52MS5MaXN0V29ya3NwYWNlTWVtYmVyc1Jl'
    'c3BvbnNlEmEKEkFkZFdvcmtzcGFjZU1lbWJlchIkLmFnZW50cy52MS5BZGRXb3Jrc3BhY2VNZW'
    '1iZXJSZXF1ZXN0GiUuYWdlbnRzLnYxLkFkZFdvcmtzcGFjZU1lbWJlclJlc3BvbnNlEmoKFVVw'
    'ZGF0ZVdvcmtzcGFjZU1lbWJlchInLmFnZW50cy52MS5VcGRhdGVXb3Jrc3BhY2VNZW1iZXJSZX'
    'F1ZXN0GiguYWdlbnRzLnYxLlVwZGF0ZVdvcmtzcGFjZU1lbWJlclJlc3BvbnNlEmoKFVJlbW92'
    'ZVdvcmtzcGFjZU1lbWJlchInLmFnZW50cy52MS5SZW1vdmVXb3Jrc3BhY2VNZW1iZXJSZXF1ZX'
    'N0GiguYWdlbnRzLnYxLlJlbW92ZVdvcmtzcGFjZU1lbWJlclJlc3BvbnNl');
