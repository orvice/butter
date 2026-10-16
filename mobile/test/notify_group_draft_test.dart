import 'package:butter_mobile/features/notify_groups/notify_group_draft.dart';
import 'package:butter_mobile/gen/agents/v1/agent.pb.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  test('rejects telegram target without destination id', () {
    final d = NotifyGroupDraft.create()
      ..name = 'ops'
      ..targets.add(NotifyTargetDraft());
    expect(d.validate(), contains('Telegram Destination'));
  });

  test('preserves metadata on edit', () {
    final g = NotifyGroup(name: 'ops', enabled: true)
      ..metadata['note'] = 'keep';
    final built = NotifyGroupDraft.edit(g).build();
    expect(built.metadata['note'], 'keep');
  });
}
