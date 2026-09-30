import { createFileRoute } from '@tanstack/react-router'
import { LinearAppEdit } from '@/features/linear/edit'

export const Route = createFileRoute('/_authenticated/linear-apps/$id/edit')({
  component: LinearAppEdit,
})
