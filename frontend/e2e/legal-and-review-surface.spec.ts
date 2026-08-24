import AxeBuilder from '@axe-core/playwright'
import { expect, test } from '@playwright/test'

// The terms page is new and the privacy page changed. Both are pages an
// affiliate reviewer opens deliberately, so a serious accessibility fault or a
// missing heading on either is worth catching before it ships.
for (const path of ['/terms', '/privacy']) {
  test(`${path} renders a single h1 and no serious accessibility violations`, async ({
    page,
  }) => {
    await page.goto(path)
    await expect(page.locator('h1')).toHaveCount(1)
    await expect(page.locator('main')).toBeVisible()

    const result = await new AxeBuilder({ page })
      .withTags(['wcag2a', 'wcag2aa', 'wcag21a', 'wcag21aa'])
      .analyze()
    const material = result.violations.filter(
      (violation) =>
        violation.impact === 'critical' || violation.impact === 'serious',
    )
    expect(material).toEqual([])
  })
}

test('the footer reaches terms, privacy, disclosure and a contact address', async ({
  page,
}) => {
  await page.goto('/')
  const footer = page.locator('footer')
  await expect(footer.getByRole('link', { name: 'Terms' })).toHaveAttribute(
    'href',
    '/terms',
  )
  await expect(footer.getByRole('link', { name: 'Privacy' })).toHaveAttribute(
    'href',
    '/privacy',
  )
  await expect(
    footer.getByRole('link', { name: 'Affiliate disclosure' }),
  ).toHaveAttribute('href', '/affiliate-disclosure')
  await expect(footer.getByRole('link', { name: 'Contact' })).toHaveAttribute(
    'href',
    'mailto:hello@unsolero.com',
  )
})

// The component gallery ships only in development. In a production build the
// route must not resolve at all, because the gallery carries invented products
// with placeholder prices. Asserting the not-found page is what makes this a
// real check: an earlier version asserted the gallery's heading was absent,
// which was true whether or not the page rendered.
test('the design system gallery is absent from the production build', async ({
  page,
}) => {
  await page.goto('/design-system')
  await expect(page.getByText('404 / Page not found')).toBeVisible()
})
