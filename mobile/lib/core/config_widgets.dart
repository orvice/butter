import 'package:flutter/material.dart';

import 'config_resource_controller.dart';
import 'url_validation.dart';

/// A scaffold that shows a configuration controller's write progress, its
/// unconfirmed-change notice and its safe write error.
class ConfigScaffold extends StatelessWidget {
  const ConfigScaffold({
    super.key,
    required this.controller,
    required this.title,
    required this.child,
  });
  final ConfigResourceController controller;
  final String title;
  final Widget child;

  @override
  Widget build(BuildContext context) => ListenableBuilder(
    listenable: controller,
    builder: (context, _) => Scaffold(
      appBar: AppBar(title: Text(title)),
      body: SafeArea(
        child: Column(
          children: [
            if (controller.writing)
              const LinearProgressIndicator(semanticsLabel: 'Saving'),
            if (controller.writeNotice != null)
              Padding(
                padding: const EdgeInsets.all(12),
                child: Semantics(
                  liveRegion: true,
                  child: Text(
                    controller.writeNotice!,
                    key: const Key('config-write-notice'),
                  ),
                ),
              ),
            if (controller.writeError != null)
              Padding(
                padding: const EdgeInsets.all(12),
                child: Semantics(
                  liveRegion: true,
                  child: Text(
                    controller.writeError!,
                    key: const Key('config-write-error'),
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

/// A line of text that reports a problem, announced to screen readers.
class ErrorLine extends StatelessWidget {
  const ErrorLine(this.text, {super.key});
  final String text;
  @override
  Widget build(BuildContext context) => Semantics(
    liveRegion: true,
    child: Text(
      text,
      style: TextStyle(color: Theme.of(context).colorScheme.error),
    ),
  );
}

/// Keep, replace or clear one stored credential. A stored value is never
/// displayed; a replacement is typed into a masked field.
class SecretField extends StatelessWidget {
  const SecretField({
    super.key,
    required this.label,
    required this.stored,
    required this.edit,
    required this.value,
    required this.onEdit,
    required this.onValue,
    this.enabled = true,
    this.error,
  });
  final String label;
  final bool stored;
  final SecretEdit edit;
  final String value;
  final ValueChanged<SecretEdit> onEdit;
  final ValueChanged<String> onValue;
  final bool enabled;
  final String? error;

  @override
  Widget build(BuildContext context) {
    // Without a stored value the only choices are none or a new one.
    final choices = stored
        ? SecretEdit.values
        : const [SecretEdit.keep, SecretEdit.replace];
    String name(SecretEdit e) => switch (e) {
      SecretEdit.keep => stored ? 'Keep' : 'None',
      SecretEdit.replace => 'Replace',
      SecretEdit.clear => 'Clear',
    };
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(label, style: Theme.of(context).textTheme.titleSmall),
        Text(
          stored
              ? 'A value is stored and is never shown.'
              : 'No value is stored.',
        ),
        const SizedBox(height: 4),
        SegmentedButton<SecretEdit>(
          key: Key('secret-choice-$label'),
          segments: [
            for (final e in choices)
              ButtonSegment(value: e, label: Text(name(e))),
          ],
          selected: {edit},
          showSelectedIcon: false,
          onSelectionChanged: enabled ? (s) => onEdit(s.first) : null,
        ),
        if (edit == SecretEdit.replace)
          TextFormField(
            key: Key('secret-value-$label'),
            initialValue: value,
            enabled: enabled,
            obscureText: true,
            autocorrect: false,
            enableSuggestions: false,
            autofillHints: const [],
            decoration: InputDecoration(labelText: 'New $label'),
            onChanged: onValue,
          ),
        if (edit == SecretEdit.clear)
          const Text('The stored value will be removed when you save.'),
        if (error != null) ErrorLine(error!),
      ],
    );
  }
}
