import 'dart:convert';
import 'dart:typed_data';

import '../../core/avatar.dart';

/// The limits `POST /api/agui/:agent_id` applies to one user turn
/// (docs/api.md § Message content). The server answers 400/413 for a turn
/// that breaks them, so the App refuses it first, before anything is sent.
const maxImageBytes = 10 * 1024 * 1024;
const maxImagesPerMessage = 10;
const maxMessageBytes = 20 * 1024 * 1024;
const maxTextBytes = 1024 * 1024;
const maxRequestBodyBytes = 32 * 1024 * 1024;

/// A turn the user can fix (too large, wrong format, nothing to send). The
/// message is safe to show as it is.
final class ChatInputError implements Exception {
  const ChatInputError(this.message);
  final String message;
  @override
  String toString() => message;
}

/// An image the user picked, checked against the server's whitelist and its
/// per-image limit. Only the bytes decide the format.
final class ChatImage {
  ChatImage._(this.name, this.bytes, this.mimeType);

  factory ChatImage.fromBytes(String name, List<int> raw) {
    final bytes = raw is Uint8List ? raw : Uint8List.fromList(raw);
    if (bytes.isEmpty) {
      throw const ChatInputError('That image is empty.');
    }
    if (bytes.length > maxImageBytes) {
      throw const ChatInputError('Each image can be at most 10 MiB.');
    }
    final type = sniffImageContentType(bytes);
    if (type == null) {
      throw const ChatInputError('Choose a PNG, JPEG, GIF or WebP image.');
    }
    return ChatImage._(name, bytes, type);
  }

  final String name;
  final Uint8List bytes;
  final String mimeType;
}

/// The images waiting in the composer. A refused image changes nothing.
final class ImageTray {
  final _images = <ChatImage>[];

  List<ChatImage> get images => List.unmodifiable(_images);
  int get totalBytes => _images.fold(0, (sum, i) => sum + i.bytes.length);

  void add(ChatImage image) {
    if (_images.length >= maxImagesPerMessage) {
      throw const ChatInputError('A message can carry at most 10 images.');
    }
    if (totalBytes + image.bytes.length > maxMessageBytes) {
      throw const ChatInputError('The images can be at most 20 MiB together.');
    }
    _images.add(image);
  }

  void removeAt(int index) => _images.removeAt(index);
  void clear() => _images.clear();
}

/// The JSON body of one run: only this turn, never the thread's history (the
/// server reads just the trailing message, and earlier images would outgrow
/// the request cap). Every run asks to detach, so leaving the page never ends
/// it, and none negotiates A2UI, sends tool declarations or a `resume`, which
/// the App does not support yet.
Uint8List buildRunRequest({
  required String threadId,
  required String runId,
  required String messageId,
  required String text,
  List<ChatImage> images = const [],
  int maxBodyBytes = maxRequestBodyBytes,
}) {
  final hasText = text.trim().isNotEmpty;
  if (!hasText && images.isEmpty) {
    throw const ChatInputError('Write a message or add an image.');
  }
  if (images.length > maxImagesPerMessage) {
    throw const ChatInputError('A message can carry at most 10 images.');
  }
  final textBytes = utf8.encode(text).length;
  if (textBytes > maxTextBytes) {
    throw const ChatInputError('The message text can be at most 1 MiB.');
  }
  final imageBytes = images.fold(0, (sum, i) => sum + i.bytes.length);
  if (textBytes + imageBytes > maxMessageBytes) {
    throw const ChatInputError(
      'The message and its images can be at most 20 MiB together.',
    );
  }
  final Object content = images.isEmpty
      ? text
      : [
          if (hasText) {'type': 'text', 'text': text},
          for (final image in images)
            {
              'type': 'image',
              'source': {
                'type': 'data',
                'value': base64Encode(image.bytes),
                'mimeType': image.mimeType,
              },
            },
        ];
  final body = Uint8List.fromList(
    utf8.encode(
      jsonEncode({
        'threadId': threadId,
        'runId': runId,
        'messages': [
          {'id': messageId, 'role': 'user', 'content': content},
        ],
        'tools': const [],
        'context': const [],
        'state': null,
        'forwardedProps': {
          'butterRun': {'detach': true},
        },
      }),
    ),
  );
  if (body.length > maxBodyBytes) {
    throw const ChatInputError('This message is too large to send.');
  }
  return body;
}
