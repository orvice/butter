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

import 'package:fixnum/fixnum.dart' as $fixnum;
import 'package:protobuf/protobuf.dart' as $pb;
import 'package:protobuf/well_known_types/google/protobuf/timestamp.pb.dart'
    as $0;

import 'agent_operation.pbenum.dart';

export 'package:protobuf/protobuf.dart' show GeneratedMessageGenericExtensions;

export 'agent_operation.pbenum.dart';

/// AgentOperation is the durable record of a cross-system Agent lifecycle Saga
/// (create, composite save, tombstone, restore, or content purge). It exposes a
/// stable operation ID, an explicit status, and per-step progress so partial
/// failures are visible and operations can be retried idempotently. It does not
/// claim a distributed transaction across MongoDB and Git: each step records its
/// own outcome and the compensation policy is best-effort.
class AgentOperation extends $pb.GeneratedMessage {
  factory AgentOperation({
    $core.String? id,
    $core.String? agentId,
    $core.String? agentName,
    AgentOperationType? type,
    AgentOperationStatus? status,
    $core.int? attemptCount,
    $core.Iterable<AgentOperationStep>? steps,
    $core.String? error,
    $core.String? requestJson,
    $core.String? baseCommitSha,
    $fixnum.Int64? expectedAgentVersion,
    $core.String? committedSha,
    $0.Timestamp? createdAt,
    $0.Timestamp? updatedAt,
    $core.String? workspaceId,
  }) {
    final result = AgentOperation._();
    if (id != null) result.id = id;
    if (agentId != null) result.agentId = agentId;
    if (agentName != null) result.agentName = agentName;
    if (type != null) result.type = type;
    if (status != null) result.status = status;
    if (attemptCount != null) result.attemptCount = attemptCount;
    if (steps != null) result.steps.addAll(steps);
    if (error != null) result.error = error;
    if (requestJson != null) result.requestJson = requestJson;
    if (baseCommitSha != null) result.baseCommitSha = baseCommitSha;
    if (expectedAgentVersion != null)
      result.expectedAgentVersion = expectedAgentVersion;
    if (committedSha != null) result.committedSha = committedSha;
    if (createdAt != null) result.createdAt = createdAt;
    if (updatedAt != null) result.updatedAt = updatedAt;
    if (workspaceId != null) result.workspaceId = workspaceId;
    return result;
  }

  AgentOperation._();

  factory AgentOperation.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      AgentOperation()..mergeFromBuffer(data, registry);
  factory AgentOperation.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      AgentOperation()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'AgentOperation',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: AgentOperation.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..aOS(2, _omitFieldNames ? '' : 'agentId')
    ..aOS(3, _omitFieldNames ? '' : 'agentName')
    ..aE<AgentOperationType>(4, _omitFieldNames ? '' : 'type',
        enumValues: AgentOperationType.values)
    ..aE<AgentOperationStatus>(5, _omitFieldNames ? '' : 'status',
        enumValues: AgentOperationStatus.values)
    ..aI(6, _omitFieldNames ? '' : 'attemptCount')
    ..pPM<AgentOperationStep>(7, _omitFieldNames ? '' : 'steps',
        subBuilder: AgentOperationStep.$_createMessage)
    ..aOS(8, _omitFieldNames ? '' : 'error')
    ..aOS(9, _omitFieldNames ? '' : 'requestJson')
    ..aOS(10, _omitFieldNames ? '' : 'baseCommitSha')
    ..aInt64(11, _omitFieldNames ? '' : 'expectedAgentVersion')
    ..aOS(12, _omitFieldNames ? '' : 'committedSha')
    ..aOM<$0.Timestamp>(30, _omitFieldNames ? '' : 'createdAt',
        subBuilder: $0.Timestamp.$_createMessage)
    ..aOM<$0.Timestamp>(31, _omitFieldNames ? '' : 'updatedAt',
        subBuilder: $0.Timestamp.$_createMessage)
    ..aOS(100, _omitFieldNames ? '' : 'workspaceId')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  AgentOperation clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  AgentOperation copyWith(void Function(AgentOperation) updates) =>
      super.copyWith((message) => updates(message as AgentOperation))
          as AgentOperation;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use AgentOperation() / AgentOperation.new instead')
  static AgentOperation create() => AgentOperation._();
  static $pb.GeneratedMessage $_createMessage() => AgentOperation._();
  @$core.override
  AgentOperation createEmptyInstance() => AgentOperation._();
  @$core.pragma('dart2js:noInline')
  static AgentOperation getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<AgentOperation>(
          AgentOperation.$_createMessage);
  static AgentOperation? _defaultInstance;

