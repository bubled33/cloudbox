package db

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/lib/pq" // Добавьте эту строку!
)

// Connect устанавливает соединение с базой данных PostgreSQL
func Connect() (*sql.DB, error) {
	// Получаем параметры подключения из переменных окружения
	host := getEnv("DB_HOST", "localhost")
	port := getEnv("DB_PORT", "5432")
	user := getEnv("DB_USER", "cloudbox")
	password := getEnv("DB_PASSWORD", "password")
	dbname := getEnv("DB_NAME", "cloudbox")
	sslmode := getEnv("DB_SSLMODE", "disable")

	// Формируем строку подключения
	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbname, sslmode)

	// Подключаемся к базе данных
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// Проверяем соединение
	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Настройка пула соединений
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)

	return db, nil
}

// getEnv возвращает значение переменной окружения или значение по умолчанию
func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
