package thinking

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/PGshen/thinking-map/server/internal/pkg/comm"
	"github.com/PGshen/thinking-map/server/internal/pkg/logger"
	"github.com/cloudwego/eino/schema"
	sse2 "github.com/gin-contrib/sse"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type StreamHandler interface {
	Handle(ctx *gin.Context) (id string, event string, sr *schema.StreamReader[*schema.Message], err error)
}

type StreamReply struct {
	handler StreamHandler
}

func NewStreamReply(handler StreamHandler) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, event, sr, err := handler.Handle(c)
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
		// 发送msgID
		c.Render(-1, sse2.Event{
			Id:    id,
			Event: comm.EventID,
			Data:  id,
		})

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
				// c.SSEvent("message", msg.Content)
				c.Render(-1, sse2.Event{
					Id:    id,
					Event: event,
					Data:  msg.Content,
				})
				c.Writer.Flush()

				// time.Sleep(100 * time.Millisecond)
			}
		}
	}
}
