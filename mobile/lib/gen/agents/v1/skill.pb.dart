// This is a generated file - do not edit.
//
// Generated from agents/v1/skill.proto.

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

export 'package:protobuf/protobuf.dart' show GeneratedMessageGenericExtensions;

/// Skill is a workspace-level shared bundle of instructions and resources
/// following the agentskills.io spec. The spec-validated name is the sole
/// identifier (unique per workspace); there is no separate generated ID.
/// Frontmatter fields are parsed from SKILL.md at write time; the SKILL.md
/// document itself is stored separately by the repository backend (ADR 0004).
class Skill extends $pb.GeneratedMessage {
  factory Skill({
    $core.String? name,
    $core.String? description,
    $core.String? license,
    $core.String? compatibility,
    $core.Iterable<$core.MapEntry<$core.String, $core.String>>? metadata,
    $core.Iterable<$core.String>? allowedTools,
    $fixnum.Int64? sizeBytes,
    $0.Timestamp? createdAt,
    $0.Timestamp? updatedAt,
    $core.String? workspaceId,
  }) {
    final result = Skill._();
    if (name != null) result.name = name;
    if (description != null) result.description = description;
    if (license != null) result.license = license;
    if (compatibility != null) result.compatibility = compatibility;
    if (metadata != null) result.metadata.addEntries(metadata);
    if (allowedTools != null) result.allowedTools.addAll(allowedTools);
    if (sizeBytes != null) result.sizeBytes = sizeBytes;
    if (createdAt != null) result.createdAt = createdAt;
    if (updatedAt != null) result.updatedAt = updatedAt;
    if (workspaceId != null) result.workspaceId = workspaceId;
    return result;
  }

  Skill._();

