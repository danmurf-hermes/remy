<script>
  import { onMount } from 'svelte'
  import { scratchpad } from '../lib/stores.js'
  import { getScratchpad, updateScratchpad } from '../lib/wails.js'

  let loading = true
  let saved = true
  let saveTimer = null
  let error = null

  onMount(async () => {
    await loadScratchpad()
  })

  async function loadScratchpad() {
    loading = true
    try {
      const content = await getScratchpad()
      scratchpad.set(content || '')
    } catch (err) {
      error = err.message || String(err)
    } finally {
      loading = false
    }
  }

  function handleInput() {
    saved = false
    if (saveTimer) {
      clearTimeout(saveTimer)
    }
    saveTimer = setTimeout(doSave, 2000)
  }

  async function doSave() {
    try {
      await updateScratchpad($scratchpad)
      saved = true
    } catch (err) {
      error = err.message || String(err)
    }
  }

  function handleKeydown(e) {
    if (e.key === 's' && (e.metaKey || e.ctrlKey)) {
      e.preventDefault()
      if (saveTimer) {
        clearTimeout(saveTimer)
      }
      doSave()
    }
  }
</script>

<div class="scratchpad-viewer">
  {#if error}
    <div class="error-banner">{error}</div>
  {/if}

  {#if loading}
    <div class="loading">Loading scratchpad...</div>
  {:else}
    <div class="scratchpad-header">
      <h3>Scratchpad</h3>
      <div class="save-status">
        {#if saved}
          <span class="saved-indicator">✓ Saved</span>
        {:else}
          <span class="unsaved-indicator">Unsaved changes...</span>
        {/if}
      </div>
    </div>
    <textarea
      class="scratchpad-textarea"
      bind:value={$scratchpad}
      on:input={handleInput}
      on:keydown={handleKeydown}
      placeholder="The agent uses this space for working memory. Edit freely — changes are auto-saved."
      rows="20"
    ></textarea>
    <div class="scratchpad-footer">
      <span class="hint"
        >Auto-saves after 2 seconds of inactivity. Press {navigator.platform.includes('Mac')
          ? 'Cmd'
          : 'Ctrl'}+S to save immediately.</span
      >
    </div>
  {/if}
</div>

<style>
  .scratchpad-viewer {
    padding: 16px;
    display: flex;
    flex-direction: column;
    height: 100%;
  }

  .error-banner {
    padding: 8px 12px;
    background: var(--danger-subtle);
    border: 1px solid var(--danger-subtle);
    border-radius: 6px;
    color: var(--danger-text);
    font-size: 13px;
    margin-bottom: 12px;
  }

  .loading {
    text-align: center;
    color: var(--text-tertiary);
    padding: 40px 0;
    font-size: 14px;
  }

  .scratchpad-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 12px;
  }

  .scratchpad-header h3 {
    margin: 0;
    font-size: 16px;
    font-weight: 600;
  }

  .save-status {
    font-size: 12px;
  }

  .saved-indicator {
    color: var(--success);
    font-weight: 500;
  }

  .unsaved-indicator {
    color: var(--warning);
    font-weight: 500;
  }

  .scratchpad-textarea {
    flex: 1;
    width: 100%;
    border: 1px solid var(--input-border);
    border-radius: 8px;
    padding: 12px;
    font-size: 14px;
    font-family: 'SF Mono', 'Fira Code', 'Consolas', monospace;
    line-height: 1.6;
    resize: none;
    outline: none;
    box-sizing: border-box;
    background: var(--input-bg);
    color: var(--text-primary);
  }

  .scratchpad-textarea:focus {
    border-color: var(--accent);
    box-shadow: 0 0 0 3px var(--focus-ring);
  }

  .scratchpad-textarea::placeholder {
    color: var(--text-tertiary);
  }

  .scratchpad-footer {
    margin-top: 8px;
  }

  .hint {
    font-size: 11px;
    color: var(--text-tertiary);
  }
</style>
