import 'dart:async';

import 'package:flutter/material.dart';

import '../../app/client_context.dart';
import '../../core/config_resource_controller.dart';
import '../../core/config_widgets.dart';
import '../../core/management_page.dart';
import '../../core/url_validation.dart';
import '../../gen/agents/v1/agent.pb.dart';
import '../../gen/agents/v1/agent_service.pb.dart';
import 'mcp_controller.dart';
import 'mcp_draft.dart';

/// Opens the Workspace's MCP servers. The list reads fresh from the server.
Future<void> openMcpServers(BuildContext context, ClientContext client) async {
  final mcp = client.mcp;
  if (mcp == null) return;
  unawaited(mcp.load());
  await openManagementPage(context, client, McpListPage(client: client));
}

String mcpTransportLabel(MCPServerTransport t) => switch (t) {
  MCPServerTransport.MCP_SERVER_TRANSPORT_STREAMABLE_HTTP => 'Streamable HTTP',
  MCPServerTransport.MCP_SERVER_TRANSPORT_SSE => 'SSE',
  _ => 'Unspecified transport',
};

/// How the server authenticates, without any credential value.
String mcpAuthSummary(MCPServer s) {
  switch (mcpAuthKindOf(s)) {
    case McpAuthKind.none:
      return 'No authentication';
    case McpAuthKind.staticHeaders:
      final names = s.headers.keys.toList()..sort();
      return 'Static headers: ${names.join(', ')}';
    case McpAuthKind.oauth2:
      final o = s.hasAuth() ? s.auth.oauth2 : null;
      final preregistered = o != null && o.clientId.isNotEmpty;
      return 'OAuth 2.0 · ${preregistered ? 'registered client' : 'dynamic registration'}';
  }
}

/// A sentence for the server's authorization state.
String mcpOAuthStateSummary(MCPOAuthConnectionStatus? s) {
  if (s == null) return 'Authorization not read yet';
  return switch (s.state) {
    MCPOAuthConnectionState.MCPO_AUTH_CONNECTION_STATE_CONNECTED =>
      s.hasConnectedAt()
          ? 'Connected since ${s.connectedAt.toDateTime().toLocal()}'
          : 'Connected',
    MCPOAuthConnectionState.MCPO_AUTH_CONNECTION_STATE_DISCONNECTED =>
      'Not connected',
    MCPOAuthConnectionState
        .MCPO_AUTH_CONNECTION_STATE_REAUTHORIZATION_REQUIRED =>
      'Authorization expired or was revoked. Connect again.',
    _ => 'The authorization state could not be determined.',
  };
}

String mcpStatusSummary(McpStatusView? v) {
  if (v == null) return 'Connection not tested';
  return switch (v.kind) {
    McpStatusKind.probing => 'Testing connection…',
    McpStatusKind.connected =>
      'Connected${v.toolCount > 0 ? ' · ${v.toolCount} tool${v.toolCount == 1 ? '' : 's'} configured' : ''}',
    McpStatusKind.disconnected => 'Not connected: ${v.reason}',
    McpStatusKind.configured => v.reason,
    McpStatusKind.error => 'Error: ${v.reason}',
    McpStatusKind.failed => 'Test failed: ${v.reason}',
  };
}

bool _statusIsProblem(McpStatusView? v) =>
    v != null &&
    (v.kind == McpStatusKind.disconnected ||
        v.kind == McpStatusKind.error ||
        v.kind == McpStatusKind.failed);

Future<void> _openDetail(
  BuildContext context,
  ClientContext client,
  String id,
) => openManagementPage(context, client, McpDetailPage(client: client, id: id));

class McpListPage extends StatelessWidget {
  const McpListPage({super.key, required this.client});
  final ClientContext client;

