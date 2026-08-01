<script>
  import Sidebar from './lib/Sidebar.svelte'
  import Chat from './lib/Chat.svelte'
  import ConversationList from './lib/ConversationList.svelte'
  import MemoryExplorer from './lib/MemoryExplorer.svelte'
  import TaskManager from './lib/TaskManager.svelte'
  import PersonaStudio from './lib/PersonaStudio.svelte'
  import ActivityLog from './lib/ActivityLog.svelte'
  import Settings from './lib/Settings.svelte'
  import { activeTab } from './lib/stores.js'
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

<style>
  .app {
    display: flex;
    height: 100vh;
    font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
    background: white;
    color: #1d1d1f;
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
