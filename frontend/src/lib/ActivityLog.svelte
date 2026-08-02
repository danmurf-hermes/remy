<script>
  import { onMount } from 'svelte'
  import { activityLog } from './stores.js'
  import { getActivityLog } from './wails.js'

  let loading = true
  let filter = 'all'
  let searchQuery = ''
  let expandedEntry = null
  let error = null

  const filters = [
    { id: 'all', label: 'All' },
    { id: 'message', label: 'Messages' },
    { id: 'retrieval', label: 'Retrievals' },
    { id: 'function', label: 'Functions' },
    { id: 'llm', label: 'LLM' },
    { id: 'consolidation', label: 'Consolidation' },
    { id: 'error', label: 'Errors' },
  ]

  const typeIcons = {
    message: '💬',
    retrieval: '🔍',
    function: '🔧',
    llm: '🤖',
    consolidation: '🧠',
    error: '⚠️',
  }

  $: filteredLog = $activityLog.filter((entry) => {
    if (filter !== 'all' && entry.type !== filter) {
      return false
    }
    if (searchQuery) {
      const q = searchQuery.toLowerCase()
      return (
        entry.type.toLowerCase().includes(q) ||
        (entry.details || '').toLowerCase().includes(q) ||
        entry.id.toLowerCase().includes(q)
      )
    }
    return true
  })

  onMount(async () => {
    await loadLog()
  })

  async function loadLog() {
    loading = true
    try {
      const result = await getActivityLog('', 50, 0)
      activityLog.set(result)
    } catch (e) {
      error = 'Failed to load activity log: ' + e.message
    }
    loading = false
  }

  function formatTime(ts) {
    if (!ts) {
      return ''
    }
    const d = new Date(ts)
    return d.toLocaleString()
  }

  function toggleExpand(entry) {
    if (expandedEntry?.id === entry.id) {
      expandedEntry = null
    } else {
      expandedEntry = entry
    }
  }

  function formatDetails(details) {
    if (!details) {
      return '{}'
    }
    try {
      return JSON.stringify(JSON.parse(details), null, 2)
    } catch {
      return details
    }
  }
</script>

