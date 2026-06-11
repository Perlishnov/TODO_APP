package middleware

import (
    "net/http"
    "strings"

    "github.com/Perlishnov/TODO_APP/internal/utils"
    "github.com/gin-gonic/gin"
    "github.com/sirupsen/logrus"
)

type AuthMiddleware struct {
    jwtUtil utils.JWTService
    logger  *logrus.Logger
}

func NewAuthMiddleware(jwtUtil utils.JWTService, logger *logrus.Logger) *AuthMiddleware {
    return &AuthMiddleware{
        jwtUtil: jwtUtil,
        logger:  logger,
    }
}

// Authenticate returns a Gin middleware that validates the JWT token
// and stores the user ID in the request context.
func (m *AuthMiddleware) Authenticate() gin.HandlerFunc {
    return func(c *gin.Context) {
        // 1. Extract Authorization header
        authHeader := c.GetHeader("Authorization")
        if authHeader == "" {
            m.logger.Warn("missing Authorization header")
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
                "error": "authorization header required",
            })
            return
        }

        // 2. Expect "Bearer <token>"
        parts := strings.SplitN(authHeader, " ", 2)
        if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
            m.logger.Warn("invalid Authorization header format")
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
                "error": "invalid authorization header format",
            })
            return
        }
        tokenString := parts[1]

        // 3. Validate token and extract claims
        claims, err := m.jwtUtil.ValidateToken(tokenString)
        if err != nil {
            m.logger.WithError(err).Warn("token validation failed")
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
                "error": "invalid or expired token",
            })
            return
        }

        // 4. Safely store user ID in context
        //    The claims struct is expected to have a field UserID (string)
        if claims.UserID == "" {
            m.logger.Error("token claims missing user ID")
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
                "error": "invalid token claims",
            })
            return
        }

        c.Set("userID", claims.UserID)
        // Optionally store other claims (email, role) if needed
        if claims.Email != "" {
            c.Set("userEmail", claims.Email)
        }
        if claims.Role != "" {
            c.Set("userRole", claims.Role)
        } else {
            // Default role for TODO app (no admin)
            c.Set("userRole", "user")
        }

        m.logger.WithField("user_id", claims.UserID).Debug("authenticated request")
        c.Next()
    }
}