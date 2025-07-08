package internal

import (
	"log"
	"net/http"
	"sync"

	"github.com/go-chi/chi/v5"

	"github.com/gorilla/websocket"
)

// Declare a global map to hold WebSocket clients
// and a mutex to protect concurrent access to this map.
var (
	clients      = make(map[string]*websocket.Conn)
	clientsMutex sync.RWMutex
)

// Declare a WebSocket upgrader to handle the upgrade from HTTP to WebSocket.
// The upgrader checks the origin of the request to allow or deny the upgrade.
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// Handles incoming WebSocket connections.
func WebSocketHandler(w http.ResponseWriter, r *http.Request) {
	playerID := chi.URLParam(r, "playerID")
	log.Print(
		"PLAYER ID: ", playerID,
	)
	if playerID == "" {
		http.Error(w, "Player ID is required", http.StatusBadRequest)
		return
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}
	log.Printf("Player %s connected via WebSocket", playerID)

	// Store the WebSocket connection in the global map.
	clientsMutex.Lock()
	clients[playerID] = conn
	clientsMutex.Unlock()

	// Listen for disconnect
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			log.Printf("Player %s disconnected: %v", playerID, err)

			clientsMutex.Lock()
			delete(clients, playerID)
			clientsMutex.Unlock()

			if err := conn.Close(); err != nil {
				log.Printf("Error closing WebSocket connection for player %s: %v", playerID, err)
			}
			break
		}
	}
}

func NotifyPlayerLobby(playerID, lobbyID string) error {
	clientsMutex.RLock()
	conn, exists := clients[playerID]
	clientsMutex.RUnlock()

	if !exists {
		return nil
	}

	message := map[string]string{
		"type":     "lobby_created",
		"lobby_id": lobbyID,
	}

	if err := conn.WriteJSON(message); err != nil {
		log.Printf("Failed to send message to player %s: %v", playerID, err)
		clientsMutex.Lock()
		delete(clients, playerID)
		clientsMutex.Unlock()
		if err := conn.Close(); err != nil {
			log.Printf("Error closing WebSocket connection for player %s: %v", playerID, err)
		}
	}
	return nil

}
