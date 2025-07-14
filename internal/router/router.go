package router

import (
	"context"
	"github.com/benlamlih/agenticx/internal/config"
	"log/slog"
	"net/http"

	scalar "github.com/MarceloPetrucio/go-scalar-api-reference"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel/trace"
)

func (s *Server) RegisterRoutes(ctx context.Context, tp trace.TracerProvider) http.Handler {
	r := gin.Default()
	r.Use(otelgin.Middleware("contract_ease", otelgin.WithTracerProvider(tp)))

	cfg := config.LoadConfig()
	frontendURL := cfg.App.FrontendURL

	slog.Info("Configuring CORS",
		"frontend_url", frontendURL,
	)

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{frontendURL},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowHeaders:     []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
	}))

	api := r.Group("/api")
	{
		api.GET("/health", s.healthHandler)
		api.GET("/docs", s.apiReferenceHandler)
	}

	slog.Info("API routes registered successfully")
	return r
}

func (s *Server) healthHandler(c *gin.Context) {
	health := s.db.Health(c)
	slog.Info("Health check performed",
		"health", health,
		"request_id", c.GetString("request_id"),
	)
	c.JSON(http.StatusOK, health)
}

func (s *Server) apiReferenceHandler(c *gin.Context) {
	htmlContent, err := scalar.ApiReferenceHTML(&scalar.Options{
		SpecURL: "./docs/build/openapi.yaml",
		CustomOptions: scalar.CustomOptions{
			PageTitle: "ContractEase API",
		},
		Theme:    "deepSpace",
		Layout:   "modern",
		DarkMode: true,
	})

	if err != nil {
		slog.Error("Failed to generate API documentation",
			"error", err,
			"request_id", c.GetString("request_id"),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate API documentation"})
		return
	}

	slog.Info("API documentation generated successfully",
		"request_id", c.GetString("request_id"),
	)

	c.Header("Content-Type", "text/html")
	c.String(http.StatusOK, htmlContent)
}
