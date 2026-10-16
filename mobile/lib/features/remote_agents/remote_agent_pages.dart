import 'dart:async';

import 'package:flutter/material.dart';

import '../../app/client_context.dart';
import '../../core/config_resource_controller.dart';
import '../../core/config_widgets.dart';
import '../../core/management_page.dart';
import '../../gen/agents/v1/agent.pb.dart';
import '../../gen/agents/v1/dashboard.pb.dart';
import 'remote_agent_controller.dart';
import 'remote_agent_draft.dart';

/// Opens the Workspace's Remote Agents. The list reads fresh from the server.
Future<void> openRemoteAgents(
  BuildContext context,
  ClientContext client,
) async {
  final remote = client.remoteAgents;
  if (remote == null) return;
  unawaited(remote.load());
  unawaited(remote.loadRuntimes());
  await openManagementPage(
    context,
    client,
    RemoteAgentListPage(client: client),
  );
}

String remoteProbeSummary(RemoteProbeView? v) {
  if (v == null) return 'Not probed';
  final latency = v.latencyMs != null ? ' · ${v.latencyMs} ms' : '';
  return switch (v.kind) {
    RemoteProbeKind.probing => 'Probing…',
    RemoteProbeKind.reachable => '${v.reason}$latency',
    RemoteProbeKind.configured => v.reason,
    RemoteProbeKind.unreachable => 'Unreachable: ${v.reason}',
    RemoteProbeKind.error => 'Error: ${v.reason}',
    RemoteProbeKind.failed => 'Probe failed: ${v.reason}',
  };
}

bool _probeIsProblem(RemoteProbeView? v) =>
    v != null &&
    (v.kind == RemoteProbeKind.unreachable ||
        v.kind == RemoteProbeKind.error ||
        v.kind == RemoteProbeKind.failed);

class RemoteAgentListPage extends StatelessWidget {
  const RemoteAgentListPage({super.key, required this.client});
  final ClientContext client;

  @override
  Widget build(BuildContext context) {
    final remote = client.remoteAgents;
    if (remote == null) return const SizedBox.shrink();
    return ConfigScaffold(
      controller: remote,
      title: 'Remote Agents',
      child: ListenableBuilder(
        listenable: remote,
        builder: (context, _) => ListView(
          padding: const EdgeInsets.all(16),
          children: [
            Text(
              '${client.workspace?.name ?? 'Workspace'} · Remote Agents',
              style: Theme.of(context).textTheme.headlineSmall,
            ),
            const SizedBox(height: 8),
            const Text(
              'A Remote Agent is an agent that runs elsewhere. Butter calls it for you over A2A or OpenCode HTTP, or hands its work to a worker connected through a daemon runtime.',
            ),
            const SizedBox(height: 12),
            Wrap(
              spacing: 12,
              children: [
                FilledButton.icon(
                  key: const Key('add-remote'),
                  onPressed: remote.writing
                      ? null
                      : () => openManagementPage(
                          context,
                          client,
                          RemoteAgentFormPage(client: client),
                        ),
                  icon: const Icon(Icons.add),
                  label: const Text('Add Remote Agent'),
                ),
                TextButton.icon(
                  key: const Key('refresh-remote'),
                  onPressed: remote.listState == ConfigLoad.loading
                      ? null
                      : remote.load,
                  icon: const Icon(Icons.refresh),
                  label: const Text('Refresh'),
                ),
              ],
            ),
            if (remote.listState == ConfigLoad.loading)
              const LinearProgressIndicator(semanticsLabel: 'Loading'),
            if (remote.listState == ConfigLoad.failed)
              Semantics(
                liveRegion: true,
                child: Text(
                  remote.listError ?? 'Remote Agents could not be loaded.',
                  key: const Key('remote-list-error'),
                  style: TextStyle(color: Theme.of(context).colorScheme.error),
                ),
              ),
            if (remote.listState == ConfigLoad.ready &&
                remote.items.isEmpty) ...[
              const SizedBox(height: 24),
              Text(
                'No Remote Agents yet',
                style: Theme.of(context).textTheme.titleMedium,
              ),
              const Text('Add one to let Agents here call it.'),
            ],
            for (final agent in remote.items)
              ListTile(
                key: Key('remote-${agent.id}'),
                isThreeLine: true,
                contentPadding: EdgeInsets.zero,
                title: Text(agent.name),
                subtitle: Text(
                  [
                    if (agent.url.isNotEmpty) agent.url,
                    remoteProtocolLabel(agent.protocol),
                    if (remote.referenceProblem(agent) case final problem?)
                      'Problem: $problem',
                    remoteProbeSummary(remote.probes[agent.id]),
                  ].join('\n'),
                ),
                trailing: IconButton(
                  tooltip: 'Probe ${agent.name}',
                  icon: Icon(
                    Icons.network_check,
                    semanticLabel: 'Probe ${agent.name}',
                  ),
                  onPressed:
                      remote.probes[agent.id]?.kind == RemoteProbeKind.probing
                      ? null
                      : () => remote.probe(agent.id),
                ),
                onTap: () => openManagementPage(
                  context,
                  client,
                  RemoteAgentDetailPage(client: client, id: agent.id),
                ),
              ),
          ],
        ),
      ),
    );
  }
}

