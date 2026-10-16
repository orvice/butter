// This is a generated file - do not edit.
//
// Generated from agents/v1/linear.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names, prefer_relative_imports

import 'dart:core' as $core;

import 'package:protobuf/protobuf.dart' as $pb;

/// LinearAppCredentialState reports whether both App secrets are stored
/// without revealing either.
class LinearAppCredentialState extends $pb.ProtobufEnum {
  static const LinearAppCredentialState
      LINEAR_APP_CREDENTIAL_STATE_UNSPECIFIED = LinearAppCredentialState._(
          0, _omitEnumNames ? '' : 'LINEAR_APP_CREDENTIAL_STATE_UNSPECIFIED');

  /// At least one of the client secret and webhook signing secret is missing.
  static const LinearAppCredentialState LINEAR_APP_CREDENTIAL_STATE_INCOMPLETE =
      LinearAppCredentialState._(
          1, _omitEnumNames ? '' : 'LINEAR_APP_CREDENTIAL_STATE_INCOMPLETE');

  /// Both secrets are stored.
  static const LinearAppCredentialState LINEAR_APP_CREDENTIAL_STATE_COMPLETE =
      LinearAppCredentialState._(
          2, _omitEnumNames ? '' : 'LINEAR_APP_CREDENTIAL_STATE_COMPLETE');

  static const $core.List<LinearAppCredentialState> values =
      <LinearAppCredentialState>[
    LINEAR_APP_CREDENTIAL_STATE_UNSPECIFIED,
    LINEAR_APP_CREDENTIAL_STATE_INCOMPLETE,
    LINEAR_APP_CREDENTIAL_STATE_COMPLETE,
  ];

  static final $core.List<LinearAppCredentialState?> _byValue =
      $pb.ProtobufEnum.$_initByValueList(values, 2);
  static LinearAppCredentialState? valueOf($core.int value) =>
      value < 0 || value >= _byValue.length ? null : _byValue[value];

  const LinearAppCredentialState._(super.value, super.name);
}

/// LinearInstallationCredentialState reports whether an installation's token
/// is usable without revealing it.
class LinearInstallationCredentialState extends $pb.ProtobufEnum {
  static const LinearInstallationCredentialState
      LINEAR_INSTALLATION_CREDENTIAL_STATE_UNSPECIFIED =
      LinearInstallationCredentialState._(
          0,
          _omitEnumNames
              ? ''
              : 'LINEAR_INSTALLATION_CREDENTIAL_STATE_UNSPECIFIED');

  /// The stored token worked, or was refreshed, on last use.
  static const LinearInstallationCredentialState
      LINEAR_INSTALLATION_CREDENTIAL_STATE_VALID =
      LinearInstallationCredentialState._(1,
          _omitEnumNames ? '' : 'LINEAR_INSTALLATION_CREDENTIAL_STATE_VALID');

  /// Linear revoked the token or refused to refresh it; install again.
  static const LinearInstallationCredentialState
      LINEAR_INSTALLATION_CREDENTIAL_STATE_NEEDS_REINSTALL =
      LinearInstallationCredentialState._(
          2,
          _omitEnumNames
              ? ''
              : 'LINEAR_INSTALLATION_CREDENTIAL_STATE_NEEDS_REINSTALL');

  static const $core.List<LinearInstallationCredentialState> values =
      <LinearInstallationCredentialState>[
    LINEAR_INSTALLATION_CREDENTIAL_STATE_UNSPECIFIED,
    LINEAR_INSTALLATION_CREDENTIAL_STATE_VALID,
    LINEAR_INSTALLATION_CREDENTIAL_STATE_NEEDS_REINSTALL,
  ];

  static final $core.List<LinearInstallationCredentialState?> _byValue =
      $pb.ProtobufEnum.$_initByValueList(values, 2);
  static LinearInstallationCredentialState? valueOf($core.int value) =>
      value < 0 || value >= _byValue.length ? null : _byValue[value];

  const LinearInstallationCredentialState._(super.value, super.name);
}

