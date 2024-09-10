package api

import (
	"context"
	"log"
	"log/slog"
	"net"
	"sync"

	"task-service/internal/items/config"
	"task-service/internal/items/msgbroker"
	"task-service/internal/items/service"
	"task-service/internal/items/storage"

	pb "task-service/genproto/task"

	sq "github.com/Masterminds/squirrel"
	"google.golang.org/grpc"
)

func Run(cfg *config.Config, logger *slog.Logger) error {
	postgresdb, err := storage.ConnecPostgrestDB(cfg)
	if err != nil {
		logger.Error("error while connecting postgres:", slog.String("err", err.Error()))
		return err
	}

	mongodb, err := storage.ConnectMongoDB(cfg)
	if err != nil {
		logger.Error("error while connecting mongodb:", slog.String("err", err.Error()))
		return err
	}

	task_collection := mongodb.Collection("tasks")

	sqrl := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	service := service.New(storage.New(
		postgresdb,
		task_collection,
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

	pb.RegisterTaskServiceServer(serverRegisterer, service)

	logger.Info("Server has started running", slog.String("port", cfg.Server.Port))
	log.Printf("Server has started running on port %s", cfg.Server.Port)

	return serverRegisterer.Serve(listener)
}
