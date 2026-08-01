<script>
  import { onMount } from 'svelte'
  import Sidebar from './lib/Sidebar.svelte'
  import Chat from './lib/Chat.svelte'
  import ConversationList from './lib/ConversationList.svelte'
  import MemoryExplorer from './lib/MemoryExplorer.svelte'
  import TaskManager from './lib/TaskManager.svelte'
  import PersonaStudio from './lib/PersonaStudio.svelte'
  import ActivityLog from './lib/ActivityLog.svelte'
  import Settings from './lib/Settings.svelte'
  import Toast from './lib/Toast.svelte'
  import { activeTab, darkMode } from './lib/stores.js'

  onMount(() => {
    // Apply initial dark mode
    if ($darkMode) {
      document.documentElement.classList.add('dark')
    }

    // Global keyboard shortcuts
    function handleKeydown(e) {
      // Cmd+K or Ctrl+K: Focus search (when on memory tab)
      if ((e.metaKey || e.ctrlKey) && e.key === 'k') {
        e.preventDefault()
        // Dispatch a custom event that memory explorer can listen for
        window.dispatchEvent(new CustomEvent('focus-search'))
      }

      // Escape: Close modals / deselect
      if (e.key === 'Escape') {
        window.dispatchEvent(new CustomEvent('escape-pressed'))
      }
    }

    window.addEventListener('keydown', handleKeydown)
    return () => window.removeEventListener('keydown', handleKeydown)
  })
</script>

<div class="app">
  <Sidebar />
  <main class="content">
    {#if $activeTab === 'chat'}
      <div class="chat-view">
        <ConversationList />
        <Chat />
      </div>
    {:else if $activeTab === 'memory'}
      <MemoryExplorer />
    {:else if $activeTab === 'tasks'}
      <TaskManager />
    {:else if $activeTab === 'personas'}
      <PersonaStudio />
    {:else if $activeTab === 'activity'}
      <ActivityLog />
    {:else if $activeTab === 'settings'}
      <Settings />
    {/if}
  </main>
</div>

<Toast />

<style>
  :global(*) {
    box-sizing: border-box;
  }

  :global(body) {
    margin: 0;
    padding: 0;
    font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
    background: var(--bg-primary);
    color: var(--text-primary);
    transition:
      background-color 0.2s,
      color 0.2s;
  }

  :global(:root) {
    --bg-primary: #ffffff;
    --bg-secondary: #f5f5f7;
    --bg-tertiary: #e8e8ed;
    --text-primary: #1d1d1f;
    --text-secondary: #6e6e73;
    --text-tertiary: #86868b;
    --border-color: #d2d2d7;
    --border-light: #e8e8ed;
    --accent: #007aff;
    --accent-hover: #0056cc;
    --danger: #ff3b30;
    --danger-hover: #d62d20;
    --success: #34c759;
    --warning: #ff9500;
    --shadow: rgba(0, 0, 0, 0.1);
    --input-bg: #ffffff;
    --card-bg: #f5f5f7;
    --hover-bg: #f0f0f5;
  }

  :global(.dark) {
    --bg-primary: #1c1c1e;
    --bg-secondary: #2c2c2e;
    --bg-tertiary: #3a3a3c;
    --text-primary: #f5f5f7;
    --text-secondary: #a1a1a6;
    --text-tertiary: #6e6e73;
    --border-color: #48484a;
    --border-light: #38383a;
    --accent: #0a84ff;
    --accent-hover: #409cff;
    --danger: #ff453a;
    --danger-hover: #ff6b5e;
    --success: #30d158;
    --warning: #ff9f0a;
    --shadow: rgba(0, 0, 0, 0.3);
    --input-bg: #2c2c2e;
    --card-bg: #2c2c2e;
    --hover-bg: #3a3a3c;
  }

  .app {
    display: flex;
    height: 100vh;
    background: var(--bg-primary);
    color: var(--text-primary);
  }

  .content {
    flex: 1;
    display: flex;
    overflow: hidden;
  }

  .chat-view {
    display: flex;
    flex: 1;
    overflow: hidden;
  }
</style>
