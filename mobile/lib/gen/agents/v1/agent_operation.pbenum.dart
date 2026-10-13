// This is a generated file - do not edit.
//
// Generated from agents/v1/agent_operation.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names, prefer_relative_imports

import 'dart:core' as $core;

import 'package:protobuf/protobuf.dart' as $pb;

/// AgentOperationType selects which lifecycle Saga a record drives.
class AgentOperationType extends $pb.ProtobufEnum {
  static const AgentOperationType AGENT_OPERATION_TYPE_UNSPECIFIED =
      AgentOperationType._(
          0, _omitEnumNames ? '' : 'AGENT_OPERATION_TYPE_UNSPECIFIED');
  static const AgentOperationType AGENT_OPERATION_TYPE_CREATE =
      AgentOperationType._(
          1, _omitEnumNames ? '' : 'AGENT_OPERATION_TYPE_CREATE');
  static const AgentOperationType AGENT_OPERATION_TYPE_UPDATE_CONFIGURATION =
      AgentOperationType._(
          2, _omitEnumNames ? '' : 'AGENT_OPERATION_TYPE_UPDATE_CONFIGURATION');
  static const AgentOperationType AGENT_OPERATION_TYPE_DELETE =
      AgentOperationType._(
          3, _omitEnumNames ? '' : 'AGENT_OPERATION_TYPE_DELETE');
  static const AgentOperationType AGENT_OPERATION_TYPE_RESTORE =
      AgentOperationType._(
          4, _omitEnumNames ? '' : 'AGENT_OPERATION_TYPE_RESTORE');

  static const $core.List<AgentOperationType> values = <AgentOperationType>[
    AGENT_OPERATION_TYPE_UNSPECIFIED,
    AGENT_OPERATION_TYPE_CREATE,
    AGENT_OPERATION_TYPE_UPDATE_CONFIGURATION,
    AGENT_OPERATION_TYPE_DELETE,
    AGENT_OPERATION_TYPE_RESTORE,
  ];

  static final $core.List<AgentOperationType?> _byValue =
      $pb.ProtobufEnum.$_initByValueList(values, 4);
  static AgentOperationType? valueOf($core.int value) =>
      value < 0 || value >= _byValue.length ? null : _byValue[value];

  const AgentOperationType._(super.value, super.name);
}

/// AgentOperationStatus is the overall state of a lifecycle Saga.
class AgentOperationStatus extends $pb.ProtobufEnum {
  static const AgentOperationStatus AGENT_OPERATION_STATUS_UNSPECIFIED =
      AgentOperationStatus._(
          0, _omitEnumNames ? '' : 'AGENT_OPERATION_STATUS_UNSPECIFIED');

  /// Persisted but not yet started.
  static const AgentOperationStatus AGENT_OPERATION_STATUS_PENDING =
      AgentOperationStatus._(
          1, _omitEnumNames ? '' : 'AGENT_OPERATION_STATUS_PENDING');

  /// At least one step is in progress.
  static const AgentOperationStatus AGENT_OPERATION_STATUS_RUNNING =
      AgentOperationStatus._(
          2, _omitEnumNames ? '' : 'AGENT_OPERATION_STATUS_RUNNING');

  /// Every step completed successfully.
  static const AgentOperationStatus AGENT_OPERATION_STATUS_SUCCEEDED =
      AgentOperationStatus._(
          3, _omitEnumNames ? '' : 'AGENT_OPERATION_STATUS_SUCCEEDED');

  /// A step failed; the operation is recoverable via RetryAgentOperation.
  static const AgentOperationStatus AGENT_OPERATION_STATUS_FAILED =
      AgentOperationStatus._(
          4, _omitEnumNames ? '' : 'AGENT_OPERATION_STATUS_FAILED');

  static const $core.List<AgentOperationStatus> values = <AgentOperationStatus>[
    AGENT_OPERATION_STATUS_UNSPECIFIED,
    AGENT_OPERATION_STATUS_PENDING,
    AGENT_OPERATION_STATUS_RUNNING,
    AGENT_OPERATION_STATUS_SUCCEEDED,
    AGENT_OPERATION_STATUS_FAILED,
  ];

  static final $core.List<AgentOperationStatus?> _byValue =
      $pb.ProtobufEnum.$_initByValueList(values, 4);
  static AgentOperationStatus? valueOf($core.int value) =>
      value < 0 || value >= _byValue.length ? null : _byValue[value];

  const AgentOperationStatus._(super.value, super.name);
}

/// AgentOperationStepKind identifies a Saga step.
class AgentOperationStepKind extends $pb.ProtobufEnum {
  static const AgentOperationStepKind AGENT_OPERATION_STEP_KIND_UNSPECIFIED =
      AgentOperationStepKind._(
          0, _omitEnumNames ? '' : 'AGENT_OPERATION_STEP_KIND_UNSPECIFIED');

  /// Create the DB row in PROVISIONING.
  static const AgentOperationStepKind AGENT_OPERATION_STEP_KIND_DB_PROVISION =
      AgentOperationStepKind._(
          1, _omitEnumNames ? '' : 'AGENT_OPERATION_STEP_KIND_DB_PROVISION');

