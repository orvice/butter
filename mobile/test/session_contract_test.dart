import 'package:butter_mobile/core/butter_api.dart';
import 'package:butter_mobile/core/server_address.dart';
import 'package:butter_mobile/features/sessions/session_controller.dart';
import 'package:butter_mobile/features/sessions/session_key.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  const server = String.fromEnvironment('BUTTER_TEST_SERVER');
  for (final prefix in ['', '/butter']) {
    group(
      'Session production contract $prefix',
      skip: server.isEmpty ? 'Run tool/contract.sh.' : false,
      () {
        final base = '$server$prefix';
        const ws = 'sessions';
        const peer = 'sessions-peer';
        late ButterApi api;
        late String token;

        setUp(() async {
          api = ButterApi(ServerAddress.parse(base, allowHttp: true));
          addTearDown(api.close);
          token = (await api.login(
            'mobile-test',
            'contract-test-password',
          )).token;
        });

        SessionController sessions({String workspace = ws}) {
          final c = SessionController(
            newApi: () => ButterApi(ServerAddress.parse(base, allowHttp: true)),
            token: token,
            workspaceId: workspace,
            onUnauthenticated: () {},
          );
          addTearDown(c.dispose);
          return c;
        }

        test('list, detail, rename, reply and pagination', () async {
          final ctl = sessions();
          await ctl.applyFilters();
          expect(ctl.sessions.map((s) => s.sessionId), contains('sess-own'));
          expect(ctl.sessions.map((s) => s.sessionId), contains('sess-bob'));
          const key = SessionKey(
            appName: 'web-chat',
            userId: 'mobile-test',
            sessionId: 'sess-own',
          );
          expect(await ctl.rename(key, 'Renamed title'), isTrue);
          await ctl.openSession(key);
          expect(ctl.detail!.events, isNotEmpty);
          final reply = await ctl.reply(
            key: key,
            agentId: 'helper',
            message: 'continue',
          );
          expect(reply.response, contains('fixture'));
        });

        test('owner may read another user session in the Workspace', () async {
          final detail = await api.getSession(
            token,
            ws,
            appName: 'web-chat',
            userId: 'bob',
            sessionId: 'sess-bob',
            numRecentEvents: 1,
          );
          expect(detail.session.sessionId, 'sess-bob');
        });

        test('workspace isolation hides peer workspace sessions', () async {
          final ctl = sessions();
          await ctl.loadPage();
          expect(
            ctl.sessions.map((s) => s.sessionId),
            isNot(contains('sess-peer')),
          );
          final peerCtl = sessions(workspace: peer);
          await peerCtl.loadPage();
          expect(
            peerCtl.sessions.map((s) => s.sessionId),
            contains('sess-peer'),
          );
        });

        test('missing agent on reply is refused', () async {
          final ctl = sessions();
          final key = SessionKey(
            appName: 'web-chat',
            userId: 'mobile-test',
            sessionId: 'sess-own',
          );
          final result = await ctl.reply(
            key: key,
            agentId: 'missing-agent',
            message: 'hi',
          );
          expect(result.response, isNull);
          expect(result.error, isNotNull);
        });

        test('delete removes the session', () async {
          final ctl = sessions();
          final key = SessionKey(
            appName: 'cron:nightly',
            userId: 'cron:nightly',
            sessionId: 'sess-cron',
          );
          await ctl.openSession(key);
          expect(await ctl.deleteSession(key), isTrue);
          await ctl.openSession(key);
          expect(ctl.detailState, SessionLoad.failed);
        });
      },
    );
  }
}
