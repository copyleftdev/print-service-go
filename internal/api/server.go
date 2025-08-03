package api

import (
	"net/http"

	"print-service/internal/api/handlers"
	"print-service/internal/api/middleware"
	"print-service/internal/infrastructure/logger"
	"print-service/internal/pkg/config"

	"github.com/gin-gonic/gin"
)

// Server represents the HTTP server
type Server struct {
	config *config.Config
	logger logger.Logger
	router *gin.Engine
}

// NewServer creates a new HTTP server
func NewServer(cfg *config.Config, logger logger.Logger) *Server {
	// Set gin mode based on environment
	if cfg.Logger.Level == "debug" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	server := &Server{
		config: cfg,
		logger: logger.With("component", "server"),
		router: router,
	}

	server.setupMiddleware()
	server.setupRoutes()

	return server
}

// Handler returns the HTTP handler
func (s *Server) Handler() http.Handler {
	return s.router
}

// setupMiddleware sets up middleware
func (s *Server) setupMiddleware() {
	// Recovery middleware
	s.router.Use(gin.Recovery())

	// CORS middleware
	s.router.Use(middleware.CORS())

	// Logging middleware
	s.router.Use(middleware.Logging(s.logger))

	// Rate limiting middleware
	s.router.Use(middleware.RateLimit())

	// Authentication middleware (for protected routes)
	// s.router.Use(middleware.Auth()) // Uncomment when auth is implemented
}

// setupRoutes sets up API routes
func (s *Server) setupRoutes() {
	// Health check routes
	healthHandler := handlers.NewHealthHandler(s.logger)
	s.router.GET("/health", healthHandler.Health)
	s.router.GET("/ready", healthHandler.Ready)

	// Metrics routes
	metricsHandler := handlers.NewMetricsHandler(s.logger)
	s.router.GET("/metrics", metricsHandler.Metrics)

	// API v1 routes
	v1 := s.router.Group("/api/v1")
	{
		// Print routes
		printHandler := handlers.NewPrintHandler(s.config, s.logger)
		v1.POST("/print", printHandler.Print)
		v1.GET("/print/:id", printHandler.GetStatus)
		v1.DELETE("/print/:id", printHandler.Cancel)
		v1.GET("/print/:id/download", printHandler.Download)

		// Job management routes
		v1.GET("/jobs", printHandler.ListJobs)
		v1.GET("/jobs/:id", printHandler.GetJob)
	}
}
