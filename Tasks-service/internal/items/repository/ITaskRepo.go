package repository

import (
	"context"

	pb "task-service/genproto/task"

	"google.golang.org/protobuf/types/known/emptypb"
)

type (
	ITaskRepo interface {
		CreateTask(ctx context.Context, req *pb.CreateTaskRequest) (*pb.CreateTaskResponse, error)
		UpdateTask(ctx context.Context, req *pb.UpdateTaskRequest) (*emptypb.Empty, error)
		DeleteTask(ctx context.Context, req *pb.DeleteTaskRequest) (*emptypb.Empty, error)
	}
)
