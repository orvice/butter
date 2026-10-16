import 'package:butter_mobile/core/automation_errors.dart';
import 'package:connectrpc/connect.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  test('run failure does not map to auto-retry hint', () {
    final msg = automationFailureMessage(
      ConnectException(Code.unavailable, 'gone'),
      AutomationOperation.run,
    );
    expect(msg, contains('connection'));
  });
}
