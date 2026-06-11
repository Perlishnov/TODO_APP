package service

import (
    "context"
    "errors"
    "testing"

    "github.com/Perlishnov/TODO_APP/internal/models"
    "github.com/Perlishnov/TODO_APP/internal/utils"
    "github.com/Perlishnov/TODO_APP/mocks"
    "github.com/sirupsen/logrus"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
)

func TestAuthService_Signup(t *testing.T) {
    logger := logrus.New()
    logger.SetLevel(logrus.FatalLevel)

    type args struct {
        req *models.CreateUserRequest
    }
    tests := []struct {
        name      string
        setupMock func(mockDAO *mocks.UserDAO, mockJWT *mocks.JWTService)
        args      args
        wantErr   bool
        errMsg    string
    }{
        {
            name: "success",
            setupMock: func(mockDAO *mocks.UserDAO, mockJWT *mocks.JWTService) {
                mockDAO.On("GetByEmail", mock.Anything, "test@example.com").
                    Return(nil, nil)
                mockDAO.On("Create", mock.Anything, mock.AnythingOfType("*models.User")).
                    Return(nil).
                    Run(func(args mock.Arguments) {
                        user := args.Get(1).(*models.User)
                        user.ID = "new-uuid"
                    })
            },
            args: args{
                req: &models.CreateUserRequest{
                    Name:     "Test User",
                    Email:    "test@example.com",
                    Password: "validpass",
                    Role:     "user",
                },
            },
            wantErr: false,
        },
        {
            name: "fail - duplicate email",
            setupMock: func(mockDAO *mocks.UserDAO, mockJWT *mocks.JWTService) {
                mockDAO.On("GetByEmail", mock.Anything, "dup@example.com").
                    Return(&models.User{Email: "dup@example.com"}, nil)
            },
            args: args{
                req: &models.CreateUserRequest{
                    Email: "dup@example.com",
                },
            },
            wantErr: true,
            errMsg:  "already exists",
        },
        {
            name: "fail - invalid role",
            setupMock: func(mockDAO *mocks.UserDAO, mockJWT *mocks.JWTService) {
                // No GetByEmail expectation – service should return before calling it
                // No Create expectation
            },
            args: args{
                req: &models.CreateUserRequest{
                    Email: "badrole@example.com",
                    Role:  "superuser",
                },
            },
            wantErr: true,
            errMsg:  "role must be either 'user' or 'admin'",
        },
        {
            name: "fail - GetByEmail database error",
            setupMock: func(mockDAO *mocks.UserDAO, mockJWT *mocks.JWTService) {
                mockDAO.On("GetByEmail", mock.Anything, "db@error.com").
                    Return(nil, errors.New("connection refused"))
            },
            args: args{
                req: &models.CreateUserRequest{
                    Email: "db@error.com",
                },
            },
            wantErr: true,
            errMsg:  "database error",
        },
        {
            name: "fail - Create returns error",
            setupMock: func(mockDAO *mocks.UserDAO, mockJWT *mocks.JWTService) {
                mockDAO.On("GetByEmail", mock.Anything, "createfail@example.com").
                    Return(nil, nil)
                mockDAO.On("Create", mock.Anything, mock.AnythingOfType("*models.User")).
                    Return(errors.New("db write failed"))
            },
            args: args{
                req: &models.CreateUserRequest{
                    Email: "createfail@example.com",
                },
            },
            wantErr: true,
            errMsg:  "failed to create user",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            mockDAO := new(mocks.UserDAO)
            mockJWT := new(mocks.JWTService)
            if tt.setupMock != nil {
                tt.setupMock(mockDAO, mockJWT)
            }

            svc := NewAuthService(mockDAO, mockJWT, logger)
            err := svc.Signup(context.Background(), tt.args.req)

            if tt.wantErr {
                assert.Error(t, err)
                if tt.errMsg != "" {
                    assert.Contains(t, err.Error(), tt.errMsg)
                }
            } else {
                assert.NoError(t, err)
            }
            mockDAO.AssertExpectations(t)
        })
    }
}

