// butter-basic-v1 component names (internal/a2ui/catalog.go).

const modelComponents = {
  'Column',
  'Row',
  'Text',
  'Card',
  'Divider',
  'KeyValue',
  'Status',
};

const formOnlyComponents = {'TextField', 'ChoicePicker', 'Button'};

const allCatalogComponents = {...modelComponents, ...formOnlyComponents};

/// Names the first component outside the catalog in [envelopes], if any.
String? unknownComponentInEnvelopes(List<Map<String, dynamic>> envelopes) {
  for (final env in envelopes) {
    final update = env['updateComponents'];
    if (update is! Map) continue;
    final components = update['components'];
    if (components is! List) continue;
    for (final c in components) {
      if (c is! Map) continue;
      final name = c['component'];
      if (name is! String || !allCatalogComponents.contains(name)) {
        return name is String && name.isNotEmpty ? name : '(unnamed)';
      }
    }
  }
  return null;
}

/// True when a read-only card envelope tries to use form-only components.
bool cardUsesFormComponents(List<Map<String, dynamic>> envelopes) {
  for (final env in envelopes) {
    final update = env['updateComponents'];
    if (update is! Map) continue;
    final components = update['components'];
    if (components is! List) continue;
    for (final c in components) {
      if (c is! Map) continue;
      final name = c['component'];
      if (name is String && formOnlyComponents.contains(name)) return true;
    }
  }
  return false;
}
