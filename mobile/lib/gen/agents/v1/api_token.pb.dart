// This is a generated file - do not edit.
//
// Generated from agents/v1/api_token.proto.

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

import 'api_token.pbenum.dart';

export 'package:protobuf/protobuf.dart' show GeneratedMessageGenericExtensions;

export 'api_token.pbenum.dart';

/// APIToken represents a stored API token. The plaintext secret is only
/// returned at create time; subsequent reads expose the prefix only.
class APIToken extends $pb.GeneratedMessage {
  factory APIToken({
    $core.String? id,
    $core.String? name,
    $core.String? prefix,
    $0.Timestamp? createdAt,
    $0.Timestamp? lastUsedAt,
    $core.bool? revoked,
    APITokenKind? kind,
    $core.Iterable<$core.String>? scopes,
    $0.Timestamp? expiresAt,
    $core.String? daemonRuntimeId,
    $core.String? workspaceId,
  }) {
    final result = APIToken._();
    if (id != null) result.id = id;
    if (name != null) result.name = name;
    if (prefix != null) result.prefix = prefix;
    if (createdAt != null) result.createdAt = createdAt;
    if (lastUsedAt != null) result.lastUsedAt = lastUsedAt;
    if (revoked != null) result.revoked = revoked;
    if (kind != null) result.kind = kind;
    if (scopes != null) result.scopes.addAll(scopes);
    if (expiresAt != null) result.expiresAt = expiresAt;
    if (daemonRuntimeId != null) result.daemonRuntimeId = daemonRuntimeId;
    if (workspaceId != null) result.workspaceId = workspaceId;
    return result;
  }

  APIToken._();

  factory APIToken.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      APIToken()..mergeFromBuffer(data, registry);
  factory APIToken.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      APIToken()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'APIToken',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: APIToken.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..aOS(2, _omitFieldNames ? '' : 'name')
    ..aOS(3, _omitFieldNames ? '' : 'prefix')
    ..aOM<$0.Timestamp>(4, _omitFieldNames ? '' : 'createdAt',
        subBuilder: $0.Timestamp.$_createMessage)
    ..aOM<$0.Timestamp>(5, _omitFieldNames ? '' : 'lastUsedAt',
        subBuilder: $0.Timestamp.$_createMessage)
    ..aOB(6, _omitFieldNames ? '' : 'revoked')
    ..aE<APITokenKind>(7, _omitFieldNames ? '' : 'kind',
        enumValues: APITokenKind.values)
    ..pPS(8, _omitFieldNames ? '' : 'scopes')
    ..aOM<$0.Timestamp>(9, _omitFieldNames ? '' : 'expiresAt',
        subBuilder: $0.Timestamp.$_createMessage)
    ..aOS(10, _omitFieldNames ? '' : 'daemonRuntimeId')
    ..aOS(100, _omitFieldNames ? '' : 'workspaceId')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  APIToken clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  APIToken copyWith(void Function(APIToken) updates) =>
      super.copyWith((message) => updates(message as APIToken)) as APIToken;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use APIToken() / APIToken.new instead')
  static APIToken create() => APIToken._();
  static $pb.GeneratedMessage $_createMessage() => APIToken._();
  @$core.override
  APIToken createEmptyInstance() => APIToken._();
  @$core.pragma('dart2js:noInline')
  static APIToken getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<APIToken>(APIToken.$_createMessage);
  static APIToken? _defaultInstance;

