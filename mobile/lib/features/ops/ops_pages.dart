import 'dart:async';

import 'package:flutter/material.dart';
import 'package:flutter/services.dart';

import '../../app/client_context.dart';
import '../../core/config_resource_controller.dart';
import '../../core/management_page.dart';
import '../cron/cron_pages.dart';
import '../daemons/daemon_pages.dart';
import '../mcp/mcp_pages.dart';
import '../remote_agents/remote_agent_pages.dart';
import 'api_token_logic.dart';
import 'integration_urls.dart';
import '../../gen/agents/v1/agent_service.pb.dart' as agentspb;
import 'invocation_logic.dart';
import 'workspace_ops_controller.dart';

Future<void> openOverview(BuildContext context, ClientContext client) async {
  final ops = client.ops;
  if (ops == null) return;
  await openManagementPage(
    context,
    client,
    _OverviewPage(client: client, ops: ops),
  );
}

Future<void> openInvocations(BuildContext context, ClientContext client) async {
  final ops = client.ops;
  if (ops == null) return;
  await openManagementPage(
    context,
    client,
    _InvocationsPage(client: client, ops: ops),
  );
}

Future<void> openApiTokens(BuildContext context, ClientContext client) async {
  final ops = client.ops;
  if (ops == null) return;
  await openManagementPage(
    context,
    client,
    _ApiTokensPage(client: client, ops: ops),
  );
}

Future<void> openIntegrationsHub(
  BuildContext context,
  ClientContext client,
) async =>
    openManagementPage(context, client, _IntegrationsHubPage(client: client));

class _OverviewPage extends StatefulWidget {
  const _OverviewPage({required this.client, required this.ops});
  final ClientContext client;
  final WorkspaceOpsController ops;
  @override
  State<_OverviewPage> createState() => _OverviewPageState();
}

