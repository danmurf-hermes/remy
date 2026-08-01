<script>
  import { onMount } from 'svelte'
  import { facts } from '../lib/stores.js'
  import { getFacts, updateFact, deleteFact } from '../lib/wails.js'

  let loading = true
  let editingId = null
  let editText = ''
  let editCategory = ''
  let editConfidence = 0
  let deletingId = null
  let error = null

  onMount(async () => {
    await loadFacts()
  })

  async function loadFacts() {
    loading = true
    try {
      const data = await getFacts('')
      facts.set(data)
    } catch (err) {
      error = err.message || String(err)
    } finally {
      loading = false
    }
  }

  $: groupedFacts = groupByCategory($facts)

  function groupByCategory(factList) {
    const groups = {}
    for (const f of factList) {
      const cat = f.category || 'uncategorized'
      if (!groups[cat]) {
        groups[cat] = []
      }
      groups[cat].push(f)
    }
    return groups
  }

  function getConfidenceColor(confidence) {
    if (confidence >= 0.8) { return '#34c759' }
    if (confidence >= 0.5) { return '#ff9500' }
    return '#ff3b30'
  }

  function startEdit(fact) {
    editingId = fact.id
    editText = fact.fact
    editCategory = fact.category
    editConfidence = fact.confidence
  }

  function cancelEdit() {
    editingId = null
  }

  async function saveEdit(fact) {
    try {
      await updateFact(fact.id, editText, editCategory, editConfidence)
      editingId = null
      await loadFacts()
    } catch (err) {
      error = err.message || String(err)
    }
  }

  function confirmDelete(id) {
    deletingId = id
  }

  function cancelDelete() {
    deletingId = null
  }

  async function doDelete(id) {
    try {
      await deleteFact(id)
      deletingId = null
      await loadFacts()
    } catch (err) {
      error = err.message || String(err)
    }
  }
</script>

