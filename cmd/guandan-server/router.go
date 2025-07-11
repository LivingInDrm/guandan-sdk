package main

import (
	"net/http"

	"github.com/gorilla/mux"
	"guandan/cmd/guandan-server/handler"
	"guandan/sdk/service"
)

// SetupRouter creates and configures the HTTP router
func SetupRouter(gameService service.GameService) *mux.Router {
	// Create handlers
	restHandler := handler.NewRestHandler(gameService)
	wsHandler := handler.NewWebSocketHandler(restHandler)
	
	// Start cleanup routine for empty rooms
	go restHandler.CleanupEmptyRooms()
	
	// Create router
	r := mux.NewRouter()
	
	// Root route redirect to API health
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/api/health", http.StatusFound)
	}).Methods("GET")
	
	// API routes
	api := r.PathPrefix("/api").Subrouter()
	
	// Room management routes
	api.HandleFunc("/room", restHandler.CreateRoom).Methods("POST")
	api.HandleFunc("/room/{id}/join", restHandler.JoinRoom).Methods("POST")
	api.HandleFunc("/room/{id}", restHandler.GetRoomInfo).Methods("GET")
	api.HandleFunc("/rooms", restHandler.ListRooms).Methods("GET")
	
	// WebSocket routes
	api.HandleFunc("/room/{id}/ws", wsHandler.HandleWebSocket)
	
	// Health check route
	api.HandleFunc("/health", restHandler.Health).Methods("GET")
	
	// CORS middleware
	r.Use(corsMiddleware)
	
	// Logging middleware
	r.Use(loggingMiddleware)
	
	return r
}

// corsMiddleware adds CORS headers
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		
		next.ServeHTTP(w, r)
	})
}

// loggingMiddleware logs HTTP requests
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}