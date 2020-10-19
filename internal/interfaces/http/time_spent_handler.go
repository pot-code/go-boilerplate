package http

import (
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/pot-code/go-boilerplate/internal/domain"
	infra "github.com/pot-code/go-boilerplate/internal/infrastructure"
	"github.com/pot-code/go-boilerplate/internal/infrastructure/auth"
)

type TimeSpentHandler struct {
	timeSpentUseCase domain.TimeSpentUseCase
	validator        infra.Validator
	jwtUtil          *auth.JWTUtil
}

func NewTimeSpentHandler(
	TimeSpentUseCase domain.TimeSpentUseCase,
	JWTUtil *auth.JWTUtil,
	Validator infra.Validator,
) *TimeSpentHandler {
	handler := &TimeSpentHandler{TimeSpentUseCase, Validator, JWTUtil}
	return handler
}

func (tsh *TimeSpentHandler) HandleGetTimeSpent(c echo.Context) (err error) {
	tsu := tsh.timeSpentUseCase
	ju := tsh.jwtUtil
	ts := c.QueryParam("ts")
	claims := ju.GetContextToken(c)
	user := new(domain.UserModel)
	user.ID = claims.UID

	// validation
	if err := tsh.validator.Empty("ts", ts); err != nil {
		return c.JSON(http.StatusBadRequest, NewRESTValidationError(http.StatusBadRequest, "Failed to validate params", err))
	}
	at, err := time.Parse(time.RFC3339, ts)
	if err != nil {
		return c.JSON(http.StatusBadRequest, NewRESTValidationError(http.StatusBadRequest, "Failed to validate params", []*infra.FieldError{{
			Domain: "ts",
			Reason: fmt.Sprintf("ts must be int RFC3339 layout, %s", err.Error()),
		}}))
	}

	timeSpent, err := tsu.GetUserTimeSpent(c.Request().Context(), user, &at)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, timeSpent)
}
