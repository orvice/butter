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

import 'package:protobuf/protobuf.dart' as $pb;

/// WorkflowNodeKind selects the behaviour of a workflow node.
class WorkflowNodeKind extends $pb.ProtobufEnum {
  static const WorkflowNodeKind WORKFLOW_NODE_KIND_UNSPECIFIED =
      WorkflowNodeKind._(
          0, _omitEnumNames ? '' : 'WORKFLOW_NODE_KIND_UNSPECIFIED');

  /// Runs one of the workflow agent's sub-agents, referenced by name.
  static const WorkflowNodeKind WORKFLOW_NODE_KIND_AGENT =
      WorkflowNodeKind._(1, _omitEnumNames ? '' : 'WORKFLOW_NODE_KIND_AGENT');

  /// Pauses the workflow (an Interrupt), asks a human the configured
  /// question, and resumes the graph when the reply arrives.
  static const WorkflowNodeKind WORKFLOW_NODE_KIND_HUMAN_INPUT =
      WorkflowNodeKind._(
          2, _omitEnumNames ? '' : 'WORKFLOW_NODE_KIND_HUMAN_INPUT');

  /// Matches its input text against the route labels of its outgoing edges
  /// (trimmed, case-insensitive exact match) and steers the branch taken.
  static const WorkflowNodeKind WORKFLOW_NODE_KIND_ROUTER =
      WorkflowNodeKind._(3, _omitEnumNames ? '' : 'WORKFLOW_NODE_KIND_ROUTER');

  /// Fan-in barrier: activates once after every predecessor completed and
  /// aggregates their outputs.
  static const WorkflowNodeKind WORKFLOW_NODE_KIND_JOIN =
      WorkflowNodeKind._(4, _omitEnumNames ? '' : 'WORKFLOW_NODE_KIND_JOIN');

  static const $core.List<WorkflowNodeKind> values = <WorkflowNodeKind>[
    WORKFLOW_NODE_KIND_UNSPECIFIED,
    WORKFLOW_NODE_KIND_AGENT,
    WORKFLOW_NODE_KIND_HUMAN_INPUT,
    WORKFLOW_NODE_KIND_ROUTER,
    WORKFLOW_NODE_KIND_JOIN,
  ];

  static final $core.List<WorkflowNodeKind?> _byValue =
      $pb.ProtobufEnum.$_initByValueList(values, 4);
  static WorkflowNodeKind? valueOf($core.int value) =>
      value < 0 || value >= _byValue.length ? null : _byValue[value];

  const WorkflowNodeKind._(super.value, super.name);
}

/// HumanInputFormFieldType selects how one form field is answered.
class HumanInputFormFieldType extends $pb.ProtobufEnum {
  static const HumanInputFormFieldType HUMAN_INPUT_FORM_FIELD_TYPE_UNSPECIFIED =
      HumanInputFormFieldType._(
          0, _omitEnumNames ? '' : 'HUMAN_INPUT_FORM_FIELD_TYPE_UNSPECIFIED');

  /// Free text, up to max_length characters.
  static const HumanInputFormFieldType HUMAN_INPUT_FORM_FIELD_TYPE_TEXT =
      HumanInputFormFieldType._(
          1, _omitEnumNames ? '' : 'HUMAN_INPUT_FORM_FIELD_TYPE_TEXT');

  /// Exactly one of the fixed options.
  static const HumanInputFormFieldType
      HUMAN_INPUT_FORM_FIELD_TYPE_SINGLE_CHOICE = HumanInputFormFieldType._(
          2, _omitEnumNames ? '' : 'HUMAN_INPUT_FORM_FIELD_TYPE_SINGLE_CHOICE');

  static const $core.List<HumanInputFormFieldType> values =
      <HumanInputFormFieldType>[
    HUMAN_INPUT_FORM_FIELD_TYPE_UNSPECIFIED,
    HUMAN_INPUT_FORM_FIELD_TYPE_TEXT,
    HUMAN_INPUT_FORM_FIELD_TYPE_SINGLE_CHOICE,
  ];

