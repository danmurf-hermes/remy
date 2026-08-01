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
    // In a real implementation, this would call a Go binding to create the persona file
    // For now, we simulate by reloading
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
    <button class="btn-primary" on:click={() => (showCreateDialog = true)}> + New Persona </button>
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
    <div class="loading">Loading personas...</div>
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
              <textarea class="textarea" rows="8" placeholder="Enter system prompt in Markdown...">
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
    color: var(--text-inverse);
    border: none;
    padding: 8px 16px;
    border-radius: 8px;
    cursor: pointer;
    font-size: 13px;
    font-weight: 500;
    transition: background 0.15s ease;
  }

  .btn-primary:hover {
    background: var(--accent-hover);
  }

  .btn-secondary {
    background: var(--tag-bg);
    color: var(--text-primary);
    border: none;
    padding: 8px 16px;
    border-radius: 8px;
    cursor: pointer;
    font-size: 13px;
    transition: background 0.15s ease;
  }

  .btn-small {
    padding: 4px 10px;
    border-radius: 6px;
    border: none;
    cursor: pointer;
    font-size: 12px;
  }

  .btn-outline {
    background: transparent;
    border: 1px solid var(--border-color);
    color: var(--text-primary);
    transition: background 0.15s ease;
  }

  .btn-outline:hover {
    background: var(--hover-bg);
  }

  .error-banner {
    background: var(--danger-subtle);
    color: var(--danger-text);
    padding: 8px 12px;
    margin: 12px 24px 0;
    border-radius: 8px;
    font-size: 13px;
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
    border-radius: 10px;
    margin-bottom: 8px;
    cursor: pointer;
    border: 1px solid transparent;
    transition: all 0.15s ease;
  }

  .persona-card:hover {
    background: var(--hover-bg);
  }

  .persona-card.active {
    background: var(--accent-subtle);
    border-color: var(--accent);
  }

  .persona-card.selected {
    background: var(--hover-bg);
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
    color: var(--text-inverse);
    padding: 1px 6px;
    border-radius: 4px;
    font-weight: 600;
  }

  .persona-tags {
    display: flex;
    gap: 4px;
    margin-bottom: 6px;
  }

  .tag {
    font-size: 10px;
    background: var(--tag-bg);
    padding: 2px 6px;
    border-radius: 4px;
    color: var(--tag-text);
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
    border-radius: 12px;
    padding: 16px;
    margin-bottom: 16px;
  }

  .model-config h4 {
    margin: 0 0 12px 0;
    font-size: 14px;
    font-weight: 600;
  }

  .input {
    width: 100%;
    padding: 8px 12px;
    border: 1px solid var(--input-border);
    border-radius: 8px;
    font-size: 13px;
    box-sizing: border-box;
    margin-bottom: 8px;
    background: var(--input-bg);
    color: var(--text-primary);
    font-family: inherit;
  }

  .input:focus {
    border-color: var(--accent);
    outline: none;
    box-shadow: 0 0 0 3px var(--focus-ring);
  }

  .slider {
    width: 100%;
    margin-bottom: 8px;
    accent-color: var(--accent);
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
    border: 1px solid var(--input-border);
    border-radius: 8px;
    font-size: 13px;
    font-family: 'SF Mono', Monaco, 'Cascadia Code', monospace;
    box-sizing: border-box;
    resize: vertical;
    background: var(--input-bg);
    color: var(--text-primary);
  }

  .textarea:focus {
    border-color: var(--accent);
    outline: none;
    box-shadow: 0 0 0 3px var(--focus-ring);
  }

  .preview {
    background: var(--bg-primary);
    border: 1px solid var(--input-border);
    border-radius: 8px;
    padding: 12px;
    font-size: 13px;
    line-height: 1.5;
    color: var(--text-primary);
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
    background: var(--overlay-bg);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 1000;
  }

  .dialog {
    background: var(--bg-primary);
    border: 1px solid var(--border-color);
    border-radius: 16px;
    padding: 24px;
    width: 480px;
    max-width: 90vw;
    max-height: 85vh;
    overflow-y: auto;
    box-shadow: 0 8px 32px var(--shadow-lg);
  }

  .dialog h3 {
    margin: 0 0 16px 0;
    font-size: 18px;
    font-weight: 600;
  }

  .dialog-actions {
    display: flex;
    gap: 8px;
    justify-content: flex-end;
    margin-top: 16px;
  }
</style>
