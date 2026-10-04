import { describe, expect, it } from 'vitest'
import { groupByRecency, recencyGroup } from './recency-groups'

// The sidebar reads times in the viewer's local time zone.
const now = new Date(2026, 9, 5, 9, 30)
const local = (day: number, hours = 0, minutes = 0) =>
  new Date(2026, 9, day, hours, minutes).toISOString()
// msAgo is exact whatever the time zone does in between (daylight saving).
const msAgo = (ms: number) => new Date(now.getTime() - ms).toISOString()
const WEEK_MS = 7 * 24 * 60 * 60 * 1000

describe('recencyGroup', () => {
  it('is today from the local start of the day on', () => {
    expect(recencyGroup(local(5, 0, 0), now)).toBe('today')
    expect(recencyGroup(local(5, 9, 29), now)).toBe('today')
    // A time ahead of the viewer's clock is still today.
    expect(recencyGroup(local(5, 23, 0), now)).toBe('today')
  })

  it('is the previous 7 days before today, counted back from now', () => {
    expect(recencyGroup(local(4, 23, 59), now)).toBe('week')
    expect(recencyGroup(msAgo(WEEK_MS - 60_000), now)).toBe('week')
    expect(recencyGroup(msAgo(WEEK_MS), now)).toBe('older')
  })

  it('is older without a readable time', () => {
    expect(recencyGroup(undefined, now)).toBe('older')
    expect(recencyGroup('', now)).toBe('older')
    expect(recencyGroup('not a time', now)).toBe('older')
  })
})

describe('groupByRecency', () => {
  const item = (id: string, time?: string) => ({ id, time })

  it('groups newest first, keeps the given order, and leaves out empty groups', () => {
    const items = [
      item('a', local(5, 9, 0)),
      item('b', msAgo(30 * 24 * 60 * 60 * 1000)),
      item('c', local(3, 12, 0)),
      item('d', local(5, 8, 0)),
      item('e'),
    ]
    const groups = groupByRecency(items, (i) => i.time, now)
    expect(
      groups.map((g) => [g.key, g.title, g.items.map((i) => i.id)])
    ).toEqual([
      ['today', 'Today', ['a', 'd']],
      ['week', 'Previous 7 days', ['c']],
      ['older', 'Older', ['b', 'e']],
    ])

    const todayOnly = groupByRecency(
      [item('x', local(5, 1))],
      (i) => i.time,
      now
    )
    expect(todayOnly.map((g) => g.key)).toEqual(['today'])
    expect(groupByRecency([], (i: { time?: string }) => i.time, now)).toEqual(
      []
    )
  })
})
