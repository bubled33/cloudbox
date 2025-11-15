package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http/httptest"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

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
	"github.com/yourusername/cloud-file-storage/internal/infra/db"
	"github.com/yourusername/cloud-file-storage/internal/infra/queue"
	"github.com/yourusername/cloud-file-storage/internal/infra/smtp"
	"github.com/yourusername/cloud-file-storage/internal/infra/storage"
	"github.com/yourusername/cloud-file-storage/internal/test"
	"github.com/yourusername/cloud-file-storage/internal/workers"
)

type TestEnv struct {
	Server         *api.Server
	DB             *test.TestDatabase
	Kafka          *test.TestKafka
	S3             *test.TestS3Storage
	AuthService    *auth_service.AuthService
	UserService    *user_service.UserService
	FileService    *file_service.FileService
	VersionService *file_version_service.FileVersionService
	MailSender     *smtp.MockMailSender

	// Репозитории
	FileCommandRepo        *db.FileCommandRepository
	FileVersionCommandRepo *db.FileVersionCommandRepository
	UOW                    *app.UnitOfWork

	// Управление воркерами
	WorkerCtx     context.Context
	CancelWorkers context.CancelFunc
}

func SetupTestEnvironment(t *testing.T) (*TestEnv, func()) {
	ctx := context.Background()

	testDB, err := test.SetupTestDatabase(ctx)
	require.NoError(t, err)
	_, filename, _, _ := runtime.Caller(0)
	projectRoot := filepath.Join(filepath.Dir(filename), "..", "..", "..", "..")
	migrationsPath, _ := filepath.Abs(filepath.Join(projectRoot, "migrations"))
	err = testDB.RunMigrations(migrationsPath)
	require.NoError(t, err)

	testKafka, err := test.SetupTestKafka(ctx)
	require.NoError(t, err)

	err = testKafka.CreateTopic(ctx, "events")
	require.NoError(t, err)

	err = testKafka.CreateTopic(ctx, "previews")
	require.NoError(t, err)

	testS3, err := test.SetupTestS3(ctx)
	require.NoError(t, err)

	eventQueryRepository := db.NewEventQueryRepository(testDB.DB)
	magicLinkQueryRepo := db.NewMagicLinkQueryRepository(testDB.DB)
	sessionQueryRepo := db.NewSessionQueryRepository(testDB.DB)
	userQueryRepo := db.NewUserQueryRepository(testDB.DB)
	fileVersionQueryRepo := db.NewFileVersionQueryRepository(testDB.DB)
	fileQueryRepo := db.NewFileQueryRepository(testDB.DB)
	publicLinkQueryRepository := db.NewPublicLinkQueryRepository(testDB.DB)

	eventCommandRepository := db.NewEventCommandRepository()
	magicLinkCommandRepo := db.NewMagicLinkCommandRepository()
	sessionCommandRepo := db.NewSessionCommandRepository()
	userCommandRepo := db.NewUserCommandRepository()
	fileVersionCommandRepo := db.NewFileVersionCommandRepository()
	fileCommandRepo := db.NewFileCommandRepository()
	publicLinkCommandRepository := db.NewPublicLinkCommandRepository()

	uow := app.NewUnitOfWork(testDB.DB)

	eventWriter := testKafka.NewWriter("events")
	eventReader := testKafka.NewReader("events", "test-events-group")
	previewWriter := testKafka.NewWriter("previews")
	previewReader := testKafka.NewReader("previews", "test-preview-group")

	eventProducer := queue.NewKafkaEventProducer(eventWriter)
	eventConsumer := queue.NewKafkaEventConsumer(eventReader)
	previewConsumer := queue.NewKafkaPreviewConsumer(previewReader)
	previewProducer := queue.NewKafkaPreviewProducer(previewWriter)

	endpoint, err := testS3.GetEndpoint(ctx)
	require.NoError(t, err)

	accessKey, secretKey := testS3.GetCredentials()

	s3Storage, err := storage.NewS3Storage(
		endpoint,
		testS3.GetBucket(),
		accessKey,
		secretKey,
		"us-east-1",
	)
	require.NoError(t, err)

	mailSender := smtp.NewMockMailSender()

	eventService := event_service.NewEventService(
		eventQueryRepository,
		eventCommandRepository,
		eventProducer,
		"test-instance",
		*uow,
	)

	magicLinkService := magic_link_service.NewMagicLinkService(
		magicLinkQueryRepo,
		magicLinkCommandRepo,
		eventService,
		*uow,
	)

	sessionService := session_service.NewSessionService(
		sessionQueryRepo,
		sessionCommandRepo,
		eventService,
		*uow,
	)

	versionService := file_version_service.NewFileVersionService(
		fileQueryRepo,
		fileCommandRepo,
		fileVersionQueryRepo,
		fileVersionCommandRepo,
		s3Storage,
		previewConsumer,
		previewProducer,
		eventService,
		*uow,
	)

	fileService := file_service.NewFileService(
		fileQueryRepo,
		fileCommandRepo,
		fileVersionQueryRepo,
		fileVersionCommandRepo,
		eventService,
		*uow,
	)

	publicLinkService := public_link_service.NewPublicLinkService(
		publicLinkQueryRepository,
		publicLinkCommandRepository,
		eventService,
		*uow,
	)

	userService := user_service.NewUserService(
		userQueryRepo,
		userCommandRepo,
		eventService,
		*uow,
		fileService,
		sessionService,
	)

	authService := auth_service.NewAuthService(
		magicLinkService,
		sessionService,
		userService,
		24*time.Hour,
		mailSender,
	)

	authHandler := auth_handlers.NewAuthHandler(authService, sessionService)
	userHandler := users_handler.NewUserHandler(userService)
	fileHandler := files_handler.NewFileHandler(versionService, fileService, publicLinkService)
	metricHandler := metrics_handler.NewMetricsHandler()

	server := api.NewServer(authHandler, userHandler, fileHandler, metricHandler, authService)

	// Создаем контекст для управления воркерами
	workerCtx, cancelWorkers := context.WithCancel(ctx)

	// Создаем и запускаем воркеры
	previewWorker := workers.NewPreviewWorker(s3Storage, previewConsumer, versionService)
	fileChecker := workers.NewFileChecker(versionService, *uow, s3Storage, time.Second*1)
	metricWorker := workers.NewMetricsWorker(eventConsumer, time.Second*1)
	publishWorker := workers.NewPublishEventsWorker(eventService, time.Second*1, 5, 3)

	// Запускаем воркеры в фоне
	go previewWorker.Handle(workerCtx)
	go fileChecker.Start(workerCtx)
	go publishWorker.Start(workerCtx)
	go metricWorker.Start(workerCtx)

	// Даем воркерам время на инициализацию
	time.Sleep(100 * time.Millisecond)

	env := &TestEnv{
		Server:                 server,
		DB:                     testDB,
		Kafka:                  testKafka,
		S3:                     testS3,
		AuthService:            authService,
		UserService:            userService,
		FileService:            fileService,
		VersionService:         versionService,
		MailSender:             mailSender,
		FileCommandRepo:        fileCommandRepo,
		FileVersionCommandRepo: fileVersionCommandRepo,
		UOW:                    uow,
		WorkerCtx:              workerCtx,
		CancelWorkers:          cancelWorkers,
	}

	cleanup := func() {
		// Останавливаем воркеры
		cancelWorkers()

		// Даем время воркерам на graceful shutdown
		time.Sleep(200 * time.Millisecond)

		// Закрываем ресурсы
		testDB.Terminate(ctx)
		testKafka.Terminate(ctx)
		testS3.Terminate(ctx)
	}

	return env, cleanup
}

