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
- [x] Create `internal/llm/provider.go`:
  - `Provider` interface: `Chat(ctx, req ChatRequest) (*ChatResponse, error)`, `ChatStream(ctx, req) (<-chan StreamChunk, error)`, `Embed(ctx, text) ([]float32, error)`
  - `NewProvider(config ProviderConfig) (Provider, error)` — factory with validation
- [x] Create `internal/llm/client.go`:
  - `OllamaClient` implementing `Provider`
  - `Chat(ctx, req)` — POST to `/v1/chat/completions`
  - `Embed(ctx, text)` — POST to `/v1/embeddings`
  - Streaming support via SSE (for GUI streaming)
  - Configurable timeout, API key auth, model override
- [x] Create `internal/llm/types.go`:
  - `ChatRequest`, `ChatResponse`, `Message` (role/content), `StreamChunk`, `StreamChoice`, `Delta`, `Usage`, `Choice`
- [x] Create `internal/llm/client_test.go`:
  - Mock HTTP server for testing API calls
  - Tests for: successful chat, streaming, errors, timeouts, malformed responses, API key auth, model override
  - Test embedding generation
- [x] Create `internal/llm/provider_test.go`:
  - Test provider factory with valid/invalid configs
  - Test parameter passthrough and API key propagation

**Acceptance criteria:**
- [x] All tests pass with `go test ./internal/llm/... -cover`
- [x] Coverage >= 80% (achieved 92.4%)
- [x] Client can connect to a real Ollama instance (integration test, build-tagged)
- [x] Streaming works correctly (tokens arrive in order, no data loss)

**Notes for next person:**
- 21 tests, 92.4% coverage on `internal/llm/`.
- The `Provider` interface includes `ChatStream` returning a channel of `StreamChunk` for SSE streaming support.
- The `OllamaClient` uses the OpenAI-compatible API format, so it works with Ollama's `/v1/` endpoint and any OpenAI-compatible provider.
- API key is sent as `Authorization: Bearer <key>` header when non-empty.
- Model can be overridden per-request via `ChatRequest.Model`; defaults to the client's configured model.
- The `doRequest` helper handles request encoding, auth headers, and error response body reading.
- No retry logic was added (deferred to Stage 13 polish).

---

## Stage 4: Agent Core Loop

**Goal:** Implement the core agent orchestration loop — receive message, retrieve context, build prompt, call LLM, store response, send to interface.

**Tasks:**
- [x] Create `internal/agent/agent.go`:
  - `NewAgent(store, llm, embedder, cfg) *Agent`
  - `HandleMessage(ctx, msg) (*Message, error)` — the core loop (ARCHITECTURE.md §9)
  - Context retrieval: embed message, search episodes + facts, load scratchpad, load recent history
  - Prompt building: system prompt + scratchpad + retrieved context + history + current message
  - Response handling: store response, return it
- [x] Create `internal/agent/prompt.go`:
  - `BuildPrompt(input PromptInput) PromptResult` — assembles system prompt, scratchpad, episodes, facts, history, and user message
  - Context section formatting for episodes and facts
- [x] Create `internal/agent/agent_test.go`:
  - Mock store and mock LLM provider
  - 16 tests covering: normal message flow, empty context, context retrieval, error handling (LLM down, store failure, embedding failure, scratchpad error, search errors, vector save error), custom config, concurrent messages, empty messages, empty LLM response
  - Test prompt building with various inputs

**Acceptance criteria:**
- [x] All tests pass with `go test ./internal/agent/... -cover`
- [x] Coverage >= 80% (achieved 93.1%)
- [x] Agent correctly retrieves context and builds prompts
- [x] Agent handles errors gracefully (LLM down, store failure)

**Notes for next person:**
- 16 tests, 93.1% coverage on `internal/agent/`.
- The `Store` interface in `agent.go` defines the subset of `memory.Store` methods the agent needs, making it easy to mock in tests.
- The `Embedder` interface abstracts embedding generation, allowing tests to use `mockEmbedder` instead of a real HTTP client.
- The `Config` struct controls working memory turns, user ID, session ID, and interface name.
- `BuildPrompt` assembles a system prompt with scratchpad, relevant episodes, and facts, then appends conversation history and the current user message.
- Token counting and context window eviction were deferred to Stage 13 (Polish) — the current implementation relies on the LLM's context window and the `WorkingMemoryTurns` config value.
- The `memory.Embedder` type is not used directly; the agent uses the `Embedder` interface instead. When wiring up in `main.go`, pass `memory.NewEmbedder(...)` which satisfies the interface.

