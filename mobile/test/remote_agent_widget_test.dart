import 'package:butter_mobile/features/remote_agents/remote_agent_controller.dart';
import 'package:butter_mobile/features/remote_agents/remote_agent_pages.dart';
import 'package:butter_mobile/gen/agents/v1/agent.pb.dart';
import 'package:butter_mobile/gen/agents/v1/dashboard.pb.dart';
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';

import 'support/config_rig.dart';

RemoteAgent a2a(String id) => RemoteAgent(
  id: id,
  name: 'Agent $id',
  url: 'https://a2a.example.com',
  protocol: RemoteAgentProtocol.REMOTE_AGENT_PROTOCOL_A2A,
);

RemoteAgent opencode(String id) => RemoteAgent(
  id: id,
  name: 'Code $id',
  url: 'https://oc.example.com',
  protocol: RemoteAgentProtocol.REMOTE_AGENT_PROTOCOL_OPENCODE_HTTP,
  username: 'oc-user',
  password: 'oc-very-secret',
);

RemoteAgent daemon(String id) => RemoteAgent(
  id: id,
  name: 'Daemon $id',
  protocol: RemoteAgentProtocol.REMOTE_AGENT_PROTOCOL_DAEMON,
  daemonRuntimeId: 'rt-1',
  acpRuntime: 'opencode',
);

Future<void> openList(WidgetTester tester, ConfigRig h) async {
  await openPage(tester, RemoteAgentListPage(client: h.client));
  await tester.runAsync(() => h.client.remoteAgents!.load());
  await tester.pumpAndSettle();
}

Future<void> openDetail(WidgetTester tester, ConfigRig h, String id) async {
  await tester.runAsync(
    () => tester.pumpWidget(
      MaterialApp(
        home: RemoteAgentDetailPage(client: h.client, id: id),
      ),
    ),
  );
  await waitFor(tester, () => h.client.remoteAgents!.item?.id == id);
}

