import 'dart:convert';

/// The server's default per-file limit (`agent_files.max_file_bytes`). The
/// real limit is configurable and is only known from a refusal, so the App
/// warns against this figure but never blocks a save on it.
const defaultFileMaxBytes = 256 * 1024;

/// Files larger than this are not opened in the App: one screen of text this
/// long is neither readable nor safe to hold and diff on a phone.
const editableFileLimit = 1024 * 1024;

const defaultFileContentType = 'text/plain; charset=utf-8';

/// `path.Clean` as Go defines it, for an absolute path.
String _clean(String rooted) {
  final out = <String>[];
  for (final part in rooted.split('/')) {
    if (part.isEmpty || part == '.') continue;
    if (part == '..') {
      if (out.isNotEmpty) out.removeLast();
      continue;
    }
    out.add(part);
  }
  return '/${out.join('/')}';
}

/// What the server's `agentfile.NormalizePath` makes of [input], or null when
/// it refuses it (empty, the root, or containing NUL). The App shows this
/// result before saving; the server remains the authority.
String? normalizeFilePath(String input) {
  var p = input.trim();
  if (p.isEmpty) return null;
  if (!p.startsWith('/')) p = '/$p';
  final clean = _clean(p);
  if (clean == '/' || clean.contains('\u0000')) return null;
  return clean;
}

/// `agentfile.NormalizePrefix`: empty or `/` means every file (returned as the
/// empty string); null means the server would refuse the prefix.
String? normalizeFilePrefix(String input) {
  final p = input.trim();
  if (p.isEmpty || p == '/') return '';
  return normalizeFilePath(p);
}

/// The size the server limits: the UTF-8 bytes of the content.
int utf8Length(String content) => utf8.encode(content).length;

/// How a mount path is stored by the runtime's toolset: an empty path is the
/// root `/`, otherwise rooted and cleaned. Two mounts that normalize to the
/// same path are collapsed by the runtime (the later one silently dropped), so
/// the editor must not allow them.
String normalizeMountPath(String input) {
  final p = input.trim();
  if (p.isEmpty) return '/';
  return _clean(p.startsWith('/') ? p : '/$p');
}

/// A mount path suggestion for a space: its name as a slug, made unique among
/// [taken] (already normalized).
String suggestMountPath(String spaceName, Iterable<String> taken) {
  final slug = spaceName
      .toLowerCase()
      .replaceAll(RegExp(r'[^a-z0-9]+'), '-')
      .replaceAll(RegExp(r'^-+|-+$'), '');
  final base = '/${slug.isEmpty ? 'files' : slug}';
  final used = taken.toSet();
  if (!used.contains(base)) return base;
  for (var i = 2; ; i++) {
    final candidate = '$base-$i';
    if (!used.contains(candidate)) return candidate;
  }
}

String formatBytes(int bytes) {
  if (bytes < 1024) return '$bytes B';
  if (bytes < 1024 * 1024) return '${(bytes / 1024).toStringAsFixed(1)} KiB';
  return '${(bytes / 1024 / 1024).toStringAsFixed(1)} MiB';
}
