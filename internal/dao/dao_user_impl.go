package dao

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Perlishnov/TODO_APP/internal/models"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type UserDAOMongo struct{
	collection *mongo.Collection
}

func NewUserDAO(db *mongo.Database, logger *logrus.Logger) UserDAO {
	collection := db.Collection("users")

	return &UserDAOMongo{collection: collection}

}

func (d *UserDAOMongo) Create(ctx context.Context, user *models.User) error{
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	
	_, err := d.collection.InsertOne(ctx, user)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return fmt.Errorf("user with email %s already exists", user.Email)
		}
		return fmt.Errorf("failed to create user: %w", err)

	}
	return nil
}

func (d *UserDAOMongo) GetByEmail(ctx context.Context, email string) (*models.User, error)  {
	var user models.User
	filter := bson.M{"email": email}
	err := d.collection.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil,nil
		}
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}
	return &user, nil
}

func (d *UserDAOMongo) GetByID(ctx context.Context, id string) (*models.User, error) {
	filter := bson.M{"_id": id}
	var user models.User
	err:= d.collection.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}
	return &user, nil
}
