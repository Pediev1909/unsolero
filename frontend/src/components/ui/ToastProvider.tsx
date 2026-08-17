import {
  useCallback,
  useEffect,
  useMemo,
  useRef,
  useState,
  type PropsWithChildren,
} from 'react'

import { Toast } from './Toast'
import { ToastContext } from './toastContext'
import type { ToastInput, ToastRecord } from './toastTypes'

const defaultDuration = 5000

export function ToastProvider({ children }: PropsWithChildren) {
  const [toasts, setToasts] = useState<ToastRecord[]>([])
  const nextId = useRef(1)
  const timers = useRef(new Map<number, ReturnType<typeof setTimeout>>())

  const dismissToast = useCallback((id: number) => {
    setToasts((current) => current.filter((toast) => toast.id !== id))
    const timer = timers.current.get(id)
    if (timer) clearTimeout(timer)
    timers.current.delete(id)
  }, [])

  const showToast = useCallback(
    (input: ToastInput) => {
      const id = nextId.current++
      setToasts((current) => [...current, { ...input, id }].slice(-3))
      const duration = input.duration ?? defaultDuration
      if (duration > 0) {
        timers.current.set(
          id,
          setTimeout(() => dismissToast(id), duration),
        )
      }
      return id
    },
    [dismissToast],
  )

  useEffect(
    () => () => {
      for (const timer of timers.current.values()) clearTimeout(timer)
    },
    [],
  )

  const value = useMemo(
    () => ({ showToast, dismissToast }),
    [dismissToast, showToast],
  )

  return (
    <ToastContext.Provider value={value}>
      {children}
      <div
        aria-atomic="false"
        className="pointer-events-none fixed inset-x-4 bottom-4 z-[100] ml-auto flex max-w-sm flex-col gap-3 sm:inset-x-auto sm:bottom-6 sm:right-6 sm:w-full"
        aria-live="polite"
      >
        {toasts.map((toast) => (
          <Toast key={toast.id} onDismiss={dismissToast} toast={toast} />
        ))}
      </div>
    </ToastContext.Provider>
  )
}
