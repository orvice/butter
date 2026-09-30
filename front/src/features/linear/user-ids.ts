export function parseUserIds(raw: string): string[] {
  return raw
    .split(/[\s,]+/)
    .map((entry) => entry.trim())
    .filter(Boolean)
}

export function formatUserIds(ids: string[] | undefined): string {
  return (ids ?? []).join('\n')
}
