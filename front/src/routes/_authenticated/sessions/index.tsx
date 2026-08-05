import { createFileRoute } from '@tanstack/react-router'
import { SessionListPage } from '@/features/sessions/list'

export const Route = createFileRoute('/_authenticated/sessions/')({
  component: SessionListPage,
})
