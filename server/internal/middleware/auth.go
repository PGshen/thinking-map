/*
 * @Date: 2025-06-18 22:56:20
 * @LastEditors: peng pgs1108pgs@gmail.com
 * @LastEditTime: 2025-06-19 00:23:06
 * @FilePath: /thinking-map/server/internal/middleware/auth.go
 */
package middleware

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/google/uuid"
	"github.com/thinking-map/server/internal/model/dto"
	"github.com/thinking-map/server/internal/service"
)

// AuthMiddleware creates a new authentication middleware
func AuthMiddleware(authService service.AuthService) app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		// Get token from Authorization header
		authHeader := string(c.GetHeader("Authorization"))
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, dto.Response{
				Code:      http.StatusUnauthorized,
				Message:   "missing authorization header",
				Data:      nil,
				Timestamp: time.Now(),
				RequestID: uuid.New().String(),
			})
			c.Abort()
			return
		}

		// Check if token is in Bearer format
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, dto.Response{
				Code:      http.StatusUnauthorized,
				Message:   "invalid authorization header format",
				Data:      nil,
				Timestamp: time.Now(),
				RequestID: uuid.New().String(),
			})
			c.Abort()
			return
		}

		// Validate token
		tokenInfo, err := authService.ValidateToken(ctx, parts[1])
		if err != nil {
			c.JSON(http.StatusUnauthorized, dto.Response{
				Code:      http.StatusUnauthorized,
				Message:   "invalid token",
				Data:      dto.ErrorData{Error: err.Error()},
				Timestamp: time.Now(),
				RequestID: uuid.New().String(),
			})
			c.Abort()
			return
		}

		// Set user info in context
		c.Set("user_id", tokenInfo.UserID)
		c.Set("username", tokenInfo.Username)

		c.Next(ctx)
	}
}
