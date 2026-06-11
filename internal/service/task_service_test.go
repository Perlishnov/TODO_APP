package service

import (
	"context"
	"errors"
	"testing"

	"github.com/Perlishnov/TODO_APP/internal/models"
	"github.com/Perlishnov/TODO_APP/mocks"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestTaskService_Create(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.FatalLevel)

	tests := []struct {
		name        string
		task        *models.Task
		setupMocks  func(mockTaskDAO *mocks.TaskDAO)
		expectedErr string
	}{
		{
			name: "success - todo status (no limit check)",
			task: &models.Task{
				Title:       "My first task",
				Description: "desc",
				Status:      "todo",
				UserID:      "user123",
			},
			setupMocks: func(mockTaskDAO *mocks.TaskDAO) {
				// CountByUserAndStatus NOT called because status != "in_progress"
				mockTaskDAO.On("Create", mock.Anything, mock.AnythingOfType("*models.Task")).Return(nil)
			},
			expectedErr: "",
		},
		{
			name: "success - in_progress status (limit not reached)",
			task: &models.Task{
				Title:       "In progress task",
				Description: "desc",
				Status:      "in_progress",
				UserID:      "user456",
			},
			setupMocks: func(mockTaskDAO *mocks.TaskDAO) {
				mockTaskDAO.On("CountByUserAndStatus", mock.Anything, "user456", "in_progress").Return(int64(0), nil)
				mockTaskDAO.On("Create", mock.Anything, mock.AnythingOfType("*models.Task")).Return(nil)
			},
			expectedErr: "",
		},
		{
			name: "fail - already 3 in_progress tasks",
			task: &models.Task{
				Title:       "Fourth task",
				Description: "desc",
				Status:      "in_progress",
				UserID:      "user789",
			},
			setupMocks: func(mockTaskDAO *mocks.TaskDAO) {
				mockTaskDAO.On("CountByUserAndStatus", mock.Anything, "user789", "in_progress").Return(int64(3), nil)
				// Create should NOT be called
			},
			expectedErr: "cannot have more than 3 tasks in progress",
		},
		{
			name: "fail - invalid status",
			task: &models.Task{
				Title:       "Invalid",
				Description: "desc",
				Status:      "unknown",
				UserID:      "userX",
			},
			setupMocks:  func(mockTaskDAO *mocks.TaskDAO) {},
			expectedErr: "invalid status",
		},
		{
			name: "fail - CountByUserAndStatus database error",
			task: &models.Task{
				Title:       "DB error task",
				Description: "desc",
				Status:      "in_progress",
				UserID:      "userErr",
			},
			setupMocks: func(mockTaskDAO *mocks.TaskDAO) {
				mockTaskDAO.On("CountByUserAndStatus", mock.Anything, "userErr", "in_progress").Return(int64(0), errors.New("db down"))
			},
			expectedErr: "failed to count tasks",
		},
		{
			name: "fail - Create database error",
			task: &models.Task{
				Title:       "Create fail",
				Description: "desc",
				Status:      "todo",
				UserID:      "userCreate",
			},
			setupMocks: func(mockTaskDAO *mocks.TaskDAO) {
				mockTaskDAO.On("Create", mock.Anything, mock.AnythingOfType("*models.Task")).Return(errors.New("duplicate key"))
			},
			expectedErr: "failed to create task",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockTaskDAO := new(mocks.TaskDAO)
			tt.setupMocks(mockTaskDAO)

			svc := NewTaskService(mockTaskDAO, nil, logger)
			err := svc.Create(context.Background(), tt.task)

			if tt.expectedErr != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr)
			} else {
				assert.NoError(t, err)
			}
			mockTaskDAO.AssertExpectations(t)
		})
	}
}

func TestTaskService_GetById(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.FatalLevel)

	tests := []struct {
		name            string
		taskID          string
		requestingUserID string
		setupMocks      func(mockTaskDAO *mocks.TaskDAO)
		expectedTask    *models.Task
		expectedErr     string
	}{
		{
			name:            "success - user owns task",
			taskID:          "task1",
			requestingUserID: "user1",
			setupMocks: func(mockTaskDAO *mocks.TaskDAO) {
				task := &models.Task{ID: "task1", UserID: "user1", Title: "My task", Description: "desc", Status: "todo"}
				mockTaskDAO.On("GetById", mock.Anything, "task1").Return(task, nil)
			},
			expectedTask: &models.Task{ID: "task1", UserID: "user1", Title: "My task", Description: "desc", Status: "todo"},
			expectedErr:  "",
		},
		{
			name:            "fail - task not found",
			taskID:          "nonexistent",
			requestingUserID: "user1",
			setupMocks: func(mockTaskDAO *mocks.TaskDAO) {
				mockTaskDAO.On("GetById", mock.Anything, "nonexistent").Return(nil, nil)
			},
			expectedErr: "task not found",
		},
		{
			name:            "fail - access denied (task belongs to another user)",
			taskID:          "task2",
			requestingUserID: "hacker",
			setupMocks: func(mockTaskDAO *mocks.TaskDAO) {
				mockTaskDAO.On("GetById", mock.Anything, "task2").Return(&models.Task{ID: "task2", UserID: "owner"}, nil)
			},
			expectedErr: "access denied",
		},
		{
			name:            "fail - GetById database error",
			taskID:          "task3",
			requestingUserID: "user3",
			setupMocks: func(mockTaskDAO *mocks.TaskDAO) {
				mockTaskDAO.On("GetById", mock.Anything, "task3").Return(nil, errors.New("db error"))
			},
			expectedErr: "failed to get task",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockTaskDAO := new(mocks.TaskDAO)
			tt.setupMocks(mockTaskDAO)

			svc := NewTaskService(mockTaskDAO, nil, logger)
			task, err := svc.GetById(context.Background(), tt.taskID, tt.requestingUserID)

			if tt.expectedErr != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr)
				assert.Nil(t, task)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedTask, task)
			}
			mockTaskDAO.AssertExpectations(t)
		})
	}
}

