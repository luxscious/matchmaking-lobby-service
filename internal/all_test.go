package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// Redis Container setup for testing
func startRedisContainer(ctx context.Context) (testcontainers.Container, string, error) {
	req := testcontainers.ContainerRequest{
		Image:        "redis:latest",
		ExposedPorts: []string{"6379/tcp"},
		WaitingFor:   wait.ForListeningPort("6379/tcp"),
	}
	redisC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, "", err
	}
	endpoint, err := redisC.Endpoint(ctx, "")
	if err != nil {
		return nil, "", err
	}
	return redisC, endpoint, nil
}

// Tests matchmaking functionality
func TestMatchmaking(t *testing.T) {
	ctx := context.Background()

	// Start ephemeral Redis container
	redisC, endpoint, err := startRedisContainer(ctx)
	assert.NoError(t, err)
	defer func() {
		if err := redisC.Terminate(ctx); err != nil {
			t.Errorf("failed to terminate redis container: %v", err)
		}
	}()

	// Setup Redis client
	rdb := redis.NewClient(&redis.Options{
		Addr: endpoint,
	})

	redisClient := &RedisClient{
		Client: rdb,
		Ctx:    ctx,
	}

	// Enqueue 5 players
	for i := 1; i <= 5; i++ {
		p := &Player{
			PlayerID:    "player" + string(rune('A'+(i-1))),
			SkillRating: 1400 + i*10,
		}
		err := redisClient.PushPlayerToQueue(p)
		assert.NoError(t, err)
	}

	// Run matchmaking once
	err = MatchPlayers(ctx, redisClient)
	assert.NoError(t, err)

	// Validate the queue is empty
	remaining, err := redisClient.GetQueuedPlayers()
	assert.NoError(t, err)
	assert.Len(t, remaining, 0)

	// Validate exactly one lobby was created
	keys, err := rdb.Keys(ctx, "lobby:*").Result()
	assert.NoError(t, err)
	assert.Len(t, keys, 1)

	// Validate lobby contents
	lobbyData, err := rdb.Get(ctx, keys[0]).Result()
	assert.NoError(t, err)

	var lobby Lobby
	err = json.Unmarshal([]byte(lobbyData), &lobby)
	assert.NoError(t, err)
	assert.Len(t, lobby.PlayerIDs, 5)
}
func TestSelectPlayersForLobby(t *testing.T) {
	players := []*Player{
		{PlayerID: "p1", SkillRating: 1400},
		{PlayerID: "p2", SkillRating: 1410},
		{PlayerID: "p3", SkillRating: 1420},
		{PlayerID: "p4", SkillRating: 1430},
		{PlayerID: "p5", SkillRating: 1440},
		{PlayerID: "p6", SkillRating: 1600}, // Outlier
	}

	selected := SelectPlayersForLobby(players, 5, 100)

	assert.NotNil(t, selected)
	assert.Len(t, selected, 5)
	assert.Equal(t, "p1", selected[0].PlayerID)
	assert.Equal(t, "p5", selected[4].PlayerID)

	// Test no match
	selectedNone := SelectPlayersForLobby(players, 5, 10)
	assert.Nil(t, selectedNone)
}
func TestSelectPlayersForLobbyWithOutliers(t *testing.T) {
	players := []*Player{
		{PlayerID: "p1", SkillRating: 1400},
		{PlayerID: "p2", SkillRating: 1410},
		{PlayerID: "p3", SkillRating: 1420},
		{PlayerID: "p4", SkillRating: 1430},
		{PlayerID: "p5", SkillRating: 1440},
		{PlayerID: "p6", SkillRating: 1600}, // Outlier
	}

	selected := SelectPlayersForLobby(players, 5, 200)

	assert.NotNil(t, selected)
	assert.Len(t, selected, 5)
	assert.Equal(t, "p1", selected[0].PlayerID)
	assert.Equal(t, "p5", selected[4].PlayerID)
}

func TestNotifyPlayerLobby(t *testing.T) {
	ctx := context.Background()

	// Start Redis container
	redisC, endpoint, err := startRedisContainer(ctx)
	assert.NoError(t, err)
	defer redisC.Terminate(ctx)

	// Redis client
	rdb := redis.NewClient(&redis.Options{Addr: endpoint})
	redisClient := &RedisClient{Client: rdb, Ctx: ctx}

	// Start HTTP server in goroutine
	router := chi.NewRouter()
	router.Get("/ws/{playerID}", WebSocketHandler)
	srv := &http.Server{Addr: ":8090", Handler: router}
	go srv.ListenAndServe()
	defer srv.Shutdown(ctx)

	// Connect WebSocket client
	wsURL := "ws://localhost:8090/ws/playerA"
	ws, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	assert.NoError(t, err)
	defer ws.Close()

	// Add 5 players
	for i := 0; i < 5; i++ {
		p := &Player{
			PlayerID:    fmt.Sprintf("player%c", 'A'+i),
			SkillRating: 1400 + i*10,
		}
		err := redisClient.PushPlayerToQueue(p)
		assert.NoError(t, err)
	}

	// Run matchmaking once
	err = MatchPlayers(ctx, redisClient)
	assert.NoError(t, err)

	// Read message
	_, msg, err := ws.ReadMessage()
	assert.NoError(t, err)

	// Parse JSON
	var notif map[string]string
	err = json.Unmarshal(msg, &notif)
	assert.NoError(t, err)

	// Validate contents
	assert.Equal(t, "lobby_created", notif["type"])
	assert.NotEmpty(t, notif["lobby_id"])
}
