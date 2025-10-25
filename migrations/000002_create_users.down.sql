-- Удаление индексов
DROP INDEX IF EXISTS idx_users_is_email_verified;
DROP INDEX IF EXISTS idx_users_created_at;
DROP INDEX IF EXISTS idx_users_email;

-- Удаление таблицы
DROP TABLE IF EXISTS users;
