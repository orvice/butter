import { createFileRoute } from '@tanstack/react-router'
import { NotifyGroupEdit } from '@/features/notify-groups/edit'

export const Route = createFileRoute('/_authenticated/notify-groups/$name/edit')({
  component: NotifyGroupEdit,
})
