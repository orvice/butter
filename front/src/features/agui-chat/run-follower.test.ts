import type { ChatModelRunResult, ThreadMessage } from '@assistant-ui/react'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import type { AGUIEvent } from '@/api/agui'
import { ApiError } from '@/api/client'
import type { UISnapshot } from './a2ui/protocol'
import type { ThreadHistory, ThreadRead, ThreadReads } from './history'
import type { RunEffects, RunEndEvent } from './run-fold'
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

// drain consumes follow to its end, as the runtime consumes resume(), and
// returns the replies it yielded.
async function drain(
  replies: AsyncGenerator<ChatModelRunResult, void, undefined>
): Promise<ChatModelRunResult[]> {
  const out: ChatModelRunResult[] = []
  for await (const reply of replies) out.push(reply)
  return out
}

// live is a run's log as attaching streams it: the events the test sends,
// as they come, until it ends the stream or breaks it off. The attach rejects
// with error when one is given, and answers null (204) when there is no log.
// Like fetch, the stream ends with an AbortError once its signal aborts.
function live(error?: unknown) {
  const queue: AGUIEvent[] = []
  let wake: (() => void) | null = null
  let state: 'open' | 'ended' | 'broken' = 'open'
  const record = { attaches: 0, closed: false }
  const attach: NonNullable<ThreadReads['attach']> = async (signal) => {
    record.attaches++
    if (error) throw error
    async function* events(): AsyncGenerator<AGUIEvent, void, undefined> {
      try {
        for (;;) {
          while (queue.length > 0) yield queue.shift()!
          if (state === 'ended') return
          if (state === 'broken') throw new TypeError('network error')
          await new Promise<void>((resolve, reject) => {
            wake = resolve
            signal?.addEventListener(
              'abort',
              () => reject(new DOMException('aborted', 'AbortError')),
              { once: true }
            )
          })
        }
      } finally {
        record.closed = true
      }
    }
    return events()
  }
  const push = (events: AGUIEvent[]) => {
    queue.push(...events)
    wake?.()
  }
  return {
    attach,
    record,
    send: (...events: AGUIEvent[]) => push(events),
    end: (...events: AGUIEvent[]) => {
      state = 'ended'
      push(events)
    },
    breakOff: () => {
      state = 'broken'
      push([])
    },
  }
}

// The log of the run the reads find: RUN_STARTED, then the state a detached
// run always opens with.
const runStarted: AGUIEvent[] = [
  { type: 'RUN_STARTED', threadId: 't', runId: RUN.runId },
  { type: 'STATE_SNAPSHOT', snapshot: { plan: 'draft' } },
]
const textStart: AGUIEvent = {
  type: 'TEXT_MESSAGE_START',
  messageId: 'm2',
  role: 'assistant',
}
const delta = (text: string): AGUIEvent => ({
  type: 'TEXT_MESSAGE_CONTENT',
  messageId: 'm2',
  delta: text,
})
const runFinished: AGUIEvent = {
  type: 'RUN_FINISHED',
  threadId: 't',
  runId: RUN.runId,
  outcome: { type: 'success' },
}
const fallback = (reason: string): AGUIEvent => ({
  type: 'CUSTOM',
  name: 'butter.fallback',
  value: { threadId: 't', runId: RUN.runId, reason },
})

const lastOf = <T>(items: readonly T[]): T | undefined =>
  items[items.length - 1]

const textOf = (reply: ChatModelRunResult | undefined) =>
  (reply?.content ?? []).flatMap((part) =>
    part.type === 'text' ? [part.text] : []
  )

