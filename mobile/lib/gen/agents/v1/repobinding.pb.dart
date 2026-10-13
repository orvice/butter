// This is a generated file - do not edit.
//
// Generated from agents/v1/repobinding.proto.

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

import 'repobinding.pbenum.dart';

export 'package:protobuf/protobuf.dart' show GeneratedMessageGenericExtensions;

export 'repobinding.pbenum.dart';

class WorkspaceRepoBinding extends $pb.GeneratedMessage {
  factory WorkspaceRepoBinding({
    $core.String? gitHostId,
    $core.String? repository,
    $core.String? branch,
    $core.String? rootPath,
    RepoBindingWriteMode? writeMode,
    $core.int? contentSchemaVersion,
    $core.bool? credentialSet,
    $0.Timestamp? credentialUpdatedAt,
    RepoBindingStatus? status,
    $0.Timestamp? createdAt,
    $0.Timestamp? updatedAt,
    $core.String? observedCommitSha,
    $core.String? activeCommitSha,
    $0.Timestamp? lastSyncedAt,
    $core.String? lastSyncError,
    $core.String? lastPublicationError,
    $core.Iterable<$core.String>? publicationErrors,
    $core.bool? webhookSecretSet,
    $core.String? workspaceId,
  }) {
    final result = WorkspaceRepoBinding._();
    if (gitHostId != null) result.gitHostId = gitHostId;
    if (repository != null) result.repository = repository;
    if (branch != null) result.branch = branch;
    if (rootPath != null) result.rootPath = rootPath;
    if (writeMode != null) result.writeMode = writeMode;
    if (contentSchemaVersion != null)
      result.contentSchemaVersion = contentSchemaVersion;
    if (credentialSet != null) result.credentialSet = credentialSet;
    if (credentialUpdatedAt != null)
      result.credentialUpdatedAt = credentialUpdatedAt;
    if (status != null) result.status = status;
    if (createdAt != null) result.createdAt = createdAt;
    if (updatedAt != null) result.updatedAt = updatedAt;
    if (observedCommitSha != null) result.observedCommitSha = observedCommitSha;
    if (activeCommitSha != null) result.activeCommitSha = activeCommitSha;
    if (lastSyncedAt != null) result.lastSyncedAt = lastSyncedAt;
    if (lastSyncError != null) result.lastSyncError = lastSyncError;
    if (lastPublicationError != null)
      result.lastPublicationError = lastPublicationError;
    if (publicationErrors != null)
      result.publicationErrors.addAll(publicationErrors);
    if (webhookSecretSet != null) result.webhookSecretSet = webhookSecretSet;
    if (workspaceId != null) result.workspaceId = workspaceId;
    return result;
  }

  WorkspaceRepoBinding._();

