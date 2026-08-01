# Remy — Personal Assistant AI Architecture

**Name:** Remy (from Latin *rememorari* — to remember)

## 1. Design Philosophy

The goal is an agent that feels like a person you message — it remembers your conversation naturally, picks up where you left off, and gets to know you over time. The user should never have to think about "sessions," "memory management," or "context windows." The agent handles all of that internally.

**Core principles:**
- **Invisible infrastructure** — the user just sends and receives messages
- **Continuous relationship** — the agent builds a persistent model of the user across all interactions
- **Biologically inspired** — memory architecture mirrors how human memory works (working, episodic, semantic, consolidation)
- **Local-first** — everything runs on the user's machine, no cloud dependencies
- **Zero external dependencies** — the binary is self-contained. No server processes, no databases to install, no runtimes. SQLite (with built-in vector support) is the only storage layer — it's just a file on disk.

### 1.1 How Remy Is Different

Most AI agents (AutoGPT, Claude Code, opencode, etc.) are designed to *do things* — run commands, edit files, browse the web, interact with your system. This makes them powerful but also introduces risk. They need sandboxing, approval flows, and careful trust management.

Remy takes the opposite approach: **it does nothing but chat.** It has no shell access, no file system access, no web access, no ability to execute code. It exists in a bubble and only knows what you tell it. This makes it:

- **Safe to run on any machine** — no sandboxing needed, no approval flows, no risk of unintended actions
- **Simple to reason about** — it can't surprise you by doing something you didn't ask for
- **Focused on what matters** — conversation, memory, and relationship, not tool orchestration

The trade-off is deliberate: Remy is not a productivity agent that does your work for you. It's a conversational companion that remembers who you are and what matters to you. If you want it to do something, you tell it to remind you later — and it will.

---

## 2. Onboarding & Installation

### 2.1 Prerequisites

- **Ollama** running locally with at least one chat model (e.g., `llama3.1:8b`) and an embedding model (e.g., `nomic-embed-text`)
- macOS, Linux, or Windows

**That's it.** Remy itself is a single binary with zero external dependencies. No database server, no runtime, no package manager. SQLite (with vector support) is compiled directly into the binary — the database is just a file on disk.

### 2.2 Installation

Remy is distributed as a single pre-built binary for each platform:

```
# macOS (ARM64)
curl -L -o remy https://github.com/yourname/remy/releases/latest/download/remy-darwin-arm64
chmod +x remy
mv remy /usr/local/bin/

# macOS (AMD64), Linux, Windows — similar per-platform binaries
```

### 2.3 Provider Model

Remy supports multiple LLM providers configured in `config.json`. Each provider has its own endpoint, auth, and model settings. The `default_provider` field specifies which one to use when a persona doesn't override it.

```json
{
  "providers": {
    "ollama": {
      "endpoint": "http://localhost:11434/v1",
      "api_key": "",
      "chat_model": "llama3.1:8b",
      "embedding_model": "nomic-embed-text",
      "parameters": {
        "temperature": 0.7,
        "max_tokens": 4096
      }
    }
  },
  "default_provider": "ollama"
}
```

**Initially only Ollama is supported.** The provider abstraction is designed so that adding OpenAI or Anthropic later is just adding a new entry to `config.json` and a new provider module — the rest of Remy doesn't change.

**Recommended local models (Ollama):**
- **Chat:** `llama3.1:8b` — good balance of quality and speed on consumer hardware. For weaker hardware (Raspberry Pi, 8GB RAM), `llama3.2:3b` or `qwen2.5:7b` (4-bit quantized).
- **Embeddings:** `nomic-embed-text` — small, fast, good quality. Works on any hardware that can run Ollama.

This means Remy itself can run on very light hardware (Raspberry Pi, old laptop) while the LLM runs elsewhere — locally on a beefier machine, or in the cloud.

### 2.4 First Run — `remy init`

```
$ remy init
```

The init command walks the user through setup:

1. **Checks for Ollama** (default) — verifies Ollama is running at `http://localhost:11434`. If not, prints instructions for installing and starting it. If using a different provider, the user can skip this and configure manually.
2. **Checks for required models** — verifies the configured chat model and embedding model are pulled. If missing, prints the `ollama pull` commands needed.
3. **Creates directory structure** — `~/.remy/` with `memory.db`, `config.json`, and `personas/default.md`.
4. **Generates a default persona** — a friendly, helpful assistant persona that the user can edit.
5. **Generates a default config** — sensible defaults for the LLM endpoint, model, and interfaces.
6. **Prints next steps** — how to start Remy, how to connect Telegram, how to customize the persona.

### 2.4 Starting Remy

```
$ remy                    # Opens the desktop GUI window
$ remy --daemon           # Starts in background (Telegram only, no GUI)
```

On first start without `init`, Remy auto-runs the init checks and guides the user through any missing dependencies.

### 2.5 Telegram Setup

