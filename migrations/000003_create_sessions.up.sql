-- Создание таблицы сессий
CREATE TABLE IF NOT EXISTS sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    
    -- Токены и метаданные (value objects хранятся как простые типы)
    token_hash VARCHAR(255) NOT NULL UNIQUE,
    refresh_token_hash VARCHAR(255) NOT NULL UNIQUE,
    device_info TEXT NOT NULL,
    ip VARCHAR(45) NOT NULL,
    
    -- Статус сессии
    is_revoked BOOLEAN NOT NULL DEFAULT FALSE,
    
    -- Временные метки
    last_used_at TIMESTAMP NOT NULL DEFAULT NOW(),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    expired_at TIMESTAMP NOT NULL,
    
    -- Внешний ключ на пользователя
    CONSTRAINT fk_sessions_user_id 
        FOREIGN KEY (user_id) 
        REFERENCES users(id) 
        ON DELETE CASCADE
);

-- Индексы для оптимизации запросов
CREATE INDEX idx_sessions_user_id ON sessions(user_id);
CREATE INDEX idx_sessions_token_hash ON sessions(token_hash);
CREATE INDEX idx_sessions_refresh_token_hash ON sessions(refresh_token_hash);
CREATE INDEX idx_sessions_expired_at ON sessions(expired_at);
CREATE INDEX idx_sessions_is_revoked ON sessions(is_revoked);
CREATE INDEX idx_sessions_last_used_at ON sessions(last_used_at);

-- Составной индекс для поиска активных сессий пользователя
CREATE INDEX idx_sessions_user_active ON sessions(user_id, is_revoked, expired_at);

-- Комментарии для документации
COMMENT ON TABLE sessions IS 'Таблица активных сессий пользователей';
COMMENT ON COLUMN sessions.token_hash IS 'Хеш токена доступа (value object TokenHash)';
COMMENT ON COLUMN sessions.refresh_token_hash IS 'Хеш refresh токена (value object TokenHash)';
COMMENT ON COLUMN sessions.device_info IS 'Информация об устройстве (value object DeviceInfo)';
COMMENT ON COLUMN sessions.ip IS 'IP адрес (value object IP), поддерживает IPv4 и IPv6';
COMMENT ON COLUMN sessions.is_revoked IS 'Флаг отзыва сессии';
COMMENT ON COLUMN sessions.last_used_at IS 'Время последнего использования сессии';
COMMENT ON COLUMN sessions.expired_at IS 'Время истечения сессии (value object ExpiresAt)';
