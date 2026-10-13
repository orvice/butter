import 'dart:async';

import 'package:butter_mobile/gen/agents/v1/agent.pb.dart';
import 'package:butter_mobile/gen/agents/v1/agent_service.pb.dart';
import 'package:butter_mobile/gen/agents/v1/auth.pb.dart';
import 'package:butter_mobile/app/client_context.dart';
import 'package:butter_mobile/core/client_storage.dart';
import 'package:flutter_test/flutter_test.dart';

import 'support/server.dart';

class MemoryStorage implements ClientStorage {
  String? server;
  StoredSession? credential;
  bool failSave = false;
  Future<void> Function()? beforeSave;
  Future<StoredSession?> Function()? readSessionOverride;
  bool failClear = false;
  @override
  Future<String?> readServer() async => server;
  @override
  Future<void> saveServer(String address) async {
    server = address;
  }

  @override
  Future<StoredSession?> readSession() async =>
      readSessionOverride == null ? credential : await readSessionOverride!();
  @override
  Future<void> saveSession(StoredSession session) async {
    if (failSave) throw StateError('OS vault is locked');
    await beforeSave?.call();
    credential = session;
  }

  @override
  Future<void> clearSession() async {
    if (failClear) throw StateError('OS vault is locked');
    credential = null;
  }
}