To connect Telegram, the user:
1. Creates a bot via [@BotFather](https://t.me/BotFather) on Telegram
2. Sets the token in `~/.remy/config.json` or via `REMY_TELEGRAM_BOT_TOKEN` env var
3. Starts Remy with `remy` (GUI) or `remy --daemon` (background, Telegram only)

### 2.6 File Layout

```
~/.remy/
├── memory.db              # SQLite database (messages, episodes, facts, vectors)
├── config.json            # User configuration
├── personas/
│   └── default.md         # Active persona (user-editable Markdown)
└── logs/
    └── remy.log           # Application log
```

---

## 3. High-Level Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                     Interfaces Layer                         │
│  ┌──────────────────┐  ┌──────────┐                         │
│  │  Desktop GUI      │  │ Telegram  │                         │
│  │  (Wails + Svelte) │  │  (optional)│                        │
│  └────────┬─────────┘  └────┬─────┘                         │
│           │                  │                                │
└───────────┼──────────────────┼───────────────────────────────┘
            │                  │
            ▼                  ▼
┌─────────────────────────────────────────────────────────────┐
│                    Message Bus (Go channels)                 │
│  All interfaces produce/consume the same message format      │
│  ┌──────────────────────────────────────────────────────┐    │
│  │  Message{ID, UserID, Text, Timestamp, InterfaceType} │    │
│  └──────────────────────────────────────────────────────┘    │
└─────────────────────────┬───────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────────┐
│                    Agent Core                                │
│                                                              │
│  ┌──────────────────────────────────────────────────────┐   │
│  │              Orchestration Loop                        │   │
│  │  ┌─────────┐  ┌─────────┐  ┌─────────┐  ┌─────────┐  │   │
│  │  │ Receive │→ │ Retrieve│→ │  Build  │→ │  Send   │  │   │
│  │  │ Message │  │ Context │  │ Prompt  │  │ to LLM  │  │   │
│  │  └─────────┘  └─────────┘  └─────────┘  └────┬────┘  │   │
│  │                                                │       │   │
│  │  ┌─────────┐  ┌─────────┐  ┌─────────┐        │       │   │
│  │  │  Store  │← │ Process │← │ Receive │◄───────┘       │   │
│  │  │ Message │  │ Response│  │ Response│                │   │
│  │  └─────────┘  └─────────┘  └─────────┘                │   │
│  └──────────────────────────────────────────────────────┘   │
│                                                              │
│  ┌──────────────────────────────────────────────────────┐   │
│  │              Memory System                             │   │
│  │  ┌──────────────┐  ┌──────────────┐  ┌────────────┐  │   │
│  │  │  Working      │  │  Episodic    │  │  Semantic  │  │   │
│  │  │  Memory       │  │  Memory      │  │  Memory    │  │   │
│  │  │  (in-context) │  │  (SQLite +   │  │  (SQLite + │  │   │
│  │  │               │  │   vectors)   │  │   vectors) │  │   │
│  │  └──────────────┘  └──────────────┘  └────────────┘  │   │
│  │                                                       │   │
│  │  ┌───────────────────────────────────────────────┐    │   │
│  │  │         Consolidation Engine (background)       │    │   │
│  │  │  ┌──────────┐  ┌──────────┐  ┌─────────────┐  │    │   │
│  │  │  │ Quick    │  │ Deep     │  │ Concept     │  │    │   │
│  │  │  │ Summarize│  │ Reflect  │  │ Extraction  │  │    │   │
│  │  │  └──────────┘  └──────────┘  └─────────────┘  │    │   │
│  │  └───────────────────────────────────────────────┘    │   │
│  └──────────────────────────────────────────────────────┘   │
└─────────────────────────┬───────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────────┐
│                    LLM Interface (Ollama)                    │
│  OpenAI-compatible API → local LLM                          │
│  Configurable model, parameters, endpoint                   │
└─────────────────────────────────────────────────────────────┘
```

---

## 3. Memory System (Biologically Inspired)

### 3.1 Working Memory (In-Context)

**What it is:** The agent's "consciousness" — everything currently in the LLM's context window.

**Contents:**
- System prompt (persona, rules)
- Persistent scratchpad (hidden from user, agent-managed)
- Recent conversation history (last ~20-30 turns, sliding window)
- Retrieved episodic/semantic context for the current turn
- Current task state (if the agent is mid-task)

**Scratchpad:** A persistent, agent-managed note area that survives across conversations. The agent uses it to track:
- Ongoing goals or tasks
- Important facts it wants to remember about the current context
- Internal state (e.g., "waiting for user to send file")

The scratchpad is stored in SQLite and loaded into the system prompt at the start of every turn. The agent can read, write, and clear it.

**Context window management:**
- When the conversation exceeds the model's context window, older messages are evicted from working memory
- Before eviction, they are summarized and stored as an episodic memory entry
- The summary is kept in working memory as a compressed representation

### 3.2 Episodic Memory (What Happened)

**What it is:** Records of specific events, conversations, and experiences — analogous to human episodic memory.

**Storage:** SQLite with sqlite-vec extension for vector embeddings.

**Schema:**
```sql
-- Core message storage
CREATE TABLE messages (
    id          TEXT PRIMARY KEY,
    user_id     TEXT NOT NULL,
    role        TEXT NOT NULL,  -- 'user' | 'assistant' | 'system'
    content     TEXT NOT NULL,
    timestamp   INTEGER NOT NULL,
    interface   TEXT NOT NULL,  -- 'gui' | 'telegram'
    session_id  TEXT           -- logical grouping
);

-- Episodic summaries (compressed representations of conversation chunks)
CREATE TABLE episodes (
    id          TEXT PRIMARY KEY,
    summary     TEXT NOT NULL,
    start_time  INTEGER NOT NULL,
    end_time    INTEGER NOT NULL,
    message_ids TEXT NOT NULL,  -- JSON array of message IDs
    importance  REAL DEFAULT 0.5,  -- 0.0 to 1.0
    topics      TEXT            -- JSON array of topic tags
);

-- Vector embeddings for semantic search over episodes
CREATE VIRTUAL TABLE episode_vectors USING vec0(
    id TEXT PRIMARY KEY,
    embedding FLOAT[768]  -- dimension depends on embedding model
);

-- Vector embeddings for individual messages (for fine-grained retrieval)
CREATE VIRTUAL TABLE message_vectors USING vec0(
    id TEXT PRIMARY KEY,
    embedding FLOAT[768]
);
```

**Retrieval:** Episodic memory is queried in two ways:
1. **Semantic search** — embed the current user message, find similar past messages/episodes via vector similarity
2. **Temporal search** — find messages/episodes from a specific time range (e.g., "what did we discuss yesterday?")

### 3.3 Semantic Memory (What I Know)

**What it is:** General knowledge about the user, the world, and concepts — analogous to human semantic memory.

**Storage:** SQLite with sqlite-vec.

**Schema:**
```sql
-- Facts about the user (preferences, traits, personal info)
CREATE TABLE facts (
    id          TEXT PRIMARY KEY,
    fact        TEXT NOT NULL,
    category    TEXT NOT NULL,  -- 'preference' | 'trait' | 'personal_info' | 'habit'
    confidence  REAL DEFAULT 0.7,
    source      TEXT,           -- reference to episode or message
    created_at  INTEGER NOT NULL,
    updated_at  INTEGER NOT NULL,
    -- e.g. { "fact": "User prefers async/await over callbacks",
    --        "category": "preference", "confidence": 0.9 }
);

-- Entities (people, places, things, concepts)
CREATE TABLE entities (
    id          TEXT PRIMARY KEY,
    name        TEXT NOT NULL,
    type        TEXT NOT NULL,  -- 'person' | 'place' | 'concept' | 'project' | 'tool'
    description TEXT,
    created_at  INTEGER NOT NULL
);

-- Relationships between entities
CREATE TABLE relationships (
    id              TEXT PRIMARY KEY,
    source_entity   TEXT NOT NULL REFERENCES entities(id),
    target_entity   TEXT NOT NULL REFERENCES entities(id),
    relationship    TEXT NOT NULL,
    -- e.g. "works_on", "lives_in", "prefers_over", "is_a"
    confidence      REAL DEFAULT 0.7,
    created_at      INTEGER NOT NULL
);

-- Vector embeddings for semantic search over facts
CREATE VIRTUAL TABLE fact_vectors USING vec0(
    id TEXT PRIMARY KEY,
    embedding FLOAT[768]
);

-- Activity log (audit trail of agent decisions)
CREATE TABLE activity_log (
    id          TEXT PRIMARY KEY,
    timestamp   INTEGER NOT NULL,
    type        TEXT NOT NULL,  -- 'message_sent' | 'message_received' | 'memory_retrieval' | 'function_call' | 'llm_request' | 'llm_response' | 'consolidation' | 'error'
    details     TEXT NOT NULL,  -- JSON with type-specific data
    message_id  TEXT,           -- reference to the related message, if any
    session_id  TEXT            -- logical grouping for a single turn
);
-- e.g. { "type": "memory_retrieval", "details": { "query": "...", "results": 3, "sources": ["episode:abc", "fact:def"] } }
-- e.g. { "type": "llm_request", "details": { "prompt_tokens": 2048, "model": "llama3.1:8b", "temperature": 0.7 } }
-- e.g. { "type": "function_call", "details": { "function": "create_reminder", "args": { "when": "tomorrow 3pm", "what": "call dentist" }, "result": "ok" } }
```

**How semantic memory is built:**
- The consolidation engine extracts facts from conversations
- Facts are deduplicated and merged (confidence increases with corroboration)
- Entities are extracted and linked via relationships
- Contradictory facts are flagged for resolution

### 3.4 Memory Retrieval Strategy (Hybrid)

The system uses a **two-tier retrieval** approach:

**Tier 1 — Automatic Context Injection (proactive):**
On every user message, the system automatically:
1. Embeds the user's message
2. Queries episodic memory (top-5 most similar episodes)
3. Queries semantic memory (top-5 most relevant facts)
4. Queries the scratchpad for any active state
5. Injects all retrieved context into the prompt as a "Context" section

This happens transparently — the agent doesn't need to "decide" to retrieve. The retrieved context is always available.

**Tier 2 — Agent-Driven Retrieval (reactive):**
The agent has internal functions it can call when it needs something specific from its own memory:
- `search_memory(query string, type string)` — search episodic or semantic memory
- `get_facts(category string)` — retrieve all facts of a category
- `get_entity(name string)` — get details about an entity
- `get_recent_conversations(hours int)` — get recent conversation summaries
- `update_scratchpad(content string)` — write to working memory

This handles edge cases where automatic retrieval didn't find what's needed, or when the agent needs to explore memory proactively.

**Why this hybrid approach:**
- Automatic retrieval handles 90%+ of cases seamlessly
- Agent-driven retrieval gives the agent agency for the remaining cases
- The user never has to think about memory — it just works
- The agent can still demonstrate intelligence by choosing when to dig deeper

---

## 4. Session & Conversation Model

### 4.1 No Rigid Sessions

Instead of fixed-duration sessions, the system uses a **continuous conversation model**:

```
Timeline:
  [User messages]────[pause 2h]────[User messages]────[pause 1d]────[User messages]
       │                              │                              │
       ▼                              ▼                              ▼
  Working memory:                Working memory:                Working memory:
  - last 20 turns                 - last 20 turns                - last 20 turns
  - scratchpad                    - scratchpad                   - scratchpad
  - auto-retrieved context        - auto-retrieved context       - auto-retrieved context
                                   - summary of prev block        - summary of prev blocks
```

**How it works:**
1. The user sends a message at any time
2. The system retrieves relevant context automatically (see 3.4)
3. The agent responds using the current working memory + retrieved context
4. When the conversation pauses (no messages for ~5 minutes), a **quick consolidation** runs:
   - The recent exchange is summarized into an episodic memory entry
   - The summary is stored with vector embeddings
5. When the user returns, the automatic retrieval finds the relevant past context
6. The user experiences a seamless conversation — no session boundaries

**Why this model:**
- Feels natural — like messaging a person who remembers
- No artificial session boundaries to confuse the user
- The agent always has relevant context without the user asking
- Consolidation happens opportunistically, not on a rigid timer

### 4.2 Multi-Interface Conversations

The user can switch between interfaces freely — chat in the GUI, then send a message from Telegram, then back to the GUI. From the user's perspective, it's one continuous conversation with the same agent.

**How it works:**

1. **Shared memory, separate working contexts** — All messages from all interfaces go into the same database (same `messages`, `episodes`, `facts` tables). The `interface` column on each message records where it came from. But each interface maintains its own working memory (last ~20 turns) because the conversation flow is different on each device.

2. **Cross-interface awareness** — When the agent responds on any interface, it automatically retrieves relevant context from *all* past messages regardless of interface. So if you ask "what was that thing I asked about earlier?" on Telegram, the agent can find it even if you asked it in the GUI.

3. **No interface silos** — The agent doesn't treat interfaces as separate conversations. It's one person talking to one agent through different windows. The automatic retrieval (Tier 1) ensures relevant context from any interface is always available.

4. **Edge case — simultaneous conversations** — If the user sends a message on Telegram while the agent is mid-response on the GUI, both messages are processed in order. The agent responds to each in turn, with the second response aware of the first (since it's in the message history).

**Why this model:**
- Feels natural — like messaging someone who has both your phone number and your chat app
- No "which interface did I say that on?" confusion
- The agent's memory is unified, only the current working context is per-interface
- Consolidation (episodic summaries, fact extraction) runs on the unified message stream

### 4.3 Context Window Management

When the conversation grows beyond the model's context window:

1. **Eviction policy:** Oldest messages are evicted first
2. **Before eviction:** The evicted block is summarized into an episode
3. **Summary retention:** The episode summary stays in working memory
4. **On retrieval:** If the user asks about evicted content, the agent can retrieve the full episode from the database

This mirrors how human memory works — recent events are vivid, older events are compressed into summaries, but can be recalled in detail when triggered.

---

## 5. Consolidation Engine

### 5.1 Two-Phase Consolidation

**Phase 1 — Quick Consolidation (after ~5 min of inactivity):**
- Summarize the recent conversation block
- Store as an episode with vector embedding
- Update the scratchpad if needed
- Extract any obvious facts (e.g., "user mentioned they like Python")

**Phase 2 — Deep Consolidation (during idle time, ~30 min of inactivity):**
- Review recent episodes and extract semantic memories:
  - Extract entities (people, places, concepts mentioned)
  - Extract facts about the user (preferences, traits, habits)
  - Build relationships between entities
  - Deduplicate and merge facts (increase confidence on corroboration)
- Reorganize memory:
  - Merge related episodes
  - Update fact confidence scores
  - Flag contradictions for resolution
- Prune low-confidence, uncorroborated facts

**Why two-phase:**
- Quick consolidation is fast and happens immediately — the user can return and the summary is already stored
- Deep consolidation is expensive (requires LLM calls) and runs during idle time
- This mirrors human memory consolidation — initial stabilization happens quickly, deeper integration happens during rest

### 5.2 Concept Learning

The agent builds a semantic model of the user over time:

1. **Extraction:** From each conversation, extract entities and facts
2. **Linking:** Connect new facts to existing entities and relationships
3. **Generalization:** When multiple facts support a pattern, create a higher-level concept
   - e.g., "user prefers async" + "user uses Go" + "user builds CLI tools" → "user is a backend developer"
4. **Update:** When new information contradicts existing knowledge, the agent can:
   - Lower confidence in the old fact
   - Flag the contradiction for the user
   - Update the fact if the new information is more recent

---

## 6. Database Architecture

### 6.1 SQLite + sqlite-vec

**Why SQLite + sqlite-vec instead of a standalone vector DB:**
- Single file — no separate server process to manage
- Zero configuration — just open the database
- sqlite-vec is a pure-C extension with no dependencies
- Can do hybrid queries (structured + vector in one query)
- Perfect for a local-first application
- Portable — the database file can be backed up, copied, or moved

**Why not a standalone vector DB (Chroma, Qdrant):**
- Adds a server process — more complexity for a local app
- Overkill for a single-user application
- sqlite-vec is "fast enough" for the scale of a personal assistant

### 6.2 Database File Layout

```
~/.remy/
├── memory.db              # Main database (SQLite + vectors)
│   ├── messages           # All messages ever exchanged
│   ├── episodes           # Summarized conversation blocks
│   ├── episode_vectors    # Vector embeddings for episodes
│   ├── message_vectors    # Vector embeddings for messages
│   ├── facts              # Semantic facts about the user
│   ├── fact_vectors       # Vector embeddings for facts
│   ├── entities           # People, places, concepts
│   ├── relationships      # Links between entities
│   ├── scratchpad         # Agent's persistent working memory
│   └── activity_log       # Audit trail of agent decisions
├── config.json            # Agent configuration
└── logs/                  # Application logs
```

---

## 7. Proactive Behavior & Scheduler

The agent should be able to initiate actions without the user explicitly asking in the moment. This requires a scheduler that runs alongside the main agent loop.

### 7.1 Scheduler Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    Scheduler Engine                           │
│                                                              │
│  ┌──────────────────────┐    ┌──────────────────────────┐   │
│  │   Task Queue         │    │   Cron-style Scheduler   │   │
│  │  (one-shot reminders)│    │  (recurring tasks)       │   │
│  └──────────┬───────────┘    └─────────────┬────────────┘   │
│             │                              │                 │
│             ▼                              ▼                 │
│  ┌──────────────────────────────────────────────────────┐   │
│  │              Task Executor                            │   │
│  │  Checks every 30s for due tasks, fires them          │   │
│  └──────────────────────┬───────────────────────────────┘   │
│                         │                                    │
│                         ▼                                    │
│  ┌──────────────────────────────────────────────────────┐   │
│  │              Action Handlers                          │   │
│  │  ┌──────────┐  ┌──────────┐  ┌──────────┐           │   │
│  │  │ Send Msg │  │ Future   │           │   │
│  │  │ to User  │  │ (web, fs)│  │ Actions  │           │   │
│  │  └──────────┘  └──────────┘  └──────────┘           │   │
│  └──────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
```

### 7.2 Task Model

Tasks are stored in SQLite and checked periodically by the scheduler:

```sql
CREATE TABLE tasks (
    id          TEXT PRIMARY KEY,
    type        TEXT NOT NULL,  -- 'reminder' | 'scheduled_message'
    status      TEXT NOT NULL DEFAULT 'pending',  -- 'pending' | 'fired' | 'cancelled'
    trigger_at  INTEGER NOT NULL,  -- Unix timestamp for one-shot
    cron_expr   TEXT,              -- Cron expression for recurring (e.g., "0 8 * * *")
    action      TEXT NOT NULL,     -- JSON describing the action
    context     TEXT,              -- JSON context to inject when firing
    created_at  INTEGER NOT NULL,
    fired_at    INTEGER
);
```

**Action types:**
- `reminder` — send a message to the user at a specific time
- `scheduled_message` — send a recurring message (e.g., daily briefing)

**Example task entries:**
```json
{ "type": "reminder", "trigger_at": 1700000000, "action": {
    "type": "send_message",
    "text": "Don't forget to call the dentist at 3pm"
}}

{ "type": "scheduled_message", "cron_expr": "0 8 * * *", "action": {
    "type": "send_message",
    "text": "Good morning! Here's your daily briefing..."
}}
```

### 7.3 How the Agent Creates Tasks

The agent creates tasks during conversation:

- `create_reminder(when string, what string)` — "remind me tomorrow at 3pm to call the dentist"
- `create_schedule(cron string, action string)` — "send me a message every morning at 8am"

The LLM parses the user's natural language request and calls the appropriate function. The scheduler stores the task and fires it at the right time.

### 7.4 How Tasks Fire

When a task is due:
1. The scheduler picks it up on its periodic check (every 30s)
2. It constructs a synthetic message as if the agent is initiating the conversation
3. The message is routed through the normal agent loop:
   - Context is retrieved (relevant memories, scratchpad)
   - The agent processes the trigger and generates a response
   - The response is sent to the user via the **active interface(s)**
4. For recurring tasks, the next occurrence is calculated and stored

**Multi-interface delivery:**

When a task fires, the agent sends the message to all active interfaces. This means:

- If the GUI is open and Telegram is connected, the reminder appears in both places simultaneously
- If the GUI is closed (minimized to tray), the reminder arrives as a desktop notification and also appears in Telegram
- If only Telegram is connected (daemon mode), the reminder arrives as a Telegram message
- The user can dismiss or acknowledge the reminder on any interface — the acknowledgment syncs to the database so other interfaces know it's been handled

**Why deliver to all interfaces:**
- The user shouldn't have to guess which device will receive the reminder
- If the user is away from their Mac but has their phone, the Telegram message catches them
- If the user is at their Mac, the desktop notification is immediate and doesn't require picking up their phone
- Duplicate delivery is acceptable because the user naturally ignores the second copy once they've seen it on one device

**Acknowledgment sync:**
- When the user responds to a reminder on any interface (e.g., "thanks, done" in Telegram), the agent marks the task as acknowledged
- Subsequent deliveries to other interfaces include a note: "✓ Acknowledged on Telegram" so the user knows it's been handled
- The desktop notification is dismissed automatically when acknowledged on another interface (via polling or push)

### 7.5 Agent Awareness of Scheduled Tasks

The agent should be aware of its own schedule. The system prompt includes a summary of upcoming tasks so the agent can:
- Answer questions like "what reminders do I have?"
- Modify or cancel tasks
- Proactively mention upcoming tasks when relevant

---

## 8. Personas

A persona defines *how* the agent behaves — its identity, tone, communication style, and constraints. Personas are separate from memory (which defines *what* the agent knows). The user can create multiple personas and switch between them at any time.

Each persona can optionally specify a **provider and model override** — a different LLM provider and/or model that works best for that personality. This lets you pair personas with the models that suit them: a creative writing persona might use a large expressive model from Ollama, while a quick-task persona uses a fast lightweight one. Personas can even use different providers (e.g., default uses Ollama, but a "polished" persona uses Anthropic).

### 8.1 Persona Definition

Each persona is a Markdown file with frontmatter for model configuration and a body describing how the agent should behave. The file is loaded into the system prompt, so the LLM uses it to shape every response.

**File location:** `~/.remy/personas/<name>.md`

**Example — `~/.remy/personas/default.md`:**
```markdown
---
provider: ollama
model: llama3.1:8b
temperature: 0.7
max_tokens: 4096
---

# Remy

You are a personal assistant named Remy. You are helpful, concise, and proactive.

## Tone
- Friendly but professional
- Use casual language, not overly formal
- Use humor occasionally but don't force it
- Be direct — don't pad responses with fluff

## Behavior
- Anticipate needs and offer help before being asked
- Remember preferences and apply them automatically
- Ask clarifying questions when something is ambiguous
- Admit when you don't know something
- Never make up facts — if unsure, say so

## Constraints
- Never share personal information with anyone
- Never execute destructive commands without confirmation
- Keep responses under 200 words unless asked for detail
- Use markdown formatting for lists and code blocks
```

**Example — `~/.remy/personas/creative.md`:**
```markdown
---
provider: ollama
model: llama3.1:70b
temperature: 0.9
max_tokens: 8192
---

# Remy (Creative)

You are a creative writing partner named Remy. You are expressive, imaginative, and enthusiastic.

## Tone
- Warm and encouraging
- Use vivid language and metaphors
- Be playful and experimental
- Ask open-ended questions to spark ideas

## Behavior
- Offer multiple creative directions
- Build on the user's ideas
- Suggest alternatives and variations
- Celebrate experimentation over perfection

## Constraints
- Never share personal information
- Keep suggestions constructive, never critical
```

**Example — `~/.remy/personas/quick.md`:**
```markdown
---
provider: ollama
model: llama3.2:3b
temperature: 0.3
max_tokens: 1024
---

# Remy (Quick)

You are a fast, efficient assistant. You prioritize speed and brevity.

## Tone
- Direct and concise
- No pleasantries, no fluff
- Use bullet points where possible

## Behavior
- Answer immediately without preamble
- If unsure, say so in one sentence
- Never ask follow-up questions unless critical
```

### 8.2 How Personas Work

1. Each persona is a standalone Markdown file in `~/.remy/personas/`
2. The file has YAML frontmatter (between `---` lines) specifying model overrides, followed by the persona body
3. The active persona is loaded into the system prompt at the start of every turn
4. If the persona specifies a model override, the agent switches to that model for the duration of the conversation. If no override is specified, the default provider model is used.
5. The user can edit persona files at any time — changes take effect on the next message
6. Personas are independent of memory — switching personas doesn't affect what the agent remembers

**Model override resolution:**
- If the persona has `provider:` in frontmatter, switch to that provider (using its endpoint and API key from config)
- If the persona has `model:` in frontmatter, use that model from the selected provider
- If neither is specified, use the default provider and model from config
- Temperature and max_tokens can also be overridden per-persona
- The provider must be configured in `config.json` — if a persona references an unconfigured provider, the agent falls back to the default and logs a warning

### 8.3 Switching Personas

The user can switch personas in several ways:

- **In conversation:** "Switch to creative mode" — the agent detects this and changes the active persona (and model if overridden)
- **Via GUI:** Persona Studio tab — click any persona to activate it immediately
- **Scheduled:** A persona could be tied to time of day (e.g., professional during work hours, casual in the evening)

When switching, the agent acknowledges the change and adjusts its behavior immediately. If the model changes, the switch happens on the next message (the current response finishes with the old model).

### 8.4 Persona + Memory Interaction

Personas and memory are orthogonal:

- **Persona** = how the agent behaves (tone, style, constraints, model)
- **Memory** = what the agent knows (facts about the user, past conversations)
- The persona influences *how* the agent uses memory — a proactive persona surfaces relevant memories more readily; a formal persona presents them more structured

This means the user can switch personas without losing any of the agent's accumulated knowledge about them.

### 8.5 Persona-Aware Scratchpad

The scratchpad is shared across personas — the agent's working memory (ongoing tasks, state) persists regardless of which persona is active. This ensures continuity even when switching modes.

---

## 9. Agent Loop (Orchestration)

The core loop runs for every user message:

```
1. RECEIVE message from interface
2. STORE message in database (with embedding)
3. RETRIEVE context:
   a. Embed user message
   b. Search episodes (top-5 similar)
   c. Search facts (top-5 relevant)
   d. Load scratchpad
   e. Load recent conversation (last ~20 turns)
4. BUILD prompt:
   a. System prompt (persona + rules)
   b. Scratchpad contents
   c. Retrieved context (episodes + facts)
   d. Conversation history
   e. Current user message
5. SEND to LLM (Ollama API)
6. RECEIVE response
7. STORE response in database (with embedding)
8. SEND response to interface
9. SCHEDULE quick consolidation (if inactivity detected)
```

---

## 10. Interfaces

### 10.1 Desktop GUI (Primary Interface — Wails + Svelte)

The primary way to interact with Remy is through a native desktop window built with **Wails** (Go backend + HTML/CSS/JS frontend). The GUI is more than just a chat window — it's a control panel for the entire agent.

**Why Wails:**
- Produces a single, self-contained binary — no separate server or browser needed
- Go backend calls the agent core directly via Go method bindings
- Frontend is standard HTML/CSS/JS (Svelte recommended) — easy to style and iterate
- Cross-platform (macOS, Linux, Windows) from one codebase
- Native window chrome — feels like a real app, not a web page

**Design principles:**

The GUI should feel like a well-crafted native app — clean, warm, and unobtrusive. The agent is a conversational companion, so the interface should feel like a comfortable messaging app, not a developer tool.

- **Calm and focused** — plenty of whitespace, muted colors, nothing flashing or competing for attention. The chat is the hero; everything else is a panel or tab away.
- **Native feel** — use system fonts (San Francisco on macOS, Segoe UI on Windows), respect system accent colors, support dark mode and light mode automatically. No custom scrollbars, no heavy shadows, no "web app" tells.
- **Warm, not cold** — a subtle accent color (a soft blue or warm amber) for the agent's messages and interactive elements. Rounded corners on bubbles, gentle transitions, no harsh borders.
- **Typography-first** — the message content is the most important thing. Use a readable font size (15-16px for body text), generous line height (1.5), and proper hierarchy. Code blocks in messages use a monospace font with a subtle background.
- **Motion with purpose** — messages fade in smoothly, the typing indicator pulses gently, tab transitions are subtle. No gratuitous animations, but enough motion to feel alive.
- **Keyboard-friendly** — all common actions have keyboard shortcuts (Cmd+Enter to send, Cmd+K to search, Cmd+, for settings, Escape to close panels, Tab to switch between tabs).

**Layout:**

The window has a fixed sidebar on the left (narrow, ~48px) with icon-only tab buttons, and the main content area fills the rest. This gives maximum space to the active tab while keeping navigation always visible and one click away.

```
┌──────────────────────────────────────────────────────────────┐
│  ● ● ●                                          Remy        │
├────┬─────────────────────────────────────────────────────────┤
│    │                                                        │
│ 💬 │  (Chat — full-width message list and input)             │
│    │                                                        │
│ 🧠 │                                                        │
│    │                                                        │
│ 📋 │                                                        │
│    │                                                        │
│ 👤 │                                                        │
│    │                                                        │
│ 📜 │                                                        │
│    │                                                        │
│ ⚙  │                                                        │
│    │                                                        │
└────┴─────────────────────────────────────────────────────────┘
```

The sidebar icons are small and muted (no labels — the tooltip shows the name on hover). The active tab's icon is highlighted with the accent color. The sidebar is always visible, so switching tabs is instant.

**Tab 1 — Chat (default):**

The chat view is a two-column layout: a narrow conversation list on the left (~200px) and the active conversation on the right. This supports future multi-conversation use but keeps the single-conversation case clean.

```
┌──────────────────────────────────────────────────────────────┐
│  Conversations          │  Remy                          │
├─────────────────────────┼────────────────────────────────────┤
│                         │                                    │
│  ● Today                │  ┌──────────────────────────┐      │
│    What do you think... │  │ Hey, what do you think   │      │
│                         │  │ about using Wails for    │      │
│  Yesterday              │  │ the GUI?                 │      │
│    Can you remind me... │  └──────────────────────────┘      │
│                         │                                    │
│  Last week              │  ┌──────────────────────────┐      │
│    How does memory...   │  │ │ Wails is a great        │      │
│                         │  │ │ choice. It produces a   │      │
│                         │  │ │ single binary and...    │      │
│                         │  │ └────────────────────────┘      │
│                         │                                    │
│                         │  ┌────────────────────────────────┐ │
│                         │  │ Type a message...        [Send] │ │
│                         │  └────────────────────────────────┘ │
└─────────────────────────┴──────────────────────────────────────┘
```

- **Conversation list** — shows recent conversations with a preview of the last message. The current conversation is highlighted. For a single-user agent, there's typically just one active conversation, but the list allows creating new "topics" or revisiting old ones.
- **Message bubbles** — user messages are right-aligned with a subtle background color. Agent messages are left-aligned with a white/light background and a small avatar (a simple "R" in a circle with the accent color). Bubbles have rounded corners and subtle shadows.
- **Streaming** — as the agent generates a response, tokens appear inline with a subtle cursor animation. The user can interrupt generation by clicking a stop button that appears next to the typing indicator.
- **Markdown rendering** — inline code gets a monospace font with a subtle background. Code blocks get syntax highlighting and a copy button. Links are clickable. Lists render with proper indentation.
- **Interface indicator** — messages from Telegram show a small Telegram icon next to the timestamp. Messages from the GUI show nothing (it's the default). This is subtle — just a tiny 12x12 icon.
- **Empty state** — when there are no messages yet, show a warm welcome message: "I'm Remy, your personal assistant. Ask me anything, or tell me about your day." with a subtle illustration or icon.
- **Scroll behavior** — auto-scrolls to the bottom on new messages. If the user has scrolled up to read history, a "Jump to bottom" button appears. Smooth scrolling.

**Tab 2 — Memory Explorer:**

A clean, card-based layout for browsing the agent's knowledge.

- **Facts panel** — facts are displayed as cards in a grid, grouped by category with section headers. Each card shows the fact text, a confidence bar (visual indicator of 0.0-1.0), and a small "source" link. Hover reveals edit and delete icons. Clicking edit opens an inline text input.
- **Episode timeline** — a vertical timeline with dots and lines. Each entry shows the date, a one-line summary, and topic tags. Clicking expands the entry to show the full summary and a "View messages" button that opens the relevant messages in the Chat tab.
- **Entity graph** — a simple force-directed graph rendered with Canvas or SVG. Nodes are entities (colored by type), edges are relationships. Clicking a node highlights it and shows a side panel with details and linked facts. The graph is pannable and zoomable.
- **Search** — a prominent search bar at the top with a toggle between "semantic" and "full-text" search. Results appear below as a list with relevance scores and snippets. Pressing Enter or clicking a result navigates to the relevant detail view.
- **Scratchpad viewer** — a simple text area showing the scratchpad content, with an "Edit" toggle that makes it editable. Changes are saved automatically on blur.

**Tab 3 — Tasks & Schedule:**

A clean list-based layout.

- **Upcoming reminders** — a list with each reminder showing: time (relative + absolute), the reminder text, and action buttons (Snooze 1h, Edit, Cancel). Past reminders are shown in a separate "Fired" section with a strikethrough style.
- **Recurring schedule** — a list of recurring tasks with a toggle switch, the cron expression in human-readable form ("Every day at 8:00 AM"), next fire time, and edit/delete buttons.
- **Create reminder** — a simple inline form that appears when clicking "+ New Reminder". A date/time picker (native OS picker), a text input for the message, and a "Create" button. For recurring, a set of preset buttons ("Daily", "Weekdays", "Weekly") plus a custom cron input.
- **Empty state** — "No reminders yet. Ask Remy to remind you about something, or create one here."

**Tab 4 — Persona Studio:**

A split-panel layout: persona list on the left, editor on the right.

- **Persona list** — a vertical list of persona cards. Each card shows the name, a small tag for the provider/model override (e.g., "Ollama · llama3.1:8b"), and a brief description. The active persona has a highlighted border and a checkmark. Clicking a card selects it and loads it into the editor.
- **Persona editor** — the right panel has two sections:
  - **Top: Model configuration** — structured form fields in a horizontal row: provider dropdown, model dropdown (populated from the provider), temperature slider, max tokens input. These fields update the frontmatter automatically.
  - **Bottom: Persona body** — a clean Markdown editor with a live preview toggle. The editor has a monospace font, line numbers, and syntax highlighting for Markdown. The preview renders the persona body as it would appear in the system prompt.
- **Create new** — a button at the top of the persona list that opens a dialog: "Name your persona" + "Create from template" / "Duplicate current". The template includes default frontmatter and a starter persona body.
- **Comparison** — a button that opens a side-by-side view of two selected personas, showing both the frontmatter fields and the body text with differences highlighted.
- **Empty state** — if no personas exist, show a prompt to create the first one.

**Tab 5 — Activity Log (Audit Trail):**

A log viewer with filtering.

- **Timeline** — a scrollable list of log entries, each with a timestamp, an icon for the type (message, retrieval, function call, LLM request, consolidation, error), and a one-line description. Clicking an entry expands it to show full details (JSON payload, prompt text, etc.).
- **Filters** — a row of filter chips at the top: "All", "Messages", "Retrievals", "Functions", "LLM", "Consolidation", "Errors". Clicking a chip filters the timeline. A date range picker is also available.
- **Prompt inspector** — when clicking an LLM-related entry, a modal opens showing the full prompt that was sent to the model, with syntax-highlighted sections (system prompt, context, conversation history, user message). A "Copy" button is available.
- **Search** — a search bar that filters log entries by text content.
- **Clear** — a button to clear the log (with confirmation). The log is also automatically pruned after a configurable number of entries.

**Tab 6 — Settings:**

A standard settings layout with sections.

- **Provider management** — a list of configured providers as cards. Each card shows the provider name, endpoint (truncated), and a status indicator (green dot = connected, red = error). Clicking a card opens an edit form. An "Add Provider" button at the bottom opens a form to add a new one.
- **Default provider** — a dropdown at the top of the provider section.
- **Model parameters** — temperature and max_tokens sliders with numeric labels.
- **Interface management** — a toggle for Telegram with the bot token input (masked, with a show/hide toggle). Connection status indicator.
- **Memory settings** — database path (with a "Browse" button that opens a native file dialog), working memory turn count (number input), consolidation delays (sliders with labels like "5 min", "30 min").
- **Data management** — "Export Database" button (opens a save dialog), "Import Database" button (opens a file picker, with a warning), "Clear All Memory" button (with a confirmation dialog that requires typing "DELETE" to confirm).
- **About** — version number, build date, Go version, links to GitHub and documentation.

**System tray:**
- Remy minimizes to the system tray when the window is closed
- Tray menu: "Open Remy", "Quit"
- Desktop notifications for reminders and scheduled messages (even when the window is closed)
- Tray icon shows a subtle indicator when the agent is processing or has unread activity

**How it works:**
- Wails binds Go methods (e.g., `SendMessage`, `GetHistory`, `GetFacts`, `GetTasks`, `SwitchPersona`, `UpdatePersona`, `SaveConfig`) to the frontend
- The frontend calls these methods directly from JavaScript
- The Go backend routes messages through the same agent core as any other interface
- Streaming responses are pushed to the frontend via Wails' event system
- The GUI subscribes to backend events (new message, task fired, consolidation complete) for real-time updates

### 10.2 Telegram Interface (Optional)

- Uses the Telegram Bot API (long polling)
- Each chat with the bot is a conversation with the agent
- Supports text messages and commands
- Runs alongside the GUI or standalone via `--daemon`

### 10.3 Interface Contract

```go
type Interface interface {
    Start(ctx context.Context, msgChan chan<- Message) error
    Send(ctx context.Context, msg Message) error
    Stop() error
}
```

All interfaces implement this contract. The GUI interface is implemented via Wails bindings that call into the agent core directly. Adding a new interface (e.g., Discord, WhatsApp, Slack) means implementing these three methods.

---

## 11. Go Best Practices

These are the conventions we follow when building Remy. They ensure the codebase is idiomatic, testable, and maintainable.

### 11.1 Project Layout

```
remy/
├── cmd/
│   └── remy/
│       └── main.go              # Entry point — minimal, just wires things together
├── internal/
│   ├── agent/
│   │   ├── agent.go             # Core agent loop
│   │   └── agent_test.go
│   ├── memory/
│   │   ├── store.go             # SQLite database operations
│   │   ├── store_test.go
│   │   ├── episodic.go          # Episodic memory
│   │   ├── semantic.go          # Semantic memory
│   │   └── vectors.go           # Vector operations
│   ├── llm/
│   │   ├── client.go            # OpenAI-compatible API client
│   │   ├── client_test.go
│   │   └── provider.go          # Provider abstraction
│   ├── interface/
│   │   ├── interface.go         # Interface contract
│   │   ├── gui/
│   │   │   └── gui.go           # Wails app setup and bindings
│   │   └── telegram/
│   │       └── telegram.go
│   ├── scheduler/
│   │   ├── scheduler.go          # Task scheduling engine
│   │   └── scheduler_test.go
│   ├── persona/
│   │   └── persona.go           # Persona loading and management
│   └── config/
│       └── config.go            # Configuration loading
├── frontend/                     # Wails frontend (Svelte)
│   ├── src/
│   │   ├── App.svelte           # Main app — tab container, routing
│   │   ├── components/
│   │   │   ├── Chat.svelte           # Chat message list + input
│   │   │   ├── MessageBubble.svelte
│   │   │   ├── MemoryExplorer.svelte
│   │   │   ├── FactList.svelte
│   │   │   ├── EpisodeTimeline.svelte
│   │   │   ├── EntityGraph.svelte
│   │   │   ├── TaskManager.svelte
│   │   │   ├── PersonaStudio.svelte
│   │   │   ├── PersonaEditor.svelte
│   │   │   ├── ActivityLog.svelte
│   │   │   ├── PromptInspector.svelte
│   │   │   ├── Settings.svelte
│   │   │   └── SystemTray.svelte
│   │   ├── lib/
│   │   │   ├── wails.js         # Wails runtime bindings
│   │   │   └── stores.js        # Svelte stores for state
│   │   └── main.js
│   ├── index.html
│   ├── package.json
│   └── vite.config.js
├── wails.json                    # Wails project config
├── go.mod
├── go.sum
└── Makefile
```

**Key conventions:**
- `cmd/` contains only the main function — business logic lives in `internal/`
- `internal/` prevents external packages from importing our internals
- Each package has a single responsibility
- Tests live alongside the code they test (`store_test.go` next to `store.go`)

### 11.2 Code Style

- **Explicit over implicit** — no magic, no reflection where a simple interface suffices
- **Errors are values** — return errors, don't panic. Use `fmt.Errorf("context: %w", err)` for wrapping
- **Interfaces are small** — 1-3 methods. Define them where they're consumed, not where they're implemented
- **Zero values are useful** — design types so their zero value is ready to use where possible
- **No init() functions** — explicit initialization via constructors (`NewStore()`, `NewAgent()`)
- **Context is always first parameter** — `ctx context.Context` is always the first argument in any function that does I/O
- **Concurrency is explicit** — use goroutines and channels deliberately. Document ownership of channels (who sends, who closes)
- **No global state** — everything is passed explicitly through constructors and function parameters

### 11.3 Testing

- **Table-driven tests** — the standard Go pattern for testing multiple cases
- **Test coverage target** — aim for 80%+ on `internal/` packages. Use `go test -cover` to measure
- **Integration tests** — use build tags (`//go:build integration`) for tests that need Ollama or a real database. Separate from unit tests
- **Mock interfaces** — use Go's `io.Writer`-style pattern: define small interfaces for dependencies, implement mocks in tests
- **No test frameworks** — standard `testing` package only. No testify, no gomega, no ginkgo

**Example table-driven test:**
```go
func TestStoreSaveMessage(t *testing.T) {
    tests := []struct {
        name    string
        msg     Message
        wantErr bool
    }{
        {"valid message", Message{ID: "1", Role: "user", Content: "hello"}, false},
        {"empty content", Message{ID: "2", Role: "user", Content: ""}, true},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            s := NewStore(t.TempDir())
            err := s.SaveMessage(context.Background(), tt.msg)
            if (err != nil) != tt.wantErr {
                t.Errorf("SaveMessage() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

### 11.4 Comments

- **Comments explain why, not what** — the code already says what it does. Comments explain the reasoning, trade-offs, or non-obvious behavior
- **No comments on obvious code** — `// SaveMessage saves a message to the database` adds nothing. `// Use a write-ahead log to allow concurrent reads during writes` is useful
- **Package-level doc** — every package has a doc comment explaining its purpose
- **Exported symbols get doc comments** — `// Store handles persistence of messages, episodes, and semantic facts.`
- **No commented-out code** — delete it. Git has the history

### 11.5 Dependency Management

- **Minimal dependencies** — prefer the standard library. Every external dependency must justify its cost
- **Vendoring** — use `go mod vendor` for reproducible builds. Check in the vendor directory
- **No dependency injection frameworks** — manual DI via constructors is the Go way

### 11.6 Build & Release

- **Single binary** — `wails build` produces a native executable with the frontend embedded
- **Cross-compilation** — use `wails build -platform` for macOS (arm64/amd64), Linux (amd64/arm64), Windows (amd64)
- **Development** — `wails dev` runs the app with hot-reload for the frontend
- **Makefile** — common targets: `build`, `dev`, `test`, `lint`, `clean`, `release`
- **Versioning** — `-ldflags="-X main.Version=$(git describe --tags)"` for build info

| Component | Choice | Rationale |
|-----------|--------|-----------|
| Language | Go | Single binary, good concurrency, cross-compilation |
| GUI Framework | Wails + Svelte | Native desktop app from Go + HTML/CSS/JS, single binary |
| Database | SQLite + sqlite-vec | Single file, no server, hybrid structured+vector queries |
| LLM API | OpenAI-compatible (Ollama) | Local models, standard API |
| Embeddings | Local model via Ollama | Same API as LLM, no extra dependency |
| Telegram | Telegram Bot API | Long polling, no webhook server needed |

### 11.1 Go Dependencies

```
github.com/mattn/go-sqlite3          # SQLite driver
github.com/asg017/sqlite-vec-go      # Vector extension for SQLite
github.com/wailsapp/wails/v2         # Wails GUI framework
github.com/go-telegram/bot           # Telegram Bot API
github.com/google/uuid                # UUID generation
github.com/robfig/cron               # Cron expression parsing
```

### 11.2 Frontend Dependencies

```
svelte          # UI framework
wailsjs/runtime # Wails JavaScript runtime bindings
```

---

## 12. Configuration

```json
{
  "providers": {
    "ollama": {
      "endpoint": "http://localhost:11434/v1",
      "api_key": "",
      "chat_model": "llama3.1:8b",
      "embedding_model": "nomic-embed-text",
      "parameters": {
        "temperature": 0.7,
        "max_tokens": 4096
      }
    },
    "openai": {
      "endpoint": "https://api.openai.com/v1",
      "api_key": "${OPENAI_API_KEY}",
      "chat_model": "gpt-4o",
      "embedding_model": "text-embedding-3-small",
      "parameters": {
        "temperature": 0.7,
        "max_tokens": 4096
      }
    }
  },
  "default_provider": "ollama",
  "memory": {
    "db_path": "~/.remy/memory.db",
    "working_memory_turns": 20,
    "quick_consolidation_delay_ms": 300000,
    "deep_consolidation_delay_ms": 1800000
  },
  "persona": {
    "active": "default",
    "directory": "~/.remy/personas/"
  },
  "interfaces": {
    "telegram": {
      "enabled": false,
      "bot_token": "${REMY_TELEGRAM_BOT_TOKEN}",
      "allowed_users": []
    }
  }
}
```

---

## 13. Future Considerations

- **Multi-agent support** — the architecture supports multiple agents with separate databases
- **Encrypted memory** — the database could be encrypted at rest
- **Memory export/import** — backup and restore the database
- **Web interface** — a local web UI for reviewing memory and configuration
- **Cross-device sync** — sync the database between machines (e.g., via Syncthing)

---

## 14. Open Questions

- **Embedding model choice** — nomic-embed-text is a good default, but needs testing for the specific use case
- **Context window size** — needs tuning based on the model and typical conversation length
- **Retrieval threshold** — how many results to retrieve per query, and what similarity threshold
- **Fact confidence decay** — how quickly should uncorroborated facts lose confidence
- **Contradiction resolution** — should the agent ask the user, or use recency as a tiebreaker
