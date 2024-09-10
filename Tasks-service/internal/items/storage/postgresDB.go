package storage

import (
	"database/sql"
	"fmt"
	"log"

	"task-service/internal/items/config"

	_ "github.com/lib/pq"
)

func ConnecPostgrestDB(config *config.Config) (*sql.DB, error) {
	connStr := fmt.Sprintf("user=%s dbname=%s password=%s host=%s port=%s sslmode=disable",
		config.Postgres.User,
		config.Postgres.DBName,
		config.Postgres.Password,
		config.Postgres.Host,
		config.Postgres.Port,
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, err
	}

	log.Printf("--------------------------- Connected to the database %s ---------------------------\n", config.Postgres.DBName)

	return db, nil
}
