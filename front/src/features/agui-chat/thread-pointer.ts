// The dashboard remembers the current AG-UI thread per workspace, agent and
// signed-in user, so a refresh returns to the same conversation (its cards
// and unanswered forms come back from the UI snapshot). Nothing else is
// kept: the draft of an unsent form does not survive a refresh.

const PREFIX = 'butter:agui-thread:'

export function threadPointerKey(
  workspaceId: string,
  userId: string,
  agentId: string
): string {
  return `${PREFIX}${workspaceId}:${userId}:${agentId}`
}

export function readThreadPointer(key: string): string | null {
  try {
    return localStorage.getItem(key)
  } catch {
    return null
  }
}

export function writeThreadPointer(key: string, threadId: string) {
  try {
    localStorage.setItem(key, threadId)
  } catch {
    // Storage unavailable (private mode): the thread just won't survive a refresh.
  }
}

export function newThreadId(): string {
  return crypto.randomUUID()
}

// currentThread returns the remembered thread for key, starting (and
// remembering) a new one when there is none.
export function currentThread(key: string): string {
  const existing = readThreadPointer(key)
  if (existing) return existing
  const id = newThreadId()
  writeThreadPointer(key, id)
  return id
}
