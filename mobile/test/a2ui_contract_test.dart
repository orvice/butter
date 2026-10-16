import 'package:butter_mobile/app/client_context.dart';
import 'package:butter_mobile/core/agui_api.dart';
import 'package:butter_mobile/core/server_address.dart';
import 'package:butter_mobile/features/a2ui/protocol.dart';
import 'package:butter_mobile/features/chat/chat_controller.dart';
import 'package:flutter_test/flutter_test.dart';

import 'support/memory_storage.dart';

Future<void> until(bool Function() done, {String? reason}) async {
  final stop = DateTime.now().add(const Duration(seconds: 30));
  while (!done()) {
    if (DateTime.now().isAfter(stop)) fail(reason ?? 'condition never held');
    await Future<void>.delayed(const Duration(milliseconds: 30));
  }
}

void main() {
  const server = String.fromEnvironment('BUTTER_TEST_SERVER');
  group(
    'A2UI production AG-UI contract',
    skip: server.isEmpty ? 'Run tool/contract.sh.' : false,
    () {
      const renderUiMessage =
          r'tool:render_ui {"fallback":"Card fallback","messages":[{"updateComponents":{"components":[{"id":"root","component":"Text","text":"Contract card"}]}}]}';

      Future<ChatController> startedChat() async {
        final client = ClientContext(
          storage: MemoryStorage(),
          allowHttp: true,
          chatIdleTimeout: const Duration(seconds: 45),
        );
        addTearDown(client.dispose);
        await client.setServer(server);
        await client.login('mobile-test', 'contract-test-password');
        await client.selectWorkspace('chat');
        final chat = client.chat!;
        await chat.loadThreads();
        chat.startThread(chat.agents.firstWhere((a) => a.agentId == 'talker'));
        return chat;
      }

      test('render_ui creates a card restored from UI snapshot', () async {
        final client = ClientContext(storage: MemoryStorage(), allowHttp: true);
        addTearDown(client.dispose);
        await client.setServer(server);
        final chat = await startedChat();
        final threadId = chat.threadId!;
        final agentId = chat.agentId!;
        expect(await chat.send(renderUiMessage), isTrue);
        await until(
          () => chat.phase == ChatPhase.idle,
          reason: 'card run ended',
        );
        expect(chat.surfaces.entries.any((e) => !e.deleted), isTrue);

        await chat.refresh();
        expect(chat.openError, isNull);
        expect(chat.surfaces.entries.any((e) => !e.deleted), isTrue);

        final agui = AguiApi(ServerAddress.parse(server, allowHttp: true));
        addTearDown(agui.close);
        final uiBody = await agui.uiSnapshot(
          token: chat.token,
          workspace: chat.workspaceId,
          agentId: agentId,
          threadId: threadId,
        );
        final snapshot = parseUiSnapshot(uiBody);
        expect(snapshot, isNotNull);
        expect(snapshot!.surfaces, isNotEmpty);
      });
    },
  );
}
