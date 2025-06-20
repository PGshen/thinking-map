/*
 * @Date: 2025-06-18 23:52:48
 * @LastEditors: peng pgs1108pgs@gmail.com
 * @LastEditTime: 2025-06-19 23:43:41
 * @FilePath: /thinking-map/server/internal/router/router.go
 */
package router

import (
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/thinking-map/server/internal/handler"
	"github.com/thinking-map/server/internal/middleware"
	"github.com/thinking-map/server/internal/pkg/sse"
	"github.com/thinking-map/server/internal/repository"
	"github.com/thinking-map/server/internal/service"
	"gorm.io/gorm"
)

// SetupRouter configures the router with all routes
func SetupRouter(
	db *gorm.DB,
	redisClient *redis.Client,
	jwtConfig service.JWTConfig,
) *gin.Engine {
	r := gin.Default()

	// Create repositories
	mapRepo := repository.NewMapRepository(db)
	nodeRepo := repository.NewThinkingNodeRepository(db)

	// Create services
	authService := service.NewAuthService(db, redisClient, jwtConfig)
	mapService := service.NewMapService(mapRepo)
	nodeService := service.NewNodeService(nodeRepo)

	// Create handlers
	authHandler := handler.NewAuthHandler(authService)
	mapHandler := handler.NewMapHandler(mapService)
	nodeHandler := handler.NewNodeHandler(nodeService)
	thinkingHandler := handler.NewThinkingHandler()
	eventManager := sse.NewEventManager()
	sseHandler := handler.NewSSEHandler(eventManager)

	// API v1 group
	v1 := r.Group("/api/v1")
	{
		// Auth routes (no auth required)
		auth := v1.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.POST("/refresh", authHandler.RefreshToken)
			auth.POST("/logout", authHandler.Logout)
		}

		// Protected routes (auth required)
		protected := v1.Group("")
		protected.Use(middleware.AuthMiddleware(authService))
		{
			// Map routes
			maps := protected.Group("/maps")
			{
				maps.POST("", mapHandler.CreateMap)
				maps.GET("", mapHandler.ListMaps)
				maps.GET("/:mapId", mapHandler.GetMap)
				maps.PUT("/:mapId", mapHandler.UpdateMap)
				maps.DELETE("/:mapId", mapHandler.DeleteMap)
			}

			// Node routes
			nodes := protected.Group("/maps/:mapId/nodes")
			{
				nodes.GET("", nodeHandler.ListNodes)
				nodes.POST("", nodeHandler.CreateNode)
				nodes.PUT("/:nodeId", nodeHandler.UpdateNode)
				nodes.DELETE("/:nodeId", nodeHandler.DeleteNode)
				nodes.GET("/:nodeId/dependencies", nodeHandler.GetDependencies)
			}

			// Thinking routes
			thinking := protected.Group("/thinking")
			{
				thinking.POST("/analyze", thinkingHandler.Analyze)
				thinking.POST("/decompose", thinkingHandler.Decompose)
				thinking.POST("/conclude", thinkingHandler.Conclude)
				thinking.POST("/chat", thinkingHandler.Chat)
			}

			// SSE routes
			sse := protected.Group("/sse")
			{
				sse.GET("/connect/:mapId", sseHandler.Connect)
			}
		}
	}

	return r
}
