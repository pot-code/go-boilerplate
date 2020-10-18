package middleware

import (
	"net/http"

	"github.com/labstack/echo/v4"
	infra "github.com/pot-code/go-boilerplate/internal/infrastructure"
	"go.uber.org/zap"
)

// ErrorHandlingOption options for error handling
type ErrorHandlingOption struct {
	Handler   func(c echo.Context, traceId string, err error)
	LoggerKey interface{}
}

// ErrorHandling handle panic returned from controller
// **DO NOT return error anymore**
func ErrorHandling(options ...*ErrorHandlingOption) echo.MiddlewareFunc {
	custom := &ErrorHandlingOption{
		Handler: func(c echo.Context, traceID string, err error) {
			c.JSON(http.StatusInternalServerError,
				infra.NewRESTStandardError(http.StatusInternalServerError, err.Error()).SetTraceID(traceID),
			)
		},
	}
	if len(options) > 0 {
		option := options[0]
		if option.Handler != nil {
			custom.Handler = option.Handler
		}
		if option.LoggerKey != nil {
			custom.LoggerKey = option.LoggerKey
		}
	}
	handler := custom.Handler

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			var (
				logger  *zap.Logger
				traceID string
			)
			if custom.LoggerKey != nil {
				logger = c.Request().Context().Value(custom.LoggerKey).(*zap.Logger)
			}
			if rid := c.Response().Header().Get(echo.HeaderXRequestID); rid != "" {
				traceID = rid
			}
			defer func() {
				if any := recover(); any != nil {
					err := any.(error)
					if logger != nil {
						logger.Error(err.Error())
					}
					handler(c, "", err)
				}
			}()
			if err := next(c); err != nil {
				if v, ok := err.(*echo.HTTPError); ok {
					c.String(v.Code, v.Error())
				} else {
					c.JSON(http.StatusInternalServerError,
						infra.NewRESTStandardError(http.StatusInternalServerError, err.Error()).SetTraceID(traceID),
					)
				}
			}
			return nil
		}
	}
}
