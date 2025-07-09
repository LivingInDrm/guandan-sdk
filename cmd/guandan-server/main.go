package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"guandan/sdk/domain"
	"guandan/sdk/event"
	"guandan/sdk/service"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Server struct {
	gameService  service.GameService
	connections  map[string]*websocket.Conn
	connMutex    sync.RWMutex
	eventBus     *event.EventBus
}

type ClientMessage struct {
	Type      string      `json:"type"`
	MatchID   string      `json:"match_id,omitempty"`
	PlayerID  string      `json:"player_id,omitempty"`
	Data      interface{} `json:"data,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
}

type ServerMessage struct {
	Type      string      `json:"type"`
	MatchID   string      `json:"match_id,omitempty"`
	Data      interface{} `json:"data,omitempty"`
	Error     string      `json:"error,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
}

type CreateMatchRequest struct {
	Players []PlayerInfo `json:"players"`
	Options service.MatchOptions `json:"options"`
}

type PlayerInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Seat string `json:"seat"`
}

type PlayCardsRequest struct {
	PlayerID string        `json:"player_id"`
	Cards    []domain.Card `json:"cards"`
}

type JoinMatchRequest struct {
	MatchID  string `json:"match_id"`
	PlayerID string `json:"player_id"`
}

func NewServer() *Server {
	gameService := service.NewGameService()
	return &Server{
		gameService: gameService,
		connections: make(map[string]*websocket.Conn),
		eventBus:    event.NewEventBus(1000),
	}
}

func (s *Server) Start(addr string) error {
	s.eventBus.Start()
	defer s.eventBus.Stop()

	// Setup HTTP routes
	http.HandleFunc("/ws", s.handleWebSocket)
	http.HandleFunc("/api/matches", s.handleMatches)
	http.HandleFunc("/api/health", s.handleHealth)
	http.HandleFunc("/", s.handleIndex)

	log.Printf("Server starting on %s", addr)
	return http.ListenAndServe(addr, nil)
}

func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}
	defer conn.Close()

	clientID := r.Header.Get("X-Client-ID")
	if clientID == "" {
		clientID = fmt.Sprintf("client-%d", time.Now().UnixNano())
	}

	s.connMutex.Lock()
	s.connections[clientID] = conn
	s.connMutex.Unlock()

	defer func() {
		s.connMutex.Lock()
		delete(s.connections, clientID)
		s.connMutex.Unlock()
	}()

	log.Printf("Client %s connected", clientID)

	for {
		var msg ClientMessage
		err := conn.ReadJSON(&msg)
		if err != nil {
			log.Printf("WebSocket read error: %v", err)
			break
		}

		msg.Timestamp = time.Now()
		s.handleClientMessage(clientID, msg)
	}
}

func (s *Server) handleClientMessage(clientID string, msg ClientMessage) {
	response := ServerMessage{
		Type:      msg.Type + "_response",
		MatchID:   msg.MatchID,
		Timestamp: time.Now(),
	}

	switch msg.Type {
	case "create_match":
		s.handleCreateMatch(clientID, msg, &response)
	case "join_match":
		s.handleJoinMatch(clientID, msg, &response)
	case "start_deal":
		s.handleStartDeal(clientID, msg, &response)
	case "play_cards":
		s.handlePlayCards(clientID, msg, &response)
	case "pass":
		s.handlePass(clientID, msg, &response)
	case "get_game_state":
		s.handleGetGameState(clientID, msg, &response)
	case "get_valid_plays":
		s.handleGetValidPlays(clientID, msg, &response)
	default:
		response.Error = "Unknown message type"
	}

	s.sendToClient(clientID, response)
}

func (s *Server) handleCreateMatch(clientID string, msg ClientMessage, response *ServerMessage) {
	var req CreateMatchRequest
	data, _ := json.Marshal(msg.Data)
	if err := json.Unmarshal(data, &req); err != nil {
		response.Error = "Invalid request data"
		return
	}

	players := make([]*domain.Player, len(req.Players))
	for i, p := range req.Players {
		seat := s.parseSeat(p.Seat)
		players[i] = domain.NewPlayer(p.ID, p.Name, seat)
	}

	matchID, err := s.gameService.CreateMatch(players, &req.Options)
	if err != nil {
		response.Error = err.Error()
		return
	}

	// Subscribe to match events
	s.gameService.Subscribe(matchID, func(e event.DomainEvent) {
		s.broadcastEvent(matchID, e)
	})

	response.Data = map[string]interface{}{
		"match_id": matchID,
		"players":  req.Players,
	}
}

