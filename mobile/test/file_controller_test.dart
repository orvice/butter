import 'dart:async';
import 'dart:io';

import 'package:butter_mobile/app/client_context.dart';
import 'package:butter_mobile/core/config_resource_controller.dart';
import 'package:butter_mobile/features/files/file_controller.dart';
import 'package:butter_mobile/features/files/file_rules.dart';
import 'package:butter_mobile/gen/agents/v1/agent.pb.dart';
import 'package:butter_mobile/gen/agents/v1/agent_file.pb.dart';
import 'package:flutter_test/flutter_test.dart';

import 'support/file_server.dart';
import 'support/memory_storage.dart';

Future<({ClientContext client, FileServerFake server})> start({
  void Function(FileServerFake)? seed,
}) async {
  final old = HttpOverrides.current;
  HttpOverrides.global = null;
  final server = FileServerFake();
  server.seedSpace('alpha', 'docs', 'Docs');
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

Future<void> settle() => Future<void>.delayed(const Duration(milliseconds: 50));

void main() {
  group('reading', () {
    test(
      'lists only the Workspace\'s spaces and the open space\'s files',
      () async {
        final h = await start(
          seed: (s) {
            s.seedSpace('beta', 'other', 'Other');
            s.seedFile('alpha', 'docs', '/notes/a.md', ['one']);
            s.seedFile('alpha', 'docs', '/b.md', ['bee', 'bee 2']);
          },
        );
        final c = h.client.fileSpaces!;
        await c.load();
        expect(c.items.map((s) => s.id), ['docs']);
        await c.openSpace('docs');
        expect(c.files.map((f) => f.path), ['/b.md', '/notes/a.md']);
        expect(c.files.first.version.toInt(), 2);
        await c.loadFiles(prefix: 'notes');
        expect(c.files.map((f) => f.path), ['/notes/a.md']);
        expect(h.server.fileCalls.map((c) => c.workspace).toSet(), {'alpha'});
      },
    );

    test(
      'an older version is read by number, and a missing one is said so',
      () async {
        final h = await start(
          seed: (s) => s.seedFile('alpha', 'docs', '/a.md', ['v1', 'v2', 'v3']),
        );
        final c = h.client.fileSpaces!;
        final old = await c.readFile('docs', '/a.md', version: 1, latest: 3);
        expect(old.document!.content, 'v1');
        expect(old.document!.isLatest, isFalse);
        expect(old.document!.latest, 3);
        final latest = await c.readFile('docs', '/a.md');
        expect(latest.document!.version, 3);
        expect(latest.document!.isLatest, isTrue);
        final gone = await c.readFile('docs', '/a.md', version: 9, latest: 3);
        expect(gone.notFound, isTrue);
        expect(gone.error, contains('not available'));
      },
    );

    test('search needs a query and finds contents', () async {
      final h = await start(
        seed: (s) => s.seedFile('alpha', 'docs', '/a.md', ['a needle here']),
      );
      final c = h.client.fileSpaces!;
      expect((await c.search('docs', '  ')).error, contains('Enter text'));
      expect(h.server.fileCount('SearchAgentFiles'), 0);
      final found = await c.search('docs', 'needle');
      expect(found.results!.single.file.path, '/a.md');
    });

    test(
      'failures are worded for the person, never with server text',
      () async {
        final h = await start();
        final c = h.client.fileSpaces!;
        h.server.fileRefusals['ListAgentFileSpaces'] = (
          400 + 12,
          'failed_precondition',
          'agent file repository not available',
        );
        await c.load();
        expect(c.listError, contains('File storage is not set up'));
        h.server.fileRefusals['ListAgentFileSpaces'] = (
          403,
          'permission_denied',
          'secret-detail',
        );
        await c.load();
        expect(c.listError, contains('do not have access'));
        expect(c.listError, isNot(contains('secret-detail')));
        h.server.fileRefusals.clear();
        h.server.filesUnimplemented = true;
        await c.load();
        expect(c.listError, contains('does not support file spaces'));
      },
    );

    test('a Workspace the account cannot reach is refused', () async {
      final h = await start(seed: (s) => s.stores.remove('alpha'));
      final c = h.client.fileSpaces!;
      await c.load();
      expect(c.listState, ConfigLoad.failed);
      expect(c.listError, contains('do not have access'));
    });

    test('an expired session signs the person out', () async {
      final h = await start();
      final c = h.client.fileSpaces!;
      h.server.expired = true;
      await c.load();
      await settle();
      expect(h.client.error, contains('Session expired'));
      expect(h.client.fileSpaces, isNull);
      expect(c.items, isEmpty);
    });
  });

  group('saving a file', () {
    test(
      'a new file is written normalized, as text, with a content type',
      () async {
        final h = await start();
        final c = h.client.fileSpaces!;
        await c.openSpace('docs');
        final saved = await c.saveFile(
          space: 'docs',
          path: ' notes//a/../todo.md ',
          content: 'héllo 文',
          baseVersion: null,
        );
        expect(saved!.path, '/notes/todo.md');
        expect(
          h.server.contentOf('alpha', 'docs', '/notes/todo.md'),
          'héllo 文',
        );
        final sent = WriteAgentFileRequest.fromBuffer(
          h.server.fileCalls
              .lastWhere((c) => c.method == 'WriteAgentFile')
              .body,
        );
        expect(sent.contentType, defaultFileContentType);
        expect(c.files.map((f) => f.path), ['/notes/todo.md']);
      },
    );

    test(
      'an edit carries the content type and metadata it started from',
      () async {
        final h = await start(
          seed: (s) => s.seedFile(
            'alpha',
            'docs',
            '/a.md',
            ['one'],
            contentType: 'text/markdown',
            metadata: {'owner': 'me', 'tag': 'x'},
          ),
        );
        final c = h.client.fileSpaces!;
        final read = (await c.readFile('docs', '/a.md')).document!;
        expect(
          await c.saveFile(
            space: 'docs',
            path: '/a.md',
            content: 'two',
            baseVersion: read.version,
            contentType: read.file.contentType,
            metadata: Map.of(read.file.metadata),
          ),
          isNotNull,
        );
        expect(h.server.versionsOf('alpha', 'docs', '/a.md'), 2);
        expect(h.server.metadataOf('alpha', 'docs', '/a.md'), {
          'owner': 'me',
          'tag': 'x',
        });
        final sent = WriteAgentFileRequest.fromBuffer(
          h.server.fileCalls
              .lastWhere((c) => c.method == 'WriteAgentFile')
              .body,
        );
        expect(sent.contentType, 'text/markdown');
      },
    );

    test('a new file over an existing one is refused before writing', () async {
      final h = await start(
        seed: (s) => s.seedFile('alpha', 'docs', '/a.md', ['theirs']),
      );
      final c = h.client.fileSpaces!;
      final saved = await c.saveFile(
        space: 'docs',
        path: '/a.md',
        content: 'mine',
        baseVersion: null,
      );
      expect(saved, isNull);
      expect(c.conflict!.kind, FileConflictKind.exists);
      expect(c.writeError, contains('already exists'));
      expect(c.unconfirmed, isFalse);
      expect(h.server.fileCount('WriteAgentFile'), 0);
      expect(h.server.contentOf('alpha', 'docs', '/a.md'), 'theirs');
    });

    test(
      'an edit over a newer version is refused, then saved on request',
      () async {
        final h = await start(
          seed: (s) => s.seedFile('alpha', 'docs', '/a.md', ['one', 'two']),
        );
        final c = h.client.fileSpaces!;
        expect(
          await c.saveFile(
            space: 'docs',
            path: '/a.md',
            content: 'mine',
            baseVersion: 1,
          ),
          isNull,
        );
        expect(c.conflict!.kind, FileConflictKind.changed);
        expect(c.conflict!.currentVersion, 2);
        expect(h.server.fileCount('WriteAgentFile'), 0);
        c.clearConflict();
        final saved = await c.saveFile(
          space: 'docs',
          path: '/a.md',
          content: 'mine',
          baseVersion: 1,
          overwrite: true,
        );
        expect(saved!.version.toInt(), 3);
        expect(h.server.versionsOf('alpha', 'docs', '/a.md'), 3);
      },
    );

    test(
      'an edit of a file deleted meanwhile is a conflict, not a recreate',
      () async {
        final h = await start();
        final c = h.client.fileSpaces!;
        expect(
          await c.saveFile(
            space: 'docs',
            path: '/gone.md',
            content: 'mine',
            baseVersion: 2,
          ),
          isNull,
        );
        expect(c.conflict!.kind, FileConflictKind.deleted);
        expect(h.server.fileCount('WriteAgentFile'), 0);
      },
    );

    test(
      'a failed look before saving writes nothing and is not unknown',
      () async {
        final h = await start(
          seed: (s) => s.seedFile('alpha', 'docs', '/a.md', ['one']),
        );
        final c = h.client.fileSpaces!;
        h.server.fileRefusals['GetAgentFile'] = (503, 'unavailable', 'down');
        expect(
          await c.saveFile(
            space: 'docs',
            path: '/a.md',
            content: 'mine',
            baseVersion: 1,
          ),
          isNull,
        );
        expect(c.writeError, contains('nothing was saved'));
        expect(c.unconfirmed, isFalse);
        expect(h.server.fileCount('WriteAgentFile'), 0);
      },
    );

    test(
      'the server\'s size refusal changes nothing and names the limit',
      () async {
        final h = await start();
        h.server.maxBytes = 10;
        final c = h.client.fileSpaces!;
        expect(
          await c.saveFile(
            space: 'docs',
            path: '/big.md',
            content: '12345678901',
            baseVersion: null,
          ),
          isNull,
        );
        expect(c.writeError, contains('10-byte limit'));
        expect(c.unconfirmed, isFalse);
        expect(h.server.contentOf('alpha', 'docs', '/big.md'), isNull);
      },
    );

    test('a path the server would refuse is never sent', () async {
      final h = await start();
      final c = h.client.fileSpaces!;
      expect(
        await c.saveFile(
          space: 'docs',
          path: '/',
          content: 'x',
          baseVersion: null,
        ),
        isNull,
      );
      expect(
        h.server.fileCalls.where((c) => c.method == 'WriteAgentFile'),
        isEmpty,
      );
    });

    test(
      'a lost reply is unconfirmed, never repeated, and checked by a read',
      () async {
        final h = await start();
        final c = h.client.fileSpaces!;
        h.server.lostFileReplies.add('WriteAgentFile');
        expect(
          await c.saveFile(
            space: 'docs',
            path: '/a.md',
            content: 'mine',
            baseVersion: null,
          ),
          isNull,
        );
        expect(c.unconfirmed, isTrue);
        expect(c.writeNotice, contains('may already'));
        h.server.lostFileReplies.clear();
        // Every further write is refused until the file has been read again.
        expect(
          await c.saveFile(
            space: 'docs',
            path: '/a.md',
            content: 'mine',
            baseVersion: null,
          ),
          isNull,
        );
        expect(h.server.fileCount('WriteAgentFile'), 1);
        expect(
          await c.verifyWrite('docs', '/a.md', expectedContent: 'mine'),
          FileVerification.applied,
        );
        expect(c.unconfirmed, isFalse);
      },
    );

    test('a verification says when the lost write did not happen', () async {
      final h = await start();
      final c = h.client.fileSpaces!;
      // The server was unreachable: the client cannot tell whether the
      // write happened (here it did not).
      h.server.fileRefusals['WriteAgentFile'] = (503, 'unavailable', 'down');
      expect(
        await c.saveFile(
          space: 'docs',
          path: '/a.md',
          content: 'mine',
          baseVersion: null,
        ),
        isNull,
      );
      expect(c.unconfirmed, isTrue);
      expect(
        await c.verifyWrite('docs', '/a.md', expectedContent: 'mine'),
        FileVerification.notApplied,
      );
      expect(c.unconfirmed, isFalse);
    });

    test('a refusal for lack of access is clear and not unknown', () async {
      final h = await start();
      final c = h.client.fileSpaces!;
      h.server.fileRefusals['WriteAgentFile'] = (
        403,
        'permission_denied',
        'no',
      );
      expect(
        await c.saveFile(
          space: 'docs',
          path: '/a.md',
          content: 'mine',
          baseVersion: null,
        ),
        isNull,
      );
      expect(c.writeError, contains('do not have access'));
      expect(c.unconfirmed, isFalse);
    });

    test('a read that began before a write is dropped, not applied', () async {
      final h = await start(
        seed: (s) => s.seedFile('alpha', 'docs', '/a.md', ['one']),
      );
      final c = h.client.fileSpaces!;
      await c.openSpace('docs');
      final gate = Completer<void>();
      h.server.fileBefore = (m) =>
          m == 'ListAgentFiles' ? gate.future : Future.value();
      final stale = c.loadFiles();
      await settle();
      await c.saveFile(
        space: 'docs',
        path: '/b.md',
        content: 'new',
        baseVersion: null,
      );
      gate.complete();
      await stale;
      // The late list was taken before /b.md existed; it must not erase it.
      expect(c.files.map((f) => f.path), contains('/b.md'));
    });

    test('a file read that began before a write is stale', () async {
      final h = await start(
        seed: (s) => s.seedFile('alpha', 'docs', '/a.md', ['one']),
      );
      final c = h.client.fileSpaces!;
      final gate = Completer<void>();
      h.server.fileBefore = (m) =>
          m == 'GetAgentFile' ? gate.future : Future.value();
      final read = c.readFile('docs', '/a.md');
      await settle();
      h.server.fileBefore = null;
      await c.saveFile(
        space: 'docs',
        path: '/a.md',
        content: 'two',
        baseVersion: 1,
      );
      gate.complete();
      expect((await read).stale, isTrue);
    });

    test('a read begun before the loss does not settle it', () async {
      final h = await start();
      final c = h.client.fileSpaces!;
      final gate = Completer<void>();
      h.server.fileBefore = (m) =>
          m == 'GetAgentFile' ? gate.future : Future.value();
      final early = c.readFile('docs', '/a.md');
      await settle();
      h.server.fileBefore = null;
      h.server.lostFileReplies.add('WriteAgentFile');
      await c.saveFile(
        space: 'docs',
        path: '/a.md',
        content: 'mine',
        baseVersion: null,
      );
      expect(c.unconfirmed, isTrue);
      gate.complete();
      await early;
      expect(c.unconfirmed, isTrue);
    });

    test(
      'suspending during a write leaves it unconfirmed, never repeated',
      () async {
        final h = await start();
        final c = h.client.fileSpaces!;
        final gate = Completer<void>();
        h.server.fileBefore = (m) =>
            m == 'WriteAgentFile' ? gate.future : Future.value();
        final save = c.saveFile(
          space: 'docs',
          path: '/a.md',
          content: 'mine',
          baseVersion: null,
        );
        await settle();
        h.client.suspend();
        gate.complete();
        expect(await save, isNull);
        expect(c.unconfirmed, isTrue);
        expect(h.server.fileCount('WriteAgentFile'), 1);
        await c.recover();
        expect(h.server.fileCount('WriteAgentFile'), 1);
      },
    );

    test('a refresh of the whole space settles an unknown save', () async {
      final h = await start();
      final c = h.client.fileSpaces!;
      await c.openSpace('docs');
      h.server.lostFileReplies.add('WriteAgentFile');
      await c.saveFile(
        space: 'docs',
        path: '/a.md',
        content: 'mine',
        baseVersion: null,
      );
      expect(c.unconfirmed, isTrue);
      await c.loadFiles(prefix: '/notes');
      expect(c.unconfirmed, isTrue, reason: 'a partial list proves nothing');
      await c.loadFiles(prefix: '');
      expect(c.unconfirmed, isFalse);
      expect(c.files.map((f) => f.path), ['/a.md']);
    });
  });

  group('deleting', () {
    test('a file is deleted once and dropped from the list', () async {
      final h = await start(
        seed: (s) => s.seedFile('alpha', 'docs', '/a.md', ['one']),
      );
      final c = h.client.fileSpaces!;
      await c.openSpace('docs');
      expect(await c.deleteFile('docs', '/a.md'), isTrue);
      expect(c.files, isEmpty);
      expect(h.server.contentOf('alpha', 'docs', '/a.md'), isNull);
    });

    test('a lost delete reply is settled by reading the file', () async {
      final h = await start(
        seed: (s) => s.seedFile('alpha', 'docs', '/a.md', ['one']),
      );
      final c = h.client.fileSpaces!;
      h.server.lostFileReplies.add('DeleteAgentFile');
      expect(await c.deleteFile('docs', '/a.md'), isFalse);
      expect(c.unconfirmed, isTrue);
      expect(await c.deleteFile('docs', '/a.md'), isFalse);
      expect(h.server.fileCount('DeleteAgentFile'), 1);
      expect(await c.verifyWrite('docs', '/a.md'), FileVerification.applied);
      expect(c.unconfirmed, isFalse);
    });

    test(
      'a space is deleted with its files, and its list is dropped',
      () async {
        final h = await start(
          seed: (s) => s.seedFile('alpha', 'docs', '/a.md', ['one']),
        );
        final c = h.client.fileSpaces!;
        await c.load();
        await c.openSpace('docs');
        expect(await c.delete('docs'), isTrue);
        expect(c.items, isEmpty);
        expect(c.files, isEmpty);
        expect(h.server.stores['alpha'], isEmpty);
      },
    );

    test('the Agents that mount a space are found across pages', () async {
      final h = await start(
        seed: (s) {
          for (var i = 0; i < 60; i++) {
            final agent = Agent(agentId: 'a$i', name: 'Agent $i');
            if (i == 3 || i == 55) {
              agent.ensureConfig().fileMounts.add(
                AgentFileMount(spaceId: 'docs', mountPath: '/docs'),
              );
            }
            s.agents['alpha']!['a$i'] = agent;
          }
        },
      );
      final refs = await h.client.fileSpaces!.agentsMounting('docs');
      expect(refs!.complete, isTrue);
      expect(refs.agents, ['Agent 3', 'Agent 55']);
    });

    test('a scan that cannot read the Agents says so', () async {
      final h = await start();
      h.server.listRefusal = (503, 'unavailable', 'down');
      final refs = await h.client.fileSpaces!.agentsMounting('docs');
      expect(refs!.complete, isFalse);
    });
  });

  test(
    'file names and contents are dropped on suspend and read on recover',
    () async {
      final h = await start(
        seed: (s) => s.seedFile('alpha', 'docs', '/a.md', ['one']),
      );
      final c = h.client.fileSpaces!;
      await c.load();
      await c.openSpace('docs');
      h.client.suspend();
      expect(c.items, isEmpty);
      expect(c.files, isEmpty);
      await c.recover();
      expect(c.items.single.id, 'docs');
      expect(c.files.single.path, '/a.md');
    },
  );

  test('another Workspace gets its own controller and its own reads', () async {
    final h = await start(
      seed: (s) => s.seedSpace('beta', 'docs', 'Beta docs'),
    );
    final first = h.client.fileSpaces!;
    await first.load();
    await h.client.selectWorkspace('beta');
    expect(first.disposed, isTrue);
    final second = h.client.fileSpaces!;
    expect(identical(first, second), isFalse);
    await second.load();
    expect(second.items.single.name, 'Beta docs');
    expect(h.server.fileCalls.last.workspace, 'beta');
  });
}
