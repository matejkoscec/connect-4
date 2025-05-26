package cache

import (
	"backend/game"
	"backend/message"
	"backend/websockets"
	"context"
	"fmt"
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
	id        uuid.UUID
	players   map[uuid.UUID]PlayerInfo
	broadcast chan websockets.WriteRequest
	game      *game.Game
	moves     chan game.Move
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
		id:        lobbyId,
		players:   make(map[uuid.UUID]PlayerInfo),
		broadcast: make(chan websockets.WriteRequest),
		game:      g,
		moves:     make(chan game.Move),
	}, nil
}

type Client struct {
	Id            uuid.UUID
	Socket        *websocket.Conn
	Notify        chan uuid.UUID
	WriteRequests chan websockets.WriteRequest
	Unregister    chan struct{}
}

func NewClient(playerId uuid.UUID, ws *websocket.Conn) *Client {
	return &Client{
		Id:            playerId,
		Socket:        ws,
		Notify:        make(chan uuid.UUID, 1),
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
	gc.idleLobbies[lobby.id] = lobby

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

			delete(gc.idleLobbies, lobby.id)
			gc.inGameLobbies[lobby.id] = lobby
			for pId := range lobby.players {
				gc.connections[pId].Notify <- lobby.id
			}
			gc.mutex.RUnlock()

			gc.startGame(ctx, lobby.id)

			lobby, err = NewLobby()
			if err != nil {
				panic(err)
			}
			gc.idleLobbies[lobby.id] = lobby
		}
	}
}

func (gc *Cache) Join(playerId uuid.UUID, ws *websocket.Conn) *Client {
	c := NewClient(playerId, ws)
	gc.mutex.Lock()
	gc.connections[playerId] = c
	gc.mutex.Unlock()

	for lobbyId, lobby := range gc.inGameLobbies {
		for pId := range lobby.players {
			if playerId == pId {
				c.Notify <- lobbyId
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
	if client, exists := gc.connections[playerId]; exists {
		close(client.Notify)
		close(client.WriteRequests)
	}

	delete(gc.connections, playerId)
	fmt.Printf("connections: %+v\nin game: %+v\n", gc.connections, gc.inGameLobbies)
	gc.mutex.Unlock()
}

func (gc *Cache) Send(lobbyId uuid.UUID, wr websockets.WriteRequest) {
	gc.inGameLobbies[lobbyId].broadcast <- wr
}

func (gc *Cache) Play(lobbyId uuid.UUID, playerId uuid.UUID, column uint8) (bool, error) {
	lobby := gc.inGameLobbies[lobbyId]
	move := game.Move{
		Column: column,
		Color:  lobby.players[playerId].Color,
	}
	return lobby.game.Make(move)
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
		}
	}
}
