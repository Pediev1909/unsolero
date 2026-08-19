import { CategoriesSection } from '../components/home/CategoriesSection'
import { ComparisonSection } from '../components/home/ComparisonSection'
import { ExampleSetupSection } from '../components/home/ExampleSetupSection'
import { FeaturedProductsSection } from '../components/home/FeaturedProductsSection'
import { FinalCtaSection } from '../components/home/FinalCtaSection'
import { Hero } from '../components/home/Hero'
import { MethodSection } from '../components/home/MethodSection'
import { PersonalizationSection } from '../components/home/PersonalizationSection'
import { PrinciplesSection } from '../components/home/PrinciplesSection'
import { SiteFooter } from '../components/layout/SiteFooter'
import { SiteHeader } from '../components/layout/SiteHeader'
import { usePageMetadata } from '../lib/seo/usePageMetadata'

export function HomePage() {
  // The server renders this page's metadata into the shell, so a cold visit
  // was already correct. A visit that arrives from another page of the site
  // was not: React Router replaces the view without touching the document, so
  // the tab kept whichever title the previous page had set.
  usePageMetadata({
    title: 'UNSOLERO — Build the right software stack',
    description:
      'Tell us what your business does, what you already run, and what you can spend. We work out what you actually need.',
  })

  return (
    <>
      <SiteHeader />
      <main id="main-content">
        <Hero />
        <MethodSection />
        <ExampleSetupSection />
        <CategoriesSection />
        <ComparisonSection />
        <PersonalizationSection />
        <FeaturedProductsSection />
        <PrinciplesSection />
        <FinalCtaSection />
      </main>
      <SiteFooter />
    </>
  )
}
