import 'package:butter_mobile/core/butter_api.dart';
import 'package:butter_mobile/core/server_address.dart';
import 'package:butter_mobile/gen/agents/v1/agent.pb.dart';
import 'package:connectrpc/connect.dart';
import 'package:flutter_test/flutter_test.dart';

Matcher code(Code code) =>
    isA<ConnectException>().having((e) => e.code, 'code', code);

void main() {
  const server = String.fromEnvironment('BUTTER_TEST_SERVER');
  for (final prefix in ['', '/butter']) {
    group(
      'Agent production Connect contract $prefix',
      skip: server.isEmpty ? 'Run tool/contract.sh.' : false,
      () {
        late ButterApi api;
        late String token;
        final tag = prefix.isEmpty ? 'root' : 'prefix';
        setUp(() async {
          api = ButterApi(
            ServerAddress.parse('$server$prefix', allowHttp: true),
          );
          addTearDown(api.close);
          token = (await api.login(
            'mobile-test',
            'contract-test-password',
          )).token;
        });
        test(
          'member creates and edits LLM without losing fields or resolving by name',
          () async {
            expect(await api.repositoryBound(token, 'alpha'), isFalse);
            final id = 'llm-$tag';
            final input = Agent(
              agentId: id,
              name: 'runtime-$tag',
              displayName: 'Visible',
              type: AgentType.AGENT_TYPE_LLM,
              labels: {'team': 'mobile'}.entries,
              metadata: {'note': 'keep'}.entries,
              enableA2a: true,
              enableOpenaiApi: true,
              enableAgui: true,
              config: AgentConfig(
                model: 'fast',
                instruction: 'Prompt',
                globalInstruction: 'Global',
                outputKey: 'answer',
                inputSchemaJson: '{"type":"object"}',
                outputSchemaJson: '{"type":"object"}',
                includeContents: LLMIncludeContents.LLM_INCLUDE_CONTENTS_NONE,
                disallowTransferToParent: true,
                disallowTransferToPeers: true,
                memory: MemoryConfig(topK: 1, threshold: 0),
                skills: ['existing-skill'],
                contextGuard: ContextGuardConfig(
                  strategy:
                      ContextGuardStrategy.CONTEXT_GUARD_STRATEGY_THRESHOLD,
                  maxTokens: 64000,
                ),
              ),
            );
            input.mergeFromBuffer([0xb0, 0x06, 9]);
            input.config.mergeFromBuffer([0xd0, 0x05, 7]);
            final created = await api.saveAgent(
              token,
              'alpha',
              input,
              create: true,
            );
            expect(
              created.lifecycleStatus,
              AgentLifecycleStatus.AGENT_LIFECYCLE_STATUS_ACTIVE,
            );
            expect(created.workspaceId, 'alpha');
            final read = await api.agent(token, 'alpha', id);
            final edited = read.deepCopy()..displayName = 'Renamed';
            edited.config.globalInstruction = 'Changed global';
            final saved = await api.saveAgent(
              token,
              'alpha',
              edited,
              create: false,
            );
            expect(saved.name, 'runtime-$tag');
            expect(saved.agentId, id);
            expect(saved.version > read.version, isTrue);
            final reread = await api.agent(token, 'alpha', id);
            expect(reread.displayName, 'Renamed');
            expect(reread.config.globalInstruction, 'Changed global');
            expect(reread.config.instruction, 'Prompt');
            expect(
              reread.config.includeContents,
              LLMIncludeContents.LLM_INCLUDE_CONTENTS_NONE,
            );
            expect(reread.config.outputSchemaJson, '{"type":"object"}');
            expect(reread.config.memory.hasTopK(), isTrue);
            expect(reread.config.memory.hasThreshold(), isTrue);
            expect(
              reread.unknownFields.getField(102)?.varints.single.toInt(),
              9,
            );
            expect(
              reread.config.unknownFields.getField(90)?.varints.single.toInt(),
              7,
            );
            expect(reread.config.skills, ['existing-skill']);
            expect(reread.labels, {'team': 'mobile'});
            expect(reread.metadata, {'note': 'keep'});
            expect((await api.agentStatus(token, 'alpha', id))?.agentId, id);
            await expectLater(
              api.agent(token, 'alpha', 'Renamed'),
              throwsA(code(Code.notFound)),
            );
            await expectLater(
              api.saveAgent(token, 'alpha', input, create: true),
              throwsA(code(Code.alreadyExists)),
            );
            await expectLater(
              api.agent(token, 'beta', id),
              throwsA(code(Code.notFound)),
            );
            await expectLater(
              api.saveAgent(token, 'foreign', input, create: true),
              throwsA(code(Code.permissionDenied)),
            );
          },
        );

        test(
          'composites retain ordered ID references and server relationship limits',
          () async {
            for (final type in [
              AgentType.AGENT_TYPE_LOOP,
              AgentType.AGENT_TYPE_SEQUENTIAL,
              AgentType.AGENT_TYPE_PARALLEL,
            ]) {
              final kind = type.value;
              final a = 'child-a-$kind-$tag', b = 'child-b-$kind-$tag';
              for (final id in [a, b]) {
                await api.saveAgent(
                  token,
                  'alpha',
                  Agent(agentId: id, name: id, type: AgentType.AGENT_TYPE_LLM),
                  create: true,
                );
              }
              final id = 'parent-$kind-$tag';
              final parent = await api.saveAgent(
                token,
                'alpha',
                Agent(
                  agentId: id,
                  name: id,
                  type: type,
                  childAgentIds: [a, b],
                  config: type == AgentType.AGENT_TYPE_LOOP
                      ? AgentConfig(maxIterations: 3)
                      : null,
                ),
                create: true,
              );
              expect(parent.childAgentIds, [a, b]);
              final reverse = parent.deepCopy()
                ..childAgentIds.setAll(0, [b, a]);
              await api.saveAgent(token, 'alpha', reverse, create: false);
              expect((await api.agent(token, 'alpha', id)).childAgentIds, [
                b,
                a,
              ]);
              expect((await api.agent(token, 'alpha', a)).agentId, a);
              final other = Agent(
                agentId: 'other-$kind-$tag',
                name: 'other-$kind-$tag',
                type: type,
                childAgentIds: [a],
              );
              await expectLater(
                api.saveAgent(token, 'alpha', other, create: true),
                throwsA(code(Code.invalidArgument)),
              );
              final self = parent.deepCopy()..childAgentIds.setAll(0, [id, a]);
              await expectLater(
                api.saveAgent(token, 'alpha', self, create: false),
                throwsA(code(Code.invalidArgument)),
              );
              final missing = parent.deepCopy()
                ..childAgentIds.setAll(0, ['missing', a]);
              await expectLater(
                api.saveAgent(token, 'alpha', missing, create: false),
                throwsA(code(Code.invalidArgument)),
              );
              final illegal = parent.deepCopy();
              illegal.ensureConfig().contextGuard = ContextGuardConfig(
                strategy: ContextGuardStrategy.CONTEXT_GUARD_STRATEGY_THRESHOLD,
              );
              await expectLater(
                api.saveAgent(token, 'alpha', illegal, create: false),
                throwsA(code(Code.invalidArgument)),
              );
            }
          },
        );

        test(
          'ContextGuard rejects wrong fields and enum; future include enum survives unrelated edit',
          () async {
            final id = 'guard-$tag';
            final input = Agent(
              agentId: id,
              name: id,
              type: AgentType.AGENT_TYPE_LLM,
              config: AgentConfig(
                contextGuard: ContextGuardConfig(
                  strategy: ContextGuardStrategy
                      .CONTEXT_GUARD_STRATEGY_SLIDING_WINDOW,
                  maxTurns: 10,
                ),
              ),
            );
            input.config.mergeFromBuffer([0xc8, 0x01, 99]);
            final created = await api.saveAgent(
              token,
              'alpha',
              input,
              create: true,
            );
            final edit = created.deepCopy()..displayName = 'Only label';
            final saved = await api.saveAgent(
              token,
              'alpha',
              edit,
              create: false,
            );
            expect(
              saved.config.unknownFields.getField(25)?.varints.single.toInt(),
              99,
            );
            final illegal = saved.deepCopy()..config.contextGuard.maxTokens = 1;
            await expectLater(
              api.saveAgent(token, 'alpha', illegal, create: false),
              throwsA(code(Code.invalidArgument)),
            );
            final negative = saved.deepCopy()
              ..config.contextGuard.maxTurns = -1;
            await expectLater(
              api.saveAgent(token, 'alpha', negative, create: false),
              throwsA(code(Code.invalidArgument)),
            );
            final unknown = saved.deepCopy()
              ..config.contextGuard.strategy =
                  ContextGuardStrategy.CONTEXT_GUARD_STRATEGY_UNSPECIFIED;
            unknown.config.contextGuard.mergeFromBuffer([8, 99]);
            await expectLater(
              api.saveAgent(token, 'alpha', unknown, create: false),
              throwsA(code(Code.invalidArgument)),
            );
          },
        );
      },
    );
  }
}
