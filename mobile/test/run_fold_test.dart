import 'dart:convert';

import 'package:butter_mobile/features/chat/chat_model.dart';
import 'package:butter_mobile/features/chat/run_fold.dart';
import 'package:flutter_test/flutter_test.dart';

RunFold fold(Iterable<Map<String, Object?>> events) {
  final fold = RunFold();
  for (final event in events) {
    fold.addData(jsonEncode(event));
  }
  return fold;
}

Map<String, Object?> started([String run = 'run-1']) => {
  'type': 'RUN_STARTED',
  'threadId': 't-1',
  'runId': run,
};
Map<String, Object?> textStart(String id) => {
  'type': 'TEXT_MESSAGE_START',
  'messageId': id,
  'role': 'assistant',
};
Map<String, Object?> text(String id, String delta) => {
  'type': 'TEXT_MESSAGE_CONTENT',
  'messageId': id,
  'delta': delta,
};
Map<String, Object?> textEnd(String id) => {
  'type': 'TEXT_MESSAGE_END',
  'messageId': id,
};
Map<String, Object?> finished([Map<String, Object?>? outcome]) => {
  'type': 'RUN_FINISHED',
  'threadId': 't-1',
  'runId': 'run-1',
  'outcome': outcome ?? {'type': 'success'},
};

