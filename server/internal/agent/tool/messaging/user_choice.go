package messaging

import (
	"context"

	"github.com/PGshen/thinking-map/server/internal/model/dto"
	"github.com/PGshen/thinking-map/server/internal/pkg/global"
	"github.com/PGshen/thinking-map/server/internal/pkg/sse"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
	"github.com/google/uuid"
)

// 用户选择消息

type UserChoiceMsgResp struct {
	Ok  bool   `json:"ok"`
	Msg string `json:"msg"`
}

func SendUserChoiceMsg(ctx context.Context, msg *dto.MsgUserChoiceEvent) (*UserChoiceMsgResp, error) {
	eventID := uuid.New().String()
	event := sse.Event{
		ID:   eventID,
		Type: dto.MsgUserChoiceEventType,
		Data: msg,
	}
	global.GetBroker().Publish(msg.MapID, event)
	return &UserChoiceMsgResp{
		Ok:  true,
		Msg: "success",
	}, nil
}

func UserChoiceTool() (tool.InvokableTool, error) {
	return utils.InferTool("user_choice", "发送选择操作给用户", SendUserChoiceMsg)
}
