<script>
  import { onMount } from 'svelte'
  import { tasks } from './stores.js'
  import { getTasks, createTask, cancelTask } from './wails.js'

  let upcomingTasks = []
  let firedTasks = []
  let loading = true
  let showNewForm = false
  let newTaskText = ''
  let newTaskDate = ''
  let newTaskTime = ''
  let newTaskCron = ''
  let error = null

  $: {
    upcomingTasks = $tasks.filter((t) => t.status === 'pending')
    firedTasks = $tasks.filter((t) => t.status === 'fired')
  }

  onMount(async () => {
    await loadTasks()
  })

  async function loadTasks() {
    loading = true
    try {
      const result = await getTasks('')
      tasks.set(result)
    } catch (e) {
      error = 'Failed to load tasks: ' + e.message
    }
    loading = false
  }

  async function handleCreate() {
    if (!newTaskText) {
      return
    }
    error = null

    let triggerAt = ''
    if (newTaskDate) {
      const dateStr = newTaskDate + (newTaskTime ? 'T' + newTaskTime : 'T12:00')
      triggerAt = String(new Date(dateStr).getTime())
    }

    const action = JSON.stringify({ text: newTaskText })
    try {
      await createTask('reminder', triggerAt, newTaskCron, action, '')
      await loadTasks()
      showNewForm = false
      newTaskText = ''
      newTaskDate = ''
      newTaskTime = ''
      newTaskCron = ''
    } catch (e) {
      error = 'Failed to create task: ' + e.message
    }
  }

  async function handleCancel(id) {
    try {
      await cancelTask(id)
      await loadTasks()
    } catch (e) {
      error = 'Failed to cancel task: ' + e.message
    }
  }

  function formatTime(ts) {
    if (!ts || ts === 0) {
      return '—'
    }
    return new Date(ts).toLocaleString()
  }

  function humanReadableCron(expr) {
    if (!expr) {
      return null
    }
    const parts = expr.split(' ')
    if (parts.length < 5) {
      return expr
    }
    let result = ''
    if (parts[1] === '*') {
      result += 'every hour'
    } else {
      result += 'at minute ' + parts[1]
    }
    if (parts[0] !== '*') {
      result = 'at ' + parts[0] + ':' + parts[1]
    }
    if (parts[4] !== '*') {
      const days = ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat']
      result +=
        ' on ' +
        parts[4]
          .split(',')
          .map((d) => days[parseInt(d)])
          .join(', ')
    }
    return result || expr
  }

  function parseAction(action) {
    try {
      const parsed = JSON.parse(action)
      return parsed.text || action
    } catch {
      return action
    }
  }
</script>