func TestTaskService_List(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.FatalLevel)

	tests := []struct {
		name        string
		userID      string
		filter      *models.TaskFilter
		setupMocks  func(mockTaskDAO *mocks.TaskDAO)
		expectedLen int
		expectedErr string
	}{
		{
			name:   "success - no filter",
			userID: "user1",
			filter: &models.TaskFilter{},
			setupMocks: func(mockTaskDAO *mocks.TaskDAO) {
				tasks := []models.Task{{ID: "t1", Title: "Task 1", UserID: "user1"}, {ID: "t2", Title: "Task 2", UserID: "user1"}}
				mockTaskDAO.On("List", mock.Anything, mock.AnythingOfType("*models.TaskFilter")).Return(tasks, nil)
			},
			expectedLen: 2,
			expectedErr: "",
		},
		{
			name:   "success - filter by status",
			userID: "user2",
			filter: &models.TaskFilter{Status: "todo"},
			setupMocks: func(mockTaskDAO *mocks.TaskDAO) {
				tasks := []models.Task{{ID: "t3", Title: "Alpha", Status: "todo", UserID: "user2"}}
				mockTaskDAO.On("List", mock.Anything, mock.AnythingOfType("*models.TaskFilter")).Return(tasks, nil)
			},
			expectedLen: 1,
			expectedErr: "",
		},
		{
			name:   "fail - DAO error",
			userID: "userErr",
			filter: &models.TaskFilter{},
			setupMocks: func(mockTaskDAO *mocks.TaskDAO) {
				mockTaskDAO.On("List", mock.Anything, mock.AnythingOfType("*models.TaskFilter")).Return(nil, errors.New("db error"))
			},
			expectedErr: "failed to list tasks",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockTaskDAO := new(mocks.TaskDAO)
			tt.setupMocks(mockTaskDAO)

			svc := NewTaskService(mockTaskDAO, nil, logger)
			tasks, err := svc.List(context.Background(), tt.filter, tt.userID)

			if tt.expectedErr != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr)
				assert.Nil(t, tasks)
			} else {
				assert.NoError(t, err)
				assert.Len(t, tasks, tt.expectedLen)
			}
			mockTaskDAO.AssertExpectations(t)
		})
	}
}