class _OverviewPageState extends State<_OverviewPage> {
  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addPostFrameCallback((_) {
      unawaited(widget.ops.loadOverview());
    });
  }

  @override
  Widget build(BuildContext context) => Scaffold(
    appBar: AppBar(title: const Text('Overview')),
    body: ListenableBuilder(
      listenable: widget.ops,
      builder: (context, _) {
        final ops = widget.ops;
        if (ops.overviewState == ConfigLoad.loading) {
          return const Center(child: CircularProgressIndicator());
        }
        if (ops.overviewState == ConfigLoad.failed) {
          return _ErrorBody(
            message: ops.overviewError ?? 'Could not load overview.',
            onRetry: () => ops.loadOverview(),
          );
        }
        return RefreshIndicator(
          onRefresh: ops.loadOverview,
          child: ListView(
            padding: const EdgeInsets.all(16),
            children: [
              if (ops.overviewIsServerWide)
                const Text(
                  'Server-wide metrics (all Workspaces). These are not limited '
                  'to the Workspace selected in the header.',
                  style: TextStyle(fontWeight: FontWeight.w600),
                )
              else
                Text(
                  'Workspace: ${widget.client.workspace?.name ?? widget.ops.workspaceId}. '
                  'Counts and recent activity below are scoped to this Workspace only.',
                ),
              const SizedBox(height: 16),
              if (ops.overviewIsServerWide && ops.serverOverview != null) ...[
                _SectionTitle('Counts (all Workspaces)'),
                _CountsGrid(
                  labels: {
                    'Agents': ops.serverOverview!.counts.activeAgents,
                    'MCP': ops.serverOverview!.counts.mcpServers,
                    'Remote': ops.serverOverview!.counts.remoteAgents,
                    'Cron jobs': ops.serverOverview!.counts.cronJobs,
                    'Sessions': ops.serverOverview!.counts.activeSessions,
                    'Daemons online':
                        ops.serverOverview!.counts.connectedDaemons,
                  },
                ),
                if (ops.serverOverview!.hasHealth()) ...[
                  const SizedBox(height: 12),
                  _SectionTitle('Health'),
                  Text(
                    'MongoDB: ${ops.serverOverview!.health.mongodb.status} · '
                    'Redis: ${ops.serverOverview!.health.redis.status} · '
                    'Runner: ${ops.serverOverview!.health.runner.status}',
                  ),
                ],
              ] else if (ops.workspaceCounts != null) ...[
                _SectionTitle('This Workspace'),
                _CountsGrid(
                  labels: {
                    'Agents': ops.workspaceCounts!.agents,
                    'MCP servers': ops.workspaceCounts!.mcpServers,
                    'Remote Agents': ops.workspaceCounts!.remoteAgents,
                    'Cron jobs': ops.workspaceCounts!.cronJobs,
                    'Daemon runtimes': ops.workspaceCounts!.daemonRuntimes,
                  },
                ),
                if (ops.serverHealth != null) ...[
                  const SizedBox(height: 12),
                  _SectionTitle('Server health'),
                  const Text(
                    'Process metrics from the Butter server (not Workspace-specific).',
                  ),
                  Text(
                    'Goroutines: ${ops.serverHealth!.goroutines} · '
                    'Memory: ${ops.serverHealth!.memoryUsedBytes} bytes',
                  ),
                ],
              ],
              const SizedBox(height: 16),
              _SectionTitle('Recent activity'),
              if (ops.activityEvents.isEmpty)
                const Text('No recent invocations to show.')
              else
                for (final e in ops.activityEvents)
                  ListTile(
                    dense: true,
                    title: Text(e.message.isEmpty ? e.id : e.message),
                    subtitle: Text(e.actor),
                  ),
              const SizedBox(height: 16),
              _SectionTitle('Cron timeline'),
              if (ops.overviewIsServerWide) ...[
                if (ops.cronBuckets.isEmpty)
                  const Text('No cron execution buckets in this range.')
                else
                  for (final b in ops.cronBuckets)
                    ListTile(
                      dense: true,
                      title: Text(
                        b.hasStart()
                            ? b.start.toDateTime().toLocal().toIso8601String()
                            : 'Bucket',
                      ),
                      subtitle: Text('OK ${b.success} · errors ${b.error}'),
                    ),
              ] else ...[
                const Text(
                  'Recent cron executions in this Workspace (see Cron jobs #464 for full management).',
                ),
                if (ops.recentCronExecutions.isEmpty)
                  const Text('No executions yet.')
                else
                  for (final ex in ops.recentCronExecutions.take(20))
                    ListTile(
                      dense: true,
                      title: Text(ex.jobName),
                      subtitle: Text(ex.status.name),
                    ),
              ],
              if (ops.overviewIsServerWide && ops.activityMetrics != null) ...[
                const SizedBox(height: 16),
                _SectionTitle('Activity metrics (7d)'),
                Text(
                  'Agent runs: ${ops.activityMetrics!.agentRuns} · '
                  'Automation runs: ${ops.activityMetrics!.automationRuns}',
                ),
              ],
            ],
          ),
        );
      },
    ),
  );
}

class _InvocationsPage extends StatefulWidget {
  const _InvocationsPage({required this.client, required this.ops});
  final ClientContext client;
  final WorkspaceOpsController ops;
  @override
  State<_InvocationsPage> createState() => _InvocationsPageState();
}

