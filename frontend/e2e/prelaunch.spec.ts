import { expect, test } from '@playwright/test'
import AxeBuilder from '@axe-core/playwright'

test.beforeEach(async ({ context }) => {
  await context.clearCookies()
  await context.addInitScript(() => {
    if (window.name !== 'test-analytics-consent') {
      localStorage.setItem('unsolero:analytics-consent:analytics-v1', 'denied')
    }
  })
})

test('public homepage and catalog retain their primary decision paths', async ({
  page,
}) => {
  await page.goto('/')
  await expect(
    page.getByRole('heading', {
      level: 1,
      name: 'Build the right software stack.',
    }),
  ).toBeVisible()
  await expect(
    page.getByRole('link', { name: 'Build My Setup' }).first(),
  ).toHaveAttribute('href', '/build')
  await expect(
    page.getByRole('link', { name: 'Explore Categories' }).first(),
  ).toHaveAttribute('href', '/#categories')
  await expectNoHorizontalOverflow(page)
  await expectNoSeriousAccessibilityViolations(page)

  await page.goto('/products')
  await expect(
    page.getByRole('heading', {
      level: 1,
      name: 'Software, judged on what matters.',
    }),
  ).toBeVisible()
  await expect(
    page.getByRole('link', { name: /View details/i }).first(),
  ).toBeVisible()
  await expectNoHorizontalOverflow(page)
  await expectNoSeriousAccessibilityViolations(page)
})

test('registration, verification, login, logout, and password reset complete through the development delivery boundary', async ({
  page,
}, testInfo) => {
  test.skip(
    testInfo.project.name !== 'desktop-chrome' ||
      process.env.E2E_BASE_URL !== undefined,
    'development delivery-boundary lifecycle runs once against the local development provider',
  )
  const suffix = crypto.randomUUID()
  const email = `phase9-${suffix}@example.invalid`
  const originalPassword = `Original-${suffix}!`
  const replacementPassword = `Replacement-${suffix}!`

  await page.goto('/register')
  await page.getByLabel('Email').fill(email)
  await page.getByLabel('Password').fill(originalPassword)
  await page.getByRole('button', { name: 'Create account' }).click()
  await expect(
    page.getByRole('heading', { name: 'Verification requested' }),
  ).toBeVisible()

  const verificationToken = await developmentToken(
    page,
    email,
    'email_verification',
  )
  await page.goto(`/verify-email#${verificationToken}`)
  await expect(
    page.getByRole('heading', { name: 'Email verified' }),
  ).toBeVisible()

  await page.goto('/login')
  await page.getByLabel('Email').fill(email)
  await page.getByLabel('Password').fill(originalPassword)
  await page.getByRole('button', { name: 'Sign in' }).click()
  await expect(page).toHaveURL(/\/account$/)
  await expect(
    page.getByRole('heading', { level: 2, name: 'Security settings' }),
  ).toBeVisible()
  await page.getByRole('button', { name: 'Sign out' }).click()
  await expect(page).toHaveURL(/\/login$/)

  await page.goto('/forgot-password')
  await page.getByLabel('Email').fill(email)
  await page.getByRole('button', { name: 'Request reset' }).click()
  await expect(page.getByRole('status')).toContainText('delivery intent')
  const resetToken = await developmentToken(page, email, 'password_reset')
  await page.goto(`/reset-password#${resetToken}`)
  await page.getByLabel('New password').fill(replacementPassword)
  await page.getByRole('button', { name: 'Replace password' }).click()
  await expect(
    page.getByRole('heading', { name: 'Password replaced' }),
  ).toBeVisible()

  await page.goto('/login')
  await page.getByLabel('Email').fill(email)
  await page.getByLabel('Password').fill(replacementPassword)
  await page.getByRole('button', { name: 'Sign in' }).click()
  await expect(page).toHaveURL(/\/account$/)
})

