import { createFileRoute } from '@tanstack/react-router'
import { AgentList } from '@/features/agents/list'

export const Route = createFileRoute('/_authenticated/agents/')({
  component: AgentList,
})
