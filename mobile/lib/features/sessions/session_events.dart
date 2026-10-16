import 'dart:convert';

import 'package:protobuf/well_known_types/google/protobuf/duration.pb.dart'
    as pbd;

import '../../gen/agents/v1/agent_service.pb.dart';

class ParsedTextPart {
  const ParsedTextPart(this.text);
  final String text;
}

class ParsedFunctionCall {
  const ParsedFunctionCall({required this.name, required this.args});
  final String name;
  final Object? args;
}

class ParsedFunctionResponse {
  const ParsedFunctionResponse({required this.name, required this.response});
  final String name;
  final Object? response;
}

/// One session event with content parsed from [SessionEvent.contentJson].
class ParsedSessionEvent {
  ParsedSessionEvent({
    required this.eventId,
    required this.author,
    required this.textParts,
    required this.functionCalls,
    required this.functionResponses,
    this.timestamp,
    this.traceUrl,
    this.traceId,
    required this.raw,
  });

  final String eventId;
  final String author;
  final List<ParsedTextPart> textParts;
  final List<ParsedFunctionCall> functionCalls;
  final List<ParsedFunctionResponse> functionResponses;
  final DateTime? timestamp;
  final String? traceUrl;
  final String? traceId;
  final SessionEvent raw;
}

ParsedSessionEvent parseSessionEvent(SessionEvent evt) {
  final author = evt.author.isEmpty ? 'unknown' : evt.author;
  final out = ParsedSessionEvent(
    eventId: evt.eventId,
    author: author,
    textParts: const [],
    functionCalls: const [],
    functionResponses: const [],
    timestamp: evt.hasTimestamp() ? evt.timestamp.toDateTime().toLocal() : null,
    traceUrl: evt.traceUrl.isEmpty ? null : evt.traceUrl,
    traceId: evt.traceId.isEmpty ? null : evt.traceId,
    raw: evt,
  );
  if (evt.contentJson.isEmpty) return out;

  Map<String, dynamic>? content;
  try {
    content = jsonDecode(evt.contentJson) as Map<String, dynamic>;
  } catch (_) {
    return ParsedSessionEvent(
      eventId: out.eventId,
      author: out.author,
      textParts: [ParsedTextPart(evt.contentJson)],
      functionCalls: out.functionCalls,
      functionResponses: out.functionResponses,
      timestamp: out.timestamp,
      traceUrl: out.traceUrl,
      traceId: out.traceId,
      raw: out.raw,
    );
  }

  final texts = <ParsedTextPart>[];
  final calls = <ParsedFunctionCall>[];
  final responses = <ParsedFunctionResponse>[];
  final parts = content['parts'];
  if (parts is List) {
    for (final part in parts) {
      if (part is! Map) continue;
      final text = part['text'];
      if (text is String && text.isNotEmpty) {
        texts.add(ParsedTextPart(text));
      }
      final call = part['functionCall'] ?? part['function_call'];
      if (call is Map && call['name'] is String) {
        calls.add(
          ParsedFunctionCall(name: call['name'] as String, args: call['args']),
        );
      }
      final resp = part['functionResponse'] ?? part['function_response'];
      if (resp is Map && resp['name'] is String) {
        responses.add(
          ParsedFunctionResponse(
            name: resp['name'] as String,
            response: resp['response'],
          ),
        );
      }
    }
  }
  return ParsedSessionEvent(
    eventId: out.eventId,
    author: out.author,
    textParts: texts,
    functionCalls: calls,
    functionResponses: responses,
    timestamp: out.timestamp,
    traceUrl: out.traceUrl,
    traceId: out.traceId,
    raw: out.raw,
  );
}

List<ParsedSessionEvent> parseSessionEvents(List<SessionEvent> events) =>
    events.map(parseSessionEvent).toList(growable: false);

/// Formats [detail.duration] from GetSession.
String formatSessionDetailDuration(pbd.Duration? duration) {
  if (duration == null) return '—';
  final secs = duration.seconds.toInt() + duration.nanos / 1e9;
  if (secs < 60) return '${secs.toStringAsFixed(1)} s';
  final mins = secs ~/ 60;
  final rem = (secs % 60).floor();
  return '${mins}m ${rem}s';
}

/// Formats a protobuf Duration string such as `12.345s`.
String formatSessionDuration(String? raw) {
  if (raw == null || raw.isEmpty) return '—';
  final match = RegExp(r'^(-?\d+(?:\.\d+)?)s$').firstMatch(raw);
  if (match == null) return raw;
  final secs = double.parse(match.group(1)!);
  if (secs < 60) return '${secs.toStringAsFixed(1)} s';
  final mins = secs ~/ 60;
  final rem = (secs % 60).floor();
  return '${mins}m ${rem}s';
}
