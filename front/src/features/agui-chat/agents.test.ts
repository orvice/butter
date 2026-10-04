import type { Agent } from '@/types/api'
import { afterEach, describe, expect, it, vi } from 'vitest'
import {
  draftAgent,
  isSelectableAgent,
  lastAgentKey,
  readLastAgent,
  rememberLastAgent,
} from './agents'

const agent = (agent_id: string, extra: Partial<Agent> = {}): Agent => ({
  name: agent_id,
  agent_id,
  ...extra,
})

describe('isSelectableAgent', () => {
  it('keeps runnable agents with an agent_id', () => {
    expect(isSelectableAgent(agent('a'))).toBe(true)
    expect(
      isSelectableAgent(
        agent('a', { lifecycle_status: 'AGENT_LIFECYCLE_STATUS_ACTIVE' })
      )
    ).toBe(true)
    expect(
      isSelectableAgent(
        agent('a', { lifecycle_status: 'AGENT_LIFECYCLE_STATUS_UNSPECIFIED' })
      )
    ).toBe(true)
  })

  it('drops agents that cannot run or cannot be addressed', () => {
    expect(
      isSelectableAgent(
        agent('a', { lifecycle_status: 'AGENT_LIFECYCLE_STATUS_DELETED' })
      )
    ).toBe(false)
    expect(isSelectableAgent({ name: 'no id' })).toBe(false)
  })
})

describe('draftAgent', () => {
  const agents = [agent('a'), agent('b'), agent('c')]

  it('starts with the agent the URL names', () => {
    expect(draftAgent(agents, 'b', 'c')?.agent_id).toBe('b')
  })

  it('falls back to the agent picked last', () => {
    expect(draftAgent(agents, undefined, 'c')?.agent_id).toBe('c')
    // An agent the URL names that cannot be opened is skipped.
    expect(draftAgent(agents, 'gone', 'c')?.agent_id).toBe('c')
  })

  it('starts with no agent rather than one nobody chose', () => {
    expect(draftAgent(agents, undefined, null)).toBeNull()
    expect(draftAgent(agents, 'gone', 'also-gone')).toBeNull()
    expect(draftAgent([], 'a', 'a')).toBeNull()
  })
})

describe('last picked agent', () => {
  afterEach(() => {
    vi.unstubAllGlobals()
  })

  function stubStorage() {
    const items = new Map<string, string>()
    vi.stubGlobal('localStorage', {
      getItem: (key: string) => items.get(key) ?? null,
      setItem: (key: string, value: string) => void items.set(key, value),
    })
    return items
  }

  it('is remembered per workspace under the key Chat uses', () => {
    const items = stubStorage()
    rememberLastAgent('ws-1', 'b')
    expect(readLastAgent('ws-1')).toBe('b')
    expect(readLastAgent('ws-2')).toBeNull()
    expect(lastAgentKey('ws-1')).toBe('butter_chat_last_agent_ws-1')
    expect([...items.keys()]).toEqual(['butter_chat_last_agent_ws-1'])
  })

  it('is not remembered without a workspace or storage', () => {
    const items = stubStorage()
    rememberLastAgent('', 'b')
    expect(readLastAgent('')).toBeNull()
    expect(items.size).toBe(0)

    vi.stubGlobal('localStorage', {
      getItem: () => {
        throw new Error('blocked')
      },
      setItem: () => {
        throw new Error('blocked')
      },
    })
    expect(() => rememberLastAgent('ws-1', 'b')).not.toThrow()
    expect(readLastAgent('ws-1')).toBeNull()
  })
})
