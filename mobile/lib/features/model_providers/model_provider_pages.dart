import 'package:flutter/material.dart';

import '../../app/client_context.dart';
import '../../core/management_page.dart';
import '../../gen/agents/v1/agent.pb.dart';
import 'provider_draft.dart';

const contextWindowExplanation =
    'Context window is how many tokens of input a model can take, including '
    'the conversation so far. It is not the maximum output tokens. This '
    'provider has no setting for output length.';

String _modelCount(int count) => count == 1 ? '1 model' : '$count models';

String _typeLabel(String type) => switch (type) {
  'openai' => 'OpenAI',
  'gemini' => 'Gemini',
  '' => 'No type',
  _ => type,
};

/// Loads the selected Workspace's providers, then opens the list. The list
/// itself offers a reload when this read failed.
Future<void> openModelProviders(
  BuildContext context,
  ClientContext client,
) async {
  final loading = client.loadModelProviders();
  final scope = client.scope;
  await loading;
  if (context.mounted && client.scope == scope && client.user != null) {
    await openManagementPage(
      context,
      client,
      ModelProviderListPage(client: client),
    );
  }
}

Future<void> _openDetail(
  BuildContext context,
  ClientContext client,
  String name,
) async {
  final loading = client.loadModelProvider(name);
  final scope = client.scope;
  await loading;
  if (context.mounted && client.scope == scope && client.user != null) {
    await openManagementPage(
      context,
      client,
      ModelProviderDetailPage(client: client, name: name),
    );
  }
}

class ModelProviderListPage extends StatelessWidget {
  const ModelProviderListPage({super.key, required this.client});
  final ClientContext client;

  @override
  Widget build(BuildContext context) => ManagementPage(
    client: client,
    title: 'Model providers',
    child: AnimatedBuilder(
      animation: client,
      builder: (context, _) {
        final providers = client.modelProviders;
        return ListView(
          padding: const EdgeInsets.all(16),
          children: [
            Text(
              '${client.workspace?.name ?? 'Workspace'} · Model providers',
              style: Theme.of(context).textTheme.headlineSmall,
            ),
            const SizedBox(height: 8),
            const Text(
              'A provider connects this Workspace to an LLM service and names the models its Agents can use.',
            ),
            const SizedBox(height: 12),
            Wrap(
              spacing: 12,
              children: [
                FilledButton.icon(
                  onPressed: client.busy
                      ? null
                      : () => openManagementPage(
                          context,
                          client,
                          ModelProviderFormPage(client: client),
                        ),
                  icon: const Icon(Icons.add),
                  label: const Text('Add provider'),
                ),
                TextButton.icon(
                  onPressed: client.busy ? null : client.loadModelProviders,
                  icon: const Icon(Icons.refresh),
                  label: Text(
                    client.modelProvidersLoaded
                        ? 'Refresh providers'
                        : 'Load providers',
                  ),
                ),
              ],
            ),
            if (client.modelProvidersLoaded && providers.isEmpty) ...[
              const SizedBox(height: 24),
              Text(
                'No model providers yet',
                style: Theme.of(context).textTheme.titleMedium,
              ),
              const Text(
                'Add a provider to make its models available to Agents in this Workspace.',
              ),
            ],
            for (final provider in providers)
              ListTile(
                isThreeLine: true,
                contentPadding: EdgeInsets.zero,
                title: Text(provider.name),
                subtitle: Text(
                  [
                    '${_typeLabel(provider.type)} · ${_modelCount(provider.models.length)}',
                    provider.baseUrl.isEmpty
                        ? 'Default provider endpoint'
                        : provider.baseUrl,
                    provider.apiKey.isEmpty ? 'No API key' : 'API key set',
                  ].join('\n'),
                ),
                trailing: const Icon(Icons.chevron_right),
                onTap: client.busy
                    ? null
                    : () => _openDetail(context, client, provider.name),
              ),
          ],
        );
      },
    ),
  );
}

class ModelProviderDetailPage extends StatelessWidget {
  const ModelProviderDetailPage({
    super.key,
    required this.client,
    required this.name,
  });
  final ClientContext client;
  final String name;

