package middleware

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/pot-code/go-boilerplate/internal/infrastructure/auth"
)

// ValidateTokenOption ...
type ValidateTokenOption struct {
	InBlackList func(token string) (bool, error)
}

// RefreshTokenOption ...
type RefreshTokenOption struct {
	Threshold time.Duration
}

// VerifyToken validate JWT
func VerifyToken(ju *auth.JWTUtil, options ...*ValidateTokenOption) echo.MiddlewareFunc {
	inBlacklist := func(string) (bool, error) { return true, nil }
	if len(options) > 0 {
		option := options[0]
		inBlacklist = option.InBlackList
	}
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			tokenStr, err := ju.ExtractToken(c)
			if err != nil {
				return c.NoContent(http.StatusUnauthorized)
			}

			if ok, err := inBlacklist(tokenStr); err != nil {
				return err
			} else if ok {
				return c.NoContent(http.StatusUnauthorized)
			}

			token, err := ju.Validate(tokenStr)
			if err == nil {
				ju.SetContextToken(c, token)
				return next(c)
			}
			return c.NoContent(http.StatusUnauthorized)
		}
	}
}

// RefreshToken refresh jwt if necessary, must be chained after ValidateMiddleware
func RefreshToken(ju *auth.JWTUtil, options ...*RefreshTokenOption) echo.MiddlewareFunc {
	threshold := 5 * time.Minute
	if len(options) > 0 {
		if option := options[0]; option.Threshold > 0 {
			threshold = option.Threshold
		}
	}
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			claims := ju.GetContextToken(c)
			if claims == nil {
				return next(c)
			}
			exp := claims.ExpiresAt
			if time.Unix(exp, 0).Sub(time.Now()) < threshold {
				ju.RefreshToken(claims)
				if tokenStr, err := ju.Sign(claims); err == nil {
					ju.SetClientToken(c, tokenStr)
				} else {
					return err
				}
			}
			return next(c)
		}
	}
}