class RemoteAgentDetailPage extends StatefulWidget {
  const RemoteAgentDetailPage({
    super.key,
    required this.client,
    required this.id,
  });
  final ClientContext client;
  final String id;
  @override
  State<RemoteAgentDetailPage> createState() => _RemoteAgentDetailPageState();
}

class _RemoteAgentDetailPageState extends State<RemoteAgentDetailPage> {
  @override
  void initState() {
    super.initState();
    final remote = widget.client.remoteAgents;
    if (remote != null) {
      unawaited(remote.loadItem(widget.id));
      unawaited(remote.loadRuntimes());
    }
  }

  Future<void> _delete(RemoteAgentController remote, RemoteAgent agent) async {
    final ok = await confirmChange(
      context,
      'Delete Remote Agent?',
      'Delete "${agent.name}"? Butter forgets this connection and any password it holds. Agents that call it stop working. Nothing is changed on the remote agent itself.',
      action: 'Delete',
    );
    if (!ok || !mounted) return;
    if (await remote.delete(agent.id) && mounted && Navigator.canPop(context)) {
      Navigator.pop(context);
    }
  }

  @override
  Widget build(BuildContext context) {
    final remote = widget.client.remoteAgents;
    if (remote == null) return const SizedBox.shrink();
    return ConfigScaffold(
      controller: remote,
      title: 'Remote Agent',
      child: ListenableBuilder(
        listenable: remote,
        builder: (context, _) {
          final agent = remote.item?.id == widget.id ? remote.item : null;
          if (agent == null) {
            return ListView(
              padding: const EdgeInsets.all(16),
              children: [
                if (remote.itemState == ConfigLoad.loading)
                  const LinearProgressIndicator(semanticsLabel: 'Loading'),
                if (remote.itemError != null)
                  Text(
                    remote.itemError!,
                    key: const Key('remote-detail-error'),
                    style: TextStyle(
                      color: Theme.of(context).colorScheme.error,
                    ),
                  ),
                TextButton(
                  onPressed: remote.itemState == ConfigLoad.loading
                      ? null
                      : () => remote.loadItem(widget.id),
                  child: const Text('Load Remote Agent'),
                ),
              ],
            );
          }
          final probe = remote.probes[agent.id];
          final editable = remoteAgentEditable(agent);
          final busy = remote.writing;
          return ListView(
            padding: const EdgeInsets.all(16),
            children: [
              SelectableText(
                agent.name,
                style: Theme.of(context).textTheme.headlineSmall,
              ),
              if (agent.url.isNotEmpty) SelectableText(agent.url),
              SelectableText('ID: ${agent.id}'),
              Text(remoteProtocolLabel(agent.protocol)),
              if (agent.protocol ==
                  RemoteAgentProtocol.REMOTE_AGENT_PROTOCOL_OPENCODE_HTTP) ...[
                Text(
                  agent.opencodeAgent.isEmpty
                      ? 'OpenCode agent: the server default'
                      : 'OpenCode agent: ${agent.opencodeAgent}',
                ),
                Text(
                  agent.opencodeModel.isEmpty
                      ? 'Model: the server default'
                      : 'Model: ${agent.opencodeModel}',
                ),
                Text(
                  agent.password.isEmpty
                      ? 'Basic authentication: none'
                      : 'Basic authentication: ${agent.username.isEmpty ? 'default username' : agent.username} · password set (hidden)',
                  key: const Key('remote-auth'),
                ),
              ],
              if (isDaemonAgent(agent)) ...[
                Text(
                  'Daemon runtime: ${remote.runtimeOf(agent.daemonRuntimeId)?.name.isNotEmpty == true ? '${remote.runtimeOf(agent.daemonRuntimeId)!.name} (${agent.daemonRuntimeId})' : agent.daemonRuntimeId}',
                  key: const Key('remote-daemon-runtime'),
                ),
                Text('Executor: ${agent.acpRuntime}'),
                Text(
                  _workerLine(remote, agent.daemonRuntimeId),
                  key: const Key('remote-daemon-worker'),
                ),
                if (remote.referenceProblem(agent) case final problem?)
                  Padding(
                    padding: const EdgeInsets.only(top: 4),
                    child: Text(
                      problem,
                      key: const Key('remote-reference-problem'),
                      style: TextStyle(
                        color: Theme.of(context).colorScheme.error,
                      ),
                    ),
                  ),
              ],
              if (!editable && !isDaemonAgent(agent))
                const Text(
                  'This protocol is not supported by this app, so it is read-only here.',
                ),
              const SizedBox(height: 12),
              Text(
                remoteProbeSummary(probe),
                key: const Key('remote-probe'),
                style: _probeIsProblem(probe)
                    ? TextStyle(color: Theme.of(context).colorScheme.error)
                    : null,
              ),
              const SizedBox(height: 12),
              Wrap(
                spacing: 12,
                children: [
                  OutlinedButton.icon(
                    key: const Key('probe-remote'),
                    onPressed: probe?.kind == RemoteProbeKind.probing
                        ? null
                        : () => remote.probe(agent.id),
                    icon: const Icon(Icons.network_check),
                    label: const Text('Probe'),
                  ),
                  if (editable)
                    FilledButton.icon(
                      key: const Key('edit-remote'),
                      onPressed: busy
                          ? null
                          : () => openManagementPage(
                              context,
                              widget.client,
                              RemoteAgentFormPage(
                                client: widget.client,
                                base: agent,
                              ),
                            ),
                      icon: const Icon(Icons.edit),
                      label: const Text('Edit'),
                    ),
                  TextButton.icon(
                    key: const Key('refresh-remote-item'),
                    onPressed: remote.itemState == ConfigLoad.loading
                        ? null
                        : () => remote.loadItem(agent.id),
                    icon: const Icon(Icons.refresh),
                    label: const Text('Refresh'),
                  ),
                  if (editable)
                    TextButton.icon(
                      key: const Key('delete-remote'),
                      onPressed: busy || remote.unconfirmed
                          ? null
                          : () => _delete(remote, agent),
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
}

/// Whether a worker is connected for the runtime, from the last read.
String _workerLine(RemoteAgentController remote, String runtimeId) {
  if (remote.online == null) return 'Worker: not checked';
  final w = remote.online!
      .where((w) => w.daemonRuntimeId == runtimeId)
      .firstOrNull;
  if (w == null) {
    return 'Worker: not connected. Runs fail until a worker connects.';
  }
  return 'Worker: connected${w.name.isEmpty ? '' : ' (${w.name})'}';
}

class RemoteAgentFormPage extends StatefulWidget {
  const RemoteAgentFormPage({super.key, required this.client, this.base});
  final ClientContext client;
  final RemoteAgent? base;
  @override
  State<RemoteAgentFormPage> createState() => _RemoteAgentFormPageState();
}

class _RemoteAgentFormPageState extends State<RemoteAgentFormPage> {
  final _form = GlobalKey<FormState>();
  late final RemoteAgentDraft _draft = widget.base == null
      ? RemoteAgentDraft.create()
      : RemoteAgentDraft.edit(widget.base!);
  bool _saved = false;
  String? _problem;

  /// Bumped when the typed password is submitted, so its field is rebuilt
  /// empty.
  int _secretGeneration = 0;

  Future<void> _save(RemoteAgentController remote) async {
    final valid = _form.currentState!.validate();
    var problem = _draft.firstError();
    if (problem == null &&
        _draft.kind == RemoteKind.daemon &&
        remote.runtimes != null &&
        remote.runtimeOf(_draft.daemonRuntimeId) == null) {
      problem =
          'Runtime "${_draft.daemonRuntimeId}" does not exist in this Workspace. Choose another one.';
    }
    if (!valid || problem != null) {
      setState(() => _problem = problem);
      return;
    }
    setState(() => _problem = null);
    FocusScope.of(context).unfocus();
    final message = _draft.build();
    final discards = _draft.discards();
    if (discards.isNotEmpty) {
      final confirmed = await confirmChange(
        context,
        'Discard stored credentials?',
        'Saving "${_draft.name.trim()}" will:\n${discards.map((d) => '• $d').join('\n')}\nThis cannot be undone.',
        action: 'Save and discard',
      );
      if (!confirmed || !mounted) return;
    }
    // The typed password is in the message now: forget it, so a refusal or an
    // unknown outcome leaves nothing secret in the form.
    setState(() {
      _draft.clearTypedSecrets();
      _secretGeneration++;
    });
    final ok = _draft.isCreate
        ? await remote.create(message)
        : await remote.update(message);
    if (!mounted || !ok) return;
    setState(() => _saved = true);
    if (Navigator.canPop(context)) Navigator.pop(context);
  }

  void _change(VoidCallback f) => setState(f);

  List<Widget> _daemonFields(
    RemoteAgentController remote,
    RemoteAgentDraft d,
    bool locked,
  ) {
    final runtimes = remote.runtimes ?? const <DaemonRuntime>[];
    final known = runtimes.any((r) => r.id == d.daemonRuntimeId);
    return [
      const SizedBox(height: 12),
      if (remote.runtimeState == ConfigLoad.loading && remote.runtimes == null)
        const LinearProgressIndicator(semanticsLabel: 'Loading runtimes'),
      if (remote.runtimeState == ConfigLoad.failed) ...[
        ErrorLine(remote.runtimeError ?? 'Runtimes could not be loaded.'),
        TextButton(
          key: const Key('reload-runtimes'),
          onPressed: remote.loadRuntimes,
          child: const Text('Load runtimes again'),
        ),
      ],
      DropdownButtonFormField<String>(
        key: const Key('remote-daemon-runtime-field'),
        initialValue: d.daemonRuntimeId.isEmpty ? null : d.daemonRuntimeId,
        isExpanded: true,
        decoration: const InputDecoration(
          labelText: 'Daemon runtime',
          helperText:
              'The runtime whose connected worker runs this agent. Create runtimes under Daemon execution.',
          helperMaxLines: 2,
        ),
        items: [
          for (final r in runtimes)
            DropdownMenuItem(
              value: r.id,
              child: Text(
                r.name.isEmpty || r.name == r.id ? r.id : '${r.name} (${r.id})',
                overflow: TextOverflow.ellipsis,
              ),
            ),
          // A stored reference to a runtime that is not there stays visible
          // so it can be seen and replaced, never silently dropped.
          if (d.daemonRuntimeId.isNotEmpty && !known && remote.runtimes != null)
            DropdownMenuItem(
              value: d.daemonRuntimeId,
              child: Text(
                '${d.daemonRuntimeId} (not in this Workspace)',
                overflow: TextOverflow.ellipsis,
              ),
            ),
        ],
        validator: (_) => d.validateRuntime(),
        onChanged: locked
            ? null
            : (v) => _change(() => d.daemonRuntimeId = v ?? ''),
      ),
      if (d.daemonRuntimeId.isNotEmpty && remote.runtimes != null && !known)
        const Padding(
          padding: EdgeInsets.only(top: 4),
          child: Text(
            'This runtime does not exist in this Workspace. Choose another one before saving.',
            key: Key('remote-runtime-missing'),
          ),
        ),
      const SizedBox(height: 12),
      Text('Executor', style: Theme.of(context).textTheme.titleSmall),
      SegmentedButton<String>(
        key: const Key('remote-executor'),
        emptySelectionAllowed: true,
        segments: [
          for (final e in daemonExecutors)
            ButtonSegment(value: e, label: Text(e)),
        ],
        selected: {if (daemonExecutors.contains(d.acpRuntime)) d.acpRuntime},
        showSelectedIcon: false,
        onSelectionChanged: locked
            ? null
            : (s) => _change(() => d.acpRuntime = s.isEmpty ? '' : s.first),
      ),
      if (d.validateAcpRuntime() case final problem?
          when d.acpRuntime.isNotEmpty)
        ErrorLine(problem),
    ];
  }

  @override
  Widget build(BuildContext context) {
    final remote = widget.client.remoteAgents;
    if (remote == null) return const SizedBox.shrink();
    return ListenableBuilder(
      listenable: remote,
      builder: (context, _) {
        final locked = remote.writing || _saved;
        final d = _draft;
        return DraftGuard(
          dirty: d.dirty && !_saved,
          busy: remote.writing,
          child: ConfigScaffold(
            controller: remote,
            title: d.isCreate ? 'Add Remote Agent' : 'Edit Remote Agent',
            child: Form(
              key: _form,
              child: SingleChildScrollView(
                padding: const EdgeInsets.all(16),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.stretch,
                  children: [
                    if (d.isCreate)
                      TextFormField(
                        key: const Key('remote-id'),
                        initialValue: d.id,
                        enabled: !locked,
                        autocorrect: false,
                        decoration: const InputDecoration(
                          labelText: 'ID',
                          helperText:
                              'Agents refer to it by this ID. It cannot be changed later.',
                          helperMaxLines: 2,
                        ),
                        validator: (_) => d.validateId(),
                        onChanged: (v) => _change(() => d.id = v),
                      )
                    else
                      SelectableText('ID: ${d.id}'),
                    TextFormField(
                      key: const Key('remote-name'),
                      initialValue: d.name,
                      enabled: !locked,
                      decoration: const InputDecoration(labelText: 'Name'),
                      validator: (_) => d.validateName(),
                      onChanged: (v) => _change(() => d.name = v),
                    ),
                    const SizedBox(height: 12),
                    Text(
                      'Protocol',
                      style: Theme.of(context).textTheme.titleSmall,
                    ),
                    SegmentedButton<RemoteKind>(
                      key: const Key('remote-kind'),
                      segments: const [
                        ButtonSegment(
                          value: RemoteKind.a2a,
                          label: Text('A2A'),
                        ),
                        ButtonSegment(
                          value: RemoteKind.opencode,
                          label: Text('OpenCode HTTP'),
                        ),
                        ButtonSegment(
                          value: RemoteKind.daemon,
                          label: Text('Daemon'),
                        ),
                      ],
                      selected: {d.kind},
                      showSelectedIcon: false,
                      onSelectionChanged: locked
                          ? null
                          : (s) {
                              _change(() => d.kind = s.first);
                              // The picker needs the Workspace's runtimes.
                              if (s.first == RemoteKind.daemon &&
                                  remote.runtimes == null &&
                                  remote.runtimeState != ConfigLoad.loading) {
                                unawaited(remote.loadRuntimes());
                              }
                            },
                    ),
                    if (!d.isCreate &&
                        d.kind == RemoteKind.a2a &&
                        widget.base!.protocol ==
                            RemoteAgentProtocol
                                .REMOTE_AGENT_PROTOCOL_OPENCODE_HTTP)
                      const Padding(
                        padding: EdgeInsets.only(top: 4),
                        child: Text(
                          'Saving as A2A removes the OpenCode agent, model, username and password.',
                        ),
                      ),
                    if (d.kind == RemoteKind.daemon)
                      ..._daemonFields(remote, d, locked)
                    else
                      TextFormField(
                        key: const Key('remote-url'),
                        initialValue: d.url,
                        enabled: !locked,
                        keyboardType: TextInputType.url,
                        autocorrect: false,
                        decoration: InputDecoration(
                          labelText: d.kind == RemoteKind.a2a
                              ? 'Agent URL'
                              : 'OpenCode server URL',
                          helperText: d.kind == RemoteKind.a2a
                              ? 'The agent\'s base URL; Butter reads its agent card from /.well-known/agent.json.'
                              : 'Where `opencode serve` listens. The Butter server connects to it.',
                          helperMaxLines: 3,
                        ),
                        validator: (_) => d.validateUrl(),
                        onChanged: (v) => _change(() => d.url = v),
                      ),
                    if (d.kind == RemoteKind.opencode) ...[
                      TextFormField(
                        key: const Key('remote-opencode-agent'),
                        initialValue: d.opencodeAgent,
                        enabled: !locked,
                        autocorrect: false,
                        decoration: const InputDecoration(
                          labelText: 'OpenCode agent (optional)',
                          helperText: 'Empty uses the server\'s default agent.',
                        ),
                        onChanged: (v) => _change(() => d.opencodeAgent = v),
                      ),
                      TextFormField(
                        key: const Key('remote-opencode-model'),
                        initialValue: d.opencodeModel,
                        enabled: !locked,
                        autocorrect: false,
                        decoration: const InputDecoration(
                          labelText: 'Model (optional)',
                          helperText:
                              'e.g. anthropic/claude-3-5-sonnet. Empty uses the server\'s model.',
                          helperMaxLines: 2,
                        ),
                        onChanged: (v) => _change(() => d.opencodeModel = v),
                      ),
                      TextFormField(
                        key: const Key('remote-username'),
                        initialValue: d.username,
                        enabled: !locked,
                        autocorrect: false,
                        decoration: const InputDecoration(
                          labelText: 'Username (optional)',
                          helperText:
                              'With a password and no username, OpenCode uses "opencode".',
                          helperMaxLines: 2,
                        ),
                        onChanged: (v) => _change(() => d.username = v),
                      ),
                      const SizedBox(height: 8),
                      SecretField(
                        label: 'Password',
                        stored: d.hasStoredPassword,
                        edit: d.passwordEdit,
                        value: d.newPassword,
                        generation: _secretGeneration,
                        enabled: !locked,
                        error: d.validatePassword(),
                        onEdit: (e) => _change(() => d.passwordEdit = e),
                        onValue: (v) => _change(() => d.newPassword = v),
                      ),
                      if (d.sendsCredentialsInClear)
                        const Padding(
                          padding: EdgeInsets.only(top: 8),
                          child: Text(
                            'This URL is plain http: the password would be sent unencrypted.',
                            key: Key('remote-cleartext-warning'),
                          ),
                        ),
                    ],
                    if (_problem != null) ...[
                      const SizedBox(height: 8),
                      ErrorLine(_problem!),
                    ],
                    const SizedBox(height: 16),
                    FilledButton(
                      key: const Key('save-remote'),
                      onPressed: locked || remote.unconfirmed
                          ? null
                          : () => _save(remote),
                      child: Text(_saved ? 'Saved' : 'Save Remote Agent'),
                    ),
                    if (remote.unconfirmed)
                      const Text(
                        'Leave this form and refresh the Remote Agent to reconcile the outcome before saving again.',
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
