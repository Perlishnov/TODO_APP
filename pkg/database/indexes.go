package database

import (
	"context"

	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type IndexConfig struct{
	Collection string
	Keys bson.D
	Unique bool
	Name string
}

func SetupIndexes(db *mongo.Database, logger *logrus.Logger) error {
	indexes := []IndexConfig{
		{
			Collection: "users",
			Keys: bson.D{{Key: "email",Value: 1}},
			Unique: true,
			Name: "idx_users_email_unique",
		},
		{
            Collection: "tasks",
            Keys:       bson.D{{Key: "user_id", Value: 1}, {Key: "status", Value: 1}},
            Name:       "idx_tasks_user_status",
        },
        {
            Collection: "tasks",
            Keys:       bson.D{{Key: "user_id", Value: 1}, {Key: "created_at", Value: -1}},
            Name:       "idx_tasks_user_created_at",
        },
        {
            Collection: "tasks",
            Keys:       bson.D{{Key: "status", Value: 1}},
            Name:       "idx_tasks_status",
        },
	}

	for _, idx := range indexes{
		indexModel := mongo.IndexModel{
			Keys: idx.Keys,
		}
		if idx.Unique {
			indexModel.Options = options.Index().SetUnique(true)
		}
		if idx.Name != "" {
			if indexModel.Options == nil {
				indexModel.Options = options.Index()
			}
			indexModel.Options.SetName(idx.Name)
		}
		coll := db.Collection(idx.Collection)
		_, err := coll.Indexes().CreateOne(context.Background(), indexModel)
		if err != nil {
			logger.WithError(err).WithField("collection", idx.Collection).Error("failed to create index")
			return err
		}
		logger.WithField("collection", idx.Collection).WithField("index", idx.Name).Info("index created/verified")
	}
	return nil
}