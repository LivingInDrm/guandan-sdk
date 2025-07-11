package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"guandan/sdk/service"
)

const (
	DefaultPort = ":3000"
)

func main() {
	// Create game service
	gameService := service.NewGameService()
	
	// Setup router
	router := SetupRouter(gameService)
	
	// Create HTTP server
	server := &http.Server{
		Addr:         getPort(),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
	
	// Start server in a goroutine
	go func() {
		log.Printf("Starting Guandan Server on %s", server.Addr)
		log.Printf("Health check: http://localhost%s/api/health", server.Addr)
		log.Printf("WebSocket endpoint: ws://localhost%s/api/room/{id}/ws", server.Addr)
		
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()
	
	// Wait for interrupt signal for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	
	log.Println("Shutting down server...")
	
	// Create context with timeout for graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	// Shutdown server
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}
	
	log.Println("Server exited")
}

// getPort returns the port from environment variable or default
func getPort() string {
	port := os.Getenv("PORT")
	if port == "" {
		port = DefaultPort
	}
	
	if port[0] != ':' {
		port = ":" + port
	}
	
	return port
}