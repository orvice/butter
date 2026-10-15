import 'dart:async';
import 'dart:convert';

/// One dispatched Server-Sent Event.
final class SseEvent {
  const SseEvent({required this.event, required this.data, this.id});

  /// The `event:` name, `message` when the event has none.
  final String event;

  /// The `data:` lines joined with a newline.
  final String data;
  final String? id;
}

/// A single event outgrew the buffer bound. The stream is abandoned rather
/// than held in memory.
final class SseOverflow implements Exception {
  const SseOverflow(this.limit);
  final int limit;
  @override
  String toString() => 'SseOverflow: one event exceeds $limit characters';
}

/// Parses a `text/event-stream` body.
///
/// Bytes may be split anywhere, including inside a multi-byte character or
/// between the CR and LF of one line ending. Lines end with LF, CRLF or CR.
/// Comment lines (the server's `: heartbeat`) are skipped. An event without
/// data is not dispatched, and one that never ends before the source does is
/// dropped, as the SSE processing model says. The text of one event (its
/// pending line plus its data lines) is bounded by [maxEventLength]
/// characters, so a hostile or broken stream cannot grow the buffer without
/// limit; malformed UTF-8 is a [FormatException], never replacement text.
Stream<SseEvent> parseSse(
  Stream<List<int>> bytes, {
  int maxEventLength = 8 * 1024 * 1024,
}) async* {
  final line = StringBuffer();
  var lineLength = 0;
  final data = <String>[];
  var dataLength = 0;
  var event = '';
  String? id;
  var skipLf = false;
  var first = true;

  SseEvent? endLine() {
    final text = line.toString();
    line.clear();
    lineLength = 0;
    if (text.isEmpty) {
      // A blank line dispatches the event.
      final name = event.isEmpty ? 'message' : event;
      final hadData = data.isNotEmpty;
      final joined = data.join('\n');
      final lastId = id;
      data.clear();
      dataLength = 0;
      event = '';
      id = null;
      return hadData ? SseEvent(event: name, data: joined, id: lastId) : null;
    }
    if (text.startsWith(':')) return null;
    final colon = text.indexOf(':');
    final field = colon < 0 ? text : text.substring(0, colon);
    var value = colon < 0 ? '' : text.substring(colon + 1);
    if (value.startsWith(' ')) value = value.substring(1);
    switch (field) {
      case 'data':
        dataLength += value.length + 1;
        if (dataLength > maxEventLength) throw SseOverflow(maxEventLength);
        data.add(value);
      case 'event':
        event = value;
      case 'id':
        if (!value.contains('\u0000')) id = value;
    }
    return null;
  }

  await for (var chunk in bytes.transform(utf8.decoder)) {
    if (first) {
      first = false;
      if (chunk.startsWith('\uFEFF')) chunk = chunk.substring(1);
    }
    for (var i = 0; i < chunk.length; i++) {
      final unit = chunk.codeUnitAt(i);
      if (unit == 0x0A) {
        if (skipLf) {
          skipLf = false;
          continue;
        }
      } else if (unit == 0x0D) {
        skipLf = true;
      } else {
        skipLf = false;
        line.writeCharCode(unit);
        lineLength++;
        if (lineLength + dataLength > maxEventLength) {
          throw SseOverflow(maxEventLength);
        }
        continue;
      }
      final dispatched = endLine();
      if (dispatched != null) yield dispatched;
    }
  }
}
