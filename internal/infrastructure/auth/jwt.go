package auth

import (
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo/v4"
	"github.com/pot-code/go-boilerplate/internal/domain"
)

// AppTokenClaims .
type AppTokenClaims struct {
	UID   string `json:"uid"`
	Email string `json:"email"`
	Name  string `json:"name"`

	jwt.StandardClaims
}

// TimeRemaining remaining time before the token get expired
func (tk *AppTokenClaims) TimeRemaining() time.Duration {
	exp := time.Unix(tk.ExpiresAt, 0)
	now := time.Now()

	if exp.Before(now) {
		return 0
	}
	return exp.Sub(now)
}

// JWTUtil .
type JWTUtil struct {
	secret    []byte
	tokenName string
	timeout   time.Duration
	method    jwt.SigningMethod
}

// NewJWTUtil create a JWTUtil instance
func NewJWTUtil(method, secret, tokenName string, timeout time.Duration) *JWTUtil {
	var signMethod jwt.SigningMethod
	switch method {
	case "HS256":
		signMethod = jwt.SigningMethodHS256
	case "HS512":
		signMethod = jwt.SigningMethodHS512
	case "ES256":
		signMethod = jwt.SigningMethodES256
	default:
		signMethod = jwt.SigningMethodHS256
	}
	bsecret := []byte(secret)
	return &JWTUtil{
		method:    signMethod,
		secret:    bsecret,
		tokenName: tokenName,
		timeout:   timeout,
	}
}

// Sign sign token
func (ju *JWTUtil) Sign(claims *AppTokenClaims) (string, error) {
	token := jwt.NewWithClaims(ju.method, claims)
	return token.SignedString(ju.secret)
}

// Validate validate token string with secret and return AppTokenClaims
func (ju *JWTUtil) Validate(tokenStr string) (*AppTokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &AppTokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		return ju.secret, nil
	})
	if err != nil {
		return nil, err
	}
	return token.Claims.(*AppTokenClaims), nil
}

// GenerateTokenStr generate user token from user model
func (ju *JWTUtil) GenerateTokenStr(user *domain.UserModel) (string, error) {
	expires := time.Now().Add(ju.timeout).Unix()
	return ju.Sign(&AppTokenClaims{
		UID:   user.ID,
		Email: user.Email,
		Name:  user.Username,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expires,
		},
	})
}

// RefreshToken set token expiration to now
func (ju *JWTUtil) RefreshToken(claims *AppTokenClaims) *AppTokenClaims {
	expires := time.Now().Add(ju.timeout).Unix()
	claims.ExpiresAt = expires
	return claims
}

// SetClientToken set token in client cookie
func (ju *JWTUtil) SetClientToken(c echo.Context, tokenStr string) {
	c.SetCookie(&http.Cookie{
		Name:     ju.tokenName,
		Value:    tokenStr,
		HttpOnly: true,
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().Add(ju.timeout),
	})
}

// ClearClientToken clear client cookie
func (ju *JWTUtil) ClearClientToken(c echo.Context) {
	c.SetCookie(&http.Cookie{
		Name:     ju.tokenName,
		Value:    "",
		HttpOnly: true,
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now(),
	})
}

// SetContextToken set token in App context
func (ju *JWTUtil) SetContextToken(c echo.Context, token *AppTokenClaims) {
	c.Set(ju.tokenName, token)
}

// GetContextToken get token from App context
func (ju *JWTUtil) GetContextToken(c echo.Context) *AppTokenClaims {
	v, ok := c.Get(ju.tokenName).(*AppTokenClaims)
	if ok {
		return v
	}
	return nil
}

// ExtractToken get token string from request
func (ju *JWTUtil) ExtractToken(c echo.Context) (string, error) {
	token, err := c.Cookie(ju.tokenName)
	if err != nil {
		return "", err
	}
	return token.Value, nil
}
