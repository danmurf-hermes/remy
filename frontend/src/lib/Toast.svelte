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
    border-radius: 8px;
    font-size: 13px;
    line-height: 1.4;
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
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
    background: #fff2f0;
    border: 1px solid #ffccc7;
    color: #cf1322;
  }

  .toast-success {
    background: #f6ffed;
    border: 1px solid #b7eb8f;
    color: #389e0d;
  }

  .toast-info {
    background: #e6f7ff;
    border: 1px solid #91d5ff;
    color: #096dd9;
  }

  .toast-warning {
    background: #fffbe6;
    border: 1px solid #ffe58f;
    color: #d48806;
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

  /* Dark mode overrides */
  :global(.dark) .toast-error {
    background: #2a1215;
    border-color: #431418;
    color: #ff7875;
  }

  :global(.dark) .toast-success {
    background: #162312;
    border-color: #274916;
    color: #95de64;
  }

  :global(.dark) .toast-info {
    background: #111d2c;
    border-color: #15395b;
    color: #69b1ff;
  }

  :global(.dark) .toast-warning {
    background: #2b1d0e;
    border-color: #543a0b;
    color: #ffd666;
  }
</style>
