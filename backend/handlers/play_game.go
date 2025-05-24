package handlers

import (
	"backend/message"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/coder/websocket"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"time"
)

func (h *Handler) PlayGame(c echo.Context) error {
	ws, err := websocket.Accept(
		c.Response(), c.Request(), &websocket.AcceptOptions{
			Subprotocols:    message.Subprotocols,
			CompressionMode: websocket.CompressionDisabled,
		},
	)
	if err != nil {
		return err
	}
	defer ws.Close(websocket.StatusNormalClosure, "")

	token := c.Get("user").(*jwt.Token)
	claims := token.Claims.(*UserClaims)

	c.Logger().Infof("%v waiting for game", claims.Username)

	ctx := c.Request().Context()

	readResults := make(chan ReadResult, 1)
	go startReader(c, ws, readResults)
	writeResults := make(chan error, 1)
	writeRequests := make(chan WriteRequest)
	defer close(writeRequests)
	go startWriter(c, ws, writeResults, writeRequests)

	writeRequests <- WriteRequest{
		MsgType: message.TypeWaitingForGame,
		Payload: message.WaitingForGamePayload{},
	}

	client := h.GameCache.Join(claims.UserID, ws)
	defer h.GameCache.Leave(claims.UserID)

	var lobbyId uuid.UUID
findLobby:
	for {
		select {
		case rr := <-readResults:
			if rr.err != nil {
				return rr.err
			}
			c.Logger().Infof("Read message %v", rr.msg)
		case wrErr := <-writeResults:
			if wrErr != nil {
				return wrErr
			}
			panic("should not happen")
		case lobbyId = <-client.ReceiveLobbyId():
			writeRequests <- WriteRequest{
				MsgType: message.TypeFoundGame,
				Payload: message.FoundGamePayload{LobbyId: lobbyId.String()},
			}
			break findLobby
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()

		case rr := <-readResults:
			if rr.err != nil {
				var msgErr message.Error
				isMsgErr := errors.As(rr.err, &msgErr)
				if !isMsgErr {
					return rr.err
				}
				writeRequests <- WriteRequest{
					MsgType: message.TypeError, Payload: message.ErrorPayload{
						Code:           websocket.StatusUnsupportedData,
						Err:            msgErr.Error(),
						ProblematicMsg: rr.msg,
					},
				}
				break
			}

			switch rr.msg.Type {
			case message.TypeChatMessage:
				var chatMsg message.ChatMessagePayload
				err = json.Unmarshal(rr.msg.Payload, &chatMsg)
				if err != nil {
					return err
				}
				h.GameCache.Send(lobbyId, chatMsg.Text)
			default:
				errStr := fmt.Sprintf("Unknown message type '%s'", rr.msg.Type)
				c.Logger().Info(errStr)
				writeRequests <- WriteRequest{
					MsgType: message.TypeError, Payload: message.ErrorPayload{
						Code:           websocket.StatusUnsupportedData,
						Err:            errStr,
						ProblematicMsg: rr.msg,
					},
				}
			}

		case wrErr := <-writeResults:
			if wrErr != nil {
				return wrErr
			}

		case msg := <-client.ReceiveMsg():
			writeRequests <- WriteRequest{
				MsgType: message.TypeChatMessage,
				Payload: message.ChatMessagePayload{
					From: "",
					Text: string(msg),
				},
			}
		}
	}
}

type ReadRequest struct {
	IgnoreTimeout bool
}

type ReadResult struct {
	msg message.Message
	err error
}

func startReader(c echo.Context, ws *websocket.Conn, readResults chan<- ReadResult) {
	ctx := c.Request().Context()
	defer close(readResults)
	for {
		select {
		case <-ctx.Done():
			readResults <- ReadResult{message.Nil, ctx.Err()}
			return
		default:
			msg, err := read(ctx, ws)
			if err != nil {
				readResults <- ReadResult{message.Nil, err}
				return
			}
			readResults <- ReadResult{msg, err}
		}
	}
}

func read(ctx context.Context, ws *websocket.Conn) (message.Message, error) {
	_, r, err := ws.Reader(ctx)
	if err != nil {
		return message.Nil, err
	}
	msg, err := message.JsonDecodeBaseMsg(json.NewDecoder(r))
	if err != nil {
		return message.Nil, err
	}

	return msg, nil
}

type WriteRequest struct {
	MsgType string
	Payload any
	Timeout time.Duration
}

func startWriter(c echo.Context, ws *websocket.Conn, writeResults chan<- error, writeRequests <-chan WriteRequest) {
	ctx := c.Request().Context()
	defer close(writeResults)
	for {
		select {
		case <-ctx.Done():
			writeResults <- ctx.Err()
			return
		case wr := <-writeRequests:
			err := writeTimeout(ctx, ws, wr)
			writeResults <- err
			if err != nil {
				return
			}
		}
	}
}

func writeTimeout(ctx context.Context, ws *websocket.Conn, wr WriteRequest) error {
	timeout := wr.Timeout
	if timeout <= 0 {
		timeout = time.Second
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	msg, err := message.NewV1(wr.MsgType, wr.Payload)
	if err != nil {
		return err
	}
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	err = ws.Write(ctx, websocket.MessageText, data)
	if err != nil {
		return err
	}

	return nil
}
