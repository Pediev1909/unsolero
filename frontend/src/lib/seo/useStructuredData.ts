import { useEffect } from 'react'

export function useStructuredData(id: string, data: unknown) {
  useEffect(() => {
    const selector = `script[data-structured-data="${id}"]`
    const existing = document.querySelector<HTMLScriptElement>(selector)
    if (data === null) {
      existing?.remove()
      return
    }
    const element = existing ?? document.createElement('script')
    element.type = 'application/ld+json'
    element.dataset.structuredData = id
    element.textContent = JSON.stringify(data).replace(/</g, '\\u003c')
    if (!existing) document.head.append(element)
    return () => element.remove()
  }, [data, id])
}
