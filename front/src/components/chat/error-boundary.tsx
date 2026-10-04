import { Component, type ReactNode } from 'react'
import { useAuiState } from '@assistant-ui/react'
import { RotateCcw, TriangleAlert } from 'lucide-react'
import { Button } from '@/components/ui/button'

interface ChatErrorBoundaryProps {
  fallback: (retry: () => void) => ReactNode
  // resetKey: a new value draws the children again after a failure.
  resetKey?: unknown
  children: ReactNode
}

interface ChatErrorBoundaryState {
  failed: boolean
  resetKey: unknown
}

// ChatErrorBoundary confines a rendering failure to what it wraps: that part
// shows its fallback, and the rest of the chat keeps working.
class ChatErrorBoundary extends Component<
  ChatErrorBoundaryProps,
  ChatErrorBoundaryState
> {
  state: ChatErrorBoundaryState = {
    failed: false,
    resetKey: this.props.resetKey,
  }

  static getDerivedStateFromProps(
    props: ChatErrorBoundaryProps,
    state: ChatErrorBoundaryState
  ): Partial<ChatErrorBoundaryState> | null {
    return props.resetKey === state.resetKey
      ? null
      : { failed: false, resetKey: props.resetKey }
  }

  static getDerivedStateFromError(): Partial<ChatErrorBoundaryState> {
    return { failed: true }
  }

  retry = () => this.setState({ failed: false })

  render() {
    return this.state.failed
      ? this.props.fallback(this.retry)
      : this.props.children
  }
}

// MessageErrorBoundary confines a failure to one message, which says it
// could not be displayed. The message is drawn again when its parts change.
export function MessageErrorBoundary({ children }: { children: ReactNode }) {
  const parts = useAuiState((s) => s.message.parts)
  return (
    <ChatErrorBoundary resetKey={parts} fallback={renderMessageError}>
      {children}
    </ChatErrorBoundary>
  )
}

function renderMessageError() {
  return (
    <p className='flex items-center gap-1.5 rounded-md border border-dashed border-border px-2.5 py-1.5 text-xs text-muted-foreground'>
      <TriangleAlert aria-hidden className='size-3.5 shrink-0' />
      This message could not be displayed.
    </p>
  )
}

// ThreadErrorBoundary stands in for a thread that failed to render, with a
// way to try again; the composer around it keeps working.
export function ThreadErrorBoundary({ children }: { children: ReactNode }) {
  return (
    <ChatErrorBoundary fallback={renderThreadError}>
      {children}
    </ChatErrorBoundary>
  )
}

function renderThreadError(retry: () => void) {
  return (
    <div
      role='alert'
      className='flex min-h-0 flex-1 flex-col items-center justify-center gap-3 px-6 text-center'
    >
      <p className='text-sm text-muted-foreground'>
        This conversation could not be displayed.
      </p>
      <Button variant='outline' size='sm' onClick={retry}>
        <RotateCcw />
        Try again
      </Button>
    </div>
  )
}