  /// Server-assigned unique identifier.
  /// @gotags: json:"id,omitempty" yaml:"id,omitempty" bson:"_id,omitempty"
  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);

  /// Human-readable label.
  /// @gotags: json:"name,omitempty" yaml:"name,omitempty" bson:"name,omitempty"
  @$pb.TagNumber(2)
  $core.String get name => $_getSZ(1);
  @$pb.TagNumber(2)
  set name($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasName() => $_has(1);
  @$pb.TagNumber(2)
  void clearName() => $_clearField(2);

  /// Display prefix (first chars of the token). Used by the dashboard to help
  /// operators identify tokens without storing the secret.
  /// @gotags: json:"prefix,omitempty" yaml:"prefix,omitempty" bson:"prefix,omitempty"
  @$pb.TagNumber(3)
  $core.String get prefix => $_getSZ(2);
  @$pb.TagNumber(3)
  set prefix($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasPrefix() => $_has(2);
  @$pb.TagNumber(3)
  void clearPrefix() => $_clearField(3);

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

  /// Time the token was last accepted by the auth middleware. May be zero.
  /// @gotags: json:"last_used_at,omitempty" yaml:"last_used_at,omitempty" bson:"last_used_at,omitempty"
  @$pb.TagNumber(5)
  $0.Timestamp get lastUsedAt => $_getN(4);
  @$pb.TagNumber(5)
  set lastUsedAt($0.Timestamp value) => $_setField(5, value);
  @$pb.TagNumber(5)
  $core.bool hasLastUsedAt() => $_has(4);
  @$pb.TagNumber(5)
  void clearLastUsedAt() => $_clearField(5);
  @$pb.TagNumber(5)
  $0.Timestamp ensureLastUsedAt() => $_ensure(4);

  /// True once the token is revoked. Revoked tokens cannot authenticate.
  /// @gotags: json:"revoked,omitempty" yaml:"revoked,omitempty" bson:"revoked,omitempty"
  @$pb.TagNumber(6)
  $core.bool get revoked => $_getBF(5);
  @$pb.TagNumber(6)
  set revoked($core.bool value) => $_setBool(5, value);
  @$pb.TagNumber(6)
  $core.bool hasRevoked() => $_has(5);
  @$pb.TagNumber(6)
  void clearRevoked() => $_clearField(6);

  /// Token kind controls which entrypoints may accept this secret.
  /// @gotags: json:"kind,omitempty" yaml:"kind,omitempty" bson:"kind,omitempty"
  @$pb.TagNumber(7)
  APITokenKind get kind => $_getN(6);
  @$pb.TagNumber(7)
  set kind(APITokenKind value) => $_setField(7, value);
  @$pb.TagNumber(7)
  $core.bool hasKind() => $_has(6);
  @$pb.TagNumber(7)
  void clearKind() => $_clearField(7);

  /// Fine-grained permission labels such as "api:*" or "daemon:connect".
  /// @gotags: json:"scopes,omitempty" yaml:"scopes,omitempty" bson:"scopes,omitempty"
  @$pb.TagNumber(8)
  $pb.PbList<$core.String> get scopes => $_getList(7);

  /// Optional expiry time. Empty means the token does not expire.
  /// @gotags: json:"expires_at,omitempty" yaml:"expires_at,omitempty" bson:"expires_at,omitempty"
  @$pb.TagNumber(9)
  $0.Timestamp get expiresAt => $_getN(8);
  @$pb.TagNumber(9)
  set expiresAt($0.Timestamp value) => $_setField(9, value);
  @$pb.TagNumber(9)
  $core.bool hasExpiresAt() => $_has(8);
  @$pb.TagNumber(9)
  void clearExpiresAt() => $_clearField(9);
  @$pb.TagNumber(9)
  $0.Timestamp ensureExpiresAt() => $_ensure(8);

  /// For daemon tokens, the daemon runtime id this token may connect as.
  /// @gotags: json:"daemon_runtime_id,omitempty" yaml:"daemon_runtime_id,omitempty" bson:"daemon_runtime_id,omitempty"
  @$pb.TagNumber(10)
  $core.String get daemonRuntimeId => $_getSZ(9);
  @$pb.TagNumber(10)
  set daemonRuntimeId($core.String value) => $_setString(9, value);
  @$pb.TagNumber(10)
  $core.bool hasDaemonRuntimeId() => $_has(9);
  @$pb.TagNumber(10)
  void clearDaemonRuntimeId() => $_clearField(10);

  /// ID of the workspace this token belongs to. Tokens authenticate within
  /// a single workspace.
  /// @gotags: json:"workspace_id,omitempty" yaml:"workspace_id,omitempty" bson:"workspace_id,omitempty"
  @$pb.TagNumber(100)
  $core.String get workspaceId => $_getSZ(10);
  @$pb.TagNumber(100)
  set workspaceId($core.String value) => $_setString(10, value);
  @$pb.TagNumber(100)
  $core.bool hasWorkspaceId() => $_has(10);
  @$pb.TagNumber(100)
  void clearWorkspaceId() => $_clearField(100);
}

class ListAPITokensRequest extends $pb.GeneratedMessage {
  factory ListAPITokensRequest() => ListAPITokensRequest._();

  ListAPITokensRequest._();

  factory ListAPITokensRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListAPITokensRequest()..mergeFromBuffer(data, registry);
  factory ListAPITokensRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListAPITokensRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListAPITokensRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: ListAPITokensRequest.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListAPITokensRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListAPITokensRequest copyWith(void Function(ListAPITokensRequest) updates) =>
      super.copyWith((message) => updates(message as ListAPITokensRequest))
          as ListAPITokensRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use ListAPITokensRequest() / ListAPITokensRequest.new instead')
  static ListAPITokensRequest create() => ListAPITokensRequest._();
  static $pb.GeneratedMessage $_createMessage() => ListAPITokensRequest._();
  @$core.override
  ListAPITokensRequest createEmptyInstance() => ListAPITokensRequest._();
  @$core.pragma('dart2js:noInline')
  static ListAPITokensRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ListAPITokensRequest>(
          ListAPITokensRequest.$_createMessage);
  static ListAPITokensRequest? _defaultInstance;
}

class ListAPITokensResponse extends $pb.GeneratedMessage {
  factory ListAPITokensResponse({
    $core.Iterable<APIToken>? tokens,
  }) {
    final result = ListAPITokensResponse._();
    if (tokens != null) result.tokens.addAll(tokens);
    return result;
  }

  ListAPITokensResponse._();

  factory ListAPITokensResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListAPITokensResponse()..mergeFromBuffer(data, registry);
  factory ListAPITokensResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListAPITokensResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListAPITokensResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: ListAPITokensResponse.$_createMessage)
    ..pPM<APIToken>(1, _omitFieldNames ? '' : 'tokens',
        subBuilder: APIToken.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListAPITokensResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListAPITokensResponse copyWith(
          void Function(ListAPITokensResponse) updates) =>
      super.copyWith((message) => updates(message as ListAPITokensResponse))
          as ListAPITokensResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use ListAPITokensResponse() / ListAPITokensResponse.new instead')
  static ListAPITokensResponse create() => ListAPITokensResponse._();
  static $pb.GeneratedMessage $_createMessage() => ListAPITokensResponse._();
  @$core.override
  ListAPITokensResponse createEmptyInstance() => ListAPITokensResponse._();
  @$core.pragma('dart2js:noInline')
  static ListAPITokensResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ListAPITokensResponse>(
          ListAPITokensResponse.$_createMessage);
  static ListAPITokensResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $pb.PbList<APIToken> get tokens => $_getList(0);
}

class CreateAPITokenRequest extends $pb.GeneratedMessage {
  factory CreateAPITokenRequest({
    $core.String? name,
  }) {
    final result = CreateAPITokenRequest._();
    if (name != null) result.name = name;
    return result;
  }

  CreateAPITokenRequest._();

  factory CreateAPITokenRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateAPITokenRequest()..mergeFromBuffer(data, registry);
  factory CreateAPITokenRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateAPITokenRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CreateAPITokenRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: CreateAPITokenRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'name')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateAPITokenRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateAPITokenRequest copyWith(
          void Function(CreateAPITokenRequest) updates) =>
      super.copyWith((message) => updates(message as CreateAPITokenRequest))
          as CreateAPITokenRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use CreateAPITokenRequest() / CreateAPITokenRequest.new instead')
  static CreateAPITokenRequest create() => CreateAPITokenRequest._();
  static $pb.GeneratedMessage $_createMessage() => CreateAPITokenRequest._();
  @$core.override
  CreateAPITokenRequest createEmptyInstance() => CreateAPITokenRequest._();
  @$core.pragma('dart2js:noInline')
  static CreateAPITokenRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<CreateAPITokenRequest>(
          CreateAPITokenRequest.$_createMessage);
  static CreateAPITokenRequest? _defaultInstance;

  /// Required label for the new token.
  @$pb.TagNumber(1)
  $core.String get name => $_getSZ(0);
  @$pb.TagNumber(1)
  set name($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasName() => $_has(0);
  @$pb.TagNumber(1)
  void clearName() => $_clearField(1);
}

class CreateAPITokenResponse extends $pb.GeneratedMessage {
  factory CreateAPITokenResponse({
    APIToken? token,
    $core.String? secret,
  }) {
    final result = CreateAPITokenResponse._();
    if (token != null) result.token = token;
    if (secret != null) result.secret = secret;
    return result;
  }

  CreateAPITokenResponse._();

  factory CreateAPITokenResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateAPITokenResponse()..mergeFromBuffer(data, registry);
  factory CreateAPITokenResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateAPITokenResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CreateAPITokenResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: CreateAPITokenResponse.$_createMessage)
    ..aOM<APIToken>(1, _omitFieldNames ? '' : 'token',
        subBuilder: APIToken.$_createMessage)
    ..aOS(2, _omitFieldNames ? '' : 'secret')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateAPITokenResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateAPITokenResponse copyWith(
          void Function(CreateAPITokenResponse) updates) =>
      super.copyWith((message) => updates(message as CreateAPITokenResponse))
          as CreateAPITokenResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use CreateAPITokenResponse() / CreateAPITokenResponse.new instead')
  static CreateAPITokenResponse create() => CreateAPITokenResponse._();
  static $pb.GeneratedMessage $_createMessage() => CreateAPITokenResponse._();
  @$core.override
  CreateAPITokenResponse createEmptyInstance() => CreateAPITokenResponse._();
  @$core.pragma('dart2js:noInline')
  static CreateAPITokenResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<CreateAPITokenResponse>(
          CreateAPITokenResponse.$_createMessage);
  static CreateAPITokenResponse? _defaultInstance;

  /// The persisted token metadata (no secret).
  @$pb.TagNumber(1)
  APIToken get token => $_getN(0);
  @$pb.TagNumber(1)
  set token(APIToken value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasToken() => $_has(0);
  @$pb.TagNumber(1)
  void clearToken() => $_clearField(1);
  @$pb.TagNumber(1)
  APIToken ensureToken() => $_ensure(0);

  /// Plaintext secret returned only at create time. Treat as sensitive.
  @$pb.TagNumber(2)
  $core.String get secret => $_getSZ(1);
  @$pb.TagNumber(2)
  set secret($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasSecret() => $_has(1);
  @$pb.TagNumber(2)
  void clearSecret() => $_clearField(2);
}

class RevokeAPITokenRequest extends $pb.GeneratedMessage {
  factory RevokeAPITokenRequest({
    $core.String? id,
  }) {
    final result = RevokeAPITokenRequest._();
    if (id != null) result.id = id;
    return result;
  }

  RevokeAPITokenRequest._();

  factory RevokeAPITokenRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      RevokeAPITokenRequest()..mergeFromBuffer(data, registry);
  factory RevokeAPITokenRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      RevokeAPITokenRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'RevokeAPITokenRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: RevokeAPITokenRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  RevokeAPITokenRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  RevokeAPITokenRequest copyWith(
          void Function(RevokeAPITokenRequest) updates) =>
      super.copyWith((message) => updates(message as RevokeAPITokenRequest))
          as RevokeAPITokenRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use RevokeAPITokenRequest() / RevokeAPITokenRequest.new instead')
  static RevokeAPITokenRequest create() => RevokeAPITokenRequest._();
  static $pb.GeneratedMessage $_createMessage() => RevokeAPITokenRequest._();
  @$core.override
  RevokeAPITokenRequest createEmptyInstance() => RevokeAPITokenRequest._();
  @$core.pragma('dart2js:noInline')
  static RevokeAPITokenRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<RevokeAPITokenRequest>(
          RevokeAPITokenRequest.$_createMessage);
  static RevokeAPITokenRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);
}

