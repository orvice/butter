import 'dart:convert';
import 'dart:typed_data';

import 'package:butter_mobile/features/chat/run_request.dart';
import 'package:flutter_test/flutter_test.dart';

const mib = 1024 * 1024;

Uint8List png([int size = 16]) {
  final bytes = Uint8List(size);
  bytes.setAll(0, [137, 80, 78, 71, 13, 10, 26, 10]);
  return bytes;
}

Uint8List jpeg() => Uint8List.fromList([255, 216, 255, 224, 0, 16]);
Uint8List gif() => Uint8List.fromList(utf8.encode('GIF89a') + [1, 0, 1, 0]);
Uint8List webp() => Uint8List.fromList(
  utf8.encode('RIFF') + [0, 0, 0, 0] + utf8.encode('WEBP'),
);

ChatImage image([int size = 16, String name = 'a.png']) =>
    ChatImage.fromBytes(name, png(size));

Map<String, dynamic> body(Uint8List bytes) =>
    jsonDecode(utf8.decode(bytes)) as Map<String, dynamic>;

Uint8List build({
  String text = 'hello',
  List<ChatImage> images = const [],
  int? maxBodyBytes,
}) => buildRunRequest(
  threadId: 't-1',
  runId: 'r-1',
  messageId: 'm-1',
  text: text,
  images: images,
  maxBodyBytes: maxBodyBytes ?? 32 * mib,
);

void main() {
  group('request body', () {
    test('a text turn is one user message, detached, with nothing else', () {
      expect(body(build(text: 'hello 你好')), {
        'threadId': 't-1',
        'runId': 'r-1',
        'messages': [
          {'id': 'm-1', 'role': 'user', 'content': 'hello 你好'},
        ],
        'tools': [],
        'context': [],
        'state': null,
        'forwardedProps': {
          'butterRun': {'detach': true},
        },
      });
    });

    test('images follow the text as inline data parts, in order', () {
      final first = ChatImage.fromBytes('one.png', png(20));
      final second = ChatImage.fromBytes('two.jpg', jpeg());
      final sent = body(build(text: 'look', images: [first, second]));
      final content = (sent['messages'] as List).single['content'] as List;
      expect(content, [
        {'type': 'text', 'text': 'look'},
        {
          'type': 'image',
          'source': {
            'type': 'data',
            'value': base64Encode(png(20)),
            'mimeType': 'image/png',
          },
        },
        {
          'type': 'image',
          'source': {
            'type': 'data',
            'value': base64Encode(jpeg()),
            'mimeType': 'image/jpeg',
          },
        },
      ]);
    });

    test('an image alone is a valid turn and carries no empty text part', () {
      final sent = body(build(text: '  ', images: [image()]));
      final content = (sent['messages'] as List).single['content'] as List;
      expect(content.map((p) => p['type']), ['image']);
    });

    test('there is never a resume, an A2UI negotiation or a client tool', () {
      final sent = body(build(images: [image()]));
      expect(sent.containsKey('resume'), isFalse);
      expect(sent['tools'], isEmpty);
      expect((sent['forwardedProps'] as Map).keys, [
        'butterRun',
      ], reason: 'butterA2UI must not be negotiated before it is supported');
    });
  });

  group('one image', () {
    test('the four whitelisted formats are recognised by their bytes', () {
      expect(ChatImage.fromBytes('x', png()).mimeType, 'image/png');
      expect(ChatImage.fromBytes('x', jpeg()).mimeType, 'image/jpeg');
      expect(ChatImage.fromBytes('x', gif()).mimeType, 'image/gif');
      expect(ChatImage.fromBytes('x', webp()).mimeType, 'image/webp');
    });

    test('the file name is never trusted for the format', () {
      expect(ChatImage.fromBytes('photo.jpg', png()).mimeType, 'image/png');
      expect(
        () => ChatImage.fromBytes(
          'photo.png',
          utf8.encode('<svg/>').toUint8List(),
        ),
        throwsA(isA<ChatInputError>()),
      );
    });

    test('empty and over 10 MiB are refused, exactly 10 MiB is accepted', () {
      expect(
        () => ChatImage.fromBytes('x', Uint8List(0)),
        throwsA(isA<ChatInputError>()),
      );
      expect(ChatImage.fromBytes('x', png(10 * mib)).bytes.length, 10 * mib);
      expect(
        () => ChatImage.fromBytes('x', png(10 * mib + 1)),
        throwsA(isA<ChatInputError>()),
      );
    });
  });

  group('image tray', () {
    test('holds ten images and refuses the eleventh', () {
      final tray = ImageTray();
      for (var i = 0; i < 10; i++) {
        tray.add(image());
      }
      expect(() => tray.add(image()), throwsA(isA<ChatInputError>()));
      expect(tray.images, hasLength(10));
    });

    test('the images together stay within 20 MiB', () {
      final tray = ImageTray();
      tray.add(image(9 * mib));
      tray.add(image(9 * mib));
      expect(() => tray.add(image(9 * mib)), throwsA(isA<ChatInputError>()));
      expect(
        tray.images,
        hasLength(2),
        reason: 'a refused image changes nothing',
      );
      tray.add(image(2 * mib));
      expect(tray.totalBytes, 20 * mib);
    });

    test('removing an image frees its room', () {
      final tray = ImageTray();
      tray.add(image(9 * mib));
      tray.add(image(9 * mib));
      tray.removeAt(0);
      tray.add(image(9 * mib));
      expect(tray.images, hasLength(2));
    });

    test('clear empties it', () {
      final tray = ImageTray()..add(image());
      tray.clear();
      expect(tray.images, isEmpty);
      expect(tray.totalBytes, 0);
    });
  });

  group('a turn at send time', () {
    test('needs text or an image', () {
      expect(() => build(text: ' \n '), throwsA(isA<ChatInputError>()));
    });

    test('text is at most 1 MiB', () {
      build(text: 'a' * mib);
      expect(
        () => build(text: 'a' * (mib + 1)),
        throwsA(isA<ChatInputError>()),
      );
    });

    test('the text and the images together are at most 20 MiB', () {
      final images = [image(10 * mib), image(9 * mib)];
      build(text: 'a' * (mib), images: images); // exactly 20 MiB
      expect(
        () => build(text: 'a' * (mib + 1), images: images),
        throwsA(isA<ChatInputError>()),
      );
    });

    test('the encoded body is at most the request cap', () {
      expect(
        () => build(images: [image(1024)], maxBodyBytes: 512),
        throwsA(isA<ChatInputError>()),
      );
      // The largest turn the limits allow still fits the real 32 MiB cap
      // after base64 inflates it.
      final largest = build(
        text: 'a' * mib,
        images: [image(10 * mib), image(9 * mib)],
      );
      expect(largest.length, lessThan(32 * mib));
    });

    test('the message keeps what the user typed, not a trimmed copy', () {
      final sent = body(build(text: '  indented\n\nlines  '));
      expect(
        (sent['messages'] as List).single['content'],
        '  indented\n\nlines  ',
      );
    });
  });
}

extension on List<int> {
  Uint8List toUint8List() => Uint8List.fromList(this);
}
