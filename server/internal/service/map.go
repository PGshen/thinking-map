package service

import (
	"context"
	"time"

	"github.com/PGshen/thinking-map/server/internal/model"
	"github.com/PGshen/thinking-map/server/internal/model/dto"
	"github.com/PGshen/thinking-map/server/internal/pkg/comm"
	"github.com/PGshen/thinking-map/server/internal/repository"
	"gorm.io/datatypes"

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
		Title:       req.Title,
		Problem:     req.Problem,
		ProblemType: req.ProblemType,
		Target:      req.Target,
		KeyPoints:   req.KeyPoints,
		Constraints: req.Constraints,
		Conclusion:  "",
		Metadata:    datatypes.JSON{},
		Status:      comm.MapStatusInitial,
	}

	rootNodeID := uuid.NewString()
	rootNode := &model.ThinkingNode{
		ID:        rootNodeID,
		MapID:     mapID,
		ParentID:  uuid.Nil.String(),
		NodeType:  comm.NodeTypeProblem,
		Question:  req.Problem,
		Target:    req.Target,
		Status:    comm.NodeStatusPending,
		Position:  model.Position{X: 0, Y: 0},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.mapRepo.Create(ctx, thinkingMap, rootNode); err != nil {
		return nil, err
	}

	resp := dto.ToMapResponse(thinkingMap)
	return &resp, nil
}

// GetMap retrieves a specific thinking map
func (s *MapService) GetMap(ctx context.Context, mapID string) (*dto.MapResponse, error) {
	thinkingMap, err := s.mapRepo.FindByID(ctx, mapID)
	if err != nil {
		return nil, err
	}
	resp := dto.ToMapResponse(thinkingMap)
	return &resp, nil
}

// ListMaps retrieves a list of thinking maps
func (s *MapService) ListMaps(ctx context.Context, query dto.MapListQuery, userID string) (*dto.MapListResponse, error) {
	// Get current time in UTC+8 (China Standard Time)
	now := time.Now().In(time.FixedZone("CST", 8*60*60))
	var startTime, endTime time.Time
	if query.DateRange != "" {
		switch query.DateRange {
		case "this-week":
			// Calculate the start of this week (Monday)
			weekday := int(now.Weekday())
			if weekday == 0 { // Sunday
				weekday = 7
			}
			startTime = now.AddDate(0, 0, -weekday+1).Truncate(24 * time.Hour)
			endTime = now
		case "last-week":
			// Calculate the start of last week
			weekday := int(now.Weekday())
			if weekday == 0 { // Sunday
				weekday = 7
			}
			startTime = now.AddDate(0, 0, -weekday-6).Truncate(24 * time.Hour)
			endTime = now.AddDate(0, 0, -weekday+1).Truncate(24 * time.Hour)
		case "this-month":
			// Calculate the start of this month
			startTime = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.FixedZone("CST", 8*60*60))
			endTime = now
		case "all-time":
			// No time filter needed
			break
		default:
			// Invalid date range, ignore the filter
			break
		}
	}
	maps, total, err := s.mapRepo.List(ctx, userID, query.Status, query.ProblemType, query.Search, startTime, endTime, query.Page, query.Limit)
	if err != nil {
		return nil, err
	}

	items := make([]dto.MapResponse, len(maps))
	for i, m := range maps {
		items[i] = dto.ToMapResponse(m)
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
	resp := dto.ToMapResponse(thinkingMap)
	return &resp, nil
}

// DeleteMap deletes a thinking map
func (s *MapService) DeleteMap(ctx context.Context, mapID string, userID string) error {
	return s.mapRepo.Delete(ctx, mapID)
}
