import 'dart:async';

import 'package:butter_mobile/app/client_context.dart';
import 'package:butter_mobile/core/config_resource_controller.dart';
import 'package:butter_mobile/features/daemons/daemon_pages.dart';
import 'package:butter_mobile/gen/agents/v1/agent.pb.dart';
import 'package:butter_mobile/gen/agents/v1/api_token.pb.dart';
import 'package:butter_mobile/gen/agents/v1/dashboard.pb.dart';
import 'package:fixnum/fixnum.dart';
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:protobuf/well_known_types/google/protobuf/duration.pb.dart'
    as pbd;

import 'support/config_rig.dart';
import 'support/daemon_server.dart';

DaemonTaskInFlight task(
  String id, {
  int progress = 40,
  String step = 'running tests',
  int seconds = 75,
}) => DaemonTaskInFlight(
  taskId: id,
  daemonRuntimeId: 'rt1',
  daemonName: 'Build box',
  acpRuntime: 'codex',
  agentName: 'Coder $id',
  currentStep: step,
  progress: progress,
  elapsed: pbd.Duration(seconds: Int64(seconds)),
);

DaemonServerFake fake(ConfigRig h) => h.server as DaemonServerFake;

Future<ConfigRig> rig(
  WidgetTester tester, {
  void Function(DaemonServerFake)? seed,
  Duration? poll,
}) => signedIn(
  tester,
  using: DaemonServerFake(),
  daemonPollInterval: poll,
  seed: (s) => seed?.call(s as DaemonServerFake),
);

/// Builds [page] in the real zone so its own first reads run on real timers.
Future<void> open(WidgetTester tester, Widget page) async {
  await tester.runAsync(() => tester.pumpWidget(MaterialApp(home: page)));
  await tester.pump();
}

/// A screen that opens the detail of `rt1`, so leaving and reopening it goes
/// through real navigation.
class _Host extends StatelessWidget {
  const _Host(this.client);
  final ClientContext client;
  @override
  Widget build(BuildContext context) => Scaffold(
    body: Builder(
      builder: (context) => TextButton(
        key: const Key('open-detail'),
        onPressed: () => Navigator.push(
          context,
          MaterialPageRoute<void>(
            builder: (_) => DaemonRuntimeDetailPage(client: client, id: 'rt1'),
          ),
        ),
        child: const Text('open'),
      ),
    ),
  );
}

Future<void> openDetailFromHost(WidgetTester tester, ConfigRig h) async {
  // The route is built, and so makes its first reads, in the real zone.
  await tester.runAsync(() async {
    await tester.tap(find.byKey(const Key('open-detail')));
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 400));
  });
  await waitFor(tester, () => h.client.daemons!.tokenState == ConfigLoad.ready);
}

void leaveDetail(WidgetTester tester) {
  tester.state<NavigatorState>(find.byType(Navigator)).pop();
}

