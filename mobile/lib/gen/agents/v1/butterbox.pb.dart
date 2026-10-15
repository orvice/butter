// This is a generated file - do not edit.
//
// Generated from agents/v1/butterbox.proto.

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

class ButterBox extends $pb.GeneratedMessage {
  factory ButterBox({
    $core.String? id,
    $core.String? name,
    $core.String? baseUrl,
    $core.bool? enabled,
    $core.bool? credentialSet,
    $0.Timestamp? credentialUpdatedAt,
    $0.Timestamp? createdAt,
    $0.Timestamp? updatedAt,
    $core.String? workspaceId,
  }) {
    final result = ButterBox._();
    if (id != null) result.id = id;
    if (name != null) result.name = name;
    if (baseUrl != null) result.baseUrl = baseUrl;
    if (enabled != null) result.enabled = enabled;
    if (credentialSet != null) result.credentialSet = credentialSet;
    if (credentialUpdatedAt != null)
      result.credentialUpdatedAt = credentialUpdatedAt;
    if (createdAt != null) result.createdAt = createdAt;
    if (updatedAt != null) result.updatedAt = updatedAt;
    if (workspaceId != null) result.workspaceId = workspaceId;
    return result;
  }

  ButterBox._();

  factory ButterBox.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ButterBox()..mergeFromBuffer(data, registry);
  factory ButterBox.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ButterBox()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ButterBox',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: ButterBox.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..aOS(2, _omitFieldNames ? '' : 'name')
    ..aOS(3, _omitFieldNames ? '' : 'baseUrl')
    ..aOB(4, _omitFieldNames ? '' : 'enabled')
    ..aOB(5, _omitFieldNames ? '' : 'credentialSet')
    ..aOM<$0.Timestamp>(6, _omitFieldNames ? '' : 'credentialUpdatedAt',
        subBuilder: $0.Timestamp.$_createMessage)
    ..aOM<$0.Timestamp>(7, _omitFieldNames ? '' : 'createdAt',
        subBuilder: $0.Timestamp.$_createMessage)
    ..aOM<$0.Timestamp>(8, _omitFieldNames ? '' : 'updatedAt',
        subBuilder: $0.Timestamp.$_createMessage)
    ..aOS(100, _omitFieldNames ? '' : 'workspaceId')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ButterBox clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ButterBox copyWith(void Function(ButterBox) updates) =>
      super.copyWith((message) => updates(message as ButterBox)) as ButterBox;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use ButterBox() / ButterBox.new instead')
  static ButterBox create() => ButterBox._();
  static $pb.GeneratedMessage $_createMessage() => ButterBox._();
  @$core.override
  ButterBox createEmptyInstance() => ButterBox._();
  @$core.pragma('dart2js:noInline')
  static ButterBox getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ButterBox>(ButterBox.$_createMessage);
  static ButterBox? _defaultInstance;

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

