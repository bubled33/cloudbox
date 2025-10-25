-- Создание таблицы событий для event sourcing / outbox pattern
CREATE TABLE IF NOT EXISTS events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    
    -- Данные события
    name VARCHAR(255) NOT NULL,
    data TEXT NOT NULL,
    
    -- Статус отправки
    sent BOOLEAN NOT NULL DEFAULT FALSE,
    
    -- Блокировка для обработки (pessimistic locking)
    locked_at TIMESTAMP NULL,
    locked_by VARCHAR(255) NULL,
    retry_count INT NOT NULL DEFAULT 0,
    
    -- Временные метки
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Индексы для оптимизации запросов
CREATE INDEX idx_events_sent ON events(sent);
CREATE INDEX idx_events_name ON events(name);
CREATE INDEX idx_events_created_at ON events(created_at);
CREATE INDEX idx_events_locked_at ON events(locked_at);

-- Составной индекс для поиска необработанных событий
CREATE INDEX idx_events_pending ON events(sent, created_at);

-- Составной индекс для поиска заблокированных событий
CREATE INDEX idx_events_locked ON events(locked_at, locked_by) 
    WHERE locked_at IS NOT NULL;

-- Индекс для поиска событий для повторной обработки
CREATE INDEX idx_events_retry ON events(sent, retry_count, created_at) 
    WHERE sent = FALSE;

-- Комментарии для документации
COMMENT ON TABLE events IS 'Таблица событий для Outbox Pattern (гарантированная доставка)';
COMMENT ON COLUMN events.name IS 'Имя/тип события (например: file.uploaded, user.created)';
COMMENT ON COLUMN events.data IS 'JSON payload события';
COMMENT ON COLUMN events.sent IS 'Флаг успешной отправки события';
COMMENT ON COLUMN events.locked_at IS 'Время блокировки события для обработки';
COMMENT ON COLUMN events.locked_by IS 'ID инстанса worker-а, обрабатывающего событие';
COMMENT ON COLUMN events.retry_count IS 'Количество попыток обработки';
