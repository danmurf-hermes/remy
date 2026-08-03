# Remy — AGENTS.md

## Commands

| Action | Command |
|--------|---------|
| Build binary | `go build -ldflags="-X main.Version=$(git describe --tags 2>/dev/null || echo dev)" -o build/remy .` |
| Dev (hot-reload GUI) | `wails dev` |
| Test all Go | `go test ./internal/... -cover` |
| Test single pkg | `go test ./internal/agent/... -cover` |
| Frontend test | `cd frontend && npm test` |
| Lint all | `golangci-lint run ./...` |
| Frontend lint | `cd frontend && npx eslint --ext .js,.svelte src/` |
| Format check | `cd frontend && npx prettier --check src/` |
| Format all | `gofmt -s -w . && cd frontend && npx prettier --write src/` |
| Full CI pipeline (mirrors GitHub Actions) | `make ci` |
| Pre-commit (fmt→lint→test) | `make pre-commit` |
| Regenerate mocks | `go generate ./internal/agent/ ./internal/llm/` |

## Mandatory: run the full CI pipeline before every push

**No matter how small the change, you MUST run `make ci` before pushing.** This mirrors `.github/workflows/ci.yml` exactly (frontend deps → frontend build → Go lint → Go test → frontend lint → frontend format check → frontend test → Go build). Never push a commit or open/update a PR until `make ci` passes locally. If any step fails, fix it before pushing — do not rely on CI to catch it.

## Architecture

- **Go module**: `github.com/danmurf/remy` (Go 1.26.5)
- **Entrypoint**: `main.go` (repo root) — Wails app entrypoint
- **Internal packages**: `config/`, `llm/`, `memory/`, `agent/` — all tested
- **Frontend**: Svelte 4 + Vite + Vitest in `frontend/`
- **GUI framework**: Wails v2 (Go backend + Svelte frontend, single binary)
- **LLM**: Ollama via OpenAI-compatible API (`/v1/chat/completions`, `/v1/embeddings`)
- **Database**: SQLite + sqlite-vec (768-dim vectors, `nomic-embed-text`)
- **Config**: `~/.remy/config.json`, DB at `~/.remy/memory.db`, personas at `~/.remy/personas/*.md`

## Key conventions

- **Mocks**: generated via `//go:generate go run go.uber.org/mock/mockgen` in `agent.go` and `provider.go`. Output goes to `mock_agent/` and `mock_lll/` dirs.
- **Migrations**: embedded SQL files in `internal/memory/migrations/` via `//go:embed`. Run on a separate DB connection from the main store.
- **Vector tables**: `vec0` virtual tables require exactly 768-dim float32 embeddings. Use `memory.SerializeVector()` / `vec.SerializeFloat32()` for storage.
- **Store**: uses `sync.RWMutex` for thread safety — all public CRUD methods lock.
- **Agent**: `Store` and `Embedder` are interfaces (not concrete types) for testability. Wire `memory.NewEmbedder(...)` as the embedder.
- **CI** (`.github/workflows/ci.yml`): lint → Go test → frontend deps → frontend lint → frontend format check → frontend test → Go build. Runs on `main` and `stage-*` branches.
- **PLAN.md**: living build plan. Update stage status and add notes when completing work. Stages are worked sequentially.
- **Frontend tests**: in `frontend/src/__tests__/` with `@testing-library/svelte` + jsdom.
- **Build output**: binary goes to `build/` dir (gitignored). A prebuilt `remy` binary also lives at repo root.
- **Version**: injected via `-ldflags="-X main.Version=..."` at build time.
- **Go tests**: prefer table-driven tests using the standard `testing` package. Not a hard rule.
- **Documentation**: use standard technical English terminology throughout.
