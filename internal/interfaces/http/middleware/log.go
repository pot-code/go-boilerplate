package middleware

import (
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
				zap.String("url.path", c.Request().RequestURI),
				zap.String("client.address", c.Request().RemoteAddr),
				zap.String("http.request.method", c.Request().Method),
				zap.Int64("http.request.body.bytes", c.Request().ContentLength),
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
			if err == nil {
				logger.Info("Ok", zap.Duration("time", endTime.Sub(startTime)), zap.Int("http.response.status_code", c.Response().Status))
			} else {
				logger.Error(err.Error(), zap.Duration("time", endTime.Sub(startTime)))
			}
			return err
		}
	}
}