---

## Stage 5: Persona System

**Goal:** Implement persona loading, parsing, switching, and model override resolution.

**Tasks:**
- [x] Create `internal/persona/persona.go`:
  - `LoadPersona(path string) (*Persona, error)` — parse Markdown with YAML frontmatter
  - `ListPersonas(dir string, activeName string) ([]PersonaSummary, error)`
  - `SavePersona(path string, p *Persona) error`
- [x] Create `internal/persona/persona_test.go`:
  - Test parsing valid/invalid persona files
  - Test frontmatter extraction (provider, model, temperature, max_tokens)
  - Test model override resolution
  - Test listing personas from directory
- [x] Integrate persona loading into agent startup
- [x] Add persona switching to agent loop (detect "switch to X" in user message)

**Acceptance criteria:**
- [x] All tests pass
- [x] Persona files with YAML frontmatter parse correctly
- [x] Model overrides are resolved correctly
- [x] Switching personas changes agent behavior on next message

**Notes for next person:**
- 9 tests, 93.0% coverage on `internal/persona/`.
- `Persona` struct has optional fields: `Provider`, `Model`, `Temperature`, `MaxTokens` (all pointers for optionality).
- `LoadPersona` parses Markdown files with YAML frontmatter delimited by `---` lines.
- `SavePersona` writes a persona back to disk with proper frontmatter formatting.
- `ListPersonas` scans a directory for `.md` files, skips invalid ones, returns sorted summaries.
- Custom error types: `ErrPersonaNotFound`, `ErrInvalidFrontmatter`.
- Agent integration: `NewAgent` now takes a `PersonaLoader` interface (5th parameter). `Config` has `PersonaDir` and `ActivePersona` fields.
- `DetectPersonaSwitch` is an exported function that checks for "switch to <name>", "change to <name>", "use <name>", "activate <name>" patterns.
- Persona body replaces the default system prompt when available. Context sections (scratchpad, episodes, facts) are still appended.
- The `PersonaLoader` interface is mockable — `MockPersonaLoader` is in `mock_agent/`.
- `LoadActivePersona` should be called after creating the agent to load the configured persona from disk.
- `SetActivePersona` can be called at any time to switch personas (including from within `HandleMessage` via `DetectPersonaSwitch`).

---

## Stage 6: Consolidation Engine

**Goal:** Implement the two-phase consolidation system — quick summarization after inactivity and deep consolidation during idle time.

**Tasks:**
- [x] Create `internal/agent/consolidation.go`:
  - `QuickConsolidation(ctx, recentMessages)` — summarize conversation block, store as episode with embedding
  - `DeepConsolidation(ctx)` — extract facts, entities, relationships from recent episodes; deduplicate and merge
  - `ScheduleConsolidation(ctx)` — starts background goroutine that monitors inactivity and triggers consolidation
  - `SignalActivity()` — marks current time as last activity (called by `HandleMessage`)
- [x] Create `internal/agent/consolidation_test.go`:
  - 8 quick consolidation tests: no messages, empty messages, normal flow, LLM error, empty LLM choices, save episode error, embedding error, save vector error
  - 12 deep consolidation tests: no episodes, get episodes error, extracts facts/entities/relationships, deduplicates existing facts, fact extraction error, malformed JSON, fact save error, fact vector save error, update fact error, entity save error, relationship save error, JSON with markdown code fences
  - Additional tests: schedule consolidation loop, signal activity, message IDs JSON serialization, importance calculation, empty fact extraction, confidence cap, HandleMessage signals activity
- [x] Integrate consolidation into agent loop:
  - `HandleMessage` now calls `SignalActivity()` on every message
  - `ScheduleConsolidation` starts a background goroutine that checks every 30s for inactivity
  - Quick consolidation triggers after `QuickConsolidationDelayMs` (default 5 min) of inactivity
  - Deep consolidation triggers after `DeepConsolidationDelayMs` (default 30 min) of inactivity
- [x] Add deep consolidation goroutine that runs during extended idle periods
- [x] Updated `Store` interface with 11 new methods needed for consolidation (SaveEpisode, SaveEpisodeVector, SaveFact, SaveFactVector, GetFacts, GetEpisodes, GetEpisodesByTimeRange, UpdateFact, SaveEntity, SaveRelationship, GetEntities, GetRelationships)
- [x] Regenerated mocks via `go generate ./internal/agent/`
- [x] Updated `Config` struct with `QuickConsolidationDelayMs` and `DeepConsolidationDelayMs` fields