class _InvocationsPageState extends State<_InvocationsPage> {
  final _agentFilter = TextEditingController();
  final _sessionFilter = TextEditingController();

  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addPostFrameCallback((_) {
      unawaited(widget.ops.loadInvocations());
    });
  }

  @override
  void dispose() {
    _agentFilter.dispose();
    _sessionFilter.dispose();
    super.dispose();
  }

  void _applyFilters() {
    widget.ops.setInvocationFilters(
      agentId: _agentFilter.text,
      sessionId: _sessionFilter.text,
    );
    unawaited(widget.ops.loadInvocations());
  }

  @override
  Widget build(BuildContext context) => Scaffold(
    appBar: AppBar(title: const Text('Operations')),
    body: ListenableBuilder(
      listenable: widget.ops,
      builder: (context, _) {
        final ops = widget.ops;
        if (ops.invocationsState == ConfigLoad.loading &&
            ops.invocations.isEmpty) {
          return const Center(child: CircularProgressIndicator());
        }
        return Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            Padding(
              padding: const EdgeInsets.all(12),
              child: Text(
                'Agent invocations in ${widget.client.workspace?.name ?? ops.workspaceId}. '
                'Cancel behavior depends on how the run started; the detail screen explains each case.',
              ),
            ),
            if (ops.invocationsError != null)
              Padding(
                padding: const EdgeInsets.symmetric(horizontal: 12),
                child: Text(
                  ops.invocationsError!,
                  style: TextStyle(color: Theme.of(context).colorScheme.error),
                ),
              ),
            Padding(
              padding: const EdgeInsets.symmetric(horizontal: 12),
              child: Column(
                children: [
                  TextField(
                    key: const Key('invocation-filter-agent'),
                    controller: _agentFilter,
                    decoration: const InputDecoration(
                      labelText: 'Filter by Agent ID',
                      isDense: true,
                    ),
                    onSubmitted: (_) => _applyFilters(),
                  ),
                  TextField(
                    key: const Key('invocation-filter-session'),
                    controller: _sessionFilter,
                    decoration: const InputDecoration(
                      labelText: 'Filter by session ID',
                      isDense: true,
                    ),
                    onSubmitted: (_) => _applyFilters(),
                  ),
                  Align(
                    alignment: Alignment.centerRight,
                    child: TextButton(
                      onPressed: _applyFilters,
                      child: const Text('Apply filters'),
                    ),
                  ),
                ],
              ),
            ),
            Expanded(
              child: RefreshIndicator(
                onRefresh: () => ops.loadInvocations(),
                child: ListView.builder(
                  itemCount:
                      ops.invocations.length +
                      (ops.invocationsNextPage.isNotEmpty ? 1 : 0),
                  itemBuilder: (context, i) {
                    if (i >= ops.invocations.length) {
                      return TextButton(
                        onPressed: () => ops.loadInvocations(more: true),
                        child: const Text('Load more'),
                      );
                    }
                    final inv = ops.invocations[i];
                    return ListTile(
                      title: Text(
                        inv.agentName.isEmpty ? inv.agentId : inv.agentName,
                      ),
                      subtitle: Text(
                        '${invocationStatusLabel(inv.status)} · ${inv.source}',
                      ),
                      onTap: () async {
                        await Navigator.push<void>(
                          context,
                          MaterialPageRoute(
                            builder: (_) => _InvocationDetailPage(
                              client: widget.client,
                              ops: ops,
                              invocationId: inv.id,
                            ),
                          ),
                        );
                      },
                    );
                  },
                ),
              ),
            ),
          ],
        );
      },
    ),
  );
}

class _InvocationDetailPage extends StatefulWidget {
  const _InvocationDetailPage({
    required this.client,
    required this.ops,
    required this.invocationId,
  });
  final ClientContext client;
  final WorkspaceOpsController ops;
  final String invocationId;
  @override
  State<_InvocationDetailPage> createState() => _InvocationDetailPageState();
}

class _InvocationDetailPageState extends State<_InvocationDetailPage> {
  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addPostFrameCallback((_) {
      unawaited(widget.ops.loadInvocationDetail(widget.invocationId));
    });
  }

  Future<void> _confirmCancel(agentspb.Invocation inv) async {
    final ok = await showDialog<bool>(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('Cancel invocation?'),
        content: Text(invocationCancelExplanation(inv)),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context, false),
            child: const Text('Keep running'),
          ),
          FilledButton(
            onPressed: () => Navigator.pop(context, true),
            child: const Text('Cancel'),
          ),
        ],
      ),
    );
    if (ok == true && mounted) {
      await widget.ops.cancelInvocation(inv.id);
    }
  }

  @override
  Widget build(BuildContext context) => Scaffold(
    appBar: AppBar(title: const Text('Invocation')),
    body: ListenableBuilder(
      listenable: widget.ops,
      builder: (context, _) {
        final inv = widget.ops.selectedInvocation;
        if (inv == null || inv.id != widget.invocationId) {
          return const Center(child: CircularProgressIndicator());
        }
        return ListView(
          padding: const EdgeInsets.all(16),
          children: [
            Text('Status: ${invocationStatusLabel(inv.status)}'),
            Text('Source: ${inv.source}'),
            Text('Session: ${inv.sessionId}'),
            Text('Agent: ${inv.agentId}'),
            const SizedBox(height: 12),
            Text(invocationCancelExplanation(inv)),
            if (widget.ops.cancelMessage != null) ...[
              const SizedBox(height: 8),
              Text(widget.ops.cancelMessage!),
            ],
            if (invocationSupportsCancel(inv)) ...[
              const SizedBox(height: 16),
              FilledButton(
                onPressed: widget.ops.cancelBusy
                    ? null
                    : () => _confirmCancel(inv),
                child: widget.ops.cancelBusy
                    ? const Text('Cancelling…')
                    : const Text('Cancel invocation'),
              ),
            ],
          ],
        );
      },
    ),
  );
}

