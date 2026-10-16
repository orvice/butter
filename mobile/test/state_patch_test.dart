import 'package:butter_mobile/core/state_patch.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  test('top-level replace updates the mirror', () {
    final next = applyAguiStateDelta(
      {'a': 1},
      [
        {'op': 'replace', 'path': '/a', 'value': 2},
      ],
    );
    expect(next, {'a': 2});
  });

  test('nested paths are refused', () {
    expect(
      applyAguiStateDelta({}, [
        {'op': 'add', 'path': '/a/b', 'value': 1},
      ]),
      isNull,
    );
  });
}
