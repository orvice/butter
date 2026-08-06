import { useNavigate } from '@tanstack/react-router'
import { toast } from 'sonner'
import { useCreateChannel } from '@/api/channels'
import { Page, PageHeader, PageScroll } from '@/components/butter/page-parts'
import { ChannelForm } from './form'
import type { AgentChannel } from '@/types/api'

export function ChannelCreate() {
  const navigate = useNavigate()
  const createMutation = useCreateChannel()

  function onSubmit(channel: AgentChannel) {
    createMutation.mutate(channel, {
      onSuccess: () => {
        toast.success('Channel created')
        navigate({ to: '/channels' })
      },
      onError: (err) => toast.error(err.message),
    })
  }

  return (
    <Page>
      <PageHeader
        className='max-w-3xl'
        title='Create Channel'
        subtitle='Bind an agent to a platform entry point like Telegram or Discord.'
      />
      <PageScroll className='max-w-3xl'>
        <ChannelForm
          mode='create'
          submitLabel='Create'
          loading={createMutation.isPending}
          onCancel={() => navigate({ to: '/channels' })}
          onSubmit={onSubmit}
        />
      </PageScroll>
    </Page>
  )
}
