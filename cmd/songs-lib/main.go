package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"effective_mobile/internal/clients/external"
	"effective_mobile/internal/config"
	deletehandler "effective_mobile/internal/http-server/handlers/song/delete"
	filterhandler "effective_mobile/internal/http-server/handlers/song/filter"
	savehandler "effective_mobile/internal/http-server/handlers/song/save"
	texthandler "effective_mobile/internal/http-server/handlers/song/text"
	updatehandler "effective_mobile/internal/http-server/handlers/song/update"
	"effective_mobile/internal/http-server/middleware/logger"
	"effective_mobile/internal/lib/logger/sl"
	songservice "effective_mobile/internal/service/song-service"
	"effective_mobile/internal/storage/postgres"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {
	cfg := config.MustLoad()

	log := setupLogger(cfg.Env)

	log.Info(
		"starting songs-lib",
		slog.String("env", cfg.Env),
	)
	log.Debug("debug messages are enabled")

	storage, err := postgres.New(cfg.DB.Port, cfg.DB.Host, cfg.DB.User, cfg.DB.Name, cfg.DB.Password, cfg.DB.SSLMode)
	if err != nil {
		panic(err)
	}

	client := external.New(log, cfg.ExternalAPI, cfg.HTTPServer.Timeout)
	service := songservice.New(storage, storage, client)

	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(logger.New(log))
	router.Use(middleware.Recoverer)

	router.Route("/songs", func(r chi.Router) {
		r.Use(middleware.BasicAuth("songs-service", map[string]string{
			cfg.HTTPServer.User: cfg.HTTPServer.Password,
		}))

		r.Post("/", savehandler.New(log, service))
		r.Patch("/{id}", updatehandler.New(log, service))
		r.Delete("/{id}", deletehandler.New(log, storage))
	})

	router.Get("/songs", filterhandler.New(log, service, cfg.PageSizeLimit))
	router.Get("/songs/{id}", texthandler.New(log, service))

	log.Info("starting server", slog.String("address", cfg.HTTPServer.Address))

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	srv := &http.Server{
		Addr:         cfg.HTTPServer.Address,
		Handler:      router,
		ReadTimeout:  cfg.HTTPServer.Timeout,
		WriteTimeout: cfg.HTTPServer.Timeout,
		IdleTimeout:  cfg.HTTPServer.IdleTimeout,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Error("failed to start server")
		}
	}()

	log.Info("server started")

	<-done
	log.Info("stopping server")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error("failed to stop server", sl.Err(err))

		return
	}

	if err := storage.Close(); err != nil {
		log.Error("failed to close storage", sl.Err(err))

		return
	}

	log.Info("server stopped")
}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envLocal:
		log = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case envDev:
		log = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case envProd:
		log = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	default:
		log = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	}

	return log
}
