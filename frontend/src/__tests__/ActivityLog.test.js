import { describe, it, expect, afterEach, vi } from 'vitest'
import { render, screen, cleanup } from '@testing-library/svelte'
import ActivityLog from '../lib/ActivityLog.svelte'

vi.mock('../lib/wails.js', () => ({
  getActivityLog: vi.fn().mockResolvedValue([]),
}))

afterEach(cleanup)

describe('ActivityLog', () => {
  it('renders the activity log header', () => {
    render(ActivityLog)
    expect(screen.getByText('Activity Log')).toBeTruthy()
  })

  it('shows filter chips', () => {
    render(ActivityLog)
    expect(screen.getByText('All')).toBeTruthy()
    expect(screen.getByText('Messages')).toBeTruthy()
    expect(screen.getByText('Errors')).toBeTruthy()
  })

  it('shows refresh button', () => {
    render(ActivityLog)
    expect(screen.getByText('🔄 Refresh')).toBeTruthy()
  })

  it('renders without crashing', () => {
    render(ActivityLog)
    expect(document.body.querySelector('.activity-log')).toBeTruthy()
  })
})
