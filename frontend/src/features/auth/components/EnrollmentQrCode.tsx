import { useEffect, useState } from 'react'
import QRCode from 'qrcode'

// The server already returns an otpauth:// URI, which is exactly what an
// authenticator's camera expects. Only the raw secret was shown, so enrolling
// meant transcribing 32 characters by hand — slow, and easy to get wrong in a
// way that only surfaces as a rejected code.
//
// Rendered as an SVG string rather than a canvas so it stays sharp at any size,
// and generated in the browser so the URI never travels anywhere new.
export function EnrollmentQrCode({ uri }: { uri: string }) {
  const [svg, setSvg] = useState('')
  const [failed, setFailed] = useState(false)

  useEffect(() => {
    let current = true
    QRCode.toString(uri, {
      type: 'svg',
      margin: 1,
      errorCorrectionLevel: 'M',
      color: { dark: '#14171c', light: '#ffffff' },
    })
      .then((value) => {
        if (current) setSvg(value)
      })
      .catch(() => {
        // The secret below the code is the complete fallback, so a failure
        // here costs convenience rather than the ability to enrol.
        if (current) setFailed(true)
      })
    return () => {
      current = false
    }
  }, [uri])

  if (failed || !svg) return null

  return (
    <div className="mt-4">
      <div
        aria-label="Enrollment QR code"
        className="inline-block border border-ink/15 bg-surface p-3 [&>svg]:size-44"
        dangerouslySetInnerHTML={{ __html: svg }}
        role="img"
      />
    </div>
  )
}
