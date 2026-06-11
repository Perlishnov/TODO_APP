package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/Perlishnov/TODO_APP/internal/dao"
	"github.com/Perlishnov/TODO_APP/internal/models"
	"github.com/Perlishnov/TODO_APP/internal/utils"
	"github.com/sirupsen/logrus"
)

type TaskService interface {
	Create(ctx context.Context, task *models.Task) error
	GetById(ctx context.Context, id, requestingUserID string) (*models.Task, error)
	List(ctx context.Context, filter *models.TaskFilter, userID string) ([]models.Task, error) // changed
	Update(ctx context.Context, task *models.Task, requestingUserID string) error
	Delete(ctx context.Context, id, requestingUserID string) error
	CountByUserAndStatus(ctx context.Context, userId string, status string) (int64, error)
}

type taskService struct {
	taskDAO dao.TaskDAO
	jwtUtil utils.JWTService
	logger  *logrus.Logger
}

func NewTaskService(taskDao dao.TaskDAO, jwtUtil utils.JWTService, logger *logrus.Logger) TaskService {
	return &taskService{
		taskDAO: taskDao,
		jwtUtil: jwtUtil,
		logger:  logger,
	}
}

func (s *taskService) Create(ctx context.Context, task *models.Task) error {
    // Validate required fields
    if task.Title == "" {
        return fmt.Errorf("title field cannot be empty")
    }
    if task.Description == "" {
        return fmt.Errorf("description field cannot be empty")
    }
    // Validate status
    validStatuses := map[string]bool{"todo": true, "in_progress": true, "done": true}
    if !validStatuses[string(task.Status)] {
        return fmt.Errorf("invalid status: must be one of todo, in_progress, done")
    }
    // Enforce max 3 in‑progress tasks if status is "in_progress"
    if task.Status == "in_progress" {
        if err := utils.ValidateTaskLimit(ctx, s.taskDAO, task.UserID, string(task.Status)); err != nil {
            return err
        }
    }
    // Proceed with creation
    if err := s.taskDAO.Create(ctx, task); err != nil {
        return fmt.Errorf("failed to create task: %w", err)
    }
    return nil
}

func (s *taskService) GetById(ctx context.Context, id, requestingUserID string) (*models.Task, error) {
	task, err := s.taskDAO.GetById(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get task: %w", err)
	}
	if task == nil {
		return nil, errors.New("task not found")
	}

	if task.UserID != requestingUserID {
		return nil, errors.New("access denied")
	}

	return task, nil
}

func (s *taskService) List(ctx context.Context, filter *models.TaskFilter, userID string) ([]models.Task, error) {
	if filter == nil {
		filter = &models.TaskFilter{}
	}
	// Enforce pagination defaults
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PageSize < 1 {
		filter.PageSize = 10
	}
	if filter.PageSize > 100 {
		filter.PageSize = 100
	}
	if filter.SortBy == "" {
		filter.SortBy = "created_at"
	}
	if filter.SortDir == "" {
		filter.SortDir = "desc"
	}

	// Always filter by the requesting user's ID – users see only their own tasks
	filter.UserID = userID

	tasks, err := s.taskDAO.List(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to list tasks: %w", err)
	}
	return tasks, nil
}

func (s *taskService) Update(ctx context.Context, task *models.Task, requestingUserID string) error {
	// Fetch existing task
	existing, err := s.taskDAO.GetById(ctx, task.ID)
	if err != nil {
		return fmt.Errorf("failed to get existing task: %w", err)
	}
	if existing == nil {
		return errors.New("task not found")
	}

	// Authorization: only owner or admin can update
	if existing.UserID != requestingUserID {
		return errors.New("access denied")
	}

	// If status is changing to "in_progress", enforce limit
	if task.Status == "in_progress" && existing.Status != "in_progress" {
		count, err := s.taskDAO.CountByUserAndStatus(ctx, existing.UserID, "in_progress")
		if err != nil {
			return fmt.Errorf("failed to count in-progress tasks: %w", err)
		}
		if count >= 3 {
			return errors.New("cannot have more than 3 tasks in progress")
		}
	}

	// Preserve fields that shouldn't be updated (optional: prevent changing UserID)
	task.UserID = existing.UserID
	task.CreatedAt = existing.CreatedAt

	// Perform update
	if err := s.taskDAO.Update(ctx, task); err != nil {
		return fmt.Errorf("failed to update task: %w", err)
	}
	return nil
}

func (s *taskService) Delete(ctx context.Context, id, requestingUserID string) error {
	// Fetch existing task
	existing, err := s.taskDAO.GetById(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get existing task: %w", err)
	}
	if existing == nil {
		return errors.New("task not found")
	}

	// Authorization: only owner
	if existing.UserID != requestingUserID {
		return errors.New("access denied")
	}
	// Perform update
	if err := s.taskDAO.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete task: %w", err)
	}
	return nil
}

func (s *taskService) CountByUserAndStatus(ctx context.Context, userId, status string) (int64, error) {
	if userId == "" {
		return 0, fmt.Errorf("Missing user property")
	}
	if status == "" {
		return 0, fmt.Errorf("Missing status property")
	}

	count, err := s.taskDAO.CountByUserAndStatus(ctx, userId, status)
	if err != nil {
		return 0, fmt.Errorf("failed to retrieve the amount tasks with status: %s that belong to the user of Id: %s", status, userId)
	}

	return count, nil
}