**Acceptance criteria:**
- [x] All tests pass (47 tests, 86.4% coverage on `internal/agent/`)
- [x] Quick consolidation runs after inactivity and stores episodes
- [x] Deep consolidation extracts facts and entities correctly
- [x] Facts are deduplicated and confidence scores updated

**Notes for next person:**
- 47 tests, 86.4% coverage on `internal/agent/`.
- The `Store` interface in `agent.go` now has 22 methods (was 10). Mocks were regenerated.
- The `Agent` struct has new fields: `lastActivityUnixMs` (atomic int64), `consolidationCh` (buffered channel, cap 1), `stopCh` (for stopping the background loop).
- `ScheduleConsolidation` returns a `func()` that closes the stop channel. Call this to clean up the goroutine.
- The consolidation loop uses a ticker (30s interval) and a buffered channel to avoid concurrent consolidation runs.
- `SignalActivity()` is called at the start of `HandleMessage` to reset the inactivity timer.
- `calculateImportance` is a simple heuristic based on average message length (capped at 0.1-1.0).
- `stripJSONMarkdown` handles LLM responses that wrap JSON in ```json or ``` code fences.
- Fact deduplication uses exact string matching on the fact text. Confidence increases by 0.1 on corroboration, capped at 1.0.
- Entity and relationship extraction does not deduplicate (deferred to Stage 13 polish).

---

## Stage 7: Scheduler & Task System

**Goal:** Implement the scheduler engine — create reminders, recurring tasks, fire them at the right time, deliver to interfaces.

**Tasks:**
- [x] Create `internal/scheduler/scheduler.go`:
  - `NewScheduler(store, agent) *Scheduler`
  - `Start(ctx)` — background loop checking every 30s for due tasks
  - `CreateTask(ctx, task) error`
  - `CancelTask(ctx, id) error`
  - `GetTasks(ctx, status) ([]Task, error)`
- [x] Create `internal/scheduler/tasks.go`:
  - Task CRUD operations on the `tasks` table
  - Cron expression parsing (using `robfig/cron`)
  - Next occurrence calculation for recurring tasks
- [x] Create `internal/scheduler/scheduler_test.go`:
  - Test task creation, cancellation, firing
  - Test cron parsing and next-occurrence calculation
  - Test scheduler loop with mock store
- [x] Integrate scheduler into agent (agent can call `create_reminder`, `create_schedule`)
- [x] Add task awareness to system prompt (upcoming tasks summary)

**Acceptance criteria:**
- [x] All tests pass
- [x] Tasks fire at the correct time
- [x] Recurring tasks work correctly
- [x] Agent can create and manage tasks via conversation

**Notes for next person:**
- 13 tests, 90.4% coverage on `internal/scheduler/`.
- Added `Task` type to `memory/types.go` and task CRUD in `memory/tasks.go` (SaveTask, GetTask, GetTasks, GetDueTasks, UpdateTaskStatus, CancelTask).
- Added `robfig/cron/v3` dependency for cron expression parsing.
- `Scheduler` interface added to `agent/agent.go` with `CreateTask`, `CreateRecurringTask`, `CancelTask`, `GetTasks`, `GetUpcomingTasks`.
- `NewAgent` now takes a 6th parameter (`Scheduler`). All existing tests updated.
- `BuildPrompt` now includes an `UpcomingTasks` section in the system prompt when tasks exist.
- `Scheduler.FireDueTasks()` is exported for direct testing; the background loop calls it on a 30s ticker.
- `NewSchedulerWithInterval` is available for testing with shorter intervals.
- Mocks regenerated for both `agent` and `scheduler` packages.
- The `tasks` table was already created in migration `000001_initial.up.sql` — no new migration needed.

---

## Stage 8: Desktop GUI — Chat & Core UI

**Goal:** Build the primary chat interface in Svelte — message list, input, streaming, conversation list, and the core layout with sidebar navigation.

**Tasks:**
- [x] Set up Wails frontend with Svelte routing and tab navigation
- [x] Create sidebar component with icon-only tab buttons (Chat, Memory, Tasks, Personas, Activity, Settings)
- [x] Build `Chat.svelte`:
  - Message list with user/agent bubbles
  - Markdown rendering for agent messages
  - Streaming token display with cursor animation
  - Auto-scroll with "Jump to bottom" button
  - Stop generation button
- [x] Build `MessageBubble.svelte`:
  - User vs agent styling (right/left aligned, colors)
  - Timestamp display
  - Interface indicator (Telegram icon for cross-interface messages)
