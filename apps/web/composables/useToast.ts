export type ToastTone = 'success' | 'error' | 'info' | 'warning'

export interface ToastMessage {
  id: number
  tone: ToastTone
  title: string
  message?: string
}

const DEFAULT_TIMEOUT_MS = 4600

export function useToast() {
  const toasts = useState<ToastMessage[]>('toast.messages', () => [])

  function remove(id: number) {
    toasts.value = toasts.value.filter((toast) => toast.id !== id)
  }

  function push(tone: ToastTone, title: string, message?: string, timeoutMs = DEFAULT_TIMEOUT_MS) {
    const id = Date.now() + Math.floor(Math.random() * 1000)
    toasts.value = [...toasts.value, { id, tone, title, message }]

    if (import.meta.client && timeoutMs > 0) {
      window.setTimeout(() => remove(id), timeoutMs)
    }

    return id
  }

  return {
    toasts,
    remove,
    success: (title: string, message?: string) => push('success', title, message),
    error: (title: string, message?: string) => push('error', title, message, 6400),
    info: (title: string, message?: string) => push('info', title, message),
    warning: (title: string, message?: string) => push('warning', title, message, 5600)
  }
}
