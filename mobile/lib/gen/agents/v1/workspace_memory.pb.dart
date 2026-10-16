// This is a generated file - do not edit.
//
// Generated from agents/v1/workspace_memory.proto.

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

import 'workspace_memory.pbenum.dart';

export 'package:protobuf/protobuf.dart' show GeneratedMessageGenericExtensions;

export 'workspace_memory.pbenum.dart';

class WorkspaceMemoryConfig extends $pb.GeneratedMessage {
  factory WorkspaceMemoryConfig({
    $core.String? workspaceId,
    $core.String? baseUrl,
    $core.bool? enabled,
    $core.bool? credentialSet,
    $0.Timestamp? credentialUpdatedAt,
    $0.Timestamp? createdAt,
    $0.Timestamp? updatedAt,
  }) {
    final result = WorkspaceMemoryConfig._();
    if (workspaceId != null) result.workspaceId = workspaceId;
    if (baseUrl != null) result.baseUrl = baseUrl;
    if (enabled != null) result.enabled = enabled;
    if (credentialSet != null) result.credentialSet = credentialSet;
    if (credentialUpdatedAt != null)
      result.credentialUpdatedAt = credentialUpdatedAt;
    if (createdAt != null) result.createdAt = createdAt;
    if (updatedAt != null) result.updatedAt = updatedAt;
    return result;
  }

  WorkspaceMemoryConfig._();

