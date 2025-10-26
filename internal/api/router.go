package api

import (
	"github.com/gin-gonic/gin"
	"github.com/yourusername/cloud-file-storage/internal/api/handlers"
)

type Server struct {
	router      *gin.Engine
	authHandler *handlers.AuthHandler
}

func NewServer(authHandler *handlers.AuthHandler) *Server {
	router := gin.Default()

	s := &Server{
		router:      router,
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
	}
}

func (s *Server) Run(addr string) error {
	return s.router.Run(addr)
}
