# Real-Time Multiplayer Tic-Tac-Toe

A high-performance, concurrent, real-time Tic-Tac-Toe engine built with Go, MongoDB, and WebSockets. This project supports multiple simultaneous game rooms, atomic state management, and an automated forfeit system for player disconnections.

## 🚀 Features

- **Real-Time Gameplay**: Low-latency communication using Gorilla WebSockets.
- **Multi-Room Support**: Concurrent game sessions identified by unique MongoDB ObjectIDs.
- **Atomic Persistence**: Uses MongoDB's `FindOneAndUpdate` to ensure the "Source of Truth" is never compromised by race conditions.
- **Concurrency Safety**: Implements `sync.Mutex` with optimized critical sections to handle multiple players joining or leaving simultaneously.
- **Automated Matchmaking**: A custom lobby system (`gameChecker`) that assigns X and O roles dynamically.
- **Forfeit System**: Automatically detects disconnections and declares the remaining player the winner to ensure fair play.
- **Live Reloading**: Configured for use with `air` for a seamless development experience.

## 🛠 Tech Stack

- **Language**: Go (Golang)
- **Database**: MongoDB
- **Communication**: WebSockets (Gorilla)
- **Concurrency Control**: `sync.Mutex` and Snapshot Patterns

## 📋 Prerequisites

- Go 1.22+ 
- MongoDB Atlas account or local instance

## ⚙️ Environment Variables

Create a `.env` file or set the following environment variable to secure your database credentials:

```bash
MONGO_URI=mongodb+srv://<username>:<password>@cluster.mongodb.net/?appName=Cluster
```

## 🏃 Getting Started

1. **Clone the repository**
   ```bash
   git clone <your-repo-url>
   cd tictactoe-backend
   ```

2. **Install dependencies**
   ```bash
   go mod tidy
   ```

3. **Run**
   ```bash
   go run .
   ```
   *The server will start at `http://localhost:8080`*

## 🔌 API Endpoints

### 1. Initialize Game
- **Endpoint**: `GET /start`
- **Description**: Creates a new game document in MongoDB and returns the initial state with a unique `gameid`.

### 2. WebSocket Connection
- **Endpoint**: `WS /ws/{id}?player={playerID}`
- **Description**: Connects to a specific game room. 
- **Lobby Logic**:
    - First to join is assigned **Player X**.
    - Second to join is assigned **Player O**.
    - Subsequent connections receive a "Game Full" error.

## 🎮 Game Logic

### State Machine
The core gameplay is handled by the `play()` function, which acts as a pure state machine:
1. **Validate**: Checks for valid turn, occupied spots, and existing win status.
2. **Apply**: Updates the 2D board array and increments the turn counter.
3. **Analyze**: Runs `checkWin` and `checkTie` algorithms.
4. **Persist**: Writes the atomic update to MongoDB and returns the updated state.

### Disconnection Handling
The `endConnection` function utilizes a **Snapshot Pattern**:
- When a player leaves, a snapshot of the remaining connections is taken.
- The mutex is released immediately to prevent blocking other rooms.
- The `deletePlayer` logic updates the DB to reflect a forfeit win for the stayer.
- The final state is broadcasted to all remaining participants in that specific room.

## 🗂 Project Structure

- `main.go`: Entry point, server configuration, and MongoDB initialization.
- `game.go`: Contains `/start` and all the functions for the game logic.
- `ws.go`: This has the websocket connection, which handles the game