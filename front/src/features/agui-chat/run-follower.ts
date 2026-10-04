import type { ChatModelRunResult, ThreadMessage } from '@assistant-ui/react'
import { AGUI_FALLBACK_EVENT, type AGUIEvent } from '@/api/agui'
import { ApiError } from '@/api/client'
import type { ThreadHistory, ThreadRead, ThreadReads } from './history'
import {
  RunFold,
  toolCallIdsOf,
  type RunEffects,
  type RunEndEvent,
} from './run-fold'

// A thread can be held by a run this page did not start: one started before
// a reload, before the page left the thread and came back, or in another tab
// (ADR-0016 decision 8). The thread's reads then show it as the run found
// it, with the turn that started the run, and name the run (`running`). The
// page shows that run's reply as running, and follows the run: it attaches
// to the run's log and streams the reply from it, as it streams a run it
// started. When the log cannot carry the run to its end, the page waits the
// run out instead: it reads the thread again, with backoff, until no run
// holds it, then shows the thread as the run left it.

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
  // ended is told the event that ended a run the page streamed from its
  // log, once its reply shows it: RUN_FINISHED, or RUN_ERROR, a failure or
  // a Stop. The reply is then the run's, and nothing is read.
  ended?: (event: RunEndEvent) => void
  // detached is told when the page stopped following the run before its
  // end: a Stop that ended it here, leaving the thread, or deleting it. The
  // run goes on, as one the page started does once its request is aborted.
  detached?: () => void
}

// FollowOptions are the history adapter's resume() options that follow
// reads: the signal that ends the run's resume, and the thread's messages
// before the run's reply.
export interface FollowOptions {
  abortSignal: AbortSignal
  messages?: readonly ThreadMessage[]
}

const NO_EFFECTS: RunEffects = { a2ui: () => {}, state: () => {} }

// RunFollower follows a run the page found holding the thread, for one
// thread's runtime. The thread's history adapter hands the run over from
// resume() (follow), which the runtime runs under the run's reply, so the
// reply shows as running until follow returns. Streamed from the run's log,
// the reply grows as the run goes, and the run's events move the A2UI store
// and the shared state (setEffects). isPolling tells the page it waits the
// run out by reading the thread instead (pollUntilEnded): the composer stays
// disabled meanwhile.
export class RunFollower {
  private readonly listeners = new Set<() => void>()
  private readonly reads: ThreadReads
  private effects: RunEffects = NO_EFFECTS
  private end: RunEnd | null = null
  private runId: string | null = null
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

  // setEffects sets what a run streamed from its log changes besides its
  // reply: the A2UI store and the shared state. The page sets them from
  // inside its runtime's provider, where the shared state's setter is,
  // before any run streams.
  setEffects(effects: RunEffects) {
    this.effects = effects
  }

  // found tells which run the thread's history found holding the thread:
  // the run follow streams, as the history ends with its turn. Without one,
  // only the UI snapshot named a run, one that started after the history was
  // read, whose turn the page does not show: follow waits that run out. So
  // it does with a log that replays another run than the history's.
  found(run: { runId: string } | undefined) {
    this.runId = run?.runId ?? null
  }

  subscribe = (listener: () => void) => {
    this.listeners.add(listener)
    return () => {
      this.listeners.delete(listener)
    }
  }

  isPolling = () => this.polling

  // follow follows the run until it ended: it is the history adapter's
  // resume(), and yields the run's reply as it goes. For the run the history
  // found (found), it attaches to the run's log (GET …/threads/:thread_id/run),
  // which replays the run from RUN_STARTED and then follows it, and streams
  // the reply from it (stream). It waits the run out by reading the thread
  // instead (pollUntilEnded) when the log cannot carry that run to its end:
  // attaching answers 204 or fails, the log is another run's, or the stream
  // ends with the butter.fallback marker or breaks off. It returns once
  // options.abortSignal aborts: the page stopped following the run, which
  // goes on (detached).
  async *follow(
    options: FollowOptions
  ): AsyncGenerator<ChatModelRunResult, void, undefined> {
    const signal = options.abortSignal
    try {
      const ended = yield* this.stream(signal, options.messages ?? [])
      if (ended || signal.aborted) return
      await this.pollUntilEnded(signal)
    } finally {
      if (signal.aborted) this.end?.detached?.()
    }
  }

  // stream streams the run's reply from its log, folded as the AG-UI runtime
  // folds a run it streams itself (RunFold), and reports whether the run
  // ended in it. messages are the thread's messages before the reply. The
  // event that ended the run goes to ended once the reply shows it.
  private async *stream(
    signal: AbortSignal,
    messages: readonly ThreadMessage[]
  ): AsyncGenerator<ChatModelRunResult, boolean, undefined> {
    if (this.runId === null) return false
    const events = await this.attach(signal)
    if (!events || signal.aborted) return false
    // The fold applies the effects the page set last.
    const effects: RunEffects = {
      a2ui: (value) => this.effects.a2ui(value),
      state: (update) => this.effects.state(update),
    }
    const fold = new RunFold(effects, toolCallIdsOf(messages))
    try {
      for await (const event of events) {
        if (signal.aborted) return false
        if (event.type === 'CUSTOM' && event.name === AGUI_FALLBACK_EVENT) {
          return false
        }
        // The log replays its run from RUN_STARTED, under the run's runId.
        if (event.type === 'RUN_STARTED' && event.runId !== this.runId) {
          return false
        }
        const reply = fold.handle(event)
        if (reply) yield reply
        const endEvent = fold.end
        if (endEvent) {
          if (!signal.aborted) this.end?.ended?.(endEvent)
          return true
        }
      }
    } catch {
      // The stream broke off before the run's end.
    }
    return false
  }

  // attach opens the run's log, or answers null when there is none to
  // follow or it cannot be read for now, or not by this server: the
  // thread's reads then tell the run's end.
  private async attach(
    signal: AbortSignal
  ): Promise<AsyncIterable<AGUIEvent> | null> {
    try {
      return (await this.reads.attach?.(signal)) ?? null
    } catch {
      return null
    }
  }

  // pollUntilEnded waits the run out by reading the thread (awaitRunEnd),
  // then shows the thread as the run left it, in place of what the stream
  // showed of it. It is the fallback of live re-attach, for a run whose log
  // cannot be followed to its end: the log is gone (204) or cannot be read,
  // or it was truncated, expired or lost (butter.fallback). It returns early,
  // showing nothing, once signal aborts: the page stopped following the run.
  // A read that fails for good goes to failed.
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
  // undefined while the page waits for no run, and while it streams the run
  // from its log: Stop then ends the stream here, as it ends one of a run
  // the page started.
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
