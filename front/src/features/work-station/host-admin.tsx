import { useState, type FormEvent } from 'react'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import type { WorkStationHost } from '@/gen/agents/v1/workstation_pb'
import { ConnectError } from '@connectrpc/connect'
import type { WorkStationHostClient } from '@/api/work-station'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
} from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'

// Global endpoint administration is kept outside workspace connection settings.
// The server independently requires a global administrator on every mutation.
export function PlaneHostAdmin({
  client,
  prefix,
}: {
  client: WorkStationHostClient
  prefix: string[]
}) {
  const [open, setOpen] = useState(false)
  const [id, setId] = useState('')
  const qc = useQueryClient()
  const hosts = useQuery({
    queryKey: [...prefix, 'plane-hosts'],
    queryFn: ({ signal }) => client.listHosts({}, { signal }),
    enabled: open,
  })
  return (
    <div className='border-t pt-6'>
      <Button variant='outline' onClick={() => setOpen(true)}>
        Manage global Plane hosts
      </Button>
      <Dialog open={open} onOpenChange={setOpen}>
        <DialogContent className='max-h-[85vh] overflow-y-auto sm:max-w-2xl'>
          <DialogHeader>
            <DialogTitle>Global Plane hosts</DialogTitle>
            <DialogDescription>
              Approved endpoints are available to every workspace. Only global
              administrators can change them. Workspace credentials are managed
              separately.
            </DialogDescription>
          </DialogHeader>
          <label className='block space-y-1 text-sm'>
            <span>Host to manage</span>
            <select
              className='block min-h-10 w-full rounded-md border bg-background px-3'
              value={id}
              onChange={(e) => setId(e.target.value)}
            >
              <option value=''>Register a host</option>
              {hosts.data?.hosts
                .filter((host) => host.id !== 'plane-cloud')
                .map((host) => (
                  <option key={host.id} value={host.id}>
                    {host.name}
                  </option>
                ))}
            </select>
          </label>
          <p className='text-sm text-muted-foreground'>
            Plane Cloud is built in. Register a self-managed host only after
            reviewing its public HTTPS endpoint and v2 profile. Connected hosts
            must keep their endpoints until their connections are removed.
          </p>
          {hosts.error && (
            <p role='alert' className='text-sm text-destructive'>
              Unable to load Plane hosts.
            </p>
          )}
          <HostForm
            key={id}
            client={client}
            host={hosts.data?.hosts.find((host) => host.id === id)}
            changed={async (nextId) => {
              await qc.invalidateQueries({
                queryKey: [...prefix, 'plane-hosts'],
              })
              setId(nextId)
            }}
          />
        </DialogContent>
      </Dialog>
    </div>
  )
}

function HostForm({
  client,
  host,
  changed,
}: {
  client: WorkStationHostClient
  host?: WorkStationHost
  changed: (id: string) => Promise<void>
}) {
  const [name, setName] = useState(host?.name ?? '')
  const [api, setApi] = useState(host?.apiBaseUrl ?? '')
  const [web, setWeb] = useState(host?.webBaseUrl ?? '')
  const [error, setError] = useState('')
  const [pending, setPending] = useState(false)
  const [confirm, setConfirm] = useState(false)
  async function run(action: () => Promise<void>) {
    setPending(true)
    setError('')
    try {
      await action()
    } catch (e) {
      setError(
        e instanceof ConnectError
          ? e.rawMessage
          : 'Unable to save this host. Check the endpoint and try again.'
      )
    } finally {
      setPending(false)
    }
  }
  function save(e: FormEvent) {
    e.preventDefault()
    void run(async () => {
      const value = {
        id: host?.id ?? '',
        name,
        apiBaseUrl: api,
        webBaseUrl: web,
        profile: 'plane-v2',
      }
      const result = host
        ? await client.updateHost({
            host: value,
            expectedRevision: host.revision,
          })
        : await client.createHost({ host: value })
      await changed(result.host?.id ?? '')
    })
  }
  return (
    <div className='space-y-4'>
      <form onSubmit={save} className='space-y-4'>
        <label className='block space-y-1 text-sm'>
          <span>Host name</span>
          <Input
            value={name}
            maxLength={128}
            required
            onChange={(e) => setName(e.target.value)}
          />
        </label>
        <label className='block space-y-1 text-sm'>
          <span>API base URL</span>
          <Input
            type='url'
            value={api}
            maxLength={2048}
            required
            placeholder='https://plane.example.com'
            onChange={(e) => setApi(e.target.value)}
          />
        </label>
        <label className='block space-y-1 text-sm'>
          <span>Web base URL</span>
          <Input
            type='url'
            value={web}
            maxLength={2048}
            required
            placeholder='https://plane.example.com'
            onChange={(e) => setWeb(e.target.value)}
          />
        </label>
        <p className='text-sm text-muted-foreground'>
          Profile: Plane v2. Private, loopback, link-local, and redirected
          destinations are refused by the outbound policy.
        </p>
        <Button type='submit' disabled={pending}>
          {host ? 'Save host' : 'Register host'}
        </Button>
      </form>
      {host &&
        (!confirm ? (
          <Button
            variant='destructive'
            disabled={pending}
            onClick={() => setConfirm(true)}
          >
            Delete host
          </Button>
        ) : (
          <div className='space-y-3 rounded-md border p-4'>
            <p className='text-sm'>
              Delete {host.name} from every workspace's approved host list?
              Connected hosts must be disconnected first.
            </p>
            <div className='flex flex-wrap gap-3'>
              <Button
                variant='destructive'
                disabled={pending}
                onClick={() =>
                  void run(async () => {
                    await client.deleteHost({
                      id: host.id,
                      expectedRevision: host.revision,
                    })
                    await changed('')
                  })
                }
              >
                Confirm host deletion
              </Button>
              <Button variant='outline' onClick={() => setConfirm(false)}>
                Cancel
              </Button>
            </div>
          </div>
        ))}
      {error && (
        <p role='alert' className='text-sm text-destructive'>
          {error}
        </p>
      )}
    </div>
  )
}
