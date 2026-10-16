import 'package:butter_mobile/features/a2ui/protocol.dart';
import 'package:butter_mobile/features/a2ui/surface_processor.dart';
import 'package:butter_mobile/features/a2ui/surface_store.dart';
import 'package:butter_mobile/features/a2ui/surface_widgets.dart';
import 'package:flutter/material.dart' hide FormField;
import 'package:flutter_test/flutter_test.dart';

FormView sampleForm() => FormView(
  interruptId: 'i1',
  token: 'tok',
  revision: 1,
  title: 'Approve',
  question: 'Go?',
  fields: [
    FormField(name: 'reason', label: 'Reason', type: 'text', maxLength: 20),
  ],
);

SurfaceEntry formEntry({FormStatus? status}) {
  const id = 'form-1';
  final processed = rebuildFromEnvelopes(id, [
    {
      'version': a2uiVersion,
      'createSurface': {'surfaceId': id, 'catalogId': catalogId},
    },
    {
      'version': a2uiVersion,
      'updateComponents': {
        'surfaceId': id,
        'components': [
          {
            'id': 'root',
            'component': 'Column',
            'children': ['field', 'btn'],
          },
          {
            'id': 'field',
            'component': 'TextField',
            'name': 'reason',
            'label': 'Reason',
          },
          {
            'id': 'btn',
            'component': 'Button',
            'action': {
              'event': {'name': submitEventName},
            },
            'child': 'btnLabel',
          },
          {'id': 'btnLabel', 'component': 'Text', 'text': 'Send'},
        ],
      },
    },
  ]);
  return SurfaceEntry(
    id: id,
    kind: SurfaceKind.form,
    revision: 1,
    seq: 0,
    origin: SurfaceOrigin.stream,
    form: sampleForm(),
    formStatus: status ?? FormStatus.pending,
    processed: processed,
  );
}

void main() {
  group('A2uiSurfaceStore form submit', () {
    test('double submit calls the submitter only once', () async {
      final store = A2uiSurfaceStore();
      store.apply(
        A2uiEventValue(
          version: a2uiVersion,
          surfaceId: 'form-1',
          kind: SurfaceKind.form,
          revision: 1,
          seq: 0,
          envelope: const {'version': a2uiVersion},
          form: sampleForm(),
        ),
      );
      store.setDraft('form-1', {'reason': 'ok'});
      var calls = 0;
      store.setSubmitter((e, values) async {
        calls++;
        await Future<void>.delayed(const Duration(milliseconds: 50));
      });
      final first = store.submitForm('form-1', store.draftFor('form-1'));
      final second = store.submitForm('form-1', store.draftFor('form-1'));
      await first;
      await second;
      expect(calls, 1);
      expect(store.entry('form-1')?.formStatus, FormStatus.answered);
    });

    test('failure keeps draft and field errors from 422', () async {
      final store = A2uiSurfaceStore()
        ..apply(
          A2uiEventValue(
            version: a2uiVersion,
            surfaceId: 'form-1',
            kind: SurfaceKind.form,
            revision: 1,
            seq: 0,
            envelope: const {'version': a2uiVersion},
            form: sampleForm(),
          ),
        );
      store.setDraft('form-1', {'reason': 'bad value'});
      store.setSubmitter((e, values) async {
        throw FormSubmitError(
          'Invalid',
          code: 'form_invalid',
          fieldErrors: {'reason': 'too long'},
        );
      });
      await store.submitForm('form-1', store.draftFor('form-1'));
      expect(store.entry('form-1')?.formStatus, FormStatus.error);
      expect(store.entry('form-1')?.fieldErrors['reason'], 'too long');
      expect(store.draftFor('form-1')['reason'], 'bad value');
    });

    test('409 form_stale shows error and keeps draft', () async {
      final store = A2uiSurfaceStore()
        ..apply(
          A2uiEventValue(
            version: a2uiVersion,
            surfaceId: 'form-1',
            kind: SurfaceKind.form,
            revision: 1,
            seq: 0,
            envelope: const {'version': a2uiVersion},
            form: sampleForm(),
          ),
        );
      store.setDraft('form-1', {'reason': 'draft text'});
      store.setSubmitter((e, values) async {
        throw FormSubmitError('Stale form', code: 'form_stale');
      });
      await store.submitForm('form-1', store.draftFor('form-1'));
      expect(store.entry('form-1')?.formStatus, FormStatus.error);
      expect(store.draftFor('form-1')['reason'], 'draft text');
    });

    test('409 form_answered locks the form', () async {
      final store = A2uiSurfaceStore()
        ..apply(
          A2uiEventValue(
            version: a2uiVersion,
            surfaceId: 'form-1',
            kind: SurfaceKind.form,
            revision: 1,
            seq: 0,
            envelope: const {'version': a2uiVersion},
            form: sampleForm(),
          ),
        );
      store.setSubmitter((e, values) async {
        throw FormSubmitError('Already answered', code: 'form_answered');
      });
      await store.submitForm('form-1', {'reason': 'x'});
      expect(store.entry('form-1')?.formStatus, FormStatus.answered);
    });

    test('success clears draft and locks the form', () async {
      final store = A2uiSurfaceStore()
        ..apply(
          A2uiEventValue(
            version: a2uiVersion,
            surfaceId: 'form-1',
            kind: SurfaceKind.form,
            revision: 1,
            seq: 0,
            envelope: const {'version': a2uiVersion},
            form: sampleForm(),
          ),
        );
      store.setDraft('form-1', {'reason': 'saved'});
      store.setSubmitter((e, values) async {});
      await store.submitForm('form-1', store.draftFor('form-1'));
      expect(store.entry('form-1')?.formStatus, FormStatus.answered);
      expect(store.draftFor('form-1'), isEmpty);
    });
  });

  testWidgets('text field keeps draft text across rebuilds', (tester) async {
    final entry = formEntry();
    var draft = <String, Object?>{'reason': 'typed draft'};
    await tester.pumpWidget(
      MaterialApp(
        home: Scaffold(
          body: StatefulBuilder(
            builder: (context, setState) => A2uiSurfaceView(
              entry: entry,
              draft: draft,
              locked: false,
              onDraftChanged: (next) => setState(() => draft = next),
              onSubmit: () {},
            ),
          ),
        ),
      ),
    );
    await tester.enterText(
      find.byKey(const Key('field-reason')),
      'typed draft',
    );
    await tester.pumpWidget(
      MaterialApp(
        home: Scaffold(
          body: A2uiSurfaceView(
            entry: entry,
            draft: draft,
            locked: false,
            onDraftChanged: (next) => draft = next,
            onSubmit: () {},
          ),
        ),
      ),
    );
    await tester.pump();
    expect(find.text('typed draft'), findsOneWidget);
  });
}
