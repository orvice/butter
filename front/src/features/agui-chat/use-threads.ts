import { useMatch } from '@tanstack/react-router'
import { useAllSessions } from '@/api/sessions'
import { useAuthStore } from '@/stores/auth-store'
import { useWorkspace } from '@/context/workspace-provider'
import { AGUI_APP_NAME, THREAD_PAGE_SIZE } from './threads'

// useThreadSessions reads the caller's threads. The server keeps the listing
// to the caller's own `agui` sessions in the selected workspace, and every
// page is read, so no thread is dropped. The sidebar's thread list and Chat
// share this one query.
export function useThreadSessions() {
  const { selectedWorkspaceId } = useWorkspace()
  const userId = useAuthStore((state) => state.auth.user?.id ?? '')
  return useAllSessions(
    {
      app_name: AGUI_APP_NAME,
      user_id: userId || undefined,
      workspace_scoped: true,
      page_size: THREAD_PAGE_SIZE,
    },
    { enabled: !!userId && !!selectedWorkspaceId }
  )
}

// useChatSearch is Chat's URL state while the page is open: the thread it
// shows, or the agent its new-chat draft starts with. It is null on every
// other page.
export function useChatSearch() {
  return (
    useMatch({
      from: '/_authenticated/chat',
      shouldThrow: false,
      select: (match) => match.search,
    }) ?? null
  )
}
