import { describe, it, expect, afterEach, vi } from 'vitest'
import { render, screen, cleanup } from '@testing-library/svelte'
import PersonaStudio from '../lib/PersonaStudio.svelte'

vi.mock('../lib/wails.js', () => ({
  getPersonas: vi.fn().mockResolvedValue([]),
  switchPersona: vi.fn().mockResolvedValue(null),
}))

afterEach(cleanup)

describe('PersonaStudio', () => {
  it('renders the persona studio header', () => {
    render(PersonaStudio)
    expect(screen.getByText('Persona Studio')).toBeTruthy()
  })

  it('shows new persona button', () => {
    render(PersonaStudio)
    expect(screen.getByText('+ New Persona')).toBeTruthy()
  })

  it('renders without crashing', () => {
    render(PersonaStudio)
    expect(document.body.querySelector('.persona-studio')).toBeTruthy()
  })
})