  factory Skill.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      Skill()..mergeFromBuffer(data, registry);
  factory Skill.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      Skill()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'Skill',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: Skill.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'name')
    ..aOS(2, _omitFieldNames ? '' : 'description')
    ..aOS(3, _omitFieldNames ? '' : 'license')
    ..aOS(4, _omitFieldNames ? '' : 'compatibility')
    ..m<$core.String, $core.String>(5, _omitFieldNames ? '' : 'metadata',
        entryClassName: 'Skill.MetadataEntry',
        keyFieldType: $pb.PbFieldType.OS,
        valueFieldType: $pb.PbFieldType.OS,
        packageName: const $pb.PackageName('agents.v1'))
    ..pPS(6, _omitFieldNames ? '' : 'allowedTools')
    ..aInt64(7, _omitFieldNames ? '' : 'sizeBytes')
    ..aOM<$0.Timestamp>(8, _omitFieldNames ? '' : 'createdAt',
        subBuilder: $0.Timestamp.$_createMessage)
    ..aOM<$0.Timestamp>(9, _omitFieldNames ? '' : 'updatedAt',
        subBuilder: $0.Timestamp.$_createMessage)
    ..aOS(100, _omitFieldNames ? '' : 'workspaceId')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Skill clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Skill copyWith(void Function(Skill) updates) =>
      super.copyWith((message) => updates(message as Skill)) as Skill;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use Skill() / Skill.new instead')
  static Skill create() => Skill._();
  static $pb.GeneratedMessage $_createMessage() => Skill._();
  @$core.override
  Skill createEmptyInstance() => Skill._();
  @$core.pragma('dart2js:noInline')
  static Skill getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<Skill>(Skill.$_createMessage);
  static Skill? _defaultInstance;

  /// @gotags: json:"name,omitempty" yaml:"name,omitempty" bson:"name,omitempty"
  @$pb.TagNumber(1)
  $core.String get name => $_getSZ(0);
  @$pb.TagNumber(1)
  set name($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasName() => $_has(0);
  @$pb.TagNumber(1)
  void clearName() => $_clearField(1);

  /// @gotags: json:"description,omitempty" yaml:"description,omitempty" bson:"description,omitempty"
  @$pb.TagNumber(2)
  $core.String get description => $_getSZ(1);
  @$pb.TagNumber(2)
  set description($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasDescription() => $_has(1);
  @$pb.TagNumber(2)
  void clearDescription() => $_clearField(2);

  /// @gotags: json:"license,omitempty" yaml:"license,omitempty" bson:"license,omitempty"
  @$pb.TagNumber(3)
  $core.String get license => $_getSZ(2);
  @$pb.TagNumber(3)
  set license($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasLicense() => $_has(2);
  @$pb.TagNumber(3)
  void clearLicense() => $_clearField(3);

  /// @gotags: json:"compatibility,omitempty" yaml:"compatibility,omitempty" bson:"compatibility,omitempty"
  @$pb.TagNumber(4)
  $core.String get compatibility => $_getSZ(3);
  @$pb.TagNumber(4)
  set compatibility($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasCompatibility() => $_has(3);
  @$pb.TagNumber(4)
  void clearCompatibility() => $_clearField(4);

  /// @gotags: json:"metadata,omitempty" yaml:"metadata,omitempty" bson:"metadata,omitempty"
  @$pb.TagNumber(5)
  $pb.PbMap<$core.String, $core.String> get metadata => $_getMap(4);

  /// Parsed from frontmatter; stored for metadata purposes, not enforced.
  /// @gotags: json:"allowed_tools,omitempty" yaml:"allowed_tools,omitempty" bson:"allowed_tools,omitempty"
  @$pb.TagNumber(6)
  $pb.PbList<$core.String> get allowedTools => $_getList(5);

  /// Size of the stored SKILL.md document in bytes.
  /// @gotags: json:"size_bytes,omitempty" yaml:"size_bytes,omitempty" bson:"size_bytes,omitempty"
  @$pb.TagNumber(7)
  $fixnum.Int64 get sizeBytes => $_getI64(6);
  @$pb.TagNumber(7)
  set sizeBytes($fixnum.Int64 value) => $_setInt64(6, value);
  @$pb.TagNumber(7)
  $core.bool hasSizeBytes() => $_has(6);
  @$pb.TagNumber(7)
  void clearSizeBytes() => $_clearField(7);

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

/// SkillResource is a file attached to a Skill under one of the spec
/// directories (references/, assets/, scripts/). Path metadata is indexed in
/// Mongo; content lives in the ContentStore (ADR 0004). The path (unique per
/// skill) is the identifier, relative to the skill root.
class SkillResource extends $pb.GeneratedMessage {
  factory SkillResource({
    $core.String? path,
    $fixnum.Int64? sizeBytes,
    $core.String? contentType,
    $0.Timestamp? createdAt,
    $0.Timestamp? updatedAt,
  }) {
    final result = SkillResource._();
    if (path != null) result.path = path;
    if (sizeBytes != null) result.sizeBytes = sizeBytes;
    if (contentType != null) result.contentType = contentType;
    if (createdAt != null) result.createdAt = createdAt;
    if (updatedAt != null) result.updatedAt = updatedAt;
    return result;
  }

  SkillResource._();

  factory SkillResource.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      SkillResource()..mergeFromBuffer(data, registry);
  factory SkillResource.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      SkillResource()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'SkillResource',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: SkillResource.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'path')
    ..aInt64(2, _omitFieldNames ? '' : 'sizeBytes')
    ..aOS(3, _omitFieldNames ? '' : 'contentType')
    ..aOM<$0.Timestamp>(4, _omitFieldNames ? '' : 'createdAt',
        subBuilder: $0.Timestamp.$_createMessage)
    ..aOM<$0.Timestamp>(5, _omitFieldNames ? '' : 'updatedAt',
        subBuilder: $0.Timestamp.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  SkillResource clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  SkillResource copyWith(void Function(SkillResource) updates) =>
      super.copyWith((message) => updates(message as SkillResource))
          as SkillResource;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use SkillResource() / SkillResource.new instead')
  static SkillResource create() => SkillResource._();
  static $pb.GeneratedMessage $_createMessage() => SkillResource._();
  @$core.override
  SkillResource createEmptyInstance() => SkillResource._();
  @$core.pragma('dart2js:noInline')
  static SkillResource getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<SkillResource>(
          SkillResource.$_createMessage);
  static SkillResource? _defaultInstance;

  /// @gotags: json:"path,omitempty" yaml:"path,omitempty" bson:"path,omitempty"
  @$pb.TagNumber(1)
  $core.String get path => $_getSZ(0);
  @$pb.TagNumber(1)
  set path($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasPath() => $_has(0);
  @$pb.TagNumber(1)
  void clearPath() => $_clearField(1);

  /// @gotags: json:"size_bytes,omitempty" yaml:"size_bytes,omitempty" bson:"size_bytes,omitempty"
  @$pb.TagNumber(2)
  $fixnum.Int64 get sizeBytes => $_getI64(1);
  @$pb.TagNumber(2)
  set sizeBytes($fixnum.Int64 value) => $_setInt64(1, value);
  @$pb.TagNumber(2)
  $core.bool hasSizeBytes() => $_has(1);
  @$pb.TagNumber(2)
  void clearSizeBytes() => $_clearField(2);

  /// @gotags: json:"content_type,omitempty" yaml:"content_type,omitempty" bson:"content_type,omitempty"
  @$pb.TagNumber(3)
  $core.String get contentType => $_getSZ(2);
  @$pb.TagNumber(3)
  set contentType($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasContentType() => $_has(2);
  @$pb.TagNumber(3)
  void clearContentType() => $_clearField(3);

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

  /// @gotags: json:"updated_at,omitempty" yaml:"updated_at,omitempty" bson:"updated_at,omitempty"
  @$pb.TagNumber(5)
  $0.Timestamp get updatedAt => $_getN(4);
  @$pb.TagNumber(5)
  set updatedAt($0.Timestamp value) => $_setField(5, value);
  @$pb.TagNumber(5)
  $core.bool hasUpdatedAt() => $_has(4);
  @$pb.TagNumber(5)
  void clearUpdatedAt() => $_clearField(5);
  @$pb.TagNumber(5)
  $0.Timestamp ensureUpdatedAt() => $_ensure(4);
}

class ListSkillsRequest extends $pb.GeneratedMessage {
  factory ListSkillsRequest() => ListSkillsRequest._();

  ListSkillsRequest._();

  factory ListSkillsRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListSkillsRequest()..mergeFromBuffer(data, registry);
  factory ListSkillsRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListSkillsRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListSkillsRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: ListSkillsRequest.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListSkillsRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListSkillsRequest copyWith(void Function(ListSkillsRequest) updates) =>
      super.copyWith((message) => updates(message as ListSkillsRequest))
          as ListSkillsRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use ListSkillsRequest() / ListSkillsRequest.new instead')
  static ListSkillsRequest create() => ListSkillsRequest._();
  static $pb.GeneratedMessage $_createMessage() => ListSkillsRequest._();
  @$core.override
  ListSkillsRequest createEmptyInstance() => ListSkillsRequest._();
  @$core.pragma('dart2js:noInline')
  static ListSkillsRequest getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<ListSkillsRequest>(
          ListSkillsRequest.$_createMessage);
  static ListSkillsRequest? _defaultInstance;
}

class ListSkillsResponse extends $pb.GeneratedMessage {
  factory ListSkillsResponse({
    $core.Iterable<Skill>? skills,
  }) {
    final result = ListSkillsResponse._();
    if (skills != null) result.skills.addAll(skills);
    return result;
  }

  ListSkillsResponse._();

  factory ListSkillsResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListSkillsResponse()..mergeFromBuffer(data, registry);
  factory ListSkillsResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListSkillsResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListSkillsResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: ListSkillsResponse.$_createMessage)
    ..pPM<Skill>(1, _omitFieldNames ? '' : 'skills',
        subBuilder: Skill.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListSkillsResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListSkillsResponse copyWith(void Function(ListSkillsResponse) updates) =>
      super.copyWith((message) => updates(message as ListSkillsResponse))
          as ListSkillsResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use ListSkillsResponse() / ListSkillsResponse.new instead')
  static ListSkillsResponse create() => ListSkillsResponse._();
  static $pb.GeneratedMessage $_createMessage() => ListSkillsResponse._();
  @$core.override
  ListSkillsResponse createEmptyInstance() => ListSkillsResponse._();
  @$core.pragma('dart2js:noInline')
  static ListSkillsResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ListSkillsResponse>(
          ListSkillsResponse.$_createMessage);
  static ListSkillsResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $pb.PbList<Skill> get skills => $_getList(0);
}

class GetSkillRequest extends $pb.GeneratedMessage {
  factory GetSkillRequest({
    $core.String? name,
  }) {
    final result = GetSkillRequest._();
    if (name != null) result.name = name;
    return result;
  }

  GetSkillRequest._();

  factory GetSkillRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetSkillRequest()..mergeFromBuffer(data, registry);
  factory GetSkillRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetSkillRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetSkillRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: GetSkillRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'name')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetSkillRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetSkillRequest copyWith(void Function(GetSkillRequest) updates) =>
      super.copyWith((message) => updates(message as GetSkillRequest))
          as GetSkillRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use GetSkillRequest() / GetSkillRequest.new instead')
  static GetSkillRequest create() => GetSkillRequest._();
  static $pb.GeneratedMessage $_createMessage() => GetSkillRequest._();
  @$core.override
  GetSkillRequest createEmptyInstance() => GetSkillRequest._();
  @$core.pragma('dart2js:noInline')
  static GetSkillRequest getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<GetSkillRequest>(
          GetSkillRequest.$_createMessage);
  static GetSkillRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get name => $_getSZ(0);
  @$pb.TagNumber(1)
  set name($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasName() => $_has(0);
  @$pb.TagNumber(1)
  void clearName() => $_clearField(1);
}

class GetSkillResponse extends $pb.GeneratedMessage {
  factory GetSkillResponse({
    Skill? skill,
    $core.String? skillMd,
  }) {
    final result = GetSkillResponse._();
    if (skill != null) result.skill = skill;
    if (skillMd != null) result.skillMd = skillMd;
    return result;
  }

  GetSkillResponse._();

  factory GetSkillResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetSkillResponse()..mergeFromBuffer(data, registry);
  factory GetSkillResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetSkillResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetSkillResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: GetSkillResponse.$_createMessage)
    ..aOM<Skill>(1, _omitFieldNames ? '' : 'skill',
        subBuilder: Skill.$_createMessage)
    ..aOS(2, _omitFieldNames ? '' : 'skillMd')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetSkillResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetSkillResponse copyWith(void Function(GetSkillResponse) updates) =>
      super.copyWith((message) => updates(message as GetSkillResponse))
          as GetSkillResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use GetSkillResponse() / GetSkillResponse.new instead')
  static GetSkillResponse create() => GetSkillResponse._();
  static $pb.GeneratedMessage $_createMessage() => GetSkillResponse._();
  @$core.override
  GetSkillResponse createEmptyInstance() => GetSkillResponse._();
  @$core.pragma('dart2js:noInline')
  static GetSkillResponse getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<GetSkillResponse>(
          GetSkillResponse.$_createMessage);
  static GetSkillResponse? _defaultInstance;

  @$pb.TagNumber(1)
  Skill get skill => $_getN(0);
  @$pb.TagNumber(1)
  set skill(Skill value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasSkill() => $_has(0);
  @$pb.TagNumber(1)
  void clearSkill() => $_clearField(1);
  @$pb.TagNumber(1)
  Skill ensureSkill() => $_ensure(0);

  /// Full SKILL.md document (frontmatter + instructions).
  @$pb.TagNumber(2)
  $core.String get skillMd => $_getSZ(1);
  @$pb.TagNumber(2)
  set skillMd($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasSkillMd() => $_has(1);
  @$pb.TagNumber(2)
  void clearSkillMd() => $_clearField(2);
}

class CreateSkillRequest extends $pb.GeneratedMessage {
  factory CreateSkillRequest({
    $core.String? name,
    $core.String? skillMd,
  }) {
    final result = CreateSkillRequest._();
    if (name != null) result.name = name;
    if (skillMd != null) result.skillMd = skillMd;
    return result;
  }

  CreateSkillRequest._();

  factory CreateSkillRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateSkillRequest()..mergeFromBuffer(data, registry);
  factory CreateSkillRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateSkillRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CreateSkillRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: CreateSkillRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'name')
    ..aOS(2, _omitFieldNames ? '' : 'skillMd')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateSkillRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateSkillRequest copyWith(void Function(CreateSkillRequest) updates) =>
      super.copyWith((message) => updates(message as CreateSkillRequest))
          as CreateSkillRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use CreateSkillRequest() / CreateSkillRequest.new instead')
  static CreateSkillRequest create() => CreateSkillRequest._();
  static $pb.GeneratedMessage $_createMessage() => CreateSkillRequest._();
  @$core.override
  CreateSkillRequest createEmptyInstance() => CreateSkillRequest._();
  @$core.pragma('dart2js:noInline')
  static CreateSkillRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<CreateSkillRequest>(
          CreateSkillRequest.$_createMessage);
  static CreateSkillRequest? _defaultInstance;

  /// Skill name; must equal the name declared in the SKILL.md frontmatter.
  @$pb.TagNumber(1)
  $core.String get name => $_getSZ(0);
  @$pb.TagNumber(1)
  set name($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasName() => $_has(0);
  @$pb.TagNumber(1)
  void clearName() => $_clearField(1);

  /// Full SKILL.md document (frontmatter + instructions).
  @$pb.TagNumber(2)
  $core.String get skillMd => $_getSZ(1);
  @$pb.TagNumber(2)
  set skillMd($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasSkillMd() => $_has(1);
  @$pb.TagNumber(2)
  void clearSkillMd() => $_clearField(2);
}

class CreateSkillResponse extends $pb.GeneratedMessage {
  factory CreateSkillResponse({
    Skill? skill,
  }) {
    final result = CreateSkillResponse._();
    if (skill != null) result.skill = skill;
    return result;
  }

  CreateSkillResponse._();

  factory CreateSkillResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateSkillResponse()..mergeFromBuffer(data, registry);
  factory CreateSkillResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateSkillResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CreateSkillResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: CreateSkillResponse.$_createMessage)
    ..aOM<Skill>(1, _omitFieldNames ? '' : 'skill',
        subBuilder: Skill.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateSkillResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateSkillResponse copyWith(void Function(CreateSkillResponse) updates) =>
      super.copyWith((message) => updates(message as CreateSkillResponse))
          as CreateSkillResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core
      .Deprecated('Use CreateSkillResponse() / CreateSkillResponse.new instead')
  static CreateSkillResponse create() => CreateSkillResponse._();
  static $pb.GeneratedMessage $_createMessage() => CreateSkillResponse._();
  @$core.override
  CreateSkillResponse createEmptyInstance() => CreateSkillResponse._();
  @$core.pragma('dart2js:noInline')
  static CreateSkillResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<CreateSkillResponse>(
          CreateSkillResponse.$_createMessage);
  static CreateSkillResponse? _defaultInstance;

  @$pb.TagNumber(1)
  Skill get skill => $_getN(0);
  @$pb.TagNumber(1)
  set skill(Skill value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasSkill() => $_has(0);
  @$pb.TagNumber(1)
  void clearSkill() => $_clearField(1);
  @$pb.TagNumber(1)
  Skill ensureSkill() => $_ensure(0);
}

class UpdateSkillRequest extends $pb.GeneratedMessage {
  factory UpdateSkillRequest({
    $core.String? name,
    $core.String? skillMd,
  }) {
    final result = UpdateSkillRequest._();
    if (name != null) result.name = name;
    if (skillMd != null) result.skillMd = skillMd;
    return result;
  }

  UpdateSkillRequest._();

  factory UpdateSkillRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpdateSkillRequest()..mergeFromBuffer(data, registry);
  factory UpdateSkillRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpdateSkillRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'UpdateSkillRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: UpdateSkillRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'name')
    ..aOS(2, _omitFieldNames ? '' : 'skillMd')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateSkillRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateSkillRequest copyWith(void Function(UpdateSkillRequest) updates) =>
      super.copyWith((message) => updates(message as UpdateSkillRequest))
          as UpdateSkillRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use UpdateSkillRequest() / UpdateSkillRequest.new instead')
  static UpdateSkillRequest create() => UpdateSkillRequest._();
  static $pb.GeneratedMessage $_createMessage() => UpdateSkillRequest._();
  @$core.override
  UpdateSkillRequest createEmptyInstance() => UpdateSkillRequest._();
  @$core.pragma('dart2js:noInline')
  static UpdateSkillRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<UpdateSkillRequest>(
          UpdateSkillRequest.$_createMessage);
  static UpdateSkillRequest? _defaultInstance;

  /// Existing skill name; the frontmatter name must match (name is immutable).
  @$pb.TagNumber(1)
  $core.String get name => $_getSZ(0);
  @$pb.TagNumber(1)
  set name($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasName() => $_has(0);
  @$pb.TagNumber(1)
  void clearName() => $_clearField(1);

  /// Full replacement SKILL.md document (frontmatter + instructions).
  @$pb.TagNumber(2)
  $core.String get skillMd => $_getSZ(1);
  @$pb.TagNumber(2)
  set skillMd($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasSkillMd() => $_has(1);
  @$pb.TagNumber(2)
  void clearSkillMd() => $_clearField(2);
}

class UpdateSkillResponse extends $pb.GeneratedMessage {
  factory UpdateSkillResponse({
    Skill? skill,
  }) {
    final result = UpdateSkillResponse._();
    if (skill != null) result.skill = skill;
    return result;
  }

  UpdateSkillResponse._();

  factory UpdateSkillResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpdateSkillResponse()..mergeFromBuffer(data, registry);
  factory UpdateSkillResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpdateSkillResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'UpdateSkillResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: UpdateSkillResponse.$_createMessage)
    ..aOM<Skill>(1, _omitFieldNames ? '' : 'skill',
        subBuilder: Skill.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateSkillResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateSkillResponse copyWith(void Function(UpdateSkillResponse) updates) =>
      super.copyWith((message) => updates(message as UpdateSkillResponse))
          as UpdateSkillResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core
      .Deprecated('Use UpdateSkillResponse() / UpdateSkillResponse.new instead')
  static UpdateSkillResponse create() => UpdateSkillResponse._();
  static $pb.GeneratedMessage $_createMessage() => UpdateSkillResponse._();
  @$core.override
  UpdateSkillResponse createEmptyInstance() => UpdateSkillResponse._();
  @$core.pragma('dart2js:noInline')
  static UpdateSkillResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<UpdateSkillResponse>(
          UpdateSkillResponse.$_createMessage);
  static UpdateSkillResponse? _defaultInstance;

  @$pb.TagNumber(1)
  Skill get skill => $_getN(0);
  @$pb.TagNumber(1)
  set skill(Skill value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasSkill() => $_has(0);
  @$pb.TagNumber(1)
  void clearSkill() => $_clearField(1);
  @$pb.TagNumber(1)
  Skill ensureSkill() => $_ensure(0);
}

class DeleteSkillRequest extends $pb.GeneratedMessage {
  factory DeleteSkillRequest({
    $core.String? name,
  }) {
    final result = DeleteSkillRequest._();
    if (name != null) result.name = name;
    return result;
  }

  DeleteSkillRequest._();

  factory DeleteSkillRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DeleteSkillRequest()..mergeFromBuffer(data, registry);
  factory DeleteSkillRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DeleteSkillRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'DeleteSkillRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: DeleteSkillRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'name')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteSkillRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteSkillRequest copyWith(void Function(DeleteSkillRequest) updates) =>
      super.copyWith((message) => updates(message as DeleteSkillRequest))
          as DeleteSkillRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use DeleteSkillRequest() / DeleteSkillRequest.new instead')
  static DeleteSkillRequest create() => DeleteSkillRequest._();
  static $pb.GeneratedMessage $_createMessage() => DeleteSkillRequest._();
  @$core.override
  DeleteSkillRequest createEmptyInstance() => DeleteSkillRequest._();
  @$core.pragma('dart2js:noInline')
  static DeleteSkillRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<DeleteSkillRequest>(
          DeleteSkillRequest.$_createMessage);
  static DeleteSkillRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get name => $_getSZ(0);
  @$pb.TagNumber(1)
  set name($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasName() => $_has(0);
  @$pb.TagNumber(1)
  void clearName() => $_clearField(1);
}

class DeleteSkillResponse extends $pb.GeneratedMessage {
  factory DeleteSkillResponse() => DeleteSkillResponse._();

  DeleteSkillResponse._();

  factory DeleteSkillResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DeleteSkillResponse()..mergeFromBuffer(data, registry);
  factory DeleteSkillResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DeleteSkillResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'DeleteSkillResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: DeleteSkillResponse.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteSkillResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteSkillResponse copyWith(void Function(DeleteSkillResponse) updates) =>
      super.copyWith((message) => updates(message as DeleteSkillResponse))
          as DeleteSkillResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core
      .Deprecated('Use DeleteSkillResponse() / DeleteSkillResponse.new instead')
  static DeleteSkillResponse create() => DeleteSkillResponse._();
  static $pb.GeneratedMessage $_createMessage() => DeleteSkillResponse._();
  @$core.override
  DeleteSkillResponse createEmptyInstance() => DeleteSkillResponse._();
  @$core.pragma('dart2js:noInline')
  static DeleteSkillResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<DeleteSkillResponse>(
          DeleteSkillResponse.$_createMessage);
  static DeleteSkillResponse? _defaultInstance;
}

class ListSkillResourcesRequest extends $pb.GeneratedMessage {
  factory ListSkillResourcesRequest({
    $core.String? skillName,
  }) {
    final result = ListSkillResourcesRequest._();
    if (skillName != null) result.skillName = skillName;
    return result;
  }

  ListSkillResourcesRequest._();

  factory ListSkillResourcesRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListSkillResourcesRequest()..mergeFromBuffer(data, registry);
  factory ListSkillResourcesRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListSkillResourcesRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListSkillResourcesRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: ListSkillResourcesRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'skillName')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListSkillResourcesRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListSkillResourcesRequest copyWith(
          void Function(ListSkillResourcesRequest) updates) =>
      super.copyWith((message) => updates(message as ListSkillResourcesRequest))
          as ListSkillResourcesRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use ListSkillResourcesRequest() / ListSkillResourcesRequest.new instead')
  static ListSkillResourcesRequest create() => ListSkillResourcesRequest._();
  static $pb.GeneratedMessage $_createMessage() =>
      ListSkillResourcesRequest._();
  @$core.override
  ListSkillResourcesRequest createEmptyInstance() =>
      ListSkillResourcesRequest._();
  @$core.pragma('dart2js:noInline')
  static ListSkillResourcesRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ListSkillResourcesRequest>(
          ListSkillResourcesRequest.$_createMessage);
  static ListSkillResourcesRequest? _defaultInstance;

  /// Skill name the resources belong to.
  @$pb.TagNumber(1)
  $core.String get skillName => $_getSZ(0);
  @$pb.TagNumber(1)
  set skillName($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasSkillName() => $_has(0);
  @$pb.TagNumber(1)
  void clearSkillName() => $_clearField(1);
}

class ListSkillResourcesResponse extends $pb.GeneratedMessage {
  factory ListSkillResourcesResponse({
    $core.Iterable<SkillResource>? resources,
  }) {
    final result = ListSkillResourcesResponse._();
    if (resources != null) result.resources.addAll(resources);
    return result;
  }

  ListSkillResourcesResponse._();

  factory ListSkillResourcesResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListSkillResourcesResponse()..mergeFromBuffer(data, registry);
  factory ListSkillResourcesResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListSkillResourcesResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListSkillResourcesResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: ListSkillResourcesResponse.$_createMessage)
    ..pPM<SkillResource>(1, _omitFieldNames ? '' : 'resources',
        subBuilder: SkillResource.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListSkillResourcesResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListSkillResourcesResponse copyWith(
          void Function(ListSkillResourcesResponse) updates) =>
      super.copyWith(
              (message) => updates(message as ListSkillResourcesResponse))
          as ListSkillResourcesResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use ListSkillResourcesResponse() / ListSkillResourcesResponse.new instead')
  static ListSkillResourcesResponse create() => ListSkillResourcesResponse._();
  static $pb.GeneratedMessage $_createMessage() =>
      ListSkillResourcesResponse._();
  @$core.override
  ListSkillResourcesResponse createEmptyInstance() =>
      ListSkillResourcesResponse._();
  @$core.pragma('dart2js:noInline')
  static ListSkillResourcesResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ListSkillResourcesResponse>(
          ListSkillResourcesResponse.$_createMessage);
  static ListSkillResourcesResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $pb.PbList<SkillResource> get resources => $_getList(0);
}

class GetSkillResourceRequest extends $pb.GeneratedMessage {
  factory GetSkillResourceRequest({
    $core.String? skillName,
    $core.String? path,
  }) {
    final result = GetSkillResourceRequest._();
    if (skillName != null) result.skillName = skillName;
    if (path != null) result.path = path;
    return result;
  }

  GetSkillResourceRequest._();

  factory GetSkillResourceRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetSkillResourceRequest()..mergeFromBuffer(data, registry);
  factory GetSkillResourceRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetSkillResourceRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetSkillResourceRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: GetSkillResourceRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'skillName')
    ..aOS(2, _omitFieldNames ? '' : 'path')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetSkillResourceRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetSkillResourceRequest copyWith(
          void Function(GetSkillResourceRequest) updates) =>
      super.copyWith((message) => updates(message as GetSkillResourceRequest))
          as GetSkillResourceRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use GetSkillResourceRequest() / GetSkillResourceRequest.new instead')
  static GetSkillResourceRequest create() => GetSkillResourceRequest._();
  static $pb.GeneratedMessage $_createMessage() => GetSkillResourceRequest._();
  @$core.override
  GetSkillResourceRequest createEmptyInstance() => GetSkillResourceRequest._();
  @$core.pragma('dart2js:noInline')
  static GetSkillResourceRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<GetSkillResourceRequest>(
          GetSkillResourceRequest.$_createMessage);
  static GetSkillResourceRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get skillName => $_getSZ(0);
  @$pb.TagNumber(1)
  set skillName($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasSkillName() => $_has(0);
  @$pb.TagNumber(1)
  void clearSkillName() => $_clearField(1);

  /// Resource path relative to the skill root, e.g. "references/api.md".
  @$pb.TagNumber(2)
  $core.String get path => $_getSZ(1);
  @$pb.TagNumber(2)
  set path($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasPath() => $_has(1);
  @$pb.TagNumber(2)
  void clearPath() => $_clearField(2);
}

class GetSkillResourceResponse extends $pb.GeneratedMessage {
  factory GetSkillResourceResponse({
    SkillResource? resource,
    $core.List<$core.int>? content,
  }) {
    final result = GetSkillResourceResponse._();
    if (resource != null) result.resource = resource;
    if (content != null) result.content = content;
    return result;
  }

  GetSkillResourceResponse._();

  factory GetSkillResourceResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetSkillResourceResponse()..mergeFromBuffer(data, registry);
  factory GetSkillResourceResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetSkillResourceResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetSkillResourceResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: GetSkillResourceResponse.$_createMessage)
    ..aOM<SkillResource>(1, _omitFieldNames ? '' : 'resource',
        subBuilder: SkillResource.$_createMessage)
    ..a<$core.List<$core.int>>(
        2, _omitFieldNames ? '' : 'content', $pb.PbFieldType.OY)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetSkillResourceResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetSkillResourceResponse copyWith(
          void Function(GetSkillResourceResponse) updates) =>
      super.copyWith((message) => updates(message as GetSkillResourceResponse))
          as GetSkillResourceResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use GetSkillResourceResponse() / GetSkillResourceResponse.new instead')
  static GetSkillResourceResponse create() => GetSkillResourceResponse._();
  static $pb.GeneratedMessage $_createMessage() => GetSkillResourceResponse._();
  @$core.override
  GetSkillResourceResponse createEmptyInstance() =>
      GetSkillResourceResponse._();
  @$core.pragma('dart2js:noInline')
  static GetSkillResourceResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<GetSkillResourceResponse>(
          GetSkillResourceResponse.$_createMessage);
  static GetSkillResourceResponse? _defaultInstance;

  @$pb.TagNumber(1)
  SkillResource get resource => $_getN(0);
  @$pb.TagNumber(1)
  set resource(SkillResource value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasResource() => $_has(0);
  @$pb.TagNumber(1)
  void clearResource() => $_clearField(1);
  @$pb.TagNumber(1)
  SkillResource ensureResource() => $_ensure(0);

  /// Raw resource content (binary-safe).
  @$pb.TagNumber(2)
  $core.List<$core.int> get content => $_getN(1);
  @$pb.TagNumber(2)
  set content($core.List<$core.int> value) => $_setBytes(1, value);
  @$pb.TagNumber(2)
  $core.bool hasContent() => $_has(1);
  @$pb.TagNumber(2)
  void clearContent() => $_clearField(2);
}

class PutSkillResourceRequest extends $pb.GeneratedMessage {
  factory PutSkillResourceRequest({
    $core.String? skillName,
    $core.String? path,
    $core.List<$core.int>? content,
    $core.String? contentType,
  }) {
    final result = PutSkillResourceRequest._();
    if (skillName != null) result.skillName = skillName;
    if (path != null) result.path = path;
    if (content != null) result.content = content;
    if (contentType != null) result.contentType = contentType;
    return result;
  }

  PutSkillResourceRequest._();

  factory PutSkillResourceRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      PutSkillResourceRequest()..mergeFromBuffer(data, registry);
  factory PutSkillResourceRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      PutSkillResourceRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'PutSkillResourceRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: PutSkillResourceRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'skillName')
    ..aOS(2, _omitFieldNames ? '' : 'path')
    ..a<$core.List<$core.int>>(
        3, _omitFieldNames ? '' : 'content', $pb.PbFieldType.OY)
    ..aOS(4, _omitFieldNames ? '' : 'contentType')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  PutSkillResourceRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  PutSkillResourceRequest copyWith(
          void Function(PutSkillResourceRequest) updates) =>
      super.copyWith((message) => updates(message as PutSkillResourceRequest))
          as PutSkillResourceRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use PutSkillResourceRequest() / PutSkillResourceRequest.new instead')
  static PutSkillResourceRequest create() => PutSkillResourceRequest._();
  static $pb.GeneratedMessage $_createMessage() => PutSkillResourceRequest._();
  @$core.override
  PutSkillResourceRequest createEmptyInstance() => PutSkillResourceRequest._();
  @$core.pragma('dart2js:noInline')
  static PutSkillResourceRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<PutSkillResourceRequest>(
          PutSkillResourceRequest.$_createMessage);
  static PutSkillResourceRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get skillName => $_getSZ(0);
  @$pb.TagNumber(1)
  set skillName($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasSkillName() => $_has(0);
  @$pb.TagNumber(1)
  void clearSkillName() => $_clearField(1);

  /// Resource path relative to the skill root; must fall under references/,
  /// assets/, or scripts/ after cleaning.
  @$pb.TagNumber(2)
  $core.String get path => $_getSZ(1);
  @$pb.TagNumber(2)
  set path($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasPath() => $_has(1);
  @$pb.TagNumber(2)
  void clearPath() => $_clearField(2);

  /// Raw resource content (binary-safe). Creates or overwrites in place.
  @$pb.TagNumber(3)
  $core.List<$core.int> get content => $_getN(2);
  @$pb.TagNumber(3)
  set content($core.List<$core.int> value) => $_setBytes(2, value);
  @$pb.TagNumber(3)
  $core.bool hasContent() => $_has(2);
  @$pb.TagNumber(3)
  void clearContent() => $_clearField(3);

  /// Optional MIME type stored as metadata; not validated.
  @$pb.TagNumber(4)
  $core.String get contentType => $_getSZ(3);
  @$pb.TagNumber(4)
  set contentType($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasContentType() => $_has(3);
  @$pb.TagNumber(4)
  void clearContentType() => $_clearField(4);
}

class PutSkillResourceResponse extends $pb.GeneratedMessage {
  factory PutSkillResourceResponse({
    SkillResource? resource,
  }) {
    final result = PutSkillResourceResponse._();
    if (resource != null) result.resource = resource;
    return result;
  }

  PutSkillResourceResponse._();

  factory PutSkillResourceResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      PutSkillResourceResponse()..mergeFromBuffer(data, registry);
  factory PutSkillResourceResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      PutSkillResourceResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'PutSkillResourceResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: PutSkillResourceResponse.$_createMessage)
    ..aOM<SkillResource>(1, _omitFieldNames ? '' : 'resource',
        subBuilder: SkillResource.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  PutSkillResourceResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  PutSkillResourceResponse copyWith(
          void Function(PutSkillResourceResponse) updates) =>
      super.copyWith((message) => updates(message as PutSkillResourceResponse))
          as PutSkillResourceResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use PutSkillResourceResponse() / PutSkillResourceResponse.new instead')
  static PutSkillResourceResponse create() => PutSkillResourceResponse._();
  static $pb.GeneratedMessage $_createMessage() => PutSkillResourceResponse._();
  @$core.override
  PutSkillResourceResponse createEmptyInstance() =>
      PutSkillResourceResponse._();
  @$core.pragma('dart2js:noInline')
  static PutSkillResourceResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<PutSkillResourceResponse>(
          PutSkillResourceResponse.$_createMessage);
  static PutSkillResourceResponse? _defaultInstance;

  @$pb.TagNumber(1)
  SkillResource get resource => $_getN(0);
  @$pb.TagNumber(1)
  set resource(SkillResource value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasResource() => $_has(0);
  @$pb.TagNumber(1)
  void clearResource() => $_clearField(1);
  @$pb.TagNumber(1)
  SkillResource ensureResource() => $_ensure(0);
}

class DeleteSkillResourceRequest extends $pb.GeneratedMessage {
  factory DeleteSkillResourceRequest({
    $core.String? skillName,
    $core.String? path,
  }) {
    final result = DeleteSkillResourceRequest._();
    if (skillName != null) result.skillName = skillName;
    if (path != null) result.path = path;
    return result;
  }

  DeleteSkillResourceRequest._();

  factory DeleteSkillResourceRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DeleteSkillResourceRequest()..mergeFromBuffer(data, registry);
  factory DeleteSkillResourceRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DeleteSkillResourceRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'DeleteSkillResourceRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: DeleteSkillResourceRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'skillName')
    ..aOS(2, _omitFieldNames ? '' : 'path')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteSkillResourceRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteSkillResourceRequest copyWith(
          void Function(DeleteSkillResourceRequest) updates) =>
      super.copyWith(
              (message) => updates(message as DeleteSkillResourceRequest))
          as DeleteSkillResourceRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use DeleteSkillResourceRequest() / DeleteSkillResourceRequest.new instead')
  static DeleteSkillResourceRequest create() => DeleteSkillResourceRequest._();
  static $pb.GeneratedMessage $_createMessage() =>
      DeleteSkillResourceRequest._();
  @$core.override
  DeleteSkillResourceRequest createEmptyInstance() =>
      DeleteSkillResourceRequest._();
  @$core.pragma('dart2js:noInline')
  static DeleteSkillResourceRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<DeleteSkillResourceRequest>(
          DeleteSkillResourceRequest.$_createMessage);
  static DeleteSkillResourceRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get skillName => $_getSZ(0);
  @$pb.TagNumber(1)
  set skillName($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasSkillName() => $_has(0);
  @$pb.TagNumber(1)
  void clearSkillName() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get path => $_getSZ(1);
  @$pb.TagNumber(2)
  set path($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasPath() => $_has(1);
  @$pb.TagNumber(2)
  void clearPath() => $_clearField(2);
}

class DeleteSkillResourceResponse extends $pb.GeneratedMessage {
  factory DeleteSkillResourceResponse() => DeleteSkillResourceResponse._();

  DeleteSkillResourceResponse._();

  factory DeleteSkillResourceResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DeleteSkillResourceResponse()..mergeFromBuffer(data, registry);
  factory DeleteSkillResourceResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DeleteSkillResourceResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'DeleteSkillResourceResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: DeleteSkillResourceResponse.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteSkillResourceResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteSkillResourceResponse copyWith(
          void Function(DeleteSkillResourceResponse) updates) =>
      super.copyWith(
              (message) => updates(message as DeleteSkillResourceResponse))
          as DeleteSkillResourceResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use DeleteSkillResourceResponse() / DeleteSkillResourceResponse.new instead')
  static DeleteSkillResourceResponse create() =>
      DeleteSkillResourceResponse._();
  static $pb.GeneratedMessage $_createMessage() =>
      DeleteSkillResourceResponse._();
  @$core.override
  DeleteSkillResourceResponse createEmptyInstance() =>
      DeleteSkillResourceResponse._();
  @$core.pragma('dart2js:noInline')
  static DeleteSkillResourceResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<DeleteSkillResourceResponse>(
          DeleteSkillResourceResponse.$_createMessage);
  static DeleteSkillResourceResponse? _defaultInstance;
}

class SkillServiceApi {
  final $pb.RpcClient _client;

  SkillServiceApi(this._client);

  $async.Future<ListSkillsResponse> listSkills(
          $pb.ClientContext? ctx, ListSkillsRequest request) =>
      _client.invoke<ListSkillsResponse>(
          ctx, 'SkillService', 'ListSkills', request, ListSkillsResponse());
  $async.Future<GetSkillResponse> getSkill(
          $pb.ClientContext? ctx, GetSkillRequest request) =>
      _client.invoke<GetSkillResponse>(
          ctx, 'SkillService', 'GetSkill', request, GetSkillResponse());
  $async.Future<CreateSkillResponse> createSkill(
          $pb.ClientContext? ctx, CreateSkillRequest request) =>
      _client.invoke<CreateSkillResponse>(
          ctx, 'SkillService', 'CreateSkill', request, CreateSkillResponse());
  $async.Future<UpdateSkillResponse> updateSkill(
          $pb.ClientContext? ctx, UpdateSkillRequest request) =>
      _client.invoke<UpdateSkillResponse>(
          ctx, 'SkillService', 'UpdateSkill', request, UpdateSkillResponse());
  $async.Future<DeleteSkillResponse> deleteSkill(
          $pb.ClientContext? ctx, DeleteSkillRequest request) =>
      _client.invoke<DeleteSkillResponse>(
          ctx, 'SkillService', 'DeleteSkill', request, DeleteSkillResponse());
  $async.Future<ListSkillResourcesResponse> listSkillResources(
          $pb.ClientContext? ctx, ListSkillResourcesRequest request) =>
      _client.invoke<ListSkillResourcesResponse>(ctx, 'SkillService',
          'ListSkillResources', request, ListSkillResourcesResponse());
  $async.Future<GetSkillResourceResponse> getSkillResource(
          $pb.ClientContext? ctx, GetSkillResourceRequest request) =>
      _client.invoke<GetSkillResourceResponse>(ctx, 'SkillService',
          'GetSkillResource', request, GetSkillResourceResponse());
  $async.Future<PutSkillResourceResponse> putSkillResource(
          $pb.ClientContext? ctx, PutSkillResourceRequest request) =>
      _client.invoke<PutSkillResourceResponse>(ctx, 'SkillService',
          'PutSkillResource', request, PutSkillResourceResponse());
  $async.Future<DeleteSkillResourceResponse> deleteSkillResource(
          $pb.ClientContext? ctx, DeleteSkillResourceRequest request) =>
      _client.invoke<DeleteSkillResourceResponse>(ctx, 'SkillService',
          'DeleteSkillResource', request, DeleteSkillResourceResponse());
}

const $core.bool _omitFieldNames =
    $core.bool.fromEnvironment('protobuf.omit_field_names');
const $core.bool _omitMessageNames =
    $core.bool.fromEnvironment('protobuf.omit_message_names');
