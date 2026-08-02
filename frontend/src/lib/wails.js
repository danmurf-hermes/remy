const isWails = typeof window !== 'undefined' && window.runtime !== undefined

function getRuntime() {
  if (isWails) {
    return window.runtime
  }
  return null
}

export async function sendMessage(text) {
  if (isWails) {
    return window.go.app.App.SendMessage(text)
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
    return window.go.app.App.SendMessageStream(text)
  }
  return null
}

export async function getHistory(limit, offset) {
  if (isWails) {
    return window.go.app.App.GetHistory(limit, offset)
  }
  return []
}

export async function getConversations() {
  if (isWails) {
    return window.go.app.App.GetConversations()
  }
  return [{ id: 'default', name: 'default', last_msg: 'No messages yet', timestamp: Date.now() }]
}

export async function getPersonas() {
  if (isWails) {
    return window.go.app.App.GetPersonas()
  }
  return [{ name: 'default', description: 'Default persona', is_active: true }]
}

export async function switchPersona(name) {
  if (isWails) {
    return window.go.app.App.SwitchPersona(name)
  }
  return null
}

export async function getActivePersona() {
  if (isWails) {
    return window.go.app.App.GetActivePersona()
  }
  return 'default'
}

// --- Memory Explorer bindings ---

export async function getFacts(category) {
  if (isWails) {
    return window.go.app.App.GetFacts(category)
  }
  // Mock data for testing
  const mockFacts = [
    {
      id: '1',
      fact: 'User prefers dark mode',
      category: 'preferences',
      confidence: 0.9,
      source: 'consolidation',
      created_at: Date.now() - 86400000,
      updated_at: Date.now() - 3600000,
    },
    {
      id: '2',
      fact: 'User works as a software engineer',
      category: 'personal',
      confidence: 0.85,
      source: 'consolidation',
      created_at: Date.now() - 172800000,
      updated_at: Date.now() - 7200000,
    },
    {
      id: '3',
      fact: 'User likes Python and Go',
      category: 'preferences',
      confidence: 0.75,
      source: 'consolidation',
      created_at: Date.now() - 259200000,
      updated_at: Date.now() - 10800000,
    },
  ]
  if (category) {
    return mockFacts.filter((f) => f.category === category)
  }
  return mockFacts
}

export async function getFact(id) {
  if (isWails) {
    return window.go.app.App.GetFact(id)
  }
  return null
}

export async function updateFact(id, fact, category, confidence) {
  if (isWails) {
    return window.go.app.App.UpdateFact(id, fact, category, confidence)
  }
  return null
}

export async function deleteFact(id) {
  if (isWails) {
    return window.go.app.App.DeleteFact(id)
  }
  return null
}

export async function getEpisodes(limit, offset) {
  if (isWails) {
    return window.go.app.App.GetEpisodes(limit, offset)
  }
  // Mock data for testing
  return [
    {
      id: '1',
      summary: 'Discussed project architecture and decided on microservices approach',
      start_time: Date.now() - 86400000,
      end_time: Date.now() - 82800000,
      message_ids: 'a,b,c',
      importance: 0.8,
      topics: 'architecture,planning',
    },
    {
      id: '2',
      summary: 'User asked about Go concurrency patterns',
      start_time: Date.now() - 172800000,
      end_time: Date.now() - 171000000,
      message_ids: 'd,e',
      importance: 0.5,
      topics: 'go,concurrency',
    },
    {
      id: '3',
      summary: 'Debugged database migration issue together',
      start_time: Date.now() - 259200000,
      end_time: Date.now() - 257400000,
      message_ids: 'f,g,h',
      importance: 0.7,
      topics: 'database,migration,debugging',
    },
  ]
}

export async function getEpisode(id) {
  if (isWails) {
    return window.go.app.App.GetEpisode(id)
  }
  return null
}

export async function getEntities() {
  if (isWails) {
    return window.go.app.App.GetEntities()
  }
  // Mock data for testing
  return [
    {
      id: '1',
      name: 'User',
      type: 'person',
      description: 'The primary user of Remy',
      created_at: Date.now() - 86400000,
    },
    {
      id: '2',
      name: 'Remy',
      type: 'ai',
      description: 'The personal AI assistant',
      created_at: Date.now() - 86400000,
    },
    {
      id: '3',
      name: 'Go',
      type: 'language',
      description: 'Go programming language',
      created_at: Date.now() - 172800000,
    },
    {
      id: '4',
      name: 'SQLite',
      type: 'technology',
      description: 'SQLite database engine',
      created_at: Date.now() - 172800000,
    },
  ]
}

