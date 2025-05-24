package game

import (
	"context"
	"fmt"
	"github.com/coder/websocket"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"sync"
)

type Cache struct {
	mutex         sync.RWMutex
	connections   map[uuid.UUID]*Client
	readyPlayersQ chan uuid.UUID
	idleLobbies   map[uuid.UUID]*Lobby
	inGameLobbies map[uuid.UUID]*Lobby

	logger echo.Logger
}

type Lobby struct {
	id        uuid.UUID
	players   map[uuid.UUID]struct{}
	broadcast chan []byte
}

type Client struct {
	id     uuid.UUID
	socket *websocket.Conn
	notify chan uuid.UUID
	send   chan []byte
}

func (c *Client) ReceiveLobbyId() chan uuid.UUID {
	return c.notify
}

func (c *Client) ReceiveMsg() chan []byte {
	return c.send
}

func NewDefaultCache(logger echo.Logger) *Cache {
	return &Cache{
		connections:   make(map[uuid.UUID]*Client),
		readyPlayersQ: make(chan uuid.UUID, 100),
		idleLobbies:   make(map[uuid.UUID]*Lobby),
		inGameLobbies: make(map[uuid.UUID]*Lobby),
		logger:        logger,
	}
}

func (gc *Cache) RunMatchmaking(ctx context.Context) {
	lobbyId, err := uuid.NewV7()
	if err != nil {
		panic(err)
	}
	lobby := &Lobby{
		id:        lobbyId,
		players:   make(map[uuid.UUID]struct{}),
		broadcast: make(chan []byte),
	}
	gc.idleLobbies[lobbyId] = lobby

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
			lobby.players[rpId] = struct{}{}
			if len(lobby.players) < 2 {
				gc.mutex.RUnlock()
				continue
			}

			delete(gc.idleLobbies, lobbyId)
			gc.inGameLobbies[lobbyId] = lobby
			for pId := range lobby.players {
				gc.connections[pId].notify <- lobbyId
			}
			gc.mutex.RUnlock()

			gc.StartChat(ctx, lobbyId)

			lobbyId, err = uuid.NewV7()
			if err != nil {
				panic(err)
			}
			lobby = &Lobby{
				id:        lobbyId,
				players:   make(map[uuid.UUID]struct{}),
				broadcast: make(chan []byte),
			}
			gc.idleLobbies[lobbyId] = lobby
		}
	}
}

func (gc *Cache) Join(playerId uuid.UUID, ws *websocket.Conn) *Client {
	c := &Client{
		id:     playerId,
		socket: ws,
		notify: make(chan uuid.UUID, 1),
		send:   make(chan []byte),
	}
	gc.mutex.Lock()
	gc.connections[playerId] = c
	gc.mutex.Unlock()

	for lobbyId, lobby := range gc.inGameLobbies {
		for pId := range lobby.players {
			if playerId == pId {
				c.notify <- lobbyId
				return c
			}
		}
	}

	gc.readyPlayersQ <- c.id

	return c
}

func (gc *Cache) Leave(playerId uuid.UUID) {
	gc.mutex.Lock()
	if client, exists := gc.connections[playerId]; exists {
		close(client.notify)
		close(client.send)
	}

	delete(gc.connections, playerId)
	fmt.Printf("connections: %+v\nin game: %+v\n", gc.connections, gc.inGameLobbies)
	gc.mutex.Unlock()
}

func (gc *Cache) Send(lobbyId uuid.UUID, message string) {
	gc.inGameLobbies[lobbyId].broadcast <- []byte(message)
}

func (gc *Cache) StartChat(ctx context.Context, lobbyId uuid.UUID) {
	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-gc.inGameLobbies[lobbyId].broadcast:
			for pId := range gc.inGameLobbies[lobbyId].players {
				gc.mutex.RLock()
				c, connected := gc.connections[pId]
				gc.mutex.RUnlock()
				if !connected {
					continue
				}
				c.send <- msg
			}
		}
	}
}
