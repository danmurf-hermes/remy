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
      <span class="hint">Auto-saves after 2 seconds of inactivity. Press {navigator.platform.includes('Mac') ? 'Cmd' : 'Ctrl'}+S to save immediately.</span>
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
    background: #fff2f0;
    border: 1px solid #ffccc7;
    border-radius: 6px;
    color: #cf1322;
    font-size: 13px;
    margin-bottom: 12px;
  }

  .loading {
    text-align: center;
    color: #86868b;
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
    color: #34c759;
    font-weight: 500;
  }

  .unsaved-indicator {
    color: #ff9500;
    font-weight: 500;
  }

  .scratchpad-textarea {
    flex: 1;
    width: 100%;
    border: 1px solid #d0d0d5;
    border-radius: 8px;
    padding: 12px;
    font-size: 14px;
    font-family: 'SF Mono', 'Fira Code', 'Consolas', monospace;
    line-height: 1.6;
    resize: none;
    outline: none;
    box-sizing: border-box;
    background: #fafafa;
    color: #1d1d1f;
  }

  .scratchpad-textarea:focus {
    border-color: #007aff;
    background: white;
  }

  .scratchpad-textarea::placeholder {
    color: #c0c0c5;
  }

  .scratchpad-footer {
    margin-top: 8px;
  }

  .hint {
    font-size: 11px;
    color: #86868b;
  }
</style>
