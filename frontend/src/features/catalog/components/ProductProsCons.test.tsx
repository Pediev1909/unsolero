import { render, screen, within } from '@testing-library/react'
import { describe, expect, it } from 'vitest'

import { ProductProsCons } from './ProductProsCons'
import { productDetailFixture } from './productDetailFixture'

describe('ProductProsCons', () => {
  it('sets strengths beside trade-offs with their scores, and keeps the scores disclaimer', () => {
    const product = productDetailFixture()
    render(
      <ProductProsCons
        strengths={product.strengths}
        useCases={product.use_cases}
        weaknesses={product.weaknesses}
      />,
    )

    const strengths = screen
      .getByRole('heading', { name: 'Strengths' })
      .closest('div')
    const tradeOffs = screen
      .getByRole('heading', { name: 'Trade-offs' })
      .closest('div')
    if (!strengths || !tradeOffs) throw new Error('columns not rendered')

    expect(within(strengths).getByText('Template library')).toBeInTheDocument()
    expect(within(strengths).getByText('86/100')).toBeInTheDocument()
    expect(within(tradeOffs).getByText('Price at scale')).toBeInTheDocument()
    expect(within(tradeOffs).getByText('74/100')).toBeInTheDocument()
    expect(
      screen.getByRole('heading', { name: 'Best use cases' }),
    ).toBeInTheDocument()
    expect(screen.getByText('Small-list newsletters')).toBeInTheDocument()
    expect(
      screen.getByText(
        /Scores are derived from structured catalog facts\. They are not customer ratings or reviews\./,
      ),
    ).toBeInTheDocument()
  })

  it('keeps the honest empty lines when a list has nothing above the threshold', () => {
    render(<ProductProsCons strengths={[]} useCases={[]} weaknesses={[]} />)

    expect(screen.getByText(/No standout strength crossed/)).toBeInTheDocument()
    expect(screen.getByText(/No material weakness crossed/)).toBeInTheDocument()
    expect(screen.getByText(/No use case crossed/)).toBeInTheDocument()
  })
})
