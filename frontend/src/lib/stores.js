import { writable } from 'svelte/store'

export const messages = writable([])
export const streamingContent = writable('')
export const isStreaming = writable(false)
export const activeTab = writable('chat')
export const conversations = writable([])
export const activeConversation = writable('default')
export const error = writable(null)
export const personas = writable([])
export const activePersona = writable('default')
