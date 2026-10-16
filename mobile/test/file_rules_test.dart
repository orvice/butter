import 'package:butter_mobile/features/files/file_rules.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  // The expectations are what the server's agentfile.NormalizePath answers
  // (internal/repo/agentfile/path.go); files_contract_test holds this mirror
  // to the running handlers for the same inputs.
  test('paths are rooted and cleaned the way the server does', () {
    expect(normalizeFilePath('a'), '/a');
    expect(normalizeFilePath('  /a//b/../c  '), '/a/c');
    expect(normalizeFilePath('/a/./b'), '/a/b');
    expect(normalizeFilePath('notes/todo.md/'), '/notes/todo.md');
    expect(normalizeFilePath('/../etc'), '/etc');
  });

  test('paths the server refuses are refused', () {
    for (final bad in ['', '   ', '/', '//', '.', '/a/..', '..', 'a\u0000b']) {
      expect(normalizeFilePath(bad), isNull, reason: 'input "$bad"');
    }
  });

  test('a prefix that is empty or the root means every file', () {
    expect(normalizeFilePrefix(''), '');
    expect(normalizeFilePrefix(' / '), '');
    expect(normalizeFilePrefix('notes/'), '/notes');
    expect(normalizeFilePrefix('/..'), isNull);
  });

  test('the size is the UTF-8 byte count, not the character count', () {
    expect(utf8Length('abc'), 3);
    expect(utf8Length('文'), 3);
    expect(utf8Length('😀'), 4);
    expect(utf8Length(''), 0);
  });

  test('mount paths normalize like the runtime toolset', () {
    expect(normalizeMountPath(''), '/');
    expect(normalizeMountPath('   '), '/');
    expect(normalizeMountPath('docs'), '/docs');
    expect(normalizeMountPath('/a//b/'), '/a/b');
    expect(normalizeMountPath('/'), '/');
    expect(normalizeMountPath('/a/../b'), '/b');
  });

  test('a suggested mount path is a slug that avoids taken ones', () {
    expect(suggestMountPath('Team Docs!', const []), '/team-docs');
    expect(suggestMountPath('Team Docs', const ['/team-docs']), '/team-docs-2');
    expect(
      suggestMountPath('Team Docs', const ['/team-docs', '/team-docs-2']),
      '/team-docs-3',
    );
    expect(suggestMountPath('!!!', const []), '/files');
  });

  test('sizes are shown in plain units', () {
    expect(formatBytes(12), '12 B');
    expect(formatBytes(2048), '2.0 KiB');
    expect(formatBytes(3 * 1024 * 1024), '3.0 MiB');
  });
}