func (s *Server) handleJoinMatch(clientID string, msg ClientMessage, response *ServerMessage) {
	var req JoinMatchRequest
	data, _ := json.Marshal(msg.Data)
	if err := json.Unmarshal(data, &req); err != nil {
		response.Error = "Invalid request data"
		return
	}

	matchState, err := s.gameService.GetMatchState(domain.MatchID(req.MatchID))
	if err != nil {
		response.Error = err.Error()
		return
	}

	response.Data = map[string]interface{}{
		"match_state": matchState,
	}
}

func (s *Server) handleStartDeal(clientID string, msg ClientMessage, response *ServerMessage) {
	if msg.MatchID == "" {
		response.Error = "Match ID required"
		return
	}

	err := s.gameService.StartNextDeal(domain.MatchID(msg.MatchID))
	if err != nil {
		response.Error = err.Error()
		return
	}

	matchState, err := s.gameService.GetMatchState(domain.MatchID(msg.MatchID))
	if err != nil {
		response.Error = err.Error()
		return
	}

	response.Data = map[string]interface{}{
		"match_state": matchState,
	}
}

func (s *Server) handlePlayCards(clientID string, msg ClientMessage, response *ServerMessage) {
	var req PlayCardsRequest
	data, _ := json.Marshal(msg.Data)
	if err := json.Unmarshal(data, &req); err != nil {
		response.Error = "Invalid request data"
		return
	}

	seat := s.findPlayerSeat(msg.MatchID, req.PlayerID)
	if seat == domain.SeatEast && req.PlayerID == "" {
		response.Error = "Player not found"
		return
	}

	err := s.gameService.PlayCards(domain.MatchID(msg.MatchID), seat, req.Cards)
	if err != nil {
		response.Error = err.Error()
		return
	}

	response.Data = map[string]interface{}{
		"success": true,
	}
}

func (s *Server) handlePass(clientID string, msg ClientMessage, response *ServerMessage) {
	playerID := msg.PlayerID
	if playerID == "" {
		response.Error = "Player ID required"
		return
	}

	seat := s.findPlayerSeat(msg.MatchID, playerID)
	if seat == domain.SeatEast && playerID == "" {
		response.Error = "Player not found"
		return
	}

	err := s.gameService.Pass(domain.MatchID(msg.MatchID), seat)
	if err != nil {
		response.Error = err.Error()
		return
	}

	response.Data = map[string]interface{}{
		"success": true,
	}
}

func (s *Server) handleGetGameState(clientID string, msg ClientMessage, response *ServerMessage) {
	if msg.MatchID == "" {
		response.Error = "Match ID required"
		return
	}

	matchState, err := s.gameService.GetMatchState(domain.MatchID(msg.MatchID))
	if err != nil {
		response.Error = err.Error()
		return
	}

	response.Data = map[string]interface{}{
		"match_state": matchState,
	}
}

func (s *Server) handleGetValidPlays(clientID string, msg ClientMessage, response *ServerMessage) {
	playerID := msg.PlayerID
	if playerID == "" {
		response.Error = "Player ID required"
		return
	}

	seat := s.findPlayerSeat(msg.MatchID, playerID)
	if seat == domain.SeatEast && playerID == "" {
		response.Error = "Player not found"
		return
	}

	validPlays, err := s.gameService.GetValidPlays(domain.MatchID(msg.MatchID), seat)
	if err != nil {
		response.Error = err.Error()
		return
	}

	response.Data = map[string]interface{}{
		"valid_plays": validPlays,
	}
}

