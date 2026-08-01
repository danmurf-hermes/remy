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
  <div class="header">
    <h2>Settings</h2>
    <button class="btn-primary" on:click={handleSave} disabled={saving}>
      {saving ? 'Saving...' : 'Save'}
    </button>
  </div>

  {#if error}
    <div class="error-banner">{error}</div>
  {/if}
  {#if success}
    <div class="success-banner">{success}</div>
  {/if}

  {#if loading}
    <div class="loading">Loading settings...</div>
  {:else}
    <div class="settings-tabs">
      {#each settingsTabs as tab}
        <button
          class="settings-tab"
          class:active={activeTab === tab.id}
          on:click={() => (activeTab = tab.id)}
        >
          {tab.label}
        </button>
      {/each}
    </div>

    <div class="settings-content">
      {#if activeTab === 'providers'}
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
          <label class="toggle-label">
            <span>Dark Mode</span>
            <input type="checkbox" bind:checked={$darkMode} class="toggle" />
          </label>
          <p class="hint">When disabled, follows system preference.</p>
        </div>
      {:else if activeTab === 'telegram'}
        <div class="section">
          <h3>Telegram Integration</h3>
          <label class="toggle-label">
            <span>Enable Telegram Bot</span>
            <input type="checkbox" bind:checked={telegramEnabled} class="toggle" />
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
    </div>
  {/if}
</div>

<style>
  .settings {
    flex: 1;
    padding: 24px;
    overflow-y: auto;
  }

  .header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 20px;
  }

  .header h2 {
    margin: 0;
    font-size: 20px;
    font-weight: 600;
  }

  .btn-primary {
    background: #0071e3;
    color: white;
    border: none;
    padding: 8px 16px;
    border-radius: 8px;
    cursor: pointer;
    font-size: 13px;
    font-weight: 500;
  }

  .btn-primary:hover {
    background: #0077ed;
  }

  .btn-primary:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .btn-secondary {
    background: #e8e8ed;
    color: #1d1d1f;
    border: none;
    padding: 8px 16px;
    border-radius: 8px;
    cursor: pointer;
    font-size: 13px;
  }

  .btn-danger {
    background: #ff3b30;
    color: white;
    border: none;
    padding: 8px 16px;
    border-radius: 8px;
    cursor: pointer;
    font-size: 13px;
  }

  .error-banner {
    background: #fff0f0;
    color: #c41e3a;
    padding: 8px 12px;
    border-radius: 8px;
    margin-bottom: 12px;
    font-size: 13px;
  }

  .success-banner {
    background: #e8f8e8;
    color: #1a7d1a;
    padding: 8px 12px;
    border-radius: 8px;
    margin-bottom: 12px;
    font-size: 13px;
  }

  .loading {
    text-align: center;
    color: #86868b;
    padding: 40px;
  }

  .settings-tabs {
    display: flex;
    gap: 4px;
    margin-bottom: 20px;
    border-bottom: 1px solid #e8e8ed;
    padding-bottom: 0;
  }

  .settings-tab {
    padding: 8px 16px;
    border: none;
    background: transparent;
    font-size: 13px;
    cursor: pointer;
    color: #6e6e73;
    border-bottom: 2px solid transparent;
    margin-bottom: -1px;
    transition: all 0.15s;
  }

  .settings-tab:hover {
    color: #1d1d1f;
  }

  .settings-tab.active {
    color: #0071e3;
    border-bottom-color: #0071e3;
  }

  .section {
    max-width: 600px;
  }

  .section h3 {
    margin: 20px 0 12px 0;
    font-size: 15px;
    font-weight: 600;
  }

  .section h3:first-child {
    margin-top: 0;
  }

  .provider-card {
    background: #f5f5f7;
    border-radius: 12px;
    padding: 16px;
    margin-bottom: 12px;
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
    border-radius: 10px;
    font-weight: 500;
  }

  .status-badge.connected {
    background: #e8f8e8;
    color: #1a7d1a;
  }

  .status-badge.disconnected {
    background: #fff0f0;
    color: #c41e3a;
  }

  .input {
    width: 100%;
    padding: 8px 12px;
    border: 1px solid #d2d2d7;
    border-radius: 8px;
    font-size: 13px;
    box-sizing: border-box;
    margin-bottom: 8px;
  }

  .slider {
    width: 100%;
    margin-bottom: 12px;
  }

  label {
    display: block;
    font-size: 12px;
    color: #6e6e73;
    margin-bottom: 4px;
  }

  .toggle-label {
    display: flex;
    justify-content: space-between;
    align-items: center;
    font-size: 13px;
    color: #1d1d1f;
    margin-bottom: 12px;
  }

  .toggle {
    width: 20px;
    height: 20px;
  }

  .hint {
    font-size: 12px;
    color: var(--text-secondary, #6e6e73);
    margin-top: -8px;
    margin-bottom: 12px;
  }

  .data-actions {
    display: flex;
    gap: 8px;
    margin-top: 8px;
  }

  .about-info {
    background: #f5f5f7;
    border-radius: 12px;
    padding: 16px;
  }

  .about-row {
    display: flex;
    justify-content: space-between;
    padding: 8px 0;
    border-bottom: 1px solid #e8e8ed;
  }

  .about-row:last-child {
    border-bottom: none;
  }

  .about-label {
    font-size: 13px;
    color: #6e6e73;
  }

  .about-value {
    font-size: 13px;
    font-weight: 500;
    color: #1d1d1f;
  }
</style>
