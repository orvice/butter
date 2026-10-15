import 'dart:async';
import 'dart:io';

import 'package:butter_mobile/app/client_context.dart';
import 'package:butter_mobile/features/butterboxes/box_controller.dart';
import 'package:butter_mobile/gen/agents/v1/butterbox.pb.dart';
import 'package:flutter_test/flutter_test.dart';

import 'support/box_server.dart';
import 'support/memory_storage.dart';

Future<({ClientContext client, BoxServer server})> start({
  void Function(BoxServer)? seed,
}) async {
  final old = HttpOverrides.current;
  HttpOverrides.global = null;
  final server = BoxServer();
  seed?.call(server);
  await server.start();
  final client = ClientContext(storage: MemoryStorage(), allowHttp: true);
  addTearDown(() async {
    client.dispose();
    await server.close();
    HttpOverrides.global = old;
  });
  await client.setServer(server.url);
  await client.login('alice', 'password');
  await client.selectWorkspace('alpha');
  return (client: client, server: server);
}

void main() {
  test('lists and reads boxes of the active Workspace only', () async {
    final h = await start(
      seed: (s) {
        s.seed('alpha', 'a1', 'Alpha box');
        s.seed('beta', 'b1', 'Beta box');
      },
    );
    final boxes = h.client.boxes!;
    await boxes.loadBoxes();
    expect(boxes.listState, BoxLoad.ready);
    expect(boxes.boxes.map((b) => b.name), ['Alpha box']);
    expect(boxes.boxes.single.credentialSet, isTrue);
    await boxes.loadBox('a1');
    expect(boxes.box?.name, 'Alpha box');
    expect(h.server.boxCalls.map((c) => c.workspace).toSet(), {'alpha'});
  });

  test('registering sends the token once; edits and reads never do', () async {
    final h = await start();
    final boxes = h.client.boxes!;
    expect(
      await boxes.create(
        name: ' Laptop ',
        baseUrl: ' https://box.test ',
        enabled: true,
        accessToken: 'tok-1',
      ),
      isTrue,
    );
    final created = h.server.boxes['alpha']!.values.single;
    expect(created.name, 'Laptop');
    expect(created.baseUrl, 'https://box.test');
    expect(created.credentialSet, isTrue);
    expect(h.server.secrets['alpha']![created.id], 'tok-1');

    await boxes.update(
      id: created.id,
      name: 'Laptop 2',
      baseUrl: 'https://box.test',
      enabled: false,
    );
    expect(h.server.secrets['alpha']![created.id], 'tok-1');
    expect(h.server.boxes['alpha']![created.id]!.enabled, isFalse);
    // The token string is on the wire exactly once: in the create request.
    final carrying = h.server.boxCalls.where(
      (c) => String.fromCharCodes(c.body).contains('tok-1'),
    );
    expect(carrying.map((c) => c.method), ['CreateButterBox']);
  });

  test('keep, replace and clear are distinct explicit writes', () async {
    final h = await start(seed: (s) => s.seed('alpha', 'a1', 'One'));
    final boxes = h.client.boxes!;
    await boxes.loadBoxes();

    // A blank replacement is refused locally: it can never erase a token.
    expect(await boxes.replaceToken('a1', '  '), isFalse);
    expect(h.server.boxWrites, 0);
    expect(h.server.secrets['alpha']!['a1'], 'secret-of-a1');

    expect(await boxes.replaceToken('a1', ' rotated '), isTrue);
    expect(h.server.secrets['alpha']!['a1'], 'rotated');
    expect(boxes.byId('a1')!.credentialSet, isTrue);

    expect(await boxes.clearToken('a1'), isTrue);
    expect(h.server.secrets['alpha']!.containsKey('a1'), isFalse);
    expect(boxes.byId('a1')!.credentialSet, isFalse);
  });

  test('delete blocked by Agent references names them and keeps the '
      'box', () async {
    final h = await start(
      seed: (s) {
        s.seed('alpha', 'a1', 'One');
        s.referencedBy['a1'] = ['coder', 'planner'];
      },
    );
    final boxes = h.client.boxes!;
    await boxes.loadBoxes();
    expect(await boxes.delete('a1'), isFalse);
    expect(boxes.writeError, contains('coder'));
    expect(boxes.writeError, contains('planner'));
    expect(boxes.writeError, contains('Agents'));
    expect(boxes.unconfirmed, isFalse);
    expect(boxes.byId('a1'), isNotNull);
    expect(h.server.boxes['alpha']!.containsKey('a1'), isTrue);

    h.server.referencedBy.clear();
    expect(await boxes.delete('a1'), isTrue);
    expect(boxes.byId('a1'), isNull);
  });

  test('a lost write reply is unconfirmed, never repeated, and cleared by '
      'a read', () async {
    final h = await start();
    final boxes = h.client.boxes!;
    h.server.loseBoxReply = true;
    expect(
      await boxes.create(
        name: 'One',
        baseUrl: 'https://box.test',
        enabled: true,
        accessToken: 'tok',
      ),
      isFalse,
    );
    expect(boxes.unconfirmed, isTrue);
    expect(boxes.writeNotice, contains('not confirmed'));
    expect(h.server.boxWrites, 1);

    // Another write is refused until a read reconciles; nothing is resent.
    h.server.loseBoxReply = false;
    expect(
      await boxes.create(
        name: 'One',
        baseUrl: 'https://box.test',
        enabled: true,
        accessToken: 'tok',
      ),
      isFalse,
    );
    expect(h.server.boxWrites, 1);

    await boxes.loadBoxes();
    expect(boxes.unconfirmed, isFalse);
    // The write had been applied: the authoritative read shows it.
    expect(boxes.boxes.map((b) => b.name), ['One']);
  });

  test('server refusals map to actionable messages', () async {
    final h = await start(seed: (s) => s.seed('alpha', 'a1', 'One'));
    final boxes = h.client.boxes!;
    await boxes.loadBoxes();
    expect(
      await boxes.create(name: 'One', baseUrl: 'https://x.test', enabled: true),
      isFalse,
    );
    expect(boxes.writeError, isNotNull);
    expect(boxes.unconfirmed, isFalse);

    h.server.boxRefusal = (
      400,
      'invalid_argument',
      'butter box base URL must be http or https',
    );
    expect(
      await boxes.update(
        id: 'a1',
        name: 'One',
        baseUrl: 'ftp://x',
        enabled: true,
      ),
      isFalse,
    );
    expect(boxes.writeError, contains('http'));
  });

  test('an unreachable box is an answer, not a failure', () async {
    final h = await start(seed: (s) => s.seed('alpha', 'a1', 'One'));
    final boxes = h.client.boxes!;
    h.server.status = GetButterBoxStatusResponse(
      reachable: false,
      error: 'dial tcp: connection refused',
    );
    await boxes.probe('a1');
    expect(boxes.probes['a1']!.state, BoxProbeState.unreachable);
    expect(boxes.probes['a1']!.detail, contains('connection refused'));

    h.server.status = GetButterBoxStatusResponse(
      reachable: true,
      activeSessions: 3,
    );
    await boxes.probe('a1');
    expect(boxes.probes['a1']!.state, BoxProbeState.reachable);
    expect(boxes.probes['a1']!.activeSessions, 3);
    expect(h.server.boxWrites, 0);
  });

  test('catalog failures say what to do and keep manual entry open', () async {
    final h = await start(seed: (s) => s.seed('alpha', 'a1', 'One'));
    final boxes = h.client.boxes!;

    h.server.piRefusal = (429, 'resource_exhausted', 'pi box is at capacity');
    await boxes.loadPiModels('a1');
    expect(boxes.piCatalogs['a1']!.state, BoxLoad.failed);
    expect(boxes.piCatalogs['a1']!.error, contains('capacity'));
    expect(boxes.piCatalogs['a1']!.error, contains('by hand'));

    h.server.cursorRefusal = (
      400,
      'failed_precondition',
      'cursor catalog: CURSOR_API_KEY is missing or invalid on the box',
    );
    await boxes.loadCursorModels('a1');
    expect(boxes.cursorCatalogs['a1']!.state, BoxLoad.failed);
    expect(boxes.cursorCatalogs['a1']!.error, contains('CURSOR_API_KEY'));

    h.server.piRefusal = null;
    h.server.piModels = [
      ButterBoxModel(
        id: 'sonnet',
        provider: 'anthropic',
        name: 'Sonnet',
        reasoning: true,
      ),
    ];
    await boxes.loadPiModels('a1');
    expect(boxes.piCatalogs['a1']!.state, BoxLoad.ready);
    expect(boxes.piCatalogs['a1']!.models.single.id, 'sonnet');
  });

  test('a late answer never reaches another Workspace', () async {
    final gate = Completer<void>();
    final h = await start(
      seed: (s) {
        s.seed('alpha', 'a1', 'Alpha box');
        s.seed('beta', 'b1', 'Beta box');
      },
    );
    h.server.beforeList = () => gate.future;
    final old = h.client.boxes!;
    final pending = old.loadBoxes();
    await Future<void>.delayed(const Duration(milliseconds: 50));
    h.server.beforeList = null;
    await h.client.selectWorkspace('beta');
    expect(old.disposed, isTrue);
    final fresh = h.client.boxes!;
    expect(identical(fresh, old), isFalse);
    expect(fresh.boxes, isEmpty);
    gate.complete();
    await pending;
    expect(old.boxes, isEmpty);
    await fresh.loadBoxes();
    expect(fresh.boxes.map((b) => b.name), ['Beta box']);
  });

  test('sign-out drops every box and catalog', () async {
    final h = await start(seed: (s) => s.seed('alpha', 'a1', 'One'));
    final boxes = h.client.boxes!;
    await boxes.loadBoxes();
    await boxes.probe('a1');
    await h.client.logout();
    expect(boxes.disposed, isTrue);
    expect(boxes.boxes, isEmpty);
    expect(boxes.probes, isEmpty);
    expect(h.client.boxes, isNull);
  });

  test(
    'an expired session signs the person out and nothing is shown',
    () async {
      final h = await start(seed: (s) => s.seed('alpha', 'a1', 'One'));
      final boxes = h.client.boxes!;
      h.server.expired = true;
      await boxes.loadBoxes();
      await Future<void>.delayed(const Duration(milliseconds: 50));
      expect(h.client.error, contains('Session expired'));
      expect(h.client.boxes, isNull);
      expect(boxes.boxes, isEmpty);
    },
  );

  test('suspending fences an in-flight probe so its late answer is '
      'dropped', () async {
    final gate = Completer<void>();
    final h = await start(seed: (s) => s.seed('alpha', 'a1', 'One'));
    final boxes = h.client.boxes!;
    h.server.beforeStatus = () => gate.future;
    final probe = boxes.probe('a1');
    await Future<void>.delayed(const Duration(milliseconds: 50));
    boxes.suspend();
    gate.complete();
    await probe;
    expect(boxes.probes.containsKey('a1'), isFalse);
  });
}
