import 'dart:async';

import 'package:flutter/material.dart';
import 'package:flutter/services.dart';

import '../../app/client_context.dart';
import '../../core/config_resource_controller.dart';
import '../../core/config_widgets.dart';
import '../../core/daemon_errors.dart';
import '../../core/management_page.dart';
import '../../gen/agents/v1/api_token.pb.dart';
import '../../gen/agents/v1/dashboard.pb.dart';
import 'daemon_controller.dart';
import 'daemon_draft.dart';

/// Opens the Workspace's daemon runtimes. Everything reads fresh.
Future<void> openDaemons(BuildContext context, ClientContext client) async {
  final daemons = client.daemons;
  if (daemons == null) return;
  unawaited(daemons.load());
  unawaited(daemons.loadOnline());
  await openManagementPage(
    context,
    client,
    DaemonRuntimeListPage(client: client),
  );
}

/// A string the worker reported about itself: shown as plain text, bounded.
String workerText(String value, {int max = 80}) {
  final clean = value.replaceAll(RegExp(r'[\u0000-\u001f]'), ' ').trim();
  return clean.length <= max ? clean : '${clean.substring(0, max)}…';
}

String elapsedLabel(int seconds) {
  if (seconds < 60) return '${seconds}s';
  final m = seconds ~/ 60, s = seconds % 60;
  if (m < 60) return '${m}m ${s}s';
  return '${m ~/ 60}h ${m % 60}m';
}

String workerSummary(DaemonStatus? w) {
  if (w == null) return 'Offline: no worker is connected';
  final parts = [
    'Online',
    if (workerText(w.name).isNotEmpty) workerText(w.name),
    '${w.activeTasks} active ${w.activeTasks == 1 ? 'task' : 'tasks'}',
  ];
  return parts.join(' · ');
}

Widget _loadError(BuildContext context, String? text, Key key) => text == null
    ? const SizedBox.shrink()
    : Semantics(
        liveRegion: true,
        child: Text(
          text,
          key: key,
          style: TextStyle(color: Theme.of(context).colorScheme.error),
        ),
      );

// ---------------------------------------------------------------- runtimes

class DaemonRuntimeListPage extends StatelessWidget {
  const DaemonRuntimeListPage({super.key, required this.client});
  final ClientContext client;

  @override
  Widget build(BuildContext context) {
    final d = client.daemons;
    if (d == null) return const SizedBox.shrink();
    return ConfigScaffold(
      controller: d,
      title: 'Daemon execution',
      child: ListenableBuilder(
        listenable: d,
        builder: (context, _) => ListView(
          padding: const EdgeInsets.all(16),
          children: [
            Text(
              '${client.workspace?.name ?? 'Workspace'} · Daemon runtimes',
              style: Theme.of(context).textTheme.headlineSmall,
            ),
            const SizedBox(height: 8),
            const Text(
              'A daemon runtime is a place a worker connects to run agent tasks. The worker runs on your own machine, never on this phone. Create a runtime, mint a worker token for it, start the worker with that token, then bind a Remote Agent to the runtime.',
            ),
            const SizedBox(height: 12),
            Wrap(
              spacing: 12,
              children: [
                FilledButton.icon(
                  key: const Key('add-daemon'),
                  onPressed: d.writing
                      ? null
                      : () => openManagementPage(
                          context,
                          client,
                          DaemonFormPage(client: client),
                        ),
                  icon: const Icon(Icons.add),
                  label: const Text('Add runtime'),
                ),
                OutlinedButton.icon(
                  key: const Key('open-daemon-tasks'),
                  onPressed: () => openManagementPage(
                    context,
                    client,
                    DaemonTasksPage(client: client),
                  ),
                  icon: const Icon(Icons.list_alt),
                  label: const Text('Tasks'),
                ),
                OutlinedButton.icon(
                  key: const Key('open-daemon-diagnostics'),
                  onPressed: () => openManagementPage(
                    context,
                    client,
                    BridgeDiagnosticsPage(client: client),
                  ),
                  icon: const Icon(Icons.monitor_heart_outlined),
                  label: const Text('Diagnostics'),
                ),
                TextButton.icon(
                  key: const Key('refresh-daemons'),
                  onPressed: d.listState == ConfigLoad.loading
                      ? null
                      : () {
                          unawaited(d.load());
                          unawaited(d.loadOnline());
                        },
                  icon: const Icon(Icons.refresh),
                  label: const Text('Refresh'),
                ),
              ],
            ),
            if (d.listState == ConfigLoad.loading)
              const LinearProgressIndicator(semanticsLabel: 'Loading'),
            _loadError(context, d.listError, const Key('daemon-list-error')),
            _loadError(
              context,
              d.onlineError,
              const Key('daemon-online-error'),
            ),
            if (d.listState == ConfigLoad.ready && d.items.isEmpty) ...[
              const SizedBox(height: 24),
              Text(
                'No daemon runtimes yet',
                style: Theme.of(context).textTheme.titleMedium,
              ),
              const Text('Add one to connect a worker.'),
            ],
            for (final r in d.items)
              ListTile(
                key: Key('daemon-${r.id}'),
                isThreeLine: true,
                contentPadding: EdgeInsets.zero,
                title: Text(r.name.isEmpty ? r.id : r.name),
                subtitle: Text(
                  [
                    r.id,
                    if (d.onlineState == ConfigLoad.ready)
                      workerSummary(d.workerOf(r.id))
                    else
                      'Worker status not read',
                  ].join('\n'),
                ),
                onTap: () => openManagementPage(
                  context,
                  client,
                  DaemonRuntimeDetailPage(client: client, id: r.id),
                ),
              ),
          ],
        ),
      ),
    );
  }
}

