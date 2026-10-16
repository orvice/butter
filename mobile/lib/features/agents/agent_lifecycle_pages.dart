import 'dart:async';

import 'package:flutter/material.dart';

import '../../app/client_context.dart';
import '../../core/agent_lifecycle_errors.dart';
import '../../core/management_page.dart';
import '../../gen/agents/v1/agent_operation.pbenum.dart';
import 'agent_lifecycle_controller.dart';

Future<void> openAgentOperations(
  BuildContext context,
  ClientContext client,
) async {
  final lifecycle = client.agentLifecycle;
  if (lifecycle == null) return;
  unawaited(lifecycle.loadOperations());
  await openManagementPage(
    context,
    client,
    AgentOperationsPage(client: client),
  );
}

class AgentOperationsPage extends StatelessWidget {
  const AgentOperationsPage({super.key, required this.client});
  final ClientContext client;

  @override
  Widget build(BuildContext context) {
    final lifecycle = client.agentLifecycle;
    if (lifecycle == null) return const SizedBox.shrink();
    return ListenableBuilder(
      listenable: lifecycle,
      builder: (context, _) => Scaffold(
        appBar: AppBar(title: const Text('Agent operations')),
        body: ListView(
          padding: const EdgeInsets.all(16),
          children: [
            const Text(
              'Durable lifecycle Sagas for create, save, tombstone and restore. '
              'These are not Agent chat runs or Invocations.',
            ),
            if (lifecycle.unconfirmed) ...[
              const SizedBox(height: 8),
              Text(lifecycle.writeError ?? lifecycleUnconfirmed),
            ],
            const SizedBox(height: 12),
            Wrap(
              spacing: 8,
              children: [
                for (final filter in <AgentOperationStatus?>[
                  null,
                  AgentOperationStatus.AGENT_OPERATION_STATUS_FAILED,
                  AgentOperationStatus.AGENT_OPERATION_STATUS_RUNNING,
                ])
                  FilterChip(
                    label: Text(
                      filter == null ? 'All' : operationStatusLabel(filter),
                    ),
                    selected: lifecycle.operationsFilter == filter,
                    onSelected: lifecycle.writing
                        ? null
                        : (_) {
                            lifecycle.operationsFilter = filter;
                            unawaited(lifecycle.loadOperations());
                          },
                  ),
              ],
            ),
            if (lifecycle.operationsState == LifecycleLoad.loading)
              const LinearProgressIndicator(semanticsLabel: 'Loading'),
            if (lifecycle.operationsError != null)
              Text(lifecycle.operationsError!),
            for (final op in lifecycle.operations)
              ListTile(
                title: Text(
                  '${operationTypeLabel(op.type)} · ${op.agentId.isEmpty ? op.agentName : op.agentId}',
                ),
                subtitle: Text('${operationStatusLabel(op.status)} · ${op.id}'),
                onTap: lifecycle.writing
                    ? null
                    : () => openOperationDetail(context, client, op.id),
              ),
            if (lifecycle.operationsNextPage.isNotEmpty)
              FilledButton(
                onPressed: lifecycle.writing
                    ? null
                    : () => lifecycle.loadOperations(more: true),
                child: const Text('Load more'),
              ),
          ],
        ),
      ),
    );
  }
}

Future<void> openOperationDetail(
  BuildContext context,
  ClientContext client,
  String operationId,
) async {
  final lifecycle = client.agentLifecycle;
  if (lifecycle == null) return;
  unawaited(lifecycle.loadOperation(operationId));
  await openManagementPage(
    context,
    client,
    AgentOperationDetailPage(client: client, operationId: operationId),
  );
}

class AgentOperationDetailPage extends StatelessWidget {
  const AgentOperationDetailPage({
    super.key,
    required this.client,
    required this.operationId,
  });
  final ClientContext client;
  final String operationId;

