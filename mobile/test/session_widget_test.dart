import 'package:butter_mobile/features/sessions/session_controller.dart';
import 'package:butter_mobile/features/sessions/session_key.dart';
import 'package:butter_mobile/features/sessions/session_pages.dart';
import 'package:butter_mobile/gen/agents/v1/agent_service.pb.dart';
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';

import 'support/config_rig.dart';
import 'support/session_server.dart';

void main() {
  testWidgets('session list and detail show title and trace id', (
    tester,
  ) async {
    final server = SessionServerFake();
    server.seedSession(
      workspace: 'alpha',
      app: 'web-chat',
      user: 'alice',
      id: 's1',
      title: 'Hello',
      events: [
        SessionEvent(
          eventId: 'e1',
          author: 'user',
          contentJson: '{"parts":[{"text":"hi"}]}',
          traceId: 'trace-abc123456789',
        ),
      ],
    );
    final h = await signedIn(tester, using: server);
    final ctl = h.client.sessions!;
    final key = SessionKey(
      appName: 'web-chat',
      userId: 'alice',
      sessionId: 's1',
    );
    await tester.runAsync(
      () =>
          tester.pumpWidget(MaterialApp(home: SessionsPage(client: h.client))),
    );
    await waitFor(tester, () => ctl.listState == SessionLoad.ready);
    expect(find.text('Hello'), findsOneWidget);
    await tester.runAsync(() => ctl.openSession(key));
    await waitFor(
      tester,
      () => ctl.detailState == SessionLoad.ready && ctl.detail != null,
    );
    await tester.runAsync(
      () => tester.pumpWidget(
        MaterialApp(
          home: SessionDetailPage(client: h.client, sessionKey: key),
        ),
      ),
    );
    await tester.pump(const Duration(milliseconds: 100));
    expect(find.textContaining('trace-abc'), findsOneWidget);
    expect(find.text('hi'), findsOneWidget);
  });
}
