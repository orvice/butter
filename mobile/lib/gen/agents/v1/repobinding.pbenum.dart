// This is a generated file - do not edit.
//
// Generated from agents/v1/repobinding.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names, prefer_relative_imports

import 'dart:core' as $core;

import 'package:protobuf/protobuf.dart' as $pb;

/// RepoBindingWriteMode selects how Butter lands changes on the repository.
class RepoBindingWriteMode extends $pb.ProtobufEnum {
  static const RepoBindingWriteMode REPO_BINDING_WRITE_MODE_UNSPECIFIED =
      RepoBindingWriteMode._(
          0, _omitEnumNames ? '' : 'REPO_BINDING_WRITE_MODE_UNSPECIFIED');

  /// Commit directly to the bound branch (default).
  static const RepoBindingWriteMode REPO_BINDING_WRITE_MODE_DIRECT_COMMIT =
      RepoBindingWriteMode._(
          1, _omitEnumNames ? '' : 'REPO_BINDING_WRITE_MODE_DIRECT_COMMIT');

  /// Open a Pull Request (GitHub) or Merge Request (GitLab) instead.
  static const RepoBindingWriteMode REPO_BINDING_WRITE_MODE_CHANGE_REQUEST =
      RepoBindingWriteMode._(
          2, _omitEnumNames ? '' : 'REPO_BINDING_WRITE_MODE_CHANGE_REQUEST');

  static const $core.List<RepoBindingWriteMode> values = <RepoBindingWriteMode>[
    REPO_BINDING_WRITE_MODE_UNSPECIFIED,
    REPO_BINDING_WRITE_MODE_DIRECT_COMMIT,
    REPO_BINDING_WRITE_MODE_CHANGE_REQUEST,
  ];

  static final $core.List<RepoBindingWriteMode?> _byValue =
      $pb.ProtobufEnum.$_initByValueList(values, 2);
  static RepoBindingWriteMode? valueOf($core.int value) =>
      value < 0 || value >= _byValue.length ? null : _byValue[value];

  const RepoBindingWriteMode._(super.value, super.name);
}

/// RepoBindingOnboardingMode selects how a freshly bound workspace reconciles
/// its existing database-managed Agent Content with the bound repository during
/// the one-time OnboardWorkspaceRepository call (issue #219).
class RepoBindingOnboardingMode extends $pb.ProtobufEnum {
  static const RepoBindingOnboardingMode
      REPO_BINDING_ONBOARDING_MODE_UNSPECIFIED = RepoBindingOnboardingMode._(
          0, _omitEnumNames ? '' : 'REPO_BINDING_ONBOARDING_MODE_UNSPECIFIED');

  /// EXPORT_CURRENT: write the workspace's existing database-managed Agent
  /// Content to the repository as a single validated commit, then read it back,
  /// validate, and publish the Active Revision.
  static const RepoBindingOnboardingMode
      REPO_BINDING_ONBOARDING_MODE_EXPORT_CURRENT = RepoBindingOnboardingMode._(
          1,
          _omitEnumNames ? '' : 'REPO_BINDING_ONBOARDING_MODE_EXPORT_CURRENT');

  /// IMPORT_REPOSITORY: adopt the repository's existing Agent Content, but only
  /// for directories that match an existing workspace Agent ID. Unknown
  /// directories remain unclaimed; no bidirectional merge is performed.
  static const RepoBindingOnboardingMode
      REPO_BINDING_ONBOARDING_MODE_IMPORT_REPOSITORY =
      RepoBindingOnboardingMode._(
          2,
          _omitEnumNames
              ? ''
              : 'REPO_BINDING_ONBOARDING_MODE_IMPORT_REPOSITORY');

  static const $core.List<RepoBindingOnboardingMode> values =
      <RepoBindingOnboardingMode>[
    REPO_BINDING_ONBOARDING_MODE_UNSPECIFIED,
    REPO_BINDING_ONBOARDING_MODE_EXPORT_CURRENT,
    REPO_BINDING_ONBOARDING_MODE_IMPORT_REPOSITORY,
  ];

  static final $core.List<RepoBindingOnboardingMode?> _byValue =
      $pb.ProtobufEnum.$_initByValueList(values, 2);
  static RepoBindingOnboardingMode? valueOf($core.int value) =>
      value < 0 || value >= _byValue.length ? null : _byValue[value];