void main() {
  testWidgets('lists this Workspace only; Daemon agents are read-only', (
    tester,
  ) async {
    final h = await signedIn(
      tester,
      seed: (s) {
        s.seedRemote('alpha', a2a('a1'));
        s.seedRemote('alpha', daemon('d1'));
        s.seedRemote('beta', a2a('b1'));
      },
    );
    await openList(tester, h);
    expect(find.text('Agent a1'), findsOneWidget);
    expect(find.text('Daemon d1'), findsOneWidget);
    expect(find.text('Agent b1'), findsNothing);

    await tester.runAsync(
      () => tester.pumpWidget(
        MaterialApp(
          home: RemoteAgentDetailPage(client: h.client, id: 'd1'),
        ),
      ),
    );
    await waitFor(tester, () => h.client.remoteAgents!.item?.id == 'd1');
    expect(find.byKey(const Key('remote-daemon-note')), findsOneWidget);
    expect(find.byKey(const Key('edit-remote')), findsNothing);
    expect(find.byKey(const Key('delete-remote')), findsNothing);
    expect(find.byKey(const Key('probe-remote')), findsOneWidget);
  });

  testWidgets('adding an OpenCode agent stores the password and never shows '
      'it', (tester) async {
    final h = await signedIn(tester);
    await openList(tester, h);
    await tapScreen(tester, find.byKey(const Key('add-remote')));
    await enter(tester, const Key('remote-id'), 'oc');
    await enter(tester, const Key('remote-name'), 'Code');
    await tapScreen(tester, find.text('OpenCode HTTP'));
    await enter(tester, const Key('remote-url'), 'https://oc.example.com');
    await enter(tester, const Key('remote-username'), 'me');
    await tapScreen(tester, find.text('Replace'));
    await enter(tester, const Key('secret-value-Password'), 'typed-secret');
    final field = tester.widget<TextField>(
      find.descendant(
        of: find.byKey(const Key('secret-value-Password')),
        matching: find.byType(TextField),
      ),
    );
    expect(field.obscureText, isTrue);
    await tapReal(tester, find.byKey(const Key('save-remote')));
    await waitFor(tester, () => h.server.writes == 1);
    await tester.pumpAndSettle();

    final stored = h.server.remotes['alpha']!['oc']!;
    expect(stored.password, 'typed-secret');
    expect(stored.username, 'me');
    expect(
      stored.protocol,
      RemoteAgentProtocol.REMOTE_AGENT_PROTOCOL_OPENCODE_HTTP,
    );
    expect(find.textContaining('typed-secret'), findsNothing);
  });

  testWidgets('a username without a password is refused before any request', (
    tester,
  ) async {
    final h = await signedIn(tester);
    await openList(tester, h);
    await tapScreen(tester, find.byKey(const Key('add-remote')));
    await enter(tester, const Key('remote-id'), 'oc');
    await enter(tester, const Key('remote-name'), 'Code');
    await tapScreen(tester, find.text('OpenCode HTTP'));
    await enter(tester, const Key('remote-url'), 'https://oc.example.com');
    await enter(tester, const Key('remote-username'), 'me');
    await tapScreen(tester, find.byKey(const Key('save-remote')));
    expect(find.textContaining('needs a password'), findsWidgets);
    expect(h.server.writes, 0);
  });

  testWidgets('editing keeps the stored password hidden and untouched', (
    tester,
  ) async {
    final h = await signedIn(
      tester,
      seed: (s) => s.seedRemote('alpha', opencode('o1')),
    );
    await openDetail(tester, h, 'o1');
    expect(find.textContaining('password set (hidden)'), findsOneWidget);
    expect(find.textContaining('oc-very-secret'), findsNothing);
    await tapScreen(tester, find.byKey(const Key('edit-remote')));
    await enter(tester, const Key('remote-name'), 'Renamed');
    await tapReal(tester, find.byKey(const Key('save-remote')));
    await waitFor(tester, () => h.server.writes == 1);
    final stored = h.server.remotes['alpha']!['o1']!;
    expect(stored.name, 'Renamed');
    expect(stored.password, 'oc-very-secret');
  });

  testWidgets('the probe is classified and hides the server text', (
    tester,
  ) async {
    final h = await signedIn(
      tester,
      seed: (s) => s.seedRemote('alpha', a2a('a1')),
    );
    await openDetail(tester, h, 'a1');
    expect(find.text('Not probed'), findsOneWidget);
    h.server.remoteState = RemoteAgentStatus_State.STATE_UNREACHABLE;
    h.server.remoteDetail = 'Get "https://u:pw-leak@host": dial tcp: refused';
    await tapReal(tester, find.byKey(const Key('probe-remote')));
    await waitFor(
      tester,
      () =>
          h.client.remoteAgents!.probes['a1']?.kind ==
          RemoteProbeKind.unreachable,
    );
    expect(find.textContaining('could not reach the endpoint'), findsOneWidget);
    expect(find.textContaining('pw-leak'), findsNothing);
    expect(h.server.writes, 0);
  });

  testWidgets('delete asks first and a cancelled dialog sends nothing', (
    tester,
  ) async {
    final h = await signedIn(
      tester,
      seed: (s) => s.seedRemote('alpha', a2a('a1')),
    );
    await openDetail(tester, h, 'a1');
    await tapReal(tester, find.byKey(const Key('delete-remote')));
    await tester.pumpAndSettle();
    expect(find.text('Delete Remote Agent?'), findsOneWidget);
    await tapDialog(tester, find.text('Cancel'));
    expect(h.server.count('DeleteRemoteAgent'), 0);
    await tapReal(tester, find.byKey(const Key('delete-remote')));
    await tester.pumpAndSettle();
    await tapDialog(tester, find.text('Delete'));
    await waitFor(tester, () => h.server.count('DeleteRemoteAgent') == 1);
    expect(h.server.remotes['alpha'], isEmpty);
  });
}
