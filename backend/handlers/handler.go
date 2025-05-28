package handlers

import (
	"backend/cache"
	"backend/config"
	"backend/generated/sqlc"
	"context"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"net/http"
)

type Handler struct {
	DB        *sqlc.Queries
	Config    config.Config
	Conn      *pgxpool.Pool
	GameCache *cache.Cache
	BaseCtx   context.Context
}

func ConfigureRoutes(h *Handler, e *echo.Echo) {
	e.GET(
		"/health", func(c echo.Context) error {
			return c.NoContent(http.StatusOK)
		},
	)

	apiV1 := e.Group("/api/v1")

	auth := apiV1.Group("/auth")
	auth.POST("/register", h.CreateUser)
	auth.POST("/login", h.LoginUser)

	jwtMiddleware := echojwt.WithConfig(
		echojwt.Config{
			SigningKey: []byte(h.Config.App.Security.JwtSecret),
			NewClaimsFunc: func(c echo.Context) jwt.Claims {
				return new(UserClaims)
			},
		},
	)

	games := apiV1.Group("/games")
	games.GET(
		"/play",
		h.PlayGame,
		func(next echo.HandlerFunc) echo.HandlerFunc {
			return func(c echo.Context) error {
				token := c.QueryParam("token")
				c.Request().Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
				return next(c)
			}
		},
		jwtMiddleware,
	)
}
