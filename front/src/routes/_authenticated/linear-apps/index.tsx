import { createFileRoute } from '@tanstack/react-router'
import { LinearAppList } from '@/features/linear/list'

export const Route = createFileRoute('/_authenticated/linear-apps/')({
  component: LinearAppList,
})
