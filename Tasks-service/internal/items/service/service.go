package service

import (
	"context"
	"log/slog"

	pb "task-service/genproto/task"

	"task-service/internal/items/repository"

	"google.golang.org/protobuf/types/known/emptypb"
)

type Service struct {
	pb.UnimplementedTaskServiceServer
	storage repository.ITaskRepo
	logger  *slog.Logger
}

func New(storage repository.ITaskRepo, logger *slog.Logger) *Service {
	return &Service{
		storage: storage,
		logger:  logger,
	}
}

func (s *Service) CreateTask(ctx context.Context, req *pb.CreateTaskRequest) (*pb.CreateTaskResponse, error) {
	s.logger.Info("CreateTask called")
	return s.storage.CreateTask(ctx, req)
}

func (s *Service) UpdateTask(ctx context.Context, req *pb.UpdateTaskRequest) (*emptypb.Empty, error) {
	s.logger.Info("UpdateTask called")
	return s.storage.UpdateTask(ctx, req)
}

func (s *Service) DeleteTask(ctx context.Context, req *pb.DeleteTaskRequest) (*emptypb.Empty, error) {
	s.logger.Info("DeleteTask called")
	return s.storage.DeleteTask(ctx, req)
}
