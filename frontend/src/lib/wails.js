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

// --- Memory Explorer bindings ---

export async function getFacts(category) {
  if (isWails) {
    return window.go.main.App.GetFacts(category)
  }
  // Mock data for testing
  const mockFacts = [
    { id: '1', fact: 'User prefers dark mode', category: 'preferences', confidence: 0.9, source: 'consolidation', created_at: Date.now() - 86400000, updated_at: Date.now() - 3600000 },
    { id: '2', fact: 'User works as a software engineer', category: 'personal', confidence: 0.85, source: 'consolidation', created_at: Date.now() - 172800000, updated_at: Date.now() - 7200000 },
    { id: '3', fact: 'User likes Python and Go', category: 'preferences', confidence: 0.75, source: 'consolidation', created_at: Date.now() - 259200000, updated_at: Date.now() - 10800000 },
  ]
  if (category) {
    return mockFacts.filter(f => f.category === category)
  }
  return mockFacts
}

export async function getFact(id) {
  if (isWails) {
    return window.go.main.App.GetFact(id)
  }
  return null
}

export async function updateFact(id, fact, category, confidence) {
  if (isWails) {
    return window.go.main.App.UpdateFact(id, fact, category, confidence)
  }
  return null
}

export async function deleteFact(id) {
  if (isWails) {
    return window.go.main.App.DeleteFact(id)
  }
  return null
}

export async function getEpisodes(limit, offset) {
  if (isWails) {
    return window.go.main.App.GetEpisodes(limit, offset)
  }
  // Mock data for testing
  return [
    { id: '1', summary: 'Discussed project architecture and decided on microservices approach', start_time: Date.now() - 86400000, end_time: Date.now() - 82800000, message_ids: 'a,b,c', importance: 0.8, topics: 'architecture,planning' },
    { id: '2', summary: 'User asked about Go concurrency patterns', start_time: Date.now() - 172800000, end_time: Date.now() - 171000000, message_ids: 'd,e', importance: 0.5, topics: 'go,concurrency' },
    { id: '3', summary: 'Debugged database migration issue together', start_time: Date.now() - 259200000, end_time: Date.now() - 257400000, message_ids: 'f,g,h', importance: 0.7, topics: 'database,migration,debugging' },
  ]
}

export async function getEpisode(id) {
  if (isWails) {
    return window.go.main.App.GetEpisode(id)
  }
  return null
}

export async function getEntities() {
  if (isWails) {
    return window.go.main.App.GetEntities()
  }
  // Mock data for testing
  return [
    { id: '1', name: 'User', type: 'person', description: 'The primary user of Remy', created_at: Date.now() - 86400000 },
    { id: '2', name: 'Remy', type: 'ai', description: 'The personal AI assistant', created_at: Date.now() - 86400000 },
    { id: '3', name: 'Go', type: 'language', description: 'Go programming language', created_at: Date.now() - 172800000 },
    { id: '4', name: 'SQLite', type: 'technology', description: 'SQLite database engine', created_at: Date.now() - 172800000 },
  ]
}

export async function getRelationships() {
  if (isWails) {
    return window.go.main.App.GetRelationships()
  }
  // Mock data for testing
  return [
    { id: '1', source_entity: 'User', target_entity: 'Remy', relationship: 'uses', confidence: 1.0, created_at: Date.now() - 86400000 },
    { id: '2', source_entity: 'Remy', target_entity: 'Go', relationship: 'built_with', confidence: 1.0, created_at: Date.now() - 86400000 },
    { id: '3', source_entity: 'Remy', target_entity: 'SQLite', relationship: 'uses', confidence: 1.0, created_at: Date.now() - 86400000 },
  ]
}

export async function getScratchpad() {
  if (isWails) {
    return window.go.main.App.GetScratchpad()
  }
  return 'This is a mock scratchpad. The agent uses this space for working memory.'
}

export async function updateScratchpad(content) {
  if (isWails) {
    return window.go.main.App.UpdateScratchpad(content)
  }
  return null
}

export async function searchMemory(query, searchType) {
  if (isWails) {
    return window.go.main.App.SearchMemory(query, searchType)
  }
  // Mock results
  return {
    facts: [],
    episodes: [],
  }
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
