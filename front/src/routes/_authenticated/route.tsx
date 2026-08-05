import { useEffect } from 'react'
import { createFileRoute, redirect } from '@tanstack/react-router'
import { useAuthStore } from '@/stores/auth-store'
import { AuthenticatedLayout } from '@/components/layout/authenticated-layout'

export const Route = createFileRoute('/_authenticated')({
  beforeLoad: ({ location }) => {
    const { token } = useAuthStore.getState().auth
    if (!token) {
      throw redirect({
        to: '/sign-in',
        search: { redirect: location.href },
      })
    }
  },
  component: AuthenticatedGuard,
})

function AuthenticatedGuard() {
  const { auth } = useAuthStore()

  // Restore the user profile (`Me`) when we only have a stored token,
  // e.g. after a full page reload.
  useEffect(() => {
    void auth.restore()
  }, [auth])

  return <AuthenticatedLayout />
}
