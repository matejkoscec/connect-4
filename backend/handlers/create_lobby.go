package handlers

import (
	"backend/generated/sqlc"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"net/http"
	"strings"
	"time"
)

type CreateLobbyRequest struct {
	CreatorId string `json:"creatorId,omitempty" validate:"required,uuid"`
}

type CreateLobbyResponse struct {
	LobbyId string
	GameId  string
}

func (h *Handler) CreateLobby(c echo.Context) error {
	var request CreateLobbyRequest
	if err := c.Bind(&request); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error()).SetInternal(err)
	}
	if err := c.Validate(request); err != nil {
		return err
	}

	lobbyId, err := uuid.NewV7()
	if err != nil {
		return err
	}
	gameId, err := uuid.NewV7()
	if err != nil {
		return err
	}
	p1Id, err := uuid.Parse(request.CreatorId)
	if err != nil {
		return err
	}
	utc := time.Now().UTC()

	ctx := c.Request().Context()
	tx, err := h.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	qtx := h.DB.WithTx(tx)

	err = qtx.CreateLobby(ctx, sqlc.CreateLobbyParams{
		ID:        lobbyId,
		Player1ID: p1Id,
		CreatedAtUtc: utc,
	})
	if err != nil {
		return err
	}

	err = qtx.CreateGame(ctx, sqlc.CreateGameParams{
		ID:      [16]byte{},
		LobbyID: lobbyId,
		State: strings.Repeat("0", 6*7),
	})
	if err != nil {
		return err
	}

	err = tx.Commit(c.Request().Context())
	if err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, CreateLobbyResponse{
		LobbyId: lobbyId.String(),
		GameId:  gameId.String(),
	})
}
