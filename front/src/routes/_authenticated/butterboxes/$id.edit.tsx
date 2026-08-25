import { createFileRoute } from '@tanstack/react-router'
import { ButterBoxEdit } from '@/features/butterboxes/edit'

export const Route = createFileRoute('/_authenticated/butterboxes/$id/edit')({
  component: ButterBoxEdit,
})
