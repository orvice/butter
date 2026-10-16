import 'package:flutter/material.dart';

import 'protocol.dart';
import 'surface_processor.dart';
import 'surface_store.dart';

/// Renders one A2UI surface, or its fallback when it cannot be shown.
class A2uiSurfaceView extends StatelessWidget {
  const A2uiSurfaceView({
    super.key,
    required this.entry,
    required this.draft,
    required this.locked,
    required this.onDraftChanged,
    required this.onSubmit,
  });

  final SurfaceEntry entry;
  final Map<String, Object?> draft;
  final bool locked;
  final void Function(Map<String, Object?> draft) onDraftChanged;
  final VoidCallback onSubmit;

  @override
  Widget build(BuildContext context) {
    if (entry.deleted) return const SizedBox.shrink();
    if (entry.error != null || entry.processed == null) {
      return _Fallback(entry: entry);
    }
    final disabled =
        locked ||
        entry.formStatus == FormStatus.submitting ||
        entry.formStatus == FormStatus.answered;
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        _SurfaceTree(
          entry: entry,
          processed: entry.processed!,
          draft: draft,
          disabled: disabled,
          fieldErrors: entry.fieldErrors,
          onDraftChanged: onDraftChanged,
          onSubmit: onSubmit,
        ),
        if (entry.kind == SurfaceKind.form) _FormStatusLine(entry: entry),
      ],
    );
  }
}

class _Fallback extends StatelessWidget {
  const _Fallback({required this.entry});
  final SurfaceEntry entry;

  @override
  Widget build(BuildContext context) {
    final scheme = Theme.of(context).colorScheme;
    return Container(
      key: Key('a2ui-fallback-${entry.id}'),
      padding: const EdgeInsets.all(12),
      decoration: BoxDecoration(
        border: Border.all(color: scheme.outlineVariant),
        borderRadius: BorderRadius.circular(8),
        color: scheme.surfaceContainerHighest.withValues(alpha: 0.4),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(
            entry.kind == SurfaceKind.form
                ? 'This form cannot be displayed here; answer in text instead.'
                : 'This card cannot be displayed; showing its text version.',
            style: Theme.of(context).textTheme.labelSmall,
          ),
          if (entry.fallback != null && entry.fallback!.isNotEmpty) ...[
            const SizedBox(height: 4),
            Text(entry.fallback!, style: Theme.of(context).textTheme.bodySmall),
          ],
          if (entry.error != null) ...[
            const SizedBox(height: 4),
            Text(
              entry.error!,
              style: TextStyle(color: scheme.error, fontSize: 12),
            ),
          ],
        ],
      ),
    );
  }
}

class _FormStatusLine extends StatelessWidget {
  const _FormStatusLine({required this.entry});
  final SurfaceEntry entry;

  @override
  Widget build(BuildContext context) {
    switch (entry.formStatus) {
      case FormStatus.submitting:
        return const Padding(
          padding: EdgeInsets.only(top: 8),
          child: Text('Submitting…'),
        );
      case FormStatus.answered:
        return Padding(
          padding: const EdgeInsets.only(top: 8),
          child: Text(
            entry.formMessage ?? 'Submitted',
            style: TextStyle(color: Theme.of(context).colorScheme.primary),
          ),
        );
      case FormStatus.error:
        if (entry.formMessage == null) return const SizedBox.shrink();
        return Padding(
          padding: const EdgeInsets.only(top: 8),
          child: Text(
            entry.formMessage!,
            style: TextStyle(color: Theme.of(context).colorScheme.error),
          ),
        );
      default:
        return const SizedBox.shrink();
    }
  }
}

class _SurfaceTree extends StatelessWidget {
  const _SurfaceTree({
    required this.entry,
    required this.processed,
    required this.draft,
    required this.disabled,
    required this.fieldErrors,
    required this.onDraftChanged,
    required this.onSubmit,
  });

  final SurfaceEntry entry;
  final ProcessedSurface processed;
  final Map<String, Object?> draft;
  final bool disabled;
  final Map<String, String> fieldErrors;
  final void Function(Map<String, Object?> draft) onDraftChanged;
  final VoidCallback onSubmit;

  @override
  Widget build(BuildContext context) {
    final root = processed.rootId ?? 'root';
    if (!processed.components.containsKey(root)) {
      return _Fallback(entry: entry);
    }
    return _ComponentNode(
      id: root,
      processed: processed,
      draft: draft,
      disabled: disabled,
      fieldErrors: fieldErrors,
      isForm: entry.kind == SurfaceKind.form,
      onDraftChanged: onDraftChanged,
      onSubmit: onSubmit,
    );
  }
}

