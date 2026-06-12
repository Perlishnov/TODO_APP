package utils

import (
	"context"
	"fmt"

	"github.com/Perlishnov/TODO_APP/internal/dao"
)

const MaxInProgressTasks = 3

// ValidateTaskLimit checks if the user can create/update a task to "IN_PROGRESS"
func ValidateTaskLimit(ctx context.Context, taskDAO dao.TaskDAO, userID, status string) error {
	if status != "IN_PROGRESS" {
		return nil // only limit IN_PROGRESS
	}
	count, err := taskDAO.CountByUserAndStatus(ctx, userID, "IN_PROGRESS")
	if err != nil {
		return fmt.Errorf("failed to count tasks: %w", err)
	}
	if count >= MaxInProgressTasks {
		return fmt.Errorf("user cannot have more than %d tasks in progress", MaxInProgressTasks)
	}
	return nil
}
