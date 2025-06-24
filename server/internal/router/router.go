/*
 * @Date: 2025-06-18 23:52:48
 * @LastEditors: peng pgs1108pgs@gmail.com
 * @LastEditTime: 2025-06-23 23:18:01
 * @FilePath: /thinking-map/server/internal/router/router.go
 */
package router

import (
	"github.com/PGshen/thinking-map/server/internal/handler"
	"github.com/PGshen/thinking-map/server/internal/middleware"
	"github.com/PGshen/thinking-map/server/internal/pkg/sse"
	"github.com/PGshen/thinking-map/server/internal/repository"
	"github.com/PGshen/thinking-map/server/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
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
	thinkingMapRepo := repository.NewThinkingMapRepository(db)
	nodeRepo := repository.NewThinkingNodeRepository(db)
	nodeDetailRepo := repository.NewNodeDetailRepository(db)

	// Create services
	authService := service.NewAuthService(db, redisClient, jwtConfig)
	mapService := service.NewMapService(thinkingMapRepo)
	nodeService := service.NewNodeService(nodeRepo, nodeDetailRepo, thinkingMapRepo)

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
				maps.PUT("/:mapId", middleware.MapOwnershipMiddleware(thinkingMapRepo), mapHandler.UpdateMap)
				maps.DELETE("/:mapId", middleware.MapOwnershipMiddleware(thinkingMapRepo), mapHandler.DeleteMap)
				maps.GET("/:mapId", middleware.MapOwnershipMiddleware(thinkingMapRepo), mapHandler.GetMap)
			}

			// Node routes
			nodes := protected.Group("/maps/:mapId/nodes")
			{
				nodes.GET("", nodeHandler.ListNodes)
				nodes.POST("", nodeHandler.CreateNode)
				nodes.PUT("/:nodeId", middleware.NodeOwnershipMiddleware(nodeRepo, thinkingMapRepo), nodeHandler.UpdateNode)
				nodes.DELETE("/:nodeId", middleware.NodeOwnershipMiddleware(nodeRepo, thinkingMapRepo), nodeHandler.DeleteNode)
				nodes.GET("/:nodeId/dependencies", middleware.NodeOwnershipMiddleware(nodeRepo, thinkingMapRepo), nodeHandler.GetDependencies)
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