  /// Human-readable name, unique within the workspace.
  /// @gotags: json:"name,omitempty" yaml:"name,omitempty" bson:"name,omitempty"
  @$pb.TagNumber(2)
  $core.String get name => $_getSZ(1);
  @$pb.TagNumber(2)
  set name($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasName() => $_has(1);
  @$pb.TagNumber(2)
  void clearName() => $_clearField(2);

  /// Base URL the box's HTTP server listens on, e.g. "https://box.example.com".
  /// PiService lives under it at /butterbox.pi.v1.PiService/; pi-web (when
  /// enabled on the box) is served at its root.
  /// @gotags: json:"base_url,omitempty" yaml:"base_url,omitempty" bson:"base_url,omitempty"
  @$pb.TagNumber(3)
  $core.String get baseUrl => $_getSZ(2);
  @$pb.TagNumber(3)
  set baseUrl($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasBaseUrl() => $_has(2);
  @$pb.TagNumber(3)
  void clearBaseUrl() => $_clearField(3);

  /// When false the box is not offered for new Pi agent bindings.
  /// @gotags: json:"enabled,omitempty" yaml:"enabled,omitempty" bson:"enabled,omitempty"
  @$pb.TagNumber(4)
  $core.bool get enabled => $_getBF(3);
  @$pb.TagNumber(4)
  set enabled($core.bool value) => $_setBool(3, value);
  @$pb.TagNumber(4)
  $core.bool hasEnabled() => $_has(3);
  @$pb.TagNumber(4)
  void clearEnabled() => $_clearField(4);

  /// Reports whether an access token is stored for this box. Derived from
  /// the credential columns on every read; never taken from client input.
  /// @gotags: json:"credential_set,omitempty" yaml:"credential_set,omitempty" bson:"credential_set,omitempty"
  @$pb.TagNumber(5)
  $core.bool get credentialSet => $_getBF(4);
  @$pb.TagNumber(5)
  set credentialSet($core.bool value) => $_setBool(4, value);
  @$pb.TagNumber(5)
  $core.bool hasCredentialSet() => $_has(4);
  @$pb.TagNumber(5)
  void clearCredentialSet() => $_clearField(5);

  /// When the access token was last set or rotated. Unset when no token is
  /// stored.
  /// @gotags: json:"credential_updated_at,omitempty" yaml:"credential_updated_at,omitempty" bson:"credential_updated_at,omitempty"
  @$pb.TagNumber(6)
  $0.Timestamp get credentialUpdatedAt => $_getN(5);
  @$pb.TagNumber(6)
  set credentialUpdatedAt($0.Timestamp value) => $_setField(6, value);
  @$pb.TagNumber(6)
  $core.bool hasCredentialUpdatedAt() => $_has(5);
  @$pb.TagNumber(6)
  void clearCredentialUpdatedAt() => $_clearField(6);
  @$pb.TagNumber(6)
  $0.Timestamp ensureCredentialUpdatedAt() => $_ensure(5);

  /// @gotags: json:"created_at,omitempty" yaml:"created_at,omitempty" bson:"created_at,omitempty"
  @$pb.TagNumber(7)
  $0.Timestamp get createdAt => $_getN(6);
  @$pb.TagNumber(7)
  set createdAt($0.Timestamp value) => $_setField(7, value);
  @$pb.TagNumber(7)
  $core.bool hasCreatedAt() => $_has(6);
  @$pb.TagNumber(7)
  void clearCreatedAt() => $_clearField(7);
  @$pb.TagNumber(7)
  $0.Timestamp ensureCreatedAt() => $_ensure(6);

  /// @gotags: json:"updated_at,omitempty" yaml:"updated_at,omitempty" bson:"updated_at,omitempty"
  @$pb.TagNumber(8)
  $0.Timestamp get updatedAt => $_getN(7);
  @$pb.TagNumber(8)
  set updatedAt($0.Timestamp value) => $_setField(8, value);
  @$pb.TagNumber(8)
  $core.bool hasUpdatedAt() => $_has(7);
  @$pb.TagNumber(8)
  void clearUpdatedAt() => $_clearField(8);
  @$pb.TagNumber(8)
  $0.Timestamp ensureUpdatedAt() => $_ensure(7);

  /// ID of the workspace this box belongs to. Stamped by the repository from
  /// request context; never taken from client input.
  /// @gotags: json:"workspace_id,omitempty" yaml:"workspace_id,omitempty" bson:"workspace_id,omitempty"
  @$pb.TagNumber(100)
  $core.String get workspaceId => $_getSZ(8);
  @$pb.TagNumber(100)
  set workspaceId($core.String value) => $_setString(8, value);
  @$pb.TagNumber(100)
  $core.bool hasWorkspaceId() => $_has(8);
  @$pb.TagNumber(100)
  void clearWorkspaceId() => $_clearField(100);
}

class ListButterBoxesRequest extends $pb.GeneratedMessage {
  factory ListButterBoxesRequest() => ListButterBoxesRequest._();

  ListButterBoxesRequest._();

  factory ListButterBoxesRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListButterBoxesRequest()..mergeFromBuffer(data, registry);
  factory ListButterBoxesRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListButterBoxesRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListButterBoxesRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: ListButterBoxesRequest.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListButterBoxesRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListButterBoxesRequest copyWith(
          void Function(ListButterBoxesRequest) updates) =>
      super.copyWith((message) => updates(message as ListButterBoxesRequest))
          as ListButterBoxesRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use ListButterBoxesRequest() / ListButterBoxesRequest.new instead')
  static ListButterBoxesRequest create() => ListButterBoxesRequest._();
  static $pb.GeneratedMessage $_createMessage() => ListButterBoxesRequest._();
  @$core.override
  ListButterBoxesRequest createEmptyInstance() => ListButterBoxesRequest._();
  @$core.pragma('dart2js:noInline')
  static ListButterBoxesRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ListButterBoxesRequest>(
          ListButterBoxesRequest.$_createMessage);
  static ListButterBoxesRequest? _defaultInstance;
}

class ListButterBoxesResponse extends $pb.GeneratedMessage {
  factory ListButterBoxesResponse({
    $core.Iterable<ButterBox>? butterBoxes,
  }) {
    final result = ListButterBoxesResponse._();
    if (butterBoxes != null) result.butterBoxes.addAll(butterBoxes);
    return result;
  }

  ListButterBoxesResponse._();

  factory ListButterBoxesResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListButterBoxesResponse()..mergeFromBuffer(data, registry);
  factory ListButterBoxesResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListButterBoxesResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListButterBoxesResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: ListButterBoxesResponse.$_createMessage)
    ..pPM<ButterBox>(1, _omitFieldNames ? '' : 'butterBoxes',
        subBuilder: ButterBox.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListButterBoxesResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListButterBoxesResponse copyWith(
          void Function(ListButterBoxesResponse) updates) =>
      super.copyWith((message) => updates(message as ListButterBoxesResponse))
          as ListButterBoxesResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use ListButterBoxesResponse() / ListButterBoxesResponse.new instead')
  static ListButterBoxesResponse create() => ListButterBoxesResponse._();
  static $pb.GeneratedMessage $_createMessage() => ListButterBoxesResponse._();
  @$core.override
  ListButterBoxesResponse createEmptyInstance() => ListButterBoxesResponse._();
  @$core.pragma('dart2js:noInline')
  static ListButterBoxesResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ListButterBoxesResponse>(
          ListButterBoxesResponse.$_createMessage);
  static ListButterBoxesResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $pb.PbList<ButterBox> get butterBoxes => $_getList(0);
}

class GetButterBoxRequest extends $pb.GeneratedMessage {
  factory GetButterBoxRequest({
    $core.String? id,
  }) {
    final result = GetButterBoxRequest._();
    if (id != null) result.id = id;
    return result;
  }

  GetButterBoxRequest._();

  factory GetButterBoxRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetButterBoxRequest()..mergeFromBuffer(data, registry);
  factory GetButterBoxRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetButterBoxRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetButterBoxRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: GetButterBoxRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetButterBoxRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetButterBoxRequest copyWith(void Function(GetButterBoxRequest) updates) =>
      super.copyWith((message) => updates(message as GetButterBoxRequest))
          as GetButterBoxRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core
      .Deprecated('Use GetButterBoxRequest() / GetButterBoxRequest.new instead')
  static GetButterBoxRequest create() => GetButterBoxRequest._();
  static $pb.GeneratedMessage $_createMessage() => GetButterBoxRequest._();
  @$core.override
  GetButterBoxRequest createEmptyInstance() => GetButterBoxRequest._();
  @$core.pragma('dart2js:noInline')
  static GetButterBoxRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<GetButterBoxRequest>(
          GetButterBoxRequest.$_createMessage);
  static GetButterBoxRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);
}

class GetButterBoxResponse extends $pb.GeneratedMessage {
  factory GetButterBoxResponse({
    ButterBox? butterBox,
  }) {
    final result = GetButterBoxResponse._();
    if (butterBox != null) result.butterBox = butterBox;
    return result;
  }

  GetButterBoxResponse._();