class DaemonRuntimeDetailPage extends StatefulWidget {
  const DaemonRuntimeDetailPage({
    super.key,
    required this.client,
    required this.id,
  });
  final ClientContext client;
  final String id;
  @override
  State<DaemonRuntimeDetailPage> createState() =>
      _DaemonRuntimeDetailPageState();
}

class _DaemonRuntimeDetailPageState extends State<DaemonRuntimeDetailPage> {
  @override
  void initState() {
    super.initState();
    final d = widget.client.daemons;
    if (d != null) {
      unawaited(d.loadItem(widget.id));
      unawaited(d.loadOnline());
      unawaited(d.loadTokens());
    }
  }

  @override
  void dispose() {
    // Leaving this screen discards a secret shown on it and fences a mint
    // still in flight: nothing is left to show it on.
    widget.client.daemons?.leaveSecretScreen(widget.id, notify: false);
    super.dispose();
  }

  Future<void> _delete(DaemonController d, DaemonRuntime runtime) async {
    final users = await d.remoteAgentsOn(runtime.id);
    if (!mounted) return;
    final label = runtime.name.isEmpty ? runtime.id : runtime.name;
    final using = users == null
        ? 'Butter could not check which Remote Agents use it.'
        : users.isEmpty
        ? 'No Remote Agent in this Workspace uses it.'
        : 'These Remote Agents run on it and will fail: ${users.map((a) => a.name.isEmpty ? a.id : a.name).join(', ')}.';
    final ok = await confirmChange(
      context,
      'Delete daemon runtime?',
      'Delete "$label"? $using Its worker tokens stop working but stay in the list until you revoke them, and a worker connected for it can no longer run tasks.',
      action: 'Delete runtime',
    );
    if (!ok || !mounted) return;
    if (await d.delete(runtime.id) && mounted && Navigator.canPop(context)) {
      Navigator.pop(context);
    }
  }

  Future<void> _mint(DaemonController d, DaemonRuntime runtime) async {
    final choice = await showDialog<MintChoice>(
      context: context,
      builder: (context) => _MintDialog(runtime: runtime),
    );
    if (choice == null || !mounted) return;
    await d.mint(runtime.id, choice.name, ttl: choice.ttl);
  }

  Future<void> _revoke(DaemonController d, APIToken token) async {
    final ok = await confirmChange(
      context,
      'Revoke worker token?',
      'Revoke "${token.prefix}…"? A worker using it can no longer connect. This cannot be undone.',
      action: 'Revoke',
    );
    if (!ok || !mounted) return;
    await d.revoke(token.id);
  }

