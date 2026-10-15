import 'dart:convert';
import 'dart:typed_data';

import 'package:butter_mobile/app/client_context.dart';
import 'package:butter_mobile/core/agui_api.dart';
import 'package:butter_mobile/core/butter_api.dart';
import 'package:butter_mobile/core/server_address.dart';
import 'package:butter_mobile/features/chat/chat_controller.dart';
import 'package:butter_mobile/features/chat/chat_model.dart';
import 'package:butter_mobile/features/chat/run_request.dart';
import 'package:flutter_test/flutter_test.dart';

import 'support/memory_storage.dart';

final png = base64Decode(
  'iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNkYPhfDwAChwGA60e6kgAAAABJRU5ErkJggg==',
);

Future<void> until(bool Function() done, {String? reason}) async {
  final stop = DateTime.now().add(const Duration(seconds: 20));
  while (!done()) {
    if (DateTime.now().isAfter(stop)) fail(reason ?? 'condition never held');
    await Future<void>.delayed(const Duration(milliseconds: 20));
  }
}

List<String> texts(ChatController chat) => [
  for (final m in chat.messages) '${m.role.name}: ${m.text}',
];

void main() {
  const server = String.fromEnvironment('BUTTER_TEST_SERVER');
  for (final prefix in ['', '/butter']) {
    group(
      'Chat production AG-UI and Session contract $prefix',
      skip: server.isEmpty ? 'Run tool/contract.sh.' : false,
      () {
        final base = '$server$prefix';

        Future<ClientContext> signedIn([String workspace = 'chat']) async {
          final client = ClientContext(
            storage: MemoryStorage(),
            allowHttp: true,
            chatIdleTimeout: const Duration(seconds: 30),
          );
          addTearDown(client.dispose);
          await client.setServer(base);
          await client.login('mobile-test', 'contract-test-password');
          await client.selectWorkspace(workspace);
          expect(client.error, isNull);
          return client;
        }

        Future<ChatController> startedChat(ClientContext client) async {
          final chat = client.chat!;
          await chat.loadThreads();
          expect(chat.threadsError, isNull);
          chat.startThread(
            chat.agents.firstWhere((a) => a.agentId == 'talker'),
          );
          return chat;
        }

        Future<void> sendAndWait(
          ChatController chat,
          String text, [
          List<ChatImage> images = const [],
        ]) async {
          expect(await chat.send(text, images), isTrue, reason: chat.sendError);
          await until(() => chat.phase == ChatPhase.idle, reason: 'run ended');
        }

        test(
          'a new thread runs detached and its history restores in a fresh session',
          () async {
            final chat = await startedChat(await signedIn());
            final threadId = chat.threadId!;
            // The production handler detaches only on a thread with a UI binding
            // and with its lease, run state and Run Log wired, so a clean run
            // proves the request was accepted as a Detached Run.
            await sendAndWait(chat, 'hello');
            expect(texts(chat), ['user: hello', 'assistant: fixture(hello)']);
            expect(chat.notice, isNull);

            await sendAndWait(chat, 'second');
            expect(texts(chat), [
              'user: hello',
              'assistant: fixture(hello)',
              'user: second',
              'assistant: fixture(second)',
            ]);

            await chat.loadThreads();
            final row = chat.threads.firstWhere((t) => t.threadId == threadId);
            expect(row.agentId, 'talker');

            final other = (await signedIn()).chat!;
            await other.loadThreads();
            final again = other.threads.firstWhere(
              (t) => t.threadId == threadId,
            );
            await other.openThread(again);
            expect(other.openError, isNull);
            expect(texts(other), texts(chat));
            expect(other.notice, isNull);
          },
        );

        test('Markdown comes back whole', () async {
          final chat = await startedChat(await signedIn());
          await sendAndWait(chat, 'markdown please');
          final reply = chat.messages.last.text;
          expect(reply, contains('# Fixture'));
          expect(reply, contains('| a | b |'));
          expect(reply, contains('```go'));
        });

        test('an image is accepted and restored from the history', () async {
          final client = await signedIn();
          final chat = await startedChat(client);
          final threadId = chat.threadId!;
          final tray = ImageTray()..add(ChatImage.fromBytes('a.png', png));
          await sendAndWait(chat, 'look', tray.images);
          expect(chat.messages.last.text, 'fixture(look)');

          final other = (await signedIn()).chat!;
          await other.loadThreads();
          await other.openThread(
            other.threads.firstWhere((t) => t.threadId == threadId),
          );
          final user = other.messages.first;
          expect(user.parts.whereType<ImagePart>(), hasLength(1));
          expect(user.parts.whereType<TextPart>().single.text, 'look');
        });

        test(
          'rename is normalized by the server and shows in the list',
          () async {
            final chat = await startedChat(await signedIn());
            final threadId = chat.threadId!;
            await sendAndWait(chat, 'name me');
            await chat.loadThreads();
            final row = chat.threads.firstWhere((t) => t.threadId == threadId);
            expect(await chat.renameThread(row, '  Plan   B '), isTrue);
            expect(
              chat.threads.firstWhere((t) => t.threadId == threadId).title,
              'Plan B',
            );
            await chat.loadThreads();
            expect(
              chat.threads.firstWhere((t) => t.threadId == threadId).title,
              'Plan B',
            );
          },
        );

        test(
          'another Workspace sees none of this Workspace\'s threads or agents',
          () async {
            final chat = await startedChat(await signedIn());
            final threadId = chat.threadId!;
            await sendAndWait(chat, 'private');

            final alpha = (await signedIn('alpha')).chat!;
            await alpha.loadThreads();
            expect(
              alpha.threads.map((t) => t.threadId),
              isNot(contains(threadId)),
            );
            expect(
              alpha.agents.map((a) => a.agentId),
              isNot(contains('talker')),
            );
          },
        );

        test(
          'a thread without a UI binding is not listed and a detached run on it is refused',
          () async {
            final client = await signedIn();
            final chat = client.chat!;
            await chat.loadThreads();
            expect(
              chat.threads.map((t) => t.threadId),
              isNot(contains('legacy')),
            );

            final address = ServerAddress.parse(base, allowHttp: true);
            final api = ButterApi(address);
            addTearDown(api.close);
            final token = (await api.login(
              'mobile-test',
              'contract-test-password',
            )).token;
            final agui = AguiApi(address);
            addTearDown(agui.close);
            final body = buildRunRequest(
              threadId: 'legacy',
              runId: 'run-legacy',
              messageId: 'm-legacy',
              text: 'hi',
              images: const [],
            );
            await expectLater(
              agui.run(
                token: token,
                workspace: 'chat',
                agentId: 'talker',
                body: body,
              ),
              throwsA(
                isA<AguiHttpError>()
                    .having((e) => e.status, 'status', 400)
                    .having((e) => e.code, 'code', 'thread_unbound'),
              ),
            );
          },
        );

        test(
          'a repeated run id is refused by the server, never run twice',
          () async {
            final client = await signedIn();
            final chat = await startedChat(client);
            final threadId = chat.threadId!;
            await sendAndWait(chat, 'once');

            final address = ServerAddress.parse(base, allowHttp: true);
            final api = ButterApi(address);
            addTearDown(api.close);
            final token = (await api.login(
              'mobile-test',
              'contract-test-password',
            )).token;
            final agui = AguiApi(address);
            addTearDown(agui.close);
            Uint8List request() => buildRunRequest(
              threadId: threadId,
              runId: 'fixed-run',
              messageId: 'fixed-message',
              text: 'again',
              images: const [],
            );
            final first = await agui.run(
              token: token,
              workspace: 'chat',
              agentId: 'talker',
              body: request(),
            );
            await first.events.drain<void>();
            first.abort();
            await expectLater(
              agui.run(
                token: token,
                workspace: 'chat',
                agentId: 'talker',
                body: request(),
              ),
              throwsA(
                isA<AguiHttpError>()
                    .having((e) => e.status, 'status', 409)
                    .having((e) => e.code, 'code', 'run_exists'),
              ),
            );
          },
        );
      },
    );
  }
}
