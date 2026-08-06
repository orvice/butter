import { createFileRoute } from '@tanstack/react-router'
import { ChannelList } from '@/features/channels/list'

export const Route = createFileRoute('/_authenticated/channels/')({
  component: ChannelList,
})