func TestTaskService_Update(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.FatalLevel)

	tests := []struct {
		name            string
		task            *models.Task
		requestingUserID string
		setupMocks      func(mockTaskDAO *mocks.TaskDAO)
		expectedErr     string
	}{
		{
			name: "success - update task (no status change to in_progress)",
			task: &models.Task{
				ID:     "task1",
				Title:  "Updated title",
				Status: "todo",
			},
			requestingUserID: "user1",
			setupMocks: func(mockTaskDAO *mocks.TaskDAO) {
				existing := &models.Task{ID: "task1", UserID: "user1", Status: "todo"}
				mockTaskDAO.On("GetById", mock.Anything, "task1").Return(existing, nil)
				mockTaskDAO.On("Update", mock.Anything, mock.AnythingOfType("*models.Task")).Return(nil)
			},
			expectedErr: "",
		},
		{
			name: "success - change from todo to in_progress when limit not reached",
			task: &models.Task{
				ID:     "task2",
				Status: "in_progress",
			},
			requestingUserID: "user2",
			setupMocks: func(mockTaskDAO *mocks.TaskDAO) {
				existing := &models.Task{ID: "task2", UserID: "user2", Status: "todo"}
				mockTaskDAO.On("GetById", mock.Anything, "task2").Return(existing, nil)
				mockTaskDAO.On("CountByUserAndStatus", mock.Anything, "user2", "in_progress").Return(int64(2), nil)
				mockTaskDAO.On("Update", mock.Anything, mock.AnythingOfType("*models.Task")).Return(nil)
			},
			expectedErr: "",
		},
		{
			name: "fail - change to in_progress when already 3 in progress",
			task: &models.Task{
				ID:     "task3",
				Status: "in_progress",
			},
			requestingUserID: "user3",
			setupMocks: func(mockTaskDAO *mocks.TaskDAO) {
				existing := &models.Task{ID: "task3", UserID: "user3", Status: "todo"}
				mockTaskDAO.On("GetById", mock.Anything, "task3").Return(existing, nil)
				mockTaskDAO.On("CountByUserAndStatus", mock.Anything, "user3", "in_progress").Return(int64(3), nil)
				// Update should NOT be called
			},
			expectedErr: "cannot have more than 3 tasks in progress",
		},
		{
			name: "fail - task not found",
			task: &models.Task{ID: "nonexistent"},
			requestingUserID: "user4",
			setupMocks: func(mockTaskDAO *mocks.TaskDAO) {
				mockTaskDAO.On("GetById", mock.Anything, "nonexistent").Return(nil, nil)
			},
			expectedErr: "task not found",
		},
		{
			name: "fail - access denied (task belongs to another user)",
			task: &models.Task{ID: "task4"},
			requestingUserID: "hacker",
			setupMocks: func(mockTaskDAO *mocks.TaskDAO) {
				existing := &models.Task{ID: "task4", UserID: "owner"}
				mockTaskDAO.On("GetById", mock.Anything, "task4").Return(existing, nil)
			},
			expectedErr: "access denied",
		},
		{
			name: "fail - GetById database error",
			task: &models.Task{ID: "task5"},
			requestingUserID: "user5",
			setupMocks: func(mockTaskDAO *mocks.TaskDAO) {
				mockTaskDAO.On("GetById", mock.Anything, "task5").Return(nil, errors.New("db error"))
			},
			expectedErr: "failed to get existing task",
		},
		{
			name: "fail - CountByUserAndStatus database error",
			task: &models.Task{
				ID:     "task6",
				Status: "in_progress",
			},
			requestingUserID: "user6",
			setupMocks: func(mockTaskDAO *mocks.TaskDAO) {
				existing := &models.Task{ID: "task6", UserID: "user6", Status: "todo"}
				mockTaskDAO.On("GetById", mock.Anything, "task6").Return(existing, nil)
				mockTaskDAO.On("CountByUserAndStatus", mock.Anything, "user6", "in_progress").Return(int64(0), errors.New("count failed"))
			},
			expectedErr: "failed to count in-progress tasks",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockTaskDAO := new(mocks.TaskDAO)
			tt.setupMocks(mockTaskDAO)

			svc := NewTaskService(mockTaskDAO, nil, logger)
			err := svc.Update(context.Background(), tt.task, tt.requestingUserID)

			if tt.expectedErr != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr)
			} else {
				assert.NoError(t, err)
			}
			mockTaskDAO.AssertExpectations(t)
		})
	}
}

func TestTaskService_Delete(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.FatalLevel)

	tests := []struct {
		name            string
		taskID          string
		requestingUserID string
		setupMocks      func(mockTaskDAO *mocks.TaskDAO)
		expectedErr     string
	}{
		{
			name:            "success - user owns task",
			taskID:          "task1",
			requestingUserID: "user1",
			setupMocks: func(mockTaskDAO *mocks.TaskDAO) {
				mockTaskDAO.On("GetById", mock.Anything, "task1").Return(&models.Task{ID: "task1", UserID: "user1"}, nil)
				mockTaskDAO.On("Delete", mock.Anything, "task1").Return(nil)
			},
			expectedErr: "",
		},
		{
			name:            "fail - task not found",
			taskID:          "nonexistent",
			requestingUserID: "user1",
			setupMocks: func(mockTaskDAO *mocks.TaskDAO) {
				mockTaskDAO.On("GetById", mock.Anything, "nonexistent").Return(nil, nil)
			},
			expectedErr: "task not found",
		},
		{
			name:            "fail - access denied (task belongs to another user)",
			taskID:          "task2",
			requestingUserID: "hacker",
			setupMocks: func(mockTaskDAO *mocks.TaskDAO) {
				mockTaskDAO.On("GetById", mock.Anything, "task2").Return(&models.Task{ID: "task2", UserID: "owner"}, nil)
			},
			expectedErr: "access denied",
		},
		{
			name:            "fail - GetById database error",
			taskID:          "task3",
			requestingUserID: "user3",
			setupMocks: func(mockTaskDAO *mocks.TaskDAO) {
				mockTaskDAO.On("GetById", mock.Anything, "task3").Return(nil, errors.New("db error"))
			},
			expectedErr: "failed to get existing task",
		},
		{
			name:            "fail - Delete database error",
			taskID:          "task4",
			requestingUserID: "user4",
			setupMocks: func(mockTaskDAO *mocks.TaskDAO) {
				mockTaskDAO.On("GetById", mock.Anything, "task4").Return(&models.Task{ID: "task4", UserID: "user4"}, nil)
				mockTaskDAO.On("Delete", mock.Anything, "task4").Return(errors.New("delete failed"))
			},
			expectedErr: "failed to delete task",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockTaskDAO := new(mocks.TaskDAO)
			tt.setupMocks(mockTaskDAO)

			svc := NewTaskService(mockTaskDAO, nil, logger)
			err := svc.Delete(context.Background(), tt.taskID, tt.requestingUserID)

			if tt.expectedErr != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr)
			} else {
				assert.NoError(t, err)
			}
			mockTaskDAO.AssertExpectations(t)
		})
	}
}