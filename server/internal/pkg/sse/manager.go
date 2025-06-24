package sse

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/PGshen/thinking-map/server/internal/model/dto"

	"github.com/google/uuid"
)

// EventManager manages SSE connections and events
type EventManager struct {
	clients map[string]map[string]chan []byte // mapID -> connectionID -> channel
	mu      sync.RWMutex
	done    chan struct{}
}

// NewEventManager creates a new EventManager
func NewEventManager() *EventManager {
	return &EventManager{
		clients: make(map[string]map[string]chan []byte),
		done:    make(chan struct{}),
	}
}

// Connect creates a new SSE connection for a map
func (m *EventManager) Connect(mapID string) (string, chan []byte) {
	m.mu.Lock()
	defer m.mu.Unlock()

	connectionID := uuid.New().String()
	eventChan := make(chan []byte, 100)

	if _, exists := m.clients[mapID]; !exists {
		m.clients[mapID] = make(map[string]chan []byte)
	}
	m.clients[mapID][connectionID] = eventChan

	// Send connection event
	event := dto.SSEConnectionResponse{
		ConnectionID: connectionID,
		MapID:        mapID,
		Timestamp:    time.Now(),
	}
	data, _ := json.Marshal(event)
	eventChan <- []byte(fmt.Sprintf("event: connected\ndata: %s\n\n", data))

	return connectionID, eventChan
}

// Disconnect removes an SSE connection
func (m *EventManager) Disconnect(mapID, connectionID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if clients, exists := m.clients[mapID]; exists {
		if ch, ok := clients[connectionID]; ok {
			// Send disconnection event
			event := dto.SSEDisconnectionResponse{
				ConnectionID: connectionID,
				Reason:       "user_disconnected",
				Timestamp:    time.Now(),
			}
			data, _ := json.Marshal(event)
			ch <- []byte(fmt.Sprintf("event: disconnected\ndata: %s\n\n", data))

			close(ch)
			delete(clients, connectionID)
		}
		if len(clients) == 0 {
			delete(m.clients, mapID)
		}
	}
}

// BroadcastEvent sends an event to all clients of a map
func (m *EventManager) BroadcastEvent(mapID, eventType string, data interface{}) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	clients, exists := m.clients[mapID]
	if !exists {
		return fmt.Errorf("no clients for map %s", mapID)
	}

	eventData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	message := fmt.Sprintf("event: %s\ndata: %s\n\n", eventType, eventData)
	for _, ch := range clients {
		select {
		case ch <- []byte(message):
		default:
			// Channel is full, skip this client
		}
	}

	return nil
}

// Close closes the event manager and all connections
func (m *EventManager) Close() {
	m.mu.Lock()
	defer m.mu.Unlock()

	close(m.done)
	for _, clients := range m.clients {
		for _, ch := range clients {
			close(ch)
		}
	}
	m.clients = make(map[string]map[string]chan []byte)
}
