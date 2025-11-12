-- Создание таблицы публичных ссылок для шаринга файлов
CREATE TABLE IF NOT EXISTS public_links (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    file_id UUID NOT NULL,
    created_by_user_id UUID NOT NULL,
    
    -- Токен для доступа
    token_hash VARCHAR(255) NOT NULL UNIQUE,
    
    -- Временные метки
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    expired_at TIMESTAMP NOT NULL,
    
    -- Внешние ключи
    CONSTRAINT fk_public_links_file_id 
        FOREIGN KEY (file_id) 
        REFERENCES files(id) 
        ON DELETE CASCADE,
    
    CONSTRAINT fk_public_links_created_by_user_id 
        FOREIGN KEY (created_by_user_id) 
        REFERENCES users(id) 
        ON DELETE CASCADE
);

-- Индексы для оптимизации запросов
CREATE INDEX idx_public_links_file_id ON public_links(file_id);
CREATE INDEX idx_public_links_created_by_user_id ON public_links(created_by_user_id);
CREATE INDEX idx_public_links_token_hash ON public_links(token_hash);
CREATE INDEX idx_public_links_expired_at ON public_links(expired_at);

-- Составной индекс для поиска активных ссылок файла
CREATE INDEX idx_public_links_file_active ON public_links(file_id, expired_at);

-- Partial index для валидных (не истёкших) ссылок
CREATE INDEX idx_public_links_valid ON public_links(token_hash, expired_at);

-- Комментарии для документации
COMMENT ON TABLE public_links IS 'Таблица публичных ссылок для шаринга файлов';
COMMENT ON COLUMN public_links.file_id IS 'ID файла, на который создана ссылка';
COMMENT ON COLUMN public_links.created_by_user_id IS 'ID пользователя, создавшего ссылку';
COMMENT ON COLUMN public_links.token_hash IS 'Хеш токена для доступа по публичной ссылке';
COMMENT ON COLUMN public_links.expired_at IS 'Время истечения ссылки';
