package dao

import (
	"context"
	"github.com/Perlishnov/TODO_APP/internal/models"
)

type UserDAO interface {
	Create(ctx context.Context, user *models.User) error
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	GetByID(ctx context.Context, id string) (*models.User, error)
}
