import 'dart:convert';

import '../../core/avatar.dart';
import 'chat_model.dart';
import 'history.dart';
import 'run_request.dart';

/// Text and images the composer can restore after a failed or stopped run.
final class RestorableTurn {
  const RestorableTurn({required this.text, required this.images});
  final String text;
  final List<ChatImage> images;

  bool get isEmpty => text.trim().isEmpty && images.isEmpty;
}

/// A run that answered an Interrupt or carried tool results sent no turn.
const emptyRestorable = RestorableTurn(text: '', images: []);

/// Builds restore input from the server's [LastRun] and the cut history.
RestorableTurn restorableFromLastRun(LastRun last, List<ChatMessage> messages) {
  final input = last.input;
  if (input == null || input.trim().isEmpty) {
    return emptyRestorable;
  }
  ChatMessage? user;
  for (var i = messages.length - 1; i >= 0; i--) {
    if (messages[i].role == ChatRole.user) {
      user = messages[i];
      break;
    }
  }
  if (user != null && user.text == input) {
    return _fromMessage(user);
  }
  return RestorableTurn(text: input, images: []);
}

RestorableTurn restorableFromUserTurn(ChatMessage turn) {
  if (turn.role != ChatRole.user) return emptyRestorable;
  return _fromMessage(turn);
}

RestorableTurn _fromMessage(ChatMessage message) {
  final images = <ChatImage>[];
  for (final part in message.parts) {
    if (part is! ImagePart) continue;
    final size = _decodedLength(part.base64);
    if (size == null || size == 0 || size > maxImageBytes) continue;
    if (images.length >= maxImagesPerMessage) break;
    try {
      final bytes = base64Decode(part.base64);
      if (sniffImageContentType(bytes) == null) continue;
      images.add(ChatImage.fromBytes('restored', bytes));
    } on FormatException {
      continue;
    }
  }
  return RestorableTurn(text: message.text, images: images);
}

int? _decodedLength(String data) {
  if (data.isEmpty || data.length % 4 != 0) return null;
  var padding = 0;
  if (data.endsWith('==')) {
    padding = 2;
  } else if (data.endsWith('=')) {
    padding = 1;
  }
  return data.length ~/ 4 * 3 - padding;
}
