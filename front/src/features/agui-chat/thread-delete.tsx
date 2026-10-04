import {
  createContext,
  useContext,
  useMemo,
  useRef,
  useState,
  type ReactNode,
  type RefObject,
} from 'react'
import { useNavigate } from '@tanstack/react-router'
import { toast } from 'sonner'
import { useDeleteSession } from '@/api/sessions'
import { DeleteDialog } from '@/components/delete-dialog'
import type { ButterAGUIAgent } from './a2ui/agent'
import { sessionIdOf } from './threads'
import { useAGUIChatSearch } from './use-threads'

// SessionAddress names one session for SessionService.
export interface SessionAddress {
  app_name: string
  user_id: string
  session_id: string
}

// ThreadDeleteTarget is a thread whose deletion is asked for.
export interface ThreadDeleteTarget {
  session: SessionAddress
  title: string
  // agentId is the agent a new-chat draft starts with if the open thread
  // goes.
  agentId: string | null
}

interface ThreadDelete {
  // requestDelete asks the user to confirm deleting a thread.
  requestDelete: (target: ThreadDeleteTarget) => void
  // openClient holds the AG-UI client of the thread AG-UI Chat has open, so
  // deleting that thread stops its run first.
  openClient: RefObject<ButterAGUIAgent | null>
}

const ThreadDeleteContext = createContext<ThreadDelete | null>(null)

// ThreadDeleteProvider is the one way a thread is deleted, from the sidebar's
// thread list or from the open thread's header: after confirmation it stops a
// run on the open thread, deletes the thread's session, and leaves a deleted
// open thread for a new-chat draft with its agent. It sits above both, so the
// sidebar can delete the thread the page has open.
export function ThreadDeleteProvider({ children }: { children: ReactNode }) {
  const navigate = useNavigate()
  const openThreadId = useAGUIChatSearch()?.thread ?? null
  const openClient = useRef<ButterAGUIAgent | null>(null)
  const deleteMutation = useDeleteSession()
  const [target, setTarget] = useState<ThreadDeleteTarget | null>(null)
  const value = useMemo<ThreadDelete>(
    () => ({ requestDelete: setTarget, openClient }),
    []
  )

  const handleDeleteConfirm = () => {
    if (!target) return
    const isOpen =
      !!openThreadId && target.session.session_id === sessionIdOf(openThreadId)
    // Stop a run on the thread before its session goes away.
    if (isOpen) openClient.current?.abortRun()
    deleteMutation.mutate(target.session, {
      onSuccess: () => {
        toast.success('Thread deleted')
        setTarget(null)
        if (isOpen) {
          void navigate({
            to: '/agui-chat',
            search: target.agentId ? { agent: target.agentId } : {},
            replace: true,
          })
        }
      },
      onError: (err) => toast.error(err.message),
    })
  }

  return (
    <ThreadDeleteContext.Provider value={value}>
      {children}
      <DeleteDialog
        open={!!target}
        onOpenChange={(open) => !open && setTarget(null)}
        title='Delete thread'
        description={`Delete thread "${target?.title ?? ''}"? Its messages, cards and forms are removed. This cannot be undone.`}
        loading={deleteMutation.isPending}
        onConfirm={handleDeleteConfirm}
      />
    </ThreadDeleteContext.Provider>
  )
}

// eslint-disable-next-line react-refresh/only-export-components
export function useThreadDelete(): ThreadDelete {
  const value = useContext(ThreadDeleteContext)
  if (!value) throw new Error('useThreadDelete needs a ThreadDeleteProvider')
  return value
}
