package service

import (
    "context"
    "errors"
    "testing"

    "github.com/Perlishnov/todoapp/internal/models"
    "github.com/Perlishnov/todoapp/mocks"
    "github.com/sirupsen/logrus"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
)

func TestTaskService_CreateTask(t *testing.T) {
    logger := logrus.New()
    logger.SetLevel(logrus.FatalLevel)

    tests := []struct {
        name           string
        userID         string
        req            *models.CreateTaskRequest
        setupMocks     func(mockTaskDAO *mocks.TaskDAO, mockUserDAO *mocks.UserDAO)
        expectedTask   *models.Task
        expectedErr    string
    }{
        {
            name:   "success - create task with TODO status",
            userID: "user123",
            req: &models.CreateTaskRequest{
                Title:       "My first task",
                Description: "Learn Go",
                Status:      "TODO",
            },
            setupMocks: func(mockTaskDAO *mocks.TaskDAO, mockUserDAO *mocks.UserDAO) {
                // No need to check in-progress count for non-IN_PROGRESS tasks.
                mockTaskDAO.On("CountByUserAndStatus", mock.Anything, "user123", "IN_PROGRESS").Return(0, nil)
                mockTaskDAO.On("Create", mock.Anything, mock.AnythingOfType("*models.Task")).Return(nil).Run(func(args mock.Arguments) {
                    task := args.Get(1).(*models.Task)
                    task.ID = "generated-id-123"
                })
            },
            expectedTask: &models.Task{
                ID:          "generated-id-123",
                Title:       "My first task",
                Description: "Learn Go",
                Status:      "TODO",
            },
            expectedErr: "",
        },
        {
            name:   "fail - user already has 3 tasks IN_PROGRESS",
            userID: "user456",
            req: &models.CreateTaskRequest{
                Title:  "Fourth task",
                Status: "IN_PROGRESS",
            },
            setupMocks: func(mockTaskDAO *mocks.TaskDAO, mockUserDAO *mocks.UserDAO) {
                mockTaskDAO.On("CountByUserAndStatus", mock.Anything, "user456", "IN_PROGRESS").Return(3, nil)
                // Create should NOT be called.
            },
            expectedErr: "user cannot have more than 3 tasks in progress",
        },
        {
            name:   "fail - invalid status",
            userID: "user789",
            req: &models.CreateTaskRequest{
                Title:  "Invalid",
                Status: "UNKNOWN",
            },
            setupMocks:   func(mockTaskDAO *mocks.TaskDAO, mockUserDAO *mocks.UserDAO) {},
            expectedErr:  "invalid status: must be TODO, IN_PROGRESS, or DONE",
        },
        {
            name:   "fail - DAO Count error",
            userID: "userErr",
            req: &models.CreateTaskRequest{
                Title:  "Task",
                Status: "IN_PROGRESS",
            },
            setupMocks: func(mockTaskDAO *mocks.TaskDAO, mockUserDAO *mocks.UserDAO) {
                mockTaskDAO.On("CountByUserAndStatus", mock.Anything, "userErr", "IN_PROGRESS").Return(0, errors.New("db connection failed"))
            },
            expectedErr: "failed to check in-progress count",
        },
        {
            name:   "fail - DAO Create error",
            userID: "userCreateErr",
            req: &models.CreateTaskRequest{
                Title:  "Will fail",
                Status: "TODO",
            },
            setupMocks: func(mockTaskDAO *mocks.TaskDAO, mockUserDAO *mocks.UserDAO) {
                mockTaskDAO.On("CountByUserAndStatus", mock.Anything, "userCreateErr", "IN_PROGRESS").Return(0, nil)
                mockTaskDAO.On("Create", mock.Anything, mock.AnythingOfType("*models.Task")).Return(errors.New("duplicate key"))
            },
            expectedErr: "failed to create task",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            mockTaskDAO := new(mocks.TaskDAO)
            mockUserDAO := new(mocks.UserDAO)
            tt.setupMocks(mockTaskDAO, mockUserDAO)

            svc := NewTaskService(mockTaskDAO, mockUserDAO, logger)
            task, err := svc.CreateTask(context.Background(), tt.userID, tt.req)

            if tt.expectedErr != "" {
                assert.Error(t, err)
                assert.Contains(t, err.Error(), tt.expectedErr)
                assert.Nil(t, task)
            } else {
                assert.NoError(t, err)
                assert.NotNil(t, task)
                assert.Equal(t, tt.expectedTask.ID, task.ID)
                assert.Equal(t, tt.expectedTask.Title, task.Title)
                assert.Equal(t, tt.expectedTask.Status, task.Status)
            }
            mockTaskDAO.AssertExpectations(t)
            mockUserDAO.AssertExpectations(t)
        })
    }
}

