package middleware

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// CORSMiddleware handle preflight request
func CORSMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		c.Response().Header().Add("Access-Control-Allow-Origin", "http://127.0.0.1:8080")
		c.Response().Header().Add("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept, X-CSRF-Token, Authorization")
		c.Response().Header().Add("Access-Control-Allow-Credentials", "true")
		if c.Request().Method == http.MethodOptions {
			return nil
		}
		return next(c)
	}
}
