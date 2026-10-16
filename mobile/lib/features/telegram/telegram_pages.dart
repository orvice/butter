import 'dart:async';

import 'package:flutter/material.dart';
import 'package:flutter/services.dart';

import '../../app/client_context.dart';
import '../../core/management_page.dart';
import '../../gen/agents/v1/agent.pb.dart';
import '../../gen/agents/v1/telegram.pb.dart';
import 'telegram_controller.dart';
import 'telegram_draft.dart';
import 'telegram_labels.dart';

Future<void> openTelegram(BuildContext context, ClientContext client) async {
  final tg = client.telegram;
  if (tg == null) return;
  unawaited(tg.loadChannels());
  unawaited(tg.loadPlatformSettings());
  await openManagementPage(context, client, TelegramHubPage(client: client));
}

class TelegramHubPage extends StatelessWidget {
  const TelegramHubPage({super.key, required this.client});
  final ClientContext client;

  @override
  Widget build(BuildContext context) => Scaffold(
    appBar: AppBar(title: const Text('Telegram')),
    body: ListView(
      children: [
        ListTile(
          leading: const Icon(Icons.router_outlined),
          title: const Text('Channels'),
          subtitle: const Text('Bot transports, tokens and receive mode'),
          onTap: () => Navigator.push<void>(
            context,
            MaterialPageRoute(
              builder: (_) => TelegramChannelListPage(client: client),
            ),
          ),
        ),
        ListTile(
          leading: const Icon(Icons.history),
          title: const Text('Processing records'),
          subtitle: const Text('Inbound updates and delivery recovery'),
          onTap: () => Navigator.push<void>(
            context,
            MaterialPageRoute(
              builder: (_) => TelegramProcessingListPage(client: client),
            ),
          ),
        ),
      ],
    ),
  );
}

mixin _TelegramPage<T extends StatefulWidget> on State<T> {
  TelegramController? get tg => (widget as dynamic).client.telegram;

  @override
  void initState() {
    super.initState();
    tg?.addListener(_onTelegramChange);
  }

  @override
  void dispose() {
    tg?.removeListener(_onTelegramChange);
    super.dispose();
  }

  void _onTelegramChange() {
    if (mounted) setState(() {});
  }

  Widget writeNotice(TelegramController c) {
    if (c.writeError == null) return const SizedBox.shrink();
    return Semantics(
      liveRegion: true,
      child: Padding(
        padding: const EdgeInsets.all(12),
        child: Text(
          c.writeError!,
          style: TextStyle(color: Theme.of(context).colorScheme.error),
        ),
      ),
    );
  }

  bool get canManage => (widget as dynamic).client.canManageWorkspaceResources;
}

class TelegramChannelListPage extends StatefulWidget {
  const TelegramChannelListPage({super.key, required this.client});
  final ClientContext client;

  @override
  State<TelegramChannelListPage> createState() =>
      _TelegramChannelListPageState();
}

class _TelegramChannelListPageState extends State<TelegramChannelListPage>
    with _TelegramPage {
  @override
  Widget build(BuildContext context) {
    final c = tg;
    return Scaffold(
      appBar: AppBar(
        title: const Text('Telegram channels'),
        actions: [
          if (canManage)
            IconButton(
              tooltip: 'Add channel',
              icon: const Icon(Icons.add),
              onPressed: c == null || c.writeBlocked
                  ? null
                  : () => Navigator.push<void>(
                      context,
                      MaterialPageRoute(
                        builder: (_) =>
                            TelegramChannelCreatePage(client: widget.client),
                      ),
                    ),
            ),
        ],
      ),
      body: c == null
          ? const Center(child: Text('Sign in and choose a Workspace.'))
          : ListView(
              padding: const EdgeInsets.all(16),
              children: [
                writeNotice(c),
                if (c.platformSettingsDenied &&
                    (c.platformWebhookBase == null ||
                        c.platformWebhookBase!.isEmpty))
                  const Padding(
                    padding: EdgeInsets.only(bottom: 12),
                    child: Text(
                      'Public webhook base URL is not configured. '
                      'A platform administrator must set it before Webhook mode can receive updates.',
                    ),
                  ),
                if (c.channelListState == TelegramLoad.loading)
                  const LinearProgressIndicator(),
                if (c.channelListError != null) Text(c.channelListError!),
                for (final ch in c.channels)
                  ListTile(
                    title: Text(ch.name.isEmpty ? ch.key : ch.name),
                    subtitle: Text(
                      '@${ch.botUsername} · ${credentialStateLabel(ch.credentialState)}',
                    ),
                    onTap: () => Navigator.push<void>(
                      context,
                      MaterialPageRoute(
                        builder: (_) => TelegramChannelDetailPage(
                          client: widget.client,
                          channelId: ch.id,
                        ),
                      ),
                    ),
                  ),
                if (c.channelListState == TelegramLoad.ready &&
                    c.channels.isEmpty)
                  const Text('No Telegram channels yet.'),
              ],
            ),
    );
  }
}

