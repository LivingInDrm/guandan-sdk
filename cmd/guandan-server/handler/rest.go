package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"guandan/cmd/guandan-server/room"
	"guandan/sdk/domain"
	"guandan/sdk/service"
)

// RestHandler handles REST API requests
type RestHandler struct {
	gameService service.GameService
	rooms       map[string]*room.RoomKernel
	roomsMutex  sync.RWMutex
}

// NewRestHandler creates a new REST handler
func NewRestHandler(gameService service.GameService) *RestHandler {
	return &RestHandler{
		gameService: gameService,
		rooms:       make(map[string]*room.RoomKernel),
	}
}

// CreateRoomRequest represents a request to create a room
type CreateRoomRequest struct {
	RoomName string `json:"roomName"`
}

// CreateRoomResponse represents a response to create a room
type CreateRoomResponse struct {
	RoomID string `json:"roomId"`
}

// JoinRoomRequest represents a request to join a room
type JoinRoomRequest struct {
	Seat int `json:"seat"`
}

// JoinRoomResponse represents a response to join a room
type JoinRoomResponse struct {
	WSUrl string `json:"wsUrl"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error string `json:"error"`
}

// CreateRoom handles POST /api/room
func (h *RestHandler) CreateRoom(w http.ResponseWriter, r *http.Request) {
	var req CreateRoomRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	// Generate room ID
	roomID := h.generateRoomID()
	
	// Create room kernel
	roomKernel := room.NewRoomKernel(roomID, h.gameService, room.DefaultRoomConfig)
	
	// Start room kernel
	if err := roomKernel.Start(); err != nil {
		h.sendError(w, "Failed to start room", http.StatusInternalServerError)
		return
	}
	
	// Store room
	h.roomsMutex.Lock()
	h.rooms[roomID] = roomKernel
	h.roomsMutex.Unlock()
	
	// Send response
	response := CreateRoomResponse{
		RoomID: roomID,
	}
	
	h.sendJSON(w, response)
}

// JoinRoom handles POST /api/room/{id}/join
func (h *RestHandler) JoinRoom(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roomID := vars["id"]
	
	var req JoinRoomRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	// Validate seat
	if req.Seat < 0 || req.Seat > 3 {
		h.sendError(w, "Invalid seat number", http.StatusBadRequest)
		return
	}
	
	// Check if room exists
	h.roomsMutex.RLock()
	roomKernel, exists := h.rooms[roomID]
	h.roomsMutex.RUnlock()
	
	if !exists {
		h.sendError(w, "Room not found", http.StatusNotFound)
		return
	}
	
	// Check if room is full
	if roomKernel.GetPlayerCount() >= 4 {
		h.sendError(w, "Room is full", http.StatusBadRequest)
		return
	}
	
	// Generate WebSocket URL
	// Check if request comes through nginx proxy (port 5173) or direct (port 8080)
	host := r.Host
	if host == "" || host == "localhost" {
		// If no host or just localhost, determine correct host based on headers
		if r.Header.Get("X-Forwarded-Proto") != "" {
			// Coming through nginx proxy
			host = "localhost:5173"
		} else {
			// Direct access
			host = "localhost:8080"
		}
	}
	wsURL := fmt.Sprintf("ws://%s/api/room/%s/ws?seat=%d", host, roomID, req.Seat)
	
	// Send response
	response := JoinRoomResponse{
		WSUrl: wsURL,
	}
	
	h.sendJSON(w, response)
}

// GetRoomInfo handles GET /api/room/{id}
func (h *RestHandler) GetRoomInfo(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roomID := vars["id"]
	
	// Check if room exists
	h.roomsMutex.RLock()
	roomKernel, exists := h.rooms[roomID]
	h.roomsMutex.RUnlock()
	
	if !exists {
		h.sendError(w, "Room not found", http.StatusNotFound)
		return
	}
	
	// Get room snapshot
	snapshot, err := roomKernel.GetSnapshot()
	if err != nil {
		h.sendError(w, "Failed to get room info", http.StatusInternalServerError)
		return
	}
	
	h.sendJSON(w, snapshot)
}

// ListRooms handles GET /api/rooms
func (h *RestHandler) ListRooms(w http.ResponseWriter, r *http.Request) {
	h.roomsMutex.RLock()
	defer h.roomsMutex.RUnlock()
	
	rooms := make([]map[string]interface{}, 0, len(h.rooms))
	
	for roomID, roomKernel := range h.rooms {
		roomInfo := map[string]interface{}{
			"roomId":      roomID,
			"playerCount": roomKernel.GetPlayerCount(),
			"maxPlayers":  4,
			"isEmpty":     roomKernel.IsEmpty(),
		}
		rooms = append(rooms, roomInfo)
	}
	
	h.sendJSON(w, rooms)
}

// Health handles GET /api/health
func (h *RestHandler) Health(w http.ResponseWriter, r *http.Request) {
	h.sendJSON(w, map[string]interface{}{
		"status":    "ok",
		"timestamp": time.Now(),
		"rooms":     len(h.rooms),
	})
}

// GetRoom returns a room by ID
func (h *RestHandler) GetRoom(roomID string) (*room.RoomKernel, bool) {
	h.roomsMutex.RLock()
	defer h.roomsMutex.RUnlock()
	
	roomKernel, exists := h.rooms[roomID]
	return roomKernel, exists
}

// RemoveRoom removes a room
func (h *RestHandler) RemoveRoom(roomID string) {
	h.roomsMutex.Lock()
	defer h.roomsMutex.Unlock()
	
	if roomKernel, exists := h.rooms[roomID]; exists {
		roomKernel.Stop()
		delete(h.rooms, roomID)
	}
}

// CleanupEmptyRooms removes empty rooms periodically
func (h *RestHandler) CleanupEmptyRooms() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	
	for range ticker.C {
		h.roomsMutex.Lock()
		for roomID, roomKernel := range h.rooms {
			if roomKernel.IsEmpty() {
				roomKernel.Stop()
				delete(h.rooms, roomID)
			}
		}
		h.roomsMutex.Unlock()
	}
}

// Helper methods

func (h *RestHandler) generateRoomID() string {
	return fmt.Sprintf("room_%d", time.Now().UnixNano())
}

func (h *RestHandler) sendJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func (h *RestHandler) sendError(w http.ResponseWriter, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(ErrorResponse{Error: message})
}

func (h *RestHandler) parseSeat(seatNum int) domain.SeatID {
	switch seatNum {
	case 0:
		return domain.SeatEast
	case 1:
		return domain.SeatSouth
	case 2:
		return domain.SeatWest
	case 3:
		return domain.SeatNorth
	default:
		return domain.SeatEast
	}
}