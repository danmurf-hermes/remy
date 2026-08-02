<script>
  import { onMount } from 'svelte'
  import { personas, activePersona } from './stores.js'
  import { getPersonas, switchPersona } from './wails.js'

  let loading = true
  let selectedPersona = null
  let editingPersona = null
  let showPreview = false
  let showCreateDialog = false
  let newPersonaName = ''
  let newPersonaProvider = 'ollama'
  let newPersonaModel = 'llama3.1:8b'
  let newPersonaTemperature = 0.7
  let newPersonaMaxTokens = 4096
  let newPersonaBody = 'You are a helpful assistant.'
  let error = null

  onMount(async () => {
    await loadPersonas()
  })

  async function loadPersonas() {
    loading = true
    try {
      const result = await getPersonas()
      personas.set(result)
      const active = result.find((p) => p.is_active)
      if (active) {
        activePersona.set(active.name)
        selectedPersona = active
      }
    } catch (e) {
      error = 'Failed to load personas: ' + e.message
    }
    loading = false
  }

  async function handleSelect(name) {
    try {
      await switchPersona(name)
      activePersona.set(name)
      await loadPersonas()
    } catch (e) {
      error = 'Failed to switch persona: ' + e.message
    }
  }

  function handleEdit(persona) {
    selectedPersona = persona
    editingPersona = { ...persona }
  }

  async function handleCreate() {
    if (!newPersonaName) {
      return
    }
    error = null
    showCreateDialog = false
    newPersonaName = ''
    await loadPersonas()
  }

  function getProviderModel(persona) {
    const desc = persona.description || ''
    const parts = desc.split(' / ')
    return {
      provider: parts[0] || 'unknown',
      model: parts[1] || 'unknown',
    }
  }
</script>

