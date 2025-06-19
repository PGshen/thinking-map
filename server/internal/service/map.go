package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/thinking-map/server/internal/model"
	"github.com/thinking-map/server/internal/model/dto"
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
	// Create the map
	mapID := uuid.New()
	thinkingMap := &model.ThinkingMap{
		ID:           mapID,
		Title:        req.Title,
		Description:  req.Description,
		RootQuestion: req.RootQuestion,
		Status:       1,
		CreatedBy:    userID,
		UpdatedBy:    userID,
	}

	// Create the root node
	rootNodeID := uuid.New()
	rootNode := &model.ThinkingNode{
		ID:       rootNodeID,
		MapID:    mapID,
		NodeType: 1, // question type
		Content:  req.RootQuestion,
		Position: model.JSONB{
			"x": 0,
			"y": 0,
		},
		Status:    1,
		CreatedBy: userID,
		UpdatedBy: userID,
	}

	// Create map and root node using repository
	if err := s.mapRepo.CreateMap(thinkingMap, rootNode); err != nil {
		return nil, err
	}

	// Return the response
	return &dto.MapResponse{
		ID:           mapID.String(),
		Title:        thinkingMap.Title,
		Description:  thinkingMap.Description,
		RootQuestion: thinkingMap.RootQuestion,
		RootNodeID:   rootNodeID.String(),
		Status:       thinkingMap.Status,
		Metadata:     make(map[string]interface{}),
		CreatedAt:    thinkingMap.CreatedAt,
		UpdatedAt:    thinkingMap.UpdatedAt,
	}, nil
}

// ListMaps retrieves a list of thinking maps
func (s *MapService) ListMaps(ctx context.Context, query dto.MapListQuery, userID uuid.UUID) (*dto.MapListResponse, error) {
	// Get maps from repository
	maps, total, err := s.mapRepo.ListMaps(userID, query.Status, query.Page, query.Limit)
	if err != nil {
		return nil, err
	}

	// Convert to response
	items := make([]dto.MapResponse, len(maps))
	for i, m := range maps {
		// Get node count for each map
		nodeCount, err := s.mapRepo.GetNodeCount(m.ID)
		if err != nil {
			return nil, err
		}

		items[i] = dto.MapResponse{
			ID:           m.ID.String(),
			Title:        m.Title,
			Description:  m.Description,
			RootQuestion: m.RootQuestion,
			Status:       m.Status,
			NodeCount:    int(nodeCount),
			CreatedAt:    m.CreatedAt,
			UpdatedAt:    m.UpdatedAt,
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
	// Get map from repository
	thinkingMap, err := s.mapRepo.GetMap(mapID, userID)
	if err != nil {
		return nil, err
	}

	// Get root node
	rootNode, err := s.mapRepo.GetRootNode(mapID)
	if err != nil {
		return nil, err
	}

	// Get node count
	nodeCount, err := s.mapRepo.GetNodeCount(mapID)
	if err != nil {
		return nil, err
	}

	return &dto.MapResponse{
		ID:           thinkingMap.ID.String(),
		Title:        thinkingMap.Title,
		Description:  thinkingMap.Description,
		RootQuestion: thinkingMap.RootQuestion,
		RootNodeID:   rootNode.ID.String(),
		Status:       thinkingMap.Status,
		Metadata:     make(map[string]interface{}),
		NodeCount:    int(nodeCount),
		CreatedAt:    thinkingMap.CreatedAt,
		UpdatedAt:    thinkingMap.UpdatedAt,
	}, nil
}

// UpdateMap updates a thinking map
func (s *MapService) UpdateMap(ctx context.Context, mapID uuid.UUID, req dto.UpdateMapRequest, userID uuid.UUID) (*dto.MapResponse, error) {
	// Prepare updates
	updates := map[string]interface{}{
		"updated_by": userID,
		"updated_at": time.Now(),
	}
	if req.Title != "" {
		updates["title"] = req.Title
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}
	if req.Status > 0 {
		updates["status"] = req.Status
	}

	// Update map using repository
	if err := s.mapRepo.UpdateMap(mapID, userID, updates); err != nil {
		return nil, err
	}

	// Get updated map
	thinkingMap, err := s.mapRepo.GetMap(mapID, userID)
	if err != nil {
		return nil, err
	}

	return &dto.MapResponse{
		ID:          thinkingMap.ID.String(),
		Title:       thinkingMap.Title,
		Description: thinkingMap.Description,
		Status:      thinkingMap.Status,
		UpdatedAt:   thinkingMap.UpdatedAt,
	}, nil
}

// DeleteMap deletes a thinking map
func (s *MapService) DeleteMap(ctx context.Context, mapID uuid.UUID, userID uuid.UUID) error {
	return s.mapRepo.DeleteMap(mapID, userID)
}
