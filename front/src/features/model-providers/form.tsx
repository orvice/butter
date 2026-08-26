import { useEffect } from 'react'
import { z } from 'zod'
import { useFieldArray, useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import type { ModelProvider } from '@/types/api'
import { Plus, Trash2 } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from '@/components/ui/form'
import { Input } from '@/components/ui/input'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { PageActions } from '@/components/butter/page-parts'
import {
  MAX_CONTEXT_WINDOW_TOKENS,
  blankModelRow,
  modelsSchema,
  modelsToRows,
  rowsToModels,
} from './model-rows'

const schema = z.object({
  name: z.string().min(1, 'Name is required'),
  type: z.string().min(1, 'Type is required'),
  api_key: z.string().optional(),
  base_url: z.string().optional(),
  models: modelsSchema,
})

type FormValues = z.infer<typeof schema>

type ModelProviderFormProps = {
  mode: 'create' | 'edit'
  initialValue?: ModelProvider
  loading?: boolean
  submitLabel: string
  onCancel: () => void
  onSubmit: (provider: ModelProvider) => void
}

function providerToFormValues(provider?: ModelProvider): FormValues {
  return {
    name: provider?.name ?? '',
    type: provider?.type || 'openai',
    api_key: provider?.api_key ?? '',
    base_url: provider?.base_url ?? '',
    models: modelsToRows(provider?.models),
  }
}

export function ModelProviderForm({
  mode,
  initialValue,
  loading,
  submitLabel,
  onCancel,
  onSubmit,
}: ModelProviderFormProps) {
  const form = useForm<FormValues>({
    resolver: zodResolver(schema),
    defaultValues: providerToFormValues(),
  })
  const { fields, append, remove } = useFieldArray({
    control: form.control,
    name: 'models',
  })

  useEffect(() => {
    form.reset(providerToFormValues(initialValue))
  }, [form, initialValue])

  function handleSubmit(values: FormValues) {
    onSubmit({
      name: values.name,
      type: values.type,
      api_key: values.api_key || undefined,
      base_url: values.base_url || undefined,
      models: rowsToModels(values.models),
    })
  }

  // Array-level errors from modelsSchema (e.g. zero rows) land under
  // errors.models.root; per-row errors are rendered inside each row.
  const modelsError =
    form.formState.errors.models?.root?.message ??
    form.formState.errors.models?.message

  return (
    <Form {...form}>
      <form
        noValidate
        onSubmit={form.handleSubmit(handleSubmit)}
        className='space-y-6'
      >
        <Card>
          <CardHeader>
            <CardTitle>Provider</CardTitle>
          </CardHeader>
          <CardContent className='space-y-4'>
            <FormField
              control={form.control}
              name='name'
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Name</FormLabel>
                  <FormControl>
                    <Input
                      placeholder='openai'
                      {...field}
                      disabled={mode === 'edit'}
                    />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />
            <FormField
              control={form.control}
              name='type'
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Type</FormLabel>
                  <Select onValueChange={field.onChange} value={field.value}>
                    <FormControl>
                      <SelectTrigger>
                        <SelectValue placeholder='Select provider type' />
                      </SelectTrigger>
                    </FormControl>
                    <SelectContent>
                      <SelectItem value='openai'>OpenAI</SelectItem>
                      <SelectItem value='gemini'>Gemini</SelectItem>
                    </SelectContent>
                  </Select>
                  <FormMessage />
                </FormItem>
              )}
            />
            <FormField
              control={form.control}
              name='api_key'
              render={({ field }) => (
                <FormItem>
                  <FormLabel>API Key</FormLabel>
                  <FormControl>
                    <Input
                      type='password'
                      placeholder='sk-... or env-resolved value'
                      {...field}
                    />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />
            <FormField
              control={form.control}
              name='base_url'
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Base URL</FormLabel>
                  <FormControl>
                    <Input placeholder='https://api.openai.com/v1' {...field} />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />
          </CardContent>
        </Card>

        <Card>
          <CardHeader className='flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between'>
            <CardTitle>Models</CardTitle>
            <Button
              type='button'
              variant='outline'
              size='sm'
              onClick={() => append(blankModelRow())}
            >
              <Plus className='mr-1 h-3 w-3' /> Add model
            </Button>
          </CardHeader>
          <CardContent className='space-y-4'>
            <p className='text-xs text-muted-foreground'>
              One row per model. The alias is optional; agents can reference a
              model by its alias.
            </p>
            {fields.length === 0 ? (
              <div className='rounded-md border border-dashed p-4 text-sm text-muted-foreground'>
                No models configured yet. Add at least one model.
              </div>
            ) : (
              <ul aria-label='Models' className='space-y-3'>
                {fields.map((row, index) => (
                  <li
                    key={row.id}
                    className='grid gap-3 rounded-md border p-3 lg:grid-cols-[minmax(0,1fr)_minmax(0,1fr)_minmax(10rem,0.8fr)_auto] lg:items-start'
                  >
                    <FormField
                      control={form.control}
                      name={`models.${index}.name`}
                      render={({ field }) => (
                        <FormItem>
                          <FormLabel>Model ID</FormLabel>
                          <FormControl>
                            <Input placeholder='gpt-4o' {...field} />
                          </FormControl>
                          <FormMessage />
                        </FormItem>
                      )}
                    />
                    <FormField
                      control={form.control}
                      name={`models.${index}.alias`}
                      render={({ field }) => (
                        <FormItem>
                          <FormLabel>Alias</FormLabel>
                          <FormControl>
                            <Input placeholder='Optional alias' {...field} />
                          </FormControl>
                          <FormMessage />
                        </FormItem>
                      )}
                    />
                    <FormField
                      control={form.control}
                      name={`models.${index}.contextWindowTokens`}
                      render={({ field }) => (
                        <FormItem>
                          <FormLabel>Context Window</FormLabel>
                          <FormControl>
                            <Input
                              type='number'
                              min={0}
                              max={MAX_CONTEXT_WINDOW_TOKENS}
                              step={1}
                              placeholder='Optional tokens'
                              {...field}
                            />
                          </FormControl>
                          <FormMessage />
                        </FormItem>
                      )}
                    />
                    <Button
                      type='button'
                      variant='ghost'
                      size='icon'
                      className='lg:mt-8'
                      onClick={() => remove(index)}
                      aria-label={`Remove model ${index + 1}`}
                    >
                      <Trash2 className='h-4 w-4' />
                    </Button>
                  </li>
                ))}
              </ul>
            )}
            {modelsError && (
              <p role='alert' className='text-sm font-medium text-destructive'>
                {modelsError}
              </p>
            )}
          </CardContent>
        </Card>

        <PageActions>
          <Button type='button' variant='outline' onClick={onCancel}>
            Cancel
          </Button>
          <Button type='submit' disabled={loading}>
            {loading ? 'Saving...' : submitLabel}
          </Button>
        </PageActions>
      </form>
    </Form>
  )
}
