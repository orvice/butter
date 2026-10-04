// Stop ends the open thread's run (ADR-0016 decision 4). Every run outlives
// its request, so aborting the request would only detach this page from a
// run that keeps going. Stop asks the server to stop the run first, then
// cancels it here, which ends its stream at once instead of waiting for the
// run to wind down.

// LocalRun is the open thread's run as this page follows it.
export interface LocalRun {
  // canCancel reports whether there is a run here to cancel.
  canCancel(): boolean
  // cancel ends the run here: it aborts the request and settles the reply.
  cancel(): void
}

// stopper returns what Stop does for one thread. stop asks the server to
// stop the thread's run, and resolves once the server accepted the Stop or
// found no run to stop.
//   - Clicks while a Stop is under way join it, so one Stop sends one request.
//   - A Stop the server refused rejects and leaves the run going here: the
//     run is still going on the server, and the page goes on showing it.
//   - A run that already ended here is not cancelled again. Its stream can
//     end with RUN_ERROR "stopped" before the Stop is answered.
export function stopper(
  stop: () => Promise<unknown>,
  run: LocalRun
): () => Promise<void> {
  let pending: Promise<void> | null = null
  return () => {
    pending ??= (async () => {
      await stop()
      if (run.canCancel()) run.cancel()
    })().finally(() => {
      pending = null
    })
    return pending
  }
}
