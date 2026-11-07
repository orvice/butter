import { useState } from 'react'
import { useInfiniteQuery } from '@tanstack/react-query'
import type { WorkStationClient } from '@/api/work-station'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'

// Options come exclusively from the workspace's visible local projection.
// Searching or paging this selector never calls a provider discovery API.
export function FacetFilter({
  client,
  prefix,
  dimension,
  label,
  value,
  onChange,
  filters,
  defaults = [],
}: {
  client: WorkStationClient
  prefix: string[]
  dimension: 'provider' | 'scope' | 'project' | 'status' | 'assignee'
  label: string
  value: string
  onChange: (value: string) => void
  filters: { kind: string; connection: string; scope: string; provider: string }
  defaults?: { id: string; name: string }[]
}) {
  const [search, setSearch] = useState('')
  const context = {
    kind: filters.kind,
    connectionId: filters.connection,
    scopeId: dimension === 'scope' ? '' : filters.scope,
    provider: dimension === 'provider' ? '' : filters.provider,
  }
  const query = useInfiniteQuery({
    queryKey: [...prefix, 'facets', dimension, context, search],
    initialPageParam: '',
    refetchInterval: 15_000,
    queryFn: ({ signal, pageParam }) =>
      client.listFilterOptions(
        { ...context, dimension, search, pageSize: 50, pageToken: pageParam },
        { signal }
      ),
    getNextPageParam: (last) => last.nextPageToken || undefined,
  })
  const options = Array.from(
    new Map(
      [...defaults, ...(query.data?.pages.flatMap((p) => p.options) ?? [])].map(
        (option) => [option.id, option]
      )
    ).values()
  )
  return (
    <div className='space-y-1 text-sm'>
      <label className='block space-y-1'>
        <span>{label}</span>
        <select
          aria-label={label}
          className='block min-h-10 w-full rounded-md border bg-background px-3 text-base sm:text-sm'
          value={value}
          onChange={(e) => onChange(e.target.value)}
        >
          {defaults.length === 0 && (
            <option value=''>All {label.toLowerCase()}s</option>
          )}
          {value && !options.some((option) => option.id === value) && (
            <option value={value}>{value}</option>
          )}
          {options.map((option) => (
            <option key={option.id} value={option.id}>
              {option.name}
            </option>
          ))}
        </select>
      </label>
      <details>
        <summary className='min-h-6 cursor-pointer text-muted-foreground'>
          Find more {label.toLowerCase()} options
        </summary>
        <div className='mt-2 space-y-2'>
          <label className='block space-y-1'>
            <span>Search {label.toLowerCase()} options</span>
            <Input
              value={search}
              maxLength={256}
              onChange={(e) => setSearch(e.target.value)}
            />
          </label>
          {query.isPending && <p role='status'>Loading options…</p>}
          {query.error && (
            <p role='alert'>Unable to load options. Retry the search.</p>
          )}
          {!query.isPending &&
            !query.error &&
            options.length === defaults.length && (
              <p className='text-muted-foreground'>
                No matching options in synced work.
              </p>
            )}
          {query.hasNextPage && (
            <Button
              type='button'
              variant='outline'
              size='sm'
              disabled={query.isFetchingNextPage}
              onClick={() => void query.fetchNextPage()}
            >
              Load more {label.toLowerCase()} options
            </Button>
          )}
        </div>
      </details>
    </div>
  )
}