  @override
  Widget build(BuildContext context) {
    final d = widget.client.daemons;
    if (d == null) return const SizedBox.shrink();
    return ConfigScaffold(
      controller: d,
      title: 'Daemon runtime',
      child: ListenableBuilder(
        listenable: d,
        builder: (context, _) {
          final r = d.item?.id == widget.id ? d.item : null;
          if (r == null) {
            return ListView(
              padding: const EdgeInsets.all(16),
              children: [
                if (d.itemState == ConfigLoad.loading)
                  const LinearProgressIndicator(semanticsLabel: 'Loading'),
                _loadError(
                  context,
                  d.itemError,
                  const Key('daemon-detail-error'),
                ),
                TextButton(
                  onPressed: d.itemState == ConfigLoad.loading
                      ? null
                      : () => d.loadItem(widget.id),
                  child: const Text('Load runtime'),
                ),
              ],
            );
          }
          final worker = d.workerOf(r.id);
          final busy = d.writing;
          final tokens = d.tokensOf(r.id);
          // Another screen on top of this one would cover the secret, so
          // nothing that opens one is offered while it is shown.
          final showingSecret = d.minted?.runtimeId == r.id;
          return ListView(
            padding: const EdgeInsets.all(16),
            children: [
              if (d.minted?.runtimeId == r.id)
                _SecretPanel(controller: d, minted: d.minted!),
              SelectableText(
                r.name.isEmpty ? r.id : r.name,
                style: Theme.of(context).textTheme.headlineSmall,
              ),
              SelectableText('ID: ${r.id}'),
              if (r.description.isNotEmpty) SelectableText(r.description),
              if (r.labels.isNotEmpty)
                Text(
                  'Labels: ${(r.labels.entries.toList()..sort((a, b) => a.key.compareTo(b.key))).map((e) => '${e.key}=${e.value}').join(', ')}',
                ),
              const SizedBox(height: 12),
              Text('Worker', style: Theme.of(context).textTheme.titleMedium),
              if (d.onlineState == ConfigLoad.loading && d.online.isEmpty)
                const LinearProgressIndicator(semanticsLabel: 'Loading'),
              _loadError(
                context,
                d.onlineError,
                const Key('daemon-online-error'),
              ),
              Text(
                d.onlineState == ConfigLoad.ready
                    ? workerSummary(worker)
                    : 'Worker status not read',
                key: const Key('daemon-worker-summary'),
              ),
              if (worker != null) ..._workerDetails(worker),
              const SizedBox(height: 12),
              Text(
                'Worker tokens',
                style: Theme.of(context).textTheme.titleMedium,
              ),
              const Text(
                'A worker token only lets a worker connect to this runtime. It is not an API token for you.',
              ),
              if (d.tokenState == ConfigLoad.loading)
                const LinearProgressIndicator(semanticsLabel: 'Loading'),
              _loadError(
                context,
                d.tokenError,
                const Key('daemon-token-error'),
              ),
              if (d.tokenState == ConfigLoad.ready && tokens.isEmpty)
                const Text('No worker token for this runtime.'),
              for (final t in tokens)
                ListTile(
                  key: Key('token-${t.id}'),
                  contentPadding: EdgeInsets.zero,
                  title: Text('${t.prefix}… · ${t.name}'),
                  subtitle: Text(_tokenLine(t)),
                  trailing: t.revoked
                      ? null
                      : TextButton(
                          key: Key('revoke-${t.id}'),
                          onPressed: busy ? null : () => _revoke(d, t),
                          child: const Text('Revoke'),
                        ),
                ),
              const SizedBox(height: 12),
              if (showingSecret)
                const Text(
                  'Save the secret above before editing or deleting this runtime.',
                  key: Key('secret-blocks-navigation'),
                ),
              Wrap(
                spacing: 12,
                children: [
                  FilledButton.icon(
                    key: const Key('mint-token'),
                    onPressed: busy || d.unconfirmed || d.minted != null
                        ? null
                        : () => _mint(d, r),
                    icon: const Icon(Icons.vpn_key_outlined),
                    label: const Text('Create worker token'),
                  ),
                  OutlinedButton.icon(
                    key: const Key('edit-daemon'),
                    onPressed: busy || showingSecret
                        ? null
                        : () => openManagementPage(
                            context,
                            widget.client,
                            DaemonFormPage(client: widget.client, base: r),
                          ),
                    icon: const Icon(Icons.edit),
                    label: const Text('Edit'),
                  ),
                  TextButton.icon(
                    key: const Key('refresh-daemon-item'),
                    onPressed: d.itemState == ConfigLoad.loading
                        ? null
                        : () {
                            unawaited(d.loadItem(r.id));
                            unawaited(d.loadOnline());
                            unawaited(d.loadTokens());
                          },
                    icon: const Icon(Icons.refresh),
                    label: const Text('Refresh'),
                  ),
                  TextButton.icon(
                    key: const Key('delete-daemon'),
                    onPressed: busy || d.unconfirmed || showingSecret
                        ? null
                        : () => _delete(d, r),
                    icon: const Icon(Icons.delete_outline),
                    label: const Text('Delete'),
                  ),
                ],
              ),
            ],
          );
        },
      ),
    );
  }

  List<Widget> _workerDetails(DaemonStatus w) => [
    if (workerText(w.version).isNotEmpty || workerText(w.os).isNotEmpty)
      Text(
        [
          if (workerText(w.version).isNotEmpty)
            'Version ${workerText(w.version)}',
          if (workerText(w.os).isNotEmpty) workerText(w.os),
        ].join(' · '),
      ),
    if (w.executors.isNotEmpty)
      Text(
        'Executors on the worker: ${w.executors.map(workerText).join(', ')}',
        key: const Key('daemon-executors'),
      ),
    if (w.acpRuntimes.isNotEmpty)
      Text(
        'ACP runtimes it supports: ${w.acpRuntimes.map(workerText).join(', ')}',
      ),
    if (w.hasConnectedAt())
      Text('Connected since ${w.connectedAt.toDateTime().toLocal()}'),
    if (w.labels.isNotEmpty)
      Text(
        'Worker labels: ${(w.labels.entries.toList()..sort((a, b) => a.key.compareTo(b.key))).map((e) => '${workerText(e.key, max: 40)}=${workerText(e.value, max: 40)}').join(', ')}',
      ),
  ];

