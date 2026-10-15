import 'dart:async';
import 'dart:convert';

import 'package:butter_mobile/core/sse.dart';
import 'package:flutter_test/flutter_test.dart';

Stream<List<int>> chunked(List<int> bytes, int size) async* {
  for (var i = 0; i < bytes.length; i += size) {
    yield bytes.sublist(i, i + size > bytes.length ? bytes.length : i + size);
  }
}

Future<List<SseEvent>> parse(
  Stream<List<int>> bytes, {
  int maxEventLength = 1 << 20,
}) => parseSse(bytes, maxEventLength: maxEventLength).toList();

Stream<List<int>> text(String value) => Stream.value(utf8.encode(value));

void main() {
  test('one event with the default name', () async {
    final events = await parse(text('data: {"type":"RUN_STARTED"}\n\n'));
    expect(events, hasLength(1));
    expect(events.single.event, 'message');
    expect(events.single.data, '{"type":"RUN_STARTED"}');
  });

  test(
    'bytes may be split anywhere, inside multi-byte characters too',
    () async {
      const payload = 'data: 你好 🙂 ünï\n\ndata: second\n\n';
      final bytes = utf8.encode(payload);
      for (final size in [1, 2, 3, 5, 7]) {
        final events = await parse(chunked(bytes, size));
        expect(events.map((e) => e.data), [
          '你好 🙂 ünï',
          'second',
        ], reason: '$size');
      }
    },
  );

  test('LF, CRLF and lone CR all end lines', () async {
    for (final eol in ['\n', '\r\n', '\r']) {
      final events = await parse(text('data: a$eol${eol}data: b$eol$eol'));
      expect(events.map((e) => e.data), [
        'a',
        'b',
      ], reason: eol.codeUnits.toString());
    }
  });

  test('CRLF split between chunks does not add a blank line', () async {
    final events = await parse(
      Stream.fromIterable([
        utf8.encode('data: a\r'),
        utf8.encode('\ndata: b\r\n\r'),
        utf8.encode('\n'),
      ]),
    );
    expect(events.map((e) => e.data), ['a\nb']);
  });

  test('several data lines are joined with a newline', () async {
    final events = await parse(
      text('data: one\ndata: two\ndata:\ndata: four\n\n'),
    );
    expect(events.single.data, 'one\ntwo\n\nfour');
  });

  test('only one space after the colon is removed', () async {
    final events = await parse(text('data:  indented\n\ndata:tight\n\n'));
    expect(events.map((e) => e.data), [' indented', 'tight']);
  });

  test('comments and heartbeats are skipped, not dispatched', () async {
    final events = await parse(
      text(': heartbeat\n\n: another\ndata: real\n\n: heartbeat\n\n'),
    );
    expect(events.map((e) => e.data), ['real']);
  });

  test(
    'event and id fields are carried, retry and unknown fields ignored',
    () async {
      final events = await parse(
        text('event: custom\nid: 7\nretry: 100\nfoo: bar\ndata: x\n\n'),
      );
      expect(events.single.event, 'custom');
      expect(events.single.id, '7');
      expect(events.single.data, 'x');
    },
  );

  test(
    'an event without data is not dispatched and does not leak a name',
    () async {
      final events = await parse(text('event: lonely\n\ndata: x\n\n'));
      expect(events, hasLength(1));
      expect(events.single.event, 'message');
    },
  );

  test('a byte order mark at the start is ignored', () async {
    final events = await parse(
      Stream.value([0xEF, 0xBB, 0xBF, ...utf8.encode('data: x\n\n')]),
    );
    expect(events.single.data, 'x');
  });

  test('an unterminated final event is discarded', () async {
    final events = await parse(text('data: done\n\ndata: cut off'));
    expect(events.map((e) => e.data), ['done']);
  });

  test('many small events stream through a small buffer', () async {
    final events = await parse(
      text('data: 0123456789\n\n' * 500),
      maxEventLength: 64,
    );
    expect(events, hasLength(500));
  });

  test('one event larger than the bound fails instead of buffering', () async {
    await expectLater(
      parse(text('data: ${'x' * 200}\n\n'), maxEventLength: 64),
      throwsA(isA<SseOverflow>()),
    );
    // The same size spread over many lines is still one event.
    await expectLater(
      parse(text('${'data: xxxxxxxxxx\n' * 20}\n'), maxEventLength: 64),
      throwsA(isA<SseOverflow>()),
    );
    // And a line that never ends is cut off as it grows.
    await expectLater(
      parse(
        Stream.fromIterable([
          for (var i = 0; i < 50; i++) utf8.encode('data: xxxxxxxxxx'),
        ]),
        maxEventLength: 64,
      ),
      throwsA(isA<SseOverflow>()),
    );
  });

  test('malformed UTF-8 is an error, not replacement text', () async {
    await expectLater(
      parse(Stream.value([...utf8.encode('data: '), 0xFF, 0x0A, 0x0A])),
      throwsA(isA<FormatException>()),
    );
  });

  test('source errors reach the consumer', () async {
    final source = Stream<List<int>>.multi((c) {
      c.add(utf8.encode('data: a\n\n'));
      c.addError(const FormatException('socket'));
      c.close();
    });
    final seen = <String>[];
    await expectLater(
      parseSse(source).forEach((e) => seen.add(e.data)),
      throwsA(isA<FormatException>()),
    );
    expect(seen, ['a']);
  });
}