class _ComponentNode extends StatelessWidget {
  const _ComponentNode({
    required this.id,
    required this.processed,
    required this.draft,
    required this.disabled,
    required this.fieldErrors,
    required this.isForm,
    required this.onDraftChanged,
    required this.onSubmit,
  });

  final String id;
  final ProcessedSurface processed;
  final Map<String, Object?> draft;
  final bool disabled;
  final Map<String, String> fieldErrors;
  final bool isForm;
  final void Function(Map<String, Object?> draft) onDraftChanged;
  final VoidCallback onSubmit;

  Map<String, dynamic>? get _c => processed.components[id];

  @override
  Widget build(BuildContext context) {
    final c = _c;
    if (c == null) return const SizedBox.shrink();
    final name = c['component'] as String? ?? '';
    switch (name) {
      case 'Column':
        return _flex(context, c, Axis.vertical);
      case 'Row':
        return _flex(context, c, Axis.horizontal);
      case 'Text':
        return _text(context, c);
      case 'Card':
        return _card(context, c);
      case 'Divider':
        return _divider(c);
      case 'KeyValue':
        return _keyValue(context, c);
      case 'Status':
        return _status(context, c);
      case 'TextField':
        return isForm
            ? _FormTextField(
                name: c['name'] as String? ?? '',
                label: resolveString(c['label'], processed.data),
                hint: resolveString(c['hint'], processed.data),
                value:
                    draft[c['name'] as String? ?? '']?.toString() ??
                    resolveString(c['value'], processed.data),
                disabled: disabled,
                errorText: fieldErrors[c['name'] as String? ?? ''],
                multiline: c['variant'] == 'multiline',
                onChanged: (v) {
                  final name = c['name'] as String? ?? '';
                  final next = Map<String, Object?>.from(draft)..[name] = v;
                  onDraftChanged(next);
                },
              )
            : const SizedBox.shrink();
      case 'ChoicePicker':
        return isForm ? _choicePicker(context, c) : const SizedBox.shrink();
      case 'Button':
        return isForm ? _button(context, c) : const SizedBox.shrink();
      default:
        return const SizedBox.shrink();
    }
  }

