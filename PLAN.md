# Remy — Build Plan

> **Status:** This is a living document. Each stage is worked on sequentially. When you pick up a stage, read this plan to see where we're up to. When you finish a stage, update the status and add notes about what was done, any deviations, and what the next person needs to know.

---

## How to Use This Plan

1. **Read the plan** — find the first stage marked `[ ]` (not started) or `[in progress]`.
2. **Read the ARCHITECTURE.md** — understand the full design before starting.
3. **Work on the stage** — implement everything listed, including tests.
4. **Update the plan** — mark the stage `[x]`, add a brief summary of what was done, any deviations from the plan, and any notes for the next person.
5. **Commit your changes** — including the updated PLAN.md.
6. **Move to the next stage** — or hand off to the next person.

---

## Quality Control Strategy

Every stage must include tests at the appropriate level:

### Go Unit Tests (required for all Go code)
- Table-driven tests using the standard `testing` package
- Mock interfaces for external dependencies (LLM, database, Telegram)
- Aim for 80%+ coverage on `internal/` packages
- Run with: `go test ./internal/...`

### Frontend Unit Tests (required for all Svelte/JS code)
- Vitest + @testing-library/svelte for component tests
- Test rendering, user interactions, and state changes
- Mock Wails runtime bindings
- Run with: `npm test` (in `frontend/`)

### Integration / BDD Tests (end-to-end)
- **Playwright** for full-stack E2E tests against the built Wails app
- Tests simulate real user flows: send a message, verify response appears, check memory persistence
- Run in CI via GitHub Actions (headless mode)
- Also: Go integration tests with build tag `//go:build integration` for tests needing Ollama or a real SQLite database

### CI Pipeline (GitHub Actions)
Every push runs:
1. `go vet ./...` — static analysis
2. `go test ./internal/... -cover` — Go unit tests with coverage
3. `npm test` (in `frontend/`) — frontend unit tests
4. `go build ./cmd/remy` — verify it compiles
5. On `main` branch: Playwright E2E tests (if a built binary is available)

---

## Stage 1: Project Scaffolding & Build System

**Goal:** Set up the Go module, Wails project, frontend skeleton, CI pipeline, and Makefile. Nothing functional yet — just the skeleton that compiles and runs.

**Tasks:**
- [x] Initialize Go module (`go mod init github.com/yourname/remy`)
- [x] Create directory structure per ARCHITECTURE.md §11.1
- [x] Set up Wails project (`wails init` with Svelte template)
- [x] Create `cmd/remy/main.go` — minimal entry point that prints "Remy starting..." and exits
- [x] Create `Makefile` with targets: `build`, `dev`, `test`, `lint`, `clean`
- [x] Create `.github/workflows/ci.yml` — runs on push/PR:
  - `go vet ./...`
  - `go test ./internal/... -cover`
  - `npm test` (in `frontend/`)
  - `go build ./cmd/remy`
- [x] Create `frontend/` skeleton with Svelte + Vite + Vitest
  - `App.svelte` — minimal "Hello Remy" component
  - Basic Vitest config with a smoke test
- [x] Create `internal/config/config.go` — load/save `config.json` from `~/.remy/`
- [x] Create `internal/config/config_test.go` — table-driven tests for config loading
- [x] Add Go dependencies: `wailsapp/wails/v2`, `google/uuid`
- [x] Add frontend dependencies: `svelte`, `@sveltejs/vite-plugin-svelte`, `vitest`, `@testing-library/svelte`
- [x] Verify `make build` produces a working binary
- [x] Verify `make test` passes (Go + frontend)
- [x] Verify CI passes on a test push

**Acceptance criteria:**
- `make build` produces a binary in `build/`
- `make test` passes with 0 failures
- CI green on push
- `remy` binary runs and prints startup message

**Notes for next person:**
- Wails dependency (`wailsapp/wails/v2`) was not added yet — it will be pulled in when the Wails project is properly initialized in Stage 8 (GUI). For now the CLI binary builds without it.
- The `wails init` step was skipped since we don't need the full Wails template yet; the frontend skeleton is standalone Svelte + Vite + Vitest. Wails integration will happen in Stage 8.
- `go vet`, `go test`, `npm test`, and `go build` all pass.
- Config package has 77.8% test coverage.

---

## Stage 2: SQLite Database Layer

**Goal:** Implement the full SQLite database layer with schema, migrations, vector support, and all CRUD operations. This is the foundation everything else depends on.

