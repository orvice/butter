import 'dart:async';

import 'package:flutter/material.dart';

import '../../app/client_context.dart';
import '../../core/management_page.dart';
import '../../gen/agents/v1/butterbox.pb.dart';
import 'box_controller.dart';
import 'box_draft.dart';

/// Opens the Workspace's ButterBoxes. The list reads fresh from the server and
/// offers a reload when that read failed.
Future<void> openButterBoxes(BuildContext context, ClientContext client) async {
  final boxes = client.boxes;
  if (boxes == null) return;
  unawaited(boxes.loadBoxes());
  await openManagementPage(context, client, ButterBoxListPage(client: client));
}

Future<void> _openDetail(
  BuildContext context,
  ClientContext client,
  String id,
) => openManagementPage(
  context,
  client,
  ButterBoxDetailPage(client: client, id: id),
);

/// A scaffold that shows the controller's progress and safe errors the way
/// [ManagementPage] does for the client.
class _BoxScaffold extends StatelessWidget {
  const _BoxScaffold({
    required this.boxes,
    required this.title,
    required this.child,
  });
  final ButterBoxController boxes;
  final String title;
  final Widget child;

  @override
  Widget build(BuildContext context) => ListenableBuilder(
    listenable: boxes,
    builder: (context, _) => Scaffold(
      appBar: AppBar(title: Text(title)),
      body: SafeArea(
        child: Column(
          children: [
            if (boxes.writing)
              const LinearProgressIndicator(semanticsLabel: 'Saving'),
            if (boxes.writeNotice != null)
              Padding(
                padding: const EdgeInsets.all(12),
                child: Semantics(
                  liveRegion: true,
                  child: Text(boxes.writeNotice!),
                ),
              ),
            if (boxes.writeError != null)
              Padding(
                padding: const EdgeInsets.all(12),
                child: Semantics(
                  liveRegion: true,
                  child: Text(
                    boxes.writeError!,
                    key: const Key('box-write-error'),
                    style: TextStyle(
                      color: Theme.of(context).colorScheme.error,
                    ),
                  ),
                ),
              ),
            Expanded(child: child),
          ],
        ),
      ),
    ),
  );
}

String probeSummary(BoxProbe? probe) {
  if (probe == null) return 'Connection not tested';
  return switch (probe.state) {
    BoxProbeState.idle => 'Connection not tested',
    BoxProbeState.probing => 'Testing connection…',
    BoxProbeState.reachable =>
      'Reachable · ${probe.activeSessions} active session${probe.activeSessions == 1 ? '' : 's'}',
    BoxProbeState.unreachable =>
      probe.detail.isEmpty ? 'Unreachable' : 'Unreachable: ${probe.detail}',
    BoxProbeState.failed => 'Test failed: ${probe.detail}',
  };
}

class ButterBoxListPage extends StatelessWidget {
  const ButterBoxListPage({super.key, required this.client});
  final ClientContext client;

