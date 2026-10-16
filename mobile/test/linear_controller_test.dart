import 'package:butter_mobile/core/linear_errors.dart';
import 'package:connectrpc/connect.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  test('revision conflict maps to reload message', () {
    expect(
      linearFailureMessage(
        ConnectException(Code.aborted, 'revision'),
        LinearOperation.update,
      ),
      contains('Reload'),
    );
  });

  test('ambiguous write flagged unknown', () {
    expect(
      linearWriteOutcomeUnknown(ConnectException(Code.unavailable, 'gone')),
      isTrue,
    );
  });
}
