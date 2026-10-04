// The sidebar groups conversations by when they were last updated, in the
// viewer's local time: today, the 7 days before, and older.

export type RecencyGroupKey = 'today' | 'week' | 'older'

const GROUP_KEYS: readonly RecencyGroupKey[] = ['today', 'week', 'older']

const GROUP_TITLES: Record<RecencyGroupKey, string> = {
  today: 'Today',
  week: 'Previous 7 days',
  older: 'Older',
}

const WEEK_MS = 7 * 24 * 60 * 60 * 1000

// recencyGroup places an update time: since the start of today, within the
// 7 days before now, or older. A conversation without a readable time counts
// as older.
export function recencyGroup(
  time: string | undefined,
  now: Date
): RecencyGroupKey {
  if (!time) return 'older'
  const updated = new Date(time)
  const startOfToday = new Date(
    now.getFullYear(),
    now.getMonth(),
    now.getDate()
  )
  if (updated >= startOfToday) return 'today'
  if (now.getTime() - updated.getTime() < WEEK_MS) return 'week'
  return 'older'
}

export interface RecencyGroup<T> {
  key: RecencyGroupKey
  title: string
  items: T[]
}

// groupByRecency splits items into those groups, newest group first. Each
// group keeps its items in the order given; an empty group is left out.
export function groupByRecency<T>(
  items: readonly T[],
  timeOf: (item: T) => string | undefined,
  now: Date
): RecencyGroup<T>[] {
  return GROUP_KEYS.map((key) => ({
    key,
    title: GROUP_TITLES[key],
    items: items.filter((item) => recencyGroup(timeOf(item), now) === key),
  })).filter((group) => group.items.length > 0)
}
