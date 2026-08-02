<script>
  import { conversations, activeConversation } from '../lib/stores.js'
  import { getConversations } from '../lib/wails.js'
  import { onMount } from 'svelte'

  onMount(async () => {
    const convs = await getConversations()
    conversations.set(convs)
  })
</script>

<aside class="conversation-list">
  <div class="header">
    <h3>Conversations</h3>
  </div>
  <div class="list">
    {#each $conversations as conv}
      <button
        class="conv-item"
        class:active={conv.id === $activeConversation}
        on:click={() => activeConversation.set(conv.id)}
      >
        <div class="conv-name">{conv.name}</div>
        <div class="conv-preview">{conv.last_msg}</div>
      </button>
    {/each}
    {#if $conversations.length === 0}
      <div class="empty">No conversations yet</div>
    {/if}
  </div>
</aside>

<style>
  .conversation-list {
    width: 220px;
    border-right: 1px solid var(--border-color);
    display: flex;
    flex-direction: column;
    background: var(--bg-secondary);
    backdrop-filter: blur(20px) saturate(180%);
    -webkit-backdrop-filter: blur(20px) saturate(180%);
  }

  .header {
    padding: 16px 16px 12px;
    border-bottom: 1px solid var(--border-light);
  }

  .header h3 {
    margin: 0;
    font-size: 12px;
    font-weight: 600;
    color: var(--text-secondary);
    text-transform: uppercase;
    letter-spacing: 0.5px;
  }

  .list {
    flex: 1;
    overflow-y: auto;
    padding: 4px 8px;
  }

  .conv-item {
    display: block;
    width: 100%;
    padding: 10px 12px;
    border: none;
    background: transparent;
    border-radius: var(--radius-md);
    cursor: pointer;
    text-align: left;
    margin-bottom: 2px;
    transition: all 0.15s ease;
  }

  .conv-item:hover {
    background: var(--hover-bg);
  }

  .conv-item.active {
    background: var(--accent);
  }

  .conv-item.active .conv-name,
  .conv-item.active .conv-preview {
    color: white;
  }

  .conv-name {
    font-size: 13px;
    font-weight: 500;
    color: var(--text-primary);
    margin-bottom: 2px;
  }

  .conv-preview {
    font-size: 11px;
    color: var(--text-tertiary);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .empty {
    padding: 20px;
    text-align: center;
    color: var(--text-tertiary);
    font-size: 12px;
  }
</style>
