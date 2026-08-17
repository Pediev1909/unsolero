import { useSyncExternalStore } from 'react'

const localChangeEvent = 'rigmark:local-storage-change'

function subscribe(key: string, callback: () => void) {
  function handleStorage(event: StorageEvent) {
    if (event.key === key) callback()
  }
  function handleLocalChange(event: Event) {
    if ((event as CustomEvent<string>).detail === key) callback()
  }
  window.addEventListener('storage', handleStorage)
  window.addEventListener(localChangeEvent, handleLocalChange)
  return () => {
    window.removeEventListener('storage', handleStorage)
    window.removeEventListener(localChangeEvent, handleLocalChange)
  }
}

export function writeLocalDocument(key: string, value: unknown) {
  window.localStorage.setItem(key, JSON.stringify(value))
  window.dispatchEvent(new CustomEvent(localChangeEvent, { detail: key }))
}

export function useLocalDocument(key: string): string {
  return useSyncExternalStore(
    (callback) => subscribe(key, callback),
    () => window.localStorage.getItem(key) ?? '',
    () => '',
  )
}