test('recommendation flow is keyboard-addressable and produces a real deterministic result', async ({
  page,
}, testInfo) => {
  test.skip(testInfo.project.name !== 'desktop-chrome', 'full flow runs once')
  await page.goto('/build')
  await chooseWithKeyboard(
    page.getByRole('radio', { name: /Run a client services business/ }),
    page,
  )
  await activateWithKeyboard(page.getByRole('button', { name: 'Next' }), page)
  await chooseWithKeyboard(
    page.getByRole('radio', { name: /No dedicated admin/ }),
    page,
  )
  await activateWithKeyboard(page.getByRole('button', { name: 'Next' }), page)
  // Budget, then the current-stack step, are both left at their defaults.
  await activateWithKeyboard(page.getByRole('button', { name: 'Next' }), page)
  await activateWithKeyboard(page.getByRole('button', { name: 'Next' }), page)
  await chooseWithKeyboard(
    page.getByRole('checkbox', { name: 'The strongest tool per job' }),
    page,
  )
  await activateWithKeyboard(page.getByRole('button', { name: 'Next' }), page)
  await chooseWithKeyboard(
    page.getByRole('checkbox', { name: 'Best value' }),
    page,
  )
  await activateWithKeyboard(page.getByRole('button', { name: 'Next' }), page)
  await activateWithKeyboard(
    page.getByRole('button', { name: 'Build my setup' }),
    page,
  )
  await expect(
    page.getByRole('heading', { level: 1, name: 'Your Personalized Stack' }),
  ).toBeVisible()
  await expect(page.getByText('Recommendation', { exact: true })).toBeVisible()
  await expect(
    page.getByText('Products we deliberately rejected'),
  ).toBeVisible()
})

test('authentication, recovery, account, and admin boundaries are explicit', async ({
  page,
}) => {
  await page.goto('/register')
  await expect(
    page.getByRole('heading', { level: 1, name: 'Build with confidence.' }),
  ).toBeVisible()
  await expect(page.getByLabel('Email')).toBeVisible()
  await expect(page.getByLabel('Password')).toBeVisible()

  await page.goto('/login')
  await expect(
    page.getByRole('heading', { level: 2, name: 'Sign in' }),
  ).toBeVisible()
  await page.getByRole('link', { name: 'Forgot your password?' }).click()
  await expect(
    page.getByRole('heading', { level: 1, name: 'Reset your password.' }),
  ).toBeVisible()

  await page.goto('/account')
  await expect(page).toHaveURL(/\/login$/)
  await page.goto('/admin')
  await expect(page).toHaveURL(/\/login$/)
})

test('authenticated account security renders while a non-admin is denied admin access', async ({
  page,
}) => {
  await page.route('**/api/auth/me', async (route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        user: {
          id: '00000000-0000-4000-8000-000000000001',
          email: 'browser-test@example.invalid',
          roles: [],
          email_verified: true,
          mfa_enabled: false,
        },
      }),
    })
  })
  await page.route('**/api/account/setups**', async (route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        setups: [],
        page: 1,
        page_size: 100,
        total: 0,
        total_pages: 0,
      }),
    })
  })
  await page.route('**/api/account/security/sessions', async (route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({ sessions: [] }),
    })
  })
  await page.goto('/account')
  await expect(
    page.getByRole('heading', { level: 2, name: 'Security settings' }),
  ).toBeVisible()
  await page.goto('/admin')
  await expect(page).toHaveURL(/\/account$/)
})

test('analytics consent stays explicit and optional', async ({ page }) => {
  await page.goto('/')
  await page.evaluate(() => {
    window.name = 'test-analytics-consent'
    localStorage.removeItem('unsolero:analytics-consent:analytics-v1')
  })
  await page.reload()
  const preferences = page.getByRole('region', {
    name: 'Analytics preferences',
  })
  await expect(preferences).toBeVisible()
  await preferences.getByRole('button', { name: 'Decline' }).click()
  await expect(preferences).toBeHidden()
  await expect
    .poll(() =>
      page.evaluate(() =>
        localStorage.getItem('unsolero:analytics-consent:analytics-v1'),
      ),
    )
    .toBe('denied')
})

