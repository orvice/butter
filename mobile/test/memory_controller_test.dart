import 'package:butter_mobile/features/memory/memory_controller.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  test('deleteMemory refuses when canManage is false', () async {
    final c = MemoryController(
      newApi: () => throw UnimplementedError(),
      token: 't',
      workspaceId: 'ws',
      canManage: () => false,
      onUnauthenticated: () {},
    );
    addTearDown(c.dispose);
    expect(await c.deleteMemory('x'), isFalse);
    expect(c.writeError, isNotNull);
  });
}
