import 'dart:async';

import 'package:flutter/material.dart';
import 'package:protobuf/well_known_types/google/protobuf/timestamp.pb.dart';

import '../../app/client_context.dart';
import '../../core/automation_errors.dart';
import '../../core/config_resource_controller.dart';
import '../../core/config_widgets.dart';
import '../../core/management_page.dart';
import '../../gen/agents/v1/agent.pb.dart';
import '../../gen/agents/v1/automation.pb.dart';
import 'automation_controller.dart';
import 'automation_draft.dart';

Future<void> openAutomations(BuildContext context, ClientContext client) async {
  final auto = client.automations;
  if (auto == null) return;
  unawaited(auto.load());
  await openManagementPage(context, client, AutomationListPage(client: client));
}

String _formatTime(Timestamp? ts) {
  if (ts == null) return '—';
  return ts.toDateTime().toLocal().toString();
}

class AutomationListPage extends StatelessWidget {
  const AutomationListPage({super.key, required this.client});
  final ClientContext client;

  @override
  Widget build(BuildContext context) {
    final auto = client.automations;
    if (auto == null) return const SizedBox.shrink();
    return ConfigScaffold(
      controller: auto,
      title: 'Automations',
      child: ListenableBuilder(
        listenable: auto,
        builder: (context, _) => ListView(
          padding: const EdgeInsets.all(16),
          children: [
            Wrap(
              spacing: 12,
              children: [
                FilledButton.icon(
                  key: const Key('add-automation'),
                  onPressed: auto.writing || auto.unconfirmed
                      ? null
                      : () => openManagementPage(
                          context,
                          client,
                          AutomationFormPage(
                            client: client,
                            draft: AutomationDraft.create(),
                          ),
                        ),
                  icon: const Icon(Icons.add),
                  label: const Text('Add automation'),
                ),
                TextButton.icon(
                  onPressed: auto.listState == ConfigLoad.loading
                      ? null
                      : auto.load,
                  icon: const Icon(Icons.refresh),
                  label: const Text('Refresh'),
                ),
              ],
            ),
            if (auto.listState == ConfigLoad.loading)
              const LinearProgressIndicator(semanticsLabel: 'Loading'),
            if (auto.listError != null) Text(auto.listError!),
            for (final a in auto.items)
              ListTile(
                title: Text(a.name),
                subtitle: Text(automationTriggerLabel(a.trigger.type)),
                onTap: () => openManagementPage(
                  context,
                  client,
                  AutomationDetailPage(client: client, name: a.name),
                ),
              ),
          ],
        ),
      ),
    );
  }
}

class AutomationDetailPage extends StatefulWidget {
  const AutomationDetailPage({
    super.key,
    required this.client,
    required this.name,
  });
  final ClientContext client;
  final String name;

  @override
  State<AutomationDetailPage> createState() => _AutomationDetailPageState();
}

class _AutomationDetailPageState extends State<AutomationDetailPage> {
  @override
  void initState() {
    super.initState();
    final auto = widget.client.automations;
    unawaited(auto?.loadItem(widget.name));
    unawaited(auto?.loadRuns(widget.name));
    unawaited(auto?.loadReferences());
  }

  Future<void> _confirmRun(AutomationController auto) async {
    final ok = await showDialog<bool>(
      context: context,
      builder: (ctx) => AlertDialog(
        title: const Text('Run automation now?'),
        content: const Text(
          'Starts one manual run on the server. If the connection drops, the run may still finish; refresh runs to see the outcome. This action is never retried automatically.',
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(ctx, false),
            child: const Text('Cancel'),
          ),
          FilledButton(
            onPressed: () => Navigator.pop(ctx, true),
            child: const Text('Run'),
          ),
        ],
      ),
    );
    if (ok == true) unawaited(auto.runNow(widget.name));
  }

  @override
  Widget build(BuildContext context) {
    final auto = widget.client.automations;
    final item = auto?.item;
    if (auto == null || item == null) {
      return const Scaffold(body: Center(child: CircularProgressIndicator()));
    }
    return Scaffold(
      appBar: AppBar(title: Text(widget.name)),
      body: ListView(
        padding: const EdgeInsets.all(16),
        children: [
          if (auto.runError != null)
            Padding(
              padding: const EdgeInsets.only(bottom: 8),
              child: Text(auto.runError!),
            ),
          SwitchListTile(
            title: const Text('Enabled'),
            value: item.enabled,
            onChanged: auto.writing || auto.unconfirmed
                ? null
                : (v) => auto.update(item.deepCopy()..enabled = v),
          ),
          ListTile(
            title: const Text('Trigger'),
            subtitle: Text(automationTriggerLabel(item.trigger.type)),
          ),
          FilledButton.icon(
            onPressed: auto.runBlocked || auto.unconfirmed || !item.enabled
                ? null
                : () => _confirmRun(auto),
            icon: auto.runSending
                ? const SizedBox(
                    width: 16,
                    height: 16,
                    child: CircularProgressIndicator(strokeWidth: 2),
                  )
                : const Icon(Icons.play_arrow),
            label: const Text('Run now'),
          ),
          const SizedBox(height: 16),
          Text('Recent runs', style: Theme.of(context).textTheme.titleMedium),
          if (auto.runsLoadError != null) Text(auto.runsLoadError!),
          for (final run in auto.runs)
            ListTile(
              title: Text(automationRunStatusLabel(run.status)),
              subtitle: Text(_formatTime(run.startedAt)),
              onTap: () => openManagementPage(
                context,
                widget.client,
                AutomationRunPage(
                  client: widget.client,
                  runId: run.id,
                  automationName: widget.name,
                ),
              ),
            ),
          TextButton(
            onPressed: auto.writing || auto.unconfirmed
                ? null
                : () => openManagementPage(
                    context,
                    widget.client,
                    AutomationFormPage(
                      client: widget.client,
                      draft: AutomationDraft.edit(item),
                    ),
                  ),
            child: const Text('Edit automation'),
          ),
        ],
      ),
    );
  }
}

