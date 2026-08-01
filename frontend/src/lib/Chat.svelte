<script>
  import { onMount, afterUpdate } from 'svelte'
  import { messages, streamingContent, isStreaming, error } from '../lib/stores.js'
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

  onMount(async () => {
    const history = await getHistory(50, 0)
    messages.set(history)

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
      error.set(err)
      streamingContent.set('')
      isStreaming.set(false)
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
      error.set(err.message || String(err))
      isStreaming.set(false)
    }
  }

  function handleKeydown(e) {
    if (e.key === 'Enter' && (e.metaKey || e.ctrlKey)) {
      handleSend()
    }
  }

  function handleStop() {
    isStreaming.set(false)
    streamingContent.set('')
  }
</script>

<div class="chat">
  <div class="message-list" bind:this={messageList} on:scroll={handleScroll}>
    {#each $messages as msg (msg.id)}
      <MessageBubble message={msg} />
    {/each}

    {#if $isStreaming}
      <div class="streaming">
        <div class="avatar">R</div>
        <div class="bubble">
          {$streamingContent}<span class="cursor">|</span>
        </div>
      </div>
    {/if}

    {#if showJumpToBottom}
      <button class="jump-btn" on:click={scrollToBottom}> ↓ Jump to bottom </button>
    {/if}
  </div>

  <div class="input-area">
    {#if $error}
      <div class="error-bar">{$error}</div>
    {/if}
    <div class="input-row">
      <textarea
        bind:value={inputText}
        on:keydown={handleKeydown}
        placeholder="Type a message..."
        disabled={$isStreaming}
        rows="1"
      ></textarea>
      {#if $isStreaming}
        <button class="stop-btn" on:click={handleStop} title="Stop generation"> ■ </button>
      {:else}
        <button
          class="send-btn"
          on:click={handleSend}
          disabled={!inputText.trim()}
          title="Send (Cmd+Enter)"
        >
          →
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
    padding: 16px;
    display: flex;
    flex-direction: column;
    scroll-behavior: smooth;
  }

  .streaming {
    display: flex;
    gap: 8px;
    max-width: 80%;
    align-self: flex-start;
  }

  .streaming .avatar {
    width: 28px;
    height: 28px;
    border-radius: 50%;
    background: #007aff;
    color: white;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 12px;
    font-weight: 600;
    flex-shrink: 0;
  }

  .streaming .bubble {
    padding: 8px 12px;
    border-radius: 12px;
    font-size: 14px;
    line-height: 1.5;
    background: #f0f0f5;
    color: #1d1d1f;
    border-bottom-left-radius: 4px;
    word-wrap: break-word;
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
    border: 1px solid #d0d0d5;
    border-radius: 20px;
    background: white;
    font-size: 12px;
    cursor: pointer;
    color: #007aff;
    box-shadow: 0 1px 4px rgba(0, 0, 0, 0.1);
  }

  .jump-btn:hover {
    background: #f0f0f5;
  }

  .input-area {
    border-top: 1px solid #e0e0e0;
    padding: 12px 16px;
    background: white;
  }

  .error-bar {
    padding: 6px 12px;
    background: #fff2f0;
    border: 1px solid #ffccc7;
    border-radius: 6px;
    color: #cf1322;
    font-size: 12px;
    margin-bottom: 8px;
  }

  .input-row {
    display: flex;
    gap: 8px;
    align-items: flex-end;
  }

  textarea {
    flex: 1;
    padding: 8px 12px;
    border: 1px solid #d0d0d5;
    border-radius: 8px;
    font-size: 14px;
    font-family: inherit;
    resize: none;
    outline: none;
    line-height: 1.4;
    min-height: 36px;
  }

  textarea:focus {
    border-color: #007aff;
  }

  textarea:disabled {
    background: #f5f5f7;
  }

  .send-btn,
  .stop-btn {
    width: 36px;
    height: 36px;
    border: none;
    border-radius: 8px;
    cursor: pointer;
    font-size: 16px;
    display: flex;
    align-items: center;
    justify-content: center;
    flex-shrink: 0;
  }

  .send-btn {
    background: #007aff;
    color: white;
  }

  .send-btn:disabled {
    background: #d0d0d5;
    cursor: not-allowed;
  }

  .send-btn:hover:not(:disabled) {
    background: #0056cc;
  }

  .stop-btn {
    background: #ff3b30;
    color: white;
  }

  .stop-btn:hover {
    background: #d62d20;
  }
</style>
