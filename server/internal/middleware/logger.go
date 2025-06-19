package middleware

import (
	"context"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/thinking-map/server/internal/pkg/logger"
	"go.uber.org/zap"
)

// Logger 日志中间件
func Logger() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		start := time.Now()
		path := string(c.Request.URI().Path())
		query := string(c.Request.URI().QueryString())
		c.Next(ctx)
		cost := time.Since(start)
		logger.Info("request",
			zap.Int("status", c.Response.StatusCode()),
			zap.String("method", string(c.Request.Method())),
			zap.String("path", path),
			zap.String("query", query),
			zap.String("ip", c.ClientIP()),
			zap.String("user-agent", string(c.Request.UserAgent())),
			zap.Duration("cost", cost),
		)
	}
}