**Tasks:**
- [x] Add `mattn/go-sqlite3` and `asg017/sqlite-vec-go` dependencies
- [x] Create `internal/memory/store.go`:
  - `NewStore(dbPath string) (*Store, error)` — opens/creates DB, runs migrations
  - Auto-creates all tables from ARCHITECTURE.md §3.2 and §3.3
  - WAL mode for concurrent reads
- [x] Create `internal/memory/messages.go`:
  - `SaveMessage(ctx, Message) error`
  - `GetMessages(ctx, limit, offset) ([]Message, error)`
  - `GetMessagesBySession(ctx, sessionID) ([]Message, error)`
  - `GetMessage(ctx, id) (Message, error)`
- [x] Create `internal/memory/episodic.go`:
  - `SaveEpisode(ctx, Episode) error`
  - `SearchEpisodes(ctx, embedding, limit) ([]Episode, error)`
  - `GetEpisodesByTimeRange(ctx, start, end) ([]Episode, error)`
- [x] Create `internal/memory/semantic.go`:
  - `SaveFact(ctx, Fact) error`
  - `SearchFacts(ctx, embedding, limit) ([]Fact, error)`
  - `GetFactsByCategory(ctx, category) ([]Fact, error)`
  - `UpdateFact(ctx, Fact) error`
  - `SaveEntity(ctx, Entity) error`
  - `SaveRelationship(ctx, Relationship) error`
- [x] Create `internal/memory/scratchpad.go`:
  - `GetScratchpad(ctx) (string, error)`
  - `UpdateScratchpad(ctx, content) error`
- [x] Create `internal/memory/activity.go`:
  - `LogActivity(ctx, ActivityEntry) error`
  - `GetActivityLog(ctx, filter, limit, offset) ([]ActivityEntry, error)`
- [x] Create `internal/memory/vectors.go`:
  - `GenerateEmbedding(ctx, text) ([]float32, error)` — calls Ollama embedding API
  - Helper functions for inserting/searching vec0 virtual tables
- [x] Create `internal/memory/store_test.go`:
  - Table-driven tests for every CRUD operation
  - Use `t.TempDir()` for isolated test databases
  - Test edge cases: empty results, duplicate IDs, large content
- [x] Create `internal/memory/vectors_test.go`:
  - Test embedding generation (mock the HTTP call)
  - Test vector search with known embeddings

**Acceptance criteria:**
- All store tests pass with `go test ./internal/memory/... -cover`
- Coverage >= 80% on `internal/memory/`
- Database file is created at the specified path
- Schema matches ARCHITECTURE.md exactly
- Vectors can be stored and searched

**Notes for next person:**
- Uses `golang-migrate/migrate/v4` with embedded SQL migration files in `internal/memory/migrations/`. To add a new migration, create `000003_<name>.up.sql` and `000003_<name>.down.sql` in that directory.
- The vec0 virtual tables require exactly 768-dimensional embeddings (matching `nomic-embed-text`). The `k` parameter is implicit via `LIMIT` in the subquery pattern used for vector search.
- Migrations run on a separate database connection to avoid the `migrate` sqlite3 driver closing the main connection.
- 31 tests, 85.8% coverage on `internal/memory/`.

---

## Stage 3: LLM Client & Provider Abstraction

**Goal:** Implement the OpenAI-compatible API client and provider abstraction so the agent can talk to Ollama (and eventually other providers).

**Tasks:**
- [ ] Create `internal/llm/provider.go`:
  - `Provider` interface: `Chat(ctx, req ChatRequest) (*ChatResponse, error)`, `Embed(ctx, text) ([]float32, error)`
  - `NewProvider(config ProviderConfig) (Provider, error)` — factory
- [ ] Create `internal/llm/client.go`:
  - `OllamaClient` implementing `Provider`
  - `Chat(ctx, req)` — POST to `/v1/chat/completions`
  - `Embed(ctx, text)` — POST to `/v1/embeddings`
  - Streaming support via SSE (for GUI streaming)
  - Configurable timeout, retry, error handling
- [ ] Create `internal/llm/types.go`:
  - `ChatRequest`, `ChatResponse`, `Message` (role/content), `StreamChunk`
- [ ] Create `internal/llm/client_test.go`:
  - Mock HTTP server for testing API calls
  - Table-driven tests for: successful chat, streaming, errors, timeouts, malformed responses
  - Test embedding generation
- [ ] Create `internal/llm/provider_test.go`:
  - Test provider factory with valid/invalid configs
  - Test fallback behavior

