package message

import (
	"encoding/json"
	"fmt"
	"github.com/coder/websocket"
)

const v1 = "v1"

var (
	Subprotocols = []string{"json.v1"}
	Nil          = Message{
		Version: v1,
		Type:    "nil",
		Payload: nil,
	}
)

type Message struct {
	Version string          `json:"version"`
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

type Error struct {
	Code           websocket.StatusCode `json:"code,omitempty"`
	Err            string               `json:"err,omitempty"`
	ProblematicMsg any                  `json:"problematicMsg"`
}

func (e Error) Error() string {
	return fmt.Sprintf("message error %d: %s %v", e.Code, e.Err, e.ProblematicMsg)
}

const TypeError = "error"
type ErrorPayload struct {
	Code           websocket.StatusCode `json:"code"`
	Err            string               `json:"err,omitempty"`
	ProblematicMsg Message              `json:"problematicMsg"`
}

func JsonDecodeBaseMsg(decoder *json.Decoder) (Message, error) {
	var msg Message
	if err := decoder.Decode(&msg); err != nil {
		return msg, err
	}

	if msg.Version == "" {
		return Nil, Error{
			Code: websocket.StatusInvalidFramePayloadData,
			Err:  "message 'version' empty or missing",
		}
	} else if msg.Type == "" {
		return Nil, Error{
			Code: websocket.StatusInvalidFramePayloadData,
			Err:  "message 'type' empty or missing",
		}
	}

	return msg, nil
}

func NewV1[T any](typ string, payload T) (Message, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return Nil, nil
	}
	return Message{Version: v1, Type: typ, Payload: data}, nil
}

const TypeWaitingForGame = "waitingForGame"
type WaitingForGamePayload struct{}

const TypeFoundGame = "foundGame"
type FoundGamePayload struct {
	LobbyId string `json:"lobbyId"`
}

const TypeChatMessage = "chatMessage"
type ChatMessagePayload struct {
	From string `json:"from"`
	Text string `json:"text"`
}