  @override
  Widget build(BuildContext context) {
    final lifecycle = client.agentLifecycle;
    if (lifecycle == null) return const SizedBox.shrink();
    return ListenableBuilder(
      listenable: lifecycle,
      builder: (context, _) {
        final op = lifecycle.operationDetail;
        return Scaffold(
          appBar: AppBar(title: const Text('Operation detail')),
          body: ListView(
            padding: const EdgeInsets.all(16),
            children: [
              if (lifecycle.detailState == LifecycleLoad.loading)
                const LinearProgressIndicator(semanticsLabel: 'Loading'),
              if (lifecycle.detailError != null) Text(lifecycle.detailError!),
              if (op != null) ...[
                SelectableText('Operation ID: ${op.id}'),
                Text(
                  'Agent: ${op.agentId.isEmpty ? op.agentName : op.agentId}',
                ),
                Text('Type: ${operationTypeLabel(op.type)}'),
                Text('Status: ${operationStatusLabel(op.status)}'),
                if (op.error.isNotEmpty) Text('Last error: ${op.error}'),
                const SizedBox(height: 12),
                const Text('Steps'),
                for (final step in op.steps)
                  ListTile(
                    dense: true,
                    title: Text(operationStepLabel(step.kind)),
                    subtitle: Text(
                      [
                        operationStepStatusLabel(step.status),
                        if (step.error.isNotEmpty) step.error,
                      ].join(' · '),
                    ),
                  ),
                if (op.status ==
                        AgentOperationStatus.AGENT_OPERATION_STATUS_FAILED &&
                    client.canManageWorkspaceResources)
                  FilledButton.icon(
                    onPressed: lifecycle.writing
                        ? null
                        : () async {
                            final ok = await lifecycle.retryOperation(op.id);
                            if (context.mounted && ok) {
                              ScaffoldMessenger.of(context).showSnackBar(
                                const SnackBar(
                                  content: Text('Operation retry finished.'),
                                ),
                              );
                            }
                          },
                    icon: const Icon(Icons.refresh),
                    label: const Text('Retry lifecycle operation'),
                  ),
              ],
            ],
          ),
        );
      },
    );
  }
}

Future<bool> confirmDeleteAgent(
  BuildContext context,
  AgentLifecycleController lifecycle,
  String agentId,
  String label,
) async {
  final accepted = await showDialog<bool>(
    context: context,
    builder: (context) => AlertDialog(
      title: const Text('Tombstone Agent?'),
      content: Text(
        'Delete (tombstone) $label ($agentId)? The Agent ID stays reserved and '
        'Git content is kept until an owner purges it. This is not the same as '
        'purging Git files.',
      ),
      actions: [
        TextButton(
          onPressed: () => Navigator.pop(context, false),
          child: const Text('Cancel'),
        ),
        FilledButton(
          onPressed: () => Navigator.pop(context, true),
          child: const Text('Tombstone'),
        ),
      ],
    ),
  );
  if (accepted != true) return false;
  return lifecycle.deleteAgent(agentId, displayName: label);
}

Future<bool> confirmRestoreAgent(
  BuildContext context,
  AgentLifecycleController lifecycle,
  String agentId,
  String label,
) async {
  final accepted = await showDialog<bool>(
    context: context,
    builder: (context) => AlertDialog(
      title: const Text('Restore Agent?'),
      content: Text(
        'Restore $label ($agentId) from its tombstone? Requires owner or admin '
        'when Git content is bound.',
      ),
      actions: [
        TextButton(
          onPressed: () => Navigator.pop(context, false),
          child: const Text('Cancel'),
        ),
        FilledButton(
          onPressed: () => Navigator.pop(context, true),
          child: const Text('Restore'),
        ),
      ],
    ),
  );
  if (accepted != true) return false;
  return lifecycle.restoreAgent(agentId);
}

Future<bool> confirmPurgeContent(
  BuildContext context,
  AgentLifecycleController lifecycle,
  String agentId,
  String label,
) async {
  final accepted = await showDialog<bool>(
    context: context,
    builder: (context) => AlertDialog(
      title: const Text('Purge Git content?'),
      content: Text(
        'Permanently remove managed Git content files for $label ($agentId)? '
        'This creates a new Git commit and cannot be undone from the App. '
        'Distinct from tombstoning the Agent record.',
      ),
      actions: [
        TextButton(
          onPressed: () => Navigator.pop(context, false),
          child: const Text('Cancel'),
        ),
        FilledButton(
          onPressed: () => Navigator.pop(context, true),
          child: const Text('Purge content'),
        ),
      ],
    ),
  );
  if (accepted != true) return false;
  return lifecycle.purgeContent(agentId);
}
