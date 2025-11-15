package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/yourusername/cloud-file-storage/docs"
	"github.com/yourusername/cloud-file-storage/internal/api"
	auth_handlers "github.com/yourusername/cloud-file-storage/internal/api/handlers/auth"
	files_handler "github.com/yourusername/cloud-file-storage/internal/api/handlers/files"
	metrics_handler "github.com/yourusername/cloud-file-storage/internal/api/handlers/metrics"
	users_handler "github.com/yourusername/cloud-file-storage/internal/api/handlers/users"
	"github.com/yourusername/cloud-file-storage/internal/app"
	auth_service "github.com/yourusername/cloud-file-storage/internal/app/auth"
	event_service "github.com/yourusername/cloud-file-storage/internal/app/event"
	file_service "github.com/yourusername/cloud-file-storage/internal/app/file"
	file_version_service "github.com/yourusername/cloud-file-storage/internal/app/file_version"
	magic_link_service "github.com/yourusername/cloud-file-storage/internal/app/magic_link"
	public_link_service "github.com/yourusername/cloud-file-storage/internal/app/public_link"
	session_service "github.com/yourusername/cloud-file-storage/internal/app/session"
	user_service "github.com/yourusername/cloud-file-storage/internal/app/user"
	"github.com/yourusername/cloud-file-storage/internal/config"
	"github.com/yourusername/cloud-file-storage/internal/infra/db"
	"github.com/yourusername/cloud-file-storage/internal/infra/queue"
	"github.com/yourusername/cloud-file-storage/internal/infra/smtp"
	"github.com/yourusername/cloud-file-storage/internal/infra/storage"
	"github.com/yourusername/cloud-file-storage/internal/workers"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

func initTracer() func() {
	exporter, err := stdouttrace.New(
		stdouttrace.WithPrettyPrint(),
	)
	if err != nil {
		log.Fatal(err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
	)
	otel.SetTracerProvider(tp)

	return func() {
		_ = tp.Shutdown(context.Background())
	}
}

// @title Cloud file storage
// @version 1.0
// @BasePath /api/v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	shutdown := initTracer()
	defer shutdown()

	if err := godotenv.Load(".env"); err != nil {
		log.Printf("Warning: .env file not found: %v", err)
	}
	httpAddr := ""
	if v := flag.Lookup("http.addr"); v != nil {
		httpAddr = v.Value.String()
	}

	cfg, err := config.Load("configs/config.base.yaml", "configs/config.dev.yaml", httpAddr)
	if err != nil {
		log.Fatal(err)
	}

	dbConn, err := db.Connect()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer dbConn.Close()

	writer, reader := queue.NewMockQueue()
	eventWriter, eventReader := queue.NewMockQueue()
	eventProducer := queue.NewKafkaEventProducer(eventWriter)
	eventConsuer := queue.NewKafkaEventConsumer(eventReader)
	previewConsumer := queue.NewKafkaPreviewConsumer(reader)
	previewProducer := queue.NewKafkaPreviewProducer(writer)

	eventQueryRepository := db.NewEventQueryRepository(dbConn)
	magicLinkQueryRepo := db.NewMagicLinkQueryRepository(dbConn)
	sessionQueryRepo := db.NewSessionQueryRepository(dbConn)
	userQueryRepo := db.NewUserQueryRepository(dbConn)
	fileVersionQueryRepo := db.NewFileVersionQueryRepository(dbConn)
	fileQueryRepo := db.NewFileQueryRepository(dbConn)
	publicLinkQueryRepository := db.NewPublicLinkQueryRepository(dbConn)

	eventCommandRepository := db.NewEventCommandRepository()
	magicLinkCommandRepo := db.NewMagicLinkCommandRepository()
	sessionCommandRepo := db.NewSessionCommandRepository()
	userCommandRepo := db.NewUserCommandRepository()
	fileVersionCommandRepo := db.NewFileVersionCommandRepository()
	fileCommandRepo := db.NewFileCommandRepository()
	publicLinkCommandRepository := db.NewPublicLinkCommandRepository()

	uow := app.NewUnitOfWork(dbConn)

	s3, err := storage.NewS3Storage(
		cfg.Immutable.S3.Endpoint,
		cfg.Immutable.S3.Bucket,
		cfg.Immutable.S3.AccessKeyID,
		cfg.Immutable.S3.SecretAccessKey,
		cfg.Immutable.S3.Region,
	)
	if err != nil {
		log.Fatalf("S3 init failed: %v", err)
	}

	mailSender := smtp.NewSMTPMailSender(
		cfg.Immutable.SMTP.Host,
		cfg.Immutable.SMTP.Port,
		cfg.Immutable.SMTP.Name,
		cfg.Immutable.SMTP.Email,
		cfg.Immutable.SMTP.Password,
	)
	fmt.Println("Key ", cfg.Immutable.S3.SecretAccessKey)
	eventService := event_service.NewEventService(eventQueryRepository, eventCommandRepository, eventProducer, "1", *uow)
	magicLinkService := magic_link_service.NewMagicLinkService(magicLinkQueryRepo, magicLinkCommandRepo, eventService, *uow)
	sessionService := session_service.NewSessionService(sessionQueryRepo, sessionCommandRepo, eventService, *uow)
	versionService := file_version_service.NewFileVersionService(fileQueryRepo, fileCommandRepo, fileVersionQueryRepo, fileVersionCommandRepo, s3, previewConsumer, previewProducer, eventService, *uow)
	fileService := file_service.NewFileService(fileQueryRepo, fileCommandRepo, fileVersionQueryRepo, fileVersionCommandRepo, eventService, *uow)
	publicLinkService := public_link_service.NewPublicLinkService(publicLinkQueryRepository, publicLinkCommandRepository, eventService, *uow)
	userService := user_service.NewUserService(userQueryRepo, userCommandRepo, eventService, *uow, fileService, sessionService)
	authService := auth_service.NewAuthService(
		magicLinkService,
		sessionService,
		userService,
		24*time.Hour,
		mailSender,
	)

	previewWorker := workers.NewPreviewWorker(s3, previewConsumer, versionService)
	fileChecker := workers.NewFileChecker(versionService, *uow, s3, time.Second*50)
	metricWorker := workers.NewMetricsWorker(eventConsuer, time.Second*5)
	publishWorker := workers.NewPublishEventsWorker(eventService, time.Second*5, 5, 3)

	authHandler := auth_handlers.NewAuthHandler(authService, sessionService)
	userHandler := users_handler.NewUserHandler(userService)
	fileHandler := files_handler.NewFileHandler(versionService, fileService, publicLinkService)
	metricHandler := metrics_handler.NewMetricsHandler()
	server := api.NewServer(authHandler, userHandler, fileHandler, metricHandler, authService)

	go previewWorker.Handle(context.Background())
	go fileChecker.Start(context.Background())
	go publishWorker.Start(context.Background())
	go metricWorker.Start(context.Background())

	if err := server.Run(cfg.Immutable.HTTP.Addr); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