func TestTaskService_UpdateTask_StatusChangeRespectsMaxLimit(t *testing.T) {
    logger := logrus.New()
    logger.SetLevel(logrus.FatalLevel)

    tests := []struct {
        name           string
        userID         string
        taskID         string
        req            *models.UpdateTaskRequest
        existingTask   *models.Task
        setupMocks     func(mockTaskDAO *mocks.TaskDAO)
        expectedErr    string
    }{
        {
            name:   "success - changing from TODO to IN_PROGRESS when limit not reached",
            userID: "user1",
            taskID: "task1",
            req: &models.UpdateTaskRequest{
                Status: ptrString("IN_PROGRESS"),
            },
            existingTask: &models.Task{ID: "task1", UserID: "user1", Status: "TODO"},
            setupMocks: func(mockTaskDAO *mocks.TaskDAO) {
                mockTaskDAO.On("GetByID", mock.Anything, "task1").Return(&models.Task{ID: "task1", UserID: "user1", Status: "TODO"}, nil)
                // Count excludes current task because its status is not IN_PROGRESS yet.
                mockTaskDAO.On("CountByUserAndStatus", mock.Anything, "user1", "IN_PROGRESS").Return(2, nil)
                mockTaskDAO.On("Update", mock.Anything, mock.AnythingOfType("*models.Task")).Return(nil)
            },
            expectedErr: "",
        },
        {
            name:   "fail - changing to IN_PROGRESS when user already has 3 including this? Actually this task is not yet IN_PROGRESS, so check should see 3 others -> fail",
            userID: "user2",
            taskID: "task2",
            req: &models.UpdateTaskRequest{
                Status: ptrString("IN_PROGRESS"),
            },
            existingTask: &models.Task{ID: "task2", UserID: "user2", Status: "TODO"},
            setupMocks: func(mockTaskDAO *mocks.TaskDAO) {
                mockTaskDAO.On("GetByID", mock.Anything, "task2").Return(&models.Task{ID: "task2", UserID: "user2", Status: "TODO"}, nil)
                mockTaskDAO.On("CountByUserAndStatus", mock.Anything, "user2", "IN_PROGRESS").Return(3, nil)
            },
            expectedErr: "user cannot have more than 3 tasks in progress",
        },
        {
            name:   "success - updating from IN_PROGRESS to DONE (no limit check needed)",
            userID: "user3",
            taskID: "task3",
            req: &models.UpdateTaskRequest{
                Status: ptrString("DONE"),
            },
            existingTask: &models.Task{ID: "task3", UserID: "user3", Status: "IN_PROGRESS"},
            setupMocks: func(mockTaskDAO *mocks.TaskDAO) {
                mockTaskDAO.On("GetByID", mock.Anything, "task3").Return(&models.Task{ID: "task3", UserID: "user3", Status: "IN_PROGRESS"}, nil)
                // No count check required because new status is not IN_PROGRESS.
                mockTaskDAO.On("Update", mock.Anything, mock.AnythingOfType("*models.Task")).Return(nil)
            },
            expectedErr: "",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            mockTaskDAO := new(mocks.TaskDAO)
            mockUserDAO := new(mocks.UserDAO)
            tt.setupMocks(mockTaskDAO)

            svc := NewTaskService(mockTaskDAO, mockUserDAO, logger)
            task, err := svc.UpdateTask(context.Background(), tt.taskID, tt.userID, tt.req)

            if tt.expectedErr != "" {
                assert.Error(t, err)
                assert.Contains(t, err.Error(), tt.expectedErr)
                assert.Nil(t, task)
            } else {
                assert.NoError(t, err)
                assert.NotNil(t, task)
            }
            mockTaskDAO.AssertExpectations(t)
        })
    }
}

