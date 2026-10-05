import type { Agent } from '@/types/api'
import { CHAT_LAST_AGENT_PREFIX } from '@/lib/constants'

// isSelectableAgent: every runnable agent can be opened here. enable_agui is
// not needed: it only gates programmatic AG-UI access (API and root tokens),
// not signed-in users.
export function isSelectableAgent(a: Agent): boolean {
  const status = a.lifecycle_status
  const runnable =
    !status ||
    status === 'AGENT_LIFECYCLE_STATUS_UNSPECIFIED' ||
    status === 'AGENT_LIFECYCLE_STATUS_ACTIVE'
  return runnable && !!a.agent_id
}

// draftAgent is the agent a new thread starts with: the one the URL names,
// else the one picked last in this workspace, else none, so a thread never
// starts with an agent nobody chose. Each counts only while it is among
// agents (the selectable ones).
export function draftAgent(
  agents: readonly Agent[],
  requested: string | undefined,
  lastPicked: string | null
): Agent | null {
  for (const id of [requested, lastPicked]) {
    if (!id) continue
    const match = agents.find((a) => a.agent_id === id)
    if (match) return match
  }
  return null
}

// The agent picked last for a new chat is remembered per workspace, under
// the key the Chat before AG-UI used (#409), so a new chat still starts from
// the agent picked there.
export function lastAgentKey(workspaceId: string): string {
  return `${CHAT_LAST_AGENT_PREFIX}${workspaceId}`
}

export function readLastAgent(workspaceId: string): string | null {
  if (!workspaceId) return null
  try {
    return localStorage.getItem(lastAgentKey(workspaceId))
  } catch {
    return null
  }
}

export function rememberLastAgent(workspaceId: string, agentId: string) {
  if (!workspaceId) return
  try {
    localStorage.setItem(lastAgentKey(workspaceId), agentId)
  } catch {
    // Storage unavailable (private mode): the pick is just not remembered.
  }
}
