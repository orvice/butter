type MemberLike = { userId: string; role: string }

/**
 * Mirrors the server's manage-role rule for workspace resources: global
 * admins, and workspace members whose role is "owner" or "admin". The
 * server enforces it; this only decides whether the UI offers edits.
 */
export function canManageWorkspace(
  isAdmin: boolean,
  userId: string | undefined,
  members: readonly MemberLike[] | undefined
): boolean {
  if (isAdmin) return true
  if (!userId || !members) return false
  const role = members.find((m) => m.userId === userId)?.role
  return role === 'owner' || role === 'admin'
}
