CREATE VIRTUAL TABLE IF NOT EXISTS episode_vectors USING vec0(
    id TEXT PRIMARY KEY,
    embedding FLOAT[768]
);

CREATE VIRTUAL TABLE IF NOT EXISTS message_vectors USING vec0(
    id TEXT PRIMARY KEY,
    embedding FLOAT[768]
);

CREATE VIRTUAL TABLE IF NOT EXISTS fact_vectors USING vec0(
    id TEXT PRIMARY KEY,
    embedding FLOAT[768]
);
