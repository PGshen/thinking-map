package messaging

import (
	"context"

	"github.com/PGshen/thinking-map/server/internal/global"
	"github.com/PGshen/thinking-map/server/internal/model"
	"github.com/PGshen/thinking-map/server/internal/model/dto"
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

func SendActionMsg(ctx context.Context, msg *dto.ActionChoice) (*dto.ActionMsgResp, error) {
	mapID := ctx.Value("mapID").(string)
	nodeID := ctx.Value("nodeID").(string)
	msgActions := dto.ActionMsgResp{}
	for _, action := range msg.Actions {
		var msgAction model.Action
		switch action {
		case ActionDecompose:
			msgAction = model.Action{
				Name:   "开始拆解",
				Method: "POST",
				URL:    "/v1/thinking/decomposition",
				Param: map[string]any{
					"mapID":  mapID,
					"nodeID": nodeID,
				},
			}
		case ActionConclude:
			msgAction = model.Action{
				Name:   "开始结论",
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
	messageID := uuid.NewString()
	eventID := uuid.NewString()
	event := sse.Event{
		ID:   eventID,
		Type: dto.MessageActionEventType,
		Data: dto.MessageActionEvent{
			NodeID:    nodeID,
			MessageID: messageID,
			Actions:   msgActions,
		},
	}
	global.GetBroker().PublishToSession(ctx.Value("mapID").(string), event)
	// 保存消息
	global.GetMessageManager().SaveDecompositionMessage(ctx, nodeID, dto.CreateMessageRequest{
		ID:          uuid.NewString(),
		MessageType: model.MsgTypeAction,
		Role:        schema.Tool,
		Content: model.MessageContent{
			Action: msgActions,
		},
	})
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
