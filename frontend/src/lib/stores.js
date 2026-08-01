import { writable } from 'svelte/store'

// --- Core stores ---
export const messages = writable([])
export const streamingContent = writable('')
export const isStreaming = writable(false)
export const activeTab = writable('chat')
export const conversations = writable([])
export const activeConversation = writable('default')
export const error = writable(null)
export const personas = writable([])
export const activePersona = writable('default')

// --- Toast notification system ---
export const toasts = writable([])
let toastId = 0

export function addToast(message, type = 'error', duration = 5000) {
  const id = ++toastId
  toasts.update((t) => [...t, { id, message, type }])
  if (duration > 0) {
    setTimeout(() => {
      toasts.update((t) => t.filter((toast) => toast.id !== id))
    }, duration)
  }
  return id
}

export function removeToast(id) {
  toasts.update((t) => t.filter((toast) => toast.id !== id))
}

// --- Dark mode ---
function getInitialDarkMode() {
  if (typeof localStorage !== 'undefined') {
    const stored = localStorage.getItem('remy-dark-mode')
    if (stored !== null) {
      return stored === 'true'
    }
  }
  if (typeof window !== 'undefined' && window.matchMedia) {
    return window.matchMedia('(prefers-color-scheme: dark)').matches
  }
  return false
}

export const darkMode = writable(getInitialDarkMode())

// Persist dark mode preference
darkMode.subscribe((value) => {
  if (typeof localStorage !== 'undefined') {
    localStorage.setItem('remy-dark-mode', String(value))
  }
  if (typeof document !== 'undefined') {
    document.documentElement.classList.toggle('dark', value)
  }
})

// --- Memory Explorer stores ---
export const facts = writable([])
export const episodes = writable([])
export const entities = writable([])
export const relationships = writable([])
export const scratchpad = writable('')
export const searchResults = writable(null)
export const searchType = writable('fulltext')
export const memorySubTab = writable('facts')

// --- Stage 10 stores ---
export const tasks = writable([])
export const activityLog = writable([])
export const config = writable(null)
