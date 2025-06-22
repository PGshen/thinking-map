package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/thinking-map/server/internal/model"
	"github.com/thinking-map/server/internal/model/dto"
	"github.com/thinking-map/server/internal/pkg/comm"
	"github.com/thinking-map/server/internal/repository"
)

type MapService struct {
	mapRepo *repository.MapRepository
}

func NewMapService(mapRepo *repository.MapRepository) *MapService {
	return &MapService{
		mapRepo: mapRepo,
	}
}

// CreateMap creates a new thinking map
func (s *MapService) CreateMap(ctx context.Context, req dto.CreateMapRequest, userID uuid.UUID) (*dto.MapResponse, error) {
	mapID := uuid.New()
	thinkingMap := &model.ThinkingMap{
		ID:          mapID,
		UserID:      userID,
		Problem:     req.Problem,
		ProblemType: req.ProblemType,
		Target:      req.Target,
		KeyPoints:   req.KeyPoints,
		Constraints: req.Constraints,
		Conclusion:  req.Conclusion,
		Status:      1,
		Metadata:    nil, // 可根据需要初始化
	}

	rootNodeID := uuid.New()
	rootNode := &model.ThinkingNode{
		ID:        rootNodeID,
		MapID:     mapID,
		ParentID:  uuid.Nil,
		NodeType:  comm.NodeTypeProblem,
		Question:  req.Problem,
		Status:    1,
		Position:  model.Position{X: 0, Y: 0},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.mapRepo.CreateMap(thinkingMap, rootNode); err != nil {
		return nil, err
	}

	return &dto.MapResponse{
		ID:          mapID.String(),
		RootNodeID:  rootNodeID.String(),
		Status:      thinkingMap.Status,
		Problem:     thinkingMap.Problem,
		ProblemType: thinkingMap.ProblemType,
		Target:      thinkingMap.Target,
		KeyPoints:   thinkingMap.KeyPoints,
		Constraints: thinkingMap.Constraints,
		Conclusion:  thinkingMap.Conclusion,
		Metadata:    thinkingMap.Metadata,
		NodeCount:   1,
		CreatedAt:   thinkingMap.CreatedAt,
		UpdatedAt:   thinkingMap.UpdatedAt,
	}, nil
}

// ListMaps retrieves a list of thinking maps
func (s *MapService) ListMaps(ctx context.Context, query dto.MapListQuery, userID uuid.UUID) (*dto.MapListResponse, error) {
	maps, total, err := s.mapRepo.ListMaps(userID, query.Status, query.Page, query.Limit)
	if err != nil {
		return nil, err
	}

	items := make([]dto.MapResponse, len(maps))
	for i, m := range maps {
		nodeCount, err := s.mapRepo.GetNodeCount(m.ID)
		if err != nil {
			return nil, err
		}
		items[i] = dto.MapResponse{
			ID:          m.ID.String(),
			Status:      m.Status,
			Problem:     m.Problem,
			ProblemType: m.ProblemType,
			Target:      m.Target,
			KeyPoints:   m.KeyPoints,
			Constraints: m.Constraints,
			Conclusion:  m.Conclusion,
			Metadata:    m.Metadata,
			NodeCount:   int(nodeCount),
			CreatedAt:   m.CreatedAt,
			UpdatedAt:   m.UpdatedAt,
		}
	}

	return &dto.MapListResponse{
		Total: int(total),
		Page:  query.Page,
		Limit: query.Limit,
		Items: items,
	}, nil
}

// GetMap retrieves a specific thinking map
func (s *MapService) GetMap(ctx context.Context, mapID uuid.UUID, userID uuid.UUID) (*dto.MapResponse, error) {
	thinkingMap, err := s.mapRepo.GetMap(mapID, userID)
	if err != nil {
		return nil, err
	}
	rootNode, err := s.mapRepo.GetRootNode(mapID)
	if err != nil {
		return nil, err
	}
	nodeCount, err := s.mapRepo.GetNodeCount(mapID)
	if err != nil {
		return nil, err
	}
	return &dto.MapResponse{
		ID:          thinkingMap.ID.String(),
		RootNodeID:  rootNode.ID.String(),
		Status:      thinkingMap.Status,
		Problem:     thinkingMap.Problem,
		ProblemType: thinkingMap.ProblemType,
		Target:      thinkingMap.Target,
		KeyPoints:   thinkingMap.KeyPoints,
		Constraints: thinkingMap.Constraints,
		Conclusion:  thinkingMap.Conclusion,
		Metadata:    thinkingMap.Metadata,
		NodeCount:   int(nodeCount),
		CreatedAt:   thinkingMap.CreatedAt,
		UpdatedAt:   thinkingMap.UpdatedAt,
	}, nil
}

// UpdateMap updates a thinking map
func (s *MapService) UpdateMap(ctx context.Context, mapID uuid.UUID, req dto.UpdateMapRequest, userID uuid.UUID) (*dto.MapResponse, error) {
	updates := map[string]interface{}{
		"updated_at": time.Now(),
	}
	if req.Status > 0 {
		updates["status"] = req.Status
	}
	if req.Problem != "" {
		updates["problem"] = req.Problem
	}
	if req.ProblemType != "" {
		updates["problem_type"] = req.ProblemType
	}
	if req.Target != "" {
		updates["target"] = req.Target
	}
	if req.KeyPoints != nil {
		updates["key_points"] = req.KeyPoints
	}
	if req.Constraints != nil {
		updates["constraints"] = req.Constraints
	}
	if req.Conclusion != "" {
		updates["conclusion"] = req.Conclusion
	}
	if err := s.mapRepo.UpdateMap(mapID, userID, updates); err != nil {
		return nil, err
	}
	thinkingMap, err := s.mapRepo.GetMap(mapID, userID)
	if err != nil {
		return nil, err
	}
	return &dto.MapResponse{
		ID:          thinkingMap.ID.String(),
		Status:      thinkingMap.Status,
		Problem:     thinkingMap.Problem,
		ProblemType: thinkingMap.ProblemType,
		Target:      thinkingMap.Target,
		KeyPoints:   thinkingMap.KeyPoints,
		Constraints: thinkingMap.Constraints,
		Conclusion:  thinkingMap.Conclusion,
		Metadata:    thinkingMap.Metadata,
		UpdatedAt:   thinkingMap.UpdatedAt,
	}, nil
}

// DeleteMap deletes a thinking map
func (s *MapService) DeleteMap(ctx context.Context, mapID uuid.UUID, userID uuid.UUID) error {
	return s.mapRepo.DeleteMap(mapID, userID)
}
