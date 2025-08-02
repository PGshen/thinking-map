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
	"github.com/cloudwego/eino/schema"
	"github.com/google/uuid"
)

const (
	ActionDecompose = "decompose"
	ActionConclude  = "conclude"
)

// 用户选择消息
type ActionChoice struct {
	Introduction string   `json:"introduction"`
	Actions      []string `json:"actions"`
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
	// 保存消息
	return &msgActions, nil
}

func ActionTool() (tool.InvokableTool, error) {
	actionTool := utils.NewTool(&schema.ToolInfo{
		Name: "sendActionMsg",
		Desc: "发送可选操作消息让用户选择",
		ParamsOneOf: schema.NewParamsOneOfByParams(
			map[string]*schema.ParameterInfo{
				"introduction": {
					Type: schema.String,
					Desc: "引导语，用于引导用户进行选择",
				},
				"actions": {
					Type: schema.Array,
					Desc: "用户可选择动作列表",
					ElemInfo: &schema.ParameterInfo{
						Type: schema.String,
						Enum: []string{ActionDecompose, ActionConclude},
					},
				},
			},
		),
	}, SendActionMsg)
	return actionTool, nil
}
