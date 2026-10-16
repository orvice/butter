import 'dart:async';

import 'package:flutter/material.dart';
import 'package:protobuf/well_known_types/google/protobuf/timestamp.pb.dart';

import '../../app/client_context.dart';
import '../../core/config_resource_controller.dart';
import '../../core/management_page.dart';
import '../../gen/agents/v1/agent.pb.dart';
import '../../gen/agents/v1/cron.pb.dart';
import 'cron_controller.dart';
import 'cron_draft.dart';

Future<void> openCronJobs(BuildContext context, ClientContext client) async {
  final cron = client.cron;
  if (cron == null) return;
  unawaited(cron.load());
  await openManagementPage(context, client, CronListPage(client: client));
}

String _formatTime(Timestamp? ts) {
  if (ts == null) return '—';
  return ts.toDateTime().toLocal().toString();
}

class CronListPage extends StatefulWidget {
  const CronListPage({super.key, required this.client});
  final ClientContext client;

  @override
  State<CronListPage> createState() => _CronListPageState();
}

class _CronListPageState extends State<CronListPage> {
  CronController? get _cron => widget.client.cron;

  @override
  void initState() {
    super.initState();
    _cron?.addListener(_onChange);
  }

  @override
  void dispose() {
    _cron?.removeListener(_onChange);
    super.dispose();
  }

  void _onChange() {
    if (mounted) setState(() {});
  }

  @override
  Widget build(BuildContext context) {
    final cron = _cron;
    if (cron == null) {
      return const Scaffold(
        body: Center(child: Text('Sign in to a Workspace.')),
      );
    }
    return Scaffold(
      appBar: AppBar(
        title: const Text('Cron jobs'),
        actions: [
          IconButton(
            onPressed: cron.listState == ConfigLoad.loading ? null : cron.load,
            icon: const Icon(Icons.refresh),
            tooltip: 'Refresh',
          ),
        ],
      ),
      floatingActionButton: FloatingActionButton.extended(
        onPressed: cron.writing || cron.unconfirmed
            ? null
            : () => _openEditor(context, CronJobDraft.create()),
        icon: const Icon(Icons.add),
        label: const Text('Add cron job'),
      ),
      body: ListView(
        padding: const EdgeInsets.all(16),
        children: [
          if (cron.writeNotice != null)
            Padding(
              padding: const EdgeInsets.only(bottom: 8),
              child: Text(cron.writeNotice!),
            ),
          if (cron.listState == ConfigLoad.loading && cron.items.isEmpty)
            const Center(child: CircularProgressIndicator()),
          if (cron.listError != null) Text(cron.listError!),
          if (cron.listState == ConfigLoad.ready && cron.items.isEmpty)
            const Text('No cron jobs in this Workspace.'),
          for (final job in cron.items)
            ListTile(
              title: Text(job.name),
              subtitle: Text(
                [
                  job.schedule,
                  if (job.timezone.isNotEmpty) job.timezone,
                  job.enabled ? 'Enabled' : 'Disabled',
                  if (job.agentId.isNotEmpty) 'Agent ${job.agentId}',
                ].join(' · '),
              ),
              onTap: () => _openDetail(context, job.name),
            ),
        ],
      ),
    );
  }

  Future<void> _openDetail(BuildContext context, String name) async {
    final cron = _cron;
    if (cron == null) return;
    await cron.loadItem(name);
    unawaited(cron.loadExecutions(name));
    unawaited(cron.loadReferences());
    if (!context.mounted) return;
    await Navigator.of(context).push<void>(
      MaterialPageRoute(
        builder: (_) => CronDetailPage(client: widget.client, jobName: name),
      ),
    );
  }

  Future<void> _openEditor(BuildContext context, CronJobDraft draft) async {
    final cron = _cron;
    if (cron == null) return;
    unawaited(cron.loadReferences());
    if (!context.mounted) return;
    await Navigator.of(context).push<void>(
      MaterialPageRoute(
        builder: (_) => CronEditorPage(client: widget.client, draft: draft),
      ),
    );
  }
}

