import 'dart:async';

import 'package:flutter/material.dart';

import '../../app/client_context.dart';
import '../../core/management_page.dart';
import '../../gen/agents/v1/workspace_memory.pb.dart';
import '../../gen/agents/v1/workspace_memory.pbenum.dart';
import 'memory_controller.dart';
import 'memory_draft.dart';

Future<void> openMemorySettings(
  BuildContext context,
  ClientContext client,
) async {
  final mem = client.memory;
  if (mem == null) return;
  await mem.loadConfig();
  if (!context.mounted) return;
  if (mem.configEnabled) {
    await mem.loadMemories();
  }
  if (!context.mounted) return;
  await openManagementPage(context, client, MemorySettingsPage(client: client));
}

class MemorySettingsPage extends StatefulWidget {
  const MemorySettingsPage({super.key, required this.client});
  final ClientContext client;

  @override
  State<MemorySettingsPage> createState() => _MemorySettingsPageState();
}

class _MemorySettingsPageState extends State<MemorySettingsPage> {
  late MemoryConfigDraft _draft;
  final _formKey = GlobalKey<FormState>();
  final _search = TextEditingController();
  WorkspaceMemoryScope _scope =
      WorkspaceMemoryScope.WORKSPACE_MEMORY_SCOPE_WORKSPACE;
  String _agentId = '';

  @override
  void initState() {
    super.initState();
    _draft = MemoryConfigDraft(base: widget.client.memory?.config);
  }

  @override
  void dispose() {
    _search.dispose();
    super.dispose();
  }

  void _syncDraft(MemoryController mem) {
    _draft = MemoryConfigDraft(base: mem.config);
  }

  Future<void> _save(MemoryController mem) async {
    if (!_formKey.currentState!.validate()) return;
    if (_draft.keyAction == MemoryApiKeyAction.clear) {
      final ok = await showDialog<bool>(
        context: context,
        builder: (ctx) => AlertDialog(
          title: const Text('Clear stored API key?'),
          content: const Text(
            'The encrypted key is removed from Butter. Saving still updates the URL and enabled flag.',
          ),
          actions: [
            TextButton(
              onPressed: () => Navigator.pop(ctx, false),
              child: const Text('Cancel'),
            ),
            FilledButton(
              onPressed: () => Navigator.pop(ctx, true),
              child: const Text('Clear on save'),
            ),
          ],
        ),
      );
      if (ok != true) return;
    }
    if (await mem.saveConfig(_draft)) {
      _syncDraft(mem);
      if (mem.configEnabled) {
        unawaited(mem.loadMemories(scope: _scope, agentId: _agentId));
      }
    }
  }

