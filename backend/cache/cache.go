package cache

import (
	"backend/game"
	"backend/message"
	"backend/websockets"
	"context"
	"encoding/json"
	"github.com/coder/websocket"
	"github.com/google/uuid"
	"sync"
	"time"
)

type Cache struct {
	mutex         sync.RWMutex
	connections   map[uuid.UUID]*Client
	readyPlayersQ chan uuid.UUID
	idleLobbies   map[uuid.UUID]*Lobby
	inGameLobbies map[uuid.UUID]*Lobby
}

type Lobby struct {
	Id        uuid.UUID
	players   map[uuid.UUID]PlayerInfo
	broadcast chan websockets.WriteRequest
	Game      *game.Game
	moves     chan game.Move
	Messages  []message.ChatMessagePayload
}

type PlayerInfo struct {
	Color game.Color
}

func NewLobby() (*Lobby, error) {
	lobbyId, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}
	g, err := game.New()
	if err != nil {
		return nil, err
	}

	return &Lobby{
		Id:        lobbyId,
		players:   make(map[uuid.UUID]PlayerInfo),
		broadcast: make(chan websockets.WriteRequest),
		Game:      g,
		moves:     make(chan game.Move),
		Messages:  make([]message.ChatMessagePayload, 0),
	}, nil
}

type Client struct {
	Id            uuid.UUID
	Socket        *websocket.Conn
	Notify        chan *Lobby
	WriteRequests chan websockets.WriteRequest
	Unregister    chan struct{}
}

func NewClient(playerId uuid.UUID, ws *websocket.Conn) *Client {
	return &Client{
		Id:            playerId,
		Socket:        ws,
		Notify:        make(chan *Lobby, 1),
		WriteRequests: make(chan websockets.WriteRequest),
		Unregister:    make(chan struct{}),
	}
}

func NewDefaultCache() *Cache {
	return &Cache{
		connections:   make(map[uuid.UUID]*Client),
		readyPlayersQ: make(chan uuid.UUID, 100),
		idleLobbies:   make(map[uuid.UUID]*Lobby),
		inGameLobbies: make(map[uuid.UUID]*Lobby),
	}
}

func (gc *Cache) RunMatchmaking(ctx context.Context) {
	lobby, err := NewLobby()
	if err != nil {
		panic(err)
	}
	gc.idleLobbies[lobby.Id] = lobby

	for {
		select {
		case <-ctx.Done():
			return
		case rpId := <-gc.readyPlayersQ:
			gc.mutex.RLock()
			_, connected := gc.connections[rpId]
			if !connected {
				gc.mutex.RUnlock()
				continue
			}
			lobby.players[rpId] = PlayerInfo{Color: game.Color(len(lobby.players) + 1)}
			if len(lobby.players) < 2 {
				gc.mutex.RUnlock()
				continue
			}

			go gc.startGame(ctx, lobby.Id)

			delete(gc.idleLobbies, lobby.Id)
			gc.inGameLobbies[lobby.Id] = lobby
			for pId := range lobby.players {
				gc.connections[pId].Notify <- lobby
			}
			gc.mutex.RUnlock()

			lobby, err = NewLobby()
			if err != nil {
				panic(err)
			}
			gc.idleLobbies[lobby.Id] = lobby
		}
	}
}

func (gc *Cache) Join(playerId uuid.UUID, ws *websocket.Conn) *Client {
	c := NewClient(playerId, ws)
	gc.mutex.Lock()
	gc.connections[playerId] = c
	gc.mutex.Unlock()

	for _, lobby := range gc.inGameLobbies {
		for pId := range lobby.players {
			if playerId == pId {
				c.Notify <- lobby
				return c
			}
		}
	}

	gc.readyPlayersQ <- c.Id

	return c
}

func (gc *Cache) PlayerInfo(lobbyId uuid.UUID, playerId uuid.UUID) PlayerInfo {
	return gc.inGameLobbies[lobbyId].players[playerId]
}

func (gc *Cache) Leave(playerId uuid.UUID) {
	gc.mutex.Lock()
	defer gc.mutex.Unlock()
	if client, exists := gc.connections[playerId]; exists {
		close(client.Notify)
		close(client.WriteRequests)
	}

	delete(gc.connections, playerId)
}

func (gc *Cache) Send(lobbyId uuid.UUID, wr websockets.WriteRequest) {
	gc.inGameLobbies[lobbyId].broadcast <- wr
}

func (gc *Cache) Play(lobbyId uuid.UUID, playerId uuid.UUID, column uint8) (int, bool, error) {
	lobby := gc.inGameLobbies[lobbyId]
	move := game.Move{
		Column: column,
		Color:  lobby.players[playerId].Color,
	}
	return lobby.Game.Make(move)
}

func (gc *Cache) startGame(ctx context.Context, lobbyId uuid.UUID) {
	lobby := gc.inGameLobbies[lobbyId]
	for {
		select {
		case <-ctx.Done():
			return
		case wr := <-lobby.broadcast:
			for pId := range lobby.players {
				gc.mutex.RLock()
				c, connected := gc.connections[pId]
				gc.mutex.RUnlock()
				if !connected {
					continue
				}
				c.WriteRequests <- wr
				time.Sleep(100 * time.Millisecond)
				if wr.MsgType == message.TypeGameOver {
					c.Unregister <- struct{}{}
				}
			}
			if wr.MsgType == message.TypeGameOver {
				delete(gc.inGameLobbies, lobbyId)
			}
			if wr.MsgType == message.TypeChat {
				var msg message.ChatMessagePayload
				_ = json.Unmarshal(wr.Payload.(json.RawMessage), &msg)
				lobby.Messages = append(lobby.Messages, msg)
			}
		}
	}
}
