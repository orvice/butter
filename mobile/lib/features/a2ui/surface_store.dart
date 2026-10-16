import 'package:flutter/foundation.dart';

import 'catalog.dart';
import 'form_logic.dart' show formValues, validateForm;
import 'protocol.dart';
import 'surface_processor.dart';

/// Everything the chat knows about one surface besides its rendered tree.
@immutable
class SurfaceEntry {
  const SurfaceEntry({
    required this.id,
    required this.kind,
    required this.revision,
    required this.seq,
    required this.origin,
    this.runId,
    this.messageId,
    this.fallback,
    this.form,
    this.deleted = false,
    this.error,
    this.formStatus,
    this.formMessage,
    this.fieldErrors = const {},
    this.processed,
  });

  final String id;
  final SurfaceKind kind;
  final int revision;
  final int seq;
  final SurfaceOrigin origin;
  final String? runId;
  final String? messageId;
  final String? fallback;
  final FormView? form;
  final bool deleted;
  final String? error;
  final FormStatus? formStatus;
  final String? formMessage;
  final Map<String, String> fieldErrors;
  final ProcessedSurface? processed;

  SurfaceEntry copyWith({
    int? revision,
    int? seq,
    String? fallback,
    FormView? form,
    bool? deleted,
    String? error,
    FormStatus? formStatus,
    String? formMessage,
    Map<String, String>? fieldErrors,
    ProcessedSurface? processed,
    bool clearError = false,
    bool clearFormMessage = false,
  }) => SurfaceEntry(
    id: id,
    kind: kind,
    revision: revision ?? this.revision,
    seq: seq ?? this.seq,
    origin: origin,
    runId: runId,
    messageId: messageId,
    fallback: fallback ?? this.fallback,
    form: form ?? this.form,
    deleted: deleted ?? this.deleted,
    error: clearError ? null : (error ?? this.error),
    formStatus: formStatus ?? this.formStatus,
    formMessage: clearFormMessage ? null : (formMessage ?? this.formMessage),
    fieldErrors: fieldErrors ?? this.fieldErrors,
    processed: processed ?? this.processed,
  );
}

typedef SubmitForm =
    Future<void> Function(SurfaceEntry entry, Map<String, String> values);

/// Holds one thread's A2UI surfaces. Events apply in (revision, seq) order.
class A2uiSurfaceStore extends ChangeNotifier {
  final _entries = <String, SurfaceEntry>{};
  final _placed = <String>{};
  SubmitForm? _submitter;
  final _drafts = <String, Map<String, Object?>>{};

  Map<String, Object?> draftFor(String surfaceId) =>
      Map<String, Object?>.from(_drafts[surfaceId] ?? const {});

  void setDraft(String surfaceId, Map<String, Object?> draft) {
    _drafts[surfaceId] = Map<String, Object?>.from(draft);
  }

  void clearDrafts() => _drafts.clear();

  SurfaceEntry? entry(String id) => _entries[id];

  Iterable<SurfaceEntry> get entries => _entries.values;

  bool isPlaced(String id) => _placed.contains(id);

  void markPlaced(Iterable<String> ids) {
    _placed
      ..clear()
      ..addAll(ids);
    notifyListeners();
  }

  SurfaceEntry? formFor(String interruptId) {
    for (final e in _entries.values) {
      if (e.deleted) continue;
      if (e.kind == SurfaceKind.form &&
          e.form?.interruptId == interruptId &&
          e.formStatus != FormStatus.answered) {
        return e;
      }
    }
    return null;
  }

  void setSubmitter(SubmitForm? submit) => _submitter = submit;

  void clear() {
    _entries.clear();
    _placed.clear();
    _drafts.clear();
    notifyListeners();
  }

