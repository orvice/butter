import '../../gen/agents/v1/agent.pb.dart';
import '../agents/agent_draft.dart';

const maxMemoryTopK = 50;

class AgentMemoryFormValues {
  AgentMemoryFormValues({
    this.enabled = false,
    this.autoRecall = true,
    this.autoCapture = true,
    this.tools = false,
    this.agentScopeWrite = false,
    this.topK = '',
    this.threshold = '',
  });

  bool enabled;
  bool autoRecall;
  bool autoCapture;
  bool tools;
  bool agentScopeWrite;
  String topK;
  String threshold;
}

AgentMemoryFormValues memoryFormFromConfig(MemoryConfig? config) {
  if (config == null || !config.enabled) {
    return AgentMemoryFormValues();
  }
  return AgentMemoryFormValues(
    enabled: true,
    autoRecall: !config.disableAutoRecall,
    autoCapture: !config.disableAutoCapture,
    tools: config.enableTools,
    agentScopeWrite: config.allowAgentScopeWrite,
    topK: config.hasTopK() ? '${config.topK}' : '',
    threshold: config.hasThreshold() ? '${config.threshold}' : '',
  );
}

bool supportsAgentMemory(Agent agent) => !isBoxLeaf(agent);

bool supportsMemoryTools(Agent agent) => isLLM(agent);

String? validateMemoryTopK(String value) {
  final trimmed = value.trim();
  if (trimmed.isEmpty) return null;
  final n = int.tryParse(trimmed);
  if (n == null || n < 1 || n > maxMemoryTopK) {
    return 'Use a whole number from 1 to $maxMemoryTopK.';
  }
  return null;
}

String? validateMemoryThreshold(String value) {
  final trimmed = value.trim();
  if (trimmed.isEmpty) return null;
  final n = double.tryParse(trimmed);
  if (n == null || n < 0 || n > 1) {
    return 'Use a number from 0 to 1.';
  }
  return null;
}

/// Writes [values] onto [draft]. Disabled memory clears the config message.
void applyMemoryForm(AgentDraft draft, AgentMemoryFormValues values) {
  if (!supportsAgentMemory(draft.agent)) {
    if (draft.agent.hasConfig()) draft.config.clearMemory();
    return;
  }
  if (!values.enabled) {
    draft.config.clearMemory();
    return;
  }
  final mc = draft.config.ensureMemory()
    ..enabled = true
    ..disableAutoRecall = !values.autoRecall
    ..disableAutoCapture = !values.autoCapture;
  if (supportsMemoryTools(draft.agent) && values.tools) {
    mc.enableTools = true;
    mc.allowAgentScopeWrite = values.agentScopeWrite;
  } else {
    mc.clearEnableTools();
    mc.clearAllowAgentScopeWrite();
  }
  final topK = values.topK.trim();
  if (topK.isEmpty) {
    mc.clearTopK();
  } else {
    mc.topK = int.parse(topK);
  }
  final threshold = values.threshold.trim();
  if (threshold.isEmpty) {
    mc.clearThreshold();
  } else {
    mc.threshold = double.parse(threshold);
  }
}
