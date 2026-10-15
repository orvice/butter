import 'package:flutter/foundation.dart';

enum ChatRole { user, assistant }

/// One piece of a message, in the order it arrived or was sent.
@immutable
sealed class MessagePart {
  const MessagePart();
}

final class TextPart extends MessagePart {
  const TextPart(this.text);
  final String text;
}

/// An inline image of a user turn, kept as the base64 the server stores so a
/// large thread is only decoded for the images that are shown.
final class ImagePart extends MessagePart {
  const ImagePart({required this.mimeType, required this.base64});
  final String mimeType;
  final String base64;
}

/// An image the history holds but the App will not show: not a whitelisted
/// format, over a limit, or not decodable. It takes the image's place so the
/// turn still reads in order.
final class UnavailableImagePart extends MessagePart {
  const UnavailableImagePart();
}

enum ToolStatus {
  /// The call and its arguments are still being written.
  calling,

  /// The call is complete and its result has not arrived.
  running,

  /// The call has a result.
  done,
}

/// A tool call the server ran, with its arguments and result. The App never
/// runs tools itself.
final class ToolPart extends MessagePart {
  const ToolPart({
    required this.id,
    required this.name,
    required this.arguments,
    required this.status,
    this.result,
  });
  final String id;
  final String name;
  final String arguments;
  final String? result;
  final ToolStatus status;
}

@immutable
final class ChatMessage {
  ChatMessage({
    required this.id,
    required this.role,
    required Iterable<MessagePart> parts,
  }) : parts = List.unmodifiable(parts);
  final String id;
  final ChatRole role;
  final List<MessagePart> parts;

  String get text => parts.whereType<TextPart>().map((p) => p.text).join('\n');
}

/// A question the run is waiting on, as RUN_FINISHED reports it.
@immutable
final class InterruptInfo {
  const InterruptInfo({
    required this.id,
    required this.reason,
    required this.message,
  });
  final String id;
  final String reason;
  final String message;
}

enum RunEnd {
  success,

  /// The run paused on a Human Input question; it did not fail.
  interrupt,
  error,

  /// A person stopped the run (the server's `stopped` code), not a lost
  /// connection.
  stopped,

  /// The observer could not follow the run to its end (the server's
  /// `butter.fallback` marker); the run itself went on.
  fallback,
}

@immutable
final class RunOutcome {
  const RunOutcome(
    this.end, {
    this.message,
    this.interrupts = const [],
    this.fallbackReason,
  });
  final RunEnd end;
  final String? message;
  final List<InterruptInfo> interrupts;
  final String? fallbackReason;
}

/// [text] cut to at most [max] characters, with an ellipsis when cut.
String boundText(String text, int max) =>
    text.length <= max ? text : '${text.substring(0, max)}…';
