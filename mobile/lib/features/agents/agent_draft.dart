import 'package:protobuf/protobuf.dart';

import '../../gen/agents/v1/agent.pb.dart';

enum GuardMode { off, threshold, slidingWindow, retained }

const editableAgentTypes = [
  AgentType.AGENT_TYPE_LLM,
  AgentType.AGENT_TYPE_LOOP,
  AgentType.AGENT_TYPE_SEQUENTIAL,
  AgentType.AGENT_TYPE_PARALLEL,
  AgentType.AGENT_TYPE_WORKFLOW,
  AgentType.AGENT_TYPE_PI,
  AgentType.AGENT_TYPE_CURSOR,
];

/// PI and CURSOR agents are leaves whose behavior lives on a ButterBox.
bool isBoxLeaf(Agent agent) =>
    agent.type == AgentType.AGENT_TYPE_PI ||
    agent.type == AgentType.AGENT_TYPE_CURSOR;

/// A Pi or Cursor leaf that holds children or a Butter-side behavior field.
/// The server refuses them on write (they cannot exist on a leaf it accepted),
/// so a record that has them is legacy or from a newer server. The editor
/// would either send them and be refused, or have to drop them silently, so
/// such a leaf is left read-only instead.
bool leafHoldsForbiddenFields(Agent agent) {
  if (!isBoxLeaf(agent)) return false;
  if (agent.childAgentIds.isNotEmpty) return true;
  if (!agent.hasConfig()) return false;
  final c = agent.config;
  return c.instruction.isNotEmpty ||
      c.globalInstruction.isNotEmpty ||
      c.mcpServers.isNotEmpty ||
      c.mcpServerIds.isNotEmpty ||
      c.skills.isNotEmpty ||
      c.fileMounts.isNotEmpty ||
      c.hasContextGuard() ||
      c.remoteAgentIds.isNotEmpty ||
      c.hasMemory() ||
      c.hasResultCards();
}

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
    // Legacy content must be detected to prevent the composite backend from
    // discarding it; it is never edited or resolved through this field.
    // ignore: deprecated_member_use_from_same_package
    agent.subAgents.isEmpty &&
    !leafHoldsForbiddenFields(agent) &&
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
    final leavingWorkflow = agent.type == AgentType.AGENT_TYPE_WORKFLOW;
    final leavingPi = agent.type == AgentType.AGENT_TYPE_PI;
    final leavingCursor = agent.type == AgentType.AGENT_TYPE_CURSOR;
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
    if (leavingWorkflow && agent.hasConfig()) {
      config.clearWorkflow();
    }
    // The children of a Workflow were derived from its nodes; without the
    // graph they would be orphans the next type neither shows nor owns.
    if (leavingWorkflow) agent.childAgentIds.clear();
    if (leavingPi && agent.hasConfig()) config.clearPi();
    if (leavingCursor && agent.hasConfig()) config.clearCursor();
    if (isBoxLeaf(agent)) {
      // A leaf's behavior lives on the box, and the server rejects every
      // Butter-side behavior field on write: drop what an earlier type left
      // in this draft rather than send something it would refuse.
      agent.childAgentIds.clear();
      config
        ..clearInstruction()
        ..clearGlobalInstruction()
        ..clearContextGuard()
        ..clearMemory()
        ..clearResultCards()
        ..clearMaxIterations()
        ..clearWorkflow();
      config.mcpServers.clear();
      config.mcpServerIds.clear();
      config.skills.clear();
      config.fileMounts.clear();
      config.remoteAgentIds.clear();
      if (type == AgentType.AGENT_TYPE_PI) {
        config.ensurePi();
      } else {
        config.ensureCursor();
      }
    }
  }

  void setGuard(GuardMode mode) {
    if (mode == GuardMode.retained) return;
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
