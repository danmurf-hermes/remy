import { describe, it, expect, afterEach } from 'vitest'
import { render, screen, cleanup } from '@testing-library/svelte'
import App from '../App.svelte'

afterEach(cleanup)

describe('App', () => {
  it('renders the sidebar', () => {
    render(App)
    expect(screen.getByTitle('Chat')).toBeTruthy()
    expect(screen.getByTitle('Memory')).toBeTruthy()
    expect(screen.getByTitle('Tasks')).toBeTruthy()
    expect(screen.getByTitle('Personas')).toBeTruthy()
    expect(screen.getByTitle('Activity')).toBeTruthy()
    expect(screen.getByTitle('Settings')).toBeTruthy()
  })

  it('shows chat view by default', () => {
    render(App)
    expect(screen.getByPlaceholderText('Message Remy…')).toBeTruthy()
  })
})
