-- Удаление индексов
DROP INDEX IF EXISTS idx_public_links_valid;
DROP INDEX IF EXISTS idx_public_links_file_active;
DROP INDEX IF EXISTS idx_public_links_expired_at;
DROP INDEX IF EXISTS idx_public_links_token_hash;
DROP INDEX IF EXISTS idx_public_links_created_by_user_id;
DROP INDEX IF EXISTS idx_public_links_file_id;

-- Удаление таблицы
DROP TABLE IF EXISTS public_links;
