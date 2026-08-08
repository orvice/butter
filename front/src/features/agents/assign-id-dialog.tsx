import { useState } from 'react'
import { toast } from 'sonner'
import { useAssignAgentID } from '@/api/agents'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Fingerprint, TriangleAlert } from 'lucide-react'
import { suggestAgentID, validateAgentID } from './agent-id'

export function AssignAgentIDDialog({
  agentName,
  open,
  onOpenChange,
}: {
  agentName: string | null
  open: boolean
  onOpenChange: (open: boolean) => void
}) {
  // Callers should pass a `key` derived from agentName so the slug suggestion
  // resets whenever the dialog targets a different agent.
  const assignMutation = useAssignAgentID()
  const [slug, setSlug] = useState(() => (agentName ? suggestAgentID(agentName) : ''))
  const [touched, setTouched] = useState(false)

  const error = validateAgentID(slug)

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className='max-w-md'>
        <DialogHeader>
          <DialogTitle className='flex items-center gap-2'>
            <Fingerprint className='size-4' /> Assign Agent ID
          </DialogTitle>
          <DialogDescription>
            Set the permanent slug identifier for <span className='font-medium'>{agentName}</span>.
            Future versions use it in APIs, relationships, and repository paths.
          </DialogDescription>
        </DialogHeader>

        <div className='space-y-2'>
          <Label htmlFor='agent-id-input'>Agent ID</Label>
          <Input
            id='agent-id-input'
            value={slug}
            onChange={(e) => {
              setSlug(e.target.value)
              setTouched(true)
            }}
            placeholder='my-agent'
            autoComplete='off'
            spellCheck={false}
            className='font-mono'
          />
          {error && (touched || slug !== '') ? (
            <p className='text-xs text-destructive'>{error}</p>
          ) : (
            <p className='text-xs text-muted-foreground'>
              1–64 lowercase letters, digits, or hyphens. Must be unique in this workspace.
            </p>
          )}
        </div>

        <div className='flex items-start gap-2 rounded-md border border-amber-500/40 bg-amber-500/10 p-3 text-xs leading-relaxed text-amber-700 dark:text-amber-400'>
          <TriangleAlert className='mt-0.5 size-4 shrink-0' />
          <span>
            The Agent ID is <strong>immutable</strong> — once assigned it can never be changed or
            reused. Choose carefully before confirming.
          </span>
        </div>

        <DialogFooter>
          <Button variant='outline' onClick={() => onOpenChange(false)}>
            Cancel
          </Button>
          <Button
            disabled={!!error || !agentName || assignMutation.isPending}
            onClick={() => {
              if (!agentName) return
              assignMutation.mutate(
                { name: agentName, agent_id: slug },
                {
                  onSuccess: () => {
                    toast.success(`Agent ID "${slug}" assigned to ${agentName}`)
                    onOpenChange(false)
                  },
                  onError: (err) => toast.error(err.message),
                },
              )
            }}
          >
            {assignMutation.isPending ? 'Assigning…' : 'Assign permanently'}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
