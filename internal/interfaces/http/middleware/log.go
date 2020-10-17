package middleware

import (
	"context"
	"time"

	"github.com/labstack/echo/v4"
	infra "github.com/pot-code/go-boilerplate/internal/infrastructure"
	"go.uber.org/zap"
)

// Logging create a logging middleware with zap logger
func Logging(base *zap.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			logger := base.With(
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
			r := c.Request()
			logger = logger.With(zap.String("trace.id", c.Response().Header().Get(echo.HeaderXRequestID)))
			nr := r.WithContext(context.WithValue(r.Context(), infra.ContextLoggerKey, logger))
			c.SetRequest(nr)

			startTime := time.Now()
			next(c)
			endTime := time.Now()
			logger.Info("", zap.Duration("time", endTime.Sub(startTime)), zap.Int("http.response.status_code", c.Response().Status))
			return nil
		}
	}
}