class TelegramChannelCreatePage extends StatefulWidget {
  const TelegramChannelCreatePage({super.key, required this.client});
  final ClientContext client;

  @override
  State<TelegramChannelCreatePage> createState() =>
      _TelegramChannelCreatePageState();
}

class _TelegramChannelCreatePageState extends State<TelegramChannelCreatePage>
    with _TelegramPage {
  final _formKey = GlobalKey<FormState>();
  late final _draft = TelegramChannelDraft.create();
  final _token = TextEditingController();

  @override
  void dispose() {
    _token.dispose();
    super.dispose();
  }

  Future<void> _save() async {
    final c = tg;
    if (c == null || !_formKey.currentState!.validate()) return;
    final err = _draft.validateCreate();
    if (err != null) {
      ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text(err)));
      return;
    }
    final token = _token.text;
    _token.clear();
    final ok = await c.createChannel(_draft, token);
    if (!mounted) return;
    if (ok && c.channel != null) {
      Navigator.pop(context);
      Navigator.push<void>(
        context,
        MaterialPageRoute(
          builder: (_) => TelegramChannelDetailPage(
            client: widget.client,
            channelId: c.channel!.id,
          ),
        ),
      );
    }
  }

  @override
  Widget build(BuildContext context) {
    final c = tg;
    return Scaffold(
      appBar: AppBar(title: const Text('Add Telegram channel')),
      body: c == null
          ? const Center(child: Text('Unavailable.'))
          : Form(
              key: _formKey,
              child: ListView(
                padding: const EdgeInsets.all(16),
                children: [
                  writeNotice(c),
                  TextFormField(
                    decoration: const InputDecoration(labelText: 'Key'),
                    initialValue: _draft.key,
                    onChanged: (v) => _draft.key = v,
                    validator: (v) =>
                        v == null || v.trim().isEmpty ? 'Required.' : null,
                  ),
                  TextFormField(
                    decoration: const InputDecoration(labelText: 'Name'),
                    initialValue: _draft.name,
                    onChanged: (v) => _draft.name = v,
                  ),
                  DropdownButtonFormField<TelegramReceiveMode>(
                    initialValue: _draft.receiveMode,
                    decoration: const InputDecoration(
                      labelText: 'Receive mode',
                    ),
                    items: const [
                      DropdownMenuItem(
                        value:
                            TelegramReceiveMode.TELEGRAM_RECEIVE_MODE_WEBHOOK,
                        child: Text('Webhook'),
                      ),
                      DropdownMenuItem(
                        value: TelegramReceiveMode
                            .TELEGRAM_RECEIVE_MODE_LONG_POLLING,
                        child: Text('Long polling'),
                      ),
                    ],
                    onChanged: (v) => setState(
                      () => _draft.receiveMode =
                          v ??
                          TelegramReceiveMode.TELEGRAM_RECEIVE_MODE_WEBHOOK,
                    ),
                  ),
                  TextFormField(
                    controller: _token,
                    obscureText: true,
                    enableSuggestions: false,
                    autocorrect: false,
                    decoration: const InputDecoration(
                      labelText: 'Bot token',
                      helperText: 'Write-only; never stored in the app.',
                    ),
                    validator: (v) =>
                        v == null || v.trim().isEmpty ? 'Required.' : null,
                  ),
                  const SizedBox(height: 16),
                  FilledButton(
                    onPressed: c.writeBlocked ? null : _save,
                    child: const Text('Create channel'),
                  ),
                ],
              ),
            ),
    );
  }
}

