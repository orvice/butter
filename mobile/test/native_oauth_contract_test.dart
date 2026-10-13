import 'dart:convert';
import 'dart:io';

import 'package:butter_mobile/app/client_context.dart';
import 'package:butter_mobile/core/butter_api.dart';
import 'package:butter_mobile/core/server_address.dart';
import 'package:connectrpc/connect.dart' show Code, ConnectException;
import 'package:flutter_test/flutter_test.dart';

import 'support/fake_oauth_platform.dart';
import 'support/memory_storage.dart';

const providerCode = 'contract-provider-code';

/// What the dashboard's /auth/oauth/callback page does after the provider
/// returns the browser: a public Connect JSON CompleteOAuthFlow call.
Future<Map<String, dynamic>> completeAsWebPage(
  String server,
  String state, {
  String code = providerCode,
}) async {
  final client = HttpClient();
  try {
    final request = await client.postUrl(
      Uri.parse('$server/api/agents.v1.AuthService/CompleteOAuthFlow'),
    );
    request.headers.contentType = ContentType.json;
    request.write(
      jsonEncode({'provider': 'github', 'code': code, 'state': state}),
    );
    final response = await request.close();
    final body = await utf8.decodeStream(response);
    if (response.statusCode != 200) {
      throw StateError('CompleteOAuthFlow ${response.statusCode}: $body');
    }
    return jsonDecode(body) as Map<String, dynamic>;
  } finally {
    client.close(force: true);
  }
}

String? returnUrlOf(Map<String, dynamic> completed) =>
    (completed['native_handoff'] as Map<String, dynamic>?)?['return_url']
        as String?;

void main() {
  const server = String.fromEnvironment('BUTTER_TEST_SERVER');
  for (final prefix in ['', '/butter']) {
    group(
      'native OAuth against real Butter Go services at "$prefix"',
      () {
        late String url;
        late MemoryStorage storage;
        late FakeOAuthPlatform platform;
        late ClientContext context;

        setUp(() async {
          url = '$server$prefix';
          storage = MemoryStorage();
          platform = FakeOAuthPlatform();
          context = ClientContext(
            storage: storage,
            allowHttp: true,
            oauth: platform,
          );
          await context.start();
          await context.setServer(url);
        });
        tearDown(() => context.dispose());

        test(
          'authorize in the browser, return to the App, redeem the session',
          () async {
            expect(context.oauthProviders.map((p) => p.name), ['github']);
            await context.beginOAuth('github');

            final authorize = platform.opened.single;
            expect(authorize.host, 'provider.invalid');
            final state = authorize.queryParameters['state']!;
            expect(context.oauthWaitingFor, 'GitHub');

            // The browser leg ends at the server's fixed callback page.
            final completed = await completeAsWebPage(url, state);
            expect(completed['token'], anyOf(isNull, isEmpty));
            final returnUrl = returnUrlOf(completed)!;
            expect(returnUrl, startsWith('ee.orz.butter:/oauth/callback?'));
            expect(returnUrl, isNot(contains('token')));
            expect(Uri.parse(returnUrl).queryParameters['state'], state);

            platform.deliver(returnUrl);
            await context.oauthIdle;
            expect(context.error, isNull);
            expect(context.user!.username, 'mobile-test');
            expect(context.stage, ClientStage.workspaces);
            expect(
              context.workspaces.map((w) => w.id),
              containsAll(['alpha', 'beta']),
            );
            await context.selectWorkspace('alpha');
            expect(context.agents.single.agentId, 'shared');
            expect(storage.pending, isNull);

            // A repeated callback and a replayed browser leg change nothing.
            platform.deliver(returnUrl);
            await context.oauthIdle;
            expect(context.user!.username, 'mobile-test');
            await expectLater(completeAsWebPage(url, state), throwsStateError);
          },
        );

        test('an app that only saw the link cannot redeem it', () async {
          await context.beginOAuth('github');
          final state = platform.opened.single.queryParameters['state']!;
          final returnUrl = returnUrlOf(await completeAsWebPage(url, state))!;
          final handoff = Uri.parse(returnUrl).queryParameters['handoff']!;

          // Another app registered for the same scheme intercepts the link, but
          // the PKCE verifier never left the real App.
          final thief = ButterApi(ServerAddress.parse(url, allowHttp: true));
          addTearDown(thief.close);
          await expectLater(
            thief.redeemOAuth('github', state, handoff, 'a' * 43),
            throwsA(
              isA<ConnectException>().having(
                (e) => e.code,
                'code',
                Code.permissionDenied,
              ),
            ),
          );

          // The attempt spent the handoff: nobody is signed in by it now.
          platform.deliver(returnUrl);
          await context.oauthIdle;
          expect(context.user, isNull);
          expect(context.error, contains('could not be completed'));
          expect(storage.credential, isNull);
        });

        test(
          'a web sign-in on the same server still gets its session',
          () async {
            // A web client begins without `native`: the callback logs it in
            // directly and never hands back to an App.
            final api = ButterApi(ServerAddress.parse(url, allowHttp: true));
            addTearDown(api.close);
            final client = HttpClient();
            addTearDown(() => client.close(force: true));
            final request = await client.postUrl(
              Uri.parse('$url/api/agents.v1.AuthService/BeginOAuthFlow'),
            );
            request.headers.contentType = ContentType.json;
            request.write(jsonEncode({'provider': 'github'}));
            final begun =
                jsonDecode(await utf8.decodeStream(await request.close()))
                    as Map<String, dynamic>;

            final completed = await completeAsWebPage(
              url,
              begun['state'] as String,
            );
            expect(completed['token'], isNotEmpty);
            expect(returnUrlOf(completed), isNull);
            final me = await api.me(completed['token'] as String);
            expect(me.user.username, 'mobile-test');
          },
        );
      },
      skip: server.isEmpty
          ? 'Run tool/contract.sh against the real Go handlers.'
          : false,
    );
  }
}
