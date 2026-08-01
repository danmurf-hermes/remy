# How Memory Works

Remy's memory system is inspired by human memory — it has three tiers that work together to create a natural, persistent conversation experience.

---

## Overview

```
┌─────────────────────────────────────────────────────────────┐
│                    Working Memory                            │
│  (Recent conversation history — last N turns)                │
├─────────────────────────────────────────────────────────────┤
│                    Episodic Memory                           │
│  (Summarized conversation blocks — "what happened")         │
├─────────────────────────────────────────────────────────────┤
│                    Semantic Memory                           │
│  (Facts, entities, relationships — "what is true")         │
└─────────────────────────────────────────────────────────────┘
```

---

## 1. Working Memory

**What it is:** The most recent conversation turns, kept in full detail.

**How it works:**
- Every message you send and every response Remy generates is stored
- The agent includes the last N turns (configurable, default 20) in the prompt
- This gives Remy immediate context for the current conversation

**Analogy:** Like your short-term memory — what was just said.

---

## 2. Episodic Memory

**What it is:** Summarized blocks of past conversation, stored as episodes with vector embeddings for semantic retrieval.

**How it works:**
- After **5 minutes of inactivity**, Remy performs a **quick consolidation**:
  1. Takes the recent conversation block
  2. Summarizes it into an episode (summary, topics, importance score)
  3. Generates a vector embedding of the summary
  4. Stores it in the episodic memory table
- When you send a new message, Remy:
  1. Generates an embedding of your message
  2. Searches episodic memory for the most relevant past episodes
  3. Includes the top matches in the prompt context

**Analogy:** Like your memory of yesterday's conversation — you don't remember every word, but you remember the gist.

### Episode Structure

Each episode contains:
- **Summary**: A concise summary of what was discussed
- **Topics**: Key topics mentioned (e.g., "project planning", "vacation plans")
- **Importance**: A score from 0.1 to 1.0 based on conversation length and content
- **Message IDs**: References to the original messages
- **Timestamp**: When the conversation occurred
- **Embedding**: 768-dimensional vector for semantic search

---

## 3. Semantic Memory

**What it is:** Facts, entities, and relationships extracted from conversations — the persistent knowledge Remy has built about you and your world.

**How it works:**
- After **30 minutes of inactivity**, Remy performs a **deep consolidation**:
  1. Reviews recent episodes
  2. Extracts facts (e.g., "User works at Acme Corp")
  3. Extracts entities (e.g., "Acme Corp" — a company)
  4. Extracts relationships (e.g., "User works_at Acme Corp")
  5. Deduplicates facts — if a fact already exists, its confidence score increases
  6. Generates vector embeddings for semantic search
- When you send a message, relevant facts are retrieved alongside episodes

**Analogy:** Like what you know about a close friend — their job, their hobbies, their family — built up over many conversations.

### Fact Structure

Each fact contains:
- **Text**: The factual statement (e.g., "User prefers dark mode")
- **Category**: A category like `preference`, `work`, `personal`, `health`, `technology`
- **Confidence**: A score from 0.1 to 1.0, increasing with corroboration
- **Source episode**: Which conversation this fact came from
- **Timestamp**: When the fact was learned
- **Embedding**: For semantic search

### Entity & Relationship Structure

**Entities** are things the user talks about:
- People, places, organizations, projects, concepts
- Each has a name, type, and description

**Relationships** connect entities:
- "User works_at Acme Corp"
- "User lives_in Portland"
- "Project X depends_on Library Y"

---

## 4. Consolidation Engine

The consolidation engine runs in the background and manages the flow of information from working → episodic → semantic memory.

### Timeline

```
Message sent → SignalActivity()
     │
     ▼
[Activity timer resets]
     │
     ├── After 5 min inactivity → Quick Consolidation
     │     └── Summarize recent messages → Store as episode with embedding
     │
     └── After 30 min inactivity → Deep Consolidation
           └── Review episodes → Extract facts, entities, relationships
                 → Deduplicate facts → Update confidence scores
```

### Quick Consolidation

- Triggers after 5 minutes of inactivity (configurable)
- Summarizes the conversation block since the last consolidation
- Stores as an episode with a vector embedding
- Episodes are searchable via semantic similarity

### Deep Consolidation

- Triggers after 30 minutes of inactivity (configurable)
- Reviews episodes created since the last deep consolidation
- Asks the LLM to extract facts, entities, and relationships
- Deduplicates facts: if a fact already exists, confidence increases by 0.1 (capped at 1.0)
- Stores new entities and relationships

---

## 5. The Scratchpad

The scratchpad is a persistent note that Remy maintains across conversations. It's included in every prompt, giving Remy a place to keep track of ongoing context.

**What it's used for:**
- Current conversation topic
- Pending questions or follow-ups
- Temporary notes that don't warrant a permanent fact

You can view and edit the scratchpad in the **Memory Explorer** tab of the desktop app.

---

## 6. Memory Explorer (GUI)

The desktop app includes a **Memory Explorer** tab where you can:

- **Facts**: Browse all learned facts grouped by category. Edit or delete facts. See confidence bars.
- **Episodes**: View the episode timeline with importance indicators. Expand entries for full details.
- **Entities**: See the entity graph with relationships visualized as a network.
- **Scratchpad**: View and edit the current scratchpad with auto-save.
- **Search**: Search memory semantically (by meaning) or by keyword (full-text).

---

## Configuration

Memory behavior can be tuned in `~/.remy/config.json`:

```json
{
  "memory": {
    "db_path": "~/.remy/memory.db",
    "working_memory_turns": 20,
    "quick_consolidation_delay_ms": 300000,
    "deep_consolidation_delay_ms": 1800000
  }
}
```

| Setting | Default | Description |
|---------|---------|-------------|
| `db_path` | `~/.remy/memory.db` | Path to the SQLite database file |
| `working_memory_turns` | 20 | Number of recent conversation turns to include in the prompt |
| `quick_consolidation_delay_ms` | 300000 (5 min) | Inactivity time before quick consolidation |
| `deep_consolidation_delay_ms` | 1800000 (30 min) | Inactivity time before deep consolidation |

---

## Technical Details

### Database

- **Engine**: SQLite with sqlite-vec extension
- **Vector dimensions**: 768 (matching `nomic-embed-text`)
- **Search**: Cosine similarity via vec0 virtual tables
- **Migrations**: Embedded SQL files in `internal/memory/migrations/`

### Embedding Model

Remy uses `nomic-embed-text` via Ollama for generating embeddings. This model produces 768-dimensional vectors optimized for semantic similarity search.

### Vector Search

When you send a message:
1. Your message is embedded using `nomic-embed-text`
2. The embedding is used to search both episodic and semantic memory
3. Results are ranked by cosine similarity
4. Top matches are included in the prompt context
