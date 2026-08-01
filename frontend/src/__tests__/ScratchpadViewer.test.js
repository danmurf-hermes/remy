import { describe, it, expect, afterEach, vi } from 'vitest'
import { render, cleanup } from '@testing-library/svelte'
import ScratchpadViewer from '../lib/ScratchpadViewer.svelte'

afterEach(cleanup)

vi.mock('../lib/wails.js', () => ({
  getScratchpad: vi.fn(() => Promise.resolve('Test scratchpad content')),
  updateScratchpad: vi.fn(() => Promise.resolve(null)),
}))

describe('ScratchpadViewer', () => {
  it('renders without crashing', () => {
    render(ScratchpadViewer)
    expect(document.body.querySelector('.scratchpad-viewer')).toBeTruthy()
  })
})
