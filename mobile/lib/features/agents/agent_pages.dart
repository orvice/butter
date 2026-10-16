import 'package:flutter/material.dart';

import '../../app/client_context.dart';
import '../../core/management_page.dart';
import '../../gen/agents/v1/agent.pb.dart';
import 'agent_content_pages.dart';
import 'agent_draft.dart';
import 'agent_form.dart';
import 'agent_lifecycle_pages.dart';
import 'leaf_draft.dart';
import 'workflow_graph_view.dart';
export 'agent_form.dart' show AgentFormPage;

Future<void> openAgentDetail(
  BuildContext context,
  ClientContext client,
  String id,
) async {
  final scope = client.scope;
  await client.loadAgent(id);
  if (context.mounted && client.scope == scope && client.user != null) {
    await openManagementPage(
      context,
      client,
      AgentDetailPage(client: client, id: id),
    );
  }
}

Future<void> openAgentEditor(
  BuildContext context,
  ClientContext client, {
  String? id,
}) async {
  final scope = client.scope;
  if (id != null) {
    await client.loadAgent(id);
    if (client.scope != scope || client.error != null) return;
  }
  await client.prepareAgentEditor();
  if (!context.mounted || client.scope != scope || client.user == null) return;
  await openManagementPage(
    context,
    client,
    AgentFormPage(
      client: client,
      base: id == null ? null : client.managedAgent,
    ),
  );
}

String agentStatusLabel(Agent agent) => unknownEnum(agent, 26) == null
    ? agent.lifecycleStatus.name.replaceFirst('AGENT_LIFECYCLE_STATUS_', '')
    : 'Unknown lifecycle (${unknownEnum(agent, 26)})';

class AgentDetailPage extends StatelessWidget {
  const AgentDetailPage({super.key, required this.client, required this.id});
  final ClientContext client;
  final String id;
  @override
  Widget build(BuildContext context) => ManagementPage(
    client: client,
    title: 'Agent details',
    child: AnimatedBuilder(
      animation: client,
      builder: (context, _) {
        final agent = client.managedAgent;
        if (agent == null || agent.agentId != id) {
          return Center(
            child: TextButton(
              onPressed: client.busy ? null : () => client.loadAgent(id),
              child: const Text('Load Agent'),
            ),
          );
        }
        final status = client.agentStatus;
        return ListView(
          padding: const EdgeInsets.all(16),
          children: [
            SelectableText(
              agent.displayName.isEmpty ? agent.name : agent.displayName,
              style: Theme.of(context).textTheme.headlineSmall,
            ),
            SelectableText('Agent ID: ${agent.agentId}'),
            Text('Type: ${agentTypeLabel(agent)}'),
            Text('Lifecycle: ${agentStatusLabel(agent)}'),
            Text(
              status == null
                  ? 'Runtime status unavailable'
                  : 'Runtime: ${status.state.name.replaceFirst('AGENT_RUNTIME_STATE_', '')} · ${status.inFlight} in flight',
            ),
            if (status?.hasLastRunAt() == true)
              Text('Last run: ${status!.lastRunAt.toDateTime().toLocal()}'),
            TextButton.icon(
              onPressed: client.busy ? null : () => client.loadAgent(id),
              icon: const Icon(Icons.refresh),
              label: const Text('Refresh Agent'),
            ),
            if (canEditAgent(agent))
              FilledButton.icon(
                onPressed: client.busy
                    ? null
                    : () => openAgentEditor(context, client, id: id),
                icon: const Icon(Icons.edit),
                label: const Text('Edit Agent'),
              )
            else
              const Text(
                'Editing this Agent type is not supported here. Use the Web dashboard; no configuration will be overwritten.',
              ),
            const SizedBox(height: 12),
            TextButton.icon(
              onPressed: client.busy
                  ? null
                  : () => openAgentOperations(context, client),
              icon: const Icon(Icons.history_toggle_off),
              label: const Text('Lifecycle operations'),
            ),
            if (client.agentContentGitOwned &&
                client.canManageWorkspaceResources)
              TextButton.icon(
                onPressed: client.busy
                    ? null
                    : () => openAgentContentEditor(context, client, id),
                icon: const Icon(Icons.description_outlined),
                label: const Text('Edit Git content'),
              ),
            if (agent.lifecycleStatus !=
                AgentLifecycleStatus.AGENT_LIFECYCLE_STATUS_DELETED)
              TextButton.icon(
                onPressed: client.busy || client.agentLifecycle == null
                    ? null
                    : () async {
                        final lifecycle = client.agentLifecycle!;
                        final label = agent.displayName.isEmpty
                            ? agent.name
                            : agent.displayName;
                        final ok = await confirmDeleteAgent(
                          context,
                          lifecycle,
                          id,
                          label,
                        );
                        if (ok && context.mounted) {
                          Navigator.pop(context);
                        }
                      },
                icon: const Icon(Icons.delete_outline),
                label: const Text('Tombstone Agent'),
              )
            else ...[
              if (client.canManageWorkspaceResources)
                TextButton.icon(
                  onPressed: client.busy || client.agentLifecycle == null
                      ? null
                      : () async {
                          final lifecycle = client.agentLifecycle!;
                          final label = agent.displayName.isEmpty
                              ? agent.name
                              : agent.displayName;
                          final ok = await confirmRestoreAgent(
                            context,
                            lifecycle,
                            id,
                            label,
                          );
                          if (ok && context.mounted) {
                            await client.loadAgent(id);
                          }
                        },
                  icon: const Icon(Icons.restore),
                  label: const Text('Restore Agent'),
                ),
              if (client.canManageWorkspaceResources)
                TextButton.icon(
                  onPressed: client.busy || client.agentLifecycle == null
                      ? null
                      : () async {
                          final lifecycle = client.agentLifecycle!;
                          final label = agent.displayName.isEmpty
                              ? agent.name
                              : agent.displayName;
                          await confirmPurgeContent(
                            context,
                            lifecycle,
                            id,
                            label,
                          );
                        },
                  icon: const Icon(Icons.delete_forever),
                  label: const Text('Purge Git content'),
                ),
            ],
            const SizedBox(height: 16),
            SelectableText(agent.description),
            if (isLLM(agent)) ...[
              Text('Model: ${agent.config.model}'),
              const Text('Instruction'),
              SelectableText(agent.config.instruction),
              const Text('Global instruction'),
              SelectableText(agent.config.globalInstruction),
              Text('Output key: ${agent.config.outputKey}'),
            ],
            if (agent.type == AgentType.AGENT_TYPE_WORKFLOW)
              ..._workflowDetail(context, agent.config.workflow),
            if (agent.type == AgentType.AGENT_TYPE_PI)
              ..._piDetail(agent.config.pi),
            if (agent.type == AgentType.AGENT_TYPE_CURSOR)
              ..._cursorDetail(agent.config.cursor),
            for (final id in agent.childAgentIds)
              ListTile(
                title: Text('Child: $id'),
                onTap: client.busy
                    ? null
                    : () => openAgentDetail(context, client, id),
              ),
            Text(
              'Interfaces: A2A ${agent.enableA2a} · OpenAI ${agent.enableOpenaiApi} · AG-UI ${agent.enableAgui}',
            ),
            for (final entry in agent.labels.entries)
              Text('Label ${entry.key}: ${entry.value}'),
            for (final entry in agent.metadata.entries)
              // Metadata is editable in the form, but never dumped in status/errors.
              Text('Metadata key: ${entry.key}'),
          ],
        );
      },
    ),
  );
}

