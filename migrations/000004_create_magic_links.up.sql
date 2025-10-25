-- Создание таблицы магических ссылок для passwordless аутентификации
CREATE TABLE IF NOT EXISTS magic_links (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    
    -- Токен и метаданные (value objects)
    token_hash VARCHAR(255) NOT NULL UNIQUE,
    device_info TEXT NOT NULL,
    purpose VARCHAR(50) NOT NULL,
    ip VARCHAR(45) NOT NULL,
    
    -- Статус использования
    is_used BOOLEAN NOT NULL DEFAULT FALSE,
    
    -- Временные метки
    used_at TIMESTAMP NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    expired_at TIMESTAMP NOT NULL,
    
    -- Внешний ключ на пользователя
    CONSTRAINT fk_magic_links_user_id 
        FOREIGN KEY (user_id) 
        REFERENCES users(id) 
        ON DELETE CASCADE,
    
    -- Проверка: used_at должен быть заполнен, если is_used = true
    CONSTRAINT chk_used_at_consistency 
        CHECK ((is_used = FALSE AND used_at IS NULL) OR (is_used = TRUE AND used_at IS NOT NULL))
);

-- Индексы для оптимизации запросов
CREATE INDEX idx_magic_links_user_id ON magic_links(user_id);
CREATE INDEX idx_magic_links_token_hash ON magic_links(token_hash);
CREATE INDEX idx_magic_links_expired_at ON magic_links(expired_at);
CREATE INDEX idx_magic_links_is_used ON magic_links(is_used);
CREATE INDEX idx_magic_links_purpose ON magic_links(purpose);

-- Составной индекс для поиска валидных ссылок
CREATE INDEX idx_magic_links_valid ON magic_links(token_hash, is_used, expired_at) 
    WHERE is_used = FALSE;

-- Составной индекс для cleanup задач
CREATE INDEX idx_magic_links_cleanup ON magic_links(expired_at, is_used);

-- Комментарии для документации
COMMENT ON TABLE magic_links IS 'Таблица магических ссылок для passwordless аутентификации';
COMMENT ON COLUMN magic_links.token_hash IS 'Хеш токена магической ссылки (value object TokenHash)';
COMMENT ON COLUMN magic_links.device_info IS 'Информация об устройстве (value object DeviceInfo)';
COMMENT ON COLUMN magic_links.purpose IS 'Назначение ссылки: login, email_verification и т.д. (value object Purpose)';
COMMENT ON COLUMN magic_links.ip IS 'IP адрес клиента (value object IP)';
COMMENT ON COLUMN magic_links.is_used IS 'Флаг использования ссылки (одноразовая)';
COMMENT ON COLUMN magic_links.used_at IS 'Время использования ссылки (NULL если не использована)';
COMMENT ON COLUMN magic_links.expired_at IS 'Время истечения ссылки (value object ExpiresAt)';
