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
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type TaskDAOMongo struct {
	collection *mongo.Collection
}

func NewTaskDAOMongo(db *mongo.Database, logger *logrus.Logger) TaskDAO {
	collection := db.Collection("tasks")

	return &TaskDAOMongo{collection: collection}
}

func (d *TaskDAOMongo) GetById(ctx context.Context, id string) (*models.Task, error) {
	objID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("invalid task id format")
	}
	filter := bson.M{"_id": objID}
	var task models.Task
	err = d.collection.FindOne(ctx, filter).Decode(&task)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to retrieve task with id %s", id)
	}

	return &task, nil
}

// dao/task_dao_mongo.go - List implementation
func (d *TaskDAOMongo) List(ctx context.Context, filter *models.TaskFilter) ([]models.Task, error) {
	// Build MongoDB filter
	mongoFilter := bson.M{}
	if filter != nil {
		if filter.Status != "" {
			mongoFilter["status"] = filter.Status
		}
		if filter.Search != "" {
			// case‑insensitive regex search on title and description
			regex := bson.M{"$regex": filter.Search, "$options": "i"}
			mongoFilter["$or"] = []bson.M{
				{"title": regex},
				{"description": regex},
			}
		}
		if filter.UserID != "" {
			mongoFilter["user_id"] = filter.UserID
		}
	}

	// Pagination
	page := 1
	pageSize := 10
	if filter != nil {
		if filter.Page > 0 {
			page = filter.Page
		}
		if filter.PageSize > 0 {
			pageSize = filter.PageSize
		}
		if pageSize > 100 {
			pageSize = 100
		}
	}
	skip := int64((page - 1) * pageSize)
	limit := int64(pageSize)

	findOpts := options.Find()
	findOpts.SetSkip(skip)
	findOpts.SetLimit(limit)

	// Sorting
	if filter != nil && filter.SortBy != "" {
		sortDir := 1 // asc
		if filter.SortDir == "desc" {
			sortDir = -1
		}
		findOpts.SetSort(bson.D{{Key: filter.SortBy, Value: sortDir}})
	} else {
		// default sort by creation date descending
		findOpts.SetSort(bson.D{{Key: "created_at", Value: -1}})
	}

	cursor, err := d.collection.Find(ctx, mongoFilter, findOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to find tasks: %w", err)
	}
	defer cursor.Close(ctx)

	var tasks []models.Task
	if err = cursor.All(ctx, &tasks); err != nil {
		return nil, fmt.Errorf("failed to decode tasks: %w", err)
	}
	return tasks, nil
}

func (d *TaskDAOMongo) Create(ctx context.Context, task *models.Task) error {
	task.CreatedAt = time.Now()
	task.UpdatedAt = time.Now()

	result, err := d.collection.InsertOne(ctx, task)
	if err != nil {
		return fmt.Errorf("failed to create task: %w", err)
	}
	if oid, ok := result.InsertedID.(bson.ObjectID); ok {
		task.ID = oid.Hex()
	}
	return nil
}

func (d *TaskDAOMongo) Update(ctx context.Context, task *models.Task) error {
	task.UpdatedAt = time.Now()
	objID, err := bson.ObjectIDFromHex(task.ID)
	if err != nil {
		return fmt.Errorf("invalid task id format")
	}
	filter := bson.M{"_id": objID}
	update := bson.M{"$set": task}
	result, err := d.collection.UpdateOne(ctx, filter, update)

	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}
	if result.MatchedCount == 0 {
		return fmt.Errorf("task with id %s not found", task.ID)
	}

	return nil
}

func (d *TaskDAOMongo) Delete(ctx context.Context, id string) error {
	objID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("invalid task id format")
	}
	filter := bson.M{"_id": objID}
	result, err := d.collection.DeleteOne(ctx, filter)

	if err != nil {
		return fmt.Errorf("failed to delete task: %w", err)
	}
	if result.DeletedCount == 0 {
		return fmt.Errorf("task with id %s not found", id)
	}
	return nil
}

func (d *TaskDAOMongo) CountByUserAndStatus(ctx context.Context, userId string, status string) (int64, error) {
	filter := bson.M{"user_id": userId, "status": status}
	return d.collection.CountDocuments(ctx, filter)
}