  String _tokenLine(APIToken t) => [
    if (t.revoked) 'Revoked',
    if (t.hasCreatedAt()) 'Created ${t.createdAt.toDateTime().toLocal()}',
    if (t.hasLastUsedAt()) 'Last used ${t.lastUsedAt.toDateTime().toLocal()}',
    if (t.hasExpiresAt()) 'Expires ${t.expiresAt.toDateTime().toLocal()}',
    if (!t.hasExpiresAt()) 'Does not expire',
  ].join(' · ');
}

/// What the person chose in the worker-token dialog.
class MintChoice {
  const MintChoice(this.name, this.ttl);
  final String name;

  /// Null: the token does not expire until it is revoked.
  final Duration? ttl;
}

/// Days until a worker token expires: blank means it never does. Null error
/// means the text is usable.
({Duration? ttl, String? error}) parseExpiryDays(String text) {
  final t = text.trim();
  if (t.isEmpty) return (ttl: null, error: null);
  final days = int.tryParse(t);
  if (days == null || days < 1 || days > 3650) {
    return (
      ttl: null,
      error: 'Enter a whole number of days from 1 to 3650, or leave it empty.',
    );
  }
  return (ttl: Duration(days: days), error: null);
}

class _MintDialog extends StatefulWidget {
  const _MintDialog({required this.runtime});
  final DaemonRuntime runtime;
  @override
  State<_MintDialog> createState() => _MintDialogState();
}

class _MintDialogState extends State<_MintDialog> {
  late final TextEditingController _name = TextEditingController(
    text: widget.runtime.name.isEmpty
        ? 'Worker for ${widget.runtime.id}'
        : 'Worker for ${widget.runtime.name}',
  );
  final _days = TextEditingController();
  String? _nameError;
  String? _daysError;

  @override
  void dispose() {
    _name.dispose();
    _days.dispose();
    super.dispose();
  }

  void _submit() {
    final name = _name.text.trim();
    final expiry = parseExpiryDays(_days.text);
    setState(() {
      _nameError = name.isEmpty ? 'Enter a name.' : null;
      _daysError = expiry.error;
    });
    if (name.isEmpty || expiry.error != null) return;
    Navigator.pop(context, MintChoice(name, expiry.ttl));
  }

  @override
  Widget build(BuildContext context) => AlertDialog(
    title: const Text('Create a worker token?'),
    content: SingleChildScrollView(
      child: Column(
        mainAxisSize: MainAxisSize.min,
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(
            'This creates one token that lets a worker connect to "${widget.runtime.id}" and nothing else. Its secret is shown once. It is separate from your personal API tokens.',
          ),
          TextField(
            key: const Key('mint-name'),
            controller: _name,
            decoration: InputDecoration(
              labelText: 'Token name',
              errorText: _nameError,
            ),
          ),
          TextField(
            key: const Key('mint-expiry-days'),
            controller: _days,
            keyboardType: TextInputType.number,
            decoration: InputDecoration(
              labelText: 'Expires after (days)',
              helperText:
                  'Empty: the token does not expire until you revoke it.',
              helperMaxLines: 2,
              errorText: _daysError,
            ),
          ),
        ],
      ),
    ),
    actions: [
      TextButton(
        onPressed: () => Navigator.pop(context),
        child: const Text('Cancel'),
      ),
      FilledButton(onPressed: _submit, child: const Text('Create token')),
    ],
  );
}

/// The one-time secret of a new worker token. It is on screen only here.
class _SecretPanel extends StatelessWidget {
  const _SecretPanel({required this.controller, required this.minted});
  final DaemonController controller;
  final MintedSecret minted;

