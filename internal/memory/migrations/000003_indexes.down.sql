-- Remove indexes added in migration 000003

DROP INDEX IF EXISTS idx_messages_timestamp;
DROP INDEX IF EXISTS idx_messages_session_id;
DROP INDEX IF EXISTS idx_messages_user_id;
DROP INDEX IF EXISTS idx_episodes_start_time;
DROP INDEX IF EXISTS idx_episodes_end_time;
DROP INDEX IF EXISTS idx_episodes_importance;
DROP INDEX IF EXISTS idx_facts_category;
DROP INDEX IF EXISTS idx_facts_confidence;
DROP INDEX IF EXISTS idx_facts_updated_at;
DROP INDEX IF EXISTS idx_activity_log_timestamp;
DROP INDEX IF EXISTS idx_activity_log_type;
DROP INDEX IF EXISTS idx_activity_log_session_id;
DROP INDEX IF EXISTS idx_tasks_status;
DROP INDEX IF EXISTS idx_tasks_trigger_at;
DROP INDEX IF EXISTS idx_tasks_type;
DROP INDEX IF EXISTS idx_entities_type;
DROP INDEX IF EXISTS idx_entities_name;
DROP INDEX IF EXISTS idx_relationships_source;
DROP INDEX IF EXISTS idx_relationships_target;
