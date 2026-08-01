import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/svelte'
import App from '../App.svelte'

describe('App', () => {
  it('renders the greeting', () => {
    render(App)
    expect(screen.getByText('Hello Remy')).toBeTruthy()
  })
})
