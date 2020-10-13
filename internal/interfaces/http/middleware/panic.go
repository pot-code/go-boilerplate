package middleware

import (
	"net/http"

	"github.com/labstack/echo/v4"
	infra "github.com/pot-code/go-boilerplate/internal/infrastructure"
	"go.uber.org/zap"
)

// PanicHandlingOption options for error handling
type PanicHandlingOption struct {
	Handler func(c echo.Context, err error)
	Logger  *zap.Logger
}

// PanicHandling handle panic returned from controller
func PanicHandling(options ...*PanicHandlingOption) echo.MiddlewareFunc {
	custom := &PanicHandlingOption{
		Handler: func(c echo.Context, err error) {
			c.JSON(http.StatusInternalServerError,
				infra.RESTStandardError{
					Code:  http.StatusInternalServerError,
					Title: err.Error(),
				},
			)
		},
	}
	if len(options) > 0 {
		option := options[0]
		handler := option.Handler
		if handler != nil {
			custom.Handler = handler
		}
		if option.Logger != nil {
			custom.Logger = option.Logger
		}
	}
	handler := custom.Handler
	logger := custom.Logger
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			defer func() {
				if any := recover(); any != nil {
					err := any.(error)
					if logger != nil {
						logger.Error(err.Error(),
							zap.String("url.path", c.Request().RequestURI),
							zap.String("http.request.method", c.Request().Method),
							zap.String("http.request.body.content", c.Request().Header.Get(echo.HeaderContentType)),
							zap.Int64("http.request.body.bytes", c.Request().ContentLength),
							zap.Strings("route.params.name", c.ParamNames()),
							zap.Strings("route.params.value", c.ParamValues()),
							zap.Int("http.response.status_code", http.StatusInternalServerError),
						)
					}
					handler(c, err)
				}
			}()
			return next(c)
		}
	}
}
