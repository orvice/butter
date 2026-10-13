// This is a generated file - do not edit.
//
// Generated from agents/v1/auth.proto.

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

import 'workspace.pb.dart' as $1;

export 'package:protobuf/protobuf.dart' show GeneratedMessageGenericExtensions;

class User extends $pb.GeneratedMessage {
  factory User({
    $core.String? id,
    $core.String? username,
    $core.String? displayName,
    $core.String? role,
    $core.bool? disabled,
    $0.Timestamp? createdAt,
    $0.Timestamp? updatedAt,
    $core.String? avatarUrl,
    $core.String? email,
    $core.String? provider,
    $core.String? externalId,
  }) {
    final result = User._();
    if (id != null) result.id = id;
    if (username != null) result.username = username;
    if (displayName != null) result.displayName = displayName;
    if (role != null) result.role = role;
    if (disabled != null) result.disabled = disabled;
    if (createdAt != null) result.createdAt = createdAt;
    if (updatedAt != null) result.updatedAt = updatedAt;
    if (avatarUrl != null) result.avatarUrl = avatarUrl;
    if (email != null) result.email = email;
    if (provider != null) result.provider = provider;
    if (externalId != null) result.externalId = externalId;
    return result;
  }

  User._();

  factory User.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      User()..mergeFromBuffer(data, registry);
  factory User.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      User()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'User',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: User.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..aOS(2, _omitFieldNames ? '' : 'username')
    ..aOS(3, _omitFieldNames ? '' : 'displayName')
    ..aOS(4, _omitFieldNames ? '' : 'role')
    ..aOB(5, _omitFieldNames ? '' : 'disabled')
    ..aOM<$0.Timestamp>(6, _omitFieldNames ? '' : 'createdAt',
        subBuilder: $0.Timestamp.$_createMessage)
    ..aOM<$0.Timestamp>(7, _omitFieldNames ? '' : 'updatedAt',
        subBuilder: $0.Timestamp.$_createMessage)
    ..aOS(8, _omitFieldNames ? '' : 'avatarUrl')
    ..aOS(9, _omitFieldNames ? '' : 'email')
    ..aOS(10, _omitFieldNames ? '' : 'provider')
    ..aOS(11, _omitFieldNames ? '' : 'externalId')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  User clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  User copyWith(void Function(User) updates) =>
      super.copyWith((message) => updates(message as User)) as User;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use User() / User.new instead')
  static User create() => User._();
  static $pb.GeneratedMessage $_createMessage() => User._();
  @$core.override
  User createEmptyInstance() => User._();
  @$core.pragma('dart2js:noInline')
  static User getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<User>(User.$_createMessage);
  static User? _defaultInstance;

