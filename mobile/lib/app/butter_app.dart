import 'dart:async';

import 'package:flutter/material.dart';

import 'client_context.dart';
import 'app_preferences.dart';
import '../core/management_page.dart';
import '../features/butterboxes/box_pages.dart';
import '../features/mcp/mcp_pages.dart';
import '../features/daemons/daemon_pages.dart';
import '../features/files/file_pages.dart';
import '../features/repo_binding/repo_binding_pages.dart';
import '../features/remote_agents/remote_agent_pages.dart';
import '../features/model_providers/model_provider_pages.dart';
import '../features/agents/agent_pages.dart';
import '../features/chat/chat_pages.dart';
import '../features/sessions/session_pages.dart';
import '../features/profile/profile_page.dart';
import '../features/settings/settings_page.dart';
import '../features/workspaces/workspace_pages.dart';

class ButterApp extends StatefulWidget {
  const ButterApp({super.key, required this.context, this.preferences});
  final ClientContext context;
  final AppPreferences? preferences;
  @override
  State<ButterApp> createState() => _ButterAppState();
}

class _ButterAppState extends State<ButterApp> with WidgetsBindingObserver {
  late final _preferences = widget.preferences ?? AppPreferences();
  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addObserver(this);
    if (widget.preferences == null) unawaited(_preferences.start());
  }

  @override
  void dispose() {
    WidgetsBinding.instance.removeObserver(this);
    if (widget.preferences == null) _preferences.dispose();
    super.dispose();
  }

  @override
  void didChangeAppLifecycleState(AppLifecycleState state) {
    if (state == AppLifecycleState.resumed) {
      unawaited(widget.context.revalidate());
    }
    if (state == AppLifecycleState.paused) widget.context.suspend();
  }

  @override
  Widget build(BuildContext context) => AnimatedBuilder(
    animation: _preferences,
    builder: (context, _) => MaterialApp(
      title: 'Butter Dev',
      themeMode: _preferences.theme,
      theme: ThemeData(
        colorSchemeSeed: Colors.amber,
        brightness: Brightness.light,
      ),
      darkTheme: ThemeData(
        colorSchemeSeed: Colors.amber,
        brightness: Brightness.dark,
      ),
      home: _Home(client: widget.context, preferences: _preferences),
    ),
  );
}

class _Home extends StatelessWidget {
  const _Home({required this.client, required this.preferences});
  final ClientContext client;
  final AppPreferences preferences;
  @override
  Widget build(BuildContext context) => AnimatedBuilder(
    animation: client,
    builder: (context, _) => Scaffold(
      appBar: AppBar(
        title: const Text('Butter Dev'),
        actions: [
          IconButton(
            tooltip: 'Appearance and display',
            icon: const Icon(
              Icons.palette,
              semanticLabel: 'Appearance and display',
            ),
            onPressed: preferences.busy
                ? null
                : () => Navigator.push<void>(
                    context,
                    MaterialPageRoute(
                      builder: (_) => SettingsPage(preferences: preferences),
                    ),
                  ),
          ),
          if (client.user != null)
            IconButton(
              tooltip: 'My profile',
              icon: const Icon(
                Icons.account_circle,
                semanticLabel: 'My profile',
              ),
              onPressed: client.busy
                  ? null
                  : () => openManagementPage(
                      context,
                      client,
                      ProfilePage(client: client),
                    ),
            ),
          if (client.address != null)
            IconButton(
              tooltip: 'Change server',
              icon: const Icon(Icons.settings, semanticLabel: 'Change server'),
              onPressed: client.busy
                  ? null
                  : () async {
                      final confirmed = await showDialog<bool>(
                        context: context,
                        builder: (context) => AlertDialog(
                          title: const Text('Change server?'),
                          content: const Text(
                            'This clears your credentials and cached data. It does not stop server runs.',
                          ),
                          actions: [
                            TextButton(
                              onPressed: () => Navigator.pop(context, false),
                              child: const Text('Keep server'),
                            ),
                            FilledButton(
                              onPressed: () => Navigator.pop(context, true),
                              child: const Text('Continue'),
                            ),
                          ],
                        ),
                      );
                      if (confirmed != true || !context.mounted) return;
                      await Navigator.push<void>(
                        context,
                        MaterialPageRoute(
                          builder: (context) => Scaffold(
                            appBar: AppBar(title: const Text('Server address')),
                            body: _ServerForm(
                              client: client,
                              onSaved: () => Navigator.pop(context),
                            ),
                          ),
                        ),
                      );
                    },
            ),
          if (client.address != null)
            IconButton(
              tooltip: 'Sign out',
              icon: const Icon(Icons.logout, semanticLabel: 'Sign out'),
              onPressed: client.busy ? null : client.logout,
            ),
        ],
      ),
      body: SafeArea(
        child: Column(
          children: [
            if (client.address != null)
              Padding(
                padding: const EdgeInsets.all(8),
                child: Column(
                  children: [
                    Text(
                      client.address.toString(),
                      maxLines: 2,
                      overflow: TextOverflow.ellipsis,
                    ),
                    if (client.user != null)
                      Text(
                        client.user!.displayName.isEmpty
                            ? client.user!.username
                            : client.user!.displayName,
                      ),
                  ],
                ),
              ),
            if (client.busy)
              const LinearProgressIndicator(semanticsLabel: 'Loading'),
            if (client.error != null)
              Semantics(
                liveRegion: true,
                child: Padding(
                  padding: const EdgeInsets.all(12),
                  child: Text(
                    client.error!,
                    style: TextStyle(
                      color: Theme.of(context).colorScheme.error,
                    ),
                  ),
                ),
              ),
            Expanded(
              child: switch (client.stage) {
                ClientStage.server => _ServerForm(client: client),
                ClientStage.login => _LoginForm(
                  key: ValueKey(client.address.toString()),
                  client: client,
                ),
                ClientStage.workspaces => _Workspaces(client: client),
                ClientStage.agents => _Agents(
                  client: client,
                  showIdentifiers: preferences.showIdentifiers,
                ),
              },
            ),
          ],
        ),
      ),
    ),
  );
}

