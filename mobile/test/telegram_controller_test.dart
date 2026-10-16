import 'package:butter_mobile/core/telegram_errors.dart';
import 'package:connectrpc/connect.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  test('maps webhook precondition to administrator hint', () {
    final msg = telegramFailureMessage(
      ConnectException(
        Code.failedPrecondition,
        'Telegram webhook base URL is not configured',
      ),
      TelegramOperation.createChannel,
    );
    expect(msg, contains('platform administrator'));
  });

  test('maps resend already delivered', () {
    final msg = telegramFailureMessage(
      ConnectException(Code.failedPrecondition, 'already delivered'),
      TelegramOperation.resend,
    );
    expect(msg, contains('already delivered'));
  });
}