<div class="activity-log">
  <div class="header">
    <h2>Activity Log</h2>
    <button class="btn-secondary" on:click={loadLog}>🔄 Refresh</button>
  </div>

  {#if error}
    <div class="error-banner">{error}</div>
  {/if}

  <div class="controls">
    <div class="filter-chips">
      {#each filters as f}
        <button class="chip" class:active={filter === f.id} on:click={() => (filter = f.id)}>
          {f.label}
        </button>
      {/each}
    </div>
    <input
      type="text"
      placeholder="Search entries…"
      bind:value={searchQuery}
      class="search-input"
    />
  </div>

  {#if loading}
    <div class="loading">Loading activity log…</div>
  {:else if filteredLog.length === 0}
    <div class="empty">
      <p>No activity entries found</p>
    </div>
  {:else}
    <div class="timeline">
      {#each filteredLog as entry (entry.id)}
        <div
          class="timeline-entry"
          class:expanded={expandedEntry?.id === entry.id}
          class:error-entry={entry.type === 'error'}
        >
          <div
            class="entry-header"
            role="button"
            tabindex="0"
            on:click={() => toggleExpand(entry)}
            on:keydown={(e) => {
              if (e.key === 'Enter' || e.key === ' ') {
                e.preventDefault()
                toggleExpand(entry)
              }
            }}
          >
            <span class="entry-icon">{typeIcons[entry.type] || '📄'}</span>
            <div class="entry-meta">
              <span class="entry-type">{entry.type}</span>
              <span class="entry-time">{formatTime(entry.timestamp)}</span>
            </div>
            <span class="expand-icon">{expandedEntry?.id === entry.id ? '▼' : '▶'}</span>
          </div>
          {#if expandedEntry?.id === entry.id}
            <div class="entry-details">
              <div class="detail-row">
                <span class="detail-label">ID:</span>
                <span class="detail-value">{entry.id}</span>
              </div>
              <div class="detail-row">
                <span class="detail-label">Session:</span>
                <span class="detail-value">{entry.session_id}</span>
              </div>
              {#if entry.message_id}
                <div class="detail-row">
                  <span class="detail-label">Message:</span>
                  <span class="detail-value">{entry.message_id}</span>
                </div>
              {/if}
              <div class="detail-json">
                <pre>{formatDetails(entry.details)}</pre>
              </div>
            </div>
          {/if}
        </div>
      {/each}
    </div>
  {/if}
</div>

<style>
  .activity-log {
    flex: 1;
    padding: 24px;
    overflow-y: auto;
  }

  .header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 16px;
  }

  .header h2 {
    margin: 0;
    font-size: 20px;
    font-weight: 600;
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

  .error-banner {
    background: rgba(255, 59, 48, 0.1);
    color: var(--danger);
    padding: 8px 12px;
    border-radius: var(--radius-md);
    margin-bottom: 12px;
    font-size: 13px;
    border: 1px solid rgba(255, 59, 48, 0.2);
  }

  .controls {
    display: flex;
    flex-direction: column;
    gap: 12px;
    margin-bottom: 16px;
  }

  .filter-chips {
    display: flex;
    flex-wrap: wrap;
    gap: 6px;
  }

  .chip {
    padding: 4px 12px;
    border-radius: var(--radius-xl);
    border: 1px solid var(--border-color);
    background: transparent;
    font-size: 12px;
    cursor: pointer;
    color: var(--text-secondary);
    transition: all 0.15s;
  }

  .chip:hover {
    background: var(--hover-bg);
  }

  .chip.active {
    background: var(--accent);
    color: white;
    border-color: var(--accent);
  }

  .search-input {
    width: 100%;
    padding: 8px 12px;
    border: 1px solid var(--border-color);
    border-radius: var(--radius-md);
    font-size: 13px;
    box-sizing: border-box;
    background: var(--input-bg);
    color: var(--text-primary);
    outline: none;
    transition: border-color 0.15s;
  }

  .search-input:focus {
    border-color: var(--accent);
    box-shadow: 0 0 0 3px rgba(0, 113, 227, 0.15);
  }

  .loading {
    text-align: center;
    color: var(--text-tertiary);
    padding: 40px;
  }

  .empty {
    text-align: center;
    color: var(--text-tertiary);
    padding: 40px;
  }

  .timeline {
    display: flex;
    flex-direction: column;
    gap: 4px;
  }

  .timeline-entry {
    border-radius: var(--radius-md);
    border: 1px solid var(--border-light);
    overflow: hidden;
    transition: all 0.15s;
    background: var(--card-bg);
    backdrop-filter: blur(10px);
    -webkit-backdrop-filter: blur(10px);
  }

  .timeline-entry:hover {
    border-color: var(--border-color);
  }

  .timeline-entry.error-entry {
    border-color: rgba(255, 59, 48, 0.3);
    background: rgba(255, 59, 48, 0.05);
  }

  .entry-header {
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 10px 12px;
    cursor: pointer;
  }

  .entry-icon {
    font-size: 16px;
    width: 24px;
    text-align: center;
  }

  .entry-meta {
    flex: 1;
    display: flex;
    flex-direction: column;
    gap: 2px;
  }

  .entry-type {
    font-size: 13px;
    font-weight: 500;
    text-transform: capitalize;
  }

  .entry-time {
    font-size: 11px;
    color: var(--text-tertiary);
  }

  .expand-icon {
    font-size: 10px;
    color: var(--text-tertiary);
  }

  .entry-details {
    padding: 0 12px 12px;
    border-top: 1px solid var(--border-light);
    padding-top: 10px;
  }

  .detail-row {
    display: flex;
    gap: 8px;
    margin-bottom: 4px;
    font-size: 12px;
  }

  .detail-label {
    color: var(--text-tertiary);
    min-width: 60px;
  }

  .detail-value {
    color: var(--text-primary);
    font-family: var(--font-mono);
  }

  .detail-json {
    margin-top: 8px;
  }

  .detail-json pre {
    background: var(--bg-tertiary);
    border-radius: var(--radius-sm);
    padding: 10px;
    font-size: 11px;
    overflow-x: auto;
    white-space: pre-wrap;
    word-break: break-all;
    margin: 0;
    font-family: var(--font-mono);
  }
</style>
