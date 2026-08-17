export function formatMinorCurrency(
  amountMinor: number,
  currency: string,
  locale?: string,
): string {
  if (!Number.isSafeInteger(amountMinor)) {
    throw new Error('Money amount must be a safe integer in minor units.')
  }
  const formatter = new Intl.NumberFormat(locale, {
    style: 'currency',
    currency,
  })
  const fractionDigits = formatter.resolvedOptions().maximumFractionDigits ?? 2
  return formatter.format(amountMinor / 10 ** fractionDigits)
}
