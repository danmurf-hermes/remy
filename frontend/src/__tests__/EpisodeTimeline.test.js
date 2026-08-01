import { describe, it, expect, afterEach, vi } from 'vitest'
import { render, cleanup } from '@testing-library/svelte'
import EpisodeTimeline from '../lib/EpisodeTimeline.svelte'

afterEach(cleanup)

const mockEpisodes = vi.hoisted(() => [
  { id: '1', summary: 'Discussed project architecture', start_time: Date.now() - 86400000, end_time: Date.now() - 82800000, message_ids: 'a,b,c', importance: 0.8, topics: 'architecture,planning' },
  { id: '2', summary: 'User asked about Go concurrency', start_time: Date.now() - 172800000, end_time: Date.now() - 171000000, message_ids: 'd,e', importance: 0.5, topics: 'go,concurrency' },
])

vi.mock('../lib/wails.js', () => ({
  getEpisodes: vi.fn(() => Promise.resolve(mockEpisodes)),
}))

describe('EpisodeTimeline', () => {
  it('renders without crashing', () => {
    render(EpisodeTimeline)
    expect(document.body.querySelector('.episode-timeline')).toBeTruthy()
  })
})