func (env *TestEnv) CleanupBetweenTests(t *testing.T) {
	ctx := context.Background()

	err := env.DB.CleanDB(ctx)
	require.NoError(t, err)

	err = env.S3.CleanBucket(ctx)
	require.NoError(t, err)
}

func (env *TestEnv) NewRequest(t *testing.T, method, path string, body io.Reader) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, body)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	w := httptest.NewRecorder()
	env.Server.ServeHTTP(w, req)

	return w
}

func (env *TestEnv) NewRequestWithAuth(t *testing.T, method, path string, body io.Reader, token string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, body)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	env.Server.ServeHTTP(w, req)

	return w
}

func (env *TestEnv) NewJSONRequest(t *testing.T, method, path string, body interface{}) *httptest.ResponseRecorder {
	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		require.NoError(t, err)
		bodyReader = bytes.NewReader(jsonBody)
	}

	return env.NewRequest(t, method, path, bodyReader)
}

func (env *TestEnv) NewJSONRequestWithAuth(t *testing.T, method, path string, body interface{}, token string) *httptest.ResponseRecorder {
	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		require.NoError(t, err)
		bodyReader = bytes.NewReader(jsonBody)
	}

	return env.NewRequestWithAuth(t, method, path, bodyReader, token)
}

func ParseJSONResponse(t *testing.T, w *httptest.ResponseRecorder) map[string]interface{} {
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	return response
}
