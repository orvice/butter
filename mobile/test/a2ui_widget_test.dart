import 'package:butter_mobile/features/a2ui/protocol.dart';
import 'package:butter_mobile/features/a2ui/surface_processor.dart';
import 'package:butter_mobile/features/a2ui/surface_store.dart';
import 'package:butter_mobile/features/a2ui/surface_widgets.dart';
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  testWidgets('a card renders Text and falls back on error', (tester) async {
    final processed = rebuildFromEnvelopes('c1', [
      {
        'version': a2uiVersion,
        'createSurface': {'surfaceId': 'c1', 'catalogId': catalogId},
      },
      {
        'version': a2uiVersion,
        'updateComponents': {
          'surfaceId': 'c1',
          'components': [
            {'id': 'root', 'component': 'Text', 'text': 'Hello card'},
          ],
        },
      },
    ]);
    final entry = SurfaceEntry(
      id: 'c1',
      kind: SurfaceKind.card,
      revision: 1,
      seq: 0,
      origin: SurfaceOrigin.stream,
      fallback: 'Hello card',
      processed: processed,
    );
    await tester.pumpWidget(
      MaterialApp(
        home: Scaffold(
          body: A2uiSurfaceView(
            entry: entry,
            draft: const {},
            locked: false,
            onDraftChanged: (_) {},
            onSubmit: () {},
          ),
        ),
      ),
    );
    expect(find.text('Hello card'), findsOneWidget);

    final broken = entry.copyWith(error: 'bad', processed: null);
    await tester.pumpWidget(
      MaterialApp(
        home: Scaffold(
          body: A2uiSurfaceView(
            entry: broken,
            draft: const {},
            locked: false,
            onDraftChanged: (_) {},
            onSubmit: () {},
          ),
        ),
      ),
    );
    expect(find.byKey(const Key('a2ui-fallback-c1')), findsOneWidget);
  });
}
