// Stop ends the open thread's run (ADR-0016 decision 4). Every run outlives
// its request, so aborting the request would only detach this page from a
// run that keeps going. Stop asks the server to stop the run first, then
// cancels it here, which ends its stream at once instead of waiting for the
// run to wind down. A run the page waits out instead of streaming, one it
// found holding the thread when it opened it, has no stream here: Stop then
// waits until the thread shows the run ended (#406).

// LocalRun is the open thread's run as this page follows it.
export interface LocalRun {
  // canCancel reports whether there is a run here to cancel.
  canCancel(): boolean
  // cancel ends the run here: it aborts the request and settles the reply.
  cancel(): void
  // untilEnded is the end of a run the page waits out: the page reads the
  // thread again at once, and the promise resolves once it shows the run
  // ended. It is undefined for a run the page streams.
  untilEnded?(): Promise<void> | undefined
}

// stopper returns what Stop does for one thread. stop asks the server to
// stop the thread's run, and resolves once the server accepted the Stop or
// found no run to stop.
//   - Clicks while a Stop is under way join it, so one Stop sends one request.
//   - A Stop the server refused rejects and leaves the run going here: the
//     run is still going on the server, and the page goes on showing it.
//   - A run that already ended here is not cancelled again. Its stream can
//     end with RUN_ERROR "stopped" before the Stop is answered.
//   - A run the page waits out is not cancelled here: the Stop lasts until
//     the page shows the run ended, so its last turn shows.
export function stopper(
  stop: () => Promise<unknown>,
  run: LocalRun
): () => Promise<void> {
  let pending: Promise<void> | null = null
  return () => {
    pending ??= (async () => {
      await stop()
      const ending = run.untilEnded?.()
      if (ending) await ending
      else if (run.canCancel()) run.cancel()
    })().finally(() => {
      pending = null
    })
    return pending
  }
}
