package models

import "time"

type TaskStatus string

const (
	StatusTodo TaskStatus = "TODO"
	StatusInprogess TaskStatus = "In Progress"
	StatusDone TaskStatus = "Done"
)


type Task struct {
    ID          string     `json:"id" bson:"_id,omitempty"`
    UserID      string     `json:"user_id" bson:"user_id"`
    Title       string     `json:"title" bson:"title"`
    Description string     `json:"description" bson:"description"`
    Status      TaskStatus `json:"status" bson:"status"`
    CreatedAt   time.Time  `json:"created_at" bson:"created_at"`
    UpdatedAt   time.Time  `json:"updated_at" bson:"updated_at"`
}

type CreateTaskRequest struct {
    Title       string     `json:"title" binding:"required"`
    Description string     `json:"description"`
    Status      TaskStatus `json:"status" binding:"omitempty,oneof=TODO IN_PROGRESS DONE"`
}

type UpdateTaskRequest struct {
    Title       *string     `json:"title"`
    Description *string     `json:"description"`
    Status      *TaskStatus `json:"status" binding:"omitempty,oneof=TODO IN_PROGRESS DONE"`
}

type TaskFilter struct {
    Status   string `form:"status"`
    Search   string `form:"search"`   // matches title or description
    Page     int    `form:"page"`
    PageSize int    `form:"page_size"`
    SortBy   string `form:"sort_by"`  // title | created_at | updated_at | status
    SortDir  string `form:"sort_dir"` // asc | desc
}