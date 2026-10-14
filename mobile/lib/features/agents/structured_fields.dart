import 'package:flutter/material.dart';

import '../../gen/agents/v1/agent.pb.dart';

/// Ordered references to independent resources, never embedded Agents or
/// display-name lookups. Missing references remain visible until removed.
class ChildAgentFields extends StatelessWidget {
  const ChildAgentFields({
    super.key,
    required this.agent,
    required this.choices,
    required this.enabled,
    required this.onChanged,
  });
  final Agent agent;
  final List<Agent> choices;
  final bool enabled;
  final VoidCallback onChanged;
  @override
  Widget build(BuildContext context) {
    final children = agent.childAgentIds;
    final available = choices
        .where(
          (a) =>
              a.agentId.isNotEmpty &&
              a.agentId != agent.agentId &&
              !children.contains(a.agentId),
        )
        .toList();
    void move(int index, int target) {
      final child = children.removeAt(index);
      children.insert(target, child);
      onChanged();
    }

    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        const SizedBox(height: 16),
        Text(
          'Ordered child Agents',
          style: Theme.of(context).textTheme.titleMedium,
        ),
        const Text(
          'One parent per child; cycles and incompatible relationships are rejected by the server.',
        ),
        for (var i = 0; i < children.length; i++)
          Card(
            child: Padding(
              padding: const EdgeInsets.all(8),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text('${i + 1}. ${children[i]}'),
                  Wrap(
                    children: [
                      IconButton(
                        tooltip: 'Move ${children[i]} up',
                        icon: const Icon(Icons.arrow_upward),
                        onPressed: enabled && i > 0
                            ? () => move(i, i - 1)
                            : null,
                      ),
                      IconButton(
                        tooltip: 'Move ${children[i]} down',
                        icon: const Icon(Icons.arrow_downward),
                        onPressed: enabled && i < children.length - 1
                            ? () => move(i, i + 1)
                            : null,
                      ),
                      IconButton(
                        tooltip: 'Remove child ${children[i]}',
                        icon: const Icon(Icons.remove_circle_outline),
                        onPressed: enabled
                            ? () {
                                children.removeAt(i);
                                onChanged();
                              }
                            : null,
                      ),
                    ],
                  ),
                ],
              ),
            ),
          ),
        DropdownButtonFormField<String>(
          key: ValueKey('child-add-${children.join('|')}'),
          isExpanded: true,
          decoration: const InputDecoration(
            labelText: 'Add existing child Agent by ID',
          ),
          items: [
            for (final a in available)
              DropdownMenuItem(
                value: a.agentId,
                child: Text(
                  '${a.agentId} · ${a.displayName.isEmpty ? a.name : a.displayName}',
                  overflow: TextOverflow.ellipsis,
                ),
              ),
          ],
          onChanged: !enabled || available.isEmpty
              ? null
              : (id) {
                  if (id != null) {
                    children.add(id);
                    onChanged();
                  }
                },
        ),
      ],
    );
  }
}

class StringMapFields extends StatefulWidget {
  const StringMapFields({
    super.key,
    required this.title,
    required this.values,
    required this.enabled,
    required this.onChanged,
  });
  final String title;
  final Map<String, String> values;
  final bool enabled;
  final VoidCallback onChanged;
  @override
  State<StringMapFields> createState() => _StringMapFieldsState();
}

class _MapRow {
  _MapRow(String key, String value)
    : key = TextEditingController(text: key),
      value = TextEditingController(text: value);
  final TextEditingController key;
  final TextEditingController value;
  void dispose() {
    key.dispose();
    value.dispose();
  }
}

class _StringMapFieldsState extends State<StringMapFields> {
  late final _rows = [
    for (final e in widget.values.entries) _MapRow(e.key, e.value),
  ];
  @override
  void dispose() {
    for (final row in _rows) {
      row.dispose();
    }
    super.dispose();
  }

  void _sync() {
    widget.values
      ..clear()
      ..addAll({for (final row in _rows) row.key.text: row.value.text});
    widget.onChanged();
  }

  @override
  Widget build(BuildContext context) => Column(
    crossAxisAlignment: CrossAxisAlignment.stretch,
    children: [
      const SizedBox(height: 16),
      Text(widget.title, style: Theme.of(context).textTheme.titleMedium),
      for (final row in _rows)
        Card(
          key: ObjectKey(row),
          child: Padding(
            padding: const EdgeInsets.all(8),
            child: Column(
              children: [
                TextFormField(
                  controller: row.key,
                  enabled: widget.enabled,
                  decoration: InputDecoration(labelText: '${widget.title} key'),
                  validator: (value) => value == null || value.isEmpty
                      ? 'Enter a key.'
                      : _rows.where((r) => r.key.text == value).length > 1
                      ? 'Keys must be unique.'
                      : null,
                  onChanged: (_) => _sync(),
                ),
                TextFormField(
                  controller: row.value,
                  enabled: widget.enabled,
                  decoration: InputDecoration(
                    labelText: '${widget.title} value',
                  ),
                  onChanged: (_) => _sync(),
                ),
                TextButton.icon(
                  onPressed: widget.enabled
                      ? () {
                          setState(() {
                            _rows.remove(row);
                          });
                          // Controllers may still be attached until this frame completes.
                          WidgetsBinding.instance.addPostFrameCallback(
                            (_) => row.dispose(),
                          );
                          _sync();
                        }
                      : null,
                  icon: const Icon(Icons.remove_circle_outline),
                  label: Text('Remove ${widget.title} entry'),
                ),
              ],
            ),
          ),
        ),
      TextButton.icon(
        onPressed: widget.enabled
            ? () {
                setState(() => _rows.add(_MapRow('', '')));
                _sync();
              }
            : null,
        icon: const Icon(Icons.add),
        label: Text('Add ${widget.title} entry'),
      ),
    ],
  );
}
