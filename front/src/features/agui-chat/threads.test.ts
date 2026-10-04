import type { SessionInfo } from '@/types/api'
import { describe, expect, it } from 'vitest'
import {
  agentThreads,
  boundAgentId,
  resolveThreadView,
  sessionIdOf,
  threadBinding,
  threadIdOf,
  threadTitle,
  type ThreadLookup,
} from './threads'

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

describe('sessionIdOf', () => {
  it('is the session a thread lives in', () => {
    expect(sessionIdOf('abc')).toBe('agui-abc')
    expect(threadIdOf(bound('abc'))).toBe('abc')
  })
})

describe('boundAgentId', () => {
  it('names the agent of this workspace’s thread', () => {
    expect(boundAgentId(bound('t1'), 'ws-1', 't1')).toBe('agent-a')
  })

  it('is null for a session that is not this thread here', () => {
    // Another workspace's thread.
    expect(
      boundAgentId(bound('t1', 'agent-a', 'ws-2'), 'ws-1', 't1')
    ).toBeNull()
    // Another thread's session.
    expect(boundAgentId(bound('t1'), 'ws-1', 't2')).toBeNull()
    // A binding for another thread.
    const mismatched = aguiSession('t1', {
      workspace_id: 'ws-1',
      agent_id: 'agent-a',
      thread_id: 'elsewhere',
    })
    expect(boundAgentId(mismatched, 'ws-1', 't1')).toBeNull()
    // No binding: the thread predates A2UI.
    expect(boundAgentId(aguiSession('t1', undefined), 'ws-1', 't1')).toBeNull()
    // Another app.
    expect(
      boundAgentId({ ...bound('t1'), app_name: 'web-chat' }, 'ws-1', 't1')
    ).toBeNull()
  })
})

describe('resolveThreadView', () => {
  const lookup = (extra: Partial<ThreadLookup> = {}): ThreadLookup => ({
    threadId: 't1',
    workspaceId: 'ws-1',
    agentIds: ['agent-a', 'agent-b'],
    ...extra,
  })

  it('opens a thread with the agent its binding names', () => {
    const session = bound('t1')
    expect(resolveThreadView(lookup({ session }))).toEqual({
      kind: 'open',
      agentId: 'agent-a',
      session,
    })
  })

  it('opens from the thread list before the session is read', () => {
    const listed = bound('t1')
    expect(resolveThreadView(lookup({ listed }))).toEqual({
      kind: 'open',
      agentId: 'agent-a',
      session: listed,
    })
    // A failed read still opens a listed thread.
    expect(
      resolveThreadView(lookup({ listed, sessionError: new Error('down') }))
    ).toMatchObject({ kind: 'open', agentId: 'agent-a' })
  })

  it('is loading until the session and the agents are read', () => {
    expect(resolveThreadView(lookup())).toEqual({ kind: 'loading' })
    expect(
      resolveThreadView(lookup({ session: bound('t1'), agentIds: undefined }))
    ).toEqual({ kind: 'loading' })
  })

  it('fails when the session or the agents cannot be read', () => {
    const error = new Error('unavailable')
    expect(resolveThreadView(lookup({ sessionError: error }))).toEqual({
      kind: 'failed',
      error,
    })
    expect(
      resolveThreadView(
        lookup({
          session: bound('t1'),
          agentIds: undefined,
          agentsError: error,
        })
      )
    ).toEqual({ kind: 'failed', error })
  })

  it('is not found without a session, even if the thread list still has it', () => {
    expect(resolveThreadView(lookup({ session: null }))).toEqual({
      kind: 'not-found',
    })
    expect(
      resolveThreadView(lookup({ session: null, listed: bound('t1') }))
    ).toEqual({ kind: 'not-found' })
  })

  it('is not found for a thread of another workspace or agent', () => {
    const notFound = { kind: 'not-found' }
    expect(
      resolveThreadView(lookup({ session: bound('t1', 'agent-a', 'ws-2') }))
    ).toEqual(notFound)
    // Bound to an agent this workspace cannot run.
    expect(
      resolveThreadView(lookup({ session: bound('t1', 'agent-gone') }))
    ).toEqual(notFound)
    // Bound to another agent than the URL names.
    expect(
      resolveThreadView(
        lookup({ session: bound('t1'), requestedAgentId: 'agent-b' })
      )
    ).toEqual(notFound)
    // Without a binding the agent is unknown.
    expect(
      resolveThreadView(lookup({ session: aguiSession('t1', undefined) }))
    ).toEqual(notFound)
  })

  it('opens a thread started here before its session exists', () => {
    expect(
      resolveThreadView(
        lookup({
          startedAgentId: 'agent-b',
          session: null,
          agentIds: undefined,
        })
      )
    ).toEqual({ kind: 'open', agentId: 'agent-b', session: null })
    const session = bound('t1', 'agent-b')
    expect(
      resolveThreadView(lookup({ startedAgentId: 'agent-b', session }))
    ).toEqual({ kind: 'open', agentId: 'agent-b', session })
  })
})
