/// Applies RFC 6902 operations to the client's AG-UI state mirror.
///
/// Butter emits top-level add/replace/remove ops; unsupported shapes return
/// null so the caller can treat the delta as failed (never silent success).
Map<String, dynamic>? applyAguiStateDelta(
  Map<String, dynamic> state,
  List<dynamic> ops,
) {
  final next = Map<String, dynamic>.from(state);
  for (final raw in ops) {
    if (raw is! Map) return null;
    final path = raw['path'];
    if (path is! String || !path.startsWith('/') || path.indexOf('/', 1) >= 0) {
      return null;
    }
    final key = _unescapeJsonPointer(path.substring(1));
    switch (raw['op']) {
      case 'add':
      case 'replace':
        next[key] = raw['value'];
      case 'remove':
        next.remove(key);
      default:
        return null;
    }
  }
  return next;
}

String _unescapeJsonPointer(String segment) =>
    segment.split('~1').join('/').split('~0').join('~');
