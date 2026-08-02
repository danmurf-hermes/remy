<script>
  import { onMount } from 'svelte'
  import { config, darkMode } from './stores.js'
  import { getConfig, updateConfig, getAvailableModels } from './wails.js'

  let loading = true
  let saving = false
  let error = null
  let success = null
  let activeTab = 'providers'

  let providers = {}
  let defaultProvider = ''
  let dbPath = ''
  let workingMemoryTurns = 20
  let quickConsolidationDelayMs = 300000
  let deepConsolidationDelayMs = 1800000
  let telegramEnabled = false
  let botToken = ''
  let allowedUsers = ''
  let temperature = 0.7
  let maxTokens = 4096
  const availableModels = {}
  const loadingModels = {}

  onMount(async () => {
    await loadConfig()
    await fetchAllModels()
  })

  async function loadConfig() {
    loading = true
    try {
      const cfg = await getConfig()
      config.set(cfg)
      providers = cfg.providers || {}
      defaultProvider = cfg.default_provider || 'ollama'
      dbPath = cfg.memory?.db_path || '~/.remy/memory.db'
      workingMemoryTurns = cfg.memory?.working_memory_turns || 20
      quickConsolidationDelayMs = cfg.memory?.quick_consolidation_delay_ms || 300000
      deepConsolidationDelayMs = cfg.memory?.deep_consolidation_delay_ms || 1800000
      telegramEnabled = cfg.interfaces?.telegram?.enabled || false
      botToken = cfg.interfaces?.telegram?.bot_token || ''
      allowedUsers = (cfg.interfaces?.telegram?.allowed_users || []).join(', ')
      const params = Object.values(providers)[0]?.parameters || {}
      temperature = params.temperature || 0.7
      maxTokens = params.max_tokens || 4096
    } catch (e) {
      error = 'Failed to load config: ' + e.message
    }
    loading = false
  }

  async function handleSave() {
    saving = true
    error = null
    success = null
    try {
      const cfg = {
        providers: Object.fromEntries(
          Object.entries(providers).map(([name, p]) => [
            name,
            {
              endpoint: p.endpoint,
              chat_model: p.chat_model,
              embedding_model: p.embedding_model,
              parameters: { temperature, max_tokens: maxTokens },
            },
          ]),
        ),
        default_provider: defaultProvider,
        memory: {
          db_path: dbPath,
          working_memory_turns: workingMemoryTurns,
          quick_consolidation_delay_ms: quickConsolidationDelayMs,
          deep_consolidation_delay_ms: deepConsolidationDelayMs,
        },
        persona: $config?.persona || { active: 'default', directory: '~/.remy/personas/' },
        interfaces: {
          telegram: {
            enabled: telegramEnabled,
            bot_token: botToken,
            allowed_users: allowedUsers
              .split(',')
              .map((u) => u.trim())
              .filter((u) => u.length > 0),
          },
        },
      }
      await updateConfig(cfg)
      config.set(cfg)
      success = 'Settings saved'
      setTimeout(() => (success = null), 3000)
    } catch (e) {
      error = 'Failed to save config: ' + e.message
    }
    saving = false
  }

  function getProviderStatus(name) {
    const p = providers[name]
    if (!p) {
      return 'disconnected'
    }
    if (p.endpoint) {
      return 'connected'
    }
    return 'disconnected'
  }

  async function fetchModels(name) {
    const p = providers[name]
    if (!p || !p.endpoint) {
      return
    }
    loadingModels[name] = true
    try {
      availableModels[name] = await getAvailableModels(p.endpoint)
    } catch (e) {
      availableModels[name] = []
    }
    loadingModels[name] = false
  }

  async function fetchAllModels() {
    for (const name of Object.keys(providers)) {
      await fetchModels(name)
    }
  }

  const settingsTabs = [
    { id: 'providers', icon: '🔌', label: 'Providers' },
    { id: 'model', icon: '🤖', label: 'Model' },
    { id: 'appearance', icon: '🎨', label: 'Appearance' },
    { id: 'telegram', icon: '📱', label: 'Telegram' },
    { id: 'memory', icon: '🧠', label: 'Memory' },
    { id: 'about', icon: 'ℹ️', label: 'About' },
  ]
</script>

