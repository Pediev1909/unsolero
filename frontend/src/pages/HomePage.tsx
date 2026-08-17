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

export function HomePage() {
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
