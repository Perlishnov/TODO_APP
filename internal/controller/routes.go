package controller

import "github.com/gin-gonic/gin"

type RegisterRoutes interface{
	RegisterRoutes(rg * gin.RouterGroup)
}