class _ServerForm extends StatefulWidget {
  const _ServerForm({required this.client, this.onSaved});
  final ClientContext client;
  final VoidCallback? onSaved;
  @override
  State<_ServerForm> createState() => _ServerFormState();
}

class _ServerFormState extends State<_ServerForm> {
  final _form = GlobalKey<FormState>();
  late final _address = TextEditingController(
    text: widget.client.address?.toString(),
  );
  @override
  void dispose() {
    _address.dispose();
    super.dispose();
  }

  Future<void> _submit() async {
    if (!_form.currentState!.validate()) return;
    FocusScope.of(context).unfocus();
    await widget.client.setServer(_address.text);
    if (mounted &&
        widget.client.address != null &&
        widget.client.error == null) {
      widget.onSaved?.call();
    }
  }

  @override
  Widget build(BuildContext context) => AnimatedBuilder(
    animation: widget.client,
    builder: (context, _) => ListView(
      padding: const EdgeInsets.all(24),
      children: [
        Text(
          'Connect to Butter',
          style: Theme.of(context).textTheme.headlineSmall,
        ),
        const SizedBox(height: 16),
        const Text(
          'Enter your one Butter server address. HTTPS is used by default.',
        ),
        Form(
          key: _form,
          child: Column(
            children: [
              TextFormField(
                key: const Key('server-address'),
                controller: _address,
                enabled: !widget.client.busy,
                keyboardType: TextInputType.url,
                autocorrect: false,
                decoration: const InputDecoration(
                  labelText: 'Server address',
                  hintText: 'https://butter.example.org',
                ),
                validator: (value) => value == null || value.trim().isEmpty
                    ? 'Enter a server address.'
                    : null,
                onFieldSubmitted: (_) => widget.client.busy ? null : _submit(),
              ),
              const SizedBox(height: 20),
              FilledButton(
                onPressed: widget.client.busy ? null : _submit,
                child: const Text('Verify and save'),
              ),
            ],
          ),
        ),
        if (widget.onSaved != null && widget.client.error != null)
          Padding(
            padding: const EdgeInsets.only(top: 16),
            child: Text(widget.client.error!),
          ),
      ],
    ),
  );
}

class _LoginForm extends StatefulWidget {
  const _LoginForm({super.key, required this.client});
  final ClientContext client;
  @override
  State<_LoginForm> createState() => _LoginFormState();
}

class _LoginFormState extends State<_LoginForm> {
  final _form = GlobalKey<FormState>();
  final _username = TextEditingController();
  final _password = TextEditingController();
  @override
  void dispose() {
    _username.dispose();
    _password.dispose();
    super.dispose();
  }

  Future<void> _submit() async {
    if (!_form.currentState!.validate()) return;
    final password = _password.text;
    _password.clear();
    FocusScope.of(context).unfocus();
    await widget.client.login(_username.text, password);
  }

