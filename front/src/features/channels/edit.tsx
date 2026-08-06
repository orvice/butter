import { useNavigate, useParams } from '@tanstack/react-router'
import { toast } from 'sonner'
import { useChannel, useUpdateChannel } from '@/api/channels'
import { Page, PageHeader, PageScroll } from '@/components/butter/page-parts'
import { Skeleton } from '@/components/ui/skeleton'
import { ChannelForm } from './form'
import type { AgentChannel } from '@/types/api'

export function ChannelEdit() {
  const { name } = useParams({ from: '/_authenticated/channels/$name/edit' })
  const navigate = useNavigate()
  const { data, isLoading } = useChannel(name)
  const updateMutation = useUpdateChannel()

  function onSubmit(channel: AgentChannel) {
    updateMutation.mutate(channel, {
      onSuccess: () => {
        toast.success('Channel updated')
        navigate({ to: '/channels' })
      },
      onError: (err) => toast.error(err.message),
    })
  }

  return (
    <Page>
      <PageHeader className='max-w-3xl' title='Edit Channel' subtitle={name} />
      <PageScroll className='max-w-3xl'>
        {isLoading ? (
          <Skeleton className='h-96' />
        ) : (
          <ChannelForm
            mode='edit'
            initialValue={data?.channel}
            submitLabel='Save'
            loading={updateMutation.isPending}
            onCancel={() => navigate({ to: '/channels' })}
            onSubmit={onSubmit}
          />
        )}
      </PageScroll>
    </Page>
  )
}
