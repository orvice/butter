import { useNavigate } from '@tanstack/react-router'
import { toast } from 'sonner'
import { useCreatePlaneApp, usePlaneProviderProfile } from '@/api/plane'
import { Page, PageHeader, PageScroll } from '@/components/butter/page-parts'
import { PlaneAppForm, type PlaneAppFormValues } from './form'
import { PlaneProviderNotice } from './shared'

export function PlaneAppCreate() {
  const navigate = useNavigate()
  const createMutation = useCreatePlaneApp()
  const { data: profile } = usePlaneProviderProfile()

  function onSubmit(values: PlaneAppFormValues) {
    createMutation.mutate(
      {
        displayName: values.displayName,
        clientId: values.clientId,
        agentId: values.agentId,
        allowedUserIds: values.allowedUserIds,
        maxRunSeconds: values.maxRunSeconds,
        clientSecret: values.clientSecret,
        webhookSecret: values.webhookSecret,
      },
      {
        onSuccess: (app) => {
          toast.success('Plane App registered')
          navigate({ to: '/plane-apps/$id/edit', params: { id: app.id } })
        },
        onError: (err) => toast.error(err.message),
      }
    )
  }

  return (
    <Page>
      <PageHeader
        className='max-w-3xl'
        title='Register Plane App'
        subtitle='Bind a Plane app to an Agent. Inbound stays off in this version.'
      />
      <PageScroll className='max-w-3xl'>
        <div className='space-y-6'>
          {profile && (
            <PlaneProviderNotice
              name={profile.name}
              apiBaseUrl={profile.apiBaseUrl}
              edition={profile.edition}
            />
          )}
          <PlaneAppForm
            mode='create'
            submitLabel='Register'
            loading={createMutation.isPending}
            onCancel={() => navigate({ to: '/plane-apps' })}
            onSubmit={onSubmit}
          />
        </div>
      </PageScroll>
    </Page>
  )
}
