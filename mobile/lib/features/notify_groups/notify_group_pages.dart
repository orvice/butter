import 'dart:async';

import 'package:flutter/material.dart';

import '../../app/client_context.dart';
import '../../core/config_resource_controller.dart';
import '../../core/config_widgets.dart';
import '../../core/management_page.dart';
import '../../gen/agents/v1/agent.pb.dart';
import '../../gen/agents/v1/telegram.pb.dart';
import 'notify_group_controller.dart';
import 'notify_group_draft.dart';

Future<void> openNotifyGroups(
  BuildContext context,
  ClientContext client,
) async {
  final ng = client.notifyGroups;
  if (ng == null) return;
  unawaited(ng.load());
  await openManagementPage(
    context,
    client,
    NotifyGroupListPage(client: client),
  );
}

class NotifyGroupListPage extends StatelessWidget {
  const NotifyGroupListPage({super.key, required this.client});
  final ClientContext client;

  @override
  Widget build(BuildContext context) {
    final ng = client.notifyGroups;
    if (ng == null) return const SizedBox.shrink();
    return ConfigScaffold(
      controller: ng,
      title: 'Notify Groups',
      child: ListenableBuilder(
        listenable: ng,
        builder: (context, _) => ListView(
          padding: const EdgeInsets.all(16),
          children: [
            Wrap(
              spacing: 12,
              children: [
                FilledButton.icon(
                  key: const Key('add-notify-group'),
                  onPressed: ng.writing || ng.unconfirmed
                      ? null
                      : () => openManagementPage(
                          context,
                          client,
                          NotifyGroupFormPage(
                            client: client,
                            draft: NotifyGroupDraft.create(),
                          ),
                        ),
                  icon: const Icon(Icons.add),
                  label: const Text('Add Notify Group'),
                ),
                TextButton.icon(
                  onPressed: ng.listState == ConfigLoad.loading
                      ? null
                      : ng.load,
                  icon: const Icon(Icons.refresh),
                  label: const Text('Refresh'),
                ),
              ],
            ),
            if (ng.listState == ConfigLoad.loading)
              const LinearProgressIndicator(semanticsLabel: 'Loading'),
            if (ng.listError != null) Text(ng.listError!),
            if (ng.listState == ConfigLoad.ready && ng.items.isEmpty)
              const Text('No Notify Groups in this Workspace.'),
            for (final g in ng.items)
              ListTile(
                title: Text(g.name),
                subtitle: Text(
                  g.enabled
                      ? '${g.targets.length} target(s)'
                      : 'Disabled · ${g.targets.length} target(s)',
                ),
                onTap: () => openManagementPage(
                  context,
                  client,
                  NotifyGroupDetailPage(client: client, name: g.name),
                ),
              ),
          ],
        ),
      ),
    );
  }
}

class NotifyGroupDetailPage extends StatefulWidget {
  const NotifyGroupDetailPage({
    super.key,
    required this.client,
    required this.name,
  });
  final ClientContext client;
  final String name;

  @override
  State<NotifyGroupDetailPage> createState() => _NotifyGroupDetailPageState();
}

class _NotifyGroupDetailPageState extends State<NotifyGroupDetailPage> {
  @override
  void initState() {
    super.initState();
    unawaited(widget.client.notifyGroups?.loadItem(widget.name));
  }

  Future<void> _delete(BuildContext context, NotifyGroupController ng) async {
    final ok = await showDialog<bool>(
      context: context,
      builder: (ctx) => AlertDialog(
        title: const Text('Delete Notify Group?'),
        content: Text('Remove "${widget.name}"? This cannot be undone.'),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(ctx, false),
            child: const Text('Cancel'),
          ),
          FilledButton(
            onPressed: () => Navigator.pop(ctx, true),
            child: const Text('Delete'),
          ),
        ],
      ),
    );
    if (ok != true || !context.mounted) return;
    if (await ng.delete(widget.name) && context.mounted) {
      Navigator.pop(context);
    }
  }

  @override
  Widget build(BuildContext context) {
    final ng = widget.client.notifyGroups;
    final group = ng?.item;
    return Scaffold(
      appBar: AppBar(title: Text(widget.name)),
      body: ng == null || group == null
          ? const Center(child: CircularProgressIndicator())
          : ListView(
              padding: const EdgeInsets.all(16),
              children: [
                SwitchListTile(
                  title: const Text('Enabled'),
                  value: group.enabled,
                  onChanged: ng.writing || ng.unconfirmed
                      ? null
                      : (v) => ng.toggleEnabled(group, v),
                ),
                for (final t in group.targets)
                  ListTile(
                    title: Text(
                      t.name.isEmpty ? notifyTargetTypeLabel(t.type) : t.name,
                    ),
                    subtitle: Text(notifyTargetTypeLabel(t.type)),
                  ),
                const SizedBox(height: 16),
                FilledButton(
                  onPressed: ng.writing || ng.unconfirmed
                      ? null
                      : () => openManagementPage(
                          context,
                          widget.client,
                          NotifyGroupFormPage(
                            client: widget.client,
                            draft: NotifyGroupDraft.edit(group),
                          ),
                        ),
                  child: const Text('Edit'),
                ),
                TextButton(
                  onPressed: ng.writing || ng.unconfirmed
                      ? null
                      : () => _delete(context, ng),
                  child: const Text('Delete Notify Group'),
                ),
              ],
            ),
    );
  }
}

class NotifyGroupFormPage extends StatefulWidget {
  const NotifyGroupFormPage({
    super.key,
    required this.client,
    required this.draft,
  });
  final ClientContext client;
  final NotifyGroupDraft draft;

