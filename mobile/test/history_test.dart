import 'dart:convert';
import 'dart:typed_data';

import 'package:butter_mobile/features/chat/chat_model.dart';
import 'package:butter_mobile/features/chat/history.dart';
import 'package:flutter_test/flutter_test.dart';

const mib = 1024 * 1024;

String pngBase64([int size = 16]) {
  final bytes = Uint8List(size)..setAll(0, [137, 80, 78, 71, 13, 10, 26, 10]);
  return base64Encode(bytes);
}

ThreadHistory parse(Map<String, Object?> json) =>
    parseThreadHistory(jsonEncode(json));

Map<String, Object?> userImage(String mime, String data) => {
  'type': 'image',
  'source': {'type': 'data', 'value': data, 'mimeType': mime},
};

void main() {
  test('the documented example becomes messages with the tool call merged', () {
    final history = parse({
      'threadId': 't-1',
      'messages': [
        {'id': 'u1', 'role': 'user', 'content': 'Which tickets expire soon?'},
        {
          'id': 'a1',
          'role': 'assistant',
          'content': 'Two tickets expire this month.',
          'toolCalls': [
            {
              'id': 'call-1',
              'type': 'function',
              'function': {'name': 'queryRecords', 'arguments': '{"q":1}'},
            },
          ],
        },
        {
          'id': 'result:call-1',
          'role': 'tool',
          'toolCallId': 'call-1',
          'content': '{"rows":2}',
        },
      ],
      'interrupts': [
        {'id': 'i1', 'reason': 'human_input', 'message': 'Approve the refund?'},
      ],
      'surfaces': [
        {'surfaceId': 'card-1', 'messageId': 'a1'},
      ],
    });
    expect(history.messages.map((m) => m.role), [
      ChatRole.user,
      ChatRole.assistant,
    ]);
    expect(history.messages.first.text, 'Which tickets expire soon?');
    final reply = history.messages.last;
    expect(
      (reply.parts.first as TextPart).text,
      'Two tickets expire this month.',
    );
    final tool = reply.parts.last as ToolPart;
    expect(tool.name, 'queryRecords');
    expect(tool.arguments, '{"q":1}');
    expect(tool.result, '{"rows":2}');
    expect(tool.status, ToolStatus.done);
    expect(history.interrupts.single.message, 'Approve the refund?');
    expect(history.running, isNull);
    expect(history.lastRun, isNull);
  });

  test('a call without a result is still running, not invented as done', () {
    final history = parse({
      'messages': [
        {
          'id': 'a1',
          'role': 'assistant',
          'content': '',
          'toolCalls': [
            {
              'id': 'c',
              'function': {'name': 'x', 'arguments': '{}'},
            },
          ],
        },
      ],
    });
    final parts = history.messages.single.parts;
    expect(parts, hasLength(1), reason: 'empty text is not a part');
    expect((parts.single as ToolPart).status, ToolStatus.running);
  });

  test('a user turn with images keeps its text and images in order', () {
    final data = pngBase64();
    final history = parse({
      'messages': [
        {
          'id': 'u1',
          'role': 'user',
          'content': [
            {'type': 'text', 'text': 'What is this?'},
            userImage('image/png', data),
            {'type': 'binary', 'mimeType': 'image/png', 'data': data},
          ],
        },
      ],
    });
    final parts = history.messages.single.parts;
    expect(parts.map((p) => p.runtimeType), [TextPart, ImagePart, ImagePart]);
    expect((parts[1] as ImagePart).base64, data);
    expect((parts[1] as ImagePart).mimeType, 'image/png');
  });

  test('an image-only turn is kept', () {
    final history = parse({
      'messages': [
        {
          'id': 'u',
          'role': 'user',
          'content': [userImage('image/png', pngBase64())],
        },
      ],
    });
    expect(history.messages.single.parts.single, isA<ImagePart>());
  });

  test('restored images follow the same format and size limits as sending', () {
    final history = parse({
      'messages': [
        {
          'id': 'u',
          'role': 'user',
          'content': [
            userImage('image/svg+xml', base64Encode(utf8.encode('<svg/>'))),
            userImage('image/png', base64Encode(utf8.encode('not an image'))),
            userImage('image/png', pngBase64(10 * mib + 1)),
            userImage('image/png', pngBase64(10 * mib)),
            userImage('image/png', '***not base64***'),
            {
              'type': 'image',
              'source': {'type': 'url', 'value': 'https://x.test/a.png'},
            },
          ],
        },
      ],
    });
    final parts = history.messages.single.parts;
    expect(parts.map((p) => p.runtimeType), [
      UnavailableImagePart,
      UnavailableImagePart,
      UnavailableImagePart,
      ImagePart,
      UnavailableImagePart,
      UnavailableImagePart,
    ]);
  });

  test('restored images stay within ten per message and 20 MiB together', () {
    final nine = pngBase64(9 * mib);
    final small = pngBase64();
    final history = parse({
      'messages': [
        {
          'id': 'u',
          'role': 'user',
          'content': [
            userImage('image/png', nine),
            userImage('image/png', nine),
            userImage('image/png', nine), // pushes the message past 20 MiB
          ],
        },
        {
          'id': 'u2',
          'role': 'user',
          'content': [
            for (var i = 0; i < 12; i++) userImage('image/png', small),
          ],
        },
      ],
    });
    expect(history.messages[0].parts.map((p) => p.runtimeType), [
      ImagePart,
      ImagePart,
      UnavailableImagePart,
    ]);
    final types = history.messages[1].parts.map((p) => p.runtimeType).toList();
    expect(types.where((t) => t == ImagePart), hasLength(10));
    expect(types.where((t) => t == UnavailableImagePart), hasLength(2));
  });

  test('a run in flight and a failed last run are reported', () {
    final running = parse({
      'messages': [
        {'id': 'u', 'role': 'user', 'content': 'go'},
      ],
      'running': {'runId': 'run-2', 'invocationId': 'inv-9'},
    });
    expect(running.running?.runId, 'run-2');
    final failed = parse({
      'messages': const [],
      'lastRun': {'status': 'failed', 'error': 'model down', 'input': 'hello'},
    });
    expect(failed.lastRun?.status, 'failed');
    expect(failed.lastRun?.error, 'model down');
    expect(failed.lastRun?.input, 'hello');
    final stopped = parse({
      'messages': const [],
      'lastRun': {'status': 'cancelled', 'error': 'stopped by user'},
    });
    expect(stopped.lastRun?.status, 'cancelled');
    expect(stopped.lastRun?.input, isNull);
  });

  test('an empty history of a thread the caller does not own is empty', () {
    final history = parse({'threadId': 't', 'messages': const []});
    expect(history.messages, isEmpty);
    expect(history.running, isNull);
  });

  test('unknown roles and fields are skipped without failing the thread', () {
    final history = parse({
      'future': true,
      'messages': [
        {'id': 's', 'role': 'system', 'content': 'hidden'},
        {'id': 'x', 'role': 'user', 'content': 7},
        {'id': 'u', 'role': 'user', 'content': 'kept', 'new_field': 1},
      ],
    });
    expect(history.messages.map((m) => m.id), ['u']);
  });

  test('something that is not a history is a format error', () {
    expect(() => parseThreadHistory('[]'), throwsFormatException);
    expect(() => parseThreadHistory('nope'), throwsFormatException);
    expect(() => parseThreadHistory('{"messages":"x"}'), throwsFormatException);
  });
}
