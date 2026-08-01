import { describe, it, expect, afterEach, vi } from 'vitest'
import { render, cleanup } from '@testing-library/svelte'
import FactList from '../lib/FactList.svelte'

afterEach(cleanup)

const mockFacts = vi.hoisted(() => [
  {
    id: '1',
    fact: 'User prefers dark mode',
    category: 'preferences',
    confidence: 0.9,
    source: 'consolidation',
    created_at: 1000,
    updated_at: 2000,
  },
  {
    id: '2',
    fact: 'User works as a software engineer',
    category: 'personal',
    confidence: 0.85,
    source: 'consolidation',
    created_at: 3000,
    updated_at: 4000,
  },
])

vi.mock('../lib/wails.js', () => ({
  getFacts: vi.fn(() => Promise.resolve(mockFacts)),
  updateFact: vi.fn(() => Promise.resolve(null)),
  deleteFact: vi.fn(() => Promise.resolve(null)),
}))

describe('FactList', () => {
  it('renders without crashing', () => {
    render(FactList)
    expect(document.body.querySelector('.fact-list')).toBeTruthy()
  })
})
