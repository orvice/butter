// Client-side mirror of the Agent ID slug rules enforced by the backend
// (internal/agent/agentid.go). The server remains the source of truth; this
// exists to give immediate feedback before submitting AssignAgentID.

export const AGENT_ID_PATTERN = /^[a-z0-9](?:[a-z0-9-]{0,62}[a-z0-9])?$/

export const RESERVED_AGENT_IDS = new Set([
  'user',
  'system',
  'admin',
  'start',
  'default',
  'api',
  'new',
])

// suggestAgentID derives a slug candidate from an agent display name:
// lowercase, non-alphanumeric runs collapsed to single hyphens, trimmed to
// the 64-char limit without leaving a dangling hyphen.
export function suggestAgentID(name: string): string {
  let slug = name
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, '-')
    .replace(/^-+|-+$/g, '')
  if (slug.length > 64) {
    slug = slug.slice(0, 64).replace(/-+$/g, '')
  }
  if (RESERVED_AGENT_IDS.has(slug)) {
    slug = `${slug}-agent`
  }
  return slug
}

// validateAgentID returns an error message, or null when the slug is valid.
export function validateAgentID(id: string): string | null {
  if (!id) return 'Agent ID is required.'
  if (!AGENT_ID_PATTERN.test(id)) {
    return 'Must be 1–64 lowercase letters, digits, or hyphens, starting and ending with a letter or digit.'
  }
  if (RESERVED_AGENT_IDS.has(id)) {
    return `"${id}" is a reserved word and cannot be used.`
  }
  return null
}
