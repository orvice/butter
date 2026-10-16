import 'dart:async';

import 'package:flutter/material.dart';

import '../../app/client_context.dart';
import '../../core/management_page.dart';
import 'agent_lifecycle_controller.dart';

Future<void> openAgentContentEditor(
  BuildContext context,
  ClientContext client,
  String agentId,
) async {
  final lifecycle = client.agentLifecycle;
  if (lifecycle == null) return;
  unawaited(lifecycle.loadContent(agentId));
  await openManagementPage(
    context,
    client,
    AgentContentPage(client: client, agentId: agentId),
  );
}

class AgentContentPage extends StatefulWidget {
  const AgentContentPage({
    super.key,
    required this.client,
    required this.agentId,
  });
  final ClientContext client;
  final String agentId;

  @override
  State<AgentContentPage> createState() => _AgentContentPageState();
}

class _AgentContentPageState extends State<AgentContentPage> {
  late final TextEditingController _description;
  late final TextEditingController _prompt;
  late final TextEditingController _global;
  late final TextEditingController _rollbackSha;

  @override
  void initState() {
    super.initState();
    _description = TextEditingController();
    _prompt = TextEditingController();
    _global = TextEditingController();
    _rollbackSha = TextEditingController();
    widget.client.agentLifecycle?.addListener(_syncFields);
  }

  void _syncFields() {
    final draft = widget.client.agentLifecycle?.contentDraft;
    if (draft == null || draft.agentId != widget.agentId) return;
    if (_description.text != draft.description) {
      _description.text = draft.description;
    }
    if (_prompt.text != draft.prompt) _prompt.text = draft.prompt;
    if (_global.text != draft.globalPrompt) _global.text = draft.globalPrompt;
  }

  @override
  void dispose() {
    widget.client.agentLifecycle?.removeListener(_syncFields);
    _description.dispose();
    _prompt.dispose();
    _global.dispose();
    _rollbackSha.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final lifecycle = widget.client.agentLifecycle;
    if (lifecycle == null) return const SizedBox.shrink();
    return ListenableBuilder(
      listenable: lifecycle,
      builder: (context, _) {
        final draft = lifecycle.contentDraft;
        if (draft != null && draft.agentId == widget.agentId) {
          draft.description = _description.text;
          draft.prompt = _prompt.text;
          draft.globalPrompt = _global.text;
        }
        return Scaffold(
          appBar: AppBar(title: const Text('Agent Content')),
          body: ListView(
            padding: const EdgeInsets.all(16),
            children: [
              Text(
                'Agent ${widget.agentId} · Git-managed description and prompts',
                style: Theme.of(context).textTheme.titleMedium,
              ),
              if (draft?.baseRevision.isNotEmpty == true)
                SelectableText('Active revision: ${draft!.baseRevision}'),
              if (lifecycle.contentState == LifecycleLoad.loading)
                const LinearProgressIndicator(semanticsLabel: 'Loading'),
              if (lifecycle.contentError != null) Text(lifecycle.contentError!),
              if (lifecycle.contentValidationErrors != null) ...[
                const SizedBox(height: 8),
                Text(lifecycle.contentValidationErrors!),
                TextButton(
                  onPressed: lifecycle.writing
                      ? null
                      : () => lifecycle.loadContent(widget.agentId),
                  child: const Text('Reload from server'),
                ),
              ],
              if (lifecycle.unconfirmed)
                Text(lifecycle.writeError ?? lifecycleUnconfirmed),
              const SizedBox(height: 12),
              TextField(
                controller: _description,
                decoration: const InputDecoration(
                  labelText: 'Description (description.md)',
                ),
                minLines: 2,
                maxLines: 6,
                readOnly: !widget.client.canManageWorkspaceResources,
              ),
              TextField(
                controller: _prompt,
                decoration: const InputDecoration(
                  labelText: 'Prompt (prompt.md)',
                ),
                minLines: 4,
                maxLines: 12,
                readOnly: !widget.client.canManageWorkspaceResources,
              ),
              TextField(
                controller: _global,
                decoration: const InputDecoration(
                  labelText: 'Global prompt (global-prompt.md)',
                ),
                minLines: 2,
                maxLines: 8,
                readOnly: !widget.client.canManageWorkspaceResources,
              ),
              if (widget.client.canManageWorkspaceResources) ...[
                const SizedBox(height: 16),
                FilledButton(
                  onPressed: lifecycle.writing || draft == null || !draft.dirty
                      ? null
                      : () async {
                          final ok = await lifecycle.commitContent();
                          if (context.mounted && ok) {
                            ScaffoldMessenger.of(context).showSnackBar(
                              SnackBar(
                                content: Text(
                                  lifecycle.lastCommitSha == null
                                      ? 'Content committed.'
                                      : 'Committed ${lifecycle.lastCommitSha}.',
                                ),
                              ),
                            );
                          }
                        },
                  child: const Text('Commit content to Git'),
                ),
                const SizedBox(height: 24),
                TextField(
                  controller: _rollbackSha,
                  decoration: const InputDecoration(
                    labelText: 'Rollback target commit SHA',
                  ),
                ),
                OutlinedButton(
                  onPressed: lifecycle.writing || _rollbackSha.text.isEmpty
                      ? null
                      : () async {
                          final ok = await lifecycle.rollbackContent(
                            _rollbackSha.text.trim(),
                          );
                          if (context.mounted && ok) {
                            ScaffoldMessenger.of(context).showSnackBar(
                              const SnackBar(
                                content: Text('Rollback committed.'),
                              ),
                            );
                          }
                        },
                  child: const Text('Rollback content to revision'),
                ),
              ] else
                const Text(
                  'Editing Git-owned content requires the workspace owner or admin role.',
                ),
            ],
          ),
        );
      },
    );
  }
}
