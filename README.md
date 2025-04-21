# ♚ ShahBoard

ShahBoard is a modern, scalable chess platform built with Go — designed for real-time multiplayer games, live game watching, and bot battles. It uses a microservices architecture and event-driven design to ensure high performance, flexibility, and easy scaling.


## 🚀 Features
- 🔄 Real-time gameplay via WebSocket  
- 💬 Real-time chat for game players  
- 📱 Multi-device session sync  
- 👀 Live Watching with viewer counts   
- 🔐 Secure Google OAuth login

<br><br>
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

<br><br>
## ♟ Game Lifecycle Flow

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

<br><br>

## ✅ Things Done So Far

- Set up project structure with Docker, config, and logging
- Built core services:
  - **Auth Service** with Google OAuth login
  - **Match Service** for matchmaking
  - **Game Service** for game creation, move validation, and state tracking
  - **WS Gateway** for WebSocket connections and real-time updates
  - **Chat Service** for in-game messaging
  - **Profile Service** for ELO and account data
- Integrated **Kafka** for event-driven architecture
- Designed and documented **game lifecycle flow** (matchmaking → gameplay → game end)
- Enabled **live game watching** with viewers list
- WebSocket event relays and client subscription management


## 🧩 Next Steps

1. **🧱 Error Package**
   - Create a shared `error` package
   - Define standard error types with codes (e.g. `ErrUnauthorized`, `ErrInvalidMove`)
   - Add helper funcs for wrapping, logging, and translating to gRPC/HTTP errors

2. **🏆 Leaderboard**
   - Track top players based on ELO or win stats
   - Endpoint/service to fetch global or regional rankings
   - Cache leaderboard with Redis for performance

3. **🔔 Notification Service**
   - Event-based notification system (e.g., “You’ve been challenged” or “Game started”)
   - WebSocket or push message support

4. **🧑‍🤝‍🧑 Invite Feature**
   - Let players directly invite others to a game
   - Support pending invite states
   - Event: `invite.sent`, `invite.accepted`, `invite.declined`
   - Use Notification Service to deliver invites via WebSocket
   - Integrate with Match and Game services on acceptance

5. **💬 Direct Message Feature**
   - Implement a private messaging system between users
   - Enable real-time chat between users (outside of games)
   - Support persistent message history
   - Integrate with Notification Service to alert users of new messages

6. **📥 Idempotent Kafka Subscribers**  
   - Add unique event IDs to all Events  
   - Store processed event IDs in-memory with TTL (e.g., 10 minutes) to avoid duplicates  
   - Use a concurrent map with expiration 
   - Ensure consumers (e.g., WS Gateway, Chat, Profile) are idempotent 

7. **📤 Reliable & Async Kafka Publisher**  
   - Make Kafka publishing asynchronous to decouple it from service layer  
   - Store unacknowledged events in a bounded in-memory queue  
   - When memory limit is reached, offload extra events to Redis (e.g., pending:events)
   - Implement retry logic with exponential backoff or fixed intervals
   - Remove event from queue only after Kafka acknowledgment

8. **⏱️ Time Control Support**
   - Add time control types: blitz, rapid, classical, custom
   - Modify game creation to accept time settings
   - Add time countdown logic in Game Service

9. **📽️ Game Replay**
   - Store full move history of each game
   - Build frontend playback system (step-through, autoplay, etc.)
   - API to fetch PGN-style game history

10. **📱 Mobile-Compatible Frontend**
   - Make the frontend responsive and optimized for phones
