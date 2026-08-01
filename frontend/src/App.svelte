<script>
  import Sidebar from './lib/Sidebar.svelte'
  import Chat from './lib/Chat.svelte'
  import ConversationList from './lib/ConversationList.svelte'
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
      <div class="placeholder-tab">
        <h2>Memory Explorer</h2>
        <p>Coming soon</p>
      </div>
    {:else if $activeTab === 'tasks'}
      <div class="placeholder-tab">
        <h2>Tasks & Schedule</h2>
        <p>Coming soon</p>
      </div>
    {:else if $activeTab === 'personas'}
      <div class="placeholder-tab">
        <h2>Persona Studio</h2>
        <p>Coming soon</p>
      </div>
    {:else if $activeTab === 'activity'}
      <div class="placeholder-tab">
        <h2>Activity Log</h2>
        <p>Coming soon</p>
      </div>
    {:else if $activeTab === 'settings'}
      <div class="placeholder-tab">
        <h2>Settings</h2>
        <p>Coming soon</p>
      </div>
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

  .placeholder-tab {
    flex: 1;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    color: #86868b;
  }

  .placeholder-tab h2 {
    margin: 0 0 8px 0;
    font-size: 20px;
    font-weight: 600;
  }

  .placeholder-tab p {
    margin: 0;
    font-size: 14px;
  }
</style>
