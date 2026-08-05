import { createFileRoute } from '@tanstack/react-router'
import { ForumListPage } from '@/features/forum/list'

export const Route = createFileRoute('/_authenticated/forum/')({
  component: ForumListPage,
})