  static final $core.List<HumanInputFormFieldType?> _byValue =
      $pb.ProtobufEnum.$_initByValueList(values, 2);
  static HumanInputFormFieldType? valueOf($core.int value) =>
      value < 0 || value >= _byValue.length ? null : _byValue[value];

  const HumanInputFormFieldType._(super.value, super.name);
}

class MCPServerAuthType extends $pb.ProtobufEnum {
  static const MCPServerAuthType MCP_SERVER_AUTH_TYPE_UNSPECIFIED =
      MCPServerAuthType._(
          0, _omitEnumNames ? '' : 'MCP_SERVER_AUTH_TYPE_UNSPECIFIED');
  static const MCPServerAuthType MCP_SERVER_AUTH_TYPE_NONE =
      MCPServerAuthType._(1, _omitEnumNames ? '' : 'MCP_SERVER_AUTH_TYPE_NONE');
  static const MCPServerAuthType MCP_SERVER_AUTH_TYPE_STATIC_HEADERS =
      MCPServerAuthType._(
          2, _omitEnumNames ? '' : 'MCP_SERVER_AUTH_TYPE_STATIC_HEADERS');
  static const MCPServerAuthType MCP_SERVER_AUTH_TYPE_OAUTH2 =
      MCPServerAuthType._(
          3, _omitEnumNames ? '' : 'MCP_SERVER_AUTH_TYPE_OAUTH2');

  static const $core.List<MCPServerAuthType> values = <MCPServerAuthType>[
    MCP_SERVER_AUTH_TYPE_UNSPECIFIED,
    MCP_SERVER_AUTH_TYPE_NONE,
    MCP_SERVER_AUTH_TYPE_STATIC_HEADERS,
    MCP_SERVER_AUTH_TYPE_OAUTH2,
  ];

  static final $core.List<MCPServerAuthType?> _byValue =
      $pb.ProtobufEnum.$_initByValueList(values, 3);
  static MCPServerAuthType? valueOf($core.int value) =>
      value < 0 || value >= _byValue.length ? null : _byValue[value];

  const MCPServerAuthType._(super.value, super.name);
}

class StreamingMode extends $pb.ProtobufEnum {
  static const StreamingMode STREAMING_MODE_UNSPECIFIED =
      StreamingMode._(0, _omitEnumNames ? '' : 'STREAMING_MODE_UNSPECIFIED');
  static const StreamingMode STREAMING_MODE_NONE =
      StreamingMode._(1, _omitEnumNames ? '' : 'STREAMING_MODE_NONE');
  static const StreamingMode STREAMING_MODE_SSE =
      StreamingMode._(2, _omitEnumNames ? '' : 'STREAMING_MODE_SSE');

  static const $core.List<StreamingMode> values = <StreamingMode>[
    STREAMING_MODE_UNSPECIFIED,
    STREAMING_MODE_NONE,
    STREAMING_MODE_SSE,
  ];

  static final $core.List<StreamingMode?> _byValue =
      $pb.ProtobufEnum.$_initByValueList(values, 2);
  static StreamingMode? valueOf($core.int value) =>
      value < 0 || value >= _byValue.length ? null : _byValue[value];

  const StreamingMode._(super.value, super.name);
}

/// AgentLifecycleStatus tracks whether an agent is runnable or blocked.
class AgentLifecycleStatus extends $pb.ProtobufEnum {
  static const AgentLifecycleStatus AGENT_LIFECYCLE_STATUS_UNSPECIFIED =
      AgentLifecycleStatus._(
          0, _omitEnumNames ? '' : 'AGENT_LIFECYCLE_STATUS_UNSPECIFIED');

  /// The agent is fully configured and runnable.
  static const AgentLifecycleStatus AGENT_LIFECYCLE_STATUS_ACTIVE =
      AgentLifecycleStatus._(
          1, _omitEnumNames ? '' : 'AGENT_LIFECYCLE_STATUS_ACTIVE');

  /// The agent (or a transitive dependency) lacks a complete Agent ID and
  /// cannot be used until the V2 migration is applied.
  static const AgentLifecycleStatus AGENT_LIFECYCLE_STATUS_MIGRATION_REQUIRED =
      AgentLifecycleStatus._(
          2, _omitEnumNames ? '' : 'AGENT_LIFECYCLE_STATUS_MIGRATION_REQUIRED');

