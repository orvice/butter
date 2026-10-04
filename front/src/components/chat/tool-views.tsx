import { useId, useState } from 'react'
import type { ToolCallMessagePartProps } from '@assistant-ui/react'
import { ChevronDown, Loader2, Wrench } from 'lucide-react'
import { cn } from '@/lib/utils'
import {
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
} from '@/components/ui/collapsible'
import {
  awaitsAnswer,
  formatJson,
  hasArgs,
  humanInputQuestion,
  renderUIFailed,
  toolCallState,
  type ToolCallState,
} from './tool-call'

// ToolCallView is how a chat shows a tool call: collapsed to its name and
// state, it expands to the call's arguments and result. Its button is named
// by the tool alone; the state is its description.
export function ToolCallView({
  toolName,
  args,
  result,
  isError,
  status,
}: ToolCallMessagePartProps) {
  const [open, setOpen] = useState(false)
  const nameId = useId()
  const stateId = useId()
  const state = toolCallState({ result, isError, status })
  const showArgs = hasArgs(args)
  const showResult = result !== undefined

  return (
    <Collapsible
      open={open}
      onOpenChange={setOpen}
      data-tool-call={toolName}
      className='group/tool my-2 max-w-[72ch] overflow-hidden rounded-md border border-border/60 bg-muted/20'
    >
      <CollapsibleTrigger
        aria-labelledby={nameId}
        aria-describedby={state === 'idle' ? undefined : stateId}
        className='flex w-full items-center gap-2 px-2.5 py-1.5 text-left transition-colors hover:bg-muted/45'
      >
        <Wrench
          aria-hidden
          className='size-3.5 shrink-0 text-muted-foreground'
        />
        <span
          id={nameId}
          className='truncate font-mono text-xs text-foreground/85'
        >
          {toolName}
        </span>
        <ToolCallStateLabel id={stateId} state={state} />
        <ChevronDown
          aria-hidden
          className='ml-auto size-3.5 shrink-0 text-muted-foreground transition-transform group-data-[state=open]/tool:rotate-180'
        />
      </CollapsibleTrigger>
      <CollapsibleContent className='space-y-2 border-t border-border/60 px-2.5 py-2'>
        {showArgs && <ToolCallSection label='Arguments' value={args} />}
        {showResult && <ToolCallSection label='Result' value={result} />}
        {!showArgs && !showResult && (
          <p className='text-xs text-muted-foreground'>No arguments.</p>
        )}
      </CollapsibleContent>
    </Collapsible>
  )
}

function ToolCallStateLabel({
  id,
  state,
}: {
  id: string
  state: ToolCallState
}) {
  switch (state) {
    case 'running':
      return (
        <span id={id} className='shrink-0 text-muted-foreground'>
          <Loader2 aria-hidden className='size-3 animate-spin' />
          <span className='sr-only'>running</span>
        </span>
      )
    case 'idle':
      return null
    default:
      return (
        <span
          id={id}
          className={cn(
            'shrink-0 text-[0.7rem]',
            state === 'failed' ? 'text-destructive' : 'text-muted-foreground'
          )}
        >
          {state}
        </span>
      )
  }
}

function ToolCallSection({ label, value }: { label: string; value: unknown }) {
  return (
    <div>
      <div className='mb-1 text-[0.68rem] text-muted-foreground'>{label}</div>
      <pre className='max-h-64 scrollbar-thin overflow-auto rounded bg-muted/55 p-2 font-mono text-xs leading-5'>
        {formatJson(value)}
      </pre>
    </div>
  )
}

// RenderUIToolView keeps a render_ui call out of the way: the card it drew
// is what the reply shows. A call that drew no card shows, with its error.
export function RenderUIToolView(props: ToolCallMessagePartProps) {
  return renderUIFailed(props.result) ? <ToolCallView {...props} /> : null
}

// HumanInputToolView shows the question a Workflow's Human Input node asks
// and, once it is answered, the answer.
export function HumanInputToolView({
  args,
  result,
  isError,
  status,
}: ToolCallMessagePartProps) {
  const waiting = awaitsAnswer({ result, isError, status })
  return (
    <div
      className={cn(
        'my-2 max-w-[72ch] rounded-lg border px-3.5 py-3 text-sm',
        waiting
          ? 'border-amber-500/35 bg-amber-500/5'
          : 'border-border/70 bg-muted/30'
      )}
    >
      <p className='font-medium text-foreground'>
        {waiting ? 'Waiting for input' : 'Human Input'}
      </p>
      <p className='mt-1 text-[0.85rem] leading-5 text-muted-foreground'>
        {humanInputQuestion(args)}
      </p>
      {result !== undefined && (
        <p className='mt-2 border-l-2 border-border pl-2 text-[0.85rem] text-foreground'>
          {formatJson(result)}
        </p>
      )}
      {waiting && (
        <p className='mt-2 text-[0.75rem] text-amber-600 dark:text-amber-400'>
          Send a message below to answer this question.
        </p>
      )}
    </div>
  )
}
