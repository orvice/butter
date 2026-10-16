import 'protocol.dart';

/// In-memory surface built from A2UI envelopes.
class ProcessedSurface {
  ProcessedSurface({
    required this.surfaceId,
    this.rootId,
    Map<String, Map<String, dynamic>>? components,
    Map<String, dynamic>? data,
    this.deleted = false,
  }) : components = components ?? {},
       data = data ?? {};

  final String surfaceId;
  String? rootId;
  final Map<String, Map<String, dynamic>> components;
  Map<String, dynamic> data;
  bool deleted;

  ProcessedSurface copy() => ProcessedSurface(
    surfaceId: surfaceId,
    rootId: rootId,
    components: Map.fromEntries(
      components.entries.map((e) => MapEntry(e.key, Map.of(e.value))),
    ),
    data: _deepCopyMap(data),
    deleted: deleted,
  );
}

Map<String, dynamic> _deepCopyMap(Map<String, dynamic> source) {
  final out = <String, dynamic>{};
  for (final entry in source.entries) {
    final v = entry.value;
    if (v is Map) {
      out[entry.key] = _deepCopyMap(Map<String, dynamic>.from(v));
    } else if (v is List) {
      out[entry.key] = List.of(v);
    } else {
      out[entry.key] = v;
    }
  }
  return out;
}

/// Applies one envelope to [surface]. Returns an error message on failure.
String? applyEnvelope(ProcessedSurface surface, Map<String, dynamic> env) {
  final op = envelopeOp(env);
  switch (op) {
    case 'createSurface':
      final create = env['createSurface'];
      if (create is! Map) return 'invalid createSurface';
      surface.deleted = false;
      return null;
    case 'updateComponents':
      final update = env['updateComponents'];
      if (update is! Map) return 'invalid updateComponents';
      final list = update['components'];
      if (list is! List) return 'invalid components';
      for (final raw in list) {
        if (raw is! Map) continue;
        final id = raw['id'];
        if (id is! String || id.isEmpty) continue;
        surface.components[id] = Map<String, dynamic>.from(raw);
        if (raw['component'] == 'root' ||
            (surface.rootId == null && id == 'root')) {
          surface.rootId = id;
        }
      }
      return null;
    case 'updateDataModel':
      final update = env['updateDataModel'];
      if (update is! Map) return 'invalid updateDataModel';
      final path = update['path'];
      if (path is! String) return 'invalid path';
      _setAtPath(surface.data, path, update['value']);
      return null;
    case 'deleteSurface':
      surface.deleted = true;
      surface.components.clear();
      surface.data = {};
      surface.rootId = null;
      return null;
    default:
      return 'unknown envelope';
  }
}

void _setAtPath(Map<String, dynamic> root, String path, Object? value) {
  if (path == '/' || path.isEmpty) {
    if (value is Map) {
      root
        ..clear()
        ..addAll(Map<String, dynamic>.from(value));
    }
    return;
  }
  if (!path.startsWith('/')) return;
  final parts = path.split('/').where((p) => p.isNotEmpty).toList();
  if (parts.isEmpty) return;
  var current = root;
  for (var i = 0; i < parts.length - 1; i++) {
    final key = parts[i];
    final next = current[key];
    if (next is! Map<String, dynamic>) {
      final created = <String, dynamic>{};
      current[key] = created;
      current = created;
    } else {
      current = next;
    }
  }
  current[parts.last] = value;
}

Object? readAtPath(Map<String, dynamic> root, String path) {
  if (path == '/' || path.isEmpty) return root;
  if (!path.startsWith('/')) return null;
  final parts = path.split('/').where((p) => p.isNotEmpty).toList();
  Object? current = root;
  for (final part in parts) {
    if (current is! Map) return null;
    current = current[part];
  }
  return current;
}

/// Resolves a dynamic string property (literal or {"path": "/…"}).
String resolveString(Object? prop, Map<String, dynamic> data) {
  if (prop is String) return prop;
  if (prop is Map && prop['path'] is String) {
    final v = readAtPath(data, prop['path'] as String);
    if (v == null) return '';
    return v.toString();
  }
  return prop == null ? '' : prop.toString();
}

ProcessedSurface rebuildFromEnvelopes(
  String surfaceId,
  List<Map<String, dynamic>> envelopes,
) {
  final surface = ProcessedSurface(surfaceId: surfaceId);
  for (final env in envelopes) {
    applyEnvelope(surface, env);
  }
  return surface;
}
