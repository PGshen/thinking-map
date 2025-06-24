/*
 * @Date: 2025-06-23 23:06:06
 * @LastEditors: peng pgs1108pgs@gmail.com
 * @LastEditTime: 2025-06-23 23:16:31
 * @FilePath: /thinking-map/server/internal/middleware/node_ownership.go
 */
package middleware

import (
	"net/http"

	"github.com/PGshen/thinking-map/server/internal/repository"

	"github.com/gin-gonic/gin"
)

// NodeOwnershipMiddleware checks if the node belongs to the user
func NodeOwnershipMiddleware(nodeRepo repository.ThinkingNode, mapRepo repository.ThinkingMap) gin.HandlerFunc {
	return func(c *gin.Context) {
		nodeID := c.Param("nodeId")
		if nodeID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"code": http.StatusBadRequest, "message": "node ID is required"})
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

		node, err := nodeRepo.FindByID(c.Request.Context(), nodeID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"code": http.StatusNotFound, "message": "node not found"})
			c.Abort()
			return
		}
		mapObj, err := mapRepo.FindByID(c.Request.Context(), node.MapID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"code": http.StatusNotFound, "message": "map not found"})
			c.Abort()
			return
		}
		if mapObj.UserID != userID {
			c.JSON(http.StatusForbidden, gin.H{"code": http.StatusForbidden, "message": "forbidden: node does not belong to user"})
			c.Abort()
			return
		}

		c.Next()
	}
}
