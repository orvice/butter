import 'package:butter_mobile/features/agents/agent_draft.dart';
import 'package:butter_mobile/features/agents/leaf_draft.dart';
import 'package:butter_mobile/gen/agents/v1/agent.pb.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  group('run limit', () {
    test('absent, explicit zero and explicit seconds stay different', () {
      final pi = PiAgentConfig(butterboxId: 'b');
      expect(piRunLimit(pi), RunLimitMode.serverDefault);
      expect(pi.hasMaxRunSeconds(), isFalse);

      pi.maxRunSeconds = 0;
      expect(piRunLimit(pi), RunLimitMode.unlimited);
      expect(pi.hasMaxRunSeconds(), isTrue);

      pi.maxRunSeconds = 90;
      expect(piRunLimit(pi), RunLimitMode.seconds);

      final cursor = CursorAgentConfig()..maxRunSeconds = 0;
      expect(cursorRunLimit(cursor), RunLimitMode.unlimited);
    });

    test('applying a mode writes exactly its wire form', () {
      final pi = PiAgentConfig(maxRunSeconds: 60);
      void apply(RunLimitMode mode, [int n = 0]) => applyRunLimit(
        mode,
        n,
        clear: pi.clearMaxRunSeconds,
        set: (v) => pi.maxRunSeconds = v,
      );
      apply(RunLimitMode.unlimited);
      expect(
        PiAgentConfig.fromBuffer(pi.writeToBuffer()).hasMaxRunSeconds(),
        isTrue,
      );
      expect(PiAgentConfig.fromBuffer(pi.writeToBuffer()).maxRunSeconds, 0);
      apply(RunLimitMode.serverDefault);
      expect(
        PiAgentConfig.fromBuffer(pi.writeToBuffer()).hasMaxRunSeconds(),
        isFalse,
      );
      apply(RunLimitMode.seconds, 45);
      expect(PiAgentConfig.fromBuffer(pi.writeToBuffer()).maxRunSeconds, 45);
    });

    test('seconds must be a positive whole number within int32', () {
      for (final bad in ['', ' ', '0', '-1', '1.5', 'abc', '2147483648']) {
        expect(validateRunSeconds(bad), isNotNull, reason: bad);
      }
      expect(validateRunSeconds('1'), isNull);
      expect(validateRunSeconds(' 1800 '), isNull);
      expect(validateRunSeconds('2147483647'), isNull);
      expect(parseRunSeconds('120'), 120);
    });
  });

  group('type switching', () {
    AgentDraft llmDraft() => AgentDraft(
      base: Agent(
        agentId: 'a',
        name: 'n',
        type: AgentType.AGENT_TYPE_LLM,
        childAgentIds: ['child'],
        config: AgentConfig(
          model: 'm',
          instruction: 'do it',
          globalInstruction: 'g',
          skills: ['writing'],
          mcpServerIds: ['mcp'],
          remoteAgentIds: ['remote'],
        ),
      ),
    );

    test('Pi and Cursor are editable types', () {
      expect(editableAgentTypes, contains(AgentType.AGENT_TYPE_PI));
      expect(editableAgentTypes, contains(AgentType.AGENT_TYPE_CURSOR));
      expect(
        canEditAgent(
          Agent(
            agentId: 'a',
            type: AgentType.AGENT_TYPE_PI,
            config: AgentConfig(pi: PiAgentConfig(butterboxId: 'b')),
          ),
        ),
        isTrue,
      );
    });

    test('becoming a leaf drops every field the server would reject', () {
      final draft = llmDraft()..setType(AgentType.AGENT_TYPE_PI);
      final built = draft.build();
      expect(built.type, AgentType.AGENT_TYPE_PI);
      expect(built.childAgentIds, isEmpty);
      expect(built.config.hasPi(), isTrue);
      expect(built.config.hasModel(), isFalse);
      expect(built.config.instruction, isEmpty);
      expect(built.config.globalInstruction, isEmpty);
      expect(built.config.skills, isEmpty);
      expect(built.config.mcpServerIds, isEmpty);
      expect(built.config.remoteAgentIds, isEmpty);
      expect(built.config.hasMemory(), isFalse);
      expect(built.config.hasResultCards(), isFalse);
      expect(built.config.hasContextGuard(), isFalse);
    });

    test('Pi to Cursor leaves no Pi binding behind', () {
      final draft = AgentDraft(
        base: Agent(
          agentId: 'a',
          type: AgentType.AGENT_TYPE_PI,
          config: AgentConfig(
            pi: PiAgentConfig(butterboxId: 'b', model: 'x', maxRunSeconds: 0),
          ),
        ),
      )..setType(AgentType.AGENT_TYPE_CURSOR);
      final built = draft.build();
      expect(built.config.hasPi(), isFalse);
      expect(built.config.hasCursor(), isTrue);
      expect(built.config.cursor.butterboxId, isEmpty);
      expect(built.config.cursor.hasMaxRunSeconds(), isFalse);
    });

    test('leaving a leaf for LLM drops its binding', () {
      final draft = AgentDraft(
        base: Agent(
          agentId: 'a',
          type: AgentType.AGENT_TYPE_CURSOR,
          config: AgentConfig(cursor: CursorAgentConfig(butterboxId: 'b')),
        ),
      )..setType(AgentType.AGENT_TYPE_LLM);
      expect(draft.build().config.hasCursor(), isFalse);
    });

    test('editing a leaf keeps unknown fields of the binding', () {
      final pi = PiAgentConfig(butterboxId: 'b');
      // Field 99 is not known to this client.
      pi.mergeFromBuffer([0x98, 0x06, 7]);
      final draft = AgentDraft(
        base: Agent(
          agentId: 'a',
          type: AgentType.AGENT_TYPE_PI,
          config: AgentConfig(pi: pi),
        ),
      );
      draft.config.ensurePi().model = 'sonnet';
      final built = draft.build();
      expect(built.config.pi.model, 'sonnet');
      expect(built.config.pi.unknownFields.isNotEmpty, isTrue);
    });
  });
}
