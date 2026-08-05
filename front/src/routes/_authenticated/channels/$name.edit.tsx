import { createFileRoute } from '@tanstack/react-router'
import { ChannelEdit } from '@/features/channels/edit'

export const Route = createFileRoute('/_authenticated/channels/$name/edit')({
  component: ChannelEdit,
})