<div class="settings">
  <div class="settings-layout">
    <nav class="settings-nav">
      <div class="settings-nav-header">Settings</div>
      {#each settingsTabs as tab}
        <button
          class="settings-nav-item"
          class:active={activeTab === tab.id}
          on:click={() => (activeTab = tab.id)}
        >
          <span class="nav-icon">{tab.icon}</span>
          <span class="nav-label">{tab.label}</span>
        </button>
      {/each}
    </nav>

    <div class="settings-content">
      <div class="settings-header">
        <h2>{settingsTabs.find((t) => t.id === activeTab)?.label}</h2>
      </div>

      {#if error}
        <div class="error-banner">{error}</div>
      {/if}
      {#if success}
        <div class="success-banner">{success}</div>
      {/if}

      {#if loading}
        <div class="loading">Loading…</div>
      {:else if activeTab === 'providers'}
        <div class="section">
          <h3>Provider Management</h3>
          {#each Object.entries(providers) as [name, p]}
            <div class="provider-card">
              <div class="provider-header">
                <span class="provider-name">{name}</span>
                <span
                  class="status-badge"
                  class:connected={getProviderStatus(name) === 'connected'}
                  class:disconnected={getProviderStatus(name) === 'disconnected'}
                >
                  {getProviderStatus(name)}
                </span>
              </div>
              <label>Endpoint <input type="text" bind:value={p.endpoint} class="input" /></label>
              <label
                >Chat Model
                <select bind:value={p.chat_model} class="input">
                  {#if loadingModels[name]}
                    <option value="">Loading…</option>
                  {:else if availableModels[name]?.length}
                    {#each availableModels[name] as m}
                      <option value={m}>{m}</option>
                    {/each}
                  {:else}
                    <option value={p.chat_model}>{p.chat_model || 'No models found'}</option>
                  {/if}
                </select>
                <button class="btn-small btn-outline" on:click={() => fetchModels(name)}
                  >Refresh</button
                >
              </label>
              <label
                >Embedding Model
                <select bind:value={p.embedding_model} class="input">
                  {#if loadingModels[name]}
                    <option value="">Loading…</option>
                  {:else if availableModels[name]?.length}
                    {#each availableModels[name] as m}
                      <option value={m}>{m}</option>
                    {/each}
                  {:else}
                    <option value={p.embedding_model}
                      >{p.embedding_model || 'No models found'}</option
                    >
                  {/if}
                </select>
                <button class="btn-small btn-outline" on:click={() => fetchModels(name)}
                  >Refresh</button
                >
              </label>
            </div>
          {/each}
          <h3>Default Provider</h3>
          <select bind:value={defaultProvider} class="input">
            {#each Object.keys(providers) as name}
              <option value={name}>{name}</option>
            {/each}
          </select>
        </div>
      {:else if activeTab === 'model'}
        <div class="section">
          <h3>Model Parameters</h3>
          <label
            >Temperature: {temperature}
            <input
              type="range"
              min="0"
              max="2"
              step="0.1"
              bind:value={temperature}
              class="slider"
            /></label
          >
          <label
            >Max Tokens: {maxTokens}
            <input
              type="range"
              min="256"
              max="32768"
              step="256"
              bind:value={maxTokens}
              class="slider"
            /></label
          >
        </div>
      {:else if activeTab === 'appearance'}
        <div class="section">
          <h3>Appearance</h3>
          <div class="setting-row">
            <div class="setting-info">
              <span class="setting-label">Dark Mode</span>
              <span class="setting-desc">When disabled, follows system preference</span>
            </div>
            <label class="switch">
              <input type="checkbox" bind:checked={$darkMode} />
              <span class="slider-track"></span>
            </label>
          </div>
        </div>
      {:else if activeTab === 'telegram'}
        <div class="section">
          <h3>Telegram Integration</h3>
          <div class="setting-row">
            <div class="setting-info">
              <span class="setting-label">Enable Telegram Bot</span>
              <span class="setting-desc">Connect Remy to Telegram</span>
            </div>
            <label class="switch">
              <input type="checkbox" bind:checked={telegramEnabled} />
              <span class="slider-track"></span>
            </label>
          </div>
          {#if telegramEnabled}
            <label
              >Bot Token <input
                type="password"
                bind:value={botToken}
                class="input"
                placeholder="Enter your bot token"
              /></label
            >
            <label
              >Allowed User IDs <input
                type="text"
                bind:value={allowedUsers}
                class="input"
                placeholder="Comma-separated Telegram user IDs (leave empty to allow all)"
              /></label
            >
          {/if}
        </div>
      {:else if activeTab === 'memory'}
        <div class="section">
          <h3>Memory Settings</h3>
          <label>Database Path <input type="text" bind:value={dbPath} class="input" /></label>
          <label
            >Working Memory Turns: {workingMemoryTurns}
            <input
              type="range"
              min="5"
              max="100"
              step="1"
              bind:value={workingMemoryTurns}
              class="slider"
            /></label
          >
          <label
            >Quick Consolidation: {quickConsolidationDelayMs / 1000}s
            <input
              type="range"
              min="60000"
              max="600000"
              step="10000"
              bind:value={quickConsolidationDelayMs}
              class="slider"
            /></label
          >
          <label
            >Deep Consolidation: {deepConsolidationDelayMs / 60000}min
            <input
              type="range"
              min="300000"
              max="7200000"
              step="60000"
              bind:value={deepConsolidationDelayMs}
              class="slider"
            /></label
          >
          <h3>Data Management</h3>
          <div class="data-actions">
            <button class="btn-secondary">Export</button>
            <button class="btn-secondary">Import</button>
            <button class="btn-danger">Clear All</button>
          </div>
        </div>
      {:else if activeTab === 'about'}
        <div class="section">
          <h3>About Remy</h3>
          <div class="about-info">
            <div class="about-row">
              <span class="about-label">Version</span><span class="about-value">0.1.0</span>
            </div>
            <div class="about-row">
              <span class="about-label">Build</span><span class="about-value">Stage 10</span>
            </div>
            <div class="about-row">
              <span class="about-label">Go Version</span><span class="about-value">1.26</span>
            </div>
            <div class="about-row">
              <span class="about-label">Frontend</span><span class="about-value"
                >Svelte 4 + Vite</span
              >
            </div>
            <div class="about-row">
              <span class="about-label">Database</span><span class="about-value"
                >SQLite + sqlite-vec</span
              >
            </div>
            <div class="about-row">
              <span class="about-label">LLM</span><span class="about-value">Ollama</span>
            </div>
          </div>
        </div>
      {/if}

      <div class="settings-footer">
        <button class="btn-primary" on:click={handleSave} disabled={saving}>
          {saving ? 'Saving…' : 'Save Changes'}
        </button>
      </div>
    </div>
  </div>
</div>

<style>
  .settings {
    flex: 1;
    display: flex;
    overflow: hidden;
  }
  .settings-layout {
    display: flex;
    flex: 1;
    overflow: hidden;
  }

  .settings-nav {
    width: 200px;
    border-right: 1px solid var(--border-color);
    padding: 20px 8px;
    background: var(--bg-secondary);
    backdrop-filter: blur(30px) saturate(200%);
    -webkit-backdrop-filter: blur(30px) saturate(200%);
    flex-shrink: 0;
  }

  .settings-nav-header {
    font-size: 11px;
    font-weight: 600;
    color: var(--text-tertiary);
    text-transform: uppercase;
    letter-spacing: 0.8px;
    padding: 0 12px 12px;
    margin-bottom: 4px;
    border-bottom: 1px solid var(--border-light);
  }

  .settings-nav-item {
    display: flex;
    align-items: center;
    gap: 8px;
    width: 100%;
    padding: 8px 12px;
    border: none;
    background: transparent;
    border-radius: 8px;
    font-size: 13px;
    text-align: left;
    cursor: pointer;
    color: var(--text-primary);
    margin-bottom: 2px;
    transition: all 0.15s;
  }

  .settings-nav-item:hover {
    background: var(--hover-bg);
  }
  .settings-nav-item.active {
    background: var(--accent);
    color: white;
    font-weight: 500;
  }
  .nav-icon {
    font-size: 14px;
  }
  .nav-label {
    font-size: 13px;
  }

  .settings-content {
    flex: 1;
    overflow-y: auto;
    padding: 24px 32px;
  }
  .settings-header h2 {
    margin: 0 0 20px;
    font-size: 22px;
    font-weight: 600;
  }

  .btn-primary {
    background: var(--accent);
    color: white;
    border: none;
    padding: 8px 20px;
    border-radius: 8px;
    cursor: pointer;
    font-size: 13px;
    font-weight: 500;
    transition: all 0.15s;
  }
  .btn-primary:hover {
    background: var(--accent-hover);
  }
  .btn-primary:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .btn-small {
    padding: 4px 10px;
    border-radius: 6px;
    border: none;
    cursor: pointer;
    font-size: 12px;
    transition: all 0.15s;
  }

  .btn-outline {
    background: transparent;
    border: 1px solid var(--border-color);
    color: var(--text-primary);
  }

  .btn-outline:hover {
    background: var(--hover-bg);
  }

  .btn-secondary {
    background: var(--bg-tertiary);
    color: var(--text-primary);
    border: none;
    padding: 8px 16px;
    border-radius: 8px;
    cursor: pointer;
    font-size: 13px;
  }
  .btn-secondary:hover {
    background: var(--hover-bg);
  }

  .btn-danger {
    background: var(--danger);
    color: white;
    border: none;
    padding: 8px 16px;
    border-radius: 8px;
    cursor: pointer;
    font-size: 13px;
  }
  .btn-danger:hover {
    background: var(--danger-hover);
  }

  .error-banner {
    background: rgba(255, 59, 48, 0.1);
    color: var(--danger);
    padding: 8px 12px;
    border-radius: 8px;
    margin-bottom: 12px;
    font-size: 13px;
    border: 1px solid rgba(255, 59, 48, 0.2);
  }
  .success-banner {
    background: rgba(52, 199, 89, 0.1);
    color: var(--success);
    padding: 8px 12px;
    border-radius: 8px;
    margin-bottom: 12px;
    font-size: 13px;
    border: 1px solid rgba(52, 199, 89, 0.2);
  }
  .loading {
    text-align: center;
    color: var(--text-tertiary);
    padding: 40px;
  }
  .section {
    max-width: 500px;
  }
  .section h3 {
    margin: 24px 0 12px;
    font-size: 15px;
    font-weight: 600;
  }
  .section h3:first-child {
    margin-top: 0;
  }

  .setting-row {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 16px 0;
    border-bottom: 1px solid var(--border-light);
  }
  .setting-info {
    display: flex;
    flex-direction: column;
    gap: 2px;
  }
  .setting-label {
    font-size: 14px;
    color: var(--text-primary);
  }
  .setting-desc {
    font-size: 12px;
    color: var(--text-tertiary);
  }

  .switch {
    position: relative;
    display: inline-block;
    width: 44px;
    height: 24px;
    flex-shrink: 0;
  }
  .switch input {
    opacity: 0;
    width: 0;
    height: 0;
  }
  .slider-track {
    position: absolute;
    cursor: pointer;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background: var(--bg-tertiary);
    border-radius: 24px;
    transition: all 0.2s;
  }
  .slider-track::before {
    content: '';
    position: absolute;
    height: 20px;
    width: 20px;
    left: 2px;
    bottom: 2px;
    background: white;
    border-radius: 50%;
    transition: all 0.2s;
    box-shadow: 0 1px 3px rgba(0, 0, 0, 0.2);
  }
  .switch input:checked + .slider-track {
    background: var(--accent);
  }
  .switch input:checked + .slider-track::before {
    transform: translateX(20px);
  }

  .provider-card {
    background: var(--card-bg);
    border: 1px solid var(--border-light);
    border-radius: 12px;
    padding: 16px;
    margin-bottom: 12px;
    backdrop-filter: blur(10px);
  }
  .provider-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 12px;
  }
  .provider-name {
    font-size: 14px;
    font-weight: 600;
  }
  .status-badge {
    font-size: 11px;
    padding: 2px 8px;
    border-radius: 20px;
    font-weight: 500;
  }
  .status-badge.connected {
    background: rgba(52, 199, 89, 0.15);
    color: var(--success);
  }
  .status-badge.disconnected {
    background: rgba(255, 59, 48, 0.1);
    color: var(--danger);
  }

  .input {
    width: 100%;
    padding: 8px 12px;
    border: 1px solid var(--border-color);
    border-radius: 8px;
    font-size: 13px;
    box-sizing: border-box;
    margin-bottom: 8px;
    background: var(--input-bg);
    color: var(--text-primary);
    outline: none;
  }
  .input:focus {
    border-color: var(--accent);
    box-shadow: 0 0 0 3px rgba(0, 113, 227, 0.12);
  }

  .slider {
    width: 100%;
    margin-bottom: 12px;
    -webkit-appearance: none;
    appearance: none;
    height: 4px;
    border-radius: 2px;
    background: var(--bg-tertiary);
    outline: none;
  }
  .slider::-webkit-slider-thumb {
    -webkit-appearance: none;
    width: 18px;
    height: 18px;
    border-radius: 50%;
    background: var(--accent);
    cursor: pointer;
    box-shadow: 0 1px 3px rgba(0, 0, 0, 0.2);
  }

  label {
    display: block;
    font-size: 12px;
    color: var(--text-secondary);
    margin-bottom: 4px;
  }
  .data-actions {
    display: flex;
    gap: 8px;
    margin-top: 8px;
  }

  .about-info {
    background: var(--card-bg);
    border: 1px solid var(--border-light);
    border-radius: 12px;
    padding: 16px;
    backdrop-filter: blur(10px);
  }
  .about-row {
    display: flex;
    justify-content: space-between;
    padding: 8px 0;
    border-bottom: 1px solid var(--border-light);
  }
  .about-row:last-child {
    border-bottom: none;
  }
  .about-label {
    font-size: 13px;
    color: var(--text-secondary);
  }
  .about-value {
    font-size: 13px;
    font-weight: 500;
    color: var(--text-primary);
  }

  .settings-footer {
    margin-top: 24px;
    padding-top: 16px;
    border-top: 1px solid var(--border-light);
  }
</style>
