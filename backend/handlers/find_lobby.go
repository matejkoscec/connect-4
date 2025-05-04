package handlers

import (
	"backend/generated/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/labstack/echo/v4"
	"net/http"
	"time"
)

type findLobbyRequest struct {
	PlayerId string `json:"playerId,omitempty" validate:"required,uuid"`
}

type FindLobbyResponse struct {
	LobbyId string `json:"lobbyId,omitempty"`
}

func (h *Handler) FindLobby(c echo.Context) error {
	var request findLobbyRequest
	if err := c.Bind(&request); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error()).SetInternal(err)
	}
	if err := c.Validate(request); err != nil {
		return err
	}

	ctx := c.Request().Context()
	tx, err := h.BeginTxWithOpts(ctx, pgx.TxOptions{
		IsoLevel: pgx.Serializable,
	})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	qtx := h.DB.WithTx(tx)

	playerId, err := uuid.Parse(request.PlayerId)
	if err != nil {
		return err
	}
	lobbyId, notFound := qtx.GetFirstFreeLobby(ctx, playerId)
	if notFound == nil {
		return c.JSON(http.StatusOK, FindLobbyResponse{lobbyId.String()})
	}

	lobbyId, err = uuid.NewV7()
	if err != nil {
		return err
	}

	err = qtx.CreateLobby(ctx, sqlc.CreateLobbyParams{
		ID:           lobbyId,
		Player1ID:    playerId,
		CreatedAtUtc: time.Now().UTC(),
	})
	if err != nil {
		return err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, FindLobbyResponse{lobbyId.String()})
}