**Acceptance criteria:**
- All tests pass with `go test ./internal/llm/... -cover`
- Coverage >= 80%
- Client can connect to a real Ollama instance (integration test, build-tagged)
- Streaming works correctly (tokens arrive in order, no data loss)

**Notes for next person:**

---

## Stage 4: Agent Core Loop

**Goal:** Implement the core agent orchestration loop — receive message, retrieve context, build prompt, call LLM, store response, send to interface.

**Tasks:**
- [ ] Create `internal/agent/agent.go`:
  - `NewAgent(store, llm, persona) *Agent`
  - `HandleMessage(ctx, msg) (*Message, error)` — the core loop (ARCHITECTURE.md §9)
  - Context retrieval: embed message, search episodes + facts, load scratchpad, load recent history
  - Prompt building: system prompt + scratchpad + retrieved context + history + current message
  - Response handling: store response, return it
- [ ] Create `internal/agent/prompt.go`:
  - `BuildSystemPrompt(persona) string`
  - `BuildContextSection(episodes, facts) string`
  - Token counting and context window management
- [ ] Create `internal/agent/agent_test.go`:
  - Mock store and mock LLM provider
  - Table-driven tests for: normal message flow, empty context, context retrieval, error handling
  - Test prompt building with various inputs
  - Test context window eviction logic

**Acceptance criteria:**
- All tests pass with `go test ./internal/agent/... -cover`
- Coverage >= 80%
- Agent correctly retrieves context and builds prompts
- Agent handles errors gracefully (LLM down, store failure)

**Notes for next person:**

---

## Stage 5: Persona System

**Goal:** Implement persona loading, parsing, switching, and model override resolution.

**Tasks:**
- [ ] Create `internal/persona/persona.go`:
  - `LoadPersona(path string) (*Persona, error)` — parse Markdown with YAML frontmatter
  - `ListPersonas(dir string) ([]PersonaSummary, error)`
  - `SavePersona(path string, p *Persona) error`
  - `GetActivePersona() *Persona`
  - `SetActivePersona(name string) error`
- [ ] Create `internal/persona/persona_test.go`:
  - Test parsing valid/invalid persona files
  - Test frontmatter extraction (provider, model, temperature, max_tokens)
  - Test model override resolution
  - Test listing personas from directory
- [ ] Integrate persona loading into agent startup
- [ ] Add persona switching to agent loop (detect "switch to X" in user message)

**Acceptance criteria:**
- All tests pass
- Persona files with YAML frontmatter parse correctly
- Model overrides are resolved correctly
- Switching personas changes agent behavior on next message

**Notes for next person:**

---

## Stage 6: Consolidation Engine

**Goal:** Implement the two-phase consolidation system — quick summarization after inactivity and deep consolidation during idle time.

**Tasks:**
- [ ] Create `internal/agent/consolidation.go`:
  - `QuickConsolidation(ctx, recentMessages)` — summarize conversation block, store as episode with embedding
  - `DeepConsolidation(ctx)` — extract facts, entities, relationships from recent episodes; deduplicate and merge
  - `ScheduleConsolidation(ctx, delay)` — run after inactivity
- [ ] Create `internal/agent/consolidation_test.go`:
  - Test quick consolidation with mock LLM
  - Test fact extraction and deduplication
  - Test entity/relationship extraction
- [ ] Integrate consolidation into agent loop (trigger after ~5 min inactivity)
- [ ] Add deep consolidation goroutine that runs during extended idle periods

**Acceptance criteria:**
- All tests pass
- Quick consolidation runs after inactivity and stores episodes
- Deep consolidation extracts facts and entities correctly
- Facts are deduplicated and confidence scores updated

**Notes for next person:**

---

## Stage 7: Scheduler & Task System

**Goal:** Implement the scheduler engine — create reminders, recurring tasks, fire them at the right time, deliver to interfaces.

**Tasks:**
- [ ] Create `internal/scheduler/scheduler.go`:
  - `NewScheduler(store, agent) *Scheduler`
  - `Start(ctx)` — background loop checking every 30s for due tasks
  - `CreateTask(ctx, task) error`
  - `CancelTask(ctx, id) error`
  - `GetTasks(ctx, status) ([]Task, error)`
- [ ] Create `internal/scheduler/tasks.go`:
  - Task CRUD operations on the `tasks` table
  - Cron expression parsing (using `robfig/cron`)
  - Next occurrence calculation for recurring tasks
- [ ] Create `internal/scheduler/scheduler_test.go`:
  - Test task creation, cancellation, firing
  - Test cron parsing and next-occurrence calculation
  - Test scheduler loop with mock store
