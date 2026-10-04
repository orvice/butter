import { describe, expect, it } from 'vitest'
import { A2UI_VERSION, type SnapshotSurface, type UISnapshot } from './protocol'
import { A2UIStore } from './store'

function card(
  surfaceId: string,
  revision: number,
  text: string
): SnapshotSurface {
  return {
    surfaceId,
    kind: 'card',
    revision,
    fallback: text,
    envelopes: [
      {
        version: A2UI_VERSION,
        createSurface: { surfaceId, catalogId: 'butter-basic-v1' },
      },
      {
        version: A2UI_VERSION,
        updateComponents: {
          surfaceId,
          components: [{ id: 'root', component: 'Text', text }],
        },
      },
    ],
  }
}

function snapshot(...surfaces: SnapshotSurface[]): UISnapshot {
  return {
    version: A2UI_VERSION,
    catalogId: 'butter-basic-v1',
    threadId: 't',
    surfaces,
  }
}

// A2UIStore.restore shows a read of the thread: the store ends up holding
// what the read holds, as a load of the thread would.
describe('A2UIStore.restore', () => {
  it('keeps a surface it holds at that revision, and moves one on to a newer one', () => {
    const store = new A2UIStore()
    store.restore(
      snapshot(card('card-1', 1, 'one'), card('card-2', 1, 'two')),
      []
    )
    const kept = store.entry('card-1')
    store.restore(
      snapshot(card('card-1', 1, 'one'), card('card-2', 2, 'two, again')),
      []
    )
    expect(store.entry('card-1')).toBe(kept)
    expect(store.entry('card-2')).toMatchObject({
      revision: 2,
      fallback: 'two, again',
      deleted: false,
    })
  })

  it('drops what the read no longer holds, and it does not come back', () => {
    const store = new A2UIStore()
    store.restore(snapshot(card('card-1', 1, 'one')), ['card-1'])
    store.restore(snapshot(), [])
    expect(store.entry('card-1')).toMatchObject({ deleted: true })
    // A late event for it changes nothing, as after its deleteSurface.
    store.apply({
      version: A2UI_VERSION,
      surfaceId: 'card-1',
      kind: 'card',
      revision: 3,
      seq: 0,
      envelope: {
        version: A2UI_VERSION,
        updateComponents: { surfaceId: 'card-1', components: [] },
      },
    })
    expect(store.entry('card-1')).toMatchObject({ deleted: true })
  })

  it('places the surfaces a reply shows, so the others show on their own', () => {
    const store = new A2UIStore()
    store.restore(
      snapshot(card('card-1', 1, 'one'), card('card-2', 1, 'two')),
      ['card-1']
    )
    expect(store.isPlaced('card-1')).toBe(true)
    expect(store.isPlaced('card-2')).toBe(false)
  })
})
