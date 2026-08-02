<script>
  /* eslint-disable svelte/no-at-html-tags */
  import { marked } from 'marked'

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

  function renderMarkdown(text) {
    return marked.parse(text, { breaks: true })
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
    <div class="content">{@html renderMarkdown(message.content)}</div>
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
    gap: 10px;
    margin-bottom: 6px;
    max-width: 78%;
    animation: messageIn 0.3s cubic-bezier(0.16, 1, 0.3, 1);
  }

  @keyframes messageIn {
    from {
      opacity: 0;
      transform: translateY(8px) scale(0.98);
    }
    to {
      opacity: 1;
      transform: translateY(0) scale(1);
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
    width: 30px;
    height: 30px;
    border-radius: 50%;
    background: linear-gradient(135deg, var(--accent), #5856d6);
    color: white;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 11px;
    font-weight: 700;
    flex-shrink: 0;
    box-shadow: 0 2px 6px rgba(0, 113, 227, 0.2);
    margin-top: 4px;
  }

  .bubble {
    padding: 10px 16px;
    border-radius: 18px;
    font-size: 14px;
    line-height: 1.5;
    word-wrap: break-word;
    position: relative;
  }

  .user-bubble {
    background: var(--accent);
    color: white;
    border-bottom-right-radius: 4px;
    box-shadow: 0 1px 4px rgba(0, 113, 227, 0.2);
  }

  .agent-bubble {
    background: var(--card-bg);
    color: var(--text-primary);
    border: 1px solid var(--border-light);
    border-bottom-left-radius: 4px;
    backdrop-filter: blur(10px);
    -webkit-backdrop-filter: blur(10px);
    box-shadow: 0 1px 3px rgba(0, 0, 0, 0.04);
  }

  .content {
    margin-bottom: 2px;
  }

  .meta {
    display: flex;
    align-items: center;
    gap: 4px;
    font-size: 10px;
    opacity: 0.6;
    margin-top: 2px;
  }

  .user .meta {
    justify-content: flex-end;
  }

  .interface {
    font-size: 11px;
  }
</style>
