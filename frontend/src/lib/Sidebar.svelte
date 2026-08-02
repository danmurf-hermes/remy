<script>
  import { activeTab } from '../lib/stores.js'

  const tabs = [
    { id: 'chat', icon: '💬', label: 'Chat' },
    { id: 'memory', icon: '🧠', label: 'Memory' },
    { id: 'tasks', icon: '📋', label: 'Tasks' },
    { id: 'personas', icon: '👤', label: 'Personas' },
    { id: 'activity', icon: '📜', label: 'Activity' },
    { id: 'settings', icon: '⚙', label: 'Settings' },
  ]
</script>

<nav class="sidebar" aria-label="Main navigation">
  <div class="sidebar-inner">
    {#each tabs as tab}
      <button
        class="tab-btn"
        class:active={$activeTab === tab.id}
        on:click={() => activeTab.set(tab.id)}
        title={tab.label}
        aria-label={tab.label}
        aria-current={$activeTab === tab.id ? 'page' : undefined}
      >
        <span class="icon" aria-hidden="true">{tab.icon}</span>
        <span class="label">{tab.label}</span>
      </button>
    {/each}
  </div>
</nav>

<style>
  .sidebar {
    display: flex;
    flex-direction: column;
    width: 68px;
    background: var(--bg-secondary);
    border-right: 1px solid var(--border-color);
    backdrop-filter: blur(20px) saturate(180%);
    -webkit-backdrop-filter: blur(20px) saturate(180%);
    padding: 12px 0;
  }

  .sidebar-inner {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 2px;
    padding: 0 8px;
  }

  .tab-btn {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    width: 52px;
    height: 48px;
    border: none;
    background: transparent;
    border-radius: var(--radius-md);
    cursor: pointer;
    transition: all 0.15s ease;
    gap: 2px;
  }

  .tab-btn:hover {
    background: var(--hover-bg);
  }

  .tab-btn.active {
    background: var(--accent);
  }

  .icon {
    font-size: 18px;
    line-height: 1;
  }

  .label {
    font-size: 9px;
    font-weight: 500;
    color: var(--text-tertiary);
    line-height: 1;
  }

  .tab-btn.active .label {
    color: white;
  }
</style>
