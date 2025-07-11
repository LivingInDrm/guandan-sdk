package handler

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"guandan/cmd/guandan-server/room"
	"guandan/sdk/domain"
)

// WebSocketHandler handles WebSocket connections
type WebSocketHandler struct {
	restHandler *RestHandler
	upgrader    websocket.Upgrader
}

// NewWebSocketHandler creates a new WebSocket handler
func NewWebSocketHandler(restHandler *RestHandler) *WebSocketHandler {
	return &WebSocketHandler{
		restHandler: restHandler,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				// Allow all origins for development
				return true
			},
		},
	}
}

// HandleWebSocket handles WebSocket connections for a specific room
func (h *WebSocketHandler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Parse room ID from URL
	vars := mux.Vars(r)
	roomID := vars["id"]
	
	// Parse seat from query parameter
	seatStr := r.URL.Query().Get("seat")
	if seatStr == "" {
		http.Error(w, "Seat parameter required", http.StatusBadRequest)
		return
	}
	
	seatNum, err := strconv.Atoi(seatStr)
	if err != nil || seatNum < 0 || seatNum > 3 {
		http.Error(w, "Invalid seat parameter", http.StatusBadRequest)
		return
	}
	
	seat := h.parseSeat(seatNum)
	
	// Get room
	roomKernel, exists := h.restHandler.GetRoom(roomID)
	if !exists {
		http.Error(w, "Room not found", http.StatusNotFound)
		return
	}
	
	// Upgrade connection to WebSocket
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}
	
	// Generate player ID
	playerID := h.generatePlayerID(roomID, seat)
	
	// Add player to room
	err = roomKernel.AddPlayer(playerID, seat, conn)
	if err != nil {
		log.Printf("Failed to add player to room: %v", err)
		conn.Close()
		return
	}
	
	// Handle connection
	h.handleConnection(roomKernel, seat, conn)
}

// handleConnection handles a WebSocket connection
func (h *WebSocketHandler) handleConnection(roomKernel *room.RoomKernel, seat domain.SeatID, conn *websocket.Conn) {
	defer func() {
		conn.Close()
		roomKernel.RemovePlayer(seat)
	}()
	
	// Set up connection parameters
	conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})
	
	// Message reading loop
	for {
		var msg room.WSMessage
		err := conn.ReadJSON(&msg)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}
		
		// Reset read deadline
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		
		// Handle message
		roomKernel.HandleMessage(seat, msg)
	}
}

// Helper methods

func (h *WebSocketHandler) generatePlayerID(roomID string, seat domain.SeatID) string {
	return fmt.Sprintf("%s_player_%s", roomID, seat)
}

func (h *WebSocketHandler) parseSeat(seatNum int) domain.SeatID {
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