class _ApiTokensPage extends StatefulWidget {
  const _ApiTokensPage({required this.client, required this.ops});
  final ClientContext client;
  final WorkspaceOpsController ops;
  @override
  State<_ApiTokensPage> createState() => _ApiTokensPageState();
}

class _ApiTokensPageState extends State<_ApiTokensPage> {
  final _name = TextEditingController();

  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addPostFrameCallback((_) {
      unawaited(widget.ops.loadPersonalTokens());
    });
  }

  @override
  void dispose() {
    widget.ops.leaveTokenScreen(notify: false);
    _name.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final canManage = widget.client.canManageWorkspaceResources;
    return Scaffold(
      appBar: AppBar(title: const Text('API tokens')),
      body: ListenableBuilder(
        listenable: widget.ops,
        builder: (context, _) {
          final ops = widget.ops;
          return ListView(
            padding: const EdgeInsets.all(16),
            children: [
              const Text(
                'Personal API tokens authenticate automation against this '
                'Workspace. They are not sign-in tokens and are separate from '
                'Daemon worker tokens (see Daemon execution #455).',
              ),
              if (!canManage)
                const Padding(
                  padding: EdgeInsets.only(top: 8),
                  child: Text(
                    'Only workspace owners and admins can create or revoke tokens.',
                  ),
                ),
              if (ops.tokenWriteError != null)
                Padding(
                  padding: const EdgeInsets.only(top: 8),
                  child: Text(
                    ops.tokenWriteError!,
                    style: TextStyle(
                      color: Theme.of(context).colorScheme.error,
                    ),
                  ),
                ),
              if (ops.minted?.secret case final secret?) ...[
                const SizedBox(height: 16),
                Card(
                  child: Padding(
                    padding: const EdgeInsets.all(12),
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.stretch,
                      children: [
                        const Text(
                          'Copy this secret now. It will not be shown again.',
                          style: TextStyle(fontWeight: FontWeight.w600),
                        ),
                        SelectableText(secret),
                        TextButton(
                          onPressed: () {
                            Clipboard.setData(ClipboardData(text: secret));
                            ScaffoldMessenger.of(context).showSnackBar(
                              const SnackBar(content: Text('Copied')),
                            );
                          },
                          child: const Text('Copy secret'),
                        ),
                        FilledButton(
                          onPressed: ops.clearMintedSecret,
                          child: const Text('Done'),
                        ),
                      ],
                    ),
                  ),
                ),
              ],
              if (canManage && ops.minted == null) ...[
                const SizedBox(height: 16),
                TextField(
                  controller: _name,
                  decoration: const InputDecoration(labelText: 'Token name'),
                ),
                FilledButton(
                  onPressed: () async {
                    final name = _name.text.trim();
                    if (name.isEmpty) return;
                    final ok = await ops.createPersonalToken(name);
                    if (ok) _name.clear();
                  },
                  child: const Text('Create token'),
                ),
              ],
              const SizedBox(height: 16),
              _SectionTitle('Tokens in this Workspace'),
              if (ops.tokenState == ConfigLoad.loading)
                const LinearProgressIndicator()
              else if (ops.personalTokens.isEmpty)
                const Text('No personal API tokens.')
              else
                for (final t in ops.personalTokens)
                  ListTile(
                    title: Text(t.name),
                    subtitle: Text(
                      '${t.prefix}… · ${apiTokenExpiryLabel(t)}'
                      '${t.revoked ? ' · revoked' : ''}',
                    ),
                    trailing: canManage && !t.revoked
                        ? IconButton(
                            icon: const Icon(Icons.delete_outline),
                            onPressed: () async {
                              final ok = await showDialog<bool>(
                                context: context,
                                builder: (context) => AlertDialog(
                                  title: const Text('Revoke token?'),
                                  content: Text(
                                    'Revoke "${t.name}"? Anything using this '
                                    'token will stop working immediately.',
                                  ),
                                  actions: [
                                    TextButton(
                                      onPressed: () =>
                                          Navigator.pop(context, false),
                                      child: const Text('Keep token'),
                                    ),
                                    FilledButton(
                                      onPressed: () =>
                                          Navigator.pop(context, true),
                                      child: const Text('Revoke'),
                                    ),
                                  ],
                                ),
                              );
                              if (ok == true && context.mounted) {
                                await ops.revokePersonalToken(t.id);
                              }
                            },
                          )
                        : null,
                  ),
            ],
          );
        },
      ),
    );
  }
}

