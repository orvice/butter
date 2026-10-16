import 'package:butter_mobile/core/cron_errors.dart';
import 'package:connectrpc/connect.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  test('maps CHANNEL rejection to actionable copy', () {
    final msg = cronFailureMessage(
      ConnectException(
        Code.invalidArgument,
        'CHANNEL delivery is no longer accepted for new or updated jobs',
      ),
      CronOperation.update,
    );
    expect(msg, contains('Legacy CHANNEL'));
  });

  test('run outcome unknown on unavailable', () {
    expect(
      cronWriteOutcomeUnknown(ConnectException(Code.unavailable, 'gone')),
      isTrue,
    );
  });

  test('run outcome known on validation errors', () {
    expect(
      cronWriteOutcomeUnknown(
        ConnectException(Code.invalidArgument, 'bad job'),
      ),
      isFalse,
    );
  });
}