  @override
  Widget build(BuildContext context) {
    final mcp = client.mcp;
    if (mcp == null) return const SizedBox.shrink();
    return ConfigScaffold(
      controller: mcp,
      title: 'MCP servers',
      child: ListenableBuilder(
        listenable: mcp,
        builder: (context, _) => ListView(
          padding: const EdgeInsets.all(16),
          children: [
            Text(
              '${client.workspace?.name ?? 'Workspace'} · MCP servers',
              style: Theme.of(context).textTheme.headlineSmall,
            ),
            const SizedBox(height: 8),
            const Text(
              'MCP servers give Agents tools. Butter connects to them over HTTP or SSE; this app never connects to one itself.',
            ),
            const SizedBox(height: 12),
            Wrap(
              spacing: 12,
              children: [
                FilledButton.icon(
                  key: const Key('add-mcp'),
                  onPressed: mcp.writing
                      ? null
                      : () => openManagementPage(
                          context,
                          client,
                          McpFormPage(client: client),
                        ),
                  icon: const Icon(Icons.add),
                  label: const Text('Add MCP server'),
                ),
                OutlinedButton.icon(
                  key: const Key('browse-presets'),
                  onPressed: () {
                    unawaited(mcp.loadPresets());
                    openManagementPage(
                      context,
                      client,
                      McpPresetsPage(client: client),
                    );
                  },
                  icon: const Icon(Icons.inventory_2_outlined),
                  label: const Text('Browse presets'),
                ),
                TextButton.icon(
                  key: const Key('refresh-mcp'),
                  onPressed: mcp.listState == ConfigLoad.loading
                      ? null
                      : mcp.load,
                  icon: const Icon(Icons.refresh),
                  label: const Text('Refresh'),
                ),
              ],
            ),
            if (mcp.listState == ConfigLoad.loading)
              const LinearProgressIndicator(semanticsLabel: 'Loading'),
            if (mcp.listState == ConfigLoad.failed)
              Semantics(
                liveRegion: true,
                child: Text(
                  mcp.listError ?? 'MCP servers could not be loaded.',
                  key: const Key('mcp-list-error'),
                  style: TextStyle(color: Theme.of(context).colorScheme.error),
                ),
              ),
            if (mcp.listState == ConfigLoad.ready && mcp.items.isEmpty) ...[
              const SizedBox(height: 24),
              Text(
                'No MCP servers yet',
                style: Theme.of(context).textTheme.titleMedium,
              ),
              const Text('Add one, or install a preset your platform offers.'),
            ],
            for (final server in mcp.items)
              ListTile(
                key: Key('mcp-${server.id}'),
                isThreeLine: true,
                contentPadding: EdgeInsets.zero,
                title: Text(server.name),
                subtitle: Text(
                  [
                    server.url,
                    [
                      mcpTransportLabel(server.transport),
                      mcpAuthSummary(server),
                      if (isInstalledPreset(server)) 'From a preset',
                    ].join(' · '),
                    mcpStatusSummary(mcp.statuses[server.id]),
                  ].join('\n'),
                ),
                trailing: IconButton(
                  tooltip: 'Test connection to ${server.name}',
                  icon: Icon(
                    Icons.network_check,
                    semanticLabel: 'Test connection to ${server.name}',
                  ),
                  onPressed:
                      mcp.statuses[server.id]?.kind == McpStatusKind.probing
                      ? null
                      : () => mcp.probe(server.id),
                ),
                onTap: () => _openDetail(context, client, server.id),
              ),
          ],
        ),
      ),
    );
  }
}

class McpDetailPage extends StatefulWidget {
  const McpDetailPage({super.key, required this.client, required this.id});
  final ClientContext client;
  final String id;
  @override
  State<McpDetailPage> createState() => _McpDetailPageState();
}

class _McpDetailPageState extends State<McpDetailPage> {
  @override
  void initState() {
    super.initState();
    final mcp = widget.client.mcp;
    if (mcp != null) {
      unawaited(mcp.loadItem(widget.id));
    }
  }

  Future<void> _delete(McpController mcp, MCPServer server) async {
    final ok = await confirmChange(
      context,
      'Delete MCP server?',
      'Delete "${server.name}"? Butter forgets this server and any authorization it holds for it. Agents that use it lose its tools. Nothing is changed on the MCP server itself.',
      action: 'Delete',
    );
    if (!ok || !mounted) return;
    if (await mcp.delete(server.id) && mounted && Navigator.canPop(context)) {
      Navigator.pop(context);
    }
  }

  Future<void> _disconnect(McpController mcp, MCPServer server) async {
    final ok = await confirmChange(
      context,
      'Disconnect "${server.name}"?',
      'Butter deletes the stored authorization for this server. Agents cannot use it until someone connects again. The account itself is not changed at the provider.',
      action: 'Disconnect',
    );
    if (ok && mounted) await mcp.disconnectOAuth(server.id);
  }

