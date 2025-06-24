package service

import (
	"context"
	"time"

	"github.com/PGshen/thinking-map/server/internal/model"
	"github.com/PGshen/thinking-map/server/internal/model/dto"
	"github.com/PGshen/thinking-map/server/internal/pkg/comm"
	"github.com/PGshen/thinking-map/server/internal/repository"

	"github.com/google/uuid"
)

type MapService struct {
	mapRepo repository.ThinkingMap
}

func NewMapService(mapRepo repository.ThinkingMap) *MapService {
	return &MapService{
		mapRepo: mapRepo,
	}
}

// CreateMap creates a new thinking map
func (s *MapService) CreateMap(ctx context.Context, req dto.CreateMapRequest, userID string) (*dto.MapResponse, error) {
	mapID := uuid.NewString()
	thinkingMap := &model.ThinkingMap{
		ID:          mapID,
		UserID:      userID,
		Problem:     req.Problem,
		ProblemType: req.ProblemType,
		Target:      req.Target,
		KeyPoints:   req.KeyPoints,
		Constraints: req.Constraints,
		Conclusion:  req.Conclusion,
		Status:      comm.MapStatusExecuting,
		Metadata:    nil, // 可根据需要初始化
	}

	rootNodeID := uuid.NewString()
	rootNode := &model.ThinkingNode{
		ID:        rootNodeID,
		MapID:     mapID,
		ParentID:  uuid.Nil.String(),
		NodeType:  comm.NodeTypeProblem,
		Question:  req.Problem,
		Status:    comm.NodeStatusPending,
		Position:  model.Position{X: 0, Y: 0},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.mapRepo.Create(ctx, thinkingMap, rootNode); err != nil {
		return nil, err
	}

	return &dto.MapResponse{
		ID:          mapID,
		RootNodeID:  rootNodeID,
		Status:      thinkingMap.Status,
		Problem:     thinkingMap.Problem,
		ProblemType: thinkingMap.ProblemType,
		Target:      thinkingMap.Target,
		KeyPoints:   thinkingMap.KeyPoints,
		Constraints: thinkingMap.Constraints,
		Conclusion:  thinkingMap.Conclusion,
		Metadata:    thinkingMap.Metadata,
		CreatedAt:   thinkingMap.CreatedAt,
		UpdatedAt:   thinkingMap.UpdatedAt,
	}, nil
}

// GetMap retrieves a specific thinking map
func (s *MapService) GetMap(ctx context.Context, mapID string) (*dto.MapResponse, error) {
	thinkingMap, err := s.mapRepo.FindByID(ctx, mapID)
	if err != nil {
		return nil, err
	}
	return &dto.MapResponse{
		ID:          thinkingMap.ID,
		Status:      thinkingMap.Status,
		Problem:     thinkingMap.Problem,
		ProblemType: thinkingMap.ProblemType,
		Target:      thinkingMap.Target,
		KeyPoints:   thinkingMap.KeyPoints,
		Constraints: thinkingMap.Constraints,
		Conclusion:  thinkingMap.Conclusion,
		Metadata:    thinkingMap.Metadata,
		CreatedAt:   thinkingMap.CreatedAt,
		UpdatedAt:   thinkingMap.UpdatedAt,
	}, nil
}

// ListMaps retrieves a list of thinking maps
func (s *MapService) ListMaps(ctx context.Context, query dto.MapListQuery, userID string) (*dto.MapListResponse, error) {
	maps, total, err := s.mapRepo.List(ctx, userID, query.Status, query.Page, query.Limit)
	if err != nil {
		return nil, err
	}

	items := make([]dto.MapResponse, len(maps))
	for i, m := range maps {
		items[i] = dto.MapResponse{
			ID:          m.ID,
			Status:      m.Status,
			Problem:     m.Problem,
			ProblemType: m.ProblemType,
			Target:      m.Target,
			KeyPoints:   m.KeyPoints,
			Constraints: m.Constraints,
			Conclusion:  m.Conclusion,
			Metadata:    m.Metadata,
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

// UpdateMap updates a thinking map
func (s *MapService) UpdateMap(ctx context.Context, mapID string, req dto.UpdateMapRequest, userID string) (*dto.MapResponse, error) {
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
	if err := s.mapRepo.Update(ctx, mapID, updates); err != nil {
		return nil, err
	}
	thinkingMap, err := s.mapRepo.FindByID(ctx, mapID)
	if err != nil {
		return nil, err
	}
	return &dto.MapResponse{
		ID:          thinkingMap.ID,
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
func (s *MapService) DeleteMap(ctx context.Context, mapID string, userID string) error {
	return s.mapRepo.Delete(ctx, mapID)
}
