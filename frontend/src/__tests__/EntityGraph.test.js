import { describe, it, expect, afterEach, vi } from 'vitest'
import { render, cleanup } from '@testing-library/svelte'
import EntityGraph from '../lib/EntityGraph.svelte'

afterEach(cleanup)

const mockEntities = vi.hoisted(() => [
  {
    id: '1',
    name: 'User',
    type: 'person',
    description: 'The primary user',
    created_at: Date.now() - 86400000,
  },
  {
    id: '2',
    name: 'Remy',
    type: 'ai',
    description: 'The AI assistant',
    created_at: Date.now() - 86400000,
  },
])
const mockRelationships = vi.hoisted(() => [
  {
    id: '1',
    source_entity: 'User',
    target_entity: 'Remy',
    relationship: 'uses',
    confidence: 1.0,
    created_at: Date.now() - 86400000,
  },
])

vi.mock('../lib/wails.js', () => ({
  getEntities: vi.fn(() => Promise.resolve(mockEntities)),
  getRelationships: vi.fn(() => Promise.resolve(mockRelationships)),
}))

describe('EntityGraph', () => {
  it('renders without crashing', () => {
    render(EntityGraph)
    expect(document.body.querySelector('.entity-graph')).toBeTruthy()
  })
})
