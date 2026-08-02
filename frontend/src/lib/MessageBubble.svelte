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
    margin-bottom: 10px;
    max-width: 75%;
    animation: messageIn 0.2s ease-out;
  }

  @keyframes messageIn {
    from {
      opacity: 0;
      transform: translateY(4px);
    }
    to {
      opacity: 1;
      transform: translateY(0);
    }
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
    color: white;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 11px;
    font-weight: 600;
    flex-shrink: 0;
    box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
  }

  .bubble {
    padding: 10px 14px;
    border-radius: var(--radius-lg);
    font-size: 14px;
    line-height: 1.5;
    word-wrap: break-word;
    box-shadow: var(--shadow-sm);
  }

  .user-bubble {
    background: var(--accent);
    color: white;
    border-bottom-right-radius: var(--radius-sm);
  }

  .agent-bubble {
    background: var(--card-bg);
    color: var(--text-primary);
    border: 1px solid var(--border-light);
    border-bottom-left-radius: var(--radius-sm);
    backdrop-filter: blur(10px);
    -webkit-backdrop-filter: blur(10px);
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