func TestTaskService_ListTasks(t *testing.T) {
    logger := logrus.New()
    logger.SetLevel(logrus.FatalLevel)

    tests := []struct {
        name         string
        userID       string
        filter       models.TaskFilter
        page, limit  int
        sort         string
        setupMocks   func(mockTaskDAO *mocks.TaskDAO)
        expectedLen  int
        expectedTotal int64
        expectedErr  string
    }{
        {
            name:   "success - no filters, default pagination",
            userID: "user1",
            filter: models.TaskFilter{},
            page:   1, limit: 10,
            sort: "created_at_desc",
            setupMocks: func(mockTaskDAO *mocks.TaskDAO) {
                tasks := []models.Task{
                    {ID: "t1", Title: "Task 1"},
                    {ID: "t2", Title: "Task 2"},
                }
                mockTaskDAO.On("List", mock.Anything, mock.Anything, 10, 0, "created_at_desc").
                    Return(tasks, int64(2), nil)
            },
            expectedLen:  2,
            expectedTotal: 2,
        },
        {
            name:   "success - filter by status",
            userID: "user2",
            filter: models.TaskFilter{Status: "TODO"},
            page:   1, limit: 5,
            sort: "title_asc",
            setupMocks: func(mockTaskDAO *mocks.TaskDAO) {
                tasks := []models.Task{{ID: "t3", Title: "Alpha", Status: "TODO"}}
                mockTaskDAO.On("List", mock.Anything, models.TaskFilter{Status: "TODO"}, 5, 0, "title_asc").
                    Return(tasks, int64(1), nil)
            },
            expectedLen:  1,
            expectedTotal: 1,
        },
        {
            name:   "fail - DAO error",
            userID: "userErr",
            filter: models.TaskFilter{},
            page:   1, limit: 10,
            sort: "created_at_desc",
            setupMocks: func(mockTaskDAO *mocks.TaskDAO) {
                mockTaskDAO.On("List", mock.Anything, mock.Anything, 10, 0, "created_at_desc").
                    Return(nil, int64(0), errors.New("db error"))
            },
            expectedErr: "failed to list tasks",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            mockTaskDAO := new(mocks.TaskDAO)
            mockUserDAO := new(mocks.UserDAO)
            tt.setupMocks(mockTaskDAO)

            svc := NewTaskService(mockTaskDAO, mockUserDAO, logger)
            tasks, total, err := svc.ListTasks(context.Background(), tt.userID, tt.filter, tt.page, tt.limit, tt.sort)

            if tt.expectedErr != "" {
                assert.Error(t, err)
                assert.Contains(t, err.Error(), tt.expectedErr)
                assert.Nil(t, tasks)
            } else {
                assert.NoError(t, err)
                assert.Len(t, tasks, tt.expectedLen)
                assert.Equal(t, tt.expectedTotal, total)
            }
            mockTaskDAO.AssertExpectations(t)
        })
    }
}

func TestTaskService_DeleteTask(t *testing.T) {
    logger := logrus.New()
    logger.SetLevel(logrus.FatalLevel)

    tests := []struct {
        name         string
        taskID       string
        userID       string
        setupMocks   func(mockTaskDAO *mocks.TaskDAO)
        expectedErr  string
    }{
        {
            name:   "success - user owns task",
            taskID: "task1",
            userID: "user1",
            setupMocks: func(mockTaskDAO *mocks.TaskDAO) {
                mockTaskDAO.On("GetByID", mock.Anything, "task1").Return(&models.Task{ID: "task1", UserID: "user1"}, nil)
                mockTaskDAO.On("Delete", mock.Anything, "task1").Return(nil)
            },
            expectedErr: "",
        },
        {
            name:   "fail - task not found",
            taskID: "nonexistent",
            userID: "user1",
            setupMocks: func(mockTaskDAO *mocks.TaskDAO) {
                mockTaskDAO.On("GetByID", mock.Anything, "nonexistent").Return(nil, nil)
            },
            expectedErr: "task not found",
        },
        {
            name:   "fail - user does not own task",
            taskID: "task2",
            userID: "wrongUser",
            setupMocks: func(mockTaskDAO *mocks.TaskDAO) {
                mockTaskDAO.On("GetByID", mock.Anything, "task2").Return(&models.Task{ID: "task2", UserID: "owner"}, nil)
            },
            expectedErr: "access denied",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            mockTaskDAO := new(mocks.TaskDAO)
            mockUserDAO := new(mocks.UserDAO)
            tt.setupMocks(mockTaskDAO)

            svc := NewTaskService(mockTaskDAO, mockUserDAO, logger)
            err := svc.DeleteTask(context.Background(), tt.taskID, tt.userID)

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

// Helper to get string pointer for UpdateTaskRequest.
func ptrString(s string) *string {
    return &s
}