  void apply(A2uiEventValue value) {
    final prev = _entries[value.surfaceId];
    if (prev?.deleted == true) return;
    if (prev != null &&
        (value.revision < prev.revision ||
            (value.revision == prev.revision && value.seq <= prev.seq))) {
      return;
    }
    var entry = prev != null
        ? prev.copyWith(
            revision: value.revision,
            seq: value.seq,
            fallback: value.fallback ?? prev.fallback,
            form: value.form ?? prev.form,
          )
        : _freshEntry(value, value.seq, SurfaceOrigin.stream);
    if (value.version != a2uiVersion) {
      entry = entry.copyWith(
        error: 'unsupported A2UI version ${value.version}',
      );
    } else {
      entry = _processEnvelope(
        entry,
        Map<String, dynamic>.from(value.envelope),
      );
    }
    entry = _afterEnvelope(entry, value.envelope);
    _entries[value.surfaceId] = entry;
    notifyListeners();
  }

  void applySnapshot(UiSnapshot snapshot) {
    for (final item in snapshot.surfaces) {
      final prev = _entries[item.surfaceId];
      if (prev?.deleted == true) continue;
      if (prev != null && !prev.deleted && prev.revision >= item.revision) {
        if (prev.error == null) continue;
      }
      var entry = _freshEntry(
        _snapshotAsEvent(item),
        item.envelopes.isEmpty ? 0 : item.envelopes.length - 1,
        prev?.origin ?? SurfaceOrigin.snapshot,
      );
      final unknown = unknownComponentInEnvelopes(item.envelopes);
      if (unknown != null) {
        entry = entry.copyWith(error: 'unsupported component $unknown');
      } else if (entry.kind == SurfaceKind.card &&
          cardUsesFormComponents(item.envelopes)) {
        entry = entry.copyWith(error: 'unsupported component in card');
      } else {
        var processed = ProcessedSurface(surfaceId: item.surfaceId);
        for (final env in item.envelopes) {
          final err = applyEnvelope(processed, Map<String, dynamic>.from(env));
          if (err != null) {
            entry = entry.copyWith(error: err);
            processed = ProcessedSurface(surfaceId: item.surfaceId);
            break;
          }
          entry = _afterEnvelope(entry, env);
        }
        if (entry.error == null) {
          entry = entry.copyWith(processed: processed);
        }
      }
      _entries[item.surfaceId] = entry;
    }
    notifyListeners();
  }

  /// Aligns the store with a thread read: snapshot surfaces plus placement.
  void restore(UiSnapshot snapshot, Iterable<String> placed) {
    final held = snapshot.surfaces.map((s) => s.surfaceId).toSet();
    for (final entry in _entries.values.toList()) {
      if (entry.deleted || held.contains(entry.id)) continue;
      _entries[entry.id] = entry.copyWith(deleted: true);
    }
    applySnapshot(snapshot);
    markPlaced(placed);
  }

  Future<void> submitForm(String surfaceId, Map<String, Object?> draft) async {
    final entry = _entries[surfaceId];
    final form = entry?.form;
    if (entry == null || form == null || entry.kind != SurfaceKind.form) {
      return;
    }
    if (entry.formStatus == FormStatus.submitting ||
        entry.formStatus == FormStatus.answered) {
      return;
    }
    final values = formValues(form, draft);
    final fieldErrors = validateForm(form, values);
    if (fieldErrors.isNotEmpty) {
      _entries[surfaceId] = entry.copyWith(
        formStatus: FormStatus.error,
        formMessage: 'Fix the highlighted fields and submit again.',
        fieldErrors: fieldErrors,
      );
      notifyListeners();
      return;
    }
    final submit = _submitter;
    if (submit == null) {
      _entries[surfaceId] = entry.copyWith(
        formStatus: FormStatus.error,
        formMessage: 'Submitting is not available right now.',
        fieldErrors: const {},
      );
      notifyListeners();
      return;
    }
    _entries[surfaceId] = entry.copyWith(
      formStatus: FormStatus.submitting,
      clearFormMessage: true,
      fieldErrors: const {},
    );
    notifyListeners();
    try {
      await submit(entry, values);
      final current = _entries[surfaceId];
      if (current != null && current.formStatus == FormStatus.submitting) {
        _entries[surfaceId] = current.copyWith(formStatus: FormStatus.answered);
        _drafts.remove(surfaceId);
      }
      notifyListeners();
    } catch (err) {
      final current = _entries[surfaceId];
      if (current == null || current.formStatus == FormStatus.answered) return;
      _handleSubmitError(current, err);
    }
  }

