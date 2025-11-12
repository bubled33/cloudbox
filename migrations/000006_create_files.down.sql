-- Удаление триггера
DROP TRIGGER IF EXISTS update_files_updated_at ON files;

-- Удаление индексов
DROP INDEX IF EXISTS idx_files_name_trgm;
DROP INDEX IF EXISTS idx_files_processing;
DROP INDEX IF EXISTS idx_files_owner_created;
DROP INDEX IF EXISTS idx_files_mime;
DROP INDEX IF EXISTS idx_files_created_at;
DROP INDEX IF EXISTS idx_files_name;
DROP INDEX IF EXISTS idx_files_status;
DROP INDEX IF EXISTS idx_files_uploaded_by_session_id;
DROP INDEX IF EXISTS idx_files_owner_id;

-- Удаление таблицы
DROP TABLE IF EXISTS files;