<div class="fact-list">
  {#if error}
    <div class="error-banner">{error}</div>
  {/if}

  {#if loading}
    <div class="loading">Loading facts...</div>
  {:else if Object.keys(groupedFacts).length === 0}
    <div class="empty-state">
      <p>No facts yet. Facts are extracted from conversations during consolidation.</p>
    </div>
  {:else}
    {#each Object.entries(groupedFacts) as [category, categoryFacts]}
      <div class="category-group">
        <h3 class="category-title">{category}</h3>
        <div class="card-grid">
          {#each categoryFacts as fact (fact.id)}
            <div class="fact-card">
              {#if editingId === fact.id}
                <div class="edit-form">
                  <textarea bind:value={editText} class="edit-textarea" rows="3"></textarea>
                  <div class="edit-row">
                    <label>
                      Category:
                      <input type="text" bind:value={editCategory} class="edit-input" />
                    </label>
                    <label>
                      Confidence:
                      <input
                        type="range"
                        min="0"
                        max="1"
                        step="0.05"
                        bind:value={editConfidence}
                      />
                      <span class="conf-value">{editConfidence.toFixed(2)}</span>
                    </label>
                  </div>
                  <div class="edit-actions">
                    <button class="save-btn" on:click={() => saveEdit(fact)}>Save</button>
                    <button class="cancel-btn" on:click={cancelEdit}>Cancel</button>
                  </div>
                </div>
              {:else if deletingId === fact.id}
                <div class="delete-confirm">
                  <p>Delete this fact?</p>
                  <p class="delete-text">"{fact.fact}"</p>
                  <div class="delete-actions">
                    <button class="delete-yes" on:click={() => doDelete(fact.id)}>Delete</button>
                    <button class="cancel-btn" on:click={cancelDelete}>Cancel</button>
                  </div>
                </div>
              {:else}
                <div class="fact-content">
                  <p class="fact-text">{fact.fact}</p>
                  <div class="fact-meta">
                    <span class="category-badge">{fact.category || 'uncategorized'}</span>
                    <span class="source-label">{fact.source || 'unknown'}</span>
                  </div>
                  <div class="confidence-bar-container">
                    <div class="confidence-bar" style="width: {fact.confidence * 100}%; background: {getConfidenceColor(fact.confidence)};"></div>
                  </div>
                  <div class="confidence-label">
                    Confidence: {(fact.confidence * 100).toFixed(0)}%
                  </div>
                </div>
                <div class="card-actions">
                  <button class="action-btn edit-btn" on:click={() => startEdit(fact)} title="Edit">✏️</button>
                  <button class="action-btn delete-btn" on:click={() => confirmDelete(fact.id)} title="Delete">🗑️</button>
                </div>
              {/if}
            </div>
          {/each}
        </div>
      </div>
    {/each}
  {/if}
</div>

<style>
  .fact-list {
    padding: 16px;
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

  .empty-state {
    text-align: center;
    color: #86868b;
    padding: 40px 20px;
    font-size: 14px;
    line-height: 1.5;
  }

  .category-group {
    margin-bottom: 24px;
  }

  .category-title {
    margin: 0 0 12px 0;
    font-size: 14px;
    font-weight: 600;
    color: #86868b;
    text-transform: uppercase;
    letter-spacing: 0.5px;
  }

  .card-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
    gap: 12px;
  }

  .fact-card {
    background: white;
    border: 1px solid #e0e0e0;
    border-radius: 10px;
    padding: 14px;
    position: relative;
    transition: box-shadow 0.15s;
  }

  .fact-card:hover {
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
  }

  .fact-card:hover .card-actions {
    opacity: 1;
  }

  .fact-content {
    margin-bottom: 8px;
  }

  .fact-text {
    margin: 0 0 8px 0;
    font-size: 14px;
    line-height: 1.5;
    color: #1d1d1f;
  }

  .fact-meta {
    display: flex;
    gap: 6px;
    margin-bottom: 8px;
    flex-wrap: wrap;
  }

  .category-badge {
    font-size: 11px;
    background: #e8f0fe;
    color: #1967d2;
    padding: 2px 8px;
    border-radius: 10px;
    font-weight: 500;
  }

  .source-label {
    font-size: 11px;
    color: #86868b;
    padding: 2px 8px;
    background: #f0f0f5;
    border-radius: 10px;
  }

  .confidence-bar-container {
    height: 6px;
    background: #f0f0f5;
    border-radius: 3px;
    overflow: hidden;
    margin-bottom: 4px;
  }

  .confidence-bar {
    height: 100%;
    border-radius: 3px;
    transition: width 0.3s;
  }

  .confidence-label {
    font-size: 11px;
    color: #86868b;
  }

  .card-actions {
    position: absolute;
    top: 8px;
    right: 8px;
    display: flex;
    gap: 4px;
    opacity: 0;
    transition: opacity 0.15s;
  }

  .action-btn {
    width: 28px;
    height: 28px;
    border: none;
    border-radius: 6px;
    cursor: pointer;
    font-size: 14px;
    display: flex;
    align-items: center;
    justify-content: center;
    background: white;
    border: 1px solid #e0e0e0;
    transition: background 0.15s;
  }

  .action-btn:hover {
    background: #f5f5f7;
  }

  .edit-form {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  .edit-textarea {
    width: 100%;
    border: 1px solid #d0d0d5;
    border-radius: 6px;
    padding: 8px;
    font-size: 14px;
    font-family: inherit;
    resize: vertical;
    box-sizing: border-box;
  }

  .edit-textarea:focus {
    border-color: #007aff;
    outline: none;
  }

  .edit-row {
    display: flex;
    gap: 12px;
    align-items: center;
    flex-wrap: wrap;
  }

  .edit-row label {
    display: flex;
    align-items: center;
    gap: 6px;
    font-size: 12px;
    color: #86868b;
  }

  .edit-input {
    border: 1px solid #d0d0d5;
    border-radius: 4px;
    padding: 4px 8px;
    font-size: 13px;
    width: 120px;
    font-family: inherit;
  }

  .edit-input:focus {
    border-color: #007aff;
    outline: none;
  }

  .conf-value {
    font-size: 12px;
    font-weight: 600;
    color: #1d1d1f;
    min-width: 32px;
  }

  .edit-actions {
    display: flex;
    gap: 6px;
  }

  .save-btn {
    padding: 6px 14px;
    background: #007aff;
    color: white;
    border: none;
    border-radius: 6px;
    font-size: 13px;
    cursor: pointer;
    font-family: inherit;
  }

  .save-btn:hover {
    background: #0056cc;
  }

  .cancel-btn {
    padding: 6px 14px;
    background: #f0f0f5;
    color: #1d1d1f;
    border: 1px solid #d0d0d5;
    border-radius: 6px;
    font-size: 13px;
    cursor: pointer;
    font-family: inherit;
  }

  .cancel-btn:hover {
    background: #e0e0e5;
  }

  .delete-confirm {
    text-align: center;
    padding: 8px 0;
  }

  .delete-confirm p {
    margin: 0 0 8px 0;
    font-size: 14px;
  }

  .delete-text {
    font-style: italic;
    color: #86868b;
    font-size: 13px !important;
  }

  .delete-actions {
    display: flex;
    gap: 6px;
    justify-content: center;
  }

  .delete-yes {
    padding: 6px 14px;
    background: #ff3b30;
    color: white;
    border: none;
    border-radius: 6px;
    font-size: 13px;
    cursor: pointer;
    font-family: inherit;
  }

  .delete-yes:hover {
    background: #d62d20;
  }
</style>
