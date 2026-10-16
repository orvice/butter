import 'dart:convert';

import '../../core/state_patch.dart';
import '../a2ui/protocol.dart';
import 'chat_model.dart';
import 'run_effects.dart';

const _maxMessage = 2000;
const _a2uiEvent = 'butter.a2ui';
const _fallbackEvent = 'butter.fallback';

String _bounded(String text) => boundText(text, _maxMessage);

class _Tool {
  _Tool(this.id, this.name);
  final String id;
  String name;
  final arguments = StringBuffer();
  String? result;
  var status = ToolStatus.calling;
}

/// Folds one run's AG-UI events into a snapshot of the assistant's reply.
///
/// [RunEffects] receives state and A2UI side channels for #462. Attach replay
/// resets the fold on `RUN_STARTED`, never appending to a stale partial reply.
class RunFold {
  RunFold([this._effects = const NullRunEffects()]);

  final RunEffects _effects;
  String? _runId;
  RunOutcome? _outcome;
  Map<String, dynamic>? _state;
  final _parts = <Object>[];
  final _text = <String, StringBuffer>{};
  final _tools = <String, _Tool>{};
  String? _firstMessageId;

  String? get runId => _runId;
  RunOutcome? get outcome => _outcome;
  Map<String, dynamic>? get sharedState =>
      _state == null ? null : Map<String, dynamic>.from(_state!);

  ChatMessage get reply => ChatMessage(
    id: _firstMessageId ?? 'run:${_runId ?? 'pending'}',
    role: ChatRole.assistant,
    parts: [
      for (final part in _parts)
        switch (part) {
          StringBuffer() => TextPart(part.toString()),
          _Tool() => ToolPart(
            id: part.id,
            name: part.name,
            arguments: part.arguments.toString(),
            result: part.result,
            status: part.status,
          ),
          _ => throw StateError('unreachable'),
        },
    ],
  );

  void addData(String data) {
    final Object? decoded;
    try {
      decoded = jsonDecode(data);
    } on FormatException {
      throw const FormatException('Unreadable event from the server.');
    }
    if (decoded is! Map<String, dynamic>) {
      throw const FormatException('Unreadable event from the server.');
    }
    add(decoded);
  }

  void add(Map<String, dynamic> event) {
    final type = event['type'];
    if (type == 'RUN_STARTED') {
      final run = event['runId'];
      if (run is! String) return;
      if (_runId != null && run != _runId) {
        throw const FormatException('The stream switched to another run.');
      }
      _resetForRun(run);
      return;
    }
    if (_outcome != null) return;
    switch (type) {
      case 'TEXT_MESSAGE_START':
        final id = event['messageId'];
        if (id is String) _openText(id);
      case 'TEXT_MESSAGE_CONTENT':
        final id = event['messageId'];
        final delta = event['delta'];
        if (id is String && delta is String) {
          (_text[id] ?? _openText(id)).write(delta);
        }
      case 'TOOL_CALL_START':
        final id = event['toolCallId'];
        if (id is String) {
          final name = event['toolCallName'];
          _tool(id).name = name is String ? name : '';
        }
      case 'TOOL_CALL_ARGS':
        final id = event['toolCallId'];
        final delta = event['delta'];
        if (id is String && delta is String) _tool(id).arguments.write(delta);
      case 'TOOL_CALL_END':
        final id = event['toolCallId'];
        if (id is String) {
          final tool = _tool(id);
          if (tool.status == ToolStatus.calling) {
            tool.status = ToolStatus.running;
          }
        }
      case 'TOOL_CALL_RESULT':
        final id = event['toolCallId'];
        if (id is String) {
          final tool = _tool(id);
          final content = event['content'];
          tool.result = content is String ? content : '';
          tool.status = ToolStatus.done;
        }
      case 'STATE_SNAPSHOT':
        final snapshot = event['snapshot'];
        if (snapshot is Map) {
          _state = Map<String, dynamic>.from(snapshot);
          _effects.state(_state);
        }
      case 'STATE_DELTA':
        final delta = event['delta'];
        if (delta is List && _state != null) {
          final next = applyAguiStateDelta(_state!, delta);
          if (next == null) {
            throw const FormatException('Unsupported state delta.');
          }
          _state = next;
          _effects.state(_state);
        } else if (delta is List && _state == null) {
          throw const FormatException('State delta before snapshot.');
        }
      case 'CUSTOM':
        final name = event['name'];
        if (name == _fallbackEvent) {
          final value = event['value'];
          final reason = value is Map ? value['reason'] : null;
          _outcome = RunOutcome(
            RunEnd.fallback,
            fallbackReason: reason is String ? reason : null,
          );
        } else if (name == _a2uiEvent && isA2uiEventValue(event['value'])) {
          _effects.a2ui(event['value']);
        }
      case 'RUN_FINISHED':
        _outcome = _finished(event['outcome']);
      case 'RUN_ERROR':
        final message = event['message'];
        _outcome = RunOutcome(
          event['code'] == 'stopped' ? RunEnd.stopped : RunEnd.error,
          message: message is String ? _bounded(message) : null,
        );
    }
  }

  void _resetForRun(String runId) {
    _runId = runId;
    _outcome = null;
    _parts.clear();
    _text.clear();
    _tools.clear();
    _firstMessageId = null;
    _state = null;
  }

  StringBuffer _openText(String id) {
    _firstMessageId ??= id;
    return _text.putIfAbsent(id, () {
      final buffer = StringBuffer();
      _parts.add(buffer);
      return buffer;
    });
  }

  _Tool _tool(String id) => _tools.putIfAbsent(id, () {
    final tool = _Tool(id, '');
    _parts.add(tool);
    return tool;
  });

  RunOutcome _finished(Object? outcome) {
    if (outcome is Map && outcome['type'] == 'interrupt') {
      final raw = outcome['interrupts'];
      return RunOutcome(
        RunEnd.interrupt,
        interrupts: [
          if (raw is List)
            for (final item in raw)
              if (item is Map && item['id'] is String)
                InterruptInfo(
                  id: item['id'] as String,
                  reason: item['reason'] is String
                      ? item['reason'] as String
                      : '',
                  message: item['message'] is String
                      ? _bounded(item['message'] as String)
                      : '',
                ),
        ],
      );
    }
    return const RunOutcome(RunEnd.success);
  }
}
