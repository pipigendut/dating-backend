package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	base "github.com/pipigendut/dating-backend/internal/delivery/http/dto"
	"github.com/pipigendut/dating-backend/pkg/auth"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			base.Error(c, http.StatusUnauthorized, "Authorization header required", nil)
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			base.Error(c, http.StatusUnauthorized, "Invalid authorization header format", nil)
			c.Abort()
			return
		}

		claims, err := auth.ValidateToken(parts[1])
		if err != nil {
			base.Error(c, http.StatusUnauthorized, "Invalid or expired token", err.Error())
			c.Abort()
			return
		}

		// Set userID in context
		c.Set("userID", claims.UserID)
		c.Next()
	}
}

func BasicAuthMiddleware(username, password string) gin.HandlerFunc {
	return gin.BasicAuth(gin.Accounts{
		username: password,
	})
}
