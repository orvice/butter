import 'dart:async';

import 'package:flutter/material.dart';

import 'client_context.dart';

class ButterApp extends StatefulWidget {
  const ButterApp({super.key, required this.context});
  final ClientContext context;
  @override
  State<ButterApp> createState() => _ButterAppState();
}

class _ButterAppState extends State<ButterApp> with WidgetsBindingObserver {
  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addObserver(this);
  }

  @override
  void dispose() {
    WidgetsBinding.instance.removeObserver(this);
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
  Widget build(BuildContext context) => MaterialApp(
    title: 'Butter Dev',
    theme: ThemeData(
      colorSchemeSeed: Colors.amber,
      brightness: Brightness.light,
    ),
    darkTheme: ThemeData(
      colorSchemeSeed: Colors.amber,
      brightness: Brightness.dark,
    ),
    home: _Home(client: widget.context),
  );
}

class _Home extends StatelessWidget {
  const _Home({required this.client});
  final ClientContext client;
  @override
  Widget build(BuildContext context) => AnimatedBuilder(
    animation: client,
    builder: (context, _) => Scaffold(
      appBar: AppBar(
        title: const Text('Butter Dev'),
        actions: [
          if (client.address != null)
            IconButton(
              tooltip: 'Change server',
              icon: const Icon(Icons.settings),
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
              icon: const Icon(Icons.logout),
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
                ClientStage.agents => _Agents(client: client),
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
    ],
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
          trailing: const Icon(Icons.chevron_right),
          onTap: client.busy
              ? null
              : () => client.selectWorkspace(workspace.id),
        ),
    ],
  );
}

class _Agents extends StatelessWidget {
  const _Agents({required this.client});
  final ClientContext client;
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
          TextButton.icon(
            onPressed: client.busy ? null : client.refreshWorkspaces,
            icon: const Icon(Icons.swap_horiz),
            label: const Text('Change Workspace'),
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
              agent.agentId,
              agent.description,
            ].where((s) => s.isNotEmpty).join('\n'),
          ),
        ),
      if (client.hasMoreAgents)
        FilledButton(
          onPressed: client.busy ? null : client.moreAgents,
          child: const Text('Load more Agents'),
        ),
    ],
  );
}
