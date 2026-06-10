package controller

import (
	"github.com/Perlishnov/TODO_APP/internal/middleware"
	"github.com/gin-gonic/gin"
)

type TaskController struct{

}

func (ctrl *TaskController) RegisterRoutes (rg *gin.RouterGroup){
	tasks := rg.Group("/tasks")
	tasks.Use(middleware.Auth())
	{
		tasks.POST("", ctrl.)
	}
}