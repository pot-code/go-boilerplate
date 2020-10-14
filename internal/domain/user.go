package domain

import (
	"context"
	"errors"
)

type UserModel struct {
	ID         string `json:"id"`
	Username   string `json:"username" validate:"required,min=6,max=31"`
	Email      string `json:"email" validate:"required,email"`
	Password   string `json:"password" validate:"required,min=6"`
	LoginRetry int    `json:"-"`
	LastLogin  int64  `json:"-"`
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