class AutomationRunPage extends StatefulWidget {
  const AutomationRunPage({
    super.key,
    required this.client,
    required this.runId,
    required this.automationName,
  });
  final ClientContext client;
  final String runId;
  final String automationName;

  @override
  State<AutomationRunPage> createState() => _AutomationRunPageState();
}

class _AutomationRunPageState extends State<AutomationRunPage> {
  @override
  void initState() {
    super.initState();
    unawaited(widget.client.automations?.loadStepRuns(widget.runId));
  }

  @override
  Widget build(BuildContext context) {
    final auto = widget.client.automations;
    return Scaffold(
      appBar: AppBar(title: Text('Run ${widget.runId}')),
      body: auto == null
          ? const Center(child: CircularProgressIndicator())
          : ListView(
              padding: const EdgeInsets.all(16),
              children: [
                if (auto.stepError != null) Text(auto.stepError!),
                for (final step in auto.stepRuns)
                  ListTile(
                    title: Text('${step.order}. ${step.stepName}'),
                    subtitle: Text(
                      '${automationStepRunStatusLabel(step.status)}'
                      '${step.attemptCount > 1 ? ' · ${step.attemptCount} attempts' : ''}'
                      '${step.truncated ? ' · output truncated' : ''}',
                    ),
                    isThreeLine:
                        step.error.isNotEmpty || step.outputJson.isNotEmpty,
                    trailing: step.error.isNotEmpty
                        ? const Icon(Icons.error_outline)
                        : null,
                    onTap: () => showDialog<void>(
                      context: context,
                      builder: (ctx) => AlertDialog(
                        title: Text(step.stepName),
                        content: SingleChildScrollView(
                          child: Column(
                            crossAxisAlignment: CrossAxisAlignment.start,
                            children: [
                              if (step.error.isNotEmpty)
                                Text('Error: ${step.error}'),
                              if (step.outputJson.isNotEmpty)
                                Text('Output: ${step.outputJson}'),
                            ],
                          ),
                        ),
                        actions: [
                          TextButton(
                            onPressed: () => Navigator.pop(ctx),
                            child: const Text('Close'),
                          ),
                        ],
                      ),
                    ),
                  ),
              ],
            ),
    );
  }
}

class AutomationFormPage extends StatefulWidget {
  const AutomationFormPage({
    super.key,
    required this.client,
    required this.draft,
  });
  final ClientContext client;
  final AutomationDraft draft;

  @override
  State<AutomationFormPage> createState() => _AutomationFormPageState();
}

class _AutomationFormPageState extends State<AutomationFormPage> {
  late final AutomationDraft _draft = widget.draft;

  @override
  void initState() {
    super.initState();
    unawaited(widget.client.automations?.loadReferences());
  }

