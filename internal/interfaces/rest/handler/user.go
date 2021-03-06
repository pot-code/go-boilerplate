package handler

import (
	"database/sql"
	"errors"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/pot-code/go-boilerplate/internal/infrastructure/auth"
	"github.com/pot-code/go-boilerplate/internal/infrastructure/driver"
	"github.com/pot-code/go-boilerplate/internal/infrastructure/validate"
	"github.com/pot-code/go-boilerplate/internal/user"
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
	jwtUtil        *auth.JWTUtil
	conn           driver.ITransactionalDB
	userRepository user.UserRepository
	kvStore        driver.KeyValueDB
	userUseCase    user.UserUseCase
	validator      validate.Validator
	maximumRetry   int
	retryTimeout   time.Duration
}

type UserLoginModel struct {
	Username string `json:"username" validate:"required,min=6,max=64"`
	Password string `json:"password" validate:"required,min=6"`
}

func (ulm *UserLoginModel) ToDomain() *user.UserModel {
	return &user.UserModel{
		Username: ulm.Username,
		Password: ulm.Password,
	}
}

type UserCheckModel struct {
	Username string `json:"username" validate:"omitempty,min=6,max=64"`
	Email    string `json:"email" validate:"omitempty,email"`
}

func (ucm *UserCheckModel) ToDomain() *user.UserModel {
	return &user.UserModel{
		Username: ucm.Username,
		Email:    ucm.Email,
	}
}

type UserSignUpModel struct {
	Username string `json:"username" validate:"required,min=6,max=31"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}

func (usm *UserSignUpModel) ToDomain() *user.UserModel {
	return &user.UserModel{
		Username: usm.Username,
		Email:    usm.Email,
		Password: usm.Password,
	}
}

// NewUserHandler create an user controller instance
func NewUserHandler(
	JWTUtil *auth.JWTUtil,
	conn driver.ITransactionalDB,
	UserRepository user.UserRepository,
	KVStore driver.KeyValueDB,
	UserUseCase user.UserUseCase,
	MaximumRetry int,
	RetryTimeout time.Duration,
	Validator validate.Validator,
) *UserHandler {
	handler := &UserHandler{JWTUtil, conn, UserRepository, KVStore, UserUseCase, Validator, MaximumRetry, RetryTimeout}
	return handler
}

// HandleSignIn ...
func (uh *UserHandler) HandleSignIn(c echo.Context) (err error) {
	ju := uh.jwtUtil
	repo := uh.userRepository
	conn := uh.conn
	ctx := c.Request().Context()

	// parse body
	post := new(UserLoginModel)
	if err = c.Bind(&post); err != nil {
		// internal := err.(*echo.HTTPError).Internal
		return c.JSON(http.StatusUnprocessableEntity,
			NewRESTStandardError(http.StatusUnprocessableEntity, "Failed to bind user entity"))
	}
	if err := uh.validator.Struct(post); err != nil {
		return c.JSON(http.StatusBadRequest,
			NewRESTValidationError(http.StatusBadRequest, "Failed to validate credentials", err))
	}

	tx, err := conn.BeginTx(ctx, &driver.TxOptions{
		Isolation: sql.LevelRepeatableRead,
	})
	if err != nil {
		return c.JSON(http.StatusInternalServerError,
			NewRESTStandardError(http.StatusInternalServerError, err.Error()))
	}
	defer tx.Commit(ctx)

	// find user
	entity, err := repo.FindByCredential(ctx, post.ToDomain())
	if err != nil {
		return c.JSON(http.StatusInternalServerError,
			NewRESTStandardError(http.StatusInternalServerError, err.Error()))
	}
	if entity == nil {
		return c.JSON(http.StatusUnauthorized, NewRESTStandardError(http.StatusUnauthorized, ErrNoSuchUser.Error()))
	}
	now := time.Now().Unix() // seconds
	if entity.LoginRetry >= uh.maximumRetry && now-entity.LastLogin < int64(uh.retryTimeout.Seconds()) {
		return c.JSON(http.StatusForbidden, NewRESTStandardError(http.StatusForbidden, ErrUserTooManyRetry.Error()))
	}

	// check credentials
	if err := bcrypt.CompareHashAndPassword([]byte(entity.Password), []byte(post.Password)); err != nil {
		if err == bcrypt.ErrMismatchedHashAndPassword {
			if entity.LoginRetry == uh.maximumRetry {
				entity.LoginRetry = 1
			} else {
				entity.LoginRetry++
			}
			entity.LastLogin = now
			repo.UpdateLogin(ctx, entity)
			return c.JSON(http.StatusUnauthorized, NewRESTStandardError(http.StatusUnauthorized, ErrNoSuchUser.Error()))
		}
		return c.JSON(http.StatusInternalServerError,
			NewRESTStandardError(http.StatusInternalServerError, "Failed to process user credential"))
	}

	// reset retry number
	entity.LoginRetry = 0
	entity.LastLogin = now
	repo.UpdateLogin(ctx, entity)

	// issue JWT
	tokenStr, err := ju.GenerateTokenStr(entity.ID, entity.Email, entity.Username)
	if err != nil {
		return err
	}
	ju.SetClientToken(c, tokenStr)
	return nil
}

// HandleSignUp ...
func (uh *UserHandler) HandleSignUp(c echo.Context) (err error) {
	UserUseCase := uh.userUseCase
	post := new(UserSignUpModel)
	ctx := c.Request().Context()

	if err = c.Bind(&post); err != nil {
		// internal := err.(*echo.HTTPError).Internal
		return c.JSON(http.StatusUnprocessableEntity,
			NewRESTStandardError(http.StatusUnprocessableEntity, "Failed to bind user entity"))
	}

	// validation
	if err := uh.validator.Struct(post); err != nil {
		return c.JSON(http.StatusBadRequest,
			NewRESTValidationError(http.StatusBadRequest, "Failed to validate fields", err))
	}

	// hash password
	if password, err := bcrypt.GenerateFromPassword([]byte(post.Password), bcrypt.MinCost); err == nil {
		post.Password = string(password)
	} else {
		return c.JSON(http.StatusInternalServerError,
			NewRESTStandardError(http.StatusInternalServerError, "Failed to process user credential"))
	}

	// register
	entity := post.ToDomain()
	entity.LastLogin = time.Now().Unix()
	_, err = UserUseCase.SignUp(ctx, entity)
	if err != nil {
		if errors.Is(err, user.ErrDuplicatedUser) {
			return c.JSON(http.StatusConflict, NewRESTStandardError(http.StatusConflict, err.Error()))
		}
		return err
	}
	return
}

// HandleSignOut ...
func (uh *UserHandler) HandleSignOut(c echo.Context) (err error) {
	ju := uh.jwtUtil
	kv := uh.kvStore

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
	UserUseCase := uh.userUseCase
	ctx := c.Request().Context()
	post := new(UserCheckModel)
	post.Username = c.QueryParam("username")
	post.Email = c.QueryParam("email")

	if err := uh.validator.AllEmpty([]string{"username", "email"}, post.Username, post.Email); err != nil {
		return c.JSON(http.StatusBadRequest, NewRESTValidationError(http.StatusBadRequest, "Failed to validate params", []*validate.FieldError{err}))
	}
	if err := uh.validator.Struct(post); err != nil {
		return c.JSON(http.StatusBadRequest,
			NewRESTValidationError(http.StatusBadRequest, "Failed to validate fields", err))
	}

	existing, err := UserUseCase.Exists(ctx, post.ToDomain())
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, existing)
}
