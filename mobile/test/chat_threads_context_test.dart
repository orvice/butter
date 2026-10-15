import 'dart:async';

import 'package:butter_mobile/app/client_context.dart';
import 'package:butter_mobile/features/chat/chat_controller.dart';
import 'package:butter_mobile/features/chat/threads.dart';
import 'package:butter_mobile/gen/agents/v1/agent_service.pb.dart';
import 'package:flutter_test/flutter_test.dart';

import 'support/chat_fixtures.dart';
import 'support/chat_rig.dart';

List<String> ids(Iterable<ThreadRow> rows) => [
  for (final r in rows) r.threadId,
];

void main() {
  group('thread list', () {
    test('shows this Workspace\'s bound threads with their agents', () async {
      final r = await rig();
      r.server.agentReply = (ws, _) async => ListAgentsResponse(
        agents: [
          agent('agent-1', 'Helper'),
          agent('agent-2', 'Writer', displayName: 'Scribe'),
        ],
      );
      r.chat.sessions.addAll([
        session('t1', title: 'Budget'),
        session('t2', workspace: 'beta'),
        session('t3', app: 'cron:nightly'),
        unbound('t4'),
        session('t5', agent: 'agent-2'),
      ]);
      final chat = r.client.chat!;
      await chat.loadThreads();

      expect(chat.threadsState, ThreadsState.ready);
      expect(ids(chat.threads), ['t1', 't5']);
      expect(chat.threads.map((t) => t.title), ['Budget', 'Scribe']);
      final call = r.chat.listCalls.single;
      expect(call.appName, 'agui');
      expect(call.userId, 'alice');
      expect(call.workspaceScoped, isTrue);
      expect(r.chat.listWorkspaces.single, 'alpha');
    });

    test('reads one page at a time and the next only when asked', () async {
      final r = await rig(pageSize: 2);
      r.chat.sessions.addAll([for (var i = 1; i <= 5; i++) session('t$i')]);
      final chat = r.client.chat!;
      await chat.loadThreads();
      expect(ids(chat.threads), ['t1', 't2']);
      expect(chat.hasMoreThreads, isTrue);
      expect(r.chat.listCalls, hasLength(1));

      await chat.moreThreads();
      expect(ids(chat.threads), ['t1', 't2', 't3', 't4']);
      await chat.moreThreads();
      expect(ids(chat.threads), ['t1', 't2', 't3', 't4', 't5']);
      expect(chat.hasMoreThreads, isFalse);
      await chat.moreThreads();
      expect(r.chat.listCalls, hasLength(3), reason: 'nothing left to read');
      expect(r.chat.listCalls[1].pageToken, '2');
    });

    test(
      'a page of only other sessions does not hide the threads after it',
      () async {
        final r = await rig(pageSize: 2);
        r.chat.sessions.addAll([
          unbound('x1'),
          session('x2', app: 'cron:job'),
          unbound('x3'),
          session('t1'),
        ]);
        final chat = r.client.chat!;
        await chat.loadThreads();
        expect(ids(chat.threads), ['t1']);
      },
    );

    test('search reads the remaining pages so a match is not missed', () async {
      final r = await rig(pageSize: 2);
      r.chat.sessions.addAll([
        session('t1', title: 'Alpha'),
        session('t2', title: 'Beta'),
        session('t3', title: 'Gamma'),
        session('t4', title: 'Quarterly plan'),
        session('t5', title: 'Delta'),
      ]);
      final chat = r.client.chat!;
      await chat.loadThreads();
      expect(r.chat.listCalls, hasLength(1));

      await chat.search(' PLAN ');
      expect(ids(chat.visibleThreads), ['t4']);
      expect(chat.hasMoreThreads, isFalse);
      expect(r.chat.listCalls, hasLength(3));

      await chat.search('');
      expect(chat.visibleThreads, hasLength(5));
    });

    test('a failed read is shown once and never repeated by itself', () async {
      final r = await rig();
      r.chat.listFailure = (403, 'permission_denied');
      final chat = r.client.chat!;
      await chat.loadThreads();
      expect(chat.threadsState, ThreadsState.failed);
      expect(chat.threadsError, isNotNull);
      expect(chat.threads, isEmpty);
      await Future<void>.delayed(const Duration(milliseconds: 50));
      expect(r.chat.listCalls, hasLength(1));
      // An explicit reload reads again and recovers.
      r.chat.listFailure = null;
      r.chat.sessions.add(session('t1'));
      await chat.loadThreads();
      expect(chat.threadsState, ThreadsState.ready);
      expect(ids(chat.threads), ['t1']);
    });

    test(
      'an expired session signs the person out instead of retrying',
      () async {
        final r = await rig();
        r.chat.listFailure = (401, 'unauthenticated');
        await r.client.chat!.loadThreads();
        expect(r.client.stage, ClientStage.login);
        expect(r.client.chat, isNull);
        expect(r.chat.listCalls, hasLength(1));
      },
    );
  });

  group('rename', () {
    test(
      'sends the normalized title for that session and shows the answer',
      () async {
        final r = await rig();
        r.chat.sessions.add(session('t1', title: 'Old'));
        final chat = r.client.chat!;
        await chat.loadThreads();

        expect(
          await chat.renameThread(chat.threads.single, '  New \n name '),
          isTrue,
        );
        final sent = r.chat.renames.single;
        expect(sent.appName, 'agui');
        expect(sent.userId, 'alice');
        expect(sent.sessionId, 'agui-t1');
        expect(sent.title, 'New name');
        expect(chat.threads.single.title, 'New name');
        expect(chat.renameError, isNull);
      },
    );

    test('a title the server would refuse is never sent', () async {
      final r = await rig();
      r.chat.sessions.add(session('t1', title: 'Old'));
      final chat = r.client.chat!;
      await chat.loadThreads();
      expect(await chat.renameThread(chat.threads.single, '   '), isFalse);
      expect(await chat.renameThread(chat.threads.single, 'x' * 101), isFalse);
      expect(r.chat.renames, isEmpty);
      expect(chat.renameError, isNotNull);
      expect(chat.threads.single.title, 'Old');
    });

    test('a refused rename keeps the old title and is not retried', () async {
      final r = await rig();
      r.chat.sessions.add(session('t1', title: 'Old'));
      r.chat.renameFailure = (503, 'unavailable');
      final chat = r.client.chat!;
      await chat.loadThreads();
      expect(await chat.renameThread(chat.threads.single, 'New'), isFalse);
      await Future<void>.delayed(const Duration(milliseconds: 50));
      expect(r.chat.renames, hasLength(1));
      expect(chat.threads.single.title, 'Old');
      expect(chat.renameError, isNotNull);
    });
  });

  group('context isolation', () {
    test(
      'another Workspace gets its own list and a late answer is dropped',
      () async {
        final r = await rig();
        r.chat.sessions.addAll([
          session('a1'),
          session('b1', workspace: 'beta'),
        ]);
        final held = Completer<void>(), reached = Completer<void>();
        r.chat.beforeList = () async {
          if (!reached.isCompleted) {
            reached.complete();
            await held.future;
          }
        };
        final first = r.client.chat!;
        final loading = first.loadThreads();
        await reached.future;

        await r.client.selectWorkspace('beta');
        held.complete();
        await loading;

        final second = r.client.chat!;
        expect(second, isNot(same(first)));
        expect(first.threads, isEmpty, reason: 'the old scope is closed');
        expect(
          second.threads,
          isEmpty,
          reason: 'nothing leaked into the new one',
        );
        await second.loadThreads();
        expect(ids(second.threads), ['b1']);
        expect(r.chat.listWorkspaces.last, 'beta');
      },
    );

    test('signing out drops the chat and its threads', () async {
      final r = await rig();
      r.chat.sessions.add(session('t1'));
      final chat = r.client.chat!;
      await chat.loadThreads();
      await r.client.logout();
      expect(r.client.chat, isNull);
      expect(chat.threads, isEmpty);
    });
  });
}