void main() {
  testWidgets('lists this Workspace\'s runtimes with their worker state', (
    tester,
  ) async {
    final h = await rig(
      tester,
      seed: (s) {
        s.seedRuntime('alpha', DaemonRuntime(id: 'rt1', name: 'Build box'));
        s.seedRuntime('alpha', DaemonRuntime(id: 'rt2', name: 'Spare'));
        s.seedRuntime('beta', DaemonRuntime(id: 'rtb', name: 'Other team'));
        s.onlineWorkers['alpha']!.add(
          DaemonStatus(
            daemonRuntimeId: 'rt1',
            name: 'laptop',
            activeTasks: 2,
            state: DaemonStatus_State.STATE_ONLINE,
          ),
        );
      },
    );
    await open(tester, DaemonRuntimeListPage(client: h.client));
    await tester.runAsync(() async {
      await h.client.daemons!.load();
      await h.client.daemons!.loadOnline();
    });
    await tester.pumpAndSettle();
    expect(find.text('Build box'), findsOneWidget);
    expect(find.text('Spare'), findsOneWidget);
    expect(find.text('Other team'), findsNothing);
    expect(
      find.textContaining('Online · laptop · 2 active tasks'),
      findsOneWidget,
    );
    expect(
      find.textContaining('Offline: no worker is connected'),
      findsOneWidget,
    );
  });

  testWidgets('adding a runtime stores it; a duplicate ID is refused', (
    tester,
  ) async {
    final h = await rig(tester);
    await open(tester, DaemonFormPage(client: h.client));
    await enter(tester, const Key('daemon-id'), 'rt1');
    await enter(tester, const Key('daemon-name'), 'Build box');
    await tapReal(tester, find.byKey(const Key('save-daemon')));
    await waitFor(tester, () => fake(h).runtimes['alpha']!.isNotEmpty);
    expect(fake(h).runtimes['alpha']!['rt1']!.name, 'Build box');
    expect(fake(h).runtimes['beta'], isEmpty);
  });

  testWidgets('a form without a name is refused before any request', (
    tester,
  ) async {
    final h = await rig(tester);
    await open(tester, DaemonFormPage(client: h.client));
    await enter(tester, const Key('daemon-id'), 'rt1');
    await tapReal(tester, find.byKey(const Key('save-daemon')));
    await tester.pumpAndSettle();
    expect(find.text('Enter a name.'), findsWidgets);
    expect(fake(h).daemonCount('CreateDaemonRuntime'), 0);
  });

  group('runtime detail', () {
    DaemonRuntime stored() => DaemonRuntime(id: 'rt1', name: 'Build box');

    testWidgets(
      'shows the worker, bounds what it reports, lists only worker tokens',
      (tester) async {
        final h = await rig(
          tester,
          seed: (s) {
            s.seedRuntime('alpha', stored());
            s.onlineWorkers['alpha']!.add(
              DaemonStatus(
                daemonRuntimeId: 'rt1',
                name: 'laptop',
                version: '1.2.3',
                os: 'linux',
                executors: ['codex', 'x' * 500],
                state: DaemonStatus_State.STATE_ONLINE,
              ),
            );
            s.tokens['alpha']!.addAll([
              APIToken(
                id: 'u1',
                name: 'personal cli token',
                prefix: 'btr_user',
                kind: APITokenKind.API_TOKEN_KIND_USER,
              ),
              APIToken(
                id: 'w1',
                name: 'Worker for Build box',
                prefix: 'btr_w1',
                kind: APITokenKind.API_TOKEN_KIND_DAEMON,
                daemonRuntimeId: 'rt1',
              ),
            ]);
          },
        );
        await open(
          tester,
          DaemonRuntimeDetailPage(client: h.client, id: 'rt1'),
        );
        await waitFor(
          tester,
          () =>
              h.client.daemons!.tokenState == ConfigLoad.ready &&
              h.client.daemons!.onlineState == ConfigLoad.ready,
        );
        expect(find.textContaining('Online · laptop'), findsOneWidget);
        expect(find.textContaining('Version 1.2.3 · linux'), findsOneWidget);
        expect(find.textContaining('x' * 100), findsNothing);
        expect(find.textContaining('btr_w1'), findsOneWidget);
        expect(find.textContaining('personal cli token'), findsNothing);
        expect(find.textContaining('btr_user'), findsNothing);
      },
    );

    testWidgets('a minted secret is shown once and gone after dismissal', (
      tester,
    ) async {
      final h = await rig(
        tester,
        seed: (s) => s.seedRuntime('alpha', stored()),
      );
      await open(tester, DaemonRuntimeDetailPage(client: h.client, id: 'rt1'));
      await waitFor(
        tester,
        () => h.client.daemons!.tokenState == ConfigLoad.ready,
      );
      await tapReal(tester, find.byKey(const Key('mint-token')));
      await tester.pumpAndSettle();
      expect(find.text('Create a worker token?'), findsOneWidget);
      await tapDialogReal(tester, find.text('Create token'));
      await waitFor(tester, () => h.client.daemons!.minted != null);
      expect(find.byKey(const Key('minted-secret')), findsOneWidget);
      expect(find.text('worker-secret-1'), findsOneWidget);
      // Not offered again while it is on screen.
      expect(
        tester
            .widget<FilledButton>(find.byKey(const Key('mint-token')))
            .onPressed,
        isNull,
      );
      await tapScreen(tester, find.byKey(const Key('dismiss-secret')));
      expect(find.text('worker-secret-1'), findsNothing);
      expect(find.byKey(const Key('minted-secret-panel')), findsNothing);
      expect(find.textContaining('btr_w1'), findsOneWidget);
    });

    testWidgets('the app going to the background wipes a shown secret', (
      tester,
    ) async {
      final h = await rig(
        tester,
        seed: (s) => s.seedRuntime('alpha', stored()),
      );
      await open(tester, DaemonRuntimeDetailPage(client: h.client, id: 'rt1'));
      await waitFor(
        tester,
        () => h.client.daemons!.tokenState == ConfigLoad.ready,
      );
      await tapReal(tester, find.byKey(const Key('mint-token')));
      await tester.pumpAndSettle();
      await tapDialogReal(tester, find.text('Create token'));
      await waitFor(tester, () => h.client.daemons!.minted != null);
      expect(find.text('worker-secret-1'), findsOneWidget);
      h.client.suspend();
      await tester.pump();
      expect(find.text('worker-secret-1'), findsNothing);
      expect(find.byKey(const Key('minted-secret-panel')), findsNothing);
    });

    testWidgets(
      'a lost mint reply shows no secret, a notice, and offers no second mint',
      (tester) async {
        final h = await rig(
          tester,
          seed: (s) => s.seedRuntime('alpha', stored()),
        );
        fake(h).lostReplies.add('CreateDaemonRuntimeToken');
        await open(
          tester,
          DaemonRuntimeDetailPage(client: h.client, id: 'rt1'),
        );
        await waitFor(
          tester,
          () => h.client.daemons!.tokenState == ConfigLoad.ready,
        );
        await tapReal(tester, find.byKey(const Key('mint-token')));
        await tester.pumpAndSettle();
        await tapDialogReal(tester, find.text('Create token'));
        await waitFor(tester, () => h.client.daemons!.unconfirmed);
        expect(find.byKey(const Key('minted-secret')), findsNothing);
        expect(find.byKey(const Key('config-write-notice')), findsOneWidget);
        expect(find.textContaining('never shown'), findsOneWidget);
        expect(
          tester
              .widget<FilledButton>(find.byKey(const Key('mint-token')))
              .onPressed,
          isNull,
        );
        expect(fake(h).daemonCount('CreateDaemonRuntimeToken'), 1);
        // Refreshing reads the tokens, which settles the notice and shows the
        // token the person did not see the secret of.
        await tapReal(tester, find.byKey(const Key('refresh-daemon-item')));
        await waitFor(tester, () => !h.client.daemons!.unconfirmed);
        expect(find.textContaining('btr_w1'), findsOneWidget);
      },
    );

    testWidgets('revoking asks first and a refusal names the role', (
      tester,
    ) async {
      final h = await rig(
        tester,
        seed: (s) {
          s.seedRuntime('alpha', stored());
          s.tokens['alpha']!.add(
            APIToken(
              id: 'w1',
              name: 'w',
              prefix: 'btr_w1',
              kind: APITokenKind.API_TOKEN_KIND_DAEMON,
              daemonRuntimeId: 'rt1',
            ),
          );
          s.daemonRefusals['RevokeAPIToken'] = (
            403,
            'permission_denied',
            'forbidden',
          );
        },
      );
      await open(tester, DaemonRuntimeDetailPage(client: h.client, id: 'rt1'));
      await waitFor(
        tester,
        () => h.client.daemons!.tokenState == ConfigLoad.ready,
      );
      await tapReal(tester, find.byKey(const Key('revoke-w1')));
      await tester.pumpAndSettle();
      expect(find.text('Revoke worker token?'), findsOneWidget);
      await tapDialog(tester, find.text('Cancel'));
      expect(fake(h).daemonCount('RevokeAPIToken'), 0);
      await tapReal(tester, find.byKey(const Key('revoke-w1')));
      await tester.pumpAndSettle();
      await tapDialogReal(tester, find.text('Revoke'));
      await waitFor(tester, () => h.client.daemons!.writeError != null);
      expect(find.textContaining('owner or admin'), findsOneWidget);
    });

    testWidgets('deleting names the Remote Agents that run on the runtime', (
      tester,
    ) async {
      final h = await rig(
        tester,
        seed: (s) {
          s.seedRuntime('alpha', stored());
          s.seedRemote(
            'alpha',
            RemoteAgent(
              id: 'ra1',
              name: 'Nightly coder',
              protocol: RemoteAgentProtocol.REMOTE_AGENT_PROTOCOL_DAEMON,
              daemonRuntimeId: 'rt1',
              acpRuntime: 'codex',
            ),
          );
        },
      );
      await open(tester, DaemonRuntimeDetailPage(client: h.client, id: 'rt1'));
      await waitFor(tester, () => h.client.daemons!.item?.id == 'rt1');
      await tapReal(tester, find.byKey(const Key('delete-daemon')));
      await waitFor(
        tester,
        () => find.byType(AlertDialog).evaluate().isNotEmpty,
      );
      expect(find.textContaining('Nightly coder'), findsOneWidget);
      expect(find.textContaining('will fail'), findsOneWidget);
      await tapDialog(tester, find.text('Cancel'));
      expect(fake(h).daemonCount('DeleteDaemonRuntime'), 0);
    });
  });

  group('tasks', () {
    testWidgets(
      'show step, progress and elapsed; the poll keeps them current',
      (tester) async {
        final h = await rig(
          tester,
          poll: const Duration(milliseconds: 50),
          seed: (s) => s.taskList['alpha']!.add(task('t1')),
        );
        await open(tester, DaemonTasksPage(client: h.client));
        await waitFor(tester, () => h.client.daemons!.tasks.isNotEmpty);
        expect(find.text('Coder t1'), findsOneWidget);
        expect(
          find.textContaining('running tests · 40% · 1m 15s'),
          findsOneWidget,
        );
        fake(h).taskList['alpha']![0] = task(
          't1',
          progress: 80,
          step: 'pushing',
        );
        await waitFor(
          tester,
          () => h.client.daemons!.tasks.first.progress == 80,
        );
        expect(find.textContaining('pushing · 80%'), findsOneWidget);
        expect(fake(h).daemonCount('ListDaemonTasks'), greaterThan(1));
      },
    );

    testWidgets('cancel success says delivered only; the task stays listed', (
      tester,
    ) async {
      final h = await rig(
        tester,
        seed: (s) => s.taskList['alpha']!.add(task('t1')),
      );
      await open(tester, DaemonTasksPage(client: h.client));
      await waitFor(tester, () => h.client.daemons!.tasks.isNotEmpty);
      await tapReal(tester, find.byKey(const Key('cancel-t1')));
      await tester.pumpAndSettle();
      expect(find.text('Ask the worker to cancel?'), findsOneWidget);
      await tapDialog(tester, find.text('Cancel'));
      expect(fake(h).daemonCount('CancelDaemonTask'), 0);
      await tapReal(tester, find.byKey(const Key('cancel-t1')));
      await tester.pumpAndSettle();
      await tapDialogReal(tester, find.text('Send cancel'));
      await waitFor(tester, () => h.client.daemons!.cancels['t1'] != null);
      await waitFor(
        tester,
        () => h.client.daemons!.cancels['t1']!.sending == false,
      );
      expect(
        find.textContaining('Cancel delivered to runtime "rt1"'),
        findsOneWidget,
      );
      expect(find.textContaining('still running'), findsWidgets);
      expect(find.text('Coder t1'), findsOneWidget);
      expect(find.textContaining('cancelled'), findsNothing);
    });

    testWidgets(
      'a worker that cannot be reached is told apart from a gone task',
      (tester) async {
        final h = await rig(
          tester,
          seed: (s) {
            s.taskList['alpha']!.add(task('t1'));
            s.daemonRefusals['CancelDaemonTask'] = (500, 'internal', 'boom');
          },
        );
        await open(tester, DaemonTasksPage(client: h.client));
        await waitFor(tester, () => h.client.daemons!.tasks.isNotEmpty);
        await tapReal(tester, find.byKey(const Key('cancel-t1')));
        await tester.pumpAndSettle();
        await tapDialogReal(tester, find.text('Send cancel'));
        await waitFor(tester, () => h.client.daemons!.cancels['t1'] != null);
        await waitFor(
          tester,
          () => h.client.daemons!.cancels['t1']!.sending == false,
        );
        expect(
          find.textContaining('could not hand the cancel to the worker'),
          findsOneWidget,
        );
        expect(find.textContaining('boom'), findsNothing);
      },
    );

    testWidgets('detail follows the task and says what is unknown once it '
        'leaves', (tester) async {
      final h = await rig(
        tester,
        seed: (s) => s.taskList['alpha']!.add(task('t1')),
      );
      await open(tester, DaemonTasksPage(client: h.client));
      await waitFor(tester, () => h.client.daemons!.tasks.isNotEmpty);
      await tapScreen(tester, find.byKey(const Key('task-t1')));
      expect(find.byKey(const Key('task-step')), findsOneWidget);
      expect(find.textContaining('Progress: 40%'), findsOneWidget);
      fake(h).taskList['alpha']!.clear();
      await tapReal(tester, find.byKey(const Key('task-refresh')));
      await waitFor(tester, () => h.client.daemons!.ended.isNotEmpty);
      expect(find.byKey(const Key('task-gone')), findsOneWidget);
      expect(
        find.textContaining('does not report whether it finished'),
        findsOneWidget,
      );
    });

    testWidgets('a failed refresh keeps the list and says it may be old', (
      tester,
    ) async {
      final h = await rig(
        tester,
        seed: (s) => s.taskList['alpha']!.add(task('t1')),
      );
      await open(tester, DaemonTasksPage(client: h.client));
      await waitFor(tester, () => h.client.daemons!.tasks.isNotEmpty);
      fake(h).daemonRefusals['ListDaemonTasks'] = (503, 'unavailable', 'down');
      await tapReal(tester, find.byKey(const Key('refresh-tasks')));
      await waitFor(tester, () => h.client.daemons!.taskError != null);
      expect(find.text('Coder t1'), findsOneWidget);
      expect(find.textContaining('may be out of date'), findsOneWidget);
    });
  });

  testWidgets('diagnostics show server-wide metrics and say so', (
    tester,
  ) async {
    final h = await rig(tester);
    await open(tester, BridgeDiagnosticsPage(client: h.client));
    await waitFor(
      tester,
      () => h.client.daemons!.diagnosticsState == ConfigLoad.ready,
    );
    expect(find.textContaining('CPU: 12.5%'), findsOneWidget);
    expect(find.textContaining('Memory: 256 MiB of 1.0 GiB'), findsOneWidget);
    expect(find.textContaining('Goroutines: 42'), findsOneWidget);
    expect(find.textContaining('shared by every Workspace'), findsOneWidget);
  });

  group('worker token screen lifetime', () {
    DaemonRuntime stored() => DaemonRuntime(id: 'rt1', name: 'Build box');

    testWidgets(
      'leaving the screen discards the secret; reopening shows none',
      (tester) async {
        final h = await rig(
          tester,
          seed: (s) => s.seedRuntime('alpha', stored()),
        );
        await open(tester, _Host(h.client));
        await openDetailFromHost(tester, h);
        await tapReal(tester, find.byKey(const Key('mint-token')));
        await tester.pumpAndSettle();
        await tapDialogReal(tester, find.text('Create token'));
        await waitFor(tester, () => h.client.daemons!.minted != null);
        expect(find.text('worker-secret-1'), findsOneWidget);
        // The secret is on screen, so nothing may open another screen over it.
        expect(
          tester
              .widget<OutlinedButton>(find.byKey(const Key('edit-daemon')))
              .onPressed,
          isNull,
        );
        expect(
          tester
              .widget<TextButton>(find.byKey(const Key('delete-daemon')))
              .onPressed,
          isNull,
        );
        final secret = h.client.daemons!.minted!;
        leaveDetail(tester);
        await tester.pumpAndSettle();
        expect(h.client.daemons!.minted, isNull);
        expect(secret.secret, isNull);
        await openDetailFromHost(tester, h);
        expect(find.byKey(const Key('minted-secret-panel')), findsNothing);
        expect(find.text('worker-secret-1'), findsNothing);
        expect(find.textContaining('btr_w1'), findsOneWidget);
      },
    );

    testWidgets('a mint answered after the screen was left shows no secret', (
      tester,
    ) async {
      final h = await rig(
        tester,
        seed: (s) => s.seedRuntime('alpha', stored()),
      );
      final gate = Completer<void>();
      fake(h).daemonBefore = (m) async {
        if (m == 'CreateDaemonRuntimeToken') await gate.future;
      };
      await open(tester, _Host(h.client));
      await openDetailFromHost(tester, h);
      await tapReal(tester, find.byKey(const Key('mint-token')));
      await tester.pumpAndSettle();
      await tapDialogReal(tester, find.text('Create token'));
      // The write is held, so its progress bar never settles: only real time
      // passes here, and the tree is pumped without waiting for it.
      for (var i = 0; i < 300; i++) {
        if (fake(h).daemonCount('CreateDaemonRuntimeToken') == 1) break;
        await tester.runAsync(
          () => Future<void>.delayed(const Duration(milliseconds: 10)),
        );
      }
      expect(fake(h).daemonCount('CreateDaemonRuntimeToken'), 1);
      expect(h.client.daemons!.writing, isTrue);
      leaveDetail(tester);
      for (var i = 0; i < 5; i++) {
        await tester.pump(const Duration(milliseconds: 200));
      }
      expect(find.byType(DaemonRuntimeDetailPage), findsNothing);
      gate.complete();
      await waitFor(tester, () => !h.client.daemons!.writing);
      expect(h.client.daemons!.minted, isNull);
      expect(h.client.daemons!.writeError, contains('discarded'));
      expect(h.client.daemons!.unconfirmed, isTrue);
      await openDetailFromHost(tester, h);
      expect(find.byKey(const Key('minted-secret-panel')), findsNothing);
      expect(find.text('worker-secret-1'), findsNothing);
      // The token is listed so the person can revoke it; none is made again.
      expect(find.textContaining('btr_w1'), findsOneWidget);
      expect(fake(h).daemonCount('CreateDaemonRuntimeToken'), 1);
    });

    Future<void> openMintDialog(WidgetTester tester, dynamic h) async {
      await open(tester, DaemonRuntimeDetailPage(client: h.client, id: 'rt1'));
      await waitFor(
        tester,
        () => h.client.daemons!.tokenState == ConfigLoad.ready,
      );
      await tapReal(tester, find.byKey(const Key('mint-token')));
      await tester.pumpAndSettle();
    }

    List<CreateDaemonRuntimeTokenRequest> mintRequests(dynamic h) => [
      for (final c in fake(h).daemonCalls)
        if (c.method == 'CreateDaemonRuntimeToken')
          CreateDaemonRuntimeTokenRequest.fromBuffer(c.body),
    ];

    testWidgets('the dialog refuses an expiry that is not a positive number '
        'of hours, before any request', (tester) async {
      final h = await rig(
        tester,
        seed: (s) => s.seedRuntime('alpha', stored()),
      );
      await openMintDialog(tester, h);
      await tester.enterText(find.byKey(const Key('mint-name')), 'ci box');
      for (final text in ['0', '-1', 'abc', 'NaN', 'Infinity', '1e3', '1,5']) {
        await tester.enterText(
          find.byKey(const Key('mint-expiry-hours')),
          text,
        );
        await tester.tap(find.text('Create token'));
        await tester.pump();
        expect(
          find.textContaining('positive number of hours'),
          findsOneWidget,
          reason: '"$text"',
        );
        expect(fake(h).daemonCount('CreateDaemonRuntimeToken'), 0);
      }
      // The dialog is still open and nothing was minted.
      expect(find.byKey(const Key('mint-expiry-hours')), findsOneWidget);
      expect(h.client.daemons!.minted, isNull);
    });

    for (final (text, seconds, nanos) in [
      ('1', 3600, 0),
      ('12', 12 * 3600, 0),
      ('0.5', 1800, 0),
      ('720', 30 * 86400, 0),
      ('0.0005', 1, 800000000),
    ]) {
      testWidgets('the dialog sends $text hours as $seconds s $nanos ns', (
        tester,
      ) async {
        final h = await rig(
          tester,
          seed: (s) => s.seedRuntime('alpha', stored()),
        );
        await openMintDialog(tester, h);
        await tester.enterText(find.byKey(const Key('mint-name')), 'ci box');
        await tester.enterText(
          find.byKey(const Key('mint-expiry-hours')),
          text,
        );
        await tapDialogReal(tester, find.text('Create token'));
        await waitFor(tester, () => h.client.daemons!.minted != null);
        final request = mintRequests(h).single;
        expect(request.name, 'ci box');
        expect(request.hasTtl(), isTrue);
        expect(request.ttl.seconds.toInt(), seconds);
        expect(request.ttl.nanos, nanos);
      });
    }

    testWidgets('an empty expiry sends none: the token does not expire', (
      tester,
    ) async {
      final h = await rig(
        tester,
        seed: (s) => s.seedRuntime('alpha', stored()),
      );
      await open(tester, DaemonRuntimeDetailPage(client: h.client, id: 'rt1'));
      await waitFor(
        tester,
        () => h.client.daemons!.tokenState == ConfigLoad.ready,
      );
      await tapReal(tester, find.byKey(const Key('mint-token')));
      await tester.pumpAndSettle();
      expect(
        find.textContaining('does not expire until you revoke'),
        findsOneWidget,
      );
      await tapDialogReal(tester, find.text('Create token'));
      await waitFor(tester, () => h.client.daemons!.minted != null);
      final request = CreateDaemonRuntimeTokenRequest.fromBuffer(
        fake(h).daemonCalls
            .firstWhere((c) => c.method == 'CreateDaemonRuntimeToken')
            .body,
      );
      expect(request.hasTtl(), isFalse);
    });
  });

  group('runtime labels', () {
    testWidgets('are added, edited and removed as rows and stored whole', (
      tester,
    ) async {
      final h = await rig(
        tester,
        seed: (s) => s.seedRuntime(
          'alpha',
          DaemonRuntime(id: 'rt1', name: 'Build box')
            ..labels['team'] = 'platform'
            ..labels['env'] = 'prod'
            ..createdBy = 'bob',
        ),
      );
      await open(
        tester,
        DaemonFormPage(
          client: h.client,
          base: fake(h).runtimes['alpha']!['rt1']!.deepCopy(),
        ),
      );
      // Sorted by key: env, team.
      expect(find.text('env'), findsOneWidget);
      expect(find.text('platform'), findsOneWidget);
      await enter(tester, const Key('label-value-0'), 'staging');
      await tapReal(tester, find.byKey(const Key('remove-label-1')));
      await tapReal(tester, find.byKey(const Key('add-label')));
      await enter(tester, const Key('label-key-1'), 'region');
      await enter(tester, const Key('label-value-1'), 'eu');
      await tapReal(tester, find.byKey(const Key('save-daemon')));
      await waitFor(
        tester,
        () => fake(h).runtimes['alpha']!['rt1']!.labels.containsKey('region'),
      );
      final saved = fake(h).runtimes['alpha']!['rt1']!;
      expect(saved.labels, {'env': 'staging', 'region': 'eu'});
      expect(saved.createdBy, 'bob');
    });

    testWidgets('a repeated key is refused before any request', (tester) async {
      final h = await rig(tester);
      await open(tester, DaemonFormPage(client: h.client));
      await enter(tester, const Key('daemon-id'), 'rt1');
      await enter(tester, const Key('daemon-name'), 'Build box');
      await tapReal(tester, find.byKey(const Key('add-label')));
      await tapReal(tester, find.byKey(const Key('add-label')));
      await enter(tester, const Key('label-key-0'), 'team');
      await enter(tester, const Key('label-key-1'), 'team');
      await tapReal(tester, find.byKey(const Key('save-daemon')));
      await tester.pumpAndSettle();
      expect(find.textContaining('used twice'), findsOneWidget);
      expect(fake(h).daemonCount('CreateDaemonRuntime'), 0);
    });
  });

  group('task polling under a slow server', () {
    testWidgets('asks one question at a time and the answers arrive', (
      tester,
    ) async {
      final h = await rig(
        tester,
        poll: const Duration(milliseconds: 20),
        seed: (s) => s.taskList['alpha']!.add(task('t1')),
      );
      var inFlight = 0, most = 0;
      fake(h).daemonBefore = (m) async {
        if (m != 'ListDaemonTasks') return;
        inFlight++;
        if (inFlight > most) most = inFlight;
        // Slower than several poll intervals.
        await Future<void>.delayed(const Duration(milliseconds: 120));
        inFlight--;
      };
      await open(tester, DaemonTasksPage(client: h.client));
      await waitFor(tester, () => h.client.daemons!.tasks.isNotEmpty);
      for (var i = 1; i <= 3; i++) {
        fake(h).taskList['alpha']![0] = task('t1', progress: 40 + i * 10);
        await waitFor(
          tester,
          () => h.client.daemons!.tasks.first.progress == 40 + i * 10,
        );
      }
      expect(most, 1);
      expect(find.textContaining('70%'), findsOneWidget);
      // Leaving stops the cycle: no read is begun after the page is gone.
      await tester.pumpWidget(const MaterialApp(home: SizedBox()));
      await tester.runAsync(
        () => Future<void>.delayed(const Duration(milliseconds: 300)),
      );
      final after = fake(h).daemonCount('ListDaemonTasks');
      await tester.runAsync(
        () => Future<void>.delayed(const Duration(milliseconds: 200)),
      );
      expect(fake(h).daemonCount('ListDaemonTasks'), after);
    });
  });
}
