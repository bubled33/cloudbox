package api

import (
	"github.com/gin-gonic/gin"
	"github.com/yourusername/cloud-file-storage/internal/api/handlers"
	"github.com/yourusername/cloud-file-storage/internal/api/middleware"
)

type Server struct {
	router      *gin.Engine
	authHandler *handlers.AuthHandler
	userHandler *handlers.UserHandler
}

func NewServer(authHandler *handlers.AuthHandler, userHandler *handlers.UserHandler) *Server {
	router := gin.Default()

	s := &Server{
		router:      router,
		userHandler: userHandler,
		authHandler: authHandler,
	}
	s.setupRoutes()
	return s
}

func (s *Server) setupRoutes() {
	v1 := s.router.Group("/api/v1")

	{
		auth := v1.Group("/auth")
		auth.POST("/request-magic-link", s.authHandler.RequestMagicLink)
		auth.GET("/verify-magic-link", s.authHandler.VerifyMagicLink) // Новый эндпоинт

		users := v1.Group("/users")
		users.Use(middleware.AuthMiddleware(s.authHandler.Auth_srv))
		{
			users.GET("/me", s.userHandler.GetMe)
		}

	}
}

func (s *Server) Run(addr string) error {
	return s.router.Run(addr)
}