void main() {
  late TestServer server;
  late MemoryStorage storage;
  late ClientContext context;
  setUp(() async {
    server = TestServer();
    await server.start();
    storage = MemoryStorage();
    context = ClientContext(storage: storage, allowHttp: true);
  });
  tearDown(() async {
    context.dispose();
    await server.close();
  });

  test(
    'verify, password login, Me and Workspace selection use a real HTTP boundary',
    () async {
      await context.start();
      expect(context.stage, ClientStage.server);
      await context.setServer(server.url);
      expect(context.stage, ClientStage.login);
      expect(storage.server, server.url);
      await context.login('alice', 'password');
      expect(context.user!.username, 'alice');
      expect(context.workspaces.map((w) => w.id), ['alpha', 'beta']);
      expect(context.stage, ClientStage.workspaces);
      await context.selectWorkspace('alpha');
      expect(context.agents.single.name, 'alpha agent');
      expect(storage.credential!.server, server.url);
      final public = server.requests.where(
        (r) =>
            r.path == '/ping' ||
            r.path.endsWith('/Login') ||
            r.path.endsWith('/ListOAuthProviders'),
      );
      expect(
        public.every((r) => r.token == null && r.workspace == null),
        isTrue,
      );
      expect(
        server.requests
            .where(
              (r) =>
                  r.path.endsWith('/Me') || r.path.endsWith('/ListWorkspaces'),
            )
            .every((r) => r.workspace == null),
        isTrue,
      );
    },
  );

  test(
    'corrupt-record cleanup during an old restore cannot delete a newer login',
    () async {
      storage.server = server.url;
      final reading = Completer<void>();
      final release = Completer<void>();
      storage.readSessionOverride = () async {
        reading.complete();
        await release.future;
        // The OS adapter deletes a corrupt record before returning null.
        storage.credential = null;
        return null;
      };
      final oldRestore = context.start();
      await reading.future;
      final newLogin = context.login('bob', 'password');
      await Future.any([
        newLogin,
        Future<void>.delayed(const Duration(milliseconds: 100)),
      ]);
      release.complete();
      await Future.wait([oldRestore, newLogin]);
      expect(context.user!.username, 'bob');
      expect(storage.credential!.token, 'bob');
    },
  );

  test(
    'failed cleanup remains visible after a post-login request fails',
    () async {
      await context.setServer(server.url);
      server.denyWorkspaces = true;
      // Allow the initial delete, but make OS deletion fail after Login returns.
      server.loginReply = (_) async {
        storage.failClear = true;
        return LoginResponse(token: 'alice');
      };
      await context.login('alice', 'password');
      expect(context.stage, ClientStage.login);
      expect(context.error, contains('could not be cleared'));
      expect(context.error, contains('permission'));
      expect(storage.credential!.token, 'alice');
    },
  );

  test(
    'backgrounding closes pending Logout observation without repeating revocation',
    () async {
      await context.setServer(server.url);
      await context.login('alice', 'password');
      final arrived = Completer<void>();
      final release = Completer<void>();
      server.beforeLogout = () async {
        arrived.complete();
        await release.future;
      };
      final logout = context.logout();
      await arrived.future;
      context.suspend();
      try {
        final cancelled = await Future.any([
          logout.then((_) => true),
          Future<bool>.delayed(const Duration(milliseconds: 100), () => false),
        ]);
        expect(cancelled, isTrue);
        expect(context.error, contains('revocation'));
        expect(storage.credential, isNull);
      } finally {
        release.complete();
        await logout;
      }
      expect(
        server.requests.where((r) => r.path.endsWith('/Logout')),
        hasLength(1),
      );
    },
  );

  test('secure storage failure cannot leave an authenticated UI', () async {
    await context.setServer(server.url);
    storage.failSave = true;
    await context.login('alice', 'password');
    expect(context.stage, ClientStage.login);
    expect(context.user, isNull);
    expect(storage.credential, isNull);
    expect(context.error, contains('secure storage'));
  });

  test(
    'slow OS writes cannot resurrect an old account after logout and relogin',
    () async {
      await context.setServer(server.url);
      final saving = Completer<void>();
      final release = Completer<void>();
      storage.beforeSave = () async {
        saving.complete();
        await release.future;
      };
      final oldLogin = context.login('alice', 'password');
      await saving.future;
      final logout = context.logout();
      final newLogin = context.login('bob', 'password');
      storage.beforeSave = null;
      release.complete();
      await Future.wait([oldLogin, logout, newLogin]);
      expect(context.user!.username, 'bob');
      expect(storage.credential!.token, 'bob');
      expect(context.error, isNull);
    },
  );

  test(
    'foreground restoration reads identity without repeating Login or writes',
    () async {
      await context.setServer(server.url);
      await context.login('alice', 'password');
      await context.selectWorkspace('alpha');
      context.suspend();
      await context.revalidate();
      expect(context.agents.single.name, 'alpha agent');
      expect(
        server.requests.where((r) => r.path.endsWith('/Login')),
        hasLength(1),
      );
      expect(server.requests.any((r) => r.path.endsWith('/Logout')), isFalse);
      server.expired = true;
      context.suspend();
      await context.revalidate();
      expect(context.stage, ClientStage.login);
      expect(storage.credential, isNull);
    },
  );

  test(
    'binary protobuf optional presence and unknown fields survive the public list API',
    () async {
      await context.setServer(server.url);
      await context.login('alice', 'password');
      final known = Agent(
        agentId: 'future',
        config: AgentConfig(pi: PiAgentConfig(maxRunSeconds: 0)),
      );
      // Independent literal protobuf wire: field 1000, varint 42.
      final bytes = [...known.writeToBuffer(), 0xc0, 0x3e, 0x2a];
      server.agentReply = (_, _) async =>
          ListAgentsResponse(agents: [Agent.fromBuffer(bytes)]);
      await context.selectWorkspace('alpha');
      expect(context.agents.single.config.pi.hasMaxRunSeconds(), isTrue);
      expect(context.agents.single.config.pi.maxRunSeconds, 0);
      expect(context.agents.single.writeToBuffer(), bytes);
      expect(PiAgentConfig().hasMaxRunSeconds(), isFalse);
    },
  );

  test(
    'switching Workspace discards an in-flight response for the same Agent ID',
    () async {
      await context.setServer(server.url);
      await context.login('alice', 'password');
      final arrived = Completer<void>();
      final release = Completer<void>();
      server.agentReply = (workspace, _) async {
        if (workspace == 'alpha') {
          arrived.complete();
          await release.future;
        }
        return ListAgentsResponse(
          agents: [Agent(agentId: 'shared', name: workspace)],
        );
      };
      final first = context.selectWorkspace('alpha');
      await arrived.future;
      await context.selectWorkspace('beta');
      release.complete();
      await first;
      expect(context.workspace!.id, 'beta');
      expect(context.agents.single.name, 'beta');
      expect(context.error, isNull);
    },
  );

  test(
    'a late login cannot restore credentials after logout or another login',
    () async {
      await context.setServer(server.url);
      final arrived = Completer<void>();
      final release = Completer<void>();
      server.loginReply = (request) async {
        if (request.username == 'alice') {
          arrived.complete();
          await release.future;
        }
        return LoginResponse(token: request.username);
      };
      final old = context.login('alice', 'password');
      await arrived.future;
      await context.logout();
      await context.login('bob', 'password');
      release.complete();
      await old;
      expect(context.user!.username, 'bob');
      expect(storage.credential!.token, 'bob');
      expect(context.workspaces, hasLength(2));
    },
  );

  test(
    'changing the server clears the account before credential-free verification',
    () async {
      await context.setServer(server.url);
      await context.login('alice', 'password');
      await context.selectWorkspace('alpha');
      final next = TestServer();
      await next.start();
      addTearDown(next.close);
      await context.setServer(next.url);
      expect(context.user, isNull);
      expect(context.agents, isEmpty);
      expect(storage.credential, isNull);
      expect(storage.server, next.url);
      expect(next.requests.every((r) => r.token == null), isTrue);
      expect(server.requests.any((r) => r.path.contains('/stop')), isFalse);
    },
  );

  test('restart restores only a server-matched secure session', () async {
    await context.setServer(server.url);
    await context.login('alice', 'password');
    final restored = ClientContext(storage: storage, allowHttp: true);
    addTearDown(restored.dispose);
    await restored.start();
    expect(restored.user!.username, 'alice');
    expect(restored.workspace, isNull);
    storage.credential = const StoredSession(
      server: 'https://other.example',
      token: 'secret',
    );
    final isolated = ClientContext(storage: storage, allowHttp: true);
    addTearDown(isolated.dispose);
    await isolated.start();
    expect(isolated.stage, ClientStage.login);
    expect(storage.credential, isNull);
    expect(server.requests.any((r) => r.token == 'Bearer secret'), isFalse);
  });

  test(
    'expired session clears identity, agents and secure credentials; 403 does not',
    () async {
      await context.setServer(server.url);
      await context.login('alice', 'password');
      server.denyAgents = true;
      await context.selectWorkspace('alpha');
      expect(context.user!.username, 'alice');
      expect(storage.credential, isNotNull);
      expect(context.error, contains('permission'));
      server.denyAgents = false;
      await context.refreshAgents();
      server.expired = true;
      await context.refreshAgents();
      expect(context.stage, ClientStage.login);
      expect(context.agents, isEmpty);
      expect(context.workspaces, isEmpty);
      expect(storage.credential, isNull);
      expect(context.error, contains('expired'));
    },
  );

  test(
    'empty and paginated agents are preserved and writes are not retried',
    () async {
      await context.setServer(server.url);
      await context.login('alice', 'password');
      server.agentReply = (_, input) async => input.pageToken.isEmpty
          ? ListAgentsResponse(
              nextPageToken: 'next',
              agents: [Agent(agentId: 'one')],
            )
          : ListAgentsResponse(agents: [Agent(agentId: 'two')]);
      await context.selectWorkspace('alpha');
      await context.moreAgents();
      expect(context.agents.map((a) => a.agentId), ['one', 'two']);
      expect(context.hasMoreAgents, isFalse);
      server.agentReply = (_, _) async => ListAgentsResponse();
      await context.refreshAgents();
      expect(context.agents, isEmpty);
      server.failLogout = true;
      await context.logout();
      expect(storage.credential, isNull);
      expect(context.stage, ClientStage.login);
      expect(context.error, contains('revocation'));
      expect(
        server.requests.where((r) => r.path.endsWith('/Logout')),
        hasLength(1),
      );
    },
  );
}
