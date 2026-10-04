import { ApiError } from '@/api/client'
import type { ThreadHistory, ThreadRead, ThreadReads } from './history'

// A thread can be held by a run this page did not start: one started before
// a reload, before the page left the thread and came back, or in another tab
// (ADR-0016 decision 8). The thread's reads then show it as the run found
// it, with the turn that started the run, and name the run (`running`). The
// page shows that run's reply as running, and waits the run out: it reads the
// thread again, with backoff, until no run holds it, then shows the thread as
// the run left it.

// POLL_DELAYS_MS paces the reads of a thread whose run the page waits out:
// the first a second after the read that found the run, then twice as long
// each time, up to five seconds.
export const POLL_DELAYS_MS: readonly number[] = [1_000, 2_000, 4_000, 5_000]

// Backoff spaces out the reads of one wait (POLL_DELAYS_MS). hurry has the
// next read go out at once and starts the delays over, for a run about to
// end: one a Stop reached.
export class Backoff {
  private attempt = 0
  private hurried = false
  private cut: (() => void) | null = null

  // wait resolves when the next read is due, and at once when signal aborts.
  wait(signal: AbortSignal): Promise<void> {
    if (signal.aborted) return Promise.resolve()
    if (this.hurried) {
      this.hurried = false
      return Promise.resolve()
    }
    const delay =
      POLL_DELAYS_MS[Math.min(this.attempt, POLL_DELAYS_MS.length - 1)]
    this.attempt++
    return new Promise((resolve) => {
      const done = () => {
        clearTimeout(timer)
        signal.removeEventListener('abort', done)
        this.cut = null
        resolve()
      }
      const timer = setTimeout(done, delay)
      signal.addEventListener('abort', done)
      this.cut = done
    })
  }

  hurry() {
    this.attempt = 0
    if (this.cut) this.cut()
    else this.hurried = true
  }
}

// transient reports a failed read worth trying again: the server, or a proxy
// in front of it, is unavailable for the moment (502, 503, 504), or the
// request got no answer, which fetch reports as a TypeError. Any other
// failure is final. So is a 409: the reads never answer one for a run.
export function transient(err: unknown): boolean {
  if (err instanceof ApiError) return TRANSIENT_STATUS.has(err.code)
  return err instanceof TypeError
}

const TRANSIENT_STATUS = new Set(['502', '503', '504'])

// awaitRunEnd reads the thread until no run holds it, spaced out by backoff,
// and returns that read: the history once it names no run, then the UI
// snapshot. A snapshot that names a run means one started in between, which
// it waits out too. A read that failed for the moment is tried again with
// the next one; any other failure rejects. It returns null once signal
// aborts, and reads nothing more.
export async function awaitRunEnd(
  reads: ThreadReads,
  signal: AbortSignal,
  backoff: Backoff = new Backoff()
): Promise<ThreadRead | null> {
  for (;;) {
    await backoff.wait(signal)
    if (signal.aborted) return null
    try {
      const history = await reads.history(signal)
      if (signal.aborted) return null
      if (history.running) continue
      const snapshot = await reads.snapshot(signal)
      if (signal.aborted) return null
      if (!snapshot.running) return { history, snapshot }
    } catch (err) {
      if (signal.aborted) return null
      if (!transient(err)) throw err
    }
  }
}

// RunEnd is where the end of a followed run goes.
export interface RunEnd {
  // show shows the thread as the read after the run found it.
  show: (read: ThreadRead) => void
  // failed reports a read that failed for good; the page offers Retry.
  failed: (err: unknown) => void
  // reading is told of each read of the history as it starts, and returns
  // what takes that read's history (useLastRun's reading): the history of
  // the read that ends the wait tells how the run ended, if it failed or
  // was stopped.
  reading?: () => (history: ThreadHistory) => void
}

// RunFollower follows a run the page found holding the thread, for one
// thread's runtime. The thread's history adapter hands the run over from
// resume() (follow), which the runtime runs under the run's reply, so the
// reply shows as running until follow returns. isPolling tells the page it
// waits the run out by reading the thread (pollUntilEnded): the composer
// stays disabled meanwhile.
export class RunFollower {
  private readonly listeners = new Set<() => void>()
  private readonly reads: ThreadReads
  private end: RunEnd | null = null
  private polling = false
  private backoff: Backoff | null = null
  private waiters: Array<() => void> = []

  constructor(reads: ThreadReads) {
    this.reads = reads
  }

  // endWith sets where the end of a followed run goes. The page sets it once
  // its runtime is up, before any run can end.
  endWith(end: RunEnd) {
    this.end = end
  }

  subscribe = (listener: () => void) => {
    this.listeners.add(listener)
    return () => {
      this.listeners.delete(listener)
    }
  }

  isPolling = () => this.polling

  // follow follows the run until it ended: it is the history adapter's
  // resume(). For now the page waits every run out (pollUntilEnded). Live
  // re-attach (#407) goes here: it attaches to the run's log
  // (GET …/threads/:thread_id/run) and streams the reply, and falls back to
  // pollUntilEnded when attaching answers 204 or the stream ends with the
  // butter.fallback marker.
  follow(signal: AbortSignal): Promise<void> {
    return this.pollUntilEnded(signal)
  }

  // pollUntilEnded waits the run out by reading the thread (awaitRunEnd),
  // then shows the thread as the run left it. It is the fallback of live
  // re-attach, for a run whose log cannot be followed to its end: the log is
  // gone (204), or it was truncated, expired or lost (butter.fallback). It
  // returns early, showing nothing, once signal aborts: the page stopped
  // following the run. A read that fails for good goes to failed.
  async pollUntilEnded(signal: AbortSignal): Promise<void> {
    const backoff = new Backoff()
    this.backoff = backoff
    this.setPolling(true)
    // takeHistory takes the history of the latest read, the one that ends
    // the wait once awaitRunEnd returns.
    let takeHistory: ((history: ThreadHistory) => void) | undefined
    const reads: ThreadReads = {
      history: (s) => {
        takeHistory = this.end?.reading?.()
        return this.reads.history(s)
      },
      snapshot: (s) => this.reads.snapshot(s),
    }
    try {
      const read = await awaitRunEnd(reads, signal, backoff)
      if (read) {
        this.end?.show(read)
        takeHistory?.(read.history)
      }
    } catch (err) {
      this.end?.failed(err)
    } finally {
      this.backoff = null
      this.setPolling(false)
      for (const resolve of this.waiters.splice(0)) resolve()
    }
  }

  // untilEnded is the end of the run the page waits out, for a Stop that
  // reached it: the thread is read again at once, as the run ends shortly
  // after, and the promise resolves once the page stopped waiting. It is
  // undefined while the page waits for no run.
  untilEnded(): Promise<void> | undefined {
    if (!this.polling) return undefined
    this.backoff?.hurry()
    return new Promise((resolve) => this.waiters.push(resolve))
  }

  private setPolling(polling: boolean) {
    this.polling = polling
    for (const listener of this.listeners) listener()
  }
}
