import { createFileRoute } from '@tanstack/react-router'
import { RemoteAgentList } from '@/features/remote-agents/list'

export const Route = createFileRoute('/_authenticated/remote-agents/')({
  component: RemoteAgentList,
})
