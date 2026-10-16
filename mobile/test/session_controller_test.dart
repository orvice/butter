import 'dart:async';
import 'dart:io';

import 'package:butter_mobile/app/client_context.dart';
import 'package:butter_mobile/features/sessions/session_key.dart';
import 'package:flutter_test/flutter_test.dart';

import 'support/memory_storage.dart';
import 'support/session_server.dart';

Future<({ClientContext client, SessionServerFake server})> start() async {
  final old = HttpOverrides.current;
  HttpOverrides.global = null;
  final server = SessionServerFake();
  server.seedSession(
    workspace: 'alpha',
    app: 'web-chat',
    user: 'alice',
    id: 's1',
    title: 'First',
  );
  server.seedSession(
    workspace: 'alpha',
    app: 'cron:nightly',
    user: 'cron:nightly',
    id: 's2',
  );
  await server.start();
  final client = ClientContext(storage: MemoryStorage(), allowHttp: true);
  addTearDown(() async {
    client.dispose();
    await server.close();
    HttpOverrides.global = old;
  });
  await client.setServer(server.url);
  await client.login('alice', 'password');
  await client.selectWorkspace('alpha');
  return (client: client, server: server);
}

Future<void> settle() => Future<void>.delayed(const Duration(milliseconds: 50));

void main() {
  test('lists, paginates and opens detail', () async {
    final h = await start();
    for (var i = 0; i < 25; i++) {
      h.server.seedSession(
        workspace: 'alpha',
        app: 'web-chat',
        user: 'alice',
        id: 'p$i',
      );
    }
    final c = h.client.sessions!;
    await c.loadPage();
    expect(c.sessions.length, 20);
    expect(c.total, 27);
    await c.nextPage();
    expect(c.sessions.length, 7);
    final key = SessionKey(
      appName: 'web-chat',
      userId: 'alice',
      sessionId: 's1',
    );
    await c.openSession(key);
    expect(c.detail?.session.sessionId, 's1');
  });

  test('rename and delete never auto-retry after a lost reply', () async {
    final h = await start();
    final c = h.client.sessions!;
    final key = SessionKey(
      appName: 'web-chat',
      userId: 'alice',
      sessionId: 's1',
    );
    h.server.lostSessionReplies.add('UpdateSessionTitle');
    expect(await c.rename(key, 'New title'), isFalse);
    expect(c.renameBlocked, isTrue);
    expect(h.server.sessionWriteCount('UpdateSessionTitle'), 1);
    expect(await c.rename(key, 'Again'), isFalse);
    expect(h.server.sessionWriteCount('UpdateSessionTitle'), 1);
    h.server.lostSessionReplies.clear();
    await c.openSession(key);
    expect(await c.rename(key, 'New title'), isTrue);
  });

  test('reply uses agent id and workspace header', () async {
    final h = await start();
    final c = h.client.sessions!;
    final key = SessionKey(
      appName: 'cron:nightly',
      userId: 'cron:nightly',
      sessionId: 's2',
    );
    final result = await c.reply(
      key: key,
      agentId: 'shared',
      message: 'resume please',
    );
    expect(result.response, isNotNull);
    expect(h.server.sessionCalls.last.method, 'ReplySession');
    expect(h.server.sessionCalls.last.workspace, 'alpha');
  });

  test(
    'a read that started before a write cannot clear unconfirmed rename',
    () async {
      final h = await start();
      final c = h.client.sessions!;
      final key = SessionKey(
        appName: 'web-chat',
        userId: 'alice',
        sessionId: 's1',
      );
      final slow = Completer<void>();
      h.server.sessionBefore = (method) async {
        if (method == 'GetSession') await slow.future;
      };
      final read = c.openSession(key);
      await settle();
      h.server.lostSessionReplies.add('UpdateSessionTitle');
      unawaited(c.rename(key, 'Lost'));
      await settle();
      slow.complete();
      await read;
      expect(c.renameBlocked, isTrue);
    },
  );
}