class LinearProcessingStatus extends $pb.ProtobufEnum {
  static const LinearProcessingStatus LINEAR_PROCESSING_STATUS_UNSPECIFIED =
      LinearProcessingStatus._(
          0, _omitEnumNames ? '' : 'LINEAR_PROCESSING_STATUS_UNSPECIFIED');

  /// A follow-up waiting behind the session's current turn.
  static const LinearProcessingStatus LINEAR_PROCESSING_STATUS_QUEUED =
      LinearProcessingStatus._(
          1, _omitEnumNames ? '' : 'LINEAR_PROCESSING_STATUS_QUEUED');

  /// Claimed from the queue; no Agent work started.
  static const LinearProcessingStatus LINEAR_PROCESSING_STATUS_RECEIVED =
      LinearProcessingStatus._(
          2, _omitEnumNames ? '' : 'LINEAR_PROCESSING_STATUS_RECEIVED');

  /// Agent work may have started. A crash from here is not safely retryable.
  static const LinearProcessingStatus LINEAR_PROCESSING_STATUS_PROCESSING =
      LinearProcessingStatus._(
          3, _omitEnumNames ? '' : 'LINEAR_PROCESSING_STATUS_PROCESSING');

  /// The reply is persisted; only posting it remains.
  static const LinearProcessingStatus
      LINEAR_PROCESSING_STATUS_READY_TO_DELIVER = LinearProcessingStatus._(
          4, _omitEnumNames ? '' : 'LINEAR_PROCESSING_STATUS_READY_TO_DELIVER');

  /// Answered.
  static const LinearProcessingStatus LINEAR_PROCESSING_STATUS_SUCCEEDED =
      LinearProcessingStatus._(
          5, _omitEnumNames ? '' : 'LINEAR_PROCESSING_STATUS_SUCCEEDED');

  /// Failed before Agent work started, or the post failed. Safe to retry or
  /// resend.
  static const LinearProcessingStatus LINEAR_PROCESSING_STATUS_FAILED =
      LinearProcessingStatus._(
          6, _omitEnumNames ? '' : 'LINEAR_PROCESSING_STATUS_FAILED');

  /// Failed after Agent work may have started. Never retried automatically:
  /// a rerun could repeat tool side effects.
  static const LinearProcessingStatus
      LINEAR_PROCESSING_STATUS_FAILED_UNCERTAIN = LinearProcessingStatus._(
          7, _omitEnumNames ? '' : 'LINEAR_PROCESSING_STATUS_FAILED_UNCERTAIN');

  /// Stopped by the user, or discarded by a stop while queued.
  static const LinearProcessingStatus LINEAR_PROCESSING_STATUS_CANCELLED =
      LinearProcessingStatus._(
          8, _omitEnumNames ? '' : 'LINEAR_PROCESSING_STATUS_CANCELLED');

  static const $core.List<LinearProcessingStatus> values =
      <LinearProcessingStatus>[
    LINEAR_PROCESSING_STATUS_UNSPECIFIED,
    LINEAR_PROCESSING_STATUS_QUEUED,
    LINEAR_PROCESSING_STATUS_RECEIVED,
    LINEAR_PROCESSING_STATUS_PROCESSING,
    LINEAR_PROCESSING_STATUS_READY_TO_DELIVER,
    LINEAR_PROCESSING_STATUS_SUCCEEDED,
    LINEAR_PROCESSING_STATUS_FAILED,
    LINEAR_PROCESSING_STATUS_FAILED_UNCERTAIN,
    LINEAR_PROCESSING_STATUS_CANCELLED,
  ];

  static final $core.List<LinearProcessingStatus?> _byValue =
      $pb.ProtobufEnum.$_initByValueList(values, 8);
  static LinearProcessingStatus? valueOf($core.int value) =>
      value < 0 || value >= _byValue.length ? null : _byValue[value];

  const LinearProcessingStatus._(super.value, super.name);
}

const $core.bool _omitEnumNames =
    $core.bool.fromEnvironment('protobuf.omit_enum_names');
