import 'package:butter_mobile/features/daemons/daemon_draft.dart';
import 'package:butter_mobile/gen/agents/v1/dashboard.pb.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  test('a new runtime needs an ID without spaces and a name', () {
    final d = DaemonRuntimeDraft.create();
    expect(d.validateId(), isNotNull);
    d.id = 'my rt';
    expect(d.validateId(), contains('spaces'));
    d.id = 'rt1';
    expect(d.validateId(), isNull);
    expect(d.validateName(), isNotNull);
    d.name = '  ';
    expect(d.firstError(), isNotNull);
    d.name = ' One ';
    expect(d.firstError(), isNull);
    final built = d.build();
    expect((built.id, built.name), ('rt1', 'One'));
  });

  test('an edit keeps the ID and carries fields it does not edit', () {
    final stored = DaemonRuntime(id: 'rt1', name: 'One', description: 'old')
      ..labels['team'] = 'platform'
      ..createdBy = 'bob'
      ..workspaceId = 'alpha';
    final d = DaemonRuntimeDraft.edit(stored);
    expect(d.dirty, isFalse);
    d.id = 'other';
    d.name = 'Two';
    d.description = '  ';
    expect(d.dirty, isTrue);
    final built = d.build();
    expect(built.id, 'rt1');
    expect(built.name, 'Two');
    expect(built.hasDescription(), isFalse);
    expect(built.labels['team'], 'platform');
    expect(built.createdBy, 'bob');
    expect(built.workspaceId, 'alpha');
    // The stored copy is never mutated.
    expect(stored.description, 'old');
  });

  group('labels', () {
    DaemonRuntime stored() => DaemonRuntime(id: 'rt1', name: 'One')
      ..labels['team'] = 'platform'
      ..labels['env'] = 'prod'
      ..createdBy = 'bob';

    test('an edit shows the stored labels, sorted, and is not dirty', () {
      final d = DaemonRuntimeDraft.edit(stored());
      expect(d.labels.map((r) => r.key), ['env', 'team']);
      expect(d.dirty, isFalse);
      expect(d.build().labels, {'team': 'platform', 'env': 'prod'});
    });

    test('add, change and remove labels', () {
      final d = DaemonRuntimeDraft.edit(stored());
      d.labels.firstWhere((r) => r.key == 'env').value = 'staging';
      d.labels.removeWhere((r) => r.key == 'team');
      d.labels.add(LabelRow(key: ' region ', value: ' eu '));
      expect(d.dirty, isTrue);
      expect(d.firstError(), isNull);
      final built = d.build();
      expect(built.labels, {'env': 'staging', 'region': 'eu'});
      expect(built.createdBy, 'bob');
    });

    test('an empty row is ignored; a blank key with a value is refused', () {
      final d = DaemonRuntimeDraft.create()
        ..id = 'rt1'
        ..name = 'One';
      d.labels.add(LabelRow());
      expect(d.firstError(), isNull);
      expect(d.build().labels, isEmpty);
      d.labels.add(LabelRow(key: '  ', value: 'x'));
      expect(d.firstError(), contains('label'));
    });

    test('a repeated key is refused', () {
      final d = DaemonRuntimeDraft.edit(stored());
      d.labels.add(LabelRow(key: 'team ', value: 'other'));
      expect(d.validateLabels(), contains('team'));
      expect(d.firstError(), isNotNull);
    });

    test('removing every label stores none', () {
      final d = DaemonRuntimeDraft.edit(stored());
      d.labels.clear();
      expect(d.dirty, isTrue);
      expect(d.build().labels, isEmpty);
    });
  });
}