  @override
  Widget build(BuildContext context) {
    final mcp = widget.client.mcp;
    if (mcp == null) return const SizedBox.shrink();
    return ConfigScaffold(
      controller: mcp,
      title: 'MCP server',
      child: ListenableBuilder(
        listenable: mcp,
        builder: (context, _) {
          final server = mcp.item?.id == widget.id ? mcp.item : null;
          if (server == null) {
            return ListView(
              padding: const EdgeInsets.all(16),
              children: [
                if (mcp.itemState == ConfigLoad.loading)
                  const LinearProgressIndicator(semanticsLabel: 'Loading'),
                if (mcp.itemError != null)
                  Text(
                    mcp.itemError!,
                    key: const Key('mcp-detail-error'),
                    style: TextStyle(
                      color: Theme.of(context).colorScheme.error,
                    ),
                  ),
                TextButton(
                  onPressed: mcp.itemState == ConfigLoad.loading
                      ? null
                      : () => mcp.loadItem(widget.id),
                  child: const Text('Load MCP server'),
                ),
              ],
            );
          }
          final status = mcp.statuses[server.id];
          final tools = mcp.tools[server.id];
          final busy = mcp.writing;
          final isOAuth = mcpAuthKindOf(server) == McpAuthKind.oauth2;
          return ListView(
            padding: const EdgeInsets.all(16),
            children: [
              SelectableText(
                server.name,
                style: Theme.of(context).textTheme.headlineSmall,
              ),
              SelectableText(server.url),
              SelectableText('ID: ${server.id}'),
              Text(mcpTransportLabel(server.transport)),
              Text(mcpAuthSummary(server), key: const Key('mcp-auth')),
              if (isInstalledPreset(server))
                const Text(
                  'Installed from a global preset. Its credentials, if any, are never shown here; editing keeps them unless you replace or remove them. The preset itself is not changed.',
                ),
              Text(
                server.toolFilter.isEmpty
                    ? 'Tools: all the server offers'
                    : 'Tools allowed: ${server.toolFilter.join(', ')}',
              ),
              if (server.timeoutSeconds > 0)
                Text('Connection timeout: ${server.timeoutSeconds} s'),
              const SizedBox(height: 12),
              Text(
                mcpStatusSummary(status),
                key: const Key('mcp-status'),
                style: _statusIsProblem(status)
                    ? TextStyle(color: Theme.of(context).colorScheme.error)
                    : null,
              ),
              if (tools != null) _ToolsView(view: tools),
              if (isOAuth) ...[
                const Divider(height: 32),
                _OAuthSection(
                  mcp: mcp,
                  server: server,
                  onDisconnect: () => _disconnect(mcp, server),
                ),
              ],
              const SizedBox(height: 12),
              Wrap(
                spacing: 12,
                children: [
                  OutlinedButton.icon(
                    key: const Key('probe-mcp'),
                    onPressed: status?.kind == McpStatusKind.probing
                        ? null
                        : () => mcp.probe(server.id),
                    icon: const Icon(Icons.network_check),
                    label: const Text('Test connection'),
                  ),
                  OutlinedButton.icon(
                    key: const Key('list-tools'),
                    onPressed: tools?.state == ConfigLoad.loading
                        ? null
                        : () => mcp.loadTools(server.id),
                    icon: const Icon(Icons.build_outlined),
                    label: const Text('List tools'),
                  ),
                  FilledButton.icon(
                    key: const Key('edit-mcp'),
                    onPressed: busy
                        ? null
                        : () => openManagementPage(
                            context,
                            widget.client,
                            McpFormPage(client: widget.client, base: server),
                          ),
                    icon: const Icon(Icons.edit),
                    label: const Text('Edit'),
                  ),
                  TextButton.icon(
                    key: const Key('refresh-mcp-item'),
                    onPressed: mcp.itemState == ConfigLoad.loading
                        ? null
                        : () => mcp.loadItem(server.id),
                    icon: const Icon(Icons.refresh),
                    label: const Text('Refresh'),
                  ),
                  TextButton.icon(
                    key: const Key('delete-mcp'),
                    onPressed: busy || mcp.unconfirmed
                        ? null
                        : () => _delete(mcp, server),
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

class _ToolsView extends StatelessWidget {
  const _ToolsView({required this.view});
  final McpToolsView view;

  @override
  Widget build(BuildContext context) {
    if (view.state == ConfigLoad.loading) {
      return const Padding(
        padding: EdgeInsets.only(top: 8),
        child: LinearProgressIndicator(semanticsLabel: 'Listing tools'),
      );
    }
    if (view.state == ConfigLoad.failed) {
      return Padding(
        padding: const EdgeInsets.only(top: 8),
        child: Text(
          'Tools could not be listed: ${view.error}',
          key: const Key('mcp-tools-error'),
          style: TextStyle(color: Theme.of(context).colorScheme.error),
        ),
      );
    }
    return Column(
      key: const Key('mcp-tools'),
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        const SizedBox(height: 8),
        Text(
          view.tools.isEmpty
              ? 'The server offers no tools.'
              : 'Tools (${view.tools.where((t) => t.allowed).length} of ${view.tools.length} allowed)',
          style: Theme.of(context).textTheme.titleSmall,
        ),
        for (final tool in view.tools)
          ListTile(
            key: Key('tool-${tool.name}'),
            dense: true,
            contentPadding: EdgeInsets.zero,
            title: Text(tool.name),
            subtitle: Text(
              [
                if (tool.description.isNotEmpty) tool.description,
                tool.allowed ? 'Allowed' : 'Blocked by the tool filter',
              ].join('\n'),
            ),
          ),
      ],
    );
  }
}

class _OAuthSection extends StatelessWidget {
  const _OAuthSection({
    required this.mcp,
    required this.server,
    required this.onDisconnect,
  });
  final McpController mcp;
  final MCPServer server;
  final VoidCallback onDisconnect;

  @override
  Widget build(BuildContext context) {
    final view = mcp.oauthOf(server.id);
    final starting = view.phase == McpOAuthPhase.starting;
    final connected =
        view.status?.state ==
        MCPOAuthConnectionState.MCPO_AUTH_CONNECTION_STATE_CONNECTED;
    final problem =
        view.phase == McpOAuthPhase.failed ||
        view.phase == McpOAuthPhase.timedOut;
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text('Authorization', style: Theme.of(context).textTheme.titleMedium),
        Text(mcpOAuthStateSummary(view.status), key: const Key('oauth-state')),
        if (view.message != null)
          Semantics(
            liveRegion: true,
            child: Text(
              view.message!,
              key: const Key('oauth-message'),
              style: problem
                  ? TextStyle(color: Theme.of(context).colorScheme.error)
                  : null,
            ),
          ),
        if (starting || view.reading)
          const LinearProgressIndicator(semanticsLabel: 'Working'),
        const SizedBox(height: 8),
        Wrap(
          spacing: 12,
          children: [
            FilledButton.icon(
              key: const Key('oauth-connect'),
              onPressed: starting || mcp.writing || view.waiting
                  ? null
                  : () => mcp.connectOAuth(server.id),
              icon: const Icon(Icons.open_in_browser),
              label: Text(
                connected ? 'Reconnect in browser' : 'Connect in browser',
              ),
            ),
            OutlinedButton.icon(
              key: const Key('oauth-check'),
              onPressed: view.reading ? null : () => mcp.loadOAuth(server.id),
              icon: const Icon(Icons.refresh),
              label: Text(view.waiting ? 'Check now' : 'Check state'),
            ),
            if (view.waiting)
              TextButton.icon(
                key: const Key('oauth-cancel'),
                onPressed: () => mcp.cancelOAuth(server.id),
                icon: const Icon(Icons.close),
                label: const Text('Stop waiting'),
              ),
            if (connected)
              TextButton.icon(
                key: const Key('oauth-disconnect'),
                onPressed: mcp.writing || mcp.unconfirmed ? null : onDisconnect,
                icon: const Icon(Icons.link_off),
                label: const Text('Disconnect'),
              ),
          ],
        ),
      ],
    );
  }
}

class McpPresetsPage extends StatelessWidget {
  const McpPresetsPage({super.key, required this.client});
  final ClientContext client;

  @override
  Widget build(BuildContext context) {
    final mcp = client.mcp;
    if (mcp == null) return const SizedBox.shrink();
    return ConfigScaffold(
      controller: mcp,
      title: 'MCP presets',
      child: ListenableBuilder(
        listenable: mcp,
        builder: (context, _) => ListView(
          padding: const EdgeInsets.all(16),
          children: [
            Text(
              'Presets from your platform',
              style: Theme.of(context).textTheme.headlineSmall,
            ),
            const SizedBox(height: 8),
            const Text(
              'A preset is an MCP server your platform administrator maintains. Installing one adds it to this Workspace; credentials it carries stay on the server and are never shown. Presets are changed only by the administrator.',
            ),
            TextButton.icon(
              key: const Key('refresh-presets'),
              onPressed: mcp.presetState == ConfigLoad.loading
                  ? null
                  : mcp.loadPresets,
              icon: const Icon(Icons.refresh),
              label: const Text('Refresh presets'),
            ),
            if (mcp.presetState == ConfigLoad.loading)
              const LinearProgressIndicator(semanticsLabel: 'Loading'),
            if (mcp.presetState == ConfigLoad.failed)
              Text(
                mcp.presetError ?? 'Presets could not be loaded.',
                key: const Key('preset-error'),
                style: TextStyle(color: Theme.of(context).colorScheme.error),
              ),
            if (mcp.presetState == ConfigLoad.ready && mcp.presets.isEmpty)
              const Text('Your platform offers no presets.'),
            for (final preset in mcp.presets)
              _PresetTile(mcp: mcp, client: client, preset: preset),
          ],
        ),
      ),
    );
  }
}

class _PresetTile extends StatelessWidget {
  const _PresetTile({
    required this.mcp,
    required this.client,
    required this.preset,
  });
  final McpController mcp;
  final ClientContext client;
  final MCPServer preset;

  @override
  Widget build(BuildContext context) {
    final existing = mcp.byId(preset.id);
    final installed = existing != null && isInstalledPreset(existing);
    final taken = existing != null && !installed;
    final hasCredentials =
        preset.headers.isNotEmpty ||
        (preset.hasAuth() && preset.auth.oauth2.clientId.isNotEmpty);
    return ListTile(
      key: Key('preset-${preset.id}'),
      isThreeLine: true,
      contentPadding: EdgeInsets.zero,
      title: Text(preset.name),
      subtitle: Text(
        [
          preset.url,
          [
            mcpTransportLabel(preset.transport),
            switch (mcpAuthKindOf(preset)) {
              McpAuthKind.none => 'No authentication',
              McpAuthKind.staticHeaders => 'Static credentials held by Butter',
              McpAuthKind.oauth2 => 'OAuth 2.0',
            },
            if (hasCredentials) 'Credentials hidden',
          ].join(' · '),
          if (installed) 'Installed in this Workspace',
          if (taken)
            'This Workspace already has a different server with this ID',
        ].join('\n'),
      ),
      trailing: TextButton(
        key: Key('install-${preset.id}'),
        onPressed: mcp.writing || mcp.unconfirmed || taken
            ? null
            : () async {
                final ok = await confirmChange(
                  context,
                  installed ? 'Reinstall preset?' : 'Install preset?',
                  installed
                      ? 'Reinstall "${preset.name}"? This overwrites this Workspace\'s copy with the platform\'s current preset, including any change made to it here.'
                      : 'Install "${preset.name}" into this Workspace? Agents here can then use its tools.',
                  action: installed ? 'Reinstall' : 'Install',
                );
                if (ok && context.mounted) await mcp.install(preset.id);
              },
        child: Text(installed ? 'Reinstall' : 'Install'),
      ),
    );
  }
}

class McpFormPage extends StatefulWidget {
  const McpFormPage({super.key, required this.client, this.base});
  final ClientContext client;
  final MCPServer? base;
  @override
  State<McpFormPage> createState() => _McpFormPageState();
}

class _McpFormPageState extends State<McpFormPage> {
  final _form = GlobalKey<FormState>();
  late final McpDraft _draft = widget.base == null
      ? McpDraft.create()
      : McpDraft.edit(widget.base!);
  bool _saved = false;
  String? _problem;

  /// Bumped when typed secrets are submitted, so their fields are rebuilt
  /// empty.
  int _secretGeneration = 0;

  Future<void> _save(McpController mcp) async {
    final valid = _form.currentState!.validate();
    final problem = _draft.firstError();
    if (!valid || problem != null) {
      setState(() => _problem = problem);
      return;
    }
    setState(() => _problem = null);
    FocusScope.of(context).unfocus();
    final save = _draft.buildSave();
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
    // What was typed is in the message now: forget it, so a refusal or an
    // unknown outcome leaves nothing secret in the form.
    setState(() {
      _draft.clearTypedSecrets();
      _secretGeneration++;
    });
    final ok = _draft.isCreate
        ? await mcp.create(save.server)
        : await mcp.saveEdit(save);
    if (!mounted || !ok) return;
    setState(() => _saved = true);
    if (Navigator.canPop(context)) Navigator.pop(context);
  }

  void _change(VoidCallback f) => setState(f);

  @override
  Widget build(BuildContext context) {
    final mcp = widget.client.mcp;
    if (mcp == null) return const SizedBox.shrink();
    return ListenableBuilder(
      listenable: mcp,
      builder: (context, _) {
        final locked = mcp.writing || _saved;
        final d = _draft;
        return DraftGuard(
          dirty: d.dirty && !_saved,
          busy: mcp.writing,
          child: ConfigScaffold(
            controller: mcp,
            title: d.isCreate ? 'Add MCP server' : 'Edit MCP server',
            child: Form(
              key: _form,
              child: SingleChildScrollView(
                padding: const EdgeInsets.all(16),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.stretch,
                  children: [
                    if (d.isCreate)
                      TextFormField(
                        key: const Key('mcp-id'),
                        initialValue: d.id,
                        enabled: !locked,
                        autocorrect: false,
                        decoration: const InputDecoration(
                          labelText: 'ID',
                          helperText:
                              'Agents refer to the server by this ID. It cannot be changed later.',
                          helperMaxLines: 2,
                        ),
                        validator: (_) => d.validateId(),
                        onChanged: (v) => _change(() => d.id = v),
                      )
                    else
                      SelectableText('ID: ${d.id}'),
                    TextFormField(
                      key: const Key('mcp-name'),
                      initialValue: d.name,
                      enabled: !locked,
                      decoration: const InputDecoration(labelText: 'Name'),
                      validator: (_) => d.validateName(),
                      onChanged: (v) => _change(() => d.name = v),
                    ),
                    TextFormField(
                      key: const Key('mcp-url'),
                      initialValue: d.url,
                      enabled: !locked,
                      keyboardType: TextInputType.url,
                      autocorrect: false,
                      decoration: const InputDecoration(
                        labelText: 'Server URL',
                        helperText:
                            'The MCP endpoint, e.g. https://mcp.example.com/mcp. The Butter server connects to it.',
                        helperMaxLines: 3,
                      ),
                      validator: (_) => d.validateUrl(),
                      onChanged: (v) => _change(() => d.url = v),
                    ),
                    const SizedBox(height: 12),
                    Text(
                      'Transport',
                      style: Theme.of(context).textTheme.titleSmall,
                    ),
                    SegmentedButton<MCPServerTransport>(
                      key: const Key('mcp-transport'),
                      segments: const [
                        ButtonSegment(
                          value: MCPServerTransport
                              .MCP_SERVER_TRANSPORT_STREAMABLE_HTTP,
                          label: Text('Streamable HTTP'),
                        ),
                        ButtonSegment(
                          value: MCPServerTransport.MCP_SERVER_TRANSPORT_SSE,
                          label: Text('SSE'),
                        ),
                      ],
                      selected: {d.transport},
                      showSelectedIcon: false,
                      onSelectionChanged: locked
                          ? null
                          : (s) => _change(() => d.transport = s.first),
                    ),
                    const SizedBox(height: 12),
                    Text(
                      'Authentication',
                      style: Theme.of(context).textTheme.titleSmall,
                    ),
                    SegmentedButton<McpAuthKind>(
                      key: const Key('mcp-auth-kind'),
                      segments: const [
                        ButtonSegment(
                          value: McpAuthKind.none,
                          label: Text('None'),
                        ),
                        ButtonSegment(
                          value: McpAuthKind.staticHeaders,
                          label: Text('Headers'),
                        ),
                        ButtonSegment(
                          value: McpAuthKind.oauth2,
                          label: Text('OAuth 2.0'),
                        ),
                      ],
                      selected: {d.authKind},
                      showSelectedIcon: false,
                      onSelectionChanged: locked
                          ? null
                          : (s) => _change(() => d.authKind = s.first),
                    ),
                    _headers(locked),
                    if (d.authKind == McpAuthKind.oauth2) _oauth(locked),
                    if (d.sendsCredentialsInClear)
                      const Padding(
                        padding: EdgeInsets.only(top: 8),
                        child: Text(
                          'This URL is plain http: the credentials above would be sent unencrypted.',
                          key: Key('mcp-cleartext-warning'),
                        ),
                      ),
                    const SizedBox(height: 12),
                    TextFormField(
                      key: const Key('mcp-tool-filter'),
                      initialValue: d.toolFilter,
                      enabled: !locked,
                      minLines: 1,
                      maxLines: 4,
                      autocorrect: false,
                      decoration: const InputDecoration(
                        labelText: 'Allowed tools (optional)',
                        helperText:
                            'One tool name per line. Empty allows every tool the server offers.',
                        helperMaxLines: 2,
                      ),
                      onChanged: (v) => _change(() => d.toolFilter = v),
                    ),
                    TextFormField(
                      key: const Key('mcp-timeout'),
                      initialValue: d.timeout,
                      enabled: !locked,
                      keyboardType: TextInputType.number,
                      decoration: const InputDecoration(
                        labelText: 'Connection timeout in seconds (optional)',
                        helperText: 'Empty uses the server default.',
                      ),
                      validator: (_) => d.validateTimeout(),
                      onChanged: (v) => _change(() => d.timeout = v),
                    ),
                    if (_problem != null) ...[
                      const SizedBox(height: 8),
                      ErrorLine(_problem!),
                    ],
                    const SizedBox(height: 16),
                    FilledButton(
                      key: const Key('save-mcp'),
                      onPressed: locked || mcp.unconfirmed
                          ? null
                          : () => _save(mcp),
                      child: Text(_saved ? 'Saved' : 'Save MCP server'),
                    ),
                    if (mcp.unconfirmed)
                      const Text(
                        'Leave this form and refresh the server to reconcile the outcome before saving again.',
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

  Widget _headers(bool locked) {
    final d = _draft;
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        const SizedBox(height: 8),
        Text('Headers', style: Theme.of(context).textTheme.titleSmall),
        Text(
          d.authKind == McpAuthKind.staticHeaders
              ? 'Sent with every request. A stored value is never shown here.'
              : 'Optional extra headers, sent with every request alongside the sign-in. A stored value is never shown here.',
        ),
        for (var i = 0; i < d.headers.length; i++)
          _headerRow(d.headers[i], i, locked),
        Align(
          alignment: Alignment.centerLeft,
          child: TextButton.icon(
            key: const Key('add-header'),
            onPressed: locked
                ? null
                : () => _change(() => d.headers.add(McpHeaderRow.added())),
            icon: const Icon(Icons.add),
            label: const Text('Add header'),
          ),
        ),
      ],
    );
  }

  Widget _headerRow(McpHeaderRow row, int i, bool locked) {
    final d = _draft;
    if (row.removed) {
      return ListTile(
        key: ObjectKey(row),
        contentPadding: EdgeInsets.zero,
        title: Text(row.key),
        subtitle: const Text('Will be removed'),
        trailing: TextButton(
          onPressed: locked ? null : () => _change(() => row.removed = false),
          child: const Text('Undo'),
        ),
      );
    }
    if (row.isStored) {
      final replacing = row.replacement != null;
      return Column(
        key: ValueKey((row, _secretGeneration)),
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          ListTile(
            contentPadding: EdgeInsets.zero,
            title: Text(row.key),
            subtitle: Text(replacing ? 'New value' : 'Value kept (hidden)'),
            trailing: Wrap(
              children: [
                TextButton(
                  key: Key('header-replace-$i'),
                  onPressed: locked
                      ? null
                      : () => _change(
                          () => row.replacement = replacing ? null : '',
                        ),
                  child: Text(replacing ? 'Keep' : 'Replace'),
                ),
                TextButton(
                  key: Key('header-remove-$i'),
                  onPressed: locked
                      ? null
                      : () => _change(() => row.removed = true),
                  child: const Text('Remove'),
                ),
              ],
            ),
          ),
          if (replacing)
            TextFormField(
              key: Key('header-value-$i'),
              initialValue: row.replacement,
              enabled: !locked,
              obscureText: true,
              autocorrect: false,
              enableSuggestions: false,
              autofillHints: const [],
              decoration: InputDecoration(
                labelText: 'New value for ${row.key}',
              ),
              validator: (_) => d.validateHeaderValue(row),
              onChanged: (v) => _change(() => row.replacement = v),
            ),
        ],
      );
    }
    return Column(
      key: ValueKey((row, _secretGeneration)),
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        TextFormField(
          key: Key('header-key-$i'),
          initialValue: row.key,
          enabled: !locked,
          autocorrect: false,
          decoration: const InputDecoration(labelText: 'Header name'),
          validator: (_) => d.validateHeaderKey(row),
          onChanged: (v) => _change(() => row.key = v),
        ),
        TextFormField(
          key: Key('header-value-$i'),
          initialValue: row.replacement,
          enabled: !locked,
          obscureText: true,
          autocorrect: false,
          enableSuggestions: false,
          autofillHints: const [],
          decoration: const InputDecoration(labelText: 'Header value'),
          validator: (_) => d.validateHeaderValue(row),
          onChanged: (v) => _change(() => row.replacement = v),
        ),
        Align(
          alignment: Alignment.centerLeft,
          child: TextButton(
            key: Key('header-remove-$i'),
            onPressed: locked
                ? null
                : () => _change(() => d.headers.remove(row)),
            child: const Text('Remove this header'),
          ),
        ),
      ],
    );
  }

  Widget _oauth(bool locked) {
    final d = _draft;
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        SwitchListTile(
          key: const Key('mcp-dynamic'),
          contentPadding: EdgeInsets.zero,
          title: const Text('Register the client automatically'),
          subtitle: const Text(
            'Butter registers itself with the authorization server when it supports that. Turn off to use a client ID you already have.',
          ),
          value: d.dynamicRegistration,
          onChanged: locked
              ? null
              : (v) => _change(() => d.dynamicRegistration = v),
        ),
        if (!d.dynamicRegistration) ...[
          TextFormField(
            key: const Key('mcp-client-id'),
            initialValue: d.clientId,
            enabled: !locked,
            autocorrect: false,
            decoration: const InputDecoration(labelText: 'Client ID'),
            validator: (_) => d.validateClientId(),
            onChanged: (v) => _change(() => d.clientId = v),
          ),
          const SizedBox(height: 8),
          SecretField(
            label: 'Client secret',
            stored: d.hasStoredSecret,
            edit: d.secretEdit,
            value: d.newSecret,
            generation: _secretGeneration,
            enabled: !locked,
            error: d.validateSecret(),
            onEdit: (e) => _change(() => d.secretEdit = e),
            onValue: (v) => _change(() => d.newSecret = v),
          ),
        ],
        TextFormField(
          key: const Key('mcp-scopes'),
          initialValue: d.scopes,
          enabled: !locked,
          autocorrect: false,
          decoration: const InputDecoration(
            labelText: 'Scopes (optional)',
            helperText: 'Separated by spaces or commas.',
          ),
          onChanged: (v) => _change(() => d.scopes = v),
        ),
        ExpansionTile(
          key: const Key('mcp-advanced'),
          title: const Text('Advanced endpoints'),
          subtitle: const Text('Only when discovery does not find them.'),
          children: [
            _url(
              'Authorization URL',
              'mcp-authorization-url',
              d.authorizationUrl,
              (v) => d.authorizationUrl = v,
              locked,
            ),
            _url(
              'Token URL',
              'mcp-token-url',
              d.tokenUrl,
              (v) => d.tokenUrl = v,
              locked,
            ),
            _url(
              'Protected resource metadata URL',
              'mcp-resource-metadata',
              d.resourceMetadataUrl,
              (v) => d.resourceMetadataUrl = v,
              locked,
            ),
            _url(
              'Authorization server URL',
              'mcp-authorization-server',
              d.authorizationServerUrl,
              (v) => d.authorizationServerUrl = v,
              locked,
            ),
            _url(
              'Resource',
              'mcp-resource',
              d.resource,
              (v) => d.resource = v,
              locked,
            ),
          ],
        ),
      ],
    );
  }

  Widget _url(
    String label,
    String key,
    String initial,
    void Function(String) set,
    bool locked,
  ) => TextFormField(
    key: Key(key),
    initialValue: initial,
    enabled: !locked,
    keyboardType: TextInputType.url,
    autocorrect: false,
    decoration: InputDecoration(labelText: '$label (optional)'),
    validator: (v) => validateHttpUrl(v, required: false),
    onChanged: (v) => _change(() => set(v)),
  );
}
