import { useEffect, useRef, useState } from 'react'
import type { ButterAGUIAgent } from './a2ui/agent'
import type { RunErrorEvent } from './errors'
import type { ThreadHistory } from './history'
import {
  NO_INPUT,
  lastRunOfHistory,
  lastRunOfRunError,
  lastRunOfStop,
  runInput,
  type LastRun,
  type TurnInput,
} from './last-run'

// ThreadLastRun is the open thread's last run when it failed or was
// stopped, as the page knows it (last-run.ts), with how the page tells it
// what the runs' own events do not.
export interface ThreadLastRun {
  lastRun: LastRun | null
  // stopped tells that this page's Stop ended the run here.
  stopped: () => void
  // reading starts a read of the thread's history, and returns what takes
  // the history read: its lastRun becomes the page's, unless a run started
  // here meanwhile, which the read predates.
  reading: () => (history: ThreadHistory) => void
  // following tells that the page follows a run it found holding the thread
  // (RunFollower), and the turn that run started from (runningInput). Like a
  // run started here, it clears the last run, which a read that predates it
  // may have set; and its RUN_ERROR and a Stop of this page quote its turn.
  following: (input: TurnInput) => void
  // runError tells the RUN_ERROR of a run the page follows from its log,
  // which never reaches the AG-UI client: the client sees the runs it starts.
  runError: (event: RunErrorEvent) => void
}

// useLastRun follows the open thread's runs through its AG-UI client. A run
// that starts clears the last run, and so does one that finishes; a
// RUN_ERROR sets it, and so does a Stop that ends the run here.
export function useLastRun(httpAgent: ButterAGUIAgent): ThreadLastRun {
  const [lastRun, setLastRun] = useState<LastRun | null>(null)
  // The input of the latest run started here, or followed here, and how
  // many started here.
  const runs = useRef<{ input: TurnInput; started: number }>({
    input: NO_INPUT,
    started: 0,
  })

  useEffect(() => {
    const sub = httpAgent.subscribe({
      onRunInitialized: ({ input }) => {
        runs.current = {
          input: runInput(input),
          started: runs.current.started + 1,
        }
      },
      onRunStartedEvent: () => setLastRun(null),
      onRunFinishedEvent: () => setLastRun(null),
      onRunErrorEvent: ({ event, input }) => {
        const ended = lastRunOfRunError(event, runInput(input))
        if (ended) setLastRun(ended)
      },
    })
    return () => sub.unsubscribe()
  }, [httpAgent])

  const [tell] = useState(() => ({
    stopped: () => setLastRun(lastRunOfStop(runs.current.input)),
    reading: () => {
      const started = runs.current.started
      return (history: ThreadHistory) => {
        if (runs.current.started === started) {
          setLastRun(lastRunOfHistory(history))
        }
      }
    },
    following: (input: TurnInput) => {
      runs.current = { ...runs.current, input }
      setLastRun(null)
    },
    runError: (event: RunErrorEvent) => {
      const ended = lastRunOfRunError(event, runs.current.input)
      if (ended) setLastRun(ended)
    },
  }))
  return { lastRun, ...tell }
}
