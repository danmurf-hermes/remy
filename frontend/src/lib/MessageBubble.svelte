<script>
  /* eslint-disable svelte/no-at-html-tags */
  export let message = {
    id: '',
    role: 'user',
    content: '',
    timestamp: 0,
    interface: 'gui',
  }

  function formatTime(ts) {
    const d = new Date(ts)
    return d.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
  }

  function formatContent(text) {
    return text
      .replace(/&/g, '&amp;')
      .replace(/</g, '&lt;')
      .replace(/>/g, '&gt;')
      .replace(/\n/g, '<br>')
  }

  $: isUser = message.role === 'user'
  $: isAgent = message.role === 'assistant'
  $: isTelegram = message.interface === 'telegram'
</script>

<div class="message" class:user={isUser} class:agent={isAgent}>
  {#if isAgent}
    <div class="avatar">R</div>
  {/if}
  <div class="bubble" class:user-bubble={isUser} class:agent-bubble={isAgent}>
    <div class="content">{@html formatContent(message.content)}</div>
    <div class="meta">
      <span class="time">{formatTime(message.timestamp)}</span>
      {#if isTelegram}
        <span class="interface" title="Sent via Telegram">📱</span>
      {/if}
    </div>
  </div>
</div>

<style>
  .message {
    display: flex;
    gap: 8px;
    margin-bottom: 12px;
    max-width: 80%;
  }

  .message.user {
    align-self: flex-end;
    flex-direction: row-reverse;
  }

  .message.agent {
    align-self: flex-start;
  }

  .avatar {
    width: 28px;
    height: 28px;
    border-radius: 50%;
    background: var(--accent);
    color: var(--text-inverse);
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 12px;
    font-weight: 600;
    flex-shrink: 0;
  }

  .bubble {
    padding: 10px 14px;
    border-radius: 14px;
    font-size: 14px;
    line-height: 1.5;
    word-wrap: break-word;
    box-shadow: 0 1px 2px var(--shadow-sm);
  }

  .user-bubble {
    background: var(--accent);
    color: var(--text-inverse);
    border-bottom-right-radius: 4px;
  }

  .agent-bubble {
    background: var(--card-bg);
    color: var(--text-primary);
    border-bottom-left-radius: 4px;
  }

  .content {
    margin-bottom: 4px;
  }

  .meta {
    display: flex;
    align-items: center;
    gap: 4px;
    font-size: 11px;
    opacity: 0.7;
  }

  .user .meta {
    justify-content: flex-end;
  }

  .interface {
    font-size: 12px;
  }
</style>
