package domain

import (
	"context"
)

type UserModel struct {
	ID       string `json:"id"`
	Username string `json:"username" validate:"required,min=6,max=31"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}

type UserUseCase interface {
	SignIn(ctx context.Context, post *UserModel) (*UserModel, error)
	SignUp(ctx context.Context, post *UserModel) (*UserModel, error)
	Exists(ctx context.Context, post *UserModel) (bool, error)
}

type UserRepository interface {
	FindByCredential(ctx context.Context, post *UserModel) (*UserModel, error)
	SaveUser(ctx context.Context, post *UserModel) error
	BeginTx(ctx context.Context) error
	EndTx(ctx context.Context, err error) error
}
