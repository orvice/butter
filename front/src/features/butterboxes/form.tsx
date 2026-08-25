import { useEffect } from 'react'
import { z } from 'zod'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import type { ButterBox } from '@/gen/agents/v1/butterbox_pb'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import {
  Form,
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from '@/components/ui/form'
import { Input } from '@/components/ui/input'
import { Switch } from '@/components/ui/switch'
import { PageActions } from '@/components/butter/page-parts'

function isHttpUrl(value: string): boolean {
  try {
    const url = new URL(value)
    return url.protocol === 'http:' || url.protocol === 'https:'
  } catch {
    return false
  }
}

const schema = z.object({
  name: z.string().min(1, 'Name is required'),
  baseUrl: z
    .string()
    .min(1, 'Base URL is required')
    .refine(isHttpUrl, 'Must be an absolute http(s) URL'),
  enabled: z.boolean(),
  token: z.string().optional(),
})

export type ButterBoxFormValues = z.infer<typeof schema>

type ButterBoxFormProps = {
  mode: 'create' | 'edit'
  initialValue?: ButterBox
  loading?: boolean
  submitLabel: string
  onCancel: () => void
  onSubmit: (values: ButterBoxFormValues) => void
}

export function ButterBoxForm({
  mode,
  initialValue,
  loading,
  submitLabel,
  onCancel,
  onSubmit,
}: ButterBoxFormProps) {
  const form = useForm<ButterBoxFormValues>({
    resolver: zodResolver(schema),
    defaultValues: {
      name: '',
      baseUrl: '',
      enabled: true,
      token: '',
    },
  })

  useEffect(() => {
    if (!initialValue) return
    form.reset({
      name: initialValue.name,
      baseUrl: initialValue.baseUrl,
      enabled: initialValue.enabled,
      token: '',
    })
  }, [form, initialValue])

  return (
    <Form {...form}>
      <form onSubmit={form.handleSubmit(onSubmit)} className='space-y-6'>
        <Card>
          <CardHeader>
            <CardTitle>ButterBox</CardTitle>
          </CardHeader>
          <CardContent className='space-y-4'>
            <FormField
              control={form.control}
              name='name'
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Name</FormLabel>
                  <FormControl>
                    <Input placeholder='dev-box' {...field} />
                  </FormControl>
                  <FormDescription>
                    Unique within the workspace.
                  </FormDescription>
                  <FormMessage />
                </FormItem>
              )}
            />
            <FormField
              control={form.control}
              name='baseUrl'
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Base URL</FormLabel>
                  <FormControl>
                    <Input placeholder='https://box.example.com' {...field} />
                  </FormControl>
                  <FormDescription>
                    The box&apos;s HTTP address. pi-web is served at its root.
                  </FormDescription>
                  <FormMessage />
                </FormItem>
              )}
            />
            <FormField
              control={form.control}
              name='enabled'
              render={({ field }) => (
                <FormItem className='flex flex-row items-center justify-between rounded-md border p-3'>
                  <div className='space-y-0.5'>
                    <FormLabel>Enabled</FormLabel>
                    <FormDescription>
                      Disabled boxes are not offered for new pi agent bindings.
                    </FormDescription>
                  </div>
                  <FormControl>
                    <Switch
                      checked={field.value}
                      onCheckedChange={field.onChange}
                    />
                  </FormControl>
                </FormItem>
              )}
            />
            {mode === 'create' && (
              <FormField
                control={form.control}
                name='token'
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Access token (optional)</FormLabel>
                    <FormControl>
                      <Input
                        type='password'
                        autoComplete='new-password'
                        placeholder=''
                        {...field}
                      />
                    </FormControl>
                    <FormDescription>
                      Write-only: encrypted at rest and never displayed again.
                    </FormDescription>
                    <FormMessage />
                  </FormItem>
                )}
              />
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
