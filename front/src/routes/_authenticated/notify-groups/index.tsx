import { createFileRoute } from '@tanstack/react-router'
import { NotifyGroupList } from '@/features/notify-groups/list'

export const Route = createFileRoute('/_authenticated/notify-groups/')({
  component: NotifyGroupList,
})
