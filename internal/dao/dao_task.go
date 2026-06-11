package dao

import (
	"context"

	"github.com/Perlishnov/TODO_APP/internal/models"
)

type TaskDAO interface{
	Create(ctx context.Context, task *models.Task) error
	GetById(ctx context.Context, id string) (*models.Task, error)
    List(ctx context.Context, filter *models.TaskFilter) ([]models.Task, error)   // changed
	Update(ctx context.Context, task *models.Task) error
	Delete(ctx context.Context, id string) error
	CountByUserAndStatus(ctx context.Context, userId string ,status string) (int64, error)
}

