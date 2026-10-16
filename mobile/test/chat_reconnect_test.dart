import 'dart:convert';

import 'package:butter_mobile/features/chat/chat_controller.dart';
import 'package:flutter_test/flutter_test.dart';

import 'support/chat_fixtures.dart';
import 'support/chat_rig.dart';
import 'support/chat_server.dart';

Future<void> until(bool Function() done, {String? reason}) async {
  final stop = DateTime.now().add(const Duration(seconds: 8));
  while (!done()) {
    if (DateTime.now().isAfter(stop)) fail(reason ?? 'condition never held');
    await Future<void>.delayed(const Duration(milliseconds: 10));
  }
}

Future<void> idle(ChatController chat) =>
    until(() => chat.phase == ChatPhase.idle && !chat.isPolling);

void main() {
  test('attach replay replaces stale partial text', () async {
    final r = await rig();
    r.chat.sessions.add(session('t1'));
    r.chat.histories['agent-1/t1'] = jsonEncode({
      'threadId': 't1',
      'messages': [
        {'id': 'u1', 'role': 'user', 'content': 'hi'},
      ],
      'running': {'runId': 'r1', 'invocationId': 'i1'},
    });
    r.chat.attachScript = (agent, thread, res) async {
      await res.event(runStarted(thread, 'r1'));
      await res.event(textStart('m1'));
      await res.event(textDelta('m1', 'full'));
      await res.event(runFinished(thread, 'r1'));
      await res.finish();
    };
    final chat = r.client.chat!;
    await chat.loadThreads();
    await chat.openThread(chat.threads.single);
    await idle(chat);
    expect(chat.messages.last.text, 'full');
    expect(chat.liveReply, isNull);
  });

  test('fallback polling replaces a partial live reply', () async {
    final r = await rig();
    r.chat.sessions.add(session('t1'));
    r.chat.histories['agent-1/t1'] = jsonEncode({
      'threadId': 't1',
      'messages': [
        {'id': 'u1', 'role': 'user', 'content': 'hi'},
      ],
      'running': {'runId': 'r1', 'invocationId': 'i1'},
    });
    r.chat.attachScript = (agent, thread, res) async {
      await res.event(runStarted(thread, 'r1'));
      await res.event(textStart('m1'));
      await res.event(textDelta('m1', 'part'));
      await res.event({
        'type': 'CUSTOM',
        'name': 'butter.fallback',
        'value': {'reason': 'truncated'},
      });
      await res.finish();
    };
    final chat = r.client.chat!;
    await chat.loadThreads();
    await chat.openThread(chat.threads.single);
    await until(() => chat.isPolling);
    r.chat.histories['agent-1/t1'] = jsonEncode({
      'threadId': 't1',
      'messages': [
        {'id': 'u1', 'role': 'user', 'content': 'hi'},
        {'id': 'a1', 'role': 'assistant', 'content': 'done on server'},
      ],
    });
    await idle(chat);
    expect(chat.messages.last.text, 'done on server');
  });

  test('stop 204 does not show stopped until history says so', () async {
    final r = await rig();
    final chat = r.client.chat!;
    await chat.loadThreads();
    chat.startThread(chat.agents.single);
    r.chat.stopAccepted = null;
    await chat.stopRun();
    expect(chat.stopError, isNull);
    expect(chat.notice?.kind, isNot(NoticeKind.stopped));
  });
}
