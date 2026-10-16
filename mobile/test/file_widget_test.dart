import 'package:butter_mobile/core/config_resource_controller.dart';
import 'package:butter_mobile/features/files/file_controller.dart';
import 'package:butter_mobile/features/files/file_pages.dart';
import 'package:butter_mobile/features/files/file_rules.dart';
import 'package:butter_mobile/gen/agents/v1/agent.pb.dart';
import 'package:butter_mobile/gen/agents/v1/agent_file.pb.dart';
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';

import 'support/config_rig.dart';
import 'support/file_server.dart';

Future<ConfigRig> rig(
  WidgetTester tester, {
  void Function(FileServerFake)? seed,
}) => signedIn(
  tester,
  using: FileServerFake(),
  seed: (s) {
    final fake = s as FileServerFake;
    fake.seedSpace('alpha', 'docs', 'Docs');
    seed?.call(fake);
  },
);

FileServerFake files(ConfigRig h) => h.server as FileServerFake;

FileSpaceController ctl(ConfigRig h) => h.client.fileSpaces!;

/// Builds [page] in the real zone; its first read keeps a progress bar
/// moving, so the tree is not settled until the caller's wait says so.
Future<void> mount(WidgetTester tester, Widget page) async {
  await tester.runAsync(() => tester.pumpWidget(MaterialApp(home: page)));
}

Future<void> openSpace(WidgetTester tester, ConfigRig h) async {
  await mount(tester, SpaceDetailPage(client: h.client, id: 'docs'));
  await waitFor(tester, () => ctl(h).filesState == ConfigLoad.ready);
}

Future<void> openFile(WidgetTester tester, ConfigRig h, String path) async {
  // The space list is what names the space in the dialogs.
  await tester.runAsync(() => ctl(h).load());
  // The space list is what names the space in the dialogs.
  await tester.runAsync(() => ctl(h).load());
  await mount(
    tester,
    FileViewPage(client: h.client, space: 'docs', path: path),
  );
  await waitFor(
    tester,
    () =>
        find.byKey(const Key('file-version')).evaluate().isNotEmpty ||
        find.byKey(const Key('file-too-large')).evaluate().isNotEmpty,
  );
}

Future<void> openEditor(
  WidgetTester tester,
  ConfigRig h, {
  FileDocument? existing,
}) async {
  await openPage(
    tester,
    FileEditPage(client: h.client, space: 'docs', existing: existing),
  );
}

Future<FileDocument> latest(ConfigRig h, String path) async =>
    (await ctl(h).readFile('docs', path)).document!;

