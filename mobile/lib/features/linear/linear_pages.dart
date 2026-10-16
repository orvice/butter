import 'dart:async';

import 'package:flutter/material.dart';
import '../../app/client_context.dart';
import '../../core/management_page.dart';
import '../../gen/agents/v1/agent.pb.dart';
import '../../gen/agents/v1/linear.pb.dart';
import '../agents/leaf_draft.dart';
import 'linear_controller.dart';
import 'linear_draft.dart';
import 'linear_install_flow.dart';
import 'linear_secrets.dart';

Future<void> openLinearApps(BuildContext context, ClientContext client) async {
  final linear = client.linear;
  if (linear == null) return;
  unawaited(linear.loadApps());
  unawaited(linear.loadPlatformSettings());
  await openManagementPage(context, client, LinearListPage(client: client));
}

class LinearListPage extends StatefulWidget {
  const LinearListPage({super.key, required this.client});
  final ClientContext client;

  @override
  State<LinearListPage> createState() => _LinearListPageState();
}

class _LinearListPageState extends State<LinearListPage> {
  LinearController? get _linear => widget.client.linear;

  @override
  void initState() {
    super.initState();
    _linear?.addListener(_onChange);
  }

  @override
  void dispose() {
    _linear?.removeListener(_onChange);
    super.dispose();
  }

  void _onChange() {
    if (mounted) setState(() {});
  }

  @override
  Widget build(BuildContext context) {
    final linear = _linear;
    if (linear == null) {
      return const Scaffold(
        body: Center(child: Text('Sign in to a Workspace.')),
      );
    }
    return Scaffold(
      appBar: AppBar(
        title: const Text('Linear Apps'),
        actions: [
          IconButton(
            onPressed: linear.listState == LinearLoad.loading
                ? null
                : () {
                    unawaited(linear.loadApps());
                    unawaited(linear.loadPlatformSettings());
                  },
            icon: const Icon(Icons.refresh),
            tooltip: 'Refresh',
          ),
        ],
      ),
      floatingActionButton: widget.client.canManageWorkspaceResources
          ? FloatingActionButton.extended(
              onPressed: linear.writing || linear.unconfirmed
                  ? null
                  : () => Navigator.push<void>(
                      context,
                      MaterialPageRoute(
                        builder: (_) => LinearEditorPage(
                          client: widget.client,
                          draft: LinearAppDraft.create(),
                          creating: true,
                        ),
                      ),
                    ),
              icon: const Icon(Icons.add),
              label: const Text('Register App'),
            )
          : null,
      body: ListView(
        padding: const EdgeInsets.all(16),
        children: [
          _PlatformUrlBanner(linear: linear),
          if (linear.listState == LinearLoad.loading)
            const LinearProgressIndicator(semanticsLabel: 'Loading'),
          if (linear.listError != null)
            Text(
              linear.listError!,
              style: TextStyle(color: Theme.of(context).colorScheme.error),
            ),
          for (final app in linear.apps)
            ListTile(
              title: Text(app.displayName.isEmpty ? app.id : app.displayName),
              subtitle: Text(
                'Agent ${app.agentId} · inbound ${app.inboundEnabled ? 'on' : 'off'}',
              ),
              onTap: () async {
                await Navigator.push<void>(
                  context,
                  MaterialPageRoute(
                    builder: (_) =>
                        LinearDetailPage(client: widget.client, appId: app.id),
                  ),
                );
              },
            ),
          ListTile(
            leading: const Icon(Icons.history),
            title: const Text('Processing deliveries'),
            onTap: () => Navigator.push<void>(
              context,
              MaterialPageRoute(
                builder: (_) => LinearProcessingListPage(client: widget.client),
              ),
            ),
          ),
        ],
      ),
    );
  }
}

class _PlatformUrlBanner extends StatelessWidget {
  const _PlatformUrlBanner({required this.linear});
  final LinearController linear;