  @override
  State<NotifyGroupFormPage> createState() => _NotifyGroupFormPageState();
}

class _NotifyGroupFormPageState extends State<NotifyGroupFormPage> {
  late final NotifyGroupDraft _draft = widget.draft;

  @override
  void initState() {
    super.initState();
    unawaited(widget.client.notifyGroups?.loadReferences());
  }

  Future<void> _save() async {
    final err = _draft.validate();
    if (err != null) {
      ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text(err)));
      return;
    }
    final ng = widget.client.notifyGroups;
    if (ng == null) return;
    final ok = _draft.isCreate
        ? await ng.create(_draft.build())
        : await ng.update(_draft.build());
    if (ok && mounted) Navigator.pop(context);
  }

  @override
  Widget build(BuildContext context) {
    final ng = widget.client.notifyGroups;
    return Scaffold(
      appBar: AppBar(
        title: Text(_draft.isCreate ? 'Add Notify Group' : 'Edit Notify Group'),
        actions: [
          TextButton(
            onPressed: ng?.writing == true ? null : _save,
            child: const Text('Save'),
          ),
        ],
      ),
      body: ListView(
        padding: const EdgeInsets.all(16),
        children: [
          if (ng?.writeError != null)
            Padding(
              padding: const EdgeInsets.only(bottom: 8),
              child: Text(ng!.writeError!),
            ),
          TextField(
            decoration: const InputDecoration(labelText: 'Name'),
            enabled: _draft.isCreate,
            controller: TextEditingController(text: _draft.name)
              ..selection = TextSelection.collapsed(offset: _draft.name.length),
            onChanged: (v) => setState(() => _draft.name = v),
          ),
          SwitchListTile(
            title: const Text('Enabled'),
            value: _draft.enabled,
            onChanged: (v) => setState(() => _draft.enabled = v),
          ),
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              Text('Targets', style: Theme.of(context).textTheme.titleMedium),
              TextButton.icon(
                onPressed: () =>
                    setState(() => _draft.targets.add(NotifyTargetDraft())),
                icon: const Icon(Icons.add),
                label: const Text('Add target'),
              ),
            ],
          ),
          for (var i = 0; i < _draft.targets.length; i++)
            _TargetEditor(
              draft: _draft.targets[i],
              destinations: ng?.telegramDestinations ?? const [],
              onChanged: () => setState(() {}),
              onRemove: () => setState(() => _draft.targets.removeAt(i)),
            ),
        ],
      ),
    );
  }
}

class _TargetEditor extends StatelessWidget {
  const _TargetEditor({
    required this.draft,
    required this.destinations,
    required this.onChanged,
    required this.onRemove,
  });
  final NotifyTargetDraft draft;
  final List<TelegramDestination> destinations;
  final VoidCallback onChanged;
  final VoidCallback onRemove;

  @override
  Widget build(BuildContext context) {
    return Card(
      margin: const EdgeInsets.only(bottom: 12),
      child: Padding(
        padding: const EdgeInsets.all(12),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                Expanded(
                  child: DropdownButtonFormField<NotifyTargetType>(
                    decoration: const InputDecoration(labelText: 'Type'),
                    initialValue: draft.type,
                    items: [
                      for (final t in NotifyTargetType.values)
                        if (t !=
                            NotifyTargetType.NOTIFY_TARGET_TYPE_UNSPECIFIED)
                          DropdownMenuItem(
                            value: t,
                            child: Text(notifyTargetTypeLabel(t)),
                          ),
                    ],
                    onChanged: (v) {
                      if (v != null) {
                        draft.type = v;
                        onChanged();
                      }
                    },
                  ),
                ),
                IconButton(
                  onPressed: onRemove,
                  icon: const Icon(Icons.delete_outline),
                ),
              ],
            ),
            TextField(
              decoration: const InputDecoration(
                labelText: 'Target name (optional)',
              ),
              controller: TextEditingController(
                text: draft.name,
              )..selection = TextSelection.collapsed(offset: draft.name.length),
              onChanged: (v) {
                draft.name = v;
                onChanged();
              },
            ),
            if (draft.type == NotifyTargetType.NOTIFY_TARGET_TYPE_TELEGRAM)
              DropdownButtonFormField<String>(
                decoration: const InputDecoration(
                  labelText: 'Telegram Destination',
                ),
                initialValue: draft.telegramDestinationId.isEmpty
                    ? null
                    : draft.telegramDestinationId,
                items: [
                  for (final d in destinations)
                    DropdownMenuItem(
                      value: d.id,
                      child: Text(d.name.isEmpty ? d.key : d.name),
                    ),
                ],
                onChanged: (v) {
                  draft.telegramDestinationId = v ?? '';
                  onChanged();
                },
              ),
            if (draft.type == NotifyTargetType.NOTIFY_TARGET_TYPE_LARK_WEBHOOK)
              TextField(
                decoration: const InputDecoration(
                  labelText: 'Lark webhook URL',
                ),
                onChanged: (v) {
                  draft.larkWebhookUrl = v;
                  onChanged();
                },
              ),
            if (draft.type ==
                NotifyTargetType.NOTIFY_TARGET_TYPE_DISCORD_WEBHOOK)
              TextField(
                decoration: const InputDecoration(
                  labelText: 'Discord webhook URL',
                ),
                onChanged: (v) {
                  draft.discordWebhookUrl = v;
                  onChanged();
                },
              ),
          ],
        ),
      ),
    );
  }
}
