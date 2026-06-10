package controller

import (
	"net/http"

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
    token, err := c.authService.Login(ctx.Request.Context(), req.Email, req.Password)
    if err != nil {
        ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
        return
    }
    ctx.JSON(http.StatusOK, gin.H{"token": token})
}

// Login godoc
// @Summary      User SignUp
// @Description  Creates a new user 
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body models.SignUpRequest true "signup credentials"
// @Success      200 {object} map[string]string "token"
// @Failure      400 {object} map[string]string "invalid request"
// @Failure      401 {object} map[string]string "invalid credentials"
// @Router       /auth/signup [post]
func (c *AuthController) SignUp(ctx *gin.Context) {
    var req models.LoginRequest
    if err := ctx.ShouldBindJSON(&req); err != nil {
        c.logger.WithError(err).Warn("invalid Signup request")
        ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
        return
    }
    token, err := c.authService.SignUp(ctx.Request.Context(), req.Email, req.Password)
    if err != nil {
        ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
        return
    }
    ctx.JSON(http.StatusOK, gin.H{"token": token})
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