  @override
  Widget build(BuildContext context) => ManagementPage(
    client: client,
    title: 'Model provider',
    child: AnimatedBuilder(
      animation: client,
      builder: (context, _) {
        final provider = client.modelProvider;
        if (provider == null || provider.name != name) {
          return Center(
            child: TextButton(
              onPressed: client.busy
                  ? null
                  : () => client.loadModelProvider(name),
              child: const Text('Load provider'),
            ),
          );
        }
        final theme = Theme.of(context);
        return ListView(
          padding: const EdgeInsets.all(16),
          children: [
            Semantics(
              header: true,
              child: SelectableText(
                provider.name,
                style: theme.textTheme.headlineSmall,
              ),
            ),
            Text('Type: ${_typeLabel(provider.type)}'),
            Text(
              provider.baseUrl.isEmpty
                  ? 'Base URL: Default provider endpoint'
                  : 'Base URL: ${provider.baseUrl}',
            ),
            // The key itself is never shown, only whether one is stored.
            Text(provider.apiKey.isEmpty ? 'API key: Not set' : 'API key: Set'),
            const SizedBox(height: 8),
            Wrap(
              spacing: 12,
              children: [
                FilledButton(
                  onPressed: client.busy
                      ? null
                      : () => openManagementPage(
                          context,
                          client,
                          ModelProviderFormPage(
                            client: client,
                            provider: provider,
                          ),
                        ),
                  child: const Text('Edit provider'),
                ),
                OutlinedButton(
                  onPressed: client.busy
                      ? null
                      : () => _confirmDelete(context, provider),
                  child: const Text('Delete provider'),
                ),
                TextButton.icon(
                  onPressed: client.busy
                      ? null
                      : () => client.loadModelProvider(name),
                  icon: const Icon(Icons.refresh),
                  label: const Text('Refresh provider'),
                ),
              ],
            ),
            const SizedBox(height: 24),
            Semantics(
              header: true,
              child: Text('Models', style: theme.textTheme.titleLarge),
            ),
            const Text(contextWindowExplanation),
            if (provider.models.isEmpty) const Text('No models.'),
            for (final (index, model) in provider.models.indexed)
              Semantics(
                container: true,
                label: 'Model ${index + 1}',
                child: Padding(
                  padding: const EdgeInsets.symmetric(vertical: 8),
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      SelectableText(
                        'Model ID: ${model.name}',
                        style: theme.textTheme.titleSmall,
                      ),
                      Text(
                        model.alias.isEmpty
                            ? 'Alias: None'
                            : 'Alias: ${model.alias}',
                      ),
                      Text(
                        model.contextWindowTokens == 0
                            ? 'Context window: Not set. The server uses the model’s built-in metadata.'
                            : 'Context window: ${formatTokens(model.contextWindowTokens)} tokens',
                      ),
                    ],
                  ),
                ),
              ),
          ],
        );
      },
    ),
  );

  Future<void> _confirmDelete(
    BuildContext context,
    ModelProvider provider,
  ) async {
    final aliases = provider.models.length;
    final confirmed = await confirmChange(
      context,
      'Delete ${provider.name}?',
      'Removes this provider and its ${_modelCount(aliases)} from the Workspace. '
          'Agents and channels that use these models may stop working. '
          'Telegram destinations that use them keep the provider from being deleted. '
          'This cannot be undone. A lost reply does not mean nothing was deleted.',
      action: 'Delete provider',
    );
    if (!confirmed || !context.mounted) return;
    final ok = await client.deleteModelProvider(provider.name);
    if (ok && context.mounted) Navigator.pop(context);
  }
}

/// One editable model row. It owns its controllers and the server's message
/// for that row, so removing a row never moves unknown fields to another.
class _Row {
  _Row([this.original])
    : name = TextEditingController(text: original?.name),
      alias = TextEditingController(text: original?.alias),
      window = TextEditingController(
        text: (original?.contextWindowTokens ?? 0) > 0
            ? '${original!.contextWindowTokens}'
            : '',
      );
  final ModelConfig? original;
  final TextEditingController name;
  final TextEditingController alias;
  final TextEditingController window;
  final nameFocus = FocusNode();
  final aliasFocus = FocusNode();
  final windowFocus = FocusNode();

