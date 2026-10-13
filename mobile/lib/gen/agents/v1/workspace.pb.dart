// This is a generated file - do not edit.
//
// Generated from agents/v1/workspace.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names, prefer_relative_imports

import 'dart:async' as $async;
import 'dart:core' as $core;

import 'package:protobuf/protobuf.dart' as $pb;
import 'package:protobuf/well_known_types/google/protobuf/timestamp.pb.dart'
    as $0;

export 'package:protobuf/protobuf.dart' show GeneratedMessageGenericExtensions;

class Workspace extends $pb.GeneratedMessage {
  factory Workspace({
    $core.String? id,
    $core.String? name,
    $core.String? slug,
    $core.String? description,
    $0.Timestamp? createdAt,
    $0.Timestamp? updatedAt,
  }) {
    final result = Workspace._();
    if (id != null) result.id = id;
    if (name != null) result.name = name;
    if (slug != null) result.slug = slug;
    if (description != null) result.description = description;
    if (createdAt != null) result.createdAt = createdAt;
    if (updatedAt != null) result.updatedAt = updatedAt;
    return result;
  }

  Workspace._();

  factory Workspace.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      Workspace()..mergeFromBuffer(data, registry);
  factory Workspace.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      Workspace()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'Workspace',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: Workspace.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..aOS(2, _omitFieldNames ? '' : 'name')
    ..aOS(3, _omitFieldNames ? '' : 'slug')
    ..aOS(4, _omitFieldNames ? '' : 'description')
    ..aOM<$0.Timestamp>(5, _omitFieldNames ? '' : 'createdAt',
        subBuilder: $0.Timestamp.$_createMessage)
    ..aOM<$0.Timestamp>(6, _omitFieldNames ? '' : 'updatedAt',
        subBuilder: $0.Timestamp.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Workspace clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Workspace copyWith(void Function(Workspace) updates) =>
      super.copyWith((message) => updates(message as Workspace)) as Workspace;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use Workspace() / Workspace.new instead')
  static Workspace create() => Workspace._();
  static $pb.GeneratedMessage $_createMessage() => Workspace._();
  @$core.override
  Workspace createEmptyInstance() => Workspace._();
  @$core.pragma('dart2js:noInline')
  static Workspace getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<Workspace>(Workspace.$_createMessage);
  static Workspace? _defaultInstance;

  /// Stable server-assigned identifier (UUID).
  /// @gotags: json:"id,omitempty" yaml:"id,omitempty" bson:"_id,omitempty"
  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);

