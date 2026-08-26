import { useEffect, useRef, useState } from 'react'
import QRCode from 'qrcode'

// The server already returns an otpauth:// URI, which is exactly what an
// authenticator's camera expects. Only the raw secret was shown, so enrolling
// meant transcribing 32 characters by hand — slow, and easy to get wrong in a
// way that only surfaces as a rejected code.
//
// Generated directly into a local canvas so the URI never travels anywhere
// new and no generated markup has to pass through an unsafe HTML sink.
export function EnrollmentQrCode({ uri }: { uri: string }) {
  const canvas = useRef<HTMLCanvasElement>(null)
  const [renderedUri, setRenderedUri] = useState<string | null>(null)
  const [failedUri, setFailedUri] = useState<string | null>(null)
  const ready = renderedUri === uri
  const failed = failedUri === uri

  useEffect(() => {
    let current = true
    if (!canvas.current) return
    QRCode.toCanvas(canvas.current, uri, {
      width: 176,
      margin: 1,
      errorCorrectionLevel: 'M',
      color: { dark: '#14171c', light: '#ffffff' },
    })
      .then(() => {
        if (current) {
          setFailedUri(null)
          setRenderedUri(uri)
        }
      })
      .catch(() => {
        // The secret below the code is the complete fallback, so a failure
        // here costs convenience rather than the ability to enrol.
        if (current) setFailedUri(uri)
      })
    return () => {
      current = false
    }
  }, [uri])

  return (
    <div className={`mt-4 ${ready && !failed ? '' : 'invisible'}`}>
      <canvas
        aria-label="Enrollment QR code"
        className="size-44 border border-ink/15 bg-surface p-3"
        ref={canvas}
        role="img"
      />
    </div>
  )
}
