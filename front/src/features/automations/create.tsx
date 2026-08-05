import { Link, useNavigate } from '@tanstack/react-router'
import { toast } from 'sonner'
import { ArrowLeft, ChevronRight } from 'lucide-react'
import { useCreateCronJob } from '@/api/cron'
import { Page, PageHeader, PageScroll } from '@/components/butter/page-parts'
import { AutomationForm } from './form'
import type { CronJob } from '@/types/api'

export function AutomationCreatePage() {
  const navigate = useNavigate()
  const createMutation = useCreateCronJob()

  function handleSubmit(job: CronJob) {
    createMutation.mutate(job, {
      onSuccess: () => {
        toast.success('Automation created')
        navigate({ to: '/automations' })
      },
      onError: (err) => toast.error(err.message),
    })
  }

  return (
    <Page>
      <PageHeader
        className='max-w-3xl'
        title='Create Automation'
        subtitle='Run an agent on a schedule and deliver the result to a webhook, channel, or notify group.'
        breadcrumb={
          <Link to='/automations' className='inline-flex items-center gap-1.5 hover:text-foreground'>
            <ArrowLeft className='size-3.5' />
            Automations
            <ChevronRight className='size-3' />
            <span className='text-foreground'>Create</span>
          </Link>
        }
      />
      <PageScroll className='max-w-3xl'>
        <AutomationForm
          mode='create'
          loading={createMutation.isPending}
          submitLabel='Create Automation'
          onCancel={() => navigate({ to: '/automations' })}
          onSubmit={handleSubmit}
        />
      </PageScroll>
    </Page>
  )
}
