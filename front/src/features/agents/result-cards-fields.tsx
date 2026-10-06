import { LayoutTemplate } from 'lucide-react'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { Label } from '@/components/ui/label'
import { RadioGroup, RadioGroupItem } from '@/components/ui/radio-group'
import type {
  ResultCardsFormValues,
  ResultCardsGeneration,
} from './result-cards-config'

interface ResultCardsConfigurationCardProps {
  value: ResultCardsFormValues
  onChange: (value: ResultCardsFormValues) => void
}

const GENERATIONS: readonly {
  value: ResultCardsGeneration
  label: string
  description: string
}[] = [
  {
    value: 'allowed',
    label: 'Allowed',
    description: 'Models can show a read-only card next to their answer.',
  },
  {
    value: 'off',
    label: 'Off for this agent and its sub-agents',
    description:
      'Models answer in text only. Cards already in a thread stay there.',
  },
]

export function ResultCardsConfigurationCard({
  value,
  onChange,
}: ResultCardsConfigurationCardProps) {
  function changeGeneration(generation: string) {
    if (generation !== 'allowed' && generation !== 'off') return
    onChange({ ...value, generation })
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle className='flex items-center gap-2'>
          <LayoutTemplate className='size-4' />
          Result cards
        </CardTitle>
        <CardDescription>
          Whether this agent&apos;s models may show results as cards in Chat.
        </CardDescription>
      </CardHeader>
      <CardContent className='space-y-3'>
        <RadioGroup
          aria-label='Result cards'
          value={value.generation}
          onValueChange={changeGeneration}
          className='grid gap-2 md:grid-cols-2'
        >
          {GENERATIONS.map((option) => {
            const id = `result-cards-${option.value}`
            return (
              <Label
                key={option.value}
                htmlFor={id}
                className={`flex cursor-pointer items-start gap-3 rounded-md border p-3 transition-colors ${
                  value.generation === option.value
                    ? 'border-primary bg-primary/10'
                    : 'hover:bg-muted'
                }`}
              >
                <RadioGroupItem
                  id={id}
                  value={option.value}
                  className='mt-0.5'
                />
                <span className='space-y-1'>
                  <span className='block text-sm font-medium'>
                    {option.label}
                  </span>
                  <span className='block text-xs text-muted-foreground'>
                    {option.description}
                  </span>
                </span>
              </Label>
            )
          })}
        </RadioGroup>
        <p className='text-xs text-muted-foreground'>
          A parent agent that turns cards off wins while this agent runs under
          it. Human Input forms still show either way.
        </p>
      </CardContent>
    </Card>
  )
}
