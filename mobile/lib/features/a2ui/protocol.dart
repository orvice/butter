import 'dart:convert';

/// Butter's A2UI-over-AG-UI contract (docs/api.md, A2UI surfaces).

const a2uiVersion = 'v0.9.1';
const catalogId = 'butter-basic-v1';
const a2uiEventName = 'butter.a2ui';
const submitEventName = 'butter.submitForm';

/// The forwardedProps declaration that makes a run A2UI-capable.
const a2uiCapability = {
  'version': a2uiVersion,
  'catalogs': [catalogId],
};

/// True when [value] is shaped like the server's butterA2UI negotiation.
bool isValidA2uiCapability(Object? value) {
  if (value is! Map) return false;
  if (value['version'] != a2uiVersion) return false;
  final catalogs = value['catalogs'];
  if (catalogs is! List || catalogs.length != 1) return false;
  return catalogs.single == catalogId;
}

enum SurfaceKind { card, form }

enum FormStatus { pending, submitting, answered, error }

enum SurfaceOrigin { stream, snapshot }

class FormOption {
  const FormOption({required this.value, required this.label});
  final String value;
  final String label;
}

class FormField {
  const FormField({
    required this.name,
    required this.label,
    required this.type,
    this.hint = '',
    this.required = false,
    this.maxLength,
    this.options = const [],
  });
  final String name;
  final String label;
  final String hint;
  final bool required;
  final String type;
  final int? maxLength;
  final List<FormOption> options;
}

/// Binds a server-built form surface to the Interrupt it answers.
class FormView {
  const FormView({
    required this.interruptId,
    required this.token,
    required this.revision,
    required this.title,
    required this.question,
    required this.fields,
  });
  final String interruptId;
  final String token;
  final int revision;
  final String title;
  final String question;
  final List<FormField> fields;
}

class A2uiEventValue {
  A2uiEventValue({
    required this.version,
    required this.surfaceId,
    required this.kind,
    required this.revision,
    required this.seq,
    required this.envelope,
    this.threadId,
    this.runId,
    this.messageId,
    this.fallback,
    this.form,
  });

  final String version;
  final String surfaceId;
  final SurfaceKind kind;
  final int revision;
  final int seq;
  final Map<String, dynamic> envelope;
  final String? threadId;
  final String? runId;
  final String? messageId;
  final String? fallback;
  final FormView? form;
}

class SnapshotSurface {
  SnapshotSurface({
    required this.surfaceId,
    required this.kind,
    required this.revision,
    required this.envelopes,
    this.runId,
    this.messageId,
    this.fallback,
    this.form,
  });
  final String surfaceId;
  final SurfaceKind kind;
  final int revision;
  final List<Map<String, dynamic>> envelopes;
  final String? runId;
  final String? messageId;
  final String? fallback;
  final FormView? form;
}

class UiSnapshot {
  UiSnapshot({
    required this.version,
    required this.catalogId,
    required this.threadId,
    required this.surfaces,
    this.running,
  });
  final String version;
  final String catalogId;
  final String threadId;
  final List<SnapshotSurface> surfaces;
  final Map<String, dynamic>? running;
}

/// Where a restored surface belongs in the thread history.
class SurfacePlacement {
  const SurfacePlacement({required this.surfaceId, this.messageId});
  final String surfaceId;
  final String? messageId;
}

bool isA2uiEventValue(Object? value) {
  if (value is! Map) return false;
  final surfaceId = value['surfaceId'];
  final revision = value['revision'];
  final seq = value['seq'];
  final envelope = value['envelope'];
  return surfaceId is String &&
      surfaceId.isNotEmpty &&
      revision is num &&
      seq is num &&
      envelope is Map;
}

A2uiEventValue? parseA2uiEventValue(Object? value) {
  if (!isA2uiEventValue(value)) return null;
  final map = value as Map;
  final kindRaw = map['kind'];
  final kind = kindRaw == 'form' ? SurfaceKind.form : SurfaceKind.card;
  return A2uiEventValue(
    version: map['version']?.toString() ?? '',
    surfaceId: map['surfaceId'] as String,
    kind: kind,
    revision: (map['revision'] as num).toInt(),
    seq: (map['seq'] as num).toInt(),
    envelope: Map<String, dynamic>.from(map['envelope'] as Map),
    threadId: map['threadId'] as String?,
    runId: map['runId'] as String?,
    messageId: map['messageId'] as String?,
    fallback: map['fallback'] as String?,
    form: parseFormView(map['form']),
  );
}

