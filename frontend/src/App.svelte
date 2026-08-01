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
      background-color 0.3s ease,
      color 0.3s ease;
  }

  :global(::-webkit-scrollbar) {
    width: 6px;
    height: 6px;
  }

  :global(::-webkit-scrollbar-track) {
    background: var(--scrollbar-track);
  }

  :global(::-webkit-scrollbar-thumb) {
    background: var(--scrollbar-thumb);
    border-radius: 3px;
  }

  :global(::-webkit-scrollbar-thumb:hover) {
    background: var(--text-tertiary);
  }

  :global(*:focus-visible) {
    outline: 2px solid var(--accent);
    outline-offset: 2px;
  }

  :global(:root) {
    --bg-primary: #ffffff;
    --bg-secondary: #f5f5f7;
    --bg-tertiary: #e8e8ed;
    --bg-card: #ffffff;
    --bg-card-hover: #fafafa;
    --bg-elevated: #ffffff;
    --text-primary: #1d1d1f;
    --text-secondary: #6e6e73;
    --text-tertiary: #86868b;
    --text-inverse: #ffffff;
    --border-color: #d2d2d7;
    --border-light: #e8e8ed;
    --border-subtle: #f0f0f5;
    --accent: #007aff;
    --accent-hover: #0056cc;
    --accent-subtle: #e8f0fe;
    --accent-text: #007aff;
    --danger: #ff3b30;
    --danger-hover: #d62d20;
    --danger-subtle: #fff2f0;
    --danger-text: #cf1322;
    --success: #34c759;
    --success-subtle: #f0fff0;
    --success-text: #1a7d1a;
    --warning: #ff9500;
    --warning-subtle: #fffbe6;
    --warning-text: #d48806;
    --shadow-sm: rgba(0, 0, 0, 0.06);
    --shadow-md: rgba(0, 0, 0, 0.1);
    --shadow-lg: rgba(0, 0, 0, 0.15);
    --input-bg: #ffffff;
    --input-border: #d2d2d7;
    --card-bg: #f5f5f7;
    --hover-bg: #f0f0f5;
    --sidebar-bg: #f5f5f7;
    --sidebar-hover: #e8e8ed;
    --sidebar-active: #dcdce0;
    --sidebar-icon: #6e6e73;
    --sidebar-icon-active: #007aff;
    --scrollbar-thumb: #c8c8cc;
    --scrollbar-track: transparent;
    --overlay-bg: rgba(0, 0, 0, 0.4);
    --code-bg: #f5f5f7;
    --code-border: #e8e8ed;
    --tag-bg: #e8e8ed;
    --tag-text: #6e6e73;
    --badge-bg: #e8f0fe;
    --badge-text: #1967d2;
    --timeline-line: #e0e0e0;
    --graph-bg: #fafafa;
    --graph-border: #e0e0e0;
    --graph-edge: #c0c0c5;
    --skeleton-bg: #e8e8ed;
    --focus-ring: rgba(0, 122, 255, 0.3);
  }

  :global(.dark) {
    --bg-primary: #1c1c1e;
    --bg-secondary: #2c2c2e;
    --bg-tertiary: #3a3a3c;
    --bg-card: #2c2c2e;
    --bg-card-hover: #353537;
    --bg-elevated: #3a3a3c;
    --text-primary: #f5f5f7;
    --text-secondary: #a1a1a6;
    --text-tertiary: #6e6e73;
    --text-inverse: #ffffff;
    --border-color: #48484a;
    --border-light: #38383a;
    --border-subtle: #333336;
    --accent: #0a84ff;
    --accent-hover: #409cff;
    --accent-subtle: #1a2a4a;
    --accent-text: #64b5f6;
    --danger: #ff453a;
    --danger-hover: #ff6b5e;
    --danger-subtle: #2a1215;
    --danger-text: #ff7875;
    --success: #30d158;
    --success-subtle: #162312;
    --success-text: #95de64;
    --warning: #ff9f0a;
    --warning-subtle: #2b1d0e;
    --warning-text: #ffd666;
    --shadow-sm: rgba(0, 0, 0, 0.2);
    --shadow-md: rgba(0, 0, 0, 0.3);
    --shadow-lg: rgba(0, 0, 0, 0.4);
    --input-bg: #2c2c2e;
    --input-border: #48484a;
    --card-bg: #2c2c2e;
    --hover-bg: #3a3a3c;
    --sidebar-bg: #2c2c2e;
    --sidebar-hover: #3a3a3c;
    --sidebar-active: #48484a;
    --sidebar-icon: #6e6e73;
    --sidebar-icon-active: #0a84ff;
    --scrollbar-thumb: #48484a;
    --scrollbar-track: transparent;
    --overlay-bg: rgba(0, 0, 0, 0.6);
    --code-bg: #2c2c2e;
    --code-border: #38383a;
    --tag-bg: #3a3a3c;
    --tag-text: #a1a1a6;
    --badge-bg: #1a2a4a;
    --badge-text: #64b5f6;
    --timeline-line: #38383a;
    --graph-bg: #2c2c2e;
    --graph-border: #38383a;
    --graph-edge: #48484a;
    --skeleton-bg: #3a3a3c;
    --focus-ring: rgba(10, 132, 255, 0.3);
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
