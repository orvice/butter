import 'package:butter_mobile/core/butter_api.dart';
import 'package:butter_mobile/core/server_address.dart';
import 'package:butter_mobile/features/memory/memory_controller.dart';
import 'package:butter_mobile/features/memory/memory_draft.dart';
import 'package:flutter_test/flutter_test.dart';

const workspace = 'memory';
const memberWorkspace = 'memory-member';

void main() {
  const server = String.fromEnvironment('BUTTER_TEST_SERVER');
  for (final prefix in ['', '/butter']) {
    group(
      'Memory production contract $prefix',
      skip: server.isEmpty ? 'Run tool/contract.sh.' : false,
      () {
        final base = '$server$prefix';
        late ButterApi api;
        late String token;

        setUp(() async {
          api = ButterApi(ServerAddress.parse(base, allowHttp: true));
          addTearDown(api.close);
          token = (await api.login(
            'mobile-test',
            'contract-test-password',
          )).token;
        });

        MemoryController mem({bool manage = true, String ws = workspace}) {
          final c = MemoryController(
            newApi: () => ButterApi(ServerAddress.parse(base, allowHttp: true)),
            token: token,
            workspaceId: ws,
            canManage: () => manage,
            onUnauthenticated: () {},
          );
          addTearDown(c.dispose);
          return c;
        }

        test(
          'delete config removes connection but not mem0 memories',
          skip: server.isEmpty
              ? 'Run tool/contract.sh.'
              : prefix.isNotEmpty
              ? 'Shared fixture; run at root prefix.'
              : false,
          () async {
            final owner = mem();
            await owner.loadConfig();
            final savedBase = owner.config!.baseUrl;
            await owner.loadMemories();
            expect(owner.memories.any((m) => m.id == 'w-old'), isTrue);
            expect(
              await owner.deleteConfig(),
              isTrue,
              reason: owner.writeError,
            );
            await owner.loadConfig();
            expect(owner.configEnabled, isFalse);

            final reconnect = MemoryConfigDraft()
              ..baseUrl = savedBase
              ..enabled = true
              ..keyAction = MemoryApiKeyAction.replace
              ..replacementKey = contractMem0Key;
            expect(
              await owner.saveConfig(reconnect),
              isTrue,
              reason: owner.writeError,
            );
            await owner.loadMemories();
            expect(owner.memories.any((m) => m.id == 'w-old'), isTrue);
          },
        );

        test(
          'connection, memories, and agent memory policy round-trip',
          skip: server.isEmpty
              ? 'Run tool/contract.sh.'
              : prefix.isNotEmpty
              ? 'Shared fixture; runs at root prefix.'
              : false,
          () async {
            final owner = mem();
            await owner.loadConfig();
            expect(owner.configEnabled, isTrue);

            await owner.testConnection();
            expect(owner.testOk, isTrue, reason: owner.testMessage);

            await owner.loadMemories();
            expect(owner.memories.any((m) => m.id == 'w-new'), isTrue);
            expect(
              owner.memories.every((m) => m.principal.isNotEmpty),
              isTrue,
              reason: 'owner sees principal provenance',
            );

            await owner.loadMemories(query: 'pnpm');
            expect(owner.memories.single.id, 'w-new');

            final member = mem(manage: false, ws: memberWorkspace);
            await member.loadMemories();
            expect(member.memories.any((m) => m.id == 'mm-1'), isTrue);
            expect(member.memories.every((m) => m.principal.isEmpty), isTrue);
            expect(await member.deleteMemory('mm-1'), isFalse);

            expect(await owner.deleteMemory('foreign-w'), isFalse);
            expect(await owner.deleteMemory('w-old'), isTrue);

            final keepDraft = MemoryConfigDraft(base: owner.config!)
              ..keyAction = MemoryApiKeyAction.keep;
            expect(
              await owner.saveConfig(keepDraft),
              isTrue,
              reason: owner.writeError,
            );
          },
        );

        test(
          'api key replace and clear round-trip',
          skip: server.isEmpty
              ? 'Run tool/contract.sh.'
              : prefix.isNotEmpty
              ? 'Shared fixture; run at root prefix.'
              : false,
          () async {
            final owner = mem();
            await owner.loadConfig();
            final cfg = owner.config!;
            final savedBase = cfg.baseUrl;
            final replace = MemoryConfigDraft(base: cfg)
              ..keyAction = MemoryApiKeyAction.replace
              ..replacementKey = contractMem0Key;
            expect(
              await owner.saveConfig(replace),
              isTrue,
              reason: owner.writeError,
            );
            await owner.testConnection();
            expect(owner.testOk, isTrue, reason: owner.testMessage);

            final clear = MemoryConfigDraft(base: owner.config!)
              ..enabled = false
              ..keyAction = MemoryApiKeyAction.clear;
            expect(
              await owner.saveConfig(clear),
              isTrue,
              reason: owner.writeError,
            );
            await owner.loadConfig();
            expect(owner.config!.credentialSet, isFalse);

            final restore = MemoryConfigDraft(base: owner.config!)
              ..baseUrl = savedBase
              ..enabled = true
              ..keyAction = MemoryApiKeyAction.replace
              ..replacementKey = contractMem0Key;
            expect(
              await owner.saveConfig(restore),
              isTrue,
              reason: owner.writeError,
            );
          },
        );

        test(
          'enabled put with rejected credentials refuses save',
          skip: server.isEmpty
              ? 'Run tool/contract.sh.'
              : prefix.isNotEmpty
              ? 'Shared fixture; run at root prefix.'
              : false,
          () async {
            final owner = mem();
            await owner.loadConfig();
            final bad = MemoryConfigDraft(base: owner.config!)
              ..keyAction = MemoryApiKeyAction.replace
              ..replacementKey = 'wrong-key';
            expect(await owner.saveConfig(bad), isFalse);
            expect(owner.writeError, isNotNull);
          },
        );

        test(
          'unreachable mem0 saves with warning',
          skip: server.isEmpty
              ? 'Run tool/contract.sh.'
              : prefix.isNotEmpty
              ? 'Shared fixture; run at root prefix.'
              : false,
          () async {
            final owner = mem();
            await owner.loadConfig();
            final savedBase = owner.config!.baseUrl;
            final draft = MemoryConfigDraft(base: owner.config!)
              ..baseUrl = 'http://127.0.0.1:1/'
              ..enabled = true
              ..keyAction = MemoryApiKeyAction.replace
              ..replacementKey = contractMem0Key;
            expect(
              await owner.saveConfig(draft),
              isTrue,
              reason: owner.writeError,
            );
            expect(owner.saveWarning, isNotEmpty);

            final restore = MemoryConfigDraft(base: owner.config!)
              ..baseUrl = savedBase
              ..enabled = true
              ..keyAction = MemoryApiKeyAction.replace
              ..replacementKey = contractMem0Key;
            expect(
              await owner.saveConfig(restore),
              isTrue,
              reason: owner.writeError,
            );
          },
        );
      },
    );
  }
}

const contractMem0Key = 'contract-mem0-key';
