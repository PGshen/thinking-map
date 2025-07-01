package thinking

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/PGshen/thinking-map/server/internal/pkg/logger"
	"github.com/cloudwego/eino/schema"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type StreamHandler interface {
	Handle(ctx *gin.Context) (*schema.StreamReader[*schema.Message], error)
}

type StreamReply struct {
	handler StreamHandler
}

func NewStreamReply(handler StreamHandler) gin.HandlerFunc {
	return func(c *gin.Context) {
		sr, err := handler.Handle(c)
		if err != nil {
			logger.Error("[Chat] Error running agent", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{
				"status": "error",
				"error":  err.Error(),
			})
			return
		}

		// 设置SSE响应头
		c.Header("Content-Type", "text/event-stream")
		c.Header("Cache-Control", "no-cache")
		c.Header("Connection", "keep-alive")
		c.Header("Access-Control-Allow-Origin", "*")

		defer func() {
			sr.Close()
			logger.Info("[StreamReply] Finished")
		}()

	outer:
		for {
			select {
			case <-c.Done():
				logger.Info("[StreamReply] Context done")
				return
			default:
				msg, err := sr.Recv()
				if errors.Is(err, io.EOF) {
					logger.Info("[StreamReply] EOF received")
					break outer
				}
				if err != nil {
					logger.Error("[StreamReply] Error receiving message", zap.Error(err))
					break outer
				}

				fmt.Print(msg.Content)
				msg.Content = strings.ReplaceAll(msg.Content, " ", "&nbsp;")

				// 使用Gin的SSE方法发送数据
				c.SSEvent("message", msg.Content)

				time.Sleep(35 * time.Millisecond)
			}
		}
	}
}
