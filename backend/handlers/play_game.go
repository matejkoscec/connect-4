package handlers

import (
	"backend/message"
	"backend/websockets"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/coder/websocket"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
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

	client := h.GameCache.Join(claims.UserID, ws)
	defer h.GameCache.Leave(claims.UserID)

	readResults := make(chan websockets.ReadResult, 1)
	go websockets.StartReader(c, ws, readResults)
	writeResults := make(chan error, 1)
	writeRequests := client.WriteRequests
	go websockets.StartWriter(c, ws, writeResults, writeRequests)

	writeRequests <- websockets.WriteRequest{
		MsgType: message.TypeWaitingForGame,
		Payload: message.WaitingForGamePayload{},
	}

	var lobbyId uuid.UUID
findLobby:
	for {
		select {
		case rr := <-readResults:
			if rr.Err != nil {
				return rr.Err
			}
			c.Logger().Infof("Read message %v", rr.Msg)
		case wrErr := <-writeResults:
			if wrErr != nil {
				return wrErr
			}
			c.Logger().Info("Written message")
		case lobbyId = <-client.Notify:
			writeRequests <- websockets.WriteRequest{
				MsgType: message.TypeFoundGame,
				Payload: message.FoundGamePayload{LobbyId: lobbyId.String()},
			}
			break findLobby
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	playerInfo := h.GameCache.PlayerInfo(lobbyId, claims.UserID)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()

		case rr := <-readResults:
			if rr.Err != nil {
				var msgErr message.Error
				isMsgErr := errors.As(rr.Err, &msgErr)
				if !isMsgErr {
					return rr.Err
				}
				writeRequests <- websockets.WriteRequest{
					MsgType: message.TypeError, Payload: message.ErrorPayload{
						Code:           websocket.StatusUnsupportedData,
						Err:            msgErr.Error(),
						ProblematicMsg: rr.Msg,
					},
				}
				break
			}

			switch rr.Msg.Type {
			case message.TypeChat:
				var chatMsg message.ChatMessagePayload
				err = json.Unmarshal(rr.Msg.Payload, &chatMsg)
				if err != nil {
					return err
				}
				h.GameCache.Send(
					lobbyId, websockets.WriteRequest{
						MsgType: rr.Msg.Type,
						Payload: rr.Msg.Payload,
					},
				)

			case message.TypePlayMove:
				var moveMsg message.PlayMovePayload
				err = json.Unmarshal(rr.Msg.Payload, &moveMsg)
				if err != nil {
					return err
				}
				isWinningMove, err := h.GameCache.Play(lobbyId, claims.UserID, moveMsg.Column)
				if err != nil {
					writeRequests <- websockets.WriteRequest{
						MsgType: message.TypeError, Payload: message.ErrorPayload{
							Code:           websocket.StatusUnsupportedData,
							Err:            err.Error(),
							ProblematicMsg: rr.Msg,
						},
					}
					break
				}

				writeRequests <- websockets.WriteRequest{
					MsgType: message.TypePlayedMove,
					Payload: message.PlayedMovePayload{
						Color:  playerInfo.Color,
						Column: moveMsg.Column,
					},
				}

				if isWinningMove {
					h.GameCache.Send(
						lobbyId, websockets.WriteRequest{
							MsgType: message.TypeGameOver,
							Payload: message.GameOverPayload{Winner: playerInfo.Color},
						},
					)
				}

			default:
				errStr := fmt.Sprintf("Unknown message type '%s'", rr.Msg.Type)
				c.Logger().Info(errStr)
				writeRequests <- websockets.WriteRequest{
					MsgType: message.TypeError, Payload: message.ErrorPayload{
						Code:           websocket.StatusUnsupportedData,
						Err:            errStr,
						ProblematicMsg: rr.Msg,
					},
				}
			}

		case wrErr := <-writeResults:
			if wrErr != nil {
				return wrErr
			}

		case <-client.Unregister:
			return nil
		}
	}
}
