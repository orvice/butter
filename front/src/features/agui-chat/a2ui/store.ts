import { useSyncExternalStore } from 'react'
import type { ReactComponentImplementation } from '@a2ui/react/v0_9'
import { MessageProcessor, type ActionPayload } from '@a2ui/web_core/v0_9'
import { butterCatalog } from './catalog'
import { formValues, validateForm } from './form'
import {
  A2UI_VERSION,
  SUBMIT_EVENT,
  envelopeOp,
  isA2UIEventValue,
  type Envelope,
  type FormView,
  type SurfaceKind,
  type UISnapshot,
} from './protocol'

export type FormStatus = 'pending' | 'submitting' | 'answered' | 'error'

// SurfaceEntry is everything the chat knows about one surface besides the
// renderer's own component and data model.
export interface SurfaceEntry {
  id: string
  kind: SurfaceKind
  revision: number
  seq: number
  // stream: arrived in this page's runs, rendered where its create event
  // landed in the conversation. snapshot: restored from the UI snapshot.
  origin: 'stream' | 'snapshot'
  runId?: string
  messageId?: string
  fallback?: string
  form?: FormView
  deleted: boolean
  // error confines a bad message or unknown catalog to this surface: it
  // renders its fallback text instead.
  error?: string
  formStatus?: FormStatus
  formMessage?: string
  fieldErrors: Record<string, string>
}

// SubmitForm delivers a validated submission; it rejects with the AG-UI
// client's HTTP error (status + parsed JSON payload) when the server
// refuses it.
export type SubmitForm = (
  entry: SurfaceEntry,
  values: Record<string, string>
) => Promise<void>

interface HttpError {
  status?: number
  payload?: {
    error?: string
    code?: string
    fieldErrors?: Record<string, string>
  }
  message?: string
}

// A2UIStore holds one thread's surfaces. Events apply in (revision, seq)
// order per surface: anything not newer than what was applied is ignored,
// so a replayed or stale message never overwrites newer content, and a
// deleted surface never comes back.
export class A2UIStore {
  readonly processor: MessageProcessor<ReactComponentImplementation>
  private readonly entries = new Map<string, SurfaceEntry>()
  private readonly placed = new Set<string>()
  private readonly listeners = new Set<() => void>()
  private version = 0
  private submitter?: SubmitForm

  constructor() {
    this.processor = new MessageProcessor<ReactComponentImplementation>(
      [butterCatalog],
      (action) => this.onAction(action)
    )
  }

  subscribe = (listener: () => void) => {
    this.listeners.add(listener)
    return () => {
      this.listeners.delete(listener)
    }
  }

  getVersion = () => this.version

  entry(id: string): SurfaceEntry | undefined {
    return this.entries.get(id)
  }

  // markPlaced records the restored surfaces a reply of the thread's history
  // shows; the others are shown on their own.
  markPlaced(ids: Iterable<string>) {
    for (const id of ids) this.placed.add(id)
    this.notify()
  }

  isPlaced(id: string): boolean {
    return this.placed.has(id)
  }

  list(): SurfaceEntry[] {
    return [...this.entries.values()]
  }

  // formFor returns the live form bound to an Interrupt, if any.
  formFor(interruptId: string): SurfaceEntry | undefined {
    return this.list().find(
      (e) =>
        e.kind === 'form' && !e.deleted && e.form?.interruptId === interruptId
    )
  }

  setSubmitter(submit: SubmitForm | undefined) {
    this.submitter = submit
  }

  dispose() {
    this.processor.dispose()
    this.listeners.clear()
  }

  private notify() {
    this.version++
    for (const listener of this.listeners) listener()
  }

  private put(entry: SurfaceEntry) {
    this.entries.set(entry.id, entry)
    this.notify()
  }

  // process feeds envelopes to the renderer; a failure is recorded on the
  // surface instead of escaping into the chat.
  private process(entry: SurfaceEntry, envelopes: Envelope[]): SurfaceEntry {
    if (entry.error) return entry
    const unknown = unknownComponent(envelopes)
    if (unknown) {
      return { ...entry, error: `unsupported component ${unknown}` }
    }
    try {
      this.processor.processMessages(
        envelopes as unknown as Parameters<
          MessageProcessor['processMessages']
        >[0]
      )
      return entry
    } catch (err) {
      return {
        ...entry,
        error: err instanceof Error ? err.message : String(err),
      }
    }
  }

  // apply takes one butter.a2ui CUSTOM event value.
  apply(value: unknown) {
    if (!isA2UIEventValue(value)) return
    const prev = this.entries.get(value.surfaceId)
    if (prev?.deleted) return
    if (
      prev &&
      (value.revision < prev.revision ||
        (value.revision === prev.revision && value.seq <= prev.seq))
    ) {
      return
    }
    let entry: SurfaceEntry = prev
      ? {
          ...prev,
          revision: value.revision,
          seq: value.seq,
          fallback: value.fallback ?? prev.fallback,
          form: value.form ?? prev.form,
        }
      : freshEntry(value, value.seq, 'stream')
    if (value.version !== A2UI_VERSION) {
      entry = { ...entry, error: `unsupported A2UI version ${value.version}` }
    } else {
      entry = this.process(entry, [value.envelope])
    }
    this.put(afterEnvelope(entry, value.envelope))
  }

