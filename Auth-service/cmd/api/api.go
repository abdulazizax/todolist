package api

import (
	"context"
	"log"
	"log/slog"
	"net"
	"sync"

	"auth-service/internal/items/config"
	"auth-service/internal/items/msgbroker"
	"auth-service/internal/items/service"
	"auth-service/internal/items/storage"

	pb "auth-service/genproto/auth"

	sq "github.com/Masterminds/squirrel"
	"google.golang.org/grpc"
)

func Run(cfg *config.Config, logger *slog.Logger) error {
	db, err := storage.ConnectDB(cfg)
	if err != nil {
		logger.Error("error while connecting postgres:", slog.String("err", err.Error()))
		return err
	}

	sqrl := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	service := service.New(storage.New(
		db,
		sqrl,
		cfg,
		logger,
	), logger)

	msgBrokers := msgbroker.InitMessageBroker(cfg)
	kafkaconsumer := msgbroker.New(service, logger, msgBrokers, &sync.WaitGroup{})
	go kafkaconsumer.StartToConsume(context.Background())

	listener, err := net.Listen("tcp", cfg.Server.Port)
	if err != nil {
		logger.Error("error while starting server:", slog.String("err", err.Error()))
		return err
	}

	serverRegisterer := grpc.NewServer()

	pb.RegisterAuthServiceServer(serverRegisterer, service)

	logger.Info("Server has started running", slog.String("port", cfg.Server.Port))
	log.Printf("Server has started running on port %s", cfg.Server.Port)

	return serverRegisterer.Serve(listener)
}
