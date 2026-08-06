import { createFileRoute } from '@tanstack/react-router'
import { AgentFilesPage } from '@/features/agent-files/list'

export const Route = createFileRoute('/_authenticated/agent-files')({
  component: AgentFilesPage,
})
