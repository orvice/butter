import 'package:protobuf/protobuf.dart';

import '../../gen/agents/v1/agent.pb.dart';

enum GuardMode { off, threshold, slidingWindow }

const editableAgentTypes = [
  AgentType.AGENT_TYPE_LLM,
  AgentType.AGENT_TYPE_LOOP,
  AgentType.AGENT_TYPE_SEQUENTIAL,
  AgentType.AGENT_TYPE_PARALLEL,
];

bool isLLM(Agent agent) =>
    agent.type == AgentType.AGENT_TYPE_LLM ||
    agent.type == AgentType.AGENT_TYPE_UNSPECIFIED;

/// Dart retains an unrecognized enum on the wire as an unknown field, while
/// its getter returns the default. Never let that default rewrite the enum.
int? unknownEnum(GeneratedMessage message, int tag) {
  final values = message.unknownFields.getField(tag)?.varints;
  return values == null || values.isEmpty ? null : values.last.toInt();
}

bool canEditAgent(Agent agent) =>
    unknownEnum(agent, 20) == null &&
    unknownEnum(agent, 26) == null &&
    (agent.lifecycleStatus ==
            AgentLifecycleStatus.AGENT_LIFECYCLE_STATUS_ACTIVE ||
        agent.lifecycleStatus ==
            AgentLifecycleStatus.AGENT_LIFECYCLE_STATUS_UNSPECIFIED) &&
    (isLLM(agent) || editableAgentTypes.contains(agent.type));

String agentTypeLabel(Agent agent) => unknownEnum(agent, 20) != null
    ? 'Unknown type (${unknownEnum(agent, 20)})'
    : agent.type.name.replaceFirst('AGENT_TYPE_', '');

/// An in-memory whole-message edit. Only user edits touch fields, so absent
/// messages, optional values and future protobuf fields survive a save.
final class AgentDraft {
  AgentDraft({Agent? base})
    : agent = base?.deepCopy() ?? Agent(type: AgentType.AGENT_TYPE_LLM);
  final Agent agent;
  Agent build() => agent.deepCopy();

  AgentConfig get config {
    if (!agent.hasConfig()) agent.config = AgentConfig();
    return agent.config;
  }

  void setType(AgentType type) {
    if (agent.type == type) return;
    agent.type = type;
    if (!isLLM(agent) && agent.hasConfig()) {
      config
        ..clearModel()
        ..clearInstruction()
        ..clearGlobalInstruction()
        ..clearDisallowTransferToParent()
        ..clearDisallowTransferToPeers()
        ..clearIncludeContents()
        ..clearOutputKey()
        ..clearInputSchemaJson()
        ..clearOutputSchemaJson()
        ..clearContextGuard();
    }
    if (type != AgentType.AGENT_TYPE_LOOP && agent.hasConfig()) {
      config.clearMaxIterations();
    }
  }

  void setGuard(GuardMode mode) {
    if (mode == GuardMode.off) {
      if (agent.hasConfig()) config.clearContextGuard();
      return;
    }
    if (!config.hasContextGuard()) config.contextGuard = ContextGuardConfig();
    final guard = config.contextGuard;
    if (mode == GuardMode.threshold) {
      guard
        ..strategy = ContextGuardStrategy.CONTEXT_GUARD_STRATEGY_THRESHOLD
        ..clearMaxTurns();
    } else {
      guard
        ..strategy = ContextGuardStrategy.CONTEXT_GUARD_STRATEGY_SLIDING_WINDOW
        ..clearMaxTokens();
    }
  }
}