  /// @gotags: json:"id,omitempty" yaml:"id,omitempty" bson:"_id,omitempty"
  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);

  /// @gotags: json:"username,omitempty" yaml:"username,omitempty" bson:"username,omitempty"
  @$pb.TagNumber(2)
  $core.String get username => $_getSZ(1);
  @$pb.TagNumber(2)
  set username($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasUsername() => $_has(1);
  @$pb.TagNumber(2)
  void clearUsername() => $_clearField(2);

  /// @gotags: json:"display_name,omitempty" yaml:"display_name,omitempty" bson:"display_name,omitempty"
  @$pb.TagNumber(3)
  $core.String get displayName => $_getSZ(2);
  @$pb.TagNumber(3)
  set displayName($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasDisplayName() => $_has(2);
  @$pb.TagNumber(3)
  void clearDisplayName() => $_clearField(3);

  /// @gotags: json:"role,omitempty" yaml:"role,omitempty" bson:"role,omitempty"
  @$pb.TagNumber(4)
  $core.String get role => $_getSZ(3);
  @$pb.TagNumber(4)
  set role($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasRole() => $_has(3);
  @$pb.TagNumber(4)
  void clearRole() => $_clearField(4);

  /// @gotags: json:"disabled,omitempty" yaml:"disabled,omitempty" bson:"disabled,omitempty"
  @$pb.TagNumber(5)
  $core.bool get disabled => $_getBF(4);
  @$pb.TagNumber(5)
  set disabled($core.bool value) => $_setBool(4, value);
  @$pb.TagNumber(5)
  $core.bool hasDisabled() => $_has(4);
  @$pb.TagNumber(5)
  void clearDisabled() => $_clearField(5);

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

  /// @gotags: json:"avatar_url,omitempty" yaml:"avatar_url,omitempty" bson:"avatar_url,omitempty"
  @$pb.TagNumber(8)
  $core.String get avatarUrl => $_getSZ(7);
  @$pb.TagNumber(8)
  set avatarUrl($core.String value) => $_setString(7, value);
  @$pb.TagNumber(8)
  $core.bool hasAvatarUrl() => $_has(7);
  @$pb.TagNumber(8)
  void clearAvatarUrl() => $_clearField(8);

  /// @gotags: json:"email,omitempty" yaml:"email,omitempty" bson:"email,omitempty"
  @$pb.TagNumber(9)
  $core.String get email => $_getSZ(8);
  @$pb.TagNumber(9)
  set email($core.String value) => $_setString(8, value);
  @$pb.TagNumber(9)
  $core.bool hasEmail() => $_has(8);
  @$pb.TagNumber(9)
  void clearEmail() => $_clearField(9);

  /// OAuth provider name (e.g. "github"). Empty for password users.
  /// @gotags: json:"provider,omitempty" yaml:"provider,omitempty" bson:"provider,omitempty"
  @$pb.TagNumber(10)
  $core.String get provider => $_getSZ(9);
  @$pb.TagNumber(10)
  set provider($core.String value) => $_setString(9, value);
  @$pb.TagNumber(10)
  $core.bool hasProvider() => $_has(9);
  @$pb.TagNumber(10)
  void clearProvider() => $_clearField(10);

  /// External (provider-issued) account ID. Stable across email changes.
  /// @gotags: json:"external_id,omitempty" yaml:"external_id,omitempty" bson:"external_id,omitempty"
  @$pb.TagNumber(11)
  $core.String get externalId => $_getSZ(10);
  @$pb.TagNumber(11)
  set externalId($core.String value) => $_setString(10, value);
  @$pb.TagNumber(11)
  $core.bool hasExternalId() => $_has(10);
  @$pb.TagNumber(11)
  void clearExternalId() => $_clearField(11);
}

class LoginRequest extends $pb.GeneratedMessage {
  factory LoginRequest({
    $core.String? username,
    $core.String? password,
  }) {
    final result = LoginRequest._();
    if (username != null) result.username = username;
    if (password != null) result.password = password;
    return result;
  }

  LoginRequest._();

  factory LoginRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      LoginRequest()..mergeFromBuffer(data, registry);
  factory LoginRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      LoginRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'LoginRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: LoginRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'username')
    ..aOS(2, _omitFieldNames ? '' : 'password')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  LoginRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  LoginRequest copyWith(void Function(LoginRequest) updates) =>
      super.copyWith((message) => updates(message as LoginRequest))
          as LoginRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use LoginRequest() / LoginRequest.new instead')
  static LoginRequest create() => LoginRequest._();
  static $pb.GeneratedMessage $_createMessage() => LoginRequest._();
  @$core.override
  LoginRequest createEmptyInstance() => LoginRequest._();
  @$core.pragma('dart2js:noInline')
  static LoginRequest getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<LoginRequest>(
          LoginRequest.$_createMessage);
  static LoginRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get username => $_getSZ(0);
  @$pb.TagNumber(1)
  set username($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasUsername() => $_has(0);
  @$pb.TagNumber(1)
  void clearUsername() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get password => $_getSZ(1);
  @$pb.TagNumber(2)
  set password($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasPassword() => $_has(1);
  @$pb.TagNumber(2)
  void clearPassword() => $_clearField(2);
}

class LoginResponse extends $pb.GeneratedMessage {
  factory LoginResponse({
    $core.String? token,
    User? user,
    $0.Timestamp? expiresAt,
    $core.Iterable<$1.Workspace>? workspaces,
  }) {
    final result = LoginResponse._();
    if (token != null) result.token = token;
    if (user != null) result.user = user;
    if (expiresAt != null) result.expiresAt = expiresAt;
    if (workspaces != null) result.workspaces.addAll(workspaces);
    return result;
  }

  LoginResponse._();

  factory LoginResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      LoginResponse()..mergeFromBuffer(data, registry);
  factory LoginResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      LoginResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'LoginResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: LoginResponse.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'token')
    ..aOM<User>(2, _omitFieldNames ? '' : 'user',
        subBuilder: User.$_createMessage)
    ..aOM<$0.Timestamp>(3, _omitFieldNames ? '' : 'expiresAt',
        subBuilder: $0.Timestamp.$_createMessage)
    ..pPM<$1.Workspace>(4, _omitFieldNames ? '' : 'workspaces',
        subBuilder: $1.Workspace.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  LoginResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  LoginResponse copyWith(void Function(LoginResponse) updates) =>
      super.copyWith((message) => updates(message as LoginResponse))
          as LoginResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use LoginResponse() / LoginResponse.new instead')
  static LoginResponse create() => LoginResponse._();
  static $pb.GeneratedMessage $_createMessage() => LoginResponse._();
  @$core.override
  LoginResponse createEmptyInstance() => LoginResponse._();
  @$core.pragma('dart2js:noInline')
  static LoginResponse getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<LoginResponse>(
          LoginResponse.$_createMessage);
  static LoginResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get token => $_getSZ(0);
  @$pb.TagNumber(1)
  set token($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasToken() => $_has(0);
  @$pb.TagNumber(1)
  void clearToken() => $_clearField(1);

  @$pb.TagNumber(2)
  User get user => $_getN(1);
  @$pb.TagNumber(2)
  set user(User value) => $_setField(2, value);
  @$pb.TagNumber(2)
  $core.bool hasUser() => $_has(1);
  @$pb.TagNumber(2)
  void clearUser() => $_clearField(2);
  @$pb.TagNumber(2)
  User ensureUser() => $_ensure(1);

  @$pb.TagNumber(3)
  $0.Timestamp get expiresAt => $_getN(2);
  @$pb.TagNumber(3)
  set expiresAt($0.Timestamp value) => $_setField(3, value);
  @$pb.TagNumber(3)
  $core.bool hasExpiresAt() => $_has(2);
  @$pb.TagNumber(3)
  void clearExpiresAt() => $_clearField(3);
  @$pb.TagNumber(3)
  $0.Timestamp ensureExpiresAt() => $_ensure(2);

  /// Workspaces this user can access. Empty for global admins with no
  /// explicit memberships — they may still pick any workspace via
  /// WorkspaceService.ListWorkspaces.
  @$pb.TagNumber(4)
  $pb.PbList<$1.Workspace> get workspaces => $_getList(3);
}

class MeRequest extends $pb.GeneratedMessage {
  factory MeRequest() => MeRequest._();

  MeRequest._();

  factory MeRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      MeRequest()..mergeFromBuffer(data, registry);
  factory MeRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      MeRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'MeRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: MeRequest.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  MeRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  MeRequest copyWith(void Function(MeRequest) updates) =>
      super.copyWith((message) => updates(message as MeRequest)) as MeRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use MeRequest() / MeRequest.new instead')
  static MeRequest create() => MeRequest._();
  static $pb.GeneratedMessage $_createMessage() => MeRequest._();
  @$core.override
  MeRequest createEmptyInstance() => MeRequest._();
  @$core.pragma('dart2js:noInline')
  static MeRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<MeRequest>(MeRequest.$_createMessage);
  static MeRequest? _defaultInstance;
}

class MeResponse extends $pb.GeneratedMessage {
  factory MeResponse({
    User? user,
  }) {
    final result = MeResponse._();
    if (user != null) result.user = user;
    return result;
  }

  MeResponse._();

  factory MeResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      MeResponse()..mergeFromBuffer(data, registry);
  factory MeResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      MeResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'MeResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: MeResponse.$_createMessage)
    ..aOM<User>(1, _omitFieldNames ? '' : 'user',
        subBuilder: User.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  MeResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  MeResponse copyWith(void Function(MeResponse) updates) =>
      super.copyWith((message) => updates(message as MeResponse)) as MeResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use MeResponse() / MeResponse.new instead')
  static MeResponse create() => MeResponse._();
  static $pb.GeneratedMessage $_createMessage() => MeResponse._();
  @$core.override
  MeResponse createEmptyInstance() => MeResponse._();
  @$core.pragma('dart2js:noInline')
  static MeResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<MeResponse>(MeResponse.$_createMessage);
  static MeResponse? _defaultInstance;

  @$pb.TagNumber(1)
  User get user => $_getN(0);
  @$pb.TagNumber(1)
  set user(User value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasUser() => $_has(0);
  @$pb.TagNumber(1)
  void clearUser() => $_clearField(1);
  @$pb.TagNumber(1)
  User ensureUser() => $_ensure(0);
}

class LogoutRequest extends $pb.GeneratedMessage {
  factory LogoutRequest() => LogoutRequest._();

  LogoutRequest._();

  factory LogoutRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      LogoutRequest()..mergeFromBuffer(data, registry);
  factory LogoutRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      LogoutRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'LogoutRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: LogoutRequest.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  LogoutRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  LogoutRequest copyWith(void Function(LogoutRequest) updates) =>
      super.copyWith((message) => updates(message as LogoutRequest))
          as LogoutRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use LogoutRequest() / LogoutRequest.new instead')
  static LogoutRequest create() => LogoutRequest._();
  static $pb.GeneratedMessage $_createMessage() => LogoutRequest._();
  @$core.override
  LogoutRequest createEmptyInstance() => LogoutRequest._();
  @$core.pragma('dart2js:noInline')
  static LogoutRequest getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<LogoutRequest>(
          LogoutRequest.$_createMessage);
  static LogoutRequest? _defaultInstance;
}

class LogoutResponse extends $pb.GeneratedMessage {
  factory LogoutResponse() => LogoutResponse._();

  LogoutResponse._();

  factory LogoutResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      LogoutResponse()..mergeFromBuffer(data, registry);
  factory LogoutResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      LogoutResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'LogoutResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: LogoutResponse.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  LogoutResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  LogoutResponse copyWith(void Function(LogoutResponse) updates) =>
      super.copyWith((message) => updates(message as LogoutResponse))
          as LogoutResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use LogoutResponse() / LogoutResponse.new instead')
  static LogoutResponse create() => LogoutResponse._();
  static $pb.GeneratedMessage $_createMessage() => LogoutResponse._();
  @$core.override
  LogoutResponse createEmptyInstance() => LogoutResponse._();
  @$core.pragma('dart2js:noInline')
  static LogoutResponse getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<LogoutResponse>(
          LogoutResponse.$_createMessage);
  static LogoutResponse? _defaultInstance;
}

class ListUsersRequest extends $pb.GeneratedMessage {
  factory ListUsersRequest() => ListUsersRequest._();

  ListUsersRequest._();

  factory ListUsersRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListUsersRequest()..mergeFromBuffer(data, registry);
  factory ListUsersRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListUsersRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListUsersRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: ListUsersRequest.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListUsersRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListUsersRequest copyWith(void Function(ListUsersRequest) updates) =>
      super.copyWith((message) => updates(message as ListUsersRequest))
          as ListUsersRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use ListUsersRequest() / ListUsersRequest.new instead')
  static ListUsersRequest create() => ListUsersRequest._();
  static $pb.GeneratedMessage $_createMessage() => ListUsersRequest._();
  @$core.override
  ListUsersRequest createEmptyInstance() => ListUsersRequest._();
  @$core.pragma('dart2js:noInline')
  static ListUsersRequest getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<ListUsersRequest>(
          ListUsersRequest.$_createMessage);
  static ListUsersRequest? _defaultInstance;
}

class ListUsersResponse extends $pb.GeneratedMessage {
  factory ListUsersResponse({
    $core.Iterable<User>? users,
  }) {
    final result = ListUsersResponse._();
    if (users != null) result.users.addAll(users);
    return result;
  }

  ListUsersResponse._();

  factory ListUsersResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListUsersResponse()..mergeFromBuffer(data, registry);
  factory ListUsersResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListUsersResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListUsersResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: ListUsersResponse.$_createMessage)
    ..pPM<User>(1, _omitFieldNames ? '' : 'users',
        subBuilder: User.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListUsersResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListUsersResponse copyWith(void Function(ListUsersResponse) updates) =>
      super.copyWith((message) => updates(message as ListUsersResponse))
          as ListUsersResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use ListUsersResponse() / ListUsersResponse.new instead')
  static ListUsersResponse create() => ListUsersResponse._();
  static $pb.GeneratedMessage $_createMessage() => ListUsersResponse._();
  @$core.override
  ListUsersResponse createEmptyInstance() => ListUsersResponse._();
  @$core.pragma('dart2js:noInline')
  static ListUsersResponse getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<ListUsersResponse>(
          ListUsersResponse.$_createMessage);
  static ListUsersResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $pb.PbList<User> get users => $_getList(0);
}

class CreateUserRequest extends $pb.GeneratedMessage {
  factory CreateUserRequest({
    $core.String? username,
    $core.String? password,
    $core.String? displayName,
    $core.String? role,
    $core.bool? disabled,
  }) {
    final result = CreateUserRequest._();
    if (username != null) result.username = username;
    if (password != null) result.password = password;
    if (displayName != null) result.displayName = displayName;
    if (role != null) result.role = role;
    if (disabled != null) result.disabled = disabled;
    return result;
  }

  CreateUserRequest._();

  factory CreateUserRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateUserRequest()..mergeFromBuffer(data, registry);
  factory CreateUserRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateUserRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CreateUserRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: CreateUserRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'username')
    ..aOS(2, _omitFieldNames ? '' : 'password')
    ..aOS(3, _omitFieldNames ? '' : 'displayName')
    ..aOS(4, _omitFieldNames ? '' : 'role')
    ..aOB(5, _omitFieldNames ? '' : 'disabled')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateUserRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateUserRequest copyWith(void Function(CreateUserRequest) updates) =>
      super.copyWith((message) => updates(message as CreateUserRequest))
          as CreateUserRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use CreateUserRequest() / CreateUserRequest.new instead')
  static CreateUserRequest create() => CreateUserRequest._();
  static $pb.GeneratedMessage $_createMessage() => CreateUserRequest._();
  @$core.override
  CreateUserRequest createEmptyInstance() => CreateUserRequest._();
  @$core.pragma('dart2js:noInline')
  static CreateUserRequest getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<CreateUserRequest>(
          CreateUserRequest.$_createMessage);
  static CreateUserRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get username => $_getSZ(0);
  @$pb.TagNumber(1)
  set username($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasUsername() => $_has(0);
  @$pb.TagNumber(1)
  void clearUsername() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get password => $_getSZ(1);
  @$pb.TagNumber(2)
  set password($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasPassword() => $_has(1);
  @$pb.TagNumber(2)
  void clearPassword() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get displayName => $_getSZ(2);
  @$pb.TagNumber(3)
  set displayName($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasDisplayName() => $_has(2);
  @$pb.TagNumber(3)
  void clearDisplayName() => $_clearField(3);

  @$pb.TagNumber(4)
  $core.String get role => $_getSZ(3);
  @$pb.TagNumber(4)
  set role($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasRole() => $_has(3);
  @$pb.TagNumber(4)
  void clearRole() => $_clearField(4);

  @$pb.TagNumber(5)
  $core.bool get disabled => $_getBF(4);
  @$pb.TagNumber(5)
  set disabled($core.bool value) => $_setBool(4, value);
  @$pb.TagNumber(5)
  $core.bool hasDisabled() => $_has(4);
  @$pb.TagNumber(5)
  void clearDisabled() => $_clearField(5);
}

class CreateUserResponse extends $pb.GeneratedMessage {
  factory CreateUserResponse({
    User? user,
  }) {
    final result = CreateUserResponse._();
    if (user != null) result.user = user;
    return result;
  }

  CreateUserResponse._();

  factory CreateUserResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateUserResponse()..mergeFromBuffer(data, registry);
  factory CreateUserResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CreateUserResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CreateUserResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: CreateUserResponse.$_createMessage)
    ..aOM<User>(1, _omitFieldNames ? '' : 'user',
        subBuilder: User.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateUserResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateUserResponse copyWith(void Function(CreateUserResponse) updates) =>
      super.copyWith((message) => updates(message as CreateUserResponse))
          as CreateUserResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use CreateUserResponse() / CreateUserResponse.new instead')
  static CreateUserResponse create() => CreateUserResponse._();
  static $pb.GeneratedMessage $_createMessage() => CreateUserResponse._();
  @$core.override
  CreateUserResponse createEmptyInstance() => CreateUserResponse._();
  @$core.pragma('dart2js:noInline')
  static CreateUserResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<CreateUserResponse>(
          CreateUserResponse.$_createMessage);
  static CreateUserResponse? _defaultInstance;

  @$pb.TagNumber(1)
  User get user => $_getN(0);
  @$pb.TagNumber(1)
  set user(User value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasUser() => $_has(0);
  @$pb.TagNumber(1)
  void clearUser() => $_clearField(1);
  @$pb.TagNumber(1)
  User ensureUser() => $_ensure(0);
}

class UpdateUserPasswordRequest extends $pb.GeneratedMessage {
  factory UpdateUserPasswordRequest({
    $core.String? id,
    $core.String? password,
  }) {
    final result = UpdateUserPasswordRequest._();
    if (id != null) result.id = id;
    if (password != null) result.password = password;
    return result;
  }

  UpdateUserPasswordRequest._();

  factory UpdateUserPasswordRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpdateUserPasswordRequest()..mergeFromBuffer(data, registry);
  factory UpdateUserPasswordRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpdateUserPasswordRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'UpdateUserPasswordRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: UpdateUserPasswordRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..aOS(2, _omitFieldNames ? '' : 'password')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateUserPasswordRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateUserPasswordRequest copyWith(
          void Function(UpdateUserPasswordRequest) updates) =>
      super.copyWith((message) => updates(message as UpdateUserPasswordRequest))
          as UpdateUserPasswordRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use UpdateUserPasswordRequest() / UpdateUserPasswordRequest.new instead')
  static UpdateUserPasswordRequest create() => UpdateUserPasswordRequest._();
  static $pb.GeneratedMessage $_createMessage() =>
      UpdateUserPasswordRequest._();
  @$core.override
  UpdateUserPasswordRequest createEmptyInstance() =>
      UpdateUserPasswordRequest._();
  @$core.pragma('dart2js:noInline')
  static UpdateUserPasswordRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<UpdateUserPasswordRequest>(
          UpdateUserPasswordRequest.$_createMessage);
  static UpdateUserPasswordRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get password => $_getSZ(1);
  @$pb.TagNumber(2)
  set password($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasPassword() => $_has(1);
  @$pb.TagNumber(2)
  void clearPassword() => $_clearField(2);
}

class UpdateUserPasswordResponse extends $pb.GeneratedMessage {
  factory UpdateUserPasswordResponse({
    User? user,
  }) {
    final result = UpdateUserPasswordResponse._();
    if (user != null) result.user = user;
    return result;
  }

  UpdateUserPasswordResponse._();

  factory UpdateUserPasswordResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpdateUserPasswordResponse()..mergeFromBuffer(data, registry);
  factory UpdateUserPasswordResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpdateUserPasswordResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'UpdateUserPasswordResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: UpdateUserPasswordResponse.$_createMessage)
    ..aOM<User>(1, _omitFieldNames ? '' : 'user',
        subBuilder: User.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateUserPasswordResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateUserPasswordResponse copyWith(
          void Function(UpdateUserPasswordResponse) updates) =>
      super.copyWith(
              (message) => updates(message as UpdateUserPasswordResponse))
          as UpdateUserPasswordResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use UpdateUserPasswordResponse() / UpdateUserPasswordResponse.new instead')
  static UpdateUserPasswordResponse create() => UpdateUserPasswordResponse._();
  static $pb.GeneratedMessage $_createMessage() =>
      UpdateUserPasswordResponse._();
  @$core.override
  UpdateUserPasswordResponse createEmptyInstance() =>
      UpdateUserPasswordResponse._();
  @$core.pragma('dart2js:noInline')
  static UpdateUserPasswordResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<UpdateUserPasswordResponse>(
          UpdateUserPasswordResponse.$_createMessage);
  static UpdateUserPasswordResponse? _defaultInstance;

  @$pb.TagNumber(1)
  User get user => $_getN(0);
  @$pb.TagNumber(1)
  set user(User value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasUser() => $_has(0);
  @$pb.TagNumber(1)
  void clearUser() => $_clearField(1);
  @$pb.TagNumber(1)
  User ensureUser() => $_ensure(0);
}

class SetUserDisabledRequest extends $pb.GeneratedMessage {
  factory SetUserDisabledRequest({
    $core.String? id,
    $core.bool? disabled,
  }) {
    final result = SetUserDisabledRequest._();
    if (id != null) result.id = id;
    if (disabled != null) result.disabled = disabled;
    return result;
  }

  SetUserDisabledRequest._();

  factory SetUserDisabledRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      SetUserDisabledRequest()..mergeFromBuffer(data, registry);
  factory SetUserDisabledRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      SetUserDisabledRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'SetUserDisabledRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: SetUserDisabledRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..aOB(2, _omitFieldNames ? '' : 'disabled')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  SetUserDisabledRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  SetUserDisabledRequest copyWith(
          void Function(SetUserDisabledRequest) updates) =>
      super.copyWith((message) => updates(message as SetUserDisabledRequest))
          as SetUserDisabledRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use SetUserDisabledRequest() / SetUserDisabledRequest.new instead')
  static SetUserDisabledRequest create() => SetUserDisabledRequest._();
  static $pb.GeneratedMessage $_createMessage() => SetUserDisabledRequest._();
  @$core.override
  SetUserDisabledRequest createEmptyInstance() => SetUserDisabledRequest._();
  @$core.pragma('dart2js:noInline')
  static SetUserDisabledRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<SetUserDisabledRequest>(
          SetUserDisabledRequest.$_createMessage);
  static SetUserDisabledRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.bool get disabled => $_getBF(1);
  @$pb.TagNumber(2)
  set disabled($core.bool value) => $_setBool(1, value);
  @$pb.TagNumber(2)
  $core.bool hasDisabled() => $_has(1);
  @$pb.TagNumber(2)
  void clearDisabled() => $_clearField(2);
}

class SetUserDisabledResponse extends $pb.GeneratedMessage {
  factory SetUserDisabledResponse({
    User? user,
  }) {
    final result = SetUserDisabledResponse._();
    if (user != null) result.user = user;
    return result;
  }

  SetUserDisabledResponse._();

  factory SetUserDisabledResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      SetUserDisabledResponse()..mergeFromBuffer(data, registry);
  factory SetUserDisabledResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      SetUserDisabledResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'SetUserDisabledResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: SetUserDisabledResponse.$_createMessage)
    ..aOM<User>(1, _omitFieldNames ? '' : 'user',
        subBuilder: User.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  SetUserDisabledResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  SetUserDisabledResponse copyWith(
          void Function(SetUserDisabledResponse) updates) =>
      super.copyWith((message) => updates(message as SetUserDisabledResponse))
          as SetUserDisabledResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use SetUserDisabledResponse() / SetUserDisabledResponse.new instead')
  static SetUserDisabledResponse create() => SetUserDisabledResponse._();
  static $pb.GeneratedMessage $_createMessage() => SetUserDisabledResponse._();
  @$core.override
  SetUserDisabledResponse createEmptyInstance() => SetUserDisabledResponse._();
  @$core.pragma('dart2js:noInline')
  static SetUserDisabledResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<SetUserDisabledResponse>(
          SetUserDisabledResponse.$_createMessage);
  static SetUserDisabledResponse? _defaultInstance;

  @$pb.TagNumber(1)
  User get user => $_getN(0);
  @$pb.TagNumber(1)
  set user(User value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasUser() => $_has(0);
  @$pb.TagNumber(1)
  void clearUser() => $_clearField(1);
  @$pb.TagNumber(1)
  User ensureUser() => $_ensure(0);
}

class UpdateProfileRequest extends $pb.GeneratedMessage {
  factory UpdateProfileRequest({
    $core.String? displayName,
    $core.String? avatarUrl,
  }) {
    final result = UpdateProfileRequest._();
    if (displayName != null) result.displayName = displayName;
    if (avatarUrl != null) result.avatarUrl = avatarUrl;
    return result;
  }

  UpdateProfileRequest._();

  factory UpdateProfileRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpdateProfileRequest()..mergeFromBuffer(data, registry);
  factory UpdateProfileRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpdateProfileRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'UpdateProfileRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: UpdateProfileRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'displayName')
    ..aOS(2, _omitFieldNames ? '' : 'avatarUrl')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateProfileRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateProfileRequest copyWith(void Function(UpdateProfileRequest) updates) =>
      super.copyWith((message) => updates(message as UpdateProfileRequest))
          as UpdateProfileRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use UpdateProfileRequest() / UpdateProfileRequest.new instead')
  static UpdateProfileRequest create() => UpdateProfileRequest._();
  static $pb.GeneratedMessage $_createMessage() => UpdateProfileRequest._();
  @$core.override
  UpdateProfileRequest createEmptyInstance() => UpdateProfileRequest._();
  @$core.pragma('dart2js:noInline')
  static UpdateProfileRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<UpdateProfileRequest>(
          UpdateProfileRequest.$_createMessage);
  static UpdateProfileRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get displayName => $_getSZ(0);
  @$pb.TagNumber(1)
  set displayName($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasDisplayName() => $_has(0);
  @$pb.TagNumber(1)
  void clearDisplayName() => $_clearField(1);

  /// Avatar URL. When unset the existing value is preserved; sending an
  /// explicit empty string clears the stored avatar.
  @$pb.TagNumber(2)
  $core.String get avatarUrl => $_getSZ(1);
  @$pb.TagNumber(2)
  set avatarUrl($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasAvatarUrl() => $_has(1);
  @$pb.TagNumber(2)
  void clearAvatarUrl() => $_clearField(2);
}

class UpdateProfileResponse extends $pb.GeneratedMessage {
  factory UpdateProfileResponse({
    User? user,
  }) {
    final result = UpdateProfileResponse._();
    if (user != null) result.user = user;
    return result;
  }

  UpdateProfileResponse._();

  factory UpdateProfileResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpdateProfileResponse()..mergeFromBuffer(data, registry);
  factory UpdateProfileResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      UpdateProfileResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'UpdateProfileResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: UpdateProfileResponse.$_createMessage)
    ..aOM<User>(1, _omitFieldNames ? '' : 'user',
        subBuilder: User.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateProfileResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UpdateProfileResponse copyWith(
          void Function(UpdateProfileResponse) updates) =>
      super.copyWith((message) => updates(message as UpdateProfileResponse))
          as UpdateProfileResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use UpdateProfileResponse() / UpdateProfileResponse.new instead')
  static UpdateProfileResponse create() => UpdateProfileResponse._();
  static $pb.GeneratedMessage $_createMessage() => UpdateProfileResponse._();
  @$core.override
  UpdateProfileResponse createEmptyInstance() => UpdateProfileResponse._();
  @$core.pragma('dart2js:noInline')
  static UpdateProfileResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<UpdateProfileResponse>(
          UpdateProfileResponse.$_createMessage);
  static UpdateProfileResponse? _defaultInstance;

  @$pb.TagNumber(1)
  User get user => $_getN(0);
  @$pb.TagNumber(1)
  set user(User value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasUser() => $_has(0);
  @$pb.TagNumber(1)
  void clearUser() => $_clearField(1);
  @$pb.TagNumber(1)
  User ensureUser() => $_ensure(0);
}

class ChangePasswordRequest extends $pb.GeneratedMessage {
  factory ChangePasswordRequest({
    $core.String? currentPassword,
    $core.String? newPassword,
  }) {
    final result = ChangePasswordRequest._();
    if (currentPassword != null) result.currentPassword = currentPassword;
    if (newPassword != null) result.newPassword = newPassword;
    return result;
  }

  ChangePasswordRequest._();

  factory ChangePasswordRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ChangePasswordRequest()..mergeFromBuffer(data, registry);
  factory ChangePasswordRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ChangePasswordRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ChangePasswordRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: ChangePasswordRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'currentPassword')
    ..aOS(2, _omitFieldNames ? '' : 'newPassword')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ChangePasswordRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ChangePasswordRequest copyWith(
          void Function(ChangePasswordRequest) updates) =>
      super.copyWith((message) => updates(message as ChangePasswordRequest))
          as ChangePasswordRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use ChangePasswordRequest() / ChangePasswordRequest.new instead')
  static ChangePasswordRequest create() => ChangePasswordRequest._();
  static $pb.GeneratedMessage $_createMessage() => ChangePasswordRequest._();
  @$core.override
  ChangePasswordRequest createEmptyInstance() => ChangePasswordRequest._();
  @$core.pragma('dart2js:noInline')
  static ChangePasswordRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ChangePasswordRequest>(
          ChangePasswordRequest.$_createMessage);
  static ChangePasswordRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get currentPassword => $_getSZ(0);
  @$pb.TagNumber(1)
  set currentPassword($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasCurrentPassword() => $_has(0);
  @$pb.TagNumber(1)
  void clearCurrentPassword() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get newPassword => $_getSZ(1);
  @$pb.TagNumber(2)
  set newPassword($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasNewPassword() => $_has(1);
  @$pb.TagNumber(2)
  void clearNewPassword() => $_clearField(2);
}

class ChangePasswordResponse extends $pb.GeneratedMessage {
  factory ChangePasswordResponse({
    User? user,
  }) {
    final result = ChangePasswordResponse._();
    if (user != null) result.user = user;
    return result;
  }

  ChangePasswordResponse._();

  factory ChangePasswordResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ChangePasswordResponse()..mergeFromBuffer(data, registry);
  factory ChangePasswordResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ChangePasswordResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ChangePasswordResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: ChangePasswordResponse.$_createMessage)
    ..aOM<User>(1, _omitFieldNames ? '' : 'user',
        subBuilder: User.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ChangePasswordResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ChangePasswordResponse copyWith(
          void Function(ChangePasswordResponse) updates) =>
      super.copyWith((message) => updates(message as ChangePasswordResponse))
          as ChangePasswordResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use ChangePasswordResponse() / ChangePasswordResponse.new instead')
  static ChangePasswordResponse create() => ChangePasswordResponse._();
  static $pb.GeneratedMessage $_createMessage() => ChangePasswordResponse._();
  @$core.override
  ChangePasswordResponse createEmptyInstance() => ChangePasswordResponse._();
  @$core.pragma('dart2js:noInline')
  static ChangePasswordResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ChangePasswordResponse>(
          ChangePasswordResponse.$_createMessage);
  static ChangePasswordResponse? _defaultInstance;

  @$pb.TagNumber(1)
  User get user => $_getN(0);
  @$pb.TagNumber(1)
  set user(User value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasUser() => $_has(0);
  @$pb.TagNumber(1)
  void clearUser() => $_clearField(1);
  @$pb.TagNumber(1)
  User ensureUser() => $_ensure(0);
}

class OAuthProvider extends $pb.GeneratedMessage {
  factory OAuthProvider({
    $core.String? name,
    $core.String? displayName,
  }) {
    final result = OAuthProvider._();
    if (name != null) result.name = name;
    if (displayName != null) result.displayName = displayName;
    return result;
  }

  OAuthProvider._();

  factory OAuthProvider.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      OAuthProvider()..mergeFromBuffer(data, registry);
  factory OAuthProvider.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      OAuthProvider()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'OAuthProvider',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: OAuthProvider.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'name')
    ..aOS(2, _omitFieldNames ? '' : 'displayName')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  OAuthProvider clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  OAuthProvider copyWith(void Function(OAuthProvider) updates) =>
      super.copyWith((message) => updates(message as OAuthProvider))
          as OAuthProvider;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use OAuthProvider() / OAuthProvider.new instead')
  static OAuthProvider create() => OAuthProvider._();
  static $pb.GeneratedMessage $_createMessage() => OAuthProvider._();
  @$core.override
  OAuthProvider createEmptyInstance() => OAuthProvider._();
  @$core.pragma('dart2js:noInline')
  static OAuthProvider getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<OAuthProvider>(
          OAuthProvider.$_createMessage);
  static OAuthProvider? _defaultInstance;

  /// Provider key (e.g. "github"). Used as the value for BeginOAuthFlow.
  @$pb.TagNumber(1)
  $core.String get name => $_getSZ(0);
  @$pb.TagNumber(1)
  set name($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasName() => $_has(0);
  @$pb.TagNumber(1)
  void clearName() => $_clearField(1);

  /// Human-readable label, e.g. "GitHub".
  @$pb.TagNumber(2)
  $core.String get displayName => $_getSZ(1);
  @$pb.TagNumber(2)
  set displayName($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasDisplayName() => $_has(1);
  @$pb.TagNumber(2)
  void clearDisplayName() => $_clearField(2);
}

class ListOAuthProvidersRequest extends $pb.GeneratedMessage {
  factory ListOAuthProvidersRequest() => ListOAuthProvidersRequest._();

  ListOAuthProvidersRequest._();

  factory ListOAuthProvidersRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListOAuthProvidersRequest()..mergeFromBuffer(data, registry);
  factory ListOAuthProvidersRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListOAuthProvidersRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListOAuthProvidersRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: ListOAuthProvidersRequest.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListOAuthProvidersRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListOAuthProvidersRequest copyWith(
          void Function(ListOAuthProvidersRequest) updates) =>
      super.copyWith((message) => updates(message as ListOAuthProvidersRequest))
          as ListOAuthProvidersRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use ListOAuthProvidersRequest() / ListOAuthProvidersRequest.new instead')
  static ListOAuthProvidersRequest create() => ListOAuthProvidersRequest._();
  static $pb.GeneratedMessage $_createMessage() =>
      ListOAuthProvidersRequest._();
  @$core.override
  ListOAuthProvidersRequest createEmptyInstance() =>
      ListOAuthProvidersRequest._();
  @$core.pragma('dart2js:noInline')
  static ListOAuthProvidersRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ListOAuthProvidersRequest>(
          ListOAuthProvidersRequest.$_createMessage);
  static ListOAuthProvidersRequest? _defaultInstance;
}

class ListOAuthProvidersResponse extends $pb.GeneratedMessage {
  factory ListOAuthProvidersResponse({
    $core.Iterable<OAuthProvider>? providers,
  }) {
    final result = ListOAuthProvidersResponse._();
    if (providers != null) result.providers.addAll(providers);
    return result;
  }

  ListOAuthProvidersResponse._();

  factory ListOAuthProvidersResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListOAuthProvidersResponse()..mergeFromBuffer(data, registry);
  factory ListOAuthProvidersResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListOAuthProvidersResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListOAuthProvidersResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: ListOAuthProvidersResponse.$_createMessage)
    ..pPM<OAuthProvider>(1, _omitFieldNames ? '' : 'providers',
        subBuilder: OAuthProvider.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListOAuthProvidersResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListOAuthProvidersResponse copyWith(
          void Function(ListOAuthProvidersResponse) updates) =>
      super.copyWith(
              (message) => updates(message as ListOAuthProvidersResponse))
          as ListOAuthProvidersResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use ListOAuthProvidersResponse() / ListOAuthProvidersResponse.new instead')
  static ListOAuthProvidersResponse create() => ListOAuthProvidersResponse._();
  static $pb.GeneratedMessage $_createMessage() =>
      ListOAuthProvidersResponse._();
  @$core.override
  ListOAuthProvidersResponse createEmptyInstance() =>
      ListOAuthProvidersResponse._();
  @$core.pragma('dart2js:noInline')
  static ListOAuthProvidersResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ListOAuthProvidersResponse>(
          ListOAuthProvidersResponse.$_createMessage);
  static ListOAuthProvidersResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $pb.PbList<OAuthProvider> get providers => $_getList(0);
}

class BeginOAuthFlowRequest extends $pb.GeneratedMessage {
  factory BeginOAuthFlowRequest({
    $core.String? provider,
    $core.String? redirectUri,
  }) {
    final result = BeginOAuthFlowRequest._();
    if (provider != null) result.provider = provider;
    if (redirectUri != null) result.redirectUri = redirectUri;
    return result;
  }

  BeginOAuthFlowRequest._();

  factory BeginOAuthFlowRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      BeginOAuthFlowRequest()..mergeFromBuffer(data, registry);
  factory BeginOAuthFlowRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      BeginOAuthFlowRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'BeginOAuthFlowRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: BeginOAuthFlowRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'provider')
    ..aOS(2, _omitFieldNames ? '' : 'redirectUri')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  BeginOAuthFlowRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  BeginOAuthFlowRequest copyWith(
          void Function(BeginOAuthFlowRequest) updates) =>
      super.copyWith((message) => updates(message as BeginOAuthFlowRequest))
          as BeginOAuthFlowRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use BeginOAuthFlowRequest() / BeginOAuthFlowRequest.new instead')
  static BeginOAuthFlowRequest create() => BeginOAuthFlowRequest._();
  static $pb.GeneratedMessage $_createMessage() => BeginOAuthFlowRequest._();
  @$core.override
  BeginOAuthFlowRequest createEmptyInstance() => BeginOAuthFlowRequest._();
  @$core.pragma('dart2js:noInline')
  static BeginOAuthFlowRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<BeginOAuthFlowRequest>(
          BeginOAuthFlowRequest.$_createMessage);
  static BeginOAuthFlowRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get provider => $_getSZ(0);
  @$pb.TagNumber(1)
  set provider($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasProvider() => $_has(0);
  @$pb.TagNumber(1)
  void clearProvider() => $_clearField(1);

  /// Where the browser should land after a successful login. The backend
  /// returns this as part of the state so CompleteOAuthFlow can echo it back.
  @$pb.TagNumber(2)
  $core.String get redirectUri => $_getSZ(1);
  @$pb.TagNumber(2)
  set redirectUri($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasRedirectUri() => $_has(1);
  @$pb.TagNumber(2)
  void clearRedirectUri() => $_clearField(2);
}

class BeginOAuthFlowResponse extends $pb.GeneratedMessage {
  factory BeginOAuthFlowResponse({
    $core.String? authorizeUrl,
    $core.String? state,
  }) {
    final result = BeginOAuthFlowResponse._();
    if (authorizeUrl != null) result.authorizeUrl = authorizeUrl;
    if (state != null) result.state = state;
    return result;
  }

  BeginOAuthFlowResponse._();

  factory BeginOAuthFlowResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      BeginOAuthFlowResponse()..mergeFromBuffer(data, registry);
  factory BeginOAuthFlowResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      BeginOAuthFlowResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'BeginOAuthFlowResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: BeginOAuthFlowResponse.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'authorizeUrl')
    ..aOS(2, _omitFieldNames ? '' : 'state')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  BeginOAuthFlowResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  BeginOAuthFlowResponse copyWith(
          void Function(BeginOAuthFlowResponse) updates) =>
      super.copyWith((message) => updates(message as BeginOAuthFlowResponse))
          as BeginOAuthFlowResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use BeginOAuthFlowResponse() / BeginOAuthFlowResponse.new instead')
  static BeginOAuthFlowResponse create() => BeginOAuthFlowResponse._();
  static $pb.GeneratedMessage $_createMessage() => BeginOAuthFlowResponse._();
  @$core.override
  BeginOAuthFlowResponse createEmptyInstance() => BeginOAuthFlowResponse._();
  @$core.pragma('dart2js:noInline')
  static BeginOAuthFlowResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<BeginOAuthFlowResponse>(
          BeginOAuthFlowResponse.$_createMessage);
  static BeginOAuthFlowResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get authorizeUrl => $_getSZ(0);
  @$pb.TagNumber(1)
  set authorizeUrl($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasAuthorizeUrl() => $_has(0);
  @$pb.TagNumber(1)
  void clearAuthorizeUrl() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get state => $_getSZ(1);
  @$pb.TagNumber(2)
  set state($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasState() => $_has(1);
  @$pb.TagNumber(2)
  void clearState() => $_clearField(2);
}

class CompleteOAuthFlowRequest extends $pb.GeneratedMessage {
  factory CompleteOAuthFlowRequest({
    $core.String? provider,
    $core.String? code,
    $core.String? state,
  }) {
    final result = CompleteOAuthFlowRequest._();
    if (provider != null) result.provider = provider;
    if (code != null) result.code = code;
    if (state != null) result.state = state;
    return result;
  }

  CompleteOAuthFlowRequest._();

  factory CompleteOAuthFlowRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CompleteOAuthFlowRequest()..mergeFromBuffer(data, registry);
  factory CompleteOAuthFlowRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CompleteOAuthFlowRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CompleteOAuthFlowRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: CompleteOAuthFlowRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'provider')
    ..aOS(2, _omitFieldNames ? '' : 'code')
    ..aOS(3, _omitFieldNames ? '' : 'state')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CompleteOAuthFlowRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CompleteOAuthFlowRequest copyWith(
          void Function(CompleteOAuthFlowRequest) updates) =>
      super.copyWith((message) => updates(message as CompleteOAuthFlowRequest))
          as CompleteOAuthFlowRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use CompleteOAuthFlowRequest() / CompleteOAuthFlowRequest.new instead')
  static CompleteOAuthFlowRequest create() => CompleteOAuthFlowRequest._();
  static $pb.GeneratedMessage $_createMessage() => CompleteOAuthFlowRequest._();
  @$core.override
  CompleteOAuthFlowRequest createEmptyInstance() =>
      CompleteOAuthFlowRequest._();
  @$core.pragma('dart2js:noInline')
  static CompleteOAuthFlowRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<CompleteOAuthFlowRequest>(
          CompleteOAuthFlowRequest.$_createMessage);
  static CompleteOAuthFlowRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get provider => $_getSZ(0);
  @$pb.TagNumber(1)
  set provider($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasProvider() => $_has(0);
  @$pb.TagNumber(1)
  void clearProvider() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get code => $_getSZ(1);
  @$pb.TagNumber(2)
  set code($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasCode() => $_has(1);
  @$pb.TagNumber(2)
  void clearCode() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get state => $_getSZ(2);
  @$pb.TagNumber(3)
  set state($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasState() => $_has(2);
  @$pb.TagNumber(3)
  void clearState() => $_clearField(3);
}

/// CompleteOAuthFlowResponse mirrors LoginResponse — duplicated rather than
/// shared so each RPC has a dedicated response type (buf STANDARD lint).
class CompleteOAuthFlowResponse extends $pb.GeneratedMessage {
  factory CompleteOAuthFlowResponse({
    $core.String? token,
    User? user,
    $0.Timestamp? expiresAt,
    $core.Iterable<$1.Workspace>? workspaces,
  }) {
    final result = CompleteOAuthFlowResponse._();
    if (token != null) result.token = token;
    if (user != null) result.user = user;
    if (expiresAt != null) result.expiresAt = expiresAt;
    if (workspaces != null) result.workspaces.addAll(workspaces);
    return result;
  }

  CompleteOAuthFlowResponse._();

  factory CompleteOAuthFlowResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CompleteOAuthFlowResponse()..mergeFromBuffer(data, registry);
  factory CompleteOAuthFlowResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CompleteOAuthFlowResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CompleteOAuthFlowResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: CompleteOAuthFlowResponse.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'token')
    ..aOM<User>(2, _omitFieldNames ? '' : 'user',
        subBuilder: User.$_createMessage)
    ..aOM<$0.Timestamp>(3, _omitFieldNames ? '' : 'expiresAt',
        subBuilder: $0.Timestamp.$_createMessage)
    ..pPM<$1.Workspace>(4, _omitFieldNames ? '' : 'workspaces',
        subBuilder: $1.Workspace.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CompleteOAuthFlowResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CompleteOAuthFlowResponse copyWith(
          void Function(CompleteOAuthFlowResponse) updates) =>
      super.copyWith((message) => updates(message as CompleteOAuthFlowResponse))
          as CompleteOAuthFlowResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use CompleteOAuthFlowResponse() / CompleteOAuthFlowResponse.new instead')
  static CompleteOAuthFlowResponse create() => CompleteOAuthFlowResponse._();
  static $pb.GeneratedMessage $_createMessage() =>
      CompleteOAuthFlowResponse._();
  @$core.override
  CompleteOAuthFlowResponse createEmptyInstance() =>
      CompleteOAuthFlowResponse._();
  @$core.pragma('dart2js:noInline')
  static CompleteOAuthFlowResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<CompleteOAuthFlowResponse>(
          CompleteOAuthFlowResponse.$_createMessage);
  static CompleteOAuthFlowResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get token => $_getSZ(0);
  @$pb.TagNumber(1)
  set token($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasToken() => $_has(0);
  @$pb.TagNumber(1)
  void clearToken() => $_clearField(1);

  @$pb.TagNumber(2)
  User get user => $_getN(1);
  @$pb.TagNumber(2)
  set user(User value) => $_setField(2, value);
  @$pb.TagNumber(2)
  $core.bool hasUser() => $_has(1);
  @$pb.TagNumber(2)
  void clearUser() => $_clearField(2);
  @$pb.TagNumber(2)
  User ensureUser() => $_ensure(1);

  @$pb.TagNumber(3)
  $0.Timestamp get expiresAt => $_getN(2);
  @$pb.TagNumber(3)
  set expiresAt($0.Timestamp value) => $_setField(3, value);
  @$pb.TagNumber(3)
  $core.bool hasExpiresAt() => $_has(2);
  @$pb.TagNumber(3)
  void clearExpiresAt() => $_clearField(3);
  @$pb.TagNumber(3)
  $0.Timestamp ensureExpiresAt() => $_ensure(2);

  @$pb.TagNumber(4)
  $pb.PbList<$1.Workspace> get workspaces => $_getList(3);
}

/// AuthService manages dashboard user authentication.
class AuthServiceApi {
  final $pb.RpcClient _client;

  AuthServiceApi(this._client);

  /// Login is public and exchanges username/password for a session token.
  $async.Future<LoginResponse> login(
          $pb.ClientContext? ctx, LoginRequest request) =>
      _client.invoke<LoginResponse>(
          ctx, 'AuthService', 'Login', request, LoginResponse());

  /// Me returns the authenticated user for the current session.
  $async.Future<MeResponse> me($pb.ClientContext? ctx, MeRequest request) =>
      _client.invoke<MeResponse>(
          ctx, 'AuthService', 'Me', request, MeResponse());

  /// Logout revokes the current session token.
  $async.Future<LogoutResponse> logout(
          $pb.ClientContext? ctx, LogoutRequest request) =>
      _client.invoke<LogoutResponse>(
          ctx, 'AuthService', 'Logout', request, LogoutResponse());

  /// ListUsers lists dashboard users. Admin only.
  $async.Future<ListUsersResponse> listUsers(
          $pb.ClientContext? ctx, ListUsersRequest request) =>
      _client.invoke<ListUsersResponse>(
          ctx, 'AuthService', 'ListUsers', request, ListUsersResponse());

  /// CreateUser creates a dashboard user. Admin only.
  $async.Future<CreateUserResponse> createUser(
          $pb.ClientContext? ctx, CreateUserRequest request) =>
      _client.invoke<CreateUserResponse>(
          ctx, 'AuthService', 'CreateUser', request, CreateUserResponse());

  /// UpdateUserPassword changes a user's password. Admin only.
  $async.Future<UpdateUserPasswordResponse> updateUserPassword(
          $pb.ClientContext? ctx, UpdateUserPasswordRequest request) =>
      _client.invoke<UpdateUserPasswordResponse>(ctx, 'AuthService',
          'UpdateUserPassword', request, UpdateUserPasswordResponse());

  /// SetUserDisabled enables or disables a user. Admin only.
  $async.Future<SetUserDisabledResponse> setUserDisabled(
          $pb.ClientContext? ctx, SetUserDisabledRequest request) =>
      _client.invoke<SetUserDisabledResponse>(ctx, 'AuthService',
          'SetUserDisabled', request, SetUserDisabledResponse());

  /// UpdateProfile updates the authenticated user's display name.
  $async.Future<UpdateProfileResponse> updateProfile(
          $pb.ClientContext? ctx, UpdateProfileRequest request) =>
      _client.invoke<UpdateProfileResponse>(ctx, 'AuthService', 'UpdateProfile',
          request, UpdateProfileResponse());

  /// ChangePassword changes the authenticated user's password.
  $async.Future<ChangePasswordResponse> changePassword(
          $pb.ClientContext? ctx, ChangePasswordRequest request) =>
      _client.invoke<ChangePasswordResponse>(ctx, 'AuthService',
          'ChangePassword', request, ChangePasswordResponse());

  /// ListOAuthProviders is public and returns the OAuth providers configured
  /// on the server. Frontends use it to decide which "Sign in with X" buttons
  /// to render on the login page.
  $async.Future<ListOAuthProvidersResponse> listOAuthProviders(
          $pb.ClientContext? ctx, ListOAuthProvidersRequest request) =>
      _client.invoke<ListOAuthProvidersResponse>(ctx, 'AuthService',
          'ListOAuthProviders', request, ListOAuthProvidersResponse());

  /// BeginOAuthFlow is public and returns the provider authorize URL plus an
  /// opaque state token. The client redirects the user to authorize_url.
  $async.Future<BeginOAuthFlowResponse> beginOAuthFlow(
          $pb.ClientContext? ctx, BeginOAuthFlowRequest request) =>
      _client.invoke<BeginOAuthFlowResponse>(ctx, 'AuthService',
          'BeginOAuthFlow', request, BeginOAuthFlowResponse());

  /// CompleteOAuthFlow is public and exchanges the provider's authorization
  /// code for a session token. The response mirrors LoginResponse.
  $async.Future<CompleteOAuthFlowResponse> completeOAuthFlow(
          $pb.ClientContext? ctx, CompleteOAuthFlowRequest request) =>
      _client.invoke<CompleteOAuthFlowResponse>(ctx, 'AuthService',
          'CompleteOAuthFlow', request, CompleteOAuthFlowResponse());
}

const $core.bool _omitFieldNames =
    $core.bool.fromEnvironment('protobuf.omit_field_names');
const $core.bool _omitMessageNames =
    $core.bool.fromEnvironment('protobuf.omit_message_names');
