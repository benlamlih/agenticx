package router

import (
	"context"
	"fmt"
	"github.com/benlamlih/agenticx/internal/config"
	database "github.com/benlamlih/agenticx/internal/repository"
	"net/http"
	"strconv"
	"time"

	"go.opentelemetry.io/otel/trace"
)

type Server struct {
	port int
	db   database.Service
}

func NewServer(db database.Service, port int) *Server {
	return &Server{
		db:   db,
		port: port,
	}
}

func BuildHTTPServer(ctx context.Context, tp trace.TracerProvider) *http.Server {
	cfg := config.LoadConfig()
	port, _ := strconv.Atoi(cfg.App.Port)
	db := database.New(ctx, cfg)

	app := NewServer(db, port)

	return &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      app.RegisterRoutes(ctx, tp),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}
}