  @override
  Widget build(BuildContext context) {
    if (linear.platformState != LinearLoad.ready) {
      return const SizedBox.shrink();
    }
    if (linear.platformSettingsDenied ||
        (linear.platformBaseUrl ?? '').isEmpty) {
      return Card(
        child: Padding(
          padding: const EdgeInsets.all(12),
          child: Text(
            linear.platformSettingsDenied
                ? 'The public Linear base URL is configured by a platform administrator. '
                      'Until it is set, callback and webhook URLs on each App stay empty and inbound cannot be enabled.'
                : 'The Linear public base URL is not configured yet. Ask a platform administrator to set it in the Web console.',
          ),
        ),
      );
    }
    return Card(
      child: ListTile(
        title: const Text('Public base URL (read-only)'),
        subtitle: Text(linear.platformBaseUrl!),
      ),
    );
  }
}

class LinearDetailPage extends StatefulWidget {
  const LinearDetailPage({
    super.key,
    required this.client,
    required this.appId,
  });
  final ClientContext client;
  final String appId;

  @override
  State<LinearDetailPage> createState() => _LinearDetailPageState();
}

class _LinearDetailPageState extends State<LinearDetailPage> {
  LinearController? get _linear => widget.client.linear;

  @override
  void initState() {
    super.initState();
    _linear?.addListener(_onChange);
    unawaited(_linear?.loadApp(widget.appId));
    unawaited(_linear?.loadInstallations(widget.appId));
  }

  @override
  void dispose() {
    _linear?.removeListener(_onChange);
    super.dispose();
  }

  void _onChange() {
    if (mounted) setState(() {});
  }

  Future<void> _confirmDelete(LinearApp app) async {
    final ok = await showDialog<bool>(
      context: context,
      builder: (ctx) => AlertDialog(
        title: const Text('Delete Linear App?'),
        content: Text(
          'Delete "${app.displayName}" and remove every installation? This cannot be undone.',
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
    if (ok != true || !mounted) return;
    if (await _linear!.deleteApp(app.id) && mounted) Navigator.pop(context);
  }

  @override
  Widget build(BuildContext context) {
    final linear = _linear;
    final app = linear?.selected;
    if (linear == null) {
      return const Scaffold(
        body: Center(child: Text('Sign in to a Workspace.')),
      );
    }
    return Scaffold(
      appBar: AppBar(
        title: Text(
          app?.displayName.isEmpty ?? true ? 'Linear App' : app!.displayName,
        ),
        actions: [
          if (widget.client.canManageWorkspaceResources && app != null)
            IconButton(
              onPressed: linear.writing
                  ? null
                  : () => Navigator.push<void>(
                      context,
                      MaterialPageRoute(
                        builder: (_) => LinearEditorPage(
                          client: widget.client,
                          draft: LinearAppDraft.fromStored(app),
                          creating: false,
                        ),
                      ),
                    ),
              icon: const Icon(Icons.edit),
              tooltip: 'Edit',
            ),
        ],
      ),
      body: ListView(
        padding: const EdgeInsets.all(16),
        children: [
          if (linear.detailState == LinearLoad.loading)
            const LinearProgressIndicator(semanticsLabel: 'Loading'),
          if (linear.detailError != null)
            Text(
              linear.detailError!,
              style: TextStyle(color: Theme.of(context).colorScheme.error),
            ),
          if (app != null) ...[
            Text('Client ID: ${app.clientId}'),
            Text('Agent: ${app.agentId}'),
            Text('Revision: ${app.revision}'),
            Text(
              'Credentials: ${app.credentialState.name.replaceAll('_', ' ')}',
            ),
            if (app.callbackUrl.isNotEmpty)
              SelectableText('Callback: ${app.callbackUrl}'),
            if (app.webhookUrl.isNotEmpty)
              SelectableText('Webhook: ${app.webhookUrl}'),
            if (app.callbackUrl.isEmpty)
              const Text(
                'Callback URL unavailable until a platform admin configures the public base URL.',
              ),
            LinearSecretsCard(
              linear: linear,
              appId: app.id,
              clientSecretSet: app.clientSecretSet,
              webhookSecretSet: app.webhookSecretSet,
              inboundEnabled: app.inboundEnabled,
              canManage: widget.client.canManageWorkspaceResources,
            ),
            SwitchListTile(
              title: const Text('Inbound enabled'),
              value: app.inboundEnabled,
              onChanged: null,
            ),
            if (widget.client.canManageWorkspaceResources) ...[
              FilledButton.tonal(
                onPressed: linear.install.phase == LinearInstallPhase.waiting
                    ? null
                    : () => linear.beginInstall(app.id),
                child: const Text('Install to organization'),
              ),
              if (linear.install.appId == app.id &&
                  linear.install.phase != LinearInstallPhase.idle)
                _InstallStatus(linear: linear),
              OutlinedButton(
                onPressed: linear.writing ? null : () => _confirmDelete(app),
                child: const Text('Delete App'),
              ),
            ],
            const SizedBox(height: 16),
            Text(
              'Installations',
              style: Theme.of(context).textTheme.titleMedium,
            ),
            if (linear.installListState == LinearLoad.loading)
              const LinearProgressIndicator(
                semanticsLabel: 'Loading installations',
              ),
            for (final inst in linear.installations)
              _InstallationTile(linear: linear, inst: inst, appId: app.id),
          ],
        ],
      ),
    );
  }
}

class _InstallStatus extends StatelessWidget {
  const _InstallStatus({required this.linear});
  final LinearController linear;

  @override
  Widget build(BuildContext context) {
    final view = linear.install;
    return Card(
      child: Padding(
        padding: const EdgeInsets.all(12),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(view.phase.name),
            if (view.message != null) Text(view.message!),
            if (view.phase == LinearInstallPhase.waiting)
              TextButton(
                onPressed: linear.stopWaitingForInstall,
                child: const Text('Stop waiting'),
              ),
          ],
        ),
      ),
    );
  }
}

class _InstallationTile extends StatelessWidget {
  const _InstallationTile({
    required this.linear,
    required this.inst,
    required this.appId,
  });
  final LinearController linear;
  final LinearInstallation inst;
  final String appId;

