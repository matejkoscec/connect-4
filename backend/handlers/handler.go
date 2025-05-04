package handlers

import (
	"backend/config"
	"backend/generated/sqlc"
	"context"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
)

type Handler struct {
	DB     *sqlc.Queries
	Config config.Config

	pool *pgxpool.Pool
}

func (h *Handler) BeginTx(ctx context.Context) (pgx.Tx, error) {
	return h.pool.Begin(ctx)
}

func (h *Handler) BeginTxWithOpts(ctx context.Context, options pgx.TxOptions) (pgx.Tx, error) {
	return h.pool.BeginTx(ctx, options)
}

func ConfigureRoutes(e *echo.Echo, cfg *config.Config, pool *pgxpool.Pool) {
	h := &Handler{
		DB:     sqlc.New(pool),
		Config: *cfg,
		pool:   pool,
	}

	apiV1 := e.Group("/api/v1")

	users := apiV1.Group("/users")
	users.POST("", h.CreateUser)
	users.POST("/login", h.LoginUser)

	lobbies := apiV1.Group("/lobbies")
	lobbies.POST("", h.CreateLobby)
	lobbies.GET("/find", h.FindLobby)
}
