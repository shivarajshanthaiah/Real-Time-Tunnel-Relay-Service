# Real-Time Tunnel Relay Service

A Golang-based real-time relay service using **WebSockets** for bidirectional communication.

## Features

* WebSocket-based **real-time message relay** between Admin and Clients.
* Central **Hub** manages active connections with concurrency-safe operations.
* Supports **direct messaging** (admin → specific client) and optional **broadcast**.
* Lightweight **REST APIs** for monitoring and health checks.
* **Graceful shutdown** handling using `context` and `sync.WaitGroup`.

## Tech Stack

* **Language:** Go (Golang)
* **Framework:** Gin (for REST endpoints)
* **Libraries:**

  * `github.com/gin-gonic/gin`
  * `github.com/gorilla/websocket`
* **Concurrency Tools:** Goroutines, Channels, Mutex, WaitGroup

## WebSocket Endpoints

### Client Connection

**Endpoint:**

```
GET /ws/client?id=<client_id>
```

**Behavior:**

* Registers the client in the Hub.
* Maintains heartbeat with keep-alive messages.
* Listens for messages from Admin.
* Cleans up on disconnect.

### Admin Connection

**Endpoint:**

```
GET /ws/admin
```

**Message Format:**

```json
{
  "target": "<client_id>",
  "message": "Hello"
}
```

**Behavior:**

* Sends messages to target clients.
* Supports optional broadcast (`"target": "*"`) to all clients.

## REST API Endpoints

### `GET /clients`

Returns a list of currently connected clients.
**Example Response:**

```json
{
  "connected_clients": ["client1", "client2", "client3"]
}
```

### `GET /health`

Health check endpoint.
**Example Response:**

```json
{
  "status": "ok"
}
```

## Test the WebSocket

### Start a Client

```bash
wscat -c ws://localhost:8080/ws/client?id=client1
```

### Start an Admin

```bash
wscat -c ws://localhost:8080/ws/admin
```

Send a message from Admin:

```json
{"target": "client1", "message": "Hello Client 1!"}
```

Client receives:

```
Hello Client 1!
```

## Graceful Shutdown

* Handles `SIGINT` and `SIGTERM` signals.
* Closes all WebSocket connections.
* Waits for goroutines to finish before exiting.
