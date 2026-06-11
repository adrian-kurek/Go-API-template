package server

import (
	"context"
	"net/http"
	"time"

	authHandler "github.com/slodkiadrianek/Go-API-template/internal/auth/handler"
	"github.com/slodkiadrianek/Go-API-template/internal/auth/routes"
)

type DependencyConfig struct {
	port        string
	authHandler authHandler.AuthHandler
}

func NewDependencyConfig(port string, authHandler authHandler.AuthHandler) *DependencyConfig {
	return &DependencyConfig{
		port:        port,
		authHandler: authHandler,
	}
}

type Server struct {
	config *DependencyConfig
	server *http.Server
	router *http.ServeMux
}

func NewServer(config *DependencyConfig) *Server {
	return &Server{
		config: config,
		router: http.NewServeMux(),
	}
}

func (s *Server) Start() error {
	s.SetupRoutes()
	s.server = &http.Server{
		Addr:         ":" + s.config.port,
		Handler:      s.router,
		ReadTimeout:  50 * time.Second,
		WriteTimeout: 50 * time.Second,
		IdleTimeout:  30 * time.Second,
	}
	return s.server.ListenAndServe()
}

func (s *Server) SetupRoutes() {
	authHandler := routes.NewAuthHandler(&s.config.authHandler)
	authHandler.SetupAuthHandlers(s.router)
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}