  ModelRowInput get input => ModelRowInput(
    original: original,
    name: name.text,
    alias: alias.text,
    contextWindow: window.text,
  );

  void dispose() {
    name.dispose();
    alias.dispose();
    window.dispose();
    nameFocus.dispose();
    aliasFocus.dispose();
    windowFocus.dispose();
  }
}

class ModelProviderFormPage extends StatefulWidget {
  const ModelProviderFormPage({super.key, required this.client, this.provider});
  final ClientContext client;

  /// The provider as the server returned it; null creates a new one.
  final ModelProvider? provider;
  @override
  State<ModelProviderFormPage> createState() => _ModelProviderFormPageState();
}

class _ModelProviderFormPageState extends State<ModelProviderFormPage> {
  final _form = GlobalKey<FormState>();
  final _nameFocus = FocusNode();
  final _urlFocus = FocusNode();
  final _keyFocus = FocusNode();
  final _addFocus = FocusNode();
  late final _name = TextEditingController(text: widget.provider?.name);
  late final _baseUrl = TextEditingController(text: widget.provider?.baseUrl);
  final _newKey = TextEditingController();
  late String _type = (widget.provider?.type.isEmpty ?? true)
      ? 'openai'
      : widget.provider!.type;
  KeyAction _keyAction = KeyAction.keep;
  late final List<_Row> _rows = [
    for (final model in widget.provider?.models ?? const <ModelConfig>[])
      _Row(model),
    if (widget.provider == null) _Row(),
  ];
  bool _dirty = false;
  String? _modelsError;

  bool get _editing => widget.provider != null;
  bool get _hasKey => widget.provider?.apiKey.isNotEmpty ?? false;
  bool get _entersKey => !_editing || _keyAction == KeyAction.replace;

  @override
  void dispose() {
    _name.dispose();
    _baseUrl.dispose();
    _newKey.dispose();
    for (final row in _rows) {
      row.dispose();
    }
    _nameFocus.dispose();
    _urlFocus.dispose();
    _keyFocus.dispose();
    _addFocus.dispose();
    super.dispose();
  }

  void _touch(VoidCallback change) => setState(() {
    change();
    _dirty = true;
  });

  /// Focus goes to the first field that fails, in reading order.
  void _focusFirstInvalid() {
    final checks = <(FocusNode, String? Function())>[
      if (!_editing) (_nameFocus, () => validateProviderName(_name.text)),
      (_urlFocus, () => validateBaseUrl(_baseUrl.text)),
      if (_keyAction == KeyAction.replace && _editing)
        (
          _keyFocus,
          () => _newKey.text.trim().isEmpty ? 'Enter the new API key.' : null,
        ),
      for (final row in _rows) ...[
        (row.nameFocus, () => validateModelId(row.name.text)),
        (row.windowFocus, () => validateContextWindow(row.window.text)),
      ],
    ];
    for (final (node, check) in checks) {
      if (check() != null) {
        node.requestFocus();
        return;
      }
    }
  }

  Future<void> _submit() async {
    final valid = _form.currentState!.validate();
    final noModels = _rows.isEmpty;
    setState(
      () => _modelsError = noModels
          ? 'Add at least one model to this provider.'
          : null,
    );
    if (!valid || noModels) {
      if (!valid) {
        _focusFirstInvalid();
      } else {
        _addFocus.requestFocus();
      }
      return;
    }
    final provider = widget.provider;
    if (_keyAction == KeyAction.clear) {
      final confirmed = await confirmChange(
        context,
        'Remove the API key?',
        'The key is deleted from ${provider!.name} and cannot be read back in this app. Agents that use this provider will fail until a key is set again, unless the server supplies one.',
        action: 'Remove key and save',
      );
      if (!confirmed || !mounted) return;
    }
    FocusScope.of(context).unfocus();
    final draft = buildModelProvider(
      base: provider,
      name: _name.text,
      type: _type,
      baseUrl: _baseUrl.text,
      key: _editing
          ? _keyAction
          : _newKey.text.trim().isNotEmpty
          ? KeyAction.replace
          : KeyAction.keep,
      newKey: _newKey.text,
      rows: [for (final row in _rows) row.input],
    );
    final messenger = ScaffoldMessenger.of(context);
    final navigator = Navigator.of(context);
    final ok = await widget.client.saveModelProvider(draft, create: !_editing);
    if (!ok || !mounted) return;
    setState(() {
      _dirty = false;
      _newKey.clear();
    });
    messenger.showSnackBar(SnackBar(content: Text('${draft.name} saved.')));
    navigator.pop();
  }

