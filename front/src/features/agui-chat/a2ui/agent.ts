import {
  HttpAgent,
  type RunAgentInput,
  type RunAgentParameters,
} from '@ag-ui/client'
import { A2UI_CAPABILITY } from './protocol'

interface ResumeEntry {
  interruptId: string
  status: 'resolved' | 'cancelled'
  payload?: unknown
}

type RunInput = RunAgentInput & { resume?: ResumeEntry[] }

// ButterAGUIAgent is the dashboard's HttpAgent. Every run declares A2UI
// support in forwardedProps — which only selects the server's built-in
// catalog — and a form submission can address exactly one Interrupt.
export class ButterAGUIAgent extends HttpAgent {
  private nextResume?: ResumeEntry

  // resumeNextRunWith makes the next run answer exactly this Interrupt.
  //
  // The AG-UI client and the assistant-ui runtime both assume a resume
  // resolves every open interrupt at once: the client refuses to start a
  // run that leaves one of its tracked interrupts unaddressed, and the
  // runtime marks the ones it was not given as cancelled. Butter has no
  // cancel — an unanswered Interrupt simply stays pending — so the run is
  // prepared with the form's entry plus placeholders that satisfy the
  // client's check, and only the form's entry is sent.
  resumeNextRunWith(interruptId: string, payload: unknown) {
    this.nextResume = { interruptId, status: 'resolved', payload }
  }

  clearNextResume() {
    this.nextResume = undefined
  }

  protected prepareRunAgentInput(
    parameters?: RunAgentParameters
  ): RunAgentInput {
    const input = super.prepareRunAgentInput(parameters) as RunInput
    const next = this.nextResume
    if (!next) return input
    const placeholders: ResumeEntry[] = this.pendingInterrupts
      .filter((i) => i.id !== next.interruptId)
      .map((i) => ({ interruptId: i.id, status: 'cancelled' }))
    return { ...input, resume: [next, ...placeholders] }
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
    const next = this.nextResume
    if (next) {
      this.nextResume = undefined
      run.resume = [next]
    }
    return super.requestInit(run)
  }
}