class TelegramChannelDetailPage extends StatefulWidget {
  const TelegramChannelDetailPage({
    super.key,
    required this.client,
    required this.channelId,
  });
  final ClientContext client;
  final String channelId;

  @override
  State<TelegramChannelDetailPage> createState() =>
      _TelegramChannelDetailPageState();
}

class _TelegramChannelDetailPageState extends State<TelegramChannelDetailPage>
    with _TelegramPage {
  TelegramChannelDraft? _draft;
  final _rotateToken = TextEditingController();

  @override
  void initState() {
    super.initState();
    unawaited(_load());
  }

  Future<void> _load() async {
    final c = tg;
    if (c == null) return;
    await c.loadChannel(widget.channelId);
    await c.loadChannelStatus(widget.channelId);
    await c.loadDestinations(channelId: widget.channelId);
    if (mounted && c.channel != null) {
      setState(() => _draft = TelegramChannelDraft.edit(c.channel!));
    }
  }

  @override
  void dispose() {
    _rotateToken.dispose();
    super.dispose();
  }

  Future<void> _saveSettings() async {
    final c = tg;
    final draft = _draft;
    if (c == null || draft == null) return;
    if (await c.updateChannel(draft) && mounted) {
      setState(() => _draft = TelegramChannelDraft.edit(c.channel!));
      ScaffoldMessenger.of(
        context,
      ).showSnackBar(const SnackBar(content: Text('Channel saved')));
    }
  }

  Future<void> _rotate() async {
    final c = tg;
    if (c == null || _rotateToken.text.trim().isEmpty) return;
    final token = _rotateToken.text;
    _rotateToken.clear();
    if (await c.rotateChannelToken(widget.channelId, token) && mounted) {
      setState(() => _draft = TelegramChannelDraft.edit(c.channel!));
      ScaffoldMessenger.of(
        context,
      ).showSnackBar(const SnackBar(content: Text('Bot token rotated')));
    }
  }

  Future<void> _toggle(bool inbound, bool outbound) async {
    final c = tg;
    final ch = c?.channel;
    if (c == null || ch == null) return;
    final result = await c.setChannelEnabled(
      channelId: ch.id,
      revision: ch.revision,
      inbound: inbound,
      outbound: outbound,
    );
    if (!mounted) return;
    if (result.ok) {
      setState(() => _draft = TelegramChannelDraft.edit(c.channel!));
      for (final w in result.warnings) {
        ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text(w)));
      }
    }
  }

  Future<void> _delete() async {
    final c = tg;
    if (c == null) return;
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (ctx) => AlertDialog(
        title: const Text('Delete channel?'),
        content: Text(
          'Removes ${c.channel?.name ?? widget.channelId} and its stored token.',
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
    if (confirmed != true || !mounted) return;
    if (await c.deleteChannel(widget.channelId) && mounted) {
      Navigator.pop(context);
    }
  }

  @override
  Widget build(BuildContext context) {
    final c = tg;
    final ch = c?.channel;
    final st = c?.channelStatus;
    final draft = _draft;
    return Scaffold(
      appBar: AppBar(
        title: Text(ch?.name.isEmpty ?? true ? 'Channel' : ch!.name),
        actions: [
          if (canManage && ch != null)
            IconButton(
              tooltip: 'Delete channel',
              icon: const Icon(Icons.delete_outline),
              onPressed: c!.writeBlocked ? null : _delete,
            ),
        ],
      ),
      body: c == null || ch == null || draft == null
          ? const Center(child: CircularProgressIndicator())
          : ListView(
              padding: const EdgeInsets.all(16),
              children: [
                writeNotice(c),
                Text('@${ch.botUsername} · bot ${ch.botId}'),
                Text(credentialStateLabel(ch.credentialState)),
                if (st != null) ...[
                  const SizedBox(height: 8),
                  Text(receiveModeLabel(st.receiveMode)),
                  if (st.webhookUrl.isNotEmpty)
                    SelectableText('Callback: ${st.webhookUrl}'),
                  if (st.webhookUrl.isEmpty &&
                      ch.receiveMode ==
                          TelegramReceiveMode.TELEGRAM_RECEIVE_MODE_WEBHOOK)
                    const Text(
                      'No callback URL yet. If blockers mention the webhook base URL, '
                      'a platform administrator must configure it.',
                    ),
                  Text(webhookStateLabel(st.webhookState)),
                  for (final b in st.blockers)
                    Text(
                      b,
                      style: TextStyle(
                        color: Theme.of(context).colorScheme.error,
                      ),
                    ),
                  for (final w in st.warnings) Text(w),
                ],
                SwitchListTile(
                  title: const Text('Receive updates'),
                  value: ch.inboundEnabled,
                  onChanged: !canManage || c.writeBlocked
                      ? null
                      : (v) => _toggle(v, v ? true : ch.outboundEnabled),
                ),
                SwitchListTile(
                  title: const Text('Send messages'),
                  value: ch.outboundEnabled,
                  onChanged: !canManage || c.writeBlocked
                      ? null
                      : (v) => _toggle(v ? ch.inboundEnabled : false, v),
                ),
                if (canManage) ...[
                  const Divider(),
                  TextFormField(
                    initialValue: draft.name,
                    decoration: const InputDecoration(labelText: 'Name'),
                    onChanged: (v) => draft.name = v,
                  ),
                  DropdownButtonFormField<TelegramReceiveMode>(
                    initialValue: draft.receiveMode,
                    decoration: const InputDecoration(
                      labelText: 'Receive mode',
                    ),
                    items: const [
                      DropdownMenuItem(
                        value:
                            TelegramReceiveMode.TELEGRAM_RECEIVE_MODE_WEBHOOK,
                        child: Text('Webhook'),
                      ),
                      DropdownMenuItem(
                        value: TelegramReceiveMode
                            .TELEGRAM_RECEIVE_MODE_LONG_POLLING,
                        child: Text('Long polling'),
                      ),
                    ],
                    onChanged: c.writeBlocked
                        ? null
                        : (v) => setState(
                            () => draft.receiveMode =
                                v ??
                                TelegramReceiveMode
                                    .TELEGRAM_RECEIVE_MODE_WEBHOOK,
                          ),
                  ),
                  FilledButton(
                    onPressed: c.writeBlocked ? null : _saveSettings,
                    child: const Text('Save settings'),
                  ),
                  const SizedBox(height: 16),
                  TextField(
                    controller: _rotateToken,
                    obscureText: true,
                    decoration: const InputDecoration(
                      labelText: 'Replace Bot token',
                    ),
                  ),
                  OutlinedButton(
                    onPressed: c.writeBlocked ? null : _rotate,
                    child: const Text('Rotate token'),
                  ),
                ],
                const Divider(),
                Row(
                  mainAxisAlignment: MainAxisAlignment.spaceBetween,
                  children: [
                    Text(
                      'Destinations',
                      style: Theme.of(context).textTheme.titleMedium,
                    ),
                    if (canManage)
                      TextButton.icon(
                        onPressed: c.writeBlocked
                            ? null
                            : () => Navigator.push<void>(
                                context,
                                MaterialPageRoute(
                                  builder: (_) => TelegramDestinationFormPage(
                                    client: widget.client,
                                    channelId: widget.channelId,
                                  ),
                                ),
                              ),
                        icon: const Icon(Icons.add),
                        label: const Text('Add'),
                      ),
                  ],
                ),
                for (final d in c.destinations)
                  ListTile(
                    title: Text(d.name.isEmpty ? d.key : d.name),
                    subtitle: Text(formatAddress(d)),
                    onTap: () => Navigator.push<void>(
                      context,
                      MaterialPageRoute(
                        builder: (_) => TelegramDestinationDetailPage(
                          client: widget.client,
                          destinationId: d.id,
                        ),
                      ),
                    ),
                  ),
              ],
            ),
    );
  }
}

class TelegramDestinationFormPage extends StatefulWidget {
  const TelegramDestinationFormPage({
    super.key,
    required this.client,
    required this.channelId,
    this.destinationId,
  });
  final ClientContext client;
  final String channelId;
  final String? destinationId;

  @override
  State<TelegramDestinationFormPage> createState() =>
      _TelegramDestinationFormPageState();
}

class _TelegramDestinationFormPageState
    extends State<TelegramDestinationFormPage>
    with _TelegramPage {
  final _formKey = GlobalKey<FormState>();
  TelegramDestinationDraft? _draft;
  bool _loading = true;

  @override
  void initState() {
    super.initState();
    unawaited(_initDraft());
  }

  Future<void> _initDraft() async {
    final c = tg;
    if (c == null) return;
    if (widget.destinationId != null) {
      await c.loadDestination(widget.destinationId!);
      if (c.destination != null) {
        _draft = TelegramDestinationDraft.edit(c.destination!);
      }
    } else {
      _draft = TelegramDestinationDraft.create(channelId: widget.channelId);
    }
    if (mounted) setState(() => _loading = false);
  }

  Agent? _selectedAgent(TelegramDestinationDraft draft) {
    for (final a in widget.client.agents) {
      if (a.agentId == draft.agentId) return a;
    }
    return null;
  }

  Future<void> _save() async {
    final c = tg;
    final draft = _draft;
    if (c == null || draft == null || !_formKey.currentState!.validate()) {
      return;
    }
    final agent = _selectedAgent(draft);
    final owns = draft.agentOwnsModel(agent);
    final err = draft.validate(ownsModel: owns);
    if (err != null) {
      ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text(err)));
      return;
    }
    final ok = draft.isCreate
        ? await c.createDestination(draft, ownsModel: owns)
        : await c.updateDestination(draft, ownsModel: owns);
    if (!mounted) return;
    if (ok) {
      Navigator.pop(context);
      if (draft.isCreate && c.destination != null) {
        Navigator.push<void>(
          context,
          MaterialPageRoute(
            builder: (_) => TelegramDestinationDetailPage(
              client: widget.client,
              destinationId: c.destination!.id,
            ),
          ),
        );
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    final c = tg;
    final draft = _draft;
    final isEdit = widget.destinationId != null;
    return Scaffold(
      appBar: AppBar(
        title: Text(isEdit ? 'Edit destination' : 'Add destination'),
      ),
      body: _loading || c == null || draft == null
          ? const Center(child: CircularProgressIndicator())
          : Form(
              key: _formKey,
              child: ListView(
                padding: const EdgeInsets.all(16),
                children: [
                  writeNotice(c),
                  if (draft.isCreate)
                    TextFormField(
                      decoration: const InputDecoration(labelText: 'Key'),
                      onChanged: (v) => draft.key = v,
                      validator: (v) =>
                          v == null || v.trim().isEmpty ? 'Required.' : null,
                    ),
                  TextFormField(
                    initialValue: draft.name,
                    decoration: const InputDecoration(labelText: 'Name'),
                    onChanged: (v) => draft.name = v,
                  ),
                  TextFormField(
                    initialValue: draft.chatId,
                    readOnly: isEdit,
                    decoration: InputDecoration(
                      labelText: 'Chat ID',
                      helperText: isEdit
                          ? 'Immutable decimal string.'
                          : 'Decimal string, e.g. -1001234567890',
                    ),
                    keyboardType: TextInputType.number,
                    inputFormatters: [
                      FilteringTextInputFormatter.allow(RegExp(r'-?\d')),
                    ],
                    onChanged: isEdit ? null : (v) => draft.chatId = v,
                    validator: isEdit
                        ? null
                        : (v) => v != null && isTelegramDecimalId(v)
                              ? null
                              : 'Invalid chat ID.',
                  ),
                  TextFormField(
                    initialValue: draft.messageThreadId,
                    readOnly: isEdit,
                    decoration: const InputDecoration(
                      labelText: 'Topic ID (optional)',
                      helperText: 'Empty means not a Forum topic address.',
                    ),
                    keyboardType: TextInputType.number,
                    inputFormatters: [FilteringTextInputFormatter.digitsOnly],
                    onChanged: isEdit ? null : (v) => draft.messageThreadId = v,
                  ),
                  SwitchListTile(
                    title: const Text('Inbound enabled'),
                    value: draft.inboundEnabled,
                    onChanged: (v) => setState(() => draft.inboundEnabled = v),
                  ),
                  SwitchListTile(
                    title: const Text('Outbound enabled'),
                    value: draft.outboundEnabled,
                    onChanged: (v) => setState(() => draft.outboundEnabled = v),
                  ),
                  DropdownButtonFormField<String>(
                    initialValue: draft.agentId.isEmpty ? null : draft.agentId,
                    decoration: const InputDecoration(
                      labelText: 'Default Agent',
                    ),
                    items: [
                      for (final a in widget.client.agents)
                        DropdownMenuItem(
                          value: a.agentId,
                          child: Text(
                            a.displayName.isEmpty ? a.name : a.displayName,
                          ),
                        ),
                    ],
                    onChanged: (v) => setState(() {
                      draft.agentId = v ?? '';
                      if (draft.agentOwnsModel(_selectedAgent(draft))) {
                        draft.model = '';
                        draft.selectableModels = '';
                      }
                    }),
                  ),
                  if (!draft.agentOwnsModel(_selectedAgent(draft))) ...[
                    DropdownButtonFormField<String>(
                      initialValue: draft.model.isEmpty ? null : draft.model,
                      decoration: const InputDecoration(
                        labelText: 'Model alias (optional)',
                      ),
                      items: [
                        for (final alias in c.modelAliases())
                          DropdownMenuItem(value: alias, child: Text(alias)),
                      ],
                      onChanged: (v) => setState(() => draft.model = v ?? ''),
                    ),
                    TextFormField(
                      initialValue: draft.selectableAgentIds,
                      decoration: const InputDecoration(
                        labelText: 'Selectable Agent IDs',
                      ),
                      onChanged: (v) => draft.selectableAgentIds = v,
                    ),
                    TextFormField(
                      initialValue: draft.selectableModels,
                      decoration: const InputDecoration(
                        labelText: 'Selectable model aliases',
                      ),
                      onChanged: (v) => draft.selectableModels = v,
                    ),
                  ],
                  DropdownButtonFormField<TelegramTriggerMode>(
                    initialValue: draft.triggerMode,
                    decoration: const InputDecoration(labelText: 'Trigger'),
                    items: TelegramTriggerMode.values
                        .where(
                          (m) =>
                              m !=
                              TelegramTriggerMode
                                  .TELEGRAM_TRIGGER_MODE_UNSPECIFIED,
                        )
                        .map(
                          (m) => DropdownMenuItem(
                            value: m,
                            child: Text(triggerModeLabel(m)),
                          ),
                        )
                        .toList(),
                    onChanged: (v) => setState(
                      () => draft.triggerMode =
                          v ?? TelegramTriggerMode.TELEGRAM_TRIGGER_MODE_ALL,
                    ),
                  ),
                  DropdownButtonFormField<TelegramSessionPolicy>(
                    initialValue: draft.sessionPolicy,
                    decoration: const InputDecoration(
                      labelText: 'Session policy',
                    ),
                    items: TelegramSessionPolicy.values
                        .where(
                          (m) =>
                              m !=
                              TelegramSessionPolicy
                                  .TELEGRAM_SESSION_POLICY_UNSPECIFIED,
                        )
                        .map(
                          (m) => DropdownMenuItem(
                            value: m,
                            child: Text(sessionPolicyLabel(m)),
                          ),
                        )
                        .toList(),
                    onChanged: (v) => setState(
                      () => draft.sessionPolicy =
                          v ??
                          TelegramSessionPolicy
                              .TELEGRAM_SESSION_POLICY_DESTINATION,
                    ),
                  ),
                  TextFormField(
                    initialValue: draft.allowedUserIds,
                    decoration: const InputDecoration(
                      labelText: 'Allowed user IDs (optional)',
                    ),
                    onChanged: (v) => draft.allowedUserIds = v,
                  ),
                  TextFormField(
                    initialValue: draft.controllerUserIds,
                    decoration: const InputDecoration(
                      labelText: 'Controller user IDs (optional)',
                    ),
                    onChanged: (v) => draft.controllerUserIds = v,
                  ),
                  DropdownButtonFormField<TelegramReplyMode>(
                    initialValue: draft.replyMode,
                    decoration: const InputDecoration(labelText: 'Reply mode'),
                    items: TelegramReplyMode.values
                        .where(
                          (m) =>
                              m !=
                              TelegramReplyMode.TELEGRAM_REPLY_MODE_UNSPECIFIED,
                        )
                        .map(
                          (m) => DropdownMenuItem(
                            value: m,
                            child: Text(replyModeLabel(m)),
                          ),
                        )
                        .toList(),
                    onChanged: (v) => setState(
                      () => draft.replyMode =
                          v ?? TelegramReplyMode.TELEGRAM_REPLY_MODE_REPLY,
                    ),
                  ),
                  SwitchListTile(
                    title: const Text('Debug output default on'),
                    value: draft.debugDefault,
                    onChanged: (v) => setState(() => draft.debugDefault = v),
                  ),
                  if (canManage)
                    FilledButton(
                      onPressed: c.writeBlocked ? null : _save,
                      child: Text(
                        isEdit ? 'Save destination' : 'Create destination',
                      ),
                    ),
                ],
              ),
            ),
    );
  }
}

class TelegramDestinationDetailPage extends StatefulWidget {
  const TelegramDestinationDetailPage({
    super.key,
    required this.client,
    required this.destinationId,
  });
  final ClientContext client;
  final String destinationId;

  @override
  State<TelegramDestinationDetailPage> createState() =>
      _TelegramDestinationDetailPageState();
}

class _TelegramDestinationDetailPageState
    extends State<TelegramDestinationDetailPage>
    with _TelegramPage {
  final _testText = TextEditingController();

  @override
  void initState() {
    super.initState();
    unawaited(tg?.loadDestination(widget.destinationId));
  }

  @override
  void dispose() {
    _testText.dispose();
    super.dispose();
  }

  Future<void> _testSend() async {
    final c = tg;
    if (c == null) return;
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (ctx) => AlertDialog(
        title: const Text('Send test message?'),
        content: const Text(
          'This sends a real Telegram message to this destination.',
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(ctx, false),
            child: const Text('Cancel'),
          ),
          FilledButton(
            onPressed: () => Navigator.pop(ctx, true),
            child: const Text('Send'),
          ),
        ],
      ),
    );
    if (confirmed != true || !mounted) return;
    final result = await c.sendTestMessage(
      widget.destinationId,
      text: _testText.text,
    );
    _testText.clear();
    if (!mounted) return;
    if (result.ok) {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text(
            result.messageIds.isEmpty
                ? 'Test message sent.'
                : 'Sent message IDs: ${result.messageIds.join(', ')}',
          ),
        ),
      );
    }
  }

  @override
  Widget build(BuildContext context) {
    final c = tg;
    final d = c?.destination;
    return Scaffold(
      appBar: AppBar(
        title: Text(d?.name.isEmpty ?? true ? 'Destination' : d!.name),
        actions: [
          if (canManage)
            IconButton(
              tooltip: 'Edit',
              icon: const Icon(Icons.edit),
              onPressed: () => Navigator.push<void>(
                context,
                MaterialPageRoute(
                  builder: (_) => TelegramDestinationFormPage(
                    client: widget.client,
                    channelId: d!.channelId,
                    destinationId: widget.destinationId,
                  ),
                ),
              ),
            ),
        ],
      ),
      body: c == null || d == null
          ? const Center(child: CircularProgressIndicator())
          : ListView(
              padding: const EdgeInsets.all(16),
              children: [
                writeNotice(c),
                Text(formatAddress(d)),
                Text(
                  d.verification.verified
                      ? 'Verified'
                      : 'Not verified yet — use test send.',
                ),
                if (canManage) ...[
                  TextField(
                    controller: _testText,
                    decoration: const InputDecoration(
                      labelText: 'Test message (optional)',
                    ),
                  ),
                  FilledButton(
                    onPressed: c.writeBlocked ? null : _testSend,
                    child: const Text('Send test message'),
                  ),
                ],
              ],
            ),
    );
  }
}

