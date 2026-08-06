import { useEffect } from 'react'
import { createFileRoute, Outlet, redirect } from '@tanstack/react-router'
import { useAuthStore } from '@/stores/auth-store'
import { WorkspaceProvider } from '@/context/workspace-provider'
import { AuthenticatedLayout } from '@/components/layout/authenticated-layout'
import { WorkspaceGate } from '@/components/layout/workspace-gate'

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

  return (
    <WorkspaceProvider>
      <AuthenticatedLayout>
        <WorkspaceGate>
          <Outlet />
        </WorkspaceGate>
      </AuthenticatedLayout>
    </WorkspaceProvider>
  )
}
