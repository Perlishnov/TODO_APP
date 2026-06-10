package dao

import (
	"context"

	"github.com/Perlishnov/TODO_APP/internal/models"
)

type AuthDAO interface{
	Signup (ctx context.Context, user *models.User) error
	Login (ctx context.Context, user *models.User) (*models.User, error)
}