class TelegramProcessingListPage extends StatefulWidget {
  const TelegramProcessingListPage({super.key, required this.client});
  final ClientContext client;

  @override
  State<TelegramProcessingListPage> createState() =>
      _TelegramProcessingListPageState();
}

class _TelegramProcessingListPageState extends State<TelegramProcessingListPage>
    with _TelegramPage {
  @override
  void initState() {
    super.initState();
    unawaited(tg?.loadProcessingRecords());
  }

  @override
  Widget build(BuildContext context) {
    final c = tg;
    return Scaffold(
      appBar: AppBar(title: const Text('Telegram processing')),
      body: c == null
          ? const Center(child: Text('Unavailable.'))
          : ListView(
              padding: const EdgeInsets.all(16),
              children: [
                writeNotice(c),
                if (c.processingState == TelegramLoad.loading)
                  const LinearProgressIndicator(),
                if (c.processingError != null) Text(c.processingError!),
                for (final r in c.records)
                  ListTile(
                    title: Text(r.id),
                    subtitle: Text(processingStatusLabel(r.status)),
                    onTap: () => Navigator.push<void>(
                      context,
                      MaterialPageRoute(
                        builder: (_) => TelegramProcessingDetailPage(
                          client: widget.client,
                          recordId: r.id,
                        ),
                      ),
                    ),
                  ),
                if (c.processingState == TelegramLoad.ready &&
                    c.records.isEmpty)
                  const Text('No processing records.'),
              ],
            ),
    );
  }
}