func (s *Server) handleMatches(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	switch r.Method {
	case "GET":
		// Return list of active matches (placeholder)
		matches := []map[string]interface{}{
			{
				"match_id": "demo",
				"players":  4,
				"status":   "waiting",
			},
		}
		json.NewEncoder(w).Encode(matches)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":    "ok",
		"timestamp": time.Now(),
	})
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	html := `
<!DOCTYPE html>
<html>
<head>
    <title>Guandan Server</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        .container { max-width: 800px; margin: 0 auto; }
        .status { background: #f0f0f0; padding: 10px; margin: 10px 0; border-radius: 5px; }
        .button { background: #007cba; color: white; padding: 10px 20px; border: none; cursor: pointer; margin: 5px; }
        .button:hover { background: #005a87; }
        .messages { height: 300px; overflow-y: auto; border: 1px solid #ccc; padding: 10px; margin: 10px 0; }
        .message { margin: 5px 0; padding: 5px; background: #f9f9f9; border-radius: 3px; }
    </style>
</head>
<body>
    <div class="container">
        <h1>Guandan Game Server</h1>
        <div class="status">
            <p>Status: <span id="status">Disconnected</span></p>
            <p>Match ID: <span id="matchId">None</span></p>
        </div>
        
        <div>
            <button class="button" onclick="connect()">Connect</button>
            <button class="button" onclick="createMatch()">Create Match</button>
            <button class="button" onclick="startDeal()">Start Deal</button>
            <button class="button" onclick="getGameState()">Get Game State</button>
        </div>
        
        <div class="messages" id="messages"></div>
    </div>

    <script>
        let ws;
        let currentMatchId = '';
        
        function connect() {
            ws = new WebSocket('ws://localhost:8080/ws');
            
            ws.onopen = function() {
                document.getElementById('status').textContent = 'Connected';
                addMessage('Connected to server');
            };
            
            ws.onmessage = function(event) {
                const msg = JSON.parse(event.data);
                addMessage('Received: ' + JSON.stringify(msg, null, 2));
                
                if (msg.type === 'create_match_response' && msg.data && msg.data.match_id) {
                    currentMatchId = msg.data.match_id;
                    document.getElementById('matchId').textContent = currentMatchId;
                }
            };
            
            ws.onclose = function() {
                document.getElementById('status').textContent = 'Disconnected';
                addMessage('Disconnected from server');
            };
        }
        
        function createMatch() {
            if (!ws) return;
            
            const msg = {
                type: 'create_match',
                data: {
                    players: [
                        {id: 'p1', name: 'Player 1', seat: 'east'},
                        {id: 'p2', name: 'Player 2', seat: 'south'},
                        {id: 'p3', name: 'Player 3', seat: 'west'},
                        {id: 'p4', name: 'Player 4', seat: 'north'}
                    ],
                    options: {
                        deal_limit: 0,
                        seed: Date.now()
                    }
                }
            };
            
            ws.send(JSON.stringify(msg));
            addMessage('Sent: create_match');
        }
        
        function startDeal() {
            if (!ws || !currentMatchId) return;
            
            const msg = {
                type: 'start_deal',
                match_id: currentMatchId
            };
            
            ws.send(JSON.stringify(msg));
            addMessage('Sent: start_deal');
        }
        
        function getGameState() {
            if (!ws || !currentMatchId) return;
            
            const msg = {
                type: 'get_game_state',
                match_id: currentMatchId
            };
            
            ws.send(JSON.stringify(msg));
            addMessage('Sent: get_game_state');
        }
        
        function addMessage(text) {
            const messages = document.getElementById('messages');
            const div = document.createElement('div');
            div.className = 'message';
            div.textContent = new Date().toLocaleTimeString() + ': ' + text;
            messages.appendChild(div);
            messages.scrollTop = messages.scrollHeight;
        }
    </script>
</body>
</html>
    `
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

func (s *Server) sendToClient(clientID string, msg ServerMessage) {
	s.connMutex.RLock()
	conn, exists := s.connections[clientID]
	s.connMutex.RUnlock()

	if !exists {
		return
	}

	err := conn.WriteJSON(msg)
	if err != nil {
		log.Printf("Error sending message to client %s: %v", clientID, err)
	}
}

func (s *Server) broadcastEvent(matchID domain.MatchID, event event.DomainEvent) {
	msg := ServerMessage{
		Type:      "game_event",
		MatchID:   string(matchID),
		Data:      event,
		Timestamp: time.Now(),
	}

	s.connMutex.RLock()
	defer s.connMutex.RUnlock()

	for clientID, conn := range s.connections {
		err := conn.WriteJSON(msg)
		if err != nil {
			log.Printf("Error broadcasting to client %s: %v", clientID, err)
		}
	}
}

func (s *Server) parseSeat(seatStr string) domain.SeatID {
	switch seatStr {
	case "east":
		return domain.SeatEast
	case "south":
		return domain.SeatSouth
	case "west":
		return domain.SeatWest
	case "north":
		return domain.SeatNorth
	default:
		return domain.SeatEast
	}
}

func (s *Server) findPlayerSeat(matchID, playerID string) domain.SeatID {
	matchState, err := s.gameService.GetMatchState(domain.MatchID(matchID))
	if err != nil {
		return domain.SeatEast
	}

	for _, player := range matchState.Players {
		if player.ID == playerID {
			return player.SeatID
		}
	}

	return domain.SeatEast
}

func main() {
	server := NewServer()
	
	log.Println("Starting Guandan Server...")
	log.Println("WebSocket endpoint: ws://localhost:8080/ws")
	log.Println("Web interface: http://localhost:8080/")
	log.Println("Health check: http://localhost:8080/api/health")
	
	if err := server.Start(":8080"); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}