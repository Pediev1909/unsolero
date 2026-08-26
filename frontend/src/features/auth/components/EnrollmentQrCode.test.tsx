import { render, screen, waitFor } from '@testing-library/react'
import QRCode from 'qrcode'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import { EnrollmentQrCode } from './EnrollmentQrCode'

vi.mock('qrcode', () => ({
  default: { toCanvas: vi.fn() },
}))

const toCanvas = vi.mocked(QRCode.toCanvas)

describe('EnrollmentQrCode', () => {
  beforeEach(() => {
    toCanvas.mockReset()
  })

  it('renders the provisioning URI directly into a canvas', async () => {
    toCanvas.mockResolvedValue(undefined)

    render(<EnrollmentQrCode uri="otpauth://totp/UNSOLERO:test" />)

    const canvas = screen.getByRole('img', { name: 'Enrollment QR code' })
    await waitFor(() =>
      expect(canvas.parentElement).not.toHaveClass('invisible'),
    )
    expect(toCanvas).toHaveBeenCalledWith(
      canvas,
      'otpauth://totp/UNSOLERO:test',
      expect.objectContaining({ width: 176 }),
    )
  })

  it('keeps the incomplete QR code hidden when generation fails', async () => {
    toCanvas.mockRejectedValue(new Error('canvas unavailable'))

    render(<EnrollmentQrCode uri="otpauth://totp/UNSOLERO:test" />)

    const canvas = screen.getByRole('img', { name: 'Enrollment QR code' })
    await waitFor(() => expect(toCanvas).toHaveBeenCalled())
    expect(canvas.parentElement).toHaveClass('invisible')
  })
})