void main() {
  testWidgets('lists this Workspace\'s file spaces only', (tester) async {
    final h = await rig(
      tester,
      seed: (s) {
        s.seedSpace('alpha', 'notes', 'Notes', description: 'Daily notes');
        s.seedSpace('beta', 'secret', 'Beta only');
      },
    );
    await openPage(tester, FileSpacesPage(client: h.client));
    await tester.runAsync(() => ctl(h).load());
    await tester.pumpAndSettle();
    expect(find.text('Docs'), findsOneWidget);
    expect(find.text('Notes'), findsOneWidget);
    expect(find.text('Beta only'), findsNothing);
    expect(find.textContaining('Daily notes'), findsOneWidget);
  });

  testWidgets('a Workspace without access shows a permission message', (
    tester,
  ) async {
    final h = await rig(tester, seed: (s) => s.stores.remove('alpha'));
    await openPage(tester, FileSpacesPage(client: h.client));
    await tester.runAsync(() => ctl(h).load());
    await tester.pumpAndSettle();
    expect(find.textContaining('do not have access'), findsOneWidget);
    expect(find.text('Docs'), findsNothing);
  });

  testWidgets('a server without file storage is explained', (tester) async {
    final h = await rig(
      tester,
      seed: (s) => s.fileRefusals['ListAgentFileSpaces'] = (
        412,
        'failed_precondition',
        'agent file repository not available',
      ),
    );
    await openPage(tester, FileSpacesPage(client: h.client));
    await tester.runAsync(() => ctl(h).load());
    await tester.pumpAndSettle();
    expect(find.textContaining('File storage is not set up'), findsOneWidget);
  });

  testWidgets('a space shows its files with size and version', (tester) async {
    final h = await rig(
      tester,
      seed: (s) {
        s.seedFile('alpha', 'docs', '/a.md', ['one', 'two two']);
        s.seedFile('alpha', 'docs', '/notes/b.md', ['b']);
      },
    );
    await openSpace(tester, h);
    expect(find.text('/a.md'), findsOneWidget);
    expect(find.textContaining('7 B · version 2'), findsOneWidget);
    expect(find.text('/notes/b.md'), findsOneWidget);
  });

  testWidgets('a path filter narrows the list', (tester) async {
    final h = await rig(
      tester,
      seed: (s) {
        s.seedFile('alpha', 'docs', '/a.md', ['one']);
        s.seedFile('alpha', 'docs', '/notes/b.md', ['b']);
      },
    );
    await openSpace(tester, h);
    await enter(tester, const Key('files-prefix'), 'notes');
    await tapReal(tester, find.byKey(const Key('apply-prefix')));
    await waitFor(
      tester,
      () =>
          ctl(h).filesPrefix == 'notes' &&
          ctl(h).filesState == ConfigLoad.ready,
    );
    expect(find.text('/a.md'), findsNothing);
    expect(find.text('/notes/b.md'), findsOneWidget);
  });

  testWidgets('searching needs text and shows matches', (tester) async {
    final h = await rig(
      tester,
      seed: (s) =>
          s.seedFile('alpha', 'docs', '/a.md', ['find the needle now']),
    );
    await openSpace(tester, h);
    await tapReal(tester, find.byKey(const Key('run-search')));
    await tester.pumpAndSettle();
    expect(find.textContaining('Enter text to search for'), findsOneWidget);
    expect(files(h).fileCount('SearchAgentFiles'), 0);
    await enter(tester, const Key('files-search'), 'needle');
    await tapReal(tester, find.byKey(const Key('run-search')));
    await waitFor(
      tester,
      () => find.byKey(const Key('search-summary')).evaluate().isNotEmpty,
    );
    expect(find.textContaining('1 match for "needle"'), findsOneWidget);
    expect(find.byKey(const Key('hit-/a.md')), findsOneWidget);
  });

  testWidgets('a new file shows the normalized path and the UTF-8 size, and '
      'is saved as text', (tester) async {
    final h = await rig(tester);
    await openEditor(tester, h);
    await enter(tester, const Key('file-path'), ' notes//a/../todo.md ');
    expect(find.text('Saved as /notes/todo.md'), findsOneWidget);
    await enter(tester, const Key('file-content'), '文文');
    expect(find.textContaining('(6 bytes)'), findsOneWidget);
    await tapReal(tester, find.byKey(const Key('save-file')));
    await waitFor(tester, () => files(h).fileCount('WriteAgentFile') == 1);
    expect(files(h).contentOf('alpha', 'docs', '/notes/todo.md'), '文文');
  });

  testWidgets('an unusable path is refused before anything is sent', (
    tester,
  ) async {
    final h = await rig(tester);
    await openEditor(tester, h);
    await enter(tester, const Key('file-path'), '/');
    expect(find.text('This is not a usable path.'), findsOneWidget);
    await enter(tester, const Key('file-content'), 'x');
    await tapReal(tester, find.byKey(const Key('save-file')));
    await tester.pumpAndSettle();
    expect(find.textContaining('Enter a file path'), findsOneWidget);
    expect(files(h).fileCount('WriteAgentFile'), 0);
  });

  testWidgets('text over the usual limit is warned about, not blocked', (
    tester,
  ) async {
    final h = await rig(tester, seed: (s) => s.maxBytes = 300 * 1024);
    await openEditor(tester, h);
    await enter(tester, const Key('file-path'), '/big.txt');
    await enter(
      tester,
      const Key('file-content'),
      'a' * (defaultFileMaxBytes + 1),
    );
    expect(find.byKey(const Key('file-over-limit')), findsOneWidget);
    // The server's limit here is higher: the warning did not block the save.
    await tapReal(tester, find.byKey(const Key('save-file')));
    await waitFor(tester, () => files(h).fileCount('WriteAgentFile') == 1);
    expect(
      files(h).contentOf('alpha', 'docs', '/big.txt')!.length,
      defaultFileMaxBytes + 1,
    );
  });

  testWidgets('a refusal for size says the limit and keeps the text', (
    tester,
  ) async {
    final h = await rig(tester, seed: (s) => s.maxBytes = 10);
    await openEditor(tester, h);
    await enter(tester, const Key('file-path'), '/big.txt');
    await enter(tester, const Key('file-content'), 'x' * 40);
    await tapReal(tester, find.byKey(const Key('save-file')));
    await waitFor(tester, () => ctl(h).writeError != null);
    expect(find.textContaining('10-byte limit'), findsOneWidget);
    expect(find.text('x' * 40), findsOneWidget);
    expect(files(h).contentOf('alpha', 'docs', '/big.txt'), isNull);
  });

  testWidgets('a file that is already there is a conflict, not an overwrite', (
    tester,
  ) async {
    final h = await rig(
      tester,
      seed: (s) => s.seedFile('alpha', 'docs', '/a.md', ['theirs']),
    );
    await openEditor(tester, h);
    await enter(tester, const Key('file-path'), '/a.md');
    await enter(tester, const Key('file-content'), 'mine');
    await tapReal(tester, find.byKey(const Key('save-file')));
    await waitFor(
      tester,
      () => find.text('A file is already there').evaluate().isNotEmpty,
    );
    expect(files(h).fileCount('WriteAgentFile'), 0);
    await tapDialog(tester, find.byKey(const Key('conflict-cancel')));
    expect(files(h).contentOf('alpha', 'docs', '/a.md'), 'theirs');
    expect(find.text('mine'), findsOneWidget, reason: 'the draft is kept');
  });

  testWidgets('an edit over a newer version offers review or overwrite', (
    tester,
  ) async {
    final h = await rig(
      tester,
      seed: (s) => s.seedFile(
        'alpha',
        'docs',
        '/a.md',
        ['one'],
        contentType: 'text/markdown',
        metadata: {'owner': 'me'},
      ),
    );
    final opened = await tester.runAsync(() => latest(h, '/a.md'));
    await openEditor(tester, h, existing: opened);
    // Somebody else saves while this draft is open.
    files(h).seedFile(
      'alpha',
      'docs',
      '/a.md',
      ['one', 'theirs'],
      contentType: 'text/markdown',
      metadata: {'owner': 'me'},
    );
    await enter(tester, const Key('file-content'), 'mine');
    await tapReal(tester, find.byKey(const Key('save-file')));
    await waitFor(
      tester,
      () => find.text('The file changed').evaluate().isNotEmpty,
    );
    expect(files(h).versionsOf('alpha', 'docs', '/a.md'), 2);
    await tapDialogReal(tester, find.byKey(const Key('conflict-overwrite')));
    await waitFor(
      tester,
      () => files(h).versionsOf('alpha', 'docs', '/a.md') == 3,
    );
    expect(files(h).contentOf('alpha', 'docs', '/a.md'), 'mine');
    // The content type and metadata the file had are carried over.
    expect(files(h).metadataOf('alpha', 'docs', '/a.md'), {'owner': 'me'});
  });

  testWidgets('a lost reply is not repeated and can be checked', (
    tester,
  ) async {
    final h = await rig(tester);
    files(h).lostFileReplies.add('WriteAgentFile');
    await openEditor(tester, h);
    await enter(tester, const Key('file-path'), '/a.md');
    await enter(tester, const Key('file-content'), 'mine');
    await tapReal(tester, find.byKey(const Key('save-file')));
    await waitFor(tester, () => ctl(h).unconfirmed);
    expect(find.textContaining('Change not confirmed'), findsOneWidget);
    // Save is off until the file has been read; nothing can repeat it.
    final save = tester.widget<FilledButton>(
      find.byKey(const Key('save-file')),
    );
    expect(save.onPressed, isNull);
    await tapReal(tester, find.byKey(const Key('verify-file')));
    await waitFor(tester, () => !ctl(h).unconfirmed);
    expect(find.byKey(const Key('verify-note')), findsOneWidget);
    expect(find.textContaining('save went through'), findsOneWidget);
    expect(files(h).fileCount('WriteAgentFile'), 1);
  });

  testWidgets('a file shows its content, version and size', (tester) async {
    final h = await rig(
      tester,
      seed: (s) => s.seedFile(
        'alpha',
        'docs',
        '/a.md',
        ['first', 'second text'],
        metadata: {'owner': 'me'},
      ),
    );
    await openFile(tester, h, '/a.md');
    expect(find.text('Version 2 of 2'), findsOneWidget);
    expect(find.text('second text'), findsOneWidget);
    expect(find.textContaining('Metadata: owner'), findsOneWidget);
    expect(find.byKey(const Key('edit-file')), findsOneWidget);
  });

  testWidgets('an earlier version is read by number and restored as a new '
      'version, keeping the current metadata', (tester) async {
    final h = await rig(
      tester,
      seed: (s) => s.seedFile(
        'alpha',
        'docs',
        '/a.md',
        ['first', 'second', 'third'],
        contentType: 'text/markdown',
        metadata: {'owner': 'me'},
      ),
    );
    await openFile(tester, h, '/a.md');
    await tapReal(tester, find.byKey(const Key('older-version')));
    await waitFor(
      tester,
      () => find.text('Version 2 of 3 (not the latest)').evaluate().isNotEmpty,
    );
    expect(find.text('second'), findsOneWidget);
    // The latest and Edit are not offered on an old version.
    expect(find.byKey(const Key('edit-file')), findsNothing);
    await enter(tester, const Key('version-jump'), '1');
    await tapReal(tester, find.byKey(const Key('go-version')));
    await waitFor(
      tester,
      () => find.text('Version 1 of 3 (not the latest)').evaluate().isNotEmpty,
    );
    expect(find.text('first'), findsOneWidget);
    await tapReal(tester, find.byKey(const Key('restore-version')));
    await tester.pumpAndSettle();
    expect(find.textContaining('as a new version 4'), findsOneWidget);
    await tapDialogReal(tester, find.text('Restore'));
    await waitFor(
      tester,
      () => files(h).versionsOf('alpha', 'docs', '/a.md') == 4,
    );
    expect(files(h).contentOf('alpha', 'docs', '/a.md'), 'first');
    expect(files(h).metadataOf('alpha', 'docs', '/a.md'), {'owner': 'me'});
    await waitFor(
      tester,
      () => find.text('Version 4 of 4').evaluate().isNotEmpty,
    );
  });

  testWidgets('a version that does not exist is said so', (tester) async {
    final h = await rig(
      tester,
      seed: (s) => s.seedFile('alpha', 'docs', '/a.md', ['first', 'second']),
    );
    await openFile(tester, h, '/a.md');
    await enter(tester, const Key('version-jump'), '9');
    await tapReal(tester, find.byKey(const Key('go-version')));
    await tester.pumpAndSettle();
    expect(find.textContaining('Enter a version from 1 to 2'), findsOneWidget);
    expect(files(h).fileCount('GetAgentFile'), 1);
  });

  testWidgets('deleting a file names it, and cancelling changes nothing', (
    tester,
  ) async {
    final h = await rig(
      tester,
      seed: (s) => s.seedFile('alpha', 'docs', '/a.md', ['one']),
    );
    await openFile(tester, h, '/a.md');
    await tapReal(tester, find.byKey(const Key('delete-file')));
    await tester.pumpAndSettle();
    expect(find.textContaining('Delete /a.md from "Docs"?'), findsOneWidget);
    await tapDialog(tester, find.text('Cancel'));
    expect(files(h).fileCount('DeleteAgentFile'), 0);
    await tapReal(tester, find.byKey(const Key('delete-file')));
    await tester.pumpAndSettle();
    await tapDialog(tester, find.text('Delete file'));
    await waitFor(tester, () => files(h).fileCount('DeleteAgentFile') == 1);
    expect(files(h).contentOf('alpha', 'docs', '/a.md'), isNull);
  });

  testWidgets('a file over 1 MiB is not opened here', (tester) async {
    final h = await rig(
      tester,
      seed: (s) {
        s.maxBytes = 4 * 1024 * 1024;
        s.seedFile('alpha', 'docs', '/huge.txt', [
          'a' * (editableFileLimit + 5),
        ]);
      },
    );
    await mount(tester, SpaceDetailPage(client: h.client, id: 'docs'));
    await waitFor(tester, () => ctl(h).filesState == ConfigLoad.ready);
    await tapScreen(tester, find.byKey(const Key('file-/huge.txt')));
    await waitFor(
      tester,
      () => find.byKey(const Key('file-too-large')).evaluate().isNotEmpty,
    );
    expect(files(h).fileCount('GetAgentFile'), 0);
  });

  testWidgets('deleting a space says how many files go and which Agents '
      'mount it', (tester) async {
    final h = await rig(
      tester,
      seed: (s) {
        s.seedFile('alpha', 'docs', '/a.md', ['one']);
        s.seedFile('alpha', 'docs', '/b.md', ['two']);
        final agent = Agent(agentId: 'reader', name: 'Reader');
        agent.ensureConfig().fileMounts.add(
          AgentFileMount(spaceId: 'docs', mountPath: '/docs'),
        );
        s.agents['alpha']!['reader'] = agent;
      },
    );
    await openSpace(tester, h);
    await tapReal(tester, find.byKey(const Key('delete-space')));
    await waitFor(tester, () => find.byType(AlertDialog).evaluate().isNotEmpty);
    expect(find.textContaining('Delete "Docs"?'), findsOneWidget);
    expect(
      find.textContaining('2 files and all their versions'),
      findsOneWidget,
    );
    expect(
      find.textContaining('Agents that mount it (Reader)'),
      findsOneWidget,
    );
    await tapDialog(tester, find.text('Cancel'));
    expect(files(h).fileCount('DeleteAgentFileSpace'), 0);
  });

  testWidgets('editing a space keeps what the form does not show', (
    tester,
  ) async {
    final h = await rig(
      tester,
      seed: (s) => s.seedSpace(
        'alpha',
        'docs',
        'Docs',
        description: 'Old',
        metadata: {'team': 'platform'},
        unknown: [99],
      ),
    );
    await mount(tester, SpaceFormPage(client: h.client, id: 'docs'));
    await waitFor(
      tester,
      () => find.byKey(const Key('space-name')).evaluate().isNotEmpty,
    );
    expect(
      find.textContaining('1 metadata entry set elsewhere'),
      findsOneWidget,
    );
    await enter(tester, const Key('space-name'), 'Docs v2');
    await tapReal(tester, find.byKey(const Key('save-space')));
    await waitFor(
      tester,
      () => files(h).fileCount('UpdateAgentFileSpace') == 1,
    );
    final stored = files(h).spaceOf('alpha', 'docs');
    expect(stored.name, 'Docs v2');
    expect(stored.description, 'Old');
    expect(stored.metadata, {'team': 'platform'});
    expect(stored.unknownFields.getField(99)?.varints.single.toInt(), 7);
  });

  testWidgets('a new space is created from a name and description', (
    tester,
  ) async {
    final h = await rig(tester);
    await openPage(tester, SpaceFormPage(client: h.client));
    await tapReal(tester, find.byKey(const Key('save-space')));
    await tester.pumpAndSettle();
    expect(find.text('Enter a name.'), findsOneWidget);
    expect(files(h).fileCount('CreateAgentFileSpace'), 0);
    await enter(tester, const Key('space-name'), 'Notes');
    await enter(tester, const Key('space-description'), 'Daily');
    await tapReal(tester, find.byKey(const Key('save-space')));
    await waitFor(
      tester,
      () => files(h).fileCount('CreateAgentFileSpace') == 1,
    );
    expect(files(h).stores['alpha']!.length, 2);
  });
}
