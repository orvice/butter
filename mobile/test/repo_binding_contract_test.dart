import 'package:butter_mobile/core/butter_api.dart';
import 'package:butter_mobile/core/server_address.dart';
import 'package:butter_mobile/features/repo_binding/repo_binding_controller.dart';
import 'package:butter_mobile/features/repo_binding/repo_binding_draft.dart';
import 'package:butter_mobile/gen/agents/v1/repobinding.pb.dart';
import 'package:flutter_test/flutter_test.dart';

const workspace = 'repo';
const memberWorkspace = 'repo-member';

void main() {
  const server = String.fromEnvironment('BUTTER_TEST_SERVER');
  for (final prefix in ['', '/butter']) {
    group(
      'Repo binding production contract $prefix',
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

        RepoBindingController repo({String ws = workspace}) {
          final c = RepoBindingController(
            newApi: () => ButterApi(ServerAddress.parse(base, allowHttp: true)),
            token: token,
            workspaceId: ws,
            canManage: () => true,
            onUnauthenticated: () {},
          );
          addTearDown(c.dispose);
          return c;
        }

        test(
          'lists git hosts and saves binding with PAT validate export',
          skip: server.isEmpty
              ? 'Run tool/contract.sh.'
              : prefix.isNotEmpty
              ? 'Shared fixture state; export runs once at root prefix.'
              : false,
          () async {
            final c = repo();
            await c.loadGitHosts();
            expect(c.gitHosts.any((h) => h.id == 'contract-git-host'), isTrue);

            await c.load();
            if (!c.hasActiveRevision) {
              final draft = RepoBindingDraft.empty()
                ..gitHostId = 'contract-git-host'
                ..repository = 'acme/agents'
                ..branch = 'main';
              expect(await c.saveBinding(draft), isTrue, reason: c.writeError);
              expect(
                await c.setCredential('ghp_contract_test_pat'),
                isTrue,
                reason: c.writeError,
              );
              expect(await c.validate(), isTrue, reason: c.writeError);
              expect(
                c.binding!.status.state,
                RepoBindingConnectionState.REPO_BINDING_CONNECTION_STATE_OK,
              );

              final onboard = await c.onboard(
                RepoBindingOnboardingMode
                    .REPO_BINDING_ONBOARDING_MODE_EXPORT_CURRENT,
              );
              expect(onboard, isNotNull, reason: c.writeError);
              expect(onboard!.published, isTrue);
            }
            expect(c.binding!.activeCommitSha, isNotEmpty);
            expect(c.hasActiveRevision, isTrue);

            expect(await c.syncRepository(), isTrue, reason: c.writeError);
            await c.loadEntries();
            expect(c.entriesError, isNull);
          },
        );

        test('member cannot save binding', () async {
          final c = repo(ws: memberWorkspace);
          await c.load();
          final draft = RepoBindingDraft.empty()
            ..gitHostId = 'contract-git-host'
            ..repository = 'x/y'
            ..branch = 'main';
          expect(await c.saveBinding(draft), isFalse);
          expect(c.writeError, contains('owner or admin'));
        });

        test('denied credential fails validation', () async {
          final c = repo();
          await c.load();
          if (c.binding == null) {
            final draft = RepoBindingDraft.empty()
              ..gitHostId = 'contract-git-host'
              ..repository = 'acme/agents'
              ..branch = 'main';
            expect(await c.saveBinding(draft), isTrue, reason: c.writeError);
          }
          expect(await c.setCredential('deny-read'), isTrue);
          expect(await c.validate(), isTrue);
          expect(
            c.binding!.status.state,
            isNot(RepoBindingConnectionState.REPO_BINDING_CONNECTION_STATE_OK),
          );
        });
      },
    );
  }
}
