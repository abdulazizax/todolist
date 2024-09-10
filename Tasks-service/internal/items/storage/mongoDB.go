package storage

import (
	"context"
	"fmt"
	"log"
	"task-service/internal/items/config"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func ConnectMongoDB(config *config.Config) (*mongo.Database, error) {
	uri := fmt.Sprintf("mongodb://%s:%s@%s:%s",
		config.Mongo.User,
		config.Mongo.Password,
		config.Mongo.Host,
		config.Mongo.Port)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOptions := options.Client().ApplyURI(uri)

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, err
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		return nil, err
	}

	db := client.Database(config.Mongo.DBName)

	log.Printf("--------------------------- Connected to the database %s --------------------------------\n", config.Mongo.DBName)

	return db, nil
}
