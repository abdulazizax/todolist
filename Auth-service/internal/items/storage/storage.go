package storage

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	pb "auth-service/genproto/auth"
	"auth-service/internal/items/config"
	jwttokens "auth-service/internal/items/jwt"
	"auth-service/internal/items/redisservice"
	"auth-service/internal/items/repository"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type Storage struct {
	redisService *redisservice.RedisService
	postgres     *sql.DB
	queryBuilder sq.StatementBuilderType
	cfg          *config.Config
	logger       *slog.Logger
}

func New(postgres *sql.DB, queryBuilder sq.StatementBuilderType, cfg *config.Config, logger *slog.Logger) repository.IAuthRepo {
	return &Storage{
		postgres:     postgres,
		queryBuilder: queryBuilder,
		cfg:          cfg,
		logger:       logger,
	}
}

func (s *Storage) Register(ctx context.Context, in *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	id := uuid.New().String()
	created_at := time.Now()

	password_hash, err := hashPassword(in.Password)
	if err != nil {
		s.logger.Error("Error while hashing password:", slog.String("err:", err.Error()))
		return nil, err
	}

	query, args, err := s.queryBuilder.Insert("users").
		Columns(
			"id",
			"email",
			"password_hash",
			"created_at",
		).Values(
		id,
		in.Email,
		password_hash,
		created_at,
	).ToSql()
	if err != nil {
		s.logger.Error("Error while building a query")
		return nil, err
	}

	_, err = s.postgres.ExecContext(ctx, query, args...)
	if err != nil {
		s.logger.Error("Error while executing a query")
		return nil, err
	}

	return &pb.RegisterResponse{
		UserId: id,
	}, nil
}

func (s *Storage) Login(ctx context.Context, in *pb.LoginRequest) (*pb.LoginResponse, error) {
	tx, err := s.postgres.BeginTx(ctx, nil)
	if err != nil {
		s.logger.Error("Error while starting a transaction")
		return nil, err
	}
	defer tx.Rollback()

	query, args, err := s.queryBuilder.Select(
		"id",
		"role",
		"password_hash",
	).From("users").
		Where(sq.Eq{"email": in.Email}).
		ToSql()
	if err != nil {
		s.logger.Error("Error while building a query")
		return nil, err
	}

	var id, role, hashedPassword string

	err = s.postgres.QueryRowContext(ctx, query, args...).Scan(&id, &role, &hashedPassword)
	if err != nil {
		s.logger.Error("Error while executing a query")
		return nil, err
	}

	_, err = checkPassword(hashedPassword, in.Password)
	if err != nil {
		s.logger.Error("Error while checking password")
		return nil, err
	}

	accessToken, err := jwttokens.GenerateAccessToken(id, in.Email, role, s.cfg.JWT.SecretKey)
	if err != nil {
		s.logger.Error("Error while generating access token")
		return nil, err
	}

	refreshToken, err := jwttokens.GenerateRefreshToken(id, in.Email, role, s.cfg.JWT.SecretKey)
	if err != nil {
		s.logger.Error("Error while generating refresh token")
		return nil, err
	}

	query, args, err = s.queryBuilder.Update("users").
		Set("refresh_token", refreshToken).
		Set("is_active", true).
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		s.logger.Error("Error while building a query")
		return nil, err
	}

	_, err = tx.ExecContext(ctx, query, args...)
	if err != nil {
		s.logger.Error("Error while executing a query")
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		s.logger.ErrorContext(ctx, "error while committing transaction", slog.String("error", err.Error()))
		return nil, err
	}

	return &pb.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *Storage) Logout(ctx context.Context, in *pb.LogoutRequest) (*pb.LogoutResponse, error) {
	tx, err := s.postgres.BeginTx(ctx, nil)
	if err != nil {
		s.logger.Error("Error while starting a transaction")
		return nil, err
	}
	defer tx.Rollback()

	query, args, err := s.queryBuilder.Update("users").
		Set("refresh_token", "").
		Set("is_active", false).
		Where(sq.Eq{"id": in.UserId}).
		ToSql()
	if err != nil {
		s.logger.Error("Error while building a query")
		return nil, err
	}

	_, err = tx.ExecContext(ctx, query, args...)
	if err != nil {
		s.logger.Error("Error while executing a query")
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		s.logger.ErrorContext(ctx, "error while committing transaction", slog.String("error", err.Error()))
		return nil, err
	}

	return &pb.LogoutResponse{
		Message: "Logout successful",
	}, nil
}

