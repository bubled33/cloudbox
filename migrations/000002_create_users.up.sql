-- Создание таблицы пользователей
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    
    -- Email и display name (value objects хранятся как простые типы)
    email VARCHAR(255) UNIQUE NOT NULL,
    display_name VARCHAR(100) NOT NULL,
    is_email_verified BOOLEAN NOT NULL DEFAULT FALSE,
    
    -- Временные метки
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Индексы для часто используемых полей
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_created_at ON users(created_at);
CREATE INDEX idx_users_is_email_verified ON users(is_email_verified);

-- Комментарии для документации
COMMENT ON TABLE users IS 'Таблица пользователей системы';
COMMENT ON COLUMN users.email IS 'Email пользователя (value object Email)';
COMMENT ON COLUMN users.display_name IS 'Отображаемое имя (value object DisplayName)';
COMMENT ON COLUMN users.is_email_verified IS 'Флаг подтверждения email';
