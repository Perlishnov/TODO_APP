package utils

import (
    "context"
    "fmt"

    "github.com/Perlishnov/TODO_APP/internal/dao"
)

const MaxInProgressTasks = 3

// ValidateTaskLimit checks if the user can create/update a task to "in_progress"
func ValidateTaskLimit(ctx context.Context, taskDAO dao.TaskDAO, userID, status string) error {
    if status != "in_progress" {
        return nil // only limit in_progress
    }
    count, err := taskDAO.CountByUserAndStatus(ctx, userID,"in_progress")
    if err != nil {
        return fmt.Errorf("failed to count tasks: %w", err)
    }
    if count >= MaxInProgressTasks {
        return fmt.Errorf("user cannot have more than %d tasks in progress", MaxInProgressTasks)
    }
    return nil
}