import type { PlaneApp } from '@/gen/agents/v1/plane_pb'

/**
 * The operator-editable shape of a Plane App as the form holds it.
 *
 * It is deliberately string-typed (that is what inputs produce) and it never
 * holds a stored secret: the two secret inputs exist only on the create form
 * and are cleared after submission, so a reload or a re-render can never
 * repopulate them from the server.
 */
export type PlaneFormState = {
  displayName: string
  clientId: string
  agentId: string
  allowedUserIds: string
  maxRunSeconds: string
  /** Create-only, write-only. Always empty after a successful submit. */
  clientSecret: string
  webhookSecret: string
}

export type PlaneAppDraft = {
  id?: string
  revision?: bigint
  displayName: string
  clientId: string
  agentId: string
  allowedUserIds: string[]
  maxRunSeconds?: number
}

export type PlaneInputOptions = {
  /** Set when editing: selects the App and carries the revision to fence on. */
  id?: string
  revision?: bigint
  /** Fixed by an Agent entry; always wins over the URL's Agent. */
  lockedAgentId?: string
}

/** Seeds the form from an App (edit) or from a locked Agent (create). */
export function toPlaneFormState(
  app?: PlaneApp,
  lockedAgentId?: string
): PlaneFormState {
  return {
    displayName: app?.displayName ?? '',
    clientId: app?.clientId ?? '',
    agentId: app?.agentId ?? lockedAgentId ?? '',
    allowedUserIds: (app?.allowedUserIds ?? []).join('\n'),
    maxRunSeconds:
      app?.maxRunSeconds === undefined ? '' : String(app.maxRunSeconds),
    clientSecret: '',
    webhookSecret: '',
  }
}

export function parsePlaneUserIDs(raw: string): string[] {
  return raw
    .split(/[\s,]+/)
    .map((entry) => entry.trim())
    .filter(Boolean)
}

function parseMaxRunSeconds(raw: string): number | undefined {
  const value = raw.trim()
  if (value === '') return undefined
  const parsed = Number(value)
  if (!Number.isInteger(parsed) || parsed < 0) return undefined
  return parsed
}

/**
 * Builds the App payload the RPC receives.
 *
 * Two rules are load-bearing:
 *
 * - The locked Agent always wins, so an Agent-scoped entry cannot register an
 *   App for a different Agent by editing the URL.
 * - The revision is carried through on edit. An update without it is refused
 *   server-side, which is what makes a stale form lose instead of silently
 *   overwriting someone else's change.
 */
export function toPlaneAppInput(
  form: PlaneFormState,
  options: PlaneInputOptions = {}
): PlaneAppDraft {
  return {
    id: options.id,
    revision: options.revision,
    displayName: form.displayName.trim(),
    clientId: form.clientId.trim(),
    agentId: options.lockedAgentId ?? form.agentId,
    allowedUserIds: parsePlaneUserIDs(form.allowedUserIds),
    maxRunSeconds: parseMaxRunSeconds(form.maxRunSeconds),
  }
}

/**
 * Decides whether a freshly fetched App should replace what is in the form.
 *
 * A background refresh (or a refetch after a revision conflict) may arrive
 * while the operator is typing. Re-seeding then would silently discard
 * unsaved input, so a dirty form only re-seeds when the caller forces it —
 * which the edit page does after an aborted save, where the user has already
 * been told the values were reloaded.
 */
export function shouldReseedPlaneForm(args: {
  dirty: boolean
  syncedKey: string | null
  nextKey: string | null
  force?: boolean
}): boolean {
  if (!args.nextKey || args.nextKey === args.syncedKey) return false
  if (args.force) return true
  return !args.dirty
}

/** Stable identity of a server revision used to decide re-seeding. */
export function planeRevisionKey(app: PlaneApp | undefined): string | null {
  if (!app) return null
  return `${app.id}:${app.revision.toString()}`
}
