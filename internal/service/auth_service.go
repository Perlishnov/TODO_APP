package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/Perlishnov/TODO_APP/internal/dao"
	"github.com/Perlishnov/TODO_APP/internal/models"
	User "github.com/Perlishnov/TODO_APP/internal/models"
	"github.com/Perlishnov/TODO_APP/internal/utils"
	"github.com/sirupsen/logrus"
)

type AuthService interface {
	Signup(ctx context.Context, request *User.CreateUserRequest) error
	Login(ctx context.Context, request *User.LoginRequest) (string, error)
	Logout(token string) error
}

type authService struct {
	userDAO dao.UserDAO
	jwtUtil utils.JWTService
	logger  *logrus.Logger
}

func NewAuthService(userDAO dao.UserDAO, jwtUtil utils.JWTService, logger *logrus.Logger) AuthService {
	return &authService{
		userDAO: userDAO,
		jwtUtil: jwtUtil,
		logger:  logger,
	}
}

func (s *authService) Signup(ctx context.Context, request *models.CreateUserRequest) error {
	
	if request.Role != "" && request.Role != "user" && request.Role != "admin" {
        return errors.New("role must be either 'user' or 'admin'")
    }
	
	existing, err := s.userDAO.GetByEmail(ctx, request.Email)

	if err != nil {
		s.logger.WithError(err).Error("failed to check existing user")
		return fmt.Errorf("database error: %w", err)
	}
	if existing != nil {
		return errors.New("user already exists")
	}
	hashedPassword, err := utils.HashPassword(request.Password)
	if err != nil {
		s.logger.WithError(err).Error("password hashing failed")
		return fmt.Errorf("failed to hash password: %w", err)
	}
	role := request.Role
	if role == "" {
		role = "user"
	}
	user := &models.User{
		Name:     request.Name,
		Email:    request.Email,
		Password: hashedPassword,
		Role:     role,
	}
	if err := s.userDAO.Create(ctx, user); err != nil {
		s.logger.WithError(err).Error("failed to create user")
		return fmt.Errorf("failed to create user: %w", err)
	}

	s.logger.WithField("user_id", user.ID).Info("user signed up")
	return nil
}

func (s *authService) Login(ctx context.Context, request *User.LoginRequest) (string, error) {
	user, err := s.userDAO.GetByEmail(ctx, request.Email)
	if err != nil {
		s.logger.WithError(err).Error("database error during login")
		return "", fmt.Errorf("database error: %w", err)
	}
	if user == nil || !utils.CheckPasswordHash(request.Password, user.Password) {
		s.logger.WithField("email", request.Email).Warn("failed login attempt")
		return "", errors.New("invalid credentials")
	}
	token, err := s.jwtUtil.GenerateToken(*user)
	if err != nil {
		s.logger.WithError(err).Error("token generation failed")
		return "", fmt.Errorf("failed to generate token %w", err)
	}
	s.logger.WithField("user_id", user.ID).Info("user logged in")
	return token, nil
}

func (s *authService) Logout(token string) error {
	s.logger.Info("logout called")
	return nil
}