class CronDetailPage extends StatefulWidget {
  const CronDetailPage({
    super.key,
    required this.client,
    required this.jobName,
  });
  final ClientContext client;
  final String jobName;

  @override
  State<CronDetailPage> createState() => _CronDetailPageState();
}

class _CronDetailPageState extends State<CronDetailPage> {
  CronController? get _cron => widget.client.cron;

  @override
  void initState() {
    super.initState();
    _cron?.addListener(_onChange);
  }

  @override
  void dispose() {
    _cron?.removeListener(_onChange);
    super.dispose();
  }

  void _onChange() {
    if (mounted) setState(() {});
  }

  Future<void> _confirmRun(CronController cron) async {
    final ok = await showDialog<bool>(
      context: context,
      builder: (ctx) => AlertDialog(
        title: const Text('Run now?'),
        content: Text(
          'Run "${widget.jobName}" immediately? This starts one execution and is not repeated if the answer is lost.',
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
    if (ok == true) unawaited(cron.runNow(widget.jobName));
  }

  @override
  Widget build(BuildContext context) {
    final cron = _cron;
    final job = cron?.item;
    if (cron == null || job == null) {
      return Scaffold(
        appBar: AppBar(title: Text(widget.jobName)),
        body: cron?.itemState == ConfigLoad.loading
            ? const Center(child: CircularProgressIndicator())
            : Text(cron?.itemError ?? 'Job not loaded.'),
      );
    }
    return Scaffold(
      appBar: AppBar(
        title: Text(job.name),
        actions: [
          IconButton(
            tooltip: 'Edit',
            onPressed: cron.writing || cron.unconfirmed
                ? null
                : () => Navigator.of(context).push<void>(
                    MaterialPageRoute(
                      builder: (_) => CronEditorPage(
                        client: widget.client,
                        draft: CronJobDraft.edit(job),
                      ),
                    ),
                  ),
            icon: const Icon(Icons.edit),
          ),
        ],
      ),
      body: ListView(
        padding: const EdgeInsets.all(16),
        children: [
          if (cron.writeNotice != null)
            Padding(
              padding: const EdgeInsets.only(bottom: 8),
              child: Text(cron.writeNotice!),
            ),
          if (cron.runError != null)
            Padding(
              padding: const EdgeInsets.only(bottom: 8),
              child: Text(cron.runError!),
            ),
          SwitchListTile(
            title: const Text('Enabled'),
            value: job.enabled,
            onChanged: cron.writing || cron.unconfirmed
                ? null
                : (v) => cron.toggleEnabled(job, v),
          ),
          ListTile(
            title: const Text('Schedule'),
            subtitle: Text(
              '${job.schedule}${job.timezone.isNotEmpty ? ' (${job.timezone})' : ''}',
            ),
          ),
          ListTile(
            title: const Text('Agent'),
            subtitle: Text(job.agentId.isEmpty ? job.agentName : job.agentId),
          ),
          if (job.input.isNotEmpty)
            ListTile(title: const Text('Input'), subtitle: Text(job.input)),
          Row(
            children: [
              FilledButton.icon(
                onPressed: cron.runBlocked || cron.unconfirmed
                    ? null
                    : () => _confirmRun(cron),
                icon: cron.runSending
                    ? const SizedBox(
                        width: 16,
                        height: 16,
                        child: CircularProgressIndicator(strokeWidth: 2),
                      )
                    : const Icon(Icons.play_arrow),
                label: const Text('Run now'),
              ),
            ],
          ),
          const SizedBox(height: 24),
          Text('Executions', style: Theme.of(context).textTheme.titleMedium),
          if (cron.executionError != null) Text(cron.executionError!),
          for (final exec in cron.executions)
            ListTile(
              title: Text(cronExecutionStatusLabel(exec.status)),
              subtitle: Text(_formatTime(exec.startedAt)),
              trailing: exec.truncated ? const Text('Truncated') : null,
              onTap: () => Navigator.of(context).push<void>(
                MaterialPageRoute(
                  builder: (_) =>
                      CronExecutionPage(client: widget.client, execution: exec),
                ),
              ),
            ),
          if (cron.executionNextToken.isNotEmpty)
            TextButton(
              onPressed: cron.executionState == ConfigLoad.loading
                  ? null
                  : () => cron.loadExecutions(widget.jobName, more: true),
              child: const Text('Load more executions'),
            ),
          const SizedBox(height: 16),
          OutlinedButton(
            onPressed: cron.writing || cron.unconfirmed
                ? null
                : () => _confirmDelete(cron),
            child: const Text('Delete cron job'),
          ),
        ],
      ),
    );
  }

  Future<void> _confirmDelete(CronController cron) async {
    final ok = await showDialog<bool>(
      context: context,
      builder: (ctx) => AlertDialog(
        title: const Text('Delete cron job?'),
        content: Text(
          'Delete "${widget.jobName}" and unschedule it? This cannot be undone from the App.',
        ),
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
    if (ok != true) return;
    final deleted = await cron.delete(widget.jobName);
    if (deleted && mounted) Navigator.of(context).pop();
  }
}

class CronExecutionPage extends StatefulWidget {
  const CronExecutionPage({
    super.key,
    required this.client,
    required this.execution,
  });
  final ClientContext client;
  final CronExecution execution;

  @override
  State<CronExecutionPage> createState() => _CronExecutionPageState();
}

class _CronExecutionPageState extends State<CronExecutionPage> {
  final _replyController = TextEditingController();

  CronController? get _cron => widget.client.cron;

  @override
  void dispose() {
    _replyController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final exec = widget.execution;
    final waiting =
        exec.status == CronExecutionStatus.CRON_EXECUTION_STATUS_WAITING_INPUT;
    return Scaffold(
      appBar: AppBar(title: Text('Execution ${exec.id}')),
      body: ListView(
        padding: const EdgeInsets.all(16),
        children: [
          ListTile(
            title: const Text('Status'),
            subtitle: Text(cronExecutionStatusLabel(exec.status)),
          ),
          ListTile(
            title: const Text('Started'),
            subtitle: Text(_formatTime(exec.startedAt)),
          ),
          if (exec.hasFinishedAt())
            ListTile(
              title: const Text('Finished'),
              subtitle: Text(_formatTime(exec.finishedAt)),
            ),
          if (exec.durationMs > 0)
            ListTile(
              title: const Text('Duration'),
              subtitle: Text('${exec.durationMs} ms'),
            ),
          if (exec.skippedReason.isNotEmpty)
            ListTile(
              title: const Text('Skipped'),
              subtitle: Text(exec.skippedReason),
            ),
          if (exec.error.isNotEmpty)
            ListTile(title: const Text('Error'), subtitle: Text(exec.error)),
          if (exec.output.isNotEmpty)
            ListTile(
              title: Text(exec.truncated ? 'Output (truncated)' : 'Output'),
              subtitle: SelectableText(exec.output),
            ),
          if (waiting) ...[
            const Divider(),
            Text(
              'Waiting for human input',
              style: Theme.of(context).textTheme.titleMedium,
            ),
            ListTile(
              title: const Text('Session'),
              subtitle: Text(
                '${exec.sessionAppName} / ${exec.sessionUserId} / ${exec.sessionId}',
              ),
            ),
            ListTile(
              title: const Text('Agent ID'),
              subtitle: Text(exec.agentId),
            ),
            TextField(
              controller: _replyController,
              decoration: const InputDecoration(
                labelText: 'Reply',
                hintText: 'Answer the workflow question',
              ),
              minLines: 2,
              maxLines: 6,
            ),
            FilledButton(
              onPressed: _cron == null
                  ? null
                  : () async {
                      final result = await _cron!.replyToExecution(
                        execution: exec,
                        message: _replyController.text,
                      );
                      if (!context.mounted) return;
                      if (result.error != null) {
                        ScaffoldMessenger.of(
                          context,
                        ).showSnackBar(SnackBar(content: Text(result.error!)));
                      } else if (result.unconfirmed) {
                        ScaffoldMessenger.of(context).showSnackBar(
                          const SnackBar(
                            content: Text(
                              'Reply not confirmed. Refresh before sending again.',
                            ),
                          ),
                        );
                      } else {
                        Navigator.of(context).pop();
                      }
                    },
              child: const Text('Send reply'),
            ),
            const Padding(
              padding: EdgeInsets.only(top: 8),
              child: Text(
                'Reply through the authorized session (ReplySession), not through Telegram or Chat threads.',
              ),
            ),
          ],
        ],
      ),
    );
  }
}

class CronEditorPage extends StatefulWidget {
  const CronEditorPage({super.key, required this.client, required this.draft});
  final ClientContext client;
  final CronJobDraft draft;

  @override
  State<CronEditorPage> createState() => _CronEditorPageState();
}

class _CronEditorPageState extends State<CronEditorPage> {
  CronController? get _cron => widget.client.cron;
  late CronJobDraft _draft;
  late final TextEditingController _nameController;
  late final TextEditingController _scheduleController;
  late final TextEditingController _inputController;
  late final TextEditingController _timezoneController;
  late final TextEditingController _timeoutController;
  late final TextEditingController _retryAttemptsController;
  late final TextEditingController _retryBackoffController;
  late final TextEditingController _maxOutputController;
  late final TextEditingController _webhookController;

  @override
  void initState() {
    super.initState();
    _draft = widget.draft;
    _nameController = TextEditingController(text: _draft.name);
    _scheduleController = TextEditingController(text: _draft.schedule);
    _inputController = TextEditingController(text: _draft.input);
    _timezoneController = TextEditingController(text: _draft.timezone);
    _timeoutController = TextEditingController(text: _draft.timeoutSeconds);
    _retryAttemptsController = TextEditingController(
      text: _draft.retryMaxAttempts,
    );
    _retryBackoffController = TextEditingController(
      text: _draft.retryBackoffSeconds,
    );
    _maxOutputController = TextEditingController(text: _draft.maxOutputBytes);
    _webhookController = TextEditingController(text: _draft.webhookUrl);
    _cron?.addListener(_onChange);
  }

  @override
  void dispose() {
    _cron?.removeListener(_onChange);
    _nameController.dispose();
    _scheduleController.dispose();
    _inputController.dispose();
    _timezoneController.dispose();
    _timeoutController.dispose();
    _retryAttemptsController.dispose();
    _retryBackoffController.dispose();
    _maxOutputController.dispose();
    _webhookController.dispose();
    super.dispose();
  }

  void _onChange() {
    if (mounted) setState(() {});
  }

  List<Agent> get _agents => widget.client.agents;

  Future<void> _save() async {
    final err = _draft.firstError();
    if (err != null) {
      ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text(err)));
      return;
    }
    CronJob job;
    try {
      job = _draft.build();
    } on FormatException catch (e) {
      ScaffoldMessenger.of(
        context,
      ).showSnackBar(SnackBar(content: Text(e.message)));
      return;
    }
    final cron = _cron;
    if (cron == null) return;
    final ok = _draft.isCreate
        ? await cron.create(job)
        : await cron.update(job);
    if (ok && mounted) Navigator.of(context).pop();
  }

  @override
  Widget build(BuildContext context) {
    final cron = _cron;
    return Scaffold(
      appBar: AppBar(
        title: Text(_draft.isCreate ? 'Add cron job' : 'Edit cron job'),
        actions: [
          TextButton(
            onPressed: cron?.writing == true ? null : _save,
            child: const Text('Save'),
          ),
        ],
      ),
      body: ListView(
        padding: const EdgeInsets.all(16),
        children: [
          if (cron?.writeError != null)
            Padding(
              padding: const EdgeInsets.only(bottom: 8),
              child: Text(cron!.writeError!),
            ),
          if (cron?.writeNotice != null)
            Padding(
              padding: const EdgeInsets.only(bottom: 8),
              child: Text(cron!.writeNotice!),
            ),
          TextField(
            decoration: const InputDecoration(labelText: 'Name'),
            enabled: _draft.isCreate,
            controller: _nameController,
            onChanged: (v) => _draft.name = v,
          ),
          TextField(
            decoration: const InputDecoration(
              labelText: 'Schedule',
              helperText: 'Cron expression or @daily, @hourly, @every 30m',
            ),
            controller: _scheduleController,
            onChanged: (v) => _draft.schedule = v,
          ),
          DropdownButtonFormField<String>(
            key: ValueKey('cron-agent-${_draft.agentId}-${_agents.length}'),
            decoration: const InputDecoration(labelText: 'Agent'),
            initialValue: _draft.agentId.isEmpty ? null : _draft.agentId,
            items: [
              for (final a in _agents)
                DropdownMenuItem(value: a.agentId, child: Text(a.name)),
            ],
            onChanged: (v) => setState(() => _draft.agentId = v ?? ''),
          ),
          TextField(
            decoration: const InputDecoration(labelText: 'Input message'),
            controller: _inputController,
            onChanged: (v) => _draft.input = v,
            maxLines: 3,
          ),
          TextField(
            decoration: const InputDecoration(
              labelText: 'Timezone',
              helperText: 'IANA name, e.g. Asia/Shanghai; empty = UTC',
            ),
            controller: _timezoneController,
            onChanged: (v) => _draft.timezone = v,
          ),
          SwitchListTile(
            title: const Text('Enabled'),
            value: _draft.enabled,
            onChanged: (v) => setState(() => _draft.enabled = v),
          ),
          Text('Reliability', style: Theme.of(context).textTheme.titleMedium),
          TextField(
            decoration: const InputDecoration(
              labelText: 'Timeout (seconds)',
              helperText: 'Empty = no cron-level timeout',
            ),
            keyboardType: TextInputType.number,
            controller: _timeoutController,
            onChanged: (v) => _draft.timeoutSeconds = v,
          ),
          TextField(
            decoration: const InputDecoration(labelText: 'Retry attempts'),
            keyboardType: TextInputType.number,
            controller: _retryAttemptsController,
            onChanged: (v) => _draft.retryMaxAttempts = v,
          ),
          TextField(
            decoration: const InputDecoration(
              labelText: 'Retry backoff (seconds)',
            ),
            keyboardType: TextInputType.number,
            controller: _retryBackoffController,
            onChanged: (v) => _draft.retryBackoffSeconds = v,
          ),
          DropdownButtonFormField<CronConcurrencyPolicy>(
            decoration: const InputDecoration(labelText: 'Concurrency'),
            initialValue: _draft.concurrencyPolicy,
            items: [
              for (final p in CronConcurrencyPolicy.values)
                if (p !=
                    CronConcurrencyPolicy.CRON_CONCURRENCY_POLICY_UNSPECIFIED)
                  DropdownMenuItem(
                    value: p,
                    child: Text(cronConcurrencyLabel(p)),
                  ),
            ],
            onChanged: (v) => setState(
              () => _draft.concurrencyPolicy =
                  v ?? CronConcurrencyPolicy.CRON_CONCURRENCY_POLICY_SKIP,
            ),
          ),
          DropdownButtonFormField<CronNotifyOn>(
            decoration: const InputDecoration(labelText: 'Notify on'),
            initialValue: _draft.notifyOn,
            items: [
              for (final n in CronNotifyOn.values)
                if (n != CronNotifyOn.CRON_NOTIFY_ON_UNSPECIFIED)
                  DropdownMenuItem(value: n, child: Text(cronNotifyOnLabel(n))),
            ],
            onChanged: (v) => setState(
              () => _draft.notifyOn = v ?? CronNotifyOn.CRON_NOTIFY_ON_FAILURE,
            ),
          ),
          DropdownButtonFormField<MaxOutputChoice>(
            decoration: const InputDecoration(labelText: 'Max output bytes'),
            initialValue: _draft.maxOutputChoice,
            items: const [
              DropdownMenuItem(
                value: MaxOutputChoice.unset,
                child: Text('Unset'),
              ),
              DropdownMenuItem(
                value: MaxOutputChoice.serverDefault,
                child: Text('Server default (0)'),
              ),
              DropdownMenuItem(
                value: MaxOutputChoice.limit,
                child: Text('Set limit'),
              ),
            ],
            onChanged: (v) => setState(
              () => _draft.maxOutputChoice = v ?? MaxOutputChoice.unset,
            ),
          ),
          if (_draft.maxOutputChoice == MaxOutputChoice.limit)
            TextField(
              decoration: const InputDecoration(labelText: 'Byte limit'),
              keyboardType: TextInputType.number,
              controller: _maxOutputController,
              onChanged: (v) => _draft.maxOutputBytes = v,
            ),
          Text('Delivery', style: Theme.of(context).textTheme.titleMedium),
          DropdownButtonFormField<CronDeliveryType>(
            initialValue: _draft.deliveryType,
            items: [
              for (final t in CronDeliveryType.values)
                if (t != CronDeliveryType.CRON_DELIVERY_TYPE_UNSPECIFIED &&
                    t != CronDeliveryType.CRON_DELIVERY_TYPE_CHANNEL)
                  DropdownMenuItem(value: t, child: Text(cronDeliveryLabel(t))),
            ],
            onChanged: (v) => setState(
              () => _draft.deliveryType =
                  v ?? CronDeliveryType.CRON_DELIVERY_TYPE_LOG,
            ),
          ),
          if (_draft.deliveryType ==
              CronDeliveryType.CRON_DELIVERY_TYPE_WEBHOOK)
            TextField(
              decoration: const InputDecoration(labelText: 'Webhook URL'),
              controller: _webhookController,
              onChanged: (v) => _draft.webhookUrl = v,
            ),
          if (_draft.deliveryType ==
              CronDeliveryType.CRON_DELIVERY_TYPE_NOTIFY_GROUP)
            DropdownButtonFormField<String>(
              key: ValueKey('cron-notify-${_draft.notifyGroupName}'),
              decoration: const InputDecoration(labelText: 'Notify Group'),
              initialValue: _draft.notifyGroupName.isEmpty
                  ? null
                  : _draft.notifyGroupName,
              items: [
                for (final g in cron?.notifyGroups ?? const [])
                  DropdownMenuItem(value: g.name, child: Text(g.name)),
              ],
              onChanged: (v) =>
                  setState(() => _draft.notifyGroupName = v ?? ''),
            ),
          if (_draft.deliveryType ==
              CronDeliveryType.CRON_DELIVERY_TYPE_TELEGRAM_DESTINATION)
            DropdownButtonFormField<String>(
              key: ValueKey('cron-tg-${_draft.telegramDestinationId}'),
              decoration: const InputDecoration(
                labelText: 'Telegram Destination',
              ),
              initialValue: _draft.telegramDestinationId.isEmpty
                  ? null
                  : _draft.telegramDestinationId,
              items: [
                for (final d in cron?.telegramDestinations ?? const [])
                  DropdownMenuItem(
                    value: d.id,
                    child: Text('${d.name} (${d.id})'),
                  ),
              ],
              onChanged: (v) =>
                  setState(() => _draft.telegramDestinationId = v ?? ''),
            ),
        ],
      ),
    );
  }
}
