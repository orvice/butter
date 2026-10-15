import 'dart:async';
import 'dart:convert';
import 'dart:typed_data';

import 'package:butter_mobile/app/client_context.dart';
import 'package:butter_mobile/features/chat/chat_controller.dart';
import 'package:butter_mobile/features/chat/chat_model.dart';
import 'package:butter_mobile/features/chat/run_request.dart';
import 'package:butter_mobile/gen/agents/v1/agent_service.pb.dart';
import 'package:flutter_test/flutter_test.dart';

import 'support/chat_fixtures.dart';
import 'support/chat_rig.dart';
import 'support/chat_server.dart';

/// 1x1 PNG.
final png = Uint8List.fromList(
  base64Decode(
    'iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNkYPhfDwAChwGA60e6kgAAAABJRU5ErkJggg==',
  ),
);

Future<void> until(bool Function() done, {String? reason}) async {
  final stop = DateTime.now().add(const Duration(seconds: 5));
  while (!done()) {
    if (DateTime.now().isAfter(stop)) fail(reason ?? 'condition never held');
    await Future<void>.delayed(const Duration(milliseconds: 10));
  }
}

Future<void> idle(ChatController chat) =>
    until(() => chat.phase == ChatPhase.idle, reason: 'run never ended');

/// A signed-in rig with one agent, an open draft on it and a controller.
Future<(Rig, ChatController)> draft({
  Duration idleTimeout = const Duration(seconds: 60),
}) async {
  final r = await rig(idleTimeout: idleTimeout);
  final chat = r.client.chat!;
  await chat.loadThreads();
  chat.startThread(chat.agents.single);
  return (r, chat);
}

List<String> texts(ChatController chat) => [
  for (final m in chat.messages) '${m.role.name}: ${m.text}',
];

