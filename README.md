# ♚ ShahBoard

ShahBoard is a modern, scalable chess platform built with Go — designed for real-time multiplayer games, live game watching, and bot battles. It uses a microservices architecture and event-driven design to ensure high performance, flexibility, and easy scaling.

---

## 🚀 Features
- 🔄 Real-time gameplay via WebSocket  
- 💬 Real-time chat for game players  
- 📱 Multi-device session sync  
- 👀 Live Watching with viewer counts   
- 🔐 Secure Google OAuth login

---

## 🏗️ Architecture Overview

ShahBoard is built on a **microservices architecture** with an **event-driven design**, ensuring high scalability, modularity, and real-time responsiveness. Services communicate using **Kafka** as a central event bus, while WebSocket connections provide interactive game and chat experiences. Services also interact through **gRPC** for synchronous operations.

---

### 🔁 Traefik — Reverse Proxy
- Entry point for all HTTP and WebSocket traffic.
- Routes requests to services dynamically.
- Handles load balancing, TLS termination, and routing rules.

---

### 🔐 Auth Service
- Supports both **Google OAuth** and **email/password-based** login.
- Issues JWTs for authentication and authorization.
- Manages user identity, sessions, and token refresh.

---

### 🎯 Match Service
- Handles **matchmaking** by finding opponents with similar skill levels.
- Performs gRPC checks with Game Service to avoid duplicate games.
- Publishes `match.created` events to Kafka when a match is found.

---

### ♟ Game Service
- Manages **game creation**, **state tracking**, and **move validation**.
- Maintains list of **live games** and **player status**.
- Could be split into a separate **Live Game Service** in the future with recommendation algorithms.
- Emits events like `game.created`, `game.moveApproved`, and `game.ended`.

---

### 💬 Chat Service
- Creates a **chat room per game**.
- Enables **real-time messaging** between players.
- Listens to Kafka events like `game.created` and `game.ended`.

---

### 👤 Profile Service  
*(Planned to split into `User Service` and `Rating Service`)*
- Maintains player profiles and account data.
- Updates and tracks **ELO ratings** after each game.
- Listens to `game.ended` events.

---

### 🌐 WS Gateway (WebSocket Gateway)
- Manages all player WebSocket connections.
- Converts WebSocket messages (moves, chat) into Kafka events.
- Subscribes to Kafka events and relays updates to clients.
- Ensures real-time experience for players and viewers.

---

### 📨 Kafka — Event Bus
- Central event broker for inter-service communication.
- Handles high-throughput, fault-tolerant messaging between services.

---

## ♟ Game Lifecycle Flow

---

## 🔍 Phase 1: Matchmaking and Game Creation

This phase begins when a player initiates a game request and ends once both players are connected to the same game via WebSocket.

### 🔄 Flow Breakdown

1. **Client → Match Request**  
   - Sends HTTP `find-match` to **Match Service**.

2. **Game Check (gRPC)**  
   - Match Service checks with **Game Service** if the player is already in a game.

3. **Queueing**  
   - Eligible players are added to match-engine queue.

4. **Match Found**  
   - Responds with `matchId` and `opponentId`.
   - Publishes `match.created` event to Kafka.

5. **Game Creation**  
   - **Game Service** consumes `match.created`:
     - Creates game
     - Assigns `gameId` and colors
     - Publishes `game.created` event

6. **Event Consumers**
   - **Chat Service** creates a chat room
   - **WS Gateway** prepares WebSocket channel

7. **Client → WebSocket**
   - Client connects to WS Gateway and sends `matchId`.

8. **WebSocket Subscription**
   - WS Gateway listens for matching `game.created` event and subscribes session.

### 🧰 Component Summary
| Component       | Responsibility                                       |
|----------------|-------------------------------------------------------|
| Match Service   | Match players, publish `match.created`               |
| Game Service    | Create games, assign colors, publish `game.created`  |
| Chat Service    | Initialize in-game chat                              |
| WS Gateway      | Manage sessions and subscriptions                    |
| Kafka           | Transport events                                     |
| gRPC            | Used for game participation checks                   |

---

## 🎮 Phase 2: Game Playing and Move Validation

This phase involves players making moves, validating them, and updating everyone in real-time.

### 🔄 Flow Breakdown

1. **Player → WebSocket Move**
2. **WS Gateway → Kafka**
   - Validates `gameId`, `playerId`
   - Publishes `game.playerMoved`

3. **Game Service → Move Validation**
   - Validates move and turn
   - Updates game state
   - Publishes `game.moveApproved`

4. **WS Gateway → Broadcast**
   - 🔍 Converts and sends update to all subscribed clients (players/viewers)

### 🧰 Component Summary
| Component       | Responsibility                                               |
|----------------|---------------------------------------------------------------|
| WS Gateway      | Validate & relay move events                                  |
| Game Service    | Validate moves, update game state                             |
| Kafka           | Relay events like `game.playerMoved` and `game.moveApproved`  |
| WebSocket       | Deliver updates to clients                                    |

---

## 🏁 Phase 3: Game End and Rating Updates

### 🔄 Flow Breakdown

1. **Trigger**
   - Last move, resign, draw, or inactivity (2 minutes)

2. **Game Service → End Game**
   - Publishes `game.ended` to Kafka

3. **WS Gateway → Notify & Unsubscribe**
4. **Chat Service → Close Room**
5. **Profile Service → Update ELO**
   - Record rating change history

### 🧰 Component Summary
| Component       | Responsibility                                     |
|----------------|-----------------------------------------------------|
| Game Service    | End game, publish `game.ended`                     |
| WS Gateway      | Notify and clean up subscriptions                  |
| Chat Service    | Close in-game chat                                 |
| Profile Service | Update ratings and history                         |
| Kafka           | Event transport                                    |

---