  @override
  Widget build(BuildContext context) {
    final needsReinstall =
        inst.credentialState ==
        LinearInstallationCredentialState
            .LINEAR_INSTALLATION_CREDENTIAL_STATE_NEEDS_REINSTALL;
    return ListTile(
      title: Text(
        inst.organizationName.isEmpty
            ? inst.organizationId
            : inst.organizationName,
      ),
      subtitle: Text(
        needsReinstall
            ? 'Needs reinstall${inst.lastCredentialError.isNotEmpty ? ': ${inst.lastCredentialError}' : ''}'
            : inst.credentialState.name.replaceAll('_', ' '),
      ),
      trailing: linear.writing
          ? null
          : IconButton(
              icon: const Icon(Icons.delete_outline),
              tooltip: 'Remove installation',
              onPressed: () async {
                final ok = await showDialog<bool>(
                  context: context,
                  builder: (ctx) => AlertDialog(
                    title: const Text('Remove installation?'),
                    content: Text(
                      'Remove ${inst.organizationName} from this App?',
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
                if (ok == true) await linear.deleteInstallation(appId, inst.id);
              },
            ),
    );
  }
}

class LinearEditorPage extends StatefulWidget {
  const LinearEditorPage({
    super.key,
    required this.client,
    required this.draft,
    required this.creating,
  });
  final ClientContext client;
  final LinearAppDraft draft;
  final bool creating;

  @override
  State<LinearEditorPage> createState() => _LinearEditorPageState();
}

class _LinearEditorPageState extends State<LinearEditorPage> {
  late LinearAppDraft _draft;
  String? _validation;
  RunLimitMode _runMode = RunLimitMode.serverDefault;
  late final TextEditingController _nameCtrl;
  late final TextEditingController _clientCtrl;
  late final TextEditingController _allowCtrl;
  late final TextEditingController _runSecondsCtrl;
  late final TextEditingController _clientSecretCtrl;
  late final TextEditingController _webhookSecretCtrl;

  @override
  void initState() {
    super.initState();
    _draft = widget.draft;
    _runMode = _draft.runLimitMode;
    _nameCtrl = TextEditingController(text: _draft.displayName);
    _clientCtrl = TextEditingController(text: _draft.clientId);
    _allowCtrl = TextEditingController(
      text: LinearAppDraft.formatAllowedUserIds(_draft.allowedUserIds),
    );
    _runSecondsCtrl = TextEditingController(
      text: _runMode == RunLimitMode.seconds
          ? '${_draft.stored.maxRunSeconds}'
          : '',
    );
    _clientSecretCtrl = TextEditingController();
    _webhookSecretCtrl = TextEditingController();
  }

  @override
  void dispose() {
    _nameCtrl.dispose();
    _clientCtrl.dispose();
    _allowCtrl.dispose();
    _runSecondsCtrl.dispose();
    _clientSecretCtrl.dispose();
    _webhookSecretCtrl.dispose();
    super.dispose();
  }

  List<Agent> get _agents =>
      widget.client.agents.where((a) => a.agentId.isNotEmpty).toList();

  Future<void> _save() async {
    _draft.displayName = _nameCtrl.text;
    _draft.clientId = _clientCtrl.text;
    _draft.allowedUserIds = LinearAppDraft.parseAllowedUserIds(_allowCtrl.text);
    _validation = _draft.validate(creating: widget.creating);
    if (_runMode == RunLimitMode.seconds) {
      final parsed = parseRunSeconds(_runSecondsCtrl.text);
      if (parsed == null) {
        _validation = validateRunSeconds(_runSecondsCtrl.text);
      } else {
        _draft.runLimitMode = RunLimitMode.seconds;
        _draft.setRunLimitSeconds(parsed);
      }
    } else {
      _draft.runLimitMode = _runMode;
    }
    setState(() {});
    if (_validation != null) return;
    final linear = widget.client.linear!;
    final ok = widget.creating
        ? await linear.createApp(_draft)
        : await linear.updateApp(_draft);
    if (ok && mounted) Navigator.pop(context);
  }

  @override
  Widget build(BuildContext context) {
    final linear = widget.client.linear;
    final canManage = widget.client.canManageWorkspaceResources;
    return Scaffold(
      appBar: AppBar(
        title: Text(
          widget.creating ? 'Register Linear App' : 'Edit Linear App',
        ),
        actions: [
          TextButton(
            onPressed: !canManage || linear?.writing == true ? null : _save,
            child: const Text('Save'),
          ),
        ],
      ),
      body: ListView(
        padding: const EdgeInsets.all(16),
        children: [
          if (_validation != null)
            Text(
              _validation!,
              style: TextStyle(color: Theme.of(context).colorScheme.error),
            ),
          if (linear?.writeError != null)
            Text(
              linear!.writeError!,
              style: TextStyle(color: Theme.of(context).colorScheme.error),
            ),
          TextField(
            decoration: const InputDecoration(labelText: 'Display name'),
            onChanged: (v) => _draft.displayName = v,
            controller: _nameCtrl,
          ),
          TextField(
            decoration: const InputDecoration(labelText: 'Client ID'),
            enabled: widget.creating,
            onChanged: (v) => _draft.clientId = v,
            controller: _clientCtrl,
          ),
          DropdownButtonFormField<String>(
            initialValue: _draft.agentId.isEmpty ? null : _draft.agentId,
            decoration: const InputDecoration(labelText: 'Routed Agent'),
            items: [
              for (final a in _agents)
                DropdownMenuItem(value: a.agentId, child: Text(a.name)),
            ],
            onChanged: canManage
                ? (v) => setState(() => _draft.agentId = v ?? '')
                : null,
          ),
          SwitchListTile(
            title: const Text('Inbound enabled'),
            value: _draft.inboundEnabled,
            onChanged: canManage
                ? (v) => setState(() => _draft.inboundEnabled = v)
                : null,
          ),
          TextField(
            decoration: const InputDecoration(
              labelText: 'Allowed Linear user IDs (one per line, empty = all)',
            ),
            maxLines: 3,
            onChanged: (v) =>
                _draft.allowedUserIds = LinearAppDraft.parseAllowedUserIds(v),
            controller: _allowCtrl,
          ),
          DropdownButtonFormField<RunLimitMode>(
            initialValue: _runMode,
            decoration: const InputDecoration(labelText: 'Max run time'),
            items: const [
              DropdownMenuItem(
                value: RunLimitMode.serverDefault,
                child: Text('Server default (1800 s)'),
              ),
              DropdownMenuItem(
                value: RunLimitMode.unlimited,
                child: Text('Unlimited (0)'),
              ),
              DropdownMenuItem(
                value: RunLimitMode.seconds,
                child: Text('Set seconds'),
              ),
            ],
            onChanged: canManage
                ? (v) => setState(() => _runMode = v ?? _runMode)
                : null,
          ),
          if (_runMode == RunLimitMode.seconds)
            TextField(
              decoration: const InputDecoration(labelText: 'Seconds'),
              keyboardType: TextInputType.number,
              controller: _runSecondsCtrl,
            ),
          if (widget.creating) ...[
            TextField(
              decoration: const InputDecoration(
                labelText: 'Client secret (write-only)',
              ),
              obscureText: true,
              onChanged: (v) => _draft.clientSecret = v,
              controller: _clientSecretCtrl,
            ),
            TextField(
              decoration: const InputDecoration(
                labelText: 'Webhook secret (write-only)',
              ),
              obscureText: true,
              onChanged: (v) => _draft.webhookSecret = v,
              controller: _webhookSecretCtrl,
            ),
          ],
        ],
      ),
    );
  }
}

class LinearProcessingListPage extends StatefulWidget {
  const LinearProcessingListPage({super.key, required this.client});
  final ClientContext client;

  @override
  State<LinearProcessingListPage> createState() =>
      _LinearProcessingListPageState();
}

class _LinearProcessingListPageState extends State<LinearProcessingListPage> {
  @override
  void initState() {
    super.initState();
    unawaited(widget.client.linear?.loadRecords());
  }

  @override
  Widget build(BuildContext context) {
    final linear = widget.client.linear!;
    return Scaffold(
      appBar: AppBar(title: const Text('Linear deliveries')),
      body: ListView(
        children: [
          if (linear.recordsState == LinearLoad.loading)
            const LinearProgressIndicator(semanticsLabel: 'Loading'),
          for (final r in linear.records)
            ListTile(
              title: Text('${r.action} · ${r.status.name}'),
              subtitle: Text(
                r.issueIdentifier.isEmpty ? r.id : r.issueIdentifier,
              ),
              onTap: () => Navigator.push<void>(
                context,
                MaterialPageRoute(
                  builder: (_) => LinearProcessingDetailPage(
                    client: widget.client,
                    recordId: r.id,
                  ),
                ),
              ),
            ),
        ],
      ),
    );
  }
}

class LinearProcessingDetailPage extends StatefulWidget {
  const LinearProcessingDetailPage({
    super.key,
    required this.client,
    required this.recordId,
  });
  final ClientContext client;
  final String recordId;

  @override
  State<LinearProcessingDetailPage> createState() =>
      _LinearProcessingDetailPageState();
}

class _LinearProcessingDetailPageState
    extends State<LinearProcessingDetailPage> {
  @override
  void initState() {
    super.initState();
    unawaited(widget.client.linear?.loadRecord(widget.recordId));
  }

  @override
  Widget build(BuildContext context) {
    final linear = widget.client.linear!;
    final r = linear.selectedRecord;
    final canResend =
        widget.client.canManageWorkspaceResources &&
        r?.status == LinearProcessingStatus.LINEAR_PROCESSING_STATUS_FAILED;
    return Scaffold(
      appBar: AppBar(title: const Text('Delivery detail')),
      body: r == null
          ? const Center(child: CircularProgressIndicator())
          : ListView(
              padding: const EdgeInsets.all(16),
              children: [
                Text('Status: ${r.status.name}'),
                if (r.error.isNotEmpty) Text('Error: ${r.error}'),
                if (r.output.isNotEmpty) SelectableText('Reply: ${r.output}'),
                if (canResend)
                  FilledButton(
                    onPressed: linear.writing
                        ? null
                        : () async {
                            final ok = await showDialog<bool>(
                              context: context,
                              builder: (ctx) => AlertDialog(
                                title: const Text('Resend persisted reply?'),
                                content: const Text(
                                  'Posts the saved reply again. The Agent is not run.',
                                ),
                                actions: [
                                  TextButton(
                                    onPressed: () => Navigator.pop(ctx, false),
                                    child: const Text('Cancel'),
                                  ),
                                  FilledButton(
                                    onPressed: () => Navigator.pop(ctx, true),
                                    child: const Text('Resend'),
                                  ),
                                ],
                              ),
                            );
                            if (ok == true) await linear.resendReply(r.id);
                          },
                    child: const Text('Resend reply'),
                  ),
              ],
            ),
    );
  }
}
