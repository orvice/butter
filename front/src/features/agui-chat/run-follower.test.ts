import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { ApiError } from '@/api/client'
import type { UISnapshot } from './a2ui/protocol'
import type { ThreadHistory, ThreadRead, ThreadReads } from './history'
import {
  Backoff,
  POLL_DELAYS_MS,
  RunFollower,
  awaitRunEnd,
  type RunEnd,
} from './run-follower'

const RUN = { runId: 'run-2', invocationId: 'inv-2' }

function history(running?: typeof RUN): ThreadHistory {
  return {
    threadId: 't',
    messages: [{ id: 'u1', role: 'user', content: 'plan the trip' }],
    interrupts: [],
    surfaces: [],
    ...(running && { running }),
  }
}

function snapshot(running?: typeof RUN): UISnapshot {
  return {
    version: 'v0.9.1',
    catalogId: 'butter-basic-v1',
    threadId: 't',
    surfaces: [],
    ...(running && { running }),
  }
}

type Answer<T> = T | Error

// scripted reads the thread from queues of answers: each read takes the
// next one, an Error rejecting it, and the thread with no run once a queue
// runs out. reads logs each read with the time it went out.
function scripted(
  histories: Array<Answer<ThreadHistory>>,
  snapshots: Array<Answer<UISnapshot>> = []
) {
  const log: string[] = []
  const answer = async <T>(queue: Array<Answer<T>>, last: T) => {
    const next = queue.length > 0 ? queue.shift()! : last
    if (next instanceof Error) throw next
    return next
  }
  const reads: ThreadReads = {
    history: (signal) => {
      log.push(`history@${Date.now()}${signal?.aborted ? ' (aborted)' : ''}`)
      return answer(histories, history())
    },
    snapshot: (signal) => {
      log.push(`snapshot@${Date.now()}${signal?.aborted ? ' (aborted)' : ''}`)
      return answer(snapshots, snapshot())
    },
  }
  return { reads, log }
}

// abortable is a read that, like fetch, rejects once its signal aborts.
function abortable(signal?: AbortSignal): Promise<never> {
  return new Promise((_, reject) =>
    signal?.addEventListener('abort', () =>
      reject(new DOMException('aborted', 'AbortError'))
    )
  )
}

beforeEach(() => {
  vi.useFakeTimers()
  vi.setSystemTime(0)
})

afterEach(() => {
  vi.useRealTimers()
})

describe('Backoff', () => {
  // waits records how long each of n waits took.
  async function waits(backoff: Backoff, n: number): Promise<number[]> {
    const signal = new AbortController().signal
    const took: number[] = []
    for (let i = 0; i < n; i++) {
      const start = Date.now()
      const waited = backoff.wait(signal)
      await vi.advanceTimersToNextTimerAsync()
      await waited
      took.push(Date.now() - start)
    }
    return took
  }

  it('waits a second, then twice as long each time, up to five seconds', async () => {
    expect(POLL_DELAYS_MS).toEqual([1_000, 2_000, 4_000, 5_000])
    expect(await waits(new Backoff(), 6)).toEqual([
      1_000, 2_000, 4_000, 5_000, 5_000, 5_000,
    ])
  })

  it('ends a wait at once when hurried, and starts the delays over', async () => {
    const backoff = new Backoff()
    const signal = new AbortController().signal
    await waits(backoff, 3)
    const start = Date.now()
    const waited = backoff.wait(signal)
    await vi.advanceTimersByTimeAsync(100)
    backoff.hurry()
    await waited
    expect(Date.now() - start).toBe(100)
    expect(await waits(backoff, 2)).toEqual([1_000, 2_000])
  })

  it('skips the next wait when hurried between two', async () => {
    const backoff = new Backoff()
    await waits(backoff, 2)
    backoff.hurry()
    const start = Date.now()
    await backoff.wait(new AbortController().signal)
    expect(Date.now()).toBe(start)
    expect(await waits(backoff, 1)).toEqual([1_000])
  })

  it('ends a wait at once when the signal aborts', async () => {
    const aborted = new AbortController()
    const waited = new Backoff().wait(aborted.signal)
    await vi.advanceTimersByTimeAsync(300)
    aborted.abort()
    await waited
    expect(Date.now()).toBe(300)
    // A wait on an aborted signal does not wait at all.
    await new Backoff().wait(aborted.signal)
    expect(Date.now()).toBe(300)
  })
})

