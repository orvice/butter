// This is a generated file - do not edit.
//
// Generated from agents/v1/githost.proto.

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

import 'githost.pbenum.dart';

export 'package:protobuf/protobuf.dart' show GeneratedMessageGenericExtensions;

export 'githost.pbenum.dart';

class GitHost extends $pb.GeneratedMessage {
  factory GitHost({
    $core.String? id,
    $core.String? name,
    GitHostKind? kind,
    $core.String? apiBaseUrl,
    $core.String? webBaseUrl,
    $0.Timestamp? createdAt,
    $0.Timestamp? updatedAt,
  }) {
    final result = GitHost._();
    if (id != null) result.id = id;
    if (name != null) result.name = name;
    if (kind != null) result.kind = kind;
    if (apiBaseUrl != null) result.apiBaseUrl = apiBaseUrl;
    if (webBaseUrl != null) result.webBaseUrl = webBaseUrl;
    if (createdAt != null) result.createdAt = createdAt;
    if (updatedAt != null) result.updatedAt = updatedAt;
    return result;
  }

  GitHost._();

  factory GitHost.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GitHost()..mergeFromBuffer(data, registry);
  factory GitHost.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GitHost()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GitHost',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: GitHost.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..aOS(2, _omitFieldNames ? '' : 'name')
    ..aE<GitHostKind>(3, _omitFieldNames ? '' : 'kind',
        enumValues: GitHostKind.values)
    ..aOS(4, _omitFieldNames ? '' : 'apiBaseUrl')
    ..aOS(5, _omitFieldNames ? '' : 'webBaseUrl')
    ..aOM<$0.Timestamp>(6, _omitFieldNames ? '' : 'createdAt',
        subBuilder: $0.Timestamp.$_createMessage)
    ..aOM<$0.Timestamp>(7, _omitFieldNames ? '' : 'updatedAt',
        subBuilder: $0.Timestamp.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GitHost clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GitHost copyWith(void Function(GitHost) updates) =>
      super.copyWith((message) => updates(message as GitHost)) as GitHost;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use GitHost() / GitHost.new instead')
  static GitHost create() => GitHost._();
  static $pb.GeneratedMessage $_createMessage() => GitHost._();
  @$core.override
  GitHost createEmptyInstance() => GitHost._();
  @$core.pragma('dart2js:noInline')
  static GitHost getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<GitHost>(GitHost.$_createMessage);
  static GitHost? _defaultInstance;

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

  /// Human-readable name shown when selecting a host (e.g. "GitHub.com").
  /// @gotags: json:"name,omitempty" yaml:"name,omitempty" bson:"name,omitempty"
  @$pb.TagNumber(2)
  $core.String get name => $_getSZ(1);
  @$pb.TagNumber(2)
  set name($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasName() => $_has(1);
  @$pb.TagNumber(2)
  void clearName() => $_clearField(2);

  /// Provider adapter used for this host.
  /// @gotags: json:"kind,omitempty" yaml:"kind,omitempty" bson:"kind,omitempty"
  @$pb.TagNumber(3)
  GitHostKind get kind => $_getN(2);
  @$pb.TagNumber(3)
  set kind(GitHostKind value) => $_setField(3, value);
  @$pb.TagNumber(3)
  $core.bool hasKind() => $_has(2);
  @$pb.TagNumber(3)
  void clearKind() => $_clearField(3);

  /// Full API root the adapter joins request paths onto, e.g.
  /// "https://api.github.com", "https://ghe.example.com/api/v3",
  /// "https://gitlab.com/api/v4", "https://gitlab.example.com/api/v4".
  /// @gotags: json:"api_base_url,omitempty" yaml:"api_base_url,omitempty" bson:"api_base_url,omitempty"
  @$pb.TagNumber(4)
  $core.String get apiBaseUrl => $_getSZ(3);
  @$pb.TagNumber(4)
  set apiBaseUrl($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasApiBaseUrl() => $_has(3);
  @$pb.TagNumber(4)
  void clearApiBaseUrl() => $_clearField(4);

  /// Optional web UI base used for building repository links in dashboards,
  /// e.g. "https://github.com". Never used for API calls.
  /// @gotags: json:"web_base_url,omitempty" yaml:"web_base_url,omitempty" bson:"web_base_url,omitempty"
  @$pb.TagNumber(5)
  $core.String get webBaseUrl => $_getSZ(4);
  @$pb.TagNumber(5)
  set webBaseUrl($core.String value) => $_setString(4, value);
  @$pb.TagNumber(5)
  $core.bool hasWebBaseUrl() => $_has(4);
  @$pb.TagNumber(5)
  void clearWebBaseUrl() => $_clearField(5);

  /// @gotags: json:"created_at,omitempty" yaml:"created_at,omitempty" bson:"created_at,omitempty"
  @$pb.TagNumber(6)
  $0.Timestamp get createdAt => $_getN(5);
  @$pb.TagNumber(6)
  set createdAt($0.Timestamp value) => $_setField(6, value);
  @$pb.TagNumber(6)
  $core.bool hasCreatedAt() => $_has(5);
  @$pb.TagNumber(6)
  void clearCreatedAt() => $_clearField(6);
  @$pb.TagNumber(6)
  $0.Timestamp ensureCreatedAt() => $_ensure(5);

  /// @gotags: json:"updated_at,omitempty" yaml:"updated_at,omitempty" bson:"updated_at,omitempty"
  @$pb.TagNumber(7)
  $0.Timestamp get updatedAt => $_getN(6);
  @$pb.TagNumber(7)
  set updatedAt($0.Timestamp value) => $_setField(7, value);
  @$pb.TagNumber(7)
  $core.bool hasUpdatedAt() => $_has(6);
  @$pb.TagNumber(7)
  void clearUpdatedAt() => $_clearField(7);
  @$pb.TagNumber(7)
  $0.Timestamp ensureUpdatedAt() => $_ensure(6);
}

class ListGitHostsRequest extends $pb.GeneratedMessage {
  factory ListGitHostsRequest() => ListGitHostsRequest._();

  ListGitHostsRequest._();

  factory ListGitHostsRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListGitHostsRequest()..mergeFromBuffer(data, registry);
  factory ListGitHostsRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListGitHostsRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListGitHostsRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: ListGitHostsRequest.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListGitHostsRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListGitHostsRequest copyWith(void Function(ListGitHostsRequest) updates) =>
      super.copyWith((message) => updates(message as ListGitHostsRequest))
          as ListGitHostsRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core
      .Deprecated('Use ListGitHostsRequest() / ListGitHostsRequest.new instead')
  static ListGitHostsRequest create() => ListGitHostsRequest._();
  static $pb.GeneratedMessage $_createMessage() => ListGitHostsRequest._();
  @$core.override
  ListGitHostsRequest createEmptyInstance() => ListGitHostsRequest._();
  @$core.pragma('dart2js:noInline')
  static ListGitHostsRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ListGitHostsRequest>(
          ListGitHostsRequest.$_createMessage);
  static ListGitHostsRequest? _defaultInstance;
}

class ListGitHostsResponse extends $pb.GeneratedMessage {
  factory ListGitHostsResponse({
    $core.Iterable<GitHost>? hosts,
  }) {
    final result = ListGitHostsResponse._();
    if (hosts != null) result.hosts.addAll(hosts);
    return result;
  }

  ListGitHostsResponse._();

  factory ListGitHostsResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListGitHostsResponse()..mergeFromBuffer(data, registry);
  factory ListGitHostsResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListGitHostsResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListGitHostsResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: ListGitHostsResponse.$_createMessage)
    ..pPM<GitHost>(1, _omitFieldNames ? '' : 'hosts',
        subBuilder: GitHost.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListGitHostsResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListGitHostsResponse copyWith(void Function(ListGitHostsResponse) updates) =>
      super.copyWith((message) => updates(message as ListGitHostsResponse))
          as ListGitHostsResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use ListGitHostsResponse() / ListGitHostsResponse.new instead')
  static ListGitHostsResponse create() => ListGitHostsResponse._();
  static $pb.GeneratedMessage $_createMessage() => ListGitHostsResponse._();
  @$core.override
  ListGitHostsResponse createEmptyInstance() => ListGitHostsResponse._();
  @$core.pragma('dart2js:noInline')
  static ListGitHostsResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ListGitHostsResponse>(
          ListGitHostsResponse.$_createMessage);
  static ListGitHostsResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $pb.PbList<GitHost> get hosts => $_getList(0);
}

class GetGitHostRequest extends $pb.GeneratedMessage {
  factory GetGitHostRequest({
    $core.String? id,
  }) {
    final result = GetGitHostRequest._();
    if (id != null) result.id = id;
    return result;
  }

  GetGitHostRequest._();

  factory GetGitHostRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetGitHostRequest()..mergeFromBuffer(data, registry);
  factory GetGitHostRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetGitHostRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetGitHostRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: GetGitHostRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetGitHostRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetGitHostRequest copyWith(void Function(GetGitHostRequest) updates) =>
      super.copyWith((message) => updates(message as GetGitHostRequest))
          as GetGitHostRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use GetGitHostRequest() / GetGitHostRequest.new instead')
  static GetGitHostRequest create() => GetGitHostRequest._();
  static $pb.GeneratedMessage $_createMessage() => GetGitHostRequest._();
  @$core.override
  GetGitHostRequest createEmptyInstance() => GetGitHostRequest._();
  @$core.pragma('dart2js:noInline')
  static GetGitHostRequest getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<GetGitHostRequest>(
          GetGitHostRequest.$_createMessage);
  static GetGitHostRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);
}

class GetGitHostResponse extends $pb.GeneratedMessage {
  factory GetGitHostResponse({
    GitHost? host,
  }) {
    final result = GetGitHostResponse._();
    if (host != null) result.host = host;
    return result;
  }

  GetGitHostResponse._();

  factory GetGitHostResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetGitHostResponse()..mergeFromBuffer(data, registry);
  factory GetGitHostResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetGitHostResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetGitHostResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: GetGitHostResponse.$_createMessage)
    ..aOM<GitHost>(1, _omitFieldNames ? '' : 'host',
        subBuilder: GitHost.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetGitHostResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetGitHostResponse copyWith(void Function(GetGitHostResponse) updates) =>
      super.copyWith((message) => updates(message as GetGitHostResponse))
          as GetGitHostResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use GetGitHostResponse() / GetGitHostResponse.new instead')
  static GetGitHostResponse create() => GetGitHostResponse._();
  static $pb.GeneratedMessage $_createMessage() => GetGitHostResponse._();
  @$core.override
  GetGitHostResponse createEmptyInstance() => GetGitHostResponse._();
  @$core.pragma('dart2js:noInline')
  static GetGitHostResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<GetGitHostResponse>(
          GetGitHostResponse.$_createMessage);
  static GetGitHostResponse? _defaultInstance;

  @$pb.TagNumber(1)
  GitHost get host => $_getN(0);
  @$pb.TagNumber(1)
  set host(GitHost value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasHost() => $_has(0);
  @$pb.TagNumber(1)
  void clearHost() => $_clearField(1);
  @$pb.TagNumber(1)
  GitHost ensureHost() => $_ensure(0);
}

class CreateGitHostRequest extends $pb.GeneratedMessage {
  factory CreateGitHostRequest({
    GitHost? host,
  }) {
    final result = CreateGitHostRequest._();
    if (host != null) result.host = host;
    return result;
  }

  CreateGitHostRequest._();

  factory CreateGitHostRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateGitHostRequest()..mergeFromBuffer(data, registry);
  factory CreateGitHostRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateGitHostRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CreateGitHostRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: CreateGitHostRequest.$_createMessage)
    ..aOM<GitHost>(1, _omitFieldNames ? '' : 'host',
        subBuilder: GitHost.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateGitHostRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateGitHostRequest copyWith(void Function(CreateGitHostRequest) updates) =>
      super.copyWith((message) => updates(message as CreateGitHostRequest))
          as CreateGitHostRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use CreateGitHostRequest() / CreateGitHostRequest.new instead')
  static CreateGitHostRequest create() => CreateGitHostRequest._();
  static $pb.GeneratedMessage $_createMessage() => CreateGitHostRequest._();
  @$core.override
  CreateGitHostRequest createEmptyInstance() => CreateGitHostRequest._();
  @$core.pragma('dart2js:noInline')
  static CreateGitHostRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<CreateGitHostRequest>(
          CreateGitHostRequest.$_createMessage);
  static CreateGitHostRequest? _defaultInstance;

  @$pb.TagNumber(1)
  GitHost get host => $_getN(0);
  @$pb.TagNumber(1)
  set host(GitHost value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasHost() => $_has(0);
  @$pb.TagNumber(1)
  void clearHost() => $_clearField(1);
  @$pb.TagNumber(1)
  GitHost ensureHost() => $_ensure(0);
}

class CreateGitHostResponse extends $pb.GeneratedMessage {
  factory CreateGitHostResponse({
    GitHost? host,
  }) {
    final result = CreateGitHostResponse._();
    if (host != null) result.host = host;
    return result;
  }

  CreateGitHostResponse._();

  factory CreateGitHostResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateGitHostResponse()..mergeFromBuffer(data, registry);
  factory CreateGitHostResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateGitHostResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CreateGitHostResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: CreateGitHostResponse.$_createMessage)
    ..aOM<GitHost>(1, _omitFieldNames ? '' : 'host',
        subBuilder: GitHost.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateGitHostResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateGitHostResponse copyWith(
          void Function(CreateGitHostResponse) updates) =>
      super.copyWith((message) => updates(message as CreateGitHostResponse))
          as CreateGitHostResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use CreateGitHostResponse() / CreateGitHostResponse.new instead')
  static CreateGitHostResponse create() => CreateGitHostResponse._();
  static $pb.GeneratedMessage $_createMessage() => CreateGitHostResponse._();
  @$core.override
  CreateGitHostResponse createEmptyInstance() => CreateGitHostResponse._();
  @$core.pragma('dart2js:noInline')
  static CreateGitHostResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<CreateGitHostResponse>(
          CreateGitHostResponse.$_createMessage);
  static CreateGitHostResponse? _defaultInstance;

  @$pb.TagNumber(1)
  GitHost get host => $_getN(0);
  @$pb.TagNumber(1)
  set host(GitHost value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasHost() => $_has(0);
  @$pb.TagNumber(1)
  void clearHost() => $_clearField(1);
  @$pb.TagNumber(1)
  GitHost ensureHost() => $_ensure(0);
}

class UpdateGitHostRequest extends $pb.GeneratedMessage {
  factory UpdateGitHostRequest({
    GitHost? host,
  }) {
    final result = UpdateGitHostRequest._();
    if (host != null) result.host = host;
    return result;
  }

  UpdateGitHostRequest._();

  factory UpdateGitHostRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpdateGitHostRequest()..mergeFromBuffer(data, registry);
  factory UpdateGitHostRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpdateGitHostRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'UpdateGitHostRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: UpdateGitHostRequest.$_createMessage)
    ..aOM<GitHost>(1, _omitFieldNames ? '' : 'host',
        subBuilder: GitHost.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateGitHostRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateGitHostRequest copyWith(void Function(UpdateGitHostRequest) updates) =>
      super.copyWith((message) => updates(message as UpdateGitHostRequest))
          as UpdateGitHostRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use UpdateGitHostRequest() / UpdateGitHostRequest.new instead')
  static UpdateGitHostRequest create() => UpdateGitHostRequest._();
  static $pb.GeneratedMessage $_createMessage() => UpdateGitHostRequest._();
  @$core.override
  UpdateGitHostRequest createEmptyInstance() => UpdateGitHostRequest._();
  @$core.pragma('dart2js:noInline')
  static UpdateGitHostRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<UpdateGitHostRequest>(
          UpdateGitHostRequest.$_createMessage);
  static UpdateGitHostRequest? _defaultInstance;

  @$pb.TagNumber(1)
  GitHost get host => $_getN(0);
  @$pb.TagNumber(1)
  set host(GitHost value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasHost() => $_has(0);
  @$pb.TagNumber(1)
  void clearHost() => $_clearField(1);
  @$pb.TagNumber(1)
  GitHost ensureHost() => $_ensure(0);
}

class UpdateGitHostResponse extends $pb.GeneratedMessage {
  factory UpdateGitHostResponse({
    GitHost? host,
  }) {
    final result = UpdateGitHostResponse._();
    if (host != null) result.host = host;
    return result;
  }

  UpdateGitHostResponse._();

  factory UpdateGitHostResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpdateGitHostResponse()..mergeFromBuffer(data, registry);
  factory UpdateGitHostResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpdateGitHostResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'UpdateGitHostResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: UpdateGitHostResponse.$_createMessage)
    ..aOM<GitHost>(1, _omitFieldNames ? '' : 'host',
        subBuilder: GitHost.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateGitHostResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateGitHostResponse copyWith(
          void Function(UpdateGitHostResponse) updates) =>
      super.copyWith((message) => updates(message as UpdateGitHostResponse))
          as UpdateGitHostResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use UpdateGitHostResponse() / UpdateGitHostResponse.new instead')
  static UpdateGitHostResponse create() => UpdateGitHostResponse._();
  static $pb.GeneratedMessage $_createMessage() => UpdateGitHostResponse._();
  @$core.override
  UpdateGitHostResponse createEmptyInstance() => UpdateGitHostResponse._();
  @$core.pragma('dart2js:noInline')
  static UpdateGitHostResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<UpdateGitHostResponse>(
          UpdateGitHostResponse.$_createMessage);
  static UpdateGitHostResponse? _defaultInstance;

  @$pb.TagNumber(1)
  GitHost get host => $_getN(0);
  @$pb.TagNumber(1)
  set host(GitHost value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasHost() => $_has(0);
  @$pb.TagNumber(1)
  void clearHost() => $_clearField(1);
  @$pb.TagNumber(1)
  GitHost ensureHost() => $_ensure(0);
}

class DeleteGitHostRequest extends $pb.GeneratedMessage {
  factory DeleteGitHostRequest({
    $core.String? id,
  }) {
    final result = DeleteGitHostRequest._();
    if (id != null) result.id = id;
    return result;
  }

  DeleteGitHostRequest._();

  factory DeleteGitHostRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DeleteGitHostRequest()..mergeFromBuffer(data, registry);
  factory DeleteGitHostRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DeleteGitHostRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'DeleteGitHostRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: DeleteGitHostRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteGitHostRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteGitHostRequest copyWith(void Function(DeleteGitHostRequest) updates) =>
      super.copyWith((message) => updates(message as DeleteGitHostRequest))
          as DeleteGitHostRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use DeleteGitHostRequest() / DeleteGitHostRequest.new instead')
  static DeleteGitHostRequest create() => DeleteGitHostRequest._();
  static $pb.GeneratedMessage $_createMessage() => DeleteGitHostRequest._();
  @$core.override
  DeleteGitHostRequest createEmptyInstance() => DeleteGitHostRequest._();
  @$core.pragma('dart2js:noInline')
  static DeleteGitHostRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<DeleteGitHostRequest>(
          DeleteGitHostRequest.$_createMessage);
  static DeleteGitHostRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);
}

class DeleteGitHostResponse extends $pb.GeneratedMessage {
  factory DeleteGitHostResponse() => DeleteGitHostResponse._();

  DeleteGitHostResponse._();

  factory DeleteGitHostResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DeleteGitHostResponse()..mergeFromBuffer(data, registry);
  factory DeleteGitHostResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DeleteGitHostResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'DeleteGitHostResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: DeleteGitHostResponse.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteGitHostResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteGitHostResponse copyWith(
          void Function(DeleteGitHostResponse) updates) =>
      super.copyWith((message) => updates(message as DeleteGitHostResponse))
          as DeleteGitHostResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use DeleteGitHostResponse() / DeleteGitHostResponse.new instead')
  static DeleteGitHostResponse create() => DeleteGitHostResponse._();
  static $pb.GeneratedMessage $_createMessage() => DeleteGitHostResponse._();
  @$core.override
  DeleteGitHostResponse createEmptyInstance() => DeleteGitHostResponse._();
  @$core.pragma('dart2js:noInline')
  static DeleteGitHostResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<DeleteGitHostResponse>(
          DeleteGitHostResponse.$_createMessage);
  static DeleteGitHostResponse? _defaultInstance;
}

/// GitHostService manages the platform-level allowlist of Git hosts.
/// List/Get are available to any authenticated user (workspace owners need
/// them to configure a binding); Create/Update/Delete require global admin.
class GitHostServiceApi {
  final $pb.RpcClient _client;

  GitHostServiceApi(this._client);

  $async.Future<ListGitHostsResponse> listGitHosts(
          $pb.ClientContext? ctx, ListGitHostsRequest request) =>
      _client.invoke<ListGitHostsResponse>(ctx, 'GitHostService',
          'ListGitHosts', request, ListGitHostsResponse());
  $async.Future<GetGitHostResponse> getGitHost(
          $pb.ClientContext? ctx, GetGitHostRequest request) =>
      _client.invoke<GetGitHostResponse>(
          ctx, 'GitHostService', 'GetGitHost', request, GetGitHostResponse());
  $async.Future<CreateGitHostResponse> createGitHost(
          $pb.ClientContext? ctx, CreateGitHostRequest request) =>
      _client.invoke<CreateGitHostResponse>(ctx, 'GitHostService',
          'CreateGitHost', request, CreateGitHostResponse());
  $async.Future<UpdateGitHostResponse> updateGitHost(
          $pb.ClientContext? ctx, UpdateGitHostRequest request) =>
      _client.invoke<UpdateGitHostResponse>(ctx, 'GitHostService',
          'UpdateGitHost', request, UpdateGitHostResponse());
  $async.Future<DeleteGitHostResponse> deleteGitHost(
          $pb.ClientContext? ctx, DeleteGitHostRequest request) =>
      _client.invoke<DeleteGitHostResponse>(ctx, 'GitHostService',
          'DeleteGitHost', request, DeleteGitHostResponse());
}

const $core.bool _omitFieldNames =
    $core.bool.fromEnvironment('protobuf.omit_field_names');
const $core.bool _omitMessageNames =
    $core.bool.fromEnvironment('protobuf.omit_message_names');