class TelegramProcessingDetailPage extends StatefulWidget {
  const TelegramProcessingDetailPage({
    super.key,
    required this.client,
    required this.recordId,
  });
  final ClientContext client;
  final String recordId;

  @override
  State<TelegramProcessingDetailPage> createState() =>
      _TelegramProcessingDetailPageState();
}

class _TelegramProcessingDetailPageState
    extends State<TelegramProcessingDetailPage>
    with _TelegramPage {
  @override
  void initState() {
    super.initState();
    unawaited(tg?.loadProcessingRecord(widget.recordId));
  }

  Future<void> _resend() async {
    final c = tg;
    if (c == null) return;
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (ctx) => AlertDialog(
        title: const Text('Resend pending delivery?'),
        content: const Text(
          'Re-sends only segments that did not land. The Agent is never run again.',
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
    if (confirmed != true || !mounted) return;
    if (await c.resendReply(widget.recordId) && mounted) {
      ScaffoldMessenger.of(
        context,
      ).showSnackBar(const SnackBar(content: Text('Resend completed')));
    }
  }

  @override
  Widget build(BuildContext context) {
    final c = tg;
    final r = c?.record;
    final canResend =
        canManage &&
        r != null &&
        r.output.isNotEmpty &&
        (r.status ==
                TelegramProcessingStatus
                    .TELEGRAM_PROCESSING_STATUS_READY_TO_DELIVER ||
            r.status ==
                TelegramProcessingStatus.TELEGRAM_PROCESSING_STATUS_FAILED);
    return Scaffold(
      appBar: AppBar(title: const Text('Processing record')),
      body: c == null || r == null
          ? const Center(child: CircularProgressIndicator())
          : ListView(
              padding: const EdgeInsets.all(16),
              children: [
                writeNotice(c),
                Text('Status: ${processingStatusLabel(r.status)}'),
                Text('Update ${r.updateId} · channel ${r.channelId}'),
                if (r.error.isNotEmpty)
                  Text(
                    r.error,
                    style: TextStyle(
                      color: Theme.of(context).colorScheme.error,
                    ),
                  ),
                if (r.output.isNotEmpty)
                  Padding(
                    padding: const EdgeInsets.only(top: 8),
                    child: Text('Stored output (${r.output.length} chars)'),
                  ),
                if (canResend)
                  FilledButton(
                    onPressed: c.writeBlocked ? null : _resend,
                    child: const Text('Resend pending segments'),
                  ),
                if (!canManage)
                  const Text(
                    'Resending requires workspace owner or admin access.',
                  ),
              ],
            ),
    );
  }
}
