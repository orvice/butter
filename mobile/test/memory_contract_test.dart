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

            final keepDraft = MemoryConfigDraft(base: owner.config)
              ..keyAction = MemoryApiKeyAction.keep;
            expect(
              await owner.saveConfig(keepDraft),
              isTrue,
              reason: owner.writeError,
            );
          },
        );
      },
    );
  }
}
