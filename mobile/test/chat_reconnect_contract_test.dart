import 'package:butter_mobile/app/client_context.dart';
import 'package:butter_mobile/features/chat/chat_controller.dart';
import 'package:flutter_test/flutter_test.dart';

import 'support/memory_storage.dart';

void main() {
  const server = String.fromEnvironment('BUTTER_TEST_SERVER');
  for (final prefix in ['', '/butter']) {
    group(
      'Chat reconnect/stop contract $prefix',
      skip: server.isEmpty ? 'Run tool/contract.sh.' : false,
      () {
        final base = '$server$prefix';

        Future<ClientContext> signedIn() async {
          final client = ClientContext(
            storage: MemoryStorage(),
            allowHttp: true,
            chatIdleTimeout: const Duration(seconds: 30),
          );
          addTearDown(client.dispose);
          await client.setServer(base);
          await client.login('mobile-test', 'contract-test-password');
          await client.selectWorkspace('chat');
          return client;
        }

        test(
          'stop on an idle thread answers 204 without claiming stopped',
          () async {
            final chat = (await signedIn()).chat!;
            await chat.loadThreads();
            chat.startThread(
              chat.agents.firstWhere((a) => a.agentId == 'talker'),
            );
            await chat.stopRun();
            expect(chat.stopError, isNull);
            expect(chat.notice?.kind, isNot(NoticeKind.stopped));
          },
        );
      },
    );
  }
}
