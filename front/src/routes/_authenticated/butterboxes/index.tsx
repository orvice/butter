import { createFileRoute } from '@tanstack/react-router'
import { ButterBoxList } from '@/features/butterboxes/list'

export const Route = createFileRoute('/_authenticated/butterboxes/')({
  component: ButterBoxList,
})