  factory WorkspaceRepoBinding.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      WorkspaceRepoBinding()..mergeFromBuffer(data, registry);
  factory WorkspaceRepoBinding.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      WorkspaceRepoBinding()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'WorkspaceRepoBinding',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: WorkspaceRepoBinding.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'gitHostId')
    ..aOS(2, _omitFieldNames ? '' : 'repository')
    ..aOS(3, _omitFieldNames ? '' : 'branch')
    ..aOS(4, _omitFieldNames ? '' : 'rootPath')
    ..aE<RepoBindingWriteMode>(5, _omitFieldNames ? '' : 'writeMode',
        enumValues: RepoBindingWriteMode.values)
    ..aI(6, _omitFieldNames ? '' : 'contentSchemaVersion')
    ..aOB(7, _omitFieldNames ? '' : 'credentialSet')
    ..aOM<$0.Timestamp>(8, _omitFieldNames ? '' : 'credentialUpdatedAt',
        subBuilder: $0.Timestamp.$_createMessage)
    ..aOM<RepoBindingStatus>(9, _omitFieldNames ? '' : 'status',
        subBuilder: RepoBindingStatus.$_createMessage)
    ..aOM<$0.Timestamp>(10, _omitFieldNames ? '' : 'createdAt',
        subBuilder: $0.Timestamp.$_createMessage)
    ..aOM<$0.Timestamp>(11, _omitFieldNames ? '' : 'updatedAt',
        subBuilder: $0.Timestamp.$_createMessage)
    ..aOS(12, _omitFieldNames ? '' : 'observedCommitSha')
    ..aOS(13, _omitFieldNames ? '' : 'activeCommitSha')
    ..aOM<$0.Timestamp>(14, _omitFieldNames ? '' : 'lastSyncedAt',
        subBuilder: $0.Timestamp.$_createMessage)
    ..aOS(15, _omitFieldNames ? '' : 'lastSyncError')
    ..aOS(16, _omitFieldNames ? '' : 'lastPublicationError')
    ..pPS(17, _omitFieldNames ? '' : 'publicationErrors')
    ..aOB(18, _omitFieldNames ? '' : 'webhookSecretSet')
    ..aOS(100, _omitFieldNames ? '' : 'workspaceId')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  WorkspaceRepoBinding clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  WorkspaceRepoBinding copyWith(void Function(WorkspaceRepoBinding) updates) =>
      super.copyWith((message) => updates(message as WorkspaceRepoBinding))
          as WorkspaceRepoBinding;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use WorkspaceRepoBinding() / WorkspaceRepoBinding.new instead')
  static WorkspaceRepoBinding create() => WorkspaceRepoBinding._();
  static $pb.GeneratedMessage $_createMessage() => WorkspaceRepoBinding._();
  @$core.override
  WorkspaceRepoBinding createEmptyInstance() => WorkspaceRepoBinding._();
  @$core.pragma('dart2js:noInline')
  static WorkspaceRepoBinding getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<WorkspaceRepoBinding>(
          WorkspaceRepoBinding.$_createMessage);
  static WorkspaceRepoBinding? _defaultInstance;

  /// References the platform-configured GitHost the repository lives on.
  /// @gotags: json:"git_host_id,omitempty" yaml:"git_host_id,omitempty" bson:"git_host_id,omitempty"
  @$pb.TagNumber(1)
  $core.String get gitHostId => $_getSZ(0);
  @$pb.TagNumber(1)
  set gitHostId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasGitHostId() => $_has(0);
  @$pb.TagNumber(1)
  void clearGitHostId() => $_clearField(1);

  /// Host-relative repository path: "owner/repo" on GitHub,
  /// "group[/subgroup]/project" on GitLab.
  /// @gotags: json:"repository,omitempty" yaml:"repository,omitempty" bson:"repository,omitempty"
  @$pb.TagNumber(2)
  $core.String get repository => $_getSZ(1);
  @$pb.TagNumber(2)
  set repository($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasRepository() => $_has(1);
  @$pb.TagNumber(2)
  void clearRepository() => $_clearField(2);

  /// Branch Agent Content is read from and written to.
  /// @gotags: json:"branch,omitempty" yaml:"branch,omitempty" bson:"branch,omitempty"
  @$pb.TagNumber(3)
  $core.String get branch => $_getSZ(2);
  @$pb.TagNumber(3)
  set branch($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasBranch() => $_has(2);
  @$pb.TagNumber(3)
  void clearBranch() => $_clearField(3);

  /// Repository-relative directory that contains the managed `agents/`
  /// subtree. Empty means the repository root. Never absolute, no traversal.
  /// @gotags: json:"root_path,omitempty" yaml:"root_path,omitempty" bson:"root_path,omitempty"
  @$pb.TagNumber(4)
  $core.String get rootPath => $_getSZ(3);
  @$pb.TagNumber(4)
  set rootPath($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasRootPath() => $_has(3);
  @$pb.TagNumber(4)
  void clearRootPath() => $_clearField(4);

  /// How Butter lands writes; defaults to DIRECT_COMMIT when unspecified.
  /// @gotags: json:"write_mode,omitempty" yaml:"write_mode,omitempty" bson:"write_mode,omitempty"
  @$pb.TagNumber(5)
  RepoBindingWriteMode get writeMode => $_getN(4);
  @$pb.TagNumber(5)
  set writeMode(RepoBindingWriteMode value) => $_setField(5, value);
  @$pb.TagNumber(5)
  $core.bool hasWriteMode() => $_has(4);
  @$pb.TagNumber(5)
  void clearWriteMode() => $_clearField(5);

  /// Repository layout version for Agent Content. 1 is the only version today.
  /// @gotags: json:"content_schema_version,omitempty" yaml:"content_schema_version,omitempty" bson:"content_schema_version,omitempty"
  @$pb.TagNumber(6)
  $core.int get contentSchemaVersion => $_getIZ(5);
  @$pb.TagNumber(6)
  set contentSchemaVersion($core.int value) => $_setSignedInt32(5, value);
  @$pb.TagNumber(6)
  $core.bool hasContentSchemaVersion() => $_has(5);
  @$pb.TagNumber(6)
  void clearContentSchemaVersion() => $_clearField(6);

  /// Whether an encrypted PAT is stored for this binding. Server-owned.
  /// @gotags: json:"credential_set,omitempty" yaml:"credential_set,omitempty" bson:"credential_set,omitempty"
  @$pb.TagNumber(7)
  $core.bool get credentialSet => $_getBF(6);
  @$pb.TagNumber(7)
  set credentialSet($core.bool value) => $_setBool(6, value);
  @$pb.TagNumber(7)
  $core.bool hasCredentialSet() => $_has(6);
  @$pb.TagNumber(7)
  void clearCredentialSet() => $_clearField(7);

  /// When the PAT was last set or replaced. Server-owned.
  /// @gotags: json:"credential_updated_at,omitempty" yaml:"credential_updated_at,omitempty" bson:"credential_updated_at,omitempty"
  @$pb.TagNumber(8)
  $0.Timestamp get credentialUpdatedAt => $_getN(7);
  @$pb.TagNumber(8)
  set credentialUpdatedAt($0.Timestamp value) => $_setField(8, value);
  @$pb.TagNumber(8)
  $core.bool hasCredentialUpdatedAt() => $_has(7);
  @$pb.TagNumber(8)
  void clearCredentialUpdatedAt() => $_clearField(8);
  @$pb.TagNumber(8)
  $0.Timestamp ensureCredentialUpdatedAt() => $_ensure(7);

  /// Result of the last validation run. Server-owned.
  /// @gotags: json:"status,omitempty" yaml:"status,omitempty" bson:"status,omitempty"
  @$pb.TagNumber(9)
  RepoBindingStatus get status => $_getN(8);
  @$pb.TagNumber(9)
  set status(RepoBindingStatus value) => $_setField(9, value);
  @$pb.TagNumber(9)
  $core.bool hasStatus() => $_has(8);
  @$pb.TagNumber(9)
  void clearStatus() => $_clearField(9);
  @$pb.TagNumber(9)
  RepoBindingStatus ensureStatus() => $_ensure(8);

  /// @gotags: json:"created_at,omitempty" yaml:"created_at,omitempty" bson:"created_at,omitempty"
  @$pb.TagNumber(10)
  $0.Timestamp get createdAt => $_getN(9);
  @$pb.TagNumber(10)
  set createdAt($0.Timestamp value) => $_setField(10, value);
  @$pb.TagNumber(10)
  $core.bool hasCreatedAt() => $_has(9);
  @$pb.TagNumber(10)
  void clearCreatedAt() => $_clearField(10);
  @$pb.TagNumber(10)
  $0.Timestamp ensureCreatedAt() => $_ensure(9);

  /// @gotags: json:"updated_at,omitempty" yaml:"updated_at,omitempty" bson:"updated_at,omitempty"
  @$pb.TagNumber(11)
  $0.Timestamp get updatedAt => $_getN(10);
  @$pb.TagNumber(11)
  set updatedAt($0.Timestamp value) => $_setField(11, value);
  @$pb.TagNumber(11)
  $core.bool hasUpdatedAt() => $_has(10);
  @$pb.TagNumber(11)
  void clearUpdatedAt() => $_clearField(11);
  @$pb.TagNumber(11)
  $0.Timestamp ensureUpdatedAt() => $_ensure(10);

  /// SHA of the most recently observed remote HEAD (set during sync).
  /// @gotags: json:"observed_commit_sha,omitempty" yaml:"observed_commit_sha,omitempty" bson:"observed_commit_sha,omitempty"
  @$pb.TagNumber(12)
  $core.String get observedCommitSha => $_getSZ(11);
  @$pb.TagNumber(12)
  set observedCommitSha($core.String value) => $_setString(11, value);
  @$pb.TagNumber(12)
  $core.bool hasObservedCommitSha() => $_has(11);
  @$pb.TagNumber(12)
  void clearObservedCommitSha() => $_clearField(12);

  /// SHA of the last revision that passed full validation and was activated.
  /// @gotags: json:"active_commit_sha,omitempty" yaml:"active_commit_sha,omitempty" bson:"active_commit_sha,omitempty"
  @$pb.TagNumber(13)
  $core.String get activeCommitSha => $_getSZ(12);
  @$pb.TagNumber(13)
  set activeCommitSha($core.String value) => $_setString(12, value);
  @$pb.TagNumber(13)
  $core.bool hasActiveCommitSha() => $_has(12);
  @$pb.TagNumber(13)
  void clearActiveCommitSha() => $_clearField(13);

  /// When the last successful sync completed.
  /// @gotags: json:"last_synced_at,omitempty" yaml:"last_synced_at,omitempty" bson:"last_synced_at,omitempty"
  @$pb.TagNumber(14)
  $0.Timestamp get lastSyncedAt => $_getN(13);
  @$pb.TagNumber(14)
  set lastSyncedAt($0.Timestamp value) => $_setField(14, value);
  @$pb.TagNumber(14)
  $core.bool hasLastSyncedAt() => $_has(13);
  @$pb.TagNumber(14)
  void clearLastSyncedAt() => $_clearField(14);
  @$pb.TagNumber(14)
  $0.Timestamp ensureLastSyncedAt() => $_ensure(13);

  /// Sanitized error from the most recent sync attempt. Empty on success.
  /// @gotags: json:"last_sync_error,omitempty" yaml:"last_sync_error,omitempty" bson:"last_sync_error,omitempty"
  @$pb.TagNumber(15)
  $core.String get lastSyncError => $_getSZ(14);
  @$pb.TagNumber(15)
  set lastSyncError($core.String value) => $_setString(14, value);
  @$pb.TagNumber(15)
  $core.bool hasLastSyncError() => $_has(14);
  @$pb.TagNumber(15)
  void clearLastSyncError() => $_clearField(15);

  /// Sanitized error from the most recent publication attempt. Empty on
  /// success. Per-agent validation errors appear in publication_errors.
  /// @gotags: json:"last_publication_error,omitempty" yaml:"last_publication_error,omitempty" bson:"last_publication_error,omitempty"
  @$pb.TagNumber(16)
  $core.String get lastPublicationError => $_getSZ(15);
  @$pb.TagNumber(16)
  set lastPublicationError($core.String value) => $_setString(15, value);
  @$pb.TagNumber(16)
  $core.bool hasLastPublicationError() => $_has(15);
  @$pb.TagNumber(16)
  void clearLastPublicationError() => $_clearField(16);

  /// Per-agent validation errors from the most recent publication attempt.
  /// @gotags: json:"publication_errors,omitempty" yaml:"publication_errors,omitempty" bson:"publication_errors,omitempty"
  @$pb.TagNumber(17)
  $pb.PbList<$core.String> get publicationErrors => $_getList(16);

  /// Whether a webhook secret is stored for this binding. Server-owned.
  /// @gotags: json:"webhook_secret_set,omitempty" yaml:"webhook_secret_set,omitempty" bson:"webhook_secret_set,omitempty"
  @$pb.TagNumber(18)
  $core.bool get webhookSecretSet => $_getBF(17);
  @$pb.TagNumber(18)
  set webhookSecretSet($core.bool value) => $_setBool(17, value);
  @$pb.TagNumber(18)
  $core.bool hasWebhookSecretSet() => $_has(17);
  @$pb.TagNumber(18)
  void clearWebhookSecretSet() => $_clearField(18);

  /// @gotags: json:"workspace_id,omitempty" yaml:"workspace_id,omitempty" bson:"workspace_id,omitempty"
  @$pb.TagNumber(100)
  $core.String get workspaceId => $_getSZ(18);
  @$pb.TagNumber(100)
  set workspaceId($core.String value) => $_setString(18, value);
  @$pb.TagNumber(100)
  $core.bool hasWorkspaceId() => $_has(18);
  @$pb.TagNumber(100)
  void clearWorkspaceId() => $_clearField(100);
}

/// RepoBindingStatus is the persisted, credential-free result of the most
/// recent validation run.
class RepoBindingStatus extends $pb.GeneratedMessage {
  factory RepoBindingStatus({
    RepoBindingConnectionState? state,
    $0.Timestamp? lastValidatedAt,
    $core.String? error,
    $core.Iterable<RepoBindingCheck>? checks,
  }) {
    final result = RepoBindingStatus._();
    if (state != null) result.state = state;
    if (lastValidatedAt != null) result.lastValidatedAt = lastValidatedAt;
    if (error != null) result.error = error;
    if (checks != null) result.checks.addAll(checks);
    return result;
  }

  RepoBindingStatus._();

  factory RepoBindingStatus.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      RepoBindingStatus()..mergeFromBuffer(data, registry);
  factory RepoBindingStatus.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      RepoBindingStatus()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'RepoBindingStatus',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: RepoBindingStatus.$_createMessage)
    ..aE<RepoBindingConnectionState>(1, _omitFieldNames ? '' : 'state',
        enumValues: RepoBindingConnectionState.values)
    ..aOM<$0.Timestamp>(2, _omitFieldNames ? '' : 'lastValidatedAt',
        subBuilder: $0.Timestamp.$_createMessage)
    ..aOS(3, _omitFieldNames ? '' : 'error')
    ..pPM<RepoBindingCheck>(4, _omitFieldNames ? '' : 'checks',
        subBuilder: RepoBindingCheck.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  RepoBindingStatus clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  RepoBindingStatus copyWith(void Function(RepoBindingStatus) updates) =>
      super.copyWith((message) => updates(message as RepoBindingStatus))
          as RepoBindingStatus;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use RepoBindingStatus() / RepoBindingStatus.new instead')
  static RepoBindingStatus create() => RepoBindingStatus._();
  static $pb.GeneratedMessage $_createMessage() => RepoBindingStatus._();
  @$core.override
  RepoBindingStatus createEmptyInstance() => RepoBindingStatus._();
  @$core.pragma('dart2js:noInline')
  static RepoBindingStatus getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<RepoBindingStatus>(
          RepoBindingStatus.$_createMessage);
  static RepoBindingStatus? _defaultInstance;

  /// @gotags: json:"state,omitempty" yaml:"state,omitempty" bson:"state,omitempty"
  @$pb.TagNumber(1)
  RepoBindingConnectionState get state => $_getN(0);
  @$pb.TagNumber(1)
  set state(RepoBindingConnectionState value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasState() => $_has(0);
  @$pb.TagNumber(1)
  void clearState() => $_clearField(1);

  /// @gotags: json:"last_validated_at,omitempty" yaml:"last_validated_at,omitempty" bson:"last_validated_at,omitempty"
  @$pb.TagNumber(2)
  $0.Timestamp get lastValidatedAt => $_getN(1);
  @$pb.TagNumber(2)
  set lastValidatedAt($0.Timestamp value) => $_setField(2, value);
  @$pb.TagNumber(2)
  $core.bool hasLastValidatedAt() => $_has(1);
  @$pb.TagNumber(2)
  void clearLastValidatedAt() => $_clearField(2);
  @$pb.TagNumber(2)
  $0.Timestamp ensureLastValidatedAt() => $_ensure(1);

  /// Sanitized failure summary. Never contains credential material.
  /// @gotags: json:"error,omitempty" yaml:"error,omitempty" bson:"error,omitempty"
  @$pb.TagNumber(3)
  $core.String get error => $_getSZ(2);
  @$pb.TagNumber(3)
  set error($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasError() => $_has(2);
  @$pb.TagNumber(3)
  void clearError() => $_clearField(3);

  /// Individual capability checks from the last run.
  /// @gotags: json:"checks,omitempty" yaml:"checks,omitempty" bson:"checks,omitempty"
  @$pb.TagNumber(4)
  $pb.PbList<RepoBindingCheck> get checks => $_getList(3);
}

/// RepoBindingCheck is one capability probe from a validation run. Check
/// names: "repository_read", "branch_exists", "write_capability",
/// "change_request_capability".
class RepoBindingCheck extends $pb.GeneratedMessage {
  factory RepoBindingCheck({
    $core.String? name,
    $core.bool? ok,
    $core.bool? required,
    $core.String? detail,
  }) {
    final result = RepoBindingCheck._();
    if (name != null) result.name = name;
    if (ok != null) result.ok = ok;
    if (required != null) result.required = required;
    if (detail != null) result.detail = detail;
    return result;
  }

  RepoBindingCheck._();

  factory RepoBindingCheck.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      RepoBindingCheck()..mergeFromBuffer(data, registry);
  factory RepoBindingCheck.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      RepoBindingCheck()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'RepoBindingCheck',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: RepoBindingCheck.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'name')
    ..aOB(2, _omitFieldNames ? '' : 'ok')
    ..aOB(3, _omitFieldNames ? '' : 'required')
    ..aOS(4, _omitFieldNames ? '' : 'detail')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  RepoBindingCheck clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  RepoBindingCheck copyWith(void Function(RepoBindingCheck) updates) =>
      super.copyWith((message) => updates(message as RepoBindingCheck))
          as RepoBindingCheck;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use RepoBindingCheck() / RepoBindingCheck.new instead')
  static RepoBindingCheck create() => RepoBindingCheck._();
  static $pb.GeneratedMessage $_createMessage() => RepoBindingCheck._();
  @$core.override
  RepoBindingCheck createEmptyInstance() => RepoBindingCheck._();
  @$core.pragma('dart2js:noInline')
  static RepoBindingCheck getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<RepoBindingCheck>(
          RepoBindingCheck.$_createMessage);
  static RepoBindingCheck? _defaultInstance;

  /// @gotags: json:"name,omitempty" yaml:"name,omitempty" bson:"name,omitempty"
  @$pb.TagNumber(1)
  $core.String get name => $_getSZ(0);
  @$pb.TagNumber(1)
  set name($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasName() => $_has(0);
  @$pb.TagNumber(1)
  void clearName() => $_clearField(1);

  /// @gotags: json:"ok,omitempty" yaml:"ok,omitempty" bson:"ok,omitempty"
  @$pb.TagNumber(2)
  $core.bool get ok => $_getBF(1);
  @$pb.TagNumber(2)
  set ok($core.bool value) => $_setBool(1, value);
  @$pb.TagNumber(2)
  $core.bool hasOk() => $_has(1);
  @$pb.TagNumber(2)
  void clearOk() => $_clearField(2);

  /// Whether this check gates the OK state for the selected write mode.
  /// @gotags: json:"required,omitempty" yaml:"required,omitempty" bson:"required,omitempty"
  @$pb.TagNumber(3)
  $core.bool get required => $_getBF(2);
  @$pb.TagNumber(3)
  set required($core.bool value) => $_setBool(2, value);
  @$pb.TagNumber(3)
  $core.bool hasRequired() => $_has(2);
  @$pb.TagNumber(3)
  void clearRequired() => $_clearField(3);

  /// Human-readable, credential-free explanation.
  /// @gotags: json:"detail,omitempty" yaml:"detail,omitempty" bson:"detail,omitempty"
  @$pb.TagNumber(4)
  $core.String get detail => $_getSZ(3);
  @$pb.TagNumber(4)
  set detail($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasDetail() => $_has(3);
  @$pb.TagNumber(4)
  void clearDetail() => $_clearField(4);
}

/// RepoBindingOverlap identifies another workspace bound to the same effective
/// repository location (same host, repository, branch, and root path).
class RepoBindingOverlap extends $pb.GeneratedMessage {
  factory RepoBindingOverlap({
    $core.String? workspaceId,
    $core.String? workspaceName,
  }) {
    final result = RepoBindingOverlap._();
    if (workspaceId != null) result.workspaceId = workspaceId;
    if (workspaceName != null) result.workspaceName = workspaceName;
    return result;
  }

  RepoBindingOverlap._();

  factory RepoBindingOverlap.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      RepoBindingOverlap()..mergeFromBuffer(data, registry);
  factory RepoBindingOverlap.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      RepoBindingOverlap()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'RepoBindingOverlap',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: RepoBindingOverlap.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'workspaceId')
    ..aOS(2, _omitFieldNames ? '' : 'workspaceName')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  RepoBindingOverlap clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  RepoBindingOverlap copyWith(void Function(RepoBindingOverlap) updates) =>
      super.copyWith((message) => updates(message as RepoBindingOverlap))
          as RepoBindingOverlap;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use RepoBindingOverlap() / RepoBindingOverlap.new instead')
  static RepoBindingOverlap create() => RepoBindingOverlap._();
  static $pb.GeneratedMessage $_createMessage() => RepoBindingOverlap._();
  @$core.override
  RepoBindingOverlap createEmptyInstance() => RepoBindingOverlap._();
  @$core.pragma('dart2js:noInline')
  static RepoBindingOverlap getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<RepoBindingOverlap>(
          RepoBindingOverlap.$_createMessage);
  static RepoBindingOverlap? _defaultInstance;

  /// @gotags: json:"workspace_id,omitempty" yaml:"workspace_id,omitempty" bson:"workspace_id,omitempty"
  @$pb.TagNumber(1)
  $core.String get workspaceId => $_getSZ(0);
  @$pb.TagNumber(1)
  set workspaceId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasWorkspaceId() => $_has(0);
  @$pb.TagNumber(1)
  void clearWorkspaceId() => $_clearField(1);

  /// @gotags: json:"workspace_name,omitempty" yaml:"workspace_name,omitempty" bson:"workspace_name,omitempty"
  @$pb.TagNumber(2)
  $core.String get workspaceName => $_getSZ(1);
  @$pb.TagNumber(2)
  set workspaceName($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasWorkspaceName() => $_has(1);
  @$pb.TagNumber(2)
  void clearWorkspaceName() => $_clearField(2);
}

class GetWorkspaceRepoBindingRequest extends $pb.GeneratedMessage {
  factory GetWorkspaceRepoBindingRequest() =>
      GetWorkspaceRepoBindingRequest._();

  GetWorkspaceRepoBindingRequest._();

  factory GetWorkspaceRepoBindingRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetWorkspaceRepoBindingRequest()..mergeFromBuffer(data, registry);
  factory GetWorkspaceRepoBindingRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetWorkspaceRepoBindingRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetWorkspaceRepoBindingRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: GetWorkspaceRepoBindingRequest.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetWorkspaceRepoBindingRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetWorkspaceRepoBindingRequest copyWith(
          void Function(GetWorkspaceRepoBindingRequest) updates) =>
      super.copyWith(
              (message) => updates(message as GetWorkspaceRepoBindingRequest))
          as GetWorkspaceRepoBindingRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use GetWorkspaceRepoBindingRequest() / GetWorkspaceRepoBindingRequest.new instead')
  static GetWorkspaceRepoBindingRequest create() =>
      GetWorkspaceRepoBindingRequest._();
  static $pb.GeneratedMessage $_createMessage() =>
      GetWorkspaceRepoBindingRequest._();
  @$core.override
  GetWorkspaceRepoBindingRequest createEmptyInstance() =>
      GetWorkspaceRepoBindingRequest._();
  @$core.pragma('dart2js:noInline')
  static GetWorkspaceRepoBindingRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<GetWorkspaceRepoBindingRequest>(
          GetWorkspaceRepoBindingRequest.$_createMessage);
  static GetWorkspaceRepoBindingRequest? _defaultInstance;
}

class GetWorkspaceRepoBindingResponse extends $pb.GeneratedMessage {
  factory GetWorkspaceRepoBindingResponse({
    WorkspaceRepoBinding? binding,
    $core.Iterable<RepoBindingOverlap>? overlaps,
  }) {
    final result = GetWorkspaceRepoBindingResponse._();
    if (binding != null) result.binding = binding;
    if (overlaps != null) result.overlaps.addAll(overlaps);
    return result;
  }

  GetWorkspaceRepoBindingResponse._();

  factory GetWorkspaceRepoBindingResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetWorkspaceRepoBindingResponse()..mergeFromBuffer(data, registry);
  factory GetWorkspaceRepoBindingResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetWorkspaceRepoBindingResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetWorkspaceRepoBindingResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: GetWorkspaceRepoBindingResponse.$_createMessage)
    ..aOM<WorkspaceRepoBinding>(1, _omitFieldNames ? '' : 'binding',
        subBuilder: WorkspaceRepoBinding.$_createMessage)
    ..pPM<RepoBindingOverlap>(2, _omitFieldNames ? '' : 'overlaps',
        subBuilder: RepoBindingOverlap.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetWorkspaceRepoBindingResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetWorkspaceRepoBindingResponse copyWith(
          void Function(GetWorkspaceRepoBindingResponse) updates) =>
      super.copyWith(
              (message) => updates(message as GetWorkspaceRepoBindingResponse))
          as GetWorkspaceRepoBindingResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use GetWorkspaceRepoBindingResponse() / GetWorkspaceRepoBindingResponse.new instead')
  static GetWorkspaceRepoBindingResponse create() =>
      GetWorkspaceRepoBindingResponse._();
  static $pb.GeneratedMessage $_createMessage() =>
      GetWorkspaceRepoBindingResponse._();
  @$core.override
  GetWorkspaceRepoBindingResponse createEmptyInstance() =>
      GetWorkspaceRepoBindingResponse._();
  @$core.pragma('dart2js:noInline')
  static GetWorkspaceRepoBindingResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<GetWorkspaceRepoBindingResponse>(
          GetWorkspaceRepoBindingResponse.$_createMessage);
  static GetWorkspaceRepoBindingResponse? _defaultInstance;

  /// Unset when the workspace has no binding.
  @$pb.TagNumber(1)
  WorkspaceRepoBinding get binding => $_getN(0);
  @$pb.TagNumber(1)
  set binding(WorkspaceRepoBinding value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasBinding() => $_has(0);
  @$pb.TagNumber(1)
  void clearBinding() => $_clearField(1);
  @$pb.TagNumber(1)
  WorkspaceRepoBinding ensureBinding() => $_ensure(0);

  @$pb.TagNumber(2)
  $pb.PbList<RepoBindingOverlap> get overlaps => $_getList(1);
}

class PutWorkspaceRepoBindingRequest extends $pb.GeneratedMessage {
  factory PutWorkspaceRepoBindingRequest({
    WorkspaceRepoBinding? binding,
  }) {
    final result = PutWorkspaceRepoBindingRequest._();
    if (binding != null) result.binding = binding;
    return result;
  }

  PutWorkspaceRepoBindingRequest._();

  factory PutWorkspaceRepoBindingRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      PutWorkspaceRepoBindingRequest()..mergeFromBuffer(data, registry);
  factory PutWorkspaceRepoBindingRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      PutWorkspaceRepoBindingRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'PutWorkspaceRepoBindingRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: PutWorkspaceRepoBindingRequest.$_createMessage)
    ..aOM<WorkspaceRepoBinding>(1, _omitFieldNames ? '' : 'binding',
        subBuilder: WorkspaceRepoBinding.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  PutWorkspaceRepoBindingRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  PutWorkspaceRepoBindingRequest copyWith(
          void Function(PutWorkspaceRepoBindingRequest) updates) =>
      super.copyWith(
              (message) => updates(message as PutWorkspaceRepoBindingRequest))
          as PutWorkspaceRepoBindingRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use PutWorkspaceRepoBindingRequest() / PutWorkspaceRepoBindingRequest.new instead')
  static PutWorkspaceRepoBindingRequest create() =>
      PutWorkspaceRepoBindingRequest._();
  static $pb.GeneratedMessage $_createMessage() =>
      PutWorkspaceRepoBindingRequest._();
  @$core.override
  PutWorkspaceRepoBindingRequest createEmptyInstance() =>
      PutWorkspaceRepoBindingRequest._();
  @$core.pragma('dart2js:noInline')
  static PutWorkspaceRepoBindingRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<PutWorkspaceRepoBindingRequest>(
          PutWorkspaceRepoBindingRequest.$_createMessage);
  static PutWorkspaceRepoBindingRequest? _defaultInstance;

  @$pb.TagNumber(1)
  WorkspaceRepoBinding get binding => $_getN(0);
  @$pb.TagNumber(1)
  set binding(WorkspaceRepoBinding value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasBinding() => $_has(0);
  @$pb.TagNumber(1)
  void clearBinding() => $_clearField(1);
  @$pb.TagNumber(1)
  WorkspaceRepoBinding ensureBinding() => $_ensure(0);
}

class PutWorkspaceRepoBindingResponse extends $pb.GeneratedMessage {
  factory PutWorkspaceRepoBindingResponse({
    WorkspaceRepoBinding? binding,
  }) {
    final result = PutWorkspaceRepoBindingResponse._();
    if (binding != null) result.binding = binding;
    return result;
  }

  PutWorkspaceRepoBindingResponse._();

  factory PutWorkspaceRepoBindingResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      PutWorkspaceRepoBindingResponse()..mergeFromBuffer(data, registry);
  factory PutWorkspaceRepoBindingResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      PutWorkspaceRepoBindingResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'PutWorkspaceRepoBindingResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: PutWorkspaceRepoBindingResponse.$_createMessage)
    ..aOM<WorkspaceRepoBinding>(1, _omitFieldNames ? '' : 'binding',
        subBuilder: WorkspaceRepoBinding.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  PutWorkspaceRepoBindingResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  PutWorkspaceRepoBindingResponse copyWith(
          void Function(PutWorkspaceRepoBindingResponse) updates) =>
      super.copyWith(
              (message) => updates(message as PutWorkspaceRepoBindingResponse))
          as PutWorkspaceRepoBindingResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use PutWorkspaceRepoBindingResponse() / PutWorkspaceRepoBindingResponse.new instead')
  static PutWorkspaceRepoBindingResponse create() =>
      PutWorkspaceRepoBindingResponse._();
  static $pb.GeneratedMessage $_createMessage() =>
      PutWorkspaceRepoBindingResponse._();
  @$core.override
  PutWorkspaceRepoBindingResponse createEmptyInstance() =>
      PutWorkspaceRepoBindingResponse._();
  @$core.pragma('dart2js:noInline')
  static PutWorkspaceRepoBindingResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<PutWorkspaceRepoBindingResponse>(
          PutWorkspaceRepoBindingResponse.$_createMessage);
  static PutWorkspaceRepoBindingResponse? _defaultInstance;

  @$pb.TagNumber(1)
  WorkspaceRepoBinding get binding => $_getN(0);
  @$pb.TagNumber(1)
  set binding(WorkspaceRepoBinding value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasBinding() => $_has(0);
  @$pb.TagNumber(1)
  void clearBinding() => $_clearField(1);
  @$pb.TagNumber(1)
  WorkspaceRepoBinding ensureBinding() => $_ensure(0);
}

class DeleteWorkspaceRepoBindingRequest extends $pb.GeneratedMessage {
  factory DeleteWorkspaceRepoBindingRequest({
    RepoBindingDetachRecovery? recovery,
  }) {
    final result = DeleteWorkspaceRepoBindingRequest._();
    if (recovery != null) result.recovery = recovery;
    return result;
  }

  DeleteWorkspaceRepoBindingRequest._();

  factory DeleteWorkspaceRepoBindingRequest.fromBuffer(
          $core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DeleteWorkspaceRepoBindingRequest()..mergeFromBuffer(data, registry);
  factory DeleteWorkspaceRepoBindingRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DeleteWorkspaceRepoBindingRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'DeleteWorkspaceRepoBindingRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: DeleteWorkspaceRepoBindingRequest.$_createMessage)
    ..aE<RepoBindingDetachRecovery>(1, _omitFieldNames ? '' : 'recovery',
        enumValues: RepoBindingDetachRecovery.values)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteWorkspaceRepoBindingRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteWorkspaceRepoBindingRequest copyWith(
          void Function(DeleteWorkspaceRepoBindingRequest) updates) =>
      super.copyWith((message) =>
              updates(message as DeleteWorkspaceRepoBindingRequest))
          as DeleteWorkspaceRepoBindingRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use DeleteWorkspaceRepoBindingRequest() / DeleteWorkspaceRepoBindingRequest.new instead')
  static DeleteWorkspaceRepoBindingRequest create() =>
      DeleteWorkspaceRepoBindingRequest._();
  static $pb.GeneratedMessage $_createMessage() =>
      DeleteWorkspaceRepoBindingRequest._();
  @$core.override
  DeleteWorkspaceRepoBindingRequest createEmptyInstance() =>
      DeleteWorkspaceRepoBindingRequest._();
  @$core.pragma('dart2js:noInline')
  static DeleteWorkspaceRepoBindingRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<DeleteWorkspaceRepoBindingRequest>(
          DeleteWorkspaceRepoBindingRequest.$_createMessage);
  static DeleteWorkspaceRepoBindingRequest? _defaultInstance;

  /// Recovery path selected when no valid Active Revision snapshot exists to
  /// materialize. Default (UNSPECIFIED) refuses the unbind in that case.
  @$pb.TagNumber(1)
  RepoBindingDetachRecovery get recovery => $_getN(0);
  @$pb.TagNumber(1)
  set recovery(RepoBindingDetachRecovery value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasRecovery() => $_has(0);
  @$pb.TagNumber(1)
  void clearRecovery() => $_clearField(1);
}

class DeleteWorkspaceRepoBindingResponse extends $pb.GeneratedMessage {
  factory DeleteWorkspaceRepoBindingResponse({
    $core.int? agentsMaterialized,
  }) {
    final result = DeleteWorkspaceRepoBindingResponse._();
    if (agentsMaterialized != null)
      result.agentsMaterialized = agentsMaterialized;
    return result;
  }

  DeleteWorkspaceRepoBindingResponse._();

  factory DeleteWorkspaceRepoBindingResponse.fromBuffer(
          $core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DeleteWorkspaceRepoBindingResponse()..mergeFromBuffer(data, registry);
  factory DeleteWorkspaceRepoBindingResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DeleteWorkspaceRepoBindingResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'DeleteWorkspaceRepoBindingResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: DeleteWorkspaceRepoBindingResponse.$_createMessage)
    ..aI(1, _omitFieldNames ? '' : 'agentsMaterialized')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteWorkspaceRepoBindingResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteWorkspaceRepoBindingResponse copyWith(
          void Function(DeleteWorkspaceRepoBindingResponse) updates) =>
      super.copyWith((message) =>
              updates(message as DeleteWorkspaceRepoBindingResponse))
          as DeleteWorkspaceRepoBindingResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use DeleteWorkspaceRepoBindingResponse() / DeleteWorkspaceRepoBindingResponse.new instead')
  static DeleteWorkspaceRepoBindingResponse create() =>
      DeleteWorkspaceRepoBindingResponse._();
  static $pb.GeneratedMessage $_createMessage() =>
      DeleteWorkspaceRepoBindingResponse._();
  @$core.override
  DeleteWorkspaceRepoBindingResponse createEmptyInstance() =>
      DeleteWorkspaceRepoBindingResponse._();
  @$core.pragma('dart2js:noInline')
  static DeleteWorkspaceRepoBindingResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<DeleteWorkspaceRepoBindingResponse>(
          DeleteWorkspaceRepoBindingResponse.$_createMessage);
  static DeleteWorkspaceRepoBindingResponse? _defaultInstance;

  /// Number of agents whose Git-owned content was materialized back into
  /// database-managed fields before the binding was removed.
  @$pb.TagNumber(1)
  $core.int get agentsMaterialized => $_getIZ(0);
  @$pb.TagNumber(1)
  set agentsMaterialized($core.int value) => $_setSignedInt32(0, value);
  @$pb.TagNumber(1)
  $core.bool hasAgentsMaterialized() => $_has(0);
  @$pb.TagNumber(1)
  void clearAgentsMaterialized() => $_clearField(1);
}

class OnboardWorkspaceRepositoryRequest extends $pb.GeneratedMessage {
  factory OnboardWorkspaceRepositoryRequest({
    RepoBindingOnboardingMode? mode,
  }) {
    final result = OnboardWorkspaceRepositoryRequest._();
    if (mode != null) result.mode = mode;
    return result;
  }

  OnboardWorkspaceRepositoryRequest._();

  factory OnboardWorkspaceRepositoryRequest.fromBuffer(
          $core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      OnboardWorkspaceRepositoryRequest()..mergeFromBuffer(data, registry);
  factory OnboardWorkspaceRepositoryRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      OnboardWorkspaceRepositoryRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'OnboardWorkspaceRepositoryRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: OnboardWorkspaceRepositoryRequest.$_createMessage)
    ..aE<RepoBindingOnboardingMode>(1, _omitFieldNames ? '' : 'mode',
        enumValues: RepoBindingOnboardingMode.values)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  OnboardWorkspaceRepositoryRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  OnboardWorkspaceRepositoryRequest copyWith(
          void Function(OnboardWorkspaceRepositoryRequest) updates) =>
      super.copyWith((message) =>
              updates(message as OnboardWorkspaceRepositoryRequest))
          as OnboardWorkspaceRepositoryRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use OnboardWorkspaceRepositoryRequest() / OnboardWorkspaceRepositoryRequest.new instead')
  static OnboardWorkspaceRepositoryRequest create() =>
      OnboardWorkspaceRepositoryRequest._();
  static $pb.GeneratedMessage $_createMessage() =>
      OnboardWorkspaceRepositoryRequest._();
  @$core.override
  OnboardWorkspaceRepositoryRequest createEmptyInstance() =>
      OnboardWorkspaceRepositoryRequest._();
  @$core.pragma('dart2js:noInline')
  static OnboardWorkspaceRepositoryRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<OnboardWorkspaceRepositoryRequest>(
          OnboardWorkspaceRepositoryRequest.$_createMessage);
  static OnboardWorkspaceRepositoryRequest? _defaultInstance;

  @$pb.TagNumber(1)
  RepoBindingOnboardingMode get mode => $_getN(0);
  @$pb.TagNumber(1)
  set mode(RepoBindingOnboardingMode value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasMode() => $_has(0);
  @$pb.TagNumber(1)
  void clearMode() => $_clearField(1);
}

class OnboardWorkspaceRepositoryResponse extends $pb.GeneratedMessage {
  factory OnboardWorkspaceRepositoryResponse({
    WorkspaceRepoBinding? binding,
    $core.String? commitSha,
    $core.bool? published,
    $core.Iterable<$core.String>? validationErrors,
    $core.int? agentsExported,
    $core.int? agentsImported,
  }) {
    final result = OnboardWorkspaceRepositoryResponse._();
    if (binding != null) result.binding = binding;
    if (commitSha != null) result.commitSha = commitSha;
    if (published != null) result.published = published;
    if (validationErrors != null)
      result.validationErrors.addAll(validationErrors);
    if (agentsExported != null) result.agentsExported = agentsExported;
    if (agentsImported != null) result.agentsImported = agentsImported;
    return result;
  }

  OnboardWorkspaceRepositoryResponse._();

  factory OnboardWorkspaceRepositoryResponse.fromBuffer(
          $core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      OnboardWorkspaceRepositoryResponse()..mergeFromBuffer(data, registry);
  factory OnboardWorkspaceRepositoryResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      OnboardWorkspaceRepositoryResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'OnboardWorkspaceRepositoryResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: OnboardWorkspaceRepositoryResponse.$_createMessage)
    ..aOM<WorkspaceRepoBinding>(1, _omitFieldNames ? '' : 'binding',
        subBuilder: WorkspaceRepoBinding.$_createMessage)
    ..aOS(2, _omitFieldNames ? '' : 'commitSha')
    ..aOB(3, _omitFieldNames ? '' : 'published')
    ..pPS(4, _omitFieldNames ? '' : 'validationErrors')
    ..aI(5, _omitFieldNames ? '' : 'agentsExported')
    ..aI(6, _omitFieldNames ? '' : 'agentsImported')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  OnboardWorkspaceRepositoryResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  OnboardWorkspaceRepositoryResponse copyWith(
          void Function(OnboardWorkspaceRepositoryResponse) updates) =>
      super.copyWith((message) =>
              updates(message as OnboardWorkspaceRepositoryResponse))
          as OnboardWorkspaceRepositoryResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use OnboardWorkspaceRepositoryResponse() / OnboardWorkspaceRepositoryResponse.new instead')
  static OnboardWorkspaceRepositoryResponse create() =>
      OnboardWorkspaceRepositoryResponse._();
  static $pb.GeneratedMessage $_createMessage() =>
      OnboardWorkspaceRepositoryResponse._();
  @$core.override
  OnboardWorkspaceRepositoryResponse createEmptyInstance() =>
      OnboardWorkspaceRepositoryResponse._();
  @$core.pragma('dart2js:noInline')
  static OnboardWorkspaceRepositoryResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<OnboardWorkspaceRepositoryResponse>(
          OnboardWorkspaceRepositoryResponse.$_createMessage);
  static OnboardWorkspaceRepositoryResponse? _defaultInstance;

  @$pb.TagNumber(1)
  WorkspaceRepoBinding get binding => $_getN(0);
  @$pb.TagNumber(1)
  set binding(WorkspaceRepoBinding value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasBinding() => $_has(0);
  @$pb.TagNumber(1)
  void clearBinding() => $_clearField(1);
  @$pb.TagNumber(1)
  WorkspaceRepoBinding ensureBinding() => $_ensure(0);

  /// SHA of the export commit (EXPORT_CURRENT). Empty for IMPORT_REPOSITORY.
  @$pb.TagNumber(2)
  $core.String get commitSha => $_getSZ(1);
  @$pb.TagNumber(2)
  set commitSha($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasCommitSha() => $_has(1);
  @$pb.TagNumber(2)
  void clearCommitSha() => $_clearField(2);

  /// Whether the Active Revision was published as a result of onboarding.
  @$pb.TagNumber(3)
  $core.bool get published => $_getBF(2);
  @$pb.TagNumber(3)
  set published($core.bool value) => $_setBool(2, value);
  @$pb.TagNumber(3)
  $core.bool hasPublished() => $_has(2);
  @$pb.TagNumber(3)
  void clearPublished() => $_clearField(3);

  /// Per-agent validation errors that prevented onboarding (empty on success).
  @$pb.TagNumber(4)
  $pb.PbList<$core.String> get validationErrors => $_getList(3);

  /// Number of agents whose content was exported to the repository
  /// (EXPORT_CURRENT).
  @$pb.TagNumber(5)
  $core.int get agentsExported => $_getIZ(4);
  @$pb.TagNumber(5)
  set agentsExported($core.int value) => $_setSignedInt32(4, value);
  @$pb.TagNumber(5)
  $core.bool hasAgentsExported() => $_has(4);
  @$pb.TagNumber(5)
  void clearAgentsExported() => $_clearField(5);

  /// Number of agents whose content was imported and activated from the
  /// repository (IMPORT_REPOSITORY).
  @$pb.TagNumber(6)
  $core.int get agentsImported => $_getIZ(5);
  @$pb.TagNumber(6)
  set agentsImported($core.int value) => $_setSignedInt32(5, value);
  @$pb.TagNumber(6)
  $core.bool hasAgentsImported() => $_has(5);
  @$pb.TagNumber(6)
  void clearAgentsImported() => $_clearField(6);
}

class SetWorkspaceRepoBindingCredentialRequest extends $pb.GeneratedMessage {
  factory SetWorkspaceRepoBindingCredentialRequest({
    $core.String? pat,
  }) {
    final result = SetWorkspaceRepoBindingCredentialRequest._();
    if (pat != null) result.pat = pat;
    return result;
  }

  SetWorkspaceRepoBindingCredentialRequest._();

  factory SetWorkspaceRepoBindingCredentialRequest.fromBuffer(
          $core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      SetWorkspaceRepoBindingCredentialRequest()
        ..mergeFromBuffer(data, registry);
  factory SetWorkspaceRepoBindingCredentialRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      SetWorkspaceRepoBindingCredentialRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'SetWorkspaceRepoBindingCredentialRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance:
          SetWorkspaceRepoBindingCredentialRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'pat')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  SetWorkspaceRepoBindingCredentialRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  SetWorkspaceRepoBindingCredentialRequest copyWith(
          void Function(SetWorkspaceRepoBindingCredentialRequest) updates) =>
      super.copyWith((message) =>
              updates(message as SetWorkspaceRepoBindingCredentialRequest))
          as SetWorkspaceRepoBindingCredentialRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use SetWorkspaceRepoBindingCredentialRequest() / SetWorkspaceRepoBindingCredentialRequest.new instead')
  static SetWorkspaceRepoBindingCredentialRequest create() =>
      SetWorkspaceRepoBindingCredentialRequest._();
  static $pb.GeneratedMessage $_createMessage() =>
      SetWorkspaceRepoBindingCredentialRequest._();
  @$core.override
  SetWorkspaceRepoBindingCredentialRequest createEmptyInstance() =>
      SetWorkspaceRepoBindingCredentialRequest._();
  @$core.pragma('dart2js:noInline')
  static SetWorkspaceRepoBindingCredentialRequest getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<
              SetWorkspaceRepoBindingCredentialRequest>(
          SetWorkspaceRepoBindingCredentialRequest.$_createMessage);
  static SetWorkspaceRepoBindingCredentialRequest? _defaultInstance;

  /// Personal access token, write-only. Encrypted at rest; never returned by
  /// any RPC and never written to logs or error details.
  @$pb.TagNumber(1)
  $core.String get pat => $_getSZ(0);
  @$pb.TagNumber(1)
  set pat($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasPat() => $_has(0);
  @$pb.TagNumber(1)
  void clearPat() => $_clearField(1);
}

class SetWorkspaceRepoBindingCredentialResponse extends $pb.GeneratedMessage {
  factory SetWorkspaceRepoBindingCredentialResponse({
    WorkspaceRepoBinding? binding,
  }) {
    final result = SetWorkspaceRepoBindingCredentialResponse._();
    if (binding != null) result.binding = binding;
    return result;
  }

  SetWorkspaceRepoBindingCredentialResponse._();

  factory SetWorkspaceRepoBindingCredentialResponse.fromBuffer(
          $core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      SetWorkspaceRepoBindingCredentialResponse()
        ..mergeFromBuffer(data, registry);
  factory SetWorkspaceRepoBindingCredentialResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      SetWorkspaceRepoBindingCredentialResponse()
        ..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'SetWorkspaceRepoBindingCredentialResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance:
          SetWorkspaceRepoBindingCredentialResponse.$_createMessage)
    ..aOM<WorkspaceRepoBinding>(1, _omitFieldNames ? '' : 'binding',
        subBuilder: WorkspaceRepoBinding.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  SetWorkspaceRepoBindingCredentialResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  SetWorkspaceRepoBindingCredentialResponse copyWith(
          void Function(SetWorkspaceRepoBindingCredentialResponse) updates) =>
      super.copyWith((message) =>
              updates(message as SetWorkspaceRepoBindingCredentialResponse))
          as SetWorkspaceRepoBindingCredentialResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use SetWorkspaceRepoBindingCredentialResponse() / SetWorkspaceRepoBindingCredentialResponse.new instead')
  static SetWorkspaceRepoBindingCredentialResponse create() =>
      SetWorkspaceRepoBindingCredentialResponse._();
  static $pb.GeneratedMessage $_createMessage() =>
      SetWorkspaceRepoBindingCredentialResponse._();
  @$core.override
  SetWorkspaceRepoBindingCredentialResponse createEmptyInstance() =>
      SetWorkspaceRepoBindingCredentialResponse._();
  @$core.pragma('dart2js:noInline')
  static SetWorkspaceRepoBindingCredentialResponse getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<
              SetWorkspaceRepoBindingCredentialResponse>(
          SetWorkspaceRepoBindingCredentialResponse.$_createMessage);
  static SetWorkspaceRepoBindingCredentialResponse? _defaultInstance;

  @$pb.TagNumber(1)
  WorkspaceRepoBinding get binding => $_getN(0);
  @$pb.TagNumber(1)
  set binding(WorkspaceRepoBinding value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasBinding() => $_has(0);
  @$pb.TagNumber(1)
  void clearBinding() => $_clearField(1);
  @$pb.TagNumber(1)
  WorkspaceRepoBinding ensureBinding() => $_ensure(0);
}

class ValidateWorkspaceRepoBindingRequest extends $pb.GeneratedMessage {
  factory ValidateWorkspaceRepoBindingRequest() =>
      ValidateWorkspaceRepoBindingRequest._();

  ValidateWorkspaceRepoBindingRequest._();

  factory ValidateWorkspaceRepoBindingRequest.fromBuffer(
          $core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ValidateWorkspaceRepoBindingRequest()..mergeFromBuffer(data, registry);
  factory ValidateWorkspaceRepoBindingRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ValidateWorkspaceRepoBindingRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ValidateWorkspaceRepoBindingRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: ValidateWorkspaceRepoBindingRequest.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ValidateWorkspaceRepoBindingRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ValidateWorkspaceRepoBindingRequest copyWith(
          void Function(ValidateWorkspaceRepoBindingRequest) updates) =>
      super.copyWith((message) =>
              updates(message as ValidateWorkspaceRepoBindingRequest))
          as ValidateWorkspaceRepoBindingRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use ValidateWorkspaceRepoBindingRequest() / ValidateWorkspaceRepoBindingRequest.new instead')
  static ValidateWorkspaceRepoBindingRequest create() =>
      ValidateWorkspaceRepoBindingRequest._();
  static $pb.GeneratedMessage $_createMessage() =>
      ValidateWorkspaceRepoBindingRequest._();
  @$core.override
  ValidateWorkspaceRepoBindingRequest createEmptyInstance() =>
      ValidateWorkspaceRepoBindingRequest._();
  @$core.pragma('dart2js:noInline')
  static ValidateWorkspaceRepoBindingRequest getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<
              ValidateWorkspaceRepoBindingRequest>(
          ValidateWorkspaceRepoBindingRequest.$_createMessage);
  static ValidateWorkspaceRepoBindingRequest? _defaultInstance;
}

class ValidateWorkspaceRepoBindingResponse extends $pb.GeneratedMessage {
  factory ValidateWorkspaceRepoBindingResponse({
    WorkspaceRepoBinding? binding,
  }) {
    final result = ValidateWorkspaceRepoBindingResponse._();
    if (binding != null) result.binding = binding;
    return result;
  }

  ValidateWorkspaceRepoBindingResponse._();

  factory ValidateWorkspaceRepoBindingResponse.fromBuffer(
          $core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ValidateWorkspaceRepoBindingResponse()..mergeFromBuffer(data, registry);
  factory ValidateWorkspaceRepoBindingResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ValidateWorkspaceRepoBindingResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ValidateWorkspaceRepoBindingResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: ValidateWorkspaceRepoBindingResponse.$_createMessage)
    ..aOM<WorkspaceRepoBinding>(1, _omitFieldNames ? '' : 'binding',
        subBuilder: WorkspaceRepoBinding.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ValidateWorkspaceRepoBindingResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ValidateWorkspaceRepoBindingResponse copyWith(
          void Function(ValidateWorkspaceRepoBindingResponse) updates) =>
      super.copyWith((message) =>
              updates(message as ValidateWorkspaceRepoBindingResponse))
          as ValidateWorkspaceRepoBindingResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use ValidateWorkspaceRepoBindingResponse() / ValidateWorkspaceRepoBindingResponse.new instead')
  static ValidateWorkspaceRepoBindingResponse create() =>
      ValidateWorkspaceRepoBindingResponse._();
  static $pb.GeneratedMessage $_createMessage() =>
      ValidateWorkspaceRepoBindingResponse._();
  @$core.override
  ValidateWorkspaceRepoBindingResponse createEmptyInstance() =>
      ValidateWorkspaceRepoBindingResponse._();
  @$core.pragma('dart2js:noInline')
  static ValidateWorkspaceRepoBindingResponse getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<
              ValidateWorkspaceRepoBindingResponse>(
          ValidateWorkspaceRepoBindingResponse.$_createMessage);
  static ValidateWorkspaceRepoBindingResponse? _defaultInstance;

  @$pb.TagNumber(1)
  WorkspaceRepoBinding get binding => $_getN(0);
  @$pb.TagNumber(1)
  set binding(WorkspaceRepoBinding value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasBinding() => $_has(0);
  @$pb.TagNumber(1)
  void clearBinding() => $_clearField(1);
  @$pb.TagNumber(1)
  WorkspaceRepoBinding ensureBinding() => $_ensure(0);
}

/// SyncWorkspaceRepository triggers a synchronization of the bound
/// repository's managed tree into the workspace-scoped DB cache. Requires
/// owner or admin role. Idempotent when the remote HEAD has not changed.
class SyncWorkspaceRepositoryRequest extends $pb.GeneratedMessage {
  factory SyncWorkspaceRepositoryRequest() =>
      SyncWorkspaceRepositoryRequest._();

  SyncWorkspaceRepositoryRequest._();

  factory SyncWorkspaceRepositoryRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      SyncWorkspaceRepositoryRequest()..mergeFromBuffer(data, registry);
  factory SyncWorkspaceRepositoryRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      SyncWorkspaceRepositoryRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'SyncWorkspaceRepositoryRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: SyncWorkspaceRepositoryRequest.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  SyncWorkspaceRepositoryRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  SyncWorkspaceRepositoryRequest copyWith(
          void Function(SyncWorkspaceRepositoryRequest) updates) =>
      super.copyWith(
              (message) => updates(message as SyncWorkspaceRepositoryRequest))
          as SyncWorkspaceRepositoryRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use SyncWorkspaceRepositoryRequest() / SyncWorkspaceRepositoryRequest.new instead')
  static SyncWorkspaceRepositoryRequest create() =>
      SyncWorkspaceRepositoryRequest._();
  static $pb.GeneratedMessage $_createMessage() =>
      SyncWorkspaceRepositoryRequest._();
  @$core.override
  SyncWorkspaceRepositoryRequest createEmptyInstance() =>
      SyncWorkspaceRepositoryRequest._();
  @$core.pragma('dart2js:noInline')
  static SyncWorkspaceRepositoryRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<SyncWorkspaceRepositoryRequest>(
          SyncWorkspaceRepositoryRequest.$_createMessage);
  static SyncWorkspaceRepositoryRequest? _defaultInstance;
}

class SyncWorkspaceRepositoryResponse extends $pb.GeneratedMessage {
  factory SyncWorkspaceRepositoryResponse({
    WorkspaceRepoBinding? binding,
    $core.int? entriesSynced,
    $core.bool? published,
    $core.Iterable<$core.String>? publicationErrors,
  }) {
    final result = SyncWorkspaceRepositoryResponse._();
    if (binding != null) result.binding = binding;
    if (entriesSynced != null) result.entriesSynced = entriesSynced;
    if (published != null) result.published = published;
    if (publicationErrors != null)
      result.publicationErrors.addAll(publicationErrors);
    return result;
  }

  SyncWorkspaceRepositoryResponse._();

  factory SyncWorkspaceRepositoryResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      SyncWorkspaceRepositoryResponse()..mergeFromBuffer(data, registry);
  factory SyncWorkspaceRepositoryResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      SyncWorkspaceRepositoryResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'SyncWorkspaceRepositoryResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: SyncWorkspaceRepositoryResponse.$_createMessage)
    ..aOM<WorkspaceRepoBinding>(1, _omitFieldNames ? '' : 'binding',
        subBuilder: WorkspaceRepoBinding.$_createMessage)
    ..aI(2, _omitFieldNames ? '' : 'entriesSynced')
    ..aOB(3, _omitFieldNames ? '' : 'published')
    ..pPS(4, _omitFieldNames ? '' : 'publicationErrors')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  SyncWorkspaceRepositoryResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  SyncWorkspaceRepositoryResponse copyWith(
          void Function(SyncWorkspaceRepositoryResponse) updates) =>
      super.copyWith(
              (message) => updates(message as SyncWorkspaceRepositoryResponse))
          as SyncWorkspaceRepositoryResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use SyncWorkspaceRepositoryResponse() / SyncWorkspaceRepositoryResponse.new instead')
  static SyncWorkspaceRepositoryResponse create() =>
      SyncWorkspaceRepositoryResponse._();
  static $pb.GeneratedMessage $_createMessage() =>
      SyncWorkspaceRepositoryResponse._();
  @$core.override
  SyncWorkspaceRepositoryResponse createEmptyInstance() =>
      SyncWorkspaceRepositoryResponse._();
  @$core.pragma('dart2js:noInline')
  static SyncWorkspaceRepositoryResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<SyncWorkspaceRepositoryResponse>(
          SyncWorkspaceRepositoryResponse.$_createMessage);
  static SyncWorkspaceRepositoryResponse? _defaultInstance;

  @$pb.TagNumber(1)
  WorkspaceRepoBinding get binding => $_getN(0);
  @$pb.TagNumber(1)
  set binding(WorkspaceRepoBinding value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasBinding() => $_has(0);
  @$pb.TagNumber(1)
  void clearBinding() => $_clearField(1);
  @$pb.TagNumber(1)
  WorkspaceRepoBinding ensureBinding() => $_ensure(0);

  /// Number of entries written to the cache during this sync.
  @$pb.TagNumber(2)
  $core.int get entriesSynced => $_getIZ(1);
  @$pb.TagNumber(2)
  set entriesSynced($core.int value) => $_setSignedInt32(1, value);
  @$pb.TagNumber(2)
  $core.bool hasEntriesSynced() => $_has(1);
  @$pb.TagNumber(2)
  void clearEntriesSynced() => $_clearField(2);

  /// Whether publication was attempted and succeeded after this sync.
  @$pb.TagNumber(3)
  $core.bool get published => $_getBF(2);
  @$pb.TagNumber(3)
  set published($core.bool value) => $_setBool(2, value);
  @$pb.TagNumber(3)
  $core.bool hasPublished() => $_has(2);
  @$pb.TagNumber(3)
  void clearPublished() => $_clearField(3);

  /// Per-agent validation errors if publication failed.
  @$pb.TagNumber(4)
  $pb.PbList<$core.String> get publicationErrors => $_getList(3);
}

/// GetRepositorySyncStatus returns the current sync state without triggering
/// a new synchronization. Available to every workspace member.
class GetRepositorySyncStatusRequest extends $pb.GeneratedMessage {
  factory GetRepositorySyncStatusRequest() =>
      GetRepositorySyncStatusRequest._();

  GetRepositorySyncStatusRequest._();

  factory GetRepositorySyncStatusRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetRepositorySyncStatusRequest()..mergeFromBuffer(data, registry);
  factory GetRepositorySyncStatusRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetRepositorySyncStatusRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetRepositorySyncStatusRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: GetRepositorySyncStatusRequest.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetRepositorySyncStatusRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetRepositorySyncStatusRequest copyWith(
          void Function(GetRepositorySyncStatusRequest) updates) =>
      super.copyWith(
              (message) => updates(message as GetRepositorySyncStatusRequest))
          as GetRepositorySyncStatusRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use GetRepositorySyncStatusRequest() / GetRepositorySyncStatusRequest.new instead')
  static GetRepositorySyncStatusRequest create() =>
      GetRepositorySyncStatusRequest._();
  static $pb.GeneratedMessage $_createMessage() =>
      GetRepositorySyncStatusRequest._();
  @$core.override
  GetRepositorySyncStatusRequest createEmptyInstance() =>
      GetRepositorySyncStatusRequest._();
  @$core.pragma('dart2js:noInline')
  static GetRepositorySyncStatusRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<GetRepositorySyncStatusRequest>(
          GetRepositorySyncStatusRequest.$_createMessage);
  static GetRepositorySyncStatusRequest? _defaultInstance;
}

class GetRepositorySyncStatusResponse extends $pb.GeneratedMessage {
  factory GetRepositorySyncStatusResponse({
    WorkspaceRepoBinding? binding,
  }) {
    final result = GetRepositorySyncStatusResponse._();
    if (binding != null) result.binding = binding;
    return result;
  }

  GetRepositorySyncStatusResponse._();

  factory GetRepositorySyncStatusResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetRepositorySyncStatusResponse()..mergeFromBuffer(data, registry);
  factory GetRepositorySyncStatusResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetRepositorySyncStatusResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetRepositorySyncStatusResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: GetRepositorySyncStatusResponse.$_createMessage)
    ..aOM<WorkspaceRepoBinding>(1, _omitFieldNames ? '' : 'binding',
        subBuilder: WorkspaceRepoBinding.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetRepositorySyncStatusResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetRepositorySyncStatusResponse copyWith(
          void Function(GetRepositorySyncStatusResponse) updates) =>
      super.copyWith(
              (message) => updates(message as GetRepositorySyncStatusResponse))
          as GetRepositorySyncStatusResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use GetRepositorySyncStatusResponse() / GetRepositorySyncStatusResponse.new instead')
  static GetRepositorySyncStatusResponse create() =>
      GetRepositorySyncStatusResponse._();
  static $pb.GeneratedMessage $_createMessage() =>
      GetRepositorySyncStatusResponse._();
  @$core.override
  GetRepositorySyncStatusResponse createEmptyInstance() =>
      GetRepositorySyncStatusResponse._();
  @$core.pragma('dart2js:noInline')
  static GetRepositorySyncStatusResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<GetRepositorySyncStatusResponse>(
          GetRepositorySyncStatusResponse.$_createMessage);
  static GetRepositorySyncStatusResponse? _defaultInstance;

  @$pb.TagNumber(1)
  WorkspaceRepoBinding get binding => $_getN(0);
  @$pb.TagNumber(1)
  set binding(WorkspaceRepoBinding value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasBinding() => $_has(0);
  @$pb.TagNumber(1)
  void clearBinding() => $_clearField(1);
  @$pb.TagNumber(1)
  WorkspaceRepoBinding ensureBinding() => $_ensure(0);
}

/// ListRepositoryEntries returns cached directory entries for a path within
/// the managed tree. Reads only from the DB cache.
class ListRepositoryEntriesRequest extends $pb.GeneratedMessage {
  factory ListRepositoryEntriesRequest({
    $core.String? path,
  }) {
    final result = ListRepositoryEntriesRequest._();
    if (path != null) result.path = path;
    return result;
  }

  ListRepositoryEntriesRequest._();

  factory ListRepositoryEntriesRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListRepositoryEntriesRequest()..mergeFromBuffer(data, registry);
  factory ListRepositoryEntriesRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListRepositoryEntriesRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListRepositoryEntriesRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: ListRepositoryEntriesRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'path')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListRepositoryEntriesRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListRepositoryEntriesRequest copyWith(
          void Function(ListRepositoryEntriesRequest) updates) =>
      super.copyWith(
              (message) => updates(message as ListRepositoryEntriesRequest))
          as ListRepositoryEntriesRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use ListRepositoryEntriesRequest() / ListRepositoryEntriesRequest.new instead')
  static ListRepositoryEntriesRequest create() =>
      ListRepositoryEntriesRequest._();
  static $pb.GeneratedMessage $_createMessage() =>
      ListRepositoryEntriesRequest._();
  @$core.override
  ListRepositoryEntriesRequest createEmptyInstance() =>
      ListRepositoryEntriesRequest._();
  @$core.pragma('dart2js:noInline')
  static ListRepositoryEntriesRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ListRepositoryEntriesRequest>(
          ListRepositoryEntriesRequest.$_createMessage);
  static ListRepositoryEntriesRequest? _defaultInstance;

  /// Repository-relative path to list. Empty lists the root.
  @$pb.TagNumber(1)
  $core.String get path => $_getSZ(0);
  @$pb.TagNumber(1)
  set path($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasPath() => $_has(0);
  @$pb.TagNumber(1)
  void clearPath() => $_clearField(1);
}

class ListRepositoryEntriesResponse extends $pb.GeneratedMessage {
  factory ListRepositoryEntriesResponse({
    $core.String? commitSha,
    $core.Iterable<RepoCacheEntry>? entries,
    $core.String? observedCommitSha,
    $core.String? activeCommitSha,
  }) {
    final result = ListRepositoryEntriesResponse._();
    if (commitSha != null) result.commitSha = commitSha;
    if (entries != null) result.entries.addAll(entries);
    if (observedCommitSha != null) result.observedCommitSha = observedCommitSha;
    if (activeCommitSha != null) result.activeCommitSha = activeCommitSha;
    return result;
  }

  ListRepositoryEntriesResponse._();

  factory ListRepositoryEntriesResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListRepositoryEntriesResponse()..mergeFromBuffer(data, registry);
  factory ListRepositoryEntriesResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ListRepositoryEntriesResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListRepositoryEntriesResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: ListRepositoryEntriesResponse.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'commitSha')
    ..pPM<RepoCacheEntry>(2, _omitFieldNames ? '' : 'entries',
        subBuilder: RepoCacheEntry.$_createMessage)
    ..aOS(3, _omitFieldNames ? '' : 'observedCommitSha')
    ..aOS(4, _omitFieldNames ? '' : 'activeCommitSha')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListRepositoryEntriesResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListRepositoryEntriesResponse copyWith(
          void Function(ListRepositoryEntriesResponse) updates) =>
      super.copyWith(
              (message) => updates(message as ListRepositoryEntriesResponse))
          as ListRepositoryEntriesResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use ListRepositoryEntriesResponse() / ListRepositoryEntriesResponse.new instead')
  static ListRepositoryEntriesResponse create() =>
      ListRepositoryEntriesResponse._();
  static $pb.GeneratedMessage $_createMessage() =>
      ListRepositoryEntriesResponse._();
  @$core.override
  ListRepositoryEntriesResponse createEmptyInstance() =>
      ListRepositoryEntriesResponse._();
  @$core.pragma('dart2js:noInline')
  static ListRepositoryEntriesResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ListRepositoryEntriesResponse>(
          ListRepositoryEntriesResponse.$_createMessage);
  static ListRepositoryEntriesResponse? _defaultInstance;

  /// Commit SHA the cache was built from (deprecated; use observed/active).
  @$pb.TagNumber(1)
  $core.String get commitSha => $_getSZ(0);
  @$pb.TagNumber(1)
  set commitSha($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasCommitSha() => $_has(0);
  @$pb.TagNumber(1)
  void clearCommitSha() => $_clearField(1);

  @$pb.TagNumber(2)
  $pb.PbList<RepoCacheEntry> get entries => $_getList(1);

  /// The commit SHA last observed from the remote branch.
  @$pb.TagNumber(3)
  $core.String get observedCommitSha => $_getSZ(2);
  @$pb.TagNumber(3)
  set observedCommitSha($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasObservedCommitSha() => $_has(2);
  @$pb.TagNumber(3)
  void clearObservedCommitSha() => $_clearField(3);

  /// The commit SHA whose content is currently in the cache.
  @$pb.TagNumber(4)
  $core.String get activeCommitSha => $_getSZ(3);
  @$pb.TagNumber(4)
  set activeCommitSha($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasActiveCommitSha() => $_has(3);
  @$pb.TagNumber(4)
  void clearActiveCommitSha() => $_clearField(4);
}

/// GetRepositoryFile returns the cached content of a single file.
class GetRepositoryFileRequest extends $pb.GeneratedMessage {
  factory GetRepositoryFileRequest({
    $core.String? path,
  }) {
    final result = GetRepositoryFileRequest._();
    if (path != null) result.path = path;
    return result;
  }

  GetRepositoryFileRequest._();

  factory GetRepositoryFileRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetRepositoryFileRequest()..mergeFromBuffer(data, registry);
  factory GetRepositoryFileRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetRepositoryFileRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetRepositoryFileRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: GetRepositoryFileRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'path')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetRepositoryFileRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetRepositoryFileRequest copyWith(
          void Function(GetRepositoryFileRequest) updates) =>
      super.copyWith((message) => updates(message as GetRepositoryFileRequest))
          as GetRepositoryFileRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use GetRepositoryFileRequest() / GetRepositoryFileRequest.new instead')
  static GetRepositoryFileRequest create() => GetRepositoryFileRequest._();
  static $pb.GeneratedMessage $_createMessage() => GetRepositoryFileRequest._();
  @$core.override
  GetRepositoryFileRequest createEmptyInstance() =>
      GetRepositoryFileRequest._();
  @$core.pragma('dart2js:noInline')
  static GetRepositoryFileRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<GetRepositoryFileRequest>(
          GetRepositoryFileRequest.$_createMessage);
  static GetRepositoryFileRequest? _defaultInstance;

  /// Repository-relative file path.
  @$pb.TagNumber(1)
  $core.String get path => $_getSZ(0);
  @$pb.TagNumber(1)
  set path($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasPath() => $_has(0);
  @$pb.TagNumber(1)
  void clearPath() => $_clearField(1);
}

class GetRepositoryFileResponse extends $pb.GeneratedMessage {
  factory GetRepositoryFileResponse({
    $core.String? commitSha,
    RepoCacheEntry? entry,
    $core.String? content,
    $core.String? observedCommitSha,
    $core.String? activeCommitSha,
  }) {
    final result = GetRepositoryFileResponse._();
    if (commitSha != null) result.commitSha = commitSha;
    if (entry != null) result.entry = entry;
    if (content != null) result.content = content;
    if (observedCommitSha != null) result.observedCommitSha = observedCommitSha;
    if (activeCommitSha != null) result.activeCommitSha = activeCommitSha;
    return result;
  }

  GetRepositoryFileResponse._();

  factory GetRepositoryFileResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetRepositoryFileResponse()..mergeFromBuffer(data, registry);
  factory GetRepositoryFileResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      GetRepositoryFileResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetRepositoryFileResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: GetRepositoryFileResponse.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'commitSha')
    ..aOM<RepoCacheEntry>(2, _omitFieldNames ? '' : 'entry',
        subBuilder: RepoCacheEntry.$_createMessage)
    ..aOS(3, _omitFieldNames ? '' : 'content')
    ..aOS(4, _omitFieldNames ? '' : 'observedCommitSha')
    ..aOS(5, _omitFieldNames ? '' : 'activeCommitSha')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetRepositoryFileResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetRepositoryFileResponse copyWith(
          void Function(GetRepositoryFileResponse) updates) =>
      super.copyWith((message) => updates(message as GetRepositoryFileResponse))
          as GetRepositoryFileResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use GetRepositoryFileResponse() / GetRepositoryFileResponse.new instead')
  static GetRepositoryFileResponse create() => GetRepositoryFileResponse._();
  static $pb.GeneratedMessage $_createMessage() =>
      GetRepositoryFileResponse._();
  @$core.override
  GetRepositoryFileResponse createEmptyInstance() =>
      GetRepositoryFileResponse._();
  @$core.pragma('dart2js:noInline')
  static GetRepositoryFileResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<GetRepositoryFileResponse>(
          GetRepositoryFileResponse.$_createMessage);
  static GetRepositoryFileResponse? _defaultInstance;

  /// Commit SHA the cache was built from (deprecated; use observed/active).
  @$pb.TagNumber(1)
  $core.String get commitSha => $_getSZ(0);
  @$pb.TagNumber(1)
  set commitSha($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasCommitSha() => $_has(0);
  @$pb.TagNumber(1)
  void clearCommitSha() => $_clearField(1);

  @$pb.TagNumber(2)
  RepoCacheEntry get entry => $_getN(1);
  @$pb.TagNumber(2)
  set entry(RepoCacheEntry value) => $_setField(2, value);
  @$pb.TagNumber(2)
  $core.bool hasEntry() => $_has(1);
  @$pb.TagNumber(2)
  void clearEntry() => $_clearField(2);
  @$pb.TagNumber(2)
  RepoCacheEntry ensureEntry() => $_ensure(1);

  /// UTF-8 file content.
  @$pb.TagNumber(3)
  $core.String get content => $_getSZ(2);
  @$pb.TagNumber(3)
  set content($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasContent() => $_has(2);
  @$pb.TagNumber(3)
  void clearContent() => $_clearField(3);

  /// The commit SHA last observed from the remote branch.
  @$pb.TagNumber(4)
  $core.String get observedCommitSha => $_getSZ(3);
  @$pb.TagNumber(4)
  set observedCommitSha($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasObservedCommitSha() => $_has(3);
  @$pb.TagNumber(4)
  void clearObservedCommitSha() => $_clearField(4);

  /// The commit SHA whose content is currently in the cache.
  @$pb.TagNumber(5)
  $core.String get activeCommitSha => $_getSZ(4);
  @$pb.TagNumber(5)
  set activeCommitSha($core.String value) => $_setString(4, value);
  @$pb.TagNumber(5)
  $core.bool hasActiveCommitSha() => $_has(4);
  @$pb.TagNumber(5)
  void clearActiveCommitSha() => $_clearField(5);
}

/// PublishWorkspaceRepository validates the current observed cache and, when
/// valid, atomically writes Agent Content snapshots and advances the active
/// revision.
class PublishWorkspaceRepositoryRequest extends $pb.GeneratedMessage {
  factory PublishWorkspaceRepositoryRequest() =>
      PublishWorkspaceRepositoryRequest._();

  PublishWorkspaceRepositoryRequest._();

  factory PublishWorkspaceRepositoryRequest.fromBuffer(
          $core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      PublishWorkspaceRepositoryRequest()..mergeFromBuffer(data, registry);
  factory PublishWorkspaceRepositoryRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      PublishWorkspaceRepositoryRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'PublishWorkspaceRepositoryRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: PublishWorkspaceRepositoryRequest.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  PublishWorkspaceRepositoryRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  PublishWorkspaceRepositoryRequest copyWith(
          void Function(PublishWorkspaceRepositoryRequest) updates) =>
      super.copyWith((message) =>
              updates(message as PublishWorkspaceRepositoryRequest))
          as PublishWorkspaceRepositoryRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use PublishWorkspaceRepositoryRequest() / PublishWorkspaceRepositoryRequest.new instead')
  static PublishWorkspaceRepositoryRequest create() =>
      PublishWorkspaceRepositoryRequest._();
  static $pb.GeneratedMessage $_createMessage() =>
      PublishWorkspaceRepositoryRequest._();
  @$core.override
  PublishWorkspaceRepositoryRequest createEmptyInstance() =>
      PublishWorkspaceRepositoryRequest._();
  @$core.pragma('dart2js:noInline')
  static PublishWorkspaceRepositoryRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<PublishWorkspaceRepositoryRequest>(
          PublishWorkspaceRepositoryRequest.$_createMessage);
  static PublishWorkspaceRepositoryRequest? _defaultInstance;
}

class PublishWorkspaceRepositoryResponse extends $pb.GeneratedMessage {
  factory PublishWorkspaceRepositoryResponse({
    WorkspaceRepoBinding? binding,
    $core.Iterable<$core.String>? validationErrors,
  }) {
    final result = PublishWorkspaceRepositoryResponse._();
    if (binding != null) result.binding = binding;
    if (validationErrors != null)
      result.validationErrors.addAll(validationErrors);
    return result;
  }

  PublishWorkspaceRepositoryResponse._();

  factory PublishWorkspaceRepositoryResponse.fromBuffer(
          $core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      PublishWorkspaceRepositoryResponse()..mergeFromBuffer(data, registry);
  factory PublishWorkspaceRepositoryResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      PublishWorkspaceRepositoryResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'PublishWorkspaceRepositoryResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: PublishWorkspaceRepositoryResponse.$_createMessage)
    ..aOM<WorkspaceRepoBinding>(1, _omitFieldNames ? '' : 'binding',
        subBuilder: WorkspaceRepoBinding.$_createMessage)
    ..pPS(2, _omitFieldNames ? '' : 'validationErrors')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  PublishWorkspaceRepositoryResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  PublishWorkspaceRepositoryResponse copyWith(
          void Function(PublishWorkspaceRepositoryResponse) updates) =>
      super.copyWith((message) =>
              updates(message as PublishWorkspaceRepositoryResponse))
          as PublishWorkspaceRepositoryResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use PublishWorkspaceRepositoryResponse() / PublishWorkspaceRepositoryResponse.new instead')
  static PublishWorkspaceRepositoryResponse create() =>
      PublishWorkspaceRepositoryResponse._();
  static $pb.GeneratedMessage $_createMessage() =>
      PublishWorkspaceRepositoryResponse._();
  @$core.override
  PublishWorkspaceRepositoryResponse createEmptyInstance() =>
      PublishWorkspaceRepositoryResponse._();
  @$core.pragma('dart2js:noInline')
  static PublishWorkspaceRepositoryResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<PublishWorkspaceRepositoryResponse>(
          PublishWorkspaceRepositoryResponse.$_createMessage);
  static PublishWorkspaceRepositoryResponse? _defaultInstance;

  @$pb.TagNumber(1)
  WorkspaceRepoBinding get binding => $_getN(0);
  @$pb.TagNumber(1)
  set binding(WorkspaceRepoBinding value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasBinding() => $_has(0);
  @$pb.TagNumber(1)
  void clearBinding() => $_clearField(1);
  @$pb.TagNumber(1)
  WorkspaceRepoBinding ensureBinding() => $_ensure(0);

  /// Per-agent validation errors (empty on success).
  @$pb.TagNumber(2)
  $pb.PbList<$core.String> get validationErrors => $_getList(1);
}

/// ConfigureWebhookSecret generates a new webhook secret for the binding.
class ConfigureWebhookSecretRequest extends $pb.GeneratedMessage {
  factory ConfigureWebhookSecretRequest() => ConfigureWebhookSecretRequest._();

  ConfigureWebhookSecretRequest._();

  factory ConfigureWebhookSecretRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ConfigureWebhookSecretRequest()..mergeFromBuffer(data, registry);
  factory ConfigureWebhookSecretRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ConfigureWebhookSecretRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ConfigureWebhookSecretRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: ConfigureWebhookSecretRequest.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ConfigureWebhookSecretRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ConfigureWebhookSecretRequest copyWith(
          void Function(ConfigureWebhookSecretRequest) updates) =>
      super.copyWith(
              (message) => updates(message as ConfigureWebhookSecretRequest))
          as ConfigureWebhookSecretRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use ConfigureWebhookSecretRequest() / ConfigureWebhookSecretRequest.new instead')
  static ConfigureWebhookSecretRequest create() =>
      ConfigureWebhookSecretRequest._();
  static $pb.GeneratedMessage $_createMessage() =>
      ConfigureWebhookSecretRequest._();
  @$core.override
  ConfigureWebhookSecretRequest createEmptyInstance() =>
      ConfigureWebhookSecretRequest._();
  @$core.pragma('dart2js:noInline')
  static ConfigureWebhookSecretRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ConfigureWebhookSecretRequest>(
          ConfigureWebhookSecretRequest.$_createMessage);
  static ConfigureWebhookSecretRequest? _defaultInstance;
}

class ConfigureWebhookSecretResponse extends $pb.GeneratedMessage {
  factory ConfigureWebhookSecretResponse({
    WorkspaceRepoBinding? binding,
    $core.String? webhookSecret,
    $core.String? callbackUrl,
  }) {
    final result = ConfigureWebhookSecretResponse._();
    if (binding != null) result.binding = binding;
    if (webhookSecret != null) result.webhookSecret = webhookSecret;
    if (callbackUrl != null) result.callbackUrl = callbackUrl;
    return result;
  }

  ConfigureWebhookSecretResponse._();

  factory ConfigureWebhookSecretResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ConfigureWebhookSecretResponse()..mergeFromBuffer(data, registry);
  factory ConfigureWebhookSecretResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ConfigureWebhookSecretResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ConfigureWebhookSecretResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: ConfigureWebhookSecretResponse.$_createMessage)
    ..aOM<WorkspaceRepoBinding>(1, _omitFieldNames ? '' : 'binding',
        subBuilder: WorkspaceRepoBinding.$_createMessage)
    ..aOS(2, _omitFieldNames ? '' : 'webhookSecret')
    ..aOS(3, _omitFieldNames ? '' : 'callbackUrl')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ConfigureWebhookSecretResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ConfigureWebhookSecretResponse copyWith(
          void Function(ConfigureWebhookSecretResponse) updates) =>
      super.copyWith(
              (message) => updates(message as ConfigureWebhookSecretResponse))
          as ConfigureWebhookSecretResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use ConfigureWebhookSecretResponse() / ConfigureWebhookSecretResponse.new instead')
  static ConfigureWebhookSecretResponse create() =>
      ConfigureWebhookSecretResponse._();
  static $pb.GeneratedMessage $_createMessage() =>
      ConfigureWebhookSecretResponse._();
  @$core.override
  ConfigureWebhookSecretResponse createEmptyInstance() =>
      ConfigureWebhookSecretResponse._();
  @$core.pragma('dart2js:noInline')
  static ConfigureWebhookSecretResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ConfigureWebhookSecretResponse>(
          ConfigureWebhookSecretResponse.$_createMessage);
  static ConfigureWebhookSecretResponse? _defaultInstance;

  @$pb.TagNumber(1)
  WorkspaceRepoBinding get binding => $_getN(0);
  @$pb.TagNumber(1)
  set binding(WorkspaceRepoBinding value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasBinding() => $_has(0);
  @$pb.TagNumber(1)
  void clearBinding() => $_clearField(1);
  @$pb.TagNumber(1)
  WorkspaceRepoBinding ensureBinding() => $_ensure(0);

  /// The webhook secret, shown only once. Store it in the Git host's webhook
  /// configuration. Never persisted in plaintext; subsequent reads only see
  /// webhook_secret_set=true on the binding.
  @$pb.TagNumber(2)
  $core.String get webhookSecret => $_getSZ(1);
  @$pb.TagNumber(2)
  set webhookSecret($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasWebhookSecret() => $_has(1);
  @$pb.TagNumber(2)
  void clearWebhookSecret() => $_clearField(2);

  /// The callback URL to configure in the Git host webhook settings.
  @$pb.TagNumber(3)
  $core.String get callbackUrl => $_getSZ(2);
  @$pb.TagNumber(3)
  set callbackUrl($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasCallbackUrl() => $_has(2);
  @$pb.TagNumber(3)
  void clearCallbackUrl() => $_clearField(3);
}

/// AcceptRepositoryBaseline accepts a force-pushed HEAD as the new baseline
/// and re-syncs.
class AcceptRepositoryBaselineRequest extends $pb.GeneratedMessage {
  factory AcceptRepositoryBaselineRequest() =>
      AcceptRepositoryBaselineRequest._();

  AcceptRepositoryBaselineRequest._();

  factory AcceptRepositoryBaselineRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      AcceptRepositoryBaselineRequest()..mergeFromBuffer(data, registry);
  factory AcceptRepositoryBaselineRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      AcceptRepositoryBaselineRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'AcceptRepositoryBaselineRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: AcceptRepositoryBaselineRequest.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  AcceptRepositoryBaselineRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  AcceptRepositoryBaselineRequest copyWith(
          void Function(AcceptRepositoryBaselineRequest) updates) =>
      super.copyWith(
              (message) => updates(message as AcceptRepositoryBaselineRequest))
          as AcceptRepositoryBaselineRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use AcceptRepositoryBaselineRequest() / AcceptRepositoryBaselineRequest.new instead')
  static AcceptRepositoryBaselineRequest create() =>
      AcceptRepositoryBaselineRequest._();
  static $pb.GeneratedMessage $_createMessage() =>
      AcceptRepositoryBaselineRequest._();
  @$core.override
  AcceptRepositoryBaselineRequest createEmptyInstance() =>
      AcceptRepositoryBaselineRequest._();
  @$core.pragma('dart2js:noInline')
  static AcceptRepositoryBaselineRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<AcceptRepositoryBaselineRequest>(
          AcceptRepositoryBaselineRequest.$_createMessage);
  static AcceptRepositoryBaselineRequest? _defaultInstance;
}

class AcceptRepositoryBaselineResponse extends $pb.GeneratedMessage {
  factory AcceptRepositoryBaselineResponse({
    WorkspaceRepoBinding? binding,
  }) {
    final result = AcceptRepositoryBaselineResponse._();
    if (binding != null) result.binding = binding;
    return result;
  }

  AcceptRepositoryBaselineResponse._();

  factory AcceptRepositoryBaselineResponse.fromBuffer(
          $core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      AcceptRepositoryBaselineResponse()..mergeFromBuffer(data, registry);
  factory AcceptRepositoryBaselineResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      AcceptRepositoryBaselineResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'AcceptRepositoryBaselineResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: AcceptRepositoryBaselineResponse.$_createMessage)
    ..aOM<WorkspaceRepoBinding>(1, _omitFieldNames ? '' : 'binding',
        subBuilder: WorkspaceRepoBinding.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  AcceptRepositoryBaselineResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  AcceptRepositoryBaselineResponse copyWith(
          void Function(AcceptRepositoryBaselineResponse) updates) =>
      super.copyWith(
              (message) => updates(message as AcceptRepositoryBaselineResponse))
          as AcceptRepositoryBaselineResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use AcceptRepositoryBaselineResponse() / AcceptRepositoryBaselineResponse.new instead')
  static AcceptRepositoryBaselineResponse create() =>
      AcceptRepositoryBaselineResponse._();
  static $pb.GeneratedMessage $_createMessage() =>
      AcceptRepositoryBaselineResponse._();
  @$core.override
  AcceptRepositoryBaselineResponse createEmptyInstance() =>
      AcceptRepositoryBaselineResponse._();
  @$core.pragma('dart2js:noInline')
  static AcceptRepositoryBaselineResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<AcceptRepositoryBaselineResponse>(
          AcceptRepositoryBaselineResponse.$_createMessage);
  static AcceptRepositoryBaselineResponse? _defaultInstance;

  @$pb.TagNumber(1)
  WorkspaceRepoBinding get binding => $_getN(0);
  @$pb.TagNumber(1)
  set binding(WorkspaceRepoBinding value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasBinding() => $_has(0);
  @$pb.TagNumber(1)
  void clearBinding() => $_clearField(1);
  @$pb.TagNumber(1)
  WorkspaceRepoBinding ensureBinding() => $_ensure(0);
}

/// ContentFileAction describes one file change within a CommitAgentContent
/// changeset.
class ContentFileAction extends $pb.GeneratedMessage {
  factory ContentFileAction({
    $core.String? path,
    ContentFileOperation? operation,
    $core.String? content,
  }) {
    final result = ContentFileAction._();
    if (path != null) result.path = path;
    if (operation != null) result.operation = operation;
    if (content != null) result.content = content;
    return result;
  }

  ContentFileAction._();

  factory ContentFileAction.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ContentFileAction()..mergeFromBuffer(data, registry);
  factory ContentFileAction.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ContentFileAction()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ContentFileAction',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: ContentFileAction.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'path')
    ..aE<ContentFileOperation>(2, _omitFieldNames ? '' : 'operation',
        enumValues: ContentFileOperation.values)
    ..aOS(3, _omitFieldNames ? '' : 'content')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ContentFileAction clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ContentFileAction copyWith(void Function(ContentFileAction) updates) =>
      super.copyWith((message) => updates(message as ContentFileAction))
          as ContentFileAction;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use ContentFileAction() / ContentFileAction.new instead')
  static ContentFileAction create() => ContentFileAction._();
  static $pb.GeneratedMessage $_createMessage() => ContentFileAction._();
  @$core.override
  ContentFileAction createEmptyInstance() => ContentFileAction._();
  @$core.pragma('dart2js:noInline')
  static ContentFileAction getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<ContentFileAction>(
          ContentFileAction.$_createMessage);
  static ContentFileAction? _defaultInstance;

  /// Repository-relative path within the managed agents subtree, e.g.
  /// "agents/my-agent/prompt.md".
  @$pb.TagNumber(1)
  $core.String get path => $_getSZ(0);
  @$pb.TagNumber(1)
  set path($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasPath() => $_has(0);
  @$pb.TagNumber(1)
  void clearPath() => $_clearField(1);

  /// Whether this action creates/updates or deletes the file.
  @$pb.TagNumber(2)
  ContentFileOperation get operation => $_getN(1);
  @$pb.TagNumber(2)
  set operation(ContentFileOperation value) => $_setField(2, value);
  @$pb.TagNumber(2)
  $core.bool hasOperation() => $_has(1);
  @$pb.TagNumber(2)
  void clearOperation() => $_clearField(2);

  /// UTF-8 Markdown content. Required for PUT, ignored for DELETE.
  @$pb.TagNumber(3)
  $core.String get content => $_getSZ(2);
  @$pb.TagNumber(3)
  set content($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasContent() => $_has(2);
  @$pb.TagNumber(3)
  void clearContent() => $_clearField(3);
}

/// CommitAgentContentRequest applies a multi-file changeset to managed Agent
/// Content and produces a single Git commit.
class CommitAgentContentRequest extends $pb.GeneratedMessage {
  factory CommitAgentContentRequest({
    $core.Iterable<ContentFileAction>? actions,
    $core.String? message,
    $core.String? baseRevision,
  }) {
    final result = CommitAgentContentRequest._();
    if (actions != null) result.actions.addAll(actions);
    if (message != null) result.message = message;
    if (baseRevision != null) result.baseRevision = baseRevision;
    return result;
  }

  CommitAgentContentRequest._();

  factory CommitAgentContentRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CommitAgentContentRequest()..mergeFromBuffer(data, registry);
  factory CommitAgentContentRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CommitAgentContentRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CommitAgentContentRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: CommitAgentContentRequest.$_createMessage)
    ..pPM<ContentFileAction>(1, _omitFieldNames ? '' : 'actions',
        subBuilder: ContentFileAction.$_createMessage)
    ..aOS(2, _omitFieldNames ? '' : 'message')
    ..aOS(3, _omitFieldNames ? '' : 'baseRevision')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CommitAgentContentRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CommitAgentContentRequest copyWith(
          void Function(CommitAgentContentRequest) updates) =>
      super.copyWith((message) => updates(message as CommitAgentContentRequest))
          as CommitAgentContentRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use CommitAgentContentRequest() / CommitAgentContentRequest.new instead')
  static CommitAgentContentRequest create() => CommitAgentContentRequest._();
  static $pb.GeneratedMessage $_createMessage() =>
      CommitAgentContentRequest._();
  @$core.override
  CommitAgentContentRequest createEmptyInstance() =>
      CommitAgentContentRequest._();
  @$core.pragma('dart2js:noInline')
  static CommitAgentContentRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<CommitAgentContentRequest>(
          CommitAgentContentRequest.$_createMessage);
  static CommitAgentContentRequest? _defaultInstance;

  /// The ordered list of file changes. At least one action is required.
  @$pb.TagNumber(1)
  $pb.PbList<ContentFileAction> get actions => $_getList(0);

  /// Optional custom commit message. When empty, Butter generates one.
  @$pb.TagNumber(2)
  $core.String get message => $_getSZ(1);
  @$pb.TagNumber(2)
  set message($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasMessage() => $_has(1);
  @$pb.TagNumber(2)
  void clearMessage() => $_clearField(2);

  /// The dashboard's cached base revision SHA. Informational: recorded in
  /// commit audit metadata but not used for conflict detection (last-write-
  /// wins applies to the latest branch HEAD).
  @$pb.TagNumber(3)
  $core.String get baseRevision => $_getSZ(2);
  @$pb.TagNumber(3)
  set baseRevision($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasBaseRevision() => $_has(2);
  @$pb.TagNumber(3)
  void clearBaseRevision() => $_clearField(3);
}

/// CommitAgentContentResponse is returned after a successful content commit.
class CommitAgentContentResponse extends $pb.GeneratedMessage {
  factory CommitAgentContentResponse({
    WorkspaceRepoBinding? binding,
    $core.String? commitSha,
    $core.Iterable<$core.String>? validationErrors,
    $core.String? changeRequestUrl,
  }) {
    final result = CommitAgentContentResponse._();
    if (binding != null) result.binding = binding;
    if (commitSha != null) result.commitSha = commitSha;
    if (validationErrors != null)
      result.validationErrors.addAll(validationErrors);
    if (changeRequestUrl != null) result.changeRequestUrl = changeRequestUrl;
    return result;
  }

  CommitAgentContentResponse._();

  factory CommitAgentContentResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CommitAgentContentResponse()..mergeFromBuffer(data, registry);
  factory CommitAgentContentResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CommitAgentContentResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CommitAgentContentResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: CommitAgentContentResponse.$_createMessage)
    ..aOM<WorkspaceRepoBinding>(1, _omitFieldNames ? '' : 'binding',
        subBuilder: WorkspaceRepoBinding.$_createMessage)
    ..aOS(2, _omitFieldNames ? '' : 'commitSha')
    ..pPS(3, _omitFieldNames ? '' : 'validationErrors')
    ..aOS(4, _omitFieldNames ? '' : 'changeRequestUrl')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CommitAgentContentResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CommitAgentContentResponse copyWith(
          void Function(CommitAgentContentResponse) updates) =>
      super.copyWith(
              (message) => updates(message as CommitAgentContentResponse))
          as CommitAgentContentResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use CommitAgentContentResponse() / CommitAgentContentResponse.new instead')
  static CommitAgentContentResponse create() => CommitAgentContentResponse._();
  static $pb.GeneratedMessage $_createMessage() =>
      CommitAgentContentResponse._();
  @$core.override
  CommitAgentContentResponse createEmptyInstance() =>
      CommitAgentContentResponse._();
  @$core.pragma('dart2js:noInline')
  static CommitAgentContentResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<CommitAgentContentResponse>(
          CommitAgentContentResponse.$_createMessage);
  static CommitAgentContentResponse? _defaultInstance;

  @$pb.TagNumber(1)
  WorkspaceRepoBinding get binding => $_getN(0);
  @$pb.TagNumber(1)
  set binding(WorkspaceRepoBinding value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasBinding() => $_has(0);
  @$pb.TagNumber(1)
  void clearBinding() => $_clearField(1);
  @$pb.TagNumber(1)
  WorkspaceRepoBinding ensureBinding() => $_ensure(0);

  /// SHA of the created Git commit.
  @$pb.TagNumber(2)
  $core.String get commitSha => $_getSZ(1);
  @$pb.TagNumber(2)
  set commitSha($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasCommitSha() => $_has(1);
  @$pb.TagNumber(2)
  void clearCommitSha() => $_clearField(2);

  /// Per-agent validation errors that prevented the commit (empty on success).
  @$pb.TagNumber(3)
  $pb.PbList<$core.String> get validationErrors => $_getList(2);

  /// URL of the created Pull Request or Merge Request (only set when the
  /// binding write mode is CHANGE_REQUEST).
  @$pb.TagNumber(4)
  $core.String get changeRequestUrl => $_getSZ(3);
  @$pb.TagNumber(4)
  set changeRequestUrl($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasChangeRequestUrl() => $_has(3);
  @$pb.TagNumber(4)
  void clearChangeRequestUrl() => $_clearField(4);
}

/// RollbackAgentContentRequest restores managed content from a previous
/// published revision.
class RollbackAgentContentRequest extends $pb.GeneratedMessage {
  factory RollbackAgentContentRequest({
    $core.String? targetCommitSha,
  }) {
    final result = RollbackAgentContentRequest._();
    if (targetCommitSha != null) result.targetCommitSha = targetCommitSha;
    return result;
  }

  RollbackAgentContentRequest._();

  factory RollbackAgentContentRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      RollbackAgentContentRequest()..mergeFromBuffer(data, registry);
  factory RollbackAgentContentRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      RollbackAgentContentRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'RollbackAgentContentRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: RollbackAgentContentRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'targetCommitSha')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  RollbackAgentContentRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  RollbackAgentContentRequest copyWith(
          void Function(RollbackAgentContentRequest) updates) =>
      super.copyWith(
              (message) => updates(message as RollbackAgentContentRequest))
          as RollbackAgentContentRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use RollbackAgentContentRequest() / RollbackAgentContentRequest.new instead')
  static RollbackAgentContentRequest create() =>
      RollbackAgentContentRequest._();
  static $pb.GeneratedMessage $_createMessage() =>
      RollbackAgentContentRequest._();
  @$core.override
  RollbackAgentContentRequest createEmptyInstance() =>
      RollbackAgentContentRequest._();
  @$core.pragma('dart2js:noInline')
  static RollbackAgentContentRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<RollbackAgentContentRequest>(
          RollbackAgentContentRequest.$_createMessage);
  static RollbackAgentContentRequest? _defaultInstance;

  /// The commit SHA of the target revision to restore. Butter reads the
  /// managed content from this commit and creates a new commit with it.
  @$pb.TagNumber(1)
  $core.String get targetCommitSha => $_getSZ(0);
  @$pb.TagNumber(1)
  set targetCommitSha($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasTargetCommitSha() => $_has(0);
  @$pb.TagNumber(1)
  void clearTargetCommitSha() => $_clearField(1);
}

/// RollbackAgentContentResponse is returned after a successful rollback.
class RollbackAgentContentResponse extends $pb.GeneratedMessage {
  factory RollbackAgentContentResponse({
    WorkspaceRepoBinding? binding,
    $core.String? commitSha,
    $core.Iterable<$core.String>? validationErrors,
    $core.String? changeRequestUrl,
  }) {
    final result = RollbackAgentContentResponse._();
    if (binding != null) result.binding = binding;
    if (commitSha != null) result.commitSha = commitSha;
    if (validationErrors != null)
      result.validationErrors.addAll(validationErrors);
    if (changeRequestUrl != null) result.changeRequestUrl = changeRequestUrl;
    return result;
  }

  RollbackAgentContentResponse._();

  factory RollbackAgentContentResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      RollbackAgentContentResponse()..mergeFromBuffer(data, registry);
  factory RollbackAgentContentResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      RollbackAgentContentResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'RollbackAgentContentResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: RollbackAgentContentResponse.$_createMessage)
    ..aOM<WorkspaceRepoBinding>(1, _omitFieldNames ? '' : 'binding',
        subBuilder: WorkspaceRepoBinding.$_createMessage)
    ..aOS(2, _omitFieldNames ? '' : 'commitSha')
    ..pPS(3, _omitFieldNames ? '' : 'validationErrors')
    ..aOS(4, _omitFieldNames ? '' : 'changeRequestUrl')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  RollbackAgentContentResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  RollbackAgentContentResponse copyWith(
          void Function(RollbackAgentContentResponse) updates) =>
      super.copyWith(
              (message) => updates(message as RollbackAgentContentResponse))
          as RollbackAgentContentResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use RollbackAgentContentResponse() / RollbackAgentContentResponse.new instead')
  static RollbackAgentContentResponse create() =>
      RollbackAgentContentResponse._();
  static $pb.GeneratedMessage $_createMessage() =>
      RollbackAgentContentResponse._();
  @$core.override
  RollbackAgentContentResponse createEmptyInstance() =>
      RollbackAgentContentResponse._();
  @$core.pragma('dart2js:noInline')
  static RollbackAgentContentResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<RollbackAgentContentResponse>(
          RollbackAgentContentResponse.$_createMessage);
  static RollbackAgentContentResponse? _defaultInstance;

  @$pb.TagNumber(1)
  WorkspaceRepoBinding get binding => $_getN(0);
  @$pb.TagNumber(1)
  set binding(WorkspaceRepoBinding value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasBinding() => $_has(0);
  @$pb.TagNumber(1)
  void clearBinding() => $_clearField(1);
  @$pb.TagNumber(1)
  WorkspaceRepoBinding ensureBinding() => $_ensure(0);

  /// SHA of the new rollback commit.
  @$pb.TagNumber(2)
  $core.String get commitSha => $_getSZ(1);
  @$pb.TagNumber(2)
  set commitSha($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasCommitSha() => $_has(1);
  @$pb.TagNumber(2)
  void clearCommitSha() => $_clearField(2);

  /// Per-agent validation errors that prevented the rollback (empty on success).
  @$pb.TagNumber(3)
  $pb.PbList<$core.String> get validationErrors => $_getList(2);

  /// URL of the created Pull Request or Merge Request (only set when the
  /// binding write mode is CHANGE_REQUEST).
  @$pb.TagNumber(4)
  $core.String get changeRequestUrl => $_getSZ(3);
  @$pb.TagNumber(4)
  set changeRequestUrl($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasChangeRequestUrl() => $_has(3);
  @$pb.TagNumber(4)
  void clearChangeRequestUrl() => $_clearField(4);
}

/// PurgeAgentContentRequest removes the managed Content files for one Agent ID.
class PurgeAgentContentRequest extends $pb.GeneratedMessage {
  factory PurgeAgentContentRequest({
    $core.String? agentId,
  }) {
    final result = PurgeAgentContentRequest._();
    if (agentId != null) result.agentId = agentId;
    return result;
  }

  PurgeAgentContentRequest._();

  factory PurgeAgentContentRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      PurgeAgentContentRequest()..mergeFromBuffer(data, registry);
  factory PurgeAgentContentRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      PurgeAgentContentRequest()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'PurgeAgentContentRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: PurgeAgentContentRequest.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'agentId')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  PurgeAgentContentRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  PurgeAgentContentRequest copyWith(
          void Function(PurgeAgentContentRequest) updates) =>
      super.copyWith((message) => updates(message as PurgeAgentContentRequest))
          as PurgeAgentContentRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use PurgeAgentContentRequest() / PurgeAgentContentRequest.new instead')
  static PurgeAgentContentRequest create() => PurgeAgentContentRequest._();
  static $pb.GeneratedMessage $_createMessage() => PurgeAgentContentRequest._();
  @$core.override
  PurgeAgentContentRequest createEmptyInstance() =>
      PurgeAgentContentRequest._();
  @$core.pragma('dart2js:noInline')
  static PurgeAgentContentRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<PurgeAgentContentRequest>(
          PurgeAgentContentRequest.$_createMessage);
  static PurgeAgentContentRequest? _defaultInstance;

  /// The immutable Agent ID whose managed Content should be purged.
  @$pb.TagNumber(1)
  $core.String get agentId => $_getSZ(0);
  @$pb.TagNumber(1)
  set agentId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasAgentId() => $_has(0);
  @$pb.TagNumber(1)
  void clearAgentId() => $_clearField(1);
}

/// PurgeAgentContentResponse is returned after a successful purge.
class PurgeAgentContentResponse extends $pb.GeneratedMessage {
  factory PurgeAgentContentResponse({
    WorkspaceRepoBinding? binding,
    $core.String? commitSha,
    $core.Iterable<$core.String>? validationErrors,
    $core.String? changeRequestUrl,
  }) {
    final result = PurgeAgentContentResponse._();
    if (binding != null) result.binding = binding;
    if (commitSha != null) result.commitSha = commitSha;
    if (validationErrors != null)
      result.validationErrors.addAll(validationErrors);
    if (changeRequestUrl != null) result.changeRequestUrl = changeRequestUrl;
    return result;
  }

  PurgeAgentContentResponse._();

  factory PurgeAgentContentResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      PurgeAgentContentResponse()..mergeFromBuffer(data, registry);
  factory PurgeAgentContentResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      PurgeAgentContentResponse()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'PurgeAgentContentResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: PurgeAgentContentResponse.$_createMessage)
    ..aOM<WorkspaceRepoBinding>(1, _omitFieldNames ? '' : 'binding',
        subBuilder: WorkspaceRepoBinding.$_createMessage)
    ..aOS(2, _omitFieldNames ? '' : 'commitSha')
    ..pPS(3, _omitFieldNames ? '' : 'validationErrors')
    ..aOS(4, _omitFieldNames ? '' : 'changeRequestUrl')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  PurgeAgentContentResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  PurgeAgentContentResponse copyWith(
          void Function(PurgeAgentContentResponse) updates) =>
      super.copyWith((message) => updates(message as PurgeAgentContentResponse))
          as PurgeAgentContentResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use PurgeAgentContentResponse() / PurgeAgentContentResponse.new instead')
  static PurgeAgentContentResponse create() => PurgeAgentContentResponse._();
  static $pb.GeneratedMessage $_createMessage() =>
      PurgeAgentContentResponse._();
  @$core.override
  PurgeAgentContentResponse createEmptyInstance() =>
      PurgeAgentContentResponse._();
  @$core.pragma('dart2js:noInline')
  static PurgeAgentContentResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<PurgeAgentContentResponse>(
          PurgeAgentContentResponse.$_createMessage);
  static PurgeAgentContentResponse? _defaultInstance;

  @$pb.TagNumber(1)
  WorkspaceRepoBinding get binding => $_getN(0);
  @$pb.TagNumber(1)
  set binding(WorkspaceRepoBinding value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasBinding() => $_has(0);
  @$pb.TagNumber(1)
  void clearBinding() => $_clearField(1);
  @$pb.TagNumber(1)
  WorkspaceRepoBinding ensureBinding() => $_ensure(0);

  /// SHA of the purge commit. Empty when the agent had no managed content.
  @$pb.TagNumber(2)
  $core.String get commitSha => $_getSZ(1);
  @$pb.TagNumber(2)
  set commitSha($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasCommitSha() => $_has(1);
  @$pb.TagNumber(2)
  void clearCommitSha() => $_clearField(2);

  /// Per-agent validation errors that prevented the purge (empty on success).
  @$pb.TagNumber(3)
  $pb.PbList<$core.String> get validationErrors => $_getList(2);

  /// URL of the created Pull Request or Merge Request (only set when the
  /// binding write mode is CHANGE_REQUEST).
  @$pb.TagNumber(4)
  $core.String get changeRequestUrl => $_getSZ(3);
  @$pb.TagNumber(4)
  set changeRequestUrl($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasChangeRequestUrl() => $_has(3);
  @$pb.TagNumber(4)
  void clearChangeRequestUrl() => $_clearField(4);
}

/// RepoCacheEntry is one node in the cached repository tree.
class RepoCacheEntry extends $pb.GeneratedMessage {
  factory RepoCacheEntry({
    $core.String? path,
    RepoCacheEntryKind? kind,
    $fixnum.Int64? size,
    $core.String? contentHash,
    $core.bool? claimed,
  }) {
    final result = RepoCacheEntry._();
    if (path != null) result.path = path;
    if (kind != null) result.kind = kind;
    if (size != null) result.size = size;
    if (contentHash != null) result.contentHash = contentHash;
    if (claimed != null) result.claimed = claimed;
    return result;
  }

  RepoCacheEntry._();

  factory RepoCacheEntry.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      RepoCacheEntry()..mergeFromBuffer(data, registry);
  factory RepoCacheEntry.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      RepoCacheEntry()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'RepoCacheEntry',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: RepoCacheEntry.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'path')
    ..aE<RepoCacheEntryKind>(2, _omitFieldNames ? '' : 'kind',
        enumValues: RepoCacheEntryKind.values)
    ..aInt64(3, _omitFieldNames ? '' : 'size')
    ..aOS(4, _omitFieldNames ? '' : 'contentHash')
    ..aOB(5, _omitFieldNames ? '' : 'claimed')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  RepoCacheEntry clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  RepoCacheEntry copyWith(void Function(RepoCacheEntry) updates) =>
      super.copyWith((message) => updates(message as RepoCacheEntry))
          as RepoCacheEntry;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use RepoCacheEntry() / RepoCacheEntry.new instead')
  static RepoCacheEntry create() => RepoCacheEntry._();
  static $pb.GeneratedMessage $_createMessage() => RepoCacheEntry._();
  @$core.override
  RepoCacheEntry createEmptyInstance() => RepoCacheEntry._();
  @$core.pragma('dart2js:noInline')
  static RepoCacheEntry getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<RepoCacheEntry>(
          RepoCacheEntry.$_createMessage);
  static RepoCacheEntry? _defaultInstance;

  /// @gotags: json:"path,omitempty" yaml:"path,omitempty" bson:"path,omitempty"
  @$pb.TagNumber(1)
  $core.String get path => $_getSZ(0);
  @$pb.TagNumber(1)
  set path($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasPath() => $_has(0);
  @$pb.TagNumber(1)
  void clearPath() => $_clearField(1);

  /// @gotags: json:"kind,omitempty" yaml:"kind,omitempty" bson:"kind,omitempty"
  @$pb.TagNumber(2)
  RepoCacheEntryKind get kind => $_getN(1);
  @$pb.TagNumber(2)
  set kind(RepoCacheEntryKind value) => $_setField(2, value);
  @$pb.TagNumber(2)
  $core.bool hasKind() => $_has(1);
  @$pb.TagNumber(2)
  void clearKind() => $_clearField(2);

  /// File size in bytes (zero for directories).
  /// @gotags: json:"size,omitempty" yaml:"size,omitempty" bson:"size,omitempty"
  @$pb.TagNumber(3)
  $fixnum.Int64 get size => $_getI64(2);
  @$pb.TagNumber(3)
  set size($fixnum.Int64 value) => $_setInt64(2, value);
  @$pb.TagNumber(3)
  $core.bool hasSize() => $_has(2);
  @$pb.TagNumber(3)
  void clearSize() => $_clearField(3);

  /// Content hash (SHA from the provider) for deduplication and change detection.
  /// @gotags: json:"content_hash,omitempty" yaml:"content_hash,omitempty" bson:"content_hash,omitempty"
  @$pb.TagNumber(4)
  $core.String get contentHash => $_getSZ(3);
  @$pb.TagNumber(4)
  set contentHash($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasContentHash() => $_has(3);
  @$pb.TagNumber(4)
  void clearContentHash() => $_clearField(4);

  /// Whether this entry maps to an Agent directory known to the workspace.
  /// @gotags: json:"claimed,omitempty" yaml:"claimed,omitempty" bson:"claimed,omitempty"
  @$pb.TagNumber(5)
  $core.bool get claimed => $_getBF(4);
  @$pb.TagNumber(5)
  set claimed($core.bool value) => $_setBool(4, value);
  @$pb.TagNumber(5)
  $core.bool hasClaimed() => $_has(4);
  @$pb.TagNumber(5)
  void clearClaimed() => $_clearField(5);
}

/// WorkspaceRepoBindingService manages the current workspace's repository
/// binding. All RPCs derive the workspace from the X-Workspace-ID header.
/// Get is available to every workspace member; the mutating RPCs require the
/// caller to be a workspace owner or admin (global admins bypass).
class WorkspaceRepoBindingServiceApi {
  final $pb.RpcClient _client;

  WorkspaceRepoBindingServiceApi(this._client);

  /// GetWorkspaceRepoBinding returns the binding (unset when none exists)
  /// plus any overlapping bindings in other workspaces. Never returns
  /// credential material.
  $async.Future<GetWorkspaceRepoBindingResponse> getWorkspaceRepoBinding(
          $pb.ClientContext? ctx, GetWorkspaceRepoBindingRequest request) =>
      _client.invoke<GetWorkspaceRepoBindingResponse>(
          ctx,
          'WorkspaceRepoBindingService',
          'GetWorkspaceRepoBinding',
          request,
          GetWorkspaceRepoBindingResponse());

  /// PutWorkspaceRepoBinding creates or replaces the workspace's binding
  /// configuration. Server-owned fields (credential, status, timestamps) are
  /// ignored on input. Every Put resets status to UNVALIDATED: each binding
  /// field (including write mode) changes what validation must prove.
  $async.Future<PutWorkspaceRepoBindingResponse> putWorkspaceRepoBinding(
          $pb.ClientContext? ctx, PutWorkspaceRepoBindingRequest request) =>
      _client.invoke<PutWorkspaceRepoBindingResponse>(
          ctx,
          'WorkspaceRepoBindingService',
          'PutWorkspaceRepoBinding',
          request,
          PutWorkspaceRepoBindingResponse());

  /// DeleteWorkspaceRepoBinding safely detaches the workspace from Git-owned
  /// Agent Content. It first materializes the Active Revision snapshot back into
  /// database-managed Agent fields and reloads the runtime, then removes the
  /// binding and its encrypted PAT. It never modifies remote Git content, and
  /// refuses to proceed when there is no valid active snapshot unless the caller
  /// selects a supported recovery path. Requires owner/admin role.
  $async.Future<DeleteWorkspaceRepoBindingResponse> deleteWorkspaceRepoBinding(
          $pb.ClientContext? ctx, DeleteWorkspaceRepoBindingRequest request) =>
      _client.invoke<DeleteWorkspaceRepoBindingResponse>(
          ctx,
          'WorkspaceRepoBindingService',
          'DeleteWorkspaceRepoBinding',
          request,
          DeleteWorkspaceRepoBindingResponse());

  /// OnboardWorkspaceRepository performs the one-time reconciliation of a freshly
  /// bound workspace's Agent Content with the repository (issue #219).
  /// EXPORT_CURRENT commits existing database-managed content to Git;
  /// IMPORT_REPOSITORY adopts repository content for matching Agent IDs. In both
  /// modes the workspace only switches to Git-owned content after the
  /// commit/import, read-back, validation, Effective Agent construction, and
  /// Active Revision publication all succeed. Requires owner/admin role.
  $async.Future<OnboardWorkspaceRepositoryResponse> onboardWorkspaceRepository(
          $pb.ClientContext? ctx, OnboardWorkspaceRepositoryRequest request) =>
      _client.invoke<OnboardWorkspaceRepositoryResponse>(
          ctx,
          'WorkspaceRepoBindingService',
          'OnboardWorkspaceRepository',
          request,
          OnboardWorkspaceRepositoryResponse());

  /// SetWorkspaceRepoBindingCredential sets or replaces the binding's PAT.
  /// The PAT is write-only: it is encrypted before storage and never returned.
  $async.Future<SetWorkspaceRepoBindingCredentialResponse>
      setWorkspaceRepoBindingCredential($pb.ClientContext? ctx,
              SetWorkspaceRepoBindingCredentialRequest request) =>
          _client.invoke<SetWorkspaceRepoBindingCredentialResponse>(
              ctx,
              'WorkspaceRepoBindingService',
              'SetWorkspaceRepoBindingCredential',
              request,
              SetWorkspaceRepoBindingCredentialResponse());

  /// ValidateWorkspaceRepoBinding probes the repository with the stored
  /// credential (visibility/read, branch existence, write capability, PR/MR
  /// capability as required by the write mode), persists the outcome on the
  /// binding, and returns it.
  $async.Future<ValidateWorkspaceRepoBindingResponse>
      validateWorkspaceRepoBinding($pb.ClientContext? ctx,
              ValidateWorkspaceRepoBindingRequest request) =>
          _client.invoke<ValidateWorkspaceRepoBindingResponse>(
              ctx,
              'WorkspaceRepoBindingService',
              'ValidateWorkspaceRepoBinding',
              request,
              ValidateWorkspaceRepoBindingResponse());

  /// SyncWorkspaceRepository reads the bound repository's managed tree and
  /// Markdown blobs into the workspace-scoped DB cache. Requires owner or
  /// admin role. The sync is idempotent when the remote HEAD has not changed.
  $async.Future<SyncWorkspaceRepositoryResponse> syncWorkspaceRepository(
          $pb.ClientContext? ctx, SyncWorkspaceRepositoryRequest request) =>
      _client.invoke<SyncWorkspaceRepositoryResponse>(
          ctx,
          'WorkspaceRepoBindingService',
          'SyncWorkspaceRepository',
          request,
          SyncWorkspaceRepositoryResponse());

  /// GetRepositorySyncStatus returns the binding with its current sync state
  /// without triggering a new synchronization.
  $async.Future<GetRepositorySyncStatusResponse> getRepositorySyncStatus(
          $pb.ClientContext? ctx, GetRepositorySyncStatusRequest request) =>
      _client.invoke<GetRepositorySyncStatusResponse>(
          ctx,
          'WorkspaceRepoBindingService',
          'GetRepositorySyncStatus',
          request,
          GetRepositorySyncStatusResponse());

  /// ListRepositoryEntries returns cached directory entries for a path within
  /// the managed tree. Reads only from the DB cache; never issues Git requests.
  $async.Future<ListRepositoryEntriesResponse> listRepositoryEntries(
          $pb.ClientContext? ctx, ListRepositoryEntriesRequest request) =>
      _client.invoke<ListRepositoryEntriesResponse>(
          ctx,
          'WorkspaceRepoBindingService',
          'ListRepositoryEntries',
          request,
          ListRepositoryEntriesResponse());

  /// GetRepositoryFile returns the cached content of a single Markdown file.
  /// Reads only from the DB cache; never issues Git requests.
  $async.Future<GetRepositoryFileResponse> getRepositoryFile(
          $pb.ClientContext? ctx, GetRepositoryFileRequest request) =>
      _client.invoke<GetRepositoryFileResponse>(
          ctx,
          'WorkspaceRepoBindingService',
          'GetRepositoryFile',
          request,
          GetRepositoryFileResponse());

  /// PublishWorkspaceRepository validates the current observed revision's
  /// agent content and, on success, atomically advances the Active Revision
  /// and reloads the runner. Fails gracefully when validation errors are
  /// present; the binding reports per-agent errors.
  $async.Future<PublishWorkspaceRepositoryResponse> publishWorkspaceRepository(
          $pb.ClientContext? ctx, PublishWorkspaceRepositoryRequest request) =>
      _client.invoke<PublishWorkspaceRepositoryResponse>(
          ctx,
          'WorkspaceRepoBindingService',
          'PublishWorkspaceRepository',
          request,
          PublishWorkspaceRepositoryResponse());

  /// ConfigureWebhookSecret generates (or replaces) a webhook secret for the
  /// binding and returns the callback URL and the secret (shown only once).
  $async.Future<ConfigureWebhookSecretResponse> configureWebhookSecret(
          $pb.ClientContext? ctx, ConfigureWebhookSecretRequest request) =>
      _client.invoke<ConfigureWebhookSecretResponse>(
          ctx,
          'WorkspaceRepoBindingService',
          'ConfigureWebhookSecret',
          request,
          ConfigureWebhookSecretResponse());

  /// AcceptRepositoryBaseline accepts a diverged (force-pushed) HEAD as the
  /// new baseline, re-syncs and publishes. Requires owner/admin role.
  $async.Future<AcceptRepositoryBaselineResponse> acceptRepositoryBaseline(
          $pb.ClientContext? ctx, AcceptRepositoryBaselineRequest request) =>
      _client.invoke<AcceptRepositoryBaselineResponse>(
          ctx,
          'WorkspaceRepoBindingService',
          'AcceptRepositoryBaseline',
          request,
          AcceptRepositoryBaselineResponse());

  /// CommitAgentContent applies a changeset of PUT and DELETE operations to
  /// managed Agent Content paths and produces a single Git commit. In
  /// DIRECT_COMMIT mode the commit lands on the bound branch; in
  /// CHANGE_REQUEST mode a PR/MR is opened instead. Content is validated
  /// before the commit is created. Requires owner/admin role.
  $async.Future<CommitAgentContentResponse> commitAgentContent(
          $pb.ClientContext? ctx, CommitAgentContentRequest request) =>
      _client.invoke<CommitAgentContentResponse>(
          ctx,
          'WorkspaceRepoBindingService',
          'CommitAgentContent',
          request,
          CommitAgentContentResponse());

  /// RollbackAgentContent restores managed Agent Content from a previously
  /// published revision in a new Git commit. The target revision's content
  /// is read from Git, validated, and committed to the bound branch (or a
  /// PR/MR in CHANGE_REQUEST mode). Requires owner/admin role.
  $async.Future<RollbackAgentContentResponse> rollbackAgentContent(
          $pb.ClientContext? ctx, RollbackAgentContentRequest request) =>
      _client.invoke<RollbackAgentContentResponse>(
          ctx,
          'WorkspaceRepoBindingService',
          'RollbackAgentContent',
          request,
          RollbackAgentContentResponse());

  /// PurgeAgentContent permanently deletes an Agent's managed Content files from
  /// the repository in a new Git commit. It is a separate owner/admin action
  /// from DeleteAgent (which only tombstones the DB entity and retains content).
  /// It refuses when the effective path is still claimed by any non-deleted
  /// Agent in this workspace or in any workspace that shares the path via an
  /// overlapping binding. Requires owner/admin role.
  $async.Future<PurgeAgentContentResponse> purgeAgentContent(
          $pb.ClientContext? ctx, PurgeAgentContentRequest request) =>
      _client.invoke<PurgeAgentContentResponse>(
          ctx,
          'WorkspaceRepoBindingService',
          'PurgeAgentContent',
          request,
          PurgeAgentContentResponse());
}

const $core.bool _omitFieldNames =
    $core.bool.fromEnvironment('protobuf.omit_field_names');
const $core.bool _omitMessageNames =
    $core.bool.fromEnvironment('protobuf.omit_message_names');
