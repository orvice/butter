import 'package:butter_mobile/features/sessions/session_events.dart';
import 'package:butter_mobile/gen/agents/v1/agent_service.pb.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:protobuf/well_known_types/google/protobuf/duration.pb.dart'
    as pbd;
import 'package:fixnum/fixnum.dart';

void main() {
  test('parseSessionEvent reads text and tool calls', () {
    final evt = SessionEvent(
      eventId: 'e1',
      author: 'user',
      contentJson: '''
{"parts":[{"text":"hi"},{"functionCall":{"name":"tool","args":{"k":1}}}]}
''',
    );
    final parsed = parseSessionEvent(evt);
    expect(parsed.textParts.single.text, 'hi');
    expect(parsed.functionCalls.single.name, 'tool');
  });

  test('formatSessionDetailDuration', () {
    expect(formatSessionDetailDuration(null), '—');
    expect(
      formatSessionDetailDuration(pbd.Duration(seconds: Int64(90))),
      '1m 30s',
    );
  });
}
