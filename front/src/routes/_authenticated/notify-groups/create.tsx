import { createFileRoute } from '@tanstack/react-router'
import { NotifyGroupCreate } from '@/features/notify-groups/create'

export const Route = createFileRoute('/_authenticated/notify-groups/create')({
  component: NotifyGroupCreate,
})
