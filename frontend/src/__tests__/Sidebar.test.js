import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/svelte'
import Sidebar from '../lib/Sidebar.svelte'

describe('Sidebar', () => {
  it('renders all tab buttons', () => {
    render(Sidebar)
    expect(screen.getByTitle('Chat')).toBeTruthy()
    expect(screen.getByTitle('Memory')).toBeTruthy()
    expect(screen.getByTitle('Tasks')).toBeTruthy()
    expect(screen.getByTitle('Personas')).toBeTruthy()
    expect(screen.getByTitle('Activity')).toBeTruthy()
    expect(screen.getByTitle('Settings')).toBeTruthy()
  })
})
