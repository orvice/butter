import { createFileRoute } from '@tanstack/react-router'
import { AgentEdit } from '@/features/agents/edit'

export const Route = createFileRoute('/_authenticated/agents/$name/edit')({
  component: AgentEdit,
})
