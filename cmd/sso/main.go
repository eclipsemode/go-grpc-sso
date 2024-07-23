package main

import (
	"github.com/eclipsemode/go-grpc-sso/internal/app"
	"github.com/eclipsemode/go-grpc-sso/internal/config"
	"github.com/eclipsemode/go-grpc-sso/internal/lib/logger/handlers/slogpretty"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {
	// инициализировать обьект кофига
	cfg := config.MustLoad()

	// инициализировать логгер

	log := setupLogger(cfg.Env)

	log.Info("starting server", slog.Any("cfg", cfg))

	// инициализировать приложение (app)
	application := app.New(log, cfg.GRPC.Port, cfg.StoragePath, cfg.TokenTTL)

	// запустить gRPC-сервер приложения
	go application.GRPCSrv.MustRun()

	//Graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	sign := <-stop

	err := application.GRPCSrv.Stop()
	if err != nil {
		log.Error("stopping server failed", slog.Any("err", err))
		os.Exit(1)
	}

	log.Info("app stopped", slog.String("signal", sign.String()))
}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envLocal:
		log = setupPrettySlog()
	case envDev:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envProd:
		log = slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}

	return log
}

func setupPrettySlog() *slog.Logger {
	opts := slogpretty.PrettyHandlerOptions{
		SlogOpts: &slog.HandlerOptions{
			Level: slog.LevelDebug,
		},
	}

	handler := opts.NewPrettyHandler(os.Stdout)

	return slog.New(handler)
}
