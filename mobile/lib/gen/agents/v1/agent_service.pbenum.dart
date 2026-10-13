// This is a generated file - do not edit.
//
// Generated from agents/v1/agent_service.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names, prefer_relative_imports

import 'dart:core' as $core;

import 'package:protobuf/protobuf.dart' as $pb;

/// AgentRuntimeState classifies the agent's most recent runtime state, derived
/// from the latest invocation.
class AgentRuntimeState extends $pb.ProtobufEnum {
  static const AgentRuntimeState AGENT_RUNTIME_STATE_UNSPECIFIED =
      AgentRuntimeState._(
          0, _omitEnumNames ? '' : 'AGENT_RUNTIME_STATE_UNSPECIFIED');
  static const AgentRuntimeState AGENT_RUNTIME_STATE_IDLE =
      AgentRuntimeState._(1, _omitEnumNames ? '' : 'AGENT_RUNTIME_STATE_IDLE');
  static const AgentRuntimeState AGENT_RUNTIME_STATE_RUNNING =
      AgentRuntimeState._(
          2, _omitEnumNames ? '' : 'AGENT_RUNTIME_STATE_RUNNING');
  static const AgentRuntimeState AGENT_RUNTIME_STATE_FAILED =
      AgentRuntimeState._(
          3, _omitEnumNames ? '' : 'AGENT_RUNTIME_STATE_FAILED');

  static const $core.List<AgentRuntimeState> values = <AgentRuntimeState>[
    AGENT_RUNTIME_STATE_UNSPECIFIED,
    AGENT_RUNTIME_STATE_IDLE,
    AGENT_RUNTIME_STATE_RUNNING,
    AGENT_RUNTIME_STATE_FAILED,
  ];

  static final $core.List<AgentRuntimeState?> _byValue =
      $pb.ProtobufEnum.$_initByValueList(values, 3);
  static AgentRuntimeState? valueOf($core.int value) =>
      value < 0 || value >= _byValue.length ? null : _byValue[value];

  const AgentRuntimeState._(super.value, super.name);
}

/// MigrationReadiness classifies an agent's readiness for the identity
/// migration.
class MigrationReadiness extends $pb.ProtobufEnum {
  static const MigrationReadiness MIGRATION_READINESS_UNSPECIFIED =
      MigrationReadiness._(
          0, _omitEnumNames ? '' : 'MIGRATION_READINESS_UNSPECIFIED');

  /// The agent has a valid Agent ID and all dependencies are resolved.
  static const MigrationReadiness MIGRATION_READINESS_READY =
      MigrationReadiness._(
          1, _omitEnumNames ? '' : 'MIGRATION_READINESS_READY');

  /// The agent has not been assigned an Agent ID.
  static const MigrationReadiness MIGRATION_READINESS_MISSING_ID =
      MigrationReadiness._(
          2, _omitEnumNames ? '' : 'MIGRATION_READINESS_MISSING_ID');

  /// The agent's Agent ID conflicts with another agent in the workspace.
  static const MigrationReadiness MIGRATION_READINESS_CONFLICT =
      MigrationReadiness._(
          3, _omitEnumNames ? '' : 'MIGRATION_READINESS_CONFLICT');

  /// The agent references sub-agents or workflow nodes whose targets
  /// have not been assigned an Agent ID.
  static const MigrationReadiness MIGRATION_READINESS_INCOMPLETE_DEPS =
      MigrationReadiness._(
          4, _omitEnumNames ? '' : 'MIGRATION_READINESS_INCOMPLETE_DEPS');

  static const $core.List<MigrationReadiness> values = <MigrationReadiness>[
    MIGRATION_READINESS_UNSPECIFIED,
    MIGRATION_READINESS_READY,
    MIGRATION_READINESS_MISSING_ID,
    MIGRATION_READINESS_CONFLICT,
    MIGRATION_READINESS_INCOMPLETE_DEPS,
  ];

  static final $core.List<MigrationReadiness?> _byValue =
      $pb.ProtobufEnum.$_initByValueList(values, 4);
  static MigrationReadiness? valueOf($core.int value) =>
      value < 0 || value >= _byValue.length ? null : _byValue[value];

  const MigrationReadiness._(super.value, super.name);
}

class MigrateMode extends $pb.ProtobufEnum {
  static const MigrateMode MIGRATE_MODE_UNSPECIFIED =
      MigrateMode._(0, _omitEnumNames ? '' : 'MIGRATE_MODE_UNSPECIFIED');

  /// Report what would be migrated without making changes.
  static const MigrateMode MIGRATE_MODE_DRY_RUN =
      MigrateMode._(1, _omitEnumNames ? '' : 'MIGRATE_MODE_DRY_RUN');

  /// Expand eligible legacy trees into independent Agent records.
  static const MigrateMode MIGRATE_MODE_APPLY =
      MigrateMode._(2, _omitEnumNames ? '' : 'MIGRATE_MODE_APPLY');

  /// Verify that all agents are fully migrated.
  static const MigrateMode MIGRATE_MODE_VERIFY =
      MigrateMode._(3, _omitEnumNames ? '' : 'MIGRATE_MODE_VERIFY');

