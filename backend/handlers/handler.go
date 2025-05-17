package handlers

import (
	"backend/config"
	"backend/generated/sqlc"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo-jwt/v4"
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

	auth := apiV1.Group("/auth")
	auth.POST("/register", h.CreateUser)
	auth.POST("/login", h.LoginUser)

	lobbies := apiV1.Group("/lobbies")
	lobbies.Use(echojwt.WithConfig(echojwt.Config{
		SigningKey: []byte(cfg.App.Security.JwtSecret),
	}))
	lobbies.POST("", h.CreateLobby)
	lobbies.GET("/find", h.FindLobby)
}