  /// Stable operation ID (UUID). Doubles as the idempotency key: re-invoking a
  /// lifecycle RPC with the same operation resumes from the first unfinished
  /// step rather than re-running completed side effects.
  /// @gotags: json:"id,omitempty" yaml:"id,omitempty" bson:"_id,omitempty"
  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);

  /// The immutable Agent ID this operation acts on.
  /// @gotags: json:"agent_id,omitempty" yaml:"agent_id,omitempty" bson:"agent_id,omitempty"
  @$pb.TagNumber(2)
  $core.String get agentId => $_getSZ(1);
  @$pb.TagNumber(2)
  set agentId($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasAgentId() => $_has(1);
  @$pb.TagNumber(2)
  void clearAgentId() => $_clearField(2);

  /// The agent's storage name (fallback identity for agents without an ID).
  /// @gotags: json:"agent_name,omitempty" yaml:"agent_name,omitempty" bson:"agent_name,omitempty"
  @$pb.TagNumber(3)
  $core.String get agentName => $_getSZ(2);
  @$pb.TagNumber(3)
  set agentName($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasAgentName() => $_has(2);
  @$pb.TagNumber(3)
  void clearAgentName() => $_clearField(3);

  /// Which lifecycle Saga this record drives.
  /// @gotags: json:"type,omitempty" yaml:"type,omitempty" bson:"type,omitempty"
  @$pb.TagNumber(4)
  AgentOperationType get type => $_getN(3);
  @$pb.TagNumber(4)
  set type(AgentOperationType value) => $_setField(4, value);
  @$pb.TagNumber(4)
  $core.bool hasType() => $_has(3);
  @$pb.TagNumber(4)
  void clearType() => $_clearField(4);

  /// Overall operation status.
  /// @gotags: json:"status,omitempty" yaml:"status,omitempty" bson:"status,omitempty"
  @$pb.TagNumber(5)
  AgentOperationStatus get status => $_getN(4);
  @$pb.TagNumber(5)
  set status(AgentOperationStatus value) => $_setField(5, value);
  @$pb.TagNumber(5)
  $core.bool hasStatus() => $_has(4);
  @$pb.TagNumber(5)
  void clearStatus() => $_clearField(5);

  /// Number of times the operation has been (re)attempted.
  /// @gotags: json:"attempt_count,omitempty" yaml:"attempt_count,omitempty" bson:"attempt_count,omitempty"
  @$pb.TagNumber(6)
  $core.int get attemptCount => $_getIZ(5);
  @$pb.TagNumber(6)
  set attemptCount($core.int value) => $_setSignedInt32(5, value);
  @$pb.TagNumber(6)
  $core.bool hasAttemptCount() => $_has(5);
  @$pb.TagNumber(6)
  void clearAttemptCount() => $_clearField(6);

  /// Ordered per-step progress records.
  /// @gotags: json:"steps,omitempty" yaml:"steps,omitempty" bson:"steps,omitempty"
  @$pb.TagNumber(7)
  $pb.PbList<AgentOperationStep> get steps => $_getList(6);

  /// Human-readable error from the most recent failure, empty when not failed.
  /// @gotags: json:"error,omitempty" yaml:"error,omitempty" bson:"error,omitempty"
  @$pb.TagNumber(8)
  $core.String get error => $_getSZ(7);
  @$pb.TagNumber(8)
  set error($core.String value) => $_setString(7, value);
  @$pb.TagNumber(8)
  $core.bool hasError() => $_has(7);
  @$pb.TagNumber(8)
  void clearError() => $_clearField(8);

  /// protojson of the originating request, so a retry can resume deterministically.
  /// @gotags: json:"request_json,omitempty" yaml:"request_json,omitempty" bson:"request_json,omitempty"
  @$pb.TagNumber(9)
  $core.String get requestJson => $_getSZ(8);
  @$pb.TagNumber(9)
  set requestJson($core.String value) => $_setString(8, value);
  @$pb.TagNumber(9)
  $core.bool hasRequestJson() => $_has(8);
  @$pb.TagNumber(9)
  void clearRequestJson() => $_clearField(9);

  /// The base commit SHA carried into content steps (informational; the content
  /// layer applies last-write-wins against the live branch HEAD).
  /// @gotags: json:"base_commit_sha,omitempty" yaml:"base_commit_sha,omitempty" bson:"base_commit_sha,omitempty"
  @$pb.TagNumber(10)
  $core.String get baseCommitSha => $_getSZ(9);
  @$pb.TagNumber(10)
  set baseCommitSha($core.String value) => $_setString(9, value);
  @$pb.TagNumber(10)
  $core.bool hasBaseCommitSha() => $_has(9);
  @$pb.TagNumber(10)
  void clearBaseCommitSha() => $_clearField(10);

  /// The agent version the caller expected (optimistic concurrency for saves).
  /// @gotags: json:"expected_agent_version,omitempty" yaml:"expected_agent_version,omitempty" bson:"expected_agent_version,omitempty"
  @$pb.TagNumber(11)
  $fixnum.Int64 get expectedAgentVersion => $_getI64(10);
  @$pb.TagNumber(11)
  set expectedAgentVersion($fixnum.Int64 value) => $_setInt64(10, value);
  @$pb.TagNumber(11)
  $core.bool hasExpectedAgentVersion() => $_has(10);
  @$pb.TagNumber(11)
  void clearExpectedAgentVersion() => $_clearField(11);

  /// The SHA produced by the content-commit step once it succeeds. Guards a
  /// retry from re-committing identical content.
  /// @gotags: json:"committed_sha,omitempty" yaml:"committed_sha,omitempty" bson:"committed_sha,omitempty"
  @$pb.TagNumber(12)
  $core.String get committedSha => $_getSZ(11);
  @$pb.TagNumber(12)
  set committedSha($core.String value) => $_setString(11, value);
  @$pb.TagNumber(12)
  $core.bool hasCommittedSha() => $_has(11);
  @$pb.TagNumber(12)
  void clearCommittedSha() => $_clearField(12);

  /// @gotags: json:"created_at,omitempty" yaml:"created_at,omitempty" bson:"created_at,omitempty"
  @$pb.TagNumber(30)
  $0.Timestamp get createdAt => $_getN(12);
  @$pb.TagNumber(30)
  set createdAt($0.Timestamp value) => $_setField(30, value);
  @$pb.TagNumber(30)
  $core.bool hasCreatedAt() => $_has(12);
  @$pb.TagNumber(30)
  void clearCreatedAt() => $_clearField(30);
  @$pb.TagNumber(30)
  $0.Timestamp ensureCreatedAt() => $_ensure(12);

  /// @gotags: json:"updated_at,omitempty" yaml:"updated_at,omitempty" bson:"updated_at,omitempty"
  @$pb.TagNumber(31)
  $0.Timestamp get updatedAt => $_getN(13);
  @$pb.TagNumber(31)
  set updatedAt($0.Timestamp value) => $_setField(31, value);
  @$pb.TagNumber(31)
  $core.bool hasUpdatedAt() => $_has(13);
  @$pb.TagNumber(31)
  void clearUpdatedAt() => $_clearField(31);
  @$pb.TagNumber(31)
  $0.Timestamp ensureUpdatedAt() => $_ensure(13);

  /// ID of the workspace this operation belongs to.
  /// @gotags: json:"workspace_id,omitempty" yaml:"workspace_id,omitempty" bson:"workspace_id,omitempty"
  @$pb.TagNumber(100)
  $core.String get workspaceId => $_getSZ(14);
  @$pb.TagNumber(100)
  set workspaceId($core.String value) => $_setString(14, value);
  @$pb.TagNumber(100)
  $core.bool hasWorkspaceId() => $_has(14);
  @$pb.TagNumber(100)
  void clearWorkspaceId() => $_clearField(100);
}

/// AgentOperationStep records the outcome of one step within a lifecycle Saga.
class AgentOperationStep extends $pb.GeneratedMessage {
  factory AgentOperationStep({
    AgentOperationStepKind? kind,
    AgentOperationStepStatus? status,
    $core.int? attemptCount,
    $core.String? error,
    $0.Timestamp? startedAt,
    $0.Timestamp? finishedAt,
  }) {
    final result = AgentOperationStep._();
    if (kind != null) result.kind = kind;
    if (status != null) result.status = status;
    if (attemptCount != null) result.attemptCount = attemptCount;
    if (error != null) result.error = error;
    if (startedAt != null) result.startedAt = startedAt;
    if (finishedAt != null) result.finishedAt = finishedAt;
    return result;
  }

  AgentOperationStep._();

  factory AgentOperationStep.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      AgentOperationStep()..mergeFromBuffer(data, registry);
  factory AgentOperationStep.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      AgentOperationStep()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'AgentOperationStep',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: AgentOperationStep.$_createMessage)
    ..aE<AgentOperationStepKind>(1, _omitFieldNames ? '' : 'kind',
        enumValues: AgentOperationStepKind.values)
    ..aE<AgentOperationStepStatus>(2, _omitFieldNames ? '' : 'status',
        enumValues: AgentOperationStepStatus.values)
    ..aI(3, _omitFieldNames ? '' : 'attemptCount')
    ..aOS(4, _omitFieldNames ? '' : 'error')
    ..aOM<$0.Timestamp>(5, _omitFieldNames ? '' : 'startedAt',
        subBuilder: $0.Timestamp.$_createMessage)
    ..aOM<$0.Timestamp>(6, _omitFieldNames ? '' : 'finishedAt',
        subBuilder: $0.Timestamp.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  AgentOperationStep clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  AgentOperationStep copyWith(void Function(AgentOperationStep) updates) =>
      super.copyWith((message) => updates(message as AgentOperationStep))
          as AgentOperationStep;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use AgentOperationStep() / AgentOperationStep.new instead')
  static AgentOperationStep create() => AgentOperationStep._();
  static $pb.GeneratedMessage $_createMessage() => AgentOperationStep._();
  @$core.override
  AgentOperationStep createEmptyInstance() => AgentOperationStep._();
  @$core.pragma('dart2js:noInline')
  static AgentOperationStep getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<AgentOperationStep>(
          AgentOperationStep.$_createMessage);
  static AgentOperationStep? _defaultInstance;

  /// @gotags: json:"kind,omitempty" yaml:"kind,omitempty" bson:"kind,omitempty"
  @$pb.TagNumber(1)
  AgentOperationStepKind get kind => $_getN(0);
  @$pb.TagNumber(1)
  set kind(AgentOperationStepKind value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasKind() => $_has(0);
  @$pb.TagNumber(1)
  void clearKind() => $_clearField(1);

  /// @gotags: json:"status,omitempty" yaml:"status,omitempty" bson:"status,omitempty"
  @$pb.TagNumber(2)
  AgentOperationStepStatus get status => $_getN(1);
  @$pb.TagNumber(2)
  set status(AgentOperationStepStatus value) => $_setField(2, value);
  @$pb.TagNumber(2)
  $core.bool hasStatus() => $_has(1);
  @$pb.TagNumber(2)
  void clearStatus() => $_clearField(2);

  /// @gotags: json:"attempt_count,omitempty" yaml:"attempt_count,omitempty" bson:"attempt_count,omitempty"
  @$pb.TagNumber(3)
  $core.int get attemptCount => $_getIZ(2);
  @$pb.TagNumber(3)
  set attemptCount($core.int value) => $_setSignedInt32(2, value);
  @$pb.TagNumber(3)
  $core.bool hasAttemptCount() => $_has(2);
  @$pb.TagNumber(3)
  void clearAttemptCount() => $_clearField(3);

  /// @gotags: json:"error,omitempty" yaml:"error,omitempty" bson:"error,omitempty"
  @$pb.TagNumber(4)
  $core.String get error => $_getSZ(3);
  @$pb.TagNumber(4)
  set error($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasError() => $_has(3);
  @$pb.TagNumber(4)
  void clearError() => $_clearField(4);

  /// @gotags: json:"started_at,omitempty" yaml:"started_at,omitempty" bson:"started_at,omitempty"
  @$pb.TagNumber(5)
  $0.Timestamp get startedAt => $_getN(4);
  @$pb.TagNumber(5)
  set startedAt($0.Timestamp value) => $_setField(5, value);
  @$pb.TagNumber(5)
  $core.bool hasStartedAt() => $_has(4);
  @$pb.TagNumber(5)
  void clearStartedAt() => $_clearField(5);
  @$pb.TagNumber(5)
  $0.Timestamp ensureStartedAt() => $_ensure(4);

  /// @gotags: json:"finished_at,omitempty" yaml:"finished_at,omitempty" bson:"finished_at,omitempty"
  @$pb.TagNumber(6)
  $0.Timestamp get finishedAt => $_getN(5);
  @$pb.TagNumber(6)
  set finishedAt($0.Timestamp value) => $_setField(6, value);
  @$pb.TagNumber(6)
  $core.bool hasFinishedAt() => $_has(5);
  @$pb.TagNumber(6)
  void clearFinishedAt() => $_clearField(6);
  @$pb.TagNumber(6)
  $0.Timestamp ensureFinishedAt() => $_ensure(5);
}

const $core.bool _omitFieldNames =
    $core.bool.fromEnvironment('protobuf.omit_field_names');
const $core.bool _omitMessageNames =
    $core.bool.fromEnvironment('protobuf.omit_message_names');
