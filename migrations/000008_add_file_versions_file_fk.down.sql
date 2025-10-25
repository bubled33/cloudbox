-- Удаление внешнего ключа
ALTER TABLE file_versions
DROP CONSTRAINT IF EXISTS fk_file_versions_file_id;
