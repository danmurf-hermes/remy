# Remy — Personal AI Assistant

[![CI](https://github.com/danmurf/remy/actions/workflows/ci.yml/badge.svg)](https://github.com/danmurf/remy/actions/workflows/ci.yml)
[![Go Version](https://img.shields.io/badge/Go-1.26+-00ADD8?logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

> **Remy** (from Latin *rememorari* — to remember) is a personal AI assistant that remembers your conversations naturally, picks up where you left off, and gets to know you over time.

Remy is not a productivity agent that runs commands or edits files. It's a **conversational companion** that lives on your machine, remembers who you are, and is always there to chat — across desktop and Telegram.

---

## ✨ Features

- **🧠 Persistent Memory** — Three-tier memory system (working, episodic, semantic) inspired by human memory. Remy remembers your conversations, learns facts about you, and builds a model of your world over time.
- **💬 Desktop GUI** — Beautiful Svelte + Wails desktop app with streaming responses, conversation management, and full memory explorer.
- **📱 Telegram Interface** — Chat with Remy from your phone via Telegram. All memory is shared across interfaces.
- **👤 Custom Personas** — Create custom personas with YAML frontmatter. Change Remy's personality, provider, or model on the fly.
- **📋 Tasks & Reminders** — "Remind me to buy milk at 5pm" — Remy handles scheduling, recurring tasks, and notifications.
- **🔍 Memory Explorer** — Browse facts, episode timeline, entity graph, and scratchpad. Search semantically or by keyword.
- **🔒 Local-First** — Everything runs on your machine. No cloud, no data leaving your computer. SQLite with built-in vector search.

---

## 🚀 Quick Start

### Prerequisites

- [Ollama](https://ollama.ai/) running locally with a chat model (e.g., `llama3.1:8b`) and an embedding model (`nomic-embed-text`)
- macOS, Linux, or Windows

### Install

```bash
# Download the latest binary for your platform from the releases page
# Or build from source:
git clone https://github.com/danmurf/remy.git
cd remy
make build
```

### Initialize

```bash
remy init
```

This checks for Ollama and required models, creates `~/.remy/` with a default config and persona, and prints next steps.

### Run

```bash
# Desktop GUI
remy

# Telegram-only mode (headless)
remy --daemon
```

---

## 📸 Screenshots

> *Screenshots coming soon.*

| Chat | Memory Explorer | Tasks |
|------|----------------|-------|
| ![Chat](docs/screenshots/chat.png) | ![Memory](docs/screenshots/memory.png) | ![Tasks](docs/screenshots/tasks.png) |

---

## 📚 Documentation

| Topic | Guide |
|-------|-------|
| Telegram Setup | [docs/telegram.md](docs/telegram.md) |
| Custom Personas | [docs/personas.md](docs/personas.md) |
| How Memory Works | [docs/memory.md](docs/memory.md) |
| Architecture | [ARCHITECTURE.md](ARCHITECTURE.md) |
| Contributing | [CONTRIBUTING.md](CONTRIBUTING.md) |

---

## 🏗️ Architecture

Remy is built with:

- **Go** backend — config, LLM client, memory store, agent loop, scheduler, consolidation engine
- **Svelte 4** frontend — desktop GUI via Wails v2
- **SQLite** + **sqlite-vec** — persistent storage with 768-dim vector search
- **Ollama** — local LLM inference via OpenAI-compatible API

See [ARCHITECTURE.md](ARCHITECTURE.md) for the full design.

---

## 🤝 Contributing

Contributions are welcome! See [CONTRIBUTING.md](CONTRIBUTING.md) for:

- Setting up the dev environment
- Running tests
- Code style guide
- How to submit PRs

---

## 📄 License

MIT — see [LICENSE](LICENSE).
