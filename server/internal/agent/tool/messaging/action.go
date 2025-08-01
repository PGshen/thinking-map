package messaging

import (
	"context"
	"fmt"

	"github.com/PGshen/thinking-map/server/internal/model"
	"github.com/PGshen/thinking-map/server/internal/model/dto"
	"github.com/PGshen/thinking-map/server/internal/pkg/global"
	"github.com/PGshen/thinking-map/server/internal/pkg/sse"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
	"github.com/google/uuid"
)

const (
	ActionDecompose = "decompose"
	ActionConclude  = "conclude"
)

// 用户选择消息
type ActionChoice struct {
	Introduction string   `json:"introduction" jsonschema:"description=引导语，用于引导用户进行选择"`
	Actions      []string `json:"actions" jsonschema:"description=用户可选择动作列表,enum=decompose,enum=conclude"`
}

type ActionMsgResp []model.Action

func SendActionMsg(ctx context.Context, msg *ActionChoice) (*ActionMsgResp, error) {
	mapID := ctx.Value("mapID").(string)
	nodeID := ctx.Value("nodeID").(string)
	msgActions := ActionMsgResp{}
	for _, action := range msg.Actions {
		var msgAction model.Action
		switch action {
		case ActionDecompose:
			msgAction = model.Action{
				Name:   "开始拆解",
				Method: "POST",
				URL:    "/api/map/node/decompose",
				Param: map[string]any{
					"mapID":  mapID,
					"nodeID": nodeID,
				},
			}
		case ActionConclude:
			msgAction = model.Action{
				Name:   "开始总结",
				Method: "POST",
				URL:    "/api/map/node/conclude",
				Param: map[string]any{
					"mapID":  mapID,
					"nodeID": nodeID,
				},
			}
		}
		msgActions = append(msgActions, msgAction)
	}
	eventID := uuid.New().String()
	event := sse.Event{
		ID:   eventID,
		Type: dto.MsgActionEventType,
		Data: msgActions,
	}
	fmt.Println("mapID", ctx.Value("mapID"))
	fmt.Println("nodeID", ctx.Value("nodeID"))
	global.GetBroker().Publish(ctx.Value("mapID").(string), event)
	return &msgActions, nil
}

func ActionTool() (tool.InvokableTool, error) {
	return utils.InferTool("action", "发送可选操作给用户选择", SendActionMsg)
}
