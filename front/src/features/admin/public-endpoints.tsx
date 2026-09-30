import { Navigate } from '@tanstack/react-router'
import { useAuth } from '@/hooks/use-auth'
import { Page, PageHeader, PageScroll } from '@/components/butter/page-parts'
import { LinearPublicBaseUrlCard } from './linear-settings'
import { TelegramWebhookBaseUrlCard } from './telegram-settings'

/**
 * The public base URLs external platforms call back to. They are
 * platform-level, not per-workspace, and usually the same address; they are
 * kept apart because each platform's settings live in their own service.
 */
export function AdminPublicEndpointsPage() {
  const { isAdmin, isLoading: isAuthLoading } = useAuth()

  if (!isAuthLoading && !isAdmin) return <Navigate to='/403' />

  return (
    <Page>
      <PageHeader
        title='Public endpoints'
        subtitle='The public address of this deployment that Telegram and Linear deliver callbacks to, for every workspace.'
      />
      <PageScroll>
        <div className='grid max-w-2xl gap-4'>
          <TelegramWebhookBaseUrlCard />
          <LinearPublicBaseUrlCard />
        </div>
      </PageScroll>
    </Page>
  )
}
