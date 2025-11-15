package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginswagger "github.com/swaggo/gin-swagger"

	auth_handlers "github.com/yourusername/cloud-file-storage/internal/api/handlers/auth"
	files_handler "github.com/yourusername/cloud-file-storage/internal/api/handlers/files"
	metrics_handler "github.com/yourusername/cloud-file-storage/internal/api/handlers/metrics"
	users_handler "github.com/yourusername/cloud-file-storage/internal/api/handlers/users"
	"github.com/yourusername/cloud-file-storage/internal/api/middleware"
	auth_service "github.com/yourusername/cloud-file-storage/internal/app/auth"
)

type Server struct {
	router         *gin.Engine
	authHandler    *auth_handlers.AuthHandler
	userHandler    *users_handler.UserHandler
	fileHandler    *files_handler.FileHandler
	metricsHandler *metrics_handler.MetricsHandler
	authSrv        *auth_service.AuthService
}

func NewServer(
	authHandler *auth_handlers.AuthHandler,
	userHandler *users_handler.UserHandler,
	fileHandler *files_handler.FileHandler,
	metricsHandler *metrics_handler.MetricsHandler,
	authSrv *auth_service.AuthService,
) *Server {
	router := gin.Default()

	s := &Server{
		router:         router,
		authHandler:    authHandler,
		userHandler:    userHandler,
		fileHandler:    fileHandler,
		metricsHandler: metricsHandler,
		authSrv:        authSrv,
	}
	s.setupRoutes()
	return s
}

func (s *Server) setupRoutes() {
	// Добавляем эндпоинт метрик без middleware и аутентификации
	s.router.GET("/api/v1/metrics", s.metricsHandler.ServeMetrics)

	s.router.GET("/swagger/*any", ginswagger.WrapHandler(swaggerfiles.Handler))

	s.router.GET("/api/v1/public-links/:token", s.fileHandler.DownloadByPublicLink)
	s.router.POST("/api/v1/public-links/:file_id", s.fileHandler.CreatePublicLink)

	v1 := s.router.Group("/api/v1")
	v1.Use(middleware.TracingMiddleware("cloud file storage"))
	v1.Use(middleware.ErrorMiddleware())

	{
		auth := v1.Group("/auth")

		v1.POST("/magic-links", s.authHandler.RequestMagicLink)
		v1.GET("/magic-links/:token", s.authHandler.VerifyMagicLink)

		v1.POST("/auth/tokens/refresh", s.authHandler.RefreshToken)

		authProtected := auth.Group("")
		authProtected.Use(middleware.AuthMiddleware(s.authSrv))

		{
			authProtected.DELETE("/sessions/current", s.authHandler.Logout)
			authProtected.DELETE("/sessions", s.authHandler.LogoutAll)
			authProtected.GET("/sessions", s.authHandler.GetActiveSessions)
			authProtected.DELETE("/sessions/:session_id", s.authHandler.RevokeSession)
		}

		users := v1.Group("/users")
		users.Use(middleware.AuthMiddleware(s.authSrv))

		{
			users.GET("/me", s.userHandler.GetMe)
			users.PATCH("/me", s.userHandler.UpdateProfile)
			users.DELETE("/me", s.userHandler.DeleteAccount)
		}

		files := v1.Group("/files")
		files.Use(middleware.AuthMiddleware(s.authSrv))

		{
			files.POST("", s.fileHandler.UploadNewFile)
			files.GET("", s.fileHandler.ListFiles)
			files.GET("/:file_id", s.fileHandler.GetFile)
			files.POST("/:file_id/versions", s.fileHandler.UploadNewVersion)
			files.GET("/:file_id/versions", s.fileHandler.GetFileVersions)
			files.GET("/:file_id/versions/:version_num/content", s.fileHandler.GetVersionDownloadURL)
			files.PATCH("/:file_id", s.fileHandler.UpdateFile)
			files.DELETE("/:file_id", s.fileHandler.DeleteFile)
			files.POST("/:file_id/versions/:version_num/restore", s.fileHandler.RestoreFileVersion)
			files.POST("/:file_id/public-links", s.fileHandler.CreatePublicLink)
			files.GET("/:file_id/public-links", s.fileHandler.GetPublicLinks)
			files.DELETE("/:file_id/public-links/:link_id", s.fileHandler.DeletePublicLink)
		}
	}
}

func (s *Server) Run(addr string) error {
	return s.router.Run(addr)
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}
