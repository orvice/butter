import { useState } from 'react'
import { toast } from 'sonner'
import { useWorkspace } from '@/context/workspace-provider'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'

/**
 * Blocks workspace-scoped content until a workspace is selected. When the
 * user has no workspaces at all, shows the first-workspace creation card.
 */
export function WorkspaceGate({ children }: { children: React.ReactNode }) {
  const { selectedWorkspaceId, workspaces, isLoading } = useWorkspace()

  if (selectedWorkspaceId) return <>{children}</>

  if (isLoading) {
    return (
      <div className='flex h-full items-center justify-center'>
        <p className='text-sm text-muted-foreground'>Loading workspaces…</p>
      </div>
    )
  }

  if (workspaces.length === 0) {
    return <WorkspaceCreateCard />
  }

  return (
    <div className='flex h-full items-center justify-center'>
      <p className='text-sm text-muted-foreground'>Preparing your workspace…</p>
    </div>
  )
}

function WorkspaceCreateCard() {
  const { createWorkspace, isCreating } = useWorkspace()
  const [name, setName] = useState('Default')
  const [slug, setSlug] = useState('default')
  const [description, setDescription] = useState('')

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    const trimmedName = name.trim()
    const trimmedSlug = slug.trim()
    if (!trimmedName) {
      toast.error('Workspace name is required')
      return
    }
    if (!trimmedSlug) {
      toast.error('Workspace slug is required')
      return
    }

    try {
      await createWorkspace({
        name: trimmedName,
        slug: trimmedSlug,
        description: description.trim(),
      })
      toast.success('Workspace created')
    } catch (err) {
      toast.error(
        err instanceof Error ? err.message : 'Failed to create workspace'
      )
    }
  }

  return (
    <div className='flex h-full items-center justify-center overflow-y-auto p-4'>
      <Card className='w-full max-w-xl'>
        <CardHeader>
          <CardTitle>Create your first workspace</CardTitle>
          <p className='text-sm text-muted-foreground'>
            Workspaces scope agents, channels, cron jobs, model providers, and
            API tokens. Create one to enter the dashboard.
          </p>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleSubmit} className='space-y-4'>
            <div className='space-y-2'>
              <label className='text-sm font-medium' htmlFor='workspace-name'>
                Name
              </label>
              <Input
                id='workspace-name'
                value={name}
                onChange={(e) => {
                  const next = e.target.value
                  setName(next)
                  setSlug(
                    next
                      .toLowerCase()
                      .trim()
                      .replace(/[^a-z0-9]+/g, '-')
                      .replace(/^-+|-+$/g, '')
                  )
                }}
                placeholder='Default'
                disabled={isCreating}
              />
            </div>
            <div className='space-y-2'>
              <label className='text-sm font-medium' htmlFor='workspace-slug'>
                Slug
              </label>
              <Input
                id='workspace-slug'
                value={slug}
                onChange={(e) =>
                  setSlug(
                    e.target.value
                      .toLowerCase()
                      .trim()
                      .replace(/[^a-z0-9-]+/g, '-')
                  )
                }
                placeholder='default'
                disabled={isCreating}
              />
            </div>
            <div className='space-y-2'>
              <label
                className='text-sm font-medium'
                htmlFor='workspace-description'
              >
                Description
              </label>
              <Textarea
                id='workspace-description'
                value={description}
                onChange={(e) => setDescription(e.target.value)}
                placeholder='Optional description'
                disabled={isCreating}
              />
            </div>
            <Button type='submit' disabled={isCreating}>
              {isCreating ? 'Creating...' : 'Create workspace'}
            </Button>
          </form>
        </CardContent>
      </Card>
    </div>
  )
}
