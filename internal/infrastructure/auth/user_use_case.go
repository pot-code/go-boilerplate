package auth

import (
	"context"

	"github.com/pot-code/go-boilerplate/internal/domain"
	"go.elastic.co/apm"
)

// UserUseCase ...
type UserUseCase struct {
	UserRepository domain.UserRepository
}

// NewUserUseCase ...
func NewUserUseCase(
	UserRepository domain.UserRepository,
) *UserUseCase {
	return &UserUseCase{
		UserRepository: UserRepository,
	}
}

// SignUp create a user
func (uu *UserUseCase) SignUp(ctx context.Context, post *domain.UserModel) (*domain.UserModel, error) {
	apmSpan, _ := apm.StartSpan(ctx, "UserUseCase.Register", "service")
	defer apmSpan.End()

	ur := uu.UserRepository
	// search for existence
	if m, err := ur.FindByCredential(ctx, post); err != nil {
		return nil, err
	} else if m != nil {
		return nil, domain.ErrDuplicatedUser
	}

	// save user
	if err := ur.SaveUser(ctx, post); err != nil {
		return nil, err
	}
	return post, nil
}

// Exists find if user exists in database
func (uu *UserUseCase) Exists(ctx context.Context, post *domain.UserModel) (bool, error) {
	apmSpan, _ := apm.StartSpan(ctx, "UserUseCase.Existing", "service")
	defer apmSpan.End()

	user, err := uu.UserRepository.FindByCredential(ctx, post)
	if err != nil {
		return false, err
	}
	if user == nil {
		return false, nil
	}
	return true, nil
}
