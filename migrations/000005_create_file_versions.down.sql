-- Удаление триггера
DROP TRIGGER IF EXISTS update_file_versions_updated_at ON file_versions;

-- Удаление индексов
DROP INDEX IF EXISTS idx_file_versions_processing;
DROP INDEX IF EXISTS idx_file_versions_file_latest;
DROP INDEX IF EXISTS idx_file_versions_created_at;
DROP INDEX IF EXISTS idx_file_versions_status;
DROP INDEX IF EXISTS idx_file_versions_s3_key;
DROP INDEX IF EXISTS idx_file_versions_uploaded_by_session_id;
DROP INDEX IF EXISTS idx_file_versions_file_id;

-- Удаление таблицы
DROP TABLE IF EXISTS file_versions;

-- Удаление ENUM типа
DROP TYPE IF EXISTS file_status;