  const RepoBindingOnboardingMode._(super.value, super.name);
}

/// RepoBindingDetachRecovery selects how DeleteWorkspaceRepoBinding proceeds
/// when there is no valid Active Agent Content snapshot to materialize back into
/// database-managed fields (issue #219). The default refuses the unbind so a
/// missing snapshot can never silently drop an Agent's live content.
class RepoBindingDetachRecovery extends $pb.ProtobufEnum {
  /// Refuse to unbind unless a valid Active Revision snapshot exists to
  /// materialize. This is the safe default.
  static const RepoBindingDetachRecovery
      REPO_BINDING_DETACH_RECOVERY_UNSPECIFIED = RepoBindingDetachRecovery._(
          0, _omitEnumNames ? '' : 'REPO_BINDING_DETACH_RECOVERY_UNSPECIFIED');

  /// Keep the current database-managed Agent Content as-is and remove the
  /// binding without materializing from Git. Chosen when the DB fields are
  /// trusted despite the snapshot being absent or invalid.
  static const RepoBindingDetachRecovery
      REPO_BINDING_DETACH_RECOVERY_KEEP_DATABASE = RepoBindingDetachRecovery._(
          1,
          _omitEnumNames ? '' : 'REPO_BINDING_DETACH_RECOVERY_KEEP_DATABASE');

  static const $core.List<RepoBindingDetachRecovery> values =
      <RepoBindingDetachRecovery>[
    REPO_BINDING_DETACH_RECOVERY_UNSPECIFIED,
    REPO_BINDING_DETACH_RECOVERY_KEEP_DATABASE,
  ];

  static final $core.List<RepoBindingDetachRecovery?> _byValue =
      $pb.ProtobufEnum.$_initByValueList(values, 1);
  static RepoBindingDetachRecovery? valueOf($core.int value) =>
      value < 0 || value >= _byValue.length ? null : _byValue[value];

  const RepoBindingDetachRecovery._(super.value, super.name);
}

/// RepoBindingConnectionState summarizes the last validation run.
class RepoBindingConnectionState extends $pb.ProtobufEnum {
  static const RepoBindingConnectionState
      REPO_BINDING_CONNECTION_STATE_UNSPECIFIED = RepoBindingConnectionState._(
          0, _omitEnumNames ? '' : 'REPO_BINDING_CONNECTION_STATE_UNSPECIFIED');

  /// The current configuration has not been validated yet (new binding, or
  /// the binding/credential changed since the last validation).
  static const RepoBindingConnectionState
      REPO_BINDING_CONNECTION_STATE_UNVALIDATED = RepoBindingConnectionState._(
          1, _omitEnumNames ? '' : 'REPO_BINDING_CONNECTION_STATE_UNVALIDATED');

  /// All checks required by the selected write mode passed.
  static const RepoBindingConnectionState REPO_BINDING_CONNECTION_STATE_OK =
      RepoBindingConnectionState._(
          2, _omitEnumNames ? '' : 'REPO_BINDING_CONNECTION_STATE_OK');

  /// At least one required check failed.
  static const RepoBindingConnectionState REPO_BINDING_CONNECTION_STATE_FAILED =
      RepoBindingConnectionState._(
          3, _omitEnumNames ? '' : 'REPO_BINDING_CONNECTION_STATE_FAILED');

  /// PAT or provider failure; active Agents continue from last-known-good.
  static const RepoBindingConnectionState
      REPO_BINDING_CONNECTION_STATE_DEGRADED = RepoBindingConnectionState._(
          4, _omitEnumNames ? '' : 'REPO_BINDING_CONNECTION_STATE_DEGRADED');

  /// Non-fast-forward history detected; requires owner/admin baseline acceptance.
  static const RepoBindingConnectionState
      REPO_BINDING_CONNECTION_STATE_DIVERGED = RepoBindingConnectionState._(
          5, _omitEnumNames ? '' : 'REPO_BINDING_CONNECTION_STATE_DIVERGED');

  static const $core.List<RepoBindingConnectionState> values =
      <RepoBindingConnectionState>[
    REPO_BINDING_CONNECTION_STATE_UNSPECIFIED,
    REPO_BINDING_CONNECTION_STATE_UNVALIDATED,
    REPO_BINDING_CONNECTION_STATE_OK,
    REPO_BINDING_CONNECTION_STATE_FAILED,
    REPO_BINDING_CONNECTION_STATE_DEGRADED,
    REPO_BINDING_CONNECTION_STATE_DIVERGED,
  ];