describe('awaitRunEnd', () => {
  it('reads the history again, with backoff, until no run holds the thread, then the snapshot', async () => {
    const { reads, log } = scripted([history(RUN), history(RUN), history()])
    const ended = awaitRunEnd(reads, new AbortController().signal)
    await vi.advanceTimersByTimeAsync(1_000 + 2_000 + 4_000)
    await expect(ended).resolves.toEqual({
      history: history(),
      snapshot: snapshot(),
    })
    expect(log).toEqual([
      'history@1000',
      'history@3000',
      'history@7000',
      'snapshot@7000',
    ])
  })

  it('waits out a run the snapshot names: one started after the history was read', async () => {
    const { reads, log } = scripted([history(), history()], [snapshot(RUN)])
    const ended = awaitRunEnd(reads, new AbortController().signal)
    await vi.advanceTimersByTimeAsync(1_000 + 2_000)
    await expect(ended).resolves.toEqual({
      history: history(),
      snapshot: snapshot(),
    })
    expect(log).toEqual([
      'history@1000',
      'snapshot@1000',
      'history@3000',
      'snapshot@3000',
    ])
  })

  it('reads again after a failure for the moment: 502, 503, 504, or no answer at all', async () => {
    const { reads, log } = scripted(
      [
        new ApiError('503', 'run state unavailable, retry later'),
        new ApiError('502', 'bad gateway'),
        new TypeError('Failed to fetch'),
        history(),
      ],
      [new ApiError('504', 'gateway timeout')]
    )
    const ended = awaitRunEnd(reads, new AbortController().signal)
    await vi.advanceTimersByTimeAsync(1_000 + 2_000 + 4_000 + 5_000 + 5_000)
    await expect(ended).resolves.toEqual({
      history: history(),
      snapshot: snapshot(),
    })
    expect(log).toEqual([
      'history@1000',
      'history@3000',
      'history@7000',
      'history@12000',
      'snapshot@12000',
      'history@17000',
      'snapshot@17000',
    ])
  })

  for (const code of ['409', '500']) {
    it(`reads nothing again after a ${code}`, async () => {
      const refused = new ApiError(code, 'refused')
      const { reads, log } = scripted([history(RUN), refused])
      const ended = awaitRunEnd(reads, new AbortController().signal)
      const failed = expect(ended).rejects.toBe(refused)
      await vi.advanceTimersByTimeAsync(1_000 + 2_000)
      await failed
      await vi.advanceTimersByTimeAsync(60_000)
      expect(log).toEqual(['history@1000', 'history@3000'])
    })
  }

  it('reads nothing more once the signal aborts during a wait', async () => {
    const { reads, log } = scripted([history(RUN)])
    const aborted = new AbortController()
    const ended = awaitRunEnd(reads, aborted.signal)
    await vi.advanceTimersByTimeAsync(1_500)
    aborted.abort()
    await expect(ended).resolves.toBeNull()
    await vi.advanceTimersByTimeAsync(60_000)
    expect(log).toEqual(['history@1000'])
  })

  it('aborts the read under way when the signal aborts', async () => {
    const log: string[] = []
    const reads: ThreadReads = {
      history: (signal) => {
        log.push('history')
        return abortable(signal)
      },
      snapshot: (signal) => {
        log.push('snapshot')
        return abortable(signal)
      },
    }
    const aborted = new AbortController()
    const ended = awaitRunEnd(reads, aborted.signal)
    await vi.advanceTimersByTimeAsync(1_000)
    aborted.abort()
    await expect(ended).resolves.toBeNull()
    await vi.advanceTimersByTimeAsync(60_000)
    expect(log).toEqual(['history'])
  })

  it('drops a read that comes back after the signal aborted', async () => {
    const aborted = new AbortController()
    const log: string[] = []
    const reads: ThreadReads = {
      history: async () => {
        log.push('history')
        // The page stops following while the read is answered.
        aborted.abort()
        return history()
      },
      snapshot: async () => {
        log.push('snapshot')
        return snapshot()
      },
    }
    const ended = awaitRunEnd(reads, aborted.signal)
    await vi.advanceTimersByTimeAsync(1_000)
    await expect(ended).resolves.toBeNull()
    expect(log).toEqual(['history'])
  })
})