  Future<void> _removeConnection(MemoryController mem) async {
    final ok = await showDialog<bool>(
      context: context,
      builder: (ctx) => AlertDialog(
        title: const Text('Remove memory connection?'),
        content: const Text(
          'The mem0 URL and API key are removed from this Workspace. Agents with memory enabled keep running, without memory. Memories already stored on the mem0 server are not deleted.',
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(ctx, false),
            child: const Text('Cancel'),
          ),
          FilledButton(
            onPressed: () => Navigator.pop(ctx, true),
            child: const Text('Remove'),
          ),
        ],
      ),
    );
    if (ok == true) await mem.deleteConfig();
  }

  @override
  Widget build(BuildContext context) {
    final mem = widget.client.memory;
    if (mem == null) return const SizedBox.shrink();
    final canManage = widget.client.canManageWorkspaceResources;
    return ListenableBuilder(
      listenable: mem,
      builder: (context, _) {
        if (mem.config != null && _draft.baseUrl.isEmpty && !mem.writing) {
          _syncDraft(mem);
        }
        return Scaffold(
          appBar: AppBar(title: const Text('Memory')),
          body: ListView(
            padding: const EdgeInsets.all(16),
            children: [
              Text(
                '${widget.client.workspace?.name ?? 'Workspace'} · Memory',
                style: Theme.of(context).textTheme.headlineSmall,
              ),
              const SizedBox(height: 8),
              const Text(
                'Connect this Workspace to a self-hosted mem0 OSS server. The phone never talks to mem0 directly; Butter holds the API key.',
              ),
              if (!canManage)
                const Padding(
                  padding: EdgeInsets.only(top: 12),
                  child: Text(
                    'Only workspace owners and admins can change these settings.',
                  ),
                ),
              if (mem.configState == MemoryLoad.loading)
                const LinearProgressIndicator(semanticsLabel: 'Loading'),
              if (mem.configError != null)
                Text(
                  mem.configError!,
                  style: TextStyle(color: Theme.of(context).colorScheme.error),
                ),
              if (mem.unconfirmed)
                Text(
                  mem.writeError ?? memoryUnconfirmed,
                  style: TextStyle(color: Theme.of(context).colorScheme.error),
                ),
              const SizedBox(height: 12),
              Form(
                key: _formKey,
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.stretch,
                  children: [
                    TextFormField(
                      key: const Key('memory-base-url'),
                      initialValue: _draft.baseUrl,
                      decoration: const InputDecoration(labelText: 'Base URL'),
                      enabled: canManage && !mem.writing,
                      onChanged: (v) => _draft.baseUrl = v,
                      validator: (v) => validateMemoryBaseUrl(v ?? ''),
                    ),
                    SwitchListTile(
                      title: const Text('Enabled'),
                      subtitle: const Text(
                        'When off, memory-enabled agents run without memory. Saving while enabled checks the server first.',
                      ),
                      value: _draft.enabled,
                      onChanged: canManage && !mem.writing
                          ? (v) => setState(() => _draft.enabled = v)
                          : null,
                    ),
                    ListTile(
                      title: Text(
                        _draft.credentialSet
                            ? 'API key: set (never shown)'
                            : 'API key: not set',
                      ),
                      subtitle: const Text(
                        'Write-only: keep, replace, or clear on save.',
                      ),
                    ),
                    if (canManage) ...[
                      SegmentedButton<MemoryApiKeyAction>(
                        segments: const [
                          ButtonSegment(
                            value: MemoryApiKeyAction.keep,
                            label: Text('Keep'),
                          ),
                          ButtonSegment(
                            value: MemoryApiKeyAction.replace,
                            label: Text('Replace'),
                          ),
                          ButtonSegment(
                            value: MemoryApiKeyAction.clear,
                            label: Text('Clear'),
                          ),
                        ],
                        selected: {_draft.keyAction},
                        onSelectionChanged: mem.writing
                            ? null
                            : (s) => setState(() => _draft.keyAction = s.first),
                      ),
                      if (_draft.keyAction == MemoryApiKeyAction.replace)
                        TextFormField(
                          key: const Key('memory-api-key'),
                          obscureText: true,
                          autocorrect: false,
                          decoration: InputDecoration(
                            labelText: _draft.credentialSet
                                ? 'New API key'
                                : 'API key',
                          ),
                          enabled: !mem.writing,
                          onChanged: (v) => _draft.replacementKey = v,
                        ),
                    ],
                    if (mem.saveWarning != null)
                      Padding(
                        padding: const EdgeInsets.only(top: 8),
                        child: Text(
                          mem.saveWarning!,
                          style: TextStyle(
                            color: Theme.of(context).colorScheme.error,
                          ),
                        ),
                      ),
                    Wrap(
                      spacing: 8,
                      runSpacing: 8,
                      children: [
                        OutlinedButton(
                          key: const Key('memory-test'),
                          onPressed: mem.config == null || mem.testing
                              ? null
                              : () => mem.testConnection(),
                          child: Text(
                            mem.testing ? 'Testing…' : 'Test connection',
                          ),
                        ),
                        if (mem.config != null && canManage)
                          OutlinedButton(
                            key: const Key('memory-remove'),
                            onPressed: mem.writing
                                ? null
                                : () => _removeConnection(mem),
                            child: const Text('Remove connection'),
                          ),
                        if (canManage)
                          FilledButton(
                            key: const Key('memory-save'),
                            onPressed: mem.writing || mem.unconfirmed
                                ? null
                                : () => _save(mem),
                            child: Text(mem.writing ? 'Saving…' : 'Save'),
                          ),
                      ],
                    ),
                    if (mem.testMessage != null)
                      Padding(
                        padding: const EdgeInsets.only(top: 8),
                        child: Text(
                          mem.testOk == true
                              ? mem.testMessage!
                              : mem.testMessage!,
                          style: TextStyle(
                            color: mem.testOk == true
                                ? Theme.of(context).colorScheme.primary
                                : Theme.of(context).colorScheme.error,
                          ),
                        ),
                      ),
                  ],
                ),
              ),
              if (mem.configEnabled) ...[
                const SizedBox(height: 24),
                _MemoriesPanel(
                  client: widget.client,
                  mem: mem,
                  scope: _scope,
                  agentId: _agentId,
                  search: _search,
                  onScopeChanged: (s) => setState(() => _scope = s),
                  onAgentChanged: (id) => setState(() => _agentId = id),
                ),
              ],
            ],
          ),
        );
      },
    );
  }
}

