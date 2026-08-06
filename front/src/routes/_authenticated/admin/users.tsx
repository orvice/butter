import { createFileRoute } from '@tanstack/react-router'
import { UserListPage } from '@/features/users/list'

export const Route = createFileRoute('/_authenticated/admin/users')({
  component: UserListPage,
})
