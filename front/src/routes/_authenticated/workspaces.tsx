import { createFileRoute } from '@tanstack/react-router'
import { WorkspacePage } from '@/features/workspaces'

export const Route = createFileRoute('/_authenticated/workspaces')({
  component: WorkspacePage,
})
