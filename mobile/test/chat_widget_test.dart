import 'dart:async';
import 'dart:convert';
import 'dart:io';

import 'package:butter_mobile/app/butter_app.dart';
import 'package:butter_mobile/app/client_context.dart';
import 'package:butter_mobile/gen/agents/v1/agent_service.pb.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:image_picker_platform_interface/image_picker_platform_interface.dart';

import 'support/chat_fixtures.dart';
import 'support/chat_server.dart';
import 'support/memory_storage.dart';
import 'support/server.dart';

final png = base64Decode(
  'iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNkYPhfDwAChwGA60e6kgAAAABJRU5ErkJggg==',
);

class GalleryPicker extends ImagePickerPlatform {
  GalleryPicker(this.files);
  List<XFile> files;
  int calls = 0;
  @override
  Future<List<XFile>> getMultiImageWithOptions({
    MultiImagePickerOptions options = const MultiImagePickerOptions(),
  }) async {
    calls++;
    return files;
  }
}

typedef Chat = ({ClientContext client, TestServer server, ChatServer chat});

/// The whole app, signed in with a Workspace chosen, on the Agents screen.
Future<Chat> openApp(
  WidgetTester tester, {
  void Function(ChatServer)? seed,
}) async {
  final oldHttp = HttpOverrides.current;
  final oldPicker = ImagePickerPlatform.instance;
  HttpOverrides.global = null;
  final server = TestServer();
  server.agentReply = (ws, _) async =>
      ListAgentsResponse(agents: [agent('shared', 'shared agent')]);
  final chat = ChatServer(server);
  seed?.call(chat);
  final client = ClientContext(storage: MemoryStorage(), allowHttp: true);
  addTearDown(() async {
    client.dispose();
    await server.close();
    HttpOverrides.global = oldHttp;
    ImagePickerPlatform.instance = oldPicker;
  });
  await tester.pumpWidget(ButterApp(context: client));
  await tester.runAsync(server.start);
  await tester.runAsync(() => client.setServer(server.url));
  await tester.runAsync(() => client.login('alice', 'password'));
  await tester.runAsync(() => client.selectWorkspace('alpha'));
  await tester.pumpAndSettle();
  return (client: client, server: server, chat: chat);
}

/// Lets real I/O (sockets) and then frames run until [done] holds.
Future<void> settleUntil(WidgetTester tester, bool Function() done) async {
  for (var i = 0; i < 300 && !done(); i++) {
    await tester.runAsync(
      () => Future<void>.delayed(const Duration(milliseconds: 10)),
    );
    await tester.pump();
  }
  expect(done(), isTrue, reason: 'never settled');
  await tester.pump();
}

Future<void> openChatList(WidgetTester tester, Chat c) async {
  await tester.tap(find.byKey(const Key('open-chat')));
  await tester.pump();
  await settleUntil(
    tester,
    () => c.client.chat!.threadsState.name != 'loading',
  );
  await tester.pumpAndSettle();
}

/// Frames only (never `pumpAndSettle`) while real I/O is pending: a spinning
/// progress indicator would let fake time race ahead of the sockets.
Future<void> frames(WidgetTester tester) async {
  for (var i = 0; i < 5; i++) {
    await tester.pump(const Duration(milliseconds: 100));
  }
}

Future<void> startNewChat(WidgetTester tester) async {
  await tester.tap(find.byKey(const Key('new-chat')));
  await tester.pumpAndSettle();
  await tester.tap(find.text('shared agent').last);
  await tester.pumpAndSettle();
}

Future<void> typeAndSend(WidgetTester tester, String text) async {
  await tester.enterText(find.byKey(const Key('composer')), text);
  await tester.pump();
  await tester.tap(find.byKey(const Key('send')));
  await tester.pump();
}

Future<void> reply(WidgetTester tester, Chat c, String markdown) async {
  c.chat.script = (post, res) async {
    await res.event(runStarted(post.threadId, post.runId));
    await res.event(textStart('m1'));
    await res.event(textDelta('m1', markdown));
    await res.event(textEnd('m1'));
    await res.event(runFinished(post.threadId, post.runId));
    await res.finish();
  };
}