class _IntegrationsHubPage extends StatelessWidget {
  const _IntegrationsHubPage({required this.client});
  final ClientContext client;

  @override
  Widget build(BuildContext context) {
    final address = client.address;
    return Scaffold(
      appBar: AppBar(title: const Text('Integrations & manage')),
      body: ListView(
        padding: const EdgeInsets.all(16),
        children: [
          const Text(
            'Entry points and API bases for this server. Copy URLs for external '
            'tools; open linked screens to manage Workspace resources.',
          ),
          if (address != null) ...[
            const SizedBox(height: 16),
            _CopyRow(label: 'Connect RPC base', value: connectApiBase(address)),
            _CopyRow(
              label: 'OpenAI-compatible API',
              value: openAiCompatibleBase(address),
            ),
          ],
          const SizedBox(height: 16),
          _SectionTitle('Workspace resources'),
          ListTile(
            title: const Text('MCP servers'),
            onTap: () => openMcpServers(context, client),
          ),
          ListTile(
            title: const Text('Remote Agents'),
            onTap: () => openRemoteAgents(context, client),
          ),
          ListTile(
            title: const Text('Daemon execution (#455)'),
            onTap: () => openDaemons(context, client),
          ),
          ListTile(
            title: const Text('Cron jobs (#464)'),
            onTap: () => openCronJobs(context, client),
          ),
          ListTile(
            title: const Text('API tokens'),
            onTap: () => openApiTokens(context, client),
          ),
          ListTile(
            title: const Text('Operations / invocations'),
            onTap: () => openInvocations(context, client),
          ),
        ],
      ),
    );
  }
}

class _CopyRow extends StatelessWidget {
  const _CopyRow({required this.label, required this.value});
  final String label;
  final String value;
  @override
  Widget build(BuildContext context) => ListTile(
    title: Text(label),
    subtitle: SelectableText(value),
    trailing: IconButton(
      icon: const Icon(Icons.copy),
      onPressed: () {
        Clipboard.setData(ClipboardData(text: value));
        ScaffoldMessenger.of(
          context,
        ).showSnackBar(SnackBar(content: Text('Copied $label')));
      },
    ),
  );
}

class _SectionTitle extends StatelessWidget {
  const _SectionTitle(this.text);
  final String text;
  @override
  Widget build(BuildContext context) => Padding(
    padding: const EdgeInsets.only(bottom: 8),
    child: Text(text, style: Theme.of(context).textTheme.titleMedium),
  );
}

class _CountsGrid extends StatelessWidget {
  const _CountsGrid({required this.labels});
  final Map<String, int> labels;
  @override
  Widget build(BuildContext context) => Wrap(
    spacing: 12,
    runSpacing: 8,
    children: [
      for (final e in labels.entries) Chip(label: Text('${e.key}: ${e.value}')),
    ],
  );
}

class _ErrorBody extends StatelessWidget {
  const _ErrorBody({required this.message, required this.onRetry});
  final String message;
  final VoidCallback onRetry;
  @override
  Widget build(BuildContext context) => Center(
    child: Column(
      mainAxisSize: MainAxisSize.min,
      children: [
        Text(message, textAlign: TextAlign.center),
        TextButton(onPressed: onRetry, child: const Text('Retry')),
      ],
    ),
  );
}
