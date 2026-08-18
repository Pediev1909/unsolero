import type { AuthUser } from '../schemas'
import { AccountDataControls } from './AccountDataControls'
import { MfaSecurity } from './MfaSecurity'
import { PasswordAndEmailSecurity } from './PasswordAndEmailSecurity'
import { SessionSecurity } from './SessionSecurity'

export function SecuritySettings({ user }: { user: AuthUser }) {
  return (
    <section
      aria-labelledby="security-settings-heading"
      className="mt-20 max-w-5xl"
    >
      <div className="border-b border-ink/15 pb-5">
        <p className="eyebrow">Protection and privacy</p>
        <h2
          className="mt-3 font-display text-4xl font-medium tracking-[-0.04em]"
          id="security-settings-heading"
        >
          Security settings
        </h2>
      </div>
      <div className="divide-y divide-ink/15">
        {[
          <PasswordAndEmailSecurity key="credentials" user={user} />,
          <SessionSecurity key="sessions" />,
          <MfaSecurity key="mfa" user={user} />,
          <AccountDataControls key="data" />,
        ].map((section) => (
          <div className="py-10" key={section.key}>
            {section}
          </div>
        ))}
      </div>
    </section>
  )
}