  void _handleSubmitError(SurfaceEntry entry, Object err) {
    String message = 'The submission failed.';
    String? code;
    Map<String, String> fieldErrors = const {};
    if (err is FormSubmitError) {
      message = err.message;
      code = err.code;
      fieldErrors = err.fieldErrors;
    }
    if (code == 'form_answered') {
      _entries[entry.id] = entry.copyWith(
        formStatus: FormStatus.answered,
        formMessage: 'This form was already submitted.',
        fieldErrors: const {},
      );
    } else {
      _entries[entry.id] = entry.copyWith(
        formStatus: FormStatus.error,
        formMessage: message,
        fieldErrors: fieldErrors,
      );
    }
    notifyListeners();
  }

  SurfaceEntry _processEnvelope(
    SurfaceEntry entry,
    Map<String, dynamic> envelope,
  ) {
    if (entry.error != null) return entry;
    final unknown = unknownComponentInEnvelopes([
      Map<String, dynamic>.from(envelope),
    ]);
    if (unknown != null) {
      return entry.copyWith(error: 'unsupported component $unknown');
    }
    if (entry.kind == SurfaceKind.card &&
        cardUsesFormComponents([Map<String, dynamic>.from(envelope)])) {
      return entry.copyWith(error: 'unsupported component in card');
    }
    var processed = entry.processed ?? ProcessedSurface(surfaceId: entry.id);
    processed = processed.copy();
    try {
      final err = applyEnvelope(processed, envelope);
      if (err != null) return entry.copyWith(error: err, processed: processed);
      return entry.copyWith(processed: processed, clearError: true);
    } catch (e) {
      return entry.copyWith(
        error: e is Exception ? e.toString() : '$e',
        processed: processed,
      );
    }
  }

  SurfaceEntry _afterEnvelope(SurfaceEntry entry, Map<String, dynamic> env) {
    if (envelopeOp(env) == 'deleteSurface') {
      return entry.copyWith(deleted: true);
    }
    if (envelopeOp(env) == 'updateDataModel') {
      final update = env['updateDataModel'];
      if (update is Map &&
          update['path'] == '/status' &&
          update['value'] == 'answered') {
        return entry.copyWith(formStatus: FormStatus.answered);
      }
    }
    return entry;
  }

  SurfaceEntry _freshEntry(
    A2uiEventValue value,
    int seq,
    SurfaceOrigin origin,
  ) => SurfaceEntry(
    id: value.surfaceId,
    kind: value.kind,
    revision: value.revision,
    seq: seq,
    origin: origin,
    runId: value.runId,
    messageId: value.messageId,
    fallback: value.fallback,
    form: value.form,
    formStatus: value.kind == SurfaceKind.form ? FormStatus.pending : null,
  );

  A2uiEventValue _snapshotAsEvent(SnapshotSurface item) => A2uiEventValue(
    version: a2uiVersion,
    surfaceId: item.surfaceId,
    kind: item.kind,
    revision: item.revision,
    seq: 0,
    envelope: item.envelopes.isNotEmpty
        ? item.envelopes.last
        : const {'version': a2uiVersion},
    runId: item.runId,
    messageId: item.messageId,
    fallback: item.fallback,
    form: item.form,
  );
}

/// Thrown when the server refuses a form submission.
final class FormSubmitError implements Exception {
  FormSubmitError(this.message, {this.code, this.fieldErrors = const {}});
  final String message;
  final String? code;
  final Map<String, String> fieldErrors;
}
