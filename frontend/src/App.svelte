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
    font-family: -apple-system, BlinkMacSystemFont, 'SF Pro', 'Helvetica Neue', sans-serif;
    background: var(--bg-primary);
    color: var(--text-primary);
    -webkit-font-smoothing: antialiased;
    -moz-osx-font-smoothing: grayscale;
    transition:
      background-color 0.2s,
      color 0.2s;
  }

  :global(:root) {
    --bg-primary: rgba(255, 255, 255, 0.72);
    --bg-secondary: rgba(246, 246, 248, 0.85);
    --bg-tertiary: rgba(232, 232, 237, 0.8);
    --text-primary: #1d1d1f;
    --text-secondary: #6e6e73;
    --text-tertiary: #86868b;
    --border-color: rgba(0, 0, 0, 0.1);
    --border-light: rgba(0, 0, 0, 0.06);
    --accent: #0071e3;
    --accent-hover: #0077ed;
    --danger: #ff3b30;
    --danger-hover: #d62d20;
    --success: #34c759;
    --warning: #ff9500;
    --shadow-sm: 0 1px 3px rgba(0, 0, 0, 0.08);
    --shadow-md: 0 4px 12px rgba(0, 0, 0, 0.1);
    --shadow-lg: 0 8px 30px rgba(0, 0, 0, 0.12);
    --input-bg: rgba(255, 255, 255, 0.8);
    --card-bg: rgba(255, 255, 255, 0.5);
    --hover-bg: rgba(0, 0, 0, 0.05);
    --frosted: backdrop-filter blur(20px) saturate(180%);
    --radius-sm: 6px;
    --radius-md: 8px;
    --radius-lg: 12px;
    --radius-xl: 20px;
    --font-mono: 'SF Mono', 'JetBrains Mono', 'Cascadia Code', monospace;
  }

  :global(.dark) {
    --bg-primary: rgba(28, 28, 30, 0.85);
    --bg-secondary: rgba(30, 30, 32, 0.9);
    --bg-tertiary: rgba(44, 44, 46, 0.8);
    --text-primary: #f5f5f7;
    --text-secondary: #a1a1a6;
    --text-tertiary: #6e6e73;
    --border-color: rgba(255, 255, 255, 0.1);
    --border-light: rgba(255, 255, 255, 0.06);
    --accent: #0a84ff;
    --accent-hover: #409cff;
    --danger: #ff453a;
    --danger-hover: #ff6b5e;
    --success: #30d158;
    --warning: #ff9f0a;
    --shadow-sm: 0 1px 3px rgba(0, 0, 0, 0.3);
    --shadow-md: 0 4px 12px rgba(0, 0, 0, 0.4);
    --shadow-lg: 0 8px 30px rgba(0, 0, 0, 0.5);
    --input-bg: rgba(44, 44, 46, 0.8);
    --card-bg: rgba(44, 44, 46, 0.6);
    --hover-bg: rgba(255, 255, 255, 0.08);
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