  factory GetButterBoxResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetButterBoxResponse()..mergeFromBuffer(data, registry);
  factory GetButterBoxResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetButterBoxResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetButterBoxResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: GetButterBoxResponse.$_createMessage)
    ..aOM<ButterBox>(1, _omitFieldNames ? '' : 'butterBox',
        subBuilder: ButterBox.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetButterBoxResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetButterBoxResponse copyWith(void Function(GetButterBoxResponse) updates) =>
      super.copyWith((message) => updates(message as GetButterBoxResponse))
          as GetButterBoxResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use GetButterBoxResponse() / GetButterBoxResponse.new instead')
  static GetButterBoxResponse create() => GetButterBoxResponse._();
  static $pb.GeneratedMessage $_createMessage() => GetButterBoxResponse._();
  @$core.override
  GetButterBoxResponse createEmptyInstance() => GetButterBoxResponse._();
  @$core.pragma('dart2js:noInline')
  static GetButterBoxResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<GetButterBoxResponse>(
          GetButterBoxResponse.$_createMessage);
  static GetButterBoxResponse? _defaultInstance;

  @$pb.TagNumber(1)
  ButterBox get butterBox => $_getN(0);
  @$pb.TagNumber(1)
  set butterBox(ButterBox value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasButterBox() => $_has(0);
  @$pb.TagNumber(1)
  void clearButterBox() => $_clearField(1);
  @$pb.TagNumber(1)
  ButterBox ensureButterBox() => $_ensure(0);
}

class CreateButterBoxRequest extends $pb.GeneratedMessage {
  factory CreateButterBoxRequest({
    $core.String? name,
    $core.String? baseUrl,
    $core.bool? enabled,
    $core.String? token,
  }) {
    final result = CreateButterBoxRequest._();
    if (name != null) result.name = name;
    if (baseUrl != null) result.baseUrl = baseUrl;
    if (enabled != null) result.enabled = enabled;
    if (token != null) result.token = token;
    return result;
  }

  CreateButterBoxRequest._();

  factory CreateButterBoxRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateButterBoxRequest()..mergeFromBuffer(data, registry);
  factory CreateButterBoxRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateButterBoxRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CreateButterBoxRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: CreateButterBoxRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'name')
    ..aOS(2, _omitFieldNames ? '' : 'baseUrl')
    ..aOB(3, _omitFieldNames ? '' : 'enabled')
    ..aOS(4, _omitFieldNames ? '' : 'token')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateButterBoxRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateButterBoxRequest copyWith(
          void Function(CreateButterBoxRequest) updates) =>
      super.copyWith((message) => updates(message as CreateButterBoxRequest))
          as CreateButterBoxRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use CreateButterBoxRequest() / CreateButterBoxRequest.new instead')
  static CreateButterBoxRequest create() => CreateButterBoxRequest._();
  static $pb.GeneratedMessage $_createMessage() => CreateButterBoxRequest._();
  @$core.override
  CreateButterBoxRequest createEmptyInstance() => CreateButterBoxRequest._();
  @$core.pragma('dart2js:noInline')
  static CreateButterBoxRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<CreateButterBoxRequest>(
          CreateButterBoxRequest.$_createMessage);
  static CreateButterBoxRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get name => $_getSZ(0);
  @$pb.TagNumber(1)
  set name($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasName() => $_has(0);
  @$pb.TagNumber(1)
  void clearName() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get baseUrl => $_getSZ(1);
  @$pb.TagNumber(2)
  set baseUrl($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasBaseUrl() => $_has(1);
  @$pb.TagNumber(2)
  void clearBaseUrl() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.bool get enabled => $_getBF(2);
  @$pb.TagNumber(3)
  set enabled($core.bool value) => $_setBool(2, value);
  @$pb.TagNumber(3)
  $core.bool hasEnabled() => $_has(2);
  @$pb.TagNumber(3)
  void clearEnabled() => $_clearField(3);

  /// Optional access token, stored through the credential seam atomically
  /// with the box. Write-only.
  @$pb.TagNumber(4)
  $core.String get token => $_getSZ(3);
  @$pb.TagNumber(4)
  set token($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasToken() => $_has(3);
  @$pb.TagNumber(4)
  void clearToken() => $_clearField(4);
}

class CreateButterBoxResponse extends $pb.GeneratedMessage {
  factory CreateButterBoxResponse({
    ButterBox? butterBox,
  }) {
    final result = CreateButterBoxResponse._();
    if (butterBox != null) result.butterBox = butterBox;
    return result;
  }

  CreateButterBoxResponse._();

  factory CreateButterBoxResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateButterBoxResponse()..mergeFromBuffer(data, registry);
  factory CreateButterBoxResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateButterBoxResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CreateButterBoxResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: CreateButterBoxResponse.$_createMessage)
    ..aOM<ButterBox>(1, _omitFieldNames ? '' : 'butterBox',
        subBuilder: ButterBox.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateButterBoxResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateButterBoxResponse copyWith(
          void Function(CreateButterBoxResponse) updates) =>
      super.copyWith((message) => updates(message as CreateButterBoxResponse))
          as CreateButterBoxResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use CreateButterBoxResponse() / CreateButterBoxResponse.new instead')
  static CreateButterBoxResponse create() => CreateButterBoxResponse._();
  static $pb.GeneratedMessage $_createMessage() => CreateButterBoxResponse._();
  @$core.override
  CreateButterBoxResponse createEmptyInstance() => CreateButterBoxResponse._();
  @$core.pragma('dart2js:noInline')
  static CreateButterBoxResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<CreateButterBoxResponse>(
          CreateButterBoxResponse.$_createMessage);
  static CreateButterBoxResponse? _defaultInstance;

  @$pb.TagNumber(1)
  ButterBox get butterBox => $_getN(0);
  @$pb.TagNumber(1)
  set butterBox(ButterBox value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasButterBox() => $_has(0);
  @$pb.TagNumber(1)
  void clearButterBox() => $_clearField(1);
  @$pb.TagNumber(1)
  ButterBox ensureButterBox() => $_ensure(0);
}

class UpdateButterBoxRequest extends $pb.GeneratedMessage {
  factory UpdateButterBoxRequest({
    $core.String? id,
    $core.String? name,
    $core.String? baseUrl,
    $core.bool? enabled,
  }) {
    final result = UpdateButterBoxRequest._();
    if (id != null) result.id = id;
    if (name != null) result.name = name;
    if (baseUrl != null) result.baseUrl = baseUrl;
    if (enabled != null) result.enabled = enabled;
    return result;
  }

  UpdateButterBoxRequest._();

  factory UpdateButterBoxRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpdateButterBoxRequest()..mergeFromBuffer(data, registry);
  factory UpdateButterBoxRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpdateButterBoxRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'UpdateButterBoxRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: UpdateButterBoxRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..aOS(2, _omitFieldNames ? '' : 'name')
    ..aOS(3, _omitFieldNames ? '' : 'baseUrl')
    ..aOB(4, _omitFieldNames ? '' : 'enabled')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateButterBoxRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateButterBoxRequest copyWith(
          void Function(UpdateButterBoxRequest) updates) =>
      super.copyWith((message) => updates(message as UpdateButterBoxRequest))
          as UpdateButterBoxRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use UpdateButterBoxRequest() / UpdateButterBoxRequest.new instead')
  static UpdateButterBoxRequest create() => UpdateButterBoxRequest._();
  static $pb.GeneratedMessage $_createMessage() => UpdateButterBoxRequest._();
  @$core.override
  UpdateButterBoxRequest createEmptyInstance() => UpdateButterBoxRequest._();
  @$core.pragma('dart2js:noInline')
  static UpdateButterBoxRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<UpdateButterBoxRequest>(
          UpdateButterBoxRequest.$_createMessage);
  static UpdateButterBoxRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get name => $_getSZ(1);
  @$pb.TagNumber(2)
  set name($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasName() => $_has(1);
  @$pb.TagNumber(2)
  void clearName() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get baseUrl => $_getSZ(2);
  @$pb.TagNumber(3)
  set baseUrl($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasBaseUrl() => $_has(2);
  @$pb.TagNumber(3)
  void clearBaseUrl() => $_clearField(3);

  @$pb.TagNumber(4)
  $core.bool get enabled => $_getBF(3);
  @$pb.TagNumber(4)
  set enabled($core.bool value) => $_setBool(3, value);
  @$pb.TagNumber(4)
  $core.bool hasEnabled() => $_has(3);
  @$pb.TagNumber(4)
  void clearEnabled() => $_clearField(4);
}

class UpdateButterBoxResponse extends $pb.GeneratedMessage {
  factory UpdateButterBoxResponse({
    ButterBox? butterBox,
  }) {
    final result = UpdateButterBoxResponse._();
    if (butterBox != null) result.butterBox = butterBox;
    return result;
  }

  UpdateButterBoxResponse._();

  factory UpdateButterBoxResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpdateButterBoxResponse()..mergeFromBuffer(data, registry);
  factory UpdateButterBoxResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpdateButterBoxResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'UpdateButterBoxResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: UpdateButterBoxResponse.$_createMessage)
    ..aOM<ButterBox>(1, _omitFieldNames ? '' : 'butterBox',
        subBuilder: ButterBox.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateButterBoxResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateButterBoxResponse copyWith(
          void Function(UpdateButterBoxResponse) updates) =>
      super.copyWith((message) => updates(message as UpdateButterBoxResponse))
          as UpdateButterBoxResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use UpdateButterBoxResponse() / UpdateButterBoxResponse.new instead')
  static UpdateButterBoxResponse create() => UpdateButterBoxResponse._();
  static $pb.GeneratedMessage $_createMessage() => UpdateButterBoxResponse._();
  @$core.override
  UpdateButterBoxResponse createEmptyInstance() => UpdateButterBoxResponse._();
  @$core.pragma('dart2js:noInline')
  static UpdateButterBoxResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<UpdateButterBoxResponse>(
          UpdateButterBoxResponse.$_createMessage);
  static UpdateButterBoxResponse? _defaultInstance;

  @$pb.TagNumber(1)
  ButterBox get butterBox => $_getN(0);
  @$pb.TagNumber(1)
  set butterBox(ButterBox value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasButterBox() => $_has(0);
  @$pb.TagNumber(1)
  void clearButterBox() => $_clearField(1);
  @$pb.TagNumber(1)
  ButterBox ensureButterBox() => $_ensure(0);
}

class DeleteButterBoxRequest extends $pb.GeneratedMessage {
  factory DeleteButterBoxRequest({
    $core.String? id,
  }) {
    final result = DeleteButterBoxRequest._();
    if (id != null) result.id = id;
    return result;
  }

  DeleteButterBoxRequest._();

  factory DeleteButterBoxRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DeleteButterBoxRequest()..mergeFromBuffer(data, registry);
  factory DeleteButterBoxRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DeleteButterBoxRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'DeleteButterBoxRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: DeleteButterBoxRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteButterBoxRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteButterBoxRequest copyWith(
          void Function(DeleteButterBoxRequest) updates) =>
      super.copyWith((message) => updates(message as DeleteButterBoxRequest))
          as DeleteButterBoxRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use DeleteButterBoxRequest() / DeleteButterBoxRequest.new instead')
  static DeleteButterBoxRequest create() => DeleteButterBoxRequest._();
  static $pb.GeneratedMessage $_createMessage() => DeleteButterBoxRequest._();
  @$core.override
  DeleteButterBoxRequest createEmptyInstance() => DeleteButterBoxRequest._();
  @$core.pragma('dart2js:noInline')
  static DeleteButterBoxRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<DeleteButterBoxRequest>(
          DeleteButterBoxRequest.$_createMessage);
  static DeleteButterBoxRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);
}

class DeleteButterBoxResponse extends $pb.GeneratedMessage {
  factory DeleteButterBoxResponse() => DeleteButterBoxResponse._();

  DeleteButterBoxResponse._();

  factory DeleteButterBoxResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DeleteButterBoxResponse()..mergeFromBuffer(data, registry);
  factory DeleteButterBoxResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DeleteButterBoxResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'DeleteButterBoxResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: DeleteButterBoxResponse.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteButterBoxResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteButterBoxResponse copyWith(
          void Function(DeleteButterBoxResponse) updates) =>
      super.copyWith((message) => updates(message as DeleteButterBoxResponse))
          as DeleteButterBoxResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use DeleteButterBoxResponse() / DeleteButterBoxResponse.new instead')
  static DeleteButterBoxResponse create() => DeleteButterBoxResponse._();
  static $pb.GeneratedMessage $_createMessage() => DeleteButterBoxResponse._();
  @$core.override
  DeleteButterBoxResponse createEmptyInstance() => DeleteButterBoxResponse._();
  @$core.pragma('dart2js:noInline')
  static DeleteButterBoxResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<DeleteButterBoxResponse>(
          DeleteButterBoxResponse.$_createMessage);
  static DeleteButterBoxResponse? _defaultInstance;
}

class SetButterBoxTokenRequest extends $pb.GeneratedMessage {
  factory SetButterBoxTokenRequest({
    $core.String? id,
    $core.String? token,
  }) {
    final result = SetButterBoxTokenRequest._();
    if (id != null) result.id = id;
    if (token != null) result.token = token;
    return result;
  }

  SetButterBoxTokenRequest._();

  factory SetButterBoxTokenRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      SetButterBoxTokenRequest()..mergeFromBuffer(data, registry);
  factory SetButterBoxTokenRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      SetButterBoxTokenRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'SetButterBoxTokenRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: SetButterBoxTokenRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..aOS(2, _omitFieldNames ? '' : 'token')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  SetButterBoxTokenRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  SetButterBoxTokenRequest copyWith(
          void Function(SetButterBoxTokenRequest) updates) =>
      super.copyWith((message) => updates(message as SetButterBoxTokenRequest))
          as SetButterBoxTokenRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use SetButterBoxTokenRequest() / SetButterBoxTokenRequest.new instead')
  static SetButterBoxTokenRequest create() => SetButterBoxTokenRequest._();
  static $pb.GeneratedMessage $_createMessage() => SetButterBoxTokenRequest._();
  @$core.override
  SetButterBoxTokenRequest createEmptyInstance() =>
      SetButterBoxTokenRequest._();
  @$core.pragma('dart2js:noInline')
  static SetButterBoxTokenRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<SetButterBoxTokenRequest>(
          SetButterBoxTokenRequest.$_createMessage);
  static SetButterBoxTokenRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);

  /// New access token; empty clears the stored token.
  @$pb.TagNumber(2)
  $core.String get token => $_getSZ(1);
  @$pb.TagNumber(2)
  set token($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasToken() => $_has(1);
  @$pb.TagNumber(2)
  void clearToken() => $_clearField(2);
}

class SetButterBoxTokenResponse extends $pb.GeneratedMessage {
  factory SetButterBoxTokenResponse({
    ButterBox? butterBox,
  }) {
    final result = SetButterBoxTokenResponse._();
    if (butterBox != null) result.butterBox = butterBox;
    return result;
  }

  SetButterBoxTokenResponse._();

  factory SetButterBoxTokenResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      SetButterBoxTokenResponse()..mergeFromBuffer(data, registry);
  factory SetButterBoxTokenResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      SetButterBoxTokenResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'SetButterBoxTokenResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: SetButterBoxTokenResponse.$_createMessage)
    ..aOM<ButterBox>(1, _omitFieldNames ? '' : 'butterBox',
        subBuilder: ButterBox.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  SetButterBoxTokenResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  SetButterBoxTokenResponse copyWith(
          void Function(SetButterBoxTokenResponse) updates) =>
      super.copyWith((message) => updates(message as SetButterBoxTokenResponse))
          as SetButterBoxTokenResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use SetButterBoxTokenResponse() / SetButterBoxTokenResponse.new instead')
  static SetButterBoxTokenResponse create() => SetButterBoxTokenResponse._();
  static $pb.GeneratedMessage $_createMessage() =>
      SetButterBoxTokenResponse._();
  @$core.override
  SetButterBoxTokenResponse createEmptyInstance() =>
      SetButterBoxTokenResponse._();
  @$core.pragma('dart2js:noInline')
  static SetButterBoxTokenResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<SetButterBoxTokenResponse>(
          SetButterBoxTokenResponse.$_createMessage);
  static SetButterBoxTokenResponse? _defaultInstance;

  @$pb.TagNumber(1)
  ButterBox get butterBox => $_getN(0);
  @$pb.TagNumber(1)
  set butterBox(ButterBox value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasButterBox() => $_has(0);
  @$pb.TagNumber(1)
  void clearButterBox() => $_clearField(1);
  @$pb.TagNumber(1)
  ButterBox ensureButterBox() => $_ensure(0);
}

class GetButterBoxStatusRequest extends $pb.GeneratedMessage {
  factory GetButterBoxStatusRequest({
    $core.String? id,
  }) {
    final result = GetButterBoxStatusRequest._();
    if (id != null) result.id = id;
    return result;
  }

  GetButterBoxStatusRequest._();

  factory GetButterBoxStatusRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetButterBoxStatusRequest()..mergeFromBuffer(data, registry);
  factory GetButterBoxStatusRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetButterBoxStatusRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetButterBoxStatusRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: GetButterBoxStatusRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetButterBoxStatusRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetButterBoxStatusRequest copyWith(
          void Function(GetButterBoxStatusRequest) updates) =>
      super.copyWith((message) => updates(message as GetButterBoxStatusRequest))
          as GetButterBoxStatusRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use GetButterBoxStatusRequest() / GetButterBoxStatusRequest.new instead')
  static GetButterBoxStatusRequest create() => GetButterBoxStatusRequest._();
  static $pb.GeneratedMessage $_createMessage() =>
      GetButterBoxStatusRequest._();
  @$core.override
  GetButterBoxStatusRequest createEmptyInstance() =>
      GetButterBoxStatusRequest._();
  @$core.pragma('dart2js:noInline')
  static GetButterBoxStatusRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<GetButterBoxStatusRequest>(
          GetButterBoxStatusRequest.$_createMessage);
  static GetButterBoxStatusRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);
}

class GetButterBoxStatusResponse extends $pb.GeneratedMessage {
  factory GetButterBoxStatusResponse({
    $core.bool? reachable,
    $core.int? activeSessions,
    $core.String? error,
  }) {
    final result = GetButterBoxStatusResponse._();
    if (reachable != null) result.reachable = reachable;
    if (activeSessions != null) result.activeSessions = activeSessions;
    if (error != null) result.error = error;
    return result;
  }

  GetButterBoxStatusResponse._();

  factory GetButterBoxStatusResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetButterBoxStatusResponse()..mergeFromBuffer(data, registry);
  factory GetButterBoxStatusResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetButterBoxStatusResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetButterBoxStatusResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: GetButterBoxStatusResponse.$_createMessage)
    ..aOB(1, _omitFieldNames ? '' : 'reachable')
    ..aI(2, _omitFieldNames ? '' : 'activeSessions')
    ..aOS(3, _omitFieldNames ? '' : 'error')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetButterBoxStatusResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetButterBoxStatusResponse copyWith(
          void Function(GetButterBoxStatusResponse) updates) =>
      super.copyWith(
              (message) => updates(message as GetButterBoxStatusResponse))
          as GetButterBoxStatusResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use GetButterBoxStatusResponse() / GetButterBoxStatusResponse.new instead')
  static GetButterBoxStatusResponse create() => GetButterBoxStatusResponse._();
  static $pb.GeneratedMessage $_createMessage() =>
      GetButterBoxStatusResponse._();
  @$core.override
  GetButterBoxStatusResponse createEmptyInstance() =>
      GetButterBoxStatusResponse._();
  @$core.pragma('dart2js:noInline')
  static GetButterBoxStatusResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<GetButterBoxStatusResponse>(
          GetButterBoxStatusResponse.$_createMessage);
  static GetButterBoxStatusResponse? _defaultInstance;

  /// True when the box answered the probe.
  @$pb.TagNumber(1)
  $core.bool get reachable => $_getBF(0);
  @$pb.TagNumber(1)
  set reachable($core.bool value) => $_setBool(0, value);
  @$pb.TagNumber(1)
  $core.bool hasReachable() => $_has(0);
  @$pb.TagNumber(1)
  void clearReachable() => $_clearField(1);

  /// Number of active pi sessions on the box. Zero when unreachable.
  @$pb.TagNumber(2)
  $core.int get activeSessions => $_getIZ(1);
  @$pb.TagNumber(2)
  set activeSessions($core.int value) => $_setSignedInt32(1, value);
  @$pb.TagNumber(2)
  $core.bool hasActiveSessions() => $_has(1);
  @$pb.TagNumber(2)
  void clearActiveSessions() => $_clearField(2);

  /// Human-readable probe failure, empty when reachable.
  @$pb.TagNumber(3)
  $core.String get error => $_getSZ(2);
  @$pb.TagNumber(3)
  set error($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasError() => $_has(2);
  @$pb.TagNumber(3)
  void clearError() => $_clearField(3);
}

class ListButterBoxModelsRequest extends $pb.GeneratedMessage {
  factory ListButterBoxModelsRequest({
    $core.String? id,
  }) {
    final result = ListButterBoxModelsRequest._();
    if (id != null) result.id = id;
    return result;
  }

  ListButterBoxModelsRequest._();

  factory ListButterBoxModelsRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListButterBoxModelsRequest()..mergeFromBuffer(data, registry);
  factory ListButterBoxModelsRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListButterBoxModelsRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListButterBoxModelsRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: ListButterBoxModelsRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListButterBoxModelsRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListButterBoxModelsRequest copyWith(
          void Function(ListButterBoxModelsRequest) updates) =>
      super.copyWith(
              (message) => updates(message as ListButterBoxModelsRequest))
          as ListButterBoxModelsRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use ListButterBoxModelsRequest() / ListButterBoxModelsRequest.new instead')
  static ListButterBoxModelsRequest create() => ListButterBoxModelsRequest._();
  static $pb.GeneratedMessage $_createMessage() =>
      ListButterBoxModelsRequest._();
  @$core.override
  ListButterBoxModelsRequest createEmptyInstance() =>
      ListButterBoxModelsRequest._();
  @$core.pragma('dart2js:noInline')
  static ListButterBoxModelsRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ListButterBoxModelsRequest>(
          ListButterBoxModelsRequest.$_createMessage);
  static ListButterBoxModelsRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);
}

class ListButterBoxModelsResponse extends $pb.GeneratedMessage {
  factory ListButterBoxModelsResponse({
    $core.Iterable<ButterBoxModel>? models,
  }) {
    final result = ListButterBoxModelsResponse._();
    if (models != null) result.models.addAll(models);
    return result;
  }

  ListButterBoxModelsResponse._();

  factory ListButterBoxModelsResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListButterBoxModelsResponse()..mergeFromBuffer(data, registry);
  factory ListButterBoxModelsResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListButterBoxModelsResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListButterBoxModelsResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: ListButterBoxModelsResponse.$_createMessage)
    ..pPM<ButterBoxModel>(1, _omitFieldNames ? '' : 'models',
        subBuilder: ButterBoxModel.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListButterBoxModelsResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListButterBoxModelsResponse copyWith(
          void Function(ListButterBoxModelsResponse) updates) =>
      super.copyWith(
              (message) => updates(message as ListButterBoxModelsResponse))
          as ListButterBoxModelsResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use ListButterBoxModelsResponse() / ListButterBoxModelsResponse.new instead')
  static ListButterBoxModelsResponse create() =>
      ListButterBoxModelsResponse._();
  static $pb.GeneratedMessage $_createMessage() =>
      ListButterBoxModelsResponse._();
  @$core.override
  ListButterBoxModelsResponse createEmptyInstance() =>
      ListButterBoxModelsResponse._();
  @$core.pragma('dart2js:noInline')
  static ListButterBoxModelsResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ListButterBoxModelsResponse>(
          ListButterBoxModelsResponse.$_createMessage);
  static ListButterBoxModelsResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $pb.PbList<ButterBoxModel> get models => $_getList(0);
}

/// ButterBoxModel is one entry of a box's pi model catalog, trimmed to what
/// agent configuration needs.
class ButterBoxModel extends $pb.GeneratedMessage {
  factory ButterBoxModel({
    $core.String? id,
    $core.String? provider,
    $core.String? name,
    $core.bool? reasoning,
  }) {
    final result = ButterBoxModel._();
    if (id != null) result.id = id;
    if (provider != null) result.provider = provider;
    if (name != null) result.name = name;
    if (reasoning != null) result.reasoning = reasoning;
    return result;
  }

  ButterBoxModel._();

  factory ButterBoxModel.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ButterBoxModel()..mergeFromBuffer(data, registry);
  factory ButterBoxModel.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ButterBoxModel()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ButterBoxModel',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: ButterBoxModel.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..aOS(2, _omitFieldNames ? '' : 'provider')
    ..aOS(3, _omitFieldNames ? '' : 'name')
    ..aOB(4, _omitFieldNames ? '' : 'reasoning')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ButterBoxModel clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ButterBoxModel copyWith(void Function(ButterBoxModel) updates) =>
      super.copyWith((message) => updates(message as ButterBoxModel))
          as ButterBoxModel;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use ButterBoxModel() / ButterBoxModel.new instead')
  static ButterBoxModel create() => ButterBoxModel._();
  static $pb.GeneratedMessage $_createMessage() => ButterBoxModel._();
  @$core.override
  ButterBoxModel createEmptyInstance() => ButterBoxModel._();
  @$core.pragma('dart2js:noInline')
  static ButterBoxModel getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<ButterBoxModel>(
          ButterBoxModel.$_createMessage);
  static ButterBoxModel? _defaultInstance;

  /// Model ID as accepted by pi (e.g. "claude-fable-5").
  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);

  /// Provider ID (e.g. "anthropic").
  @$pb.TagNumber(2)
  $core.String get provider => $_getSZ(1);
  @$pb.TagNumber(2)
  set provider($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasProvider() => $_has(1);
  @$pb.TagNumber(2)
  void clearProvider() => $_clearField(2);

  /// Human-readable model name.
  @$pb.TagNumber(3)
  $core.String get name => $_getSZ(2);
  @$pb.TagNumber(3)
  set name($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasName() => $_has(2);
  @$pb.TagNumber(3)
  void clearName() => $_clearField(3);

  /// True when the model supports reasoning/thinking levels.
  @$pb.TagNumber(4)
  $core.bool get reasoning => $_getBF(3);
  @$pb.TagNumber(4)
  set reasoning($core.bool value) => $_setBool(3, value);
  @$pb.TagNumber(4)
  $core.bool hasReasoning() => $_has(3);
  @$pb.TagNumber(4)
  void clearReasoning() => $_clearField(4);
}

class ListButterBoxCursorModelsRequest extends $pb.GeneratedMessage {
  factory ListButterBoxCursorModelsRequest({
    $core.String? id,
  }) {
    final result = ListButterBoxCursorModelsRequest._();
    if (id != null) result.id = id;
    return result;
  }

  ListButterBoxCursorModelsRequest._();

  factory ListButterBoxCursorModelsRequest.fromBuffer(
          $core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListButterBoxCursorModelsRequest()..mergeFromBuffer(data, registry);
  factory ListButterBoxCursorModelsRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListButterBoxCursorModelsRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListButterBoxCursorModelsRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: ListButterBoxCursorModelsRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListButterBoxCursorModelsRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListButterBoxCursorModelsRequest copyWith(
          void Function(ListButterBoxCursorModelsRequest) updates) =>
      super.copyWith(
              (message) => updates(message as ListButterBoxCursorModelsRequest))
          as ListButterBoxCursorModelsRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use ListButterBoxCursorModelsRequest() / ListButterBoxCursorModelsRequest.new instead')
  static ListButterBoxCursorModelsRequest create() =>
      ListButterBoxCursorModelsRequest._();
  static $pb.GeneratedMessage $_createMessage() =>
      ListButterBoxCursorModelsRequest._();
  @$core.override
  ListButterBoxCursorModelsRequest createEmptyInstance() =>
      ListButterBoxCursorModelsRequest._();
  @$core.pragma('dart2js:noInline')
  static ListButterBoxCursorModelsRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ListButterBoxCursorModelsRequest>(
          ListButterBoxCursorModelsRequest.$_createMessage);
  static ListButterBoxCursorModelsRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);
}

class ListButterBoxCursorModelsResponse extends $pb.GeneratedMessage {
  factory ListButterBoxCursorModelsResponse({
    $core.Iterable<ButterBoxCursorModel>? models,
  }) {
    final result = ListButterBoxCursorModelsResponse._();
    if (models != null) result.models.addAll(models);
    return result;
  }

  ListButterBoxCursorModelsResponse._();

  factory ListButterBoxCursorModelsResponse.fromBuffer(
          $core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListButterBoxCursorModelsResponse()..mergeFromBuffer(data, registry);
  factory ListButterBoxCursorModelsResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListButterBoxCursorModelsResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListButterBoxCursorModelsResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: ListButterBoxCursorModelsResponse.$_createMessage)
    ..pPM<ButterBoxCursorModel>(1, _omitFieldNames ? '' : 'models',
        subBuilder: ButterBoxCursorModel.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListButterBoxCursorModelsResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListButterBoxCursorModelsResponse copyWith(
          void Function(ListButterBoxCursorModelsResponse) updates) =>
      super.copyWith((message) =>
              updates(message as ListButterBoxCursorModelsResponse))
          as ListButterBoxCursorModelsResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use ListButterBoxCursorModelsResponse() / ListButterBoxCursorModelsResponse.new instead')
  static ListButterBoxCursorModelsResponse create() =>
      ListButterBoxCursorModelsResponse._();
  static $pb.GeneratedMessage $_createMessage() =>
      ListButterBoxCursorModelsResponse._();
  @$core.override
  ListButterBoxCursorModelsResponse createEmptyInstance() =>
      ListButterBoxCursorModelsResponse._();
  @$core.pragma('dart2js:noInline')
  static ListButterBoxCursorModelsResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ListButterBoxCursorModelsResponse>(
          ListButterBoxCursorModelsResponse.$_createMessage);
  static ListButterBoxCursorModelsResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $pb.PbList<ButterBoxCursorModel> get models => $_getList(0);
}

/// ButterBoxCursorModel is one entry of a box's Cursor model catalog.
class ButterBoxCursorModel extends $pb.GeneratedMessage {
  factory ButterBoxCursorModel({
    $core.String? id,
    $core.String? name,
  }) {
    final result = ButterBoxCursorModel._();
    if (id != null) result.id = id;
    if (name != null) result.name = name;
    return result;
  }

  ButterBoxCursorModel._();

  factory ButterBoxCursorModel.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ButterBoxCursorModel()..mergeFromBuffer(data, registry);
  factory ButterBoxCursorModel.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ButterBoxCursorModel()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ButterBoxCursorModel',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: ButterBoxCursorModel.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..aOS(2, _omitFieldNames ? '' : 'name')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ButterBoxCursorModel clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ButterBoxCursorModel copyWith(void Function(ButterBoxCursorModel) updates) =>
      super.copyWith((message) => updates(message as ButterBoxCursorModel))
          as ButterBoxCursorModel;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use ButterBoxCursorModel() / ButterBoxCursorModel.new instead')
  static ButterBoxCursorModel create() => ButterBoxCursorModel._();
  static $pb.GeneratedMessage $_createMessage() => ButterBoxCursorModel._();
  @$core.override
  ButterBoxCursorModel createEmptyInstance() => ButterBoxCursorModel._();
  @$core.pragma('dart2js:noInline')
  static ButterBoxCursorModel getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ButterBoxCursorModel>(
          ButterBoxCursorModel.$_createMessage);
  static ButterBoxCursorModel? _defaultInstance;

  /// Model ID as accepted by Cursor (e.g. "composer-2.5").
  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);

  /// Human-readable model name.
  @$pb.TagNumber(2)
  $core.String get name => $_getSZ(1);
  @$pb.TagNumber(2)
  set name($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasName() => $_has(1);
  @$pb.TagNumber(2)
  void clearName() => $_clearField(2);
}

/// ButterBoxService manages the workspace's registered ButterBoxes.
class ButterBoxServiceApi {
  final $pb.RpcClient _client;

  ButterBoxServiceApi(this._client);

  $async.Future<ListButterBoxesResponse> listButterBoxes(
          $pb.ClientContext? ctx, ListButterBoxesRequest request) =>
      _client.invoke<ListButterBoxesResponse>(ctx, 'ButterBoxService',
          'ListButterBoxes', request, ListButterBoxesResponse());
  $async.Future<GetButterBoxResponse> getButterBox(
          $pb.ClientContext? ctx, GetButterBoxRequest request) =>
      _client.invoke<GetButterBoxResponse>(ctx, 'ButterBoxService',
          'GetButterBox', request, GetButterBoxResponse());
  $async.Future<CreateButterBoxResponse> createButterBox(
          $pb.ClientContext? ctx, CreateButterBoxRequest request) =>
      _client.invoke<CreateButterBoxResponse>(ctx, 'ButterBoxService',
          'CreateButterBox', request, CreateButterBoxResponse());
  $async.Future<UpdateButterBoxResponse> updateButterBox(
          $pb.ClientContext? ctx, UpdateButterBoxRequest request) =>
      _client.invoke<UpdateButterBoxResponse>(ctx, 'ButterBoxService',
          'UpdateButterBox', request, UpdateButterBoxResponse());
  $async.Future<DeleteButterBoxResponse> deleteButterBox(
          $pb.ClientContext? ctx, DeleteButterBoxRequest request) =>
      _client.invoke<DeleteButterBoxResponse>(ctx, 'ButterBoxService',
          'DeleteButterBox', request, DeleteButterBoxResponse());

  /// SetButterBoxToken sets, rotates, or clears (empty token) the box access
  /// token. The token is write-only: encrypted at rest, never read back.
  $async.Future<SetButterBoxTokenResponse> setButterBoxToken(
          $pb.ClientContext? ctx, SetButterBoxTokenRequest request) =>
      _client.invoke<SetButterBoxTokenResponse>(ctx, 'ButterBoxService',
          'SetButterBoxToken', request, SetButterBoxTokenResponse());

  /// GetButterBoxStatus probes the box and reports reachability and the
  /// number of active pi sessions.
  $async.Future<GetButterBoxStatusResponse> getButterBoxStatus(
          $pb.ClientContext? ctx, GetButterBoxStatusRequest request) =>
      _client.invoke<GetButterBoxStatusResponse>(ctx, 'ButterBoxService',
          'GetButterBoxStatus', request, GetButterBoxStatusResponse());

  /// ListButterBoxModels reports the box's pi model catalog (the models pi is
  /// configured with on the box).
  $async.Future<ListButterBoxModelsResponse> listButterBoxModels(
          $pb.ClientContext? ctx, ListButterBoxModelsRequest request) =>
      _client.invoke<ListButterBoxModelsResponse>(ctx, 'ButterBoxService',
          'ListButterBoxModels', request, ListButterBoxModelsResponse());

  /// ListButterBoxCursorModels reports the box's Cursor model catalog (from
  /// the box's CursorService, which asks Cursor's API with the box's key).
  $async.Future<ListButterBoxCursorModelsResponse> listButterBoxCursorModels(
          $pb.ClientContext? ctx, ListButterBoxCursorModelsRequest request) =>
      _client.invoke<ListButterBoxCursorModelsResponse>(
          ctx,
          'ButterBoxService',
          'ListButterBoxCursorModels',
          request,
          ListButterBoxCursorModelsResponse());
}

const $core.bool _omitFieldNames =
    $core.bool.fromEnvironment('protobuf.omit_field_names');
const $core.bool _omitMessageNames =
    $core.bool.fromEnvironment('protobuf.omit_message_names');
