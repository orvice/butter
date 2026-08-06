import { createFileRoute } from '@tanstack/react-router'
import { RemoteAgentEdit } from '@/features/remote-agents/edit'

export const Route = createFileRoute('/_authenticated/remote-agents/$id/edit')({
  component: RemoteAgentEdit,
})
