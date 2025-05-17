package handlers

import (
	"backend/generated/sqlc"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"net/http"
	"time"
)

type CreateLobbyRequest struct {
	CreatorId string `json:"creatorId,omitempty" validate:"required,uuid"`
}

type CreateLobbyResponse struct {
	LobbyId string `json:"lobbyId"`
}

func (h *Handler) CreateLobby(c echo.Context) error {
	var request CreateLobbyRequest
	if err := c.Bind(&request); err != nil {
		return err
	}
	if err := c.Validate(request); err != nil {
		return err
	}

	ctx := c.Request().Context()
	tx, err := h.Conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	qtx := h.DB.WithTx(tx)

	p1Id, err := uuid.Parse(request.CreatorId)
	if err != nil {
		return err
	}

	_, notFound := qtx.GetUserById(ctx, p1Id)
	if notFound == nil {
		return echo.NewHTTPError(http.StatusConflict, "user already has an active lobby")
	}

	lobbyId, err := uuid.NewV7()
	if err != nil {
		return err
	}
	err = qtx.CreateLobby(ctx, sqlc.CreateLobbyParams{
		ID:           lobbyId,
		Player1ID:    p1Id,
		CreatedAtUtc: time.Now().UTC(),
		IsPrivate:    true,
	})
	if err != nil {
		return err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, CreateLobbyResponse{lobbyId.String()})
}