func (s *Storage) CreateAdmin(ctx context.Context, in *pb.CreateAdminRequest) (*pb.CreateAdminResponse, error) {
	updated_at := time.Now()

	query, args, err := s.queryBuilder.Update("users").
		Set("role", "admin").
		Set("updated_at", updated_at).
		Where(sq.Eq{"id": in.UserId}).
		Where("deleted_at IS NULL").
		ToSql()
	if err != nil {
		s.logger.Error("Error while building a query")
		return nil, err
	}

	_, err = s.postgres.ExecContext(ctx, query, args...)
	if err != nil {
		s.logger.Error("Error while executing a query")
		return nil, err
	}

	return &pb.CreateAdminResponse{
		Message: "Admin created successfully",
	}, nil
}

func (s *Storage) DeleteUser(ctx context.Context, in *pb.DeleteUserRequest) (*pb.DeleteUserResponse, error) {
	deleted_at := time.Now()

	query, args, err := s.queryBuilder.Update("users").
		Set("deleted_at", deleted_at).
		Where(sq.Eq{"id": in.UserId}).
		ToSql()
	if err != nil {
		s.logger.Error("Error while building a query")
		return nil, err
	}

	_, err = s.postgres.ExecContext(ctx, query, args...)
	if err != nil {
		s.logger.Error("Error while executing a query")
		return nil, err
	}

	return &pb.DeleteUserResponse{
		Message: "User deleted successfully",
	}, nil
}

func (s *Storage) GetUserByEmail(ctx context.Context, in *pb.GetUserByEmailRequest) (*pb.RegisterResponse, error) {
	query, args, err := s.queryBuilder.Select("id").
		From("users").
		Where(sq.Eq{"email": in.Email}).
		ToSql()
	if err != nil {
		s.logger.Error("Error while building a query")
		return nil, err
	}

	var id string

	err = s.postgres.QueryRowContext(ctx, query, args...).Scan(&id)
	if err != nil {
		s.logger.Error("Error while executing a query")
		return nil, err
	}

	return &pb.RegisterResponse{
		UserId: id,
	}, nil
}

func (s *Storage) UpdatePasswordService(ctx context.Context, req *pb.UpdateUserPasswordRequest) (*pb.RawResponse, error) {
	err := s.verifyEmail(context.Background(), req.Email, int(req.VerificationCode))
	if err != nil {
		s.logger.Error("Error while verifying email")
		return nil, err
	}

	tx, err := s.postgres.BeginTx(ctx, nil)
	if err != nil {
		s.logger.Error("Error while starting a transaction")
		return nil, err
	}
	defer tx.Rollback()

	query, args, err := s.queryBuilder.Update("users").
		Set("password_hash", req.NewPassword).
		Where(sq.Eq{"id": req.UserId}).
		ToSql()
	if err != nil {
		s.logger.Error("Error while building a query")
		return nil, err
	}

	_, err = tx.ExecContext(ctx, query, args...)
	if err != nil {
		s.logger.Error("Error while executing a query")
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		s.logger.ErrorContext(ctx, "error while committing transaction", slog.String("error", err.Error()))
		return nil, err
	}

	return &pb.RawResponse{
		Message: "Password updated successfully",
	}, nil
}

func (s *Storage) SendVerificationCode(ctx context.Context, req *pb.SendVerificationCodeRequest) (*pb.RawResponse, error) {
	query, args, err := s.queryBuilder.Select("email").
		From("users").
		Where(sq.Eq{"email": req.Email}).
		ToSql()
	if err != nil {
		s.logger.Error("Error while building a query")
		return nil, err
	}

	var email string
	err = s.postgres.QueryRowContext(ctx, query, args...).Scan(&email)
	if err != nil {
		s.logger.Error("Error while executing a query")
		return nil, err
	}

	if email != req.Email {
		s.logger.Info("Email not found")
		return nil, fmt.Errorf("email not found")
	}

	if err := s.sendVerificationCode(context.Background(), req.Email); err != nil {
		s.logger.Error("Error while sending verification code")
		return nil, err
	}

	return &pb.RawResponse{
		Message: "Verification code sent to your email!",
	}, nil
}

func hashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

func checkPassword(hashedPassword, password string) (bool, error) {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil, err
}
