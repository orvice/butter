import 'package:flutter/material.dart';

import 'linear_controller.dart';

/// Write-only client and webhook secrets on an existing App (#466).
class LinearSecretsCard extends StatefulWidget {
  const LinearSecretsCard({
    super.key,
    required this.linear,
    required this.appId,
    required this.clientSecretSet,
    required this.webhookSecretSet,
    required this.inboundEnabled,
    required this.canManage,
  });

  final LinearController linear;
  final String appId;
  final bool clientSecretSet;
  final bool webhookSecretSet;
  final bool inboundEnabled;
  final bool canManage;

  @override
  State<LinearSecretsCard> createState() => _LinearSecretsCardState();
}

class _LinearSecretsCardState extends State<LinearSecretsCard> {
  final _clientCtrl = TextEditingController();
  final _webhookCtrl = TextEditingController();

  @override
  void dispose() {
    _clientCtrl.dispose();
    _webhookCtrl.dispose();
    super.dispose();
  }

  Future<void> _replace({required bool client}) async {
    final value = (client ? _clientCtrl : _webhookCtrl).text.trim();
    if (value.isEmpty) return;
    final ok = await widget.linear.putCredentials(
      widget.appId,
      clientSecret: client ? value : null,
      webhookSecret: client ? null : value,
    );
    if (ok && mounted) {
      if (client) {
        _clientCtrl.clear();
      } else {
        _webhookCtrl.clear();
      }
      setState(() {});
    }
  }

  Future<void> _clear({required bool client}) async {
    final label = client ? 'client secret' : 'webhook signing secret';
    final ok = await showDialog<bool>(
      context: context,
      builder: (ctx) => AlertDialog(
        title: Text('Clear $label?'),
        content: Text(
          widget.inboundEnabled
              ? 'Turn inbound off before clearing secrets required for delivery.'
              : 'The stored $label will be removed. This cannot be undone from the App.',
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(ctx, false),
            child: const Text('Cancel'),
          ),
          FilledButton(
            onPressed: () => Navigator.pop(ctx, true),
            child: const Text('Clear'),
          ),
        ],
      ),
    );
    if (ok != true) return;
    await widget.linear.putCredentials(
      widget.appId,
      clientSecret: client ? '' : null,
      webhookSecret: client ? null : '',
    );
  }

  @override
  Widget build(BuildContext context) {
    if (!widget.canManage) return const SizedBox.shrink();
    return Card(
      child: Padding(
        padding: const EdgeInsets.all(12),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(
              'Secrets (write-only)',
              style: Theme.of(context).textTheme.titleMedium,
            ),
            const SizedBox(height: 8),
            Text(
              widget.inboundEnabled
                  ? 'Disable inbound before clearing a secret.'
                  : 'Encrypted at rest; never shown again after saving.',
            ),
            const SizedBox(height: 12),
            _SecretRow(
              label: 'Client secret',
              isSet: widget.clientSecretSet,
              controller: _clientCtrl,
              enabled: !widget.linear.writing,
              onReplace: () => _replace(client: true),
              onClear: widget.inboundEnabled
                  ? null
                  : () => _clear(client: true),
            ),
            const SizedBox(height: 12),
            _SecretRow(
              label: 'Webhook signing secret',
              isSet: widget.webhookSecretSet,
              controller: _webhookCtrl,
              enabled: !widget.linear.writing,
              onReplace: () => _replace(client: false),
              onClear: widget.inboundEnabled
                  ? null
                  : () => _clear(client: false),
            ),
          ],
        ),
      ),
    );
  }
}

class _SecretRow extends StatelessWidget {
  const _SecretRow({
    required this.label,
    required this.isSet,
    required this.controller,
    required this.enabled,
    required this.onReplace,
    this.onClear,
  });

  final String label;
  final bool isSet;
  final TextEditingController controller;
  final bool enabled;
  final VoidCallback onReplace;
  final VoidCallback? onClear;

  @override
  Widget build(BuildContext context) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text('$label (${isSet ? 'set' : 'not set'})'),
        Row(
          children: [
            Expanded(
              child: TextField(
                controller: controller,
                obscureText: true,
                enabled: enabled,
                decoration: InputDecoration(
                  hintText: isSet ? 'Replace…' : 'Set…',
                ),
                onSubmitted: (_) => onReplace(),
              ),
            ),
            TextButton(
              onPressed: enabled ? onReplace : null,
              child: Text(isSet ? 'Replace' : 'Set'),
            ),
            if (onClear != null)
              TextButton(
                onPressed: enabled && isSet ? onClear : null,
                child: const Text('Clear'),
              ),
          ],
        ),
      ],
    );
  }
}
