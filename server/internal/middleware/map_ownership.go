package middleware

import (
	"net/http"

	"github.com/PGshen/thinking-map/server/internal/repository"

	"github.com/gin-gonic/gin"
)

// MapOwnershipMiddleware checks if the map belongs to the user
func MapOwnershipMiddleware(mapRepo repository.ThinkingMap) gin.HandlerFunc {
	return func(c *gin.Context) {
		mapID := c.Param("mapId")
		if mapID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"code": http.StatusBadRequest, "message": "map ID is required"})
			c.Abort()
			return
		}

		userIDValue, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"code": http.StatusUnauthorized, "message": "unauthorized"})
			c.Abort()
			return
		}
		userID, ok := userIDValue.(string)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"code": http.StatusInternalServerError, "message": "invalid user id"})
			c.Abort()
			return
		}

		mapObj, err := mapRepo.FindByID(c.Request.Context(), mapID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"code": http.StatusNotFound, "message": "map not found"})
			c.Abort()
			return
		}
		if mapObj.UserID != userID {
			c.JSON(http.StatusForbidden, gin.H{"code": http.StatusForbidden, "message": "forbidden: map does not belong to user"})
			c.Abort()
			return
		}

		c.Next()
	}
}
