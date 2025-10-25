-- Удаление индексов
DROP INDEX IF EXISTS idx_magic_links_cleanup;
DROP INDEX IF EXISTS idx_magic_links_valid;
DROP INDEX IF EXISTS idx_magic_links_purpose;
DROP INDEX IF EXISTS idx_magic_links_is_used;
DROP INDEX IF EXISTS idx_magic_links_expired_at;
DROP INDEX IF EXISTS idx_magic_links_token_hash;
DROP INDEX IF EXISTS idx_magic_links_user_id;

-- Удаление таблицы
DROP TABLE IF EXISTS magic_links;
