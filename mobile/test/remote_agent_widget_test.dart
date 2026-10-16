import 'package:butter_mobile/core/config_resource_controller.dart';
import 'package:butter_mobile/features/remote_agents/remote_agent_controller.dart';
import 'package:butter_mobile/features/remote_agents/remote_agent_pages.dart';
import 'package:butter_mobile/gen/agents/v1/agent.pb.dart';
import 'package:butter_mobile/gen/agents/v1/dashboard.pb.dart';
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';

import 'support/config_rig.dart';
import 'support/daemon_server.dart';

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
  testWidgets('lists this Workspace only and flags a Daemon agent whose '
      'runtime is gone', (tester) async {
    final h = await signedIn(
      tester,
      using: DaemonServerFake(),
      seed: (s) {
        s.seedRemote('alpha', a2a('a1'));
        s.seedRemote('alpha', daemon('d1'));
        s.seedRemote('beta', a2a('b1'));
      },
    );
    await openList(tester, h);
    await tester.runAsync(() => h.client.remoteAgents!.loadRuntimes());
    await tester.pumpAndSettle();
    expect(find.text('Agent a1'), findsOneWidget);
    expect(find.text('Daemon d1'), findsOneWidget);
    expect(find.text('Agent b1'), findsNothing);
    expect(
      find.textContaining('Runtime "rt-1" does not exist in this Workspace'),
      findsOneWidget,
    );
  });

  testWidgets('a Daemon agent shows its runtime and worker and can be edited '
      'and deleted', (tester) async {
    final h = await signedIn(
      tester,
      using: DaemonServerFake(),
      seed: (s) {
        final fake = s as DaemonServerFake;
        fake.seedRuntime('alpha', DaemonRuntime(id: 'rt-1', name: 'Build box'));
        fake.onlineWorkers['alpha']!.add(
          DaemonStatus(
            daemonRuntimeId: 'rt-1',
            name: 'laptop',
            acpRuntimes: ['opencode', 'codex'],
            state: DaemonStatus_State.STATE_ONLINE,
          ),
        );
        s.seedRemote('alpha', daemon('d1'));
      },
    );
    await openDetail(tester, h, 'd1');
    await waitFor(
      tester,
      () => h.client.remoteAgents!.runtimeState == ConfigLoad.ready,
    );
    expect(find.textContaining('Build box (rt-1)'), findsOneWidget);
    expect(find.textContaining('Worker: connected (laptop)'), findsOneWidget);
    expect(find.byKey(const Key('remote-reference-problem')), findsNothing);
    expect(find.byKey(const Key('edit-remote')), findsOneWidget);
    expect(find.byKey(const Key('delete-remote')), findsOneWidget);
    expect(find.byKey(const Key('probe-remote')), findsOneWidget);
  });

  testWidgets('a Daemon agent without a worker or runtime says what is wrong', (
    tester,
  ) async {
    final h = await signedIn(
      tester,
      using: DaemonServerFake(),
      seed: (s) => s.seedRemote('alpha', daemon('d1')),
    );
    await openDetail(tester, h, 'd1');
    await waitFor(
      tester,
      () => h.client.remoteAgents!.runtimeState == ConfigLoad.ready,
    );
    expect(find.byKey(const Key('remote-reference-problem')), findsOneWidget);
    expect(
      find.textContaining('does not exist in this Workspace'),
      findsWidgets,
    );
  });

  testWidgets('adding a Daemon agent stores the runtime and executor, no URL', (
    tester,
  ) async {
    final h = await signedIn(
      tester,
      using: DaemonServerFake(),
      seed: (s) => (s as DaemonServerFake)
        ..seedRuntime('alpha', DaemonRuntime(id: 'rt-1', name: 'Build box')),
    );
    await openList(tester, h);
    await tapScreen(tester, find.byKey(const Key('add-remote')));
    await enter(tester, const Key('remote-id'), 'dd');
    await enter(tester, const Key('remote-name'), 'Coder');
    await tapReal(tester, find.text('Daemon'));
    await waitFor(
      tester,
      () => h.client.remoteAgents!.runtimeState == ConfigLoad.ready,
    );
    expect(find.byKey(const Key('remote-url')), findsNothing);
    await tapScreen(
      tester,
      find.byKey(const Key('remote-daemon-runtime-field')),
    );
    await tapScreen(tester, find.text('Build box (rt-1)').last);
    await tapScreen(tester, find.text('codex'));
    await tapReal(tester, find.byKey(const Key('save-remote')));
    await waitFor(tester, () => h.server.count('CreateRemoteAgent') == 1);
    final stored = h.server.remotes['alpha']!['dd']!;
    expect(stored.protocol, RemoteAgentProtocol.REMOTE_AGENT_PROTOCOL_DAEMON);
    expect(stored.daemonRuntimeId, 'rt-1');
    expect(stored.acpRuntime, 'codex');
    expect(stored.hasUrl(), isFalse);
  });

  testWidgets('saving a Daemon agent for a runtime that is gone is refused '
      'before any request', (tester) async {
    final h = await signedIn(
      tester,
      using: DaemonServerFake(),
      seed: (s) => s.seedRemote('alpha', daemon('d1')),
    );
    await openDetail(tester, h, 'd1');
    await waitFor(
      tester,
      () => h.client.remoteAgents!.runtimeState == ConfigLoad.ready,
    );
    await tapScreen(tester, find.byKey(const Key('edit-remote')));
    await enter(tester, const Key('remote-name'), 'Renamed');
    await tapReal(tester, find.byKey(const Key('save-remote')));
    await tester.pumpAndSettle();
    expect(
      find.textContaining('does not exist in this Workspace'),
      findsWidgets,
    );
    expect(h.server.count('UpdateRemoteAgent'), 0);
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

  testWidgets('clearing the password asks first and names the agent; cancel '
      'sends nothing', (tester) async {
    final h = await signedIn(
      tester,
      seed: (s) => s.seedRemote('alpha', opencode('o1')),
    );
    await openDetail(tester, h, 'o1');
    await tapScreen(tester, find.byKey(const Key('edit-remote')));
    // A username needs a password, so clearing one clears the other.
    await enter(tester, const Key('remote-username'), '');
    await tapScreen(
      tester,
      find.descendant(
        of: find.byKey(const Key('secret-choice-Password')),
        matching: find.text('Clear'),
      ),
    );
    await tapReal(tester, find.byKey(const Key('save-remote')));
    await tester.pumpAndSettle();
    expect(find.text('Discard stored credentials?'), findsOneWidget);
    expect(find.textContaining('"Code o1"'), findsOneWidget);
    expect(find.textContaining('Clear the stored password'), findsOneWidget);
    await tapDialog(tester, find.text('Cancel'));
    expect(h.server.count('UpdateRemoteAgent'), 0);
    expect(h.server.remotes['alpha']!['o1']!.password, 'oc-very-secret');

    await tapReal(tester, find.byKey(const Key('save-remote')));
    await tester.pumpAndSettle();
    await tapDialogReal(tester, find.text('Save and discard'));
    await waitFor(tester, () => h.server.count('UpdateRemoteAgent') == 1);
    expect(h.server.remotes['alpha']!['o1']!.password, isEmpty);
  });

  testWidgets('switching an OpenCode agent to A2A asks before discarding its '
      'credentials', (tester) async {
    final h = await signedIn(
      tester,
      seed: (s) => s.seedRemote('alpha', opencode('o1')),
    );
    await openDetail(tester, h, 'o1');
    await tapScreen(tester, find.byKey(const Key('edit-remote')));
    await tapScreen(
      tester,
      find.descendant(
        of: find.byKey(const Key('remote-kind')),
        matching: find.text('A2A'),
      ),
    );
    await tapReal(tester, find.byKey(const Key('save-remote')));
    await tester.pumpAndSettle();
    expect(find.text('Discard stored credentials?'), findsOneWidget);
    expect(find.textContaining('Switch to A2A'), findsOneWidget);
    await tapDialog(tester, find.text('Cancel'));
    expect(h.server.count('UpdateRemoteAgent'), 0);
  });

  testWidgets('a rename or a rotated password asks nothing', (tester) async {
    final h = await signedIn(
      tester,
      seed: (s) => s.seedRemote('alpha', opencode('o1')),
    );
    await openDetail(tester, h, 'o1');
    await tapScreen(tester, find.byKey(const Key('edit-remote')));
    await enter(tester, const Key('remote-name'), 'Renamed');
    await tapScreen(
      tester,
      find.descendant(
        of: find.byKey(const Key('secret-choice-Password')),
        matching: find.text('Replace'),
      ),
    );
    await enter(tester, const Key('secret-value-Password'), 'rotated');
    await tapReal(tester, find.byKey(const Key('save-remote')));
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 400));
    expect(find.text('Discard stored credentials?'), findsNothing);
    await waitFor(tester, () => h.server.count('UpdateRemoteAgent') == 1);
    expect(h.server.remotes['alpha']!['o1']!.password, 'rotated');
  });

  for (final outcome in ['denied', 'lost reply']) {
    testWidgets('a typed password is forgotten after a $outcome save and '
        'never sent again by itself', (tester) async {
      final h = await signedIn(
        tester,
        seed: (s) => s.seedRemote('alpha', opencode('o1')),
      );
      await openDetail(tester, h, 'o1');
      await tapScreen(tester, find.byKey(const Key('edit-remote')));
      await enter(tester, const Key('remote-name'), 'Kept name');
      await tapScreen(
        tester,
        find.descendant(
          of: find.byKey(const Key('secret-choice-Password')),
          matching: find.text('Replace'),
        ),
      );
      await enter(tester, const Key('secret-value-Password'), 'typed-secret');
      if (outcome == 'denied') {
        h.server.refusals['UpdateRemoteAgent'] = (
          403,
          'permission_denied',
          'no',
        );
      } else {
        h.server.loseReply = true;
      }
      await tapReal(tester, find.byKey(const Key('save-remote')));
      await waitFor(tester, () => h.server.count('UpdateRemoteAgent') == 1);

      String typed(Key key) => tester
          .widget<TextField>(
            find.descendant(
              of: find.byKey(key),
              matching: find.byType(TextField),
            ),
          )
          .controller!
          .text;
      expect(typed(const Key('remote-name')), 'Kept name');
      expect(typed(const Key('secret-value-Password')), isEmpty);
      expect(find.textContaining('typed-secret'), findsNothing);

      h.server.refusals.clear();
      h.server.loseReply = false;
      await tester.runAsync(
        () => Future<void>.delayed(const Duration(milliseconds: 100)),
      );
      await tapReal(tester, find.byKey(const Key('save-remote')));
      await tester.pumpAndSettle();
      expect(h.server.count('UpdateRemoteAgent'), 1);
    });
  }

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