void main() {
  group('sending', () {
    test(
      'a new thread runs detached under the agent and shows the reply',
      () async {
        final (r, chat) = await draft();
        final threadId = chat.threadId!;
        expect(
          r.chat.posts,
          isEmpty,
          reason: 'a draft sends nothing by itself',
        );

        expect(await chat.send('hi'), isTrue);
        await idle(chat);

        final post = r.chat.posts.single;
        expect(post.agentId, 'shared');
        expect(post.token, startsWith('Bearer '));
        expect(post.workspace, 'alpha');
        expect(post.threadId, threadId);
        expect(post.body['forwardedProps'], {
          'butterRun': {'detach': true},
        });
        expect(post.content, 'hi');
        expect(texts(chat), ['user: hi', 'assistant: ok']);
        expect(chat.liveReply, isNull);
        expect(chat.notice, isNull);
        expect(chat.sendError, isNull);
      },
    );

    test('the reply is visible while it streams', () async {
      final (r, chat) = await draft();
      final more = Completer<void>();
      r.chat.script = (post, res) async {
        await res.event(runStarted(post.threadId, post.runId));
        await res.event(textStart('m1'));
        await res.event(textDelta('m1', 'Hel'));
        await res.hold(more.future);
        await res.event(textDelta('m1', 'lo'));
        await res.event(textEnd('m1'));
        await res.event(runFinished(post.threadId, post.runId));
        await res.finish();
      };
      await chat.send('hi');
      await until(() => chat.liveReply?.text == 'Hel');
      expect(chat.phase, ChatPhase.streaming);
      expect(chat.canSend, isFalse, reason: 'one run per thread at a time');

      more.complete();
      await idle(chat);
      expect(texts(chat), ['user: hi', 'assistant: Hello']);
    });

    test(
      'a later message sends only itself, never the history or old images',
      () async {
        final (r, chat) = await draft();
        final tray = ImageTray()..add(ChatImage.fromBytes('a.png', png));
        await chat.send('first', tray.images);
        await idle(chat);
        await chat.send('second');
        await idle(chat);

        expect(r.chat.posts, hasLength(2));
        final first = r.chat.posts[0], second = r.chat.posts[1];
        expect(first.content, isA<List>());
        expect(
          (first.content! as List).map((p) => (p as Map)['type']),
          containsAll(['text', 'image']),
        );
        expect(second.messages, hasLength(1));
        expect(second.content, 'second');
        expect(second.threadId, first.threadId);
        expect(second.runId, isNot(first.runId));
        expect(jsonEncode(second.body), isNot(contains('image')));
      },
    );

    test('text no one can send is refused before any request', () async {
      final (r, chat) = await draft();
      expect(await chat.send('   '), isFalse);
      expect(chat.sendError, isNotNull);
      expect(r.chat.posts, isEmpty);
    });
  });

  group('leaving and failing', () {
    test(
      'leaving the page only detaches: no stop, no retry, no new request',
      () async {
        final (r, chat) = await draft();
        final never = Completer<void>();
        r.chat.script = (post, res) async {
          await res.event(runStarted(post.threadId, post.runId));
          await res.event(textDelta('m1', 'x'));
          await res.hold(never.future);
        };
        await chat.send('hi');
        await until(() => chat.phase == ChatPhase.streaming);

        chat.leave();
        expect(chat.threadId, isNull);
        expect(chat.phase, ChatPhase.idle);
        await r.chat.posts.single.disconnected.future.timeout(
          const Duration(seconds: 5),
        );
        await Future<void>.delayed(const Duration(milliseconds: 100));
        expect(r.chat.stops, isEmpty);
        expect(r.chat.posts, hasLength(1));
        never.complete();
      },
    );

    test(
      'a stream that drops is a lost connection, not a stop, and is not retried',
      () async {
        final (r, chat) = await draft();
        r.chat.script = (post, res) async {
          await res.event(runStarted(post.threadId, post.runId));
          await res.event(textStart('m1'));
          await res.event(textDelta('m1', 'part'));
          await res.drop();
        };
        await chat.send('hi');
        await idle(chat);

        expect(chat.notice?.kind, NoticeKind.connectionLost);
        expect(texts(chat), ['user: hi', 'assistant: part']);
        expect(r.chat.posts, hasLength(1));
        expect(r.chat.stops, isEmpty);
        expect(r.chat.historyReads, isEmpty, reason: 'nothing is read unasked');
      },
    );

    test(
      'a request that may have arrived is marked unconfirmed and never resent',
      () async {
        final (r, chat) = await draft();
        r.chat.script = (post, res) async => res.drop();
        expect(
          await chat.send('hi'),
          isTrue,
          reason: 'the turn left the composer',
        );
        await idle(chat);

        expect(chat.notice?.kind, NoticeKind.unconfirmed);
        expect(texts(chat), ['user: hi']);
        expect(r.chat.posts, hasLength(1));
        await Future<void>.delayed(const Duration(milliseconds: 150));
        expect(r.chat.posts, hasLength(1));
      },
    );

    test(
      'refresh reads the thread once and shows what the server kept',
      () async {
        final (r, chat) = await draft();
        r.chat.script = (post, res) async => res.drop();
        await chat.send('hi');
        await idle(chat);
        r.chat.histories['shared/${chat.threadId}'] = jsonEncode({
          'threadId': chat.threadId,
          'messages': [
            {'id': 'u1', 'role': 'user', 'content': 'hi'},
            {'id': 'a1', 'role': 'assistant', 'content': 'hello'},
          ],
        });

        await chat.refresh();
        expect(r.chat.historyReads, ['shared/${chat.threadId}']);
        expect(texts(chat), ['user: hi', 'assistant: hello']);
        expect(chat.notice, isNull);
        expect(r.chat.posts, hasLength(1));
      },
    );

    for (final (status, body, text) in [
      (409, {'code': 'busy'}, 'busy'),
      (400, {'code': 'thread_unbound'}, 'new thread'),
      (403, {'error': 'x'}, 'agent'),
      (413, {'error': 'x'}, 'too large'),
    ]) {
      test(
        'a $status answer keeps the turn in the composer and sends nothing again',
        () async {
          final (r, chat) = await draft();
          r.chat.script = (post, res) async => res.fail(status, body);
          expect(await chat.send('hi'), isFalse);

          expect(chat.sendError, contains(text));
          expect(
            chat.messages,
            isEmpty,
            reason: 'the optimistic turn is withdrawn',
          );
          expect(chat.phase, ChatPhase.idle);
          expect(r.chat.posts, hasLength(1));
          await Future<void>.delayed(const Duration(milliseconds: 100));
          expect(r.chat.posts, hasLength(1));
        },
      );
    }

    test('an expired session signs the person out', () async {
      final (r, chat) = await draft();
      r.chat.script = (post, res) async => res.fail(401);
      await chat.send('hi');
      await until(() => r.client.stage == ClientStage.login);
      expect(r.client.chat, isNull);
      expect(r.chat.posts, hasLength(1));
    });

    test(
      'a failed run is shown with its reason; a stop is not a failure',
      () async {
        final (r, chat) = await draft();
        r.chat.script = (post, res) async {
          await res.event(runStarted(post.threadId, post.runId));
          await res.event({
            'type': 'RUN_ERROR',
            'message': 'model unavailable',
          });
          await res.finish();
        };
        await chat.send('hi');
        await idle(chat);
        expect(chat.notice?.kind, NoticeKind.failed);
        expect(chat.notice?.text, contains('model unavailable'));

        r.chat.script = (post, res) async {
          await res.event(runStarted(post.threadId, post.runId));
          await res.event({
            'type': 'RUN_ERROR',
            'code': 'stopped',
            'message': 'stopped',
          });
          await res.finish();
        };
        await chat.send('again');
        await idle(chat);
        expect(chat.notice?.kind, NoticeKind.stopped);
      },
    );

    test(
      'a stream the server could not finish falls back to one read',
      () async {
        final (r, chat) = await draft();
        r.chat.script = (post, res) async {
          await res.event(runStarted(post.threadId, post.runId));
          await res.event({
            'type': 'CUSTOM',
            'name': 'butter.fallback',
            'value': {'reason': 'log_truncated'},
          });
          await res.finish();
        };
        await chat.send('hi');
        await idle(chat);

        expect(chat.notice?.kind, NoticeKind.fallback);
        expect(r.chat.posts, hasLength(1));
        expect(r.chat.historyReads, isEmpty);
      },
    );

    test(
      'a run that waits for input says so and does not claim forms',
      () async {
        final (r, chat) = await draft();
        r.chat.script = (post, res) async {
          await res.event(runStarted(post.threadId, post.runId));
          await res.event({
            'type': 'RUN_FINISHED',
            'threadId': post.threadId,
            'runId': post.runId,
            'outcome': {
              'type': 'interrupt',
              'interrupts': [
                {
                  'id': 'i1',
                  'reason': 'input_required',
                  'message': 'Which one?',
                },
              ],
            },
          });
          await res.finish();
        };
        await chat.send('hi');
        await idle(chat);

        expect(chat.notice?.kind, NoticeKind.waitingInput);
        expect(chat.interrupts.single.message, 'Which one?');
      },
    );

    test('a silent stream times out; heartbeats keep it alive', () async {
      final (r, chat) = await draft(
        idleTimeout: const Duration(milliseconds: 250),
      );
      final never = Completer<void>();
      r.chat.script = (post, res) async {
        await res.event(runStarted(post.threadId, post.runId));
        await res.hold(never.future);
      };
      await chat.send('hi');
      await idle(chat);
      expect(chat.notice?.kind, NoticeKind.connectionLost);
      never.complete();

      r.chat.script = (post, res) async {
        await res.event(runStarted(post.threadId, post.runId));
        for (var i = 0; i < 8; i++) {
          await Future<void>.delayed(const Duration(milliseconds: 80));
          await res.heartbeat();
        }
        await res.event(runFinished(post.threadId, post.runId));
        await res.finish();
      };
      await chat.send('again');
      await idle(chat);
      expect(chat.notice, isNull);
    });
  });

  group('opening a thread', () {
    test(
      'restores messages, tool results and images from the history',
      () async {
        final r = await rig();
        r.server.agentReply = (ws, _) async =>
            ListAgentsResponse(agents: [agent('agent-1', 'Helper')]);
        r.chat.sessions.add(session('t1', title: 'Budget'));
        r.chat.histories['agent-1/t1'] = jsonEncode({
          'threadId': 't1',
          'messages': [
            {
              'id': 'u1',
              'role': 'user',
              'content': [
                {'type': 'text', 'text': 'look'},
                {
                  'type': 'image',
                  'source': {
                    'type': 'data',
                    'value': base64Encode(png),
                    'mimeType': 'image/png',
                  },
                },
              ],
            },
            {
              'id': 'a1',
              'role': 'assistant',
              'content': 'ok',
              'toolCalls': [
                {
                  'id': 'c1',
                  'type': 'function',
                  'function': {'name': 'search', 'arguments': '{}'},
                },
              ],
            },
            {
              'id': 't1',
              'role': 'tool',
              'toolCallId': 'c1',
              'content': 'found',
            },
          ],
        });
        final chat = r.client.chat!;
        await chat.loadThreads();
        await chat.openThread(chat.threads.single);

        expect(chat.threadId, 't1');
        expect(chat.agent?.agentId, 'agent-1');
        expect(chat.opening, isFalse);
        expect(chat.messages, hasLength(2));
        expect(chat.messages[0].parts.whereType<ImagePart>(), hasLength(1));
        final tool = chat.messages[1].parts.whereType<ToolPart>().single;
        expect(tool.result, 'found');
        expect(tool.status, ToolStatus.done);
        expect(r.chat.historyReads, ['agent-1/t1']);
        expect(r.chat.posts, isEmpty, reason: 'reading history starts no run');
      },
    );

    test(
      'a thread another run holds is shown as running and cannot be sent to',
      () async {
        final r = await rig();
        r.chat.sessions.add(session('t1'));
        r.chat.histories['agent-1/t1'] = jsonEncode({
          'threadId': 't1',
          'messages': [
            {'id': 'u1', 'role': 'user', 'content': 'hi'},
          ],
          'running': {'runId': 'r1', 'invocationId': 'i1'},
        });
        final chat = r.client.chat!;
        await chat.loadThreads();
        await chat.openThread(chat.threads.single);

        expect(chat.notice?.kind, NoticeKind.running);
        expect(chat.canSend, isFalse);
        expect(r.chat.posts, isEmpty);
      },
    );

    test('a failed read is shown and not retried on its own', () async {
      final r = await rig();
      r.chat.sessions.add(session('t1'));
      final chat = r.client.chat!;
      await chat.loadThreads();
      r.chat.histories['agent-1/t1'] = 'not json';
      await chat.openThread(chat.threads.single);

      expect(chat.openError, isNotNull);
      expect(chat.messages, isEmpty);
      await Future<void>.delayed(const Duration(milliseconds: 100));
      expect(r.chat.historyReads, hasLength(1));
    });
  });

  group('context changes', () {
    test(
      'switching Workspace mid-stream detaches and drops the thread',
      () async {
        final (r, chat) = await draft();
        final never = Completer<void>();
        r.chat.script = (post, res) async {
          await res.event(runStarted(post.threadId, post.runId));
          await res.event(textDelta('m1', 'x'));
          await res.hold(never.future);
        };
        await chat.send('hi');
        await until(() => chat.phase == ChatPhase.streaming);

        await r.client.selectWorkspace('beta');
        await r.chat.posts.single.disconnected.future.timeout(
          const Duration(seconds: 5),
        );
        expect(chat.disposed, isTrue);
        expect(chat.messages, isEmpty);
        expect(r.client.chat, isNot(same(chat)));
        expect(r.client.chat!.threadId, isNull);
        expect(r.chat.stops, isEmpty);
        never.complete();
      },
    );

    test(
      'going to the background detaches the stream; coming back only reads',
      () async {
        final (r, chat) = await draft();
        final never = Completer<void>();
        r.chat.script = (post, res) async {
          await res.event(runStarted(post.threadId, post.runId));
          await res.event(textStart('m1'));
          await res.event(textDelta('m1', 'x'));
          await res.hold(never.future);
        };
        await chat.send('hi');
        await until(() => chat.phase == ChatPhase.streaming);

        r.client.suspend();
        await r.chat.posts.single.disconnected.future.timeout(
          const Duration(seconds: 5),
        );
        await idle(chat);
        expect(chat.notice?.kind, NoticeKind.connectionLost);

        r.chat.histories['shared/${chat.threadId}'] = jsonEncode({
          'threadId': chat.threadId,
          'messages': [
            {'id': 'u1', 'role': 'user', 'content': 'hi'},
            {'id': 'a1', 'role': 'assistant', 'content': 'done'},
          ],
        });
        await r.client.revalidate();
        await until(() => texts(chat).last == 'assistant: done');
        expect(texts(chat), ['user: hi', 'assistant: done']);
        expect(r.chat.posts, hasLength(1));
        never.complete();
      },
    );
  });
}