  @override
  Widget build(BuildContext context) {
    final client = widget.client;
    final theme = Theme.of(context);
    return ManagementPage(
      client: client,
      title: _editing ? 'Edit provider' : 'Add provider',
      child: AnimatedBuilder(
        animation: client,
        builder: (context, _) => DraftGuard(
          dirty: _dirty,
          busy: client.busy,
          child: Form(
            key: _form,
            onChanged: () => setState(() => _dirty = true),
            child: ListView(
              padding: const EdgeInsets.all(24),
              keyboardDismissBehavior: ScrollViewKeyboardDismissBehavior.onDrag,
              children: [
                TextFormField(
                  controller: _name,
                  focusNode: _nameFocus,
                  readOnly: _editing,
                  enabled: !client.busy,
                  autocorrect: false,
                  decoration: InputDecoration(
                    labelText: 'Name',
                    helperText: _editing
                        ? 'The name cannot be changed. Create a new provider to use another name.'
                        : 'Identifies this provider in the Workspace, for example openai.',
                  ),
                  validator: _editing ? null : validateProviderName,
                ),
                const SizedBox(height: 16),
                DropdownButtonFormField<String>(
                  initialValue: _type,
                  decoration: const InputDecoration(
                    labelText: 'Type',
                    helperText:
                        'Choose OpenAI for OpenAI-compatible services and set a base URL.',
                  ),
                  items: [
                    for (final type in {...providerTypes, _type})
                      DropdownMenuItem(
                        value: type,
                        child: Text(
                          providerTypes.contains(type)
                              ? _typeLabel(type)
                              : '$type (not supported by this app)',
                        ),
                      ),
                  ],
                  onChanged: client.busy
                      ? null
                      : (value) => _touch(() => _type = value ?? _type),
                ),
                const SizedBox(height: 16),
                TextFormField(
                  controller: _baseUrl,
                  focusNode: _urlFocus,
                  enabled: !client.busy,
                  keyboardType: TextInputType.url,
                  autocorrect: false,
                  decoration: const InputDecoration(
                    labelText: 'Base URL',
                    hintText: 'https://api.openai.com/v1',
                    helperText:
                        'Optional. Leave empty to use the provider’s default endpoint.',
                  ),
                  validator: validateBaseUrl,
                ),
                const SizedBox(height: 24),
                Semantics(
                  header: true,
                  child: Text('API key', style: theme.textTheme.titleMedium),
                ),
                if (_editing) ...[
                  Text(
                    _hasKey
                        ? 'A key is set for this provider. This app never shows saved keys.'
                        : 'No key is set for this provider.',
                  ),
                  RadioGroup<KeyAction>(
                    groupValue: _keyAction,
                    onChanged: (value) {
                      if (client.busy || value == null) return;
                      _touch(() {
                        _keyAction = value;
                        if (value != KeyAction.replace) _newKey.clear();
                      });
                    },
                    child: Column(
                      children: [
                        RadioListTile<KeyAction>(
                          contentPadding: EdgeInsets.zero,
                          value: KeyAction.keep,
                          title: Text(
                            _hasKey ? 'Keep current key' : 'Keep without a key',
                          ),
                        ),
                        RadioListTile<KeyAction>(
                          contentPadding: EdgeInsets.zero,
                          value: KeyAction.replace,
                          title: Text(_hasKey ? 'Replace key' : 'Set a key'),
                        ),
                        if (_hasKey)
                          const RadioListTile<KeyAction>(
                            contentPadding: EdgeInsets.zero,
                            value: KeyAction.clear,
                            title: Text('Remove key'),
                          ),
                      ],
                    ),
                  ),
                ],
                if (_entersKey)
                  TextFormField(
                    controller: _newKey,
                    focusNode: _keyFocus,
                    enabled: !client.busy,
                    obscureText: true,
                    autocorrect: false,
                    enableSuggestions: false,
                    enableIMEPersonalizedLearning: false,
                    decoration: InputDecoration(
                      labelText: _editing ? 'New API key' : 'API key',
                      helperText: _editing
                          ? 'Sent to the server when you save.'
                          : 'Optional. Leave empty if the server supplies the key.',
                    ),
                    validator: (value) =>
                        _editing && (value ?? '').trim().isEmpty
                        ? 'Enter the new API key.'
                        : null,
                  ),
                const SizedBox(height: 24),
                Semantics(
                  header: true,
                  child: Text('Models', style: theme.textTheme.titleMedium),
                ),
                const Text(
                  'Model ID is the name the provider uses. Agents and channels can refer to a model by its alias instead.',
                ),
                const SizedBox(height: 4),
                const Text(contextWindowExplanation),
                for (final (index, row) in _rows.indexed)
                  Semantics(
                    key: ObjectKey(row),
                    container: true,
                    explicitChildNodes: true,
                    label: 'Model ${index + 1}',
                    child: Card(
                      margin: const EdgeInsets.only(top: 12),
                      child: Padding(
                        padding: const EdgeInsets.all(12),
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Row(
                              children: [
                                Expanded(
                                  child: Text(
                                    'Model ${index + 1}',
                                    style: theme.textTheme.titleSmall,
                                  ),
                                ),
                                IconButton(
                                  tooltip: row.name.text.trim().isEmpty
                                      ? 'Remove model ${index + 1}'
                                      : 'Remove model ${row.name.text.trim()}',
                                  icon: const Icon(Icons.delete_outline),
                                  onPressed: client.busy
                                      ? null
                                      : () => _touch(() {
                                          _rows.remove(row);
                                          // Disposed after the frame that
                                          // stops using its controllers.
                                          WidgetsBinding.instance
                                              .addPostFrameCallback(
                                                (_) => row.dispose(),
                                              );
                                        }),
                                ),
                              ],
                            ),
                            TextFormField(
                              controller: row.name,
                              focusNode: row.nameFocus,
                              enabled: !client.busy,
                              autocorrect: false,
                              decoration: const InputDecoration(
                                labelText: 'Model ID',
                                hintText: 'gpt-4o',
                              ),
                              validator: validateModelId,
                            ),
                            const SizedBox(height: 12),
                            TextFormField(
                              controller: row.alias,
                              focusNode: row.aliasFocus,
                              enabled: !client.busy,
                              autocorrect: false,
                              decoration: const InputDecoration(
                                labelText: 'Alias',
                                helperText: 'Optional.',
                              ),
                            ),
                            const SizedBox(height: 12),
                            TextFormField(
                              controller: row.window,
                              focusNode: row.windowFocus,
                              enabled: !client.busy,
                              keyboardType: TextInputType.number,
                              decoration: const InputDecoration(
                                labelText: 'Context window (tokens)',
                                helperText:
                                    'Optional. Empty or 0 uses the model’s built-in metadata.',
                              ),
                              validator: validateContextWindow,
                            ),
                          ],
                        ),
                      ),
                    ),
                  ),
                if (_modelsError != null)
                  Padding(
                    padding: const EdgeInsets.only(top: 8),
                    child: Semantics(
                      liveRegion: true,
                      child: Text(
                        _modelsError!,
                        style: TextStyle(color: theme.colorScheme.error),
                      ),
                    ),
                  ),
                Align(
                  alignment: AlignmentDirectional.centerStart,
                  child: TextButton.icon(
                    focusNode: _addFocus,
                    onPressed: client.busy
                        ? null
                        : () => _touch(() {
                            _rows.add(_Row());
                            _modelsError = null;
                          }),
                    icon: const Icon(Icons.add),
                    label: const Text('Add model'),
                  ),
                ),
                const SizedBox(height: 16),
                if (_editing)
                  const Text(
                    'Saving replaces this provider with the values loaded when you opened it. The server has no concurrent-edit protection, so refresh first if someone else may have changed it.',
                  ),
                const SizedBox(height: 12),
                FilledButton(
                  onPressed: client.busy ? null : _submit,
                  child: Text(_editing ? 'Save changes' : 'Create provider'),
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }
}
