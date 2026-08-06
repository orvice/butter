import { createFileRoute } from '@tanstack/react-router'
import { ChannelCreate } from '@/features/channels/create'

export const Route = createFileRoute('/_authenticated/channels/create')({
  component: ChannelCreate,
})