- [ ] Integrate scheduler into agent (agent can call `create_reminder`, `create_schedule`)
- [ ] Add task awareness to system prompt (upcoming tasks summary)

**Acceptance criteria:**
- All tests pass
- Tasks fire at the correct time
- Recurring tasks work correctly
- Agent can create and manage tasks via conversation

**Notes for next person:**

---

## Stage 8: Desktop GUI — Chat & Core UI

**Goal:** Build the primary chat interface in Svelte — message list, input, streaming, conversation list, and the core layout with sidebar navigation.

**Tasks:**
- [ ] Set up Wails frontend with Svelte routing and tab navigation
- [ ] Create sidebar component with icon-only tab buttons (Chat, Memory, Tasks, Personas, Activity, Settings)
- [ ] Build `Chat.svelte`:
  - Message list with user/agent bubbles
  - Markdown rendering for agent messages
  - Streaming token display with cursor animation
  - Auto-scroll with "Jump to bottom" button
  - Stop generation button
- [ ] Build `MessageBubble.svelte`:
  - User vs agent styling (right/left aligned, colors)
  - Timestamp display
  - Interface indicator (Telegram icon for cross-interface messages)
- [ ] Build conversation list sidebar within Chat tab
- [ ] Implement Wails Go bindings for chat:
  - `SendMessage(text string) (Message, error)`
  - `GetHistory(limit, offset) ([]Message, error)`
  - `StreamResponse` — push events to frontend via Wails event system
- [ ] Create `frontend/src/lib/wails.js` — typed wrappers around Wails runtime
- [ ] Create `frontend/src/lib/stores.js` — Svelte stores for messages, streaming state, active tab
- [ ] Write frontend unit tests:
  - `MessageBubble.test.js` — renders user/agent messages correctly
  - `Chat.test.js` — message list renders, input works, streaming updates
  - `Sidebar.test.js` — tab switching works
- [ ] Write Playwright E2E test:
  - Launch app, send a message, verify response appears in chat

**Acceptance criteria:**
- `make dev` launches the app with hot-reload
- Chat works end-to-end: type message → agent responds → response streams in
- All frontend tests pass
- Playwright test passes (if binary is built)

**Notes for next person:**

---

## Stage 9: Desktop GUI — Memory Explorer

**Goal:** Build the Memory Explorer tab — facts panel, episode timeline, entity graph, search, scratchpad viewer.

**Tasks:**
- [ ] Build `MemoryExplorer.svelte` — tab container with sub-views
- [ ] Build `FactList.svelte`:
  - Card grid grouped by category
  - Confidence bar visualization
  - Edit/delete on hover
  - Inline editing
- [ ] Build `EpisodeTimeline.svelte`:
  - Vertical timeline with dots and lines
  - Expandable entries with full summary
  - "View messages" button linking to Chat tab
- [ ] Build `EntityGraph.svelte`:
  - Force-directed graph (Canvas or SVG)
  - Nodes colored by entity type
  - Click to highlight and show details panel
  - Pan and zoom
- [ ] Build search bar with semantic/full-text toggle
- [ ] Build `ScratchpadViewer.svelte` — editable text area with auto-save
- [ ] Implement Wails Go bindings:
  - `GetFacts(category)`, `UpdateFact(id, fact)`, `DeleteFact(id)`
  - `GetEpisodes(limit, offset)`, `GetEpisode(id)`
  - `GetEntities()`, `GetRelationships()`
  - `SearchMemory(query, type)` — semantic or full-text
  - `GetScratchpad()`, `UpdateScratchpad(content)`
- [ ] Write frontend unit tests for each component
- [ ] Write Playwright E2E test: browse facts, search memory, view episode timeline

**Acceptance criteria:**
- Memory Explorer tab shows all sub-views
- Facts are displayed, editable, deletable
- Episode timeline renders correctly
- Entity graph renders and is interactive
- Search works (semantic and full-text)
- All tests pass

**Notes for next person:**

---

## Stage 10: Desktop GUI — Tasks, Personas, Activity Log, Settings

**Goal:** Build the remaining four tabs — Tasks & Schedule, Persona Studio, Activity Log, and Settings.

**Tasks:**
- [ ] Build `TaskManager.svelte`:
  - Upcoming reminders list with snooze/edit/cancel
  - Fired reminders section
  - Recurring schedule with toggle, human-readable cron, next fire time
  - "New Reminder" inline form with date/time picker
