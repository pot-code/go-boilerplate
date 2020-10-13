package middleware

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// NoRouteMatched no matched route handler
func NoRouteMatched() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			err := next(c)
			if v, ok := err.(*echo.HTTPError); ok && v.Code == http.StatusNotFound {
				return c.NoContent(v.Code)
			}
			return err
		}
	}
}