  /// Write Agent Content as a single Git commit.
  static const AgentOperationStepKind AGENT_OPERATION_STEP_KIND_CONTENT_COMMIT =
      AgentOperationStepKind._(
          2, _omitEnumNames ? '' : 'AGENT_OPERATION_STEP_KIND_CONTENT_COMMIT');

  /// Sync the repository and publish the Active Revision.
  static const AgentOperationStepKind AGENT_OPERATION_STEP_KIND_SYNC_PUBLISH =
      AgentOperationStepKind._(
          3, _omitEnumNames ? '' : 'AGENT_OPERATION_STEP_KIND_SYNC_PUBLISH');

  /// Flip lifecycle to ACTIVE and reload the runner.
  static const AgentOperationStepKind AGENT_OPERATION_STEP_KIND_ACTIVATE =
      AgentOperationStepKind._(
          4, _omitEnumNames ? '' : 'AGENT_OPERATION_STEP_KIND_ACTIVATE');

  /// Flip lifecycle to DELETED (tombstone) and reload the runner.
  static const AgentOperationStepKind AGENT_OPERATION_STEP_KIND_TOMBSTONE =
      AgentOperationStepKind._(
          5, _omitEnumNames ? '' : 'AGENT_OPERATION_STEP_KIND_TOMBSTONE');

  /// Flip lifecycle from DELETED back to ACTIVE and reload the runner.
  static const AgentOperationStepKind AGENT_OPERATION_STEP_KIND_RESTORE_DB =
      AgentOperationStepKind._(
          6, _omitEnumNames ? '' : 'AGENT_OPERATION_STEP_KIND_RESTORE_DB');

  /// Apply an operational config patch to the DB (composite save) with
  /// optimistic concurrency, then reload the runner.
  static const AgentOperationStepKind AGENT_OPERATION_STEP_KIND_DB_PATCH =
      AgentOperationStepKind._(
          7, _omitEnumNames ? '' : 'AGENT_OPERATION_STEP_KIND_DB_PATCH');

  static const $core.List<AgentOperationStepKind> values =
      <AgentOperationStepKind>[
    AGENT_OPERATION_STEP_KIND_UNSPECIFIED,
    AGENT_OPERATION_STEP_KIND_DB_PROVISION,
    AGENT_OPERATION_STEP_KIND_CONTENT_COMMIT,
    AGENT_OPERATION_STEP_KIND_SYNC_PUBLISH,
    AGENT_OPERATION_STEP_KIND_ACTIVATE,
    AGENT_OPERATION_STEP_KIND_TOMBSTONE,
    AGENT_OPERATION_STEP_KIND_RESTORE_DB,
    AGENT_OPERATION_STEP_KIND_DB_PATCH,
  ];

  static final $core.List<AgentOperationStepKind?> _byValue =
      $pb.ProtobufEnum.$_initByValueList(values, 7);
  static AgentOperationStepKind? valueOf($core.int value) =>
      value < 0 || value >= _byValue.length ? null : _byValue[value];

  const AgentOperationStepKind._(super.value, super.name);
}

/// AgentOperationStepStatus is the outcome of a single Saga step.
class AgentOperationStepStatus extends $pb.ProtobufEnum {
  static const AgentOperationStepStatus
      AGENT_OPERATION_STEP_STATUS_UNSPECIFIED = AgentOperationStepStatus._(
          0, _omitEnumNames ? '' : 'AGENT_OPERATION_STEP_STATUS_UNSPECIFIED');
  static const AgentOperationStepStatus AGENT_OPERATION_STEP_STATUS_PENDING =
      AgentOperationStepStatus._(
          1, _omitEnumNames ? '' : 'AGENT_OPERATION_STEP_STATUS_PENDING');
  static const AgentOperationStepStatus AGENT_OPERATION_STEP_STATUS_SUCCEEDED =
      AgentOperationStepStatus._(
          2, _omitEnumNames ? '' : 'AGENT_OPERATION_STEP_STATUS_SUCCEEDED');
  static const AgentOperationStepStatus AGENT_OPERATION_STEP_STATUS_FAILED =
      AgentOperationStepStatus._(
          3, _omitEnumNames ? '' : 'AGENT_OPERATION_STEP_STATUS_FAILED');
  static const AgentOperationStepStatus AGENT_OPERATION_STEP_STATUS_SKIPPED =
      AgentOperationStepStatus._(
          4, _omitEnumNames ? '' : 'AGENT_OPERATION_STEP_STATUS_SKIPPED');

  static const $core.List<AgentOperationStepStatus> values =
      <AgentOperationStepStatus>[
    AGENT_OPERATION_STEP_STATUS_UNSPECIFIED,
    AGENT_OPERATION_STEP_STATUS_PENDING,
    AGENT_OPERATION_STEP_STATUS_SUCCEEDED,
    AGENT_OPERATION_STEP_STATUS_FAILED,
    AGENT_OPERATION_STEP_STATUS_SKIPPED,
  ];

  static final $core.List<AgentOperationStepStatus?> _byValue =
      $pb.ProtobufEnum.$_initByValueList(values, 4);
  static AgentOperationStepStatus? valueOf($core.int value) =>
      value < 0 || value >= _byValue.length ? null : _byValue[value];

  const AgentOperationStepStatus._(super.value, super.name);
}

const $core.bool _omitEnumNames =
    $core.bool.fromEnvironment('protobuf.omit_enum_names');
