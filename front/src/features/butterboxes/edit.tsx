import { useState } from 'react'
import { useNavigate, useParams } from '@tanstack/react-router'
import { timestampDate } from '@bufbuild/protobuf/wkt'
import { ExternalLink, KeyRound } from 'lucide-react'
import { toast } from 'sonner'
import {
  useButterBox,
  useSetButterBoxToken,
  useUpdateButterBox,
} from '@/api/butterboxes'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Skeleton } from '@/components/ui/skeleton'
import { Page, PageHeader, PageScroll } from '@/components/butter/page-parts'
import { ButterBoxForm, type ButterBoxFormValues } from './form'
import { BoxStatusLine } from './status-cell'

export function ButterBoxEdit() {
  const { id } = useParams({ from: '/_authenticated/butterboxes/$id/edit' })
  const navigate = useNavigate()
  const { data: box, isLoading } = useButterBox(id)
  const updateMutation = useUpdateButterBox()
  const setToken = useSetButterBoxToken()
  const [rotationToken, setRotationToken] = useState('')

  function onSubmit(values: ButterBoxFormValues) {
    updateMutation.mutate(
      {
        id,
        name: values.name,
        baseUrl: values.baseUrl,
        enabled: values.enabled,
      },
      {
        onSuccess: () => {
          toast.success('ButterBox updated')
          navigate({ to: '/butterboxes' })
        },
        onError: (err) => toast.error(err.message),
      }
    )
  }

  function rotateToken() {
    setToken.mutate(
      { id, token: rotationToken.trim() },
      {
        onSuccess: () => {
          toast.success('Access token updated')
          setRotationToken('')
        },
        onError: (err) => toast.error(err.message),
      }
    )
  }

  function clearToken() {
    setToken.mutate(
      { id, token: '' },
      {
        onSuccess: () => toast.success('Access token cleared'),
        onError: (err) => toast.error(err.message),
      }
    )
  }

  if (isLoading) {
    return (
      <div className='p-6'>
        <Skeleton className='h-96 w-full' />
      </div>
    )
  }

  return (
    <Page>
      <PageHeader
        className='max-w-3xl'
        title='Edit ButterBox'
        subtitle={box ? <BoxStatusLine id={box.id} /> : id}
        actions={
          box ? (
            <Button size='sm' variant='outline' asChild>
              <a href={box.baseUrl} target='_blank' rel='noreferrer'>
                <ExternalLink className='size-4' />
                pi-web
              </a>
            </Button>
          ) : undefined
        }
      />
      <PageScroll className='max-w-3xl'>
        <div className='space-y-6'>
          <ButterBoxForm
            mode='edit'
            submitLabel='Save'
            loading={updateMutation.isPending}
            initialValue={box}
            onCancel={() => navigate({ to: '/butterboxes' })}
            onSubmit={onSubmit}
          />

          <Card>
            <CardHeader className='pb-2'>
              <CardTitle className='flex items-center gap-2 text-base'>
                <KeyRound className='h-4 w-4' />
                Rotate access token
              </CardTitle>
            </CardHeader>
            <CardContent className='space-y-3'>
              <p className='text-sm text-muted-foreground'>
                The token is write-only: it is encrypted at rest and never
                displayed again.{' '}
                {box?.credentialSet
                  ? `A token is currently set${
                      box.credentialUpdatedAt
                        ? ` (updated ${timestampDate(box.credentialUpdatedAt).toLocaleString()})`
                        : ''
                    }.`
                  : 'No token is currently set.'}
              </p>
              <Input
                type='password'
                autoComplete='new-password'
                aria-label='New access token'
                value={rotationToken}
                onChange={(e) => setRotationToken(e.target.value)}
                placeholder=''
              />
              <div className='flex justify-end gap-2'>
                <Button
                  variant='outline'
                  onClick={clearToken}
                  disabled={!box?.credentialSet || setToken.isPending}
                >
                  Clear token
                </Button>
                <Button
                  variant='outline'
                  onClick={rotateToken}
                  disabled={!rotationToken.trim() || setToken.isPending}
                >
                  Rotate
                </Button>
              </div>
            </CardContent>
          </Card>
        </div>
      </PageScroll>
    </Page>
  )
}