- [ ] Build `PersonaStudio.svelte`:
  - Persona list with cards (name, provider/model tag, active indicator)
  - Split-panel: list left, editor right
  - Model configuration form (provider, model, temperature, max_tokens)
  - Markdown editor with live preview toggle
  - Create new persona dialog
  - Persona comparison view
- [ ] Build `ActivityLog.svelte`:
  - Scrollable timeline with type icons
  - Filter chips (All, Messages, Retrievals, Functions, LLM, Consolidation, Errors)
  - Expandable entries with full JSON details
  - Prompt inspector modal with syntax-highlighted sections
  - Search bar
- [ ] Build `Settings.svelte`:
  - Provider management cards with status indicator
  - Default provider dropdown
  - Model parameter sliders
  - Telegram toggle + bot token input
  - Memory settings (db path, working memory turns, consolidation delays)
  - Data management (export/import/clear)
  - About section
- [ ] Implement Wails Go bindings for all tabs
- [ ] Implement system tray with menu and notifications
- [ ] Write frontend unit tests for each component
- [ ] Write Playwright E2E tests for key flows:
  - Create a reminder, verify it appears in task list
  - Switch persona, verify behavior change
  - Browse activity log, inspect a prompt
  - Change a setting, verify it persists

**Acceptance criteria:**
- All tabs render and are functional
- Tasks can be created, snoozed, cancelled
- Personas can be created, edited, switched
- Activity log shows entries with filtering
- Settings persist across restarts
- System tray works (minimize, notifications)
- All tests pass

**Notes for next person:**

---

## Stage 11: Telegram Interface

**Goal:** Implement the optional Telegram interface — long-polling bot that connects to the same agent core.

**Tasks:**
- [ ] Create `internal/interface/telegram/telegram.go`:
  - Implement `Interface` contract (Start, Send, Stop)
  - Long-polling via `go-telegram/bot`
  - Handle text messages and commands
  - Map Telegram chat IDs to user IDs
- [ ] Create `internal/interface/telegram/telegram_test.go`:
  - Mock Telegram API server
  - Test message receiving and sending
  - Test error handling (network issues, API errors)
- [ ] Integrate Telegram into main startup:
  - Start Telegram interface if configured
  - Support `--daemon` flag (GUI-less mode, Telegram only)
- [ ] Add cross-interface awareness:
  - Telegram messages show in GUI chat with Telegram icon
  - GUI messages are accessible from Telegram context
- [ ] Write integration test: send message via Telegram mock, verify it reaches agent

**Acceptance criteria:**
- Telegram bot connects and responds to messages
- Messages from Telegram appear in GUI chat history
- Agent has access to full memory across interfaces
- `--daemon` mode works (no GUI, Telegram only)
- All tests pass

**Notes for next person:**

---

## Stage 12: `remy init` Command & First-Run Experience

**Goal:** Implement the `remy init` command that walks users through setup, and the first-run experience when starting without init.

**Tasks:**
- [ ] Create `cmd/remy/init.go`:
  - Check for Ollama (or configured provider)
  - Check for required models
  - Create `~/.remy/` directory structure
  - Generate default `config.json`
  - Generate default persona (`personas/default.md`)
  - Print next steps
- [ ] Implement first-run detection:
  - If `~/.remy/` doesn't exist on startup, auto-run init checks
  - Guide user through missing dependencies interactively
- [ ] Write tests for init logic (mock file system, mock Ollama check)
- [ ] Update `main.go` to handle `init` subcommand and first-run flow

**Acceptance criteria:**
- `remy init` creates all necessary files and directories
- First-run without init auto-detects and guides user
- All tests pass

**Notes for next person:**

---

## Stage 13: Polish, Edge Cases & Performance

**Goal:** Hardening pass — error handling, edge cases, performance optimization, and UX polish.

**Tasks:**
- [ ] Error handling audit:
  - All errors are properly wrapped with context
  - No panics in production code paths
  - Graceful degradation when LLM is down
  - Database connection errors are recoverable
- [ ] Edge case testing:
  - Very long messages (context window overflow)
  - Rapid-fire messages (concurrent handling)
  - Empty messages
  - Special characters in messages (Unicode, emoji, markdown injection)
  - Database file locked/permissions issues
  - Config file corruption
  - Persona file with invalid frontmatter
- [ ] Performance:
  - Profile database queries (add indexes where needed)
  - Profile LLM client (connection pooling, timeout tuning)
  - Profile frontend rendering (virtual list for large message histories)
  - Memory usage monitoring