  // applySnapshot restores surfaces from the persisted session. A surface
  // the client already holds at the same or a newer revision is left alone.
  applySnapshot(snapshot: UISnapshot) {
    for (const item of snapshot.surfaces ?? []) {
      const prev = this.entries.get(item.surfaceId)
      if (
        prev?.deleted ||
        (prev && !prev.error && prev.revision >= item.revision)
      ) {
        continue
      }
      if (prev) {
        try {
          this.processor.processMessages([
            {
              version: A2UI_VERSION,
              deleteSurface: { surfaceId: item.surfaceId },
            },
          ] as unknown as Parameters<MessageProcessor['processMessages']>[0])
        } catch {
          // The renderer may never have created it; rebuilding is the point.
        }
      }
      let entry = freshEntry(
        item,
        Math.max(0, (item.envelopes?.length ?? 1) - 1),
        prev?.origin ?? 'snapshot'
      )
      entry = this.process(entry, item.envelopes)
      for (const env of item.envelopes) entry = afterEnvelope(entry, env)
      this.entries.set(entry.id, entry)
    }
    this.notify()
  }

  private async onAction(action: ActionPayload) {
    if (action.name !== SUBMIT_EVENT) return
    const entry = this.entries.get(action.surfaceId)
    const form = entry?.form
    if (!entry || !form || entry.kind !== 'form') return
    if (entry.formStatus === 'submitting' || entry.formStatus === 'answered') {
      return
    }
    const draft = (action.context?.values ?? {}) as Record<string, unknown>
    const values = formValues(form, draft)
    const fieldErrors = validateForm(form, values)
    if (Object.keys(fieldErrors).length > 0) {
      this.put({
        ...entry,
        formStatus: 'error',
        formMessage: 'Fix the highlighted fields and submit again.',
        fieldErrors,
      })
      return
    }
    const submit = this.submitter
    if (!submit) {
      this.put({
        ...entry,
        formStatus: 'error',
        formMessage: 'Submitting is not available right now.',
        fieldErrors: {},
      })
      return
    }
    this.put({
      ...entry,
      formStatus: 'submitting',
      formMessage: undefined,
      fieldErrors: {},
    })
    try {
      await submit(entry, values)
      const current = this.entries.get(entry.id)
      if (current && current.formStatus === 'submitting') {
        this.put({ ...current, formStatus: 'answered' })
      }
    } catch (err) {
      const current = this.entries.get(entry.id)
      // The server may have consumed the answer before the run failed; its
      // answered marker wins over the error.
      if (!current || current.formStatus === 'answered') return
      const httpErr = err as HttpError
      const message =
        httpErr?.payload?.error ??
        (err instanceof Error ? err.message : 'The submission failed.')
      if (httpErr?.payload?.code === 'form_answered') {
        this.put({
          ...current,
          formStatus: 'answered',
          formMessage: 'This form was already submitted.',
          fieldErrors: {},
        })
        return
      }
      this.put({
        ...current,
        formStatus: 'error',
        formMessage: message,
        fieldErrors: httpErr?.payload?.fieldErrors ?? {},
      })
    }
  }
}

// unknownComponent names the first component outside butter-basic-v1. The
// renderer would draw a placeholder for it; the surface shows its readable
// fallback instead.
function unknownComponent(envelopes: Envelope[]): string | undefined {
  for (const env of envelopes) {
    const update = env.updateComponents as
      { components?: Array<{ component?: unknown }> } | undefined
    for (const c of update?.components ?? []) {
      const name = typeof c?.component === 'string' ? c.component : ''
      if (!butterCatalog.components.has(name)) return name || '(unnamed)'
    }
  }
  return undefined
}

// freshEntry is a surface as first seen, from a live event or a snapshot.
function freshEntry(
  surface: {
    surfaceId: string
    kind: SurfaceKind
    revision: number
    runId?: string
    messageId?: string
    fallback?: string
    form?: FormView
  },
  seq: number,
  origin: SurfaceEntry['origin']
): SurfaceEntry {
  return {
    id: surface.surfaceId,
    kind: surface.kind === 'form' ? 'form' : 'card',
    revision: surface.revision,
    seq,
    origin,
    runId: surface.runId,
    messageId: surface.messageId,
    fallback: surface.fallback,
    form: surface.form,
    deleted: false,
    formStatus: surface.kind === 'form' ? 'pending' : undefined,
    fieldErrors: {},
  }
}

// afterEnvelope folds what an envelope means for the chat into the entry.
function afterEnvelope(entry: SurfaceEntry, env: Envelope): SurfaceEntry {
  switch (envelopeOp(env)) {
    case 'deleteSurface':
      return { ...entry, deleted: true }
    case 'updateDataModel': {
      const update = env.updateDataModel as { path?: string; value?: unknown }
      if (
        entry.kind === 'form' &&
        update?.path === '/status' &&
        update.value === 'answered'
      ) {
        return { ...entry, formStatus: 'answered' }
      }
      return entry
    }
    default:
      return entry
  }
}

// useA2UIStore re-renders the caller whenever the store changes.
export function useA2UIStore(store: A2UIStore): A2UIStore {
  useSyncExternalStore(store.subscribe, store.getVersion)
  return store
}
