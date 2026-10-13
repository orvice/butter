// This is a generated file - do not edit.
//
// Generated from agents/v1/agent_file.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names, prefer_relative_imports

import 'dart:async' as $async;
import 'dart:core' as $core;

import 'package:fixnum/fixnum.dart' as $fixnum;
import 'package:protobuf/protobuf.dart' as $pb;
import 'package:protobuf/well_known_types/google/protobuf/timestamp.pb.dart'
    as $0;

import 'agent_file.pbenum.dart';

export 'package:protobuf/protobuf.dart' show GeneratedMessageGenericExtensions;

export 'agent_file.pbenum.dart';

/// AgentFileSpace is a workspace-owned collection of files that agents can mount.
class AgentFileSpace extends $pb.GeneratedMessage {
  factory AgentFileSpace({
    $core.String? id,
    $core.String? name,
    $core.String? description,
    $core.Iterable<$core.MapEntry<$core.String, $core.String>>? metadata,
    $0.Timestamp? createdAt,
    $0.Timestamp? updatedAt,
    $core.String? workspaceId,
  }) {
    final result = AgentFileSpace._();
    if (id != null) result.id = id;
    if (name != null) result.name = name;
    if (description != null) result.description = description;
    if (metadata != null) result.metadata.addEntries(metadata);
    if (createdAt != null) result.createdAt = createdAt;
    if (updatedAt != null) result.updatedAt = updatedAt;
    if (workspaceId != null) result.workspaceId = workspaceId;
    return result;
  }

  AgentFileSpace._();

  factory AgentFileSpace.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      AgentFileSpace()..mergeFromBuffer(data, registry);
  factory AgentFileSpace.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      AgentFileSpace()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'AgentFileSpace',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: AgentFileSpace.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..aOS(2, _omitFieldNames ? '' : 'name')
    ..aOS(3, _omitFieldNames ? '' : 'description')
    ..m<$core.String, $core.String>(4, _omitFieldNames ? '' : 'metadata',
        entryClassName: 'AgentFileSpace.MetadataEntry',
        keyFieldType: $pb.PbFieldType.OS,
        valueFieldType: $pb.PbFieldType.OS,
        packageName: const $pb.PackageName('agents.v1'))
    ..aOM<$0.Timestamp>(5, _omitFieldNames ? '' : 'createdAt',
        subBuilder: $0.Timestamp.$_createMessage)
    ..aOM<$0.Timestamp>(6, _omitFieldNames ? '' : 'updatedAt',
        subBuilder: $0.Timestamp.$_createMessage)
    ..aOS(100, _omitFieldNames ? '' : 'workspaceId')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  AgentFileSpace clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  AgentFileSpace copyWith(void Function(AgentFileSpace) updates) =>
      super.copyWith((message) => updates(message as AgentFileSpace))
          as AgentFileSpace;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use AgentFileSpace() / AgentFileSpace.new instead')
  static AgentFileSpace create() => AgentFileSpace._();
  static $pb.GeneratedMessage $_createMessage() => AgentFileSpace._();
  @$core.override
  AgentFileSpace createEmptyInstance() => AgentFileSpace._();
  @$core.pragma('dart2js:noInline')
  static AgentFileSpace getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<AgentFileSpace>(
          AgentFileSpace.$_createMessage);
  static AgentFileSpace? _defaultInstance;

