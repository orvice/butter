import 'package:butter_mobile/features/agents/agent_content_draft.dart';
import 'package:butter_mobile/gen/agents/v1/agent.pb.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  test('commitActions emits only changed Git files', () {
    final agent = Agent(
      agentId: 'a1',
      description: 'desc',
      config: AgentConfig(instruction: 'prompt', globalInstruction: 'global'),
    );
    final draft = AgentContentDraft.fromAgent('a1', agent);
    expect(draft.commitActions(), isEmpty);
    draft.prompt = 'new prompt';
    final actions = draft.commitActions();
    expect(actions, hasLength(1));
    expect(actions.single.path, 'agents/a1/prompt.md');
    expect(actions.single.content, 'new prompt');
  });
}
