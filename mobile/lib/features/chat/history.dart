import 'dart:convert';
import 'dart:typed_data';

import '../../core/avatar.dart';
import 'chat_model.dart';
import 'run_request.dart';

/// The run holding a thread, as the read names it.
final class RunningRun {
  const RunningRun({required this.runId, required this.invocationId});
  final String runId;
  final String invocationId;
}

/// How the thread's latest run ended when it did not succeed.
final class LastRun {
  const LastRun({required this.status, required this.error, this.input});

  /// `failed`, or `cancelled` when a person stopped the run.
  final String status;
  final String error;

  /// The text the user sent, as the server kept it; null when the turn had
  /// none (a form answer, tool results, only images).
  final String? input;
}

/// A thread read: the conversation and what holds it.
final class ThreadHistory {
  ThreadHistory({
    required Iterable<ChatMessage> messages,
    required Iterable<InterruptInfo> interrupts,
    this.running,
    this.lastRun,
  }) : messages = List.unmodifiable(messages),
       interrupts = List.unmodifiable(interrupts);
  final List<ChatMessage> messages;
  final List<InterruptInfo> interrupts;
  final RunningRun? running;
  final LastRun? lastRun;
}

String? _string(Object? value) => value is String ? value : null;

/// Bytes a base64 string decodes to, or null when it is not well-formed base64
/// of whole groups.
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

/// Reads an image of a user turn the way sending reads one: a whitelisted
/// format by its bytes, at most 10 MiB, at most 10 per turn and 20 MiB
/// together. Anything else is shown as unavailable, never decoded in full.
class _ImageBudget {
  var count = 0;
  var bytes = 0;

  MessagePart read(String? mimeType, String? data) {
    if (data == null || mimeType == null) return const UnavailableImagePart();
    final size = _decodedLength(data);
    if (size == null || size == 0 || size > maxImageBytes) {
      return const UnavailableImagePart();
    }
    if (count >= maxImagesPerMessage || bytes + size > maxMessageBytes) {
      return const UnavailableImagePart();
    }
    final Uint8List head;
    try {
      head = base64Decode(
        data.substring(0, data.length < 16 ? data.length : 16),
      );
    } on FormatException {
      return const UnavailableImagePart();
    }
    final sniffed = sniffImageContentType(head);
    if (sniffed == null) return const UnavailableImagePart();
    count++;
    bytes += size;
    return ImagePart(mimeType: sniffed, base64: data);
  }
}

List<MessagePart>? _userParts(Object? content) {
  if (content is String) return [if (content.isNotEmpty) TextPart(content)];
  if (content is! List) return null;
  final budget = _ImageBudget();
  final parts = <MessagePart>[];
  for (final item in content) {
    if (item is! Map) continue;
    switch (item['type']) {
      case 'text':
        final text = _string(item['text']);
        if (text != null && text.isNotEmpty) parts.add(TextPart(text));
      case 'image':
        final source = item['source'];
        parts.add(
          source is Map && source['type'] == 'data'
              ? budget.read(
                  _string(source['mimeType']),
                  _string(source['value']),
                )
              : const UnavailableImagePart(),
        );
      case 'binary':
        parts.add(
          budget.read(_string(item['mimeType']), _string(item['data'])),
        );
    }
  }
  return parts;
}

/// Parses `GET /api/agui/:agent_id/threads/:thread_id/messages`.
///
/// Each run is one assistant message whose tool calls are followed by `tool`
/// messages with their results; those are merged into the calls. Roles and
/// fields this App does not know are skipped, so a newer server does not
/// break the thread.
ThreadHistory parseThreadHistory(String body) {
  final Object? decoded;
  try {
    decoded = jsonDecode(body);
  } on FormatException {
    throw const FormatException('Unreadable thread from the server.');
  }
  if (decoded is! Map || decoded['messages'] is! List) {
    throw const FormatException('Unreadable thread from the server.');
  }
  final rawMessages = decoded['messages'] as List;

  final results = <String, String>{};
  for (final raw in rawMessages) {
    if (raw is Map && raw['role'] == 'tool') {
      final id = _string(raw['toolCallId']);
      if (id != null) results[id] = _string(raw['content']) ?? '';
    }
  }

  final messages = <ChatMessage>[];
  for (final raw in rawMessages) {
    if (raw is! Map) continue;
    final id = _string(raw['id']) ?? '';
    switch (raw['role']) {
      case 'user':
        final parts = _userParts(raw['content']);
        if (parts != null) {
          messages.add(ChatMessage(id: id, role: ChatRole.user, parts: parts));
        }
      case 'assistant':
        final text = _string(raw['content']) ?? '';
        final calls = raw['toolCalls'];
        messages.add(
          ChatMessage(
            id: id,
            role: ChatRole.assistant,
            parts: [
              if (text.isNotEmpty) TextPart(text),
              if (calls is List)
                for (final call in calls)
                  if (call is Map && _string(call['id']) != null)
                    () {
                      final callId = call['id'] as String;
                      final function = call['function'];
                      final result = results[callId];
                      return ToolPart(
                        id: callId,
                        name: function is Map
                            ? _string(function['name']) ?? ''
                            : '',
                        arguments: function is Map
                            ? _string(function['arguments']) ?? ''
                            : '',
                        result: result,
                        status: result == null
                            ? ToolStatus.running
                            : ToolStatus.done,
                      );
                    }(),
            ],
          ),
        );
    }
  }

  final interrupts = <InterruptInfo>[
    if (decoded['interrupts'] is List)
      for (final item in decoded['interrupts'] as List)
        if (item is Map && _string(item['id']) != null)
          InterruptInfo(
            id: item['id'] as String,
            reason: _string(item['reason']) ?? '',
            message: _string(item['message']) ?? '',
          ),
  ];

  final running = decoded['running'];
  final last = decoded['lastRun'];
  return ThreadHistory(
    messages: messages,
    interrupts: interrupts,
    running: running is Map && _string(running['runId']) != null
        ? RunningRun(
            runId: running['runId'] as String,
            invocationId: _string(running['invocationId']) ?? '',
          )
        : null,
    lastRun: last is Map && _string(last['status']) != null
        ? LastRun(
            status: last['status'] as String,
            error: _string(last['error']) ?? '',
            input: _string(last['input']),
          )
        : null,
  );
}