  /// @gotags: json:"id,omitempty" yaml:"id,omitempty" bson:"_id,omitempty"
  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);

  /// @gotags: json:"name,omitempty" yaml:"name,omitempty" bson:"name,omitempty"
  @$pb.TagNumber(2)
  $core.String get name => $_getSZ(1);
  @$pb.TagNumber(2)
  set name($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasName() => $_has(1);
  @$pb.TagNumber(2)
  void clearName() => $_clearField(2);

  /// @gotags: json:"description,omitempty" yaml:"description,omitempty" bson:"description,omitempty"
  @$pb.TagNumber(3)
  $core.String get description => $_getSZ(2);
  @$pb.TagNumber(3)
  set description($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasDescription() => $_has(2);
  @$pb.TagNumber(3)
  void clearDescription() => $_clearField(3);

  /// @gotags: json:"metadata,omitempty" yaml:"metadata,omitempty" bson:"metadata,omitempty"
  @$pb.TagNumber(4)
  $pb.PbMap<$core.String, $core.String> get metadata => $_getMap(3);

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

  /// @gotags: json:"workspace_id,omitempty" yaml:"workspace_id,omitempty" bson:"workspace_id,omitempty"
  @$pb.TagNumber(100)
  $core.String get workspaceId => $_getSZ(6);
  @$pb.TagNumber(100)
  set workspaceId($core.String value) => $_setString(6, value);
  @$pb.TagNumber(100)
  $core.bool hasWorkspaceId() => $_has(6);
  @$pb.TagNumber(100)
  void clearWorkspaceId() => $_clearField(100);
}

/// AgentFile stores metadata for a text file inside an AgentFileSpace. File
/// content is stored separately by the repository backend.
class AgentFile extends $pb.GeneratedMessage {
  factory AgentFile({
    $core.String? id,
    $core.String? spaceId,
    $core.String? path,
    $core.String? contentType,
    $fixnum.Int64? sizeBytes,
    $fixnum.Int64? version,
    $core.Iterable<$core.MapEntry<$core.String, $core.String>>? metadata,
    $0.Timestamp? createdAt,
    $0.Timestamp? updatedAt,
    $core.String? workspaceId,
  }) {
    final result = AgentFile._();
    if (id != null) result.id = id;
    if (spaceId != null) result.spaceId = spaceId;
    if (path != null) result.path = path;
    if (contentType != null) result.contentType = contentType;
    if (sizeBytes != null) result.sizeBytes = sizeBytes;
    if (version != null) result.version = version;
    if (metadata != null) result.metadata.addEntries(metadata);
    if (createdAt != null) result.createdAt = createdAt;
    if (updatedAt != null) result.updatedAt = updatedAt;
    if (workspaceId != null) result.workspaceId = workspaceId;
    return result;
  }

  AgentFile._();

  factory AgentFile.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      AgentFile()..mergeFromBuffer(data, registry);
  factory AgentFile.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      AgentFile()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'AgentFile',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: AgentFile.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..aOS(2, _omitFieldNames ? '' : 'spaceId')
    ..aOS(3, _omitFieldNames ? '' : 'path')
    ..aOS(4, _omitFieldNames ? '' : 'contentType')
    ..aInt64(5, _omitFieldNames ? '' : 'sizeBytes')
    ..aInt64(6, _omitFieldNames ? '' : 'version')
    ..m<$core.String, $core.String>(7, _omitFieldNames ? '' : 'metadata',
        entryClassName: 'AgentFile.MetadataEntry',
        keyFieldType: $pb.PbFieldType.OS,
        valueFieldType: $pb.PbFieldType.OS,
        packageName: const $pb.PackageName('agents.v1'))
    ..aOM<$0.Timestamp>(8, _omitFieldNames ? '' : 'createdAt',
        subBuilder: $0.Timestamp.$_createMessage)
    ..aOM<$0.Timestamp>(9, _omitFieldNames ? '' : 'updatedAt',
        subBuilder: $0.Timestamp.$_createMessage)
    ..aOS(100, _omitFieldNames ? '' : 'workspaceId')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  AgentFile clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  AgentFile copyWith(void Function(AgentFile) updates) =>
      super.copyWith((message) => updates(message as AgentFile)) as AgentFile;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use AgentFile() / AgentFile.new instead')
  static AgentFile create() => AgentFile._();
  static $pb.GeneratedMessage $_createMessage() => AgentFile._();
  @$core.override
  AgentFile createEmptyInstance() => AgentFile._();
  @$core.pragma('dart2js:noInline')
  static AgentFile getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<AgentFile>(AgentFile.$_createMessage);
  static AgentFile? _defaultInstance;

  /// @gotags: json:"id,omitempty" yaml:"id,omitempty" bson:"_id,omitempty"
  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);

  /// @gotags: json:"space_id,omitempty" yaml:"space_id,omitempty" bson:"space_id,omitempty"
  @$pb.TagNumber(2)
  $core.String get spaceId => $_getSZ(1);
  @$pb.TagNumber(2)
  set spaceId($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasSpaceId() => $_has(1);
  @$pb.TagNumber(2)
  void clearSpaceId() => $_clearField(2);

  /// Absolute path within the file space, e.g. "/notes/todo.md".
  /// @gotags: json:"path,omitempty" yaml:"path,omitempty" bson:"path,omitempty"
  @$pb.TagNumber(3)
  $core.String get path => $_getSZ(2);
  @$pb.TagNumber(3)
  set path($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasPath() => $_has(2);
  @$pb.TagNumber(3)
  void clearPath() => $_clearField(3);

  /// @gotags: json:"content_type,omitempty" yaml:"content_type,omitempty" bson:"content_type,omitempty"
  @$pb.TagNumber(4)
  $core.String get contentType => $_getSZ(3);
  @$pb.TagNumber(4)
  set contentType($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasContentType() => $_has(3);
  @$pb.TagNumber(4)
  void clearContentType() => $_clearField(4);

  /// @gotags: json:"size_bytes,omitempty" yaml:"size_bytes,omitempty" bson:"size_bytes,omitempty"
  @$pb.TagNumber(5)
  $fixnum.Int64 get sizeBytes => $_getI64(4);
  @$pb.TagNumber(5)
  set sizeBytes($fixnum.Int64 value) => $_setInt64(4, value);
  @$pb.TagNumber(5)
  $core.bool hasSizeBytes() => $_has(4);
  @$pb.TagNumber(5)
  void clearSizeBytes() => $_clearField(5);

  /// Monotonic content version. Each write creates the next version.
  /// @gotags: json:"version,omitempty" yaml:"version,omitempty" bson:"version,omitempty"
  @$pb.TagNumber(6)
  $fixnum.Int64 get version => $_getI64(5);
  @$pb.TagNumber(6)
  set version($fixnum.Int64 value) => $_setInt64(5, value);
  @$pb.TagNumber(6)
  $core.bool hasVersion() => $_has(5);
  @$pb.TagNumber(6)
  void clearVersion() => $_clearField(6);

  /// @gotags: json:"metadata,omitempty" yaml:"metadata,omitempty" bson:"metadata,omitempty"
  @$pb.TagNumber(7)
  $pb.PbMap<$core.String, $core.String> get metadata => $_getMap(6);

  /// @gotags: json:"created_at,omitempty" yaml:"created_at,omitempty" bson:"created_at,omitempty"
  @$pb.TagNumber(8)
  $0.Timestamp get createdAt => $_getN(7);
  @$pb.TagNumber(8)
  set createdAt($0.Timestamp value) => $_setField(8, value);
  @$pb.TagNumber(8)
  $core.bool hasCreatedAt() => $_has(7);
  @$pb.TagNumber(8)
  void clearCreatedAt() => $_clearField(8);
  @$pb.TagNumber(8)
  $0.Timestamp ensureCreatedAt() => $_ensure(7);

  /// @gotags: json:"updated_at,omitempty" yaml:"updated_at,omitempty" bson:"updated_at,omitempty"
  @$pb.TagNumber(9)
  $0.Timestamp get updatedAt => $_getN(8);
  @$pb.TagNumber(9)
  set updatedAt($0.Timestamp value) => $_setField(9, value);
  @$pb.TagNumber(9)
  $core.bool hasUpdatedAt() => $_has(8);
  @$pb.TagNumber(9)
  void clearUpdatedAt() => $_clearField(9);
  @$pb.TagNumber(9)
  $0.Timestamp ensureUpdatedAt() => $_ensure(8);

  /// @gotags: json:"workspace_id,omitempty" yaml:"workspace_id,omitempty" bson:"workspace_id,omitempty"
  @$pb.TagNumber(100)
  $core.String get workspaceId => $_getSZ(9);
  @$pb.TagNumber(100)
  set workspaceId($core.String value) => $_setString(9, value);
  @$pb.TagNumber(100)
  $core.bool hasWorkspaceId() => $_has(9);
  @$pb.TagNumber(100)
  void clearWorkspaceId() => $_clearField(100);
}

/// AgentFileMount grants an agent access to a file space under a virtual path.
class AgentFileMount extends $pb.GeneratedMessage {
  factory AgentFileMount({
    $core.String? spaceId,
    $core.String? mountPath,
    AgentFileMountPermission? permission,
  }) {
    final result = AgentFileMount._();
    if (spaceId != null) result.spaceId = spaceId;
    if (mountPath != null) result.mountPath = mountPath;
    if (permission != null) result.permission = permission;
    return result;
  }

  AgentFileMount._();

  factory AgentFileMount.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      AgentFileMount()..mergeFromBuffer(data, registry);
  factory AgentFileMount.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      AgentFileMount()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'AgentFileMount',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: AgentFileMount.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'spaceId')
    ..aOS(2, _omitFieldNames ? '' : 'mountPath')
    ..aE<AgentFileMountPermission>(3, _omitFieldNames ? '' : 'permission',
        enumValues: AgentFileMountPermission.values)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  AgentFileMount clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  AgentFileMount copyWith(void Function(AgentFileMount) updates) =>
      super.copyWith((message) => updates(message as AgentFileMount))
          as AgentFileMount;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use AgentFileMount() / AgentFileMount.new instead')
  static AgentFileMount create() => AgentFileMount._();
  static $pb.GeneratedMessage $_createMessage() => AgentFileMount._();
  @$core.override
  AgentFileMount createEmptyInstance() => AgentFileMount._();
  @$core.pragma('dart2js:noInline')
  static AgentFileMount getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<AgentFileMount>(
          AgentFileMount.$_createMessage);
  static AgentFileMount? _defaultInstance;

  /// @gotags: json:"space_id,omitempty" yaml:"space_id,omitempty"
  @$pb.TagNumber(1)
  $core.String get spaceId => $_getSZ(0);
  @$pb.TagNumber(1)
  set spaceId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasSpaceId() => $_has(0);
  @$pb.TagNumber(1)
  void clearSpaceId() => $_clearField(1);

  /// Virtual root exposed to the agent, e.g. "/docs".
  /// @gotags: json:"mount_path,omitempty" yaml:"mount_path,omitempty"
  @$pb.TagNumber(2)
  $core.String get mountPath => $_getSZ(1);
  @$pb.TagNumber(2)
  set mountPath($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasMountPath() => $_has(1);
  @$pb.TagNumber(2)
  void clearMountPath() => $_clearField(2);

  /// @gotags: json:"permission,omitempty" yaml:"permission,omitempty"
  @$pb.TagNumber(3)
  AgentFileMountPermission get permission => $_getN(2);
  @$pb.TagNumber(3)
  set permission(AgentFileMountPermission value) => $_setField(3, value);
  @$pb.TagNumber(3)
  $core.bool hasPermission() => $_has(2);
  @$pb.TagNumber(3)
  void clearPermission() => $_clearField(3);
}

class ListAgentFileSpacesRequest extends $pb.GeneratedMessage {
  factory ListAgentFileSpacesRequest() => ListAgentFileSpacesRequest._();

  ListAgentFileSpacesRequest._();

  factory ListAgentFileSpacesRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListAgentFileSpacesRequest()..mergeFromBuffer(data, registry);
  factory ListAgentFileSpacesRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListAgentFileSpacesRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListAgentFileSpacesRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: ListAgentFileSpacesRequest.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListAgentFileSpacesRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListAgentFileSpacesRequest copyWith(
          void Function(ListAgentFileSpacesRequest) updates) =>
      super.copyWith(
              (message) => updates(message as ListAgentFileSpacesRequest))
          as ListAgentFileSpacesRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use ListAgentFileSpacesRequest() / ListAgentFileSpacesRequest.new instead')
  static ListAgentFileSpacesRequest create() => ListAgentFileSpacesRequest._();
  static $pb.GeneratedMessage $_createMessage() =>
      ListAgentFileSpacesRequest._();
  @$core.override
  ListAgentFileSpacesRequest createEmptyInstance() =>
      ListAgentFileSpacesRequest._();
  @$core.pragma('dart2js:noInline')
  static ListAgentFileSpacesRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ListAgentFileSpacesRequest>(
          ListAgentFileSpacesRequest.$_createMessage);
  static ListAgentFileSpacesRequest? _defaultInstance;
}

class ListAgentFileSpacesResponse extends $pb.GeneratedMessage {
  factory ListAgentFileSpacesResponse({
    $core.Iterable<AgentFileSpace>? spaces,
  }) {
    final result = ListAgentFileSpacesResponse._();
    if (spaces != null) result.spaces.addAll(spaces);
    return result;
  }

  ListAgentFileSpacesResponse._();

  factory ListAgentFileSpacesResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListAgentFileSpacesResponse()..mergeFromBuffer(data, registry);
  factory ListAgentFileSpacesResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListAgentFileSpacesResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListAgentFileSpacesResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: ListAgentFileSpacesResponse.$_createMessage)
    ..pPM<AgentFileSpace>(1, _omitFieldNames ? '' : 'spaces',
        subBuilder: AgentFileSpace.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListAgentFileSpacesResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListAgentFileSpacesResponse copyWith(
          void Function(ListAgentFileSpacesResponse) updates) =>
      super.copyWith(
              (message) => updates(message as ListAgentFileSpacesResponse))
          as ListAgentFileSpacesResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use ListAgentFileSpacesResponse() / ListAgentFileSpacesResponse.new instead')
  static ListAgentFileSpacesResponse create() =>
      ListAgentFileSpacesResponse._();
  static $pb.GeneratedMessage $_createMessage() =>
      ListAgentFileSpacesResponse._();
  @$core.override
  ListAgentFileSpacesResponse createEmptyInstance() =>
      ListAgentFileSpacesResponse._();
  @$core.pragma('dart2js:noInline')
  static ListAgentFileSpacesResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ListAgentFileSpacesResponse>(
          ListAgentFileSpacesResponse.$_createMessage);
  static ListAgentFileSpacesResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $pb.PbList<AgentFileSpace> get spaces => $_getList(0);
}

class GetAgentFileSpaceRequest extends $pb.GeneratedMessage {
  factory GetAgentFileSpaceRequest({
    $core.String? id,
  }) {
    final result = GetAgentFileSpaceRequest._();
    if (id != null) result.id = id;
    return result;
  }

  GetAgentFileSpaceRequest._();

  factory GetAgentFileSpaceRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetAgentFileSpaceRequest()..mergeFromBuffer(data, registry);
  factory GetAgentFileSpaceRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetAgentFileSpaceRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetAgentFileSpaceRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: GetAgentFileSpaceRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetAgentFileSpaceRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetAgentFileSpaceRequest copyWith(
          void Function(GetAgentFileSpaceRequest) updates) =>
      super.copyWith((message) => updates(message as GetAgentFileSpaceRequest))
          as GetAgentFileSpaceRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use GetAgentFileSpaceRequest() / GetAgentFileSpaceRequest.new instead')
  static GetAgentFileSpaceRequest create() => GetAgentFileSpaceRequest._();
  static $pb.GeneratedMessage $_createMessage() => GetAgentFileSpaceRequest._();
  @$core.override
  GetAgentFileSpaceRequest createEmptyInstance() =>
      GetAgentFileSpaceRequest._();
  @$core.pragma('dart2js:noInline')
  static GetAgentFileSpaceRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<GetAgentFileSpaceRequest>(
          GetAgentFileSpaceRequest.$_createMessage);
  static GetAgentFileSpaceRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);
}

class GetAgentFileSpaceResponse extends $pb.GeneratedMessage {
  factory GetAgentFileSpaceResponse({
    AgentFileSpace? space,
  }) {
    final result = GetAgentFileSpaceResponse._();
    if (space != null) result.space = space;
    return result;
  }

  GetAgentFileSpaceResponse._();

  factory GetAgentFileSpaceResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetAgentFileSpaceResponse()..mergeFromBuffer(data, registry);
  factory GetAgentFileSpaceResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetAgentFileSpaceResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetAgentFileSpaceResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: GetAgentFileSpaceResponse.$_createMessage)
    ..aOM<AgentFileSpace>(1, _omitFieldNames ? '' : 'space',
        subBuilder: AgentFileSpace.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetAgentFileSpaceResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetAgentFileSpaceResponse copyWith(
          void Function(GetAgentFileSpaceResponse) updates) =>
      super.copyWith((message) => updates(message as GetAgentFileSpaceResponse))
          as GetAgentFileSpaceResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use GetAgentFileSpaceResponse() / GetAgentFileSpaceResponse.new instead')
  static GetAgentFileSpaceResponse create() => GetAgentFileSpaceResponse._();
  static $pb.GeneratedMessage $_createMessage() =>
      GetAgentFileSpaceResponse._();
  @$core.override
  GetAgentFileSpaceResponse createEmptyInstance() =>
      GetAgentFileSpaceResponse._();
  @$core.pragma('dart2js:noInline')
  static GetAgentFileSpaceResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<GetAgentFileSpaceResponse>(
          GetAgentFileSpaceResponse.$_createMessage);
  static GetAgentFileSpaceResponse? _defaultInstance;

  @$pb.TagNumber(1)
  AgentFileSpace get space => $_getN(0);
  @$pb.TagNumber(1)
  set space(AgentFileSpace value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasSpace() => $_has(0);
  @$pb.TagNumber(1)
  void clearSpace() => $_clearField(1);
  @$pb.TagNumber(1)
  AgentFileSpace ensureSpace() => $_ensure(0);
}

class CreateAgentFileSpaceRequest extends $pb.GeneratedMessage {
  factory CreateAgentFileSpaceRequest({
    AgentFileSpace? space,
  }) {
    final result = CreateAgentFileSpaceRequest._();
    if (space != null) result.space = space;
    return result;
  }

  CreateAgentFileSpaceRequest._();

  factory CreateAgentFileSpaceRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateAgentFileSpaceRequest()..mergeFromBuffer(data, registry);
  factory CreateAgentFileSpaceRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateAgentFileSpaceRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CreateAgentFileSpaceRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: CreateAgentFileSpaceRequest.$_createMessage)
    ..aOM<AgentFileSpace>(1, _omitFieldNames ? '' : 'space',
        subBuilder: AgentFileSpace.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateAgentFileSpaceRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateAgentFileSpaceRequest copyWith(
          void Function(CreateAgentFileSpaceRequest) updates) =>
      super.copyWith(
              (message) => updates(message as CreateAgentFileSpaceRequest))
          as CreateAgentFileSpaceRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use CreateAgentFileSpaceRequest() / CreateAgentFileSpaceRequest.new instead')
  static CreateAgentFileSpaceRequest create() =>
      CreateAgentFileSpaceRequest._();
  static $pb.GeneratedMessage $_createMessage() =>
      CreateAgentFileSpaceRequest._();
  @$core.override
  CreateAgentFileSpaceRequest createEmptyInstance() =>
      CreateAgentFileSpaceRequest._();
  @$core.pragma('dart2js:noInline')
  static CreateAgentFileSpaceRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<CreateAgentFileSpaceRequest>(
          CreateAgentFileSpaceRequest.$_createMessage);
  static CreateAgentFileSpaceRequest? _defaultInstance;

  @$pb.TagNumber(1)
  AgentFileSpace get space => $_getN(0);
  @$pb.TagNumber(1)
  set space(AgentFileSpace value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasSpace() => $_has(0);
  @$pb.TagNumber(1)
  void clearSpace() => $_clearField(1);
  @$pb.TagNumber(1)
  AgentFileSpace ensureSpace() => $_ensure(0);
}

class CreateAgentFileSpaceResponse extends $pb.GeneratedMessage {
  factory CreateAgentFileSpaceResponse({
    AgentFileSpace? space,
  }) {
    final result = CreateAgentFileSpaceResponse._();
    if (space != null) result.space = space;
    return result;
  }

  CreateAgentFileSpaceResponse._();

  factory CreateAgentFileSpaceResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateAgentFileSpaceResponse()..mergeFromBuffer(data, registry);
  factory CreateAgentFileSpaceResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateAgentFileSpaceResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CreateAgentFileSpaceResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: CreateAgentFileSpaceResponse.$_createMessage)
    ..aOM<AgentFileSpace>(1, _omitFieldNames ? '' : 'space',
        subBuilder: AgentFileSpace.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateAgentFileSpaceResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateAgentFileSpaceResponse copyWith(
          void Function(CreateAgentFileSpaceResponse) updates) =>
      super.copyWith(
              (message) => updates(message as CreateAgentFileSpaceResponse))
          as CreateAgentFileSpaceResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use CreateAgentFileSpaceResponse() / CreateAgentFileSpaceResponse.new instead')
  static CreateAgentFileSpaceResponse create() =>
      CreateAgentFileSpaceResponse._();
  static $pb.GeneratedMessage $_createMessage() =>
      CreateAgentFileSpaceResponse._();
  @$core.override
  CreateAgentFileSpaceResponse createEmptyInstance() =>
      CreateAgentFileSpaceResponse._();
  @$core.pragma('dart2js:noInline')
  static CreateAgentFileSpaceResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<CreateAgentFileSpaceResponse>(
          CreateAgentFileSpaceResponse.$_createMessage);
  static CreateAgentFileSpaceResponse? _defaultInstance;

  @$pb.TagNumber(1)
  AgentFileSpace get space => $_getN(0);
  @$pb.TagNumber(1)
  set space(AgentFileSpace value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasSpace() => $_has(0);
  @$pb.TagNumber(1)
  void clearSpace() => $_clearField(1);
  @$pb.TagNumber(1)
  AgentFileSpace ensureSpace() => $_ensure(0);
}

class UpdateAgentFileSpaceRequest extends $pb.GeneratedMessage {
  factory UpdateAgentFileSpaceRequest({
    AgentFileSpace? space,
  }) {
    final result = UpdateAgentFileSpaceRequest._();
    if (space != null) result.space = space;
    return result;
  }

  UpdateAgentFileSpaceRequest._();

  factory UpdateAgentFileSpaceRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpdateAgentFileSpaceRequest()..mergeFromBuffer(data, registry);
  factory UpdateAgentFileSpaceRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpdateAgentFileSpaceRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'UpdateAgentFileSpaceRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: UpdateAgentFileSpaceRequest.$_createMessage)
    ..aOM<AgentFileSpace>(1, _omitFieldNames ? '' : 'space',
        subBuilder: AgentFileSpace.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateAgentFileSpaceRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateAgentFileSpaceRequest copyWith(
          void Function(UpdateAgentFileSpaceRequest) updates) =>
      super.copyWith(
              (message) => updates(message as UpdateAgentFileSpaceRequest))
          as UpdateAgentFileSpaceRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use UpdateAgentFileSpaceRequest() / UpdateAgentFileSpaceRequest.new instead')
  static UpdateAgentFileSpaceRequest create() =>
      UpdateAgentFileSpaceRequest._();
  static $pb.GeneratedMessage $_createMessage() =>
      UpdateAgentFileSpaceRequest._();
  @$core.override
  UpdateAgentFileSpaceRequest createEmptyInstance() =>
      UpdateAgentFileSpaceRequest._();
  @$core.pragma('dart2js:noInline')
  static UpdateAgentFileSpaceRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<UpdateAgentFileSpaceRequest>(
          UpdateAgentFileSpaceRequest.$_createMessage);
  static UpdateAgentFileSpaceRequest? _defaultInstance;

  @$pb.TagNumber(1)
  AgentFileSpace get space => $_getN(0);
  @$pb.TagNumber(1)
  set space(AgentFileSpace value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasSpace() => $_has(0);
  @$pb.TagNumber(1)
  void clearSpace() => $_clearField(1);
  @$pb.TagNumber(1)
  AgentFileSpace ensureSpace() => $_ensure(0);
}

class UpdateAgentFileSpaceResponse extends $pb.GeneratedMessage {
  factory UpdateAgentFileSpaceResponse({
    AgentFileSpace? space,
  }) {
    final result = UpdateAgentFileSpaceResponse._();
    if (space != null) result.space = space;
    return result;
  }

  UpdateAgentFileSpaceResponse._();

  factory UpdateAgentFileSpaceResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpdateAgentFileSpaceResponse()..mergeFromBuffer(data, registry);
  factory UpdateAgentFileSpaceResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpdateAgentFileSpaceResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'UpdateAgentFileSpaceResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: UpdateAgentFileSpaceResponse.$_createMessage)
    ..aOM<AgentFileSpace>(1, _omitFieldNames ? '' : 'space',
        subBuilder: AgentFileSpace.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateAgentFileSpaceResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateAgentFileSpaceResponse copyWith(
          void Function(UpdateAgentFileSpaceResponse) updates) =>
      super.copyWith(
              (message) => updates(message as UpdateAgentFileSpaceResponse))
          as UpdateAgentFileSpaceResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use UpdateAgentFileSpaceResponse() / UpdateAgentFileSpaceResponse.new instead')
  static UpdateAgentFileSpaceResponse create() =>
      UpdateAgentFileSpaceResponse._();
  static $pb.GeneratedMessage $_createMessage() =>
      UpdateAgentFileSpaceResponse._();
  @$core.override
  UpdateAgentFileSpaceResponse createEmptyInstance() =>
      UpdateAgentFileSpaceResponse._();
  @$core.pragma('dart2js:noInline')
  static UpdateAgentFileSpaceResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<UpdateAgentFileSpaceResponse>(
          UpdateAgentFileSpaceResponse.$_createMessage);
  static UpdateAgentFileSpaceResponse? _defaultInstance;

  @$pb.TagNumber(1)
  AgentFileSpace get space => $_getN(0);
  @$pb.TagNumber(1)
  set space(AgentFileSpace value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasSpace() => $_has(0);
  @$pb.TagNumber(1)
  void clearSpace() => $_clearField(1);
  @$pb.TagNumber(1)
  AgentFileSpace ensureSpace() => $_ensure(0);
}

class DeleteAgentFileSpaceRequest extends $pb.GeneratedMessage {
  factory DeleteAgentFileSpaceRequest({
    $core.String? id,
  }) {
    final result = DeleteAgentFileSpaceRequest._();
    if (id != null) result.id = id;
    return result;
  }

  DeleteAgentFileSpaceRequest._();

  factory DeleteAgentFileSpaceRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DeleteAgentFileSpaceRequest()..mergeFromBuffer(data, registry);
  factory DeleteAgentFileSpaceRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DeleteAgentFileSpaceRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'DeleteAgentFileSpaceRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: DeleteAgentFileSpaceRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteAgentFileSpaceRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteAgentFileSpaceRequest copyWith(
          void Function(DeleteAgentFileSpaceRequest) updates) =>
      super.copyWith(
              (message) => updates(message as DeleteAgentFileSpaceRequest))
          as DeleteAgentFileSpaceRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use DeleteAgentFileSpaceRequest() / DeleteAgentFileSpaceRequest.new instead')
  static DeleteAgentFileSpaceRequest create() =>
      DeleteAgentFileSpaceRequest._();
  static $pb.GeneratedMessage $_createMessage() =>
      DeleteAgentFileSpaceRequest._();
  @$core.override
  DeleteAgentFileSpaceRequest createEmptyInstance() =>
      DeleteAgentFileSpaceRequest._();
  @$core.pragma('dart2js:noInline')
  static DeleteAgentFileSpaceRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<DeleteAgentFileSpaceRequest>(
          DeleteAgentFileSpaceRequest.$_createMessage);
  static DeleteAgentFileSpaceRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);
}

class DeleteAgentFileSpaceResponse extends $pb.GeneratedMessage {
  factory DeleteAgentFileSpaceResponse() => DeleteAgentFileSpaceResponse._();

  DeleteAgentFileSpaceResponse._();

  factory DeleteAgentFileSpaceResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DeleteAgentFileSpaceResponse()..mergeFromBuffer(data, registry);
  factory DeleteAgentFileSpaceResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DeleteAgentFileSpaceResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'DeleteAgentFileSpaceResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: DeleteAgentFileSpaceResponse.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteAgentFileSpaceResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteAgentFileSpaceResponse copyWith(
          void Function(DeleteAgentFileSpaceResponse) updates) =>
      super.copyWith(
              (message) => updates(message as DeleteAgentFileSpaceResponse))
          as DeleteAgentFileSpaceResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use DeleteAgentFileSpaceResponse() / DeleteAgentFileSpaceResponse.new instead')
  static DeleteAgentFileSpaceResponse create() =>
      DeleteAgentFileSpaceResponse._();
  static $pb.GeneratedMessage $_createMessage() =>
      DeleteAgentFileSpaceResponse._();
  @$core.override
  DeleteAgentFileSpaceResponse createEmptyInstance() =>
      DeleteAgentFileSpaceResponse._();
  @$core.pragma('dart2js:noInline')
  static DeleteAgentFileSpaceResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<DeleteAgentFileSpaceResponse>(
          DeleteAgentFileSpaceResponse.$_createMessage);
  static DeleteAgentFileSpaceResponse? _defaultInstance;
}

class ListAgentFilesRequest extends $pb.GeneratedMessage {
  factory ListAgentFilesRequest({
    $core.String? spaceId,
    $core.String? pathPrefix,
  }) {
    final result = ListAgentFilesRequest._();
    if (spaceId != null) result.spaceId = spaceId;
    if (pathPrefix != null) result.pathPrefix = pathPrefix;
    return result;
  }

  ListAgentFilesRequest._();

  factory ListAgentFilesRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListAgentFilesRequest()..mergeFromBuffer(data, registry);
  factory ListAgentFilesRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListAgentFilesRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListAgentFilesRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: ListAgentFilesRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'spaceId')
    ..aOS(2, _omitFieldNames ? '' : 'pathPrefix')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListAgentFilesRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListAgentFilesRequest copyWith(
          void Function(ListAgentFilesRequest) updates) =>
      super.copyWith((message) => updates(message as ListAgentFilesRequest))
          as ListAgentFilesRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use ListAgentFilesRequest() / ListAgentFilesRequest.new instead')
  static ListAgentFilesRequest create() => ListAgentFilesRequest._();
  static $pb.GeneratedMessage $_createMessage() => ListAgentFilesRequest._();
  @$core.override
  ListAgentFilesRequest createEmptyInstance() => ListAgentFilesRequest._();
  @$core.pragma('dart2js:noInline')
  static ListAgentFilesRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ListAgentFilesRequest>(
          ListAgentFilesRequest.$_createMessage);
  static ListAgentFilesRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get spaceId => $_getSZ(0);
  @$pb.TagNumber(1)
  set spaceId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasSpaceId() => $_has(0);
  @$pb.TagNumber(1)
  void clearSpaceId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get pathPrefix => $_getSZ(1);
  @$pb.TagNumber(2)
  set pathPrefix($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasPathPrefix() => $_has(1);
  @$pb.TagNumber(2)
  void clearPathPrefix() => $_clearField(2);
}

class ListAgentFilesResponse extends $pb.GeneratedMessage {
  factory ListAgentFilesResponse({
    $core.Iterable<AgentFile>? files,
  }) {
    final result = ListAgentFilesResponse._();
    if (files != null) result.files.addAll(files);
    return result;
  }

  ListAgentFilesResponse._();

  factory ListAgentFilesResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListAgentFilesResponse()..mergeFromBuffer(data, registry);
  factory ListAgentFilesResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListAgentFilesResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListAgentFilesResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: ListAgentFilesResponse.$_createMessage)
    ..pPM<AgentFile>(1, _omitFieldNames ? '' : 'files',
        subBuilder: AgentFile.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListAgentFilesResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListAgentFilesResponse copyWith(
          void Function(ListAgentFilesResponse) updates) =>
      super.copyWith((message) => updates(message as ListAgentFilesResponse))
          as ListAgentFilesResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use ListAgentFilesResponse() / ListAgentFilesResponse.new instead')
  static ListAgentFilesResponse create() => ListAgentFilesResponse._();
  static $pb.GeneratedMessage $_createMessage() => ListAgentFilesResponse._();
  @$core.override
  ListAgentFilesResponse createEmptyInstance() => ListAgentFilesResponse._();
  @$core.pragma('dart2js:noInline')
  static ListAgentFilesResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ListAgentFilesResponse>(
          ListAgentFilesResponse.$_createMessage);
  static ListAgentFilesResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $pb.PbList<AgentFile> get files => $_getList(0);
}

class GetAgentFileRequest extends $pb.GeneratedMessage {
  factory GetAgentFileRequest({
    $core.String? spaceId,
    $core.String? path,
    $fixnum.Int64? version,
  }) {
    final result = GetAgentFileRequest._();
    if (spaceId != null) result.spaceId = spaceId;
    if (path != null) result.path = path;
    if (version != null) result.version = version;
    return result;
  }

  GetAgentFileRequest._();

  factory GetAgentFileRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetAgentFileRequest()..mergeFromBuffer(data, registry);
  factory GetAgentFileRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetAgentFileRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetAgentFileRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: GetAgentFileRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'spaceId')
    ..aOS(2, _omitFieldNames ? '' : 'path')
    ..aInt64(3, _omitFieldNames ? '' : 'version')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetAgentFileRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetAgentFileRequest copyWith(void Function(GetAgentFileRequest) updates) =>
      super.copyWith((message) => updates(message as GetAgentFileRequest))
          as GetAgentFileRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core
      .Deprecated('Use GetAgentFileRequest() / GetAgentFileRequest.new instead')
  static GetAgentFileRequest create() => GetAgentFileRequest._();
  static $pb.GeneratedMessage $_createMessage() => GetAgentFileRequest._();
  @$core.override
  GetAgentFileRequest createEmptyInstance() => GetAgentFileRequest._();
  @$core.pragma('dart2js:noInline')
  static GetAgentFileRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<GetAgentFileRequest>(
          GetAgentFileRequest.$_createMessage);
  static GetAgentFileRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get spaceId => $_getSZ(0);
  @$pb.TagNumber(1)
  set spaceId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasSpaceId() => $_has(0);
  @$pb.TagNumber(1)
  void clearSpaceId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get path => $_getSZ(1);
  @$pb.TagNumber(2)
  set path($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasPath() => $_has(1);
  @$pb.TagNumber(2)
  void clearPath() => $_clearField(2);

  @$pb.TagNumber(3)
  $fixnum.Int64 get version => $_getI64(2);
  @$pb.TagNumber(3)
  set version($fixnum.Int64 value) => $_setInt64(2, value);
  @$pb.TagNumber(3)
  $core.bool hasVersion() => $_has(2);
  @$pb.TagNumber(3)
  void clearVersion() => $_clearField(3);
}

class GetAgentFileResponse extends $pb.GeneratedMessage {
  factory GetAgentFileResponse({
    AgentFile? file,
    $core.String? content,
  }) {
    final result = GetAgentFileResponse._();
    if (file != null) result.file = file;
    if (content != null) result.content = content;
    return result;
  }

  GetAgentFileResponse._();

  factory GetAgentFileResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetAgentFileResponse()..mergeFromBuffer(data, registry);
  factory GetAgentFileResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetAgentFileResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetAgentFileResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: GetAgentFileResponse.$_createMessage)
    ..aOM<AgentFile>(1, _omitFieldNames ? '' : 'file',
        subBuilder: AgentFile.$_createMessage)
    ..aOS(2, _omitFieldNames ? '' : 'content')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetAgentFileResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetAgentFileResponse copyWith(void Function(GetAgentFileResponse) updates) =>
      super.copyWith((message) => updates(message as GetAgentFileResponse))
          as GetAgentFileResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use GetAgentFileResponse() / GetAgentFileResponse.new instead')
  static GetAgentFileResponse create() => GetAgentFileResponse._();
  static $pb.GeneratedMessage $_createMessage() => GetAgentFileResponse._();
  @$core.override
  GetAgentFileResponse createEmptyInstance() => GetAgentFileResponse._();
  @$core.pragma('dart2js:noInline')
  static GetAgentFileResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<GetAgentFileResponse>(
          GetAgentFileResponse.$_createMessage);
  static GetAgentFileResponse? _defaultInstance;

  @$pb.TagNumber(1)
  AgentFile get file => $_getN(0);
  @$pb.TagNumber(1)
  set file(AgentFile value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasFile() => $_has(0);
  @$pb.TagNumber(1)
  void clearFile() => $_clearField(1);
  @$pb.TagNumber(1)
  AgentFile ensureFile() => $_ensure(0);

  @$pb.TagNumber(2)
  $core.String get content => $_getSZ(1);
  @$pb.TagNumber(2)
  set content($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasContent() => $_has(1);
  @$pb.TagNumber(2)
  void clearContent() => $_clearField(2);
}

class WriteAgentFileRequest extends $pb.GeneratedMessage {
  factory WriteAgentFileRequest({
    $core.String? spaceId,
    $core.String? path,
    $core.String? content,
    $core.String? contentType,
    $core.Iterable<$core.MapEntry<$core.String, $core.String>>? metadata,
  }) {
    final result = WriteAgentFileRequest._();
    if (spaceId != null) result.spaceId = spaceId;
    if (path != null) result.path = path;
    if (content != null) result.content = content;
    if (contentType != null) result.contentType = contentType;
    if (metadata != null) result.metadata.addEntries(metadata);
    return result;
  }

  WriteAgentFileRequest._();

  factory WriteAgentFileRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      WriteAgentFileRequest()..mergeFromBuffer(data, registry);
  factory WriteAgentFileRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      WriteAgentFileRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'WriteAgentFileRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: WriteAgentFileRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'spaceId')
    ..aOS(2, _omitFieldNames ? '' : 'path')
    ..aOS(3, _omitFieldNames ? '' : 'content')
    ..aOS(4, _omitFieldNames ? '' : 'contentType')
    ..m<$core.String, $core.String>(5, _omitFieldNames ? '' : 'metadata',
        entryClassName: 'WriteAgentFileRequest.MetadataEntry',
        keyFieldType: $pb.PbFieldType.OS,
        valueFieldType: $pb.PbFieldType.OS,
        packageName: const $pb.PackageName('agents.v1'))
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  WriteAgentFileRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  WriteAgentFileRequest copyWith(
          void Function(WriteAgentFileRequest) updates) =>
      super.copyWith((message) => updates(message as WriteAgentFileRequest))
          as WriteAgentFileRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use WriteAgentFileRequest() / WriteAgentFileRequest.new instead')
  static WriteAgentFileRequest create() => WriteAgentFileRequest._();
  static $pb.GeneratedMessage $_createMessage() => WriteAgentFileRequest._();
  @$core.override
  WriteAgentFileRequest createEmptyInstance() => WriteAgentFileRequest._();
  @$core.pragma('dart2js:noInline')
  static WriteAgentFileRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<WriteAgentFileRequest>(
          WriteAgentFileRequest.$_createMessage);
  static WriteAgentFileRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get spaceId => $_getSZ(0);
  @$pb.TagNumber(1)
  set spaceId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasSpaceId() => $_has(0);
  @$pb.TagNumber(1)
  void clearSpaceId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get path => $_getSZ(1);
  @$pb.TagNumber(2)
  set path($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasPath() => $_has(1);
  @$pb.TagNumber(2)
  void clearPath() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get content => $_getSZ(2);
  @$pb.TagNumber(3)
  set content($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasContent() => $_has(2);
  @$pb.TagNumber(3)
  void clearContent() => $_clearField(3);

  @$pb.TagNumber(4)
  $core.String get contentType => $_getSZ(3);
  @$pb.TagNumber(4)
  set contentType($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasContentType() => $_has(3);
  @$pb.TagNumber(4)
  void clearContentType() => $_clearField(4);

  @$pb.TagNumber(5)
  $pb.PbMap<$core.String, $core.String> get metadata => $_getMap(4);
}

class WriteAgentFileResponse extends $pb.GeneratedMessage {
  factory WriteAgentFileResponse({
    AgentFile? file,
  }) {
    final result = WriteAgentFileResponse._();
    if (file != null) result.file = file;
    return result;
  }

  WriteAgentFileResponse._();

  factory WriteAgentFileResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      WriteAgentFileResponse()..mergeFromBuffer(data, registry);
  factory WriteAgentFileResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      WriteAgentFileResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'WriteAgentFileResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: WriteAgentFileResponse.$_createMessage)
    ..aOM<AgentFile>(1, _omitFieldNames ? '' : 'file',
        subBuilder: AgentFile.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  WriteAgentFileResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  WriteAgentFileResponse copyWith(
          void Function(WriteAgentFileResponse) updates) =>
      super.copyWith((message) => updates(message as WriteAgentFileResponse))
          as WriteAgentFileResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use WriteAgentFileResponse() / WriteAgentFileResponse.new instead')
  static WriteAgentFileResponse create() => WriteAgentFileResponse._();
  static $pb.GeneratedMessage $_createMessage() => WriteAgentFileResponse._();
  @$core.override
  WriteAgentFileResponse createEmptyInstance() => WriteAgentFileResponse._();
  @$core.pragma('dart2js:noInline')
  static WriteAgentFileResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<WriteAgentFileResponse>(
          WriteAgentFileResponse.$_createMessage);
  static WriteAgentFileResponse? _defaultInstance;

  @$pb.TagNumber(1)
  AgentFile get file => $_getN(0);
  @$pb.TagNumber(1)
  set file(AgentFile value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasFile() => $_has(0);
  @$pb.TagNumber(1)
  void clearFile() => $_clearField(1);
  @$pb.TagNumber(1)
  AgentFile ensureFile() => $_ensure(0);
}

class DeleteAgentFileRequest extends $pb.GeneratedMessage {
  factory DeleteAgentFileRequest({
    $core.String? spaceId,
    $core.String? path,
  }) {
    final result = DeleteAgentFileRequest._();
    if (spaceId != null) result.spaceId = spaceId;
    if (path != null) result.path = path;
    return result;
  }

  DeleteAgentFileRequest._();

  factory DeleteAgentFileRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DeleteAgentFileRequest()..mergeFromBuffer(data, registry);
  factory DeleteAgentFileRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DeleteAgentFileRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'DeleteAgentFileRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: DeleteAgentFileRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'spaceId')
    ..aOS(2, _omitFieldNames ? '' : 'path')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteAgentFileRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteAgentFileRequest copyWith(
          void Function(DeleteAgentFileRequest) updates) =>
      super.copyWith((message) => updates(message as DeleteAgentFileRequest))
          as DeleteAgentFileRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use DeleteAgentFileRequest() / DeleteAgentFileRequest.new instead')
  static DeleteAgentFileRequest create() => DeleteAgentFileRequest._();
  static $pb.GeneratedMessage $_createMessage() => DeleteAgentFileRequest._();
  @$core.override
  DeleteAgentFileRequest createEmptyInstance() => DeleteAgentFileRequest._();
  @$core.pragma('dart2js:noInline')
  static DeleteAgentFileRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<DeleteAgentFileRequest>(
          DeleteAgentFileRequest.$_createMessage);
  static DeleteAgentFileRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get spaceId => $_getSZ(0);
  @$pb.TagNumber(1)
  set spaceId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasSpaceId() => $_has(0);
  @$pb.TagNumber(1)
  void clearSpaceId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get path => $_getSZ(1);
  @$pb.TagNumber(2)
  set path($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasPath() => $_has(1);
  @$pb.TagNumber(2)
  void clearPath() => $_clearField(2);
}

class DeleteAgentFileResponse extends $pb.GeneratedMessage {
  factory DeleteAgentFileResponse() => DeleteAgentFileResponse._();

  DeleteAgentFileResponse._();

  factory DeleteAgentFileResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DeleteAgentFileResponse()..mergeFromBuffer(data, registry);
  factory DeleteAgentFileResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DeleteAgentFileResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'DeleteAgentFileResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: DeleteAgentFileResponse.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteAgentFileResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteAgentFileResponse copyWith(
          void Function(DeleteAgentFileResponse) updates) =>
      super.copyWith((message) => updates(message as DeleteAgentFileResponse))
          as DeleteAgentFileResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use DeleteAgentFileResponse() / DeleteAgentFileResponse.new instead')
  static DeleteAgentFileResponse create() => DeleteAgentFileResponse._();
  static $pb.GeneratedMessage $_createMessage() => DeleteAgentFileResponse._();
  @$core.override
  DeleteAgentFileResponse createEmptyInstance() => DeleteAgentFileResponse._();
  @$core.pragma('dart2js:noInline')
  static DeleteAgentFileResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<DeleteAgentFileResponse>(
          DeleteAgentFileResponse.$_createMessage);
  static DeleteAgentFileResponse? _defaultInstance;
}

class SearchAgentFilesRequest extends $pb.GeneratedMessage {
  factory SearchAgentFilesRequest({
    $core.String? spaceId,
    $core.String? query,
    $core.int? limit,
  }) {
    final result = SearchAgentFilesRequest._();
    if (spaceId != null) result.spaceId = spaceId;
    if (query != null) result.query = query;
    if (limit != null) result.limit = limit;
    return result;
  }

  SearchAgentFilesRequest._();

  factory SearchAgentFilesRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      SearchAgentFilesRequest()..mergeFromBuffer(data, registry);
  factory SearchAgentFilesRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      SearchAgentFilesRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'SearchAgentFilesRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: SearchAgentFilesRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'spaceId')
    ..aOS(2, _omitFieldNames ? '' : 'query')
    ..aI(3, _omitFieldNames ? '' : 'limit')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  SearchAgentFilesRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  SearchAgentFilesRequest copyWith(
          void Function(SearchAgentFilesRequest) updates) =>
      super.copyWith((message) => updates(message as SearchAgentFilesRequest))
          as SearchAgentFilesRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use SearchAgentFilesRequest() / SearchAgentFilesRequest.new instead')
  static SearchAgentFilesRequest create() => SearchAgentFilesRequest._();
  static $pb.GeneratedMessage $_createMessage() => SearchAgentFilesRequest._();
  @$core.override
  SearchAgentFilesRequest createEmptyInstance() => SearchAgentFilesRequest._();
  @$core.pragma('dart2js:noInline')
  static SearchAgentFilesRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<SearchAgentFilesRequest>(
          SearchAgentFilesRequest.$_createMessage);
  static SearchAgentFilesRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get spaceId => $_getSZ(0);
  @$pb.TagNumber(1)
  set spaceId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasSpaceId() => $_has(0);
  @$pb.TagNumber(1)
  void clearSpaceId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get query => $_getSZ(1);
  @$pb.TagNumber(2)
  set query($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasQuery() => $_has(1);
  @$pb.TagNumber(2)
  void clearQuery() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.int get limit => $_getIZ(2);
  @$pb.TagNumber(3)
  set limit($core.int value) => $_setSignedInt32(2, value);
  @$pb.TagNumber(3)
  $core.bool hasLimit() => $_has(2);
  @$pb.TagNumber(3)
  void clearLimit() => $_clearField(3);
}

class SearchAgentFilesResponse extends $pb.GeneratedMessage {
  factory SearchAgentFilesResponse({
    $core.Iterable<AgentFileSearchResult>? results,
  }) {
    final result = SearchAgentFilesResponse._();
    if (results != null) result.results.addAll(results);
    return result;
  }

  SearchAgentFilesResponse._();

  factory SearchAgentFilesResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      SearchAgentFilesResponse()..mergeFromBuffer(data, registry);
  factory SearchAgentFilesResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      SearchAgentFilesResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'SearchAgentFilesResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: SearchAgentFilesResponse.$_createMessage)
    ..pPM<AgentFileSearchResult>(1, _omitFieldNames ? '' : 'results',
        subBuilder: AgentFileSearchResult.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  SearchAgentFilesResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  SearchAgentFilesResponse copyWith(
          void Function(SearchAgentFilesResponse) updates) =>
      super.copyWith((message) => updates(message as SearchAgentFilesResponse))
          as SearchAgentFilesResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use SearchAgentFilesResponse() / SearchAgentFilesResponse.new instead')
  static SearchAgentFilesResponse create() => SearchAgentFilesResponse._();
  static $pb.GeneratedMessage $_createMessage() => SearchAgentFilesResponse._();
  @$core.override
  SearchAgentFilesResponse createEmptyInstance() =>
      SearchAgentFilesResponse._();
  @$core.pragma('dart2js:noInline')
  static SearchAgentFilesResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<SearchAgentFilesResponse>(
          SearchAgentFilesResponse.$_createMessage);
  static SearchAgentFilesResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $pb.PbList<AgentFileSearchResult> get results => $_getList(0);
}

class AgentFileSearchResult extends $pb.GeneratedMessage {
  factory AgentFileSearchResult({
    AgentFile? file,
    $core.Iterable<$core.String>? snippets,
  }) {
    final result = AgentFileSearchResult._();
    if (file != null) result.file = file;
    if (snippets != null) result.snippets.addAll(snippets);
    return result;
  }

  AgentFileSearchResult._();

  factory AgentFileSearchResult.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      AgentFileSearchResult()..mergeFromBuffer(data, registry);
  factory AgentFileSearchResult.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      AgentFileSearchResult()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'AgentFileSearchResult',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: AgentFileSearchResult.$_createMessage)
    ..aOM<AgentFile>(1, _omitFieldNames ? '' : 'file',
        subBuilder: AgentFile.$_createMessage)
    ..pPS(2, _omitFieldNames ? '' : 'snippets')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  AgentFileSearchResult clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  AgentFileSearchResult copyWith(
          void Function(AgentFileSearchResult) updates) =>
      super.copyWith((message) => updates(message as AgentFileSearchResult))
          as AgentFileSearchResult;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use AgentFileSearchResult() / AgentFileSearchResult.new instead')
  static AgentFileSearchResult create() => AgentFileSearchResult._();
  static $pb.GeneratedMessage $_createMessage() => AgentFileSearchResult._();
  @$core.override
  AgentFileSearchResult createEmptyInstance() => AgentFileSearchResult._();
  @$core.pragma('dart2js:noInline')
  static AgentFileSearchResult getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<AgentFileSearchResult>(
          AgentFileSearchResult.$_createMessage);
  static AgentFileSearchResult? _defaultInstance;

  @$pb.TagNumber(1)
  AgentFile get file => $_getN(0);
  @$pb.TagNumber(1)
  set file(AgentFile value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasFile() => $_has(0);
  @$pb.TagNumber(1)
  void clearFile() => $_clearField(1);
  @$pb.TagNumber(1)
  AgentFile ensureFile() => $_ensure(0);

  @$pb.TagNumber(2)
  $pb.PbList<$core.String> get snippets => $_getList(1);
}

class AgentFileServiceApi {
  final $pb.RpcClient _client;

  AgentFileServiceApi(this._client);

  $async.Future<ListAgentFileSpacesResponse> listAgentFileSpaces(
          $pb.ClientContext? ctx, ListAgentFileSpacesRequest request) =>
      _client.invoke<ListAgentFileSpacesResponse>(ctx, 'AgentFileService',
          'ListAgentFileSpaces', request, ListAgentFileSpacesResponse());
  $async.Future<GetAgentFileSpaceResponse> getAgentFileSpace(
          $pb.ClientContext? ctx, GetAgentFileSpaceRequest request) =>
      _client.invoke<GetAgentFileSpaceResponse>(ctx, 'AgentFileService',
          'GetAgentFileSpace', request, GetAgentFileSpaceResponse());
  $async.Future<CreateAgentFileSpaceResponse> createAgentFileSpace(
          $pb.ClientContext? ctx, CreateAgentFileSpaceRequest request) =>
      _client.invoke<CreateAgentFileSpaceResponse>(ctx, 'AgentFileService',
          'CreateAgentFileSpace', request, CreateAgentFileSpaceResponse());
  $async.Future<UpdateAgentFileSpaceResponse> updateAgentFileSpace(
          $pb.ClientContext? ctx, UpdateAgentFileSpaceRequest request) =>
      _client.invoke<UpdateAgentFileSpaceResponse>(ctx, 'AgentFileService',
          'UpdateAgentFileSpace', request, UpdateAgentFileSpaceResponse());
  $async.Future<DeleteAgentFileSpaceResponse> deleteAgentFileSpace(
          $pb.ClientContext? ctx, DeleteAgentFileSpaceRequest request) =>
      _client.invoke<DeleteAgentFileSpaceResponse>(ctx, 'AgentFileService',
          'DeleteAgentFileSpace', request, DeleteAgentFileSpaceResponse());
  $async.Future<ListAgentFilesResponse> listAgentFiles(
          $pb.ClientContext? ctx, ListAgentFilesRequest request) =>
      _client.invoke<ListAgentFilesResponse>(ctx, 'AgentFileService',
          'ListAgentFiles', request, ListAgentFilesResponse());
  $async.Future<GetAgentFileResponse> getAgentFile(
          $pb.ClientContext? ctx, GetAgentFileRequest request) =>
      _client.invoke<GetAgentFileResponse>(ctx, 'AgentFileService',
          'GetAgentFile', request, GetAgentFileResponse());
  $async.Future<WriteAgentFileResponse> writeAgentFile(
          $pb.ClientContext? ctx, WriteAgentFileRequest request) =>
      _client.invoke<WriteAgentFileResponse>(ctx, 'AgentFileService',
          'WriteAgentFile', request, WriteAgentFileResponse());
  $async.Future<DeleteAgentFileResponse> deleteAgentFile(
          $pb.ClientContext? ctx, DeleteAgentFileRequest request) =>
      _client.invoke<DeleteAgentFileResponse>(ctx, 'AgentFileService',
          'DeleteAgentFile', request, DeleteAgentFileResponse());
  $async.Future<SearchAgentFilesResponse> searchAgentFiles(
          $pb.ClientContext? ctx, SearchAgentFilesRequest request) =>
      _client.invoke<SearchAgentFilesResponse>(ctx, 'AgentFileService',
          'SearchAgentFiles', request, SearchAgentFilesResponse());
}

const $core.bool _omitFieldNames =
    $core.bool.fromEnvironment('protobuf.omit_field_names');
const $core.bool _omitMessageNames =
    $core.bool.fromEnvironment('protobuf.omit_message_names');
