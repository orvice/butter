// This is a generated file - do not edit.
//
// Generated from agents/v1/linear.proto.

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

import 'linear.pbenum.dart';

export 'package:protobuf/protobuf.dart' show GeneratedMessageGenericExtensions;

export 'linear.pbenum.dart';

/// LinearApp is one Linear OAuth application registered in a workspace.
class LinearApp extends $pb.GeneratedMessage {
  factory LinearApp({
    $core.String? id,
    $core.String? displayName,
    $core.String? clientId,
    $core.String? agentId,
    $core.bool? inboundEnabled,
    $core.Iterable<$core.String>? allowedUserIds,
    $core.int? maxRunSeconds,
    $fixnum.Int64? revision,
    LinearAppCredentialState? credentialState,
    $core.bool? clientSecretSet,
    $core.bool? webhookSecretSet,
    $core.String? callbackUrl,
    $core.String? webhookUrl,
    $0.Timestamp? createdAt,
    $0.Timestamp? updatedAt,
    $core.String? workspaceId,
  }) {
    final result = LinearApp._();
    if (id != null) result.id = id;
    if (displayName != null) result.displayName = displayName;
    if (clientId != null) result.clientId = clientId;
    if (agentId != null) result.agentId = agentId;
    if (inboundEnabled != null) result.inboundEnabled = inboundEnabled;
    if (allowedUserIds != null) result.allowedUserIds.addAll(allowedUserIds);
    if (maxRunSeconds != null) result.maxRunSeconds = maxRunSeconds;
    if (revision != null) result.revision = revision;
    if (credentialState != null) result.credentialState = credentialState;
    if (clientSecretSet != null) result.clientSecretSet = clientSecretSet;
    if (webhookSecretSet != null) result.webhookSecretSet = webhookSecretSet;
    if (callbackUrl != null) result.callbackUrl = callbackUrl;
    if (webhookUrl != null) result.webhookUrl = webhookUrl;
    if (createdAt != null) result.createdAt = createdAt;
    if (updatedAt != null) result.updatedAt = updatedAt;
    if (workspaceId != null) result.workspaceId = workspaceId;
    return result;
  }

  LinearApp._();

  factory LinearApp.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      LinearApp()..mergeFromBuffer(data, registry);
  factory LinearApp.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      LinearApp()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'LinearApp',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: LinearApp.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..aOS(2, _omitFieldNames ? '' : 'displayName')
    ..aOS(3, _omitFieldNames ? '' : 'clientId')
    ..aOS(4, _omitFieldNames ? '' : 'agentId')
    ..aOB(5, _omitFieldNames ? '' : 'inboundEnabled')
    ..pPS(6, _omitFieldNames ? '' : 'allowedUserIds')
    ..aI(7, _omitFieldNames ? '' : 'maxRunSeconds')
    ..aInt64(8, _omitFieldNames ? '' : 'revision')
    ..aE<LinearAppCredentialState>(9, _omitFieldNames ? '' : 'credentialState',
        enumValues: LinearAppCredentialState.values)
    ..aOB(10, _omitFieldNames ? '' : 'clientSecretSet')
    ..aOB(11, _omitFieldNames ? '' : 'webhookSecretSet')
    ..aOS(12, _omitFieldNames ? '' : 'callbackUrl')
    ..aOS(13, _omitFieldNames ? '' : 'webhookUrl')
    ..aOM<$0.Timestamp>(14, _omitFieldNames ? '' : 'createdAt',
        subBuilder: $0.Timestamp.$_createMessage)
    ..aOM<$0.Timestamp>(15, _omitFieldNames ? '' : 'updatedAt',
        subBuilder: $0.Timestamp.$_createMessage)
    ..aOS(100, _omitFieldNames ? '' : 'workspaceId')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  LinearApp clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  LinearApp copyWith(void Function(LinearApp) updates) =>
      super.copyWith((message) => updates(message as LinearApp)) as LinearApp;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use LinearApp() / LinearApp.new instead')
  static LinearApp create() => LinearApp._();
  static $pb.GeneratedMessage $_createMessage() => LinearApp._();
  @$core.override
  LinearApp createEmptyInstance() => LinearApp._();
  @$core.pragma('dart2js:noInline')
  static LinearApp getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<LinearApp>(LinearApp.$_createMessage);
  static LinearApp? _defaultInstance;

