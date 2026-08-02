<script>
  import { onMount, afterUpdate } from 'svelte'
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

    // Listen for escape to blur the input
    window.addEventListener('escape-pressed', () => {
      if (inputEl) {
        inputEl.blur()
      }
    })
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
          {$streamingContent}<span class="cursor" aria-hidden="true">|</span>
        </div>
      </div>
    {/if}

    {#if showJumpToBottom}
      <button class="jump-btn" on:click={scrollToBottom} aria-label="Scroll to bottom of messages">
        ↓ Jump to bottom
      </button>
    {/if}
  </div>

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
          ↑
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

  .message-list {
    flex: 1;
    overflow-y: auto;
    padding: 20px 24px;
    display: flex;
    flex-direction: column;
    scroll-behavior: smooth;
  }

  .streaming {
    display: flex;
    gap: 8px;
    max-width: 75%;
    align-self: flex-start;
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

  .streaming .avatar {
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

  .streaming .bubble {
    padding: 10px 14px;
    border-radius: var(--radius-lg);
    font-size: 14px;
    line-height: 1.5;
    background: var(--card-bg);
    color: var(--text-primary);
    border: 1px solid var(--border-light);
    border-bottom-left-radius: var(--radius-sm);
    word-wrap: break-word;
    box-shadow: var(--shadow-sm);
    backdrop-filter: blur(10px);
    -webkit-backdrop-filter: blur(10px);
  }

  .cursor {
    animation: blink 1s step-end infinite;
  }

  @keyframes blink {
    50% {
      opacity: 0;
    }
  }

  .jump-btn {
    position: sticky;
    bottom: 8px;
    align-self: center;
    padding: 6px 16px;
    border: 1px solid var(--border-color);
    border-radius: var(--radius-xl);
    background: var(--bg-primary);
    font-size: 12px;
    cursor: pointer;
    color: var(--accent);
    box-shadow: var(--shadow-sm);
    backdrop-filter: blur(10px);
    -webkit-backdrop-filter: blur(10px);
  }

  .jump-btn:hover {
    background: var(--hover-bg);
  }

  .input-area {
    border-top: 1px solid var(--border-light);
    padding: 12px 20px;
    background: var(--bg-primary);
    backdrop-filter: blur(20px) saturate(180%);
    -webkit-backdrop-filter: blur(20px) saturate(180%);
  }

  .input-row {
    display: flex;
    gap: 8px;
    align-items: flex-end;
  }

  textarea {
    flex: 1;
    padding: 10px 14px;
    border: 1px solid var(--border-color);
    border-radius: var(--radius-lg);
    font-size: 14px;
    font-family: inherit;
    resize: none;
    outline: none;
    line-height: 1.4;
    min-height: 40px;
    background: var(--input-bg);
    color: var(--text-primary);
    backdrop-filter: blur(10px);
    -webkit-backdrop-filter: blur(10px);
  }

  textarea:focus {
    border-color: var(--accent);
    box-shadow: 0 0 0 3px rgba(0, 113, 227, 0.15);
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
    font-size: 18px;
    display: flex;
    align-items: center;
    justify-content: center;
    flex-shrink: 0;
    transition: all 0.15s ease;
  }

  .send-btn {
    background: var(--accent);
    color: white;
    box-shadow: 0 1px 3px rgba(0, 0, 0, 0.15);
  }

  .send-btn:disabled {
    background: var(--bg-tertiary);
    color: var(--text-tertiary);
    cursor: not-allowed;
    box-shadow: none;
  }

  .send-btn:hover:not(:disabled) {
    background: var(--accent-hover);
    transform: scale(1.05);
  }

  .stop-btn {
    background: var(--danger);
    color: white;
    box-shadow: 0 1px 3px rgba(0, 0, 0, 0.15);
  }

  .stop-btn:hover {
    background: var(--danger-hover);
    transform: scale(1.05);
  }
</style>