  static const $core.List<MigrateMode> values = <MigrateMode>[
    MIGRATE_MODE_UNSPECIFIED,
    MIGRATE_MODE_DRY_RUN,
    MIGRATE_MODE_APPLY,
    MIGRATE_MODE_VERIFY,
  ];

  static final $core.List<MigrateMode?> _byValue =
      $pb.ProtobufEnum.$_initByValueList(values, 3);
  static MigrateMode? valueOf($core.int value) =>
      value < 0 || value >= _byValue.length ? null : _byValue[value];

  const MigrateMode._(super.value, super.name);
}

/// InvocationStatus tracks the lifecycle of an agent invocation.
class InvocationStatus extends $pb.ProtobufEnum {
  static const InvocationStatus INVOCATION_STATUS_UNSPECIFIED =
      InvocationStatus._(
          0, _omitEnumNames ? '' : 'INVOCATION_STATUS_UNSPECIFIED');
  static const InvocationStatus INVOCATION_STATUS_RUNNING =
      InvocationStatus._(1, _omitEnumNames ? '' : 'INVOCATION_STATUS_RUNNING');
  static const InvocationStatus INVOCATION_STATUS_SUCCEEDED =
      InvocationStatus._(
          2, _omitEnumNames ? '' : 'INVOCATION_STATUS_SUCCEEDED');
  static const InvocationStatus INVOCATION_STATUS_FAILED =
      InvocationStatus._(3, _omitEnumNames ? '' : 'INVOCATION_STATUS_FAILED');
  static const InvocationStatus INVOCATION_STATUS_QUEUED =
      InvocationStatus._(4, _omitEnumNames ? '' : 'INVOCATION_STATUS_QUEUED');
  static const InvocationStatus INVOCATION_STATUS_CANCELLED =
      InvocationStatus._(
          5, _omitEnumNames ? '' : 'INVOCATION_STATUS_CANCELLED');

  static const $core.List<InvocationStatus> values = <InvocationStatus>[
    INVOCATION_STATUS_UNSPECIFIED,
    INVOCATION_STATUS_RUNNING,
    INVOCATION_STATUS_SUCCEEDED,
    INVOCATION_STATUS_FAILED,
    INVOCATION_STATUS_QUEUED,
    INVOCATION_STATUS_CANCELLED,
  ];

  static final $core.List<InvocationStatus?> _byValue =
      $pb.ProtobufEnum.$_initByValueList(values, 5);
  static InvocationStatus? valueOf($core.int value) =>
      value < 0 || value >= _byValue.length ? null : _byValue[value];

  const InvocationStatus._(super.value, super.name);
}

class MCPOAuthConnectionState extends $pb.ProtobufEnum {
  static const MCPOAuthConnectionState MCPO_AUTH_CONNECTION_STATE_UNSPECIFIED =
      MCPOAuthConnectionState._(
          0, _omitEnumNames ? '' : 'MCPO_AUTH_CONNECTION_STATE_UNSPECIFIED');
  static const MCPOAuthConnectionState MCPO_AUTH_CONNECTION_STATE_DISCONNECTED =
      MCPOAuthConnectionState._(
          1, _omitEnumNames ? '' : 'MCPO_AUTH_CONNECTION_STATE_DISCONNECTED');
  static const MCPOAuthConnectionState MCPO_AUTH_CONNECTION_STATE_CONNECTED =
      MCPOAuthConnectionState._(
          2, _omitEnumNames ? '' : 'MCPO_AUTH_CONNECTION_STATE_CONNECTED');
  static const MCPOAuthConnectionState
      MCPO_AUTH_CONNECTION_STATE_REAUTHORIZATION_REQUIRED =
      MCPOAuthConnectionState._(
          3,
          _omitEnumNames
              ? ''
              : 'MCPO_AUTH_CONNECTION_STATE_REAUTHORIZATION_REQUIRED');
  static const MCPOAuthConnectionState MCPO_AUTH_CONNECTION_STATE_ERROR =
      MCPOAuthConnectionState._(
          4, _omitEnumNames ? '' : 'MCPO_AUTH_CONNECTION_STATE_ERROR');

  static const $core.List<MCPOAuthConnectionState> values =
      <MCPOAuthConnectionState>[
    MCPO_AUTH_CONNECTION_STATE_UNSPECIFIED,
    MCPO_AUTH_CONNECTION_STATE_DISCONNECTED,
    MCPO_AUTH_CONNECTION_STATE_CONNECTED,
    MCPO_AUTH_CONNECTION_STATE_REAUTHORIZATION_REQUIRED,
    MCPO_AUTH_CONNECTION_STATE_ERROR,
  ];

  static final $core.List<MCPOAuthConnectionState?> _byValue =
      $pb.ProtobufEnum.$_initByValueList(values, 4);
  static MCPOAuthConnectionState? valueOf($core.int value) =>
      value < 0 || value >= _byValue.length ? null : _byValue[value];

  const MCPOAuthConnectionState._(super.value, super.name);
}

const $core.bool _omitEnumNames =
    $core.bool.fromEnvironment('protobuf.omit_enum_names');
