package middleware

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/pot-code/go-boilerplate/internal/infrastructure/logging"
	"go.uber.org/zap"
)

type LoggingConfig struct {
	// Skipper defines a function to skip middleware.
	Skipper middleware.Skipper
}

// Logging create a logging middleware with zap logger
func Logging(base *zap.Logger, options ...*LoggingConfig) echo.MiddlewareFunc {
	cfg := &LoggingConfig{
		Skipper: middleware.DefaultSkipper,
	}
	if len(options) > 0 {
		option := options[0]
		if option.Skipper != nil {
			cfg.Skipper = option.Skipper
		}
	}
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if cfg.Skipper(c) {
				return next(c)
			}
			err := next(c)
			rid := c.Response().Header().Get(echo.HeaderXRequestID)
			logger := base.With(
				zap.String("trace.id", rid),
				zap.String("url.path", c.Request().RequestURI),
				zap.String("client.address", c.Request().RemoteAddr),
				zap.String("http.request.method", c.Request().Method),
				zap.Int64("http.request.body.byte", c.Request().ContentLength),
			)
			if len(c.ParamNames()) > 0 {
				logger = logger.With(
					zap.Strings("route.params.name", c.ParamNames()),
					zap.Strings("route.params.value", c.ParamValues()),
				)
			}
			code := c.Response().Status
			logger.Debug(http.StatusText(code), zap.Int("http.response.status_code", code))
			return err
		}
	}
}

// SetTraceLogger set logger binding with trace ID into context
func SetTraceLogger(base *zap.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			r := c.Request()
			logger := base.With(zap.String("trace.id", c.Response().Header().Get(echo.HeaderXRequestID)))
			nr := r.WithContext(logging.SetLoggerInContext(r.Context(), logger))
			c.SetRequest(nr)
			return next(c)
		}
	}
}