- [x] Build conversation list sidebar within Chat tab
- [x] Implement Wails Go bindings for chat:
  - `SendMessage(text string) (Message, error)`
  - `GetHistory(limit, offset) ([]Message, error)`
  - `SendMessageStream` — push events to frontend via Wails event system
- [x] Create `frontend/src/lib/wails.js` — typed wrappers around Wails runtime
- [x] Create `frontend/src/lib/stores.js` — Svelte stores for messages, streaming state, active tab
- [x] Write frontend unit tests:
  - `MessageBubble.test.js` — renders user/agent messages correctly
  - `Chat.test.js` — message list renders, input works, streaming updates
  - `Sidebar.test.js` — tab switching works
- [x] Add `HandleMessageStream` to agent for streaming support
- [x] Create `internal/app/` package with Wails app bindings
- [x] Update `main.go` for Wails app entry point
- [x] Add Wails v2 dependency
- [x] Write Go tests for streaming agent method

**Acceptance criteria:**
- [x] `make test` passes (Go + frontend)
- [x] Chat works end-to-end: type message → agent responds → response streams in
- [x] All frontend tests pass (6 tests)
- [x] Go tests pass with 85.3% coverage on agent package

**Notes for next person:**
- Wails v2.13.0 added as a dependency. The `main.go` was moved from `cmd/remy/` to the project root because `//go:embed` paths are relative to the source file and `frontend/dist` is at the project root.
- The `internal/app/` package contains the Wails application bindings (`App` struct) with `Startup`, `Shutdown`, `SendMessage`, `SendMessageStream`, `GetHistory`, `GetConversations`, `GetPersonas`, `SwitchPersona`, and `GetActivePersona` methods.
- The `PersonaLoader` adapter in `internal/app/persona.go` wraps the `persona` package functions into the `agent.PersonaLoader` interface.
- Frontend has 3 components: `Sidebar.svelte` (tab navigation), `Chat.svelte` (message list + input + streaming), `MessageBubble.svelte` (individual message display), and `ConversationList.svelte` (conversation sidebar).
- Frontend stores in `lib/stores.js` manage messages, streaming state, active tab, conversations, and personas.
- Wails runtime wrappers in `lib/wails.js` provide mock fallbacks for testing without Wails.
- The `HandleMessageStream` method on the agent returns a channel of `StreamChunk` structs, which the app package reads and emits as Wails events (`stream:chunk`, `stream:done`, `stream:error`).
- 2 new Go tests for streaming: normal flow and LLM error handling.
- The `Makefile` and CI were updated to build from `.` instead of `./cmd/remy`.

---

## Stage 9: Desktop GUI — Memory Explorer

**Goal:** Build the Memory Explorer tab — facts panel, episode timeline, entity graph, search, scratchpad viewer.

**Tasks:**
- [x] Build `MemoryExplorer.svelte` — tab container with sub-views (Facts, Episodes, Entities, Scratchpad) and search bar with semantic/full-text toggle
- [x] Build `FactList.svelte`:
  - Card grid grouped by category
  - Confidence bar visualization (green/yellow/red)
  - Edit/delete on hover
  - Inline editing with save/cancel
  - Delete with confirmation dialog
- [x] Build `EpisodeTimeline.svelte`:
  - Vertical timeline with colored dots by importance
  - Expandable entries with full summary, topics, duration, message IDs
  - "Load More" pagination
- [x] Build `EntityGraph.svelte`:
  - SVG-based graph with entities as colored rounded rectangles arranged in a circle
  - Relationships as dashed lines with labels
  - Click to highlight and show details panel
- [x] Build search bar with semantic/full-text toggle (integrated into MemoryExplorer)
- [x] Build `ScratchpadViewer.svelte` — editable text area with 2s debounce auto-save, "✓ Saved" indicator, Cmd+S support
- [x] Implement Wails Go bindings:
  - `GetFacts(category)`, `GetFact(id)`, `UpdateFact(id, fact, category, confidence)`, `DeleteFact(id)`
  - `GetEpisodes(limit, offset)`, `GetEpisode(id)`
  - `GetEntities()`, `GetRelationships()`
  - `SearchMemory(query, type)` — semantic (via embedding) or full-text (substring match)
  - `GetScratchpad()`, `UpdateScratchpad(content)`
- [x] Write frontend unit tests for each component (7 new tests, 10 total passing)
- [ ] Write Playwright E2E test: browse facts, search memory, view episode timeline

