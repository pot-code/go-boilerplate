package middleware

import (
	"net/http"

	"github.com/labstack/echo/v4"
	infra "github.com/pot-code/go-boilerplate/internal/infrastructure"
	"go.uber.org/zap"
)

// ErrorHandlingOption options for error handling
type ErrorHandlingOption struct {
	Handler func(c echo.Context, traceId string, err error)
	Logger  *zap.Logger
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
		if option.Logger != nil {
			custom.Logger = option.Logger
		}
	}
	handler := custom.Handler
	logger := custom.Logger
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			var traceID string
			if rid := c.Request().Header.Get(echo.HeaderXRequestID); rid != "" {
				traceID = rid
			}
			defer func() {
				if any := recover(); any != nil {
					err := any.(error)
					if logger != nil {
						logger.Error(err.Error(),
							zap.String("url.path", c.Request().RequestURI),
							zap.String("client.address", c.Request().RemoteAddr),
							zap.String("http.request.method", c.Request().Method),
							zap.Int64("http.request.body.bytes", c.Request().ContentLength),
							zap.Strings("route.params.name", c.ParamNames()),
							zap.Strings("route.params.value", c.ParamValues()),
							zap.String("trace.id", traceID),
						)
					}
					handler(c, "", err)
				}
			}()
			if err := next(c); err != nil {
				c.JSON(http.StatusInternalServerError,
					infra.NewRESTStandardError(http.StatusInternalServerError, err.Error()).SetTraceID(traceID),
				)
			}
			return nil
		}
	}
}
