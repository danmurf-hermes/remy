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

  function formatContent(text) {
    const html = marked.parse(text, { breaks: true })
    return html
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

  .content :global(p) {
    margin: 0 0 6px;
  }

  .content :global(p:last-child) {
    margin-bottom: 0;
  }

  .content :global(code) {
    background: rgba(0, 0, 0, 0.08);
    padding: 2px 6px;
    border-radius: 4px;
    font-size: 13px;
    font-family: 'SF Mono', 'Fira Code', monospace;
  }

  .content :global(pre) {
    background: rgba(0, 0, 0, 0.08);
    padding: 12px;
    border-radius: 10px;
    overflow-x: auto;
    margin: 8px 0;
    font-size: 13px;
    line-height: 1.4;
  }

  .content :global(pre code) {
    background: none;
    padding: 0;
    border-radius: 0;
  }

  .content :global(ul),
  .content :global(ol) {
    margin: 4px 0;
    padding-left: 20px;
  }

  .content :global(li) {
    margin-bottom: 2px;
  }

  .content :global(strong) {
    font-weight: 600;
  }

  .content :global(blockquote) {
    border-left: 3px solid var(--accent);
    margin: 8px 0;
    padding: 4px 12px;
    opacity: 0.85;
  }

  .content :global(a) {
    color: var(--accent);
    text-decoration: none;
  }

  .content :global(a:hover) {
    text-decoration: underline;
  }

  .content :global(h1),
  .content :global(h2),
  .content :global(h3),
  .content :global(h4) {
    margin: 10px 0 4px;
    font-weight: 600;
  }

  .content :global(h1) {
    font-size: 16px;
  }
  .content :global(h2) {
    font-size: 15px;
  }
  .content :global(h3) {
    font-size: 14px;
  }
  .content :global(h4) {
    font-size: 14px;
  }

  .content :global(hr) {
    border: none;
    border-top: 1px solid var(--border-light);
    margin: 12px 0;
  }

  .content :global(table) {
    border-collapse: collapse;
    margin: 8px 0;
    font-size: 13px;
    width: 100%;
  }

  .content :global(th),
  .content :global(td) {
    border: 1px solid var(--border-light);
    padding: 6px 10px;
    text-align: left;
  }

  .content :global(th) {
    font-weight: 600;
    background: rgba(0, 0, 0, 0.04);
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
