import 'dart:convert';

import 'package:butter_mobile/features/chat/run_request.dart';
import 'package:butter_mobile/features/chat/threads.dart';
import 'package:protobuf/well_known_types/google/protobuf/struct.pb.dart';
import 'package:flutter_test/flutter_test.dart';

import 'support/chat_fixtures.dart';

List<String> ids(List<ThreadRow> rows) => [for (final r in rows) r.threadId];

void main() {
  group('which sessions are this Workspace\'s threads', () {
    test(
      'a bound AG-UI session of this Workspace is a thread of its agent',
      () {
        final rows = threadRows(
          [session('t1')],
          'alpha',
          [agent('agent-1', 'Helper')],
        );
        expect(rows, hasLength(1));
        expect(rows.single.threadId, 't1');
        expect(rows.single.agentId, 'agent-1');
        expect(rows.single.agent?.name, 'Helper');
      },
    );

    test(
      'another Workspace\'s thread is left out even for the same agent id',
      () {
        final rows = threadRows(
          [session('t1', workspace: 'beta'), session('t2')],
          'alpha',
          [agent('agent-1', 'Helper')],
        );
        expect(ids(rows), ['t2']);
      },
    );

    test('other apps and unbound history never appear', () {
      final rows = threadRows(
        [
          session('t1', app: 'cron:nightly'),
          unbound('t2'),
          session('t3', sessionId: 'plain-id'),
          session('t4'),
        ],
        'alpha',
        const [],
      );
      expect(ids(rows), ['t4']);
    });

    test('a binding for another thread, or a broken one, is not trusted', () {
      final rows = threadRows(
        [
          session('t1', bindingThread: 'other'),
          session('t2', binding: Value(stringValue: '{not json')),
          session('t3', binding: Value(stringValue: '"a string"')),
          session(
            't4',
            binding: Value(
              stringValue: jsonEncode({
                'workspace_id': 'alpha',
                'agent_id': '',
                'thread_id': 't4',
              }),
            ),
          ),
          session('t5', binding: Value(numberValue: 5)),
          session('t6'),
        ],
        'alpha',
        const [],
      );
      expect(ids(rows), ['t6']);
    });

    test('a binding stored as a structure is read like a JSON string', () {
      final value = Value(
        structValue: Struct(
          fields: {
            'workspace_id': Value(stringValue: 'alpha'),
            'agent_id': Value(stringValue: 'agent-1'),
            'thread_id': Value(stringValue: 't1'),
          }.entries,
        ),
      );
      expect(
        ids(threadRows([session('t1', binding: value)], 'alpha', const [])),
        ['t1'],
      );
    });

    test('the server\'s order (newest first) is kept', () {
      final rows = threadRows(
        [
          session('new', updated: 9),
          session('mid', updated: 5),
          session('old', updated: 1),
        ],
        'alpha',
        const [],
      );
      expect(ids(rows), ['new', 'mid', 'old']);
    });

    test('a thread whose agent is not in the list is still listed', () {
      final rows = threadRows(
        [session('t1', agent: 'gone')],
        'alpha',
        [agent('agent-1', 'x')],
      );
      expect(rows.single.agent, isNull);
      expect(rows.single.agentId, 'gone');
    });
  });

  group('titles', () {
    test('its own title first, then the agent, then a placeholder', () {
      final rows = threadRows(
        [
          session('own', title: '  Quarterly plan '),
          session('named'),
          session('shown', agent: 'agent-2'),
          session('lost', agent: 'unknown'),
        ],
        'alpha',
        [
          agent('agent-1', 'helper'),
          agent('agent-2', 'internal-name', displayName: 'Friendly name'),
        ],
      );
      expect(rows.map((r) => r.title), [
        'Quarterly plan',
        'helper',
        'Friendly name',
        'Untitled thread',
      ]);
    });
  });

  group('search', () {
    final rows = threadRows(
      [
        session('a', title: 'Budget Review'),
        session('b', title: '会议纪要'),
        session('c'),
      ],
      'alpha',
      [agent('agent-1', 'Helper')],
    );

    test('matches the title ignoring case and surrounding space', () {
      expect(ids(searchThreads(rows, '  bUdGet ')), ['a']);
      expect(ids(searchThreads(rows, '纪要')), ['b']);
    });

    test('also matches the agent name an untitled thread shows', () {
      expect(ids(searchThreads(rows, 'help')), ['c']);
    });

    test('an empty query keeps everything and no match keeps nothing', () {
      expect(searchThreads(rows, ' '), hasLength(3));
      expect(searchThreads(rows, 'zzz'), isEmpty);
    });
  });

  group('a new title', () {
    test('is trimmed and kept to one line', () {
      expect(normalizeThreadTitle('  Plan \n  for\tQ3 '), 'Plan for Q3');
    });

    test('cannot be blank', () {
      expect(
        () => normalizeThreadTitle(' \n '),
        throwsA(isA<ChatInputError>()),
      );
    });

    test('is at most 100 code points, however they are encoded', () {
      expect(normalizeThreadTitle('a' * 100), hasLength(100));
      expect(
        () => normalizeThreadTitle('a' * 101),
        throwsA(isA<ChatInputError>()),
      );
      // 100 emoji are 200 UTF-16 units but 100 code points: allowed.
      expect(normalizeThreadTitle('🙂' * 100).runes.length, 100);
      expect(
        () => normalizeThreadTitle('🙂' * 101),
        throwsA(isA<ChatInputError>()),
      );
    });
  });
}
