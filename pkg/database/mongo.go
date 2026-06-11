package database

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/Perlishnov/TODO_APP/internal/config"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)


func NewMongoConnection( config *config.Config ,logger *logrus.Logger) (*mongo.Database, error)  {
	uri := config.ConnectionUri

	if uri == "" {
        uri = "mongodb://localhost:27017"
    }

    dbName := os.Getenv("DB_NAME")
    if dbName == "" {
        dbName = "task_api_db"
    }

	clientOps := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(clientOps)

	if err != nil {
		return nil, fmt.Errorf("")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5 * time.Second)
	defer cancel()

	if err := client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}
	logger.Info("Connected to MongoDB")
	return client.Database(dbName), nil
}