  Widget _flex(BuildContext context, Map<String, dynamic> c, Axis axis) {
    final children = _childIds(c['children']);
    return Flex(
      direction: axis,
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        for (final childId in children)
          _ComponentNode(
            id: childId,
            processed: processed,
            draft: draft,
            disabled: disabled,
            fieldErrors: fieldErrors,
            isForm: isForm,
            onDraftChanged: onDraftChanged,
            onSubmit: onSubmit,
          ),
      ],
    );
  }

  Widget _text(BuildContext context, Map<String, dynamic> c) {
    final text = resolveString(c['text'], processed.data);
    final variant = c['variant'] as String? ?? 'body';
    final style = switch (variant) {
      'h1' => Theme.of(context).textTheme.titleLarge,
      'h2' => Theme.of(context).textTheme.titleMedium,
      'h3' => Theme.of(context).textTheme.titleSmall,
      'h4' => Theme.of(context).textTheme.labelLarge,
      'h5' => Theme.of(context).textTheme.labelSmall,
      'caption' => Theme.of(context).textTheme.bodySmall,
      _ => Theme.of(context).textTheme.bodyMedium,
    };
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 2),
      child: Text(text, style: style),
    );
  }

  Widget _card(BuildContext context, Map<String, dynamic> c) {
    final childId = c['child'];
    if (childId is! String) return const SizedBox.shrink();
    return Card(
      margin: const EdgeInsets.symmetric(vertical: 4),
      child: Padding(
        padding: const EdgeInsets.all(12),
        child: _ComponentNode(
          id: childId,
          processed: processed,
          draft: draft,
          disabled: disabled,
          fieldErrors: fieldErrors,
          isForm: isForm,
          onDraftChanged: onDraftChanged,
          onSubmit: onSubmit,
        ),
      ),
    );
  }

  Widget _divider(Map<String, dynamic> c) {
    final axis = c['axis'] as String? ?? 'horizontal';
    return axis == 'vertical'
        ? const VerticalDivider(width: 16)
        : const Divider(height: 16);
  }

  Widget _keyValue(BuildContext context, Map<String, dynamic> c) {
    final label = resolveString(c['label'], processed.data);
    final value = resolveString(c['value'], processed.data);
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 2),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Expanded(
            flex: 2,
            child: Text(label, style: Theme.of(context).textTheme.labelMedium),
          ),
          Expanded(flex: 3, child: Text(value)),
        ],
      ),
    );
  }

  Widget _status(BuildContext context, Map<String, dynamic> c) {
    final text = resolveString(c['text'], processed.data);
    final tone = c['tone'] as String? ?? 'neutral';
    final color = switch (tone) {
      'success' => Colors.green.shade700,
      'warning' => Colors.orange.shade800,
      'error' => Theme.of(context).colorScheme.error,
      'info' => Theme.of(context).colorScheme.primary,
      _ => Theme.of(context).colorScheme.onSurfaceVariant,
    };
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
      decoration: BoxDecoration(
        color: color.withValues(alpha: 0.12),
        borderRadius: BorderRadius.circular(4),
      ),
      child: Text(text, style: TextStyle(color: color)),
    );
  }

  Widget _choicePicker(BuildContext context, Map<String, dynamic> c) {
    final name = c['name'] as String? ?? '';
    final label = resolveString(c['label'], processed.data);
    final selected = draft[name]?.toString() ?? '';
    final options = <Map<String, String>>[];
    if (c['options'] is List) {
      for (final opt in c['options'] as List) {
        if (opt is Map) {
          final v = opt['value'];
          final l = opt['label'];
          if (v is String && l is String) {
            options.add({'value': v, 'label': l});
          }
        }
      }
    }
    return Padding(
      padding: const EdgeInsets.only(bottom: 8),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          if (label.isNotEmpty)
            Text(label, style: Theme.of(context).textTheme.labelLarge),
          if (fieldErrors[name] != null)
            Text(
              fieldErrors[name]!,
              style: TextStyle(
                color: Theme.of(context).colorScheme.error,
                fontSize: 12,
              ),
            ),
          DropdownButtonFormField<String>(
            key: Key('choice-$name'),
            initialValue: selected.isEmpty ? null : selected,
            decoration: InputDecoration(errorText: fieldErrors[name]),
            items: [
              for (final opt in options)
                DropdownMenuItem(
                  value: opt['value'],
                  child: Text(opt['label']!),
                ),
            ],
            onChanged: disabled
                ? null
                : (v) {
                    if (v == null) return;
                    final next = Map<String, Object?>.from(draft)..[name] = v;
                    onDraftChanged(next);
                  },
          ),
        ],
      ),
    );
  }

  Widget _button(BuildContext context, Map<String, dynamic> c) {
    final action = c['action'];
    if (action is! Map) return const SizedBox.shrink();
    final event = action['event'];
    if (event is! Map || event['name'] != submitEventName) {
      return const SizedBox.shrink();
    }
    final childId = c['child'];
    final label = childId is String
        ? resolveString(processed.components[childId]?['text'], processed.data)
        : 'Submit';
    return Align(
      alignment: Alignment.centerRight,
      child: FilledButton(
        key: Key('submit-${processed.surfaceId}'),
        onPressed: disabled ? null : onSubmit,
        child: Text(label.isEmpty ? 'Submit' : label),
      ),
    );
  }

  List<String> _childIds(Object? raw) {
    if (raw is! List) return const [];
    return [
      for (final item in raw)
        if (item is String) item,
    ];
  }
}

/// Keeps one [TextEditingController] across rebuilds so focus and drafts survive.
class _FormTextField extends StatefulWidget {
  const _FormTextField({
    required this.name,
    required this.label,
    required this.hint,
    required this.value,
    required this.disabled,
    required this.errorText,
    required this.multiline,
    required this.onChanged,
  });

  final String name;
  final String label;
  final String hint;
  final String value;
  final bool disabled;
  final String? errorText;
  final bool multiline;
  final ValueChanged<String> onChanged;

  @override
  State<_FormTextField> createState() => _FormTextFieldState();
}

class _FormTextFieldState extends State<_FormTextField> {
  late final TextEditingController _controller;
  final _focus = FocusNode();

  @override
  void initState() {
    super.initState();
    _controller = TextEditingController(text: widget.value)
      ..selection = TextSelection.collapsed(offset: widget.value.length);
  }

  @override
  void didUpdateWidget(_FormTextField oldWidget) {
    super.didUpdateWidget(oldWidget);
    if (widget.value != _controller.text &&
        widget.value != oldWidget.value &&
        !_focus.hasFocus) {
      _controller.text = widget.value;
      _controller.selection = TextSelection.collapsed(
        offset: widget.value.length,
      );
    }
  }

  @override
  void dispose() {
    _focus.dispose();
    _controller.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.only(bottom: 8),
      child: TextField(
        key: Key('field-${widget.name}'),
        focusNode: _focus,
        enabled: !widget.disabled,
        decoration: InputDecoration(
          labelText: widget.label.isEmpty ? null : widget.label,
          hintText: widget.hint.isEmpty ? null : widget.hint,
          errorText: widget.errorText,
        ),
        controller: _controller,
        onChanged: widget.onChanged,
        maxLines: widget.multiline ? 4 : 1,
      ),
    );
  }
}
