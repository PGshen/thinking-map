package middleware

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
)

// CORS 跨域中间件
func CORS() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		c.Response.Header.Set("Access-Control-Allow-Origin", "*")
		c.Response.Header.Set("Access-Control-Allow-Credentials", "true")
		c.Response.Header.Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Response.Header.Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if string(c.Request.Method()) == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next(ctx)
	}
}