  @override
  Widget build(BuildContext context) {
    final secret = minted.secret;
    return Card(
      key: const Key('minted-secret-panel'),
      margin: const EdgeInsets.only(bottom: 16),
      child: Padding(
        padding: const EdgeInsets.all(12),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(
              'Worker token created',
              style: Theme.of(context).textTheme.titleMedium,
            ),
            const Text(
              'Copy this secret now and give it to the worker. It is shown only once and cannot be shown again; leaving this screen or the app discards it.',
            ),
            const SizedBox(height: 8),
            if (secret != null)
              SelectableText(
                secret,
                key: const Key('minted-secret'),
                style: const TextStyle(fontFamily: 'monospace'),
              ),
            const SizedBox(height: 8),
            Wrap(
              spacing: 12,
              children: [
                OutlinedButton.icon(
                  key: const Key('copy-secret'),
                  onPressed: secret == null
                      ? null
                      : () => Clipboard.setData(ClipboardData(text: secret)),
                  icon: const Icon(Icons.copy),
                  label: const Text('Copy'),
                ),
                FilledButton(
                  key: const Key('dismiss-secret'),
                  onPressed: controller.dismissSecret,
                  child: const Text('I have saved it'),
                ),
              ],
            ),
            const Text(
              'Copying puts the secret on the clipboard, which other apps may read until you copy something else.',
            ),
          ],
        ),
      ),
    );
  }
}

// -------------------------------------------------------------------- form

class DaemonFormPage extends StatefulWidget {
  const DaemonFormPage({super.key, required this.client, this.base});
  final ClientContext client;
  final DaemonRuntime? base;
  @override
  State<DaemonFormPage> createState() => _DaemonFormPageState();
}

class _DaemonFormPageState extends State<DaemonFormPage> {
  final _form = GlobalKey<FormState>();
  late final DaemonRuntimeDraft _draft = widget.base == null
      ? DaemonRuntimeDraft.create()
      : DaemonRuntimeDraft.edit(widget.base!);
  bool _saved = false;
  String? _problem;

  Future<void> _save(DaemonController d) async {
    final valid = _form.currentState!.validate();
    final problem = _draft.firstError();
    if (!valid || problem != null) {
      setState(() => _problem = problem);
      return;
    }
    setState(() => _problem = null);
    FocusScope.of(context).unfocus();
    final message = _draft.build();
    final ok = _draft.isCreate
        ? await d.create(message)
        : await d.update(message);
    if (!mounted || !ok) return;
    setState(() => _saved = true);
    if (Navigator.canPop(context)) Navigator.pop(context);
  }

  List<Widget> _labelEditor(bool locked) {
    final rows = _draft.labels;
    return [
      const SizedBox(height: 16),
      Text('Labels', style: Theme.of(context).textTheme.titleSmall),
      const Text('Optional key and value pairs that describe the runtime.'),
      for (var i = 0; i < rows.length; i++)
        Row(
          key: ObjectKey(rows[i]),
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Expanded(
              child: TextFormField(
                key: Key('label-key-$i'),
                initialValue: rows[i].key,
                enabled: !locked,
                autocorrect: false,
                decoration: const InputDecoration(labelText: 'Key'),
                onChanged: (v) => setState(() => rows[i].key = v),
              ),
            ),
            const SizedBox(width: 8),
            Expanded(
              child: TextFormField(
                key: Key('label-value-$i'),
                initialValue: rows[i].value,
                enabled: !locked,
                autocorrect: false,
                decoration: const InputDecoration(labelText: 'Value'),
                onChanged: (v) => setState(() => rows[i].value = v),
              ),
            ),
            IconButton(
              key: Key('remove-label-$i'),
              tooltip: 'Remove label',
              onPressed: locked ? null : () => setState(() => rows.removeAt(i)),
              icon: const Icon(Icons.close),
            ),
          ],
        ),
      Align(
        alignment: Alignment.centerLeft,
        child: TextButton.icon(
          key: const Key('add-label'),
          onPressed: locked ? null : () => setState(() => rows.add(LabelRow())),
          icon: const Icon(Icons.add),
          label: const Text('Add label'),
        ),
      ),
    ];
  }

  @override
  Widget build(BuildContext context) {
    final d = widget.client.daemons;
    if (d == null) return const SizedBox.shrink();
    return ListenableBuilder(
      listenable: d,
      builder: (context, _) {
        final locked = d.writing || _saved;
        final draft = _draft;
        return DraftGuard(
          dirty: draft.dirty && !_saved,
          busy: d.writing,
          child: ConfigScaffold(
            controller: d,
            title: draft.isCreate
                ? 'Add daemon runtime'
                : 'Edit daemon runtime',
            child: Form(
              key: _form,
              child: SingleChildScrollView(
                padding: const EdgeInsets.all(16),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.stretch,
                  children: [
                    if (draft.isCreate)
                      TextFormField(
                        key: const Key('daemon-id'),
                        initialValue: draft.id,
                        enabled: !locked,
                        autocorrect: false,
                        decoration: const InputDecoration(
                          labelText: 'ID',
                          helperText:
                              'Remote Agents and worker tokens refer to it by this ID. It cannot be changed later.',
                          helperMaxLines: 2,
                        ),
                        validator: (_) => draft.validateId(),
                        onChanged: (v) => setState(() => draft.id = v),
                      )
                    else
                      SelectableText('ID: ${draft.id}'),
                    TextFormField(
                      key: const Key('daemon-name'),
                      initialValue: draft.name,
                      enabled: !locked,
                      decoration: const InputDecoration(labelText: 'Name'),
                      validator: (_) => draft.validateName(),
                      onChanged: (v) => setState(() => draft.name = v),
                    ),
                    TextFormField(
                      key: const Key('daemon-description'),
                      initialValue: draft.description,
                      enabled: !locked,
                      maxLines: 3,
                      decoration: const InputDecoration(
                        labelText: 'Description (optional)',
                      ),
                      onChanged: (v) => setState(() => draft.description = v),
                    ),
                    ..._labelEditor(locked),
                    if (_problem != null) ...[
                      const SizedBox(height: 8),
                      ErrorLine(_problem!),
                    ],
                    const SizedBox(height: 16),
                    FilledButton(
                      key: const Key('save-daemon'),
                      onPressed: locked || d.unconfirmed
                          ? null
                          : () => _save(d),
                      child: Text(_saved ? 'Saved' : 'Save runtime'),
                    ),
                    if (d.unconfirmed)
                      const Text(
                        'Leave this form and refresh to reconcile the outcome before saving again.',
                      ),
                  ],
                ),
              ),
            ),
          ),
        );
      },
    );
  }
}