  factory WorkspaceMemoryConfig.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      WorkspaceMemoryConfig()..mergeFromBuffer(data, registry);
  factory WorkspaceMemoryConfig.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      WorkspaceMemoryConfig()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'WorkspaceMemoryConfig',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: WorkspaceMemoryConfig.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'workspaceId')
    ..aOS(2, _omitFieldNames ? '' : 'baseUrl')
    ..aOB(3, _omitFieldNames ? '' : 'enabled')
    ..aOB(4, _omitFieldNames ? '' : 'credentialSet')
    ..aOM<$0.Timestamp>(5, _omitFieldNames ? '' : 'credentialUpdatedAt',
        subBuilder: $0.Timestamp.$_createMessage)
    ..aOM<$0.Timestamp>(6, _omitFieldNames ? '' : 'createdAt',
        subBuilder: $0.Timestamp.$_createMessage)
    ..aOM<$0.Timestamp>(7, _omitFieldNames ? '' : 'updatedAt',
        subBuilder: $0.Timestamp.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  WorkspaceMemoryConfig clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  WorkspaceMemoryConfig copyWith(
          void Function(WorkspaceMemoryConfig) updates) =>
      super.copyWith((message) => updates(message as WorkspaceMemoryConfig))
          as WorkspaceMemoryConfig;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use WorkspaceMemoryConfig() / WorkspaceMemoryConfig.new instead')
  static WorkspaceMemoryConfig create() => WorkspaceMemoryConfig._();
  static $pb.GeneratedMessage $_createMessage() => WorkspaceMemoryConfig._();
  @$core.override
  WorkspaceMemoryConfig createEmptyInstance() => WorkspaceMemoryConfig._();
  @$core.pragma('dart2js:noInline')
  static WorkspaceMemoryConfig getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<WorkspaceMemoryConfig>(
          WorkspaceMemoryConfig.$_createMessage);
  static WorkspaceMemoryConfig? _defaultInstance;

  /// ID of the workspace this config belongs to. Stamped by the repository
  /// from request context; never taken from client input.
  /// @gotags: json:"workspace_id,omitempty" yaml:"workspace_id,omitempty" bson:"workspace_id,omitempty"
  @$pb.TagNumber(1)
  $core.String get workspaceId => $_getSZ(0);
  @$pb.TagNumber(1)
  set workspaceId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasWorkspaceId() => $_has(0);
  @$pb.TagNumber(1)
  void clearWorkspaceId() => $_clearField(1);

  /// Base URL of the mem0 OSS REST server, e.g. "https://mem0.example.com".
  /// @gotags: json:"base_url,omitempty" yaml:"base_url,omitempty" bson:"base_url,omitempty"
  @$pb.TagNumber(2)
  $core.String get baseUrl => $_getSZ(1);
  @$pb.TagNumber(2)
  set baseUrl($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasBaseUrl() => $_has(1);
  @$pb.TagNumber(2)
  void clearBaseUrl() => $_clearField(2);

  /// When false, memory-enabled agents in this workspace run without memory.
  /// @gotags: json:"enabled,omitempty" yaml:"enabled,omitempty" bson:"enabled,omitempty"
  @$pb.TagNumber(3)
  $core.bool get enabled => $_getBF(2);
  @$pb.TagNumber(3)
  set enabled($core.bool value) => $_setBool(2, value);
  @$pb.TagNumber(3)
  $core.bool hasEnabled() => $_has(2);
  @$pb.TagNumber(3)
  void clearEnabled() => $_clearField(3);

  /// Reports whether a mem0 API key is stored. Derived from the credential
  /// columns on every read; never taken from client input.
  /// @gotags: json:"credential_set,omitempty" yaml:"credential_set,omitempty" bson:"credential_set,omitempty"
  @$pb.TagNumber(4)
  $core.bool get credentialSet => $_getBF(3);
  @$pb.TagNumber(4)
  set credentialSet($core.bool value) => $_setBool(3, value);
  @$pb.TagNumber(4)
  $core.bool hasCredentialSet() => $_has(3);
  @$pb.TagNumber(4)
  void clearCredentialSet() => $_clearField(4);

  /// When the API key was last set or rotated. Unset when no key is stored.
  /// @gotags: json:"credential_updated_at,omitempty" yaml:"credential_updated_at,omitempty" bson:"credential_updated_at,omitempty"
  @$pb.TagNumber(5)
  $0.Timestamp get credentialUpdatedAt => $_getN(4);
  @$pb.TagNumber(5)
  set credentialUpdatedAt($0.Timestamp value) => $_setField(5, value);
  @$pb.TagNumber(5)
  $core.bool hasCredentialUpdatedAt() => $_has(4);
  @$pb.TagNumber(5)
  void clearCredentialUpdatedAt() => $_clearField(5);
  @$pb.TagNumber(5)
  $0.Timestamp ensureCredentialUpdatedAt() => $_ensure(4);

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

class GetWorkspaceMemoryConfigRequest extends $pb.GeneratedMessage {
  factory GetWorkspaceMemoryConfigRequest() =>
      GetWorkspaceMemoryConfigRequest._();

  GetWorkspaceMemoryConfigRequest._();

  factory GetWorkspaceMemoryConfigRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetWorkspaceMemoryConfigRequest()..mergeFromBuffer(data, registry);
  factory GetWorkspaceMemoryConfigRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetWorkspaceMemoryConfigRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetWorkspaceMemoryConfigRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: GetWorkspaceMemoryConfigRequest.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetWorkspaceMemoryConfigRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetWorkspaceMemoryConfigRequest copyWith(
          void Function(GetWorkspaceMemoryConfigRequest) updates) =>
      super.copyWith(
              (message) => updates(message as GetWorkspaceMemoryConfigRequest))
          as GetWorkspaceMemoryConfigRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use GetWorkspaceMemoryConfigRequest() / GetWorkspaceMemoryConfigRequest.new instead')
  static GetWorkspaceMemoryConfigRequest create() =>
      GetWorkspaceMemoryConfigRequest._();
  static $pb.GeneratedMessage $_createMessage() =>
      GetWorkspaceMemoryConfigRequest._();
  @$core.override
  GetWorkspaceMemoryConfigRequest createEmptyInstance() =>
      GetWorkspaceMemoryConfigRequest._();
  @$core.pragma('dart2js:noInline')
  static GetWorkspaceMemoryConfigRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<GetWorkspaceMemoryConfigRequest>(
          GetWorkspaceMemoryConfigRequest.$_createMessage);
  static GetWorkspaceMemoryConfigRequest? _defaultInstance;
}

class GetWorkspaceMemoryConfigResponse extends $pb.GeneratedMessage {
  factory GetWorkspaceMemoryConfigResponse({
    WorkspaceMemoryConfig? config,
  }) {
    final result = GetWorkspaceMemoryConfigResponse._();
    if (config != null) result.config = config;
    return result;
  }

  GetWorkspaceMemoryConfigResponse._();

  factory GetWorkspaceMemoryConfigResponse.fromBuffer(
          $core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetWorkspaceMemoryConfigResponse()..mergeFromBuffer(data, registry);
  factory GetWorkspaceMemoryConfigResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetWorkspaceMemoryConfigResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetWorkspaceMemoryConfigResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: GetWorkspaceMemoryConfigResponse.$_createMessage)
    ..aOM<WorkspaceMemoryConfig>(1, _omitFieldNames ? '' : 'config',
        subBuilder: WorkspaceMemoryConfig.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetWorkspaceMemoryConfigResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetWorkspaceMemoryConfigResponse copyWith(
          void Function(GetWorkspaceMemoryConfigResponse) updates) =>
      super.copyWith(
              (message) => updates(message as GetWorkspaceMemoryConfigResponse))
          as GetWorkspaceMemoryConfigResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use GetWorkspaceMemoryConfigResponse() / GetWorkspaceMemoryConfigResponse.new instead')
  static GetWorkspaceMemoryConfigResponse create() =>
      GetWorkspaceMemoryConfigResponse._();
  static $pb.GeneratedMessage $_createMessage() =>
      GetWorkspaceMemoryConfigResponse._();
  @$core.override
  GetWorkspaceMemoryConfigResponse createEmptyInstance() =>
      GetWorkspaceMemoryConfigResponse._();
  @$core.pragma('dart2js:noInline')
  static GetWorkspaceMemoryConfigResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<GetWorkspaceMemoryConfigResponse>(
          GetWorkspaceMemoryConfigResponse.$_createMessage);
  static GetWorkspaceMemoryConfigResponse? _defaultInstance;

  /// Unset when the workspace has no config.
  @$pb.TagNumber(1)
  WorkspaceMemoryConfig get config => $_getN(0);
  @$pb.TagNumber(1)
  set config(WorkspaceMemoryConfig value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasConfig() => $_has(0);
  @$pb.TagNumber(1)
  void clearConfig() => $_clearField(1);
  @$pb.TagNumber(1)
  WorkspaceMemoryConfig ensureConfig() => $_ensure(0);
}

class PutWorkspaceMemoryConfigRequest extends $pb.GeneratedMessage {
  factory PutWorkspaceMemoryConfigRequest({
    $core.String? baseUrl,
    $core.bool? enabled,
    $core.String? apiKey,
  }) {
    final result = PutWorkspaceMemoryConfigRequest._();
    if (baseUrl != null) result.baseUrl = baseUrl;
    if (enabled != null) result.enabled = enabled;
    if (apiKey != null) result.apiKey = apiKey;
    return result;
  }

  PutWorkspaceMemoryConfigRequest._();

  factory PutWorkspaceMemoryConfigRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      PutWorkspaceMemoryConfigRequest()..mergeFromBuffer(data, registry);
  factory PutWorkspaceMemoryConfigRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      PutWorkspaceMemoryConfigRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'PutWorkspaceMemoryConfigRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: PutWorkspaceMemoryConfigRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'baseUrl')
    ..aOB(2, _omitFieldNames ? '' : 'enabled')
    ..aOS(3, _omitFieldNames ? '' : 'apiKey')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  PutWorkspaceMemoryConfigRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  PutWorkspaceMemoryConfigRequest copyWith(
          void Function(PutWorkspaceMemoryConfigRequest) updates) =>
      super.copyWith(
              (message) => updates(message as PutWorkspaceMemoryConfigRequest))
          as PutWorkspaceMemoryConfigRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use PutWorkspaceMemoryConfigRequest() / PutWorkspaceMemoryConfigRequest.new instead')
  static PutWorkspaceMemoryConfigRequest create() =>
      PutWorkspaceMemoryConfigRequest._();
  static $pb.GeneratedMessage $_createMessage() =>
      PutWorkspaceMemoryConfigRequest._();
  @$core.override
  PutWorkspaceMemoryConfigRequest createEmptyInstance() =>
      PutWorkspaceMemoryConfigRequest._();
  @$core.pragma('dart2js:noInline')
  static PutWorkspaceMemoryConfigRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<PutWorkspaceMemoryConfigRequest>(
          PutWorkspaceMemoryConfigRequest.$_createMessage);
  static PutWorkspaceMemoryConfigRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get baseUrl => $_getSZ(0);
  @$pb.TagNumber(1)
  set baseUrl($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasBaseUrl() => $_has(0);
  @$pb.TagNumber(1)
  void clearBaseUrl() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.bool get enabled => $_getBF(1);
  @$pb.TagNumber(2)
  set enabled($core.bool value) => $_setBool(1, value);
  @$pb.TagNumber(2)
  $core.bool hasEnabled() => $_has(1);
  @$pb.TagNumber(2)
  void clearEnabled() => $_clearField(2);

  /// Write-only mem0 API key. Absent keeps the stored key, empty clears it,
  /// any other value sets or rotates it.
  @$pb.TagNumber(3)
  $core.String get apiKey => $_getSZ(2);
  @$pb.TagNumber(3)
  set apiKey($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasApiKey() => $_has(2);
  @$pb.TagNumber(3)
  void clearApiKey() => $_clearField(3);
}

class PutWorkspaceMemoryConfigResponse extends $pb.GeneratedMessage {
  factory PutWorkspaceMemoryConfigResponse({
    WorkspaceMemoryConfig? config,
    $core.String? warning,
  }) {
    final result = PutWorkspaceMemoryConfigResponse._();
    if (config != null) result.config = config;
    if (warning != null) result.warning = warning;
    return result;
  }

  PutWorkspaceMemoryConfigResponse._();

  factory PutWorkspaceMemoryConfigResponse.fromBuffer(
          $core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      PutWorkspaceMemoryConfigResponse()..mergeFromBuffer(data, registry);
  factory PutWorkspaceMemoryConfigResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      PutWorkspaceMemoryConfigResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'PutWorkspaceMemoryConfigResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: PutWorkspaceMemoryConfigResponse.$_createMessage)
    ..aOM<WorkspaceMemoryConfig>(1, _omitFieldNames ? '' : 'config',
        subBuilder: WorkspaceMemoryConfig.$_createMessage)
    ..aOS(2, _omitFieldNames ? '' : 'warning')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  PutWorkspaceMemoryConfigResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  PutWorkspaceMemoryConfigResponse copyWith(
          void Function(PutWorkspaceMemoryConfigResponse) updates) =>
      super.copyWith(
              (message) => updates(message as PutWorkspaceMemoryConfigResponse))
          as PutWorkspaceMemoryConfigResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use PutWorkspaceMemoryConfigResponse() / PutWorkspaceMemoryConfigResponse.new instead')
  static PutWorkspaceMemoryConfigResponse create() =>
      PutWorkspaceMemoryConfigResponse._();
  static $pb.GeneratedMessage $_createMessage() =>
      PutWorkspaceMemoryConfigResponse._();
  @$core.override
  PutWorkspaceMemoryConfigResponse createEmptyInstance() =>
      PutWorkspaceMemoryConfigResponse._();
  @$core.pragma('dart2js:noInline')
  static PutWorkspaceMemoryConfigResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<PutWorkspaceMemoryConfigResponse>(
          PutWorkspaceMemoryConfigResponse.$_createMessage);
  static PutWorkspaceMemoryConfigResponse? _defaultInstance;

  @$pb.TagNumber(1)
  WorkspaceMemoryConfig get config => $_getN(0);
  @$pb.TagNumber(1)
  set config(WorkspaceMemoryConfig value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasConfig() => $_has(0);
  @$pb.TagNumber(1)
  void clearConfig() => $_clearField(1);
  @$pb.TagNumber(1)
  WorkspaceMemoryConfig ensureConfig() => $_ensure(0);

  /// Non-empty when the config was saved but the probe could not reach the
  /// server.
  @$pb.TagNumber(2)
  $core.String get warning => $_getSZ(1);
  @$pb.TagNumber(2)
  set warning($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasWarning() => $_has(1);
  @$pb.TagNumber(2)
  void clearWarning() => $_clearField(2);
}

class DeleteWorkspaceMemoryConfigRequest extends $pb.GeneratedMessage {
  factory DeleteWorkspaceMemoryConfigRequest() =>
      DeleteWorkspaceMemoryConfigRequest._();

  DeleteWorkspaceMemoryConfigRequest._();

  factory DeleteWorkspaceMemoryConfigRequest.fromBuffer(
          $core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DeleteWorkspaceMemoryConfigRequest()..mergeFromBuffer(data, registry);
  factory DeleteWorkspaceMemoryConfigRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DeleteWorkspaceMemoryConfigRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'DeleteWorkspaceMemoryConfigRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: DeleteWorkspaceMemoryConfigRequest.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteWorkspaceMemoryConfigRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteWorkspaceMemoryConfigRequest copyWith(
          void Function(DeleteWorkspaceMemoryConfigRequest) updates) =>
      super.copyWith((message) =>
              updates(message as DeleteWorkspaceMemoryConfigRequest))
          as DeleteWorkspaceMemoryConfigRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use DeleteWorkspaceMemoryConfigRequest() / DeleteWorkspaceMemoryConfigRequest.new instead')
  static DeleteWorkspaceMemoryConfigRequest create() =>
      DeleteWorkspaceMemoryConfigRequest._();
  static $pb.GeneratedMessage $_createMessage() =>
      DeleteWorkspaceMemoryConfigRequest._();
  @$core.override
  DeleteWorkspaceMemoryConfigRequest createEmptyInstance() =>
      DeleteWorkspaceMemoryConfigRequest._();
  @$core.pragma('dart2js:noInline')
  static DeleteWorkspaceMemoryConfigRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<DeleteWorkspaceMemoryConfigRequest>(
          DeleteWorkspaceMemoryConfigRequest.$_createMessage);
  static DeleteWorkspaceMemoryConfigRequest? _defaultInstance;
}

class DeleteWorkspaceMemoryConfigResponse extends $pb.GeneratedMessage {
  factory DeleteWorkspaceMemoryConfigResponse() =>
      DeleteWorkspaceMemoryConfigResponse._();

  DeleteWorkspaceMemoryConfigResponse._();

  factory DeleteWorkspaceMemoryConfigResponse.fromBuffer(
          $core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DeleteWorkspaceMemoryConfigResponse()..mergeFromBuffer(data, registry);
  factory DeleteWorkspaceMemoryConfigResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DeleteWorkspaceMemoryConfigResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'DeleteWorkspaceMemoryConfigResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: DeleteWorkspaceMemoryConfigResponse.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteWorkspaceMemoryConfigResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteWorkspaceMemoryConfigResponse copyWith(
          void Function(DeleteWorkspaceMemoryConfigResponse) updates) =>
      super.copyWith((message) =>
              updates(message as DeleteWorkspaceMemoryConfigResponse))
          as DeleteWorkspaceMemoryConfigResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use DeleteWorkspaceMemoryConfigResponse() / DeleteWorkspaceMemoryConfigResponse.new instead')
  static DeleteWorkspaceMemoryConfigResponse create() =>
      DeleteWorkspaceMemoryConfigResponse._();
  static $pb.GeneratedMessage $_createMessage() =>
      DeleteWorkspaceMemoryConfigResponse._();
  @$core.override
  DeleteWorkspaceMemoryConfigResponse createEmptyInstance() =>
      DeleteWorkspaceMemoryConfigResponse._();
  @$core.pragma('dart2js:noInline')
  static DeleteWorkspaceMemoryConfigResponse getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<
              DeleteWorkspaceMemoryConfigResponse>(
          DeleteWorkspaceMemoryConfigResponse.$_createMessage);
  static DeleteWorkspaceMemoryConfigResponse? _defaultInstance;
}

class TestWorkspaceMemoryConnectionRequest extends $pb.GeneratedMessage {
  factory TestWorkspaceMemoryConnectionRequest() =>
      TestWorkspaceMemoryConnectionRequest._();

  TestWorkspaceMemoryConnectionRequest._();

  factory TestWorkspaceMemoryConnectionRequest.fromBuffer(
          $core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      TestWorkspaceMemoryConnectionRequest()..mergeFromBuffer(data, registry);
  factory TestWorkspaceMemoryConnectionRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      TestWorkspaceMemoryConnectionRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'TestWorkspaceMemoryConnectionRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: TestWorkspaceMemoryConnectionRequest.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  TestWorkspaceMemoryConnectionRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  TestWorkspaceMemoryConnectionRequest copyWith(
          void Function(TestWorkspaceMemoryConnectionRequest) updates) =>
      super.copyWith((message) =>
              updates(message as TestWorkspaceMemoryConnectionRequest))
          as TestWorkspaceMemoryConnectionRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use TestWorkspaceMemoryConnectionRequest() / TestWorkspaceMemoryConnectionRequest.new instead')
  static TestWorkspaceMemoryConnectionRequest create() =>
      TestWorkspaceMemoryConnectionRequest._();
  static $pb.GeneratedMessage $_createMessage() =>
      TestWorkspaceMemoryConnectionRequest._();
  @$core.override
  TestWorkspaceMemoryConnectionRequest createEmptyInstance() =>
      TestWorkspaceMemoryConnectionRequest._();
  @$core.pragma('dart2js:noInline')
  static TestWorkspaceMemoryConnectionRequest getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<
              TestWorkspaceMemoryConnectionRequest>(
          TestWorkspaceMemoryConnectionRequest.$_createMessage);
  static TestWorkspaceMemoryConnectionRequest? _defaultInstance;
}

class TestWorkspaceMemoryConnectionResponse extends $pb.GeneratedMessage {
  factory TestWorkspaceMemoryConnectionResponse({
    $core.bool? ok,
    $core.String? error,
  }) {
    final result = TestWorkspaceMemoryConnectionResponse._();
    if (ok != null) result.ok = ok;
    if (error != null) result.error = error;
    return result;
  }

  TestWorkspaceMemoryConnectionResponse._();

  factory TestWorkspaceMemoryConnectionResponse.fromBuffer(
          $core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      TestWorkspaceMemoryConnectionResponse()..mergeFromBuffer(data, registry);
  factory TestWorkspaceMemoryConnectionResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      TestWorkspaceMemoryConnectionResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'TestWorkspaceMemoryConnectionResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance:
          TestWorkspaceMemoryConnectionResponse.$_createMessage)
    ..aOB(1, _omitFieldNames ? '' : 'ok')
    ..aOS(2, _omitFieldNames ? '' : 'error')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  TestWorkspaceMemoryConnectionResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  TestWorkspaceMemoryConnectionResponse copyWith(
          void Function(TestWorkspaceMemoryConnectionResponse) updates) =>
      super.copyWith((message) =>
              updates(message as TestWorkspaceMemoryConnectionResponse))
          as TestWorkspaceMemoryConnectionResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use TestWorkspaceMemoryConnectionResponse() / TestWorkspaceMemoryConnectionResponse.new instead')
  static TestWorkspaceMemoryConnectionResponse create() =>
      TestWorkspaceMemoryConnectionResponse._();
  static $pb.GeneratedMessage $_createMessage() =>
      TestWorkspaceMemoryConnectionResponse._();
  @$core.override
  TestWorkspaceMemoryConnectionResponse createEmptyInstance() =>
      TestWorkspaceMemoryConnectionResponse._();
  @$core.pragma('dart2js:noInline')
  static TestWorkspaceMemoryConnectionResponse getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<
              TestWorkspaceMemoryConnectionResponse>(
          TestWorkspaceMemoryConnectionResponse.$_createMessage);
  static TestWorkspaceMemoryConnectionResponse? _defaultInstance;

  /// True when the server answered the probe and accepted the API key.
  @$pb.TagNumber(1)
  $core.bool get ok => $_getBF(0);
  @$pb.TagNumber(1)
  set ok($core.bool value) => $_setBool(0, value);
  @$pb.TagNumber(1)
  $core.bool hasOk() => $_has(0);
  @$pb.TagNumber(1)
  void clearOk() => $_clearField(1);

  /// Human-readable probe failure, empty when ok.
  @$pb.TagNumber(2)
  $core.String get error => $_getSZ(1);
  @$pb.TagNumber(2)
  set error($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasError() => $_has(1);
  @$pb.TagNumber(2)
  void clearError() => $_clearField(2);
}

class WorkspaceMemory extends $pb.GeneratedMessage {
  factory WorkspaceMemory({
    $core.String? id,
    $core.String? memory,
    WorkspaceMemoryScope? scope,
    $core.String? agentId,
    $core.String? channel,
    $core.String? principal,
    $core.String? sessionId,
    $0.Timestamp? createdAt,
    $0.Timestamp? updatedAt,
    $core.double? score,
  }) {
    final result = WorkspaceMemory._();
    if (id != null) result.id = id;
    if (memory != null) result.memory = memory;
    if (scope != null) result.scope = scope;
    if (agentId != null) result.agentId = agentId;
    if (channel != null) result.channel = channel;
    if (principal != null) result.principal = principal;
    if (sessionId != null) result.sessionId = sessionId;
    if (createdAt != null) result.createdAt = createdAt;
    if (updatedAt != null) result.updatedAt = updatedAt;
    if (score != null) result.score = score;
    return result;
  }

  WorkspaceMemory._();

  factory WorkspaceMemory.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      WorkspaceMemory()..mergeFromBuffer(data, registry);
  factory WorkspaceMemory.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      WorkspaceMemory()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'WorkspaceMemory',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: WorkspaceMemory.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..aOS(2, _omitFieldNames ? '' : 'memory')
    ..aE<WorkspaceMemoryScope>(3, _omitFieldNames ? '' : 'scope',
        enumValues: WorkspaceMemoryScope.values)
    ..aOS(4, _omitFieldNames ? '' : 'agentId')
    ..aOS(5, _omitFieldNames ? '' : 'channel')
    ..aOS(6, _omitFieldNames ? '' : 'principal')
    ..aOS(7, _omitFieldNames ? '' : 'sessionId')
    ..aOM<$0.Timestamp>(8, _omitFieldNames ? '' : 'createdAt',
        subBuilder: $0.Timestamp.$_createMessage)
    ..aOM<$0.Timestamp>(9, _omitFieldNames ? '' : 'updatedAt',
        subBuilder: $0.Timestamp.$_createMessage)
    ..aD(10, _omitFieldNames ? '' : 'score')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  WorkspaceMemory clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  WorkspaceMemory copyWith(void Function(WorkspaceMemory) updates) =>
      super.copyWith((message) => updates(message as WorkspaceMemory))
          as WorkspaceMemory;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use WorkspaceMemory() / WorkspaceMemory.new instead')
  static WorkspaceMemory create() => WorkspaceMemory._();
  static $pb.GeneratedMessage $_createMessage() => WorkspaceMemory._();
  @$core.override
  WorkspaceMemory createEmptyInstance() => WorkspaceMemory._();
  @$core.pragma('dart2js:noInline')
  static WorkspaceMemory getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<WorkspaceMemory>(
          WorkspaceMemory.$_createMessage);
  static WorkspaceMemory? _defaultInstance;

  /// mem0 memory ID.
  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);

  /// The stored memory text.
  @$pb.TagNumber(2)
  $core.String get memory => $_getSZ(1);
  @$pb.TagNumber(2)
  set memory($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasMemory() => $_has(1);
  @$pb.TagNumber(2)
  void clearMemory() => $_clearField(2);

  /// Which scope holds the memory.
  @$pb.TagNumber(3)
  WorkspaceMemoryScope get scope => $_getN(2);
  @$pb.TagNumber(3)
  set scope(WorkspaceMemoryScope value) => $_setField(3, value);
  @$pb.TagNumber(3)
  $core.bool hasScope() => $_has(2);
  @$pb.TagNumber(3)
  void clearScope() => $_clearField(3);

  /// For Agent Memory, the owning Agent ID. For Workspace Memory, the Agent
  /// whose turn produced it, when known (provenance).
  @$pb.TagNumber(4)
  $core.String get agentId => $_getSZ(3);
  @$pb.TagNumber(4)
  set agentId($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasAgentId() => $_has(3);
  @$pb.TagNumber(4)
  void clearAgentId() => $_clearField(4);

  /// Entry point the memory came from (provenance), e.g. "web-chat",
  /// "telegram", "cron:<job>".
  @$pb.TagNumber(5)
  $core.String get channel => $_getSZ(4);
  @$pb.TagNumber(5)
  set channel($core.String value) => $_setString(4, value);
  @$pb.TagNumber(5)
  $core.bool hasChannel() => $_has(4);
  @$pb.TagNumber(5)
  void clearChannel() => $_clearField(5);

  /// Who sent the turn that produced the memory (provenance). Only returned
  /// to workspace owners/admins and global admins; empty for other members.
  @$pb.TagNumber(6)
  $core.String get principal => $_getSZ(5);
  @$pb.TagNumber(6)
  set principal($core.String value) => $_setString(5, value);
  @$pb.TagNumber(6)
  $core.bool hasPrincipal() => $_has(5);
  @$pb.TagNumber(6)
  void clearPrincipal() => $_clearField(6);

  /// Session the memory came from, when known (provenance).
  @$pb.TagNumber(7)
  $core.String get sessionId => $_getSZ(6);
  @$pb.TagNumber(7)
  set sessionId($core.String value) => $_setString(6, value);
  @$pb.TagNumber(7)
  $core.bool hasSessionId() => $_has(6);
  @$pb.TagNumber(7)
  void clearSessionId() => $_clearField(7);

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

  /// mem0 relevance score; set on search results only.
  @$pb.TagNumber(10)
  $core.double get score => $_getN(9);
  @$pb.TagNumber(10)
  set score($core.double value) => $_setDouble(9, value);
  @$pb.TagNumber(10)
  $core.bool hasScore() => $_has(9);
  @$pb.TagNumber(10)
  void clearScore() => $_clearField(10);
}

class ListWorkspaceMemoriesRequest extends $pb.GeneratedMessage {
  factory ListWorkspaceMemoriesRequest({
    WorkspaceMemoryScope? scope,
    $core.String? agentId,
  }) {
    final result = ListWorkspaceMemoriesRequest._();
    if (scope != null) result.scope = scope;
    if (agentId != null) result.agentId = agentId;
    return result;
  }

  ListWorkspaceMemoriesRequest._();

  factory ListWorkspaceMemoriesRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListWorkspaceMemoriesRequest()..mergeFromBuffer(data, registry);
  factory ListWorkspaceMemoriesRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListWorkspaceMemoriesRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListWorkspaceMemoriesRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: ListWorkspaceMemoriesRequest.$_createMessage)
    ..aE<WorkspaceMemoryScope>(1, _omitFieldNames ? '' : 'scope',
        enumValues: WorkspaceMemoryScope.values)
    ..aOS(2, _omitFieldNames ? '' : 'agentId')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListWorkspaceMemoriesRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListWorkspaceMemoriesRequest copyWith(
          void Function(ListWorkspaceMemoriesRequest) updates) =>
      super.copyWith(
              (message) => updates(message as ListWorkspaceMemoriesRequest))
          as ListWorkspaceMemoriesRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use ListWorkspaceMemoriesRequest() / ListWorkspaceMemoriesRequest.new instead')
  static ListWorkspaceMemoriesRequest create() =>
      ListWorkspaceMemoriesRequest._();
  static $pb.GeneratedMessage $_createMessage() =>
      ListWorkspaceMemoriesRequest._();
  @$core.override
  ListWorkspaceMemoriesRequest createEmptyInstance() =>
      ListWorkspaceMemoriesRequest._();
  @$core.pragma('dart2js:noInline')
  static ListWorkspaceMemoriesRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ListWorkspaceMemoriesRequest>(
          ListWorkspaceMemoriesRequest.$_createMessage);
  static ListWorkspaceMemoriesRequest? _defaultInstance;

  @$pb.TagNumber(1)
  WorkspaceMemoryScope get scope => $_getN(0);
  @$pb.TagNumber(1)
  set scope(WorkspaceMemoryScope value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasScope() => $_has(0);
  @$pb.TagNumber(1)
  void clearScope() => $_clearField(1);

  /// Required for WORKSPACE_MEMORY_SCOPE_AGENT.
  @$pb.TagNumber(2)
  $core.String get agentId => $_getSZ(1);
  @$pb.TagNumber(2)
  set agentId($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasAgentId() => $_has(1);
  @$pb.TagNumber(2)
  void clearAgentId() => $_clearField(2);
}

class ListWorkspaceMemoriesResponse extends $pb.GeneratedMessage {
  factory ListWorkspaceMemoriesResponse({
    $core.Iterable<WorkspaceMemory>? memories,
    $core.bool? truncated,
    $core.int? limit,
  }) {
    final result = ListWorkspaceMemoriesResponse._();
    if (memories != null) result.memories.addAll(memories);
    if (truncated != null) result.truncated = truncated;
    if (limit != null) result.limit = limit;
    return result;
  }

  ListWorkspaceMemoriesResponse._();

  factory ListWorkspaceMemoriesResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListWorkspaceMemoriesResponse()..mergeFromBuffer(data, registry);
  factory ListWorkspaceMemoriesResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListWorkspaceMemoriesResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListWorkspaceMemoriesResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: ListWorkspaceMemoriesResponse.$_createMessage)
    ..pPM<WorkspaceMemory>(1, _omitFieldNames ? '' : 'memories',
        subBuilder: WorkspaceMemory.$_createMessage)
    ..aOB(2, _omitFieldNames ? '' : 'truncated')
    ..aI(3, _omitFieldNames ? '' : 'limit')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListWorkspaceMemoriesResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListWorkspaceMemoriesResponse copyWith(
          void Function(ListWorkspaceMemoriesResponse) updates) =>
      super.copyWith(
              (message) => updates(message as ListWorkspaceMemoriesResponse))
          as ListWorkspaceMemoriesResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use ListWorkspaceMemoriesResponse() / ListWorkspaceMemoriesResponse.new instead')
  static ListWorkspaceMemoriesResponse create() =>
      ListWorkspaceMemoriesResponse._();
  static $pb.GeneratedMessage $_createMessage() =>
      ListWorkspaceMemoriesResponse._();
  @$core.override
  ListWorkspaceMemoriesResponse createEmptyInstance() =>
      ListWorkspaceMemoriesResponse._();
  @$core.pragma('dart2js:noInline')
  static ListWorkspaceMemoriesResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ListWorkspaceMemoriesResponse>(
          ListWorkspaceMemoriesResponse.$_createMessage);
  static ListWorkspaceMemoriesResponse? _defaultInstance;

  /// Newest first.
  @$pb.TagNumber(1)
  $pb.PbList<WorkspaceMemory> get memories => $_getList(0);

  /// True when more memories exist than the listing returns.
  @$pb.TagNumber(2)
  $core.bool get truncated => $_getBF(1);
  @$pb.TagNumber(2)
  set truncated($core.bool value) => $_setBool(1, value);
  @$pb.TagNumber(2)
  $core.bool hasTruncated() => $_has(1);
  @$pb.TagNumber(2)
  void clearTruncated() => $_clearField(2);

  /// The listing cap applied.
  @$pb.TagNumber(3)
  $core.int get limit => $_getIZ(2);
  @$pb.TagNumber(3)
  set limit($core.int value) => $_setSignedInt32(2, value);
  @$pb.TagNumber(3)
  $core.bool hasLimit() => $_has(2);
  @$pb.TagNumber(3)
  void clearLimit() => $_clearField(3);
}

class SearchWorkspaceMemoriesRequest extends $pb.GeneratedMessage {
  factory SearchWorkspaceMemoriesRequest({
    $core.String? query,
    WorkspaceMemoryScope? scope,
    $core.String? agentId,
    $core.int? topK,
  }) {
    final result = SearchWorkspaceMemoriesRequest._();
    if (query != null) result.query = query;
    if (scope != null) result.scope = scope;
    if (agentId != null) result.agentId = agentId;
    if (topK != null) result.topK = topK;
    return result;
  }

  SearchWorkspaceMemoriesRequest._();

  factory SearchWorkspaceMemoriesRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      SearchWorkspaceMemoriesRequest()..mergeFromBuffer(data, registry);
  factory SearchWorkspaceMemoriesRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      SearchWorkspaceMemoriesRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'SearchWorkspaceMemoriesRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: SearchWorkspaceMemoriesRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'query')
    ..aE<WorkspaceMemoryScope>(2, _omitFieldNames ? '' : 'scope',
        enumValues: WorkspaceMemoryScope.values)
    ..aOS(3, _omitFieldNames ? '' : 'agentId')
    ..aI(4, _omitFieldNames ? '' : 'topK')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  SearchWorkspaceMemoriesRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  SearchWorkspaceMemoriesRequest copyWith(
          void Function(SearchWorkspaceMemoriesRequest) updates) =>
      super.copyWith(
              (message) => updates(message as SearchWorkspaceMemoriesRequest))
          as SearchWorkspaceMemoriesRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use SearchWorkspaceMemoriesRequest() / SearchWorkspaceMemoriesRequest.new instead')
  static SearchWorkspaceMemoriesRequest create() =>
      SearchWorkspaceMemoriesRequest._();
  static $pb.GeneratedMessage $_createMessage() =>
      SearchWorkspaceMemoriesRequest._();
  @$core.override
  SearchWorkspaceMemoriesRequest createEmptyInstance() =>
      SearchWorkspaceMemoriesRequest._();
  @$core.pragma('dart2js:noInline')
  static SearchWorkspaceMemoriesRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<SearchWorkspaceMemoriesRequest>(
          SearchWorkspaceMemoriesRequest.$_createMessage);
  static SearchWorkspaceMemoriesRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get query => $_getSZ(0);
  @$pb.TagNumber(1)
  set query($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasQuery() => $_has(0);
  @$pb.TagNumber(1)
  void clearQuery() => $_clearField(1);

  @$pb.TagNumber(2)
  WorkspaceMemoryScope get scope => $_getN(1);
  @$pb.TagNumber(2)
  set scope(WorkspaceMemoryScope value) => $_setField(2, value);
  @$pb.TagNumber(2)
  $core.bool hasScope() => $_has(1);
  @$pb.TagNumber(2)
  void clearScope() => $_clearField(2);

  /// Required for WORKSPACE_MEMORY_SCOPE_AGENT.
  @$pb.TagNumber(3)
  $core.String get agentId => $_getSZ(2);
  @$pb.TagNumber(3)
  set agentId($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasAgentId() => $_has(2);
  @$pb.TagNumber(3)
  void clearAgentId() => $_clearField(3);

  /// Maximum results; 0 means 20, capped at 100.
  @$pb.TagNumber(4)
  $core.int get topK => $_getIZ(3);
  @$pb.TagNumber(4)
  set topK($core.int value) => $_setSignedInt32(3, value);
  @$pb.TagNumber(4)
  $core.bool hasTopK() => $_has(3);
  @$pb.TagNumber(4)
  void clearTopK() => $_clearField(4);
}

class SearchWorkspaceMemoriesResponse extends $pb.GeneratedMessage {
  factory SearchWorkspaceMemoriesResponse({
    $core.Iterable<WorkspaceMemory>? memories,
  }) {
    final result = SearchWorkspaceMemoriesResponse._();
    if (memories != null) result.memories.addAll(memories);
    return result;
  }

  SearchWorkspaceMemoriesResponse._();

  factory SearchWorkspaceMemoriesResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      SearchWorkspaceMemoriesResponse()..mergeFromBuffer(data, registry);
  factory SearchWorkspaceMemoriesResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      SearchWorkspaceMemoriesResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'SearchWorkspaceMemoriesResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: SearchWorkspaceMemoriesResponse.$_createMessage)
    ..pPM<WorkspaceMemory>(1, _omitFieldNames ? '' : 'memories',
        subBuilder: WorkspaceMemory.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  SearchWorkspaceMemoriesResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  SearchWorkspaceMemoriesResponse copyWith(
          void Function(SearchWorkspaceMemoriesResponse) updates) =>
      super.copyWith(
              (message) => updates(message as SearchWorkspaceMemoriesResponse))
          as SearchWorkspaceMemoriesResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use SearchWorkspaceMemoriesResponse() / SearchWorkspaceMemoriesResponse.new instead')
  static SearchWorkspaceMemoriesResponse create() =>
      SearchWorkspaceMemoriesResponse._();
  static $pb.GeneratedMessage $_createMessage() =>
      SearchWorkspaceMemoriesResponse._();
  @$core.override
  SearchWorkspaceMemoriesResponse createEmptyInstance() =>
      SearchWorkspaceMemoriesResponse._();
  @$core.pragma('dart2js:noInline')
  static SearchWorkspaceMemoriesResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<SearchWorkspaceMemoriesResponse>(
          SearchWorkspaceMemoriesResponse.$_createMessage);
  static SearchWorkspaceMemoriesResponse? _defaultInstance;

  /// Most relevant first.
  @$pb.TagNumber(1)
  $pb.PbList<WorkspaceMemory> get memories => $_getList(0);
}

class DeleteWorkspaceMemoryRequest extends $pb.GeneratedMessage {
  factory DeleteWorkspaceMemoryRequest({
    $core.String? memoryId,
  }) {
    final result = DeleteWorkspaceMemoryRequest._();
    if (memoryId != null) result.memoryId = memoryId;
    return result;
  }

  DeleteWorkspaceMemoryRequest._();

  factory DeleteWorkspaceMemoryRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DeleteWorkspaceMemoryRequest()..mergeFromBuffer(data, registry);
  factory DeleteWorkspaceMemoryRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DeleteWorkspaceMemoryRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'DeleteWorkspaceMemoryRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: DeleteWorkspaceMemoryRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'memoryId')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteWorkspaceMemoryRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteWorkspaceMemoryRequest copyWith(
          void Function(DeleteWorkspaceMemoryRequest) updates) =>
      super.copyWith(
              (message) => updates(message as DeleteWorkspaceMemoryRequest))
          as DeleteWorkspaceMemoryRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use DeleteWorkspaceMemoryRequest() / DeleteWorkspaceMemoryRequest.new instead')
  static DeleteWorkspaceMemoryRequest create() =>
      DeleteWorkspaceMemoryRequest._();
  static $pb.GeneratedMessage $_createMessage() =>
      DeleteWorkspaceMemoryRequest._();
  @$core.override
  DeleteWorkspaceMemoryRequest createEmptyInstance() =>
      DeleteWorkspaceMemoryRequest._();
  @$core.pragma('dart2js:noInline')
  static DeleteWorkspaceMemoryRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<DeleteWorkspaceMemoryRequest>(
          DeleteWorkspaceMemoryRequest.$_createMessage);
  static DeleteWorkspaceMemoryRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get memoryId => $_getSZ(0);
  @$pb.TagNumber(1)
  set memoryId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasMemoryId() => $_has(0);
  @$pb.TagNumber(1)
  void clearMemoryId() => $_clearField(1);
}

class DeleteWorkspaceMemoryResponse extends $pb.GeneratedMessage {
  factory DeleteWorkspaceMemoryResponse() => DeleteWorkspaceMemoryResponse._();

  DeleteWorkspaceMemoryResponse._();

  factory DeleteWorkspaceMemoryResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DeleteWorkspaceMemoryResponse()..mergeFromBuffer(data, registry);
  factory DeleteWorkspaceMemoryResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DeleteWorkspaceMemoryResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'DeleteWorkspaceMemoryResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: DeleteWorkspaceMemoryResponse.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteWorkspaceMemoryResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteWorkspaceMemoryResponse copyWith(
          void Function(DeleteWorkspaceMemoryResponse) updates) =>
      super.copyWith(
              (message) => updates(message as DeleteWorkspaceMemoryResponse))
          as DeleteWorkspaceMemoryResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use DeleteWorkspaceMemoryResponse() / DeleteWorkspaceMemoryResponse.new instead')
  static DeleteWorkspaceMemoryResponse create() =>
      DeleteWorkspaceMemoryResponse._();
  static $pb.GeneratedMessage $_createMessage() =>
      DeleteWorkspaceMemoryResponse._();
  @$core.override
  DeleteWorkspaceMemoryResponse createEmptyInstance() =>
      DeleteWorkspaceMemoryResponse._();
  @$core.pragma('dart2js:noInline')
  static DeleteWorkspaceMemoryResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<DeleteWorkspaceMemoryResponse>(
          DeleteWorkspaceMemoryResponse.$_createMessage);
  static DeleteWorkspaceMemoryResponse? _defaultInstance;
}

/// WorkspaceMemoryConfigService manages the workspace's mem0 connection. Any
/// member may read it; owners, admins, and global admins may change it.
class WorkspaceMemoryConfigServiceApi {
  final $pb.RpcClient _client;

  WorkspaceMemoryConfigServiceApi(this._client);

  /// GetWorkspaceMemoryConfig returns the config, or an unset `config` when
  /// the workspace has none.
  $async.Future<GetWorkspaceMemoryConfigResponse> getWorkspaceMemoryConfig(
          $pb.ClientContext? ctx, GetWorkspaceMemoryConfigRequest request) =>
      _client.invoke<GetWorkspaceMemoryConfigResponse>(
          ctx,
          'WorkspaceMemoryConfigService',
          'GetWorkspaceMemoryConfig',
          request,
          GetWorkspaceMemoryConfigResponse());

  /// PutWorkspaceMemoryConfig creates or replaces the config. When the result
  /// is enabled, the server is probed first: a rejected API key refuses the
  /// save, while an unreachable server saves and reports a warning.
  $async.Future<PutWorkspaceMemoryConfigResponse> putWorkspaceMemoryConfig(
          $pb.ClientContext? ctx, PutWorkspaceMemoryConfigRequest request) =>
      _client.invoke<PutWorkspaceMemoryConfigResponse>(
          ctx,
          'WorkspaceMemoryConfigService',
          'PutWorkspaceMemoryConfig',
          request,
          PutWorkspaceMemoryConfigResponse());

  /// DeleteWorkspaceMemoryConfig removes the config and its API key. Agents
  /// that enable memory keep running, without memory.
  $async.Future<DeleteWorkspaceMemoryConfigResponse>
      deleteWorkspaceMemoryConfig($pb.ClientContext? ctx,
              DeleteWorkspaceMemoryConfigRequest request) =>
          _client.invoke<DeleteWorkspaceMemoryConfigResponse>(
              ctx,
              'WorkspaceMemoryConfigService',
              'DeleteWorkspaceMemoryConfig',
              request,
              DeleteWorkspaceMemoryConfigResponse());

  /// TestWorkspaceMemoryConnection probes the stored config's server with the
  /// stored API key and reports the outcome.
  $async.Future<TestWorkspaceMemoryConnectionResponse>
      testWorkspaceMemoryConnection($pb.ClientContext? ctx,
              TestWorkspaceMemoryConnectionRequest request) =>
          _client.invoke<TestWorkspaceMemoryConnectionResponse>(
              ctx,
              'WorkspaceMemoryConfigService',
              'TestWorkspaceMemoryConnection',
              request,
              TestWorkspaceMemoryConnectionResponse());
}

class WorkspaceMemoryServiceApi {
  final $pb.RpcClient _client;

  WorkspaceMemoryServiceApi(this._client);

  /// ListWorkspaceMemories returns the newest memories of one scope, capped
  /// at `limit` (mem0 OSS listing has no pagination; search reaches the rest).
  $async.Future<ListWorkspaceMemoriesResponse> listWorkspaceMemories(
          $pb.ClientContext? ctx, ListWorkspaceMemoriesRequest request) =>
      _client.invoke<ListWorkspaceMemoriesResponse>(
          ctx,
          'WorkspaceMemoryService',
          'ListWorkspaceMemories',
          request,
          ListWorkspaceMemoriesResponse());

  /// SearchWorkspaceMemories runs a semantic search in one scope.
  $async.Future<SearchWorkspaceMemoriesResponse> searchWorkspaceMemories(
          $pb.ClientContext? ctx, SearchWorkspaceMemoriesRequest request) =>
      _client.invoke<SearchWorkspaceMemoriesResponse>(
          ctx,
          'WorkspaceMemoryService',
          'SearchWorkspaceMemories',
          request,
          SearchWorkspaceMemoriesResponse());

  /// DeleteWorkspaceMemory deletes one memory of this workspace. A memory
  /// that does not exist or belongs to another workspace is NotFound.
  $async.Future<DeleteWorkspaceMemoryResponse> deleteWorkspaceMemory(
          $pb.ClientContext? ctx, DeleteWorkspaceMemoryRequest request) =>
      _client.invoke<DeleteWorkspaceMemoryResponse>(
          ctx,
          'WorkspaceMemoryService',
          'DeleteWorkspaceMemory',
          request,
          DeleteWorkspaceMemoryResponse());
}

const $core.bool _omitFieldNames =
    $core.bool.fromEnvironment('protobuf.omit_field_names');
const $core.bool _omitMessageNames =
    $core.bool.fromEnvironment('protobuf.omit_message_names');
