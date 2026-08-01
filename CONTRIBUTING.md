# Contributing to Remy

Thank you for your interest in contributing to Remy! This guide will help you get started.

---

## Table of Contents

- [Development Environment](#development-environment)
- [Project Structure](#project-structure)
- [Running Tests](#running-tests)
- [Code Style](#code-style)
- [Building and Releasing](#building-and-releasing)
- [How to Submit PRs](#how-to-submit-prs)
- [Branch Naming](#branch-naming)

---

## Development Environment

### Prerequisites

| Tool | Version | Purpose |
|------|---------|---------|
| [Go](https://go.dev/dl/) | 1.26+ | Backend |
| [Node.js](https://nodejs.org/) | 20+ | Frontend build |
| [npm](https://www.npmjs.com/) | 10+ | Frontend dependencies |
| [Ollama](https://ollama.ai/) | Latest | Local LLM (for integration tests) |
| [golangci-lint](https://golangci-lint.run/) | v2.12+ | Go linting |
| [Wails](https://wails.io/) | v2.13.0+ | Desktop GUI framework |

### Quick Setup

```bash
# Clone the repo
git clone https://github.com/danmurf/remy.git
cd remy

# Install frontend dependencies
cd frontend && npm install && cd ..

# Verify everything works
make test
```

### Required Ollama Models

```bash
ollama pull llama3.1:8b    # Chat model (or your preferred model)
ollama pull nomic-embed-text  # Embedding model (required for memory)
```

---

## Project Structure

```
remy/
├── main.go                  # Entry point (Wails app)
├── internal/
│   ├── agent/               # Core agent loop, consolidation, prompt building
│   ├── app/                 # Wails app bindings (Go ↔ Svelte bridge)
│   ├── config/              # Config loading/saving (~/.remy/config.json)
│   ├── interface/
│   │   └── telegram/        # Telegram bot interface
│   ├── llm/                 # LLM provider abstraction + Ollama client
│   ├── memory/              # SQLite store, vectors, migrations
│   ├── persona/             # Persona loading/parsing (YAML frontmatter)
│   └── scheduler/           # Task scheduling and reminders
├── frontend/
│   ├── src/
│   │   ├── App.svelte       # Root component with tab navigation
│   │   ├── lib/
│   │   │   ├── Chat.svelte  # Chat interface
│   │   │   ├── MemoryExplorer.svelte
│   │   │   ├── TaskManager.svelte
│   │   │   ├── PersonaStudio.svelte
│   │   │   ├── ActivityLog.svelte
│   │   │   ├── Settings.svelte
│   │   │   ├── stores.js    # Svelte stores
│   │   │   └── wails.js     # Wails runtime wrappers
│   │   └── __tests__/       # Frontend unit tests
│   └── package.json
├── .github/workflows/ci.yml # CI pipeline
├── Makefile                 # Build targets
├── ARCHITECTURE.md          # Full architecture documentation
└── PLAN.md                  # Living build plan
```

---

## Running Tests

### Go Tests

```bash
# All Go tests
go test ./internal/... -cover

# Single package
go test ./internal/agent/... -cover

# With race detection
go test -race ./internal/...
```

### Frontend Tests

```bash
cd frontend
npm test
```

### Full Test Suite

```bash
make test
```

This runs both Go and frontend tests.

---

## Code Style

### Go

- Follow standard Go conventions (`gofmt`, `go vet`, `golangci-lint`)
- Use table-driven tests with the standard `testing` package
- Mock interfaces for external dependencies (LLM, database, Telegram)
- Prefer `slog` for structured logging
- Wrap errors with context using `fmt.Errorf("...: %w", err)`
- Pass large structs by pointer, not by value

Run linting:

```bash
make lint-go
```

### Frontend (Svelte/JavaScript)

- Use ES module syntax (`import`/`export`)
- Curly braces required on all control flow statements (`if`, `for`, etc.)
- ARIA labels and keyboard handlers on interactive elements
- Prefer `const` over `let` where possible
- Component files use `.svelte` extension

Run linting and formatting:

```bash
make lint-frontend
make fmt
```

### Pre-Commit Check

Before committing, run the full pre-commit suite:

```bash
make pre-commit
```

This runs formatting, linting, and all tests.

---

## Building and Releasing

### Development Build

```bash
make build
```

Produces `build/remy` with version injected from git tags.

### Hot-Reload Development

```bash
wails dev
```

Starts the Wails dev server with live frontend reloading.

### Release Build

Releases are automated via GitHub Actions. To trigger a release:

```bash
# Tag the commit
git tag v1.0.0
git push origin v1.0.0
```

The release workflow builds binaries for macOS (arm64, amd64), Linux (amd64, arm64), and Windows (amd64), runs the full test suite, and creates a GitHub release with release notes.

---

## How to Submit PRs

1. **Fork the repo** on GitHub
2. **Create a branch** following the naming convention (see below)
3. **Make your changes** — keep commits focused and well-described
4. **Run the full test suite** — `make pre-commit`
5. **Push to your fork** and open a PR to `danmurf/remy` `main`
6. **Wait for CI** to pass — if it fails, fix and re-push
7. **Request review** if needed

### PR Checklist

- [ ] All tests pass locally (`make test`)
- [ ] Linting passes (`make lint`)
- [ ] Formatting is correct (`make fmt`)
- [ ] New code has tests
- [ ] Documentation is updated (if applicable)
- [ ] Commits are signed

### Commit Messages

Use [Conventional Commits](https://www.conventionalcommits.org/):

```
type(scope): short description

Longer explanation if needed. Wrap at 72 characters.
```

Types: `feat`, `fix`, `refactor`, `docs`, `test`, `ci`, `chore`, `perf`

Examples:
```
feat(agent): add streaming response support
fix(memory): handle empty embedding results gracefully
docs: add Telegram setup guide
```

---

## Branch Naming

| Pattern | Purpose |
|---------|---------|
| `stage-*` | Build plan stages (e.g., `stage-15-docs`) |
| `feat/*` | New features |
| `fix/*` | Bug fixes |
| `docs/*` | Documentation |
| `refactor/*` | Code restructuring |
| `ci/*` | CI/CD changes |

---

## Additional Resources

- [ARCHITECTURE.md](ARCHITECTURE.md) — Full system design
- [PLAN.md](PLAN.md) — Build plan and stage tracking
- [AGENTS.md](AGENTS.md) — Quick reference for AI agents working on the codebase