/// A widget test whose requests start inside the test's fake-time zone: it
/// ends by letting their (already moot) deadline timers run out.
void chatTest(String name, Future<void> Function(WidgetTester) body) {
  testWidgets(name, (tester) async {
    await body(tester);
    await tester.pump(const Duration(minutes: 2));
  });
}

bool idle(Chat c) => c.client.chat?.phase.name == 'idle';

void main() {
  chatTest(
    'threads are listed with their agent, search filters, other Workspaces stay out',
    (tester) async {
      final c = await openApp(
        tester,
        seed: (chat) => chat.sessions.addAll([
          session('t1', title: 'Budget plan', agent: 'shared'),
          session('t2', title: 'Holiday ideas', agent: 'shared'),
          session('t3', title: 'Beta secret', workspace: 'beta'),
          unbound('t4'),
        ]),
      );
      c.server.agentReply = (ws, _) async =>
          ListAgentsResponse(agents: [agent('shared', 'shared agent')]);
      await tester.runAsync(c.client.refreshAgents);
      await openChatList(tester, c);

      expect(find.text('Budget plan'), findsOneWidget);
      expect(find.text('Holiday ideas'), findsOneWidget);
      expect(find.text('Beta secret'), findsNothing);
      expect(find.text('shared agent'), findsWidgets);

      await tester.enterText(find.byKey(const Key('thread-search')), 'holi');
      await tester.pumpAndSettle();
      expect(find.text('Budget plan'), findsNothing);
      expect(find.text('Holiday ideas'), findsOneWidget);
    },
  );

  chatTest(
    'a thread can be renamed; a blank title is refused before any request',
    (tester) async {
      final c = await openApp(
        tester,
        seed: (chat) => chat.sessions.add(session('t1', title: 'Old name')),
      );
      c.server.agentReply = (ws, _) async =>
          ListAgentsResponse(agents: [agent('agent-1', 'Helper')]);
      await openChatList(tester, c);

      await tester.tap(find.byTooltip('Thread options'));
      await tester.pumpAndSettle();
      await tester.tap(find.text('Rename'));
      await tester.pumpAndSettle();
      await tester.enterText(find.byKey(const Key('rename-field')), '   ');
      await tester.tap(find.text('Save'));
      await tester.pumpAndSettle();
      expect(c.chat.renames, isEmpty);
      expect(find.byKey(const Key('rename-field')), findsOneWidget);

      await tester.enterText(find.byKey(const Key('rename-field')), 'New name');
      await tester.tap(find.text('Save'));
      await settleUntil(
        tester,
        () =>
            find.text('New name').evaluate().length > 1 ||
            c.chat.renames.isNotEmpty &&
                find.text('Old name').evaluate().isEmpty,
      );
      await tester.pumpAndSettle();
      expect(c.chat.renames.single.title, 'New name');
      expect(find.text('New name'), findsOneWidget);
      expect(find.text('Old name'), findsNothing);
    },
  );

  chatTest('a new chat sends and renders Markdown: table, code with copy', (
    tester,
  ) async {
    final c = await openApp(tester);
    await reply(
      tester,
      c,
      '# Plan\n\n| a | b |\n|---|---|\n| 1 | 2 |\n\n```dart\nprint(1);\n```\n',
    );
    String? copied;
    tester.binding.defaultBinaryMessenger.setMockMethodCallHandler(
      SystemChannels.platform,
      (call) async {
        if (call.method == 'Clipboard.setData') {
          copied = (call.arguments as Map)['text'] as String?;
        }
        return null;
      },
    );
    addTearDown(
      () => tester.binding.defaultBinaryMessenger.setMockMethodCallHandler(
        SystemChannels.platform,
        null,
      ),
    );
    await openChatList(tester, c);
    await startNewChat(tester);
    await typeAndSend(tester, 'show me');
    await settleUntil(tester, () => idle(c));

    expect(find.text('show me'), findsOneWidget);
    expect(find.text('Plan'), findsOneWidget);
    expect(find.text('a'), findsOneWidget);
    expect(find.text('2'), findsOneWidget);
    expect(find.textContaining('print(1);'), findsOneWidget);
    expect(find.text('shared agent'), findsWidgets, reason: 'agent identity');

    await tester.tap(find.byTooltip('Copy code'));
    await tester.pump();
    expect(copied, contains('print(1);'));
    // The composer was emptied by the send.
    expect(
      tester
          .widget<TextField>(find.byKey(const Key('composer')))
          .controller!
          .text,
      isEmpty,
    );
  });

  chatTest(
    'a tool call shows its status and, when opened, its arguments and result',
    (tester) async {
      final c = await openApp(tester);
      final proceed = Completer<void>();
      c.chat.script = (post, res) async {
        await res.event(runStarted(post.threadId, post.runId));
        await res.event({
          'type': 'TOOL_CALL_START',
          'toolCallId': 'c1',
          'toolCallName': 'search_docs',
        });
        await res.event({
          'type': 'TOOL_CALL_ARGS',
          'toolCallId': 'c1',
          'delta': '{"q":"butter"}',
        });
        await res.event({'type': 'TOOL_CALL_END', 'toolCallId': 'c1'});
        await res.hold(proceed.future);
        await res.event({
          'type': 'TOOL_CALL_RESULT',
          'messageId': 'r1',
          'toolCallId': 'c1',
          'content': 'three hits',
        });
        await res.event(runFinished(post.threadId, post.runId));
        await res.finish();
      };
      await openChatList(tester, c);
      await startNewChat(tester);
      await typeAndSend(tester, 'look');
      await settleUntil(
        tester,
        () => find.text('search_docs').evaluate().isNotEmpty,
      );
      expect(find.text('Running…'), findsOneWidget);

      proceed.complete();
      await settleUntil(tester, () => idle(c));
      expect(find.text('Done'), findsOneWidget);
      await tester.tap(find.text('search_docs'));
      await tester.pumpAndSettle();
      expect(find.textContaining('"q":"butter"'), findsOneWidget);
      expect(find.text('three hits'), findsOneWidget);
    },
  );

  chatTest('images are picked, previewed, removable, sent once and shown', (
    tester,
  ) async {
    final c = await openApp(tester);
    ImagePickerPlatform.instance = GalleryPicker([
      XFile.fromData(Uint8List.fromList(png), name: 'a.png'),
      XFile.fromData(Uint8List.fromList(png), name: 'b.png'),
    ]);
    await openChatList(tester, c);
    await startNewChat(tester);

    await tester.tap(find.byTooltip('Attach image'));
    await tester.pumpAndSettle();
    expect(find.byTooltip('Remove image 1'), findsOneWidget);
    expect(find.byTooltip('Remove image 2'), findsOneWidget);
    await tester.tap(find.byTooltip('Remove image 2'));
    await tester.pumpAndSettle();
    expect(find.byTooltip('Remove image 2'), findsNothing);

    await typeAndSend(tester, 'see this');
    await settleUntil(tester, () => idle(c));
    final parts = c.chat.posts.single.content! as List;
    expect(parts.where((p) => (p as Map)['type'] == 'image'), hasLength(1));
    expect(
      find.byTooltip('Remove image 1'),
      findsNothing,
      reason: 'tray cleared',
    );
    expect(find.byType(Image), findsOneWidget);

    await typeAndSend(tester, 'and now words only');
    await settleUntil(tester, () => c.chat.posts.length == 2 && idle(c));
    expect(c.chat.posts[1].content, 'and now words only');
  });

  chatTest(
    'leaving the page detaches: no stop on leave; Stop is enabled while streaming',
    (tester) async {
      final c = await openApp(tester);
      final never = Completer<void>();
      c.chat.script = (post, res) async {
        await res.event(runStarted(post.threadId, post.runId));
        await res.event(textStart('m1'));
        await res.event(textDelta('m1', 'working'));
        await res.hold(never.future);
      };
      await openChatList(tester, c);
      await startNewChat(tester);
      await typeAndSend(tester, 'go');
      await settleUntil(
        tester,
        () => find.text('working').evaluate().isNotEmpty,
      );

      final stop = find.widgetWithIcon(IconButton, Icons.stop_circle_outlined);
      expect(stop, findsOneWidget);
      expect(tester.widget<IconButton>(stop).tooltip, 'Stop run');
      expect(tester.widget<IconButton>(stop).onPressed, isNotNull);

      await tester.pageBack();
      await frames(tester);
      await tester.runAsync(
        () => c.chat.posts.single.disconnected.future.timeout(
          const Duration(seconds: 5),
        ),
      );
      await tester.runAsync(
        () => Future<void>.delayed(const Duration(milliseconds: 100)),
      );
      expect(c.chat.stops, isEmpty);
      expect(c.chat.posts, hasLength(1));
      never.complete();
    },
  );

  chatTest(
    'a dropped stream says the run may still be going; Refresh reads once',
    (tester) async {
      final c = await openApp(tester);
      c.chat.script = (post, res) async {
        await res.event(runStarted(post.threadId, post.runId));
        await res.event(textStart('m1'));
        await res.event(textDelta('m1', 'half'));
        await res.drop();
      };
      await openChatList(tester, c);
      await startNewChat(tester);
      await typeAndSend(tester, 'go');
      await settleUntil(tester, () => idle(c));

      expect(find.textContaining('may still be going'), findsOneWidget);
      expect(find.text('half'), findsOneWidget);
      expect(find.textContaining('stopped'), findsNothing);
      expect(c.chat.historyReads, isEmpty);

      await tester.tap(find.text('Refresh'));
      await settleUntil(tester, () => c.chat.historyReads.length == 1);
      await settleUntil(tester, () => !c.client.chat!.opening);
      await tester.pumpAndSettle();
      expect(c.chat.historyReads, hasLength(1));
      expect(c.chat.posts, hasLength(1));
    },
  );

  chatTest(
    'a thread without a UI binding is refused with a plain message and the text is kept',
    (tester) async {
      final c = await openApp(tester);
      c.chat.script = (post, res) async =>
          res.fail(400, {'error': 'x', 'code': 'thread_unbound'});
      await openChatList(tester, c);
      await startNewChat(tester);
      await typeAndSend(tester, 'keep me');
      await settleUntil(tester, () => idle(c));
      await tester.pumpAndSettle();

      expect(find.textContaining('Start a new thread'), findsOneWidget);
      expect(
        tester
            .widget<TextField>(find.byKey(const Key('composer')))
            .controller!
            .text,
        'keep me',
      );
      expect(c.chat.posts, hasLength(1));
    },
  );

  chatTest(
    'an opened thread shows its history, restored images and unavailable ones',
    (tester) async {
      final c = await openApp(
        tester,
        seed: (chat) {
          chat.sessions.add(session('t1', title: 'Pictures', agent: 'shared'));
          chat.histories['shared/t1'] = jsonEncode({
            'threadId': 't1',
            'messages': [
              {
                'id': 'u1',
                'role': 'user',
                'content': [
                  {'type': 'text', 'text': 'first look'},
                  {
                    'type': 'image',
                    'source': {
                      'type': 'data',
                      'value': base64Encode(png),
                      'mimeType': 'image/png',
                    },
                  },
                  {
                    'type': 'image',
                    'source': {
                      'type': 'data',
                      'value': base64Encode(utf8.encode('not an image at all')),
                      'mimeType': 'image/png',
                    },
                  },
                ],
              },
              {'id': 'a1', 'role': 'assistant', 'content': 'nice **photo**'},
            ],
          });
        },
      );
      c.server.agentReply = (ws, _) async =>
          ListAgentsResponse(agents: [agent('shared', 'shared agent')]);
      await tester.runAsync(c.client.refreshAgents);
      await openChatList(tester, c);
      await tester.tap(find.text('Pictures'));
      await settleUntil(tester, () => c.chat.historyReads.isNotEmpty);
      await settleUntil(tester, () => !c.client.chat!.opening);
      await tester.pumpAndSettle();

      expect(find.text('first look'), findsOneWidget);
      expect(find.textContaining('photo'), findsOneWidget);
      expect(find.byType(Image), findsOneWidget);
      expect(find.text('Image unavailable'), findsOneWidget);
      expect(c.chat.posts, isEmpty);
    },
  );
}
