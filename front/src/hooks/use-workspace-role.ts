import { useWorkspaceMembers } from '@/api/workspaces'
import { canManageWorkspace } from '@/lib/workspace-role'
import { useAuth } from '@/hooks/use-auth'
import { useWorkspace } from '@/hooks/use-workspace'

/** Whether the signed-in user may manage the selected workspace's resources. */
export function useCanManageWorkspace(): {
  canManage: boolean
  isLoading: boolean
} {
  const { user, isAdmin } = useAuth()
  const { selectedWorkspaceId } = useWorkspace()
  const { data, isLoading } = useWorkspaceMembers(
    isAdmin ? '' : selectedWorkspaceId
  )
  return {
    canManage: canManageWorkspace(isAdmin, user?.id, data?.members),
    isLoading: !isAdmin && isLoading,
  }
}
