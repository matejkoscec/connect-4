package handlers

import (
	"backend/generated/sqlc"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"time"
)

type createUserRequest struct {
	Username string `json:"username,omitempty" validate:"required"`
	Email    string `json:"email,omitempty" validate:"required,email"`
	Password string `json:"password,omitempty" validate:"required"`
}

func (h *Handler) CreateUser(c echo.Context) error {
	var request createUserRequest
	if err := c.Bind(&request); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error()).SetInternal(err)
	}
	if err := c.Validate(request); err != nil {
		return err
	}

	_, notFound := h.DB.GetUserByUsernameOrEmail(c.Request().Context(), sqlc.GetUserByUsernameOrEmailParams{
		Username: request.Username,
		Email:    request.Email,
	})
	if notFound == nil {
		return echo.NewHTTPError(http.StatusConflict, "User with that username or email already exists")
	}

	password, err := bcrypt.GenerateFromPassword([]byte(request.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	id, err := uuid.NewV7()
	if err != nil {
		return err
	}

	err = h.DB.CreateUser(c.Request().Context(), sqlc.CreateUserParams{
		ID:           id,
		Username:     request.Username,
		Email:        request.Email,
		Password:     password,
		CreatedAtUtc: time.Now().UTC(),
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error()).SetInternal(err)
	}

	c.Logger().Infof("Added new user: %v", request.Username)

	return c.String(http.StatusCreated, id.String())
}
