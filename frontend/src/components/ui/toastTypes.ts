export type ToastVariant = 'neutral' | 'success' | 'error'

export interface ToastInput {
  title: string
  description?: string
  variant?: ToastVariant
  duration?: number
}

export interface ToastRecord extends ToastInput {
  id: number
}

export interface ToastContextValue {
  showToast: (toast: ToastInput) => number
  dismissToast: (id: number) => void
}
