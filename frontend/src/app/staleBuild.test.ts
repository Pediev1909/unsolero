import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

import { isStaleBuildError, recoverFromStaleBuild } from './staleBuild'

describe('isStaleBuildError', () => {
  it('recognises the wording each browser uses for a missing chunk', () => {
    const messages = [
      'Failed to fetch dynamically imported module: https://unsolero.com/assets/CheckEmailPage-DF9btwbh.js',
      'error loading dynamically imported module',
      'Importing a module script failed.',
    ]
    for (const message of messages) {
      expect(isStaleBuildError(new Error(message))).toBe(true)
    }
  })

  it('leaves ordinary application errors to the error page', () => {
    expect(isStaleBuildError(new Error('Cannot read properties of null'))).toBe(
      false,
    )
    expect(isStaleBuildError(undefined)).toBe(false)
  })
})

describe('recoverFromStaleBuild', () => {
  const reload = vi.fn()

  beforeEach(() => {
    reload.mockClear()
    window.sessionStorage.clear()
    Object.defineProperty(window, 'location', {
      configurable: true,
      value: { ...window.location, reload },
    })
  })

  afterEach(() => {
    window.sessionStorage.clear()
  })

  it('reloads once so the browser picks up the current chunk names', () => {
    expect(recoverFromStaleBuild()).toBe(true)
    expect(reload).toHaveBeenCalledTimes(1)
  })

  it('gives up rather than looping when the reload did not help', () => {
    recoverFromStaleBuild()
    reload.mockClear()

    // A second failure inside the window means reloading is not fixing it, so
    // the visitor should get the error page instead of an endless refresh.
    expect(recoverFromStaleBuild()).toBe(false)
    expect(reload).not.toHaveBeenCalled()
  })
})
