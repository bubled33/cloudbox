-- Создание таблицы файлов (основная информация о файле)
CREATE TABLE IF NOT EXISTS files (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id UUID NOT NULL,
    uploaded_by_session_id UUID NOT NULL,
    
    -- Метаданные файла (value objects)
    name VARCHAR(500) NOT NULL,
    mime VARCHAR(100) NOT NULL,
    preview_s3_key VARCHAR(1000) NULL,
    
    -- Статус файла (используется ENUM из file_versions)
    status file_status NOT NULL DEFAULT 'uploaded',
    
    -- Характеристики текущей версии
    size BIGINT NOT NULL CHECK (size >= 0),
    version_num INT NOT NULL DEFAULT 1 CHECK (version_num > 0),
    
    -- Временные метки
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    
    -- Внешние ключи
    CONSTRAINT fk_files_owner_id 
        FOREIGN KEY (owner_id) 
        REFERENCES users(id) 
        ON DELETE CASCADE,
    
    CONSTRAINT fk_files_uploaded_by_session_id 
        FOREIGN KEY (uploaded_by_session_id) 
        REFERENCES sessions(id) 
        ON DELETE SET NULL
);

-- Индексы для оптимизации запросов
CREATE INDEX idx_files_owner_id ON files(owner_id);
CREATE INDEX idx_files_uploaded_by_session_id ON files(uploaded_by_session_id);
CREATE INDEX idx_files_status ON files(status);
CREATE INDEX idx_files_name ON files(name);
CREATE INDEX idx_files_created_at ON files(created_at);
CREATE INDEX idx_files_mime ON files(mime);

-- Составной индекс для поиска файлов пользователя
CREATE INDEX idx_files_owner_created ON files(owner_id, created_at DESC);

-- Partial индекс для файлов в обработке
CREATE INDEX idx_files_processing ON files(status, created_at);

-- Full-text search индекс для поиска по имени (опционально)
CREATE INDEX idx_files_name_trgm ON files USING gin(name gin_trgm_ops);

-- Комментарии для документации
COMMENT ON TABLE files IS 'Таблица файлов (основная информация о текущем состоянии файла)';
COMMENT ON COLUMN files.owner_id IS 'ID владельца файла';
COMMENT ON COLUMN files.uploaded_by_session_id IS 'ID сессии, загрузившей файл';
COMMENT ON COLUMN files.name IS 'Имя файла (value object FileName)';
COMMENT ON COLUMN files.mime IS 'MIME тип файла (value object MimeType)';
COMMENT ON COLUMN files.preview_s3_key IS 'Ключ превью в S3 (nullable, value object S3Key)';
COMMENT ON COLUMN files.status IS 'Текущий статус файла';
COMMENT ON COLUMN files.size IS 'Размер текущей версии в байтах (value object FileSize)';
COMMENT ON COLUMN files.version_num IS 'Номер текущей версии файла (value object FileVersionNum)';

-- Триггер для автоматического обновления updated_at
CREATE TRIGGER update_files_updated_at 
    BEFORE UPDATE ON files
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
