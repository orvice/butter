import { createFileRoute } from '@tanstack/react-router'
import { ForumThreadPage } from '@/features/forum/thread'

export const Route = createFileRoute('/_authenticated/forum/$id')({
  component: ForumThreadPage,
})
