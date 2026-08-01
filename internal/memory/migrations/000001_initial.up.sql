CREATE TABLE IF NOT EXISTS messages (
    id          TEXT PRIMARY KEY,
    user_id     TEXT NOT NULL,
    role        TEXT NOT NULL,
    content     TEXT NOT NULL,
    timestamp   INTEGER NOT NULL,
    interface   TEXT NOT NULL,
    session_id  TEXT
);

CREATE TABLE IF NOT EXISTS episodes (
    id          TEXT PRIMARY KEY,
    summary     TEXT NOT NULL,
    start_time  INTEGER NOT NULL,
    end_time    INTEGER NOT NULL,
    message_ids TEXT NOT NULL,
    importance  REAL DEFAULT 0.5,
    topics      TEXT
);

CREATE TABLE IF NOT EXISTS facts (
    id          TEXT PRIMARY KEY,
    fact        TEXT NOT NULL,
    category    TEXT NOT NULL,
    confidence  REAL DEFAULT 0.7,
    source      TEXT,
    created_at  INTEGER NOT NULL,
    updated_at  INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS entities (
    id          TEXT PRIMARY KEY,
    name        TEXT NOT NULL,
    type        TEXT NOT NULL,
    description TEXT,
    created_at  INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS relationships (
    id              TEXT PRIMARY KEY,
    source_entity   TEXT NOT NULL REFERENCES entities(id),
    target_entity   TEXT NOT NULL REFERENCES entities(id),
    relationship    TEXT NOT NULL,
    confidence      REAL DEFAULT 0.7,
    created_at      INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS scratchpad (
    id      TEXT PRIMARY KEY,
    content TEXT NOT NULL DEFAULT '',
    updated_at INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS activity_log (
    id          TEXT PRIMARY KEY,
    timestamp   INTEGER NOT NULL,
    type        TEXT NOT NULL,
    details     TEXT NOT NULL,
    message_id  TEXT,
    session_id  TEXT
);

CREATE TABLE IF NOT EXISTS tasks (
    id          TEXT PRIMARY KEY,
    type        TEXT NOT NULL,
    status      TEXT NOT NULL DEFAULT 'pending',
    trigger_at  INTEGER NOT NULL,
    cron_expr   TEXT,
    action      TEXT NOT NULL,
    context     TEXT,
    created_at  INTEGER NOT NULL,
    fired_at    INTEGER
);
