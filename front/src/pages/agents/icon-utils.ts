export function agentIconUrl(agent: { metadata?: Record<string, string> }) {
  return agent.metadata?.icon_url || agent.metadata?.avatar_url || "";
}
