package main

import (
	"log/slog"
	"net/http"

	"github.com/manuelgomezsw/loopi-api/internal/infrastructure/database"
	"github.com/manuelgomezsw/loopi-api/internal/interface/router"
	"github.com/manuelgomezsw/loopi-api/pkg/config"
	"github.com/manuelgomezsw/loopi-api/pkg/logger"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load configuration", "error", err)
		panic(err)
	}

	log := logger.New(cfg.Log.Level, cfg.Log.Format)
	slog.SetDefault(log)

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
