<script>
  /* eslint-disable svelte/no-at-html-tags */
  import { activeTab } from '../lib/stores.js'

  const tabs = [
    {
      id: 'chat',
      label: 'Chat',
      icon: `<svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"/></svg>`,
    },
    {
      id: 'memory',
      label: 'Memory',
      icon: `<svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M22 12h-4l-3 9L9 3l-3 9H2"/></svg>`,
    },
    {
      id: 'tasks',
      label: 'Tasks',
      icon: `<svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M9 11l3 3L22 4"/><path d="M21 12v7a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h11"/></svg>`,
    },
    {
      id: 'personas',
      label: 'Personas',
      icon: `<svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"/><circle cx="12" cy="7" r="4"/></svg>`,
    },
    {
      id: 'activity',
      label: 'Activity',
      icon: `<svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="22 12 18 12 15 21 9 3 6 12 2 12"/></svg>`,
    },
    {
      id: 'settings',
      label: 'Settings',
      icon: `<svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="3"/><path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1 0 2.83 2 2 0 0 1-2.83 0l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-2 2 2 2 0 0 1-2-2v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83 0 2 2 0 0 1 0-2.83l.06-.06A1.65 1.65 0 0 0 4.68 15a1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1-2-2 2 2 0 0 1 2-2h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 0-2.83 2 2 0 0 1 2.83 0l.06.06A1.65 1.65 0 0 0 9 4.68a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 2-2 2 2 0 0 1 2 2v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 0 2 2 0 0 1 0 2.83l-.06.06A1.65 1.65 0 0 0 19.4 9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 2 2 2 2 0 0 1-2 2h-.09a1.65 1.65 0 0 0-1.51 1z"/></svg>`,
    },
  ]
</script>

<nav class="sidebar" aria-label="Main navigation">
  <div class="sidebar-top">
    {#each tabs as tab}
      <button
        class="tab-btn"
        class:active={$activeTab === tab.id}
        on:click={() => activeTab.set(tab.id)}
        title={tab.label}
        aria-label={tab.label}
        aria-current={$activeTab === tab.id ? 'page' : undefined}
      >
        <span class="icon" aria-hidden="true">{@html tab.icon}</span>
      </button>
    {/each}
  </div>
  <div class="sidebar-bottom">
    <div class="logo-mark" title="Remy">R</div>
  </div>
</nav>

<style>
  .sidebar {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: space-between;
    width: 48px;
    background: var(--sidebar-bg);
    border-right: 1px solid var(--border-light);
    padding: 8px 0;
    flex-shrink: 0;
  }

  .sidebar-top {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 2px;
  }

  .sidebar-bottom {
    display: flex;
    flex-direction: column;
    align-items: center;
  }

  .logo-mark {
    width: 28px;
    height: 28px;
    border-radius: 8px;
    background: var(--accent);
    color: white;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 13px;
    font-weight: 700;
    opacity: 0.8;
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
    transition: all 0.15s ease;
    color: var(--sidebar-icon);
  }

  .tab-btn:hover {
    background: var(--sidebar-hover);
    color: var(--text-primary);
  }

  .tab-btn.active {
    background: var(--sidebar-active);
    color: var(--sidebar-icon-active);
  }

  .icon {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 20px;
    height: 20px;
  }

  .icon :global(svg) {
    display: block;
  }
</style>
