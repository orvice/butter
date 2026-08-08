import { createFileRoute } from '@tanstack/react-router'
import { AdminGitHostsPage } from '@/features/admin/git-hosts'

export const Route = createFileRoute('/_authenticated/admin/git-hosts')({
  component: AdminGitHostsPage,
})
