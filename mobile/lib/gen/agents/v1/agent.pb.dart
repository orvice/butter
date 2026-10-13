// This is a generated file - do not edit.
//
// Generated from agents/v1/agent.proto.

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

import 'agent.pbenum.dart';
import 'agent_file.pb.dart' as $1;

export 'package:protobuf/protobuf.dart' show GeneratedMessageGenericExtensions;

export 'agent.pbenum.dart';

/// Agent describes a declarative ADK agent configuration.
class Agent extends $pb.GeneratedMessage {
  factory Agent({
    $core.String? name,
    $core.String? description,
    @$core.Deprecated('This field is deprecated.')
    $core.Iterable<Agent>? subAgents,
    $core.Iterable<$core.MapEntry<$core.String, $core.String>>? labels,
    $core.Iterable<$core.MapEntry<$core.String, $core.String>>? metadata,
    AgentConfig? config,
    AgentType? type,
    $core.bool? enableA2a,
    $core.bool? enableOpenaiApi,
    $core.String? agentId,
    $core.String? displayName,
    $core.Iterable<$core.String>? childAgentIds,
    AgentLifecycleStatus? lifecycleStatus,
    $core.String? legacyName,
    $fixnum.Int64? version,
    $0.Timestamp? createdAt,
    $0.Timestamp? updatedAt,
    $0.Timestamp? deletedAt,
    $core.bool? enableAgui,
    $core.String? workspaceId,
  }) {
    final result = Agent._();
    if (name != null) result.name = name;
    if (description != null) result.description = description;
    if (subAgents != null) result.subAgents.addAll(subAgents);
    if (labels != null) result.labels.addEntries(labels);
    if (metadata != null) result.metadata.addEntries(metadata);
    if (config != null) result.config = config;
    if (type != null) result.type = type;
    if (enableA2a != null) result.enableA2a = enableA2a;
    if (enableOpenaiApi != null) result.enableOpenaiApi = enableOpenaiApi;
    if (agentId != null) result.agentId = agentId;
    if (displayName != null) result.displayName = displayName;
    if (childAgentIds != null) result.childAgentIds.addAll(childAgentIds);
    if (lifecycleStatus != null) result.lifecycleStatus = lifecycleStatus;
    if (legacyName != null) result.legacyName = legacyName;
    if (version != null) result.version = version;
    if (createdAt != null) result.createdAt = createdAt;
    if (updatedAt != null) result.updatedAt = updatedAt;
    if (deletedAt != null) result.deletedAt = deletedAt;
    if (enableAgui != null) result.enableAgui = enableAgui;
    if (workspaceId != null) result.workspaceId = workspaceId;
    return result;
  }

  Agent._();