<div class="task-manager">
  <div class="header">
    <h2>Tasks & Schedule</h2>
    <button class="btn-primary" on:click={() => (showNewForm = !showNewForm)}>
      {showNewForm ? 'Cancel' : '+ New Reminder'}
    </button>
  </div>

  {#if error}
    <div class="error-banner">{error}</div>
  {/if}

  {#if showNewForm}
    <div class="new-task-form">
      <h3>New Reminder</h3>
      <input
        type="text"
        placeholder="What do you need to remember?"
        bind:value={newTaskText}
        class="input"
      />
      <div class="form-row">
        <label>
          Date
          <input type="date" bind:value={newTaskDate} class="input" />
        </label>
        <label>
          Time
          <input type="time" bind:value={newTaskTime} class="input" />
        </label>
      </div>
      <label>
        Cron expression (optional, for recurring)
        <input type="text" placeholder="e.g. 0 9 * * 1-5" bind:value={newTaskCron} class="input" />
      </label>
      <button class="btn-primary" on:click={handleCreate}>Create Reminder</button>
    </div>
  {/if}

  {#if loading}
    <div class="loading">Loading tasks…</div>
  {:else}
    <div class="sections">
      <section class="task-section">
        <h3>Upcoming ({upcomingTasks.length})</h3>
        {#if upcomingTasks.length === 0}
          <p class="empty">No upcoming reminders</p>
        {:else}
          {#each upcomingTasks as task (task.id)}
            <div class="task-card">
              <div class="task-info">
                <span class="task-text">{parseAction(task.action)}</span>
                <span class="task-time">
                  {#if task.cron_expr}
                    🔄 {humanReadableCron(task.cron_expr)} — next: {formatTime(task.trigger_at)}
                  {:else}
                    ⏰ {formatTime(task.trigger_at)}
                  {/if}
                </span>
              </div>
              <div class="task-actions">
                <button class="btn-small btn-danger" on:click={() => handleCancel(task.id)}>
                  Cancel
                </button>
              </div>
            </div>
          {/each}
        {/if}
      </section>

      <section class="task-section">
        <h3>Fired ({firedTasks.length})</h3>
        {#if firedTasks.length === 0}
          <p class="empty">No fired reminders</p>
        {:else}
          {#each firedTasks as task (task.id)}
            <div class="task-card fired">
              <div class="task-info">
                <span class="task-text">{parseAction(task.action)}</span>
                <span class="task-time">✅ Fired at {formatTime(task.fired_at)}</span>
              </div>
            </div>
          {/each}
        {/if}
      </section>
    </div>
  {/if}
</div>

<style>
  .task-manager {
    flex: 1;
    padding: 24px;
    overflow-y: auto;
  }

  .header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 20px;
  }

  .header h2 {
    margin: 0;
    font-size: 20px;
    font-weight: 600;
  }

  .btn-primary {
    background: var(--accent);
    color: white;
    border: none;
    padding: 8px 16px;
    border-radius: var(--radius-md);
    cursor: pointer;
    font-size: 13px;
    font-weight: 500;
    transition: all 0.15s ease;
  }

  .btn-primary:hover {
    background: var(--accent-hover);
  }

  .btn-small {
    padding: 4px 10px;
    border-radius: var(--radius-sm);
    border: none;
    cursor: pointer;
    font-size: 12px;
    transition: all 0.15s ease;
  }

  .btn-danger {
    background: var(--danger);
    color: white;
  }

  .btn-danger:hover {
    background: var(--danger-hover);
  }

  .error-banner {
    background: rgba(255, 59, 48, 0.1);
    color: var(--danger);
    padding: 8px 12px;
    border-radius: var(--radius-md);
    margin-bottom: 12px;
    font-size: 13px;
    border: 1px solid rgba(255, 59, 48, 0.2);
  }

  .new-task-form {
    background: var(--card-bg);
    border: 1px solid var(--border-light);
    border-radius: var(--radius-lg);
    padding: 16px;
    margin-bottom: 20px;
    backdrop-filter: blur(10px);
    -webkit-backdrop-filter: blur(10px);
  }

  .new-task-form h3 {
    margin: 0 0 12px 0;
    font-size: 15px;
  }

  .input {
    width: 100%;
    padding: 8px 12px;
    border: 1px solid var(--border-color);
    border-radius: var(--radius-md);
    font-size: 13px;
    box-sizing: border-box;
    margin-bottom: 8px;
    background: var(--input-bg);
    color: var(--text-primary);
    outline: none;
    transition: border-color 0.15s;
  }

  .input:focus {
    border-color: var(--accent);
    box-shadow: 0 0 0 3px rgba(0, 113, 227, 0.15);
  }

  .form-row {
    display: flex;
    gap: 12px;
  }

  .form-row label {
    flex: 1;
    font-size: 12px;
    color: var(--text-secondary);
  }

  label {
    display: block;
    font-size: 12px;
    color: var(--text-secondary);
    margin-bottom: 4px;
  }

  .loading {
    text-align: center;
    color: var(--text-tertiary);
    padding: 40px;
  }

  .sections {
    display: flex;
    flex-direction: column;
    gap: 24px;
  }

  .task-section h3 {
    margin: 0 0 12px 0;
    font-size: 15px;
    font-weight: 600;
  }

  .empty {
    color: var(--text-tertiary);
    font-size: 13px;
    padding: 12px 0;
  }

  .task-card {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 12px 16px;
    background: var(--card-bg);
    border: 1px solid var(--border-light);
    border-radius: var(--radius-lg);
    margin-bottom: 8px;
    backdrop-filter: blur(10px);
    -webkit-backdrop-filter: blur(10px);
    transition: all 0.15s ease;
  }

  .task-card:hover {
    border-color: var(--border-color);
  }

  .task-card.fired {
    opacity: 0.6;
  }

  .task-info {
    display: flex;
    flex-direction: column;
    gap: 4px;
  }

  .task-text {
    font-size: 14px;
    font-weight: 500;
  }

  .task-time {
    font-size: 12px;
    color: var(--text-secondary);
  }

  .task-actions {
    display: flex;
    gap: 6px;
  }
</style>
