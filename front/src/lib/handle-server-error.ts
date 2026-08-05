import { ConnectError } from '@connectrpc/connect'
import { toast } from 'sonner'

export function handleServerError(error: unknown) {
  if (import.meta.env.DEV) {
    // eslint-disable-next-line no-console
    console.log(error)
  }

  let errMsg = 'Something went wrong!'

  if (error instanceof ConnectError && error.rawMessage.length > 0) {
    errMsg = error.rawMessage
  }

  toast.error(errMsg)
}
