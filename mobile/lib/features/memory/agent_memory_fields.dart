import 'package:flutter/material.dart';

import '../../gen/agents/v1/agent.pb.dart';
import 'agent_memory_config.dart';

class AgentMemoryFields extends StatelessWidget {
  const AgentMemoryFields({
    super.key,
    required this.agent,
    required this.values,
    required this.workspaceMemoryReady,
    required this.enabled,
    required this.onChanged,
  });

  final Agent agent;
  final AgentMemoryFormValues values;
  final bool workspaceMemoryReady;
  final bool enabled;
  final VoidCallback onChanged;

  @override
  Widget build(BuildContext context) {
    if (!supportsAgentMemory(agent)) return const SizedBox.shrink();
    void set(void Function(AgentMemoryFormValues v) fn) {
      fn(values);
      onChanged();
    }

    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        Text('Memory', style: Theme.of(context).textTheme.titleMedium),
        const Text(
          'Recall and capture use the workspace mem0 connection. Only the root agent\'s settings apply when invoked directly.',
        ),
        if (!workspaceMemoryReady)
          const Text(
            'This workspace has no enabled mem0 connection yet. Configure one under Memory on the Agents page.',
          ),
        SwitchListTile(
          title: const Text('Enable memory'),
          value: values.enabled,
          onChanged: enabled ? (v) => set((s) => s.enabled = v) : null,
        ),
        if (values.enabled) ...[
          SwitchListTile(
            title: const Text('Auto recall'),
            subtitle: const Text(
              'Inject relevant memories before each model call (never saved to the session).',
            ),
            value: values.autoRecall,
            onChanged: enabled ? (v) => set((s) => s.autoRecall = v) : null,
          ),
          SwitchListTile(
            title: const Text('Auto capture'),
            subtitle: const Text(
              'Send successful turns to Workspace Memory (secrets redacted server-side).',
            ),
            value: values.autoCapture,
            onChanged: enabled ? (v) => set((s) => s.autoCapture = v) : null,
          ),
          if (supportsMemoryTools(agent)) ...[
            SwitchListTile(
              title: const Text('Memory tools'),
              subtitle: const Text(
                'Offer search_memory and add_memory to the model.',
              ),
              value: values.tools,
              onChanged: enabled ? (v) => set((s) => s.tools = v) : null,
            ),
            if (values.tools)
              SwitchListTile(
                title: const Text('Allow agent-scope writes'),
                subtitle: const Text(
                  'Lets add_memory store Agent Memory, not only the shared pool.',
                ),
                value: values.agentScopeWrite,
                onChanged: enabled
                    ? (v) => set((s) => s.agentScopeWrite = v)
                    : null,
              ),
          ] else
            const Text(
              'Memory tools are only available on LLM agents. LLM sub-agents still receive recalled memories.',
            ),
          TextFormField(
            key: const Key('memory-top-k'),
            initialValue: values.topK,
            decoration: const InputDecoration(
              labelText: 'Memories per turn',
              helperText:
                  'Across both scopes, 1–50. Empty uses the server default (5).',
            ),
            enabled: enabled,
            onChanged: (v) => set((s) => s.topK = v),
            autovalidateMode: AutovalidateMode.onUserInteraction,
            validator: (v) => validateMemoryTopK(v ?? ''),
          ),
          TextFormField(
            key: const Key('memory-threshold'),
            initialValue: values.threshold,
            decoration: const InputDecoration(
              labelText: 'Relevance threshold',
              helperText:
                  'Minimum mem0 score, 0–1. Empty uses the server default (0.3).',
            ),
            enabled: enabled,
            onChanged: (v) => set((s) => s.threshold = v),
            autovalidateMode: AutovalidateMode.onUserInteraction,
            validator: (v) => validateMemoryThreshold(v ?? ''),
          ),
        ],
      ],
    );
  }
}
