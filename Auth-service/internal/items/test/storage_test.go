package test

import (
	pb "auth-service/genproto/auth"
	"auth-service/internal/items/config"
	"auth-service/internal/items/repository"
	"auth-service/internal/items/storage"
	"database/sql"

	"context"
	"testing"

	sq "github.com/Masterminds/squirrel"

	"log"
	"log/slog"
	"os"
)

func setupStorage() (repository.IAuthRepo, *sql.DB) {
	config, err := config.New()
	if err != nil {
		log.Fatalln(err)
	}

	logFile, err := os.OpenFile("test.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer logFile.Close()

	logger := slog.New(slog.NewJSONHandler(logFile, nil))

	db, err := storage.ConnectDB(config)
	if err != nil {
		logger.Error("error while connecting postgres:", slog.String("err:", err.Error()))
	}

	sqrl := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	return storage.New(
		db,
		sqrl,
		config,
		logger,
	), db
}

func TestRegister(t *testing.T) {
	storage, db := setupStorage()
	ctx := context.Background()

	test := pb.RegisterRequest{
		Email:    "test@gmail.com",
		Password: "test12345",
	}

	_, err := storage.Register(ctx, &test)
	if err != nil {
		t.Errorf("error while registering: %v", err)
	}

	query := "DELETE FROM users WHERE email = $1"
	_, err = db.ExecContext(ctx, query, test.Email)
	if err != nil {
		t.Errorf("error while deleting user: %v", err)
	}
}