  @override
  Widget build(BuildContext context) => ListView(
    padding: const EdgeInsets.all(24),
    children: [
      Text('Sign in', style: Theme.of(context).textTheme.headlineSmall),
      Form(
        key: _form,
        child: Column(
          children: [
            TextFormField(
              key: const Key('username'),
              controller: _username,
              enabled: !widget.client.busy,
              autofillHints: const [AutofillHints.username],
              decoration: const InputDecoration(labelText: 'Username'),
              validator: (value) => value == null || value.trim().isEmpty
                  ? 'Enter your username.'
                  : null,
            ),
            TextFormField(
              key: const Key('password'),
              controller: _password,
              enabled: !widget.client.busy,
              obscureText: true,
              autocorrect: false,
              enableSuggestions: false,
              autofillHints: const [AutofillHints.password],
              decoration: const InputDecoration(labelText: 'Password'),
              validator: (value) => value == null || value.isEmpty
                  ? 'Enter your password.'
                  : null,
              onFieldSubmitted: (_) => widget.client.busy ? null : _submit(),
            ),
            const SizedBox(height: 20),
            FilledButton(
              key: const Key('sign-in'),
              onPressed: widget.client.busy ? null : _submit,
              child: const Text('Sign in'),
            ),
          ],
        ),
      ),
      if (widget.client.oauthWaitingFor case final provider?)
        _OAuthWaiting(client: widget.client, provider: provider)
      else if (widget.client.oauthProviders.isNotEmpty)
        _OAuthProviders(client: widget.client),
    ],
  );
}

/// Sign-in through a provider's page in the system browser.
class _OAuthProviders extends StatelessWidget {
  const _OAuthProviders({required this.client});
  final ClientContext client;
  @override
  Widget build(BuildContext context) => Padding(
    padding: const EdgeInsets.only(top: 24),
    child: Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        const Divider(),
        const SizedBox(height: 8),
        for (final provider in client.oauthProviders)
          Padding(
            padding: const EdgeInsets.only(top: 8),
            child: OutlinedButton(
              key: Key('oauth-${provider.name}'),
              onPressed: client.busy
                  ? null
                  : () => client.beginOAuth(provider.name),
              child: Text(
                'Continue with ${provider.displayName.isEmpty ? provider.name : provider.displayName}',
              ),
            ),
          ),
      ],
    ),
  );
}

/// Shown while the browser leg is open; the App never sees the provider page.
class _OAuthWaiting extends StatelessWidget {
  const _OAuthWaiting({required this.client, required this.provider});
  final ClientContext client;
  final String provider;
  @override
  Widget build(BuildContext context) => Padding(
    padding: const EdgeInsets.only(top: 24),
    child: Semantics(
      liveRegion: true,
      child: Card(
        child: Padding(
          padding: const EdgeInsets.all(16),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(
                'Finish signing in with $provider in your browser.',
                style: Theme.of(context).textTheme.titleMedium,
              ),
              const SizedBox(height: 8),
              const Text('Butter opens again when you are done.'),
              const SizedBox(height: 8),
              Align(
                alignment: Alignment.centerRight,
                child: TextButton(
                  key: const Key('oauth-cancel'),
                  onPressed: client.cancelOAuth,
                  child: const Text('Cancel'),
                ),
              ),
            ],
          ),
        ),
      ),
    ),
  );
}

class _Workspaces extends StatelessWidget {
  const _Workspaces({required this.client});
  final ClientContext client;
  @override
  Widget build(BuildContext context) => ListView(
    padding: const EdgeInsets.all(16),
    children: [
      Text(
        'Choose Workspace',
        style: Theme.of(context).textTheme.headlineSmall,
      ),
      FilledButton.icon(
        onPressed: client.busy
            ? null
            : () => openManagementPage(
                context,
                client,
                WorkspaceFormPage(client: client),
              ),
        icon: const Icon(Icons.add),
        label: const Text('Create Workspace'),
      ),
      TextButton.icon(
        onPressed: client.busy ? null : client.refreshWorkspaces,
        icon: const Icon(Icons.refresh),
        label: const Text('Refresh Workspaces'),
      ),
      if (!client.busy && client.workspaces.isEmpty)
        const Text(
          'No accessible Workspaces. Ask a Workspace owner for access.',
        ),
      for (final workspace in client.workspaces)
        ListTile(
          title: Text(workspace.name),
          subtitle: Text(
            workspace.description.isEmpty
                ? workspace.id
                : workspace.description,
          ),
          trailing: IconButton(
            tooltip: 'Manage ${workspace.name}',
            icon: Icon(
              Icons.manage_accounts,
              semanticLabel: 'Manage ${workspace.name}',
            ),
            onPressed: client.busy
                ? null
                : () => openWorkspaceDetails(context, client, workspace.id),
          ),
          onTap: client.busy
              ? null
              : () => client.selectWorkspace(workspace.id),
        ),
    ],
  );
}

