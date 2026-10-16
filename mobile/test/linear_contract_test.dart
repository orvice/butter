import 'dart:io' as io;

import 'package:butter_mobile/core/butter_api.dart';
import 'package:butter_mobile/core/server_address.dart';
import 'package:butter_mobile/features/linear/linear_controller.dart';
import 'package:butter_mobile/features/linear/linear_draft.dart';
import 'package:butter_mobile/gen/agents/v1/linear.pb.dart';
import 'package:connectrpc/connect.dart';
import 'package:flutter_test/flutter_test.dart';

const workspace = 'linear';
const memberWorkspace = 'linear-member';
const seededClientId = 'contract-linear-client';

LinearApp seededApp(Iterable<LinearApp> apps) =>
    apps.firstWhere((a) => a.clientId == seededClientId);

void main() {
  const server = String.fromEnvironment('BUTTER_TEST_SERVER');
  for (final prefix in ['', '/butter']) {
    group(
      'Linear production contract $prefix',
      skip: server.isEmpty ? 'Run tool/contract.sh.' : false,
      () {
        final base = '$server$prefix';
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

        LinearController linear() {
          final c = LinearController(
            newApi: () => ButterApi(ServerAddress.parse(base, allowHttp: true)),
            token: token,
            workspaceId: workspace,
            canManage: () => true,
            onUnauthenticated: () {},
          );
          addTearDown(c.dispose);
          return c;
        }

        test('lists seeded app without leaking secrets', () async {
          final c = linear();
          await c.loadApps();
          final app = seededApp(c.apps);
          expect(app.displayName.isNotEmpty, isTrue);
          final wire = app.writeToJson();
          expect(wire.contains('lin-client-secret'), isFalse);
          expect(wire.contains('lin-webhook-secret'), isFalse);
        });

        test('member workspace create is permission denied', () async {
          final memberApi = ButterApi(
            ServerAddress.parse(base, allowHttp: true),
          );
          addTearDown(memberApi.close);
          final draft = LinearApp()
            ..displayName = 'Member app'
            ..clientId =
                'member-client-$prefix-${DateTime.now().microsecondsSinceEpoch}'
            ..agentId = 'linear-agent';
          expect(
            () => memberApi.createLinearApp(
              token,
              memberWorkspace,
              draft,
              clientSecret: 's',
              webhookSecret: 'w',
            ),
            throwsA(
              isA<ConnectException>().having(
                (e) => e.code,
                'code',
                Code.permissionDenied,
              ),
            ),
          );
        });

        test('revision conflict on stale update', () async {
          final c = linear();
          await c.loadApps();
          final created = await api.createLinearApp(
            token,
            workspace,
            LinearApp()
              ..displayName = 'CAS probe $prefix'
              ..clientId =
                  'cas-client-$prefix-${DateTime.now().microsecondsSinceEpoch}'
              ..agentId = 'linear-agent',
            clientSecret: 'secret',
            webhookSecret: 'webhook',
          );
          final stale = LinearAppDraft.fromStored(created);
          stale.displayName = 'First rename';
          expect(await c.updateApp(stale), isTrue);
          stale.displayName = 'Second on old revision';
          expect(await c.updateApp(stale), isFalse);
          expect(c.writeError, contains('Reload'));
        });

        test('resend failed delivery without invoking agent', () async {
          final c = linear();
          await c.loadRecords();
          final failed = c.records.where(
            (r) =>
                r.status ==
                LinearProcessingStatus.LINEAR_PROCESSING_STATUS_FAILED,
          );
          expect(failed, isNotEmpty);
          expect(await c.resendReply(failed.first.id), isTrue);
          expect(
            c.selectedRecord?.status,
            LinearProcessingStatus.LINEAR_PROCESSING_STATUS_SUCCEEDED,
          );
        });

        test('put credentials keep absent does not clear', () async {
          final c = linear();
          await c.loadApps();
          final app = seededApp(c.apps);
          expect(await c.putCredentials(app.id), isTrue);
          await c.loadApps();
          expect(seededApp(c.apps).clientSecretSet, isTrue);
        });

        test(
          'oauth callback adds installation without app token in url',
          () async {
            final c = linear();
            await c.loadApps();
            final app = seededApp(c.apps);
            await c.loadInstallations(app.id);
            final before = c.installations.length;
            final authorizeUrl = await api.beginLinearInstall(
              token,
              workspace,
              app.id,
              returnUrl: '/linear-apps/${app.id}',
            );
            expect(authorizeUrl.contains('token='), isFalse);
            final state = Uri.parse(authorizeUrl).queryParameters['state'];
            expect(state, isNotEmpty);
            final client = io.HttpClient();
            addTearDown(() => client.close(force: true));
            final req =
                await client.getUrl(
                    Uri.parse(
                      '$base/api/linear/oauth/callback?state=${Uri.encodeComponent(state!)}&code=contract-install-$prefix',
                    ),
                  )
                  ..followRedirects = false;
            final resp = await req.close();
            expect(resp.statusCode, io.HttpStatus.found);
            await resp.drain();
            await c.loadInstallations(app.id);
            expect(c.installations.length, greaterThan(before));
          },
        );
      },
    );
  }
}
