import type { PlaneApp } from '@/gen/agents/v1/plane_pb'
import { describe, expect, it } from 'vitest'
import {
  parsePlaneUserIDs,
  planeRevisionKey,
  shouldReseedPlaneForm,
  toPlaneAppInput,
  toPlaneFormState,
} from './form-state'

function app(overrides: Partial<PlaneApp> = {}): PlaneApp {
  return {
    id: 'app-1',
    displayName: 'Support',
    providerProfile: 'cloud',
    clientId: 'client-1',
    agentId: 'support',
    inboundEnabled: false,
    allowedUserIds: [],
    revision: 3n,
    credentialState: 1,
    clientSecretSet: false,
    webhookSecretSet: false,
    $typeName: 'agents.v1.PlaneApp',
    ...overrides,
  } as PlaneApp
}

describe('toPlaneFormState', () => {
  it('starts with inbound off and no secrets, even for an App with secrets', () => {
    const state = toPlaneFormState(
      app({ clientSecretSet: true, webhookSecretSet: true })
    )
    // The form has no inbound field at all: #497 cannot enable it, so there
    // is nothing for a user to flip.
    expect('inboundEnabled' in state).toBe(false)
    expect(state.clientSecret).toBe('')
    expect(state.webhookSecret).toBe('')
  })

  it('seeds the Agent from a locked Agent entry when there is no App yet', () => {
    expect(toPlaneFormState(undefined, 'support').agentId).toBe('support')
    // An existing App's binding always wins over the entry's Agent.
    expect(
      toPlaneFormState(app({ agentId: 'research' }), 'support').agentId
    ).toBe('research')
  })
})

describe('toPlaneAppInput', () => {
  it('keeps the locked Agent even when the form says otherwise', () => {
    const input = toPlaneAppInput(
      { ...toPlaneFormState(app()), agentId: 'someone-else' },
      { lockedAgentId: 'support' }
    )
    expect(input.agentId).toBe('support')
  })

  it('carries the revision on edit so a stale form loses instead of overwriting', () => {
    const input = toPlaneAppInput(toPlaneFormState(app()), {
      id: 'app-1',
      revision: 3n,
    })
    expect(input.id).toBe('app-1')
    expect(input.revision).toBe(3n)
  })

  it('never sends secrets or an inbound flag', () => {
    const form = {
      ...toPlaneFormState(app()),
      clientSecret: 'client-secret',
      webhookSecret: 'webhook-secret',
    }
    const input = toPlaneAppInput(form)
    expect(JSON.stringify(input)).not.toContain('client-secret')
    expect(JSON.stringify(input)).not.toContain('webhook-secret')
    expect('inboundEnabled' in input).toBe(false)
  })

  it('translates an empty max run time to absent and 0 to unlimited', () => {
    const base = toPlaneFormState(app())
    expect(toPlaneAppInput({ ...base, maxRunSeconds: '' }).maxRunSeconds).toBe(
      undefined
    )
    expect(toPlaneAppInput({ ...base, maxRunSeconds: '0' }).maxRunSeconds).toBe(
      0
    )
    expect(
      toPlaneAppInput({ ...base, maxRunSeconds: 'not-a-number' }).maxRunSeconds
    ).toBe(undefined)
  })

  it('canonicalises the allowlist input', () => {
    expect(parsePlaneUserIDs(' a \n b, c  ')).toEqual(['a', 'b', 'c'])
  })
})

describe('shouldReseedPlaneForm', () => {
  it('re-seeds a clean form when a newer revision arrives', () => {
    expect(
      shouldReseedPlaneForm({
        dirty: false,
        syncedKey: 'app-1:1',
        nextKey: 'app-1:2',
      })
    ).toBe(true)
  })

  it('never overwrites unsaved input on a background refresh', () => {
    expect(
      shouldReseedPlaneForm({
        dirty: true,
        syncedKey: 'app-1:1',
        nextKey: 'app-1:2',
      })
    ).toBe(false)
  })

  it('ignores a refetch that did not change the revision', () => {
    expect(
      shouldReseedPlaneForm({
        dirty: false,
        syncedKey: 'app-1:1',
        nextKey: 'app-1:1',
      })
    ).toBe(false)
  })

  it('only replaces a dirty form after an explicit conflict reload', () => {
    expect(
      shouldReseedPlaneForm({
        dirty: true,
        syncedKey: 'app-1:1',
        nextKey: 'app-1:2',
        force: true,
      })
    ).toBe(true)
  })

  it('does not switch identities just because a refresh landed', () => {
    expect(
      shouldReseedPlaneForm({
        dirty: true,
        syncedKey: 'app-1:1',
        nextKey: 'app-2:1',
      })
    ).toBe(false)
  })

  it('keys revisions by identity and revision', () => {
    expect(planeRevisionKey(app())).toBe('app-1:3')
    expect(planeRevisionKey(undefined)).toBeNull()
  })
})
