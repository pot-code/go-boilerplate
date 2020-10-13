package http

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/pot-code/go-boilerplate/internal/domain"
	infra "github.com/pot-code/go-boilerplate/internal/infrastructure"
	"github.com/pot-code/go-boilerplate/internal/infrastructure/driver"
)

// UserHandler user related operations
type UserHandler struct {
	JWTUtil     *infra.JWTUtil
	TokenStore  driver.KeyValueDB
	UserUseCase domain.UserUseCase
	Validator   infra.Validator
}

// NewUserHandler create an user controller instance
func NewUserHandler(
	UserUseCase domain.UserUseCase,
	JWTUtil *infra.JWTUtil,
	KVStore driver.KeyValueDB,
	Validator infra.Validator,
) *UserHandler {
	handler := &UserHandler{
		JWTUtil:     JWTUtil,
		UserUseCase: UserUseCase,
		Validator:   Validator,
		TokenStore:  KVStore,
	}
	return handler
}

// HandleSignIn ...
func (uh *UserHandler) HandleSignIn(c echo.Context) (err error) {
	UserUseCase := uh.UserUseCase
	ju := uh.JWTUtil

	// parse body
	post := new(domain.UserModel)
	if err = c.Bind(&post); err != nil {
		internal := err.(*echo.HTTPError).Internal
		return c.JSON(http.StatusUnprocessableEntity,
			infra.NewRESTStandardError(http.StatusUnprocessableEntity, "Failed to bind user entity").SetDetail(internal.Error()))
	}

	user, err := UserUseCase.SignIn(c.Request().Context(), post)
	if err != nil {
		if errors.Is(err, domain.ErrNoSuchUser) {
			return c.JSON(http.StatusUnauthorized, infra.NewRESTStandardError(http.StatusUnauthorized, err.Error()))
		}
		return
	}

	// issue JWT
	tokenStr, err := ju.GenerateTokenStr(user)
	if err != nil {
		return err
	}
	ju.SetClientToken(c, tokenStr)
	return nil
}

// HandleSignUp ...
func (uh *UserHandler) HandleSignUp(c echo.Context) (err error) {
	UserUseCase := uh.UserUseCase
	post := new(domain.UserModel)

	if err = c.Bind(&post); err != nil {
		internal := err.(*echo.HTTPError).Internal
		return c.JSON(http.StatusUnprocessableEntity,
			infra.NewRESTStandardError(http.StatusUnprocessableEntity, "Failed to bind user entity").SetDetail(internal.Error()))
	}

	// validation
	if err := uh.Validator.Struct(post); err != nil {
		return c.JSON(http.StatusBadRequest,
			infra.NewRESTValidationError(http.StatusBadRequest, "Failed to validate fields", err))
	}

	// register
	_, err = UserUseCase.SignUp(c.Request().Context(), post)
	if err != nil {
		if errors.Is(err, domain.ErrDuplicatedUser) {
			return c.JSON(http.StatusConflict, infra.NewRESTStandardError(http.StatusConflict, err.Error()))
		}
		return err
	}
	return
}

// HandleSignOut ...
func (uh *UserHandler) HandleSignOut(c echo.Context) (err error) {
	ju := uh.JWTUtil
	kv := uh.TokenStore

	if tokenStr, err := ju.ExtractToken(c); err == nil {
		if token, err := ju.Validate(tokenStr); err == nil {
			ju.ClearClientToken(c)
			return kv.SetEX(tokenStr, "", token.TimeRemaining())
		}
		return c.NoContent(http.StatusUnauthorized)
	}
	return nil
}

// HandleUserExists ...
func (uh *UserHandler) HandleUserExists(c echo.Context) (err error) {
	UserUseCase := uh.UserUseCase
	post := new(domain.UserModel)
	post.Username = c.QueryParam("username")
	post.Email = c.QueryParam("email")

	if err := uh.Validator.AllEmpty([]string{"username", "email"}, post.Username, post.Email); err != nil {
		return c.JSON(http.StatusBadRequest, infra.NewRESTValidationError(http.StatusBadRequest, "Failed to validate params", err))
	}

	existing, err := UserUseCase.Exists(c.Request().Context(), post)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, existing)
}
