package middleware

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// ErrorHandlingOption options for error handling
type ErrorHandlingOption struct {
	Handler func(c echo.Context, err error)
}

// ErrorHandling handle panic returned from controller
// **DO NOT return error anymore**
func ErrorHandling(options ...*ErrorHandlingOption) echo.MiddlewareFunc {
	custom := &ErrorHandlingOption{
		Handler: func(c echo.Context, err error) {
			c.String(http.StatusInternalServerError, err.Error())
		},
	}
	if len(options) > 0 {
		option := options[0]
		if option.Handler != nil {
			custom.Handler = option.Handler
		}
	}
	handler := custom.Handler
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			defer func() {
				if any := recover(); any != nil {
					err := any.(error)
					handler(c, err)
				}
			}()
			if err := next(c); err != nil {
				if v, ok := err.(*echo.HTTPError); ok {
					c.String(v.Code, v.Error())
				} else {
					handler(c, err)
				}
			}
			return nil
		}
	}
}