<div class="persona-studio">
  <div class="header">
    <h2>Persona Studio</h2>
    <button class="btn-primary" on:click={() => (showCreateDialog = true)}>+ New Persona</button>
  </div>

  {#if error}
    <div class="error-banner">{error}</div>
  {/if}

  {#if showCreateDialog}
    <div class="dialog-overlay">
      <button
        class="dialog-close-bg"
        on:click={() => (showCreateDialog = false)}
        on:keydown={(e) => {
          if (e.key === 'Escape') {
            showCreateDialog = false
          }
        }}
      ></button>
      <!-- svelte-ignore a11y-click-events-have-key-events a11y-no-static-element-interactions -->
      <div class="dialog" on:click|stopPropagation>
        <h3>Create New Persona</h3>
        <label>
          Name
          <input
            type="text"
            bind:value={newPersonaName}
            class="input"
            placeholder="e.g. creative-writer"
          />
        </label>
        <label>
          Provider
          <select bind:value={newPersonaProvider} class="input">
            <option value="ollama">Ollama</option>
            <option value="openai">OpenAI</option>
          </select>
        </label>
        <label>
          Model
          <input type="text" bind:value={newPersonaModel} class="input" />
        </label>
        <label>
          Temperature: {newPersonaTemperature}
          <input
            type="range"
            min="0"
            max="2"
            step="0.1"
            bind:value={newPersonaTemperature}
            class="slider"
          />
        </label>
        <label>
          Max Tokens: {newPersonaMaxTokens}
          <input
            type="range"
            min="256"
            max="16384"
            step="256"
            bind:value={newPersonaMaxTokens}
            class="slider"
          />
        </label>
        <label>
          System Prompt
          <textarea bind:value={newPersonaBody} class="textarea" rows="4"></textarea>
        </label>
        <div class="dialog-actions">
          <button class="btn-secondary" on:click={() => (showCreateDialog = false)}>Cancel</button>
          <button class="btn-primary" on:click={handleCreate}>Create</button>
        </div>
      </div>
    </div>
  {/if}

  {#if loading}
    <div class="loading">Loading personas…</div>
  {:else}
    <div class="split-panel">
      <div class="list-panel">
        <h3>Personas ({$personas.length})</h3>
        {#each $personas as persona (persona.name)}
          {@const pm = getProviderModel(persona)}
          <div
            class="persona-card"
            class:active={persona.is_active}
            class:selected={selectedPersona?.name === persona.name}
            role="button"
            tabindex="0"
            on:click={() => handleEdit(persona)}
            on:keydown={(e) => {
              if (e.key === 'Enter' || e.key === ' ') {
                e.preventDefault()
                handleEdit(persona)
              }
            }}
          >
            <div class="persona-name">
              {persona.name}
              {#if persona.is_active}
                <span class="active-badge">Active</span>
              {/if}
            </div>
            <div class="persona-tags">
              <span class="tag">{pm.provider}</span>
              <span class="tag">{pm.model}</span>
            </div>
            {#if !persona.is_active}
              <button
                class="btn-small btn-outline"
                on:click|stopPropagation={() => handleSelect(persona.name)}
              >
                Activate
              </button>
            {/if}
          </div>
        {/each}
      </div>

      <div class="editor-panel">
        {#if editingPersona}
          <div class="editor-header">
            <h3>{editingPersona.name}</h3>
            <button class="btn-small btn-outline" on:click={() => (showPreview = !showPreview)}>
              {showPreview ? 'Edit' : 'Preview'}
            </button>
          </div>

          <div class="model-config">
            <h4>Model Configuration</h4>
            <label>
              Provider
              <input type="text" bind:value={editingPersona.provider} class="input" disabled />
            </label>
            <label>
              Model
              <input type="text" bind:value={editingPersona.model} class="input" />
            </label>
            <label>
              Temperature: 0.7
              <input type="range" min="0" max="2" step="0.1" value="0.7" class="slider" />
            </label>
            <label>
              Max Tokens: 4096
              <input type="range" min="256" max="16384" step="256" value="4096" class="slider" />
            </label>
          </div>

          <div class="markdown-editor">
            <h4>System Prompt</h4>
            {#if showPreview}
              <div class="preview">
                <p>You are a helpful AI assistant named Remy.</p>
                <p>You have access to the user's conversation history and memory.</p>
                <p>Be concise, accurate, and helpful.</p>
              </div>
            {:else}
              <textarea class="textarea" rows="8" placeholder="Enter system prompt in Markdown…">
                You are a helpful AI assistant named Remy. You have access to the user's
                conversation history and memory. Be concise, accurate, and helpful.
              </textarea>
            {/if}
          </div>
        {:else}
          <div class="no-selection">
            <p>Select a persona to edit</p>
          </div>
        {/if}
      </div>
    </div>
  {/if}
</div>

<style>
  .persona-studio {
    flex: 1;
    display: flex;
    flex-direction: column;
    overflow: hidden;
  }

  .header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 24px 24px 0;
  }

  .header h2 {
    margin: 0;
    font-size: 20px;
    font-weight: 600;
  }

  .btn-primary {
    background: var(--accent);
    color: white;
    border: none;
    padding: 8px 16px;
    border-radius: var(--radius-md);
    cursor: pointer;
    font-size: 13px;
    font-weight: 500;
    transition: all 0.15s ease;
  }

  .btn-primary:hover {
    background: var(--accent-hover);
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

  .btn-small {
    padding: 4px 10px;
    border-radius: var(--radius-sm);
    border: none;
    cursor: pointer;
    font-size: 12px;
    transition: all 0.15s ease;
  }

  .btn-outline {
    background: transparent;
    border: 1px solid var(--border-color);
    color: var(--text-primary);
  }

  .btn-outline:hover {
    background: var(--hover-bg);
  }

  .error-banner {
    background: rgba(255, 59, 48, 0.1);
    color: var(--danger);
    padding: 8px 12px;
    margin: 12px 24px 0;
    border-radius: var(--radius-md);
    font-size: 13px;
    border: 1px solid rgba(255, 59, 48, 0.2);
  }

  .loading {
    text-align: center;
    color: var(--text-tertiary);
    padding: 40px;
  }

  .split-panel {
    display: flex;
    flex: 1;
    overflow: hidden;
    padding: 16px 24px 24px;
    gap: 16px;
  }

  .list-panel {
    width: 280px;
    flex-shrink: 0;
    overflow-y: auto;
  }

  .list-panel h3 {
    margin: 0 0 12px 0;
    font-size: 14px;
    font-weight: 600;
    color: var(--text-secondary);
  }

  .persona-card {
    padding: 12px;
    border-radius: var(--radius-lg);
    margin-bottom: 8px;
    cursor: pointer;
    border: 1px solid transparent;
    transition: all 0.15s;
    background: var(--card-bg);
    backdrop-filter: blur(10px);
    -webkit-backdrop-filter: blur(10px);
  }

  .persona-card:hover {
    border-color: var(--border-color);
  }

  .persona-card.active {
    background: rgba(0, 113, 227, 0.1);
    border-color: var(--accent);
  }

  .persona-card.selected {
    border-color: var(--border-color);
  }

  .persona-name {
    font-size: 14px;
    font-weight: 500;
    margin-bottom: 4px;
    display: flex;
    align-items: center;
    gap: 6px;
  }

  .active-badge {
    font-size: 10px;
    background: var(--success);
    color: white;
    padding: 1px 6px;
    border-radius: var(--radius-sm);
    font-weight: 600;
  }

  .persona-tags {
    display: flex;
    gap: 4px;
    margin-bottom: 6px;
  }

  .tag {
    font-size: 10px;
    background: var(--bg-tertiary);
    padding: 2px 6px;
    border-radius: var(--radius-sm);
    color: var(--text-tertiary);
  }

  .editor-panel {
    flex: 1;
    overflow-y: auto;
  }

  .editor-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 16px;
  }

  .editor-header h3 {
    margin: 0;
    font-size: 18px;
    font-weight: 600;
  }

  .model-config {
    background: var(--card-bg);
    border: 1px solid var(--border-light);
    border-radius: var(--radius-lg);
    padding: 16px;
    margin-bottom: 16px;
    backdrop-filter: blur(10px);
    -webkit-backdrop-filter: blur(10px);
  }

  .model-config h4 {
    margin: 0 0 12px 0;
    font-size: 14px;
    font-weight: 600;
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
    margin-bottom: 8px;
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

  .markdown-editor h4 {
    margin: 0 0 8px 0;
    font-size: 14px;
    font-weight: 600;
  }

  .textarea {
    width: 100%;
    padding: 12px;
    border: 1px solid var(--border-color);
    border-radius: var(--radius-md);
    font-size: 13px;
    font-family: var(--font-mono);
    box-sizing: border-box;
    resize: vertical;
    background: var(--input-bg);
    color: var(--text-primary);
    outline: none;
    transition: border-color 0.15s;
  }

  .textarea:focus {
    border-color: var(--accent);
    box-shadow: 0 0 0 3px rgba(0, 113, 227, 0.15);
  }

  .preview {
    background: var(--card-bg);
    border: 1px solid var(--border-light);
    border-radius: var(--radius-md);
    padding: 12px;
    font-size: 13px;
    line-height: 1.5;
    backdrop-filter: blur(10px);
    -webkit-backdrop-filter: blur(10px);
  }

  .no-selection {
    display: flex;
    align-items: center;
    justify-content: center;
    height: 200px;
    color: var(--text-tertiary);
  }

  .dialog-overlay {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background: rgba(0, 0, 0, 0.4);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 1000;
    backdrop-filter: blur(4px);
    -webkit-backdrop-filter: blur(4px);
  }

  .dialog-close-bg {
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    border: none;
    background: transparent;
    cursor: default;
  }

  .dialog {
    background: var(--bg-primary);
    border-radius: var(--radius-lg);
    padding: 24px;
    min-width: 400px;
    max-width: 500px;
    box-shadow: var(--shadow-lg);
    position: relative;
    backdrop-filter: blur(20px) saturate(180%);
    -webkit-backdrop-filter: blur(20px) saturate(180%);
    border: 1px solid var(--border-color);
  }

  .dialog h3 {
    margin: 0 0 16px 0;
    font-size: 17px;
    font-weight: 600;
  }

  .dialog-actions {
    display: flex;
    justify-content: flex-end;
    gap: 8px;
    margin-top: 16px;
  }
</style>