  /// The DB row exists but the agent's initial Agent Content has not yet been
  /// committed, validated, and published. A create Saga holds the agent here
  /// until it activates. Not runnable.
  static const AgentLifecycleStatus AGENT_LIFECYCLE_STATUS_PROVISIONING =
      AgentLifecycleStatus._(
          3, _omitEnumNames ? '' : 'AGENT_LIFECYCLE_STATUS_PROVISIONING');

  /// A lifecycle Saga left the agent in a partial state (e.g. content committed
  /// but publication failed). Recoverable via RetryAgentOperation. Not runnable.
  static const AgentLifecycleStatus AGENT_LIFECYCLE_STATUS_ERROR =
      AgentLifecycleStatus._(
          4, _omitEnumNames ? '' : 'AGENT_LIFECYCLE_STATUS_ERROR');

  /// A soft-delete (tombstone) is in progress. Not runnable.
  static const AgentLifecycleStatus AGENT_LIFECYCLE_STATUS_DELETING =
      AgentLifecycleStatus._(
          5, _omitEnumNames ? '' : 'AGENT_LIFECYCLE_STATUS_DELETING');

  /// The agent is soft-deleted (tombstone). Its Agent ID stays reserved and its
  /// Git Agent Content is retained; RestoreAgent reactivates the same entity.
  /// Not runnable.
  static const AgentLifecycleStatus AGENT_LIFECYCLE_STATUS_DELETED =
      AgentLifecycleStatus._(
          6, _omitEnumNames ? '' : 'AGENT_LIFECYCLE_STATUS_DELETED');

  static const $core.List<AgentLifecycleStatus> values = <AgentLifecycleStatus>[
    AGENT_LIFECYCLE_STATUS_UNSPECIFIED,
    AGENT_LIFECYCLE_STATUS_ACTIVE,
    AGENT_LIFECYCLE_STATUS_MIGRATION_REQUIRED,
    AGENT_LIFECYCLE_STATUS_PROVISIONING,
    AGENT_LIFECYCLE_STATUS_ERROR,
    AGENT_LIFECYCLE_STATUS_DELETING,
    AGENT_LIFECYCLE_STATUS_DELETED,
  ];

  static final $core.List<AgentLifecycleStatus?> _byValue =
      $pb.ProtobufEnum.$_initByValueList(values, 6);
  static AgentLifecycleStatus? valueOf($core.int value) =>
      value < 0 || value >= _byValue.length ? null : _byValue[value];

  const AgentLifecycleStatus._(super.value, super.name);
}

/// AgentType selects which concrete ADK agent type to instantiate.
class AgentType extends $pb.ProtobufEnum {
  static const AgentType AGENT_TYPE_UNSPECIFIED =
      AgentType._(0, _omitEnumNames ? '' : 'AGENT_TYPE_UNSPECIFIED');
  static const AgentType AGENT_TYPE_LLM =
      AgentType._(1, _omitEnumNames ? '' : 'AGENT_TYPE_LLM');
  static const AgentType AGENT_TYPE_LOOP =
      AgentType._(2, _omitEnumNames ? '' : 'AGENT_TYPE_LOOP');
  static const AgentType AGENT_TYPE_SEQUENTIAL =
      AgentType._(3, _omitEnumNames ? '' : 'AGENT_TYPE_SEQUENTIAL');
  static const AgentType AGENT_TYPE_PARALLEL =
      AgentType._(4, _omitEnumNames ? '' : 'AGENT_TYPE_PARALLEL');
  static const AgentType AGENT_TYPE_WORKFLOW =
      AgentType._(5, _omitEnumNames ? '' : 'AGENT_TYPE_WORKFLOW');

  /// Leaf agent backed by a pi coding agent running on a ButterBox
  /// (ADR-0011). Configured via AgentConfig.pi; may not have children.
  static const AgentType AGENT_TYPE_PI =
      AgentType._(6, _omitEnumNames ? '' : 'AGENT_TYPE_PI');