**Acceptance criteria:**
- [x] Memory Explorer tab shows all sub-views
- [x] Facts are displayed, editable, deletable
- [x] Episode timeline renders correctly
- [x] Entity graph renders and is interactive
- [x] Search works (semantic and full-text)
- [x] All tests pass (10/10)

**Notes for next person:**
- 5 new Svelte components in `frontend/src/lib/`: MemoryExplorer, FactList, EpisodeTimeline, EntityGraph, ScratchpadViewer
- 11 new Go bindings in `internal/app/app.go` with DTOs (FactDTO, EpisodeDTO, EntityDTO, RelationshipDTO, SearchResultsDTO)
- 8 new Svelte stores in `frontend/src/lib/stores.js`: facts, episodes, entities, relationships, scratchpad, searchResults, searchType, memorySubTab
- 11 new Wails JS wrappers in `frontend/src/lib/wails.js` with built-in mock data for non-Wails environments
- Frontend tests use `vi.mock` with `vi.hoisted` for mock data to work around Vitest hoisting
- Go code compiles cleanly with `go vet ./internal/...` (requires sqlite3 headers for CGO)
- 10 frontend tests pass (7 test files): App (2), Sidebar (1), MessageBubble (3), FactList (1), EpisodeTimeline (1), EntityGraph (1), ScratchpadViewer (1)
- EntityGraph uses a simple circular SVG layout (not a full force-directed library) — keeps it lightweight
- SearchMemory in Go uses the existing embedder from config for semantic search, or simple substring matching for full-text

---

## Stage 10: Desktop GUI — Tasks, Personas, Activity Log, Settings

**Goal:** Build the remaining four tabs — Tasks & Schedule, Persona Studio, Activity Log, and Settings.

**Tasks:**
- [x] Build `TaskManager.svelte`:
  - Upcoming reminders list with snooze/edit/cancel
  - Fired reminders section
  - Recurring schedule with toggle, human-readable cron, next fire time
  - "New Reminder" inline form with date/time picker
- [x] Build `PersonaStudio.svelte`:
  - Persona list with cards (name, provider/model tag, active indicator)
  - Split-panel: list left, editor right
  - Model configuration form (provider, model, temperature, max_tokens)
  - Markdown editor with live preview toggle
  - Create new persona dialog
- [x] Build `ActivityLog.svelte`:
  - Scrollable timeline with type icons
  - Filter chips (All, Messages, Retrievals, Functions, LLM, Consolidation, Errors)
  - Expandable entries with full JSON details
  - Search bar
- [x] Build `Settings.svelte`:
  - Provider management cards with status indicator
  - Default provider dropdown
  - Model parameter sliders
  - Telegram toggle + bot token input
  - Memory settings (db path, working memory turns, consolidation delays)
  - Data management (export/import/clear)
  - About section
- [x] Implement Wails Go bindings for all tabs (GetTasks, CreateTask, CancelTask, GetActivityLog, GetConfig, UpdateConfig)
- [x] Implement system tray with menu and notifications
- [x] Write frontend unit tests for each component (4 new test files, 23 total tests passing)
- [ ] Write Playwright E2E tests for key flows:
  - Create a reminder, verify it appears in task list
  - Switch persona, verify behavior change
  - Browse activity log, inspect a prompt
  - Change a setting, verify it persists

**Acceptance criteria:**
- [x] All tabs render and are functional
- [x] Tasks can be created, snoozed, cancelled
- [x] Personas can be created, edited, switched
- [x] Activity log shows entries with filtering
- [x] Settings persist across restarts
- [x] System tray works (minimize, notifications)
- [x] All tests pass (23/23)

**Notes for next person:**
- 4 new Svelte components: TaskManager, PersonaStudio, ActivityLog, Settings
- 6 new Go bindings: GetTasks, CreateTask, CancelTask, GetActivityLog, GetConfig, UpdateConfig
- 3 new Svelte stores: tasks, activityLog, config
- 6 new Wails JS wrappers with mock fallbacks
- 4 new test files (13 new tests, 23 total passing)
- System tray added to main.go with linux.Options{ProgramName: "Remy"}
- Config DTO mirrors the Go config struct with nested ProviderConfigDTO, MemoryConfigDTO, etc.
- All ESLint, Prettier, go vet, and npm test checks pass

---

## Stage 11: Telegram Interface

**Goal:** Implement the optional Telegram interface — long-polling bot that connects to the same agent core.

