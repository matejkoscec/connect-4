package handlers

import (
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"time"
)

type loginUserRequest struct {
	Username string `json:"username,omitempty" validate:"required"`
	Password string `json:"password,omitempty" validate:"required"`
}

type UserClaims struct {
	UserID   uuid.UUID `json:"user_id"`
	Username string    `json:"username"`
	jwt.RegisteredClaims
}

func (h *Handler) LoginUser(c echo.Context) error {
	var request loginUserRequest
	if err := c.Bind(&request); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error()).SetInternal(err)
	}
	if err := c.Validate(request); err != nil {
		return err
	}

	user, notFound := h.DB.GetUserByUsername(c.Request().Context(), request.Username)
	if notFound != nil {
		return echo.NewHTTPError(http.StatusNotFound, fmt.Sprintf("User with username %s not found", request.Username))
	}

	if err := bcrypt.CompareHashAndPassword(user.Password, []byte(request.Password)); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Username or password is incorrect")
	}

	claims := UserClaims{
		user.ID,
		user.Username,
		jwt.RegisteredClaims{
			Issuer:    "connect-4",
			Subject:   user.Username,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24)),
			NotBefore: jwt.NewNumericDate(time.Now()),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(h.Config.App.Security.JwtSecret))
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to sign token").SetInternal(err)
	}

	return c.JSON(http.StatusOK, map[string]string{"token": signedToken})
}
