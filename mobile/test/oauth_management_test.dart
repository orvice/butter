import 'dart:convert';
import 'dart:io';

import 'package:butter_mobile/app/app_preferences.dart';
import 'package:butter_mobile/app/client_context.dart';
import 'package:butter_mobile/core/client_storage.dart';
import 'package:butter_mobile/gen/agents/v1/auth.pb.dart';
import 'package:flutter/material.dart';
import 'package:flutter_secure_storage/flutter_secure_storage.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:shared_preferences_platform_interface/shared_preferences_async_platform_interface.dart';

import 'device_storage_test.dart' show PreferencesPlatform;
import 'support/fake_oauth_platform.dart';
import 'support/management_server.dart';
import 'support/memory_storage.dart';

const _returnUri = 'ee.orz.butter:/oauth/callback';

/// Native OAuth sign-in (#448) and Workspace/self-service management (#449)
/// share one ClientContext lifetime; these cover where they meet.
void main() {
  TestWidgetsFlutterBinding.ensureInitialized();
  // The binding blocks real sockets; this test talks to a loopback fixture.
  HttpOverrides.global = null;
  late ManagementServer server;
  late FakeOAuthPlatform platform;

  setUp(() async {
    server = ManagementServer()
      ..oauthProviders = [OAuthProvider(name: 'github', displayName: 'GitHub')];
    await server.start();
    platform = FakeOAuthPlatform();
  });
  tearDown(() => server.close());

  ClientContext newClient(ClientStorage storage) => ClientContext(
    storage: storage,
    allowHttp: true,
    oauth: platform,
    newVerifier: () => 'dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk',
  );

  Future<void> signInWithOAuth(ClientContext client) async {
    await client.start();
    await client.setServer(server.url);
    await client.beginOAuth('github');
    platform.deliver('$_returnUri?handoff=good-1&state=state-1');
    await client.oauthIdle;
  }

  group('after native OAuth sign-in', () {
    late MemoryStorage storage;
    late ClientContext client;
    setUp(() async {
      storage = MemoryStorage();
      client = newClient(storage);
      addTearDown(client.dispose);
      await signInWithOAuth(client);
      expect(client.stage, ClientStage.workspaces);
    });

    test(
      'profile and Workspace management use the OAuth session; logout closes them',
      () async {
        expect(storage.credential!.token, 'oauth-github');
        expect(await client.saveProfile('Via OAuth'), isTrue);
        expect(client.user!.displayName, 'Via OAuth');
        await client.openWorkspace('alpha');
        expect(client.canManageWorkspace, isTrue);
        expect(
          await client.saveWorkspace(
            client.managedWorkspace!.deepCopy()..name = 'Renamed',
          ),
          isTrue,
        );
        expect(client.workspaces.first.name, 'Renamed');
        // Every authorized management call carried the OAuth session token and
        // none of the public OAuth calls carried a credential.
        final authorized = server.requests.where(
          (r) =>
              r.path.endsWith('/UpdateProfile') ||
              r.path.endsWith('/GetWorkspace') ||
              r.path.endsWith('/UpdateWorkspace'),
        );
        expect(authorized, isNotEmpty);
        expect(authorized.map((r) => r.token).toSet(), {'Bearer oauth-github'});
        expect(
          server.requests
              .where(
                (r) =>
                    r.path.endsWith('/BeginOAuthFlow') ||
                    r.path.endsWith('/RedeemOAuthHandoff'),
              )
              .every((r) => r.token == null),
          isTrue,
        );

        await client.logout();
        expect(client.stage, ClientStage.login);
        expect(client.user, isNull);
        expect(client.managedWorkspace, isNull);
        expect(client.members, isEmpty);
        expect(client.workspaces, isEmpty);
        expect(storage.credential, isNull);
        expect(storage.pending, isNull);
        expect(
          server.requests.where((r) => r.path.endsWith('/Logout')).single.token,
          'Bearer oauth-github',
        );
        // Management is gone with the session: no write can be issued.
        final writes = server.writes;
        expect(await client.saveProfile('After logout'), isFalse);
        expect(await client.saveWorkspace(server.spaces.first), isFalse);
        expect(server.writes, writes);
      },
    );

    test(
      'a duplicate or late return link after logout never signs in again',
      () async {
        await client.logout();
        platform.deliver('$_returnUri?handoff=good-1&state=state-1');
        await client.oauthIdle;
        expect(server.redeems, hasLength(1));
        expect(client.user, isNull);
        expect(client.stage, ClientStage.login);
        expect(storage.credential, isNull);
      },
    );

    test(
      'an expired OAuth session during a write clears management and signs out',
      () async {
        await client.openWorkspace('alpha');
        expect(client.members, isNotEmpty);
        server.expired = true;
        expect(await client.saveProfile('Expired'), isFalse);
        expect(client.user, isNull);
        expect(client.managedWorkspace, isNull);
        expect(client.members, isEmpty);
        expect(client.error, contains('Session expired'));
        expect(storage.credential, isNull);
        expect(storage.pending, isNull);
        // The expired session can be replaced by a new OAuth sign-in, which
        // starts with no management state of the previous one.
        server.expired = false;
        await client.beginOAuth('github');
        platform.deliver('$_returnUri?handoff=good-2&state=state-2');
        await client.oauthIdle;
        expect(client.stage, ClientStage.workspaces);
        expect(client.managedWorkspace, isNull);
        expect(client.members, isEmpty);
        expect(storage.credential!.token, 'oauth-github');
      },
    );

    test(
      'a new sign-in flow closes the old Workspace management scope',
      () async {
        await client.openWorkspace('alpha');
        final scope = client.scope;
        await client.logout();
        expect(client.scope, greaterThan(scope));
        await client.beginOAuth('github');
        expect(client.managedWorkspace, isNull);
        expect(client.oauthWaitingFor, 'GitHub');
        expect(client.user, isNull);
      },
    );
  });

  group('clearing credentials never touches device preferences', () {
    late PreferencesPlatform preferences;
    late DeviceStorage storage;
    late ClientContext client;
    late AppPreferences appearance;

    Map<String, dynamic> stored() =>
        jsonDecode(preferences.values['butter.appearance']!)
            as Map<String, dynamic>;

    setUp(() async {
      preferences = PreferencesPlatform();
      SharedPreferencesAsyncPlatform.instance = preferences;
      FlutterSecureStorage.setMockInitialValues({});
      storage = DeviceStorage();
      client = newClient(storage);
      appearance = AppPreferences();
      addTearDown(() {
        appearance.dispose();
        client.dispose();
      });
      await appearance.start();
      expect(
        await appearance.save(theme: ThemeMode.dark, showIdentifiers: true),
        isTrue,
      );
    });

    test(
      'logout and server replacement keep appearance, drop OAuth state',
      () async {
        await signInWithOAuth(client);
        expect(await storage.readSession(), isNotNull);
        await client.logout();
        expect(await storage.readSession(), isNull);
        expect(await storage.readPendingOAuth(), isNull);
        expect(stored(), {'theme': 'dark', 'showIdentifiers': true});
        expect(await storage.readServer(), server.url);

        // A pending flow is removed when the server is replaced, and so is the
        // session, while the appearance record stays exactly as saved.
        await client.beginOAuth('github');
        expect(await storage.readPendingOAuth(), isNotNull);
        await client.setServer(server.url);
        expect(await storage.readPendingOAuth(), isNull);
        expect(await storage.readSession(), isNull);
        expect(stored(), {'theme': 'dark', 'showIdentifiers': true});
      },
    );

    test('the PKCE verifier lives only in secure storage', () async {
      await client.start();
      await client.setServer(server.url);
      await client.beginOAuth('github');
      expect(
        (await storage.readPendingOAuth())!.verifier,
        'dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk',
      );
      expect(preferences.values.keys.toSet(), {
        'butter.server',
        'butter.appearance',
      });
      expect(preferences.values.values.join(), isNot(contains('dBjftJeZ4CVP')));
    });

    test('saving or failing to save appearance keeps pending OAuth and session '
        'intact', () async {
      await client.start();
      await client.setServer(server.url);
      await client.beginOAuth('github');
      final pending = (await storage.readPendingOAuth())!;

      preferences.failSave = true;
      expect(
        await appearance.save(theme: ThemeMode.light, showIdentifiers: false),
        isFalse,
      );
      preferences.failSave = false;
      expect(
        await appearance.save(theme: ThemeMode.light, showIdentifiers: false),
        isTrue,
      );

      expect((await storage.readPendingOAuth())!.state, pending.state);
      expect(client.oauthWaitingFor, 'GitHub');
      platform.deliver('$_returnUri?handoff=good-1&state=${pending.state}');
      await client.oauthIdle;
      expect(client.stage, ClientStage.workspaces);
      expect((await storage.readSession())!.token, 'oauth-github');
      expect(await storage.readPendingOAuth(), isNull);
      expect(stored(), {'theme': 'light', 'showIdentifiers': false});
    });
  });
}