- [ ] UX polish:
  - Loading states for all async operations
  - Error toasts for failures
  - Keyboard shortcuts (Cmd+Enter, Cmd+K, Cmd+,, Escape, Tab)
  - Dark mode / light mode (follow system preference)
  - Responsive window resizing
  - Accessibility (ARIA labels, keyboard navigation, focus management)
- [ ] Write additional tests for all edge cases
- [ ] Run full test suite and verify 80%+ coverage

**Acceptance criteria:**
- Full test suite passes
- Coverage >= 80% on `internal/`
- No known panics or crashes
- App feels responsive and polished

**Notes for next person:**

---

## Stage 14: CI/CD & Release Pipeline

**Goal:** Set up automated releases — build binaries for all platforms, run full test suite, publish GitHub releases.

**Tasks:**
- [ ] Update `.github/workflows/ci.yml`:
  - Add Playwright E2E test step (requires building the binary first)
  - Add integration test step (requires Ollama — use `nomic-ai/ollama` GitHub Action)
  - Add coverage reporting (upload to Codecov or similar)
  - Add linting (golangci-lint for Go, eslint for frontend)
- [ ] Create `.github/workflows/release.yml`:
  - Triggered on tag push (`v*`)
  - Build for macOS (arm64, amd64), Linux (amd64, arm64), Windows (amd64)
  - Run full test suite
  - Create GitHub release with binaries attached
  - Generate release notes from git log
- [ ] Add `Makefile` targets: `release`, `release-dry-run`
- [ ] Add version injection via `-ldflags`
- [ ] Document release process in CONTRIBUTING.md

**Acceptance criteria:**
- CI passes on every push
- Release workflow produces binaries for all platforms
- GitHub release is created with binaries and release notes
- Playwright tests run in CI

**Notes for next person:**

---

## Stage 15: Documentation & Contributing Guide

**Goal:** Write comprehensive documentation for users and contributors.

**Tasks:**
- [ ] Write `README.md`:
  - What is Remy?
  - Quick start (install, init, chat)
  - Features overview
  - Screenshots
  - Links to detailed docs
- [ ] Write `CONTRIBUTING.md`:
  - How to set up the dev environment
  - How to run tests
  - How to build and release
  - Code style guide (reference ARCHITECTURE.md §11)
  - How to submit PRs
- [ ] Write `docs/telegram.md` — how to set up Telegram bot
- [ ] Write `docs/personas.md` — how to create and customize personas
- [ ] Write `docs/memory.md` — how memory works (for curious users)
- [ ] Add help text to CLI (`remy --help`)

**Acceptance criteria:**
- README is clear and complete
- CONTRIBUTING.md has enough detail for a new contributor
- All CLI commands have help text
- Documentation is consistent with the actual behavior

**Notes for next person:**

---

## Summary of Testing Strategy

| Test Type | What | Tool | When |
|-----------|------|------|------|
| Go unit tests | All `internal/` packages | `go test` | Every push (CI) |
| Frontend unit tests | Svelte components | Vitest + @testing-library/svelte | Every push (CI) |
| Go integration tests | LLM client, database | `go test -tags=integration` | CI (with Ollama) |
| E2E tests | Full app flows | Playwright | CI (on main, with built binary) |
| Coverage check | 80%+ on `internal/` | `go test -cover` | Every push (CI) |

## Current Status

| Stage | Status | Started | Completed | Notes |
|-------|--------|---------|-----------|-------|
| 1. Project Scaffolding | [x] | 2026-08-01 | 2026-08-01 | CLI binary, config, frontend skeleton, CI, Makefile all working |
| 2. SQLite Database Layer | [ ] | — | — | |
| 3. LLM Client & Provider | [ ] | — | — | |
| 4. Agent Core Loop | [ ] | — | — | |
| 5. Persona System | [ ] | — | — | |
| 6. Consolidation Engine | [ ] | — | — | |
| 7. Scheduler & Tasks | [ ] | — | — | |
| 8. GUI — Chat & Core UI | [ ] | — | — | |
| 9. GUI — Memory Explorer | [ ] | — | — | |
| 10. GUI — Tasks, Personas, Activity, Settings | [ ] | — | — | |
| 11. Telegram Interface | [ ] | — | — | |
| 12. `remy init` & First-Run | [ ] | — | — | |
| 13. Polish & Edge Cases | [ ] | — | — | |
| 14. CI/CD & Release Pipeline | [ ] | — | — | |
| 15. Documentation | [ ] | — | — | |
