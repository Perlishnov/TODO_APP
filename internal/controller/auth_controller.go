package controller

import (
	"net/http"
	"strings"

	"github.com/Perlishnov/TODO_APP/internal/models"
	"github.com/Perlishnov/TODO_APP/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type AuthController struct {
	authService service.AuthService
	logger      *logrus.Logger
}

func NewAuthController(authService service.AuthService, logger *logrus.Logger) *AuthController {
	return &AuthController{authService: authService, logger: logger}
}

func (ctrl *AuthController) RegisterRoutes(rg *gin.RouterGroup) {
	authRoutes := rg.Group("/auth")
	{
		authRoutes.POST("/signup", ctrl.SignUp)
		authRoutes.POST("login", ctrl.Login)
		authRoutes.POST("/logout", ctrl.Logout)
	}
}

// Login godoc
// @Summary      User login
// @Description  Authenticates a user and returns a JWT token.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body models.LoginRequest true "Login credentials"
// @Success      200 {object} map[string]string "token"
// @Failure      400 {object} map[string]string "invalid request"
// @Failure      401 {object} map[string]string "invalid credentials"
// @Router       /auth/login [post]
func (c *AuthController) Login(ctx *gin.Context) {
	var req models.LoginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.logger.WithError(err).Warn("invalid login request")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	token, err := c.authService.Login(ctx.Request.Context(), &req)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"token": token})
}

// Signup godoc
// @Summary      User SignUp
// @Description  Creates a new user and returns a JWT token (auto-login)
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body models.CreateUserRequest true "signup credentials"
// @Success      201 {object} map[string]string "token"
// @Failure      400 {object} map[string]string "invalid request body or validation error"
// @Failure      409 {object} map[string]string "email already exists"
// @Failure      500 {object} map[string]string "internal server error"
// @Router       /auth/signup [post]
func (c *AuthController) SignUp(ctx *gin.Context) {
	var req models.CreateUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.logger.WithError(err).Warn("invalid Signup request")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	err := c.authService.Signup(ctx, &req)
	if err != nil {
		c.logger.WithError(err).WithField("email", req.Email).Error("signup faieled")

		if strings.Contains(err.Error(), "already exists") {
			ctx.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}

		if strings.Contains(err.Error(), "invalid role") || strings.Contains(err.Error(), "password") {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})

		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"message": "Signed up sucessfully"})
}

// Logout godoc
// @Summary      User logout
// @Description  Logs out the user (client must discard token).
// @Tags         auth
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} map[string]string "message"
// @Failure      500 {object} map[string]string "internal error"
// @Router       /auth/logout [post]
func (c *AuthController) Logout(ctx *gin.Context) {
	token := ctx.GetHeader("Authorization")
	if err := c.authService.Logout(token); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "logged out"})
}
