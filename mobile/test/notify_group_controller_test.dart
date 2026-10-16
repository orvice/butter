import 'package:butter_mobile/core/notify_group_errors.dart';
import 'package:connectrpc/connect.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  test('maps legacy telegram field rejection', () {
    final msg = notifyGroupFailureMessage(
      ConnectException(
        Code.invalidArgument,
        'notify_group.targets[0].telegram.bot_token is no longer accepted',
      ),
      NotifyGroupOperation.create,
    );
    expect(msg, contains('Legacy raw Telegram'));
  });
}
