<script>
  import { toasts, removeToast } from './stores.js'
</script>

<div class="toast-container" role="alert" aria-live="polite">
  {#each $toasts as toast (toast.id)}
    <div class="toast toast-{toast.type}" role="alert">
      <span class="toast-message">{toast.message}</span>
      <button
        class="toast-close"
        on:click={() => removeToast(toast.id)}
        aria-label="Dismiss notification"
      >
        ×
      </button>
    </div>
  {/each}
</div>

<style>
  .toast-container {
    position: fixed;
    top: 16px;
    right: 16px;
    z-index: 10000;
    display: flex;
    flex-direction: column;
    gap: 8px;
    max-width: 400px;
  }

  .toast {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 12px 16px;
    border-radius: var(--radius-lg);
    font-size: 13px;
    line-height: 1.4;
    box-shadow: var(--shadow-lg);
    animation: slideIn 0.3s cubic-bezier(0.16, 1, 0.3, 1);
    backdrop-filter: blur(20px) saturate(180%);
    -webkit-backdrop-filter: blur(20px) saturate(180%);
    border: 1px solid var(--border-color);
  }

  @keyframes slideIn {
    from {
      transform: translateX(120%);
      opacity: 0;
    }
    to {
      transform: translateX(0);
      opacity: 1;
    }
  }

  .toast-error {
    background: rgba(255, 59, 48, 0.15);
    color: var(--danger);
  }

  .toast-success {
    background: rgba(52, 199, 89, 0.15);
    color: var(--success);
  }

  .toast-info {
    background: rgba(10, 132, 255, 0.15);
    color: var(--accent);
  }

  .toast-warning {
    background: rgba(255, 149, 0, 0.15);
    color: var(--warning);
  }

  .toast-message {
    flex: 1;
  }

  .toast-close {
    background: none;
    border: none;
    cursor: pointer;
    font-size: 16px;
    line-height: 1;
    padding: 0 4px;
    opacity: 0.6;
    color: inherit;
  }

  .toast-close:hover {
    opacity: 1;
  }
</style>