void main() {
  test('text deltas build one reply and the run ends in success', () {
    final f = fold([
      started(),
      {'type': 'STATE_SNAPSHOT', 'snapshot': {}},
      textStart('m1'),
      text('m1', '你好, '),
      text('m1', '**world**'),
      textEnd('m1'),
      finished(),
    ]);
    expect(f.runId, 'run-1');
    expect(f.reply.role, ChatRole.assistant);
    expect(f.reply.parts, hasLength(1));
    expect((f.reply.parts.single as TextPart).text, '你好, **world**');
    expect(f.outcome?.end, RunEnd.success);
  });

  test('the reply is a live snapshot while the run goes on', () {
    final f = fold([started(), textStart('m1'), text('m1', 'par')]);
    expect(f.outcome, isNull);
    expect((f.reply.parts.single as TextPart).text, 'par');
    f.addData(jsonEncode(text('m1', 'tial')));
    expect((f.reply.parts.single as TextPart).text, 'partial');
  });

  test('a snapshot taken earlier does not change as the run continues', () {
    final f = fold([started(), textStart('m1'), text('m1', 'a')]);
    final before = f.reply;
    f.addData(jsonEncode(text('m1', 'b')));
    expect((before.parts.single as TextPart).text, 'a');
  });

  test('tool calls show their arguments, running state and result', () {
    final f = fold([
      started(),
      {
        'type': 'TOOL_CALL_START',
        'toolCallId': 'c1',
        'toolCallName': 'queryRecords',
        'parentMessageId': 'm1',
      },
      {'type': 'TOOL_CALL_ARGS', 'toolCallId': 'c1', 'delta': '{"q":'},
    ]);
    var tool = f.reply.parts.single as ToolPart;
    expect(tool.name, 'queryRecords');
    expect(tool.status, ToolStatus.calling);
    f.addData(
      jsonEncode({'type': 'TOOL_CALL_ARGS', 'toolCallId': 'c1', 'delta': '1}'}),
    );
    f.addData(jsonEncode({'type': 'TOOL_CALL_END', 'toolCallId': 'c1'}));
    tool = f.reply.parts.single as ToolPart;
    expect(tool.arguments, '{"q":1}');
    expect(tool.status, ToolStatus.running);
    f.addData(
      jsonEncode({
        'type': 'TOOL_CALL_RESULT',
        'messageId': 'result:c1',
        'toolCallId': 'c1',
        'content': '{"rows":2}',
      }),
    );
    tool = f.reply.parts.single as ToolPart;
    expect(tool.status, ToolStatus.done);
    expect(tool.result, '{"rows":2}');
  });

  test('text and tool calls keep the order they arrived in', () {
    final f = fold([
      started(),
      textStart('m1'),
      text('m1', 'Let me look.'),
      textEnd('m1'),
      {'type': 'TOOL_CALL_START', 'toolCallId': 'c1', 'toolCallName': 'search'},
      {'type': 'TOOL_CALL_END', 'toolCallId': 'c1'},
      textStart('m2'),
      text('m2', 'Found it.'),
    ]);
    expect(f.reply.parts.map((p) => p.runtimeType), [
      TextPart,
      ToolPart,
      TextPart,
    ]);
  });

  test('a result for a call the fold never saw is kept, not dropped', () {
    final f = fold([
      started(),
      {
        'type': 'TOOL_CALL_RESULT',
        'messageId': 'r',
        'toolCallId': 'ghost',
        'content': 'x',
      },
    ]);
    final tool = f.reply.parts.single as ToolPart;
    expect(tool.id, 'ghost');
    expect(tool.status, ToolStatus.done);
  });

  test('an interrupt outcome lists the questions and is not a failure', () {
    final f = fold([
      started(),
      finished({
        'type': 'interrupt',
        'interrupts': [
          {'id': 'int-7', 'reason': 'human_input', 'message': 'Approve?'},
        ],
      }),
    ]);
    expect(f.outcome?.end, RunEnd.interrupt);
    expect(f.outcome?.interrupts.single.id, 'int-7');
    expect(f.outcome?.interrupts.single.message, 'Approve?');
  });

  test('a failure carries the server reason', () {
    final f = fold([
      started(),
      {'type': 'RUN_ERROR', 'message': 'model unavailable', 'runId': 'run-1'},
    ]);
    expect(f.outcome?.end, RunEnd.error);
    expect(f.outcome?.message, 'model unavailable');
  });

  test('the stop code is told apart from a failure', () {
    final f = fold([
      started(),
      {
        'type': 'RUN_ERROR',
        'code': 'stopped',
        'message': 'stopped by user',
        'runId': 'run-1',
      },
    ]);
    expect(f.outcome?.end, RunEnd.stopped);
  });

  test('the fallback marker is neither success nor failure', () {
    final f = fold([
      started(),
      textStart('m1'),
      text('m1', 'partial'),
      {
        'type': 'CUSTOM',
        'name': 'butter.fallback',
        'value': {'threadId': 't-1', 'runId': 'run-1', 'reason': 'truncated'},
      },
    ]);
    expect(f.outcome?.end, RunEnd.fallback);
    expect(f.outcome?.fallbackReason, 'truncated');
    expect((f.reply.parts.single as TextPart).text, 'partial');
  });

  test('other CUSTOM events, state events and unknown types are ignored', () {
    final f = fold([
      started(),
      {
        'type': 'CUSTOM',
        'name': 'butter.a2ui',
        'value': {'x': 1},
      },
      {'type': 'STATE_DELTA', 'delta': []},
      {'type': 'SOMETHING_NEW', 'extra': true},
      textStart('m1'),
      text('m1', 'ok'),
    ]);
    expect(f.reply.parts, hasLength(1));
    expect(f.outcome, isNull);
  });

  test('nothing after the end of the run changes the result', () {
    final f = fold([
      started(),
      textStart('m1'),
      text('m1', 'done'),
      finished(),
      text('m1', ' extra'),
      {'type': 'RUN_ERROR', 'message': 'late'},
    ]);
    expect(f.outcome?.end, RunEnd.success);
    expect((f.reply.parts.single as TextPart).text, 'done');
  });

  test('an unreadable event is a format error', () {
    expect(() => RunFold().addData('{not json'), throwsFormatException);
    expect(() => RunFold().addData('[1]'), throwsFormatException);
  });

  test('a RUN_STARTED for another run is refused', () {
    final f = fold([started('run-1')]);
    expect(
      () => f.addData(jsonEncode(started('run-2'))),
      throwsFormatException,
    );
  });
}
