package main

import (
	"log"
	"time"

	"github.com/yourusername/cloud-file-storage/internal/api"
	"github.com/yourusername/cloud-file-storage/internal/api/handlers"
	"github.com/yourusername/cloud-file-storage/internal/app"
	auth_service "github.com/yourusername/cloud-file-storage/internal/app/auth"
	event_service "github.com/yourusername/cloud-file-storage/internal/app/event"
	magic_link_service "github.com/yourusername/cloud-file-storage/internal/app/magic_link"
	session_service "github.com/yourusername/cloud-file-storage/internal/app/session"
	user_service "github.com/yourusername/cloud-file-storage/internal/app/user"
	"github.com/yourusername/cloud-file-storage/internal/infra/db"
	"github.com/yourusername/cloud-file-storage/internal/infra/queue"
)

func main() {
	// Инициализация базы данных
	dbConn, err := db.Connect()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer dbConn.Close()
	queue := queue.NewKafkaEventProducer(&queue.MockWriter{})
	// Инициализация репозиториев
	eventQueryRepository := db.NewEventQueryRepository(dbConn)
	magicLinkQueryRepo := db.NewMagicLinkQueryRepository(dbConn)
	sessionQueryRepo := db.NewSessionQueryRepository(dbConn)
	userQueryRepo := db.NewUserQueryRepository(dbConn)

	eventCommandRepository := db.NewEventCommandRepository()
	magicLinkCommandRepo := db.NewMagicLinkCommandRepository()
	sessionCommandRepo := db.NewSessionCommandRepository()
	userCommandRepo := db.NewUserCommandRepository()

	uow := app.NewUnitOfWork(dbConn)

	// Инициализация сервисов
	eventService := event_service.NewEventService(eventQueryRepository, eventCommandRepository, queue, "1")
	magicLinkService := magic_link_service.NewMagicLinkService(magicLinkQueryRepo, magicLinkCommandRepo, eventService, *uow)
	sessionService := session_service.NewSessionService(sessionQueryRepo, sessionCommandRepo, eventService, *uow)

	userService := user_service.NewUserService(userQueryRepo, userCommandRepo, eventService, *uow)
	authService := auth_service.NewAuthService(
		magicLinkService,
		sessionService,
		userService,
		24*time.Hour, // TTL для сессий - 24 часа
	)

	// Инициализация обработчиков
	authHandler := handlers.NewAuthHandler(authService)
	userHandler := handlers.NewUserHandler(userService)
	// Создание и запуск сервера
	server := api.NewServer(authHandler, userHandler)

	err = server.Run(":8080")
	if err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