  /// Leaf agent backed by a Cursor SDK Bridge session running on a ButterBox.
  /// Configured via AgentConfig.cursor; may not have children.
  static const AgentType AGENT_TYPE_CURSOR =
      AgentType._(7, _omitEnumNames ? '' : 'AGENT_TYPE_CURSOR');

  static const $core.List<AgentType> values = <AgentType>[
    AGENT_TYPE_UNSPECIFIED,
    AGENT_TYPE_LLM,
    AGENT_TYPE_LOOP,
    AGENT_TYPE_SEQUENTIAL,
    AGENT_TYPE_PARALLEL,
    AGENT_TYPE_WORKFLOW,
    AGENT_TYPE_PI,
    AGENT_TYPE_CURSOR,
  ];

  static final $core.List<AgentType?> _byValue =
      $pb.ProtobufEnum.$_initByValueList(values, 7);
  static AgentType? valueOf($core.int value) =>
      value < 0 || value >= _byValue.length ? null : _byValue[value];

  const AgentType._(super.value, super.name);
}

/// LLMIncludeContents matches llmagent.IncludeContents values.
class LLMIncludeContents extends $pb.ProtobufEnum {
  static const LLMIncludeContents LLM_INCLUDE_CONTENTS_UNSPECIFIED =
      LLMIncludeContents._(
          0, _omitEnumNames ? '' : 'LLM_INCLUDE_CONTENTS_UNSPECIFIED');
  static const LLMIncludeContents LLM_INCLUDE_CONTENTS_DEFAULT =
      LLMIncludeContents._(
          1, _omitEnumNames ? '' : 'LLM_INCLUDE_CONTENTS_DEFAULT');
  static const LLMIncludeContents LLM_INCLUDE_CONTENTS_NONE =
      LLMIncludeContents._(
          2, _omitEnumNames ? '' : 'LLM_INCLUDE_CONTENTS_NONE');

  static const $core.List<LLMIncludeContents> values = <LLMIncludeContents>[
    LLM_INCLUDE_CONTENTS_UNSPECIFIED,
    LLM_INCLUDE_CONTENTS_DEFAULT,
    LLM_INCLUDE_CONTENTS_NONE,
  ];

  static final $core.List<LLMIncludeContents?> _byValue =
      $pb.ProtobufEnum.$_initByValueList(values, 2);
  static LLMIncludeContents? valueOf($core.int value) =>
      value < 0 || value >= _byValue.length ? null : _byValue[value];

  const LLMIncludeContents._(super.value, super.name);
}

class MCPServerTransport extends $pb.ProtobufEnum {
  static const MCPServerTransport MCP_SERVER_TRANSPORT_UNSPECIFIED =
      MCPServerTransport._(
          0, _omitEnumNames ? '' : 'MCP_SERVER_TRANSPORT_UNSPECIFIED');
  static const MCPServerTransport MCP_SERVER_TRANSPORT_STREAMABLE_HTTP =
      MCPServerTransport._(
          2, _omitEnumNames ? '' : 'MCP_SERVER_TRANSPORT_STREAMABLE_HTTP');
  static const MCPServerTransport MCP_SERVER_TRANSPORT_SSE =
      MCPServerTransport._(3, _omitEnumNames ? '' : 'MCP_SERVER_TRANSPORT_SSE');

  static const $core.List<MCPServerTransport> values = <MCPServerTransport>[
    MCP_SERVER_TRANSPORT_UNSPECIFIED,
    MCP_SERVER_TRANSPORT_STREAMABLE_HTTP,
    MCP_SERVER_TRANSPORT_SSE,
  ];

  static final $core.List<MCPServerTransport?> _byValue =
      $pb.ProtobufEnum.$_initByValueList(values, 3);
  static MCPServerTransport? valueOf($core.int value) =>
      value < 0 || value >= _byValue.length ? null : _byValue[value];

  const MCPServerTransport._(super.value, super.name);
}

