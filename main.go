package main

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/manuelgomezsw/loopi-api/internal/infrastructure/database"
	"github.com/manuelgomezsw/loopi-api/internal/interface/router"
	"github.com/manuelgomezsw/loopi-api/pkg/config"
	"github.com/manuelgomezsw/loopi-api/pkg/logger"
	"github.com/manuelgomezsw/loopi-api/pkg/observability"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load configuration", "error", err)
		panic(err)
	}

	log := logger.New(cfg.Log.Level, cfg.Log.Format)
	slog.SetDefault(log)

	// Initialize OpenTelemetry tracer → Cloud Trace (no-op when GOOGLE_CLOUD_PROJECT is unset)
	shutdownTracer, err := observability.InitTracer(
		cfg.Tracing.ProjectID,
		cfg.Tracing.ServiceName,
		cfg.Tracing.ServiceVersion,
		cfg.Tracing.Env,
	)
	if err != nil {
		log.Error("failed to initialize tracer", "error", err)
		panic(err)
	}
	defer shutdownTracer(context.Background())

	// Initialize OpenTelemetry meter → Cloud Monitoring (no-op when GOOGLE_CLOUD_PROJECT is unset)
	shutdownMeter, err := observability.InitMeter(cfg.Tracing.ProjectID)
	if err != nil {
		log.Error("failed to initialize meter", "error", err)
		panic(err)
	}
	defer shutdownMeter(context.Background())

	db, err := database.NewMySQL(cfg.Database)
	if err != nil {
		log.Error("failed to connect to database", "error", err)
		panic(err)
	}
	defer db.Close()

	log.Info("database connected")

	r := router.New(db, cfg, log)

	addr := ":" + cfg.Server.Port
	log.Info("server starting", "addr", addr)

	if err := http.ListenAndServe(addr, r); err != nil {
		log.Error("server stopped", "error", err)
		panic(err)
	}
}
