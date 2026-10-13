import 'dart:async';

import 'package:butter_mobile/app/client_context.dart';
import 'package:butter_mobile/core/client_storage.dart' show PendingOAuth;
import 'package:butter_mobile/gen/agents/v1/auth.pb.dart';
import 'package:flutter_test/flutter_test.dart';

import 'support/fake_oauth_platform.dart';
import 'support/memory_storage.dart';
import 'support/server.dart';

// RFC 7636 Appendix B: an independent, published verifier/challenge pair.
const rfcVerifier = 'dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk';
const rfcChallenge = 'E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM';
const returnUri = 'ee.orz.butter:/oauth/callback';

void main() {
  late TestServer server;
  late MemoryStorage storage;
  late FakeOAuthPlatform platform;
  late ClientContext context;
  var clock = DateTime.utc(2026, 10, 13, 8);

  ClientContext newContext({
    MemoryStorage? withStorage,
    Duration timeout = const Duration(minutes: 10),
  }) => ClientContext(
    storage: withStorage ?? storage,
    allowHttp: true,
    oauth: platform,
    newVerifier: () => rfcVerifier,
    now: () => clock,
    oauthTimeout: timeout,
  );

  setUp(() async {
    clock = DateTime.utc(2026, 10, 13, 8);
    server = TestServer()
      ..oauthProviders = [OAuthProvider(name: 'github', displayName: 'GitHub')];
    await server.start();
    storage = MemoryStorage();
    platform = FakeOAuthPlatform();
    context = newContext();
    await context.start();
    await context.setServer(server.url);
  });
  tearDown(() async {
    context.dispose();
    await server.close();
  });

  String link(String state, {String handoff = 'good-1'}) =>
      '$returnUri?handoff=$handoff&state=$state';

  test(
    'system browser sign-in: begin with PKCE, return link, redeem, session',
    () async {
      expect(context.oauthProviders.map((p) => p.displayName), ['GitHub']);

      await context.beginOAuth('github');

      // Public Begin: the registered return URI and the S256 challenge only.
      final begin = server.begins.single;
      expect(begin.provider, 'github');
      expect(begin.native.returnUri, returnUri);
      expect(begin.native.codeChallenge, rfcChallenge);
      expect(
        server.requests.where((r) => r.path.endsWith('/BeginOAuthFlow')),
        everyElement(
          predicate<({String path, String? token, String? workspace})>(
            (r) => r.token == null && r.workspace == null,
          ),
        ),
      );
      // The provider page opens in the system browser, and the App waits.
      expect(platform.opened.single.toString(), startsWith('https://provider'));
      expect(context.oauthWaitingFor, 'GitHub');
      expect(context.user, isNull);
      // The flow survives a process death: it is in secure storage, server-bound.
      expect(storage.pending!.server, server.url);
      expect(storage.pending!.provider, 'github');
      expect(storage.pending!.state, 'state-1');
      expect(storage.pending!.verifier, rfcVerifier);

      platform.deliver(link('state-1'));
      await context.oauthIdle;

      final redeem = server.redeems.single;
      expect(redeem.provider, 'github');
      expect(redeem.state, 'state-1');
      expect(redeem.handoffCode, 'good-1');
      expect(redeem.codeVerifier, rfcVerifier);
      expect(
        server.requests
            .where((r) => r.path.endsWith('/RedeemOAuthHandoff'))
            .every((r) => r.token == null && r.workspace == null),
        isTrue,
      );
      expect(context.user!.username, 'oauth-github');
      expect(context.stage, ClientStage.workspaces);
      expect(context.oauthWaitingFor, isNull);
      expect(context.error, isNull);
      expect(storage.credential!.token, 'oauth-github');
      expect(storage.credential!.server, server.url);
      expect(storage.pending, isNull);
    },
  );

  group('the browser leg can fail or be abandoned', () {
    test(
      'cancelling forgets the flow; a later valid link is ignored',
      () async {
        await context.beginOAuth('github');
        await context.cancelOAuth();
        expect(context.oauthWaitingFor, isNull);
        expect(storage.pending, isNull);

        platform.deliver(link('state-1'));
        await context.oauthIdle;
        expect(server.redeems, isEmpty);
        expect(context.user, isNull);
        expect(context.error, isNull);
      },
    );

    test('a browser leg that never returns times out', () async {
      context.dispose();
      context = newContext(timeout: const Duration(milliseconds: 30));
      await context.start();
      await context.beginOAuth('github');
      expect(context.oauthWaitingFor, 'GitHub');

      await Future<void>.delayed(const Duration(milliseconds: 120));
      expect(context.oauthWaitingFor, isNull);
      expect(storage.pending, isNull);
      expect(context.error, contains('timed out'));

      platform.deliver(link('state-1'));
      await context.oauthIdle;
      expect(server.redeems, isEmpty);
      expect(context.user, isNull);
    });

    test('a return link without a handoff ends the attempt', () async {
      await context.beginOAuth('github');
      platform.deliver('$returnUri?state=state-1&error=access_denied');
      await context.oauthIdle;
      expect(server.redeems, isEmpty);
      expect(context.oauthWaitingFor, isNull);
      expect(storage.pending, isNull);
      expect(context.error, contains('not completed'));
      expect(context.stage, ClientStage.login);
    });

    test('a refused Begin shows why and opens nothing', () async {
      server.beginFailure = (400, 'failed_precondition');
      await context.beginOAuth('github');
      expect(context.error, contains('not enabled'));
      expect(platform.opened, isEmpty);
      expect(storage.pending, isNull);
      expect(context.oauthWaitingFor, isNull);
      expect(context.stage, ClientStage.login);
    });

    test('a browser that cannot be opened ends the attempt', () async {
      platform.openResult = false;
      await context.beginOAuth('github');
      expect(context.error, contains('browser'));
      expect(storage.pending, isNull);
      expect(context.oauthWaitingFor, isNull);
    });

    test('only web URLs are ever opened in the browser', () async {
      for (final unsafe in [
        'javascript:alert(1)',
        'file:///etc/passwd',
        'ee.orz.butter:/oauth/callback?handoff=good-1&state=state-1',
      ]) {
        server.beginAuthorizeUrl = unsafe;
        await context.beginOAuth('github');
        expect(platform.opened, isEmpty, reason: unsafe);
        expect(storage.pending, isNull, reason: unsafe);
        expect(context.oauthWaitingFor, isNull, reason: unsafe);
        expect(context.error, isNotNull, reason: unsafe);
      }
    });

    test('a spent or unknown handoff signs nobody in', () async {
      await context.beginOAuth('github');
      platform.deliver(link('state-1', handoff: 'bad-1'));
      await context.oauthIdle;
      expect(context.user, isNull);
      expect(context.stage, ClientStage.login);
      expect(context.error, contains('could not be completed'));
      expect(storage.credential, isNull);
      expect(storage.pending, isNull);
      expect(context.oauthWaitingFor, isNull);
    });
  });

  group('return links are matched to the flow', () {
    test('a duplicate callback redeems exactly once', () async {
      final gate = Completer<void>();
      server.redeemReply = (input) async {
        await gate.future;
        return RedeemOAuthHandoffResponse(
          token: 'oauth-github',
          user: User(id: 'u', username: 'u'),
        );
      };
      await context.beginOAuth('github');
      platform.deliver(link('state-1'));
      platform.deliver(link('state-1')); // while the first is in flight
      gate.complete();
      await context.oauthIdle;
      platform.deliver(link('state-1')); // and after it finished
      await context.oauthIdle;
      expect(server.redeems, hasLength(1));
      expect(context.user!.username, 'oauth-github');
      expect(context.error, isNull);
    });

    test('a link for another state or path is ignored, not fatal', () async {
      await context.beginOAuth('github');
      platform.deliver(link('state-from-someone-else'));
      platform.deliver('ee.orz.butter:/other?handoff=good-1&state=state-1');
      await context.oauthIdle;
      expect(server.redeems, isEmpty);
      expect(context.oauthWaitingFor, 'GitHub');
      expect(storage.pending, isNotNull);

      platform.deliver(link('state-1'));
      await context.oauthIdle;
      expect(context.user!.username, 'oauth-github');
    });
  });

  group('the flow survives the App being rebuilt', () {
    test('backgrounding does not lose a flow in the browser', () async {
      await context.beginOAuth('github');
      context.suspend(); // the browser took the foreground
      platform.deliver(link('state-1'));
      await context.oauthIdle;
      expect(context.user!.username, 'oauth-github');
    });

    test('a new process finishes the flow from secure storage', () async {
      await context.beginOAuth('github');
      context.dispose(); // the OS killed the App while in the browser
      platform = FakeOAuthPlatform();
      context = newContext();
      await context.start();
      expect(context.oauthWaitingFor, 'GitHub');

      platform.deliver(link('state-1'));
      await context.oauthIdle;
      expect(server.redeems.single.codeVerifier, rfcVerifier);
      expect(context.user!.username, 'oauth-github');
      expect(storage.pending, isNull);
    });

    test('a link that launches the App cold waits for the restore', () async {
      await context.beginOAuth('github');
      context.dispose();
      platform = FakeOAuthPlatform();
      context = newContext();
      final starting = context.start();
      platform.deliver(link('state-1')); // before storage was read
      await starting;
      await context.oauthIdle;
      expect(context.user!.username, 'oauth-github');
    });

    test('an expired flow is forgotten on restart', () async {
      await context.beginOAuth('github');
      context.dispose();
      clock = clock.add(const Duration(minutes: 11));
      platform = FakeOAuthPlatform();
      context = newContext();
      await context.start();
      expect(context.oauthWaitingFor, isNull);
      expect(storage.pending, isNull);

      platform.deliver(link('state-1'));
      await context.oauthIdle;
      expect(server.redeems, isEmpty);
    });

    test('a flow begun against another server is never redeemed', () async {
      await context.beginOAuth('github');
      context.dispose();
      storage.pending = PendingOAuth(
        server: 'https://elsewhere.example',
        provider: storage.pending!.provider,
        state: storage.pending!.state,
        verifier: storage.pending!.verifier,
        expiresAt: storage.pending!.expiresAt,
      );
      platform = FakeOAuthPlatform();
      context = newContext();
      await context.start();
      expect(context.oauthWaitingFor, isNull);
      expect(storage.pending, isNull);

      platform.deliver(link('state-1'));
      await context.oauthIdle;
      expect(server.redeems, isEmpty);
    });
  });

  group('contexts are not mixed', () {
    test('changing the server abandons the flow', () async {
      await context.beginOAuth('github');
      await context.setServer(server.url);
      expect(context.oauthWaitingFor, isNull);
      expect(storage.pending, isNull);

      platform.deliver(link('state-1'));
      await context.oauthIdle;
      expect(server.redeems, isEmpty);
    });

    test('a late redemption cannot override a newer password login', () async {
      final gate = Completer<void>();
      server.redeemReply = (input) async {
        await gate.future;
        return RedeemOAuthHandoffResponse(
          token: 'oauth-github',
          user: User(id: 'u', username: 'u'),
        );
      };
      await context.beginOAuth('github');
      platform.deliver(link('state-1'));
      await context.login('bob', 'password');
      gate.complete();
      await context.oauthIdle;
      expect(context.user!.username, 'bob');
      expect(storage.credential!.token, 'bob');
    });

    test(
      'starting a password login drops a flow still in the browser',
      () async {
        await context.beginOAuth('github');
        await context.login('bob', 'password');
        expect(context.oauthWaitingFor, isNull);
        expect(storage.pending, isNull);
        platform.deliver(link('state-1'));
        await context.oauthIdle;
        expect(server.redeems, isEmpty);
        expect(context.user!.username, 'bob');
      },
    );
  });
}
