import type { SessionInfo } from '@/types/api'
import { describe, expect, it } from 'vitest'
import { agentThreads, threadBinding, threadIdOf, threadTitle } from './threads'

function aguiSession(
  threadId: string,
  binding: Record<string, unknown> | string | undefined,
  extra: Partial<SessionInfo> = {}
): SessionInfo {
  return {
    session_id: `agui-${threadId}`,
    app_name: 'agui',
    user_id: 'u1',
    state:
      binding === undefined
        ? {}
        : {
            'butter:a2ui:binding':
              typeof binding === 'string' ? binding : JSON.stringify(binding),
          },
    ...extra,
  }
}

function bound(threadId: string, agentId = 'agent-a', workspaceId = 'ws-1') {
  return aguiSession(threadId, {
    principal: 'u1',
    workspace_id: workspaceId,
    agent_id: agentId,
    thread_id: threadId,
  })
}

describe('threadBinding', () => {
  it('parses the JSON-string binding', () => {
    expect(threadBinding(bound('t1'))).toEqual({
      workspaceId: 'ws-1',
      agentId: 'agent-a',
      threadId: 't1',
    })
  })

  it('reads a session without a binding, or with a broken one, as unbound', () => {
    expect(threadBinding(aguiSession('t1', undefined))).toBeNull()
    expect(threadBinding(aguiSession('t1', '{not json'))).toBeNull()
    expect(threadBinding(aguiSession('t1', { agent_id: 'agent-a' }))).toBeNull()
  })
})

describe('threadIdOf', () => {
  it('strips the agui- session prefix', () => {
    expect(threadIdOf(bound('abc'))).toBe('abc')
    expect(threadIdOf({ ...bound('abc'), session_id: 'other' })).toBeNull()
  })
})

describe('agentThreads', () => {
  it('keeps only this agent and workspace, in the given order', () => {
    const sessions = [
      bound('t3'),
      bound('t2', 'agent-b'),
      bound('t1'),
      bound('t0', 'agent-a', 'ws-2'),
    ]
    expect(
      agentThreads(sessions, 'ws-1', 'agent-a').map((s) => s.session_id)
    ).toEqual(['agui-t3', 'agui-t1'])
  })

  it('drops unbound threads, other apps, and bindings for another thread', () => {
    const mismatched = aguiSession('t2', {
      workspace_id: 'ws-1',
      agent_id: 'agent-a',
      thread_id: 'elsewhere',
    })
    const sessions = [
      aguiSession('t1', undefined),
      { ...bound('t3'), app_name: 'web-chat' },
      mismatched,
    ]
    expect(agentThreads(sessions, 'ws-1', 'agent-a')).toEqual([])
  })
})

describe('threadTitle', () => {
  it('falls back for an untitled thread', () => {
    expect(threadTitle({ ...bound('t1'), title: '  Trip plan ' })).toBe(
      'Trip plan'
    )
    expect(threadTitle(bound('t1'))).toBe('Untitled thread')
  })
})