  @override
  Widget build(BuildContext context) {
    final boxes = client.boxes;
    if (boxes == null) return const SizedBox.shrink();
    return _BoxScaffold(
      boxes: boxes,
      title: 'ButterBoxes',
      child: ListenableBuilder(
        listenable: boxes,
        builder: (context, _) => ListView(
          padding: const EdgeInsets.all(16),
          children: [
            Text(
              '${client.workspace?.name ?? 'Workspace'} · ButterBoxes',
              style: Theme.of(context).textTheme.headlineSmall,
            ),
            const SizedBox(height: 8),
            const Text(
              'A ButterBox is an agent VM that runs Pi and Cursor Agents. Butter holds its access token and talks to it; this app never connects to a Box.',
            ),
            const SizedBox(height: 12),
            Wrap(
              spacing: 12,
              children: [
                FilledButton.icon(
                  key: const Key('register-box'),
                  onPressed: boxes.writing
                      ? null
                      : () => openManagementPage(
                          context,
                          client,
                          ButterBoxFormPage(client: client),
                        ),
                  icon: const Icon(Icons.add),
                  label: const Text('Register ButterBox'),
                ),
                TextButton.icon(
                  key: const Key('refresh-boxes'),
                  onPressed: boxes.listState == BoxLoad.loading
                      ? null
                      : boxes.loadBoxes,
                  icon: const Icon(Icons.refresh),
                  label: const Text('Refresh ButterBoxes'),
                ),
              ],
            ),
            if (boxes.listState == BoxLoad.loading)
              const LinearProgressIndicator(semanticsLabel: 'Loading'),
            if (boxes.listState == BoxLoad.failed)
              Semantics(
                liveRegion: true,
                child: Text(
                  boxes.listError ?? 'ButterBoxes could not be loaded.',
                  key: const Key('box-list-error'),
                  style: TextStyle(color: Theme.of(context).colorScheme.error),
                ),
              ),
            if (boxes.listState == BoxLoad.ready && boxes.boxes.isEmpty) ...[
              const SizedBox(height: 24),
              Text(
                'No ButterBoxes yet',
                style: Theme.of(context).textTheme.titleMedium,
              ),
              const Text('Register a box to run Pi or Cursor Agents on it.'),
            ],
            for (final box in boxes.boxes)
              ListTile(
                key: Key('box-${box.id}'),
                isThreeLine: true,
                contentPadding: EdgeInsets.zero,
                title: Text(box.name),
                subtitle: Text(
                  [
                    box.baseUrl,
                    [
                      box.enabled ? 'Enabled' : 'Disabled',
                      box.credentialSet
                          ? 'Access token set'
                          : 'No access token',
                    ].join(' · '),
                    probeSummary(boxes.probes[box.id]),
                  ].join('\n'),
                ),
                trailing: IconButton(
                  tooltip: 'Test connection to ${box.name}',
                  icon: Icon(
                    Icons.network_check,
                    semanticLabel: 'Test connection to ${box.name}',
                  ),
                  onPressed:
                      boxes.probes[box.id]?.state == BoxProbeState.probing
                      ? null
                      : () => boxes.probe(box.id),
                ),
                onTap: () => _openDetail(context, client, box.id),
              ),
          ],
        ),
      ),
    );
  }
}

class ButterBoxDetailPage extends StatefulWidget {
  const ButterBoxDetailPage({
    super.key,
    required this.client,
    required this.id,
  });
  final ClientContext client;
  final String id;
  @override
  State<ButterBoxDetailPage> createState() => _ButterBoxDetailPageState();
}

class _ButterBoxDetailPageState extends State<ButterBoxDetailPage> {
  @override
  void initState() {
    super.initState();
    widget.client.boxes?.loadBox(widget.id);
  }

  Future<void> _replaceToken(ButterBoxController boxes, ButterBox box) async {
    final token = await showDialog<String>(
      context: context,
      builder: (_) => _TokenDialog(name: box.name),
    );
    if (token == null || !mounted) return;
    await boxes.replaceToken(box.id, token);
  }

  Future<void> _clearToken(ButterBoxController boxes, ButterBox box) async {
    final ok = await confirmChange(
      context,
      'Clear access token?',
      'Butter will stop sending a token to "${box.name}". Agents on it fail if the box requires one. The old token is not recoverable from here.',
      action: 'Clear token',
    );
    if (ok && mounted) await boxes.clearToken(box.id);
  }

  Future<void> _delete(ButterBoxController boxes, ButterBox box) async {
    final ok = await confirmChange(
      context,
      'Delete ButterBox?',
      'Delete "${box.name}"? Butter forgets this box and its access token; nothing is deleted on the box itself. A box that Agents still use cannot be deleted.',
      action: 'Delete',
    );
    if (!ok || !mounted) return;
    if (await boxes.delete(box.id) && mounted && Navigator.canPop(context)) {
      Navigator.pop(context);
    }
  }