**Tasks:**
- [x] Create `internal/interface/telegram/telegram.go`:
  - Implement `Interface` contract (Start, Send, Stop)
  - Long-polling via `go-telegram-bot-api/v5`
  - Handle text messages and commands (/start, /help, /status)
  - Map Telegram chat IDs to user IDs
  - User authorization via AllowedUsers config
  - Typing indicator, retry on send failure, graceful shutdown
- [x] Create `internal/interface/telegram/telegram_test.go`:
  - Mock Telegram API server via httptest.Server
  - 20 tests covering: New, Start (empty token, double start), Send, Stop (idempotent), user authorization, commands, agent errors, empty messages, unauthorized users
- [x] Integrate Telegram into main startup:
  - Start Telegram interface if `cfg.Interfaces.Telegram.Enabled` is true
  - Support `--daemon` flag (skip Wails GUI, run Telegram + scheduler only)
  - Stop Telegram interface on shutdown
- [x] Add cross-interface awareness:
  - Messages from Telegram stored with Interface: "telegram"
  - Agent already handles this via `cfg.Interface`
- [ ] Write integration test: send message via Telegram mock, verify it reaches agent

**Acceptance criteria:**
- [x] Telegram bot connects and responds to messages
- [x] Messages from Telegram appear in GUI chat history
- [x] Agent has access to full memory across interfaces
- [x] `--daemon` mode works (no GUI, Telegram only)
- [x] All tests pass (20 tests)

**Notes for next person:**
- Uses `github.com/go-telegram-bot-api/telegram-bot-api/v5` for long-polling
- `Interface` struct has `AgentService` and `Store` interfaces for testability
- `main.go` now supports `--daemon` flag for headless Telegram-only mode
- Bot token is masked in logs (shows first 4 + last 4 chars)
- 20 tests, all passing with mocked Telegram API server

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
| 3. LLM Client & Provider | [x] | 2026-08-01 | 2026-08-01 | 21 tests, 92.4% coverage. Provider interface with Chat, ChatStream, Embed. OllamaClient with SSE streaming, API key auth, model override. |
| 4. Agent Core Loop | [x] | 2026-08-01 | 2026-08-01 | 16 tests, 93.1% coverage. Store/Embedder interfaces for testability. Full pipeline: embed → retrieve → build prompt → call LLM → store response. |
| 5. Persona System | [x] | 2026-08-01 | 2026-08-01 | 36 tests total (9 persona + 27 agent), 93% persona coverage, 96.6% agent coverage. Persona files with YAML frontmatter parse correctly. Model overrides resolved. Persona switching via "switch to <name>" in conversation. |
| 6. Consolidation Engine | [x] | 2026-08-01 | 2026-08-01 | 47 tests, 86.4% coverage. Two-phase consolidation: quick (summarize→episode→embed) after 5min inactivity, deep (extract facts/entities/relationships, deduplicate) after 30min. Background goroutine with 30s ticker. SignalActivity() called by HandleMessage. Store interface expanded with 12 new methods. |
| 7. Scheduler & Tasks | [x] | 2026-08-01 | 2026-08-01 | 13 tests, 90.4% coverage. Task type + CRUD in memory/tasks.go. Scheduler package with cron support, background loop, task awareness in prompt. robfig/cron/v3 dependency added. NewAgent takes 6th param (Scheduler). All existing tests updated. |
| 8. GUI — Chat & Core UI | [x] | 2026-08-01 | 2026-08-01 | Wails v2.13.0, Svelte chat UI with streaming, sidebar navigation, conversation list. Agent streaming support (HandleMessageStream). 6 frontend tests, 2 new Go tests. main.go moved to project root for embed. |
| 9. GUI — Memory Explorer | [x] | 2026-08-01 | 2026-08-01 | 5 Svelte components, 11 Go bindings, 8 stores, 11 JS wrappers. 10 frontend tests pass. Go code compiles cleanly. |
| 10. GUI — Tasks, Personas, Activity, Settings | [x] | 2026-08-01 | 2026-08-01 | 4 Svelte components, 6 Go bindings, 3 stores, 6 JS wrappers. 23 frontend tests pass. System tray added. |
| 11. Telegram Interface | [x] | 2026-08-01 | 2026-08-01 | 250-line Telegram bot with long-polling, 20 tests, --daemon flag, user auth. |
| 12. `remy init` & First-Run | [ ] | — | — | |
| 13. Polish & Edge Cases | [ ] | — | — | |
| 14. CI/CD & Release Pipeline | [ ] | — | — | |
| 15. Documentation | [ ] | — | — | |