// ------------------------------------------------------------------- tasks

/// Tasks in flight, read again on a timer while the page is open and the app
/// is in the foreground.
class DaemonTasksPage extends StatefulWidget {
  const DaemonTasksPage({super.key, required this.client});
  final ClientContext client;
  @override
  State<DaemonTasksPage> createState() => _DaemonTasksPageState();
}

class _DaemonTasksPageState extends State<DaemonTasksPage> {
  Timer? _timer;
  bool _closed = false;

  @override
  void initState() {
    super.initState();
    final d = widget.client.daemons;
    if (d == null) return;
    // One read at a time: the next is scheduled when the last has ended, so a
    // slow server is never asked again before it answered.
    unawaited(_cycle(d.loadTasks()));
  }

  Future<void> _cycle(Future<void> read) async {
    await read;
    if (_closed) return;
    _timer = Timer(widget.client.daemonPollInterval, () {
      if (_closed) return;
      final state = WidgetsBinding.instance.lifecycleState;
      final current = widget.client.daemons;
      if ((state != null && state != AppLifecycleState.resumed) ||
          current == null ||
          current.disposed) {
        // Nothing to ask now; look again after another interval.
        unawaited(_cycle(Future.value()));
        return;
      }
      unawaited(_cycle(current.loadTasks()));
    });
  }

  @override
  void dispose() {
    _closed = true;
    _timer?.cancel();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final d = widget.client.daemons;
    if (d == null) return const SizedBox.shrink();
    return ConfigScaffold(
      controller: d,
      title: 'Daemon tasks',
      child: ListenableBuilder(
        listenable: d,
        builder: (context, _) => ListView(
          padding: const EdgeInsets.all(16),
          children: [
            Text(
              'Tasks in flight',
              style: Theme.of(context).textTheme.headlineSmall,
            ),
            const Text(
              'Only tasks running now are listed. When a task finishes it leaves the list; the server does not keep or report how it ended.',
            ),
            if (d.tasksReadAt != null)
              Text(
                'Read at ${_clock(d.tasksReadAt!)}. Elapsed times are as of that read.',
                key: const Key('tasks-read-at'),
              ),
            Align(
              alignment: Alignment.centerLeft,
              child: TextButton.icon(
                key: const Key('refresh-tasks'),
                onPressed: d.taskState == ConfigLoad.loading
                    ? null
                    : d.loadTasks,
                icon: const Icon(Icons.refresh),
                label: const Text('Refresh'),
              ),
            ),
            if (d.taskState == ConfigLoad.loading && d.tasksReadAt == null)
              const LinearProgressIndicator(semanticsLabel: 'Loading'),
            if (d.taskError != null)
              _loadError(
                context,
                d.tasksReadAt == null
                    ? d.taskError
                    : '${d.taskError} Showing the list from ${_clock(d.tasksReadAt!)}, which may be out of date.',
                const Key('daemon-task-error'),
              ),
            if (d.taskState == ConfigLoad.ready &&
                d.tasks.isEmpty &&
                d.taskError == null)
              const Text('No tasks in flight.'),
            for (final t in d.tasks) _TaskTile(client: widget.client, task: t),
            if (d.ended.isNotEmpty) ...[
              const SizedBox(height: 16),
              Text(
                'Left the list',
                style: Theme.of(context).textTheme.titleMedium,
              ),
              const Text(
                'These stopped being in flight while this page was open. Whether they finished, failed or were cancelled is not reported.',
              ),
              for (final e in d.ended)
                ListTile(
                  key: Key('ended-${e.task.taskId}'),
                  contentPadding: EdgeInsets.zero,
                  title: Text(_taskTitle(e.task)),
                  subtitle: Text(
                    e.cancelDelivered
                        ? 'A cancel was delivered before it left. Outcome not reported.'
                        : 'Outcome not reported.',
                  ),
                ),
            ],
          ],
        ),
      ),
    );
  }
}

