-- Создание ENUM типа для статуса версии файла
CREATE TYPE file_status AS ENUM ('uploaded', 'processing', 'ready', 'failed');

-- Создание таблицы версий файлов
CREATE TABLE IF NOT EXISTS file_versions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    file_id UUID NOT NULL,
    uploaded_by_session_id UUID NOT NULL,
    
    -- Хранение и метаданные файла (value objects)
    s3_key VARCHAR(1000) NOT NULL,
    mime VARCHAR(100) NOT NULL,
    preview_s3_key VARCHAR(1000) NULL,
    
    -- Статус обработки
    status file_status NOT NULL DEFAULT 'uploaded',
    
    -- Характеристики файла
    size BIGINT NOT NULL CHECK (size >= 0),
    version_num INT NOT NULL CHECK (version_num > 0),
    
    -- Временные метки
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),

    
    CONSTRAINT fk_file_versions_uploaded_by_session_id 
        FOREIGN KEY (uploaded_by_session_id) 
        REFERENCES sessions(id) 
        ON DELETE SET NULL,
    
    -- Уникальность версии для каждого файла
    CONSTRAINT uq_file_versions_file_version 
        UNIQUE (file_id, version_num)
);

-- Индексы для оптимизации запросов
CREATE INDEX idx_file_versions_file_id ON file_versions(file_id);
CREATE INDEX idx_file_versions_uploaded_by_session_id ON file_versions(uploaded_by_session_id);
CREATE INDEX idx_file_versions_s3_key ON file_versions(s3_key);
CREATE INDEX idx_file_versions_status ON file_versions(status);
CREATE INDEX idx_file_versions_created_at ON file_versions(created_at);

-- Составной индекс для получения последней версии файла
CREATE INDEX idx_file_versions_file_latest ON file_versions(file_id, version_num DESC);

-- Partial индекс для версий в обработке
CREATE INDEX idx_file_versions_processing ON file_versions(status, created_at) 
    WHERE status IN ('uploaded', 'processing');

-- Комментарии для документации
COMMENT ON TABLE file_versions IS 'Таблица версий файлов (поддержка версионирования)';
COMMENT ON COLUMN file_versions.file_id IS 'ID родительского файла';
COMMENT ON COLUMN file_versions.uploaded_by_session_id IS 'ID сессии, загрузившей эту версию';
COMMENT ON COLUMN file_versions.s3_key IS 'Ключ файла в S3 (value object S3Key)';
COMMENT ON COLUMN file_versions.mime IS 'MIME тип файла (value object MimeType)';
COMMENT ON COLUMN file_versions.preview_s3_key IS 'Ключ превью в S3 (nullable, value object S3Key)';
COMMENT ON COLUMN file_versions.status IS 'Статус обработки: uploaded, processing, ready, failed';
COMMENT ON COLUMN file_versions.size IS 'Размер файла в байтах (value object FileSize)';
COMMENT ON COLUMN file_versions.version_num IS 'Номер версии файла (value object FileVersionNum)';

-- Триггер для автоматического обновления updated_at
CREATE TRIGGER update_file_versions_updated_at 
    BEFORE UPDATE ON file_versions
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
