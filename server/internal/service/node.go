package service

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/thinking-map/server/internal/model"
	"github.com/thinking-map/server/internal/model/dto"
	"github.com/thinking-map/server/internal/repository"
)

type NodeService struct {
	nodeRepo repository.ThinkingNode
}

func NewNodeService(nodeRepo repository.ThinkingNode) *NodeService {
	return &NodeService{
		nodeRepo: nodeRepo,
	}
}

// ListNodes 获取某个map下的所有节点
func (s *NodeService) ListNodes(ctx context.Context, mapID string) ([]dto.NodeResponse, error) {
	uid, err := uuid.Parse(mapID)
	if err != nil {
		return nil, errors.New("invalid mapID")
	}
	nodes, err := s.nodeRepo.FindByMapID(ctx, uid)
	if err != nil {
		return nil, err
	}
	var res []dto.NodeResponse
	for _, n := range nodes {
		res = append(res, modelToNodeResponse(n))
	}
	return res, nil
}

// CreateNode 创建节点
func (s *NodeService) CreateNode(ctx context.Context, mapID string, req dto.CreateNodeRequest, userID uuid.UUID) (*dto.NodeResponse, error) {
	mapUUID, err := uuid.Parse(mapID)
	if err != nil {
		return nil, errors.New("invalid mapID")
	}
	parentUUID, err := uuid.Parse(req.ParentID)
	if err != nil {
		return nil, errors.New("invalid parentID")
	}
	posBytes, _ := json.Marshal(req.Position)
	var pos model.JSONB
	_ = json.Unmarshal(posBytes, &pos)
	node := &model.ThinkingNode{
		ID:        uuid.New(),
		MapID:     mapUUID,
		ParentID:  parentUUID,
		NodeType:  nodeTypeStrToInt(req.NodeType),
		Content:   req.Question, // 这里假设Content存储Question
		Position:  pos,
		Status:    1,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		CreatedBy: userID,
		UpdatedBy: userID,
	}
	if err := s.nodeRepo.Create(ctx, node); err != nil {
		return nil, err
	}
	resp := modelToNodeResponse(node)
	resp.MapID = mapID
	resp.ParentID = req.ParentID
	resp.Question = req.Question
	resp.Target = req.Target
	resp.Context = req.Context
	return &resp, nil
}

// UpdateNode 更新节点
func (s *NodeService) UpdateNode(ctx context.Context, nodeID string, req dto.UpdateNodeRequest, userID uuid.UUID) (*dto.NodeResponse, error) {
	uid, err := uuid.Parse(nodeID)
	if err != nil {
		return nil, errors.New("invalid nodeID")
	}
	node, err := s.nodeRepo.FindByID(ctx, uid)
	if err != nil {
		return nil, err
	}
	if req.Question != "" {
		node.Content = req.Question
	}
	if (req.Position != dto.Position{}) {
		posBytes, _ := json.Marshal(req.Position)
		var pos model.JSONB
		_ = json.Unmarshal(posBytes, &pos)
		node.Position = pos
	}
	node.UpdatedAt = time.Now()
	node.UpdatedBy = userID
	if err := s.nodeRepo.Update(ctx, node); err != nil {
		return nil, err
	}
	resp := modelToNodeResponse(node)
	resp.Question = req.Question
	resp.Target = req.Target
	resp.Context = req.Context
	return &resp, nil
}

// DeleteNode 删除节点
func (s *NodeService) DeleteNode(ctx context.Context, nodeID string) error {
	uid, err := uuid.Parse(nodeID)
	if err != nil {
		return errors.New("invalid nodeID")
	}
	return s.nodeRepo.Delete(ctx, uid)
}

// modelToNodeResponse 将model.ThinkingNode转为dto.NodeResponse
func modelToNodeResponse(n *model.ThinkingNode) dto.NodeResponse {
	var pos dto.Position
	posBytes, _ := json.Marshal(n.Position)
	_ = json.Unmarshal(posBytes, &pos)
	return dto.NodeResponse{
		ID:        n.ID.String(),
		MapID:     n.MapID.String(),
		ParentID:  n.ParentID.String(),
		NodeType:  nodeTypeIntToStr(n.NodeType),
		Question:  n.Content, // 假设Content存储Question
		Status:    n.Status,
		Position:  pos,
		CreatedAt: n.CreatedAt,
		UpdatedAt: n.UpdatedAt,
	}
}

// nodeTypeStrToInt 字符串类型转int
func nodeTypeStrToInt(t string) int {
	switch t {
	case "question":
		return 1
	case "analysis":
		return 2
	case "target":
		return 3
	default:
		return 1
	}
}

// nodeTypeIntToStr int类型转字符串
func nodeTypeIntToStr(t int) string {
	switch t {
	case 1:
		return "question"
	case 2:
		return "analysis"
	case 3:
		return "target"
	default:
		return "question"
	}
}