String _clock(DateTime t) {
  String two(int n) => n.toString().padLeft(2, '0');
  return '${two(t.hour)}:${two(t.minute)}:${two(t.second)}';
}

String _taskTitle(DaemonTaskInFlight t) {
  final agent = workerText(t.agentName);
  return agent.isEmpty ? 'Task ${workerText(t.taskId, max: 12)}' : agent;
}

String taskProgressLabel(DaemonTaskInFlight t) =>
    '${t.progress.clamp(0, 100)}%';

class _TaskTile extends StatelessWidget {
  const _TaskTile({required this.client, required this.task});
  final ClientContext client;
  final DaemonTaskInFlight task;

  @override
  Widget build(BuildContext context) {
    final d = client.daemons!;
    final cancel = d.cancels[task.taskId];
    return Card(
      key: Key('task-${task.taskId}'),
      child: ListTile(
        isThreeLine: true,
        title: Text(_taskTitle(task)),
        subtitle: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(
              [
                if (workerText(task.currentStep).isNotEmpty)
                  workerText(task.currentStep)
                else
                  'No step reported',
                taskProgressLabel(task),
                elapsedLabel(task.elapsed.seconds.toInt()),
              ].join(' · '),
            ),
            LinearProgressIndicator(
              value: task.progress.clamp(0, 100) / 100,
              semanticsLabel: 'Progress ${taskProgressLabel(task)}',
            ),
            if (cancel != null)
              Text(
                cancel.message,
                key: Key('cancel-result-${task.taskId}'),
                style:
                    cancel.kind == CancelKind.delivered ||
                        cancel.kind == CancelKind.sending
                    ? null
                    : TextStyle(color: Theme.of(context).colorScheme.error),
              ),
          ],
        ),
        trailing: TextButton(
          key: Key('cancel-${task.taskId}'),
          onPressed: cancel?.sending == true
              ? null
              : () => confirmCancel(context, d, task),
          child: const Text('Cancel'),
        ),
        onTap: () => openManagementPage(
          context,
          client,
          DaemonTaskDetailPage(client: client, taskId: task.taskId),
        ),
      ),
    );
  }
}

/// Asks, then hands the cancel to the server. The wording says what a cancel
/// is: a request to the worker, not a stop.
Future<void> confirmCancel(
  BuildContext context,
  DaemonController d,
  DaemonTaskInFlight task,
) async {
  final ok = await confirmChange(
    context,
    'Ask the worker to cancel?',
    'Butter will hand a cancel request for "${_taskTitle(task)}" to runtime "${task.daemonRuntimeId}". The task keeps running until the worker reports back; the worker decides whether and how it stops.',
    action: 'Send cancel',
  );
  if (!ok || d.disposed) return;
  await d.cancel(task);
}

class DaemonTaskDetailPage extends StatelessWidget {
  const DaemonTaskDetailPage({
    super.key,
    required this.client,
    required this.taskId,
  });
  final ClientContext client;
  final String taskId;