  /// Human-readable name.
  /// @gotags: json:"name,omitempty" yaml:"name,omitempty" bson:"name,omitempty"
  @$pb.TagNumber(2)
  $core.String get name => $_getSZ(1);
  @$pb.TagNumber(2)
  set name($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasName() => $_has(1);
  @$pb.TagNumber(2)
  void clearName() => $_clearField(2);

  /// URL-safe short identifier, unique across workspaces.
  /// @gotags: json:"slug,omitempty" yaml:"slug,omitempty" bson:"slug,omitempty"
  @$pb.TagNumber(3)
  $core.String get slug => $_getSZ(2);
  @$pb.TagNumber(3)
  set slug($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasSlug() => $_has(2);
  @$pb.TagNumber(3)
  void clearSlug() => $_clearField(3);

  /// Free-form description.
  /// @gotags: json:"description,omitempty" yaml:"description,omitempty" bson:"description,omitempty"
  @$pb.TagNumber(4)
  $core.String get description => $_getSZ(3);
  @$pb.TagNumber(4)
  set description($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasDescription() => $_has(3);
  @$pb.TagNumber(4)
  void clearDescription() => $_clearField(4);

  /// @gotags: json:"created_at,omitempty" yaml:"created_at,omitempty" bson:"created_at,omitempty"
  @$pb.TagNumber(5)
  $0.Timestamp get createdAt => $_getN(4);
  @$pb.TagNumber(5)
  set createdAt($0.Timestamp value) => $_setField(5, value);
  @$pb.TagNumber(5)
  $core.bool hasCreatedAt() => $_has(4);
  @$pb.TagNumber(5)
  void clearCreatedAt() => $_clearField(5);
  @$pb.TagNumber(5)
  $0.Timestamp ensureCreatedAt() => $_ensure(4);

  /// @gotags: json:"updated_at,omitempty" yaml:"updated_at,omitempty" bson:"updated_at,omitempty"
  @$pb.TagNumber(6)
  $0.Timestamp get updatedAt => $_getN(5);
  @$pb.TagNumber(6)
  set updatedAt($0.Timestamp value) => $_setField(6, value);
  @$pb.TagNumber(6)
  $core.bool hasUpdatedAt() => $_has(5);
  @$pb.TagNumber(6)
  void clearUpdatedAt() => $_clearField(6);
  @$pb.TagNumber(6)
  $0.Timestamp ensureUpdatedAt() => $_ensure(5);
}

/// WorkspaceMember is a (workspace, user) binding with a role inside that
/// workspace. Global "admin" users still see every workspace.
class WorkspaceMember extends $pb.GeneratedMessage {
  factory WorkspaceMember({
    $core.String? workspaceId,
    $core.String? userId,
    $core.String? role,
    $0.Timestamp? createdAt,
  }) {
    final result = WorkspaceMember._();
    if (workspaceId != null) result.workspaceId = workspaceId;
    if (userId != null) result.userId = userId;
    if (role != null) result.role = role;
    if (createdAt != null) result.createdAt = createdAt;
    return result;
  }

  WorkspaceMember._();

  factory WorkspaceMember.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      WorkspaceMember()..mergeFromBuffer(data, registry);
  factory WorkspaceMember.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      WorkspaceMember()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'WorkspaceMember',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: WorkspaceMember.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'workspaceId')
    ..aOS(2, _omitFieldNames ? '' : 'userId')
    ..aOS(3, _omitFieldNames ? '' : 'role')
    ..aOM<$0.Timestamp>(4, _omitFieldNames ? '' : 'createdAt',
        subBuilder: $0.Timestamp.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  WorkspaceMember clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  WorkspaceMember copyWith(void Function(WorkspaceMember) updates) =>
      super.copyWith((message) => updates(message as WorkspaceMember))
          as WorkspaceMember;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use WorkspaceMember() / WorkspaceMember.new instead')
  static WorkspaceMember create() => WorkspaceMember._();
  static $pb.GeneratedMessage $_createMessage() => WorkspaceMember._();
  @$core.override
  WorkspaceMember createEmptyInstance() => WorkspaceMember._();
  @$core.pragma('dart2js:noInline')
  static WorkspaceMember getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<WorkspaceMember>(
          WorkspaceMember.$_createMessage);
  static WorkspaceMember? _defaultInstance;

  /// @gotags: json:"workspace_id,omitempty" yaml:"workspace_id,omitempty" bson:"workspace_id,omitempty"
  @$pb.TagNumber(1)
  $core.String get workspaceId => $_getSZ(0);
  @$pb.TagNumber(1)
  set workspaceId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasWorkspaceId() => $_has(0);
  @$pb.TagNumber(1)
  void clearWorkspaceId() => $_clearField(1);

  /// @gotags: json:"user_id,omitempty" yaml:"user_id,omitempty" bson:"user_id,omitempty"
  @$pb.TagNumber(2)
  $core.String get userId => $_getSZ(1);
  @$pb.TagNumber(2)
  set userId($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasUserId() => $_has(1);
  @$pb.TagNumber(2)
  void clearUserId() => $_clearField(2);

  /// Role inside the workspace: "owner", "admin", "member".
  /// @gotags: json:"role,omitempty" yaml:"role,omitempty" bson:"role,omitempty"
  @$pb.TagNumber(3)
  $core.String get role => $_getSZ(2);
  @$pb.TagNumber(3)
  set role($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasRole() => $_has(2);
  @$pb.TagNumber(3)
  void clearRole() => $_clearField(3);

  /// @gotags: json:"created_at,omitempty" yaml:"created_at,omitempty" bson:"created_at,omitempty"
  @$pb.TagNumber(4)
  $0.Timestamp get createdAt => $_getN(3);
  @$pb.TagNumber(4)
  set createdAt($0.Timestamp value) => $_setField(4, value);
  @$pb.TagNumber(4)
  $core.bool hasCreatedAt() => $_has(3);
  @$pb.TagNumber(4)
  void clearCreatedAt() => $_clearField(4);
  @$pb.TagNumber(4)
  $0.Timestamp ensureCreatedAt() => $_ensure(3);
}

class ListWorkspacesRequest extends $pb.GeneratedMessage {
  factory ListWorkspacesRequest() => ListWorkspacesRequest._();

  ListWorkspacesRequest._();

  factory ListWorkspacesRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListWorkspacesRequest()..mergeFromBuffer(data, registry);
  factory ListWorkspacesRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListWorkspacesRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListWorkspacesRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: ListWorkspacesRequest.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListWorkspacesRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListWorkspacesRequest copyWith(
          void Function(ListWorkspacesRequest) updates) =>
      super.copyWith((message) => updates(message as ListWorkspacesRequest))
          as ListWorkspacesRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use ListWorkspacesRequest() / ListWorkspacesRequest.new instead')
  static ListWorkspacesRequest create() => ListWorkspacesRequest._();
  static $pb.GeneratedMessage $_createMessage() => ListWorkspacesRequest._();
  @$core.override
  ListWorkspacesRequest createEmptyInstance() => ListWorkspacesRequest._();
  @$core.pragma('dart2js:noInline')
  static ListWorkspacesRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ListWorkspacesRequest>(
          ListWorkspacesRequest.$_createMessage);
  static ListWorkspacesRequest? _defaultInstance;
}

class ListWorkspacesResponse extends $pb.GeneratedMessage {
  factory ListWorkspacesResponse({
    $core.Iterable<Workspace>? workspaces,
  }) {
    final result = ListWorkspacesResponse._();
    if (workspaces != null) result.workspaces.addAll(workspaces);
    return result;
  }

  ListWorkspacesResponse._();

  factory ListWorkspacesResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListWorkspacesResponse()..mergeFromBuffer(data, registry);
  factory ListWorkspacesResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListWorkspacesResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListWorkspacesResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: ListWorkspacesResponse.$_createMessage)
    ..pPM<Workspace>(1, _omitFieldNames ? '' : 'workspaces',
        subBuilder: Workspace.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListWorkspacesResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListWorkspacesResponse copyWith(
          void Function(ListWorkspacesResponse) updates) =>
      super.copyWith((message) => updates(message as ListWorkspacesResponse))
          as ListWorkspacesResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use ListWorkspacesResponse() / ListWorkspacesResponse.new instead')
  static ListWorkspacesResponse create() => ListWorkspacesResponse._();
  static $pb.GeneratedMessage $_createMessage() => ListWorkspacesResponse._();
  @$core.override
  ListWorkspacesResponse createEmptyInstance() => ListWorkspacesResponse._();
  @$core.pragma('dart2js:noInline')
  static ListWorkspacesResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ListWorkspacesResponse>(
          ListWorkspacesResponse.$_createMessage);
  static ListWorkspacesResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $pb.PbList<Workspace> get workspaces => $_getList(0);
}

class GetWorkspaceRequest extends $pb.GeneratedMessage {
  factory GetWorkspaceRequest({
    $core.String? id,
  }) {
    final result = GetWorkspaceRequest._();
    if (id != null) result.id = id;
    return result;
  }

  GetWorkspaceRequest._();

  factory GetWorkspaceRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetWorkspaceRequest()..mergeFromBuffer(data, registry);
  factory GetWorkspaceRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetWorkspaceRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetWorkspaceRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: GetWorkspaceRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetWorkspaceRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetWorkspaceRequest copyWith(void Function(GetWorkspaceRequest) updates) =>
      super.copyWith((message) => updates(message as GetWorkspaceRequest))
          as GetWorkspaceRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core
      .Deprecated('Use GetWorkspaceRequest() / GetWorkspaceRequest.new instead')
  static GetWorkspaceRequest create() => GetWorkspaceRequest._();
  static $pb.GeneratedMessage $_createMessage() => GetWorkspaceRequest._();
  @$core.override
  GetWorkspaceRequest createEmptyInstance() => GetWorkspaceRequest._();
  @$core.pragma('dart2js:noInline')
  static GetWorkspaceRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<GetWorkspaceRequest>(
          GetWorkspaceRequest.$_createMessage);
  static GetWorkspaceRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);
}

class GetWorkspaceResponse extends $pb.GeneratedMessage {
  factory GetWorkspaceResponse({
    Workspace? workspace,
  }) {
    final result = GetWorkspaceResponse._();
    if (workspace != null) result.workspace = workspace;
    return result;
  }

  GetWorkspaceResponse._();

  factory GetWorkspaceResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetWorkspaceResponse()..mergeFromBuffer(data, registry);
  factory GetWorkspaceResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetWorkspaceResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetWorkspaceResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: GetWorkspaceResponse.$_createMessage)
    ..aOM<Workspace>(1, _omitFieldNames ? '' : 'workspace',
        subBuilder: Workspace.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetWorkspaceResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetWorkspaceResponse copyWith(void Function(GetWorkspaceResponse) updates) =>
      super.copyWith((message) => updates(message as GetWorkspaceResponse))
          as GetWorkspaceResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use GetWorkspaceResponse() / GetWorkspaceResponse.new instead')
  static GetWorkspaceResponse create() => GetWorkspaceResponse._();
  static $pb.GeneratedMessage $_createMessage() => GetWorkspaceResponse._();
  @$core.override
  GetWorkspaceResponse createEmptyInstance() => GetWorkspaceResponse._();
  @$core.pragma('dart2js:noInline')
  static GetWorkspaceResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<GetWorkspaceResponse>(
          GetWorkspaceResponse.$_createMessage);
  static GetWorkspaceResponse? _defaultInstance;

  @$pb.TagNumber(1)
  Workspace get workspace => $_getN(0);
  @$pb.TagNumber(1)
  set workspace(Workspace value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasWorkspace() => $_has(0);
  @$pb.TagNumber(1)
  void clearWorkspace() => $_clearField(1);
  @$pb.TagNumber(1)
  Workspace ensureWorkspace() => $_ensure(0);
}

class CreateWorkspaceRequest extends $pb.GeneratedMessage {
  factory CreateWorkspaceRequest({
    Workspace? workspace,
  }) {
    final result = CreateWorkspaceRequest._();
    if (workspace != null) result.workspace = workspace;
    return result;
  }

  CreateWorkspaceRequest._();

  factory CreateWorkspaceRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateWorkspaceRequest()..mergeFromBuffer(data, registry);
  factory CreateWorkspaceRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateWorkspaceRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CreateWorkspaceRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: CreateWorkspaceRequest.$_createMessage)
    ..aOM<Workspace>(1, _omitFieldNames ? '' : 'workspace',
        subBuilder: Workspace.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateWorkspaceRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateWorkspaceRequest copyWith(
          void Function(CreateWorkspaceRequest) updates) =>
      super.copyWith((message) => updates(message as CreateWorkspaceRequest))
          as CreateWorkspaceRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use CreateWorkspaceRequest() / CreateWorkspaceRequest.new instead')
  static CreateWorkspaceRequest create() => CreateWorkspaceRequest._();
  static $pb.GeneratedMessage $_createMessage() => CreateWorkspaceRequest._();
  @$core.override
  CreateWorkspaceRequest createEmptyInstance() => CreateWorkspaceRequest._();
  @$core.pragma('dart2js:noInline')
  static CreateWorkspaceRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<CreateWorkspaceRequest>(
          CreateWorkspaceRequest.$_createMessage);
  static CreateWorkspaceRequest? _defaultInstance;

  @$pb.TagNumber(1)
  Workspace get workspace => $_getN(0);
  @$pb.TagNumber(1)
  set workspace(Workspace value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasWorkspace() => $_has(0);
  @$pb.TagNumber(1)
  void clearWorkspace() => $_clearField(1);
  @$pb.TagNumber(1)
  Workspace ensureWorkspace() => $_ensure(0);
}

class CreateWorkspaceResponse extends $pb.GeneratedMessage {
  factory CreateWorkspaceResponse({
    Workspace? workspace,
  }) {
    final result = CreateWorkspaceResponse._();
    if (workspace != null) result.workspace = workspace;
    return result;
  }

  CreateWorkspaceResponse._();

  factory CreateWorkspaceResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateWorkspaceResponse()..mergeFromBuffer(data, registry);
  factory CreateWorkspaceResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateWorkspaceResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CreateWorkspaceResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: CreateWorkspaceResponse.$_createMessage)
    ..aOM<Workspace>(1, _omitFieldNames ? '' : 'workspace',
        subBuilder: Workspace.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateWorkspaceResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateWorkspaceResponse copyWith(
          void Function(CreateWorkspaceResponse) updates) =>
      super.copyWith((message) => updates(message as CreateWorkspaceResponse))
          as CreateWorkspaceResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use CreateWorkspaceResponse() / CreateWorkspaceResponse.new instead')
  static CreateWorkspaceResponse create() => CreateWorkspaceResponse._();
  static $pb.GeneratedMessage $_createMessage() => CreateWorkspaceResponse._();
  @$core.override
  CreateWorkspaceResponse createEmptyInstance() => CreateWorkspaceResponse._();
  @$core.pragma('dart2js:noInline')
  static CreateWorkspaceResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<CreateWorkspaceResponse>(
          CreateWorkspaceResponse.$_createMessage);
  static CreateWorkspaceResponse? _defaultInstance;

  @$pb.TagNumber(1)
  Workspace get workspace => $_getN(0);
  @$pb.TagNumber(1)
  set workspace(Workspace value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasWorkspace() => $_has(0);
  @$pb.TagNumber(1)
  void clearWorkspace() => $_clearField(1);
  @$pb.TagNumber(1)
  Workspace ensureWorkspace() => $_ensure(0);
}

class UpdateWorkspaceRequest extends $pb.GeneratedMessage {
  factory UpdateWorkspaceRequest({
    Workspace? workspace,
  }) {
    final result = UpdateWorkspaceRequest._();
    if (workspace != null) result.workspace = workspace;
    return result;
  }

  UpdateWorkspaceRequest._();

  factory UpdateWorkspaceRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpdateWorkspaceRequest()..mergeFromBuffer(data, registry);
  factory UpdateWorkspaceRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpdateWorkspaceRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'UpdateWorkspaceRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: UpdateWorkspaceRequest.$_createMessage)
    ..aOM<Workspace>(1, _omitFieldNames ? '' : 'workspace',
        subBuilder: Workspace.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateWorkspaceRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateWorkspaceRequest copyWith(
          void Function(UpdateWorkspaceRequest) updates) =>
      super.copyWith((message) => updates(message as UpdateWorkspaceRequest))
          as UpdateWorkspaceRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use UpdateWorkspaceRequest() / UpdateWorkspaceRequest.new instead')
  static UpdateWorkspaceRequest create() => UpdateWorkspaceRequest._();
  static $pb.GeneratedMessage $_createMessage() => UpdateWorkspaceRequest._();
  @$core.override
  UpdateWorkspaceRequest createEmptyInstance() => UpdateWorkspaceRequest._();
  @$core.pragma('dart2js:noInline')
  static UpdateWorkspaceRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<UpdateWorkspaceRequest>(
          UpdateWorkspaceRequest.$_createMessage);
  static UpdateWorkspaceRequest? _defaultInstance;

  @$pb.TagNumber(1)
  Workspace get workspace => $_getN(0);
  @$pb.TagNumber(1)
  set workspace(Workspace value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasWorkspace() => $_has(0);
  @$pb.TagNumber(1)
  void clearWorkspace() => $_clearField(1);
  @$pb.TagNumber(1)
  Workspace ensureWorkspace() => $_ensure(0);
}

class UpdateWorkspaceResponse extends $pb.GeneratedMessage {
  factory UpdateWorkspaceResponse({
    Workspace? workspace,
  }) {
    final result = UpdateWorkspaceResponse._();
    if (workspace != null) result.workspace = workspace;
    return result;
  }

  UpdateWorkspaceResponse._();

  factory UpdateWorkspaceResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpdateWorkspaceResponse()..mergeFromBuffer(data, registry);
  factory UpdateWorkspaceResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpdateWorkspaceResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'UpdateWorkspaceResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: UpdateWorkspaceResponse.$_createMessage)
    ..aOM<Workspace>(1, _omitFieldNames ? '' : 'workspace',
        subBuilder: Workspace.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateWorkspaceResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateWorkspaceResponse copyWith(
          void Function(UpdateWorkspaceResponse) updates) =>
      super.copyWith((message) => updates(message as UpdateWorkspaceResponse))
          as UpdateWorkspaceResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use UpdateWorkspaceResponse() / UpdateWorkspaceResponse.new instead')
  static UpdateWorkspaceResponse create() => UpdateWorkspaceResponse._();
  static $pb.GeneratedMessage $_createMessage() => UpdateWorkspaceResponse._();
  @$core.override
  UpdateWorkspaceResponse createEmptyInstance() => UpdateWorkspaceResponse._();
  @$core.pragma('dart2js:noInline')
  static UpdateWorkspaceResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<UpdateWorkspaceResponse>(
          UpdateWorkspaceResponse.$_createMessage);
  static UpdateWorkspaceResponse? _defaultInstance;

  @$pb.TagNumber(1)
  Workspace get workspace => $_getN(0);
  @$pb.TagNumber(1)
  set workspace(Workspace value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasWorkspace() => $_has(0);
  @$pb.TagNumber(1)
  void clearWorkspace() => $_clearField(1);
  @$pb.TagNumber(1)
  Workspace ensureWorkspace() => $_ensure(0);
}

class DeleteWorkspaceRequest extends $pb.GeneratedMessage {
  factory DeleteWorkspaceRequest({
    $core.String? id,
  }) {
    final result = DeleteWorkspaceRequest._();
    if (id != null) result.id = id;
    return result;
  }

  DeleteWorkspaceRequest._();

  factory DeleteWorkspaceRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DeleteWorkspaceRequest()..mergeFromBuffer(data, registry);
  factory DeleteWorkspaceRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DeleteWorkspaceRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'DeleteWorkspaceRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: DeleteWorkspaceRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteWorkspaceRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteWorkspaceRequest copyWith(
          void Function(DeleteWorkspaceRequest) updates) =>
      super.copyWith((message) => updates(message as DeleteWorkspaceRequest))
          as DeleteWorkspaceRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use DeleteWorkspaceRequest() / DeleteWorkspaceRequest.new instead')
  static DeleteWorkspaceRequest create() => DeleteWorkspaceRequest._();
  static $pb.GeneratedMessage $_createMessage() => DeleteWorkspaceRequest._();
  @$core.override
  DeleteWorkspaceRequest createEmptyInstance() => DeleteWorkspaceRequest._();
  @$core.pragma('dart2js:noInline')
  static DeleteWorkspaceRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<DeleteWorkspaceRequest>(
          DeleteWorkspaceRequest.$_createMessage);
  static DeleteWorkspaceRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);
}

class DeleteWorkspaceResponse extends $pb.GeneratedMessage {
  factory DeleteWorkspaceResponse() => DeleteWorkspaceResponse._();

  DeleteWorkspaceResponse._();

  factory DeleteWorkspaceResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DeleteWorkspaceResponse()..mergeFromBuffer(data, registry);
  factory DeleteWorkspaceResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DeleteWorkspaceResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'DeleteWorkspaceResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: DeleteWorkspaceResponse.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteWorkspaceResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteWorkspaceResponse copyWith(
          void Function(DeleteWorkspaceResponse) updates) =>
      super.copyWith((message) => updates(message as DeleteWorkspaceResponse))
          as DeleteWorkspaceResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use DeleteWorkspaceResponse() / DeleteWorkspaceResponse.new instead')
  static DeleteWorkspaceResponse create() => DeleteWorkspaceResponse._();
  static $pb.GeneratedMessage $_createMessage() => DeleteWorkspaceResponse._();
  @$core.override
  DeleteWorkspaceResponse createEmptyInstance() => DeleteWorkspaceResponse._();
  @$core.pragma('dart2js:noInline')
  static DeleteWorkspaceResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<DeleteWorkspaceResponse>(
          DeleteWorkspaceResponse.$_createMessage);
  static DeleteWorkspaceResponse? _defaultInstance;
}

class ListWorkspaceMembersRequest extends $pb.GeneratedMessage {
  factory ListWorkspaceMembersRequest({
    $core.String? workspaceId,
  }) {
    final result = ListWorkspaceMembersRequest._();
    if (workspaceId != null) result.workspaceId = workspaceId;
    return result;
  }

  ListWorkspaceMembersRequest._();

  factory ListWorkspaceMembersRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListWorkspaceMembersRequest()..mergeFromBuffer(data, registry);
  factory ListWorkspaceMembersRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListWorkspaceMembersRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListWorkspaceMembersRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: ListWorkspaceMembersRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'workspaceId')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListWorkspaceMembersRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListWorkspaceMembersRequest copyWith(
          void Function(ListWorkspaceMembersRequest) updates) =>
      super.copyWith(
              (message) => updates(message as ListWorkspaceMembersRequest))
          as ListWorkspaceMembersRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use ListWorkspaceMembersRequest() / ListWorkspaceMembersRequest.new instead')
  static ListWorkspaceMembersRequest create() =>
      ListWorkspaceMembersRequest._();
  static $pb.GeneratedMessage $_createMessage() =>
      ListWorkspaceMembersRequest._();
  @$core.override
  ListWorkspaceMembersRequest createEmptyInstance() =>
      ListWorkspaceMembersRequest._();
  @$core.pragma('dart2js:noInline')
  static ListWorkspaceMembersRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ListWorkspaceMembersRequest>(
          ListWorkspaceMembersRequest.$_createMessage);
  static ListWorkspaceMembersRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get workspaceId => $_getSZ(0);
  @$pb.TagNumber(1)
  set workspaceId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasWorkspaceId() => $_has(0);
  @$pb.TagNumber(1)
  void clearWorkspaceId() => $_clearField(1);
}

class ListWorkspaceMembersResponse extends $pb.GeneratedMessage {
  factory ListWorkspaceMembersResponse({
    $core.Iterable<WorkspaceMember>? members,
  }) {
    final result = ListWorkspaceMembersResponse._();
    if (members != null) result.members.addAll(members);
    return result;
  }

  ListWorkspaceMembersResponse._();

  factory ListWorkspaceMembersResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListWorkspaceMembersResponse()..mergeFromBuffer(data, registry);
  factory ListWorkspaceMembersResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListWorkspaceMembersResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListWorkspaceMembersResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: ListWorkspaceMembersResponse.$_createMessage)
    ..pPM<WorkspaceMember>(1, _omitFieldNames ? '' : 'members',
        subBuilder: WorkspaceMember.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListWorkspaceMembersResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListWorkspaceMembersResponse copyWith(
          void Function(ListWorkspaceMembersResponse) updates) =>
      super.copyWith(
              (message) => updates(message as ListWorkspaceMembersResponse))
          as ListWorkspaceMembersResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use ListWorkspaceMembersResponse() / ListWorkspaceMembersResponse.new instead')
  static ListWorkspaceMembersResponse create() =>
      ListWorkspaceMembersResponse._();
  static $pb.GeneratedMessage $_createMessage() =>
      ListWorkspaceMembersResponse._();
  @$core.override
  ListWorkspaceMembersResponse createEmptyInstance() =>
      ListWorkspaceMembersResponse._();
  @$core.pragma('dart2js:noInline')
  static ListWorkspaceMembersResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ListWorkspaceMembersResponse>(
          ListWorkspaceMembersResponse.$_createMessage);
  static ListWorkspaceMembersResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $pb.PbList<WorkspaceMember> get members => $_getList(0);
}

class AddWorkspaceMemberRequest extends $pb.GeneratedMessage {
  factory AddWorkspaceMemberRequest({
    $core.String? workspaceId,
    $core.String? userId,
    $core.String? role,
  }) {
    final result = AddWorkspaceMemberRequest._();
    if (workspaceId != null) result.workspaceId = workspaceId;
    if (userId != null) result.userId = userId;
    if (role != null) result.role = role;
    return result;
  }

  AddWorkspaceMemberRequest._();

  factory AddWorkspaceMemberRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      AddWorkspaceMemberRequest()..mergeFromBuffer(data, registry);
  factory AddWorkspaceMemberRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      AddWorkspaceMemberRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'AddWorkspaceMemberRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: AddWorkspaceMemberRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'workspaceId')
    ..aOS(2, _omitFieldNames ? '' : 'userId')
    ..aOS(3, _omitFieldNames ? '' : 'role')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  AddWorkspaceMemberRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  AddWorkspaceMemberRequest copyWith(
          void Function(AddWorkspaceMemberRequest) updates) =>
      super.copyWith((message) => updates(message as AddWorkspaceMemberRequest))
          as AddWorkspaceMemberRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use AddWorkspaceMemberRequest() / AddWorkspaceMemberRequest.new instead')
  static AddWorkspaceMemberRequest create() => AddWorkspaceMemberRequest._();
  static $pb.GeneratedMessage $_createMessage() =>
      AddWorkspaceMemberRequest._();
  @$core.override
  AddWorkspaceMemberRequest createEmptyInstance() =>
      AddWorkspaceMemberRequest._();
  @$core.pragma('dart2js:noInline')
  static AddWorkspaceMemberRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<AddWorkspaceMemberRequest>(
          AddWorkspaceMemberRequest.$_createMessage);
  static AddWorkspaceMemberRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get workspaceId => $_getSZ(0);
  @$pb.TagNumber(1)
  set workspaceId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasWorkspaceId() => $_has(0);
  @$pb.TagNumber(1)
  void clearWorkspaceId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get userId => $_getSZ(1);
  @$pb.TagNumber(2)
  set userId($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasUserId() => $_has(1);
  @$pb.TagNumber(2)
  void clearUserId() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get role => $_getSZ(2);
  @$pb.TagNumber(3)
  set role($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasRole() => $_has(2);
  @$pb.TagNumber(3)
  void clearRole() => $_clearField(3);
}

class AddWorkspaceMemberResponse extends $pb.GeneratedMessage {
  factory AddWorkspaceMemberResponse({
    WorkspaceMember? member,
  }) {
    final result = AddWorkspaceMemberResponse._();
    if (member != null) result.member = member;
    return result;
  }

  AddWorkspaceMemberResponse._();

  factory AddWorkspaceMemberResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      AddWorkspaceMemberResponse()..mergeFromBuffer(data, registry);
  factory AddWorkspaceMemberResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      AddWorkspaceMemberResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'AddWorkspaceMemberResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: AddWorkspaceMemberResponse.$_createMessage)
    ..aOM<WorkspaceMember>(1, _omitFieldNames ? '' : 'member',
        subBuilder: WorkspaceMember.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  AddWorkspaceMemberResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  AddWorkspaceMemberResponse copyWith(
          void Function(AddWorkspaceMemberResponse) updates) =>
      super.copyWith(
              (message) => updates(message as AddWorkspaceMemberResponse))
          as AddWorkspaceMemberResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use AddWorkspaceMemberResponse() / AddWorkspaceMemberResponse.new instead')
  static AddWorkspaceMemberResponse create() => AddWorkspaceMemberResponse._();
  static $pb.GeneratedMessage $_createMessage() =>
      AddWorkspaceMemberResponse._();
  @$core.override
  AddWorkspaceMemberResponse createEmptyInstance() =>
      AddWorkspaceMemberResponse._();
  @$core.pragma('dart2js:noInline')
  static AddWorkspaceMemberResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<AddWorkspaceMemberResponse>(
          AddWorkspaceMemberResponse.$_createMessage);
  static AddWorkspaceMemberResponse? _defaultInstance;

  @$pb.TagNumber(1)
  WorkspaceMember get member => $_getN(0);
  @$pb.TagNumber(1)
  set member(WorkspaceMember value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasMember() => $_has(0);
  @$pb.TagNumber(1)
  void clearMember() => $_clearField(1);
  @$pb.TagNumber(1)
  WorkspaceMember ensureMember() => $_ensure(0);
}

class UpdateWorkspaceMemberRequest extends $pb.GeneratedMessage {
  factory UpdateWorkspaceMemberRequest({
    $core.String? workspaceId,
    $core.String? userId,
    $core.String? role,
  }) {
    final result = UpdateWorkspaceMemberRequest._();
    if (workspaceId != null) result.workspaceId = workspaceId;
    if (userId != null) result.userId = userId;
    if (role != null) result.role = role;
    return result;
  }

  UpdateWorkspaceMemberRequest._();

  factory UpdateWorkspaceMemberRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpdateWorkspaceMemberRequest()..mergeFromBuffer(data, registry);
  factory UpdateWorkspaceMemberRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpdateWorkspaceMemberRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'UpdateWorkspaceMemberRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: UpdateWorkspaceMemberRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'workspaceId')
    ..aOS(2, _omitFieldNames ? '' : 'userId')
    ..aOS(3, _omitFieldNames ? '' : 'role')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateWorkspaceMemberRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateWorkspaceMemberRequest copyWith(
          void Function(UpdateWorkspaceMemberRequest) updates) =>
      super.copyWith(
              (message) => updates(message as UpdateWorkspaceMemberRequest))
          as UpdateWorkspaceMemberRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use UpdateWorkspaceMemberRequest() / UpdateWorkspaceMemberRequest.new instead')
  static UpdateWorkspaceMemberRequest create() =>
      UpdateWorkspaceMemberRequest._();
  static $pb.GeneratedMessage $_createMessage() =>
      UpdateWorkspaceMemberRequest._();
  @$core.override
  UpdateWorkspaceMemberRequest createEmptyInstance() =>
      UpdateWorkspaceMemberRequest._();
  @$core.pragma('dart2js:noInline')
  static UpdateWorkspaceMemberRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<UpdateWorkspaceMemberRequest>(
          UpdateWorkspaceMemberRequest.$_createMessage);
  static UpdateWorkspaceMemberRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get workspaceId => $_getSZ(0);
  @$pb.TagNumber(1)
  set workspaceId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasWorkspaceId() => $_has(0);
  @$pb.TagNumber(1)
  void clearWorkspaceId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get userId => $_getSZ(1);
  @$pb.TagNumber(2)
  set userId($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasUserId() => $_has(1);
  @$pb.TagNumber(2)
  void clearUserId() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get role => $_getSZ(2);
  @$pb.TagNumber(3)
  set role($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasRole() => $_has(2);
  @$pb.TagNumber(3)
  void clearRole() => $_clearField(3);
}

class UpdateWorkspaceMemberResponse extends $pb.GeneratedMessage {
  factory UpdateWorkspaceMemberResponse({
    WorkspaceMember? member,
  }) {
    final result = UpdateWorkspaceMemberResponse._();
    if (member != null) result.member = member;
    return result;
  }

  UpdateWorkspaceMemberResponse._();

  factory UpdateWorkspaceMemberResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpdateWorkspaceMemberResponse()..mergeFromBuffer(data, registry);
  factory UpdateWorkspaceMemberResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpdateWorkspaceMemberResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'UpdateWorkspaceMemberResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: UpdateWorkspaceMemberResponse.$_createMessage)
    ..aOM<WorkspaceMember>(1, _omitFieldNames ? '' : 'member',
        subBuilder: WorkspaceMember.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateWorkspaceMemberResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateWorkspaceMemberResponse copyWith(
          void Function(UpdateWorkspaceMemberResponse) updates) =>
      super.copyWith(
              (message) => updates(message as UpdateWorkspaceMemberResponse))
          as UpdateWorkspaceMemberResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use UpdateWorkspaceMemberResponse() / UpdateWorkspaceMemberResponse.new instead')
  static UpdateWorkspaceMemberResponse create() =>
      UpdateWorkspaceMemberResponse._();
  static $pb.GeneratedMessage $_createMessage() =>
      UpdateWorkspaceMemberResponse._();
  @$core.override
  UpdateWorkspaceMemberResponse createEmptyInstance() =>
      UpdateWorkspaceMemberResponse._();
  @$core.pragma('dart2js:noInline')
  static UpdateWorkspaceMemberResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<UpdateWorkspaceMemberResponse>(
          UpdateWorkspaceMemberResponse.$_createMessage);
  static UpdateWorkspaceMemberResponse? _defaultInstance;

  @$pb.TagNumber(1)
  WorkspaceMember get member => $_getN(0);
  @$pb.TagNumber(1)
  set member(WorkspaceMember value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasMember() => $_has(0);
  @$pb.TagNumber(1)
  void clearMember() => $_clearField(1);
  @$pb.TagNumber(1)
  WorkspaceMember ensureMember() => $_ensure(0);
}

class RemoveWorkspaceMemberRequest extends $pb.GeneratedMessage {
  factory RemoveWorkspaceMemberRequest({
    $core.String? workspaceId,
    $core.String? userId,
  }) {
    final result = RemoveWorkspaceMemberRequest._();
    if (workspaceId != null) result.workspaceId = workspaceId;
    if (userId != null) result.userId = userId;
    return result;
  }

  RemoveWorkspaceMemberRequest._();

  factory RemoveWorkspaceMemberRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      RemoveWorkspaceMemberRequest()..mergeFromBuffer(data, registry);
  factory RemoveWorkspaceMemberRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      RemoveWorkspaceMemberRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'RemoveWorkspaceMemberRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: RemoveWorkspaceMemberRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'workspaceId')
    ..aOS(2, _omitFieldNames ? '' : 'userId')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  RemoveWorkspaceMemberRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  RemoveWorkspaceMemberRequest copyWith(
          void Function(RemoveWorkspaceMemberRequest) updates) =>
      super.copyWith(
              (message) => updates(message as RemoveWorkspaceMemberRequest))
          as RemoveWorkspaceMemberRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use RemoveWorkspaceMemberRequest() / RemoveWorkspaceMemberRequest.new instead')
  static RemoveWorkspaceMemberRequest create() =>
      RemoveWorkspaceMemberRequest._();
  static $pb.GeneratedMessage $_createMessage() =>
      RemoveWorkspaceMemberRequest._();
  @$core.override
  RemoveWorkspaceMemberRequest createEmptyInstance() =>
      RemoveWorkspaceMemberRequest._();
  @$core.pragma('dart2js:noInline')
  static RemoveWorkspaceMemberRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<RemoveWorkspaceMemberRequest>(
          RemoveWorkspaceMemberRequest.$_createMessage);
  static RemoveWorkspaceMemberRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get workspaceId => $_getSZ(0);
  @$pb.TagNumber(1)
  set workspaceId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasWorkspaceId() => $_has(0);
  @$pb.TagNumber(1)
  void clearWorkspaceId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get userId => $_getSZ(1);
  @$pb.TagNumber(2)
  set userId($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasUserId() => $_has(1);
  @$pb.TagNumber(2)
  void clearUserId() => $_clearField(2);
}

class RemoveWorkspaceMemberResponse extends $pb.GeneratedMessage {
  factory RemoveWorkspaceMemberResponse() => RemoveWorkspaceMemberResponse._();

  RemoveWorkspaceMemberResponse._();

  factory RemoveWorkspaceMemberResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      RemoveWorkspaceMemberResponse()..mergeFromBuffer(data, registry);
  factory RemoveWorkspaceMemberResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      RemoveWorkspaceMemberResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'RemoveWorkspaceMemberResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: RemoveWorkspaceMemberResponse.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  RemoveWorkspaceMemberResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  RemoveWorkspaceMemberResponse copyWith(
          void Function(RemoveWorkspaceMemberResponse) updates) =>
      super.copyWith(
              (message) => updates(message as RemoveWorkspaceMemberResponse))
          as RemoveWorkspaceMemberResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use RemoveWorkspaceMemberResponse() / RemoveWorkspaceMemberResponse.new instead')
  static RemoveWorkspaceMemberResponse create() =>
      RemoveWorkspaceMemberResponse._();
  static $pb.GeneratedMessage $_createMessage() =>
      RemoveWorkspaceMemberResponse._();
  @$core.override
  RemoveWorkspaceMemberResponse createEmptyInstance() =>
      RemoveWorkspaceMemberResponse._();
  @$core.pragma('dart2js:noInline')
  static RemoveWorkspaceMemberResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<RemoveWorkspaceMemberResponse>(
          RemoveWorkspaceMemberResponse.$_createMessage);
  static RemoveWorkspaceMemberResponse? _defaultInstance;
}

/// WorkspaceService manages workspaces and memberships.
class WorkspaceServiceApi {
  final $pb.RpcClient _client;

  WorkspaceServiceApi(this._client);

  /// ListWorkspaces returns all workspaces the caller can access. Global
  /// admins receive every workspace; other users receive only the ones they
  /// are members of.
  $async.Future<ListWorkspacesResponse> listWorkspaces(
          $pb.ClientContext? ctx, ListWorkspacesRequest request) =>
      _client.invoke<ListWorkspacesResponse>(ctx, 'WorkspaceService',
          'ListWorkspaces', request, ListWorkspacesResponse());

  /// GetWorkspace returns a single workspace by id.
  $async.Future<GetWorkspaceResponse> getWorkspace(
          $pb.ClientContext? ctx, GetWorkspaceRequest request) =>
      _client.invoke<GetWorkspaceResponse>(ctx, 'WorkspaceService',
          'GetWorkspace', request, GetWorkspaceResponse());

  /// CreateWorkspace creates a new workspace. The caller is added as the
  /// initial owner.
  $async.Future<CreateWorkspaceResponse> createWorkspace(
          $pb.ClientContext? ctx, CreateWorkspaceRequest request) =>
      _client.invoke<CreateWorkspaceResponse>(ctx, 'WorkspaceService',
          'CreateWorkspace', request, CreateWorkspaceResponse());

  /// UpdateWorkspace updates workspace metadata.
  $async.Future<UpdateWorkspaceResponse> updateWorkspace(
          $pb.ClientContext? ctx, UpdateWorkspaceRequest request) =>
      _client.invoke<UpdateWorkspaceResponse>(ctx, 'WorkspaceService',
          'UpdateWorkspace', request, UpdateWorkspaceResponse());

  /// DeleteWorkspace permanently removes a workspace and all of its scoped
  /// configs. Caller must be an owner.
  $async.Future<DeleteWorkspaceResponse> deleteWorkspace(
          $pb.ClientContext? ctx, DeleteWorkspaceRequest request) =>
      _client.invoke<DeleteWorkspaceResponse>(ctx, 'WorkspaceService',
          'DeleteWorkspace', request, DeleteWorkspaceResponse());

  /// ListWorkspaceMembers returns the membership list of a workspace.
  $async.Future<ListWorkspaceMembersResponse> listWorkspaceMembers(
          $pb.ClientContext? ctx, ListWorkspaceMembersRequest request) =>
      _client.invoke<ListWorkspaceMembersResponse>(ctx, 'WorkspaceService',
          'ListWorkspaceMembers', request, ListWorkspaceMembersResponse());

  /// AddWorkspaceMember adds a user to a workspace with a given role.
  $async.Future<AddWorkspaceMemberResponse> addWorkspaceMember(
          $pb.ClientContext? ctx, AddWorkspaceMemberRequest request) =>
      _client.invoke<AddWorkspaceMemberResponse>(ctx, 'WorkspaceService',
          'AddWorkspaceMember', request, AddWorkspaceMemberResponse());

  /// UpdateWorkspaceMember changes a member's role.
  $async.Future<UpdateWorkspaceMemberResponse> updateWorkspaceMember(
          $pb.ClientContext? ctx, UpdateWorkspaceMemberRequest request) =>
      _client.invoke<UpdateWorkspaceMemberResponse>(ctx, 'WorkspaceService',
          'UpdateWorkspaceMember', request, UpdateWorkspaceMemberResponse());

  /// RemoveWorkspaceMember removes a user from a workspace.
  $async.Future<RemoveWorkspaceMemberResponse> removeWorkspaceMember(
          $pb.ClientContext? ctx, RemoveWorkspaceMemberRequest request) =>
      _client.invoke<RemoveWorkspaceMemberResponse>(ctx, 'WorkspaceService',
          'RemoveWorkspaceMember', request, RemoveWorkspaceMemberResponse());
}

const $core.bool _omitFieldNames =
    $core.bool.fromEnvironment('protobuf.omit_field_names');
const $core.bool _omitMessageNames =
    $core.bool.fromEnvironment('protobuf.omit_message_names');
