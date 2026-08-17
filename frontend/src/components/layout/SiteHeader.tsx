import { Menu } from 'lucide-react'
import { useState, type ReactNode } from 'react'

import { cn } from '../../lib/styles/cn'
import { Button } from '../ui/Button'
import { Container } from '../ui/Container'
import { SkipLink } from '../ui/SkipLink'
import { BrandMark } from './BrandMark'
import { MobileNavigation } from './MobileNavigation'
import { Navigation } from './Navigation'
import { primaryNavigation } from './navigationItems'

interface SiteHeaderProps {
  position?: 'overlay' | 'static' | 'sticky'
  showNavigation?: boolean
  actions?: ReactNode
  className?: string
}

const positions = {
  overlay: 'absolute inset-x-0 top-0',
  static: 'relative',
  sticky: 'sticky inset-x-0 top-0 bg-canvas/95 backdrop-blur-sm',
}

export function SiteHeader({
  position = 'overlay',
  showNavigation = true,
  actions,
  className,
}: SiteHeaderProps) {
  const [mobileOpen, setMobileOpen] = useState(false)

  return (
    <>
      <SkipLink />
      <header
        className={cn(
          'z-40 border-b border-ink/10',
          positions[position],
          className,
        )}
      >
        <Container className="flex h-18 items-center justify-between sm:h-20">
          <BrandMark />

          {showNavigation && (
            <div className="hidden md:block">
              <Navigation items={primaryNavigation} />
            </div>
          )}

          {actions && <div className="ml-auto md:ml-0">{actions}</div>}

          {showNavigation && (
            <Button
              aria-expanded={mobileOpen}
              aria-label="Open navigation"
              className={cn('md:hidden', Boolean(actions) && 'ml-2')}
              onClick={() => setMobileOpen(true)}
              size="sm"
              variant="quiet"
            >
              <Menu aria-hidden="true" size={20} />
              <span className="hidden xs:inline">Menu</span>
            </Button>
          )}
        </Container>
      </header>
      {showNavigation && (
        <MobileNavigation
          items={primaryNavigation}
          onOpenChange={setMobileOpen}
          open={mobileOpen}
        />
      )}
    </>
  )
}