/// The stored graph, drawn from the same model the editor changes, followed
/// by the same facts as text.
List<Widget> _workflowDetail(BuildContext context, WorkflowConfig workflow) => [
  const SizedBox(height: 16),
  Text('Workflow graph', style: Theme.of(context).textTheme.titleMedium),
  WorkflowGraphView(config: workflow),
  for (final node in workflow.nodes)
    ListTile(
      contentPadding: EdgeInsets.zero,
      leading: Icon(workflowKindIcon(node)),
      title: SelectableText(node.name.isEmpty ? '(unnamed)' : node.name),
      subtitle: Text(
        [
          workflowKindLabel(node),
          if (node.agentId.isNotEmpty) 'Agent ${node.agentId}',
          if (node.parallelWorker) 'parallel worker',
          if (node.hasForm() && node.form.fields.isNotEmpty)
            'form with ${node.form.fields.length} field${node.form.fields.length == 1 ? '' : 's'}',
          if (node.timeoutSeconds > 0) 'timeout ${node.timeoutSeconds}s',
          if (node.hasRetry() && node.retry.maxAttempts > 1)
            'up to ${node.retry.maxAttempts} attempts',
        ].join(' · '),
      ),
    ),
  for (final edge in workflow.edges) Text(describeWorkflowEdge(edge)),
];

String _runLimitText(RunLimitMode mode, int seconds) => switch (mode) {
  RunLimitMode.serverDefault => 'server default (1800 seconds)',
  RunLimitMode.unlimited => 'unlimited',
  RunLimitMode.seconds => '$seconds seconds',
};

String _orDefault(String value, String fallback) =>
    value.isEmpty ? fallback : value;

/// A Pi leaf's binding. Its instructions and tools live on the box.
List<Widget> _piDetail(PiAgentConfig pi) => [
  const SizedBox(height: 16),
  SelectableText('ButterBox ID: ${pi.butterboxId}'),
  SelectableText(
    'Working directory: ${_orDefault(pi.workingDir, "box default")}',
  ),
  Text('Provider: ${_orDefault(pi.provider, "pi default")}'),
  Text('Model: ${_orDefault(pi.model, "pi default")}'),
  Text('Thinking level: ${_orDefault(pi.thinkingLevel, "default")}'),
  Text('Maximum run time: ${_runLimitText(piRunLimit(pi), pi.maxRunSeconds)}'),
  const Text(
    'Instructions and tools are configured on the box (AGENTS.md and .pi).',
  ),
];

/// A Cursor leaf's binding. Its rules and tools live on the box.
List<Widget> _cursorDetail(CursorAgentConfig cursor) => [
  const SizedBox(height: 16),
  SelectableText('ButterBox ID: ${cursor.butterboxId}'),
  SelectableText(
    'Working directory: ${_orDefault(cursor.workingDir, "box default")}',
  ),
  Text('Model: ${_orDefault(cursor.model, "Cursor default")}'),
  Text('Mode: ${_orDefault(cursor.mode, "default (agent)")}'),
  Text(
    'Maximum run time: ${_runLimitText(cursorRunLimit(cursor), cursor.maxRunSeconds)}',
  ),
  const Text(
    'Rules and tools are configured on the box (.cursor/rules and mcp.json).',
  ),
];
