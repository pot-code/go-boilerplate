package middleware

import (
	"context"
	"time"

	"github.com/labstack/echo/v4"
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
			startTime := time.Now()
			err := next(c)
			endTime := time.Now()
			logger.Debug("", zap.Duration("time", endTime.Sub(startTime)), zap.Int("http.response.status_code", c.Response().Status))
			return err
		}
	}
}

// SetTraceLogger set logger binding with trace ID into context
func SetTraceLogger(base *zap.Logger, loggerKey interface{}) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			r := c.Request()
			logger := base.With(zap.String("trace.id", c.Response().Header().Get(echo.HeaderXRequestID)))
			nr := r.WithContext(context.WithValue(r.Context(), loggerKey, logger))
			c.SetRequest(nr)
			return next(c)
		}
	}
}
