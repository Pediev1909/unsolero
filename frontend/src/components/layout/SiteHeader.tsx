import { Menu } from 'lucide-react'
import { useState, type ReactNode } from 'react'

import { cn } from '../../lib/styles/cn'
import { Button } from '../ui/Button'
import { Container } from '../ui/Container'
import { SkipLink } from '../ui/SkipLink'
import { BrandMark } from './BrandMark'
import { HeaderSearch } from './HeaderSearch'
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
        <Container className="flex h-18 items-center gap-4 sm:h-20">
          <BrandMark />

          {showNavigation && (
            <div className="ml-auto hidden md:block">
              <Navigation items={primaryNavigation} />
            </div>
          )}

          {/* Search sits last in the bar and first in the reading order of
              things a lost visitor will try. It is hidden below lg only
              because the bar cannot hold it; the drawer carries it there. */}
          {showNavigation && (
            <div className="hidden lg:block">
              <HeaderSearch />
            </div>
          )}

          {actions && (
            <div className={cn('ml-auto', showNavigation && 'md:ml-0')}>
              {actions}
            </div>
          )}

          {showNavigation && (
            <Button
              aria-expanded={mobileOpen}
              aria-label="Open navigation"
              // Written as one branch or the other rather than as two
              // classes, because cn concatenates and does not merge: ml-auto
              // and ml-2 would both survive and the winner would be decided
              // by stylesheet order.
              className={
                actions ? 'ml-2 md:hidden' : 'ml-auto md:hidden'
              }
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
