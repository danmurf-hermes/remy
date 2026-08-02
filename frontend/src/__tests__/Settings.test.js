import { describe, it, expect, afterEach, vi } from 'vitest'
import { render, screen, cleanup } from '@testing-library/svelte'
import Settings from '../lib/Settings.svelte'

vi.mock('../lib/wails.js', () => ({
  getConfig: vi.fn().mockResolvedValue({
    providers: {
      ollama: {
        endpoint: 'http://localhost:11434/v1',
        chat_model: 'llama3.1:8b',
        embedding_model: 'nomic-embed-text',
        parameters: { temperature: 0.7, max_tokens: 4096 },
      },
    },
    default_provider: 'ollama',
    memory: {
      db_path: '~/.remy/memory.db',
      working_memory_turns: 20,
      quick_consolidation_delay_ms: 300000,
      deep_consolidation_delay_ms: 1800000,
    },
    persona: { active: 'default', directory: '~/.remy/personas/' },
    interfaces: {
      telegram: { enabled: false, bot_token: '', allowed_users: [] },
    },
  }),
  updateConfig: vi.fn().mockResolvedValue(null),
}))

afterEach(cleanup)

describe('Settings', () => {
  it('renders the settings header', () => {
    render(Settings)
    expect(screen.getByRole('heading', { name: 'Providers' })).toBeTruthy()
  })

  it('shows save button', () => {
    render(Settings)
    expect(screen.getByText('Save')).toBeTruthy()
  })

  it('renders without crashing', () => {
    render(Settings)
    expect(document.body.querySelector('.settings')).toBeTruthy()
  })
})
