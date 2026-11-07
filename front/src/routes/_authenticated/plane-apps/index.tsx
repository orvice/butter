import { createFileRoute } from '@tanstack/react-router'
import { PlaneAppList } from '@/features/plane/list'

export const Route = createFileRoute('/_authenticated/plane-apps/')({
  component: PlaneAppList,
})