class _Agents extends StatelessWidget {
  const _Agents({required this.client, required this.showIdentifiers});
  final ClientContext client;
  final bool showIdentifiers;
  @override
  Widget build(BuildContext context) => ListView(
    padding: const EdgeInsets.all(16),
    children: [
      Text(
        '${client.workspace!.name} · Agents',
        style: Theme.of(context).textTheme.headlineSmall,
      ),
      Wrap(
        spacing: 12,
        children: [
          FilledButton.tonalIcon(
            key: const Key('open-chat'),
            onPressed: client.busy
                ? null
                : () => openManagementPage(
                    context,
                    client,
                    ThreadListPage(client: client),
                  ),
            icon: const Icon(Icons.chat_bubble_outline),
            label: const Text('Chat'),
          ),
          TextButton.icon(
            onPressed: client.busy
                ? null
                : () => openWorkspaceDetails(
                    context,
                    client,
                    client.workspace!.id,
                  ),
            icon: const Icon(Icons.manage_accounts),
            label: const Text('Manage Workspace'),
          ),
          TextButton.icon(
            onPressed: client.busy
                ? null
                : () => openModelProviders(context, client),
            icon: const Icon(Icons.memory),
            label: const Text('Model providers'),
          ),
          TextButton.icon(
            key: const Key('open-boxes'),
            onPressed: client.busy
                ? null
                : () => openButterBoxes(context, client),
            icon: const Icon(Icons.dns_outlined),
            label: const Text('ButterBoxes'),
          ),
          TextButton.icon(
            key: const Key('open-mcp'),
            onPressed: client.busy
                ? null
                : () => openMcpServers(context, client),
            icon: const Icon(Icons.extension_outlined),
            label: const Text('MCP servers'),
          ),
          TextButton.icon(
            key: const Key('open-files'),
            onPressed: client.busy
                ? null
                : () => openFileSpaces(context, client),
            icon: const Icon(Icons.folder_outlined),
            label: const Text('File spaces'),
          ),
          TextButton.icon(
            key: const Key('open-repo-binding'),
            onPressed: client.busy
                ? null
                : () => openRepoBinding(context, client),
            icon: const Icon(Icons.account_tree_outlined),
            label: const Text('Git repository'),
          ),
          TextButton.icon(
            key: const Key('open-sessions'),
            onPressed: client.busy ? null : () => openSessions(context, client),
            icon: const Icon(Icons.history),
            label: const Text('Sessions'),
          ),
          TextButton.icon(
            key: const Key('open-remote-agents'),
            onPressed: client.busy
                ? null
                : () => openRemoteAgents(context, client),
            icon: const Icon(Icons.hub_outlined),
            label: const Text('Remote Agents'),
          ),
          TextButton.icon(
            key: const Key('open-daemons'),
            onPressed: client.busy ? null : () => openDaemons(context, client),
            icon: const Icon(Icons.developer_board),
            label: const Text('Daemon execution'),
          ),
          TextButton.icon(
            onPressed: client.busy ? null : client.refreshWorkspaces,
            icon: const Icon(Icons.swap_horiz),
            label: const Text('Change Workspace'),
          ),
          FilledButton.icon(
            onPressed: client.busy
                ? null
                : () => openAgentEditor(context, client),
            icon: const Icon(Icons.add),
            label: const Text('Create Agent'),
          ),
          TextButton.icon(
            onPressed: client.busy ? null : client.refreshAgents,
            icon: const Icon(Icons.refresh),
            label: const Text('Refresh Agents'),
          ),
        ],
      ),
      if (!client.busy && client.agents.isEmpty && client.error == null)
        const Text('No Agents in this Workspace.'),
      for (final agent in client.agents)
        ListTile(
          title: Text(
            agent.displayName.isEmpty ? agent.name : agent.displayName,
          ),
          subtitle: Text(
            [
              if (showIdentifiers) agent.agentId,
              agent.description,
              agentStatusLabel(agent),
            ].where((s) => s.isNotEmpty).join('\n'),
          ),
          onTap: client.busy
              ? null
              : () => openAgentDetail(context, client, agent.agentId),
        ),
      if (client.hasMoreAgents)
        FilledButton(
          onPressed: client.busy ? null : client.moreAgents,
          child: const Text('Load more Agents'),
        ),
    ],
  );
}