class _MemoriesPanel extends StatelessWidget {
  const _MemoriesPanel({
    required this.client,
    required this.mem,
    required this.scope,
    required this.agentId,
    required this.search,
    required this.onScopeChanged,
    required this.onAgentChanged,
  });

  final ClientContext client;
  final MemoryController mem;
  final WorkspaceMemoryScope scope;
  final String agentId;
  final TextEditingController search;
  final ValueChanged<WorkspaceMemoryScope> onScopeChanged;
  final ValueChanged<String> onAgentChanged;

  @override
  Widget build(BuildContext context) {
    final canDelete = client.canManageWorkspaceResources;
    final showPrincipal = client.canManageWorkspaceResources;
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        Text('Stored memories', style: Theme.of(context).textTheme.titleMedium),
        const Text(
          'Listed newest first (up to the server cap). Search reaches older entries. Deletes require owner or admin.',
        ),
        SegmentedButton<WorkspaceMemoryScope>(
          segments: const [
            ButtonSegment(
              value: WorkspaceMemoryScope.WORKSPACE_MEMORY_SCOPE_WORKSPACE,
              label: Text('Workspace'),
            ),
            ButtonSegment(
              value: WorkspaceMemoryScope.WORKSPACE_MEMORY_SCOPE_AGENT,
              label: Text('Agent'),
            ),
          ],
          selected: {scope},
          onSelectionChanged: (s) {
            onScopeChanged(s.first);
            unawaited(mem.loadMemories(scope: s.first, agentId: agentId));
          },
        ),
        if (scope == WorkspaceMemoryScope.WORKSPACE_MEMORY_SCOPE_AGENT)
          DropdownMenu<String>(
            label: const Text('Agent ID'),
            initialSelection: agentId.isEmpty ? null : agentId,
            dropdownMenuEntries: [
              for (final a in client.agentChoices)
                DropdownMenuEntry(value: a.agentId, label: a.name),
            ],
            onSelected: (v) {
              if (v != null) {
                onAgentChanged(v);
                unawaited(mem.loadMemories(scope: scope, agentId: v));
              }
            },
          ),
        Row(
          children: [
            Expanded(
              child: TextField(
                controller: search,
                decoration: const InputDecoration(labelText: 'Search'),
                onSubmitted: (q) =>
                    mem.loadMemories(scope: scope, agentId: agentId, query: q),
              ),
            ),
            IconButton(
              tooltip: 'Search',
              onPressed: () => mem.loadMemories(
                scope: scope,
                agentId: agentId,
                query: search.text,
              ),
              icon: const Icon(Icons.search),
            ),
          ],
        ),
        if (mem.memoriesTruncated)
          Text(
            'Showing the newest ${mem.memoriesLimit} memories; more exist on the server. Use search to find others.',
          ),
        if (mem.memoriesState == MemoryLoad.loading)
          const LinearProgressIndicator(semanticsLabel: 'Loading memories'),
        if (mem.memoriesError != null)
          Text(
            mem.memoriesError!,
            style: TextStyle(color: Theme.of(context).colorScheme.error),
          ),
        for (final m in mem.memories)
          Card(
            child: ListTile(
              title: Text(m.memory),
              subtitle: Text(
                [
                  if (m.agentId.isNotEmpty) 'Agent: ${m.agentId}',
                  if (m.channel.isNotEmpty) 'Channel: ${m.channel}',
                  if (showPrincipal && m.principal.isNotEmpty)
                    'Principal: ${m.principal}',
                  if (m.score > 0) 'Score: ${m.score.toStringAsFixed(2)}',
                ].join(' · '),
              ),
              trailing: canDelete
                  ? IconButton(
                      key: Key('delete-memory-${m.id}'),
                      tooltip: 'Delete memory',
                      onPressed: mem.writing
                          ? null
                          : () async {
                              final ok = await showDialog<bool>(
                                context: context,
                                builder: (ctx) => AlertDialog(
                                  title: const Text('Delete this memory?'),
                                  content: Text(
                                    'Removes it from the mem0 server for this workspace. This cannot be undone.\n\n${m.memory}',
                                  ),
                                  actions: [
                                    TextButton(
                                      onPressed: () =>
                                          Navigator.pop(ctx, false),
                                      child: const Text('Cancel'),
                                    ),
                                    FilledButton(
                                      onPressed: () => Navigator.pop(ctx, true),
                                      child: const Text('Delete'),
                                    ),
                                  ],
                                ),
                              );
                              if (ok == true) await mem.deleteMemory(m.id);
                            },
                      icon: const Icon(Icons.delete_outline),
                    )
                  : null,
            ),
          ),
      ],
    );
  }
}
