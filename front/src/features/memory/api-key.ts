/**
 * The mem0 API key field is write-only: an empty field keeps the stored
 * key, "clear" removes it, and anything typed sets or rotates it. The
 * result maps onto PutWorkspaceMemoryConfigRequest.api_key presence.
 */
export function apiKeyForSave(values: {
  apiKey: string
  clearKey: boolean
}): string | undefined {
  if (values.clearKey) return ''
  const key = values.apiKey.trim()
  return key === '' ? undefined : key
}
