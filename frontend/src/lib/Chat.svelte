<script>
  /* eslint-disable svelte/no-at-html-tags */
  import { marked } from 'marked'
  import { onMount, onDestroy, afterUpdate } from 'svelte'
  import { messages, streamingContent, isStreaming, addToast } from '../lib/stores.js'
  import {
    sendMessageStream,
    getHistory,
    onStreamChunk,
    onStreamDone,
    onStreamError,
  } from '../lib/wails.js'
  import MessageBubble from './MessageBubble.svelte'

  let inputText = ''
  let messageList
  let showJumpToBottom = false
  let inputEl

  onMount(async () => {
    try {
      const history = await getHistory(50, 0)
      messages.set(history)
    } catch (e) {
      addToast('Failed to load message history: ' + e.message, 'error')
    }

    onStreamChunk((chunk) => {
      streamingContent.update((prev) => prev + chunk)
    })

    onStreamDone(() => {
      const finalContent = $streamingContent
      if (finalContent) {
        messages.update((msgs) => [
          ...msgs,
          {
            id: Date.now().toString(),
            role: 'assistant',
            content: finalContent,
            timestamp: Date.now(),
            interface: 'gui',
          },
        ])
      }
      streamingContent.set('')
      isStreaming.set(false)
    })

    onStreamError((err) => {
      addToast(err, 'error')
      streamingContent.set('')
      isStreaming.set(false)
    })

    window.addEventListener('escape-pressed', () => {
      if (inputEl) {
        inputEl.blur()
      }
    })
  })

  onDestroy(() => {
    // Save any in-progress streaming content before unmounting
    const pendingContent = $streamingContent
    if (pendingContent) {
      messages.update((msgs) => [
        ...msgs,
        {
          id: Date.now().toString(),
          role: 'assistant',
          content: pendingContent,
          timestamp: Date.now(),
          interface: 'gui',
        },
      ])
    }
    streamingContent.set('')
    isStreaming.set(false)

    // Clean up stream event listeners to prevent duplicates on re-mount
    const runtime = window.runtime
    if (runtime) {
      runtime.EventsOff('stream:chunk')
      runtime.EventsOff('stream:done')
      runtime.EventsOff('stream:error')
    }
  })

  afterUpdate(() => {
    if (messageList && !showJumpToBottom) {
      messageList.scrollTop = messageList.scrollHeight
    }
  })

  function handleScroll() {
    if (!messageList) {
      return
    }
    const threshold = 100
    showJumpToBottom =
      messageList.scrollHeight - messageList.scrollTop - messageList.clientHeight > threshold
  }

  function scrollToBottom() {
    if (messageList) {
      messageList.scrollTop = messageList.scrollHeight
      showJumpToBottom = false
    }
  }

  async function handleSend() {
    const text = inputText.trim()
    if (!text || $isStreaming) {
      return
    }

    inputText = ''
    messages.update((msgs) => [
      ...msgs,
      {
        id: Date.now().toString(),
        role: 'user',
        content: text,
        timestamp: Date.now(),
        interface: 'gui',
      },
    ])

    isStreaming.set(true)
    streamingContent.set('')

    try {
      await sendMessageStream(text)
    } catch (err) {
      addToast(err.message || String(err), 'error')
      isStreaming.set(false)
    }
  }

  function handleKeydown(e) {
    if (e.key === 'Enter' && (e.metaKey || e.ctrlKey)) {
      e.preventDefault()
      handleSend()
    }
  }

  function handleStop() {
    isStreaming.set(false)
    streamingContent.set('')
  }
</script>

