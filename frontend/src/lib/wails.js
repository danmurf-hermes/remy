const isWails = typeof window !== 'undefined' && window.runtime !== undefined

function getRuntime() {
  if (isWails) {
    return window.runtime
  }
  return null
}

export async function sendMessage(text) {
  if (isWails) {
    return window.go.main.App.SendMessage(text)
  }
  return {
    id: 'mock',
    role: 'assistant',
    content: 'Mock response: ' + text,
    timestamp: Date.now(),
    interface: 'gui',
  }
}

export async function sendMessageStream(text) {
  if (isWails) {
    return window.go.main.App.SendMessageStream(text)
  }
  return null
}

export async function getHistory(limit, offset) {
  if (isWails) {
    return window.go.main.App.GetHistory(limit, offset)
  }
  return []
}

export async function getConversations() {
  if (isWails) {
    return window.go.main.App.GetConversations()
  }
  return [{ id: 'default', name: 'default', last_msg: 'No messages yet', timestamp: Date.now() }]
}

export async function getPersonas() {
  if (isWails) {
    return window.go.main.App.GetPersonas()
  }
  return [{ name: 'default', description: 'Default persona', is_active: true }]
}

export async function switchPersona(name) {
  if (isWails) {
    return window.go.main.App.SwitchPersona(name)
  }
  return null
}

export async function getActivePersona() {
  if (isWails) {
    return window.go.main.App.GetActivePersona()
  }
  return 'default'
}

export function onStreamChunk(callback) {
  const runtime = getRuntime()
  if (runtime) {
    runtime.EventsOn('stream:chunk', callback)
  }
}

export function onStreamDone(callback) {
  const runtime = getRuntime()
  if (runtime) {
    runtime.EventsOn('stream:done', callback)
  }
}

export function onStreamError(callback) {
  const runtime = getRuntime()
  if (runtime) {
    runtime.EventsOn('stream:error', callback)
  }
}