describe('RunFollower', () => {
  // follower follows a run through reads, recording what it shows, what it
  // reports as failed, and each change of isPolling.
  function follower(reads: ThreadReads) {
    const shown: ThreadRead[] = []
    const failures: unknown[] = []
    const polling: boolean[] = []
    const f = new RunFollower(reads)
    const end: RunEnd = {
      show: (read) => shown.push(read),
      failed: (err) => failures.push(err),
    }
    f.endWith(end)
    f.subscribe(() => polling.push(f.isPolling()))
    return { f, shown, failures, polling }
  }

  it('polls until the run ended, then shows the thread as the read after found it', async () => {
    const { reads } = scripted([history(RUN), history()])
    const { f, shown, failures, polling } = follower(reads)
    const followed = f.follow(new AbortController().signal)
    expect(f.isPolling()).toBe(true)
    await vi.advanceTimersByTimeAsync(1_000)
    expect(shown).toEqual([])
    await vi.advanceTimersByTimeAsync(2_000)
    await followed
    expect(shown).toEqual([{ history: history(), snapshot: snapshot() }])
    expect(failures).toEqual([])
    expect(polling).toEqual([true, false])
  })

  it('waits a run out when called on its own, as live re-attach falls back to it', async () => {
    const { reads, log } = scripted([history(RUN), history()])
    const { f, shown, polling } = follower(reads)
    const polled = f.pollUntilEnded(new AbortController().signal)
    expect(f.isPolling()).toBe(true)
    await vi.advanceTimersByTimeAsync(1_000 + 2_000)
    await polled
    expect(shown).toEqual([{ history: history(), snapshot: snapshot() }])
    expect(polling).toEqual([true, false])
    expect(log).toEqual(['history@1000', 'history@3000', 'snapshot@3000'])
  })

  it('hands the history of the read that ends the wait to the last run, through the reading started with it', async () => {
    const stopped = {
      ...history(),
      lastRun: { status: 'cancelled', error: 'stopped by user' },
    }
    const { reads } = scripted([history(RUN), stopped])
    const shown: ThreadRead[] = []
    const started: number[] = []
    const taken: Array<[number, ThreadHistory]> = []
    const f = new RunFollower(reads)
    f.endWith({
      show: (read) => shown.push(read),
      failed: () => {},
      reading: () => {
        const read = started.push(Date.now())
        return (h) => taken.push([read, h])
      },
    })
    const polled = f.pollUntilEnded(new AbortController().signal)
    await vi.advanceTimersByTimeAsync(1_000 + 2_000)
    await polled
    // Each read of the history started a reading; only the last one's
    // taker got a history, the one that ended the wait.
    expect(started).toEqual([1_000, 3_000])
    expect(taken).toEqual([[2, stopped]])
    expect(shown).toEqual([{ history: stopped, snapshot: snapshot() }])
  })

  it('reports a read that failed for good, and stops polling', async () => {
    const refused = new ApiError('409', 'a run is in progress on this thread')
    const { reads, log } = scripted([refused])
    const { f, shown, failures, polling } = follower(reads)
    const followed = f.follow(new AbortController().signal)
    await vi.advanceTimersByTimeAsync(1_000)
    await followed
    expect(failures).toEqual([refused])
    expect(shown).toEqual([])
    expect(polling).toEqual([true, false])
    await vi.advanceTimersByTimeAsync(60_000)
    expect(log).toEqual(['history@1000'])
  })

  it('shows nothing, and reports nothing, once the page stopped following the run', async () => {
    const { reads } = scripted([history(RUN), history()])
    const { f, shown, failures, polling } = follower(reads)
    const aborted = new AbortController()
    const followed = f.follow(aborted.signal)
    await vi.advanceTimersByTimeAsync(1_000)
    aborted.abort()
    await followed
    expect(shown).toEqual([])
    expect(failures).toEqual([])
    expect(polling).toEqual([true, false])
  })

  it('reads the thread again at once when hurried, and resolves once the end shows', async () => {
    const { reads, log } = scripted([history(RUN), history(RUN), history()])
    const { f, shown } = follower(reads)
    const followed = f.follow(new AbortController().signal)
    await vi.advanceTimersByTimeAsync(1_000 + 500)
    // A Stop reached the run halfway through the second wait.
    let ended = false
    const ending = f.untilEnded()
    expect(ending).toBeDefined()
    void ending!.then(() => {
      ended = true
      // The end shows before the wait for it is over.
      expect(shown).toHaveLength(1)
    })
    await vi.advanceTimersByTimeAsync(0)
    expect(log).toEqual(['history@1000', 'history@1500'])
    expect(ended).toBe(false)
    // The run took a moment to stop; the delays started over.
    await vi.advanceTimersByTimeAsync(1_000)
    await followed
    expect(log).toEqual([
      'history@1000',
      'history@1500',
      'history@2500',
      'snapshot@2500',
    ])
    expect(ended).toBe(true)
  })

  it('has no end to wait for while it polls for no run', async () => {
    const { reads } = scripted([history(RUN)])
    const { f } = follower(reads)
    expect(f.untilEnded()).toBeUndefined()
    const followed = f.follow(new AbortController().signal)
    expect(f.untilEnded()).toBeDefined()
    await vi.advanceTimersByTimeAsync(3_000)
    await followed
    expect(f.untilEnded()).toBeUndefined()
  })
})
