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

<nav class="sidebar">
  {#each tabs as tab}
    <button
      class="tab-btn"
      class:active={$activeTab === tab.id}
      on:click={() => activeTab.set(tab.id)}
      title={tab.label}
    >
      <span class="icon">{tab.icon}</span>
    </button>
  {/each}
</nav>

<style>
  .sidebar {
    display: flex;
    flex-direction: column;
    align-items: center;
    width: 48px;
    background: #f5f5f7;
    border-right: 1px solid #e0e0e0;
    padding: 8px 0;
    gap: 4px;
  }

  .tab-btn {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 36px;
    height: 36px;
    border: none;
    background: transparent;
    border-radius: 8px;
    cursor: pointer;
    transition:
      background 0.15s,
      opacity 0.15s;
    opacity: 0.5;
  }

  .tab-btn:hover {
    background: #e8e8ed;
    opacity: 0.8;
  }

  .tab-btn.active {
    background: #e8e8ed;
    opacity: 1;
  }

  .icon {
    font-size: 18px;
    line-height: 1;
  }
</style>