  factory Agent.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      Agent()..mergeFromBuffer(data, registry);
  factory Agent.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      Agent()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'Agent',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: Agent.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'name')
    ..aOS(2, _omitFieldNames ? '' : 'description')
    ..pPM<Agent>(3, _omitFieldNames ? '' : 'subAgents',
        subBuilder: Agent.$_createMessage)
    ..m<$core.String, $core.String>(4, _omitFieldNames ? '' : 'labels',
        entryClassName: 'Agent.LabelsEntry',
        keyFieldType: $pb.PbFieldType.OS,
        valueFieldType: $pb.PbFieldType.OS,
        packageName: const $pb.PackageName('agents.v1'))
    ..m<$core.String, $core.String>(5, _omitFieldNames ? '' : 'metadata',
        entryClassName: 'Agent.MetadataEntry',
        keyFieldType: $pb.PbFieldType.OS,
        valueFieldType: $pb.PbFieldType.OS,
        packageName: const $pb.PackageName('agents.v1'))
    ..aOM<AgentConfig>(6, _omitFieldNames ? '' : 'config',
        subBuilder: AgentConfig.$_createMessage)
    ..aE<AgentType>(20, _omitFieldNames ? '' : 'type',
        enumValues: AgentType.values)
    ..aOB(21, _omitFieldNames ? '' : 'enableA2a')
    ..aOB(22, _omitFieldNames ? '' : 'enableOpenaiApi')
    ..aOS(23, _omitFieldNames ? '' : 'agentId')
    ..aOS(24, _omitFieldNames ? '' : 'displayName')
    ..pPS(25, _omitFieldNames ? '' : 'childAgentIds')
    ..aE<AgentLifecycleStatus>(26, _omitFieldNames ? '' : 'lifecycleStatus',
        enumValues: AgentLifecycleStatus.values)
    ..aOS(27, _omitFieldNames ? '' : 'legacyName')
    ..aInt64(28, _omitFieldNames ? '' : 'version')
    ..aOM<$0.Timestamp>(29, _omitFieldNames ? '' : 'createdAt',
        subBuilder: $0.Timestamp.$_createMessage)
    ..aOM<$0.Timestamp>(30, _omitFieldNames ? '' : 'updatedAt',
        subBuilder: $0.Timestamp.$_createMessage)
    ..aOM<$0.Timestamp>(31, _omitFieldNames ? '' : 'deletedAt',
        subBuilder: $0.Timestamp.$_createMessage)
    ..aOB(32, _omitFieldNames ? '' : 'enableAgui')
    ..aOS(100, _omitFieldNames ? '' : 'workspaceId')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Agent clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Agent copyWith(void Function(Agent) updates) =>
      super.copyWith((message) => updates(message as Agent)) as Agent;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use Agent() / Agent.new instead')
  static Agent create() => Agent._();
  static $pb.GeneratedMessage $_createMessage() => Agent._();
  @$core.override
  Agent createEmptyInstance() => Agent._();
  @$core.pragma('dart2js:noInline')
  static Agent getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<Agent>(Agent.$_createMessage);
  static Agent? _defaultInstance;

  /// Name must be unique within an agent tree.
  /// @gotags: json:"name,omitempty" yaml:"name,omitempty"
  @$pb.TagNumber(1)
  $core.String get name => $_getSZ(0);
  @$pb.TagNumber(1)
  set name($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasName() => $_has(0);
  @$pb.TagNumber(1)
  void clearName() => $_clearField(1);

  /// Description helps the LLM decide when to delegate to this agent.
  /// @gotags: json:"description,omitempty" yaml:"description,omitempty"
  @$pb.TagNumber(2)
  $core.String get description => $_getSZ(1);
  @$pb.TagNumber(2)
  set description($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasDescription() => $_has(1);
  @$pb.TagNumber(2)
  void clearDescription() => $_clearField(2);

  /// Legacy embedded sub-agent tree. Deprecated (issue #241): not writable,
  /// never used for construction — children are referenced via
  /// child_agent_ids. Retained only so historical records decode.
  /// @gotags: json:"sub_agents,omitempty" yaml:"sub_agents,omitempty"
  @$core.Deprecated('This field is deprecated.')
  @$pb.TagNumber(3)
  $pb.PbList<Agent> get subAgents => $_getList(2);

  /// Labels are intended for indexing, routing, or coarse-grained filtering.
  /// @gotags: json:"labels,omitempty" yaml:"labels,omitempty"
  @$pb.TagNumber(4)
  $pb.PbMap<$core.String, $core.String> get labels => $_getMap(3);

  /// Metadata stores non-routing annotations that callers may want to keep.
  /// @gotags: json:"metadata,omitempty" yaml:"metadata,omitempty"
  @$pb.TagNumber(5)
  $pb.PbMap<$core.String, $core.String> get metadata => $_getMap(4);

  /// Config contains execution and integration settings for this agent.
  /// @gotags: json:"config,omitempty" yaml:"config,omitempty"
  @$pb.TagNumber(6)
  AgentConfig get config => $_getN(5);
  @$pb.TagNumber(6)
  set config(AgentConfig value) => $_setField(6, value);
  @$pb.TagNumber(6)
  $core.bool hasConfig() => $_has(5);
  @$pb.TagNumber(6)
  void clearConfig() => $_clearField(6);
  @$pb.TagNumber(6)
  AgentConfig ensureConfig() => $_ensure(5);

  /// Type indicates which ADK agent implementation this node represents.
  /// @gotags: json:"type,omitempty" yaml:"type,omitempty"
  @$pb.TagNumber(20)
  AgentType get type => $_getN(6);
  @$pb.TagNumber(20)
  set type(AgentType value) => $_setField(20, value);
  @$pb.TagNumber(20)
  $core.bool hasType() => $_has(6);
  @$pb.TagNumber(20)
  void clearType() => $_clearField(20);

  /// Whether this agent should be exposed via A2A protocol over HTTP.
  /// @gotags: json:"enable_a2a,omitempty" yaml:"enable_a2a,omitempty"
  @$pb.TagNumber(21)
  $core.bool get enableA2a => $_getBF(7);
  @$pb.TagNumber(21)
  set enableA2a($core.bool value) => $_setBool(7, value);
  @$pb.TagNumber(21)
  $core.bool hasEnableA2a() => $_has(7);
  @$pb.TagNumber(21)
  void clearEnableA2a() => $_clearField(21);

  /// Whether this agent should be exposed via the OpenAI-compatible API.
  /// @gotags: json:"enable_openai_api,omitempty" yaml:"enable_openai_api,omitempty"
  @$pb.TagNumber(22)
  $core.bool get enableOpenaiApi => $_getBF(8);
  @$pb.TagNumber(22)
  set enableOpenaiApi($core.bool value) => $_setBool(8, value);
  @$pb.TagNumber(22)
  $core.bool hasEnableOpenaiApi() => $_has(8);
  @$pb.TagNumber(22)
  void clearEnableOpenaiApi() => $_clearField(22);

  /// Immutable, workspace-unique slug identifier assigned at creation.
  /// (workspace_id, agent_id) is the Agent's logical primary key everywhere:
  /// persistence, APIs, and relationships. Once set it cannot be changed or
  /// reused.
  /// @gotags: json:"agent_id,omitempty" yaml:"agent_id,omitempty" bson:"agent_id,omitempty"
  @$pb.TagNumber(23)
  $core.String get agentId => $_getSZ(9);
  @$pb.TagNumber(23)
  set agentId($core.String value) => $_setString(9, value);
  @$pb.TagNumber(23)
  $core.bool hasAgentId() => $_has(9);
  @$pb.TagNumber(23)
  void clearAgentId() => $_clearField(23);

  /// Mutable human-readable label for UI display. Does not participate in
  /// identity resolution or runtime lookup.
  /// @gotags: json:"display_name,omitempty" yaml:"display_name,omitempty" bson:"display_name,omitempty"
  @$pb.TagNumber(24)
  $core.String get displayName => $_getSZ(10);
  @$pb.TagNumber(24)
  set displayName($core.String value) => $_setString(10, value);
  @$pb.TagNumber(24)
  $core.bool hasDisplayName() => $_has(10);
  @$pb.TagNumber(24)
  void clearDisplayName() => $_clearField(24);

  /// Ordered references to child Agent IDs for LLM, Loop, Sequential, and
  /// Parallel agents. Replaces embedded sub_agents for V2 agents.
  /// @gotags: json:"child_agent_ids,omitempty" yaml:"child_agent_ids,omitempty" bson:"child_agent_ids,omitempty"
  @$pb.TagNumber(25)
  $pb.PbList<$core.String> get childAgentIds => $_getList(11);

  /// Lifecycle status tracks whether the agent is runnable or blocked.
  /// @gotags: json:"lifecycle_status,omitempty" yaml:"lifecycle_status,omitempty" bson:"lifecycle_status,omitempty"
  @$pb.TagNumber(26)
  AgentLifecycleStatus get lifecycleStatus => $_getN(12);
  @$pb.TagNumber(26)
  set lifecycleStatus(AgentLifecycleStatus value) => $_setField(26, value);
  @$pb.TagNumber(26)
  $core.bool hasLifecycleStatus() => $_has(12);
  @$pb.TagNumber(26)
  void clearLifecycleStatus() => $_clearField(26);

  /// Original agent name preserved during the V2 migration observation period
  /// so the cutover can be rolled back without data loss.
  /// @gotags: json:"legacy_name,omitempty" yaml:"legacy_name,omitempty" bson:"legacy_name,omitempty"
  @$pb.TagNumber(27)
  $core.String get legacyName => $_getSZ(13);
  @$pb.TagNumber(27)
  set legacyName($core.String value) => $_setString(13, value);
  @$pb.TagNumber(27)
  $core.bool hasLegacyName() => $_has(13);
  @$pb.TagNumber(27)
  void clearLegacyName() => $_clearField(27);

  /// Monotonic optimistic-concurrency counter, bumped on every operational
  /// config write. Composite saves (UpdateAgentConfiguration) pass the version
  /// they observed as expected_agent_version; a mismatch aborts the write.
  /// Legacy rows created before this field carry version 0.
  /// @gotags: json:"version,omitempty" yaml:"version,omitempty" bson:"version,omitempty"
  @$pb.TagNumber(28)
  $fixnum.Int64 get version => $_getI64(14);
  @$pb.TagNumber(28)
  set version($fixnum.Int64 value) => $_setInt64(14, value);
  @$pb.TagNumber(28)
  $core.bool hasVersion() => $_has(14);
  @$pb.TagNumber(28)
  void clearVersion() => $_clearField(28);

  /// When the agent entity was first created.
  /// @gotags: json:"created_at,omitempty" yaml:"created_at,omitempty" bson:"created_at,omitempty"
  @$pb.TagNumber(29)
  $0.Timestamp get createdAt => $_getN(15);
  @$pb.TagNumber(29)
  set createdAt($0.Timestamp value) => $_setField(29, value);
  @$pb.TagNumber(29)
  $core.bool hasCreatedAt() => $_has(15);
  @$pb.TagNumber(29)
  void clearCreatedAt() => $_clearField(29);
  @$pb.TagNumber(29)
  $0.Timestamp ensureCreatedAt() => $_ensure(15);

  /// When the agent entity was last written.
  /// @gotags: json:"updated_at,omitempty" yaml:"updated_at,omitempty" bson:"updated_at,omitempty"
  @$pb.TagNumber(30)
  $0.Timestamp get updatedAt => $_getN(16);
  @$pb.TagNumber(30)
  set updatedAt($0.Timestamp value) => $_setField(30, value);
  @$pb.TagNumber(30)
  $core.bool hasUpdatedAt() => $_has(16);
  @$pb.TagNumber(30)
  void clearUpdatedAt() => $_clearField(30);
  @$pb.TagNumber(30)
  $0.Timestamp ensureUpdatedAt() => $_ensure(16);

  /// When the agent was soft-deleted (tombstoned). Set when lifecycle_status
  /// becomes DELETED; cleared on restore.
  /// @gotags: json:"deleted_at,omitempty" yaml:"deleted_at,omitempty" bson:"deleted_at,omitempty"
  @$pb.TagNumber(31)
  $0.Timestamp get deletedAt => $_getN(17);
  @$pb.TagNumber(31)
  set deletedAt($0.Timestamp value) => $_setField(31, value);
  @$pb.TagNumber(31)
  $core.bool hasDeletedAt() => $_has(17);
  @$pb.TagNumber(31)
  void clearDeletedAt() => $_clearField(31);
  @$pb.TagNumber(31)
  $0.Timestamp ensureDeletedAt() => $_ensure(17);

  /// Whether this agent should be exposed via the AG-UI protocol endpoint
  /// (POST /api/agui/{agent_id}). Independent of enable_openai_api: the two
  /// protocols address different clients and are controlled separately.
  /// @gotags: json:"enable_agui,omitempty" yaml:"enable_agui,omitempty"
  @$pb.TagNumber(32)
  $core.bool get enableAgui => $_getBF(18);
  @$pb.TagNumber(32)
  set enableAgui($core.bool value) => $_setBool(18, value);
  @$pb.TagNumber(32)
  $core.bool hasEnableAgui() => $_has(18);
  @$pb.TagNumber(32)
  void clearEnableAgui() => $_clearField(32);

  /// ID of the workspace this agent belongs to. Server-enforced from request
  /// context on writes; included in responses so clients know the scope.
  /// @gotags: json:"workspace_id,omitempty" yaml:"workspace_id,omitempty" bson:"workspace_id,omitempty"
  @$pb.TagNumber(100)
  $core.String get workspaceId => $_getSZ(19);
  @$pb.TagNumber(100)
  set workspaceId($core.String value) => $_setString(19, value);
  @$pb.TagNumber(100)
  $core.bool hasWorkspaceId() => $_has(19);
  @$pb.TagNumber(100)
  void clearWorkspaceId() => $_clearField(100);
}

/// AgentConfig groups agent-level runtime and integration settings.
class AgentConfig extends $pb.GeneratedMessage {
  factory AgentConfig({
    AgentRuntime? runtime,
    $core.Iterable<MCPServer>? mcpServers,
    ContextGuardConfig? contextGuard,
    $core.Iterable<$core.String>? mcpServerIds,
    $core.Iterable<$core.String>? remoteAgentIds,
    $core.Iterable<$1.AgentFileMount>? fileMounts,
    $core.Iterable<$core.String>? skills,
    MemoryConfig? memory,
    ResultCardConfig? resultCards,
    $core.String? model,
    $core.String? instruction,
    $core.String? globalInstruction,
    $core.bool? disallowTransferToParent,
    $core.bool? disallowTransferToPeers,
    LLMIncludeContents? includeContents,
    $core.String? outputKey,
    $core.String? inputSchemaJson,
    $core.String? outputSchemaJson,
    $core.int? maxIterations,
    WorkflowConfig? workflow,
    PiAgentConfig? pi,
    CursorAgentConfig? cursor,
  }) {
    final result = AgentConfig._();
    if (runtime != null) result.runtime = runtime;
    if (mcpServers != null) result.mcpServers.addAll(mcpServers);
    if (contextGuard != null) result.contextGuard = contextGuard;
    if (mcpServerIds != null) result.mcpServerIds.addAll(mcpServerIds);
    if (remoteAgentIds != null) result.remoteAgentIds.addAll(remoteAgentIds);
    if (fileMounts != null) result.fileMounts.addAll(fileMounts);
    if (skills != null) result.skills.addAll(skills);
    if (memory != null) result.memory = memory;
    if (resultCards != null) result.resultCards = resultCards;
    if (model != null) result.model = model;
    if (instruction != null) result.instruction = instruction;
    if (globalInstruction != null) result.globalInstruction = globalInstruction;
    if (disallowTransferToParent != null)
      result.disallowTransferToParent = disallowTransferToParent;
    if (disallowTransferToPeers != null)
      result.disallowTransferToPeers = disallowTransferToPeers;
    if (includeContents != null) result.includeContents = includeContents;
    if (outputKey != null) result.outputKey = outputKey;
    if (inputSchemaJson != null) result.inputSchemaJson = inputSchemaJson;
    if (outputSchemaJson != null) result.outputSchemaJson = outputSchemaJson;
    if (maxIterations != null) result.maxIterations = maxIterations;
    if (workflow != null) result.workflow = workflow;
    if (pi != null) result.pi = pi;
    if (cursor != null) result.cursor = cursor;
    return result;
  }

  AgentConfig._();

  factory AgentConfig.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      AgentConfig()..mergeFromBuffer(data, registry);
  factory AgentConfig.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      AgentConfig()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'AgentConfig',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: AgentConfig.$_createMessage)
    ..aOM<AgentRuntime>(1, _omitFieldNames ? '' : 'runtime',
        subBuilder: AgentRuntime.$_createMessage)
    ..pPM<MCPServer>(2, _omitFieldNames ? '' : 'mcpServers',
        subBuilder: MCPServer.$_createMessage)
    ..aOM<ContextGuardConfig>(3, _omitFieldNames ? '' : 'contextGuard',
        subBuilder: ContextGuardConfig.$_createMessage)
    ..pPS(4, _omitFieldNames ? '' : 'mcpServerIds')
    ..pPS(5, _omitFieldNames ? '' : 'remoteAgentIds')
    ..pPM<$1.AgentFileMount>(6, _omitFieldNames ? '' : 'fileMounts',
        subBuilder: $1.AgentFileMount.$_createMessage)
    ..pPS(7, _omitFieldNames ? '' : 'skills')
    ..aOM<MemoryConfig>(8, _omitFieldNames ? '' : 'memory',
        subBuilder: MemoryConfig.$_createMessage)
    ..aOM<ResultCardConfig>(9, _omitFieldNames ? '' : 'resultCards',
        subBuilder: ResultCardConfig.$_createMessage)
    ..aOS(20, _omitFieldNames ? '' : 'model')
    ..aOS(21, _omitFieldNames ? '' : 'instruction')
    ..aOS(22, _omitFieldNames ? '' : 'globalInstruction')
    ..aOB(23, _omitFieldNames ? '' : 'disallowTransferToParent')
    ..aOB(24, _omitFieldNames ? '' : 'disallowTransferToPeers')
    ..aE<LLMIncludeContents>(25, _omitFieldNames ? '' : 'includeContents',
        enumValues: LLMIncludeContents.values)
    ..aOS(26, _omitFieldNames ? '' : 'outputKey')
    ..aOS(27, _omitFieldNames ? '' : 'inputSchemaJson')
    ..aOS(28, _omitFieldNames ? '' : 'outputSchemaJson')
    ..aI(30, _omitFieldNames ? '' : 'maxIterations',
        fieldType: $pb.PbFieldType.OU3)
    ..aOM<WorkflowConfig>(40, _omitFieldNames ? '' : 'workflow',
        subBuilder: WorkflowConfig.$_createMessage)
    ..aOM<PiAgentConfig>(50, _omitFieldNames ? '' : 'pi',
        subBuilder: PiAgentConfig.$_createMessage)
    ..aOM<CursorAgentConfig>(51, _omitFieldNames ? '' : 'cursor',
        subBuilder: CursorAgentConfig.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  AgentConfig clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  AgentConfig copyWith(void Function(AgentConfig) updates) =>
      super.copyWith((message) => updates(message as AgentConfig))
          as AgentConfig;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use AgentConfig() / AgentConfig.new instead')
  static AgentConfig create() => AgentConfig._();
  static $pb.GeneratedMessage $_createMessage() => AgentConfig._();
  @$core.override
  AgentConfig createEmptyInstance() => AgentConfig._();
  @$core.pragma('dart2js:noInline')
  static AgentConfig getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<AgentConfig>(
          AgentConfig.$_createMessage);
  static AgentConfig? _defaultInstance;

  /// Runtime contains execution behavior toggles that map to ADK run config.
  /// @gotags: json:"runtime,omitempty" yaml:"runtime,omitempty"
  @$pb.TagNumber(1)
  AgentRuntime get runtime => $_getN(0);
  @$pb.TagNumber(1)
  set runtime(AgentRuntime value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasRuntime() => $_has(0);
  @$pb.TagNumber(1)
  void clearRuntime() => $_clearField(1);
  @$pb.TagNumber(1)
  AgentRuntime ensureRuntime() => $_ensure(0);

  /// MCP servers define external tool providers this agent can connect to.
  /// @gotags: json:"mcp_servers,omitempty" yaml:"mcp_servers,omitempty"
  @$pb.TagNumber(2)
  $pb.PbList<MCPServer> get mcpServers => $_getList(1);

  /// Context guard configures automatic context window management.
  /// @gotags: json:"context_guard,omitempty" yaml:"context_guard,omitempty"
  @$pb.TagNumber(3)
  ContextGuardConfig get contextGuard => $_getN(2);
  @$pb.TagNumber(3)
  set contextGuard(ContextGuardConfig value) => $_setField(3, value);
  @$pb.TagNumber(3)
  $core.bool hasContextGuard() => $_has(2);
  @$pb.TagNumber(3)
  void clearContextGuard() => $_clearField(3);
  @$pb.TagNumber(3)
  ContextGuardConfig ensureContextGuard() => $_ensure(2);

  /// IDs referencing top-level shared MCP server configs.
  /// @gotags: json:"mcp_server_ids,omitempty" yaml:"mcp_server_ids,omitempty"
  @$pb.TagNumber(4)
  $pb.PbList<$core.String> get mcpServerIds => $_getList(3);

  /// IDs referencing top-level shared remote agent configs.
  /// @gotags: json:"remote_agent_ids,omitempty" yaml:"remote_agent_ids,omitempty"
  @$pb.TagNumber(5)
  $pb.PbList<$core.String> get remoteAgentIds => $_getList(4);

  /// File spaces mounted into this agent's built-in agent_files toolset.
  /// @gotags: json:"file_mounts,omitempty" yaml:"file_mounts,omitempty"
  @$pb.TagNumber(6)
  $pb.PbList<$1.AgentFileMount> get fileMounts => $_getList(5);

  /// Names of workspace-level Skills attached to this agent's built-in skill
  /// toolset. Empty means no skill toolset is attached.
  /// @gotags: json:"skills,omitempty" yaml:"skills,omitempty"
  @$pb.TagNumber(7)
  $pb.PbList<$core.String> get skills => $_getList(6);

  /// Memory configures mem0-backed Workspace/Agent Memory (ADR-0013). Only
  /// the invocation's root agent's config applies; PI and CURSOR agents
  /// reject it on write.
  /// @gotags: json:"memory,omitempty" yaml:"memory,omitempty"
  @$pb.TagNumber(8)
  MemoryConfig get memory => $_getN(7);
  @$pb.TagNumber(8)
  set memory(MemoryConfig value) => $_setField(8, value);
  @$pb.TagNumber(8)
  $core.bool hasMemory() => $_has(7);
  @$pb.TagNumber(8)
  void clearMemory() => $_clearField(8);
  @$pb.TagNumber(8)
  MemoryConfig ensureMemory() => $_ensure(7);

  /// Card Policy for Result Cards (ADR-0014). It narrows down the agent tree;
  /// PI and CURSOR agents reject it on write.
  /// @gotags: json:"result_cards,omitempty" yaml:"result_cards,omitempty"
  @$pb.TagNumber(9)
  ResultCardConfig get resultCards => $_getN(8);
  @$pb.TagNumber(9)
  set resultCards(ResultCardConfig value) => $_setField(9, value);
  @$pb.TagNumber(9)
  $core.bool hasResultCards() => $_has(8);
  @$pb.TagNumber(9)
  void clearResultCards() => $_clearField(9);
  @$pb.TagNumber(9)
  ResultCardConfig ensureResultCards() => $_ensure(8);

  /// Model identifier (for example: "gemini-2.5-pro").
  /// @gotags: json:"model,omitempty" yaml:"model,omitempty"
  @$pb.TagNumber(20)
  $core.String get model => $_getSZ(9);
  @$pb.TagNumber(20)
  set model($core.String value) => $_setString(9, value);
  @$pb.TagNumber(20)
  $core.bool hasModel() => $_has(9);
  @$pb.TagNumber(20)
  void clearModel() => $_clearField(20);

  /// Instruction guides agent behavior for this node.
  /// @gotags: json:"instruction,omitempty" yaml:"instruction,omitempty"
  @$pb.TagNumber(21)
  $core.String get instruction => $_getSZ(10);
  @$pb.TagNumber(21)
  set instruction($core.String value) => $_setString(10, value);
  @$pb.TagNumber(21)
  $core.bool hasInstruction() => $_has(10);
  @$pb.TagNumber(21)
  void clearInstruction() => $_clearField(21);

  /// Global instruction applies across the entire agent tree (root takes effect).
  /// @gotags: json:"global_instruction,omitempty" yaml:"global_instruction,omitempty"
  @$pb.TagNumber(22)
  $core.String get globalInstruction => $_getSZ(11);
  @$pb.TagNumber(22)
  set globalInstruction($core.String value) => $_setString(11, value);
  @$pb.TagNumber(22)
  $core.bool hasGlobalInstruction() => $_has(11);
  @$pb.TagNumber(22)
  void clearGlobalInstruction() => $_clearField(22);

  /// Prevents transferring control to parent/peer agents.
  /// @gotags: json:"disallow_transfer_to_parent,omitempty" yaml:"disallow_transfer_to_parent,omitempty"
  @$pb.TagNumber(23)
  $core.bool get disallowTransferToParent => $_getBF(12);
  @$pb.TagNumber(23)
  set disallowTransferToParent($core.bool value) => $_setBool(12, value);
  @$pb.TagNumber(23)
  $core.bool hasDisallowTransferToParent() => $_has(12);
  @$pb.TagNumber(23)
  void clearDisallowTransferToParent() => $_clearField(23);

  /// @gotags: json:"disallow_transfer_to_peers,omitempty" yaml:"disallow_transfer_to_peers,omitempty"
  @$pb.TagNumber(24)
  $core.bool get disallowTransferToPeers => $_getBF(13);
  @$pb.TagNumber(24)
  set disallowTransferToPeers($core.bool value) => $_setBool(13, value);
  @$pb.TagNumber(24)
  $core.bool hasDisallowTransferToPeers() => $_has(13);
  @$pb.TagNumber(24)
  void clearDisallowTransferToPeers() => $_clearField(24);

  /// Controls whether prior conversation contents are included in model requests.
  /// @gotags: json:"include_contents,omitempty" yaml:"include_contents,omitempty"
  @$pb.TagNumber(25)
  LLMIncludeContents get includeContents => $_getN(14);
  @$pb.TagNumber(25)
  set includeContents(LLMIncludeContents value) => $_setField(25, value);
  @$pb.TagNumber(25)
  $core.bool hasIncludeContents() => $_has(14);
  @$pb.TagNumber(25)
  void clearIncludeContents() => $_clearField(25);

  /// Optional session state key to store agent output.
  /// @gotags: json:"output_key,omitempty" yaml:"output_key,omitempty"
  @$pb.TagNumber(26)
  $core.String get outputKey => $_getSZ(15);
  @$pb.TagNumber(26)
  set outputKey($core.String value) => $_setString(15, value);
  @$pb.TagNumber(26)
  $core.bool hasOutputKey() => $_has(15);
  @$pb.TagNumber(26)
  void clearOutputKey() => $_clearField(26);

  /// Optional schemas encoded as JSON text.
  /// @gotags: json:"input_schema_json,omitempty" yaml:"input_schema_json,omitempty"
  @$pb.TagNumber(27)
  $core.String get inputSchemaJson => $_getSZ(16);
  @$pb.TagNumber(27)
  set inputSchemaJson($core.String value) => $_setString(16, value);
  @$pb.TagNumber(27)
  $core.bool hasInputSchemaJson() => $_has(16);
  @$pb.TagNumber(27)
  void clearInputSchemaJson() => $_clearField(27);

  /// @gotags: json:"output_schema_json,omitempty" yaml:"output_schema_json,omitempty"
  @$pb.TagNumber(28)
  $core.String get outputSchemaJson => $_getSZ(17);
  @$pb.TagNumber(28)
  set outputSchemaJson($core.String value) => $_setString(17, value);
  @$pb.TagNumber(28)
  $core.bool hasOutputSchemaJson() => $_has(17);
  @$pb.TagNumber(28)
  void clearOutputSchemaJson() => $_clearField(28);

  /// Max loop iterations. 0 means run until escalation/termination.
  /// @gotags: json:"max_iterations,omitempty" yaml:"max_iterations,omitempty"
  @$pb.TagNumber(30)
  $core.int get maxIterations => $_getIZ(18);
  @$pb.TagNumber(30)
  set maxIterations($core.int value) => $_setUnsignedInt32(18, value);
  @$pb.TagNumber(30)
  $core.bool hasMaxIterations() => $_has(18);
  @$pb.TagNumber(30)
  void clearMaxIterations() => $_clearField(30);

  /// Workflow graph of nodes and edges executed by the ADK workflow engine.
  /// @gotags: json:"workflow,omitempty" yaml:"workflow,omitempty"
  @$pb.TagNumber(40)
  WorkflowConfig get workflow => $_getN(19);
  @$pb.TagNumber(40)
  set workflow(WorkflowConfig value) => $_setField(40, value);
  @$pb.TagNumber(40)
  $core.bool hasWorkflow() => $_has(19);
  @$pb.TagNumber(40)
  void clearWorkflow() => $_clearField(40);
  @$pb.TagNumber(40)
  WorkflowConfig ensureWorkflow() => $_ensure(19);

  /// Binding of a PI agent to one ButterBox and one working directory.
  /// Required when type is PI; meaningless otherwise.
  /// @gotags: json:"pi,omitempty" yaml:"pi,omitempty"
  @$pb.TagNumber(50)
  PiAgentConfig get pi => $_getN(20);
  @$pb.TagNumber(50)
  set pi(PiAgentConfig value) => $_setField(50, value);
  @$pb.TagNumber(50)
  $core.bool hasPi() => $_has(20);
  @$pb.TagNumber(50)
  void clearPi() => $_clearField(50);
  @$pb.TagNumber(50)
  PiAgentConfig ensurePi() => $_ensure(20);

  /// Binding of a CURSOR agent to one ButterBox and one working directory.
  /// Required when type is CURSOR; meaningless otherwise.
  /// @gotags: json:"cursor,omitempty" yaml:"cursor,omitempty"
  @$pb.TagNumber(51)
  CursorAgentConfig get cursor => $_getN(21);
  @$pb.TagNumber(51)
  set cursor(CursorAgentConfig value) => $_setField(51, value);
  @$pb.TagNumber(51)
  $core.bool hasCursor() => $_has(21);
  @$pb.TagNumber(51)
  void clearCursor() => $_clearField(51);
  @$pb.TagNumber(51)
  CursorAgentConfig ensureCursor() => $_ensure(21);
}

/// CursorAgentConfig binds a CURSOR agent to the ButterBox that runs its
/// Cursor SDK Bridge sessions (CursorService) and to one working directory on
/// that box. Like a PI agent, a CURSOR agent's tools and instructions are
/// configured on the box — the working directory's .cursor/rules and
/// mcp.json — and the Cursor API key lives in the box's environment, never in
/// butter.
class CursorAgentConfig extends $pb.GeneratedMessage {
  factory CursorAgentConfig({
    $core.String? butterboxId,
    $core.String? workingDir,
    $core.String? model,
    $core.String? mode,
    $core.int? maxRunSeconds,
  }) {
    final result = CursorAgentConfig._();
    if (butterboxId != null) result.butterboxId = butterboxId;
    if (workingDir != null) result.workingDir = workingDir;
    if (model != null) result.model = model;
    if (mode != null) result.mode = mode;
    if (maxRunSeconds != null) result.maxRunSeconds = maxRunSeconds;
    return result;
  }

  CursorAgentConfig._();

  factory CursorAgentConfig.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CursorAgentConfig()..mergeFromBuffer(data, registry);
  factory CursorAgentConfig.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      CursorAgentConfig()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CursorAgentConfig',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: CursorAgentConfig.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'butterboxId')
    ..aOS(2, _omitFieldNames ? '' : 'workingDir')
    ..aOS(3, _omitFieldNames ? '' : 'model')
    ..aOS(4, _omitFieldNames ? '' : 'mode')
    ..aI(5, _omitFieldNames ? '' : 'maxRunSeconds')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CursorAgentConfig clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CursorAgentConfig copyWith(void Function(CursorAgentConfig) updates) =>
      super.copyWith((message) => updates(message as CursorAgentConfig))
          as CursorAgentConfig;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use CursorAgentConfig() / CursorAgentConfig.new instead')
  static CursorAgentConfig create() => CursorAgentConfig._();
  static $pb.GeneratedMessage $_createMessage() => CursorAgentConfig._();
  @$core.override
  CursorAgentConfig createEmptyInstance() => CursorAgentConfig._();
  @$core.pragma('dart2js:noInline')
  static CursorAgentConfig getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<CursorAgentConfig>(
          CursorAgentConfig.$_createMessage);
  static CursorAgentConfig? _defaultInstance;

  /// The workspace ButterBox that hosts this agent's Cursor sessions.
  /// @gotags: json:"butterbox_id,omitempty" yaml:"butterbox_id,omitempty"
  @$pb.TagNumber(1)
  $core.String get butterboxId => $_getSZ(0);
  @$pb.TagNumber(1)
  set butterboxId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasButterboxId() => $_has(0);
  @$pb.TagNumber(1)
  void clearButterboxId() => $_clearField(1);

  /// Working directory for Cursor sessions, absolute or relative to the box's
  /// sandbox root. Empty uses the box server's working directory.
  /// @gotags: json:"working_dir,omitempty" yaml:"working_dir,omitempty"
  @$pb.TagNumber(2)
  $core.String get workingDir => $_getSZ(1);
  @$pb.TagNumber(2)
  set workingDir($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasWorkingDir() => $_has(1);
  @$pb.TagNumber(2)
  void clearWorkingDir() => $_clearField(2);

  /// Optional Cursor model ID (e.g. "composer-2.5"). Empty uses the bridge
  /// default. The dropdown is fed by ButterBoxService.ListButterBoxCursorModels.
  /// @gotags: json:"model,omitempty" yaml:"model,omitempty"
  @$pb.TagNumber(3)
  $core.String get model => $_getSZ(2);
  @$pb.TagNumber(3)
  set model($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasModel() => $_has(2);
  @$pb.TagNumber(3)
  void clearModel() => $_clearField(3);

  /// Agent mode: "agent" (default when empty) or "plan".
  /// @gotags: json:"mode,omitempty" yaml:"mode,omitempty"
  @$pb.TagNumber(4)
  $core.String get mode => $_getSZ(3);
  @$pb.TagNumber(4)
  set mode($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasMode() => $_has(3);
  @$pb.TagNumber(4)
  void clearMode() => $_clearField(4);

  /// Bounds one turn's run on the box; exceeding it aborts the run. Unset
  /// defaults to 1800 seconds; an explicit 0 means unlimited.
  /// @gotags: json:"max_run_seconds,omitempty" yaml:"max_run_seconds,omitempty"
  @$pb.TagNumber(5)
  $core.int get maxRunSeconds => $_getIZ(4);
  @$pb.TagNumber(5)
  set maxRunSeconds($core.int value) => $_setSignedInt32(4, value);
  @$pb.TagNumber(5)
  $core.bool hasMaxRunSeconds() => $_has(4);
  @$pb.TagNumber(5)
  void clearMaxRunSeconds() => $_clearField(5);
}

/// PiAgentConfig binds a PI agent to the ButterBox that runs its pi coding
/// agent and to one working directory on that box (ADR-0011). A PI agent's
/// tools and instructions are configured on the box — in the working
/// directory's AGENTS.md and .pi/ — not in butter, so the LLM/MCP/skill
/// fields of AgentConfig are rejected on write for PI agents.
class PiAgentConfig extends $pb.GeneratedMessage {
  factory PiAgentConfig({
    $core.String? butterboxId,
    $core.String? workingDir,
    $core.String? provider,
    $core.String? model,
    $core.String? thinkingLevel,
    $core.int? maxRunSeconds,
  }) {
    final result = PiAgentConfig._();
    if (butterboxId != null) result.butterboxId = butterboxId;
    if (workingDir != null) result.workingDir = workingDir;
    if (provider != null) result.provider = provider;
    if (model != null) result.model = model;
    if (thinkingLevel != null) result.thinkingLevel = thinkingLevel;
    if (maxRunSeconds != null) result.maxRunSeconds = maxRunSeconds;
    return result;
  }

  PiAgentConfig._();

  factory PiAgentConfig.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      PiAgentConfig()..mergeFromBuffer(data, registry);
  factory PiAgentConfig.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      PiAgentConfig()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'PiAgentConfig',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: PiAgentConfig.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'butterboxId')
    ..aOS(2, _omitFieldNames ? '' : 'workingDir')
    ..aOS(3, _omitFieldNames ? '' : 'provider')
    ..aOS(4, _omitFieldNames ? '' : 'model')
    ..aOS(5, _omitFieldNames ? '' : 'thinkingLevel')
    ..aI(6, _omitFieldNames ? '' : 'maxRunSeconds')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  PiAgentConfig clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  PiAgentConfig copyWith(void Function(PiAgentConfig) updates) =>
      super.copyWith((message) => updates(message as PiAgentConfig))
          as PiAgentConfig;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use PiAgentConfig() / PiAgentConfig.new instead')
  static PiAgentConfig create() => PiAgentConfig._();
  static $pb.GeneratedMessage $_createMessage() => PiAgentConfig._();
  @$core.override
  PiAgentConfig createEmptyInstance() => PiAgentConfig._();
  @$core.pragma('dart2js:noInline')
  static PiAgentConfig getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<PiAgentConfig>(
          PiAgentConfig.$_createMessage);
  static PiAgentConfig? _defaultInstance;

  /// The workspace ButterBox that hosts this agent's pi sessions.
  /// @gotags: json:"butterbox_id,omitempty" yaml:"butterbox_id,omitempty"
  @$pb.TagNumber(1)
  $core.String get butterboxId => $_getSZ(0);
  @$pb.TagNumber(1)
  set butterboxId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasButterboxId() => $_has(0);
  @$pb.TagNumber(1)
  void clearButterboxId() => $_clearField(1);

  /// Working directory for pi sessions, absolute or relative to the box's
  /// sandbox root. Empty uses the box server's working directory.
  /// @gotags: json:"working_dir,omitempty" yaml:"working_dir,omitempty"
  @$pb.TagNumber(2)
  $core.String get workingDir => $_getSZ(1);
  @$pb.TagNumber(2)
  set workingDir($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasWorkingDir() => $_has(1);
  @$pb.TagNumber(2)
  void clearWorkingDir() => $_clearField(2);

  /// Optional pi model provider (e.g. "anthropic"). Empty uses pi's default
  /// on the box. Model identity is pi's, not butter's: ModelProviders are
  /// not involved.
  /// @gotags: json:"provider,omitempty" yaml:"provider,omitempty"
  @$pb.TagNumber(3)
  $core.String get provider => $_getSZ(2);
  @$pb.TagNumber(3)
  set provider($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasProvider() => $_has(2);
  @$pb.TagNumber(3)
  void clearProvider() => $_clearField(3);

  /// Optional model pattern or ID as accepted by pi's --model flag. The
  /// dropdown is fed by ButterBoxService.ListButterBoxModels.
  /// @gotags: json:"model,omitempty" yaml:"model,omitempty"
  @$pb.TagNumber(4)
  $core.String get model => $_getSZ(3);
  @$pb.TagNumber(4)
  set model($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasModel() => $_has(3);
  @$pb.TagNumber(4)
  void clearModel() => $_clearField(4);

  /// Optional thinking level: off, minimal, low, medium, high, xhigh, max.
  /// @gotags: json:"thinking_level,omitempty" yaml:"thinking_level,omitempty"
  @$pb.TagNumber(5)
  $core.String get thinkingLevel => $_getSZ(4);
  @$pb.TagNumber(5)
  set thinkingLevel($core.String value) => $_setString(4, value);
  @$pb.TagNumber(5)
  $core.bool hasThinkingLevel() => $_has(4);
  @$pb.TagNumber(5)
  void clearThinkingLevel() => $_clearField(5);

  /// Bounds one turn's run on the box; exceeding it aborts the run. Unset
  /// defaults to 1800 seconds; an explicit 0 means unlimited. The bound
  /// exists to stop runaway loops from holding a session lease and a box
  /// process slot forever — long legitimate runs raise the number.
  /// @gotags: json:"max_run_seconds,omitempty" yaml:"max_run_seconds,omitempty"
  @$pb.TagNumber(6)
  $core.int get maxRunSeconds => $_getIZ(5);
  @$pb.TagNumber(6)
  set maxRunSeconds($core.int value) => $_setSignedInt32(5, value);
  @$pb.TagNumber(6)
  $core.bool hasMaxRunSeconds() => $_has(5);
  @$pb.TagNumber(6)
  void clearMaxRunSeconds() => $_clearField(6);
}

/// WorkflowConfig declares a Workflow Agent's directed graph as two explicit
/// lists: nodes (name, kind, per-node options) and edges (from/to node names
/// plus an optional route label). See ADR 0001.
class WorkflowConfig extends $pb.GeneratedMessage {
  factory WorkflowConfig({
    $core.Iterable<WorkflowNode>? nodes,
    $core.Iterable<WorkflowEdge>? edges,
  }) {
    final result = WorkflowConfig._();
    if (nodes != null) result.nodes.addAll(nodes);
    if (edges != null) result.edges.addAll(edges);
    return result;
  }

  WorkflowConfig._();

  factory WorkflowConfig.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      WorkflowConfig()..mergeFromBuffer(data, registry);
  factory WorkflowConfig.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      WorkflowConfig()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'WorkflowConfig',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: WorkflowConfig.$_createMessage)
    ..pPM<WorkflowNode>(1, _omitFieldNames ? '' : 'nodes',
        subBuilder: WorkflowNode.$_createMessage)
    ..pPM<WorkflowEdge>(2, _omitFieldNames ? '' : 'edges',
        subBuilder: WorkflowEdge.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  WorkflowConfig clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  WorkflowConfig copyWith(void Function(WorkflowConfig) updates) =>
      super.copyWith((message) => updates(message as WorkflowConfig))
          as WorkflowConfig;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use WorkflowConfig() / WorkflowConfig.new instead')
  static WorkflowConfig create() => WorkflowConfig._();
  static $pb.GeneratedMessage $_createMessage() => WorkflowConfig._();
  @$core.override
  WorkflowConfig createEmptyInstance() => WorkflowConfig._();
  @$core.pragma('dart2js:noInline')
  static WorkflowConfig getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<WorkflowConfig>(
          WorkflowConfig.$_createMessage);
  static WorkflowConfig? _defaultInstance;

  /// Nodes are the steps of the graph. Node names must be unique and must not
  /// use the reserved "START" sentinel.
  /// @gotags: json:"nodes,omitempty" yaml:"nodes,omitempty"
  @$pb.TagNumber(1)
  $pb.PbList<WorkflowNode> get nodes => $_getList(0);

  /// Edges are directed connections between declared nodes. The reserved name
  /// "START" may appear in `from` to mark the graph's entry point(s).
  /// @gotags: json:"edges,omitempty" yaml:"edges,omitempty"
  @$pb.TagNumber(2)
  $pb.PbList<WorkflowEdge> get edges => $_getList(1);
}

/// WorkflowNode is a single step in a Workflow Agent's graph.
class WorkflowNode extends $pb.GeneratedMessage {
  factory WorkflowNode({
    $core.String? name,
    WorkflowNodeKind? kind,
    @$core.Deprecated('This field is deprecated.') $core.String? agent,
    $core.String? question,
    $core.String? agentId,
    $core.bool? parallelWorker,
    WorkflowRetryConfig? retry,
    $core.int? timeoutSeconds,
    HumanInputForm? form,
  }) {
    final result = WorkflowNode._();
    if (name != null) result.name = name;
    if (kind != null) result.kind = kind;
    if (agent != null) result.agent = agent;
    if (question != null) result.question = question;
    if (agentId != null) result.agentId = agentId;
    if (parallelWorker != null) result.parallelWorker = parallelWorker;
    if (retry != null) result.retry = retry;
    if (timeoutSeconds != null) result.timeoutSeconds = timeoutSeconds;
    if (form != null) result.form = form;
    return result;
  }

  WorkflowNode._();

  factory WorkflowNode.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      WorkflowNode()..mergeFromBuffer(data, registry);
  factory WorkflowNode.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      WorkflowNode()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'WorkflowNode',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: WorkflowNode.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'name')
    ..aE<WorkflowNodeKind>(2, _omitFieldNames ? '' : 'kind',
        enumValues: WorkflowNodeKind.values)
    ..aOS(3, _omitFieldNames ? '' : 'agent')
    ..aOS(4, _omitFieldNames ? '' : 'question')
    ..aOS(5, _omitFieldNames ? '' : 'agentId')
    ..aOB(10, _omitFieldNames ? '' : 'parallelWorker')
    ..aOM<WorkflowRetryConfig>(11, _omitFieldNames ? '' : 'retry',
        subBuilder: WorkflowRetryConfig.$_createMessage)
    ..aI(12, _omitFieldNames ? '' : 'timeoutSeconds')
    ..aOM<HumanInputForm>(13, _omitFieldNames ? '' : 'form',
        subBuilder: HumanInputForm.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  WorkflowNode clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  WorkflowNode copyWith(void Function(WorkflowNode) updates) =>
      super.copyWith((message) => updates(message as WorkflowNode))
          as WorkflowNode;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use WorkflowNode() / WorkflowNode.new instead')
  static WorkflowNode create() => WorkflowNode._();
  static $pb.GeneratedMessage $_createMessage() => WorkflowNode._();
  @$core.override
  WorkflowNode createEmptyInstance() => WorkflowNode._();
  @$core.pragma('dart2js:noInline')
  static WorkflowNode getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<WorkflowNode>(
          WorkflowNode.$_createMessage);
  static WorkflowNode? _defaultInstance;

  /// Name identifies the node within the graph. Must be unique; "START" is
  /// reserved for the entry sentinel used in edges.
  /// @gotags: json:"name,omitempty" yaml:"name,omitempty"
  @$pb.TagNumber(1)
  $core.String get name => $_getSZ(0);
  @$pb.TagNumber(1)
  set name($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasName() => $_has(0);
  @$pb.TagNumber(1)
  void clearName() => $_clearField(1);

  /// Kind selects the node behaviour.
  /// @gotags: json:"kind,omitempty" yaml:"kind,omitempty"
  @$pb.TagNumber(2)
  WorkflowNodeKind get kind => $_getN(1);
  @$pb.TagNumber(2)
  set kind(WorkflowNodeKind value) => $_setField(2, value);
  @$pb.TagNumber(2)
  $core.bool hasKind() => $_has(1);
  @$pb.TagNumber(2)
  void clearKind() => $_clearField(2);

  /// Legacy AGENT node reference by runtime name. Deprecated (issue #241):
  /// never resolved — AGENT nodes must reference their target via agent_id.
  /// Retained only so historical records decode.
  /// @gotags: json:"agent,omitempty" yaml:"agent,omitempty"
  @$core.Deprecated('This field is deprecated.')
  @$pb.TagNumber(3)
  $core.String get agent => $_getSZ(2);
  @$core.Deprecated('This field is deprecated.')
  @$pb.TagNumber(3)
  set agent($core.String value) => $_setString(2, value);
  @$core.Deprecated('This field is deprecated.')
  @$pb.TagNumber(3)
  $core.bool hasAgent() => $_has(2);
  @$core.Deprecated('This field is deprecated.')
  @$pb.TagNumber(3)
  void clearAgent() => $_clearField(3);

  /// HUMAN_INPUT nodes: the question presented to the human.
  /// @gotags: json:"question,omitempty" yaml:"question,omitempty"
  @$pb.TagNumber(4)
  $core.String get question => $_getSZ(3);
  @$pb.TagNumber(4)
  set question($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasQuestion() => $_has(3);
  @$pb.TagNumber(4)
  void clearQuestion() => $_clearField(4);

  /// AGENT nodes (V2): references an independent Agent by its workspace-scoped
  /// agent_id. Preferred over the legacy name-based `agent` field.
  /// @gotags: json:"agent_id,omitempty" yaml:"agent_id,omitempty"
  @$pb.TagNumber(5)
  $core.String get agentId => $_getSZ(4);
  @$pb.TagNumber(5)
  set agentId($core.String value) => $_setString(4, value);
  @$pb.TagNumber(5)
  $core.bool hasAgentId() => $_has(4);
  @$pb.TagNumber(5)
  void clearAgentId() => $_clearField(5);

  /// Runs the node once per item of a list-typed input, concurrently, then
  /// aggregates outputs.
  /// @gotags: json:"parallel_worker,omitempty" yaml:"parallel_worker,omitempty"
  @$pb.TagNumber(10)
  $core.bool get parallelWorker => $_getBF(5);
  @$pb.TagNumber(10)
  set parallelWorker($core.bool value) => $_setBool(5, value);
  @$pb.TagNumber(10)
  $core.bool hasParallelWorker() => $_has(5);
  @$pb.TagNumber(10)
  void clearParallelWorker() => $_clearField(10);

  /// Retry policy for failed activations. Unset means no retries.
  /// @gotags: json:"retry,omitempty" yaml:"retry,omitempty"
  @$pb.TagNumber(11)
  WorkflowRetryConfig get retry => $_getN(6);
  @$pb.TagNumber(11)
  set retry(WorkflowRetryConfig value) => $_setField(11, value);
  @$pb.TagNumber(11)
  $core.bool hasRetry() => $_has(6);
  @$pb.TagNumber(11)
  void clearRetry() => $_clearField(11);
  @$pb.TagNumber(11)
  WorkflowRetryConfig ensureRetry() => $_ensure(6);

  /// Bounds a single activation of the node. 0 means no per-node timeout.
  /// @gotags: json:"timeout_seconds,omitempty" yaml:"timeout_seconds,omitempty"
  @$pb.TagNumber(12)
  $core.int get timeoutSeconds => $_getIZ(7);
  @$pb.TagNumber(12)
  set timeoutSeconds($core.int value) => $_setSignedInt32(7, value);
  @$pb.TagNumber(12)
  $core.bool hasTimeoutSeconds() => $_has(7);
  @$pb.TagNumber(12)
  void clearTimeoutSeconds() => $_clearField(12);

  /// HUMAN_INPUT nodes: optional form presentation of the question. An
  /// A2UI-capable AG-UI client renders it as fields and submits a structured
  /// answer; every other entry point receives the question followed by field
  /// instructions and may answer in plain text. The form is a presentation
  /// and submission format, not an ADK response schema: the successor still
  /// receives a string. Unset keeps the plain-text question.
  /// @gotags: json:"form,omitempty" yaml:"form,omitempty"
  @$pb.TagNumber(13)
  HumanInputForm get form => $_getN(8);
  @$pb.TagNumber(13)
  set form(HumanInputForm value) => $_setField(13, value);
  @$pb.TagNumber(13)
  $core.bool hasForm() => $_has(8);
  @$pb.TagNumber(13)
  void clearForm() => $_clearField(13);
  @$pb.TagNumber(13)
  HumanInputForm ensureForm() => $_ensure(8);
}

/// HumanInputForm is the form a HUMAN_INPUT node presents to an A2UI-capable
/// client. Its definition is frozen into the request-input event when the node
/// pauses, so later edits never change the rules of a form already shown.
class HumanInputForm extends $pb.GeneratedMessage {
  factory HumanInputForm({
    $core.String? title,
    $core.Iterable<HumanInputFormField>? fields,
  }) {
    final result = HumanInputForm._();
    if (title != null) result.title = title;
    if (fields != null) result.fields.addAll(fields);
    return result;
  }

  HumanInputForm._();

  factory HumanInputForm.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      HumanInputForm()..mergeFromBuffer(data, registry);
  factory HumanInputForm.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      HumanInputForm()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'HumanInputForm',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: HumanInputForm.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'title')
    ..pPM<HumanInputFormField>(2, _omitFieldNames ? '' : 'fields',
        subBuilder: HumanInputFormField.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  HumanInputForm clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  HumanInputForm copyWith(void Function(HumanInputForm) updates) =>
      super.copyWith((message) => updates(message as HumanInputForm))
          as HumanInputForm;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use HumanInputForm() / HumanInputForm.new instead')
  static HumanInputForm create() => HumanInputForm._();
  static $pb.GeneratedMessage $_createMessage() => HumanInputForm._();
  @$core.override
  HumanInputForm createEmptyInstance() => HumanInputForm._();
  @$core.pragma('dart2js:noInline')
  static HumanInputForm getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<HumanInputForm>(
          HumanInputForm.$_createMessage);
  static HumanInputForm? _defaultInstance;

  /// Heading shown above the fields. Empty uses the node's question.
  /// @gotags: json:"title,omitempty" yaml:"title,omitempty"
  @$pb.TagNumber(1)
  $core.String get title => $_getSZ(0);
  @$pb.TagNumber(1)
  set title($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasTitle() => $_has(0);
  @$pb.TagNumber(1)
  void clearTitle() => $_clearField(1);

  /// Ordered fields, 1 to 20. The submitted answer is a JSON object with one
  /// string member per field, in this order.
  /// @gotags: json:"fields,omitempty" yaml:"fields,omitempty"
  @$pb.TagNumber(2)
  $pb.PbList<HumanInputFormField> get fields => $_getList(1);
}

/// HumanInputFormField is one field of a Human Input form.
class HumanInputFormField extends $pb.GeneratedMessage {
  factory HumanInputFormField({
    $core.String? name,
    $core.String? label,
    $core.String? hint,
    $core.bool? required,
    HumanInputFormFieldType? type,
    $core.int? maxLength,
    $core.Iterable<HumanInputFormOption>? options,
  }) {
    final result = HumanInputFormField._();
    if (name != null) result.name = name;
    if (label != null) result.label = label;
    if (hint != null) result.hint = hint;
    if (required != null) result.required = required;
    if (type != null) result.type = type;
    if (maxLength != null) result.maxLength = maxLength;
    if (options != null) result.options.addAll(options);
    return result;
  }

  HumanInputFormField._();

  factory HumanInputFormField.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      HumanInputFormField()..mergeFromBuffer(data, registry);
  factory HumanInputFormField.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      HumanInputFormField()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'HumanInputFormField',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: HumanInputFormField.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'name')
    ..aOS(2, _omitFieldNames ? '' : 'label')
    ..aOS(3, _omitFieldNames ? '' : 'hint')
    ..aOB(4, _omitFieldNames ? '' : 'required')
    ..aE<HumanInputFormFieldType>(5, _omitFieldNames ? '' : 'type',
        enumValues: HumanInputFormFieldType.values)
    ..aI(6, _omitFieldNames ? '' : 'maxLength')
    ..pPM<HumanInputFormOption>(7, _omitFieldNames ? '' : 'options',
        subBuilder: HumanInputFormOption.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  HumanInputFormField clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  HumanInputFormField copyWith(void Function(HumanInputFormField) updates) =>
      super.copyWith((message) => updates(message as HumanInputFormField))
          as HumanInputFormField;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core
      .Deprecated('Use HumanInputFormField() / HumanInputFormField.new instead')
  static HumanInputFormField create() => HumanInputFormField._();
  static $pb.GeneratedMessage $_createMessage() => HumanInputFormField._();
  @$core.override
  HumanInputFormField createEmptyInstance() => HumanInputFormField._();
  @$core.pragma('dart2js:noInline')
  static HumanInputFormField getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<HumanInputFormField>(
          HumanInputFormField.$_createMessage);
  static HumanInputFormField? _defaultInstance;

  /// Key of this field in the submitted JSON object. Unique within the form;
  /// letters, digits and underscores, starting with a letter or underscore.
  /// @gotags: json:"name,omitempty" yaml:"name,omitempty"
  @$pb.TagNumber(1)
  $core.String get name => $_getSZ(0);
  @$pb.TagNumber(1)
  set name($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasName() => $_has(0);
  @$pb.TagNumber(1)
  void clearName() => $_clearField(1);

  /// Label shown next to the input.
  /// @gotags: json:"label,omitempty" yaml:"label,omitempty"
  @$pb.TagNumber(2)
  $core.String get label => $_getSZ(1);
  @$pb.TagNumber(2)
  set label($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasLabel() => $_has(1);
  @$pb.TagNumber(2)
  void clearLabel() => $_clearField(2);

  /// Optional helper text shown under the input.
  /// @gotags: json:"hint,omitempty" yaml:"hint,omitempty"
  @$pb.TagNumber(3)
  $core.String get hint => $_getSZ(2);
  @$pb.TagNumber(3)
  set hint($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasHint() => $_has(2);
  @$pb.TagNumber(3)
  void clearHint() => $_clearField(3);

  /// Whether the field must be answered before the form can be submitted.
  /// @gotags: json:"required,omitempty" yaml:"required,omitempty"
  @$pb.TagNumber(4)
  $core.bool get required => $_getBF(3);
  @$pb.TagNumber(4)
  set required($core.bool value) => $_setBool(3, value);
  @$pb.TagNumber(4)
  $core.bool hasRequired() => $_has(3);
  @$pb.TagNumber(4)
  void clearRequired() => $_clearField(4);

  /// How the field is answered.
  /// @gotags: json:"type,omitempty" yaml:"type,omitempty"
  @$pb.TagNumber(5)
  HumanInputFormFieldType get type => $_getN(4);
  @$pb.TagNumber(5)
  set type(HumanInputFormFieldType value) => $_setField(5, value);
  @$pb.TagNumber(5)
  $core.bool hasType() => $_has(4);
  @$pb.TagNumber(5)
  void clearType() => $_clearField(5);

  /// TEXT fields: maximum answer length in characters, at most 2000. 0 uses
  /// 2000. Must be 0 for SINGLE_CHOICE fields.
  /// @gotags: json:"max_length,omitempty" yaml:"max_length,omitempty"
  @$pb.TagNumber(6)
  $core.int get maxLength => $_getIZ(5);
  @$pb.TagNumber(6)
  set maxLength($core.int value) => $_setSignedInt32(5, value);
  @$pb.TagNumber(6)
  $core.bool hasMaxLength() => $_has(5);
  @$pb.TagNumber(6)
  void clearMaxLength() => $_clearField(6);

  /// SINGLE_CHOICE fields: the fixed options, 1 to 50, with unique values.
  /// Must be empty for TEXT fields.
  /// @gotags: json:"options,omitempty" yaml:"options,omitempty"
  @$pb.TagNumber(7)
  $pb.PbList<HumanInputFormOption> get options => $_getList(6);
}

/// HumanInputFormOption is one fixed choice of a SINGLE_CHOICE field.
class HumanInputFormOption extends $pb.GeneratedMessage {
  factory HumanInputFormOption({
    $core.String? value,
    $core.String? label,
  }) {
    final result = HumanInputFormOption._();
    if (value != null) result.value = value;
    if (label != null) result.label = label;
    return result;
  }

  HumanInputFormOption._();

  factory HumanInputFormOption.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      HumanInputFormOption()..mergeFromBuffer(data, registry);
  factory HumanInputFormOption.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      HumanInputFormOption()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'HumanInputFormOption',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: HumanInputFormOption.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'value')
    ..aOS(2, _omitFieldNames ? '' : 'label')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  HumanInputFormOption clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  HumanInputFormOption copyWith(void Function(HumanInputFormOption) updates) =>
      super.copyWith((message) => updates(message as HumanInputFormOption))
          as HumanInputFormOption;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use HumanInputFormOption() / HumanInputFormOption.new instead')
  static HumanInputFormOption create() => HumanInputFormOption._();
  static $pb.GeneratedMessage $_createMessage() => HumanInputFormOption._();
  @$core.override
  HumanInputFormOption createEmptyInstance() => HumanInputFormOption._();
  @$core.pragma('dart2js:noInline')
  static HumanInputFormOption getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<HumanInputFormOption>(
          HumanInputFormOption.$_createMessage);
  static HumanInputFormOption? _defaultInstance;

  /// Value submitted when the option is chosen.
  /// @gotags: json:"value,omitempty" yaml:"value,omitempty"
  @$pb.TagNumber(1)
  $core.String get value => $_getSZ(0);
  @$pb.TagNumber(1)
  set value($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasValue() => $_has(0);
  @$pb.TagNumber(1)
  void clearValue() => $_clearField(1);

  /// Text shown for the option.
  /// @gotags: json:"label,omitempty" yaml:"label,omitempty"
  @$pb.TagNumber(2)
  $core.String get label => $_getSZ(1);
  @$pb.TagNumber(2)
  set label($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasLabel() => $_has(1);
  @$pb.TagNumber(2)
  void clearLabel() => $_clearField(2);
}

/// WorkflowRetryConfig is a serializable subset of the ADK retry policy.
/// Unset fields fall back to the engine defaults.
class WorkflowRetryConfig extends $pb.GeneratedMessage {
  factory WorkflowRetryConfig({
    $core.int? maxAttempts,
    $core.int? initialDelaySeconds,
    $core.int? maxDelaySeconds,
    $core.double? backoffFactor,
  }) {
    final result = WorkflowRetryConfig._();
    if (maxAttempts != null) result.maxAttempts = maxAttempts;
    if (initialDelaySeconds != null)
      result.initialDelaySeconds = initialDelaySeconds;
    if (maxDelaySeconds != null) result.maxDelaySeconds = maxDelaySeconds;
    if (backoffFactor != null) result.backoffFactor = backoffFactor;
    return result;
  }

  WorkflowRetryConfig._();

  factory WorkflowRetryConfig.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      WorkflowRetryConfig()..mergeFromBuffer(data, registry);
  factory WorkflowRetryConfig.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      WorkflowRetryConfig()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'WorkflowRetryConfig',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: WorkflowRetryConfig.$_createMessage)
    ..aI(1, _omitFieldNames ? '' : 'maxAttempts')
    ..aI(2, _omitFieldNames ? '' : 'initialDelaySeconds')
    ..aI(3, _omitFieldNames ? '' : 'maxDelaySeconds')
    ..aD(4, _omitFieldNames ? '' : 'backoffFactor')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  WorkflowRetryConfig clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  WorkflowRetryConfig copyWith(void Function(WorkflowRetryConfig) updates) =>
      super.copyWith((message) => updates(message as WorkflowRetryConfig))
          as WorkflowRetryConfig;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core
      .Deprecated('Use WorkflowRetryConfig() / WorkflowRetryConfig.new instead')
  static WorkflowRetryConfig create() => WorkflowRetryConfig._();
  static $pb.GeneratedMessage $_createMessage() => WorkflowRetryConfig._();
  @$core.override
  WorkflowRetryConfig createEmptyInstance() => WorkflowRetryConfig._();
  @$core.pragma('dart2js:noInline')
  static WorkflowRetryConfig getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<WorkflowRetryConfig>(
          WorkflowRetryConfig.$_createMessage);
  static WorkflowRetryConfig? _defaultInstance;

  /// Maximum number of attempts including the first. 0 or 1 means no retries.
  /// @gotags: json:"max_attempts,omitempty" yaml:"max_attempts,omitempty"
  @$pb.TagNumber(1)
  $core.int get maxAttempts => $_getIZ(0);
  @$pb.TagNumber(1)
  set maxAttempts($core.int value) => $_setSignedInt32(0, value);
  @$pb.TagNumber(1)
  $core.bool hasMaxAttempts() => $_has(0);
  @$pb.TagNumber(1)
  void clearMaxAttempts() => $_clearField(1);

  /// Delay before the first retry, in seconds.
  /// @gotags: json:"initial_delay_seconds,omitempty" yaml:"initial_delay_seconds,omitempty"
  @$pb.TagNumber(2)
  $core.int get initialDelaySeconds => $_getIZ(1);
  @$pb.TagNumber(2)
  set initialDelaySeconds($core.int value) => $_setSignedInt32(1, value);
  @$pb.TagNumber(2)
  $core.bool hasInitialDelaySeconds() => $_has(1);
  @$pb.TagNumber(2)
  void clearInitialDelaySeconds() => $_clearField(2);

  /// Cap on the delay between retries, in seconds.
  /// @gotags: json:"max_delay_seconds,omitempty" yaml:"max_delay_seconds,omitempty"
  @$pb.TagNumber(3)
  $core.int get maxDelaySeconds => $_getIZ(2);
  @$pb.TagNumber(3)
  set maxDelaySeconds($core.int value) => $_setSignedInt32(2, value);
  @$pb.TagNumber(3)
  $core.bool hasMaxDelaySeconds() => $_has(2);
  @$pb.TagNumber(3)
  void clearMaxDelaySeconds() => $_clearField(3);

  /// Per-attempt multiplier applied to the delay.
  /// @gotags: json:"backoff_factor,omitempty" yaml:"backoff_factor,omitempty"
  @$pb.TagNumber(4)
  $core.double get backoffFactor => $_getN(3);
  @$pb.TagNumber(4)
  set backoffFactor($core.double value) => $_setDouble(3, value);
  @$pb.TagNumber(4)
  $core.bool hasBackoffFactor() => $_has(3);
  @$pb.TagNumber(4)
  void clearBackoffFactor() => $_clearField(4);
}

/// WorkflowEdge is a directed connection between two nodes, optionally guarded
/// by a Route label.
class WorkflowEdge extends $pb.GeneratedMessage {
  factory WorkflowEdge({
    $core.String? from,
    $core.String? to,
    $core.String? route,
    $core.bool? isDefault,
  }) {
    final result = WorkflowEdge._();
    if (from != null) result.from = from;
    if (to != null) result.to = to;
    if (route != null) result.route = route;
    if (isDefault != null) result.isDefault = isDefault;
    return result;
  }

  WorkflowEdge._();

  factory WorkflowEdge.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      WorkflowEdge()..mergeFromBuffer(data, registry);
  factory WorkflowEdge.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      WorkflowEdge()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'WorkflowEdge',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: WorkflowEdge.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'from')
    ..aOS(2, _omitFieldNames ? '' : 'to')
    ..aOS(3, _omitFieldNames ? '' : 'route')
    ..aOB(4, _omitFieldNames ? '' : 'isDefault')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  WorkflowEdge clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  WorkflowEdge copyWith(void Function(WorkflowEdge) updates) =>
      super.copyWith((message) => updates(message as WorkflowEdge))
          as WorkflowEdge;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use WorkflowEdge() / WorkflowEdge.new instead')
  static WorkflowEdge create() => WorkflowEdge._();
  static $pb.GeneratedMessage $_createMessage() => WorkflowEdge._();
  @$core.override
  WorkflowEdge createEmptyInstance() => WorkflowEdge._();
  @$core.pragma('dart2js:noInline')
  static WorkflowEdge getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<WorkflowEdge>(
          WorkflowEdge.$_createMessage);
  static WorkflowEdge? _defaultInstance;

  /// Source node name, or the reserved "START" sentinel for entry edges.
  /// @gotags: json:"from,omitempty" yaml:"from,omitempty"
  @$pb.TagNumber(1)
  $core.String get from => $_getSZ(0);
  @$pb.TagNumber(1)
  set from($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasFrom() => $_has(0);
  @$pb.TagNumber(1)
  void clearFrom() => $_clearField(1);

  /// Target node name. Must reference a declared node.
  /// @gotags: json:"to,omitempty" yaml:"to,omitempty"
  @$pb.TagNumber(2)
  $core.String get to => $_getSZ(1);
  @$pb.TagNumber(2)
  set to($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasTo() => $_has(1);
  @$pb.TagNumber(2)
  void clearTo() => $_clearField(2);

  /// Route label: the edge is taken only when the emitting node's output
  /// carries a matching route value. Empty means the edge always fires.
  /// @gotags: json:"route,omitempty" yaml:"route,omitempty"
  @$pb.TagNumber(3)
  $core.String get route => $_getSZ(2);
  @$pb.TagNumber(3)
  set route($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasRoute() => $_has(2);
  @$pb.TagNumber(3)
  void clearRoute() => $_clearField(3);

  /// Default edge: taken when no other route label on the same source node
  /// matches. Mutually exclusive with route.
  /// @gotags: json:"is_default,omitempty" yaml:"is_default,omitempty"
  @$pb.TagNumber(4)
  $core.bool get isDefault => $_getBF(3);
  @$pb.TagNumber(4)
  set isDefault($core.bool value) => $_setBool(3, value);
  @$pb.TagNumber(4)
  $core.bool hasIsDefault() => $_has(3);
  @$pb.TagNumber(4)
  void clearIsDefault() => $_clearField(4);
}

/// AgentRuntime captures runtime behavior flags that are safe to serialize.
class AgentRuntime extends $pb.GeneratedMessage {
  factory AgentRuntime({
    StreamingMode? streamingMode,
    $core.bool? saveInputBlobsAsArtifacts,
  }) {
    final result = AgentRuntime._();
    if (streamingMode != null) result.streamingMode = streamingMode;
    if (saveInputBlobsAsArtifacts != null)
      result.saveInputBlobsAsArtifacts = saveInputBlobsAsArtifacts;
    return result;
  }

  AgentRuntime._();

  factory AgentRuntime.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      AgentRuntime()..mergeFromBuffer(data, registry);
  factory AgentRuntime.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      AgentRuntime()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'AgentRuntime',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: AgentRuntime.$_createMessage)
    ..aE<StreamingMode>(1, _omitFieldNames ? '' : 'streamingMode',
        enumValues: StreamingMode.values)
    ..aOB(2, _omitFieldNames ? '' : 'saveInputBlobsAsArtifacts')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  AgentRuntime clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  AgentRuntime copyWith(void Function(AgentRuntime) updates) =>
      super.copyWith((message) => updates(message as AgentRuntime))
          as AgentRuntime;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use AgentRuntime() / AgentRuntime.new instead')
  static AgentRuntime create() => AgentRuntime._();
  static $pb.GeneratedMessage $_createMessage() => AgentRuntime._();
  @$core.override
  AgentRuntime createEmptyInstance() => AgentRuntime._();
  @$core.pragma('dart2js:noInline')
  static AgentRuntime getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<AgentRuntime>(
          AgentRuntime.$_createMessage);
  static AgentRuntime? _defaultInstance;

  /// @gotags: json:"streaming_mode,omitempty" yaml:"streaming_mode,omitempty"
  @$pb.TagNumber(1)
  StreamingMode get streamingMode => $_getN(0);
  @$pb.TagNumber(1)
  set streamingMode(StreamingMode value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasStreamingMode() => $_has(0);
  @$pb.TagNumber(1)
  void clearStreamingMode() => $_clearField(1);

  /// @gotags: json:"save_input_blobs_as_artifacts,omitempty" yaml:"save_input_blobs_as_artifacts,omitempty"
  @$pb.TagNumber(2)
  $core.bool get saveInputBlobsAsArtifacts => $_getBF(1);
  @$pb.TagNumber(2)
  set saveInputBlobsAsArtifacts($core.bool value) => $_setBool(1, value);
  @$pb.TagNumber(2)
  $core.bool hasSaveInputBlobsAsArtifacts() => $_has(1);
  @$pb.TagNumber(2)
  void clearSaveInputBlobsAsArtifacts() => $_clearField(2);
}

/// MCPServer describes a serializable MCP server connection config.
class MCPServer extends $pb.GeneratedMessage {
  factory MCPServer({
    $core.String? name,
    MCPServerTransport? transport,
    $core.String? url,
    $core.Iterable<$core.MapEntry<$core.String, $core.String>>? headers,
    $core.Iterable<$core.String>? toolFilter,
    $core.Iterable<$core.MapEntry<$core.String, $core.String>>? metadata,
    $core.String? id,
    $core.int? timeoutSeconds,
    MCPServerAuth? auth,
    $core.String? workspaceId,
  }) {
    final result = MCPServer._();
    if (name != null) result.name = name;
    if (transport != null) result.transport = transport;
    if (url != null) result.url = url;
    if (headers != null) result.headers.addEntries(headers);
    if (toolFilter != null) result.toolFilter.addAll(toolFilter);
    if (metadata != null) result.metadata.addEntries(metadata);
    if (id != null) result.id = id;
    if (timeoutSeconds != null) result.timeoutSeconds = timeoutSeconds;
    if (auth != null) result.auth = auth;
    if (workspaceId != null) result.workspaceId = workspaceId;
    return result;
  }

  MCPServer._();

  factory MCPServer.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      MCPServer()..mergeFromBuffer(data, registry);
  factory MCPServer.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      MCPServer()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'MCPServer',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: MCPServer.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'name')
    ..aE<MCPServerTransport>(2, _omitFieldNames ? '' : 'transport',
        enumValues: MCPServerTransport.values)
    ..aOS(6, _omitFieldNames ? '' : 'url')
    ..m<$core.String, $core.String>(7, _omitFieldNames ? '' : 'headers',
        entryClassName: 'MCPServer.HeadersEntry',
        keyFieldType: $pb.PbFieldType.OS,
        valueFieldType: $pb.PbFieldType.OS,
        packageName: const $pb.PackageName('agents.v1'))
    ..pPS(8, _omitFieldNames ? '' : 'toolFilter')
    ..m<$core.String, $core.String>(9, _omitFieldNames ? '' : 'metadata',
        entryClassName: 'MCPServer.MetadataEntry',
        keyFieldType: $pb.PbFieldType.OS,
        valueFieldType: $pb.PbFieldType.OS,
        packageName: const $pb.PackageName('agents.v1'))
    ..aOS(10, _omitFieldNames ? '' : 'id')
    ..aI(11, _omitFieldNames ? '' : 'timeoutSeconds')
    ..aOM<MCPServerAuth>(12, _omitFieldNames ? '' : 'auth',
        subBuilder: MCPServerAuth.$_createMessage)
    ..aOS(100, _omitFieldNames ? '' : 'workspaceId')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  MCPServer clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  MCPServer copyWith(void Function(MCPServer) updates) =>
      super.copyWith((message) => updates(message as MCPServer)) as MCPServer;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use MCPServer() / MCPServer.new instead')
  static MCPServer create() => MCPServer._();
  static $pb.GeneratedMessage $_createMessage() => MCPServer._();
  @$core.override
  MCPServer createEmptyInstance() => MCPServer._();
  @$core.pragma('dart2js:noInline')
  static MCPServer getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<MCPServer>(MCPServer.$_createMessage);
  static MCPServer? _defaultInstance;

  /// @gotags: json:"name,omitempty" yaml:"name,omitempty"
  @$pb.TagNumber(1)
  $core.String get name => $_getSZ(0);
  @$pb.TagNumber(1)
  set name($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasName() => $_has(0);
  @$pb.TagNumber(1)
  void clearName() => $_clearField(1);

  /// @gotags: json:"transport,omitempty" yaml:"transport,omitempty"
  @$pb.TagNumber(2)
  MCPServerTransport get transport => $_getN(1);
  @$pb.TagNumber(2)
  set transport(MCPServerTransport value) => $_setField(2, value);
  @$pb.TagNumber(2)
  $core.bool hasTransport() => $_has(1);
  @$pb.TagNumber(2)
  void clearTransport() => $_clearField(2);

  /// Used by remote HTTP/SSE transports.
  /// @gotags: json:"url,omitempty" yaml:"url,omitempty"
  @$pb.TagNumber(6)
  $core.String get url => $_getSZ(2);
  @$pb.TagNumber(6)
  set url($core.String value) => $_setString(2, value);
  @$pb.TagNumber(6)
  $core.bool hasUrl() => $_has(2);
  @$pb.TagNumber(6)
  void clearUrl() => $_clearField(6);

  /// @gotags: json:"headers,omitempty" yaml:"headers,omitempty"
  @$pb.TagNumber(7)
  $pb.PbMap<$core.String, $core.String> get headers => $_getMap(3);

  /// Optional allowlist of tools exposed from this server.
  /// @gotags: json:"tool_filter,omitempty" yaml:"tool_filter,omitempty"
  @$pb.TagNumber(8)
  $pb.PbList<$core.String> get toolFilter => $_getList(4);

  /// @gotags: json:"metadata,omitempty" yaml:"metadata,omitempty"
  @$pb.TagNumber(9)
  $pb.PbMap<$core.String, $core.String> get metadata => $_getMap(5);

  /// Unique identifier for referencing this server from agent configs.
  /// Required when used in the top-level mcp_server_configs list; optional when inline.
  /// @gotags: json:"id,omitempty" yaml:"id,omitempty"
  @$pb.TagNumber(10)
  $core.String get id => $_getSZ(6);
  @$pb.TagNumber(10)
  set id($core.String value) => $_setString(6, value);
  @$pb.TagNumber(10)
  $core.bool hasId() => $_has(6);
  @$pb.TagNumber(10)
  void clearId() => $_clearField(10);

  /// Optional connection timeout in seconds for MCP clients.
  /// @gotags: json:"timeout_seconds,omitempty" yaml:"timeout_seconds,omitempty"
  @$pb.TagNumber(11)
  $core.int get timeoutSeconds => $_getIZ(7);
  @$pb.TagNumber(11)
  set timeoutSeconds($core.int value) => $_setSignedInt32(7, value);
  @$pb.TagNumber(11)
  $core.bool hasTimeoutSeconds() => $_has(7);
  @$pb.TagNumber(11)
  void clearTimeoutSeconds() => $_clearField(11);

  /// Authentication settings for remote MCP transports. If omitted, existing
  /// configs preserve their current behavior: static headers are sent when
  /// present, otherwise no auth is applied.
  /// @gotags: json:"auth,omitempty" yaml:"auth,omitempty"
  @$pb.TagNumber(12)
  MCPServerAuth get auth => $_getN(8);
  @$pb.TagNumber(12)
  set auth(MCPServerAuth value) => $_setField(12, value);
  @$pb.TagNumber(12)
  $core.bool hasAuth() => $_has(8);
  @$pb.TagNumber(12)
  void clearAuth() => $_clearField(12);
  @$pb.TagNumber(12)
  MCPServerAuth ensureAuth() => $_ensure(8);

  /// ID of the workspace this server config belongs to.
  /// @gotags: json:"workspace_id,omitempty" yaml:"workspace_id,omitempty" bson:"workspace_id,omitempty"
  @$pb.TagNumber(100)
  $core.String get workspaceId => $_getSZ(9);
  @$pb.TagNumber(100)
  set workspaceId($core.String value) => $_setString(9, value);
  @$pb.TagNumber(100)
  $core.bool hasWorkspaceId() => $_has(9);
  @$pb.TagNumber(100)
  void clearWorkspaceId() => $_clearField(100);
}

class MCPServerAuth extends $pb.GeneratedMessage {
  factory MCPServerAuth({
    MCPServerAuthType? type,
    MCPServerOAuth2Config? oauth2,
  }) {
    final result = MCPServerAuth._();
    if (type != null) result.type = type;
    if (oauth2 != null) result.oauth2 = oauth2;
    return result;
  }

  MCPServerAuth._();

  factory MCPServerAuth.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      MCPServerAuth()..mergeFromBuffer(data, registry);
  factory MCPServerAuth.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      MCPServerAuth()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'MCPServerAuth',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: MCPServerAuth.$_createMessage)
    ..aE<MCPServerAuthType>(1, _omitFieldNames ? '' : 'type',
        enumValues: MCPServerAuthType.values)
    ..aOM<MCPServerOAuth2Config>(2, _omitFieldNames ? '' : 'oauth2',
        subBuilder: MCPServerOAuth2Config.$_createMessage)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  MCPServerAuth clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  MCPServerAuth copyWith(void Function(MCPServerAuth) updates) =>
      super.copyWith((message) => updates(message as MCPServerAuth))
          as MCPServerAuth;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use MCPServerAuth() / MCPServerAuth.new instead')
  static MCPServerAuth create() => MCPServerAuth._();
  static $pb.GeneratedMessage $_createMessage() => MCPServerAuth._();
  @$core.override
  MCPServerAuth createEmptyInstance() => MCPServerAuth._();
  @$core.pragma('dart2js:noInline')
  static MCPServerAuth getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<MCPServerAuth>(
          MCPServerAuth.$_createMessage);
  static MCPServerAuth? _defaultInstance;

  /// @gotags: json:"type,omitempty" yaml:"type,omitempty"
  @$pb.TagNumber(1)
  MCPServerAuthType get type => $_getN(0);
  @$pb.TagNumber(1)
  set type(MCPServerAuthType value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasType() => $_has(0);
  @$pb.TagNumber(1)
  void clearType() => $_clearField(1);

  /// @gotags: json:"oauth2,omitempty" yaml:"oauth2,omitempty"
  @$pb.TagNumber(2)
  MCPServerOAuth2Config get oauth2 => $_getN(1);
  @$pb.TagNumber(2)
  set oauth2(MCPServerOAuth2Config value) => $_setField(2, value);
  @$pb.TagNumber(2)
  $core.bool hasOauth2() => $_has(1);
  @$pb.TagNumber(2)
  void clearOauth2() => $_clearField(2);
  @$pb.TagNumber(2)
  MCPServerOAuth2Config ensureOauth2() => $_ensure(1);
}

class MCPServerOAuth2Config extends $pb.GeneratedMessage {
  factory MCPServerOAuth2Config({
    $core.String? clientId,
    $core.String? clientSecret,
    $core.Iterable<$core.String>? scopes,
    $core.String? authorizationUrl,
    $core.String? tokenUrl,
    $core.String? resourceMetadataUrl,
    $core.String? authorizationServerUrl,
    $core.String? resource,
  }) {
    final result = MCPServerOAuth2Config._();
    if (clientId != null) result.clientId = clientId;
    if (clientSecret != null) result.clientSecret = clientSecret;
    if (scopes != null) result.scopes.addAll(scopes);
    if (authorizationUrl != null) result.authorizationUrl = authorizationUrl;
    if (tokenUrl != null) result.tokenUrl = tokenUrl;
    if (resourceMetadataUrl != null)
      result.resourceMetadataUrl = resourceMetadataUrl;
    if (authorizationServerUrl != null)
      result.authorizationServerUrl = authorizationServerUrl;
    if (resource != null) result.resource = resource;
    return result;
  }

  MCPServerOAuth2Config._();

  factory MCPServerOAuth2Config.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      MCPServerOAuth2Config()..mergeFromBuffer(data, registry);
  factory MCPServerOAuth2Config.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      MCPServerOAuth2Config()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'MCPServerOAuth2Config',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: MCPServerOAuth2Config.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'clientId')
    ..aOS(2, _omitFieldNames ? '' : 'clientSecret')
    ..pPS(3, _omitFieldNames ? '' : 'scopes')
    ..aOS(4, _omitFieldNames ? '' : 'authorizationUrl')
    ..aOS(5, _omitFieldNames ? '' : 'tokenUrl')
    ..aOS(6, _omitFieldNames ? '' : 'resourceMetadataUrl')
    ..aOS(7, _omitFieldNames ? '' : 'authorizationServerUrl')
    ..aOS(8, _omitFieldNames ? '' : 'resource')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  MCPServerOAuth2Config clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  MCPServerOAuth2Config copyWith(
          void Function(MCPServerOAuth2Config) updates) =>
      super.copyWith((message) => updates(message as MCPServerOAuth2Config))
          as MCPServerOAuth2Config;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use MCPServerOAuth2Config() / MCPServerOAuth2Config.new instead')
  static MCPServerOAuth2Config create() => MCPServerOAuth2Config._();
  static $pb.GeneratedMessage $_createMessage() => MCPServerOAuth2Config._();
  @$core.override
  MCPServerOAuth2Config createEmptyInstance() => MCPServerOAuth2Config._();
  @$core.pragma('dart2js:noInline')
  static MCPServerOAuth2Config getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<MCPServerOAuth2Config>(
          MCPServerOAuth2Config.$_createMessage);
  static MCPServerOAuth2Config? _defaultInstance;

  /// Optional pre-registered OAuth client ID. If omitted, the server may use
  /// dynamic registration when supported by the authorization server.
  /// @gotags: json:"client_id,omitempty" yaml:"client_id,omitempty"
  @$pb.TagNumber(1)
  $core.String get clientId => $_getSZ(0);
  @$pb.TagNumber(1)
  set clientId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasClientId() => $_has(0);
  @$pb.TagNumber(1)
  void clearClientId() => $_clearField(1);

  /// Optional pre-registered OAuth client secret. Returned by the API only as
  /// part of the MCP server config today; token material is stored separately.
  /// @gotags: json:"client_secret,omitempty" yaml:"client_secret,omitempty"
  @$pb.TagNumber(2)
  $core.String get clientSecret => $_getSZ(1);
  @$pb.TagNumber(2)
  set clientSecret($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasClientSecret() => $_has(1);
  @$pb.TagNumber(2)
  void clearClientSecret() => $_clearField(2);

  /// Requested OAuth scopes.
  /// @gotags: json:"scopes,omitempty" yaml:"scopes,omitempty"
  @$pb.TagNumber(3)
  $pb.PbList<$core.String> get scopes => $_getList(2);

  /// Optional authorization endpoint override. Discovery is used when empty.
  /// @gotags: json:"authorization_url,omitempty" yaml:"authorization_url,omitempty"
  @$pb.TagNumber(4)
  $core.String get authorizationUrl => $_getSZ(3);
  @$pb.TagNumber(4)
  set authorizationUrl($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasAuthorizationUrl() => $_has(3);
  @$pb.TagNumber(4)
  void clearAuthorizationUrl() => $_clearField(4);

  /// Optional token endpoint override. Discovery is used when empty.
  /// @gotags: json:"token_url,omitempty" yaml:"token_url,omitempty"
  @$pb.TagNumber(5)
  $core.String get tokenUrl => $_getSZ(4);
  @$pb.TagNumber(5)
  set tokenUrl($core.String value) => $_setString(4, value);
  @$pb.TagNumber(5)
  $core.bool hasTokenUrl() => $_has(4);
  @$pb.TagNumber(5)
  void clearTokenUrl() => $_clearField(5);

  /// Optional protected resource metadata URL override.
  /// @gotags: json:"resource_metadata_url,omitempty" yaml:"resource_metadata_url,omitempty"
  @$pb.TagNumber(6)
  $core.String get resourceMetadataUrl => $_getSZ(5);
  @$pb.TagNumber(6)
  set resourceMetadataUrl($core.String value) => $_setString(5, value);
  @$pb.TagNumber(6)
  $core.bool hasResourceMetadataUrl() => $_has(5);
  @$pb.TagNumber(6)
  void clearResourceMetadataUrl() => $_clearField(6);

  /// Optional authorization server issuer/base URL override.
  /// @gotags: json:"authorization_server_url,omitempty" yaml:"authorization_server_url,omitempty"
  @$pb.TagNumber(7)
  $core.String get authorizationServerUrl => $_getSZ(6);
  @$pb.TagNumber(7)
  set authorizationServerUrl($core.String value) => $_setString(6, value);
  @$pb.TagNumber(7)
  $core.bool hasAuthorizationServerUrl() => $_has(6);
  @$pb.TagNumber(7)
  void clearAuthorizationServerUrl() => $_clearField(7);

  /// Optional resource/audience value. Defaults to the MCP server URL.
  /// @gotags: json:"resource,omitempty" yaml:"resource,omitempty"
  @$pb.TagNumber(8)
  $core.String get resource => $_getSZ(7);
  @$pb.TagNumber(8)
  set resource($core.String value) => $_setString(7, value);
  @$pb.TagNumber(8)
  $core.bool hasResource() => $_has(7);
  @$pb.TagNumber(8)
  void clearResource() => $_clearField(8);
}

/// ContextGuardConfig configures automatic context window management for an agent.
class ContextGuardConfig extends $pb.GeneratedMessage {
  factory ContextGuardConfig({
    ContextGuardStrategy? strategy,
    $core.int? maxTurns,
    $core.int? maxTokens,
  }) {
    final result = ContextGuardConfig._();
    if (strategy != null) result.strategy = strategy;
    if (maxTurns != null) result.maxTurns = maxTurns;
    if (maxTokens != null) result.maxTokens = maxTokens;
    return result;
  }

  ContextGuardConfig._();

  factory ContextGuardConfig.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ContextGuardConfig()..mergeFromBuffer(data, registry);
  factory ContextGuardConfig.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ContextGuardConfig()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ContextGuardConfig',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: ContextGuardConfig.$_createMessage)
    ..aE<ContextGuardStrategy>(1, _omitFieldNames ? '' : 'strategy',
        enumValues: ContextGuardStrategy.values)
    ..aI(2, _omitFieldNames ? '' : 'maxTurns')
    ..aI(3, _omitFieldNames ? '' : 'maxTokens')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ContextGuardConfig clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ContextGuardConfig copyWith(void Function(ContextGuardConfig) updates) =>
      super.copyWith((message) => updates(message as ContextGuardConfig))
          as ContextGuardConfig;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use ContextGuardConfig() / ContextGuardConfig.new instead')
  static ContextGuardConfig create() => ContextGuardConfig._();
  static $pb.GeneratedMessage $_createMessage() => ContextGuardConfig._();
  @$core.override
  ContextGuardConfig createEmptyInstance() => ContextGuardConfig._();
  @$core.pragma('dart2js:noInline')
  static ContextGuardConfig getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ContextGuardConfig>(
          ContextGuardConfig.$_createMessage);
  static ContextGuardConfig? _defaultInstance;

  /// Strategy to use for context compaction.
  /// @gotags: json:"strategy,omitempty" yaml:"strategy,omitempty"
  @$pb.TagNumber(1)
  ContextGuardStrategy get strategy => $_getN(0);
  @$pb.TagNumber(1)
  set strategy(ContextGuardStrategy value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasStrategy() => $_has(0);
  @$pb.TagNumber(1)
  void clearStrategy() => $_clearField(1);

  /// Maximum number of content entries before compaction (sliding window only).
  /// Zero/unset keeps the existing default of 20.
  /// @gotags: json:"max_turns,omitempty" yaml:"max_turns,omitempty"
  @$pb.TagNumber(2)
  $core.int get maxTurns => $_getIZ(1);
  @$pb.TagNumber(2)
  set maxTurns($core.int value) => $_setSignedInt32(1, value);
  @$pb.TagNumber(2)
  $core.bool hasMaxTurns() => $_has(1);
  @$pb.TagNumber(2)
  void clearMaxTurns() => $_clearField(2);

  /// Agent Context Override for the input context window in tokens (threshold
  /// strategy only). This is not maximum output tokens. Zero/unset inherits
  /// the selected model's configured or embedded context metadata.
  /// @gotags: json:"max_tokens,omitempty" yaml:"max_tokens,omitempty"
  @$pb.TagNumber(3)
  $core.int get maxTokens => $_getIZ(2);
  @$pb.TagNumber(3)
  set maxTokens($core.int value) => $_setSignedInt32(2, value);
  @$pb.TagNumber(3)
  $core.bool hasMaxTokens() => $_has(2);
  @$pb.TagNumber(3)
  void clearMaxTokens() => $_clearField(3);
}

/// MemoryConfig controls an agent's use of Workspace Memory and Agent Memory
/// (ADR-0013). Memory needs an enabled WorkspaceMemoryConfig; without one the
/// agent runs without memory. Booleans are phrased so that the zero value is
/// the default: enabling memory turns on automatic Memory Recall and Memory
/// Capture, while the memory tools stay off until requested.
class MemoryConfig extends $pb.GeneratedMessage {
  factory MemoryConfig({
    $core.bool? enabled,
    $core.bool? disableAutoRecall,
    $core.bool? disableAutoCapture,
    $core.bool? enableTools,
    $core.bool? allowAgentScopeWrite,
    $core.int? topK,
    $core.double? threshold,
  }) {
    final result = MemoryConfig._();
    if (enabled != null) result.enabled = enabled;
    if (disableAutoRecall != null) result.disableAutoRecall = disableAutoRecall;
    if (disableAutoCapture != null)
      result.disableAutoCapture = disableAutoCapture;
    if (enableTools != null) result.enableTools = enableTools;
    if (allowAgentScopeWrite != null)
      result.allowAgentScopeWrite = allowAgentScopeWrite;
    if (topK != null) result.topK = topK;
    if (threshold != null) result.threshold = threshold;
    return result;
  }

  MemoryConfig._();

  factory MemoryConfig.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      MemoryConfig()..mergeFromBuffer(data, registry);
  factory MemoryConfig.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      MemoryConfig()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'MemoryConfig',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: MemoryConfig.$_createMessage)
    ..aOB(1, _omitFieldNames ? '' : 'enabled')
    ..aOB(2, _omitFieldNames ? '' : 'disableAutoRecall')
    ..aOB(3, _omitFieldNames ? '' : 'disableAutoCapture')
    ..aOB(4, _omitFieldNames ? '' : 'enableTools')
    ..aOB(5, _omitFieldNames ? '' : 'allowAgentScopeWrite')
    ..aI(6, _omitFieldNames ? '' : 'topK')
    ..aD(7, _omitFieldNames ? '' : 'threshold', fieldType: $pb.PbFieldType.OF)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  MemoryConfig clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  MemoryConfig copyWith(void Function(MemoryConfig) updates) =>
      super.copyWith((message) => updates(message as MemoryConfig))
          as MemoryConfig;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use MemoryConfig() / MemoryConfig.new instead')
  static MemoryConfig create() => MemoryConfig._();
  static $pb.GeneratedMessage $_createMessage() => MemoryConfig._();
  @$core.override
  MemoryConfig createEmptyInstance() => MemoryConfig._();
  @$core.pragma('dart2js:noInline')
  static MemoryConfig getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<MemoryConfig>(
          MemoryConfig.$_createMessage);
  static MemoryConfig? _defaultInstance;

  /// Master switch. When false, the rest of the config has no effect.
  /// @gotags: json:"enabled,omitempty" yaml:"enabled,omitempty"
  @$pb.TagNumber(1)
  $core.bool get enabled => $_getBF(0);
  @$pb.TagNumber(1)
  set enabled($core.bool value) => $_setBool(0, value);
  @$pb.TagNumber(1)
  $core.bool hasEnabled() => $_has(0);
  @$pb.TagNumber(1)
  void clearEnabled() => $_clearField(1);

  /// Skip the automatic per-turn Memory Recall.
  /// @gotags: json:"disable_auto_recall,omitempty" yaml:"disable_auto_recall,omitempty"
  @$pb.TagNumber(2)
  $core.bool get disableAutoRecall => $_getBF(1);
  @$pb.TagNumber(2)
  set disableAutoRecall($core.bool value) => $_setBool(1, value);
  @$pb.TagNumber(2)
  $core.bool hasDisableAutoRecall() => $_has(1);
  @$pb.TagNumber(2)
  void clearDisableAutoRecall() => $_clearField(2);

  /// Skip the automatic per-turn Memory Capture into Workspace Memory.
  /// @gotags: json:"disable_auto_capture,omitempty" yaml:"disable_auto_capture,omitempty"
  @$pb.TagNumber(3)
  $core.bool get disableAutoCapture => $_getBF(2);
  @$pb.TagNumber(3)
  set disableAutoCapture($core.bool value) => $_setBool(2, value);
  @$pb.TagNumber(3)
  $core.bool hasDisableAutoCapture() => $_has(2);
  @$pb.TagNumber(3)
  void clearDisableAutoCapture() => $_clearField(3);

  /// Mount the search_memory / add_memory tools (LLM root agents only).
  /// @gotags: json:"enable_tools,omitempty" yaml:"enable_tools,omitempty"
  @$pb.TagNumber(4)
  $core.bool get enableTools => $_getBF(3);
  @$pb.TagNumber(4)
  set enableTools($core.bool value) => $_setBool(3, value);
  @$pb.TagNumber(4)
  $core.bool hasEnableTools() => $_has(3);
  @$pb.TagNumber(4)
  void clearEnableTools() => $_clearField(4);

  /// Allow add_memory to write Agent Memory (scope = agent). Automatic
  /// capture never writes Agent Memory.
  /// @gotags: json:"allow_agent_scope_write,omitempty" yaml:"allow_agent_scope_write,omitempty"
  @$pb.TagNumber(5)
  $core.bool get allowAgentScopeWrite => $_getBF(4);
  @$pb.TagNumber(5)
  set allowAgentScopeWrite($core.bool value) => $_setBool(4, value);
  @$pb.TagNumber(5)
  $core.bool hasAllowAgentScopeWrite() => $_has(4);
  @$pb.TagNumber(5)
  void clearAllowAgentScopeWrite() => $_clearField(5);

  /// Maximum memories injected per turn across both scopes. Unset means 5.
  /// @gotags: json:"top_k,omitempty" yaml:"top_k,omitempty"
  @$pb.TagNumber(6)
  $core.int get topK => $_getIZ(5);
  @$pb.TagNumber(6)
  set topK($core.int value) => $_setSignedInt32(5, value);
  @$pb.TagNumber(6)
  $core.bool hasTopK() => $_has(5);
  @$pb.TagNumber(6)
  void clearTopK() => $_clearField(6);

  /// Minimum mem0 relevance score, in [0, 1]. Unset means 0.3.
  /// @gotags: json:"threshold,omitempty" yaml:"threshold,omitempty"
  @$pb.TagNumber(7)
  $core.double get threshold => $_getN(6);
  @$pb.TagNumber(7)
  set threshold($core.double value) => $_setFloat(6, value);
  @$pb.TagNumber(7)
  $core.bool hasThreshold() => $_has(6);
  @$pb.TagNumber(7)
  void clearThreshold() => $_clearField(7);
}

/// ResultCardConfig is an Agent's Card Policy: whether the models at and below
/// it in a run may render Result Cards with render_ui (ADR-0014, the Card
/// Policy amendment). Human Input Forms are outside it.
class ResultCardConfig extends $pb.GeneratedMessage {
  factory ResultCardConfig({
    ResultCardGeneration? generation,
    ResultCardPresentation? presentation,
  }) {
    final result = ResultCardConfig._();
    if (generation != null) result.generation = generation;
    if (presentation != null) result.presentation = presentation;
    return result;
  }

  ResultCardConfig._();

  factory ResultCardConfig.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ResultCardConfig()..mergeFromBuffer(data, registry);
  factory ResultCardConfig.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ResultCardConfig()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ResultCardConfig',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: ResultCardConfig.$_createMessage)
    ..aE<ResultCardGeneration>(1, _omitFieldNames ? '' : 'generation',
        enumValues: ResultCardGeneration.values)
    ..aE<ResultCardPresentation>(2, _omitFieldNames ? '' : 'presentation',
        enumValues: ResultCardPresentation.values)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ResultCardConfig clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ResultCardConfig copyWith(void Function(ResultCardConfig) updates) =>
      super.copyWith((message) => updates(message as ResultCardConfig))
          as ResultCardConfig;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use ResultCardConfig() / ResultCardConfig.new instead')
  static ResultCardConfig create() => ResultCardConfig._();
  static $pb.GeneratedMessage $_createMessage() => ResultCardConfig._();
  @$core.override
  ResultCardConfig createEmptyInstance() => ResultCardConfig._();
  @$core.pragma('dart2js:noInline')
  static ResultCardConfig getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<ResultCardConfig>(
          ResultCardConfig.$_createMessage);
  static ResultCardConfig? _defaultInstance;

  /// Unset inherits the parent's generation; a run's root allows cards.
  /// @gotags: json:"generation,omitempty" yaml:"generation,omitempty"
  @$pb.TagNumber(1)
  ResultCardGeneration get generation => $_getN(0);
  @$pb.TagNumber(1)
  set generation(ResultCardGeneration value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasGeneration() => $_has(0);
  @$pb.TagNumber(1)
  void clearGeneration() => $_clearField(1);

  /// Unset inherits the parent's presentation; a run's root uses AUTO.
  /// @gotags: json:"presentation,omitempty" yaml:"presentation,omitempty"
  @$pb.TagNumber(2)
  ResultCardPresentation get presentation => $_getN(1);
  @$pb.TagNumber(2)
  set presentation(ResultCardPresentation value) => $_setField(2, value);
  @$pb.TagNumber(2)
  $core.bool hasPresentation() => $_has(1);
  @$pb.TagNumber(2)
  void clearPresentation() => $_clearField(2);
}

/// RemoteAgent describes a remote agent reachable over a network protocol.
class RemoteAgent extends $pb.GeneratedMessage {
  factory RemoteAgent({
    $core.String? id,
    $core.String? name,
    $core.String? url,
    RemoteAgentProtocol? protocol,
    $core.String? daemonRuntimeId,
    $core.String? acpRuntime,
    $core.String? opencodeAgent,
    $core.String? opencodeModel,
    $core.String? username,
    $core.String? password,
    $core.String? workspaceId,
  }) {
    final result = RemoteAgent._();
    if (id != null) result.id = id;
    if (name != null) result.name = name;
    if (url != null) result.url = url;
    if (protocol != null) result.protocol = protocol;
    if (daemonRuntimeId != null) result.daemonRuntimeId = daemonRuntimeId;
    if (acpRuntime != null) result.acpRuntime = acpRuntime;
    if (opencodeAgent != null) result.opencodeAgent = opencodeAgent;
    if (opencodeModel != null) result.opencodeModel = opencodeModel;
    if (username != null) result.username = username;
    if (password != null) result.password = password;
    if (workspaceId != null) result.workspaceId = workspaceId;
    return result;
  }

  RemoteAgent._();

  factory RemoteAgent.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      RemoteAgent()..mergeFromBuffer(data, registry);
  factory RemoteAgent.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      RemoteAgent()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'RemoteAgent',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: RemoteAgent.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..aOS(2, _omitFieldNames ? '' : 'name')
    ..aOS(3, _omitFieldNames ? '' : 'url')
    ..aE<RemoteAgentProtocol>(4, _omitFieldNames ? '' : 'protocol',
        enumValues: RemoteAgentProtocol.values)
    ..aOS(6, _omitFieldNames ? '' : 'daemonRuntimeId')
    ..aOS(7, _omitFieldNames ? '' : 'acpRuntime')
    ..aOS(8, _omitFieldNames ? '' : 'opencodeAgent')
    ..aOS(9, _omitFieldNames ? '' : 'opencodeModel')
    ..aOS(10, _omitFieldNames ? '' : 'username')
    ..aOS(11, _omitFieldNames ? '' : 'password')
    ..aOS(100, _omitFieldNames ? '' : 'workspaceId')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  RemoteAgent clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  RemoteAgent copyWith(void Function(RemoteAgent) updates) =>
      super.copyWith((message) => updates(message as RemoteAgent))
          as RemoteAgent;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use RemoteAgent() / RemoteAgent.new instead')
  static RemoteAgent create() => RemoteAgent._();
  static $pb.GeneratedMessage $_createMessage() => RemoteAgent._();
  @$core.override
  RemoteAgent createEmptyInstance() => RemoteAgent._();
  @$core.pragma('dart2js:noInline')
  static RemoteAgent getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<RemoteAgent>(
          RemoteAgent.$_createMessage);
  static RemoteAgent? _defaultInstance;

  /// Unique identifier for referencing this remote agent from agent configs.
  /// @gotags: json:"id,omitempty" yaml:"id,omitempty"
  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);

  /// Human-readable name for the remote agent.
  /// @gotags: json:"name,omitempty" yaml:"name,omitempty"
  @$pb.TagNumber(2)
  $core.String get name => $_getSZ(1);
  @$pb.TagNumber(2)
  set name($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasName() => $_has(1);
  @$pb.TagNumber(2)
  void clearName() => $_clearField(2);

  /// Endpoint URL for communicating with the remote agent.
  /// Required for A2A protocol; unused for DAEMON protocol.
  /// @gotags: json:"url,omitempty" yaml:"url,omitempty"
  @$pb.TagNumber(3)
  $core.String get url => $_getSZ(2);
  @$pb.TagNumber(3)
  set url($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasUrl() => $_has(2);
  @$pb.TagNumber(3)
  void clearUrl() => $_clearField(3);

  /// Communication protocol for this remote agent.
  /// @gotags: json:"protocol,omitempty" yaml:"protocol,omitempty"
  @$pb.TagNumber(4)
  RemoteAgentProtocol get protocol => $_getN(3);
  @$pb.TagNumber(4)
  set protocol(RemoteAgentProtocol value) => $_setField(4, value);
  @$pb.TagNumber(4)
  $core.bool hasProtocol() => $_has(3);
  @$pb.TagNumber(4)
  void clearProtocol() => $_clearField(4);

  /// Runtime identifier for daemon protocol routing.
  /// Required when protocol is DAEMON; references a workspace DaemonRuntime.
  /// @gotags: json:"daemon_runtime_id,omitempty" yaml:"daemon_runtime_id,omitempty"
  @$pb.TagNumber(6)
  $core.String get daemonRuntimeId => $_getSZ(4);
  @$pb.TagNumber(6)
  set daemonRuntimeId($core.String value) => $_setString(4, value);
  @$pb.TagNumber(6)
  $core.bool hasDaemonRuntimeId() => $_has(4);
  @$pb.TagNumber(6)
  void clearDaemonRuntimeId() => $_clearField(6);

  /// ACP runtime to launch through the selected DaemonRuntime.
  /// Required when protocol is DAEMON. v1 supports "opencode" and "codex".
  /// @gotags: json:"acp_runtime,omitempty" yaml:"acp_runtime,omitempty"
  @$pb.TagNumber(7)
  $core.String get acpRuntime => $_getSZ(5);
  @$pb.TagNumber(7)
  set acpRuntime($core.String value) => $_setString(5, value);
  @$pb.TagNumber(7)
  $core.bool hasAcpRuntime() => $_has(5);
  @$pb.TagNumber(7)
  void clearAcpRuntime() => $_clearField(7);

  /// Optional opencode agent name (`/agent` API). Used by OPENCODE_HTTP only;
  /// when empty the server's default agent handles the request.
  /// @gotags: json:"opencode_agent,omitempty" yaml:"opencode_agent,omitempty"
  @$pb.TagNumber(8)
  $core.String get opencodeAgent => $_getSZ(6);
  @$pb.TagNumber(8)
  set opencodeAgent($core.String value) => $_setString(6, value);
  @$pb.TagNumber(8)
  $core.bool hasOpencodeAgent() => $_has(6);
  @$pb.TagNumber(8)
  void clearOpencodeAgent() => $_clearField(8);

  /// Optional opencode model identifier (e.g. "anthropic/claude-3-5-sonnet").
  /// Used by OPENCODE_HTTP only; defaults to the server-side configured model.
  /// @gotags: json:"opencode_model,omitempty" yaml:"opencode_model,omitempty"
  @$pb.TagNumber(9)
  $core.String get opencodeModel => $_getSZ(7);
  @$pb.TagNumber(9)
  set opencodeModel($core.String value) => $_setString(7, value);
  @$pb.TagNumber(9)
  $core.bool hasOpencodeModel() => $_has(7);
  @$pb.TagNumber(9)
  void clearOpencodeModel() => $_clearField(9);

  /// Optional HTTP basic-auth username for OPENCODE_HTTP. Defaults to
  /// "opencode" on the opencode server when password is set.
  /// @gotags: json:"username,omitempty" yaml:"username,omitempty"
  @$pb.TagNumber(10)
  $core.String get username => $_getSZ(8);
  @$pb.TagNumber(10)
  set username($core.String value) => $_setString(8, value);
  @$pb.TagNumber(10)
  $core.bool hasUsername() => $_has(8);
  @$pb.TagNumber(10)
  void clearUsername() => $_clearField(10);

  /// Optional HTTP basic-auth password for OPENCODE_HTTP.
  /// @gotags: json:"password,omitempty" yaml:"password,omitempty"
  @$pb.TagNumber(11)
  $core.String get password => $_getSZ(9);
  @$pb.TagNumber(11)
  set password($core.String value) => $_setString(9, value);
  @$pb.TagNumber(11)
  $core.bool hasPassword() => $_has(9);
  @$pb.TagNumber(11)
  void clearPassword() => $_clearField(11);

  /// ID of the workspace this remote agent belongs to.
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

/// ModelConfig describes a single model within a provider.
class ModelConfig extends $pb.GeneratedMessage {
  factory ModelConfig({
    $core.String? name,
    $core.String? alias,
    $core.int? contextWindowTokens,
  }) {
    final result = ModelConfig._();
    if (name != null) result.name = name;
    if (alias != null) result.alias = alias;
    if (contextWindowTokens != null)
      result.contextWindowTokens = contextWindowTokens;
    return result;
  }

  ModelConfig._();

  factory ModelConfig.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ModelConfig()..mergeFromBuffer(data, registry);
  factory ModelConfig.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ModelConfig()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ModelConfig',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: ModelConfig.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'name')
    ..aOS(2, _omitFieldNames ? '' : 'alias')
    ..aI(3, _omitFieldNames ? '' : 'contextWindowTokens',
        fieldType: $pb.PbFieldType.OU3)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ModelConfig clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ModelConfig copyWith(void Function(ModelConfig) updates) =>
      super.copyWith((message) => updates(message as ModelConfig))
          as ModelConfig;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use ModelConfig() / ModelConfig.new instead')
  static ModelConfig create() => ModelConfig._();
  static $pb.GeneratedMessage $_createMessage() => ModelConfig._();
  @$core.override
  ModelConfig createEmptyInstance() => ModelConfig._();
  @$core.pragma('dart2js:noInline')
  static ModelConfig getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<ModelConfig>(
          ModelConfig.$_createMessage);
  static ModelConfig? _defaultInstance;

  /// The actual provider model identifier (e.g. "gemini-2.5-pro", "gpt-4o").
  /// @gotags: json:"name,omitempty" yaml:"name,omitempty"
  @$pb.TagNumber(1)
  $core.String get name => $_getSZ(0);
  @$pb.TagNumber(1)
  set name($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasName() => $_has(0);
  @$pb.TagNumber(1)
  void clearName() => $_clearField(1);

  /// Short alias for referencing this model (e.g. "pro", "flash").
  /// Agents and channels use alias to reference models.
  /// @gotags: json:"alias,omitempty" yaml:"alias,omitempty"
  @$pb.TagNumber(2)
  $core.String get alias => $_getSZ(1);
  @$pb.TagNumber(2)
  set alias($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasAlias() => $_has(1);
  @$pb.TagNumber(2)
  void clearAlias() => $_clearField(2);

  /// Optional provider model input context-window capacity in tokens.
  /// Zero means unset; ContextGuard delegates to embedded model metadata.
  /// This is not the model's maximum output-token configuration.
  /// @gotags: json:"context_window_tokens,omitempty" yaml:"context_window_tokens,omitempty"
  @$pb.TagNumber(3)
  $core.int get contextWindowTokens => $_getIZ(2);
  @$pb.TagNumber(3)
  set contextWindowTokens($core.int value) => $_setUnsignedInt32(2, value);
  @$pb.TagNumber(3)
  $core.bool hasContextWindowTokens() => $_has(2);
  @$pb.TagNumber(3)
  void clearContextWindowTokens() => $_clearField(3);
}

/// ModelProvider maps a set of model names to a provider backend.
class ModelProvider extends $pb.GeneratedMessage {
  factory ModelProvider({
    $core.String? name,
    $core.String? type,
    $core.String? apiKey,
    $core.String? baseUrl,
    $core.Iterable<ModelConfig>? models,
    $core.String? workspaceId,
  }) {
    final result = ModelProvider._();
    if (name != null) result.name = name;
    if (type != null) result.type = type;
    if (apiKey != null) result.apiKey = apiKey;
    if (baseUrl != null) result.baseUrl = baseUrl;
    if (models != null) result.models.addAll(models);
    if (workspaceId != null) result.workspaceId = workspaceId;
    return result;
  }

  ModelProvider._();

  factory ModelProvider.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ModelProvider()..mergeFromBuffer(data, registry);
  factory ModelProvider.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      ModelProvider()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ModelProvider',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: ModelProvider.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'name')
    ..aOS(2, _omitFieldNames ? '' : 'type')
    ..aOS(3, _omitFieldNames ? '' : 'apiKey')
    ..aOS(4, _omitFieldNames ? '' : 'baseUrl')
    ..pPM<ModelConfig>(5, _omitFieldNames ? '' : 'models',
        subBuilder: ModelConfig.$_createMessage)
    ..aOS(100, _omitFieldNames ? '' : 'workspaceId')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ModelProvider clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ModelProvider copyWith(void Function(ModelProvider) updates) =>
      super.copyWith((message) => updates(message as ModelProvider))
          as ModelProvider;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use ModelProvider() / ModelProvider.new instead')
  static ModelProvider create() => ModelProvider._();
  static $pb.GeneratedMessage $_createMessage() => ModelProvider._();
  @$core.override
  ModelProvider createEmptyInstance() => ModelProvider._();
  @$core.pragma('dart2js:noInline')
  static ModelProvider getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<ModelProvider>(
          ModelProvider.$_createMessage);
  static ModelProvider? _defaultInstance;

  /// Human-readable name for this provider entry (e.g. "openai", "ollama").
  /// @gotags: json:"name,omitempty" yaml:"name,omitempty"
  @$pb.TagNumber(1)
  $core.String get name => $_getSZ(0);
  @$pb.TagNumber(1)
  set name($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasName() => $_has(0);
  @$pb.TagNumber(1)
  void clearName() => $_clearField(1);

  /// Provider type: "gemini" or "openai".
  /// @gotags: json:"type,omitempty" yaml:"type,omitempty"
  @$pb.TagNumber(2)
  $core.String get type => $_getSZ(1);
  @$pb.TagNumber(2)
  set type($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasType() => $_has(1);
  @$pb.TagNumber(2)
  void clearType() => $_clearField(2);

  /// API key for authentication. Can be injected via env vars.
  /// @gotags: json:"api_key,omitempty" yaml:"api_key,omitempty"
  @$pb.TagNumber(3)
  $core.String get apiKey => $_getSZ(2);
  @$pb.TagNumber(3)
  set apiKey($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasApiKey() => $_has(2);
  @$pb.TagNumber(3)
  void clearApiKey() => $_clearField(3);

  /// Base URL for OpenAI-compatible endpoints (e.g. "http://localhost:11434/v1").
  /// @gotags: json:"base_url,omitempty" yaml:"base_url,omitempty"
  @$pb.TagNumber(4)
  $core.String get baseUrl => $_getSZ(3);
  @$pb.TagNumber(4)
  set baseUrl($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasBaseUrl() => $_has(3);
  @$pb.TagNumber(4)
  void clearBaseUrl() => $_clearField(4);

  /// Models this provider serves.
  /// @gotags: json:"models,omitempty" yaml:"models,omitempty"
  @$pb.TagNumber(5)
  $pb.PbList<ModelConfig> get models => $_getList(4);

  /// ID of the workspace this provider belongs to.
  /// @gotags: json:"workspace_id,omitempty" yaml:"workspace_id,omitempty" bson:"workspace_id,omitempty"
  @$pb.TagNumber(100)
  $core.String get workspaceId => $_getSZ(5);
  @$pb.TagNumber(100)
  set workspaceId($core.String value) => $_setString(5, value);
  @$pb.TagNumber(100)
  $core.bool hasWorkspaceId() => $_has(5);
  @$pb.TagNumber(100)
  void clearWorkspaceId() => $_clearField(100);
}

/// NotifyGroup groups outbound notification targets for cron delivery.
class NotifyGroup extends $pb.GeneratedMessage {
  factory NotifyGroup({
    $core.String? name,
    $core.bool? enabled,
    $core.Iterable<NotifyTarget>? targets,
    $core.Iterable<$core.MapEntry<$core.String, $core.String>>? metadata,
    $core.String? workspaceId,
  }) {
    final result = NotifyGroup._();
    if (name != null) result.name = name;
    if (enabled != null) result.enabled = enabled;
    if (targets != null) result.targets.addAll(targets);
    if (metadata != null) result.metadata.addEntries(metadata);
    if (workspaceId != null) result.workspaceId = workspaceId;
    return result;
  }

  NotifyGroup._();

  factory NotifyGroup.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      NotifyGroup()..mergeFromBuffer(data, registry);
  factory NotifyGroup.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      NotifyGroup()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'NotifyGroup',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: NotifyGroup.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'name')
    ..aOB(2, _omitFieldNames ? '' : 'enabled')
    ..pPM<NotifyTarget>(3, _omitFieldNames ? '' : 'targets',
        subBuilder: NotifyTarget.$_createMessage)
    ..m<$core.String, $core.String>(10, _omitFieldNames ? '' : 'metadata',
        entryClassName: 'NotifyGroup.MetadataEntry',
        keyFieldType: $pb.PbFieldType.OS,
        valueFieldType: $pb.PbFieldType.OS,
        packageName: const $pb.PackageName('agents.v1'))
    ..aOS(100, _omitFieldNames ? '' : 'workspaceId')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  NotifyGroup clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  NotifyGroup copyWith(void Function(NotifyGroup) updates) =>
      super.copyWith((message) => updates(message as NotifyGroup))
          as NotifyGroup;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use NotifyGroup() / NotifyGroup.new instead')
  static NotifyGroup create() => NotifyGroup._();
  static $pb.GeneratedMessage $_createMessage() => NotifyGroup._();
  @$core.override
  NotifyGroup createEmptyInstance() => NotifyGroup._();
  @$core.pragma('dart2js:noInline')
  static NotifyGroup getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<NotifyGroup>(
          NotifyGroup.$_createMessage);
  static NotifyGroup? _defaultInstance;

  /// Unique name identifying this notify group.
  /// @gotags: json:"name,omitempty" yaml:"name,omitempty"
  @$pb.TagNumber(1)
  $core.String get name => $_getSZ(0);
  @$pb.TagNumber(1)
  set name($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasName() => $_has(0);
  @$pb.TagNumber(1)
  void clearName() => $_clearField(1);

  /// Whether this group is enabled. Disabled groups are skipped at delivery time.
  /// @gotags: json:"enabled,omitempty" yaml:"enabled,omitempty"
  @$pb.TagNumber(2)
  $core.bool get enabled => $_getBF(1);
  @$pb.TagNumber(2)
  set enabled($core.bool value) => $_setBool(1, value);
  @$pb.TagNumber(2)
  $core.bool hasEnabled() => $_has(1);
  @$pb.TagNumber(2)
  void clearEnabled() => $_clearField(2);

  /// Notification targets in this group.
  /// @gotags: json:"targets,omitempty" yaml:"targets,omitempty"
  @$pb.TagNumber(3)
  $pb.PbList<NotifyTarget> get targets => $_getList(2);

  /// @gotags: json:"metadata,omitempty" yaml:"metadata,omitempty"
  @$pb.TagNumber(10)
  $pb.PbMap<$core.String, $core.String> get metadata => $_getMap(3);

  /// ID of the workspace this notify group belongs to.
  /// @gotags: json:"workspace_id,omitempty" yaml:"workspace_id,omitempty" bson:"workspace_id,omitempty"
  @$pb.TagNumber(100)
  $core.String get workspaceId => $_getSZ(4);
  @$pb.TagNumber(100)
  set workspaceId($core.String value) => $_setString(4, value);
  @$pb.TagNumber(100)
  $core.bool hasWorkspaceId() => $_has(4);
  @$pb.TagNumber(100)
  void clearWorkspaceId() => $_clearField(100);
}

/// NotifyTarget is a single outbound destination in a notify group.
class NotifyTarget extends $pb.GeneratedMessage {
  factory NotifyTarget({
    $core.String? name,
    $core.bool? enabled,
    NotifyTargetType? type,
    TelegramNotifyTarget? telegram,
    LarkNotifyTarget? lark,
    DiscordNotifyTarget? discord,
    $core.Iterable<$core.MapEntry<$core.String, $core.String>>? metadata,
  }) {
    final result = NotifyTarget._();
    if (name != null) result.name = name;
    if (enabled != null) result.enabled = enabled;
    if (type != null) result.type = type;
    if (telegram != null) result.telegram = telegram;
    if (lark != null) result.lark = lark;
    if (discord != null) result.discord = discord;
    if (metadata != null) result.metadata.addEntries(metadata);
    return result;
  }

  NotifyTarget._();

  factory NotifyTarget.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      NotifyTarget()..mergeFromBuffer(data, registry);
  factory NotifyTarget.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      NotifyTarget()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'NotifyTarget',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: NotifyTarget.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'name')
    ..aOB(2, _omitFieldNames ? '' : 'enabled')
    ..aE<NotifyTargetType>(3, _omitFieldNames ? '' : 'type',
        enumValues: NotifyTargetType.values)
    ..aOM<TelegramNotifyTarget>(4, _omitFieldNames ? '' : 'telegram',
        subBuilder: TelegramNotifyTarget.$_createMessage)
    ..aOM<LarkNotifyTarget>(5, _omitFieldNames ? '' : 'lark',
        subBuilder: LarkNotifyTarget.$_createMessage)
    ..aOM<DiscordNotifyTarget>(6, _omitFieldNames ? '' : 'discord',
        subBuilder: DiscordNotifyTarget.$_createMessage)
    ..m<$core.String, $core.String>(10, _omitFieldNames ? '' : 'metadata',
        entryClassName: 'NotifyTarget.MetadataEntry',
        keyFieldType: $pb.PbFieldType.OS,
        valueFieldType: $pb.PbFieldType.OS,
        packageName: const $pb.PackageName('agents.v1'))
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  NotifyTarget clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  NotifyTarget copyWith(void Function(NotifyTarget) updates) =>
      super.copyWith((message) => updates(message as NotifyTarget))
          as NotifyTarget;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use NotifyTarget() / NotifyTarget.new instead')
  static NotifyTarget create() => NotifyTarget._();
  static $pb.GeneratedMessage $_createMessage() => NotifyTarget._();
  @$core.override
  NotifyTarget createEmptyInstance() => NotifyTarget._();
  @$core.pragma('dart2js:noInline')
  static NotifyTarget getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<NotifyTarget>(
          NotifyTarget.$_createMessage);
  static NotifyTarget? _defaultInstance;

  /// Human-readable target name for logging and UI display.
  /// @gotags: json:"name,omitempty" yaml:"name,omitempty"
  @$pb.TagNumber(1)
  $core.String get name => $_getSZ(0);
  @$pb.TagNumber(1)
  set name($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasName() => $_has(0);
  @$pb.TagNumber(1)
  void clearName() => $_clearField(1);

  /// Whether this target is enabled. Disabled targets are skipped.
  /// @gotags: json:"enabled,omitempty" yaml:"enabled,omitempty"
  @$pb.TagNumber(2)
  $core.bool get enabled => $_getBF(1);
  @$pb.TagNumber(2)
  set enabled($core.bool value) => $_setBool(1, value);
  @$pb.TagNumber(2)
  $core.bool hasEnabled() => $_has(1);
  @$pb.TagNumber(2)
  void clearEnabled() => $_clearField(2);

  /// Target platform.
  /// @gotags: json:"type,omitempty" yaml:"type,omitempty"
  @$pb.TagNumber(3)
  NotifyTargetType get type => $_getN(2);
  @$pb.TagNumber(3)
  set type(NotifyTargetType value) => $_setField(3, value);
  @$pb.TagNumber(3)
  $core.bool hasType() => $_has(2);
  @$pb.TagNumber(3)
  void clearType() => $_clearField(3);

  /// Telegram target config.
  /// @gotags: json:"telegram,omitempty" yaml:"telegram,omitempty"
  @$pb.TagNumber(4)
  TelegramNotifyTarget get telegram => $_getN(3);
  @$pb.TagNumber(4)
  set telegram(TelegramNotifyTarget value) => $_setField(4, value);
  @$pb.TagNumber(4)
  $core.bool hasTelegram() => $_has(3);
  @$pb.TagNumber(4)
  void clearTelegram() => $_clearField(4);
  @$pb.TagNumber(4)
  TelegramNotifyTarget ensureTelegram() => $_ensure(3);

  /// Lark/Feishu webhook target config.
  /// @gotags: json:"lark,omitempty" yaml:"lark,omitempty"
  @$pb.TagNumber(5)
  LarkNotifyTarget get lark => $_getN(4);
  @$pb.TagNumber(5)
  set lark(LarkNotifyTarget value) => $_setField(5, value);
  @$pb.TagNumber(5)
  $core.bool hasLark() => $_has(4);
  @$pb.TagNumber(5)
  void clearLark() => $_clearField(5);
  @$pb.TagNumber(5)
  LarkNotifyTarget ensureLark() => $_ensure(4);

  /// Discord webhook target config.
  /// @gotags: json:"discord,omitempty" yaml:"discord,omitempty"
  @$pb.TagNumber(6)
  DiscordNotifyTarget get discord => $_getN(5);
  @$pb.TagNumber(6)
  set discord(DiscordNotifyTarget value) => $_setField(6, value);
  @$pb.TagNumber(6)
  $core.bool hasDiscord() => $_has(5);
  @$pb.TagNumber(6)
  void clearDiscord() => $_clearField(6);
  @$pb.TagNumber(6)
  DiscordNotifyTarget ensureDiscord() => $_ensure(5);

  /// @gotags: json:"metadata,omitempty" yaml:"metadata,omitempty"
  @$pb.TagNumber(10)
  $pb.PbMap<$core.String, $core.String> get metadata => $_getMap(6);
}

/// TelegramNotifyTarget addresses one Telegram Destination (issue #264).
///
/// Fields 1–4 held a raw Bot Token, chat ID, parse mode, and thread ID. They
/// remain reserved-but-declared so historical records still decode, but new
/// create/update calls reject any value in them: a Bot Token duplicated into
/// every notify group is exactly the fragmentation `destination_id` removes.
class TelegramNotifyTarget extends $pb.GeneratedMessage {
  factory TelegramNotifyTarget({
    @$core.Deprecated('This field is deprecated.') $core.String? botToken,
    @$core.Deprecated('This field is deprecated.') $core.String? chatId,
    @$core.Deprecated('This field is deprecated.') $core.String? parseMode,
    @$core.Deprecated('This field is deprecated.')
    $fixnum.Int64? messageThreadId,
    $core.String? destinationId,
  }) {
    final result = TelegramNotifyTarget._();
    if (botToken != null) result.botToken = botToken;
    if (chatId != null) result.chatId = chatId;
    if (parseMode != null) result.parseMode = parseMode;
    if (messageThreadId != null) result.messageThreadId = messageThreadId;
    if (destinationId != null) result.destinationId = destinationId;
    return result;
  }

  TelegramNotifyTarget._();

  factory TelegramNotifyTarget.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      TelegramNotifyTarget()..mergeFromBuffer(data, registry);
  factory TelegramNotifyTarget.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      TelegramNotifyTarget()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'TelegramNotifyTarget',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: TelegramNotifyTarget.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'botToken')
    ..aOS(2, _omitFieldNames ? '' : 'chatId')
    ..aOS(3, _omitFieldNames ? '' : 'parseMode')
    ..aInt64(4, _omitFieldNames ? '' : 'messageThreadId')
    ..aOS(5, _omitFieldNames ? '' : 'destinationId')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  TelegramNotifyTarget clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  TelegramNotifyTarget copyWith(void Function(TelegramNotifyTarget) updates) =>
      super.copyWith((message) => updates(message as TelegramNotifyTarget))
          as TelegramNotifyTarget;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated(
      'Use TelegramNotifyTarget() / TelegramNotifyTarget.new instead')
  static TelegramNotifyTarget create() => TelegramNotifyTarget._();
  static $pb.GeneratedMessage $_createMessage() => TelegramNotifyTarget._();
  @$core.override
  TelegramNotifyTarget createEmptyInstance() => TelegramNotifyTarget._();
  @$core.pragma('dart2js:noInline')
  static TelegramNotifyTarget getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<TelegramNotifyTarget>(
          TelegramNotifyTarget.$_createMessage);
  static TelegramNotifyTarget? _defaultInstance;

  /// Deprecated: rejected on write. Use destination_id.
  /// @gotags: json:"bot_token,omitempty" yaml:"bot_token,omitempty"
  @$core.Deprecated('This field is deprecated.')
  @$pb.TagNumber(1)
  $core.String get botToken => $_getSZ(0);
  @$core.Deprecated('This field is deprecated.')
  @$pb.TagNumber(1)
  set botToken($core.String value) => $_setString(0, value);
  @$core.Deprecated('This field is deprecated.')
  @$pb.TagNumber(1)
  $core.bool hasBotToken() => $_has(0);
  @$core.Deprecated('This field is deprecated.')
  @$pb.TagNumber(1)
  void clearBotToken() => $_clearField(1);

  /// Deprecated: rejected on write. Use destination_id.
  /// @gotags: json:"chat_id,omitempty" yaml:"chat_id,omitempty"
  @$core.Deprecated('This field is deprecated.')
  @$pb.TagNumber(2)
  $core.String get chatId => $_getSZ(1);
  @$core.Deprecated('This field is deprecated.')
  @$pb.TagNumber(2)
  set chatId($core.String value) => $_setString(1, value);
  @$core.Deprecated('This field is deprecated.')
  @$pb.TagNumber(2)
  $core.bool hasChatId() => $_has(1);
  @$core.Deprecated('This field is deprecated.')
  @$pb.TagNumber(2)
  void clearChatId() => $_clearField(2);

  /// Deprecated: rejected on write. Markdown handling is now central.
  /// @gotags: json:"parse_mode,omitempty" yaml:"parse_mode,omitempty"
  @$core.Deprecated('This field is deprecated.')
  @$pb.TagNumber(3)
  $core.String get parseMode => $_getSZ(2);
  @$core.Deprecated('This field is deprecated.')
  @$pb.TagNumber(3)
  set parseMode($core.String value) => $_setString(2, value);
  @$core.Deprecated('This field is deprecated.')
  @$pb.TagNumber(3)
  $core.bool hasParseMode() => $_has(2);
  @$core.Deprecated('This field is deprecated.')
  @$pb.TagNumber(3)
  void clearParseMode() => $_clearField(3);

  /// Deprecated: rejected on write. The Destination owns the Forum Topic.
  /// @gotags: json:"message_thread_id,omitempty" yaml:"message_thread_id,omitempty"
  @$core.Deprecated('This field is deprecated.')
  @$pb.TagNumber(4)
  $fixnum.Int64 get messageThreadId => $_getI64(3);
  @$core.Deprecated('This field is deprecated.')
  @$pb.TagNumber(4)
  set messageThreadId($fixnum.Int64 value) => $_setInt64(3, value);
  @$core.Deprecated('This field is deprecated.')
  @$pb.TagNumber(4)
  $core.bool hasMessageThreadId() => $_has(3);
  @$core.Deprecated('This field is deprecated.')
  @$pb.TagNumber(4)
  void clearMessageThreadId() => $_clearField(4);

  /// The Telegram Destination this target delivers to. The Destination owns
  /// the chat, the optional Forum Topic, and the credential used to reach it.
  /// @gotags: json:"destination_id,omitempty" yaml:"destination_id,omitempty" bson:"destination_id,omitempty"
  @$pb.TagNumber(5)
  $core.String get destinationId => $_getSZ(4);
  @$pb.TagNumber(5)
  set destinationId($core.String value) => $_setString(4, value);
  @$pb.TagNumber(5)
  $core.bool hasDestinationId() => $_has(4);
  @$pb.TagNumber(5)
  void clearDestinationId() => $_clearField(5);
}

class LarkNotifyTarget extends $pb.GeneratedMessage {
  factory LarkNotifyTarget({
    $core.String? webhookUrl,
    $core.String? secret,
  }) {
    final result = LarkNotifyTarget._();
    if (webhookUrl != null) result.webhookUrl = webhookUrl;
    if (secret != null) result.secret = secret;
    return result;
  }

  LarkNotifyTarget._();

  factory LarkNotifyTarget.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      LarkNotifyTarget()..mergeFromBuffer(data, registry);
  factory LarkNotifyTarget.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      LarkNotifyTarget()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'LarkNotifyTarget',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: LarkNotifyTarget.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'webhookUrl')
    ..aOS(2, _omitFieldNames ? '' : 'secret')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  LarkNotifyTarget clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  LarkNotifyTarget copyWith(void Function(LarkNotifyTarget) updates) =>
      super.copyWith((message) => updates(message as LarkNotifyTarget))
          as LarkNotifyTarget;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core.Deprecated('Use LarkNotifyTarget() / LarkNotifyTarget.new instead')
  static LarkNotifyTarget create() => LarkNotifyTarget._();
  static $pb.GeneratedMessage $_createMessage() => LarkNotifyTarget._();
  @$core.override
  LarkNotifyTarget createEmptyInstance() => LarkNotifyTarget._();
  @$core.pragma('dart2js:noInline')
  static LarkNotifyTarget getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<LarkNotifyTarget>(
          LarkNotifyTarget.$_createMessage);
  static LarkNotifyTarget? _defaultInstance;

  /// @gotags: json:"webhook_url,omitempty" yaml:"webhook_url,omitempty"
  @$pb.TagNumber(1)
  $core.String get webhookUrl => $_getSZ(0);
  @$pb.TagNumber(1)
  set webhookUrl($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasWebhookUrl() => $_has(0);
  @$pb.TagNumber(1)
  void clearWebhookUrl() => $_clearField(1);

  /// Optional signing secret for Lark custom bot webhooks.
  /// @gotags: json:"secret,omitempty" yaml:"secret,omitempty"
  @$pb.TagNumber(2)
  $core.String get secret => $_getSZ(1);
  @$pb.TagNumber(2)
  set secret($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasSecret() => $_has(1);
  @$pb.TagNumber(2)
  void clearSecret() => $_clearField(2);
}

class DiscordNotifyTarget extends $pb.GeneratedMessage {
  factory DiscordNotifyTarget({
    $core.String? webhookUrl,
    $core.String? username,
    $core.String? avatarUrl,
    $core.String? threadId,
  }) {
    final result = DiscordNotifyTarget._();
    if (webhookUrl != null) result.webhookUrl = webhookUrl;
    if (username != null) result.username = username;
    if (avatarUrl != null) result.avatarUrl = avatarUrl;
    if (threadId != null) result.threadId = threadId;
    return result;
  }

  DiscordNotifyTarget._();

  factory DiscordNotifyTarget.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DiscordNotifyTarget()..mergeFromBuffer(data, registry);
  factory DiscordNotifyTarget.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      DiscordNotifyTarget()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'DiscordNotifyTarget',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'agents.v1'),
      createEmptyInstance: DiscordNotifyTarget.$_createMessage)
    ..aOS(1, _omitFieldNames ? '' : 'webhookUrl')
    ..aOS(2, _omitFieldNames ? '' : 'username')
    ..aOS(3, _omitFieldNames ? '' : 'avatarUrl')
    ..aOS(4, _omitFieldNames ? '' : 'threadId')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DiscordNotifyTarget clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DiscordNotifyTarget copyWith(void Function(DiscordNotifyTarget) updates) =>
      super.copyWith((message) => updates(message as DiscordNotifyTarget))
          as DiscordNotifyTarget;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  @$core
      .Deprecated('Use DiscordNotifyTarget() / DiscordNotifyTarget.new instead')
  static DiscordNotifyTarget create() => DiscordNotifyTarget._();
  static $pb.GeneratedMessage $_createMessage() => DiscordNotifyTarget._();
  @$core.override
  DiscordNotifyTarget createEmptyInstance() => DiscordNotifyTarget._();
  @$core.pragma('dart2js:noInline')
  static DiscordNotifyTarget getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<DiscordNotifyTarget>(
          DiscordNotifyTarget.$_createMessage);
  static DiscordNotifyTarget? _defaultInstance;

  /// @gotags: json:"webhook_url,omitempty" yaml:"webhook_url,omitempty"
  @$pb.TagNumber(1)
  $core.String get webhookUrl => $_getSZ(0);
  @$pb.TagNumber(1)
  set webhookUrl($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasWebhookUrl() => $_has(0);
  @$pb.TagNumber(1)
  void clearWebhookUrl() => $_clearField(1);

  /// @gotags: json:"username,omitempty" yaml:"username,omitempty"
  @$pb.TagNumber(2)
  $core.String get username => $_getSZ(1);
  @$pb.TagNumber(2)
  set username($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasUsername() => $_has(1);
  @$pb.TagNumber(2)
  void clearUsername() => $_clearField(2);

  /// @gotags: json:"avatar_url,omitempty" yaml:"avatar_url,omitempty"
  @$pb.TagNumber(3)
  $core.String get avatarUrl => $_getSZ(2);
  @$pb.TagNumber(3)
  set avatarUrl($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasAvatarUrl() => $_has(2);
  @$pb.TagNumber(3)
  void clearAvatarUrl() => $_clearField(3);

  /// @gotags: json:"thread_id,omitempty" yaml:"thread_id,omitempty"
  @$pb.TagNumber(4)
  $core.String get threadId => $_getSZ(3);
  @$pb.TagNumber(4)
  set threadId($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasThreadId() => $_has(3);
  @$pb.TagNumber(4)
  void clearThreadId() => $_clearField(4);
}

const $core.bool _omitFieldNames =
    $core.bool.fromEnvironment('protobuf.omit_field_names');
const $core.bool _omitMessageNames =
    $core.bool.fromEnvironment('protobuf.omit_message_names');