describe('RunFollower', () => {
  // follower follows RUN, which the thread's history found, through reads.
  // It records what it shows, what it reports as failed, how a run streamed
  // from its log ended, how often the page stopped following it before its
  // end, and each change of isPolling.
  function follower(reads: ThreadReads, effects?: RunEffects) {
    const shown: ThreadRead[] = []
    const failures: unknown[] = []
    const ended: RunEndEvent[] = []
    const detached: number[] = []
    const polling: boolean[] = []
    const f = new RunFollower(reads)
    f.found(RUN)
    if (effects) f.setEffects(effects)
    const end: RunEnd = {
      show: (read) => shown.push(read),
      failed: (err) => failures.push(err),
      ended: (event) => ended.push(event),
      detached: () => detached.push(Date.now()),
    }
    f.endWith(end)
    f.subscribe(() => polling.push(f.isPolling()))
    return { f, shown, failures, ended, detached, polling }
  }

  const follow = (f: RunFollower, signal = new AbortController().signal) =>
    drain(f.follow({ abortSignal: signal }))

  describe('from the run’s log', () => {
    it('streams the reply as the log replays and follows the run, and reads nothing', async () => {
      const { reads, log } = scripted([])
      const run = live()
      const { f, shown, ended, detached, polling } = follower({
        ...reads,
        attach: run.attach,
      })
      const replies: ChatModelRunResult[] = []
      const followed = (async () => {
        for await (const reply of f.follow({
          abortSignal: new AbortController().signal,
        })) {
          replies.push(reply)
        }
      })()
      // The replay: what the run sent before the page attached.
      run.send(...runStarted, textStart, delta('Three days'))
      await vi.advanceTimersByTimeAsync(0)
      expect(textOf(lastOf(replies))).toEqual(['Three days'])
      expect(lastOf(replies)?.status).toEqual({ type: 'running' })
      // Then the run goes on.
      run.send(delta(' in Lisbon.'))
      await vi.advanceTimersByTimeAsync(0)
      expect(textOf(lastOf(replies))).toEqual(['Three days in Lisbon.'])
      expect(f.isPolling()).toBe(false)
      expect(f.untilEnded()).toBeUndefined()
      expect(ended).toEqual([])

      run.end({ type: 'TEXT_MESSAGE_END', messageId: 'm2' }, runFinished)
      await followed
      expect(lastOf(replies)).toEqual({
        content: [{ type: 'text', text: 'Three days in Lisbon.' }],
        status: { type: 'complete', reason: 'unknown' },
      })
      expect(ended).toEqual([{ type: 'RUN_FINISHED' }])
      // The run ended in the stream: nothing is read, and nothing waited.
      await vi.advanceTimersByTimeAsync(60_000)
      expect(log).toEqual([])
      expect(shown).toEqual([])
      expect(polling).toEqual([])
      expect(detached).toEqual([])
      expect(run.record.closed).toBe(true)
    })

    it('hands butter.a2ui to the A2UI store and the state events to the shared state', async () => {
      const run = live()
      const applied: unknown[] = []
      let state: unknown = { stale: true }
      const { f } = follower(
        { ...scripted([]).reads, attach: run.attach },
        {
          a2ui: (value) => applied.push(value),
          state: (update) => (state = update(state)),
        }
      )
      const card = {
        version: 'v0.9.1',
        surfaceId: 'card-1',
        kind: 'card',
        revision: 1,
        seq: 0,
        envelope: { version: 'v0.9.1', createSurface: { surfaceId: 'card-1' } },
      }
      run.end(
        ...runStarted,
        { type: 'CUSTOM', name: 'butter.a2ui', value: card },
        {
          type: 'STATE_DELTA',
          delta: [{ op: 'replace', path: '/plan', value: 'final' }],
        },
        runFinished
      )
      const replies = await follow(f)
      expect(applied).toEqual([card])
      expect(state).toEqual({ plan: 'final' })
      expect(lastOf(replies)?.content).toEqual([
        { type: 'data', name: 'butter.a2ui', data: card },
      ])
    })

    it('tells how a run that failed or was stopped ended, once its reply shows it', async () => {
      const run = live()
      const { f, ended, shown } = follower({
        ...scripted([]).reads,
        attach: run.attach,
      })
      const stopped: AGUIEvent = {
        type: 'RUN_ERROR',
        code: 'stopped',
        message: 'stopped by user',
        runId: RUN.runId,
      }
      run.end(...runStarted, textStart, delta('Day one: Alfama.'), stopped)
      const replies = await follow(f)
      expect(lastOf(replies)).toEqual({
        content: [{ type: 'text', text: 'Day one: Alfama.' }],
        status: {
          type: 'incomplete',
          reason: 'error',
          error: 'stopped by user',
        },
      })
      expect(ended).toEqual([
        { type: 'RUN_ERROR', code: 'stopped', message: 'stopped by user' },
      ])
      expect(shown).toEqual([])
    })

    it('leaves out a result for a call of a reply before the run’s', async () => {
      const run = live()
      const { f } = follower({ ...scripted([]).reads, attach: run.attach })
      run.end(
        ...runStarted,
        {
          type: 'TOOL_CALL_RESULT',
          messageId: 'm2',
          toolCallId: 'call-0',
          content: '{"approved":true}',
          role: 'tool',
        },
        runFinished
      )
      const earlier = [
        {
          id: 'a1',
          role: 'assistant',
          content: [
            {
              type: 'tool-call',
              toolCallId: 'call-0',
              toolName: 'confirm',
              args: {},
              argsText: '{}',
            },
          ],
        },
      ] as unknown as ThreadMessage[]
      const replies = await drain(
        f.follow({
          abortSignal: new AbortController().signal,
          messages: earlier,
        })
      )
      expect(lastOf(replies)?.content).toEqual([])
    })

    it('stops reading the log once the page stopped following the run, and tells only that', async () => {
      const { reads, log } = scripted([])
      const run = live()
      const { f, ended, detached, shown, polling } = follower({
        ...reads,
        attach: run.attach,
      })
      const aborted = new AbortController()
      const followed = follow(f, aborted.signal)
      run.send(...runStarted, textStart, delta('Three days'))
      await vi.advanceTimersByTimeAsync(0)
      expect(detached).toEqual([])
      aborted.abort()
      const replies = await followed
      expect(textOf(lastOf(replies))).toEqual(['Three days'])
      expect(run.record.closed).toBe(true)
      expect(detached).toHaveLength(1)
      // The run's end, if it comes, reaches no one.
      run.end(runFinished)
      await vi.advanceTimersByTimeAsync(60_000)
      expect(ended).toEqual([])
      expect(shown).toEqual([])
      expect(polling).toEqual([])
      expect(log).toEqual([])
      expect(detached).toHaveLength(1)
    })

    it('streams only the run the history found, and waits out one that took the thread after it', async () => {
      const { reads, log } = scripted([history()])
      const run = live()
      const { f, shown, ended, detached } = follower({
        ...reads,
        attach: run.attach,
      })
      const followed = follow(f)
      // The run the reads found ended, and another took the thread before
      // the page attached: the log is that one's.
      run.send(
        { type: 'RUN_STARTED', threadId: 't', runId: 'run-3' },
        { type: 'STATE_SNAPSHOT', snapshot: {} },
        textStart,
        delta('Another reply')
      )
      await vi.advanceTimersByTimeAsync(1_000)
      expect(await followed).toEqual([])
      expect(run.record.closed).toBe(true)
      expect(log).toEqual(['history@1000', 'snapshot@1000'])
      expect(shown).toEqual([{ history: history(), snapshot: snapshot() }])
      expect(ended).toEqual([])
      expect(detached).toEqual([])
    })
  })

  describe('falls back to waiting the run out', () => {
    it('without attaching, when only the UI snapshot named a run: the history lacks its turn', async () => {
      const { reads, log } = scripted([history()])
      const run = live()
      const { f, shown, polling } = follower({ ...reads, attach: run.attach })
      f.found(undefined)
      const followed = follow(f)
      await vi.advanceTimersByTimeAsync(1_000)
      expect(await followed).toEqual([])
      expect(run.record.attaches).toBe(0)
      expect(log).toEqual(['history@1000', 'snapshot@1000'])
      expect(shown).toEqual([{ history: history(), snapshot: snapshot() }])
      expect(polling).toEqual([true, false])
    })

    it('when there is no log to follow (204)', async () => {
      const { reads, log } = scripted([history(RUN), history()])
      const attached: number[] = []
      const { f, shown, polling } = follower({
        ...reads,
        attach: async () => {
          attached.push(Date.now())
          return null
        },
      })
      const followed = follow(f)
      await vi.advanceTimersByTimeAsync(1_000 + 2_000)
      expect(await followed).toEqual([])
      expect(attached).toEqual([0])
      expect(shown).toEqual([{ history: history(), snapshot: snapshot() }])
      expect(polling).toEqual([true, false])
      expect(log).toEqual(['history@1000', 'history@3000', 'snapshot@3000'])
    })

    it('when the stream ends with the fallback marker, after what it streamed', async () => {
      const { reads, log } = scripted([history(RUN), history()])
      const run = live()
      const { f, shown, ended, polling } = follower({
        ...reads,
        attach: run.attach,
      })
      const followed = follow(f)
      run.send(...runStarted, textStart, delta('Three days'))
      await vi.advanceTimersByTimeAsync(0)
      expect(f.isPolling()).toBe(false)
      run.send(fallback('truncated'))
      await vi.advanceTimersByTimeAsync(0)
      // The page reads the thread instead, and the composer waits with it.
      expect(f.isPolling()).toBe(true)
      await vi.advanceTimersByTimeAsync(1_000 + 2_000)
      const replies = await followed
      expect(textOf(lastOf(replies))).toEqual(['Three days'])
      // The marker is no part of the reply.
      expect(
        replies.some((r) => r.content?.some((p) => p.type === 'data'))
      ).toBe(false)
      expect(shown).toEqual([{ history: history(), snapshot: snapshot() }])
      expect(ended).toEqual([])
      expect(polling).toEqual([true, false])
      expect(log).toEqual(['history@1000', 'history@3000', 'snapshot@3000'])
      expect(run.record.closed).toBe(true)
    })

    it('when the stream breaks off before the run’s end', async () => {
      const { reads } = scripted([history()])
      const run = live()
      const { f, shown, ended } = follower({ ...reads, attach: run.attach })
      const followed = follow(f)
      run.send(...runStarted, textStart, delta('Three days'))
      await vi.advanceTimersByTimeAsync(0)
      run.breakOff()
      await vi.advanceTimersByTimeAsync(1_000)
      await followed
      expect(shown).toEqual([{ history: history(), snapshot: snapshot() }])
      expect(ended).toEqual([])
    })

    for (const [what, err] of [
      ['fails for the moment', new ApiError('503', 'run log unavailable')],
      ['gets no answer', new TypeError('Failed to fetch')],
      ['is not served (404)', new ApiError('404', 'not found')],
    ] as const) {
      it(`when attaching ${what}`, async () => {
        const { reads } = scripted([history()])
        const run = live(err)
        const { f, shown, failures } = follower({
          ...reads,
          attach: run.attach,
        })
        const followed = follow(f)
        await vi.advanceTimersByTimeAsync(1_000)
        await followed
        expect(run.record.attaches).toBe(1)
        expect(shown).toEqual([{ history: history(), snapshot: snapshot() }])
        expect(failures).toEqual([])
      })
    }
  })

  it('polls until the run ended, then shows the thread as the read after found it', async () => {
    const { reads } = scripted([history(RUN), history()])
    const { f, shown, failures, polling } = follower(reads)
    const followed = follow(f)
    await vi.advanceTimersByTimeAsync(0)
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
    const followed = follow(f)
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
    const { f, shown, failures, detached, polling } = follower(reads)
    const aborted = new AbortController()
    const followed = follow(f, aborted.signal)
    await vi.advanceTimersByTimeAsync(1_000)
    aborted.abort()
    await followed
    expect(shown).toEqual([])
    expect(failures).toEqual([])
    expect(polling).toEqual([true, false])
    // The page stopped following the run before its end.
    expect(detached).toHaveLength(1)
  })

  it('reads the thread again at once when hurried, and resolves once the end shows', async () => {
    const { reads, log } = scripted([history(RUN), history(RUN), history()])
    const { f, shown } = follower(reads)
    const followed = follow(f)
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
    const followed = follow(f)
    await vi.advanceTimersByTimeAsync(0)
    expect(f.untilEnded()).toBeDefined()
    await vi.advanceTimersByTimeAsync(3_000)
    await followed
    expect(f.untilEnded()).toBeUndefined()
  })
})