FormView? parseFormView(Object? raw) {
  if (raw is! Map) return null;
  final interruptId = raw['interruptId'];
  final token = raw['token'];
  final revision = raw['revision'];
  final title = raw['title'];
  final question = raw['question'];
  if (interruptId is! String ||
      token is! String ||
      revision is! num ||
      title is! String ||
      question is! String) {
    return null;
  }
  final fields = <FormField>[];
  if (raw['fields'] is List) {
    for (final item in raw['fields'] as List) {
      if (item is! Map) continue;
      final name = item['name'];
      final label = item['label'];
      final type = item['type'];
      if (name is! String || label is! String || type is! String) continue;
      final options = <FormOption>[];
      if (item['options'] is List) {
        for (final opt in item['options'] as List) {
          if (opt is! Map) continue;
          final v = opt['value'];
          final l = opt['label'];
          if (v is String && l is String) {
            options.add(FormOption(value: v, label: l));
          }
        }
      }
      fields.add(
        FormField(
          name: name,
          label: label,
          type: type,
          hint: item['hint'] as String? ?? '',
          required: item['required'] == true,
          maxLength: item['maxLength'] is num
              ? (item['maxLength'] as num).toInt()
              : null,
          options: options,
        ),
      );
    }
  }
  return FormView(
    interruptId: interruptId,
    token: token,
    revision: revision.toInt(),
    title: title,
    question: question,
    fields: fields,
  );
}

SnapshotSurface? parseSnapshotSurface(Object? raw) {
  if (raw is! Map) return null;
  final surfaceId = raw['surfaceId'];
  final revision = raw['revision'];
  if (surfaceId is! String || revision is! num) return null;
  final kindRaw = raw['kind'];
  final kind = kindRaw == 'form' ? SurfaceKind.form : SurfaceKind.card;
  final envelopes = <Map<String, dynamic>>[];
  if (raw['envelopes'] is List) {
    for (final env in raw['envelopes'] as List) {
      if (env is Map) envelopes.add(Map<String, dynamic>.from(env));
    }
  }
  return SnapshotSurface(
    surfaceId: surfaceId,
    kind: kind,
    revision: revision.toInt(),
    envelopes: envelopes,
    runId: raw['runId'] as String?,
    messageId: raw['messageId'] as String?,
    fallback: raw['fallback'] as String?,
    form: parseFormView(raw['form']),
  );
}

UiSnapshot? parseUiSnapshot(String body) {
  Object? decoded;
  try {
    decoded = jsonDecode(body);
  } on FormatException {
    return null;
  }
  if (decoded is! Map) return null;
  final version = decoded['version'];
  final cat = decoded['catalogId'];
  final threadId = decoded['threadId'];
  if (version is! String || cat is! String || threadId is! String) return null;
  final surfaces = <SnapshotSurface>[];
  if (decoded['surfaces'] is List) {
    for (final item in decoded['surfaces'] as List) {
      final s = parseSnapshotSurface(item);
      if (s != null) surfaces.add(s);
    }
  }
  return UiSnapshot(
    version: version,
    catalogId: cat,
    threadId: threadId,
    surfaces: surfaces,
    running: decoded['running'] is Map
        ? Map<String, dynamic>.from(decoded['running'] as Map)
        : null,
  );
}

SurfacePlacement? parseSurfacePlacement(Object? raw) {
  if (raw is! Map) return null;
  final surfaceId = raw['surfaceId'];
  if (surfaceId is! String || surfaceId.isEmpty) return null;
  return SurfacePlacement(
    surfaceId: surfaceId,
    messageId: raw['messageId'] as String?,
  );
}

String? envelopeOp(Map<String, dynamic> env) {
  for (final op in [
    'createSurface',
    'updateComponents',
    'updateDataModel',
    'deleteSurface',
  ]) {
    if (env.containsKey(op)) return op;
  }
  return null;
}
