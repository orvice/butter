import { useState } from 'react'
import { LinearAppCredentialState } from '@/gen/agents/v1/linear_pb'
import { Check, Copy, TriangleAlert } from 'lucide-react'
import { toast } from 'sonner'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'

/** A read-only value the operator pastes into Linear, with a copy button. */
export function CopyField({
  id,
  label,
  value,
  emptyHint,
}: {
  id: string
  label: string
  value: string
  emptyHint: string
}) {
  const [copied, setCopied] = useState(false)
  return (
    <div className='space-y-1'>
      <label htmlFor={id} className='text-sm font-medium'>
        {label}
      </label>
      <div className='flex gap-2'>
        <Input
          id={id}
          readOnly
          value={value}
          placeholder={emptyHint}
          className='font-mono text-xs'
        />
        <Button
          type='button'
          variant='outline'
          size='icon'
          aria-label={`Copy ${label}`}
          disabled={!value}
          onClick={async () => {
            try {
              await navigator.clipboard.writeText(value)
              setCopied(true)
              setTimeout(() => setCopied(false), 1500)
            } catch {
              toast.error('Could not copy to the clipboard')
            }
          }}
        >
          {copied ? (
            <Check className='h-4 w-4' />
          ) : (
            <Copy className='h-4 w-4' />
          )}
        </Button>
      </div>
    </div>
  )
}

export function CredentialStateBadge({
  state,
}: {
  state: LinearAppCredentialState
}) {
  if (state === LinearAppCredentialState.COMPLETE) {
    return <Badge variant='outline'>Secrets set</Badge>
  }
  return (
    <Badge variant='outline' className='text-muted-foreground'>
      Secrets missing
    </Badge>
  )
}

/**
 * Anyone admitted to a Linear App can drive its Agent. For a Pi or Cursor
 * Agent that means running commands on the ButterBox (ADR-0015 §10).
 */
export function BoxAgentWarning() {
  return (
    <Alert>
      <TriangleAlert className='h-4 w-4' />
      <AlertTitle>This Agent runs commands on a ButterBox</AlertTitle>
      <AlertDescription>
        Anyone admitted to this Linear App can make the Agent run commands on
        its box, including anything the box user can reach. Restrict the
        allowlist to people you would give that access.
      </AlertDescription>
    </Alert>
  )
}