  /// Service-generated immutable identifier. The webhook URL is derived from
  /// it, so it never changes.
  /// @gotags: json:"id,omitempty" bson:"id,omitempty"
  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);

  /// Mutable display name, also the channel name of the Agent's turns.
  /// @gotags: json:"display_name,omitempty" bson:"display_name,omitempty"
  @$pb.TagNumber(2)
  $core.String get displayName => $_getSZ(1);
  @$pb.TagNumber(2)
  set displayName($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasDisplayName() => $_has(1);
  @$pb.TagNumber(2)
  void clearDisplayName() => $_clearField(2);

  /// The Linear OAuth application's public client ID. Globally unique: one
  /// Linear app can be registered in exactly one Butter workspace.
  /// @gotags: json:"client_id,omitempty" bson:"client_id,omitempty"
  @$pb.TagNumber(3)
  $core.String get clientId => $_getSZ(2);
  @$pb.TagNumber(3)
  set clientId($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasClientId() => $_has(2);
  @$pb.TagNumber(3)
  void clearClientId() => $_clearField(3);

  /// The Agent every Linear Agent Session of this App is routed to. Must name
  /// an active Agent of the workspace.
  /// @gotags: json:"agent_id,omitempty" bson:"agent_id,omitempty"
  @$pb.TagNumber(4)
  $core.String get agentId => $_getSZ(3);
  @$pb.TagNumber(4)
  set agentId($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasAgentId() => $_has(3);
  @$pb.TagNumber(4)
  void clearAgentId() => $_clearField(4);

  /// Desired inbound state: whether Linear deliveries are accepted. Enabling
  /// requires both secrets and a durable queue.
  /// @gotags: json:"inbound_enabled,omitempty" bson:"inbound_enabled,omitempty"
  @$pb.TagNumber(5)
  $core.bool get inboundEnabled => $_getBF(4);
  @$pb.TagNumber(5)
  set inboundEnabled($core.bool value) => $_setBool(4, value);
  @$pb.TagNumber(5)
  $core.bool hasInboundEnabled() => $_has(4);
  @$pb.TagNumber(5)
  void clearInboundEnabled() => $_clearField(5);

  /// Linear user IDs admitted to open or prompt a session. Empty admits every
  /// member of the installed Linear organizations.
  /// @gotags: json:"allowed_user_ids,omitempty" bson:"allowed_user_ids,omitempty"
  @$pb.TagNumber(6)
  $pb.PbList<$core.String> get allowedUserIds => $_getList(5);

  /// Maximum run time of one turn, in seconds. Absent means 1800; an explicit
  /// 0 means unlimited.
  /// @gotags: json:"max_run_seconds,omitempty" bson:"max_run_seconds,omitempty"
  @$pb.TagNumber(7)
  $core.int get maxRunSeconds => $_getIZ(6);
  @$pb.TagNumber(7)
  set maxRunSeconds($core.int value) => $_setSignedInt32(6, value);
  @$pb.TagNumber(7)
  $core.bool hasMaxRunSeconds() => $_has(6);
  @$pb.TagNumber(7)
  void clearMaxRunSeconds() => $_clearField(7);

  /// Optimistic concurrency token. Updates must send the revision they read.
  /// @gotags: json:"revision,omitempty" bson:"revision,omitempty"
  @$pb.TagNumber(8)
  $fixnum.Int64 get revision => $_getI64(7);
  @$pb.TagNumber(8)
  set revision($fixnum.Int64 value) => $_setInt64(7, value);
  @$pb.TagNumber(8)
  $core.bool hasRevision() => $_has(7);
  @$pb.TagNumber(8)
  void clearRevision() => $_clearField(8);

  /// Server-owned. Never carries credential material.
  /// @gotags: json:"credential_state,omitempty" bson:"credential_state,omitempty"
  @$pb.TagNumber(9)
  LinearAppCredentialState get credentialState => $_getN(8);
  @$pb.TagNumber(9)
  set credentialState(LinearAppCredentialState value) => $_setField(9, value);
  @$pb.TagNumber(9)
  $core.bool hasCredentialState() => $_has(8);
  @$pb.TagNumber(9)
  void clearCredentialState() => $_clearField(9);

  /// Whether the OAuth client secret is stored. Server-owned.
  /// @gotags: json:"client_secret_set,omitempty" bson:"client_secret_set,omitempty"
  @$pb.TagNumber(10)
  $core.bool get clientSecretSet => $_getBF(9);
  @$pb.TagNumber(10)
  set clientSecretSet($core.bool value) => $_setBool(9, value);
  @$pb.TagNumber(10)
  $core.bool hasClientSecretSet() => $_has(9);
  @$pb.TagNumber(10)
  void clearClientSecretSet() => $_clearField(10);

  /// Whether the webhook signing secret is stored. Server-owned.
  /// @gotags: json:"webhook_secret_set,omitempty" bson:"webhook_secret_set,omitempty"
  @$pb.TagNumber(11)
  $core.bool get webhookSecretSet => $_getBF(10);
  @$pb.TagNumber(11)
  set webhookSecretSet($core.bool value) => $_setBool(10, value);
  @$pb.TagNumber(11)
  $core.bool hasWebhookSecretSet() => $_has(10);
  @$pb.TagNumber(11)
  void clearWebhookSecretSet() => $_clearField(11);

  /// The OAuth callback URL to register with the Linear app. Output only:
  /// derived from the platform Linear base URL, empty until it is configured.
  /// @gotags: json:"callback_url,omitempty" bson:"callback_url,omitempty"
  @$pb.TagNumber(12)
  $core.String get callbackUrl => $_getSZ(11);
  @$pb.TagNumber(12)
  set callbackUrl($core.String value) => $_setString(11, value);
  @$pb.TagNumber(12)
  $core.bool hasCallbackUrl() => $_has(11);
  @$pb.TagNumber(12)
  void clearCallbackUrl() => $_clearField(12);

  /// The webhook URL to register with the Linear app. Output only: derived
  /// from the platform Linear base URL and this App's ID.
  /// @gotags: json:"webhook_url,omitempty" bson:"webhook_url,omitempty"
  @$pb.TagNumber(13)
  $core.String get webhookUrl => $_getSZ(12);
  @$pb.TagNumber(13)
  set webhookUrl($core.String value) => $_setString(12, value);
  @$pb.TagNumber(13)
  $core.bool hasWebhookUrl() => $_has(12);
  @$pb.TagNumber(13)
  void clearWebhookUrl() => $_clearField(13);

  /// @gotags: json:"created_at,omitempty" bson:"created_at,omitempty"
  @$pb.TagNumber(14)
  $0.Timestamp get createdAt => $_getN(13);
  @$pb.TagNumber(14)
  set createdAt($0.Timestamp value) => $_setField(14, value);
  @$pb.TagNumber(14)
  $core.bool hasCreatedAt() => $_has(13);
  @$pb.TagNumber(14)
  void clearCreatedAt() => $_clearField(14);
  @$pb.TagNumber(14)
  $0.Timestamp ensureCreatedAt() => $_ensure(13);

  /// @gotags: json:"updated_at,omitempty" bson:"updated_at,omitempty"
  @$pb.TagNumber(15)
  $0.Timestamp get updatedAt => $_getN(14);
  @$pb.TagNumber(15)
  set updatedAt($0.Timestamp value) => $_setField(15, value);
  @$pb.TagNumber(15)
  $core.bool hasUpdatedAt() => $_has(14);
  @$pb.TagNumber(15)
  void clearUpdatedAt() => $_clearField(15);
  @$pb.TagNumber(15)
  $0.Timestamp ensureUpdatedAt() => $_ensure(14);

  /// @gotags: json:"workspace_id,omitempty" bson:"workspace_id,omitempty"
  @$pb.TagNumber(100)
  $core.String get workspaceId => $_getSZ(15);
  @$pb.TagNumber(100)
  set workspaceId($core.String value) => $_setString(15, value);
  @$pb.TagNumber(100)
  $core.bool hasWorkspaceId() => $_has(15);
  @$pb.TagNumber(100)
  void clearWorkspaceId() => $_clearField(100);
}

class ListLinearAppsRequest extends $pb.GeneratedMessage {
  factory ListLinearAppsRequest() => ListLinearAppsRequest._();

  ListLinearAppsRequest._();

  factory ListLinearAppsRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListLinearAppsRequest()..mergeFromBuffer(data, registry);
  factory ListLinearAppsRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListLinearAppsRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListLinearAppsRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: ListLinearAppsRequest.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListLinearAppsRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListLinearAppsRequest copyWith(
          void Function(ListLinearAppsRequest) updates) =>
      super.copyWith((message) => updates(message as ListLinearAppsRequest))
          as ListLinearAppsRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use ListLinearAppsRequest() / ListLinearAppsRequest.new instead')
  static ListLinearAppsRequest create() => ListLinearAppsRequest._();
  static $pb.GeneratedMessage $_createMessage() => ListLinearAppsRequest._();
  @$core.override
  ListLinearAppsRequest createEmptyInstance() => ListLinearAppsRequest._();
  @$core.pragma('dart2js:noInline')
  static ListLinearAppsRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ListLinearAppsRequest>(
          ListLinearAppsRequest.$_createMessage);
  static ListLinearAppsRequest? _defaultInstance;
}

class ListLinearAppsResponse extends $pb.GeneratedMessage {
  factory ListLinearAppsResponse({
    $core.Iterable<LinearApp>? apps,
  }) {
    final result = ListLinearAppsResponse._();
    if (apps != null) result.apps.addAll(apps);
    return result;
  }

  ListLinearAppsResponse._();

  factory ListLinearAppsResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListLinearAppsResponse()..mergeFromBuffer(data, registry);
  factory ListLinearAppsResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListLinearAppsResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListLinearAppsResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: ListLinearAppsResponse.$_createMessage)
    ..pPM<LinearApp>(1, _omitFieldNames ? '' : 'apps',
        subBuilder: LinearApp.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListLinearAppsResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListLinearAppsResponse copyWith(
          void Function(ListLinearAppsResponse) updates) =>
      super.copyWith((message) => updates(message as ListLinearAppsResponse))
          as ListLinearAppsResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use ListLinearAppsResponse() / ListLinearAppsResponse.new instead')
  static ListLinearAppsResponse create() => ListLinearAppsResponse._();
  static $pb.GeneratedMessage $_createMessage() => ListLinearAppsResponse._();
  @$core.override
  ListLinearAppsResponse createEmptyInstance() => ListLinearAppsResponse._();
  @$core.pragma('dart2js:noInline')
  static ListLinearAppsResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ListLinearAppsResponse>(
          ListLinearAppsResponse.$_createMessage);
  static ListLinearAppsResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $pb.PbList<LinearApp> get apps => $_getList(0);
}

class GetLinearAppRequest extends $pb.GeneratedMessage {
  factory GetLinearAppRequest({
    $core.String? id,
  }) {
    final result = GetLinearAppRequest._();
    if (id != null) result.id = id;
    return result;
  }

  GetLinearAppRequest._();

  factory GetLinearAppRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetLinearAppRequest()..mergeFromBuffer(data, registry);
  factory GetLinearAppRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetLinearAppRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetLinearAppRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: GetLinearAppRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetLinearAppRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetLinearAppRequest copyWith(void Function(GetLinearAppRequest) updates) =>
      super.copyWith((message) => updates(message as GetLinearAppRequest))
          as GetLinearAppRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core
      .Deprecated('Use GetLinearAppRequest() / GetLinearAppRequest.new instead')
  static GetLinearAppRequest create() => GetLinearAppRequest._();
  static $pb.GeneratedMessage $_createMessage() => GetLinearAppRequest._();
  @$core.override
  GetLinearAppRequest createEmptyInstance() => GetLinearAppRequest._();
  @$core.pragma('dart2js:noInline')
  static GetLinearAppRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<GetLinearAppRequest>(
          GetLinearAppRequest.$_createMessage);
  static GetLinearAppRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);
}

class GetLinearAppResponse extends $pb.GeneratedMessage {
  factory GetLinearAppResponse({
    LinearApp? app,
  }) {
    final result = GetLinearAppResponse._();
    if (app != null) result.app = app;
    return result;
  }

  GetLinearAppResponse._();

  factory GetLinearAppResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetLinearAppResponse()..mergeFromBuffer(data, registry);
  factory GetLinearAppResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetLinearAppResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetLinearAppResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: GetLinearAppResponse.$_createMessage)
    ..aOM<LinearApp>(1, _omitFieldNames ? '' : 'app',
        subBuilder: LinearApp.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetLinearAppResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetLinearAppResponse copyWith(void Function(GetLinearAppResponse) updates) =>
      super.copyWith((message) => updates(message as GetLinearAppResponse))
          as GetLinearAppResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use GetLinearAppResponse() / GetLinearAppResponse.new instead')
  static GetLinearAppResponse create() => GetLinearAppResponse._();
  static $pb.GeneratedMessage $_createMessage() => GetLinearAppResponse._();
  @$core.override
  GetLinearAppResponse createEmptyInstance() => GetLinearAppResponse._();
  @$core.pragma('dart2js:noInline')
  static GetLinearAppResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<GetLinearAppResponse>(
          GetLinearAppResponse.$_createMessage);
  static GetLinearAppResponse? _defaultInstance;

  @$pb.TagNumber(1)
  LinearApp get app => $_getN(0);
  @$pb.TagNumber(1)
  set app(LinearApp value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasApp() => $_has(0);
  @$pb.TagNumber(1)
  void clearApp() => $_clearField(1);
  @$pb.TagNumber(1)
  LinearApp ensureApp() => $_ensure(0);
}

class CreateLinearAppRequest extends $pb.GeneratedMessage {
  factory CreateLinearAppRequest({
    LinearApp? app,
    $core.String? clientSecret,
    $core.String? webhookSecret,
  }) {
    final result = CreateLinearAppRequest._();
    if (app != null) result.app = app;
    if (clientSecret != null) result.clientSecret = clientSecret;
    if (webhookSecret != null) result.webhookSecret = webhookSecret;
    return result;
  }

  CreateLinearAppRequest._();

  factory CreateLinearAppRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateLinearAppRequest()..mergeFromBuffer(data, registry);
  factory CreateLinearAppRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateLinearAppRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CreateLinearAppRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: CreateLinearAppRequest.$_createMessage)
    ..aOM<LinearApp>(1, _omitFieldNames ? '' : 'app',
        subBuilder: LinearApp.$_createMessage)
    ..aOS(2, _omitFieldNames ? '' : 'clientSecret')
    ..aOS(3, _omitFieldNames ? '' : 'webhookSecret')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateLinearAppRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateLinearAppRequest copyWith(
          void Function(CreateLinearAppRequest) updates) =>
      super.copyWith((message) => updates(message as CreateLinearAppRequest))
          as CreateLinearAppRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use CreateLinearAppRequest() / CreateLinearAppRequest.new instead')
  static CreateLinearAppRequest create() => CreateLinearAppRequest._();
  static $pb.GeneratedMessage $_createMessage() => CreateLinearAppRequest._();
  @$core.override
  CreateLinearAppRequest createEmptyInstance() => CreateLinearAppRequest._();
  @$core.pragma('dart2js:noInline')
  static CreateLinearAppRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<CreateLinearAppRequest>(
          CreateLinearAppRequest.$_createMessage);
  static CreateLinearAppRequest? _defaultInstance;

  /// Only display_name, client_id, agent_id, inbound_enabled,
  /// allowed_user_ids and max_run_seconds are read. Identity, revision,
  /// timestamps and credential state are server-owned.
  @$pb.TagNumber(1)
  LinearApp get app => $_getN(0);
  @$pb.TagNumber(1)
  set app(LinearApp value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasApp() => $_has(0);
  @$pb.TagNumber(1)
  void clearApp() => $_clearField(1);
  @$pb.TagNumber(1)
  LinearApp ensureApp() => $_ensure(0);

  /// OAuth client secret, write-only. Encrypted before storage and never
  /// returned by any RPC.
  @$pb.TagNumber(2)
  $core.String get clientSecret => $_getSZ(1);
  @$pb.TagNumber(2)
  set clientSecret($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasClientSecret() => $_has(1);
  @$pb.TagNumber(2)
  void clearClientSecret() => $_clearField(2);

  /// Webhook signing secret, write-only.
  @$pb.TagNumber(3)
  $core.String get webhookSecret => $_getSZ(2);
  @$pb.TagNumber(3)
  set webhookSecret($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasWebhookSecret() => $_has(2);
  @$pb.TagNumber(3)
  void clearWebhookSecret() => $_clearField(3);
}

class CreateLinearAppResponse extends $pb.GeneratedMessage {
  factory CreateLinearAppResponse({
    LinearApp? app,
  }) {
    final result = CreateLinearAppResponse._();
    if (app != null) result.app = app;
    return result;
  }

  CreateLinearAppResponse._();

  factory CreateLinearAppResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateLinearAppResponse()..mergeFromBuffer(data, registry);
  factory CreateLinearAppResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateLinearAppResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CreateLinearAppResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: CreateLinearAppResponse.$_createMessage)
    ..aOM<LinearApp>(1, _omitFieldNames ? '' : 'app',
        subBuilder: LinearApp.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateLinearAppResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateLinearAppResponse copyWith(
          void Function(CreateLinearAppResponse) updates) =>
      super.copyWith((message) => updates(message as CreateLinearAppResponse))
          as CreateLinearAppResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use CreateLinearAppResponse() / CreateLinearAppResponse.new instead')
  static CreateLinearAppResponse create() => CreateLinearAppResponse._();
  static $pb.GeneratedMessage $_createMessage() => CreateLinearAppResponse._();
  @$core.override
  CreateLinearAppResponse createEmptyInstance() => CreateLinearAppResponse._();
  @$core.pragma('dart2js:noInline')
  static CreateLinearAppResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<CreateLinearAppResponse>(
          CreateLinearAppResponse.$_createMessage);
  static CreateLinearAppResponse? _defaultInstance;

  @$pb.TagNumber(1)
  LinearApp get app => $_getN(0);
  @$pb.TagNumber(1)
  set app(LinearApp value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasApp() => $_has(0);
  @$pb.TagNumber(1)
  void clearApp() => $_clearField(1);
  @$pb.TagNumber(1)
  LinearApp ensureApp() => $_ensure(0);
}

class UpdateLinearAppRequest extends $pb.GeneratedMessage {
  factory UpdateLinearAppRequest({
    LinearApp? app,
  }) {
    final result = UpdateLinearAppRequest._();
    if (app != null) result.app = app;
    return result;
  }

  UpdateLinearAppRequest._();

  factory UpdateLinearAppRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpdateLinearAppRequest()..mergeFromBuffer(data, registry);
  factory UpdateLinearAppRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpdateLinearAppRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'UpdateLinearAppRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: UpdateLinearAppRequest.$_createMessage)
    ..aOM<LinearApp>(1, _omitFieldNames ? '' : 'app',
        subBuilder: LinearApp.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateLinearAppRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateLinearAppRequest copyWith(
          void Function(UpdateLinearAppRequest) updates) =>
      super.copyWith((message) => updates(message as UpdateLinearAppRequest))
          as UpdateLinearAppRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use UpdateLinearAppRequest() / UpdateLinearAppRequest.new instead')
  static UpdateLinearAppRequest create() => UpdateLinearAppRequest._();
  static $pb.GeneratedMessage $_createMessage() => UpdateLinearAppRequest._();
  @$core.override
  UpdateLinearAppRequest createEmptyInstance() => UpdateLinearAppRequest._();
  @$core.pragma('dart2js:noInline')
  static UpdateLinearAppRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<UpdateLinearAppRequest>(
          UpdateLinearAppRequest.$_createMessage);
  static UpdateLinearAppRequest? _defaultInstance;

  /// `id` selects the App and `revision` must match the stored one.
  @$pb.TagNumber(1)
  LinearApp get app => $_getN(0);
  @$pb.TagNumber(1)
  set app(LinearApp value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasApp() => $_has(0);
  @$pb.TagNumber(1)
  void clearApp() => $_clearField(1);
  @$pb.TagNumber(1)
  LinearApp ensureApp() => $_ensure(0);
}

class UpdateLinearAppResponse extends $pb.GeneratedMessage {
  factory UpdateLinearAppResponse({
    LinearApp? app,
  }) {
    final result = UpdateLinearAppResponse._();
    if (app != null) result.app = app;
    return result;
  }

  UpdateLinearAppResponse._();

  factory UpdateLinearAppResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpdateLinearAppResponse()..mergeFromBuffer(data, registry);
  factory UpdateLinearAppResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpdateLinearAppResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'UpdateLinearAppResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: UpdateLinearAppResponse.$_createMessage)
    ..aOM<LinearApp>(1, _omitFieldNames ? '' : 'app',
        subBuilder: LinearApp.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateLinearAppResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateLinearAppResponse copyWith(
          void Function(UpdateLinearAppResponse) updates) =>
      super.copyWith((message) => updates(message as UpdateLinearAppResponse))
          as UpdateLinearAppResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use UpdateLinearAppResponse() / UpdateLinearAppResponse.new instead')
  static UpdateLinearAppResponse create() => UpdateLinearAppResponse._();
  static $pb.GeneratedMessage $_createMessage() => UpdateLinearAppResponse._();
  @$core.override
  UpdateLinearAppResponse createEmptyInstance() => UpdateLinearAppResponse._();
  @$core.pragma('dart2js:noInline')
  static UpdateLinearAppResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<UpdateLinearAppResponse>(
          UpdateLinearAppResponse.$_createMessage);
  static UpdateLinearAppResponse? _defaultInstance;

  @$pb.TagNumber(1)
  LinearApp get app => $_getN(0);
  @$pb.TagNumber(1)
  set app(LinearApp value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasApp() => $_has(0);
  @$pb.TagNumber(1)
  void clearApp() => $_clearField(1);
  @$pb.TagNumber(1)
  LinearApp ensureApp() => $_ensure(0);
}

class PutLinearAppCredentialsRequest extends $pb.GeneratedMessage {
  factory PutLinearAppCredentialsRequest({
    $core.String? appId,
    $core.String? clientSecret,
    $core.String? webhookSecret,
  }) {
    final result = PutLinearAppCredentialsRequest._();
    if (appId != null) result.appId = appId;
    if (clientSecret != null) result.clientSecret = clientSecret;
    if (webhookSecret != null) result.webhookSecret = webhookSecret;
    return result;
  }

  PutLinearAppCredentialsRequest._();

  factory PutLinearAppCredentialsRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      PutLinearAppCredentialsRequest()..mergeFromBuffer(data, registry);
  factory PutLinearAppCredentialsRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      PutLinearAppCredentialsRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'PutLinearAppCredentialsRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: PutLinearAppCredentialsRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'appId')
    ..aOS(2, _omitFieldNames ? '' : 'clientSecret')
    ..aOS(3, _omitFieldNames ? '' : 'webhookSecret')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  PutLinearAppCredentialsRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  PutLinearAppCredentialsRequest copyWith(
          void Function(PutLinearAppCredentialsRequest) updates) =>
      super.copyWith(
              (message) => updates(message as PutLinearAppCredentialsRequest))
          as PutLinearAppCredentialsRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use PutLinearAppCredentialsRequest() / PutLinearAppCredentialsRequest.new instead')
  static PutLinearAppCredentialsRequest create() =>
      PutLinearAppCredentialsRequest._();
  static $pb.GeneratedMessage $_createMessage() =>
      PutLinearAppCredentialsRequest._();
  @$core.override
  PutLinearAppCredentialsRequest createEmptyInstance() =>
      PutLinearAppCredentialsRequest._();
  @$core.pragma('dart2js:noInline')
  static PutLinearAppCredentialsRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<PutLinearAppCredentialsRequest>(
          PutLinearAppCredentialsRequest.$_createMessage);
  static PutLinearAppCredentialsRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get appId => $_getSZ(0);
  @$pb.TagNumber(1)
  set appId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasAppId() => $_has(0);
  @$pb.TagNumber(1)
  void clearAppId() => $_clearField(1);

  /// Absent keeps the stored secret; empty clears it.
  @$pb.TagNumber(2)
  $core.String get clientSecret => $_getSZ(1);
  @$pb.TagNumber(2)
  set clientSecret($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasClientSecret() => $_has(1);
  @$pb.TagNumber(2)
  void clearClientSecret() => $_clearField(2);

  /// Absent keeps the stored secret; empty clears it.
  @$pb.TagNumber(3)
  $core.String get webhookSecret => $_getSZ(2);
  @$pb.TagNumber(3)
  set webhookSecret($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasWebhookSecret() => $_has(2);
  @$pb.TagNumber(3)
  void clearWebhookSecret() => $_clearField(3);
}

class PutLinearAppCredentialsResponse extends $pb.GeneratedMessage {
  factory PutLinearAppCredentialsResponse({
    LinearApp? app,
  }) {
    final result = PutLinearAppCredentialsResponse._();
    if (app != null) result.app = app;
    return result;
  }

  PutLinearAppCredentialsResponse._();

  factory PutLinearAppCredentialsResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      PutLinearAppCredentialsResponse()..mergeFromBuffer(data, registry);
  factory PutLinearAppCredentialsResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      PutLinearAppCredentialsResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'PutLinearAppCredentialsResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: PutLinearAppCredentialsResponse.$_createMessage)
    ..aOM<LinearApp>(1, _omitFieldNames ? '' : 'app',
        subBuilder: LinearApp.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  PutLinearAppCredentialsResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  PutLinearAppCredentialsResponse copyWith(
          void Function(PutLinearAppCredentialsResponse) updates) =>
      super.copyWith(
              (message) => updates(message as PutLinearAppCredentialsResponse))
          as PutLinearAppCredentialsResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use PutLinearAppCredentialsResponse() / PutLinearAppCredentialsResponse.new instead')
  static PutLinearAppCredentialsResponse create() =>
      PutLinearAppCredentialsResponse._();
  static $pb.GeneratedMessage $_createMessage() =>
      PutLinearAppCredentialsResponse._();
  @$core.override
  PutLinearAppCredentialsResponse createEmptyInstance() =>
      PutLinearAppCredentialsResponse._();
  @$core.pragma('dart2js:noInline')
  static PutLinearAppCredentialsResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<PutLinearAppCredentialsResponse>(
          PutLinearAppCredentialsResponse.$_createMessage);
  static PutLinearAppCredentialsResponse? _defaultInstance;

  @$pb.TagNumber(1)
  LinearApp get app => $_getN(0);
  @$pb.TagNumber(1)
  set app(LinearApp value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasApp() => $_has(0);
  @$pb.TagNumber(1)
  void clearApp() => $_clearField(1);
  @$pb.TagNumber(1)
  LinearApp ensureApp() => $_ensure(0);
}

class DeleteLinearAppRequest extends $pb.GeneratedMessage {
  factory DeleteLinearAppRequest({
    $core.String? id,
  }) {
    final result = DeleteLinearAppRequest._();
    if (id != null) result.id = id;
    return result;
  }

  DeleteLinearAppRequest._();

  factory DeleteLinearAppRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DeleteLinearAppRequest()..mergeFromBuffer(data, registry);
  factory DeleteLinearAppRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DeleteLinearAppRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'DeleteLinearAppRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: DeleteLinearAppRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteLinearAppRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteLinearAppRequest copyWith(
          void Function(DeleteLinearAppRequest) updates) =>
      super.copyWith((message) => updates(message as DeleteLinearAppRequest))
          as DeleteLinearAppRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use DeleteLinearAppRequest() / DeleteLinearAppRequest.new instead')
  static DeleteLinearAppRequest create() => DeleteLinearAppRequest._();
  static $pb.GeneratedMessage $_createMessage() => DeleteLinearAppRequest._();
  @$core.override
  DeleteLinearAppRequest createEmptyInstance() => DeleteLinearAppRequest._();
  @$core.pragma('dart2js:noInline')
  static DeleteLinearAppRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<DeleteLinearAppRequest>(
          DeleteLinearAppRequest.$_createMessage);
  static DeleteLinearAppRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);
}

class DeleteLinearAppResponse extends $pb.GeneratedMessage {
  factory DeleteLinearAppResponse() => DeleteLinearAppResponse._();

  DeleteLinearAppResponse._();

  factory DeleteLinearAppResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DeleteLinearAppResponse()..mergeFromBuffer(data, registry);
  factory DeleteLinearAppResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DeleteLinearAppResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'DeleteLinearAppResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: DeleteLinearAppResponse.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteLinearAppResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteLinearAppResponse copyWith(
          void Function(DeleteLinearAppResponse) updates) =>
      super.copyWith((message) => updates(message as DeleteLinearAppResponse))
          as DeleteLinearAppResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use DeleteLinearAppResponse() / DeleteLinearAppResponse.new instead')
  static DeleteLinearAppResponse create() => DeleteLinearAppResponse._();
  static $pb.GeneratedMessage $_createMessage() => DeleteLinearAppResponse._();
  @$core.override
  DeleteLinearAppResponse createEmptyInstance() => DeleteLinearAppResponse._();
  @$core.pragma('dart2js:noInline')
  static DeleteLinearAppResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<DeleteLinearAppResponse>(
          DeleteLinearAppResponse.$_createMessage);
  static DeleteLinearAppResponse? _defaultInstance;
}

/// LinearInstallation is a Linear App installed into one Linear
/// organization. Its access and refresh tokens are not proto fields: they sit
/// behind the same credential seam as the App's secrets.
class LinearInstallation extends $pb.GeneratedMessage {
  factory LinearInstallation({
    $core.String? id,
    $core.String? appId,
    $core.String? organizationId,
    $core.String? organizationName,
    $core.String? appUserId,
    $core.Iterable<$core.String>? scopes,
    LinearInstallationCredentialState? credentialState,
    $core.String? lastCredentialError,
    $0.Timestamp? installedAt,
    $0.Timestamp? updatedAt,
    $core.String? workspaceId,
  }) {
    final result = LinearInstallation._();
    if (id != null) result.id = id;
    if (appId != null) result.appId = appId;
    if (organizationId != null) result.organizationId = organizationId;
    if (organizationName != null) result.organizationName = organizationName;
    if (appUserId != null) result.appUserId = appUserId;
    if (scopes != null) result.scopes.addAll(scopes);
    if (credentialState != null) result.credentialState = credentialState;
    if (lastCredentialError != null)
      result.lastCredentialError = lastCredentialError;
    if (installedAt != null) result.installedAt = installedAt;
    if (updatedAt != null) result.updatedAt = updatedAt;
    if (workspaceId != null) result.workspaceId = workspaceId;
    return result;
  }

  LinearInstallation._();

  factory LinearInstallation.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      LinearInstallation()..mergeFromBuffer(data, registry);
  factory LinearInstallation.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      LinearInstallation()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'LinearInstallation',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: LinearInstallation.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..aOS(2, _omitFieldNames ? '' : 'appId')
    ..aOS(3, _omitFieldNames ? '' : 'organizationId')
    ..aOS(4, _omitFieldNames ? '' : 'organizationName')
    ..aOS(5, _omitFieldNames ? '' : 'appUserId')
    ..pPS(6, _omitFieldNames ? '' : 'scopes')
    ..aE<LinearInstallationCredentialState>(
        7, _omitFieldNames ? '' : 'credentialState',
        enumValues: LinearInstallationCredentialState.values)
    ..aOS(8, _omitFieldNames ? '' : 'lastCredentialError')
    ..aOM<$0.Timestamp>(9, _omitFieldNames ? '' : 'installedAt',
        subBuilder: $0.Timestamp.$_createMessage)
    ..aOM<$0.Timestamp>(10, _omitFieldNames ? '' : 'updatedAt',
        subBuilder: $0.Timestamp.$_createMessage)
    ..aOS(100, _omitFieldNames ? '' : 'workspaceId')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  LinearInstallation clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  LinearInstallation copyWith(void Function(LinearInstallation) updates) =>
      super.copyWith((message) => updates(message as LinearInstallation))
          as LinearInstallation;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use LinearInstallation() / LinearInstallation.new instead')
  static LinearInstallation create() => LinearInstallation._();
  static $pb.GeneratedMessage $_createMessage() => LinearInstallation._();
  @$core.override
  LinearInstallation createEmptyInstance() => LinearInstallation._();
  @$core.pragma('dart2js:noInline')
  static LinearInstallation getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<LinearInstallation>(
          LinearInstallation.$_createMessage);
  static LinearInstallation? _defaultInstance;

  /// @gotags: json:"id,omitempty" bson:"id,omitempty"
  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);

  /// @gotags: json:"app_id,omitempty" bson:"app_id,omitempty"
  @$pb.TagNumber(2)
  $core.String get appId => $_getSZ(1);
  @$pb.TagNumber(2)
  set appId($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasAppId() => $_has(1);
  @$pb.TagNumber(2)
  void clearAppId() => $_clearField(2);

  /// The Linear organization (Linear calls it a workspace) the App is
  /// installed in. Unique per App.
  /// @gotags: json:"organization_id,omitempty" bson:"organization_id,omitempty"
  @$pb.TagNumber(3)
  $core.String get organizationId => $_getSZ(2);
  @$pb.TagNumber(3)
  set organizationId($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasOrganizationId() => $_has(2);
  @$pb.TagNumber(3)
  void clearOrganizationId() => $_clearField(3);

  /// @gotags: json:"organization_name,omitempty" bson:"organization_name,omitempty"
  @$pb.TagNumber(4)
  $core.String get organizationName => $_getSZ(3);
  @$pb.TagNumber(4)
  set organizationName($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasOrganizationName() => $_has(3);
  @$pb.TagNumber(4)
  void clearOrganizationName() => $_clearField(4);

  /// The App's own user in that organization (its `viewer.id`), which is
  /// what people delegate issues to and mention.
  /// @gotags: json:"app_user_id,omitempty" bson:"app_user_id,omitempty"
  @$pb.TagNumber(5)
  $core.String get appUserId => $_getSZ(4);
  @$pb.TagNumber(5)
  set appUserId($core.String value) => $_setString(4, value);
  @$pb.TagNumber(5)
  $core.bool hasAppUserId() => $_has(4);
  @$pb.TagNumber(5)
  void clearAppUserId() => $_clearField(5);

  /// Scopes Linear granted.
  /// @gotags: json:"scopes,omitempty" bson:"scopes,omitempty"
  @$pb.TagNumber(6)
  $pb.PbList<$core.String> get scopes => $_getList(5);

  /// Server-owned. Never carries credential material.
  /// @gotags: json:"credential_state,omitempty" bson:"credential_state,omitempty"
  @$pb.TagNumber(7)
  LinearInstallationCredentialState get credentialState => $_getN(6);
  @$pb.TagNumber(7)
  set credentialState(LinearInstallationCredentialState value) =>
      $_setField(7, value);
  @$pb.TagNumber(7)
  $core.bool hasCredentialState() => $_has(6);
  @$pb.TagNumber(7)
  void clearCredentialState() => $_clearField(7);

  /// Sanitized reason the installation needs reinstalling. Empty otherwise.
  /// @gotags: json:"last_credential_error,omitempty" bson:"last_credential_error,omitempty"
  @$pb.TagNumber(8)
  $core.String get lastCredentialError => $_getSZ(7);
  @$pb.TagNumber(8)
  set lastCredentialError($core.String value) => $_setString(7, value);
  @$pb.TagNumber(8)
  $core.bool hasLastCredentialError() => $_has(7);
  @$pb.TagNumber(8)
  void clearLastCredentialError() => $_clearField(8);

  /// First install time; kept when the App is installed again.
  /// @gotags: json:"installed_at,omitempty" bson:"installed_at,omitempty"
  @$pb.TagNumber(9)
  $0.Timestamp get installedAt => $_getN(8);
  @$pb.TagNumber(9)
  set installedAt($0.Timestamp value) => $_setField(9, value);
  @$pb.TagNumber(9)
  $core.bool hasInstalledAt() => $_has(8);
  @$pb.TagNumber(9)
  void clearInstalledAt() => $_clearField(9);
  @$pb.TagNumber(9)
  $0.Timestamp ensureInstalledAt() => $_ensure(8);

  /// @gotags: json:"updated_at,omitempty" bson:"updated_at,omitempty"
  @$pb.TagNumber(10)
  $0.Timestamp get updatedAt => $_getN(9);
  @$pb.TagNumber(10)
  set updatedAt($0.Timestamp value) => $_setField(10, value);
  @$pb.TagNumber(10)
  $core.bool hasUpdatedAt() => $_has(9);
  @$pb.TagNumber(10)
  void clearUpdatedAt() => $_clearField(10);
  @$pb.TagNumber(10)
  $0.Timestamp ensureUpdatedAt() => $_ensure(9);

  /// @gotags: json:"workspace_id,omitempty" bson:"workspace_id,omitempty"
  @$pb.TagNumber(100)
  $core.String get workspaceId => $_getSZ(10);
  @$pb.TagNumber(100)
  set workspaceId($core.String value) => $_setString(10, value);
  @$pb.TagNumber(100)
  $core.bool hasWorkspaceId() => $_has(10);
  @$pb.TagNumber(100)
  void clearWorkspaceId() => $_clearField(100);
}

class BeginLinearInstallRequest extends $pb.GeneratedMessage {
  factory BeginLinearInstallRequest({
    $core.String? appId,
    $core.String? returnUrl,
  }) {
    final result = BeginLinearInstallRequest._();
    if (appId != null) result.appId = appId;
    if (returnUrl != null) result.returnUrl = returnUrl;
    return result;
  }

  BeginLinearInstallRequest._();

  factory BeginLinearInstallRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      BeginLinearInstallRequest()..mergeFromBuffer(data, registry);
  factory BeginLinearInstallRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      BeginLinearInstallRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'BeginLinearInstallRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: BeginLinearInstallRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'appId')
    ..aOS(2, _omitFieldNames ? '' : 'returnUrl')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  BeginLinearInstallRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  BeginLinearInstallRequest copyWith(
          void Function(BeginLinearInstallRequest) updates) =>
      super.copyWith((message) => updates(message as BeginLinearInstallRequest))
          as BeginLinearInstallRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use BeginLinearInstallRequest() / BeginLinearInstallRequest.new instead')
  static BeginLinearInstallRequest create() => BeginLinearInstallRequest._();
  static $pb.GeneratedMessage $_createMessage() =>
      BeginLinearInstallRequest._();
  @$core.override
  BeginLinearInstallRequest createEmptyInstance() =>
      BeginLinearInstallRequest._();
  @$core.pragma('dart2js:noInline')
  static BeginLinearInstallRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<BeginLinearInstallRequest>(
          BeginLinearInstallRequest.$_createMessage);
  static BeginLinearInstallRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get appId => $_getSZ(0);
  @$pb.TagNumber(1)
  set appId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasAppId() => $_has(0);
  @$pb.TagNumber(1)
  void clearAppId() => $_clearField(1);

  /// Where the browser lands after the install: a relative path, or an
  /// absolute URL on the configured Linear base URL's or dashboard's origin.
  /// Anything else falls back to the App's page.
  @$pb.TagNumber(2)
  $core.String get returnUrl => $_getSZ(1);
  @$pb.TagNumber(2)
  set returnUrl($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasReturnUrl() => $_has(1);
  @$pb.TagNumber(2)
  void clearReturnUrl() => $_clearField(2);
}

class BeginLinearInstallResponse extends $pb.GeneratedMessage {
  factory BeginLinearInstallResponse({
    $core.String? authorizeUrl,
  }) {
    final result = BeginLinearInstallResponse._();
    if (authorizeUrl != null) result.authorizeUrl = authorizeUrl;
    return result;
  }

  BeginLinearInstallResponse._();

  factory BeginLinearInstallResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      BeginLinearInstallResponse()..mergeFromBuffer(data, registry);
  factory BeginLinearInstallResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      BeginLinearInstallResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'BeginLinearInstallResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: BeginLinearInstallResponse.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'authorizeUrl')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  BeginLinearInstallResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  BeginLinearInstallResponse copyWith(
          void Function(BeginLinearInstallResponse) updates) =>
      super.copyWith(
              (message) => updates(message as BeginLinearInstallResponse))
          as BeginLinearInstallResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use BeginLinearInstallResponse() / BeginLinearInstallResponse.new instead')
  static BeginLinearInstallResponse create() => BeginLinearInstallResponse._();
  static $pb.GeneratedMessage $_createMessage() =>
      BeginLinearInstallResponse._();
  @$core.override
  BeginLinearInstallResponse createEmptyInstance() =>
      BeginLinearInstallResponse._();
  @$core.pragma('dart2js:noInline')
  static BeginLinearInstallResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<BeginLinearInstallResponse>(
          BeginLinearInstallResponse.$_createMessage);
  static BeginLinearInstallResponse? _defaultInstance;

  /// Linear's authorize URL to open in the browser.
  @$pb.TagNumber(1)
  $core.String get authorizeUrl => $_getSZ(0);
  @$pb.TagNumber(1)
  set authorizeUrl($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasAuthorizeUrl() => $_has(0);
  @$pb.TagNumber(1)
  void clearAuthorizeUrl() => $_clearField(1);
}

class ListLinearInstallationsRequest extends $pb.GeneratedMessage {
  factory ListLinearInstallationsRequest({
    $core.String? appId,
  }) {
    final result = ListLinearInstallationsRequest._();
    if (appId != null) result.appId = appId;
    return result;
  }

  ListLinearInstallationsRequest._();

  factory ListLinearInstallationsRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListLinearInstallationsRequest()..mergeFromBuffer(data, registry);
  factory ListLinearInstallationsRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListLinearInstallationsRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListLinearInstallationsRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: ListLinearInstallationsRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'appId')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListLinearInstallationsRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListLinearInstallationsRequest copyWith(
          void Function(ListLinearInstallationsRequest) updates) =>
      super.copyWith(
              (message) => updates(message as ListLinearInstallationsRequest))
          as ListLinearInstallationsRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use ListLinearInstallationsRequest() / ListLinearInstallationsRequest.new instead')
  static ListLinearInstallationsRequest create() =>
      ListLinearInstallationsRequest._();
  static $pb.GeneratedMessage $_createMessage() =>
      ListLinearInstallationsRequest._();
  @$core.override
  ListLinearInstallationsRequest createEmptyInstance() =>
      ListLinearInstallationsRequest._();
  @$core.pragma('dart2js:noInline')
  static ListLinearInstallationsRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ListLinearInstallationsRequest>(
          ListLinearInstallationsRequest.$_createMessage);
  static ListLinearInstallationsRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get appId => $_getSZ(0);
  @$pb.TagNumber(1)
  set appId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasAppId() => $_has(0);
  @$pb.TagNumber(1)
  void clearAppId() => $_clearField(1);
}

class ListLinearInstallationsResponse extends $pb.GeneratedMessage {
  factory ListLinearInstallationsResponse({
    $core.Iterable<LinearInstallation>? installations,
  }) {
    final result = ListLinearInstallationsResponse._();
    if (installations != null) result.installations.addAll(installations);
    return result;
  }

  ListLinearInstallationsResponse._();

  factory ListLinearInstallationsResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListLinearInstallationsResponse()..mergeFromBuffer(data, registry);
  factory ListLinearInstallationsResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListLinearInstallationsResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListLinearInstallationsResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: ListLinearInstallationsResponse.$_createMessage)
    ..pPM<LinearInstallation>(1, _omitFieldNames ? '' : 'installations',
        subBuilder: LinearInstallation.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListLinearInstallationsResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListLinearInstallationsResponse copyWith(
          void Function(ListLinearInstallationsResponse) updates) =>
      super.copyWith(
              (message) => updates(message as ListLinearInstallationsResponse))
          as ListLinearInstallationsResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use ListLinearInstallationsResponse() / ListLinearInstallationsResponse.new instead')
  static ListLinearInstallationsResponse create() =>
      ListLinearInstallationsResponse._();
  static $pb.GeneratedMessage $_createMessage() =>
      ListLinearInstallationsResponse._();
  @$core.override
  ListLinearInstallationsResponse createEmptyInstance() =>
      ListLinearInstallationsResponse._();
  @$core.pragma('dart2js:noInline')
  static ListLinearInstallationsResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ListLinearInstallationsResponse>(
          ListLinearInstallationsResponse.$_createMessage);
  static ListLinearInstallationsResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $pb.PbList<LinearInstallation> get installations => $_getList(0);
}

class DeleteLinearInstallationRequest extends $pb.GeneratedMessage {
  factory DeleteLinearInstallationRequest({
    $core.String? appId,
    $core.String? id,
  }) {
    final result = DeleteLinearInstallationRequest._();
    if (appId != null) result.appId = appId;
    if (id != null) result.id = id;
    return result;
  }

  DeleteLinearInstallationRequest._();

  factory DeleteLinearInstallationRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DeleteLinearInstallationRequest()..mergeFromBuffer(data, registry);
  factory DeleteLinearInstallationRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DeleteLinearInstallationRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'DeleteLinearInstallationRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: DeleteLinearInstallationRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'appId')
    ..aOS(2, _omitFieldNames ? '' : 'id')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteLinearInstallationRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteLinearInstallationRequest copyWith(
          void Function(DeleteLinearInstallationRequest) updates) =>
      super.copyWith(
              (message) => updates(message as DeleteLinearInstallationRequest))
          as DeleteLinearInstallationRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use DeleteLinearInstallationRequest() / DeleteLinearInstallationRequest.new instead')
  static DeleteLinearInstallationRequest create() =>
      DeleteLinearInstallationRequest._();
  static $pb.GeneratedMessage $_createMessage() =>
      DeleteLinearInstallationRequest._();
  @$core.override
  DeleteLinearInstallationRequest createEmptyInstance() =>
      DeleteLinearInstallationRequest._();
  @$core.pragma('dart2js:noInline')
  static DeleteLinearInstallationRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<DeleteLinearInstallationRequest>(
          DeleteLinearInstallationRequest.$_createMessage);
  static DeleteLinearInstallationRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get appId => $_getSZ(0);
  @$pb.TagNumber(1)
  set appId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasAppId() => $_has(0);
  @$pb.TagNumber(1)
  void clearAppId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get id => $_getSZ(1);
  @$pb.TagNumber(2)
  set id($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasId() => $_has(1);
  @$pb.TagNumber(2)
  void clearId() => $_clearField(2);
}

class DeleteLinearInstallationResponse extends $pb.GeneratedMessage {
  factory DeleteLinearInstallationResponse() =>
      DeleteLinearInstallationResponse._();

  DeleteLinearInstallationResponse._();

  factory DeleteLinearInstallationResponse.fromBuffer(
          $core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DeleteLinearInstallationResponse()..mergeFromBuffer(data, registry);
  factory DeleteLinearInstallationResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DeleteLinearInstallationResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'DeleteLinearInstallationResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: DeleteLinearInstallationResponse.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteLinearInstallationResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteLinearInstallationResponse copyWith(
          void Function(DeleteLinearInstallationResponse) updates) =>
      super.copyWith(
              (message) => updates(message as DeleteLinearInstallationResponse))
          as DeleteLinearInstallationResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use DeleteLinearInstallationResponse() / DeleteLinearInstallationResponse.new instead')
  static DeleteLinearInstallationResponse create() =>
      DeleteLinearInstallationResponse._();
  static $pb.GeneratedMessage $_createMessage() =>
      DeleteLinearInstallationResponse._();
  @$core.override
  DeleteLinearInstallationResponse createEmptyInstance() =>
      DeleteLinearInstallationResponse._();
  @$core.pragma('dart2js:noInline')
  static DeleteLinearInstallationResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<DeleteLinearInstallationResponse>(
          DeleteLinearInstallationResponse.$_createMessage);
  static DeleteLinearInstallationResponse? _defaultInstance;
}

/// LinearSettings holds platform-level Linear configuration.
class LinearSettings extends $pb.GeneratedMessage {
  factory LinearSettings({
    $core.String? publicBaseUrl,
    $0.Timestamp? updatedAt,
  }) {
    final result = LinearSettings._();
    if (publicBaseUrl != null) result.publicBaseUrl = publicBaseUrl;
    if (updatedAt != null) result.updatedAt = updatedAt;
    return result;
  }

  LinearSettings._();

  factory LinearSettings.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      LinearSettings()..mergeFromBuffer(data, registry);
  factory LinearSettings.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      LinearSettings()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'LinearSettings',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: LinearSettings.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'publicBaseUrl')
    ..aOM<$0.Timestamp>(2, _omitFieldNames ? '' : 'updatedAt',
        subBuilder: $0.Timestamp.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  LinearSettings clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  LinearSettings copyWith(void Function(LinearSettings) updates) =>
      super.copyWith((message) => updates(message as LinearSettings))
          as LinearSettings;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use LinearSettings() / LinearSettings.new instead')
  static LinearSettings create() => LinearSettings._();
  static $pb.GeneratedMessage $_createMessage() => LinearSettings._();
  @$core.override
  LinearSettings createEmptyInstance() => LinearSettings._();
  @$core.pragma('dart2js:noInline')
  static LinearSettings getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<LinearSettings>(
          LinearSettings.$_createMessage);
  static LinearSettings? _defaultInstance;

  /// Public base URL Linear reaches this deployment at, e.g.
  /// "https://butter.example.com". Callback and webhook URLs are derived from
  /// it and immutable IDs.
  /// @gotags: json:"public_base_url,omitempty" bson:"public_base_url,omitempty"
  @$pb.TagNumber(1)
  $core.String get publicBaseUrl => $_getSZ(0);
  @$pb.TagNumber(1)
  set publicBaseUrl($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasPublicBaseUrl() => $_has(0);
  @$pb.TagNumber(1)
  void clearPublicBaseUrl() => $_clearField(1);

  /// @gotags: json:"updated_at,omitempty" bson:"updated_at,omitempty"
  @$pb.TagNumber(2)
  $0.Timestamp get updatedAt => $_getN(1);
  @$pb.TagNumber(2)
  set updatedAt($0.Timestamp value) => $_setField(2, value);
  @$pb.TagNumber(2)
  $core.bool hasUpdatedAt() => $_has(1);
  @$pb.TagNumber(2)
  void clearUpdatedAt() => $_clearField(2);
  @$pb.TagNumber(2)
  $0.Timestamp ensureUpdatedAt() => $_ensure(1);
}

class GetLinearSettingsRequest extends $pb.GeneratedMessage {
  factory GetLinearSettingsRequest() => GetLinearSettingsRequest._();

  GetLinearSettingsRequest._();

  factory GetLinearSettingsRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetLinearSettingsRequest()..mergeFromBuffer(data, registry);
  factory GetLinearSettingsRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetLinearSettingsRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetLinearSettingsRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: GetLinearSettingsRequest.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetLinearSettingsRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetLinearSettingsRequest copyWith(
          void Function(GetLinearSettingsRequest) updates) =>
      super.copyWith((message) => updates(message as GetLinearSettingsRequest))
          as GetLinearSettingsRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use GetLinearSettingsRequest() / GetLinearSettingsRequest.new instead')
  static GetLinearSettingsRequest create() => GetLinearSettingsRequest._();
  static $pb.GeneratedMessage $_createMessage() => GetLinearSettingsRequest._();
  @$core.override
  GetLinearSettingsRequest createEmptyInstance() =>
      GetLinearSettingsRequest._();
  @$core.pragma('dart2js:noInline')
  static GetLinearSettingsRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<GetLinearSettingsRequest>(
          GetLinearSettingsRequest.$_createMessage);
  static GetLinearSettingsRequest? _defaultInstance;
}

class GetLinearSettingsResponse extends $pb.GeneratedMessage {
  factory GetLinearSettingsResponse({
    LinearSettings? settings,
  }) {
    final result = GetLinearSettingsResponse._();
    if (settings != null) result.settings = settings;
    return result;
  }

  GetLinearSettingsResponse._();

  factory GetLinearSettingsResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetLinearSettingsResponse()..mergeFromBuffer(data, registry);
  factory GetLinearSettingsResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetLinearSettingsResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetLinearSettingsResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: GetLinearSettingsResponse.$_createMessage)
    ..aOM<LinearSettings>(1, _omitFieldNames ? '' : 'settings',
        subBuilder: LinearSettings.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetLinearSettingsResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetLinearSettingsResponse copyWith(
          void Function(GetLinearSettingsResponse) updates) =>
      super.copyWith((message) => updates(message as GetLinearSettingsResponse))
          as GetLinearSettingsResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use GetLinearSettingsResponse() / GetLinearSettingsResponse.new instead')
  static GetLinearSettingsResponse create() => GetLinearSettingsResponse._();
  static $pb.GeneratedMessage $_createMessage() =>
      GetLinearSettingsResponse._();
  @$core.override
  GetLinearSettingsResponse createEmptyInstance() =>
      GetLinearSettingsResponse._();
  @$core.pragma('dart2js:noInline')
  static GetLinearSettingsResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<GetLinearSettingsResponse>(
          GetLinearSettingsResponse.$_createMessage);
  static GetLinearSettingsResponse? _defaultInstance;

  @$pb.TagNumber(1)
  LinearSettings get settings => $_getN(0);
  @$pb.TagNumber(1)
  set settings(LinearSettings value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasSettings() => $_has(0);
  @$pb.TagNumber(1)
  void clearSettings() => $_clearField(1);
  @$pb.TagNumber(1)
  LinearSettings ensureSettings() => $_ensure(0);
}

class UpdateLinearSettingsRequest extends $pb.GeneratedMessage {
  factory UpdateLinearSettingsRequest({
    LinearSettings? settings,
  }) {
    final result = UpdateLinearSettingsRequest._();
    if (settings != null) result.settings = settings;
    return result;
  }

  UpdateLinearSettingsRequest._();

  factory UpdateLinearSettingsRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpdateLinearSettingsRequest()..mergeFromBuffer(data, registry);
  factory UpdateLinearSettingsRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpdateLinearSettingsRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'UpdateLinearSettingsRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: UpdateLinearSettingsRequest.$_createMessage)
    ..aOM<LinearSettings>(1, _omitFieldNames ? '' : 'settings',
        subBuilder: LinearSettings.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateLinearSettingsRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateLinearSettingsRequest copyWith(
          void Function(UpdateLinearSettingsRequest) updates) =>
      super.copyWith(
              (message) => updates(message as UpdateLinearSettingsRequest))
          as UpdateLinearSettingsRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use UpdateLinearSettingsRequest() / UpdateLinearSettingsRequest.new instead')
  static UpdateLinearSettingsRequest create() =>
      UpdateLinearSettingsRequest._();
  static $pb.GeneratedMessage $_createMessage() =>
      UpdateLinearSettingsRequest._();
  @$core.override
  UpdateLinearSettingsRequest createEmptyInstance() =>
      UpdateLinearSettingsRequest._();
  @$core.pragma('dart2js:noInline')
  static UpdateLinearSettingsRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<UpdateLinearSettingsRequest>(
          UpdateLinearSettingsRequest.$_createMessage);
  static UpdateLinearSettingsRequest? _defaultInstance;

  @$pb.TagNumber(1)
  LinearSettings get settings => $_getN(0);
  @$pb.TagNumber(1)
  set settings(LinearSettings value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasSettings() => $_has(0);
  @$pb.TagNumber(1)
  void clearSettings() => $_clearField(1);
  @$pb.TagNumber(1)
  LinearSettings ensureSettings() => $_ensure(0);
}

class UpdateLinearSettingsResponse extends $pb.GeneratedMessage {
  factory UpdateLinearSettingsResponse({
    LinearSettings? settings,
  }) {
    final result = UpdateLinearSettingsResponse._();
    if (settings != null) result.settings = settings;
    return result;
  }

  UpdateLinearSettingsResponse._();

  factory UpdateLinearSettingsResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpdateLinearSettingsResponse()..mergeFromBuffer(data, registry);
  factory UpdateLinearSettingsResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpdateLinearSettingsResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'UpdateLinearSettingsResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: UpdateLinearSettingsResponse.$_createMessage)
    ..aOM<LinearSettings>(1, _omitFieldNames ? '' : 'settings',
        subBuilder: LinearSettings.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateLinearSettingsResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateLinearSettingsResponse copyWith(
          void Function(UpdateLinearSettingsResponse) updates) =>
      super.copyWith(
              (message) => updates(message as UpdateLinearSettingsResponse))
          as UpdateLinearSettingsResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use UpdateLinearSettingsResponse() / UpdateLinearSettingsResponse.new instead')
  static UpdateLinearSettingsResponse create() =>
      UpdateLinearSettingsResponse._();
  static $pb.GeneratedMessage $_createMessage() =>
      UpdateLinearSettingsResponse._();
  @$core.override
  UpdateLinearSettingsResponse createEmptyInstance() =>
      UpdateLinearSettingsResponse._();
  @$core.pragma('dart2js:noInline')
  static UpdateLinearSettingsResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<UpdateLinearSettingsResponse>(
          UpdateLinearSettingsResponse.$_createMessage);
  static UpdateLinearSettingsResponse? _defaultInstance;

  @$pb.TagNumber(1)
  LinearSettings get settings => $_getN(0);
  @$pb.TagNumber(1)
  set settings(LinearSettings value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasSettings() => $_has(0);
  @$pb.TagNumber(1)
  void clearSettings() => $_clearField(1);
  @$pb.TagNumber(1)
  LinearSettings ensureSettings() => $_ensure(0);
}

/// LinearProcessingRecord is the auditable state of one accepted delivery.
class LinearProcessingRecord extends $pb.GeneratedMessage {
  factory LinearProcessingRecord({
    $core.String? id,
    $core.String? workspaceId,
    $core.String? appId,
    $core.String? installationId,
    $core.String? agentSessionId,
    $core.String? action,
    $core.String? deliveryId,
    LinearProcessingStatus? status,
    $core.String? invocationId,
    $core.int? attempts,
    $fixnum.Int64? appRevision,
    $core.String? error,
    $core.String? output,
    $core.String? outputType,
    $core.String? outputSignal,
    $core.String? outputSignalMetadata,
    $core.bool? delivered,
    $core.bool? deadLettered,
    $core.String? promptingUserId,
    $core.String? issueIdentifier,
    $0.Timestamp? createdAt,
    $0.Timestamp? updatedAt,
    $0.Timestamp? expiresAt,
    $core.String? agentId,
  }) {
    final result = LinearProcessingRecord._();
    if (id != null) result.id = id;
    if (workspaceId != null) result.workspaceId = workspaceId;
    if (appId != null) result.appId = appId;
    if (installationId != null) result.installationId = installationId;
    if (agentSessionId != null) result.agentSessionId = agentSessionId;
    if (action != null) result.action = action;
    if (deliveryId != null) result.deliveryId = deliveryId;
    if (status != null) result.status = status;
    if (invocationId != null) result.invocationId = invocationId;
    if (attempts != null) result.attempts = attempts;
    if (appRevision != null) result.appRevision = appRevision;
    if (error != null) result.error = error;
    if (output != null) result.output = output;
    if (outputType != null) result.outputType = outputType;
    if (outputSignal != null) result.outputSignal = outputSignal;
    if (outputSignalMetadata != null)
      result.outputSignalMetadata = outputSignalMetadata;
    if (delivered != null) result.delivered = delivered;
    if (deadLettered != null) result.deadLettered = deadLettered;
    if (promptingUserId != null) result.promptingUserId = promptingUserId;
    if (issueIdentifier != null) result.issueIdentifier = issueIdentifier;
    if (createdAt != null) result.createdAt = createdAt;
    if (updatedAt != null) result.updatedAt = updatedAt;
    if (expiresAt != null) result.expiresAt = expiresAt;
    if (agentId != null) result.agentId = agentId;
    return result;
  }

  LinearProcessingRecord._();

  factory LinearProcessingRecord.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      LinearProcessingRecord()..mergeFromBuffer(data, registry);
  factory LinearProcessingRecord.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      LinearProcessingRecord()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'LinearProcessingRecord',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: LinearProcessingRecord.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..aOS(2, _omitFieldNames ? '' : 'workspaceId')
    ..aOS(3, _omitFieldNames ? '' : 'appId')
    ..aOS(4, _omitFieldNames ? '' : 'installationId')
    ..aOS(5, _omitFieldNames ? '' : 'agentSessionId')
    ..aOS(6, _omitFieldNames ? '' : 'action')
    ..aOS(7, _omitFieldNames ? '' : 'deliveryId')
    ..aE<LinearProcessingStatus>(8, _omitFieldNames ? '' : 'status',
        enumValues: LinearProcessingStatus.values)
    ..aOS(9, _omitFieldNames ? '' : 'invocationId')
    ..aI(10, _omitFieldNames ? '' : 'attempts')
    ..aInt64(11, _omitFieldNames ? '' : 'appRevision')
    ..aOS(12, _omitFieldNames ? '' : 'error')
    ..aOS(13, _omitFieldNames ? '' : 'output')
    ..aOS(14, _omitFieldNames ? '' : 'outputType')
    ..aOS(15, _omitFieldNames ? '' : 'outputSignal')
    ..aOS(16, _omitFieldNames ? '' : 'outputSignalMetadata')
    ..aOB(17, _omitFieldNames ? '' : 'delivered')
    ..aOB(18, _omitFieldNames ? '' : 'deadLettered')
    ..aOS(19, _omitFieldNames ? '' : 'promptingUserId')
    ..aOS(20, _omitFieldNames ? '' : 'issueIdentifier')
    ..aOM<$0.Timestamp>(21, _omitFieldNames ? '' : 'createdAt',
        subBuilder: $0.Timestamp.$_createMessage)
    ..aOM<$0.Timestamp>(22, _omitFieldNames ? '' : 'updatedAt',
        subBuilder: $0.Timestamp.$_createMessage)
    ..aOM<$0.Timestamp>(23, _omitFieldNames ? '' : 'expiresAt',
        subBuilder: $0.Timestamp.$_createMessage)
    ..aOS(24, _omitFieldNames ? '' : 'agentId')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  LinearProcessingRecord clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  LinearProcessingRecord copyWith(
          void Function(LinearProcessingRecord) updates) =>
      super.copyWith((message) => updates(message as LinearProcessingRecord))
          as LinearProcessingRecord;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use LinearProcessingRecord() / LinearProcessingRecord.new instead')
  static LinearProcessingRecord create() => LinearProcessingRecord._();
  static $pb.GeneratedMessage $_createMessage() => LinearProcessingRecord._();
  @$core.override
  LinearProcessingRecord createEmptyInstance() => LinearProcessingRecord._();
  @$core.pragma('dart2js:noInline')
  static LinearProcessingRecord getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<LinearProcessingRecord>(
          LinearProcessingRecord.$_createMessage);
  static LinearProcessingRecord? _defaultInstance;

  /// @gotags: json:"id,omitempty" bson:"id,omitempty"
  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);

  /// @gotags: json:"workspace_id,omitempty" bson:"workspace_id,omitempty"
  @$pb.TagNumber(2)
  $core.String get workspaceId => $_getSZ(1);
  @$pb.TagNumber(2)
  set workspaceId($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasWorkspaceId() => $_has(1);
  @$pb.TagNumber(2)
  void clearWorkspaceId() => $_clearField(2);

  /// @gotags: json:"app_id,omitempty" bson:"app_id,omitempty"
  @$pb.TagNumber(3)
  $core.String get appId => $_getSZ(2);
  @$pb.TagNumber(3)
  set appId($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasAppId() => $_has(2);
  @$pb.TagNumber(3)
  void clearAppId() => $_clearField(3);

  /// @gotags: json:"installation_id,omitempty" bson:"installation_id,omitempty"
  @$pb.TagNumber(4)
  $core.String get installationId => $_getSZ(3);
  @$pb.TagNumber(4)
  set installationId($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasInstallationId() => $_has(3);
  @$pb.TagNumber(4)
  void clearInstallationId() => $_clearField(4);

  /// @gotags: json:"agent_session_id,omitempty" bson:"agent_session_id,omitempty"
  @$pb.TagNumber(5)
  $core.String get agentSessionId => $_getSZ(4);
  @$pb.TagNumber(5)
  set agentSessionId($core.String value) => $_setString(4, value);
  @$pb.TagNumber(5)
  $core.bool hasAgentSessionId() => $_has(4);
  @$pb.TagNumber(5)
  void clearAgentSessionId() => $_clearField(5);

  /// created or prompted.
  /// @gotags: json:"action,omitempty" bson:"action,omitempty"
  @$pb.TagNumber(6)
  $core.String get action => $_getSZ(5);
  @$pb.TagNumber(6)
  set action($core.String value) => $_setString(5, value);
  @$pb.TagNumber(6)
  $core.bool hasAction() => $_has(5);
  @$pb.TagNumber(6)
  void clearAction() => $_clearField(6);

  /// Linear's delivery ID (or the body hash); unique with app_id.
  /// @gotags: json:"delivery_id,omitempty" bson:"delivery_id,omitempty"
  @$pb.TagNumber(7)
  $core.String get deliveryId => $_getSZ(6);
  @$pb.TagNumber(7)
  set deliveryId($core.String value) => $_setString(6, value);
  @$pb.TagNumber(7)
  $core.bool hasDeliveryId() => $_has(6);
  @$pb.TagNumber(7)
  void clearDeliveryId() => $_clearField(7);

  /// @gotags: json:"status,omitempty" bson:"status,omitempty"
  @$pb.TagNumber(8)
  LinearProcessingStatus get status => $_getN(7);
  @$pb.TagNumber(8)
  set status(LinearProcessingStatus value) => $_setField(8, value);
  @$pb.TagNumber(8)
  $core.bool hasStatus() => $_has(7);
  @$pb.TagNumber(8)
  void clearStatus() => $_clearField(8);

  /// Stable across retries; shared by follow-ups run in one turn.
  /// @gotags: json:"invocation_id,omitempty" bson:"invocation_id,omitempty"
  @$pb.TagNumber(9)
  $core.String get invocationId => $_getSZ(8);
  @$pb.TagNumber(9)
  set invocationId($core.String value) => $_setString(8, value);
  @$pb.TagNumber(9)
  $core.bool hasInvocationId() => $_has(8);
  @$pb.TagNumber(9)
  void clearInvocationId() => $_clearField(9);

  /// How many times the record has been claimed.
  /// @gotags: json:"attempts,omitempty" bson:"attempts,omitempty"
  @$pb.TagNumber(10)
  $core.int get attempts => $_getIZ(9);
  @$pb.TagNumber(10)
  set attempts($core.int value) => $_setSignedInt32(9, value);
  @$pb.TagNumber(10)
  $core.bool hasAttempts() => $_has(9);
  @$pb.TagNumber(10)
  void clearAttempts() => $_clearField(10);

  /// App revision frozen at acceptance.
  /// @gotags: json:"app_revision,omitempty" bson:"app_revision,omitempty"
  @$pb.TagNumber(11)
  $fixnum.Int64 get appRevision => $_getI64(10);
  @$pb.TagNumber(11)
  set appRevision($fixnum.Int64 value) => $_setInt64(10, value);
  @$pb.TagNumber(11)
  $core.bool hasAppRevision() => $_has(10);
  @$pb.TagNumber(11)
  void clearAppRevision() => $_clearField(11);

  /// Sanitized failure summary. Never contains credential material.
  /// @gotags: json:"error,omitempty" bson:"error,omitempty"
  @$pb.TagNumber(12)
  $core.String get error => $_getSZ(11);
  @$pb.TagNumber(12)
  set error($core.String value) => $_setString(11, value);
  @$pb.TagNumber(12)
  $core.bool hasError() => $_has(11);
  @$pb.TagNumber(12)
  void clearError() => $_clearField(12);

  /// The complete reply, persisted before it is posted.
  /// @gotags: json:"output,omitempty" bson:"output,omitempty"
  @$pb.TagNumber(13)
  $core.String get output => $_getSZ(12);
  @$pb.TagNumber(13)
  set output($core.String value) => $_setString(12, value);
  @$pb.TagNumber(13)
  $core.bool hasOutput() => $_has(12);
  @$pb.TagNumber(13)
  void clearOutput() => $_clearField(13);

  /// The activity type the reply is posted as: response, elicitation or
  /// error.
  /// @gotags: json:"output_type,omitempty" bson:"output_type,omitempty"
  @$pb.TagNumber(14)
  $core.String get outputType => $_getSZ(13);
  @$pb.TagNumber(14)
  set outputType($core.String value) => $_setString(13, value);
  @$pb.TagNumber(14)
  $core.bool hasOutputType() => $_has(13);
  @$pb.TagNumber(14)
  void clearOutputType() => $_clearField(14);

  /// An elicitation's signal and its metadata as JSON, e.g. select options.
  /// @gotags: json:"output_signal,omitempty" bson:"output_signal,omitempty"
  @$pb.TagNumber(15)
  $core.String get outputSignal => $_getSZ(14);
  @$pb.TagNumber(15)
  set outputSignal($core.String value) => $_setString(14, value);
  @$pb.TagNumber(15)
  $core.bool hasOutputSignal() => $_has(14);
  @$pb.TagNumber(15)
  void clearOutputSignal() => $_clearField(15);

  /// @gotags: json:"output_signal_metadata,omitempty" bson:"output_signal_metadata,omitempty"
  @$pb.TagNumber(16)
  $core.String get outputSignalMetadata => $_getSZ(15);
  @$pb.TagNumber(16)
  set outputSignalMetadata($core.String value) => $_setString(15, value);
  @$pb.TagNumber(16)
  $core.bool hasOutputSignalMetadata() => $_has(15);
  @$pb.TagNumber(16)
  void clearOutputSignalMetadata() => $_clearField(16);

  /// Whether the reply reached Linear.
  /// @gotags: json:"delivered,omitempty" bson:"delivered,omitempty"
  @$pb.TagNumber(17)
  $core.bool get delivered => $_getBF(16);
  @$pb.TagNumber(17)
  set delivered($core.bool value) => $_setBool(16, value);
  @$pb.TagNumber(17)
  $core.bool hasDelivered() => $_has(16);
  @$pb.TagNumber(17)
  void clearDelivered() => $_clearField(17);

  /// Whether the record needs a person.
  /// @gotags: json:"dead_lettered,omitempty" bson:"dead_lettered,omitempty"
  @$pb.TagNumber(18)
  $core.bool get deadLettered => $_getBF(17);
  @$pb.TagNumber(18)
  set deadLettered($core.bool value) => $_setBool(17, value);
  @$pb.TagNumber(18)
  $core.bool hasDeadLettered() => $_has(17);
  @$pb.TagNumber(18)
  void clearDeadLettered() => $_clearField(18);

  /// @gotags: json:"prompting_user_id,omitempty" bson:"prompting_user_id,omitempty"
  @$pb.TagNumber(19)
  $core.String get promptingUserId => $_getSZ(18);
  @$pb.TagNumber(19)
  set promptingUserId($core.String value) => $_setString(18, value);
  @$pb.TagNumber(19)
  $core.bool hasPromptingUserId() => $_has(18);
  @$pb.TagNumber(19)
  void clearPromptingUserId() => $_clearField(19);

  /// @gotags: json:"issue_identifier,omitempty" bson:"issue_identifier,omitempty"
  @$pb.TagNumber(20)
  $core.String get issueIdentifier => $_getSZ(19);
  @$pb.TagNumber(20)
  set issueIdentifier($core.String value) => $_setString(19, value);
  @$pb.TagNumber(20)
  $core.bool hasIssueIdentifier() => $_has(19);
  @$pb.TagNumber(20)
  void clearIssueIdentifier() => $_clearField(20);

  /// @gotags: json:"created_at,omitempty" bson:"created_at,omitempty"
  @$pb.TagNumber(21)
  $0.Timestamp get createdAt => $_getN(20);
  @$pb.TagNumber(21)
  set createdAt($0.Timestamp value) => $_setField(21, value);
  @$pb.TagNumber(21)
  $core.bool hasCreatedAt() => $_has(20);
  @$pb.TagNumber(21)
  void clearCreatedAt() => $_clearField(21);
  @$pb.TagNumber(21)
  $0.Timestamp ensureCreatedAt() => $_ensure(20);

  /// @gotags: json:"updated_at,omitempty" bson:"updated_at,omitempty"
  @$pb.TagNumber(22)
  $0.Timestamp get updatedAt => $_getN(21);
  @$pb.TagNumber(22)
  set updatedAt($0.Timestamp value) => $_setField(22, value);
  @$pb.TagNumber(22)
  $core.bool hasUpdatedAt() => $_has(21);
  @$pb.TagNumber(22)
  void clearUpdatedAt() => $_clearField(22);
  @$pb.TagNumber(22)
  $0.Timestamp ensureUpdatedAt() => $_ensure(21);

  /// When the record becomes eligible for TTL cleanup (30 days).
  /// @gotags: json:"expires_at,omitempty" bson:"expires_at,omitempty"
  @$pb.TagNumber(23)
  $0.Timestamp get expiresAt => $_getN(22);
  @$pb.TagNumber(23)
  set expiresAt($0.Timestamp value) => $_setField(23, value);
  @$pb.TagNumber(23)
  $core.bool hasExpiresAt() => $_has(22);
  @$pb.TagNumber(23)
  void clearExpiresAt() => $_clearField(23);
  @$pb.TagNumber(23)
  $0.Timestamp ensureExpiresAt() => $_ensure(22);

  /// The Agent the turn ran on, set when Agent work starts. A session's
  /// lease is per Agent, so settling a dead holder's records must not touch
  /// another Agent's turn after the App is re-pointed.
  /// @gotags: json:"agent_id,omitempty" bson:"agent_id,omitempty"
  @$pb.TagNumber(24)
  $core.String get agentId => $_getSZ(23);
  @$pb.TagNumber(24)
  set agentId($core.String value) => $_setString(23, value);
  @$pb.TagNumber(24)
  $core.bool hasAgentId() => $_has(23);
  @$pb.TagNumber(24)
  void clearAgentId() => $_clearField(24);
}

class ListLinearProcessingRecordsRequest extends $pb.GeneratedMessage {
  factory ListLinearProcessingRecordsRequest({
    $core.String? appId,
    LinearProcessingStatus? status,
    $core.int? pageSize,
  }) {
    final result = ListLinearProcessingRecordsRequest._();
    if (appId != null) result.appId = appId;
    if (status != null) result.status = status;
    if (pageSize != null) result.pageSize = pageSize;
    return result;
  }

  ListLinearProcessingRecordsRequest._();

  factory ListLinearProcessingRecordsRequest.fromBuffer(
          $core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListLinearProcessingRecordsRequest()..mergeFromBuffer(data, registry);
  factory ListLinearProcessingRecordsRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListLinearProcessingRecordsRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListLinearProcessingRecordsRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: ListLinearProcessingRecordsRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'appId')
    ..aE<LinearProcessingStatus>(2, _omitFieldNames ? '' : 'status',
        enumValues: LinearProcessingStatus.values)
    ..aI(3, _omitFieldNames ? '' : 'pageSize')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListLinearProcessingRecordsRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListLinearProcessingRecordsRequest copyWith(
          void Function(ListLinearProcessingRecordsRequest) updates) =>
      super.copyWith((message) =>
              updates(message as ListLinearProcessingRecordsRequest))
          as ListLinearProcessingRecordsRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use ListLinearProcessingRecordsRequest() / ListLinearProcessingRecordsRequest.new instead')
  static ListLinearProcessingRecordsRequest create() =>
      ListLinearProcessingRecordsRequest._();
  static $pb.GeneratedMessage $_createMessage() =>
      ListLinearProcessingRecordsRequest._();
  @$core.override
  ListLinearProcessingRecordsRequest createEmptyInstance() =>
      ListLinearProcessingRecordsRequest._();
  @$core.pragma('dart2js:noInline')
  static ListLinearProcessingRecordsRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ListLinearProcessingRecordsRequest>(
          ListLinearProcessingRecordsRequest.$_createMessage);
  static ListLinearProcessingRecordsRequest? _defaultInstance;

  /// Optional filters.
  @$pb.TagNumber(1)
  $core.String get appId => $_getSZ(0);
  @$pb.TagNumber(1)
  set appId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasAppId() => $_has(0);
  @$pb.TagNumber(1)
  void clearAppId() => $_clearField(1);

  @$pb.TagNumber(2)
  LinearProcessingStatus get status => $_getN(1);
  @$pb.TagNumber(2)
  set status(LinearProcessingStatus value) => $_setField(2, value);
  @$pb.TagNumber(2)
  $core.bool hasStatus() => $_has(1);
  @$pb.TagNumber(2)
  void clearStatus() => $_clearField(2);

  /// Maximum records to return. Defaults to 50, capped at 200.
  @$pb.TagNumber(3)
  $core.int get pageSize => $_getIZ(2);
  @$pb.TagNumber(3)
  set pageSize($core.int value) => $_setSignedInt32(2, value);
  @$pb.TagNumber(3)
  $core.bool hasPageSize() => $_has(2);
  @$pb.TagNumber(3)
  void clearPageSize() => $_clearField(3);
}

class ListLinearProcessingRecordsResponse extends $pb.GeneratedMessage {
  factory ListLinearProcessingRecordsResponse({
    $core.Iterable<LinearProcessingRecord>? records,
  }) {
    final result = ListLinearProcessingRecordsResponse._();
    if (records != null) result.records.addAll(records);
    return result;
  }

  ListLinearProcessingRecordsResponse._();

  factory ListLinearProcessingRecordsResponse.fromBuffer(
          $core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListLinearProcessingRecordsResponse()..mergeFromBuffer(data, registry);
  factory ListLinearProcessingRecordsResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListLinearProcessingRecordsResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListLinearProcessingRecordsResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: ListLinearProcessingRecordsResponse.$_createMessage)
    ..pPM<LinearProcessingRecord>(1, _omitFieldNames ? '' : 'records',
        subBuilder: LinearProcessingRecord.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListLinearProcessingRecordsResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListLinearProcessingRecordsResponse copyWith(
          void Function(ListLinearProcessingRecordsResponse) updates) =>
      super.copyWith((message) =>
              updates(message as ListLinearProcessingRecordsResponse))
          as ListLinearProcessingRecordsResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use ListLinearProcessingRecordsResponse() / ListLinearProcessingRecordsResponse.new instead')
  static ListLinearProcessingRecordsResponse create() =>
      ListLinearProcessingRecordsResponse._();
  static $pb.GeneratedMessage $_createMessage() =>
      ListLinearProcessingRecordsResponse._();
  @$core.override
  ListLinearProcessingRecordsResponse createEmptyInstance() =>
      ListLinearProcessingRecordsResponse._();
  @$core.pragma('dart2js:noInline')
  static ListLinearProcessingRecordsResponse getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<
              ListLinearProcessingRecordsResponse>(
          ListLinearProcessingRecordsResponse.$_createMessage);
  static ListLinearProcessingRecordsResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $pb.PbList<LinearProcessingRecord> get records => $_getList(0);
}

class GetLinearProcessingRecordRequest extends $pb.GeneratedMessage {
  factory GetLinearProcessingRecordRequest({
    $core.String? id,
  }) {
    final result = GetLinearProcessingRecordRequest._();
    if (id != null) result.id = id;
    return result;
  }

  GetLinearProcessingRecordRequest._();

  factory GetLinearProcessingRecordRequest.fromBuffer(
          $core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetLinearProcessingRecordRequest()..mergeFromBuffer(data, registry);
  factory GetLinearProcessingRecordRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetLinearProcessingRecordRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetLinearProcessingRecordRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: GetLinearProcessingRecordRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetLinearProcessingRecordRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetLinearProcessingRecordRequest copyWith(
          void Function(GetLinearProcessingRecordRequest) updates) =>
      super.copyWith(
              (message) => updates(message as GetLinearProcessingRecordRequest))
          as GetLinearProcessingRecordRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use GetLinearProcessingRecordRequest() / GetLinearProcessingRecordRequest.new instead')
  static GetLinearProcessingRecordRequest create() =>
      GetLinearProcessingRecordRequest._();
  static $pb.GeneratedMessage $_createMessage() =>
      GetLinearProcessingRecordRequest._();
  @$core.override
  GetLinearProcessingRecordRequest createEmptyInstance() =>
      GetLinearProcessingRecordRequest._();
  @$core.pragma('dart2js:noInline')
  static GetLinearProcessingRecordRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<GetLinearProcessingRecordRequest>(
          GetLinearProcessingRecordRequest.$_createMessage);
  static GetLinearProcessingRecordRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);
}

class GetLinearProcessingRecordResponse extends $pb.GeneratedMessage {
  factory GetLinearProcessingRecordResponse({
    LinearProcessingRecord? record,
  }) {
    final result = GetLinearProcessingRecordResponse._();
    if (record != null) result.record = record;
    return result;
  }

  GetLinearProcessingRecordResponse._();

  factory GetLinearProcessingRecordResponse.fromBuffer(
          $core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetLinearProcessingRecordResponse()..mergeFromBuffer(data, registry);
  factory GetLinearProcessingRecordResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetLinearProcessingRecordResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetLinearProcessingRecordResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: GetLinearProcessingRecordResponse.$_createMessage)
    ..aOM<LinearProcessingRecord>(1, _omitFieldNames ? '' : 'record',
        subBuilder: LinearProcessingRecord.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetLinearProcessingRecordResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetLinearProcessingRecordResponse copyWith(
          void Function(GetLinearProcessingRecordResponse) updates) =>
      super.copyWith((message) =>
              updates(message as GetLinearProcessingRecordResponse))
          as GetLinearProcessingRecordResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use GetLinearProcessingRecordResponse() / GetLinearProcessingRecordResponse.new instead')
  static GetLinearProcessingRecordResponse create() =>
      GetLinearProcessingRecordResponse._();
  static $pb.GeneratedMessage $_createMessage() =>
      GetLinearProcessingRecordResponse._();
  @$core.override
  GetLinearProcessingRecordResponse createEmptyInstance() =>
      GetLinearProcessingRecordResponse._();
  @$core.pragma('dart2js:noInline')
  static GetLinearProcessingRecordResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<GetLinearProcessingRecordResponse>(
          GetLinearProcessingRecordResponse.$_createMessage);
  static GetLinearProcessingRecordResponse? _defaultInstance;

  @$pb.TagNumber(1)
  LinearProcessingRecord get record => $_getN(0);
  @$pb.TagNumber(1)
  set record(LinearProcessingRecord value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasRecord() => $_has(0);
  @$pb.TagNumber(1)
  void clearRecord() => $_clearField(1);
  @$pb.TagNumber(1)
  LinearProcessingRecord ensureRecord() => $_ensure(0);
}

class ResendLinearReplyRequest extends $pb.GeneratedMessage {
  factory ResendLinearReplyRequest({
    $core.String? id,
  }) {
    final result = ResendLinearReplyRequest._();
    if (id != null) result.id = id;
    return result;
  }

  ResendLinearReplyRequest._();

  factory ResendLinearReplyRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ResendLinearReplyRequest()..mergeFromBuffer(data, registry);
  factory ResendLinearReplyRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ResendLinearReplyRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ResendLinearReplyRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: ResendLinearReplyRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ResendLinearReplyRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ResendLinearReplyRequest copyWith(
          void Function(ResendLinearReplyRequest) updates) =>
      super.copyWith((message) => updates(message as ResendLinearReplyRequest))
          as ResendLinearReplyRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use ResendLinearReplyRequest() / ResendLinearReplyRequest.new instead')
  static ResendLinearReplyRequest create() => ResendLinearReplyRequest._();
  static $pb.GeneratedMessage $_createMessage() => ResendLinearReplyRequest._();
  @$core.override
  ResendLinearReplyRequest createEmptyInstance() =>
      ResendLinearReplyRequest._();
  @$core.pragma('dart2js:noInline')
  static ResendLinearReplyRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ResendLinearReplyRequest>(
          ResendLinearReplyRequest.$_createMessage);
  static ResendLinearReplyRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);
}

class ResendLinearReplyResponse extends $pb.GeneratedMessage {
  factory ResendLinearReplyResponse({
    LinearProcessingRecord? record,
  }) {
    final result = ResendLinearReplyResponse._();
    if (record != null) result.record = record;
    return result;
  }

  ResendLinearReplyResponse._();

  factory ResendLinearReplyResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ResendLinearReplyResponse()..mergeFromBuffer(data, registry);
  factory ResendLinearReplyResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ResendLinearReplyResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ResendLinearReplyResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: ResendLinearReplyResponse.$_createMessage)
    ..aOM<LinearProcessingRecord>(1, _omitFieldNames ? '' : 'record',
        subBuilder: LinearProcessingRecord.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ResendLinearReplyResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ResendLinearReplyResponse copyWith(
          void Function(ResendLinearReplyResponse) updates) =>
      super.copyWith((message) => updates(message as ResendLinearReplyResponse))
          as ResendLinearReplyResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use ResendLinearReplyResponse() / ResendLinearReplyResponse.new instead')
  static ResendLinearReplyResponse create() => ResendLinearReplyResponse._();
  static $pb.GeneratedMessage $_createMessage() =>
      ResendLinearReplyResponse._();
  @$core.override
  ResendLinearReplyResponse createEmptyInstance() =>
      ResendLinearReplyResponse._();
  @$core.pragma('dart2js:noInline')
  static ResendLinearReplyResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ResendLinearReplyResponse>(
          ResendLinearReplyResponse.$_createMessage);
  static ResendLinearReplyResponse? _defaultInstance;

  @$pb.TagNumber(1)
  LinearProcessingRecord get record => $_getN(0);
  @$pb.TagNumber(1)
  set record(LinearProcessingRecord value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasRecord() => $_has(0);
  @$pb.TagNumber(1)
  void clearRecord() => $_clearField(1);
  @$pb.TagNumber(1)
  LinearProcessingRecord ensureRecord() => $_ensure(0);
}

/// LinearAppService manages a workspace's Linear Apps. Members may read;
/// workspace owners and admins manage.
class LinearAppServiceApi {
  final $pb.RpcClient _client;

  LinearAppServiceApi(this._client);

  $async.Future<ListLinearAppsResponse> listLinearApps(
          $pb.ClientContext? ctx, ListLinearAppsRequest request) =>
      _client.invoke<ListLinearAppsResponse>(ctx, 'LinearAppService',
          'ListLinearApps', request, ListLinearAppsResponse());
  $async.Future<GetLinearAppResponse> getLinearApp(
          $pb.ClientContext? ctx, GetLinearAppRequest request) =>
      _client.invoke<GetLinearAppResponse>(ctx, 'LinearAppService',
          'GetLinearApp', request, GetLinearAppResponse());

  /// CreateLinearApp registers a Linear OAuth app routed to one Agent.
  $async.Future<CreateLinearAppResponse> createLinearApp(
          $pb.ClientContext? ctx, CreateLinearAppRequest request) =>
      _client.invoke<CreateLinearAppResponse>(ctx, 'LinearAppService',
          'CreateLinearApp', request, CreateLinearAppResponse());

  /// UpdateLinearApp replaces the mutable fields under the optimistic
  /// revision check. The ID and client ID are immutable.
  $async.Future<UpdateLinearAppResponse> updateLinearApp(
          $pb.ClientContext? ctx, UpdateLinearAppRequest request) =>
      _client.invoke<UpdateLinearAppResponse>(ctx, 'LinearAppService',
          'UpdateLinearApp', request, UpdateLinearAppResponse());

  /// PutLinearAppCredentials sets or clears the write-only secrets.
  $async.Future<PutLinearAppCredentialsResponse> putLinearAppCredentials(
          $pb.ClientContext? ctx, PutLinearAppCredentialsRequest request) =>
      _client.invoke<PutLinearAppCredentialsResponse>(
          ctx,
          'LinearAppService',
          'PutLinearAppCredentials',
          request,
          PutLinearAppCredentialsResponse());

  /// DeleteLinearApp removes the App together with its installations.
  $async.Future<DeleteLinearAppResponse> deleteLinearApp(
          $pb.ClientContext? ctx, DeleteLinearAppRequest request) =>
      _client.invoke<DeleteLinearAppResponse>(ctx, 'LinearAppService',
          'DeleteLinearApp', request, DeleteLinearAppResponse());

  /// BeginLinearInstall starts installing the App into a Linear organization:
  /// it returns Linear's authorize URL, carrying a single-use state that the
  /// public OAuth callback consumes.
  $async.Future<BeginLinearInstallResponse> beginLinearInstall(
          $pb.ClientContext? ctx, BeginLinearInstallRequest request) =>
      _client.invoke<BeginLinearInstallResponse>(ctx, 'LinearAppService',
          'BeginLinearInstall', request, BeginLinearInstallResponse());
  $async.Future<ListLinearInstallationsResponse> listLinearInstallations(
          $pb.ClientContext? ctx, ListLinearInstallationsRequest request) =>
      _client.invoke<ListLinearInstallationsResponse>(
          ctx,
          'LinearAppService',
          'ListLinearInstallations',
          request,
          ListLinearInstallationsResponse());

  /// DeleteLinearInstallation removes an installation, revoking its token at
  /// Linear as a best effort.
  $async.Future<DeleteLinearInstallationResponse> deleteLinearInstallation(
          $pb.ClientContext? ctx, DeleteLinearInstallationRequest request) =>
      _client.invoke<DeleteLinearInstallationResponse>(
          ctx,
          'LinearAppService',
          'DeleteLinearInstallation',
          request,
          DeleteLinearInstallationResponse());
}

/// =============================================================================
/// LinearAdminService manages platform-level Linear settings.
///
/// The public base URL is global, not per-workspace: it names the public
/// address of this deployment. Only global administrators may read or change
/// it, so workspace input can never redirect Linear webhooks or OAuth
/// callbacks.
/// =============================================================================
class LinearAdminServiceApi {
  final $pb.RpcClient _client;

  LinearAdminServiceApi(this._client);

  $async.Future<GetLinearSettingsResponse> getLinearSettings(
          $pb.ClientContext? ctx, GetLinearSettingsRequest request) =>
      _client.invoke<GetLinearSettingsResponse>(ctx, 'LinearAdminService',
          'GetLinearSettings', request, GetLinearSettingsResponse());
  $async.Future<UpdateLinearSettingsResponse> updateLinearSettings(
          $pb.ClientContext? ctx, UpdateLinearSettingsRequest request) =>
      _client.invoke<UpdateLinearSettingsResponse>(ctx, 'LinearAdminService',
          'UpdateLinearSettings', request, UpdateLinearSettingsResponse());
}

/// LinearProcessingService exposes processing records. Members read;
/// owners and admins resend a reply that was produced but not posted. There
/// is deliberately no rerun action.
class LinearProcessingServiceApi {
  final $pb.RpcClient _client;

  LinearProcessingServiceApi(this._client);

  $async.Future<ListLinearProcessingRecordsResponse>
      listLinearProcessingRecords($pb.ClientContext? ctx,
              ListLinearProcessingRecordsRequest request) =>
          _client.invoke<ListLinearProcessingRecordsResponse>(
              ctx,
              'LinearProcessingService',
              'ListLinearProcessingRecords',
              request,
              ListLinearProcessingRecordsResponse());
  $async.Future<GetLinearProcessingRecordResponse> getLinearProcessingRecord(
          $pb.ClientContext? ctx, GetLinearProcessingRecordRequest request) =>
      _client.invoke<GetLinearProcessingRecordResponse>(
          ctx,
          'LinearProcessingService',
          'GetLinearProcessingRecord',
          request,
          GetLinearProcessingRecordResponse());

  /// ResendLinearReply posts the persisted reply of a FAILED record. It never
  /// invokes the Agent, and it is refused when no reply was persisted.
  $async.Future<ResendLinearReplyResponse> resendLinearReply(
          $pb.ClientContext? ctx, ResendLinearReplyRequest request) =>
      _client.invoke<ResendLinearReplyResponse>(ctx, 'LinearProcessingService',
          'ResendLinearReply', request, ResendLinearReplyResponse());
}

const $core.bool _omitFieldNames =
    $core.bool.fromEnvironment('protobuf.omit_field_names');
const $core.bool _omitMessageNames =
    $core.bool.fromEnvironment('protobuf.omit_message_names');
