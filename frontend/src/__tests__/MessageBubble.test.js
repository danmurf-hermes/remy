import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/svelte'
import MessageBubble from '../lib/MessageBubble.svelte'

describe('MessageBubble', () => {
  it('renders a user message', () => {
    const message = {
      id: '1',
      role: 'user',
      content: 'Hello',
      timestamp: 1700000000000,
      interface: 'gui',
    }
    render(MessageBubble, { props: { message } })
    expect(screen.getByText('Hello')).toBeTruthy()
  })

  it('renders an agent message', () => {
    const message = {
      id: '2',
      role: 'assistant',
      content: 'Hi there!',
      timestamp: 1700000000000,
      interface: 'gui',
    }
    render(MessageBubble, { props: { message } })
    expect(screen.getByText('Hi there!')).toBeTruthy()
    expect(screen.getByText('R')).toBeTruthy()
  })

  it('shows telegram indicator for telegram messages', () => {
    const message = {
      id: '3',
      role: 'user',
      content: 'Hello',
      timestamp: 1700000000000,
      interface: 'telegram',
    }
    render(MessageBubble, { props: { message } })
    expect(screen.getByTitle('Sent via Telegram')).toBeTruthy()
  })
})
