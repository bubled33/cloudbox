package main

import (
	"log"
	"time"

	"github.com/yourusername/cloud-file-storage/internal/api"
	"github.com/yourusername/cloud-file-storage/internal/api/handlers"
	auth_service "github.com/yourusername/cloud-file-storage/internal/app/auth"
	magic_link_service "github.com/yourusername/cloud-file-storage/internal/app/magic_link"
	session_service "github.com/yourusername/cloud-file-storage/internal/app/session"
	"github.com/yourusername/cloud-file-storage/internal/infra/db"
	"github.com/yourusername/cloud-file-storage/internal/infra/db/repositories"
)

func main() {
	// Инициализация базы данных
	dbConn, err := db.Connect()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer dbConn.Close()

	// Инициализация репозиториев
	userRepo := repositories.NewUserRepository(dbConn)
	magicLinkRepo := repositories.NewMagicLinkRepository(dbConn)
	sessionRepo := repositories.NewSessionRepository(dbConn)

	// Инициализация сервисов
	magicLinkService := magic_link_service.NewMagicLinkService(magicLinkRepo, userRepo)
	sessionService := session_service.NewSessionService(sessionRepo)
	authService := auth_service.NewAuthService(
		magicLinkService,
		sessionService,
		24*time.Hour, // TTL для сессий - 24 часа
	)

	// Инициализация обработчиков
	authHandler := handlers.NewAuthHandler(authService)

	// Создание и запуск сервера
	server := api.NewServer(authHandler)

	err = server.Run(":8080")
	if err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}