func TestAuthService_Login(t *testing.T) {
    logger := logrus.New()
    logger.SetLevel(logrus.FatalLevel)

    hashedPass, _ := utils.HashPassword("correctpass")

    type args struct {
        req *models.LoginRequest
    }
    tests := []struct {
        name      string
        setupMock func(mockDAO *mocks.UserDAO, mockJWT *mocks.JWTService)
        args      args
        wantToken string
        wantErr   bool
        errMsg    string
    }{
        {
            name: "success",
            setupMock: func(mockDAO *mocks.UserDAO, mockJWT *mocks.JWTService) {
                user := &models.User{
                    ID:       "user123",
                    Email:    "john@example.com",
                    Password: hashedPass,
                    Role:     "user",
                }
                mockDAO.On("GetByEmail", mock.Anything, "john@example.com").
                    Return(user, nil)
                // Use mock.Anything to avoid field‑by‑field comparison
                mockJWT.On("GenerateToken", mock.AnythingOfType("models.User")).Return("jwt-token-xyz", nil)
            },
            args: args{
                req: &models.LoginRequest{
                    Email:    "john@example.com",
                    Password: "correctpass",
                },
            },
            wantToken: "jwt-token-xyz",
            wantErr:   false,
        },
        {
            name: "fail - user not found",
            setupMock: func(mockDAO *mocks.UserDAO, mockJWT *mocks.JWTService) {
                mockDAO.On("GetByEmail", mock.Anything, "unknown@example.com").
                    Return(nil, nil)
            },
            args: args{
                req: &models.LoginRequest{
                    Email:    "unknown@example.com",
                    Password: "anything",
                },
            },
            wantErr: true,
            errMsg:  "invalid credentials",
        },
        {
            name: "fail - wrong password",
            setupMock: func(mockDAO *mocks.UserDAO, mockJWT *mocks.JWTService) {
                user := &models.User{
                    Email:    "john@example.com",
                    Password: hashedPass,
                }
                mockDAO.On("GetByEmail", mock.Anything, "john@example.com").
                    Return(user, nil)
            },
            args: args{
                req: &models.LoginRequest{
                    Email:    "john@example.com",
                    Password: "wrongpass",
                },
            },
            wantErr: true,
            errMsg:  "invalid credentials",
        },
        {
            name: "fail - GetByEmail database error",
            setupMock: func(mockDAO *mocks.UserDAO, mockJWT *mocks.JWTService) {
                mockDAO.On("GetByEmail", mock.Anything, "error@example.com").
                    Return(nil, errors.New("db down"))
            },
            args: args{
                req: &models.LoginRequest{
                    Email: "error@example.com",
                },
            },
            wantErr: true,
            errMsg:  "database error",
        },
        {
            name: "fail - token generation error",
            setupMock: func(mockDAO *mocks.UserDAO, mockJWT *mocks.JWTService) {
                user := &models.User{
                    ID:    "uid",
                    Email: "tokenfail@example.com",
                    Role:  "user",
                    Password: hashedPass,
                }
                mockDAO.On("GetByEmail", mock.Anything, "tokenfail@example.com").
                    Return(user, nil)
                mockJWT.On("GenerateToken", mock.AnythingOfType("models.User")).Return("", errors.New("signing failed"))
            },
            args: args{
                req: &models.LoginRequest{
                    Email:    "tokenfail@example.com",
                    Password: "correctpass",
                },
            },
            wantErr: true,
            errMsg:  "failed to generate token",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            mockDAO := new(mocks.UserDAO)
            mockJWT := new(mocks.JWTService)
            if tt.setupMock != nil {
                tt.setupMock(mockDAO, mockJWT)
            }

            svc := NewAuthService(mockDAO, mockJWT, logger)
            token, err := svc.Login(context.Background(), tt.args.req)

            if tt.wantErr {
                assert.Error(t, err)
                if tt.errMsg != "" {
                    assert.Contains(t, err.Error(), tt.errMsg)
                }
                assert.Empty(t, token)
            } else {
                assert.NoError(t, err)
                assert.Equal(t, tt.wantToken, token)
            }
            mockDAO.AssertExpectations(t)
            mockJWT.AssertExpectations(t)
        })
    }
}

func TestAuthService_Logout(t *testing.T) {
    logger := logrus.New()
    logger.SetLevel(logrus.FatalLevel)
    svc := NewAuthService(nil, nil, logger)
    err := svc.Logout("some-token")
    assert.NoError(t, err)
}