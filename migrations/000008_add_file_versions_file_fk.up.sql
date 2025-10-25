-- Добавление внешнего ключа от file_versions к files
ALTER TABLE file_versions
ADD CONSTRAINT fk_file_versions_file_id 
    FOREIGN KEY (file_id) 
    REFERENCES files(id) 
    ON DELETE CASCADE;
