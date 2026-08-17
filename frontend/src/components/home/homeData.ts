import type { ProductCardData } from '../product'

export interface SetupItem {
  name: string
  reason: string
  priceMinor: number
}

export interface ComparisonProduct {
  name: string
  shortName: string
  priceMinor: number
  maximumWeight: string
  beginnerScore: number
  apartmentScore: number
  verdict: string
  recommended?: boolean
}

export const exampleSetup = {
  profile: {
    goal: 'Build muscle',
    space: 'Small apartment',
    experience: 'Beginner',
    budgetMinor: 70000,
    owned: 'Pull-up bar',
  },
  items: [
    {
      name: 'Demo Civic Select 16 Dumbbell Pair',
      reason: 'Broad exercise range without taking over the room.',
      priceMinor: 19900,
    },
    {
      name: 'Demo Oak & Iron Foldaway Flat Bench',
      reason: 'Adds supported pressing and folds away after training.',
      priceMinor: 14900,
    },
    {
      name: 'Demo Kinetic House Starter Band Set',
      reason: 'Low-cost accessory work, warm-ups, and progression.',
      priceMinor: 3900,
    },
  ] satisfies SetupItem[],
  totalMinor: 38700,
  remainingMinor: 31300,
  rejected:
    'A rack, barbell, and plates were left out: they duplicate useful movement patterns at this stage and consume most of the available space and budget.',
  upgrade:
    'Add a heavier adjustable kettlebell when the dumbbells stop challenging lower-body movements.',
}

export const comparisonProducts: ComparisonProduct[] = [
  {
    name: 'Demo Civic Select 16 Dumbbell Pair',
    shortName: 'Civic Select 16',
    priceMinor: 19900,
    maximumWeight: '16 kg / hand',
    beginnerScore: 97,
    apartmentScore: 96,
    verdict: 'Best fit for this brief',
    recommended: true,
  },
  {
    name: 'Demo QuietForm Dial 20 Dumbbell Pair',
    shortName: 'QuietForm Dial 20',
    priceMinor: 27900,
    maximumWeight: '20 kg / hand',
    beginnerScore: 93,
    apartmentScore: 95,
    verdict: 'More range, higher cost',
  },
  {
    name: 'Demo Northline Nest 24 Dumbbell Pair',
    shortName: 'Northline Nest 24',
    priceMinor: 32900,
    maximumWeight: '24 kg / hand',
    beginnerScore: 91,
    apartmentScore: 91,
    verdict: 'Better long-term headroom',
  },
]

export const featuredProducts: ProductCardData[] = [
  {
    id: 'demo-civic-select-16-pair',
    href: '/products/demo-civic-select-16-pair',
    name: 'Demo Civic Select 16 Dumbbell Pair',
    brand: 'Demo Civic Strength',
    category: 'Adjustable dumbbells',
    priceMinor: 19900,
    currency: 'USD',
    image: {
      src: '/images/demo-adjustable-dumbbells.webp',
      alt: 'Illustrative studio image of a fictional adjustable dumbbell pair',
    },
    badge: { label: 'Demo product', variant: 'neutral' },
  },
  {
    id: 'demo-oak-iron-foldaway-flat-bench',
    href: '/products/demo-oak-iron-foldaway-flat-bench',
    name: 'Demo Oak & Iron Foldaway Flat Bench',
    brand: 'Demo Oak & Iron',
    category: 'Benches',
    priceMinor: 14900,
    currency: 'USD',
    image: {
      src: '/images/demo-foldaway-bench.webp',
      alt: 'Illustrative studio image of a fictional flat workout bench',
    },
    badge: { label: 'Demo product', variant: 'neutral' },
  },
  {
    id: 'demo-range-lab-adjustable-20-kettlebell',
    href: '/products/demo-range-lab-adjustable-20-kettlebell',
    name: 'Demo Range Lab Adjustable 20 Kettlebell',
    brand: 'Demo Range Lab',
    category: 'Kettlebells',
    priceMinor: 15900,
    currency: 'USD',
    image: {
      src: '/images/demo-adjustable-kettlebell.webp',
      alt: 'Illustrative studio image of a fictional adjustable kettlebell',
    },
    badge: { label: 'Demo product', variant: 'neutral' },
  },
  {
    id: 'demo-kinetic-house-starter-band-set',
    href: '/products/demo-kinetic-house-starter-band-set',
    name: 'Demo Kinetic House Starter Band Set',
    brand: 'Demo Kinetic House',
    category: 'Resistance bands',
    priceMinor: 3900,
    currency: 'USD',
    image: {
      src: '/images/demo-resistance-bands.webp',
      alt: 'Illustrative studio image of a fictional resistance band set',
    },
    badge: { label: 'Demo product', variant: 'neutral' },
  },
]