/// RemoteAgentProtocol selects the communication protocol for a remote agent.
class RemoteAgentProtocol extends $pb.ProtobufEnum {
  static const RemoteAgentProtocol REMOTE_AGENT_PROTOCOL_UNSPECIFIED =
      RemoteAgentProtocol._(
          0, _omitEnumNames ? '' : 'REMOTE_AGENT_PROTOCOL_UNSPECIFIED');
  static const RemoteAgentProtocol REMOTE_AGENT_PROTOCOL_A2A =
      RemoteAgentProtocol._(
          1, _omitEnumNames ? '' : 'REMOTE_AGENT_PROTOCOL_A2A');
  static const RemoteAgentProtocol REMOTE_AGENT_PROTOCOL_DAEMON =
      RemoteAgentProtocol._(
          2, _omitEnumNames ? '' : 'REMOTE_AGENT_PROTOCOL_DAEMON');

  /// Direct HTTP connection to a running `opencode serve` instance.
  static const RemoteAgentProtocol REMOTE_AGENT_PROTOCOL_OPENCODE_HTTP =
      RemoteAgentProtocol._(
          3, _omitEnumNames ? '' : 'REMOTE_AGENT_PROTOCOL_OPENCODE_HTTP');

  static const $core.List<RemoteAgentProtocol> values = <RemoteAgentProtocol>[
    REMOTE_AGENT_PROTOCOL_UNSPECIFIED,
    REMOTE_AGENT_PROTOCOL_A2A,
    REMOTE_AGENT_PROTOCOL_DAEMON,
    REMOTE_AGENT_PROTOCOL_OPENCODE_HTTP,
  ];

  static final $core.List<RemoteAgentProtocol?> _byValue =
      $pb.ProtobufEnum.$_initByValueList(values, 3);
  static RemoteAgentProtocol? valueOf($core.int value) =>
      value < 0 || value >= _byValue.length ? null : _byValue[value];

  const RemoteAgentProtocol._(super.value, super.name);
}

/// ContextGuardStrategy selects the compaction strategy for context management.
class ContextGuardStrategy extends $pb.ProtobufEnum {
  static const ContextGuardStrategy CONTEXT_GUARD_STRATEGY_UNSPECIFIED =
      ContextGuardStrategy._(
          0, _omitEnumNames ? '' : 'CONTEXT_GUARD_STRATEGY_UNSPECIFIED');
  static const ContextGuardStrategy CONTEXT_GUARD_STRATEGY_THRESHOLD =
      ContextGuardStrategy._(
          1, _omitEnumNames ? '' : 'CONTEXT_GUARD_STRATEGY_THRESHOLD');
  static const ContextGuardStrategy CONTEXT_GUARD_STRATEGY_SLIDING_WINDOW =
      ContextGuardStrategy._(
          2, _omitEnumNames ? '' : 'CONTEXT_GUARD_STRATEGY_SLIDING_WINDOW');

  static const $core.List<ContextGuardStrategy> values = <ContextGuardStrategy>[
    CONTEXT_GUARD_STRATEGY_UNSPECIFIED,
    CONTEXT_GUARD_STRATEGY_THRESHOLD,
    CONTEXT_GUARD_STRATEGY_SLIDING_WINDOW,
  ];

  static final $core.List<ContextGuardStrategy?> _byValue =
      $pb.ProtobufEnum.$_initByValueList(values, 2);
  static ContextGuardStrategy? valueOf($core.int value) =>
      value < 0 || value >= _byValue.length ? null : _byValue[value];

  const ContextGuardStrategy._(super.value, super.name);
}

/// ResultCardGeneration only narrows: nothing below a DISABLED agent can turn
/// Result Cards back on, so there is no ENABLED value.
class ResultCardGeneration extends $pb.ProtobufEnum {
  /// Inherit the parent's generation.
  static const ResultCardGeneration RESULT_CARD_GENERATION_UNSPECIFIED =
      ResultCardGeneration._(
          0, _omitEnumNames ? '' : 'RESULT_CARD_GENERATION_UNSPECIFIED');

  /// No Result Cards for this agent and every agent below it in the run.
  static const ResultCardGeneration RESULT_CARD_GENERATION_DISABLED =
      ResultCardGeneration._(
          1, _omitEnumNames ? '' : 'RESULT_CARD_GENERATION_DISABLED');

