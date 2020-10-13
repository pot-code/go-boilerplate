package http

import (
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/pot-code/go-boilerplate/internal/domain"
	infra "github.com/pot-code/go-boilerplate/internal/infrastructure"
)

type TimeSpentHandler struct {
	TimeSpentUseCase domain.TimeSpentUseCase
	Validator        infra.Validator
	JWTUtil          *infra.JWTUtil
}

func NewTimeSpentHandler(
	TimeSpentUseCase domain.TimeSpentUseCase,
	JWTUtil *infra.JWTUtil,
	Validator infra.Validator,
) *TimeSpentHandler {
	handler := &TimeSpentHandler{
		TimeSpentUseCase: TimeSpentUseCase,
		Validator:        Validator,
		JWTUtil:          JWTUtil,
	}
	return handler
}

func (tsh *TimeSpentHandler) HandleGetTimeSpent(c echo.Context) (err error) {
	tsu := tsh.TimeSpentUseCase
	ju := tsh.JWTUtil
	ts := c.QueryParam("ts")
	claims := ju.GetContextToken(c)
	user := new(domain.UserModel)
	user.ID = claims.UID

	// validation
	if err := tsh.Validator.Empty("user ID", user.ID); err != nil {
		return c.JSON(http.StatusBadRequest, infra.NewRESTValidationError(http.StatusBadRequest, "Failed to validate params", err))
	}
	if err := tsh.Validator.Empty("ts", ts); err != nil {
		return c.JSON(http.StatusBadRequest, infra.NewRESTValidationError(http.StatusBadRequest, "Failed to validate params", err))
	}
	at, err := time.Parse(time.RFC3339, ts)
	if err != nil {
		return c.JSON(http.StatusBadRequest, infra.NewRESTValidationError(http.StatusBadRequest, "Failed to validate params", []*infra.FieldError{{
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
