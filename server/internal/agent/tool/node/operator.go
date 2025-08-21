package node

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

// CreateNodeRequest 创建节点请求参数
type CreateNodeRequest struct {
	NodeID   string  `json:"nodeID"`
	NodeType string  `json:"nodeType"`
	Question string  `json:"question"`
	Target   string  `json:"target"`
	X        float64 `json:"x"`
	Y        float64 `json:"y"`
}

// UpdateNodeRequest 更新节点请求参数
type UpdateNodeRequest struct {
	NodeID   string  `json:"nodeID"`
	Question string  `json:"question"`
	Target   string  `json:"target"`
	X        float64 `json:"x"`
	Y        float64 `json:"y"`
}

// DeleteNodeRequest 删除节点请求参数
type DeleteNodeRequest struct {
	NodeID string `json:"nodeID"`
}

// DeleteNodeResponse 删除节点响应
type DeleteNodeResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// CreateNodeFunc 创建节点函数
func CreateNodeFunc(ctx context.Context, req *CreateNodeRequest) (*dto.NodeResponse, error) {
	// 构建DTO请求
	mapID := ctx.Value("mapID").(string)
	parentID := ctx.Value("nodeID").(string)
	createReq := dto.CreateNodeRequest{
		MapID:    mapID,
		ParentID: parentID,
		NodeType: req.NodeType,
		Question: req.Question,
		Target:   req.Target,
		Position: model.Position{
			X: req.X,
			Y: req.Y,
		},
	}

	// 调用全局节点操作器
	resp, err := global.GetNodeOperator().CreateNode(ctx, createReq)
	if err != nil {
		return nil, err
	}
	// 发送节点创建事件
	global.GetBroker().PublishToSession(mapID, sse.Event{
		ID:   uuid.NewString(),
		Type: dto.NodeCreatedEventType,
		Data: dto.NodeCreatedEvent{
			NodeID:   resp.ID,
			ParentID: parentID,
			NodeType: resp.NodeType,
			Question: resp.Question,
			Target:   resp.Target,
			Position: resp.Position,
		},
	})
	// 发送节点创建消息
	global.GetBroker().PublishToSession(mapID, sse.Event{
		ID:   uuid.NewString(),
		Type: dto.MessageNoticeEventType,
		Data: dto.MessageNoticeEvent{
			NodeID:    resp.ID,
			MessageID: uuid.NewString(),
			Notice: model.Notice{
				Type:    model.NoticeTypeSuccess,
				Name:    "节点创建",
				Content: resp.Question,
			},
		},
	})
	// 保存消息到数据库
	global.GetMessageManager().SaveDecompositionMessage(ctx, parentID, dto.CreateMessageRequest{
		ID:          uuid.NewString(),
		MessageType: model.MsgTypeNotice,
		Role:        schema.Tool,
		Content: model.MessageContent{
			Notice: model.Notice{
				Type:    model.NoticeTypeSuccess,
				Name:    "节点创建",
				Content: resp.Question,
			},
		},
	})

	return resp, nil
}

// UpdateNodeFunc 更新节点函数
func UpdateNodeFunc(ctx context.Context, req *UpdateNodeRequest) (*dto.NodeResponse, error) {
	// 构建DTO请求
	updateReq := dto.UpdateNodeRequest{
		Question: req.Question,
		Target:   req.Target,
	}

	// 如果提供了坐标，则更新位置
	if req.X != 0 || req.Y != 0 {
		updateReq.Position = model.Position{
			X: req.X,
			Y: req.Y,
		}
	}

	// 调用全局节点操作器
	resp, err := global.GetNodeOperator().UpdateNode(ctx, req.NodeID, updateReq)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// DeleteNodeFunc 删除节点函数
func DeleteNodeFunc(ctx context.Context, req *DeleteNodeRequest) (*DeleteNodeResponse, error) {
	// 调用全局节点操作器
	err := global.GetNodeOperator().DeleteNode(ctx, req.NodeID)
	if err != nil {
		return nil, err
	}

	// 返回成功消息
	return &DeleteNodeResponse{
		Success: true,
		Message: "节点删除成功",
	}, nil
}

// CreateNodeTool 创建节点工具
func CreateNodeTool() (tool.InvokableTool, error) {
	tool := utils.NewTool(&schema.ToolInfo{
		Name: "createNode",
		Desc: "创建新的思维导图节点",
		ParamsOneOf: schema.NewParamsOneOfByParams(
			map[string]*schema.ParameterInfo{
				"nodeID": {
					Type:     schema.String,
					Desc:     "节点ID: uuid格式，如954b9be3-cae8-43bf-9ac6-b423bb62f8f4",
					Required: true,
				},
				"nodeType": {
					Type:     schema.String,
					Desc:     "节点类型，如problem、information、analysis、generation、evaluation等",
					Required: true,
				},
				"question": {
					Type:     schema.String,
					Desc:     "节点问题描述",
					Required: true,
				},
				"target": {
					Type: schema.String,
					Desc: "节点目标描述",
				},
				"x": {
					Type: schema.Number,
					Desc: "节点X坐标位置",
				},
				"y": {
					Type: schema.Number,
					Desc: "节点Y坐标位置",
				},
			},
		),
	}, CreateNodeFunc)
	return tool, nil
}

// UpdateNodeTool 更新节点工具
func UpdateNodeTool() (tool.InvokableTool, error) {
	tool := utils.NewTool(&schema.ToolInfo{
		Name: "updateNode",
		Desc: "更新现有思维导图节点的信息",
		ParamsOneOf: schema.NewParamsOneOfByParams(
			map[string]*schema.ParameterInfo{
				"nodeID": {
					Type:     schema.String,
					Desc:     "要更新的节点ID",
					Required: true,
				},
				"question": {
					Type: schema.String,
					Desc: "更新的节点问题描述",
				},
				"target": {
					Type: schema.String,
					Desc: "更新的节点目标描述",
				},
				"x": {
					Type: schema.Number,
					Desc: "更新的节点X坐标位置",
				},
				"y": {
					Type: schema.Number,
					Desc: "更新的节点Y坐标位置",
				},
			},
		),
	}, UpdateNodeFunc)
	return tool, nil
}

// DeleteNodeTool 删除节点工具
func DeleteNodeTool() (tool.InvokableTool, error) {
	tool := utils.NewTool(&schema.ToolInfo{
		Name: "deleteNode",
		Desc: "删除指定的思维导图节点",
		ParamsOneOf: schema.NewParamsOneOfByParams(
			map[string]*schema.ParameterInfo{
				"nodeID": {
					Type:     schema.String,
					Desc:     "要删除的节点ID",
					Required: true,
				},
			},
		),
	}, DeleteNodeFunc)
	return tool, nil
}

// GetAllNodeTools 获取所有节点操作工具
func GetAllNodeTools() ([]tool.BaseTool, error) {
	createNodeTool, err := CreateNodeTool()
	if err != nil {
		return nil, err
	}

	updateNodeTool, err := UpdateNodeTool()
	if err != nil {
		return nil, err
	}

	deleteNodeTool, err := DeleteNodeTool()
	if err != nil {
		return nil, err
	}

	return []tool.BaseTool{createNodeTool, updateNodeTool, deleteNodeTool}, nil
}
