<script>
  import { memorySubTab, searchType, searchResults } from '../lib/stores.js'
  import { searchMemory } from '../lib/wails.js'
  import FactList from './FactList.svelte'
  import EpisodeTimeline from './EpisodeTimeline.svelte'
  import EntityGraph from './EntityGraph.svelte'
  import ScratchpadViewer from './ScratchpadViewer.svelte'

  const subTabs = [
    { id: 'facts', label: 'Facts' },
    { id: 'episodes', label: 'Episodes' },
    { id: 'entities', label: 'Entities' },
    { id: 'scratchpad', label: 'Scratchpad' },
  ]

  let searchQuery = ''

  async function handleSearch() {
    if (!searchQuery.trim()) {
      searchResults.set(null)
      return
    }
    try {
      const results = await searchMemory(searchQuery.trim(), $searchType)
      searchResults.set(results)
    } catch (err) {
      console.error('Search failed:', err)
    }
  }

  function clearSearch() {
    searchQuery = ''
    searchResults.set(null)
  }
</script>

<div class="memory-explorer">
  <div class="search-bar">
    <div class="search-input-wrapper">
      <span class="search-icon">🔍</span>
      <input
        type="text"
        bind:value={searchQuery}
        placeholder="Search memory..."
        on:input={handleSearch}
      />
      {#if searchQuery}
        <button class="clear-btn" on:click={clearSearch}>✕</button>
      {/if}
    </div>
    <div class="search-type-toggle">
      <button
        class="toggle-btn"
        class:active={$searchType === 'fulltext'}
        on:click={() => searchType.set('fulltext')}
      >
        Full Text
      </button>
      <button
        class="toggle-btn"
        class:active={$searchType === 'semantic'}
        on:click={() => searchType.set('semantic')}
      >
        Semantic
      </button>
    </div>
  </div>

  {#if $searchResults}
    <div class="search-results">
      <div class="results-header">
        <h3>Search Results</h3>
        <button class="close-results" on:click={clearSearch}>✕</button>
      </div>
      {#if $searchResults.facts && $searchResults.facts.length > 0}
        <div class="result-section">
          <h4>Facts ({$searchResults.facts.length})</h4>
          {#each $searchResults.facts as fact}
            <div class="result-item">
              <p class="result-text">{fact.fact}</p>
              <span class="result-category">{fact.category}</span>
            </div>
          {/each}
        </div>
      {/if}
      {#if $searchResults.episodes && $searchResults.episodes.length > 0}
        <div class="result-section">
          <h4>Episodes ({$searchResults.episodes.length})</h4>
          {#each $searchResults.episodes as ep}
            <div class="result-item">
              <p class="result-text">{ep.summary}</p>
              <span class="result-importance">Importance: {ep.importance.toFixed(2)}</span>
            </div>
          {/each}
        </div>
      {/if}
      {#if (!$searchResults.facts || $searchResults.facts.length === 0) && (!$searchResults.episodes || $searchResults.episodes.length === 0)}
        <p class="no-results">No results found</p>
      {/if}
    </div>
  {:else}
    <nav class="sub-nav">
      {#each subTabs as tab}
        <button
          class="sub-tab"
          class:active={$memorySubTab === tab.id}
          on:click={() => memorySubTab.set(tab.id)}
        >
          {tab.label}
        </button>
      {/each}
    </nav>

    <div class="sub-content">
      {#if $memorySubTab === 'facts'}
        <FactList />
      {:else if $memorySubTab === 'episodes'}
        <EpisodeTimeline />
      {:else if $memorySubTab === 'entities'}
        <EntityGraph />
      {:else if $memorySubTab === 'scratchpad'}
        <ScratchpadViewer />
      {/if}
    </div>
  {/if}
</div>

<style>
  .memory-explorer {
    display: flex;
    flex-direction: column;
    height: 100%;
    overflow: hidden;
  }

  .search-bar {
    display: flex;
    gap: 8px;
    padding: 12px 16px;
    border-bottom: 1px solid var(--border-light);
    background: var(--bg-primary);
    align-items: center;
  }

  .search-input-wrapper {
    flex: 1;
    display: flex;
    align-items: center;
    background: var(--bg-secondary);
    border-radius: 8px;
    padding: 0 12px;
    border: 1px solid var(--border-light);
    transition: border-color 0.15s ease;
  }

  .search-input-wrapper:focus-within {
    border-color: var(--accent);
    box-shadow: 0 0 0 3px var(--focus-ring);
  }

  .search-icon {
    font-size: 14px;
    margin-right: 8px;
    opacity: 0.5;
  }

  .search-input-wrapper input {
    flex: 1;
    border: none;
    background: transparent;
    padding: 8px 0;
    font-size: 14px;
    outline: none;
    font-family: inherit;
    color: var(--text-primary);
  }

  .search-input-wrapper input::placeholder {
    color: var(--text-tertiary);
  }

  .clear-btn {
    border: none;
    background: transparent;
    cursor: pointer;
    font-size: 14px;
    opacity: 0.5;
    padding: 4px;
    color: var(--text-secondary);
  }

  .clear-btn:hover {
    opacity: 1;
  }

  .search-type-toggle {
    display: flex;
    gap: 0;
    border: 1px solid var(--border-light);
    border-radius: 8px;
    overflow: hidden;
    flex-shrink: 0;
  }

  .toggle-btn {
    border: none;
    background: var(--bg-primary);
    padding: 6px 12px;
    font-size: 12px;
    cursor: pointer;
    font-family: inherit;
    transition: all 0.15s ease;
    color: var(--text-secondary);
  }

  .toggle-btn.active {
    background: var(--accent);
    color: var(--text-inverse);
  }

  .toggle-btn:not(.active):hover {
    background: var(--hover-bg);
  }

  .sub-nav {
    display: flex;
    gap: 0;
    border-bottom: 1px solid var(--border-light);
    padding: 0 16px;
    background: var(--bg-secondary);
    flex-shrink: 0;
  }

  .sub-tab {
    border: none;
    background: transparent;
    padding: 10px 16px;
    font-size: 13px;
    font-weight: 500;
    cursor: pointer;
    color: var(--text-tertiary);
    border-bottom: 2px solid transparent;
    font-family: inherit;
    transition: all 0.15s ease;
  }

  .sub-tab:hover {
    color: var(--text-primary);
  }

  .sub-tab.active {
    color: var(--accent);
    border-bottom-color: var(--accent);
  }

  .sub-content {
    flex: 1;
    overflow-y: auto;
  }

  .search-results {
    flex: 1;
    overflow-y: auto;
    padding: 16px;
  }

  .results-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 16px;
  }

  .results-header h3 {
    margin: 0;
    font-size: 16px;
    font-weight: 600;
  }

  .close-results {
    border: none;
    background: transparent;
    cursor: pointer;
    font-size: 16px;
    opacity: 0.5;
    padding: 4px 8px;
    color: var(--text-secondary);
  }

  .close-results:hover {
    opacity: 1;
  }

  .result-section {
    margin-bottom: 20px;
  }

  .result-section h4 {
    margin: 0 0 8px 0;
    font-size: 14px;
    font-weight: 600;
    color: var(--text-tertiary);
  }

  .result-item {
    padding: 10px 12px;
    background: var(--card-bg);
    border-radius: 8px;
    margin-bottom: 8px;
  }

  .result-text {
    margin: 0 0 4px 0;
    font-size: 14px;
    line-height: 1.4;
    color: var(--text-primary);
  }

  .result-category,
  .result-importance {
    font-size: 11px;
    color: var(--text-tertiary);
    background: var(--tag-bg);
    padding: 2px 6px;
    border-radius: 4px;
  }

  .no-results {
    text-align: center;
    color: var(--text-tertiary);
    font-size: 14px;
    padding: 40px 0;
  }
</style>
