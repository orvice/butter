import { createFileRoute } from '@tanstack/react-router'
import { DaemonListPage } from '@/features/daemons/list'

export const Route = createFileRoute('/_authenticated/daemons')({
  component: DaemonListPage,
})