test('catalog failures provide bounded recovery instead of fabricated data', async ({
  page,
}) => {
  await page.route('**/api/catalog/products**', async (route) => {
    await route.fulfill({
      status: 429,
      contentType: 'application/json',
      body: JSON.stringify({
        error: { code: 'rate_limited', message: 'Please try again later.' },
      }),
    })
  })
  await page.goto('/products')
  await expect(page.getByText(/try again/i).first()).toBeVisible()
  await expect(page.getByRole('link', { name: /View details/i })).toHaveCount(0)
})

test('server errors remain honest and recoverable', async ({ page }) => {
  await page.route(
    '**/api/catalog/products/clickup-unlimited',
    async (route) => {
      await route.fulfill({
        status: 500,
        contentType: 'application/json',
        body: JSON.stringify({
          error: { code: 'internal_error', message: 'Request failed.' },
        }),
      })
    },
  )
  await page.goto('/products/clickup-unlimited')
  await expect(
    page.getByRole('heading', { name: 'Product unavailable' }),
  ).toBeVisible()
  await expect(page.getByRole('button', { name: /try again/i })).toBeVisible()
})

test('affiliate CTAs use the tracked backend boundary and disclose the relationship', async ({
  page,
}) => {
  await page.goto('/products/clickup-unlimited')
  const merchantLink = page.getByRole('link', { name: /View at /i }).first()
  if ((await merchantLink.count()) === 0) {
    test.skip(
      true,
      'no affiliate offer is configured yet; the boundary is covered by backend tests',
    )
  }
  await expect(merchantLink).toBeVisible()
  await expect(merchantLink).toHaveAttribute(
    'href',
    /\/api\/affiliate\/click\//,
  )
  await expect(merchantLink).toHaveAttribute('rel', /sponsored/)
  await expect(merchantLink).toHaveAttribute('target', '_blank')
  await expect(
    page.getByText(/commission never changes product ranking/i),
  ).toBeVisible()
  const path = await merchantLink.getAttribute('href')
  expect(path).not.toBeNull()
  const redirect = await page.request.get(path!, { maxRedirects: 0 })
  expect(redirect.status()).toBe(302)
  expect(redirect.headers().location).toMatch(/^https:\/\/[^/]+\.invalid\//)
})

test('core public layouts do not overflow the required viewport widths', async ({
  context,
}, testInfo) => {
  test.skip(
    testInfo.project.name !== 'desktop-chrome',
    'explicit viewport matrix runs once',
  )
  test.setTimeout(180_000)
  for (const width of [320, 375, 390, 430, 768, 1024, 1280, 1440, 1920]) {
    await test.step(`${width}px`, async () => {
      const page = await context.newPage()
      try {
        await page.setViewportSize({
          width,
          height: width < 768 ? 844 : 1000,
        })
        await page.goto('/', { waitUntil: 'domcontentloaded' })
        await expect(
          page.getByRole('heading', {
            level: 1,
            name: 'Build the right software stack.',
          }),
        ).toBeVisible()
        await expectNoHorizontalOverflow(page)
        await page.goto('/products', {
          waitUntil: 'domcontentloaded',
        })
        await expect(
          page.getByRole('heading', {
            level: 1,
            name: 'Software, judged on what matters.',
          }),
        ).toBeVisible()
        await expectNoHorizontalOverflow(page)
      } finally {
        await page.close()
      }
    })
  }
})

test('modal focus and dismissal work from the keyboard', async ({ page }) => {
  await page.goto('/design-system')
  const openModal = page.getByRole('button', { name: 'Open modal' })
  await openModal.focus()
  await page.keyboard.press('Enter')
  const dialog = page.getByRole('dialog', { name: 'Confirm tool choice' })
  await expect(dialog).toBeVisible()
  await expect(
    dialog.getByRole('button', { name: 'Close dialog' }),
  ).toBeFocused()
  await page.keyboard.press('Escape')
  await expect(dialog).toBeHidden()
})

test('interactive controls expose accessible names and keyboard focus', async ({
  page,
}) => {
  await page.goto('/login')
  const unnamed = await page
    .locator('button, input, select, textarea, a[href]')
    .evaluateAll(
      (nodes) =>
        nodes.filter((node) => {
          const element = node as HTMLElement
          const aria =
            element.getAttribute('aria-label') ??
            element.getAttribute('aria-labelledby')
          const text = element.textContent?.trim()
          const id = element.id
          const label = id
            ? document.querySelector(`label[for="${CSS.escape(id)}"]`)
            : null
          return !aria && !text && !label
        }).length,
    )
  expect(unnamed).toBe(0)

  // Tab has nowhere to land until the page is interactive. Pressing it against
  // a still-hydrating document put focus on the body and left ":focus"
  // matching nothing, which read as a keyboard-accessibility failure rather
  // than the race it was.
  await expect(page.getByLabel('Email')).toBeVisible()
  await page.keyboard.press('Tab')
  await expect(page.locator(':focus')).toBeVisible()
  await expectNoSeriousAccessibilityViolations(page)
})

async function expectNoHorizontalOverflow(
  page: import('@playwright/test').Page,
) {
  await expect
    .poll(() =>
      page.evaluate(
        () => document.documentElement.scrollWidth <= window.innerWidth + 1,
      ),
    )
    .toBe(true)
}

async function developmentToken(
  page: import('@playwright/test').Page,
  email: string,
  kind: 'email_verification' | 'password_reset',
) {
  return expect
    .poll(
      async () => {
        const response = await page.request.get(
          `/api/dev/email-deliveries?recipient=${encodeURIComponent(email)}`,
        )
        if (!response.ok()) return ''
        const payload: unknown = await response.json()
        if (!isRecord(payload) || !Array.isArray(payload.messages)) return ''
        for (let index = payload.messages.length - 1; index >= 0; index -= 1) {
          const message: unknown = payload.messages[index]
          if (
            isRecord(message) &&
            message.kind === kind &&
            typeof message.token === 'string'
          ) {
            return message.token
          }
        }
        return ''
      },
      { message: `development ${kind} token`, timeout: 10_000 },
    )
    .not.toBe('')
    .then(async () => {
      const response = await page.request.get(
        `/api/dev/email-deliveries?recipient=${encodeURIComponent(email)}`,
      )
      const payload: unknown = await response.json()
      if (!isRecord(payload) || !Array.isArray(payload.messages)) {
        throw new Error('development email response is malformed')
      }
      for (let index = payload.messages.length - 1; index >= 0; index -= 1) {
        const message: unknown = payload.messages[index]
        if (
          isRecord(message) &&
          message.kind === kind &&
          typeof message.token === 'string'
        ) {
          return message.token
        }
      }
      throw new Error(`development ${kind} token is unavailable`)
    })
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null
}

async function expectNoSeriousAccessibilityViolations(
  page: import('@playwright/test').Page,
) {
  const result = await new AxeBuilder({ page })
    .withTags(['wcag2a', 'wcag2aa', 'wcag21a', 'wcag21aa'])
    .analyze()
  const materialViolations = result.violations.filter(
    (violation) =>
      violation.impact === 'critical' || violation.impact === 'serious',
  )
  expect(materialViolations).toEqual([])
}

async function chooseWithKeyboard(
  control: import('@playwright/test').Locator,
  page: import('@playwright/test').Page,
) {
  await control.focus()
  await page.keyboard.press('Space')
  await expect(control).toBeChecked()
}

async function activateWithKeyboard(
  control: import('@playwright/test').Locator,
  page: import('@playwright/test').Page,
) {
  await control.focus()
  await page.keyboard.press('Enter')
}
