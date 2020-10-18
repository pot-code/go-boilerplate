package http

import (
	"database/sql"
	"errors"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/pot-code/go-boilerplate/internal/domain"
	infra "github.com/pot-code/go-boilerplate/internal/infrastructure"
	"github.com/pot-code/go-boilerplate/internal/infrastructure/auth"
	"github.com/pot-code/go-boilerplate/internal/infrastructure/driver"
	"golang.org/x/crypto/bcrypt"
)

var (
	// ErrNoSuchUser failed to validate the credential
	ErrNoSuchUser = errors.New("No such user or password is incorrect")
	// ErrUserTooManyRetry excess maximum retry count
	ErrUserTooManyRetry = errors.New("Excess maximum retry count")
)

// UserHandler user related operations
type UserHandler struct {
	JWTUtil        *auth.JWTUtil
	UserRepository *auth.UserRepository
	KVStore        driver.KeyValueDB
	UserUseCase    domain.UserUseCase
	Validator      infra.Validator
	MaximumRetry   int
	RetryTimeout   time.Duration
}

type UserLoginModel struct {
	Username string `json:"username" validate:"required,min=6,max=64"`
	Password string `json:"password" validate:"required,min=6"`
}

func (ulm *UserLoginModel) ToDomain() *domain.UserModel {
	return &domain.UserModel{
		Username: ulm.Username,
		Password: ulm.Password,
	}
}

type UserCheckModel struct {
	Username string `json:"username" validate:"omitempty,min=6,max=64"`
	Email    string `json:"email" validate:"omitempty,email"`
}

func (ucm *UserCheckModel) ToDomain() *domain.UserModel {
	return &domain.UserModel{
		Username: ucm.Username,
		Email:    ucm.Email,
	}
}

// NewUserHandler create an user controller instance
func NewUserHandler(
	JWTUtil *auth.JWTUtil,
	UserRepository *auth.UserRepository,
	KVStore driver.KeyValueDB,
	UserUseCase domain.UserUseCase,
	MaximumRetry int,
	RetryTimeout time.Duration,
	Validator infra.Validator,
) *UserHandler {
	handler := &UserHandler{JWTUtil, UserRepository, KVStore, UserUseCase, Validator, MaximumRetry, RetryTimeout}
	return handler
}

// HandleSignIn ...
func (uh *UserHandler) HandleSignIn(c echo.Context) (err error) {
	ju := uh.JWTUtil
	repo := uh.UserRepository
	conn := repo.Conn
	ctx := c.Request().Context()

	// parse body
	post := new(UserLoginModel)
	if err = c.Bind(&post); err != nil {
		// internal := err.(*echo.HTTPError).Internal
		return c.JSON(http.StatusUnprocessableEntity,
			infra.NewRESTStandardError(http.StatusUnprocessableEntity, "Failed to bind user entity"))
	}
	if err := uh.Validator.Struct(post); err != nil {
		return c.JSON(http.StatusBadRequest,
			infra.NewRESTValidationError(http.StatusBadRequest, "Failed to validate credentials", err))
	}

	tx, err := conn.BeginTx(ctx, &driver.TxOptions{
		Isolation: sql.LevelRepeatableRead,
	})
	if err != nil {
		return c.JSON(http.StatusInternalServerError,
			infra.NewRESTStandardError(http.StatusInternalServerError, err.Error()))
	}
	defer tx.Commit(ctx)

	// find user
	user, err := repo.FindByCredential(ctx, post.ToDomain())
	if err != nil {
		return c.JSON(http.StatusInternalServerError,
			infra.NewRESTStandardError(http.StatusInternalServerError, err.Error()))
	}
	if user == nil {
		return c.JSON(http.StatusUnauthorized, infra.NewRESTStandardError(http.StatusUnauthorized, ErrNoSuchUser.Error()))
	}
	now := time.Now().Unix() // seconds
	if user.LoginRetry >= uh.MaximumRetry && now-user.LastLogin < int64(uh.RetryTimeout.Seconds()) {
		return c.JSON(http.StatusForbidden, infra.NewRESTStandardError(http.StatusForbidden, ErrUserTooManyRetry.Error()))
	}

	// check credentials
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(post.Password)); err != nil {
		if err == bcrypt.ErrMismatchedHashAndPassword {
			if user.LoginRetry == uh.MaximumRetry {
				user.LoginRetry = 1
			} else {
				user.LoginRetry++
			}
			user.LastLogin = now
			repo.UpdateUser(ctx, user)
			return c.JSON(http.StatusUnauthorized, infra.NewRESTStandardError(http.StatusUnauthorized, ErrNoSuchUser.Error()))
		}
		return c.JSON(http.StatusInternalServerError,
			infra.NewRESTStandardError(http.StatusInternalServerError, "Failed to process user credential"))
	}

	// reset retry number
	user.LoginRetry = 0
	user.LastLogin = now
	repo.UpdateUser(ctx, user)

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
	ctx := c.Request().Context()

	if err = c.Bind(&post); err != nil {
		// internal := err.(*echo.HTTPError).Internal
		return c.JSON(http.StatusUnprocessableEntity,
			infra.NewRESTStandardError(http.StatusUnprocessableEntity, "Failed to bind user entity"))
	}

	// validation
	if err := uh.Validator.Struct(post); err != nil {
		return c.JSON(http.StatusBadRequest,
			infra.NewRESTValidationError(http.StatusBadRequest, "Failed to validate fields", err))
	}

	// hash password
	if password, err := bcrypt.GenerateFromPassword([]byte(post.Password), bcrypt.MinCost); err == nil {
		post.Password = string(password)
	} else {
		return c.JSON(http.StatusInternalServerError,
			infra.NewRESTStandardError(http.StatusInternalServerError, "Failed to process user credential"))
	}

	// register
	_, err = UserUseCase.SignUp(ctx, post)
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
	kv := uh.KVStore

	if tokenStr, err := ju.ExtractToken(c); err == nil {
		if token, err := ju.Validate(tokenStr); err == nil {
			ju.ClearClientToken(c)
			return kv.SetEX(tokenStr, "", token.TimeRemaining())
		}
		return c.NoContent(http.StatusForbidden)
	}
	return nil
}

// HandleUserExists ...
func (uh *UserHandler) HandleUserExists(c echo.Context) (err error) {
	UserUseCase := uh.UserUseCase
	ctx := c.Request().Context()
	post := new(UserCheckModel)
	post.Username = c.QueryParam("username")
	post.Email = c.QueryParam("email")

	if err := uh.Validator.AllEmpty([]string{"username", "email"}, post.Username, post.Email); err != nil {
		return c.JSON(http.StatusBadRequest, infra.NewRESTValidationError(http.StatusBadRequest, "Failed to validate params", []*infra.FieldError{err}))
	}
	if err := uh.Validator.Struct(post); err != nil {
		return c.JSON(http.StatusBadRequest,
			infra.NewRESTValidationError(http.StatusBadRequest, "Failed to validate fields", err))
	}

	existing, err := UserUseCase.Exists(ctx, post.ToDomain())
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, existing)
}