export async function getRelationships() {
  if (isWails) {
    return window.go.app.App.GetRelationships()
  }
  // Mock data for testing
  return [
    {
      id: '1',
      source_entity: 'User',
      target_entity: 'Remy',
      relationship: 'uses',
      confidence: 1.0,
      created_at: Date.now() - 86400000,
    },
    {
      id: '2',
      source_entity: 'Remy',
      target_entity: 'Go',
      relationship: 'built_with',
      confidence: 1.0,
      created_at: Date.now() - 86400000,
    },
    {
      id: '3',
      source_entity: 'Remy',
      target_entity: 'SQLite',
      relationship: 'uses',
      confidence: 1.0,
      created_at: Date.now() - 86400000,
    },
  ]
}

export async function getScratchpad() {
  if (isWails) {
    return window.go.app.App.GetScratchpad()
  }
  return 'This is a mock scratchpad. The agent uses this space for working memory.'
}

export async function updateScratchpad(content) {
  if (isWails) {
    return window.go.app.App.UpdateScratchpad(content)
  }
  return null
}

export async function searchMemory(query, searchType) {
  if (isWails) {
    return window.go.app.App.SearchMemory(query, searchType)
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

// --- Stage 10: Tasks, Activity Log, Config bindings ---

export async function getTasks(status) {
  if (isWails) {
    return window.go.app.App.GetTasks(status)
  }
  // Mock data for testing
  const now = Date.now()
  return [
    {
      id: 'task-1',
      type: 'reminder',
      status: 'pending',
      trigger_at: now + 3600000,
      cron_expr: '',
      action: '{"text":"Check email"}',
      context: '',
      created_at: now - 86400000,
      fired_at: 0,
    },
    {
      id: 'task-2',
      type: 'reminder',
      status: 'pending',
      trigger_at: now + 7200000,
      cron_expr: '',
      action: '{"text":"Stand up meeting"}',
      context: '',
      created_at: now - 43200000,
      fired_at: 0,
    },
    {
      id: 'task-3',
      type: 'reminder',
      status: 'fired',
      trigger_at: now - 3600000,
      cron_expr: '',
      action: '{"text":"Morning coffee"}',
      context: '',
      created_at: now - 86400000,
      fired_at: now - 3600000,
    },
    {
      id: 'task-4',
      type: 'reminder',
      status: 'pending',
      trigger_at: now + 86400000,
      cron_expr: '0 9 * * *',
      action: '{"text":"Daily standup"}',
      context: '',
      created_at: now - 172800000,
      fired_at: 0,
    },
  ]
}

export async function createTask(taskType, triggerAt, cronExpr, action, context) {
  if (isWails) {
    return window.go.app.App.CreateTask(taskType, triggerAt, cronExpr, action, context)
  }
  return {
    id: 'new-task-' + Date.now(),
    type: taskType,
    status: 'pending',
    trigger_at: parseInt(triggerAt) || Date.now() + 3600000,
    cron_expr: cronExpr || '',
    action: action,
    context: context || '',
    created_at: Date.now(),
    fired_at: 0,
  }
}

export async function cancelTask(id) {
  if (isWails) {
    return window.go.app.App.CancelTask(id)
  }
  return null
}

export async function getActivityLog(filter, limit, offset) {
  if (isWails) {
    return window.go.app.App.GetActivityLog(filter, limit, offset)
  }
  // Mock data for testing
  const now = Date.now()
  return [
    {
      id: 'act-1',
      timestamp: now - 60000,
      type: 'message',
      details: '{"role":"user","content":"Hello Remy"}',
      message_id: 'msg-1',
      session_id: 'default',
    },
    {
      id: 'act-2',
      timestamp: now - 55000,
      type: 'retrieval',
      details: '{"query":"user preferences","results":2}',
      message_id: 'msg-1',
      session_id: 'default',
    },
    {
      id: 'act-3',
      timestamp: now - 50000,
      type: 'llm',
      details: '{"model":"llama3.1:8b","tokens":145,"duration_ms":3200}',
      message_id: 'msg-1',
      session_id: 'default',
    },
    {
      id: 'act-4',
      timestamp: now - 45000,
      type: 'consolidation',
      details: '{"episodes_created":1,"facts_extracted":3}',
      message_id: '',
      session_id: 'default',
    },
    {
      id: 'act-5',
      timestamp: now - 120000,
      type: 'error',
      details: '{"error":"LLM request timed out","retry":true}',
      message_id: 'msg-0',
      session_id: 'default',
    },
  ]
}

export async function getConfig() {
  if (isWails) {
    return window.go.app.App.GetConfig()
  }
  return {
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
    persona: {
      active: 'default',
      directory: '~/.remy/personas/',
    },
    interfaces: {
      telegram: {
        enabled: false,
        bot_token: '',
        allowed_users: [],
      },
    },
  }
}

export async function updateConfig(config) {
  if (isWails) {
    return window.go.app.App.UpdateConfig(config)
  }
  return null
}

export async function getAvailableModels(endpoint) {
  if (isWails) {
    return window.go.app.App.GetAvailableModels(endpoint)
  }
  return ['llama3.1:8b', 'llama3.2:3b', 'nomic-embed-text', 'mistral:7b']
}
