-- Удаление индексов
DROP INDEX IF EXISTS idx_events_retry;
DROP INDEX IF EXISTS idx_events_locked;
DROP INDEX IF EXISTS idx_events_pending;
DROP INDEX IF EXISTS idx_events_locked_at;
DROP INDEX IF EXISTS idx_events_created_at;
DROP INDEX IF EXISTS idx_events_name;
DROP INDEX IF EXISTS idx_events_sent;

-- Удаление таблицы
DROP TABLE IF EXISTS events;
