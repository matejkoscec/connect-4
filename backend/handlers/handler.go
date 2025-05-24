package handlers

import (
	"backend/config"
	"backend/game"
	"backend/generated/sqlc"
	"context"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"net/http"
)

type Handler struct {
	DB        *sqlc.Queries
	Config    config.Config
	Conn      *pgxpool.Pool
	GameCache *game.Cache
	BaseCtx   context.Context
}

func ConfigureRoutes(h *Handler, e *echo.Echo) {
	e.GET("/health", func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})

	apiV1 := e.Group("/api/v1")

	auth := apiV1.Group("/auth")
	auth.POST("/register", h.CreateUser)
	auth.POST("/login", h.LoginUser)

	jwtMiddleware := echojwt.WithConfig(echojwt.Config{
		SigningKey: []byte(h.Config.App.Security.JwtSecret),
		NewClaimsFunc: func(c echo.Context) jwt.Claims {
			return new(UserClaims)
		},
	})

	games := apiV1.Group("/games", jwtMiddleware)
	games.GET("/play/ws", h.PlayGame)
}
