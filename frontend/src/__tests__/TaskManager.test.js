import { describe, it, expect, afterEach, vi } from 'vitest'
import { render, screen, cleanup } from '@testing-library/svelte'
import TaskManager from '../lib/TaskManager.svelte'

vi.mock('../lib/wails.js', () => ({
  getTasks: vi.fn().mockResolvedValue([]),
  createTask: vi.fn().mockResolvedValue({ id: 'new-task' }),
  cancelTask: vi.fn().mockResolvedValue(null),
}))

afterEach(cleanup)

describe('TaskManager', () => {
  it('renders the task manager header', () => {
    render(TaskManager)
    expect(screen.getByText('Tasks & Schedule')).toBeTruthy()
  })

  it('shows a new reminder button', () => {
    render(TaskManager)
    expect(screen.getByText('+ New Reminder')).toBeTruthy()
  })

  it('renders without crashing', () => {
    render(TaskManager)
    expect(document.body.querySelector('.task-manager')).toBeTruthy()
  })
})
