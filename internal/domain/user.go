package domain

import (
	"context"
	"errors"
)

type UserModel struct {
	ID         string
	Username   string
	Email      string
	Password   string
	LoginRetry int
	LastLogin  int64
}

// ErrDuplicatedUser unique key constraint violation
var ErrDuplicatedUser = errors.New("Username or email is already registered")

type UserUseCase interface {
	SignUp(ctx context.Context, post *UserModel) (*UserModel, error)
	Exists(ctx context.Context, post *UserModel) (bool, error)
}

type UserRepository interface {
	FindByCredential(ctx context.Context, post *UserModel) (*UserModel, error)
	UpdateUser(ctx context.Context, post *UserModel) error
	SaveUser(ctx context.Context, post *UserModel) error
}