  @override
  Widget build(BuildContext context) {
    final boxes = widget.client.boxes;
    if (boxes == null) return const SizedBox.shrink();
    return _BoxScaffold(
      boxes: boxes,
      title: 'ButterBox',
      child: ListenableBuilder(
        listenable: boxes,
        builder: (context, _) {
          final box = boxes.box?.id == widget.id ? boxes.box : null;
          if (box == null) {
            return ListView(
              padding: const EdgeInsets.all(16),
              children: [
                if (boxes.boxState == BoxLoad.loading)
                  const LinearProgressIndicator(semanticsLabel: 'Loading'),
                if (boxes.boxError != null)
                  Text(
                    boxes.boxError!,
                    key: const Key('box-detail-error'),
                    style: TextStyle(
                      color: Theme.of(context).colorScheme.error,
                    ),
                  ),
                TextButton(
                  onPressed: boxes.boxState == BoxLoad.loading
                      ? null
                      : () => boxes.loadBox(widget.id),
                  child: const Text('Load ButterBox'),
                ),
              ],
            );
          }
          final probe = boxes.probes[box.id];
          final busy = boxes.writing;
          return ListView(
            padding: const EdgeInsets.all(16),
            children: [
              SelectableText(
                box.name,
                style: Theme.of(context).textTheme.headlineSmall,
              ),
              SelectableText(box.baseUrl),
              SelectableText('ID: ${box.id}'),
              Text(
                box.enabled
                    ? 'Enabled'
                    : 'Disabled — not offered for new Agent bindings',
              ),
              Text(
                box.credentialSet
                    ? 'Access token: set${box.hasCredentialUpdatedAt() ? ' · updated ${box.credentialUpdatedAt.toDateTime().toLocal()}' : ''}'
                    : 'Access token: none',
                key: const Key('box-token-state'),
              ),
              const Text(
                'The token is write-only: it is encrypted on the server and never shown again.',
              ),
              const SizedBox(height: 12),
              Text(
                probeSummary(probe),
                key: const Key('box-probe'),
                style:
                    probe?.state == BoxProbeState.unreachable ||
                        probe?.state == BoxProbeState.failed
                    ? TextStyle(color: Theme.of(context).colorScheme.error)
                    : null,
              ),
              const SizedBox(height: 12),
              Wrap(
                spacing: 12,
                children: [
                  OutlinedButton.icon(
                    key: const Key('probe-box'),
                    onPressed: probe?.state == BoxProbeState.probing
                        ? null
                        : () => boxes.probe(box.id),
                    icon: const Icon(Icons.network_check),
                    label: const Text('Test connection'),
                  ),
                  FilledButton.icon(
                    key: const Key('edit-box'),
                    onPressed: busy
                        ? null
                        : () => openManagementPage(
                            context,
                            widget.client,
                            ButterBoxFormPage(client: widget.client, base: box),
                          ),
                    icon: const Icon(Icons.edit),
                    label: const Text('Edit'),
                  ),
                  OutlinedButton.icon(
                    key: const Key('replace-token'),
                    onPressed: busy || boxes.unconfirmed
                        ? null
                        : () => _replaceToken(boxes, box),
                    icon: const Icon(Icons.key),
                    label: Text(
                      box.credentialSet
                          ? 'Replace access token'
                          : 'Set access token',
                    ),
                  ),
                  if (box.credentialSet)
                    OutlinedButton.icon(
                      key: const Key('clear-token'),
                      onPressed: busy || boxes.unconfirmed
                          ? null
                          : () => _clearToken(boxes, box),
                      icon: const Icon(Icons.key_off),
                      label: const Text('Clear access token'),
                    ),
                  TextButton.icon(
                    key: const Key('refresh-box'),
                    onPressed: boxes.boxState == BoxLoad.loading
                        ? null
                        : () => boxes.loadBox(box.id),
                    icon: const Icon(Icons.refresh),
                    label: const Text('Refresh'),
                  ),
                  TextButton.icon(
                    key: const Key('delete-box'),
                    onPressed: busy || boxes.unconfirmed
                        ? null
                        : () => _delete(boxes, box),
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

/// Asks for the replacement token. The text lives only in this dialog; it is
/// handed back once and never stored or logged.
class _TokenDialog extends StatefulWidget {
  const _TokenDialog({required this.name});
  final String name;
  @override
  State<_TokenDialog> createState() => _TokenDialogState();
}

class _TokenDialogState extends State<_TokenDialog> {
  final _controller = TextEditingController();
  bool _empty = true;

  @override
  void dispose() {
    _controller.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) => AlertDialog(
    title: const Text('Access token'),
    content: Column(
      mainAxisSize: MainAxisSize.min,
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(
          'Enter the access token for "${widget.name}". It replaces the stored token and is never shown again. To remove the token, use Clear access token.',
        ),
        TextField(
          key: const Key('token-field'),
          controller: _controller,
          obscureText: true,
          autocorrect: false,
          enableSuggestions: false,
          autofillHints: const [],
          decoration: const InputDecoration(labelText: 'New access token'),
          onChanged: (v) => setState(() => _empty = v.trim().isEmpty),
        ),
      ],
    ),
    actions: [
      TextButton(
        onPressed: () => Navigator.pop(context),
        child: const Text('Cancel'),
      ),
      FilledButton(
        key: const Key('save-token'),
        onPressed: _empty
            ? null
            : () => Navigator.pop(context, _controller.text.trim()),
        child: const Text('Save token'),
      ),
    ],
  );
}

class ButterBoxFormPage extends StatefulWidget {
  const ButterBoxFormPage({super.key, required this.client, this.base});
  final ClientContext client;
  final ButterBox? base;
  @override
  State<ButterBoxFormPage> createState() => _ButterBoxFormPageState();
}

class _ButterBoxFormPageState extends State<ButterBoxFormPage> {
  final _form = GlobalKey<FormState>();
  late final _name = TextEditingController(text: widget.base?.name ?? '');
  late final _url = TextEditingController(text: widget.base?.baseUrl ?? '');
  final _token = TextEditingController();
  late bool _enabled = widget.base?.enabled ?? true;
  bool _dirty = false;
  bool _saved = false;

  @override
  void dispose() {
    _name.dispose();
    _url.dispose();
    _token.dispose();
    super.dispose();
  }

  Future<void> _save(ButterBoxController boxes) async {
    if (!_form.currentState!.validate()) return;
    FocusScope.of(context).unfocus();
    final base = widget.base;
    final ok = base == null
        ? await boxes.create(
            name: _name.text,
            baseUrl: _url.text,
            enabled: _enabled,
            accessToken: _token.text.trim(),
          )
        : await boxes.update(
            id: base.id,
            name: _name.text,
            baseUrl: _url.text,
            enabled: _enabled,
          );
    if (!mounted || !ok) return;
    // The token field is sensitive: drop it as soon as it was sent.
    _token.clear();
    setState(() {
      _saved = true;
      _dirty = false;
    });
    if (Navigator.canPop(context)) Navigator.pop(context);
  }

  @override
  Widget build(BuildContext context) {
    final boxes = widget.client.boxes;
    if (boxes == null) return const SizedBox.shrink();
    return ListenableBuilder(
      listenable: boxes,
      builder: (context, _) {
        final locked = boxes.writing || _saved;
        return DraftGuard(
          dirty: _dirty,
          busy: boxes.writing,
          child: _BoxScaffold(
            boxes: boxes,
            title: widget.base == null
                ? 'Register ButterBox'
                : 'Edit ButterBox',
            child: Form(
              key: _form,
              child: SingleChildScrollView(
                padding: const EdgeInsets.all(16),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.stretch,
                  children: [
                    TextFormField(
                      key: const Key('box-name'),
                      controller: _name,
                      enabled: !locked,
                      decoration: const InputDecoration(
                        labelText: 'Name',
                        helperText: 'Unique within the Workspace.',
                      ),
                      validator: validateBoxName,
                      onChanged: (_) => setState(() => _dirty = true),
                    ),
                    TextFormField(
                      key: const Key('box-url'),
                      controller: _url,
                      enabled: !locked,
                      keyboardType: TextInputType.url,
                      autocorrect: false,
                      decoration: const InputDecoration(
                        labelText: 'Base URL',
                        helperText:
                            'The box\'s HTTP address, e.g. https://box.example.com. The Butter server connects to it.',
                        helperMaxLines: 3,
                      ),
                      validator: validateBoxUrl,
                      onChanged: (_) => setState(() => _dirty = true),
                    ),
                    SwitchListTile(
                      key: const Key('box-enabled'),
                      contentPadding: EdgeInsets.zero,
                      title: const Text('Enabled'),
                      subtitle: const Text(
                        'A disabled box is not offered for new Pi or Cursor Agent bindings. Agents already bound to it keep working.',
                      ),
                      value: _enabled,
                      onChanged: locked
                          ? null
                          : (v) => setState(() {
                              _enabled = v;
                              _dirty = true;
                            }),
                    ),
                    if (widget.base == null)
                      TextFormField(
                        key: const Key('box-token'),
                        controller: _token,
                        enabled: !locked,
                        obscureText: true,
                        autocorrect: false,
                        enableSuggestions: false,
                        autofillHints: const [],
                        decoration: const InputDecoration(
                          labelText: 'Access token (optional)',
                          helperText:
                              'Write-only: encrypted on the server and never shown again.',
                        ),
                        onChanged: (_) => setState(() => _dirty = true),
                      )
                    else
                      const Padding(
                        padding: EdgeInsets.symmetric(vertical: 8),
                        child: Text(
                          'The access token is managed separately: use Replace or Clear access token on the box\'s page. Saving here never changes it.',
                        ),
                      ),
                    const SizedBox(height: 16),
                    FilledButton(
                      key: const Key('save-box'),
                      onPressed: locked || boxes.unconfirmed
                          ? null
                          : () => _save(boxes),
                      child: Text(_saved ? 'Saved' : 'Save ButterBox'),
                    ),
                    if (boxes.unconfirmed)
                      const Text(
                        'Leave this form and refresh the ButterBox to reconcile the outcome before saving again.',
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
