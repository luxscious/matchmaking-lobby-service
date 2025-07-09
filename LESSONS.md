## Websocket Connection Handling #39

✅ How to upgrade HTTP to WebSocket
✅ How to store live connections safely
✅ How to clean up on disconnect
✅ How to notify clients with structured JSON

## REST Endpoint: Join Queue #40

✅ How to push structured JSON data into Redis
✅ How to create composable handlers with dependency injection (func(redisClient) http.HandlerFunc)
✅ How to test and confirm correct queueing

## Matchmaking Loop #41

✅ How to design a continuous background loop that processes queued players
✅ How to separate orchestration (looping) from pure selection logic
✅ How to atomically remove matched players from Redis using transactions
✅ How to create lobbies and persist them in Redis with clean JSON structures
✅ How to send real-time notifications to WebSocket clients when a lobby is created
✅ How to test deterministic matchmaking behavior using isolated Redis containers (testcontainers-go)
