package handlers

import (
	"backend/generated/sqlc"
	"github.com/labstack/echo/v4"
)

type Handler struct {
	DB *sqlc.Queries
}

func Configure(e *echo.Echo, h *Handler) {
	apiV1 := e.Group("/api/v1")

	users := apiV1.Group("/users")
	users.POST("", h.CreateUser)
}
