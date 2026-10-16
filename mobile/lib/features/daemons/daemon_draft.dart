import '../../gen/agents/v1/dashboard.pb.dart';

/// One label of a runtime, as typed.
class LabelRow {
  LabelRow({this.key = '', this.value = ''});
  String key;
  String value;
  bool get isEmpty => key.trim().isEmpty && value.trim().isEmpty;
}

/// The editable form of a daemon runtime. It builds the complete message from
/// the stored copy it started from: labels, creation data, the Workspace and
/// fields this App does not know yet are carried over, because an update
/// replaces the whole object.
class DaemonRuntimeDraft {
  DaemonRuntimeDraft.create() : base = null {
    _initial = snapshot();
    _initialLabels = _labelSnapshot();
  }

  DaemonRuntimeDraft.edit(DaemonRuntime runtime)
    : base = runtime.deepCopy(),
      id = runtime.id,
      name = runtime.name,
      description = runtime.description {
    labels.addAll([
      for (final e
          in runtime.labels.entries.toList()
            ..sort((a, b) => a.key.compareTo(b.key)))
        LabelRow(key: e.key, value: e.value),
    ]);
    _initial = snapshot();
    _initialLabels = _labelSnapshot();
  }

  final DaemonRuntime? base;
  bool get isCreate => base == null;

  String id = '';
  String name = '';
  String description = '';

  /// The labels, one row each, in the order they are shown.
  final List<LabelRow> labels = [];

  late final String _initial;
  late final String _initialLabels;

  String _labelSnapshot() =>
      [for (final r in labels) '${r.key}\u0001${r.value}'].join('\u0000');

  String snapshot() => [id, name, description, _labelSnapshot()].join('\u0000');
  bool get dirty => snapshot() != _initial;

  String? validateId() {
    if (!isCreate) return null;
    final t = id.trim();
    if (t.isEmpty) return 'Enter an ID.';
    if (t.contains(RegExp(r'\s'))) return 'An ID cannot contain spaces.';
    return null;
  }

  String? validateName() => name.trim().isEmpty ? 'Enter a name.' : null;

  /// A label needs a key, and a key appears once.
  String? validateLabels() {
    final seen = <String>{};
    for (final r in labels) {
      if (r.isEmpty) continue;
      final k = r.key.trim();
      if (k.isEmpty) return 'Give each label a key, or remove the row.';
      if (!seen.add(k)) return 'The label key "$k" is used twice.';
    }
    return null;
  }

  String? firstError() => validateId() ?? validateName() ?? validateLabels();

  /// The complete message to store. Call only after [firstError] is null.
  DaemonRuntime build() {
    final r = base?.deepCopy() ?? DaemonRuntime();
    if (isCreate) r.id = id.trim();
    r.name = name.trim();
    final text = description.trim();
    if (text.isEmpty) {
      r.clearDescription();
    } else {
      r.description = text;
    }
    // Labels are edited as rows, so once touched the rows are the whole map;
    // an untouched set is left exactly as stored.
    if (_labelSnapshot() != _initialLabels) {
      r.labels
        ..clear()
        ..addEntries([
          for (final row in labels)
            if (!row.isEmpty) MapEntry(row.key.trim(), row.value.trim()),
        ]);
    }
    return r;
  }
}
