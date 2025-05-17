package handlers

import (
	"backend/config"
	"backend/generated/sqlc"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"net/http"
)

type Handler struct {
	DB     *sqlc.Queries
	Config config.Config
	Conn   *pgxpool.Pool
}

func ConfigureRoutes(e *echo.Echo, cfg *config.Config, pool *pgxpool.Pool) {
	h := &Handler{
		DB:     sqlc.New(pool),
		Config: *cfg,
		Conn:   pool,
	}

	e.GET("/health", func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})

	apiV1 := e.Group("/api/v1")

	users := apiV1.Group("/users")
	users.POST("", h.CreateUser)
	users.POST("/login", h.LoginUser)

	lobbies := apiV1.Group("/lobbies")
	lobbies.POST("", h.CreateLobby)
	lobbies.GET("/find", h.FindLobby)
}
