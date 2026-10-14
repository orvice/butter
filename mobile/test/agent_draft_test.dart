import 'package:butter_mobile/features/agents/agent_draft.dart';
import 'package:butter_mobile/gen/agents/v1/agent.pb.dart';
import 'package:fixnum/fixnum.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  test('future types and lifecycle-blocked Agents cannot be edited', () {
    final future = Agent()..mergeFromBuffer([0xa0, 0x01, 99]);
    expect(canEditAgent(future), isFalse);
    expect(canEditAgent(Agent(type: AgentType.AGENT_TYPE_WORKFLOW)), isFalse);
    expect(canEditAgent(Agent(type: AgentType.AGENT_TYPE_PI)), isFalse);
    expect(
      canEditAgent(
        Agent(
          type: AgentType.AGENT_TYPE_LLM,
          lifecycleStatus: AgentLifecycleStatus.AGENT_LIFECYCLE_STATUS_DELETED,
        ),
      ),
      isFalse,
    );
    expect(canEditAgent(Agent(type: AgentType.AGENT_TYPE_LLM)), isTrue);
  });

  test('ContextGuard mode edits clear only the other strategy fields', () {
    final draft = AgentDraft(
      base: Agent(
        type: AgentType.AGENT_TYPE_LLM,
        config: AgentConfig(
          contextGuard: ContextGuardConfig(
            strategy: ContextGuardStrategy.CONTEXT_GUARD_STRATEGY_THRESHOLD,
            maxTokens: 64000,
          )..mergeFromBuffer([0x78, 8]),
        ),
      ),
    );
    draft.setGuard(GuardMode.slidingWindow);
    expect(draft.agent.config.contextGuard.hasMaxTokens(), isFalse);
    draft.agent.config.contextGuard.maxTurns = 12;
    expect(draft.build().config.contextGuard.maxTurns, 12);
    expect(
      draft
          .build()
          .config
          .contextGuard
          .unknownFields
          .getField(15)
          ?.varints
          .single
          .toInt(),
      8,
    );
    draft.setGuard(GuardMode.off);
    expect(draft.build().config.hasContextGuard(), isFalse);
  });

  test(
    'changing LLM to Loop clears LLM-only behavior without losing associations',
    () {
      final draft = AgentDraft(
        base: Agent(
          type: AgentType.AGENT_TYPE_LLM,
          config: AgentConfig(
            instruction: 'prompt',
            model: 'alias',
            skills: ['skill'],
            mcpServerIds: ['mcp'],
            contextGuard: ContextGuardConfig(maxTokens: 1),
          ),
        ),
      );
      draft.setType(AgentType.AGENT_TYPE_LOOP);
      draft.agent.config.maxIterations = 4;
      expect(draft.build().config.hasInstruction(), isFalse);
      expect(draft.build().config.hasModel(), isFalse);
      expect(draft.build().config.hasContextGuard(), isFalse);
      expect(draft.build().config.skills, ['skill']);
      expect(draft.build().config.mcpServerIds, ['mcp']);
      expect(draft.build().config.maxIterations, 4);
    },
  );

  test(
    'editing a label preserves optional, unknown enum and nested fields',
    () {
      final config = AgentConfig(
        skills: ['existing-skill'],
        memory: MemoryConfig(enabled: true, topK: 0),
      );
      // include_contents = future enum 99; future config field 90 = 7.
      config.mergeFromBuffer([0xc8, 0x01, 99, 0xd0, 0x05, 7]);
      final base = Agent(
        agentId: 'stable-id',
        name: 'runtime-name',
        displayName: 'Before',
        config: config,
        version: Int64.parseInt('9007199254740993'),
        type: AgentType.AGENT_TYPE_LLM,
      )..mergeFromBuffer([0xb0, 0x06, 9]);
      final draft = AgentDraft(base: base);
      draft.agent.displayName = 'After';
      final saved = Agent.fromBuffer(draft.build().writeToBuffer());
      expect(saved.displayName, 'After');
      expect(saved.agentId, 'stable-id');
      expect(saved.version.toString(), '9007199254740993');
      expect(saved.config.writeToBuffer(), config.writeToBuffer());
      expect(saved.unknownFields.getField(102)?.varints.single.toInt(), 9);
      expect(saved.config.memory.hasTopK(), isTrue);
      expect(base.displayName, 'Before');
    },
  );
}
