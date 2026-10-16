import 'package:butter_mobile/core/memory_errors.dart';
import 'package:butter_mobile/features/agents/agent_draft.dart';
import 'package:butter_mobile/features/memory/agent_memory_config.dart';
import 'package:butter_mobile/features/memory/memory_draft.dart';
import 'package:butter_mobile/gen/agents/v1/agent.pb.dart';
import 'package:butter_mobile/gen/agents/v1/workspace_memory.pb.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  test('api key keep omits optional field', () {
    final draft =
        MemoryConfigDraft(base: WorkspaceMemoryConfig(credentialSet: true))
          ..baseUrl = 'http://mem0.test'
          ..keyAction = MemoryApiKeyAction.keep;
    expect(draft.buildPut().hasApiKey(), isFalse);
  });

  test('api key clear sends empty string', () {
    final draft = MemoryConfigDraft(
      base: WorkspaceMemoryConfig(credentialSet: true),
    )..keyAction = MemoryApiKeyAction.clear;
    expect(draft.buildPut().apiKey, '');
  });

  test('typed secret clears after submission', () {
    final draft = MemoryConfigDraft()
      ..keyAction = MemoryApiKeyAction.replace
      ..replacementKey = 'secret';
    draft.clearTypedSecret();
    expect(draft.replacementKey, isEmpty);
    expect(draft.keyAction, MemoryApiKeyAction.keep);
  });

  test('memory config disabled clears message', () {
    final agent = Agent(
      type: AgentType.AGENT_TYPE_LLM,
      config: AgentConfig(memory: MemoryConfig()..enabled = true),
    );
    final draft = AgentDraft(base: agent);
    applyMemoryForm(draft, AgentMemoryFormValues());
    expect(draft.agent.config.hasMemory(), isFalse);
  });

  test('classifyMem0ProbeError hides raw credential payloads', () {
    expect(
      classifyMem0ProbeError('HTTP 401: {"detail":"secret-key"}'),
      'The mem0 server rejected the API key.',
    );
  });

  test('pi agent drops memory and cards on create build', () {
    final agent = Agent(
      type: AgentType.AGENT_TYPE_PI,
      config: AgentConfig(
        pi: PiAgentConfig(butterboxId: 'b'),
        memory: MemoryConfig()..enabled = true,
        resultCards: ResultCardConfig(
          generation: ResultCardGeneration.RESULT_CARD_GENERATION_DISABLED,
        ),
      ),
    );
    final built = AgentDraft(base: agent).buildForCreate();
    expect(built.config.hasMemory(), isFalse);
    expect(built.config.hasResultCards(), isFalse);
  });
}