  static const $core.List<ResultCardGeneration> values = <ResultCardGeneration>[
    RESULT_CARD_GENERATION_UNSPECIFIED,
    RESULT_CARD_GENERATION_DISABLED,
  ];

  static final $core.List<ResultCardGeneration?> _byValue =
      $pb.ProtobufEnum.$_initByValueList(values, 1);
  static ResultCardGeneration? valueOf($core.int value) =>
      value < 0 || value >= _byValue.length ? null : _byValue[value];

  const ResultCardGeneration._(super.value, super.name);
}

/// ResultCardPresentation is a hint, carried in render_ui's description, for
/// how readily models show Result Cards. The nearest explicit value on the
/// path from the run's root wins. It never forces a card.
class ResultCardPresentation extends $pb.ProtobufEnum {
  /// Inherit the parent's presentation.
  static const ResultCardPresentation RESULT_CARD_PRESENTATION_UNSPECIFIED =
      ResultCardPresentation._(
          0, _omitEnumNames ? '' : 'RESULT_CARD_PRESENTATION_UNSPECIFIED');

  /// The model decides when a card helps.
  static const ResultCardPresentation RESULT_CARD_PRESENTATION_AUTO =
      ResultCardPresentation._(
          1, _omitEnumNames ? '' : 'RESULT_CARD_PRESENTATION_AUTO');

  /// Ask the model to show structured results in a card as well as in text.
  static const ResultCardPresentation RESULT_CARD_PRESENTATION_PREFERRED =
      ResultCardPresentation._(
          2, _omitEnumNames ? '' : 'RESULT_CARD_PRESENTATION_PREFERRED');

  static const $core.List<ResultCardPresentation> values =
      <ResultCardPresentation>[
    RESULT_CARD_PRESENTATION_UNSPECIFIED,
    RESULT_CARD_PRESENTATION_AUTO,
    RESULT_CARD_PRESENTATION_PREFERRED,
  ];

  static final $core.List<ResultCardPresentation?> _byValue =
      $pb.ProtobufEnum.$_initByValueList(values, 2);
  static ResultCardPresentation? valueOf($core.int value) =>
      value < 0 || value >= _byValue.length ? null : _byValue[value];

  const ResultCardPresentation._(super.value, super.name);
}

class NotifyTargetType extends $pb.ProtobufEnum {
  static const NotifyTargetType NOTIFY_TARGET_TYPE_UNSPECIFIED =
      NotifyTargetType._(
          0, _omitEnumNames ? '' : 'NOTIFY_TARGET_TYPE_UNSPECIFIED');
  static const NotifyTargetType NOTIFY_TARGET_TYPE_TELEGRAM =
      NotifyTargetType._(
          1, _omitEnumNames ? '' : 'NOTIFY_TARGET_TYPE_TELEGRAM');
  static const NotifyTargetType NOTIFY_TARGET_TYPE_LARK_WEBHOOK =
      NotifyTargetType._(
          2, _omitEnumNames ? '' : 'NOTIFY_TARGET_TYPE_LARK_WEBHOOK');
  static const NotifyTargetType NOTIFY_TARGET_TYPE_DISCORD_WEBHOOK =
      NotifyTargetType._(
          3, _omitEnumNames ? '' : 'NOTIFY_TARGET_TYPE_DISCORD_WEBHOOK');

  static const $core.List<NotifyTargetType> values = <NotifyTargetType>[
    NOTIFY_TARGET_TYPE_UNSPECIFIED,
    NOTIFY_TARGET_TYPE_TELEGRAM,
    NOTIFY_TARGET_TYPE_LARK_WEBHOOK,
    NOTIFY_TARGET_TYPE_DISCORD_WEBHOOK,
  ];

  static final $core.List<NotifyTargetType?> _byValue =
      $pb.ProtobufEnum.$_initByValueList(values, 3);
  static NotifyTargetType? valueOf($core.int value) =>
      value < 0 || value >= _byValue.length ? null : _byValue[value];

  const NotifyTargetType._(super.value, super.name);
}

const $core.bool _omitEnumNames =
    $core.bool.fromEnvironment('protobuf.omit_enum_names');