  Future<void> _save() async {
    final err = _draft.validate();
    if (err != null) {
      ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text(err)));
      return;
    }
    final auto = widget.client.automations;
    if (auto == null) return;
    final ok = _draft.isCreate
        ? await auto.create(_draft.build())
        : await auto.update(_draft.build());
    if (ok && mounted) Navigator.pop(context);
  }

  @override
  Widget build(BuildContext context) {
    final auto = widget.client.automations;
    final agents = widget.client.agents;
    return Scaffold(
      appBar: AppBar(
        title: Text(_draft.isCreate ? 'Add automation' : 'Edit automation'),
        actions: [
          TextButton(
            onPressed: auto?.writing == true ? null : _save,
            child: const Text('Save'),
          ),
        ],
      ),
      body: ListView(
        padding: const EdgeInsets.all(16),
        children: [
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
          DropdownButtonFormField<AutomationTriggerType>(
            decoration: const InputDecoration(labelText: 'Trigger'),
            initialValue: _draft.triggerType,
            items: [
              for (final t in AutomationTriggerType.values)
                if (t !=
                    AutomationTriggerType.AUTOMATION_TRIGGER_TYPE_UNSPECIFIED)
                  DropdownMenuItem(
                    value: t,
                    enabled: automationTriggerUsable(t),
                    child: Text(automationTriggerLabel(t)),
                  ),
            ],
            onChanged: (v) {
              if (v != null && automationTriggerUsable(v)) {
                setState(() => _draft.triggerType = v);
              }
            },
          ),
          if (_draft.triggerType ==
              AutomationTriggerType.AUTOMATION_TRIGGER_TYPE_SCHEDULE) ...[
            TextField(
              decoration: const InputDecoration(labelText: 'Schedule'),
              controller: TextEditingController(text: _draft.schedule)
                ..selection = TextSelection.collapsed(
                  offset: _draft.schedule.length,
                ),
              onChanged: (v) => setState(() => _draft.schedule = v),
            ),
            TextField(
              decoration: const InputDecoration(
                labelText: 'Timezone',
                helperText: 'IANA name, or empty for UTC',
              ),
              onChanged: (v) => setState(() => _draft.timezone = v),
            ),
          ],
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              Text('Steps', style: Theme.of(context).textTheme.titleMedium),
              TextButton.icon(
                onPressed: () => setState(() => _draft.steps.add(StepDraft())),
                icon: const Icon(Icons.add),
                label: const Text('Add step'),
              ),
            ],
          ),
          for (var i = 0; i < _draft.steps.length; i++)
            _StepEditor(
              draft: _draft.steps[i],
              agents: agents,
              notifyGroups: auto?.notifyGroups ?? const [],
              onChanged: () => setState(() {}),
              onRemove: () => setState(() => _draft.steps.removeAt(i)),
            ),
        ],
      ),
    );
  }
}

class _StepEditor extends StatelessWidget {
  const _StepEditor({
    required this.draft,
    required this.agents,
    required this.notifyGroups,
    required this.onChanged,
    required this.onRemove,
  });
  final StepDraft draft;
  final List<Agent> agents;
  final List<NotifyGroup> notifyGroups;
  final VoidCallback onChanged;
  final VoidCallback onRemove;

  @override
  Widget build(BuildContext context) {
    return Card(
      margin: const EdgeInsets.only(bottom: 12),
      child: Padding(
        padding: const EdgeInsets.all(12),
        child: Column(
          children: [
            Row(
              children: [
                Expanded(
                  child: TextField(
                    decoration: const InputDecoration(labelText: 'Step name'),
                    onChanged: (v) {
                      draft.name = v;
                      onChanged();
                    },
                  ),
                ),
                IconButton(
                  onPressed: onRemove,
                  icon: const Icon(Icons.delete_outline),
                ),
              ],
            ),
            DropdownButtonFormField<AutomationStepType>(
              decoration: const InputDecoration(labelText: 'Step type'),
              initialValue: draft.type,
              items: [
                for (final t in AutomationStepType.values)
                  if (t != AutomationStepType.AUTOMATION_STEP_TYPE_UNSPECIFIED)
                    DropdownMenuItem(
                      value: t,
                      child: Text(automationStepTypeLabel(t)),
                    ),
              ],
              onChanged: (v) {
                if (v != null) {
                  draft.type = v;
                  onChanged();
                }
              },
            ),
            if (draft.type ==
                AutomationStepType.AUTOMATION_STEP_TYPE_INVOKE_AGENT)
              DropdownButtonFormField<String>(
                decoration: const InputDecoration(labelText: 'Agent'),
                items: [
                  for (final a in agents)
                    DropdownMenuItem(value: a.agentId, child: Text(a.name)),
                ],
                onChanged: (v) {
                  draft.agentId = v ?? '';
                  onChanged();
                },
              ),
            if (draft.type ==
                AutomationStepType.AUTOMATION_STEP_TYPE_CALL_WEBHOOK)
              TextField(
                decoration: const InputDecoration(labelText: 'Webhook URL'),
                onChanged: (v) {
                  draft.webhookUrl = v;
                  onChanged();
                },
              ),
            if (draft.type ==
                AutomationStepType.AUTOMATION_STEP_TYPE_SEND_NOTIFY_GROUP)
              DropdownButtonFormField<String>(
                decoration: const InputDecoration(labelText: 'Notify Group'),
                items: [
                  for (final g in notifyGroups)
                    DropdownMenuItem(value: g.name, child: Text(g.name)),
                ],
                onChanged: (v) {
                  draft.notifyGroupName = v ?? '';
                  onChanged();
                },
              ),
          ],
        ),
      ),
    );
  }
}
