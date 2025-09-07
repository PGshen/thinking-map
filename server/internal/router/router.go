/*
 * @Date: 2025-06-18 23:52:48
 * @LastEditors: peng pgs1108pgs@gmail.com
 * @LastEditTime: 2025-06-25 00:16:23
 * @FilePath: /thinking-map/server/internal/router/router.go
 */
package router

import (
	"github.com/PGshen/thinking-map/server/internal/global"
	"github.com/PGshen/thinking-map/server/internal/handler"
	thinkinghandler "github.com/PGshen/thinking-map/server/internal/handler/thinking"
	"github.com/PGshen/thinking-map/server/internal/middleware"
	"github.com/PGshen/thinking-map/server/internal/repository"
	"github.com/PGshen/thinking-map/server/internal/service"

	"time"

	"github.com/gin-contrib/cors"
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
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "http://127.0.0.1:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-REFRESH-TOKEN", "Cache-Control"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Create repositories
	mapRepo := repository.NewThinkingMapRepository(db)
	nodeRepo := repository.NewThinkingNodeRepository(db)
	messageRepo := repository.NewMessageRepository(db)

	// Create services
	authService := service.NewAuthService(db, redisClient, jwtConfig)
	mapService := service.NewMapService(mapRepo)
	nodeService := service.NewNodeService(nodeRepo, mapRepo)
	contextManager := service.NewContextManager(nodeRepo, mapRepo, messageRepo)
	understandingService := service.NewUnderstandingService(messageRepo, nodeRepo)
	decompositionService := service.NewDecompositionService(contextManager, nodeRepo)

	// Create handlers
	authHandler := handler.NewAuthHandler(authService)
	mapHandler := handler.NewMapHandler(mapService)
	nodeHandler := handler.NewNodeHandler(nodeService)
	understandingHandler := thinkinghandler.NewUnderstandingHandler(understandingService)
	decompositionHandler := thinkinghandler.NewDecompositionHandler(decompositionService)
	repeaterHandler := thinkinghandler.NewRepeaterHandler()

	// 使用全局 broker
	sseHandler := handler.NewSSEHandler(global.GetBroker(), mapRepo)

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
				maps.PUT("/:mapID", middleware.MapOwnershipMiddleware(mapRepo), mapHandler.UpdateMap)
				maps.DELETE("/:mapID", middleware.MapOwnershipMiddleware(mapRepo), mapHandler.DeleteMap)
				maps.GET("/:mapID", middleware.MapOwnershipMiddleware(mapRepo), mapHandler.GetMap)
			}

			// Node routes
			nodes := protected.Group("/maps/:mapID/nodes")
			{
				nodes.GET("", middleware.MapOwnershipMiddleware(mapRepo), nodeHandler.ListNodes)
				nodes.POST("", nodeHandler.CreateNode)
				nodes.GET("/executable-nodes", middleware.MapOwnershipMiddleware(mapRepo), nodeHandler.ExecutableNodes)
				nodes.PUT("/:nodeID", middleware.NodeOwnershipMiddleware(nodeRepo, mapRepo), nodeHandler.UpdateNode)
				nodes.DELETE("/:nodeID", middleware.NodeOwnershipMiddleware(nodeRepo, mapRepo), nodeHandler.DeleteNode)
				nodes.PUT("/:nodeID/context", middleware.NodeOwnershipMiddleware(nodeRepo, mapRepo), nodeHandler.UpdateNodeContext)
				nodes.PUT("/:nodeID/context/reset", middleware.NodeOwnershipMiddleware(nodeRepo, mapRepo), nodeHandler.ResetNodeContext)
				nodes.GET("/:nodeID/messages", middleware.NodeOwnershipMiddleware(nodeRepo, mapRepo), nodeHandler.GetNodeMessages)
			}

			// Thinking routes
			thinking := protected.Group("/thinking")
			{
				thinking.POST("/understanding", thinkinghandler.NewStreamReply(understandingHandler))
				// thinking.POST("/decomposition", thinkinghandler.NewStreamReply(decompositionHandler))
				thinking.POST("/decomposition", decompositionHandler.Handle)
				thinking.POST("/repeat", thinkinghandler.NewStreamReply(repeaterHandler))
			}

			// SSE routes
			sse := protected.Group("/sse")
			{
				sse.GET("/connect/:mapID", sseHandler.Connect)
				sse.POST("/send-event/:mapID", sseHandler.SendEvent)
			}
		}
	}

	return r
}
