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
    <button class="new-btn" title="New conversation">+</button>
  </div>
  <div class="list">
    {#each $conversations as conv}
      <button
        class="conv-item"
        class:active={conv.id === $activeConversation}
        on:click={() => activeConversation.set(conv.id)}
      >
        <div class="conv-avatar">{conv.name.charAt(0).toUpperCase()}</div>
        <div class="conv-info">
          <div class="conv-name">{conv.name}</div>
          <div class="conv-preview">{conv.last_msg}</div>
        </div>
      </button>
    {/each}
    {#if $conversations.length === 0}
      <div class="empty">No conversations yet</div>
    {/if}
  </div>
</aside>

<style>
  .conversation-list {
    width: 240px;
    border-right: 1px solid var(--border-color);
    display: flex;
    flex-direction: column;
    background: var(--bg-secondary);
    backdrop-filter: blur(30px) saturate(200%);
    -webkit-backdrop-filter: blur(30px) saturate(200%);
  }

  .header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 16px 16px 12px;
    border-bottom: 1px solid var(--border-light);
  }

  .header h3 {
    margin: 0;
    font-size: 12px;
    font-weight: 600;
    color: var(--text-secondary);
    text-transform: uppercase;
    letter-spacing: 0.8px;
  }

  .new-btn {
    width: 24px;
    height: 24px;
    border: none;
    background: var(--accent);
    color: white;
    border-radius: 6px;
    cursor: pointer;
    font-size: 14px;
    font-weight: 600;
    display: flex;
    align-items: center;
    justify-content: center;
    transition: all 0.2s;
    box-shadow: 0 1px 3px rgba(0, 113, 227, 0.2);
  }

  .new-btn:hover {
    transform: scale(1.1);
    background: var(--accent-hover);
  }

  .list {
    flex: 1;
    overflow-y: auto;
    padding: 4px 8px;
  }

  .conv-item {
    display: flex;
    gap: 10px;
    width: 100%;
    padding: 10px 10px;
    border: none;
    background: transparent;
    border-radius: 10px;
    cursor: pointer;
    text-align: left;
    margin-bottom: 2px;
    transition: all 0.2s;
  }

  .conv-item:hover {
    background: var(--hover-bg);
  }

  .conv-item.active {
    background: var(--accent);
    box-shadow: 0 2px 6px rgba(0, 113, 227, 0.15);
  }

  .conv-item.active .conv-name,
  .conv-item.active .conv-preview {
    color: white;
  }

  .conv-avatar {
    width: 36px;
    height: 36px;
    border-radius: 10px;
    background: linear-gradient(135deg, var(--accent), #5856d6);
    color: white;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 14px;
    font-weight: 700;
    flex-shrink: 0;
    box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
  }

  .conv-item.active .conv-avatar {
    background: rgba(255, 255, 255, 0.2);
  }

  .conv-info {
    flex: 1;
    min-width: 0;
    display: flex;
    flex-direction: column;
    justify-content: center;
    gap: 2px;
  }

  .conv-name {
    font-size: 13px;
    font-weight: 500;
    color: var(--text-primary);
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
