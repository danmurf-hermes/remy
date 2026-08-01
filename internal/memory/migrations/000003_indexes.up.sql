-- Add indexes for common query patterns

-- Messages: queries by timestamp (recent messages, history)
CREATE INDEX IF NOT EXISTS idx_messages_timestamp ON messages(timestamp DESC);

-- Messages: queries by session
CREATE INDEX IF NOT EXISTS idx_messages_session_id ON messages(session_id);

-- Messages: queries by user
CREATE INDEX IF NOT EXISTS idx_messages_user_id ON messages(user_id);

-- Episodes: queries by time range
CREATE INDEX IF NOT EXISTS idx_episodes_start_time ON episodes(start_time DESC);
CREATE INDEX IF NOT EXISTS idx_episodes_end_time ON episodes(end_time DESC);

-- Episodes: queries by importance
CREATE INDEX IF NOT EXISTS idx_episodes_importance ON episodes(importance DESC);

-- Facts: queries by category
CREATE INDEX IF NOT EXISTS idx_facts_category ON facts(category);

-- Facts: queries by confidence
CREATE INDEX IF NOT EXISTS idx_facts_confidence ON facts(confidence DESC);

-- Facts: queries by updated_at (for consolidation)
CREATE INDEX IF NOT EXISTS idx_facts_updated_at ON facts(updated_at DESC);

-- Activity log: queries by timestamp
CREATE INDEX IF NOT EXISTS idx_activity_log_timestamp ON activity_log(timestamp DESC);

-- Activity log: queries by type
CREATE INDEX IF NOT EXISTS idx_activity_log_type ON activity_log(type);

-- Activity log: queries by session
CREATE INDEX IF NOT EXISTS idx_activity_log_session_id ON activity_log(session_id);

-- Tasks: queries by status
CREATE INDEX IF NOT EXISTS idx_tasks_status ON tasks(status);

-- Tasks: queries by trigger_at (for scheduler)
CREATE INDEX IF NOT EXISTS idx_tasks_trigger_at ON tasks(trigger_at);

-- Tasks: queries by type
CREATE INDEX IF NOT EXISTS idx_tasks_type ON tasks(type);

-- Entities: queries by type
CREATE INDEX IF NOT EXISTS idx_entities_type ON entities(type);

-- Entities: queries by name
CREATE INDEX IF NOT EXISTS idx_entities_name ON entities(name);

-- Relationships: queries by source entity
CREATE INDEX IF NOT EXISTS idx_relationships_source ON relationships(source_entity);

-- Relationships: queries by target entity
CREATE INDEX IF NOT EXISTS idx_relationships_target ON relationships(target_entity);
