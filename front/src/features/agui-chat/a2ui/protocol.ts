import type { AGUIRunningRun } from '@/api/agui'

// Butter's A2UI-over-AG-UI contract (docs/api.md, "A2UI surfaces"). AG-UI
// stays the transport: each A2UI v0.9.1 message arrives as one CUSTOM event
// named EVENT_NAME, and the UI snapshot endpoint returns the same envelopes
// for a thread's current surfaces.

export const A2UI_VERSION = 'v0.9.1'
export const CATALOG_ID = 'butter-basic-v1'
export const EVENT_NAME = 'butter.a2ui'
export const SUBMIT_EVENT = 'butter.submitForm'

// The forwardedProps declaration that makes a run A2UI-capable. It selects
// what the server ships; it never uploads components.
export const A2UI_CAPABILITY = {
  version: A2UI_VERSION,
  catalogs: [CATALOG_ID],
} as const

export type SurfaceKind = 'card' | 'form'

export interface FormOption {
  value: string
  label: string
}

export interface FormField {
  name: string
  label: string
  hint?: string
  required?: boolean
  type: 'text' | 'single_choice'
  maxLength?: number
  options?: FormOption[]
}

// FormView binds a server-built form surface to the Interrupt it answers.
export interface FormView {
  interruptId: string
  token: string
  revision: number
  title: string
  question: string
  fields: FormField[]
}

export type Envelope = Record<string, unknown> & { version?: string }

// A2UIEventValue is the value of one butter.a2ui CUSTOM event.
export interface A2UIEventValue {
  version: string
  surfaceId: string
  kind: SurfaceKind
  revision: number
  seq: number
  threadId?: string
  runId?: string
  messageId?: string
  envelope: Envelope
  fallback?: string
  form?: FormView
}

export interface SnapshotSurface {
  surfaceId: string
  kind: SurfaceKind
  revision: number
  runId?: string
  messageId?: string
  fallback?: string
  form?: FormView
  envelopes: Envelope[]
}

export interface UISnapshot {
  version: string
  catalogId: string
  threadId: string
  surfaces: SnapshotSurface[]
  // running names the run in flight: the surfaces are then those the run
  // found, and the forms still open after its first turn.
  running?: AGUIRunningRun
}

export function isA2UIEventValue(v: unknown): v is A2UIEventValue {
  if (!v || typeof v !== 'object') return false
  const o = v as Record<string, unknown>
  return (
    typeof o.surfaceId === 'string' &&
    o.surfaceId !== '' &&
    typeof o.revision === 'number' &&
    typeof o.seq === 'number' &&
    !!o.envelope &&
    typeof o.envelope === 'object'
  )
}

export function envelopeOp(env: Envelope): string | undefined {
  for (const op of [
    'createSurface',
    'updateComponents',
    'updateDataModel',
    'deleteSurface',
  ]) {
    if (op in env) return op
  }
  return undefined
}
