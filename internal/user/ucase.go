package user

import (
	"context"

	"go.elastic.co/apm"
)

// UserUseCaseImpl ...
type UserUseCaseImpl struct {
	UserRepository UserRepository
}

var _ UserUseCase = &UserUseCaseImpl{}

// NewUserUseCase ...
func NewUserUseCase(
	UserRepository UserRepository,
) *UserUseCaseImpl {
	return &UserUseCaseImpl{
		UserRepository: UserRepository,
	}
}

// SignUp create a user
func (uu *UserUseCaseImpl) SignUp(ctx context.Context, post *UserModel) (*UserModel, error) {
	apmSpan, _ := apm.StartSpan(ctx, "UserUseCaseImpl.Register", "service")
	defer apmSpan.End()

	ur := uu.UserRepository
	// search for existence
	if m, err := ur.FindByCredential(ctx, post); err != nil {
		return nil, err
	} else if m != nil {
		return nil, ErrDuplicatedUser
	}

	// save user
	if err := ur.SaveUser(ctx, post); err != nil {
		return nil, err
	}
	return post, nil
}

// Exists find if user exists in database
func (uu *UserUseCaseImpl) Exists(ctx context.Context, post *UserModel) (bool, error) {
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
