-- Удаление индексов
DROP INDEX IF EXISTS idx_sessions_user_active;
DROP INDEX IF EXISTS idx_sessions_last_used_at;
DROP INDEX IF EXISTS idx_sessions_is_revoked;
DROP INDEX IF EXISTS idx_sessions_expired_at;
DROP INDEX IF EXISTS idx_sessions_refresh_token_hash;
DROP INDEX IF EXISTS idx_sessions_token_hash;
DROP INDEX IF EXISTS idx_sessions_user_id;

-- Удаление таблицы
DROP TABLE IF EXISTS sessions;
