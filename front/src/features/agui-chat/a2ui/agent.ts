import {
  HttpAgent,
  type RunAgentInput,
  type RunAgentParameters,
} from '@ag-ui/client'
import { A2UI_CAPABILITY } from './protocol'

export interface ResumeEntry {
  interruptId: string
  status: 'resolved' | 'cancelled'
  payload?: unknown
}

type RunInput = RunAgentInput & { resume?: ResumeEntry[] }

// NextRun is how the next run answers the thread's open Interrupts. Butter
// answers them one at a time (ADR-0002, docs/api.md "Human-in-the-loop"):
//   - answer: one resume entry for one Interrupt; the others stay open;
//   - message: no resume at all; the server answers its oldest open
//     Interrupt with the run's trailing user message.
export type NextRun =
  { kind: 'answer'; entry: ResumeEntry } | { kind: 'message' }

// clientResume is the resume the AG-UI client checks before it starts the
// run. The client refuses a run that leaves an Interrupt it tracks
// unaddressed, so each one the run does not answer gets a placeholder. The
// placeholders never leave the client (wireResume).
export function clientResume(
  next: NextRun,
  open: ReadonlyArray<{ id: string }>
): ResumeEntry[] {
  const answered = next.kind === 'answer' ? [next.entry] : []
  const placeholders = open
    .filter((i) => !answered.some((e) => e.interruptId === i.id))
    .map((i): ResumeEntry => ({ interruptId: i.id, status: 'cancelled' }))
  return [...answered, ...placeholders]
}

// wireResume is the resume the request carries: the one answer, or none.
export function wireResume(next: NextRun): ResumeEntry[] | undefined {
  return next.kind === 'answer' ? [next.entry] : undefined
}

// ButterAGUIAgent is the dashboard's HttpAgent. Every run declares A2UI
// support in forwardedProps, which selects the server's built-in catalog and
// makes the run's RUN_FINISHED list every Interrupt still open. A run armed
// while Interrupts are open answers one of them: the one it addresses, or the
// server's oldest.
//
// The AG-UI client and the assistant-ui runtime both assume a resume
// resolves every open interrupt at once: the client refuses to start a run
// that leaves one of its tracked interrupts unaddressed, and the runtime
// marks the ones it was not given as cancelled. Butter has no cancel — an
// unanswered Interrupt stays open, and RUN_FINISHED lists it again — so an
// armed run is prepared with placeholders that satisfy the client's check,
// and its request carries only the answer, if any.
export class ButterAGUIAgent extends HttpAgent {
  private nextRun?: NextRun

  // resumeNextRunWith makes the next run answer exactly this Interrupt.
  resumeNextRunWith(interruptId: string, payload: unknown) {
    this.nextRun = {
      kind: 'answer',
      entry: { interruptId, status: 'resolved', payload },
    }
  }

  // sendNextRunAsMessage makes the next run a plain message with no resume,
  // which the server matches to its oldest open Interrupt. The client cannot
  // pick that one itself: RUN_FINISHED lists the Interrupts a run raised
  // before the older ones still open.
  sendNextRunAsMessage() {
    this.nextRun = { kind: 'message' }
  }

  clearNextRun() {
    this.nextRun = undefined
  }

  protected prepareRunAgentInput(
    parameters?: RunAgentParameters
  ): RunAgentInput {
    const input = super.prepareRunAgentInput(parameters) as RunInput
    if (!this.nextRun) return input
    return {
      ...input,
      resume: clientResume(this.nextRun, this.pendingInterrupts),
    }
  }

  protected requestInit(input: RunAgentInput): RequestInit {
    const run: RunInput = {
      ...(input as RunInput),
      forwardedProps: {
        ...((input.forwardedProps as Record<string, unknown> | undefined) ??
          {}),
        butterA2UI: A2UI_CAPABILITY,
      },
    }
    const next = this.nextRun
    if (next) {
      this.nextRun = undefined
      const resume = wireResume(next)
      if (resume) run.resume = resume
      else delete run.resume
    }
    return super.requestInit(run)
  }
}
