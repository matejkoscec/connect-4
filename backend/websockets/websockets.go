package websockets

import (
	"backend/message"
	"context"
	"encoding/json"
	"github.com/coder/websocket"
	"github.com/labstack/echo/v4"
	"time"
)

type ReadRequest struct {
	IgnoreTimeout bool
}

type ReadResult struct {
	Msg message.Message
	Err error
}

func StartReader(c echo.Context, ws *websocket.Conn, readResults chan<- ReadResult) {
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
}

func StartWriter(c echo.Context, ws *websocket.Conn, writeResults chan<- error, writeRequests <-chan WriteRequest) {
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
	timeout := time.Second
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
