import 'package:butter_mobile/features/a2ui/protocol.dart';
import 'package:butter_mobile/features/a2ui/surface_store.dart';
import 'package:flutter_test/flutter_test.dart';

SnapshotSurface card(String id, int revision, String text) => SnapshotSurface(
  surfaceId: id,
  kind: SurfaceKind.card,
  revision: revision,
  fallback: text,
  envelopes: [
    {
      'version': a2uiVersion,
      'createSurface': {'surfaceId': id, 'catalogId': catalogId},
    },
    {
      'version': a2uiVersion,
      'updateComponents': {
        'surfaceId': id,
        'components': [
          {'id': 'root', 'component': 'Text', 'text': text},
        ],
      },
    },
  ],
);

UiSnapshot snapshot(List<SnapshotSurface> surfaces) => UiSnapshot(
  version: a2uiVersion,
  catalogId: catalogId,
  threadId: 't',
  surfaces: surfaces,
);

void main() {
  test(
    'restore keeps revision and drops surfaces the read no longer holds',
    () {
      final store = A2uiSurfaceStore();
      store.restore(snapshot([card('c1', 1, 'one'), card('c2', 1, 'two')]), []);
      store.restore(
        snapshot([card('c1', 1, 'one'), card('c2', 2, 'two again')]),
        [],
      );
      expect(store.entry('c1')?.revision, 1);
      expect(store.entry('c2')?.revision, 2);
      store.restore(snapshot([]), []);
      expect(store.entry('c1')?.deleted, isTrue);
      store.apply(
        A2uiEventValue(
          version: a2uiVersion,
          surfaceId: 'c1',
          kind: SurfaceKind.card,
          revision: 3,
          seq: 0,
          envelope: {
            'version': a2uiVersion,
            'updateComponents': {
              'surfaceId': 'c1',
              'components': [
                {'id': 'root', 'component': 'Text', 'text': 'late'},
              ],
            },
          },
        ),
      );
      expect(store.entry('c1')?.deleted, isTrue);
    },
  );

  test('stale revision is ignored', () {
    final store = A2uiSurfaceStore();
    store.apply(
      A2uiEventValue(
        version: a2uiVersion,
        surfaceId: 'c1',
        kind: SurfaceKind.card,
        revision: 2,
        seq: 1,
        envelope: {
          'version': a2uiVersion,
          'updateComponents': {
            'surfaceId': 'c1',
            'components': [
              {'id': 'root', 'component': 'Text', 'text': 'new'},
            ],
          },
        },
        fallback: 'new',
      ),
    );
    store.apply(
      A2uiEventValue(
        version: a2uiVersion,
        surfaceId: 'c1',
        kind: SurfaceKind.card,
        revision: 1,
        seq: 0,
        envelope: {
          'version': a2uiVersion,
          'updateComponents': {
            'surfaceId': 'c1',
            'components': [
              {'id': 'root', 'component': 'Text', 'text': 'old'},
            ],
          },
        },
        fallback: 'old',
      ),
    );
    expect(store.entry('c1')?.fallback, 'new');
  });

  test('unknown component becomes a safe fallback error', () {
    final store = A2uiSurfaceStore();
    store.apply(
      A2uiEventValue(
        version: a2uiVersion,
        surfaceId: 'c1',
        kind: SurfaceKind.card,
        revision: 1,
        seq: 0,
        envelope: {
          'version': a2uiVersion,
          'updateComponents': {
            'surfaceId': 'c1',
            'components': [
              {'id': 'root', 'component': 'AlienWidget', 'text': 'x'},
            ],
          },
        },
        fallback: 'readable',
      ),
    );
    expect(store.entry('c1')?.error, contains('AlienWidget'));
    expect(store.entry('c1')?.fallback, 'readable');
  });
}
