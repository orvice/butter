import 'package:butter_mobile/features/linear/linear_draft.dart';
import 'package:butter_mobile/features/agents/leaf_draft.dart';
import 'package:butter_mobile/gen/agents/v1/linear.pb.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  test('run limit absent vs zero vs seconds', () {
    final draft = LinearAppDraft.create();
    expect(draft.runLimitMode, RunLimitMode.serverDefault);
    draft.runLimitMode = RunLimitMode.unlimited;
    expect(draft.buildForCreate().hasMaxRunSeconds(), isTrue);
    expect(draft.buildForCreate().maxRunSeconds, 0);
    draft.runLimitMode = RunLimitMode.serverDefault;
    expect(draft.buildForCreate().hasMaxRunSeconds(), isFalse);
  });

  test('clone preserves unknown fields on stored app', () {
    final app = LinearApp()
      ..displayName = 'x'
      ..agentId = 'a';
    final draft = LinearAppDraft.fromStored(app);
    draft.displayName = 'y';
    expect(draft.buildForUpdate().agentId, 'a');
  });

  test('parse allowed user ids', () {
    expect(LinearAppDraft.parseAllowedUserIds('a\nb, c'), ['a', 'b', 'c']);
  });
}
