package memory

import (
	"database/sql"
	"embed"
	"fmt"
	"sync"

	sq "github.com/Masterminds/squirrel"
	vec "github.com/asg017/sqlite-vec-go-bindings/cgo"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/mattn/go-sqlite3" // SQLite driver registration
)

var sb = sq.StatementBuilder.PlaceholderFormat(sq.Question)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// Store wraps a SQLite database and provides thread-safe CRUD operations
// for messages, episodes, facts, entities, relationships, and the scratchpad.
type Store struct {
	db *sql.DB
	mu sync.RWMutex
}

// NewStore opens or creates a SQLite database at the given path, runs
// migrations, and returns a ready-to-use Store.
func NewStore(dbPath string) (*Store, error) {
	vec.Auto()

	db, err := sql.Open("sqlite3", dbPath+"?_journal_mode=WAL&_busy_timeout=5000")
	if err != nil {
		return nil, fmt.Errorf("opening database: %w", err)
	}

	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("pinging database: %w", err)
	}

	s := &Store{db: db}
	if err := s.migrate(dbPath); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("running migrations: %w", err)
	}

	return s, nil
}

// Close closes the underlying database connection.
func (s *Store) Close() error {
	return s.db.Close()
}

// DB returns the underlying *sql.DB for advanced use cases.
func (s *Store) DB() *sql.DB {
	return s.db
}

func (s *Store) migrate(dbPath string) error {
	src, err := iofs.New(migrationsFS, "migrations")
	if err != nil {
		return fmt.Errorf("creating migration source: %w", err)
	}
	defer func() { _ = src.Close() }()

	migrateDB, err := sql.Open("sqlite3", dbPath+"?_journal_mode=WAL&_busy_timeout=5000")
	if err != nil {
		return fmt.Errorf("opening migration database: %w", err)
	}
	defer func() { _ = migrateDB.Close() }()

	driver, err := sqlite3.WithInstance(migrateDB, &sqlite3.Config{})
	if err != nil {
		return fmt.Errorf("creating migration driver: %w", err)
	}

	m, err := migrate.NewWithInstance("iofs", src, "sqlite3", driver)
	if err != nil {
		return fmt.Errorf("creating migrator: %w", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("applying migrations: %w", err)
	}

	if _, dbErr := m.Close(); dbErr != nil {
		return fmt.Errorf("closing migrator: %w", dbErr)
	}

	return nil
}
