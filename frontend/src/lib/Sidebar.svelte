<script>
  import { activeTab } from '../lib/stores.js'
  import {
    MessageSquare,
    Brain,
    ListTodo,
    User,
    Activity,
    Settings as SettingsIcon,
  } from 'lucide-svelte'

  const tabs = [
    { id: 'chat', icon: MessageSquare, label: 'Chat' },
    { id: 'memory', icon: Brain, label: 'Memory' },
    { id: 'tasks', icon: ListTodo, label: 'Tasks' },
    { id: 'personas', icon: User, label: 'Personas' },
    { id: 'activity', icon: Activity, label: 'Activity' },
    { id: 'settings', icon: SettingsIcon, label: 'Settings' },
  ]
</script>

<nav class="sidebar" aria-label="Main navigation">
  <div class="sidebar-inner">
    <div class="logo-area">
      <div class="logo">R</div>
    </div>
    <div class="tabs-group">
      {#each tabs as tab}
        <button
          class="tab-btn"
          class:active={$activeTab === tab.id}
          on:click={() => activeTab.set(tab.id)}
          title={tab.label}
          aria-label={tab.label}
          aria-current={$activeTab === tab.id ? 'page' : undefined}
        >
          <svelte:component this={tab.icon} size={18} strokeWidth={1.5} />
          <span class="label">{tab.label}</span>
        </button>
      {/each}
    </div>
  </div>
</nav>

<style>
  .sidebar {
    display: flex;
    flex-direction: column;
    width: 64px;
    background: var(--bg-secondary);
    border-right: 1px solid var(--border-color);
    backdrop-filter: blur(30px) saturate(200%);
    -webkit-backdrop-filter: blur(30px) saturate(200%);
    padding: 0;
    user-select: none;
  }

  .sidebar-inner {
    display: flex;
    flex-direction: column;
    align-items: center;
    height: 100%;
    padding: 16px 0;
    gap: 0;
  }

  .logo-area {
    padding: 0 0 16px 0;
    margin-bottom: 8px;
    border-bottom: 1px solid var(--border-light);
    width: 100%;
    display: flex;
    justify-content: center;
  }

  .logo {
    width: 32px;
    height: 32px;
    border-radius: 8px;
    background: linear-gradient(135deg, var(--accent), #5856d6);
    color: white;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 14px;
    font-weight: 700;
    font-family: 'Inter', -apple-system, BlinkMacSystemFont, sans-serif;
    box-shadow: 0 2px 8px rgba(0, 113, 227, 0.3);
  }

  .tabs-group {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 2px;
    flex: 1;
  }

  .tab-btn {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    width: 48px;
    height: 44px;
    border: none;
    background: transparent;
    border-radius: 10px;
    cursor: pointer;
    gap: 1px;
    transition: all 0.2s cubic-bezier(0.16, 1, 0.3, 1);
    position: relative;
    color: var(--text-tertiary);
  }

  .tab-btn:hover {
    background: var(--hover-bg);
    color: var(--text-primary);
  }

  .tab-btn.active {
    background: var(--accent);
    color: white;
    box-shadow: 0 2px 8px rgba(0, 113, 227, 0.25);
  }

  .tab-btn.active::after {
    content: '';
    position: absolute;
    left: -8px;
    top: 50%;
    transform: translateY(-50%);
    width: 3px;
    height: 20px;
    background: var(--accent);
    border-radius: 0 3px 3px 0;
  }

  .label {
    font-size: 8px;
    font-weight: 500;
    color: inherit;
    line-height: 1;
    letter-spacing: 0.2px;
  }
</style>