<div class="chat" role="region" aria-label="Chat conversation">
  <div class="chat-header">
    <div class="chat-header-info">
      <h2 class="chat-title">Chat</h2>
      <span class="chat-subtitle">Online</span>
    </div>
  </div>

  <div
    class="message-list"
    bind:this={messageList}
    on:scroll={handleScroll}
    role="log"
    aria-label="Messages"
    aria-live="polite"
  >
    {#each $messages as msg (msg.id)}
      <MessageBubble message={msg} />
    {/each}

    {#if $isStreaming}
      <div class="streaming" role="status" aria-label="Assistant is typing">
        <div class="avatar" aria-hidden="true">R</div>
        <div class="bubble">
          <!-- eslint-disable-next-line svelte/no-at-html-tags -->
          {@html marked.parse($streamingContent, { breaks: true })}
        </div>
      </div>
    {/if}

    {#if showJumpToBottom}
      <button class="jump-btn" on:click={scrollToBottom} aria-label="Scroll to bottom of messages">
        ↓
      </button>
    {/if}
  </div>

  {#if $isStreaming}
    <div class="typing-indicator" aria-live="polite">
      <span class="typing-dot"></span>
      <span class="typing-dot"></span>
      <span class="typing-dot"></span>
      <span class="typing-label">Remy is typing</span>
    </div>
  {/if}

  <div class="input-area">
    <div class="input-row">
      <textarea
        bind:this={inputEl}
        bind:value={inputText}
        on:keydown={handleKeydown}
        placeholder="Message Remy…"
        disabled={$isStreaming}
        rows="1"
        aria-label="Message input"
      ></textarea>
      {#if $isStreaming}
        <button
          class="stop-btn"
          on:click={handleStop}
          title="Stop generation"
          aria-label="Stop generation"
        >
          ■
        </button>
      {:else}
        <button
          class="send-btn"
          on:click={handleSend}
          disabled={!inputText.trim()}
          title="Send (Cmd+Enter)"
          aria-label="Send message"
        >
          <svg
            width="18"
            height="18"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            stroke-width="2.5"
            stroke-linecap="round"
            stroke-linejoin="round"
          >
            <line x1="22" y1="2" x2="11" y2="13" />
            <polygon points="22 2 15 22 11 13 2 9 22 2" />
          </svg>
        </button>
      {/if}
    </div>
  </div>
</div>

<style>
  .chat {
    display: flex;
    flex-direction: column;
    height: 100%;
    flex: 1;
  }

  .chat-header {
    padding: 12px 20px;
    border-bottom: 1px solid var(--border-light);
    background: var(--bg-primary);
    backdrop-filter: blur(30px) saturate(200%);
    -webkit-backdrop-filter: blur(30px) saturate(200%);
    flex-shrink: 0;
  }

  .chat-header-info {
    display: flex;
    align-items: baseline;
    gap: 8px;
  }

  .chat-title {
    margin: 0;
    font-size: 15px;
    font-weight: 600;
    color: var(--text-primary);
  }

  .chat-subtitle {
    font-size: 11px;
    color: var(--success);
    font-weight: 500;
  }

  .message-list {
    flex: 1;
    overflow-y: auto;
    padding: 16px 20px;
    display: flex;
    flex-direction: column;
    scroll-behavior: smooth;
  }

  .streaming {
    display: flex;
    gap: 10px;
    max-width: 78%;
    align-self: flex-start;
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

  .streaming .avatar {
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

  .streaming .bubble {
    padding: 10px 16px;
    border-radius: 18px;
    font-size: 14px;
    line-height: 1.5;
    background: var(--card-bg);
    color: var(--text-primary);
    border: 1px solid var(--border-light);
    border-bottom-left-radius: 4px;
    word-wrap: break-word;
    box-shadow: 0 1px 3px rgba(0, 0, 0, 0.04);
    backdrop-filter: blur(10px);
    -webkit-backdrop-filter: blur(10px);
  }

  .streaming .bubble :global(p) {
    margin: 0;
  }

  .streaming .bubble :global(code) {
    background: rgba(0, 0, 0, 0.08);
    padding: 2px 6px;
    border-radius: 4px;
    font-size: 13px;
    font-family: 'SF Mono', 'Fira Code', monospace;
  }

  .streaming .bubble :global(pre) {
    background: rgba(0, 0, 0, 0.08);
    padding: 12px;
    border-radius: 10px;
    overflow-x: auto;
    margin: 8px 0;
    font-size: 13px;
    line-height: 1.4;
  }

  .streaming .bubble :global(pre code) {
    background: none;
    padding: 0;
    border-radius: 0;
  }

  .typing-indicator {
    display: flex;
    align-items: center;
    gap: 4px;
    padding: 6px 20px 2px;
    font-size: 12px;
    color: var(--text-tertiary);
  }

  .typing-dot {
    width: 5px;
    height: 5px;
    border-radius: 50%;
    background: var(--text-tertiary);
    animation: typingBounce 1.4s ease-in-out infinite;
  }

  .typing-dot:nth-child(2) {
    animation-delay: 0.2s;
  }

  .typing-dot:nth-child(3) {
    animation-delay: 0.4s;
  }

  .typing-label {
    margin-left: 6px;
  }

  @keyframes typingBounce {
    0%,
    60%,
    100% {
      transform: translateY(0);
      opacity: 0.4;
    }
    30% {
      transform: translateY(-4px);
      opacity: 1;
    }
  }

  .jump-btn {
    position: sticky;
    bottom: 8px;
    align-self: center;
    width: 36px;
    height: 36px;
    border: 1px solid var(--border-color);
    border-radius: 50%;
    background: var(--bg-primary);
    font-size: 16px;
    cursor: pointer;
    color: var(--accent);
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
    backdrop-filter: blur(10px);
    -webkit-backdrop-filter: blur(10px);
    display: flex;
    align-items: center;
    justify-content: center;
    transition: all 0.2s;
  }

  .jump-btn:hover {
    background: var(--hover-bg);
    transform: translateY(-2px);
  }

  .input-area {
    border-top: 1px solid var(--border-light);
    padding: 12px 20px 16px;
    background: var(--bg-primary);
    backdrop-filter: blur(30px) saturate(200%);
    -webkit-backdrop-filter: blur(30px) saturate(200%);
  }

  .input-row {
    display: flex;
    gap: 8px;
    align-items: flex-end;
  }

  textarea {
    flex: 1;
    padding: 10px 16px;
    border: 1px solid var(--border-color);
    border-radius: 20px;
    font-size: 14px;
    font-family: inherit;
    resize: none;
    outline: none;
    line-height: 1.4;
    min-height: 40px;
    max-height: 120px;
    background: var(--input-bg);
    color: var(--text-primary);
    backdrop-filter: blur(10px);
    -webkit-backdrop-filter: blur(10px);
    transition: all 0.2s;
  }

  textarea:focus {
    border-color: var(--accent);
    box-shadow: 0 0 0 3px rgba(0, 113, 227, 0.12);
  }

  textarea:disabled {
    opacity: 0.6;
  }

  textarea::placeholder {
    color: var(--text-tertiary);
  }

  .send-btn,
  .stop-btn {
    width: 40px;
    height: 40px;
    border: none;
    border-radius: 50%;
    cursor: pointer;
    display: flex;
    align-items: center;
    justify-content: center;
    flex-shrink: 0;
    transition: all 0.2s cubic-bezier(0.16, 1, 0.3, 1);
  }

  .send-btn {
    background: var(--accent);
    color: white;
    box-shadow: 0 2px 8px rgba(0, 113, 227, 0.3);
  }

  .send-btn:disabled {
    background: var(--bg-tertiary);
    color: var(--text-tertiary);
    cursor: not-allowed;
    box-shadow: none;
  }

  .send-btn:hover:not(:disabled) {
    background: var(--accent-hover);
    transform: scale(1.08);
  }

  .stop-btn {
    background: var(--danger);
    color: white;
    box-shadow: 0 2px 8px rgba(255, 59, 48, 0.3);
  }

  .stop-btn:hover {
    background: var(--danger-hover);
    transform: scale(1.08);
  }
</style>
