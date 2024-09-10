package storage

import (
	"context"
	"database/sql"
	"log/slog"
	"time"

	pb "task-service/genproto/task"
	"task-service/internal/items/config"
	"task-service/internal/items/redisservice"
	"task-service/internal/items/repository"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"google.golang.org/protobuf/types/known/emptypb"
)

type Storage struct {
	redisService *redisservice.RedisService
	postgres     *sql.DB
	mongo        *mongo.Collection
	queryBuilder sq.StatementBuilderType
	cfg          *config.Config
	logger       *slog.Logger
}

func New(postgres *sql.DB, mongo *mongo.Collection, queryBuilder sq.StatementBuilderType, cfg *config.Config, logger *slog.Logger) repository.ITaskRepo {
	return &Storage{
		postgres:     postgres,
		mongo:        mongo,
		queryBuilder: queryBuilder,
		cfg:          cfg,
		logger:       logger,
	}
}

func (s *Storage) CreateTask(ctx context.Context, req *pb.CreateTaskRequest) (*pb.CreateTaskResponse, error) {
	taskID := uuid.New().String()
	createdAt := time.Now()

	tx, err := s.postgres.BeginTx(ctx, nil)
	if err != nil {
		s.logger.Error("Error while starting transaction", slog.Any("err", err))
		return nil, err
	}

	query, args, err := s.queryBuilder.Insert("tasks").
		Columns(
			"id",
			"user_id",
			"title",
			"created_at",
			"updated_at").
		Values(
			taskID,
			req.Metadata.UserId,
			req.Metadata.Title,
			createdAt,
			createdAt,
		).ToSql()
	if err != nil {
		s.logger.Error("Error while building query", slog.Any("err", err))
		tx.Rollback()
		return nil, err
	}

	_, err = tx.ExecContext(ctx, query, args...)
	if err != nil {
		s.logger.Error("Error while executing query", slog.Any("err", err))
		tx.Rollback()
		return nil, err
	}

	taskDetails := bson.M{
		"id":          taskID,
		"description": req.Details.Description,
		"status":      req.Details.Status,
		"priority":    req.Details.Priority,
		"due_date":    req.Details.DueDate.AsTime(),
		"updated_at":  createdAt,
	}

	_, err = s.mongo.InsertOne(ctx, taskDetails)
	if err != nil {
		s.logger.Error("Error while inserting into MongoDB", slog.Any("err", err))
		tx.Rollback()
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		s.logger.Error("Error while committing transaction", slog.Any("err", err))
		return nil, err
	}

	return &pb.CreateTaskResponse{Id: taskID}, nil
}

func (s *Storage) UpdateTask(ctx context.Context, req *pb.UpdateTaskRequest) (*emptypb.Empty, error) {
	tx, err := s.postgres.BeginTx(ctx, nil)
	if err != nil {
		s.logger.Error("Error while starting transaction", slog.Any("err", err))
		return nil, err
	}

	query, args, err := s.queryBuilder.Update("tasks").
		Set("title", req.Metadata.Title).
		Set("updated_at", time.Now()).
		Where(sq.Eq{"id": req.Id}).
		ToSql()
	if err != nil {
		s.logger.Error("Error while building query", slog.Any("err", err))
		tx.Rollback()
		return nil, err
	}

	_, err = tx.ExecContext(ctx, query, args...)
	if err != nil {
		s.logger.Error("Error while executing query", slog.Any("err", err))
		tx.Rollback()
		return nil, err
	}

	filter := bson.M{"id": req.Id}
	update := bson.M{
		"$set": bson.M{
			"description": req.Details.Description,
			"status":      req.Details.Status,
			"priority":    req.Details.Priority,
			"due_date":    req.Details.DueDate.AsTime(),
			"updated_at":  time.Now(),
		},
	}

	_, err = s.mongo.UpdateOne(ctx, filter, update)
	if err != nil {
		s.logger.Error("Error while updating MongoDB", slog.Any("err", err))
		tx.Rollback()
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		s.logger.Error("Error while committing transaction", slog.Any("err", err))
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (s *Storage) DeleteTask(ctx context.Context, req *pb.DeleteTaskRequest) (*emptypb.Empty, error) {
	deleted_at := time.Now()

	// Yangi transaction boshlash
	tx, err := s.postgres.BeginTx(ctx, nil)
	if err != nil {
		s.logger.Error("Error while starting transaction", slog.Any("err", err))
		return nil, err
	}

	// PostgreSQL uchun so'rovni tayyorlash
	query, args, err := s.queryBuilder.Update("tasks").
		Set("deleted_at", deleted_at). // O'chirilgan vaqtni saqlash
		Where(sq.Eq{"id": req.Id}).
		ToSql()
	if err != nil {
		s.logger.Error("Error while building query", slog.Any("err", err))
		tx.Rollback()
		return nil, err
	}

	// PostgreSQL ma'lumotlar bazasida so'rovni bajarish
	_, err = tx.ExecContext(ctx, query, args...)
	if err != nil {
		s.logger.Error("Error while executing query", slog.Any("err", err))
		tx.Rollback()
		return nil, err
	}

	// MongoDB uchun so'rovni tayyorlash
	filter := bson.M{"id": req.Id}
	update := bson.M{
		"$set": bson.M{
			"deleted_at": deleted_at, // O'chirilgan vaqtni saqlash
		},
	}

	// MongoDB ma'lumotlar bazasida so'rovni bajarish
	_, err = s.mongo.UpdateOne(ctx, filter, update)
	if err != nil {
		s.logger.Error("Error while updating MongoDB document", slog.Any("err", err))
		tx.Rollback()
		return nil, err
	}

	// Transactionni commit qilish
	if err = tx.Commit(); err != nil {
		s.logger.Error("Error while committing transaction", slog.Any("err", err))
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

// func (s *Storage) ListTasks(ctx context.Context, req *pb.ListTasksRequest) (*pb.ListTasksResponse, error) {
// 	query, args, err := s.queryBuilder.Select("id", "title", "created_at").
// 		From("tasks").
// 		Where(sq.Eq{"user_id": req.UserId}).
// 		ToSql()
// 	if err != nil {
// 		s.logger.Error("Error while building query", slog.Any("err", err))
// 		return nil, err
// 	}

// 	rows, err := s.postgres.QueryContext(ctx, query, args...)
// 	if err != nil {
// 		s.logger.Error("Error while executing query", slog.Any("err", err))
// 		return nil, err
// 	}
// 	defer rows.Close()

// 	var tasks []*pb.TaskMetadata
// 	for rows.Next() {
// 		var task pb.TaskMetadata
// 		var createdAt time.Time

// 		if err := rows.Scan(&task.Id, &task.Title, &createdAt); err != nil {
// 			s.logger.Error("Error while scanning row", slog.Any("err", err))
// 			return nil, err
// 		}

// 		task.CreatedAt = timestamppb.New(createdAt)
// 		tasks = append(tasks, &task)
// 	}

// 	if err := rows.Err(); err != nil {
// 		s.logger.Error("Error while iterating rows", slog.Any("err", err))
// 		return nil, err
// 	}

// 	return &pb.ListTasksResponse{Tasks: tasks}, nil
// }