class RevokeAPITokenResponse extends $pb.GeneratedMessage {
  factory RevokeAPITokenResponse({
    APIToken? token,
  }) {
    final result = RevokeAPITokenResponse._();
    if (token != null) result.token = token;
    return result;
  }

  RevokeAPITokenResponse._();

  factory RevokeAPITokenResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      RevokeAPITokenResponse()..mergeFromBuffer(data, registry);
  factory RevokeAPITokenResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      RevokeAPITokenResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'RevokeAPITokenResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: RevokeAPITokenResponse.$_createMessage)
    ..aOM<APIToken>(1, _omitFieldNames ? '' : 'token',
        subBuilder: APIToken.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  RevokeAPITokenResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  RevokeAPITokenResponse copyWith(
          void Function(RevokeAPITokenResponse) updates) =>
      super.copyWith((message) => updates(message as RevokeAPITokenResponse))
          as RevokeAPITokenResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use RevokeAPITokenResponse() / RevokeAPITokenResponse.new instead')
  static RevokeAPITokenResponse create() => RevokeAPITokenResponse._();
  static $pb.GeneratedMessage $_createMessage() => RevokeAPITokenResponse._();
  @$core.override
  RevokeAPITokenResponse createEmptyInstance() => RevokeAPITokenResponse._();
  @$core.pragma('dart2js:noInline')
  static RevokeAPITokenResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<RevokeAPITokenResponse>(
          RevokeAPITokenResponse.$_createMessage);
  static RevokeAPITokenResponse? _defaultInstance;

  @$pb.TagNumber(1)
  APIToken get token => $_getN(0);
  @$pb.TagNumber(1)
  set token(APIToken value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasToken() => $_has(0);
  @$pb.TagNumber(1)
  void clearToken() => $_clearField(1);
  @$pb.TagNumber(1)
  APIToken ensureToken() => $_ensure(0);
}

class APITokenServiceApi {
  final $pb.RpcClient _client;

  APITokenServiceApi(this._client);

  $async.Future<ListAPITokensResponse> listAPITokens(
          $pb.ClientContext? ctx, ListAPITokensRequest request) =>
      _client.invoke<ListAPITokensResponse>(ctx, 'APITokenService',
          'ListAPITokens', request, ListAPITokensResponse());
  $async.Future<CreateAPITokenResponse> createAPIToken(
          $pb.ClientContext? ctx, CreateAPITokenRequest request) =>
      _client.invoke<CreateAPITokenResponse>(ctx, 'APITokenService',
          'CreateAPIToken', request, CreateAPITokenResponse());
  $async.Future<RevokeAPITokenResponse> revokeAPIToken(
          $pb.ClientContext? ctx, RevokeAPITokenRequest request) =>
      _client.invoke<RevokeAPITokenResponse>(ctx, 'APITokenService',
          'RevokeAPIToken', request, RevokeAPITokenResponse());
}

const $core.bool _omitFieldNames =
    $core.bool.fromEnvironment('protobuf.omit_field_names');
const $core.bool _omitMessageNames =
    $core.bool.fromEnvironment('protobuf.omit_message_names');
