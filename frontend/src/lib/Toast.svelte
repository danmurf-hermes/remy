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
    padding: 10px 16px;
    border-radius: 10px;
    font-size: 13px;
    line-height: 1.4;
    box-shadow: 0 4px 12px var(--shadow-md);
    animation: slideIn 0.2s ease-out;
  }

  @keyframes slideIn {
    from {
      transform: translateX(100%);
      opacity: 0;
    }
    to {
      transform: translateX(0);
      opacity: 1;
    }
  }

  .toast-error {
    background: var(--danger-subtle);
    border: 1px solid var(--danger);
    color: var(--danger-text);
  }

  .toast-success {
    background: var(--success-subtle);
    border: 1px solid var(--success);
    color: var(--success-text);
  }

  .toast-info {
    background: var(--accent-subtle);
    border: 1px solid var(--accent);
    color: var(--accent-text);
  }

  .toast-warning {
    background: var(--warning-subtle);
    border: 1px solid var(--warning);
    color: var(--warning-text);
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
