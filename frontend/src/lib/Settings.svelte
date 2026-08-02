<script>
  import { onMount } from 'svelte'
  import { config, darkMode } from './stores.js'
  import { getConfig, updateConfig } from './wails.js'

  let loading = true
  let saving = false
  let error = null
  let success = null
  let activeTab = 'providers'

  // Editable fields
  let providers = {}
  let defaultProvider = ''
  let dbPath = ''
  let workingMemoryTurns = 20
  let quickConsolidationDelayMs = 300000
  let deepConsolidationDelayMs = 1800000
  let telegramEnabled = false
  let botToken = ''
  let temperature = 0.7
  let maxTokens = 4096

  onMount(async () => {
    await loadConfig()
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
            allowed_users: $config?.interfaces?.telegram?.allowed_users || [],
          },
        },
      }
      await updateConfig(cfg)
      config.set(cfg)
      success = 'Settings saved successfully'
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

  const settingsTabs = [
    { id: 'providers', label: 'Providers' },
    { id: 'model', label: 'Model' },
    { id: 'appearance', label: 'Appearance' },
    { id: 'telegram', label: 'Telegram' },
    { id: 'memory', label: 'Memory' },
    { id: 'about', label: 'About' },
  ]
</script>

<div class="settings">
  <div class="settings-layout">
    <nav class="settings-nav">
      {#each settingsTabs as tab}
        <button
          class="settings-nav-item"
          class:active={activeTab === tab.id}
          on:click={() => (activeTab = tab.id)}
        >
          {tab.label}
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
        <div class="loading">Loading settings…</div>
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
              <label>
                Endpoint
                <input type="text" bind:value={p.endpoint} class="input" />
              </label>
              <label>
                Chat Model
                <input type="text" bind:value={p.chat_model} class="input" />
              </label>
              <label>
                Embedding Model
                <input type="text" bind:value={p.embedding_model} class="input" />
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
          <label>
            Temperature: {temperature}
            <input
              type="range"
              min="0"
              max="2"
              step="0.1"
              bind:value={temperature}
              class="slider"
            />
          </label>
          <label>
            Max Tokens: {maxTokens}
            <input
              type="range"
              min="256"
              max="32768"
              step="256"
              bind:value={maxTokens}
              class="slider"
            />
          </label>
        </div>
      {:else if activeTab === 'appearance'}
        <div class="section">
          <h3>Appearance</h3>
          <label class="toggle-row">
            <span>Dark Mode</span>
            <label class="switch">
              <input type="checkbox" bind:checked={$darkMode} />
              <span class="slider-track"></span>
            </label>
          </label>
          <p class="hint">When disabled, follows system preference.</p>
        </div>
      {:else if activeTab === 'telegram'}
        <div class="section">
          <h3>Telegram Integration</h3>
          <label class="toggle-row">
            <span>Enable Telegram Bot</span>
            <label class="switch">
              <input type="checkbox" bind:checked={telegramEnabled} />
              <span class="slider-track"></span>
            </label>
          </label>
          {#if telegramEnabled}
            <label>
              Bot Token
              <input
                type="password"
                bind:value={botToken}
                class="input"
                placeholder="Enter your bot token"
              />
            </label>
          {/if}
        </div>
      {:else if activeTab === 'memory'}
        <div class="section">
          <h3>Memory Settings</h3>
          <label>
            Database Path
            <input type="text" bind:value={dbPath} class="input" />
          </label>
          <label>
            Working Memory Turns: {workingMemoryTurns}
            <input
              type="range"
              min="5"
              max="100"
              step="1"
              bind:value={workingMemoryTurns}
              class="slider"
            />
          </label>
          <label>
            Quick Consolidation Delay: {quickConsolidationDelayMs / 1000}s
            <input
              type="range"
              min="60000"
              max="600000"
              step="10000"
              bind:value={quickConsolidationDelayMs}
              class="slider"
            />
          </label>
          <label>
            Deep Consolidation Delay: {deepConsolidationDelayMs / 60000}min
            <input
              type="range"
              min="300000"
              max="7200000"
              step="60000"
              bind:value={deepConsolidationDelayMs}
              class="slider"
            />
          </label>

          <h3>Data Management</h3>
          <div class="data-actions">
            <button class="btn-secondary">Export Data</button>
            <button class="btn-secondary">Import Data</button>
            <button class="btn-danger">Clear All Data</button>
          </div>
        </div>
      {:else if activeTab === 'about'}
        <div class="section">
          <h3>About Remy</h3>
          <div class="about-info">
            <div class="about-row">
              <span class="about-label">Version</span>
              <span class="about-value">0.1.0</span>
            </div>
            <div class="about-row">
              <span class="about-label">Build</span>
              <span class="about-value">Stage 10</span>
            </div>
            <div class="about-row">
              <span class="about-label">Go Version</span>
              <span class="about-value">1.26</span>
            </div>
            <div class="about-row">
              <span class="about-label">Frontend</span>
              <span class="about-value">Svelte 4 + Vite</span>
            </div>
            <div class="about-row">
              <span class="about-label">Database</span>
              <span class="about-value">SQLite + sqlite-vec</span>
            </div>
            <div class="about-row">
              <span class="about-label">LLM Provider</span>
              <span class="about-value">Ollama (OpenAI-compatible)</span>
            </div>
          </div>
        </div>
      {/if}

      <div class="settings-footer">
        <button class="btn-primary" on:click={handleSave} disabled={saving}>
          {saving ? 'Saving…' : 'Save'}
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
    padding: 20px 12px;
    background: var(--bg-secondary);
    backdrop-filter: blur(20px) saturate(180%);
    -webkit-backdrop-filter: blur(20px) saturate(180%);
    flex-shrink: 0;
  }

  .settings-nav-item {
    display: block;
    width: 100%;
    padding: 8px 12px;
    border: none;
    background: transparent;
    border-radius: var(--radius-md);
    font-size: 13px;
    text-align: left;
    cursor: pointer;
    color: var(--text-primary);
    transition: all 0.15s ease;
    margin-bottom: 2px;
  }

  .settings-nav-item:hover {
    background: var(--hover-bg);
  }

  .settings-nav-item.active {
    background: var(--accent);
    color: white;
    font-weight: 500;
  }

  .settings-content {
    flex: 1;
    overflow-y: auto;
    padding: 24px 32px;
  }

  .settings-header h2 {
    margin: 0 0 20px 0;
    font-size: 22px;
    font-weight: 600;
  }

  .btn-primary {
    background: var(--accent);
    color: white;
    border: none;
    padding: 8px 20px;
    border-radius: var(--radius-md);
    cursor: pointer;
    font-size: 13px;
    font-weight: 500;
    transition: all 0.15s ease;
  }

  .btn-primary:hover {
    background: var(--accent-hover);
  }

  .btn-primary:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .btn-secondary {
    background: var(--bg-tertiary);
    color: var(--text-primary);
    border: none;
    padding: 8px 16px;
    border-radius: var(--radius-md);
    cursor: pointer;
    font-size: 13px;
    transition: all 0.15s ease;
  }

  .btn-secondary:hover {
    background: var(--hover-bg);
  }

  .btn-danger {
    background: var(--danger);
    color: white;
    border: none;
    padding: 8px 16px;
    border-radius: var(--radius-md);
    cursor: pointer;
    font-size: 13px;
    transition: all 0.15s ease;
  }

  .btn-danger:hover {
    background: var(--danger-hover);
  }

  .error-banner {
    background: rgba(255, 59, 48, 0.1);
    color: var(--danger);
    padding: 8px 12px;
    border-radius: var(--radius-md);
    margin-bottom: 12px;
    font-size: 13px;
    border: 1px solid rgba(255, 59, 48, 0.2);
  }

  .success-banner {
    background: rgba(52, 199, 89, 0.1);
    color: var(--success);
    padding: 8px 12px;
    border-radius: var(--radius-md);
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
    margin: 24px 0 12px 0;
    font-size: 15px;
    font-weight: 600;
  }

  .section h3:first-child {
    margin-top: 0;
  }

  .provider-card {
    background: var(--card-bg);
    border: 1px solid var(--border-light);
    border-radius: var(--radius-lg);
    padding: 16px;
    margin-bottom: 12px;
    backdrop-filter: blur(10px);
    -webkit-backdrop-filter: blur(10px);
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
    border-radius: var(--radius-xl);
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
    border-radius: var(--radius-md);
    font-size: 13px;
    box-sizing: border-box;
    margin-bottom: 8px;
    background: var(--input-bg);
    color: var(--text-primary);
    outline: none;
    transition: border-color 0.15s;
  }

  .input:focus {
    border-color: var(--accent);
    box-shadow: 0 0 0 3px rgba(0, 113, 227, 0.15);
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

  .toggle-row {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 12px 0;
    border-bottom: 1px solid var(--border-light);
  }

  .toggle-row span {
    font-size: 14px;
    color: var(--text-primary);
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

  .hint {
    font-size: 12px;
    color: var(--text-tertiary);
    margin: 8px 0 0 0;
  }

  .data-actions {
    display: flex;
    gap: 8px;
    margin-top: 8px;
  }

  .about-info {
    background: var(--card-bg);
    border: 1px solid var(--border-light);
    border-radius: var(--radius-lg);
    padding: 16px;
    backdrop-filter: blur(10px);
    -webkit-backdrop-filter: blur(10px);
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
