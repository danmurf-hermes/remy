<script>
  import { onMount } from 'svelte'
  import { episodes } from '../lib/stores.js'
  import { getEpisodes } from '../lib/wails.js'

  let loading = true
  let expandedId = null
  let offset = 0
  const limit = 10
  let hasMore = true
  let error = null

  onMount(async () => {
    await loadEpisodes()
  })

  async function loadEpisodes() {
    loading = true
    try {
      const data = await getEpisodes(limit, offset)
      if (data.length < limit) {
        hasMore = false
      }
      episodes.update((prev) => [...prev, ...data])
    } catch (err) {
      error = err.message || String(err)
    } finally {
      loading = false
    }
  }

  async function loadMore() {
    offset += limit
    await loadEpisodes()
  }

  function toggleExpand(id) {
    expandedId = expandedId === id ? null : id
  }

  function formatDate(ts) {
    if (!ts) {
      return ''
    }
    const d = new Date(ts)
    return d.toLocaleDateString('en-US', {
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    })
  }

  function getImportanceColor(imp) {
    if (imp >= 0.7) {
      return '#34c759'
    }
    if (imp >= 0.4) {
      return '#ff9500'
    }
    return '#86868b'
  }

  function formatDuration(start, end) {
    if (!start || !end) {
      return 'N/A'
    }
    const diff = end - start
    const mins = Math.round(diff / 60000)
    if (mins < 1) {
      return 'Less than a minute'
    }
    if (mins < 60) {
      return `${mins} min`
    }
    const hours = Math.floor(mins / 60)
    const remaining = mins % 60
    if (remaining === 0) {
      return `${hours}h`
    }
    return `${hours}h ${remaining}m`
  }
</script>

<div class="episode-timeline">
  {#if error}
    <div class="error-banner">{error}</div>
  {/if}

  {#if $episodes.length === 0 && !loading}
    <div class="empty-state">
      <p>No episodes yet. Episodes are created during consolidation.</p>
    </div>
  {:else}
    <div class="timeline">
      {#each $episodes as ep, i (ep.id)}
        <div class="timeline-item">
          <div class="timeline-dot" style="background: {getImportanceColor(ep.importance)};"></div>
          {#if i < $episodes.length - 1}
            <div class="timeline-line"></div>
          {/if}
          <div class="timeline-content" class:expanded={expandedId === ep.id}>
            <div
              class="timeline-header"
              role="button"
              tabindex="0"
              on:click={() => toggleExpand(ep.id)}
              on:keydown={(e) => {
                if (e.key === 'Enter' || e.key === ' ') {
                  e.preventDefault()
                  toggleExpand(ep.id)
                }
              }}
            >
              <div class="timeline-date">{formatDate(ep.end_time || ep.start_time)}</div>
              <div class="timeline-importance">
                <span
                  class="importance-dot"
                  style="background: {getImportanceColor(ep.importance)};"
                ></span>
                {(ep.importance * 100).toFixed(0)}%
              </div>
            </div>
            <p class="timeline-summary">
              {ep.summary}
            </p>
            {#if expandedId === ep.id}
              <div class="timeline-details">
                {#if ep.topics}
                  <div class="detail-row">
                    <span class="detail-label">Topics:</span>
                    <div class="topic-tags">
                      {#each ep.topics.split(',') as topic}
                        <span class="topic-tag">{topic.trim()}</span>
                      {/each}
                    </div>
                  </div>
                {/if}
                {#if ep.message_ids}
                  <div class="detail-row">
                    <span class="detail-label">Messages:</span>
                    <span class="detail-value">{ep.message_ids}</span>
                  </div>
                {/if}
                <div class="detail-row">
                  <span class="detail-label">Duration:</span>
                  <span class="detail-value">{formatDuration(ep.start_time, ep.end_time)}</span>
                </div>
              </div>
            {/if}
          </div>
        </div>
      {/each}
    </div>

    {#if hasMore}
      <div class="load-more">
        <button on:click={loadMore} disabled={loading}>
          {loading ? 'Loading...' : 'Load More'}
        </button>
      </div>
    {/if}
  {/if}
</div>

<style>
  .episode-timeline {
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

  .empty-state {
    text-align: center;
    color: #86868b;
    padding: 40px 20px;
    font-size: 14px;
  }

  .timeline {
    position: relative;
    padding-left: 20px;
  }

  .timeline-item {
    position: relative;
    padding-bottom: 20px;
  }

  .timeline-dot {
    position: absolute;
    left: -16px;
    top: 4px;
    width: 12px;
    height: 12px;
    border-radius: 50%;
    border: 2px solid white;
    box-shadow: 0 0 0 1px #e0e0e0;
    z-index: 1;
  }

  .timeline-line {
    position: absolute;
    left: -11px;
    top: 16px;
    bottom: 0;
    width: 2px;
    background: #e0e0e0;
  }

  .timeline-content {
    background: white;
    border: 1px solid #e0e0e0;
    border-radius: 10px;
    padding: 12px 14px;
    cursor: pointer;
    transition: box-shadow 0.15s;
  }

  .timeline-content:hover {
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
  }

  .timeline-content.expanded {
    border-color: #007aff;
  }

  .timeline-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 6px;
  }

  .timeline-date {
    font-size: 12px;
    color: #86868b;
    font-weight: 500;
  }

  .timeline-importance {
    display: flex;
    align-items: center;
    gap: 4px;
    font-size: 11px;
    color: #86868b;
  }

  .importance-dot {
    width: 6px;
    height: 6px;
    border-radius: 50%;
    display: inline-block;
  }

  .timeline-summary {
    margin: 0;
    font-size: 14px;
    line-height: 1.5;
    color: #1d1d1f;
  }

  .timeline-details {
    margin-top: 10px;
    padding-top: 10px;
    border-top: 1px solid #f0f0f5;
  }

  .detail-row {
    display: flex;
    align-items: flex-start;
    gap: 8px;
    margin-bottom: 6px;
    font-size: 13px;
  }

  .detail-label {
    color: #86868b;
    font-weight: 500;
    flex-shrink: 0;
    min-width: 60px;
  }

  .detail-value {
    color: #1d1d1f;
  }

  .topic-tags {
    display: flex;
    gap: 4px;
    flex-wrap: wrap;
  }

  .topic-tag {
    font-size: 11px;
    background: #e8f0fe;
    color: #1967d2;
    padding: 2px 8px;
    border-radius: 10px;
  }

  .load-more {
    text-align: center;
    padding: 16px;
  }

  .load-more button {
    padding: 8px 24px;
    background: #f0f0f5;
    border: 1px solid #d0d0d5;
    border-radius: 8px;
    font-size: 14px;
    cursor: pointer;
    font-family: inherit;
    color: #1d1d1f;
    transition: background 0.15s;
  }

  .load-more button:hover:not(:disabled) {
    background: #e0e0e5;
  }

  .load-more button:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }
</style>
