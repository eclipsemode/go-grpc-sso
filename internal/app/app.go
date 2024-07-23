package app

import (
	grpcapp "github.com/eclipsemode/go-grpc-sso/internal/app/grpc"
	"github.com/eclipsemode/go-grpc-sso/internal/services/auth"
	"github.com/eclipsemode/go-grpc-sso/internal/storage/sqlite"
	"log/slog"
	"time"
)

type App struct {
	GRPCSrv *grpcapp.App
}

func New(
	log *slog.Logger,
	grpcPort int,
	storagePath string,
	tokenTTL time.Duration,
) *App {
	// инициализировать хранилище (storage)
	storage, err := sqlite.NewStorage(storagePath)
	if err != nil {
		panic(err)
	}

	// init suite service (suite)

	authService := auth.New(log, storage, storage, storage, tokenTTL)

	grpcApp := grpcapp.New(log, authService, grpcPort)

	return &App{
		GRPCSrv: grpcApp,
	}
}