  static final $core.List<RepoBindingConnectionState?> _byValue =
      $pb.ProtobufEnum.$_initByValueList(values, 5);
  static RepoBindingConnectionState? valueOf($core.int value) =>
      value < 0 || value >= _byValue.length ? null : _byValue[value];

  const RepoBindingConnectionState._(super.value, super.name);
}

/// ContentFileOperation selects whether a changeset entry creates/updates or
/// removes a managed file.
class ContentFileOperation extends $pb.ProtobufEnum {
  static const ContentFileOperation CONTENT_FILE_OPERATION_UNSPECIFIED =
      ContentFileOperation._(
          0, _omitEnumNames ? '' : 'CONTENT_FILE_OPERATION_UNSPECIFIED');

  /// Create or replace the file at the given path.
  static const ContentFileOperation CONTENT_FILE_OPERATION_PUT =
      ContentFileOperation._(
          1, _omitEnumNames ? '' : 'CONTENT_FILE_OPERATION_PUT');

  /// Remove the file at the given path.
  static const ContentFileOperation CONTENT_FILE_OPERATION_DELETE =
      ContentFileOperation._(
          2, _omitEnumNames ? '' : 'CONTENT_FILE_OPERATION_DELETE');

  static const $core.List<ContentFileOperation> values = <ContentFileOperation>[
    CONTENT_FILE_OPERATION_UNSPECIFIED,
    CONTENT_FILE_OPERATION_PUT,
    CONTENT_FILE_OPERATION_DELETE,
  ];

  static final $core.List<ContentFileOperation?> _byValue =
      $pb.ProtobufEnum.$_initByValueList(values, 2);
  static ContentFileOperation? valueOf($core.int value) =>
      value < 0 || value >= _byValue.length ? null : _byValue[value];

  const ContentFileOperation._(super.value, super.name);
}

/// RepoCacheEntryKind classifies a cached tree entry.
class RepoCacheEntryKind extends $pb.ProtobufEnum {
  static const RepoCacheEntryKind REPO_CACHE_ENTRY_KIND_UNSPECIFIED =
      RepoCacheEntryKind._(
          0, _omitEnumNames ? '' : 'REPO_CACHE_ENTRY_KIND_UNSPECIFIED');
  static const RepoCacheEntryKind REPO_CACHE_ENTRY_KIND_FILE =
      RepoCacheEntryKind._(
          1, _omitEnumNames ? '' : 'REPO_CACHE_ENTRY_KIND_FILE');
  static const RepoCacheEntryKind REPO_CACHE_ENTRY_KIND_DIRECTORY =
      RepoCacheEntryKind._(
          2, _omitEnumNames ? '' : 'REPO_CACHE_ENTRY_KIND_DIRECTORY');
  static const RepoCacheEntryKind REPO_CACHE_ENTRY_KIND_SYMLINK =
      RepoCacheEntryKind._(
          3, _omitEnumNames ? '' : 'REPO_CACHE_ENTRY_KIND_SYMLINK');
  static const RepoCacheEntryKind REPO_CACHE_ENTRY_KIND_SUBMODULE =
      RepoCacheEntryKind._(
          4, _omitEnumNames ? '' : 'REPO_CACHE_ENTRY_KIND_SUBMODULE');

  static const $core.List<RepoCacheEntryKind> values = <RepoCacheEntryKind>[
    REPO_CACHE_ENTRY_KIND_UNSPECIFIED,
    REPO_CACHE_ENTRY_KIND_FILE,
    REPO_CACHE_ENTRY_KIND_DIRECTORY,
    REPO_CACHE_ENTRY_KIND_SYMLINK,
    REPO_CACHE_ENTRY_KIND_SUBMODULE,
  ];

  static final $core.List<RepoCacheEntryKind?> _byValue =
      $pb.ProtobufEnum.$_initByValueList(values, 4);
  static RepoCacheEntryKind? valueOf($core.int value) =>
      value < 0 || value >= _byValue.length ? null : _byValue[value];

  const RepoCacheEntryKind._(super.value, super.name);
}

const $core.bool _omitEnumNames =
    $core.bool.fromEnvironment('protobuf.omit_enum_names');