  @override
  Widget build(BuildContext context) {
    final d = client.daemons;
    if (d == null) return const SizedBox.shrink();
    return ConfigScaffold(
      controller: d,
      title: 'Daemon task',
      child: ListenableBuilder(
        listenable: d,
        builder: (context, _) {
          final task = d.tasks.where((t) => t.taskId == taskId).firstOrNull;
          final ended = d.ended
              .where((e) => e.task.taskId == taskId)
              .firstOrNull;
          final cancel = d.cancels[taskId];
          return ListView(
            padding: const EdgeInsets.all(16),
            children: [
              if (task == null)
                Text(
                  ended != null
                      ? 'This task is no longer in flight. The server does not report whether it finished, failed or was cancelled.'
                      : 'This task is not in flight. It may have finished, or the list was not read yet. Refresh the tasks.',
                  key: const Key('task-gone'),
                )
              else ...[
                SelectableText(
                  _taskTitle(task),
                  style: Theme.of(context).textTheme.headlineSmall,
                ),
                SelectableText('Task: ${task.taskId}'),
                Text(
                  'Runtime: ${workerText(task.daemonName).isEmpty ? task.daemonRuntimeId : '${workerText(task.daemonName)} (${task.daemonRuntimeId})'}',
                ),
                if (task.acpRuntime.isNotEmpty)
                  Text('Executor: ${workerText(task.acpRuntime)}'),
                Text(
                  'Step: ${workerText(task.currentStep).isEmpty ? 'none reported' : workerText(task.currentStep)}',
                  key: const Key('task-step'),
                ),
                Text('Progress: ${taskProgressLabel(task)}'),
                LinearProgressIndicator(
                  value: task.progress.clamp(0, 100) / 100,
                  semanticsLabel: 'Progress ${taskProgressLabel(task)}',
                ),
                Text(
                  'Elapsed: ${elapsedLabel(task.elapsed.seconds.toInt())}'
                  '${d.tasksReadAt == null ? '' : ' as of ${_clock(d.tasksReadAt!)}'}',
                ),
                if (task.hasStartedAt())
                  Text('Started ${task.startedAt.toDateTime().toLocal()}'),
              ],
              if (cancel != null) ...[
                const SizedBox(height: 8),
                Text(cancel.message, key: const Key('task-cancel-result')),
              ],
              _loadError(context, d.taskError, const Key('daemon-task-error')),
              const SizedBox(height: 12),
              Wrap(
                spacing: 12,
                children: [
                  if (task != null)
                    OutlinedButton(
                      key: const Key('task-cancel'),
                      onPressed: cancel?.sending == true
                          ? null
                          : () => confirmCancel(context, d, task),
                      child: const Text('Cancel task'),
                    ),
                  TextButton.icon(
                    key: const Key('task-refresh'),
                    onPressed: d.taskState == ConfigLoad.loading
                        ? null
                        : d.loadTasks,
                    icon: const Icon(Icons.refresh),
                    label: const Text('Refresh'),
                  ),
                ],
              ),
            ],
          );
        },
      ),
    );
  }
}

// ------------------------------------------------------------- diagnostics

class BridgeDiagnosticsPage extends StatefulWidget {
  const BridgeDiagnosticsPage({super.key, required this.client});
  final ClientContext client;
  @override
  State<BridgeDiagnosticsPage> createState() => _BridgeDiagnosticsPageState();
}

String bytesLabel(int bytes) {
  const mb = 1024 * 1024;
  if (bytes >= 1024 * mb) {
    return '${(bytes / (1024 * mb)).toStringAsFixed(1)} GiB';
  }
  return '${(bytes / mb).toStringAsFixed(0)} MiB';
}

class _BridgeDiagnosticsPageState extends State<BridgeDiagnosticsPage> {
  @override
  void initState() {
    super.initState();
    final d = widget.client.daemons;
    if (d != null) unawaited(d.loadDiagnostics());
  }

  @override
  Widget build(BuildContext context) {
    final d = widget.client.daemons;
    if (d == null) return const SizedBox.shrink();
    return ConfigScaffold(
      controller: d,
      title: 'Server diagnostics',
      child: ListenableBuilder(
        listenable: d,
        builder: (context, _) {
          final x = d.diagnostics;
          return ListView(
            padding: const EdgeInsets.all(16),
            children: [
              Text(
                'Butter server',
                style: Theme.of(context).textTheme.headlineSmall,
              ),
              const Text(
                'These are the Butter server\'s own metrics, shared by every Workspace. They do not describe a worker.',
              ),
              if (d.diagnosticsState == ConfigLoad.loading)
                const LinearProgressIndicator(semanticsLabel: 'Loading'),
              _loadError(
                context,
                d.diagnosticsError,
                const Key('diagnostics-error'),
              ),
              if (x != null) ...[
                Text('CPU: ${x.cpuPercent.toStringAsFixed(1)}%'),
                Text(
                  'Memory: ${bytesLabel(x.memoryUsedBytes.toInt())}'
                  '${x.memoryLimitBytes > 0 ? ' of ${bytesLabel(x.memoryLimitBytes.toInt())}' : ''}',
                ),
                Text('Goroutines: ${x.goroutines}'),
                if (x.hasCheckedAt())
                  Text('Checked ${x.checkedAt.toDateTime().toLocal()}'),
                if (x.latency.isNotEmpty) ...[
                  const SizedBox(height: 8),
                  Text(
                    'Recent latency',
                    style: Theme.of(context).textTheme.titleMedium,
                  ),
                  for (final p in x.latency.reversed.take(10))
                    Text(
                      '${p.hasTimestamp() ? _clock(p.timestamp.toDateTime().toLocal()) : '—'} · ${p.latencyMs} ms',
                    ),
                ],
              ],
              TextButton.icon(
                key: const Key('refresh-diagnostics'),
                onPressed: d.diagnosticsState == ConfigLoad.loading
                    ? null
                    : d.loadDiagnostics,
                icon: const Icon(Icons.refresh),
                label: const Text('Refresh'),
              ),
            ],
          );
        },
      ),
    );
  }
}
